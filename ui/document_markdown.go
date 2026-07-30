package ui

import (
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// MarkdownToDocumentSpec converts Markdown (GFM via goldmark) into DocumentSpec
// when ui/markdown is imported; otherwise returns an empty DocumentSpec with ID/Title.
//
// Deprecated: Runtime preview and Notepad use [preview.MarkdownView] instead.
// This entry remains for compile-time tests and optional DocumentSpec tooling only.
//
// Markdown is compile-time authoring input only; callers compile with BuildDocumentSpec.
// Blank-import github.com/ledocorp/gru/ui/markdown to enable the goldmark bridge.
//
// See delete/_DEPRECATION_CANDIDATES.md.
func MarkdownToDocumentSpec(id, title, source string) DocumentSpec {
	return markdownToDocumentSpecGoldmark(id, title, source)
}

// ParseMarkdownInline parses inline Markdown markers into TextSpans.
// Exported for the optional ui/markdown goldmark bridge.
func ParseMarkdownInline(text string) []TextSpan {
	return parseMarkdownInline(text)
}

// PlainInlineText concatenates span text without styling.
func PlainInlineText(spans []TextSpan) string {
	return plainInlineText(spans)
}

// MarkdownAnchorSlug builds a heading fragment id from Markdown or plain text.
func MarkdownAnchorSlug(source string) string {
	return markdownAnchorSlug(source)
}

// ParseListLine detects unordered/ordered/task list item prefixes.
func ParseListLine(line string) (text string, ordered, task, done bool, ok bool) {
	return parseListLine(line)
}

func markdownAnchorSlug(source string) string {
	plain := plainInlineText(parseMarkdownInline(source))
	s := strings.ToLower(strings.TrimSpace(plain))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func plainInlineText(spans []TextSpan) string {
	var b strings.Builder
	for _, sp := range spans {
		b.WriteString(sp.Text)
	}
	return b.String()
}

func parseListLine(line string) (text string, ordered, task, done bool, ok bool) {
	if strings.HasPrefix(line, "- [ ] ") {
		return strings.TrimSpace(line[6:]), false, true, false, true
	}
	if strings.HasPrefix(line, "- [x] ") || strings.HasPrefix(line, "- [X] ") {
		return strings.TrimSpace(line[6:]), false, true, true, true
	}
	if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
		return strings.TrimSpace(line[2:]), false, false, false, true
	}
	dot := strings.IndexByte(line, '.')
	if dot <= 0 || dot+1 >= len(line) || line[dot+1] != ' ' {
		return "", false, false, false, false
	}
	if _, err := strconv.Atoi(line[:dot]); err != nil {
		return "", false, false, false, false
	}
	return strings.TrimSpace(line[dot+2:]), true, false, false, true
}

type inlineFlags struct {
	bold   bool
	italic bool
	strike bool
}

func parseMarkdownInline(text string) []TextSpan {
	return flattenInlineSpans(parseMarkdownInlineFlags(text, inlineFlags{}))
}

func parseMarkdownInlineFlags(text string, flags inlineFlags) []TextSpan {
	var spans []TextSpan
	for len(text) > 0 {
		if text[0] == '\\' && len(text) > 1 {
			r, size := utf8.DecodeRuneInString(text[1:])
			if size == 0 {
				size = 1
			}
			spans = append(spans, applyInlineFlags(TextSpan{Text: string(r)}, flags))
			text = text[1+size:]
			continue
		}
		start := markdownNextSpecial(text)
		if start > 0 {
			spans = append(spans, applyInlineFlags(TextSpan{Text: text[:start]}, flags))
			text = text[start:]
			continue
		}
		if start < 0 {
			spans = append(spans, applyInlineFlags(TextSpan{Text: text}, flags))
			break
		}
		switch {
		case strings.HasPrefix(text, "***"):
			if end := strings.Index(text[3:], "***"); end >= 0 {
				inner := text[3 : 3+end]
				spans = append(spans, parseMarkdownInlineFlags(inner, inlineFlags{bold: true, italic: true, strike: flags.strike})...)
				text = text[3+end+3:]
				continue
			}
		case strings.HasPrefix(text, "**"):
			if end := strings.Index(text[2:], "**"); end >= 0 {
				inner := text[2 : 2+end]
				spans = append(spans, parseMarkdownInlineFlags(inner, inlineFlags{bold: true, italic: flags.italic, strike: flags.strike})...)
				text = text[2+end+2:]
				continue
			}
		case strings.HasPrefix(text, "__"):
			if end := strings.Index(text[2:], "__"); end >= 0 {
				inner := text[2 : 2+end]
				spans = append(spans, parseMarkdownInlineFlags(inner, inlineFlags{bold: true, italic: flags.italic, strike: flags.strike})...)
				text = text[2+end+2:]
				continue
			}
		case strings.HasPrefix(text, "~~"):
			if end := strings.Index(text[2:], "~~"); end >= 0 {
				inner := text[2 : 2+end]
				spans = append(spans, parseMarkdownInlineFlags(inner, inlineFlags{bold: flags.bold, italic: flags.italic, strike: true})...)
				text = text[2+end+2:]
				continue
			}
		case strings.HasPrefix(text, "`"):
			if end := strings.Index(text[1:], "`"); end >= 0 {
				spans = append(spans, applyInlineFlags(TextSpan{Text: text[1 : 1+end], Variant: "code"}, flags))
				text = text[1+end+1:]
				continue
			}
		case strings.HasPrefix(text, "!["):
			if span, rest, ok := markdownImageSpan(text); ok {
				spans = append(spans, applyInlineFlags(span, flags))
				text = rest
				continue
			}
		case strings.HasPrefix(text, "["):
			if span, rest, ok := markdownLinkSpan(text); ok {
				spans = append(spans, applyInlineFlags(span, flags))
				text = rest
				continue
			}
		case strings.HasPrefix(text, "*"):
			if !strings.HasPrefix(text, "**") {
				if end := strings.Index(text[1:], "*"); end >= 0 {
					content := text[1 : 1+end]
					if content != "" && !unicode.IsSpace(rune(content[0])) {
						spans = append(spans, applyInlineFlags(TextSpan{Text: content, Italic: true}, flags))
						text = text[1+end+1:]
						continue
					}
				}
			}
		case strings.HasPrefix(text, "_"):
			if !strings.HasPrefix(text, "__") {
				if end := strings.Index(text[1:], "_"); end >= 0 {
					content := text[1 : 1+end]
					if content != "" && !unicode.IsSpace(rune(content[0])) {
						spans = append(spans, applyInlineFlags(TextSpan{Text: content, Italic: true}, flags))
						text = text[1+end+1:]
						continue
					}
				}
			}
		}
		spans = append(spans, applyInlineFlags(TextSpan{Text: text[:1]}, flags))
		text = text[1:]
	}
	if len(spans) == 0 {
		return []TextSpan{{Text: ""}}
	}
	return spans
}

func applyInlineFlags(span TextSpan, flags inlineFlags) TextSpan {
	span.Bold = span.Bold || flags.bold
	span.Italic = span.Italic || flags.italic
	span.Strike = span.Strike || flags.strike
	return span
}

func flattenInlineSpans(spans []TextSpan) []TextSpan {
	if len(spans) <= 1 {
		return spans
	}
	out := make([]TextSpan, 0, len(spans))
	for _, sp := range spans {
		if len(out) > 0 && spansMergeable(out[len(out)-1], sp) {
			last := &out[len(out)-1]
			last.Text += sp.Text
			continue
		}
		out = append(out, sp)
	}
	return out
}

func spansMergeable(a, b TextSpan) bool {
	return a.Style == b.Style && a.Variant == b.Variant && a.Bold == b.Bold &&
		a.Italic == b.Italic && a.Strike == b.Strike && a.Link == b.Link &&
		a.LinkTitle == b.LinkTitle && a.Color.A == b.Color.A
}

func markdownNextSpecial(text string) int {
	next := -1
	for _, marker := range []string{"\\", "***", "**", "__", "~~", "`", "![", "[", "*", "_"} {
		if idx := strings.Index(text, marker); idx >= 0 && (next < 0 || idx < next) {
			next = idx
		}
	}
	return next
}

func markdownLinkSpan(text string) (TextSpan, string, bool) {
	closeText := strings.Index(text, "](")
	if closeText <= 0 {
		return TextSpan{}, text, false
	}
	closeURL := strings.Index(text[closeText+2:], ")")
	if closeURL < 0 {
		return TextSpan{}, text, false
	}
	label := text[1:closeText]
	urlPart := text[closeText+2 : closeText+2+closeURL]
	url := strings.TrimSpace(urlPart)
	title := ""
	if qi := strings.Index(urlPart, " \""); qi >= 0 {
		url = strings.TrimSpace(urlPart[:qi])
		titlePart := strings.TrimSpace(urlPart[qi+1:])
		titlePart = strings.Trim(titlePart, "\"")
		title = titlePart
	}
	return TextSpan{Text: label, Link: url, LinkTitle: title}, text[closeText+2+closeURL+1:], true
}

func markdownImageSpan(text string) (TextSpan, string, bool) {
	if !strings.HasPrefix(text, "![") {
		return TextSpan{}, text, false
	}
	closeAlt := strings.Index(text, "](")
	if closeAlt <= 2 {
		return TextSpan{}, text, false
	}
	closeURL := strings.Index(text[closeAlt+2:], ")")
	if closeURL < 0 {
		return TextSpan{}, text, false
	}
	alt := text[2:closeAlt]
	url := text[closeAlt+2 : closeAlt+2+closeURL]
	label := alt
	if label == "" {
		label = "image"
	}
	return TextSpan{Text: "[Image: " + label + "]", Link: url, Variant: "muted"}, text[closeAlt+2+closeURL+1:], true
}
