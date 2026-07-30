// Package ui (continued) — egui-style programmable panel features (Phase C3–C4).
//
// PanelFeatures controls the whole panel: title bar, collapse, close, move, resize,
// scroll, anchor, and drag mode. Attach via NewPanel (default behavior) or mutate
// through Panel.Features() and call SyncFeatures().
//
// # LLM Prompt Template
//
//	p := ui.NewPanel("win", "Settings", 40, 40, 360, 280)
//	p.SetCollapsible(true).SetClosable(true).SetMovable(true).SetResizable(true)
//	p.SetDragMode(ui.PanelDragTitleBar)
//	p.SetVScroll(true)
package ui

// PanelAnchorH is horizontal anchor within the parent bounds.
type PanelAnchorH int

const (
	PanelAnchorLeft PanelAnchorH = iota
	PanelAnchorHCenter
	PanelAnchorRight
)

// PanelAnchorV is vertical anchor within the parent bounds.
type PanelAnchorV int

const (
	PanelAnchorTop PanelAnchorV = iota
	PanelAnchorVCenter
	PanelAnchorBottom
)

// PanelDragMode selects where a drag gesture moves the panel (requires Movable).
type PanelDragMode int

const (
	PanelDragOff PanelDragMode = iota
	// PanelDragOnTouch — drag from the header/chrome row (not body content).
	PanelDragOnTouch
	// PanelDragTitleBar — drag from the title band only.
	PanelDragTitleBar
	// PanelDragAnywhere — drag from anywhere on the panel shell (except controls).
	PanelDragAnywhere
)

// PanelFeatures is the egui-style capability set for a Panel shell.
type PanelFeatures struct {
	TitleBar    bool
	Collapsible bool
	Closable    bool
	Movable     bool
	Resizable   bool
	Constrain   bool
	HScroll     bool
	VScroll     bool
	Anchored    bool
	AnchorX     PanelAnchorH
	AnchorY     PanelAnchorV
	DragMode    PanelDragMode
	OnDismiss   func()
	// FloatPosition places the panel at FloatX/FloatY within the parent (absolute layout).
	FloatPosition bool
	FloatX        float32
	FloatY        float32
	// MinWidth is the resize floor; 0 uses panelFloatMinW for movable windows else panelMinW.
	MinWidth float32
	// CloseOnEscape dismisses on Escape when the pointer is over the panel.
	CloseOnEscape bool
}

// DefaultPanelFeatures returns production defaults (title bar on, rest off).
func DefaultPanelFeatures() PanelFeatures {
	return PanelFeatures{
		TitleBar: true,
		AnchorX:  PanelAnchorLeft,
		AnchorY:  PanelAnchorTop,
		DragMode: PanelDragOff,
	}
}

// Features returns the live feature set for programmatic toggling.
func (p *Panel) Features() *PanelFeatures {
	return surfaceFeaturesConfig(&p.SurfaceShell)
}

// SyncFeatures reapplies the current feature flags (header mode, scroll host, collapse…).
func (p *Panel) SyncFeatures() {
	surfaceSyncFeatures(&p.SurfaceShell)
}

// SetTitleBar toggles the dark title bar (HeaderModeTitleBar vs HeaderModeNone).
func (p *Panel) SetTitleBar(on bool) *Panel {
	p.Features().TitleBar = on
	p.SyncFeatures()
	return p
}

// SetCollapsible toggles +/- collapse control in the header chrome.
func (p *Panel) SetCollapsible(on bool) *Panel {
	p.Features().Collapsible = on
	p.SyncFeatures()
	return p
}

// SetClosable toggles the × dismiss control in the header chrome.
func (p *Panel) SetClosable(on bool) *Panel {
	p.Features().Closable = on
	p.SyncFeatures()
	return p
}

// SetMovable toggles whether drag modes can reposition the panel.
func (p *Panel) SetMovable(on bool) *Panel {
	p.Features().Movable = on
	p.SyncFeatures()
	return p
}

// SetResizable toggles the bottom-right resize grip.
func (p *Panel) SetResizable(on bool) *Panel {
	p.Features().Resizable = on
	p.SyncFeatures()
	return p
}

// SetConstrain toggles clamping move/resize to the parent bounds.
func (p *Panel) SetConstrain(on bool) *Panel {
	p.Features().Constrain = on
	p.SyncFeatures()
	return p
}

// SetHScroll wraps the body in a horizontal scroll viewport when true.
func (p *Panel) SetHScroll(on bool) *Panel {
	p.Features().HScroll = on
	p.SyncFeatures()
	return p
}

// SetVScroll wraps the body in a vertical scroll viewport when true.
func (p *Panel) SetVScroll(on bool) *Panel {
	p.Features().VScroll = on
	p.SyncFeatures()
	return p
}

// SetAnchored toggles anchor positioning within the parent on layout.
func (p *Panel) SetAnchored(on bool) *Panel {
	p.Features().Anchored = on
	p.SyncFeatures()
	return p
}

// SetAnchor sets horizontal and vertical anchor when Anchored is true.
func (p *Panel) SetAnchor(x PanelAnchorH, y PanelAnchorV) *Panel {
	f := p.Features()
	f.AnchorX = x
	f.AnchorY = y
	p.SyncFeatures()
	return p
}

// SetDragMode sets how mouse drag moves the panel (requires SetMovable(true)).
func (p *Panel) SetDragMode(mode PanelDragMode) *Panel {
	p.Features().DragMode = mode
	p.SyncFeatures()
	return p
}

// SetCloseOnEscape toggles Escape-key dismiss when the pointer is over the panel.
func (p *Panel) SetCloseOnEscape(on bool) *Panel {
	p.Features().CloseOnEscape = on
	p.SyncFeatures()
	return p
}

// SetMinWidth sets the minimum width when Resizable is enabled (0 = default).
func (p *Panel) SetMinWidth(w float32) *Panel {
	p.Features().MinWidth = w
	return p
}

// SetFloatPosition places the panel at (x, y) relative to its parent's bounds on layout.
func (p *Panel) SetFloatPosition(x, y float32) *Panel {
	f := p.Features()
	f.FloatPosition = true
	f.FloatX = x
	f.FloatY = y
	f.Anchored = false
	p.SyncFeatures()
	return p
}

// SetOnDismiss sets the callback invoked when the user clicks the × button.
func (p *Panel) SetOnDismiss(fn func()) *Panel {
	p.Features().OnDismiss = fn
	if d := p.DismissBehavior(); d != nil {
		d.OnDismiss = fn
	}
	return p
}

// CollapseBehavior returns the collapse plugin when Collapsible is enabled.
func (p *Panel) CollapseBehavior() *CollapseBehavior {
	return surfaceCollapseBehavior(&p.SurfaceShell)
}

// DismissBehavior returns the dismiss plugin when Closable is enabled.
func (p *Panel) DismissBehavior() *DismissBehavior {
	return surfaceDismissBehavior(&p.SurfaceShell)
}
