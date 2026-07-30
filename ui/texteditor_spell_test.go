package ui

import "testing"

func TestTextEditorSpellMisspelling(t *testing.T) {
	ed := NewTextEditor("ed", "Hello teh world", 0, 0, 200, 80)
	ed.SetSpellChecker(NewSimpleSpellChecker())
	ed.SetSpellCheckEnabled(true)
	ed.FlushSpellCheck()
	if len(ed.spellMiss) != 1 {
		t.Fatalf("spellMiss = %v, want one (teh)", ed.spellMiss)
	}
}

func TestTextEditorAutoCorrect(t *testing.T) {
	ed := NewTextEditor("ed", "teh", 0, 0, 200, 80)
	ed.cursor = 3
	ed.SetSpellAutoCorrect(true)
	ed.SetSpellAutoCorrectTable(map[string]string{"teh": "the"})
	if !ed.trySpellAutoCorrectBeforeSpace() {
		t.Fatal("expected auto-correct")
	}
	if ed.Text.Get() != "the" {
		t.Fatalf("text = %q, want the", ed.Text.Get())
	}
}
