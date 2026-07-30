package ui

import (
	"fmt"
	"testing"
)

// TestDocumentSpecBlockRecipes audits buildDocBlockAt: each JSON type must compile
// to the same runtime widget types as the Go recipes in docs/DOCUMENT_SPEC_GO_RECIPES.md.
func TestDocumentSpecBlockRecipes(t *testing.T) {
	ctx := NewBuildContext()
	ctx.Actions["act"] = func() {}

	type check func(Node) error

	assert := func(name string, block DocBlock, fn check) {
		t.Helper()
		node, err := buildDocBlockAt(block, ctx, "children[0]")
		if err != nil {
			t.Fatalf("%s: build error: %v", name, err)
		}
		if err := fn(node); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
	}

	column := func(c *Container) error {
		if c.FlexDirection != FlexColumn {
			return errf("FlexDirection = %v, want FlexColumn", c.FlexDirection)
		}
		if c.styleName != "transparent" {
			return errf("style = %q, want transparent", c.styleName)
		}
		return nil
	}

	assert("column", DocBlock{Type: "column", ID: "col"}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok {
			return errf("type %T, want *Container", n)
		}
		return column(c)
	})

	assert("page", DocBlock{Type: "page", ID: "page"}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok {
			return errf("type %T, want *Container", n)
		}
		return column(c)
	})

	assert("row", DocBlock{Type: "row", ID: "row"}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok {
			return errf("type %T, want *Container", n)
		}
		if c.FlexDirection != FlexRow {
			return errf("FlexDirection = %v, want FlexRow", c.FlexDirection)
		}
		return nil
	})

	assert("buttonRow", DocBlock{Type: "buttonRow", ID: "br"}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok || c.FlexDirection != FlexRow {
			return errf("buttonRow: got %T", n)
		}
		return nil
	})

	assert("form", DocBlock{Type: "form", ID: "form"}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok {
			return errf("type %T", n)
		}
		if err := column(c); err != nil {
			return err
		}
		if c.GetStyle().Padding != 8 {
			return errf("padding = %.0f, want 8", c.GetStyle().Padding)
		}
		return nil
	})

	assert("field", DocBlock{Type: "field", ID: "fld", Label: "Name", Children: []DocBlock{
		{Type: "input", ID: "in"},
	}}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok {
			return errf("type %T", n)
		}
		if _, ok := c.Children()[0].(*RichText); !ok {
			return errf("first child %T, want *RichText caption", c.Children()[0])
		}
		if _, ok := c.Children()[1].(*TextInput); !ok {
			return errf("second child %T, want *TextInput", c.Children()[1])
		}
		return nil
	})

	assert("section", DocBlock{Type: "section", ID: "sec", Title: "Page"}, func(n Node) error {
		if _, ok := n.(*Panel); !ok {
			return errf("type %T, want *Panel", n)
		}
		return nil
	})

	assert("card", DocBlock{Type: "card", ID: "card", Title: "Group"}, func(n Node) error {
		if _, ok := n.(*Card); !ok {
			return errf("type %T, want *Card", n)
		}
		return nil
	})

	assert("callout", DocBlock{Type: "callout", ID: "co", Title: "Tip", Text: "Note"}, func(n Node) error {
		card, ok := n.(*Card)
		if !ok {
			return errf("type %T, want *Card", n)
		}
		want := CurrentThemeV2().Components["card"].Variants["callout"].BorderColor
		if card.GetStyle().BorderColor != want {
			return errf("callout theme mismatch")
		}
		if len(card.Children()) != 1 {
			return errf("children = %d, want 1 RichText", len(card.Children()))
		}
		if _, ok := card.Children()[0].(*RichText); !ok {
			return errf("child %T, want *RichText", card.Children()[0])
		}
		return nil
	})

	assert("code", DocBlock{Type: "code", ID: "cd", Text: "fn main() {}"}, func(n Node) error {
		card, ok := n.(*Card)
		if !ok {
			return errf("type %T, want *Card", n)
		}
		codeWant := CurrentThemeV2().Components["card"].Variants["code"].BorderColor
		if card.GetStyle().BorderColor != codeWant {
			return errf("code theme mismatch")
		}
		scroll, ok := card.Children()[0].(*Viewport)
		if !ok || scroll.Orientation != ScrollHorizontal {
			return errf("child %T, want horizontal *Viewport", card.Children()[0])
		}
		var rt *RichText
		var walk func(Node)
		walk = func(node Node) {
			if r, ok := node.(*RichText); ok {
				rt = r
			}
			for _, ch := range node.Children() {
				walk(ch)
			}
		}
		walk(scroll)
		if rt == nil {
			return errf("missing RichText in code scroll host")
		}
		if len(rt.Spans) == 0 || rt.Spans[0].Variant != "" {
			return errf("code block should not use inline-code variant, got %q", rt.Spans[0].Variant)
		}
		if !rt.GetStyle().Mono {
			return errf("expected mono code-block body")
		}
		return nil
	})

	assert("viewport", DocBlock{Type: "viewport", ID: "vp", Height: 200}, func(n Node) error {
		vp, ok := n.(*Viewport)
		if !ok {
			return errf("type %T, want *Viewport", n)
		}
		if vp.AutoHeight {
			return errf("fixed height viewport should not be AutoHeight")
		}
		return nil
	})

	assert("text", DocBlock{Type: "text", ID: "t", Text: "Hello"}, func(n Node) error {
		if _, ok := n.(*RichText); !ok {
			return errf("type %T, want *RichText", n)
		}
		return nil
	})

	assert("divider", DocBlock{Type: "divider", ID: "div"}, func(n Node) error {
		if _, ok := n.(*Separator); !ok {
			return errf("type %T, want *Separator", n)
		}
		return nil
	})

	assert("list", DocBlock{Type: "list", ID: "list", Items: []string{"a", "b"}}, func(n Node) error {
		c, ok := n.(*Container)
		if !ok || len(c.Children()) != 2 {
			return errf("list container children = %d", len(c.Children()))
		}
		for i, ch := range c.Children() {
			row, ok := ch.(*Container)
			if !ok {
				return errf("item %d type %T, want *Container row", i, ch)
			}
			found := false
			for _, rc := range row.Children() {
				if _, ok := rc.(*RichText); ok {
					found = true
					break
				}
			}
			if !found {
				return errf("item %d row has no RichText", i)
			}
		}
		return nil
	})

	assert("progressBar", DocBlock{Type: "progressBar", ID: "pb", Value: 50, Min: 0, Max: 100, Width: 200}, func(n Node) error {
		pb, ok := n.(*ProgressBar)
		if !ok {
			return errf("type %T, want *ProgressBar", n)
		}
		if got := pb.Value.Get(); got < 0.49 || got > 0.51 {
			return errf("value = %.2f, want ~0.5", got)
		}
		return nil
	})

	assert("badge", DocBlock{Type: "badge", ID: "chip", Text: "New", Variant: "success"}, func(n Node) error {
		b, ok := n.(*Badge)
		if !ok {
			return errf("type %T, want *Badge", n)
		}
		if b.Variant != BadgeSuccess {
			return errf("variant = %v, want success", b.Variant)
		}
		return nil
	})

	assert("button", DocBlock{Type: "button", ID: "btn", Text: "Go", OnClick: "act"}, func(n Node) error {
		btn, ok := n.(*Button)
		if !ok {
			return errf("type %T, want *Button", n)
		}
		if !btn.IsAutoHeight() {
			return errf("button should be AutoHeight")
		}
		if btn.OnClick == nil {
			return errf("OnClick not wired")
		}
		return nil
	})

	assert("input", DocBlock{Type: "input", ID: "in", Value: "x"}, func(n Node) error {
		if _, ok := n.(*TextInput); !ok {
			return errf("type %T, want *TextInput", n)
		}
		return nil
	})

	assert("dropdown", DocBlock{Type: "dropdown", ID: "dd", Options: []string{"A"}}, func(n Node) error {
		if _, ok := n.(*Dropdown); !ok {
			return errf("type %T, want *Dropdown", n)
		}
		return nil
	})

	assert("checkbox", DocBlock{Type: "checkbox", ID: "cb", Label: "Ok"}, func(n Node) error {
		wrap, ok := n.(*Container)
		if !ok {
			return errf("type %T, want label wrap *Container", n)
		}
		row, ok := wrap.Children()[0].(*Container)
		if !ok {
			return errf("wrap child %T", wrap.Children()[0])
		}
		if _, ok := row.Children()[1].(*Checkbox); !ok {
			return errf("checkbox child %T", row.Children()[1])
		}
		return nil
	})

	assert("toggle", DocBlock{Type: "toggle", ID: "tg", Label: "On"}, func(n Node) error {
		wrap, ok := n.(*Container)
		if !ok {
			return errf("type %T", n)
		}
		row, _ := wrap.Children()[0].(*Container)
		if _, ok := row.Children()[1].(*Toggle); !ok {
			return errf("toggle child %T", row.Children()[1])
		}
		return nil
	})

	assert("listTile", DocBlock{Type: "listTile", ID: "lt", Title: "Wi-Fi", Text: "On", Props: map[string]any{"trailing": "toggle"}, Checked: true}, func(n Node) error {
		lt, ok := n.(*ListTile)
		if !ok {
			return errf("type %T, want *ListTile", n)
		}
		if lt.Title != "Wi-Fi" {
			return errf("title = %q", lt.Title)
		}
		if _, ok := lt.Trailing.(*Toggle); !ok {
			return errf("trailing %T, want *Toggle", lt.Trailing)
		}
		return nil
	})

	assert("appBar", DocBlock{Type: "appBar", ID: "bar", Title: "Settings", Props: map[string]any{
		"leadingIcon": "<", "trailingIcon": "..",
	}}, func(n Node) error {
		bar, ok := n.(*AppBar)
		if !ok {
			return errf("type %T, want *AppBar", n)
		}
		if bar.Leading == nil || bar.Trailing == nil {
			return errf("expected leading and trailing actions")
		}
		return nil
	})

	chipSel := true
	assert("chip", DocBlock{Type: "chip", ID: "c", Text: "Go", Variant: "primary", Selectable: &chipSel, Checked: true}, func(n Node) error {
		b, ok := n.(*Badge)
		if !ok {
			return errf("type %T, want *Badge", n)
		}
		if b.Selected == nil || !b.Selected.Get() {
			return errf("expected selectable badge")
		}
		return nil
	})

	assert("bottomnav", DocBlock{Type: "bottomnav", ID: "nav", Items: []string{"Home", "Search"}, Props: map[string]any{
		"icons": []string{"H", "S"},
	}, SelectedIndex: 1}, func(n Node) error {
		bn, ok := n.(*BottomNavigationBar)
		if !ok {
			return errf("type %T", n)
		}
		if bn.Selected.Get() != 1 {
			return errf("selected = %d", bn.Selected.Get())
		}
		return nil
	})

	assert("fab", DocBlock{Type: "fab", ID: "fab", Text: "+", OnClick: "act"}, func(n Node) error {
		f, ok := n.(*FAB)
		if !ok {
			return errf("type %T", n)
		}
		if f.OnClick == nil {
			return errf("OnClick not wired")
		}
		return nil
	})

	assert("avatar", DocBlock{Type: "avatar", ID: "av", Text: "AB", Props: map[string]any{"statusOnline": true}}, func(n Node) error {
		a, ok := n.(*Avatar)
		if !ok {
			return errf("type %T", n)
		}
		if !a.ShowStatus || !a.StatusOnline {
			return errf("expected status dot")
		}
		return nil
	})

	assert("breadcrumbs", DocBlock{Type: "breadcrumbs", ID: "bc", Items: []string{"A", "B"}}, func(n Node) error {
		if _, ok := n.(*Breadcrumbs); !ok {
			return errf("type %T", n)
		}
		return nil
	})

	assert("combobox", DocBlock{Type: "combobox", ID: "cb", Text: "B", Options: []string{"A", "B", "C"}}, func(n Node) error {
		cb, ok := n.(*ComboBox)
		if !ok {
			return errf("type %T", n)
		}
		if cb.Selected.Get() != "B" {
			return errf("selected = %q", cb.Selected.Get())
		}
		return nil
	})

	assert("dateRangePicker", DocBlock{Type: "dateRangePicker", ID: "dr", Props: map[string]any{"start": "2026-01-01", "end": "2026-01-31"}}, func(n Node) error {
		dr, ok := n.(*DateRangePicker)
		if !ok {
			return errf("type %T", n)
		}
		if dr.Start.Get().Format("2006-01-02") != "2026-01-01" {
			return errf("start = %v", dr.Start.Get())
		}
		if dr.End.Get().Format("2006-01-02") != "2026-01-31" {
			return errf("end = %v", dr.End.Get())
		}
		return nil
	})

	assert("pagination", DocBlock{Type: "pagination", ID: "pg", Value: 2, Max: 5}, func(n Node) error {
		pg, ok := n.(*Pagination)
		if !ok {
			return errf("type %T", n)
		}
		if pg.Current.Get() != 1 {
			return errf("page = %d, want 1 (value 2 → index 1)", pg.Current.Get())
		}
		return nil
	})

	assert("rating", DocBlock{Type: "rating", ID: "r", Value: 3, Max: 5}, func(n Node) error {
		rt, ok := n.(*Rating)
		if !ok {
			return errf("type %T, want *Rating", n)
		}
		if got := int(rt.Value.Get() + 0.5); got != 3 {
			return errf("value = %d, want 3", got)
		}
		return nil
	})

	assert("radioGroup", DocBlock{Type: "radioGroup", ID: "rg", Label: "Pick", Options: []string{"A", "B"}}, func(n Node) error {
		col, ok := n.(*Container)
		if !ok {
			return errf("type %T", n)
		}
		if _, ok := col.Children()[1].(*RadioGroup); !ok {
			return errf("radio child %T", col.Children()[1])
		}
		return nil
	})

	ctx.BoolSignals["autosave"] = NewSignal(false)
	ctx.BoolSignals["gridLines"] = NewSignal(true)
	ctx.Float64Signals["zoom"] = NewSignal(100.0)
	assert("toolbar", DocBlock{
		Type: "toolbar",
		ID:   "state-bar",
		Props: map[string]any{
			"overflow":     true,
			"overflowKind": "scroll",
			"itemGap":      float64(10),
		},
		Children: []DocBlock{
			{
				Type:  "toolbarGroup",
				ID:    "state",
				Label: "State",
				Children: []DocBlock{
					{Type: "toggleLabel", ID: "autosave", Label: "Autosave", Props: map[string]any{"bind": "autosave"}},
					{Type: "separator", ID: "sep"},
					{Type: "menu", ID: "font", Label: "Font size", Options: []string{"10 pt", "12 pt"}},
					{Type: "wordToggle", ID: "grid", Label: "Gridlines", Props: map[string]any{"bind": "gridLines"}},
					{Type: "spinBox", ID: "zoom", Min: 50, Max: 200, Value: 100, Props: map[string]any{"bind": "zoom", "step": float64(5)}},
					{Type: "button", ID: "sync", Text: "Sync", OnClick: "act"},
				},
			},
		},
	}, func(n Node) error {
		tb, ok := n.(*Toolbar)
		if !ok {
			return errf("type %T, want *Toolbar", n)
		}
		if !tb.Overflow || tb.OverflowKind != ToolbarOverflowScroll {
			return errf("overflow scroll not set")
		}
		if tb.ItemGap != 10 {
			return errf("itemGap = %.0f, want 10", tb.ItemGap)
		}
		if len(tb.Groups) != 1 || tb.Groups[0].ID != "state" {
			return errf("groups = %d", len(tb.Groups))
		}
		if len(tb.Groups[0].items) < 5 {
			return errf("item count = %d, want >= 5", len(tb.Groups[0].items))
		}
		return nil
	})

	assert("searchBar", DocBlock{
		Type:        "searchBar",
		ID:          "find",
		Placeholder: "Search fruits…",
		Height:      38,
		Props: map[string]any{
			"debounceDelay": float64(0.3),
		},
	}, func(n Node) error {
		sb, ok := n.(*SearchBar)
		if !ok {
			return errf("type %T, want *SearchBar", n)
		}
		if sb.Placeholder != "Search fruits…" {
			return errf("placeholder = %q", sb.Placeholder)
		}
		if sb.DebounceDelay != 0.3 {
			return errf("debounce = %.1f", sb.DebounceDelay)
		}
		return nil
	})

	assert("tabView", DocBlock{
		Type:   "tabView",
		ID:     "demo-tabs",
		Height: 200,
		Children: []DocBlock{
			{Type: "tab", ID: "t1", Title: "Overview", Children: []DocBlock{
				{Type: "text", ID: "t1txt", Text: "Overview tab"},
			}},
			{Type: "tab", ID: "t2", Title: "Settings", Children: []DocBlock{
				{Type: "text", ID: "t2txt", Text: "Settings tab"},
			}},
		},
	}, func(n Node) error {
		tv, ok := n.(*TabView)
		if !ok {
			return errf("type %T, want *TabView", n)
		}
		if len(tv.tabs) != 2 {
			return errf("tabs = %d, want 2", len(tv.tabs))
		}
		if tv.tabs[0].title != "Overview" {
			return errf("tab0 = %q", tv.tabs[0].title)
		}
		return nil
	})

	assert("dataTable", DocBlock{
		Type:   "dataTable",
		ID:     "emp-table",
		Height: 340,
		Props: map[string]any{
			"columns": []any{
				map[string]any{"title": "Name", "width": float64(160), "key": "name", "sortable": true},
				map[string]any{"title": "Dept", "width": float64(120), "key": "dept"},
			},
			"rows": []any{
				map[string]any{"name": "Alice", "dept": "Eng"},
				map[string]any{"name": "Bob", "dept": "Sales"},
			},
		},
	}, func(n Node) error {
		dt, ok := n.(*DataTable[docDataTableRow])
		if !ok {
			return errf("type %T, want *DataTable", n)
		}
		if len(dt.Columns) != 2 {
			return errf("columns = %d", len(dt.Columns))
		}
		if dt.binding.Len() != 2 {
			return errf("rows = %d", dt.binding.Len())
		}
		if dt.Columns[0].Title != "Name" || !dt.Columns[0].Sortable {
			return errf("col0 sortable/name")
		}
		return nil
	})

	assert("slider", DocBlock{Type: "slider", ID: "sl", Label: "Vol", Min: 0, Max: 10, Value: 5}, func(n Node) error {
		col, ok := n.(*Container)
		if !ok {
			return errf("type %T", n)
		}
		if _, ok := col.Children()[1].(*Slider); !ok {
			return errf("slider child %T", col.Children()[1])
		}
		return nil
	})

	collapsible := true
	vScroll := true
	assert("card-capabilities", DocBlock{
		Type:  "card",
		ID:    "cap-card",
		Title: "Advanced",
		Capabilities: &DocBlockCapabilities{
			Collapsible: &collapsible,
			VScroll:     &vScroll,
		},
		Children: []DocBlock{
			{Type: "text", ID: "cap-body", Text: "Scrollable collapsible body"},
		},
	}, func(n Node) error {
		c, ok := n.(*Card)
		if !ok {
			return errf("type %T, want *Card", n)
		}
		f := c.Features()
		if f == nil || !f.Collapsible {
			return errf("Collapsible = false, want true")
		}
		if !f.VScroll {
			return errf("VScroll = false, want true")
		}
		return nil
	})

	assert("section-capabilities-props", DocBlock{
		Type:  "section",
		ID:    "cap-section",
		Title: "Panel",
		Props: map[string]any{
			"capabilities": map[string]any{
				"collapsible": true,
				"vScroll":     true,
			},
		},
		Children: []DocBlock{
			{Type: "text", ID: "sec-body", Text: "via props.capabilities"},
		},
	}, func(n Node) error {
		p, ok := n.(*Panel)
		if !ok {
			return errf("type %T, want *Panel", n)
		}
		f := p.Features()
		if f == nil || !f.Collapsible || !f.VScroll {
			return errf("capabilities not applied: %+v", f)
		}
		return nil
	})
}

func errf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}
