//go:build linux

package ui

/*
#cgo linux pkg-config: wayland-client

#include <stdint.h>
#include <string.h>
#include <wayland-client-core.h>

// GLFW 3.4 (raylib-go v0.55.1) dual X11+Wayland layout offsets into the
// statically linked _glfw / GLFWwindow. Verified with offsetof against the
// same -D_GLFW_X11 -D_GLFW_WAYLAND flags raylib uses by default.
//
// Wayland-only builds (-tags wayland) use a slightly smaller window header
// (toplevel at +944); Go selects the constant via build tags below.

extern char _glfw;
// Incomplete array so fortified memset does not assume sizeof(CORE)==1.
extern unsigned char CORE[];

#define XDG_TOPLEVEL_MOVE 5
#define XDG_TOPLEVEL_RESIZE 6
#define GRU_GLFW_MOUSE_BUTTON_LAST 7
#define GRU_MAX_MOUSE_BUTTONS 8
#define GRU_MAX_TOUCH_POINTS 8

// CoreData offsets (raylib 5.5 / raylib-go v0.55.1) — MouseButtonCallback writes
// CORE.Input.Mouse/Touch; clearing GLFW alone does not unstick IsMouseButtonPressed.
#define GRU_CORE_MOUSE_CUR 1906
#define GRU_CORE_MOUSE_PREV 1914
#define GRU_CORE_TOUCH_CUR 2040
#define GRU_CORE_TOUCH_PREV 2048

static void gru_wl_clear_mouse_buttons(void *glfw_window, uintptr_t mouse_off)
{
	if (!glfw_window)
		return;
	char *buttons = (char *)glfw_window + mouse_off;
	for (int i = 0; i <= GRU_GLFW_MOUSE_BUTTON_LAST; i++)
		buttons[i] = 0;
	// Raylib mirrors presses via callback into CORE — force both current and
	// previous to 0 so the next real press is seen as a rising edge.
	memset(CORE + GRU_CORE_MOUSE_CUR, 0, GRU_MAX_MOUSE_BUTTONS);
	memset(CORE + GRU_CORE_MOUSE_PREV, 0, GRU_MAX_MOUSE_BUTTONS);
	memset(CORE + GRU_CORE_TOUCH_CUR, 0, GRU_MAX_TOUCH_POINTS);
	memset(CORE + GRU_CORE_TOUCH_PREV, 0, GRU_MAX_TOUCH_POINTS);
}

// When xdg_toplevel.move/resize steals the pointer, GLFW often never gets a
// RELEASE. Heal: if GLFW says up but CORE still shows down, clear CORE so the
// next PRESS is a rising edge. Returns 1 when a stuck-down bit was cleared
// (caller must ignore IsMouseButtonReleased this frame — it would be synthetic
// and can fire title-bar Close).
static int gru_wl_heal_mouse_buttons(void *glfw_window, uintptr_t mouse_off)
{
	if (!glfw_window)
		return 0;
	char *buttons = (char *)glfw_window + mouse_off;
	int healed = 0;
	for (int i = 0; i <= GRU_GLFW_MOUSE_BUTTON_LAST; i++) {
		if (buttons[i] != 0)
			continue;
		if (CORE[GRU_CORE_MOUSE_CUR + i] != 0) {
			CORE[GRU_CORE_MOUSE_CUR + i] = 0;
			healed = 1;
		}
		if (i < GRU_MAX_TOUCH_POINTS && CORE[GRU_CORE_TOUCH_CUR + i] != 0) {
			CORE[GRU_CORE_TOUCH_CUR + i] = 0;
			healed = 1;
		}
	}
	return healed;
}

static int gru_wl_interactive(void *glfw_window, uintptr_t toplevel_off,
	uintptr_t seat_off, uintptr_t serial_off, uintptr_t mouse_off, int resize_edges)
{
	if (!glfw_window)
		return 0;
	struct wl_proxy *toplevel = *(struct wl_proxy **)((char *)glfw_window + toplevel_off);
	struct wl_seat *seat = *(struct wl_seat **)((char *)&_glfw + seat_off);
	uint32_t serial = *(uint32_t *)((char *)&_glfw + serial_off);
	if (!toplevel || !seat || serial == 0)
		return 0;
	if (resize_edges < 0) {
		wl_proxy_marshal_flags(toplevel, XDG_TOPLEVEL_MOVE, NULL,
			wl_proxy_get_version(toplevel), 0, seat, serial);
	} else {
		wl_proxy_marshal_flags(toplevel, XDG_TOPLEVEL_RESIZE, NULL,
			wl_proxy_get_version(toplevel), 0, seat, serial, (uint32_t)resize_edges);
	}
	gru_wl_clear_mouse_buttons(glfw_window, mouse_off);
	return 1;
}
*/
import "C"

import (
	"os"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Offsets match raylib-go v0.55.1 default Linux build (_GLFW_X11|_GLFW_WAYLAND).
// See wayland_interactive_offsets_*.go for -tags wayland variant.
var (
	glfwWinToplevelOff    = uintptr(960)
	glfwLibSeatOff        = uintptr(133928)
	glfwLibSerialOff      = uintptr(134132)
	glfwWinMouseButtonsOff = uintptr(144) // same for wayland-only / dual
)

func waylandSessionActive() bool {
	if os.Getenv("WAYLAND_DISPLAY") != "" {
		if os.Getenv("GLFW_PLATFORM") == "x11" {
			return false
		}
		return true
	}
	return os.Getenv("XDG_SESSION_TYPE") == "wayland" && os.Getenv("GLFW_PLATFORM") != "x11"
}

// platformInteractiveMoveResizeSupported is true on a live Wayland session
// where xdg_toplevel.move/resize can drive the compositor.
func platformInteractiveMoveResizeSupported() bool {
	return waylandSessionActive()
}

// platformHealPointerButtons clears raylib CORE button bits that stayed down
// after a compositor-owned move/resize stole the pointer without a RELEASE.
// Call once per frame before reading IsMouseButtonPressed on Wayland.
// Returns true when a stuck-down bit was cleared (synthetic Released — ignore
// for title-bar Close).
func platformHealPointerButtons() bool {
	if !waylandSessionActive() || !rl.IsWindowReady() {
		return false
	}
	handle := rl.GetWindowHandle()
	if handle == nil {
		return false
	}
	return C.gru_wl_heal_mouse_buttons(unsafe.Pointer(handle), C.uintptr_t(glfwWinMouseButtonsOff)) != 0
}

func platformBeginInteractiveMove() bool {
	if !waylandSessionActive() || !rl.IsWindowReady() {
		return false
	}
	handle := rl.GetWindowHandle()
	if handle == nil {
		return false
	}
	ok := C.gru_wl_interactive(unsafe.Pointer(handle),
		C.uintptr_t(glfwWinToplevelOff),
		C.uintptr_t(glfwLibSeatOff),
		C.uintptr_t(glfwLibSerialOff),
		C.uintptr_t(glfwWinMouseButtonsOff),
		C.int(-1))
	// Mouse buttons cleared in C; raylib picks up RELEASE on the next
	// EndDrawing → PollInputEvents (do not poll mid-Update).
	return ok != 0
}

func platformBeginInteractiveResize(edge resizeEdge) bool {
	if !waylandSessionActive() || !rl.IsWindowReady() {
		return false
	}
	edges := waylandResizeEdges(edge)
	if edges == 0 {
		return false
	}
	handle := rl.GetWindowHandle()
	if handle == nil {
		return false
	}
	ok := C.gru_wl_interactive(unsafe.Pointer(handle),
		C.uintptr_t(glfwWinToplevelOff),
		C.uintptr_t(glfwLibSeatOff),
		C.uintptr_t(glfwLibSerialOff),
		C.uintptr_t(glfwWinMouseButtonsOff),
		C.int(edges))
	return ok != 0
}

func waylandResizeEdges(edge resizeEdge) int {
	// xdg_toplevel_resize_edge
	switch edge {
	case edgeN:
		return 1
	case edgeS:
		return 2
	case edgeW:
		return 4
	case edgeNW:
		return 5
	case edgeSW:
		return 6
	case edgeE:
		return 8
	case edgeNE:
		return 9
	case edgeSE:
		return 10
	default:
		return 0
	}
}
