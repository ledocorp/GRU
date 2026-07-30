package ui

import (
	"encoding/json"
	"testing"
)

func TestFlexSpacingUnmarshalToken(t *testing.T) {
	var block DocBlock
	if err := json.Unmarshal([]byte(`{"type":"card","padding":"md","gap":"sm"}`), &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := block.Padding.Float(); got != 16 {
		t.Fatalf("padding = %.0f, want 16 (md)", got)
	}
	if got := block.Gap.Float(); got != 8 {
		t.Fatalf("gap = %.0f, want 8 (sm)", got)
	}
}

func TestFlexSpacingUnmarshalNumber(t *testing.T) {
	var block DocBlock
	if err := json.Unmarshal([]byte(`{"type":"card","padding":14}`), &block); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got := block.Padding.Float(); got != 14 {
		t.Fatalf("padding = %.0f, want 14", got)
	}
}

func TestDocBlockStyleColorToken(t *testing.T) {
	bg := "surface.card"
	border := "border.subtle"
	st, err := (&DocBlockStyle{BackgroundColor: &bg, BorderColor: &border}).compile()
	if err != nil {
		t.Fatal(err)
	}
	wantBg := CurrentTheme["card"].BackgroundColor
	if st.BackgroundColor != wantBg {
		t.Fatalf("background = %+v, want %+v", st.BackgroundColor, wantBg)
	}
}

func TestBuildDocumentSpecBadgeBlock(t *testing.T) {
	node, err := BuildDocumentSpec(DocumentSpec{
		Children: []DocBlock{{
			Type:    "badge",
			ID:      "tag",
			Text:    "Beta",
			Variant: "primary",
		}},
	}, nil)
	if err != nil {
		t.Fatalf("BuildDocumentSpec: %v", err)
	}
	badge := node.(*Container).Children()[0].(*Badge)
	if badge.Text.Get() != "Beta" {
		t.Fatalf("text = %q, want Beta", badge.Text.Get())
	}
	if badge.Variant != BadgePrimary {
		t.Fatalf("variant = %v, want primary", badge.Variant)
	}
}
