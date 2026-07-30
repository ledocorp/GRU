// Package ui (continued) — Panel facade over SurfaceShell.
//
// See node.go and panel_features.go for collapse/move/resize options.
//
// Demo scenes: **Form Demo**, **Batch 2 SearchBar**, **Settings (DocumentSpec)**,
// **Widgets Demo**, most batch demos (panel-per-section pattern).
package ui

// Panel is a facade over SurfaceShell with a dark title-bar header (HeaderModeTitleBar).
//
// # LLM Prompt Template
//
//	p := ui.NewPanel("settings", "Settings", 0, 0, 0, 230)
//	p.SetFlexGrow(1)
//	p.AddChild(ui.NewLabel("name-label", "Name:", 0, 0, 120, 28))
//	vp.AddChild(p)
//
// For scrollable body content, PanelFeatures enables v-scroll (see panel_features.go).
// See examples/form_demo.go and examples/document_app_demo.go.
type Panel struct {
	SurfaceShell
}

// NewPanel creates a Panel with the given title and bounds.
func NewPanel(id, title string, x, y, w, h float32) *Panel {
	c := NewContainer(id, x, y, w, h)
	p := &Panel{
		SurfaceShell: SurfaceShell{
			Container:   *c,
			Title:       title,
			TitleHeight: 40,
			headerMode:  HeaderModeTitleBar,
		},
	}
	p.Gap = 12
	p.cachePolicy = CacheAuto
	p.attachBody(id)
	p.styleName = "panel"
	p.Element.SetStyleVariant("panel", "default")
	p.AttachBehavior(NewPanelFeaturesBehavior())
	return p
}

// AddChild attaches content to the body or scroll host; parent is the attach target
// so ancestor clip walks (viewport scissor) stay correct.
func (p *Panel) AddChild(child Node) {
	target := p.panelContentTarget()
	if vp, ok := target.(*Viewport); ok {
		vp.AddChild(child)
	} else if p.body != nil {
		p.body.AddChild(child)
	}
	child.SetParent(target)
	p.applySurfaceBodyTypographyToChild(child)
}

func (p *Panel) panelContentTarget() Node {
	return surfaceContentTarget(&p.SurfaceShell)
}

// ApplyPanelBodyTextColor syncs direct Label/RichText children to panel chrome.
func (p *Panel) ApplyPanelBodyTextColor() {
	p.applySurfaceBodyTypographyToChildren()
}

// EnableCollapse enables Collapsible and returns the collapse plugin.
func (p *Panel) EnableCollapse(initialExpanded bool) *CollapseBehavior {
	p.SetCollapsible(true)
	cb := p.CollapseBehavior()
	if cb != nil {
		cb.Expanded.Set(initialExpanded)
	}
	return cb
}

// SetHeaderMode sets the surface title band style (title bar, inset, glass, or none).
func (p *Panel) SetHeaderMode(mode SurfaceHeaderMode) {
	p.headerMode = mode
	p.MarkDirty()
}

// SetStylePreset applies a named visual preset to the panel chrome.
func (p *Panel) SetStylePreset(name string, props PresetProps) error {
	if err := p.Element.SetStylePreset(name, props); err != nil {
		return err
	}
	applyVisualSurfaceLayout(&p.TitleHeight, &p.Gap, name)
	p.ApplyPanelBodyTextColor()
	return nil
}
