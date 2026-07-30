package preview

import (
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	extast "github.com/yuin/goldmark/extension/ast"
	gmtext "github.com/yuin/goldmark/text"

	"github.com/ledocorp/gru/ui"
)

var mdParser = goldmark.New(
	goldmark.WithExtensions(
		extension.GFM,
		extension.Footnote,
	),
)

// BuildMarkdownNodes parses GFM source and returns a vertical stack of preview nodes.
func BuildMarkdownNodes(idPrefix, source string) []ui.Node {
	return BuildMarkdownNodesWithHandler(idPrefix, source, nil, nil)
}

// ParseMarkdownBlocks parses GFM and returns top-level AST blocks (safe to use on a worker).
func ParseMarkdownBlocks(source string) (blocks []ast.Node, src []byte) {
	src = []byte(strings.ReplaceAll(source, "\r\n", "\n"))
	doc := mdParser.Parser().Parse(gmtext.NewReader(src))
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		blocks = append(blocks, n)
	}
	return reorderFootnotesLast(blocks), src
}

func parseDocument(source string) (ast.Node, []byte) {
	src := []byte(strings.ReplaceAll(source, "\r\n", "\n"))
	return mdParser.Parser().Parse(gmtext.NewReader(src)), src
}

// BuildMarkdownNodesWithHandler builds nodes and optionally fills anchors and link clicks.
func BuildMarkdownNodesWithHandler(idPrefix, source string, anchors map[string]ui.Node, onLink func(string)) []ui.Node {
	doc, src := parseDocument(source)
	ctx := newBuildCtx(idPrefix, anchors, onLink)
	ctx.source = src
	ctx.footnoteIndexToRef = collectFootnoteIndexToRef(doc)
	var blocks []ast.Node
	for n := doc.FirstChild(); n != nil; n = n.NextSibling() {
		blocks = append(blocks, n)
	}
	blocks = reorderFootnotesLast(blocks)
	var out []ui.Node
	for i, n := range blocks {
		id := fmtID(ctx.idPrefix, i)
		if fl, ok := n.(*extast.FootnoteList); ok {
			for _, def := range ctx.buildFootnoteDefs(id, fl) {
				if def != nil {
					out = append(out, def)
				}
			}
			continue
		}
		if node := ctx.blockToNode(id, n); node != nil {
			out = append(out, node)
		}
	}
	return out
}

// BuildMarkdownBlock builds one top-level block node.
func BuildMarkdownBlock(ctx *mdBuildCtx, id string, n ast.Node) ui.Node {
	if ctx == nil {
		return nil
	}
	return ctx.blockToNode(id, n)
}

func fmtID(prefix string, i int) string {
	return prefix + "-b" + itoa(i)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var b [12]byte
	n := len(b)
	for i > 0 {
		n--
		b[n] = byte('0' + i%10)
		i /= 10
	}
	return string(b[n:])
}

func (ctx *mdBuildCtx) blockToNode(id string, n ast.Node) ui.Node {
	switch nn := n.(type) {
	case *ast.Heading:
		return ctx.buildHeading(id, nn)
	case *ast.Paragraph:
		return ctx.buildParagraph(id, nn)
	case *ast.FencedCodeBlock:
		return ctx.buildCodeFence(id, nn)
	case *ast.CodeBlock:
		return ctx.buildCodeBlock(id, nn)
	case *ast.ThematicBreak:
		return ctx.buildThematicBreak(id)
	case *ast.Blockquote:
		return ctx.buildBlockquote(id, nn, 0)
	case *ast.List:
		return ctx.buildList(id, nn)
	case *extast.Table:
		return ctx.buildTable(id, nn)
	case *extast.FootnoteList:
		return nil // expanded by callers — one RichText per definition, no wrapper
	case *extast.Footnote:
		// Goldmark collects definitions in FootnoteList; loose nodes duplicate the bottom section.
		return nil
	case *ast.HTMLBlock:
		return ctx.buildHTMLBlock(id, nn)
	default:
		return nil
	}
}

func (ctx *mdBuildCtx) buildHTMLBlock(id string, b *ast.HTMLBlock) ui.Node {
	var buf strings.Builder
	lines := b.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(ctx.source[seg.Start:seg.Stop])
	}
	raw := strings.TrimSpace(buf.String())
	if raw == "" {
		return nil
	}
	spans := trimTrailingNewlineSpans(htmlFragmentToSpans(raw, inlineFlags{}))
	rt := ctx.newRichText(id+"-body", spans)
	card := ui.NewCard(id, "HTML", 0, 0, 0, 0)
	card.TitleHeight = 28
	card.AutoHeight = true
	card.Gap = 8
	card.SetStyleVariant("card", "code")
	if st, ok := ui.MergedComponentVariantStyle("card", "code"); ok {
		st.Padding = ui.DocMarkdownCardPadding
		card.SetStyleOverrides(st)
	}
	card.AddChild(rt)
	return card
}

func trimTrailingNewlineSpans(spans []ui.TextSpan) []ui.TextSpan {
	for len(spans) > 0 {
		last := spans[len(spans)-1]
		t := strings.TrimRight(last.Text, "\n")
		if t == last.Text {
			break
		}
		if t == "" {
			spans = spans[:len(spans)-1]
			continue
		}
		last.Text = t
		spans[len(spans)-1] = last
		break
	}
	return spans
}

func (ctx *mdBuildCtx) newRichText(id string, spans []ui.TextSpan) *ui.RichText {
	rt := ui.NewRichText(id, spans, 0, 0, 0, 0)
	rt.Wrap = true
	rt.AutoHeight = true
	rt.SetStyle("richtext-preview")
	ctx.wireRichText(rt)
	return rt
}

func (ctx *mdBuildCtx) buildHeading(id string, h *ast.Heading) ui.Node {
	spans := inlineSpans(h, ctx.source, inlineFlags{}, ctx)
	text := plainInline(spans)
	hv := "h" + itoa(h.Level)
	for i := range spans {
		switch spans[i].Variant {
		case "code", "footnote-ref", "math", "math-display", "muted":
			continue
		}
		spans[i].Bold = true
		spans[i].Variant = hv
	}
	rt := ctx.newRichText(id, spans)
	if slug := headingAnchorID(h.Level, text); slug != "" {
		ctx.registerAnchor(slug, rt)
	}
	return rt
}


func (ctx *mdBuildCtx) buildCodeFence(id string, b *ast.FencedCodeBlock) ui.Node {
	langRaw := strings.TrimSpace(string(b.Language(ctx.source)))
	text := fencedText(b, ctx.source)
	if isMathLang(langRaw) {
		return ctx.buildMathCard(id, text, true)
	}
	return ctx.buildCodeCard(id, formatLang(langRaw), langRaw, text)
}

func (ctx *mdBuildCtx) buildCodeBlock(id string, b *ast.CodeBlock) ui.Node {
	var buf strings.Builder
	lines := b.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(ctx.source[seg.Start:seg.Stop])
	}
	return ctx.buildCodeCard(id, "", "", strings.TrimRight(buf.String(), "\n"))
}

func (ctx *mdBuildCtx) buildCodeCard(id, langTitle, langRaw, text string) ui.Node {
	card := ui.NewCard(id, langTitle, 0, 0, 0, 0)
	if langTitle != "" {
		card.TitleHeight = 28
	} else {
		card.Title = ""
		card.TitleHeight = 0
	}
	card.AutoHeight = true
	card.Gap = 0
	card.SetStyleVariant("card", "code")
	if st, ok := ui.MergedComponentVariantStyle("card", "code"); ok {
		st.Padding = ui.DocMarkdownCardPadding
		card.SetStyleOverrides(st)
	}
	var spans []ui.TextSpan
	if strings.TrimSpace(langRaw) != "" {
		spans = highlightSpans(text, langRaw)
	} else {
		spans = []ui.TextSpan{{Text: text}}
	}
	rt := ui.NewRichText(id+"-code", spans, 0, 0, 0, 0)
	rt.SetStyle("richtext-code-block")
	rt.SetStyleOverrides(ui.Style{Padding: 0})
	rt.Wrap = false
	rt.AutoHeight = true
	rt.LineGap = 2
	ctx.wireRichText(rt)
	card.AddChild(ui.WrapCodeBlockHorizontalScroll(id, rt))
	return card
}

func (ctx *mdBuildCtx) buildThematicBreak(id string) ui.Node {
	wrap := ui.NewContainer(id, 0, 0, 0, 0)
	wrap.SetStyle("preview-hr-wrap")
	wrap.AutoHeight = true
	sep := ui.NewSeparator(id+"-rule", "", 0, 0, 0, 18)
	wrap.AddChild(sep)
	return wrap
}

// buildBlockquote renders blockquotes as Card shells (same pattern as code fences).
func (ctx *mdBuildCtx) buildBlockquote(id string, bq *ast.Blockquote, depth int) ui.Node {
	variant := "blockquote"
	if depth > 0 {
		variant = "blockquote-nested"
	}
	card := ui.NewCard(id, "", 0, 0, 0, 0)
	card.Title = ""
	card.TitleHeight = 0
	card.AutoHeight = true
	card.Gap = 8
	card.SetStyleVariant("card", variant)
	if st, ok := ui.MergedComponentVariantStyle("card", variant); ok {
		card.SetStyleOverrides(st)
	}

	childIdx := 0
	for c := bq.FirstChild(); c != nil; c = c.NextSibling() {
		switch nn := c.(type) {
		case *ast.Blockquote:
			if nested := ctx.buildBlockquote(id+"-nest"+itoa(childIdx), nn, depth+1); nested != nil {
				card.AddChild(nested)
			}
		case *ast.FencedCodeBlock:
			if node := ctx.buildCodeFence(id+"-code"+itoa(childIdx), nn); node != nil {
				card.AddChild(node)
			}
		case *ast.Paragraph:
			if img := paragraphImage(nn); img != nil {
				if node := buildImageBlock(ctx, id+"-img"+itoa(childIdx), img, ctx.source); node != nil {
					card.AddChild(node)
				}
			} else if body, ok := parseDisplayMathParagraph(strings.TrimSpace(blockPlainText(nn, ctx.source))); ok {
				if node := ctx.buildMathCard(id+"-math"+itoa(childIdx), body, true); node != nil {
					card.AddChild(node)
				}
			} else {
				spans := inlineSpans(nn, ctx.source, inlineFlags{}, ctx)
				if body, ok := parseDisplayMathParagraph(plainInline(spans)); ok {
					if node := ctx.buildMathCard(id+"-math"+itoa(childIdx), body, true); node != nil {
						card.AddChild(node)
					}
				} else {
					card.AddChild(ctx.blockquoteParagraph(id+"-p"+itoa(childIdx), spans))
				}
			}
		default:
			spans := inlineSpans(nn, ctx.source, inlineFlags{}, ctx)
			card.AddChild(ctx.blockquoteParagraph(id+"-p"+itoa(childIdx), spans))
		}
		childIdx++
	}
	if childIdx == 0 {
		return nil
	}
	return card
}

func (ctx *mdBuildCtx) blockquoteParagraph(id string, spans []ui.TextSpan) *ui.RichText {
	rt := ctx.newRichText(id, spans)
	rt.SetStyle("richtext-blockquote")
	return rt
}

func (ctx *mdBuildCtx) buildList(id string, list *ast.List) ui.Node {
	entries := flattenList(list, ctx.source, ctx, 0, list.IsOrdered(), make([]int, 8))
	col := ui.NewContainer(id, 0, 0, 0, 0)
	col.SetStyle("transparent")
	col.LayoutType = ui.LayoutFlex
	col.FlexDirection = ui.FlexColumn
	col.Gap = 10
	col.AutoHeight = true
	for i, e := range entries {
		row := ui.NewContainer(id+"-i"+itoa(i), 0, 0, 0, 0)
		row.SetStyle("transparent")
		row.LayoutType = ui.LayoutFlex
		row.FlexDirection = ui.FlexRow
		row.Gap = 10
		row.AutoHeight = true
		if e.depth > 0 {
			inset := ui.NewContainer(id+"-in"+itoa(i), 0, 0, float32(e.depth)*22, 0)
			inset.SetStyle("transparent")
			row.AddChild(inset)
		}
		if e.task {
			row.AddChild(taskListMarker(id+"-cb"+itoa(i), e.checked))
		}
		if !e.task {
			prefix := "• "
			if e.ordered && e.num > 0 {
				prefix = itoa(e.num) + ". "
			}
			e.spans = append([]ui.TextSpan{{Text: prefix, Variant: "list-marker"}}, e.spans...)
		}
		rt := ctx.newRichText(id+"-t"+itoa(i), e.spans)
		rt.SetFlexGrow(1)
		row.AddChild(rt)
		col.AddChild(row)
	}
	return col
}

func (ctx *mdBuildCtx) buildFootnoteDefs(id string, fl *extast.FootnoteList) []ui.Node {
	var out []ui.Node
	i := 0
	for c := fl.FirstChild(); c != nil; c = c.NextSibling() {
		if fn, ok := c.(*extast.Footnote); ok {
			if node := ctx.buildFootnoteDef(id+"-fn"+itoa(i), fn); node != nil {
				out = append(out, node)
			}
			i++
		}
	}
	return out
}

func (ctx *mdBuildCtx) buildFootnoteDef(id string, fn *extast.Footnote) ui.Node {
	ref := strings.TrimSpace(string(fn.Ref))
	if ref == "" {
		ref = ctx.footnoteRef(fn.Index)
	}
	defAnchor := footnoteDefAnchor(ref)

	body := sanitizeFootnoteBodySpans(footnoteBodySpans(fn, ctx.source, ctx))
	if strings.TrimSpace(plainInline(body)) == "" {
		return nil
	}

	spans := append([]ui.TextSpan{{Text: footnoteBracketLabel(ref) + " ", Variant: "footnote-ref"}}, body...)
	rt := ctx.newRichText(id+"-text", spans)
	rt.SetStyle("richtext-preview")
	rt.AutoHeight = true
	ctx.registerAnchor(defAnchor, rt)
	return rt
}

func (ctx *mdBuildCtx) buildTable(id string, table *extast.Table) ui.Node {
	pt := parseTable(table, ctx)
	if len(pt.colKeys) == 0 {
		return nil
	}
	ncols := len(pt.colKeys)
	colWidths := ctx.tableColumnWidths(pt)
	tableW := float32(0)
	for _, w := range colWidths {
		tableW += w
	}
	if ncols > 1 {
		tableW += float32(ncols-1) * 8 // row Gap between cells
	}

	lane := ui.NewContainer(id+"-lane", 0, 0, 0, 0)
	lane.LayoutType = ui.LayoutFlex
	lane.FlexDirection = ui.FlexColumn
	lane.Gap = 0
	lane.AutoHeight = true
	lane.MinWidth = tableW
	lane.PreferredWidth = tableW
	lane.SetFlexGrow(1) // fill host when wider than intrinsic table width
	lane.SetStyle("transparent")
	lane.AddChild(ctx.tableRowSpans(id+"-hdr", pt.colKeys, pt.headerSpans, true, colWidths))
	for ri, row := range pt.bodySpans {
		lane.AddChild(ctx.tableRowSpans(id+"-r"+itoa(ri), pt.colKeys, row, false, colWidths))
	}

	scroll := ui.NewHorizontalViewport(id+"-hscroll", 0, 0, 0, 0)
	scroll.AutoHeight = true
	scroll.Gap = 0
	// Do not SetFlexGrow here — parent card is a column; flex-grow expands height
	// (often to the intrinsic probe band) and distorts row layout.
	scroll.SetStyle("list-flush")
	scroll.AddChild(lane)

	outer := ui.NewCard(id, "", 0, 0, 0, 0)
	outer.Title = ""
	outer.TitleHeight = 0
	outer.AutoHeight = true
	outer.Gap = 0
	outer.ClipChildren = false
	outer.SetStyleVariant("card", "table")
	if st, ok := ui.MergedComponentVariantStyle("card", "table"); ok {
		st.Padding = ui.DocMarkdownCardPadding
		outer.SetStyleOverrides(st)
	}
	outer.AddChild(scroll)
	return outer
}

const tableColFloor = float32(56)
const tableColPad = float32(20)
// tableColMax caps column width so long cell copy wraps a few lines instead of
// forcing one ultra-wide column; the horizontal scroll host still handles overflow.
const tableColMax = float32(200)

func (ctx *mdBuildCtx) tableColumnWidths(pt parsedTable) []float32 {
	st := ui.GetThemeStyle("richtext-preview")
	fs := ui.EffectiveFontSize(st)
	n := len(pt.colKeys)
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		maxW := tableColFloor
		texts := []string{pt.colKeys[i]}
		if i < len(pt.headerSpans) {
			texts = append(texts, plainInline(pt.headerSpans[i]))
		}
		for _, row := range pt.bodySpans {
			if i < len(row) {
				texts = append(texts, plainInline(row[i]))
			}
		}
		for _, t := range texts {
			t = strings.TrimSpace(t)
			if t == "" {
				continue
			}
			w := ui.MeasureText(t, fs) + tableColPad
			if w > maxW {
				maxW = w
			}
		}
		if maxW > tableColMax {
			maxW = tableColMax
		}
		out[i] = maxW
	}
	return out
}

func (ctx *mdBuildCtx) tableRowSpans(id string, cols []string, cellSpans [][]ui.TextSpan, header bool, colWidths []float32) ui.Node {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	if header {
		row.SetStyle("table-header-row")
	} else {
		row.SetStyle("table-body-row")
	}
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 8
	row.AutoHeight = true
	for i := range cols {
		spans := []ui.TextSpan{{Text: ""}}
		if i < len(cellSpans) {
			spans = cellSpans[i]
		}
		colW := tableColFloor
		if i < len(colWidths) {
			colW = colWidths[i]
		}
		rt := ctx.newRichText(id+"-c"+itoa(i), spans)
		rt.PreferredWidth = colW
		rt.SetFlexGrow(1)
		row.AddChild(rt)
	}
	return row
}

type listEntry struct {
	spans   []ui.TextSpan
	depth   int
	num     int
	ordered bool
	task    bool
	checked bool
}

func flattenList(list *ast.List, source []byte, ctx *mdBuildCtx, depth int, ordered bool, counters []int) []listEntry {
	var out []listEntry
	for item := list.FirstChild(); item != nil; item = item.NextSibling() {
		li, ok := item.(*ast.ListItem)
		if !ok {
			continue
		}
		task, checked := taskItemFromListItem(li, source)
		var text string
		var spans []ui.TextSpan
		if task {
			spans = listItemBodySpans(li, source, ctx)
			text = strings.TrimSpace(plainInline(spans))
			if len(spans) == 0 {
				_, _, body := parseTaskPrefix(listItemPlainText(li, source))
				text = body
				if text != "" {
					spans = parseInlineMarkdown(text)
				}
			}
		} else {
			text, spans = listItemContent(li, source, ctx)
		}
		num := 0
		if ordered && depth < len(counters) {
			counters[depth]++
			for d := depth + 1; d < len(counters); d++ {
				counters[d] = 0
			}
			num = counters[depth]
		}
		if text != "" || task {
			if len(spans) == 0 && text != "" {
				spans = []ui.TextSpan{{Text: text}}
			}
			out = append(out, listEntry{
				spans: spans, depth: depth, num: num, ordered: ordered,
				task: task, checked: checked,
			})
		}
		for c := li.FirstChild(); c != nil; c = c.NextSibling() {
			if nested, ok := c.(*ast.List); ok {
				out = append(out, flattenList(nested, source, ctx, depth+1, nested.IsOrdered(), counters)...)
			}
		}
	}
	return out
}

func tableColKey(title string, index int) string {
	k := strings.ToLower(strings.TrimSpace(title))
	k = strings.ReplaceAll(k, " ", "_")
	if k == "" {
		return "col" + itoa(index)
	}
	return k
}

func formatLang(lang string) string {
	lang = strings.TrimSpace(lang)
	if lang == "" {
		return ""
	}
	if len(lang) == 1 {
		return strings.ToUpper(lang)
	}
	return strings.ToUpper(lang[:1]) + strings.ToLower(lang[1:])
}

func fencedText(b *ast.FencedCodeBlock, source []byte) string {
	var buf strings.Builder
	lines := b.Lines()
	for i := 0; i < lines.Len(); i++ {
		seg := lines.At(i)
		buf.Write(source[seg.Start:seg.Stop])
	}
	return strings.TrimRight(buf.String(), "\n")
}
