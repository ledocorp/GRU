//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch24SliderScene{} }) }

type batch24SliderScene struct {
	BaseScene
}

func (s *batch24SliderScene) Title() string { return "Batch 24 · Slider" }

func (s *batch24SliderScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5SliderPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func sliderCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func (s *batch24SliderScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b24",
		"Widget Batch 24 · Slider",
		"Drag the thumb to pick a value in a range — live value label on the right.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b24-grid", 12)

	pVol := ui.NewPanel("b24-vol", "Volume", 0, 0, 0, 0)
	pVol.AutoHeight = true
	setSpans5SliderPanel(pVol, 12, 12, 6, 5, 5)
	pVol.Gap = 10
	pVol.TitleHeight = 32

	volSlider := ui.NewSlider("b24-vol-sl", 0, 100, 62, 0, 0, 0, 36)
	volSlider.ValueFmt = "%.0f"
	volLbl := ui.NewLabel("b24-vol-lbl", "", 0, 0, 0, 0)
	volLbl.SetStyle("form-value")
	volLbl.Wrap = true
	volSlider.Value.Subscribe(func() {
		volLbl.Text.Set(fmt.Sprintf("Output level: %d%%", int(volSlider.Value.Get())))
	})
	volSlider.Value.Set(volSlider.Value.Get())

	pVol.AddChild(sliderCaption("b24-cap", "Click the track or drag the thumb — value updates every frame while held."))
	pVol.AddChild(volSlider)
	pVol.AddChild(volLbl)

	pFine := ui.NewPanel("b24-fine", "Fine control", 0, 0, 0, 0)
	pFine.AutoHeight = true
	setSpans5SliderPanel(pFine, 12, 12, 6, 5, 5)
	pFine.Gap = 10
	pFine.TitleHeight = 32
	opSlider := ui.NewSlider("b24-op", 0, 1, 0.85, 0, 0, 0, 36)
	opSlider.ValueFmt = "%.2f"
	opSlider.ShowValue = true
	pFine.AddChild(sliderCaption("b24-fine-cap", "Decimal format via ValueFmt — good for opacity, mix, and balance."))
	pFine.AddChild(opSlider)

	pQuiet := ui.NewPanel("b24-quiet", "No value column", 0, 0, 0, 0)
	pQuiet.AutoHeight = true
	setSpans5SliderPanel(pQuiet, 12, 12, 12, 12, 12)
	pQuiet.Gap = 10
	pQuiet.TitleHeight = 32
	balSlider := ui.NewSlider("b24-bal", -50, 50, 0, 0, 0, 0, 36)
	balSlider.ShowValue = false
	balNote := ui.NewLabel("b24-bal-note", "Centered — drag to pan left/right", 0, 0, 0, 0)
	balNote.SetStyle("form-value")
	balSlider.Value.Subscribe(func() {
		v := int(balSlider.Value.Get())
		if v == 0 {
			balNote.Text.Set("Centered — drag to pan left/right")
		} else if v > 0 {
			balNote.Text.Set(fmt.Sprintf("Pan right %d", v))
		} else {
			balNote.Text.Set(fmt.Sprintf("Pan left %d", -v))
		}
	})

	pQuiet.AddChild(sliderCaption("b24-quiet-cap", "ShowValue=false — pair with your own label when the built-in column is too narrow."))
	pQuiet.AddChild(balSlider)
	pQuiet.AddChild(balNote)

	grid.AddChild(pVol)
	grid.AddChild(pFine)
	grid.AddChild(pQuiet)
	page.Body.AddChild(grid)
}
