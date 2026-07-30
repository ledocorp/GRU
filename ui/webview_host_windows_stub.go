//go:build windows && !webview2

package ui

import rl "github.com/gen2brain/raylib-go/raylib"

func webViewHostPlatformSupported() bool {
	// Live WebView2 loader ships behind -tags webview2.
	return true
}

func afterSyncWebViewHosts() {
	webViewForceBoundsSync = false
}

type webViewHostStub struct {
	status WebViewHostStatus
	errMsg string
}

func newWebViewHostPlatform(_ uintptr) WebViewHost {
	if !webViewHostPlatformSupported() {
		return &webViewHostStub{status: WebViewHostUnsupported}
	}
	return &webViewHostStub{status: WebViewHostStub, errMsg: "WebView2 loader not linked — build with -tags webview2 (B+2.2)"}
}

func (h *webViewHostStub) Navigate(string) error        { return nil }
func (h *webViewHostStub) SetHTML(string) error       { return nil }
func (h *webViewHostStub) SetBounds(rl.Rectangle) error { return nil }
func (h *webViewHostStub) SetVisible(bool)            {}
func (h *webViewHostStub) SetFocus()                  {}
func (h *webViewHostStub) Blur()                      {}
func (h *webViewHostStub) PostMessage(string) error   { return nil }
func (h *webViewHostStub) Destroy()                   {}
func (h *webViewHostStub) Status() WebViewHostStatus  { return h.status }
func (h *webViewHostStub) LastError() string          { return h.errMsg }

func (h *webViewHostStub) bindOnMessage(func(string)) {}
