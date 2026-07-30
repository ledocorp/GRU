package ui

import "testing"

// TestDocumentSpecBlockTypesCompileSmoke ensures every canonical type is accepted by the compiler.
func TestDocumentSpecBlockTypesCompileSmoke(t *testing.T) {
	ctx := NewBuildContext()
	ctx.Actions["act"] = func() {}
	chipSel := true

	minimal := func(typ string) DocBlock {
		b := DocBlock{Type: typ, ID: "smoke-" + typ, Title: "T", Text: "x", Width: 120, Height: 40}
		switch typ {
		case "field":
			b.Label = "L"
			b.Children = []DocBlock{{Type: "input", ID: "in"}}
		case "list", "breadcrumbs", "bottomnav", "bottomNavigation", "bottomNav":
			b.Items = []string{"a"}
		case "dropdown", "radioGroup", "radio", "combobox", "comboBox":
			b.Options = []string{"a"}
		case "toolbar":
			b.Children = []DocBlock{{Type: "toolbarGroup", ID: "g", Title: "G", Children: []DocBlock{
				{Type: "button", ID: "b", Text: "Go", OnClick: "act"},
			}}}
		case "tabView", "tabview":
			b.Children = []DocBlock{{Type: "tab", ID: "tab", Title: "One", Children: []DocBlock{
				{Type: "text", ID: "t", Text: "hi"},
			}}}
		case "dataTable", "datatable":
			b.Height = 120
			b.Props = map[string]any{
				"columns": []any{"Name"},
				"rows":    []any{map[string]any{"Name": "x"}},
			}
		case "table":
			b.Props = map[string]any{
				"columns": []any{"Col"},
				"rows":    []any{map[string]any{"Col": "v"}},
			}
		case "dateRangePicker", "daterangepicker":
			b.Props = map[string]any{"start": "2026-01-01", "end": "2026-01-31"}
		case "chip":
			b.Selectable = &chipSel
		case "listTile", "listtile":
			b.Title = "Title"
		case "presetRow":
			b.PresetItems = []PresetTileSpec{{Preset: "surface-card", Text: "Demo tile"}}
		case "backdrop":
			b.Children = []DocBlock{{Type: "text", ID: "t", Text: "on gradient"}}
		case "pagination":
			b.Value = 1
			b.Max = 3
		case "rating":
			b.Value = 3
			b.Max = 5
		case "slider":
			b.Min = 0
			b.Max = 10
			b.Value = 5
		case "progressBar", "progress":
			b.Value = 50
			b.Max = 100
		case "button":
			b.OnClick = "act"
		case "fab":
			b.OnClick = "act"
		}
		return b
	}

	seen := make(map[string]bool, len(DocumentSpecBlockTypes))
	for _, typ := range DocumentSpecBlockTypes {
		if typ == "" {
			continue
		}
		if seen[typ] {
			t.Fatalf("duplicate block type %q in DocumentSpecBlockTypes", typ)
		}
		seen[typ] = true
		block := minimal(typ)
		block.Type = typ
		if _, err := buildDocBlockAt(block, ctx, "smoke"); err != nil {
			t.Errorf("type %q: %v", typ, err)
		}
	}
}
