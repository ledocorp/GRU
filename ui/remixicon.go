// Package ui (continued) — title-bar Remix glyphs (subset of remixicon.ttf).
//
// Window chrome uses direct codepoints; all other widgets use ui.Icons.Draw
// which resolves the same TTF via remix_icon.go.
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Title-bar chrome glyphs (ri-*-line) from remixicon.css v4.9.1.
const (
	RemixClose                 rune = 0xEB99 // ri-close-line
	RemixSubtract              rune = 0xF1AF // ri-subtract-line (minimize)
	RemixSquare                rune = 0xF3DC // ri-square-line (maximize)
	RemixCheckboxMultipleBlank rune = 0xEB87 // ri-checkbox-multiple-blank-line (restore)
	RemixImageLine           rune = 0xEE4B // ri-image-line
)

var remixTitleBarCPs = []int32{
	int32(RemixClose),
	int32(RemixSubtract),
	int32(RemixSquare),
	int32(RemixCheckboxMultipleBlank),
}

// InitRemixIcon loads title-bar glyphs; called from InitIcons.
func InitRemixIcon(atlasSize int) {
	initRemixIconAtlas(int32(atlasSize))
}

// UnloadRemixIcon releases the shared Remix atlas (also called from Icons.UnloadAll).
func UnloadRemixIcon() {
	unloadRemixIcons()
}

// RemixIconReady reports whether Remix glyphs can be drawn.
func RemixIconReady() bool {
	return remixIcons.ready
}

// RemixIconSummary is a one-line status for startup logging.
func RemixIconSummary() string {
	return remixIconSummary()
}

func drawRemixIcon(dst rl.Rectangle, cp rune, tint rl.Color, strokeScale float32) bool {
	if !remixIcons.ready && !remixIcons.failed {
		initRemixIconAtlas(128)
	}
	// If maximize/restore was skipped at init, try one ensure via codepoint pack.
	if remixIcons.ready {
		remixIcons.mu.Lock()
		_, has := remixIcons.loadedCP[int32(cp)]
		remixIcons.mu.Unlock()
		if !has {
			ensureTitleBarCodepoint(cp)
		}
	}
	remixIcons.mu.Lock()
	font := remixIcons.font
	loaded := remixIcons.loadedCP
	ready := remixIcons.ready
	remixIcons.mu.Unlock()
	if !ready {
		return false
	}
	return remixDrawCodepoint(font, loaded, dst, cp, tint, strokeScale)
}
