package examples

import "github.com/ledocorp/gru/ui"

// NewBatchPageGrid returns the standard shrink-wrap 12-column grid for batch demos
// (docs/LAYOUT_CONTRACTS.md §7). Panels use intrinsic height; rows do not equal-fill
// the viewport unless you opt into NewBatchEqualFillGrid.
func NewBatchPageGrid(id string, gap float32) *ui.Container {
	grid := ui.NewContainer(id, 0, 0, 0, 0)
	grid.LayoutType = ui.LayoutGrid
	grid.GridColumns = 12
	grid.Gap = gap
	grid.SetStyle("page-grid")
	grid.GridRowSizing = ui.GridRowSizingShrinkWrap
	grid.SetFlexGrow(1)
	return grid
}

// NewBatchEqualFillGrid fills the viewport band with equal-height rows — use only
// when the demo intentionally fills one screen with no page scroll (e.g. a single
// full-viewport fixture). Split showcases with multiple sections should use
// NewBatchPageGrid + fixed section heights (see batch8_demo.go).
func NewBatchEqualFillGrid(id string, gap float32) *ui.Container {
	grid := NewBatchPageGrid(id, gap)
	grid.AutoHeight = false
	grid.GridRowSizing = ui.GridRowSizingEqualFill
	return grid
}
