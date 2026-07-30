package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestToolbarAbsorbsVerticalWheelOnRibbonBand(t *testing.T) {
	tb := NewToolbar("ribbon", 0, 0, 480, 128)
	tb.Ribbon = true
	tb.RibbonStacked = true
	tb.Overflow = true
	tb.OverflowKind = ToolbarOverflowScroll
	tb.SetBounds(rl.NewRectangle(0, 0, 480, 128))
	tb.scrollActive = true
	tb.scrollLaneRect = rl.NewRectangle(22, 36, 436, 84)

	mouse := rl.Vector2{X: 240, Y: 72}
	if !tb.absorbsRibbonWheel(mouse) {
		t.Fatal("ribbon band should absorb wheel")
	}
	if !deepHasWheelConsumer([]Node{tb}, mouse, -1) {
		t.Fatal("deepHasWheelConsumer should find toolbar")
	}
}

func TestResolveWheelOwnerRibbonBlocksPage(t *testing.T) {
	resetWheelGesturesForTest()
	root := NewContainer("root", 0, 0, 480, 600)
	tb := NewToolbar("ribbon", 0, 0, 480, 128)
	tb.Ribbon = true
	tb.RibbonStacked = true
	tb.Overflow = true
	tb.OverflowKind = ToolbarOverflowScroll
	tb.SetBounds(rl.NewRectangle(0, 0, 480, 128))
	tb.scrollActive = true
	tb.scrollLaneRect = rl.NewRectangle(22, 36, 436, 84)
	page := NewViewport("page-vp", 0, 128, 480, 472)
	page.contentHeight = 2000
	root.AddChild(tb)
	root.AddChild(page)
	root.Layout()

	mouse := rl.Vector2{X: 240, Y: 72}
	owner := resolveWheelScrollOwner(page, root, mouse, -1)
	if owner != nil {
		t.Fatalf("wheel over ribbon should not assign page viewport, got %v", owner.ID())
	}
}
