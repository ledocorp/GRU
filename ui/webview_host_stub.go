//go:build !windows

package ui

import rl "github.com/gen2brain/raylib-go/raylib"

func webViewHostPlatformSupported() bool { return false }

func afterSyncWebViewHosts() {
	webViewForceBoundsSync = false
}

type webViewHostStub struct {
	status WebViewHostStatus
	errMsg string
}

func newWebViewHostPlatform(_ uintptr) WebViewHost {
	return &webViewHostStub{status: WebViewHostUnsupported}
}

func (h *webViewHostStub) Navigate(string) error     { return nil }
func (h *webViewHostStub) SetHTML(string) error    { return nil }
func (h *webViewHostStub) SetBounds(rl.Rectangle) error { return nil }
func (h *webViewHostStub) SetVisible(bool)           {}
func (h *webViewHostStub) SetFocus()                 {}
func (h *webViewHostStub) Blur()                     {}
func (h *webViewHostStub) PostMessage(string) error  { return nil }
func (h *webViewHostStub) Destroy()                  {}
func (h *webViewHostStub) Status() WebViewHostStatus { return h.status }
func (h *webViewHostStub) LastError() string         { return h.errMsg }

func (h *webViewHostStub) bindOnMessage(func(string)) {}
