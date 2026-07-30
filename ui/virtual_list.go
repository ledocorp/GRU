// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ItemTemplate is a function that produces a draw-only Node for a single list item.
//
// The template is called every frame for each visible item. It must return a
// lightweight, stateless Node (typically a Label) because the returned node is
// positioned and drawn immediately — it is never added to the widget tree and
// its Update/Layout methods are not called. Avoid allocating expensive state
// inside the template; use the binding for reactive data instead.
type ItemTemplate[T any] func(item T, index int, isSelected bool) Node

// VirtualList is a self-contained scrollable list widget with view culling.
//
// # Self-Contained Scrolling
//
// VirtualList manages its own scrollY independently of any parent Viewport.
// It implements HandlesWheelScroll() so the parent Viewport defers wheel
// events to it when the cursor is over the list.
//
// # View Culling
//
// Only the items whose Y range overlaps the visible window are rendered each
// frame. With 1 000+ items at itemHeight=30, this means at most ~14 Draw calls
// per frame regardless of total list size.
//
// # Scissor Clipping
//
// Row drawing uses [drawChildrenInRectClip] (same contract as [DataTable]) so
// Label Truncate cannot break the active clip stack. The body clip intersects
// ancestor viewports via [intersectRectsWithViewportAncestors].
//
// # LLM Prompt Template
//
//	binding := ui.NewListBinding([]string{"alpha", "beta", "gamma"})
//	vl := ui.NewVirtualList("files", binding, func(item string, i int, sel bool) ui.Node {
//	    lbl := ui.NewLabel(fmt.Sprintf("row-%d", i), item, 0, 0, 0, 28)
//	    if sel {
//	        lbl.SetStyleVariant("label", "selected")
//	    }
//	    return lbl
//	}, 30, 0, 0, 0, 400)
//	panel.AddChild(vl)
//
// Demo scenes: **Batch 2 SearchBar** (fruit list), JSON `virtualList` blocks.
type VirtualList[T any] struct {
	Element
	binding        *ListBinding[T]
	template       ItemTemplate[T]
	itemHeight     float32
	scrollY        float32
	contentHeight  float32
	scrollbarWidth float32

	// Scrollbar drag (same model as Viewport.handleScrollbarInput).
	sbDragging        bool
	sbDragStartMouse  float32
	sbDragStartScroll float32
}

// NewVirtualList creates a new VirtualList with the given binding and template.
func NewVirtualList[T any](id string, binding *ListBinding[T], template ItemTemplate[T], itemHeight float32, x, y, w, h float32) *VirtualList[T] {
	vl := &VirtualList[T]{
		Element:        NewElement(id, x, y, w, h),
		binding:        binding,
		template:       template,
		itemHeight:     itemHeight,
		scrollY:        0,
		contentHeight:  0,
		scrollbarWidth: 8,
	}
	binding.SubscribeItems(func() { vl.MarkDirty() })
	binding.SubscribeSelection(func() { vl.MarkDirty() })
	return vl
}

func (vl *VirtualList[T]) clampScroll() {
	if vl.scrollY < 0 {
		vl.scrollY = 0
	}
	maxScroll := vl.maxScrollY()
	if vl.scrollY > maxScroll {
		vl.scrollY = maxScroll
	}
}

func (vl *VirtualList[T]) scrollGeometry() (contentW, contentH, sbW float32) {
	style := vl.GetStyle()
	b := vl.bounds
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	contentH = b.Height - 2*bw
	contentW = b.Width - 2*bw
	sbW = 0
	if vl.contentHeight > contentH {
		sbW = vl.scrollbarWidth
		contentW -= sbW
	}
	return contentW, contentH, sbW
}

// maxScrollY is how far the list can scroll vertically.
func (vl *VirtualList[T]) maxScrollY() float32 {
	_, contentH, _ := vl.scrollGeometry()
	d := vl.contentHeight - contentH
	if d < 0 {
		return 0
	}
	return d
}

// Update handles wheel scrolling, scrollbar drag, and item selection.
func (vl *VirtualList[T]) Update(dt float32) {
	if vl.IsHidden() {
		return
	}
	bounds := vl.Bounds()
	mouse := rl.GetMousePosition()

	vl.handleScrollbarInput()

	wheel := rl.GetMouseWheelMove()
	if wheel != 0 && rl.CheckCollisionPointRec(mouse, bounds) {
		vl.scrollY -= wheel * vl.itemHeight * 3
		vl.clampScroll()
		vl.MarkDirty()
	}

	if vl.sbDragging {
		return
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, bounds) {
		if vl.hitScrollbar(mouse) {
			return
		}
		style := vl.GetStyle()
		bw := style.BorderWidth
		if bw <= 0 {
			bw = 1
		}
		localY := mouse.Y - (bounds.Y + bw) + vl.scrollY
		clickedIndex := int(localY / vl.itemHeight)
		if clickedIndex >= 0 && clickedIndex < vl.binding.Len() {
			vl.binding.SetSelectedIndex(clickedIndex)
		}
	}
}

func (vl *VirtualList[T]) hitScrollbar(mouse rl.Vector2) bool {
	_, _, sbW := vl.scrollGeometry()
	if sbW <= 0 {
		return false
	}
	b := vl.bounds
	style := vl.GetStyle()
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	trackX := b.X + b.Width - bw - sbW
	track := rl.NewRectangle(trackX, b.Y+bw, sbW, b.Height-2*bw)
	return rl.CheckCollisionPointRec(mouse, track)
}

func (vl *VirtualList[T]) thumbGeom(viewportH, barW float32) (trackX, trackY, thumbY, thumbH, maxScroll float32) {
	b := vl.bounds
	style := vl.GetStyle()
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}
	trackX = b.X + b.Width - bw - barW
	trackY = b.Y + bw
	trackH := viewportH

	thumbH = (trackH / vl.contentHeight) * trackH
	if thumbH < 16 {
		thumbH = 16
	}
	if thumbH > trackH {
		thumbH = trackH
	}
	maxScroll = vl.contentHeight - trackH
	if maxScroll > 0 {
		thumbY = trackY + (vl.scrollY/maxScroll)*(trackH-thumbH)
	} else {
		thumbY = trackY
	}
	return trackX, trackY, thumbY, thumbH, maxScroll
}

func (vl *VirtualList[T]) handleScrollbarInput() {
	_, contentH, sbW := vl.scrollGeometry()
	if sbW <= 0 {
		vl.sbDragging = false
		return
	}

	bounds := vl.Bounds()
	mouse := rl.GetMousePosition()

	if vl.sbDragging {
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
			vl.sbDragging = false
			return
		}
		_, _, _, thumbH, maxScroll := vl.thumbGeom(contentH, sbW)
		den := contentH - thumbH
		if maxScroll > 0 && den > 0 {
			vl.scrollY = vl.sbDragStartScroll + (mouse.Y-vl.sbDragStartMouse)*(maxScroll/den)
			vl.clampScroll()
			vl.MarkDirty()
		}
		return
	}

	if !rl.CheckCollisionPointRec(mouse, bounds) {
		return
	}

	trackX, trackY, thumbY, thumbH, maxScroll := vl.thumbGeom(contentH, sbW)
	track := rl.NewRectangle(trackX, trackY, sbW, contentH)
	thumbRec := rl.NewRectangle(trackX+1, thumbY, sbW-2, thumbH)

	if rl.IsMouseButtonDown(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, thumbRec) {
		vl.sbDragging = true
		vl.sbDragStartMouse = mouse.Y
		vl.sbDragStartScroll = vl.scrollY
		PointerClickMarkUsed()
		return
	}

	if rl.CheckCollisionPointRec(mouse, track) &&
		(PointerClickConsume(track) || rl.IsMouseButtonPressed(rl.MouseLeftButton)) {
		if maxScroll > 0 && contentH > 1 {
			r := (mouse.Y - trackY) / contentH
			if r < 0 {
				r = 0
			}
			if r > 1 {
				r = 1
			}
			vl.scrollY = r * maxScroll
			vl.clampScroll()
			vl.MarkDirty()
			PointerClickMarkUsed()
		}
	}
}

// Layout calculates the total content height and clamps scrollY.
func (vl *VirtualList[T]) Layout() {
	vl.contentHeight = float32(vl.binding.Len()) * vl.itemHeight
	vl.clampScroll()
	vl.layoutDirty = false
}

// Draw renders visible rows with DataTable-style scissor clipping.
func (vl *VirtualList[T]) Draw() {
	if vl.IsHidden() {
		return
	}
	style := vl.GetStyle()
	b := vl.bounds
	bw := style.BorderWidth
	if bw <= 0 {
		bw = 1
	}

	// Widget chrome — same pattern as DataTable (simple fill + hairline border).
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if style.BorderWidth > 0 {
		rl.DrawRectangleLinesEx(b, style.BorderWidth, style.BorderColor)
	}

	contentW, contentH, sbW := vl.scrollGeometry()
	contentX := b.X + bw
	contentY := b.Y + bw

	startIndex := int(vl.scrollY / vl.itemHeight)
	endIndex := int((vl.scrollY+contentH)/vl.itemHeight) + 1
	if startIndex < 0 {
		startIndex = 0
	}
	if endIndex > vl.binding.Len() {
		endIndex = vl.binding.Len()
	}

	bodyClip := rl.NewRectangle(contentX, contentY, contentW, contentH)
	bodyClip = intersectRectsWithViewportAncestors(bodyClip, vl)
	if bodyClip.Width > 0 && bodyClip.Height > 0 && endIndex > startIndex {
		rows := make([]Node, 0, endIndex-startIndex)
		for i := startIndex; i < endIndex; i++ {
			item := vl.binding.GetItem(i)
			isSelected := (i == vl.binding.GetSelectedIndex())
			node := vl.template(item, i, isSelected)
			itemY := contentY + float32(i)*vl.itemHeight - vl.scrollY
			node.SetBounds(rl.NewRectangle(contentX, itemY, contentW, vl.itemHeight))
			rows = append(rows, node)
		}
		drawChildrenInRectClip(vl, bodyClip, rows)
	}

	if sbW > 0 {
		sbClip := rl.NewRectangle(b.X, b.Y, b.Width, b.Height)
		sbClip = intersectRectsWithViewportAncestors(sbClip, vl)
		if sbClip.Width > 0 && sbClip.Height > 0 {
			beginScissorFromRect(sbClip)
			vl.drawScrollbar(contentH, sbW)
			rl.EndScissorMode()
		}
	}
}

func (vl *VirtualList[T]) drawScrollbar(viewportH, barW float32) {
	trackX, trackY, thumbY, thumbH, _ := vl.thumbGeom(viewportH, barW)
	track := rl.NewRectangle(trackX, trackY, barW, viewportH)
	rl.DrawRectangleRounded(track, 1, 6, rl.NewColor(229, 231, 235, 96))
	thumb := rl.NewRectangle(trackX+1, thumbY, barW-2, thumbH)
	rl.DrawRectangleRounded(thumb, 1, 6, rl.NewColor(148, 163, 184, 210))
}

func (vl *VirtualList[T]) HandlesWheelScroll() bool { return true }

func (vl *VirtualList[T]) AbsorbsParentWheel(wheel float32) bool {
	max := vl.maxScrollY()
	if max <= 0 {
		return false
	}
	const eps = float32(0.5)
	if wheel < 0 && vl.scrollY >= max-eps {
		return false
	}
	if wheel > 0 && vl.scrollY <= eps {
		return false
	}
	return true
}

func (vl *VirtualList[T]) UsesScissor() bool { return true }

func (vl *VirtualList[T]) IsInteractive() bool { return true }
