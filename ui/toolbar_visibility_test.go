package ui

import "testing"

func TestSyncRibbonItemVisibilityNoRepeatShow(t *testing.T) {
	tb := NewToolbar("ribbon", 0, 0, 400, 120)
	tb.Ribbon = true
	tb.RibbonStacked = true
	tb.ribbonTabNames = []string{"Home"}
	btn := NewIconButton("save", "", "Save", 0, 0, 72, 72)
	btn.Stacked = true
	btn.styleName = "toolbar-cell"
	tb.Groups = []*ToolbarGroup{{
		Label:    "Home",
		TabIndex: 0,
		items: []*ToolbarItem{{
			id:       "save",
			itemType: ToolbarItemIconButton,
			widget:   btn,
		}},
	}}
	tb.layoutDirty = false
	tb.drawDirty = false
	btn.layoutDirty = false
	btn.drawDirty = false

	tb.syncRibbonItemVisibility()
	if btn.IsDirty() || btn.DbgDrawDirty() {
		t.Fatal("syncRibbonItemVisibility must not Show() already-visible widgets")
	}

	tb.syncRibbonItemVisibility()
	if tb.IsDirty() || tb.DbgDrawDirty() {
		t.Fatal("second sync must stay clean")
	}
}
