package ui

import "testing"

func TestShapedGlyphCacheLatinOptInOnly(t *testing.T) {
	prevBase := BaseSupersamplingScale
	prevLatin := shapedLatinMeasureActive
	t.Cleanup(func() {
		BaseSupersamplingScale = prevBase
		shapedLatinMeasureActive = prevLatin
		t.Setenv("GRU_SHAPED_GLYPH_CACHE", "")
		t.Setenv("GORY_SHAPED_GLYPH_CACHE", "")
	})
	t.Setenv("GRU_SHAPED_GLYPH_CACHE", "")
	shapedLatinMeasureActive = true
	BaseSupersamplingScale = 1
	if shapedGlyphCacheEnabled() {
		t.Fatal("Latin glyph-cache must not auto-enable (wrap measure≠draw advances)")
	}
	t.Setenv("GRU_SHAPED_GLYPH_CACHE", "1")
	if !shapedGlyphCacheEnabled() {
		t.Fatal("env=1 enables experiment")
	}
}

func TestPreferShapedTextForQualityOff(t *testing.T) {
	if PreferShapedTextForQuality() {
		t.Fatal("Studio must not auto-prefer shaped for lean SSAA")
	}
}
