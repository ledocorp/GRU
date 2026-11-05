//go:build !cgo || gru_nofx

package ui

import (
	"image"

	"github.com/go-text/typesetting/font"
)

func tryRasterShapedGlyphFX(outline font.GlyphOutline, scale float64, w, h int, ox, oy float64) (image.Image, bool) {
	return nil, false
}

func fxGlyphBackendActive() bool { return false }
