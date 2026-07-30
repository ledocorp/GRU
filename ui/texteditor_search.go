package ui

import (
	"strings"
)

// SelectRange selects the half-open byte range [lo, hi) in the buffer and scrolls the caret into view.
func (ed *TextEditor) SelectRange(lo, hi int) {
	text := ed.Text.Get()
	n := len(text)
	if lo < 0 {
		lo = 0
	}
	if hi < 0 {
		hi = 0
	}
	if lo > n {
		lo = n
	}
	if hi > n {
		hi = n
	}
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo == hi {
		ed.cursor = lo
		ed.clearSelection()
	} else {
		ed.selAnchor = lo
		ed.cursor = hi
	}
	ed.ensureCursorVisible()
	ed.markContentDirty()
	ed.flushContentDirty()
	ed.RestartCaretCycle()
}

// FindForward locates needle in text starting at start (inclusive). When wrap is true, continues from index 0.
func (ed *TextEditor) FindForward(needle string, start int, matchCase, wrap bool) (lo, hi int, ok bool) {
	if needle == "" {
		return 0, 0, false
	}
	text := ed.Text.Get()
	if start < 0 {
		start = 0
	}
	if start > len(text) {
		start = len(text)
	}
	if i := indexFrom(text, needle, start, matchCase); i >= 0 {
		return i, i + len(needle), true
	}
	if !wrap || start == 0 {
		return 0, 0, false
	}
	if i := indexFrom(text, needle, 0, matchCase); i >= 0 && i < start {
		return i, i + len(needle), true
	}
	return 0, 0, false
}

// FindBackward locates the last needle occurrence strictly before before (exclusive). When wrap is true, continues from the end.
func (ed *TextEditor) FindBackward(needle string, before int, matchCase, wrap bool) (lo, hi int, ok bool) {
	if needle == "" {
		return 0, 0, false
	}
	text := ed.Text.Get()
	if before < 0 {
		before = 0
	}
	if before > len(text) {
		before = len(text)
	}
	if i := lastIndexBefore(text, needle, before, matchCase); i >= 0 {
		return i, i + len(needle), true
	}
	if !wrap || before >= len(text) {
		return 0, 0, false
	}
	if i := lastIndexBefore(text, needle, len(text), matchCase); i >= 0 && i >= before {
		return i, i + len(needle), true
	}
	return 0, 0, false
}

// FindNext selects the next match after the caret (or after the current selection). Returns false when not found.
func (ed *TextEditor) FindNext(needle string, matchCase bool) bool {
	start := ed.cursor
	if ed.hasSelection() {
		_, hi := ed.selectionRange()
		start = hi
	}
	lo, hi, ok := ed.FindForward(needle, start, matchCase, true)
	if !ok {
		return false
	}
	ed.SelectRange(lo, hi)
	return true
}

// FindPrevious selects the previous match before the caret (or before the current selection). Returns false when not found.
func (ed *TextEditor) FindPrevious(needle string, matchCase bool) bool {
	before := ed.cursor
	if ed.hasSelection() {
		lo, _ := ed.selectionRange()
		before = lo
	}
	lo, hi, ok := ed.FindBackward(needle, before, matchCase, true)
	if !ok {
		return false
	}
	ed.SelectRange(lo, hi)
	return true
}

// ReplaceSelection replaces the current selection with repl and records undo. Returns false when nothing is selected.
func (ed *TextEditor) ReplaceSelection(repl string) bool {
	if !ed.hasSelection() {
		return false
	}
	ed.pushUndo()
	lo, hi := ed.selectionRange()
	text := ed.Text.Get()
	ed.setText(text[:lo] + repl + text[hi:])
	ed.cursor = lo + len(repl)
	ed.clearSelection()
	ed.notifyChange()
	ed.markContentDirty()
	ed.flushContentDirty()
	ed.RestartCaretCycle()
	return true
}

// ReplaceOne replaces the current selection when it matches needle, otherwise finds the next match and replaces it.
func (ed *TextEditor) ReplaceOne(needle, repl string, matchCase bool) bool {
	if needle == "" {
		return false
	}
	if ed.hasSelection() && textMatches(ed.selectedText(), needle, matchCase) {
		return ed.ReplaceSelection(repl)
	}
	if !ed.FindNext(needle, matchCase) {
		return false
	}
	return ed.ReplaceSelection(repl)
}

// ReplaceAll replaces every occurrence of needle with repl. Returns the number of replacements.
func (ed *TextEditor) ReplaceAll(needle, repl string, matchCase bool) int {
	if needle == "" {
		return 0
	}
	text := ed.Text.Get()
	var b strings.Builder
	b.Grow(len(text))
	count := 0
	at := 0
	for {
		i := indexFrom(text, needle, at, matchCase)
		if i < 0 {
			b.WriteString(text[at:])
			break
		}
		b.WriteString(text[at:i])
		b.WriteString(repl)
		count++
		at = i + len(needle)
	}
	if count == 0 {
		return 0
	}
	ed.pushUndo()
	ed.setText(b.String())
	ed.cursor = 0
	ed.clearSelection()
	ed.notifyChange()
	ed.markContentDirty()
	ed.flushContentDirty()
	ed.RestartCaretCycle()
	return count
}

// GoToLine moves the caret to the start of the 1-based line number. Returns false when line is out of range.
func (ed *TextEditor) GoToLine(line int) bool {
	if line < 1 {
		line = 1
	}
	text := ed.Text.Get()
	maxLine := 1
	if len(text) > 0 {
		maxLine = strings.Count(text, "\n") + 1
	}
	if line > maxLine {
		return false
	}
	off := lineColToOffset(text, line-1, 0)
	ed.clearSelection()
	ed.cursor = off
	ed.ensureCursorVisible()
	ed.markContentDirty()
	ed.flushContentDirty()
	ed.RestartCaretCycle()
	return true
}

func indexFrom(s, sub string, from int, matchCase bool) int {
	if sub == "" || from > len(s) {
		return -1
	}
	if from < 0 {
		from = 0
	}
	rest := s[from:]
	if matchCase {
		if i := strings.Index(rest, sub); i >= 0 {
			return from + i
		}
		return -1
	}
	for i := 0; i+len(sub) <= len(rest); i++ {
		if strings.EqualFold(rest[i:i+len(sub)], sub) {
			return from + i
		}
	}
	return -1
}

func lastIndexBefore(s, sub string, before int, matchCase bool) int {
	if sub == "" || before <= 0 {
		return -1
	}
	if before > len(s) {
		before = len(s)
	}
	head := s[:before]
	if matchCase {
		return strings.LastIndex(head, sub)
	}
	last := -1
	for i := 0; i+len(sub) <= len(head); i++ {
		if strings.EqualFold(head[i:i+len(sub)], sub) {
			last = i
		}
	}
	return last
}

func textMatches(haystack, needle string, matchCase bool) bool {
	if matchCase {
		return haystack == needle
	}
	return strings.EqualFold(haystack, needle)
}
