//go:build cgo && !gru_nofx

package ui

import "github.com/ledocorp/gru/ui/fxglyph"

func tryFxBoostStrikeAlpha(pix []byte, contrast float32) {
	if contrast <= 0 || !fxglyph.Enabled() || fxOptOut("GRU_FX_GLYPH", "GORY_FX_GLYPH") {
		return
	}
	_ = fxglyph.BoostRGBAAlpha(pix, contrast)
}
