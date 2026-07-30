package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// InteractionOverlayPainter is implemented by widgets whose hover, press, focus,
// or drag visuals can be redrawn on top of the cached SSAA framebuffer without
// invalidating the document cache (MarkDrawDirty).
type InteractionOverlayPainter interface {
	InteractionOverlayActive() bool
	DrawInteractionOverlay()
}

// CollectInteractionOverlayWake reports WakeOverlay while any interaction overlay
// is active so the main loop stays at ActiveFPS during hover exploration.
func CollectInteractionOverlayWake(root Node) WakeSummary {
	var out WakeSummary
	collectInteractionOverlayWake(root, &out)
	return out
}

func collectInteractionOverlayWake(n Node, out *WakeSummary) {
	if n == nil || n.IsHidden() {
		return
	}
	if p, ok := n.(InteractionOverlayPainter); ok && p.InteractionOverlayActive() {
		// TextEditor caret blinks via overlay but must not hold ActiveFPS — see
		// blink → frozen (cache) → hidden in texteditor.go.
		if _, isEditor := n.(*TextEditor); isEditor {
			// still draw overlay below
		} else if ib, isIcon := n.(*IconButton); isIcon && ib.stackedRibbonCell() {
			// ribbon hover chrome at idle FPS is fine
		} else {
			out.Add(WakeOverlay, "interaction-overlay")
		}
	}
	for _, child := range n.Children() {
		collectInteractionOverlayWake(child, out)
	}
}

// DrawInteractionOverlays paints transient interaction chrome after the cached UI
// blit. Uses screen-space scissor when the widget sits inside a viewport.
func DrawInteractionOverlays(root Node) {
	drawInteractionOverlays(root)
}

func drawInteractionOverlays(n Node) {
	if n == nil || n.IsHidden() {
		return
	}
	if p, ok := n.(InteractionOverlayPainter); ok && p.InteractionOverlayActive() {
		clip, hasClip := ancestorClipBounds(n)
		if hasClip {
			if clip.Width >= 1 && clip.Height >= 1 {
				// Use the shared scissor wrapper so clipping stays correct in both
				// 1x fallback and 2x SSAA overlay passes.
				beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
				p.DrawInteractionOverlay()
				rl.EndScissorMode()
			}
		} else {
			p.DrawInteractionOverlay()
		}
	}
	for _, child := range n.Children() {
		drawInteractionOverlays(child)
	}
}
