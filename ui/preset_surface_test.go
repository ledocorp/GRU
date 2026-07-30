package ui

import "testing"

func TestVisualSurfacePresetLayoutMetrics(t *testing.T) {
	card := NewCard("c", "neo-glow-card", 0, 0, 200, 0)
	if err := card.SetStylePreset("neo-glow-card", PresetProps{}); err != nil {
		t.Fatal(err)
	}
	if card.TitleHeight != PresetSurfaceTitleHeight {
		t.Fatalf("title height = %v, want %v", card.TitleHeight, PresetSurfaceTitleHeight)
	}
	if card.Gap != PresetSurfaceBodyGap {
		t.Fatalf("gap = %v, want %v", card.Gap, PresetSurfaceBodyGap)
	}
	st := card.GetStyle()
	if st.Padding != PresetSurfacePadding {
		t.Fatalf("padding = %v, want %v", st.Padding, PresetSurfacePadding)
	}
	if st.FontSize != PresetSurfaceFontSize {
		t.Fatalf("font = %v, want %v", st.FontSize, PresetSurfaceFontSize)
	}

	panel := NewPanel("p", "glass-panel", 0, 0, 200, 0)
	if err := panel.SetStylePreset("glass-panel", PresetProps{}); err != nil {
		t.Fatal(err)
	}
	if panel.TitleHeight != PresetSurfaceTitleHeight {
		t.Fatalf("panel title height = %v, want %v", panel.TitleHeight, PresetSurfaceTitleHeight)
	}
}
