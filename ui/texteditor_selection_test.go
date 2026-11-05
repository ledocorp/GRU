package ui

import "testing"

func TestTextEditorSelectedTextClampsOutOfRangeSelection(t *testing.T) {
	ed := NewTextEditor("sel-test", "", 0, 0, 200, 100)
	ed.selAnchor = 0
	ed.cursor = 260
	if got := ed.selectedText(); got != "" {
		t.Fatalf("selectedText() = %q, want empty for stale selection", got)
	}
	ed.Text.Set("hello")
	ed.selAnchor = 0
	ed.cursor = 10
	if got := ed.selectedText(); got != "hello" {
		t.Fatalf("selectedText() = %q, want %q", got, "hello")
	}
}

func TestTextEditorInsertReplacesSelection(t *testing.T) {
	ed := NewTextEditor("sel-replace", "abcdef", 0, 0, 200, 100)
	ed.Text.Set("abcdef")
	ed.selAnchor = 2
	ed.cursor = 5 // select "cde"
	ed.insertAtCursor("X")
	if got := ed.Text.Get(); got != "abXf" {
		t.Fatalf("after replace Text = %q, want %q", got, "abXf")
	}
	if ed.hasSelection() {
		t.Fatal("selection should be cleared after insert")
	}
	if ed.cursor != 3 {
		t.Fatalf("cursor = %d, want 3", ed.cursor)
	}
}
