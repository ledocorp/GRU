package ui

import "testing"

func TestDismissBehaviorHidesShell(t *testing.T) {
	p := NewPanel("p", "Closable", 0, 0, 320, 200)
	called := false
	p.SetClosable(true)
	p.SetOnDismiss(func() { called = true })
	d := p.DismissBehavior()
	if d == nil {
		t.Fatal("expected DismissBehavior")
	}

	d.Dismiss()
	if !p.IsHidden() {
		t.Fatal("panel should be hidden after dismiss")
	}
	if !called {
		t.Fatal("OnDismiss should run")
	}
}

func TestEscapeBehaviorRequiresDismiss(t *testing.T) {
	p := NewPanel("p", "Esc", 0, 0, 320, 200)
	p.SetClosable(true).SetCloseOnEscape(true)
	if p.escape == nil {
		t.Fatal("expected EscapeBehavior on shell")
	}
	if !p.escape.Enabled {
		t.Fatal("escape should be enabled")
	}
}

func TestCardClosableParity(t *testing.T) {
	c := NewCard("c", "Card", 0, 0, 320, 200)
	if c.panelFeatures == nil {
		t.Fatal("card should attach PanelFeaturesBehavior")
	}
	c.SetCollapsible(true).SetClosable(true)
	if c.CollapseBehavior() == nil {
		t.Fatal("expected collapse on card")
	}
	if c.DismissBehavior() == nil {
		t.Fatal("expected dismiss on card")
	}
	if c.headerMode != HeaderModeInset {
		t.Fatal("card should keep inset header by default")
	}
}

func TestCardSetTitleBarFalse(t *testing.T) {
	c := NewCard("c", "Card", 0, 0, 320, 200)
	c.SetTitleBar(false)
	if c.headerMode != HeaderModeNone {
		t.Fatalf("headerMode = %v, want none", c.headerMode)
	}
}
