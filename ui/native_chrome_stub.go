//go:build !windows

package ui

// ApplyNativeBorderlessRoundedCorners is a no-op off Windows (no DWM corner API).
func ApplyNativeBorderlessRoundedCorners(enabled bool) {
	_ = enabled
}

// nativeBorderlessUsesOSCornerClip is false off Windows — drawn rounded fill is required.
func nativeBorderlessUsesOSCornerClip() bool { return false }
