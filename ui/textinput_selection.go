// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var textInputSelectionFill = rl.NewColor(79, 70, 229, 60)

func textInputCtrlDown() bool {
	return rl.IsKeyDown(rl.KeyLeftControl) || rl.IsKeyDown(rl.KeyRightControl)
}

func textInputShiftDown() bool {
	return rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift)
}

func (ti *TextInput) hasSelection() bool {
	return ti.selAnchor >= 0 && ti.selAnchor != ti.cursor
}

func (ti *TextInput) selectionRange() (lo, hi int) {
	if !ti.hasSelection() {
		return 0, 0
	}
	lo, hi = ti.selAnchor, ti.cursor
	if lo > hi {
		lo, hi = hi, lo
	}
	return lo, hi
}

func (ti *TextInput) selectedText() string {
	lo, hi := ti.selectionRange()
	if lo >= hi {
		return ""
	}
	return ti.Text.Get()[lo:hi]
}

func (ti *TextInput) clearSelection() {
	if ti.selAnchor >= 0 {
		ti.selAnchor = -1
		ti.MarkDrawDirty()
	}
}

func (ti *TextInput) deleteSelection() bool {
	if !ti.hasSelection() {
		return false
	}
	lo, hi := ti.selectionRange()
	text := ti.Text.Get()
	ti.Text.Set(text[:lo] + text[hi:])
	ti.cursor = lo
	ti.selAnchor = -1
	ti.updateScroll()
	ti.MarkDrawDirty()
	return true
}

func (ti *TextInput) selectAll() {
	text := ti.Text.Get()
	if len(text) == 0 {
		return
	}
	ti.selAnchor = 0
	ti.cursor = len(text)
	ti.updateScroll()
	ti.MarkDrawDirty()
}

func (ti *TextInput) cursorFromMouseX(mouseX float32, style Style, bounds rl.Rectangle, pad int32) int {
	text := ti.Text.Get()
	innerX := mouseX - bounds.X - float32(pad) + float32(ti.scrollOffset)
	if innerX <= 0 {
		return 0
	}
	lo, hi := 0, len(text)
	for lo < hi {
		mid := (lo + hi) / 2
		if float32(measureTextS(text[:mid], style)) < innerX {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo > 0 {
		wPrev := float32(measureTextS(text[:lo-1], style))
		wCur := float32(measureTextS(text[:lo], style))
		if innerX-wPrev < wCur-innerX {
			return lo - 1
		}
	}
	if lo > len(text) {
		return len(text)
	}
	return lo
}

func (ti *TextInput) updateMouseSelection(mouse rl.Vector2, style Style, bounds rl.Rectangle, pad int32) {
	if IsContextMenuOpen() {
		return
	}
	if ti.Disabled {
		return
	}
	if rl.IsMouseButtonPressed(rl.MouseRightButton) && (ti.hovered || ti.keyboardActive()) {
		ti.showContextMenu(mouse)
		return
	}
	if !ti.hovered {
		if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
			ti.dragSelect = false
		}
		return
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		pos := ti.cursorFromMouseX(mouse.X, style, bounds, pad)
		if textInputShiftDown() {
			if ti.selAnchor < 0 {
				ti.selAnchor = ti.cursor
			}
			ti.cursor = pos
		} else {
			ti.selAnchor = pos
			ti.cursor = pos
			ti.dragSelect = true
		}
		ti.updateScroll()
		ti.MarkDrawDirty()
	}
	if ti.dragSelect && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		pos := ti.cursorFromMouseX(mouse.X, style, bounds, pad)
		if pos != ti.cursor {
			ti.cursor = pos
			ti.updateScroll()
			ti.MarkDrawDirty()
		}
	}
	if rl.IsMouseButtonReleased(rl.MouseLeftButton) {
		if ti.dragSelect && ti.selAnchor == ti.cursor {
			ti.selAnchor = -1
			ti.MarkDrawDirty()
		}
		ti.dragSelect = false
	}
}

func (ti *TextInput) updateSelectionShortcuts() {
	if !ti.keyboardActive() || ti.Disabled {
		return
	}
	ctrl := textInputCtrlDown()
	if ctrl && rl.IsKeyPressed(rl.KeyA) {
		ti.selectAll()
		return
	}
	if ctrl && rl.IsKeyPressed(rl.KeyC) {
		if selected := ti.selectedText(); selected != "" {
			rl.SetClipboardText(selected)
		}
		return
	}
	if ctrl && rl.IsKeyPressed(rl.KeyX) {
		if selected := ti.selectedText(); selected != "" {
			rl.SetClipboardText(selected)
			ti.deleteSelection()
		}
		return
	}
	if ctrl && rl.IsKeyPressed(rl.KeyV) {
		clip := strings.TrimRight(rl.GetClipboardText(), "\x00")
		if clip == "" {
			return
		}
		clip = strings.ReplaceAll(clip, "\r\n", "\n")
		clip = strings.ReplaceAll(clip, "\r", "\n")
		clip = strings.ReplaceAll(clip, "\n", "")
		if ti.hasSelection() {
			ti.deleteSelection()
		}
		text := ti.Text.Get()
		ti.Text.Set(text[:ti.cursor] + clip + text[ti.cursor:])
		ti.cursor += len(clip)
		ti.updateScroll()
		ti.MarkDrawDirty()
	}
}

func (ti *TextInput) showContextMenu(mouse rl.Vector2) {
	if d := ActiveDocument(); d != nil {
		d.SetFocus(ti)
	}
	selected := ti.selectedText()
	clip := strings.TrimRight(rl.GetClipboardText(), "\x00")
	items := []ContextMenuItem{
		{Label: "Cut", Disabled: selected == "", Action: func() {
			if text := ti.selectedText(); text != "" {
				rl.SetClipboardText(text)
				ti.deleteSelection()
			}
		}},
		{Label: "Copy", Disabled: selected == "", Action: func() {
			if text := ti.selectedText(); text != "" {
				rl.SetClipboardText(text)
			}
		}},
		{Label: "Paste", Disabled: clip == "", Action: func() {
			clip := strings.TrimRight(rl.GetClipboardText(), "\x00")
			if clip == "" {
				return
			}
			clip = strings.ReplaceAll(clip, "\r\n", "\n")
			clip = strings.ReplaceAll(clip, "\r", "\n")
			clip = strings.ReplaceAll(clip, "\n", "")
			if ti.hasSelection() {
				ti.deleteSelection()
			}
			text := ti.Text.Get()
			ti.Text.Set(text[:ti.cursor] + clip + text[ti.cursor:])
			ti.cursor += len(clip)
			ti.updateScroll()
			ti.MarkDrawDirty()
		}},
		{Divider: true},
		{Label: "Select all", Action: func() { ti.selectAll() }},
	}
	ShowContextMenu(items, mouse.X, mouse.Y)
}

func (ti *TextInput) drawSelectionHighlight(text string, bounds rl.Rectangle, pad int32, posY int32, style Style) {
	if !ti.hasSelection() {
		return
	}
	lo, hi := ti.selectionRange()
	x1 := int32(bounds.X) + pad + measureTextS(text[:lo], style) - ti.scrollOffset
	x2 := int32(bounds.X) + pad + measureTextS(text[:hi], style) - ti.scrollOffset
	if x2 > x1 {
		rl.DrawRectangle(x1, posY, x2-x1, int32(EffectiveFontSize(style)), textInputSelectionFill)
	}
}
