// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─────────────────────────────────────────────────────────────────────────────
// Column types
// ─────────────────────────────────────────────────────────────────────────────

// ColumnAlign controls horizontal text alignment within a data cell.
type ColumnAlign int

const (
	ColumnAlignLeft   ColumnAlign = iota // Text starts at the left edge of the cell
	ColumnAlignCenter                    // Text is centred in the cell
	ColumnAlignRight                     // Text is right-aligned in the cell
)

// Column defines a single column of a DataTable.
// T is the row data type — it must match the DataTable's type parameter.
//
// At minimum, set Title, Width, and Render. Set Sortable + SortLess to enable
// header-click sorting for this column. Use CellDraw instead of Render when
// the cell needs custom graphics (e.g. a badge, progress bar, or icon).
type Column[T any] struct {
	// Title is the column header label.
	Title string

	// Width is the column's fixed pixel width. Must be > 0.
	Width float32

	// Align controls horizontal text alignment in data cells. Default: Left.
	Align ColumnAlign

	// Sortable, when true, allows the user to click the header to sort by
	// this column. Requires SortLess to be non-nil.
	Sortable bool

	// SortLess is a less-than comparator used by sort.SliceStable when this
	// column is active. Return true when a should appear before b.
	// If nil the column is treated as non-sortable.
	SortLess func(a, b T) bool

	// Render returns the display string for a cell. Called when CellDraw is
	// nil. If both Render and CellDraw are nil the cell is drawn empty.
	Render func(item T) string

	// CellDraw, when non-nil, fully replaces Render. It must paint within
	// bounds — a scissor region covering the entire data area is already
	// active; content drawn outside bounds may overlap adjacent cells.
	CellDraw func(item T, rowIndex int, isSelected bool, bounds rl.Rectangle)
}

// ─────────────────────────────────────────────────────────────────────────────
// Layout constants
// ─────────────────────────────────────────────────────────────────────────────

const (
	dtHeaderH      float32 = 36
	dtScrollbarW   float32 = 13
	dtScrollbarH   float32 = 13
	dtSortArrowW   float32 = 16
	dtCellPadX     float32 = 16
	dtCellFontSize int32   = 15
	dtHeaderFS     int32   = 15
	dtResizeHandleW float32 = 6
	dtMinColW       float32 = 48
)

// ─────────────────────────────────────────────────────────────────────────────
// Palette
// ─────────────────────────────────────────────────────────────────────────────

var (
	dtHeaderBg         = rl.NewColor(247, 248, 252, 255)
	dtHeaderBorderCol  = rl.NewColor(218, 220, 232, 255)
	dtColSepCol        = rl.NewColor(220, 222, 234, 255)
	dtRowBg            = rl.NewColor(255, 255, 255, 255)
	dtAltRowBg         = rl.NewColor(249, 250, 253, 255)
	dtHoverOverlayBg   = rl.NewColor(99, 102, 241, 42) // translucent overlay (no SSAA bust)
	dtSelectedRowBg    = rl.NewColor(224, 228, 253, 255)
	dtRowSepCol        = rl.NewColor(242, 244, 249, 255)
	dtCellText         = rl.NewColor(31, 41, 55, 255)
	dtSelectedText     = rl.NewColor(30, 27, 75, 255)
	dtSortedHeaderText = rl.NewColor(79, 70, 229, 255)
	dtScrollTrack      = rl.NewColor(234, 236, 242, 255)
	dtScrollThumb      = rl.NewColor(162, 168, 190, 255)
)

// ─────────────────────────────────────────────────────────────────────────────
// dataTableWidget — non-generic Inspector interface
// ─────────────────────────────────────────────────────────────────────────────

// dataTableWidget is implemented by every DataTable[T] regardless of T.
// It lets the Inspector display row count, column count, and sort state
// without knowing the concrete row type.
type dataTableWidget interface {
	Node
	dtInfo() (colCount, rowCount, sortCol int, sortAsc bool, selectedRow int)
}

// ─────────────────────────────────────────────────────────────────────────────
// DataTable
// ─────────────────────────────────────────────────────────────────────────────

// DataTable is a high-performance virtual table with sortable column headers,
// row selection, and horizontal scrolling.
//
// # Virtualization
//
// Only the rows whose Y range overlaps the visible window are drawn each
// frame. With thousands of rows at RowHeight=30, at most ~20 rows are
// rendered per frame regardless of total data size.
//
// # Sorting
//
// Clicking a sortable column header cycles through ascending → descending →
// (clicking a different column resets direction). The sort is applied to an
// index slice (sortIndices) so the original binding order is never mutated.
// The binding's selected index always refers to the original binding order.
//
// # Scrolling
//
// Vertical:   mouse wheel when the cursor is over the table (not over other panes).
// Horizontal: Shift + wheel, or wheel on the horizontal scrollbar track.
// Horizontal wheel uses the opposite sign from vertical so the whole grid pans
// left as you scroll down (new columns enter from the right).
//
// # Scissor clipping
//
// DataTable applies scissors so the header band and row band clip separately
// (rows never paint over the fixed header), intersected with the parent
// Viewport's ClipBounds() so cells never escape the enclosing panel. Scrollbars
// are drawn in a separate scissor pass.
//
// # Usage
//
//	type Person struct { Name, City string; Age int }
//	binding := ui.NewListBinding(people)
//	cols := []ui.Column[Person]{
//	    {Title: "Name", Width: 160, Sortable: true,
//	        SortLess: func(a, b Person) bool { return a.Name < b.Name },
//	        Render: func(p Person) string { return p.Name }},
//	    {Title: "City", Width: 130,
//	        Render: func(p Person) string { return p.City }},
//	    {Title: "Age", Width: 70, Align: ui.ColumnAlignRight,
//	        Sortable: true,
//	        SortLess: func(a, b Person) bool { return a.Age < b.Age },
//	        Render: func(p Person) string { return fmt.Sprintf("%d", p.Age) }},
//	}
//	table := ui.NewDataTable("people-table", cols, binding, 0, 0, 0, 400)
//	panel.AddChild(table)
//
// # LLM Prompt Template
//
//	binding := ui.NewListBinding(rows)
//	cols := []ui.Column[Row]{{Title: "Name", Width: 160, Render: func(r Row) string { return r.Name }}}
//	table := ui.NewDataTable("table", cols, binding, 0, 0, 0, 340)
//	panel.AddChild(table)
//
// Demo scenes: **Batch 6 DataTable**, **Inbox Demo**.
type DataTable[T any] struct {
	Element

	// Columns is the ordered list of column definitions. Set at construction.
	Columns []Column[T]

	// RowHeight is the pixel height of each data row. Default: 30.
	RowHeight float32

	// CellFontSize overrides data cell text size (0 = default 15).
	CellFontSize int32

	// HeaderFontSize overrides column header text size (0 = default 15).
	HeaderFontSize int32

	// HeaderHeight overrides the column header band height (0 = derived from HeaderFontSize).
	HeaderHeight float32

	// ShowAlternatingRows draws a faint background on odd-indexed rows.
	ShowAlternatingRows bool

	// OnRowClick fires with the binding index when the user clicks a row.
	OnRowClick func(bindingIndex int)

	// ResizableColumns enables Excel-style drag handles on column header edges.
	ResizableColumns bool

	// binding holds the data and selection state.
	binding *ListBinding[T]

	// colWidths holds live column widths (initialized from Columns[].Width).
	colWidths []float32

	// Column resize drag.
	resizingCol      int
	resizeDragMouseX float32
	resizeDragStartW float32
	hoverResizeCol   int

	// Scroll state.
	scrollY float32
	scrollX float32

	// Computed content dimensions (updated in Layout).
	contentH float32
	contentW float32

	// Sort state.
	sortCol     int   // active sort column index; -1 = no sort
	sortAsc     bool  // true = ascending
	sortIndices []int // display row i → binding index sortIndices[i]

	// Hover state (display row index, or -1).
	hoverRow       int
	hoverHeaderCol int

	// Scrollbar drag (thumb or track jump).
	sbDragging      bool
	sbDragVert      bool
	sbDragStartMouse float32 // X for horizontal, Y for vertical
	sbDragStartScroll float32
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// NewDataTable creates a DataTable.
//
//	id      — unique node ID
//	cols    — column definitions (Title, Width, Render, … required per column)
//	binding — reactive data + selection source; call binding.SetItems to refresh
//	x, y    — position (overridden by parent layout)
//	w       — width; 0 = flex cross-axis stretch from parent Container
//	h       — height; required (the table has no natural height)
func NewDataTable[T any](id string, cols []Column[T], binding *ListBinding[T], x, y, w, h float32) *DataTable[T] {
	dt := &DataTable[T]{
		Element:             NewElement(id, x, y, w, h),
		Columns:             cols,
		binding:             binding,
		RowHeight:           36,
		ShowAlternatingRows: true,
		sortCol:             -1,
		sortAsc:             true,
		hoverRow:            -1,
		hoverHeaderCol:      -1,
		resizingCol:         -1,
		hoverResizeCol:      -1,
	}
	dt.styleName = "datatable"
	dt.colWidths = make([]float32, len(cols))
	for i, col := range cols {
		dt.colWidths[i] = col.Width
	}

	// Initial sort index (identity).
	dt.rebuildSortIndices()

	// Rebuild when data changes; re-sort maintains the active sort column.
	binding.SubscribeItems(func() {
		dt.rebuildSortIndices()
		dt.MarkDirty()
	})
	// Selection change needs only a redraw (row highlight changes).
	binding.SubscribeSelection(func() {
		dt.MarkDrawDirty()
	})

	return dt
}

// scrollGeometry resolves visible data area and scrollbar thickness. Vertical
// and horizontal bars affect each other's client size (classic corner case).
func (dt *DataTable[T]) scrollGeometry() (dataAreaW, dataAreaH, vertSBW, horizSBH float32) {
	b := dt.bounds
	vertSBW, horizSBH = 0, 0
	for i := 0; i < 8; i++ {
		dataAreaW = b.Width - vertSBW
		dataAreaH = b.Height - dt.headerHeight() - horizSBH
		if dataAreaH < 1 {
			dataAreaH = 1
		}
		if dataAreaW < 1 {
			dataAreaW = 1
		}
		nv := float32(0)
		if dt.contentH > dataAreaH {
			nv = dtScrollbarW
		}
		nh := float32(0)
		if dt.contentW > dataAreaW {
			nh = dtScrollbarH
		}
		if nv == vertSBW && nh == horizSBH {
			break
		}
		vertSBW, horizSBH = nv, nh
	}
	dataAreaW = b.Width - vertSBW
	dataAreaH = b.Height - dt.headerHeight() - horizSBH
	if dataAreaH < 1 {
		dataAreaH = 1
	}
	if dataAreaW < 1 {
		dataAreaW = 1
	}
	return dataAreaW, dataAreaH, vertSBW, horizSBH
}

// horizThumbGeom reports track Y, thumb X/width, and max horizontal scroll.
func (dt *DataTable[T]) horizThumbGeom(b rl.Rectangle, dataAreaW float32) (trackY, thumbX, thumbW, maxScroll float32) {
	trackY = b.Y + b.Height - dtScrollbarH
	maxScroll = dt.contentW - dataAreaW
	if maxScroll < 0 {
		maxScroll = 0
	}
	if dt.contentW <= dataAreaW || dataAreaW <= 0 {
		thumbX = b.X
		thumbW = dataAreaW
		return trackY, thumbX, thumbW, maxScroll
	}
	thumbW = (dataAreaW / dt.contentW) * dataAreaW
	if thumbW < 20 {
		thumbW = 20
	}
	if maxScroll > 0 {
		thumbX = b.X + (dt.scrollX/maxScroll)*(dataAreaW-thumbW)
	} else {
		thumbX = b.X
	}
	return trackY, thumbX, thumbW, maxScroll
}

// vertThumbGeom reports track X, thumb Y/height, and max vertical scroll.
func (dt *DataTable[T]) vertThumbGeom(b rl.Rectangle, dataAreaH float32) (trackX, thumbY, thumbH, maxScroll float32) {
	trackX = b.X + b.Width - dtScrollbarW
	maxScroll = dt.contentH - dataAreaH
	if maxScroll < 0 {
		maxScroll = 0
	}
	trackH := dataAreaH
	if trackH < 1 {
		trackH = 1
	}
	if dt.contentH <= dataAreaH || dataAreaH <= 0 {
		thumbY = b.Y + dt.headerHeight()
		thumbH = trackH
		return trackX, thumbY, thumbH, maxScroll
	}
	thumbH = (dataAreaH / dt.contentH) * trackH
	if thumbH < 20 {
		thumbH = 20
	}
	if maxScroll > 0 {
		thumbY = b.Y + dt.headerHeight() + (dt.scrollY/maxScroll)*(trackH-thumbH)
	} else {
		thumbY = b.Y + dt.headerHeight()
	}
	return trackX, thumbY, thumbH, maxScroll
}

// ─────────────────────────────────────────────────────────────────────────────
// Node interface
// ─────────────────────────────────────────────────────────────────────────────

// IsInteractive returns true — DataTable handles mouse input.
func (dt *DataTable[T]) IsInteractive() bool { return true }

// UsesScissor returns true — DataTable uses BeginScissorMode in Draw.
func (dt *DataTable[T]) UsesScissor() bool { return true }

// HandlesWheelScroll returns true so a parent Viewport defers wheel events
// to the table instead of scrolling the whole panel.
func (dt *DataTable[T]) HandlesWheelScroll() bool { return true }

// AbsorbsParentWheel implements wheelScrollLimiter. Shift+Wheel belongs to the
// table when horizontal overflow exists; otherwise the outer page viewport can
// consume it before DataTable.Update gets a chance to pan columns.
func (dt *DataTable[T]) AbsorbsParentWheel(wheel float32) bool {
	dataAreaW, dataAreaH, _, horizSBH := dt.scrollGeometry()
	needsH := horizSBH > 0
	needsV := dt.contentH > dataAreaH
	maxX := dt.contentW - dataAreaW
	if maxX < 0 {
		maxX = 0
	}
	// Wide tables with no vertical overflow: keep page scroll from stealing wheel.
	if needsH && !needsV && maxX > 0 {
		return true
	}
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	if shift && needsH && maxX > 0 {
		const eps = float32(0.5)
		if wheel < 0 && dt.scrollX >= maxX-eps {
			return false
		}
		if wheel > 0 && dt.scrollX <= eps {
			return false
		}
		return true
	}
	if !needsV {
		return false
	}
	maxY := dt.contentH - dataAreaH
	const eps = float32(0.5)
	if wheel < 0 && dt.scrollY >= maxY-eps {
		return false
	}
	if wheel > 0 && dt.scrollY <= eps {
		return false
	}
	return true
}

// Layout recomputes content dimensions and scroll clamping.
func (dt *DataTable[T]) Layout() {
	dataAreaW, _, _, _ := dt.scrollGeometry()
	dt.contentW = dt.totalColumnsWidth()
	if extra := dataAreaW - dt.contentW; extra > 0.5 {
		dt.contentW = dataAreaW
	}
	n := dt.binding.Len()
	if len(dt.sortIndices) != n {
		dt.rebuildSortIndices()
	}
	dt.contentH = float32(n) * dt.RowHeight
	dt.clampScrollY()
	dt.clampScrollX()
	dt.layoutDirty = false
}

// Update handles wheel scroll, hover detection, column header clicks, and row
// selection.
func (dt *DataTable[T]) Update(_ float32) {
	if dt.IsHidden() {
		return
	}
	b := dt.bounds
	mouse := rl.GetMousePosition()
	inBounds := rl.CheckCollisionPointRec(mouse, b)

	dataAreaW, dataAreaH, vertSBW, horizSBH := dt.scrollGeometry()
	needsVScroll := vertSBW > 0
	needsHScroll := horizSBH > 0
	hasHorizBar := needsHScroll
	contentOverflowY := dt.contentH > dataAreaH

	horizTrackRect := rl.NewRectangle(b.X, b.Y+b.Height-dtScrollbarH, dataAreaW, dtScrollbarH)
	vertTrackRect := rl.NewRectangle(b.X+b.Width-dtScrollbarW, b.Y+dt.headerHeight(), dtScrollbarW, dataAreaH)
	trackY, thumbX, thumbW, maxHScroll := dt.horizThumbGeom(b, dataAreaW)

	// ── Column resize drag ────────────────────────────────────────────────────
	if dt.resizingCol >= 0 {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			delta := mouse.X - dt.resizeDragMouseX
			nw := dt.resizeDragStartW + delta
			if nw < dtMinColW {
				nw = dtMinColW
			}
			if dt.resizingCol < len(dt.colWidths) {
				dt.colWidths[dt.resizingCol] = nw
				dt.Columns[dt.resizingCol].Width = nw
			}
			dt.MarkDirty()
		} else {
			dt.resizingCol = -1
		}
	}

	// ── Scrollbar drag (thumb) ───────────────────────────────────────────────
	if dt.sbDragging {
		if !rl.IsMouseButtonDown(rl.MouseLeftButton) {
			dt.sbDragging = false
		} else {
			if dt.sbDragVert {
				_, _, vThumbH, vMax := dt.vertThumbGeom(b, dataAreaH)
				t := mouse.Y - dt.sbDragStartMouse
				den := dataAreaH - vThumbH
				if vMax > 0 && den > 0 {
					dt.scrollY = dt.sbDragStartScroll + t*(vMax/den)
				}
				dt.clampScrollY()
			} else {
				den := dataAreaW - thumbW
				if maxHScroll > 0 && den > 0 {
					t := mouse.X - dt.sbDragStartMouse
					dt.scrollX = dt.sbDragStartScroll + t*(maxHScroll/den)
				}
				dt.clampScrollX()
			}
			dt.MarkDirty()
		}
	}

	// ── Wheel scroll (only when the pointer is over this table so file preview
	// and other panes can claim the wheel in the same frame).
	wheelV := rl.Vector2{}
	if !dt.sbDragging && inBounds {
		wheelV = rl.GetMouseWheelMoveV()
	}
	shift := rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
	maxHScrollContent := dt.contentW - dataAreaW
	if maxHScrollContent < 0 {
		maxHScrollContent = 0
	}
	if inBounds && !dt.sbDragging {
		// Trackpad / tilt wheel horizontal axis.
		if hasHorizBar && maxHScrollContent > 0 && wheelV.X != 0 {
			dt.scrollX -= wheelV.X * dt.RowHeight * 2
			dt.clampScrollX()
			dt.MarkDirty()
		}
	}
	wheelY := wheelV.Y
	if wheelY != 0 && inBounds && !dt.sbDragging {
		if hasHorizBar && maxHScrollContent > 0 && (shift || !contentOverflowY) {
			dt.scrollX -= wheelY * dt.RowHeight * 2
			dt.clampScrollX()
			dt.MarkDirty()
		} else if hasHorizBar && rl.CheckCollisionPointRec(mouse, horizTrackRect) && !shift {
			dt.scrollX -= wheelY * dt.RowHeight * 2
			dt.clampScrollX()
			dt.MarkDirty()
		} else if contentOverflowY {
			dt.scrollY -= wheelY * dt.RowHeight * 3
			dt.clampScrollY()
			dt.MarkDirty()
		}
	}

	// ── New scrollbar interactions (press) ────────────────────────────────────
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && inBounds && !dt.sbDragging {
		if needsHScroll && rl.CheckCollisionPointRec(mouse, horizTrackRect) {
			thumbRec := rl.NewRectangle(thumbX+2, trackY+2, thumbW-4, dtScrollbarH-4)
			if rl.CheckCollisionPointRec(mouse, thumbRec) {
				dt.sbDragging = true
				dt.sbDragVert = false
				dt.sbDragStartMouse = mouse.X
				dt.sbDragStartScroll = dt.scrollX
			} else {
				// Track jump: click position → scroll ratio.
				if maxHScroll > 0 && dataAreaW > 1 {
					r := (mouse.X - b.X) / dataAreaW
					if r < 0 {
						r = 0
					}
					if r > 1 {
						r = 1
					}
					dt.scrollX = r * maxHScroll
					dt.clampScrollX()
					dt.MarkDirty()
				}
			}
		} else if needsVScroll && rl.CheckCollisionPointRec(mouse, vertTrackRect) {
			_, vThumbY, vThumbH, vMax := dt.vertThumbGeom(b, dataAreaH)
			thumbRec := rl.NewRectangle(b.X+b.Width-dtScrollbarW+2, vThumbY+2, dtScrollbarW-4, vThumbH-4)
			if rl.CheckCollisionPointRec(mouse, thumbRec) {
				dt.sbDragging = true
				dt.sbDragVert = true
				dt.sbDragStartMouse = mouse.Y
				dt.sbDragStartScroll = dt.scrollY
			} else if vMax > 0 {
				r := (mouse.Y - (b.Y + dt.headerHeight())) / dataAreaH
				if r < 0 {
					r = 0
				}
				if r > 1 {
					r = 1
				}
				dt.scrollY = r * vMax
				dt.clampScrollY()
				dt.MarkDirty()
			}
		}
	}

	// ── Header hover ─────────────────────────────────────────────────────────
	headerRect := rl.NewRectangle(b.X, b.Y, dataAreaW, dt.headerHeight())
	prevHeaderHover := dt.hoverHeaderCol
	prevResizeHover := dt.hoverResizeCol
	dt.hoverHeaderCol = -1
	dt.hoverResizeCol = -1
	if rl.CheckCollisionPointRec(mouse, headerRect) {
		xInCol := mouse.X - b.X + dt.scrollX
		if dt.ResizableColumns {
			if col, on := dt.resizeColHit(xInCol, dataAreaW); on {
				dt.hoverResizeCol = col
				rl.SetMouseCursor(rl.MouseCursorResizeEW)
			}
		}
		if dt.hoverResizeCol < 0 {
			dt.hoverHeaderCol = dt.colIndexAt(xInCol, dataAreaW)
		}
	}

	// ── Row hover ────────────────────────────────────────────────────────────
	dataRect := rl.NewRectangle(b.X, b.Y+dt.headerHeight(), dataAreaW, dataAreaH)
	dt.hoverRow = -1
	if rl.CheckCollisionPointRec(mouse, dataRect) {
		localY := mouse.Y - (b.Y + dt.headerHeight()) + dt.scrollY
		row := int(localY / dt.RowHeight)
		if row >= 0 && row < len(dt.sortIndices) {
			dt.hoverRow = row
		}
	}

	if dt.hoverHeaderCol != prevHeaderHover || dt.hoverResizeCol != prevResizeHover {
		dt.MarkDrawDirty()
	}

	// ── Click handling (cells; skip when interacting with scrollbars) ──────
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && inBounds && dt.resizingCol < 0 {
		relY := mouse.Y - b.Y
		onHBar := needsHScroll && relY >= b.Height-dtScrollbarH
		onVBar := needsVScroll && mouse.X >= b.X+b.Width-dtScrollbarW
		if onHBar || onVBar {
			// Handled above.
		} else if relY < dt.headerHeight() {
			xInCol := mouse.X - b.X + dt.scrollX
			if dt.ResizableColumns {
				if col, on := dt.resizeColHit(xInCol, dataAreaW); on {
					dt.resizingCol = col
					dt.resizeDragMouseX = mouse.X
					dt.resizeDragStartW = dt.colWidths[col]
					dt.MarkDrawDirty()
					return
				}
			}
			colIdx := dt.colIndexAt(xInCol, dataAreaW)
			if colIdx >= 0 && colIdx < len(dt.Columns) {
				col := dt.Columns[colIdx]
				if col.Sortable && col.SortLess != nil {
					if dt.sortCol == colIdx {
						dt.sortAsc = !dt.sortAsc
					} else {
						dt.sortCol = colIdx
						dt.sortAsc = true
					}
					dt.rebuildSortIndices()
					dt.MarkDrawDirty()
				}
			}
		} else if relY < dt.headerHeight()+dataAreaH {
			localY := mouse.Y - (b.Y + dt.headerHeight()) + dt.scrollY
			row := int(localY / dt.RowHeight)
			if row >= 0 && row < len(dt.sortIndices) {
				bindingIdx := dt.sortIndices[row]
				dt.binding.SetSelectedIndex(bindingIdx)
				if dt.OnRowClick != nil {
					dt.OnRowClick(bindingIdx)
				}
				dt.MarkDrawDirty()
			}
		}
	}
}

// Draw renders the table.
func (dt *DataTable[T]) Draw() {
	if dt.IsHidden() {
		return
	}
	dt.drawInternal()
	dt.drawDirty = false
}

// InteractionOverlayActive is true while a data row is hovered — translucent
// highlight paints above the SSAA cache without MarkDrawDirty lag.
func (dt *DataTable[T]) InteractionOverlayActive() bool {
	return !dt.IsHidden() && dt.hoverRow >= 0
}

// DrawInteractionOverlay paints a low-alpha hover wash over the hovered row.
func (dt *DataTable[T]) DrawInteractionOverlay() {
	if !dt.InteractionOverlayActive() {
		return
	}
	b := dt.Bounds()
	dataAreaW, dataAreaH, _, _ := dt.scrollGeometry()
	if dataAreaW < 1 || dataAreaH < 1 {
		return
	}
	rowY := b.Y + dt.headerHeight() + float32(dt.hoverRow)*dt.RowHeight - dt.scrollY
	row := rl.NewRectangle(b.X, rowY, dataAreaW, dt.RowHeight)
	bodyClip := rl.NewRectangle(b.X, b.Y+dt.headerHeight(), dataAreaW, dataAreaH)
	clip := intersectRects(row, bodyClip)
	if clip.Width > 0 && clip.Height > 0 {
		rl.DrawRectangleRec(clip, dtHoverOverlayBg)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Inspector interface
// ─────────────────────────────────────────────────────────────────────────────

// dtInfo implements dataTableWidget for the Inspector panel.
func (dt *DataTable[T]) dtInfo() (colCount, rowCount, sortCol int, sortAsc bool, selectedRow int) {
	return len(dt.Columns), dt.binding.Len(), dt.sortCol, dt.sortAsc, dt.binding.GetSelectedIndex()
}

// ─────────────────────────────────────────────────────────────────────────────
// Drawing
// ─────────────────────────────────────────────────────────────────────────────

func (dt *DataTable[T]) cellFontSize() int32 {
	if dt.CellFontSize > 0 {
		return dt.CellFontSize
	}
	return dtCellFontSize
}

func (dt *DataTable[T]) headerFontSize() int32 {
	if dt.HeaderFontSize > 0 {
		return dt.HeaderFontSize
	}
	return dtHeaderFS
}

func (dt *DataTable[T]) headerHeight() float32 {
	if dt.HeaderHeight > 0 {
		return dt.HeaderHeight
	}
	fs := dt.headerFontSize()
	if fs > dtHeaderFS {
		return float32(fs) + 22
	}
	return dtHeaderH
}

func (dt *DataTable[T]) drawInternal() {
	b := dt.bounds

	dataAreaW, dataAreaH, vertSBW, horizSBH := dt.scrollGeometry()
	needsVScroll := vertSBW > 0
	needsHScroll := horizSBH > 0

	// ── Widget background ─────────────────────────────────────────────────────
	style := dt.GetStyle()
	rl.DrawRectangleRec(b, style.BackgroundColor)
	if style.BorderWidth > 0 {
		rl.DrawRectangleLinesEx(b, style.BorderWidth, style.BorderColor)
	}

	// ── Scissor: header vs body are separate bands so horizontal scroll reads as
	// a viewport over row cells; clip intersects parent viewport once per band.
	baseClip := rl.NewRectangle(b.X, b.Y, dataAreaW, b.Height-horizSBH)
	if vp := findViewport(dt); vp != nil {
		baseClip = intersectRects(baseClip, vp.ClipBounds())
	}
	if baseClip.Width <= 0 || baseClip.Height <= 0 {
		return
	}

	// Header and body use separate scissors so partially scrolled rows never
	// paint over the fixed column header band.
	headerClip := intersectRects(rl.NewRectangle(b.X, b.Y, b.Width, dt.headerHeight()), baseClip)
	bodyClip := intersectRects(rl.NewRectangle(b.X, b.Y+dt.headerHeight(), dataAreaW, dataAreaH), baseClip)

	if headerClip.Width > 0 && headerClip.Height > 0 {
		beginScissorMode(int32(headerClip.X), int32(headerClip.Y), int32(headerClip.Width), int32(headerClip.Height))
		dt.drawHeader(b, dataAreaW)
		rl.EndScissorMode()
	}
	if bodyClip.Width > 0 && bodyClip.Height > 0 {
		beginScissorMode(int32(bodyClip.X), int32(bodyClip.Y), int32(bodyClip.Width), int32(bodyClip.Height))
		dt.drawRows(b, dataAreaW, dataAreaH)
		rl.EndScissorMode()
	}

	// ── Scrollbars (own scissor pass) ─────────────────────────────────────────
	sbClip := rl.NewRectangle(b.X, b.Y, b.Width, b.Height)
	if vp := findViewport(dt); vp != nil {
		sbClip = intersectRects(sbClip, vp.ClipBounds())
	}
	if sbClip.Width > 0 && sbClip.Height > 0 {
		beginScissorMode(int32(sbClip.X), int32(sbClip.Y), int32(sbClip.Width), int32(sbClip.Height))
		if needsVScroll {
			dt.drawVertScrollbar(b, dataAreaH)
		}
		if needsHScroll {
			dt.drawHorizScrollbar(b, dataAreaW)
		}
		// Corner fill between scrollbars.
		if needsVScroll && needsHScroll {
			rl.DrawRectangleRec(
				rl.NewRectangle(b.X+dataAreaW, b.Y+b.Height-dtScrollbarH, dtScrollbarW, dtScrollbarH),
				dtScrollTrack)
		}
		rl.EndScissorMode()
	}
}

// drawHeader renders the column header row.
func (dt *DataTable[T]) drawHeader(b rl.Rectangle, dataAreaW float32) {
	// Header background.
	rl.DrawRectangleRec(rl.NewRectangle(b.X, b.Y, dataAreaW, dt.headerHeight()), dtHeaderBg)
	// Bottom border.
	rl.DrawLineEx(
		rl.NewVector2(b.X, b.Y+dt.headerHeight()-1),
		rl.NewVector2(b.X+dataAreaW, b.Y+dt.headerHeight()-1),
		1.5, dtHeaderBorderCol)

	x := b.X - dt.scrollX
	colWidths := dt.effectiveColumnWidths(dataAreaW)
	for i, col := range dt.Columns {
		cellX := x
		cellW := colWidths[i]

		if cellX+cellW > b.X && cellX < b.X+dataAreaW {
			isHovered := dt.hoverHeaderCol == i && col.Sortable && col.SortLess != nil
			isSorted := dt.sortCol == i

			// Hover highlight.
			if isHovered {
				rl.DrawRectangleRec(
					rl.NewRectangle(cellX, b.Y, cellW, dt.headerHeight()-1),
					rl.NewColor(232, 235, 252, 255))
			}

			hdrPadX := float32(10)
			textX := cellX + hdrPadX
			textColor := dtCellText
			if isSorted {
				textColor = dtSortedHeaderText
			}
			hdrStyle := Style{FontSize: dt.headerFontSize(), Bold: true, TextColor: textColor}
			titleMaxW := cellW - 2*hdrPadX
			if col.Sortable && col.SortLess != nil {
				titleMaxW -= dtSortArrowW
			}
			if titleMaxW < 8 {
				titleMaxW = 8
			}
			titleText := col.Title
			if MeasureText(col.Title, float32(hdrStyle.FontSize)) > titleMaxW {
				titleText = truncateTextS(col.Title, titleMaxW, hdrStyle)
			}
			drawTextS(titleText, int32(textX), TextPosY(
				rl.NewRectangle(cellX, b.Y, cellW, dt.headerHeight()-1), hdrStyle), hdrStyle)

			// Sort arrow.
			arrowX := cellX + cellW - dtCellPadX - dtSortArrowW/2
			arrowY := b.Y + dt.headerHeight()/2
			if isSorted {
				dtDrawSortArrow(arrowX, arrowY, dt.sortAsc, dtSortedHeaderText)
			} else if isHovered {
				dtDrawSortArrow(arrowX, arrowY, true, rl.NewColor(180, 186, 210, 160))
			}

			// Column separator / resize affordance.
			if i < len(dt.Columns)-1 && col.Title != "" && dt.Columns[i+1].Title != "" {
				sepX := cellX + cellW
				if sepX > b.X && sepX < b.X+dataAreaW {
					sepCol := dtColSepCol
					if dt.ResizableColumns && dt.hoverResizeCol == i {
						sepCol = rl.NewColor(79, 70, 229, 255)
					}
					rl.DrawLineEx(
						rl.NewVector2(sepX, b.Y+4),
						rl.NewVector2(sepX, b.Y+dt.headerHeight()-4),
						1.5, sepCol)
				}
			}
		}
		x += colWidths[i]
	}
}

// drawRows renders the visible data rows.
func (dt *DataTable[T]) drawRows(b rl.Rectangle, dataAreaW, dataAreaH float32) {
	dataStartY := b.Y + dt.headerHeight()
	n := len(dt.sortIndices)
	bodyClip := rl.NewRectangle(b.X, dataStartY, dataAreaW, dataAreaH)
	if vp := findViewport(dt); vp != nil {
		bodyClip = intersectRects(bodyClip, vp.ClipBounds())
	}

	// Visible row range — virtual rendering.
	startRow := int(dt.scrollY / dt.RowHeight)
	endRow := int((dt.scrollY+dataAreaH)/dt.RowHeight) + 1
	if startRow < 0 {
		startRow = 0
	}
	if endRow > n {
		endRow = n
	}

	for dispRow := startRow; dispRow < endRow; dispRow++ {
		bindingIdx := dt.sortIndices[dispRow]
		item := dt.binding.GetItem(bindingIdx)
		isSelected := bindingIdx == dt.binding.GetSelectedIndex()
		rowY := dataStartY + float32(dispRow)*dt.RowHeight - dt.scrollY

		var rowBg rl.Color
		if isSelected {
			rowBg = dtSelectedRowBg
		} else if dt.ShowAlternatingRows && dispRow%2 == 1 {
			rowBg = dtAltRowBg
		} else {
			rowBg = dtRowBg
		}

		// Per-cell backgrounds so zebra/selection track horizontal column scroll.
		x := b.X - dt.scrollX
		colWidths := dt.effectiveColumnWidths(dataAreaW)
		for i := range dt.Columns {
			cellX := x
			cellW := colWidths[i]
			if cellX+cellW > b.X && cellX < b.X+dataAreaW {
				cellRect := rl.NewRectangle(cellX, rowY, cellW, dt.RowHeight)
				clip := intersectRects(cellRect, bodyClip)
				if clip.Width > 0 && clip.Height > 0 {
					rl.DrawRectangleRec(clip, rowBg)
				}
			}
			x += colWidths[i]
		}

		// Row bottom separator (viewport width).
		rl.DrawLineEx(
			rl.NewVector2(b.X, rowY+dt.RowHeight-0.5),
			rl.NewVector2(b.X+dataAreaW, rowY+dt.RowHeight-0.5),
			1, dtRowSepCol)

		// Cell text / custom draw (same x walk).
		x = b.X - dt.scrollX
		for i, col := range dt.Columns {
			cellX := x
			cellW := colWidths[i]

			if cellX+cellW > b.X && cellX < b.X+dataAreaW {
				cellRect := rl.NewRectangle(cellX, rowY, cellW, dt.RowHeight)

				if col.CellDraw != nil {
					clip := intersectRects(cellRect, bodyClip)
					if clip.Width > 0 && clip.Height > 0 {
						beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
						col.CellDraw(item, bindingIdx, isSelected, cellRect)
						rl.EndScissorMode()
					}
				} else if col.Render != nil {
					text := col.Render(item)
					textColor := dtCellText
					if isSelected {
						textColor = dtSelectedText
					}
					cellStyle := Style{FontSize: dt.cellFontSize(), TextColor: textColor}
					padL := dtCellPadX
					padR := dtCellPadX
					textMaxW := cellW - padL - padR
					if textMaxW < 4 {
						textMaxW = 4
					}
					text = truncateTextS(text, textMaxW, cellStyle)
					textW := measureTextS(text, cellStyle)
					var textX float32
					switch col.Align {
					case ColumnAlignCenter:
						textX = cellX + (cellW-float32(textW))/2
					case ColumnAlignRight:
						textX = cellX + cellW - float32(textW) - padR
					default: // ColumnAlignLeft
						textX = cellX + padL
					}
					clip := intersectRects(cellRect, bodyClip)
					if clip.Width > 0 && clip.Height > 0 {
						beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
						drawTextS(text, int32(textX), toolbarTextPosY(cellRect, cellStyle), cellStyle)
						rl.EndScissorMode()
					}
				}
			}
			x += colWidths[i]
		}
	}
}

// drawVertScrollbar renders the vertical scrollbar.
func (dt *DataTable[T]) drawVertScrollbar(b rl.Rectangle, dataAreaH float32) {
	trackX := b.X + b.Width - dtScrollbarW
	trackH := dataAreaH
	if trackH < 1 {
		trackH = 1
	}
	// Gutter must match header fill — otherwise a white strip appears above the track.
	rl.DrawRectangleRec(rl.NewRectangle(trackX, b.Y, dtScrollbarW, dt.headerHeight()), dtHeaderBg)
	rl.DrawRectangleRec(rl.NewRectangle(trackX, b.Y+dt.headerHeight(), dtScrollbarW, trackH), dtScrollTrack)

	if dt.contentH <= dataAreaH {
		return
	}
	thumbH := (dataAreaH / dt.contentH) * trackH
	if thumbH < 20 {
		thumbH = 20
	}
	maxScroll := dt.contentH - dataAreaH
	var thumbY float32
	if maxScroll > 0 {
		thumbY = b.Y + dt.headerHeight() + (dt.scrollY/maxScroll)*(trackH-thumbH)
	} else {
		thumbY = b.Y + dt.headerHeight()
	}
	rl.DrawRectangleRounded(
		rl.NewRectangle(trackX+2, thumbY+2, dtScrollbarW-4, thumbH-4),
		0.5, 4, dtScrollThumb)
}

// drawHorizScrollbar renders the horizontal scrollbar.
func (dt *DataTable[T]) drawHorizScrollbar(b rl.Rectangle, dataAreaW float32) {
	trackY := b.Y + b.Height - dtScrollbarH
	rl.DrawRectangleRec(rl.NewRectangle(b.X, trackY, dataAreaW, dtScrollbarH), dtScrollTrack)

	if dt.contentW <= dataAreaW {
		return
	}
	thumbW := (dataAreaW / dt.contentW) * dataAreaW
	if thumbW < 20 {
		thumbW = 20
	}
	maxScroll := dt.contentW - dataAreaW
	var thumbX float32
	if maxScroll > 0 {
		thumbX = b.X + (dt.scrollX/maxScroll)*(dataAreaW-thumbW)
	} else {
		thumbX = b.X
	}
	rl.DrawRectangleRounded(
		rl.NewRectangle(thumbX+2, trackY+2, thumbW-4, dtScrollbarH-4),
		0.5, 4, dtScrollThumb)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// totalColumnsWidth returns the sum of all column widths.
func (dt *DataTable[T]) totalColumnsWidth() float32 {
	w := float32(0)
	for _, cw := range dt.colWidths {
		w += cw
	}
	return w
}

// effectiveColumnWidths returns per-column widths, stretching the last column to
// fill the data area when fixed widths leave a trailing gap.
func (dt *DataTable[T]) effectiveColumnWidths(dataAreaW float32) []float32 {
	out := make([]float32, len(dt.colWidths))
	total := float32(0)
	for i, cw := range dt.colWidths {
		out[i] = cw
		total += cw
	}
	if len(out) > 0 && total > 0 && dataAreaW > total+0.5 {
		out[len(out)-1] += dataAreaW - total
	}
	return out
}

// resizeColHit reports whether x (column-space coords) is on a column resize handle.
func (dt *DataTable[T]) resizeColHit(xInColSpace, dataAreaW float32) (col int, onHandle bool) {
	if !dt.ResizableColumns || len(dt.colWidths) == 0 {
		return -1, false
	}
	colWidths := dt.effectiveColumnWidths(dataAreaW)
	x := float32(0)
	for i, w := range colWidths {
		edge := x + w
		if xInColSpace >= edge-dtResizeHandleW && xInColSpace <= edge+dtResizeHandleW {
			return i, true
		}
		x += w
	}
	return -1, false
}

// colIndexAt returns the column index for a given x offset measured from the
// start of the column grid (0 = left edge of the first column, independent of
// scrollX). Returns -1 if x falls outside all columns.
func (dt *DataTable[T]) colIndexAt(xInColSpace, dataAreaW float32) int {
	colWidths := dt.effectiveColumnWidths(dataAreaW)
	x := float32(0)
	for i, w := range colWidths {
		if xInColSpace >= x && xInColSpace < x+w {
			return i
		}
		x += w
	}
	return -1
}

// ScrollToBindingIndex scrolls vertically so the binding row is visible.
func (dt *DataTable[T]) ScrollToBindingIndex(bindingIdx int) {
	if dt == nil || bindingIdx < 0 {
		return
	}
	if len(dt.sortIndices) != dt.binding.Len() {
		dt.rebuildSortIndices()
	}
	dispRow := -1
	for i, bi := range dt.sortIndices {
		if bi == bindingIdx {
			dispRow = i
			break
		}
	}
	if dispRow < 0 {
		return
	}
	_, dataAreaH, _, _ := dt.scrollGeometry()
	rowTop := float32(dispRow) * dt.RowHeight
	rowBottom := rowTop + dt.RowHeight
	if rowTop < dt.scrollY {
		dt.scrollY = rowTop
	} else if rowBottom > dt.scrollY+dataAreaH {
		dt.scrollY = rowBottom - dataAreaH
	}
	dt.clampScrollY()
	dt.MarkDrawDirty()
}

// clampScrollY keeps scrollY in [0, max].
func (dt *DataTable[T]) clampScrollY() {
	_, dataAreaH, _, _ := dt.scrollGeometry()
	maxScroll := dt.contentH - dataAreaH
	if maxScroll < 0 {
		maxScroll = 0
	}
	if dt.scrollY < 0 {
		dt.scrollY = 0
	}
	if dt.scrollY > maxScroll {
		dt.scrollY = maxScroll
	}
}

// clampScrollX keeps scrollX in [0, max].
func (dt *DataTable[T]) clampScrollX() {
	dataAreaW, _, _, _ := dt.scrollGeometry()
	maxScroll := dt.contentW - dataAreaW
	if maxScroll < 0 {
		maxScroll = 0
	}
	if dt.scrollX < 0 {
		dt.scrollX = 0
	}
	if dt.scrollX > maxScroll {
		dt.scrollX = maxScroll
	}
}

// rebuildSortIndices rebuilds the sortIndices slice using the current sort
// column and direction. If sortCol == -1 the result is the identity mapping.
func (dt *DataTable[T]) rebuildSortIndices() {
	n := dt.binding.Len()
	if len(dt.sortIndices) != n {
		dt.sortIndices = make([]int, n)
	}
	for i := range dt.sortIndices {
		dt.sortIndices[i] = i
	}
	if dt.sortCol < 0 || dt.sortCol >= len(dt.Columns) {
		return
	}
	col := dt.Columns[dt.sortCol]
	if col.SortLess == nil {
		return
	}
	items := dt.binding.GetItems()
	asc := dt.sortAsc
	sort.SliceStable(dt.sortIndices, func(i, j int) bool {
		a := items[dt.sortIndices[i]]
		b := items[dt.sortIndices[j]]
		if asc {
			return col.SortLess(a, b)
		}
		return col.SortLess(b, a)
	})
}

// dtDrawSortArrow draws a small triangle indicating sort direction at (cx, cy).
func dtDrawSortArrow(cx, cy float32, ascending bool, color rl.Color) {
	sz := float32(4)
	if ascending {
		// ▲ up
		rl.DrawTriangle(
			rl.NewVector2(cx, cy-sz),
			rl.NewVector2(cx+sz, cy+sz),
			rl.NewVector2(cx-sz, cy+sz),
			color)
	} else {
		// ▼ down
		rl.DrawTriangle(
			rl.NewVector2(cx-sz, cy-sz),
			rl.NewVector2(cx+sz, cy-sz),
			rl.NewVector2(cx, cy+sz),
			color)
	}
}
