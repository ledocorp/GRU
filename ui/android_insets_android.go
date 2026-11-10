//go:build android

package ui

/*
#cgo LDFLAGS: -landroid -llog
#include <stdint.h>
#include <android/native_activity.h>
#include <android/native_window.h>
#include <android/window.h>
#include <android_native_app_glue.h>

extern struct android_app *GetAndroidApp(void);

static ARect gruContentRect;
static int gruContentRectValid;
static int gruInsetFlagsApplied;

static void gruOnContentRectChanged(ANativeActivity *activity, const ARect *rect) {
	(void)activity;
	if (rect == NULL) {
		return;
	}
	gruContentRect = *rect;
	gruContentRectValid = 1;
}

// EnsureAndroidInsetFlags clears raylib's AWINDOW_FLAG_FULLSCREEN and requests
// layout-in-screen + inset-decor so the WM reports a content ARect that clears
// the status / nav bars. Call once after rl.InitWindow.
static int gruEnsureAndroidInsetFlags(void) {
	struct android_app *app = GetAndroidApp();
	if (app == NULL || app->activity == NULL) {
		return -1;
	}
	if (app->activity->callbacks != NULL) {
		app->activity->callbacks->onContentRectChanged = gruOnContentRectChanged;
	}
	if (!gruInsetFlagsApplied) {
		ANativeActivity_setWindowFlags(app->activity,
			AWINDOW_FLAG_FORCE_NOT_FULLSCREEN |
				AWINDOW_FLAG_LAYOUT_IN_SCREEN |
				AWINDOW_FLAG_LAYOUT_INSET_DECOR,
			AWINDOW_FLAG_FULLSCREEN);
		gruInsetFlagsApplied = 1;
	}
	return 0;
}

// gruReadContentRect fills window size and content rectangle in window pixels.
// Returns 1 when a non-empty inset was observed.
static int gruReadContentRect(int32_t *winW, int32_t *winH,
	int32_t *left, int32_t *top, int32_t *right, int32_t *bottom) {
	struct android_app *app = GetAndroidApp();
	*winW = *winH = 0;
	*left = *top = *right = *bottom = 0;
	if (app == NULL) {
		return 0;
	}
	if (app->window != NULL) {
		*winW = ANativeWindow_getWidth(app->window);
		*winH = ANativeWindow_getHeight(app->window);
	}
	ARect r = app->contentRect;
	if (gruContentRectValid) {
		r = gruContentRect;
	}
	*left = r.left;
	*top = r.top;
	*right = r.right;
	*bottom = r.bottom;
	if (*winW < 1 || *winH < 1) {
		return 0;
	}
	// Empty / unset rect → treat as full window (no insets yet).
	if (r.right <= r.left || r.bottom <= r.top) {
		*left = 0;
		*top = 0;
		*right = *winW;
		*bottom = *winH;
		return 0;
	}
	return 1;
}
*/
import "C"

import (
	"fmt"
	"os"
)

// EnsureAndroidInsetFlags clears raylib fullscreen and enables inset-decor layout.
func EnsureAndroidInsetFlags() {
	if C.gruEnsureAndroidInsetFlags() != 0 {
		fmt.Fprintln(os.Stderr, "gru android: EnsureAndroidInsetFlags — no ANativeActivity")
	}
}

// AndroidContentInsets returns system-bar padding in window pixels suitable for
// Document.SetChromeTop / SetChromeBottom (and optional horizontal shell pad).
// When the WM has not yet published a content rect, falls back to ~48dp top /
// ~24dp bottom at 420dpi (Pixel / gru_api34 AVD).
func AndroidContentInsets(screenW, screenH int32) (top, bottom, left, right float32) {
	EnsureAndroidInsetFlags()
	var winW, winH, l, t, r, b C.int32_t
	ok := C.gruReadContentRect(&winW, &winH, &l, &t, &r, &b)
	w := int32(winW)
	h := int32(winH)
	if w < 1 {
		w = screenW
	}
	if h < 1 {
		h = screenH
	}
	if ok != 0 && w > 0 && h > 0 {
		top = float32(t)
		left = float32(l)
		if int32(r) < w {
			right = float32(w - int32(r))
		}
		if int32(b) < h {
			bottom = float32(h - int32(b))
		}
		// Clamp nonsense (e.g. transient 0-size during rotate).
		if top < 0 {
			top = 0
		}
		if bottom < 0 {
			bottom = 0
		}
		if top+bottom >= float32(h) {
			top, bottom = 0, 0
		}
		return top, bottom, left, right
	}
	// Fallback: status ~48dp, gesture nav ~24dp @ 420dpi (2.625×).
	if screenW >= 1000 {
		return 126, 63, 0, 0
	}
	return 72, 48, 0, 0
}

// ApplyAndroidSafeInsets sets Document chrome from the content rect (plus any
// extraBottom for a launcher footer). Returns true when chrome values changed.
//
// Soft-IME overlap (extraBottom from [IMEInsetHeight]) replaces the system
// bottom inset when larger — do not sum, or gesture bar is double-counted.
func ApplyAndroidSafeInsets(doc *Document, extraBottom float32, screenW, screenH int32) bool {
	if doc == nil || !AndroidGLES() {
		return false
	}
	top, bottom, _, _ := AndroidContentInsets(screenW, screenH)
	if extraBottom > bottom {
		bottom = extraBottom
	} else if extraBottom > 0 {
		// Small IME report still lifts above the gesture/nav band.
		bottom += extraBottom
	}
	prevTop, prevBottom := doc.ChromeTop(), doc.ChromeBottom()
	doc.SetChromeTop(top)
	doc.SetChromeBottom(bottom)
	return prevTop != top || prevBottom != bottom
}
