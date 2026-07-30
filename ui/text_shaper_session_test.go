package ui

import "testing"

func TestShapedTextSessionRestoresSDF(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	prev := TextEngineModeActive()
	sess, ok := BeginShapedTextSession()
	if !ok {
		t.Fatal("BeginShapedTextSession failed")
	}
	if TextEngineModeActive() != TextEngineShaped {
		t.Fatalf("mode = %v, want shaped", TextEngineModeActive())
	}
	sess.End()
	if TextEngineModeActive() != prev {
		t.Fatalf("mode after End = %v, want %v", TextEngineModeActive(), prev)
	}
}

func TestShapedSessionLatinMeasureStaysSDF(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	sess, ok := BeginShapedTextSession()
	if !ok {
		t.Fatal("session failed")
	}
	defer sess.End()
	SetTextEngineMode(TextEngineShaped)
	if _, ok := shapedMeasureTextF("Hello", 16, false, false, false, false); ok {
		t.Fatal("Latin measure should stay SDF during demo session")
	}
}

func TestDevanagariFaceWhenAvailable(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	initShapedScriptFaces()
	if shapedDevanagari == nil {
		t.Skip("no Devanagari font on this machine")
	}
	run, ok := shapedShapeText(ComplexScriptDevanagariSample, 16, false, false, false, false)
	if !ok || run.out.Face != shapedDevanagari {
		t.Fatal("Devanagari sample should shape with Devanagari face")
	}
	gid, ok := shapedDevanagari.NominalGlyph('न')
	if !ok {
		t.Fatal("missing Devanagari glyph")
	}
	img, ok := rasterShapedGlyphImage(shapedDevanagari, gid, 24)
	if !ok || !shapedGlyphCacheHasInk(img) {
		t.Fatal("Devanagari raster failed")
	}
}
