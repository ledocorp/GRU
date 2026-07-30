// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// MapPin is a normalized pin on a MapPlaceholder (0..1 in X/Y).
type MapPin struct {
	X, Y    float32
	Label   string
	Phosphor string
}

// MapPlaceholder shows a static map image with pin markers (Strategy 2 #30).
//
// # LLM Prompt Template
//
//	m := ui.NewMapPlaceholder("map", "", []ui.MapPin{{X: 0.3, Y: 0.4, Label: "HQ"}}, 0, 0, 400, 240)
//	panel.AddChild(m)
//
// Demo scenes: **Batch 10** (media panel).
type MapPlaceholder struct {
	Element
	MapPath string
	Pins    []MapPin
	mapImg  *Image
}

// NewMapPlaceholder creates a map placeholder. mapPath may be empty for a flat grid.
func NewMapPlaceholder(id, mapPath string, pins []MapPin, x, y, w, h float32) *MapPlaceholder {
	m := &MapPlaceholder{
		Element: NewElement(id, x, y, w, h),
		MapPath: mapPath,
		Pins:    pins,
	}
	m.styleName = "mapplaceholder"
	if mapPath != "" {
		m.mapImg = NewImage(id+"-map", mapPath, 0, 0, 0, 0)
		m.mapImg.SetParent(m)
	}
	return m
}

func (m *MapPlaceholder) Layout() {
	if m.IsHidden() {
		return
	}
	b := m.Bounds()
	if m.mapImg != nil {
		layoutSetBounds(m.mapImg, b)
		m.mapImg.Layout()
	}
	m.layoutDirty = false
}

func (m *MapPlaceholder) Update(dt float32) {
	if m.IsHidden() {
		return
	}
	m.Layout()
	if m.mapImg != nil {
		m.mapImg.Update(dt)
	}
}

func (m *MapPlaceholder) Draw() { m.drawInternal() }

func (m *MapPlaceholder) drawInternal() {
	if m.IsHidden() {
		return
	}
	b := m.Bounds()
	style := m.GetStyle()
	if m.mapImg != nil {
		m.mapImg.Draw()
	} else {
		bg := style.BackgroundColor
		if bg.A == 0 {
			bg = rl.NewColor(220, 235, 220, 255)
		}
		rl.DrawRectangleRec(b, bg)
		// simple grid
		grid := rl.NewColor(180, 200, 180, 120)
		for i := 1; i < 8; i++ {
			x := b.X + b.Width*float32(i)/8
			rl.DrawLineEx(rl.NewVector2(x, b.Y), rl.NewVector2(x, b.Y+b.Height), 1, grid)
			y := b.Y + b.Height*float32(i)/8
			rl.DrawLineEx(rl.NewVector2(b.X, y), rl.NewVector2(b.X+b.Width, y), 1, grid)
		}
	}
	for _, pin := range m.Pins {
		px := b.X + pin.X*b.Width
		py := b.Y + pin.Y*b.Height
		rl.DrawCircleV(rl.NewVector2(px, py), 6, rl.NewColor(220, 38, 38, 255))
		rl.DrawCircleV(rl.NewVector2(px, py), 3, rl.White)
		if pin.Phosphor != "" && Phosphor.Available(pin.Phosphor, PhosphorFill) {
			dst := snapPhosphorRect(rl.NewRectangle(px-10, py-28, 20, 20))
			Phosphor.Draw(dst, pin.Phosphor, PhosphorFill, rl.NewColor(79, 70, 229, 255))
		}
		if pin.Label != "" {
			ts := GetThemeStyle("form-label")
			ts.FontSize = 11
			drawTextS(pin.Label, int32(px+10), int32(py-6), ts)
		}
	}
}
