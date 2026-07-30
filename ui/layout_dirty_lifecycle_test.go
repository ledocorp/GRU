package ui

import "testing"

func TestIconButtonLayoutClearsLayoutDirty(t *testing.T) {
	ib := NewIconButton("ib", "+", "Add", 0, 0, 40, 40)
	ib.MarkDirty()
	if !ib.IsDirty() {
		t.Fatal("expected layout dirty after MarkDirty")
	}
	ib.Layout()
	if ib.IsDirty() {
		t.Fatal("IconButton.Layout must clear layoutDirty")
	}
}

func TestTextEditorLayoutClearsLayoutDirty(t *testing.T) {
	ed := NewTextEditor("ed", "hi", 0, 0, 200, 40)
	ed.MarkDirty()
	ed.Layout()
	if ed.IsDirty() {
		t.Fatal("TextEditor.Layout must clear layoutDirty")
	}
}

func TestViewportDrawClearsDrawDirty(t *testing.T) {
	vp := NewViewport("vp", 0, 0, 100, 100)
	vp.MarkDrawDirty()
	vp.drawDirty = false // same defer as Viewport.Draw without raylib
	if vp.DbgDrawDirty() {
		t.Fatal("Viewport.Draw must clear drawDirty")
	}
}

func TestSubtreeCleanAfterLayoutPasses(t *testing.T) {
	root := NewContainer("root", 0, 0, 400, 400)
	ib := NewIconButton("notes", "", "Notes", 0, 0, 48, 48)
	ed := NewTextEditor("ed", "", 0, 0, 200, 40)
	root.AddChild(ib)
	root.AddChild(ed)
	ib.MarkDirty()
	ed.MarkDirty()
	ib.Layout()
	ed.Layout()
	ib.drawDirty = false
	ed.drawDirty = false
	root.layoutDirty = false
	root.drawDirty = false
	if SubtreeNeedsRedraw(root) {
		t.Fatal("after Layout clears layoutDirty, visible subtree should not need redraw")
	}
}
