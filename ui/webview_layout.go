// Package ui (continued) — WebView HWND layout jump handling (maximize, restore).
package ui

import (
	"sync"
	"time"
)

var (
	webViewLayoutJumpMu    sync.Mutex
	webViewLayoutJumpUntil time.Time
)

// NotifyWebViewLayoutJump marks every live host for a bounds refresh and keeps
// resize FPS elevated (maximize button, title-bar double-click, or zoom edge).
// Does not hide hosts — hiding caused stale HWND overlays and blank video until
// a client-size edge fired; see ARCHITECTURE.md §11.4 and WEBVIEW2_HOST.md §9.
func NotifyWebViewLayoutJump() {
	MarkWebViewHostsResize()
	webViewLayoutJumpMu.Lock()
	webViewLayoutJumpUntil = time.Now().Add(900 * time.Millisecond)
	webViewLayoutJumpMu.Unlock()
}

// WebViewLayoutJumpActive reports whether a recent OS window geometry jump should
// keep resize FPS and force WebView sync (main loop).
func WebViewLayoutJumpActive(now time.Time) bool {
	webViewLayoutJumpMu.Lock()
	until := webViewLayoutJumpUntil
	webViewLayoutJumpMu.Unlock()
	return !until.IsZero() && now.Before(until)
}
