package ui

// WindowPositionSupported reports whether Get/SetWindowPosition are usable.
// Wayland returns false — hosts must not call SetWindowPosition (GLFW 65548).
func WindowPositionSupported() bool {
	return windowPositionSupported()
}
