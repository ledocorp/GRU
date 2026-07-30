package ui

import (
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func resetWheelGesturesForTest() {
	pageWheelGestureUntil = time.Time{}
	nestedWheelGestureUntil = time.Time{}
	wheelGestureStickyNested = nil
}

func TestScrollGestureActiveHoldWindow(t *testing.T) {
	resetWheelGesturesForTest()
	if ScrollGestureActive() {
		t.Fatal("expected inactive before gesture")
	}
	notePageWheelGesture()
	if !ScrollGestureActive() {
		t.Fatal("expected active after page wheel gesture")
	}
}

func TestResolveWheelOwnerPageGestureBlocksNested(t *testing.T) {
	resetWheelGesturesForTest()
	page := NewViewport("page-vp", 0, 0, 800, 600)
	panel := NewPanel("panel", "Panel", 10, 10, 400, 300)
	inner := NewViewport("inner-vp", 20, 80, 360, 200)
	inner.contentHeight = 900
	rt := NewRichText("rt", []TextSpan{{Text: "x"}}, 0, 0, 0, 0)
	inner.AddChild(rt)
	panel.AddChild(inner)
	page.AddChild(panel)

	notePageWheelGesture()
	mouse := rl.Vector2{X: 200, Y: 150}
	owner := resolveWheelScrollOwner(page, page, mouse, -1)
	if owner != page {
		t.Fatalf("page gesture should keep page as owner, got %v", owner.ID())
	}
}

func TestResolveWheelOwnerPicksNestedOnHover(t *testing.T) {
	resetWheelGesturesForTest()
	page := NewViewport("page-vp", 0, 0, 800, 600)
	panel := NewPanel("panel", "Panel", 10, 10, 400, 300)
	inner := NewViewport("inner-vp", 20, 80, 360, 200)
	inner.contentHeight = 900
	panel.AddChild(inner)
	page.AddChild(panel)
	page.Layout()
	inner.contentHeight = 900

	b := inner.Bounds()
	mouse := rl.Vector2{X: b.X + b.Width/2, Y: b.Y + b.Height/2}
	owner := resolveWheelScrollOwner(page, page, mouse, -1)
	if owner != inner {
		t.Fatalf("hover over nested scroll region should select inner, got %v", owner.ID())
	}
}

func TestResolveWheelOwnerStickyNestedBlocksPage(t *testing.T) {
	resetWheelGesturesForTest()
	page := NewViewport("page-vp", 0, 0, 800, 600)
	inner := NewViewport("inner-vp", 20, 80, 360, 200)
	inner.contentHeight = 900
	page.AddChild(inner)
	page.Layout()

	wheelGestureStickyNested = inner
	noteNestedWheelGesture()
	owner := resolveWheelScrollOwner(page, page, rl.Vector2{X: 400, Y: 10}, -1)
	if owner != inner {
		t.Fatalf("active nested gesture should stay on nested, got %v", owner.ID())
	}
}

func TestResolveWheelOwnerSplitSiblingPane(t *testing.T) {
	resetWheelGesturesForTest()
	root := NewContainer("root", 0, 0, 800, 600)
	editor := NewViewport("editor-vp", 0, 0, 400, 600)
	editor.contentHeight = 1200
	preview := NewViewport("preview-vp", 400, 0, 400, 600)
	preview.contentHeight = 1200
	root.AddChild(editor)
	root.AddChild(preview)
	root.Layout()
	editor.SetBounds(rl.NewRectangle(0, 0, 400, 600))
	preview.SetBounds(rl.NewRectangle(400, 0, 400, 600))
	editor.contentHeight = 1200
	preview.contentHeight = 1200

	notePageWheelGesture()
	b := preview.Bounds()
	mouse := rl.Vector2{X: b.X + b.Width/2, Y: b.Y + b.Height/2}
	owner := resolveWheelScrollOwner(editor, root, mouse, -1)
	if owner != preview {
		t.Fatalf("cursor over sibling preview pane should scroll preview, got %v", owner.ID())
	}
}
