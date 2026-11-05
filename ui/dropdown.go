// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// listRowHoverBg matches DataTable row hover fill (updated by SetAppearance).
var listRowHoverBg = rl.NewColor(237, 239, 254, 255)

// Dropdown is an interactive widget that shows a selected value and expands
// a menu of options when clicked.
//
// SelectedIndex is a Signal[int] — subscribe or read via ControlSnapshot when
// compiled from DocumentSpec (`type`: `dropdown`).
//
// # LLM Prompt Template
//
//	dd := ui.NewDropdown("density", []string{"Compact", "Comfortable"}, 0, 0, 0, 260, 40)
//	form.AddChild(dd)
//
// Demo scenes: **Batch 19 Dropdown**, **Filters (Go)**, **Form Demo**, `pages/settings.gru`.
//
// # Overlay Drawing
//
// When isOpen is true the option list is drawn below the widget bounds.  To
// prevent the option list from being clipped by a parent Viewport's scissor,
// Viewport.Draw() separates children into "regular" and "overlay" groups:
// regular children are drawn inside the scissor; open Dropdowns are drawn
// after EndScissorMode() so their options are always fully visible.
//
// # Z-Index
//
// Dropdown's ZIndex defaults to 5 so the closed widget draws above siblings
// with ZIndex 0 when their bounds overlap.
type Dropdown struct {
	Element
	Options       []string     // List of available options
	SelectedIndex *Signal[int] // Reactive selected index
	// FaceLabel, when set, is always shown on the closed face (toolbar menus).
	FaceLabel      string
	Disabled       bool
	isOpen         bool // Whether dropdown is expanded
	hovered        bool // Internal hover state
	hoveredIndex   int  // Which option is hovered (-1 = none)
	popupScrollY   float32
	popupOpenAbove bool
	popupListCap   float32
}

// NewDropdown creates a new Dropdown with the given options.
// Default style is "dropdown" (white bg, border, rounded corners).
func NewDropdown(id string, options []string, selectedIndex int, x, y, w, h float32) *Dropdown {
	if selectedIndex < 0 || selectedIndex >= len(options) {
		selectedIndex = 0
	}
	d := &Dropdown{
		Element:       NewElement(id, x, y, w, h),
		Options:       options,
		SelectedIndex: NewSignal(selectedIndex),
		isOpen:        false,
		hovered:       false,
		hoveredIndex:  -1,
	}
	d.SelectedIndex.Subscribe(func() { d.MarkDirty() })
	d.ZIndex = 5
	d.styleName = "dropdown"
	d.Element.SetStyleVariant("dropdown", "default")
	return d
}

// optionH returns the pixel height of each row in the expanded popup.
// Rows are at least 32 px for parity with compact fields (e.g. SearchBar) and
// capped at 40 px so very tall dropdown faces do not create oversized rows.
func (d *Dropdown) optionH() float32 {
	h := d.Bounds().Height
	if h < 32 {
		h = 32
	}
	if h > 40 {
		h = 40
	}
	return h
}

// OpenFromToolbarOverflow positions the face on rowRect and opens the option list
// (overflow menu mode when the dropdown was hidden from the main bar).
func (d *Dropdown) OpenFromToolbarOverflow(rowRect rl.Rectangle) {
	if d.Disabled || len(d.Options) == 0 {
		return
	}
	d.setBoundsNoMark(rowRect)
	d.isOpen = true
	d.hoveredIndex = -1
	d.syncMenuPopupPlacement()
	d.scrollToSelected()
	ensureMenuPopupVisible(d, d.PopupBounds)
	d.MarkDrawDirty()
}

// Update implements Node.Update by handling mouse input.
func (d *Dropdown) Update(_ float32) {
	if d.IsHidden() {
		return
	}
	if d.Disabled {
		if d.isOpen || d.hovered || d.hoveredIndex != -1 {
			d.isOpen = false
			d.hovered = false
			d.hoveredIndex = -1
			d.popupScrollY = 0
			d.MarkDrawDirty()
		}
		return
	}
	mouse := rl.GetMousePosition()
	if WidgetBlockedByOverlay(mouse, d.OverlayExemptRects()...) {
		d.ClearOverlayPointerState()
		return
	}
	bounds := d.Bounds()
	prevIsOpen := d.isOpen
	prevHovIdx := d.hoveredIndex
	prevHovered := d.hovered
	d.hovered = rl.CheckCollisionPointRec(mouse, bounds)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if d.hovered {
			d.isOpen = !d.isOpen
			d.hoveredIndex = -1
			if d.isOpen {
				d.syncMenuPopupPlacement()
				d.scrollToSelected()
				ensureMenuPopupVisible(d, d.PopupBounds)
			} else {
				d.popupScrollY = 0
			}
		} else if d.isOpen {
			for i, ob := range d.getOptionBounds() {
				if rl.CheckCollisionPointRec(mouse, ob) {
					d.SelectedIndex.Set(i)
					PointerClickMarkUsed()
					break
				}
			}
			d.isOpen = false
			d.popupScrollY = 0
		}
	}

	if d.isOpen {
		d.hoveredIndex = -1
		for i, ob := range d.getOptionBounds() {
			if rl.CheckCollisionPointRec(mouse, ob) {
				d.hoveredIndex = i
				break
			}
		}
		if rl.CheckCollisionPointRec(mouse, d.PopupBounds()) {
			wheel := rl.GetMouseWheelMove()
			if wheel != 0 {
				d.popupScrollY -= wheel * d.optionH() * 2
				d.popupScrollY = clampMenuPopupScroll(d.popupScrollY, d.listContentHeight(), d.listViewportH())
				d.MarkDrawDirty()
			}
		}
	}

	if d.isOpen != prevIsOpen {
		d.MarkDrawDirty()
	} else if d.isOpen && d.hoveredIndex != prevHovIdx {
		d.MarkDrawDirty()
	} else if !d.isOpen && d.hovered != prevHovered {
		d.MarkDrawDirty()
	}
}

// OverlayExemptRects implements overlayExempter.
func (d *Dropdown) OverlayExemptRects() []rl.Rectangle {
	if !d.isOpen {
		return nil
	}
	return []rl.Rectangle{d.Bounds(), d.PopupBounds()}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (d *Dropdown) ClearOverlayPointerState() {
	if !d.isOpen && !d.hovered && d.hoveredIndex == -1 {
		return
	}
	d.isOpen = false
	d.hovered = false
	d.hoveredIndex = -1
	d.popupScrollY = 0
	d.popupOpenAbove = false
	d.popupListCap = 0
	d.MarkDrawDirty()
}

// IsOpen returns true when the dropdown popup is currently expanded.
// Used by Viewport to collect open dropdowns for overlay drawing.
func (d *Dropdown) IsOpen() bool { return d.isOpen }

// InteractionOverlayActive implements InteractionOverlayPainter.
func (d *Dropdown) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (d *Dropdown) DrawInteractionOverlay() {}

// getOptionBounds returns the screen rectangles for each option row.
func (d *Dropdown) getOptionBounds() []rl.Rectangle {
	pop := d.PopupBounds()
	popW := pop.Width
	if popW <= 0 {
		popW = d.popupWidth()
	}
	oh := d.optionH()
	listTop := pop.Y
	if listTop <= 0 {
		listTop = d.listTopY()
	}
	result := make([]rl.Rectangle, len(d.Options))
	for i := range d.Options {
		result[i] = rl.NewRectangle(
			pop.X,
			listTop+float32(i)*oh-d.popupScrollY,
			popW,
			oh,
		)
	}
	return result
}

func (d *Dropdown) popupWidth() float32 {
	return menuPopupListWidth(d, d.Bounds().Width, d.Options, d.GetStyle())
}

func (d *Dropdown) syncMenuPopupPlacement() {
	if !d.isOpen {
		d.popupOpenAbove = false
		d.popupListCap = 0
		return
	}
	p := computeMenuPopupPlacement(d, d.Bounds(), 0, d.listContentHeight(), d.optionH())
	d.popupOpenAbove = p.openAbove
	d.popupListCap = p.listCap
}

func (d *Dropdown) listTopY() float32 {
	b := d.Bounds()
	if d.popupOpenAbove {
		return b.Y - d.listViewportH()
	}
	return b.Y + b.Height
}

func (d *Dropdown) listContentHeight() float32 {
	return d.optionH() * float32(len(d.Options))
}

func (d *Dropdown) listViewportH() float32 {
	h := d.listContentHeight()
	max := menuPopupMaxHeight
	if d.isOpen && d.popupListCap > 0 {
		max = d.popupListCap
	}
	if h > max {
		return max
	}
	return h
}

// PopupBounds returns the on-screen list viewport while open.
func (d *Dropdown) PopupBounds() rl.Rectangle {
	if !d.isOpen {
		return rl.Rectangle{}
	}
	b := d.Bounds()
	viewH := d.listViewportH()
	popW := d.popupWidth()
	y := b.Y + b.Height
	if d.popupOpenAbove {
		y = b.Y - viewH
	}
	return clampMenuPopupRect(d, rl.NewRectangle(b.X, y, popW, viewH))
}

func (d *Dropdown) scrollToSelected() {
	idx := d.SelectedIndex.Get()
	if idx < 0 {
		idx = 0
	}
	if idx >= len(d.Options) {
		idx = len(d.Options) - 1
	}
	oh := d.optionH()
	d.popupScrollY = float32(idx)*oh - d.listViewportH()*0.35
	d.popupScrollY = clampMenuPopupScroll(d.popupScrollY, d.listContentHeight(), d.listViewportH())
}

// Layout implements Node.Layout (no-op for leaf widgets).
func (d *Dropdown) Layout() { d.layoutDirty = false }

// Draw implements Node.Draw by rendering the dropdown button (not the popup).
func (d *Dropdown) Draw() { d.drawInternal() }

func (d *Dropdown) drawInternal() {
	if d.IsHidden() {
		return
	}
	state := StyleStateNone
	if d.hovered {
		state |= StyleStateHover
	}
	if d.isOpen {
		state |= StyleStateOpen
	}
	if d.Disabled {
		state |= StyleStateDisabled
	}
	style, _ := d.ResolveStyle(state)
	// Face type size tracks list-tile titles; TextColor stays on the widget style
	// (toolbar-menu*) so SetAppearance light/dark patches apply every frame.
	faceStyle := style
	if tile := GetThemeStyle("list-tile"); tile.FontSize > 0 {
		faceStyle.FontSize = tile.FontSize
		faceStyle.FontDensity = tile.FontDensity
		faceStyle.MinFontSize = tile.MinFontSize
		faceStyle.MaxFontSize = tile.MaxFontSize
		faceStyle.NoCaptionBold = tile.NoCaptionBold
	}
	// Match TextInput face layout: chrome fills the slot; text at Padding from the
	// left. Avoid fieldPaintBounds + content + extra pad (that clipped "github"/"Light").
	layoutBounds := snapControlRect(d.Bounds())
	chromeName := d.StyleName()
	if chromeName == "" || chromeName == "dropdown" {
		chromeName = "toolbar-menu"
	}
	drawToolbarChrome(layoutBounds, chromeName, style, d.hovered, d.isOpen, false)

	pad := float32(style.Padding)
	if pad < 8 {
		pad = 8
	}
	const caretReserve = float32(34)
	faceText := d.FaceLabel
	if faceText == "" {
		if idx := d.SelectedIndex.Get(); idx >= 0 && idx < len(d.Options) {
			faceText = d.Options[idx]
		}
	}
	textW := layoutBounds.Width - pad - caretReserve
	if textW < 8 {
		textW = 8
	}
	textRect := rl.NewRectangle(layoutBounds.X+pad, layoutBounds.Y, textW, layoutBounds.Height)
	posX := int32(textRect.X + 0.5)
	posY := toolbarTextPosY(textRect, faceStyle)
	bw := style.BorderWidth
	if bw < 1 {
		bw = 1
	}
	clip := rl.NewRectangle(
		layoutBounds.X+bw, layoutBounds.Y+bw,
		layoutBounds.Width-2*bw, layoutBounds.Height-2*bw,
	)
	if hasActiveDrawClip {
		clip = intersectRects(clip, activeDrawClip)
	}
	if clip.Width >= 1 && clip.Height >= 1 {
		beginScissorFromRect(clip)
		drawTextS(truncateTextS(faceText, textRect.Width, faceStyle), posX, posY, faceStyle)
		endNestedScissor()
	}
	caretTint := faceStyle.TextColor
	if caretTint.A == 0 {
		caretTint = rl.NewColor(100, 104, 120, 255)
	} else if caretTint.A > 160 {
		caretTint.A = 160
	}
	content := toolbarContentRect(layoutBounds, style)
	drawToolbarMenuCaret(content, d.isOpen, faceStyle, caretTint)
}

// drawToolbarMenuCaret draws a caret beside toolbar menu label text (bottom-aligned with face text).
func drawToolbarMenuCaret(content rl.Rectangle, open bool, textStyle Style, tint rl.Color) {
	const caretSlot = float32(18)
	dstY := toolbarAccessoryY(content, textStyle, caretSlot)
	dst := rl.NewRectangle(
		content.X+content.Width-caretSlot-8,
		dstY,
		caretSlot, caretSlot,
	)
	name := PhosphorCaretDown
	if open {
		name = PhosphorCaretUp
	}
	if remixDrawIcon(dst, name, PhosphorRegular, tint, 1) {
		return
	}
	fs := float32(10)
	glyph := "▾"
	if open {
		glyph = "▴"
	}
	tw := measureTextF(glyph, fs, false, false, false, false)
	drawTextF(glyph, dst.X+(dst.Width-tw)/2, dst.Y+(dst.Height-fs)/2, fs, tint, false, false, false, false)
}

// DrawPopup renders the expanded option list.
// It is called by the enclosing Viewport AFTER EndScissorMode so the popup
// is never clipped by the viewport content area and always appears on top.
func (d *Dropdown) DrawPopup() {
	if !d.isOpen || len(d.Options) == 0 {
		return
	}
	drawMenuPopupInHost(d, func() {
		d.drawPopupUnclipped()
	})
}

func (d *Dropdown) drawPopupUnclipped() {
	style := d.GetStyle()
	// Same type size as the closed face / list-tile titles.
	itemStyle := style
	if tile := GetThemeStyle("list-tile"); tile.FontSize > 0 {
		itemStyle.FontSize = tile.FontSize
		itemStyle.FontDensity = tile.FontDensity
		itemStyle.MinFontSize = tile.MinFontSize
		itemStyle.MaxFontSize = tile.MaxFontSize
		itemStyle.NoCaptionBold = tile.NoCaptionBold
	}
	oh := d.optionH()
	popupRect := d.PopupBounds()
	panelBg, panelBorder, divider := menuPopupChromeColors(style)

	shadowRect := rl.NewRectangle(
		popupRect.X+3, popupRect.Y+3,
		popupRect.Width, popupRect.Height,
	)
	rl.DrawRectangleRec(shadowRect, rl.NewColor(0, 0, 0, 22))

	panelRoundness := float32(0.08)
	DrawRoundedTessInsetBorder(popupRect, panelRoundness, 1, panelBorder, panelBg)

	beginScissorFromRect(popupRect)
	selectedIdx := d.SelectedIndex.Get()
	pad := int32(style.Padding + 0.5)
	if pad < 2 {
		pad = 2
	}
	listTop := popupRect.Y
	popW := popupRect.Width
	textPad := style.Padding + 8
	if textPad < 12 {
		textPad = 12
	}
	for i := range d.Options {
		ob := rl.NewRectangle(popupRect.X, listTop+float32(i)*oh-d.popupScrollY, popW, oh)
		if ob.Y+ob.Height <= popupRect.Y || ob.Y >= popupRect.Y+popupRect.Height {
			continue
		}
		var optBg rl.Color
		var textCol rl.Color

		switch {
		case i == selectedIdx:
			optBg = rl.NewColor(0, 0, 0, 0)
			textCol = DropdownSelectedTextColor()
		case i == d.hoveredIndex:
			optBg = listRowHoverBg
			textCol = style.TextColor
		default:
			optBg = rl.NewColor(0, 0, 0, 0)
			textCol = style.TextColor
		}

		if optBg.A > 0 {
			rl.DrawRectangleRec(ob, optBg)
		}

		if i < len(d.Options)-1 {
			lineY := int32(ob.Y + ob.Height - 1)
			rl.DrawLine(
				int32(ob.X)+pad, lineY,
				int32(ob.X+ob.Width)-pad, lineY,
				divider,
			)
		}

		optPosX := int32(ob.X) + int32(textPad+0.5)
		optPosY := TextPosY(ob, itemStyle)
		optStyle := itemStyle
		optStyle.TextColor = textCol
		if i == selectedIdx {
			optStyle.Bold = true
		}
		optTextMax := ob.Width - 2*textPad
		if optTextMax < 8 {
			optTextMax = 8
		}
		drawTextS(truncateTextS(d.Options[i], optTextMax, optStyle), optPosX, optPosY, optStyle)
	}
	rl.EndScissorMode()
}

// IsInteractive implements Node.IsInteractive.
func (d *Dropdown) IsInteractive() bool { return !d.Disabled }
