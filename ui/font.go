// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── Font state ──────────────────────────────────────────────────────────────

// DefaultFont is the bitmap atlas fallback, loaded by InitFonts.
var DefaultFont rl.Font

// fontReady is true once a TTF atlas has been loaded successfully.
var fontReady bool

// SDFFont holds the Signed Distance Field atlas. When sdfReady, all draw calls
// route through sdfShader/sdfShaderBold for resolution-independent crispness.
var SDFFont rl.Font

// Pre-compiled SDF shaders: normal weight and simulated bold weight.
// Using two static shaders avoids runtime uniform writes and unsafe.Pointer.
var sdfShader rl.Shader     // normal weight — edge at 0.50
var sdfShaderBold rl.Shader // bold weight   — edge at 0.41 (glyph expands ~9%)

// sdfReady is true after InitSDFFont succeeds (full chrome SDF atlas).
var sdfReady bool

// sdfShaderReady is true when SDF shaders are loaded for aux faces (mono /
// PreviewFont) without a chrome SDF atlas — S1 ft-grey chrome path.
var sdfShaderReady bool

// LoadedUIFontPath is the TTF file path used by InitFonts or InitSDFFont, or ""
// when no candidate loaded (raylib default text is used). For F6 typography debug.
var LoadedUIFontPath string

// sdfGlyphs keeps the C-allocated GlyphInfo slice alive for the lifetime of
// SDFFont. The slice header is a Go object; the backing data is C heap.
var sdfGlyphs []rl.GlyphInfo

// ─── SDF fragment shaders ────────────────────────────────────────────────────
// Both shaders use fwidth()-based adaptive smoothing: the AA band automatically
// scales with glyph screen size, giving sharp edges at large sizes and smooth
// AA at small sizes — identical to GPU-accelerated text in Flutter/Skia.
//
// When framebuffer SSAA is reduced (lean / 1×), [sdfAAWidthFactor] widens the
// band so Latin chrome is less stair-steppy without MSDF (GRP-4).

func sdfAAWidthFactor() float32 {
	s := EffectiveSupersamplingScale()
	if s <= 1.01 {
		return 1.2
	}
	if s <= 1.51 {
		return 0.95
	}
	return 0.75
}

func sdfFragSource(edge, aaWidth float32) string {
	return fmt.Sprintf(`#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
out vec4 finalColor;
void main() {
    float dist = texture(texture0, fragTexCoord).a;
    float w = fwidth(dist) * %.3f;
    float alpha = smoothstep(%.2f - w, %.2f + w, dist);
    finalColor = vec4(fragColor.rgb * colDiffuse.rgb, fragColor.a * colDiffuse.a * alpha);
}`, aaWidth, edge, edge)
}

// ─── Configuration ───────────────────────────────────────────────────────────

// FontSpacing is the inter-character spacing in pixels. 0.3 produces tight,
// dense tracking closer to Flutter's default rendering than the looser 0.5.
const FontSpacing = float32(0.3)

// GlobalFontScale is an extra multiplier on [EffectiveFontSize]. Keep at 1.0;
// responsive sizing is handled by [RootFontSize] / [TypeScaleReference] in type_scale.go.
const GlobalFontScale = float32(1.0)

// fontCandidates — default UI body font (menus, labels, chrome). NOT mono.
// Segoe first on Windows (hinted, sharper small UI). Bundled Poppins is opt-in via
// putting it first only when GRU_UI_FONT=poppins (see PreferBundledUIFont).
// Fira Code is editor/code only — see monoFontCandidates in font_stack.go.
var fontCandidates = []string{
	"C:/Windows/Fonts/segoeui.ttf",
	"C:/Windows/Fonts/segoeuib.ttf",
	"assets/fonts/poppins/Poppins-Regular.ttf",
	"assets/fonts/Poppins/static/Poppins-Regular.ttf",
	"assets/fonts/Poppins-Regular.ttf",
	"assets/fonts/Roboto-Regular.ttf",
	"C:/Windows/Fonts/calibri.ttf",
	"C:/Windows/Fonts/tahoma.ttf",
	"C:/Windows/Fonts/arial.ttf",
}

// resolvedFontCandidates returns fontCandidates, optionally preferring Poppins.
func resolvedFontCandidates() []string {
	v := strings.ToLower(strings.TrimSpace(EnvOr("GRU_UI_FONT", "GORY_UI_FONT")))
	if v == "poppins" || v == "bundled" {
		return []string{
			"assets/fonts/poppins/Poppins-Regular.ttf",
			"assets/fonts/Poppins/static/Poppins-Regular.ttf",
			"assets/fonts/Poppins-Regular.ttf",
			"C:/Windows/Fonts/segoeui.ttf",
			"C:/Windows/Fonts/segoeuib.ttf",
			"assets/fonts/Roboto-Regular.ttf",
			"C:/Windows/Fonts/calibri.ttf",
			"C:/Windows/Fonts/tahoma.ttf",
			"C:/Windows/Fonts/arial.ttf",
		}
	}
	return fontCandidates
}

// UIFontFaceSummary returns the active UI TTF basename for dogfood logs.
func UIFontFaceSummary() string {
	p := LoadedUIFontPath
	if p == "" {
		return "none"
	}
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' || p[i] == '\\' {
			return p[i+1:]
		}
	}
	return p
}

func loadFontFromMemory(fileData []byte, atlasSize int32, cps []int32) rl.Font {
	glyphs := rl.LoadFontData(fileData, atlasSize, cps, int32(len(cps)), 0)
	if len(glyphs) == 0 {
		return rl.Font{}
	}
	glyphRecs := make([]*rl.Rectangle, 1)
	atlas := rl.GenImageFontAtlas(glyphs, glyphRecs, atlasSize, 4, 0)
	if atlas.Width == 0 || glyphRecs[0] == nil {
		return rl.Font{}
	}
	tex := rl.LoadTextureFromImage(&atlas)
	rl.UnloadImage(&atlas)
	return rl.Font{
		BaseSize:     atlasSize,
		CharsCount:   int32(len(glyphs)),
		CharsPadding: 4,
		Texture:      tex,
		Recs:         glyphRecs[0],
		Chars:        &glyphs[0],
	}
}

// ─── Initialisation ──────────────────────────────────────────────────────────

// InitFonts loads the best available TTF as a high-res bitmap atlas fallback.
// Call once after rl.InitWindow() and before the draw loop.
func InitFonts(atlasSize int32) {
	cp := sdfCodepoints()
	cps := make([]int32, len(cp))
	for i, r := range cp {
		cps[i] = int32(r)
	}
	for _, path := range resolvedFontCandidates() {
		fileData, err := ReadAssetFile(path)
		if err != nil {
			continue
		}
		f := loadFontFromMemory(fileData, atlasSize, cps)
		if f.BaseSize > 0 {
			rl.SetTextureFilter(f.Texture, rl.FilterBilinear)
			DefaultFont = f
			fontReady = true
			LoadedUIFontPath = path
			return
		}
	}
	DefaultFont = rl.GetFontDefault()
}

// minRenderPx is the smallest SDF size drawTextF will render (fainter below ~11px).
const minRenderPx = float32(11)

// sdfCodepoints builds the set of Unicode codepoints baked into the SDF atlas.
// Covering ASCII + Latin-1 supplement + common UI symbols prevents the
// "?" fallback glyph that appears for any character not in the atlas.
func sdfCodepoints() []rune {
	var cp []rune
	// Standard printable ASCII (32–126)
	for r := rune(32); r <= 126; r++ {
		cp = append(cp, r)
	}
	// Latin-1 supplement (160–255): European letters, currency, common symbols
	for r := rune(160); r <= 255; r++ {
		cp = append(cp, r)
	}
	// Selected Unicode UI symbols
	cp = append(cp,
		'\u2013', '\u2014', // en dash –, em dash —
		'\u2026',           // ellipsis …
		'\u2022',           // bullet •
		'\u2018', '\u2019', // left/right single quotes ''
		'\u201C', '\u201D', // left/right double quotes ""
		'\u2610', '\u2611', // ballot box ☐ ☑
		'\u2715',           // cross ✕
		'\u25B6', '\u25C0', // play ▶ / back ◀
		'\u2190', '\u2192', '\u2191', '\u2193', // ←→↑↓
		'\u00D7',           // multiplication ×
		'\u2212',           // minus sign −
		'\u25CF',           // black circle ●
		'\u00B0',           // degree °
		'\u00B1',           // plus-minus ±
		'\u2265', '\u2264', // ≥ ≤
		'\u2248', // approximately ≈
		'\u03C0', // pi π
	)
	// Box-drawing omitted from the default SDF atlas — those glyphs oversized the
	// atlas cells and spammed raylib "size bigger than expected" warnings. UI
	// separators use theme borders; rare box chars can fall back to tofu.
	return cp
}

// compiles both normal and bold SDF shaders. Returns true on success.
//
// Recommended: sdfAtlasSize = 176 for high-quality SDF at all UI sizes while
// deferring a full MSDF renderer.
// Call after rl.InitWindow() and after InitFonts.
func InitSDFFont(sdfAtlasSize int32) bool {
	// Desktop GLSL 330 SDF shaders do not compile on Android OpenGL ES.
	if AndroidGLES() {
		return false
	}
	// Bundled multi-face stack (Poppins/Inter) is opt-in; default is Segoe via fontCandidates below.
	if useBundledUIFontStack && initUIFontStack(sdfAtlasSize) {
		aa := sdfAAWidthFactor()
		sdfShader = rl.LoadShaderFromMemory("", sdfFragSource(0.50, aa))
		if sdfShader.ID == 0 {
			unloadUIFontStack()
			return false
		}
		sdfShaderBold = rl.LoadShaderFromMemory("", sdfFragSource(0.41, aa))
		if sdfShaderBold.ID == 0 {
			rl.UnloadShader(sdfShader)
			unloadUIFontStack()
			return false
		}
		sdfReady = true
		sdfShaderReady = true
		InitShapedFonts()
		return true
	}
	// Legacy single-font fallback (Segoe / Roboto).
	for _, path := range resolvedFontCandidates() {
		fileData, err := ReadAssetFile(path)
		if err != nil {
			continue
		}

		// Rasterise glyphs as SDF. Pass explicit codepoints so every character
		// used in the UI has a glyph; missing codepoints would fall back to '?'.
		cp := sdfCodepoints()
		cps := make([]int32, len(cp))
		for i, r := range cp {
			cps[i] = int32(r)
		}
		glyphs := rl.LoadFontData(fileData, sdfAtlasSize, cps, int32(len(cps)), int32(rl.FontSdf))
		if len(glyphs) == 0 {
			continue
		}

		// Pack into atlas; glyphRecs[0] → C-allocated Rectangle array.
		glyphRecs := make([]*rl.Rectangle, 1)
		atlas := rl.GenImageFontAtlas(glyphs, glyphRecs, sdfAtlasSize, 4, 0)
		if atlas.Width == 0 || glyphRecs[0] == nil {
			continue
		}

		tex := rl.LoadTextureFromImage(&atlas)
		rl.SetTextureFilter(tex, rl.FilterBilinear)
		rl.UnloadImage(&atlas)

		SDFFont = rl.Font{
			BaseSize:     sdfAtlasSize,
			CharsCount:   int32(len(glyphs)),
			CharsPadding: 4,
			Texture:      tex,
			Recs:         glyphRecs[0],
			Chars:        &glyphs[0],
		}

		// Compile both weight variants. Empty vertex shader = raylib default.
		// Bold: edge 0.41 expands outline ~9% without a separate bold TTF.
		aa := sdfAAWidthFactor()
		sdfShader = rl.LoadShaderFromMemory("", sdfFragSource(0.50, aa))
		if sdfShader.ID == 0 {
			rl.UnloadFont(SDFFont)
			continue
		}
		sdfShaderBold = rl.LoadShaderFromMemory("", sdfFragSource(0.41, aa))
		if sdfShaderBold.ID == 0 {
			rl.UnloadShader(sdfShader)
			rl.UnloadFont(SDFFont)
			continue
		}
		sdfGlyphs = glyphs
		sdfReady = true
		sdfShaderReady = true
		LoadedUIFontPath = path
		InitShapedFonts()
		return true
	}
	return false
}

// ensureSDFShaders loads the SDF fragment shaders without a chrome atlas.
// Used by InitS1AuxFonts so Fira (Mono) and Inter (PreviewFont) can draw under S1.
func ensureSDFShaders() bool {
	if AndroidGLES() {
		return false
	}
	if sdfShaderReady && sdfShader.ID != 0 && sdfShaderBold.ID != 0 {
		return true
	}
	aa := sdfAAWidthFactor()
	sh := rl.LoadShaderFromMemory("", sdfFragSource(0.50, aa))
	if sh.ID == 0 {
		return false
	}
	bold := rl.LoadShaderFromMemory("", sdfFragSource(0.41, aa))
	if bold.ID == 0 {
		rl.UnloadShader(sh)
		return false
	}
	if sdfShader.ID != 0 {
		rl.UnloadShader(sdfShader)
	}
	if sdfShaderBold.ID != 0 {
		rl.UnloadShader(sdfShaderBold)
	}
	sdfShader = sh
	sdfShaderBold = bold
	sdfShaderReady = true
	return true
}

// InitS1AuxFonts loads Mono (Fira) + PreviewFont (Inter) on the same S1 sticky
// LCD pattern as chrome Segoe (separate atlases, identical bake/draw).
func InitS1AuxFonts() {
	InitFTGreyAuxFaces()
}

// sdfFaceDrawReady reports whether SDF shader + a face atlas can draw (chrome or aux).
func sdfFaceDrawReady(font rl.Font) bool {
	if font.BaseSize <= 0 {
		return false
	}
	return (sdfReady || sdfShaderReady) && sdfShader.ID != 0
}

// UnloadSDFFonts releases GPU and C resources. Call during shutdown.
func UnloadSDFFonts() {
	if sdfReady || sdfShaderReady {
		if sdfShader.ID != 0 {
			rl.UnloadShader(sdfShader)
		}
		if sdfShaderBold.ID != 0 {
			rl.UnloadShader(sdfShaderBold)
		}
		sdfReady = false
		sdfShaderReady = false
	}
	sdfMeasureCacheClear()
	unloadFTLCDFont()
	unloadMSDFFont()
	unloadFTGreyFont()
	unloadUIFontStack()
	unloadComplexScriptUIFont()
	unloadShapedGlyphCache()
	unloadShapedFonts()
	_ = sdfGlyphs // GC root for C glyph table until unload
	sdfGlyphs = nil
	SDFFont = rl.Font{}
}

// UIFontBackend reports which raster path drawTextF uses.
func UIFontBackend() string {
	if textEngineMode == TextEngineShaped && shapedUI.ready {
		return "shaped"
	}
	if msdfActive(false, false) {
		return "msdf"
	}
	if ftLCDActive(false, false) {
		return "ft-lcd"
	}
	if ftGreyActive(false, false) {
		if PreferFTGreyStrikes() {
			n := 0
			if ftGreyStrikes != nil {
				n = len(ftGreyStrikes)
			}
			if PreferFTLCDChrome() {
				if dwriteLCDAvailable() {
					if n == 0 {
						return "ft-grey+strikes+dwrite-lcd(empty!)"
					}
					return fmt.Sprintf("ft-grey+strikes+dwrite-lcd(%d)", n)
				}
				return fmt.Sprintf("ft-grey+strikes+lcd(%d)", n)
			}
			return fmt.Sprintf("ft-grey+strikes(%d)", n)
		}
		return "ft-grey"
	}
	if sdfReady {
		return "sdf"
	}
	if fontReady {
		return "bitmap"
	}
	return "raylib-default"
}

// ─── Internal helpers ────────────────────────────────────────────────────────

// fontDensity returns the effective size multiplier for a Style, defaulting
// to 1.0 when FontDensity is zero (the zero-value of float32).
func fontDensity(s Style) float32 {
	if s.FontDensity <= 0 {
		return 1.0
	}
	return s.FontDensity
}

// drawTextF is the core text primitive: float coordinates, float font size,
// and a bold flag that selects the pre-compiled bold SDF shader.
// fontSize is clamped to minRenderPx so SDF coverage never drops to noise.
func drawTextF(text string, x, y, fontSize float32, color rl.Color, bold, italic, mono, preview bool) {
	if text == "" {
		return
	}
	if fontSize < minRenderPx {
		fontSize = minRenderPx
	}
	if shapedTextReady() {
		if shapedDrawTextF(text, x, y, fontSize, color, bold, italic, mono, preview) {
			return
		}
	}
	if msdfActive(mono, preview) {
		drawTextMSDF(text, x, y, fontSize, color)
		return
	}
	if ftLCDActive(mono, preview) {
		drawTextFTLCD(text, x, y, fontSize, color)
		return
	}
	if ftGreyActive(mono, preview) {
		drawTextFTGreyKind(text, x, y, fontSize, color, ftFaceFromFlags(mono, preview), bold)
		return
	}
	pos := rl.NewVector2(x, y)
	font := pickUIFont(bold, italic, mono, preview)
	useBoldShader := uiFontUsesBoldShader(bold, italic, mono, preview)
	if sdfFaceDrawReady(font) {
		shader := sdfShader
		if useBoldShader {
			shader = sdfShaderBold
		}
		rl.BeginShaderMode(shader)
		rl.DrawTextEx(font, text, pos, fontSize, FontSpacing, color)
		rl.EndShaderMode()
	} else if fontReady {
		rl.DrawTextEx(DefaultFont, text, pos, fontSize, FontSpacing, color)
	} else {
		rl.DrawText(text, int32(x), int32(y), int32(fontSize), color)
	}
}

func measureTextF(text string, fontSize float32, bold, italic, mono, preview bool) float32 {
	if fontSize < minRenderPx {
		fontSize = minRenderPx
	}
	if shapedTextReady() && textDrawUsesShaped(text, fontSize, bold, italic, mono, preview) {
		if w, ok := shapedMeasureTextF(text, fontSize, bold, italic, mono, preview); ok {
			return w
		}
	}
	return cachedMeasureTextSDF(text, fontSize, bold, italic, mono, preview)
}

func measureTextSDF(text string, fontSize float32, bold, italic, mono, preview bool) float32 {
	if msdfActive(mono, preview) {
		return measureTextMSDF(text, fontSize)
	}
	if ftLCDActive(mono, preview) {
		return measureTextFTLCD(text, fontSize)
	}
	if ftGreyActive(mono, preview) {
		return measureTextFTGreyKind(text, fontSize, ftFaceFromFlags(mono, preview), bold)
	}
	useBoldShader := uiFontUsesBoldShader(bold, italic, mono, preview)
	measureBold := bold && !useBoldShader
	font := pickUIFont(measureBold, italic, mono, preview)
	var w float32
	if sdfFaceDrawReady(font) {
		w = rl.MeasureTextEx(font, text, fontSize, FontSpacing).X
	} else if fontReady {
		w = rl.MeasureTextEx(DefaultFont, text, fontSize, FontSpacing).X
	} else {
		w = float32(rl.MeasureText(text, int32(fontSize)))
	}
	if useBoldShader {
		w *= sdfBoldShaderMeasureScale
	}
	return w
}

// sdfBoldShaderMeasureScale approximates extra ink width from sdfShaderBold (edge ~0.41 vs 0.50).
const sdfBoldShaderMeasureScale = float32(1.055)

// measureTextDrawAligned returns pixel width using the same backend as drawTextS.
func measureTextDrawAligned(text string, s Style) float32 {
	if text == "" {
		return 0
	}
	fs := EffectiveFontSize(s)
	bold, italic, mono, preview := styleDrawBold(s), s.Italic, s.Mono, s.PreviewFont
	if shapedTextReady() && textDrawUsesShaped(text, fs, bold, italic, mono, preview) {
		if w, ok := shapedMeasureTextF(text, fs, bold, italic, mono, preview); ok {
			return w
		}
	}
	return cachedMeasureTextSDF(text, fs, bold, italic, mono, preview)
}

// EditorMeasureWidth reports prefix width for editor hit-testing and wrap layout.
// Uses the same backend as drawTextS/measureTextS so caret scroll and clicks align.
func EditorMeasureWidth(text string, s Style) float32 {
	return measureTextDrawAligned(text, s)
}

// ─── Public / widget-facing API ──────────────────────────────────────────────

// EffectiveFontSize returns the pixel size drawTextS/measureTextS will use
// (FontDensity, MinFontSize, and minRenderPx clamp). Use for vertical centering
// in widgets — do not use raw Style.FontSize for layout baselines.
//
// S1 (GRU_TEXT_STRIKES=1): sizes snap onto the sticky strike ladder so small
// chrome (tabs, captions) draws 1:1 instead of fractional rem → soft Path A.
func EffectiveFontSize(s Style) float32 {
	// Theme FontSize tokens are authored in [TypeScaleReference] units (16px).
	// RootFontSize is computed from window width (see type_scale.go) and acts
	// like rem-based html { font-size } so all text scales coherently.
	fs := float32(s.FontSize) * fontDensity(s) * GlobalFontScale * typeScaleFactor()
	if s.MinFontSize > 0 && fs < float32(s.MinFontSize) {
		fs = float32(s.MinFontSize)
	}
	if s.MaxFontSize > 0 && fs > float32(s.MaxFontSize) {
		fs = float32(s.MaxFontSize)
	}
	if fs < minRenderPx {
		fs = minRenderPx
	}
	// Quantize onto sticky ladders whenever S1 strikes are on — do not wait for
	// atlas ready (measure/layout before InitFTGrey must still hit exact sizes).
	if PreferFTGreyStrikes() && PreferFTGreyChrome() {
		switch {
		case s.Mono:
			fs = QuantizeFTGreyAuxSize(fs)
		case s.PreviewFont:
			fs = QuantizeFTGreyAuxSize(fs)
		default:
			fs = QuantizeFTGreyChromeSize(fs)
		}
	}
	return fs
}

// TextPosY returns the Y coordinate to draw single-line text vertically centred
// in bounds using [EffectiveFontSize].
func TextPosY(bounds rl.Rectangle, s Style) int32 {
	h := EffectiveFontSize(s)
	// Sticky FT/DWrite ink can exceed rem fontSize — center on the real line box
	// or tops/descenders clip inside Label scissors.
	if ftGreyActive(s.Mono, s.PreviewFont) {
		if box := TextLineBoxHeight(s); box > h {
			h = box
		}
	}
	if h > bounds.Height {
		return int32(bounds.Y)
	}
	y := int32(bounds.Y) + (int32(bounds.Height)-int32(h))/2
	if y < int32(bounds.Y) {
		y = int32(bounds.Y)
	}
	return y
}

// controlTextPosYMath returns draw Y (line-box top) so a cap box of height cap
// sits vertically centered in [boundsY, boundsY+boundsH].
func controlTextPosYMath(boundsY, boundsH, cap, ascent float32) int32 {
	if ascent < 1 {
		ascent = 1
	}
	if cap < 1 {
		cap = ascent * 0.75
	}
	if cap > ascent {
		cap = ascent
	}
	mid := boundsY + boundsH*0.5
	return int32(mid + cap*0.5 - ascent + 0.5)
}

// controlTextPosY centres labels in compact chrome (buttons, badges, chips).
func controlTextPosY(bounds rl.Rectangle, s Style) int32 {
	ascent := TextInkHeight(s)
	cap := TextCapHeight(s)
	fs := EffectiveFontSize(s)
	if ascent < 1 {
		ascent = fs
	}
	if cap < 1 {
		cap = fs * 0.75
	}
	return controlTextPosYMath(bounds.Y, bounds.Height, cap, ascent)
}

// toolbarTextPosY vertically centres toolbar face text using measured glyph height
// (SDF cap height), which reads more balanced than TextPosY inside compact buttons.
func toolbarTextPosY(bounds rl.Rectangle, s Style) int32 {
	if msdfActive(s.Mono, s.PreviewFont) || ftLCDActive(s.Mono, s.PreviewFont) || ftGreyActive(s.Mono, s.PreviewFont) {
		return controlTextPosY(bounds, s)
	}
	fs := EffectiveFontSize(s)
	probe := "Ag"
	var textH float32
	if sdfReady {
		textH = rl.MeasureTextEx(SDFFont, probe, fs, FontSpacing).Y
	} else if fontReady {
		textH = rl.MeasureTextEx(DefaultFont, probe, fs, FontSpacing).Y
	} else {
		textH = fs
	}
	if textH <= 0 || textH > bounds.Height {
		return TextPosY(bounds, s)
	}
	return int32(bounds.Y) + int32((bounds.Height-textH)/2)
}

// statusBarTextPosY vertically centres status-strip labels using measured glyph
// height (same metrics as toolbarTextPosY — avoids clipping descenders on the bottom edge).
func statusBarTextPosY(bounds rl.Rectangle, s Style) int32 {
	return toolbarTextPosY(bounds, s)
}

// toolbarTextHeight returns the rendered cap height for toolbar face text.
func toolbarTextHeight(s Style) float32 {
	fs := EffectiveFontSize(s)
	if msdfActive(s.Mono, s.PreviewFont) {
		return msdfInkHeight(fs)
	}
	if ftLCDActive(s.Mono, s.PreviewFont) {
		return ftLCDInkHeight(fs)
	}
	if ftGreyActive(s.Mono, s.PreviewFont) {
		return ftGreyInkHeightKind(fs, ftFaceFromFlags(s.Mono, s.PreviewFont), styleDrawBold(s))
	}
	if sdfReady {
		return rl.MeasureTextEx(SDFFont, "Ag", fs, FontSpacing).Y
	}
	if fontReady {
		return rl.MeasureTextEx(DefaultFont, "Ag", fs, FontSpacing).Y
	}
	return fs
}

// toolbarTextBottom is the Y coordinate of the bottom edge of toolbar face text in bounds.
func toolbarTextBottom(bounds rl.Rectangle, s Style) float32 {
	return float32(toolbarTextPosY(bounds, s)) + toolbarTextHeight(s)
}

// TextInkBottomY returns a Y coordinate just under glyphs when drawTextS uses lineTopY
// as the raylib position (used by TextEditor spell underlines).
func TextInkBottomY(lineTopY float32, s Style) float32 {
	return lineTopY + TextInkHeight(s) + 1
}

// TextInkHeight returns the rendered cap height for styled text, honouring Mono /
// PreviewFont / bold the same way drawTextS does. TextEditor uses this for the
// caret ("pipe") so I-beam height matches Fira Code / UI face ink, not raw token size.
func TextInkHeight(s Style) float32 {
	fs := EffectiveFontSize(s)
	if msdfActive(s.Mono, s.PreviewFont) {
		return msdfInkHeight(fs)
	}
	if ftLCDActive(s.Mono, s.PreviewFont) {
		return ftLCDInkHeight(fs)
	}
	if ftGreyActive(s.Mono, s.PreviewFont) {
		return ftGreyInkHeightKind(fs, ftFaceFromFlags(s.Mono, s.PreviewFont), styleDrawBold(s))
	}
	font := pickUIFont(styleDrawBold(s), s.Italic, s.Mono, s.PreviewFont)
	if sdfReady && font.BaseSize > 0 {
		if h := rl.MeasureTextEx(font, "Ag", fs, FontSpacing).Y; h > 0 {
			return h
		}
	}
	if fontReady {
		if h := rl.MeasureTextEx(DefaultFont, "Ag", fs, FontSpacing).Y; h > 0 {
			return h
		}
	}
	return fs
}

// drawStyledCaret paints the I-beam aligned to sticky-face ink (not the tall
// line-box top from TextPosY). Using EffectiveFontSize from textTopY made the
// pipe hang above Latin glyphs (search / inputs / editors).
func drawStyledCaret(x, textTopY int32, s Style, col rl.Color) {
	ink := TextInkHeight(s)
	if ink < 1 {
		ink = EffectiveFontSize(s)
	}
	box := EffectiveFontSize(s)
	if ftGreyActive(s.Mono, s.PreviewFont) {
		if lb := TextLineBoxHeight(s); lb > box {
			box = lb
		}
	}
	top := textTopY
	if box > ink+0.5 {
		top = textTopY + int32((box-ink)*0.5+0.5)
	}
	h := int32(ink + 0.5)
	if h < 8 {
		h = 8
	}
	rl.DrawLine(x, top, x, top+h, col)
}

// TextLineBoxHeight is ascent+descent for AutoHeight Labels (ft-grey sticky faces).
// TextInkHeight alone under-sizes the scissor and clips descenders / bold tops.
func TextLineBoxHeight(s Style) float32 {
	fs := EffectiveFontSize(s)
	if ftGreyActive(s.Mono, s.PreviewFont) {
		return ftGreyLineBoxHeightKind(fs, ftFaceFromFlags(s.Mono, s.PreviewFont), styleDrawBold(s))
	}
	ink := TextInkHeight(s)
	if ink > fs {
		return ink + fs*0.2
	}
	return fs
}

// TextCapHeight returns typical capital ink height above the baseline.
func TextCapHeight(s Style) float32 {
	fs := EffectiveFontSize(s)
	h := TextInkHeight(s)
	if h > 1 && h < fs*1.05 {
		return h
	}
	return fs * 0.75
}

// toolbarAccessoryY places a control of height h vertically centered on toolbar face text.
func toolbarAccessoryY(content rl.Rectangle, s Style, h float32) float32 {
	textTop := float32(toolbarTextPosY(content, s))
	textBottom := toolbarTextBottom(content, s)
	mid := (textTop + textBottom) / 2
	return mid - h/2
}

// styleDrawBold resolves whether draw/measure should use a bold face or SDF bold shader.
// When the Poppins stack is loaded, only explicit Style.Bold applies (no caption fake-bold).
func styleDrawBold(s Style) bool {
	if s.Bold {
		return true
	}
	if s.NoCaptionBold || fonts.ready {
		return false
	}
	return s.FontSize > 0 && s.FontSize < 15
}

// drawTextS draws text honouring the full Style: FontDensity scaling, Bold
// weight simulation, per-style MinFontSize override, and automatic bold
// promotion only for caption-sized theme keys (FontSize < 15).
func drawTextS(text string, x, y int32, s Style) {
	fs := EffectiveFontSize(s)
	drawTextF(text, float32(x), float32(y), fs, s.TextColor, styleDrawBold(s), s.Italic, s.Mono, s.PreviewFont)
}

func measureTextS(text string, s Style) int32 {
	return int32(measureTextDrawAligned(text, s))
}

// MeasureTextS reports pixel width for styled text (demos, diagnostics).
func MeasureTextS(text string, s Style) int32 {
	return measureTextS(text, s)
}

// drawText draws text at integer coordinates without a style context.
// All existing callers that pass explicit fontSize/color continue to work.
func drawText(text string, x, y int32, fontSize int32, color rl.Color) {
	drawTextF(text, float32(x), float32(y), float32(fontSize)*GlobalFontScale, color, false, false, false, false)
}

func measureText(text string, fontSize int32) int32 {
	return int32(measureTextF(text, float32(fontSize)*GlobalFontScale, false, false, false, false))
}

// chromeFontSize converts a theme token (e.g. 15) to the pixel size DrawText uses.
// Under S1 strikes, snaps onto the sticky ladder (same as EffectiveFontSize chrome).
func chromeFontSize(token float32) float32 {
	fs := token * GlobalFontScale * typeScaleFactor()
	if PreferFTGreyStrikes() && PreferFTGreyChrome() {
		return QuantizeFTGreyChromeSize(fs)
	}
	return fs
}

// textSkipLCD suppresses ClearType LCD for draws over unknown/transparent backdrops.
var textSkipLCD int

// DrawTextOnTransparent draws S1 grey strikes without LCD (layered transparent hosts).
func DrawTextOnTransparent(text string, x, y, fontSize float32, color rl.Color) {
	textSkipLCD++
	defer func() { textSkipLCD-- }()
	drawTextF(text, x, y, chromeFontSize(fontSize), color, false, false, false, false)
}

// DrawTextOnTransparentBold is semibold [DrawTextOnTransparent].
func DrawTextOnTransparentBold(text string, x, y, fontSize float32, color rl.Color) {
	textSkipLCD++
	defer func() { textSkipLCD-- }()
	drawTextF(text, x, y, chromeFontSize(fontSize), color, true, false, false, false)
}

// MeasureTextOnTransparent matches [DrawTextOnTransparent] metrics.
func MeasureTextOnTransparent(text string, fontSize float32) float32 {
	textSkipLCD++
	defer func() { textSkipLCD-- }()
	return measureTextF(text, chromeFontSize(fontSize), false, false, false, false)
}

// MeasureTextOnTransparentDevicePx matches [DrawTextOnTransparentDevicePx].
func MeasureTextOnTransparentDevicePx(text string, px float32) float32 {
	textSkipLCD++
	defer func() { textSkipLCD-- }()
	return measureTextF(text, px, true, false, false, false)
}

// DrawTextOnTransparentDevicePx draws an exact device-px grey sticky over transparency.
func DrawTextOnTransparentDevicePx(text string, x, y, px float32, color rl.Color) {
	textSkipLCD++
	defer func() { textSkipLCD-- }()
	drawTextF(text, x, y, px, color, true, false, false, false)
}

// DrawText is a public SDF-aware text draw function for use by main.go and
// overlay systems (inspector, tooltips). fontSize is a theme token (same grid as
// Style.FontSize) and scales with RootFontSize like widget text.
func DrawText(text string, x, y float32, fontSize float32, color rl.Color) {
	drawTextF(text, x, y, chromeFontSize(fontSize), color, false, false, false, false)
}

// DrawTextBold draws launcher/overlay chrome text in semibold weight.
func DrawTextBold(text string, x, y, fontSize float32, color rl.Color) {
	drawTextF(text, x, y, chromeFontSize(fontSize), color, true, false, false, false)
}

func MeasureText(text string, fontSize float32) float32 {
	return measureTextF(text, chromeFontSize(fontSize), false, false, false, false)
}
