//go:build !grudemo

package examples

// PublicDemoMode is false for Studio / Notepad / Prism builds.
func PublicDemoMode() bool { return false }
