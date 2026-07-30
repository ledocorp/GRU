// Package ui — PlainText helpers (single-span RichText with document flex contract).
package ui

// NewPlainText creates auto-height wrapped text with the same flex/reflow contract
// as DocumentSpec "text" blocks. Prefer this over Label for form captions, hints,
// status lines, and any copy that lives in a flex column and may wrap on resize.
//
// # LLM Prompt Template
//
//	hint := ui.NewPlainText("hint", "form-hint", "Helper copy wraps on resize.", 0, 0, 0, 0)
//	col.AddChild(hint)
//
// PlainText is a *RichText with Selectable=false and Wrap=true.
func NewPlainText(id, styleName, text string, x, y, w, h float32) *RichText {
	rt := NewRichText(id, []TextSpan{{Text: text}}, x, y, w, h)
	rt.SetStyle(styleName)
	rt.Selectable = false
	rt.Wrap = true
	return rt
}

// NewPlainTextSpans is like NewPlainText but accepts pre-built spans (variants, links).
func NewPlainTextSpans(id, styleName string, spans []TextSpan, x, y, w, h float32) *RichText {
	rt := NewRichText(id, spans, x, y, w, h)
	rt.SetStyle(styleName)
	rt.Selectable = false
	rt.Wrap = true
	return rt
}

// BindPlainText keeps a PlainText node's spans in sync with a string signal.
func BindPlainText(rt *RichText, text *Signal[string]) {
	if rt == nil || text == nil {
		return
	}
	sync := func() {
		rt.Spans = []TextSpan{{Text: text.Get()}}
		rt.InvalidateAutoHeightMeasure()
		rt.MarkDirty()
		markAutoHeightLayoutHostDirty(rt)
	}
	text.Subscribe(func() { sync() })
	sync()
}
