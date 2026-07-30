package preview

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/ledocorp/gru/ui"
)

type inlineFlags struct {
	bold   bool
	italic bool
	strike bool
}

func inlineSpans(n ast.Node, source []byte, flags inlineFlags, ctx *mdBuildCtx) []ui.TextSpan {
	var spans []ui.TextSpan
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		spans = append(spans, inlineNode(c, source, flags, ctx)...)
	}
	if len(spans) == 0 {
		return []ui.TextSpan{{Text: ""}}
	}
	return expandInlineMath(mergeSpans(spans))
}

func inlineNode(n ast.Node, source []byte, flags inlineFlags, ctx *mdBuildCtx) []ui.TextSpan {
	switch nn := n.(type) {
	case *ast.Text:
		v := nn.Value(source)
		if len(v) == 0 {
			return nil
		}
		s := unescapeMarkdownPunct(string(v))
		// Soft breaks are line wraps in source, not hard paragraphs — keep flowing text.
		if nn.SoftLineBreak() {
			s += " "
		} else if nn.HardLineBreak() {
			s += "\n"
		}
		if s == "" {
			return nil
		}
		return []ui.TextSpan{applyFlags(ui.TextSpan{Text: s}, flags)}
	case *ast.String:
		if len(nn.Value) == 0 {
			return nil
		}
		return []ui.TextSpan{applyFlags(ui.TextSpan{Text: unescapeMarkdownPunct(string(nn.Value))}, flags)}
	case *ast.Emphasis:
		f := flags
		if nn.Level >= 2 {
			f.bold = true
		} else {
			f.italic = true
		}
		return inlineSpans(nn, source, f, ctx)
	case *extast.Strikethrough:
		f := flags
		f.strike = true
		return inlineSpans(nn, source, f, ctx)
	case *ast.CodeSpan:
		return []ui.TextSpan{applyFlags(ui.TextSpan{Text: codeSpanText(nn, source), Variant: "code"}, flags)}
	case *ast.Link:
		label := plainInline(inlineSpans(nn, source, flags, ctx))
		return []ui.TextSpan{applyFlags(ui.TextSpan{
			Text:      label,
			Link:      string(nn.Destination),
			LinkTitle: string(nn.Title),
		}, flags)}
	case *ast.Image:
		alt := plainInline(inlineSpans(nn, source, flags, ctx))
		if alt == "" {
			alt = "image"
		}
		return []ui.TextSpan{applyFlags(ui.TextSpan{
			Text:      alt,
			Link:      string(nn.Destination),
			LinkTitle: string(nn.Title),
			Variant:   "muted",
		}, flags)}
	case *extast.FootnoteLink:
		refKey := itoa(nn.Index)
		if ctx != nil {
			refKey = ctx.footnoteRef(nn.Index)
		}
		return []ui.TextSpan{{
			Text:    footnoteBracketLabel(refKey),
			Variant: "footnote-ref",
			Link:    "#" + footnoteDefAnchor(refKey),
		}}
	case *extast.FootnoteBacklink:
		// Goldmark injects backlinks into footnote bodies; we render our own return control.
		return nil
	case *ast.RawHTML:
		return htmlFragmentToSpans(string(nn.Text(source)), flags)
	default:
		if n.ChildCount() > 0 {
			return inlineSpans(n, source, flags, ctx)
		}
		return nil
	}
}

func codeSpanText(n *ast.CodeSpan, source []byte) string {
	return plainInline(inlineSpans(n, source, inlineFlags{}, nil))
}

func applyFlags(span ui.TextSpan, flags inlineFlags) ui.TextSpan {
	span.Bold = span.Bold || flags.bold
	span.Italic = span.Italic || flags.italic
	span.Strike = span.Strike || flags.strike
	return span
}

func plainInline(spans []ui.TextSpan) string {
	var b strings.Builder
	for _, s := range spans {
		b.WriteString(s.Text)
	}
	return b.String()
}

func mergeSpans(spans []ui.TextSpan) []ui.TextSpan {
	if len(spans) <= 1 {
		return spans
	}
	out := make([]ui.TextSpan, 0, len(spans))
	for _, sp := range spans {
		if len(out) > 0 && spansMergeable(out[len(out)-1], sp) {
			out[len(out)-1].Text += sp.Text
			continue
		}
		out = append(out, sp)
	}
	return out
}

func spansMergeable(a, b ui.TextSpan) bool {
	return a.Style == b.Style && a.Variant == b.Variant && a.Bold == b.Bold &&
		a.Italic == b.Italic && a.Strike == b.Strike && a.Link == b.Link &&
		a.LinkTitle == b.LinkTitle && a.Color == b.Color
}

func footnoteBodySpans(fn *extast.Footnote, source []byte, ctx *mdBuildCtx) []ui.TextSpan {
	var spans []ui.TextSpan
	for c := fn.FirstChild(); c != nil; c = c.NextSibling() {
		if len(spans) > 0 {
			spans = append(spans, ui.TextSpan{Text: " "})
		}
		spans = append(spans, inlineSpans(c, source, inlineFlags{}, ctx)...)
	}
	return spans
}

// sanitizeFootnoteBodySpans drops Goldmark backlink tokens and other empty runs.
func sanitizeFootnoteBodySpans(spans []ui.TextSpan) []ui.TextSpan {
	out := make([]ui.TextSpan, 0, len(spans))
	for _, sp := range spans {
		t := strings.TrimSpace(sp.Text)
		if t == "" || t == "↩" || t == "⏎" {
			continue
		}
		if sp.Link != "" && sp.Variant == "footnote-ref" {
			continue
		}
		out = append(out, sp)
	}
	return out
}

func listItemContent(li *ast.ListItem, source []byte, ctx *mdBuildCtx) (string, []ui.TextSpan) {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *ast.List, *extast.TaskCheckBox:
			continue
		default:
			spans := inlineSpans(c, source, inlineFlags{}, ctx)
			return strings.TrimSpace(plainInline(spans)), spans
		}
	}
	return "", nil
}

// parseInlineMarkdown is a small fallback for table cells (backticks, bold).
func parseInlineMarkdown(text string) []ui.TextSpan {
	return parseInlineFlags(text, inlineFlags{})
}

func parseInlineFlags(text string, flags inlineFlags) []ui.TextSpan {
	var spans []ui.TextSpan
	for len(text) > 0 {
		if text[0] == '`' {
			if end := strings.Index(text[1:], "`"); end >= 0 {
				spans = append(spans, applyFlags(ui.TextSpan{Text: text[1 : 1+end], Variant: "code"}, flags))
				text = text[1+end+1:]
				continue
			}
		}
		if strings.HasPrefix(text, "**") {
			if end := strings.Index(text[2:], "**"); end >= 0 {
				spans = append(spans, parseInlineFlags(text[2:2+end], inlineFlags{bold: true, strike: flags.strike, italic: flags.italic})...)
				text = text[2+end+2:]
				continue
			}
		}
		start := nextSpecial(text)
		if start > 0 {
			spans = append(spans, applyFlags(ui.TextSpan{Text: text[:start]}, flags))
			text = text[start:]
			continue
		}
		if start < 0 {
			spans = append(spans, applyFlags(ui.TextSpan{Text: text}, flags))
			break
		}
		spans = append(spans, applyFlags(ui.TextSpan{Text: text[:1]}, flags))
		text = text[1:]
	}
	if len(spans) == 0 {
		return []ui.TextSpan{{Text: ""}}
	}
	return mergeSpans(spans)
}

func nextSpecial(text string) int {
	next := -1
	for _, m := range []string{"**", "`"} {
		if idx := strings.Index(text, m); idx >= 0 && (next < 0 || idx < next) {
			next = idx
		}
	}
	return next
}

// unescapeMarkdownPunct mirrors goldmark's HTML writer: a '\' before ASCII
// punctuation is a CommonMark escape and must not appear in preview text.
func unescapeMarkdownPunct(s string) string {
	if !strings.ContainsRune(s, '\\') {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && isMarkdownEscapableByte(s[i+1]) {
			b.WriteByte(s[i+1])
			i++
			continue
		}
		b.WriteByte(s[i])
	}
	return b.String()
}

func isMarkdownEscapableByte(c byte) bool {
	switch c {
	case '!', '"', '#', '$', '%', '&', '\'', '(', ')', '*', '+', ',', '-', '.', '/',
		':', ';', '<', '=', '>', '?', '@', '[', '\\', ']', '^', '_', '`', '{', '|', '}', '~':
		return true
	default:
		return false
	}
}

