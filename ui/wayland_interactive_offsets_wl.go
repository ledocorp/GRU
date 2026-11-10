//go:build linux && !android && wayland

package ui

func init() {
	// raylib-go -tags wayland: -D_GLFW_WAYLAND only
	glfwWinToplevelOff = 944
	glfwLibSeatOff = 133928
	glfwLibSerialOff = 134132
}
