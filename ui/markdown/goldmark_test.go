package markdown

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestMarkdownToDocumentSpecHeading(t *testing.T) {
	spec := ui.MarkdownToDocumentSpec("t", "T", "## Hello **world**")
	if len(spec.Children) != 1 {
		t.Fatalf("got %d blocks", len(spec.Children))
	}
	spans := spec.Children[0].Spans
	if len(spans) < 2 || !spans[1].Bold || spans[1].Text != "world" {
		t.Fatalf("heading inline bold: %+v", spans)
	}
	if spec.Children[0].ID != "hello-world" {
		t.Fatalf("anchor id = %q, want hello-world", spec.Children[0].ID)
	}
}
