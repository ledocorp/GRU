//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"time"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &filtersScene{} }) }

// filtersScene demonstrates ComboBox, DateRangePicker, and ColorWell.
// Recipe: CP-SHELL-PAGE + Panel hosts (early widget demo polish).
type filtersScene struct {
	BaseScene
	country *ui.Signal[string]
	start   *ui.Signal[time.Time]
	end     *ui.Signal[time.Time]
	accent  *ui.Signal[rl.Color]
}

func (s *filtersScene) Title() string { return "Filters (Go)" }

func (s *filtersScene) Build(doc *ui.Document) {
	countries := []string{
		"United States", "United Kingdom", "Canada", "Germany", "France",
		"Japan", "Australia", "Brazil", "India", "Mexico", "South Korea", "Italy",
	}
	s.country = ui.NewSignal("United States")
	now := time.Now()
	s.start = ui.NewSignal(dateOnly(now.AddDate(0, 0, -14)))
	s.end = ui.NewSignal(dateOnly(now))
	s.accent = ui.NewSignal(rl.NewColor(79, 70, 229, 255))

	PreloadScenePhosphor(doc)
	page := MountAppPage(doc, "filters",
		"Filters",
		"ComboBox, DateRangePicker, and ColorWell — searchable country, range, accent.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("filters-grid", 12)

	main := ui.NewPanel("filters-main", "Report filters", 0, 0, 0, 0)
	main.AutoHeight = true
	main.Gap = 12
	main.TitleHeight = 32
	main.SetColSpan(ui.BreakpointXS, 12)
	main.SetColSpan(ui.BreakpointSM, 12)
	main.SetColSpan(ui.BreakpointMD, 12)
	main.SetColSpan(ui.BreakpointLG, 8)
	main.SetColSpan(ui.BreakpointXL, 8)

	main.AddChild(batchCaption("filters-hint",
		"Pick a country, date range, and accent. Summary updates from shared signals."))

	countryRow := filterFieldRow("filters-country-row", "Country", 88)
	countryCB := ui.NewComboBox("filters-country", countries, s.country, 0, 0, 0, 40)
	countryCB.SetFlexGrow(1)
	countryRow.AddChild(countryCB)
	main.AddChild(countryRow)

	rangeRow := filterFieldRow("filters-range-row", "Date range", 88)
	rangePick := ui.NewDateRangePicker("filters-range", s.start, s.end, 0, 0, 0, 40)
	rangePick.SetFlexGrow(1)
	rangeRow.AddChild(rangePick)
	main.AddChild(rangeRow)

	accentRow := filterFieldRow("filters-accent-row", "Accent", 88)
	accentWell := ui.NewColorWell("filters-accent", s.accent.Get(), []rl.Color{
		rl.NewColor(79, 70, 229, 255),
		rl.NewColor(220, 38, 38, 255),
		rl.NewColor(234, 88, 12, 255),
		rl.NewColor(22, 163, 74, 255),
		rl.NewColor(37, 99, 235, 255),
		rl.NewColor(124, 58, 237, 255),
	}, 0, 0, 0, 0)
	accentWell.Value = s.accent
	accentRow.AddChild(accentWell)
	main.AddChild(accentRow)

	summary, summaryDisplay := FlexCopyPair("filters-summary", "form-value", "")
	refresh := func() {
		st, en := s.start.Get(), s.end.Get()
		st, en = normalizeDemoRange(st, en)
		ac := s.accent.Get()
		summaryDisplay.Set(fmt.Sprintf("%s · %s – %s · accent #%02X%02X%02X",
			s.country.Get(),
			st.Format("2006-01-02"),
			en.Format("2006-01-02"),
			ac.R, ac.G, ac.B,
		))
	}
	s.country.Subscribe(refresh)
	s.start.Subscribe(refresh)
	s.end.Subscribe(refresh)
	s.accent.Subscribe(refresh)
	refresh()
	main.AddChild(summary)

	side := ui.NewPanel("filters-side", "Tips", 0, 0, 0, 0)
	side.AutoHeight = true
	side.Gap = 10
	side.TitleHeight = 32
	side.SetColSpan(ui.BreakpointXS, 12)
	side.SetColSpan(ui.BreakpointSM, 12)
	side.SetColSpan(ui.BreakpointMD, 12)
	side.SetColSpan(ui.BreakpointLG, 4)
	side.SetColSpan(ui.BreakpointXL, 4)
	side.AddChild(batchCaption("filters-side-hint",
		"ComboBox filters a long option list as you type. DateRangePicker keeps start/end on one calendar."))

	grid.AddChild(main)
	grid.AddChild(side)
	page.Body.AddChild(grid)
}

func filterFieldRow(id, label string, labelW float32) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 12
	row.AutoHeight = true
	row.SetStyle("transparent")
	lbl := ui.NewPlainText(id+"-lbl", "form-label", label, 0, 0, 0, 0)
	lbl.PreferredWidth = labelW
	row.AddChild(lbl)
	return row
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func normalizeDemoRange(a, b time.Time) (time.Time, time.Time) {
	a, b = dateOnly(a), dateOnly(b)
	if b.Before(a) {
		a, b = b, a
	}
	return a, b
}
