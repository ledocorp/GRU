// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ChartKind selects line or bar rendering.
type ChartKind int

const (
	ChartLine ChartKind = iota
	ChartBar
)

// Chart plots a float series in a bounded rect (Strategy 2 #25).
//
// # LLM Prompt Template
//
//	series := []float32{12, 28, 22, 40, 35}
//	c := ui.NewChart("cpu", ui.ChartLine, series, 0, 0, 400, 160)
//	c.YMin, c.YMax = 0, 50
//	panel.AddChild(c)
//
// Demo scenes: **Batch 10**, **Widget Showcase**.
type Chart struct {
	Element
	Kind   ChartKind
	Series []float32
	YMin   float32
	YMax   float32
	Title  string
}

// NewChart creates a chart. series may be nil (draws empty frame).
func NewChart(id string, kind ChartKind, series []float32, x, y, w, h float32) *Chart {
	c := &Chart{
		Element: NewElement(id, x, y, w, h),
		Kind:    kind,
		Series:  series,
		YMin:    0,
		YMax:    100,
	}
	c.styleName = "chart"
	return c
}

// SetSeries replaces data and marks dirty.
func (c *Chart) SetSeries(data []float32) {
	c.Series = data
	c.MarkDirty()
}

func (c *Chart) Update(_ float32) {}

func (c *Chart) Layout() { c.layoutDirty = false }

func (c *Chart) Draw() { c.drawInternal() }

func (c *Chart) plotRect() rl.Rectangle {
	b := c.Bounds()
	pad := float32(8)
	top := float32(24)
	if c.Title == "" {
		top = pad
	}
	return rl.NewRectangle(b.X+pad, b.Y+top, b.Width-2*pad, b.Height-top-pad)
}

func (c *Chart) mapY(v float32, plot rl.Rectangle) float32 {
	span := c.YMax - c.YMin
	if span <= 0 {
		span = 1
	}
	norm := (v - c.YMin) / span
	if norm < 0 {
		norm = 0
	}
	if norm > 1 {
		norm = 1
	}
	return plot.Y + plot.Height*(1-norm)
}

func (c *Chart) drawInternal() {
	if c.IsHidden() {
		return
	}
	b := c.Bounds()
	style := c.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRounded(b, 0.04, 8, style.BackgroundColor)
	}
	if style.BorderWidth > 0 && style.BorderColor.A > 0 {
		rl.DrawRectangleRoundedLinesEx(b, 0.04, 8, style.BorderWidth, style.BorderColor)
	}
	if c.Title != "" {
		ts := style
		ts.FontSize = 14
		ts.Bold = true
		drawTextS(c.Title, int32(b.X+10), int32(b.Y+6), ts)
	}

	plot := c.plotRect()
	grid := rl.NewColor(220, 224, 235, 255)
	rl.DrawRectangleLinesEx(plot, 1, grid)

	lineCol := GetThemeStyle("progress-fill").BackgroundColor
	if lineCol.A == 0 {
		lineCol = rl.NewColor(79, 70, 229, 255)
	}
	barCol := rl.NewColor(99, 102, 241, 220)

	n := len(c.Series)
	if n == 0 {
		hint := GetThemeStyle("form-value")
		drawTextS("No data", int32(plot.X+8), int32(plot.Y+plot.Height/2-8), hint)
		return
	}

	switch c.Kind {
	case ChartBar:
		barW := plot.Width / float32(n)
		if barW < 2 {
			barW = 2
		}
		gap := barW * 0.15
		barW -= gap
		for i, v := range c.Series {
			x := plot.X + float32(i)* (plot.Width/float32(n)) + gap/2
			y := c.mapY(v, plot)
			h := plot.Y + plot.Height - y
			if h < 1 {
				h = 1
			}
			rl.DrawRectangleRec(rl.NewRectangle(x, y, barW, h), barCol)
		}
	default:
		if n == 1 {
			x := plot.X + plot.Width/2
			y := c.mapY(c.Series[0], plot)
			rl.DrawCircleV(rl.NewVector2(x, y), 4, lineCol)
			return
		}
		for i := 1; i < n; i++ {
			x0 := plot.X + plot.Width*float32(i-1)/float32(n-1)
			x1 := plot.X + plot.Width*float32(i)/float32(n-1)
			y0 := c.mapY(c.Series[i-1], plot)
			y1 := c.mapY(c.Series[i], plot)
			rl.DrawLineEx(rl.NewVector2(x0, y0), rl.NewVector2(x1, y1), 2, lineCol)
		}
	}
}
