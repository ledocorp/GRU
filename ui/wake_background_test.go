package ui

import "testing"

func TestWakeSummaryForIdlePolicyStripsHover(t *testing.T) {
	var in WakeSummary
	in.Add(WakeOverlay|WakeAnimation|WakeInput, "test")
	out := WakeSummaryForIdlePolicy(in)
	if out.Reasons != (WakeAnimation | WakeInput) {
		t.Fatalf("idle policy wake = %v want animation|input (overlay stripped)", out.Reasons)
	}
}

func TestWakeSummaryForIdlePolicyKeepsAnimation(t *testing.T) {
	var in WakeSummary
	in.Add(WakeAnimation, "spinner")
	out := WakeSummaryForIdlePolicy(in)
	if out.Reasons&WakeAnimation == 0 {
		t.Fatal("WakeAnimation must reach idle policy for AnimationFPS")
	}
}

func TestWakeSummaryForBackgroundStripsInput(t *testing.T) {
	var in WakeSummary
	in.Add(WakeInput|WakeAnimation|WakeOverlay, "test")
	out := WakeSummaryForBackground(in)
	if out.Reasons != 0 {
		t.Fatalf("background wake = %v want 0", out.Reasons)
	}
}

func TestWakeSummaryForBackgroundKeepsResize(t *testing.T) {
	var in WakeSummary
	in.Add(WakeResize, "resize")
	out := WakeSummaryForBackground(in)
	if out.Reasons&WakeResize == 0 {
		t.Fatal("expected resize wake preserved")
	}
}

func TestClearDrawDirtySubtreeClearsRoot(t *testing.T) {
	root := NewContainer("root", 0, 0, 100, 100)
	child := NewLabel("lbl", "hi", 0, 0, 0, 0)
	root.AddChild(child)
	root.MarkDrawDirty()
	child.MarkDrawDirty()
	if !root.DbgDrawDirty() {
		t.Fatal("setup: root draw dirty")
	}
	ClearDrawDirtySubtree(root)
	if root.DbgDrawDirty() || child.DbgDrawDirty() {
		t.Fatal("draw dirty should be cleared")
	}
}
