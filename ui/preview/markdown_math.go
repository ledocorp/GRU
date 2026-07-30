package preview

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

func isMathLang(lang string) bool {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "latex", "math", "tex":
		return true
	default:
		return false
	}
}

func parseDisplayMathParagraph(text string) (body string, ok bool) {
	text = strings.TrimSpace(text)
	if !strings.HasPrefix(text, "$$") || !strings.HasSuffix(text, "$$") || len(text) < 4 {
		return "", false
	}
	body = strings.TrimSpace(text[2 : len(text)-2])
	return body, body != ""
}

func expandInlineMath(spans []ui.TextSpan) []ui.TextSpan {
	if len(spans) == 0 {
		return spans
	}
	var out []ui.TextSpan
	for _, sp := range spans {
		if sp.Link != "" || sp.Variant == "code" || sp.Variant == "footnote-ref" || sp.Variant == "math" {
			out = append(out, sp)
			continue
		}
		out = append(out, splitMathInSpan(sp)...)
	}
	return mergeSpans(out)
}

func inheritSpanStyle(parent ui.TextSpan, text, variant string) ui.TextSpan {
	sp := ui.TextSpan{
		Text:    text,
		Variant: variant,
		Bold:    parent.Bold,
		Italic:  parent.Italic,
		Strike:  parent.Strike,
		Color:   parent.Color,
	}
	if variant == "" {
		sp.Variant = parent.Variant
	}
	return sp
}

func splitMathInSpan(span ui.TextSpan) []ui.TextSpan {
	s := span.Text
	if s == "" || !strings.Contains(s, "$") {
		return []ui.TextSpan{span}
	}
	var out []ui.TextSpan
	i := 0
	for i < len(s) {
		if s[i] != '$' {
			j := strings.IndexByte(s[i:], '$')
			if j < 0 {
				out = append(out, inheritSpanStyle(span, s[i:], ""))
				break
			}
			out = append(out, inheritSpanStyle(span, s[i:i+j], ""))
			i += j
			continue
		}
		if i+1 < len(s) && s[i+1] == '$' {
			rest := s[i+2:]
			close := strings.Index(rest, "$$")
			if close < 0 {
				out = append(out, inheritSpanStyle(span, s[i:], ""))
				break
			}
			body := rest[:close]
			out = append(out, inheritSpanStyle(span, body, "math"))
			i += 2 + close + 2
			continue
		}
		rest := s[i+1:]
		close := strings.IndexByte(rest, '$')
		if close < 0 {
			out = append(out, inheritSpanStyle(span, s[i:], ""))
			break
		}
		body := rest[:close]
		if body != "" {
			out = append(out, inheritSpanStyle(span, body, "math"))
		}
		i += 1 + close + 1
	}
	if len(out) == 0 {
		return []ui.TextSpan{span}
	}
	return out
}
