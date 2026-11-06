//go:build linux && !android

package main

import (
	"github.com/ledocorp/gru/ui"
	"github.com/ledocorp/gru/x11util"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func applyNativeBorderlessRoundedCorners(enabled bool) {
	_ = enabled
}

func platformWindowFocus() {
	if !rl.IsWindowReady() {
		return
	}
	// GetWindowHandle() on Linux is GLFWwindow*, not an X11 Window id (raylib TODO).
	// X11/XWayland: raise by title over a side X connection. Pure Wayland: GLFW focus.
	if ui.GLFWPlatformIsX11() {
		if x11util.RaiseNotepadWindow() {
			return
		}
	}
	ui.FocusPlatformWindow()
}

func setupPlatformWindowHooks() {}

func platformWindowMinimize() {
	rl.MinimizeWindow()
}

func platformWindowToggleMaximize() {
	ui.NotifyWebViewLayoutJump()
	if rl.IsWindowMaximized() {
		rl.RestoreWindow()
	} else {
		rl.MaximizeWindow()
	}
}

func platformWindowShow() {
	if rl.IsWindowMinimized() {
		rl.RestoreWindow()
	}
	if rl.IsWindowHidden() {
		rl.ClearWindowState(uint32(rl.FlagWindowHidden))
	}
	platformWindowFocus()
}

func platformWindowClose(closeRequested *bool) {
	if closeRequested != nil {
		*closeRequested = true
	}
}
