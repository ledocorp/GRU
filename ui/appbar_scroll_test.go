package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestViewportRepositionOnlyShiftsAppBarSlots(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 400, 300)
	vp.SetStyle("page-scroll")

	bar := NewAppBar("bar", "Settings", 0, 0, 0, 0)
	menu := NewIconButton("menu", "", "Menu", 0, 0, 40, 40)
	save := NewButton("save", "Save", 0, 0, 72, 36)
	bar.SetLeading(menu)
	bar.AddTrailing(save)

	vp.AddChild(bar)
	vp.SetBounds(rl.NewRectangle(0, 0, 400, 300))
	vp.Layout()

	menuBefore := menu.Bounds()
	saveBefore := save.Bounds()

	vp.ScrollY = 60
	vp.scrollDirty = true
	vp.MarkDirty()
	vp.Layout()

	menuAfter := menu.Bounds()
	saveAfter := save.Bounds()
	wantMenuY := menuBefore.Y - 60
	wantSaveY := saveBefore.Y - 60
	if absF(menuAfter.Y-wantMenuY) > 0.5 {
		t.Fatalf("menu Y after scroll = %.1f, want %.1f", menuAfter.Y, wantMenuY)
	}
	if absF(saveAfter.Y-wantSaveY) > 0.5 {
		t.Fatalf("save Y after scroll = %.1f, want %.1f", saveAfter.Y, wantSaveY)
	}
}
