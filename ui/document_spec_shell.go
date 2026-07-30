// Package ui (continued) — DocumentSpec shell widgets (Sprint G2).
package ui

import (
	"fmt"
	"strings"
)

// docDataTableRow is a JSON-friendly table row (field key → cell text).
type docDataTableRow map[string]string

// buildDocSearchBar compiles a SearchBar block (see batch2_demo.go).
func buildDocSearchBar(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	id := block.ID
	if id == "" {
		id = "search"
	}
	placeholder := block.Placeholder
	if placeholder == "" {
		placeholder = block.Text
	}
	if placeholder == "" {
		placeholder = "Search…"
	}
	h := block.Height
	if h <= 0 {
		h = 38
	}
	w := docBlockWidth(block)
	sb := NewSearchBar(id, placeholder, 0, 0, w, h)
	if block.Props != nil {
		if v, ok := block.Props["debounceDelay"].(float64); ok && v >= 0 {
			sb.DebounceDelay = float32(v)
		}
	}
	registerDocControlValue(ctx, id, func() any { return sb.Query.Get() })
	applyDocStyle(&sb.Element, block)
	applyDocLayout(&sb.Element, block)
	return sb, nil
}

// buildDocTabView compiles a TabView from child tab blocks.
func buildDocTabView(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	id := block.ID
	if id == "" {
		id = "tabs"
	}
	if len(block.Children) == 0 {
		return nil, docBlockError(block, path, "tabView requires tab children")
	}
	w := docBlockWidth(block)
	h := block.Height
	if h <= 0 {
		h = 320
	}
	tv := NewTabView(id, 0, 0, w, h)
	for i, tab := range block.Children {
		tabPath := fmt.Sprintf("%s.children[%d]", path, i)
		title := tab.Title
		if title == "" {
			title = tab.Label
		}
		if title == "" {
			title = fmt.Sprintf("Tab %d", i+1)
		}
		content := NewContainer(fmt.Sprintf("%s-tab-%d", id, i), 0, 0, w, h-tabBarH)
		content.SetStyle("transparent")
		content.FlexDirection = FlexColumn
		content.Gap = 8
		if err := addDocChildren(content, tab.Children, ctx, tabPath); err != nil {
			return nil, err
		}
		tv.AddTab(title, content)
	}
	idx := block.SelectedIndex
	if block.Value != nil && docBlockValueIsNumber(block.Value) {
		idx = int(docBlockNumericValue(block, 0))
	}
	if idx < 0 {
		idx = 0
	}
	if idx >= len(tv.tabs) {
		idx = len(tv.tabs) - 1
	}
	if len(tv.tabs) > 0 {
		tv.Active.Set(idx)
	}
	applyDocStyle(&tv.Element, block)
	applyDocLayout(&tv.Element, block)
	return tv, nil
}

func buildDocDataTable(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	id := block.ID
	if id == "" {
		id = "datatable"
	}
	cols, err := parseDocTableColumns(block, path)
	if err != nil {
		return nil, err
	}
	rows, err := parseDocTableRows(block, cols, path)
	if err != nil {
		return nil, err
	}
	w := docBlockWidth(block)
	h := block.Height
	if h <= 0 {
		h = 240
	}
	binding := NewListBinding(rows)
	uiCols := make([]Column[docDataTableRow], len(cols))
	for i, c := range cols {
		key := c.key
		col := Column[docDataTableRow]{
			Title:  c.title,
			Width:  c.width,
			Align:  c.align,
			Render: func(item docDataTableRow) string { return item[key] },
		}
		if c.sortable {
			col.Sortable = true
			col.SortLess = func(a, b docDataTableRow) bool {
				return strings.ToLower(a[key]) < strings.ToLower(b[key])
			}
		}
		uiCols[i] = col
	}
	dt := NewDataTable(id, uiCols, binding, 0, 0, w, h)
	if rh := docBlockFloat(block, "rowHeight", 0); rh > 0 {
		dt.RowHeight = rh
	}
	registerDocControlValue(ctx, id, func() any { return binding.GetSelectedIndex() })
	applyDocStyle(&dt.Element, block)
	applyDocLayout(&dt.Element, block)
	return dt, nil
}

func parseDocTableColumns(block DocBlock, path string) ([]docTableColumnSpec, error) {
	raw := docBlockPropsValue(block, "columns")
	if raw == nil {
		return nil, docBlockError(block, path, "dataTable requires props.columns")
	}
	switch arr := raw.(type) {
	case []any:
		out := make([]docTableColumnSpec, 0, len(arr))
		for i, el := range arr {
			switch c := el.(type) {
			case string:
				key := docTableFieldKey(c, i)
				out = append(out, docTableColumnSpec{
					title: c,
					width: 120,
					key:   key,
				})
			case map[string]any:
				spec, err := docTableColumnFromMap(c, i)
				if err != nil {
					return nil, docBlockError(block, path, "columns[%d]: %v", i, err)
				}
				out = append(out, spec)
			default:
				return nil, docBlockError(block, path, "columns[%d]: expected string or object", i)
			}
		}
		if len(out) == 0 {
			return nil, docBlockError(block, path, "columns must not be empty")
		}
		return out, nil
	default:
		return nil, docBlockError(block, path, "props.columns must be an array")
	}
}

func docTableColumnFromMap(m map[string]any, index int) (docTableColumnSpec, error) {
	title, _ := m["title"].(string)
	if title == "" {
		title, _ = m["label"].(string)
	}
	if title == "" {
		title = fmt.Sprintf("Col %d", index+1)
	}
	key, _ := m["key"].(string)
	if key == "" {
		key, _ = m["field"].(string)
	}
	if key == "" {
		key = docTableFieldKey(title, index)
	}
	width := float32(120)
	if w, ok := m["width"].(float64); ok && w > 0 {
		width = float32(w)
	}
	align := ColumnAlignLeft
	if a, ok := m["align"].(string); ok {
		switch strings.ToLower(a) {
		case "center", "centre":
			align = ColumnAlignCenter
		case "right":
			align = ColumnAlignRight
		}
	}
	sortable, _ := m["sortable"].(bool)
	return docTableColumnSpec{title: title, width: width, key: key, align: align, sortable: sortable}, nil
}

func docTableFieldKey(title string, index int) string {
	k := strings.ToLower(strings.TrimSpace(title))
	k = strings.ReplaceAll(k, " ", "_")
	if k == "" {
		return fmt.Sprintf("col%d", index)
	}
	return k
}

// DocTableFieldKey normalizes a table column title into a row map key.
// Exported for the optional ui/markdown goldmark bridge.
func DocTableFieldKey(title string, index int) string {
	return docTableFieldKey(title, index)
}

func parseDocTableRows(block DocBlock, cols []docTableColumnSpec, path string) ([]docDataTableRow, error) {
	raw := docBlockPropsValue(block, "rows")
	if raw == nil {
		return []docDataTableRow{}, nil
	}
	arr, ok := raw.([]any)
	if !ok {
		return nil, docBlockError(block, path, "props.rows must be an array")
	}
	out := make([]docDataTableRow, 0, len(arr))
	for i, el := range arr {
		switch row := el.(type) {
		case map[string]any:
			m := make(docDataTableRow, len(cols))
			for _, c := range cols {
				if v, ok := row[c.key]; ok {
					m[c.key] = fmt.Sprint(v)
				} else if v, ok := row[c.title]; ok {
					m[c.key] = fmt.Sprint(v)
				}
			}
			out = append(out, m)
		case []any:
			m := make(docDataTableRow, len(cols))
			for j, c := range cols {
				if j < len(row) {
					m[c.key] = fmt.Sprint(row[j])
				}
			}
			out = append(out, m)
		default:
			return nil, docBlockError(block, path, "rows[%d]: expected object or array", i)
		}
	}
	return out, nil
}

func docBlockPropsValue(block DocBlock, key string) any {
	if block.Props == nil {
		return nil
	}
	return block.Props[key]
}

type docTableColumnSpec struct {
	title    string
	width    float32
	key      string
	align    ColumnAlign
	sortable bool
}

// buildDocTableCard renders markdown tables: rounded card, shaded header row, body rows
// with flex columns so cell text does not collide.
func buildDocTableCard(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	id := block.ID
	if id == "" {
		id = "table"
	}
	cols, err := parseDocTableColumns(block, path)
	if err != nil {
		return nil, err
	}
	rows, err := parseDocTableRows(block, cols, path)
	if err != nil {
		return nil, err
	}

	card := NewCard(id, "", 0, 0, docBlockWidthFromCtx(block, ctx), 0)
	applyDocStyle(&card.Element, block)
	configureDocMarkdownCard(card, "table", "")
	card.Gap = 0

	card.AddChild(docTableHeaderRow(id+"-hdr", cols, block, ctx))
	for ri, row := range rows {
		card.AddChild(docTableDataRow(fmt.Sprintf("%s-row-%d", id, ri), cols, row, block, ctx))
	}

	applyDocLayout(&card.Element, block)
	applyCardChromeTextToBody(card, block)
	if err := applyDocBlockCapabilities(block, &card.SurfaceShell, ctx, path); err != nil {
		return nil, err
	}
	return card, nil
}

func docTableHeaderRow(id string, cols []docTableColumnSpec, block DocBlock, ctx *BuildContext) Node {
	row := NewContainer(id, 0, 0, 0, 0)
	row.SetStyle("table-header-row")
	row.FlexDirection = FlexRow
	row.Gap = 12
	row.AutoHeight = true
	for i, c := range cols {
		rt := NewRichText(fmt.Sprintf("%s-c%d", id, i), []TextSpan{{Text: c.title, Bold: true}}, 0, 0, 0, 0)
		rt.SetFlexGrow(1)
		applyDocRichText(rt, block, ctx)
		row.AddChild(rt)
	}
	return row
}

func docTableDataRow(id string, cols []docTableColumnSpec, data docDataTableRow, block DocBlock, ctx *BuildContext) Node {
	row := NewContainer(id, 0, 0, 0, 0)
	row.SetStyle("table-body-row")
	row.FlexDirection = FlexRow
	row.Gap = 12
	row.AutoHeight = true
	for i, c := range cols {
		cell := data[c.key]
		rt := NewRichText(fmt.Sprintf("%s-c%d", id, i), parseMarkdownInline(cell), 0, 0, 0, 0)
		rt.SetFlexGrow(1)
		applyDocRichText(rt, block, ctx)
		row.AddChild(rt)
	}
	return row
}
