// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// drawPopupRoundedBorder strokes a floating overlay (calendar, menu popup) on the
// outer layout bounds. Unlike chromeBorderBounds inset, this keeps all four edges
// equally visible when no parent scissor clips the right gutter.
func drawPopupRoundedBorder(bounds rl.Rectangle, cornerRadius, lineWidth float32, color rl.Color) {
	if lineWidth <= 0 {
		return
	}
	snap := rl.NewRectangle(
		float32(int32(bounds.X+0.5)),
		float32(int32(bounds.Y+0.5)),
		float32(int32(bounds.Width+0.5)),
		float32(int32(bounds.Height+0.5)),
	)
	roundness := chromeRoundness(snap, cornerRadius)
	rl.DrawRectangleRoundedLinesEx(snap, roundness, 16, lineWidth, color)
}

// chromeBorderBounds insets a rectangle so rounded/box border strokes stay inside
// the layout bounds (raylib draws centered on the path; without inset the left
// stroke is sheared by viewport scissors — the right side often looked fine only
// because the scrollbar gutter left extra room).
func chromeBorderBounds(bounds rl.Rectangle, borderWidth float32) rl.Rectangle {
	if borderWidth <= 0 {
		return bounds
	}
	return rl.NewRectangle(
		bounds.X+borderWidth,
		bounds.Y+borderWidth,
		bounds.Width-2*borderWidth,
		bounds.Height-2*borderWidth,
	)
}

// chromeExpandBounds expands a rectangle outward (negative inset).
func chromeExpandBounds(bounds rl.Rectangle, expand float32) rl.Rectangle {
	if expand <= 0 {
		return bounds
	}
	return rl.NewRectangle(
		bounds.X-expand,
		bounds.Y-expand,
		bounds.Width+2*expand,
		bounds.Height+2*expand,
	)
}

// RaisedSurfaceShadowBleed is extra space below a nested Card inside a parent
// Panel/Card body so drawRaisedSurfaceShadow (extends ~6px below the child bounds)
// is not clipped by the parent's padded body scissor.
const RaisedSurfaceShadowBleed = float32(8)

// SurfaceGlowBleed is the maximum clip expansion per side for nested Cards with
// preset glow so outer halo rings are not sheared by the parent body scissor.
const SurfaceGlowBleed = float32(10)

// chromeFillBounds insets the fill rect so background color stays inside an inset
// border stroke (raylib draws lines centered on the path; a full-bleed fill
// otherwise shows past the visible outer edge of the border).
func chromeFillBounds(bounds rl.Rectangle, borderWidth float32) rl.Rectangle {
	return chromeBorderBounds(bounds, borderWidth)
}

// chromeRoundness converts a pixel corner radius into raylib's 0..1 roundness.
func chromeRoundness(bounds rl.Rectangle, cornerRadius float32) float32 {
	if cornerRadius <= 0 {
		return 0
	}
	shorter := bounds.Width
	if bounds.Height < shorter {
		shorter = bounds.Height
	}
	if shorter <= 0 {
		return 0
	}
	r := cornerRadius / (shorter / 2)
	if r > 1 {
		return 1
	}
	return r
}

// drawSurfacePresetGlow draws a soft indigo halo for neo-glow presets.
// Uses low-alpha filled rings for blur-like falloff (no shader). Real backdrop
// blur remains a Phase 4 option — see docs/VISUAL_PRESETS_PLAN.md §11.
func drawSurfacePresetGlow(bounds rl.Rectangle, cornerRadius float32, intensity float32, borderWidth float32) {
	if intensity <= 0 {
		return
	}
	scale := 0.45 + 0.55*intensity

	// Soft bloom — low-alpha filled rings (blur-like falloff without a shader).
	for i, expand := range []float32{1, 2.5, 4.5, 7} {
		exp := expand * scale
		b := chromeExpandBounds(bounds, exp)
		r := chromeRoundness(b, cornerRadius+exp*0.25)
		alphas := []uint8{20, 13, 8, 4}
		a := uint8(float32(alphas[i]) * intensity)
		col := rl.NewColor(129, 140, 248, a)
		if r > 0 {
			rl.DrawRectangleRounded(b, r, 8, col)
		} else {
			rl.DrawRectangleRec(b, col)
		}
	}

	// Single subtle edge accent on the border path.
	if borderWidth > 0 {
		b := chromeBorderBounds(bounds, borderWidth*0.35)
		r := chromeRoundness(b, cornerRadius-borderWidth*0.35)
		col := rl.NewColor(165, 180, 252, uint8(42*intensity))
		if r > 0 {
			rl.DrawRectangleRoundedLinesEx(b, r, 10, 1.2, col)
		} else {
			rl.DrawRectangleLinesEx(b, 1.2, col)
		}
	}
}

// drawNeoGlowAmbientShadow adds indigo-tinted soft layers under neo-glow cards.
func drawNeoGlowAmbientShadow(bounds rl.Rectangle, roundness float32, intensity float32, hoverLift bool) {
	if intensity <= 0 {
		return
	}
	const seg int32 = 8
	lift := float32(0)
	if hoverLift {
		lift = -2
	}
	for i, yExtra := range []float32{4, 8} {
		b := rl.NewRectangle(bounds.X+float32(i), bounds.Y+6+lift+yExtra*intensity*0.5, bounds.Width+2-float32(i), bounds.Height)
		a := uint8(float32([]uint8{32, 20}[i]) * intensity)
		col := rl.NewColor(79, 70, 229, a)
		if roundness > 0 {
			rl.DrawRectangleRounded(b, roundness, seg, col)
		} else {
			rl.DrawRectangleRec(b, col)
		}
	}
}

// drawGlassPanelSheen adds a native top gradient highlight inside the border.
func drawGlassPanelSheen(bounds rl.Rectangle, roundness float32, style Style) {
	inner := chromeFillBounds(bounds, style.BorderWidth)
	if inner.Width < 2 || inner.Height < 2 {
		return
	}
	sheenH := inner.Height * 0.34
	if sheenH < 6 {
		sheenH = 6
	}
	if sheenH > inner.Height-2 {
		sheenH = inner.Height - 2
	}
	sheen := rl.NewRectangle(inner.X, inner.Y, inner.Width, sheenH)
	DrawCornerGradientRect(
		sheen,
		rl.NewColor(255, 255, 255, 48),
		rl.NewColor(255, 255, 255, 0),
		rl.NewColor(255, 255, 255, 32),
		rl.NewColor(255, 255, 255, 0),
	)
	_ = roundness
	highlightY := inner.Y + 1.5
	rl.DrawLine(
		int32(inner.X+2), int32(highlightY),
		int32(inner.X+inner.Width-2), int32(highlightY),
		rl.NewColor(255, 255, 255, 90),
	)
}

// drawFlatBottomChromeFill paints a fill with rounded top corners and square bottom
// corners — used for standalone HeaderBand strips (no body below).
func drawFlatBottomChromeFill(fillBounds rl.Rectangle, cornerRadius float32, color rl.Color) {
	r := chromeRoundness(fillBounds, cornerRadius)
	if r <= 0 {
		rl.DrawRectangleRec(fillBounds, color)
		return
	}
	rl.DrawRectangleRounded(fillBounds, r, 6, color)
	roundZone := cornerRadius
	if roundZone > fillBounds.Height/2 {
		roundZone = fillBounds.Height / 2
	}
	if roundZone > 0 && roundZone < fillBounds.Height {
		square := rl.NewRectangle(
			fillBounds.X,
			fillBounds.Y+roundZone,
			fillBounds.Width,
			fillBounds.Height-roundZone,
		)
		rl.DrawRectangleRec(square, color)
	}
}

// drawFlatBottomChromeBorder strokes a flat-bottom chrome outline (square bottom corners).
func drawFlatBottomChromeBorder(bounds rl.Rectangle, borderWidth, cornerRadius float32, color rl.Color) {
	if borderWidth <= 0 {
		return
	}
	x, y, w, h := bounds.X, bounds.Y, bounds.Width, bounds.Height
	cr := cornerRadius
	if cr > h/2 {
		cr = h / 2
	}
	ix, iy := int32(x), int32(y)
	iw, ih := int32(w), int32(h)
	ibw := int32(borderWidth + 0.5)
	icr := int32(cr + 0.5)
	rl.DrawLine(ix+icr, iy, ix+iw-icr, iy, color)
	rl.DrawLine(ix, iy+icr, ix, iy+ih-ibw, color)
	rl.DrawLine(ix+iw-ibw, iy+icr, ix+iw-ibw, iy+ih-ibw, color)
	rl.DrawLine(ix, iy+ih-ibw, ix+iw, iy+ih-ibw, color)
}

// drawRaisedSurfaceShadow draws the shared soft drop stack under Panel and Card.
//
// Soft DrawRectangleRounded layers (not seam-safe circles): low-alpha corner
// circles over rect strips produced blocky L-artifacts under SSAA on hello's card.
// Mid-seams in translucent shadow are not visible the way they are on opaque fills.
func drawRaisedSurfaceShadow(bounds rl.Rectangle, roundness float32) {
	outer := rl.NewRectangle(bounds.X+2, bounds.Y+5, bounds.Width+1, bounds.Height+1)
	inner := rl.NewRectangle(bounds.X+3, bounds.Y+4, bounds.Width, bounds.Height)
	if roundness > 0.001 {
		rl.DrawRectangleRounded(outer, roundness, 16, rl.NewColor(0, 0, 0, 14))
		rl.DrawRectangleRounded(inner, roundness, 16, rl.NewColor(0, 0, 0, 26))
		return
	}
	rl.DrawRectangleRec(outer, rl.NewColor(0, 0, 0, 14))
	rl.DrawRectangleRec(inner, rl.NewColor(0, 0, 0, 26))
}
