// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	paginationDefaultH = float32(40)
	paginationGap      = float32(6)
	paginationPadX     = float32(10)
	paginationCellPad  = float32(3)
	paginationNavW     = float32(36)
	paginationMinPageW = float32(36)
	paginationEllipsisW = float32(28)
	// paginationShowAllMax — list every page inline up to this count; above uses ellipsis window.
	paginationShowAllMax = 12
)

type paginationSlotKind int

const (
	pgSlotPrev paginationSlotKind = iota
	pgSlotNext
	pgSlotPage
	pgSlotEllipsis
)

type paginationSlot struct {
	kind  paginationSlotKind
	page  int // 0-based when kind == pgSlotPage
	label string
}

// Pagination is a page control strip: pinned prev/next (Phosphor carets) and a scrollable
// middle of page numbers with ellipsis when TotalPages exceeds the inline threshold.
// Pair with a separate "Showing X–Y of Z" label in table footers. Current is zero-based.
//
// # LLM Prompt Template
//
//	page := ui.NewSignal(0)
//	pg := ui.NewPagination("pages", 8, page, 0, 0, 0, 0)
//	page.Subscribe(func() { reloadPage(page.Get()) })
//	footer.AddChild(pg)
//
// Demo scenes: **Batch 13 Pagination**, **List Detail Demo**.
type Pagination struct {
	Element
	TotalPages int
	Current    *Signal[int]
	hoverIdx   int
	pressedIdx int
	scrollX    float32
}

// NewPagination creates pagination. Pass w=0 for intrinsic width.
func NewPagination(id string, totalPages int, current *Signal[int], x, y, w, h float32) *Pagination {
	if current == nil {
		current = NewSignal(0)
	}
	if totalPages <= 0 {
		totalPages = 1
	}
	if current.Get() < 0 {
		current.Set(0)
	}
	if current.Get() >= totalPages {
		current.Set(totalPages - 1)
	}
	p := &Pagination{
		Element:    NewElement(id, x, y, w, h),
		TotalPages: totalPages,
		Current:    current,
		hoverIdx:   -1,
		pressedIdx: -1,
	}
	p.styleName = "pagination"
	p.SetFlexGrow(0)
	if h == 0 {
		p.AutoHeight = true
	}
	if w == 0 {
		b := p.Bounds()
		b.Width = p.stripWidth(p.buildSlots())
		p.setBoundsNoMark(b)
	}
	p.Current.Subscribe(func() {
		p.MarkDirty()
		p.MarkDrawDirty()
	})
	return p
}

func (p *Pagination) labelStyle() Style {
	fs := p.GetStyle()
	if fs.FontSize <= 0 {
		fs.FontSize = 14
	}
	if fs.TextColor.A == 0 {
		fs.TextColor = rl.NewColor(45, 48, 62, 255)
	}
	return fs
}

func (p *Pagination) slotWidth(slot paginationSlot) float32 {
	switch slot.kind {
	case pgSlotEllipsis:
		return paginationEllipsisW
	case pgSlotPrev, pgSlotNext:
		return paginationNavW
	default:
		fs := p.labelStyle()
		tw := float32(measureTextS(slot.label, fs))
		w := tw + 2*paginationPadX
		if w < paginationMinPageW {
			w = paginationMinPageW
		}
		return w
	}
}

func (p *Pagination) buildSlots() []paginationSlot {
	n := p.TotalPages
	if n <= 0 {
		n = 1
	}
	cur := p.Current.Get()
	if cur < 0 {
		cur = 0
	}
	if cur >= n {
		cur = n - 1
	}

	slots := []paginationSlot{{kind: pgSlotPrev}}
	if n <= paginationShowAllMax {
		for i := 0; i < n; i++ {
			slots = append(slots, paginationSlot{kind: pgSlotPage, page: i, label: fmt.Sprintf("%d", i+1)})
		}
	} else {
		const boundary = 1
		const sibling = 1
		show := make(map[int]struct{})
		for i := 0; i < boundary && i < n; i++ {
			show[i] = struct{}{}
		}
		for i := n - boundary; i < n; i++ {
			if i >= 0 {
				show[i] = struct{}{}
			}
		}
		for i := cur - sibling; i <= cur+sibling; i++ {
			if i >= 0 && i < n {
				show[i] = struct{}{}
			}
		}
		pages := make([]int, 0, len(show))
		for pg := range show {
			pages = append(pages, pg)
		}
		sort.Ints(pages)
		for i, pg := range pages {
			if i > 0 && pg-pages[i-1] > 1 {
				slots = append(slots, paginationSlot{kind: pgSlotEllipsis, label: "…"})
			}
			slots = append(slots, paginationSlot{kind: pgSlotPage, page: pg, label: fmt.Sprintf("%d", pg+1)})
		}
	}
	slots = append(slots, paginationSlot{kind: pgSlotNext})
	return slots
}

func (p *Pagination) middleSlots(slots []paginationSlot) []paginationSlot {
	if len(slots) <= 2 {
		return nil
	}
	return slots[1 : len(slots)-1]
}

func (p *Pagination) middleWidths(slots []paginationSlot) []float32 {
	mid := p.middleSlots(slots)
	ws := make([]float32, len(mid))
	for i, s := range mid {
		ws[i] = p.slotWidth(s)
	}
	return ws
}

func (p *Pagination) middleContentWidth(slots []paginationSlot) float32 {
	ws := p.middleWidths(slots)
	if len(ws) == 0 {
		return 0
	}
	var total float32
	for _, w := range ws {
		total += w
	}
	total += float32(len(ws)-1) * paginationGap
	return total
}

func (p *Pagination) stripWidth(slots []paginationSlot) float32 {
	return 2*paginationNavW + 2*paginationCellPad + 2*paginationGap + p.middleContentWidth(slots)
}

// layoutFrame returns the centered strip origin and visible middle viewport width.
// When the strip is wider than the widget, prev/next pin to the widget edges
// (origin stays at b.X; middleViewW is the scroll viewport between gutters).
func (p *Pagination) layoutFrame() (origin, middleViewW float32) {
	b := p.Bounds()
	slots := p.buildSlots()
	total := p.stripWidth(slots)
	midW := p.middleContentWidth(slots)
	gutter := paginationNavW + paginationCellPad + paginationGap

	if total <= b.Width {
		origin = b.X + (b.Width-total)/2
		middleViewW = midW
		return origin, middleViewW
	}
	origin = b.X
	middleViewW = b.Width - 2*gutter
	if middleViewW < 0 {
		middleViewW = 0
	}
	return origin, middleViewW
}

func (p *Pagination) stripOverflow() bool {
	b := p.Bounds()
	return p.stripWidth(p.buildSlots()) > b.Width+0.5
}

func (p *Pagination) prevBounds() rl.Rectangle {
	b := p.Bounds()
	y, h := p.navYAndH(b)
	if p.stripOverflow() {
		return rl.NewRectangle(b.X+paginationCellPad, y, paginationNavW, h)
	}
	origin, _ := p.layoutFrame()
	return rl.NewRectangle(origin+paginationCellPad, y, paginationNavW, h)
}

func (p *Pagination) nextBounds() rl.Rectangle {
	b := p.Bounds()
	y, h := p.navYAndH(b)
	if p.stripOverflow() {
		return rl.NewRectangle(b.X+b.Width-paginationNavW-paginationCellPad, y, paginationNavW, h)
	}
	origin, _ := p.layoutFrame()
	total := p.stripWidth(p.buildSlots())
	return rl.NewRectangle(origin+total-paginationNavW-paginationCellPad, y, paginationNavW, h)
}

func (p *Pagination) navYAndH(b rl.Rectangle) (y, h float32) {
	y = b.Y + paginationCellPad
	h = b.Height - 2*paginationCellPad
	if h > paginationDefaultH {
		h = paginationDefaultH
		y = b.Y + (b.Height-h)/2
	}
	if h < 28 {
		h = b.Height
		y = b.Y
	}
	return y, h
}

func (p *Pagination) middleStripBounds() rl.Rectangle {
	b := p.Bounds()
	y, h := p.navYAndH(b)
	_, viewW := p.layoutFrame()
	if p.stripOverflow() {
		left := b.X + paginationNavW + paginationCellPad + paginationGap
		return rl.NewRectangle(left, y, viewW, h)
	}
	origin, _ := p.layoutFrame()
	left := origin + paginationNavW + paginationCellPad + paginationGap
	return rl.NewRectangle(left, y, viewW, h)
}

func (p *Pagination) clampScroll(middleW, viewportW float32) {
	max := middleW - viewportW
	if max < 0 {
		max = 0
	}
	if p.scrollX < 0 {
		p.scrollX = 0
	}
	if p.scrollX > max {
		p.scrollX = max
	}
}

func (p *Pagination) controlBounds(idx int) rl.Rectangle {
	slots := p.buildSlots()
	if idx == 0 {
		return p.prevBounds()
	}
	if idx == len(slots)-1 {
		return p.nextBounds()
	}
	mid := p.middleSlots(slots)
	midIdx := idx - 1
	if midIdx < 0 || midIdx >= len(mid) {
		return rl.Rectangle{}
	}
	strip := p.middleStripBounds()
	ws := p.middleWidths(slots)
	var x float32
	for i := 0; i < midIdx; i++ {
		x += ws[i] + paginationGap
	}
	return rl.NewRectangle(strip.X+x-p.scrollX, strip.Y, ws[midIdx], strip.Height)
}

func (p *Pagination) hitControl(mouse rl.Vector2) int {
	slots := p.buildSlots()
	for i := range slots {
		if slots[i].kind == pgSlotEllipsis {
			continue
		}
		if rl.CheckCollisionPointRec(mouse, p.controlBounds(i)) {
			return i
		}
	}
	return -1
}

// IsInteractive implements Node.
func (p *Pagination) IsInteractive() bool { return true }

// HandlesWheelScroll implements wheelScrollLimiter for horizontal overflow.
func (p *Pagination) HandlesWheelScroll() bool { return true }

func (p *Pagination) scrollToCurrentSlot() {
	slots := p.buildSlots()
	cur := p.Current.Get()
	slotIdx := -1
	for i, s := range slots {
		if s.kind == pgSlotPage && s.page == cur {
			slotIdx = i
			break
		}
	}
	if slotIdx <= 0 || slotIdx >= len(slots)-1 {
		return
	}
	midIdx := slotIdx - 1
	ws := p.middleWidths(slots)
	strip := p.middleStripBounds()
	var rel float32
	for i := 0; i < midIdx; i++ {
		rel += ws[i] + paginationGap
	}
	minS := rel + ws[midIdx] - strip.Width
	maxS := rel
	if minS < 0 {
		minS = 0
	}
	if p.scrollX < minS {
		p.scrollX = minS
	}
	if p.scrollX > maxS {
		p.scrollX = maxS
	}
}

// Update handles prev/next and page selection.
func (p *Pagination) Update(_ float32) {
	if p.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	slots := p.buildSlots()
	middleW := p.middleContentWidth(slots)
	strip := p.middleStripBounds()
	prevCur := p.Current.Get()
	p.clampScroll(middleW, strip.Width)

	if rl.CheckCollisionPointRec(mouse, strip) && middleW > strip.Width+0.5 {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			p.scrollX -= wheel * paginationNavW * 2
			p.clampScroll(middleW, strip.Width)
			p.MarkDrawDirty()
		}
	}

	prevHover := p.hoverIdx
	prevPress := p.pressedIdx
	p.hoverIdx = p.hitControl(mouse)
	p.pressedIdx = -1
	if p.hoverIdx >= 0 && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		p.pressedIdx = p.hoverIdx
	}
	if p.hoverIdx != prevHover || p.pressedIdx != prevPress {
		p.MarkDrawDirty()
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && p.hoverIdx >= 0 {
		cur := p.Current.Get()
		slot := slots[p.hoverIdx]
		switch slot.kind {
		case pgSlotPrev:
			if cur > 0 {
				p.Current.Set(cur - 1)
			}
		case pgSlotNext:
			if cur < p.TotalPages-1 {
				p.Current.Set(cur + 1)
			}
		case pgSlotPage:
			p.Current.Set(slot.page)
		}
		p.MarkDrawDirty()
	}
	if p.Current.Get() != prevCur {
		p.scrollToCurrentSlot()
	}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (p *Pagination) ClearOverlayPointerState() {
	if p.hoverIdx < 0 && p.pressedIdx < 0 {
		return
	}
	p.hoverIdx = -1
	p.pressedIdx = -1
	p.MarkDrawDirty()
}

// Layout clamps intrinsic size — pagination is always a single-line strip.
func (p *Pagination) Layout() {
	defer func() { p.layoutDirty = false }()
	b := p.Bounds()
	if p.IsAutoHeight() {
		b.Height = paginationDefaultH
	} else if b.Height > paginationDefaultH+0.5 {
		b.Height = paginationDefaultH
	}
	need := p.stripWidth(p.buildSlots())
	if b.Width <= 0 {
		b.Width = need
	} else if b.Width < need && !p.stripOverflow() {
		b.Width = need
	}
	p.clampScroll(p.middleContentWidth(p.buildSlots()), p.middleStripBounds().Width)
	p.setBoundsNoMark(b)
}

// controlCount returns prev + page slots + next (for tests).
func (p *Pagination) controlCount() int { return len(p.buildSlots()) }

func drawPaginationNavIcon(rect rl.Rectangle, name string, col rl.Color) {
	iconSize := float32(18)
	if rect.Height-8 < iconSize {
		iconSize = rect.Height - 8
	}
	if iconSize < 12 {
		iconSize = 12
	}
	dst := rl.NewRectangle(
		rect.X+(rect.Width-iconSize)/2,
		rect.Y+(rect.Height-iconSize)/2,
		iconSize,
		iconSize,
	)
	if col.A == 0 {
		col = rl.NewColor(45, 48, 62, 255)
	}
	DrawPhosphorIcon(dst, name, PhosphorRegular, col)
}

// Draw implements Node.Draw.
func (p *Pagination) Draw() { p.drawInternal() }

func (p *Pagination) drawSlot(i int, slot paginationSlot, cur int, clip rl.Rectangle) {
	rect := p.controlBounds(i)
	if rect.Width <= 0 || rect.Height <= 0 {
		return
	}
	if slot.kind != pgSlotPrev && slot.kind != pgSlotNext {
		if rect.X+rect.Width < clip.X || rect.X > clip.X+clip.Width {
			return
		}
	}
	fs := p.labelStyle()

	if slot.kind == pgSlotEllipsis {
		fs.TextColor = rl.NewColor(140, 144, 160, 255)
		tw := measureTextS(slot.label, fs)
		drawTextS(slot.label, int32(rect.X+(rect.Width-float32(tw))/2), TextPosY(rect, fs), fs)
		return
	}

	selected := slot.kind == pgSlotPage && slot.page == cur
	hovered := p.hoverIdx == i
	pressed := p.pressedIdx == i
	disabled := (slot.kind == pgSlotPrev && cur <= 0) ||
		(slot.kind == pgSlotNext && cur >= p.TotalPages-1)

	inner := rect
	if hovered || pressed || selected {
		insetBtn := float32(2)
		if pressed {
			insetBtn = 3
		}
		inner = rl.NewRectangle(
			rect.X+insetBtn,
			rect.Y+insetBtn,
			rect.Width-2*insetBtn,
			rect.Height-2*insetBtn,
		)
	}

	var bg rl.Color
	switch {
	case selected:
		bg = rl.NewColor(232, 234, 255, 255)
	case pressed:
		bg = rl.NewColor(220, 224, 252, 255)
	case hovered && !disabled:
		bg = rl.NewColor(242, 244, 250, 255)
	default:
		bg = rl.NewColor(0, 0, 0, 0)
	}

	if bg.A > 0 && inner.Width > 2 && inner.Height > 2 {
		radius := inner.Height * 0.28
		drawRoundedControl(inner, inner.Width, inner.Height, radius, bg)
	}

	iconCol := fs.TextColor
	if disabled {
		iconCol = rl.NewColor(180, 184, 200, 255)
	}

	switch slot.kind {
	case pgSlotPrev:
		drawPaginationNavIcon(rect, PhosphorCaretLeft, iconCol)
		return
	case pgSlotNext:
		drawPaginationNavIcon(rect, PhosphorCaretRight, iconCol)
		return
	}

	drawFs := fs
	if selected {
		drawFs.Bold = true
		drawFs.TextColor = rl.NewColor(79, 70, 229, 255)
	}
	tw := measureTextS(slot.label, drawFs)
	tx := int32(rect.X + (rect.Width-float32(tw))/2)
	drawTextS(slot.label, tx, TextPosY(rect, drawFs), drawFs)
}

func (p *Pagination) drawInternal() {
	if p.IsHidden() {
		return
	}
	cur := p.Current.Get()
	if cur < 0 {
		cur = 0
	}
	if cur >= p.TotalPages {
		cur = p.TotalPages - 1
	}
	slots := p.buildSlots()
	middleW := p.middleContentWidth(slots)
	strip := p.middleStripBounds()
	p.clampScroll(middleW, strip.Width)

	// Prev — always pinned left, never clipped.
	if len(slots) > 0 {
		p.drawSlot(0, slots[0], cur, p.Bounds())
	}

	// Middle page numbers — clipped to strip between nav buttons.
	if strip.Width > 0 && len(slots) > 2 {
		beginScissorFromRect(strip)
		for i := 1; i < len(slots)-1; i++ {
			p.drawSlot(i, slots[i], cur, strip)
		}
		rl.EndScissorMode()
	}

	// Next — always pinned right, never clipped.
	if len(slots) > 1 {
		p.drawSlot(len(slots)-1, slots[len(slots)-1], cur, p.Bounds())
	}
}

func (p *Pagination) InteractionOverlayActive() bool { return false }

func (p *Pagination) DrawInteractionOverlay() {}
