package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// textDrawUsesShaped reports whether drawTextF would take the shaped glyph path
// for this string (Latin/mono editor text uses SDF — see shapedDrawTextF).
func textDrawUsesShaped(text string, fontSize float32, bold, italic, mono, preview bool) bool {
	if text == "" {
		return false
	}
	if !shapedTextReady() {
		return false
	}
	if shapedNeedsSDFFallbackMeasure(text) {
		return false
	}
	if shapedGlyphCacheEnabled() {
		return true
	}
	return shapedTextUsesComplexScript(text)
}

// shapedDrawTextF rasterizes complex-script text per shaped glyph (T1.2/T1.5).
// Latin and box-drawing strings return false so drawTextF uses the legacy whole-string
// SDF path — HarfBuzz per-glyph advances do not match raylib atlas glyph metrics and
// produced visible random letter gaps when mixed with shaped measure.
func shapedDrawTextF(text string, x, y, fontSize float32, color rl.Color, bold, italic, mono, preview bool) bool {
	if text == "" {
		return true
	}
	if shapedNeedsSDFFallbackMeasure(text) {
		return false
	}
	useCache := shapedGlyphCacheEnabled()
	if !useCache && !shapedTextUsesComplexScript(text) {
		return false
	}
	run, ok := shapedShapeText(text, fontSize, bold, italic, mono, preview)
	if !ok || len(run.runes) == 0 {
		return ok
	}

	// Complex script: raster shaped GlyphIDs via TTF outlines (T1.8). Per-rune SDF
	// from editor mono atlas cannot render Arabic (Notepad uses Mono=true).
	if useCache || shapedTextUsesComplexScript(text) {
		return shapedDrawRunGlyphCache(run, x, y, fontSize, color)
	}
	return false
}
