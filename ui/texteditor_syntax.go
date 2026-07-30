package ui

import (
	"sync/atomic"
)

const textEditorSyntaxDebounceSec = float32(0.55)

// TextEditorLargeBuffer — above this size syntax runs async on a worker (VS Code-style).
const TextEditorLargeBuffer = 8192

// textEditorMarkdownAsyncMin — markdown uses Chroma on the full buffer; run async
// above this size so medium/large .md files stay responsive like VS Code.
const textEditorMarkdownAsyncMin = 2048

var syntaxWorkerSeq atomic.Uint64

// SetDocument enables async syntax highlighting via QueueMain (call from scene Build).
func (ed *TextEditor) SetDocument(doc *Document) {
	ed.uiDoc = doc
}

// SetSyntaxLanguage enables Chroma highlighting (empty disables).
func (ed *TextEditor) SetSyntaxLanguage(lang string) {
	lang = NormalizeSyntaxLanguage(lang)
	if ed.syntaxLang == lang {
		return
	}
	ed.syntaxLang = lang
	ed.markSyntaxDirty()
}

// SyntaxLanguage returns the active Chroma lexer name, or "" when off.
func (ed *TextEditor) SyntaxLanguage() string { return ed.syntaxLang }

// FlushSyntaxHighlight recomputes highlight spans (async when buffer is non-trivial).
func (ed *TextEditor) FlushSyntaxHighlight() {
	if ed.syntaxLang == "" {
		ed.syntaxSpans = nil
		ed.syntaxDirty = false
		ed.clearSyntaxDirtyLines()
		return
	}
	if !ed.syntaxHighlightAllowed() {
		ed.syntaxDirty = false
		ed.syntaxSpans = nil
		ed.clearSyntaxDirtyLines()
		ed.markContentDirty()
		return
	}
	text := ed.Text.Get()
	if !ed.syntaxHighlightAsync(text) {
		ed.syntaxDirty = false
		ed.syntaxSpans = HighlightSyntax(text, ed.syntaxLang)
		ed.clearSyntaxDirtyLines()
		ed.markContentDirty()
		return
	}
	ed.syntaxDirty = false
	ed.scheduleSyntaxHighlight(text, ed.syntaxLang)
}

func (ed *TextEditor) markSyntaxDirty() {
	if ed.syntaxLang == "" {
		ed.syntaxSpans = nil
		return
	}
	line, _ := textOffsetLineCol(ed.Text.Get(), ed.cursor)
	if !ed.syntaxDirty || ed.syntaxDirtyLineStart < 0 {
		ed.syntaxDirtyLineStart = line
		ed.syntaxDirtyLineEnd = line
	} else {
		if line < ed.syntaxDirtyLineStart {
			ed.syntaxDirtyLineStart = line
		}
		if line > ed.syntaxDirtyLineEnd {
			ed.syntaxDirtyLineEnd = line
		}
	}
	ed.syntaxDirty = true
	ed.syntaxTimer = textEditorSyntaxDebounceSec
}

func (ed *TextEditor) clearSyntaxDirtyLines() {
	ed.syntaxDirtyLineStart = -1
	ed.syntaxDirtyLineEnd = -1
}

func (ed *TextEditor) lineInSyntaxDirtyRange(text string, byteStart int) bool {
	if !ed.syntaxDirty || ed.syntaxDirtyLineStart < 0 {
		return false
	}
	line, _ := textOffsetLineCol(text, byteStart)
	return line >= ed.syntaxDirtyLineStart && line <= ed.syntaxDirtyLineEnd
}

func (ed *TextEditor) tickSyntaxRefresh(dt float32) {
	if !ed.syntaxDirty || ed.syntaxLang == "" {
		return
	}
	ed.syntaxTimer -= dt
	if ed.syntaxTimer > 0 {
		return
	}
	ed.syntaxDirty = false
	if !ed.syntaxHighlightAllowed() {
		ed.syntaxSpans = nil
		ed.clearSyntaxDirtyLines()
		return
	}
	ed.scheduleSyntaxHighlight(ed.Text.Get(), ed.syntaxLang)
}

func (ed *TextEditor) scheduleSyntaxHighlight(text, lang string) {
	if ed.uiDoc == nil || lang == "" {
		return
	}
	if text == ed.syntaxCacheText && lang == ed.syntaxCacheLang && len(ed.syntaxCacheSpans) > 0 {
		if !syntaxSpansEqual(ed.syntaxSpans, ed.syntaxCacheSpans) {
			ed.syntaxSpans = ed.syntaxCacheSpans
			ed.markContentDirty()
			ed.flushContentDirty()
		}
		ed.clearSyntaxDirtyLines()
		return
	}
	ed.syntaxGen++
	gen := ed.syntaxGen
	job := syntaxWorkerSeq.Add(1)
	doc := ed.uiDoc
	go func() {
		spans := HighlightSyntax(text, lang)
		_ = job
		doc.QueueMain(func() {
			if ed.syntaxGen != gen || ed.syntaxLang != lang || ed.Text.Get() != text {
				return
			}
			ed.syntaxCacheText = text
			ed.syntaxCacheLang = lang
			ed.syntaxCacheSpans = spans
			ed.clearSyntaxDirtyLines()
			if syntaxSpansEqual(ed.syntaxSpans, spans) {
				return
			}
			ed.syntaxSpans = spans
			ed.markContentDirty()
			ed.flushContentDirty()
		})
	}()
}

func (ed *TextEditor) syntaxHighlightAllowed() bool {
	return ed.syntaxLang != ""
}

func (ed *TextEditor) syntaxHighlightAsync(text string) bool {
	if len(text) > TextEditorLargeBuffer {
		return true
	}
	return NormalizeSyntaxLanguage(ed.syntaxLang) == "markdown" && len(text) > textEditorMarkdownAsyncMin
}

func (ed *TextEditor) drawSyntaxLine(spans []TextSpan, x, y int32, base Style) {
	cx := float32(x)
	for _, sp := range spans {
		if sp.Text == "" {
			continue
		}
		st := base
		if sp.Color.A != 0 {
			st.TextColor = sp.Color
		}
		drawTextS(sp.Text, int32(cx), y, st)
		cx += EditorMeasureWidth(sp.Text, st)
	}
}

func syntaxSpansEqual(a, b []TextSpan) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i].Text != b[i].Text || a[i].Color != b[i].Color {
			return false
		}
	}
	return true
}

func syntaxSpansForRange(spans []TextSpan, start, length int) []TextSpan {
	if length <= 0 || len(spans) == 0 {
		return nil
	}
	end := start + length
	pos := 0
	var out []TextSpan
	for _, sp := range spans {
		if sp.Text == "" {
			continue
		}
		spStart := pos
		spEnd := pos + len(sp.Text)
		pos = spEnd
		if spEnd <= start {
			continue
		}
		if spStart >= end {
			break
		}
		lo := 0
		hi := len(sp.Text)
		if spStart < start {
			lo = start - spStart
		}
		if spEnd > end {
			hi = len(sp.Text) - (spEnd - end)
		}
		if lo >= hi {
			continue
		}
		seg := sp
		seg.Text = sp.Text[lo:hi]
		out = append(out, seg)
	}
	return out
}
