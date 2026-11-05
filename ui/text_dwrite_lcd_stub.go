//go:build !windows || !cgo

package ui

import "image"

func dwriteLCDAvailable() bool { return false }

var (
	dwriteBlendGamma     float32 = 1.8
	dwriteBlendContrast  float32 = 0.5
	dwriteClearTypeLevel float32 = 0.5
)

func packDWriteLCDStrikeAtlas(pixelSize int32, originX float32, fontPath string, symmetric bool) (*image.RGBA, map[rune]ftGreyGlyph, bool) {
	_ = pixelSize
	_ = originX
	_ = fontPath
	_ = symmetric
	return nil, nil, false
}

func packDWriteGreyStrikeAtlas(pixelSize int32, fontPath string) (*image.RGBA, map[rune]ftGreyGlyph, bool) {
	_ = pixelSize
	_ = fontPath
	return nil, nil, false
}
