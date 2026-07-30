//go:build windows

package ui

import (
	"syscall"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	dwmwaWindowCornerPreference = 33
	dwmwcpDoNotRound            = 0
	dwmwcpRound                 = 2
)

var procDwmSetWindowAttribute = syscall.NewLazyDLL("dwmapi.dll").NewProc("DwmSetWindowAttribute")

// ApplyNativeBorderlessRoundedCorners asks DWM to round the window silhouette (Win11+).
// Must stay in sync with drawn fill: on Windows we paint a solid rect and let DWM
// clip corners (avoids DrawRectangleRounded seam artifacts on large empty areas).
func ApplyNativeBorderlessRoundedCorners(enabled bool) {
	if !rl.IsWindowReady() {
		return
	}
	hwnd := uintptr(rl.GetWindowHandle())
	if hwnd == 0 {
		return
	}
	pref := int32(dwmwcpDoNotRound)
	if enabled {
		pref = dwmwcpRound
	}
	procDwmSetWindowAttribute.Call(
		hwnd,
		uintptr(dwmwaWindowCornerPreference),
		uintptr(unsafe.Pointer(&pref)),
		unsafe.Sizeof(pref),
	)
}

// nativeBorderlessUsesOSCornerClip reports that DWM (not DrawRectangleRounded)
// provides the rounded window silhouette.
func nativeBorderlessUsesOSCornerClip() bool { return true }
