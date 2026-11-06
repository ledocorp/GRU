package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// T1.9 — DPR-scaled source atlases (SDF text + Remix icons). Layout stays logical;
// FlagWindowHighdpi remains off. SSAA is not multiplied by DisplayScale.
//
// Important: do **not** grow atlases when SSAA drops. Lean/GRU_SSAA=1 must shrink
// the framebuffer RT, not inflate SDF/Remix (that caused WS ↑ while quality ↓).

const (
	SDFAtlasBasePx   int32 = 176
	RemixAtlasBasePx int32 = 512
	BitmapFontBasePx int32 = 176
	// FTGreyAtlasBasePx — single greyscale atlas (~48×DPI). Keep lean for WS;
	// exact-size strikes are opt-in (GRU_TEXT_STRIKES=1), not denser default bake.
	FTGreyAtlasBasePx int32 = 48
)

var atlasForGlyphScale float32

// EffectiveGlyphDisplayScale is DisplayScale clamped for glyph raster / T1.9.
// Linux keeps DisplayScale=1 (no compositor densify).
func EffectiveGlyphDisplayScale() float32 {
	s := DisplayScale
	if s < 1 {
		return 1
	}
	if s > 2.5 {
		return 2.5
	}
	return s
}

func scaleAtlasPx(base int32) int32 {
	if base <= 0 {
		return base
	}
	s := float32(base) * DisplayScale
	if s < float32(base) {
		s = float32(base)
	}
	cap := float32(base) * 2.5
	if s > cap {
		s = cap
	}
	return int32(s + 0.5)
}

// EffectiveSDFAtlasSize returns the SDF TTF raster size for the current DPI.
func EffectiveSDFAtlasSize() int32 { return scaleAtlasPx(SDFAtlasBasePx) }

// EffectiveRemixAtlasSize returns the Remix icon-font atlas size for the current DPI.
// Android GLES emulators often cap textures at 4096² — a 1280px glyph size with a
// growing codepoint set overflows GenImageFontAtlas (black FAB / missing chrome icons).
func EffectiveRemixAtlasSize() int32 {
	if AndroidGLES() {
		// 320: enough for shell chrome (home/search/user/add) without exceeding
		// emulator max_texture_size when the warm set grows (was 1280→GL errors).
		return 320
	}
	return scaleAtlasPx(RemixAtlasBasePx)
}

// EffectiveBitmapFontAtlasSize returns the bitmap UI font fallback atlas size.
// Android skips SDF; baking near ~2× peak rem size keeps bilinear scale mild
// (a 440px atlas drawn at ~45px looks soft/mushy on GLES).
func EffectiveBitmapFontAtlasSize() int32 {
	if AndroidGLES() {
		return 112
	}
	return scaleAtlasPx(BitmapFontBasePx)
}

// EffectiveFTGreyAtlasSize is the greyscale Latin bake size (~48×DPI).
func EffectiveFTGreyAtlasSize() int32 { return scaleAtlasPx(FTGreyAtlasBasePx) }

// InitDisplayAwareAtlases loads icon + text atlases scaled by [DisplayScale].
// When ft-grey is preferred and init succeeds, skip SDF + full bitmap atlases (GRP-7).
// Call after [RefreshDisplayScale] and rl.InitWindow.
func InitDisplayAwareAtlases() {
	ApplyRenderQualityFromEnv()
	RefreshDisplayScale()
	RefreshDensity()
	atlasForGlyphScale = EffectiveGlyphDisplayScale()
	InitIcons(EffectiveRemixAtlasSize())
	if LoadedUIFontPath == "" {
		LoadedUIFontPath = firstExisting(resolvedFontCandidates()...)
	}
	if PreferMSDFChrome() && InitMSDFFont(EffectiveFTGreyAtlasSize()) {
		return
	}
	// S1 (desktop + Android GLES): ft-grey (+ exact strikes / outline AA) first.
	// LCD/DWrite attach only when available (Windows). Bitmap+SDF remain fallback
	// (Android used to force bitmap-only — that skipped Medium/SemiBold soft-bold
	// and outline bake lessons from linux-s1-text.md).
	if PreferFTGreyStrikes() && PreferFTGreyChrome() && InitFTGreyFont(EffectiveFTGreyAtlasSize()) {
		if PreferFTLCDChrome() {
			_ = ensureFTLCDShader()
		}
		InitS1AuxFonts()
		return
	}
	if PreferFTLCDChrome() && InitFTLCDFont(EffectiveFTGreyAtlasSize()) {
		return
	}
	if PreferFTGreyChrome() && InitFTGreyFont(EffectiveFTGreyAtlasSize()) {
		return
	}
	InitFonts(EffectiveBitmapFontAtlasSize())
	InitSDFFont(EffectiveSDFAtlasSize())
}

func unloadBitmapUIFont() {
	if fontReady && DefaultFont.Texture.ID != 0 {
		rl.UnloadFont(DefaultFont)
	}
	fontReady = false
	DefaultFont = rl.Font{}
}

func reinitDisplayAwareAtlases() {
	mode := textEngineMode
	latin := shapedLatinMeasureActive

	UnloadSDFFonts()
	unloadBitmapUIFont()
	Icons.UnloadAll()
	unloadShapedFonts()

	InitIcons(EffectiveRemixAtlasSize())
	if LoadedUIFontPath == "" {
		LoadedUIFontPath = firstExisting(resolvedFontCandidates()...)
	}
	if PreferMSDFChrome() && InitMSDFFont(EffectiveFTGreyAtlasSize()) {
		// ok
	} else if PreferFTGreyStrikes() && PreferFTGreyChrome() && InitFTGreyFont(EffectiveFTGreyAtlasSize()) {
		if PreferFTLCDChrome() {
			_ = ensureFTLCDShader()
		}
		InitS1AuxFonts()
	} else if PreferFTLCDChrome() && InitFTLCDFont(EffectiveFTGreyAtlasSize()) {
		// ok
	} else if PreferFTGreyChrome() && InitFTGreyFont(EffectiveFTGreyAtlasSize()) {
		// ok
	} else {
		InitFonts(EffectiveBitmapFontAtlasSize())
		InitSDFFont(EffectiveSDFAtlasSize())
	}

	SetTextEngineMode(mode)
	shapedLatinMeasureActive = latin
	if mode == TextEngineShaped {
		initShapedScriptFaces()
		EnsureComplexScriptUIFont()
	}
}

// ReinitDisplayAwareAtlasesIfNeeded rebuilds text/icon atlases when DisplayScale
// changes (e.g. window moved to another monitor). Returns true when reloaded.
func ReinitDisplayAwareAtlasesIfNeeded() bool {
	RefreshDisplayScale()
	RefreshDensity()
	next := EffectiveGlyphDisplayScale()
	prev := atlasForGlyphScale
	if prev > 0 && displayScaleNearlyEqual(next, prev) {
		return false
	}
	reinitDisplayAwareAtlases()
	atlasForGlyphScale = next
	return true
}

func displayScaleNearlyEqual(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.02
}
