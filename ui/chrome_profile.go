// Package ui (continued) — Visual Presets chrome profiles (Phase 2).
//
// ChromeProfile draws extras beyond flat Style fields: shadow stacks, glow halos,
// glass sheen. Resolved from preset name or Theme v2 component/variant.
// See docs/VISUAL_PRESETS_PLAN.md §5 Phase 2.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ChromeKind identifies which chrome profile resolves for an element.
type ChromeKind int

const (
	ChromeDefault ChromeKind = iota
	ChromeNeoGlow
	ChromeGlass
)

// SurfaceChromeCtx carries resolved draw inputs for Card/Panel chrome.
type SurfaceChromeCtx struct {
	Bounds          rl.Rectangle
	Style           Style
	CornerRadius    float32
	Roundness       float32
	GlowIntensity   float32
	HoverLift       bool
	SkipOuterShadow bool
}

// ChromeProfile draws chrome beyond flat Style fills and borders.
type ChromeProfile interface {
	DrawShadow(ctx SurfaceChromeCtx)
	// DrawOverFill runs after the surface fill, before the border stroke.
	DrawOverFill(ctx SurfaceChromeCtx)
	// DrawPostBorder runs after the border stroke (glow rings, etc.).
	DrawPostBorder(ctx SurfaceChromeCtx)
}

// BakedChromeProfile is an optional extension for static chrome baked to a texture.
// Full async bake is Phase 4; profiles implement this only when ready to cache.
type BakedChromeProfile interface {
	ChromeProfile
	WantsBake() bool
}

type chromeDefaultProfile struct{}

func (chromeDefaultProfile) DrawShadow(ctx SurfaceChromeCtx) {
	if ctx.SkipOuterShadow {
		return
	}
	drawRaisedSurfaceShadow(ctx.Bounds, ctx.Roundness)
}

func (chromeDefaultProfile) DrawOverFill(_ SurfaceChromeCtx)  {}
func (chromeDefaultProfile) DrawPostBorder(_ SurfaceChromeCtx) {}

type chromeNeoGlowProfile struct{}

func (chromeNeoGlowProfile) DrawShadow(ctx SurfaceChromeCtx) {
	if ctx.SkipOuterShadow {
		return
	}
	drawRaisedSurfaceShadow(ctx.Bounds, ctx.Roundness)
	intensity := ctx.GlowIntensity
	if intensity <= 0 {
		intensity = 0.5
	}
	drawNeoGlowAmbientShadow(ctx.Bounds, ctx.Roundness, intensity, ctx.HoverLift)
}

func (chromeNeoGlowProfile) DrawOverFill(_ SurfaceChromeCtx) {}

func (chromeNeoGlowProfile) DrawPostBorder(ctx SurfaceChromeCtx) {
	intensity := ctx.GlowIntensity
	if intensity <= 0 {
		intensity = 0.5
	}
	drawSurfacePresetGlow(ctx.Bounds, ctx.CornerRadius, intensity, ctx.Style.BorderWidth)
}

func (chromeNeoGlowProfile) WantsBake() bool { return false }

type chromeGlassProfile struct{}

func (chromeGlassProfile) DrawShadow(ctx SurfaceChromeCtx) {
	if ctx.SkipOuterShadow {
		return
	}
	drawRaisedSurfaceShadow(ctx.Bounds, ctx.Roundness)
}

func (chromeGlassProfile) DrawOverFill(ctx SurfaceChromeCtx) {
	drawGlassPanelSheen(ctx.Bounds, ctx.CornerRadius, ctx.Style)
}

func (chromeGlassProfile) DrawPostBorder(_ SurfaceChromeCtx) {}

func (chromeGlassProfile) WantsBake() bool { return false }

var (
	chromeProfileDefault = chromeDefaultProfile{}
	chromeProfileNeoGlow = chromeNeoGlowProfile{}
	chromeProfileGlass   = chromeGlassProfile{}
)

// ChromeKind resolves the chrome profile key from preset + Theme v2 fields.
func (e *Element) ChromeKind() ChromeKind {
	switch e.presetName {
	case "neo-glow-card":
		return ChromeNeoGlow
	case "glass-panel":
		return ChromeGlass
	case "glass-panel-dark":
		return ChromeGlass
	case "glass-card":
		return ChromeGlass
	}
	if e.styleComponent == "card" && e.styleVariant == "neo-glow" {
		return ChromeNeoGlow
	}
	if e.styleComponent == "card" && e.styleVariant == "glass" {
		return ChromeGlass
	}
	if e.styleComponent == "panel" && (e.styleVariant == "glass" || e.styleVariant == "glass-dark") {
		return ChromeGlass
	}
	return ChromeDefault
}

// ResolveChromeProfile returns the draw profile for an element.
func ResolveChromeProfile(e *Element) ChromeProfile {
	switch e.ChromeKind() {
	case ChromeNeoGlow:
		return chromeProfileNeoGlow
	case ChromeGlass:
		return chromeProfileGlass
	default:
		return chromeProfileDefault
	}
}

// PresetHoverLift reports whether the preset requested a static lifted shadow.
func (e *Element) PresetHoverLift() bool { return e.presetHoverLift }

// PresetName returns the last applied visual preset name, or "".
func (e *Element) PresetName() string { return e.presetName }

// SurfaceChromeCtxFor builds a chrome context from element state and layout bounds.
func SurfaceChromeCtxFor(e *Element, bounds rl.Rectangle, style Style, skipOuterShadow bool) SurfaceChromeCtx {
	return SurfaceChromeCtx{
		Bounds:          bounds,
		Style:           style,
		CornerRadius:    style.CornerRadius,
		Roundness:       chromeRoundness(bounds, style.CornerRadius),
		GlowIntensity:   e.ChromeGlowIntensity(),
		HoverLift:       e.PresetHoverLift(),
		SkipOuterShadow: skipOuterShadow,
	}
}
