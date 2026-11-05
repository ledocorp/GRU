//go:build windows

package ui

import (
	"fmt"
	"syscall"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// GWL_EXSTYLE = -20; encoded as uintptr for Get/SetWindowLongPtrW.
	gwlExStyle  = ^uintptr(19)
	wsExLayered = 0x00080000
	lwaAlpha    = 0x00000002
	lwaColorKey = 0x00000001

	wcaAccentPolicy              = 19
	accentDisabled               = 0
	accentEnableBlurBehind       = 3
	accentEnableAcrylicBlurBehind = 4
)

type accentPolicy struct {
	AccentState   int32
	AccentFlags   int32
	GradientColor uint32 // ABGR
	AnimationId   int32
}

type windowCompositionAttribData struct {
	Attribute int32
	Data      uintptr
	SizeOfData uint32
}

var (
	user32                         = syscall.NewLazyDLL("user32.dll")
	procGetWindowLongPtrW          = user32.NewProc("GetWindowLongPtrW")
	procSetWindowLongPtrW          = user32.NewProc("SetWindowLongPtrW")
	procSetLayeredWindowAttributes = user32.NewProc("SetLayeredWindowAttributes")
	procSetWindowCompositionAttribute = user32.NewProc("SetWindowCompositionAttribute")
)

func hostWindowTransparencyHint() string {
	return "Windows: Raylib SetWindowOpacity uses layered LWA_ALPHA. " +
		"FLAG_WINDOW_TRANSPARENT (per-pixel framebuffer alpha) must be set before InitWindow — " +
		"use go run ./cmd/transparentdemo for that path."
}

// EnableHostLayeredStyle adds WS_EX_LAYERED on the Raylib HWND (idempotent).
func EnableHostLayeredStyle() bool {
	hwnd := hostHWND()
	if hwnd == 0 {
		return false
	}
	style, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
	if style&wsExLayered != 0 {
		return true
	}
	procSetWindowLongPtrW.Call(hwnd, gwlExStyle, style|wsExLayered)
	return true
}

// ApplyHostLayeredAlpha sets whole-window alpha via SetLayeredWindowAttributes.
// Prefer SetHostWindowOpacity for normal use; this is the explicit Win32 path for demos.
func ApplyHostLayeredAlpha(alpha uint8) {
	hwnd := hostHWND()
	if hwnd == 0 {
		return
	}
	EnableHostLayeredStyle()
	procSetLayeredWindowAttributes.Call(hwnd, 0, uintptr(alpha), uintptr(lwaAlpha))
}

// HostWindowExStyleReport returns WS_EX_LAYERED status for UI readouts.
func HostWindowExStyleReport() string {
	hwnd := hostHWND()
	if hwnd == 0 {
		return "hwnd=0 (window not ready)"
	}
	style, _, _ := procGetWindowLongPtrW.Call(hwnd, gwlExStyle)
	layered := style&wsExLayered != 0
	return fmt.Sprintf("HWND=0x%X  WS_EX_LAYERED=%v  exStyle=0x%X", hwnd, layered, style)
}

func hostHWND() uintptr {
	if !rl.IsWindowReady() {
		return 0
	}
	return uintptr(rl.GetWindowHandle())
}

// ClearHostColorKey clears any color-key layered mode (no-op if unused).
func ClearHostColorKey() {
	hwnd := hostHWND()
	if hwnd == 0 {
		return
	}
	// Re-assert alpha mode so a prior COLORKEY experiment does not stick.
	_ = unsafe.Sizeof(lwaColorKey)
	EnableHostLayeredStyle()
	procSetLayeredWindowAttributes.Call(hwnd, 0, 255, uintptr(lwaAlpha))
}

// EnableHostAcrylicBlur asks DWM for acrylic/blur-behind on the Raylib HWND.
// tintABGR is Windows ABGR (e.g. 0xA8000000 = dark ~66% transparent). Returns
// false if the API is missing or the call fails (common with some GL visuals).
func EnableHostAcrylicBlur(tintABGR uint32) bool {
	return setHostAccent(accentEnableAcrylicBlurBehind, tintABGR)
}

// EnableHostBlurBehind asks DWM for classic blur-behind (fallback if acrylic fails).
func EnableHostBlurBehind(tintABGR uint32) bool {
	return setHostAccent(accentEnableBlurBehind, tintABGR)
}

// DisableHostBlurBehind clears DWM accent blur/acrylic.
func DisableHostBlurBehind() {
	setHostAccent(accentDisabled, 0)
}

func setHostAccent(state int32, tintABGR uint32) bool {
	hwnd := hostHWND()
	if hwnd == 0 || procSetWindowCompositionAttribute.Find() != nil {
		return false
	}
	policy := accentPolicy{
		AccentState:   state,
		AccentFlags:   2, // draw all borders
		GradientColor: tintABGR,
	}
	data := windowCompositionAttribData{
		Attribute:  wcaAccentPolicy,
		Data:       uintptr(unsafe.Pointer(&policy)),
		SizeOfData: uint32(unsafe.Sizeof(policy)),
	}
	r, _, _ := procSetWindowCompositionAttribute.Call(hwnd, uintptr(unsafe.Pointer(&data)))
	return r != 0
}
