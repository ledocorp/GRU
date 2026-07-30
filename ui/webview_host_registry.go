package ui

import "sync"

var (
	webViewHostMu      sync.Mutex
	webViewHostAlive   []WebViewHost
	webViewHostsDirty  bool // bounds moved this frame; Present pumps more Win32 messages
	webViewPumpBudget  int  // scaled up when target FPS drops (keeps COM dispatch rate)
)

func markWebViewHostsDirty() {
	webViewHostsDirty = true
}

var webViewForceBoundsSync bool

// MarkWebViewHostsResize arms a one-shot force-sync pass on the next SyncWebViewHosts
// call. Call from main when the client area changes (edge resize, maximize, DPI).
//
// While WebViewForceBoundsSyncActive, syncHost keeps the host visible and retains
// last COM bounds when layout briefly reports zero size — prevents blank web after resize.
// See docs/WEBVIEW2_HOST.md §13.7 and ARCHITECTURE.md WebView2 section.
func MarkWebViewHostsResize() {
	webViewForceBoundsSync = true
	markWebViewHostsDirty()
}

// WebViewForceBoundsSyncActive reports whether the current SyncWebViewHosts pass
// should retain last COM bounds when visibleHostRect is transiently empty (resize relayout).
func WebViewForceBoundsSyncActive() bool {
	return webViewForceBoundsSync
}

// WebViewHostsActive reports whether any live WebView2 host is registered.
func WebViewHostsActive() bool {
	return activeWebViewHostCount() > 0
}

// UpdateWebViewPresentBudget scales the Win32 message pump when target FPS is low so
// WebView2 COM callbacks still get serviced (~100 dispatches/sec).
func UpdateWebViewPresentBudget(targetFPS int) {
	if activeWebViewHostCount() == 0 {
		webViewHostMu.Lock()
		webViewPumpBudget = 0
		webViewHostMu.Unlock()
		return
	}
	if targetFPS <= 0 {
		targetFPS = WebViewIdleFPS
	}
	const targetDispatchPerSec = 100
	pump := (targetDispatchPerSec + targetFPS - 1) / targetFPS
	if pump < 6 {
		pump = 6
	}
	if pump > 16 {
		pump = 16
	}
	webViewHostMu.Lock()
	webViewPumpBudget = pump
	webViewHostMu.Unlock()
}

func webViewPresentPumpMax(base int) int {
	webViewHostMu.Lock()
	budget := webViewPumpBudget
	webViewHostMu.Unlock()
	if budget > base {
		return budget
	}
	return base
}

// deliverWebViewMessage runs panel OnMessage on the main goroutine (COM posts off-thread).
// FillClient chrome.resize is noted on the COM thread (mutex coalesce) — never QueueMain —
// so resize storms cannot fill the 128-slot buffer and drop end events (webviewhello lock).
func deliverWebViewMessage(payload string, cb func(string)) {
	if noteFillClientChromeResize(payload) {
		return
	}
	if cb == nil {
		return
	}
	if doc := ActiveDocument(); doc != nil {
		msg := payload
		doc.QueueMain(func() { cb(msg) })
		return
	}
	cb(payload)
}

func registerWebViewHost(h WebViewHost) {
	if h == nil {
		return
	}
	webViewHostMu.Lock()
	webViewHostAlive = append(webViewHostAlive, h)
	webViewHostMu.Unlock()
}

func unregisterWebViewHost(h WebViewHost) {
	if h == nil {
		return
	}
	webViewHostMu.Lock()
	for i, alive := range webViewHostAlive {
		if alive == h {
			webViewHostAlive = append(webViewHostAlive[:i], webViewHostAlive[i+1:]...)
			break
		}
	}
	webViewHostMu.Unlock()
}

func activeWebViewHostCount() int {
	webViewHostMu.Lock()
	n := len(webViewHostAlive)
	webViewHostMu.Unlock()
	return n
}

// DestroyAllWebViewHosts tears down every live OS web surface.
// Call before replacing the scene tree (main.go loadScene) so WebView2
// does not keep drawing over the launcher or the next demo.
func DestroyAllWebViewHosts() {
	webViewHostMu.Lock()
	hosts := webViewHostAlive
	webViewHostAlive = nil
	webViewHostMu.Unlock()
	for _, h := range hosts {
		h.Destroy()
	}
}
