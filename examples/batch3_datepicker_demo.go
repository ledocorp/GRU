//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"time"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch3DatePickerScene{} }) }

// batch3DatePickerScene demonstrates DatePicker and the global calendar overlay.
//
// Two grid panels:
//   - Schedule — primary picker, live value, scrollable body (popup is not clipped).
//   - Variants — format, placeholder, and narrow-width fields.
type batch3DatePickerScene struct {
	BaseScene
}

func (s *batch3DatePickerScene) Title() string { return "Batch 3 · DatePicker" }

func (s *batch3DatePickerScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5DatePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func dp3Caption(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-label",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func dp3Body(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-value",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func (s *batch3DatePickerScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "dp3",
		"Widget Batch 3 · DatePicker",
		"Compact field with a screen-space calendar popup (DatePickerMgr). Popup is never clipped by panel or page scroll.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("dp3-grid", 16)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel: Schedule (main)
	// ══════════════════════════════════════════════════════════════════════════
	pMain := ui.NewPanel("dp3-main", "Schedule", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5DatePanel(pMain, 12, 12, 6, 7, 7)
	pMain.Gap = 10
	pMain.TitleHeight = 32

	pMain.AddChild(dp3Caption("dp3-desc", "Pick a date — the calendar opens via DatePickerMgr above scroll clips."))

	mainDate := ui.NewSignal(time.Date(2026, 5, 15, 0, 0, 0, 0, time.Local))
	dpMain := ui.NewDatePicker("dp3-picker", mainDate, 0, 0, 0, 38)

	valueLbl := ui.NewLabel("dp3-value", "", 0, 0, 0, 0)
	valueLbl.SetStyle("form-value")
	valueLbl.Wrap = true
	refreshValue := func() {
		t := mainDate.Get()
		if t.IsZero() {
			valueLbl.Text.Set("Value: (none)")
		} else {
			valueLbl.Text.Set(fmt.Sprintf("Value: %s", t.In(time.Local).Format(time.RFC3339)))
		}
		valueLbl.MarkDirty()
	}
	mainDate.Subscribe(refreshValue)
	refreshValue()

	fillerVP := ui.NewViewport("dp3-filler-vp", 0, 0, 0, 220)
	fillerVP.Gap = 6
	fillerVP.SetStyle("list")
	fillerBox := ui.NewContainer("dp3-filler", 0, 0, 0, 0)
	fillerBox.FlexDirection = ui.FlexColumn
	fillerBox.Gap = 4
	fillerBox.SetStyle("transparent")
	for i := 0; i < 10; i++ {
		line := dp3Body(fmt.Sprintf("dp3-line-%d", i),
			fmt.Sprintf("Panel body line %d — scroll the page; open the picker to confirm the calendar draws on top.", i+1))
		fillerBox.AddChild(line)
	}
	fillerVP.AddChild(fillerBox)

	pMain.AddChild(dpMain)
	pMain.AddChild(valueLbl)
	pMain.AddChild(fillerVP)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel: Variants
	// ══════════════════════════════════════════════════════════════════════════
	pVariants := ui.NewPanel("dp3-variants", "Variants", 0, 0, 0, 0)
	pVariants.AutoHeight = true
	setSpans5DatePanel(pVariants, 12, 12, 6, 5, 5)
	pVariants.Gap = 10
	pVariants.TitleHeight = 32

	usDate := ui.NewSignal(time.Date(2026, 12, 25, 0, 0, 0, 0, time.Local))
	dpUS := ui.NewDatePicker("dp3-us", usDate, 0, 0, 0, 36)
	dpUS.DateFormat = "01/02/2006"
	v1Status := ui.NewLabel("dp3-v1-status", "", 0, 0, 0, 0)
	v1Status.SetStyle("form-value")
	v1Status.Wrap = true
	usDate.Subscribe(func() {
		t := usDate.Get()
		if t.IsZero() {
			v1Status.Text.Set("Displayed: (placeholder)")
		} else {
			v1Status.Text.Set(fmt.Sprintf("Displayed: %s", t.Format(dpUS.DateFormat)))
		}
		v1Status.MarkDirty()
	})

	longDate := ui.NewSignal(time.Date(2026, 7, 4, 0, 0, 0, 0, time.Local))
	dpLong := ui.NewDatePicker("dp3-long", longDate, 0, 0, 0, 36)
	dpLong.DateFormat = "Monday, January 2, 2006"
	v2Status := ui.NewLabel("dp3-v2-status", "", 0, 0, 0, 0)
	v2Status.SetStyle("form-value")
	v2Status.Wrap = true
	longDate.Subscribe(func() {
		t := longDate.Get()
		if t.IsZero() {
			v2Status.Text.Set("Displayed: (placeholder)")
		} else {
			v2Status.Text.Set(fmt.Sprintf("Displayed: %s", t.Format(dpLong.DateFormat)))
		}
		v2Status.MarkDirty()
	})

	emptyDate := ui.NewSignal(time.Time{})
	dpEmpty := ui.NewDatePicker("dp3-empty", emptyDate, 0, 0, 0, 36)
	dpEmpty.Placeholder = "Choose your start date…"

	narrowRow := ui.NewContainer("dp3-narrow-row", 0, 0, 0, 36)
	narrowRow.FlexDirection = ui.FlexRow
	narrowRow.SetStyle("transparent")
	narrowDate := ui.NewSignal(time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local))
	dpNarrow := ui.NewDatePicker("dp3-narrow", narrowDate, 0, 0, 200, 36)
	narrowRow.AddChild(dpNarrow)

	pVariants.AddChild(dp3Caption("dp3-v1-title", "US format · MM/DD/YYYY"))
	pVariants.AddChild(dpUS)
	pVariants.AddChild(v1Status)
	pVariants.AddChild(dp3Caption("dp3-v2-title", "Long format · weekday + month name"))
	pVariants.AddChild(dpLong)
	pVariants.AddChild(v2Status)
	pVariants.AddChild(dp3Caption("dp3-v3-title", "Empty value · custom placeholder"))
	pVariants.AddChild(dpEmpty)
	pVariants.AddChild(dp3Caption("dp3-v4-title", "Fixed 200 px wide · same keyboard shortcuts"))
	pVariants.AddChild(narrowRow)
	pVariants.AddChild(dp3Body("dp3-hint",
		"Arrows · day · PgUp/PgDn · month · Enter commit · Esc cancel · click outside cancels"))

	usDate.Set(usDate.Get())
	longDate.Set(longDate.Get())

	grid.AddChild(pMain)
	grid.AddChild(pVariants)
	page.Body.AddChild(grid)
}
