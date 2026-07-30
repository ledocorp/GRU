// Package ui (continued) — viewport float overlay host.
//
// FloatLayer is a transparent absolute container that Viewport lays out over the
// visible client area (not scrolled). Add movable/resizable panels as children.
//
// # LLM Prompt Template
//
//	fl := ui.NewFloatLayer("tools")
//	fl.SetZIndex(500)
//	win := ui.NewPanel("palette", "Tools", 24, 0, 320, 280)
//	win.SetFloatPosition(24, 80).SetMovable(true).SetResizable(true)
//	fl.AddChild(win)
//	vp.AddChild(fl) // last child, high ZIndex
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// FloatOverlay marks a node hosted above viewport scroll content.
type FloatOverlay interface {
	Node
	IsFloatOverlay() bool
}

// FloatLayer hosts floating panels above a Viewport's scrollable content.
// Viewport pins the layer to floatOverlayHostRect (full viewport bounds) each layout pass.
type FloatLayer struct {
	Container
}

// NewFloatLayer creates a transparent absolute overlay host.
func NewFloatLayer(id string) *FloatLayer {
	c := NewContainer(id, 0, 0, 0, 0)
	fl := &FloatLayer{Container: *c}
	fl.LayoutType = LayoutAbsolute
	fl.SetStyle("transparent")
	return fl
}

// IsFloatOverlay reports that Viewport should pin this layer to the client band.
func (fl *FloatLayer) IsFloatOverlay() bool { return true }

// Layout runs child layout; bounds are assigned by the parent Viewport.
func (fl *FloatLayer) Layout() {
	if fl.IsHidden() {
		return
	}
	for _, ch := range fl.children {
		if !ch.IsHidden() {
			ch.Layout()
		}
	}
	fl.rebuildSortedCache()
	fl.layoutDirty = false
	fl.lastLayoutPassW = fl.bounds.Width
	fl.lastLayoutPassH = fl.bounds.Height
	fl.lastLayoutPassValid = true
	syncLayoutExtent(fl)
}

func nodeIsFloatOverlay(n Node) bool {
	if n == nil {
		return false
	}
	fo, ok := n.(FloatOverlay)
	return ok && fo.IsFloatOverlay()
}

func collectFloatOverlays(nodes []Node) []Node {
	var out []Node
	for _, child := range nodes {
		if child.IsHidden() {
			continue
		}
		if nodeIsFloatOverlay(child) {
			out = append(out, child)
		}
	}
	return out
}

func (vp *Viewport) layoutFloatOverlays() {
	rect := vp.floatOverlayHostRect()
	for _, child := range vp.children {
		if child.IsHidden() || !nodeIsFloatOverlay(child) {
			continue
		}
		type boundsWriter interface{ setBoundsNoMark(rl.Rectangle) }
		if bw, ok := child.(boundsWriter); ok {
			cur := child.Bounds()
			if cur != rect {
				bw.setBoundsNoMark(rect)
				child.MarkDirty()
			}
		} else {
			child.SetBounds(rect)
		}
	}
}
