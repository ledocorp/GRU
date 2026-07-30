package ui

import "testing"

func TestBuildDocumentSpecCapabilitiesCollapsible(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:  "card",
			ID:    "cap-card",
			Title: "Filters",
			Capabilities: &DocBlockCapabilities{
				Collapsible: boolPtr(true),
				Collapsed:   boolPtr(true),
			},
			Children: []DocBlock{{
				Type: "text",
				ID:   "cap-text",
				Text: "Body",
			}},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	card := node.(*Container).Children()[0].(*Card)
	cb := card.CollapseBehavior()
	if cb == nil {
		t.Fatal("expected CollapseBehavior when collapsible")
	}
	if cb.Expanded.Get() {
		t.Fatal("expected initially collapsed")
	}
}

func TestBuildDocumentSpecCapabilitiesVScroll(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:  "section",
			ID:    "scroll-section",
			Title: "Scrollable",
			Capabilities: &DocBlockCapabilities{
				VScroll: boolPtr(true),
			},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	panel := node.(*Container).Children()[0].(*Panel)
	if !panel.Features().VScroll {
		t.Fatal("expected VScroll enabled on section panel")
	}
}

func TestBuildDocumentSpecSurfaceBlockPanel(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:    "surface",
			ID:      "glass-surface",
			Variant: "panel",
			Title:   "Glass",
			Preset:  "glass-panel",
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	if _, ok := node.(*Container).Children()[0].(*Panel); !ok {
		t.Fatalf("surface variant panel should compile to *Panel, got %T", node.(*Container).Children()[0])
	}
}

func TestBuildDocumentSpecSurfaceBlockFromPreset(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:   "surface",
			ID:     "neo-surface",
			Title:  "Neo",
			Preset: "neo-glow-card",
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	if _, ok := node.(*Container).Children()[0].(*Card); !ok {
		t.Fatalf("neo-glow-card preset should compile to *Card, got %T", node.(*Container).Children()[0])
	}
}

func TestBuildDocumentSpecCapabilitiesOnDismiss(t *testing.T) {
	ctx := NewBuildContext()
	dismissed := false
	ctx.Actions["hideCard"] = func() { dismissed = true }

	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:  "card",
			ID:    "dismiss-card",
			Title: "Closable",
			Capabilities: &DocBlockCapabilities{
				OnDismiss: "hideCard",
			},
		}},
	}, ctx)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	card := node.(*Container).Children()[0].(*Card)
	if card.DismissBehavior() == nil {
		t.Fatal("onDismiss should enable closable dismiss behavior")
	}
	card.Features().OnDismiss()
	if !dismissed {
		t.Fatal("onDismiss action was not wired")
	}
}

func TestBuildDocumentSpecRejectsUnknownCapabilityDragMode(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type: "card",
			ID:   "bad-drag",
			Capabilities: &DocBlockCapabilities{
				DragMode: "invalid",
			},
		}},
	}, nil)
	if err == nil {
		t.Fatal("expected unknown dragMode error")
	}
}

func TestBuildDocumentSpecRejectsUnknownSurfaceVariant(t *testing.T) {
	_, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:    "surface",
			ID:      "bad-surface",
			Variant: "sidebar",
		}},
	}, nil)
	if err == nil {
		t.Fatal("expected unknown surface variant error")
	}
}

func TestBuildDocumentSpecCapabilitiesFromProps(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type: "card",
			ID:   "props-cap",
			Props: map[string]any{
				"capabilities": map[string]any{
					"closable": true,
				},
			},
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	card := node.(*Container).Children()[0].(*Card)
	if card.DismissBehavior() == nil {
		t.Fatal("expected dismiss from props.capabilities")
	}
}

func boolPtr(v bool) *bool { return &v }
