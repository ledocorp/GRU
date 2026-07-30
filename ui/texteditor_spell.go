package ui

import (
	"strings"
	"unicode/utf8"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const textEditorSpellDebounceSec = float32(0.22)

var spellUnderlineColor = rl.NewColor(239, 48, 48, 255)

// textEditorCaretColor is the I-beam caret stroke (updated by SetAppearance).
var textEditorCaretColor = rl.NewColor(40, 42, 54, 255)

// SetSpellCheckEnabled turns red squiggle underlines on or off.
func (ed *TextEditor) SetSpellCheckEnabled(on bool) {
	if ed.spellEnabled == on {
		return
	}
	ed.spellEnabled = on
	ed.markSpellDirty()
}

// SpellCheckEnabled reports whether spell underlines are active.
func (ed *TextEditor) SpellCheckEnabled() bool { return ed.spellEnabled }

// SetSpellChecker sets the dictionary hook (nil disables checking).
func (ed *TextEditor) SetSpellChecker(c SpellChecker) {
	ed.spellChecker = c
	ed.markSpellDirty()
}

// SetSpellAutoCorrect enables replace-on-space when a table entry matches the prior word.
func (ed *TextEditor) SetSpellAutoCorrect(on bool) {
	ed.spellAutoCorrect = on
}

// SetSpellAutoCorrectTable sets lowercase typo → correction pairs (e.g. "teh" → "the").
func (ed *TextEditor) SetSpellAutoCorrectTable(m map[string]string) {
	ed.spellCorrect = m
}

func (ed *TextEditor) markSpellDirty() {
	if !ed.spellEnabled || ed.spellChecker == nil {
		ed.spellMiss = nil
		return
	}
	ed.spellDirty = true
	ed.spellTimer = textEditorSpellDebounceSec
}

// FlushSpellCheck recomputes misspellings immediately (e.g. after enabling in settings).
func (ed *TextEditor) FlushSpellCheck() {
	if !ed.spellEnabled || ed.spellChecker == nil {
		ed.spellMiss = nil
		ed.spellDirty = false
		return
	}
	ed.spellDirty = false
	ed.spellMiss = MisspelledRanges(ed.Text.Get(), ed.spellChecker)
	ed.markContentDirty()
}

func (ed *TextEditor) spellCheckActive() bool {
	return ed.spellEnabled && ed.spellChecker != nil
}

func (ed *TextEditor) tickSpellRefresh(dt float32) {
	if !ed.spellDirty || !ed.spellCheckActive() {
		return
	}
	ed.spellTimer -= dt
	if ed.spellTimer > 0 {
		return
	}
	ed.spellDirty = false
	ed.spellMiss = MisspelledRanges(ed.Text.Get(), ed.spellChecker)
	ed.markContentDirty()
}

// trySpellAutoCorrectBeforeSpace replaces the word ending at the cursor when it
// matches the auto-correct table. Call immediately before inserting a space.
func (ed *TextEditor) trySpellAutoCorrectBeforeSpace() bool {
	if !ed.spellAutoCorrect || len(ed.spellCorrect) == 0 {
		return false
	}
	text := ed.Text.Get()
	start, word := ed.wordEndingAt(text, ed.cursor)
	if word == "" {
		return false
	}
	key := strings.ToLower(word)
	repl, ok := ed.spellCorrect[key]
	if !ok || repl == "" || repl == word {
		return false
	}
	ed.replaceTextRange(start, ed.cursor, repl)
	return true
}

func (ed *TextEditor) wordEndingAt(text string, cursor int) (start int, word string) {
	if cursor <= 0 || cursor > len(text) {
		return 0, ""
	}
	i := cursor
	for i > 0 {
		r, size := utf8.DecodeLastRuneInString(text[:i])
		if !isSpellInnerRune(r) {
			break
		}
		i -= size
	}
	return i, text[i:cursor]
}

func (ed *TextEditor) replaceTextRange(lo, hi int, with string) {
	if lo < 0 || hi < lo {
		return
	}
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	ed.removeSelection()
	text := ed.Text.Get()
	if hi > len(text) {
		hi = len(text)
	}
	ed.setText(text[:lo] + with + text[hi:])
	ed.cursor = lo + len(with)
	ed.clearSelection()
	ed.ensureCursorVisible()
	ed.markContentDirty()
	ed.markSpellDirty()
}

func (ed *TextEditor) drawSpellUnderlines(text string, posX int32, baseY, lh float32, style Style, clip rl.Rectangle) {
	if len(ed.spellMiss) == 0 {
		return
	}
	lines, starts := ed.displayLines(style)
	first, last := ed.visibleLineRange(baseY, lh, clip, len(lines))
	for i := first; i <= last; i++ {
		line := lines[i]
		lineStart := starts[i]
		lineEnd := lineStart + len(line)
		y := baseY + float32(i)*lh
		for _, miss := range ed.spellMiss {
			lo, hi := miss[0], miss[1]
			if hi <= lineStart || lo >= lineEnd {
				continue
			}
			selLo := 0
			selHi := len(line)
			if lo > lineStart {
				selLo = lo - lineStart
			}
			if hi < lineEnd {
				selHi = hi - lineStart
			}
			if selLo >= selHi {
				continue
			}
			x1 := float32(posX) + float32(measureTextS(line[:selLo], style))
			w := float32(measureTextS(line[selLo:selHi], style))
			inkBottom := TextInkBottomY(float32(y), style)
			drawSpellSquiggle(x1, inkBottom, w)
		}
	}
}

// drawSpellSquiggle draws a dense red zigzag under a word (Windows Notepad–style).
func drawSpellSquiggle(x, baseline, w float32) {
	if w < 1 {
		return
	}
	const (
		segLen = float32(2.5) // horizontal step between points
		amp    = float32(1.25) // peak above/below baseline
	)
	xEnd := x + w
	px, py := x, baseline
	toggle := true
	for px < xEnd-0.25 {
		nx := px + segLen
		if nx > xEnd {
			nx = xEnd
		}
		ny := baseline + amp
		if toggle {
			ny = baseline - amp
		}
		rl.DrawLine(int32(px+0.5), int32(py+0.5), int32(nx+0.5), int32(ny+0.5), spellUnderlineColor)
		px, py = nx, ny
		toggle = !toggle
	}
}
