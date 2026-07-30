// Package ui — text shaping backend (Engine T1).
//
// T1 replaces the SDF-only measure/draw path with shaped glyphs via
// go-text/typesetting, keeping drawTextF/measureTextF contracts stable.
// See docs/T1_TEXT_ENGINE.md.
//
// LLM Prompt Template:
//   "Read docs/T1_TEXT_ENGINE.md and ui/text_shaper.go. Implement the next
//    T1.x package without breaking UIFontBackend() sdf fallback."
package ui

// TextEngineMode selects the active text raster backend.
type TextEngineMode int

const (
	// TextEngineSDF uses the raylib SDF atlas (default until T1.6).
	TextEngineSDF TextEngineMode = iota
	// TextEngineShaped uses go-text/typesetting shaped glyphs (T1 target).
	TextEngineShaped
)

// shapedLatinMeasureActive gates shaped measure for Latin strings. Notepad/editor
// needs true for caret parity; Text Engine demo session sets false so chrome stays SDF.
var shapedLatinMeasureActive = true

// TextEngineModeActive reports the configured engine mode.
func TextEngineModeActive() TextEngineMode {
	return textEngineMode
}

// SetTextEngineMode switches the backend (tests and T1 rollout).
// Shaped measure (T1.1) and shaped draw (T1.2) activate when faces are loaded.
func SetTextEngineMode(mode TextEngineMode) {
	textEngineMode = mode
	shapedMeasureCacheClear()
}

// TextEngineBackendName extends UIFontBackend with T1 rollout state.
func TextEngineBackendName() string {
	switch textEngineMode {
	case TextEngineShaped:
		if shapedGlyphCacheEnabled() {
			return "shaped+glyph-cache"
		}
		if shapedTextReady() {
			return "shaped+outline-glyphs"
		}
		if sdfReady {
			return "shaped+atlas-glyphs"
		}
		return "shaped"
	default:
		return UIFontBackend()
	}
}

// shapedTextReady reports whether shaped measure/draw is wired (T1.1+).
func shapedTextReady() bool {
	return textEngineMode == TextEngineShaped && shapedUI.ready
}

// textEngineMode is the global backend. Notepad release (-tags notepad) flips to
// TextEngineShaped in main.applyTextEngineMode (T1.6-R); dev builds stay SDF unless
// GRU_SHAPED_TEXT=1 (GORY_SHAPED_TEXT alias).
var textEngineMode = TextEngineSDF

// ShapedTextSession scopes shaped mode to one scene (Text Engine demo). End() restores
// the previous backend so other demos keep normal SDF text.
type ShapedTextSession struct {
	prev             TextEngineMode
	prevLatinMeasure bool
	active           bool
}

// BeginShapedTextSession enables shaped complex-script draw for the current scene only.
func BeginShapedTextSession() (ShapedTextSession, bool) {
	if !InitShapedFonts() {
		return ShapedTextSession{}, false
	}
	initShapedScriptFaces()
	s := ShapedTextSession{
		prev:             textEngineMode,
		prevLatinMeasure: shapedLatinMeasureActive,
		active:           true,
	}
	shapedLatinMeasureActive = false
	SetTextEngineMode(TextEngineShaped)
	return s, true
}

// End restores the text backend active before BeginShapedTextSession.
func (s *ShapedTextSession) End() {
	if !s.active {
		return
	}
	SetTextEngineMode(s.prev)
	shapedLatinMeasureActive = s.prevLatinMeasure
	unloadShapedGlyphCache()
	initShapedGlyphCache()
	s.active = false
}

// Active reports whether the session is still open.
func (s *ShapedTextSession) Active() bool { return s.active }

// ShapedDevanagariFaceReady reports whether a Devanagari TTF was found.
func ShapedDevanagariFaceReady() bool {
	initShapedScriptFaces()
	return shapedDevanagari != nil
}

// EnableShapedTextEngine activates shaped mode globally (Notepad / env opt-in).
// Demos should prefer BeginShapedTextSession + Destroy End().
func EnableShapedTextEngine() bool {
	if !InitShapedFonts() {
		return false
	}
	initShapedScriptFaces()
	SetTextEngineMode(TextEngineShaped)
	EnsureComplexScriptUIFont()
	return shapedTextReady()
}

// PrewarmShapedGlyphs uploads shaped glyph textures for text (main thread).
// Use in the Text Engine demo so Arabic/Devanagari appear on the first frame.
func PrewarmShapedGlyphs(text string, fontSize float32) {
	if !shapedTextReady() || text == "" {
		return
	}
	run, ok := shapedShapeText(text, fontSize, false, false, shapedEffectiveMono(false, text), false)
	if !ok || run.out.Face == nil {
		return
	}
	for _, g := range run.out.Glyphs {
		_, _ = shapedEnsureGlyphTexture(run.out.Face, g.GlyphID, fontSize)
	}
}
