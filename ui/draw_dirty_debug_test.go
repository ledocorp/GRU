package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestShowHideIdempotent(t *testing.T) {
	el := NewElement("x", 0, 0, 10, 10)
	el.layoutDirty = false
	el.drawDirty = false
	el.Show()
	if el.IsDirty() || el.DbgDrawDirty() {
		t.Fatal("Show when already visible must not mark dirty")
	}
	el.Hide()
	if !el.IsDirty() {
		t.Fatal("Hide should mark dirty once")
	}
	el.layoutDirty = false
	el.drawDirty = false
	el.Hide()
	if el.IsDirty() || el.DbgDrawDirty() {
		t.Fatal("Hide when already hidden must not mark dirty")
	}
}

func TestSetBoundsIdempotent(t *testing.T) {
	el := NewElement("x", 0, 0, 10, 10)
	el.layoutDirty = false
	el.drawDirty = false
	el.SetBounds(rl.Rectangle{X: 0, Y: 0, Width: 10, Height: 10})
	if el.IsDirty() || el.DbgDrawDirty() {
		t.Fatal("SetBounds with same rect must not mark dirty")
	}
}

func TestCollectDirtyReportsFindsDrawDirty(t *testing.T) {
	root := NewContainer("root", 0, 0, 100, 100)
	child := NewLabel("child", "hi", 0, 0, 0, 0)
	root.AddChild(child)
	child.MarkDrawDirty()
	reports := CollectDirtyReports(root, 4)
	if len(reports) == 0 {
		t.Fatal("expected dirty report")
	}
	found := false
	for _, r := range reports {
		if r.ID == "child" && r.DrawDirty {
			found = true
		}
	}
	if !found {
		t.Fatalf("child draw dirty not reported: %+v", reports)
	}
}
