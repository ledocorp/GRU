// Package ui — optional caret / idle diagnostics (enable from main via F5).
package ui

import (
	"fmt"
	"sync"
)

var caretDebugMu sync.Mutex

// CaretDebugEnabled gates stderr traces for caret color, space input, and idle blockers.
var CaretDebugEnabled bool

// CaretDebugLine writes one throttled line when CaretDebugEnabled is true.
func CaretDebugLine(format string, args ...any) {
	if !CaretDebugEnabled {
		return
	}
	caretDebugMu.Lock()
	defer caretDebugMu.Unlock()
	fmt.Printf("Gru caret-debug: "+format+"\n", args...)
}

// LogFocusedEditorCaret emits one line about the focused TextEditor (if any).
func LogFocusedEditorCaret(root Node) {
	if !CaretDebugEnabled || root == nil {
		return
	}
	ed := findFocusedTextEditor(root)
	if ed == nil {
		CaretDebugLine("no focused TextEditor")
		return
	}
	ink := ed.caretInkColor()
	CaretDebugLine(
		"caret id=%s phase=%d blink=%d shown=%t cursor=%d ink=%d,%d,%d themeCaret=%d,%d,%d dark=%t",
		ed.ID(), ed.caretPhase, ed.blinkPhase, ed.caretShown(), ed.cursor,
		ink.R, ink.G, ink.B, textEditorCaretColor.R, textEditorCaretColor.G, textEditorCaretColor.B,
		ThemeIsDark(),
	)
}

func findFocusedTextEditor(n Node) *TextEditor {
	if n == nil {
		return nil
	}
	if ed, ok := n.(*TextEditor); ok && ed.IsFocused() {
		return ed
	}
	for _, ch := range n.Children() {
		if ed := findFocusedTextEditor(ch); ed != nil {
			return ed
		}
	}
	return nil
}

func tailBytes(s string, n int) string {
	if n <= 0 || s == "" {
		return s
	}
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
