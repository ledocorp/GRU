// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"math"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─────────────────────────────────────────────────────────────────────────────
// Types
// ─────────────────────────────────────────────────────────────────────────────

// StepState describes the visual state of a single step in a Stepper.
type StepState int

const (
	// StepPending: step has not been reached yet — hollow ring, grey label.
	StepPending StepState = iota
	// StepActive: the step currently in focus — filled indigo circle, dark bold label.
	StepActive
	// StepCompleted: step has been completed — filled indigo circle with a ✓, dark label.
	StepCompleted
)

// StepItem holds the content for a single step in the sequence.
type StepItem struct {
	Title    string // Primary label shown next to or below the step circle (required).
	Subtitle string // Secondary description shown below Title; omit for single-line steps.
}

// StepperDirection controls the axis along which steps are arranged.
type StepperDirection int

const (
	// StepperHorizontal lays steps left-to-right with labels below each circle.
	StepperHorizontal StepperDirection = iota
	// StepperVertical lays steps top-to-bottom with labels to the right of each circle.
	StepperVertical
)

// ─────────────────────────────────────────────────────────────────────────────
// Stepper
// ─────────────────────────────────────────────────────────────────────────────

// Stepper is a horizontal or vertical progress indicator showing a labelled
// sequence of steps. It supports smooth animated transitions, optional
// click-to-jump navigation, and reactive control via CurrentStep.
//
// # LLM Prompt Template
//
//	st := ui.NewStepper("wizard", []ui.StepItem{
//	    {Title: "Account", Subtitle: "Create login"},
//	    {Title: "Confirm", Subtitle: "Review"},
//	}, 0, 0, 0, 80)
//	st.Clickable = true
//	panel.AddChild(st)
//
// Demo scenes: **Batch 4 Stepper**.
//
// Basic usage:
//
//	steps := []ui.StepItem{
//	    {Title: "Account",  Subtitle: "Create your login"},
//	    {Title: "Profile",  Subtitle: "Fill in your details"},
//	    {Title: "Confirm",  Subtitle: "Review and submit"},
//	}
//	st := ui.NewStepper("wizard", steps, 0, 0, 0, 80)
//	st.Clickable = true
//	parent.AddChild(st)
//
//	// Advance programmatically:
//	nextBtn.OnClick = func() { st.CurrentStep.Set(st.CurrentStep.Get() + 1) }
//
// For horizontal steppers the recommended height is roughly 80 px (circle + label +
// subtitle); for vertical steppers allow ~60 px per step.
type Stepper struct {
	Element

	// Steps is the ordered list of step definitions. Provide at construction time.
	Steps []StepItem

	// CurrentStep is the zero-based index of the currently active step.
	// Assign via CurrentStep.Set(i) to trigger the slide animation.
	// Valid range: [0, len(Steps)-1]; out-of-range values are clamped silently.
	CurrentStep *Signal[int]

	// Direction controls horizontal (default) or vertical layout.
	Direction StepperDirection

	// Clickable, when true, allows the user to jump to any step by clicking
	// its circle. OnStepClick is called on each successful click.
	Clickable bool

	// OnStepClick fires with the clicked step index when Clickable is true.
	// May be nil.
	OnStepClick func(index int)

	// CircleRadius is the radius of each step circle in pixels. Default: 18.
	CircleRadius float32

	// LineThick is the thickness of the connecting line in pixels. Default: 2.
	LineThick float32

	// animProgress is the animated step position in [0, N-1] space.
	// The connecting line between step i and i+1 is filled by the fraction
	//   clamp(animProgress - i, 0, 1)
	// which correctly handles both forward and backward transitions.
	animProgress float32

	// tween drives animProgress when CurrentStep changes.
	tween *Tween

	// hoverStep is the step index under the mouse cursor, or -1 when none.
	hoverStep int
}

// ─────────────────────────────────────────────────────────────────────────────
// Constructor
// ─────────────────────────────────────────────────────────────────────────────

// NewStepper creates a Stepper at the given bounds. For horizontal steppers
// a height of 80 is usually sufficient; for vertical steppers allow 60 px
// per step (e.g. h = len(steps) * 60).
//
//	id      — unique node ID
//	steps   — ordered slice of StepItem (may be replaced at runtime by a new
//	          NewStepper call if steps change; signal re-subscribe is required)
//	x, y    — position (overridden by parent layout)
//	w, h    — explicit dimensions; set either to 0 for parent cross-axis stretch
func NewStepper(id string, steps []StepItem, x, y, w, h float32) *Stepper {
	st := &Stepper{
		Element:      NewElement(id, x, y, w, h),
		Steps:        steps,
		CurrentStep:  NewSignal(0),
		Direction:    StepperHorizontal,
		Clickable:    false,
		CircleRadius: 18,
		LineThick:    2.5,
		animProgress: 0,
		hoverStep:    -1,
	}
	st.styleName = "stepper"

	// Subscribe: any CurrentStep change triggers an animated transition from
	// the current visual position (animProgress) to the new target.
	st.CurrentStep.Subscribe(func() {
		n := len(st.Steps)
		if n == 0 {
			return
		}
		next := st.CurrentStep.Get()
		// Silent clamp — avoids re-entrant Set calls.
		if next < 0 {
			next = 0
		}
		if next >= n {
			next = n - 1
		}
		st.startTransition(float32(next))
	})

	return st
}

// ─────────────────────────────────────────────────────────────────────────────
// Node interface
// ─────────────────────────────────────────────────────────────────────────────

// SetStyle sets the theme key used for the label styles.
func (st *Stepper) SetStyle(name string) { st.styleName = name }

// IsInteractive returns true when Clickable is enabled so the engine routes
// mouse events to this widget.
func (st *Stepper) IsInteractive() bool { return st.Clickable }

// UsesScissor reports false — Stepper clips nothing.
func (st *Stepper) UsesScissor() bool { return false }

// Layout is a no-op; Stepper is a leaf widget with caller-defined bounds.
func (st *Stepper) Layout() { st.layoutDirty = false }

// ─────────────────────────────────────────────────────────────────────────────
// Update
// ─────────────────────────────────────────────────────────────────────────────

// Update advances the transition animation and processes hover / click input.
func (st *Stepper) Update(dt float32) {
	if st.IsHidden() {
		return
	}

	// Advance tween.
	if st.tween != nil {
		if st.tween.IsActive {
			st.tween.Update(dt)
		} else {
			st.tween = nil
		}
	}

	// Hover and click (only when Clickable).
	if !st.Clickable {
		if st.hoverStep != -1 {
			st.hoverStep = -1
			st.MarkDrawDirty()
		}
		return
	}

	mouse := rl.GetMousePosition()
	prevHover := st.hoverStep
	st.hoverStep = -1
	r := st.CircleRadius
	for i, c := range st.circleCenters() {
		dx := mouse.X - c.X
		dy := mouse.Y - c.Y
		if dx*dx+dy*dy <= r*r {
			st.hoverStep = i
			break
		}
	}
	if st.hoverStep != prevHover {
		st.MarkDrawDirty()
	}

	if st.hoverStep >= 0 && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		idx := st.hoverStep
		if idx != st.CurrentStep.Get() {
			st.CurrentStep.Set(idx)
		}
		if st.OnStepClick != nil {
			st.OnStepClick(idx)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Draw
// ─────────────────────────────────────────────────────────────────────────────

// Draw renders the stepper into the current draw pass.
func (st *Stepper) Draw() {
	if st.IsHidden() {
		return
	}
	st.drawInternal()
	st.drawDirty = false
}

func (st *Stepper) drawInternal() {
	n := len(st.Steps)
	if n == 0 {
		return
	}

	centers := st.circleCenters()
	r := st.CircleRadius
	cur := st.CurrentStep.Get()
	// Clamp cur to valid range for display purposes (signal value may be
	// in flight if the widget is rebuilt).
	if cur < 0 {
		cur = 0
	}
	if cur >= n {
		cur = n - 1
	}

	// ── Connecting lines ──────────────────────────────────────────────────────
	for i := 0; i < n-1; i++ {
		c1 := centers[i]
		c2 := centers[i+1]

		// Fill fraction for this segment: how far animProgress has moved past step i.
		fillFrac := st.animProgress - float32(i)
		if fillFrac < 0 {
			fillFrac = 0
		}
		if fillFrac > 1 {
			fillFrac = 1
		}

		if st.Direction == StepperHorizontal {
			lineX := c1.X + r + 3
			lineY := c1.Y - st.LineThick/2
			lineW := c2.X - r - 3 - lineX
			if lineW < 1 {
				lineW = 1
			}
			// Background (empty) segment.
			rl.DrawRectangleRec(rl.NewRectangle(lineX, lineY, lineW, st.LineThick), stepLineEmpty)
			// Filled (completed) portion.
			if fillFrac > 0 {
				rl.DrawRectangleRec(rl.NewRectangle(lineX, lineY, lineW*fillFrac, st.LineThick), stepLineFilled)
			}
		} else {
			lineX := c1.X - st.LineThick/2
			lineY := c1.Y + r + 3
			lineH := c2.Y - r - 3 - lineY
			if lineH < 1 {
				lineH = 1
			}
			rl.DrawRectangleRec(rl.NewRectangle(lineX, lineY, st.LineThick, lineH), stepLineEmpty)
			if fillFrac > 0 {
				rl.DrawRectangleRec(rl.NewRectangle(lineX, lineY, st.LineThick, lineH*fillFrac), stepLineFilled)
			}
		}
	}

	// ── Step circles + labels ─────────────────────────────────────────────────
	for i, step := range st.Steps {
		c := centers[i]

		var state StepState
		if i < cur {
			state = StepCompleted
		} else if i == cur {
			state = StepActive
		} else {
			state = StepPending
		}

		hovered := st.hoverStep == i && st.Clickable
		st.drawStepCircle(c, r, state, i+1, hovered)
		st.drawStepLabel(c, r, step, state)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Drawing helpers
// ─────────────────────────────────────────────────────────────────────────────

// drawStepCircle renders the circle indicator for one step.
func (st *Stepper) drawStepCircle(center rl.Vector2, r float32, state StepState, num int, hovered bool) {
	cx, cy := center.X, center.Y
	segs := int32(math.Max(24, math.Pi*float64(r)))

	// Keep numbers visually centered inside fixed-size circles. They use the SDF
	// path, but opt out of the larger body-text floor that would crowd the icon.
	numFontSize := int32(r * 0.78)
	if numFontSize < 9 {
		numFontSize = 9
	}
	ns := GetThemeStyle("default")
	ns.FontSize = numFontSize
	ns.MinFontSize = 10
	ns.Bold = true
	ns.Padding = 0
	ns.BorderWidth = 0

	switch state {
	case StepPending:
		// White fill + grey border ring.
		rl.DrawCircleV(center, r, stepCirclePendingFill)
		if hovered {
			// Faint indigo tint on hover.
			rl.DrawCircleV(center, r, rl.NewColor(79, 70, 229, 18))
		}
		rl.DrawRing(center, r-2, r, 0, 360, segs, stepCirclePendingBorder)
		// Grey number.
		ns.TextColor = stepCirclePendingBorder
		nStr := fmt.Sprintf("%d", num)
		nW := measureTextS(nStr, ns)
		nH := int32(EffectiveFontSize(ns))
		drawTextS(nStr, int32(cx)-nW/2, int32(cy)-nH/2, ns)

	case StepActive:
		// Subtle outer glow ring.
		glowSegs := int32(math.Max(24, math.Pi*float64(r+5)))
		rl.DrawCircleV(center, r+5, rl.NewColor(79, 70, 229, 28))
		_ = glowSegs
		// Filled indigo circle.
		rl.DrawCircleV(center, r, stepCircleActive)
		if hovered {
			// Brighter on hover.
			rl.DrawCircleV(center, r, rl.NewColor(255, 255, 255, 20))
		}
		// White number.
		ns.TextColor = rl.White
		nStr := fmt.Sprintf("%d", num)
		nW := measureTextS(nStr, ns)
		nH := int32(EffectiveFontSize(ns))
		drawTextS(nStr, int32(cx)-nW/2, int32(cy)-nH/2, ns)

	case StepCompleted:
		// Filled indigo circle.
		rl.DrawCircleV(center, r, stepCircleDone)
		if hovered {
			rl.DrawCircleV(center, r, rl.NewColor(255, 255, 255, 20))
		}
		icon := r * 1.15
		inner := rl.NewRectangle(cx-icon/2, cy-icon/2, icon, icon)
		Phosphor.EnsureLoaded(PhosphorCheck, PhosphorRegular)
		Phosphor.Draw(inner, PhosphorCheck, PhosphorRegular, rl.White)
	}
}

// drawStepLabel renders the title (and optional subtitle) for one step.
func (st *Stepper) drawStepLabel(center rl.Vector2, r float32, step StepItem, state StepState) {
	titleStyle := GetThemeStyle("stepper-title")
	subStyle := GetThemeStyle("stepper-subtitle")

	if state == StepPending {
		titleStyle.TextColor = stepTitlePending
		subStyle.TextColor = stepTitlePending
	}
	// Active step uses bold title.
	titleStyle.Bold = state == StepActive

	titleW := measureTextS(step.Title, titleStyle)

	if st.Direction == StepperHorizontal {
		// Labels centred below the circle.
		titleH := int32(EffectiveFontSize(titleStyle))
		labelY := int32(center.Y+r) + 9
		drawTextS(step.Title, int32(center.X)-titleW/2, labelY, titleStyle)
		if step.Subtitle != "" {
			subW := measureTextS(step.Subtitle, subStyle)
			drawTextS(step.Subtitle, int32(center.X)-subW/2, labelY+titleH+3, subStyle)
		}
	} else {
		// Labels to the right of the circle.
		labelX := int32(center.X+r) + 12
		var labelY int32
		titleH := int32(EffectiveFontSize(titleStyle))
		if step.Subtitle != "" {
			// Vertically centre both lines around the circle midpoint.
			subH := int32(EffectiveFontSize(subStyle))
			totalH := titleH + subH + 3
			labelY = int32(center.Y) - totalH/2
		} else {
			labelY = int32(center.Y) - titleH/2
		}
		drawTextS(step.Title, labelX, labelY, titleStyle)
		if step.Subtitle != "" {
			drawTextS(step.Subtitle, labelX, labelY+titleH+3, subStyle)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Geometry helpers
// ─────────────────────────────────────────────────────────────────────────────

// circleCenters computes the centre of each step circle from the current bounds.
func (st *Stepper) circleCenters() []rl.Vector2 {
	n := len(st.Steps)
	if n == 0 {
		return nil
	}
	b := st.bounds
	r := st.CircleRadius
	centers := make([]rl.Vector2, n)

	if st.Direction == StepperHorizontal {
		pitch := b.Width / float32(n)
		cy := b.Y + r + 6
		for i := range st.Steps {
			cx := b.X + pitch*float32(i) + pitch*0.5
			centers[i] = rl.NewVector2(cx, cy)
		}
	} else {
		pitch := b.Height / float32(n)
		cx := b.X + r + 10
		for i := range st.Steps {
			cy := b.Y + pitch*float32(i) + pitch*0.5
			centers[i] = rl.NewVector2(cx, cy)
		}
	}
	return centers
}

// startTransition begins an animated move of animProgress toward target.
// The duration scales with distance (0.25 s/step) clamped to [0.12, 0.60] s.
func (st *Stepper) startTransition(target float32) {
	start := st.animProgress
	dist := target - start
	if dist < 0 {
		dist = -dist
	}
	dur := dist * 0.25
	if dur < 0.12 {
		dur = 0.12
	}
	if dur > 0.60 {
		dur = 0.60
	}

	st.tween = NewTween(start, target, dur, EaseInOutQuad,
		func(v float32) {
			st.animProgress = v
			st.MarkDrawDirty()
		},
		func() {
			st.animProgress = target
			st.tween = nil
			st.MarkDrawDirty()
		},
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Palette (package-level; unexported)
// ─────────────────────────────────────────────────────────────────────────────

var (
	stepCircleActive        = rl.NewColor(79, 70, 229, 255)   // indigo-600
	stepCircleDone          = rl.NewColor(79, 70, 229, 255)   // indigo-600
	stepCirclePendingFill   = rl.NewColor(255, 255, 255, 255) // white
	stepCirclePendingBorder = rl.NewColor(203, 207, 220, 255) // cool grey

	stepLineFilled = rl.NewColor(79, 70, 229, 255)   // indigo-600
	stepLineEmpty  = rl.NewColor(218, 220, 228, 255) // light grey

	stepTitlePending = rl.NewColor(150, 155, 170, 255) // grey
)
