package ui

import (
	"strings"
	"testing"
)

func TestParseMarkdownInlineBoldUnderscore(t *testing.T) {
	spans := parseMarkdownInline("hello __world__")
	if len(spans) != 2 {
		t.Fatalf("got %d spans, want 2", len(spans))
	}
	if spans[0].Text != "hello " || spans[0].Bold {
		t.Fatalf("first span: %+v", spans[0])
	}
	if spans[1].Text != "world" || !spans[1].Bold {
		t.Fatalf("second span: %+v", spans[1])
	}
}

func TestMarkdownParagraphPreservesLineBreaks(t *testing.T) {
	spec := MarkdownToDocumentSpec("t", "T", "line one\nline two\n\nnext para")
	var texts []string
	for _, b := range spec.Children {
		if b.Type == "text" && len(b.Spans) > 0 {
			texts = append(texts, b.Spans[0].Text)
		}
	}
	if len(texts) < 2 {
		t.Fatalf("expected at least 2 text blocks, got %v", texts)
	}
	// CommonMark/goldmark joins soft line breaks within a paragraph into one line.
	if !strings.Contains(texts[0], "line one") || !strings.Contains(texts[0], "line two") {
		t.Fatalf("paragraph should include both lines, got %q", texts[0])
	}
}

func TestMarkdownHeadingInlineBold(t *testing.T) {
	spec := MarkdownToDocumentSpec("t", "T", "## Hello **world**")
	if len(spec.Children) != 1 {
		t.Fatalf("got %d blocks", len(spec.Children))
	}
	spans := spec.Children[0].Spans
	if len(spans) < 2 || !spans[1].Bold || spans[1].Text != "world" {
		t.Fatalf("heading inline bold: %+v", spans)
	}
}

func TestMarkdownHeadingNoFixedHeight(t *testing.T) {
	spec := MarkdownToDocumentSpec("t", "T", "# Title")
	if len(spec.Children) != 1 {
		t.Fatalf("got %d blocks", len(spec.Children))
	}
	if spec.Children[0].Height != 0 {
		t.Fatalf("headings should use AutoHeight, got height %v", spec.Children[0].Height)
	}
}

func TestParseMarkdownInlineStrike(t *testing.T) {
	spans := parseMarkdownInline("gone ~~soon~~")
	if len(spans) != 2 || !spans[1].Strike || spans[1].Text != "soon" {
		t.Fatalf("strike span: %+v", spans)
	}
}

func TestParseMarkdownInlineUnderscoreItalic(t *testing.T) {
	spans := parseMarkdownInline("say _hello_")
	if len(spans) != 2 || !spans[1].Italic || spans[1].Text != "hello" {
		t.Fatalf("italic span: %+v", spans)
	}
}

func TestParseMarkdownInlineBoldWithCode(t *testing.T) {
	spans := parseMarkdownInline("**Bold with `inline code` inside**")
	foundBold := false
	foundCode := false
	for _, sp := range spans {
		if sp.Bold && strings.Contains(sp.Text, "Bold") {
			foundBold = true
		}
		if sp.Variant == "code" && sp.Text == "inline code" {
			foundCode = true
		}
	}
	if !foundBold || !foundCode {
		t.Fatalf("expected bold + code spans, got %+v", spans)
	}
}

func TestMarkdownHorizontalRuleUnderscores(t *testing.T) {
	spec := MarkdownToDocumentSpec("t", "T", "___")
	if len(spec.Children) != 1 || spec.Children[0].Type != "divider" {
		t.Fatalf("want divider for ___, got %+v", spec.Children)
	}
}

func TestMarkdownNestedListNumbers(t *testing.T) {
	src := "1. First\n2. Second\n   1. Nested\n3. Third"
	spec := MarkdownToDocumentSpec("t", "T", src)
	if len(spec.Children) != 1 || !spec.Children[0].Ordered {
		t.Fatalf("expected one ordered list, got %+v", spec.Children)
	}
	nums := spec.Children[0].Props["numbers"].([]int)
	if len(nums) < 3 || nums[2] != 1 {
		t.Fatalf("nested ordered number = %v, want third item numbered 1 at depth", nums)
	}
}

func TestMarkdownEscapeAsterisk(t *testing.T) {
	spans := parseMarkdownInline(`\*not italic\*`)
	if len(spans) != 1 || spans[0].Text != "*not italic*" || spans[0].Italic {
		t.Fatalf("escaped asterisks: %+v", spans)
	}
}

func TestMarkdownTableDataTable(t *testing.T) {
	src := "| Name | Value |\n| --- | --- |\n| Alpha | 1 |\n| Beta | 2 |"
	spec := MarkdownToDocumentSpec("t", "T", src)
	if len(spec.Children) != 1 || spec.Children[0].Type != "table" {
		t.Fatalf("expected table block, got %+v", spec.Children)
	}
	cols := spec.Children[0].Props["columns"].([]any)
	if len(cols) != 2 {
		t.Fatalf("columns = %v", cols)
	}
	rows := spec.Children[0].Props["rows"].([]any)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
}

func TestMarkdownHeadingAnchorID(t *testing.T) {
	spec := MarkdownToDocumentSpec("t", "T", "## Hello **World**")
	if len(spec.Children) != 1 {
		t.Fatalf("got %d blocks", len(spec.Children))
	}
	if spec.Children[0].ID != "hello-world" {
		t.Fatalf("anchor id = %q, want hello-world", spec.Children[0].ID)
	}
}

func TestMarkdownBlockquoteNested(t *testing.T) {
	src := "> Outer line\n>\n> > Inner line\n>\n> > Second inner"
	spec := MarkdownToDocumentSpec("t", "T", src)
	if len(spec.Children) != 1 || spec.Children[0].Type != "callout" {
		t.Fatalf("want one callout, got %+v", spec.Children)
	}
	if len(spec.Children[0].Children) != 0 {
		t.Fatalf("blockquote should be one callout body, got children %+v", spec.Children[0].Children)
	}
	body := plainInlineText(spec.Children[0].Spans)
	if !strings.Contains(body, "Outer line") || !strings.Contains(body, "Inner line") {
		t.Fatalf("nested quote text missing: %q", body)
	}
}

func TestMarkdownUnorderedNestedList(t *testing.T) {
	src := "- Item 1\n- Item 2\n  - Nested A\n  - Nested B\n- Item 3"
	spec := MarkdownToDocumentSpec("t", "T", src)
	if len(spec.Children) != 1 || spec.Children[0].Type != "list" {
		t.Fatalf("expected one list, got %+v", spec.Children)
	}
	depths := spec.Children[0].Props["depths"].([]int)
	if len(depths) < 3 || depths[2] != 1 {
		t.Fatalf("depths = %v, want nested depth 1 at item 3", depths)
	}
}

func TestMarkdownListItemInline(t *testing.T) {
	spec := MarkdownToDocumentSpec("t", "T", "- item **bold**")
	if len(spec.Children) != 1 || spec.Children[0].Type != "list" {
		t.Fatalf("expected list block, got %+v", spec.Children[0])
	}
	node, err := buildDocBlock(spec.Children[0], NewBuildContext())
	if err != nil {
		t.Fatal(err)
	}
	c, ok := node.(*Container)
	if !ok {
		t.Fatalf("expected list container, got %T", node)
	}
	var rt *RichText
	var findRT func([]Node)
	findRT = func(nodes []Node) {
		for _, ch := range nodes {
			if r, ok := ch.(*RichText); ok {
				rt = r
				return
			}
			findRT(ch.Children())
			if rt != nil {
				return
			}
		}
	}
	findRT(c.Children())
	if rt == nil {
		t.Fatalf("expected RichText list item among %d children", len(c.Children()))
	}
	foundBold := false
	for _, sp := range rt.Spans {
		if sp.Bold && sp.Text == "bold" {
			foundBold = true
		}
	}
	if !foundBold {
		t.Fatalf("list item missing bold span: %+v", rt.Spans)
	}
}
