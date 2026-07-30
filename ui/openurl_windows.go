//go:build windows

package ui

import (
	"os/exec"
)

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}
