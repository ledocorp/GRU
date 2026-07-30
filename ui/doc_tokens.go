// Package ui — compile-time design tokens for DocumentSpec (Phase 3).
package ui

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// flexSpacing is padding or gap on DocBlock: a pixel float or a Theme v2 spacing token (e.g. "md").
type flexSpacing float32

func (f flexSpacing) Float() float32 { return float32(f) }

// UnmarshalJSON accepts a number or a spacing token name resolved at compile time.
func (f *flexSpacing) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*f = 0
		return nil
	}
	var num float64
	if err := json.Unmarshal(data, &num); err == nil {
		if num < 0 {
			return fmt.Errorf("spacing must be >= 0")
		}
		*f = flexSpacing(num)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("spacing must be a number or token string")
	}
	v, err := ResolveSpacingToken(s)
	if err != nil {
		return err
	}
	*f = flexSpacing(v)
	return nil
}

func defaultTokenSet() TokenSet {
	return TokenSet{
		Colors: map[string]rl.Color{
			"surface.panel":  CurrentTheme["panel"].BackgroundColor,
			"surface.card":   CurrentTheme["card"].BackgroundColor,
			"accent.primary": CurrentTheme["primary"].BackgroundColor,
			"accent.danger":  CurrentTheme["danger"].BackgroundColor,
			"text.muted":     CurrentTheme["richtext-muted"].TextColor,
			"border.subtle":  CurrentTheme["card"].BorderColor,
		},
		Spacing: map[string]float32{
			"xs": 4,
			"sm": 8,
			"md": 16,
			"lg": 20,
			"xl": 24,
		},
		Radii: map[string]float32{
			"sm": 6,
			"md": 8,
			"lg": 10,
		},
	}
}

// ResolveSpacingToken maps a Theme v2 spacing token name to pixels at compile time.
func ResolveSpacingToken(name string) (float32, error) {
	key := strings.TrimSpace(name)
	if key == "" {
		return 0, fmt.Errorf("empty spacing token")
	}
	if currentThemeV2 != nil {
		if v, ok := currentThemeV2.Tokens.Spacing[key]; ok {
			return v, nil
		}
	}
	return 0, fmt.Errorf("unknown spacing token %q", name)
}

// ResolveColorToken maps a Theme v2 color token name to an rl.Color at compile time.
func ResolveColorToken(name string) (rl.Color, error) {
	key := strings.TrimSpace(name)
	if key == "" {
		return rl.Color{}, fmt.Errorf("empty color token")
	}
	if currentThemeV2 != nil {
		if c, ok := currentThemeV2.Tokens.Colors[key]; ok && c.A != 0 {
			return c, nil
		}
	}
	return rl.Color{}, fmt.Errorf("unknown color token %q", name)
}

// resolveStyleColor parses #RRGGBB hex or a Theme v2 color token reference.
func resolveStyleColor(raw string) (rl.Color, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return rl.Color{}, fmt.Errorf("empty color")
	}
	if strings.HasPrefix(s, "#") {
		return parseHexColor(s)
	}
	return ResolveColorToken(s)
}

// ResolveRadiusToken maps a Theme v2 radii token name to pixels at compile time.
func ResolveRadiusToken(name string) (float32, error) {
	key := strings.TrimSpace(name)
	if key == "" {
		return 0, fmt.Errorf("empty radius token")
	}
	if currentThemeV2 != nil {
		if r, ok := currentThemeV2.Tokens.Radii[key]; ok {
			return r, nil
		}
	}
	return 0, fmt.Errorf("unknown radius token %q", name)
}

func docBlockSpacingFromProps(block DocBlock, key string, fallback float32) float32 {
	if block.Props == nil {
		return fallback
	}
	raw, ok := block.Props[key]
	if !ok {
		return fallback
	}
	switch v := raw.(type) {
	case float32:
		return v
	case float64:
		return float32(v)
	case int:
		return float32(v)
	case string:
		if px, err := ResolveSpacingToken(v); err == nil {
			return px
		}
		if px, err := strconv.ParseFloat(v, 32); err == nil {
			return float32(px)
		}
	}
	return fallback
}
