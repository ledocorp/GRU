package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

const textEditorDragThreshold = float32(4)

func (ed *TextEditor) updateMouseSelection(mouse rl.Vector2) {
	if IsContextMenuOpen() || OverlayBlocksSceneInput() {
		return
	}
	if !ed.hovered {
		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			ed.dragSelect = false
		}
		return
	}
	drawStyle := ed.editorStyle(ed.GetStyle())
	bounds := ed.Bounds()
	pad := int32(drawStyle.Padding + 0.5)
	if pad < 8 {
		pad = 8
	}

	// Ignore clicks on the horizontal scrollbar track.
	if !ed.WordWrap {
		vp, _ := ed.horizontalScrollbarHost()
		track, _, maxX := ed.horizScrollbarGeom(vp, drawStyle)
		if maxX > 0 && rl.CheckCollisionPointRec(mouse, track) {
			if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
				ed.dragSelect = false
			}
			return
		}
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !PointerClickHandled() {
		pos := ed.cursorFromMouse(mouse, drawStyle, bounds, pad)
		ed.selectPressMouse = mouse
		ed.selectDragging = false
		if textInputShiftDown() {
			if ed.selAnchor < 0 {
				ed.selAnchor = ed.cursor
			}
			ed.cursor = pos
			ed.selectDragging = true
		} else {
			ed.selAnchor = pos
			ed.cursor = pos
			ed.dragSelect = true
		}
		ed.RestartCaretCycle()
		ed.ensureCursorVisible()
		ed.markContentDirty()
	}
	if ed.dragSelect && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if !ed.selectDragging {
			dx := mouse.X - ed.selectPressMouse.X
			dy := mouse.Y - ed.selectPressMouse.Y
			if dx*dx+dy*dy < textEditorDragThreshold*textEditorDragThreshold {
				return
			}
			ed.selectDragging = true
		}
		pos := ed.cursorFromMouse(mouse, drawStyle, bounds, pad)
		if pos != ed.cursor {
			ed.cursor = pos
			ed.ensureCursorVisible()
			ed.markContentDirty()
		}
	}
	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		if ed.dragSelect && (!ed.selectDragging || ed.selAnchor == ed.cursor) {
			ed.selAnchor = -1
			ed.markContentDirty()
		}
		ed.dragSelect = false
		ed.selectDragging = false
	}
}

func (ed *TextEditor) cursorFromMouse(mouse rl.Vector2, style Style, bounds rl.Rectangle, pad int32) int {
	text := ed.Text.Get()
	relX := mouse.X - bounds.X - float32(pad) + ed.scrollX
	relY := mouse.Y - bounds.Y - float32(pad)
	if findViewport(ed) == nil {
		relY += ed.scrollY
	}
	if relY < 0 {
		relY = 0
	}
	lh := ed.lineHeight(style)
	if lh < 1 {
		lh = 1
	}
	lineIdx := int(relY / lh)
	lines, starts := ed.displayLines(style)
	if len(lines) == 0 {
		return 0
	}
	if lineIdx < 0 {
		lineIdx = 0
	}
	if lineIdx >= len(lines) {
		return len(text)
	}
	col := cursorOffsetInLine(lines[lineIdx], relX, style)
	off := starts[lineIdx] + col
	if off > len(text) {
		return len(text)
	}
	return off
}

func cursorOffsetInLine(line string, innerX float32, style Style) int {
	if off, ok := shapedCaretOffsetAtX(line, innerX, style); ok {
		return off
	}
	if innerX <= 0 {
		return 0
	}
	lo, hi := 0, len(line)
	for lo < hi {
		mid := (lo + hi) / 2
		if EditorMeasureWidth(line[:mid], style) < innerX {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo > 0 {
		wPrev := EditorMeasureWidth(line[:lo-1], style)
		wCur := EditorMeasureWidth(line[:lo], style)
		if innerX-wPrev < wCur-innerX {
			return lo - 1
		}
	}
	if lo > len(line) {
		return len(line)
	}
	return lo
}
