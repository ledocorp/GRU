// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ScrollOrientation controls the scroll axis of a Viewport.
type ScrollOrientation int

const (
	ScrollVertical   ScrollOrientation = iota // default: scroll on Y axis
	ScrollHorizontal                          // scroll on X axis (scrollbar at bottom)
)

// Viewport is a bounded, scrollable container with strict scissor clipping.
//
// # LLM Prompt Template
//
//	vp := ui.NewViewport("scroll", 0, 0, 0, 0)
//	vp.SetStyle("page-scroll")
//	vp.FlexDirection = ui.FlexColumn
//	vp.SetFlexGrow(1)
//	body := ui.NewContainer("body", 0, 0, 0, 0)
//	body.LayoutType = ui.LayoutFlex
//	body.FlexDirection = ui.FlexColumn
//	body.AutoHeight = true
//	vp.AddChild(body)
//	shell.AddChild(vp)
//
// Always use vp.AddChild (not Element.AddChild) so ClipBounds() resolves correctly.
//
// Demo scenes: **Notepad (Go)** (editor + preview viewports), **Shell Desktop Demo**,
// **Responsive Demo**, **Batch 8 Split** (nested viewports). See layout.go for
// page-scroll column template.
//
// # Responsibilities
//
// Viewport is the single source of truth for how far its children have been
// scrolled. Each Layout() call positions children at (x, y - ScrollY) so the
// visible portion always matches ScrollY. Because any child mutation marks
// the Viewport dirty (via MarkDirty propagation), Layout runs every frame
// the scroll offset changes — keeping visual position and hit-test position
// in perfect sync.
//
// # Scroll Fast-Path
//
// When the ONLY change is ScrollY (no child added, removed, resized, or
// reflowed), and this viewport's own bounds match the last full layoutFlex pass,
// Layout uses a lightweight repositionOnly pass that updates child Y positions
// without re-running calculateContentHeight or child.Layout(). If the parent
// reflowed us to a new width/height, we always run full layoutFlex so flex
// children (grids, panels) get new cross-axis sizes — never reuse the fast-path
// after a bounds change.
//
// # Scissor Clipping
//
// raylib's scissor mode is not stackable: calling BeginScissorMode inside a
// child's Draw() replaces whatever scissor was active. Viewport handles this
// by re-applying its own scissor only after children that report
// UsesScissor() == true (TextInput, Slider, ProgressBar, VirtualList).
// Static children (Label, Button, Panel) never need the re-apply, which
// reduces GPU state changes and lets raylib's batcher accumulate more draw
// calls before flushing.
//
// # Dropdown Overlay
//
// Open Dropdown menus are drawn AFTER EndScissorMode so their option list is
// never clipped by the content area. They are identified by checking whether
// the child is a *Dropdown with isOpen == true.
//
// # AddChild Override
//
// Viewport.AddChild must be used instead of the embedded Element.AddChild so
// that children's parent pointer is set to the *Viewport (not the embedded
// *Element). Children use this to type-assert their parent to *Viewport and
// call ClipBounds() for their own scissor intersection.
type Viewport struct {
	Container
	ScrollY         float32           // Current vertical scroll position (0 = top)
	ScrollX         float32           // Current horizontal scroll position (0 = left)
	Orientation     ScrollOrientation // Scroll axis: ScrollVertical (default) or ScrollHorizontal
	contentHeight   float32           // Total height of all children (vertical mode)
	contentWidth    float32           // Total width of all children (horizontal mode)
	scrollbarWidth  float32           // Width of the vertical scrollbar
	scrollbarHeight float32           // Height of the horizontal scrollbar
	scrollDirty     bool              // true when only scroll offset changed (enables fast-path)
	// lastFlexW/H record bounds after the last full vertical layoutFlex. Used so
	// the scroll reposition fast-path does not run when the parent resized this
	// viewport — children must receive layoutFlex again for correct widths.
	lastFlexW     float32
	lastFlexH     float32
	lastFlexValid bool
	// lastCommittedLayout* record vp.bounds after any completed Layout(); if the
	// parent assigned new bounds without leaving us layout-dirty, we still run.
	lastCommittedLayoutW     float32
	lastCommittedLayoutH     float32
	lastCommittedLayoutValid bool
	// ContentClipBleed expands the child scissor for nested Card/Panel shadows.
	// Panel-injected scroll hosts set this to 0 so body text cannot paint outside the band.
	ContentClipBleed float32
	// WheelScrollStep is pixels per wheel tick (0 = default 40). Notepad editor uses 80.
	WheelScrollStep float32
	// Scrollbar drag (thumb or track jump).
	sbDragging        bool
	sbDragHoriz       bool
	sbDragStartMouse  float32 // Y for vertical, X for horizontal
	sbDragStartScroll float32
}

// NewViewport creates a new Viewport with flex-column layout and clipping enabled.
func NewViewport(id string, x, y, w, h float32) *Viewport {
	vp := &Viewport{
		Container: Container{
			Element: NewElement(id, x, y, w, h),
		},
		ScrollY:         0,
		ScrollX:         0,
		contentHeight:   0,
		contentWidth:    0,
		scrollbarWidth:  8,
		scrollbarHeight: 8,
	}
	vp.LayoutType = LayoutFlex
	vp.FlexDirection = FlexColumn
	vp.Gap = 12
	vp.ClipChildren = true
	vp.ContentClipBleed = viewportContentBleed
	vp.styleName = "default"
	if h == 0 {
		vp.AutoHeight = true
	}
	return vp
}

// SetStyle sets the viewport theme and applies scroll-container defaults.
// page-scroll and settings-scroll disable ContentClipBleed so panel borders
// cannot paint into sibling chrome (MenuBar, scrollbar gutter).
func (vp *Viewport) SetStyle(name string) {
	vp.Container.SetStyle(name)
	if name == "page-scroll" || name == "settings-scroll" {
		vp.ContentClipBleed = 0
	}
}

// SetFlexGrow assigns flex-grow and clears AutoHeight so the viewport fills the
// parent scroll band instead of shrink-wrapping to scroll content height.
func (vp *Viewport) SetFlexGrow(g float32) {
	vp.FlexGrow = g
	if g > 0 {
		vp.AutoHeight = false
	}
	vp.MarkDirty()
}

// NewHorizontalViewport creates a Viewport that scrolls horizontally.
// Children are laid out in a row and the scrollbar appears at the bottom.
func NewHorizontalViewport(id string, x, y, w, h float32) *Viewport {
	vp := NewViewport(id, x, y, w, h)
	vp.Orientation = ScrollHorizontal
	vp.FlexDirection = FlexRow
	return vp
}

// viewportContentBleed expands the draw scissor slightly so Card/Panel shadows
// and button chrome are not clipped at the viewport edge.
const viewportContentBleed = float32(10)

// pageScrollBarWidth is the vertical track width for page-scroll viewports.
const pageScrollBarWidth = float32(10)

// viewportScrollTrackColor and viewportScrollThumbColor are shared scrollbar fills.
var (
	viewportScrollTrackLight = rl.NewColor(229, 231, 235, 255)
	viewportScrollThumbLight = rl.NewColor(148, 163, 184, 255)
	viewportScrollTrackColor = viewportScrollTrackLight
	viewportScrollThumbColor = viewportScrollThumbLight
)

// SetViewportScrollbarColors updates shared viewport scrollbar fills (call from SetAppearance).
func SetViewportScrollbarColors(track, thumb rl.Color) {
	viewportScrollTrackColor = track
	viewportScrollThumbColor = thumb
}

// pageScrollBarVertInset is a small inset at the top/bottom of the page scrollbar
// track so it is not flush with the viewport edge.
const pageScrollBarVertInset = float32(1)

// previewScrollPadL/V inset markdown preview scroll host on the left/top/bottom.
const previewScrollPadL = float32(12)
const previewScrollPadV = float32(4)

// pageScrollPadH is the horizontal inset for page-scroll (content + scrollbar band).
const pageScrollPadH = float32(16)

// pageScrollPadV is the vertical inset for page-scroll — kept small so the page
// uses the full band below the title bar and above the launcher nav.
const pageScrollPadV = float32(4)

// settingsScrollPadH is horizontal inset for desktop in-shell settings (MenuBar
// / StatusBar bands stay flush — no white gutter above/below the scroll client).
const settingsScrollPadH = float32(16)

// horizontalContentBorderSlack extends the horizontal content scissor slightly
// above the scrollbar gutter so inset card border strokes are not sheared off.
const horizontalContentBorderSlack = float32(8)

const defaultWheelScrollStep = float32(40)

func (vp *Viewport) wheelScrollStep() float32 {
	if vp.WheelScrollStep > 0 {
		return vp.WheelScrollStep
	}
	return defaultWheelScrollStep
}

// HandlesWheelScroll implements the wheelConsumer interface.
//
// Only vertical viewports opt in. A horizontal strip inside a vertically
// scrolling page must not block page scroll when the cursor is over it — the
// strip still scrolls on X in Update() from the same wheel delta.
//
// Vertical viewports opt in. Exclusive routing (page vs nested) is handled by
// PrepareWheelScroll / WheelScrollOwner before Update.
func (vp *Viewport) HandlesWheelScroll() bool {
	return vp.Orientation == ScrollVertical
}

func (vp *Viewport) isPageLevelVerticalViewport() bool {
	return vp.Orientation == ScrollVertical && !vp.hasAncestorVerticalViewport()
}

func (vp *Viewport) hasAncestorVerticalViewport() bool {
	for p := vp.ParentNode(); p != nil; p = p.ParentNode() {
		if ancestor, ok := p.(*Viewport); ok && ancestor != vp && ancestor.Orientation == ScrollVertical {
			return true
		}
	}
	return false
}

// AbsorbsParentWheel implements wheelScrollLimiter. When this viewport is at
// its scroll limit in the wheel direction, the parent may scroll instead.
func (vp *Viewport) AbsorbsParentWheel(wheel float32) bool {
	if vp.Orientation == ScrollHorizontal {
		return false
	}
	max := vp.overflowScrollY()
	if max <= 0 {
		return false
	}
	const eps = float32(0.5)
	if wheel < 0 && vp.ScrollY >= max-eps {
		return false
	}
	if wheel > 0 && vp.ScrollY <= eps {
		return false
	}
	return true
}

// overflowScrollX returns how far horizontal content exceeds the viewport width.
// Matches calculateContentWidth vs outer bounds (same model as layoutHorizontal clamp).
func (vp *Viewport) overflowScrollX() float32 {
	d := vp.contentWidth - vp.bounds.Width
	if d < 0 {
		return 0
	}
	return d
}

// overflowScrollY returns how far vertical content exceeds the viewport height.
// When a Dropdown or ComboBox popup is open and extends below the visible client,
// extra scroll range is included so the viewport can scroll to reveal the full list.
func (vp *Viewport) overflowScrollY() float32 {
	d := vp.contentHeight - vp.bounds.Height
	if d < 0 {
		d = 0
	}
	if ext := vp.openMenuPopupScrollExtension(); ext > 0 {
		d += ext
	}
	return d
}

// openMenuPopupScrollExtension is additional scroll slack while a menu popup
// hangs below the padded client bottom (popups are not part of contentHeight).
func (vp *Viewport) 	openMenuPopupScrollExtension() float32 {
	if vp.IsHidden() || vp.Orientation == ScrollHorizontal {
		return 0
	}
	client := vp.viewportPaddedClientRect()
	clientTop := client.Y
	clientBottom := client.Y + client.Height
	var extra float32
	for _, dd := range collectOpenDropdowns(vp.children) {
		pop := dd.PopupBounds()
		if pop.Height <= 0 {
			continue
		}
		if over := pop.Y + pop.Height - clientBottom; over > extra {
			extra = over
		}
		if over := clientTop - pop.Y; over > extra {
			extra = over
		}
	}
	for _, cb := range collectOpenComboBoxes(vp.children) {
		pop := cb.PopupBounds()
		if pop.Height <= 0 {
			continue
		}
		if over := pop.Y + pop.Height - clientBottom; over > extra {
			extra = over
		}
		if over := clientTop - pop.Y; over > extra {
			extra = over
		}
	}
	return extra
}

// parent (instead of the embedded *Element). Children rely on this to
// type-assert their parent to *Viewport and call ClipBounds() when they need
// to intersect their own scissor rectangle with the Viewport boundary.
func (vp *Viewport) AddChild(child Node) {
	vp.children = append(vp.children, child)
	child.SetParent(vp)
	vp.MarkDirty()
}

// ClipBounds returns the content clipping rectangle for this Viewport.
//
// Children that apply their own scissor (e.g. VirtualList) call this method
// to intersect their scissor with the Viewport's visible area, preventing
// content from escaping outside the grey panel background.
func (vp *Viewport) ClipBounds() rl.Rectangle {
	x, y, w, h := vp.contentScissorRect()
	return rl.NewRectangle(x, y, w, h)
}

// Update handles mouse wheel scrolling and forwards input to all children.
//
// Wheel events are consumed by the Viewport unless a child under the cursor
// implements the wheelConsumer interface (HandlesWheelScroll() bool returns true).
// Regular interactive widgets — Button, Slider, TextInput, Checkbox, Dropdown —
// do NOT implement this interface, so the Viewport still scrolls when the mouse
// is over them. Only VirtualList opts in, so it can manage its own independent
// scroll offset without conflicting with the outer Viewport scroll.
func (vp *Viewport) Update(dt float32) {
	if vp.IsHidden() {
		return
	}

	vp.handleScrollbarInput()

	wheel := rl.GetMouseWheelMove()
	if wheel != 0 && !vp.sbDragging && !chromeWindowMoving {
		if vp.Orientation == ScrollHorizontal {
			mouse := rl.GetMousePosition()
			if rl.CheckCollisionPointRec(mouse, vp.Bounds()) {
				oldX := vp.ScrollX
				vp.ScrollX += wheel * vp.wheelScrollStep()
				maxScrollX := vp.overflowScrollX()
				if vp.ScrollX < 0 {
					vp.ScrollX = 0
				}
				if vp.ScrollX > maxScrollX {
					vp.ScrollX = maxScrollX
				}
				if vp.ScrollX != oldX {
					vp.scrollDirty = true
					vp.MarkDirty()
				}
			}
		} else if WheelScrollOwner() == vp {
			mouse := rl.GetMousePosition()
			if deepHasWheelConsumer(vp.children, mouse, wheel) {
				// Nested wheel consumer (DataTable, VirtualList, …) owns this tick.
			} else {
			oldScrollY := vp.ScrollY
			vp.ScrollY -= wheel * vp.wheelScrollStep()
			maxScroll := vp.overflowScrollY()
			if vp.ScrollY < 0 {
				vp.ScrollY = 0
			}
			if vp.ScrollY > maxScroll {
				vp.ScrollY = maxScroll
			}
			if vp.ScrollY != oldScrollY {
				vp.scrollDirty = true
				vp.MarkDirty()
				vp.MarkDrawDirty()
			}
			if vp.isPageLevelVerticalViewport() {
				notePageWheelGesture()
			} else {
				noteNestedWheelGesture()
			}
			}
		}
	}
	// After wheel scroll, child bounds are still from the previous Layout pass until
	// repositionOnly runs. Sync Y positions before Update so clicks hit widgets
	// under the cursor on the same frame (especially switches in list rows).
	if vp.scrollDirty && vp.Orientation != ScrollHorizontal && !viewportFlexGeomStale(vp) {
		childNeedsLayout := false
		for _, child := range vp.children {
			if !child.IsHidden() && child.IsDirty() {
				childNeedsLayout = true
				break
			}
		}
		if !childNeedsLayout {
			vp.repositionOnly()
		}
	}
	UpdateChildrenOverlayAware(vp.Children(), dt)
}

// viewportFlexGeomStale is true when vp.bounds differ from the dimensions used
// in the last full vertical layoutFlex (lastFlex*). Parent SetBounds can change
// our width without toggling layoutDirty; this forces a full flex pass.
// viewportFlexGeomStale is true when the last full layoutFlex geometry does not
// match vp.bounds — including parent-driven growth (wider/taller than last pass)
// so flex children and grids must not use the scroll reposition fast-path.
func viewportFlexGeomStale(vp *Viewport) bool {
	return !vp.lastFlexValid || vp.bounds.Width != vp.lastFlexW || vp.bounds.Height != vp.lastFlexH
}

// Layout positions children at their scrolled Y coordinates.
// When vp.bounds differ from the last full layoutFlex snapshot (lastFlex*),
// layout always runs full layoutFlex (scroll repositionOnly is skipped).
func (vp *Viewport) Layout() {
	if vp.IsHidden() {
		vp.layoutDirty = false
		return
	}
	if viewportScrollRecovering() {
		vp.scrollDirty = false
		vp.lastFlexValid = false
		vp.lastCommittedLayoutValid = false
		vp.lastFlexW = 0
		vp.lastFlexH = 0
		vp.layoutDirty = true
	}
	boundsChanged := viewportFlexGeomStale(vp)
	if boundsChanged {
		vp.layoutDirty = true
		vp.scrollDirty = false
	}
	if !vp.IsDirty() && !vp.scrollDirty && !boundsChanged && !viewportScrollRecovering() {
		if SubtreeLayoutDirty(vp) {
			goto layoutViewport
		}
		return
	}
layoutViewport:
	if vp.Orientation == ScrollHorizontal {
		vp.layoutHorizontal()
		vp.scrollDirty = false
		vp.layoutDirty = false
		vp.lastCommittedLayoutW = vp.bounds.Width
		vp.lastCommittedLayoutH = vp.bounds.Height
		vp.lastCommittedLayoutValid = true
		vp.lastLayoutPassW = vp.bounds.Width
		vp.lastLayoutPassH = vp.bounds.Height
		vp.lastLayoutPassValid = true
		return
	}

	childNeedsLayout := false
	for _, child := range vp.children {
		if !child.IsHidden() && child.IsDirty() {
			childNeedsLayout = true
			break
		}
	}

	if vp.scrollDirty && !childNeedsLayout && !boundsChanged {
		vp.repositionOnly()
	} else {
		prevFlexW := vp.lastFlexW
		savedW := vp.bounds.Width
		if boundsChanged {
			for _, ch := range vp.children {
				if !ch.IsHidden() {
					ch.MarkDirty()
				}
			}
		}
		vp.layoutVerticalScrollContent()
		// Widen after shrink: run flex again so column children (grids, panels)
		// pick up cross-axis width from layoutFlex; then subtree Layout() sees it.
		if vp.lastFlexValid && savedW > prevFlexW+0.5 {
			for _, ch := range vp.children {
				if !ch.IsHidden() {
					ch.MarkDirty()
				}
			}
			vp.layoutVerticalScrollContent()
		}

		for _, child := range vp.children {
			if child.IsHidden() || nodeIsFloatOverlay(child) {
				continue
			}
			b := child.Bounds()
			b.Y -= vp.ScrollY
			scrollTranslate(child, b)
		}

		vp.layoutFloatOverlays()

		for _, child := range vp.children {
			if !child.IsHidden() {
				child.Layout()
			}
		}
		vp.restackVerticalChildrenAfterLayout()
		// Measure scroll range after subtree layout — AutoHeight panels and xs
		// grids grow in child.Layout(); an early sum underestimates maxScroll at
		// minimum width and makes wheel scroll feel stuck.
		vp.calculateContentHeight()
		if maxScroll := vp.overflowScrollY(); vp.ScrollY > maxScroll {
			vp.ScrollY = maxScroll
		}
		vp.lastFlexValid = true
		vp.lastFlexW = vp.bounds.Width
		vp.lastFlexH = vp.bounds.Height
	}

	vp.scrollDirty = false
	vp.layoutDirty = false
	vp.lastCommittedLayoutW = vp.bounds.Width
	vp.lastCommittedLayoutH = vp.bounds.Height
	vp.lastCommittedLayoutValid = true
	vp.lastLayoutPassW = vp.bounds.Width
	vp.lastLayoutPassH = vp.bounds.Height
	vp.lastLayoutPassValid = true
}

// layoutHorizontal positions children left-to-right using shared layoutFlex(),
// then offsets each child’s X by −ScrollX.
func (vp *Viewport) layoutHorizontal() {
	if viewportScrollRecovering() {
		vp.scrollDirty = false
		vp.lastFlexValid = false
		vp.lastCommittedLayoutValid = false
		vp.lastFlexW = 0
		vp.lastFlexH = 0
		vp.layoutDirty = true
	}
	if viewportFlexGeomStale(vp) {
		vp.scrollDirty = false
		vp.lastFlexValid = false
	}
	boundsChangedSinceLayout := !vp.lastCommittedLayoutValid || vp.lastCommittedLayoutW != vp.bounds.Width || vp.lastCommittedLayoutH != vp.bounds.Height
	if boundsChangedSinceLayout || viewportFlexGeomStale(vp) {
		for _, ch := range vp.children {
			if !ch.IsHidden() {
				ch.MarkDirty()
			}
		}
	}
	// Route through the shared layoutFlex() so FlexGrow distribution and
	// cross-axis stretch (h=0 → fill height) work identically to Container.
	// Temporarily shrink bounds.Height by scrollbarHeight so that h=0 children
	// don’t overlap the horizontal scrollbar gutter.
	savedH := vp.bounds.Height
	vp.bounds.Height -= vp.scrollbarHeight
	vp.Container.layoutFlex()
	vp.bounds.Height = savedH

	// Apply horizontal scroll offset without triggering MarkDirty.
	for _, child := range vp.children {
		if child.IsHidden() {
			continue
		}
		b := child.Bounds()
		b.X -= vp.ScrollX
		scrollTranslate(child, b)
	}

	// Recurse into every child after horizontal flex layout.
	for _, child := range vp.children {
		if !child.IsHidden() {
			child.Layout()
		}
	}
	vp.calculateContentWidth()
	if maxScroll := vp.overflowScrollX(); vp.ScrollX > maxScroll {
		vp.ScrollX = maxScroll
	}
	// Shrink-wrap height for AutoHeight horizontal strips (e.g. preview chrome wider than panel).
	if vp.AutoHeight && vp.GetFlexGrow() == 0 {
		style := vp.GetStyle()
		padding := style.Padding
		maxChildH := float32(0)
		for _, child := range vp.children {
			if child.IsHidden() {
				continue
			}
			h := child.Bounds().Height
			// FlexRow lanes may still report h=0 until their own Layout() pass even
			// though descendants (e.g. fixed-height Cards) are already positioned.
			if sub := nodeSubtreeBottom(child); sub > child.Bounds().Y {
				if contentH := sub - child.Bounds().Y; contentH > h {
					h = contentH
				}
			}
			if h > maxChildH {
				maxChildH = h
			}
		}
		want := maxChildH + 2*padding + vp.scrollbarHeight
		if want > 0 && (vp.bounds.Height < want-0.5 || vp.bounds.Height > want+0.5) {
			vp.bounds.Height = want
		}
	}
	vp.lastFlexValid = true
	vp.lastFlexW = vp.bounds.Width
	vp.lastFlexH = vp.bounds.Height
	vp.lastCommittedLayoutW = vp.bounds.Width
	vp.lastCommittedLayoutH = vp.bounds.Height
	vp.lastCommittedLayoutValid = true
	vp.lastLayoutPassW = vp.bounds.Width
	vp.lastLayoutPassH = vp.bounds.Height
	vp.lastLayoutPassValid = true
}

// calculateContentWidth sums all children widths plus padding (horizontal mode).
func (vp *Viewport) calculateContentWidth() {
	total := float32(0)
	for i, child := range vp.children {
		if child.IsHidden() {
			continue
		}
		if i > 0 {
			total += vp.Gap
		}
		total += child.Bounds().Width
	}
	padL, _, padR, _ := vp.scrollContentPadding()
	vp.contentWidth = total + padL + padR
}

// repositionOnly updates child Y positions for the current ScrollY without
// re-running flex layout or child.Layout(). Used by the scroll fast-path.
//
// # Root-cause of the scroll regression (Phase 2)
//
// The original Phase 2 implementation only moved each direct Viewport child's
// bounds.Y (e.g. the Panel box itself) but never propagated the same Y offset
// to the Panel's own children (Labels, TextInputs, etc.). Those inner widgets
// kept their pre-scroll absolute Y coordinates, causing them to appear frozen
// in place while only the Panel border shifted.
//
// # Fix: shiftSubtreeY
//
// After moving a direct child, we recursively walk its subtree and apply the
// same Δy to every descendant's bounds — silently, without triggering
// MarkDirty propagation.
//
// # UsesScissor children (VirtualList, TextInput, Slider, ProgressBar)
//
// These widgets derive their clip rectangles from their own bounds at draw
// time, so a bounds update alone is sufficient. As an additional safety guard,
// we still call Layout() on them (it is always idempotent and for VirtualList
// it has no dirty guard — running it recalculates contentHeight and clamps
// scrollY in case the binding changed just before the scroll).
func (vp *Viewport) repositionOnly() {
	padL, padT, _, _ := vp.scrollContentPadding()
	x := vp.bounds.X + padL
	y := vp.bounds.Y + padT
	moved := false
	for _, child := range vp.children {
		if child.IsHidden() || nodeIsFloatOverlay(child) {
			continue
		}
		b := child.Bounds()
		newY := y - vp.ScrollY
		dy := newY - b.Y
		if dy != 0 {
			moved = true
		}
		b.X = x
		b.Y = newY
		scrollTranslate(child, b)

		// Cascade the same Δy through the entire child subtree.
		if dy != 0 {
			if ss, ok := child.(scrollSubtreeShifter); ok {
				ss.ShiftScrollSubtreeInternal(dy)
			} else {
				shiftSubtreeY(child.Children(), dy)
			}
		}

		// Widgets that own a scissor region or manage their own scroll
		// get an explicit Layout() to refresh contentHeight / clip state.
		if child.UsesScissor() {
			child.Layout()
		}

		y += b.Height + vp.Gap
	}
	vp.layoutFloatOverlays()
	if moved {
		vp.MarkDrawDirty()
	}
}

// layoutVerticalScrollContent lays out vertical viewport children without
// clamping their main-axis height to the visible viewport. A viewport is a
// scroll container: children must be allowed to keep their intrinsic content
// height so calculateContentHeight can expose overflow through ScrollY.
func (vp *Viewport) layoutVerticalScrollContent() {
	padL, padT, _, padB := vp.scrollContentPadding()
	availableW := vp.scrollContentWidthBudget(vp.bounds)
	availableH := vp.bounds.Height - padT - padB
	if availableW < 0 {
		availableW = 0
	}
	if availableH < 0 {
		availableH = 0
	}

	x := vp.bounds.X + padL
	y := vp.bounds.Y + padT

	for _, child := range vp.children {
		if child.IsHidden() || nodeIsFloatOverlay(child) || child.GetFlexGrow() != 0 || !child.IsAutoHeight() {
			continue
		}
		b := child.Bounds()
		if b.Width == 0 || b.Width > availableW {
			b.Width = availableW
		}
		b.Height = 0
		child.SetBounds(b)
		child.Layout()
	}

	nonHidden := 0
	fixedH := float32(0)
	totalGrow := float32(0)
	for _, child := range vp.children {
		if child.IsHidden() || nodeIsFloatOverlay(child) {
			continue
		}
		nonHidden++
		if fg := child.GetFlexGrow(); fg > 0 {
			totalGrow += fg
		} else {
			fixedH += child.Bounds().Height
		}
	}

	gaps := float32(0)
	if nonHidden > 1 {
		gaps = float32(nonHidden-1) * vp.Gap
	}
	remaining := availableH - fixedH - gaps
	if remaining < 0 {
		remaining = 0
	}

	for _, child := range vp.children {
		if child.IsHidden() || nodeIsFloatOverlay(child) {
			continue
		}
		b := child.Bounds()
		b.X = x
		b.Y = y
		if fg := child.GetFlexGrow(); fg > 0 && totalGrow > 0 {
			b.Height = remaining * (fg / totalGrow)
		}
		if shouldStretchColumnCrossAxis(&vp.Container, child, availableW) {
			b.Width = availableW
		}
		child.SetBounds(b)
		y += b.Height + vp.Gap
	}

	vp.lastLayoutPassW = vp.bounds.Width
	vp.lastLayoutPassH = vp.bounds.Height
	vp.lastLayoutPassValid = true
}

// restackVerticalChildrenAfterLayout fixes main-axis Y (and cross-axis width when
// stale) after AutoHeight children (Header, preview blocks) change height during
// child.Layout(). Viewport positions children before the subtree pass; without
// this, wrapped headers overlap flex-grow panels below until the next frame.
func (vp *Viewport) restackVerticalChildrenAfterLayout() {
	if vp.Orientation == ScrollHorizontal {
		return
	}
	padL, padT, _, _ := vp.scrollContentPadding()
	x := vp.bounds.X + padL
	contentY := vp.bounds.Y + padT
	availableW := vp.scrollContentWidthBudget(vp.bounds)
	for _, child := range vp.children {
		if child.IsHidden() || nodeIsFloatOverlay(child) {
			continue
		}
		b := child.Bounds()
		widthChanged := false
		if shouldStretchColumnCrossAxis(&vp.Container, child, availableW) && absF(b.Width-availableW) > 0.5 {
			b.Width = availableW
			widthChanged = true
		}
		b.X = x
		b.Y = contentY - vp.ScrollY
		if widthChanged {
			child.SetBounds(b)
			child.Layout()
		} else {
			layoutSetBounds(child, b)
		}
		contentY += child.Bounds().Height + vp.Gap
	}
}

// shiftSubtreeY translates every node in the slice — and all their descendants
// — by dy pixels on the Y axis, without triggering MarkDirty propagation.
//
// This is O(N) in the total number of descendant widgets and runs only when the
// parent Viewport's scroll changed (the fast-path). On typical scenes with a
// few Panels each containing ~5–15 widgets this is 30–80 operations — much
// cheaper than a full recursive Layout() pass.
//
// scrollSubtreeShifter is implemented by Card/Panel shells whose user content
// lives in an internal RaisedSurface body (see ShiftScrollSubtreeInternal).
type scrollSubtreeShifter interface {
	ShiftScrollSubtreeInternal(dy float32)
}

func shiftSubtreeY(nodes []Node, dy float32) {
	for _, n := range nodes {
		b := n.Bounds()
		b.Y += dy
		scrollTranslate(n, b)
		if ss, ok := n.(scrollSubtreeShifter); ok {
			ss.ShiftScrollSubtreeInternal(dy)
			continue
		}
		if len(n.Children()) > 0 {
			shiftSubtreeY(n.Children(), dy)
		}
	}
}

// scrollTranslate updates a child's bounds without triggering MarkDirty.
// Uses the unexported setBoundsNoMark interface implemented by Element,
// falling back to SetBounds (which triggers MarkDirty) for any future node
// type that does not embed Element.
func scrollTranslate(n Node, b rl.Rectangle) {
	type boundsWriter interface{ setBoundsNoMark(rl.Rectangle) }
	if bw, ok := n.(boundsWriter); ok {
		bw.setBoundsNoMark(b)
		return
	}
	n.SetBounds(b) // fallback: triggers MarkDirty but Layout will run
}

// calculateContentHeight sums all children heights plus padding.
func (vp *Viewport) calculateContentHeight() {
	total := float32(0)
	first := true
	for _, child := range vp.children {
		if child.IsHidden() || nodeIsFloatOverlay(child) {
			continue
		}
		if !first {
			total += vp.Gap
		}
		first = false
		total += child.Bounds().Height
	}
	_, padT, _, padB := vp.scrollContentPadding()
	vp.contentHeight = total + padT + padB
}

// collectOpenDropdowns walks the full node subtree and finds every Dropdown
// whose popup is currently open. The Viewport calls this after EndScissorMode
// so popups are drawn on top of all other content, unclipped.
func collectOpenDropdowns(nodes []Node) []*Dropdown {
	var result []*Dropdown
	for _, child := range nodes {
		if child.IsHidden() {
			continue
		}
		if dd, ok := child.(*Dropdown); ok && dd.IsOpen() {
			result = append(result, dd)
		}
		result = append(result, collectOpenDropdowns(child.Children())...)
	}
	return result
}

func collectOpenComboBoxes(nodes []Node) []*ComboBox {
	var result []*ComboBox
	for _, child := range nodes {
		if child.IsHidden() {
			continue
		}
		if cb, ok := child.(*ComboBox); ok && cb.IsOpen() {
			result = append(result, cb)
		}
		result = append(result, collectOpenComboBoxes(child.Children())...)
	}
	return result
}

func collectOpenToolbars(nodes []Node) []*Toolbar {
	var result []*Toolbar
	for _, child := range nodes {
		if child.IsHidden() {
			continue
		}
		if tb, ok := child.(*Toolbar); ok && tb.overflowOpen {
			result = append(result, tb)
		}
		result = append(result, collectOpenToolbars(child.Children())...)
	}
	return result
}

func (vp *Viewport) scrollContentPadding() (left, top, right, bottom float32) {
	switch vp.styleName {
	case "page-scroll":
		return pageScrollPadH, pageScrollPadV, pageScrollPadH, pageScrollPadV
	case "settings-scroll":
		return settingsScrollPadH, 0, settingsScrollPadH, 0
	case "preview-scroll":
		// Horizontal inset on the left; lane padding handles right gutter before scrollbar.
		return previewScrollPadL, previewScrollPadV, 0, previewScrollPadV
	}
	p := vp.GetStyle().Padding
	return p, p, p, p
}

func (vp *Viewport) verticalScrollbarWidth() float32 {
	switch vp.styleName {
	case "page-scroll", "preview-scroll", "editor-scroll":
		return pageScrollBarWidth
	}
	return vp.scrollbarWidth
}

func (vp *Viewport) scrollBarLeadingGap() float32 {
	return 0
}

func (vp *Viewport) floatOverlayHostRect() rl.Rectangle {
	b := vp.Bounds()
	if b.Width < 1 {
		b.Width = 1
	}
	if b.Height < 1 {
		b.Height = 1
	}
	return b
}

func (vp *Viewport) floatOverlayClipRect() rl.Rectangle {
	if vp.Orientation == ScrollVertical {
		x, y, w, h := vp.contentScissorRect()
		return rl.NewRectangle(x, y, w, h)
	}
	return vp.floatOverlayHostRect()
}

// viewportPaddedClientRect is the symmetric inset band inside the viewport bounds.
// page-scroll content, clip, and float overlays share this rectangle.
func (vp *Viewport) viewportPaddedClientRect() rl.Rectangle {
	b := vp.Bounds()
	padL, padT, padR, padB := vp.scrollContentPadding()
	w := b.Width - padL - padR
	h := b.Height - padT - padB
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return rl.NewRectangle(b.X+padL, b.Y+padT, w, h)
}

func (vp *Viewport) scrollContentWidthBudget(bounds rl.Rectangle) float32 {
	padL, _, padR, _ := vp.scrollContentPadding()
	if vp.styleName == "page-scroll" {
		w := bounds.Width - padL - padR
		if w < 1 {
			return 1
		}
		return w
	}
	if vp.styleName == "preview-scroll" {
		w := bounds.Width - padL - vp.verticalScrollbarWidth()
		if w < 1 {
			return 1
		}
		return w
	}
	w := bounds.Width - padL - padR - vp.verticalScrollbarWidth() - vp.scrollBarLeadingGap()
	if w < 1 {
		return 1
	}
	return w
}

// pageScrollVertTrackRect returns the vertical scrollbar track for page-scroll
// viewports: centered in the right padding band, inset slightly from top/bottom.
func (vp *Viewport) pageScrollVertTrackRect(bounds rl.Rectangle) (trackX, trackY, trackW, trackH float32) {
	_, padT, padR, padB := vp.scrollContentPadding()
	barW := vp.verticalScrollbarWidth()
	trackW = barW
	trackX = bounds.X + bounds.Width - padR + (padR-barW)/2
	if trackX < bounds.X {
		trackX = bounds.X
	}
	trackY = bounds.Y + padT + pageScrollBarVertInset
	trackH = bounds.Height - padT - padB - 2*pageScrollBarVertInset
	if trackH < 1 {
		trackH = 1
	}
	return trackX, trackY, trackW, trackH
}

// contentScissorRect returns the scissor rectangle for child drawing, expanded by
// viewportContentBleed so raised surfaces are not clipped at the content edge.
func (vp *Viewport) contentScissorRect() (x, y, w, h float32) {
	bounds := vp.Bounds()
	padL, padT, padR, padB := vp.scrollContentPadding()
	if vp.Orientation == ScrollHorizontal {
		x = bounds.X + padL
		y = bounds.Y + padT
		w = bounds.Width - padL - padR
		h = bounds.Height - padT - padB - vp.scrollbarHeight + horizontalContentBorderSlack
	} else {
		switch vp.styleName {
		case "page-scroll":
			client := vp.viewportPaddedClientRect()
			x, y, w, h = client.X, client.Y, client.Width, client.Height
		case "preview-scroll":
			x = bounds.X + padL
			y = bounds.Y + padT
			w = bounds.Width - padL - vp.verticalScrollbarWidth()
			h = bounds.Height - padT - padB
		default:
			x = bounds.X + padL
			y = bounds.Y + padT
			w = vp.scrollContentWidthBudget(bounds)
			h = bounds.Height - padT - padB
		}
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}

	// Horizontal strips clip to the card/panel body; bleed would let nested cards
	// paint into the caption row above the scroll band (responsive demo).
	// page-scroll keeps bleed at 0 so content stays inside the padded band.
	bleed := vp.ContentClipBleed
	if vp.Orientation == ScrollHorizontal || vp.styleName == "page-scroll" {
		bleed = 0
	}
	x -= bleed
	y -= bleed
	w += bleed * 2
	h += bleed * 2

	if x < bounds.X {
		w -= bounds.X - x
		x = bounds.X
	}
	if y < bounds.Y {
		h -= bounds.Y - y
		y = bounds.Y
	}
	right := bounds.X + bounds.Width
	bottom := bounds.Y + bounds.Height
	if x+w > right {
		w = right - x
	}
	if y+h > bottom {
		h = bottom - y
	}
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return x, y, w, h
}

// Draw renders the background, clips children to the content area, then draws the scrollbar.
func (vp *Viewport) Draw() {
	defer func() { vp.drawDirty = false }()
	if vp.IsHidden() {
		return
	}

	style := vp.GetStyle()
	bounds := vp.Bounds()

	// Background and border drawn outside the scissor.
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangle(int32(bounds.X), int32(bounds.Y), int32(bounds.Width), int32(bounds.Height), style.BackgroundColor)
		if style.BorderWidth > 0 {
			rl.DrawRectangleLinesEx(bounds, float32(style.BorderWidth), style.BorderColor)
		}
	}

	contentX, contentY, contentW, contentH := vp.contentScissorRect()

	contentClip := intersectRectsWithViewportAncestors(
		rl.NewRectangle(contentX, contentY, contentW, contentH), vp)
	if contentClip.Width >= 1 && contentClip.Height >= 1 {
		beginScissorFromRect(contentClip)
		setActiveDrawClip(contentClip)
		for _, child := range vp.children {
			if nodeIsFloatOverlay(child) {
				continue
			}
			child.Draw()
			beginScissorFromRect(contentClip)
			setActiveDrawClip(contentClip)
		}
		clearActiveDrawClip()
		rl.EndScissorMode()
	}

	for _, overlay := range collectFloatOverlays(vp.children) {
		clip := vp.floatOverlayClipRect()
		clip = intersectRectsWithViewportAncestors(clip, vp)
		if clip.Width < 1 || clip.Height < 1 {
			continue
		}
		beginScissorFromRect(clip)
		overlay.Draw()
		rl.EndScissorMode()
	}

	// Scrollbars are drawn after the content scissor ends (track lives in the
	// chrome gutter), but must still be clipped to this Viewport's outer bounds.
	// Otherwise, when a parent Viewport restores its scissor between Panel
	// children, the thumb/track can paint on top of earlier siblings (e.g. the
	// grid section above a horizontal strip).
	ob := intersectRectsWithViewportAncestors(vp.Bounds(), vp)
	if ob.Width >= 1 && ob.Height >= 1 {
		beginScissorFromRect(ob)
		if vp.Orientation == ScrollHorizontal {
			if vp.overflowScrollX() > 0 {
				vp.drawHorizontalScrollbar()
			}
		} else {
			if vp.overflowScrollY() > 0 {
				vp.drawScrollbar()
			}
		}
		rl.EndScissorMode()
	}

	// Dropdown / ComboBox popups are drawn at screen level (DrawOpenMenuPopups).
}

// drawScrollbar renders the vertical scrollbar track and thumb.
func (vp *Viewport) drawScrollbar() {
	bounds := vp.Bounds()
	trackX, trackY, thumbY, thumbH, trackW, trackH := vp.vertScrollbarDrawGeom(bounds)
	barW := trackW
	if barW < 1 {
		barW = vp.verticalScrollbarWidth()
	}

	track := rl.NewRectangle(trackX, trackY, barW, trackH)
	rl.DrawRectangleRounded(track, 1, 6, viewportScrollTrackColor)

	thumb := rl.NewRectangle(trackX+1, thumbY, barW-2, thumbH)
	rl.DrawRectangleRounded(thumb, 1, 6, viewportScrollThumbColor)
}

// drawHorizontalScrollbar renders a horizontal scrollbar track and thumb at the bottom.
func (vp *Viewport) drawHorizontalScrollbar() {
	bounds := vp.Bounds()
	trackY, thumbX, thumbW, _ := vp.horizThumbGeom(bounds)

	track := rl.NewRectangle(bounds.X, trackY, bounds.Width, vp.scrollbarHeight)
	rl.DrawRectangleRounded(track, 1, 6, viewportScrollTrackColor)

	thumb := rl.NewRectangle(thumbX, trackY+1, thumbW, vp.scrollbarHeight-2)
	rl.DrawRectangleRounded(thumb, 1, 6, viewportScrollThumbColor)
}

// vertTrackRect returns the vertical scrollbar track rectangle.
func (vp *Viewport) vertTrackRect(bounds rl.Rectangle) (trackX, trackY, trackW, trackH float32) {
	if vp.styleName == "page-scroll" {
		return vp.pageScrollVertTrackRect(bounds)
	}
	barW := vp.verticalScrollbarWidth()
	trackX = bounds.X + bounds.Width - barW
	trackY = bounds.Y
	trackH = bounds.Height
	if vp.styleName == "preview-scroll" {
		trackY += pageScrollBarVertInset
		trackH -= 2 * pageScrollBarVertInset
		if trackH < 1 {
			trackH = 1
		}
	}
	return trackX, trackY, barW, trackH
}

// vertThumbGeom returns vertical scrollbar track X, thumb Y/H, and max scroll.
func (vp *Viewport) vertThumbGeom(bounds rl.Rectangle) (trackX, thumbY, thumbH, maxScroll float32) {
	trackX, trackY, _, trackH := vp.vertTrackRect(bounds)
	maxScroll = vp.overflowScrollY()
	if vp.contentHeight > 0 {
		thumbH = (trackH / vp.contentHeight) * trackH
	} else {
		thumbH = trackH
	}
	if thumbH < 24 {
		thumbH = 24
	}
	if thumbH > trackH {
		thumbH = trackH
	}
	if maxScroll > 0 {
		thumbY = trackY + (vp.ScrollY/maxScroll)*(trackH-thumbH)
		if thumbY < trackY {
			thumbY = trackY
		}
		if thumbY+thumbH > trackY+trackH {
			thumbY = trackY + trackH - thumbH
		}
	} else {
		thumbY = trackY
	}
	return trackX, thumbY, thumbH, maxScroll
}

// vertScrollbarDrawGeom returns track and thumb geometry for drawing and hit tests.
func (vp *Viewport) vertScrollbarDrawGeom(bounds rl.Rectangle) (trackX, trackY, thumbY, thumbH, trackW, trackH float32) {
	trackX, trackY, trackW, trackH = vp.vertTrackRect(bounds)
	_, thumbY, thumbH, _ = vp.vertThumbGeom(bounds)
	return trackX, trackY, thumbY, thumbH, trackW, trackH
}

// horizThumbGeom returns horizontal track Y, thumb X/W, and max scroll.
func (vp *Viewport) horizThumbGeom(bounds rl.Rectangle) (trackY, thumbX, thumbW, maxScroll float32) {
	trackY = bounds.Y + bounds.Height - vp.scrollbarHeight
	thumbW = (bounds.Width / vp.contentWidth) * bounds.Width
	if thumbW < 24 {
		thumbW = 24
	}
	maxScroll = vp.overflowScrollX()
	if maxScroll > 0 {
		thumbX = bounds.X + (vp.ScrollX/maxScroll)*(bounds.Width-thumbW)
		if thumbX < bounds.X {
			thumbX = bounds.X
		}
		if thumbX+thumbW > bounds.X+bounds.Width {
			thumbX = bounds.X + bounds.Width - thumbW
		}
	} else {
		thumbX = bounds.X
	}
	return trackY, thumbX, thumbW, maxScroll
}

// NodeOffsetYWithin returns the node's Y offset within ancestor's local space.
func NodeOffsetYWithin(node, ancestor Node) float32 {
	if node == nil || ancestor == nil {
		return 0
	}
	return node.Bounds().Y - ancestor.Bounds().Y
}

// measureContentOffsetY relayouts at ScrollY=0 and returns document offset within contentRoot.
func (vp *Viewport) measureContentOffsetY(node, contentRoot Node) float32 {
	if node == nil || contentRoot == nil {
		return 0
	}
	vp.ScrollY = 0
	vp.scrollDirty = true
	for _, ch := range vp.children {
		if !ch.IsHidden() {
			ch.MarkDirty()
		}
	}
	vp.Layout()
	return NodeOffsetYWithin(node, contentRoot)
}

// ScrollToShowNode scrolls the viewport so node is visible within contentRoot.
func (vp *Viewport) ScrollToShowNode(node, contentRoot Node) {
	if vp == nil || node == nil || contentRoot == nil {
		return
	}
	margin := float32(12)
	contentY := vp.measureContentOffsetY(node, contentRoot)
	vp.ScrollY = contentY - margin
	if vp.ScrollY < 0 {
		vp.ScrollY = 0
	}
	vp.clampScrollY()
	vp.scrollDirty = true
	vp.MarkDirty()
}

func (vp *Viewport) clampScrollY() {
	if vp.ScrollY < 0 {
		vp.ScrollY = 0
	}
	if max := vp.overflowScrollY(); vp.ScrollY > max {
		vp.ScrollY = max
	}
}

func (vp *Viewport) clampScrollX() {
	if vp.ScrollX < 0 {
		vp.ScrollX = 0
	}
	if max := vp.overflowScrollX(); vp.ScrollX > max {
		vp.ScrollX = max
	}
}

// handleScrollbarInput supports thumb drag and track click-to-jump on the viewport scrollbar.
func (vp *Viewport) handleScrollbarInput() {
	bounds := vp.Bounds()
	mouse := rl.GetMousePosition()
	inBounds := rl.CheckCollisionPointRec(mouse, bounds)

	if vp.sbDragging {
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
			vp.sbDragging = false
		} else if vp.sbDragHoriz {
			_, _, thumbW, maxScroll := vp.horizThumbGeom(bounds)
			den := bounds.Width - thumbW
			if maxScroll > 0 && den > 0 {
				vp.ScrollX = vp.sbDragStartScroll + (mouse.X-vp.sbDragStartMouse)*(maxScroll/den)
				vp.clampScrollX()
				vp.scrollDirty = true
				vp.MarkDirty()
			}
		} else {
			_, _, thumbH, maxScroll := vp.vertThumbGeom(bounds)
			_, _, _, trackH := vp.vertTrackRect(bounds)
			den := trackH - thumbH
			if maxScroll > 0 && den > 0 {
				vp.ScrollY = vp.sbDragStartScroll + (mouse.Y-vp.sbDragStartMouse)*(maxScroll/den)
				vp.clampScrollY()
				vp.scrollDirty = true
				vp.MarkDirty()
			}
		}
		return
	}

	if !inBounds {
		return
	}

	if vp.Orientation == ScrollHorizontal {
		if vp.overflowScrollX() <= 0 {
			return
		}
		trackY, thumbX, thumbW, maxScroll := vp.horizThumbGeom(bounds)
		track := rl.NewRectangle(bounds.X, trackY, bounds.Width, vp.scrollbarHeight)
		thumbRec := rl.NewRectangle(thumbX+1, trackY+1, thumbW-2, vp.scrollbarHeight-2)
		// Thumb drag: hold on thumb (not only the one-frame Pressed edge — matches pointer latch).
		if rl.IsMouseButtonDown(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, thumbRec) {
			vp.sbDragging = true
			vp.sbDragHoriz = true
			vp.sbDragStartMouse = mouse.X
			vp.sbDragStartScroll = vp.ScrollX
			PointerClickMarkUsed()
			return
		}
		if rl.CheckCollisionPointRec(mouse, track) &&
			(PointerClickConsume(track) || rl.IsMouseButtonPressed(rl.MouseLeftButton)) {
			if maxScroll > 0 && bounds.Width > 1 {
				r := (mouse.X - bounds.X) / bounds.Width
				if r < 0 {
					r = 0
				}
				if r > 1 {
					r = 1
				}
				vp.ScrollX = r * maxScroll
				vp.clampScrollX()
				vp.scrollDirty = true
				vp.MarkDirty()
				PointerClickMarkUsed()
			}
		}
		return
	}

	if vp.overflowScrollY() <= 0 {
		return
	}
	trackX, trackY, thumbY, thumbH, trackW, trackH := vp.vertScrollbarDrawGeom(bounds)
	track := rl.NewRectangle(trackX, trackY, trackW, trackH)
	thumbRec := rl.NewRectangle(trackX+1, thumbY, trackW-2, thumbH)
	if rl.IsMouseButtonDown(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, thumbRec) {
		vp.sbDragging = true
		vp.sbDragHoriz = false
		vp.sbDragStartMouse = mouse.Y
		vp.sbDragStartScroll = vp.ScrollY
		PointerClickMarkUsed()
		return
	}
	if rl.CheckCollisionPointRec(mouse, track) &&
		(PointerClickConsume(track) || rl.IsMouseButtonPressed(rl.MouseLeftButton)) {
		if maxScroll := vp.overflowScrollY(); maxScroll > 0 && trackH > 1 {
			r := (mouse.Y - trackY) / trackH
			if r < 0 {
				r = 0
			}
			if r > 1 {
				r = 1
			}
			vp.ScrollY = r * maxScroll
			vp.clampScrollY()
			vp.scrollDirty = true
			vp.MarkDirty()
			PointerClickMarkUsed()
		}
	}
}
