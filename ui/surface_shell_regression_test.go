package ui

import "testing"

func TestSurfaceShellRemoveChild(t *testing.T) {
	p := NewPanel("p", "Panel", 0, 0, 200, 100)
	lbl := NewLabel("gone", "bye", 0, 0, 0, 20)
	p.AddChild(lbl)
	if len(p.Children()) != 1 {
		t.Fatalf("children = %d, want 1", len(p.Children()))
	}
	p.RemoveChild("gone")
	if len(p.Children()) != 0 {
		t.Fatalf("children = %d after remove, want 0", len(p.Children()))
	}
}

func TestSurfaceShellInvalidateLayoutPassCacheClearsBody(t *testing.T) {
	p := NewPanel("p", "Panel", 0, 0, 200, 100)
	p.AddChild(NewLabel("l", "x", 0, 0, 0, 20))
	p.Layout()
	p.body.lastLayoutPassValid = true
	p.InvalidateLayoutPassCache()
	if p.body.lastLayoutPassValid {
		t.Fatal("body layout pass cache should invalidate with shell")
	}
}

func TestMarkResizeLayoutDirtySubtreeIncludesSurfaceBody(t *testing.T) {
	p := NewPanel("p", "Panel", 0, 0, 200, 100)
	p.AddChild(NewLabel("l", "x", 0, 0, 0, 20))
	p.Layout()
	p.body.lastLayoutPassValid = true
	p.body.layoutDirty = false
	MarkResizeLayoutDirtySubtree(p)
	if !p.body.layoutDirty {
		t.Fatal("resize dirty walk should mark internal body dirty")
	}
}
