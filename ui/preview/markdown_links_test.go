package preview

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestInternalLinkAnchorRegistration(t *testing.T) {
	src := "### 1. Headings\n\nParagraph with [jump](#1-headings).\n"
	anchors := make(map[string]ui.Node)
	nodes := BuildMarkdownNodesWithHandler("t", src, anchors, nil)
	if len(nodes) < 2 {
		t.Fatalf("nodes = %d, want at least 2", len(nodes))
	}
	target, ok := anchors["1-headings"]
	if !ok || target == nil {
		t.Fatalf("anchors[%q] missing; keys=%v", "1-headings", anchorKeys(anchors))
	}
	lane := ui.NewContainer("lane", 0, 0, 400, 600)
	for _, n := range nodes {
		lane.AddChild(n)
	}
	if got := resolvePreviewJumpTarget(lane, anchors, "1-headings"); got != target {
		t.Fatalf("resolvePreviewJumpTarget = %v, want %v", got, target)
	}
}

func anchorKeys(m map[string]ui.Node) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
