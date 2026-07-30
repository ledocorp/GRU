// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"math"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	spinBtnW     float32 = 34
	spinValuePad float32 = 8
)

// SpinBox is a numeric stepper with − / + buttons and keyboard nudging when focused.
//
// Value is clamped to [Min, Max] on every change. Set DecimalPlaces to control
// the value label (0 = integers).
//
// # LLM Prompt Template
//
//	vol := ui.NewSignal(50.0)
//	sb := ui.NewSpinBox("volume", vol, 0, 100, 5, 0, 0, 160, 36)
//	form.AddChild(sb)
//
// Demo scenes: **Batch 16 SpinBox**.
type SpinBox struct {
	Element
	Value          *Signal[float64]
	Min            float64
	Max            float64
	Step           float64
	DecimalPlaces  int
	Disabled       bool
	focused        bool
	hovered        bool
	hoverDec       bool
	hoverInc       bool
}

// NewSpinBox creates a SpinBox bound to value. Bounds width should be ≥ 120 px.
func NewSpinBox(id string, value *Signal[float64], min, max, step float64, x, y, w, h float32) *SpinBox {
	if value == nil {
		value = NewSignal(min)
	}
	if step <= 0 {
		step = 1
	}
	sb := &SpinBox{
		Element:       NewElement(id, x, y, w, h),
		Value:         value,
		Min:           min,
		Max:           max,
		Step:          step,
		DecimalPlaces: 0,
	}
	sb.styleName = "spinbox"
	sb.Element.SetStyleVariant("spinbox", "default")
	sb.clampAndSet(sb.Value.Get())
	sb.Value.Subscribe(func() { sb.MarkDirty() })
	sb.On(EventFocus, func(Event) { sb.focused = !sb.Disabled })
	sb.On(EventBlur, func(Event) { sb.focused = false })
	return sb
}

// IsInteractive implements Node.
func (sb *SpinBox) IsInteractive() bool { return true }

func (sb *SpinBox) decBounds() rl.Rectangle {
	b := sb.Bounds()
	return rl.NewRectangle(b.X, b.Y, spinBtnW, b.Height)
}

func (sb *SpinBox) incBounds() rl.Rectangle {
	b := sb.Bounds()
	return rl.NewRectangle(b.X+b.Width-spinBtnW, b.Y, spinBtnW, b.Height)
}

func (sb *SpinBox) valueBounds() rl.Rectangle {
	b := sb.Bounds()
	return rl.NewRectangle(b.X+spinBtnW, b.Y, b.Width-2*spinBtnW, b.Height)
}

func (sb *SpinBox) clampAndSet(v float64) {
	if v < sb.Min {
		v = sb.Min
	}
	if v > sb.Max {
		v = sb.Max
	}
	if sb.Value.Get() != v {
		sb.Value.Set(v)
	}
}

func (sb *SpinBox) nudge(mult float64) {
	sb.clampAndSet(sb.Value.Get() + sb.Step*mult)
}

// Update handles buttons, focus click, and keyboard.
func (sb *SpinBox) Update(_ float32) {
	if sb.IsHidden() || sb.Disabled {
		if sb.hovered || sb.focused {
			sb.hovered = false
			sb.hoverDec = false
			sb.hoverInc = false
			sb.MarkDrawDirty()
		}
		return
	}

	mouse := rl.GetMousePosition()
	dec := sb.decBounds()
	inc := sb.incBounds()
	val := sb.valueBounds()

	prevH := sb.hovered
	prevD := sb.hoverDec
	prevI := sb.hoverInc
	sb.hoverDec = rl.CheckCollisionPointRec(mouse, dec)
	sb.hoverInc = rl.CheckCollisionPointRec(mouse, inc)
	sb.hovered = sb.hoverDec || sb.hoverInc || rl.CheckCollisionPointRec(mouse, val)
	if sb.hovered != prevH || sb.hoverDec != prevD || sb.hoverInc != prevI {
		sb.MarkDrawDirty()
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if sb.hoverDec {
			sb.nudge(-1)
		} else if sb.hoverInc {
			sb.nudge(1)
		} else if rl.CheckCollisionPointRec(mouse, val) {
			// Focus is set by the app/demo via doc.SetFocus(sb) on click.
		}
	}

	if !sb.focused {
		return
	}
	key := func(k int32) bool {
		return rl.IsKeyPressed(k) || rl.IsKeyPressedRepeat(k)
	}
	switch {
	case key(rl.KeyUp), key(rl.KeyRight):
		sb.nudge(1)
	case key(rl.KeyDown), key(rl.KeyLeft):
		sb.nudge(-1)
	case key(rl.KeyPageUp):
		sb.nudge(10)
	case key(rl.KeyPageDown):
		sb.nudge(-10)
	}
}

// Layout is a no-op for this leaf widget.
func (sb *SpinBox) Layout() { sb.layoutDirty = false }

// Draw implements Node.Draw.
func (sb *SpinBox) Draw() { sb.drawInternal() }

func (sb *SpinBox) formatValue() string {
	v := sb.Value.Get()
	if sb.DecimalPlaces <= 0 {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', sb.DecimalPlaces, 64)
}

func (sb *SpinBox) drawInternal() {
	if sb.IsHidden() {
		return
	}
	state := StyleStateNone
	if sb.hovered {
		state |= StyleStateHover
	}
	if sb.focused {
		state |= StyleStateFocus
	}
	if sb.Disabled {
		state |= StyleStateDisabled
	}
	style, _ := sb.ResolveStyle(state)
	b := sb.Bounds()
	radius := style.CornerRadius
	if radius <= 0 {
		radius = 6
	}
	roundness := radius / (b.Height / 2)
	if roundness > 1 {
		roundness = 1
	}

	bg := style.BackgroundColor
	if bg.A == 0 {
		bg = rl.NewColor(255, 255, 255, 255)
	}
	border := style.BorderColor
	if border.A == 0 {
		border = rl.NewColor(210, 213, 228, 255)
	}
	rl.DrawRectangleRounded(b, roundness, 6, bg)
	rl.DrawRectangleRoundedLinesEx(b, roundness, 6, style.BorderWidth, border)

	btnStyle := style
	btnStyle.Bold = true
	btnStyle.FontSize = 18
	if btnStyle.TextColor.A == 0 {
		btnStyle.TextColor = rl.NewColor(79, 70, 229, 255)
	}
	drawSpinBtn(sb.decBounds(), "−", sb.hoverDec, btnStyle)
	drawSpinBtn(sb.incBounds(), "+", sb.hoverInc, btnStyle)

	valStyle := style
	valStyle.Bold = false
	if valStyle.FontSize <= 0 {
		valStyle.FontSize = 15
	}
	text := sb.formatValue()
	tw := measureTextS(text, valStyle)
	vb := sb.valueBounds()
	vx := int32(vb.X + (vb.Width-float32(tw))/2)
	vy := TextPosY(vb, valStyle)
	drawTextS(text, vx, vy, valStyle)
}

func drawSpinBtn(rect rl.Rectangle, label string, hover bool, style Style) {
	if hover {
		rl.DrawRectangleRounded(rect, 0.2, 4, rl.NewColor(0, 0, 0, 12))
	}
	tw := measureTextS(label, style)
	x := int32(rect.X + (rect.Width-float32(tw))/2)
	y := TextPosY(rect, style)
	drawTextS(label, x, y, style)
}
