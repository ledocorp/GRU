package syntax

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestHighlightSyntaxGo(t *testing.T) {
	spans := ui.HighlightSyntax("package main\n\nfunc main() {}\n", "go")
	if len(spans) == 0 {
		t.Fatal("no spans")
	}
	colors := 0
	for _, sp := range spans {
		if sp.Color.A != 0 {
			colors++
		}
	}
	if colors < 2 {
		t.Fatalf("expected colored tokens, got %d colored of %d spans", colors, len(spans))
	}
}

func TestNormalizeSyntaxLanguage(t *testing.T) {
	if got := ui.NormalizeSyntaxLanguage("golang"); got != "go" {
		t.Fatalf("golang = %q", got)
	}
}
