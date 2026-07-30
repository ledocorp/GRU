// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Slider is an interactive widget for selecting a float32 value in a range.
//
// Value is a Signal[float32] in [MinValue, MaxValue].  Dragging the thumb
// updates Value continuously.  Effects subscribed to Value.Get() re-run on
// every frame the mouse is held.  The thumb snaps to the click position on
// mouse-down and continues tracking until mouse-up.
//
// # LLM Prompt Template
//
//	vol := ui.NewSignal(float32(5))
//	sl := ui.NewSlider("volume", 0, 10, vol.Get(), 0, 0, 260, 36)
//	sl.Value = vol
//	field.AddChild(sl)
//
// DocumentSpec: `{ "type": "slider", "label": "Volume", "min": 0, "max": 10, "value": 5 }`.
// Demo scenes: **Form Demo**, **Widgets Demo**, `pages/gallery.gru`.
//
// When ShowValue is true (the default) the current value is drawn to the
// right of the track in a fixed-width label column (60 px).  Set ValueFmt to
// a fmt format string (default "%.0f") to control how the number is printed.
//
// Slider resolves through the Theme v2 "slider/default" component and applies
// hover and pressed states to the track and thumb when available.
type Slider struct {
	Element
	Value     *Signal[float32] // Reactive value
	MinValue  float32          // Range minimum
	MaxValue  float32          // Range maximum
	ShowValue bool             // Draw current value to the right of the track
	ValueFmt  string           // fmt format string for the value label, e.g. "%.1f"
	dragging  bool
	hovered   bool
}

const sliderLabelW = float32(60) // reserved width for the value label

// NewSlider creates a Slider with the given range and initial value.
// ShowValue defaults to true so the current value is always visible.
func NewSlider(id string, minVal, maxVal, initialVal float32, x, y, w, h float32) *Slider {
	if initialVal < minVal {
		initialVal = minVal
	}
	if initialVal > maxVal {
		initialVal = maxVal
	}
	s := &Slider{
		Element:   NewElement(id, x, y, w, h),
		Value:     NewSignal(initialVal),
		MinValue:  minVal,
		MaxValue:  maxVal,
		ShowValue: true,
		ValueFmt:  "%.0f",
	}
	s.styleName = "slider"
	s.Element.SetStyleVariant("slider", "default")
	s.Value.Subscribe(func() { s.MarkDrawDirty() })
	return s
}

// trackBounds returns the rectangle used for the interactive track area.
// When ShowValue is true, sliderLabelW pixels on the right are reserved for
// the value text so the track does not extend behind the label.
func (s *Slider) trackBounds() rl.Rectangle {
	b := s.Bounds()
	if s.ShowValue {
		w := b.Width - sliderLabelW
		if w < 1 {
			w = 1 // avoid zero/negative track when flex squeezes the widget too narrow
		}
		return rl.NewRectangle(b.X, b.Y, w, b.Height)
	}
	return b
}

// Update implements Node.Update by handling mouse input.
func (s *Slider) Update(_ float32) {
	if s.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	tb := s.trackBounds()
	prevHovered := s.hovered
	prevDragging := s.dragging
	s.hovered = rl.CheckCollisionPointRec(mouse, tb)

	if s.hovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		s.dragging = true
	}
	if s.dragging {
		if rl.IsMouseButtonDown(rl.MouseLeftButton) {
			tw := tb.Width
			if tw < 1 {
				tw = 1
			}
			ratio := (mouse.X - tb.X) / tw
			if ratio < 0 {
				ratio = 0
			}
			if ratio > 1 {
				ratio = 1
			}
			s.Value.Set(s.MinValue + ratio*(s.MaxValue-s.MinValue))
			s.MarkDrawDirty()
		} else {
			s.dragging = false
			s.MarkDrawDirty()
		}
	}
	if s.hovered != prevHovered || s.dragging != prevDragging {
		s.MarkDrawDirty()
	}
}

// IsDragging reports whether the user is dragging the thumb (for coalescing layout).
func (s *Slider) IsDragging() bool { return s.dragging }

// InteractionOverlayActive implements InteractionOverlayPainter.
// Track/thumb redraw in the SSAA cache so the thumb does not ghost while dragging.
func (s *Slider) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (s *Slider) DrawInteractionOverlay() {}

// Layout implements Node.Layout (no-op for leaf widgets).
func (s *Slider) Layout() { s.layoutDirty = false }

// Draw implements Node.Draw.
func (s *Slider) Draw() { s.drawInternal() }

func (s *Slider) drawInternal() {
	if s.IsHidden() {
		return
	}

	state := StyleStateNone
	if s.hovered {
		state |= StyleStateHover
	}
	if s.dragging {
		state |= StyleStatePressed
	}
	style, stateApplied := s.ResolveStyle(state)
	tb := s.trackBounds()

	// ── Track background ──────────────────────────────────────────────────────
	trackH := float32(6)
	trackY := tb.Y + (tb.Height-trackH)/2
	trackRect := rl.NewRectangle(tb.X, trackY, tb.Width, trackH)
	rl.DrawRectangleRounded(trackRect, 1.0, 6, rl.NewColor(220, 222, 230, 255))

	// ── Filled section (indigo) up to the thumb ───────────────────────────────
	value := s.Value.Get()
	ratio := (value - s.MinValue) / (s.MaxValue - s.MinValue)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	// Determine ancestor Viewport clip for all sub-scissors in this widget.
	var vpClip rl.Rectangle
	hasVP := false
	if vp := findViewport(s); vp != nil {
		vpClip = vp.ClipBounds()
		hasVP = true
	}
	restoreVP := func() {
		if hasVP {
			beginScissorMode(int32(vpClip.X), int32(vpClip.Y), int32(vpClip.Width), int32(vpClip.Height))
		}
	}

	fillW := ratio * tb.Width
	if fillW > 0 {
		fillColor := rl.NewColor(79, 70, 229, 255)
		if s.hovered || s.dragging {
			fillColor = rl.NewColor(99, 90, 249, 255)
		}
		// Scissor to fillW intersected with the Viewport clip so the fill pill
		// never bleeds outside the scroll container on partial scroll-off.
		fillRect := rl.NewRectangle(tb.X, trackY-1, fillW, trackH+2)
		if hasVP {
			fillRect = intersectRects(fillRect, vpClip)
		}
		if fillRect.Width > 0 && fillRect.Height > 0 {
			beginScissorMode(int32(fillRect.X), int32(fillRect.Y), int32(fillRect.Width), int32(fillRect.Height))
			rl.DrawRectangleRounded(trackRect, 1.0, 6, fillColor)
			rl.EndScissorMode()
			restoreVP()
		}
	}

	// ── Thumb ─────────────────────────────────────────────────────────────────
	thumbX := tb.X + ratio*tb.Width
	thumbSize := float32(18)
	if s.dragging {
		thumbSize = 20
	}
	thumbRect := rl.NewRectangle(
		thumbX-thumbSize/2,
		tb.Y+(tb.Height-thumbSize)/2,
		thumbSize, thumbSize,
	)

	// Soft glow ring behind the thumb on hover / drag
	if s.hovered || s.dragging {
		glowSize := thumbSize + 10
		glowRect := rl.NewRectangle(
			thumbX-glowSize/2,
			tb.Y+(tb.Height-glowSize)/2,
			glowSize, glowSize,
		)
		rl.DrawRectangleRounded(glowRect, 1.0, 8, rl.NewColor(79, 70, 229, 24))
	}

	thumbBorder := focusRingIndigo
	if stateApplied && style.BorderColor.A != 0 {
		thumbBorder = style.BorderColor
	}
	rl.DrawRectangleRounded(thumbRect, 1.0, 8, rl.NewColor(255, 255, 255, 255))
	rl.DrawRectangleRoundedLinesEx(thumbRect, 1.0, 8, 2, thumbBorder)

	// ── Value label ───────────────────────────────────────────────────────────
	if s.ShowValue {
		b := s.Bounds()
		txt := fmt.Sprintf(s.ValueFmt, value)
		tw := measureTextS(txt, style)
		// Centre the text inside the label column
		labelX := int32(b.X+b.Width-sliderLabelW) + (int32(sliderLabelW)-tw)/2
		labelY := int32(b.Y) + (int32(b.Height)-int32(EffectiveFontSize(style)))/2
		drawTextS(txt, labelX, labelY, style)
	}
}

// IsInteractive implements Node.IsInteractive.
func (s *Slider) IsInteractive() bool { return true }

// UsesScissor implements Node.UsesScissor.
// Slider clips the fill bar to the filled width via BeginScissorMode.
func (s *Slider) UsesScissor() bool { return true }
