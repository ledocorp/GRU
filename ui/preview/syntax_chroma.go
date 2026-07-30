package preview

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

// highlightSpans tokenizes source with Chroma and maps to RichText spans.
func highlightSpans(source, language string) []ui.TextSpan {
	return ui.HighlightSyntax(source, language)
}

// highlightedCodeRichText builds a Chroma-colored RichText block for preview.
func highlightedCodeRichText(id, langRaw, text string) *ui.RichText {
	var spans []ui.TextSpan
	if strings.TrimSpace(langRaw) != "" {
		spans = highlightSpans(text, langRaw)
	} else {
		spans = []ui.TextSpan{{Text: text}}
	}
	rt := ui.NewRichText(id, spans, 0, 0, 0, 0)
	rt.SetStyle("richtext-code-block")
	rt.Wrap = false
	rt.AutoHeight = true
	rt.LineGap = 2
	return rt
}
