package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestWidgetBlockedByOverlayBlocksForeignComboFace(t *testing.T) {
	top := NewComboBox("top", []string{"A", "B"}, NewSignal("A"), 10, 10, 120, 40)
	top.isOpen = true
	overlayHitRects = []rl.Rectangle{top.PopupBounds(), top.Bounds()}

	mouse := rl.NewVector2(70, 80) // inside top popup and over where bottom face would be
	if !rl.CheckCollisionPointRec(mouse, top.PopupBounds()) {
		t.Fatal("test setup: mouse should be inside top popup")
	}
	if WidgetBlockedByOverlay(mouse, top.OverlayExemptRects()...) {
		t.Fatal("open combobox should stay interactive over its own popup")
	}
	if !WidgetBlockedByOverlay(mouse) {
		t.Fatal("foreign widget at same point should be blocked by open combobox popup")
	}
}

func TestViewportOpenPopupExtendsScrollRange(t *testing.T) {
	vp := NewViewport("vp", 100, 50, 300, 200)
	vp.styleName = "default"
	vp.contentHeight = 500
	vp.lastFlexValid = true
	vp.lastFlexW = 300
	vp.lastFlexH = 200
	vp.ScrollY = 300 // max scroll without an open popup

	cb := NewComboBox("cb", []string{"One", "Two", "Three"}, NewSignal("One"), 110, 110, 160, 40)
	vp.AddChild(cb)
	cb.SetParent(vp)
	cb.isOpen = true

	baseMax := float32(500 - 200)
	if vp.ScrollY != baseMax {
		t.Fatalf("setup ScrollY = %v, want %v", vp.ScrollY, baseMax)
	}
	ext := vp.openMenuPopupScrollExtension()
	if ext <= 0 {
		t.Fatal("open popup should extend scroll range below the client")
	}
	if vp.overflowScrollY() <= baseMax {
		t.Fatalf("overflowScrollY = %v, want > %v while popup hangs below client", vp.overflowScrollY(), baseMax)
	}
}

func TestScrollViewportRevealsPopupBelowClient(t *testing.T) {
	vp := NewViewport("vp", 100, 50, 300, 200)
	vp.styleName = "default"
	vp.contentHeight = 600
	vp.lastFlexValid = true
	vp.lastFlexW = 300
	vp.lastFlexH = 200

	row := NewContainer("row", 0, 0, 280, 40)
	vp.AddChild(row)
	cb := NewComboBox("cb", []string{"One", "Two", "Three", "Four"}, NewSignal("One"), 0, 0, 160, 40)
	row.AddChild(cb)

	row.SetBounds(rl.NewRectangle(110, 210, 280, 40))
	cb.SetBounds(rl.NewRectangle(110, 210, 160, 40))
	cb.SetParent(row)
	row.SetParent(vp)

	cb.isOpen = true
	pop := cb.PopupBounds()
	client := vp.viewportPaddedClientRect()
	if pop.Y+pop.Height <= client.Y+client.Height {
		t.Fatalf("test setup: popup %v should extend below client %v", pop, client)
	}

	before := vp.ScrollY
	scrollAncestorViewportsToRevealRect(cb, pop)
	if vp.ScrollY <= before {
		t.Fatalf("ScrollY = %v, want increase from %v to reveal popup bottom", vp.ScrollY, before)
	}
	revealed := cb.PopupBounds()
	if revealed.Y+revealed.Height > client.Y+client.Height+1 {
		t.Fatalf("popup still clipped after scroll: %v vs client bottom %v", revealed, client.Y+client.Height)
	}
}
