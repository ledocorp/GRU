//go:build linux && !wayland

package ui

func init() {
	// raylib-go default: -D_GLFW_X11 -D_GLFW_WAYLAND
	glfwWinToplevelOff = 960
	glfwLibSeatOff = 133928
	glfwLibSerialOff = 134132
}
