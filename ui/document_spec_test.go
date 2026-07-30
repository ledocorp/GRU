package ui

import "testing"

func TestBuildDocumentSpecCompilesBasicBlocks(t *testing.T) {
	ctx := NewBuildContext()
	clicked := false
	ctx.Actions["save"] = func() { clicked = true }

	node, err := BuildDocumentSpec(DocumentSpec{
		ID: "doc-test",
		Children: []DocBlock{
			{
				Type:  "text",
				ID:    "intro",
				Spans: []TextSpan{{Text: "Hello ", Bold: true}, {Text: "document"}},
			},
			{
				Type:      "button",
				ID:        "save",
				Text:      "Save",
				Component: "button",
				Variant:   "primary",
				OnClick:   "save",
				Width:     120,
			},
		},
	}, ctx)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}

	root, ok := node.(*Container)
	if !ok {
		t.Fatalf("root type = %T, want *Container", node)
	}
	if got := len(root.Children()); got != 2 {
		t.Fatalf("root children = %d, want 2", got)
	}
	if _, ok := root.Children()[0].(*RichText); !ok {
		t.Fatalf("first child type = %T, want *RichText", root.Children()[0])
	}
	btn, ok := root.Children()[1].(*Button)
	if !ok {
		t.Fatalf("second child type = %T, want *Button", root.Children()[1])
	}
	if btn.OnClick == nil {
		t.Fatal("button action was not wired")
	}
	if !btn.IsAutoHeight() {
		t.Fatal("document button should use intrinsic height")
	}
	if got := btn.GetPreferredWidth(); got != 120 {
		t.Fatalf("button preferred width = %.0f, want 120", got)
	}
	btn.OnClick()
	if !clicked {
		t.Fatal("button action did not run")
	}
}

func TestBuildDocumentSpecRejectsUnknownBlock(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "section", ID: "parent", Children: []DocBlock{{Type: "unknown", ID: "bad"}}}},
	}, nil)
	if err == nil {
		t.Fatal("expected unsupported block type error")
	}
	if got := err.Error(); got != `ui/document_spec: children[0].children[0] (id "bad", type "unknown"): unsupported block type "unknown"` {
		t.Fatalf("error = %q", got)
	}
}

func TestBuildDocumentSpecRejectsInvalidLayoutFields(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "text", ID: "bad-width", Width: -1, Text: "invalid"}},
	}, nil)
	if err == nil {
		t.Fatal("expected invalid width error")
	}

	_, err = BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "text", ID: "bad-width-range", MinWidth: 200, MaxWidth: 100, Text: "invalid"}},
	}, nil)
	if err == nil {
		t.Fatal("expected invalid minWidth/maxWidth error")
	}
}

func TestBuildDocumentSpecRejectsUnknownAction(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "button", ID: "missing-action", Text: "Run", OnClick: "missing"}},
	}, NewBuildContext())
	if err == nil {
		t.Fatal("expected unknown action error")
	}
	if got := err.Error(); got != `ui/document_spec: children[0] (id "missing-action", type "button"): unknown action "missing"` {
		t.Fatalf("error = %q", got)
	}
}

func TestBuildDocumentSpecCompilesControls(t *testing.T) {
	ctx := NewBuildContext()
	docSpec := DocumentSpec{
		Children: []DocBlock{
			{
				Type:        "input",
				ID:          "name-input",
				Value:       "Ada",
				Placeholder: "Name",
				Width:       220,
			},
			{
				Type:          "dropdown",
				ID:            "tone-dropdown",
				Options:       []string{"Default", "Primary", "Danger"},
				SelectedIndex: 2,
				Width:         180,
			},
			{
				Type:    "checkbox",
				ID:      "subscribe-checkbox",
				Label:   "Subscribe",
				Checked: true,
			},
			{
				Type:  "toggle",
				ID:    "public-toggle",
				Label: "Public",
			},
			{
				Type:          "radioGroup",
				ID:            "density-radio",
				Label:         "Density",
				Options:       []string{"Compact", "Comfortable"},
				SelectedIndex: 1,
			},
			{
				Type:        "slider",
				ID:          "volume-slider",
				Label:       "Volume",
				Min:         0,
				Max:         10,
				Value:       4.5,
				ValueFormat: "%.1f",
			},
		},
	}
	node, err := BuildDocumentSpec(docSpec, ctx)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}
	root := node.(*Container)

	input, ok := root.Children()[0].(*TextInput)
	if !ok {
		t.Fatalf("first child type = %T, want *TextInput", root.Children()[0])
	}
	if got := input.Text.Get(); got != "Ada" {
		t.Fatalf("input value = %q, want Ada", got)
	}
	if input.Placeholder != "Name" {
		t.Fatalf("input placeholder = %q, want Name", input.Placeholder)
	}
	if got := input.GetPreferredWidth(); got != 220 {
		t.Fatalf("input preferred width = %.0f, want 220", got)
	}

	dropdown, ok := root.Children()[1].(*Dropdown)
	if !ok {
		t.Fatalf("second child type = %T, want *Dropdown", root.Children()[1])
	}
	if got := dropdown.SelectedIndex.Get(); got != 2 {
		t.Fatalf("dropdown selected index = %d, want 2", got)
	}
	if got := len(dropdown.Options); got != 3 {
		t.Fatalf("dropdown options = %d, want 3", got)
	}

	checkboxWrap, ok := root.Children()[2].(*Container)
	if !ok {
		t.Fatalf("third child type = %T, want labeled *Container", root.Children()[2])
	}
	checkboxRow, ok := checkboxWrap.Children()[0].(*Container)
	if !ok {
		t.Fatalf("checkbox wrap child type = %T, want *Container row", checkboxWrap.Children()[0])
	}
	checkbox, ok := checkboxRow.Children()[1].(*Checkbox)
	if !ok {
		t.Fatalf("labeled checkbox child type = %T, want *Checkbox", checkboxRow.Children()[1])
	}
	if !checkbox.Value.Get() {
		t.Fatal("checkbox checked value = false, want true")
	}

	toggleWrap, ok := root.Children()[3].(*Container)
	if !ok {
		t.Fatalf("fourth child type = %T, want labeled *Container", root.Children()[3])
	}
	toggleRow, ok := toggleWrap.Children()[0].(*Container)
	if !ok {
		t.Fatalf("toggle wrap child type = %T, want *Container row", toggleWrap.Children()[0])
	}
	toggle, ok := toggleRow.Children()[1].(*Toggle)
	if !ok {
		t.Fatalf("labeled toggle child type = %T, want *Toggle", toggleRow.Children()[1])
	}
	if toggle.Value.Get() {
		t.Fatal("toggle checked value = true, want false")
	}

	radioField, ok := root.Children()[4].(*Container)
	if !ok {
		t.Fatalf("fifth child type = %T, want labeled *Container", root.Children()[4])
	}
	radio, ok := radioField.Children()[1].(*RadioGroup)
	if !ok {
		t.Fatalf("radio field child type = %T, want *RadioGroup", radioField.Children()[1])
	}
	if got := radio.Selected.Get(); got != 1 {
		t.Fatalf("radio selected index = %d, want 1", got)
	}

	sliderField, ok := root.Children()[5].(*Container)
	if !ok {
		t.Fatalf("sixth child type = %T, want labeled *Container", root.Children()[5])
	}
	slider, ok := sliderField.Children()[1].(*Slider)
	if !ok {
		t.Fatalf("slider field child type = %T, want *Slider", sliderField.Children()[1])
	}
	if got := slider.Value.Get(); got != 4.5 {
		t.Fatalf("slider value = %.1f, want 4.5", got)
	}

	input.Text.Set("Grace")
	dropdown.SelectedIndex.Set(1)
	toggle.Value.Set(true)
	radio.Selected.Set(0)
	slider.Value.Set(8.25)
	snapshot := ctx.ControlSnapshot()
	if snapshot["name-input"] != "Grace" {
		t.Fatalf("input snapshot = %v, want Grace", snapshot["name-input"])
	}
	if snapshot["tone-dropdown"] != "Primary" {
		t.Fatalf("dropdown snapshot = %v, want Primary", snapshot["tone-dropdown"])
	}
	if snapshot["subscribe-checkbox"] != true {
		t.Fatalf("checkbox snapshot = %v, want true", snapshot["subscribe-checkbox"])
	}
	if snapshot["public-toggle"] != true {
		t.Fatalf("toggle snapshot = %v, want true", snapshot["public-toggle"])
	}
	if snapshot["density-radio"] != "Compact" {
		t.Fatalf("radio snapshot = %v, want Compact", snapshot["density-radio"])
	}
	if snapshot["volume-slider"] != float32(8.25) {
		t.Fatalf("slider snapshot = %v, want 8.25", snapshot["volume-slider"])
	}

	// Rebuild and restore — hot-reload path.
	rebuilt, err := BuildDocumentSpec(docSpec, ctx)
	if err != nil {
		t.Fatalf("rebuild: %v", err)
	}
	ApplyControlSnapshot(rebuilt, snapshot)
	reRoot := rebuilt.(*Container)
	reInput := reRoot.Children()[0].(*TextInput)
	if got := reInput.Text.Get(); got != "Grace" {
		t.Fatalf("restored input = %q, want Grace", got)
	}
	reDropdown := reRoot.Children()[1].(*Dropdown)
	if reDropdown.Options[reDropdown.SelectedIndex.Get()] != "Primary" {
		t.Fatalf("restored dropdown = %q, want Primary", reDropdown.Options[reDropdown.SelectedIndex.Get()])
	}
}

func TestDocBlockFieldCaptionUsesCaptionStyle(t *testing.T) {
	lbl := docBlockFieldCaption("name-label", "Name").(*RichText)
	if lbl.styleName != "form-field-caption" {
		t.Fatalf("caption style = %q, want form-field-caption", lbl.styleName)
	}
	st := lbl.GetStyle()
	if st.FontSize != 19 {
		t.Fatalf("caption font size = %d, want 19", st.FontSize)
	}
	if !st.Bold {
		t.Fatal("caption should be bold")
	}
}

func TestBuildDocumentSpecCompilesFormAndField(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type: "form",
			ID:   "settings-form",
			Children: []DocBlock{{
				Type:  "field",
				ID:    "name-field",
				Label: "Name",
				Children: []DocBlock{{
					Type:        "input",
					ID:          "name-input",
					Placeholder: "Name",
				}},
			}},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}
	root := node.(*Container)
	form, ok := root.Children()[0].(*Container)
	if !ok {
		t.Fatalf("form type = %T, want *Container", root.Children()[0])
	}
	if got := form.GetStyle().Padding; got != 8 {
		t.Fatalf("form default padding = %.0f, want 8", got)
	}
	field, ok := form.Children()[0].(*Container)
	if !ok {
		t.Fatalf("field type = %T, want *Container", form.Children()[0])
	}
	if got := field.GetStyle().Padding; got != 4 {
		t.Fatalf("field default padding = %.0f, want 4", got)
	}
	caption, ok := field.Children()[0].(*RichText)
	if !ok {
		t.Fatalf("field first child type = %T, want *RichText", field.Children()[0])
	}
	if caption.styleName != "form-field-caption" {
		t.Fatalf("field caption style = %q, want form-field-caption", caption.styleName)
	}
	if st := caption.GetStyle(); st.FontSize != 19 {
		t.Fatalf("field caption font size = %d, want 19", st.FontSize)
	}
	if _, ok := field.Children()[1].(*TextInput); !ok {
		t.Fatalf("field second child type = %T, want *TextInput", field.Children()[1])
	}
}

func TestBuildDocumentSpecRejectsInvalidDropdown(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "dropdown", ID: "empty-options"}},
	}, nil)
	if err == nil {
		t.Fatal("expected empty dropdown options error")
	}

	_, err = BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:          "dropdown",
			ID:            "bad-selected",
			Options:       []string{"One"},
			SelectedIndex: 4,
		}},
	}, nil)
	if err == nil {
		t.Fatal("expected selectedIndex bounds error")
	}
}

func TestBuildDocumentSpecRejectsInvalidRadioAndSlider(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "radioGroup", ID: "empty-radio"}},
	}, nil)
	if err == nil {
		t.Fatal("expected empty radio options error")
	}

	_, err = BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:          "radioGroup",
			ID:            "bad-radio-selected",
			Options:       []string{"One"},
			SelectedIndex: 2,
		}},
	}, nil)
	if err == nil {
		t.Fatal("expected radio selectedIndex bounds error")
	}

	_, err = BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "slider", ID: "bad-slider", Min: 10, Max: 5}},
	}, nil)
	if err == nil {
		t.Fatal("expected invalid slider range error")
	}

	_, err = BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "slider", ID: "bad-slider-value", Value: "not-a-number"}},
	}, nil)
	if err == nil {
		t.Fatal("expected invalid slider value type error")
	}
}

func TestBuildDocumentSpecRejectsInvalidInputValue(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{Type: "input", ID: "bad-input-value", Value: 12}},
	}, nil)
	if err == nil {
		t.Fatal("expected invalid input value type error")
	}
}

func TestBuildDocumentSpecAppliesLayoutFields(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:     "card",
			ID:       "layout-card",
			Title:    "Layout Card",
			Width:    240,
			MinWidth: 180,
			MaxWidth: 320,
			Padding:  flexSpacing(22),
			Children: []DocBlock{{
				Type: "text",
				ID:   "layout-text",
				Text: "Layout fields",
			}},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}
	root := node.(*Container)
	card := root.Children()[0].(*Card)
	if got := card.GetPreferredWidth(); got != 240 {
		t.Fatalf("preferred width = %.0f, want 240", got)
	}
	if got := card.GetMinWidth(); got != 180 {
		t.Fatalf("min width = %.0f, want 180", got)
	}
	if got := card.GetMaxWidth(); got != 320 {
		t.Fatalf("max width = %.0f, want 320", got)
	}
	if got := card.GetStyle().Padding; got != 22 {
		t.Fatalf("padding = %.0f, want 22", got)
	}
}

func TestLoadDocumentSpecFromJSON(t *testing.T) {
	ctx := NewBuildContext()
	ctx.Actions["run"] = func() {}

	node, err := LoadDocumentSpec([]byte(`{
		"id": "json-doc",
		"children": [
			{
				"type": "section",
				"id": "intro-section",
				"title": "Intro",
				"children": [
					{
						"type": "text",
						"id": "intro-text",
						"selectable": false,
						"spans": [
							{ "text": "JSON-authored ", "bold": true },
							{ "text": "document block" }
						]
					},
					{
						"type": "button",
						"id": "json-button",
						"text": "Run",
						"onClick": "run",
						"width": 140
					},
					{
						"type": "card",
						"id": "json-card",
						"title": "Card",
						"children": [
							{ "type": "text", "id": "json-card-text", "text": "Card text" }
						]
					},
					{
						"type": "viewport",
						"id": "json-viewport",
						"height": 80,
						"children": [
							{ "type": "text", "id": "json-viewport-text", "text": "Viewport text" }
						]
					},
					{
						"type": "divider",
						"id": "json-divider"
					},
					{
						"type": "callout",
						"id": "json-callout",
						"title": "Note",
						"text": "Callout text"
					},
					{
						"type": "list",
						"id": "json-list",
						"items": ["One", "Two"]
					}
				]
			}
		]
	}`), ctx)
	if err != nil {
		t.Fatalf("LoadDocumentSpec returned error: %v", err)
	}
	root, ok := node.(*Container)
	if !ok {
		t.Fatalf("root type = %T, want *Container", node)
	}
	if got := len(root.Children()); got != 1 {
		t.Fatalf("root children = %d, want 1", got)
	}
	if _, ok := root.Children()[0].(*Panel); !ok {
		t.Fatalf("first child type = %T, want *Panel", root.Children()[0])
	}
	panel := root.Children()[0].(*Panel)
	if got := len(panel.Children()); got != 7 {
		t.Fatalf("panel children = %d, want 7", got)
	}
	rt, ok := panel.Children()[0].(*RichText)
	if !ok {
		t.Fatalf("panel first child type = %T, want *RichText", panel.Children()[0])
	}
	if rt.Selectable {
		t.Fatal("typed selectable=false was not applied")
	}
	btn, ok := panel.Children()[1].(*Button)
	if !ok {
		t.Fatalf("panel second child type = %T, want *Button", panel.Children()[1])
	}
	if got := btn.GetPreferredWidth(); got != 140 {
		t.Fatalf("json button preferred width = %.0f, want 140", got)
	}
	if _, ok := panel.Children()[2].(*Card); !ok {
		t.Fatalf("panel third child type = %T, want *Card", panel.Children()[2])
	}
	if _, ok := panel.Children()[3].(*Viewport); !ok {
		t.Fatalf("panel fourth child type = %T, want *Viewport", panel.Children()[3])
	}
	if _, ok := panel.Children()[4].(*Separator); !ok {
		t.Fatalf("panel fifth child type = %T, want *Separator", panel.Children()[4])
	}
	if _, ok := panel.Children()[5].(*Card); !ok {
		t.Fatalf("panel sixth child type = %T, want *Card", panel.Children()[5])
	}
	list, ok := panel.Children()[6].(*Container)
	if !ok {
		t.Fatalf("panel seventh child type = %T, want *Container", panel.Children()[6])
	}
	if got := len(list.Children()); got != 2 {
		t.Fatalf("list children = %d, want 2", got)
	}
}

func TestMarkdownToDocumentSpec(t *testing.T) {
	spec := MarkdownToDocumentSpec("md-doc", "Markdown Demo", `# Heading

This is **bold**, *italic*, `+"`code`"+`, and a [link](https://example.com).

`+"```go"+`
fmt.Println("hello")
`+"```"+`

- One
- Two

---

> A callout note`)

	if spec.ID != "md-doc" {
		t.Fatalf("spec ID = %q, want md-doc", spec.ID)
	}
	if got := len(spec.Children); got != 6 {
		t.Fatalf("markdown children = %d, want 6", got)
	}
	if spec.Children[0].Type != "text" || spec.Children[0].Spans[0].Variant != "h1" {
		t.Fatalf("first block = %+v, want h1 text", spec.Children[0])
	}
	if spec.Children[0].Height != 0 {
		t.Fatalf("heading height = %.0f, want 0 (AutoHeight RichText)", spec.Children[0].Height)
	}
	if spec.Children[2].Type != "code" || spec.Children[2].Title != "go" {
		t.Fatalf("third block = %+v, want fenced go code block", spec.Children[2])
	}
	if spec.Children[3].Type != "list" || len(spec.Children[3].Items) != 2 {
		t.Fatalf("fourth block = %+v, want two-item list", spec.Children[3])
	}
	if spec.Children[4].Type != "divider" {
		t.Fatalf("fifth block type = %q, want divider", spec.Children[4].Type)
	}
	if spec.Children[5].Type != "callout" || spec.Children[5].Variant != "blockquote" {
		t.Fatalf("sixth block = %+v, want blockquote callout", spec.Children[5])
	}
	if _, err := BuildDocumentSpec(spec, nil); err != nil {
		t.Fatalf("Markdown spec did not compile: %v", err)
	}
}

func TestBuildDocumentSpecFormatsListAndCallout(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{
			{Type: "list", ID: "list", Items: []string{"One"}},
			{Type: "callout", ID: "callout", Text: "Note"},
			{Type: "code", ID: "code", Text: "fmt.Println(\"hello\")"},
		},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}
	root := node.(*Container)
	list := root.Children()[0].(*Container)
	row := list.Children()[0].(*Container)
	var item *RichText
	for _, ch := range row.Children() {
		if rt, ok := ch.(*RichText); ok {
			item = rt
			break
		}
	}
	if item == nil {
		t.Fatalf("list row missing RichText among %d children", len(row.Children()))
	}
	if len(item.Spans) != 2 || item.Spans[0].Variant != "list-marker" {
		t.Fatalf("list item spans = %+v, want marker + item", item.Spans)
	}
	callout := root.Children()[1].(*Card)
	style := callout.GetStyle()
	want := CurrentThemeV2().Components["card"].Variants["callout"].BorderColor
	if style.BorderColor != want {
		t.Fatalf("callout border = %+v, want %+v", style.BorderColor, want)
	}
	if style.BackgroundColor.R == 255 && style.BackgroundColor.G == 255 && style.BackgroundColor.B == 255 {
		t.Fatalf("callout background still default white: %+v", style.BackgroundColor)
	}
	code := root.Children()[2].(*Card)
	codeStyle := code.GetStyle()
	codeWant := CurrentThemeV2().Components["card"].Variants["code"].BorderColor
	if codeStyle.BorderColor != codeWant {
		t.Fatalf("code border = %+v, want %+v", codeStyle.BorderColor, codeWant)
	}
}

func TestBuildDocumentSpecPresetNeoGlow(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:   "card",
			ID:     "neo",
			Title:  "Neo",
			Preset: "neo-glow-card",
			Props:  map[string]any{"glowIntensity": 0.4},
			Text:   "Preset path",
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	card := node.(*Container).Children()[0].(*Card)
	st := card.GetStyle()
	wantBG := CurrentThemeV2().Components["card"].Variants["neo-glow"].BackgroundColor
	if st.BackgroundColor != wantBG {
		t.Fatalf("background = %+v, want %+v", st.BackgroundColor, wantBG)
	}
	btnNode, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:   "button",
			ID:     "go",
			Text:   "Preset",
			Preset: "primary-button",
			Width:  120,
		}},
	}, nil)
	if err != nil {
		t.Fatalf("button preset: %v", err)
	}
	btn := btnNode.(*Container).Children()[0].(*Button)
	btnSt := btn.GetStyle()
	wantPrimary := CurrentThemeV2().Components["button"].Variants["primary"].BackgroundColor
	if btnSt.BackgroundColor != wantPrimary {
		t.Fatalf("button bg = %+v, want %+v", btnSt.BackgroundColor, wantPrimary)
	}
}

func TestBuildDocumentSpecUnknownPresetRejected(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:   "card",
			ID:     "bad",
			Preset: "not-real",
		}},
	}, nil)
	if err == nil {
		t.Fatal("expected unknown preset error")
	}
}

func TestBuildDocumentSpecCardPaddingZero(t *testing.T) {
	pad := float32(0)
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:  "card",
			ID:    "flush",
			Title: "padding: 0",
			Style: &DocBlockStyle{Padding: &pad},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	card := node.(*Container).Children()[0].(*Card)
	if got := card.GetStyle().Padding; got != 0 {
		t.Fatalf("style padding = %.0f, want 0", got)
	}
}

func TestBuildDocumentSpecDocBlockStyleOverrides(t *testing.T) {
	bgHex := "#FEF3C7"
	borderHex := "#D97706"
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type: "card",
			ID:   "styled-card",
			Style: &DocBlockStyle{
				BackgroundColor: &bgHex,
				BorderColor:     &borderHex,
			},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}
	card := node.(*Container).Children()[0].(*Card)
	style := card.GetStyle()
	wantBg, err := parseHexColor(bgHex)
	if err != nil {
		t.Fatal(err)
	}
	if style.BackgroundColor != wantBg {
		t.Fatalf("background = %+v, want %+v", style.BackgroundColor, wantBg)
	}
	wantBorder, err := parseHexColor(borderHex)
	if err != nil {
		t.Fatal(err)
	}
	if style.BorderColor != wantBorder {
		t.Fatalf("border = %+v, want %+v", style.BorderColor, wantBorder)
	}
}

func TestBuildDocumentSpecRejectsInvalidDocBlockStyle(t *testing.T) {
	bad := "not-a-color"
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:  "card",
			ID:    "bad-style",
			Style: &DocBlockStyle{BackgroundColor: &bad},
		}},
	}, nil)
	if err == nil {
		t.Fatal("expected invalid style error")
	}
}

func TestBuildDocumentSpecWiresLinkHandler(t *testing.T) {
	var clicked string
	ctx := NewBuildContext()
	ctx.LinkHandler = func(link string) {
		clicked = link
	}

	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:  "text",
			ID:    "linked-text",
			Spans: []TextSpan{{Text: "Docs", Link: "demo://docs"}},
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}

	rt := node.(*Container).Children()[0].(*RichText)
	if rt.OnLinkClick == nil {
		t.Fatal("link handler was not wired")
	}
	rt.OnLinkClick("demo://docs")
	if clicked != "demo://docs" {
		t.Fatalf("clicked link = %q, want demo://docs", clicked)
	}
}

func TestBuildDocumentSpecStabilityFixtureCompiles(t *testing.T) {
	ctx := NewBuildContext()
	ctx.Actions["fixtureSnapshot"] = func() {}
	ctx.Actions["fixturePrimaryAck"] = func() {}

	mdBlocks := MarkdownToDocumentSpec("fixture-md", "Fixture", "## Markdown slice\n\n- list item\n\n> callout").Children

	spec := DocumentSpec{
		ID: "doc-theme-fixture-root",
		Children: []DocBlock{{
			Type:  "section",
			ID:    "doc-theme-fixture-section",
			Title: "Stability Fixture",
			Gap:   flexSpacing(12),
			Children: append([]DocBlock{
				{
					Type: "text",
					ID:   "doc-theme-fixture-checklist",
					Spans: []TextSpan{
						{Text: "Verify scroll, selection, and controls.", Variant: "muted"},
					},
				},
			}, append(mdBlocks, DocBlock{
				Type: "form",
				ID:   "doc-theme-fixture-form",
				Children: []DocBlock{
					{
						Type:  "input",
						ID:    "doc-theme-fixture-input",
						Value: "Ada",
					},
					{
						Type:          "dropdown",
						ID:            "doc-theme-fixture-dropdown",
						Options:       []string{"Default", "Primary"},
						SelectedIndex: 1,
					},
					{
						Type:    "checkbox",
						ID:      "doc-theme-fixture-checkbox",
						Checked: true,
					},
					{
						Type: "toggle",
						ID:   "doc-theme-fixture-toggle",
					},
					{
						Type:          "radioGroup",
						ID:            "doc-theme-fixture-radio",
						Options:       []string{"Compact", "Comfortable"},
						SelectedIndex: 0,
					},
					{
						Type:  "slider",
						ID:    "doc-theme-fixture-slider",
						Min:   0,
						Max:   100,
						Value: float64(42),
					},
				},
			})...),
		}},
	}

	root, err := BuildDocumentSpec(spec, ctx)
	if err != nil {
		t.Fatalf("BuildDocumentSpec returned error: %v", err)
	}
	if _, ok := root.(*Container); !ok {
		t.Fatalf("root type = %T, want *Container", root)
	}

	wantIDs := []string{
		"doc-theme-fixture-input",
		"doc-theme-fixture-dropdown",
		"doc-theme-fixture-checkbox",
		"doc-theme-fixture-toggle",
		"doc-theme-fixture-radio",
		"doc-theme-fixture-slider",
	}
	for _, id := range wantIDs {
		if _, ok := ctx.ControlValue(id); !ok {
			t.Fatalf("control value getter missing for %q", id)
		}
	}

	snapshot := ctx.ControlSnapshot()
	if snapshot["doc-theme-fixture-input"] != "Ada" {
		t.Fatalf("input snapshot = %v, want Ada", snapshot["doc-theme-fixture-input"])
	}
	if snapshot["doc-theme-fixture-checkbox"] != true {
		t.Fatalf("checkbox snapshot = %v, want true", snapshot["doc-theme-fixture-checkbox"])
	}
}
