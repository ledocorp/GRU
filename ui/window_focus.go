package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// PlatformWindowFocused reports whether the host should accept keyboard input.
// On Android NativeActivity, raylib IsWindowFocused is desktop-oriented and can
// stay false even while the activity is resumed — treat GLES as focused.
func PlatformWindowFocused() bool {
	if AndroidGLES() {
		return true
	}
	return rl.IsWindowFocused()
}
