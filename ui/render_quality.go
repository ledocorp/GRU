package ui

import (
	"fmt"
	"strings"
	"sync"
)

// RenderQuality is the user/product SSAA + text policy (FX Phase 4).
// Distinct from [RenderCapability] (what the GPU actually initialized).
type RenderQuality int

const (
	// RenderQualityNative — SSAA 1× framebuffer (ship default).
	RenderQualityNative RenderQuality = iota
	// RenderQualityMax — full-scene 2× SSAA (testing / legacy quality tier).
	RenderQualityMax
	// RenderQualityBalanced — 1.5× SSAA (testing).
	RenderQualityBalanced
	// RenderQualityLean — 1.5× SSAA RT + glyph ≥3× (testing).
	RenderQualityLean
)

// CurrentRenderQuality is the active policy. Set by [ApplyRenderQualityFromEnv]
// or [SetRenderQuality] before [InitSupersampling].
var CurrentRenderQuality = RenderQualityNative

// GlyphSupersampleFloor is the minimum multiplier for shaped/outline glyph CPU
// raster. Native ship default keeps glyph floor at 4× while RT stays 1×.
var GlyphSupersampleFloor float32 = 4

var ssaa1WarnOnce sync.Once
var minimalWarnOnce sync.Once
var s1DefaultsLogOnce sync.Once

func applyS1Defaults() {
	applyStudioDogfoodQuality()
	s1DefaultsLogOnce.Do(func() {
		fmt.Println("Gru: S1 defaults — lean SSAA 1× + ft-grey strikes (+ LCD when available)")
	})
}

func applyStudioDogfoodQuality() {
	SetRenderQuality(RenderQualityLean)
	BaseSupersamplingScale = 1
	if GlyphSupersampleFloor < 4 {
		GlyphSupersampleFloor = 4
	}
}

func printSSAA1Warning() {
	ssaa1WarnOnce.Do(func() {
			fmt.Println("Gru: GRU_SSAA=1 — smaller RT; atlases stay DPI-sized (not boosted); Latin outline + fx AA when shaped")
	})
}

// SetRenderQuality applies a quality tier (SSAA base + glyph floor).
// Call before InitSupersampling / ApplyDisplayAwareSupersampling.
func SetRenderQuality(q RenderQuality) {
	CurrentRenderQuality = q
	switch q {
	case RenderQualityNative:
		BaseSupersamplingScale = 1
		GlyphSupersampleFloor = 4
	case RenderQualityLean:
		BaseSupersamplingScale = 1.5
		GlyphSupersampleFloor = 3
	case RenderQualityBalanced:
		BaseSupersamplingScale = 1.5
		GlyphSupersampleFloor = 3
	default:
		CurrentRenderQuality = RenderQualityMax
		BaseSupersamplingScale = 2
		GlyphSupersampleFloor = 1
	}
}

// ApplyRenderQualityFromEnv reads GRU_RENDER_QUALITY (GORY alias) and optional
// GRU_SSAA / GORY_SSAA override. Safe to call multiple times; no-op when unset
// (keeps Max defaults) unless SSAA override is set alone.
//
//	GRU_RENDER_QUALITY=max|balanced|lean
//	GRU_SSAA=1|1.5|2          // optional explicit framebuffer scale
func ApplyRenderQualityFromEnv() {
	qName := strings.ToLower(strings.TrimSpace(EnvOr("GRU_RENDER_QUALITY", "GORY_RENDER_QUALITY")))
	ssaa := strings.TrimSpace(EnvOr("GRU_SSAA", "GORY_SSAA"))
	if qName == "" && ssaa == "" {
		applyS1Defaults()
		return
	}
	switch qName {
	case "lean", "low":
		SetRenderQuality(RenderQualityLean)
	case "balanced", "bal", "1.5":
		SetRenderQuality(RenderQualityBalanced)
	case "max", "high", "2", "2x", "2.0":
		SetRenderQuality(RenderQualityMax)
	case "1x", "1.0", "minimal":
		// Explicit true 1× framebuffer — expect aliased SDF; RSS still atlas-heavy.
		SetRenderQuality(RenderQualityLean)
		BaseSupersamplingScale = 1
		GlyphSupersampleFloor = 4
		minimalWarnOnce.Do(func() {
			fmt.Println("Gru: RENDER_QUALITY minimal/1× — smaller SSAA RT; atlases DPI-only; Latin outline when shaped")
		})
	case "":
		SetRenderQuality(RenderQualityNative)
	default:
		fmt.Printf("Gru: unknown GRU_RENDER_QUALITY=%q (use max|balanced|lean|minimal)\n", qName)
	}

	if ssaa != "" {
		switch ssaa {
		case "1", "1.0":
			BaseSupersamplingScale = 1
			// Outline/complex glyphs: 3× still stair-steps Arabic/Devanagari at true 1× RT.
			if GlyphSupersampleFloor < 4 {
				GlyphSupersampleFloor = 4
			}
			printSSAA1Warning()
		case "1.5":
			BaseSupersamplingScale = 1.5
		case "2", "2.0":
			BaseSupersamplingScale = 2
		default:
			fmt.Printf("Gru: unknown GRU_SSAA=%q (use 1|1.5|2)\n", ssaa)
		}
		if qName == "" && BaseSupersamplingScale <= 1.5 && GlyphSupersampleFloor < 2 {
			GlyphSupersampleFloor = 2
			if BaseSupersamplingScale <= 1 {
				CurrentRenderQuality = RenderQualityLean
			} else {
				CurrentRenderQuality = RenderQualityBalanced
			}
		}
	}
}

// RenderQualityString returns a short label for logs / dogfood.
func RenderQualityString() string {
	switch CurrentRenderQuality {
	case RenderQualityNative:
		return "native (SSAA 1× RT, glyph≥4×)"
	case RenderQualityLean:
		if EffectiveSupersamplingScale() <= 1.01 {
			return "lean (SSAA 1× RT, glyph≥4×; atlas-bound WS)"
		}
		return "lean (SSAA 1.5×, glyph≥3×)"
	case RenderQualityBalanced:
		return "balanced (SSAA 1.5×, glyph≥3×)"
	default:
		return "max (SSAA 2×)"
	}
}

// EffectiveGlyphSupersample is max(scene SSAA, GlyphSupersampleFloor).
func EffectiveGlyphSupersample() float32 {
	s := EffectiveSupersamplingScale()
	if s < 1 {
		s = 1
	}
	f := GlyphSupersampleFloor
	if f < 1 {
		f = 1
	}
	if f > s {
		return f
	}
	return s
}
