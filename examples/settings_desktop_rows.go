// Package examples — desktop settings row helpers (panel + flat layouts).
package examples

import "github.com/ledocorp/gru/ui"

func settingsToggleRow(id, title, subtitle string, tog *ui.Toggle) *ui.ListTile {
	tile := ui.NewListTile(id, title, subtitle, 0, 0, 0, 0)
	tile.SetTrailing(tog)
	return tile
}

func settingsIconToggleRow(id, title, subtitle string, icon string, tog *ui.Toggle) *ui.ListTile {
	tile := settingsToggleRow(id, title, subtitle, tog)
	tile.SetLeading(settingsRowIcon(id+"-icon", icon))
	return tile
}

func settingsFlatToggleRow(id, title, subtitle string, icon string, tog *ui.Toggle) *ui.ListTile {
	tile := settingsIconToggleRow(id, title, subtitle, icon, tog)
	return tile
}

func settingsRowIcon(id string, icon string) *ui.Icon {
	return ui.NewIcon(id, icon, ui.PhosphorRegular, 0, 0, 22, 22)
}

func settingsFlatComboRow(id, title, subtitle string, icon string, cb *ui.ComboBox) *ui.ListTile {
	tile := ui.NewListTile(id, title, subtitle, 0, 0, 0, 0)
	tile.SetLeading(settingsRowIcon(id+"-icon", icon))
	cb.PreferredWidth = 168
	cb.MaxWidth = 168
	cb.MinWidth = 120
	tile.SetTrailing(cb)
	return tile
}

func settingsFlatDropdownRow(id, title, subtitle string, icon string, dd *ui.Dropdown) *ui.ListTile {
	tile := ui.NewListTile(id, title, subtitle, 0, 0, 0, 0)
	tile.SetLeading(settingsRowIcon(id+"-icon", icon))
	dd.SetStyle("toolbar-menu")
	dd.PreferredWidth = 168
	dd.MaxWidth = 168
	dd.MinWidth = 120
	tile.SetTrailing(dd)
	return tile
}

// settingsPageTitleRow is a caret-back + title row for in-shell settings (no AppBar).
func settingsPageTitleRow(id, title string, onBack func()) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 4
	row.AutoHeight = true
	row.SetStyle("transparent")

	ui.Phosphor.EnsureLoaded(ui.PhosphorCaretLeft, ui.PhosphorRegular)
	back := ui.NewIconButton(id+"-back", "", "", 0, 0, 36, 36)
	back.SetStyle("appbar-icon")
	back.SetPhosphorIcon(ui.PhosphorCaretLeft, ui.PhosphorRegular)
	back.OnClick = onBack
	row.AddChild(back)

	lbl := ui.NewLabel(id+"-title", title, 0, 0, 0, 0)
	lbl.SetStyle("settings-page-title")
	lbl.Align = ui.LabelAlignLeft
	row.AddChild(lbl)
	return row
}

// settingsSection returns top spacing + a section title for settings-style pages.
func settingsSection(id, title string) *ui.Container {
	wrap := ui.NewContainer(id+"-wrap", 0, 0, 0, 0)
	wrap.LayoutType = ui.LayoutFlex
	wrap.FlexDirection = ui.FlexColumn
	wrap.SetStyle("settings-section-wrap")
	spacer := ui.NewContainer(id+"-sp", 0, 0, 0, 16)
	spacer.SetStyle("transparent")
	lbl := ui.NewLabel(id, title, 0, 0, 0, 0)
	lbl.SetStyle("settings-section")
	lbl.Align = ui.LabelAlignLeft
	wrap.AddChild(spacer)
	wrap.AddChild(lbl)
	return wrap
}

