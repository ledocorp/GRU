// Package ui (continued) — toolbar horizontal scroll (integrated gutter columns).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	tbScrollGutterW       = float32(26) // flat toolbar scroll column width
	tbRibbonScrollGutterW = float32(22) // ribbon: dedicated caret columns (narrower than icon cells)
)

// scrollGutterW returns the width of one scroll gutter column for this toolbar.
func (tb *Toolbar) scrollGutterW() float32 {
	if tb.Ribbon && tb.RibbonStacked {
		return tbRibbonScrollGutterW
	}
	return tbScrollGutterW
}

// scrollBandVertical is the Y/H band for gutters and the scroll lane (inset from toolbar borders).
func (tb *Toolbar) scrollBandVertical(inner rl.Rectangle, itemY, itemH, bw float32) (y, h float32) {
	if tb.Ribbon {
		y = itemY
		h = inner.Y + inner.Height - y - bw
	} else {
		y = inner.Y + bw
		h = inner.Height - 2*bw
	}
	if h < 1 {
		h = 1
	}
	return y, h
}

// computeScrollChromeRects sets gutter columns and the scroll lane from toolbar geometry.
// Gutters are omitted on the edge: no left column at scrollX==0, no right at scroll end.
// Ribbon items scroll in the lane and paint under the fixed gutter columns.
func (tb *Toolbar) computeScrollChromeRects(itemY, itemH, bw float32) {
	inner := tb.innerClipRect()
	bandY, bandH := tb.scrollBandVertical(inner, itemY, itemH, bw)
	contentLeft, contentRight := tb.scrollContentBounds()
	gw := tb.scrollGutterW()

	laneLeft := contentLeft
	laneRight := contentRight
	tb.scrollLeftRect = rl.Rectangle{}
	tb.scrollRightRect = rl.Rectangle{}

	if tb.showLeftScrollGutter() {
		tb.scrollLeftRect = snapControlRect(rl.NewRectangle(contentLeft, bandY, gw, bandH))
		laneLeft = contentLeft + gw
	}
	if tb.showRightScrollGutter() {
		rightX := contentRight - gw
		tb.scrollRightRect = snapControlRect(rl.NewRectangle(rightX, bandY, gw, bandH))
		laneRight = rightX
	}
	laneW := laneRight - laneLeft
	if laneW < 1 {
		laneW = 1
	}
	tb.scrollLaneRect = snapLayoutRect(rl.NewRectangle(laneLeft, bandY, laneW, bandH))
}

func (tb *Toolbar) refreshScrollChrome() {
	if !tb.scrollActive {
		tb.scrollLeftShown = false
		return
	}
	tb.updateScrollLeftGutterState()
	itemY, itemH := tb.flatRowMetrics()
	bw := tb.GetStyle().BorderWidth
	if bw <= 0 {
		bw = 1
	}
	tb.computeScrollChromeRects(itemY, itemH, bw)
}

func (tb *Toolbar) itemBandRect() rl.Rectangle {
	if tb.scrollActive {
		return tb.scrollLaneRect
	}
	inner := tb.innerClipRect()
	itemY, itemH := tb.flatRowMetrics()
	return rl.NewRectangle(inner.X, itemY, inner.Width, itemH)
}

func (tb *Toolbar) scrollContentBounds() (contentLeft, contentRight float32) {
	inner := tb.innerClipRect()
	bw := tb.GetStyle().BorderWidth
	if bw <= 0 {
		bw = 1
	}
	return inner.X + bw, inner.X + inner.Width - bw
}

func (tb *Toolbar) contentSpan() float32 {
	if len(tb.itemRects) == 0 {
		return 0
	}
	first := tb.itemRects[0]
	last := tb.itemRects[len(tb.itemRects)-1]
	return (last.X + last.Width) - first.X
}

func (tb *Toolbar) clampScrollXValue(x float32) float32 {
	if x < 0 {
		x = 0
	}
	for i := 0; i < 4; i++ {
		tb.refreshScrollChrome()
		max := tb.maxScrollX()
		if x > max {
			x = max
		}
		if x < 0 {
			x = 0
		}
	}
	return x
}

func (tb *Toolbar) applyScrollBounds() {
	flat := tb.flatItems()
	for i, item := range flat {
		if i >= len(tb.itemRects) {
			break
		}
		item.hidden = false
		if item.widget == nil {
			continue
		}
		r := tb.itemRects[i]
		r.X -= tb.scrollX
		if el, ok := item.widget.(interface{ setBoundsNoMark(rl.Rectangle) }); ok {
			el.setBoundsNoMark(r)
		} else {
			item.widget.SetBounds(r)
		}
		// Always Layout — compound items (WordToggle → nested Toggle) keep
		// children at absolute coords; skipping Layout leaves switches behind.
		item.widget.Layout()
	}
}

func (tb *Toolbar) flatItems() []*ToolbarItem {
	var out []*ToolbarItem
	for _, g := range tb.activeGroups() {
		for _, item := range g.items {
			out = append(out, item)
		}
	}
	return out
}

// scrollGutterReserve is width withheld when deciding if horizontal scroll is needed.
func (tb *Toolbar) scrollGutterReserve() float32 {
	return tb.scrollGutterW()
}

func (tb *Toolbar) updateScrollLeftGutterState() {
	if !tb.scrollActive {
		tb.scrollLeftShown = false
		return
	}
	const showAt = float32(0.75)
	const hideAt = float32(0.25)
	if tb.scrollLeftShown {
		if tb.scrollX <= hideAt {
			tb.scrollLeftShown = false
		}
	} else if tb.scrollX >= showAt {
		tb.scrollLeftShown = true
	}
}

func (tb *Toolbar) showLeftScrollGutter() bool {
	return tb.scrollActive && tb.scrollLeftShown
}

// maxScrollX is the farthest pan when the right gutter is hidden (one gutter column
// reserved at most — left only after scrollX > 0).
func (tb *Toolbar) maxScrollX() float32 {
	if !tb.scrollActive {
		return 0
	}
	contentLeft, contentRight := tb.scrollContentBounds()
	contentW := contentRight - contentLeft
	span := tb.contentSpan()
	reserve := tb.scrollGutterReserve()
	maxEnd := span - (contentW - reserve)
	if maxEnd < 0 {
		return 0
	}
	return maxEnd
}

// showRightScrollGutter uses maxScrollX, not span-view, so hiding the right column does
// not widen the viewport and immediately re-show the gutter (feedback loop).
func (tb *Toolbar) showRightScrollGutter() bool {
	if !tb.scrollActive {
		return false
	}
	contentLeft, contentRight := tb.scrollContentBounds()
	contentW := contentRight - contentLeft
	span := tb.contentSpan()
	maxEnd := tb.maxScrollX()
	if tb.scrollX >= maxEnd-0.5 {
		return false
	}
	reserve := tb.scrollGutterReserve()
	if tb.scrollX <= 0.5 {
		return span > contentW-reserve+0.5
	}
	return true
}

func (tb *Toolbar) clearToolbarItemPointerState() {
	for _, item := range tb.flatItems() {
		if item.widget == nil {
			continue
		}
		if c, ok := item.widget.(overlayPointerClearer); ok {
			c.ClearOverlayPointerState()
		}
	}
}

// HandlesWheelScroll implements wheelConsumer — horizontal ribbon scroll maps the
// vertical wheel axis and must block the document viewport on the same tick.
func (tb *Toolbar) HandlesWheelScroll() bool {
	if !tb.usesHorizontalScroll() {
		return false
	}
	// Stacked ribbon at narrow widths always maps wheel → horizontal pan over its bounds.
	if tb.Ribbon && tb.RibbonStacked {
		return true
	}
	return tb.scrollActive
}

// AbsorbsParentWheel implements wheelScrollLimiter.
func (tb *Toolbar) AbsorbsParentWheel(wheel float32) bool {
	if wheel == 0 {
		return false
	}
	return tb.absorbsRibbonWheel(rl.GetMousePosition())
}

func (tb *Toolbar) absorbsRibbonWheel(mouse rl.Vector2) bool {
	if !tb.HandlesWheelScroll() {
		return false
	}
	if !rl.CheckCollisionPointRec(mouse, tb.Bounds()) {
		return false
	}
	if tb.Ribbon && tb.RibbonStacked {
		return true
	}
	tb.refreshScrollChrome()
	if rl.CheckCollisionPointRec(mouse, tb.itemBandRect()) {
		return true
	}
	return tb.mouseOnScrollGutter(mouse)
}

func (tb *Toolbar) mouseOnScrollGutter(mouse rl.Vector2) bool {
	if !tb.scrollActive {
		return false
	}
	tb.refreshScrollChrome()
	if tb.showLeftScrollGutter() && rl.CheckCollisionPointRec(mouse, tb.scrollLeftRect) {
		return true
	}
	if tb.showRightScrollGutter() && rl.CheckCollisionPointRec(mouse, tb.scrollRightRect) {
		return true
	}
	return false
}

func (tb *Toolbar) updateHorizontalScroll(mouse rl.Vector2) {
	if !tb.scrollActive {
		return
	}
	prev := tb.scrollX
	tb.refreshScrollChrome()

	band := tb.itemBandRect()
	onBand := rl.CheckCollisionPointRec(mouse, band)
	changed := false
	if onBand {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			tb.scrollX -= wheel * tbScrollWheelScale
			changed = true
		}
	}

	// Ribbon: wheel-only scroll; flat toolbar keeps click-to-page on gutters.
	if !tb.Ribbon || !tb.RibbonStacked {
		if tb.showLeftScrollGutter() && rl.CheckCollisionPointRec(mouse, tb.scrollLeftRect) &&
			rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			tb.scrollX -= tbScrollStep
			changed = true
		}
		if tb.showRightScrollGutter() && rl.CheckCollisionPointRec(mouse, tb.scrollRightRect) &&
			rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
			tb.scrollX += tbScrollStep
			changed = true
		}
	}

	if !changed {
		return
	}

	tb.scrollX = tb.clampScrollXValue(tb.scrollX)
	if absF(tb.scrollX-prev) > 0.01 {
		tb.refreshScrollChrome()
		tb.applyScrollBounds()
		tb.MarkDrawDirty()
	}
}

// scrollCaretCenterY matches stacked ribbon icon placement (see drawStackedRibbonCell).
func (tb *Toolbar) scrollCaretCenterY(itemY float32) float32 {
	_, itemH := tb.flatRowMetrics()
	if !tb.Ribbon || !tb.RibbonStacked {
		return itemY + itemH/2
	}
	_, bandPad, cellH, _ := tb.ribbonStackedMetrics()
	const innerPad = float32(7)
	capStyle := GetThemeStyle("toolbar-cell")
	capStyle.FontSize = tbRibbonCaptionFS
	labelH := EffectiveFontSize(capStyle)
	cellTop := itemY + bandPad
	textTop := cellTop + cellH - innerPad - labelH
	iconAreaH := textTop - cellTop - innerPad
	if iconAreaH < 10 {
		iconAreaH = 10
	}
	iconSize := phosphorIconSize(iconAreaH, 22)
	if iconSize > iconAreaH {
		iconSize = iconAreaH
	}
	iconY := cellTop + innerPad + (iconAreaH-iconSize)/2
	// Caret glyphs read optically high vs filled toolbar icons.
	const caretOpticalDown = float32(2)
	return iconY + iconSize/2 + caretOpticalDown
}

// drawScrollGutters paints dedicated gutter columns on top of the clipped item lane.
func (tb *Toolbar) drawScrollGutters() {
	if !tb.scrollActive {
		return
	}
	tb.refreshScrollChrome()

	ribbonStyle := GetThemeStyle("toolbar-ribbon")
	bg := ribbonStyle.BackgroundColor
	if bg.A == 0 {
		bg = GetThemeStyle("toolbar").BackgroundColor
	}
	if bg.A == 0 {
		bg = rl.NewColor(255, 255, 255, 255)
	}
	iconCol := GetThemeStyle("toolbar-ribbon-tab-active").TextColor
	if iconCol.A == 0 {
		iconCol = GetThemeStyle("toolbar-cell").TextColor
	}
	if iconCol.A == 0 {
		iconCol = ribbonStyle.TextColor
	}

	itemY, _ := tb.flatRowMetrics()
	caretCY := tb.scrollCaretCenterY(itemY)

	drawGutter := func(r rl.Rectangle, left bool) {
		if r.Width < 1 || r.Height < 1 {
			return
		}
		rl.DrawRectangleRec(r, bg)
		iconSize := float32(14)
		if tb.Ribbon && tb.RibbonStacked {
			iconSize = 15
		}
		if iconSize > r.Width-4 {
			iconSize = r.Width - 4
		}
		dst := rl.NewRectangle(
			r.X+(r.Width-iconSize)/2,
			caretCY-iconSize/2,
			iconSize, iconSize,
		)
		name := PhosphorCaretRight
		if left {
			name = PhosphorCaretLeft
		}
		if !Phosphor.Draw(dst, name, PhosphorRegular, iconCol) {
			glyph := ">"
			if left {
				glyph = "<"
			}
			fs := float32(14)
			tw := measureTextF(glyph, fs, false, false, false, false)
			drawTextF(glyph, dst.X+(dst.Width-tw)/2, dst.Y+(dst.Height-fs)/2, fs, iconCol, true, false, false, false)
		}
	}

	if tb.showLeftScrollGutter() {
		drawGutter(tb.scrollLeftRect, true)
	}
	if tb.showRightScrollGutter() {
		drawGutter(tb.scrollRightRect, false)
	}
}
