// Package ui (continued) — native raylib gradient helpers for preset chrome.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// DrawVerticalGradientRect fills bounds with a native vertical gradient.
func DrawVerticalGradientRect(bounds rl.Rectangle, top, bottom rl.Color) {
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	rl.DrawRectangleGradientV(
		int32(bounds.X), int32(bounds.Y),
		int32(bounds.Width), int32(bounds.Height),
		top, bottom,
	)
}

// DrawHorizontalGradientRect fills bounds with a native horizontal gradient.
func DrawHorizontalGradientRect(bounds rl.Rectangle, left, right rl.Color) {
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	rl.DrawRectangleGradientH(
		int32(bounds.X), int32(bounds.Y),
		int32(bounds.Width), int32(bounds.Height),
		left, right,
	)
}

// DrawCornerGradientRect fills bounds with a four-corner gradient.
//
// Raylib's DrawRectangleGradientEx parameter names do not match vertex order:
// colors are applied as topLeft, bottomLeft, bottomRight, topRight (see raylib PR #4980).
func DrawCornerGradientRect(bounds rl.Rectangle, topLeft, bottomLeft, topRight, bottomRight rl.Color) {
	if bounds.Width < 1 || bounds.Height < 1 {
		return
	}
	rl.DrawRectangleGradientEx(bounds, topLeft, bottomLeft, bottomRight, topRight)
}
