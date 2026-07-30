package ui

import (
	"testing"
)

func TestGridRowSizingEqualFillDistributesViewportBand(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 600, 400)
	vp.AutoHeight = false
	vp.Gap = 12

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 1
	grid.AutoHeight = false
	grid.GridRowSizing = GridRowSizingEqualFill
	grid.SetFlexGrow(1)
	grid.Gap = 12

	for i := 0; i < 4; i++ {
		p := NewPanel("p", "Split", 0, 0, 0, 0)
		p.AutoHeight = false
		p.TitleHeight = 32
		p.AddChild(NewLabel("help", "Help", 0, 0, 0, 18))
		sv := NewSplitView("sv", SplitHorizontal,
			NewLabel("left", "L", 0, 0, 0, 24),
			NewLabel("right", "R", 0, 0, 0, 24),
			0, 0, 0, 0)
		sv.SetFlexGrow(1)
		p.AddChild(sv)
		grid.AddChild(p)
	}

	vp.AddChild(grid)
	vp.SetBounds(vp.Bounds())
	vp.Layout()

	heights := make([]float32, 0, 4)
	for _, ch := range grid.Children() {
		heights = append(heights, ch.Bounds().Height)
	}
	if len(heights) != 4 {
		t.Fatalf("children = %d", len(heights))
	}
	for i, h := range heights {
		if h < 60 || h > 120 {
			t.Fatalf("panel %d height %.0f outside equal-share band", i, h)
		}
	}
	sv := grid.Children()[0].(*Panel).Children()[1].(*SplitView)
	if sv.Bounds().Height > heights[0] {
		t.Fatalf("split height %.0f exceeds panel %.0f", sv.Bounds().Height, heights[0])
	}
}
