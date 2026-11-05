// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── RadioGroup ──────────────────────────────────────────────────────────────
//
// RadioGroup presents a set of mutually exclusive options. Exactly one option
// can be selected at a time; clicking a selected option keeps it selected.
// The selected index is exposed as a reactive Signal[int] so subscribers are
// notified whenever the selection changes.
//
// # Basic usage
//
//	rg := ui.NewRadioGroup("theme",
//	    []string{"System Default", "Light Mode", "Dark Mode"},
//	    20, 20, 220, 108)
//	rg.Selected.Set(0)
//	rg.Selected.Subscribe(func() {
//	    fmt.Println("selected:", rg.Selected.Get())
//	})
//
// # Layout
//
// Set Vertical = true (default) for a top-to-bottom column of options, or
// false for a left-to-right row. RowH controls the per-option height (or
// width in horizontal mode). The indicator circle is always drawn first
// (left / top), the label beside it.
//
// # Disabled options
//
// Mark individual options non-interactive by setting elements of the Disabled
// slice before or after construction:
//
//	rg.Disabled[2] = true   // "Dark Mode" is greyed out and unclickable
//
// # Visual design
//
// Each option row draws:
//   - A hover highlight rectangle (rounded, low-opacity indigo).
//   - An outer ring (border): grey when unselected, indigo when selected.
//   - An inner disc: hollow when unselected, indigo fill + white centre dot
//     when selected.
//   - The label text to the right of the circle.
//
// Both the outer ring and disc use G1.1 coverage AA at SSAA 1× (DrawCoverageRing /
// DrawCoverageCircle); with scene SSAA they fall back to tessellated primitives.
//
// # Style
//
// RadioGroup resolves through the Theme v2 "radio/default" component when it
// is available. The checked, hover, and disabled states drive the ring, fill,
// and row-highlight colors; legacy theme keys remain a fallback path through
// Element.ResolveStyle.
//
// # Inspector integration
//
// The Inspector pane shows the option list, current selection index and label,
// disabled mask, and layout direction.

const (
	rgCircleR   = float32(8)  // outer radius of the radio indicator
	rgCircleGap = float32(8)  // gap between circle right edge and label
	rgRowPadX   = float32(10) // left/right padding inside each option row
)

// RadioGroup is a set of mutually exclusive toggle options.
//
// # LLM Prompt Template
//
//	rg := ui.NewRadioGroup("theme", []string{"Light", "Dark"}, 0, 0, 220, 108)
//	rg.Selected.Subscribe(func() { applyTheme(rg.Selected.Get()) })
//	form.AddChild(rg)
//
// Demo scenes: **Batch 1**, **Batch 20 RadioGroup**, **Widgets Demo**.
type RadioGroup struct {
	Element

	// Options is the list of label strings shown for each radio button.
	Options []string

	// Selected is the index of the active option. -1 means nothing is selected.
	// Setting it via Set() notifies all subscribers and triggers a redraw.
	Selected *Signal[int]

	// Disabled is a per-option disable mask. A true entry makes that option
	// unclickable and renders it greyed out. The slice is automatically grown
	// to len(Options) when needed.
	Disabled []bool

	// Vertical controls the layout direction.
	// true  = options stacked top-to-bottom (default)
	// false = options arranged left-to-right
	Vertical bool

	// RowH is the per-option row height (vertical mode) or width (horizontal).
	// Default: 32. Should be at least 2×CircleRadius + 4 to avoid clipping.
	RowH float32

	// CircleRadius is the outer radius of the indicator circle in pixels.
	// Increase for larger, touch-friendly targets; decrease for compact lists.
	// Default: 8.
	CircleRadius float32

	hovered int // index of the currently hovered option, or -1
}

// NewRadioGroup creates a RadioGroup with the given options at position (x, y).
// Selected defaults to -1 (nothing selected); call rg.Selected.Set(0) to
// pre-select the first option. Disabled defaults to all false.
func NewRadioGroup(id string, options []string, x, y, w, h float32) *RadioGroup {
	rg := &RadioGroup{
		Element:      NewElement(id, x, y, w, h),
		Options:      options,
		Selected:     NewSignal(-1),
		Disabled:     make([]bool, len(options)),
		Vertical:     true,
		RowH:         32,
		CircleRadius: 8,
		hovered:      -1,
	}
	rg.styleName = "radio"
	rg.Element.SetStyleVariant("radio", "default")
	rg.Selected.Subscribe(func() { rg.MarkDrawDirty() })
	return rg
}

// SetStyle sets the base theme key (default: "radio").
func (rg *RadioGroup) SetStyle(name string) { rg.styleName = name }

// IsInteractive reports that RadioGroup accepts mouse input.
func (rg *RadioGroup) IsInteractive() bool { return true }

// UsesScissor reports that RadioGroup does not open a scissor region.
func (rg *RadioGroup) UsesScissor() bool { return false }

// Layout is a no-op; the widget uses its assigned bounds directly.
func (rg *RadioGroup) Layout() { rg.layoutDirty = false }

// Update handles hover detection and click selection.
func (rg *RadioGroup) Update(_ float32) {
	if rg.IsHidden() {
		return
	}
	// Grow Disabled slice if Options were added after construction.
	for len(rg.Disabled) < len(rg.Options) {
		rg.Disabled = append(rg.Disabled, false)
	}

	b := rg.Bounds()
	mouse := rl.GetMousePosition()
	prevHovered := rg.hovered
	rg.hovered = -1

	for i := range rg.Options {
		row := rg.rowRect(b, i)
		if rl.CheckCollisionPointRec(mouse, row) {
			if !rg.isDisabled(i) {
				rg.hovered = i
				if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
					rg.Selected.Set(i)
				}
			}
		}
	}
	if rg.hovered != prevHovered {
		rg.MarkDrawDirty()
	}
}

// Draw renders all radio options.
func (rg *RadioGroup) Draw() {
	rg.drawInternal()
}

// drawInternal performs the actual rendering.
func (rg *RadioGroup) drawInternal() {
	if rg.IsHidden() {
		return
	}
	// Grow Disabled slice defensively (Draw may be called before Update).
	for len(rg.Disabled) < len(rg.Options) {
		rg.Disabled = append(rg.Disabled, false)
	}

	b := rg.Bounds()
	selectedIdx := rg.Selected.Get()

	styleNorm, _ := rg.ResolveStyle(StyleStateNone)
	styleSel, _ := rg.ResolveStyle(StyleStateChecked)
	styleHov, _ := rg.ResolveStyle(StyleStateHover)
	styleDis, _ := rg.ResolveStyle(StyleStateDisabled)

	for i, label := range rg.Options {
		row := rg.rowRect(b, i)
		isSelected := i == selectedIdx
		isHovered := i == rg.hovered
		isDisabled := rg.isDisabled(i)

		// ── Hover highlight ───────────────────────────────────────────────────
		if isHovered && !isDisabled {
			hRect := rl.NewRectangle(row.X+2, row.Y+2, row.Width-4, row.Height-4)
			DrawRoundedFill(hRect, 0.4, styleHov.BackgroundColor)
		}

		// ── Circle indicator ──────────────────────────────────────────────────
		cx, cy := rg.circleCenter(row)
		cr := rg.CircleRadius
		innerR := cr - 2 // border width = 2px

		if isDisabled {
			DrawCoverageRing(cx, cy, innerR, cr, 0, 360, styleDis.BorderColor)
			DrawCoverageCircle(cx, cy, innerR, styleDis.BackgroundColor)
		} else if isSelected {
			// Filled indigo disc.
			DrawCoverageCircle(cx, cy, cr, styleSel.BackgroundColor)
			// Outer border ring (thin, slightly darker).
			DrawCoverageRing(cx, cy, cr-1.5, cr, 0, 360, styleSel.BorderColor)
			// Inner white dot (≈40% of radius).
			DrawCoverageCircle(cx, cy, cr*0.38, rl.NewColor(255, 255, 255, 255))
		} else {
			// Hollow: fill full disc then one outer ring (innerR fill + ring = double AA).
			DrawCoverageCircle(cx, cy, cr, styleNorm.BackgroundColor)
			DrawCoverageRing(cx, cy, cr-1.15, cr, 0, 360, styleNorm.BorderColor)
		}

		// ── Label text ────────────────────────────────────────────────────────
		var ts Style
		switch {
		case isDisabled:
			ts = styleDis
		case isSelected:
			ts = styleSel
			ts.TextColor = styleNorm.TextColor // keep label in dark text
		default:
			ts = styleNorm
		}

		textX := int32(cx) + int32(cr) + int32(rgCircleGap)
		textY := int32(cy) - ts.FontSize/2
		drawTextS(label, textX, textY, ts)
	}
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

// isDisabled returns true if option i is disabled (bounds-checked).
func (rg *RadioGroup) isDisabled(i int) bool {
	if i < 0 || i >= len(rg.Disabled) {
		return false
	}
	return rg.Disabled[i]
}

// rowRect returns the bounding rectangle for option i.
func (rg *RadioGroup) rowRect(b rl.Rectangle, i int) rl.Rectangle {
	if rg.Vertical {
		return rl.NewRectangle(b.X, b.Y+float32(i)*rg.RowH, b.Width, rg.RowH)
	}
	colW := b.Width / float32(len(rg.Options))
	return rl.NewRectangle(b.X+float32(i)*colW, b.Y, colW, b.Height)
}

// circleCenter returns the screen-space centre of the indicator circle for
// the given row rectangle. In vertical mode the circle is left-padded; in
// horizontal mode it is top-centred so the label can sit below if desired
// (currently label is to the right in both modes).
func (rg *RadioGroup) circleCenter(row rl.Rectangle) (cx, cy float32) {
	cx = row.X + rgRowPadX + rg.CircleRadius
	cy = row.Y + row.Height/2
	return
}

// InteractionOverlayActive implements InteractionOverlayPainter.
// Hover redraws in the SSAA cache so option labels stay crisp.
func (rg *RadioGroup) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (rg *RadioGroup) DrawInteractionOverlay() {}
