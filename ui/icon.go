// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Icon is a Phosphor glyph. Set OnClick for a chromeless clickable icon (no button slab).
//
// Example:
//
//	ic := ui.NewIcon("nav-home", ui.PhosphorHouse, ui.PhosphorRegular, 0, 0, 24, 24)
type Icon struct {
	Element
	Name    string
	Weight  PhosphorWeight
	OnClick func()
	hovered bool
}

// NewIcon creates an Icon. Pass equal w and h for a square glyph.
func NewIcon(id, name string, weight PhosphorWeight, x, y, w, h float32) *Icon {
	ic := &Icon{
		Element: NewElement(id, x, y, w, h),
		Name:    name,
		Weight:  weight,
	}
	if weight == "" {
		ic.Weight = PhosphorRegular
	}
	return ic
}

func (ic *Icon) Update(_ float32) {
	if ic.IsHidden() || ic.OnClick == nil {
		return
	}
	mouse := rl.GetMousePosition()
	hovered := rl.CheckCollisionPointRec(mouse, ic.Bounds())
	if hovered != ic.hovered {
		ic.hovered = hovered
		ic.MarkDrawDirty()
	}
	if hovered {
		rl.SetMouseCursor(rl.MouseCursorPointingHand)
		if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
			ic.OnClick()
		}
	}
}

func (ic *Icon) Layout() { ic.layoutDirty = false }

func (ic *Icon) Draw() {
	if ic.IsHidden() || ic.Name == "" {
		return
	}
	col := ic.GetStyle().TextColor
	if col.A == 0 {
		col = rl.NewColor(40, 42, 58, 255)
	}
	b := ic.Bounds()
	slot := b.Width
	if b.Height < slot {
		slot = b.Height
	}
	size := phosphorIconDrawSize(slot, 0)
	dst := snapPhosphorRect(rl.NewRectangle(
		b.X+(b.Width-size)/2,
		b.Y+(b.Height-size)/2,
		size, size,
	))
	Phosphor.EnsureLoaded(ic.Name, ic.Weight)
	DrawPhosphorIcon(dst, ic.Name, ic.Weight, col)
}

func (ic *Icon) IsInteractive() bool { return ic.OnClick != nil }
