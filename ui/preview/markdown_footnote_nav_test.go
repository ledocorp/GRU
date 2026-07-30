package preview

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

func buildPreviewLane(t *testing.T, src string) ui.Node {
	t.Helper()
	nodes := BuildMarkdownNodes("t", src)
	lane := ui.NewContainer("lane", 0, 0, 0, 0)
	for _, n := range nodes {
		lane.AddChild(n)
	}
	return lane
}

func TestFindFirstFootnoteRef(t *testing.T) {
	src := "First[^1] here.\n\nSecond[^1] later.\n\n[^1]: Note one.\n\n[^2]: Note two."
	lane := buildPreviewLane(t, src)

	all := collectInlineFootnoteRefs(lane, "1")
	if len(all) < 2 {
		t.Fatalf("inline [^1] refs = %d, want 2", len(all))
	}
	first := findFirstFootnoteRef(lane, "1")
	if first == nil {
		t.Fatal("missing first [^1] ref")
	}
	if first != all[0] {
		t.Fatalf("findFirst=%s all[0]=%s", first.(*ui.RichText).ID(), all[0].(*ui.RichText).ID())
	}
}

func collectInlineFootnoteRefs(root ui.Node, ref string) []ui.Node {
	wantLink := "#" + footnoteDefAnchor(ref)
	var out []ui.Node
	walkPreviewNodes(root, func(n ui.Node) bool {
		rt, ok := n.(*ui.RichText)
		if !ok {
			return false
		}
		for _, sp := range rt.Spans {
			if sp.Variant == "footnote-ref" && sp.Link == wantLink {
				out = append(out, rt)
				break
			}
		}
		return false
	})
	return out
}

func TestFindFootnoteDef(t *testing.T) {
	src := "Text[^1] and [^2].\n\n[^1]: Foot one.\n\n[^2]: Foot two."
	lane := buildPreviewLane(t, src)

	def1 := findFootnoteDef(lane, "1")
	if def1 == nil {
		t.Fatal("missing fn-1 definition")
	}
	def2 := findFootnoteDef(lane, "2")
	if def2 == nil {
		t.Fatal("missing fn-2 definition")
	}
	if def1 == def2 {
		t.Fatal("definitions must differ")
	}
	if findFirstFootnoteRef(lane, "1") == def1 {
		t.Fatal("return target must not be the footnote definition")
	}
}

func TestResolvePreviewJumpTarget(t *testing.T) {
	src := "See[^1] here.\n\n[^1]: Note."
	lane := buildPreviewLane(t, src)

	if n := resolvePreviewJumpTarget(lane, nil, "fn-1"); n == nil {
		t.Fatal("fn-1 jump target missing")
	}
	if n := resolvePreviewJumpTarget(lane, nil, "fn-ref-1"); n == nil {
		t.Fatal("fn-ref-1 jump target missing")
	}
	if _, ok := resolvePreviewJumpTarget(lane, nil, "fn-ref-1").(*ui.RichText); !ok {
		t.Fatalf("return jump = %T, want inline RichText", resolvePreviewJumpTarget(lane, nil, "fn-ref-1"))
	}
}
