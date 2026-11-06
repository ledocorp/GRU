//go:build !android

package ui

// PlatformShowSoftInput is a no-op off Android.
func PlatformShowSoftInput() {}

// PlatformHideSoftInput is a no-op off Android.
func PlatformHideSoftInput() {}

// PlatformHapticTick is a no-op off Android.
func PlatformHapticTick() {}

// ArmIMEForUserGesture is a no-op off Android.
func ArmIMEForUserGesture() {}

func syncIMEForFocus(n Node) {}
