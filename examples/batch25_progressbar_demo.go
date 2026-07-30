//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch25ProgressBarScene{} }) }

// batch25ProgressBarScene demonstrates ProgressBar fill and captions.
// Recipe: CP-SHELL-PAGE + NewBatchPageGrid.
type batch25ProgressBarScene struct {
	BaseScene
	upload *ui.Signal[float32]
}

func (s *batch25ProgressBarScene) Title() string { return "Batch 25 · ProgressBar" }

func (s *batch25ProgressBarScene) OnUpdate(_ *ui.Document, dt float32) {
	if s.upload == nil {
		return
	}
	v := s.upload.Get()
	if v >= 1 {
		return
	}
	v += dt * 0.12
	if v > 1 {
		v = 1
	}
	s.upload.Set(v)
	ui.Wake(ui.WakeAnimation, "b25-upload")
}

func setSpans5ProgressPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func (s *batch25ProgressBarScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b25",
		"ProgressBar",
		"Read-only fill bar — value is 0.0–1.0; label shows percent when tall enough.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b25-grid", 12)

	pUpload := ui.NewPanel("b25-up", "Upload", 0, 0, 0, 0)
	pUpload.AutoHeight = true
	setSpans5ProgressPanel(pUpload, 12, 12, 6, 5, 5)
	pUpload.Gap = 10
	pUpload.TitleHeight = 32

	s.upload = ui.NewSignal(float32(0.35))
	bar := ui.NewProgressBar("b25-bar", s.upload.Get(), 0, 0, 0, 24)
	bar.Value = s.upload
	status, statusDisplay := FlexCopyPair("b25-status", "form-value", "")
	s.upload.Subscribe(func() {
		statusDisplay.Set(fmt.Sprintf("Transferring backup.zip — %d%%", int(s.upload.Get()*100)))
	})
	s.upload.Set(s.upload.Get())

	reset := ui.NewButton("b25-reset", "Restart demo", 0, 0, 120, 36)
	reset.SetStyle("button")
	reset.OnClick = func() {
		s.upload.Set(0)
	}

	pUpload.AddChild(batchCaption("b25-cap", "Non-interactive — bind Value to a Signal or drive from OnUpdate."))
	pUpload.AddChild(bar)
	pUpload.AddChild(status)
	pUpload.AddChild(reset)

	pSteps := ui.NewPanel("b25-steps", "Step milestones", 0, 0, 0, 0)
	pSteps.AutoHeight = true
	setSpans5ProgressPanel(pSteps, 12, 12, 6, 5, 5)
	pSteps.Gap = 10
	pSteps.TitleHeight = 32
	labels := []string{"Indexing", "Compressing", "Uploading"}
	vals := []float32{1, 0.72, 0.18}
	for i, lbl := range labels {
		pSteps.AddChild(ui.NewPlainText(fmt.Sprintf("b25-lbl-%d", i), "form-label", lbl, 0, 0, 0, 0))
		pSteps.AddChild(ui.NewProgressBar(fmt.Sprintf("b25-pb-%d", i), vals[i], 0, 0, 0, 20))
	}

	pThin := ui.NewPanel("b25-thin", "Thin track", 0, 0, 0, 0)
	pThin.AutoHeight = true
	setSpans5ProgressPanel(pThin, 12, 12, 12, 12, 12)
	pThin.Gap = 10
	pThin.TitleHeight = 32
	thin := ui.NewProgressBar("b25-thin-bar", 0.55, 0, 0, 0, 12)
	pThin.AddChild(batchCaption("b25-thin-cap", "Under 18px tall the percent label is hidden — use a separate caption."))
	pThin.AddChild(thin)

	grid.AddChild(pUpload)
	grid.AddChild(pSteps)
	grid.AddChild(pThin)
	page.Body.AddChild(grid)
}
