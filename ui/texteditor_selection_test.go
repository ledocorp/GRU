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
