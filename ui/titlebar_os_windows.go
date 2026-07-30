//go:build windows

package ui

import (
	"syscall"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	tbOsWmNCLButtonDblClk = 0x00A3
	tbOsHtCaption         = 2
	tbOsSwMaximize        = 3
	tbOsSwRestore         = 9
)

var (
	tbOsUser32              = syscall.NewLazyDLL("user32.dll")
	tbOsPostMessageW        = tbOsUser32.NewProc("PostMessageW")
	tbOsShowWindow          = tbOsUser32.NewProc("ShowWindow")
	tbOsIsZoomed            = tbOsUser32.NewProc("IsZoomed")
	tbOsGetDoubleClickTime  = tbOsUser32.NewProc("GetDoubleClickTime")
)

// WireBorderlessTitleBarOS installs Windows double-click maximize + click interval
// for thin hosts (calc/hello/webviewhello). Studio also wires this via main.
func WireBorderlessTitleBarOS() {
	if ms, _, _ := tbOsGetDoubleClickTime.Call(); ms > 0 && ms < 2000 {
		SetTitleBarDoubleClickInterval(float64(ms) / 1000.0)
	}
	SetTitleBarNativeDoubleClickMaximize(func() {
		NotifyWebViewLayoutJump()
		if !rl.IsWindowReady() {
			return
		}
		hwnd := uintptr(rl.GetWindowHandle())
		if hwnd == 0 {
			return
		}
		wasZoomed := false
		if z, _, _ := tbOsIsZoomed.Call(hwnd); z != 0 {
			wasZoomed = true
		}
		tbOsPostMessageW.Call(hwnd, tbOsWmNCLButtonDblClk, tbOsHtCaption, 0)
		nowZoomed := false
		if z, _, _ := tbOsIsZoomed.Call(hwnd); z != 0 {
			nowZoomed = true
		}
		if nowZoomed == wasZoomed {
			if wasZoomed {
				tbOsShowWindow.Call(hwnd, tbOsSwRestore)
			} else {
				tbOsShowWindow.Call(hwnd, tbOsSwMaximize)
			}
		}
	})
}
