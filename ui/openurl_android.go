//go:build android

// Android-only OpenBrowser stub. Desktop: openurl_windows.go / openurl_other.go.
// See docs/ANDROID_CODE.md §3.
package ui

import "errors"

// OpenBrowser is not wired on Android v1 (no xdg-open). Preview links are no-op.
func OpenBrowser(url string) error {
	_ = url
	return errors.New("OpenBrowser unavailable on Android")
}
