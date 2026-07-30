// Package ui (continued)
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// PropertyRow is one key/value pair in a PropertyTable.
type PropertyRow struct {
	Key   string
	Value *Signal[string]
}

// PropertyTable is a scrollable key/value editor (Strategy 2 #31).
// Rows are laid out via the shared Form container so labels and TextInputs
// align and clip like the rest of the toolkit.
//
// # LLM Prompt Template
//
//	rows := []ui.PropertyRow{{Key: "Theme", Value: ui.NewSignal("Dark")}}
//	pt := ui.NewPropertyTable("props", rows, 0, 0, 360, 200)
//	panel.AddChild(pt)
//
// Demo scenes: **Batch 10** (PropertyTable panel).
type PropertyTable struct {
	Element
	Rows []PropertyRow
	form *Form
}

// NewPropertyTable creates a property table with one TextInput per row.
func NewPropertyTable(id string, rows []PropertyRow, x, y, w, h float32) *PropertyTable {
	pt := &PropertyTable{
		Element: NewElement(id, x, y, w, h),
		Rows:    rows,
	}
	pt.styleName = "propertytable"
	pt.form = NewForm(id+"-form", 0, 0, w, h)
	pt.form.SetStyle("transparent")
	pt.form.LabelW = 96
	pt.form.RowH = 36
	pt.form.Gap = 8
	pt.form.FieldGap = 8
	pt.AddChild(pt.form)

	for _, r := range rows {
		val := ""
		if r.Value != nil {
			val = r.Value.Get()
		}
		ti := NewTextInput(id+"-v-"+r.Key, val, 0, 0, 0, 0)
		ti.SetStyle("input")
		if r.Value != nil {
			sig := r.Value
			ti.Text.Subscribe(func() { sig.Set(ti.Text.Get()) })
			sig.Subscribe(func() {
				if !ti.IsFocused() {
					ti.Text.Set(sig.Get())
				}
			})
		}
		pt.form.AddField(r.Key, ti)
	}
	return pt
}

// Children implements Node.
func (pt *PropertyTable) Children() []Node {
	if pt.form == nil {
		return nil
	}
	return []Node{pt.form}
}

// IsInteractive implements Node.
func (pt *PropertyTable) IsInteractive() bool { return len(pt.Rows) > 0 }

func (pt *PropertyTable) Layout() {
	if pt.IsHidden() || pt.form == nil {
		return
	}
	if !pt.IsDirty() {
		return
	}
	layoutSetBounds(pt.form, pt.Bounds())
	pt.form.MarkDirty()
	pt.form.Layout()
	pt.layoutDirty = false
}

func (pt *PropertyTable) Update(dt float32) {
	if pt.IsHidden() {
		return
	}
	pt.form.Update(dt)
}

func (pt *PropertyTable) Draw() { pt.drawInternal() }

func (pt *PropertyTable) drawInternal() {
	if pt.IsHidden() || pt.form == nil {
		return
	}
	b := pt.Bounds()
	style := pt.GetStyle()
	if style.BackgroundColor.A > 0 {
		rl.DrawRectangleRec(b, style.BackgroundColor)
	}
	if style.BorderWidth > 0 && style.BorderColor.A > 0 {
		rl.DrawRectangleLinesEx(b, style.BorderWidth, style.BorderColor)
	}
	pt.form.Draw()
}
