// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
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

// sdfReady is true after InitSDFFont succeeds.
var sdfReady bool

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

const sdfFragNormal = `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
out vec4 finalColor;
void main() {
    float dist = texture(texture0, fragTexCoord).a;
    float w = fwidth(dist) * 0.75;
    float alpha = smoothstep(0.50 - w, 0.50 + w, dist);
    finalColor = vec4(fragColor.rgb * colDiffuse.rgb, fragColor.a * colDiffuse.a * alpha);
}`

// Bold variant: edge shifted from 0.50 → 0.41, expanding the glyph outline by
// ~9% of the SDF range, which produces a convincing weight increase without a
// separate bold TTF.
const sdfFragBold = `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
uniform sampler2D texture0;
uniform vec4 colDiffuse;
out vec4 finalColor;
void main() {
    float dist = texture(texture0, fragTexCoord).a;
    float w = fwidth(dist) * 0.75;
    float alpha = smoothstep(0.41 - w, 0.41 + w, dist);
    finalColor = vec4(fragColor.rgb * colDiffuse.rgb, fragColor.a * colDiffuse.a * alpha);
}`

// ─── Configuration ───────────────────────────────────────────────────────────

// FontSpacing is the inter-character spacing in pixels. 0.3 produces tight,
// dense tracking closer to Flutter's default rendering than the looser 0.5.
const FontSpacing = float32(0.3)

// GlobalFontScale is an extra multiplier on [EffectiveFontSize]. Keep at 1.0;
// responsive sizing is handled by [RootFontSize] / [TypeScaleReference] in type_scale.go.
const GlobalFontScale = float32(1.0)

// fontCandidates — default UI body font (menus, labels, chrome). NOT mono.
// Fira Code is editor/code only — see monoFontCandidates in font_stack.go.
var fontCandidates = []string{
	"C:/Windows/Fonts/segoeui.ttf",
	"C:/Windows/Fonts/segoeuib.ttf",
	"assets/fonts/poppins/Poppins-Regular.ttf",
	"assets/fonts/Poppins/static/Poppins-Regular.ttf",
	"assets/fonts/Roboto-Regular.ttf",
	"C:/Windows/Fonts/calibri.ttf",
	"C:/Windows/Fonts/tahoma.ttf",
	"C:/Windows/Fonts/arial.ttf",
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
	for _, path := range fontCandidates {
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
	// Box-drawing characters (U+2500–U+257F): used in separator labels like
	// "── new section ──" and decorative dividers.
	for r := rune(0x2500); r <= 0x257F; r++ {
		cp = append(cp, r)
	}
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
		sdfShader = rl.LoadShaderFromMemory("", sdfFragNormal)
		if sdfShader.ID == 0 {
			unloadUIFontStack()
			return false
		}
		sdfShaderBold = rl.LoadShaderFromMemory("", sdfFragBold)
		if sdfShaderBold.ID == 0 {
			rl.UnloadShader(sdfShader)
			unloadUIFontStack()
			return false
		}
		sdfReady = true
		InitShapedFonts()
		return true
	}
	// Legacy single-font fallback (Segoe / Roboto).
	for _, path := range fontCandidates {
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
		sdfShader = rl.LoadShaderFromMemory("", sdfFragNormal)
		if sdfShader.ID == 0 {
			rl.UnloadFont(SDFFont)
			continue
		}
		sdfShaderBold = rl.LoadShaderFromMemory("", sdfFragBold)
		if sdfShaderBold.ID == 0 {
			rl.UnloadShader(sdfShader)
			rl.UnloadFont(SDFFont)
			continue
		}
		sdfGlyphs = glyphs
		sdfReady = true
		LoadedUIFontPath = path
		InitShapedFonts()
		return true
	}
	return false
}

// UnloadSDFFonts releases GPU and C resources. Call during shutdown.
func UnloadSDFFonts() {
	if sdfReady {
		rl.UnloadShader(sdfShader)
		rl.UnloadShader(sdfShaderBold)
		sdfReady = false
	}
	sdfMeasureCacheClear()
	unloadUIFontStack()
	unloadComplexScriptUIFont()
	unloadShapedGlyphCache()
	unloadShapedFonts()
	_ = sdfGlyphs // GC root for C glyph table until unload
	sdfGlyphs = nil
	SDFFont = rl.Font{}
}

// UIFontBackend reports which raster path drawTextF uses ("sdf", "bitmap", or
// "raylib-default"). Used by the F6 typography overlay and stderr report.
func UIFontBackend() string {
	if textEngineMode == TextEngineShaped && shapedUI.ready {
		return "shaped"
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
	pos := rl.NewVector2(x, y)
	font := pickUIFont(bold, italic, mono, preview)
	useBoldShader := uiFontUsesBoldShader(bold, italic, mono, preview)
	if sdfReady && font.BaseSize > 0 {
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
	useBoldShader := uiFontUsesBoldShader(bold, italic, mono, preview)
	measureBold := bold && !useBoldShader
	font := pickUIFont(measureBold, italic, mono, preview)
	var w float32
	if sdfReady && font.BaseSize > 0 {
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
	return fs
}

// TextPosY returns the Y coordinate to draw single-line text vertically centred
// in bounds using [EffectiveFontSize].
func TextPosY(bounds rl.Rectangle, s Style) int32 {
	fs := EffectiveFontSize(s)
	return int32(bounds.Y) + (int32(bounds.Height)-int32(fs))/2
}

// toolbarTextPosY vertically centres toolbar face text using measured glyph height
// (SDF cap height), which reads more balanced than TextPosY inside compact buttons.
func toolbarTextPosY(bounds rl.Rectangle, s Style) int32 {
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
func chromeFontSize(token float32) float32 {
	return token * GlobalFontScale * typeScaleFactor()
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
