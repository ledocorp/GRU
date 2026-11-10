//go:build linux && !android

package ui

/*
#cgo LDFLAGS: -lX11
#include <stdlib.h>
#include <X11/Xlib.h>
#include <X11/Xutil.h>
// Raylib statically links GLFW; resolve at link time.
extern void glfwGetFramebufferSize(void *window, int *width, int *height);
extern void *glfwGetPrimaryMonitor(void);
extern void glfwGetMonitorContentScale(void *monitor, float *xscale, float *yscale);
extern void glfwGetWindowContentScale(void *window, float *xscale, float *yscale);
extern void glfwSetWindowAttrib(void *window, int attrib, int value);
extern void glfwFocusWindow(void *window);
extern void glfwInitHint(int hint, int value);
extern void glfwWindowHintString(int hint, const char *value);
extern int glfwGetPlatform(void);
extern unsigned long glfwGetX11Window(void *window);
extern void *glfwGetX11Display(void);

static void gru_x11_threads_once(void) {
	static int done = 0;
	if (done) return;
	done = 1;
	XInitThreads();
}

static void gru_x11_set_wm_class(void *glfw_window, const char *id) {
	if (glfw_window == NULL || id == NULL || id[0] == 0) {
		return;
	}
	if (glfwGetPlatform() != 0x00060004) { // GLFW_PLATFORM_X11
		return;
	}
	Display *dpy = (Display *)glfwGetX11Display();
	Window win = (Window)glfwGetX11Window(glfw_window);
	if (dpy == NULL || win == 0) {
		return;
	}
	XClassHint hint;
	hint.res_name = (char *)id;
	hint.res_class = (char *)id;
	XSetClassHint(dpy, win, &hint);
	XFlush(dpy);
}
*/
import "C"

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// GLFW constants (must match glfw3.h).
const (
	glfwDecorated              = 0x00020005
	glfwWaylandLibdecor        = 0x00053001
	glfwWaylandDisableLibdecor = 0x00038002
	glfwPlatformHint           = 0x00050003 // GLFW_PLATFORM init hint
	glfwPlatformX11            = 0x00060004
	glfwPlatformWayland        = 0x00060003
	glfwX11ClassName           = 0x00024001 // GLFW_X11_CLASS_NAME
	glfwX11InstanceName        = 0x00024002 // GLFW_X11_INSTANCE_NAME
	glfwWaylandAppID           = 0x00026001 // GLFW_WAYLAND_APP_ID
)

// PrepareLinuxWindowHints runs before glfwInit for both X11 and Wayland.
//
// Linux present policy (not compositor-specific): window size is the client
// size; no DPI zoom / no denser-FB layout. Wayland: disable libdecor so
// FlagWindowUndecorated stays client-side only; SCALE_FRAMEBUFFER is forced
// off after raylib's hint reset (see wayland_scale_framebuffer_linux.go) so
// layout px == FB on GNOME/KDE/Sway/Cosmic alike. X11: usually already 1:1.
//
// When GLFW_PLATFORM=x11 is set, also apply the GLFW_PLATFORM init hint — the
// env var alone is unreliable with raylib's statically linked GLFW.
func PrepareLinuxWindowHints() {
	C.glfwInitHint(glfwWaylandLibdecor, glfwWaylandDisableLibdecor)
	if strings.EqualFold(strings.TrimSpace(os.Getenv("GLFW_PLATFORM")), "x11") {
		// GLFW + GTK/WebKit both talk to Xlib — must enable threads before either inits.
		C.gru_x11_threads_once()
		C.glfwInitHint(glfwPlatformHint, glfwPlatformX11)
	}
}

// EnsureX11ThreadsForWebKit calls XInitThreads when live WebKit may share the
// display with GLFW. Safe to call multiple times; no-op if not needed yet.
func EnsureX11ThreadsForWebKit() {
	if !GLFWPlatformIsX11() && !strings.EqualFold(strings.TrimSpace(os.Getenv("GLFW_PLATFORM")), "x11") {
		return
	}
	C.gru_x11_threads_once()
}

// PrepareLinuxAppID sets Wayland app_id + X11 WM_CLASS before InitWindow.
//
// Cosmic/GNOME match running windows to .desktop via app_id == StartupWMClass.
// Without this, the dock shows a generic gear instead of Icon= from grudemo.desktop.
// Call after PrepareLinuxWindowHints, before rl.InitWindow. id should match the
// desktop file (e.g. "grudemo" for packaging/linux/grudemo.desktop).
func PrepareLinuxAppID(id string) {
	id = strings.TrimSpace(id)
	if id == "" {
		return
	}
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))
	C.glfwWindowHintString(glfwWaylandAppID, cID)
	C.glfwWindowHintString(glfwX11ClassName, cID)
	C.glfwWindowHintString(glfwX11InstanceName, cID)
	linuxAppID = id
}

var linuxAppID string

// ReinforceLinuxAppID re-applies X11 WM_CLASS after InitWindow (Cosmic/XWayland
// dock matching). Wayland app_id was set via hint before create.
func ReinforceLinuxAppID() {
	id := strings.TrimSpace(linuxAppID)
	if id == "" || !rl.IsWindowReady() {
		return
	}
	handle := rl.GetWindowHandle()
	if handle == nil {
		return
	}
	cID := C.CString(id)
	defer C.free(unsafe.Pointer(cID))
	C.gru_x11_set_wm_class(handle, cID)
}

// GLFWPlatformIsX11 reports the real runtime platform after InitWindow.
// Safe on Wayland — does not call glfwGetX11Window (that emits 65550).
func GLFWPlatformIsX11() bool {
	if !rl.IsWindowReady() {
		return false
	}
	return int(C.glfwGetPlatform()) == glfwPlatformX11
}

// GLFWPlatformIsWayland reports the real runtime platform after InitWindow.
func GLFWPlatformIsWayland() bool {
	if !rl.IsWindowReady() {
		return os.Getenv("WAYLAND_DISPLAY") != "" &&
			!strings.EqualFold(strings.TrimSpace(os.Getenv("GLFW_PLATFORM")), "x11")
	}
	return int(C.glfwGetPlatform()) == glfwPlatformWayland
}

// platformFramebufferSize returns the real GLFW EGL buffer size (may exceed
// GetScreenWidth/Height under Wayland fractional scale). ok=false if unavailable.
func platformFramebufferSize() (w, h int32, ok bool) {
	handle := rl.GetWindowHandle()
	if handle == nil {
		return 0, 0, false
	}
	var cw, ch C.int
	C.glfwGetFramebufferSize(handle, &cw, &ch)
	if cw < 1 || ch < 1 {
		return 0, 0, false
	}
	return int32(cw), int32(ch), true
}

// platformWindowContentScale is the mapped window's GLFW content scale (Wayland
// fractional scale). Diagnostics only — Linux present does not zoom or densify
// from this value.
func platformWindowContentScale() (scale float32, ok bool) {
	handle := rl.GetWindowHandle()
	if handle == nil {
		return 1, false
	}
	var x, y C.float
	C.glfwGetWindowContentScale(handle, &x, &y)
	s := (float32(x) + float32(y)) * 0.5
	if s < 1 {
		s = 1
	}
	if s > 2.5 {
		s = 2.5
	}
	return s, true
}

// platformMonitorContentScale returns the primary monitor's GLFW content scale.
// Used while the window is still hidden (Wayland reports window DPI=1 until mapped).
func platformMonitorContentScale() (scale float32, ok bool) {
	mon := C.glfwGetPrimaryMonitor()
	if mon == nil {
		return 1, false
	}
	var x, y C.float
	C.glfwGetMonitorContentScale(mon, &x, &y)
	s := (float32(x) + float32(y)) * 0.5
	if s < 1 {
		s = 1
	}
	if s > 2.5 {
		s = 2.5
	}
	return s, true
}

var presentFBLogCount atomic.Uint32

func logPresentFB(sw, sh, fbW, fbH int32) {
	if os.Getenv("GRU_DEBUG_CLIENT") != "1" {
		return
	}
	n := presentFBLogCount.Add(1)
	if n > 6 && n%120 != 0 {
		return
	}
	fbScale := float32(1)
	if sw > 1 && sh > 1 {
		fbScale = (float32(fbW)/float32(sw) + float32(fbH)/float32(sh)) * 0.5
	}
	content := float32(1)
	if cs, ok := platformWindowContentScale(); ok {
		content = cs
	}
	fmt.Fprintf(os.Stderr, "gru present: logical=%dx%d framebuffer=%dx%d dpi=%.2f fbScale=%.2f contentScale=%.2f msaa=off(linux)\n",
		sw, sh, fbW, fbH, DisplayScale, fbScale, content)
}

// EnsureWindowUndecorated re-asserts GLFW_DECORATED=false after map/show.
func EnsureWindowUndecorated() {
	if !rl.IsWindowReady() {
		return
	}
	rl.SetWindowState(uint32(rl.FlagWindowUndecorated))
	handle := rl.GetWindowHandle()
	if handle != nil {
		C.glfwSetWindowAttrib(handle, glfwDecorated, 0)
	}
}

// FocusPlatformWindow requests keyboard focus on the GLFW window (Wayland raise
// fallback when X11 _NET_ACTIVE_WINDOW is unavailable).
func FocusPlatformWindow() {
	if !rl.IsWindowReady() {
		return
	}
	handle := rl.GetWindowHandle()
	if handle == nil {
		return
	}
	C.glfwFocusWindow(handle)
}

// platformKeepDisplayScaleUnity is true on Linux (X11 and Wayland): DisplayScale
// stays 1 so rem/atlases are not inflated by compositor DPI. Window size is
// whatever the user/client is; we do not "adjust for" fractional scale.
// Windows still follows DPI via RefreshDisplayScale.
func platformKeepDisplayScaleUnity() bool { return true }

// PreferOSWindowChrome is reserved for hosts that opt into SSD. Default false —
// Gru uses custom TitleBar; libdecor is disabled on Wayland.
func PreferOSWindowChrome() bool { return false }

// BeginPresentReveal maps a FlagWindowHidden window and briefly polls so the
// client size is stable before first paint. SSAA stays 1×.
func BeginPresentReveal() {
	if !rl.IsWindowReady() {
		return
	}
	keepCSD := rl.IsWindowState(uint32(rl.FlagWindowUndecorated))
	if keepCSD {
		EnsureWindowUndecorated()
	}
	if rl.IsWindowHidden() {
		rl.ClearWindowState(uint32(rl.FlagWindowHidden))
	}
	if keepCSD {
		EnsureWindowUndecorated()
	}
	RefreshDisplayScale()
	for i := 0; i < 16; i++ {
		rl.PollInputEvents()
		RefreshDisplayScale()
		if keepCSD {
			EnsureWindowUndecorated()
		}
		sw := int32(rl.GetScreenWidth())
		sh := int32(rl.GetScreenHeight())
		fbW, fbH, ok := platformFramebufferSize()
		if !ok {
			continue
		}
		if fbW > sw+1 || fbH > sh+1 || i >= 3 {
			return
		}
	}
}

// FinishPresentReveal re-asserts undecorated after the first scaled present.
func FinishPresentReveal() {
	if rl.IsWindowReady() && rl.IsWindowState(uint32(rl.FlagWindowUndecorated)) {
		EnsureWindowUndecorated()
	}
}
