package preview

import (
	"strings"

	"github.com/yuin/goldmark/ast"

	"github.com/ledocorp/gru/ui"
)

// paragraphNode builds a paragraph RichText block.
func (ctx *mdBuildCtx) paragraphNode(id string, spans []ui.TextSpan) ui.Node {
	return ctx.newRichText(id, spans)
}

func (ctx *mdBuildCtx) buildParagraph(id string, p *ast.Paragraph) ui.Node {
	if img := paragraphImage(p); img != nil {
		return buildImageBlock(ctx, id, img, ctx.source)
	}
	// Detect $$ display math on raw paragraph text BEFORE expandInlineMath
	// strips the delimiters into a lone math-variant span.
	if body, ok := parseDisplayMathParagraph(strings.TrimSpace(blockPlainText(p, ctx.source))); ok {
		return ctx.buildMathCard(id, body, true)
	}
	spans := inlineSpans(p, ctx.source, inlineFlags{}, ctx)
	return ctx.paragraphNode(id, spans)
}
