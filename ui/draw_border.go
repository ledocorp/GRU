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

// DrawRoundedInsetBorder paints a rounded border as an outer fill minus an inset fill
// (crisper than DrawRectangleRoundedLinesEx at 1px under SSAA).
func DrawRoundedInsetBorder(outer rl.Rectangle, roundness, bw float32, borderCol, fillCol rl.Color) {
	drawRoundedInsetBorder(outer, roundness, bw, borderCol, fillCol)
}

// DrawRoundedFill draws a solid rounded rectangle with smooth pill corners.
func DrawRoundedFill(bounds rl.Rectangle, roundness float32, fill rl.Color) {
	bounds = snapControlRect(bounds)
	if roundness > 0 {
		rl.DrawRectangleRounded(bounds, roundness, 32, fill)
	} else {
		rl.DrawRectangleRec(bounds, fill)
	}
}

// drawRoundedInsetBorder paints a rounded border as an outer fill minus an inset fill
// (crisper than DrawRectangleRoundedLinesEx at 1px under SSAA).
//
// Use for compact controls where a hairline must read clearly: Checkbox, ColorPicker
// swatch, SearchBar, VirtualList frame, and similar square/rounded fields.
func drawRoundedInsetBorder(outer rl.Rectangle, roundness, bw float32, borderCol, fillCol rl.Color) {
	outer = snapControlRect(outer)
	if bw <= 0 {
		if roundness > 0 {
			rl.DrawRectangleRounded(outer, roundness, 32, fillCol)
		} else {
			rl.DrawRectangleRec(outer, fillCol)
		}
		return
	}
	if roundness > 0 {
		rl.DrawRectangleRounded(outer, roundness, 32, borderCol)
	} else {
		rl.DrawRectangleRec(outer, borderCol)
	}
	inner := rl.NewRectangle(outer.X+bw, outer.Y+bw, outer.Width-2*bw, outer.Height-2*bw)
	if inner.Width <= 0 || inner.Height <= 0 {
		return
	}
	innerRound := roundness
	if outer.Width > 0 && outer.Height > 0 {
		short := outer.Width
		if outer.Height < short {
			short = outer.Height
		}
		innerShort := inner.Width
		if inner.Height < innerShort {
			innerShort = inner.Height
		}
		if short > 0 && innerShort > 0 {
			r := (roundness * short / 2)
			if r > bw {
				r -= bw
			} else {
				r = 0
			}
			innerRound = (r * 2) / innerShort
			if innerRound > 1 {
				innerRound = 1
			}
		}
	}
	if innerRound > 0 {
		rl.DrawRectangleRounded(inner, innerRound, 32, fillCol)
	} else {
		rl.DrawRectangleRec(inner, fillCol)
	}
}
