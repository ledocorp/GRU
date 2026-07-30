// Package ui (continued) — SSAA-safe fills for large raised surfaces.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// drawSeamSafeRoundedFill paints a rounded rect with center/side strips + corner
// circles. Avoids DrawRectangleRounded triangulation, which leaves a vertical
// mid-seam under 2× SSAA on wide empty Panel/Card fills (calc Desk card).
func drawSeamSafeRoundedFill(r rl.Rectangle, cornerRadius float32, col rl.Color) {
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
	// Horizontal band through the middle (full width).
	rl.DrawRectangleRec(rl.NewRectangle(r.X+cr, r.Y, r.Width-2*cr, r.Height), col)
	// Left / right strips between corner circles.
	rl.DrawRectangleRec(rl.NewRectangle(r.X, r.Y+cr, cr, r.Height-2*cr), col)
	rl.DrawRectangleRec(rl.NewRectangle(r.X+r.Width-cr, r.Y+cr, cr, r.Height-2*cr), col)
	// Corners.
	rl.DrawCircleV(rl.NewVector2(r.X+cr, r.Y+cr), cr, col)
	rl.DrawCircleV(rl.NewVector2(r.X+r.Width-cr, r.Y+cr), cr, col)
	rl.DrawCircleV(rl.NewVector2(r.X+cr, r.Y+r.Height-cr), cr, col)
	rl.DrawCircleV(rl.NewVector2(r.X+r.Width-cr, r.Y+r.Height-cr), cr, col)
}

// drawSeamSafeInsetBorder paints border+fill without DrawRectangleRounded seams.
func drawSeamSafeInsetBorder(outer rl.Rectangle, cornerRadius, bw float32, borderCol, fillCol rl.Color) {
	outer = snapControlRect(outer)
	if bw <= 0 {
		drawSeamSafeRoundedFill(outer, cornerRadius, fillCol)
		return
	}
	bw = float32(int32(bw + 0.5))
	if bw < 1 {
		bw = 1
	}
	drawSeamSafeRoundedFill(outer, cornerRadius, borderCol)
	inner := rl.NewRectangle(outer.X+bw, outer.Y+bw, outer.Width-2*bw, outer.Height-2*bw)
	if inner.Width < 1 || inner.Height < 1 {
		return
	}
	innerCR := cornerRadius - bw
	if innerCR < 0 {
		innerCR = 0
	}
	drawSeamSafeRoundedFill(inner, innerCR, fillCol)
}
