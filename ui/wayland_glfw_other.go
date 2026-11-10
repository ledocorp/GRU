//go:build !linux

package ui

// PrepareLinuxWindowHints is a no-op off Linux.
func PrepareLinuxWindowHints() {}

// PrepareLinuxAppID is a no-op off Linux.
func PrepareLinuxAppID(string) {}

// ReinforceLinuxAppID is a no-op off Linux.
func ReinforceLinuxAppID() {}

// EnsureX11ThreadsForWebKit is a no-op off Linux.
func EnsureX11ThreadsForWebKit() {}

// EnsureWindowUndecorated is a no-op off Linux (Windows uses DWM / flags elsewhere).
func EnsureWindowUndecorated() {}

// BeginPresentReveal is a no-op off Linux.
func BeginPresentReveal() {}

// FinishPresentReveal is a no-op off Linux.
func FinishPresentReveal() {}

// FocusPlatformWindow is a no-op off Linux (hosts use OS raise APIs).
func FocusPlatformWindow() {}

// platformFramebufferSize is unavailable off Linux (use GetRenderWidth fallback).
func platformFramebufferSize() (w, h int32, ok bool) { return 0, 0, false }

func platformMonitorContentScale() (scale float32, ok bool) { return 1, false }

// platformKeepDisplayScaleUnity is false off Linux — Windows/macOS follow DPI.
func platformKeepDisplayScaleUnity() bool { return false }

func logPresentFB(sw, sh, fbW, fbH int32) {}

// GLFWPlatformIsX11 is false off Linux.
func GLFWPlatformIsX11() bool { return false }

// GLFWPlatformIsWayland is false off Linux.
func GLFWPlatformIsWayland() bool { return false }
