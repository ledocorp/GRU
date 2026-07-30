//go:build windows

package tray

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	wmUser     = 0x0400
	wmTrayIcon = wmUser + 1
	wmCommand   = 0x0111
	nimAdd      = 0x00000000
	nimModify   = 0x00000001
	nimDelete   = 0x00000002
	nifMessage  = 0x00000001
	nifIcon     = 0x00000002
	nifTip      = 0x00000004
	idShow uint32 = 1
	idQuit uint32 = 2
)

var (
	user32             = windows.NewLazySystemDLL("user32.dll")
	shell32            = windows.NewLazySystemDLL("shell32.dll")
	procRegisterClass  = user32.NewProc("RegisterClassExW")
	procCreateWindow   = user32.NewProc("CreateWindowExW")
	procDefWindowProc  = user32.NewProc("DefWindowProcW")
	procGetMessage     = user32.NewProc("GetMessageW")
	procTranslateMsg   = user32.NewProc("TranslateMessage")
	procDispatchMsg    = user32.NewProc("DispatchMessageW")
	procPostQuitMsg    = user32.NewProc("PostQuitMessage")
	procDestroyWindow  = user32.NewProc("DestroyWindow")
	procLoadImage      = user32.NewProc("LoadImageW")
	procCreatePopup    = user32.NewProc("CreatePopupMenu")
	procAppendMenu     = user32.NewProc("AppendMenuW")
	procTrackPopup     = user32.NewProc("TrackPopupMenu")
	procSetForeground  = user32.NewProc("SetForegroundWindow")
	procShellNotify    = shell32.NewProc("Shell_NotifyIconW")
	procGetModule      = windows.NewLazySystemDLL("kernel32.dll").NewProc("GetModuleHandleW")

	winTrayOnce sync.Once
	winTrayCfg  Config
)

// Start launches the Windows notification-area icon on a dedicated OS thread.
func Start(c Config) {
	if len(c.Icon) == 0 {
		return
	}
	winTrayCfg = c
	winTrayOnce.Do(func() {
		go func() {
			runtime.LockOSThread()
			runWinTray(c)
		}()
	})
}

// Stop removes the tray icon (posts quit to the tray thread).
func Stop() {
	// Best-effort; tray thread exits when the app closes.
}

func runWinTray(c Config) {
	iconPath, err := writeTrayICO(c.Icon)
	if err != nil {
		return
	}

	className, _ := windows.UTF16PtrFromString("GruTrayWindow")
	hInst, _, _ := procGetModule.Call(0)

	wndClass := wndClassEx{
		Style:      0,
		WndProc:    syscall.NewCallback(trayWndProc),
		Instance:   windows.Handle(hInst),
		ClassName:  className,
	}
	wndClass.Size = uint32(unsafe.Sizeof(wndClass))
	if r, _, _ := procRegisterClass.Call(uintptr(unsafe.Pointer(&wndClass))); r == 0 {
		return
	}

	title, _ := windows.UTF16PtrFromString("GruTray")
	hwnd, _, _ := procCreateWindow.Call(
		0,
		uintptr(unsafe.Pointer(className)),
		uintptr(unsafe.Pointer(title)),
		0, 0, 0, 0, 0,
		0, 0, hInst, 0,
	)
	if hwnd == 0 {
		return
	}
	defer procDestroyWindow.Call(hwnd)

	iconPathPtr, _ := windows.UTF16PtrFromString(iconPath)
	icon, _, _ := procLoadImage.Call(
		0,
		uintptr(unsafe.Pointer(iconPathPtr)),
		1, // IMAGE_ICON
		0, 0,
		0x00000010|0x00000040, // LR_LOADFROMFILE | LR_DEFAULTSIZE
	)
	if icon == 0 {
		return
	}

	nid := notifyIconData{
		Wnd:             windows.Handle(hwnd),
		ID:              1,
		Flags:           nifMessage | nifIcon | nifTip,
		CallbackMessage: wmTrayIcon,
		Icon:            windows.Handle(icon),
	}
	nid.Size = uint32(unsafe.Sizeof(nid))
	if c.Tooltip != "" {
		tip, _ := windows.UTF16FromString(c.Tooltip)
		copy(nid.Tip[:], tip)
	}
	if r, _, _ := procShellNotify.Call(nimAdd, uintptr(unsafe.Pointer(&nid))); r == 0 {
		return
	}
	defer procShellNotify.Call(nimDelete, uintptr(unsafe.Pointer(&nid)))

	var msg struct {
		HWnd    windows.Handle
		Message uint32
		WParam  uintptr
		LParam  uintptr
		Time    uint32
		Pt      struct{ X, Y int32 }
	}
	for {
		ret, _, _ := procGetMessage.Call(
			uintptr(unsafe.Pointer(&msg)),
			0, 0, 0,
		)
		if ret == 0 || ret == ^uintptr(0) {
			break
		}
		procTranslateMsg.Call(uintptr(unsafe.Pointer(&msg)))
		procDispatchMsg.Call(uintptr(unsafe.Pointer(&msg)))
	}
	if c.OnQuit != nil {
		c.OnQuit()
	}
}

func trayWndProc(hwnd, msg, wParam, lParam uintptr) uintptr {
	switch uint32(msg) {
	case wmTrayIcon:
		switch lParam {
		case 0x0205, 0x0204: // WM_LBUTTONUP, WM_RBUTTONUP
			showTrayMenu(windows.Handle(hwnd))
		}
	case wmCommand:
		switch uint32(wParam) {
		case idShow:
			if winTrayCfg.OnShow != nil {
				winTrayCfg.OnShow()
			}
		case idQuit:
			procPostQuitMsg.Call(0)
		}
	}
	r, _, _ := procDefWindowProc.Call(hwnd, msg, wParam, lParam)
	return r
}

func showTrayMenu(hwnd windows.Handle) {
	menu, _, _ := procCreatePopup.Call()
	if menu == 0 {
		return
	}
	defer user32.NewProc("DestroyMenu").Call(menu)

	showLabel, _ := windows.UTF16PtrFromString("Show")
	quitLabel, _ := windows.UTF16PtrFromString("Quit")
	const mfString = 0x00000000
	procAppendMenu.Call(menu, mfString, uintptr(idShow), uintptr(unsafe.Pointer(showLabel)))
	procAppendMenu.Call(menu, mfString, uintptr(idQuit), uintptr(unsafe.Pointer(quitLabel)))

	procSetForeground.Call(uintptr(hwnd))
	var pt struct{ X, Y int32 }
	user32.NewProc("GetCursorPos").Call(uintptr(unsafe.Pointer(&pt)))
	const tpmRightButton = 0x0002
	procTrackPopup.Call(
		menu,
		tpmRightButton,
		uintptr(pt.X),
		uintptr(pt.Y),
		0,
		uintptr(hwnd),
		0,
	)
}

func writeTrayICO(ico []byte) (string, error) {
	sum := md5.Sum(ico)
	name := "gru-tray-" + hex.EncodeToString(sum[:]) + ".ico"
	path := filepath.Join(os.TempDir(), name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.WriteFile(path, ico, 0o644); err != nil {
			return "", err
		}
	}
	return path, nil
}

type wndClassEx struct {
	Size, Style                        uint32
	WndProc                            uintptr
	ClsExtra, WndExtra                 int32
	Instance, Icon, Cursor, Background windows.Handle
	MenuName, ClassName                *uint16
	IconSm                             windows.Handle
}

type notifyIconData struct {
	Size                       uint32
	Wnd                        windows.Handle
	ID, Flags, CallbackMessage uint32
	Icon                       windows.Handle
	Tip                        [128]uint16
	State, StateMask           uint32
	Info                       [256]uint16
	Timeout, Version           uint32
	InfoTitle                  [64]uint16
	InfoFlags                  uint32
	Guid                       windows.GUID
	BalloonIcon                windows.Handle
}
