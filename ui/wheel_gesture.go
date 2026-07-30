// Package ui (continued)
package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// wheelGestureWindow is how long a page or nested scroll gesture stays "active"
// after the last wheel tick on that target. Keeps ScrollFPS / wake alive between
// slow ticks so deep idle (10 FPS) does not miss wheel input.
const wheelGestureWindow = 420 * time.Millisecond

var (
	pageWheelGestureUntil    time.Time
	nestedWheelGestureUntil   time.Time
	wheelGestureStickyNested *Viewport
	wheelScrollOwner         *Viewport
	chromeWindowMoving       bool
	chromeTitleBarDragging   bool
	wheelSuppressAboveY      float32
)

// SetWheelSuppressBandY skips wheel routing when the cursor is above y (borderless title bar).
func SetWheelSuppressBandY(y float32) {
	wheelSuppressAboveY = y
}

// SetChromeWindowMoving suppresses wheel scrolling while the borderless title
// bar is dragging or edge-resizing the window (avoids editor scroll racing).
func SetChromeWindowMoving(moving bool) {
	chromeWindowMoving = moving
}

// SetChromeTitleBarDragging tracks borderless title-bar move drags. main.go sets
// this from titleBar.Update each frame. Drives WebViewDeferChromeRaise (§13.5).
func SetChromeTitleBarDragging(dragging bool) {
	chromeTitleBarDragging = dragging
}

// ChromeWindowMoving reports whether wheel scroll should be suppressed for chrome drags.
func ChromeWindowMoving() bool { return chromeWindowMoving }

// WebViewDeferChromeRaise reports whether PresentWebViewHosts should skip raising
// panel-matched Chrome/WebView child HWNDs this frame. True during borderless
// title-bar drag or when the cursor is in the title band.
//
// Band height prefers SetOverlayChromeInsets; falls back to Document.ChromeTop
// then TitleBarHeight so Full Client still defers when insets were missed.
//
// The host stays visible (Show still runs) — only the HWND_TOP raise is deferred.
// Hiding during title-bar drag caused blank web panels (May 2026 regression).
// See docs/WEBVIEW2_HOST.md §13.5.
func WebViewDeferChromeRaise() bool {
	if chromeTitleBarDragging {
		return true
	}
	top := overlayChromeTop
	if top <= 0 {
		if doc := ActiveDocument(); doc != nil && doc.ChromeTop() > 0 {
			top = doc.ChromeTop()
		} else {
			top = TitleBarHeight
		}
	}
	if top > 0 && rl.GetMousePosition().Y < top {
		return true
	}
	return false
}

// ResetWheelScrollGesture clears sticky scroll ownership after chrome gestures.
func ResetWheelScrollGesture() {
	pageWheelGestureUntil = time.Time{}
	nestedWheelGestureUntil = time.Time{}
	wheelGestureStickyNested = nil
	wheelScrollOwner = nil
}

func notePageWheelGesture() {
	pageWheelGestureUntil = time.Now().Add(wheelGestureWindow)
}

func pageWheelGestureActive() bool {
	return time.Now().Before(pageWheelGestureUntil)
}

func noteNestedWheelGesture() {
	nestedWheelGestureUntil = time.Now().Add(wheelGestureWindow)
}

func nestedWheelGestureActive() bool {
	return time.Now().Before(nestedWheelGestureUntil)
}

// ScrollGestureActive reports whether a wheel scroll gesture is still in its
// hold window (page or nested viewport). The main loop uses this to wake
// ScrollFPS before the next frame would otherwise run at DeepIdleFPS.
func ScrollGestureActive() bool {
	return pageWheelGestureActive() || nestedWheelGestureActive()
}

// WheelScrollOwner is the sole vertical scroll target for the current wheel tick.
// Set by PrepareWheelScroll before the widget tree Update pass.
func WheelScrollOwner() *Viewport { return wheelScrollOwner }

// PrepareWheelScroll resolves which Viewport may consume the current mouse wheel
// event. Call once per frame before doc.Root.Update so only one scroll target
// moves per tick (page vs nested region).
func PrepareWheelScroll(root Node) {
	wheelScrollOwner = nil
	if chromeWindowMoving || CalendarPopupOpen() {
		return
	}
	if wheelSuppressAboveY > 0 {
		if rl.GetMousePosition().Y < wheelSuppressAboveY {
			ResetWheelScrollGesture()
			return
		}
	}
	mouse := rl.GetMousePosition()
	if root != nil {
		for _, dd := range collectOpenDropdowns(root.Children()) {
			if rl.CheckCollisionPointRec(mouse, dd.PopupBounds()) {
				return
			}
		}
		for _, cb := range collectOpenComboBoxes(root.Children()) {
			if rl.CheckCollisionPointRec(mouse, cb.PopupBounds()) {
				return
			}
		}
		// Prime scroll target while hovering (before the first wheel tick after focus return).
		if vp := findDeepestVerticalWheelViewportUnderMouse(root, mouse); vp != nil {
			wheelGestureStickyNested = vp
		}
	}
	wheel := rl.GetMouseWheelMove()
	if wheel == 0 {
		return
	}
	if wheelGestureStickyNested != nil && rl.CheckCollisionPointRec(mouse, wheelGestureStickyNested.Bounds()) {
		if wheelGestureStickyNested.Orientation == ScrollVertical &&
			wheelGestureStickyNested.overflowScrollY() > 0 &&
			wheelGestureStickyNested.AbsorbsParentWheel(wheel) {
			wheelScrollOwner = wheelGestureStickyNested
			if wheelGestureStickyNested.isPageLevelVerticalViewport() {
				notePageWheelGesture()
			} else {
				noteNestedWheelGesture()
			}
			return
		}
	}
	page := findPageLevelViewport(root)
	wheelScrollOwner = resolveWheelScrollOwner(page, root, mouse, wheel)
}

func resolveWheelScrollOwner(page *Viewport, root Node, mouse rl.Vector2, wheel float32) *Viewport {
	// Chrome siblings (ribbon toolbar, menubar) sit outside the page viewport but
	// must consume wheel ticks so horizontal ribbon scroll does not also scroll the doc.
	if root != nil && deepHasWheelConsumer(root.Children(), mouse, wheel) {
		wheelGestureStickyNested = nil
		return nil
	}

	// Prefer the scroll viewport under the cursor (split editor/preview panes, nested VPs).
	deepest := findDeepestVerticalWheelViewport(root, mouse, wheel)

	if page != nil && pageWheelGestureActive() {
		// Keep scrolling the page unless the cursor moved to a separate scroll
		// viewport outside the page subtree (e.g. split editor/preview panes).
		if deepest == nil || deepest == page || isNodeDescendantOf(deepest, page) {
			wheelGestureStickyNested = nil
			return page
		}
	}

	if nestedWheelGestureActive() && wheelGestureStickyNested != nil {
		return wheelGestureStickyNested
	}

	if deepest != nil && viewportCanAbsorbWheel(deepest, wheel) {
		wheelGestureStickyNested = deepest
		return deepest
	}

	if page != nil {
		wheelGestureStickyNested = nil
		return page
	}
	wheelGestureStickyNested = nil
	return nil
}

func isNodeDescendantOf(n Node, ancestor Node) bool {
	for p := n.ParentNode(); p != nil; p = p.ParentNode() {
		if p == ancestor {
			return true
		}
	}
	return false
}

func viewportCanAbsorbWheel(vp *Viewport, wheel float32) bool {
	if vp.Orientation != ScrollVertical || vp.overflowScrollY() <= 0 {
		return false
	}
	return vp.AbsorbsParentWheel(wheel)
}

func findPageLevelViewport(root Node) *Viewport {
	var found *Viewport
	var walk func(Node)
	walk = func(n Node) {
		if found != nil || n.IsHidden() {
			return
		}
		if vp, ok := n.(*Viewport); ok && vp.Orientation == ScrollVertical && !vp.hasAncestorVerticalViewport() {
			found = vp
			return
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
	return found
}

func findDeepestVerticalWheelViewport(root Node, mouse rl.Vector2, wheel float32) *Viewport {
	var best *Viewport
	bestDepth := -1
	var walk func(Node)
	walk = func(n Node) {
		if n.IsHidden() {
			return
		}
		if vp, ok := n.(*Viewport); ok && vp.Orientation == ScrollVertical {
			if rl.CheckCollisionPointRec(mouse, vp.Bounds()) && vp.overflowScrollY() > 0 {
				depth := viewportNestDepth(vp)
				if depth > bestDepth {
					best = vp
					bestDepth = depth
				}
			}
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
	return best
}

// findDeepestVerticalWheelViewportUnderMouse selects the nested scroll host under the
// cursor even before overflow is known (e.g. markdown preview still building).
func findDeepestVerticalWheelViewportUnderMouse(root Node, mouse rl.Vector2) *Viewport {
	var best *Viewport
	bestDepth := -1
	var walk func(Node)
	walk = func(n Node) {
		if n.IsHidden() {
			return
		}
		if vp, ok := n.(*Viewport); ok && vp.Orientation == ScrollVertical {
			if rl.CheckCollisionPointRec(mouse, vp.Bounds()) {
				depth := viewportNestDepth(vp)
				if depth > bestDepth {
					best = vp
					bestDepth = depth
				}
			}
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
	return best
}

func viewportNestDepth(vp *Viewport) int {
	depth := 0
	for p := vp.ParentNode(); p != nil; p = p.ParentNode() {
		if _, ok := p.(*Viewport); ok {
			depth++
		}
	}
	return depth
}
