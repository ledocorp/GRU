package ui

import "testing"

func TestWebViewProductionPolicy(t *testing.T) {
	p := WebViewProductionPolicy()
	if p.DevToolsEnabled {
		t.Fatal("production should disable devtools")
	}
	if p.DefaultContextMenusEnabled {
		t.Fatal("production should disable context menus")
	}
	if p.BrowserAcceleratorKeysEnabled {
		t.Fatal("production should disable browser accelerator keys")
	}
	if p.HostObjectsAllowed {
		t.Fatal("production should disable host objects")
	}
}

func TestAllowedWebNavigation(t *testing.T) {
	prod := WebViewProductionPolicy()
	if !allowedWebNavigation(prod, nil, "https://example.com") {
		t.Fatal("https should be allowed")
	}
	if allowedWebNavigation(prod, nil, "http://evil.com") {
		t.Fatal("insecure http should be blocked")
	}
	if !allowedWebNavigation(prod, nil, "http://localhost:8080") {
		t.Fatal("localhost http should be allowed")
	}
	wv := &WebViewPanel{AllowHTTP: true}
	if !allowedWebNavigation(prod, wv, "http://intranet.local/") {
		t.Fatal("panel AllowHTTP should permit http")
	}
}

func TestInitWebViewHostPolicyEnvOverride(t *testing.T) {
	t.Setenv("GRU_WEBVIEW_DEV", "1")
	t.Setenv("GORY_WEBVIEW_DEV", "")
	InitWebViewHostPolicy(true)
	p := WebViewHostPolicyCurrent()
	if !p.DevToolsEnabled {
		t.Fatal("GRU_WEBVIEW_DEV should force dev policy")
	}
	t.Setenv("GRU_WEBVIEW_DEV", "")
	InitWebViewHostPolicy(false)
	if !WebViewHostPolicyCurrent().DevToolsEnabled {
		t.Fatal("dev build should enable devtools")
	}
}

func TestInitWebViewHostPolicyLegacyEnvOverride(t *testing.T) {
	t.Setenv("GRU_WEBVIEW_DEV", "")
	t.Setenv("GORY_WEBVIEW_DEV", "1")
	InitWebViewHostPolicy(true)
	p := WebViewHostPolicyCurrent()
	if !p.DevToolsEnabled {
		t.Fatal("GORY_WEBVIEW_DEV alias should force dev policy")
	}
	t.Setenv("GORY_WEBVIEW_DEV", "")
}

func TestWebViewAcceleratorHandled(t *testing.T) {
	prod := WebViewProductionPolicy()
	if !webViewAcceleratorHandled(0x7B, prod) {
		t.Fatal("F12 should be blocked in production")
	}
	if webViewAcceleratorHandled(0x7B, WebViewDevPolicy()) {
		t.Fatal("F12 should pass through in dev policy")
	}
}
