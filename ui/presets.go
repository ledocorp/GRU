// Package ui (continued) — Visual Presets (Theme v2 named recipes).
//
// Presets map a stable name to component + variant (+ optional pinned Style).
// They resolve through Element.SetStylePreset → ResolveStyle like component/variant.
// See docs/VISUAL_PRESETS_PLAN.md.
package ui

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// VisualPreset is a named Theme v2 recipe.
type VisualPreset struct {
	Name      string
	Component string
	Variant   string
	// PinStyle applies MergedComponentVariantStyle as overrides so tinted card
	// chrome stays visible when legacy styleName would otherwise win.
	PinStyle bool
}

// PresetProps holds optional parametric tweaks merged onto pinned styles.
type PresetProps struct {
	GlowIntensity *float32
	HoverLift     *bool
	RadiusScale   *float32
}

var presetRegistry = defaultPresetRegistry()

func defaultPresetRegistry() map[string]VisualPreset {
	return map[string]VisualPreset{
		"surface-card":  {Name: "surface-card", Component: "card", Variant: "default"},
		"callout-tip":   {Name: "callout-tip", Component: "card", Variant: "callout", PinStyle: true},
		"code-block":    {Name: "code-block", Component: "card", Variant: "code", PinStyle: true},
		"primary-button": {Name: "primary-button", Component: "button", Variant: "primary"},
		"ghost-button":  {Name: "ghost-button", Component: "button", Variant: "ghost"},
		"danger-button": {Name: "danger-button", Component: "button", Variant: "danger"},
		"glass-panel":      {Name: "glass-panel", Component: "panel", Variant: "glass", PinStyle: true},
		"glass-panel-dark": {Name: "glass-panel-dark", Component: "panel", Variant: "glass-dark", PinStyle: true},
		"glass-card":       {Name: "glass-card", Component: "card", Variant: "glass", PinStyle: true},
		"neo-glow-card": {Name: "neo-glow-card", Component: "card", Variant: "neo-glow", PinStyle: true},
	}
}

// LookupPreset returns a registered visual preset by name.
func LookupPreset(name string) (VisualPreset, bool) {
	p, ok := presetRegistry[name]
	return p, ok
}

// ListPresetNames returns sorted preset names (for tooling/tests).
func ListPresetNames() []string {
	out := make([]string, 0, len(presetRegistry))
	for name := range presetRegistry {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

// LoadPresetsJSON merges preset definitions from JSON into the active registry.
// Shape: { "neo-glow-card": { "component": "card", "variant": "neo-glow", "pinStyle": true } }
func LoadPresetsJSON(data []byte) error {
	var raw map[string]struct {
		Component string `json:"component"`
		Variant   string `json:"variant"`
		PinStyle  bool   `json:"pinStyle"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("ui/presets: %w", err)
	}
	for name, entry := range raw {
		if entry.Component == "" || entry.Variant == "" {
			return fmt.Errorf("ui/presets: preset %q needs component and variant", name)
		}
		presetRegistry[name] = VisualPreset{
			Name:      name,
			Component: entry.Component,
			Variant:   entry.Variant,
			PinStyle:  entry.PinStyle,
		}
	}
	bumpPresetsRevision()
	return nil
}

// bumpPresetsRevision invalidates resolved-style caches after preset registry changes.
func bumpPresetsRevision() {
	themeRevisionV2++
}

// SetStylePreset applies a named visual preset and optional parametric props.
func (e *Element) SetStylePreset(name string, props PresetProps) error {
	p, ok := LookupPreset(name)
	if !ok {
		return fmt.Errorf("ui/presets: unknown preset %q", name)
	}
	e.presetGlowIntensity = 0
	e.presetGlowSet = props.GlowIntensity != nil
	if props.GlowIntensity != nil {
		e.presetGlowIntensity = clampFloat(*props.GlowIntensity, 0, 1)
	} else if name == "neo-glow-card" {
		e.presetGlowIntensity = 0.45
	}
	e.presetHoverLift = props.HoverLift != nil && *props.HoverLift
	e.SetStyleVariant(p.Component, p.Variant)
	e.presetName = name
	if p.PinStyle {
		st, ok := MergedComponentVariantStyle(p.Component, p.Variant)
		if !ok {
			return fmt.Errorf("ui/presets: preset %q variant not in theme", name)
		}
		st = normalizeVisualSurfaceStyle(st)
		st = applyPresetProps(st, props)
		e.SetStyleOverrides(st)
		e.MarkDrawDirty()
		return nil
	}
	if props.hasAny() {
		e.SetStyleOverrides(applyPresetProps(Style{}, props))
	}
	e.MarkDrawDirty()
	return nil
}

func (p PresetProps) hasAny() bool {
	return p.GlowIntensity != nil || p.HoverLift != nil || p.RadiusScale != nil
}

// PresetPropsFromMap parses DocumentSpec block props into PresetProps.
func PresetPropsFromMap(m map[string]any) PresetProps {
	if len(m) == 0 {
		return PresetProps{}
	}
	var props PresetProps
	if v, ok := mapFloat(m, "glowIntensity"); ok {
		props.GlowIntensity = &v
	}
	if v, ok := mapBool(m, "hoverLift"); ok {
		props.HoverLift = &v
	}
	if v, ok := mapFloat(m, "radiusScale"); ok {
		props.RadiusScale = &v
	}
	return props
}

func applyPresetProps(base Style, props PresetProps) Style {
	st := base
	if props.GlowIntensity != nil {
		g := clampFloat(*props.GlowIntensity, 0, 1)
		if g > 0 {
			bc := st.BorderColor
			st.BorderColor = rl.NewColor(
				lerpU8(bc.R, 129, g*0.35),
				lerpU8(bc.G, 140, g*0.35),
				lerpU8(bc.B, 248, g*0.35),
				bc.A,
			)
			st.BorderWidth += g * 1.5
		}
	}
	if props.RadiusScale != nil {
		s := clampFloat(*props.RadiusScale, 0.5, 2)
		if st.CornerRadius > 0 {
			st.CornerRadius *= s
		} else {
			st.CornerRadius = 8 * s
		}
	}
	if props.HoverLift != nil && *props.HoverLift && st.CornerRadius > 0 {
		st.CornerRadius += 1
	}
	return st
}

func clampFloat(v, lo, hi float32) float32 {
	return float32(math.Max(float64(lo), math.Min(float64(hi), float64(v))))
}

func lerpU8(a, b uint8, t float32) uint8 {
	return uint8(float32(a) + (float32(b)-float32(a))*t)
}

func mapFloat(m map[string]any, key string) (float32, bool) {
	v, ok := m[key]
	if !ok {
		return 0, false
	}
	switch n := v.(type) {
	case float64:
		return float32(n), true
	case float32:
		return n, true
	case int:
		return float32(n), true
	case json.Number:
		f, err := n.Float64()
		if err != nil {
			return 0, false
		}
		return float32(f), true
	default:
		return 0, false
	}
}

func mapBool(m map[string]any, key string) (bool, bool) {
	v, ok := m[key]
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	return b, ok
}

// SetStylePreset applies a named visual preset to the button chrome.
func (b *Button) SetStylePreset(name string, props PresetProps) error {
	return b.Element.SetStylePreset(name, props)
}
