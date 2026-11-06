//go:build linux && !android

package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// S1HostWindowFlags is the lean present config for Linux (X11 and Wayland).
// No FlagMsaa4xHint: MSAA resolve softens 1× sticky text — Windows keeps MSAA
// via s1_window_flags_other.go.
func S1HostWindowFlags() uint32 {
	return uint32(rl.FlagVsyncHint |
		rl.FlagWindowResizable | rl.FlagWindowUndecorated | rl.FlagWindowHidden)
}
