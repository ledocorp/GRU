// Package ui (continued) — WebView2 production policy and URL rules.
//
// See docs/WEBVIEW2_HOST.md §4 (security) and §11 (production).
package ui

import (
	"strings"
	"sync"
)

// WebViewHostPolicy controls developer vs production browser capabilities for
// every live WebView2 host. Per-panel overrides use WebViewPanel policy fields.
//
// # LLM Prompt Template
//
//	ui.InitWebViewHostPolicy(releaseBuild)
//	wv := ui.NewWebViewPanel("app", "https://app.example.com", 0, 0, 0, 0)
//	// Optional per-panel dev override:
//	wv.DevToolsEnabled = boolPtr(true)
type WebViewHostPolicy struct {
	// DevToolsEnabled allows Edge DevTools (F12, Ctrl+Shift+I). Off in production.
	DevToolsEnabled bool
	// DefaultContextMenusEnabled allows the native web right-click menu.
	DefaultContextMenusEnabled bool
	// BrowserAcceleratorKeysEnabled forwards browser shortcuts (Ctrl+P, Ctrl+F, …) to the page.
	BrowserAcceleratorKeysEnabled bool
	StatusBarEnabled              bool
	ZoomControlEnabled            bool
	// HostObjectsAllowed exposes COM host objects to script — keep false for untrusted HTML.
	HostObjectsAllowed bool
	AutofillEnabled    bool
	SwipeNavigationEnabled bool
	// AllowInsecureHTTP permits non-localhost http:// navigations when true.
	AllowInsecureHTTP bool
	// AllowFileBundle permits file:// when true (bundle path validation is a follow-up).
	AllowFileBundle bool
}

var (
	webViewPolicyMu sync.RWMutex
	webViewPolicy   = WebViewDevPolicy()
)

// WebViewDevPolicy is the default for non-release builds and GRU_WEBVIEW_DEV=1 (GORY alias).
func WebViewDevPolicy() WebViewHostPolicy {
	return WebViewHostPolicy{
		DevToolsEnabled:               true,
		DefaultContextMenusEnabled:    true,
		BrowserAcceleratorKeysEnabled: true,
		StatusBarEnabled:              true,
		ZoomControlEnabled:            true,
		HostObjectsAllowed:            false,
		AutofillEnabled:               false,
		SwipeNavigationEnabled:        false,
		AllowInsecureHTTP:             false,
		AllowFileBundle:               false,
	}
}

// WebViewProductionPolicy locks down embedded web for shipping builds.
func WebViewProductionPolicy() WebViewHostPolicy {
	return WebViewHostPolicy{
		DevToolsEnabled:               false,
		DefaultContextMenusEnabled:    false,
		BrowserAcceleratorKeysEnabled: false,
		StatusBarEnabled:              false,
		ZoomControlEnabled:            false,
		HostObjectsAllowed:            false,
		AutofillEnabled:               false,
		SwipeNavigationEnabled:        false,
		AllowInsecureHTTP:             false,
		AllowFileBundle:               false,
	}
}

// InitWebViewHostPolicy selects dev vs production defaults.
// GRU_WEBVIEW_DEV=1 (GORY_WEBVIEW_DEV alias) forces dev policy even in release builds.
func InitWebViewHostPolicy(releaseBuild bool) {
	if EnvOr("GRU_WEBVIEW_DEV", "GORY_WEBVIEW_DEV") == "1" {
		SetWebViewHostPolicy(WebViewDevPolicy())
		return
	}
	if releaseBuild {
		SetWebViewHostPolicy(WebViewProductionPolicy())
		return
	}
	SetWebViewHostPolicy(WebViewDevPolicy())
}

// SetWebViewHostPolicy replaces the global WebView2 host policy.
func SetWebViewHostPolicy(p WebViewHostPolicy) {
	webViewPolicyMu.Lock()
	webViewPolicy = p
	webViewPolicyMu.Unlock()
}

// WebViewHostPolicyCurrent returns the active global policy.
func WebViewHostPolicyCurrent() WebViewHostPolicy {
	webViewPolicyMu.RLock()
	p := webViewPolicy
	webViewPolicyMu.RUnlock()
	return p
}

// effectivePolicy merges global policy with optional per-panel overrides.
func (wv *WebViewPanel) effectivePolicy() WebViewHostPolicy {
	if wv == nil {
		return WebViewHostPolicyCurrent()
	}
	p := WebViewHostPolicyCurrent()
	if wv.DevToolsEnabled != nil {
		p.DevToolsEnabled = *wv.DevToolsEnabled
	}
	if wv.DefaultContextMenusEnabled != nil {
		p.DefaultContextMenusEnabled = *wv.DefaultContextMenusEnabled
	}
	if wv.BrowserAcceleratorKeysEnabled != nil {
		p.BrowserAcceleratorKeysEnabled = *wv.BrowserAcceleratorKeysEnabled
	}
	if wv.AllowHTTP {
		p.AllowInsecureHTTP = true
	}
	if wv.AllowFileBundle {
		p.AllowFileBundle = true
	}
	return p
}

// allowedWebNavigation reports whether url may be loaded under policy + panel flags.
func allowedWebNavigation(policy WebViewHostPolicy, wv *WebViewPanel, url string) bool {
	lower := strings.ToLower(strings.TrimSpace(url))
	if strings.HasPrefix(lower, "https://") {
		return true
	}
	if strings.HasPrefix(lower, "http://localhost") || strings.HasPrefix(lower, "http://127.0.0.1") {
		return true
	}
	allowHTTP := policy.AllowInsecureHTTP || (wv != nil && wv.AllowHTTP)
	if allowHTTP && strings.HasPrefix(lower, "http://") {
		return true
	}
	allowFile := policy.AllowFileBundle || (wv != nil && wv.AllowFileBundle)
	if allowFile && strings.HasPrefix(lower, "file://") {
		return true
	}
	return false
}

// webViewAcceleratorHandled consumes devtools shortcuts when policy disables them.
func webViewAcceleratorHandled(vk uint, policy WebViewHostPolicy) bool {
	if policy.DevToolsEnabled {
		return false
	}
	const vkF12 = 0x7B
	return vk == vkF12
}
