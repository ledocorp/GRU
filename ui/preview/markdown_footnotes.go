package preview

import (
	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"
)

// collectFootnoteIndexToRef maps goldmark footnote indices to label refs ([^1] → "1").
func collectFootnoteIndexToRef(doc ast.Node) map[int]string {
	out := make(map[int]string)
	if doc == nil {
		return out
	}
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		collectFootnoteIndexNode(n, entering, out)
		return ast.WalkContinue, nil
	})
	return out
}

// collectFootnoteIndexFromBlocks walks parsed top-level blocks only (avoids a second full parse).
func collectFootnoteIndexFromBlocks(blocks []ast.Node) map[int]string {
	out := make(map[int]string)
	for _, b := range blocks {
		_ = ast.Walk(b, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
			collectFootnoteIndexNode(n, entering, out)
			return ast.WalkContinue, nil
		})
	}
	return out
}

func collectFootnoteIndexNode(n ast.Node, entering bool, out map[int]string) {
	if !entering {
		return
	}
	fn, ok := n.(*extast.Footnote)
	if !ok {
		return
	}
	ref := string(fn.Ref)
	if ref == "" {
		return
	}
	out[fn.Index] = ref
}

// reorderFootnotesLast moves FootnoteList blocks to the document tail so preview
// definitions render at the bottom (Goldmark often places them right after the ref).
func reorderFootnotesLast(blocks []ast.Node) []ast.Node {
	if len(blocks) == 0 {
		return blocks
	}
	var lists []ast.Node
	out := make([]ast.Node, 0, len(blocks))
	for _, b := range blocks {
		switch b.(type) {
		case *extast.FootnoteList:
			lists = append(lists, b)
		case *extast.Footnote:
			// Loose defs duplicate FootnoteList; blockToNode skips these too.
		default:
			out = append(out, b)
		}
	}
	if len(lists) == 0 {
		return blocks
	}
	return append(out, lists...)
}

func footnoteDefAnchor(ref string) string {
	return "fn-" + markdownAnchorSlug(ref)
}

func footnoteBracketLabel(ref string) string {
	return "[" + ref + "]"
}
