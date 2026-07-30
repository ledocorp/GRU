//go:build !windows && !android

package ui

import (
	"os/exec"
	"runtime"
)

// OpenBrowser opens url in the system default browser.
func OpenBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	default:
		return exec.Command("xdg-open", url).Start()
	}
}
