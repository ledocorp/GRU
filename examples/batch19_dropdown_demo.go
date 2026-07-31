//go:build !notepad

// Package examples (continued)
package examples

import (
	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch19DropdownScene{} }) }

type batch19DropdownScene struct {
	BaseScene
}

func (s *batch19DropdownScene) Title() string { return "Batch 19 · Dropdown" }

func (s *batch19DropdownScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5DropPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func dropCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func (s *batch19DropdownScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b19",
		"Widget Batch 19 · Dropdown",
		"Classic select field — click to expand, pick one option. Popup scrolls for long lists.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b19-grid", 12)

	pMain := ui.NewPanel("b19-main", "Status", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5DropPanel(pMain, 12, 12, 6, 5, 5)
	pMain.Gap = 10
	pMain.TitleHeight = 32
	statusOpts := []string{"Draft", "In review", "Approved", "Shipped", "Archived"}
	dd := ui.NewDropdown("b19-status", statusOpts, 0, 0, 0, 0, 40)
	statusLbl := ui.NewLabel("b19-status-lbl", "", 0, 0, 0, 0)
	statusLbl.SetStyle("form-value")
	dd.SelectedIndex.Subscribe(func() {
		i := dd.SelectedIndex.Get()
		if i >= 0 && i < len(statusOpts) {
			statusLbl.Text.Set("Selected: " + statusOpts[i])
		}
	})
	dd.SelectedIndex.Set(dd.SelectedIndex.Get())
	pMain.AddChild(dropCaption("b19-cap", "Single-select list — no type-to-filter (see ComboBox batch)."))
	pMain.AddChild(dd)
	pMain.AddChild(statusLbl)

	pLong := ui.NewPanel("b19-long", "Long list", 0, 0, 0, 0)
	pLong.AutoHeight = true
	setSpans5DropPanel(pLong, 12, 12, 6, 5, 5)
	pLong.Gap = 10
	pLong.TitleHeight = 32
	timezones := []string{
		"UTC−12:00", "UTC−11:00", "UTC−10:00", "UTC−09:00", "UTC−08:00",
		"UTC−07:00", "UTC−06:00", "UTC−05:00", "UTC−04:00", "UTC−03:00",
		"UTC−02:00", "UTC−01:00", "UTC±00:00", "UTC+01:00", "UTC+02:00",
		"UTC+03:00", "UTC+04:00", "UTC+05:00", "UTC+06:00", "UTC+07:00",
		"UTC+08:00", "UTC+09:00", "UTC+10:00", "UTC+11:00", "UTC+12:00",
	}
	tz := ui.NewDropdown("b19-tz", timezones, 12, 0, 0, 0, 40)
	pLong.AddChild(dropCaption("b19-long-cap", "Twenty-five options — popup scrolls inside the menu, not the page."))
	pLong.AddChild(tz)

	pDisabled := ui.NewPanel("b19-off", "Disabled", 0, 0, 0, 0)
	pDisabled.AutoHeight = true
	setSpans5DropPanel(pDisabled, 12, 12, 12, 12, 12)
	pDisabled.Gap = 10
	pDisabled.TitleHeight = 32
	off := ui.NewDropdown("b19-off-dd", []string{"Cannot open"}, 0, 0, 0, 0, 40)
	off.Disabled = true
	pDisabled.AddChild(off)

	grid.AddChild(pMain)
	grid.AddChild(pLong)
	grid.AddChild(pDisabled)
	page.Body.AddChild(grid)
}
