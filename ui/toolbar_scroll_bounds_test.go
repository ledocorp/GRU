package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestApplyScrollBoundsDoesNotLayoutDirtyRoot(t *testing.T) {
	root := NewContainer("root", 0, 0, 480, 128)
	tb := NewToolbar("ribbon", 0, 0, 480, 128)
	tb.Ribbon = true
	tb.RibbonStacked = true
	tb.Overflow = true
	tb.OverflowKind = ToolbarOverflowScroll
	tb.SetBounds(rl.NewRectangle(0, 0, 480, 128))
	root.AddChild(tb)
	root.Layout()
	tb.Layout()

	tb.scrollActive = true
	tb.itemRects = []rl.Rectangle{
		rl.NewRectangle(30, 40, 72, 72),
		rl.NewRectangle(110, 40, 72, 72),
	}
	tb.scrollX = 12

	ClearDrawDirtySubtree(root)
	root.layoutDirty = false
	tb.layoutDirty = false

	tb.applyScrollBounds()
	if root.IsDirty() {
		t.Fatal("scroll pan should not propagate layout dirty to root")
	}
}
