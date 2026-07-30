package ui

// NewPageGrid returns a full-size Container using LayoutGrid with 12 columns.
//
// This is the engine primitive for a **responsive page grid layout** (the
// top-level composition pattern): place it as the first child under an
// absolute-sized Document root, then add cells (Viewports, Panels, etc.).
// The layout type constant is still named LayoutGrid — "page grid" refers to
// this documented pattern, not a separate engine.
//
// Typical stack: Root (LayoutAbsolute) → page grid → Viewport(s) → Panel → flex.
// See ARCHITECTURE.md §5.6 and examples.MountPageGrid / MountFlexPageShell.
func NewPageGrid(id string, width, height float32) *Container {
	c := NewContainer(id, 0, 0, width, height)
	c.LayoutType = LayoutGrid
	c.GridColumns = 12
	c.Gap = 12
	c.SetStyle("page-grid")
	return c
}
