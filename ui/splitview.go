// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── SplitView ───────────────────────────────────────────────────────────────
//
// SplitView divides a rectangular area into two panes separated by a draggable
// splitter bar.  Either pane can hold any Node — a Panel, a Container, a
// VirtualList, or even a nested SplitView.
//
// # Direction
//
// SplitHorizontal (default): left pane | splitter | right pane
// SplitVertical:             top pane  / splitter / bottom pane
//
// # Ratio
//
// Ratio is a reactive Signal[float32] clamped to [0,1] that describes what
// fraction of the splittable space belongs to the first pane.  0.5 is an
// equal split.  Changing Ratio programmatically updates child bounds immediately
// and triggers a relayout.
//
//	sv.Ratio.Set(0.3) // give 30 % to the first pane
//
// # Minimum sizes
//
// MinFirst and MinSecond define the minimum pixel size of each pane in the
// split direction.  The splitter cannot be dragged past a position that would
// violate either minimum; Ratio is also clamped on resize (parent SetBounds).
//
// # Splitter thickness
//
// SplitterW is the pixel thickness of the divider bar (default 5).  The bar
// is centred on the split point so it intrudes 2.5 px into each pane, which
// is invisible to the eye at normal UI scales.
//
// # Children
//
// Use SetFirst / SetSecond to assign the two pane contents (or pass them to
// NewSplitView).  The standard Node.Children() method returns [first, second]
// so Inspector tree-walking works without any special casing.
//
// # Events
//
// OnRatioChanged is called after every drag frame with the new Ratio value.
// The cursor changes to a resize arrow while hovering or dragging the splitter.
//
// # Style keys
//
//   - "splitview"            — background of the whole widget area (optional)
//   - "splitview-splitter"   — idle splitter fill + edge colour; hover/drag tint toward primary indigo (polish §4.3)
//
// # Example — horizontal split inside a Panel
//
//	sv := ui.NewSplitView("main-split",
//	    ui.SplitHorizontal,
//	    leftPanel, rightPanel,
//	    0, 0, 900, 600)
//	sv.Ratio.Set(0.4)
//	sv.MinFirst  = 160
//	sv.MinSecond = 120
//	viewport.AddChild(sv)
//
// # LLM Prompt Template
//
//	sv := ui.NewSplitView("editor-split", ui.SplitHorizontal, tree, editor, 0, 0, 0, 0)
//	sv.Ratio.Set(0.28)
//	sv.MinFirst, sv.MinSecond = 180, 240
//	sv.SetFlexGrow(1)
//	parent.AddChild(sv)
//
// Demo scenes: **Batch 8 SplitView**, **Notepad** (editor/preview + main split).

// SplitDirection specifies whether the split is left/right or top/bottom.
type SplitDirection int

const (
	// SplitHorizontal splits the area into a left pane and a right pane.
	SplitHorizontal SplitDirection = iota
	// SplitVertical splits the area into a top pane and a bottom pane.
	SplitVertical
)

// ─── Constants ───────────────────────────────────────────────────────────────

const (
	// svDefaultSplitterW is the hit-target thickness of the splitter bar (px).
	svDefaultSplitterW = float32(5)
	// svDefaultMinPane is the default minimum pixel size for each pane.
	svDefaultMinPane = float32(40)
	// svHoverExtra is extra hit-test padding on each side of the splitter (px).
	svHoverExtra = float32(6)
)

// ─── SplitView ───────────────────────────────────────────────────────────────

// SplitView is a two-pane resizable container.
type SplitView struct {
	Element

	// Direction is the orientation of the split (default SplitHorizontal).
	Direction SplitDirection

	// Ratio is the fraction [0,1] of splittable space assigned to the first pane.
	// Listen to changes with Ratio.Subscribe(fn).
	Ratio *Signal[float32]

	// MinFirst is the minimum pixel size of the first pane (default 40).
	MinFirst float32

	// MinSecond is the minimum pixel size of the second pane (default 40).
	MinSecond float32

	// SplitterW is the width / height of the divider bar in pixels (default 5).
	SplitterW float32

	// OnRatioChanged is called after each drag-frame with the updated ratio.
	OnRatioChanged func(ratio float32)

	// first / second hold the two pane content nodes.
	first  Node
	second Node

	// drag state
	dragging   bool
	dragMouse0 float32 // mouse coord (X or Y) at drag start
	dragRatio0 float32 // Ratio value at drag start
	hoverSplit bool    // mouse is hovering over the splitter bar

	lastLayoutPassW     float32
	lastLayoutPassH     float32
	lastLayoutPassValid bool
}

// NewSplitView creates a SplitView with the given direction and initial pane
// contents.  Either first or second may be nil; use SetFirst / SetSecond later.
//
//	sv := ui.NewSplitView("split", ui.SplitHorizontal, leftNode, rightNode, 0, 0, 800, 600)
func NewSplitView(id string, dir SplitDirection, first, second Node, x, y, w, h float32) *SplitView {
	sv := &SplitView{
		Element:   NewElement(id, x, y, w, h),
		Direction: dir,
		Ratio:     NewSignal(float32(0.5)),
		MinFirst:  svDefaultMinPane,
		MinSecond: svDefaultMinPane,
		SplitterW: svDefaultSplitterW,
	}
	sv.styleName = "splitview"
	sv.Ratio.Subscribe(func() {
		sv.MarkDirty()
		sv.MarkDrawDirty()
	})
	if first != nil {
		sv.SetFirst(first)
	}
	if second != nil {
		sv.SetSecond(second)
	}
	return sv
}

// SetFirst assigns the first pane content node (left or top depending on Direction).
func (sv *SplitView) SetFirst(n Node) {
	sv.first = n
	n.SetParent(sv)
	sv.MarkDirty()
}

// SetSecond assigns the second pane content node (right or bottom).
func (sv *SplitView) SetSecond(n Node) {
	sv.second = n
	n.SetParent(sv)
	sv.MarkDirty()
}

// GetFirst returns the first pane content node (may be nil).
func (sv *SplitView) GetFirst() Node { return sv.first }

// GetSecond returns the second pane content node (may be nil).
func (sv *SplitView) GetSecond() Node { return sv.second }

// Children returns [first, second] (non-nil only) so the Inspector tree-walk
// and deepHasWheelConsumer reach both panes.
func (sv *SplitView) Children() []Node {
	var out []Node
	if sv.first != nil {
		out = append(out, sv.first)
	}
	if sv.second != nil {
		out = append(out, sv.second)
	}
	return out
}

// AddChild appends to the first or second slot in order of calls.
// The first AddChild call sets the first pane; the second call sets the second.
// Subsequent calls are ignored (SplitView supports exactly two panes).
func (sv *SplitView) AddChild(child Node) {
	if sv.first == nil {
		sv.SetFirst(child)
	} else if sv.second == nil {
		sv.SetSecond(child)
	}
	// More than 2 children: silently ignored — SplitView is a two-pane widget.
}

// IsInteractive returns true because the splitter bar handles drag input.
func (sv *SplitView) IsInteractive() bool { return true }

// ─── Geometry helpers ────────────────────────────────────────────────────────

// splitMetrics returns the pixel positions derived from the current Ratio
// and SplitterW for the given outer bounds.
//
// Returns (firstSize, splitterOrigin) where:
//   - firstSize     — px allocated to the first pane
//   - splitterOrigin — coordinate (X or Y) of the left/top edge of the splitter
//
// The second pane starts at splitterOrigin + SplitterW.
func (sv *SplitView) firstPaneHidden() bool {
	return sv.first != nil && sv.first.IsHidden()
}

func (sv *SplitView) effectiveSplitterW(b rl.Rectangle) float32 {
	if sv.firstPaneHidden() || (sv.second != nil && sv.second.IsHidden()) {
		return 0
	}
	var total float32
	if sv.Direction == SplitHorizontal {
		total = b.Width
	} else {
		total = b.Height
	}
	avail := total - sv.SplitterW
	if avail < 0 {
		avail = 0
	}
	if avail <= 0 {
		return 0
	}
	return sv.SplitterW
}

func (sv *SplitView) splitMetrics(b rl.Rectangle) (firstSize, splitterOrigin float32) {
	var total float32
	if sv.Direction == SplitHorizontal {
		total = b.Width
	} else {
		total = b.Height
	}
	splitW := sv.effectiveSplitterW(b)
	avail := total - splitW
	if avail < 0 {
		avail = 0
	}
	r := sv.clampRatio(sv.Ratio.Get(), avail)
	firstSize = r * avail
	if sv.Direction == SplitHorizontal {
		splitterOrigin = b.X + firstSize
	} else {
		splitterOrigin = b.Y + firstSize
	}
	return firstSize, splitterOrigin
}

// splitterRect returns the screen rectangle of the splitter bar.
func (sv *SplitView) splitterRect() rl.Rectangle {
	b := sv.Bounds()
	splitW := sv.effectiveSplitterW(b)
	if splitW <= 0 {
		return rl.Rectangle{}
	}
	_, origin := sv.splitMetrics(b)
	if sv.Direction == SplitHorizontal {
		return rl.NewRectangle(origin, b.Y, splitW, b.Height)
	}
	return rl.NewRectangle(b.X, origin, b.Width, splitW)
}

// hitRect returns a slightly larger rectangle around the splitter for easier
// mouse hit-testing without requiring pixel-perfect cursor placement.
func (sv *SplitView) hitRect() rl.Rectangle {
	r := sv.splitterRect()
	if sv.Direction == SplitHorizontal {
		return rl.NewRectangle(r.X-svHoverExtra, r.Y, r.Width+2*svHoverExtra, r.Height)
	}
	return rl.NewRectangle(r.X, r.Y-svHoverExtra, r.Width, r.Height+2*svHoverExtra)
}

// clampRatio clamps r to keep both panes at or above their minimum sizes.
// avail is the total splittable space (outer dimension minus SplitterW).
func (sv *SplitView) clampRatio(r, avail float32) float32 {
	if avail <= 0 {
		return 0.5
	}
	minR := sv.MinFirst / avail
	maxR := float32(1) - sv.MinSecond/avail
	if minR > maxR {
		minR = 0.5
		maxR = 0.5
	}
	if r < minR {
		r = minR
	}
	if r > maxR {
		r = maxR
	}
	return r
}

// firstRect returns the bounds for the first pane.
func (sv *SplitView) firstRect() rl.Rectangle {
	b := sv.Bounds()
	if sv.firstPaneHidden() {
		return rl.Rectangle{}
	}
	if sv.second != nil && sv.second.IsHidden() {
		return b
	}
	firstSize, _ := sv.splitMetrics(b)
	if sv.Direction == SplitHorizontal {
		return rl.NewRectangle(b.X, b.Y, firstSize, b.Height)
	}
	return rl.NewRectangle(b.X, b.Y, b.Width, firstSize)
}

// secondRect returns the bounds for the second pane.
func (sv *SplitView) secondRect() rl.Rectangle {
	b := sv.Bounds()
	if sv.firstPaneHidden() {
		return b
	}
	if sv.second != nil && sv.second.IsHidden() {
		return rl.Rectangle{}
	}
	_, origin := sv.splitMetrics(b)
	splitW := sv.effectiveSplitterW(b)
	if sv.Direction == SplitHorizontal {
		x2 := origin + splitW
		return rl.NewRectangle(x2, b.Y, b.Width-x2+b.X, b.Height)
	}
	y2 := origin + splitW
	return rl.NewRectangle(b.X, y2, b.Width, b.Height-y2+b.Y)
}

// ─── Update ──────────────────────────────────────────────────────────────────

// Update handles splitter drag input and delegates to both pane content nodes.
func (sv *SplitView) Update(dt float32) {
	if sv.IsHidden() {
		return
	}

	mouse := rl.GetMousePosition()
	hr := sv.hitRect()
	prevHover := sv.hoverSplit

	// ── Drag logic ────────────────────────────────────────────────────────────

	if sv.dragging {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			// Compute new ratio from mouse delta.
			b := sv.Bounds()
			var total, mouseCoord float32
			if sv.Direction == SplitHorizontal {
				total = b.Width
				mouseCoord = mouse.X
			} else {
				total = b.Height
				mouseCoord = mouse.Y
			}
			avail := total - sv.SplitterW
			if avail > 0 {
				delta := mouseCoord - sv.dragMouse0
				deltaR := delta / avail
				newR := sv.clampRatio(sv.dragRatio0+deltaR, avail)
				if newR != sv.Ratio.Get() {
					sv.Ratio.Set(newR) // triggers MarkDirty via subscriber
					if sv.OnRatioChanged != nil {
						sv.OnRatioChanged(newR)
					}
				}
			}
		} else {
			// Mouse released.
			sv.dragging = false
			rl.SetMouseCursor(rl.MouseCursorDefault)
		}
	}

	// ── Hover / click on splitter ─────────────────────────────────────────────

	if !sv.dragging {
		if sv.effectiveSplitterW(sv.Bounds()) <= 0 {
			sv.hoverSplit = false
		} else {
			sv.hoverSplit = rl.CheckCollisionPointRec(mouse, hr)
		}
		if sv.hoverSplit {
			if sv.Direction == SplitHorizontal {
				rl.SetMouseCursor(rl.MouseCursorResizeEW)
			} else {
				rl.SetMouseCursor(rl.MouseCursorResizeNS)
			}
			if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
				sv.dragging = true
				if sv.Direction == SplitHorizontal {
					sv.dragMouse0 = mouse.X
				} else {
					sv.dragMouse0 = mouse.Y
				}
				sv.dragRatio0 = sv.Ratio.Get()
			}
		} else if prevHover {
			// Leaving the splitter band — release resize cursor so borderless
			// chrome and other widgets can set I-beam / default on this frame.
			rl.SetMouseCursor(rl.MouseCursorDefault)
		}
	}

	if sv.hoverSplit != prevHover {
		sv.MarkDrawDirty()
	}

	// ── Delegate to children ──────────────────────────────────────────────────

	if sv.first != nil {
		sv.first.Update(dt)
	}
	if sv.second != nil {
		sv.second.Update(dt)
	}
}

// ─── Layout ──────────────────────────────────────────────────────────────────

// Layout recomputes the bounds of both pane contents from the current Ratio
// and recurses into any dirty child subtrees.
func (sv *SplitView) Layout() {
	if sv.IsHidden() {
		return
	}
	geom := !sv.lastLayoutPassValid || sv.lastLayoutPassW != sv.bounds.Width || sv.lastLayoutPassH != sv.bounds.Height
	if !sv.IsDirty() && !geom {
		return
	}
	// Always recompute child bounds from ratio (cheap).
	if sv.first != nil {
		fr := sv.firstRect()
		layoutChildAfterSetBounds(sv.first, fr)
		finalizeSplitPaneContainment(sv.first, fr)
	}
	if sv.second != nil {
		sr := sv.secondRect()
		layoutChildAfterSetBounds(sv.second, sr)
		finalizeSplitPaneContainment(sv.second, sr)
	}
	sv.layoutDirty = false
	sv.lastLayoutPassW = sv.bounds.Width
	sv.lastLayoutPassH = sv.bounds.Height
	sv.lastLayoutPassValid = true
}

// ─── Draw ────────────────────────────────────────────────────────────────────

// Draw renders the two panes and the splitter bar.
func (sv *SplitView) Draw() {
	defer func() { sv.drawDirty = false }()
	if sv.IsHidden() {
		return
	}

	if sv.first != nil {
		sv.drawPane(sv.first, sv.firstRect())
	}
	if sv.second != nil {
		sv.drawPane(sv.second, sv.secondRect())
	}

	sv.drawSplitter()
}

func (sv *SplitView) drawPane(n Node, pane rl.Rectangle) {
	if n == nil || pane.Width < 1 || pane.Height < 1 {
		return
	}
	beginScissorFromRect(pane)
	n.Draw()
	rl.EndScissorMode()
}

// drawSplitter renders a minimal 1px divider with a soft glow on hover/drag.
// Idle/hover colours come from "splitview-splitter".
func (sv *SplitView) drawSplitter() {
	if sv.effectiveSplitterW(sv.Bounds()) <= 0 {
		return
	}
	sr := sv.splitterRect()
	st := GetThemeStyle("splitview-splitter")
	lineCol := st.BorderColor
	if lineCol.A == 0 {
		lineCol = st.BackgroundColor
	}
	if lineCol.A == 0 {
		lineCol = rl.NewColor(190, 192, 205, 255)
	}

	active := sv.dragging || sv.hoverSplit
	if active {
		lineCol = lerpColor(lineCol, focusRingIndigo, 0.72)
		glow := lineCol
		if sv.dragging {
			glow.A = 72
		} else {
			glow.A = 48
		}
		rl.DrawRectangleRec(sr, glow)
	}

	if sv.Direction == SplitHorizontal {
		x := int32(sr.X + sr.Width/2)
		rl.DrawLine(x, int32(sr.Y+2), x, int32(sr.Y+sr.Height-2), lineCol)
	} else {
		y := int32(sr.Y + sr.Height/2)
		rl.DrawLine(int32(sr.X+2), y, int32(sr.X+sr.Width-2), y, lineCol)
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

// directionString returns a human-readable name for the split direction.
func (sv *SplitView) directionString() string {
	if sv.Direction == SplitHorizontal {
		return "horizontal"
	}
	return "vertical"
}

// firstPaneLabel returns a display label for the first pane slot.
func (sv *SplitView) firstPaneLabel() string {
	if sv.Direction == SplitHorizontal {
		return "left"
	}
	return "top"
}

// secondPaneLabel returns a display label for the second pane slot.
func (sv *SplitView) secondPaneLabel() string {
	if sv.Direction == SplitHorizontal {
		return "right"
	}
	return "bottom"
}

// SplitViewInfo returns display strings for Inspector use.
func (sv *SplitView) SplitViewInfo() (direction, ratioStr, sizeStr string) {
	b := sv.Bounds()
	avail := b.Width - sv.SplitterW
	if sv.Direction == SplitVertical {
		avail = b.Height - sv.SplitterW
	}
	r := sv.Ratio.Get()
	firstPx := r * avail
	secondPx := avail - firstPx
	direction = sv.directionString()
	ratioStr = fmt.Sprintf("%.3f  (%s=%.0f  %s=%.0f)",
		r, sv.firstPaneLabel(), firstPx, sv.secondPaneLabel(), secondPx)
	sizeStr = fmt.Sprintf("%.0f × %.0f  (avail=%.0f)", b.Width, b.Height, avail)
	return direction, ratioStr, sizeStr
}
