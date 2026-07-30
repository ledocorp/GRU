// T1.6-R — automated smoke for Gru Notepad on the default shaped text path.
package ui

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func enableShapedNotepadText(t *testing.T) {
	t.Helper()
	if !InitShapedFonts() {
		t.Skip("no TTF faces for shaped engine")
	}
	initShapedScriptFaces()
	EnsureComplexScriptUIFont()
	SetTextEngineMode(TextEngineShaped)
}

func TestNotepadShapedEditorMeasureAndCaret(t *testing.T) {
	enableShapedNotepadText(t)
	defer SetTextEngineMode(TextEngineSDF)

	ed := NewTextEditor("np-ed", "Hello Notepad", 0, 0, 400, 120)
	st := GetThemeStyle("text-editor")
	st.Mono = true
	ed.SetStyleOverrides(st)
	ed.SetBounds(rl.NewRectangle(0, 0, 400, 120))
	ed.Layout()

	line := ed.Text.Get()
	style := ed.editorStyle(st)
	stops := shapedCaretStopsFromStyle(line, style)
	if len(stops) < 2 {
		t.Fatalf("caret stops = %v", stops)
	}
	fs := EffectiveFontSize(style)
	raw, ok := shapedMeasureTextFRaw(line, fs, styleDrawBold(style), style.Italic, style.Mono, style.PreviewFont)
	if !ok || raw <= 0 {
		t.Fatalf("shaped measure raw = %.1f ok=%v", raw, ok)
	}
}

func TestNotepadShapedArabicEditorMeasure(t *testing.T) {
	enableShapedNotepadText(t)
	defer SetTextEngineMode(TextEngineSDF)

	sample := ComplexScriptArabicSample
	ed := NewTextEditor("np-ar", sample, 0, 0, 480, 120)
	st := GetThemeStyle("text-editor")
	st.Mono = true
	ed.SetStyleOverrides(st)
	ed.SetBounds(rl.NewRectangle(0, 0, 480, 120))
	ed.Layout()

	style := ed.editorStyle(st)
	w := EditorMeasureWidth(sample, style)
	if w <= 0 {
		t.Fatal("Arabic editor measure width should be positive")
	}
	stops := shapedCaretStopsFromStyle(sample, style)
	if len(stops) < 2 {
		t.Fatalf("Arabic caret stops = %v", stops)
	}
	off, ok := shapedCaretOffsetAtX(sample, w*0.5, style)
	if !ok {
		t.Fatal("Arabic caret X hit-test failed")
	}
	if off < 0 || off > len(sample) {
		t.Fatalf("caret offset = %d out of range for len %d", off, len(sample))
	}
}

func TestNotepadShapedBackendName(t *testing.T) {
	enableShapedNotepadText(t)
	defer SetTextEngineMode(TextEngineSDF)

	name := TextEngineBackendName()
	if !strings.HasPrefix(name, "shaped") {
		t.Fatalf("backend = %q, want shaped* for notepad profile", name)
	}
}

func TestNotepadShapedIdleAfterEditorLayout(t *testing.T) {
	enableShapedNotepadText(t)
	defer SetTextEngineMode(TextEngineSDF)

	ed := NewTextEditor("np-ed", "Idle probe", 0, 0, 320, 80)
	ed.SetBounds(rl.NewRectangle(0, 0, 320, 80))
	vp := NewViewport("np-vp", 0, 0, 320, 200)
	vp.AddChild(ed)
	root := NewContainer("np-root", 0, 0, 320, 200)
	root.LayoutType = LayoutFlex
	root.FlexDirection = FlexColumn
	root.AddChild(vp)
	root.SetBounds(rl.NewRectangle(0, 0, 320, 200))

	root.Layout()
	SimulateCacheHitFrame(root)
	AssertIdleReady(t, root, "notepad editor + viewport")
}
