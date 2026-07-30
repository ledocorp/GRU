package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestCodeFenceSyntaxHighlight(t *testing.T) {
	src := "```go\npackage main\n\nfunc main() {}\n```"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T", nodes[0])
	}
	rt := firstRichText(card)
	if rt == nil {
		t.Fatal("missing code RichText")
	}
	colored := distinctSpanColors(rt.Spans)
	if len(colored) < 2 {
		t.Fatalf("expected multiple token colors, got %d: %+v", len(colored), rt.Spans)
	}
}

func TestHighlightSpansDistinctColors(t *testing.T) {
	spans := highlightSpans("package main\n\nfunc main() {\n\tfmt.Println(\"hi\")\n}", "go")
	colors := distinctSpanColors(spans)
	if len(colors) < 3 {
		t.Fatalf("distinct colors = %d, want >= 3; spans=%d", len(colors), len(spans))
	}
}

func distinctSpanColors(spans []ui.TextSpan) map[[4]uint8]int {
	out := make(map[[4]uint8]int)
	for _, sp := range spans {
		if sp.Color.A == 0 {
			continue
		}
		key := [4]uint8{sp.Color.R, sp.Color.G, sp.Color.B, sp.Color.A}
		out[key]++
	}
	return out
}

func TestRenderLatexPNG(t *testing.T) {
	png, err := renderLatexPNG(`x = 42`, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(png) < 8 || string(png[:8]) != "\x89PNG\r\n\x1a\n" {
		t.Fatalf("not PNG header, len=%d", len(png))
	}
	png2, err := renderLatexPNG(`\int\frac{\partial x}{x}`, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(png2) < len(png)/2 {
		t.Fatalf("display PNG suspiciously small: %d vs inline %d", len(png2), len(png))
	}
	_, err = renderLatexPNG(`E=mc^2`, false)
	if err == nil {
		t.Fatal("expected error for unsupported superscript syntax")
	}
}

func TestMathFence(t *testing.T) {
	src := "```latex\nE = mc^2\n```"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	card, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T", nodes[0])
	}
	if card.Title != "LaTeX" {
		t.Fatalf("title = %q", card.Title)
	}
}

func TestInlineMath(t *testing.T) {
	src := "Energy is $E=mc^2$ in text."
	nodes := BuildMarkdownNodes("t", src)
	rt, ok := nodes[0].(*ui.RichText)
	if !ok {
		t.Fatalf("node = %T", nodes[0])
	}
	found := false
	for _, sp := range rt.Spans {
		if sp.Variant == "math" && strings.Contains(sp.Text, "E=mc^2") {
			found = true
		}
	}
	if !found {
		t.Fatalf("spans = %+v", rt.Spans)
	}
}

func TestMathInBlockquote(t *testing.T) {
	src := "> Quote with inline $a^2+b^2=c^2$.\n>\n> ```latex\n> \\frac{a}{b}\n> ```"
	nodes := BuildMarkdownNodes("t", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes = %d", len(nodes))
	}
	root, ok := nodes[0].(*ui.Card)
	if !ok {
		t.Fatalf("node = %T", nodes[0])
	}
	var mathInline, latexCard bool
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if rt, ok := n.(*ui.RichText); ok {
			for _, sp := range rt.Spans {
				if sp.Variant == "math" {
					mathInline = true
				}
			}
		}
		if c, ok := n.(*ui.Card); ok && c.Title == "LaTeX" {
			latexCard = true
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
	if !mathInline {
		t.Fatal("expected inline math inside blockquote")
	}
	if !latexCard {
		t.Fatal("expected latex fence inside blockquote")
	}
}

func TestShowcaseBuildsMixedSections(t *testing.T) {
	data, err := os.ReadFile("docs/fixtures/markdown_feature_showcase.md")
	if err != nil {
		data, err = os.ReadFile(filepath.Join("..", "docs", "fixtures", "markdown_feature_showcase.md"))
	}
	if err != nil {
		t.Skip("showcase fixture not found")
	}
	nodes := BuildMarkdownNodes("showcase", string(data))
	if len(nodes) < 12 {
		t.Fatalf("showcase nodes = %d, want rich fixture", len(nodes))
	}
}
