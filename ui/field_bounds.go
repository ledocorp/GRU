// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// fieldClipInset is extra inset beyond [Style.BorderWidth] so rounded field
// strokes stay fully inside parent Panel body scissors.
const fieldClipInset = float32(1)

// effectiveFieldFaceStyle tightens font size and padding so face text fits compact fields.
func effectiveFieldFaceStyle(bounds rl.Rectangle, style Style) Style {
	pad := toolbarStylePadding(style)
	border := style.BorderWidth
	if border < 1 {
		border = 1
	}
	maxFS := bounds.Height - 2*pad - 2*border - 2
	if maxFS < 12 {
		maxFS = 12
	}
	// Cap rendered size via MaxFontSize — do not rewrite FontSize (tokens are
	// rem-scaled by EffectiveFontSize; assigning pixel caps there overshoots).
	if fs := EffectiveFontSize(style); fs > maxFS {
		style.MaxFontSize = int32(maxFS)
	}
	if bounds.Height <= 34 && style.Padding > 4 {
		style.Padding = 4
	}
	return style
}

// fieldPaintBounds returns bounds inset for drawing bordered field chrome
// (DatePicker, Dropdown, TextInput). Layout/hit-test bounds stay unchanged.
func fieldPaintBounds(bounds rl.Rectangle, style Style) rl.Rectangle {
	inset := style.BorderWidth + fieldClipInset
	if inset < fieldClipInset {
		inset = fieldClipInset
	}
	if bounds.Width <= 2*inset || bounds.Height <= 2*inset {
		return bounds
	}
	return rl.NewRectangle(bounds.X+inset, bounds.Y+inset, bounds.Width-2*inset, bounds.Height-2*inset)
}
