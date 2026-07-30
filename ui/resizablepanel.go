// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	rpDefaultSplitterW = float32(5)
	rpDefaultMinPane   = float32(48)
	rpHoverExtra       = float32(3)
	rpSnapRelease      = float32(0.035) // snap on mouse-up within 3.5% of a preset
	rpSnapMagnet       = float32(0.055) // pull boundary while dragging within 5.5%
	rpRatioQuantize    = float32(0.01)  // round displayed splits to 1%
)

// DefaultResizableSnapPoints are cumulative boundaries for magnetic snap (¼, ½, ¾).
var DefaultResizableSnapPoints = []float32{0.25, 0.5, 0.75}

// ResizablePanel divides space among three or more panes with draggable splitters.
// For two panes, prefer SplitView. Ratios are normalized panel fractions (sum = 1).
//
// Example:
//
//	rp := ui.NewResizablePanel("workspace", ui.SplitHorizontal,
//	    []ui.Node{tree, editor, props},
//	    []float32{0.22, 0.56, 0.22}, 0, 0, 0, 400)
//	rp.MinSizes = []float32{120, 200, 120}
type ResizablePanel struct {
	Element
	Direction SplitDirection
	Panels    []Node
	Ratios    []float32
	MinSizes  []float32
	MaxSizes  []float32 // optional max pane px; 0 = no cap
	SnapPoints []float32 // cumulative boundaries (0–1) to snap on release
	SplitterW float32

	OnRatiosChanged func(ratios []float32)

	dragging   int  // splitter index, -1 = none
	dragMouse0 float32
	dragRatios []float32
	hoverSplit int

	lastLayoutPassW     float32
	lastLayoutPassH     float32
	lastLayoutPassValid bool
}

// NewResizablePanel creates a multi-pane splitter. ratios may be nil for equal split.
func NewResizablePanel(id string, dir SplitDirection, panels []Node, ratios []float32, x, y, w, h float32) *ResizablePanel {
	if len(panels) < 2 {
		if len(panels) == 1 {
			panels = []Node{panels[0], NewLabel(id+"-pad", "", 0, 0, 0, 0)}
		} else {
			panels = []Node{
				NewLabel(id+"-a", "Pane A", 0, 0, 0, 0),
				NewLabel(id+"-b", "Pane B", 0, 0, 0, 0),
			}
		}
	}
	rp := &ResizablePanel{
		Element:   NewElement(id, x, y, w, h),
		Direction: dir,
		Panels:    panels,
		SplitterW: rpDefaultSplitterW,
		hoverSplit: -1,
		dragging:  -1,
	}
	rp.styleName = "resizablepanel"
	rp.Ratios = normalizeRatios(ratios, len(panels))
	rp.dragRatios = append([]float32(nil), rp.Ratios...)
	rp.SnapPoints = append([]float32(nil), DefaultResizableSnapPoints...)
	for _, p := range panels {
		if p != nil {
			p.SetParent(rp)
		}
	}
	return rp
}

func normalizeRatios(ratios []float32, n int) []float32 {
	out := make([]float32, n)
	if len(ratios) == n {
		copy(out, ratios)
	} else {
		for i := range out {
			out[i] = 1 / float32(n)
		}
	}
	var sum float32
	for _, r := range out {
		sum += r
	}
	if sum <= 0 {
		for i := range out {
			out[i] = 1 / float32(n)
		}
		return out
	}
	for i := range out {
		out[i] /= sum
	}
	return out
}

// Children implements Node.
func (rp *ResizablePanel) Children() []Node { return rp.Panels }

// AddChild appends a pane and rebalances ratios equally.
func (rp *ResizablePanel) AddChild(child Node) {
	if child == nil {
		return
	}
	rp.Panels = append(rp.Panels, child)
	child.SetParent(rp)
	rp.Ratios = normalizeRatios(nil, len(rp.Panels))
	rp.dragRatios = append([]float32(nil), rp.Ratios...)
	rp.MarkDirty()
}

func (rp *ResizablePanel) paneCount() int { return len(rp.Panels) }

func (rp *ResizablePanel) splitterCount() int {
	n := rp.paneCount()
	if n <= 1 {
		return 0
	}
	return n - 1
}

func (rp *ResizablePanel) availSize(b rl.Rectangle) float32 {
	var total float32
	if rp.Direction == SplitHorizontal {
		total = b.Width
	} else {
		total = b.Height
	}
	splitters := float32(rp.splitterCount()) * rp.SplitterW
	avail := total - splitters
	if avail < 1 {
		return 1
	}
	return avail
}

func (rp *ResizablePanel) paneRect(i int) rl.Rectangle {
	b := rp.Bounds()
	sizes := rp.paneSizes()
	off := rp.paneOffsetPx(i, sizes)
	if rp.Direction == SplitHorizontal {
		return rl.NewRectangle(b.X+off, b.Y, sizes[i], b.Height)
	}
	return rl.NewRectangle(b.X, b.Y+off, b.Width, sizes[i])
}

func (rp *ResizablePanel) splitterRect(i int) rl.Rectangle {
	b := rp.Bounds()
	sizes := rp.paneSizes()
	off := rp.paneOffsetPx(i, sizes) + sizes[i]
	if rp.Direction == SplitHorizontal {
		return rl.NewRectangle(b.X+off, b.Y, rp.SplitterW, b.Height)
	}
	return rl.NewRectangle(b.X, b.Y+off, b.Width, rp.SplitterW)
}

func (rp *ResizablePanel) splitterHitRect(i int) rl.Rectangle {
	r := rp.splitterRect(i)
	if rp.Direction == SplitHorizontal {
		return rl.NewRectangle(r.X-rpHoverExtra, r.Y, r.Width+2*rpHoverExtra, r.Height)
	}
	return rl.NewRectangle(r.X, r.Y-rpHoverExtra, r.Width, r.Height+2*rpHoverExtra)
}

func (rp *ResizablePanel) minSize(i int) float32 {
	if i >= 0 && i < len(rp.MinSizes) && rp.MinSizes[i] > 0 {
		return rp.MinSizes[i]
	}
	return rpDefaultMinPane
}

func (rp *ResizablePanel) maxRatio(i int, avail float32) float32 {
	if avail <= 0 {
		return 1
	}
	if i >= 0 && i < len(rp.MaxSizes) && rp.MaxSizes[i] > 0 {
		return rp.MaxSizes[i] / avail
	}
	return 1
}

func (rp *ResizablePanel) clampRatios() {
	avail := rp.availSize(rp.Bounds())
	if avail <= 0 {
		return
	}
	n := len(rp.Ratios)
	mins := make([]float32, n)
	var minPxSum float32
	for i := range rp.Ratios {
		mins[i] = rp.minSize(i)
		minPxSum += mins[i]
	}
	// When minimums exceed available space, scale proportionally (no jamming).
	if minPxSum >= avail-0.5 {
		for i := range rp.Ratios {
			rp.Ratios[i] = mins[i] / minPxSum
		}
		return
	}
	for i := range rp.Ratios {
		mins[i] = mins[i] / avail
	}
	for i := range rp.Ratios {
		maxR := rp.maxRatio(i, avail)
		if rp.Ratios[i] > maxR {
			excess := rp.Ratios[i] - maxR
			rp.Ratios[i] = maxR
			if j := rp.largestRatioIndexExcept(i, mins); j >= 0 {
				rp.Ratios[j] += excess
			}
		}
	}
	for i := range rp.Ratios {
		if rp.Ratios[i] >= mins[i] {
			continue
		}
		need := mins[i] - rp.Ratios[i]
		rp.Ratios[i] = mins[i]
		for need > 0.0001 {
			j := rp.largestRatioIndexExcept(i, mins)
			if j < 0 {
				break
			}
			canTake := rp.Ratios[j] - mins[j]
			if canTake <= 0 {
				break
			}
			take := need
			if take > canTake {
				take = canTake
			}
			rp.Ratios[j] -= take
			need -= take
		}
	}
	normalizeRatiosSlice(rp.Ratios)
}

func normalizeRatiosSlice(ratios []float32) {
	var sum float32
	for _, r := range ratios {
		sum += r
	}
	if sum > 0 && (sum < 0.999 || sum > 1.001) {
		for i := range ratios {
			ratios[i] /= sum
		}
	}
}

// paneSizes returns pixel pane sizes; the last pane absorbs rounding remainder.
func (rp *ResizablePanel) paneSizes() []float32 {
	avail := rp.availSize(rp.Bounds())
	n := len(rp.Ratios)
	out := make([]float32, n)
	if n == 0 || avail <= 0 {
		return out
	}
	var used float32
	for i := 0; i < n-1; i++ {
		out[i] = rp.Ratios[i] * avail
		used += out[i]
	}
	out[n-1] = avail - used
	if out[n-1] < 0 {
		out[n-1] = 0
	}
	return out
}

func (rp *ResizablePanel) paneOffsetPx(i int, sizes []float32) float32 {
	var off float32
	for j := 0; j < i; j++ {
		off += sizes[j] + rp.SplitterW
	}
	return off
}

func (rp *ResizablePanel) quantizeRatios() {
	if rpRatioQuantize <= 0 {
		return
	}
	for i := range rp.Ratios {
		q := float32(int32(rp.Ratios[i]/rpRatioQuantize+0.5)) * rpRatioQuantize
		if q < 0 {
			q = 0
		}
		rp.Ratios[i] = q
	}
	var sum float32
	for _, r := range rp.Ratios {
		sum += r
	}
	if sum > 0 && (sum < 0.999 || sum > 1.001) {
		normalizeRatiosSlice(rp.Ratios)
	}
}

func (rp *ResizablePanel) cumulativeThrough(splitter int) float32 {
	var cum float32
	for j := 0; j <= splitter && j < len(rp.Ratios); j++ {
		cum += rp.Ratios[j]
	}
	return cum
}

func (rp *ResizablePanel) snapBoundary(splitter int, epsilon float32) bool {
	if splitter < 0 || len(rp.SnapPoints) == 0 {
		return false
	}
	cum := rp.cumulativeThrough(splitter)
	bestDist := float32(1)
	var best float32
	for _, sp := range rp.SnapPoints {
		d := cum - sp
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			best = sp
		}
	}
	if bestDist >= epsilon {
		return false
	}
	delta := best - cum
	rp.Ratios[splitter] += delta
	rp.Ratios[splitter+1] -= delta
	rp.clampRatios()
	return true
}

func (rp *ResizablePanel) largestRatioIndexExcept(skip int, mins []float32) int {
	best := -1
	var bestVal float32
	for i, r := range rp.Ratios {
		if i == skip {
			continue
		}
		flex := r - mins[i]
		if flex <= 0 {
			continue
		}
		if best < 0 || r > bestVal {
			best = i
			bestVal = r
		}
	}
	return best
}

func (rp *ResizablePanel) notifyRatiosChanged() {
	if rp.OnRatiosChanged != nil {
		rp.OnRatiosChanged(append([]float32(nil), rp.Ratios...))
	}
}

func (rp *ResizablePanel) applySnap() {
	if rp.dragging < 0 {
		return
	}
	rp.snapBoundary(rp.dragging, rpSnapRelease)
	rp.quantizeRatios()
	rp.clampRatios()
}

// IsInteractive implements Node.
func (rp *ResizablePanel) IsInteractive() bool { return rp.splitterCount() > 0 }

// Layout assigns pane bounds from current ratios.
func (rp *ResizablePanel) Layout() {
	if rp.IsHidden() {
		return
	}
	geom := !rp.lastLayoutPassValid || rp.lastLayoutPassW != rp.bounds.Width || rp.lastLayoutPassH != rp.bounds.Height
	if !rp.IsDirty() && !geom {
		return
	}
	rp.clampRatios()
	for i, p := range rp.Panels {
		if p == nil {
			continue
		}
		pr := rp.paneRect(i)
		layoutChildAfterSetBounds(p, pr)
		finalizeSplitPaneContainment(p, pr)
	}
	rp.layoutDirty = false
	rp.lastLayoutPassW = rp.bounds.Width
	rp.lastLayoutPassH = rp.bounds.Height
	rp.lastLayoutPassValid = true
}

// Update handles splitter drag and pane updates.
func (rp *ResizablePanel) Update(dt float32) {
	if rp.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	prevHover := rp.hoverSplit
	rp.hoverSplit = -1
	for i := 0; i < rp.splitterCount(); i++ {
		if rl.CheckCollisionPointRec(mouse, rp.splitterHitRect(i)) {
			rp.hoverSplit = i
			break
		}
	}
	if rp.hoverSplit != prevHover {
		rp.MarkDrawDirty()
	}

	if rp.dragging >= 0 {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			b := rp.Bounds()
			avail := rp.availSize(b)
			var coord float32
			if rp.Direction == SplitHorizontal {
				coord = mouse.X - b.X
			} else {
				coord = mouse.Y - b.Y
			}
			delta := (coord - rp.dragMouse0) / avail
			i := rp.dragging
			prevLeft, prevRight := rp.Ratios[i], rp.Ratios[i+1]
			rp.Ratios[i] = rp.dragRatios[i] + delta
			rp.Ratios[i+1] = rp.dragRatios[i+1] - delta
			rp.clampRatios()
			rp.snapBoundary(i, rpSnapMagnet)
			if rp.Ratios[i] != prevLeft || rp.Ratios[i+1] != prevRight {
				rp.MarkDirty()
				rp.MarkDrawDirty()
				rp.notifyRatiosChanged()
			}
		} else {
			rp.applySnap()
			rp.dragging = -1
			rp.notifyRatiosChanged()
		}
	} else if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rp.hoverSplit >= 0 {
		rp.dragging = rp.hoverSplit
		rp.dragRatios = append([]float32(nil), rp.Ratios...)
		b := rp.Bounds()
		if rp.Direction == SplitHorizontal {
			rp.dragMouse0 = mouse.X - b.X
		} else {
			rp.dragMouse0 = mouse.Y - b.Y
		}
	}

	for _, p := range rp.Panels {
		if p != nil {
			p.Update(dt)
		}
	}
}

// Draw implements Node.Draw.
func (rp *ResizablePanel) Draw() { rp.drawInternal() }

func (rp *ResizablePanel) drawInternal() {
	if rp.IsHidden() {
		return
	}
	style := rp.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(rp.Bounds(), style.BackgroundColor)
	}
	for _, p := range rp.Panels {
		if p != nil {
			p.Draw()
		}
	}
	splitStyle := GetThemeStyle("splitview-splitter")
	for i := 0; i < rp.splitterCount(); i++ {
		r := rp.splitterRect(i)
		col := splitStyle.BackgroundColor
		if i == rp.hoverSplit || rp.dragging == i {
			col = rl.NewColor(79, 70, 229, 200)
		}
		rl.DrawRectangleRec(r, col)
	}
}
