//go:build linux && !android

package ui

// Wayland does not expose global window positions (GLFW FEATURE_UNAVAILABLE).
// Trust the real GLFW platform after InitWindow — never the GLFW_PLATFORM env
// alone (env can say x11 while GLFW is still on Wayland → SetWindowPosition spam).
func windowPositionSupported() bool {
	return GLFWPlatformIsX11()
}
