// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Separator is a non-interactive horizontal rule, optionally labelled.
//
// When Label is non-empty the text is centred on the line and the line is
// drawn in two segments on either side of the text.  When Label is empty a
// plain rule spanning the full widget width is drawn.
//
// Typical sizes:
//
//	ui.NewSeparator("s1", "",       0, 0, 800, 12)  // plain — 12 px tall
//	ui.NewSeparator("s2", "─ Options ─", 0, 0, 800, 28)  // labelled — 28 px
type Separator struct {
	Element
	Label string // centred text; empty = plain rule
}

// NewSeparator creates a Separator. Pass an empty label for a plain rule.
func NewSeparator(id, label string, x, y, w, h float32) *Separator {
	s := &Separator{
		Element: NewElement(id, x, y, w, h),
		Label:   label,
	}
	s.styleName = "separator"
	return s
}

// Update is a no-op — Separator is not interactive.
func (s *Separator) Update(_ float32) {}

// Layout is a no-op for leaf widgets.
func (s *Separator) Layout() {
	s.layoutDirty = false
}

// Draw implements Node.Draw.
func (s *Separator) Draw() {
	defer func() { s.drawDirty = false }()
	s.drawInternal()
}

func (s *Separator) drawInternal() {
	if s.IsHidden() {
		return
	}

	style := s.GetStyle()
	b := s.Bounds()
	midY := b.Y + b.Height/2

	lw := style.BorderWidth
	if lw <= 0 {
		lw = 1
	}
	col := style.BorderColor

	if s.Label == "" {
		rl.DrawLineEx(
			rl.NewVector2(b.X, midY),
			rl.NewVector2(b.X+b.Width, midY),
			lw, col,
		)
		return
	}

	// Labelled: text centred with line segments on each side.
	textW := float32(measureTextS(s.Label, style))
	gap := float32(10)
	textX := b.X + (b.Width-textW)/2
	ink := TextInkHeight(style)
	if ink < 1 {
		ink = EffectiveFontSize(style)
	}
	textY := int32(midY - ink/2)

	if textX > b.X+gap {
		rl.DrawLineEx(
			rl.NewVector2(b.X, midY),
			rl.NewVector2(textX-gap, midY),
			lw, col,
		)
	}
	drawTextS(s.Label, int32(textX), textY, style)

	rightStart := textX + textW + gap
	if rightStart < b.X+b.Width {
		rl.DrawLineEx(
			rl.NewVector2(rightStart, midY),
			rl.NewVector2(b.X+b.Width, midY),
			lw, col,
		)
	}
}

// IsInteractive implements Node.IsInteractive.
func (s *Separator) IsInteractive() bool { return false }
