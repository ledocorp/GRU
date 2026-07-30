// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	checkboxBorderW = float32(1)
	// CheckboxDefaultSize is the standard square used in forms and demos (24×24).
	CheckboxDefaultSize = float32(24)
	// CheckboxSizeCompact is a denser form control (18×18).
	CheckboxSizeCompact = float32(18)
	// CheckboxSizeSmall is the densest readable square (18×18). Prefer
	// CheckboxSizeCompact / CheckboxDefaultSize in forms — 16px is too small.
	CheckboxSizeSmall = float32(18)
)

// Checkbox is an interactive boolean toggle.
//
// Value is a Signal[bool]. Clicking the checkbox flips the value. Effects
// subscribed to Value.Get() re-run whenever the checkbox is toggled, allowing
// labels or other widgets to react automatically.
//
// Set Disabled to true to render the disabled Theme v2 state and ignore mouse
// interaction. The widget resolves through the "checkbox/default" component and
// applies the "checked" and "disabled" visual states.
//
// When a flex parent assigns bounds larger than the preferred square, paint and
// hit-testing stay on a centered square (see paintRect) so label rows do not
// get oversized hit targets or visual gaps.
//
// # LLM Prompt Template
//
//	notify := ui.NewCheckbox("notify", true, 0, 0, 24, 24)
//	notify.Value.Subscribe(func() { /* persist */ })
//	form.AddChild(notify)
//
// Demo scenes: **Form Demo**, **Batch 1**.
type Checkbox struct {
	Element
	Value       *Signal[bool]
	Disabled    bool
	hovered     bool
	lastHovered bool
}

// NewCheckbox creates a new Checkbox with the given initial value.
// Pass CheckboxDefaultSize for w and h in forms; smaller sizes (e.g. 20) work
// when PreferredWidth/Height are pinned for compact toolbars.
func NewCheckbox(id string, initialValue bool, x, y, w, h float32) *Checkbox {
	c := &Checkbox{
		Element: NewElement(id, x, y, w, h),
		Value:   NewSignal(initialValue),
	}
	if w > 0 {
		c.PreferredWidth = w
	} else if h > 0 {
		c.PreferredWidth = h
	}
	c.styleName = "checkbox"
	c.Element.SetStyleVariant("checkbox", "default")
	c.Value.Subscribe(func() { c.MarkDirty() })
	return c
}

// GetPreferredWidth implements flex width hinting (defaults to CheckboxDefaultSize).
func (c *Checkbox) GetPreferredWidth() float32 {
	if c.PreferredWidth > 0 {
		return c.PreferredWidth
	}
	if w := c.Bounds().Width; w > 0 {
		return w
	}
	return CheckboxDefaultSize
}

// GetPreferredHeight returns the preferred square height for flex rows.
func (c *Checkbox) GetPreferredHeight() float32 {
	return c.GetPreferredWidth()
}

// paintRect is the centered square used for draw and hit-testing.
func (c *Checkbox) paintRect() rl.Rectangle {
	b := snapControlRect(c.Bounds())
	w := c.GetPreferredWidth()
	h := c.GetPreferredHeight()
	if b.Width > 0 && w > b.Width {
		w = b.Width
	}
	if b.Height > 0 && h > b.Height {
		h = b.Height
	}
	if b.Width > w+0.5 || b.Height > h+0.5 {
		return snapControlRect(rl.NewRectangle(
			b.X+(b.Width-w)/2,
			b.Y+(b.Height-h)/2,
			w, h,
		))
	}
	return rl.NewRectangle(b.X, b.Y, w, h)
}

// Update implements Node.Update by handling mouse input.
func (c *Checkbox) Update(dt float32) {
	if c.IsHidden() {
		return
	}
	if c.Disabled {
		c.hovered = false
		return
	}

	mouse := rl.GetMousePosition()
	hit := c.paintRect()
	c.hovered = rl.CheckCollisionPointRec(mouse, hit)
	c.lastHovered = c.hovered

	if c.hovered && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		c.Value.Set(!c.Value.Get())
	}
}

// Layout implements Node.Layout (leaf — bounds come from parent).
func (c *Checkbox) Layout() { c.layoutDirty = false }

// InteractionOverlayActive implements InteractionOverlayPainter.
func (c *Checkbox) InteractionOverlayActive() bool {
	return !c.IsHidden() && !c.Disabled && c.hovered && !c.Value.Get()
}

// DrawInteractionOverlay paints unchecked hover border without busting the SSAA cache.
func (c *Checkbox) DrawInteractionOverlay() {
	if !c.InteractionOverlayActive() {
		return
	}
	bounds := c.paintRect()
	style, _ := c.ResolveStyle(StyleStateNone)
	shorter := bounds.Width
	if bounds.Height < shorter {
		shorter = bounds.Height
	}
	roundness := float32(0)
	if style.CornerRadius > 0 && shorter > 0 {
		r := style.CornerRadius
		if shorter < CheckboxDefaultSize-2 {
			r = r * (shorter / CheckboxDefaultSize)
			if r < 3 {
				r = 3
			}
		}
		roundness = r / (shorter / 2)
		if roundness > 1 {
			roundness = 1
		}
	}
	drawRoundedInsetBorder(bounds, roundness, checkboxBorderW, rl.NewColor(79, 70, 229, 255), style.BackgroundColor)
}

// Draw implements Node.Draw by rendering the checkbox.
func (c *Checkbox) Draw() {
	c.drawInternal()
}

func (c *Checkbox) drawInternal() {
	if c.IsHidden() {
		return
	}

	bounds := c.paintRect()

	state := StyleStateNone
	if c.Value.Get() {
		state |= StyleStateChecked
	}
	if c.Disabled {
		state |= StyleStateDisabled
	}
	style, _ := c.ResolveStyle(state)
	if c.Disabled {
		style = mergeStyle(style, disabledControlStyle())
	}

	shorter := bounds.Width
	if bounds.Height < shorter {
		shorter = bounds.Height
	}
	roundness := float32(0)
	if style.CornerRadius > 0 && shorter > 0 {
		r := style.CornerRadius
		if shorter < CheckboxDefaultSize-2 {
			r = r * (shorter / CheckboxDefaultSize)
			if r < 3 {
				r = 3
			}
		}
		roundness = r / (shorter / 2)
		if roundness > 1 {
			roundness = 1
		}
	}

	bw := checkboxBorderW
	borderCol := style.BorderColor
	fillCol := style.BackgroundColor
	if c.Value.Get() {
		borderCol = fillCol
	}

	drawRoundedInsetBorder(bounds, roundness, bw, borderCol, fillCol)

	inner := rl.NewRectangle(bounds.X+bw, bounds.Y+bw, bounds.Width-2*bw, bounds.Height-2*bw)
	if c.Value.Get() {
		pad := inner.Width * 0.13
		iconSize := inner.Width - 2*pad
		if iconSize < 8 {
			iconSize = 8
		}
		dst := rl.NewRectangle(inner.X+pad, inner.Y+pad, iconSize, iconSize)
		DrawPhosphorIcon(dst, PhosphorCheck, PhosphorBold, style.TextColor)
	}
}

// IsInteractive implements Node.IsInteractive.
func (c *Checkbox) IsInteractive() bool { return !c.Disabled }

// CheckboxCaption is a compact checkbox + single-line caption for toolbars.
// Layout and draw are self-contained so flex parents do not clip the label.
type CheckboxCaption struct {
	Element
	Text     string
	Value    *Signal[bool]
	cb       *Checkbox
	boxSize  float32
	fontSize int32
	gap      float32
}

// NewCheckboxCaption builds a toolbar-style checkbox row.
func NewCheckboxCaption(id, text string, value *Signal[bool], boxSize float32, fontSize int32, gap float32) *CheckboxCaption {
	if value == nil {
		value = NewSignal(false)
	}
	if boxSize <= 0 {
		boxSize = CheckboxDefaultSize
	}
	if fontSize <= 0 {
		fontSize = 15
	}
	cc := &CheckboxCaption{
		Element:  NewElement(id, 0, 0, 0, 0),
		Text:     text,
		Value:    value,
		boxSize:  boxSize,
		fontSize: fontSize,
		gap:      gap,
	}
	cc.cb = NewCheckbox(id+"-cb", value.Get(), 0, 0, boxSize, boxSize)
	cc.cb.Value = value
	value.Subscribe(func() { cc.MarkDrawDirty() })
	return cc
}

// Checkbox returns the inner toggle for enable/disable wiring.
func (cc *CheckboxCaption) Checkbox() *Checkbox { return cc.cb }

// SetDisabled greys out the control and ignores clicks.
func (cc *CheckboxCaption) SetDisabled(disabled bool) {
	if cc == nil || cc.cb == nil {
		return
	}
	cc.cb.Disabled = disabled
	cc.MarkDrawDirty()
}

// NaturalWidth measures the full caption row width.
func (cc *CheckboxCaption) NaturalWidth() float32 {
	st := cc.captionStyle()
	return cc.boxSize + cc.gap + float32(MeasureTextS(cc.Text, st)) + 4
}

// GetPreferredWidth implements flex width hinting.
func (cc *CheckboxCaption) GetPreferredWidth() float32 {
	if cc.PreferredWidth > 0 {
		return cc.PreferredWidth
	}
	return cc.NaturalWidth()
}

func (cc *CheckboxCaption) captionStyle() Style {
	st := GetThemeStyle("toolbar-caption")
	if cc.fontSize > 0 {
		st.FontSize = cc.fontSize
	}
	return st
}

func (cc *CheckboxCaption) Layout() {
	b := cc.Bounds()
	box := cc.boxSize
	if b.Height > 0 && box > b.Height {
		box = b.Height
	}
	cbY := b.Y + (b.Height-box)/2
	if cbY < b.Y {
		cbY = b.Y
	}
	cc.cb.setBoundsNoMark(rl.NewRectangle(b.X, cbY, box, box))
	cc.layoutDirty = false
}

func (cc *CheckboxCaption) Update(dt float32) {
	if cc.IsHidden() {
		return
	}
	cc.cb.Update(dt)
}

func (cc *CheckboxCaption) Draw() {
	if cc.IsHidden() {
		return
	}
	cc.cb.Draw()
	st := cc.captionStyle()
	if cc.cb.Disabled {
		st.TextColor = rl.NewColor(160, 164, 180, 255)
	}
	cb := cc.cb.Bounds()
	textX := cb.X + cb.Width + cc.gap
	textRect := rl.NewRectangle(textX, cc.Bounds().Y, cc.Bounds().X+cc.Bounds().Width-textX, cc.Bounds().Height)
	drawTextS(cc.Text, int32(textRect.X), toolbarTextPosY(textRect, st), st)
	cc.drawDirty = false
}

func (cc *CheckboxCaption) IsInteractive() bool { return true }

// InteractionOverlayActive implements InteractionOverlayPainter.
func (cc *CheckboxCaption) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (cc *CheckboxCaption) DrawInteractionOverlay() {}
