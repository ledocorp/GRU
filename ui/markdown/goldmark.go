// Package markdown registers a goldmark (GFM) DocumentSpec bridge into package ui.
//
//	import _ "github.com/ledocorp/gru/ui/markdown"
package markdown

import (
	"fmt"
	"strings"

	"github.com/ledocorp/gru/ui"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	gmtext "github.com/yuin/goldmark/text"
)

var goldmarkParser = goldmark.New(goldmark.WithExtensions(extension.GFM))

type inlineFlags struct {
	bold   bool
	italic bool
	strike bool
}

func applyInlineFlags(span ui.TextSpan, flags inlineFlags) ui.TextSpan {
	span.Bold = span.Bold || flags.bold
	span.Italic = span.Italic || flags.italic
	span.Strike = span.Strike || flags.strike
	return span
}

func flattenInlineSpans(spans []ui.TextSpan) []ui.TextSpan {
	if len(spans) <= 1 {
		return spans
	}
	out := make([]ui.TextSpan, 0, len(spans))
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

func spansMergeable(a, b ui.TextSpan) bool {
	return a.Style == b.Style && a.Variant == b.Variant && a.Bold == b.Bold &&
		a.Italic == b.Italic && a.Strike == b.Strike && a.Link == b.Link &&
		a.LinkTitle == b.LinkTitle && a.Color.A == b.Color.A
}

// markdownToDocumentSpecGoldmark parses Markdown with goldmark (GFM) and maps the
// AST to DocumentSpec. Markdown remains compile-time input only.
func markdownToDocumentSpecGoldmark(id, title, source string) ui.DocumentSpec {
	if id == "" {
		id = "markdown-document"
	}
	spec := ui.DocumentSpec{ID: id, Title: title}
	src := []byte(strings.ReplaceAll(source, "\r\n", "\n"))
	reader := gmtext.NewReader(src)
	doc := goldmarkParser.Parser().Parse(reader)
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		if block := goldmarkBlockToDoc(n, src); block.Type != "" {
			spec.Children = append(spec.Children, block)
		}
	}
	return spec
}

func goldmarkBlockToDoc(n ast.Node, source []byte) ui.DocBlock {
	switch nn := n.(type) {
	case *ast.Heading:
		level := nn.Level
		text := goldmarkPlainInline(nn, source)
		spans := goldmarkInlineSpans(nn, source, inlineFlags{})
		hv := fmt.Sprintf("h%d", level)
		for i := range spans {
			spans[i].Bold = true
			if spans[i].Variant == "" {
				spans[i].Variant = hv
			}
		}
		return ui.DocBlock{
			Type:  "text",
			ID:    ui.MarkdownAnchorSlug(text),
			Spans: spans,
		}
	case *ast.Paragraph:
		return ui.DocBlock{Type: "text", Spans: goldmarkInlineSpans(nn, source, inlineFlags{})}
	case *ast.FencedCodeBlock:
		return ui.DocBlock{Type: "code", Title: string(nn.Language(source)), Text: goldmarkFencedCode(nn, source)}
	case *ast.CodeBlock:
		return ui.DocBlock{Type: "code", Text: goldmarkCodeBlock(nn, source)}
	case *ast.ThematicBreak:
		return ui.DocBlock{Type: "divider"}
	case *ast.Blockquote:
		return ui.DocBlock{Type: "callout", Variant: "blockquote", Spans: goldmarkBlockquoteSpans(nn, source, 0)}
	case *ast.List:
		return goldmarkListBlock(nn, source)
	case *extast.Table:
		return goldmarkTableBlock(nn, source)
	default:
		return ui.DocBlock{}
	}
}

func goldmarkFencedCode(b *ast.FencedCodeBlock, source []byte) string {
	var buf strings.Builder
	lines := b.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(source[seg.Start:seg.Stop])
	}
	return strings.TrimRight(buf.String(), "\n")
}

func goldmarkCodeBlock(b *ast.CodeBlock, source []byte) string {
	var buf strings.Builder
	lines := b.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(source[seg.Start:seg.Stop])
	}
	return strings.TrimRight(buf.String(), "\n")
}

func goldmarkBlockquoteSpans(bq *ast.Blockquote, source []byte, depth int) []ui.TextSpan {
	var spans []ui.TextSpan
	indent := strings.Repeat("    ", depth)
	for c := bq.FirstChild(); c != nil; c = c.NextSibling() {
		if len(spans) > 0 {
			spans = append(spans, ui.TextSpan{Text: "\n"})
		}
		switch nn := c.(type) {
		case *ast.Blockquote:
			spans = append(spans, goldmarkBlockquoteSpans(nn, source, depth+1)...)
		default:
			line := goldmarkInlineSpans(nn, source, inlineFlags{})
			for i := range line {
				line[i].Italic = true
			}
			if indent != "" && len(line) > 0 {
				line[0].Text = indent + line[0].Text
			}
			spans = append(spans, line...)
		}
	}
	if len(spans) == 0 {
		return []ui.TextSpan{{Text: ""}}
	}
	return spans
}

type goldmarkListEntry struct {
	text  string
	spans []ui.TextSpan
	depth int
	num   int
	task  bool
	done  bool
}

func goldmarkListBlock(list *ast.List, source []byte) ui.DocBlock {
	var entries []goldmarkListEntry
	counters := make([]int, 12)
	goldmarkWalkList(list, source, 0, list.IsOrdered(), counters, &entries)
	if len(entries) == 0 {
		return ui.DocBlock{Type: "list"}
	}
	items := make([]string, len(entries))
	depths := make([]int, len(entries))
	nums := make([]int, len(entries))
	tasks := make([]bool, len(entries))
	done := make([]bool, len(entries))
	spanItems := make([]any, len(entries))
	for i, e := range entries {
		items[i] = e.text
		depths[i] = e.depth
		nums[i] = e.num
		tasks[i] = e.task
		done[i] = e.done
		if len(e.spans) > 0 {
			spanItems[i] = e.spans
		}
	}
	props := map[string]any{
		"depths":   depths,
		"numbers":  nums,
		"task":     tasks,
		"taskDone": done,
	}
	if len(spanItems) > 0 {
		props["itemSpans"] = spanItems
	}
	return ui.DocBlock{
		Type:    "list",
		Ordered: list.IsOrdered(),
		Items:   items,
		Props:   props,
	}
}

func goldmarkWalkList(list *ast.List, source []byte, depth int, parentOrdered bool, counters []int, out *[]goldmarkListEntry) {
	ordered := list.IsOrdered()
	if !parentOrdered && depth == 0 {
		parentOrdered = ordered
	}
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		li, ok := item.(*ast.ListItem)
		if !ok {
			continue
		}
		task, done := goldmarkListItemTask(li)
		text, spans := goldmarkListItemContent(li, source)
		if text != "" {
			goldmarkAppendListEntry(out, text, spans, depth, ordered, counters, task, done)
		}
		for c := li.FirstChild(); c != nil; c = c.NextSibling() {
			if nested, ok := c.(*ast.List); ok {
				goldmarkWalkList(nested, source, depth+1, nested.IsOrdered(), counters, out)
			}
		}
	}
}

func goldmarkAppendListEntry(out *[]goldmarkListEntry, text string, spans []ui.TextSpan, depth int, ordered bool, counters []int, task, done bool) {
	num := 0
	if ordered && depth < len(counters) {
		counters[depth]++
		for d := depth + 1; d < len(counters); d++ {
			counters[d] = 0
		}
		num = counters[depth]
	}
	*out = append(*out, goldmarkListEntry{text: text, spans: spans, depth: depth, num: num, task: task, done: done})
}

func goldmarkTableBlock(table *extast.Table, source []byte) ui.DocBlock {
	var header []string
	var bodyRows []map[string]any
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch row := child.(type) {
		case *extast.TableHeader:
			header = goldmarkTableRowCells(row, source)
		case *extast.TableRow:
			cells := goldmarkTableRowCells(row, source)
			if len(header) == 0 && len(bodyRows) == 0 {
				header = cells
				continue
			}
			bodyRows = append(bodyRows, goldmarkTableRowMap(header, cells))
		}
	}
	if len(header) == 0 {
		return ui.DocBlock{Type: "text", Spans: []ui.TextSpan{{Text: ""}}}
	}
	columns := make([]any, len(header))
	for j, title := range header {
		columns[j] = title
	}
	rows := make([]any, len(bodyRows))
	for i, m := range bodyRows {
		rows[i] = m
	}
	return ui.DocBlock{
		Type: "table",
		Props: map[string]any{
			"columns": columns,
			"rows":    rows,
		},
	}
}

func goldmarkTableRowCells(row ast.Node, source []byte) []string {
	var cells []string
	for c := row.FirstChild(); c != nil; c = c.NextSibling() {
		cell, ok := c.(*extast.TableCell)
		if !ok {
			continue
		}
		cells = append(cells, goldmarkTableCellText(cell, source))
	}
	return cells
}

func goldmarkTableCellText(cell *extast.TableCell, source []byte) string {
	var parts []string
	for c := cell.FirstChild(); c != nil; c = c.NextSibling() {
		if t := goldmarkPlainInline(c, source); t != "" {
			parts = append(parts, t)
		}
	}
	return strings.Join(parts, " ")
}

func goldmarkTableRowMap(header, cells []string) map[string]any {
	row := make(map[string]any, len(header))
	for j, col := range header {
		key := ui.DocTableFieldKey(col, j)
		val := ""
		if j < len(cells) {
			val = cells[j]
		}
		row[key] = val
	}
	return row
}

func goldmarkPlainInline(n ast.Node, source []byte) string {
	return ui.PlainInlineText(goldmarkInlineSpans(n, source, inlineFlags{}))
}

func goldmarkListItemTask(li *ast.ListItem) (task, done bool) {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		if cb, ok := c.(*extast.TaskCheckBox); ok {
			return true, cb.IsChecked
		}
	}
	return false, false
}

func goldmarkRawTextFromNode(n ast.Node, source []byte) string {
	var buf strings.Builder
	_ = ast.Walk(n, func(nn ast.Node, entering bool) (ast.WalkStatus, error) {
		if entering {
			if t, ok := nn.(*ast.Text); ok {
				seg := t.Segment
				buf.Write(source[seg.Start:seg.Stop])
			}
		}
		return ast.WalkContinue, nil
	})
	return buf.String()
}

func goldmarkListItemContent(li *ast.ListItem, source []byte) (string, []ui.TextSpan) {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *ast.List, *extast.TaskCheckBox:
			continue
		default:
			spans := goldmarkInlineSpans(c, source, inlineFlags{})
			text := strings.TrimSpace(ui.PlainInlineText(spans))
			if text != "" {
				return text, spans
			}
		}
	}
	text := goldmarkListItemMarkdown(li, source)
	if text == "" {
		return "", nil
	}
	return text, ui.ParseMarkdownInline(text)
}

func goldmarkListItemMarkdown(li *ast.ListItem, source []byte) string {
	var buf strings.Builder
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		if _, ok := c.(*ast.List); ok {
			continue
		}
		buf.WriteString(goldmarkRawTextFromNode(c, source))
	}
	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		return ""
	}
	trimmed := strings.TrimSpace(raw)
	if text, _, _, _, ok := ui.ParseListLine(trimmed); ok {
		return text
	}
	if strings.HasPrefix(trimmed, "- [ ] ") {
		return strings.TrimSpace(trimmed[6:])
	}
	if len(trimmed) >= 6 && (strings.HasPrefix(trimmed, "- [x] ") || strings.HasPrefix(trimmed, "- [X] ")) {
		return strings.TrimSpace(trimmed[6:])
	}
	return trimmed
}

func goldmarkCodeSpanText(n *ast.CodeSpan, source []byte) string {
	return ui.PlainInlineText(goldmarkInlineSpans(n, source, inlineFlags{}))
}

func goldmarkInlineSpans(n ast.Node, source []byte, flags inlineFlags) []ui.TextSpan {
	var spans []ui.TextSpan
	for c := n.FirstChild(); c != nil; c = c.NextSibling() {
		spans = append(spans, goldmarkInlineNode(c, source, flags)...)
	}
	if len(spans) == 0 {
		return []ui.TextSpan{{Text: ""}}
	}
	return flattenInlineSpans(spans)
}

func goldmarkInlineNode(n ast.Node, source []byte, flags inlineFlags) []ui.TextSpan {
	switch nn := n.(type) {
	case *ast.Text:
		seg := nn.Segment
		s := string(source[seg.Start:seg.Stop])
		if s == "" {
			return nil
		}
		return []ui.TextSpan{applyInlineFlags(ui.TextSpan{Text: s}, flags)}
	case *ast.String:
		s := string(nn.Value)
		if s == "" {
			return nil
		}
		return []ui.TextSpan{applyInlineFlags(ui.TextSpan{Text: s}, flags)}
	case *ast.Emphasis:
		f := flags
		if nn.Level >= 2 {
			f.bold = true
		} else {
			f.italic = true
		}
		return goldmarkInlineSpans(nn, source, f)
	case *extast.Strikethrough:
		f := flags
		f.strike = true
		return goldmarkInlineSpans(nn, source, f)
	case *ast.CodeSpan:
		return []ui.TextSpan{applyInlineFlags(ui.TextSpan{Text: goldmarkCodeSpanText(nn, source), Variant: "code"}, flags)}
	case *ast.Link:
		label := ui.PlainInlineText(goldmarkInlineSpans(nn, source, flags))
		return []ui.TextSpan{applyInlineFlags(ui.TextSpan{
			Text:      label,
			Link:      string(nn.Destination),
			LinkTitle: string(nn.Title),
		}, flags)}
	case *ast.Image:
		alt := ui.PlainInlineText(goldmarkInlineSpans(nn, source, flags))
		if alt == "" {
			alt = "image"
		}
		return []ui.TextSpan{applyInlineFlags(ui.TextSpan{
			Text:    "[Image: " + alt + "]",
			Link:    string(nn.Destination),
			Variant: "muted",
		}, flags)}
	default:
		if n.ChildCount() > 0 {
			return goldmarkInlineSpans(n, source, flags)
		}
		return nil
	}
}
