package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// T1.9 — DPR-scaled source atlases (SDF text + Remix icons). Layout stays logical;
// FlagWindowHighdpi remains off. SSAA is not multiplied by DisplayScale.

const (
	SDFAtlasBasePx   int32 = 176
	RemixAtlasBasePx int32 = 512
	BitmapFontBasePx int32 = 176
)

var atlasForDisplayScale float32

// EffectiveGlyphDisplayScale is DisplayScale clamped for glyph raster strike keys.
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
func EffectiveRemixAtlasSize() int32 { return scaleAtlasPx(RemixAtlasBasePx) }

// EffectiveBitmapFontAtlasSize returns the bitmap UI font fallback atlas size.
func EffectiveBitmapFontAtlasSize() int32 { return scaleAtlasPx(BitmapFontBasePx) }

// InitDisplayAwareAtlases loads bitmap, Remix, and SDF atlases scaled by [DisplayScale].
// Call after [RefreshDisplayScale] and rl.InitWindow.
func InitDisplayAwareAtlases() {
	RefreshDisplayScale()
	atlasForDisplayScale = DisplayScale
	InitFonts(EffectiveBitmapFontAtlasSize())
	InitIcons(EffectiveRemixAtlasSize())
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

	InitFonts(EffectiveBitmapFontAtlasSize())
	InitIcons(EffectiveRemixAtlasSize())
	InitSDFFont(EffectiveSDFAtlasSize())

	SetTextEngineMode(mode)
	shapedLatinMeasureActive = latin
	if mode == TextEngineShaped {
		initShapedScriptFaces()
		EnsureComplexScriptUIFont()
	}
}

// ReinitDisplayAwareAtlasesIfNeeded rebuilds text/icon atlases when [DisplayScale] changes
// (e.g. window moved to another monitor). Returns true when atlases were reloaded.
func ReinitDisplayAwareAtlasesIfNeeded() bool {
	prev := atlasForDisplayScale
	RefreshDisplayScale()
	if prev > 0 && displayScaleNearlyEqual(DisplayScale, prev) {
		return false
	}
	reinitDisplayAwareAtlases()
	atlasForDisplayScale = DisplayScale
	return true
}

func displayScaleNearlyEqual(a, b float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d < 0.02
}
