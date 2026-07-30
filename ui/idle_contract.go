// Package ui — idle / SSAA cache contract for deep FPS (see docs/IDLE_INVARIANTS.md).
package ui

// IdleReady reports whether the visible widget tree can skip a full SSAA redraw
// and blit the cached superTarget. When false, the main loop stays at ActiveFPS.
//
// Equivalent to !SubtreeNeedsRedraw(root). Hidden subtrees are ignored.
func IdleReady(root Node) bool {
	return !SubtreeNeedsRedraw(root)
}

// SubtreeLayoutDirty reports whether any visible node needs Layout (layoutDirty).
// Used when scheduleLayoutPass marks a deep node without bubbling to root.
func SubtreeLayoutDirty(n Node) bool {
	if n == nil || n.IsHidden() {
		return false
	}
	if n.IsDirty() {
		return true
	}
	for _, ch := range n.Children() {
		if SubtreeLayoutDirty(ch) {
			return true
		}
	}
	return false
}

// NotIdleReason returns a compact summary of the first dirty visible nodes, or "" when idle-ready.
func NotIdleReason(root Node) string {
	if IdleReady(root) {
		return ""
	}
	return FormatDirtyReports(CollectDirtyReports(root, 6))
}

// SimulateCacheHitFrame clears draw-only dirty flags after a successful blit or
// full SSAA redraw pass, matching main.go cache-hit and post-redraw handling.
// Layout dirty is unchanged — those nodes still block idle until Layout clears them.
func SimulateCacheHitFrame(root Node) {
	ClearDrawDirtySubtree(root)
}

// AssertIdleReady fails tests when the tree cannot reach deep idle after layout settle.
func AssertIdleReady(t testingTB, root Node, context string) {
	t.Helper()
	if IdleReady(root) {
		return
	}
	t.Fatalf("%s: tree not idle-ready: %s", context, NotIdleReason(root))
}

// testingTB is the subset of testing.T used by AssertIdleReady.
type testingTB interface {
	Helper()
	Fatalf(format string, args ...any)
}
