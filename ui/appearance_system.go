//go:build (!windows && !linux) || android

package ui

func systemPrefersDarkAppearance() bool {
	return false
}
