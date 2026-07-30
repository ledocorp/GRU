package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestSplitViewLayoutDoesNotReDirtyRoot(t *testing.T) {
	left := NewLabel("left", "Left", 0, 0, 0, 24)
	right := NewLabel("right", "Right", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 400, 200)
	root := NewContainer("root", 0, 0, 400, 200)
	root.AddChild(sv)
	root.Layout()
	root.layoutDirty = false

	sv.MarkDirty()
	root.Layout()
	if root.IsDirty() {
		t.Fatal("root should stay clean after split relayout with unchanged pane bounds")
	}
	SimulateCacheHitFrame(root)
	if SubtreeNeedsRedraw(root) {
		t.Fatal("root subtree should not need redraw after idle split layout")
	}
}

func TestSplitViewLayoutRelayoutsWhenBoundsChange(t *testing.T) {
	left := NewLabel("left", "Left", 0, 0, 0, 24)
	right := NewLabel("right", "Right", 0, 0, 0, 24)
	sv := NewSplitView("sv", SplitHorizontal, left, right, 0, 0, 400, 200)
	sv.Layout()
	w0 := left.Bounds().Width
	sv.SetBounds(rl.NewRectangle(0, 0, 600, 200))
	sv.MarkDirty()
	sv.Layout()
	if left.Bounds().Width <= w0+10 {
		t.Fatalf("left width %.0f should grow after split widen (was %.0f)", left.Bounds().Width, w0)
	}
}

func TestSplitViewRelayoutsFlexPaneWithoutPriorDirty(t *testing.T) {
	notes := NewContainer("notes", 0, 0, 0, 0)
	notes.LayoutType = LayoutFlex
	notes.FlexDirection = FlexColumn
	notes.SetFlexGrow(1)
	hdr := NewLabel("hdr", "Open Notes", 0, 0, 0, 24)
	notes.AddChild(hdr)
	editor := NewContainer("editor", 0, 0, 0, 0)
	editor.SetFlexGrow(1)
	sv := NewSplitView("main", SplitHorizontal, notes, editor, 0, 0, 800, 400)
	sv.Layout()
	notes.layoutDirty = false
	editor.layoutDirty = false
	w0 := notes.Bounds().Width
	sv.Ratio.Set(0.35)
	sv.Layout()
	if notes.Bounds().Width == w0 {
		t.Fatalf("notes pane width should change after split ratio drag (stuck at %.0f)", w0)
	}
	if hdr.Bounds().Width < notes.Bounds().Width*0.9 {
		t.Fatalf("notes header should flex to pane width (hdr=%.0f pane=%.0f)", hdr.Bounds().Width, notes.Bounds().Width)
	}
}
