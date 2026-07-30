package ui

import (
	"testing"
)

func TestFlexGrowGridStretchesRowsToViewportBand(t *testing.T) {
	const w, h = float32(800), float32(700)
	shell := NewContainer("shell", 0, 0, w, h)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn

	vp := NewViewport("vp", 0, 0, 0, 0)
	vp.SetFlexGrow(1)

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	grid.SetFlexGrow(1)

	p := NewPanel("p", "Split", 0, 0, 0, 0)
	p.SetColSpan(BreakpointXS, 12)
	p.AddChild(NewLabel("help", "Help", 0, 0, 0, 18))
	left := NewLabel("left", "A", 0, 0, 0, 24)
	right := NewLabel("right", "B", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 0, 0)
	sv.SetFlexGrow(1)
	p.AddChild(sv)
	grid.AddChild(p)

	vp.AddChild(grid)
	shell.AddChild(vp)
	root := NewContainer("root", 0, 0, w, h)
	root.AddChild(shell)
	root.Layout()

	if p.Bounds().Height < 400 {
		t.Fatalf("panel height %.0f should grow with viewport band", p.Bounds().Height)
	}
	if sv.Bounds().Height < 300 {
		t.Fatalf("split height %.0f should fill stretched panel", sv.Bounds().Height)
	}
}
