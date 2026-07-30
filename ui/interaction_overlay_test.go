package ui

import "testing"

func TestStackedRibbonCellHoverUsesOverlayNotCacheDirty(t *testing.T) {
	btn := NewIconButton("save", "", "Save", 0, 0, 72, 72)
	btn.Stacked = true
	btn.styleName = "toolbar-cell"
	btn.hovered = true
	if !btn.stackedRibbonCell() {
		t.Fatal("expected stacked ribbon cell")
	}
	if btn.InteractionOverlayActive() {
		t.Fatal("stacked ribbon hover should redraw in SSAA cache, not overlay")
	}
}

func TestStackedRibbonCellCheckedRedrawsInCache(t *testing.T) {
	btn := NewIconButton("status", "", "Status", 0, 0, 72, 72)
	btn.Stacked = true
	btn.styleName = "toolbar-cell"
	btn.Checked = NewSignal(true)
	if btn.InteractionOverlayActive() {
		t.Fatal("stacked ribbon checked chrome belongs in SSAA cache")
	}
}

func TestButtonHoverMarksDrawDirtyNotOverlay(t *testing.T) {
	b := NewButton("b", "Go", 0, 0, 80, 36)
	b.hovered = true
	if b.InteractionOverlayActive() {
		t.Fatal("hover should redraw in SSAA cache, not interaction overlay")
	}
}

func TestCollectInteractionOverlayWakeEmpty(t *testing.T) {
	root := NewContainer("root", 0, 0, 100, 100)
	wake := CollectInteractionOverlayWake(root)
	if wake.Any() {
		t.Fatalf("expected no wake on idle tree, got %v", wake.Reasons)
	}
}

func TestInteractionOverlayWakeSkipsTextEditorCaret(t *testing.T) {
	ed := NewTextEditor("ed", "x", 0, 0, 100, 40)
	ed.focused = true
	ed.caretPhase = caretPhaseBlink
	ed.blinkPhase = 0
	if !ed.InteractionOverlayActive() {
		t.Fatal("expected caret overlay active on blink visible half")
	}
	root := NewContainer("root", 0, 0, 200, 200)
	root.AddChild(ed)
	wake := CollectInteractionOverlayWake(root)
	if wake.Any() {
		t.Fatalf("TextEditor caret must not wake ActiveFPS: %v", wake.Reasons)
	}
}

func TestTextEditorFrozenCaretUsesCacheNotOverlay(t *testing.T) {
	ed := NewTextEditor("ed", "hello", 0, 0, 100, 40)
	ed.focused = true
	ed.caretPhase = caretPhaseFrozen
	if ed.InteractionOverlayActive() {
		t.Fatal("frozen caret must bake into SSAA cache so idle FPS can drop")
	}
}

func TestInteractionOverlayWakeSkipsStackedRibbonCell(t *testing.T) {
	root := NewContainer("root", 0, 0, 200, 200)
	for _, setup := range []func(*IconButton){
		func(btn *IconButton) { btn.hovered = true },
		func(btn *IconButton) { btn.Checked = NewSignal(true) },
	} {
		btn := NewIconButton("save", "", "Save", 0, 0, 72, 72)
		btn.Stacked = true
		btn.styleName = "toolbar-cell"
		setup(btn)
		root.AddChild(btn)
		wake := CollectInteractionOverlayWake(root)
		if wake.Any() {
			t.Fatalf("stacked ribbon overlay must not wake ActiveFPS: %v", wake.Reasons)
		}
		root.RemoveChild(btn.ID())
	}
}

func TestInteractionOverlayWakeIconHover(t *testing.T) {
	btn := NewIconButton("btn", "", "Save", 0, 0, 40, 40)
	btn.hovered = true
	root := NewContainer("root", 0, 0, 200, 200)
	root.AddChild(btn)
	wake := CollectInteractionOverlayWake(root)
	if !wake.Any() {
		t.Fatal("hovered icon should wake overlay exploration")
	}
}

func TestTextInputNoInitialSelection(t *testing.T) {
	ti := NewTextInput("ti", "Jane Doe", 0, 0, 200, 32)
	if ti.hasSelection() {
		t.Fatal("new TextInput with text must not start with a selection")
	}
	if ti.selAnchor != -1 {
		t.Fatalf("selAnchor = %d, want -1", ti.selAnchor)
	}
}

func TestTextInputSelectionRange(t *testing.T) {
	ti := NewTextInput("ti", "hello", 0, 0, 100, 32)
	ti.selAnchor = 1
	ti.cursor = 4
	if got := ti.selectedText(); got != "ell" {
		t.Fatalf("selectedText = %q, want ell", got)
	}
	ti.selAnchor = 4
	ti.cursor = 1
	if got := ti.selectedText(); got != "ell" {
		t.Fatalf("reversed selection = %q, want ell", got)
	}
}
