package preview

import (
	"github.com/yuin/goldmark/ast"
)

// markdownSection is a group of top-level blocks (usually headed by H1/H2).
type markdownSection struct {
	blocks []ast.Node
}

// partitionMarkdownBlocks splits top-level blocks into sections at H1/H2 boundaries.
// FootnoteList and other blocks stay in document order inside their section.
func partitionMarkdownBlocks(blocks []ast.Node) []markdownSection {
	var sections []markdownSection
	var current []ast.Node
	flush := func() {
		if len(current) == 0 {
			return
		}
		sections = append(sections, markdownSection{blocks: current})
		current = nil
	}
	for _, b := range blocks {
		if h, ok := b.(*ast.Heading); ok && h.Level <= 2 && len(current) > 0 {
			flush()
		}
		current = append(current, b)
	}
	flush()
	return sections
}
