package ui

import (
	"testing"

	"github.com/go-text/typesetting/di"
)

func TestShapedGlyphRasterSSAA(t *testing.T) {
	prev := BaseSupersamplingScale
	prevDPI := DisplayScale
	BaseSupersamplingScale = 2
	DisplayScale = 1
	defer func() {
		BaseSupersamplingScale = prev
		DisplayScale = prevDPI
	}()
	if shapedGlyphRasterPx(16) != 32 {
		t.Fatalf("raster px = %v, want 32 at 2x SSAA", shapedGlyphRasterPx(16))
	}
}

func TestShapedGlyphOriginInTexture(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	face := pickShapedFace(false, false, false, false)
	if face == nil {
		t.Fatal("no face")
	}
	gid, ok := face.NominalGlyph('A')
	if !ok {
		t.Fatal("no glyph")
	}
	img, ok := rasterShapedGlyphImage(face, gid, 16)
	if !ok || !shapedGlyphCacheHasInk(img) {
		t.Fatal("raster failed")
	}
}

func TestShapedRunInkWidthAtLeastAdvance(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	run, ok := shapedShapeText(ComplexScriptArabicSample, 18, false, false, false, false)
	if !ok {
		t.Fatal("shape failed")
	}
	ink, ok := shapedRunVisualWidth(run, 18)
	if !ok || ink <= 0 {
		t.Fatalf("ink width = %.2f ok=%v", ink, ok)
	}
	adv := float32(run.out.Advance.Ceil())
	if ink < adv*0.5 || ink > adv*1.5 {
		t.Fatalf("ink %.2f outside reasonable range of advance %.2f", ink, adv)
	}
}

func TestShapedParagraphDirectionPureArabic(t *testing.T) {
	dir := shapedTextParagraphDirection([]rune(ComplexScriptArabicSample))
	if dir != di.DirectionRTL {
		t.Fatalf("direction = %v, want RTL", dir)
	}
}

func TestShapedAtlasAlloc(t *testing.T) {
	initShapedGlyphAtlas()
	slot, ok := shapedAtlasAlloc(32, 32)
	if !ok {
		t.Fatal("alloc failed")
	}
	if slot.Dx() != 32 || slot.Dy() != 32 {
		t.Fatalf("slot = %v", slot)
	}
}
