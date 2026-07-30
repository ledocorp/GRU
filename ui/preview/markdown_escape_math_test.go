package preview

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestUnescapeMarkdownPunct(t *testing.T) {
	got := unescapeMarkdownPunct(`\*not italic\*`)
	if got != `*not italic*` {
		t.Fatalf("got %q", got)
	}
}

func TestEscapeListItemRendersLiteralAsterisks(t *testing.T) {
	nodes := BuildMarkdownNodes("e", "- \\*not italic\\*\n")
	if len(nodes) != 1 {
		t.Fatalf("nodes=%d", len(nodes))
	}
	var text string
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		if rt, ok := n.(*ui.RichText); ok {
			for _, s := range rt.Spans {
				text += s.Text
			}
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	walk(nodes[0])
	if strings.Contains(text, `\*`) {
		t.Fatalf("backslash left in text: %q", text)
	}
	if !strings.Contains(text, `*not italic*`) {
		t.Fatalf("text=%q", text)
	}
}

func TestDisplayMathBecomesLatexCard(t *testing.T) {
	src := "$$\n\\int x\n$$\n"
	nodes := BuildMarkdownNodes("m", src)
	if len(nodes) != 1 {
		t.Fatalf("nodes=%d", len(nodes))
	}
	card, ok := nodes[0].(*ui.Card)
	if !ok || card.Title != "LaTeX" {
		t.Fatalf("got %T title=%v", nodes[0], func() string {
			if c, ok := nodes[0].(*ui.Card); ok {
				return c.Title
			}
			return ""
		}())
	}
}

func TestShowcaseTailMathEscapeHTML(t *testing.T) {
	paths := []string{
		"docs/fixtures/markdown_feature_showcase.md",
		filepath.Join("..", "..", "docs", "fixtures", "markdown_feature_showcase.md"),
	}
	var data []byte
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			data = b
			break
		}
	}
	if data == nil {
		t.Skip("fixture missing")
	}
	idx := strings.Index(string(data), "### 10. Math")
	if idx < 0 {
		t.Fatal("math section missing")
	}
	nodes := BuildMarkdownNodes("show", string(data)[idx:])
	var sawLatexCard, sawEscapeOK, sawHTMLBold bool
	var walk func(ui.Node)
	walk = func(n ui.Node) {
		switch x := n.(type) {
		case *ui.Card:
			if x.Title == "LaTeX" {
				sawLatexCard = true
			}
		case *ui.RichText:
			var txt string
			for _, s := range x.Spans {
				txt += s.Text
				if s.Bold && strings.Contains(s.Text, "HTML") {
					sawHTMLBold = true
				}
			}
			if strings.Contains(txt, "*not italic*") && !strings.Contains(txt, `\*`) {
				sawEscapeOK = true
			}
		}
		for _, c := range n.Children() {
			walk(c)
		}
	}
	for _, n := range nodes {
		walk(n)
	}
	if !sawLatexCard {
		t.Fatal("expected LaTeX card for display/fenced math")
	}
	if !sawEscapeOK {
		t.Fatal("expected unescaped *not italic*")
	}
	if !sawHTMLBold {
		t.Fatal("expected bold HTML span")
	}
	var sawHTMLCard bool
	var walk2 func(ui.Node)
	walk2 = func(n ui.Node) {
		if c, ok := n.(*ui.Card); ok && c.Title == "HTML" {
			sawHTMLCard = true
		}
		for _, ch := range n.Children() {
			walk2(ch)
		}
	}
	for _, n := range nodes {
		walk2(n)
	}
	if !sawHTMLCard {
		t.Fatal("expected HTML card wrapper")
	}
}
