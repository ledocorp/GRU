package ui

import "testing"

func TestTextEditorFindNextWrap(t *testing.T) {
	ed := NewTextEditor("t", "aba", 0, 0, 100, 100)
	ed.cursor = len("aba")
	if !ed.FindNext("a", true) {
		t.Fatal("expected wrap find")
	}
	lo, hi := ed.selectionRange()
	if lo != 0 || hi != 1 {
		t.Fatalf("got range %d:%d", lo, hi)
	}
}

func TestTextEditorFindPrevious(t *testing.T) {
	ed := NewTextEditor("t", "aba", 0, 0, 100, 100)
	ed.cursor = 1
	if !ed.FindPrevious("a", true) {
		t.Fatal("expected find previous")
	}
	lo, hi := ed.selectionRange()
	if lo != 0 || hi != 1 {
		t.Fatalf("got range %d:%d", lo, hi)
	}
}

func TestTextEditorReplaceAll(t *testing.T) {
	ed := NewTextEditor("t", "foo foo", 0, 0, 100, 100)
	n := ed.ReplaceAll("foo", "bar", true)
	if n != 2 || ed.Text.Get() != "bar bar" {
		t.Fatalf("replace all: n=%d text=%q", n, ed.Text.Get())
	}
}

func TestTextEditorGoToLine(t *testing.T) {
	ed := NewTextEditor("t", "a\nb\nc", 0, 0, 100, 100)
	if !ed.GoToLine(2) {
		t.Fatal("GoToLine failed")
	}
	line, col := ed.CursorLineCol()
	if line != 2 || col != 1 {
		t.Fatalf("cursor at %d:%d", line, col)
	}
	if ed.GoToLine(99) {
		t.Fatal("expected out of range")
	}
}
