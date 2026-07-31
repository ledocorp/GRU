//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch15ComboBoxScene{} }) }

type batch15ComboBoxScene struct {
	BaseScene
}

func (s *batch15ComboBoxScene) Title() string { return "Batch 15 · ComboBox" }

func (s *batch15ComboBoxScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5ComboPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func comboCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

var comboCountries = []string{
	"Argentina", "Australia", "Brazil", "Canada", "Chile", "China", "Colombia",
	"Denmark", "Egypt", "Finland", "France", "Germany", "Greece", "India",
	"Indonesia", "Ireland", "Italy", "Japan", "Kenya", "Mexico", "Netherlands",
	"New Zealand", "Nigeria", "Norway", "Poland", "Portugal", "Singapore",
	"South Africa", "South Korea", "Spain", "Sweden", "Switzerland", "Turkey",
	"United Kingdom", "United States", "Vietnam",
}

func (s *batch15ComboBoxScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b15",
		"Widget Batch 15 · ComboBox",
		"Searchable dropdown — type to filter, then pick an option.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b15-grid", 12)

	pMain := ui.NewPanel("b15-main", "Country", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5ComboPanel(pMain, 12, 12, 6, 7, 7)
	pMain.Gap = 10
	pMain.TitleHeight = 32
	country := ui.NewSignal("United States")
	cb := ui.NewComboBox("b15-country", comboCountries, country, 0, 0, 0, 40)
	status := ui.NewLabel("b15-status", "", 0, 0, 0, 0)
	status.SetStyle("form-value")
	status.Wrap = true
	country.Subscribe(func() {
		status.Text.Set("Selected: " + country.Get())
	})
	country.Set(country.Get())
	pMain.AddChild(comboCaption("b15-cap", "Open the list and type to narrow options (e.g. \"united\")."))
	pMain.AddChild(cb)
	pMain.AddChild(status)

	pVariants := ui.NewPanel("b15-variants", "Variants", 0, 0, 0, 0)
	pVariants.AutoHeight = true
	setSpans5ComboPanel(pVariants, 12, 12, 6, 5, 5)
	pVariants.Gap = 10
	pVariants.TitleHeight = 32

	lang := ui.NewSignal("Go")
	langs := []string{"Go", "Rust", "TypeScript", "Python", "C#", "Java", "Kotlin", "Swift"}
	cbLang := ui.NewComboBox("b15-lang", langs, lang, 0, 0, 0, 36)

	empty := ui.NewSignal("")
	cbEmpty := ui.NewComboBox("b15-empty", []string{"Alpha", "Beta", "Gamma"}, empty, 0, 0, 0, 36)
	cbEmpty.Placeholder = "Pick a letter…"

	pVariants.AddChild(comboCaption("b15-v1", "Short list — already has a selection (Go)."))
	pVariants.AddChild(cbLang)
	pVariants.AddChild(comboCaption("b15-v2", "Placeholder when unset — starts empty; type or pick."))
	pVariants.AddChild(cbEmpty)

	pScroll := ui.NewPanel("b15-scroll", "Scroll test", 0, 0, 0, 0)
	pScroll.AutoHeight = true
	setSpans5ComboPanel(pScroll, 12, 12, 12, 12, 12)
	pScroll.Gap = 10
	pScroll.TitleHeight = 32

	region := ui.NewSignal(comboCountries[0])
	cbRegion := ui.NewComboBox("b15-region", comboCountries, region, 0, 0, 0, 40)

	lines := ui.NewContainer("b15-lines", 0, 0, 0, 0)
	lines.FlexDirection = ui.FlexColumn
	lines.Gap = 6
	lines.SetStyle("transparent")
	for i := 0; i < 8; i++ {
		lbl := ui.NewLabel(fmt.Sprintf("b15-line-%d", i),
			fmt.Sprintf("Body line %d — scroll the page with a calendar open; comboboxes below should not react.", i+1),
			0, 0, 0, 0)
		lbl.SetStyle("form-value")
		lbl.Wrap = true
		lines.AddChild(lbl)
	}

	pScroll.AddChild(comboCaption("b15-scroll-cap",
		"Long page body below the picker — verifies calendar overlay blocks wheel/hover on fields underneath."))
	pScroll.AddChild(cbRegion)
	pScroll.AddChild(lines)

	grid.AddChild(pMain)
	grid.AddChild(pVariants)
	grid.AddChild(pScroll)
	page.Body.AddChild(grid)
}
