// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	listTilePadX       float32 = 14
	listTileLeadingGap float32 = 12
	listTileIconGap    float32 = 16
	listTileLeadingW   float32 = 40
	listTileH         float32 = 56
	listTileDenseH    float32 = 48
	listTileToggleW   float32 = 47
	listTileToggleH   float32 = 25
)

// ListTileRowMode controls how a list row participates in pointer input.
type ListTileRowMode int

const (
	// ListTileNavigation — full-row tap and hover (chevron, value label, drill-down).
	ListTileNavigation ListTileRowMode = iota
	// ListTileSwitchOnly — locked row: only the trailing switch is interactive.
	ListTileSwitchOnly
)

// ListTile is a settings-style list row: optional leading/trailing slots with
// title and subtitle in the center column.
//
// # Row modes
//
//   - ListTileSwitchOnly (locked): body is inert — no OnClick, no hover, no inspector
//     selection on the row chrome. Only the trailing Toggle receives clicks.
//     Hovering the switch wakes ActiveFPS (CollectListTileSwitchWake) before the click.
//     Set automatically when SetTrailing receives a *Toggle, or call SetRowMode explicitly.
//   - ListTileNavigation: full-row OnClick and hover; use for chevron / value rows.
//
// Slotted children are exposed via Children() so hit-testing reaches the switch.
//
// # LLM Prompt Template
//
//	tile := ui.NewListTile("wifi", "Wi-Fi", "Home network", 0, 0, 0, 0)
//	tile.SetRowMode(ui.ListTileSwitchOnly)
//	tile.SetTrailing(ui.NewToggle("wifi-tog", true, 0, 0, 52, 28))
//	list.AddChild(tile)
//
// Demo scenes: **Batch 21 ListTile**, **Settings Demo**, **Demo Directory**.
//
// Navigation row: set `OnClick` and optional trailing chevron label; switch rows use `ListTileSwitchOnly`.
type ListTile struct {
	Element
	Title    string
	Subtitle string
	Dense    bool
	RowMode  ListTileRowMode
	Leading  Node
	Trailing Node
	OnClick  func()
	hovered  bool
	Selected bool
}

// NewListTile creates a list row. Pass w=0 and h=0 for flex width and intrinsic height.
func NewListTile(id, title, subtitle string, x, y, w, h float32) *ListTile {
	lt := &ListTile{
		Element:  NewElement(id, x, y, w, h),
		Title:    title,
		Subtitle: subtitle,
		RowMode:  ListTileNavigation,
	}
	lt.styleName = "list-tile"
	if h == 0 {
		lt.AutoHeight = true
	}
	return lt
}

// SetRowMode sets navigation vs locked switch-only interaction.
func (lt *ListTile) SetRowMode(mode ListTileRowMode) {
	lt.RowMode = mode
	if mode == ListTileSwitchOnly {
		lt.OnClick = nil
		lt.hovered = false
		if tg, ok := lt.Trailing.(*Toggle); ok {
			tg.hostedInListTile = true
		}
	} else if tg, ok := lt.Trailing.(*Toggle); ok {
		tg.hostedInListTile = false
	}
	lt.MarkDrawDirty()
}

// SwitchOnly reports whether the row body is locked (only the switch is interactive).
func (lt *ListTile) SwitchOnly() bool { return lt.RowMode == ListTileSwitchOnly }

// SetLeading assigns the optional left slot widget (Badge, IconButton, etc.).
func (lt *ListTile) SetLeading(n Node) { lt.Leading = n; lt.MarkDirty() }

// SetTrailing assigns the optional right slot widget (Toggle, Label, etc.).
// A *Toggle sets ListTileSwitchOnly automatically.
func (lt *ListTile) SetTrailing(n Node) {
	if tg, ok := lt.Trailing.(*Toggle); ok {
		tg.hostedInListTile = false
	}
	lt.Trailing = n
	if _, ok := n.(*Toggle); ok {
		lt.SetRowMode(ListTileSwitchOnly)
	} else if lt.RowMode == ListTileSwitchOnly {
		lt.SetRowMode(ListTileNavigation)
	}
	lt.MarkDirty()
}

// Children exposes slotted widgets for hit-testing, inspector, and animation walks.
func (lt *ListTile) Children() []Node {
	out := make([]Node, 0, 2)
	if lt.Leading != nil {
		out = append(out, lt.Leading)
	}
	if lt.Trailing != nil {
		out = append(out, lt.Trailing)
	}
	return out
}

// IsInteractive implements Node — the row chrome is never interactive; use OnClick
// (navigation) or the trailing Toggle (switch-only).
func (lt *ListTile) IsInteractive() bool { return false }

// Update handles row hover/click and advances slotted children.
func (lt *ListTile) Update(dt float32) {
	if lt.IsHidden() {
		return
	}
	lt.layoutSlots()

	// Trailing dropdowns/comboboxes must receive input even when a popup
	// overlay is active (including this row's own open menu).
	if lt.Leading != nil {
		UpdateNodeOverlayAware(lt.Leading, dt)
	}
	if lt.Trailing != nil {
		UpdateNodeOverlayAware(lt.Trailing, dt)
	}

	if lt.SwitchOnly() {
		lt.hovered = false
		return
	}

	mouse := rl.GetMousePosition()
	if WidgetBlockedByOverlay(mouse) {
		if lt.hovered {
			lt.hovered = false
			lt.MarkDrawDirty()
		}
		return
	}

	b := lt.Bounds()
	prev := lt.hovered
	if lt.OnClick != nil {
		lt.hovered = rl.CheckCollisionPointRec(mouse, b)
	} else {
		lt.hovered = false
	}
	if lt.hovered != prev {
		lt.MarkDrawDirty()
	}

	if lt.OnClick == nil {
		return
	}
	if listTileTrailingAbsorbsClick(lt.Trailing, PointerClickPosition()) {
		return
	}
	if PointerClickConsume(b) {
		lt.OnClick()
	}
}

// listTileTrailingAbsorbsClick is true when the latched click targets trailing
// controls (overflow menu, toggle, dropdown) so navigation OnClick does not run.
func listTileTrailingAbsorbsClick(trailing Node, click rl.Vector2) bool {
	if trailing == nil {
		return false
	}
	return nodeAbsorbsListTileClick(trailing, click)
}

func nodeAbsorbsListTileClick(n Node, click rl.Vector2) bool {
	if n == nil || n.IsHidden() {
		return false
	}
	switch w := n.(type) {
	case *IconButton:
		b := w.Bounds()
		return b.Width > 0 && b.Height > 0 && rl.CheckCollisionPointRec(click, b)
	case *Toggle:
		return rl.CheckCollisionPointRec(click, w.hitBounds())
	case *Container:
		for _, ch := range w.Children() {
			if nodeAbsorbsListTileClick(ch, click) {
				return true
			}
		}
	}
	b := n.Bounds()
	if b.Width > 0 && b.Height > 0 && n.IsInteractive() && rl.CheckCollisionPointRec(click, b) {
		return true
	}
	return false
}

// CollectListTileSwitchWake returns WakeInput when the cursor is over a
// ListTileSwitchOnly trailing toggle. The main loop merges this before pointer
// handling so idle FPS rises to ActiveFPS on hover — the first click then lands
// at full rate (global WakeOnMouseMove stays off for the rest of the UI).
func CollectListTileSwitchWake(root Node) WakeSummary {
	var out WakeSummary
	collectListTileSwitchWake(root, &out)
	return out
}

func collectListTileSwitchWake(root Node, out *WakeSummary) {
	if root == nil || root.IsHidden() {
		return
	}
	if lt, ok := root.(*ListTile); ok && lt.SwitchOnly() {
		lt.layoutSlots()
		if tg, ok := lt.Trailing.(*Toggle); ok {
			if rl.CheckCollisionPointRec(rl.GetMousePosition(), tg.hitBounds()) {
				out.Add(WakeInput, lt.ID())
			}
		}
	}
	for _, ch := range root.Children() {
		collectListTileSwitchWake(ch, out)
	}
}

// ProcessSwitchListTilePointers runs switch-only row input before the main Update
// pass so toggles get the latched click while geometry is fresh and navigation
// rows cannot consume it first.
func ProcessSwitchListTilePointers(root Node, dt float32) {
	if root == nil || root.IsHidden() {
		return
	}
	if lt, ok := root.(*ListTile); ok && lt.SwitchOnly() {
		lt.processSwitchPointer(dt)
	}
	for _, ch := range root.Children() {
		ProcessSwitchListTilePointers(ch, dt)
	}
}

func (lt *ListTile) processSwitchPointer(dt float32) {
	lt.layoutSlots()
	tg, ok := lt.Trailing.(*Toggle)
	if !ok {
		return
	}
	tg.updateHostedPointer(dt)
	PointerClickSuppressTileBody(lt.Bounds(), tg.hitBounds())
}

func listTileLeadingTextGap(lt *ListTile) float32 {
	if lt.Leading == nil {
		return 0
	}
	if _, ok := lt.Leading.(*Icon); ok {
		return listTileIconGap
	}
	return listTileLeadingGap
}

func listTileDrawChrome(b rl.Rectangle, style Style, selected, hovered bool) {
	bg := style.BackgroundColor
	if bg.A == 0 {
		bg = rl.NewColor(255, 255, 255, 255)
	}
	if selected {
		bg = listOptionSelectedBg
	}
	roundness := buttonCornerRoundness(b.Width, b.Height, style.CornerRadius)
	if style.BorderWidth > 0 || style.CornerRadius > 0 {
		borderCol := style.BorderColor
		if borderCol.A == 0 {
			borderCol = rl.NewColor(232, 234, 240, 255)
		}
		bw := style.BorderWidth
		if bw <= 0 {
			bw = 1
		}
		drawRoundedInsetBorder(b, roundness, bw, borderCol, bg)
	} else if bg.A > 0 {
		if roundness > 0 {
			rl.DrawRectangleRounded(b, roundness, 8, bg)
		} else {
			rl.DrawRectangleRec(b, bg)
		}
	}
	if hovered && !selected {
		if listTileHoverOverlay.A > 0 {
			if roundness > 0 {
				rl.DrawRectangleRounded(b, roundness, 8, listTileHoverOverlay)
			} else {
				rl.DrawRectangleRec(b, listTileHoverOverlay)
			}
		}
	}
}

func listTileLeadingSize(n Node) (w, h float32) {
	if b, ok := n.(*Badge); ok {
		w, h = b.Bounds().Width, b.Bounds().Height
		if w <= 0 {
			w = b.PreferredWidth
		}
		if h <= 0 {
			h = 24
		}
		if w <= 0 {
			w = 20
		}
		return w, h
	}
	if lbl, ok := n.(*Label); ok {
		h = float32(22)
		style := lbl.GetStyle()
		w = float32(measureTextS(lbl.Text.Get(), style))
		if w < 8 {
			w = 8
		}
		return w, h
	}
	b := n.Bounds()
	if b.Width > 0 && b.Height > 0 {
		return b.Width, b.Height
	}
	return listTileLeadingW, listTileLeadingW
}

func listTileTrailingSize(n Node, dense bool) (w, h float32) {
	if _, ok := n.(*Toggle); ok {
		return listTileToggleW, listTileToggleH
	}
	if cb, ok := n.(*ComboBox); ok {
		w := cb.PreferredWidth
		if w <= 0 {
			w = cb.Bounds().Width
		}
		if w <= 0 {
			w = 168
		}
		if mx := cb.MaxWidth; mx > 0 && w > mx {
			w = mx
		}
		h := cb.Bounds().Height
		if h <= 0 {
			h = 40
		}
		return w, h
	}
	if dd, ok := n.(*Dropdown); ok {
		w := dd.PreferredWidth
		if w <= 0 {
			w = dd.Bounds().Width
		}
		if w <= 0 {
			w = 168
		}
		if mx := dd.MaxWidth; mx > 0 && w > mx {
			w = mx
		}
		h := dd.Bounds().Height
		if h <= 0 {
			h = 32
		}
		return w, h
	}
	if ib, ok := n.(*IconButton); ok {
		w := ib.Bounds().Width
		if w <= 0 {
			w = 32
		}
		h := ib.Bounds().Height
		if h <= 0 {
			h = 32
		}
		return w, h
	}
	if btn, ok := n.(*Button); ok {
		w := btn.PreferredWidth
		if w <= 0 {
			w = btn.Bounds().Width
		}
		if w <= 0 {
			w = 72
		}
		h := btn.Bounds().Height
		if h <= 0 {
			h = 32
		}
		return w, h
	}
	if cw, ok := n.(*ColorWell); ok {
		return cw.intrinsicWidth(), colorWellDefaultH
	}
	if cp, ok := n.(*ColorPicker); ok {
		w := cp.GetPreferredWidth()
		if w <= 0 {
			w = 52
		}
		h := cp.Bounds().Height
		if h <= 0 {
			h = 36
		}
		return w, h
	}
	if c, ok := n.(*Container); ok && c.LayoutType == LayoutFlex && c.FlexDirection == FlexRow {
		var w, maxH float32
		for i, ch := range c.Children() {
			if i > 0 {
				w += c.Gap
			}
			cw, chh := listTileTrailingSize(ch, dense)
			w += cw
			if chh > maxH {
				maxH = chh
			}
		}
		if c.PreferredWidth > w {
			w = c.PreferredWidth
		}
		if w > 0 {
			if maxH <= 0 {
				maxH = 36
			}
			return w, maxH
		}
	}
	if lbl, ok := n.(*Label); ok {
		h = float32(24)
		if dense {
			h = 22
		}
		if lbl.PreferredWidth > 0 {
			return lbl.PreferredWidth, h
		}
		style := lbl.GetStyle()
		w = float32(measureTextS(lbl.Text.Get(), style)) + 8
		if w < 48 {
			w = 48
		}
		return w, h
	}
	b := n.Bounds()
	if b.Width > 0 && b.Height > 0 {
		return b.Width, b.Height
	}
	return 56, 28
}

func listTileTrailingMinWidth(n Node) float32 {
	if c, ok := n.(*Container); ok && c.PreferredWidth > 0 {
		return c.PreferredWidth
	}
	if ib, ok := n.(*IconButton); ok && ib.PreferredWidth > 0 {
		return ib.PreferredWidth
	}
	return 0
}

// contentBandHeight is the vertical space for title/subtitle and slotted icons.
func (lt *ListTile) contentBandHeight() float32 {
	base := listTileH
	if lt.Dense {
		base = listTileDenseH
	}
	if lt.Subtitle == "" {
		return base
	}
	style := lt.GetStyle()
	titleStyle := style
	if titleStyle.FontSize <= 0 {
		titleStyle.FontSize = 18
	}
	var subStyle Style
	if themeSub, ok := CurrentTheme["list-tile-subtitle"]; ok {
		subStyle = themeSub
	} else {
		subStyle = titleStyle
		subStyle.FontSize = titleStyle.FontSize - 2
		if subStyle.FontSize < 12 {
			subStyle.FontSize = 12
		}
	}
	block := float32(EffectiveFontSize(titleStyle)) + 6 + float32(EffectiveFontSize(subStyle))
	want := block + 10 // descenders (g, y) need a little room below the subtitle baseline
	if want < base {
		return base
	}
	return want
}

// Layout positions leading/trailing slots and optional intrinsic height.
func (lt *ListTile) Layout() {
	b := lt.Bounds()
	wantH := lt.contentBandHeight()
	if lt.IsAutoHeight() && b.Height < wantH-0.5 {
		b.Height = wantH
		lt.setBoundsNoMark(b)
	}
	lt.layoutSlots()
	lt.layoutDirty = false
}

// layoutSlots assigns leading/trailing bounds in screen space (also called from Update).
func (lt *ListTile) layoutSlots() {
	b := lt.Bounds()
	if b.Width < 1 || b.Height < 1 {
		return
	}
	wantH := lt.contentBandHeight()
	innerY := b.Y + (b.Height-wantH)/2
	if innerY < b.Y {
		innerY = b.Y
	}
	slotH := wantH
	if slotH > b.Height {
		slotH = b.Height
	}

	if lt.Leading != nil {
		lt.Leading.SetParent(lt)
		lw, lh := listTileLeadingSize(lt.Leading)
		lx := b.X + listTilePadX
		ly := innerY + (slotH-lh)/2
		setSlotBounds(lt.Leading, rl.NewRectangle(lx, ly, lw, lh))
		lt.Leading.Layout()
	}
	if lt.Trailing != nil {
		lt.Trailing.SetParent(lt) // so MarkDrawDirty reaches doc.Root (SSAA cache invalidation)
		tw, th := listTileTrailingSize(lt.Trailing, lt.Dense)
		leadingReserve := listTilePadX
		if lt.Leading != nil {
			lw, _ := listTileLeadingSize(lt.Leading)
			leadingReserve += lw + listTileLeadingGap + listTileIconGap
		}
		minTitleW := float32(72)
		trMin := listTileTrailingMinWidth(lt.Trailing)
		if trMin > 0 {
			if room := b.Width - leadingReserve - listTilePadX - trMin; room < minTitleW {
				minTitleW = room
				if minTitleW < 24 {
					minTitleW = 24
				}
			}
		}
		if maxTw := b.Width - leadingReserve - listTilePadX - minTitleW; maxTw > 48 && tw > maxTw {
			tw = maxTw
		}
		if trMin > 0 && tw < trMin {
			tw = trMin
		}
		if dd, ok := lt.Trailing.(*Dropdown); ok && tw > 0 {
			dd.MaxWidth = tw
		}
		tx := b.X + b.Width - listTilePadX - tw
		ty := innerY + (slotH-th)/2
		setSlotBounds(lt.Trailing, rl.NewRectangle(tx, ty, tw, th))
		lt.Trailing.Layout()
	}
}

func setSlotBounds(n Node, r rl.Rectangle) {
	if sameRect(n.Bounds(), r) {
		return
	}
	if el, ok := n.(interface{ setBoundsNoMark(rl.Rectangle) }); ok {
		el.setBoundsNoMark(r)
	} else {
		n.SetBounds(r)
	}
}

func sameRect(a, b rl.Rectangle) bool {
	return a.X == b.X && a.Y == b.Y && a.Width == b.Width && a.Height == b.Height
}

// Draw renders the row chrome, text, and slotted children.
func (lt *ListTile) Draw() {
	defer func() { lt.drawDirty = false }()
	lt.drawInternal()
}

func (lt *ListTile) drawInternal() {
	if lt.IsHidden() {
		return
	}
	b := lt.Bounds()
	style := lt.GetStyle()

	if style.BackgroundColor.A > 0 || style.BorderWidth > 0 {
		listTileDrawChrome(b, style, lt.Selected, lt.hovered)
	} else {
		if lt.Selected {
			rl.DrawRectangleRec(b, listOptionSelectedBg)
		}
		if lt.hovered {
			if listTileHoverOverlay.A > 0 {
				rl.DrawRectangleRec(b, listTileHoverOverlay)
			}
		}
	}

	textX := b.X + listTilePadX
	if lt.Leading != nil {
		textX = lt.Leading.Bounds().X + lt.Leading.Bounds().Width + listTileLeadingTextGap(lt)
	}
	textRight := b.X + b.Width - listTilePadX
	if lt.Trailing != nil {
		if tg, ok := lt.Trailing.(*Toggle); ok {
			textRight = tg.PillBounds().X - 8
		} else {
			textRight = lt.Trailing.Bounds().X - 8
		}
	}
	maxTextW := textRight - textX
	if maxTextW < 20 {
		maxTextW = 20
	}

	titleStyle := style
	if titleStyle.FontSize <= 0 {
		titleStyle.FontSize = 18
	}

	var subStyle Style
	if themeSub, ok := CurrentTheme["list-tile-subtitle"]; ok {
		subStyle = themeSub
	} else {
		subStyle = titleStyle
		subStyle.FontSize = titleStyle.FontSize - 2
		if subStyle.FontSize < 12 {
			subStyle.FontSize = 12
		}
		subStyle.Bold = false
		subStyle.TextColor = rl.NewColor(100, 106, 128, 255)
	}

	titleFS := EffectiveFontSize(titleStyle)
	subFS := EffectiveFontSize(subStyle)

	if lt.Subtitle == "" {
		titleY := int32(b.Y) + (int32(b.Height)-int32(titleFS))/2
		drawTextS(truncateTextS(lt.Title, maxTextW, titleStyle), int32(textX), titleY, titleStyle)
	} else {
		blockH := titleFS + 6 + subFS
		startY := b.Y + (b.Height-float32(blockH))/2
		drawTextS(truncateTextS(lt.Title, maxTextW, titleStyle), int32(textX), int32(startY), titleStyle)
		drawTextS(truncateTextS(lt.Subtitle, maxTextW, subStyle), int32(textX), int32(startY+float32(titleFS)+6), subStyle)
	}

	if lt.Leading != nil {
		lt.Leading.Draw()
	}
	if lt.Trailing != nil {
		lt.Trailing.Draw()
	}
}

// InspectorPickTarget implements inspector pick rules for locked switch rows.
// When the row is switch-only, the row chrome is not selectable — only slotted nodes are.
func (lt *ListTile) InspectorPickTarget(childHit Node) Node {
	if lt.SwitchOnly() {
		return childHit
	}
	return lt
}
