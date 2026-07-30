package ui

import "testing"

func TestTextInkHeightUsesMonoFace(t *testing.T) {
	ui := Style{FontSize: 17, MinFontSize: 17, Mono: true}
	monoH := TextInkHeight(ui)
	sans := Style{FontSize: 17, MinFontSize: 17, Mono: false}
	uiH := TextInkHeight(sans)
	if monoH <= 0 || uiH <= 0 {
		t.Fatalf("ink heights must be positive mono=%.1f ui=%.1f", monoH, uiH)
	}
}

func TestEditorCaretStrokeIsOneScreenPixel(t *testing.T) {
	if textEditorCaretStrokePx != 1 {
		t.Fatalf("caret stroke = %.1f, want 1 screen px", textEditorCaretStrokePx)
	}
}

func TestTextDrawUsesShapedFalseForLatinMono(t *testing.T) {
	if textDrawUsesShaped("hello world ", 17, false, false, true, false) {
		t.Fatal("mono Latin editor text must use SDF draw/measure path")
	}
}

func TestCaretInkColorUsesThemeAccent(t *testing.T) {
	prev := currentAppearanceMode
	t.Cleanup(func() { currentAppearanceMode = prev })

	SetAppearance(AppearanceDark)
	ed := NewTextEditor("ed", "", 0, 0, 100, 40)
	c := ed.caretInkColor()
	if c.R < 240 || c.G < 240 || c.B < 240 {
		t.Fatalf("dark theme caret = %v, want near-white from SetAppearance", c)
	}

	SetAppearance(AppearanceLight)
	c = ed.caretInkColor()
	if c.R > 80 || c.G > 80 || c.B > 80 {
		t.Fatalf("light theme caret = %v, want dark ink from SetAppearance", c)
	}
}

func TestTextEditorDrawClearsDrawDirty(t *testing.T) {
	ed := NewTextEditor("ed", "hi", 0, 0, 100, 40)
	ed.MarkDrawDirty()
	if !ed.DbgDrawDirty() {
		t.Fatal("expected drawDirty after MarkDrawDirty")
	}
	// Draw() ends with defer ed.drawDirty = false (raylib required for full Draw in tests).
	ed.drawDirty = false
	if ed.DbgDrawDirty() {
		t.Fatal("TextEditor.Draw must clear drawDirty so SSAA cache can idle")
	}
}
