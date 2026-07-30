package ui

import "testing"

func TestSubtreeNeedsRedrawFindsChildDrawDirty(t *testing.T) {
	root := NewContainer("root", 0, 0, 400, 400)
	child := NewLabel("child", "ok", 0, 0, 100, 20)
	root.AddChild(child)
	root.layoutDirty = false
	root.drawDirty = false
	child.layoutDirty = false
	child.drawDirty = true

	if !SubtreeNeedsRedraw(root) {
		t.Fatal("child drawDirty must request a cache refresh")
	}
}

func TestShowMarksFlexSiblingsDirty(t *testing.T) {
	col := NewContainer("col", 0, 0, 400, 400)
	col.LayoutType = LayoutFlex
	col.FlexDirection = FlexColumn
	body := NewContainer("body", 0, 0, 0, 0)
	body.SetFlexGrow(1)
	status := NewStatusBar("status", 0, 0, 0, 0)
	col.AddChild(body)
	col.AddChild(status)
	status.Hide()
	col.layoutDirty = false
	body.layoutDirty = false
	status.layoutDirty = false

	status.Show()
	if !col.IsDirty() {
		t.Fatal("flex parent must reflow when status bar is shown")
	}
}
