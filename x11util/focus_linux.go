//go:build linux

// Package x11util activates Gru Notepad windows on X11 (WSLg, Linux desktop).
package x11util

/*
#cgo linux LDFLAGS: -lX11
#include <X11/Xlib.h>
#include <X11/Xatom.h>
#include <string.h>
#include <stdlib.h>

static int gru_xio_error(Display *dpy) {
	(void)dpy;
	return 0;
}

static int gru_xerror(Display *dpy, XErrorEvent *ev) {
	(void)dpy;
	(void)ev;
	return 0;
}

static void gru_install_x11_handlers(void) {
	XSetIOErrorHandler(gru_xio_error);
	XSetErrorHandler(gru_xerror);
}

static int gru_window_viewable(Display *dpy, Window w) {
	XWindowAttributes attrs;
	if (!XGetWindowAttributes(dpy, w, &attrs)) {
		return 0;
	}
	return attrs.map_state == IsViewable;
}

static int gru_title_matches(Display *dpy, Window w, const char *exact, const char *contains) {
	Atom netWMName = XInternAtom(dpy, "_NET_WM_NAME", False);
	Atom utf8 = XInternAtom(dpy, "UTF8_STRING", True);
	Atom actual;
	int fmt;
	unsigned long n, after;
	unsigned char *data = NULL;

	if (utf8 != None &&
		XGetWindowProperty(dpy, w, netWMName, 0, 1024, False, utf8,
			&actual, &fmt, &n, &after, &data) == Success && data && n > 0) {
		int ok = 0;
		if (exact && strcmp((char *)data, exact) == 0) {
			ok = 1;
		} else if (contains && strstr((char *)data, contains) != NULL) {
			ok = 1;
		}
		XFree(data);
		if (ok) {
			return 1;
		}
	}

	char *name = NULL;
	if (XFetchName(dpy, w, &name) && name) {
		int ok = 0;
		if (exact && strcmp(name, exact) == 0) {
			ok = 1;
		} else if (contains && strstr(name, contains) != NULL) {
			ok = 1;
		}
		XFree(name);
		return ok;
	}
	return 0;
}

// gru_find_toplevel searches _NET_CLIENT_LIST_STACKING (top-level only — avoids BadWindow on child surfaces).
static Window gru_find_toplevel(Display *dpy, Window root, const char *exact, const char *contains) {
	Atom listAtom = XInternAtom(dpy, "_NET_CLIENT_LIST_STACKING", False);
	Atom actual;
	int fmt;
	unsigned long n, after;
	unsigned char *data = NULL;
	if (XGetWindowProperty(dpy, root, listAtom, 0, 4096, False, XA_WINDOW,
			&actual, &fmt, &n, &after, &data) != Success || !data || n == 0) {
		return None;
	}
	Window *windows = (Window *)data;
	Window found = None;
	for (unsigned long i = 0; i < n; i++) {
		Window w = windows[i];
		if (gru_window_viewable(dpy, w) && gru_title_matches(dpy, w, exact, contains)) {
			found = w;
			break;
		}
	}
	XFree(data);
	return found;
}

static int gru_activate(Display *dpy, Window w) {
	if (dpy == NULL || w == None || !gru_window_viewable(dpy, w)) {
		return 0;
	}
	gru_install_x11_handlers();
	XRaiseWindow(dpy, w);
	XEvent ev;
	memset(&ev, 0, sizeof(ev));
	ev.xclient.type = ClientMessage;
	ev.xclient.window = w;
	ev.xclient.message_type = XInternAtom(dpy, "_NET_ACTIVE_WINDOW", False);
	ev.xclient.format = 32;
	ev.xclient.data.l[0] = 1;
	ev.xclient.data.l[1] = CurrentTime;
	ev.xclient.data.l[2] = 0;
	Window root = DefaultRootWindow(dpy);
	XSendEvent(dpy, root, False,
		SubstructureRedirectMask | SubstructureNotifyMask, &ev);
	XSync(dpy, False);
	return 1;
}

static int gru_activate_by_title(const char *exact, const char *contains) {
	gru_install_x11_handlers();
	Display *dpy = XOpenDisplay(NULL);
	if (!dpy) {
		return 0;
	}
	Window root = DefaultRootWindow(dpy);
	Window w = gru_find_toplevel(dpy, root, exact, contains);
	int ok = 0;
	if (w != None) {
		ok = gru_activate(dpy, w);
	}
	XCloseDisplay(dpy);
	return ok;
}
*/
import "C"
import "unsafe"

const (
	notepadTitleExact    = "Gru Notepad"
	notepadTitleContains = "Notepad (Go)"
)

// RaiseNotepadWindow finds a visible Gru Notepad top-level window and requests focus.
// Used when a second process forwards a file to the running instance.
func RaiseNotepadWindow() bool {
	exact := C.CString(notepadTitleExact)
	defer C.free(unsafe.Pointer(exact))
	contains := C.CString(notepadTitleContains)
	defer C.free(unsafe.Pointer(contains))
	return C.gru_activate_by_title(exact, contains) != 0
}
