// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── Spinner ─────────────────────────────────────────────────────────────────
//
// Spinner is an animated loading indicator that supports two modes:
//
//   - Indeterminate (default): a 270° arc rotates continuously at Speed °/s.
//     The arc length breathes via a sinusoidal pulse so the animation feels
//     alive rather than mechanical.
//   - Determinate: Progress (0.0–1.0) drives a static filled arc. A percentage
//     label is drawn at the centre of the ring.
//
// An optional text label can be positioned below (LabelBelow, default) or to
// the right (LabelRight) of the ring. Leave Label empty to omit it.
//
// # Usage
//
//	// Indeterminate spinner
//	sp := ui.NewSpinner("loading", 40, 200, 200)
//	sp.Active.Set(true)
//
//	// Determinate spinner with label
//	sp := ui.NewSpinner("upload", 60, 200, 200)
//	sp.Determinate = true
//	sp.Progress.Set(0.42)
//	sp.Label = "Uploading…"
//	sp.LabelPos = ui.LabelBelow
//	sp.Active.Set(true)
//
// # Sizing
//
// The size parameter is the outer diameter of the ring. The widget height is
// size when LabelBelow is used (the label extends below the bounding box of
// the ring), or size when LabelRight (label sits beside the ring at mid-height).
// Adjust the node height via SetBounds after construction if you need extra
// space for the label.
//
// # Style
//
// Styled via the "spinner" theme key (transparent background). The ring colour
// is controlled by the Color field (default: indigo #4F46E5). The track
// (background ring) is always a desaturated grey.
//
// # Main-loop integration
//
//	// In Build: add to any container/panel/card/modal body.
//	pane.AddChild(spinner)
//	// No special wiring required — Update/Draw are called automatically.
//
// # Inspector integration
//
// The Inspector shows: active, determinate, progress, label, speed.

// SpinnerLabelPos controls where the optional text label is drawn.
type SpinnerLabelPos int

const (
	LabelBelow SpinnerLabelPos = iota // Label centred below the ring (default)
	LabelRight                        // Label to the right of the ring, vertically centred
)

// Spinner is a circular animated or progress ring.
//
// # LLM Prompt Template
//
//	sp := ui.NewSpinner("loading", 40, 0, 0)
//	sp.Active.Set(true)
//	modalBody.AddChild(sp)
//
// Demo scenes: **Batch 1**, **Widgets Demo**.
type Spinner struct {
	Element

	// Active controls whether the spinner animates. When false the ring is
	// still drawn at its current angle (frozen) — it does not disappear.
	// Set Hidden() to fully hide the widget.
	Active *Signal[bool]

	// Progress is only used in Determinate mode. Range 0.0–1.0.
	Progress *Signal[float32]

	// Determinate switches from the rotating arc to a static filled arc driven
	// by Progress. A percentage label is drawn at the ring centre.
	Determinate bool

	// Color is the arc fill colour. Defaults to indigo #4F46E5.
	Color rl.Color

	// TrackColor is the background ring colour. Defaults to light grey.
	TrackColor rl.Color

	// Speed is the rotation speed in degrees per second (indeterminate only).
	// Default: 360 (one full revolution per second).
	Speed float32

	// Label is optional text drawn beside or below the ring (see LabelPos).
	// Empty string = no label.
	Label string

	// LabelPos controls the label placement. Default: LabelBelow.
	LabelPos SpinnerLabelPos

	// Internal animation state.
	angle   float32 // accumulated rotation angle [0, 360)
	breathT float32 // time accumulator for arc-length breathing [0, 2π)
}

// NewSpinner creates a Spinner with the given outer diameter at position (x, y).
// Active defaults to false; call sp.Active.Set(true) to start animating.
func NewSpinner(id string, x, y, size float32) *Spinner {
	sp := &Spinner{
		Element:    NewElement(id, x, y, size, size),
		Active:     NewSignal(false),
		Progress:   NewSignal(float32(0)),
		Color:      rl.NewColor(79, 70, 229, 255),
		TrackColor: rl.NewColor(220, 222, 235, 255),
		Speed:      360,
	}
	sp.styleName = "spinner"
	sp.Active.Subscribe(func() { sp.MarkDrawDirty() })
	sp.Progress.Subscribe(func() { sp.MarkDrawDirty() })
	return sp
}

// SetStyle sets the theme key used for background / border lookup.
func (sp *Spinner) SetStyle(name string) { sp.styleName = name }

// IsInteractive reports that Spinner does not handle mouse/keyboard input.
func (sp *Spinner) IsInteractive() bool { return false }

// UsesScissor reports that Spinner does not open a scissor region.
func (sp *Spinner) UsesScissor() bool { return false }

// AnimationActive reports whether this spinner needs time-based redraws.
func (sp *Spinner) AnimationActive() bool {
	return !sp.IsHidden() && sp.Active.Get() && !sp.Determinate
}

// AnimationSource returns a compact source label for perf diagnostics.
func (sp *Spinner) AnimationSource() string { return sp.ID() }

// Layout grows bounds when LabelBelow needs space below the ring.
func (sp *Spinner) Layout() {
	defer func() { sp.layoutDirty = false }()
	if sp.IsHidden() {
		return
	}
	want := sp.preferredHeight()
	b := sp.Bounds()
	if want > 0 && b.Height < want-0.5 {
		b.Height = want
		sp.setBoundsNoMark(b)
	}
}

// Update advances the indeterminate rotation and breathing pulse each frame.
func (sp *Spinner) Update(dt float32) {
	if sp.IsHidden() || !sp.Active.Get() {
		return
	}
	if sp.Determinate {
		return
	}
	sp.angle += sp.Speed * dt
	if sp.angle >= 360 {
		sp.angle -= 360
	}
	// Breathing pulse: full cycle every ~1.4 s (2π / 4.5 rad/s).
	sp.breathT += dt * 4.5
	if sp.breathT >= 2*math.Pi {
		sp.breathT -= 2 * math.Pi
	}
}

// Draw renders the spinner ring, optional centre text, and optional label.
func (sp *Spinner) Draw() {
	sp.drawInternal()
}

// drawInternal performs the actual rendering.
func (sp *Spinner) drawInternal() {
	if sp.IsHidden() {
		return
	}

	geom, ok := sp.geometry()
	if !ok {
		return
	}

	// ── Track (full ring) ─────────────────────────────────────────────────────
	rl.DrawRing(rl.NewVector2(geom.cx, geom.cy), geom.inner, geom.outer, 0, 360, 36, sp.TrackColor)

	// ── Arc ───────────────────────────────────────────────────────────────────
	if sp.Determinate {
		// Static filled arc from top (−90°) proportional to Progress.
		p := sp.Progress.Get()
		if p < 0 {
			p = 0
		}
		if p > 1 {
			p = 1
		}
		sweep := p * 360
		if sweep > 0.5 {
			rl.DrawRing(rl.NewVector2(geom.cx, geom.cy), geom.inner, geom.outer,
				-90, -90+sweep, geom.steps, sp.Color)
		}

		// Percentage label at ring centre.
		pct := fmt.Sprintf("%d%%", int(p*100))
		s := spinnerLabelStyle()
		s.TextColor = rl.NewColor(40, 42, 60, 255)
		tw := measureTextS(pct, s)
		tx := int32(geom.cx) - tw/2
		ty := int32(geom.cy) - int32(EffectiveFontSize(s)/2)
		drawTextS(pct, tx, ty, s)
	} else if !sp.AnimationActive() {
		sp.drawIndeterminateArc(geom)
	}

	// ── Optional label ────────────────────────────────────────────────────────
	if sp.Label != "" {
		ls := spinnerLabelStyle()
		lw := measureTextS(sp.Label, ls)
		lh := EffectiveFontSize(ls)

		switch sp.LabelPos {
		case LabelBelow:
			lx := int32(geom.bounds.X) + int32(geom.bounds.Width/2) - lw/2
			ly := int32(geom.bounds.Y+geom.ringSize) + int32(labelGap)
			drawTextS(sp.Label, lx, ly, ls)
		case LabelRight:
			lx := int32(geom.bounds.X+geom.ringSize) + int32(labelGap)
			ly := int32(geom.cy) - int32(lh/2)
			drawTextS(sp.Label, lx, ly, ls)
		}
	}
}

// DrawAnimationOverlay draws a full native-layer spinner overlay. Drawing the
// track and arc together avoids visual drift between a 2x SSAA cached base and
// a 1x animated arc.
func (sp *Spinner) DrawAnimationOverlay() {
	if !sp.AnimationActive() {
		return
	}
	geom, ok := sp.geometry()
	if !ok {
		return
	}
	rl.DrawRing(rl.NewVector2(geom.cx, geom.cy), geom.inner, geom.outer, 0, 360, 36, sp.TrackColor)
	sp.drawIndeterminateArc(geom)
}

type spinnerGeometry struct {
	bounds   rl.Rectangle
	ringSize float32
	cx       float32
	cy       float32
	outer    float32
	inner    float32
	steps    int32
}

const labelGap = float32(6)

func (sp *Spinner) preferredHeight() float32 {
	b := sp.Bounds()
	d := b.Width
	if d < 4 {
		d = 64
	}
	if sp.Label != "" && sp.LabelPos == LabelBelow {
		return d + labelGap + EffectiveFontSize(spinnerLabelStyle())
	}
	if b.Height > d {
		return b.Height
	}
	return d
}

func spinnerLabelStyle() Style {
	s := GetThemeStyle("form-value")
	s.TextColor = rl.NewColor(80, 82, 100, 255)
	return s
}

func (sp *Spinner) geometry() (spinnerGeometry, bool) {
	b := sp.Bounds()
	labelH := EffectiveFontSize(spinnerLabelStyle())
	ringSize := b.Width
	if b.Height < ringSize {
		ringSize = b.Height
	}
	if sp.Label != "" && sp.LabelPos == LabelBelow {
		ringSize = b.Width
		if b.Height-labelH-labelGap < ringSize {
			ringSize = b.Height - labelH - labelGap
		}
	}
	if sp.Label != "" && sp.LabelPos == LabelRight && b.Height < ringSize {
		ringSize = b.Height
	}
	if ringSize < 4 {
		return spinnerGeometry{}, false
	}
	cx := b.X + ringSize/2
	cy := b.Y + ringSize/2
	outer := ringSize / 2
	return spinnerGeometry{
		bounds:   b,
		ringSize: ringSize,
		cx:       cx,
		cy:       cy,
		outer:    outer,
		inner:    outer * 0.68,
		steps:    int32(math.Max(36, math.Pi*float64(outer))),
	}, true
}

func (sp *Spinner) drawIndeterminateArc(geom spinnerGeometry) {
	breathSweep := float32(210) + 60*float32(0.5+0.5*math.Sin(float64(sp.breathT)))
	startDeg := sp.angle - 90
	endDeg := startDeg + breathSweep
	rl.DrawRing(rl.NewVector2(geom.cx, geom.cy), geom.inner, geom.outer, startDeg, endDeg, geom.steps, sp.Color)
}
