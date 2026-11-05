package ui

import "testing"

func TestScaleAtlasPxAt100(t *testing.T) {
	DisplayScale = 1
	if got := scaleAtlasPx(SDFAtlasBasePx); got != SDFAtlasBasePx {
		t.Fatalf("scaleAtlasPx at 1x = %d, want %d", got, SDFAtlasBasePx)
	}
}

func TestScaleAtlasPxAt125(t *testing.T) {
	DisplayScale = 1.25
	want := int32(220) // 176 * 1.25
	if got := scaleAtlasPx(SDFAtlasBasePx); got != want {
		t.Fatalf("scaleAtlasPx at 125%% = %d, want %d", got, want)
	}
	if got := scaleAtlasPx(RemixAtlasBasePx); got != 640 {
		t.Fatalf("remix atlas at 125%% = %d, want 640", got)
	}
}

func TestScaleAtlasPxIgnoresSSAA(t *testing.T) {
	prevBase := BaseSupersamplingScale
	t.Cleanup(func() { BaseSupersamplingScale = prevBase })
	DisplayScale = 1.25
	BaseSupersamplingScale = 2
	a := scaleAtlasPx(SDFAtlasBasePx)
	BaseSupersamplingScale = 1
	b := scaleAtlasPx(SDFAtlasBasePx)
	if a != b {
		t.Fatalf("atlas must not grow when SSAA drops: max=%d ssaa1=%d", a, b)
	}
}

func TestEffectiveSupersamplingIgnoresDPI(t *testing.T) {
	prevDPI := DisplayScale
	prevSSAA := BaseSupersamplingScale
	DisplayScale = 1.5
	BaseSupersamplingScale = 2
	defer func() {
		DisplayScale = prevDPI
		BaseSupersamplingScale = prevSSAA
	}()
	if got := EffectiveSupersamplingScale(); got != 2 {
		t.Fatalf("SSAA = %.2f, want 2 (not multiplied by DPI)", got)
	}
}

func TestShapedGlyphRasterIncludesDPI(t *testing.T) {
	prevDPI := DisplayScale
	prevSSAA := BaseSupersamplingScale
	prevFloor := GlyphSupersampleFloor
	DisplayScale = 1.25
	BaseSupersamplingScale = 2
	GlyphSupersampleFloor = 1 // follow scene SSAA; S1 default floor 4 would mask DPI×SSAA
	defer func() {
		DisplayScale = prevDPI
		BaseSupersamplingScale = prevSSAA
		GlyphSupersampleFloor = prevFloor
	}()
	want := float32(16 * 2 * 1.25)
	if got := shapedGlyphRasterPx(16); got != want {
		t.Fatalf("raster px = %v, want %v", got, want)
	}
}

func TestDisplayScaleNearlyEqual(t *testing.T) {
	if !displayScaleNearlyEqual(1.25, 1.26) {
		t.Fatal("expected near-equal")
	}
	if displayScaleNearlyEqual(1.0, 1.25) {
		t.Fatal("expected different")
	}
}
