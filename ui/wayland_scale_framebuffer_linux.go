//go:build linux && !android

package ui

/*
#cgo LDFLAGS: -Wl,--wrap=glfwDefaultWindowHints

// Linux Wayland present policy: layout px == framebuffer (same idea as Windows
// S1 / typical X11). Raylib's InitWindow → glfwDefaultWindowHints() resets
// GLFW_SCALE_FRAMEBUFFER to TRUE, which makes FB = logical × contentScale on
// any fractional-scale compositor (GNOME, KDE, Sway, Cosmic, …). We wrap the
// reset and force the hint off so the window stays the size the client set —
// no DPI expand, no Camera densify. Harmless on X11 (FB already matches).

extern void __real_glfwDefaultWindowHints(void);
extern void glfwWindowHint(int hint, int value);

#define GRU_GLFW_SCALE_FRAMEBUFFER 0x0002200D

void __wrap_glfwDefaultWindowHints(void) {
	__real_glfwDefaultWindowHints();
	glfwWindowHint(GRU_GLFW_SCALE_FRAMEBUFFER, 0);
}
*/
import "C"
