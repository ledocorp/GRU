package ui

import (
	"math"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestTextEngineBackendNameSDFDefault(t *testing.T) {
	SetTextEngineMode(TextEngineSDF)
	name := TextEngineBackendName()
	if name != UIFontBackend() {
		t.Fatalf("backend = %q, want %q", name, UIFontBackend())
	}
}

func TestTextEngineShapedMeasureReady(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces for shaped engine")
	}
	SetTextEngineMode(TextEngineShaped)
	if !shapedTextReady() {
		t.Fatal("shaped measure should be ready after InitShapedFonts")
	}
	w, ok := shapedMeasureTextF("Ag", 16, false, false, false, false)
	if !ok || w <= 0 {
		t.Fatalf("shapedMeasureTextF = %.2f ok=%v", w, ok)
	}
	SetTextEngineMode(TextEngineSDF)
}

func TestShapedMeasureRatioApplication(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	shapedMeasureRatio[16] = 0.748
	shapedMeasureRatioReady = true
	shapedPreviewMeasureRatio[16] = 0.712
	shapedPreviewMeasureRatioReady = true
	raw, ok := shapedMeasureTextFRaw("Hello world", 16, false, false, false, false)
	if !ok || raw <= 0 {
		t.Fatalf("raw = %.2f ok=%v", raw, ok)
	}
	got, ok := shapedMeasureTextF("Hello world", 16, false, false, false, false)
	if !ok {
		t.Fatal("shapedMeasureTextF failed")
	}
	want := raw * 0.748
	if math.Abs(float64(got-want)) > 0.01 {
		t.Fatalf("calibrated = %.2f want %.2f", got, want)
	}
	if _, ok := shapedMeasureTextF("Hello world", 16, false, false, false, true); ok {
		t.Fatal("Latin preview should bypass shaped measure (T1.7)")
	}
}

func TestRichTextPreviewStyleUsesSDFMeasureWhenShaped(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	if !sdfReady {
		t.Skip("SDF font not loaded in test environment")
	}
	rt := NewRichText("rt-test", []TextSpan{{Text: "Preview body", Style: "richtext-preview-body"}}, 0, 0, 400, 0)
	rt.styleName = "richtext-preview"
	SetTextEngineMode(TextEngineShaped)
	st := rt.spanStyle(rt.Spans[0])
	if !st.PreviewFont {
		t.Fatal("preview span should set PreviewFont")
	}
	if _, ok := shapedMeasureTextF("Preview body", EffectiveFontSize(st), styleDrawBold(st), st.Italic, st.Mono, true); ok {
		t.Fatal("Latin preview should bypass shaped measure (T1.7)")
	}
	w := measureTextS("Preview body", st)
	if w <= 0 {
		t.Fatalf("preview measure width = %d", w)
	}
	SetTextEngineMode(TextEngineSDF)
}

func TestShapedMeasureCacheHit(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	shapedMeasureCacheClear()
	shapedMeasureRatio[16] = 0.75
	shapedMeasureRatioReady = true
	w1, ok := shapedMeasureTextF("cache probe", 16, false, false, false, false)
	if !ok || w1 <= 0 {
		t.Fatalf("first measure = %.2f ok=%v", w1, ok)
	}
	w2, ok := shapedMeasureTextF("cache probe", 16, false, false, false, false)
	if !ok || w2 != w1 {
		t.Fatalf("cache hit = %.2f want %.2f", w2, w1)
	}
}

func TestShapedPreviewFacesOptional(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	if !initShapedPreviewFaces() {
		t.Skip("no Inter preview fonts on disk")
	}
	if shapedPreview.regular == nil {
		t.Fatal("preview regular face missing")
	}
}

func TestShapedMeasureSDFFallbackSymbols(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	if !shapedNeedsSDFFallbackMeasure("── section ──") {
		t.Fatal("box-drawing should fall back to SDF measure")
	}
	if shapedNeedsSDFFallbackMeasure("Hello world") {
		t.Fatal("Latin probe should use shaped measure")
	}
	_, ok := shapedMeasureTextF("── section ──", 16, false, false, false, false)
	if ok {
		t.Fatal("box-drawing should return ok=false for shaped measure")
	}
}

func TestShapedMeasureMonotonic(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	short, _ := shapedMeasureTextFRaw("Hi", 16, false, false, false, false)
	long, _ := shapedMeasureTextFRaw("Hello world", 16, false, false, false, false)
	if long <= short {
		t.Fatalf("longer string should measure wider: short=%.1f long=%.1f", short, long)
	}
}

func TestShapedShapeTextGlyphCount(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	run, ok := shapedShapeText("Hello", 16, false, false, false, false)
	if !ok {
		t.Fatal("shapedShapeText failed")
	}
	if len(run.out.Glyphs) == 0 {
		t.Fatal("expected shaped glyphs")
	}
	if len(run.runes) != 5 {
		t.Fatalf("runes = %d, want 5", len(run.runes))
	}
}

func TestShapedDrawFallbackSymbols(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	SetTextEngineMode(TextEngineShaped)
	if shapedDrawTextF("── section ──", 0, 0, 16, rl.White, false, false, false, false) {
		t.Fatal("box-drawing draw should fall back")
	}
	SetTextEngineMode(TextEngineSDF)
}

func TestShapedDrawLatinFallsBackToSDFString(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	if shapedDrawTextF("Hello world", 0, 0, 16, rl.White, false, false, false, false) {
		t.Fatal("Latin draw should use whole-string SDF, not per-glyph placement")
	}
}

func TestEditorMeasureWidthMatchesSDFDrawWhenLatinShaped(t *testing.T) {
	if !InitShapedFonts() || !sdfReady {
		t.Skip("fonts not loaded")
	}
	SetTextEngineMode(TextEngineShaped)
	st := Style{FontSize: 16, Mono: true}
	text := "func main() {"
	got := EditorMeasureWidth(text, st)
	want := measureTextS(text, st)
	SetTextEngineMode(TextEngineSDF)
	if int32(got) != want {
		t.Fatalf("EditorMeasureWidth = %.0f, measureTextS = %d", got, want)
	}
}

func TestShapedCaretStopsASCII(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	SetTextEngineMode(TextEngineShaped)
	stops := shapedCaretStops("abc", 16, false, false, false, false)
	SetTextEngineMode(TextEngineSDF)
	if len(stops) < 2 {
		t.Fatalf("stops = %v", stops)
	}
	if stops[0] != 0 || stops[len(stops)-1] != 3 {
		t.Fatalf("stops = %v, want [0..3]", stops)
	}
}

func TestShapedCaretOffsetAtX(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	SetTextEngineMode(TextEngineShaped)
	s := GetThemeStyle("input")
	s.Mono = true
	off, ok := shapedCaretOffsetAtX("Hello", 0, s)
	if !ok || off != 0 {
		t.Fatalf("offset at 0 = %d ok=%v", off, ok)
	}
	SetTextEngineMode(TextEngineSDF)
}

func TestShapedDrawEmptyOK(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("no TTF faces")
	}
	if !shapedDrawTextF("", 0, 0, 16, rl.White, false, false, false, false) {
		t.Fatal("empty shaped draw should succeed")
	}
}
