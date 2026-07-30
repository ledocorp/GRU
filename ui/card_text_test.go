package ui

import "testing"

func TestApplyCardBodyTextColorNeoGlow(t *testing.T) {
	card := NewCard("c", "Title", 0, 0, 200, 120)
	glow := float32(0.5)
	if err := card.SetStylePreset("neo-glow-card", PresetProps{GlowIntensity: &glow}); err != nil {
		t.Fatal(err)
	}
	lbl := NewLabel("body", "Hello", 0, 0, 0, 0)
	lbl.SetStyle("form-value")
	card.AddChild(lbl)

	chrome := card.GetStyle()
	wantColor := chrome.TextColor
	got := lbl.GetStyle()
	if got.TextColor != wantColor {
		t.Fatalf("label text color = %+v, want %+v", got.TextColor, wantColor)
	}
	if chrome.FontSize > 0 && got.FontSize != chrome.FontSize {
		t.Fatalf("label font size = %v, want chrome %v", got.FontSize, chrome.FontSize)
	}
	if lbl.StyleName() != "form-value" {
		t.Fatalf("style = %q, want form-value preserved", lbl.StyleName())
	}
}

func TestDefaultPanelDoesNotSyncBodyText(t *testing.T) {
	panel := NewPanel("p", "Regular", 0, 0, 200, 120)
	lbl := NewLabel("body", "Hello", 0, 0, 0, 0)
	lbl.SetStyle("form-value")
	panel.AddChild(lbl)

	got := lbl.GetStyle()
	if got.TextColor == panel.GetStyle().TextColor && got.FontSize == panel.GetStyle().FontSize {
		t.Fatal("default panel should not override label typography")
	}
}

func TestApplyPanelBodyTextTypographyAfterPreset(t *testing.T) {
	panel := NewPanel("p", "Glass", 0, 0, 200, 120)
	if err := panel.SetStylePreset("glass-panel-dark", PresetProps{}); err != nil {
		t.Fatal(err)
	}
	lbl := NewLabel("body", "Hello", 0, 0, 0, 0)
	panel.AddChild(lbl)

	chrome := panel.GetStyle()
	got := lbl.GetStyle()
	if got.TextColor != chrome.TextColor {
		t.Fatalf("label color = %+v, want %+v", got.TextColor, chrome.TextColor)
	}
	if chrome.FontSize > 0 && got.FontSize != chrome.FontSize {
		t.Fatalf("label font size = %v, want %v", got.FontSize, chrome.FontSize)
	}
}

func TestApplyCardBodyTextColorRichText(t *testing.T) {
	card := NewCard("c", "", 0, 0, 200, 100)
	card.SetStyleVariant("card", "neo-glow")
	rt := NewRichText("t", []TextSpan{{Text: "Hi"}}, 0, 0, 0, 0)
	card.AddChild(rt)

	chrome := card.GetStyle()
	got := rt.GetStyle()
	if got.TextColor != chrome.TextColor {
		t.Fatalf("richtext color = %+v, want %+v", got.TextColor, chrome.TextColor)
	}
	if chrome.FontSize > 0 && got.FontSize != chrome.FontSize {
		t.Fatalf("richtext font size = %v, want %v", got.FontSize, chrome.FontSize)
	}
	if rt.StyleName() != "richtext-on-dark" {
		t.Fatalf("style = %q, want richtext-on-dark", rt.StyleName())
	}
}
