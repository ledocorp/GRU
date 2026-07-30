package preview

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/ledocorp/gru/ui"
)

// isGFMAlignRow reports GFM separator rows (e.g. |:---|:---:|---:|).
func isGFMAlignRow(cells []string) bool {
	if len(cells) == 0 {
		return false
	}
	n := 0
	for _, c := range cells {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		n++
		if !tableAlignCell(c) {
			return false
		}
	}
	return n > 0
}

func tableAlignCell(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return true
	}
	if len(s) >= 3 && strings.HasPrefix(s, ":") && strings.HasSuffix(s, ":") {
		for _, r := range s[1 : len(s)-1] {
			if r != '-' {
				return false
			}
		}
		return true
	}
	for _, r := range s {
		if r != '-' {
			return false
		}
	}
	return len(s) > 0
}

func tableRowCells(row ast.Node) []*extast.TableCell {
	var cells []*extast.TableCell
	for c := row.FirstChild(); c != nil; c = c.NextSibling() {
		if cell, ok := c.(*extast.TableCell); ok {
			cells = append(cells, cell)
		}
	}
	return cells
}

func tableCellPlain(cell *extast.TableCell, source []byte, ctx *mdBuildCtx) string {
	return strings.TrimSpace(plainInline(inlineSpans(cell, source, inlineFlags{}, ctx)))
}

type parsedTable struct {
	colKeys     []string
	headerSpans [][]ui.TextSpan
	bodySpans   [][][]ui.TextSpan
}

func parseTable(table *extast.Table, ctx *mdBuildCtx) parsedTable {
	var out parsedTable
	for child := table.FirstChild(); child != nil; child = child.NextSibling() {
		switch row := child.(type) {
		case *extast.TableHeader:
			cells := tableRowCells(row)
			for i, cell := range cells {
				out.colKeys = append(out.colKeys, tableColKey(tableCellPlain(cell, ctx.source, ctx), i))
				sp := inlineSpans(cell, ctx.source, inlineFlags{}, ctx)
				for j := range sp {
					sp[j].Bold = true
				}
				out.headerSpans = append(out.headerSpans, sp)
			}
		case *extast.TableRow:
			cells := tableRowCells(row)
			plain := make([]string, len(cells))
			for i, cell := range cells {
				plain[i] = tableCellPlain(cell, ctx.source, ctx)
			}
			if isGFMAlignRow(plain) {
				continue
			}
			if len(out.colKeys) == 0 && len(out.headerSpans) == 0 {
				for i, cell := range cells {
					out.colKeys = append(out.colKeys, tableColKey(plain[i], i))
					sp := inlineSpans(cell, ctx.source, inlineFlags{}, ctx)
					for j := range sp {
						sp[j].Bold = true
					}
					out.headerSpans = append(out.headerSpans, sp)
				}
				continue
			}
			var rowSpans [][]ui.TextSpan
			for _, cell := range cells {
				rowSpans = append(rowSpans, inlineSpans(cell, ctx.source, inlineFlags{}, ctx))
			}
			out.bodySpans = append(out.bodySpans, rowSpans)
		}
	}
	return out
}
