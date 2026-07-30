// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const tbWordTogglePad = float32(6) // extra space left/right of label + switch

// ToolbarWordToggle is a flat-toolbar row item: static label text + compact switch.
type ToolbarWordToggle struct {
	Element
	Text   string
	Value  *Signal[bool]
	toggle *Toggle
}

// NewToolbarWordToggle creates a label + toggle pair for command bars.
func NewToolbarWordToggle(id, text string, value *Signal[bool]) *ToolbarWordToggle {
	if value == nil {
		value = NewSignal(false)
	}
	wt := &ToolbarWordToggle{
		Element: NewElement(id, 0, 0, 0, 0),
		Text:    text,
		Value:   value,
	}
	wt.toggle = NewToggle(id+"-switch", value.Get(), 0, 0, tbToggleW, tbToggleH)
	wt.toggle.AutoHeight = false
	wt.toggle.Value = value
	wt.styleName = "toolbar"
	value.Subscribe(func() { wt.MarkDrawDirty() })
	return wt
}

// NaturalWidth implements toolbar width budgeting for ToolbarItemCustom.
func (wt *ToolbarWordToggle) NaturalWidth() float32 {
	st := GetThemeStyle("toolbar-btn")
	gap := float32(10)
	return tbWordTogglePad*2 + float32(measureTextS(wt.Text, st)) + gap + tbToggleW
}

// GetPreferredWidth implements flex row width hinting.
func (wt *ToolbarWordToggle) GetPreferredWidth() float32 {
	if wt.PreferredWidth > 0 {
		return wt.PreferredWidth
	}
	return wt.NaturalWidth()
}

// SetDisabled greys out the switch and ignores clicks.
func (wt *ToolbarWordToggle) SetDisabled(disabled bool) {
	if wt == nil || wt.toggle == nil {
		return
	}
	wt.toggle.Disabled = disabled
	wt.MarkDrawDirty()
}

func (wt *ToolbarWordToggle) wordToggleContent(b rl.Rectangle) (content rl.Rectangle, st Style) {
	st = GetThemeStyle("toolbar-btn")
	content = toolbarContentRect(b, st)
	content.X += tbWordTogglePad
	return content, st
}

func (wt *ToolbarWordToggle) Layout() {
	b := wt.Bounds()
	content, st := wt.wordToggleContent(b)
	textW := float32(measureTextS(wt.Text, st))
	tw := tbToggleW
	th := tbToggleH
	tx := content.X + textW + 10
	ty := toolbarAccessoryY(content, st, th)
	wt.toggle.setBoundsNoMark(rl.NewRectangle(tx, ty, tw, th))
	wt.layoutDirty = false
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (wt *ToolbarWordToggle) ClearOverlayPointerState() {
	wt.toggle.ClearOverlayPointerState()
}

func (wt *ToolbarWordToggle) Update(dt float32) {
	if wt.IsHidden() {
		return
	}
	wt.toggle.Update(dt)
}

func (wt *ToolbarWordToggle) Draw() {
	if wt.IsHidden() {
		return
	}
	b := wt.Bounds()
	content, st := wt.wordToggleContent(b)
	st.TextColor = rl.NewColor(73, 80, 87, 255)
	drawTextS(wt.Text, int32(content.X), toolbarTextPosY(content, st), st)
	wt.toggle.Draw()
}

func (wt *ToolbarWordToggle) IsInteractive() bool { return true }
