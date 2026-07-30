// Package examples — FlexCopy helpers for flex-column body copy in demos.
//
// Use FlexCopy / FlexCopyPair instead of NewLabel + form-* styles in scroll pages
// and panel bodies. Label is for fixed chrome only (see docs/HYGIENE_AUDIT.md §7).
package examples

import "github.com/ledocorp/gru/ui"

// FlexCopy is PlainText for flex-column copy (hints, captions, status, body text).
func FlexCopy(id, style, text string) *ui.RichText {
	return ui.NewPlainText(id, style, text, 0, 0, 0, 0)
}

// FlexCopyPair returns PlainText bound to a display signal you update.
func FlexCopyPair(id, style, initial string) (*ui.RichText, *ui.Signal[string]) {
	display := ui.NewSignal(initial)
	rt := ui.NewPlainText(id, style, "", 0, 0, 0, 0)
	ui.BindPlainText(rt, display)
	return rt, display
}

// FlexCopyMirror keeps PlainText in sync with src, optionally with a string prefix.
func FlexCopyMirror(id, style string, src *ui.Signal[string], prefix string) *ui.RichText {
	display := ui.NewSignal("")
	rt := ui.NewPlainText(id, style, "", 0, 0, 0, 0)
	ui.BindPlainText(rt, display)
	sync := func() {
		msg := src.Get()
		if prefix != "" {
			msg = prefix + msg
		}
		display.Set(msg)
	}
	src.Subscribe(sync)
	sync()
	return rt
}
