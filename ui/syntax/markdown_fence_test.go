package syntax

import (
	"strings"
	"testing"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestHighlightMarkdownSQLFence(t *testing.T) {
	ui.SetChromaStyle("pygments")
	src := "# Title\n\nSome prose here.\n\n```sql\nSELECT * FROM users;\n```\n\nMore prose."
	spans := ui.HighlightSyntax(src, "markdown")
	joined := spanText(spans)
	if joined != src {
		t.Fatalf("span text mismatch:\n%q\nvs\n%q", joined, src)
	}
	sqlKeyword := rl.NewColor(0, 128, 0, 255)
	for _, sp := range spans {
		if sp.Text == "Some" || sp.Text == "More" {
			if sp.Color == sqlKeyword {
				t.Fatalf("prose highlighted as SQL keyword: %q", sp.Text)
			}
		}
	}
	if !spanContainsColor(spans, "SELECT", sqlKeyword) {
		t.Fatal("expected SELECT highlighted inside sql fence")
	}
}

func TestHighlightMarkdownPythonAfterProseSection(t *testing.T) {
	ui.SetChromaStyle("github")
	// Mimics section 7 (live markdown prose) then section 8 python fence.
	src := "### 7. Markdown\n\n# Project Title\n\n## Overview\nSample doc.\n\n---\n\n**Last updated**: June 2026\n\n### 8. Python\n\n```python\ndef greet(name: str) -> str:\n    \"\"\"Return a greeting.\"\"\"\n    return f\"Hello, {name}!\"\n\nif __name__ == \"__main__\":\n    print(greet(\"Pythonista\"))\n```\n\n### 9. Go\n\n```go\npackage main\n\nfunc main() {}\n```\n"
	spans := ui.HighlightSyntax(src, "markdown")
	if spanText(spans) != src {
		t.Fatal("span text mismatch")
	}
	defSpan := findSpanContaining(spans, "def")
	if defSpan == nil || defSpan.Color.A == 0 {
		t.Fatal("python def should be colored inside fence")
	}
	goSpan := findSpanContaining(spans, "package")
	if goSpan == nil || goSpan.Color.A == 0 {
		t.Fatal("go fence should still colorize")
	}
}

func TestHighlightMarkdownPythonAfterBareFenceLine(t *testing.T) {
	ui.SetChromaStyle("github")
	// A lone ``` line in prose (common in markdown tutorials) must not swallow the next fence.
	src := "### Markdown tips\n\nEnd a block with:\n\n```\n\n### Python\n\n```python\nprint(\"hi\")\n```\n"
	spans := ui.HighlightSyntax(src, "markdown")
	printSpan := findSpanContaining(spans, "print")
	if printSpan == nil {
		t.Fatal("missing print span")
	}
	if printSpan.Color.A == 0 {
		t.Fatalf("python not highlighted after bare ``` line; got plain %q", printSpan.Text)
	}
}

func TestHighlightMarkdownUnclosedFence(t *testing.T) {
	ui.SetChromaStyle("pygments")
	src := "# Doc\n\n```sql\nSELECT 1;\n\n# still in fence"
	spans := ui.HighlightSyntax(src, "markdown")
	for _, sp := range spans {
		if strings.Contains(sp.Text, "Doc") && sp.Color.G == 128 && sp.Color.R == 0 && sp.Color.B == 0 {
			t.Fatalf("heading outside fence colored as SQL: %q", sp.Text)
		}
	}
}

func TestMarkdownFenceLangAliases(t *testing.T) {
	tests := map[string]string{
		"```python": "python",
		"```py":     "python",
		"```Python": "python",
		"```go":     "go",
	}
	for line, want := range tests {
		if got := markdownFenceLang(line); got != want {
			t.Fatalf("%q lang = %q want %q", line, got, want)
		}
	}
}

func spanText(spans []ui.TextSpan) string {
	var b strings.Builder
	for _, sp := range spans {
		b.WriteString(sp.Text)
	}
	return b.String()
}

func spanContainsColor(spans []ui.TextSpan, substr string, want rl.Color) bool {
	for _, sp := range spans {
		if strings.Contains(sp.Text, substr) && sp.Color == want {
			return true
		}
	}
	return false
}

func findSpanContaining(spans []ui.TextSpan, substr string) *ui.TextSpan {
	for i := range spans {
		if strings.Contains(spans[i].Text, substr) {
			return &spans[i]
		}
	}
	return nil
}
