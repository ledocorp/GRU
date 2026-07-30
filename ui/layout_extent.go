// Package ui — layout extent measurement and upward propagation (see docs/LAYOUT_CONTRACTS.md §5).
package ui

import "math"

// nodeLayoutExtentHeight returns the vertical span n occupies for grid row sizing
// and AutoHeight parents. Intrinsic hosts use subtree bottom; fixed/fill use bounds.
func nodeLayoutExtentHeight(n Node) float32 {
	if n.IsHidden() {
		return 0
	}
	b := n.Bounds()
	if !n.IsAutoHeight() || n.GetFlexGrow() > 0 {
		if b.Height > 0 {
			return b.Height
		}
		return 0
	}
	bottom := nodeSubtreeBottom(n)
	h := bottom - b.Y
	if h < b.Height {
		return b.Height
	}
	return h
}

// syncLayoutExtent records subtree bottom after Layout and marks ancestors dirty
// when extent changed. Skips the first record and changes smaller than 0.5 px.
func syncLayoutExtent(n Node) {
	if n == nil || n.IsHidden() {
		return
	}
	bottom := nodeSubtreeBottom(n)
	if recorder, ok := n.(layoutExtentRecorder); ok {
		recorder.recordLayoutExtentBottom(bottom)
	}
}

type layoutExtentRecorder interface {
	recordLayoutExtentBottom(bottom float32)
}

func (e *Element) recordLayoutExtentBottom(bottom float32) {
	if e.layoutExtentValid && math.Abs(float64(bottom-e.layoutExtentBottom)) > 0.5 {
		if e.parent != nil {
			e.parent.MarkDirty()
		}
	}
	e.layoutExtentBottom = bottom
	e.layoutExtentValid = true
}
