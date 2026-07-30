package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Toolbar chrome must stretch with flex-column parents on resize (notepad ribbon).
func TestToolbarStretchesInFlexColumnOnResize(t *testing.T) {
	root := NewContainer("root", 0, 0, 800, 200)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.SetStyle("transparent")

	tb := NewToolbar("ribbon", 0, 0, 0, 128)
	tb.Ribbon = true
	tb.RibbonStacked = true
	tb.AddRibbonTab("Home")
	g := tb.AddSection(0, "file", "File")
	tb.AddButton(g.ID, "new", "New", nil)

	root.AddChild(tb)
	root.SetBounds(root.Bounds())
	root.Layout()

	if w := tb.Bounds().Width; w < 799 {
		t.Fatalf("initial toolbar width %.0f, want ~800", w)
	}

	root.SetBounds(rl.NewRectangle(0, 0, 1200, 200))
	root.MarkDirty()
	root.Layout()

	if w := tb.Bounds().Width; w < 1199 {
		t.Fatalf("after grow toolbar width %.0f, want ~1200", w)
	}
}

func TestFlexChildFillCrossWidthIncludesToolbar(t *testing.T) {
	tb := NewToolbar("tb", 0, 0, 400, 54)
	if !flexChildFillCrossWidth(tb, 900, 400) {
		t.Fatal("toolbar narrower than parent should fill cross-axis width")
	}
}
