package ui

import "testing"

func TestSubtreeNeedsRedrawSkipsHidden(t *testing.T) {
	root := NewContainer("root", 0, 0, 100, 100)
	hidden := NewContainer("hidden", 0, 0, 50, 50)
	visible := NewLabel("visible", "hi", 0, 0, 40, 20)
	root.AddChild(hidden)
	root.AddChild(visible)
	hidden.Hide()
	// Simulate layout+draw after visibility change; only the hidden leaf may stay dirty.
	root.layoutDirty = false
	root.drawDirty = false
	visible.layoutDirty = false
	visible.drawDirty = false
	hidden.MarkDrawDirty()
	if SubtreeNeedsRedraw(root) {
		t.Fatal("hidden drawDirty must not force document redraw")
	}
	visible.MarkDrawDirty()
	if !SubtreeNeedsRedraw(root) {
		t.Fatal("visible drawDirty must force document redraw")
	}
}

func TestMenuBarDrawClearsDrawDirty(t *testing.T) {
	mb := NewMenuBar("mb", []MenuBarMenu{{Label: "File", Items: nil}}, 0, 0, 200, 0)
	mb.MarkDrawDirty()
	mb.drawDirty = false // simulate Draw defer without raylib
	if mb.DbgDrawDirty() {
		t.Fatal("MenuBar.Draw must clear drawDirty")
	}
}
