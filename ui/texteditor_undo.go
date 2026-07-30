package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const editorUndoLimit = 128

type editorSnapshot struct {
	text      string
	cursor    int
	selAnchor int
}

func (ed *TextEditor) captureSnapshot() editorSnapshot {
	return editorSnapshot{
		text:      ed.Text.Get(),
		cursor:    ed.cursor,
		selAnchor: ed.selAnchor,
	}
}

func (ed *TextEditor) pushUndo() {
	if ed.undoSuspend {
		return
	}
	ed.undoStack = append(ed.undoStack, ed.captureSnapshot())
	if len(ed.undoStack) > editorUndoLimit {
		ed.undoStack = ed.undoStack[1:]
	}
	ed.redoStack = ed.redoStack[:0]
}

func (ed *TextEditor) restoreSnapshot(snap editorSnapshot) {
	ed.undoSuspend = true
	ed.Text.Set(snap.text)
	ed.cursor = snap.cursor
	ed.selAnchor = snap.selAnchor
	ed.clampScrollY()
	ed.ensureCursorVisible()
	ed.notifyChange()
	ed.markContentDirty()
	ed.flushContentDirty()
	ed.undoSuspend = false
}

func (ed *TextEditor) clearUndoHistory() {
	ed.undoStack = ed.undoStack[:0]
	ed.redoStack = ed.redoStack[:0]
}

// CanUndo reports whether Undo would change the buffer.
func (ed *TextEditor) CanUndo() bool { return len(ed.undoStack) > 0 }

// CanRedo reports whether Redo would change the buffer.
func (ed *TextEditor) CanRedo() bool { return len(ed.redoStack) > 0 }

// HasSelection reports whether a non-empty range is selected.
func (ed *TextEditor) HasSelection() bool { return ed.hasSelection() }

// Undo reverts the last edit.
func (ed *TextEditor) Undo() bool {
	if len(ed.undoStack) == 0 {
		return false
	}
	ed.redoStack = append(ed.redoStack, ed.captureSnapshot())
	snap := ed.undoStack[len(ed.undoStack)-1]
	ed.undoStack = ed.undoStack[:len(ed.undoStack)-1]
	ed.restoreSnapshot(snap)
	ed.RestartCaretCycle()
	return true
}

// Redo reapplies the last undone edit.
func (ed *TextEditor) Redo() bool {
	if len(ed.redoStack) == 0 {
		return false
	}
	ed.undoStack = append(ed.undoStack, ed.captureSnapshot())
	snap := ed.redoStack[len(ed.redoStack)-1]
	ed.redoStack = ed.redoStack[:len(ed.redoStack)-1]
	ed.restoreSnapshot(snap)
	ed.RestartCaretCycle()
	return true
}

// Copy copies the selection, or the whole document when nothing is selected.
func (ed *TextEditor) Copy() {
	if ed.Disabled {
		return
	}
	if sel := ed.selectedText(); sel != "" {
		rl.SetClipboardText(sel)
		return
	}
	rl.SetClipboardText(ed.Text.Get())
}

// Cut copies and removes the current selection.
func (ed *TextEditor) Cut() {
	if ed.Disabled || !ed.hasSelection() {
		return
	}
	rl.SetClipboardText(ed.selectedText())
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	ed.removeSelection()
	ed.markContentDirty()
	ed.flushContentDirty()
}

// Paste inserts clipboard text at the caret.
func (ed *TextEditor) Paste() {
	if ed.Disabled {
		return
	}
	clip := normalizeEditorClipboard(rl.GetClipboardText())
	if clip == "" {
		return
	}
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	ed.removeSelection()
	text := ed.Text.Get()
	ed.setText(text[:ed.cursor] + clip + text[ed.cursor:])
	ed.cursor += len(clip)
	ed.clearSelection()
	ed.markContentDirty()
	ed.Layout()
	ed.ensureCursorVisible()
	ed.flushContentDirty()
}

// Delete removes the current selection, or the character after the caret.
func (ed *TextEditor) Delete() {
	if ed.Disabled {
		return
	}
	if ed.hasSelection() {
		ed.bumpCaretAfterEdit()
		ed.pushUndo()
		ed.removeSelection()
		ed.markContentDirty()
		ed.flushContentDirty()
		return
	}
	ed.deleteAfterCursor()
}

// SelectAll selects the entire buffer.
func (ed *TextEditor) SelectAll() {
	text := ed.Text.Get()
	if len(text) == 0 {
		return
	}
	ed.selAnchor = 0
	ed.cursor = len(text)
	ed.ensureCursorVisible()
	ed.markContentDirty()
	ed.flushContentDirty()
	ed.RestartCaretCycle()
}

func normalizeEditorClipboard(s string) string {
	s = strings.TrimRight(s, "\x00")
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	// Trim runaway trailing blank lines from pasted files (not editable as "space").
	for strings.HasSuffix(s, "\n\n\n") {
		s = strings.TrimSuffix(s, "\n")
	}
	return s
}

// removeSelection deletes the selected range without recording undo.
func (ed *TextEditor) removeSelection() bool {
	if !ed.hasSelection() {
		return false
	}
	lo, hi := ed.selectionRange()
	text := ed.Text.Get()
	ed.setText(text[:lo] + text[hi:])
	ed.cursor = lo
	ed.selAnchor = -1
	ed.ensureCursorVisible()
	return true
}
