// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ─── Form ────────────────────────────────────────────────────────────────────
//
// Form is a labelled field layout container. Each row pairs a right-aligned
// label with a child widget. Bounds are computed by Layout so callers only
// need to size the Form itself.
//
//	f := ui.NewForm("settings", 10, 10, 500, 400)
//	f.AddField("Username", ui.NewTextInput("user", "", 0, 0, 0, 0))
//	f.AddField("Volume",   ui.NewSlider("vol", 0, 0, 0, 0, 0, 100))
//	f.AddField("Notify",   ui.NewToggle("tg",  false, 0, 0, 0, 0))
//
// # Layout Modes
//
// Two-column (default, Vertical = false): the label is right-aligned inside a
// fixed-width left column (LabelW). The control fills the remaining width.
//
// Vertical (Vertical = true): the label appears as a small-caps caption one
// row above the control; both span the full content width. This mode suits
// narrow forms or wide controls (TextInput, Dropdown).
//
// # Validation
//
// Call SetError(label, msg) to mark a field invalid. An error causes a 3-px
// red left-border on the control row and a small error message text below the
// widget. Call ClearError(label) or ClearAllErrors() to reset.
//
// # Style
//
// Theme keys used:
//   - "form"        — container background, border, padding, and corner radius
//   - "form-label"  — label text style (muted colour, 14 pt minimum)
//   - "form-error"  — inline error message text (red, 14 pt minimum)
//
// # Inspector integration
//
// The Inspector pane shows field count, layout direction, and per-field labels
// with [error] annotations.
//
// # Nesting
//
// Form works correctly inside a Viewport, Panel, Card, or Modal. Wrap the
// Form in a Viewport when the field list may exceed the visible height.

const formErrorLineH = float32(16) // vertical space reserved for an error message

// formField is one entry in the Form's field list.
type formField struct {
	label    string  // display label for this field
	widget   Node    // the control widget
	widgetW  float32 // 0 = fill available control-column width, >0 = fixed px width
	widgetH  float32 // 0 = use Form.RowH, >0 = fixed px height
	errorMsg string  // empty means "no error"
}

// Form is a labelled-field layout container.
type Form struct {
	Element

	// LabelW is the fixed width of the left label column in two-column mode.
	// Default: 130. Ignored when Vertical = true.
	LabelW float32

	// RowH is the allocated height for each control widget (default 36).
	RowH float32

	// LabelLineH is the label caption height in Vertical mode (default 20).
	LabelLineH float32

	// Gap is the vertical spacing between field rows (default 10).
	Gap float32

	// FieldGap is the gap between the label column and the control column in
	// two-column mode (default 12). Ignored when Vertical = true.
	FieldGap float32

	// Vertical switches to stacked layout: label caption above, control below.
	Vertical bool

	fields     []formField
	layoutAtW  float32 // last width fields were positioned for (reflow on change)
}

// NewForm creates a Form at the given position and size.
func NewForm(id string, x, y, w, h float32) *Form {
	f := &Form{
		Element:    NewElement(id, x, y, w, h),
		LabelW:     130,
		RowH:       36,
		LabelLineH: 20,
		Gap:        10,
		FieldGap:   12,
	}
	f.styleName = "form"
	return f
}

// SetStyle sets the base theme key (default: "form").
func (f *Form) SetStyle(name string) { f.styleName = name }

// AddField appends a labelled control row and returns the Form for chaining.
// The widget stretches to fill the available control-column width at Form.RowH.
// Use AddFieldSized when the widget should keep a compact natural size.
func (f *Form) AddField(label string, widget Node) *Form {
	return f.AddFieldSized(label, widget, 0, 0)
}

// AddFieldSized appends a labelled control row with explicit size overrides.
// Pass widgetW=0 to fill the control column; pass widgetH=0 to use Form.RowH.
// Useful for compact controls like Toggle (56×28) or Checkbox (28×28) that
// should not be stretched to fill the full column width.
func (f *Form) AddFieldSized(label string, widget Node, widgetW, widgetH float32) *Form {
	if widget != nil {
		// Register as a structural child so the Inspector tree and
		// parent-chain traversal (findViewport, etc.) work correctly.
		f.Element.AddChild(widget)
	}
	f.fields = append(f.fields, formField{
		label:   label,
		widget:  widget,
		widgetW: widgetW,
		widgetH: widgetH,
	})
	f.MarkDirty()
	return f
}

// SetError marks the field with the given label as invalid and shows msg below
// the control widget. Triggers a redraw+layout so the extra error line appears.
func (f *Form) SetError(label, msg string) {
	for i := range f.fields {
		if f.fields[i].label == label {
			if f.fields[i].errorMsg != msg {
				f.fields[i].errorMsg = msg
				f.MarkDirty()
			}
			return
		}
	}
}

// ClearError removes the error state for the field with the given label.
func (f *Form) ClearError(label string) {
	f.SetError(label, "")
}

// ClearAllErrors removes all field error states.
func (f *Form) ClearAllErrors() {
	changed := false
	for i := range f.fields {
		if f.fields[i].errorMsg != "" {
			f.fields[i].errorMsg = ""
			changed = true
		}
	}
	if changed {
		f.MarkDirty()
	}
}

// FieldCount returns the number of registered fields.
func (f *Form) FieldCount() int { return len(f.fields) }

// IsInteractive returns false — Form is a layout container; its children
// handle their own input.
func (f *Form) IsInteractive() bool { return false }

// UsesScissor returns false — Form does not open a scissor region.
func (f *Form) UsesScissor() bool { return false }

// Update delegates input processing to all child widgets.
func (f *Form) Update(dt float32) {
	if f.IsHidden() {
		return
	}
	UpdateChildrenOverlayAware(f.Children(), dt)
}

// Layout positions each control widget according to the current layout mode.
func (f *Form) Layout() {
	if f.IsHidden() {
		return
	}
	w := f.Bounds().Width
	widthChanged := f.layoutAtW > 1 && (w > f.layoutAtW+0.5 || w < f.layoutAtW-0.5)
	if !f.IsDirty() && !widthChanged {
		return
	}

	if f.AutoHeight {
		wantH := f.intrinsicHeight()
		b := f.Bounds()
		if b.Height < wantH-0.5 || b.Height > wantH+0.5 {
			b.Height = wantH
			f.bounds = b
		}
	}

	style := GetThemeStyle(f.styleName)
	pad := style.Padding
	if pad == 0 {
		pad = 10
	}

	b := f.Bounds()
	contentX := b.X + pad
	contentW := b.Width - 2*pad
	y := b.Y + pad

	for i := range f.fields {
		ff := &f.fields[i]
		if ff.widget == nil {
			y += f.rowTotalH(ff) + f.Gap
			continue
		}

		var wx, wy, ww, wh float32
		wh = f.RowH
		if ff.widgetH > 0 {
			wh = ff.widgetH
		}
		if f.Vertical {
			wx = contentX
			wy = y + f.LabelLineH + 4 // 4px gap between caption and control
			ww = contentW
			if ff.widgetW > 0 {
				ww = ff.widgetW
			}
		} else {
			wx = contentX + f.LabelW + f.FieldGap
			wy = y + (f.RowH-wh)/2 // vertically centre compact widgets
			ww = contentW - f.LabelW - f.FieldGap
			if ff.widgetW > 0 {
				ww = ff.widgetW
			}
		}
		ff.widget.SetBounds(rl.NewRectangle(wx, wy, ww, wh))
		ff.widget.Layout()

		y += f.rowTotalH(ff) + f.Gap
	}

	if w > 1 {
		f.layoutAtW = w
	}
	f.layoutDirty = false
}

// rowTotalH returns the full vertical space consumed by one field row,
// including the label caption (vertical mode), error line, and widget height.
func (f *Form) rowTotalH(ff *formField) float32 {
	wh := f.RowH
	if ff.widgetH > 0 {
		wh = ff.widgetH
	}
	h := wh
	if f.Vertical {
		h += f.LabelLineH + 4
	}
	if ff.errorMsg != "" {
		h += formErrorLineH
	}
	return h
}

// intrinsicHeight returns the padded height needed for all registered fields.
func (f *Form) intrinsicHeight() float32 {
	style := GetThemeStyle(f.styleName)
	pad := style.Padding
	if pad == 0 {
		pad = 10
	}
	if len(f.fields) == 0 {
		return pad * 2
	}
	var total float32
	for i := range f.fields {
		total += f.rowTotalH(&f.fields[i])
		if i < len(f.fields)-1 {
			total += f.Gap
		}
	}
	return total + 2*pad
}

// Draw renders the form background, labels, and all child widgets.
func (f *Form) Draw() {
	if f.IsHidden() {
		return
	}

	b := f.Bounds()
	style := GetThemeStyle(f.styleName)
	lblStyle := GetThemeStyle("form-label")
	errStyle := GetThemeStyle("form-error")
	pad := style.Padding
	if pad == 0 {
		pad = 10
	}

	// ── Container background ──────────────────────────────────────────────
	if style.CornerRadius > 0 {
		r := style.CornerRadius / b.Height
		if r > 0.5 {
			r = 0.5
		}
		if style.BackgroundColor.A > 0 {
			rl.DrawRectangleRounded(b, r, 8, style.BackgroundColor)
		}
		if style.BorderWidth > 0 {
			rl.DrawRectangleRoundedLinesEx(b, r, 8, style.BorderWidth, style.BorderColor)
		}
	} else {
		if style.BackgroundColor.A > 0 {
			rl.DrawRectangleRec(b, style.BackgroundColor)
		}
		if style.BorderWidth > 0 {
			rl.DrawRectangleLinesEx(b, style.BorderWidth, style.BorderColor)
		}
	}

	// ── Field rows ────────────────────────────────────────────────────────
	contentX := b.X + pad
	contentW := b.Width - 2*pad
	y := b.Y + pad

	// For viewport-scissor restore after each widget that uses a scissor region.
	vp := findViewport(f)
	var vpClip rl.Rectangle
	if vp != nil {
		vpClip = vp.ClipBounds()
	}

	for _, ff := range f.fields {
		rowH := f.rowTotalH(&ff)
		wh := f.RowH
		if ff.widgetH > 0 {
			wh = ff.widgetH
		}

		// Red left-accent bar for invalid fields.
		if ff.errorMsg != "" {
			accentRect := rl.NewRectangle(b.X, y, 3, rowH-formErrorLineH)
			rl.DrawRectangleRec(accentRect, rl.NewColor(220, 38, 38, 220))
		}

		if f.Vertical {
			lblFS := EffectiveFontSize(lblStyle)
			lx := int32(contentX)
			ly := int32(y) + (int32(f.LabelLineH)-int32(lblFS))/2
			drawTextS(ff.label, lx, ly, lblStyle)
		} else {
			lblFS := EffectiveFontSize(lblStyle)
			lw := float32(measureTextS(ff.label, lblStyle))
			lx := contentX + f.LabelW - lw
			if lx < contentX {
				lx = contentX
			}
			// Centre labels in the full row band (same as compact widgets), not
			// in widgetH alone — otherwise Notify/checkbox labels sit high.
			labelBand := f.RowH
			if ff.errorMsg != "" {
				labelBand = rowH - formErrorLineH
			}
			ly := y + (labelBand-lblFS)/2
			drawTextS(ff.label, int32(lx), int32(ly), lblStyle)

			divX := contentX + f.LabelW + f.FieldGap/2
			rl.DrawRectangleRec(
				rl.NewRectangle(divX, y+4, 1, wh-8),
				rl.NewColor(210, 213, 228, 180),
			)
		}

		// Draw child widget.
		if ff.widget != nil {
			ff.widget.Draw()
			// Restore the Viewport scissor if the widget clobbered it.
			// (TextInput, Slider, VirtualList all call EndScissorMode internally.)
			if vp != nil && ff.widget.UsesScissor() {
				beginScissorMode(
					int32(vpClip.X), int32(vpClip.Y),
					int32(vpClip.Width), int32(vpClip.Height),
				)
			}
		}

		// Inline error message below the control.
		if ff.errorMsg != "" {
			errY := y + wh
			if f.Vertical {
				errY += f.LabelLineH + 4
			}
			ex := int32(contentX + 4)
			ey := int32(errY) + 2
			drawTextS(ff.errorMsg, ex, ey, errStyle)
		}

		y += rowH + f.Gap
	}

	// Required width of all descendants drawn above must be within b — no
	// further action needed (Form does not scissor-clip its children).
	_ = contentW
}
