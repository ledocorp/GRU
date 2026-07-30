// Package ui (continued) — Container layout: flex, grid, responsive, absolute.
//
// See node.go for the full package documentation.
//
// # Container vs Viewport
//
// Container arranges children; Viewport adds scroll + clip. Use Container for
// toolbars, form rows, grid page regions, and stacked panels. Nest Viewport
// inside a grid cell when content may overflow vertically.
//
// # LLM Prompt Template — responsive page grid (12 columns)
//
//	page := ui.NewContainer("page", 0, 0, 0, 0)
//	page.LayoutType = ui.LayoutGrid
//	page.GridColumns = 12
//	page.Gap = 12
//	page.SetFlexGrow(1)
//
//	sidebar := ui.NewPanel("sidebar", "Nav", 0, 0, 0, 0)
//	sidebar.SetColSpan(ui.BreakpointMD, 3)
//	sidebar.SetColSpan(ui.BreakpointLG, 2)
//	page.AddChild(sidebar)
//
//	main := ui.NewPanel("main", "Content", 0, 0, 0, 0)
//	main.SetColSpan(ui.BreakpointMD, 9)
//	main.SetColSpan(ui.BreakpointLG, 10)
//	main.SetFlexGrow(1)
//	page.AddChild(main)
//
//	root.AddChild(page) // Document.Root is LayoutAbsolute; grid lives inside
//
// # LLM Prompt Template — vertical flex stack
//
//	col := ui.NewContainer("col", 0, 0, 0, 0)
//	col.LayoutType = ui.LayoutFlex
//	col.FlexDirection = ui.FlexColumn
//	col.Gap = 8
//	col.SetFlexGrow(1)
//	col.AddChild(ui.NewLabel("lbl", "Hello", 0, 0, 0, 0))
//
// Leaf widgets with Layout() overrides must clear layoutDirty before return
// (see docs/IDLE_INVARIANTS.md).
package ui

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// LayoutType defines how children are arranged.
type LayoutType int

const (
	LayoutNone       LayoutType = iota // No layout, manual positioning
	LayoutFlex                         // Flexbox-like layout
	LayoutGrid                         // Grid layout (future)
	LayoutAbsolute                     // Absolute positioning
	LayoutResponsive                   // Auto row↔column based on container width
)

// FlexDirection specifies the direction of flex layout.
type FlexDirection int

const (
	FlexRow    FlexDirection = iota // Horizontal layout
	FlexColumn                      // Vertical layout
)

// GridRowSizing selects how LayoutGrid assigns row height (docs/LAYOUT_CONTRACTS.md §4).
type GridRowSizing int

const (
	// GridRowSizingShrinkWrap — row height = max intrinsic cell extent (default).
	GridRowSizingShrinkWrap GridRowSizing = iota
	// GridRowSizingEqualFill — divide grid inner height equally among rows (opt-in).
	GridRowSizingEqualFill
)

// Container is a layout box that holds and positions child nodes.
//
// Containers support flex layout (row or column), an optional gap between
// children, and optional scissor clipping (ClipChildren). They are the
// primary building block for composing screens.
//
// Container vs Viewport: Container is a plain, non-scrolling layout box.
// Use Viewport when you need scrolling and strict clipping of children that
// may overflow the visible area.
type Container struct {
	Element
	LayoutType    LayoutType
	FlexDirection FlexDirection
	Gap           float32 // Spacing between children
	FlexWrap      bool    // FlexRow only: wrap children to the next line when they overflow
	ClipChildren  bool    // Whether to clip children to content area (default false)
	GridColumns   int           // Number of grid columns for LayoutGrid (default 12)
	GridRowSizing GridRowSizing // Row height policy for LayoutGrid (default ShrinkWrap)
	sortedCache   []Node        // ZIndex-sorted children cache (invalidated on layout)
	// lastLayoutPass* record c.bounds after the last completed Layout(). When
	// the container's size changes (window resize, flex parent reflow), we mark
	// flex/grid/responsive children dirty so nested Panels and grids run Layout
	// even though their Bounds() no longer show cross-axis 0 (width was filled
	// on the previous pass).
	lastLayoutPassW     float32
	lastLayoutPassH     float32
	lastLayoutPassValid bool
	// Future: AlignItems, etc.
}

// NewContainer creates a new Container with default flex layout.
func NewContainer(id string, x, y, w, h float32) *Container {
	c := &Container{
		Element:       NewElement(id, x, y, w, h),
		LayoutType:    LayoutFlex,
		FlexDirection: FlexColumn,
		Gap:           12,
		ClipChildren:  false, // Default: no clipping
	}
	c.styleName = "default"
	c.AutoHeight = (h == 0) // h=0 at creation → intrinsic (shrink-wrap) sizing
	return c
}

// GetStyle returns the resolved style for this container. Theme v2 components
// and DocBlock overrides live on Element; legacy flat styleName-only nodes
// still read CurrentTheme.
func (c *Container) GetStyle() Style {
	if c.styleComponent != "" || c.styleOverrides != nil || c.stylePatch != nil {
		return c.Element.GetStyle()
	}
	if s, ok := CurrentTheme[c.styleName]; ok {
		return s
	}
	return DefaultStyle
}

// Update implements Node.Update by updating all children.
func (c *Container) Update(dt float32) {
	if c.IsHidden() {
		return
	}
	UpdateChildrenOverlayAware(c.Children(), dt)
}

// InvalidateLayoutPassCache clears the committed layout snapshot for this
// container so the next Layout() treats bounds as changed vs the last pass.
// Document.Resize calls this on direct *Container children (page shell) so
// layoutFlex always re-runs and assigns new sizes to flex-grow / w=0 children.
func (c *Container) InvalidateLayoutPassCache() {
	c.lastLayoutPassValid = false
	c.MarkDirty()
}

// flexChildDependsOnParentFlex is true when this flex child's size is driven
// by the parent's inner width/height (cross-axis 0) or FlexGrow, when the
// child was effectively full-width / full-height on the previous layout pass
// (so parent growth must re-run flex), or when the child is a grid/responsive
// container whose columns depend on parent width.
// shouldStretchColumnCrossAxis reports whether a FlexColumn child should receive
// the parent's full inner width this pass (w=0, flex-grow, grid host, stale
// full-width after parent resize, or wider than the parent).
func shouldStretchColumnCrossAxis(c *Container, ch Node, availableW float32) bool {
	if ch.GetFlexGrow() > 0 || nodeHostsLayoutGrid(ch) {
		return true
	}
	b := ch.Bounds()
	if b.Width == 0 {
		return true
	}
	if availableW <= 0 {
		return false
	}
	if b.Width > availableW {
		return true
	}
	if flexChildFillCrossWidth(ch, availableW, b.Width) {
		return true
	}
	pad := c.GetStyle().Padding
	if c.lastLayoutPassValid {
		oldInner := c.lastLayoutPassW - 2*pad
		if oldInner > 1 && b.Width >= oldInner*0.97 {
			return true
		}
	}
	return false
}

// layoutFlexInRect runs layoutFlex in rect without leaving c.bounds mutated.
// lastLayoutPass* are recorded from rect (used by Panel/Card body layout).
func (c *Container) layoutFlexInRect(rect rl.Rectangle) {
	saved := c.bounds
	c.bounds = rect
	c.layoutFlex()
	c.lastLayoutPassW = rect.Width
	c.lastLayoutPassH = rect.Height
	c.lastLayoutPassValid = true
	c.bounds = saved
}

func flexChildDependsOnParentFlex(c *Container, ch Node) bool {
	if ch.GetFlexGrow() > 0 || nodeHasPinnedWidth(ch) {
		return true
	}
	switch ch.(type) {
	case *SplitView, *Accordion, *ResizablePanel:
		return true
	case *Toolbar, *MenuBar, *StatusBar:
		return true
	}
	b := ch.Bounds()
	if c.FlexDirection == FlexColumn {
		if b.Width == 0 {
			return true
		}
		if c.lastLayoutPassValid {
			pad := c.GetStyle().Padding
			oldInner := c.lastLayoutPassW - 2*pad
			if oldInner > 1 && b.Width >= oldInner*0.97 {
				return true
			}
		}
	} else {
		if b.Height == 0 {
			return true
		}
		if c.lastLayoutPassValid {
			pad := c.GetStyle().Padding
			oldInnerH := c.lastLayoutPassH - 2*pad
			if oldInnerH > 1 && b.Height >= oldInnerH*0.97 {
				return true
			}
		}
	}
	if nodeHostsLayoutGrid(ch) {
		return true
	}
	return false
}

// nodeHostsLayoutGrid is true when ch is a Container or Panel whose layout mode
// is grid/responsive (Panel embeds Container — type switch on *Container alone
// misses *Panel grid sections).
func nodeHostsLayoutGrid(ch Node) bool {
	switch v := ch.(type) {
	case *Container:
		return v.LayoutType == LayoutGrid || v.LayoutType == LayoutResponsive
	case *Panel:
		return v.LayoutType == LayoutGrid || v.LayoutType == LayoutResponsive
	default:
		return false
	}
}

// gridBreakpointForWidth picks the ColSpan tier for layoutGrid. Below
// MinClientWidth we always use BreakpointXS so grids stack to one column at
// narrow client widths even when CurrentBreakpoint would still be SM.
func gridBreakpointForWidth(containerWidth float32) Breakpoint {
	if containerWidth < float32(MinClientWidth) {
		return BreakpointXS
	}
	return CurrentBreakpoint(containerWidth)
}

// Layout implements Node.Layout by arranging children and recursing.
//
// Partial layout: if a child's layoutDirty flag is false it has not changed
// size or position since the last Layout pass, so we skip its subtree
// entirely — critical for dense scenes with many static Panels and Labels.
//
// # LLM Prompt Template
//
//	col := ui.NewContainer("col", 0, 0, 0, 0)
//	col.LayoutType = ui.LayoutFlex
//	col.FlexDirection = ui.FlexColumn
//	col.Gap = 8
//	col.SetFlexGrow(1)
//	col.AddChild(ui.NewLabel("lbl", "Hello", 0, 0, 0, 0))
//
// Resize: when this container's bounds differ from the last completed layout
// pass, flex children that depend on parent size (FlexGrow, cross-axis w/h=0,
// nested LayoutGrid / LayoutResponsive) are MarkDirty. On recurse, those
// children run Layout even if a stale flag left them briefly "clean", so the
// page shell always pushes new bounds into the main Viewport.
func (c *Container) Layout() {
	if c.IsHidden() {
		c.layoutDirty = false
		return
	}
	boundsChanged := !c.lastLayoutPassValid || c.lastLayoutPassW != c.bounds.Width || c.lastLayoutPassH != c.bounds.Height
	orphanDirtyChild := false
	if !c.IsDirty() && !boundsChanged {
		orphanDirtyChild = SubtreeLayoutDirty(c)
		if !orphanDirtyChild {
			return
		}
	}

	if boundsChanged {
		switch c.LayoutType {
		case LayoutFlex:
			for _, ch := range c.children {
				if ch.IsHidden() {
					continue
				}
				if flexChildDependsOnParentFlex(c, ch) {
					ch.MarkDirty()
				}
			}
			// Parent flex area grew: re-run subtree layout so FlexGrow children and
			// inner panels restore widths/heights after a widen (not only shrink).
			if c.lastLayoutPassValid {
				grew := c.bounds.Width > c.lastLayoutPassW+0.5 || c.bounds.Height > c.lastLayoutPassH+0.5
				if grew {
					for _, ch := range c.children {
						if !ch.IsHidden() {
							ch.MarkDirty()
						}
					}
				}
			}
		case LayoutGrid, LayoutResponsive:
			for _, ch := range c.children {
				if !ch.IsHidden() {
					ch.MarkDirty()
				}
			}
		case LayoutAbsolute:
			if c.parent == nil {
				for _, ch := range c.children {
					if !ch.IsHidden() {
						ch.MarkDirty()
					}
				}
			}
		}
	}

	if !orphanDirtyChild {
		switch c.LayoutType {
		case LayoutFlex:
			c.layoutFlex()
			c.syncPageShellScrollViewport()
		case LayoutGrid:
			c.layoutGrid()
		case LayoutAbsolute:
			// Children keep their manually set bounds
		case LayoutResponsive:
			c.layoutResponsive()
		}
	}

	for _, child := range c.Children() {
		if child.IsHidden() {
			continue
		}
		recurse := child.IsDirty() || SubtreeLayoutDirty(child)
		if boundsChanged {
			switch c.LayoutType {
			case LayoutFlex:
				recurse = recurse || flexChildDependsOnParentFlex(c, child)
			case LayoutGrid, LayoutResponsive:
				recurse = true
			}
		}
		if recurse {
			child.Layout()
		}
	}

	// Nested Layout (Card/table shells) can change AutoHeight block heights after
	// pass 2 stacked Y — reflow shrink-wrap columns only (not flex-grow panels).
	if c.LayoutType == LayoutFlex && c.FlexDirection == FlexColumn {
		c.reflowFlexColumnChildY()
	}

	// Rebuild the ZIndex sort cache now that positions are settled.
	c.rebuildSortedCache()
	c.layoutDirty = false

	c.lastLayoutPassW = c.bounds.Width
	c.lastLayoutPassH = c.bounds.Height
	c.lastLayoutPassValid = true
}

// InsertChildBefore inserts child immediately before before in this container's
// child list. If before is not found, child is appended.
func (c *Container) InsertChildBefore(before Node, child Node) {
	idx := -1
	for i, ch := range c.children {
		if ch == before {
			idx = i
			break
		}
	}
	if idx < 0 {
		c.AddChild(child)
		return
	}
	c.children = append(c.children[:idx], append([]Node{child}, c.children[idx:]...)...)
	child.SetParent(c)
	c.MarkDirty()
}

// syncPageShellScrollViewport assigns the flush-right horizontal band to each
// page-scroll viewport child (see SyncShellScrollViewportWidth).
func (c *Container) syncPageShellScrollViewport() {
	if c.styleName != "page-shell" {
		return
	}
	for _, ch := range c.children {
		vp, ok := ch.(*Viewport)
		if !ok || vp.styleName != "page-scroll" {
			continue
		}
		SyncShellScrollViewportWidth(c, vp)
	}
}

// flexIntrinsicProbeH marks AutoHeight column measure passes that use a tall
// probe band (see RaisedSurface layoutContent). Flex-grow children must not
// fill that band or SplitView / Viewport hosts become thousands of pixels tall.
const flexIntrinsicProbeH = float32(2048)

func flexColumnIntrinsicProbe(c *Container, availableH float32) bool {
	return c.FlexDirection == FlexColumn && c.AutoHeight && c.GetFlexGrow() == 0 && availableH >= flexIntrinsicProbeH
}

func flexGrowIntrinsicMinHeight(ch Node) float32 {
	switch v := ch.(type) {
	case *SplitView:
		if v.Direction == SplitVertical {
			return v.MinFirst + v.MinSecond + v.SplitterW + 24
		}
		return 160
	case *Viewport:
		return 0
	case *ResizablePanel:
		return 180
	}
	return 0
}

// flexContentInsets returns the inner content inset for flex layout. page-shell
// uses asymmetric horizontal inset (left/top/bottom = padding, right = 0) so
// the main page viewport extends flush to the window's right edge for the
// scrollbar gutter; page-scroll viewport padding supplies symmetric content margins.
func flexContentInsets(c *Container) (left, top, right, bottom float32) {
	p := c.GetStyle().Padding
	if c.styleName == "page-shell" {
		return p, p, 0, p
	}
	return p, p, p, p
}

// layoutFlex positions children in a single direction (FlexRow or FlexColumn).
//
// Children are placed sequentially starting at (bounds.X+padding, bounds.Y+padding)
// with Gap pixels of space between them.
//
// Flex grow: children with FlexGrow > 0 share the remaining space after all
// fixed-size children are placed. The distribution is proportional to each
// child's FlexGrow value (e.g. FlexGrow=2 gets twice as much as FlexGrow=1).
// Multiple flex children are all supported — this uses a two-pass algorithm:
// pass 1 sums fixed sizes; pass 2 assigns proportional shares of the remainder.
//
// When a flex parent's bounds **grow** vs the last completed `Container.Layout`
// pass, `Layout` marks all flex children dirty before calling `layoutFlex` so
// flex-grow and cross-axis stretch re-run with the new inner size.
//
// Cross-axis stretch: a child whose size in the cross axis is exactly 0 is
// automatically expanded to fill the container's available cross-axis space.
// In FlexColumn mode (main axis = Y) this means width=0 → availableWidth.
// In FlexRow    mode (main axis = X) this means height=0 → availableHeight,
// including when the row has AutoHeight from h=0 but a flex parent already
// assigned a positive bounds.Height (batch page rows).
func (c *Container) layoutFlex() {
	if len(c.children) == 0 {
		return
	}
	// Parent inner area grew vs last Container.Layout snapshot: flex-grow and
	// cross-axis stretch (w=0 / h=0) children must see MarkDirty so expand-after-
	// shrink reflows even when this layoutFlex is reached without a prior Layout()
	// pass (e.g. Viewport calls layoutFlex on the embedded Container directly).
	if c.lastLayoutPassValid {
		grew := c.bounds.Width > c.lastLayoutPassW+0.5 || c.bounds.Height > c.lastLayoutPassH+0.5
		shrank := c.bounds.Width < c.lastLayoutPassW-0.5 || c.bounds.Height < c.lastLayoutPassH-0.5
		if grew || shrank {
			for _, ch := range c.children {
				if ch.IsHidden() {
					continue
				}
				if flexChildDependsOnParentFlex(c, ch) {
					ch.MarkDirty()
					continue
				}
				if shrank && c.FlexDirection == FlexColumn && ch.GetFlexGrow() == 0 && ch.IsAutoHeight() {
					ch.MarkDirty()
					continue
				}
				if grew {
					if ch.GetFlexGrow() > 0 {
						ch.MarkDirty()
						continue
					}
					b := ch.Bounds()
					if c.FlexDirection == FlexColumn && b.Width == 0 {
						ch.MarkDirty()
					} else if c.FlexDirection == FlexRow && b.Height == 0 {
						ch.MarkDirty()
					}
					if nodeHostsLayoutGrid(ch) {
						ch.MarkDirty()
					}
				}
			}
		}
	}

	style := c.GetStyle()
	padding := style.Padding
	padL, padT, padR, padB := flexContentInsets(c)
	availableW := c.bounds.Width - padL - padR
	availableH := c.bounds.Height - padT - padB

	x := c.bounds.X + padL
	y := c.bounds.Y + padT

	if c.FlexDirection == FlexRow && c.FlexWrap {
		c.layoutFlexRowWrap(x, y, availableW, availableH, padding)
	} else if c.FlexDirection == FlexRow {
		// ── Pre-measure pass (FlexRow AutoHeight only) ────────────────────
		// When this FlexRow container is AutoHeight, its height is derived from
		// its tallest child. Children that are also AutoHeight must be measured
		// first so their intrinsic heights are known before we track maxChildH.
		// (For FlexGrow children we use availableW as a width approximation;
		// their heights rarely depend on width so the result is still correct.)
		if c.AutoHeight {
			// Reserve main-axis space for non-flex-grow siblings so flex-grow
			// AutoHeight children (wrapped list copy, labels beside icons) measure
			// at their final width, not the full row width.
			fixedW := float32(0)
			growAuto := 0
			autoCount := 0
			for _, child := range c.children {
				if child.IsHidden() || !child.IsAutoHeight() {
					continue
				}
				autoCount++
				if child.GetFlexGrow() > 0 {
					growAuto++
					continue
				}
				bw := child.Bounds().Width
				if bw == 0 {
					if pw, ok := flexChildPreferredWidth(child); ok {
						bw = pw
					}
				}
				if bw > 0 {
					fixedW += bw
				}
			}
			gapW := float32(0)
			if autoCount > 1 {
				gapW = float32(autoCount-1) * c.Gap
			}
			remaining := availableW - fixedW - gapW
			if remaining < 0 {
				remaining = 0
			}
			growShare := remaining
			if growAuto > 1 {
				growShare = remaining / float32(growAuto)
			}
			for _, child := range c.children {
				if child.IsHidden() || !child.IsAutoHeight() {
					continue
				}
				b := child.Bounds()
				if child.GetFlexGrow() > 0 {
					b.Width = growShare
				} else if b.Width == 0 {
					if pw, ok := flexChildPreferredWidth(child); ok {
						b.Width = pw
					}
				}
				if availableW > 0 && b.Width > availableW {
					b.Width = availableW
				}
				b.Height = 0
				child.SetBounds(b)
				child.Layout()
			}
		}

		// Pass 1: count visible children, sum fixed widths and total flex-grow.
		nonHidden := 0
		fixedW := float32(0)
		totalGrow := float32(0)
		shareWCount := 0
		for _, child := range c.children {
			if child.IsHidden() {
				continue
			}
			nonHidden++
			if fg := child.GetFlexGrow(); fg > 0 {
				totalGrow += fg
			} else {
				w := child.Bounds().Width
				if nodeHasPinnedWidth(child) {
					if pw, ok := flexChildPreferredWidth(child); ok {
						w = pw
					}
				}
				if w == 0 {
					if pw, ok := flexChildPreferredWidth(child); ok {
						fixedW += pw
					} else {
						shareWCount++
					}
				} else {
					fixedW += w
				}
			}
		}
		gaps := float32(0)
		if nonHidden > 1 {
			gaps = float32(nonHidden-1) * c.Gap
		}
		remaining := availableW - fixedW - gaps
		if remaining < 0 {
			remaining = 0
		}

		// Pass 2: position each child; assign flex-grow widths and cross-axis
		// stretch (height=0 → availableH) proportionally.
		maxChildH := float32(0) // used for FlexRow AutoHeight below
		equalShare := float32(0)
		if shareWCount > 0 && totalGrow == 0 {
			equalShare = remaining / float32(shareWCount)
		}
		for _, child := range c.children {
			if child.IsHidden() {
				continue
			}
			b := child.Bounds()
			b.X = x
			b.Y = y
			prevW := b.Width
			type widthHinter interface {
				GetPreferredWidth() float32
				GetMinWidth() float32
				GetMaxWidth() float32
			}
			if pw, pinned := PinnedMainAxisWidth(child); pinned {
				b.Width = pw
				if availableH > 0 {
					b.Height = availableH
				}
			} else if fg := child.GetFlexGrow(); fg > 0 && totalGrow > 0 {
				b.Width = remaining * (fg / totalGrow)
			} else if pw, ok := flexChildPreferredWidth(child); ok {
				b.Width = pw
			} else if b.Width == 0 && equalShare > 0 {
				b.Width = equalShare
			} else if wh, ok := child.(widthHinter); ok {
				mn := wh.GetMinWidth()
				mx := wh.GetMaxWidth()
				if pw := wh.GetPreferredWidth(); pw > 0 {
					b.Width = pw
				}
				if mn > 0 && b.Width < mn {
					b.Width = mn
				}
				if mx > 0 && b.Width > mx {
					b.Width = mx
				}
				if availableW > 0 && b.Width > availableW {
					b.Width = availableW
					if mn > 0 && b.Width < mn {
						b.Width = mn
					}
				}
			}
			// Cross-axis stretch: height=0 fills a row with a definite height.
			// Pure AutoHeight rows (no flex-grow on the row itself) shrink-wrap to
			// the tallest child — do not stretch to the parent's probe band (4096)
			// during intrinsic measure. Rows with SetFlexGrow(1) already received
			// a main-axis band from the parent and must stretch cross-axis children.
			rowHasDefiniteHeight := availableH > 0 && (!c.AutoHeight || c.GetFlexGrow() > 0)
			if b.Height == 0 && rowHasDefiniteHeight {
				b.Height = availableH
			} else if availableH > 0 && b.Height > availableH {
				b.Height = availableH
			} else if child.GetFlexGrow() > 0 && rowHasDefiniteHeight && b.Height < availableH {
				b.Height = availableH
			}
			if b.Height > maxChildH {
				maxChildH = b.Height
			}
			child.SetBounds(b)
			if child.GetFlexGrow() > 0 && child.IsAutoHeight() && absF(b.Width-prevW) > 0.5 {
				child.Layout()
				if h := child.Bounds().Height; h > maxChildH {
					maxChildH = h
				}
			}
			x += b.Width + c.Gap
		}

		// Cross-axis center: mixed-height row children (pipe + badge, icon + label).
		for _, child := range c.children {
			if child.IsHidden() {
				continue
			}
			b := child.Bounds()
			if b.Height < maxChildH {
				b.Y += (maxChildH - b.Height) / 2
				child.SetBounds(b)
			}
		}

		// ── Intrinsic height (FlexRow AutoHeight containers only) ─────────
		// Shrink-wrap to the tallest child, but never shrink below a height a
		// flex parent already assigned (e.g. column child with SetFlexGrow).
		if c.AutoHeight && nonHidden > 0 {
			intrinsic := maxChildH + padT + padB
			if c.GetFlexGrow() == 0 {
				c.bounds.Height = intrinsic
			} else if intrinsic > c.bounds.Height {
				c.bounds.Height = intrinsic
			}
		}
	} else {
		// ── Pre-measure pass (FlexColumn only) ────────────────────────────
		// For AutoHeight children (created with h=0), reset their height to 0
		// and call Layout() so they compute their intrinsic height BEFORE pass 1
		// sums fixedH. This is the "measure" phase of a two-phase layout system,
		// equivalent to CSS shrink-to-fit / Flutter's MainAxisSize.min.
		for i, child := range c.children {
			if child.IsHidden() || child.GetFlexGrow() != 0 {
				continue
			}
			if !child.IsAutoHeight() {
				continue
			}
			b := child.Bounds()
			// Apply cross-axis stretch so the child knows its available width.
			if flexChildFillCrossWidth(child, availableW, b.Width) {
				b.Width = availableW
			}
			b.Height = 0 // reset: force fresh intrinsic-height computation
			// Flex-wrap rows at the bottom of a fixed-height column must measure
			// against remaining vertical space, not the full parent height.
			if !c.AutoHeight && availableH > 0 && isFlexRowWrapContainer(child) {
				used := float32(0)
				gaps := 0
				for j := 0; j < i; j++ {
					prior := c.children[j]
					if prior.IsHidden() || prior.GetFlexGrow() > 0 {
						continue
					}
					if gaps > 0 {
						used += c.Gap
					}
					gaps++
					used += prior.Bounds().Height
				}
				rem := availableH - used
				if rem < 0 {
					rem = 0
				}
				if rem > 0 {
					b.Height = rem
				}
			}
			child.SetBounds(b)
			child.Layout() // child computes its own intrinsic height → updates bounds.Height
		}

		// Pass 1: count visible children, sum fixed heights and total flex-grow.
		nonHidden := 0
		fixedH := float32(0)
		totalGrow := float32(0)
		for _, child := range c.children {
			if child.IsHidden() {
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
			gaps = float32(nonHidden-1) * c.Gap
		}
		remaining := availableH - fixedH - gaps
		if remaining < 0 {
			remaining = 0
		}
		intrinsicProbe := flexColumnIntrinsicProbe(c, availableH)

		// Pass 2a: assign flex-grow heights and cross-axis width; measure AutoHeight
		// at the new width before stacking on Y (preview headings/tables on resize).
		for _, child := range c.children {
			if child.IsHidden() {
				continue
			}
			b := child.Bounds()
			b.X = x
			if fg := child.GetFlexGrow(); fg > 0 && totalGrow > 0 {
				if intrinsicProbe {
					if minH := flexGrowIntrinsicMinHeight(child); minH > 0 {
						b.Height = minH
					}
				} else {
					b.Height = remaining * (fg / totalGrow)
				}
			}
			type widthHinter interface {
				GetPreferredWidth() float32
				GetMinWidth() float32
				GetMaxWidth() float32
			}
			if wh, ok := child.(widthHinter); ok {
				if pw := wh.GetPreferredWidth(); pw > 0 {
					b.Width = pw
				} else if shouldStretchColumnCrossAxis(c, child, availableW) {
					b.Width = availableW
				}
				if availableW > 0 && b.Width > availableW {
					b.Width = availableW
				}
				if mn := wh.GetMinWidth(); mn > 0 && b.Width < mn {
					b.Width = mn
				}
				if mx := wh.GetMaxWidth(); mx > 0 && b.Width > mx {
					b.Width = mx
				}
			} else if shouldStretchColumnCrossAxis(c, child, availableW) {
				b.Width = availableW
			}
			child.SetBounds(b)
			if child.GetFlexGrow() == 0 && child.IsAutoHeight() {
				child.Layout()
			}
		}

		// Pass 2b: stack children on Y using heights from 2a / pre-measure.
		y = c.flexColumnStackY(x, y, availableH, padB)
	}
}

// flexColumnStackY assigns main-axis Y for each visible flex-column child after
// widths and heights are known. Returns Y just past the last child + trailing Gap.
func (c *Container) flexColumnStackY(x, startY, availableH, padB float32) float32 {
	maxContentY := startY + availableH
	y := startY
	for _, child := range c.children {
		if child.IsHidden() {
			continue
		}
		b := child.Bounds()
		b.X = x
		b.Y = y
		if !c.AutoHeight && availableH > 0 && b.Y+b.Height > maxContentY {
			b.Height = maxContentY - b.Y
			if b.Height < 0 {
				b.Height = 0
			}
		}
		layoutSetBounds(child, b)
		y += child.Bounds().Height + c.Gap
	}
	c.flexColumnUpdateIntrinsicHeight(startY, y, padB)
	return y
}

// flexColumnUpdateIntrinsicHeight shrink-wraps an AutoHeight flex column to its content.
func (c *Container) flexColumnUpdateIntrinsicHeight(startY, yAfterGap, padB float32) {
	if !c.AutoHeight || c.GetFlexGrow() > 0 {
		return
	}
	nonHidden := 0
	for _, child := range c.children {
		if !child.IsHidden() {
			nonHidden++
		}
	}
	if nonHidden == 0 {
		return
	}
	contentEnd := yAfterGap - c.Gap
	for _, child := range c.children {
		if child.IsHidden() {
			continue
		}
		if sub := nodeSubtreeBottom(child); sub > contentEnd {
			contentEnd = sub
		}
	}
	if contentEnd < startY {
		contentEnd = startY
	}
	c.bounds.Height = contentEnd - c.bounds.Y + padB
}

// reflowFlexColumnChildY re-stacks flex-column children after nested Layout passes
// (e.g. preview table Card shells) may have changed AutoHeight block heights.
// Only shrink-wrap (AutoHeight) columns — fixed-height panel bodies rely on
// pass 2 flex-grow bands and must not be reclamped here.
func (c *Container) reflowFlexColumnChildY() {
	if c.LayoutType != LayoutFlex || c.FlexDirection != FlexColumn || len(c.children) == 0 {
		return
	}
	if !c.AutoHeight || c.GetFlexGrow() > 0 {
		return
	}
	padL, padT, _, padB := flexContentInsets(c)
	x := c.bounds.X + padL
	startY := c.bounds.Y + padT
	availableH := c.bounds.Height - padT - padB
	c.flexColumnStackY(x, startY, availableH, padB)
}

// layoutFlexRowWrap places FlexRow children left-to-right, wrapping to the next
// line when the next child would exceed availableW. Flex-grow is not supported
// in wrap mode (children keep their measured width). AutoHeight containers
// grow vertically to fit all wrapped lines.
func nodeMinWidth(n Node) float32 {
	type minW interface {
		GetMinWidth() float32
	}
	if m, ok := n.(minW); ok {
		return m.GetMinWidth()
	}
	return 0
}

func isFlexRowWrapContainer(n Node) bool {
	c, ok := n.(*Container)
	return ok && c.FlexDirection == FlexRow && c.FlexWrap
}

func (c *Container) layoutFlexRowWrap(x, y, availableW, availableH, padding float32) {
	lineX := x
	lineY := y
	lineH := float32(0)
	maxX := x + availableW
	contentMaxY := float32(0)
	if !c.AutoHeight && availableH > 0 {
		contentMaxY = y + availableH
	}
	nonHidden := 0

	for _, child := range c.children {
		if child.IsHidden() {
			continue
		}
		nonHidden++
		b := child.Bounds()
		w := b.Width
		if mw := nodeMinWidth(child); mw > 0 {
			if w <= 0 || w+0.5 < mw {
				w = mw
			}
		}
		if w <= 0 && child.GetFlexGrow() == 0 {
			if child.IsAutoHeight() {
				measureW := availableW
				if mw := nodeMinWidth(child); mw > 0 && availableW >= mw {
					measureW = mw
				}
				b.Width = measureW
				b.Height = 0
				child.SetBounds(b)
				child.Layout()
				w = child.Bounds().Width
				if w <= 0 {
					w = measureW
				}
				// Raised surfaces may expand to the row width during Layout; clamp
				// back to the wrap measure width so tiles sit side-by-side.
				if mw := nodeMinWidth(child); mw > 0 && measureW <= mw+0.5 && w > mw+0.5 {
					b = child.Bounds()
					b.Width = mw
					child.SetBounds(b)
					child.MarkDirty()
					child.Layout()
					w = mw
				}
			}
		}
		if w <= 0 {
			continue
		}
		measured := child.Bounds()
		if measured.Width > 0 {
			w = measured.Width
		}
		h := measured.Height
		if h <= 0 {
			if availableH > 0 {
				h = availableH
			} else {
				h = 24
			}
		}
		if lineX > x && lineX+w > maxX+0.5 {
			lineY += lineH + c.Gap
			lineX = x
			lineH = 0
		}
		if contentMaxY > y && lineY+h > contentMaxY+0.5 {
			if lineX > x {
				lineY += lineH + c.Gap
				lineX = x
				lineH = 0
			}
			if lineY+h > contentMaxY+0.5 {
				continue
			}
		}
		b.X = lineX
		b.Y = lineY
		b.Width = w
		b.Height = h
		if availableH > 0 && h > availableH {
			b.Height = availableH
		}
		child.SetBounds(b)
		lineX += w + c.Gap
		if b.Height > lineH {
			lineH = b.Height
		}
	}

	if c.AutoHeight && nonHidden > 0 {
		contentBottom := lineY + lineH
		for _, child := range c.children {
			if child.IsHidden() {
				continue
			}
			if sub := nodeSubtreeBottom(child); sub > contentBottom {
				contentBottom = sub
			}
		}
		computed := contentBottom - c.bounds.Y + padding
		if availableH > 0 && computed > availableH {
			computed = availableH
		}
		if c.GetFlexGrow() == 0 {
			c.bounds.Height = computed
		} else if computed > c.bounds.Height {
			c.bounds.Height = computed
		}
	}
}

// SetFlexWrap enables or disables FlexRow line wrapping (see [Container.FlexWrap]).
func (c *Container) SetFlexWrap(on bool) {
	if c.FlexWrap == on {
		return
	}
	c.FlexWrap = on
	c.MarkDirty()
}

// layoutResponsive automatically picks FlexRow or FlexColumn based on the
// container's current width.  Breakpoints mirror CSS media-query conventions
// but are evaluated against this container's own width, not the window width.
//
// Width ≥ 600 px → FlexRow   (side-by-side children)
// Width < 600 px → FlexColumn (stacked children)
//
// The active FlexDirection is restored after layout so the field retains its
// default value (useful for programmatic inspection).
func (c *Container) layoutResponsive() {
	saved := c.FlexDirection
	if c.bounds.Width >= 600 {
		c.FlexDirection = FlexRow
	} else {
		c.FlexDirection = FlexColumn
	}
	c.layoutFlex()
	c.FlexDirection = saved
}

// layoutGrid implements a 12-column (or GridColumns-column) responsive grid.
//
// Children declare how many columns they span per breakpoint via SetColSpan.
// The current breakpoint is derived from the container's own width so that the
// grid reflows automatically when the window is resized — no manual wiring
// required.
//
// Algorithm:
//  1. Determine the active breakpoint from the container width.
//  2. Compute the pixel width of one column unit: (availW - gaps) / GridColumns.
//  3. Walk children in order. Track how many column units have been consumed
//     on the current row. When a child's span would overflow the row width,
//     start a new row.
//  4. Children that span the full row (ColSpanAt == GridColumns) always start
//     on their own row.
//  5. Row height = the tallest child on that row (respects existing h values).
//
// RowSpan > 1 is reserved for future implementation; currently treated as 1.
func (c *Container) layoutGrid() {
	if len(c.children) == 0 {
		return
	}

	cols := c.GridColumns
	if cols <= 0 {
		cols = 12
	}

	style := c.GetStyle()
	padding := style.Padding
	availW := c.bounds.Width - 2*padding
	// colW is the pixel width of a single column unit (before gap).
	// Total gaps on a full row = (cols-1) * Gap.
	colW := (availW - float32(cols-1)*c.Gap) / float32(cols)
	if colW < 0 {
		colW = 0
	}

	bp := gridBreakpointForWidth(c.bounds.Width)

	// gridSpacer is the interface through which layoutGrid reads ColSpan hints.
	// Element implements it; any custom Node can opt in by implementing it too.
	type gridSpacer interface {
		ColSpanAt(Breakpoint) int
	}

	startX := c.bounds.X + padding
	startY := c.bounds.Y + padding
	assignedOuterH := c.bounds.Height

	// colCursor tracks how many column units are filled on the current row.
	// rowX is the left edge of the next child.
	// rowMaxH is the tallest child on the current row (for row height).
	// rowTopY is the Y of the current row's top edge.
	colCursor := 0
	rowX := startX
	rowTopY := startY
	rowMaxH := float32(0)

	// pendingRow collects (child, bounds) pairs for the current row so we can
	// set final positions after we know the row height.
	type pending struct {
		child  Node
		bounds rl.Rectangle
	}
	var rowPending []pending
	type gridRowInfo struct {
		children []Node
		topY     float32
		height   float32
	}
	var gridRows []gridRowInfo
	rowTopAtFlush := startY
	flushRow := func() {
		if len(rowPending) == 0 {
			return
		}
		rh := rowMaxH
		rowChildren := make([]Node, 0, len(rowPending))
		for _, p := range rowPending {
			b := p.bounds
			// Cross-axis stretch: if child height is 0, fill the row height.
			if b.Height == 0 {
				b.Height = rh
			}
			p.child.SetBounds(b)
			rowChildren = append(rowChildren, p.child)
		}
		gridRows = append(gridRows, gridRowInfo{
			children: rowChildren,
			topY:     rowTopAtFlush,
			height:   rh,
		})
		rowPending = rowPending[:0]
		colCursor = 0
		rowTopY += rowMaxH + c.Gap
		rowTopAtFlush = rowTopY
		rowX = startX
		rowMaxH = 0
	}

	for _, child := range c.children {
		if child.IsHidden() {
			continue
		}

		span := cols // default: full row
		if gs, ok := child.(gridSpacer); ok {
			span = gs.ColSpanAt(bp)
		}
		if span > cols {
			span = cols
		}
		if span <= 0 {
			span = cols // default: full row
		}

		// If this child doesn't fit on the current row, flush and start a new one.
		if colCursor > 0 && colCursor+span > cols {
			flushRow()
		}

		// Pixel width for this span: span column units + (span-1) internal gaps.
		childW := colW*float32(span) + c.Gap*float32(span-1)

		b := child.Bounds()
		b.X = rowX
		b.Y = rowTopY
		b.Width = childW
		// h=0 / AutoHeight: measure intrinsic height so rowMaxH is not stuck at 0
		// (panels with flex-grow list regions need a real cell height).
		if b.Height == 0 && child.IsAutoHeight() {
			// Width-first: measure intrinsic height at assigned cell width (Phase D1).
			mb := b
			mb.Width = childW
			mb.Height = 0
			child.SetBounds(mb)
			child.Layout()
			b = child.Bounds()
			b.X = rowX
			b.Y = rowTopY
			b.Width = childW
		}
		extH := b.Height
		if child.IsAutoHeight() && child.GetFlexGrow() == 0 {
			extH = nodeLayoutExtentHeight(child)
			b.Height = extH
		}
		if extH > rowMaxH {
			rowMaxH = extH
		}

		rowPending = append(rowPending, pending{child, b})
		colCursor += span
		rowX += childW + c.Gap

		// If the row is exactly full, flush it.
		if colCursor >= cols {
			flushRow()
		}
	}

	// Flush any partial last row.
	flushRow()

	contentH := rowTopY - startY - c.Gap
	if contentH < 0 {
		contentH = 0
	}
	computed := contentH + 2*padding
	innerTarget := assignedOuterH - 2*padding
	if innerTarget < 0 {
		innerTarget = 0
	}

	distributed := false
	// EqualFill: opt-in viewport band split (docs/LAYOUT_CONTRACTS.md §4).
	if c.GridRowSizing == GridRowSizingEqualFill &&
		c.GetFlexGrow() > 0 && innerTarget > contentH+0.5 && len(gridRows) > 0 {
		perRow := innerTarget / float32(len(gridRows))
		y := startY
		for _, row := range gridRows {
			for _, ch := range row.children {
				b := ch.Bounds()
				b.Y = y
				b.Height = perRow
				ch.SetBounds(b)
				ch.MarkDirty()
			}
			y += perRow + c.Gap
		}
		c.bounds.Height = assignedOuterH
		distributed = true
		for _, row := range gridRows {
			for _, ch := range row.children {
				if !ch.IsHidden() {
					ch.Layout()
				}
			}
		}
	}

	if c.AutoHeight && rowTopY > startY {
		if !distributed {
			c.bounds.Height = computed
		}
	} else if !c.AutoHeight && distributed {
		c.bounds.Height = assignedOuterH
	}
}

// rebuildSortedCache rebuilds the ZIndex-sorted child slice.
// Called at the end of Layout() when child order may have changed.
func (c *Container) rebuildSortedCache() {
	c.sortedCache = make([]Node, len(c.children))
	copy(c.sortedCache, c.children)
	sort.Slice(c.sortedCache, func(i, j int) bool {
		return c.sortedCache[i].GetZIndex() < c.sortedCache[j].GetZIndex()
	})
}

// sortedChildren returns the ZIndex-sorted child slice, building it on first use.
func (c *Container) sortedChildren() []Node {
	if c.sortedCache == nil {
		c.rebuildSortedCache()
	}
	return c.sortedCache
}

// Draw implements Node.Draw by drawing children (no background by default).
func (c *Container) Draw() {
	c.drawInternal() // Direct call, no caching delegation
	c.drawDirty = false
}

// Draw renders the container background, border, and all children.
//
// Children are sorted by ZIndex before drawing so higher-ZIndex widgets
// (e.g. open Dropdowns) always appear on top of lower-ZIndex siblings.
// If ClipChildren is true, a scissor rect is set to the content area
// (bounds inset by style.Padding) before drawing children.
//
// When ClipChildren is true, drawChildrenInRectClip re-applies the content
// scissor after each child (Labels, TextInput, etc. call EndScissorMode).
// For scroll + width simulation, prefer a Viewport inside the host.
func (c *Container) drawInternal() {
	if c.IsHidden() {
		return
	}

	style := c.GetStyle()
	bounds := c.Bounds()

	// Draw background first
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangle(int32(bounds.X), int32(bounds.Y), int32(bounds.Width), int32(bounds.Height), style.BackgroundColor)
		if style.BorderWidth > 0 {
			rl.DrawRectangleLinesEx(bounds, float32(style.BorderWidth), style.BorderColor)
		}
	}

	// Sort children by ZIndex for layering (cached, rebuilt only on Layout)
	sortedChildren := c.sortedChildren()

	// Optionally clip children to content area (per-child scissor restore — same
	// model as Viewport.Draw and Panel/Card body clip).
	if c.ClipChildren {
		padding := style.Padding
		contentClip := containerContentDrawClip(bounds, padding, sortedChildren)
		drawChildrenInRectClip(c, contentClip, sortedChildren)
	} else {
		// Draw children, restoring the ancestor Viewport's scissor after each
		// child that may have clobbered it (ProgressBar, TextInput, etc.).
		drawChildrenWithScissorRestore(c, sortedChildren)
	}
}
