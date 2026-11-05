// Package ui (continued) — G1.1 analytic coverage AA for rounded fills + rings at SSAA 1×.
//
// Shader outputs a = tint.a * cover. Must blend onto an **opaque** dest in the
// main RT (BeginSuperFrame page fill). Onto Blank, then alpha-blit over the
// screen fill, the fringe doubles (fat list-tile / panel frames). Remaining
// “thick hairline” after that fix is fringe width + BorderWidth snap (G1.2),
// not a second font. docs/PRESENT_1X_COMPOSITE.md · docs/GEOMETRY_AA_1X.md
package ui

import (
	"math"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// PreferGeometryCoverageAA is true when scene SSAA cannot soften tessellated
// chrome (S1 dogfood). See docs/GEOMETRY_AA_1X.md G1.1.
func PreferGeometryCoverageAA() bool {
	return EffectiveSupersamplingScale() <= 1.01
}

// Coverage AA fringe (device px). Fills stay tighter so pills/toggles don’t go soft;
// rings/1px borders need a wider fringe or high-contrast edges stair-step.
// Hairline AA is for 1px inset borders / thin rings — fill AA stacked twice reads as 2px.
const (
	covAAPad       = float32(1.35) // quad bleed so smoothstep isn’t clipped
	covFillAAMin   = float32(0.45)
	covFillAAMax   = float32(0.90)
	// Thin rings (calendar selection) need a wider fringe than solid fills.
	covStrokeAAMin = float32(0.90)
	covStrokeAAMax = float32(1.85)
	covHairlineAAMin = float32(0.08)
	covHairlineAAMax = float32(0.50)
)

var (
	coverageRRectShader rl.Shader
	coverageRRectOnce   sync.Once
	coverageRRectOK     bool
	covLocSize          int32 = -1
	covLocRadius        int32 = -1
	covLocPad           int32 = -1
	covLocAAMin         int32 = -1
	covLocAAMax         int32 = -1

	coverageRingShader rl.Shader
	coverageRingOnce   sync.Once
	coverageRingOK     bool
	covRingLocSize     int32 = -1
	covRingLocInner    int32 = -1
	covRingLocOuter    int32 = -1
	covRingLocStart    int32 = -1
	covRingLocEnd      int32 = -1
	covRingLocAAMin    int32 = -1
	covRingLocAAMax    int32 = -1

	covWhiteTex     rl.Texture2D
	covWhiteTexOnce sync.Once
)

func ensureCoverageWhiteTex() bool {
	covWhiteTexOnce.Do(func() {
		img := rl.GenImageColor(4, 4, rl.White)
		covWhiteTex = rl.LoadTextureFromImage(img)
		rl.UnloadImage(img)
		if covWhiteTex.ID != 0 {
			// Point sample — bilinear on a tiny white tex softens coverage fringes.
			rl.SetTextureFilter(covWhiteTex, rl.FilterPoint)
		}
	})
	return covWhiteTex.ID != 0
}

func coverageRRectFragSource() string {
	// Screen-stable rounded-box SDF with fwidth AA (Impeller/Flutter-class).
	// Quad is padded (uPad); uSize is the logical control size so the AA fringe
	// is not clipped by the draw rectangle.
	return `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
out vec4 finalColor;
uniform vec2 uSize;
uniform float uRadius;
uniform float uPad;
uniform float uAAMin;
uniform float uAAMax;
uniform sampler2D texture0;

float sdRoundBox(vec2 p, vec2 b, float r) {
	vec2 q = abs(p) - b + vec2(r);
	return length(max(q, 0.0)) + min(max(q.x, q.y), 0.0) - r;
}

void main() {
	vec2 full = uSize + vec2(uPad * 2.0);
	vec2 p = (fragTexCoord - vec2(0.5)) * full;
	vec2 halfSz = uSize * 0.5;
	float rad = clamp(uRadius, 0.0, min(halfSz.x, halfSz.y));
	float d = sdRoundBox(p, halfSz, rad);
	float aa = clamp(fwidth(d), uAAMin, uAAMax);
	float cover = 1.0 - smoothstep(-aa, aa, d);
	vec4 tint = fragColor;
	finalColor = vec4(tint.rgb, tint.a * cover);
}`
}

func coverageRingFragSource() string {
	// Annulus SDF + optional angular wedge (raylib DrawRing angle space: Y-down,
	// 0° = +X / east, positive CCW — matches atan(p.y, p.x) in this UV space).
	return `#version 330
in vec2 fragTexCoord;
in vec4 fragColor;
out vec4 finalColor;
uniform vec2 uSize;
uniform float uInner;
uniform float uOuter;
uniform float uStartDeg;
uniform float uEndDeg;
uniform float uAAMin;
uniform float uAAMax;
uniform sampler2D texture0;

void main() {
	vec2 p = (fragTexCoord - vec2(0.5)) * uSize;
	float r = length(p);
	float d = max(r - uOuter, uInner - r);
	float aa = clamp(fwidth(d), uAAMin, uAAMax);
	float cover = 1.0 - smoothstep(-aa, aa, d);

	float span = uEndDeg - uStartDeg;
	// Full ring (or near-full): skip angular mask.
	if (span < 359.5 && span > -359.5) {
		while (span < 0.0) span += 360.0;
		while (span >= 360.0) span -= 360.0;
		if (span < 359.5) {
			float ang = degrees(atan(p.y, p.x));
			float a = ang - uStartDeg;
			while (a < 0.0) a += 360.0;
			while (a >= 360.0) a -= 360.0;
			float aaA = max(fwidth(a), 0.55);
			float angCover = smoothstep(-aaA, aaA, a) * (1.0 - smoothstep(span - aaA, span + aaA, a));
			cover *= angCover;
		}
	}

	vec4 tint = fragColor;
	finalColor = vec4(tint.rgb, tint.a * cover);
}`
}

func ensureCoverageRRectShader() bool {
	coverageRRectOnce.Do(func() {
		if !ensureCoverageWhiteTex() {
			return
		}
		sh := rl.LoadShaderFromMemory("", coverageRRectFragSource())
		if sh.ID == 0 {
			return
		}
		coverageRRectShader = sh
		covLocSize = rl.GetShaderLocation(sh, "uSize")
		covLocRadius = rl.GetShaderLocation(sh, "uRadius")
		covLocPad = rl.GetShaderLocation(sh, "uPad")
		covLocAAMin = rl.GetShaderLocation(sh, "uAAMin")
		covLocAAMax = rl.GetShaderLocation(sh, "uAAMax")
		coverageRRectOK = covLocSize >= 0
	})
	return coverageRRectOK
}

func ensureCoverageRingShader() bool {
	coverageRingOnce.Do(func() {
		if !ensureCoverageWhiteTex() {
			return
		}
		sh := rl.LoadShaderFromMemory("", coverageRingFragSource())
		if sh.ID == 0 {
			return
		}
		coverageRingShader = sh
		covRingLocSize = rl.GetShaderLocation(sh, "uSize")
		covRingLocInner = rl.GetShaderLocation(sh, "uInner")
		covRingLocOuter = rl.GetShaderLocation(sh, "uOuter")
		covRingLocStart = rl.GetShaderLocation(sh, "uStartDeg")
		covRingLocEnd = rl.GetShaderLocation(sh, "uEndDeg")
		covRingLocAAMin = rl.GetShaderLocation(sh, "uAAMin")
		covRingLocAAMax = rl.GetShaderLocation(sh, "uAAMax")
		coverageRingOK = covRingLocSize >= 0 && covRingLocInner >= 0 && covRingLocOuter >= 0
	})
	return coverageRingOK
}

func setCoverageFillAA() {
	setCoverageFillAARange(covFillAAMin, covFillAAMax)
}

func setCoverageFillAARange(aaMin, aaMax float32) {
	if covLocAAMin >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocAAMin, []float32{aaMin}, rl.ShaderUniformFloat)
	}
	if covLocAAMax >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocAAMax, []float32{aaMax}, rl.ShaderUniformFloat)
	}
}

// drawCoverageRoundedFill draws a rounded rect with analytic edge coverage (G1.1).
// Falls back to tessellated DrawRectangleRounded when the shader is unavailable.
func drawCoverageRoundedFill(r rl.Rectangle, cornerRadius float32, col rl.Color) {
	r = snapControlRect(r)
	if r.Width < 1 || r.Height < 1 {
		return
	}
	cr := cornerRadius
	if cr < 0.5 {
		rl.DrawRectangleRec(r, col)
		return
	}
	if cr > r.Width/2 {
		cr = r.Width / 2
	}
	if cr > r.Height/2 {
		cr = r.Height / 2
	}
	if !PreferGeometryCoverageAA() || !ensureCoverageRRectShader() {
		rl.DrawRectangleRounded(r, cornerRoundnessFromRadius(r, cr), 48, col)
		return
	}
	pad := covAAPad
	dst := rl.NewRectangle(r.X-pad, r.Y-pad, r.Width+2*pad, r.Height+2*pad)
	rl.BeginShaderMode(coverageRRectShader)
	if covLocSize >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocSize, []float32{r.Width, r.Height}, rl.ShaderUniformVec2)
	}
	if covLocRadius >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocRadius, []float32{cr}, rl.ShaderUniformFloat)
	}
	if covLocPad >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocPad, []float32{pad}, rl.ShaderUniformFloat)
	}
	setCoverageFillAA()
	src := rl.NewRectangle(0, 0, float32(covWhiteTex.Width), float32(covWhiteTex.Height))
	rl.DrawTexturePro(covWhiteTex, src, dst, rl.NewVector2(0, 0), 0, col)
	rl.EndShaderMode()
}

func drawCoverageRoundedFillAA(r rl.Rectangle, cornerRadius, aaMin, aaMax float32, col rl.Color) {
	r = snapControlRect(r)
	if r.Width < 1 || r.Height < 1 {
		return
	}
	cr := cornerRadius
	if cr < 0.5 {
		rl.DrawRectangleRec(r, col)
		return
	}
	if cr > r.Width/2 {
		cr = r.Width / 2
	}
	if cr > r.Height/2 {
		cr = r.Height / 2
	}
	if !PreferGeometryCoverageAA() || !ensureCoverageRRectShader() {
		rl.DrawRectangleRounded(r, cornerRoundnessFromRadius(r, cr), 48, col)
		return
	}
	pad := covAAPad
	dst := rl.NewRectangle(r.X-pad, r.Y-pad, r.Width+2*pad, r.Height+2*pad)
	rl.BeginShaderMode(coverageRRectShader)
	if covLocSize >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocSize, []float32{r.Width, r.Height}, rl.ShaderUniformVec2)
	}
	if covLocRadius >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocRadius, []float32{cr}, rl.ShaderUniformFloat)
	}
	if covLocPad >= 0 {
		rl.SetShaderValue(coverageRRectShader, covLocPad, []float32{pad}, rl.ShaderUniformFloat)
	}
	setCoverageFillAARange(aaMin, aaMax)
	src := rl.NewRectangle(0, 0, float32(covWhiteTex.Width), float32(covWhiteTex.Height))
	rl.DrawTexturePro(covWhiteTex, src, dst, rl.NewVector2(0, 0), 0, col)
	rl.EndShaderMode()
}

// drawCoverageCircle draws a filled circle with G1.1 coverage AA.
// Uses the ring shader (inner=0) so the disc center is not distorted by
// snapControlRect on a square bbox (that made rings and day labels disagree).
func drawCoverageCircle(cx, cy, radius float32, col rl.Color) {
	if radius < 0.5 {
		return
	}
	if PreferGeometryCoverageAA() && ensureCoverageRingShader() {
		drawCoverageAnnulus(cx, cy, 0, radius, 0, 360, col, covFillAAMin, covFillAAMax)
		return
	}
	rl.DrawCircleV(rl.NewVector2(cx, cy), radius, col)
}

// DrawCoverageCircle is the exported G1.1 circle fill (falls back when SSAA > 1).
func DrawCoverageCircle(cx, cy, radius float32, col rl.Color) {
	if PreferGeometryCoverageAA() {
		drawCoverageCircle(cx, cy, radius, col)
		return
	}
	rl.DrawCircleV(rl.NewVector2(cx, cy), radius, col)
}

// DrawCoverageRing draws an annular arc (full ring when sweep ≥ ~360°) with G1.1 coverage.
// Angles match raylib DrawRing (degrees, Y-down, 0° = east).
func DrawCoverageRing(cx, cy, inner, outer, startDeg, endDeg float32, col rl.Color) {
	DrawCoverageRingAA(cx, cy, inner, outer, startDeg, endDeg, col, false)
}

// DrawCoverageRingSharp uses tighter AA for thick loader/gauge strokes (less fuzzy).
func DrawCoverageRingSharp(cx, cy, inner, outer, startDeg, endDeg float32, col rl.Color) {
	DrawCoverageRingAA(cx, cy, inner, outer, startDeg, endDeg, col, true)
}

// DrawCoverageRingAA is the shared annulus path. sharp=true tightens stroke fringe.
func DrawCoverageRingAA(cx, cy, inner, outer, startDeg, endDeg float32, col rl.Color, sharp bool) {
	if outer < 0.5 || outer <= inner {
		return
	}
	if !PreferGeometryCoverageAA() || !ensureCoverageRingShader() {
		segs := int32(math.Max(48, math.Pi*float64(outer)))
		rl.DrawRing(rl.NewVector2(cx, cy), inner, outer, startDeg, endDeg, segs, col)
		return
	}
		aaMin, aaMax := covStrokeAAMin, covStrokeAAMax
	if outer-inner <= 2.5 {
		aaMin, aaMax = covHairlineAAMin, covHairlineAAMax
	} else if sharp {
		// Middle ground — C13 sharp (0.28–0.72) stair-stepped; full stroke was fuzzy.
		aaMin, aaMax = float32(0.50), float32(1.05)
	}
	drawCoverageAnnulus(cx, cy, inner, outer, startDeg, endDeg, col, aaMin, aaMax)
}

func drawCoverageAnnulus(cx, cy, inner, outer, startDeg, endDeg float32, col rl.Color, aaMin, aaMax float32) {
	pad := covAAPad
	side := (outer + pad) * 2
	dst := rl.NewRectangle(cx-outer-pad, cy-outer-pad, side, side)
	rl.BeginShaderMode(coverageRingShader)
	if covRingLocSize >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocSize, []float32{side, side}, rl.ShaderUniformVec2)
	}
	if covRingLocInner >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocInner, []float32{inner}, rl.ShaderUniformFloat)
	}
	if covRingLocOuter >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocOuter, []float32{outer}, rl.ShaderUniformFloat)
	}
	if covRingLocStart >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocStart, []float32{startDeg}, rl.ShaderUniformFloat)
	}
	if covRingLocEnd >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocEnd, []float32{endDeg}, rl.ShaderUniformFloat)
	}
	if covRingLocAAMin >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocAAMin, []float32{aaMin}, rl.ShaderUniformFloat)
	}
	if covRingLocAAMax >= 0 {
		rl.SetShaderValue(coverageRingShader, covRingLocAAMax, []float32{aaMax}, rl.ShaderUniformFloat)
	}
	src := rl.NewRectangle(0, 0, float32(covWhiteTex.Width), float32(covWhiteTex.Height))
	rl.DrawTexturePro(covWhiteTex, src, dst, rl.NewVector2(0, 0), 0, col)
	rl.EndShaderMode()
}

func cornerRoundnessFromRadius(r rl.Rectangle, cr float32) float32 {
	short := r.Width
	if r.Height < short {
		short = r.Height
	}
	if short < 1 {
		return 0
	}
	roundness := (cr * 2) / short
	if roundness > 1 {
		return 1
	}
	return roundness
}
