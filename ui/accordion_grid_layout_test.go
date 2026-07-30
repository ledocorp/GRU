package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAccordionGridDoesNotStretchRowGaps(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 400, 800)
	vp.AutoHeight = false
	vp.Gap = 12

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.AutoHeight = true
	grid.SetFlexGrow(1)
	grid.Gap = 12

	for i := 0; i < 3; i++ {
		p := NewPanel("p", "Panel", 0, 0, 0, 0)
		p.AutoHeight = true
		p.TitleHeight = 32
		acc := NewAccordion("acc", "Section", 0, 0, 0, 0)
		acc.AddChild(NewLabel("lbl", "Short body copy.", 0, 0, 0, 18))
		p.AddChild(acc)
		grid.AddChild(p)
	}

	vp.AddChild(grid)
	vp.SetBounds(rl.NewRectangle(0, 0, 400, 800))
	vp.Layout()

	h0 := grid.Children()[0].Bounds().Height
	h1 := grid.Children()[1].Bounds().Height
	gap := grid.Children()[1].Bounds().Y - (grid.Children()[0].Bounds().Y + h0)
	if gap > grid.Gap+40 {
		t.Fatalf("row gap %.0f too large (grid gap %.0f)", gap, grid.Gap)
	}
	if h0 > 200 {
		t.Fatalf("accordion panel height %.0f should shrink-wrap, not fill viewport", h0)
	}
	_ = h1
}
