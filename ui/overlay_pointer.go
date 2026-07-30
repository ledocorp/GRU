// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// overlayExempter is implemented by widgets that own a floating popup. Only those
// popup regions (and the trigger while open) exempt pointer input from blocking.
type overlayExempter interface {
	OverlayExemptRects() []rl.Rectangle
}

// overlayPointerClearer clears hover/press state when a foreign overlay covers
// this widget so stale highlights do not linger under popups.
type overlayPointerClearer interface {
	ClearOverlayPointerState()
}

func overlayExemptRects(n Node) []rl.Rectangle {
	if e, ok := n.(overlayExempter); ok {
		return e.OverlayExemptRects()
	}
	return nil
}

func clearOverlayPointer(n Node) {
	if c, ok := n.(overlayPointerClearer); ok {
		c.ClearOverlayPointerState()
	}
}

// overlayBlockApplies is true only for leaf interactive widgets. Composite shells
// (Panel, Viewport, Form, Container) must keep updating descendants even when a
// floating popup sits over part of their bounds.
func overlayBlockApplies(n Node) bool {
	if !n.IsInteractive() {
		return false
	}
	return len(n.Children()) == 0
}

// UpdateChildrenOverlayAware runs Update on each child, skipping pointer handling
// when a foreign floating overlay covers the cursor.
func UpdateChildrenOverlayAware(children []Node, dt float32) {
	for _, ch := range children {
		UpdateNodeOverlayAware(ch, dt)
	}
}

// UpdateNodeOverlayAware dispatches Update while respecting floating popup occlusion.
func UpdateNodeOverlayAware(n Node, dt float32) {
	if n.IsHidden() {
		return
	}
	if c, ok := n.(*Container); ok {
		UpdateChildrenOverlayAware(c.Children(), dt)
		return
	}
	if overlayBlockApplies(n) {
		mouse := rl.GetMousePosition()
		if WidgetBlockedByOverlay(mouse, overlayExemptRects(n)...) {
			clearOverlayPointer(n)
			// Keep Update (keyboard) for the focused field under a modal/popup.
			if d := ActiveDocument(); d == nil || d.Focused != n {
				return
			}
		}
	}
	n.Update(dt)
}
