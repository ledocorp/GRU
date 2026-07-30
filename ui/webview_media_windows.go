//go:build windows && webview2

package ui

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"

	"github.com/wailsapp/go-webview2/pkg/edge"
)

// WebViewMediaVHost is the HTTPS virtual host mapped to the repo assets/ folder.
const WebViewMediaVHost = "gru.media"

// ErrWebViewMediaVHostPending means the WebView2 controller is not ready for
// SetVirtualHostNameToFolderMapping yet — retry on the next syncHost pass.
var ErrWebViewMediaVHostPending = errors.New("webview media vhost pending")

var (
	webViewMediaVHostMu sync.Mutex
	webViewMediaVHostOK = map[uintptr]bool{}
)

func chromiumInstanceKey(ch *edge.Chromium) uintptr {
	if ch == nil {
		return 0
	}
	return uintptr(unsafe.Pointer(ch))
}

// EnsureWebViewMediaVHost maps https://gru.media/ → assets/ for local bundles.
// Each live Chromium instance needs its own mapping (not a process-global once).
func EnsureWebViewMediaVHost(ch *edge.Chromium) error {
	if ch == nil {
		return ErrWebViewMediaVHostPending
	}
	key := chromiumInstanceKey(ch)
	webViewMediaVHostMu.Lock()
	if webViewMediaVHostOK[key] {
		webViewMediaVHostMu.Unlock()
		return nil
	}
	webViewMediaVHostMu.Unlock()

	if ch.GetController() == nil {
		return ErrWebViewMediaVHostPending
	}
	wv3 := ch.GetICoreWebView2_3()
	if wv3 == nil {
		return ErrWebViewMediaVHostPending
	}
	root := resolveAssetsRootForWebView()
	if err := wv3.SetVirtualHostNameToFolderMapping(
		WebViewMediaVHost,
		root,
		edge.COREWEBVIEW2_HOST_RESOURCE_ACCESS_KIND_ALLOW,
	); err != nil {
		return err
	}
	webViewMediaVHostMu.Lock()
	webViewMediaVHostOK[key] = true
	webViewMediaVHostMu.Unlock()
	return nil
}

func ensureMediaVHostForNavigate(ch *edge.Chromium, url string) error {
	if ch == nil {
		return ErrWebViewMediaVHostPending
	}
	lower := strings.ToLower(strings.TrimSpace(url))
	if !strings.Contains(lower, WebViewMediaVHost) {
		return nil
	}
	return EnsureWebViewMediaVHost(ch)
}

func resolveAssetsRootForWebView() string {
	for _, base := range assetRootCandidates() {
		if st, err := os.Stat(base); err == nil && st.IsDir() {
			if abs, err := filepath.Abs(base); err == nil {
				return abs
			}
			return base
		}
	}
	return filepath.Join("assets")
}

func assetRootCandidates() []string {
	var out []string
	out = append(out, "assets")
	if wd, err := os.Getwd(); err == nil {
		out = append(out, filepath.Join(wd, "assets"), filepath.Join(wd, "..", "assets"))
	}
	if exe, err := os.Executable(); err == nil {
		out = append(out, filepath.Join(filepath.Dir(exe), "assets"))
	}
	return out
}
