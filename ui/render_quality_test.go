package ui

import "testing"

func TestRenderQualityNativeDefault(t *testing.T) {
	prevQ := CurrentRenderQuality
	prevBase := BaseSupersamplingScale
	prevFloor := GlyphSupersampleFloor
	t.Cleanup(func() {
		CurrentRenderQuality = prevQ
		BaseSupersamplingScale = prevBase
		GlyphSupersampleFloor = prevFloor
	})

	SetRenderQuality(RenderQualityNative)
	if EffectiveSupersamplingScale() != 1 {
		t.Fatalf("scene SSAA=%v want 1 (native ship default)", EffectiveSupersamplingScale())
	}
	if EffectiveGlyphSupersample() != 4 {
		t.Fatalf("glyph SSAA=%v want 4", EffectiveGlyphSupersample())
	}
}

func TestRenderQualityLeanGlyphFloor(t *testing.T) {
	prevQ := CurrentRenderQuality
	prevBase := BaseSupersamplingScale
	prevFloor := GlyphSupersampleFloor
	t.Cleanup(func() {
		CurrentRenderQuality = prevQ
		BaseSupersamplingScale = prevBase
		GlyphSupersampleFloor = prevFloor
	})

	SetRenderQuality(RenderQualityLean)
	if EffectiveSupersamplingScale() != 1.5 {
		t.Fatalf("scene SSAA=%v want 1.5 (quality floor until exact-size/LCD unlock 1×)", EffectiveSupersamplingScale())
	}
	if EffectiveGlyphSupersample() != 3 {
		t.Fatalf("glyph SSAA=%v want 3", EffectiveGlyphSupersample())
	}
	if shapedGlyphSSAA() != 3 {
		t.Fatalf("shapedGlyphSSAA=%v want 3", shapedGlyphSSAA())
	}
}

func TestRenderQualityMaxFollowsScene(t *testing.T) {
	prevQ := CurrentRenderQuality
	prevBase := BaseSupersamplingScale
	prevFloor := GlyphSupersampleFloor
	t.Cleanup(func() {
		CurrentRenderQuality = prevQ
		BaseSupersamplingScale = prevBase
		GlyphSupersampleFloor = prevFloor
	})

	SetRenderQuality(RenderQualityMax)
	if EffectiveGlyphSupersample() != 2 {
		t.Fatalf("glyph=%v want 2 at max", EffectiveGlyphSupersample())
	}
	BaseSupersamplingScale = 1.5
	GlyphSupersampleFloor = 1
	if EffectiveGlyphSupersample() != 1.5 {
		t.Fatalf("glyph=%v want follow 1.5", EffectiveGlyphSupersample())
	}
}

func TestShapedGlyphRasterPxAtLean(t *testing.T) {
	prevQ := CurrentRenderQuality
	prevBase := BaseSupersamplingScale
	prevFloor := GlyphSupersampleFloor
	prevDPI := DisplayScale
	t.Cleanup(func() {
		CurrentRenderQuality = prevQ
		BaseSupersamplingScale = prevBase
		GlyphSupersampleFloor = prevFloor
		DisplayScale = prevDPI
	})
	DisplayScale = 1
	SetRenderQuality(RenderQualityLean)
	if shapedGlyphRasterPx(16) != 48 {
		t.Fatalf("raster px=%v want 48 (16× glyph floor 3)", shapedGlyphRasterPx(16))
	}
	BaseSupersamplingScale = 1
	GlyphSupersampleFloor = 4
	if shapedGlyphRasterPx(16) != 64 {
		t.Fatalf("raster px at SSAA1=%v want 64 (16× glyph floor 4)", shapedGlyphRasterPx(16))
	}
}
