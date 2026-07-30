// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ToolbarItemType identifies what kind of item a ToolbarItem is.
type ToolbarItemType int

const (
	// ToolbarItemButton is a regular clickable button item.
	ToolbarItemButton ToolbarItemType = iota
	// ToolbarItemIconButton is an icon-bearing clickable button item.
	ToolbarItemIconButton
	// ToolbarItemToggle is a stateful toggle item.
	ToolbarItemToggle
	// ToolbarItemSeparator is a thin vertical divider between items.
	ToolbarItemSeparator
	// ToolbarItemCustom is any arbitrary Node embedded in the toolbar.
	ToolbarItemCustom
)

// toolbar layout constants (polish §4.2 — spacing aligned with Header / Panel title rhythm)
const (
	tbDefaultH   = float32(54) // default flat toolbar height (row edge pad + inner btn padding + type)
	tbPad        = float32(12) // left/right outer padding (4px rhythm; matches page grid xs)
	tbItemGap    = float32(4)  // gap between items within a group
	tbGroupGap   = float32(10) // gap between groups (on each side of the sep line)
	tbSepW       = float32(1)  // visible separator line width
	tbSepH       = float32(28) // separator visible height (centred vertically)
	tbRibbonTabH = float32(32) // height of ribbon tab bar
	tbOvflW      = float32(28) // width of the overflow "..." button
	tbMinItemW   = float32(28) // minimum item width (separators excluded)

	tbRibbonStackedCellW = float32(58) // ribbon icon-above-label cell width
	tbRibbonStackedCellH = float32(54) // ribbon stacked cell height
	tbRibbonCaptionFS  = int32(11)    // uniform caption under stacked icons (all cells)
	tbSectionLabelH      = float32(18) // min band for section caption (see ribbonSectionLabelHeight)
	tbRibbonSectionGap   = float32(40) // gap between ribbon sections (divider sits in center)
	tbRibbonItemGap      = float32(6)  // gap between cells within a ribbon section
	tbRibbonCompactItemGap = float32(2) // gap after toolbar-menu fields (find, theme dropdown)
	tbRibbonCompactCustomScale = float32(0.8) // find/theme field height vs stacked icon cell
	tbRibbonCompactMinW        = float32(88)  // floor when ribbon flex-shrinks toolbar-menu fields
	tbScrollStep         = float32(80) // chevron / page scroll step
	tbScrollWheelScale   = float32(32) // pixels per wheel notch

	// tbRowEdgePad is space between the toolbar inner edge and control chrome (flat row
	// and ribbon band below tabs). tbRibbonRowGap matches it above section captions.
	tbRowEdgePad   = float32(6)
	tbRibbonRowGap = tbRowEdgePad

	tbIconOnlyW = float32(40) // flat icon-only slot (2× toolbar-btn inner pad + icon)
	tbToggleW    = float32(34) // compact toggle track width (~1.9× height)
	tbToggleH    = float32(18) // compact toggle track height
	tbOvflRowH = float32(44) // overflow menu row height (matches toolbar-btn type scale)
)

// ─────────────────────────────────────────────────────────────────────────────
// ToolbarItem
// ─────────────────────────────────────────────────────────────────────────────

// ToolbarItem is an entry managed by a Toolbar.  It wraps an underlying Node
// widget (Button, IconButton, Toggle, …) together with metadata used by the
// toolbar to position, show/hide, and draw the item.
type ToolbarItem struct {
	id         string
	itemType   ToolbarItemType
	widget     Node // nil for ToolbarItemSeparator
	hidden     bool // true when pushed past the overflow threshold
	suppressed bool // app-controlled; hidden on the active ribbon tab (overflow layout)
	menuLabel  string // overflow menu text (tooltips for icon-only commands)
}

// ─────────────────────────────────────────────────────────────────────────────
// ToolbarGroup
// ─────────────────────────────────────────────────────────────────────────────

// ToolbarGroup is a named collection of ToolbarItems.
//
// Command bar: TabIndex = -1 (default via AddGroup). Ribbon: use AddSection so
// TabIndex selects the ribbon tab and Label is the section caption (Font, Insert, …).
type ToolbarGroup struct {
	ID       string
	Label    string
	TabIndex int // -1 flat bar; >=0 ribbon tab index
	items    []*ToolbarItem
}

// ─────────────────────────────────────────────────────────────────────────────
// Toolbar
// ─────────────────────────────────────────────────────────────────────────────

// Toolbar is a horizontal bar of grouped interactive items.
//
// # Layout
//
// Items are laid out left-to-right.  Each group is separated from the next by
// a thin vertical rule.  When Overflow is true and the total item width exceeds
// the toolbar's content area, items that don't fit are hidden and a "..." button
// appears at the right edge.  Clicking "..." opens an inline popup listing the
// hidden items by label.
//
// # Ribbon Mode
//
// When Ribbon is true a tab row is rendered at the top of the toolbar
// (tbRibbonTabH px tall) and only the items belonging to the active group are
// shown in the item row below.  Clicking a tab changes the active group.
//
// # Styles
//
//	"toolbar"              — overall background, border, padding (polish §4.2: padding aligns with dense rows)
//	"toolbar-separator"    — colour of group divider lines
//	"toolbar-ribbon-tab"   — inactive ribbon tab style
//	"toolbar-ribbon-tab-active" — active ribbon tab style
//
// # Usage
//
//	tb := ui.NewToolbar("main-tb", 0, 0, 1200, 0)
//	tb.AddGroup("file", "File")
//	tb.AddButton("file", "new",  "New",  func() { ... })
//	tb.AddButton("file", "open", "Open", func() { ... })
//	tb.AddSeparator("file", "sep1")
//	tb.AddButton("file", "save", "Save", func() { ... })
//	tb.AddGroup("view", "View")
//	tb.AddIconButton("view", "zoom", "+", "Zoom In", func() { ... })
//
// # LLM Prompt Template
//
//	tb := ui.NewToolbar("state-bar", 0, 0, 0, 54)
//	tb.Overflow = true
//	tb.AddGroup("state", "State")
//	tb.AddToggleLabel("state", "autosave", "Autosave", autoSave)
//
// DocumentSpec: `type`: `toolbar` with `toolbarGroup` children — see docs/DOCUMENT_SPEC_GO_RECIPES.md.
// Demo scenes: **Batch 9 Live Demo**, Notepad ribbon (Go), `pages/gallery.gru`.
type Toolbar struct {
	Element

	// Groups is the ordered list of item groups.
	Groups []*ToolbarGroup

	// Ribbon enables tab strip + section rows (see AddRibbonTab / AddSection).
	Ribbon bool

	// RibbonStacked lays out ribbon cells icon-above-caption (DevExpress-style).
	RibbonStacked bool

	// HideRibbonTabs suppresses the tab strip; ActiveGroup still selects the visible section.
	HideRibbonTabs bool

	// HideRibbonSectionLabels suppresses stacked section captions (File, Clipboard, …).
	HideRibbonSectionLabels bool

	// ribbonTabNames is the ordered tab strip; sections reference tabs by index.
	ribbonTabNames []string

	// ActiveGroup is the selected ribbon tab index.
	ActiveGroup *Signal[int]

	// Overflow enables automatic overflow handling.
	Overflow bool

	// OverflowKind selects scroll (default) vs "…" text menu — see docs/TOOLBAR_OVERFLOW.md.
	OverflowKind ToolbarOverflowKind

	// scrollX pans the item band when OverflowKind is ToolbarOverflowScroll.
	scrollX float32
	// contentWidth is the total logical width of the active item row (set in Layout).
	contentWidth float32
	// scrollActive when content is wider than the bar (reserves left/right gutter columns).
	scrollActive bool
	scrollLeftRect  rl.Rectangle
	scrollRightRect rl.Rectangle
	scrollLaneRect  rl.Rectangle // clip + hit-test band for scrolling items
	// scrollLeftShown hysteresis for the left scroll gutter (avoids 60 FPS redraw loops).
	scrollLeftShown bool

	// overflowFrom is the index (into the flat item list) of the first hidden item.
	// -1 means nothing is hidden.
	overflowFrom int

	// overflowOpen tracks whether the overflow popup is visible.
	overflowOpen bool

	// overflowRect is the bounds of the "..." button (set in Layout).
	overflowRect rl.Rectangle

	// tabRects holds the hit-rects for ribbon tabs (one per group).
	tabRects []rl.Rectangle

	// itemRects holds the resolved bounds of each visible item widget.
	// Indexed the same as a flat walk of all items across all groups.
	// Used by Draw/Update for overflow popup item highlighting.
	itemRects []rl.Rectangle

	// hoverOvfl tracks whether the mouse is over the overflow button.
	hoverOvfl bool

	// ovflHoverIdx is the index (into the overflow-hidden items) being hovered
	// inside the popup, or -1 when nothing is hovered.
	ovflHoverIdx int

	// hoverTab is the ribbon tab index being hovered, or -1.
	hoverTab int

	// sectionRects is each visible ribbon section's bounds (set in Layout).
	sectionRects []rl.Rectangle

	// sectionSeparatorX is the X coordinate of vertical dividers between ribbon sections.
	sectionSeparatorX []float32

	// ribbonDividerMidY is the vertical center of stacked icon cells (set in Layout).
	ribbonDividerMidY float32

	// ItemGap overrides horizontal spacing between items in a group (0 = default).
	ItemGap float32

	// lastLayoutW tracks bounds width from the prior Layout pass so ribbon tab
	// strip and scroll lane refresh when a flex parent stretches the bar on resize.
	lastLayoutW float32
}

// NewToolbar creates an empty Toolbar.
//
//	h — pass 0 to auto-size to tbDefaultH (or tbDefaultH + tbRibbonTabH when
//	    Ribbon is enabled).  Explicit h is used as-is.
func NewToolbar(id string, x, y, w, h float32) *Toolbar {
	if h == 0 {
		h = tbDefaultH
	}
	tb := &Toolbar{
		Element:      NewElement(id, x, y, w, h),
		ActiveGroup:  NewSignal(0),
		overflowFrom: -1,
		ovflHoverIdx: -1,
		hoverTab:     -1,
	}
	tb.styleName = "toolbar"

	tb.ActiveGroup.Subscribe(func() {
		tb.scrollX = 0
		tb.syncRibbonItemVisibility()
		tb.MarkDirty()
		tb.MarkDrawDirty()
	})

	return tb
}

func (tb *Toolbar) showRibbonTabs() bool {
	return tb.Ribbon && !tb.HideRibbonTabs
}

// SetHideRibbonTabs toggles the ribbon tab strip without disabling ribbon sections.
func (tb *Toolbar) SetHideRibbonTabs(hide bool) {
	if tb.HideRibbonTabs == hide {
		return
	}
	tb.HideRibbonTabs = hide
	tb.MarkDirty()
	tb.MarkDrawDirty()
}

// SetHideRibbonSectionLabels toggles stacked section captions under ribbon icon rows.
func (tb *Toolbar) SetHideRibbonSectionLabels(hide bool) {
	if tb.HideRibbonSectionLabels == hide {
		return
	}
	tb.HideRibbonSectionLabels = hide
	tb.MarkDirty()
	tb.MarkDrawDirty()
}

func (tb *Toolbar) showRibbonSectionLabels() bool {
	return tb.Ribbon && tb.RibbonStacked && !tb.HideRibbonSectionLabels
}

func (tb *Toolbar) ribbonSectionLabelHeight() float32 {
	if !tb.showRibbonSectionLabels() {
		return 0
	}
	h := EffectiveFontSize(GetThemeStyle("form-label"))
	if h < float32(tbSectionLabelH) {
		h = float32(tbSectionLabelH)
	}
	return h + 4
}

// ribbonCellScale ties stacked ribbon cell geometry to the active type scale (with
// a touch-target floor so icons stay tappable on narrow windows).
func (tb *Toolbar) ribbonCellScale() float32 {
	if !tb.Ribbon || !tb.RibbonStacked {
		return 1
	}
	cellStyle := GetThemeStyle("toolbar-cell")
	ref := float32(cellStyle.FontSize)
	if ref <= 0 {
		ref = 18
	}
	scale := EffectiveFontSize(cellStyle) / ref
	if scale < 0.88 {
		scale = 0.88
	}
	if scale > 1.12 {
		scale = 1.12
	}
	return scale
}

func (tb *Toolbar) stackedCellW() float32 {
	w := tbRibbonStackedCellW * tb.ribbonCellScale()
	if w < 48 {
		return 48
	}
	return w
}

func (tb *Toolbar) stackedCellH() float32 {
	h := tbRibbonStackedCellH * tb.ribbonCellScale()
	if h < 44 {
		return 44
	}
	return h
}

// ribbonCompactCustomH is the laid-out height for toolbar-menu TextInput/Dropdown cells (80% of icon cell).
func (tb *Toolbar) ribbonCompactCustomH() float32 {
	_, _, cellH, _ := tb.ribbonStackedMetrics()
	return cellH * tbRibbonCompactCustomScale
}

// ─── Group helpers ────────────────────────────────────────────────────────────

// AddGroup appends a flat command-bar group (TabIndex -1).
func (tb *Toolbar) AddGroup(id, label string) *ToolbarGroup {
	g := &ToolbarGroup{ID: id, Label: label, TabIndex: -1}
	tb.Groups = append(tb.Groups, g)
	tb.MarkDirty()
	return g
}

// AddRibbonTab registers a ribbon tab and returns its index.
func (tb *Toolbar) AddRibbonTab(name string) int {
	idx := len(tb.ribbonTabNames)
	tb.ribbonTabNames = append(tb.ribbonTabNames, name)
	tb.Ribbon = true
	tb.RibbonStacked = true
	tb.MarkDirty()
	return idx
}

// AddSection adds a ribbon section (group of cells) under tabIdx.
func (tb *Toolbar) AddSection(tabIdx int, id, sectionLabel string) *ToolbarGroup {
	g := &ToolbarGroup{ID: id, Label: sectionLabel, TabIndex: tabIdx}
	tb.Groups = append(tb.Groups, g)
	tb.MarkDirty()
	return g
}

// groupByID returns the first group whose ID matches, or nil.
func (tb *Toolbar) groupByID(id string) *ToolbarGroup {
	for _, g := range tb.Groups {
		if g.ID == id {
			return g
		}
	}
	return nil
}

// Widget returns the node for a toolbar item id, or nil.
func (tb *Toolbar) Widget(itemID string) Node {
	for _, g := range tb.Groups {
		for _, item := range g.items {
			if item.id == itemID {
				return item.widget
			}
		}
	}
	return nil
}

// SetItemSuppressed hides a toolbar item on the active ribbon tab without
// removing it from overflow accounting. Use for context-only cells (e.g. Back).
func (tb *Toolbar) SetItemSuppressed(itemID string, suppressed bool) {
	if tb == nil {
		return
	}
	for _, g := range tb.Groups {
		for _, item := range g.items {
			if item.id != itemID {
				continue
			}
			if item.suppressed == suppressed {
				return
			}
			item.suppressed = suppressed
			tb.syncRibbonItemVisibility()
			tb.MarkDirty()
			return
		}
	}
}

// ─── Item builder helpers ─────────────────────────────────────────────────────

// AddButton adds a Button item to the named group.  Returns tb for chaining.
func (tb *Toolbar) AddButton(groupID, id, label string, onClick func()) *Toolbar {
	btn := NewButton(id, label, 0, 0, 0, 0)
	btn.AutoHeight = false
	btn.SetStyle("toolbar-btn")
	btn.OnClick = onClick
	btn.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemButton, widget: btn, menuLabel: label})
	return tb
}

// AddPhosphorIconButton adds an IconButton with a Remix/Phosphor glyph. Returns tb for chaining.
//
// label is visible caption text (ribbon stacked cells). tooltip is hover help; when empty,
// label is used. For icon-only flat commands pass label "" and set tooltip explicitly.
func (tb *Toolbar) AddPhosphorIconButton(groupID, id, phosphorName, label, tooltip string, onClick func()) *Toolbar {
	btn := NewIconButton(id, "", label, 0, 0, 0, 0)
	btn.AutoHeight = false
	if tb.Ribbon && tb.RibbonStacked {
		btn.Stacked = true
		btn.SetStyle("toolbar-cell")
	} else {
		btn.SetStyle("toolbar-btn")
	}
	btn.SetPhosphorIcon(phosphorName, PhosphorRegular)
	btn.OnClick = onClick
	btn.SetParent(tb)
	tip := tooltip
	if tip == "" {
		tip = label
	}
	if tip != "" {
		SetTooltip(btn, tip)
	}
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemIconButton, widget: btn, menuLabel: tip})
	return tb
}

// AddPhosphorRibbonToggle adds a stacked ribbon cell that stays highlighted while
// checked is true. checked must mirror app state; onClick flips that state.
func (tb *Toolbar) AddPhosphorRibbonToggle(groupID, id, phosphorName, label, tooltip string, checked *Signal[bool], onClick func()) *Toolbar {
	btn := NewIconButton(id, "", label, 0, 0, 0, 0)
	btn.AutoHeight = false
	btn.Stacked = true
	btn.SetStyle("toolbar-cell")
	btn.SetPhosphorIcon(phosphorName, PhosphorRegular)
	btn.Checked = checked
	btn.ToggleCheckedOnClick = false
	if checked != nil {
		checked.Subscribe(func() { btn.MarkDrawDirty() })
	}
	btn.OnClick = onClick
	btn.SetParent(tb)
	tip := tooltip
	if tip == "" {
		tip = label
	}
	if tip != "" {
		SetTooltip(btn, tip)
	}
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemIconButton, widget: btn, menuLabel: tip})
	return tb
}

// syncRibbonItemVisibility hides widgets on inactive ribbon tabs so their checked
// interaction overlays do not bleed through when another tab is selected.
func (tb *Toolbar) syncRibbonItemVisibility() {
	if !tb.Ribbon || len(tb.ribbonTabNames) == 0 {
		return
	}
	activeIdx := tb.ActiveGroup.Get()
	if activeIdx < 0 || activeIdx >= len(tb.ribbonTabNames) {
		activeIdx = 0
	}
	for _, g := range tb.Groups {
		inTab := g.TabIndex == activeIdx
		for _, item := range g.items {
			if item.widget == nil {
				continue
			}
			show := inTab && !item.hidden && !item.suppressed
			if show {
				if item.widget.IsHidden() {
					item.widget.Show()
				}
			} else {
				if !item.widget.IsHidden() {
					item.widget.Hide()
				}
				if clearer, ok := item.widget.(overlayPointerClearer); ok {
					clearer.ClearOverlayPointerState()
				}
			}
		}
	}
}

// AddIconButton adds an IconButton item.  symbol is the glyph drawn as icon;
// label is the visible text.  Returns tb for chaining.
func (tb *Toolbar) AddIconButton(groupID, id, symbol, label string, onClick func()) *Toolbar {
	btn := NewIconButton(id, symbol, label, 0, 0, 0, 0)
	btn.AutoHeight = false
	if tb.Ribbon && tb.RibbonStacked {
		btn.Stacked = true
		btn.SetStyle("toolbar-cell")
	} else {
		btn.SetStyle("toolbar-btn")
	}
	btn.OnClick = onClick
	btn.SetParent(tb)
	tip := label
	if tip == "" {
		tip = id
	}
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemIconButton, widget: btn, menuLabel: tip})
	return tb
}

// AddToggleLabel adds a clickable word that toggles value (accent when on; no switch).
func (tb *Toolbar) AddToggleLabel(groupID, id, label string, value *Signal[bool]) *Toolbar {
	btn := NewButton(id, label, 0, 0, 0, 0)
	btn.AutoHeight = false
	btn.ToggleBinding = value
	btn.SetStyle("toolbar-toggle-label")
	btn.OnClick = func() {
		value.Set(!value.Get())
		btn.MarkDrawDirty()
	}
	value.Subscribe(func() { btn.MarkDrawDirty() })
	btn.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemButton, widget: btn})
	return tb
}

// AddSpinBox adds a compact − value + stepper for flat command bars.
func (tb *Toolbar) AddSpinBox(groupID, id string, value *Signal[float64], min, max, step float64) *Toolbar {
	sb := NewToolbarSpinBox(id, value, min, max, step)
	sb.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemCustom, widget: sb, menuLabel: "Value"})
	return tb
}

// AddWordToggle adds static label text with a compact switch (flat command bars).
func (tb *Toolbar) AddWordToggle(groupID, id, label string, value *Signal[bool]) *Toolbar {
	wt := NewToolbarWordToggle(id, label, value)
	wt.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemCustom, widget: wt, menuLabel: label})
	return tb
}

// AddMenuButton adds a toolbar menu control (fixed label + chevron, options below).
func (tb *Toolbar) AddMenuButton(groupID, id, faceLabel string, options []string, onSelect func(index int)) *Toolbar {
	dd := NewDropdown(id, options, 0, float32(0), float32(0), float32(0), float32(0))
	dd.AutoHeight = false
	dd.FaceLabel = faceLabel
	dd.SetStyle("toolbar-menu")
	if onSelect != nil {
		dd.SelectedIndex.Subscribe(func() { onSelect(dd.SelectedIndex.Get()) })
	}
	dd.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemCustom, widget: dd, menuLabel: faceLabel})
	return tb
}

// AddToggle adds a Toggle item.  value is the shared Signal; pass nil to create
// an internal one.  Returns tb for chaining.
func (tb *Toolbar) AddToggle(groupID, id string, value *Signal[bool]) *Toolbar {
	if value == nil {
		value = NewSignal(false)
	}
	tog := NewToggle(id, value.Get(), 0, 0, tbToggleW, tbToggleH)
	tog.Value = value
	tog.Value.Subscribe(func() {
		if tog.tween == nil {
			if tog.Value.Get() {
				tog.thumbOffset = 1
			} else {
				tog.thumbOffset = 0
			}
		}
		tog.MarkDirty()
	})
	tog.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemToggle, widget: tog})
	return tb
}

// AddSeparator adds a visual separator within the named group.
// Returns tb for chaining.
func (tb *Toolbar) AddSeparator(groupID, id string) *Toolbar {
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemSeparator, widget: nil})
	return tb
}

// AddCustom adds any Node as a toolbar item.  Returns tb for chaining.
func (tb *Toolbar) AddCustom(groupID, id string, widget Node) *Toolbar {
	widget.SetParent(tb)
	tb.addItem(groupID, &ToolbarItem{id: id, itemType: ToolbarItemCustom, widget: widget})
	return tb
}

// addItem appends item to the group identified by groupID; if no such group
// exists a new one is created automatically.
func (tb *Toolbar) addItem(groupID string, item *ToolbarItem) {
	g := tb.groupByID(groupID)
	if g == nil {
		g = tb.AddGroup(groupID, groupID)
	}
	g.items = append(g.items, item)
	tb.MarkDirty()
}

// ─── Node interface ───────────────────────────────────────────────────────────

// IsInteractive always returns true for Toolbar.
func (tb *Toolbar) IsInteractive() bool { return true }

// Children returns all item widgets that are Nodes (for the inspector / tree walk).
func (tb *Toolbar) Children() []Node {
	var out []Node
	for _, g := range tb.Groups {
		for _, item := range g.items {
			if item.widget != nil {
				out = append(out, item.widget)
			}
		}
	}
	return out
}

// AddChild satisfies the Node interface but is a no-op — use the typed
// AddButton / AddToggle / other typed helpers instead.
func (tb *Toolbar) AddChild(_ Node) {}

// innerClipRect is the padded interior used for scissor and interaction overlays.
func (tb *Toolbar) innerClipRect() rl.Rectangle {
	b := tb.bounds
	style := tb.GetStyle()
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	return rl.NewRectangle(b.X+bw, b.Y+bw, b.Width-2*bw, b.Height-2*bw)
}

// itemGap returns horizontal spacing between items in a group.
func (tb *Toolbar) itemGap() float32 {
	if tb.ItemGap > 0 {
		return tb.ItemGap
	}
	if tb.Ribbon && tb.RibbonStacked {
		return tbRibbonItemGap
	}
	return tbItemGap
}

// flatRowMetrics returns the item band inside toolbar chrome (below ribbon tabs when set).
func (tb *Toolbar) flatRowMetrics() (itemY, itemH float32) {
	inner := tb.innerClipRect()
	itemY = inner.Y
	itemH = inner.Height
	if tb.showRibbonTabs() {
		itemY += tbRibbonTabH
		itemH -= tbRibbonTabH
	}
	if itemH < 1 && !tb.Ribbon {
		itemH = tbDefaultH
	}
	return itemY, itemH
}

// ribbonStackedMetrics splits the ribbon item band into even top/bottom pad, cell, and caption gap.
func (tb *Toolbar) ribbonStackedMetrics() (bandY, bandPad, cellH, labelGap float32) {
	itemY, bandH := tb.flatRowMetrics()
	bandY = itemY
	cellH = tb.stackedCellH()
	labelGap = tbRibbonRowGap
	if !tb.showRibbonSectionLabels() {
		labelGap = 0
	}
	labelH := tb.ribbonSectionLabelHeight()
	rest := bandH - cellH - labelGap - labelH
	bandPad = rest / 2
	if bandPad < 0 {
		bandPad = 0
	}
	return bandY, bandPad, cellH, labelGap
}

// overflowButtonRect returns the "…" button in current toolbar bounds (scroll-safe).
func (tb *Toolbar) overflowButtonRect() rl.Rectangle {
	if !tb.Overflow || tb.overflowFrom < 0 {
		return rl.Rectangle{}
	}
	b := tb.bounds
	itemY, itemH := tb.flatRowMetrics()
	ow := tbOvflW
	oh := itemH - 2*tbRowEdgePad
	if oh < tbToggleH {
		oh = tbToggleH
	}
	ox := b.X + b.Width - ow - tbPad
	oy := itemY + (itemH-oh)/2
	return snapLayoutRect(rl.NewRectangle(ox, oy, ow, oh))
}

// ─── Update ───────────────────────────────────────────────────────────────────

// Update handles mouse interaction for ribbon tabs, the overflow button,
// the overflow popup, and all visible item widgets.
func (tb *Toolbar) Update(dt float32) {
	if tb.IsHidden() {
		return
	}

	mouse := rl.GetMousePosition()
	prevHoverTab := tb.hoverTab
	prevHoverOvfl := tb.hoverOvfl
	prevOverflowOpen := tb.overflowOpen
	prevOvflHoverIdx := tb.ovflHoverIdx

	// ── Ribbon tab clicks ────────────────────────────────────────────────────
	if tb.showRibbonTabs() {
		tb.hoverTab = -1
		for i, r := range tb.tabRects {
			if rl.CheckCollisionPointRec(mouse, r) {
				tb.hoverTab = i
				if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
					tb.ActiveGroup.Set(i)
					tb.overflowOpen = false
				}
			}
		}
	}

	// ── Overflow button (menu mode only) ─────────────────────────────────────
	if tb.usesOverflowMenu() && tb.overflowFrom >= 0 {
		ovflBtn := tb.overflowButtonRect()
		tb.overflowRect = ovflBtn
		tb.hoverOvfl = rl.CheckCollisionPointRec(mouse, ovflBtn)
		if tb.hoverOvfl && rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			tb.overflowOpen = !tb.overflowOpen
		}
	} else {
		tb.overflowOpen = false
		tb.hoverOvfl = false
	}

	// ── Horizontal scroll (wheel + chevrons) ─────────────────────────────────
	if tb.usesHorizontalScroll() && rl.CheckCollisionPointRec(mouse, tb.Bounds()) {
		tb.updateHorizontalScroll(mouse)
	}
	if tb.mouseOnScrollGutter(mouse) {
		// Clear item hover when over chevrons/gutter, but do not return —
		// compound controls (Dropdown, SpinBox) still need Update this frame.
		tb.clearToolbarItemPointerState()
	}

	// ── Overflow popup item hover / click (menu mode only) ────────────────────
	if tb.overflowOpen {
		hiddenItems := tb.hiddenItems()
		popup := tb.overflowPopupRect(hiddenItems)
		tb.ovflHoverIdx = -1
		for i, item := range hiddenItems {
			r := rl.NewRectangle(popup.X, popup.Y+float32(i)*tbOvflRowH, popup.Width, tbOvflRowH)
			if rl.CheckCollisionPointRec(mouse, r) {
				tb.ovflHoverIdx = i
				if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
					row := rl.NewRectangle(r.X+6, r.Y+4, r.Width-12, tbOvflRowH-8)
					tb.activateOverflowItem(item, row)
					tb.overflowOpen = false
				}
			}
		}
		// Dismiss popup on click outside.
		if rl.IsMouseButtonPressed(rl.MouseButtonLeft) && tb.ovflHoverIdx < 0 && !tb.hoverOvfl {
			tb.overflowOpen = false
		}
		if tb.hoverTab != prevHoverTab || tb.hoverOvfl != prevHoverOvfl || tb.overflowOpen != prevOverflowOpen || tb.ovflHoverIdx != prevOvflHoverIdx {
			tb.MarkDrawDirty()
		}
		if tb.overflowOpen {
			return // menu open — don't hit hidden items under the bar
		}
	}

	if tb.hoverTab != prevHoverTab || tb.hoverOvfl != prevHoverOvfl || tb.overflowOpen != prevOverflowOpen || tb.ovflHoverIdx != prevOvflHoverIdx {
		tb.MarkDrawDirty()
	}

	// ── Delegate to visible item widgets ──────────────────────────────────────
	for _, g := range tb.activeGroups() {
		for _, item := range g.items {
			if item.hidden || item.suppressed || item.widget == nil || item.widget.IsHidden() {
				continue
			}
			item.widget.Update(dt)
		}
	}
}

// ─── Layout ───────────────────────────────────────────────────────────────────

// Layout positions all item widgets within the toolbar bounds.
func (tb *Toolbar) Layout() {
	if tb.IsHidden() {
		tb.layoutDirty = false
		return
	}
	if !tb.IsDirty() {
		return
	}

	b := tb.bounds
	if tb.lastLayoutW > 0 && absF(b.Width-tb.lastLayoutW) > 0.5 {
		tb.scrollX = 0
		tb.MarkDrawDirty()
	}
	tb.lastLayoutW = b.Width

	itemY, itemH := tb.flatRowMetrics()

	groups := tb.activeGroups()
	tb.sectionRects = make([]rl.Rectangle, 0, len(groups))
	sepCap := 0
	if len(groups) > 1 {
		sepCap = len(groups) - 1
	}
	tb.sectionSeparatorX = make([]float32, 0, sepCap)
	type itemMeta = toolbarItemMeta
	var flat []itemMeta
	for _, g := range groups {
		for _, item := range g.items {
			if item.suppressed {
				continue
			}
			flat = append(flat, itemMeta{item, tb.naturalWidth(item, itemH)})
		}
	}

	inner := tb.innerClipRect()
	tbScrollBW := tb.GetStyle().BorderWidth
	if tbScrollBW <= 0 {
		tbScrollBW = 1
	}

	scrollViewport := tb.ribbonScrollViewport(inner, tbScrollBW)
	tb.ribbonFlexCompactItems(flat, groups, scrollViewport)

	// Measure total width needed (ribbon uses section/compact gaps).
	totalW := tb.measureFlatTotalW(flat, groups)

	menuMode := tb.usesOverflowMenu()
	tb.overflowFrom = -1
	tb.scrollActive = false
	if tb.usesHorizontalScroll() && totalW > scrollViewport+0.5 {
		tb.scrollActive = true
	}
	itemStartX := inner.X + tbScrollBW + tbPad
	if !tb.scrollActive {
		itemStartX = inner.X + tbPad
	}
	if tb.scrollActive {
		tb.refreshScrollChrome()
		if !menuMode {
			tb.scrollX = tb.clampScrollXValue(tb.scrollX)
		}
	}
	cumW := snapLayoutUnit(itemStartX - b.X)
	flatIdx := 0
	tb.itemRects = make([]rl.Rectangle, len(flat))

	tb.ribbonDividerMidY = 0
	for gi, g := range groups {
		if gi > 0 {
			if tb.Ribbon && tb.RibbonStacked {
				gapStart := b.X + cumW
				sepX := snapLayoutUnit(gapStart + tbRibbonSectionGap/2)
				tb.sectionSeparatorX = append(tb.sectionSeparatorX, sepX)
				cumW = gapStart - b.X + tbRibbonSectionGap
			} else {
				cumW += tbGroupGap*2 + tbSepW
			}
		}
		var sectionFirst, sectionLast rl.Rectangle
		hasSectionItem := false
		groupItems := tb.itemsForGroupMeta(flat, g)
		for ii, im := range groupItems {
			if ii > 0 {
				cumW += tb.ribbonItemGapBetween(groupItems[ii-1].item, im.item)
			}
			rx := snapLayoutUnit(b.X + cumW)
			rw := im.w
			var rh, ry float32
			// Final height first, then vertical centre — never seed ry from item
			// width (that pushed wide Find/menus outside the bar and staggered
			// overflow buttons by label length).
			if im.item.itemType == ToolbarItemSeparator {
				rh = tbSepH
				ry = itemY + (itemH-rh)/2
			} else if im.item.itemType == ToolbarItemToggle {
				rw = tbToggleW
				rh = tbToggleH
				ry = itemY + (itemH-rh)/2
			} else if tb.Ribbon && tb.RibbonStacked {
				_, bandPad, cellH, _ := tb.ribbonStackedMetrics()
				if tb.isCompactRibbonCustom(im.item) {
					rh = tb.ribbonCompactCustomH()
					ry = itemY + bandPad + (cellH-rh)/2
				} else {
					rh = cellH
					ry = itemY + bandPad
					if im.item.itemType == ToolbarItemIconButton || im.item.itemType == ToolbarItemButton {
						if rw < tb.stackedCellW() {
							rw = tb.stackedCellW()
						}
					}
				}
			} else {
				rh = itemH - 2*tbRowEdgePad
				if rh < tbToggleH {
					rh = itemH
				}
				ry = itemY + (itemH-rh)/2
			}

			rect := snapLayoutRect(rl.NewRectangle(rx, ry, rw, rh))
			tb.itemRects[flatIdx] = rect
			if tb.Ribbon && tb.RibbonStacked && tb.ribbonDividerMidY == 0 &&
				(im.item.itemType == ToolbarItemButton || im.item.itemType == ToolbarItemIconButton) {
				tb.ribbonDividerMidY = rect.Y + rect.Height/2
			}

			if tb.overflowFrom < 0 || flatIdx < tb.overflowFrom {
				if !hasSectionItem {
					sectionFirst = rect
					hasSectionItem = true
				}
				sectionLast = rect
			}

			im.item.hidden = false

			cumW = rect.X - b.X + rect.Width
			flatIdx++
		}
		if tb.Ribbon && tb.RibbonStacked && hasSectionItem {
			tb.sectionRects = append(tb.sectionRects, rl.NewRectangle(
				sectionFirst.X, itemY,
				(sectionLast.X+sectionLast.Width)-sectionFirst.X,
				itemH,
			))
		}
	}

	if menuMode && len(tb.itemRects) > 0 {
		inner := tb.innerClipRect()
		limit := inner.X + inner.Width - tbOvflW - tb.itemGap()
		last := tb.itemRects[len(tb.itemRects)-1]
		needMenu := last.X+last.Width > inner.X+inner.Width+0.5
		tb.overflowFrom = -1
		if needMenu {
			for i, r := range tb.itemRects {
				if r.X+r.Width > limit+0.5 {
					tb.overflowFrom = i
					break
				}
			}
			if tb.overflowFrom < 0 {
				tb.overflowFrom = len(tb.itemRects)
			}
		}
		for i := range flat {
			if i >= len(tb.itemRects) {
				break
			}
			flat[i].item.hidden = tb.overflowFrom >= 0 && i >= tb.overflowFrom
		}
	}

	tb.contentWidth = cumW + tbPad
	if tb.usesHorizontalScroll() {
		if tb.scrollActive {
			tb.refreshScrollChrome()
			tb.applyScrollBounds()
		} else {
			tb.scrollX = 0
			for i := range flat {
				item := flat[i].item
				if item.hidden || item.widget == nil {
					continue
				}
				r := tb.itemRects[i]
				if el, ok := item.widget.(interface{ setBoundsNoMark(rl.Rectangle) }); ok {
					el.setBoundsNoMark(r)
				} else {
					item.widget.SetBounds(r)
				}
				item.widget.Layout()
			}
		}
	} else {
		for i := range flat {
			item := flat[i].item
			if item.hidden || item.widget == nil {
				continue
			}
			item.widget.SetBounds(tb.itemRects[i])
			item.widget.Layout()
		}
	}

	if tb.usesOverflowMenu() {
		tb.overflowRect = tb.overflowButtonRect()
	}

	if tb.Ribbon && tb.showRibbonTabs() {
		if len(tb.ribbonTabNames) > 0 {
			tb.layoutRibbonTabs(b)
		} else {
			tb.layoutRibbonTabsLegacy(b)
		}
	}

	tb.syncRibbonItemVisibility()

	_ = totalW // suppress unused warning
	tb.layoutDirty = false
}

// layoutRibbonTabsLegacy positions tabs when each ToolbarGroup is its own tab.
func (tb *Toolbar) layoutRibbonTabsLegacy(b rl.Rectangle) {
	inner := tb.innerClipRect()
	tb.tabRects = make([]rl.Rectangle, len(tb.Groups))
	x := inner.X + tbPad
	for i, g := range tb.Groups {
		tw := float32(measureTextS(g.Label, GetThemeStyle("toolbar-ribbon-tab"))) + 24
		tb.tabRects[i] = rl.NewRectangle(x, inner.Y, tw, tbRibbonTabH)
		x += tw + 4
	}
}

func (tb *Toolbar) layoutRibbonTabs(b rl.Rectangle) {
	inner := tb.innerClipRect()
	tb.tabRects = make([]rl.Rectangle, len(tb.ribbonTabNames))
	x := snapLayoutUnit(inner.X + tbPad)
	tabStyle := GetThemeStyle("toolbar-ribbon-tab")
	tabGap := float32(4)
	for i, name := range tb.ribbonTabNames {
		tw := snapLayoutUnit(float32(measureTextS(name, tabStyle)) + 24)
		tb.tabRects[i] = snapLayoutRect(rl.NewRectangle(x, inner.Y, tw, tbRibbonTabH))
		x = tb.tabRects[i].X + tb.tabRects[i].Width + tabGap
	}
}

type toolbarItemMeta struct {
	item *ToolbarItem
	w    float32
}

func (tb *Toolbar) ribbonItemGapBetween(prev, cur *ToolbarItem) float32 {
	gap := tb.itemGap()
	if tb.Ribbon && tb.RibbonStacked {
		if tb.isCompactRibbonCustom(prev) || tb.isCompactRibbonCustom(cur) {
			return tbRibbonCompactItemGap
		}
	}
	return gap
}

func (tb *Toolbar) compactRibbonMinW(item *ToolbarItem) float32 {
	minW := tbRibbonCompactMinW
	if item != nil && item.widget != nil {
		if m, ok := item.widget.(interface{ GetMinWidth() float32 }); ok {
			if v := m.GetMinWidth(); v > 0 {
				minW = v
			}
		}
	}
	return minW
}

func (tb *Toolbar) measureFlatTotalW(flat []toolbarItemMeta, groups []*ToolbarGroup) float32 {
	totalW := tbPad * 2
	for gi, g := range groups {
		if gi > 0 {
			if tb.Ribbon && tb.RibbonStacked {
				totalW += tbRibbonSectionGap
			} else {
				totalW += tbGroupGap*2 + tbSepW
			}
		}
		groupItems := tb.itemsForGroupMeta(flat, g)
		for ii, im := range groupItems {
			if ii > 0 {
				totalW += tb.ribbonItemGapBetween(groupItems[ii-1].item, im.item)
			}
			totalW += im.w
		}
	}
	return totalW
}

func (tb *Toolbar) itemsForGroupMeta(flat []toolbarItemMeta, g *ToolbarGroup) []toolbarItemMeta {
	var out []toolbarItemMeta
	for _, im := range flat {
		for _, gi := range g.items {
			if gi == im.item {
				out = append(out, im)
				break
			}
		}
	}
	return out
}

// ribbonScrollViewport is the horizontal budget for ribbon items before scroll gutters.
// Flex and scroll activation both use this so Search (find bar + icons) matches other tabs.
func (tb *Toolbar) ribbonScrollViewport(inner rl.Rectangle, bw float32) float32 {
	v := inner.Width - 2*bw - 2*tbPad
	if tb.usesHorizontalScroll() {
		v -= tb.scrollGutterReserve()
	}
	if v < 1 {
		return 1
	}
	return v
}

// ribbonFlexCompactItems shrinks toolbar-menu fields before enabling ribbon scroll.
func (tb *Toolbar) ribbonFlexCompactItems(flat []toolbarItemMeta, groups []*ToolbarGroup, viewport float32) {
	if !tb.Ribbon || !tb.RibbonStacked {
		return
	}
	totalW := tb.measureFlatTotalW(flat, groups)
	if totalW <= viewport+0.5 {
		return
	}
	deficit := totalW - viewport
	var flexIdx []int
	var flexSum, minSum float32
	for i := range flat {
		if !tb.isCompactRibbonCustom(flat[i].item) {
			continue
		}
		flexIdx = append(flexIdx, i)
		flexSum += flat[i].w
		minSum += tb.compactRibbonMinW(flat[i].item)
	}
	shrinkable := flexSum - minSum
	if shrinkable <= 0 || deficit <= 0 {
		return
	}
	shrink := deficit
	if shrink > shrinkable {
		shrink = shrinkable
	}
	scale := (flexSum - shrink) / flexSum
	for _, i := range flexIdx {
		minW := tb.compactRibbonMinW(flat[i].item)
		flat[i].w = max32(flat[i].w*scale, minW)
	}
}

// isCompactRibbonCustom reports toolbar-menu controls that stay short in stacked ribbon bands.
func (tb *Toolbar) isCompactRibbonCustom(item *ToolbarItem) bool {
	if item == nil || item.itemType != ToolbarItemCustom || item.widget == nil {
		return false
	}
	switch w := item.widget.(type) {
	case *Dropdown:
		return w.styleName == "toolbar-menu"
	case *TextInput:
		return w.styleName == "toolbar-menu"
	case *ToolbarSpinBox:
		return true
	default:
		return false
	}
}

// naturalWidth returns the preferred width of a toolbar item.
func (tb *Toolbar) naturalWidth(item *ToolbarItem, _ float32) float32 {
	switch item.itemType {
	case ToolbarItemSeparator:
		return tbSepW + 8
	case ToolbarItemToggle:
		return tbToggleW
	case ToolbarItemCustom:
		if dd, ok := item.widget.(*Dropdown); ok {
			if dd.PreferredWidth > 0 {
				w := dd.PreferredWidth
				if mx := dd.MaxWidth; mx > 0 && w > mx {
					w = mx
				}
				return max32(w, tbMinItemW)
			}
			label := dd.FaceLabel
			if label == "" && len(dd.Options) > 0 {
				label = dd.Options[dd.SelectedIndex.Get()]
			}
			st := dd.GetStyle()
			return max32(buttonNaturalWidth(label, st)+38, tbMinItemW)
		}
		if cb, ok := item.widget.(*ComboBox); ok {
			w := cb.PreferredWidth
			if w <= 0 {
				w = 140
			}
			return max32(w, tbMinItemW)
		}
		if ti, ok := item.widget.(*TextInput); ok {
			w := ti.PreferredWidth
			if w <= 0 {
				w = 140
			}
			if mx := ti.MaxWidth; mx > 0 && w > mx {
				w = mx
			}
			return max32(w, tbMinItemW)
		}
		if wt, ok := item.widget.(*ToolbarWordToggle); ok {
			return max32(wt.NaturalWidth(), tbMinItemW)
		}
		if sb, ok := item.widget.(*ToolbarSpinBox); ok {
			return max32(sb.NaturalWidth(), tbMinItemW)
		}
		if item.widget != nil {
			if pw, ok := flexChildPreferredWidth(item.widget); ok {
				return max32(pw, tbMinItemW)
			}
			if w := item.widget.Bounds().Width; w > 0 {
				return max32(w, tbMinItemW)
			}
		}
		return 60
	default:
		if item.widget != nil {
			if ib, ok := item.widget.(*IconButton); ok && ib.Stacked {
				return tb.stackedCellW()
			}
			if btn, ok := item.widget.(*Button); ok {
				w := buttonNaturalWidth(btn.Text.Get(), btn.GetStyle())
				if btn.styleName == "toolbar-btn" {
					w = buttonToolbarNaturalWidth(btn.Text.Get(), btn.GetStyle())
				}
				return max32(w, tbMinItemW)
			}
			if ibtn, ok := item.widget.(*IconButton); ok {
				if ibtn.Label.Get() == "" {
					return tbIconOnlyW
				}
				tw := float32(measureTextS(ibtn.Label.Get(), ibtn.GetStyle()))
				sw := float32(measureText(ibtn.Symbol, 16))
				return max32(sw+tw+24, tbMinItemW)
			}
		}
		return 60
	}
}

// activeGroups returns groups visible in the current layout pass.
func (tb *Toolbar) activeGroups() []*ToolbarGroup {
	if !tb.Ribbon {
		return tb.Groups
	}
	if len(tb.ribbonTabNames) == 0 {
		idx := tb.ActiveGroup.Get()
		if idx < 0 || idx >= len(tb.Groups) {
			idx = 0
		}
		if len(tb.Groups) == 0 {
			return nil
		}
		return tb.Groups[idx : idx+1]
	}
	idx := tb.ActiveGroup.Get()
	if idx < 0 || idx >= len(tb.ribbonTabNames) {
		idx = 0
	}
	out := make([]*ToolbarGroup, 0, 4)
	for _, g := range tb.Groups {
		if g.TabIndex == idx {
			out = append(out, g)
		}
	}
	return out
}

// hiddenItems returns the flat list of items that have been hidden by overflow.
func (tb *Toolbar) hiddenItems() []*ToolbarItem {
	var out []*ToolbarItem
	for _, g := range tb.activeGroups() {
		for _, item := range g.items {
			if item.hidden {
				out = append(out, item)
			}
		}
	}
	return out
}

// ─── Draw ─────────────────────────────────────────────────────────────────────

// Draw renders the toolbar background, ribbon tabs (if enabled), item widgets,
// group separators, overflow button, and (if open) the overflow popup.
func (tb *Toolbar) Draw() {
	defer func() { tb.drawDirty = false }()
	if tb.IsHidden() {
		return
	}

	b := tb.bounds
	style := tb.GetStyle()
	sepStyle := GetThemeStyle("toolbar-separator")

	borderCol := style.BorderColor
	if borderCol.A == 0 {
		borderCol = rl.NewColor(222, 226, 230, 255)
	}
	inner := snapControlRect(b)
	if tb.Ribbon {
		// Flush under title bar: fill only + bottom hairline (no top/double rule).
		rl.DrawRectangleRec(inner, style.BackgroundColor)
		rl.DrawLineEx(
			rl.NewVector2(inner.X, inner.Y+inner.Height-0.5),
			rl.NewVector2(inner.X+inner.Width, inner.Y+inner.Height-0.5),
			1, borderCol,
		)
	} else {
		// ── Background (inset border — matches Checkbox / SearchBar) ───────────────
		r := style.CornerRadius
		roundness := float32(0)
		if r > 0 && b.Height > 0 {
			shorter := b.Width
			if b.Height < shorter {
				shorter = b.Height
			}
			roundness = r / (shorter / 2)
			if roundness > 1 {
				roundness = 1
			}
		}
		bw := style.BorderWidth
		if bw <= 0 {
			bw = 1
		}
		drawRoundedInsetBorder(inner, roundness, bw, borderCol, style.BackgroundColor)
	}

	itemY, itemH := tb.flatRowMetrics()
	groups := tb.activeGroups()
	innerClip := tb.innerClipRect()

	if tb.showRibbonTabs() {
		tb.drawRibbonTabs()
	}

	withClipRestore(tb, innerClip, func() {
		laneClip := innerClip
		if tb.scrollActive {
			tb.refreshScrollChrome()
			if tb.scrollLaneRect.Width > 0 {
				laneClip = tb.scrollLaneRect
			}
		}
		drawItems := func(clip rl.Rectangle) {
			clip = intersectRectsWithViewportAncestors(clip, tb)
			if clip.Width < 1 || clip.Height < 1 {
				return
			}
			beginScissorFromRect(clip)
			setActiveDrawClip(clip)
			for gi, g := range groups {
				if gi > 0 && !(tb.Ribbon && tb.RibbonStacked) {
					firstRect := tb.firstVisibleItemRect(g)
					if firstRect.Width > 0 {
						sepX := firstRect.X - tbGroupGap - tbSepW/2
						sepMidY := itemY + itemH/2
						rl.DrawLineEx(
							rl.NewVector2(sepX, sepMidY-tbSepH/2),
							rl.NewVector2(sepX, sepMidY+tbSepH/2),
							tbSepW, sepStyle.BackgroundColor,
						)
					}
				}

				for _, item := range g.items {
					if item.hidden || item.suppressed {
						continue
					}
					if item.widget != nil && item.widget.IsHidden() {
						continue
					}
					if item.itemType == ToolbarItemSeparator {
						if item.widget == nil {
							continue
						}
						r2 := item.widget.Bounds()
						if r2.X+r2.Width < clip.X || r2.X > clip.X+clip.Width {
							continue
						}
						midX := r2.X + r2.Width/2
						midY := itemY + itemH/2
						if tb.Ribbon && tb.RibbonStacked {
							midY = r2.Y + r2.Height/2
						}
						rl.DrawLineEx(
							rl.NewVector2(midX, midY-tbSepH/2),
							rl.NewVector2(midX, midY+tbSepH/2),
							tbSepW, sepStyle.BackgroundColor,
						)
						continue
					}
					if item.widget != nil {
						wb := item.widget.Bounds()
						// Fully outside the lane — skip (avoids EndScissor leaks
						// painting Sync/Clear over the Panel).
						if wb.X+wb.Width < clip.X+0.5 || wb.X > clip.X+clip.Width-0.5 {
							continue
						}
						item.widget.Draw()
						beginScissorFromRect(clip)
						setActiveDrawClip(clip)
					}
				}
			}
			tb.drawRibbonSections(groups)
			if tb.usesOverflowMenu() && tb.overflowFrom >= 0 {
				tb.drawOverflowButton()
			}
		}
		drawItems(laneClip)
		if tb.scrollActive {
			// Gutters sit on the chrome; clip to inner so carets cannot leave the bar.
			beginScissorFromRect(intersectRectsWithViewportAncestors(innerClip, tb))
			setActiveDrawClip(innerClip)
			tb.drawScrollGutters()
		}
	})

	// ── Overflow popup ────────────────────────────────────────────────────────
	if tb.overflowOpen && findViewport(tb) == nil {
		tb.drawOverflowPopup()
	}
}

// drawRibbonTabs renders the tab row (Bootstrap / DevExpress ribbon tab strip).
func (tb *Toolbar) drawRibbonTabs() {
	inner := tb.innerClipRect()
	active := tb.ActiveGroup.Get()
	tabStyle := GetThemeStyle("toolbar-ribbon-tab")
	activeStyle := GetThemeStyle("toolbar-ribbon-tab-active")
	ribbonStyle := GetThemeStyle("toolbar-ribbon")
	stripCol := ribbonStyle.BackgroundColor
	if stripCol.A == 0 {
		stripCol = rl.NewColor(241, 243, 247, 255)
	}

	rl.DrawRectangleRec(rl.NewRectangle(inner.X, inner.Y, inner.Width, tbRibbonTabH), stripCol)

	names := tb.ribbonTabNames
	if len(names) == 0 {
		for _, g := range tb.Groups {
			names = append(names, g.Label)
		}
	}

	for i, name := range names {
		if i >= len(tb.tabRects) {
			break
		}
		r := tb.tabRects[i]
		isActive := i == active

		ts := tabStyle
		if isActive {
			ts.TextColor = activeStyle.TextColor
			ts.Bold = activeStyle.Bold
		}

		textW := float32(measureTextS(name, ts))
		tx := int32(r.X + (r.Width-textW)/2)
		ty := TextPosY(r, ts)
		drawTextS(name, tx, ty, ts)
	}

	// Tab/content separation uses ribbon background only — title bar supplies the top hairline.
}

func (tb *Toolbar) drawRibbonSections(groups []*ToolbarGroup) {
	if !tb.Ribbon || !tb.RibbonStacked {
		return
	}
	capStyle := GetThemeStyle("form-label")
	sepStyle := GetThemeStyle("toolbar-separator")
	for i, g := range groups {
		if i >= len(tb.sectionRects) {
			break
		}
		sr := tb.sectionRects[i]
		scroll := float32(0)
		if tb.usesHorizontalScroll() {
			scroll = tb.scrollX
		}
		drawSR := sr
		drawSR.X -= scroll
		if i > 0 && i-1 < len(tb.sectionSeparatorX) {
			sepX := tb.sectionSeparatorX[i-1] - scroll
			midY := tb.ribbonDividerMidY
			if midY <= 0 {
				_, bandPad, cellH, _ := tb.ribbonStackedMetrics()
				bandY, _ := tb.flatRowMetrics()
				midY = bandY + bandPad + cellH/2
			}
			sepH := tb.stackedCellH() - 8
			if sepH > tbSepH {
				sepH = tbSepH
			}
			rl.DrawLineEx(
				rl.NewVector2(sepX, midY-sepH/2),
				rl.NewVector2(sepX, midY+sepH/2),
				tbSepW, sepStyle.BackgroundColor,
			)
		}
		if g.Label == "" || !tb.showRibbonSectionLabels() {
			continue
		}
		textW := float32(measureTextS(g.Label, capStyle))
		tx := int32(drawSR.X + (drawSR.Width-textW)/2)
		bandY, bandPad, cellH, labelGap := tb.ribbonStackedMetrics()
		labelBandY := bandY + bandPad + cellH + labelGap
		labelBandH := tb.ribbonSectionLabelHeight()
		if labelBandH < 1 {
			continue
		}
		ty := int32(labelBandY + (labelBandH-EffectiveFontSize(capStyle))/2)
		drawTextS(g.Label, tx, ty, capStyle)
	}
}

// drawOverflowButton renders the "..." button at the right edge of the toolbar.
func (tb *Toolbar) drawOverflowButton() {
	r := tb.overflowButtonRect()
	bg := rl.NewColor(255, 255, 255, 255)
	border := rl.NewColor(222, 226, 230, 255)
	if tb.hoverOvfl || tb.overflowOpen {
		bg = rl.NewColor(243, 244, 246, 255)
	}
	drawRoundedInsetBorder(snapControlRect(r), 0.35, 1, border, bg)
	dotStyle := GetThemeStyle("toolbar-btn")
	content := toolbarContentRect(r, dotStyle)
	fs := EffectiveFontSize(dotStyle)
	textW := float32(measureTextF("...", fs, false, false, false, false))
	drawTextF("...", content.X+(content.Width-textW)/2, content.Y+(content.Height-fs)/2, fs, rl.NewColor(80, 82, 110, 255), false, false, false, false)
}

// drawOverflowPopup renders an inline popup listing hidden item labels.
func (tb *Toolbar) drawOverflowPopup() {
	hidden := tb.hiddenItems()
	if len(hidden) == 0 {
		return
	}

	bg := tb.overflowPopupRect(hidden)
	menuStyle := GetThemeStyle("contextmenu")
	panelBg := menuStyle.BackgroundColor
	panelBorder := menuStyle.BorderColor
	if panelBg.A == 0 {
		panelBg = rl.NewColor(255, 255, 255, 255)
	}
	if panelBorder.A == 0 {
		panelBorder = rl.NewColor(210, 213, 228, 255)
	}
	rl.DrawRectangleRounded(bg, 0.15, 4, panelBg)
	rl.DrawRectangleRoundedLinesEx(bg, 0.15, 4, 1, panelBorder)

	rowStyle := GetThemeStyle("toolbar-btn")
	hoverStyle := GetThemeStyle("contextmenu-hover")
	for i, item := range hidden {
		ry := bg.Y + float32(i)*tbOvflRowH
		row := rl.NewRectangle(bg.X+6, ry+4, bg.Width-12, tbOvflRowH-8)
		if i == tb.ovflHoverIdx {
			hoverBg := hoverStyle.BackgroundColor
			if hoverBg.A == 0 {
				hoverBg = rl.NewColor(243, 244, 246, 255)
			}
			rl.DrawRectangleRounded(row, 0.28, 6, hoverBg)
		}
		label := tb.itemLabel(item)
		textColor := rowStyle.TextColor
		if i == tb.ovflHoverIdx {
			textColor = rl.NewColor(79, 70, 229, 255)
		}
		tx := int32(bg.X) + 14
		ty := int32(ry) + (int32(tbOvflRowH)-int32(EffectiveFontSize(rowStyle)))/2
		rowDraw := rowStyle
		rowDraw.TextColor = textColor
		drawTextS(label, tx, ty, rowDraw)
	}
}

func (tb *Toolbar) overflowPopupRect(hidden []*ToolbarItem) rl.Rectangle {
	popupW := float32(160)
	rowStyle := GetThemeStyle("toolbar-btn")
	fs := EffectiveFontSize(rowStyle)
	for _, item := range hidden {
		w := measureTextF(tb.itemLabel(item), fs, false, false, false, false) + 32
		if w > popupW {
			popupW = w
		}
	}
	popupH := float32(len(hidden)) * tbOvflRowH
	ovfl := tb.overflowButtonRect()
	popupX := ovfl.X + ovfl.Width - popupW
	popupY := ovfl.Y + ovfl.Height + 4
	return rl.NewRectangle(popupX, popupY, popupW, popupH)
}

// firstVisibleItemRect returns the bounds of the first non-hidden item in g,
// or zero-value if none.
func (tb *Toolbar) firstVisibleItemRect(g *ToolbarGroup) rl.Rectangle {
	for _, item := range g.items {
		if !item.hidden && item.widget != nil {
			return item.widget.Bounds()
		}
	}
	return rl.Rectangle{}
}

// itemLabel returns a human-readable label for an item (for the overflow popup).
func (tb *Toolbar) itemLabel(item *ToolbarItem) string {
	if item.menuLabel != "" {
		return item.menuLabel
	}
	if item.widget == nil {
		if item.itemType == ToolbarItemSeparator {
			return "Separator"
		}
		return item.id
	}
	if btn, ok := item.widget.(*Button); ok {
		if t := btn.Text.Get(); t != "" {
			return t
		}
	}
	if ibtn, ok := item.widget.(*IconButton); ok {
		if t := ibtn.Label.Get(); t != "" {
			return t
		}
	}
	if dd, ok := item.widget.(*Dropdown); ok {
		if dd.FaceLabel != "" {
			return dd.FaceLabel
		}
	}
	if wt, ok := item.widget.(*ToolbarWordToggle); ok {
		return wt.Text
	}
	if sb, ok := item.widget.(*ToolbarSpinBox); ok {
		return "Value " + sb.formatValue()
	}
	return item.id
}

// ─── Inspector helpers ────────────────────────────────────────────────────────

// ToolbarInfo returns a summary string for the inspector.
func (tb *Toolbar) ToolbarInfo() (groupCount, itemCount, activeGroup string) {
	total := 0
	for _, g := range tb.Groups {
		total += len(g.items)
	}
	return fmt.Sprintf("%d", len(tb.Groups)),
		fmt.Sprintf("%d", total),
		fmt.Sprintf("%d", tb.ActiveGroup.Get())
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func max32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
