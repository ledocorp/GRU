package ui

import "strings"

// buildDocBlockquote renders markdown blockquotes: left accent bar + italic body
// (no full grey Card — matches reference preview).
func buildDocBlockquote(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	id := block.ID
	if id == "" {
		id = "blockquote"
	}
	row := NewContainer(id, 0, 0, docBlockWidthFromCtx(block, ctx), 0)
	row.SetStyle("transparent")
	row.FlexDirection = FlexRow
	row.Gap = 10
	row.AutoHeight = true
	applyDocStyle(&row.Element, block)

	bar := NewContainer(id+"-bar", 0, 0, 3, 0)
	bar.SetStyle("blockquote-accent")
	bar.AutoHeight = true

	spans := block.Spans
	if len(spans) == 0 && block.Text != "" {
		spans = []TextSpan{{Text: block.Text}}
	}
	spans = docApplySyntaxHighlightSpans(spans, ctx)
	for i := range spans {
		spans[i].Italic = true
	}
	rt := NewRichText(id+"-text", spans, 0, 0, 0, 0)
	rt.SetStyle("richtext-blockquote")
	rt.SetFlexGrow(1)
	applyDocRichText(rt, block, ctx)

	row.AddChild(bar)
	row.AddChild(rt)
	applyDocLayout(&row.Element, block)
	return row, nil
}

func formatCodeLang(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	if len(lang) == 1 {
		return strings.ToUpper(lang)
	}
	return strings.ToUpper(lang[:1]) + strings.ToLower(lang[1:])
}
