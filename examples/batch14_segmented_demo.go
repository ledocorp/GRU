//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch14SegmentedScene{} }) }

type batch14SegmentedScene struct {
	BaseScene
}

func (s *batch14SegmentedScene) Title() string { return "Batch 14 · SegmentedControl" }

func (s *batch14SegmentedScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5SegPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func segBody(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{Text: text, Variant: "form-value"}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func (s *batch14SegmentedScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b14",
		"Widget Batch 14 · SegmentedControl",
		"Mutually exclusive pill tabs on one rounded track.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b14-grid", 12)

	pFilter := ui.NewPanel("b14-filter", "Filters", 0, 0, 0, 0)
	pFilter.AutoHeight = true
	setSpans5SegPanel(pFilter, 12, 12, 6, 4, 4)
	pFilter.Gap = 10
	pFilter.TitleHeight = 32
	filter := ui.NewSignal(0)
	seg := ui.NewSegmentedControl("b14-filter-seg", []string{"All", "Active", "Done"}, filter, 0, 0, 0, 36)
	filterLbl := ui.NewLabel("b14-filter-lbl", "Selected: All", 0, 0, 0, 0)
	filterLbl.SetStyle("form-value")
	filter.Subscribe(func() {
		names := []string{"All", "Active", "Done"}
		i := filter.Get()
		if i < 0 || i >= len(names) {
			i = 0
		}
		filterLbl.Text.Set("Selected: " + names[i])
	})
	pFilter.AddChild(segBody("b14-filter-cap", "Click a segment — only one option stays selected."))
	pFilter.AddChild(seg)
	pFilter.AddChild(filterLbl)

	pView := ui.NewPanel("b14-view", "View modes", 0, 0, 0, 0)
	pView.AutoHeight = true
	setSpans5SegPanel(pView, 12, 12, 6, 4, 4)
	pView.Gap = 10
	pView.TitleHeight = 32
	view := ui.NewSignal(1)
	viewSeg := ui.NewSegmentedControl("b14-view-seg", []string{"List", "Grid", "Board"}, view, 0, 0, 280, 36)
	viewLbl := ui.NewLabel("b14-view-lbl", "", 0, 0, 0, 0)
	viewLbl.SetStyle("form-value")
	view.Subscribe(func() {
		viewLbl.Text.Set(fmt.Sprintf("View index: %d", view.Get()))
	})
	view.Set(view.Get())
	pView.AddChild(viewSeg)
	pView.AddChild(viewLbl)

	pCompact := ui.NewPanel("b14-compact", "Compact", 0, 0, 0, 0)
	pCompact.AutoHeight = true
	setSpans5SegPanel(pCompact, 12, 12, 12, 4, 4)
	pCompact.Gap = 10
	pCompact.TitleHeight = 32
	period := ui.NewSignal(0)
	periodSeg := ui.NewSegmentedControl("b14-period", []string{"Day", "Week", "Month", "Year"}, period, 0, 0, 360, 36)
	pCompact.AddChild(periodSeg)

	grid.AddChild(pFilter)
	grid.AddChild(pView)
	grid.AddChild(pCompact)
	page.Body.AddChild(grid)
}
