//go:build !linux || android

package ui

func windowPositionSupported() bool { return true }
