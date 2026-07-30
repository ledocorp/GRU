package ui

import "testing"

func TestTextEditorInsertSpaceAtCursor(t *testing.T) {
	ed := NewTextEditor("ed", "hello", 0, 0, 200, 40)
	ed.cursor = 5
	ed.focused = true
	ed.insertAtCursor(" ")
	if got := ed.Text.Get(); got != "hello " {
		t.Fatalf("text = %q, want %q", got, "hello ")
	}
	if ed.cursor != 6 {
		t.Fatalf("cursor = %d, want 6", ed.cursor)
	}
}

func TestTextEditorCaretPrefixMatchesMeasureTextS(t *testing.T) {
	SetAppearance(AppearanceLight)
	ed := NewTextEditor("ed", "abc def", 0, 0, 400, 40)
	ed.SetStyleOverrides(Style{FontSize: 17, MinFontSize: 17, Mono: true})
	ed.focused = true
	ed.cursor = 4 // after "abc "
	style := ed.editorStyle(ed.GetStyle())
	lines, starts := ed.displayLines(style)
	vLine := visualLineIndex(starts, ed.cursor)
	col := ed.cursor - starts[vLine]
	prefix := lines[vLine][:col]
	w := measureTextS(prefix, style)
	if w <= 0 && !sdfReady && !fontReady {
		t.Skip("SDF fonts not loaded in unit test")
	}
	if w <= 0 {
		t.Fatalf("prefix %q width = %d, want > 0", prefix, w)
	}
	// Mid-buffer space: prefix width must include the space character.
	if col != 4 || prefix != "abc " {
		t.Fatalf("setup col=%d prefix=%q", col, prefix)
	}
}

func TestSetAppearanceUpdatesEditorCaretColor(t *testing.T) {
	prev := currentAppearanceMode
	t.Cleanup(func() { currentAppearanceMode = prev })

	SetAppearance(AppearanceDark)
	dark := textEditorCaretColor
	SetAppearance(AppearanceLight)
	light := textEditorCaretColor
	if dark.R == light.R && dark.G == light.G && dark.B == light.B {
		t.Fatalf("caret color unchanged across themes: dark=%v light=%v", dark, light)
	}
	if dark.R < 240 {
		t.Fatalf("dark caret = %v, want near-white", dark)
	}
}
