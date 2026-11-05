package ui

import (
	"image"
	"image/color"
	"testing"
)

func TestBuildMaskGammaLUTMonotonicMid(t *testing.T) {
	table := buildMaskGammaLUT(maskGammaSrcDark, maskGammaContrastDefault, 0)
	if table[0] != 0 {
		t.Fatalf("table[0]=%d want 0", table[0])
	}
	if table[255] != 255 {
		t.Fatalf("table[255]=%d want 255", table[255])
	}
	ident := buildMaskGammaLUT(maskGammaSrcDark, 0, 0)
	// Contrast must change the correcting curve vs contrast=0.
	diff := false
	for i := 1; i < 255; i++ {
		if table[i] != ident[i] {
			diff = true
			break
		}
	}
	if !diff {
		t.Fatal("Chrome-class contrast LUT should differ from contrast 0")
	}
	// 1.0 should push mid coverage harder than legacy 0.5.
	legacy := buildMaskGammaLUT(maskGammaSrcDark, 0.5, 0)
	if table[128] <= legacy[128] {
		t.Fatalf("contrast 1.0 mid=%d should exceed 0.5 mid=%d", table[128], legacy[128])
	}
}

func TestEnhanceS1StrikeAtlasGreyscale(t *testing.T) {
	atlas := image.NewRGBA(image.Rect(0, 0, 1, 1))
	atlas.SetRGBA(0, 0, color.RGBA{255, 255, 255, 128})
	before := atlas.Pix[3]
	enhanceS1StrikeAtlas(atlas, false)
	if atlas.Pix[3] == before && s1MaskGammaContrast() > 0 {
		t.Fatalf("enhance should remap mid alpha (before=%d after=%d)", before, atlas.Pix[3])
	}
}

func TestS1MaskGammaContrastClamp(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_CONTRAST", "")
		t.Setenv("GORY_TEXT_CONTRAST", "")
	})
	t.Setenv("GRU_TEXT_CONTRAST", "")
	if s1MaskGammaContrast() != maskGammaContrastDefault {
		t.Fatalf("default contrast=%v", s1MaskGammaContrast())
	}
	t.Setenv("GRU_TEXT_CONTRAST", "0.8")
	if s1MaskGammaContrast() != 0.8 {
		t.Fatalf("want 0.8 got %v", s1MaskGammaContrast())
	}
}
