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
		if !entering {
			return ast.WalkContinue, nil
		}
		fn, ok := n.(*extast.Footnote)
		if !ok {
			return ast.WalkContinue, nil
		}
		ref := string(fn.Ref)
		if ref == "" {
			return ast.WalkContinue, nil
		}
		out[fn.Index] = ref
		return ast.WalkContinue, nil
	})
	return out
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
