// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"encoding/json"
	"fmt"
)

// JSONNode is the JSON schema for a single UI node.
// All fields are optional; sensible defaults are applied per type.
//
// Supported types: viewport, panel, container, button, label, textinput,
// slider, checkbox, toggle, dropdown, progressbar, virtuallist.
type JSONNode struct {
	// Core identity
	Type  string `json:"type"` // Required widget type name.
	ID    string `json:"id"`
	Style string `json:"style"` // Named style from CurrentTheme.
	// Theme v2 fields. Component+Variant is preferred for new JSON; Style
	// remains the legacy flat theme key and fallback component name.
	Component string `json:"component"`
	Variant   string `json:"variant"`
	Overrides *Style `json:"overrides"`

	// Geometry
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Width  float32 `json:"width"`
	Height float32 `json:"height"`

	// Layout (container, viewport, panel)
	Direction string  `json:"direction"` // "row" or "column" (default)
	Layout    string  `json:"layout"`    // "grid" overrides flex
	Gap       float32 `json:"gap"`

	// Text-bearing widgets
	Text        string `json:"text"`        // Button label, Label text
	Placeholder string `json:"placeholder"` // TextInput hint

	// Panel
	Title       string  `json:"title"`
	TitleHeight float32 `json:"titleHeight"`

	// Numeric widgets
	Min   float32 `json:"min"`
	Max   float32 `json:"max"`
	Value float32 `json:"value"` // Slider initial value or ProgressBar fill

	// Boolean widgets
	Checked bool `json:"checked"` // Checkbox, Toggle initial state

	// Dropdown
	Options       []string `json:"options"`
	SelectedIndex int      `json:"selectedIndex"`

	// VirtualList
	ItemHeight float32  `json:"itemHeight"`
	Items      []string `json:"items"` // Static item list

	// Data binding — keys into BuildContext maps
	BindingKey string `json:"bindingKey"`

	// Interaction
	OnClick string `json:"onClick"` // Key into BuildContext.Actions

	// Spinner
	SpinnerSize  float32 `json:"spinnerSize"`  // diameter in px (default 40)
	SpinnerSpeed float32 `json:"spinnerSpeed"` // degrees/second (default 360)
	Active       bool    `json:"active"`       // Spinner: initial active state

	// RadioGroup
	// Options is reused from Dropdown above.
	Vertical  bool    `json:"vertical"`  // RadioGroup layout direction
	RadioRowH float32 `json:"radioRowH"` // row height per option (default 36)

	// TabView — tabs are specified as child JSONNodes.
	// Each child's "text" field becomes the tab title; its own Children are
	// placed in a Container used as the tab's content node.
	TabBarH float32 `json:"tabBarH"` // override tab bar height (default 34)

	// TreeView — items are expressed using the standard Children recursion.
	// Set type="treeview" and nest children using type="treeitem".
	// A treeitem's "text" becomes the label; its Children are sub-items.
	TreeExpanded bool `json:"treeExpanded"` // initial expanded state for treeitem

	// Tree (standard recursive Children)
	Children []JSONNode `json:"children"`
}

// BuildContext carries the runtime dependencies needed when instantiating
// JSON-defined UI trees. Pass a populated BuildContext to
// LoadFromJSONWithContext so widgets referencing bindingKey or onClick can
// connect to live application state.
type BuildContext struct {
	// Actions maps onClick name strings to handler functions.
	Actions map[string]func()
	// TextSignals maps bindingKey strings to *Signal[string] for Label/TextInput.
	TextSignals map[string]*Signal[string]
	// ListBindings maps bindingKey strings to *ListBinding[string] for VirtualList.
	ListBindings map[string]*ListBinding[string]
	// ValueSignals maps bindingKey strings to *Signal[float32] for Slider/ProgressBar.
	ValueSignals map[string]*Signal[float32]
	// BoolSignals maps bindingKey strings to *Signal[bool] for Checkbox/Toggle.
	BoolSignals map[string]*Signal[bool]
	// Float64Signals maps binding keys to *Signal[float64] for ToolbarSpinBox and similar.
	Float64Signals map[string]*Signal[float64]
	// ControlValues maps generated control IDs to getter functions. DocumentSpec
	// controls register here so app code can read current form values by block ID.
	ControlValues map[string]func() any
	// LinkHandler receives RichText link clicks from generated document content.
	LinkHandler func(link string)
	// PreviewTypography uses Inter (when loaded) for generated RichText in markdown preview.
	PreviewTypography bool
	// SyntaxHighlight enables Chroma on code blocks and inline-code pill styling in
	// DocumentSpec trees. nil = enabled. Set false on ctx or a block to disable subtree.
	SyntaxHighlight *bool
	// ContentWidth is the available width for document blocks (markdown preview lane).
	ContentWidth float32
}

// NewBuildContext returns an empty BuildContext with all maps initialised.
func NewBuildContext() *BuildContext {
	return &BuildContext{
		Actions:       make(map[string]func()),
		TextSignals:   make(map[string]*Signal[string]),
		ListBindings:  make(map[string]*ListBinding[string]),
		ValueSignals:  make(map[string]*Signal[float32]),
		BoolSignals:    make(map[string]*Signal[bool]),
		Float64Signals: make(map[string]*Signal[float64]),
		ControlValues: make(map[string]func() any),
	}
}

// RegisterControlValue exposes a generated control's current value by ID.
func (ctx *BuildContext) RegisterControlValue(id string, getter func() any) {
	if ctx == nil || id == "" || getter == nil {
		return
	}
	if ctx.ControlValues == nil {
		ctx.ControlValues = make(map[string]func() any)
	}
	ctx.ControlValues[id] = getter
}

// ControlValue returns the latest value for a generated control ID.
func (ctx *BuildContext) ControlValue(id string) (any, bool) {
	if ctx == nil || ctx.ControlValues == nil {
		return nil, false
	}
	getter, ok := ctx.ControlValues[id]
	if !ok || getter == nil {
		return nil, false
	}
	return getter(), true
}

// ControlSnapshot returns a point-in-time copy of all generated control values.
func (ctx *BuildContext) ControlSnapshot() map[string]any {
	out := map[string]any{}
	if ctx == nil || ctx.ControlValues == nil {
		return out
	}
	for id, getter := range ctx.ControlValues {
		if getter != nil {
			out[id] = getter()
		}
	}
	return out
}

// ApplyControlSnapshot restores DocumentSpec control values on a freshly built
// tree. IDs missing from the new tree or with incompatible types are skipped.
// Dropdown and radio values are option label strings (same as ControlSnapshot).
func ApplyControlSnapshot(root Node, snap map[string]any) {
	if root == nil || len(snap) == 0 {
		return
	}
	for id, val := range snap {
		if val == nil {
			continue
		}
		n := FindNodeByID(root, id)
		if n == nil {
			continue
		}
		applyControlValue(n, val)
	}
}

func applyControlValue(n Node, val any) {
	switch w := n.(type) {
	case *TextInput:
		if s, ok := val.(string); ok {
			w.Text.Set(s)
		}
	case *Dropdown:
		label, ok := val.(string)
		if !ok {
			return
		}
		for i, opt := range w.Options {
			if opt == label {
				w.SelectedIndex.Set(i)
				return
			}
		}
	case *Checkbox:
		if b, ok := val.(bool); ok {
			w.Value.Set(b)
		}
	case *Toggle:
		if b, ok := val.(bool); ok {
			w.Value.Set(b)
		}
	case *RadioGroup:
		label, ok := val.(string)
		if !ok {
			return
		}
		for i, opt := range w.Options {
			if opt == label {
				w.Selected.Set(i)
				return
			}
		}
	case *Slider:
		switch v := val.(type) {
		case float32:
			w.Value.Set(v)
		case float64:
			w.Value.Set(float32(v))
		case int:
			w.Value.Set(float32(v))
		}
	}
}

// LoadFromJSON parses data and appends the top-level node to the document root.
// Use LoadFromJSONWithContext when you need action callbacks or data bindings.
func (d *Document) LoadFromJSON(data []byte) error {
	return d.LoadFromJSONWithContext(data, NewBuildContext())
}

// LoadFromJSONWithContext parses data using ctx for callbacks and bindings,
// then appends the resulting node to d.Root.
func (d *Document) LoadFromJSONWithContext(data []byte, ctx *BuildContext) error {
	var root JSONNode
	if err := json.Unmarshal(data, &root); err != nil {
		return fmt.Errorf("ui/json: parse error: %w", err)
	}
	node, err := buildNodeFromJSON(&root, ctx)
	if err != nil {
		return err
	}
	d.Add(node)
	return nil
}

// buildNodeFromJSON converts a JSONNode tree into live Nodes recursively.
func buildNodeFromJSON(j *JSONNode, ctx *BuildContext) (Node, error) {
	if j.ID == "" {
		j.ID = j.Type + "-auto"
	}
	switch j.Type {
	case "viewport":
		return jsonViewport(j, ctx)
	case "panel":
		return jsonPanel(j, ctx)
	case "container":
		return jsonContainer(j, ctx)
	case "button":
		return jsonButton(j, ctx), nil
	case "label":
		return jsonLabel(j, ctx), nil
	case "textinput":
		return jsonTextInput(j, ctx), nil
	case "slider":
		return jsonSlider(j, ctx), nil
	case "checkbox":
		return jsonCheckbox(j, ctx), nil
	case "toggle":
		return jsonToggle(j, ctx), nil
	case "dropdown":
		return jsonDropdown(j, ctx), nil
	case "progressbar":
		return jsonProgressBar(j, ctx), nil
	case "virtuallist":
		return jsonVirtualList(j, ctx), nil
	case "tabview":
		return jsonTabView(j, ctx)
	case "spinner":
		return jsonSpinner(j, ctx)
	case "radiogroup":
		return jsonRadioGroup(j, ctx)
	case "treeview":
		return jsonTreeView(j, ctx)
	default:
		lbl := NewLabel(j.ID, fmt.Sprintf("[unknown widget: %q]", j.Type), j.X, j.Y, j.Width, j.Height)
		return lbl, nil
	}
}

// jsonChildren builds all children of j and returns them.
func jsonChildren(j *JSONNode, ctx *BuildContext) ([]Node, error) {
	out := make([]Node, 0, len(j.Children))
	for i := range j.Children {
		child, err := buildNodeFromJSON(&j.Children[i], ctx)
		if err != nil {
			return nil, fmt.Errorf("child[%d] of %q: %w", i, j.ID, err)
		}
		out = append(out, child)
	}
	return out, nil
}

// applyFlexDir sets direction and gap on a Container from a JSONNode.
func applyFlexDir(c *Container, j *JSONNode) {
	if j.Direction == "row" {
		c.FlexDirection = FlexRow
	} else {
		c.FlexDirection = FlexColumn
	}
	if j.Gap > 0 {
		c.Gap = j.Gap
	}
}

type jsonStyleTarget interface {
	SetStyle(string)
	SetStyleVariant(string, string)
	SetStyleOverrides(Style)
}

func applyJSONStyle(n jsonStyleTarget, j *JSONNode) {
	if n == nil || j == nil {
		return
	}
	if j.Component != "" || j.Variant != "" {
		component := j.Component
		if component == "" {
			component = j.Style
		}
		if component == "" {
			component = j.Type
		}
		n.SetStyleVariant(component, j.Variant)
	} else if j.Style != "" {
		n.SetStyle(j.Style)
	}
	if j.Overrides != nil {
		n.SetStyleOverrides(*j.Overrides)
	}
}

func jsonViewport(j *JSONNode, ctx *BuildContext) (*Viewport, error) {
	vp := NewViewport(j.ID, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(vp, j)
	applyFlexDir(&vp.Container, j)
	children, err := jsonChildren(j, ctx)
	if err != nil {
		return nil, fmt.Errorf("viewport %q: %w", j.ID, err)
	}
	for _, child := range children {
		vp.AddChild(child)
	}
	return vp, nil
}

func jsonPanel(j *JSONNode, ctx *BuildContext) (*Panel, error) {
	p := NewPanel(j.ID, j.Title, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(p, j)
	if j.TitleHeight > 0 {
		p.TitleHeight = j.TitleHeight
	}
	if j.Gap > 0 {
		p.Gap = j.Gap
	}
	children, err := jsonChildren(j, ctx)
	if err != nil {
		return nil, fmt.Errorf("panel %q: %w", j.ID, err)
	}
	for _, child := range children {
		p.AddChild(child)
	}
	return p, nil
}

func jsonContainer(j *JSONNode, ctx *BuildContext) (*Container, error) {
	c := NewContainer(j.ID, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(c, j)
	if j.Layout == "grid" {
		c.LayoutType = LayoutGrid
	}
	applyFlexDir(c, j)
	children, err := jsonChildren(j, ctx)
	if err != nil {
		return nil, fmt.Errorf("container %q: %w", j.ID, err)
	}
	for _, child := range children {
		c.AddChild(child)
	}
	return c, nil
}

func jsonButton(j *JSONNode, ctx *BuildContext) *Button {
	btn := NewButton(j.ID, j.Text, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(btn, j)
	if j.OnClick != "" && ctx != nil {
		if fn, ok := ctx.Actions[j.OnClick]; ok {
			btn.OnClick = fn
		}
	}
	return btn
}

func jsonLabel(j *JSONNode, ctx *BuildContext) *Label {
	lbl := NewLabel(j.ID, j.Text, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(lbl, j)
	if j.BindingKey != "" && ctx != nil {
		if sig, ok := ctx.TextSignals[j.BindingKey]; ok {
			NewEffect(func() { lbl.Text.Set(sig.Get()) })
		}
	}
	return lbl
}

func jsonTextInput(j *JSONNode, ctx *BuildContext) *TextInput {
	hint := j.Placeholder
	if hint == "" {
		hint = j.Text
	}
	inp := NewTextInput(j.ID, hint, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(inp, j)
	if j.BindingKey != "" && ctx != nil {
		if sig, ok := ctx.TextSignals[j.BindingKey]; ok {
			// Widget → binding
			inp.Text.Subscribe(func() { sig.Set(inp.Text.Get()) })
			// Binding → widget (external writes)
			sig.Subscribe(func() {
				if sig.Get() != inp.Text.Get() {
					inp.Text.Set(sig.Get())
				}
			})
		}
	}
	return inp
}

func jsonSlider(j *JSONNode, ctx *BuildContext) *Slider {
	max := j.Max
	if max == 0 {
		max = 100
	}
	sl := NewSlider(j.ID, j.Min, max, j.Value, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(sl, j)
	if j.BindingKey != "" && ctx != nil {
		if sig, ok := ctx.ValueSignals[j.BindingKey]; ok {
			sl.Value.Subscribe(func() { sig.Set(sl.Value.Get()) })
			sig.Subscribe(func() {
				if sig.Get() != sl.Value.Get() {
					sl.Value.Set(sig.Get())
				}
			})
		}
	}
	return sl
}

func jsonCheckbox(j *JSONNode, ctx *BuildContext) *Checkbox {
	cb := NewCheckbox(j.ID, j.Checked, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(cb, j)
	if j.BindingKey != "" && ctx != nil {
		if sig, ok := ctx.BoolSignals[j.BindingKey]; ok {
			cb.Value.Subscribe(func() { sig.Set(cb.Value.Get()) })
			sig.Subscribe(func() {
				if sig.Get() != cb.Value.Get() {
					cb.Value.Set(sig.Get())
				}
			})
		}
	}
	return cb
}

func jsonToggle(j *JSONNode, ctx *BuildContext) *Toggle {
	tg := NewToggle(j.ID, j.Checked, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(tg, j)
	if j.BindingKey != "" && ctx != nil {
		if sig, ok := ctx.BoolSignals[j.BindingKey]; ok {
			tg.Value.Subscribe(func() { sig.Set(tg.Value.Get()) })
			sig.Subscribe(func() {
				if sig.Get() != tg.Value.Get() {
					tg.Value.Set(sig.Get())
				}
			})
		}
	}
	return tg
}

func jsonDropdown(j *JSONNode, _ *BuildContext) *Dropdown {
	opts := j.Options
	if len(opts) == 0 {
		opts = []string{"(empty)"}
	}
	dd := NewDropdown(j.ID, opts, j.SelectedIndex, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(dd, j)
	return dd
}

func jsonProgressBar(j *JSONNode, ctx *BuildContext) *ProgressBar {
	pb := NewProgressBar(j.ID, j.Value, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(pb, j)
	if j.BindingKey != "" && ctx != nil {
		if sig, ok := ctx.ValueSignals[j.BindingKey]; ok {
			NewEffect(func() { pb.Value.Set(sig.Get()) })
		}
	}
	return pb
}

func jsonVirtualList(j *JSONNode, ctx *BuildContext) *VirtualList[string] {
	itemH := j.ItemHeight
	if itemH <= 0 {
		itemH = 28
	}
	var binding *ListBinding[string]
	if j.BindingKey != "" && ctx != nil {
		if lb, ok := ctx.ListBindings[j.BindingKey]; ok {
			binding = lb
		}
	}
	if binding == nil {
		items := j.Items
		if items == nil {
			items = []string{}
		}
		binding = NewListBinding(items)
	}
	tmpl := func(item string, index int, isSelected bool) Node {
		lbl := NewLabel(fmt.Sprintf("%s-item-%d", j.ID, index), item, 0, 0, 0, itemH)
		if isSelected {
			lbl.SetStyle("list-selected")
		} else {
			lbl.SetStyle("default")
		}
		return lbl
	}
	vl := NewVirtualList(j.ID, binding, tmpl, itemH, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(vl, j)
	return vl
}

// ── Batch 1 JSON builders ─────────────────────────────────────────────────────

// jsonTabView builds a TabView from a JSONNode. Each child JSONNode becomes a
// tab: the child's "text" field is the tab title and its own Children are
// placed in a flex-column Container used as the tab content.
func jsonTabView(j *JSONNode, ctx *BuildContext) (Node, error) {
	tv := NewTabView(j.ID, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(tv, j)
	for i := range j.Children {
		tab := &j.Children[i]
		title := tab.Text
		if title == "" {
			title = fmt.Sprintf("Tab %d", i+1)
		}
		// Build content container from tab's children
		content := NewContainer(fmt.Sprintf("%s-tab%d", j.ID, i), 0, 0, j.Width, j.Height-tabBarH)
		subChildren, err := jsonChildren(tab, ctx)
		if err != nil {
			return nil, fmt.Errorf("tabview %q tab[%d]: %w", j.ID, i, err)
		}
		for _, c := range subChildren {
			content.AddChild(c)
		}
		tv.AddTab(title, content)
	}
	return tv, nil
}

// jsonSpinner builds a Spinner from a JSONNode.
func jsonSpinner(j *JSONNode, _ *BuildContext) (Node, error) {
	size := j.SpinnerSize
	if size <= 0 {
		size = j.Width
	}
	if size <= 0 {
		size = 40
	}
	sp := NewSpinner(j.ID, j.X, j.Y, size)
	if j.SpinnerSpeed > 0 {
		sp.Speed = j.SpinnerSpeed
	}
	sp.Active.Set(j.Active)
	applyJSONStyle(sp, j)
	return sp, nil
}

// jsonRadioGroup builds a RadioGroup from a JSONNode.
// Options are taken from the "options" field (same as Dropdown).
func jsonRadioGroup(j *JSONNode, _ *BuildContext) (Node, error) {
	opts := j.Options
	if len(opts) == 0 {
		opts = []string{"Option A", "Option B"}
	}
	rg := NewRadioGroup(j.ID, opts, j.X, j.Y, j.Width, j.Height)
	if j.Vertical || (!j.Vertical && j.RadioRowH == 0) {
		rg.Vertical = true
	} else {
		rg.Vertical = j.Vertical
	}
	if j.RadioRowH > 0 {
		rg.RowH = j.RadioRowH
	}
	if j.SelectedIndex >= 0 && j.SelectedIndex < len(opts) {
		rg.Selected.Set(j.SelectedIndex)
	}
	applyJSONStyle(rg, j)
	return rg, nil
}

// jsonBuildTree recursively converts JSONNode trees rooted at a "treeitem"
// into *TreeNode data nodes.
func jsonBuildTree(j *JSONNode) *TreeNode {
	n := NewTreeNode(j.ID, j.Text)
	n.expanded = j.TreeExpanded
	for i := range j.Children {
		child := jsonBuildTree(&j.Children[i])
		n.Children = append(n.Children, child)
	}
	return n
}

// jsonTreeView builds a TreeView from a JSONNode. Top-level children become
// children of a synthetic root node labelled "(root)".
func jsonTreeView(j *JSONNode, _ *BuildContext) (Node, error) {
	root := NewTreeNode("root", "(root)")
	root.expanded = true
	for i := range j.Children {
		root.Children = append(root.Children, jsonBuildTree(&j.Children[i]))
	}
	tv := NewTreeView(j.ID, root, j.X, j.Y, j.Width, j.Height)
	applyJSONStyle(tv, j)
	return tv, nil
}

// LoadTheme replaces CurrentTheme with the theme loaded from JSON data.
func LoadTheme(data []byte) error {
	var t Theme
	if err := json.Unmarshal(data, &t); err != nil {
		return fmt.Errorf("ui/json: theme parse error: %w", err)
	}
	CurrentTheme = t
	return nil
}
