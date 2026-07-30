// Package ui (continued) — Card surface features facade (Phase C3 parity with Panel).
package ui

// Features returns the live feature set for programmatic toggling.
func (c *Card) Features() *PanelFeatures {
	return surfaceFeaturesConfig(&c.SurfaceShell)
}

// SyncFeatures reapplies the current feature flags.
func (c *Card) SyncFeatures() {
	surfaceSyncFeatures(&c.SurfaceShell)
}

// SetTitleBar toggles the title band (inset header vs none for Card).
func (c *Card) SetTitleBar(on bool) *Card {
	c.Features().TitleBar = on
	c.SyncFeatures()
	return c
}

// SetCollapsible toggles collapse control in the header chrome.
func (c *Card) SetCollapsible(on bool) *Card {
	c.Features().Collapsible = on
	c.SyncFeatures()
	return c
}

// SetClosable toggles the × dismiss control.
func (c *Card) SetClosable(on bool) *Card {
	c.Features().Closable = on
	c.SyncFeatures()
	return c
}

// SetCloseOnEscape toggles Escape-key dismiss when the pointer is over the card.
func (c *Card) SetCloseOnEscape(on bool) *Card {
	c.Features().CloseOnEscape = on
	c.SyncFeatures()
	return c
}

// SetOnDismiss sets the callback invoked when the user dismisses the card.
func (c *Card) SetOnDismiss(fn func()) *Card {
	c.Features().OnDismiss = fn
	if d := c.DismissBehavior(); d != nil {
		d.OnDismiss = fn
	}
	return c
}

// SetVScroll wraps the body in a vertical scroll viewport when true.
func (c *Card) SetVScroll(on bool) *Card {
	c.Features().VScroll = on
	c.SyncFeatures()
	return c
}

// SetHScroll wraps the body in a horizontal scroll viewport when true.
func (c *Card) SetHScroll(on bool) *Card {
	c.Features().HScroll = on
	c.SyncFeatures()
	return c
}

// CollapseBehavior returns the collapse plugin when Collapsible is enabled.
func (c *Card) CollapseBehavior() *CollapseBehavior {
	return surfaceCollapseBehavior(&c.SurfaceShell)
}

// DismissBehavior returns the dismiss plugin when Closable is enabled.
func (c *Card) DismissBehavior() *DismissBehavior {
	return surfaceDismissBehavior(&c.SurfaceShell)
}

// EnableCollapse enables Collapsible and returns the collapse plugin.
func (c *Card) EnableCollapse(initialExpanded bool) *CollapseBehavior {
	c.SetCollapsible(true)
	cb := c.CollapseBehavior()
	if cb != nil {
		cb.Expanded.Set(initialExpanded)
	}
	return cb
}

// SetHeaderMode sets the surface title band style.
func (c *Card) SetHeaderMode(mode SurfaceHeaderMode) {
	c.headerMode = mode
	c.MarkDirty()
}
