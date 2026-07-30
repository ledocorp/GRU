// Package ui (continued) — transient draw clip for template/orphan nodes.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// activeDrawClip is the clip rect imposed by the current drawChildrenInRectClip
// pass. Template nodes (VirtualList rows) are not in the widget tree, so Label
// intersects truncate scissors with this rect.
var activeDrawClip rl.Rectangle
var hasActiveDrawClip bool

func setActiveDrawClip(clip rl.Rectangle) {
	if clip.Width > 0 && clip.Height > 0 {
		activeDrawClip = clip
		hasActiveDrawClip = true
		return
	}
	hasActiveDrawClip = false
}

func clearActiveDrawClip() { hasActiveDrawClip = false }

// endNestedScissor closes a widget-local BeginScissorMode without destroying an
// outer clip (toolbar lane, panel body, viewport). Prefer this over bare
// rl.EndScissorMode inside leaf widgets.
func endNestedScissor() {
	if hasActiveDrawClip && activeDrawClip.Width >= 1 && activeDrawClip.Height >= 1 {
		beginScissorFromRect(activeDrawClip)
		return
	}
	rl.EndScissorMode()
}

// effectiveLabelClip returns label bounds intersected with the active draw clip.
func effectiveLabelClip(bounds rl.Rectangle) rl.Rectangle {
	if !hasActiveDrawClip {
		return bounds
	}
	return intersectRects(bounds, activeDrawClip)
}
