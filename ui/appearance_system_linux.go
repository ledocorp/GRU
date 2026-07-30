//go:build linux

package ui

import (
	"os/exec"
	"strings"
)

// systemPrefersDarkAppearance reads GNOME color-scheme when gsettings is available.
func systemPrefersDarkAppearance() bool {
	if out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "color-scheme").Output(); err == nil {
		s := strings.ToLower(strings.TrimSpace(string(out)))
		if strings.Contains(s, "dark") {
			return true
		}
		if strings.Contains(s, "light") {
			return false
		}
	}
	if out, err := exec.Command("gsettings", "get", "org.gnome.desktop.interface", "gtk-theme").Output(); err == nil {
		s := strings.ToLower(string(out))
		return strings.Contains(s, "dark")
	}
	return false
}
