// Package ui (continued)
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	gaugeArcStart = float32(135)
	gaugeArcEnd   = float32(405) // 270° sweep
	gaugeSegments = int32(48)
)

// Gauge is a radial 0..1 progress indicator (Strategy 2 #26).
//
// # LLM Prompt Template
//
//	pct := ui.NewSignal(float32(0.72))
//	g := ui.NewGauge("disk", pct, "Disk", 0, 0, 96, 96)
//	row.AddChild(g)
//
// Demo scenes: **Batch 10**.
type Gauge struct {
	Element
	Value *Signal[float32]
	Label string
}

// NewGauge creates a gauge. value is clamped to [0,1] when drawn.
func NewGauge(id string, value *Signal[float32], label string, x, y, w, h float32) *Gauge {
	if value == nil {
		value = NewSignal(float32(0))
	}
	g := &Gauge{
		Element: NewElement(id, x, y, w, h),
		Value:   value,
		Label:   label,
	}
	g.styleName = "gauge"
	g.Value.Subscribe(func() { g.MarkDirty() })
	return g
}

func (g *Gauge) Update(_ float32) {}

func (g *Gauge) Layout() { g.layoutDirty = false }

func (g *Gauge) Draw() { g.drawInternal() }

func (g *Gauge) drawInternal() {
	if g.IsHidden() {
		return
	}
	b := g.Bounds()
	style := g.GetStyle()
	cx := b.X + b.Width/2
	cy := b.Y + b.Height/2
	slot := b.Width
	if b.Height < slot {
		slot = b.Height
	}
	outer := slot * 0.42
	inner := outer * 0.72

	track := style.BackgroundColor
	if track.A == 0 {
		track = rl.NewColor(230, 232, 240, 255)
	}
	fill := GetThemeStyle("progress-fill").BackgroundColor
	if fill.A == 0 {
		fill = rl.NewColor(79, 70, 229, 255)
	}

	center := rl.NewVector2(cx, cy)
	rl.DrawRing(center, inner, outer, gaugeArcStart, gaugeArcEnd, gaugeSegments, track)

	v := g.Value.Get()
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	if v > 0.001 {
		end := gaugeArcStart + (gaugeArcEnd-gaugeArcStart)*v
		rl.DrawRing(center, inner, outer, gaugeArcStart, end, gaugeSegments, fill)
	}

	pctStyle := style
	pctStyle.FontSize = 18
	pctStyle.Bold = true
	pctText := fmt.Sprintf("%.0f%%", v*100)
	pw := measureTextS(pctText, pctStyle)
	drawTextS(pctText, int32(cx-float32(pw)/2), int32(cy-10), pctStyle)

	if g.Label != "" {
		lbl := style
		lbl.FontSize = 13
		lbl.Bold = false
		lw := measureTextS(g.Label, lbl)
		drawTextS(g.Label, int32(cx-float32(lw)/2), int32(cy+outer*0.35), lbl)
	}
}
