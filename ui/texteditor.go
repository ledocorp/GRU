// Package ui (continued) — TextEditor multiline field.
//
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	textEditorCaretBlinkSec  = float32(4)
	textEditorCaretFrozenSec = float32(4)
	textEditorCaretStrokePx  = float32(1) // screen px; overlay is 1× after SSAA blit
	textEditorScrollBarSz    = float32(8)
)

type caretShowPhase int

const (
	caretPhaseBlink caretShowPhase = iota
	caretPhaseFrozen
	caretPhaseHidden
)

// TextEditor is a multiline plain-text field with a blinking caret, selection,
// undo/redo, clipboard, word wrap, horizontal scroll, and zoom.
//
// Caret lifecycle (idle FPS): blink (~4s, overlay on visible half-cycles only) →
// frozen (~4s, solid caret baked in SSAA cache; overlay off) → hidden (deep idle).
//
// Focus via [Document.SetFocus]. Emits [EventFocus] / [EventBlur] like [TextInput].
// Focus does not clear an existing selection (caret placement is via mouse/keys).
// Use [TextEditor.OnChange] to track dirty state in app scenes.
//
// Built-in keys when focused: Ctrl+Z/Y undo/redo, Ctrl+C/X/V/A clipboard/select-all,
// arrows/shift selection, Backspace/Delete. Optional spell-check squiggles and
// replace-on-space auto-correct via SetSpellCheckEnabled / SetSpellChecker.
// Scenes add file/zoom shortcuts in OnUpdate.
//
// Place inside a vertical [Viewport] for page scroll; set [TextEditor.AutoHeight] so
// content height drives the scroll range. Horizontal overflow uses an internal bar
// at the viewport bottom unless [TextEditor.WordWrap] is on.
//
// # LLM Prompt Template
//
//	vp := ui.NewViewport("editor-vp", 0, 0, 0, 0)
//	vp.SetFlexGrow(1)
//	ed := ui.NewTextEditor("editor", initialText, 0, 0, 0, 0)
//	ed.SetFlexGrow(1)
//	ed.AutoHeight = true
//	vp.AddChild(ed)
//	shell.AddChild(vp)
//	doc.SetFocus(ed)
//
// Primary reference: **Notepad (Go)** — examples/notepad_demo.go.
// Syntax samples: **Syntax Highlight Demo** — examples/syntax_highlight_demo.go.
//
// Example:
//
//	ed := ui.NewTextEditor("body", "", 0, 0, 0, 0)
//	ed.SetFlexGrow(1)
//	doc.SetFocus(ed)
type TextEditor struct {
	Element
	Text        *Signal[string]
	Placeholder string
	Disabled    bool
	OnChange    func()

	focused     bool
	hovered     bool
	lastHovered bool
	cursor      int
	selAnchor   int // -1 = no selection
	blinkTimer          float32
	blinkPhase          int
	caretPhase          caretShowPhase
	caretPhaseTimer     float32
	lastWindowFocused   bool
	hScrollDragging     bool
	hScrollDragMouse    float32
	hScrollDragScroll   float32
	dragSelect          bool
	selectPressMouse    rl.Vector2
	selectDragging      bool
	scrollY      float32
	scrollX      float32
	WordWrap     bool
	Zoom         float32 // 1.0 = 100%; scales editor type
	lastInnerW   float32
	frameDirty   bool // coalesce MarkDrawDirty to once per Update
	// Word-wrap line cache — reflow is expensive on large buffers.
	wrapCacheValid  bool
	wrapCacheText   string
	wrapCacheInnerW float32
	wrapCacheWrap   bool
	wrapCacheLines  []string
	wrapCacheStarts []int
	undoStack    []editorSnapshot
	redoStack    []editorSnapshot
	undoSuspend  bool

	syntaxLang  string
	syntaxSpans []TextSpan
	syntaxDirty bool
	syntaxTimer float32
	syntaxGen   int
	syntaxCacheText  string
	syntaxCacheLang  string
	syntaxCacheSpans []TextSpan
	uiDoc       *Document
	// Lines [start,end] still being edited; outside this range keep last highlight pass.
	syntaxDirtyLineStart int
	syntaxDirtyLineEnd   int

	spellEnabled    bool
	spellChecker    SpellChecker
	spellAutoCorrect bool
	spellCorrect    map[string]string
	spellMiss       [][2]int
	spellDirty      bool
	spellTimer      float32
}

// NewTextEditor creates a multiline editor. Pass h=0 with flex parent for fill height.
func NewTextEditor(id, text string, x, y, w, h float32) *TextEditor {
	ed := &TextEditor{
		Element:              NewElement(id, x, y, w, h),
		Text:                 NewSignal(text),
		cursor:               len(text),
		selAnchor:            -1,
		syntaxDirtyLineStart: -1,
		syntaxDirtyLineEnd:   -1,
	}
	ed.styleName = "text-editor"
	ed.Zoom = 1
	ed.ZIndex = 10
	ed.Text.Subscribe(func() {
		ed.invalidateDisplayLinesCache()
		ed.clampScrollY()
		ed.markSyntaxDirty()
		ed.markSpellDirty()
		ed.markContentDirty()
	})
	ed.On(EventFocus, func(Event) {
		ed.focused = !ed.Disabled
		if ed.focused {
			ed.RestartCaretCycle()
		}
	})
	ed.On(EventBlur, func(Event) {
		ed.focused = false
		ed.selAnchor = -1
		ed.caretPhase = caretPhaseHidden
		ed.caretPhaseTimer = 0
	})
	return ed
}

func (ed *TextEditor) editorStyle(base Style) Style {
	if ed.Zoom <= 0 || ed.Zoom == 1 {
		return base
	}
	s := base
	s.FontSize = int32(float32(base.FontSize) * ed.Zoom)
	if base.MinFontSize > 0 {
		s.MinFontSize = int32(float32(base.MinFontSize) * ed.Zoom)
	}
	return s
}

func (ed *TextEditor) lineHeight(style Style) float32 {
	return EffectiveFontSize(ed.editorStyle(style)) + 4
}

func (ed *TextEditor) markContentDirty() {
	ed.frameDirty = true
	if ed.focused {
		NoteTypingGesture()
	}
}

func (ed *TextEditor) flushContentDirty() {
	if ed.frameDirty {
		ed.frameDirty = false
		if ed.IsAutoHeight() {
			ed.Layout()
		}
		ed.MarkDrawDirty()
	}
}

// RestartCaretCycle begins blink → frozen → hidden (used on focus and window clicks).
func (ed *TextEditor) RestartCaretCycle() {
	if ed.Disabled {
		return
	}
	ed.caretPhase = caretPhaseBlink
	ed.caretPhaseTimer = textEditorCaretBlinkSec
	ed.blinkTimer = 0
	ed.blinkPhase = 0
}

func (ed *TextEditor) caretShown() bool {
	if ed.Disabled || !ed.keyboardActive() || !rl.IsWindowFocused() {
		return false
	}
	return ed.caretPhase == caretPhaseBlink || ed.caretPhase == caretPhaseFrozen
}

func (ed *TextEditor) hideCaret() {
	ed.caretPhase = caretPhaseHidden
	ed.caretPhaseTimer = 0
}

func (ed *TextEditor) updateCaretTimeline(dt float32) {
	if !ed.keyboardActive() || !rl.IsWindowFocused() {
		ed.hideCaret()
		return
	}
	switch ed.caretPhase {
	case caretPhaseBlink:
		ed.blinkTimer += dt
		ed.blinkPhase = int(ed.blinkTimer*2) % 2
		ed.caretPhaseTimer -= dt
		if ed.caretPhaseTimer <= 0 {
			ed.caretPhase = caretPhaseFrozen
			ed.caretPhaseTimer = textEditorCaretFrozenSec
			ed.blinkPhase = 0
			ed.MarkDrawDirty()
		}
	case caretPhaseFrozen:
		ed.caretPhaseTimer -= dt
		if ed.caretPhaseTimer <= 0 {
			ed.caretPhase = caretPhaseHidden
			ed.MarkDrawDirty()
		}
	case caretPhaseHidden:
		// idle
	}
}

func (ed *TextEditor) innerTextWidth(style Style) float32 {
	pad := ed.GetStyle().Padding
	if pad < 8 {
		pad = 8
	}
	w := ed.Bounds().Width - 2*pad
	if vp := findViewport(ed); vp != nil {
		w = vp.scrollContentWidthBudget(vp.Bounds()) - 2*pad
	}
	if w < 48 {
		w = 48
	}
	return w
}

func (ed *TextEditor) notifyChange() {
	if ed.OnChange != nil {
		ed.OnChange()
	}
}

func (ed *TextEditor) setText(s string) {
	ed.Text.Set(s)
	ed.notifyChange()
}

func (ed *TextEditor) hasSelection() bool {
	return ed.selAnchor >= 0 && ed.selAnchor != ed.cursor
}

func (ed *TextEditor) selectionRange() (lo, hi int) {
	if !ed.hasSelection() {
		return 0, 0
	}
	text := ed.Text.Get()
	n := len(text)
	lo, hi = ed.selAnchor, ed.cursor
	if lo > hi {
		lo, hi = hi, lo
	}
	if lo < 0 {
		lo = 0
	}
	if hi > n {
		hi = n
	}
	if lo >= hi {
		return 0, 0
	}
	return lo, hi
}

func (ed *TextEditor) selectedText() string {
	lo, hi := ed.selectionRange()
	if lo >= hi {
		return ""
	}
	return ed.Text.Get()[lo:hi]
}

// ClearSelection removes the active text selection (caret-only).
func (ed *TextEditor) ClearSelection() {
	ed.clearSelection()
}

func (ed *TextEditor) clearSelection() {
	if ed.selAnchor >= 0 {
		ed.selAnchor = -1
		ed.markContentDirty()
	}
}

func (ed *TextEditor) deleteSelection() bool {
	if !ed.hasSelection() {
		return false
	}
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	ed.removeSelection()
	ed.markContentDirty()
	return true
}

func (ed *TextEditor) bumpCaretAfterEdit() {
	if ed.caretPhase == caretPhaseHidden {
		ed.RestartCaretCycle()
	}
	ed.blinkPhase = 0
}

func (ed *TextEditor) insertAtCursor(insert string) {
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	ed.removeSelection()
	text := ed.Text.Get()
	ed.setText(text[:ed.cursor] + insert + text[ed.cursor:])
	ed.cursor += len(insert)
	ed.clearSelection()
	ed.invalidateDisplayLinesCache()
	ed.ensureCursorVisible()
	ed.markContentDirty()
}

func (ed *TextEditor) deleteBeforeCursor() bool {
	if ed.cursor <= 0 {
		return false
	}
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	text := ed.Text.Get()
	ed.setText(text[:ed.cursor-1] + text[ed.cursor:])
	ed.cursor--
	ed.ensureCursorVisible()
	ed.markContentDirty()
	return true
}

func (ed *TextEditor) deleteAfterCursor() bool {
	text := ed.Text.Get()
	if ed.cursor >= len(text) {
		return false
	}
	ed.bumpCaretAfterEdit()
	ed.pushUndo()
	ed.setText(text[:ed.cursor] + text[ed.cursor+1:])
	ed.ensureCursorVisible()
	ed.markContentDirty()
	return true
}

func (ed *TextEditor) moveCursor(delta int) {
	ed.bumpCaretAfterEdit()
	text := ed.Text.Get()
	ed.cursor += delta
	if ed.cursor < 0 {
		ed.cursor = 0
	}
	if ed.cursor > len(text) {
		ed.cursor = len(text)
	}
	ed.ensureCursorVisible()
	ed.markContentDirty()
}

func (ed *TextEditor) moveVertical(dir int) {
	ed.bumpCaretAfterEdit()
	text := ed.Text.Get()
	line, col := textOffsetLineCol(text, ed.cursor)
	line += dir
	if line < 0 {
		ed.cursor = 0
		ed.ensureCursorVisible()
		ed.markContentDirty()
		return
	}
	if ed.WordWrap {
		ed.moveVerticalWrapped(dir)
		return
	}
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		ed.cursor = len(text)
		ed.ensureCursorVisible()
		ed.markContentDirty()
		return
	}
	targetCol := col
	if targetCol > len(lines[line]) {
		targetCol = len(lines[line])
	}
	ed.cursor = lineColToOffset(text, line, targetCol)
	ed.ensureCursorVisible()
	ed.markContentDirty()
}

func (ed *TextEditor) moveVerticalWrapped(dir int) {
	style := ed.GetStyle()
	lines, _ := ed.displayLines(ed.editorStyle(style))
	if len(lines) == 0 {
		return
	}
	_, starts := ed.displayLines(ed.editorStyle(style))
	cur := visualLineIndex(starts, ed.cursor)
	next := cur + dir
	if next < 0 {
		ed.cursor = 0
	} else if next >= len(lines) {
		ed.cursor = len(ed.Text.Get())
	} else {
		col := ed.cursor
		if cur < len(starts) {
			col = ed.cursor - starts[cur]
		}
		if col < 0 {
			col = 0
		}
		line := lines[next]
		if col > len(line) {
			col = len(line)
		}
		start := 0
		if next < len(starts) {
			start = starts[next]
		}
		ed.cursor = start + col
		if ed.cursor > len(ed.Text.Get()) {
			ed.cursor = len(ed.Text.Get())
		}
	}
	ed.ensureCursorVisible()
	ed.markContentDirty()
}

func (ed *TextEditor) moveHome(end bool) {
	ed.bumpCaretAfterEdit()
	text := ed.Text.Get()
	line, _ := textOffsetLineCol(text, ed.cursor)
	if end {
		ed.cursor = lineColToOffset(text, line, len(lineAt(text, line)))
	} else {
		ed.cursor = lineColToOffset(text, line, 0)
	}
	ed.ensureCursorVisible()
	ed.markContentDirty()
}

// drawPlaceholder renders hint text wrapped to the editor inner width (always wraps).
func (ed *TextEditor) drawPlaceholder(posX int32, baseY, lh float32, style Style) {
	innerW := ed.innerTextWidth(style)
	lines := wrapEditorLines(ed.Placeholder, innerW, style)
	if len(lines) == 0 {
		drawTextS(ed.Placeholder, posX, int32(baseY), style)
		return
	}
	for i, line := range lines {
		drawTextS(line, posX, int32(baseY+float32(i)*lh), style)
	}
}

// displayLines returns drawable lines (wrapped when WordWrap is on).
func (ed *TextEditor) displayLines(style Style) ([]string, []int) {
	text := ed.Text.Get()
	innerW := ed.innerTextWidth(style)
	if ed.wrapCacheValid && text == ed.wrapCacheText &&
		innerW == ed.wrapCacheInnerW && ed.WordWrap == ed.wrapCacheWrap {
		return ed.wrapCacheLines, ed.wrapCacheStarts
	}
	lines, starts := ed.buildDisplayLines(text, innerW, style)
	ed.wrapCacheLines = lines
	ed.wrapCacheStarts = starts
	ed.wrapCacheText = text
	ed.wrapCacheInnerW = innerW
	ed.wrapCacheWrap = ed.WordWrap
	ed.wrapCacheValid = true
	return lines, starts
}

func (ed *TextEditor) invalidateDisplayLinesCache() {
	ed.wrapCacheValid = false
}

func (ed *TextEditor) buildDisplayLines(text string, innerW float32, style Style) ([]string, []int) {
	if !ed.WordWrap || innerW < 48 {
		parts := strings.Split(text, "\n")
		starts := make([]int, len(parts))
		off := 0
		for i, p := range parts {
			starts[i] = off
			off += len(p) + 1
		}
		return parts, starts
	}
	var lines []string
	var starts []int
	off := 0
	paras := strings.Split(text, "\n")
	for pi, para := range paras {
		wrapped := wrapEditorLines(para, innerW, style)
		if len(wrapped) == 0 {
			lines = append(lines, "")
			starts = append(starts, off)
		} else {
			pos := off
			for _, wl := range wrapped {
				lines = append(lines, wl)
				starts = append(starts, pos)
				pos += len(wl)
			}
		}
		off += len(para)
		if pi < len(paras)-1 {
			off++ // newline
		}
	}
	return lines, starts
}

// wrapEditorLines soft-wraps a paragraph into contiguous source substrings so
// byte offsets (caret, selection, mouse) stay aligned with the buffer.
// Does not use strings.Fields — spaces and repeated whitespace are preserved.
func wrapEditorLines(para string, maxW float32, style Style) []string {
	if maxW < 8 {
		if para == "" {
			return nil
		}
		return []string{para}
	}
	if para == "" {
		return []string{""}
	}
	limit := maxW - 4
	var lines []string
	i := 0
	for i < len(para) {
		if para[i] == '\n' {
			lines = append(lines, "")
			i++
			continue
		}
		best := i
		lastBreak := -1
		for j := i; j < len(para) && para[j] != '\n'; j++ {
			if float32(measureTextS(para[i:j+1], style)) > limit {
				break
			}
			best = j + 1
			if para[j] == ' ' || para[j] == '\t' {
				lastBreak = j + 1
			}
		}
		if best == i {
			if para[i] == ' ' || para[i] == '\t' {
				lines = append(lines, para[i:i+1])
				i++
				continue
			}
			rest := para[i:]
			chunks := breakWordToWidth(rest, maxW, style)
			if len(chunks) == 0 || chunks[0] == "" {
				lines = append(lines, para[i:i+1])
				i++
				continue
			}
			lines = append(lines, chunks[0])
			i += len(chunks[0])
			continue
		}
		cut := best
		if best < len(para) && lastBreak > i {
			cut = lastBreak
		}
		lines = append(lines, para[i:cut])
		i = cut
	}
	return lines
}

func breakWordToWidth(word string, maxW float32, style Style) []string {
	if word == "" {
		return nil
	}
	if float32(measureTextS(word, style)) <= maxW-4 {
		return []string{word}
	}
	var out []string
	rest := word
	for len(rest) > 0 {
		lo, hi := 1, len(rest)
		best := 1
		for lo <= hi {
			mid := (lo + hi) / 2
			if float32(measureTextS(rest[:mid], style)) <= maxW-4 {
				best = mid
				lo = mid + 1
			} else {
				hi = mid - 1
			}
		}
		out = append(out, rest[:best])
		rest = rest[best:]
	}
	return out
}

// absorbsHorizontalBarWheel reports whether the cursor is over the editor's
// horizontal scrollbar so the parent Viewport skips vertical wheel scroll.
func (ed *TextEditor) absorbsHorizontalBarWheel(mouse rl.Vector2) bool {
	if ed.IsHidden() || ed.Disabled || ed.WordWrap {
		return false
	}
	style := ed.editorStyle(ed.GetStyle())
	vp, _ := ed.horizontalScrollbarHost()
	track, _, maxX := ed.horizScrollbarGeom(vp, style)
	return maxX > 0 && rl.CheckCollisionPointRec(mouse, track)
}

func visualLineIndex(starts []int, cursor int) int {
	best := 0
	for i, start := range starts {
		if start <= cursor {
			best = i
		}
	}
	return best
}

func (ed *TextEditor) contentHeight(style Style) float32 {
	text := ed.Text.Get()
	if text == "" && ed.Placeholder != "" {
		lines := wrapEditorLines(ed.Placeholder, ed.innerTextWidth(style), style)
		if len(lines) == 0 {
			return ed.lineHeight(style)
		}
		return float32(len(lines)) * ed.lineHeight(style)
	}
	lines, _ := ed.displayLines(style)
	if len(lines) == 0 {
		return ed.lineHeight(style)
	}
	return float32(len(lines)) * ed.lineHeight(style)
}

func (ed *TextEditor) viewportInnerHeight(style Style) float32 {
	pad := style.Padding
	if pad < 8 {
		pad = 8
	}
	innerH := ed.Bounds().Height - 2*pad
	if vp := findViewport(ed); vp != nil {
		clip := vp.ClipBounds()
		innerH = clip.Height - 2*pad
	}
	if !ed.WordWrap && ed.maxScrollX(style) > 0 {
		innerH -= textEditorScrollBarSz
	}
	if innerH < ed.lineHeight(style) {
		return ed.lineHeight(style)
	}
	return innerH
}

func (ed *TextEditor) viewportInnerWidth(style Style) float32 {
	return ed.innerTextWidth(style)
}

func (ed *TextEditor) effectiveScrollY() float32 {
	if vp := findViewport(ed); vp != nil {
		return vp.ScrollY
	}
	return ed.scrollY
}

// textDrawBaseY is the Y baseline for the first text line. When the editor lives
// inside a Viewport, the viewport already scrolls the widget; subtracting
// ScrollY here would double-apply vertical scroll.
func (ed *TextEditor) textDrawBaseY(layoutBounds rl.Rectangle, pad int32) float32 {
	baseY := layoutBounds.Y + float32(pad)
	if findViewport(ed) == nil {
		baseY -= ed.scrollY
	}
	return baseY
}

func (ed *TextEditor) setEffectiveScrollY(y float32) {
	if vp := findViewport(ed); vp != nil {
		vp.ScrollY = y
		vp.scrollDirty = true
		vp.MarkDirty()
		return
	}
	ed.scrollY = y
}

func (ed *TextEditor) maxScrollY() float32 {
	style := ed.editorStyle(ed.GetStyle())
	contentH := ed.contentHeight(style)
	pad := ed.GetStyle().Padding
	if pad < 8 {
		pad = 8
	}
	contentH += 2 * pad
	innerH := ed.viewportInnerHeight(style)
	if contentH <= innerH {
		return 0
	}
	return contentH - innerH
}

func (ed *TextEditor) clampScrollY() {
	maxY := ed.maxScrollY()
	y := ed.effectiveScrollY()
	if y > maxY {
		y = maxY
	}
	if y < 0 {
		y = 0
	}
	ed.setEffectiveScrollY(y)
}

func (ed *TextEditor) maxScrollX(style Style) float32 {
	lines, _ := ed.displayLines(style)
	innerW := ed.innerTextWidth(ed.GetStyle()) - 4
	if innerW < 1 {
		innerW = 1
	}
	var maxLineW float32
	for _, line := range lines {
		w := float32(measureTextS(line, style))
		if w > maxLineW {
			maxLineW = w
		}
	}
	if maxLineW <= innerW {
		return 0
	}
	return maxLineW - innerW
}

func (ed *TextEditor) clampScrollX(style Style) {
	maxX := ed.maxScrollX(style)
	if ed.scrollX > maxX {
		ed.scrollX = maxX
	}
	if ed.scrollX < 0 {
		ed.scrollX = 0
	}
}

func (ed *TextEditor) ensureCursorVisible() {
	style := ed.editorStyle(ed.GetStyle())
	lh := ed.lineHeight(style)
	_, starts := ed.displayLines(style)

	vLine := visualLineIndex(starts, ed.cursor)

	top := float32(vLine) * lh
	bottom := top + lh
	pad := ed.GetStyle().Padding
	if pad < 8 {
		pad = 8
	}
	if vp := findViewport(ed); vp != nil {
		caretContentTop := pad + top
		caretContentBottom := caretContentTop + lh
		_, padT, _, padB := vp.scrollContentPadding()
		innerH := vp.Bounds().Height - padT - padB
		if innerH < lh {
			innerH = lh
		}
		margin := float32(8)
		scrollY := vp.ScrollY
		if caretContentTop < scrollY+margin {
			scrollY = caretContentTop - margin
		}
		if caretContentBottom > scrollY+innerH-margin {
			scrollY = caretContentBottom - innerH + margin
		}
		if scrollY < 0 {
			scrollY = 0
		}
		if vp.ScrollY != scrollY {
			vp.ScrollY = scrollY
			vp.clampScrollY()
			vp.scrollDirty = true
			vp.MarkDirty()
		}
	} else {
		innerH := ed.viewportInnerHeight(style)
		scrollY := ed.scrollY
		if top < scrollY {
			scrollY = top
		}
		if bottom-scrollY > innerH {
			scrollY = bottom - innerH
		}
		if ed.scrollY != scrollY {
			ed.scrollY = scrollY
			ed.clampScrollY()
		}
	}

	if !ed.WordWrap {
		text := ed.Text.Get()
		byteStart := 0
		if vLine < len(starts) {
			byteStart = starts[vLine]
		}
		if byteStart < 0 {
			byteStart = 0
		}
		if ed.cursor < byteStart {
			byteStart = ed.cursor
		}
		prefix := text[byteStart:ed.cursor]
		cursorX := float32(measureTextS(prefix, style))
		innerW := ed.innerTextWidth(ed.GetStyle()) - 4
		if innerW < 1 {
			innerW = 1
		}
		if cursorX < ed.scrollX {
			ed.scrollX = cursorX
		}
		if cursorX-ed.scrollX > innerW {
			ed.scrollX = cursorX - innerW
		}
		ed.clampScrollX(style)
	}
}

func lineAt(text string, line int) string {
	lines := strings.Split(text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

func textOffsetLineCol(text string, offset int) (line, col int) {
	if offset < 0 {
		offset = 0
	}
	if offset > len(text) {
		offset = len(text)
	}
	line = 0
	col = 0
	for i := 0; i < offset; i++ {
		if text[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}
	return line, col
}

func lineColToOffset(text string, line, col int) int {
	if line <= 0 {
		if col < 0 {
			return 0
		}
		lines := strings.Split(text, "\n")
		if len(lines) == 0 {
			return 0
		}
		if col > len(lines[0]) {
			return len(lines[0])
		}
		return col
	}
	lines := strings.Split(text, "\n")
	if line >= len(lines) {
		return len(text)
	}
	off := 0
	for i := 0; i < line; i++ {
		off += len(lines[i]) + 1
	}
	if col > len(lines[line]) {
		col = len(lines[line])
	}
	return off + col
}

func (ed *TextEditor) keyboardActive() bool {
	if ed.Disabled {
		return false
	}
	if WebViewHostHoldsKeyboard() {
		return false
	}
	if ed.focused {
		return true
	}
	d := ActiveDocument()
	return d != nil && d.Focused == ed
}

func (ed *TextEditor) Update(dt float32) {
	if ed.IsHidden() {
		return
	}
	ed.tickSyntaxRefresh(dt)
	ed.tickSpellRefresh(dt)
	mouse := rl.GetMousePosition()
	ed.hovered = !ed.Disabled && rl.CheckCollisionPointRec(mouse, ed.Bounds())
	// Click-to-focus runs after layout via RouteScenePointerFocus (main loop) so
	// bounds match the latched press; do not SetFocus here — stale flex bounds
	// break WebViewPanel handoff (§5 WEBVIEW2_HOST.md).
	if ed.hovered != ed.lastHovered {
		ed.lastHovered = ed.hovered
	}
	winFocused := rl.IsWindowFocused()
	if !winFocused {
		ed.hideCaret()
	} else if !ed.lastWindowFocused && ed.focused {
		ed.RestartCaretCycle()
	}
	ed.lastWindowFocused = winFocused
	if winFocused {
		ed.updateCaretTimeline(dt)
	}

	if !ed.Disabled {
		ed.handleHorizontalScrollbarInput()
	}

	wheel := rl.GetMouseWheelMove()
	if wheel != 0 && !ed.Disabled && !ChromeWindowMoving() {
		style := ed.editorStyle(ed.GetStyle())
		vp, _ := ed.horizontalScrollbarHost()
		track, _, maxX := ed.horizScrollbarGeom(vp, style)
		if !ed.WordWrap && maxX > 0 && rl.CheckCollisionPointRec(mouse, track) {
			ed.scrollX -= wheel * ed.lineHeight(style) * 0.75
			ed.clampScrollX(style)
			ed.markContentDirty()
		}
	}

	if ed.hovered && !ed.Disabled && !ScenePointerBlocked() {
		if rl.IsMouseButtonPressed(rl.MouseRightButton) {
			ed.ShowContextMenuAt(mouse)
		}
		ed.updateMouseSelection(mouse)
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 && !ChromeWindowMoving() {
			style := ed.editorStyle(ed.GetStyle())
			maxX := ed.maxScrollX(style)
			vertScroll := false
			if vp := findViewport(ed); vp != nil {
				vertScroll = vp.overflowScrollY() > 0
			}
			onHBar := false
			if maxX > 0 {
				vp, _ := ed.horizontalScrollbarHost()
				track, _, _ := ed.horizScrollbarGeom(vp, style)
				onHBar = rl.CheckCollisionPointRec(mouse, track)
			}
			if !ed.WordWrap && maxX > 0 && (onHBar || !vertScroll || textInputShiftDown() || ed.hScrollDragging) {
				ed.scrollX -= wheel * ed.lineHeight(style) * 0.75
				ed.clampScrollX(style)
				ed.markContentDirty()
			} else if findViewport(ed) == nil {
				y := ed.effectiveScrollY() - wheel*ed.lineHeight(style)*0.75
				ed.setEffectiveScrollY(y)
				ed.clampScrollY()
				ed.markContentDirty()
			}
		}
	}

	if !ed.keyboardActive() || ed.Disabled {
		if CaretDebugEnabled && (rl.IsKeyPressed(rl.KeySpace) || rl.GetKeyPressed() == int32(rl.KeySpace)) {
			CaretDebugLine(
				"space-blocked id=%s focused=%t disabled=%t cursor=%d (focus another widget?)",
				ed.ID(), ed.focused, ed.Disabled, ed.cursor,
			)
		}
		ed.flushContentDirty()
		return
	}

	keyFired := func(key int32) bool {
		return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
	}

	shift := textInputShiftDown()

	if textInputCtrlDown() {
		if rl.IsKeyPressed(rl.KeyZ) {
			if shift {
				ed.Redo()
			} else {
				ed.Undo()
			}
		} else if rl.IsKeyPressed(rl.KeyY) {
			ed.Redo()
		} else if rl.IsKeyPressed(rl.KeyC) {
			ed.Copy()
		} else if rl.IsKeyPressed(rl.KeyX) {
			ed.Cut()
		} else if rl.IsKeyPressed(rl.KeyV) {
			ed.Paste()
		} else if rl.IsKeyPressed(rl.KeyA) {
			ed.SelectAll()
		}
	}

	if keyFired(rl.KeyBackspace) {
		if !ed.deleteSelection() {
			ed.deleteBeforeCursor()
		}
	} else if keyFired(rl.KeyDelete) {
		if !ed.deleteSelection() {
			ed.deleteAfterCursor()
		}
	} else if keyFired(rl.KeyLeft) {
		if shift {
			if ed.selAnchor < 0 {
				ed.selAnchor = ed.cursor
			}
		} else {
			ed.clearSelection()
		}
		ed.moveCursor(-1)
	} else if keyFired(rl.KeyRight) {
		if shift {
			if ed.selAnchor < 0 {
				ed.selAnchor = ed.cursor
			}
		} else {
			ed.clearSelection()
		}
		ed.moveCursor(1)
	} else if keyFired(rl.KeyUp) {
		if shift {
			if ed.selAnchor < 0 {
				ed.selAnchor = ed.cursor
			}
		} else {
			ed.clearSelection()
		}
		ed.moveVertical(-1)
	} else if keyFired(rl.KeyDown) {
		if shift {
			if ed.selAnchor < 0 {
				ed.selAnchor = ed.cursor
			}
		} else {
			ed.clearSelection()
		}
		ed.moveVertical(1)
	} else if keyFired(rl.KeyHome) {
		if textInputCtrlDown() {
			if shift {
				if ed.selAnchor < 0 {
					ed.selAnchor = ed.cursor
				}
			} else {
				ed.clearSelection()
			}
			ed.cursor = 0
			ed.ensureCursorVisible()
			ed.bumpCaretAfterEdit()
			ed.markContentDirty()
		} else {
			if shift {
				if ed.selAnchor < 0 {
					ed.selAnchor = ed.cursor
				}
			} else {
				ed.clearSelection()
			}
			ed.moveHome(false)
		}
	} else if keyFired(rl.KeyEnd) {
		if textInputCtrlDown() {
			if shift {
				if ed.selAnchor < 0 {
					ed.selAnchor = ed.cursor
				}
			} else {
				ed.clearSelection()
			}
			ed.cursor = len(ed.Text.Get())
			ed.ensureCursorVisible()
			ed.bumpCaretAfterEdit()
			ed.markContentDirty()
		} else {
			if shift {
				if ed.selAnchor < 0 {
					ed.selAnchor = ed.cursor
				}
			} else {
				ed.clearSelection()
			}
			ed.moveHome(true)
		}
	} else if keyFired(rl.KeyEnter) {
		ed.insertAtCursor("\n")
	} else if keyFired(rl.KeyTab) {
		ed.insertAtCursor("    ")
	}

	spaceFromKey := keyFired(rl.KeySpace)
	if !spaceFromKey && rl.GetKeyPressed() == int32(rl.KeySpace) {
		spaceFromKey = true
	}
	if spaceFromKey {
		before := ed.cursor
		textBefore := ed.Text.Get()
		ed.clearSelection()
		ed.trySpellAutoCorrectBeforeSpace()
		ed.insertAtCursor(" ")
		CaretDebugLine(
			"space id=%s focused=%t cursor %d→%d textLen %d→%d tail=%q",
			ed.ID(), ed.focused, before, ed.cursor, len(textBefore), len(ed.Text.Get()), tailBytes(ed.Text.Get(), 12),
		)
	}

	char := rl.GetCharPressed()
	for char != 0 {
		if char >= 32 || char == '\t' {
			if char == ' ' {
				if !spaceFromKey {
					before := ed.cursor
					textBefore := ed.Text.Get()
					ed.clearSelection()
					ed.trySpellAutoCorrectBeforeSpace()
					ed.insertAtCursor(" ")
					CaretDebugLine(
						"space-char id=%s focused=%t cursor %d→%d textLen %d→%d tail=%q",
						ed.ID(), ed.focused, before, ed.cursor, len(textBefore), len(ed.Text.Get()), tailBytes(ed.Text.Get(), 12),
					)
				}
				char = rl.GetCharPressed()
				continue
			}
			ed.clearSelection()
			ed.insertAtCursor(string(char))
		}
		char = rl.GetCharPressed()
	}
	ed.flushContentDirty()
}

func (ed *TextEditor) Layout() {
	style := ed.editorStyle(ed.GetStyle())
	pad := ed.GetStyle().Padding
	if pad < 8 {
		pad = 8
	}
	wantH := ed.contentHeight(style) + 2*pad
	if wantH < ed.lineHeight(style)+2*pad {
		wantH = ed.lineHeight(style) + 2*pad
	}
	if vp := findViewport(ed); vp != nil {
		clip := vp.ClipBounds()
		if ed.IsAutoHeight() {
			b := ed.Bounds()
			changed := false
			if b.Height < wantH-0.5 || b.Height > wantH+0.5 {
				b.Height = wantH
				changed = true
			}
			if b.Width < clip.Width-0.5 || b.Width > clip.Width+0.5 {
				b.Width = clip.Width
				changed = true
			}
			if changed {
				ed.setBoundsNoMark(b)
				vp.MarkDirty()
			}
		}
	} else if ed.IsAutoHeight() {
		b := ed.Bounds()
		if b.Height < wantH-0.5 || b.Height > wantH+0.5 {
			b.Height = wantH
			ed.setBoundsNoMark(b)
		}
	}
	w := ed.innerTextWidth(ed.GetStyle())
	if ed.lastInnerW != 0 && (w > ed.lastInnerW+0.5 || w < ed.lastInnerW-0.5) {
		ed.invalidateDisplayLinesCache()
		if ed.WordWrap {
			ed.MarkDrawDirty()
		} else if ed.Text.Get() == "" && ed.Placeholder != "" {
			ed.markContentDirty()
		}
	}
	ed.lastInnerW = w
	ed.layoutDirty = false
}

func (ed *TextEditor) Draw() {
	defer func() { ed.drawDirty = false }()
	if ed.IsHidden() {
		return
	}
	baseStyle, _ := ed.ResolveStyle(StyleStateNone)
	drawStyle := ed.editorStyle(baseStyle)
	layoutBounds := ed.Bounds()
	pad := int32(drawStyle.Padding + 0.5)
	if pad < 8 {
		pad = 8
	}

	if drawStyle.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(layoutBounds, drawStyle.BackgroundColor)
	}

	var editorVP *Viewport
	innerBounds := layoutBounds
	if vp := findViewport(ed); vp != nil {
		editorVP = vp
		innerBounds = intersectRects(innerBounds, vp.ClipBounds())
		if innerBounds.Width <= 0 || innerBounds.Height <= 0 {
			return
		}
		if !ed.WordWrap && ed.maxScrollX(drawStyle) > 0 {
			innerBounds.Height -= textEditorScrollBarSz
			if innerBounds.Height < 1 {
				innerBounds.Height = 1
			}
		}
	}
	beginScissorMode(int32(innerBounds.X), int32(innerBounds.Y), int32(innerBounds.Width), int32(innerBounds.Height))

	text := ed.Text.Get()
	lh := ed.lineHeight(drawStyle)
	posX := int32(layoutBounds.X) + pad - int32(ed.scrollX)
	baseY := ed.textDrawBaseY(layoutBounds, pad)

	if ed.hasSelection() {
		ed.drawSelection(text, posX, baseY, lh, drawStyle, innerBounds)
	}
	if text == "" && ed.Placeholder != "" {
		ph := drawStyle
		ph.TextColor = rl.NewColor(drawStyle.TextColor.R, drawStyle.TextColor.G, drawStyle.TextColor.B, 120)
		ed.drawPlaceholder(posX, baseY, lh, ph)
	} else {
		lines, starts := ed.displayLines(drawStyle)
		first, last := ed.visibleLineRange(baseY, lh, innerBounds, len(lines))
		// Keep prior Chroma colors outside the edited line range; plain text only on dirty lines.
		useSyntax := ed.syntaxHighlightAllowed() && len(ed.syntaxSpans) > 0
		for i := first; i <= last; i++ {
			line := lines[i]
			y := int32(baseY + float32(i)*lh)
			byteStart := 0
			if i < len(starts) {
				byteStart = starts[i]
			}
			plainLine := ed.lineInSyntaxDirtyRange(text, byteStart)
			if useSyntax && !plainLine {
				length := len(line)
				if i < len(starts) {
					ed.drawSyntaxLine(syntaxSpansForRange(ed.syntaxSpans, starts[i], length), posX, y, drawStyle)
				} else {
					drawTextS(line, posX, y, drawStyle)
				}
			} else {
				drawTextS(line, posX, y, drawStyle)
			}
		}
		if ed.spellCheckActive() && len(ed.spellMiss) > 0 {
			ed.drawSpellUnderlines(text, posX, baseY, lh, drawStyle, innerBounds)
		}
	}

	if ed.caretShown() && ed.caretPhase == caretPhaseFrozen {
		ed.drawCaret(drawStyle, layoutBounds, pad, baseY, lh)
	}

	rl.EndScissorMode()

	if !ed.WordWrap {
		ed.drawHorizontalScrollbar(drawStyle, editorVP)
	}
}

func (ed *TextEditor) IsFocused() bool { return ed.keyboardActive() }

func (ed *TextEditor) IsInteractive() bool { return !ed.Disabled }

func (ed *TextEditor) UsesScissor() bool { return true }

// AnimationActive is false — caret blinks via InteractionOverlay so the document cache can idle.
func (ed *TextEditor) AnimationActive() bool { return false }

func (ed *TextEditor) AnimationSource() string { return ed.ID() }

func (ed *TextEditor) InteractionOverlayActive() bool {
	return ed.caretShown() && ed.caretPhase == caretPhaseBlink && ed.blinkPhase == 0
}

func (ed *TextEditor) DrawInteractionOverlay() {
	if !ed.caretShown() || ed.caretPhase != caretPhaseBlink {
		return
	}
	baseStyle, _ := ed.ResolveStyle(StyleStateNone)
	drawStyle := ed.editorStyle(baseStyle)
	layoutBounds := ed.Bounds()
	pad := int32(drawStyle.Padding + 0.5)
	if pad < 8 {
		pad = 8
	}
	innerBounds := layoutBounds
	if vp := findViewport(ed); vp != nil {
		innerBounds = intersectRects(innerBounds, vp.ClipBounds())
		if innerBounds.Width <= 0 || innerBounds.Height <= 0 {
			return
		}
		if !ed.WordWrap && ed.maxScrollX(drawStyle) > 0 {
			innerBounds.Height -= textEditorScrollBarSz
			if innerBounds.Height < 1 {
				innerBounds.Height = 1
			}
		}
	}
	beginScissorMode(int32(innerBounds.X), int32(innerBounds.Y), int32(innerBounds.Width), int32(innerBounds.Height))

	lh := ed.lineHeight(drawStyle)
	baseY := ed.textDrawBaseY(layoutBounds, pad)
	ed.drawCaret(drawStyle, layoutBounds, pad, baseY, lh)
	rl.EndScissorMode()
}

func (ed *TextEditor) horizontalScrollbarTrackRect(vp *Viewport, style Style) rl.Rectangle {
	maxScroll := ed.maxScrollX(style)
	if maxScroll <= 0 {
		return rl.Rectangle{}
	}
	host := ed.Bounds()
	if vp != nil {
		host = vp.Bounds()
	}
	trackY := host.Y + host.Height - textEditorScrollBarSz
	padL, _, padR := float32(0), float32(0), float32(0)
	gutterR := float32(0)
	if vp != nil {
		padL, _, padR, _ = vp.scrollContentPadding()
		if vp.overflowScrollY() > 0 {
			gutterR = vp.verticalScrollbarWidth() + vp.scrollBarLeadingGap()
		}
	}
	w := host.Width - padL - padR - gutterR
	if w < 1 {
		w = 1
	}
	return rl.NewRectangle(host.X+padL, trackY, w, textEditorScrollBarSz)
}

func (ed *TextEditor) horizScrollbarGeom(vp *Viewport, style Style) (track, thumb rl.Rectangle, maxScroll float32) {
	maxScroll = ed.maxScrollX(style)
	track = ed.horizontalScrollbarTrackRect(vp, style)
	if maxScroll <= 0 || track.Width < 16 {
		return track, thumb, 0
	}
	innerW := ed.viewportInnerWidth(style) - 4
	if innerW < 1 {
		innerW = 1
	}
	ratio := innerW / (innerW + maxScroll)
	thumbW := track.Width * ratio
	if thumbW < 24 {
		thumbW = 24
	}
	if thumbW > track.Width {
		thumbW = track.Width
	}
	scrollRatio := ed.scrollX / maxScroll
	thumbX := track.X + (track.Width-thumbW)*scrollRatio
	thumb = rl.NewRectangle(thumbX+1, track.Y+1, thumbW-2, textEditorScrollBarSz-2)
	return track, thumb, maxScroll
}

func (ed *TextEditor) drawHorizontalScrollbar(style Style, vp *Viewport) {
	track, thumb, maxScroll := ed.horizScrollbarGeom(vp, style)
	if maxScroll <= 0 {
		return
	}
	rl.DrawRectangleRounded(track, 1, 6, viewportScrollTrackColor)
	rl.DrawRectangleRounded(thumb, 1, 6, viewportScrollThumbColor)
}

func (ed *TextEditor) horizontalScrollbarHost() (*Viewport, rl.Rectangle) {
	vp := findViewport(ed)
	if vp != nil {
		return vp, vp.Bounds()
	}
	return nil, ed.Bounds()
}

func (ed *TextEditor) handleHorizontalScrollbarInput() {
	if ed.WordWrap {
		return
	}
	vp, _ := ed.horizontalScrollbarHost()
	style := ed.editorStyle(ed.GetStyle())
	track, thumb, maxScroll := ed.horizScrollbarGeom(vp, style)
	if maxScroll <= 0 {
		ed.hScrollDragging = false
		return
	}
	mouse := rl.GetMousePosition()

	if ed.hScrollDragging {
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
			ed.hScrollDragging = false
		} else {
			den := track.Width - thumb.Width
			if den > 0 {
				ed.scrollX = ed.hScrollDragScroll + (mouse.X-ed.hScrollDragMouse)*(maxScroll/den)
				ed.clampScrollX(style)
				ed.markContentDirty()
			}
		}
		return
	}

	if rl.IsMouseButtonDown(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, thumb) {
		ed.hScrollDragging = true
		ed.hScrollDragMouse = mouse.X
		ed.hScrollDragScroll = ed.scrollX
		PointerClickMarkUsed()
		return
	}
	if rl.CheckCollisionPointRec(mouse, track) &&
		(PointerClickConsume(track) || rl.IsMouseButtonPressed(rl.MouseLeftButton)) {
		r := (mouse.X - track.X) / track.Width
		if r < 0 {
			r = 0
		}
		if r > 1 {
			r = 1
		}
		ed.scrollX = r * maxScroll
		ed.clampScrollX(style)
		ed.markContentDirty()
		PointerClickMarkUsed()
	}
}

func (ed *TextEditor) caretInkColor() rl.Color {
	return textEditorCaretColor
}

func editorSyntaxPrefixWidth(spans []TextSpan, prefixBytes int, base Style) int32 {
	if prefixBytes <= 0 {
		return 0
	}
	var w int32
	pos := 0
	for _, sp := range spans {
		if sp.Text == "" {
			continue
		}
		n := len(sp.Text)
		if pos >= prefixBytes {
			break
		}
		st := base
		if sp.Color.A != 0 {
			st.TextColor = sp.Color
		}
		if pos+n <= prefixBytes {
			w += int32(EditorMeasureWidth(sp.Text, st))
			pos += n
			continue
		}
		w += int32(EditorMeasureWidth(sp.Text[:prefixBytes-pos], st))
		break
	}
	return w
}

// drawCaret paints the editor I-beam. Blink phase uses the interaction overlay;
// frozen phase bakes into the SSAA cache (see Draw).
func (ed *TextEditor) drawCaret(drawStyle Style, layoutBounds rl.Rectangle, pad int32, baseY, lh float32) {
	text := ed.Text.Get()
	lines, starts := ed.displayLines(drawStyle)
	vLine := visualLineIndex(starts, ed.cursor)
	byteStart := 0
	if vLine < len(starts) {
		byteStart = starts[vLine]
	}
	if byteStart < 0 {
		byteStart = 0
	}
	if ed.cursor < byteStart {
		byteStart = ed.cursor
	}
	prefix := text[byteStart:ed.cursor]
	posX := int32(layoutBounds.X) + pad - int32(ed.scrollX)
	var prefixW int32
	lineLen := 0
	if vLine < len(lines) {
		lineLen = len(lines[vLine])
	}
	if ed.syntaxHighlightAllowed() && len(ed.syntaxSpans) > 0 && !ed.lineInSyntaxDirtyRange(text, byteStart) && prefix != "" {
		spans := syntaxSpansForRange(ed.syntaxSpans, byteStart, lineLen)
		prefixW = editorSyntaxPrefixWidth(spans, len(prefix), drawStyle)
	} else {
		prefixW = measureTextS(prefix, drawStyle)
	}
	cursorX := posX + prefixW
	cursorY := int32(baseY + float32(vLine)*lh)
	fs := int32(EffectiveFontSize(drawStyle))
	rl.DrawLine(cursorX, cursorY, cursorX, cursorY+fs, ed.caretInkColor())
}

// SetWordWrap enables soft line wrapping to the editor inner width.
func (ed *TextEditor) SetWordWrap(on bool) {
	if ed.WordWrap == on {
		return
	}
	ed.WordWrap = on
	ed.invalidateDisplayLinesCache()
	ed.scrollX = 0
	ed.lastInnerW = 0
	ed.setEffectiveScrollY(0)
	ed.markContentDirty()
	ed.MarkDirty()
	ed.flushContentDirty()
}

// SetZoom sets editor scale (1.0 = 100%). Clamped 0.5–3.0.
func (ed *TextEditor) SetZoom(z float32) {
	if z < 0.5 {
		z = 0.5
	}
	if z > 3 {
		z = 3
	}
	ed.Zoom = z
	ed.markContentDirty()
}

// Insert appends text at the caret (e.g. demo fixtures, programmatic edits).
func (ed *TextEditor) Insert(insert string) {
	if ed.Disabled || insert == "" {
		return
	}
	ed.insertAtCursor(insert)
}

// SetTextContent replaces the buffer and moves the caret to the end (e.g. after Open).
func (ed *TextEditor) SetTextContent(text string) {
	ed.invalidateDisplayLinesCache()
	ed.Text.Set(text)
	ed.cursor = len(text)
	ed.selAnchor = -1
	ed.setEffectiveScrollY(0)
	ed.scrollX = 0
	ed.clearUndoHistory()
	ed.markContentDirty()
	ed.MarkDirty()
	ed.flushContentDirty()
}

// CursorLineCol returns 1-based line and column for status bars.
func (ed *TextEditor) CursorLineCol() (line, col int) {
	l, c := textOffsetLineCol(ed.Text.Get(), ed.cursor)
	return l + 1, c + 1
}

func (ed *TextEditor) visibleLineRange(baseY, lh float32, clip rl.Rectangle, lineCount int) (first, last int) {
	if lineCount <= 0 {
		return 0, -1
	}
	const margin = 1
	first = int((clip.Y-baseY)/lh) - margin
	if first < 0 {
		first = 0
	}
	last = int((clip.Y+clip.Height-baseY)/lh) + margin
	if last >= lineCount {
		last = lineCount - 1
	}
	if first > last {
		return 0, -1
	}
	return first, last
}

func (ed *TextEditor) drawSelection(text string, posX int32, baseY, lh float32, style Style, clip rl.Rectangle) {
	lo, hi := ed.selectionRange()
	if lo >= hi {
		return
	}
	lines, starts := ed.displayLines(style)
	fs := int32(EffectiveFontSize(style))
	first, last := ed.visibleLineRange(baseY, lh, clip, len(lines))
	for i := first; i <= last; i++ {
		line := lines[i]
		lineStart := starts[i]
		lineEnd := lineStart + len(line)
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
		x1 := posX + measureTextS(line[:selLo], style)
		x2 := posX + measureTextS(line[:selHi], style)
		y := int32(baseY + float32(i)*lh)
		rl.DrawRectangle(x1, y, x2-x1, fs, textInputSelectionFill)
	}
}
