//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"time"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch11DateRangeScene{} }) }

// batch11DateRangeScene demonstrates DateRangePicker and the global range calendar.
type batch11DateRangeScene struct {
	BaseScene
}

func (s *batch11DateRangeScene) Title() string { return "Batch 11 · DateRangePicker" }

func (s *batch11DateRangeScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5RangePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func drCaption(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-label",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func drBody(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-value",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func (s *batch11DateRangeScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b11",
		"Widget Batch 11 · DateRangePicker",
		"Pick a start day, then an end day on one calendar. Popup draws in screen space like DatePicker.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b11-grid", 16)

	// ── Panel: Trip dates ─────────────────────────────────────────────────────
	pTrip := ui.NewPanel("b11-trip", "Trip dates", 0, 0, 0, 0)
	pTrip.AutoHeight = true
	setSpans5RangePanel(pTrip, 12, 12, 6, 7, 7)
	pTrip.Gap = 10
	pTrip.TitleHeight = 32

	start := ui.NewSignal(time.Date(2026, 6, 10, 0, 0, 0, 0, time.Local))
	end := ui.NewSignal(time.Date(2026, 6, 17, 0, 0, 0, 0, time.Local))
	drp := ui.NewDateRangePicker("b11-range", start, end, 0, 0, 0, 38)

	status := ui.NewLabel("b11-status", "", 0, 0, 0, 0)
	status.SetStyle("form-value")
	status.Wrap = true
	refresh := func() {
		s0, e0 := start.Get(), end.Get()
		if s0.IsZero() && e0.IsZero() {
			status.Text.Set("Range: (none)")
			return
		}
		if s0.IsZero() || e0.IsZero() {
			status.Text.Set("Range: (incomplete — pick both endpoints)")
			return
		}
		if e0.Before(s0) {
			s0, e0 = e0, s0
		}
		days := int(e0.Sub(s0).Hours()/24) + 1
		status.Text.Set(fmt.Sprintf("Range: %s → %s (%d nights)",
			s0.Format("2006-01-02"), e0.Format("2006-01-02"), days-1))
	}
	start.Subscribe(refresh)
	end.Subscribe(refresh)
	refresh()

	pTrip.AddChild(drCaption("b11-trip-cap", "Click start, then end. Range commits when the second day is chosen."))
	pTrip.AddChild(drp)
	pTrip.AddChild(status)

	// ── Panel: Variants ───────────────────────────────────────────────────────
	pVar := ui.NewPanel("b11-variants", "Variants", 0, 0, 0, 0)
	pVar.AutoHeight = true
	setSpans5RangePanel(pVar, 12, 12, 6, 5, 5)
	pVar.Gap = 10
	pVar.TitleHeight = 32

	reportStart := ui.NewSignal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	reportEnd := ui.NewSignal(time.Date(2026, 1, 31, 0, 0, 0, 0, time.Local))
	drpReport := ui.NewDateRangePicker("b11-report", reportStart, reportEnd, 0, 0, 0, 36)
	drpReport.DateFormat = "01/02/2006"

	emptyStart := ui.NewSignal(time.Time{})
	emptyEnd := ui.NewSignal(time.Time{})
	drpEmpty := ui.NewDateRangePicker("b11-empty", emptyStart, emptyEnd, 0, 0, 0, 36)
	drpEmpty.Placeholder = "Filter by date range…"

	narrowRow := ui.NewContainer("b11-narrow", 0, 0, 0, 36)
	narrowRow.FlexDirection = ui.FlexRow
	narrowRow.SetStyle("transparent")
	nStart := ui.NewSignal(time.Date(2026, 3, 1, 0, 0, 0, 0, time.Local))
	nEnd := ui.NewSignal(time.Date(2026, 3, 7, 0, 0, 0, 0, time.Local))
	drpNarrow := ui.NewDateRangePicker("b11-narrow-drp", nStart, nEnd, 0, 0, 240, 36)
	narrowRow.AddChild(drpNarrow)

	pVar.AddChild(drCaption("b11-v1", "US format · MM/DD/YYYY"))
	pVar.AddChild(drpReport)
	pVar.AddChild(drCaption("b11-v2", "Empty value · custom placeholder"))
	pVar.AddChild(drpEmpty)
	pVar.AddChild(drCaption("b11-v3", "Fixed 240 px wide field"))
	pVar.AddChild(narrowRow)
	pVar.AddChild(drBody("b11-hint",
		"Keyboard: arrows move focus · PgUp/PgDn change month · Esc cancels · click outside closes."))

	grid.AddChild(pTrip)
	grid.AddChild(pVar)
	page.Body.AddChild(grid)
}
