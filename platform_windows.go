//go:build windows

package main

import (
	"syscall"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	wmNCLButtonDblClk = 0x00A3
	wmClose           = 0x0010
	htCaption         = 2

	swMinimize = 6
	swMaximize = 3
	swRestore  = 9
)

var (
	modUser32               = syscall.NewLazyDLL("user32.dll")
	procPostMessageW        = modUser32.NewProc("PostMessageW")
	procShowWindow          = modUser32.NewProc("ShowWindow")
	procIsZoomed            = modUser32.NewProc("IsZoomed")
	procGetDoubleClickTime  = modUser32.NewProc("GetDoubleClickTime")
	procSetForegroundWindow = modUser32.NewProc("SetForegroundWindow")
	procIsIconic            = modUser32.NewProc("IsIconic")
)

func windowHWND() uintptr {
	if !rl.IsWindowReady() {
		return 0
	}
	return uintptr(rl.GetWindowHandle())
}

func windowIsZoomed(hwnd uintptr) bool {
	if hwnd == 0 {
		return false
	}
	z, _, _ := procIsZoomed.Call(hwnd)
	return z != 0
}

// applyNativeBorderlessRoundedCorners asks DWM to round the window silhouette (Win11+).
func applyNativeBorderlessRoundedCorners(enabled bool) {
	ui.ApplyNativeBorderlessRoundedCorners(enabled)
}

// platformWindowMinimize uses ShowWindow (smoother than raylib on Win32).
func platformWindowMinimize() {
	hwnd := windowHWND()
	if hwnd == 0 {
		return
	}
	procShowWindow.Call(hwnd, swMinimize)
}

// platformWindowToggleMaximize uses ShowWindow for native maximize/restore animation.
func platformWindowToggleMaximize() {
	ui.NotifyWebViewLayoutJump()
	hwnd := windowHWND()
	if hwnd == 0 {
		return
	}
	if windowIsZoomed(hwnd) {
		procShowWindow.Call(hwnd, swRestore)
	} else {
		procShowWindow.Call(hwnd, swMaximize)
	}
}

// platformWindowClose posts WM_CLOSE; main still sets closeRequested for safe teardown.
func platformWindowClose(closeRequested *bool) {
	hwnd := windowHWND()
	if hwnd != 0 {
		procPostMessageW.Call(hwnd, wmClose, 0, 0)
	}
	if closeRequested != nil {
		*closeRequested = true
	}
}

// nativeTitleBarToggleMaximize is title-bar double-click: NC dblclk then ShowWindow fallback.
func nativeTitleBarToggleMaximize() {
	ui.NotifyWebViewLayoutJump()
	hwnd := windowHWND()
	if hwnd == 0 {
		return
	}
	wasZoomed := windowIsZoomed(hwnd)
	procPostMessageW.Call(hwnd, wmNCLButtonDblClk, htCaption, 0)
	if windowIsZoomed(hwnd) != wasZoomed {
		return
	}
	platformWindowToggleMaximize()
}

// platformWindowShow restores a minimized/hidden window and brings it to the foreground.
func platformWindowShow() {
	hwnd := windowHWND()
	if hwnd == 0 {
		return
	}
	if iconic, _, _ := procIsIconic.Call(hwnd); iconic != 0 {
		procShowWindow.Call(hwnd, swRestore)
	}
	if rl.IsWindowHidden() {
		rl.ClearWindowState(uint32(rl.FlagWindowHidden))
	}
	platformWindowFocus()
}

// platformWindowFocus brings the game window to the foreground after launch.
func platformWindowFocus() {
	hwnd := windowHWND()
	if hwnd == 0 {
		return
	}
	procSetForegroundWindow.Call(hwnd)
}

func setupPlatformWindowHooks() {
	if ms, _, _ := procGetDoubleClickTime.Call(); ms > 0 && ms < 2000 {
		ui.SetTitleBarDoubleClickInterval(float64(ms) / 1000.0)
	}
	ui.SetTitleBarNativeDoubleClickMaximize(nativeTitleBarToggleMaximize)
}
