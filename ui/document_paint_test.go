package ui

import "testing"

func TestInvalidatePaintMarksDrawDirty(t *testing.T) {
	doc := NewDocument(640, 480)
	ClearDrawDirtySubtree(doc.Root)
	if doc.Root.DbgDrawDirty() {
		t.Fatal("expected clean draw flags before InvalidatePaint")
	}
	doc.InvalidatePaint()
	if !doc.Root.DbgDrawDirty() {
		t.Fatal("InvalidatePaint must mark root draw-dirty after scene swap")
	}
}
