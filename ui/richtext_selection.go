// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const richTextDragThreshold = float32(4)

// ClearSelection removes the current RichText token selection.
func (rt *RichText) ClearSelection() {
	if rt.selectStart < 0 && rt.selectEnd < 0 && !rt.selecting && !rt.selectPress {
		return
	}
	rt.selecting = false
	rt.selectPress = false
	rt.selectAnchor = -1
	rt.selectStart = -1
	rt.selectEnd = -1
	rt.MarkDrawDirty()
}

// SelectedText returns the currently highlighted token text.
func (rt *RichText) SelectedText() string {
	if !rt.Selectable || rt.selectStart < 0 || rt.selectEnd < 0 {
		return ""
	}
	lo, hi := rt.selectionRange()
	var b strings.Builder
	for _, tok := range rt.tokens() {
		if tok.index < lo || tok.index > hi {
			continue
		}
		b.WriteString(tok.text)
	}
	return strings.TrimSpace(b.String())
}

// PlainText returns all span text without styling metadata.
func (rt *RichText) PlainText() string {
	var b strings.Builder
	for _, span := range rt.Spans {
		b.WriteString(span.Text)
	}
	return strings.TrimSpace(b.String())
}

func (rt *RichText) updateSelection(mouse rl.Vector2) {
	// ContextMenuMgr runs after document Update in the main loop. When the user
	// clicks a RichText context-menu item, do not let this node clear selection
	// before the menu action has a chance to read/copy it.
	if IsContextMenuOpen() {
		return
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if !rt.pointInTextContent(mouse) {
			rt.ClearSelection()
			return
		}
		rt.selectPress = true
		rt.selectPressMouse = mouse
		rt.selectDragging = false
		rt.selecting = false
		rt.selectAnchor = rt.selectableTokenAt(mouse)
		rt.selectStart = -1
		rt.selectEnd = -1
		rt.MarkDrawDirty()
	}
	if rt.selectPress && !rt.selecting && rt.selectAnchor >= 0 && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		if !rt.selectDragging {
			dx := mouse.X - rt.selectPressMouse.X
			dy := mouse.Y - rt.selectPressMouse.Y
			if dx*dx+dy*dy < richTextDragThreshold*richTextDragThreshold {
				return
			}
			rt.selectDragging = true
		}
		idx := rt.selectableTokenAt(mouse)
		if idx >= 0 && idx != rt.selectAnchor {
			rt.selecting = true
		}
	}
	if rt.selecting && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		idx := rt.selectableTokenAt(mouse)
		if idx >= 0 {
			start, end := rt.selectAnchor, idx
			if start > end {
				start, end = end, start
			}
			if rt.selectStart != start || rt.selectEnd != end {
				rt.selectStart = start
				rt.selectEnd = end
				rt.MarkDrawDirty()
			}
		}
	}
	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		if rt.selectPress && !rt.selecting {
			rt.selectStart = -1
			rt.selectEnd = -1
			rt.MarkDrawDirty()
		}
		rt.selectPress = false
		rt.selectDragging = false
		rt.selecting = false
		rt.selectAnchor = -1
	}
	if rt.hasSelection() && richTextCopyShortcutPressed() {
		if selected := rt.SelectedText(); selected != "" {
			rl.SetClipboardText(selected)
		}
	}
	if rl.IsMouseButtonPressed(rl.MouseRightButton) && rt.pointInTextContent(mouse) {
		rt.showContextMenu(mouse)
	}
}

func (rt *RichText) hasSelection() bool {
	return rt.selectStart >= 0 && rt.selectEnd >= 0
}

func (rt *RichText) tokenSelected(idx int) bool {
	if !rt.hasSelection() {
		return false
	}
	lo, hi := rt.selectionRange()
	return idx >= lo && idx <= hi
}

func (rt *RichText) selectionRange() (int, int) {
	lo, hi := rt.selectStart, rt.selectEnd
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi
}

func richTextCopyShortcutPressed() bool {
	ctrl := rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
	return ctrl && rl.IsKeyPressed(rl.KeyC)
}

func (rt *RichText) showContextMenu(mouse rl.Vector2) {
	selected := rt.SelectedText()
	items := []ContextMenuItem{
		{Label: "Copy", Disabled: selected == "", Action: func() {
			if text := rt.SelectedText(); text != "" {
				rl.SetClipboardText(text)
			}
		}},
		{Label: "Select all", Action: func() {
			rt.SelectAll()
		}},
		{Divider: true},
		{Label: "Copy all text", Action: func() {
			if text := rt.PlainText(); text != "" {
				rl.SetClipboardText(text)
			}
		}},
	}
	ShowContextMenu(items, mouse.X, mouse.Y)
}

// SelectAll highlights every non-empty token in the RichText block.
func (rt *RichText) SelectAll() {
	tokens := rt.tokens()
	first, last := -1, -1
	for _, tok := range tokens {
		if strings.TrimSpace(tok.text) == "" {
			continue
		}
		if first < 0 {
			first = tok.index
		}
		last = tok.index
	}
	if first < 0 {
		return
	}
	rt.selectStart = first
	rt.selectEnd = last
	rt.selecting = false
	rt.MarkDrawDirty()
}

func (rt *RichText) selectableTokenAt(point rl.Vector2) int {
	if idx := rt.tokenAt(point); idx >= 0 {
		return idx
	}
	return rt.tokenAtLineX(point)
}
