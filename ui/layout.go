// Package ui (continued) — flex/grid layout, hit-testing, and resize helpers.
//
// See node.go for the full package documentation.
//
// # Layout contract (summary)
//
// Parents assign child bounds with layoutSetBounds during Layout(); never call
// child.SetBounds from inside a parent's Layout pass — that bubbles MarkDirty
// to the root and breaks idle FPS (see docs/IDLE_INVARIANTS.md).
//
// Every widget that overrides Layout must set layoutDirty = false before return,
// including leaf widgets with no child arrangement.
//
// # LLM Prompt Template — scrollable page column
//
//	vp := ui.NewViewport("page-scroll", 0, 0, 0, 0)
//	vp.SetStyle("page-scroll")
//	vp.FlexDirection = ui.FlexColumn
//	vp.SetFlexGrow(1)
//	body := ui.NewContainer("body", 0, 0, 0, 0)
//	body.LayoutType = ui.LayoutFlex
//	body.FlexDirection = ui.FlexColumn
//	body.AutoHeight = true
//	body.Gap = 8
//	vp.AddChild(body)
//	root.AddChild(vp)
//
// Window resize: call doc.Resize(w, h) or doc.RelayoutTreeForResize(); do not
// only root.SetBounds — flex children need MarkResizeLayoutDirtySubtree.
package ui

import (
	"reflect"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// panelChromeBleed is the default inset expansion for Panel/Card body layout and
// draw scissors so rounded strokes are not sheared at the clip edge.
const panelChromeBleed = float32(4)

// LayoutGridUnit is the 4px spacing rhythm (doc_tokens "xs"). Use snapLayoutUnit and
// snapLayoutRect for dense toolbars and control strips so positions match page grids.
const LayoutGridUnit = float32(4)

// snapLayoutUnit rounds v to the nearest LayoutGridUnit.
func snapLayoutUnit(v float32) float32 {
	if LayoutGridUnit <= 0 {
		return v
	}
	return float32(int32((v/LayoutGridUnit)+0.5)) * LayoutGridUnit
}

// snapLayoutRect snaps position and size to the layout grid (minimum 1 unit per axis).
func snapLayoutRect(r rl.Rectangle) rl.Rectangle {
	w := snapLayoutUnit(r.Width)
	h := snapLayoutUnit(r.Height)
	if w < LayoutGridUnit {
		w = LayoutGridUnit
	}
	if h < LayoutGridUnit {
		h = LayoutGridUnit
	}
	return rl.NewRectangle(snapLayoutUnit(r.X), snapLayoutUnit(r.Y), w, h)
}

// panelChipChromeBleed is used when direct children include pill badges or buttons
// (raylib rounded meshes can extend ~6px past layout bounds).
const panelChipChromeBleed = float32(6)

// layoutSetBounds assigns child bounds during a parent Layout pass without
// bubbling MarkDirty upward. SetBounds during layout would keep the root dirty
// every frame and block idle FPS (SplitView, ResizablePanel, grid flex).
//
// # LLM Prompt Template
//
//	// Inside parent Layout() — never child.SetBounds (bubbles MarkDirty to root):
//	layoutSetBounds(child, rl.NewRectangle(x, y, w, h))
//	if child.IsDirty() {
//	    child.Layout()
//	}
func layoutSetBounds(n Node, r rl.Rectangle) {
	if n == nil {
		return
	}
	cur := n.Bounds()
	if cur.X == r.X && cur.Y == r.Y && cur.Width == r.Width && cur.Height == r.Height {
		return
	}
	if bw, ok := n.(interface{ setBoundsNoMark(rl.Rectangle) }); ok {
		bw.setBoundsNoMark(r)
	} else {
		n.SetBounds(r)
	}
}

// layoutChildAfterSetBounds assigns pane bounds and runs Layout when bounds
// changed or the child was already dirty. Never calls MarkDirty on the child —
// avoids pinning the document root dirty every frame (Notepad notes + preview splits).
func layoutChildAfterSetBounds(n Node, r rl.Rectangle) {
	if n == nil {
		return
	}
	before := n.Bounds()
	layoutSetBounds(n, r)
	if before != n.Bounds() {
		scheduleLayoutPass(n)
		markDrawDirtyNode(n)
	}
	if n.IsDirty() {
		n.Layout()
	}
}

// markDrawDirtyNode invalidates SSAA / render caches when layout assigns new bounds
// without SetBounds (setBoundsNoMark). Bubbles upward only — no layoutDirty.
func markDrawDirtyNode(n Node) {
	if d, ok := n.(interface{ MarkDrawDirty() }); ok {
		d.MarkDrawDirty()
	}
}

// scheduleLayoutPass marks layoutDirty on n without bubbling to ancestors.
func scheduleLayoutPass(n Node) {
	switch v := n.(type) {
	case *Container:
		v.layoutDirty = true
	case *SplitView:
		v.layoutDirty = true
	case *Viewport:
		v.layoutDirty = true
	case *ResizablePanel:
		v.layoutDirty = true
	default:
		if el, ok := n.(interface{ scheduleLayoutPassLocal() }); ok {
			el.scheduleLayoutPassLocal()
		}
	}
}

func (e *Element) scheduleLayoutPassLocal() { e.layoutDirty = true }

// panelBodyChromeBleed picks bleed for layout clamp and body scissor.
func panelBodyChromeBleed(children []Node) float32 {
	bleed := panelChromeBleed
	for _, ch := range children {
		if nodeUsesRoundedChipDraw(ch) {
			return panelChipChromeBleed
		}
	}
	return bleed
}

// nodeUsesRoundedChipDraw reports whether n or a shallow descendant draws rounded chips.
func nodeUsesRoundedChipDraw(n Node) bool {
	switch v := n.(type) {
	case *Badge:
		return v.Shape != BadgeShapeRect
	case *Button:
		return true
	case *Container:
		for _, ch := range v.children {
			if nodeUsesRoundedChipDraw(ch) {
				return true
			}
		}
	case *Card:
		for _, ch := range v.Children() {
			if nodeUsesRoundedChipDraw(ch) {
				return true
			}
		}
	}
	return false
}

func panelBodyContentExtents(bodyRect rl.Rectangle, padding, bleed float32) (minX, maxX, minY, maxY float32) {
	minX = bodyRect.X + padding - bleed
	maxX = bodyRect.X + bodyRect.Width - padding + bleed
	minY = bodyRect.Y + padding - bleed
	maxY = bodyRect.Y + bodyRect.Height - padding + bleed
	if minX < bodyRect.X {
		minX = bodyRect.X
	}
	if maxX > bodyRect.X+bodyRect.Width {
		maxX = bodyRect.X + bodyRect.Width
	}
	if minY < bodyRect.Y {
		minY = bodyRect.Y
	}
	if maxY > bodyRect.Y+bodyRect.Height {
		maxY = bodyRect.Y + bodyRect.Height
	}
	if maxX < minX {
		maxX = minX
	}
	if maxY < minY {
		maxY = minY
	}
	return minX, maxX, minY, maxY
}

// flexChildFillCrossWidth reports whether a vertical stack child should take the
// full available inner width. Cards and flex columns only shrank children when
// b.Width > availW, so after a narrow window pass RichText/viewports stayed
// narrow when the parent grew again.
func flexChildFillCrossWidth(ch Node, availW, childW float32) bool {
	if availW <= 0 {
		return false
	}
	if nodeHasFixedCrossWidth(ch) {
		return false
	}
	if childW == 0 || childW > availW {
		return true
	}
	if childW >= availW-0.5 {
		return false
	}
	switch ch.(type) {
	case *RichText, *Viewport, *Container, *Panel, *Card:
		return true
	case *Toolbar, *MenuBar, *StatusBar:
		// Desktop chrome bars must track parent width after resize; they are not
		// AutoHeight but still fill the cross-axis (notepad ribbon, menubar strip).
		return true
	}
	return ch.IsAutoHeight() && ch.GetFlexGrow() == 0
}

// PinnedMainAxisWidth returns the pinned width when Min/Max/Preferred all match.
func PinnedMainAxisWidth(ch Node) (float32, bool) {
	if !nodeHasPinnedWidth(ch) {
		return 0, false
	}
	type widthHinter interface {
		GetPreferredWidth() float32
	}
	wh, ok := ch.(widthHinter)
	if !ok {
		return 0, false
	}
	return wh.GetPreferredWidth(), true
}

// flexChildPreferredWidth returns a non-zero natural width when the child
// implements GetPreferredWidth (e.g. Button shrink-wrap in flex rows/columns).
func flexChildPreferredWidth(ch Node) (float32, bool) {
	type widthHinter interface {
		GetPreferredWidth() float32
	}
	wh, ok := ch.(widthHinter)
	if !ok {
		return 0, false
	}
	pw := wh.GetPreferredWidth()
	if pw <= 0 {
		return 0, false
	}
	return pw, true
}

// nodeHasPinnedWidth is true when PreferredWidth, MinWidth, and MaxWidth all pin the
// same main-axis size (navigation rail, preview lanes at simWidth, etc.).
func nodeHasPinnedWidth(ch Node) bool {
	type widthHinter interface {
		GetPreferredWidth() float32
		GetMinWidth() float32
		GetMaxWidth() float32
	}
	wh, ok := ch.(widthHinter)
	if !ok {
		return false
	}
	pw, mn, mx := wh.GetPreferredWidth(), wh.GetMinWidth(), wh.GetMaxWidth()
	if pw <= 0 || mx <= 0 {
		return false
	}
	if mx > pw+0.5 {
		return false
	}
	if mn > 0 && (mn < pw-0.5 || mn > pw+0.5) {
		return false
	}
	return true
}

// nodeHasFixedCrossWidth is true when PreferredWidth and MaxWidth pin the same width
// (preview lanes at simWidth, horizontal viewports capped to layout width).
func nodeHasFixedCrossWidth(ch Node) bool {
	return nodeHasPinnedWidth(ch)
}

// nodeLayoutBottomExtent returns the bottom Y used when sizing an AutoHeight
// Panel/Card body. Nested Cards reserve RaisedSurfaceShadowBleed below their
// layout bounds for the shared drop shadow.
func nodeLayoutBottomExtent(ch Node) float32 {
	b := ch.Bounds()
	bottom := b.Y + b.Height
	if _, ok := ch.(*Card); ok {
		bottom += RaisedSurfaceShadowBleed
	}
	return bottom
}

// nodeSubtreeBottom returns the lowest layout Y of n and visible descendants.
// Used after child Layout() so AutoHeight hosts include nested grids/lists.
func nodeSubtreeBottom(n Node) float32 {
	if n.IsHidden() {
		return 0
	}
	if acc, ok := n.(*Accordion); ok {
		b := acc.Bounds()
		// Always use shell bounds (title + body band). Walking expanded children
		// can disagree with animH by body padding and leave sibling buttons unflush.
		return b.Y + b.Height
	}
	// Scroll/split hosts report outer bounds only; inner overflow scrolls inside.
	if _, ok := n.(*Viewport); ok {
		return nodeLayoutBottomExtent(n)
	}
	// Toolbar clips item widgets; descendants must not inflate AutoHeight panels.
	if _, ok := n.(*Toolbar); ok {
		return nodeLayoutBottomExtent(n)
	}
	if _, ok := n.(*SplitView); ok {
		return nodeLayoutBottomExtent(n)
	}
	if _, ok := n.(*ResizablePanel); ok {
		return nodeLayoutBottomExtent(n)
	}
	bottom := nodeLayoutBottomExtent(n)
	type childLister interface {
		Children() []Node
	}
	cl, ok := n.(childLister)
	if !ok {
		return bottom
	}
	for _, ch := range cl.Children() {
		if ch.IsHidden() {
			continue
		}
		if sub := nodeSubtreeBottom(ch); sub > bottom {
			bottom = sub
		}
	}
	return bottom
}

// subtreeHasFlexGrowFill is true when any descendant uses flex-grow to fill a
// parent band (e.g. SplitView in a grid-stretched panel).
func subtreeHasFlexGrowFill(nodes []Node) bool {
	for _, ch := range nodes {
		if ch.IsHidden() {
			continue
		}
		if ch.GetFlexGrow() > 0 {
			return true
		}
		if kids := ch.Children(); len(kids) > 0 {
			if subtreeHasFlexGrowFill(kids) {
				return true
			}
		}
	}
	return false
}

// panelBodyFillHeight picks intrinsic vs assigned band per docs/LAYOUT_CONTRACTS.md §3.
func panelBodyFillHeight(host Node, origH, intrinsicH float32, hostFlexGrow float32, children []Node) float32 {
	if origH >= flexIntrinsicProbeH {
		return intrinsicH
	}
	if hostFlexGrow > 0 && origH > 0 {
		return origH
	}
	if !host.IsAutoHeight() && origH > 0 {
		return origH
	}
	if subtreeHasFlexGrowFill(children) && origH > 0 {
		return origH
	}
	if origH > 0 && origH < intrinsicH-0.5 && nestedInFixedHeightSurface(host) {
		return origH
	}
	return intrinsicH
}

// nestedInFixedHeightSurface is true when a host sits inside a non-AutoHeight panel/card body.
func nestedInFixedHeightSurface(n Node) bool {
	for p := n.ParentNode(); p != nil; p = p.ParentNode() {
		switch parent := p.(type) {
		case *Panel:
			if !parent.AutoHeight {
				return true
			}
		case *Card:
			if !parent.AutoHeight {
				return true
			}
		case *SurfaceShell:
			if !parent.AutoHeight {
				return true
			}
		}
	}
	return false
}

// allowAutoHeightSubtreeFinalize is false when a fixed-height Panel/Card parent
// already clamped this host — growing here would spill past the parent body.
func allowAutoHeightSubtreeFinalize(host Node) bool {
	p := host.ParentNode()
	for p != nil {
		switch parent := p.(type) {
		case *Panel:
			if !parent.AutoHeight || parent.GetFlexGrow() > 0 {
				return false
			}
		case *Card:
			if !parent.AutoHeight || parent.GetFlexGrow() > 0 {
				return false
			}
		}
		p = p.ParentNode()
	}
	return true
}

// finalizeAutoHeightFromSubtree syncs an AutoHeight Panel/Card body after
// descendants have run Layout(). layoutContent sizes from direct-child bounds
// only, which misses nested grid rows and leaves stale height when content shrinks.
func finalizeAutoHeightFromSubtree(
	host Node,
	bounds *rl.Rectangle,
	titleOff, padding float32,
	children []Node,
	flexGrow float32,
	layoutFlexInRect func(rect rl.Rectangle),
	layoutLabels func(),
	clampChildren func(bodyRect rl.Rectangle, padding float32),
) {
	if len(children) == 0 || !allowAutoHeightSubtreeFinalize(host) {
		return
	}
	origY := bounds.Y
	origH := bounds.Height
	bodyY := origY + titleOff

	measureIntrinsic := func() float32 {
		contentEnd := bodyY
		for _, ch := range children {
			if ch.IsHidden() {
				continue
			}
			if bottom := nodeSubtreeBottom(ch); bottom > contentEnd {
				contentEnd = bottom
			}
		}
		contentEnd += padding
		bodyUsed := contentEnd - bodyY
		if bodyUsed < 0 {
			bodyUsed = 0
		}
		return titleOff + bodyUsed
	}

	syncBody := func(finalH float32) {
		bounds.Y = origY
		bounds.Height = finalH
		finalBodyH := finalH - titleOff
		if finalBodyH < 0 {
			finalBodyH = 0
		}
		bodyRect := rl.NewRectangle(bounds.X, bodyY, bounds.Width, finalBodyH)
		layoutFlexInRect(bodyRect)
		layoutLabels()
		// Nested LayoutGrid / flex must run at the final body width (preview chrome).
		for _, ch := range children {
			if !ch.IsHidden() {
				ch.Layout()
			}
		}
	}

	intrinsic := measureIntrinsic()
	if flexGrow > 0 && origH > intrinsic+0.5 {
		return
	}
	if intrinsic >= origH-0.5 && intrinsic <= origH+0.5 {
		return
	}
	syncBody(intrinsic)
	// Second measure: grid reflow may change row count after width/height sync.
	if intrinsic2 := measureIntrinsic(); intrinsic2 > intrinsic+0.5 {
		intrinsic = intrinsic2
		syncBody(intrinsic)
	}
}

// panelBodyDrawClipRect is the padded body scissor for Panel/Card child draws.
// When a direct child is a Card, the clip extends to the body bottom so nested
// shadows are not sheared.
func panelBodyDrawClipRect(bodyRect rl.Rectangle, padding float32, children []Node) rl.Rectangle {
	minX, maxX, minY, maxY := panelBodyContentExtents(bodyRect, padding, panelBodyChromeBleed(children))
	bodyRight := bodyRect.X + bodyRect.Width
	bodyBottom := bodyRect.Y + bodyRect.Height
	shadowBleed := false
	for _, ch := range children {
		if ch.IsHidden() {
			continue
		}
		if card, ok := ch.(*Card); ok {
			st := card.GetStyle()
			bleed := st.CornerRadius
			if bleed < 6 {
				bleed = 6
			}
			if bw := st.BorderWidth; bw > 0 && bleed < bw+4 {
				bleed = bw + 4
			}
			if minX-bleed > bodyRect.X {
				minX -= bleed
			} else {
				minX = bodyRect.X
			}
			if minY-bleed > bodyRect.Y {
				minY -= bleed
			} else {
				minY = bodyRect.Y
			}
			if maxX+bleed < bodyRight {
				maxX += bleed
			} else {
				maxX = bodyRight
			}
			if !shadowBleed {
				maxY += RaisedSurfaceShadowBleed
				if maxY > bodyBottom {
					maxY = bodyBottom
				}
				shadowBleed = true
			}
			if maxY+bleed < bodyBottom {
				maxY += bleed
			} else if maxY < bodyBottom {
				maxY = bodyBottom
			}
		}
		if g, ok := ch.(interface{ ChromeGlowIntensity() float32 }); ok {
			if bleed := g.ChromeGlowIntensity() * SurfaceGlowBleed; bleed > 0 {
				if minX-bleed > bodyRect.X {
					minX -= bleed
				} else {
					minX = bodyRect.X
				}
				if minY-bleed > bodyRect.Y {
					minY -= bleed
				} else {
					minY = bodyRect.Y
				}
				if maxX+bleed < bodyRight {
					maxX += bleed
				} else {
					maxX = bodyRight
				}
				if maxY+bleed < bodyBottom {
					maxY += bleed
				} else {
					maxY = bodyBottom
				}
			}
		}
	}
	return rl.NewRectangle(minX, minY, maxX-minX, maxY-minY)
}

// containerContentDrawClip is the scissor for ClipChildren containers. When
// direct children include Cards, the clip extends slightly on the bottom
// so chromeBorderBounds strokes are not sheared at the content edge.
func containerContentDrawClip(bounds rl.Rectangle, padding float32, children []Node) rl.Rectangle {
	minX := bounds.X + padding
	minY := bounds.Y + padding
	maxX := bounds.X + bounds.Width - padding
	maxY := bounds.Y + bounds.Height - padding
	hasCard := false
	for _, ch := range children {
		if ch.IsHidden() {
			continue
		}
		if _, ok := ch.(*Card); ok {
			hasCard = true
			break
		}
	}
	if hasCard {
		// Bottom bleed only — extending maxX made preview lanes look padded on the right.
		bleed := panelChromeBleed
		maxY += bleed
		bBottom := bounds.Y + bounds.Height
		if maxY > bBottom {
			maxY = bBottom
		}
	}
	w := maxX - minX
	h := maxY - minY
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return rl.NewRectangle(minX, minY, w, h)
}

// surfaceBodyContentClip returns the padded body content band of the nearest
// Panel/Card ancestors. RichText/Label use this at draw time so scissor rects
// stay inside chrome even when AutoHeight layout over-estimates height.
func surfaceBodyContentClip(n Node) (rl.Rectangle, bool) {
	var clip rl.Rectangle
	has := false
	for cur := n.ParentNode(); cur != nil; cur = cur.ParentNode() {
		var band rl.Rectangle
		ok := false
		switch p := cur.(type) {
		case *Panel:
			band, ok = panelBodyContentClip(p)
		case *Card:
			band, ok = cardBodyContentClip(p)
		}
		if !ok {
			continue
		}
		if !has {
			clip = band
			has = true
		} else {
			clip = intersectRects(clip, band)
		}
	}
	return clip, has
}

func panelBodyContentClip(p *Panel) (rl.Rectangle, bool) {
	style := p.GetStyle()
	titleH := p.bodyTitleHeight()
	b := p.Bounds()
	if b.Width < 1 || b.Height < titleH+1 {
		return rl.Rectangle{}, false
	}
	body := rl.NewRectangle(b.X, b.Y+titleH, b.Width, b.Height-titleH)
	minX, maxX, minY, maxY := panelBodyContentExtents(body, style.Padding, panelBodyChromeBleed(p.Children()))
	w := maxX - minX
	h := maxY - minY
	if w < 1 || h < 1 {
		return rl.Rectangle{}, false
	}
	return rl.NewRectangle(minX, minY, w, h), true
}

func cardBodyContentClip(c *Card) (rl.Rectangle, bool) {
	style := c.GetStyle()
	titleH := c.bodyTitleHeight()
	b := c.Bounds()
	if b.Width < 1 || b.Height < titleH+1 {
		return rl.Rectangle{}, false
	}
	body := rl.NewRectangle(b.X, b.Y+titleH, b.Width, b.Height-titleH)
	minX, maxX, minY, maxY := panelBodyContentExtents(body, style.Padding, panelBodyChromeBleed(c.Children()))
	w := maxX - minX
	h := maxY - minY
	if w < 1 || h < 1 {
		return rl.Rectangle{}, false
	}
	return rl.NewRectangle(minX, minY, w, h), true
}

// clampChildrenToBodyRect clamps each child's bounds to the padded body band.
// Used by Panel and Card after layout so controls respect chrome insets the same
// way in layout and draw.
func clampChildrenToBodyRect(children []Node, bodyRect rl.Rectangle, padding float32) {
	minX, maxX, minY, maxY := panelBodyContentExtents(bodyRect, padding, panelBodyChromeBleed(children))
	bodyBottom := bodyRect.Y + bodyRect.Height
	for _, ch := range children {
		if ch.IsHidden() {
			continue
		}
		b := ch.Bounds()
		origW := b.Width
		x2 := b.X + b.Width
		y2 := b.Y + b.Height
		childMaxY := maxY
		if _, nested := ch.(*Card); nested {
			childMaxY = maxY + RaisedSurfaceShadowBleed
			if childMaxY > bodyBottom {
				childMaxY = bodyBottom
			}
		}
		if b.X < minX {
			b.X = minX
		}
		if b.Y < minY {
			b.Y = minY
		}
		if x2 > maxX {
			x2 = maxX
		}
		heightClamped := false
		if y2 > childMaxY {
			y2 = childMaxY
			heightClamped = true
		}
		b.Width = x2 - b.X
		b.Height = y2 - b.Y
		if b.Width < 0 {
			b.Width = 0
		}
		if b.Height < 0 {
			b.Height = 0
		}
		widthClamped := origW > 0 && b.Width+0.5 < origW
		layoutSetBounds(ch, b)
		if widthClamped || heightClamped {
			ch.MarkDirty()
			ch.Layout()
			if heightClamped {
				relayoutClampedSurfaceChild(ch)
			}
		}
	}
}

// relayoutClampedSurfaceChild reflows a panel/card child after its height was
// capped by clampChildrenToBodyRect so nested AutoHeight content respects the band.
func relayoutClampedSurfaceChild(ch Node) {
	ch.Layout()
	switch v := ch.(type) {
	case *Card:
		style := v.GetStyle()
		titleOff := bodyTitleHeightFor(v)
		b := v.Bounds()
		bodyH := b.Height - titleOff
		if bodyH < 0 {
			bodyH = 0
		}
		bodyRect := rl.NewRectangle(b.X, b.Y+titleOff, b.Width, bodyH)
		layoutBodyLabels(v.Children(), true)
		clampChildrenToBodyRect(v.Children(), bodyRect, style.Padding)
	case *Panel:
		style := v.GetStyle()
		titleOff := bodyTitleHeightFor(v)
		b := v.Bounds()
		bodyH := b.Height - titleOff
		if bodyH < 0 {
			bodyH = 0
		}
		bodyRect := rl.NewRectangle(b.X, b.Y+titleOff, b.Width, bodyH)
		layoutBodyLabels(v.Children(), true)
		clampChildrenToBodyRect(v.Children(), bodyRect, style.Padding)
	case *Accordion:
		v.Layout()
	case *SplitView:
		v.MarkDirty()
		v.Layout()
		finalizeSplitPaneContainment(v.first, v.firstRect())
		finalizeSplitPaneContainment(v.second, v.secondRect())
	case *ResizablePanel:
		v.MarkDirty()
		v.Layout()
	}
}

// Size represents a 2D size with width and height.
type Size struct {
	Width  float32
	Height float32
}

// NewSize creates a new Size.
func NewSize(w, h float32) Size {
	return Size{Width: w, Height: h}
}

// findViewport walks the parent chain of n and returns the first *Viewport
// ancestor, or nil if the node is not inside a Viewport.
// Use this in widgets that need to intersect their scissor rect with the
// enclosing Viewport's ClipBounds() (e.g. TextInput, custom leaf widgets).
func findViewport(n Node) *Viewport {
	p := n.ParentNode()
	for p != nil {
		if vp, ok := p.(*Viewport); ok {
			return vp
		}
		p = p.ParentNode()
	}
	return nil
}

// hitTestChildOrder returns children sorted by ZIndex high → low so hit-testing
// matches paint order (topmost sibling wins). See ARCHITECTURE.md §5.6 / §9.
func hitTestChildOrder(children []Node) []Node {
	if len(children) <= 1 {
		return children
	}
	sorted := make([]Node, len(children))
	copy(sorted, children)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].GetZIndex() > sorted[j].GetZIndex()
	})
	return sorted
}

// FindNodeByID walks root depth-first and returns the first node with the given ID.
//
// Example (focus a field after navigation):
//
//	if ti, ok := ui.FindNodeByID(doc.Root, "email").(*ui.TextInput); ok {
//	    doc.SetFocus(ti)
//	}
func FindNodeByID(root Node, id string) Node {
	if root == nil || id == "" {
		return nil
	}
	if root.ID() == id {
		return root
	}
	for _, child := range root.Children() {
		if found := FindNodeByID(child, id); found != nil {
			return found
		}
	}
	return nil
}

// FindInteractiveAt walks root and returns the deepest interactive node under point p.
// Siblings are tested in ZIndex order (highest first), matching draw order.
//
// Example (scene OnUpdate click-to-focus):
//
//	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
//	    if hit := ui.FindInteractiveAt(doc.Root, rl.GetMousePosition()); hit != nil {
//	        doc.SetFocus(hit)
//	    }
//	}
func FindInteractiveAt(root Node, p rl.Vector2) Node {
	if root.IsHidden() {
		return nil
	}
	for _, child := range hitTestChildOrder(root.Children()) {
		if child.IsHidden() {
			continue
		}
		if hit := FindInteractiveAt(child, p); hit != nil {
			return hit
		}
		if child.IsInteractive() && rl.CheckCollisionPointRec(p, child.Bounds()) {
			return child
		}
	}
	return nil
}

// drawChildrenInRectClip draws children inside clip, intersected with ancestor
// viewports, and re-applies the scissor after each child. Label Wrap/Truncate
// and other widgets call EndScissorMode and would otherwise break a single
// Begin/End pair on ClipChildren containers (grid lanes, etc.).
func drawChildrenInRectClip(parent Node, clip rl.Rectangle, children []Node) {
	clip = intersectRectsWithViewportAncestors(clip, parent)
	if clip.Width < 1 || clip.Height < 1 {
		drawChildrenWithScissorRestore(parent, children)
		return
	}
	beginScissorFromRect(clip)
	setActiveDrawClip(clip)
	for _, ch := range children {
		if ch.IsHidden() {
			continue
		}
		ch.Draw()
		beginScissorFromRect(clip)
		setActiveDrawClip(clip)
	}
	clearActiveDrawClip()
	rl.EndScissorMode()
}

// withClipRestore draws inside clip while preserving an ancestor scissor (e.g. panel
// body or page viewport). Widgets such as Toolbar must use this instead of a bare
// EndScissorMode so scrolling does not leak content past the parent clip.
func withClipRestore(parent Node, clip rl.Rectangle, draw func()) {
	clip = intersectRectsWithViewportAncestors(clip, parent)
	parentClip, hadParent := ancestorClipBounds(parent)
	if clip.Width < 1 || clip.Height < 1 {
		draw()
		return
	}
	beginScissorFromRect(clip)
	setActiveDrawClip(clip)
	draw()
	clearActiveDrawClip()
	if hadParent && parentClip.Width >= 1 && parentClip.Height >= 1 {
		parentClip = intersectRectsWithViewportAncestors(parentClip, parent)
		if parentClip.Width >= 1 && parentClip.Height >= 1 {
			beginScissorFromRect(parentClip)
			setActiveDrawClip(parentClip)
			return
		}
	}
	rl.EndScissorMode()
}

// drawChildrenInPaddedBodyClip draws children inside bodyClip (panel/card padded
// body), intersected with every ancestor Viewport.ClipBounds(), and re-applies
// the body scissor after each child so scissor-using widgets cannot leak past
// the chrome.
func drawChildrenInPaddedBodyClip(parent Node, bodyClip rl.Rectangle, children []Node) {
	bodyClip = intersectRectsWithViewportAncestors(bodyClip, parent)
	if bodyClip.Width < 1 || bodyClip.Height < 1 {
		drawChildrenWithScissorRestore(parent, children)
		return
	}
	drawChildrenInRectClip(parent, bodyClip, children)
}

// drawChildrenWithScissorRestore draws each node in children and, after every
// draw call, re-applies the active ancestor scissor rectangle if there is one.
// Widgets such as TextInput, Slider, ProgressBar, VirtualList, and Label (when
// Truncate=true) call EndScissorMode on exit, which destroys whatever scissor
// was active. Accordion's animated body clip has the same issue for nested
// Containers, so the restore handles both Viewport and Accordion ancestors.
//
// The function is used by Container.drawInternal() when ClipChildren is false.
// Panel and Card prefer drawChildrenInPaddedBodyClip for strict body clipping.
func drawChildrenWithScissorRestore(parent Node, children []Node) {
	clip, hasClip := ancestorClipBounds(parent)
	if !hasClip {
		// Not inside a clipping ancestor — just draw normally.
		for _, child := range children {
			child.Draw()
		}
		return
	}
	if clip.Width < 1 || clip.Height < 1 {
		return
	}
	cx, cy := int32(clip.X), int32(clip.Y)
	cw, ch := int32(clip.Width), int32(clip.Height)
	for _, child := range children {
		child.Draw()
		// Restore the scissor so the next sibling cannot escape the clipping
		// boundary even if this child clobbered it.
		beginScissorMode(cx, cy, cw, ch)
	}
}

func absF(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}

// intersectRects returns the intersection of two rectangles.
// If the rectangles do not overlap the returned width or height will be ≤ 0;
// callers should check for that before passing the result to BeginScissorMode.
func intersectRects(a, b rl.Rectangle) rl.Rectangle {
	x1 := a.X
	if b.X > x1 {
		x1 = b.X
	}
	y1 := a.Y
	if b.Y > y1 {
		y1 = b.Y
	}
	x2 := a.X + a.Width
	if b.X+b.Width < x2 {
		x2 = b.X + b.Width
	}
	y2 := a.Y + a.Height
	if b.Y+b.Height < y2 {
		y2 = b.Y + b.Height
	}
	return rl.NewRectangle(x1, y1, x2-x1, y2-y1)
}

// intersectRectsWithViewportAncestors returns r intersected with every
// Viewport.ClipBounds() along start's parent chain (outermost viewport first
// as we walk upward). Used by Panel.Draw so body clipping matches nested
// scroll viewports the same way Viewport.Draw clips its children.
func intersectRectsWithViewportAncestors(r rl.Rectangle, start Node) rl.Rectangle {
	if start == nil {
		return r
	}
	out := r
	for cur := start.ParentNode(); cur != nil; cur = cur.ParentNode() {
		if vp, ok := cur.(*Viewport); ok {
			out = intersectRects(out, vp.ClipBounds())
		}
	}
	return out
}

// ancestorClipBounds returns the active clip imposed by ancestors that own
// clipping state. This includes Viewport content clips and Accordion's animated
// body clip, which matters for nested Containers inside collapsing accordions.
func ancestorClipBounds(start Node) (rl.Rectangle, bool) {
	var out rl.Rectangle
	hasClip := false
	for cur := start.ParentNode(); cur != nil; cur = cur.ParentNode() {
		var clip rl.Rectangle
		ownsClip := false
		switch n := cur.(type) {
		case *Viewport:
			clip = n.ClipBounds()
			ownsClip = true
		case *Accordion:
			if n.animH <= 0 {
				return rl.NewRectangle(0, 0, 0, 0), true
			}
			b := n.Bounds()
			clip = rl.NewRectangle(b.X, b.Y+n.TitleHeight, b.Width, n.animH)
			ownsClip = true
		case *Toolbar:
			clip = n.itemBandRect()
			ownsClip = clip.Width >= 1 && clip.Height >= 1
		}
		if !ownsClip {
			continue
		}
		if !hasClip {
			out = clip
			hasClip = true
		} else {
			out = intersectRects(out, clip)
		}
	}
	return out, hasClip
}

// viewportScrollRecoveryFrames forces full layoutFlex in every Viewport for N frames
// after minimize/restore (see ArmViewportScrollRecovery in main; default 2).
var viewportScrollRecoveryFrames int

// ArmViewportScrollRecovery sets how many upcoming frames skip Viewport scroll fast-path.
func ArmViewportScrollRecovery(frames int) {
	if frames > viewportScrollRecoveryFrames {
		viewportScrollRecoveryFrames = frames
	}
}

// ViewportScrollRecoveryActive reports whether scroll recovery layout is in progress.
func ViewportScrollRecoveryActive() bool {
	return viewportScrollRecoveryFrames > 0
}

// TickViewportScrollRecovery advances the recovery counter (call once per frame after Layout).
func TickViewportScrollRecovery() {
	if viewportScrollRecoveryFrames > 0 {
		viewportScrollRecoveryFrames--
	}
}

func viewportScrollRecovering() bool {
	return viewportScrollRecoveryFrames > 0
}

// InvalidateAutoHeightTextMeasures clears intrinsic-height caches on Labels and
// RichText so the next layout pass remeasures at the flex-assigned cell width.
func InvalidateAutoHeightTextMeasures(root Node) {
	var walk func(Node)
	walk = func(n Node) {
		if n == nil || n.IsHidden() {
			return
		}
		if l, ok := n.(*Label); ok && l.IsAutoHeight() {
			l.measure.reset()
			l.MarkDirty()
		}
		if rt, ok := n.(*RichText); ok && rt.IsAutoHeight() {
			rt.InvalidateAutoHeightMeasure()
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
}

// MarkAutoHeightHostDirty marks viewport/flex ancestors dirty when an AutoHeight
// node's intrinsic height changed (async math texture, wrapped PlainText, etc.).
func MarkAutoHeightHostDirty(n Node) {
	markAutoHeightLayoutHostDirty(n)
}

// markAutoHeightLayoutHostDirty marks viewport ancestors dirty when an AutoHeight
// text block changes height so flex columns restack siblings (fixes overlap until
// resize on .gru pages, directory status lines, and wrapped card copy).
func markAutoHeightLayoutHostDirty(n Node) {
	for p := n.ParentNode(); p != nil; p = p.ParentNode() {
		p.MarkDirty()
		if vp, ok := p.(*Viewport); ok {
			vp.lastFlexValid = false
			break
		}
	}
}

// InvalidateViewportScrollFastPath walks the tree and clears Viewport scroll
// fast-path state so the next Layout runs full layoutFlex (used after minimize/
// restore when scroll can feel stuck for a few frames).
func InvalidateViewportScrollFastPath(root Node) {
	var walk func(Node)
	walk = func(n Node) {
		if n == nil || n.IsHidden() {
			return
		}
		if vp, ok := n.(*Viewport); ok {
			vp.scrollDirty = false
			vp.lastFlexValid = false
			vp.lastCommittedLayoutValid = false
			vp.MarkDirty()
			maxScroll := vp.overflowScrollY()
			if vp.ScrollY > maxScroll {
				vp.ScrollY = maxScroll
			}
			if vp.ScrollY < 0 {
				vp.ScrollY = 0
			}
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
}

// MarkResizeLayoutDirtySubtree marks root and every descendant as needing layout
// and draw. It walks Node.Children() so composite widgets (TabView tabs,
// Accordion content, SplitView panes, FilePicker, …) are included — not only
// Element.children. Document.Resize calls this after applying window bounds so
// grids and ColSpan breakpoints cannot be skipped by partial-layout passes.
func MarkResizeLayoutDirtySubtree(root Node) {
	var walk func(Node)
	walk = func(n Node) {
		if n == nil || n.IsHidden() {
			return
		}
		if el := nodeElementPtrForResize(n); el != nil {
			el.layoutDirty = true
			el.drawDirty = true
			if el.cachePolicy != CacheNever {
				el.cacheDirty = true
			}
		}
		for _, ch := range layoutWalkChildren(n) {
			walk(ch)
		}
	}
	walk(root)
}

// layoutWalkChildren returns children for dirty/resize walks. Surface shells include
// the internal body node so resize reflow reaches RaisedSurface layout state.
func layoutWalkChildren(n Node) []Node {
	type walkChildren interface {
		layoutWalkChildren() []Node
	}
	if wc, ok := n.(walkChildren); ok {
		return wc.layoutWalkChildren()
	}
	return n.Children()
}

// nodeElementPtrForResize returns the embedded *Element for n, or nil.
// Uses reflection so generic widgets (e.g. DataTable) and *Container wrappers
// stay covered without listing every Node implementation.
func nodeElementPtrForResize(n Node) *Element {
	v := reflect.ValueOf(n)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		return nil
	}
	ve := v.Elem()
	if ve.Kind() != reflect.Struct {
		return nil
	}
	// FilePicker-style: *Container field
	if c := ve.FieldByName("Container"); c.IsValid() {
		if c.Kind() == reflect.Ptr && !c.IsNil() {
			ce := c.Elem()
			if e := ce.FieldByName("Element"); e.IsValid() && e.CanAddr() {
				return e.Addr().Interface().(*Element)
			}
		}
		if c.Kind() == reflect.Struct {
			if e := c.FieldByName("Element"); e.IsValid() && e.CanAddr() {
				return e.Addr().Interface().(*Element)
			}
		}
	}
	if e := ve.FieldByName("Element"); e.IsValid() && e.CanAddr() {
		return e.Addr().Interface().(*Element)
	}
	return nil
}

// wheelConsumer is the interface that widgets implement to claim ownership of
// mouse-wheel scroll events while the cursor is over them. Viewport checks for
// this before deciding whether to scroll itself.
type wheelConsumer interface{ HandlesWheelScroll() bool }

// wheelScrollLimiter is optional. When implemented, deepHasWheelConsumer only
// blocks the parent when the widget can still scroll in the wheel direction.
type wheelScrollLimiter interface {
	AbsorbsParentWheel(wheel float32) bool
}

// deepHasWheelConsumer recursively searches the subtree rooted at nodes and
// returns true if any non-Viewport descendant under mouse still absorbs vertical
// wheel scroll (DataTable, VirtualList, TreeView, etc.). Viewports are resolved
// separately via findDeepestVerticalWheelViewport in wheel_gesture.go.
func deepHasWheelConsumer(nodes []Node, mouse rl.Vector2, wheel float32) bool {
	for _, n := range nodes {
		if n.IsHidden() {
			continue
		}
		if te, ok := n.(*TextEditor); ok && wheel != 0 && te.absorbsHorizontalBarWheel(mouse) {
			return true
		}
		if !rl.CheckCollisionPointRec(mouse, n.Bounds()) {
			continue
		}
		if tb, ok := n.(*Toolbar); ok && wheel != 0 && tb.absorbsRibbonWheel(mouse) {
			return true
		}
		if _, isVP := n.(*Viewport); !isVP {
			if wc, ok := n.(wheelConsumer); ok && wc.HandlesWheelScroll() {
				if lim, ok := n.(wheelScrollLimiter); ok {
					if lim.AbsorbsParentWheel(wheel) {
						return true
					}
				} else {
					return true
				}
			}
		}
		if deepHasWheelConsumer(n.Children(), mouse, wheel) {
			return true
		}
	}
	return false
}
