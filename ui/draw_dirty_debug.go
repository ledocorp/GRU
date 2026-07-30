// Package ui — draw-dirty tree diagnostics for idle FPS investigation.
package ui

import (
	"fmt"
	"strings"
)

// DirtyNodeReport is one node that still has layout and/or draw dirty set.
type DirtyNodeReport struct {
	ID         string
	Type       string
	LayoutDirty bool
	DrawDirty   bool
	Path       string
}

const dirtyReportMax = 8

// CollectDirtyReports walks the tree and returns nodes with pending dirty flags.
// limit caps results (0 = dirtyReportMax).
func CollectDirtyReports(root Node, limit int) []DirtyNodeReport {
	if limit <= 0 {
		limit = dirtyReportMax
	}
	var out []DirtyNodeReport
	collectDirtyReports(root, "", &out, limit)
	return out
}

func collectDirtyReports(n Node, path string, out *[]DirtyNodeReport, limit int) {
	if n == nil || len(*out) >= limit || n.IsHidden() {
		return
	}
	id := n.ID()
	if id == "" {
		id = "?"
	}
	cur := path
	if cur == "" {
		cur = id
	} else {
		cur = path + "/" + id
	}
	layoutDirty := n.IsDirty()
	drawDirty := false
	if d, ok := n.(interface{ DbgDrawDirty() bool }); ok {
		drawDirty = d.DbgDrawDirty()
	}
	if layoutDirty || drawDirty {
		*out = append(*out, DirtyNodeReport{
			ID:          id,
			Type:        fmt.Sprintf("%T", n),
			LayoutDirty: layoutDirty,
			DrawDirty:   drawDirty,
			Path:        cur,
		})
	}
	for _, ch := range n.Children() {
		collectDirtyReports(ch, cur, out, limit)
	}
}

// FormatDirtyReports renders a compact single-line summary for stderr logging.
func FormatDirtyReports(reports []DirtyNodeReport) string {
	if len(reports) == 0 {
		return "none"
	}
	parts := make([]string, 0, len(reports))
	for _, r := range reports {
		flags := make([]string, 0, 2)
		if r.LayoutDirty {
			flags = append(flags, "layout")
		}
		if r.DrawDirty {
			flags = append(flags, "draw")
		}
		parts = append(parts, fmt.Sprintf("%s[%s](%s)", r.ID, strings.Join(flags, "+"), shortTypeName(r.Type)))
	}
	return strings.Join(parts, ", ")
}

func shortTypeName(full string) string {
	if i := strings.LastIndex(full, "."); i >= 0 {
		return full[i+1:]
	}
	return full
}
