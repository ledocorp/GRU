package ui

import "testing"

func TestWrapEditorLinesPreservesTrailingSpace(t *testing.T) {
	style := Style{FontSize: 17, MinFontSize: 17, Mono: true}
	lines := wrapEditorLines("hello ", 800, style)
	if len(lines) != 1 || lines[0] != "hello " {
		t.Fatalf("lines = %#v, want [hello ]", lines)
	}
}

func TestWrapEditorLinesPreservesDoubleSpace(t *testing.T) {
	style := Style{FontSize: 17, MinFontSize: 17, Mono: true}
	src := "hello  world"
	lines := wrapEditorLines(src, 800, style)
	if len(lines) != 1 || lines[0] != src {
		t.Fatalf("lines = %#v, want %q", lines, src)
	}
}

func TestBuildDisplayLinesStartsMatchBuffer(t *testing.T) {
	ed := NewTextEditor("ed", "hello world", 0, 0, 400, 40)
	ed.WordWrap = true
	style := ed.editorStyle(Style{FontSize: 17, MinFontSize: 17, Mono: true})
	ed.wrapCacheValid = false
	lines, starts := ed.buildDisplayLines("hello world", 800, style)
	if len(lines) == 0 || len(starts) != len(lines) {
		t.Fatalf("lines=%d starts=%d", len(lines), len(starts))
	}
	joined := ""
	for _, ln := range lines {
		joined += ln
	}
	if joined != "hello world" {
		t.Fatalf("joined = %q, want hello world", joined)
	}
	if starts[0] != 0 {
		t.Fatalf("starts[0] = %d, want 0", starts[0])
	}
	ed.cursor = 6 // after "hello "
	vLine := visualLineIndex(starts, ed.cursor)
	if vLine != 0 {
		t.Fatalf("vLine = %d, want 0 for cursor in first wrapped line", vLine)
	}
	prefix := "hello world"[starts[vLine]:ed.cursor]
	if prefix != "hello " {
		t.Fatalf("prefix = %q, want %q", prefix, "hello ")
	}
}

func TestTextEditorLayoutWordWrapWidthChangeMarksDrawOnly(t *testing.T) {
	ed := NewTextEditor("ed", "hi", 0, 0, 200, 40)
	ed.WordWrap = true
	ed.lastInnerW = 100
	ed.frameDirty = false
	ed.Layout()
	if ed.frameDirty {
		t.Fatal("word-wrap width reflow must not set frameDirty (Layout/flush loop)")
	}
}
