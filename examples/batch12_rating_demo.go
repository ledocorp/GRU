//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch12RatingScene{} }) }

// batch12RatingScene demonstrates the Rating star input widget.
type batch12RatingScene struct {
	BaseScene
}

func (s *batch12RatingScene) Title() string { return "Batch 12 · Rating" }

func (s *batch12RatingScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5RatingPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func rtCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func rtBody(id, text string) *ui.RichText { return batchCaption(id, text) }

func ratingRow(id, label string, value *ui.Signal[float32], stars int) (*ui.Container, *ui.RichText) {
	row := ui.NewContainer(id+"-row", 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 12
	row.SetStyle("transparent")

	lbl := ui.NewPlainText(id+"-lbl", "form-label", label, 0, 0, 0, 0)
	lbl.PreferredWidth = 100

	rating := ui.NewRating(id+"-rating", value, stars, 0, 0, 0, 0)
	rating.SetFlexGrow(1)

	status, statusDisplay := FlexCopyPair(id+"-val", "form-value", "")
	refresh := func() {
		statusDisplay.Set(fmt.Sprintf("%.0f / %d", value.Get(), stars))
	}
	value.Subscribe(refresh)
	refresh()

	row.AddChild(lbl)
	row.AddChild(rating)
	return row, status
}

func (s *batch12RatingScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b12",
		"Widget Batch 12 · Rating",
		"Click or hover stars to set a score. Uses Phosphor star icons with Unicode fallback.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b12-grid", 12)

	// ── Panel: Interactive ────────────────────────────────────────────────────
	pMain := ui.NewPanel("b12-main", "Interactive", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5RatingPanel(pMain, 12, 12, 6, 4, 4)
	pMain.Gap = 10
	pMain.TitleHeight = 32

	pMain.AddChild(rtBody("b12-intro", "Each row is an independent rating control backed by a Signal."))

	overall := ui.NewSignal(float32(4))
	row1, st1 := ratingRow("b12-overall", "Overall", overall, 5)
	pMain.AddChild(row1)
	pMain.AddChild(st1)

	quality := ui.NewSignal(float32(3))
	row2, st2 := ratingRow("b12-quality", "Quality", quality, 5)
	pMain.AddChild(row2)
	pMain.AddChild(st2)

	btnRow := ui.NewContainer("b12-btns", 0, 0, 0, 0)
	btnRow.FlexDirection = ui.FlexRow
	btnRow.Gap = 8
	btnRow.SetStyle("transparent")
	btnReset := ui.NewButton("b12-reset", "Reset all", 0, 0, 0, 34)
	btnReset.SetStyle("button")
	btnReset.OnClick = func() {
		overall.Set(0)
		quality.Set(0)
	}
	btnFive := ui.NewButton("b12-five", "Set 5 stars", 0, 0, 0, 34)
	btnFive.SetStyle("primary")
	btnFive.OnClick = func() {
		overall.Set(5)
		quality.Set(5)
	}
	btnRow.AddChild(btnReset)
	btnRow.AddChild(btnFive)
	pMain.AddChild(btnRow)

	// ── Panel: Star counts ────────────────────────────────────────────────────
	pCounts := ui.NewPanel("b12-counts", "Star counts", 0, 0, 0, 0)
	pCounts.AutoHeight = true
	setSpans5RatingPanel(pCounts, 12, 12, 6, 4, 4)
	pCounts.Gap = 10
	pCounts.TitleHeight = 32

	pCounts.AddChild(rtCaption("b12-count-cap", "Max stars changes the strip width (intrinsic sizing)."))

	three := ui.NewSignal(float32(2))
	row3, st3 := ratingRow("b12-three", "3-star", three, 3)
	pCounts.AddChild(row3)
	pCounts.AddChild(st3)

	ten := ui.NewSignal(float32(7))
	row10, st10 := ratingRow("b12-ten", "10-star", ten, 10)
	pCounts.AddChild(row10)
	pCounts.AddChild(st10)

	// ── Panel: Reactive readout ───────────────────────────────────────────────
	pReactive := ui.NewPanel("b12-reactive", "Reactive", 0, 0, 0, 0)
	pReactive.AutoHeight = true
	setSpans5RatingPanel(pReactive, 12, 12, 12, 4, 4)
	pReactive.Gap = 10
	pReactive.TitleHeight = 32

	score := ui.NewSignal(float32(0))
	drpRating := ui.NewRating("b12-live", score, 5, 0, 0, 0, 0)

	liveLbl := ui.NewLabel("b12-live-lbl", "Score: 0 — drag stars or use buttons.", 0, 0, 0, 0)
	liveLbl.SetStyle("form-value")
	liveLbl.Wrap = true
	score.Subscribe(func() {
		v := score.Get()
		msg := "Unrated"
		if v > 0 {
			msg = fmt.Sprintf("%.0f star", v)
			if v != 1 {
				msg += "s"
			}
		}
		liveLbl.Text.Set("Score: " + msg)
	})

	pReactive.AddChild(rtBody("b12-react-note", "External code can write score.Set(n) — the widget redraws from the signal."))
	pReactive.AddChild(drpRating)
	pReactive.AddChild(liveLbl)

	presetRow := ui.NewContainer("b12-presets", 0, 0, 0, 0)
	presetRow.FlexDirection = ui.FlexRow
	presetRow.SetFlexWrap(true)
	presetRow.Gap = 8
	presetRow.SetStyle("transparent")
	for i := 1; i <= 5; i++ {
		n := float32(i)
		btn := ui.NewButton(fmt.Sprintf("b12-preset-%d", i), fmt.Sprintf("%d ★", i), 0, 0, 0, 34)
		btn.SetStyle("button")
		btn.OnClick = func() { score.Set(n) }
		presetRow.AddChild(btn)
	}
	pReactive.AddChild(presetRow)

	grid.AddChild(pMain)
	grid.AddChild(pCounts)
	grid.AddChild(pReactive)
	page.Body.AddChild(grid)
}
