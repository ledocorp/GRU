//go:build !linux

package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// S1HostWindowFlags is the Windows/desktop S1 host config: MSAA + vsync +
// borderless chrome. Linux omits MSAA (see s1_window_flags_linux.go).
func S1HostWindowFlags() uint32 {
	return uint32(rl.FlagMsaa4xHint | rl.FlagVsyncHint |
		rl.FlagWindowResizable | rl.FlagWindowUndecorated | rl.FlagWindowHidden)
}
