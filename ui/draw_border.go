// Package ui (continued) — crisp control borders on the logical pixel grid.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// snapControlRect aligns a control rect to whole logical pixels (reduces SSAA fringe on 1px chrome).
func snapControlRect(r rl.Rectangle) rl.Rectangle {
	return rl.NewRectangle(
		float32(int32(r.X+0.5)),
		float32(int32(r.Y+0.5)),
		float32(int32(r.Width+0.5)),
		float32(int32(r.Height+0.5)),
	)
}

// SnapControlRect aligns a control rect to whole logical pixels.
func SnapControlRect(r rl.Rectangle) rl.Rectangle {
	return snapControlRect(r)
}

// snapChromeBorderWidth maps theme BorderWidth to device px.
// SSAA 1×: 1.5 stays 1px (round-up to 2 made Panel/Card frames look doubled).
// SSAA >1: still round to nearest whole pixel for the downsample.
func snapChromeBorderWidth(bw float32) float32 {
	if bw <= 0 {
		return 0
	}
	if RenderScale <= 1.01 {
		if bw <= 1.5 {
			return 1
		}
		out := float32(int32(bw + 0.5))
		if out < 1 {
			return 1
		}
		return out
	}
	out := float32(int32(bw + 0.5))
	if out < 1 {
		return 1
	}
	return out
}

// DrawRoundedTessInsetBorder paints a 1px rounded frame via tessellated punch-out
// (outer border-colored rounded rect, then inset fill). No coverage-AA fringe and
// no LinesEx stroke.
//
// Mastered 1px recipe for Studio footer chips (C10) and page controls that share
// drawRoundedInsetBorder / drawSeamSafeInsetBorder (checkbox, TextInput, ListTile,
// Panel/Card). See docs/PRESENT_1X_COMPOSITE.md §4 C10–C11.
//
// LLM Prompt Template — 1px chip / control:
//   btn := ui.SnapControlRect(...)
//   ui.DrawRoundedTessInsetBorder(btn, 0.25, 1, borderCol, fillCol)
//   pop := ui.PushLCDSurfaceBackdrop(fillCol)
//   ui.DrawChromeText(...); ui.PopLCDSurfaceBackdrop(pop)
func DrawRoundedTessInsetBorder(outer rl.Rectangle, roundness, bw float32, borderCol, fillCol rl.Color) {
	outer = snapControlRect(outer)
	if bw <= 0 {
		if roundness > 0 {
			rl.DrawRectangleRounded(outer, roundness, 48, fillCol)
		} else {
			rl.DrawRectangleRec(outer, fillCol)
		}
		return
	}
	bw = snapChromeBorderWidth(bw)
	if roundness > 0 {
		rl.DrawRectangleRounded(outer, roundness, 64, borderCol)
	} else {
		rl.DrawRectangleRec(outer, borderCol)
	}
	inner := rl.NewRectangle(outer.X+bw, outer.Y+bw, outer.Width-2*bw, outer.Height-2*bw)
	if inner.Width <= 0 || inner.Height <= 0 {
		return
	}
	innerRound := insetRoundness(outer, roundness, bw)
	if innerRound > 0 {
		rl.DrawRectangleRounded(inner, innerRound, 64, fillCol)
	} else {
		rl.DrawRectangleRec(inner, fillCol)
	}
}

// DrawRoundedInsetBorder paints a rounded border as an outer fill minus an inset fill
// (crisper than DrawRectangleRoundedLinesEx at 1px under SSAA).
func DrawRoundedInsetBorder(outer rl.Rectangle, roundness, bw float32, borderCol, fillCol rl.Color) {
	drawRoundedInsetBorder(outer, roundness, bw, borderCol, fillCol)
}

// DrawRoundedFill draws a solid rounded rectangle with smooth pill corners.
// SSAA 1×: G1.1 coverage SDF — same family as Avatar’s DrawCoverageCircle (C21).
// C20 strip+disc removed AA and stair-stepped buttons/checkboxes; do not revive.
func DrawRoundedFill(bounds rl.Rectangle, roundness float32, fill rl.Color) {
	drawRoundedFillAA(bounds, roundness, fill, covFillAAMin, covFillAAMax)
}

func drawRoundedFillAA(bounds rl.Rectangle, roundness float32, fill rl.Color, aaMin, aaMax float32) {
	bounds = snapControlRect(bounds)
	if roundness <= 0 {
		rl.DrawRectangleRec(bounds, fill)
		return
	}
	short := bounds.Width
	if bounds.Height < short {
		short = bounds.Height
	}
	cr := roundness * short / 2
	if PreferGeometryCoverageAA() {
		drawCoverageRoundedFillAA(bounds, cr, aaMin, aaMax, fill)
		return
	}
	rl.DrawRectangleRounded(bounds, roundness, 48, fill)
}

func insetRoundness(outer rl.Rectangle, roundness, bw float32) float32 {
	if roundness <= 0 || outer.Width <= 0 || outer.Height <= 0 {
		return 0
	}
	innerW := outer.Width - 2*bw
	innerH := outer.Height - 2*bw
	if innerW <= 0 || innerH <= 0 {
		return 0
	}
	short := outer.Width
	if outer.Height < short {
		short = outer.Height
	}
	innerShort := innerW
	if innerH < innerShort {
		innerShort = innerH
	}
	r := (roundness * short / 2)
	if r > bw {
		r -= bw
	} else {
		r = 0
	}
	out := (r * 2) / innerShort
	if out > 1 {
		return 1
	}
	return out
}

// drawRoundedInsetBorder — SSAA 1× uses tessellated punch-out (C10/C11).
// Coverage fill + LinesEx left an uneven dark fringe on checkbox / TextInput / ListTile.
// Dual coverage punch-out still reads ~2px — do not revive that.
func drawRoundedInsetBorder(outer rl.Rectangle, roundness, bw float32, borderCol, fillCol rl.Color) {
	outer = snapControlRect(outer)
	if bw <= 0 {
		if roundness > 0 {
			DrawRoundedFill(outer, roundness, fillCol)
		} else {
			rl.DrawRectangleRec(outer, fillCol)
		}
		return
	}
	bw = snapChromeBorderWidth(bw)
	if PreferGeometryCoverageAA() && bw <= 1.01 {
		short := outer.Width
		if outer.Height < short {
			short = outer.Height
		}
		// Large chrome (panels): tess punch stays crisp. Tiny controls (checkbox
		// 18–24, soft chips): tess stair-steps — coverage punch matches Avatar (C21).
		if short <= 36 {
			DrawRoundedFill(outer, roundness, borderCol)
			inner := rl.NewRectangle(outer.X+bw, outer.Y+bw, outer.Width-2*bw, outer.Height-2*bw)
			if inner.Width > 0 && inner.Height > 0 {
				DrawRoundedFill(inner, insetRoundness(outer, roundness, bw), fillCol)
			}
			return
		}
		DrawRoundedTessInsetBorder(outer, roundness, bw, borderCol, fillCol)
		return
	}
	// SSAA >1 or thick borders: coverage punch-out (downsample softens the edge).
	if roundness > 0 {
		DrawRoundedFill(outer, roundness, borderCol)
	} else {
		rl.DrawRectangleRec(outer, borderCol)
	}
	inner := rl.NewRectangle(outer.X+bw, outer.Y+bw, outer.Width-2*bw, outer.Height-2*bw)
	if inner.Width <= 0 || inner.Height <= 0 {
		return
	}
	innerRound := insetRoundness(outer, roundness, bw)
	if innerRound > 0 {
		DrawRoundedFill(inner, innerRound, fillCol)
	} else {
		rl.DrawRectangleRec(inner, fillCol)
	}
}
