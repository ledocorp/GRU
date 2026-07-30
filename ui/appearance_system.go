//go:build !windows && !linux

package ui

func systemPrefersDarkAppearance() bool {
	return false
}
