package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestListTileLayoutSlotsPositionsTrailingToggle(t *testing.T) {
	lt := NewListTile("t", "Wi-Fi", "Home", 100, 200, 400, 56)
	tog := NewToggle("tog", true, 0, 0, 52, 28)
	lt.SetTrailing(tog)

	lt.layoutSlots()

	b := tog.Bounds()
	if b.Width != 47 || b.Height != 25 {
		t.Fatalf("toggle bounds = %v, want 47x25", b)
	}
	wantX := float32(100) + 400 - listTilePadX - 47
	wantY := float32(200) + (56-25)/2.0
	if b.X != wantX || b.Y != wantY {
		t.Fatalf("toggle at (%v,%v), want (%v,%v)", b.X, b.Y, wantX, wantY)
	}
	center := rl.NewVector2(wantX+23.5, wantY+12.5)
	if !rl.CheckCollisionPointRec(center, tog.hitBounds()) {
		t.Fatal("center of toggle slot should hit")
	}
	tog.hostedInListTile = true
	hit := tog.hitBounds()
	if hit.Y < b.Y || hit.Y+hit.Height > b.Y+b.Height {
		t.Fatalf("hosted hit %v bleeds outside slot %v", hit, b)
	}
}

func TestListTileSwitchOnlyChildrenAndInspectorPick(t *testing.T) {
	lt := NewListTile("wifi", "Wi-Fi", "Net", 0, 0, 300, 56)
	tog := NewToggle("tog", false, 0, 0, 52, 28)
	lt.SetTrailing(tog)
	if !lt.SwitchOnly() {
		t.Fatal("SetTrailing(Toggle) should set switch-only mode")
	}
	if len(lt.Children()) != 1 || lt.Children()[0] != tog {
		t.Fatalf("Children = %v, want trailing toggle only", lt.Children())
	}
	if pick := lt.InspectorPickTarget(nil); pick != nil {
		t.Fatalf("InspectorPickTarget(nil) = %v, want nil for locked body", pick)
	}
	if pick := lt.InspectorPickTarget(tog); pick != tog {
		t.Fatalf("InspectorPickTarget(toggle) = %v, want toggle", pick)
	}
	lt.SetRowMode(ListTileNavigation)
	if lt.SwitchOnly() {
		t.Fatal("navigation mode should not be switch-only")
	}
	if pick := lt.InspectorPickTarget(nil); pick != lt {
		t.Fatal("navigation row without child hit should pick tile")
	}
}

func TestListTileSlotParentForDrawDirty(t *testing.T) {
	lt := NewListTile("t", "Wi-Fi", "", 0, 0, 300, 56)
	panel := NewContainer("panel", 0, 0, 400, 200)
	panel.AddChild(lt)
	tog := NewToggle("tog", false, 0, 0, 52, 28)
	lt.SetTrailing(tog)
	lt.layoutSlots()
	if tog.ParentNode() != lt {
		t.Fatalf("toggle parent = %v, want list tile", tog.ParentNode())
	}
	tog.MarkDrawDirty()
	if !panel.DbgDrawDirty() {
		t.Fatal("toggle MarkDrawDirty should bubble to panel via list tile")
	}
}

func TestListTileSwitchOnlyClearsOnClick(t *testing.T) {
	lt := NewListTile("t", "Wi-Fi", "", 0, 0, 200, 56)
	lt.OnClick = func() {}
	lt.SetTrailing(NewToggle("tog", false, 0, 0, 52, 28))
	if lt.OnClick != nil {
		t.Fatal("switch-only row must clear OnClick")
	}
}

func TestListTileAvailableTextWidthShrinksWithTrailing(t *testing.T) {
	lt := NewListTile("t", "Title", "Subtitle", 0, 0, 280, 56)
	lt.Layout()
	b := lt.Bounds()
	textX := b.X + listTilePadX
	maxOpen := b.X + b.Width - listTilePadX - textX

	lt.SetTrailing(NewToggle("tog", false, 0, 0, 52, 28))
	lt.layoutSlots()
	tw := lt.Trailing.Bounds()
	textRight := tw.X - 8
	maxWithToggle := textRight - textX

	if maxWithToggle >= maxOpen {
		t.Fatalf("text band with toggle %v should be narrower than without %v", maxWithToggle, maxOpen)
	}
	if maxWithToggle < 20 {
		t.Fatalf("maxWithToggle = %v, want at least 20", maxWithToggle)
	}
}
