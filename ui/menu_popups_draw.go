// Package ui (continued)
package ui

// DrawOpenMenuPopups renders Dropdown and ComboBox popups (and toolbar overflow
// menus) at screen level so they are never clipped by ribbon scissor or nested
// viewports. Call from drawScreenOverlays after the cached UI pass.
func DrawOpenMenuPopups(root Node) {
	if root == nil {
		return
	}
	for _, dd := range collectOpenDropdowns(root.Children()) {
		dd.DrawPopup()
	}
	for _, cb := range collectOpenComboBoxes(root.Children()) {
		cb.DrawPopup()
	}
	for _, tb := range collectOpenToolbars(root.Children()) {
		tb.drawOverflowPopup()
	}
}
