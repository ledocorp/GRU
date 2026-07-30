//go:build !windows

package ui

// WireBorderlessTitleBarOS is a no-op off Windows (thin hosts still call it).
func WireBorderlessTitleBarOS() {}
