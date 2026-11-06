package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// ClientLayoutSize is the Document / SSAA / chrome size: always the GLFW client
// (GetScreenWidth/Height). Linux X11 and Wayland use the window as-is — never
// substitute a denser framebuffer or apply DPI zoom. Windows is unchanged
// (client already equals the framebuffer).
func ClientLayoutSize() (w, h int32) {
	w = int32(rl.GetScreenWidth())
	h = int32(rl.GetScreenHeight())
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	ApplyClientInputScale(1, 1)
	return w, h
}

// ApplyClientInputScale maps mouse coords into layout space (1,1 = identity).
func ApplyClientInputScale(sx, sy float32) {
	if sx < 0.5 {
		sx = 1
	}
	if sy < 0.5 {
		sy = 1
	}
	rl.SetMouseScale(sx, sy)
}

// RawFramebufferScale is FB/logical for diagnostics only (not a present zoom).
func RawFramebufferScale() float32 {
	sw := int32(rl.GetScreenWidth())
	sh := int32(rl.GetScreenHeight())
	if sw < 1 || sh < 1 {
		return 1
	}
	fbW, fbH, ok := platformFramebufferSize()
	if !ok {
		return 1
	}
	return (float32(fbW)/float32(sw) + float32(fbH)/float32(sh)) * 0.5
}
