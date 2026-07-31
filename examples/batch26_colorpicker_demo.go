//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &batch26ColorPickerScene{} }) }

// batch26ColorPickerScene demonstrates ColorPicker (HSV popup) — split from Batch 2.
type batch26ColorPickerScene struct {
	BaseScene
}

func (s *batch26ColorPickerScene) Title() string { return "Batch 26 · ColorPicker" }

func (s *batch26ColorPickerScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5ColorPickerPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func cpCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

// cpSwatchRow is a Bootstrap input-group row: compact swatch + readout label.
func cpSwatchRow(id string, initial rl.Color, showAlpha bool, readoutFn func(rl.Color) string) *ui.Container {
	row := ui.NewContainer(id+"-row", 0, 0, 0, 0)
	row.FlexDirection = ui.FlexRow
	row.Gap = 12
	row.SetStyle("transparent")

	cp := ui.NewColorPicker(id, initial, 0, 0, 52, 36)
	cp.ShowAlpha = showAlpha

	readout := ui.NewLabel(id+"-readout", readoutFn(initial), 0, 0, 0, 36)
	readout.Align = ui.LabelAlignLeft
	readout.SetStyle("form-value")
	cp.Value.Subscribe(func() {
		readout.Text.Set(readoutFn(cp.Value.Get()))
		readout.MarkDirty()
	})

	row.AddChild(cp)
	row.AddChild(readout)
	return row
}

func (s *batch26ColorPickerScene) Build(doc *ui.Document) {
	PreloadScenePhosphor(doc)
	page := MountAppPage(doc, "b26",
		"Widget Batch 26 · ColorPicker",
		"Compact swatch opens an HSV popup. Preset rows without popup → Batch 18 · ColorWell.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b26-grid", 12)

	pBasic := ui.NewPanel("b26-basic", "Solid colors", 0, 0, 0, 0)
	pBasic.AutoHeight = true
	setSpans5ColorPickerPanel(pBasic, 12, 12, 6, 5, 5)
	pBasic.Gap = 8
	pBasic.TitleHeight = 32

	hexReadout := func(c rl.Color) string {
		return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
	}

	pBasic.AddChild(cpCaption("b26-basic-cap", "Click a 52×36 swatch to open the HSV popup. Esc or click outside to close."))
	pBasic.AddChild(cpSwatchRow("b26-indigo", rl.NewColor(79, 70, 229, 255), false, hexReadout))
	pBasic.AddChild(cpSwatchRow("b26-emerald", rl.NewColor(16, 185, 129, 255), false, hexReadout))

	pAlpha := ui.NewPanel("b26-alpha", "With alpha", 0, 0, 0, 0)
	pAlpha.AutoHeight = true
	setSpans5ColorPickerPanel(pAlpha, 12, 12, 6, 5, 5)
	pAlpha.Gap = 8
	pAlpha.TitleHeight = 32

	rgbaReadout := func(c rl.Color) string {
		return fmt.Sprintf("rgba(%d, %d, %d, %d)", c.R, c.G, c.B, c.A)
	}

	pAlpha.AddChild(cpCaption("b26-alpha-cap", "ShowAlpha adds an opacity slider in the popup."))
	pAlpha.AddChild(cpSwatchRow("b26-coral", rl.NewColor(239, 97, 87, 180), true, rgbaReadout))

	pForm := ui.NewPanel("b26-form", "Form row", 0, 0, 0, 0)
	pForm.AutoHeight = true
	setSpans5ColorPickerPanel(pForm, 12, 12, 12, 12, 12)
	pForm.Gap = 8
	pForm.TitleHeight = 32

	formRow := ui.NewContainer("b26-form-row", 0, 0, 0, 0)
	formRow.FlexDirection = ui.FlexRow
	formRow.Gap = 16
	formRow.SetStyle("transparent")
	lbl := ui.NewLabel("b26-label", "Border color", 0, 0, 120, 36)
	lbl.Align = ui.LabelAlignLeft
	lbl.SetStyle("form-label")
	border := ui.NewColorPicker("b26-border", rl.NewColor(100, 116, 139, 255), 0, 0, 52, 36)
	borderReadout := ui.NewLabel("b26-border-readout", "", 0, 0, 0, 36)
	borderReadout.Align = ui.LabelAlignLeft
	borderReadout.SetStyle("form-value")
	border.Value.Subscribe(func() {
		c := border.Value.Get()
		borderReadout.Text.Set(fmt.Sprintf("Selected #%02X%02X%02X", c.R, c.G, c.B))
		borderReadout.MarkDirty()
	})
	border.Value.Set(border.Value.Get())
	formRow.AddChild(lbl)
	formRow.AddChild(border)
	formRow.AddChild(borderReadout)

	pForm.AddChild(cpCaption("b26-form-cap", "Typical label + swatch + readout in a settings form."))
	pForm.AddChild(formRow)

	grid.AddChild(pBasic)
	grid.AddChild(pAlpha)
	grid.AddChild(pForm)
	page.Body.AddChild(grid)
}
