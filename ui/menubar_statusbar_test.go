package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestMenuBarClickOpensDropdown(t *testing.T) {
	item := []ContextMenuItem{{Label: "New"}}
	mb := NewMenuBar("mb", []MenuBarMenu{
		{Label: "File", Items: item},
		{Label: "Edit", Items: item},
	}, 0, 0, 400, 0)
	mb.SetBounds(rl.NewRectangle(0, 48, 400, menuBarDefaultH))
	mb.Layout()

	ptrClickPending = true
	ptrClickPos = rl.NewVector2(mb.menuRects[0].X+4, mb.menuRects[0].Y+4)
	ptrClickUsed = false
	scenePointerBlocked = false

	mb.Update(0)
	if !IsContextMenuOpen() {
		t.Fatal("click on File should open dropdown")
	}
	if mb.openMenuIdx != 0 {
		t.Fatalf("openMenuIdx = %d, want 0", mb.openMenuIdx)
	}
	CloseContextMenu()
}

func TestMenuBarClickWorksWhenHoverSuppressed(t *testing.T) {
	item := []ContextMenuItem{{Label: "Cut"}}
	mb := NewMenuBar("mb", []MenuBarMenu{
		{Label: "File", Items: item},
		{Label: "Edit", Items: item},
	}, 0, 0, 400, 0)
	mb.SetBounds(rl.NewRectangle(0, 48, 400, menuBarDefaultH))
	mb.Layout()

	mb.suppressHoverUntilLeave = true
	ptrClickPending = true
	ptrClickPos = rl.NewVector2(mb.menuRects[1].X+4, mb.menuRects[1].Y+4)
	ptrClickUsed = false
	scenePointerBlocked = false

	mb.Update(0)
	if !IsContextMenuOpen() {
		t.Fatal("click on Edit should open dropdown even when hover suppressed")
	}
	if mb.suppressHoverUntilLeave {
		t.Fatal("successful click should clear hover suppress")
	}
	CloseContextMenu()
}

func TestMenuBarLayoutRects(t *testing.T) {
	mb := NewMenuBar("mb", []MenuBarMenu{
		{Label: "File"},
		{Label: "Edit"},
	}, 0, 0, 400, 0)
	mb.SetBounds(rl.NewRectangle(0, 48, 400, menuBarDefaultH))
	mb.Layout()
	if len(mb.menuRects) != 2 {
		t.Fatalf("menuRects = %d, want 2", len(mb.menuRects))
	}
	if mb.menuRects[0].X < menuBarPadX-0.5 || mb.menuRects[0].X > menuBarPadX+0.5 {
		t.Fatalf("first menu x = %v", mb.menuRects[0].X)
	}
	if mb.menuRects[1].X <= mb.menuRects[0].X {
		t.Fatalf("menus should lay out left-to-right")
	}
}

func TestStatusBarSections(t *testing.T) {
	sb := NewStatusBar("sb", 0, 0, 600, 0)
	left := NewLabel("l", "Ready", 0, 0, 0, 0)
	center := NewLabel("c", "Mode", 0, 0, 0, 0)
	right := NewLabel("r", "Ln 1", 0, 0, 0, 0)
	sb.AddLeft(left)
	sb.AddCenter(center)
	sb.AddRight(right)
	sb.SetBounds(rl.NewRectangle(0, 500, 600, statusBarDefaultH))
	sb.Layout()

	if left.Bounds().X < statusBarPadX-0.5 {
		t.Fatalf("left x = %v", left.Bounds().X)
	}
	if right.Bounds().X+right.Bounds().Width > 600-statusBarPadX+0.5 {
		t.Fatalf("right overflow: %v", right.Bounds())
	}
	cx := center.Bounds().X + center.Bounds().Width/2
	if cx < 280 || cx > 320 {
		t.Fatalf("center not near middle: cx=%v", cx)
	}
}
