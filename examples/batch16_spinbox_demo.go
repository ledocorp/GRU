//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch16SpinBoxScene{} }) }

type batch16SpinBoxScene struct {
	BaseScene
}

func (s *batch16SpinBoxScene) Title() string { return "Batch 16 · SpinBox" }

func (s *batch16SpinBoxScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5SpinPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func spinCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func (s *batch16SpinBoxScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b16",
		"Widget Batch 16 · SpinBox",
		"Numeric stepper with − / + buttons, clamped range, and keyboard nudging when focused.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b16-grid", 12)

	pQty := ui.NewPanel("b16-qty", "Quantity", 0, 0, 0, 0)
	pQty.AutoHeight = true
	setSpans5SpinPanel(pQty, 12, 12, 6, 4, 4)
	pQty.Gap = 10
	pQty.TitleHeight = 32
	qty := ui.NewSignal(1.0)
	qtySpin := ui.NewSpinBox("b16-qty-spin", qty, 1, 99, 1, 0, 0, 180, 36)
	qtyLbl := ui.NewLabel("b16-qty-lbl", "", 0, 0, 0, 0)
	qtyLbl.SetStyle("form-value")
	qty.Subscribe(func() {
		qtyLbl.Text.Set(fmt.Sprintf("Cart quantity: %.0f", qty.Get()))
	})
	qty.Set(qty.Get())
	pQty.AddChild(spinCaption("b16-qty-cap", "Integer steps from 1 to 99."))
	pQty.AddChild(qtySpin)
	pQty.AddChild(qtyLbl)

	pVol := ui.NewPanel("b16-vol", "Volume", 0, 0, 0, 0)
	pVol.AutoHeight = true
	setSpans5SpinPanel(pVol, 12, 12, 6, 4, 4)
	pVol.Gap = 10
	pVol.TitleHeight = 32
	vol := ui.NewSignal(50.0)
	volSpin := ui.NewSpinBox("b16-vol-spin", vol, 0, 100, 0.5, 0, 0, 200, 36)
	volSpin.DecimalPlaces = 1
	volLbl := ui.NewLabel("b16-vol-lbl", "", 0, 0, 0, 0)
	volLbl.SetStyle("form-value")
	vol.Subscribe(func() {
		volLbl.Text.Set(fmt.Sprintf("Level: %.1f%%", vol.Get()))
	})
	vol.Set(vol.Get())
	pVol.AddChild(spinCaption("b16-vol-cap", "Half-point steps with one decimal place."))
	pVol.AddChild(volSpin)
	pVol.AddChild(volLbl)

	pFocus := ui.NewPanel("b16-focus", "Keyboard", 0, 0, 0, 0)
	pFocus.AutoHeight = true
	setSpans5SpinPanel(pFocus, 12, 12, 12, 4, 4)
	pFocus.Gap = 10
	pFocus.TitleHeight = 32
	port := ui.NewSignal(8080.0)
	portSpin := ui.NewSpinBox("b16-port", port, 1024, 65535, 1, 0, 0, 220, 36)
	portSpin.DecimalPlaces = 0
	pFocus.AddChild(spinCaption("b16-focus-cap",
		"Click the value column to focus, then use arrow keys (Page Up/Down jumps by 10)."))
	pFocus.AddChild(portSpin)

	grid.AddChild(pQty)
	grid.AddChild(pVol)
	grid.AddChild(pFocus)
	page.Body.AddChild(grid)
}
