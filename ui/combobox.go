// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const comboFilterH float32 = 36

// ComboBox is a searchable dropdown: type to filter options, then pick one.
// The closed face shows [ComboBox.Selected]; the open popup draws a filter row
// and matching options (same overlay pattern as [Dropdown]).
//
// # LLM Prompt Template
//
//	cb := ui.NewComboBox("country", countries, selected, 0, 0, 0, 40)
//	// selected is *Signal[string]; filter while open narrows Options.
type ComboBox struct {
	Element
	Options   []string
	Selected  *Signal[string]
	Disabled  bool
	Placeholder string

	isOpen       bool
	hovered      bool
	hoveredIndex int
	filter       string
	filterFocus  bool
	popupScrollY float32
	popupOpenAbove bool
	popupListCap   float32
}

// NewComboBox creates a ComboBox. selected must not be nil; if its value is not
// in options, the first option is used when the list is non-empty.
func NewComboBox(id string, options []string, selected *Signal[string], x, y, w, h float32) *ComboBox {
	if selected == nil {
		panic("ui.NewComboBox: selected must not be nil")
	}
	cur := selected.Get()
	if cur == "" && len(options) > 0 {
		cur = options[0]
		selected.Set(cur)
	}
	c := &ComboBox{
		Element:     NewElement(id, x, y, w, h),
		Options:     options,
		Selected:    selected,
		Placeholder: "Select…",
		hoveredIndex: -1,
	}
	c.Selected.Subscribe(func() { c.MarkDirty() })
	c.ZIndex = 5
	c.styleName = "combobox"
	c.Element.SetStyleVariant("combobox", "default")
	return c
}

func (c *ComboBox) optionH() float32 {
	h := c.Bounds().Height
	if h < 32 {
		h = 32
	}
	if h > 40 {
		h = 40
	}
	return h
}

func (c *ComboBox) filteredOptions() []string {
	if c.filter == "" {
		return c.Options
	}
	q := strings.ToLower(strings.TrimSpace(c.filter))
	if q == "" {
		return c.Options
	}
	out := make([]string, 0, len(c.Options))
	for _, o := range c.Options {
		if strings.Contains(strings.ToLower(o), q) {
			out = append(out, o)
		}
	}
	return out
}

func (c *ComboBox) filterRect() rl.Rectangle {
	b := c.Bounds()
	if c.popupOpenAbove {
		return rl.NewRectangle(b.X, b.Y-comboFilterH, b.Width, comboFilterH)
	}
	return rl.NewRectangle(b.X, b.Y+b.Height, b.Width, comboFilterH)
}

func (c *ComboBox) listTopY() float32 {
	b := c.Bounds()
	if c.popupOpenAbove {
		return b.Y - comboFilterH - c.listViewportH()
	}
	return b.Y + b.Height + comboFilterH
}

func (c *ComboBox) syncMenuPopupPlacement() {
	if !c.isOpen {
		c.popupOpenAbove = false
		c.popupListCap = 0
		return
	}
	p := computeMenuPopupPlacement(c, c.Bounds(), comboFilterH, c.listContentHeight(), c.optionH())
	c.popupOpenAbove = p.openAbove
	c.popupListCap = p.listCap
}

func (c *ComboBox) getOptionBounds() []rl.Rectangle {
	b := c.Bounds()
	opts := c.filteredOptions()
	n := len(opts)
	if n == 0 {
		n = 1
	}
	oh := c.optionH()
	listTop := c.listTopY()
	popW := c.popupWidth()
	out := make([]rl.Rectangle, n)
	for i := 0; i < n; i++ {
		out[i] = rl.NewRectangle(
			b.X,
			listTop+float32(i)*oh-c.popupScrollY,
			popW,
			oh,
		)
	}
	return out
}

func (c *ComboBox) popupWidth() float32 {
	return menuPopupListWidth(c, c.Bounds().Width, c.filteredOptions(), c.GetStyle())
}

func (c *ComboBox) listContentHeight() float32 {
	opts := c.filteredOptions()
	n := len(opts)
	if n == 0 {
		n = 1
	}
	return c.optionH() * float32(n)
}

func (c *ComboBox) listViewportH() float32 {
	h := c.listContentHeight()
	max := menuPopupMaxHeight
	if c.isOpen && c.popupListCap > 0 {
		max = c.popupListCap
	}
	if h > max {
		return max
	}
	return h
}

func (c *ComboBox) PopupBounds() rl.Rectangle {
	if !c.isOpen {
		return rl.Rectangle{}
	}
	b := c.Bounds()
	h := comboFilterH + c.listViewportH()
	y := b.Y + b.Height
	if c.popupOpenAbove {
		y = b.Y - h
	}
	return rl.NewRectangle(b.X, y, c.popupWidth(), h)
}

func (c *ComboBox) popupRect() rl.Rectangle {
	b := c.PopupBounds()
	if b.Width == 0 {
		return rl.Rectangle{}
	}
	return b
}

func (c *ComboBox) absorbFilterKeys() {
	for {
		ch := rl.GetCharPressed()
		if ch == 0 {
			break
		}
		if ch >= 32 && ch != 127 {
			c.filter += string(rune(ch))
			c.MarkDrawDirty()
		}
	}
	if rl.IsKeyPressed(rl.KeyBackspace) && len(c.filter) > 0 {
		c.filter = c.filter[:len(c.filter)-1]
		c.MarkDrawDirty()
	}
}

// Update implements Node.Update.
func (c *ComboBox) Update(_ float32) {
	if c.IsHidden() {
		return
	}
	if c.Disabled {
		if c.isOpen || c.hovered || c.hoveredIndex != -1 {
			c.isOpen = false
			c.hovered = false
			c.hoveredIndex = -1
			c.filter = ""
			c.popupScrollY = 0
			c.MarkDrawDirty()
		}
		return
	}

	mouse := rl.GetMousePosition()
	if WidgetBlockedByOverlay(mouse, c.OverlayExemptRects()...) {
		c.ClearOverlayPointerState()
		return
	}

	bounds := c.Bounds()
	prevOpen := c.isOpen
	prevHov := c.hoveredIndex
	prevHovered := c.hovered

	c.hovered = rl.CheckCollisionPointRec(mouse, bounds)

	if c.isOpen {
		c.absorbFilterKeys()
		prevCap := c.popupListCap
		c.syncMenuPopupPlacement()
		if c.popupListCap != prevCap {
			c.popupScrollY = clampMenuPopupScroll(c.popupScrollY, c.listContentHeight(), c.listViewportH())
		}
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		switch {
		case c.hovered:
			if c.isOpen {
				c.isOpen = false
				c.filter = ""
				c.filterFocus = false
			} else {
				c.isOpen = true
				c.filter = ""
				c.filterFocus = true
				c.hoveredIndex = -1
				c.popupScrollY = 0
				c.syncMenuPopupPlacement()
				ensureMenuPopupVisible(c, c.PopupBounds)
			}
			PointerClickMarkUsed()
		case c.isOpen:
			fr := c.filterRect()
			if rl.CheckCollisionPointRec(mouse, fr) {
				c.filterFocus = true
			} else {
				picked := false
				for i, ob := range c.getOptionBounds() {
					if rl.CheckCollisionPointRec(mouse, ob) {
						opts := c.filteredOptions()
						if i >= 0 && i < len(opts) {
							c.Selected.Set(opts[i])
						}
						picked = true
						PointerClickMarkUsed()
						break
					}
				}
				if picked || !rl.CheckCollisionPointRec(mouse, c.PopupBounds()) {
					c.isOpen = false
					c.filter = ""
					c.filterFocus = false
					c.popupScrollY = 0
					if picked {
						PointerClickMarkUsed()
					}
				}
			}
		}
	}

	if c.isOpen {
		c.hoveredIndex = -1
		for i, ob := range c.getOptionBounds() {
			if rl.CheckCollisionPointRec(mouse, ob) {
				c.hoveredIndex = i
				break
			}
		}
		if rl.CheckCollisionPointRec(mouse, c.PopupBounds()) {
			wheel := rl.GetMouseWheelMove()
			if wheel != 0 && !rl.CheckCollisionPointRec(mouse, c.filterRect()) {
				c.popupScrollY -= wheel * c.optionH() * 2
				c.popupScrollY = clampMenuPopupScroll(c.popupScrollY, c.listContentHeight(), c.listViewportH())
				c.MarkDrawDirty()
			}
		}
	}

	if c.isOpen != prevOpen || (c.isOpen && c.hoveredIndex != prevHov) {
		c.MarkDrawDirty()
	} else if !c.isOpen && c.hovered != prevHovered {
		c.MarkDrawDirty()
	}
}

// OverlayExemptRects implements overlayExempter.
func (c *ComboBox) OverlayExemptRects() []rl.Rectangle {
	if !c.isOpen {
		return nil
	}
	return []rl.Rectangle{c.Bounds(), c.PopupBounds()}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (c *ComboBox) ClearOverlayPointerState() {
	if !c.isOpen && !c.hovered && c.hoveredIndex == -1 {
		return
	}
	c.isOpen = false
	c.hovered = false
	c.hoveredIndex = -1
	c.filter = ""
	c.popupScrollY = 0
	c.popupOpenAbove = false
	c.popupListCap = 0
	c.MarkDrawDirty()
}

// IsOpen reports whether the popup is expanded (for Viewport overlay drawing).
func (c *ComboBox) IsOpen() bool { return c.isOpen }

// DrawPopup renders the filter row and option list outside the viewport scissor.
func (c *ComboBox) DrawPopup() {
	if !c.isOpen {
		return
	}
	drawMenuPopupInHost(c, func() {
		c.drawPopupUnclipped()
	})
}

func (c *ComboBox) drawPopupUnclipped() {
	style := c.GetStyle()
	pop := c.popupRect()
	panelBg, panelBorder, divider := menuPopupChromeColors(style)
	filterBg, filterFocusBg := menuPopupFilterColors(style)

	shadowRect := rl.NewRectangle(pop.X+3, pop.Y+3, pop.Width, pop.Height)
	rl.DrawRectangleRec(shadowRect, rl.NewColor(0, 0, 0, 22))

	panelRoundness := float32(0.08)
	rl.DrawRectangleRounded(pop, panelRoundness, 8, panelBg)
	rl.DrawRectangleRoundedLinesEx(pop, panelRoundness, 8, 1, panelBorder)

	fr := c.filterRect()
	if c.filterFocus {
		filterBg = filterFocusBg
	}
	rl.DrawRectangleRec(fr, filterBg)
	rl.DrawLine(int32(fr.X)+8, int32(fr.Y+fr.Height-1), int32(fr.X+fr.Width)-8, int32(fr.Y+fr.Height-1), divider)

	pad := int32(style.Padding + 0.5)
	if pad < 8 {
		pad = 8
	}
	hint := c.filter
	if hint == "" {
		hint = "Search…"
		filterStyle := style
		filterStyle.TextColor = rl.NewColor(140, 142, 155, 255)
		drawTextS(hint, int32(fr.X)+pad, TextPosY(fr, style), filterStyle)
	} else {
		drawTextS(hint, int32(fr.X)+pad, TextPosY(fr, style), style)
	}

	opts := c.filteredOptions()
	selected := c.Selected.Get()
	listArea := rl.NewRectangle(pop.X, pop.Y, pop.Width, c.listViewportH())
	if !c.popupOpenAbove {
		listArea.Y = pop.Y + comboFilterH
	}
	beginScissorFromRect(listArea)
	listTop := listArea.Y
	oh := c.optionH()
	if len(opts) == 0 {
		ob := rl.NewRectangle(pop.X, listTop, pop.Width, oh)
		optStyle := style
		optStyle.TextColor = rl.NewColor(140, 142, 155, 255)
		drawTextS("(no matches)", int32(ob.X)+pad, int32(ob.Y)+(int32(ob.Height)-int32(optStyle.FontSize))/2, optStyle)
	} else {
		for i := range opts {
			ob := rl.NewRectangle(pop.X, listTop+float32(i)*oh-c.popupScrollY, pop.Width, oh)
			if ob.Y+ob.Height <= listArea.Y || ob.Y >= listArea.Y+listArea.Height {
				continue
			}
			opt := opts[i]
			var optBg rl.Color
			var textCol rl.Color
			switch {
			case opt == selected:
				optBg = rl.NewColor(0, 0, 0, 0)
				textCol = DropdownSelectedTextColor()
			case i == c.hoveredIndex:
				optBg = listRowHoverBg
				textCol = style.TextColor
			default:
				optBg = rl.NewColor(0, 0, 0, 0)
				textCol = style.TextColor
			}
			if optBg.A > 0 {
				rl.DrawRectangleRec(ob, optBg)
			}
			if i < len(opts)-1 {
				lineY := int32(ob.Y + ob.Height - 1)
				rl.DrawLine(int32(ob.X)+pad, lineY, int32(ob.X+ob.Width)-pad, lineY, divider)
			}
			optStyle := style
			optStyle.TextColor = textCol
			if opt == selected {
				optStyle.Bold = true
			}
			drawTextS(opt, int32(ob.X)+pad, int32(ob.Y)+(int32(ob.Height)-int32(optStyle.FontSize))/2, optStyle)
		}
	}
	rl.EndScissorMode()
}

func (c *ComboBox) InteractionOverlayActive() bool { return false }

func (c *ComboBox) DrawInteractionOverlay() {}

func (c *ComboBox) Layout() { c.layoutDirty = false }

func (c *ComboBox) Draw() { c.drawInternal() }

func (c *ComboBox) drawInternal() {
	if c.IsHidden() {
		return
	}
	state := StyleStateNone
	if c.hovered {
		state |= StyleStateHover
	}
	if c.isOpen {
		state |= StyleStateOpen
	}
	if c.Disabled {
		state |= StyleStateDisabled
	}
	style, stateApplied := c.ResolveStyle(state)
	bounds := c.Bounds()

	bg := style.BackgroundColor
	if !stateApplied && (c.isOpen || c.hovered) {
		bg = rl.ColorBrightness(bg, 0.045)
	}

	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	borderCol := style.BorderColor
	if borderCol.A == 0 {
		borderCol = rl.NewColor(218, 222, 232, 255)
	}
	if !stateApplied && c.isOpen {
		borderCol = lerpColor(borderCol, focusRingIndigo, 0.28)
	} else if !stateApplied && c.hovered {
		borderCol = rl.NewColor(210, 214, 224, 255)
	}

	roundness := buttonCornerRoundness(bounds.Width, bounds.Height, style.CornerRadius)
	drawRoundedInsetBorder(bounds, roundness, bw, borderCol, bg)

	label := c.Selected.Get()
	if label == "" {
		label = c.Placeholder
	}
	textStyle := style
	arrowReserve := float32(34)
	textMaxW := bounds.Width - style.Padding - arrowReserve
	if textMaxW < 24 {
		textMaxW = 24
	}
	if label == "" || label == c.Placeholder {
		textStyle.TextColor = rl.NewColor(140, 142, 155, 255)
	}
	drawTextS(truncateTextS(label, textMaxW, textStyle), int32(bounds.X+style.Padding+0.5), TextPosY(bounds, style), textStyle)

	c.drawDropArrow(bounds)
}

func (c *ComboBox) drawDropArrow(bounds rl.Rectangle) {
	icon := PhosphorArrowDropDown
	if c.isOpen {
		icon = PhosphorArrowDropUp
	}
	Phosphor.EnsureLoaded(icon, PhosphorRegular)
	sz := float32(20)
	dst := rl.NewRectangle(
		bounds.X+bounds.Width-sz-10,
		bounds.Y+(bounds.Height-sz)/2,
		sz, sz,
	)
	col := rl.NewColor(110, 112, 128, 255)
	if !Phosphor.Draw(dst, icon, PhosphorRegular, col) {
		// Unicode fallback
		arrowR := float32(7)
		arrowX := dst.X + 2
		arrowY := dst.Y + dst.Height/2
		if c.isOpen {
			rl.DrawTriangle(
				rl.NewVector2(arrowX, arrowY+arrowR*0.5),
				rl.NewVector2(arrowX+arrowR*2, arrowY+arrowR*0.5),
				rl.NewVector2(arrowX+arrowR, arrowY-arrowR*0.5),
				col,
			)
		} else {
			rl.DrawTriangle(
				rl.NewVector2(arrowX, arrowY-arrowR*0.5),
				rl.NewVector2(arrowX+arrowR*2, arrowY-arrowR*0.5),
				rl.NewVector2(arrowX+arrowR, arrowY+arrowR*0.5),
				col,
			)
		}
	}
}

func (c *ComboBox) IsInteractive() bool { return !c.Disabled }
