// Package ui (continued) — WebView2 host abstraction (B+2).
//
// Default builds use a stub host; Windows live WebView2 arrives behind -tags webview2.
// See docs/WEBVIEW2_HOST.md.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// WebViewHostStatus reports platform host availability for a WebViewPanel.
type WebViewHostStatus string

const (
	WebViewHostUnsupported WebViewHostStatus = "unsupported" // OS/build cannot host WebView2
	WebViewHostStub        WebViewHostStatus = "stub"        // contract wired; no live browser yet
	WebViewHostReady       WebViewHostStatus = "ready"       // live host attached
	WebViewHostError       WebViewHostStatus = "error"
)

// WebViewHost is the OS child surface behind WebViewPanel.
type WebViewHost interface {
	Navigate(url string) error
	SetHTML(html string) error
	SetBounds(r rl.Rectangle) error
	SetVisible(visible bool)
	SetFocus()
	Blur()
	PostMessage(json string) error
	Destroy()
	Status() WebViewHostStatus
	LastError() string
}

// WebViewHostSupported reports whether this build/OS can attach a live WebView2 host.
func WebViewHostSupported() bool {
	return webViewHostPlatformSupported()
}

// NewWebViewHost creates a platform host for parentHWND (raylib window handle on Windows).
// Returns stub host when live WebView2 is not linked.
func NewWebViewHost(parentHWND uintptr) WebViewHost {
	return newWebViewHostPlatform(parentHWND)
}

// SyncWebViewHosts walks doc.Root and syncs every WebViewPanel OS host: bounds
// (visibleHostRect), visibility, navigation (once via contentSynced), theme bridge,
// and opaque pointer policy (never WS_EX_TRANSPARENT while visible).
//
// Call after layout when any web panel may have moved, been clipped, or occluded.
// Pair with PresentWebViewHosts after rl.EndDrawing. See docs/WEBVIEW2_HOST.md §13.6.
//
// Example (main loop):
//
//	if layoutDirty { doc.Root.Layout() }
//	ui.SyncWebViewHosts(doc)
//	ui.RouteScenePointerFocus(doc)
//	// ... draw ...
//	rl.EndDrawing()
//	ui.PresentWebViewHosts()
func SyncWebViewHosts(doc *Document) {
	if doc == nil || doc.Root == nil {
		return
	}
	syncWebViewHostsWalk(doc.Root, doc)
	afterSyncWebViewHosts()
}

// PresentWebViewHosts finishes the frame for every live WebView2 host: Show(),
// bounds-scoped HWND raise (or defer when WebViewDeferChromeRaise), demote stray
// Chrome children, and pump Win32 messages on the parent HWND for COM dispatch.
//
// Must run immediately after rl.EndDrawing() so the HWND stacks above the GL swap chain.
// See docs/WEBVIEW2_HOST.md §13.5–§13.6.
func PresentWebViewHosts() {
	presentWebViewHostsPlatform()
}

func syncWebViewHostsWalk(n Node, doc *Document) {
	if n == nil {
		return
	}
	if wv, ok := n.(*WebViewPanel); ok {
		wv.syncHost(doc)
	}
	for _, ch := range n.Children() {
		syncWebViewHostsWalk(ch, doc)
	}
}
