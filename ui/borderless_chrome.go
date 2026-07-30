// Package ui (continued) — borderless window chrome rounding state.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// borderlessRoundedChrome is set each frame from the title bar (e.g. while dragging
// out of maximized) so draw/fill paths can restore rounded corners immediately.
var borderlessRoundedChrome bool

// SetBorderlessRoundedChrome updates whether borderless mode draws rounded corners.
func SetBorderlessRoundedChrome(rounded bool) {
	borderlessRoundedChrome = rounded
}

// BorderlessChromeRounded reports whether borderless window fill/title chrome
// should use rounded corners (false when edge-to-edge maximized).
func BorderlessChromeRounded() bool {
	if borderlessRoundedChrome {
		return true
	}
	return !rl.IsWindowMaximized()
}
