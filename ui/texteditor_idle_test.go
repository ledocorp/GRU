package ui

import "testing"

// TestTextEditorIdleDrawDirtyClears verifies MarkDrawDirty + Draw clears drawDirty
// so SubtreeNeedsRedraw returns false after a cache frame (idle FPS prerequisite).
func TestTextEditorIdleDrawDirtyClears(t *testing.T) {
	ed := NewTextEditor("ed", "hello", 0, 0, 200, 40)
	ed.focused = true
	ed.markContentDirty()
	ed.flushContentDirty()
	if !ed.DbgDrawDirty() {
		t.Fatal("expected drawDirty after flushContentDirty")
	}
	// Simulate Draw() end without raylib (same defer as TextEditor.Draw).
	ed.drawDirty = false
	if ed.DbgDrawDirty() {
		t.Fatal("drawDirty should be clear after Draw for idle cache hit")
	}
}

func TestTextEditorCaretPhaseTransitionMarksDrawDirty(t *testing.T) {
	ed := NewTextEditor("ed", "", 0, 0, 100, 40)
	ed.focused = true
	ed.caretPhase = caretPhaseBlink
	ed.caretPhaseTimer = 0.001
	ed.updateCaretTimeline(0.01)
	if ed.caretPhase != caretPhaseFrozen {
		t.Fatalf("phase = %d, want frozen", ed.caretPhase)
	}
	if !ed.DbgDrawDirty() {
		t.Fatal("blink→frozen must MarkDrawDirty to bake caret into SSAA cache")
	}
}
