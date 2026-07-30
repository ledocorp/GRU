//go:build !windows && !linux

package main

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func applyNativeBorderlessRoundedCorners(enabled bool) {
	_ = enabled
}

func platformWindowFocus() {}

func platformWindowShow() {
	platformWindowFocus()
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

func platformWindowClose(closeRequested *bool) {
	if closeRequested != nil {
		*closeRequested = true
	}
}
