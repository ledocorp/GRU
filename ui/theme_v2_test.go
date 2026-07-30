package ui

import "testing"

func TestLoadThemeV2(t *testing.T) {
	previous := CurrentThemeV2()
	defer SetThemeV2(previous)

	data := []byte(`{
		"name": "Test Theme",
		"tokens": {
			"colors": {
				"brand.primary": "#4F46E5",
				"surface.transparent": "#00000000"
			},
			"spacing": {
				"md": 12
			},
			"radii": {
				"control": 6
			}
		},
		"components": {
			"button": {
				"base": {
					"fontSize": 20,
					"padding": 10,
					"cornerRadius": 6
				},
				"variants": {
					"primary": {
						"backgroundColor": "#4F46E5",
						"textColor": "#FFFFFF"
					}
				},
				"states": {
					"hover": {
						"borderColor": "#A5B4FC"
					}
				}
			}
		}
	}`)

	if err := LoadThemeV2(data); err != nil {
		t.Fatalf("LoadThemeV2 returned error: %v", err)
	}

	theme := CurrentThemeV2()
	if theme.Name != "Test Theme" {
		t.Fatalf("theme name = %q, want Test Theme", theme.Name)
	}
	if got := theme.Tokens.Colors["brand.primary"]; got.R != 0x4F || got.G != 0x46 || got.B != 0xE5 || got.A != 0xFF {
		t.Fatalf("brand.primary parsed as %+v", got)
	}
	if got := theme.Tokens.Colors["surface.transparent"]; got.A != 0 {
		t.Fatalf("transparent token alpha = %d, want 0", got.A)
	}
	button := theme.Components["button"]
	if button.Base.FontSize != 20 || button.Base.Padding != 10 || button.Base.CornerRadius != 6 {
		t.Fatalf("button base parsed as %+v", button.Base)
	}
	if got := button.Variants["primary"].BackgroundColor; got.R != 0x4F || got.G != 0x46 || got.B != 0xE5 || got.A != 0xFF {
		t.Fatalf("button primary background parsed as %+v", got)
	}
	if got := button.States["hover"].BorderColor; got.R != 0xA5 || got.G != 0xB4 || got.B != 0xFC || got.A != 0xFF {
		t.Fatalf("button hover border parsed as %+v", got)
	}
}

func TestLoadThemeV2InvalidColor(t *testing.T) {
	if err := LoadThemeV2([]byte(`{
		"name": "Bad Theme",
		"tokens": {
			"colors": {
				"brand.primary": "#XYZ"
			}
		}
	}`)); err == nil {
		t.Fatal("LoadThemeV2 accepted an invalid color")
	}
}

func TestResolveStyleLegacyFallbacks(t *testing.T) {
	previous := CurrentThemeV2()
	defer SetThemeV2(previous)
	SetThemeV2(&ThemeV2{Name: "empty"})

	e := NewElement("legacy", 0, 0, 0, 0)
	e.SetStyleVariant("badge", "primary")

	style := e.GetStyle()
	want := CurrentTheme["badge-primary"].BackgroundColor
	if style.BackgroundColor != want {
		t.Fatalf("legacy component-variant background = %+v, want %+v", style.BackgroundColor, want)
	}
}

func TestBadgeUsesThemeV2Variant(t *testing.T) {
	badge := NewBadge("badge", "Primary", BadgePrimary, 0, 0, 0, 0)

	style := badge.GetStyle()
	want := CurrentTheme["badge-primary"].BackgroundColor
	if style.BackgroundColor != want {
		t.Fatalf("badge primary background = %+v, want %+v", style.BackgroundColor, want)
	}
	if badge.GetPreferredWidth() <= 0 && badge.Bounds().Width <= 0 {
		t.Fatal("auto-sized badge did not compute a width")
	}
}

func TestPanelAndCardUseThemeV2Defaults(t *testing.T) {
	panel := NewPanel("panel", "Panel", 0, 0, 200, 120)
	card := NewCard("card", "Card", 0, 0, 200, 120)

	panelStyle := panel.GetStyle()
	if panelStyle.BackgroundColor != CurrentTheme["panel"].BackgroundColor {
		t.Fatalf("panel background = %+v, want legacy panel background", panelStyle.BackgroundColor)
	}

	cardStyle := card.GetStyle()
	if cardStyle.BackgroundColor != CurrentTheme["card"].BackgroundColor {
		t.Fatalf("card background = %+v, want legacy card background", cardStyle.BackgroundColor)
	}
}

func TestInputAndDropdownUseThemeV2Defaults(t *testing.T) {
	input := NewTextInput("input", "value", 0, 0, 200, 40)
	dropdown := NewDropdown("dropdown", []string{"One", "Two"}, 0, 0, 0, 200, 40)

	inputStyle := input.GetStyle()
	if inputStyle.BackgroundColor != CurrentTheme["input"].BackgroundColor {
		t.Fatalf("input background = %+v, want legacy input background", inputStyle.BackgroundColor)
	}

	dropdownStyle := dropdown.GetStyle()
	if dropdownStyle.BackgroundColor != CurrentTheme["dropdown"].BackgroundColor {
		t.Fatalf("dropdown background = %+v, want legacy dropdown background", dropdownStyle.BackgroundColor)
	}
}

func TestInputAndDropdownResolveThemeV2States(t *testing.T) {
	input := NewTextInput("input", "value", 0, 0, 200, 40)
	focusStyle, inputStateApplied := input.ResolveStyle(StyleStateFocus)
	if !inputStateApplied {
		t.Fatal("input focus state was not applied")
	}
	if focusStyle.BorderColor != CurrentThemeV2().Components["input"].States["focus"].BorderColor {
		t.Fatalf("input focus border = %+v, want Theme v2 focus border", focusStyle.BorderColor)
	}

	dropdown := NewDropdown("dropdown", []string{"One", "Two"}, 0, 0, 0, 200, 40)
	openStyle, dropdownStateApplied := dropdown.ResolveStyle(StyleStateOpen)
	if !dropdownStateApplied {
		t.Fatal("dropdown open state was not applied")
	}
	if openStyle.BackgroundColor != CurrentThemeV2().Components["dropdown"].States["open"].BackgroundColor {
		t.Fatalf("dropdown open background = %+v, want Theme v2 open background", openStyle.BackgroundColor)
	}
}

func TestRemainingFormControlsResolveThemeV2States(t *testing.T) {
	checkbox := NewCheckbox("checkbox", true, 0, 0, 26, 26)
	checkboxStyle, checkboxStateApplied := checkbox.ResolveStyle(StyleStateChecked)
	if !checkboxStateApplied {
		t.Fatal("checkbox checked state was not applied")
	}
	if checkboxStyle.BackgroundColor != CurrentThemeV2().Components["checkbox"].States["checked"].BackgroundColor {
		t.Fatalf("checkbox checked background = %+v, want Theme v2 checked background", checkboxStyle.BackgroundColor)
	}

	toggle := NewToggle("toggle", true, 0, 0, 48, 24)
	toggleStyle, toggleStateApplied := toggle.ResolveStyle(StyleStateChecked)
	if !toggleStateApplied {
		t.Fatal("toggle checked state was not applied")
	}
	if toggleStyle.BackgroundColor != CurrentThemeV2().Components["toggle"].States["checked"].BackgroundColor {
		t.Fatalf("toggle checked background = %+v, want Theme v2 checked background", toggleStyle.BackgroundColor)
	}

	radio := NewRadioGroup("radio", []string{"One"}, 0, 0, 200, 32)
	radioStyle, radioStateApplied := radio.ResolveStyle(StyleStateChecked)
	if !radioStateApplied {
		t.Fatal("radio checked state was not applied")
	}
	if radioStyle.BackgroundColor != CurrentThemeV2().Components["radio"].States["checked"].BackgroundColor {
		t.Fatalf("radio checked background = %+v, want Theme v2 checked background", radioStyle.BackgroundColor)
	}

	slider := NewSlider("slider", 0, 10, 5, 0, 0, 200, 36)
	sliderStyle, sliderStateApplied := slider.ResolveStyle(StyleStateHover)
	if !sliderStateApplied {
		t.Fatal("slider hover state was not applied")
	}
	if sliderStyle.BorderColor != CurrentThemeV2().Components["slider"].States["hover"].BorderColor {
		t.Fatalf("slider hover border = %+v, want Theme v2 hover border", sliderStyle.BorderColor)
	}
}
