// Package syntax registers Chroma-backed syntax highlighting into package ui.
//
// Blank-import this package to enable ui.HighlightSyntax / ui.SetChromaStyle:
//
//	import _ "github.com/ledocorp/gru/ui/syntax"
package syntax

import (
	"math"
	"strings"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const syntaxHighlightMaxBytes = 512 * 1024

func highlightSyntax(source, language string) []ui.TextSpan {
	source = strings.ReplaceAll(source, "\r\n", "\n")
	if len(source) > syntaxHighlightMaxBytes {
		return []ui.TextSpan{{Text: source}}
	}
	lang := ui.NormalizeSyntaxLanguage(language)
	if lang == "" {
		return []ui.TextSpan{{Text: source}}
	}
	if lang == "markdown" {
		return highlightMarkdownSource(source)
	}
	return chromaHighlightString(source, lang)
}

func chromaHighlightString(source, lang string) []ui.TextSpan {
	lexer := lexers.Get(lang)
	if lexer == nil {
		lexer = lexers.Analyse(source)
	}
	if lexer == nil {
		lexer = lexers.Get("text")
	}
	lexer = chroma.Coalesce(lexer)
	it, err := lexer.Tokenise(nil, source)
	if err != nil {
		return []ui.TextSpan{{Text: source}}
	}
	tokens := it.Tokens()
	if len(tokens) == 0 {
		return []ui.TextSpan{{Text: source}}
	}
	style := resolveChromaStyle()
	if style == nil {
		return []ui.TextSpan{{Text: source}}
	}
	spans := make([]ui.TextSpan, 0, len(tokens))
	editorBg := ui.GetThemeStyle("text-editor").BackgroundColor
	for _, tok := range tokens {
		if tok == chroma.EOF {
			continue
		}
		sp := ui.TextSpan{Text: tok.Value}
		if strings.TrimSpace(tok.Value) == "" {
			spans = append(spans, sp)
			continue
		}
		entry := style.Get(tok.Type)
		if col, ok := syntaxTokenColor(tok.Type, &entry, editorBg); ok {
			sp.Color = col
		}
		spans = append(spans, sp)
	}
	return mergeSyntaxSpans(spans)
}

func syntaxTokenColor(t chroma.TokenType, entry *chroma.StyleEntry, editorBg rl.Color) (rl.Color, bool) {
	if entry != nil && entry.Colour.IsSet() {
		c := entry.Colour
		col := rl.NewColor(c.Red(), c.Green(), c.Blue(), 255)
		if syntaxColorContrastsEditor(col, editorBg) {
			return col, true
		}
	}
	col := chromaTokenColor(t, entry)
	if syntaxColorContrastsEditor(col, editorBg) {
		return col, true
	}
	return rl.Color{}, false
}

// syntaxColorContrastsEditor rejects Chroma theme colours that vanish on the editor surface
// (e.g. github SQL whitespace #FFFFFF on a light editor, monokai #F8F8F2 on white).
func syntaxColorContrastsEditor(fg, bg rl.Color) bool {
	if fg.A == 0 {
		return false
	}
	const minContrast = 2.2
	return colorContrastRatio(fg, bg) >= minContrast
}

func colorContrastRatio(fg, bg rl.Color) float64 {
	l1 := colorRelativeLuminance(fg)
	l2 := colorRelativeLuminance(bg)
	if l1 < l2 {
		l1, l2 = l2, l1
	}
	if l2 <= 0 {
		return 21
	}
	return (l1 + 0.05) / (l2 + 0.05)
}

func colorRelativeLuminance(c rl.Color) float64 {
	srgb := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	r, g, b := srgb(c.R), srgb(c.G), srgb(c.B)
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func highlightMarkdownSource(source string) []ui.TextSpan {
	return highlightMarkdownEmbedded(source)
}

// highlightMarkdownEmbedded applies markdown highlighting to prose and defers to the
// fenced language (```sql, ```go, …) inside code blocks only.
func highlightMarkdownEmbedded(source string) []ui.TextSpan {
	if source == "" {
		return nil
	}
	var spans []ui.TextSpan
	i := 0
	for i < len(source) {
		fenceAt := nextMarkdownFenceOpen(source, i)
		if fenceAt < 0 {
			spans = append(spans, chromaSegment(source[i:], "markdown")...)
			break
		}
		if fenceAt > i {
			spans = append(spans, chromaSegment(source[i:fenceAt], "markdown")...)
		}
		headerEnd := strings.Index(source[fenceAt:], "\n")
		if headerEnd < 0 {
			spans = append(spans, chromaSegment(source[fenceAt:], "markdown")...)
			break
		}
		headerEnd += fenceAt
		lang := markdownFenceLang(source[fenceAt:headerEnd])
		contentStart := headerEnd
		if contentStart < len(source) && source[contentStart] == '\n' {
			contentStart++
		}
		if lang == "" && skipBareMarkdownFenceLine(source, fenceAt, contentStart) {
			lineEnd := contentStart
			spans = append(spans, chromaSegment(source[fenceAt:lineEnd], "markdown")...)
			i = lineEnd
			continue
		}
		closeAt := findMarkdownFenceClose(source, contentStart)
		if closeAt < 0 {
			// Unclosed fence: do not SQL-highlight the tail — keep as markdown.
			spans = append(spans, chromaSegment(source[fenceAt:], "markdown")...)
			break
		}
		spans = append(spans, chromaSegment(source[fenceAt:contentStart], "markdown")...)
		if lang != "" && closeAt > contentStart {
			spans = append(spans, chromaSegment(source[contentStart:closeAt], lang)...)
		} else if closeAt > contentStart {
			spans = append(spans, plainSegment(source[contentStart:closeAt])...)
		}
		closeLineEnd := closeAt
		for closeLineEnd < len(source) && source[closeLineEnd] != '\n' {
			closeLineEnd++
		}
		next := closeLineEnd
		if next < len(source) && source[next] == '\n' {
			next++
		}
		spans = append(spans, chromaSegment(source[closeAt:next], "markdown")...)
		i = next
	}
	return mergeSyntaxSpans(spans)
}

func chromaSegment(text, lang string) []ui.TextSpan {
	if text == "" {
		return nil
	}
	return chromaHighlightString(text, lang)
}

func plainSegment(text string) []ui.TextSpan {
	if text == "" {
		return nil
	}
	return []ui.TextSpan{{Text: text}}
}

func nextMarkdownFenceOpen(source string, from int) int {
	if at := markdownFenceLineAt(source, from); at >= 0 {
		return at
	}
	for {
		idx := strings.Index(source[from:], "\n```")
		if idx < 0 {
			return -1
		}
		at := from + idx + 1
		if markdownFenceLineAt(source, at) >= 0 {
			return at
		}
		from = at + 3
	}
}

func markdownFenceLineAt(source string, at int) int {
	if at+3 > len(source) || source[at:at+3] != "```" {
		return -1
	}
	if at > 0 && source[at-1] != '\n' {
		return -1
	}
	return at
}

func markdownFenceLang(openLine string) string {
	openLine = strings.TrimSpace(openLine)
	if !strings.HasPrefix(openLine, "```") {
		return ""
	}
	lang := ui.NormalizeSyntaxLanguage(strings.TrimSpace(strings.TrimPrefix(openLine, "```")))
	if lang == "markdown" || lang == "md" {
		return ""
	}
	return lang
}

// skipBareMarkdownFenceLine reports whether a ``` line without a language tag is
// prose (e.g. a tutorial line showing fence syntax) rather than a real code block.
func skipBareMarkdownFenceLine(source string, fenceAt, contentStart int) bool {
	if strings.TrimSpace(source[fenceAt:contentStart]) != "```" {
		return false
	}
	closeAt := findMarkdownFenceClose(source, contentStart)
	if closeAt < 0 {
		return true
	}
	inner := source[contentStart:closeAt]
	if strings.TrimSpace(inner) == "" {
		return false
	}
	if strings.Contains(inner, "\n```") {
		return true
	}
	first := strings.TrimSpace(inner)
	if i := strings.IndexByte(first, '\n'); i >= 0 {
		first = first[:i]
	}
	first = strings.TrimSpace(first)
	if strings.HasPrefix(first, "#") || strings.HasPrefix(first, ">") ||
		strings.HasPrefix(first, "- ") || strings.HasPrefix(first, "* ") ||
		strings.HasPrefix(first, "1.") {
		return true
	}
	return false
}

func findMarkdownFenceClose(source string, from int) int {
	for pos := from; pos <= len(source)-3; {
		if pos > from && source[pos-1] != '\n' {
			pos++
			continue
		}
		if source[pos:pos+3] != "```" {
			pos++
			continue
		}
		lineEnd := pos + 3
		for lineEnd < len(source) && source[lineEnd] != '\n' {
			lineEnd++
		}
		if strings.TrimSpace(source[pos:lineEnd]) == "```" {
			return pos
		}
		pos = lineEnd + 1
	}
	return -1
}

func mergeSyntaxSpans(spans []ui.TextSpan) []ui.TextSpan {
	if len(spans) <= 1 {
		return spans
	}
	out := make([]ui.TextSpan, 0, len(spans))
	for _, sp := range spans {
		if len(out) > 0 && syntaxSpansMergeable(out[len(out)-1], sp) {
			out[len(out)-1].Text += sp.Text
			continue
		}
		out = append(out, sp)
	}
	return out
}

func syntaxSpansMergeable(a, b ui.TextSpan) bool {
	return a.Style == b.Style && a.Variant == b.Variant && a.Bold == b.Bold &&
		a.Italic == b.Italic && a.Strike == b.Strike && a.Link == b.Link &&
		a.LinkTitle == b.LinkTitle && a.Color == b.Color
}

func chromaTokenColor(t chroma.TokenType, entry *chroma.StyleEntry) rl.Color {
	if entry != nil && entry.Colour.IsSet() {
		c := entry.Colour
		return rl.NewColor(c.Red(), c.Green(), c.Blue(), 255)
	}
	switch t {
	case chroma.Keyword, chroma.KeywordConstant, chroma.KeywordDeclaration, chroma.KeywordNamespace, chroma.KeywordType:
		return rl.NewColor(249, 38, 114, 255)
	case chroma.Name, chroma.NameFunction, chroma.NameClass, chroma.NameDecorator:
		return rl.NewColor(166, 226, 46, 255)
	case chroma.LiteralString, chroma.LiteralStringAffix, chroma.LiteralStringChar, chroma.LiteralStringDelimiter,
		chroma.LiteralStringEscape, chroma.LiteralStringHeredoc, chroma.LiteralStringInterpol, chroma.LiteralStringOther,
		chroma.LiteralStringRegex, chroma.LiteralStringSingle, chroma.LiteralStringSymbol:
		return rl.NewColor(230, 219, 116, 255)
	case chroma.LiteralNumber, chroma.LiteralNumberBin, chroma.LiteralNumberFloat, chroma.LiteralNumberHex,
		chroma.LiteralNumberInteger, chroma.LiteralNumberOct:
		return rl.NewColor(174, 129, 255, 255)
	case chroma.Comment, chroma.CommentMultiline, chroma.CommentSingle, chroma.CommentSpecial, chroma.CommentPreproc:
		return rl.NewColor(117, 113, 94, 255)
	case chroma.Operator, chroma.Punctuation:
		return rl.NewColor(248, 248, 242, 255)
	default:
		return rl.NewColor(248, 248, 242, 255)
	}
}
