package ui

import (
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ShowContextMenuAt opens the editor cut/copy/paste menu at the given screen position.
func (ed *TextEditor) ShowContextMenuAt(mouse rl.Vector2) {
	if ed.Disabled {
		return
	}
	if d := ActiveDocument(); d != nil {
		d.SetFocus(ed)
	}
	selected := ed.selectedText()
	clip := strings.TrimRight(rl.GetClipboardText(), "\x00")
	items := []ContextMenuItem{
		{Label: "Cut", Disabled: selected == "", Action: func() { ed.Cut() }},
		{Label: "Copy", Disabled: selected == "" && ed.Text.Get() == "", Action: func() { ed.Copy() }},
		{Label: "Paste", Disabled: clip == "", Action: func() { ed.Paste() }},
		{Divider: true},
		{Label: "Select all", Action: func() { ed.SelectAll() }},
	}
	ShowContextMenu(items, mouse.X, mouse.Y)
}
