// Package ui (continued)
package ui

import (
	"math"
	"strconv"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	tbSpinBoxW     = float32(120)
	tbSpinBtnW     = float32(28) // wide − / + targets for easier clicks
	tbSpinDividerInset = float32(0.18) // top/bottom inset; smaller = longer dividers
)

// ToolbarSpinBox is one toolbar control: a single chrome frame with − value + zones.
// Hit-testing and drawing are unified; use [Toolbar.AddSpinBox] to place it on a bar.
type ToolbarSpinBox struct {
	Element
	Value         *Signal[float64]
	Min           float64
	Max           float64
	Step          float64
	DecimalPlaces int
	Disabled      bool
	hovered       bool
	hoverDec      bool
	hoverInc      bool
}

// NewToolbarSpinBox creates a toolbar-height numeric stepper.
func NewToolbarSpinBox(id string, value *Signal[float64], min, max, step float64) *ToolbarSpinBox {
	if value == nil {
		value = NewSignal(min)
	}
	if step <= 0 {
		step = 1
	}
	sb := &ToolbarSpinBox{
		Element: NewElement(id, 0, 0, tbSpinBoxW, 0),
		Value:   value,
		Min:     min,
		Max:     max,
		Step:    step,
	}
	sb.AutoHeight = false
	sb.styleName = "toolbar-menu"
	value.Subscribe(func() { sb.MarkDrawDirty() })
	return sb
}

// NaturalWidth implements toolbar width budgeting.
func (sb *ToolbarSpinBox) NaturalWidth() float32 { return tbSpinBoxW }

func (sb *ToolbarSpinBox) decBounds() rl.Rectangle {
	b := sb.Bounds()
	return rl.NewRectangle(b.X, b.Y, tbSpinBtnW, b.Height)
}

func (sb *ToolbarSpinBox) incBounds() rl.Rectangle {
	b := sb.Bounds()
	return rl.NewRectangle(b.X+b.Width-tbSpinBtnW, b.Y, tbSpinBtnW, b.Height)
}

func (sb *ToolbarSpinBox) valueBounds() rl.Rectangle {
	b := sb.Bounds()
	return rl.NewRectangle(b.X+tbSpinBtnW, b.Y, b.Width-2*tbSpinBtnW, b.Height)
}

func (sb *ToolbarSpinBox) clampAndSet(v float64) {
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

func (sb *ToolbarSpinBox) nudge(mult float64) {
	sb.clampAndSet(sb.Value.Get() + sb.Step*mult)
}

func (sb *ToolbarSpinBox) formatValue() string {
	v := sb.Value.Get()
	if sb.DecimalPlaces <= 0 {
		return strconv.FormatInt(int64(math.Round(v)), 10)
	}
	return strconv.FormatFloat(v, 'f', sb.DecimalPlaces, 64)
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (sb *ToolbarSpinBox) ClearOverlayPointerState() {
	if !sb.hovered && !sb.hoverDec && !sb.hoverInc {
		return
	}
	sb.hovered = false
	sb.hoverDec = false
	sb.hoverInc = false
	sb.MarkDrawDirty()
}

func (sb *ToolbarSpinBox) Update(_ float32) {
	if sb.IsHidden() || sb.Disabled {
		return
	}
	mouse := rl.GetMousePosition()
	dec := sb.decBounds()
	inc := sb.incBounds()
	prevD, prevI := sb.hoverDec, sb.hoverInc
	sb.hoverDec = rl.CheckCollisionPointRec(mouse, dec)
	sb.hoverInc = rl.CheckCollisionPointRec(mouse, inc)
	sb.hovered = sb.hoverDec || sb.hoverInc
	if sb.hoverDec != prevD || sb.hoverInc != prevI {
		sb.MarkDrawDirty()
	}
	if rl.IsMouseButtonPressed(rl.MouseButtonLeft) {
		if sb.hoverDec {
			sb.nudge(-1)
		} else if sb.hoverInc {
			sb.nudge(1)
		}
	}
}

func (sb *ToolbarSpinBox) Layout() { sb.layoutDirty = false }

func (sb *ToolbarSpinBox) Draw() {
	if sb.IsHidden() {
		return
	}
	style := GetThemeStyle("toolbar-menu")
	b := sb.Bounds()
	drawToolbarChrome(b, "toolbar-btn", style, sb.hovered, false, false)

	divCol := rl.NewColor(222, 226, 230, 255)
	pad := b.Height * tbSpinDividerInset
	divY0 := b.Y + pad
	divY1 := b.Y + b.Height - pad
	rl.DrawLineEx(rl.NewVector2(b.X+tbSpinBtnW, divY0), rl.NewVector2(b.X+tbSpinBtnW, divY1), 1, divCol)
	rl.DrawLineEx(rl.NewVector2(b.X+b.Width-tbSpinBtnW, divY0), rl.NewVector2(b.X+b.Width-tbSpinBtnW, divY1), 1, divCol)

	btnStyle := style
	btnStyle.Bold = true
	btnStyle.FontSize = 14
	btnStyle.TextColor = rl.NewColor(79, 70, 229, 255)

	drawToolbarSpinBtn(sb.decBounds(), "−", sb.hoverDec, btnStyle)
	drawToolbarSpinBtn(sb.incBounds(), "+", sb.hoverInc, btnStyle)

	valStyle := style
	valStyle.Bold = false
	text := sb.formatValue()
	tw := measureTextS(text, valStyle)
	vb := sb.valueBounds()
	vx := int32(vb.X + (vb.Width-float32(tw))/2)
	vy := toolbarTextPosY(vb, valStyle)
	drawTextS(text, vx, vy, valStyle)
}

func drawToolbarSpinBtn(rect rl.Rectangle, label string, hover bool, style Style) {
	if hover {
		rl.DrawRectangleRec(rect, rl.NewColor(79, 70, 229, 20))
	}
	tw := measureTextS(label, style)
	x := int32(rect.X + (rect.Width-float32(tw))/2)
	y := toolbarTextPosY(rect, style)
	drawTextS(label, x, y, style)
}

func (sb *ToolbarSpinBox) IsInteractive() bool { return true }
