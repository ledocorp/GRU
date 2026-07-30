package ui

import "testing"

func TestResolveChromeProfileNeoGlow(t *testing.T) {
	card := NewCard("c", "T", 0, 0, 100, 80)
	if err := card.SetStylePreset("neo-glow-card", PresetProps{}); err != nil {
		t.Fatal(err)
	}
	if card.PresetName() != "neo-glow-card" {
		t.Fatalf("preset name = %q", card.PresetName())
	}
	if card.ChromeKind() != ChromeNeoGlow {
		t.Fatalf("kind = %v, want ChromeNeoGlow", card.ChromeKind())
	}
	if _, ok := ResolveChromeProfile(&card.Element).(chromeNeoGlowProfile); !ok {
		t.Fatal("expected neo-glow chrome profile")
	}
}

func TestResolveChromeProfileGlassVariant(t *testing.T) {
	p := NewPanel("p", "Glass", 0, 0, 200, 120)
	p.SetStyleVariant("panel", "glass")
	if p.ChromeKind() != ChromeGlass {
		t.Fatalf("kind = %v, want ChromeGlass", p.ChromeKind())
	}
}

func TestResolveChromeProfileDefault(t *testing.T) {
	card := NewCard("c", "", 0, 0, 80, 60)
	if card.ChromeKind() != ChromeDefault {
		t.Fatalf("kind = %v", card.ChromeKind())
	}
}

func TestLoadPresetsJSONBumpsRevision(t *testing.T) {
	before := themeRevisionV2
	if err := LoadPresetsJSON([]byte(`{"x-preset":{"component":"card","variant":"default"}}`)); err != nil {
		t.Fatal(err)
	}
	if themeRevisionV2 <= before {
		t.Fatalf("revision = %d, want > %d", themeRevisionV2, before)
	}
}

func TestPresetHoverLift(t *testing.T) {
	card := NewCard("c", "", 0, 0, 80, 60)
	lift := true
	if err := card.SetStylePreset("neo-glow-card", PresetProps{HoverLift: &lift}); err != nil {
		t.Fatal(err)
	}
	if !card.PresetHoverLift() {
		t.Fatal("expected hover lift")
	}
}
