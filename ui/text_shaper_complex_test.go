package ui



import (

	"testing"



	"github.com/go-text/typesetting/di"

)



func TestComplexScriptArabicShapesRTL(t *testing.T) {

	if !InitShapedFonts() {

		t.Skip("no TTF faces")

	}

	SetTextEngineMode(TextEngineShaped)

	run, ok := shapedShapeText(ComplexScriptArabicSample, 16, false, false, false, false)

	SetTextEngineMode(TextEngineSDF)

	if !ok {

		t.Fatal("shape failed")

	}

	if run.out.Direction != di.DirectionRTL {

		t.Fatalf("direction = %v, want RTL", run.out.Direction)

	}

	w, ok := shapedMeasureTextF(ComplexScriptArabicSample, 16, false, false, false, false)

	if !ok || w <= 0 {

		t.Fatalf("measure = %.1f ok=%v", w, ok)

	}

}

func TestMixedLabelParagraphDirectionLTR(t *testing.T) {
	mixed := "Arabic sample: " + ComplexScriptArabicSample
	dir := shapedTextParagraphDirection([]rune(mixed))
	if dir != di.DirectionLTR {
		t.Fatalf("mixed paragraph direction = %v, want LTR", dir)
	}
	run, ok := shapedShapeText(mixed, 18, false, false, false, false)
	if !ok {
		t.Fatal("shape mixed failed")
	}
	if run.out.Direction != di.DirectionLTR {
		t.Fatalf("shaped direction = %v, want LTR for mixed prefix", run.out.Direction)
	}
}



func TestComplexScriptNotLatinFallback(t *testing.T) {

	if shapedNeedsSDFFallbackMeasure(ComplexScriptArabicSample) {

		t.Fatal("Arabic should use shaped path")

	}

	if !shapedNeedsSDFFallbackMeasure("── box ──") {

		t.Fatal("box drawing should fall back")

	}

}



func TestComplexScriptFixtureRunes(t *testing.T) {

	cp := complexScriptFixtureRunes()

	if len(cp) < 20 {

		t.Fatalf("fixture codepoints = %d", len(cp))

	}

}



func TestComplexScriptIgnoresEditorMono(t *testing.T) {

	if !InitShapedFonts() {

		t.Skip("no TTF faces")

	}

	shapedEnsureMono()

	if !shapedUI.monoReady || shapedUI.mono == nil {

		t.Skip("no mono shaped face")

	}

	regular := pickShapedFace(false, false, false, false)

	run, ok := shapedShapeText("مرحبا", 16, false, false, true, false)

	if !ok || run.out.Face == nil {

		t.Fatal("shape failed")

	}

	if run.out.Face == shapedUI.mono {

		t.Fatal("Arabic must not shape with editor mono face (Fira Code lacks Arabic)")

	}

	if regular != nil && run.out.Face != regular {

		t.Logf("note: shaped face differs from regular UI face (stack may vary)")

	}

	if !shapedTextUsesComplexScript("مرحبا") {

		t.Fatal("expected complex script detection")

	}

	if shapedEffectiveMono(true, "مرحبا") {

		t.Fatal("mono should be disabled for Arabic in editor")

	}

}



func TestShapedArabicUsesDrawPath(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	SetTextEngineMode(TextEngineShaped)
	defer SetTextEngineMode(TextEngineSDF)
	if !textDrawUsesShaped(ComplexScriptArabicSample, 16, false, false, false, false) {
		t.Fatal("Arabic should use shaped draw when engine is shaped")
	}
	run, ok := shapedShapeText(ComplexScriptArabicSample, 16, false, false, false, false)
	if !ok || run.out.Face == nil || len(run.out.Glyphs) == 0 {
		t.Fatalf("shape failed: ok=%v face=%v glyphs=%d", ok, run.out.Face, len(run.out.Glyphs))
	}
}

func TestComplexScriptArabicRasterInk(t *testing.T) {

	if !InitShapedFonts() {

		t.Skip("no TTF faces")

	}

	face := pickShapedFace(false, false, shapedEffectiveMono(true, "مرحبا"), false)

	if face == nil {

		t.Fatal("no face")

	}

	gid, ok := face.NominalGlyph('م')

	if !ok {

		t.Fatal("no glyph for Arabic meem")

	}

	img, ok := rasterShapedGlyphImage(face, gid, 24)

	if !ok || !shapedGlyphCacheHasInk(img) {

		t.Fatal("expected Arabic glyph raster")

	}

}


