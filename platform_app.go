// Platform helpers for desktop vs Android (raylib NativeActivity).
//
// This file is compiled on every GOOS; branches use isAndroidApp() so desktop
// keeps custom chrome and default window size (MSAA on Windows only). Android-only
// entry is main_android_init.go (build tag). See docs/ANDROID_CODE.md.
package main

import (
	"runtime"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func isAndroidApp() bool { return runtime.GOOS == "android" }

// appUsesCustomChrome is false on Android — fullscreen NativeActivity, no Gru title bar.
func appUsesCustomChrome() bool { return !isAndroidApp() }

func appInitWindowFlags() uint32 {
	if isAndroidApp() {
		// MSAA breaks EGL context creation on Android emulators/devices.
		return uint32(rl.FlagVsyncHint)
	}
	// Linux/Wayland: no MSAA — resolve softens 1× present (same as S1HostWindowFlags).
	// Windows keeps MSAA per docs/architecture/rendering.md.
	flags := uint32(rl.FlagVsyncHint)
	if runtime.GOOS != "linux" {
		flags |= rl.FlagMsaa4xHint
	}
	if appUsesCustomChrome() {
		flags |= rl.FlagWindowResizable | rl.FlagWindowUndecorated
		// Hide until the first frame is drawn so users never see empty chrome
		// while atlases / SSAA / scene Build run (Studio, grudemo, and release).
		flags |= rl.FlagWindowHidden
	}
	return flags
}

func appInitWindowSize() (w, h int32) {
	if isAndroidApp() {
		// Hint portrait before EGL init (manifest also locks portrait).
		// Raylib replaces these with ANativeWindow dimensions after InitWindow.
		return 1080, 1920
	}
	return initWindowW, initWindowH
}

// appNormalizeAndroidClientSize swaps dimensions when the GL surface reports
// landscape width/height on a portrait-locked activity (stale APK or raylib quirk).
func appNormalizeAndroidClientSize(w, h int32) (int32, int32) {
	if !isAndroidApp() || w <= 0 || h <= 0 {
		return w, h
	}
	if w > h {
		return h, w
	}
	return w, h
}

func appInitWindowTitle() string {
	if isAndroidApp() {
		return "Gru Notepad"
	}
	return desktopInitWindowTitle()
}

// androidBackPressed returns true when the hardware back key should close the app.
func androidBackPressed() bool {
	if !isAndroidApp() {
		return false
	}
	return rl.IsKeyPressed(rl.KeyBack)
}
