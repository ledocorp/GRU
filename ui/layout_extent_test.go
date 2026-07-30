package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestGridRowSizingShrinkWrapDoesNotEqualFill(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 600, 400)
	vp.AutoHeight = false

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 1
	grid.AutoHeight = false
	grid.GridRowSizing = GridRowSizingShrinkWrap
	grid.SetFlexGrow(1)

	p := NewPanel("p", "Short", 0, 0, 0, 0)
	p.AddChild(NewLabel("lbl", "One line", 0, 0, 0, 18))
	grid.AddChild(p)

	vp.AddChild(grid)
	vp.SetBounds(vp.Bounds())
	vp.Layout()

	// Shrink-wrap: panel stays intrinsic; grid may still fill viewport via FlexGrow.
	if p.Bounds().Height > 120 {
		t.Fatalf("panel height %.0f should shrink-wrap, not equal-fill", p.Bounds().Height)
	}
}

func TestGridPreOpenedAccordionUsesLayoutExtent(t *testing.T) {
	grid := NewContainer("grid", 0, 0, 320, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.AutoHeight = true
	grid.Gap = 12

	panel := NewPanel("p", "Acc", 0, 0, 0, 0)
	panel.SetColSpan(BreakpointXS, 12)
	panel.Gap = 6

	acc := NewAccordion("acc", "Open", 0, 0, 0, 0)
	for i := 0; i < 4; i++ {
		lbl := NewLabel("l", "Wrapped label line content here", 0, 0, 0, 0)
		lbl.AutoHeight = true
		lbl.Wrap = true
		acc.AddChild(lbl)
	}
	acc.Expanded.Set(true)
	panel.AddChild(acc)
	grid.AddChild(panel)

	grid.SetBounds(grid.Bounds())
	grid.Layout()

	panelBottom := panel.Bounds().Y + panel.Bounds().Height
	accBottom := nodeSubtreeBottom(acc)
	if accBottom > panelBottom+2 {
		t.Fatalf("pre-open accordion bottom %.0f exceeds panel %.0f", accBottom, panelBottom)
	}
}

func TestSyncLayoutExtentMarksParentOnChange(t *testing.T) {
	parent := NewContainer("parent", 0, 0, 200, 0)
	parent.AutoHeight = true
	child := NewLabel("c", "Hi", 0, 0, 0, 18)
	parent.AddChild(child)

	parent.SetBounds(parent.Bounds())
	parent.Layout()
	parent.layoutDirty = false

	child.SetBounds(rl.NewRectangle(0, 0, 200, 48))
	syncLayoutExtent(child)

	if !parent.IsDirty() {
		t.Fatal("parent should be dirty after child extent change")
	}
}
