// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	colorWellCellSize  float32 = 28
	colorWellCellGap   float32 = 6
	colorWellDefaultH  float32 = 28
	colorWellRingPad   float32 = 2
)

// ColorWell is a compact row of preset color swatches for forms. Click a swatch
// to set Value; the selected swatch shows a focus ring. Composes the same color
// signal pattern as ColorPicker without opening the full HSV popup.
//
// # LLM Prompt Template
//
//	swatches := []rl.Color{rl.Red, rl.Green, rl.Blue}
//	cw := ui.NewColorWell("accent", rl.Red, swatches, 0, 0, 0, 0)
//	cw.Value.Subscribe(func() { applyAccent(cw.Value.Get()) })
//	form.AddChild(cw)
//
// Demo scenes: **Batch 26** (HSV picker family).
type ColorWell struct {
	Element
	Value    *Signal[rl.Color]
	Swatches []rl.Color
	hoverIdx int
}

// NewColorWell creates a color well. Pass w=0 for flex width; h=0 for intrinsic height.
func NewColorWell(id string, initial rl.Color, swatches []rl.Color, x, y, w, h float32) *ColorWell {
	if len(swatches) == 0 {
		swatches = []rl.Color{
			rl.NewColor(239, 68, 68, 255),
			rl.NewColor(249, 115, 22, 255),
			rl.NewColor(234, 179, 8, 255),
			rl.NewColor(34, 197, 94, 255),
			rl.NewColor(59, 130, 246, 255),
			rl.NewColor(139, 92, 246, 255),
		}
	}
	cw := &ColorWell{
		Element:  NewElement(id, x, y, w, h),
		Value:    NewSignal(initial),
		Swatches: swatches,
		hoverIdx: -1,
	}
	cw.styleName = "colorwell"
	if h == 0 {
		cw.AutoHeight = true
	}
	cw.Value.Subscribe(func() { cw.MarkDrawDirty() })
	return cw
}

// IsInteractive implements Node.
func (cw *ColorWell) IsInteractive() bool { return true }

// GetPreferredWidth reports the intrinsic swatch-row width for flex/list layouts.
func (cw *ColorWell) GetPreferredWidth() float32 { return cw.intrinsicWidth() }

// GetPreferredHeight reports row height for list-tile trailing slots.
func (cw *ColorWell) GetPreferredHeight() float32 { return colorWellDefaultH }

func (cw *ColorWell) cellCount() int { return len(cw.Swatches) }

func (cw *ColorWell) intrinsicWidth() float32 {
	n := cw.cellCount()
	if n <= 0 {
		return colorWellCellSize
	}
	return float32(n)*colorWellCellSize + float32(n-1)*colorWellCellGap
}

func (cw *ColorWell) cellRect(index int) rl.Rectangle {
	b := cw.Bounds()
	x := b.X + float32(index)*(colorWellCellSize+colorWellCellGap)
	y := b.Y + (b.Height-colorWellCellSize)/2
	return rl.NewRectangle(x, y, colorWellCellSize, colorWellCellSize)
}

func (cw *ColorWell) selectedIndex() int {
	cur := cw.Value.Get()
	for i, c := range cw.Swatches {
		if colorsEqual(c, cur) {
			return i
		}
	}
	return -1
}

func colorsEqual(a, b rl.Color) bool {
	return a.R == b.R && a.G == b.G && a.B == b.B && a.A == b.A
}

// Layout sets intrinsic size when AutoHeight is enabled.
func (cw *ColorWell) Layout() {
	defer func() { cw.layoutDirty = false }()
	if !cw.IsAutoHeight() {
		return
	}
	b := cw.Bounds()
	wantW := cw.intrinsicWidth()
	wantH := colorWellDefaultH
	if b.Width < wantW-0.5 || b.Width > wantW+0.5 || b.Height < wantH-0.5 || b.Height > wantH+0.5 {
		if cw.bounds.Width <= 0 {
			b.Width = wantW
		}
		b.Height = wantH
		cw.setBoundsNoMark(b)
	}
}

// Update handles hover and swatch selection.
func (cw *ColorWell) Update(_ float32) {
	if cw.IsHidden() || cw.cellCount() == 0 {
		return
	}
	mouse := rl.GetMousePosition()
	prev := cw.hoverIdx
	cw.hoverIdx = -1
	for i := 0; i < cw.cellCount(); i++ {
		if rl.CheckCollisionPointRec(mouse, cw.cellRect(i)) {
			cw.hoverIdx = i
			break
		}
	}
	if cw.hoverIdx != prev {
		cw.MarkDrawDirty()
	}
	if cw.hoverIdx < 0 {
		return
	}
	if PointerClickConsume(cw.cellRect(cw.hoverIdx)) {
		cw.Value.Set(cw.Swatches[cw.hoverIdx])
	}
}

// Draw implements Node.Draw.
func (cw *ColorWell) Draw() { cw.drawInternal() }

func (cw *ColorWell) drawInternal() {
	if cw.IsHidden() || cw.cellCount() == 0 {
		return
	}
	sel := cw.selectedIndex()
	for i, col := range cw.Swatches {
		r := cw.cellRect(i)
		if i == sel || i == cw.hoverIdx {
			ring := rl.NewRectangle(r.X-colorWellRingPad, r.Y-colorWellRingPad,
				r.Width+colorWellRingPad*2, r.Height+colorWellRingPad*2)
			rc := rl.NewColor(79, 70, 229, 255)
			if i == cw.hoverIdx && i != sel {
				rc.A = 180
			}
			rl.DrawRectangleRounded(ring, 0.25, 6, rc)
		}
		rl.DrawRectangleRounded(r, 0.2, 6, col)
		bc := rl.NewColor(0, 0, 0, 40)
		rl.DrawRectangleRoundedLinesEx(r, 0.2, 6, 1, bc)
	}
}
