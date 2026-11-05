package ui

import (
	"fmt"

	"github.com/ledocorp/gru/ui/fxglyph"
	"github.com/ledocorp/gru/ui/fxmeasure"
)

// FxDogfoodHUD is a one-line status for Editor Wrap / Text Engine dogfood scenes.
// fx is native in CGO builds; GRU_FX_*=0 opts out for A/B testing.
func FxDogfoodHUD() string {
	wrap := EditorWrapPathName()
	glyph := fxGlyphDogfoodState()
	link := "native"
	if !fxmeasure.Enabled() && !fxglyph.Enabled() {
		link = "stub"
	}
	return fmt.Sprintf("fx_link=%s wrap=%s glyph=%s", link, wrap, glyph)
}

func fxGlyphDogfoodState() string {
	if fxGlyphBackendActive() {
		return "on"
	}
	if fxOptOut("GRU_FX_GLYPH", "GORY_FX_GLYPH") {
		return "gg-fallback"
	}
	if !fxglyph.Enabled() {
		return "stub"
	}
	return "gg"
}
