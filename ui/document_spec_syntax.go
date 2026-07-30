package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// docSyntaxHighlightEnabled reports whether Chroma colors and inline-code pill styling
// apply in the current DocumentSpec build scope. nil on BuildContext.SyntaxHighlight
// means enabled (default).
func docSyntaxHighlightEnabled(ctx *BuildContext) bool {
	if ctx == nil || ctx.SyntaxHighlight == nil {
		return true
	}
	return *ctx.SyntaxHighlight
}

// docBlockSyntaxHighlightFlag reads an optional syntaxHighlight override from a block
// (typed field or props map for .gru JSON).
func docBlockSyntaxHighlightFlag(block DocBlock) *bool {
	if block.SyntaxHighlight != nil {
		return block.SyntaxHighlight
	}
	if block.Props == nil {
		return nil
	}
	v, ok := block.Props["syntaxHighlight"].(bool)
	if !ok {
		return nil
	}
	return &v
}

// docCtxWithBlockSyntaxHighlight returns ctx scoped by block.SyntaxHighlight when set.
func docCtxWithBlockSyntaxHighlight(ctx *BuildContext, block DocBlock) *BuildContext {
	flag := docBlockSyntaxHighlightFlag(block)
	if flag == nil {
		return ctx
	}
	if ctx == nil {
		ctx = NewBuildContext()
	}
	c := *ctx
	c.SyntaxHighlight = flag
	return &c
}

// docApplySyntaxHighlightSpans strips inline-code pill styling when highlighting is off.
func docApplySyntaxHighlightSpans(spans []TextSpan, ctx *BuildContext) []TextSpan {
	if docSyntaxHighlightEnabled(ctx) || len(spans) == 0 {
		return spans
	}
	out := make([]TextSpan, len(spans))
	copy(out, spans)
	for i := range out {
		if out[i].Variant == "code" {
			out[i].Variant = ""
		}
		out[i].Color = rlColorClear()
	}
	return out
}

func rlColorClear() rl.Color {
	return rl.Color{}
}

// docCodeBlockSpans returns Chroma-colored spans when highlighting is enabled and a
// language is known; otherwise plain mono body text.
func docCodeBlockSpans(text, lang string, ctx *BuildContext) []TextSpan {
	text = strings.TrimRight(text, "\n")
	lang = strings.TrimSpace(lang)
	if !docSyntaxHighlightEnabled(ctx) || lang == "" {
		return []TextSpan{{Text: text}}
	}
	return HighlightSyntax(text, lang)
}
