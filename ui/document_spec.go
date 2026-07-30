// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// DocumentSpec is declarative JSON that compiles into Gru's retained Node tree.
// Go is the engine: this package only maps DocBlock types to existing New* calls
// (Panel, Card, RichText, …). JSON cannot request behavior Go does not support.
// See docs/GO_FIRST_UI.md.
type DocumentSpec struct {
	ID              string     `json:"id"`
	Title           string     `json:"title"`
	Theme           string     `json:"theme"`
	SyntaxHighlight *bool      `json:"syntaxHighlight"` // page scope: false disables Chroma + inline-code pills
	Children        []DocBlock `json:"children"`
}

// DocBlock is one declarative block in a DocumentSpec. Common layout and
// control properties are represented as typed JSON fields so generated specs
// can be validated before they compile; Props remains an escape hatch for
// experimental metadata.
type DocBlock struct {
	Type          string         `json:"type"`
	ID            string         `json:"id"`
	Title         string         `json:"title"`
	Component     string         `json:"component"`
	Variant       string         `json:"variant"`
	Preset        string         `json:"preset"`
	Label         string         `json:"label"`
	Text          string         `json:"text"`
	Value         any            `json:"value"`
	Placeholder   string         `json:"placeholder"`
	Spans         []TextSpan     `json:"spans"`
	Children      []DocBlock     `json:"children"`
	Width         float32        `json:"width"`
	Height        float32        `json:"height"`
	MinWidth      float32        `json:"minWidth"`
	MaxWidth      float32        `json:"maxWidth"`
	Gap           flexSpacing    `json:"gap"`
	Padding       flexSpacing    `json:"padding"`
	FlexGrow      *float32       `json:"flexGrow"`
	Wrap          *bool          `json:"wrap"`
	ClipChildren  *bool          `json:"clipChildren"`
	PresetItems   []PresetTileSpec
	MinTileWidth  float32        `json:"minTileWidth"`
	Columns       int            `json:"columns"`
	Backdrop      *bool          `json:"backdrop"`
	Selectable    *bool          `json:"selectable"`
	OnClick       string         `json:"onClick"`
	Items         []string       `json:"items"`
	SyntaxHighlight *bool        `json:"syntaxHighlight"` // section/container scope for descendants
	Ordered       bool           `json:"ordered"`
	Checked       bool           `json:"checked"`
	Disabled      bool           `json:"disabled"`
	Min           float32        `json:"min"`
	Max           float32        `json:"max"`
	Vertical      *bool          `json:"vertical"`
	RowHeight     float32        `json:"rowHeight"`
	ShowValue     *bool          `json:"showValue"`
	ValueFormat   string         `json:"valueFormat"`
	Options       []string       `json:"options"`
	SelectedIndex int            `json:"selectedIndex"`
	// Style is optional compile-time chrome (#RRGGBB hex colors, padding, etc.).
	// Merged after component/variant resolution — see docs/DOCUMENT_SPEC_AUTHORING.md.
	Style *DocBlockStyle `json:"style"`
	// Capabilities toggles egui-style shell features (collapse, scroll, dismiss, …).
	// See docs/DOCUMENT_SPEC_AUTHORING.md — Surfaces capabilities (C6).
	Capabilities *DocBlockCapabilities `json:"capabilities"`
	Props map[string]any `json:"props"`
}

// UnmarshalJSON routes "items" to PresetItems for presetRow blocks and to Items otherwise.
func (b *DocBlock) UnmarshalJSON(data []byte) error {
	type Alias DocBlock
	aux := struct {
		Items json.RawMessage `json:"items"`
		*Alias
	}{
		Alias: (*Alias)(b),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if len(aux.Items) == 0 {
		return nil
	}
	if b.Type == "presetRow" {
		return json.Unmarshal(aux.Items, &b.PresetItems)
	}
	return json.Unmarshal(aux.Items, &b.Items)
}

// DocBlockStyle holds optional per-block visual overrides (Theme v2 hex color strings).
type DocBlockStyle styleJSON

func (s *DocBlockStyle) compile() (Style, error) {
	if s == nil {
		return Style{}, nil
	}
	return styleJSON(*s).toStyle(Style{})
}

// BuildDocumentSpec compiles DocumentSpec into the same retained UI nodes as
// hand-written Go. Supported blocks include page, section, card, surface, viewport, row, column,
// form, field, text, buttonRow, divider, callout, code, presetRow, list, button, input,
// dropdown, checkbox, toggle, radioGroup, slider, badge, progressBar, listTile, appBar,
// chip, rating, bottomnav, fab, avatar, breadcrumbs, pagination, toolbar, searchBar,
// tabView, and dataTable. Generated controls
// register current-value getters on BuildContext when they have an ID.
//
// Each block type maps to a documented Go recipe — see docs/DOCUMENT_SPEC_GO_RECIPES.md
// and ui/document_spec_recipe_test.go.
func BuildDocumentSpec(spec DocumentSpec, ctx *BuildContext) (Node, error) {
	if ctx == nil {
		ctx = NewBuildContext()
	}
	if spec.SyntaxHighlight != nil {
		c := *ctx
		c.SyntaxHighlight = spec.SyntaxHighlight
		ctx = &c
	}
	id := spec.ID
	if id == "" {
		id = "document-spec"
	}
	root := NewContainer(id, 0, 0, 0, 0)
	root.SetStyle("transparent")
	root.FlexDirection = FlexColumn
	root.Gap = 12
	if ctx != nil && ctx.ContentWidth > 0 {
		root.PreferredWidth = ctx.ContentWidth
	}
	for i := range spec.Children {
		child, err := buildDocBlockAt(spec.Children[i], ctx, fmt.Sprintf("children[%d]", i))
		if err != nil {
			return nil, err
		}
		root.AddChild(child)
	}
	return root, nil
}

// ParseDocumentSpec unmarshals JSON into a DocumentSpec without compiling.
func ParseDocumentSpec(data []byte) (DocumentSpec, error) {
	var spec DocumentSpec
	if err := json.Unmarshal(data, &spec); err != nil {
		return DocumentSpec{}, fmt.Errorf("ui/document_spec: parse error: %w", err)
	}
	return spec, nil
}

// LoadDocumentSpec parses a JSON document spec and compiles it into ordinary
// retained UI nodes.
func LoadDocumentSpec(data []byte, ctx *BuildContext) (Node, error) {
	spec, err := ParseDocumentSpec(data)
	if err != nil {
		return nil, err
	}
	return BuildDocumentSpec(spec, ctx)
}

func buildDocBlock(block DocBlock, ctx *BuildContext) (Node, error) {
	return buildDocBlockAt(block, ctx, "block")
}

func buildDocBlockAt(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	ctx = docCtxWithBlockSyntaxHighlight(ctx, block)
	if block.Width == 0 && ctx != nil && ctx.ContentWidth > 0 {
		block.Width = ctx.ContentWidth
	}
	if err := validateDocBlock(block, path); err != nil {
		return nil, err
	}
	id := block.ID
	if id == "" {
		id = "doc-" + block.Type
	}
	switch block.Type {
	case "page", "column", "":
		c := NewContainer(id, 0, 0, block.Width, block.Height)
		c.SetStyle("transparent")
		c.FlexDirection = FlexColumn
		c.Gap = docBlockGap(block, 10)
		if err := addDocChildren(c, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocStyle(&c.Element, block)
		applyDocLayout(&c.Element, block)
		return c, nil
	case "form":
		c := NewContainer(id, 0, 0, block.Width, block.Height)
		c.SetStyle("transparent")
		c.FlexDirection = FlexColumn
		c.Gap = docBlockGap(block, 16)
		if err := addDocChildren(c, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocStyle(&c.Element, block)
		applyDocLayout(&c.Element, block)
		if docBlockPadding(block) == 0 {
			applyElementPadding(&c.Element, 8)
		}
		return c, nil
	case "field":
		c := NewContainer(id, 0, 0, block.Width, block.Height)
		c.SetStyle("transparent")
		c.FlexDirection = FlexColumn
		c.Gap = docBlockGap(block, 6)
		if label := docBlockLabel(block); label != "" {
			c.AddChild(docBlockFieldCaption(id+"-label", label))
		}
		if err := addDocChildren(c, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocStyle(&c.Element, block)
		applyDocLayout(&c.Element, block)
		if docBlockPadding(block) == 0 {
			applyElementPadding(&c.Element, 4)
		}
		return c, nil
	case "row", "buttonRow":
		c := NewContainer(id, 0, 0, block.Width, block.Height)
		c.SetStyle("transparent")
		c.FlexDirection = FlexRow
		c.Gap = docBlockGap(block, 10)
		c.AutoHeight = true
		applyDocContainerFlags(c, block)
		if err := addDocChildren(c, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocStyle(&c.Element, block)
		applyDocLayout(&c.Element, block)
		return c, nil
	case "card":
		card := NewCard(id, block.Title, 0, 0, block.Width, block.Height)
		card.Gap = docBlockGap(block, 10)
		if block.Height > 0 {
			// Fixed height (e.g. vScroll demos) — do not shrink-wrap to content.
			card.AutoHeight = false
		}
		if err := addDocChildren(card, block.Children, ctx, path); err != nil {
			return nil, err
		}
		if block.Preset != "" {
			_ = card.SetStylePreset(block.Preset, PresetPropsFromMap(block.Props))
		} else {
			applyDocStyle(&card.Element, block)
		}
		if v := block.Variant; v == "callout" || v == "code" {
			applyCardRecipeVariant(card, v)
		}
		applyDocLayout(&card.Element, block)
		applyCardChromeTextToBody(card, block)
		if err := applyDocBlockCapabilities(block, &card.SurfaceShell, ctx, path); err != nil {
			return nil, err
		}
		return card, nil
	case "callout":
		if block.Variant == "blockquote" {
			return buildDocBlockquote(block, ctx, path)
		}
		card := NewCard(id, block.Title, 0, 0, block.Width, block.Height)
		applyDocStyle(&card.Element, block)
		configureDocMarkdownCard(card, "callout", block.Title)
		if block.Text != "" || len(block.Spans) > 0 {
			spans := block.Spans
			if len(spans) == 0 {
				spans = []TextSpan{{Text: block.Text}}
			}
			spans = docApplySyntaxHighlightSpans(spans, ctx)
			rt := NewRichText(id+"-text", spans, 0, 0, 0, 0)
			rt.SetStyle("richtext-callout")
			applyDocRichText(rt, block, ctx)
			card.AddChild(rt)
		}
		if err := addDocChildren(card, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocLayout(&card.Element, block)
		applyCardChromeTextToBody(card, block)
		if err := applyDocBlockCapabilities(block, &card.SurfaceShell, ctx, path); err != nil {
			return nil, err
		}
		return card, nil
	case "code":
		langTitle := formatCodeLang(block.Title)
		card := NewCard(id, langTitle, 0, 0, block.Width, block.Height)
		if langTitle != "" {
			card.TitleHeight = 28
		}
		applyDocStyle(&card.Element, block)
		configureDocMarkdownCard(card, "code", langTitle)
		rt := NewRichText(id+"-text", docCodeBlockSpans(block.Text, block.Title, ctx), 0, 0, 0, 0)
		rt.SetStyle("richtext-code-block")
		applyDocRichText(rt, block, ctx)
		rt.Wrap = false
		rt.AutoHeight = true
		rt.LineGap = 2
		card.AddChild(WrapCodeBlockHorizontalScroll(id, rt))
		if err := addDocChildren(card, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocLayout(&card.Element, block)
		applyCardChromeTextToBody(card, block)
		if err := applyDocBlockCapabilities(block, &card.SurfaceShell, ctx, path); err != nil {
			return nil, err
		}
		return card, nil
	case "viewport":
		vp := NewViewport(id, 0, 0, block.Width, block.Height)
		vp.SetStyle("transparent")
		vp.Gap = docBlockGap(block, 10)
		if block.Height > 0 {
			vp.AutoHeight = false
		}
		if err := addDocChildren(vp, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocStyle(&vp.Element, block)
		applyDocLayout(&vp.Element, block)
		return vp, nil
	case "section":
		// section → Panel: padded document region with body clamp (not just layout).
		// Use for page surfaces in a viewport; use card for sub-groups inside.
		title := block.Title
		if title == "" {
			title = block.Text
		}
		p := NewPanel(id, title, 0, 0, block.Width, block.Height)
		p.Gap = docBlockGap(block, 10)
		if err := addDocChildren(p, block.Children, ctx, path); err != nil {
			return nil, err
		}
		if block.Preset != "" {
			_ = p.SetStylePreset(block.Preset, PresetPropsFromMap(block.Props))
		} else {
			applyDocStyle(&p.Element, block)
		}
		applyDocLayout(&p.Element, block)
		if err := applyDocBlockCapabilities(block, &p.SurfaceShell, ctx, path); err != nil {
			return nil, err
		}
		return p, nil
	case "surface":
		// surface → Panel or Card from variant / preset component (C6).
		return buildDocSurfaceAt(block, ctx, path, surfaceKindAuto)
	case "backdrop":
		d := NewDemoBackdrop(id, 0, 0, block.Width, block.Height)
		d.Gap = docBlockGap(block, 12)
		applyDocContainerFlags(&d.Container, block)
		if err := addDocChildren(d, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocLayout(&d.Element, block)
		if padding, ok := docBlockPaddingOverride(block); ok {
			applyElementPadding(&d.Element, padding)
		}
		return d, nil
	case "presetRow":
		if len(block.PresetItems) == 0 {
			return nil, fmt.Errorf("%s: presetRow requires items", path)
		}
		opts := DefaultPresetRowOptions()
		opts.Gap = docBlockGap(block, opts.Gap)
		if block.MinTileWidth > 0 {
			opts.MinTileWidth = block.MinTileWidth
		}
		if block.Columns > 0 {
			opts.Columns = block.Columns
		}
		if block.Backdrop != nil {
			opts.UseBackdrop = *block.Backdrop
		}
		if block.ClipChildren != nil {
			opts.ClipBackdrop = *block.ClipChildren
		}
		node, err := BuildPresetRow(id, block.PresetItems, opts)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		return node, nil
	case "text":
		spans := block.Spans
		if len(spans) == 0 {
			spans = []TextSpan{{Text: block.Text}}
		}
		spans = docApplySyntaxHighlightSpans(spans, ctx)
		rt := NewRichText(id, spans, 0, 0, block.Width, block.Height)
		applyDocRichText(rt, block, ctx)
		applyDocStyle(&rt.Element, block)
		applyDocLayout(&rt.Element, block)
		// Fixed-height markers (e.g. gallery pipe "|") must not AutoHeight-grow
		// past the badge row — that desyncs cross-axis centering.
		if block.Height > 0 {
			rt.AutoHeight = false
			rt.Wrap = false
		}
		return rt, nil
	case "divider", "separator":
		label := block.Title
		if label == "" {
			label = block.Text
		}
		h := block.Height
		if h == 0 {
			h = 12
			if label != "" {
				h = 28
			}
		}
		sep := NewSeparator(id, label, 0, 0, block.Width, h)
		applyDocStyle(&sep.Element, block)
		applyDocLayout(&sep.Element, block)
		return sep, nil
	case "list":
		c := NewContainer(id, 0, 0, block.Width, block.Height)
		c.SetStyle("transparent")
		c.FlexDirection = FlexColumn
		c.Gap = docBlockGap(block, 2)
		depths := docBlockIntSliceProp(block, "depths")
		numbers := docBlockIntSliceProp(block, "numbers")
		tasks := docBlockBoolSliceProp(block, "task")
		taskDone := docBlockBoolSliceProp(block, "taskDone")
		if len(block.Items) > 0 {
			itemSpans := docBlockTextSpanSliceProp(block, "itemSpans")
			for i, item := range block.Items {
				depth := 0
				if i < len(depths) {
					depth = depths[i]
				}
				isTask := i < len(tasks) && tasks[i]
				done := isTask && i < len(taskDone) && taskDone[i]
				spans := parseMarkdownInline(item)
				if i < len(itemSpans) && len(itemSpans[i]) > 0 {
					spans = itemSpans[i]
				}
				spans = docApplySyntaxHighlightSpans(spans, ctx)
				row := NewContainer(fmt.Sprintf("%s-item-%d", id, i), 0, 0, block.Width, 0)
				row.SetStyle("transparent")
				row.FlexDirection = FlexRow
				row.Gap = docBlockGap(block, 8)
				row.AutoHeight = true
				if depth > 0 {
					inset := NewContainer(fmt.Sprintf("%s-item-%d-inset", id, i), 0, 0, float32(depth)*22, 0)
					inset.SetStyle("transparent")
					row.AddChild(inset)
				}
				if isTask {
					cbSize := float32(20)
					cb := NewCheckbox(fmt.Sprintf("%s-item-%d-cb", id, i), done, 0, 0, cbSize, cbSize)
					cb.Disabled = true
					row.AddChild(cb)
					rt := NewRichText(fmt.Sprintf("%s-item-%d-text", id, i), spans, 0, 0, 0, 0)
					rt.LineGap = 2
					rt.SetFlexGrow(1)
					applyDocRichText(rt, block, ctx)
					row.AddChild(rt)
					c.AddChild(row)
					continue
				}
				prefix := "• "
				if block.Ordered {
					n := i + 1
					if i < len(numbers) && numbers[i] > 0 {
						n = numbers[i]
					}
					prefix = fmt.Sprintf("%d. ", n)
				}
				spans = append([]TextSpan{{Text: prefix, Variant: "list-marker"}}, spans...)
				rt := NewRichText(fmt.Sprintf("%s-item-%d-text", id, i), spans, 0, 0, 0, 0)
				rt.LineGap = 2
				rt.SetFlexGrow(1)
				applyDocRichText(rt, block, ctx)
				row.AddChild(rt)
				c.AddChild(row)
			}
		}
		if err := addDocChildren(c, block.Children, ctx, path); err != nil {
			return nil, err
		}
		applyDocStyle(&c.Element, block)
		applyDocLayout(&c.Element, block)
		return c, nil
	case "badge":
		text := block.Text
		if text == "" {
			text = block.Label
		}
		if text == "" {
			text = "Badge"
		}
		badge := NewBadge(id, text, docBadgeVariant(block.Variant), 0, 0, docBlockWidth(block), block.Height)
		if block.Component != "" && block.Variant != "" {
			badge.SetStyleVariant(block.Component, block.Variant)
		}
		applyDocLayout(&badge.Element, block)
		return badge, nil
	case "progressBar", "progress":
		minVal, maxVal := docBlockRange(block, 0, 100)
		if minVal >= maxVal {
			return nil, docBlockError(block, path, "min must be < max")
		}
		raw := docBlockNumericValue(block, minVal)
		norm := (raw - minVal) / (maxVal - minVal)
		if norm < 0 {
			norm = 0
		}
		if norm > 1 {
			norm = 1
		}
		pbW := docBlockWidth(block)
		if pbW == 0 {
			pbW = 280
		}
		pbH := block.Height
		if pbH == 0 {
			pbH = 12
		}
		pb := NewProgressBar(id, norm, 0, 0, pbW, pbH)
		if pbW > 0 {
			pb.PreferredWidth = pbW
		}
		registerDocControlValue(ctx, id, func() any { return pb.Value.Get() })
		applyDocStyle(&pb.Element, block)
		applyDocLayout(&pb.Element, block)
		return pb, nil
	case "button":
		btnW := docBlockWidth(block)
		btn := NewButton(id, block.Text, 0, 0, btnW, 0)
		btn.AutoHeight = true
		if btnW > 0 {
			btn.PreferredWidth = btnW
		}
		actionName := docBlockOnClick(block)
		if actionName != "" {
			if ctx.Actions == nil || ctx.Actions[actionName] == nil {
				return nil, docBlockError(block, path, "unknown action %q", actionName)
			}
			btn.OnClick = ctx.Actions[actionName]
		}
		applyDocStyle(&btn.Element, block)
		applyDocLayout(&btn.Element, block)
		return btn, nil
	case "input":
		inputW := docBlockWidth(block)
		inputH := block.Height
		if inputH == 0 {
			inputH = 40
		}
		input := NewTextInput(id, docBlockValue(block), 0, 0, inputW, inputH)
		input.Placeholder = docBlockPlaceholder(block)
		input.Disabled = block.Disabled
		if inputW > 0 {
			input.PreferredWidth = inputW
		}
		registerDocControlValue(ctx, id, func() any { return input.Text.Get() })
		applyDocStyle(&input.Element, block)
		applyDocLayout(&input.Element, block)
		return input, nil
	case "dropdown":
		dropdownW := docBlockWidth(block)
		dropdownH := block.Height
		if dropdownH == 0 {
			dropdownH = 40
		}
		dropdown := NewDropdown(id, docBlockOptions(block), docBlockSelectedIndex(block), 0, 0, dropdownW, dropdownH)
		if dropdownW > 0 {
			dropdown.PreferredWidth = dropdownW
		}
		dropdown.Disabled = block.Disabled
		registerDocControlValue(ctx, id, func() any {
			idx := dropdown.SelectedIndex.Get()
			if idx < 0 || idx >= len(dropdown.Options) {
				return ""
			}
			return dropdown.Options[idx]
		})
		applyDocStyle(&dropdown.Element, block)
		applyDocLayout(&dropdown.Element, block)
		return dropdown, nil
	case "checkbox":
		cbW, cbH := docBlockControlSize(block, 26, 26)
		checkbox := NewCheckbox(id, docBlockChecked(block), 0, 0, cbW, cbH)
		checkbox.Disabled = block.Disabled
		registerDocControlValue(ctx, id, func() any { return checkbox.Value.Get() })
		applyDocStyle(&checkbox.Element, block)
		applyDocLayout(&checkbox.Element, block)
		return docBlockLabeledControl(block, checkbox), nil
	case "toggle":
		toggleW, toggleH := docBlockControlSize(block, 48, 24)
		toggle := NewToggle(id, docBlockChecked(block), 0, 0, toggleW, toggleH)
		toggle.Disabled = block.Disabled
		registerDocControlValue(ctx, id, func() any { return toggle.Value.Get() })
		applyDocStyle(&toggle.Element, block)
		applyDocLayout(&toggle.Element, block)
		return docBlockLabeledControl(block, toggle), nil
	case "radioGroup", "radio":
		rgW := docBlockWidth(block)
		if rgW == 0 {
			rgW = 260
		}
		rowH := docBlockRowHeight(block)
		rgH := block.Height
		if rgH == 0 {
			if docBlockVertical(block, true) {
				rgH = rowH * float32(len(docBlockOptions(block)))
			} else {
				rgH = rowH
			}
		}
		rg := NewRadioGroup(id, docBlockOptions(block), 0, 0, rgW, rgH)
		rg.RowH = rowH
		rg.Vertical = docBlockVertical(block, true)
		rg.Selected.Set(docBlockSelectedIndex(block))
		if rgW > 0 {
			rg.PreferredWidth = rgW
		}
		registerDocControlValue(ctx, id, func() any {
			idx := rg.Selected.Get()
			if idx < 0 || idx >= len(rg.Options) {
				return ""
			}
			return rg.Options[idx]
		})
		applyDocStyle(&rg.Element, block)
		applyDocLayout(&rg.Element, block)
		return docBlockStackedLabeledControl(block, rg), nil
	case "slider":
		minVal, maxVal := docBlockRange(block, 0, 100)
		sliderW := docBlockWidth(block)
		if sliderW == 0 {
			sliderW = 260
		}
		sliderH := block.Height
		if sliderH == 0 {
			sliderH = 36
		}
		slider := NewSlider(id, minVal, maxVal, docBlockNumericValue(block, minVal), 0, 0, sliderW, sliderH)
		if block.ShowValue != nil {
			slider.ShowValue = *block.ShowValue
		}
		if block.ValueFormat != "" {
			slider.ValueFmt = block.ValueFormat
		}
		if sliderW > 0 {
			slider.PreferredWidth = sliderW
		}
		registerDocControlValue(ctx, id, func() any { return slider.Value.Get() })
		applyDocStyle(&slider.Element, block)
		applyDocLayout(&slider.Element, block)
		return docBlockStackedLabeledControl(block, slider), nil
	case "listTile", "listtile":
		title := block.Title
		if title == "" {
			title = block.Label
		}
		if title == "" {
			title = "Row"
		}
		subtitle := block.Text
		lt := NewListTile(id, title, subtitle, 0, 0, docBlockWidth(block), 0)
		if docBlockStringProp(block, "trailing") == "toggle" {
			tw, th := docBlockControlSize(block, 52, 28)
			lt.SetTrailing(NewToggle(id+"-tog", docBlockChecked(block), 0, 0, tw, th))
		} else if chev := docBlockStringProp(block, "trailing"); chev != "" {
			lt.SetTrailing(NewLabel(id+"-trail", chev, 0, 0, 24, 24))
		}
		if action := docBlockOnClick(block); action != "" {
			if ctx.Actions == nil || ctx.Actions[action] == nil {
				return nil, docBlockError(block, path, "unknown action %q", action)
			}
			lt.OnClick = ctx.Actions[action]
		}
		applyDocStyle(&lt.Element, block)
		applyDocLayout(&lt.Element, block)
		return lt, nil
	case "appBar", "appbar":
		title := block.Title
		if title == "" {
			title = "App"
		}
		bar := NewAppBar(id, title, 0, 0, docBlockWidth(block), 0)
		if sym := docBlockStringProp(block, "leadingIcon"); sym != "" {
			ib := NewIconButton(id+"-lead", sym, "", 0, 0, 40, 40)
			bar.SetLeading(ib)
		}
		if sym := docBlockStringProp(block, "trailingIcon"); sym != "" {
			bar.AddTrailing(NewIconButton(id+"-trail", sym, "", 0, 0, 40, 40))
		}
		applyDocStyle(&bar.Element, block)
		applyDocLayout(&bar.Element, block)
		return bar, nil
	case "chip":
		// Alias for badge with optional Selected / CloseButton (filter chips).
		text := block.Text
		if text == "" {
			text = block.Label
		}
		if text == "" {
			text = "Chip"
		}
		h := block.Height
		if h == 0 {
			h = chipDefaultH
		}
		badge := NewBadge(id, text, docBadgeVariant(block.Variant), 0, 0, docBlockWidth(block), h)
		if docBlockSelectable(block, false) {
			badge.Selected = NewSignal(docBlockChecked(block))
			registerDocControlValue(ctx, id, func() any { return badge.Selected.Get() })
		}
		if docBlockBool(block, "dismissible", false) {
			badge.CloseButton = true
			if action := docBlockOnClick(block); action != "" {
				if ctx.Actions == nil || ctx.Actions[action] == nil {
					return nil, docBlockError(block, path, "unknown action %q", action)
				}
				badge.OnClose = ctx.Actions[action]
			} else {
				badge.OnClose = func() { badge.Hide() }
			}
		}
		applyDocLayout(&badge.Element, block)
		return badge, nil
	case "rating":
		maxStars := int(block.Max)
		if maxStars <= 0 {
			maxStars = int(docBlockFloat(block, "maxStars", ratingDefaultStars))
		}
		if maxStars <= 0 {
			maxStars = ratingDefaultStars
		}
		val := docBlockNumericValue(block, 0)
		sig := NewSignal(val)
		registerDocControlValue(ctx, id, func() any { return sig.Get() })
		rt := NewRating(id, sig, maxStars, 0, 0, docBlockWidth(block), block.Height)
		applyDocLayout(&rt.Element, block)
		return rt, nil
	case "bottomnav", "bottomNavigation", "bottomNav":
		items := docBlockBottomNavItems(block)
		sel := NewSignal(docBlockSelectedIndex(block))
		bn := NewBottomNavigationBar(id, items, sel, 0, 0, docBlockWidth(block), block.Height)
		registerDocControlValue(ctx, id, func() any { return bn.Selected.Get() })
		applyDocLayout(&bn.Element, block)
		return bn, nil
	case "fab":
		icon := docBlockStringProp(block, "icon")
		if icon == "" {
			icon = block.Text
		}
		if icon == "" {
			icon = "+"
		}
		fab := NewFAB(id, icon, block.Label, nil, 0, 0, 0, 0)
		if action := docBlockOnClick(block); action != "" {
			if ctx.Actions == nil || ctx.Actions[action] == nil {
				return nil, docBlockError(block, path, "unknown action %q", action)
			}
			fab.OnClick = ctx.Actions[action]
		}
		applyDocLayout(&fab.Element, block)
		return fab, nil
	case "avatar":
		initials := block.Text
		if initials == "" {
			initials = block.Label
		}
		if initials == "" {
			initials = "?"
		}
		av := NewAvatar(id, docBlockStringProp(block, "imagePath"), initials, 0, 0, docBlockWidth(block), block.Height)
		if docBlockBool(block, "statusOnline", false) {
			av.ShowStatus = true
			av.StatusOnline = true
		}
		applyDocLayout(&av.Element, block)
		return av, nil
	case "breadcrumbs":
		items := block.Items
		if len(items) == 0 {
			items = []string{"Home"}
		}
		bc := NewBreadcrumbs(id, items, 0, 0, docBlockWidth(block), block.Height)
		if action := docBlockOnClick(block); action != "" {
			if ctx.Actions == nil || ctx.Actions[action] == nil {
				return nil, docBlockError(block, path, "unknown action %q", action)
			}
			bc.OnClick = func(i int) { ctx.Actions[action]() }
		}
		applyDocLayout(&bc.Element, block)
		return bc, nil
	case "combobox", "comboBox":
		opts := docBlockOptions(block)
		initial := block.Text
		if initial == "" && len(opts) > 0 {
			initial = opts[0]
		}
		sel := NewSignal(initial)
		h := block.Height
		if h == 0 {
			h = 40
		}
		cb := NewComboBox(id, opts, sel, 0, 0, docBlockWidth(block), h)
		registerDocControlValue(ctx, id, func() any { return sel.Get() })
		applyDocLayout(&cb.Element, block)
		return cb, nil
	case "dateRangePicker", "daterangepicker":
		start := docBlockParseDate(docBlockStringProp(block, "start"))
		end := docBlockParseDate(docBlockStringProp(block, "end"))
		startSig := NewSignal(start)
		endSig := NewSignal(end)
		h := block.Height
		if h == 0 {
			h = 40
		}
		drp := NewDateRangePicker(id, startSig, endSig, 0, 0, docBlockWidth(block), h)
		registerDocControlValue(ctx, id+"_start", func() any {
			t := startSig.Get()
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		})
		registerDocControlValue(ctx, id+"_end", func() any {
			t := endSig.Get()
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		})
		applyDocLayout(&drp.Element, block)
		return drp, nil
	case "pagination":
		total := int(block.Max)
		if total <= 0 {
			total = int(docBlockFloat(block, "totalPages", 1))
		}
		if total <= 0 {
			total = 1
		}
		if total > 9 {
			total = 9
		}
		page := int(docBlockNumericValue(block, 0))
		if page > 0 {
			page--
		}
		if page < 0 {
			page = 0
		}
		if page >= total {
			page = total - 1
		}
		cur := NewSignal(page)
		registerDocControlValue(ctx, id, func() any { return cur.Get() })
		pg := NewPagination(id, total, cur, 0, 0, docBlockWidth(block), block.Height)
		applyDocLayout(&pg.Element, block)
		return pg, nil
	case "toolbar":
		return buildDocToolbar(block, ctx, path)
	case "searchBar", "searchbar":
		return buildDocSearchBar(block, ctx, path)
	case "tabView", "tabview":
		return buildDocTabView(block, ctx, path)
	case "table":
		return buildDocTableCard(block, ctx, path)
	case "dataTable", "datatable":
		return buildDocDataTable(block, ctx, path)
	default:
		return nil, docBlockError(block, path, "unsupported block type %q", block.Type)
	}
}

func registerDocControlValue(ctx *BuildContext, id string, getter func() any) {
	if ctx != nil {
		ctx.RegisterControlValue(id, getter)
	}
}

func addDocChildren(parent Node, children []DocBlock, ctx *BuildContext, parentPath string) error {
	for i := range children {
		child, err := buildDocBlockAt(children[i], ctx, fmt.Sprintf("%s.children[%d]", parentPath, i))
		if err != nil {
			return err
		}
		parent.AddChild(child)
	}
	return nil
}

func applyDocRichText(rt *RichText, block DocBlock, ctx *BuildContext) {
	rt.Selectable = docBlockSelectable(block, true)
	if ctx != nil && ctx.LinkHandler != nil {
		rt.OnLinkClick = ctx.LinkHandler
	}
	if ctx != nil && ctx.PreviewTypography {
		if rt.styleName == "richtext" || rt.styleName == "" {
			rt.SetStyle("richtext-preview")
		}
	}
}

func docBlockLabeledControl(block DocBlock, control Node) Node {
	label := docBlockLabel(block)
	if label == "" {
		return control
	}
	controlID := control.ID()
	controlH := control.Bounds().Height
	row := NewContainer(controlID+"-row", 0, 0, block.Width, 0)
	row.SetStyle("transparent")
	row.FlexDirection = FlexRow
	row.Gap = 0
	row.AutoHeight = true
	leftInset := NewContainer(controlID+"-left-inset", 0, 0, 4, controlH)
	leftInset.SetStyle("transparent")
	row.AddChild(leftInset)
	row.AddChild(control)
	labelInset := NewContainer(controlID+"-label-inset", 0, 0, docBlockGap(block, 10), controlH)
	labelInset.SetStyle("transparent")
	row.AddChild(labelInset)
	lbl := NewPlainText(controlID+"-label", "form-value", label, 0, 0, 0, 0)
	lbl.SetFlexGrow(1)
	row.AddChild(lbl)

	wrap := NewContainer(controlID+"-wrap", 0, 0, block.Width, 0)
	wrap.SetStyle("transparent")
	wrap.FlexDirection = FlexColumn
	wrap.Gap = 0
	wrap.AddChild(row)
	applyElementPadding(&wrap.Element, 4)
	applyDocLayout(&wrap.Element, block)
	return wrap
}

func docBlockStackedLabeledControl(block DocBlock, control Node) Node {
	label := docBlockLabel(block)
	if label == "" {
		return control
	}
	controlID := control.ID()
	col := NewContainer(controlID+"-field", 0, 0, block.Width, 0)
	col.SetStyle("transparent")
	col.FlexDirection = FlexColumn
	col.Gap = docBlockGap(block, 8)
	col.AddChild(docBlockFieldCaption(controlID+"-label", label))
	col.AddChild(control)
	applyDocLayout(&col.Element, block)
	return col
}

// docBlockFieldCaption is the field title above controls (e.g. "Density" above a
// radio group, "Name" above an input). Uses the dedicated form-field-caption
// theme key so captions are visibly larger than inline control labels.
func docBlockFieldCaption(id, text string) Node {
	return NewPlainText(id, "form-field-caption", text, 0, 0, 0, 0)
}

func validateDocBlock(block DocBlock, path string) error {
	if block.Width < 0 {
		return docBlockError(block, path, "width must be >= 0")
	}
	if block.Height < 0 {
		return docBlockError(block, path, "height must be >= 0")
	}
	if block.Gap.Float() < 0 {
		return docBlockError(block, path, "gap must be >= 0")
	}
	if block.Padding.Float() < 0 {
		return docBlockError(block, path, "padding must be >= 0")
	}
	if block.MinWidth < 0 {
		return docBlockError(block, path, "minWidth must be >= 0")
	}
	if block.MaxWidth < 0 {
		return docBlockError(block, path, "maxWidth must be >= 0")
	}
	if block.MinWidth > 0 && block.MaxWidth > 0 && block.MinWidth > block.MaxWidth {
		return docBlockError(block, path, "minWidth must be <= maxWidth")
	}
	if block.Style != nil {
		if _, err := block.Style.compile(); err != nil {
			return docBlockError(block, path, "style: %v", err)
		}
	}
	if block.Preset != "" {
		if _, ok := LookupPreset(block.Preset); !ok {
			return docBlockError(block, path, "unknown preset %q", block.Preset)
		}
	}
	if caps := blockEffectiveCapabilities(block); caps != nil {
		if err := validateDocCapabilities(caps, block, path); err != nil {
			return err
		}
	}
	if block.Type == "surface" {
		if block.Variant != "" {
			if _, err := resolveSurfaceDocKind(block); err != nil {
				return docBlockError(block, path, "%s", err.Error())
			}
		}
	}
	if block.Type == "dropdown" {
		options := docBlockOptions(block)
		if len(options) == 0 {
			return docBlockError(block, path, "dropdown options must not be empty")
		}
		selected := docBlockSelectedIndex(block)
		if selected < 0 || selected >= len(options) {
			return docBlockError(block, path, "selectedIndex must be between 0 and %d", len(options)-1)
		}
	}
	if block.Type == "input" {
		if block.Value != nil {
			if _, ok := block.Value.(string); !ok {
				return docBlockError(block, path, "value must be a string")
			}
		}
	}
	if block.Type == "radioGroup" || block.Type == "radio" {
		options := docBlockOptions(block)
		if len(options) == 0 {
			return docBlockError(block, path, "radioGroup options must not be empty")
		}
		selected := docBlockSelectedIndex(block)
		if selected < -1 || selected >= len(options) {
			return docBlockError(block, path, "selectedIndex must be between -1 and %d", len(options)-1)
		}
	}
	if block.Type == "slider" || block.Type == "progressBar" || block.Type == "progress" {
		minVal, maxVal := docBlockRange(block, 0, 100)
		if minVal >= maxVal {
			return docBlockError(block, path, "min must be < max")
		}
		if block.Value != nil && !docBlockValueIsNumber(block.Value) {
			return docBlockError(block, path, "value must be a number")
		}
	}
	if block.Type == "rating" {
		if block.Value != nil && !docBlockValueIsNumber(block.Value) {
			return docBlockError(block, path, "value must be a number")
		}
	}
	return nil
}

func docBlockError(block DocBlock, path, format string, args ...any) error {
	detail := fmt.Sprintf(format, args...)
	typeName := block.Type
	if typeName == "" {
		typeName = "page"
	}
	if block.ID != "" {
		return fmt.Errorf("ui/document_spec: %s (id %q, type %q): %s", path, block.ID, typeName, detail)
	}
	return fmt.Errorf("ui/document_spec: %s (type %q): %s", path, typeName, detail)
}

func docBlockIntSliceProp(block DocBlock, key string) []int {
	if block.Props == nil {
		return nil
	}
	raw, ok := block.Props[key]
	if !ok {
		return nil
	}
	switch arr := raw.(type) {
	case []int:
		return arr
	case []any:
		out := make([]int, len(arr))
		for i, v := range arr {
			switch n := v.(type) {
			case int:
				out[i] = n
			case float64:
				out[i] = int(n)
			}
		}
		return out
	default:
		return nil
	}
}

func docBlockTextSpanSliceProp(block DocBlock, key string) [][]TextSpan {
	if block.Props == nil {
		return nil
	}
	raw, ok := block.Props[key]
	if !ok {
		return nil
	}
	switch arr := raw.(type) {
	case [][]TextSpan:
		return arr
	case []any:
		out := make([][]TextSpan, len(arr))
		for i, v := range arr {
			if sp, ok := v.([]TextSpan); ok {
				out[i] = sp
			}
		}
		return out
	default:
		return nil
	}
}

func docBlockBoolSliceProp(block DocBlock, key string) []bool {
	if block.Props == nil {
		return nil
	}
	raw, ok := block.Props[key]
	if !ok {
		return nil
	}
	switch arr := raw.(type) {
	case []bool:
		return arr
	case []any:
		out := make([]bool, len(arr))
		for i, v := range arr {
			if b, ok := v.(bool); ok {
				out[i] = b
			}
		}
		return out
	default:
		return nil
	}
}

func applyDocStyle(e *Element, block DocBlock) {
	if block.Preset != "" {
		props := PresetPropsFromMap(block.Props)
		if err := e.SetStylePreset(block.Preset, props); err != nil {
			// validateDocBlock should reject unknown presets before build.
			return
		}
		return
	}
	if block.Component != "" {
		e.SetStyleVariant(block.Component, block.Variant)
	} else if block.Variant != "" {
		e.SetVariant(block.Variant)
	}
}

// applyCardRecipeVariant pins callout/code (and similar) card chrome from Theme v2.
// SetStyleOverrides keeps tint visible even if variant resolution falls back to legacy "card".
func applyCardRecipeVariant(card *Card, variant string) {
	card.SetStyleVariant("card", variant)
	if st, ok := MergedComponentVariantStyle("card", variant); ok {
		card.SetStyleOverrides(st)
	}
}

// configureDocMarkdownCard applies minimal preview chrome: no title bar when title is
// empty, tight body gap, and lite Theme v2 padding (callout, code, table).
func configureDocMarkdownCard(card *Card, variant, title string) {
	if title == "" {
		card.Title = ""
		card.TitleHeight = 0
	} else {
		card.Title = title
	}
	card.Gap = 4
	card.AutoHeight = true
	applyCardRecipeVariant(card, variant)
	if st, ok := MergedComponentVariantStyle("card", variant); ok {
		st.Padding = DocMarkdownCardPadding
		card.SetStyleOverrides(st)
	}
}

func applyDocLayout(e *Element, block DocBlock) {
	if w := docBlockWidth(block); w > 0 {
		e.PreferredWidth = w
	}
	if minW := docBlockMinWidth(block); minW > 0 {
		e.MinWidth = minW
	}
	if maxW := docBlockMaxWidth(block); maxW > 0 {
		e.MaxWidth = maxW
	}
	if block.FlexGrow != nil && *block.FlexGrow >= 0 {
		e.SetFlexGrow(*block.FlexGrow)
	}
	applyDocBlockStyleOverrides(e, block)
	if padding, ok := docBlockPaddingOverride(block); ok {
		applyElementPadding(e, padding)
	}
}

func applyDocContainerFlags(c *Container, block DocBlock) {
	if block.Wrap != nil {
		c.SetFlexWrap(*block.Wrap)
	}
	if block.ClipChildren != nil {
		c.ClipChildren = *block.ClipChildren
	}
}

// applyCardChromeTextToBody syncs body text color from card chrome to Label/RichText
// children. Explicit block style textColor wins when provided.
func applyCardChromeTextToBody(card *Card, block DocBlock) {
	if block.Style != nil {
		st, err := block.Style.compile()
		if err == nil && (st.TextColor.A != 0 || st.FontSize > 0) {
			hint := bodyTypographyHintFromChrome(st)
			if st.TextColor.A != 0 {
				hint.TextColor = st.TextColor
			}
			if st.FontSize > 0 {
				hint.FontSize = st.FontSize
			}
			for _, ch := range card.Children() {
				applySurfaceBodyTypography(ch, hint, chromeStyleIsDark(card.GetStyle()))
			}
			return
		}
	}
	card.ApplyCardBodyTextColor()
}

func applyDocBlockStyleOverrides(e *Element, block DocBlock) {
	if block.Style == nil {
		return
	}
	e.mergeStylePatch(styleJSON(*block.Style))
}

func applyElementPadding(e *Element, padding float32) {
	p := padding
	e.mergeStylePatch(styleJSON{Padding: &p})
}

func docBadgeVariant(variant string) BadgeVariant {
	switch strings.ToLower(strings.TrimSpace(variant)) {
	case "primary":
		return BadgePrimary
	case "success":
		return BadgeSuccess
	case "warning":
		return BadgeWarning
	case "danger":
		return BadgeDanger
	case "info":
		return BadgeInfo
	default:
		return BadgeDefault
	}
}

func docBlockGap(block DocBlock, fallback float32) float32 {
	if v := block.Gap.Float(); v > 0 {
		return v
	}
	if v := docBlockSpacingFromProps(block, "gap", 0); v > 0 {
		return v
	}
	return docBlockFloat(block, "gap", fallback)
}

func docBlockWidth(block DocBlock) float32 {
	if block.Width > 0 {
		return block.Width
	}
	if w := docBlockFloat(block, "width", 0); w > 0 {
		return w
	}
	return 0
}

func docBlockWidthFromCtx(block DocBlock, ctx *BuildContext) float32 {
	if w := docBlockWidth(block); w > 0 {
		return w
	}
	if ctx != nil && ctx.ContentWidth > 0 {
		return ctx.ContentWidth
	}
	return 0
}

func docBlockMinWidth(block DocBlock) float32 {
	if block.MinWidth > 0 {
		return block.MinWidth
	}
	return docBlockFloat(block, "minWidth", 0)
}

func docBlockMaxWidth(block DocBlock) float32 {
	if block.MaxWidth > 0 {
		return block.MaxWidth
	}
	return docBlockFloat(block, "maxWidth", 0)
}

func docBlockPadding(block DocBlock) float32 {
	if v := block.Padding.Float(); v > 0 {
		return v
	}
	if v := docBlockSpacingFromProps(block, "padding", 0); v > 0 {
		return v
	}
	return docBlockFloat(block, "padding", 0)
}

// docBlockPaddingOverride reports an explicit padding on the block (including 0).
func docBlockPaddingOverride(block DocBlock) (float32, bool) {
	if v := block.Padding.Float(); v > 0 {
		return v, true
	}
	if block.Props != nil {
		if _, ok := block.Props["padding"]; ok {
			return docBlockSpacingFromProps(block, "padding", 0), true
		}
	}
	return 0, false
}

func docBlockSelectable(block DocBlock, fallback bool) bool {
	if block.Selectable != nil {
		return *block.Selectable
	}
	return docBlockBool(block, "selectable", fallback)
}

func docBlockOnClick(block DocBlock) string {
	if block.OnClick != "" {
		return block.OnClick
	}
	if block.Props == nil {
		return ""
	}
	actionName, _ := block.Props["onClick"].(string)
	return actionName
}

func docBlockStringProp(block DocBlock, key string) string {
	if block.Props == nil {
		return ""
	}
	s, _ := block.Props[key].(string)
	return s
}

func docBlockStringSliceProp(block DocBlock, key string) []string {
	if block.Props == nil {
		return nil
	}
	raw, ok := block.Props[key]
	if !ok {
		return nil
	}
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, e := range v {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}

func docBlockBottomNavItems(block DocBlock) []BottomNavItem {
	labels := block.Items
	if len(labels) == 0 {
		return []BottomNavItem{{Icon: "·", Label: "-"}}
	}
	icons := docBlockStringSliceProp(block, "icons")
	badges := docBlockStringSliceProp(block, "badges")
	out := make([]BottomNavItem, len(labels))
	for i, label := range labels {
		item := BottomNavItem{Label: label}
		if i < len(icons) {
			item.Icon = icons[i]
		}
		if item.Icon == "" {
			item.Icon = "·"
		}
		if i < len(badges) {
			item.Badge = badges[i]
		}
		out[i] = item
	}
	return out
}

func docBlockLabel(block DocBlock) string {
	if block.Label != "" {
		return block.Label
	}
	if block.Text != "" {
		return block.Text
	}
	if block.Title != "" {
		return block.Title
	}
	if block.Props == nil {
		return ""
	}
	label, _ := block.Props["label"].(string)
	return label
}

func docBlockChecked(block DocBlock) bool {
	if block.Checked {
		return true
	}
	return docBlockBool(block, "checked", false)
}

func docBlockControlSize(block DocBlock, fallbackW, fallbackH float32) (float32, float32) {
	w := block.Width
	if w == 0 {
		w = docBlockFloat(block, "width", fallbackW)
	}
	if w == 0 {
		w = fallbackW
	}
	h := block.Height
	if h == 0 {
		h = docBlockFloat(block, "height", fallbackH)
	}
	if h == 0 {
		h = fallbackH
	}
	return w, h
}

func docBlockVertical(block DocBlock, fallback bool) bool {
	if block.Vertical != nil {
		return *block.Vertical
	}
	return docBlockBool(block, "vertical", fallback)
}

func docBlockRowHeight(block DocBlock) float32 {
	if block.RowHeight > 0 {
		return block.RowHeight
	}
	return docBlockFloat(block, "rowHeight", 32)
}

func docBlockRange(block DocBlock, fallbackMin, fallbackMax float32) (float32, float32) {
	minVal := block.Min
	if minVal == 0 {
		minVal = docBlockFloat(block, "min", fallbackMin)
	}
	maxVal := block.Max
	if maxVal == 0 {
		maxVal = docBlockFloat(block, "max", fallbackMax)
	}
	return minVal, maxVal
}

func docBlockNumericValue(block DocBlock, fallback float32) float32 {
	switch v := block.Value.(type) {
	case float32:
		return v
	case float64:
		return float32(v)
	case int:
		return float32(v)
	default:
		return docBlockFloat(block, "value", fallback)
	}
}

func docBlockValueIsNumber(value any) bool {
	switch value.(type) {
	case float32, float64, int, int32, int64:
		return true
	default:
		return false
	}
}

func docBlockValue(block DocBlock) string {
	if value, ok := block.Value.(string); ok && value != "" {
		return value
	}
	if block.Props == nil {
		return ""
	}
	value, _ := block.Props["value"].(string)
	return value
}

func docBlockPlaceholder(block DocBlock) string {
	if block.Placeholder != "" {
		return block.Placeholder
	}
	if block.Props == nil {
		return ""
	}
	placeholder, _ := block.Props["placeholder"].(string)
	return placeholder
}

func docBlockOptions(block DocBlock) []string {
	if len(block.Options) > 0 {
		return block.Options
	}
	if block.Props == nil {
		return nil
	}
	raw, ok := block.Props["options"]
	if !ok {
		return nil
	}
	switch opts := raw.(type) {
	case []string:
		return opts
	case []any:
		out := make([]string, 0, len(opts))
		for _, item := range opts {
			text, ok := item.(string)
			if !ok {
				return nil
			}
			out = append(out, text)
		}
		return out
	default:
		return nil
	}
}

func docBlockSelectedIndex(block DocBlock) int {
	if block.SelectedIndex != 0 {
		return block.SelectedIndex
	}
	if block.Props == nil {
		return 0
	}
	switch v := block.Props["selectedIndex"].(type) {
	case float64:
		return int(v)
	case int:
		return v
	default:
		return 0
	}
}

func docBlockFloat(block DocBlock, key string, fallback float32) float32 {
	if block.Props == nil {
		return fallback
	}
	switch v := block.Props[key].(type) {
	case float32:
		return v
	case float64:
		return float32(v)
	case int:
		return float32(v)
	case string:
		if px, err := ResolveSpacingToken(v); err == nil {
			return px
		}
		value, err := strconv.ParseFloat(v, 32)
		if err != nil {
			return fallback
		}
		return float32(value)
	default:
		return fallback
	}
}

func docBlockBool(block DocBlock, key string, fallback bool) bool {
	if block.Props == nil {
		return fallback
	}
	v, ok := block.Props[key].(bool)
	if !ok {
		return fallback
	}
	return v
}

// docBlockParseDate parses YYYY-MM-DD from DocumentSpec props; zero time on empty/invalid.
func docBlockParseDate(s string) time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}
	}
	t, err := time.ParseInLocation("2006-01-02", s, time.Local)
	if err != nil {
		return time.Time{}
	}
	return dateTruncLocal(t)
}
