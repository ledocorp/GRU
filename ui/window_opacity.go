package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// SetHostWindowOpacity sets whole-window opacity via Raylib (0–1).
// On Windows this typically enables WS_EX_LAYERED + LWA_ALPHA under the hood.
// Values below 0.25 are clamped so the window stays findable.
func SetHostWindowOpacity(opacity float32) {
	if !rl.IsWindowReady() {
		return
	}
	if opacity < 0.25 {
		opacity = 0.25
	}
	if opacity > 1 {
		opacity = 1
	}
	rl.SetWindowOpacity(opacity)
}

// ResetHostWindowOpacity restores full opacity.
func ResetHostWindowOpacity() {
	SetHostWindowOpacity(1)
}

// HostWindowTransparencyHint returns a short platform note for demos.
func HostWindowTransparencyHint() string {
	return hostWindowTransparencyHint()
}
