package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestLookupPresetDefaults(t *testing.T) {
	p, ok := LookupPreset("neo-glow-card")
	if !ok || p.Component != "card" || p.Variant != "neo-glow" || !p.PinStyle {
		t.Fatalf("neo-glow-card = %+v ok=%v", p, ok)
	}
	if _, ok := LookupPreset("not-a-preset"); ok {
		t.Fatal("expected missing preset")
	}
}

func TestSetStylePresetNeoGlowCard(t *testing.T) {
	card := NewCard("p", "Title", 0, 0, 200, 120)
	glow := float32(0.6)
	if err := card.SetStylePreset("neo-glow-card", PresetProps{GlowIntensity: &glow}); err != nil {
		t.Fatal(err)
	}
	st := card.GetStyle()
	want := CurrentThemeV2().Components["card"].Variants["neo-glow"].BackgroundColor
	if st.BackgroundColor != want {
		t.Fatalf("background = %+v, want neo-glow %+v", st.BackgroundColor, want)
	}
	if st.BorderWidth <= 2 {
		t.Fatalf("expected glowIntensity to widen border, got %.2f", st.BorderWidth)
	}
	if card.PresetGlowIntensity() != glow {
		t.Fatalf("preset glow = %.2f, want %.2f", card.PresetGlowIntensity(), glow)
	}
	if card.ChromeGlowIntensity() != glow {
		t.Fatalf("chrome glow = %.2f, want %.2f", card.ChromeGlowIntensity(), glow)
	}
}

func TestChromeGlowIntensityFromVariant(t *testing.T) {
	card := NewCard("v", "", 0, 0, 100, 80)
	card.SetStyleVariant("card", "neo-glow")
	if g := card.ChromeGlowIntensity(); g <= 0 {
		t.Fatalf("neo-glow variant should default glow, got %.2f", g)
	}
	card.SetStylePreset("neo-glow-card", PresetProps{})
	if g := card.ChromeGlowIntensity(); g <= 0 {
		t.Fatalf("neo-glow-card preset should default glow, got %.2f", g)
	}
	off := float32(0)
	if err := card.SetStylePreset("neo-glow-card", PresetProps{GlowIntensity: &off}); err != nil {
		t.Fatal(err)
	}
	if g := card.ChromeGlowIntensity(); g != 0 {
		t.Fatalf("explicit glowIntensity 0 should disable glow, got %.2f", g)
	}
}

func TestSetStylePresetPrimaryButton(t *testing.T) {
	btn := NewButton("b", "Go", 0, 0, 80, 36)
	if err := btn.SetStylePreset("primary-button", PresetProps{}); err != nil {
		t.Fatal(err)
	}
	st := btn.GetStyle()
	want := CurrentThemeV2().Components["button"].Variants["primary"].BackgroundColor
	if st.BackgroundColor != want {
		t.Fatalf("background = %+v, want primary %+v", st.BackgroundColor, want)
	}
}

func TestPresetPropsFromMap(t *testing.T) {
	props := PresetPropsFromMap(map[string]any{
		"glowIntensity": 0.5,
		"hoverLift":     true,
	})
	if props.GlowIntensity == nil || *props.GlowIntensity != 0.5 {
		t.Fatalf("glow = %v", props.GlowIntensity)
	}
	if props.HoverLift == nil || !*props.HoverLift {
		t.Fatalf("hoverLift = %v", props.HoverLift)
	}
}

func TestLoadPresetsJSON(t *testing.T) {
	const extra = `{
	  "brand-card": { "component": "card", "variant": "default", "pinStyle": false }
	}`
	if err := LoadPresetsJSON([]byte(extra)); err != nil {
		t.Fatal(err)
	}
	p, ok := LookupPreset("brand-card")
	if !ok || p.Component != "card" {
		t.Fatalf("brand-card = %+v ok=%v", p, ok)
	}
}

func TestApplyPresetPropsRadiusScale(t *testing.T) {
	base := Style{CornerRadius: 8, BorderColor: rl.NewColor(10, 10, 10, 255), BorderWidth: 2}
	scale := float32(1.5)
	got := applyPresetProps(base, PresetProps{RadiusScale: &scale})
	if got.CornerRadius != 12 {
		t.Fatalf("radius = %.1f, want 12", got.CornerRadius)
	}
}
