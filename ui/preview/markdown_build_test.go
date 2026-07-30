package preview

import (
	"strings"
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestBuildMarkdownNodesHeadingsAndCode(t *testing.T) {
	src := "# Title\n\nParagraph with **bold**.\n\n```go\nfmt.Println(\"hi\")\n```"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) < 3 {
		t.Fatalf("got %d nodes, want at least 3", len(nodes))
	}
	if rt, ok := nodes[0].(*ui.RichText); !ok || !strings.Contains(plainInline(rt.Spans), "Title") {
		t.Fatalf("first node = %T, want heading RichText", nodes[0])
	}
	if _, ok := nodes[2].(*ui.Card); !ok {
		t.Fatalf("third node = %T, want code Card", nodes[2])
	}
}

func TestBuildMarkdownNodesTable(t *testing.T) {
	src := "| A | B |\n|---|---|\n| 1 | 2 |"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("got %d nodes, want 1 table card", len(nodes))
	}
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T, want Card", nodes[0])
	}
	if len(card.Children()) != 1 {
		t.Fatalf("card children = %d, want horizontal scroll host", len(card.Children()))
	}
	scroll, ok := card.Children()[0].(*ui.Viewport)
	if !ok {
		t.Fatalf("child = %T, want horizontal Viewport", card.Children()[0])
	}
	if scroll.GetFlexGrow() != 0 {
		t.Fatalf("table scroll FlexGrow = %v, want 0 (column parent — grow stretches height)", scroll.GetFlexGrow())
	}
	if card.ClipChildren {
		t.Fatal("table card should not clip children — horizontal scroll owns overflow")
	}
	rows := tableBodyRowCount(card)
	if rows != 2 {
		t.Fatalf("table rows = %d, want header + one body (align row hidden)", rows)
	}
}

// tableBodyRowCount walks Card → HorizontalViewport → lane and counts table rows.
func tableBodyRowCount(card *ui.Card) int {
	if card == nil || len(card.Children()) == 0 {
		return 0
	}
	scroll, ok := card.Children()[0].(*ui.Viewport)
	if !ok || len(scroll.Children()) == 0 {
		return 0
	}
	lane, ok := scroll.Children()[0].(*ui.Container)
	if !ok {
		return 0
	}
	return len(lane.Children())
}

func TestHeadingLevelsDistinctSize(t *testing.T) {
	src := "### Three\n\n#### Four\n"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 2 {
		t.Fatalf("got %d nodes, want 2 headings", len(nodes))
	}
	h3, ok := nodes[0].(*ui.RichText)
	if !ok {
		t.Fatalf("node 0 = %T, want RichText", nodes[0])
	}
	h4, ok := nodes[1].(*ui.RichText)
	if !ok {
		t.Fatalf("node 1 = %T, want RichText", nodes[1])
	}
	if h3.Spans[0].Variant != "h3" {
		t.Fatalf("h3 variant = %q", h3.Spans[0].Variant)
	}
	if h4.Spans[0].Variant != "h4" {
		t.Fatalf("h4 variant = %q", h4.Spans[0].Variant)
	}
	h3tok := ui.CurrentTheme["richtext-h3"].FontSize
	h4tok := ui.CurrentTheme["richtext-h4"].FontSize
	if h3tok <= h4tok {
		t.Fatalf("h3 token=%d h4 token=%d, want h3 larger", h3tok, h4tok)
	}
}

func TestCodeBlockHorizontalScrollHost(t *testing.T) {
	src := "package main\n\nfunc main() {\n\tfmt.Println(\"this line is intentionally very long so the code block should scroll horizontally instead of clipping at the card edge\")\n}"
	nodes := BuildMarkdownNodes("t", "```go\n"+src+"\n```")
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T, want Card", nodes[0])
	}
	var scroll *ui.Viewport
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if vp, ok := n.(*ui.Viewport); ok && vp.Orientation == ui.ScrollHorizontal {
			scroll = vp
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(card)
	if scroll == nil {
		t.Fatal("missing horizontal viewport for code block")
	}
}

func TestCodeBlockPreservesSmallMono(t *testing.T) {
	src := "```go\nfmt.Println(\"hi\")\n```"
	nodes := BuildMarkdownNodes("t", src)
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T, want Card", nodes[0])
	}
	var rt *ui.RichText
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if r, ok := n.(*ui.RichText); ok && r.StyleName() == "richtext-code-block" {
			rt = r
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(card)
	if rt == nil {
		t.Fatal("missing richtext-code-block")
	}
	if rt.GetStyle().FontSize != 14 {
		t.Fatalf("code block token=%d, want 14 (not surface-body 18)", rt.GetStyle().FontSize)
	}
}

func TestTableHeaderBold(t *testing.T) {
	src := "| **Header** | B |\n|:---|:---:|\n| 1 | 2 |"
	nodes := BuildMarkdownNodes("t", src)
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T", nodes[0])
	}
	found := false
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if rt, ok := n.(*ui.RichText); ok {
			for _, sp := range rt.Spans {
				if sp.Bold && strings.Contains(sp.Text, "Header") {
					found = true
				}
			}
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(card)
	if !found {
		t.Fatal("expected bold Header in table header cell")
	}
}

func TestMarkdownViewSetMarkdown(t *testing.T) {
	mv := NewMarkdownView("mv", 0, 0, 400, 300)
	mv.SetMarkdown("# Hello\n\nWorld.")
	if len(mv.Scroll().Children()) != 1 {
		t.Fatalf("scroll children = %d, want 1 lane", len(mv.Scroll().Children()))
	}
	mv.Update(0)
	if mv.building {
		t.Fatal("expected incremental build to finish after Update")
	}
}

func TestMarkdownViewStartBuildSingleLane(t *testing.T) {
	mv := NewMarkdownView("mv", 0, 0, 400, 300)
	mv.startIncrementalBuild(nil, nil)
	if len(mv.Scroll().Children()) != 1 {
		t.Fatalf("scroll children = %d, want 1", len(mv.Scroll().Children()))
	}
	mv.startIncrementalBuild(nil, nil)
	if len(mv.Scroll().Children()) != 1 {
		t.Fatalf("after second build scroll children = %d, want 1", len(mv.Scroll().Children()))
	}
}

func TestMarkdownViewBlockOrder(t *testing.T) {
	mv := NewMarkdownView("mv", 0, 0, 400, 600)
	mv.SetMarkdown("# First\n\nMiddle.\n\n## Second\n\nEnd.")
	mv.Update(0)
	lane := mv.lane
	if lane == nil {
		t.Fatal("no lane")
	}
	ids := childIDs(lane)
	if len(ids) < 3 {
		t.Fatalf("lane children = %v", ids)
	}
	if !strings.Contains(ids[0], "-b0") {
		t.Fatalf("first block id = %q, want -b0 (top of doc)", ids[0])
	}
	if !strings.Contains(ids[len(ids)-1], "-b3") {
		t.Fatalf("last block id = %q, want -b3 (end of doc)", ids[len(ids)-1])
	}
	if strings.Contains(ids[0], "-b3") {
		t.Fatal("document order is reversed")
	}
}

func childIDs(lane *ui.Container) []string {
	var ids []string
	for _, ch := range lane.Children() {
		ids = append(ids, ch.ID())
	}
	return ids
}

func TestMarkdownAnchorSlug(t *testing.T) {
	if got := markdownAnchorSlug("1-headings"); got != "1-headings" {
		t.Fatalf("slug = %q", got)
	}
	if got := markdownAnchorSlug("1. Headings"); got != "1-headings" {
		t.Fatalf("slug = %q, want 1-headings", got)
	}
}

func TestFootnoteBracketLabel(t *testing.T) {
	if footnoteBracketLabel("1") != "[1]" {
		t.Fatalf("label = %q", footnoteBracketLabel("1"))
	}
	if footnoteBracketLabel("demo") != "[demo]" {
		t.Fatalf("label = %q", footnoteBracketLabel("demo"))
	}
}

func TestNestedBlockquoteUsesNestedCard(t *testing.T) {
	src := "> Outer line.\n>\n> > Nested inner line."
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d, want 1 blockquote", len(nodes))
	}
	outer, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("outer = %T, want Card", nodes[0])
	}
	if len(outer.Children()) == 0 {
		t.Fatal("outer has no children")
	}
	if _, ok := outer.Children()[len(outer.Children())-1].(*ui.Card); !ok {
		t.Fatalf("nested = %T, want nested Card", outer.Children()[len(outer.Children())-1])
	}
}

func TestFootnoteBlocksNoDuplicateDefs(t *testing.T) {
	src := "Line with[^1] ref.\n\nLine two[^2].\n\n[^1]: First note.\n\n[^2]: Second note."
	lane := buildPreviewLane(t, src)
	n := countFootnoteDefs(lane)
	if n != 2 {
		t.Fatalf("footnote defs = %d, want 2", n)
	}
}

func TestUndefinedFootnoteNoPhantomShell(t *testing.T) {
	src := "Only a reference[^1] with no definition block."
	lane := buildPreviewLane(t, src)
	if n := countFootnoteDefs(lane); n != 0 {
		t.Fatalf("footnote defs = %d, want 0 for undefined ref", n)
	}
}

func TestFootnoteScrollTargetOffset(t *testing.T) {
	src := "Intro.\n\n## Section\n\nBody[^1] here.\n\nMore filler.\n\n[^1]: Footnote at the bottom."
	lane := buildPreviewLane(t, src)
	def := findFootnoteDef(lane, "1")
	if def == nil {
		t.Fatal("missing footnote definition")
	}
	inline := findFirstFootnoteRef(lane, "1")
	if inline == nil {
		t.Fatal("missing inline footnote ref")
	}
	if def == inline {
		t.Fatal("definition must not be the inline ref node")
	}
}

func TestFootnoteBodyRendersUserText(t *testing.T) {
	src := "Text with a note[^1] inline.\n\n[^1]: This is the footnote text at the bottom. It can use **formatting**."
	lane := buildPreviewLane(t, src)
	def := findFootnoteDef(lane, "1")
	rt, ok := def.(*ui.RichText)
	if !ok {
		t.Fatalf("footnote def = %T, want RichText", def)
	}
	text := plainInline(rt.Spans)
	if !strings.Contains(text, "footnote text at the bottom") {
		t.Fatalf("footnote body = %q", text)
	}
	if !strings.Contains(text, "formatting") {
		t.Fatalf("footnote body missing bold word: %q", text)
	}
	if strings.Contains(text, "It can contain") {
		t.Fatalf("unexpected showcase text in footnote: %q", text)
	}
}

func TestMarkdownViewUserFootnoteAfterShowcaseCache(t *testing.T) {
	showcaseSrc := "Body[^1] here.\n\n[^1]: This is the footnote text. It can contain **formatting** and even `code`.\n"
	ParseMarkdownBlocksCached(showcaseSrc)

	userSrc := "Text[^1] here.\n\n[^1]: This is the footnote text at the bottom. It can use **formatting**. Yay!"
	mv := NewMarkdownView("mv", 0, 0, 400, 600)
	mv.SetMarkdown(userSrc)
	for mv.building {
		mv.Update(0)
	}
	if mv.lane == nil {
		t.Fatal("no preview lane")
	}
	def := findFootnoteDef(mv.lane, "1")
	rt, ok := def.(*ui.RichText)
	if !ok {
		t.Fatalf("footnote def = %T, want RichText", def)
	}
	text := plainInline(rt.Spans)
	if !strings.Contains(text, "Yay!") {
		t.Fatalf("user footnote = %q, want Yay!", text)
	}
	if strings.Contains(text, "It can contain") {
		t.Fatalf("showcase footnote leaked into notepad preview: %q", text)
	}
}

func TestFootnotesDeferredToDocumentEnd(t *testing.T) {
	src := "First[^1] here.\n\n[^1]: Inline def.\n\n## Later\n\nMore body."
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) < 2 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	last := nodes[len(nodes)-1]
	if _, ok := last.(*ui.RichText); !ok {
		t.Fatalf("last node %T should be footnote RichText", last)
	}
}

func countFootnoteDefs(root ui.Node) int {
	n := 0
	walkPreviewNodes(root, func(node ui.Node) bool {
		rt, ok := node.(*ui.RichText)
		if !ok {
			return false
		}
		for _, sp := range rt.Spans {
			if sp.Variant == "footnote-ref" && sp.Link == "" && strings.HasPrefix(sp.Text, "[") {
				n++
				return false
			}
		}
		return false
	})
	return n
}

func TestFootnoteReturnUsesFirstRefSite(t *testing.T) {
	src := "First[^1] here.\n\nSecond[^1] later.\n\nAlso[^2].\n\n[^1]: Note one.\n\n[^2]: Note two."
	lane := buildPreviewLane(t, src)
	all := collectInlineFootnoteRefs(lane, "1")
	if len(all) < 2 {
		t.Fatalf("inline [^1] refs = %d, want at least 2", len(all))
	}
	first := findFirstFootnoteRef(lane, "1")
	if first != all[0] {
		t.Fatalf("return target = %p, want first inline ref %p", first, all[0])
	}
	if findFirstFootnoteRef(lane, "2") == nil {
		t.Fatal("missing [^2] ref site")
	}
}

func TestFootnoteAnchorMap(t *testing.T) {
	src := "Text[^1] and [^2].\n\n[^1]: Foot one.\n\n[^2]: Foot two."
	lane := buildPreviewLane(t, src)
	if findFootnoteDef(lane, "1") == nil {
		t.Fatal("missing fn-1 definition")
	}
	if findFootnoteDef(lane, "2") == nil {
		t.Fatal("missing fn-2 definition")
	}
	if findFirstFootnoteRef(lane, "1") == findFootnoteDef(lane, "1") {
		t.Fatal("return target must not be the footnote definition")
	}
}

func TestParseTaskPrefix(t *testing.T) {
	task, checked, body := parseTaskPrefix("[x] Done item")
	if !task || !checked || body != "Done item" {
		t.Fatalf("task=%v checked=%v body=%q", task, checked, body)
	}
	task, checked, body = parseTaskPrefix("[ ] Todo")
	if !task || checked || body != "Todo" {
		t.Fatalf("open task: task=%v checked=%v body=%q", task, checked, body)
	}
	task, _, _ = parseTaskPrefix("• normal")
	if task {
		t.Fatal("bullet should not be task")
	}
}

func TestBuildMarkdownTaskListIcons(t *testing.T) {
	src := "- [x] One\n- [ ] Two"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	col, ok := nodes[0].(*ui.Container)
	if !ok {
		t.Fatalf("root = %T", nodes[0])
	}
	var icons int
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if _, ok := n.(*ui.Icon); ok {
			icons++
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(col)
	if icons < 2 {
		t.Fatalf("icons = %d, want 2 task markers", icons)
	}
}

func TestParagraphFootnoteRefWrap(t *testing.T) {
	src := "See[^1] here.\n\n[^1]: Note."
	lane := buildPreviewLane(t, src)
	if findFirstFootnoteRef(lane, "1") == nil {
		t.Fatal("missing inline footnote ref")
	}
	if _, ok := findFirstFootnoteRef(lane, "1").(*ui.RichText); !ok {
		t.Fatalf("inline ref = %T, want RichText", findFirstFootnoteRef(lane, "1"))
	}
}

func TestHTMLFragmentToSpans(t *testing.T) {
	spans := htmlFragmentToSpans(`<p>This is a paragraph with <strong>HTML</strong>.</p>`, inlineFlags{})
	text := plainInline(spans)
	if !strings.Contains(text, "paragraph") || !strings.Contains(text, "HTML") {
		t.Fatalf("text = %q", text)
	}
	bold := false
	for _, sp := range spans {
		if sp.Bold && strings.Contains(sp.Text, "HTML") {
			bold = true
		}
	}
	if !bold {
		t.Fatal("expected bold HTML span")
	}
}
