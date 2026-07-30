// Package ui (continued) — WebView2 flash/passthrough diagnostics.
package ui

import (
	"fmt"
	"sync"
	"time"
)

// WebViewDebugEnabled logs passthrough, visibility, and present transitions to stderr.
// Set GRU_WEBVIEW_DEBUG=1 (GORY_WEBVIEW_DEBUG alias) or press F12 in the demo launcher.
var WebViewDebugEnabled bool

var (
	webViewDebugMu   sync.Mutex
	webViewDebugLast time.Time
)

func logWebViewDebug(tag, detail string) {
	if !WebViewDebugEnabled {
		return
	}
	webViewDebugMu.Lock()
	now := time.Now()
	if now.Sub(webViewDebugLast) < 40*time.Millisecond {
		webViewDebugMu.Unlock()
		return
	}
	webViewDebugLast = now
	webViewDebugMu.Unlock()
	fmt.Printf("Gru webview: %s %s\n", tag, detail)
}
