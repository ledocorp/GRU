// Package ui provides a retained-mode UI engine built on top of raylib-go.
//
// # Architecture Overview
//
// Gru uses a retained widget tree where every widget is a Node. Each frame
// the engine runs three sequential phases:
//
//  1. Update(dt) — process input, run animations, mutate reactive state.
//  2. Layout()   — recompute positions/sizes for any dirty subtree.
//  3. Draw()     — render every visible node to the screen.
//
// # Dirty Propagation
//
// Whenever a widget mutates its own state (scroll position, text value, signal
// change, etc.) it calls MarkDirty(). MarkDirty bubbles upward through the
// parent chain so the root Container knows that Layout must run this frame.
// Critically, propagation works for any parent type — Container, Viewport, or
// any other Node — because Element.parent is stored as the Node interface.
//
// # Reactive Signals
//
// Signal[T] is a generic observable value. Any code that reads a Signal inside
// a NewEffect callback becomes a subscriber automatically. When the Signal
// changes, all subscribers re-run. Widgets expose their text/value/state as
// Signal fields so the application can bind labels to counters, sliders to
// volume, etc. without manual wiring.
//
// # Layout & layering stack (contract)
//
// Build scenes so dependencies flow inward — each layer assumes the ones outside
// it are already correct:
//
//  1. **Page grid (responsive page grid layout)** — a `Container` with
//     `LayoutGrid` and `ColSpanAt` / breakpoints. This is the product term;
//     the API constant remains `LayoutGrid`.
//     establishes column widths and row flow for page regions.
//  2. **ZIndex** orders *siblings* within the same parent for both Draw and hit-testing
//     (`FindInteractiveAt` uses the same high-to-low order as sorted drawing).
//  3. **Viewport** clips and scrolls content laid out inside its bounds; children
//     should stay within the flex/grid region the Viewport was given.
//  4. **Panel** (or Card) is a titled surface inside that viewport: body layout
//     delegates to the same flex engine as Container.
//  5. **Flex** (`LayoutFlex` / `LayoutResponsive`) arranges widgets inside the panel
//     body (or any Container). It does not replace grid at the page level — nest
//     flex rows/columns *inside* grid cells or stacked regions as needed.
//
// `Document.Root` uses `LayoutAbsolute` so scenes can place a full-window shell;
// use grid/flex *inside* that shell for responsive content. See ARCHITECTURE.md §5.6.
//
// # Clipping Hierarchy
//
// raylib's BeginScissorMode / EndScissorMode are NOT stackable — each call
// replaces the active scissor rectangle. Gru handles this explicitly:
//   - Viewport re-applies its scissor after each child Draw() call so that
//     children which open their own scissor (e.g. VirtualList, TextInput)
//     do not leave the viewport unclipped.
//   - VirtualList and TextInput intersect their own clip rect with the parent
//     Viewport's ClipBounds() so their content can never bleed outside the
//     viewport panel. The package-level findViewport(n) / intersectRects()
//     helpers in layout.go centralise this logic.
//
// # Viewport vs Container
//
// Container is a plain layout box. Viewport embeds Container and adds
// scroll state, a scrollbar, and strict scissor clipping. Children of a
// Viewport should be added with Viewport.AddChild so their parent pointer is
// set to the *Viewport (not the embedded *Element), enabling type-assertions
// in child widgets that need to query ClipBounds().
//
// # Widget Summary
//
//   - Button       — clickable label; hover highlight, scale-bounce animation
//   - IconButton   — Button with a leading symbol glyph or PNG icon
//   - Label        — reactive text display
//   - TextInput    — editable text with cursor, focus, horizontal scroll; scissor-clipped to parent Viewport
//   - Slider       — draggable range selector with indigo fill + thumb glow
//   - Checkbox     — rounded boolean toggle, indigo when checked, hover glow ring
//   - Toggle       — animated sliding thumb switch
//   - Dropdown     — option list; open popup rendered above other widgets via ZIndex
//   - ProgressBar  — pill-shaped fill bar with scissor-cut rounded fill
//   - Header       — large title + optional subtitle; optional indigo left accent bar
//   - Separator    — horizontal divider line with optional centred label
//   - Image        — lazy-loading texture widget with FitContain/FitStretch/FitCover
//   - Panel        — titled container: dark title bar + body; multi-layer shadow
//   - Card         — lightweight Panel variant: white bg, 8px rounded corners, no title bar colour
//   - Container    — flex layout box (row or column)
//   - Viewport     — scrollable, scissor-clipped container with scrollbar
//   - VirtualList  — large scrollable list with view culling (O(visible) draw calls)
//
// # Upcoming: Advanced Rendering
//
// See ADVANCED_VISUALS.MD for the planned SDF/MSDF font upgrade, supersampled
// RenderTexture (2× SSAA), and gg-based vector shapes. The upgrade path is
// confined to ui/font.go, ui/render.go (new), and main.go window flags — no
// widget draw code changes required.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// Node represents a UI element in the widget tree.
// All concrete widgets implement this interface by embedding Element and
// overriding the methods they need.
type Node interface {
	EventEmitter
	// ID returns the unique string identifier of the node.
	ID() string
	// Parent returns the parent as *Container if possible, else nil.
	// NOTE: If the node's parent is a *Viewport (which embeds Container but
	// is not a *Container), this returns nil. Use ParentNode() for a type-safe
	// walk of the full parent chain (e.g. to find an ancestor *Viewport).
	Parent() *Container
	// ParentNode returns the raw parent Node regardless of its concrete type.
	// Prefer this over Parent() when you need to walk the ancestry chain to
	// find a *Viewport, *Panel, or other concrete ancestor.
	ParentNode() Node
	// SetParent sets the parent node reference. Called by AddChild.
	SetParent(Node)
	// Children returns the slice of child nodes.
	Children() []Node
	// AddChild appends a child and sets its parent pointer.
	AddChild(Node)
	// RemoveChild removes the first child with the given ID.
	RemoveChild(string)
	// Bounds returns the axis-aligned bounding rectangle in screen space.
	Bounds() rl.Rectangle
	// SetBounds updates the bounding rectangle and calls MarkDirty.
	SetBounds(rl.Rectangle)
	// Update is called once per frame before Layout and Draw. Widgets handle
	// mouse/keyboard input and mutate reactive state here.
	Update(dt float32)
	// Layout recomputes child positions. It is a no-op if IsDirty() is false.
	// Leaf widgets must still clear layoutDirty in Layout() — see IDLE_INVARIANTS.md.
	Layout()
	// Draw renders the node to the screen. Widgets that use scissor clipping
	// must ensure BeginScissorMode/EndScissorMode are balanced within their
	// Draw call so the parent's scissor state is not corrupted.
	Draw()
	// MarkDirty flags this node as needing both relayout and redraw, propagating
	// upward so the root Container runs Layout this frame.
	MarkDirty()
	// MarkDrawDirty flags this node as needing a visual redraw only (no layout).
	// Use this for purely visual changes — hover, blink, color — where the
	// widget bounds do not change. Bubbles up through the parent chain so parent
	// caches know to re-render, but does NOT trigger a Layout() pass.
	MarkDrawDirty()
	// IsDirty returns true if Layout needs to run for this subtree.
	IsDirty() bool
	// UsesScissor reports whether this widget opens a BeginScissorMode region
	// inside its own Draw() call. Viewport uses this to skip the redundant
	// scissor re-application after children that never alter the scissor state.
	UsesScissor() bool
	// IsInteractive returns true for widgets that handle mouse/keyboard input
	// (Button, TextInput, Slider, Checkbox, Dropdown, VirtualList).
	IsInteractive() bool
	// Hide makes the node invisible and skips Update/Layout/Draw.
	Hide()
	// Show makes the node visible again.
	Show()
	// IsHidden returns true if the node is currently hidden.
	IsHidden() bool
	// GetZIndex returns the Z-index used for draw ordering within a container.
	GetZIndex() int
	// SetZIndex sets the Z-index for draw ordering.
	SetZIndex(int)
	// GetFlexGrow returns the flex-grow factor for the parent's flex layout.
	GetFlexGrow() float32
	// SetFlexGrow sets the flex-grow factor and marks the node dirty.
	SetFlexGrow(float32)
	// IsAutoHeight reports whether this node should compute its own height from
	// content. True for Container/Panel/Viewport nodes created with h=0.
	// When true, layoutFlex (FlexColumn), layoutGrid, and Panel.layoutContent
	// will shrink-wrap the node's bounds.Height to fit its children after layout.
	IsAutoHeight() bool
	// StyleName returns the style key applied to this node (e.g. "button", "panel-title").
	StyleName() string
}

// ─── CachePolicy ─────────────────────────────────────────────────────────────

// CachePolicy controls whether a widget renders into a RenderTexture and reuses
// the result on subsequent clean frames.
//
//   - CacheNever  — always draw directly (default; safe for all widgets)
//   - CacheAuto   — reserved for future SSAA-aware auto-caching logic
//   - CacheAlways — force render-texture caching (caller ensures SSAA compat)
type CachePolicy int

const (
	CacheNever  CachePolicy = iota // No caching — draw every frame
	CacheAuto                      // Engine-managed (reserved; currently behaves as Never)
	CacheAlways                    // Explicit caller-managed caching
)

// # LLM Prompt Template — custom leaf widget
//
//	type Meter struct { ui.Element; value float32 }
//
//	func NewMeter(id string) *Meter {
//	    m := &Meter{Element: ui.NewElement(id, 0, 0, 120, 24)}
//	    return m
//	}
//	func (m *Meter) Update(dt float32) { /* read input, mutate state */ }
//	func (m *Meter) Layout() { m.layoutDirty = false } // required — idle FPS contract
//	func (m *Meter) Draw() { /* raylib draws inside m.Bounds() */ }
//
// Use MarkDrawDirty for hover/color-only changes; MarkDirty when bounds or
// structure change. See docs/IDLE_INVARIANTS.md.
//
// Element is the base struct embedded by every widget.
//
// Dirty flags use a two-level split:
//   - layoutDirty: size or position changed; Layout() + Draw() must run.
//   - drawDirty:   visual content changed (color, text, hover); Draw() only.
//
// MarkDirty() sets both flags and propagates upward; the root Container's
// layoutDirty being true is what triggers the Layout() call in main.go.
// MarkDrawDirty() sets only drawDirty and propagates upward as draw-only,
// so a hover change on a button never triggers a full Layout pass.
//
// Invariant: e.parent is always either nil (root) or the Node passed to the
// most recent SetParent call. Viewport.AddChild sets parent = *Viewport;
// plain Container.AddChild (via Element.AddChild) sets parent = *Element.
type Element struct {
	id          string
	parent      Node
	children    []Node
	bounds      rl.Rectangle
	layoutDirty bool // needs Layout() + Draw() this frame
	drawDirty   bool // needs Draw() only (no layout change)
	hidden      bool
	styleName   string // legacy theme style name
	// Theme v2 fields. Empty component/variant means "use styleName through the
	// legacy CurrentTheme map." This keeps SetStyle-compatible behavior intact.
	styleComponent      string
	styleVariant        string
	styleOverrides      *Style
	// stylePatch holds DocBlock/JSON overrides with pointer semantics so explicit
	// zero values (e.g. padding: 0) are not dropped by mergeStyle.
	stylePatch          *styleJSON
	styleDirty          bool
	styleRevision       uint64
	styleState          StyleState
	styleStateApplied   bool
	presetGlowIntensity float32 // parametric glow from SetStylePreset (0..1)
	presetGlowSet       bool    // true when glowIntensity prop was explicitly provided
	presetName          string  // last SetStylePreset name; drives ChromeProfile resolution
	presetHoverLift     bool    // static lifted shadow when hoverLift prop is true
	resolvedStyle       Style
	resolvedStyleValid  bool
	// Cache fields — set via SetCachePolicy; currently CacheAlways only.
	// CacheAuto is reserved for future SSAA-aware automatic caching.
	cachePolicy   CachePolicy
	renderTexture rl.RenderTexture2D
	cacheDirty    bool
	eventHandlers map[EventType][]func(Event)
	ZIndex        int
	FlexGrow      float32
	AutoHeight    bool // true = shrink-wrap height to content (set when created with h=0)
	// Responsive size hints — used by LayoutResponsive and flex layout.
	// A value of 0 means "unconstrained" (no constraint applied).
	MinWidth       float32 // Minimum content width (px); flex will not shrink below this
	MaxWidth       float32 // Maximum content width (px); flex will not grow above this
	PreferredWidth float32 // Preferred width hint; overrides natural width when non-zero

	// Grid layout hints — used by LayoutGrid on the parent Container.
	//
	// ColSpan holds the column span for each responsive breakpoint tier in
	// the order [XS, SM, MD, LG, XL]. A value of 0 means "inherit the next
	// larger tier's span" (mobile-first cascade). A ColSpan of 12 fills
	// the full row; 6 takes half. Use SetColSpan / ColSpanAt helpers.
	//
	//   child.SetColSpan(ui.BreakpointXS, 12)  // full-width on mobile
	//   child.SetColSpan(ui.BreakpointMD, 6)   // half-width on tablet+
	//
	ColSpan [5]int // indexed by Breakpoint constant (0=XS … 4=XL)
	RowSpan int    // number of grid rows the child spans (0 or 1 = one row)

	// layoutExtentBottom tracks nodeSubtreeBottom after Layout for extent propagation
	// (docs/LAYOUT_CONTRACTS.md §5).
	layoutExtentBottom float32
	layoutExtentValid  bool
}

// NewElement creates a new Element with the given ID and bounds.
func NewElement(id string, x, y, w, h float32) Element {
	return Element{
		id:            id,
		bounds:        rl.NewRectangle(x, y, w, h),
		layoutDirty:   true,
		drawDirty:     true,
		styleName:     "default",
		cachePolicy:   CacheNever,
		eventHandlers: make(map[EventType][]func(Event)),
	}
}

// ID implements Node.ID.
func (e *Element) ID() string { return e.id }

// Parent implements Node.Parent.
func (e *Element) Parent() *Container {
	if c, ok := e.parent.(*Container); ok {
		return c
	}
	return nil
}

// ParentNode implements Node.ParentNode — returns the raw parent Node.
func (e *Element) ParentNode() Node { return e.parent }

// SetParent implements Node.SetParent.
func (e *Element) SetParent(p Node) { e.parent = p }

// Children implements Node.Children.
func (e *Element) Children() []Node { return e.children }

// Bounds implements Node.Bounds.
func (e *Element) Bounds() rl.Rectangle { return e.bounds }

// SetBounds implements Node.SetBounds.
func (e *Element) SetBounds(r rl.Rectangle) {
	if e.bounds == r {
		return
	}
	e.bounds = r
	e.MarkDirty()
}

// setBoundsNoMark updates the bounds without triggering MarkDirty propagation.
// Used exclusively by Viewport.repositionOnly() for the scroll fast-path:
// we want child Y positions updated silently so the next frame does not think
// a layout change occurred. The method is unexported (lowercase) so it cannot
// be called from application code.
func (e *Element) setBoundsNoMark(r rl.Rectangle) {
	if e.bounds == r {
		return
	}
	e.bounds = r
	// Mark only draw-dirty on self (position changed, redraw needed) but do
	// NOT propagate layoutDirty upward — the Viewport already knows it needs
	// a draw this frame from its own dirty state.
	e.drawDirty = true
}

// AddChild implements Node.AddChild.
// Note: Primarily intended for Container widgets.
func (e *Element) AddChild(child Node) {
	e.children = append(e.children, child)
	child.SetParent(e)
	e.MarkDirty()
}

// RemoveChild implements Node.RemoveChild.
func (e *Element) RemoveChild(id string) {
	for i, c := range e.children {
		if c.ID() == id {
			e.children = append(e.children[:i], e.children[i+1:]...)
			e.MarkDirty()
			break
		}
	}
}

// ReplaceChildAt swaps the child at index i and marks the parent dirty.
// Used by demo hot-reload to replace a compiled .gru body under a Viewport header.
func (e *Element) ReplaceChildAt(i int, child Node) {
	if i < 0 || i >= len(e.children) {
		return
	}
	e.children[i] = child
	child.SetParent(e)
	e.MarkDirty()
}

// GetZIndex returns the Z-index for layering.
func (e *Element) GetZIndex() int {
	return e.ZIndex
}

// SetZIndex sets the Z-index for layering.
func (e *Element) SetZIndex(z int) {
	e.ZIndex = z
}

// GetFlexGrow returns the flex grow factor.
func (e *Element) GetFlexGrow() float32 {
	return e.FlexGrow
}

// SetFlexGrow sets the flex grow factor.
func (e *Element) SetFlexGrow(g float32) {
	e.FlexGrow = g
	e.MarkDirty()
}

// IsAutoHeight implements Node.IsAutoHeight.
func (e *Element) IsAutoHeight() bool { return e.AutoHeight }

// GetMinWidth returns the minimum width hint (0 = unconstrained).
func (e *Element) GetMinWidth() float32 { return e.MinWidth }

// GetMaxWidth returns the maximum width hint (0 = unconstrained).
func (e *Element) GetMaxWidth() float32 { return e.MaxWidth }

// GetPreferredWidth returns the preferred width hint (0 = use natural width).
func (e *Element) GetPreferredWidth() float32 { return e.PreferredWidth }

// SetColSpan sets the number of grid columns this element spans at the given
// breakpoint tier. The value cascades upward (mobile-first): if a tier is 0
// the next smaller set tier is used.
//
//	el.SetColSpan(ui.BreakpointXS, 12) // full row on mobile
//	el.SetColSpan(ui.BreakpointMD, 6)  // half row on tablet+
func (e *Element) SetColSpan(bp Breakpoint, cols int) {
	if bp < 0 || int(bp) >= len(e.ColSpan) {
		return
	}
	e.ColSpan[bp] = cols
	e.MarkDirty()
}

// ColSpanAt returns the effective column span at the given breakpoint,
// cascading down to smaller tiers if the requested one is unset (0).
// Falls back to 12 (full row) if no tier is set at all.
func (e *Element) ColSpanAt(bp Breakpoint) int {
	// Walk from bp down to XS looking for a set value.
	for i := int(bp); i >= 0; i-- {
		if e.ColSpan[i] != 0 {
			return e.ColSpan[i]
		}
	}
	return 12 // default: full row
}

// GetRowSpan returns the grid row span (1 if unset).
func (e *Element) GetRowSpan() int {
	if e.RowSpan <= 0 {
		return 1
	}
	return e.RowSpan
}

// MarkDirty flags this node and all ancestors as needing layout + redraw.
//
// Propagation stops only at the root (parent == nil). Because parent is
// stored as the Node interface, this correctly bubbles through Viewport,
// Container, or any other parent type — no type-assertion required.
// This is the key mechanism that triggers Layout() from the root each frame
// whenever any descendant changes scroll position, text, or size.
func (e *Element) MarkDirty() {
	e.layoutDirty = true
	e.drawDirty = true
	if e.cachePolicy != CacheNever {
		e.cacheDirty = true
	}
	// Propagate to any parent node (Viewport, Container, etc.).
	if e.parent != nil {
		e.parent.MarkDirty()
	}
}

// MarkDrawDirty flags this node as needing a visual redraw only.
//
// Unlike MarkDirty, this does NOT set layoutDirty, so a full Layout() pass
// is not triggered. Use for purely visual changes (hover highlight, blink
// cursor, color animation) where bounds do not change. Bubbles upward as
// draw-dirty so any ancestor's render-texture cache is invalidated.
func (e *Element) MarkDrawDirty() {
	e.drawDirty = true
	if e.cachePolicy != CacheNever {
		e.cacheDirty = true
	}
	if e.hidden || e.parent == nil {
		return
	}
	e.parent.MarkDrawDirty()
}

// IsDirty implements Node.IsDirty — returns true if Layout() must run.
func (e *Element) IsDirty() bool { return e.layoutDirty }

// UsesScissor implements Node.UsesScissor — false for most widgets.
// Override in TextInput, VirtualList, Slider, ProgressBar.
func (e *Element) UsesScissor() bool { return false }

// Update implements Node.Update (default no-op).
func (e *Element) Update(dt float32) {}

// Layout implements Node.Layout (default no-op). Clears layoutDirty so a completed
// pass does not pin NeedsRedraw / SSAA cache hits — override and call through when
// you replace this with real layout work.
func (e *Element) Layout() { e.layoutDirty = false }

// IsInteractive implements Node.IsInteractive.
func (e *Element) IsInteractive() bool { return false }

// propagateVisibilityReflow marks a flex parent and its siblings dirty when
// visibility changes so flex-grow children reclaim space immediately (status
// bar show/hide, collapsed panels) without waiting for an unrelated click.
func (e *Element) propagateVisibilityReflow() {
	p := e.Parent()
	if p == nil {
		return
	}
	p.MarkDirty()
	if p.LayoutType == LayoutFlex {
		for _, ch := range p.Children() {
			if ch != e {
				ch.MarkDirty()
			}
		}
	}
}

// Hide implements Node.Hide.
func (e *Element) Hide() {
	if e.hidden {
		return
	}
	e.hidden = true
	e.propagateVisibilityReflow()
	e.MarkDirty()
	e.MarkDrawDirty()
}

// Show implements Node.Show.
func (e *Element) Show() {
	if !e.hidden {
		return
	}
	e.hidden = false
	e.propagateVisibilityReflow()
	e.MarkDirty()
	e.MarkDrawDirty()
}

// IsHidden implements Node.IsHidden.
func (e *Element) IsHidden() bool { return e.hidden }

// On implements EventEmitter.On.
func (e *Element) On(eventType EventType, handler func(Event)) {
	e.eventHandlers[eventType] = append(e.eventHandlers[eventType], handler)
}

// Emit implements EventEmitter.Emit.
func (e *Element) Emit(eventType EventType, data interface{}) {
	event := Event{Type: eventType, Target: e, Data: data, Bubble: true}
	if handlers, ok := e.eventHandlers[eventType]; ok {
		for _, handler := range handlers {
			handler(event)
		}
	}
	// Simple bubbling: if not handled and has parent, emit on parent
	if event.Bubble && e.parent != nil {
		e.parent.Emit(eventType, data)
	}
}

// GetStyle returns the resolved style for this element from Theme v2 or the
// legacy CurrentTheme fallback.
func (e *Element) GetStyle() Style {
	style, _ := e.ResolveStyle(StyleStateNone)
	return style
}

// ResolveStyle returns the concrete Style for this element and whether a Theme
// v2 state style participated in resolution.
func (e *Element) ResolveStyle(state StyleState) (Style, bool) {
	if e == nil {
		return DefaultStyle, false
	}
	if e.resolvedStyleValid &&
		!e.styleDirty &&
		e.styleRevision == themeRevisionV2 &&
		e.styleState == state {
		return e.resolvedStyle, e.styleStateApplied
	}

	style := DefaultStyle
	stateApplied := false
	resolvedV2 := false
	if currentThemeV2 != nil && e.styleComponent != "" {
		if comp, ok := currentThemeV2.Components[e.styleComponent]; ok {
			resolvedV2 = true
			style = comp.Base
			if e.styleVariant != "" {
				if variant, ok := comp.Variants[e.styleVariant]; ok {
					style = mergeStyle(style, variant)
				}
			}
			for _, name := range styleStateNames(state) {
				if st, ok := comp.States[name]; ok {
					style = mergeStyle(style, st)
					stateApplied = true
				}
			}
		}
	}
	if !resolvedV2 {
		style = e.resolveLegacyStyle()
	}
	if e.styleOverrides != nil {
		style = mergeStyle(style, *e.styleOverrides)
	}
	if e.stylePatch != nil {
		if patched, err := e.stylePatch.toStyle(style); err == nil {
			style = patched
		}
	}

	e.resolvedStyle = style
	e.resolvedStyleValid = true
	e.styleDirty = false
	e.styleRevision = themeRevisionV2
	e.styleState = state
	e.styleStateApplied = stateApplied
	return style, stateApplied
}

func (e *Element) resolveLegacyStyle() Style {
	names := []string{e.styleName}
	if e.styleComponent != "" {
		names = []string{
			legacyComponentVariantName(e.styleComponent, e.styleVariant),
			legacyVariantName(e.styleVariant),
			e.styleComponent,
			e.styleName,
		}
	}
	for _, name := range names {
		if name == "" {
			continue
		}
		if style, ok := CurrentTheme[name]; ok {
			return style
		}
	}
	return DefaultStyle
}

func legacyComponentVariantName(component, variant string) string {
	if component == "" || variant == "" || variant == "default" {
		return ""
	}
	return component + "-" + variant
}

func legacyVariantName(variant string) string {
	if variant == "default" {
		return ""
	}
	return variant
}

// SetStyle sets the legacy style name for this element.
func (e *Element) SetStyle(name string) {
	e.styleName = name
	e.styleComponent = ""
	e.styleVariant = ""
	e.styleDirty = true
	e.resolvedStyleValid = false
	e.MarkDirty()
}

// SetStyleVariant opts into Theme v2 component + variant resolution.
func (e *Element) SetStyleVariant(component, variant string) {
	e.styleComponent = component
	e.styleVariant = variant
	e.styleDirty = true
	e.resolvedStyleValid = false
	e.MarkDirty()
}

// SetVariant changes only the Theme v2 variant for this element. If no component
// has been set yet, the legacy styleName is used as the component name.
func (e *Element) SetVariant(variant string) {
	if e.styleComponent == "" {
		e.styleComponent = e.styleName
	}
	e.styleVariant = variant
	e.styleDirty = true
	e.resolvedStyleValid = false
	e.MarkDirty()
}

// SetStyleOverrides adds concrete inline overrides for JSON/document generated
// UI. Overrides compose after base, variant, and state styles.
func (e *Element) SetStyleOverrides(style Style) {
	e.styleOverrides = &style
	e.styleDirty = true
	e.resolvedStyleValid = false
	e.MarkDirty()
}

// ClearStyleOverrides removes inline overrides.
func (e *Element) ClearStyleOverrides() {
	e.styleOverrides = nil
	e.stylePatch = nil
	e.presetName = ""
	e.presetHoverLift = false
	e.presetGlowIntensity = 0
	e.presetGlowSet = false
	e.styleDirty = true
	e.resolvedStyleValid = false
	e.MarkDirty()
}

// PresetGlowIntensity returns the stored glow prop from SetStylePreset (0 when unset).
func (e *Element) PresetGlowIntensity() float32 { return e.presetGlowIntensity }

// ChromeGlowIntensity returns the glow amount used when drawing surface chrome.
// Preset props (including explicit 0) win; card/neo-glow variant gets a default.
func (e *Element) ChromeGlowIntensity() float32 {
	if e.presetGlowSet {
		return e.presetGlowIntensity
	}
	if e.presetGlowIntensity > 0 {
		return e.presetGlowIntensity
	}
	if e.styleComponent == "card" && e.styleVariant == "neo-glow" {
		return 0.5
	}
	return 0
}

// mergeStylePatch merges a JSON style patch (supports explicit zeros).
func (e *Element) mergeStylePatch(add styleJSON) {
	e.stylePatch = mergeStyleJSON(e.stylePatch, add)
	e.styleDirty = true
	e.resolvedStyleValid = false
}

// StyleName implements Node.StyleName.
func (e *Element) StyleName() string { return e.styleName }

// SetCachePolicy sets the render-texture caching strategy for this element.
//
//   - CacheNever  — draw directly every frame (default; correct for all widgets)
//   - CacheAuto   — reserved; currently behaves as CacheNever
//   - CacheAlways — caller explicitly opts into RenderTexture caching
//
// NOTE: CacheAlways creates a texture at the widget's current bounds size.
// When SSAA is active the texture is rendered at 1× and bilinearly upscaled,
// which is acceptable for shape/background content but will soften SDF text.
// Prefer CacheAlways only on widgets whose content is purely geometric.
func (e *Element) SetCachePolicy(p CachePolicy) {
	if p == e.cachePolicy {
		return
	}
	if e.cachePolicy == CacheAlways && e.renderTexture.ID != 0 {
		rl.UnloadRenderTexture(e.renderTexture)
		e.renderTexture = rl.RenderTexture2D{}
	}
	e.cachePolicy = p
	if p == CacheAlways {
		tw, th := clampDim(int32(e.bounds.Width), int32(e.bounds.Height))
		e.renderTexture = rl.LoadRenderTexture(tw, th)
		e.cacheDirty = true
	}
}

// EnableCache is a convenience wrapper: SetCachePolicy(CacheAlways).
// Preserved for backward compatibility.
func (e *Element) EnableCache() { e.SetCachePolicy(CacheAlways) }

// DisableCache is a convenience wrapper: SetCachePolicy(CacheNever).
// Preserved for backward compatibility.
func (e *Element) DisableCache() { e.SetCachePolicy(CacheNever) }

// Draw implements Node.Draw with optional render-texture caching.
// If caching is disabled (the default), drawInternal is called directly.
// Widgets override drawInternal, not Draw, so caching applies transparently.
//
// After drawing, drawDirty is cleared on this node. layoutDirty is cleared
// by Layout() when the layout pass runs.
func (e *Element) Draw() {
	if e.cachePolicy == CacheAlways && e.renderTexture.ID != 0 {
		if e.cacheDirty {
			rl.BeginTextureMode(e.renderTexture)
			rl.PushMatrix()
			rl.Translatef(-e.bounds.X, e.bounds.Y+e.bounds.Height, 0)
			rl.Scalef(1, -1, 1)
			e.drawInternal()
			rl.PopMatrix()
			rl.EndTextureMode()
			e.cacheDirty = false
		}
		rl.DrawTextureRec(e.renderTexture.Texture,
			rl.NewRectangle(0, 0, e.bounds.Width, -e.bounds.Height),
			rl.NewVector2(e.bounds.X, e.bounds.Y), rl.White)
	} else {
		e.drawInternal()
	}
	e.drawDirty = false
}

// drawInternal performs the actual drawing (overridden by subclasses).
// Default implementation does nothing (no debug drawing).
func (e *Element) drawInternal() {
	// No debug drawing in production
}

// ─── Inspector accessors ──────────────────────────────────────────────────────
// These are read-only accessors exposing internal state for the Inspector
// overlay. They are prefixed with Dbg to signal that they are debug-only and
// should not be used by application code.

// DbgLayoutDirty returns true if this node needs a layout pass this frame.
func (e *Element) DbgLayoutDirty() bool { return e.layoutDirty }

// DbgDrawDirty returns true if this node needs a visual redraw this frame.
func (e *Element) DbgDrawDirty() bool { return e.drawDirty }

// ClearDrawDirtySubtree clears draw-only dirty flags without touching layout.
// Call when the window loses focus so caret-blink / hover chrome does not force
// a full SSAA redraw every few frames while the app is in the background.
func ClearDrawDirtySubtree(root Node) {
	if root == nil {
		return
	}
	clearDrawDirtySubtree(root)
}

func clearDrawDirtySubtree(n Node) {
	if c, ok := n.(interface{ ClearDrawDirtyFlag() }); ok {
		c.ClearDrawDirtyFlag()
	}
	for _, ch := range n.Children() {
		clearDrawDirtySubtree(ch)
	}
}

// ClearLayoutDirtyFlag clears the layout-only dirty bit after a custom Layout pass.
func (e *Element) ClearLayoutDirtyFlag() { e.layoutDirty = false }

// ClearDrawDirtyFlag clears the draw-only dirty bit (Element / embedded widgets).
func (e *Element) ClearDrawDirtyFlag() { e.drawDirty = false }

// SubtreeNeedsRedraw reports whether any visible node in the tree needs layout or draw.
// Hidden subtrees are skipped — chrome that is Show/Hide toggled must not pin SSAA at 60 FPS.
func SubtreeNeedsRedraw(n Node) bool {
	if n == nil || n.IsHidden() {
		return false
	}
	if n.IsDirty() {
		return true
	}
	if d, ok := n.(interface{ DbgDrawDirty() bool }); ok && d.DbgDrawDirty() {
		return true
	}
	for _, ch := range n.Children() {
		if SubtreeNeedsRedraw(ch) {
			return true
		}
	}
	return false
}

// WakeInteractive groups wakes that should keep ActiveFPS while the window is focused.
const WakeInteractive = WakeInput | WakeScroll | WakeKeyboard | WakeResize | WakeScene | WakeDataUpdate

// WakeSummaryForIdlePolicy strips hover/overlay wakes so a focused window can
// reach DeepIdleFPS while the cursor rests over chrome. WakeAnimation is kept so
// Spinner / accordion tweens can settle at AnimationFPS (not DeepIdleFPS 10).
func WakeSummaryForIdlePolicy(in WakeSummary) WakeSummary {
	if in.Reasons == 0 {
		return in
	}
	return WakeSummary{Reasons: in.Reasons &^ WakeOverlay}
}

// WakeSummaryForBackground strips input/animation/overlay wakes when the native
// window is unfocused so idle policy can reach DeepIdleFPS.
func WakeSummaryForBackground(in WakeSummary) WakeSummary {
	if in.Reasons == 0 {
		return in
	}
	return WakeSummary{Reasons: in.Reasons & (WakeResize | WakeScene)}
}

// DbgCachePolicy returns the cache policy name as a short string.
func (e *Element) DbgCachePolicy() string {
	switch e.cachePolicy {
	case CacheNever:
		return "never"
	case CacheAuto:
		return "auto"
	case CacheAlways:
		return "always"
	default:
		return "?"
	}
}

// InvalidateResolvedStyle clears cached theme styles after SetAppearance.
func (e *Element) InvalidateResolvedStyle() {
	e.styleDirty = true
	e.resolvedStyleValid = false
	e.MarkDrawDirty()
}

// MarkTreeThemeDirty invalidates resolved styles after SetAppearance so widgets
// pick up new theme colors on the next draw.
func MarkTreeThemeDirty(root Node) {
	markTreeThemeDirtyDepth(root, 0, 128)
}

func markTreeThemeDirtyDepth(root Node, depth, maxDepth int) {
	if root == nil || depth > maxDepth {
		return
	}
	if inv, ok := root.(interface{ InvalidateResolvedStyle() }); ok {
		inv.InvalidateResolvedStyle()
	}
	for _, ch := range root.Children() {
		markTreeThemeDirtyDepth(ch, depth+1, maxDepth)
	}
}
