// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// overlayChromeTop/Bottom reserve the borderless title bar and launcher nav
// from screen-space overlays (drawer, notification center, etc.).
// Scene content uses Document.SetChromeTop/Bottom; call SetOverlayChromeInsets
// from main each frame when borderless chrome is active.
var overlayChromeTop, overlayChromeBottom float32

// SetOverlayChromeInsets configures the safe overlay band below the title bar
// and above the bottom nav. Also updates DrawerMgr insets (same contract).
func SetOverlayChromeInsets(top, bottom float32) {
	if top < 0 {
		top = 0
	}
	if bottom < 0 {
		bottom = 0
	}
	overlayChromeTop = top
	overlayChromeBottom = bottom
	DrawerMgr.SetContentInsets(top, bottom)
}

// OverlayContentBand returns the rectangle overlays may occupy (full width,
// between top/bottom chrome). Pass screen width and height in layout pixels.
func OverlayContentBand(sw, sh float32) rl.Rectangle {
	top := overlayChromeTop
	bottom := overlayChromeBottom
	h := sh - top - bottom
	if h < 1 {
		h = 1
	}
	return rl.NewRectangle(0, top, sw, h)
}
