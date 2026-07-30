package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// AnimationOverlayDrawer is implemented by widgets whose time-varying pixels can
// be drawn after the cached document blit. This keeps the main UI render texture
// clean while small animations continue to move.
type AnimationOverlayDrawer interface {
	AnimationReporter
	DrawAnimationOverlay()
}

// DrawAnimationOverlays draws active animation overlays in tree order. Call this
// after the cached UI has been presented and before screen-level managers such
// as tooltips, modals, and toasts.
//
// screenExclusions are extra screen rects (e.g. launcher nav bar) where animation
// pixels must not paint — keeps chrome in the SSAA cache instead of repainting at 1×.
func DrawAnimationOverlays(root Node, screenExclusions ...rl.Rectangle) {
	exclusions := collectChromeExclusions(root)
	exclusions = append(exclusions, screenExclusions...)
	drawAnimationOverlays(root, exclusions)
}

func collectChromeExclusions(root Node) []rl.Rectangle {
	var rects []rl.Rectangle
	collectChromeExclusionsWalk(root, &rects)
	return rects
}

func collectChromeExclusionsWalk(n Node, rects *[]rl.Rectangle) {
	if n == nil || n.IsHidden() {
		return
	}
	if _, ok := n.(*BottomNavigationBar); ok {
		*rects = append(*rects, n.Bounds())
		return
	}
	if c, ok := n.(*Container); ok && c.StyleName() == "appshell-footer" {
		*rects = append(*rects, c.Bounds())
		return
	}
	for _, child := range n.Children() {
		collectChromeExclusionsWalk(child, rects)
	}
}

// clipExcludeRegions shrinks clip so post-blit animation pixels do not paint
// over pinned shell chrome (bottom nav, footer, launcher nav).
func clipExcludeRegions(clip rl.Rectangle, exclusions []rl.Rectangle) rl.Rectangle {
	out := clip
	for _, ex := range exclusions {
		if ex.Width < 1 || ex.Height < 1 {
			continue
		}
		if !rectsOverlap(out, ex) {
			continue
		}
		// Chrome band at the bottom of the clip (typical bottom nav / footer / nav strip).
		if ex.Y > out.Y && ex.Y < out.Y+out.Height {
			if h := ex.Y - out.Y; h < out.Height {
				out.Height = h
			}
		}
	}
	if out.Height < 0 {
		out.Height = 0
	}
	return out
}

func rectsOverlap(a, b rl.Rectangle) bool {
	return a.X < b.X+b.Width && a.X+a.Width > b.X &&
		a.Y < b.Y+b.Height && a.Y+a.Height > b.Y
}

func drawAnimationOverlays(n Node, exclusions []rl.Rectangle) {
	if n == nil || n.IsHidden() {
		return
	}
	if drawer, ok := n.(AnimationOverlayDrawer); ok && drawer.AnimationActive() {
		clip, hasClip := ancestorClipBounds(n)
		if !hasClip {
			clip = n.Bounds()
		}
		clip = clipExcludeRegions(clip, exclusions)
		if clip.Width >= 1 && clip.Height >= 1 {
			rl.BeginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
			drawer.DrawAnimationOverlay()
			rl.EndScissorMode()
		}
	}
	for _, child := range n.Children() {
		drawAnimationOverlays(child, exclusions)
	}
}
