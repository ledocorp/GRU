// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── TreeView ────────────────────────────────────────────────────────────────
//
// TreeView displays a hierarchical tree of labelled nodes. Users can expand /
// collapse branches and click or keyboard-navigate to select items. The
// selected node is exposed via the reactive Selected signal.
//
// # Basic usage
//
//	root := ui.NewTreeNode("root", "Project")
//	src := root.AddChild(ui.NewTreeNode("src", "src/"))
//	src.AddChild(ui.NewTreeNode("main", "main.go"))
//	src.AddChild(ui.NewTreeNode("ui",   "ui/"))
//
//	tv := ui.NewTreeView("tree", root, 10, 10, 260, 400)
//	tv.Selected.Subscribe(func() {
//	    n := tv.Selected.Get()
//	    if n != nil { fmt.Println("selected:", n.Label) }
//	})
//
// # Expand / Collapse
//
// Clicking a node's toggle arrow, or pressing Left/Right while the node is
// focused, expands or collapses it. The reveal animation tweens the child
// area height from 0 to full over 150 ms (EaseOutQuad). Collapsing tweens
// back to 0 then hides.
//
// # Keyboard Navigation
//
// TreeView receives focus the first time a row is clicked. While focused:
//
//	Up / Down  — move selection one visible row up / down
//	Right      — expand selected node (or descend if already expanded)
//	Left       — collapse selected node (or ascend to parent)
//	Home / End — jump to first / last visible row
//
// # Icons
//
// Each TreeNode carries an optional Icon string (typically a Unicode symbol,
// e.g. "📁", "📄", "⚙"). When non-empty the icon is drawn before the label
// in the node style's TextColor.
//
// # Scissor clipping
//
// TreeView clips its content to its own bounds via BeginScissorMode. When
// nested inside a Viewport the scissor rectangle is intersected with the
// Viewport's ClipBounds() so content never escapes the outer panel.
// UsesScissor() returns true so the Viewport re-applies its own scissor after
// each TreeView draw call.
//
// # Style
//
// Theme keys used:
//   - "treeview"          — container background and default text style
//   - "treeview-selected" — highlighted row (BackgroundColor = fill, TextColor = label)
//   - "treeview-hover"    — hover row highlight (must exist in theme)
//
// # Inspector integration
//
// The Inspector pane shows total node count, expanded node count, visible row
// count, scroll position, and the selected node's label + path.

const (
	treeRowH      = float32(22) // height of each visible row
	treeBottomPad = float32(8)  // extra space below the last row (avoids scissor clip)
	treeIndent  = float32(16)   // horizontal indent per depth level
	treeToggleW = float32(14)   // width reserved for the expand/collapse arrow
	treeIconW   = float32(16)   // width reserved for the optional icon glyph
	treeAnimDur = float32(0.15) // expand/collapse animation duration (seconds)
)

// ─── TreeNode ────────────────────────────────────────────────────────────────

// TreeNode is one node in the tree data model. Build trees by creating nodes
// with NewTreeNode and calling AddChild.
type TreeNode struct {
	// ID is a unique string identifier used for path tracking.
	ID string

	// Label is the display string shown in the row.
	Label string

	// Icon is an optional glyph or Unicode symbol drawn before the label.
	// Leave empty for no icon.
	Icon string

	// FolderGlyph draws a built-in vector folder tab (no font glyph) before the
	// label. Use for directory rows when the UI font lacks emoji coverage.
	FolderGlyph bool

	// Children are the sub-nodes. Use AddChild to append.
	Children []*TreeNode

	// Data is an arbitrary caller-supplied payload; nil by default.
	Data any

	// expanded is the current logical expand state.
	expanded bool

	// clipH is the current animated clip height for child reveal/hide.
	// 0 = fully collapsed, ≥ childFullH = fully expanded.
	clipH float32

	// animating is true while a Tween is running for this node.
	animating bool
}

// NewTreeNode creates a tree data node with the given unique ID and label.
func NewTreeNode(id, label string) *TreeNode {
	return &TreeNode{ID: id, Label: label}
}

// AddChild appends child to this node and returns child for chaining.
func (n *TreeNode) AddChild(child *TreeNode) *TreeNode {
	n.Children = append(n.Children, child)
	return child
}

// IsExpanded reports whether this node is currently in the expanded state.
func (n *TreeNode) IsExpanded() bool { return n.expanded }

// SetExpanded sets the expanded state without animation. Prefer calling
// TreeView.toggleNode for the animated version.
func (n *TreeNode) SetExpanded(v bool) {
	n.expanded = v
	if v {
		n.clipH = n.childFullHeight()
	} else {
		n.clipH = 0
	}
}

// childFullHeight returns the total pixel height needed to show all direct
// children (not recursively — the tree handles recursion via flatRows).
// Used to set the animation endpoint.
func (n *TreeNode) childFullHeight() float32 {
	return treeRowH * float32(n.countVisible())
}

// countVisible counts total visible rows under this node (recursively),
// assuming all currently expanded sub-nodes are also fully expanded.
func (n *TreeNode) countVisible() int {
	count := 0
	for _, c := range n.Children {
		count++ // the child row itself
		if c.expanded {
			count += c.countVisible()
		}
	}
	return count
}

// ─── flatRow ─────────────────────────────────────────────────────────────────

// flatRow is an entry in the linearised visible-row list rebuilt each frame.
type flatRow struct {
	node  *TreeNode
	depth int
}

// ─── TreeView ────────────────────────────────────────────────────────────────

// TreeView is a scrollable, selectable tree widget.
//
// # LLM Prompt Template
//
//	root := ui.NewTreeNode("root", "Project")
//	src := root.AddChild(ui.NewTreeNode("src", "src/"))
//	src.AddChild(ui.NewTreeNode("main.go", "main.go"))
//	tv := ui.NewTreeView("tree", root, 0, 0, 0, 0)
//	tv.Selected.Subscribe(func() { /* react */ })
//	panel.AddChild(tv)
//
// Demo scenes: **Batch 1**, **Widgets Demo**, **FilePicker** (folder tree).
type TreeView struct {
	Element

	// Root is the root node of the tree. It is always shown as the first row.
	// Swap Root and call Rebuild() to change the whole tree at runtime.
	Root *TreeNode

	// Selected holds the currently selected *TreeNode, or nil.
	// Subscribe to react to selection changes.
	Selected *Signal[*TreeNode]

	// ScrollY is the vertical scroll offset in pixels (0 = top).
	ScrollY float32

	// ShowRoot controls whether the root node itself is rendered as a row.
	// When false only the root's children are shown. Default: true.
	ShowRoot bool

	// focused is true when the widget has keyboard focus (set on first click).
	focused bool

	rows    []flatRow // linearised visible rows; rebuilt when dirty
	dirty   bool      // set to trigger a Rebuild on next Update/Draw
	tweens  []*Tween  // active expand/collapse animations
	hovered int       // hovered row index, -1 = none
}

// NewTreeView creates a TreeView displaying the given root node.
// ShowRoot defaults to true; set tv.ShowRoot = false to hide the root row.
func NewTreeView(id string, root *TreeNode, x, y, w, h float32) *TreeView {
	tv := &TreeView{
		Element:  NewElement(id, x, y, w, h),
		Root:     root,
		Selected: NewSignal[*TreeNode](nil),
		ShowRoot: true,
		dirty:    true,
		hovered:  -1,
	}
	tv.styleName = "treeview"
	tv.Selected.Subscribe(func() { tv.MarkDrawDirty() })
	return tv
}

// SetStyle sets the base theme key (default: "treeview").
func (tv *TreeView) SetStyle(name string) { tv.styleName = name }

// IsInteractive reports that TreeView accepts mouse and keyboard input.
func (tv *TreeView) IsInteractive() bool { return true }

// UsesScissor reports that TreeView opens its own scissor region.
func (tv *TreeView) UsesScissor() bool { return true }

// Layout clamps internal scroll when the parent assigns new bounds.
func (tv *TreeView) Layout() {
	defer func() { tv.layoutDirty = false }()
	maxScroll := tv.contentScrollHeight() - tv.bounds.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if tv.ScrollY > maxScroll {
		tv.ScrollY = maxScroll
	}
	if tv.ScrollY < 0 {
		tv.ScrollY = 0
	}
}

// HandlesWheelScroll opts the TreeView into consuming wheel events itself so
// an outer Viewport does not scroll while the mouse is over the tree.
func (tv *TreeView) HandlesWheelScroll() bool { return true }

func (tv *TreeView) contentScrollHeight() float32 {
	pad := float32(GetThemeStyle(tv.styleName).Padding)
	return pad + float32(len(tv.rows))*treeRowH + treeBottomPad
}

// AbsorbsParentWheel implements wheelScrollLimiter.
func (tv *TreeView) AbsorbsParentWheel(wheel float32) bool {
	b := tv.Bounds()
	max := tv.contentScrollHeight() - b.Height
	if max <= 0 {
		return false
	}
	const eps = float32(0.5)
	if wheel < 0 && tv.ScrollY >= max-eps {
		return false
	}
	if wheel > 0 && tv.ScrollY <= eps {
		return false
	}
	return true
}

// Rebuild linearises the visible tree into tv.rows. Call after structural
// changes to the tree (add/remove nodes). The widget calls this automatically
// when dirty.
func (tv *TreeView) Rebuild() {
	tv.rows = tv.rows[:0]
	if tv.Root != nil {
		if tv.ShowRoot {
			tv.flattenNode(tv.Root, 0)
		} else {
			for _, c := range tv.Root.Children {
				tv.flattenNode(c, 0)
			}
		}
	}
	tv.dirty = false
}

func (tv *TreeView) flattenNode(n *TreeNode, depth int) {
	tv.rows = append(tv.rows, flatRow{node: n, depth: depth})
	if n.expanded || n.clipH > 0 {
		for _, c := range n.Children {
			tv.flattenNode(c, depth+1)
		}
	}
}

// ─── Update ──────────────────────────────────────────────────────────────────

// Update handles mouse interaction, keyboard navigation, and tween advances.
func (tv *TreeView) Update(dt float32) {
	if tv.IsHidden() {
		return
	}

	// Advance animations.
	alive := tv.tweens[:0]
	for _, tw := range tv.tweens {
		tw.Update(dt)
		if tw.IsActive {
			alive = append(alive, tw)
		}
	}
	if len(alive) != len(tv.tweens) {
		tv.tweens = alive
		tv.dirty = true
		tv.MarkDrawDirty()
	}

	if tv.dirty {
		tv.Rebuild()
	}

	b := tv.Bounds()
	mouse := rl.GetMousePosition()
	inBounds := rl.CheckCollisionPointRec(mouse, b)
	prevHovered := tv.hovered
	tv.hovered = -1

	// Lose keyboard focus when user clicks outside this tree.
	if !inBounds && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		tv.focused = false
	}

	// Wheel scroll (only when mouse is inside bounds).
	if inBounds {
		wheel := rl.GetMouseWheelMove()
		if wheel != 0 {
			tv.ScrollY -= wheel * treeRowH * 3
			tv.clampScroll()
			tv.MarkDrawDirty()
		}
	}

	// Row hit-testing. Rows are drawn at b.Y + pad + i*treeRowH - ScrollY.
	if inBounds {
		pad := float32(GetThemeStyle(tv.styleName).Padding)
		y := b.Y + pad - tv.ScrollY
		for i, row := range tv.rows {
			rowTop := y + float32(i)*treeRowH
			rowRect := rl.NewRectangle(b.X, rowTop, b.Width, treeRowH)
			if rl.CheckCollisionPointRec(mouse, rowRect) {
				tv.hovered = i
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					tv.focused = true
					// Toggle expand on click of toggle zone or anywhere on the row.
					if len(row.node.Children) > 0 {
						toggleRect := rl.NewRectangle(
							b.X+float32(row.depth)*treeIndent+4,
							rowTop, treeToggleW+4, treeRowH)
						if rl.CheckCollisionPointRec(mouse, toggleRect) {
							tv.toggleNode(row.node)
						} else {
							// Click anywhere else on a parent row also selects it.
						}
					}
					tv.Selected.Set(row.node)
				}
			}
		}
	}

	if tv.hovered != prevHovered {
		tv.MarkDrawDirty()
	}

	// Keyboard navigation (only when focused).
	if tv.focused {
		tv.handleKeys()
	}

	if tv.dirty {
		tv.Rebuild()
	}
}

// handleKeys processes arrow key navigation while the tree is focused.
func (tv *TreeView) handleKeys() {
	sel := tv.Selected.Get()
	selIdx := -1
	for i, r := range tv.rows {
		if r.node == sel {
			selIdx = i
			break
		}
	}

	switch {
	case rl.IsKeyPressed(rl.KeyDown):
		next := selIdx + 1
		if next < len(tv.rows) {
			tv.Selected.Set(tv.rows[next].node)
			tv.scrollToRow(next)
		}
	case rl.IsKeyPressed(rl.KeyUp):
		prev := selIdx - 1
		if prev >= 0 {
			tv.Selected.Set(tv.rows[prev].node)
			tv.scrollToRow(prev)
		}
	case rl.IsKeyPressed(rl.KeyRight):
		if sel != nil && len(sel.Children) > 0 {
			if !sel.expanded {
				tv.toggleNode(sel)
			}
		}
	case rl.IsKeyPressed(rl.KeyLeft):
		if sel != nil {
			if sel.expanded {
				tv.toggleNode(sel)
			} else {
				// Move to parent.
				parent := tv.findParent(tv.Root, sel)
				if parent != nil {
					tv.Selected.Set(parent)
					for i, r := range tv.rows {
						if r.node == parent {
							tv.scrollToRow(i)
							break
						}
					}
				}
			}
		}
	case rl.IsKeyPressed(rl.KeyHome):
		if len(tv.rows) > 0 {
			tv.Selected.Set(tv.rows[0].node)
			tv.ScrollY = 0
			tv.MarkDrawDirty()
		}
	case rl.IsKeyPressed(rl.KeyEnd):
		if len(tv.rows) > 0 {
			last := len(tv.rows) - 1
			tv.Selected.Set(tv.rows[last].node)
			tv.scrollToRow(last)
		}
	}
}

// toggleNode animates expand/collapse for node n.
func (tv *TreeView) toggleNode(n *TreeNode) {
	if n.animating {
		return // Don't interrupt mid-animation.
	}
	fullH := treeRowH * float32(n.countVisible())
	if fullH == 0 {
		return
	}
	if !n.expanded {
		// Expand: grow clipH from 0 to fullH.
		n.expanded = true
		n.clipH = 0
		start := n.clipH
		tw := NewTween(start, fullH, treeAnimDur, EaseOutQuad,
			func(v float32) { n.clipH = v; tv.dirty = true; tv.MarkDrawDirty() },
			func() { n.clipH = fullH; n.animating = false; tv.dirty = true },
		)
		n.animating = true
		tv.tweens = append(tv.tweens, tw)
	} else {
		// Collapse: shrink clipH from fullH to 0, then set expanded=false.
		start := n.clipH
		tw := NewTween(start, 0, treeAnimDur, EaseOutQuad,
			func(v float32) { n.clipH = v; tv.dirty = true; tv.MarkDrawDirty() },
			func() { n.expanded = false; n.clipH = 0; n.animating = false; tv.dirty = true },
		)
		n.animating = true
		tv.tweens = append(tv.tweens, tw)
	}
}

// findParent returns the direct parent of target within the subtree rooted at n.
func (tv *TreeView) findParent(n, target *TreeNode) *TreeNode {
	for _, c := range n.Children {
		if c == target {
			return n
		}
		if p := tv.findParent(c, target); p != nil {
			return p
		}
	}
	return nil
}

// scrollToRow adjusts ScrollY so row i is fully visible.
func (tv *TreeView) scrollToRow(i int) {
	b := tv.Bounds()
	rowTop := float32(i) * treeRowH
	rowBot := rowTop + treeRowH
	if rowTop < tv.ScrollY {
		tv.ScrollY = rowTop
	} else if rowBot > tv.ScrollY+b.Height {
		tv.ScrollY = rowBot - b.Height
	}
	tv.clampScroll()
	tv.MarkDrawDirty()
}

// clampScroll keeps ScrollY within valid bounds.
func (tv *TreeView) clampScroll() {
	b := tv.Bounds()
	maxScroll := tv.contentScrollHeight() - b.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if tv.ScrollY < 0 {
		tv.ScrollY = 0
	}
	if tv.ScrollY > maxScroll {
		tv.ScrollY = maxScroll
	}
}

// ─── Draw ─────────────────────────────────────────────────────────────────────

// Draw renders the tree with scissor clipping.
func (tv *TreeView) Draw() {
	if tv.IsHidden() {
		return
	}
	if tv.dirty {
		tv.Rebuild()
	}

	b := tv.Bounds()
	style := GetThemeStyle("treeview")
	styleSel := GetThemeStyle("treeview-selected")
	styleHov := GetThemeStyle("treeview-hover")

	// Container background + border.
	dim := mathMaxF32(b.Width, b.Height)
	corner := float32(style.CornerRadius) / dim
	rl.DrawRectangleRounded(b, corner, 6, style.BackgroundColor)
	if style.BorderWidth > 0 {
		rl.DrawRectangleRoundedLinesEx(b, corner, 6, style.BorderWidth, style.BorderColor)
	}

	// Build scissor rect, intersected with any ancestor Viewport.
	clipRect := rl.NewRectangle(b.X+1, b.Y+1, b.Width-2, b.Height-2)
	if vp := findViewport(tv); vp != nil {
		clipRect = intersectRects(clipRect, vp.ClipBounds())
	}
	if clipRect.Width <= 0 || clipRect.Height <= 0 {
		return
	}
	beginScissorMode(int32(clipRect.X), int32(clipRect.Y), int32(clipRect.Width), int32(clipRect.Height))

	sel := tv.Selected.Get()
	pad := float32(style.Padding)

	for i, row := range tv.rows {
		ry := b.Y + pad + float32(i)*treeRowH - tv.ScrollY
		if ry+treeRowH < clipRect.Y || ry > clipRect.Y+clipRect.Height {
			continue
		}

		isSelected := row.node == sel
		isHovered := tv.hovered == i

		rowRect := rl.NewRectangle(b.X+1, ry, b.Width-2, treeRowH)

		// Row background.
		if isSelected {
			rl.DrawRectangleRounded(rowRect, 0.3, 4, styleSel.BackgroundColor)
		} else if isHovered {
			rl.DrawRectangleRec(rowRect, styleHov.BackgroundColor)
		}

		// Indent.
		indent := float32(row.depth)*treeIndent + pad
		x := b.X + indent

		// Expand/collapse triangle.
		midY := ry + treeRowH/2
		if len(row.node.Children) > 0 {
			col := style.TextColor
			if isSelected {
				col = styleSel.TextColor
			}
			if row.node.expanded || row.node.clipH > 0 {
				// ▼ down arrow
				rl.DrawTriangle(
					rl.NewVector2(x+2, midY-3),
					rl.NewVector2(x+treeToggleW-2, midY-3),
					rl.NewVector2(x+treeToggleW/2, midY+4),
					col)
			} else {
				// ▶ right arrow
				rl.DrawTriangle(
					rl.NewVector2(x+3, midY-5),
					rl.NewVector2(x+3, midY+5),
					rl.NewVector2(x+treeToggleW-1, midY),
					col)
			}
		}
		x += treeToggleW + 2

		// Icon (if present).
		s := style
		if isSelected {
			s = styleSel
		}
		if row.node.Icon != "" {
			ty := int32(ry) + (int32(treeRowH)-s.FontSize)/2
			drawTextS(row.node.Icon, int32(x), ty, s)
			x += treeIconW
		} else if row.node.FolderGlyph {
			drawTreeFolderGlyph(x+1, ry+4, treeRowH-8, s.TextColor)
			x += treeIconW
		}

		// Label text.
		ty := int32(ry) + (int32(treeRowH)-s.FontSize)/2
		drawTextS(row.node.Label, int32(x), ty, s)
	}

	rl.EndScissorMode()

	// Scrollbar (if content overflows).
	totalH := tv.contentScrollHeight()
	if totalH > b.Height {
		sbW := float32(6)
		ratio := b.Height / totalH
		sbH := b.Height * ratio
		sbY := b.Y + (tv.ScrollY/totalH)*b.Height
		sbRect := rl.NewRectangle(b.X+b.Width-sbW-2, sbY, sbW, sbH)
		rl.DrawRectangleRounded(sbRect, 1, 4, rl.NewColor(180, 182, 210, 160))
	}
}

// drawTreeFolderGlyph draws a small outline folder tab + body (no font glyphs).
func drawTreeFolderGlyph(x, y, h float32, col rl.Color) {
	tabW := float32(6)
	tabH := float32(5)
	bodyW := float32(13)
	bodyH := h - tabH
	if bodyH < 3 {
		bodyH = 3
	}
	rl.DrawRectangleLinesEx(rl.NewRectangle(x, y, tabW, tabH), 1, col)
	rl.DrawRectangleLinesEx(rl.NewRectangle(x, y+tabH-1, bodyW, bodyH), 1, col)
}

// ─── Inspector helpers ────────────────────────────────────────────────────────

// TotalNodeCount returns the total number of nodes in the tree (all depths).
func (tv *TreeView) TotalNodeCount() int {
	if tv.Root == nil {
		return 0
	}
	return countAllNodes(tv.Root)
}

func countAllNodes(n *TreeNode) int {
	total := 1
	for _, c := range n.Children {
		total += countAllNodes(c)
	}
	return total
}

// ExpandedNodeCount returns the number of nodes that are currently expanded.
func (tv *TreeView) ExpandedNodeCount() int {
	if tv.Root == nil {
		return 0
	}
	return countExpanded(tv.Root)
}

func countExpanded(n *TreeNode) int {
	total := 0
	if n.expanded {
		total++
	}
	for _, c := range n.Children {
		total += countExpanded(c)
	}
	return total
}

// SelectedPath returns the path from root to the selected node as a
// slash-separated ID string, e.g. "root/src/main". Returns "" if nothing
// is selected.
func (tv *TreeView) SelectedPath() string {
	sel := tv.Selected.Get()
	if sel == nil || tv.Root == nil {
		return ""
	}
	path := buildPath(tv.Root, sel)
	return path
}

func buildPath(n, target *TreeNode) string {
	if n == target {
		return n.ID
	}
	for _, c := range n.Children {
		if p := buildPath(c, target); p != "" {
			return fmt.Sprintf("%s/%s", n.ID, p)
		}
	}
	return ""
}

// ─── math helper ─────────────────────────────────────────────────────────────

// math.Max on float32 via the standard library (float64 cast).
func mathMaxF32(a, b float32) float32 {
	if a > b {
		return a
	}
	return b
}
