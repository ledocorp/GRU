//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch20RadioGroupScene{} }) }

type batch20RadioGroupScene struct {
	BaseScene
}

func (s *batch20RadioGroupScene) Title() string { return "Batch 20 · RadioGroup" }

func (s *batch20RadioGroupScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5RadioPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func radioCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func (s *batch20RadioGroupScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b20",
		"Widget Batch 20 · RadioGroup",
		"Mutually exclusive options — vertical column or horizontal row.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b20-grid", 12)

	pTheme := ui.NewPanel("b20-theme", "Appearance", 0, 0, 0, 0)
	pTheme.AutoHeight = true
	setSpans5RadioPanel(pTheme, 12, 12, 6, 5, 5)
	pTheme.Gap = 10
	pTheme.TitleHeight = 32
	themeOpts := []string{"System default", "Light", "Dark"}
	theme := ui.NewRadioGroup("b20-theme-rg", themeOpts, 0, 0, 0, 120)
	theme.Selected.Set(0)
	themeLbl := ui.NewLabel("b20-theme-lbl", "", 0, 0, 0, 0)
	themeLbl.SetStyle("form-value")
	theme.Selected.Subscribe(func() {
		i := theme.Selected.Get()
		if i >= 0 && i < len(themeOpts) {
			themeLbl.Text.Set("Theme: " + themeOpts[i])
		}
	})
	theme.Selected.Set(theme.Selected.Get())
	pTheme.AddChild(radioCaption("b20-cap", "Vertical stack — one choice active at a time."))
	pTheme.AddChild(theme)
	pTheme.AddChild(themeLbl)

	pAlign := ui.NewPanel("b20-align", "Horizontal", 0, 0, 0, 0)
	pAlign.AutoHeight = true
	setSpans5RadioPanel(pAlign, 12, 12, 6, 5, 5)
	pAlign.Gap = 10
	pAlign.TitleHeight = 32
	alignOpts := []string{"Left", "Center", "Right", "Justify"}
	align := ui.NewRadioGroup("b20-align-rg", alignOpts, 0, 0, 0, 48)
	align.Vertical = false
	align.Selected.Set(1)
	alignLbl := ui.NewLabel("b20-align-lbl", "", 0, 0, 0, 0)
	alignLbl.SetStyle("form-value")
	align.Selected.Subscribe(func() {
		alignLbl.Text.Set(fmt.Sprintf("Alignment index: %d", align.Selected.Get()))
	})
	pAlign.AddChild(align)
	pAlign.AddChild(alignLbl)

	pDisabled := ui.NewPanel("b20-disabled", "Disabled option", 0, 0, 0, 0)
	pDisabled.AutoHeight = true
	setSpans5RadioPanel(pDisabled, 12, 12, 12, 12, 12)
	pDisabled.Gap = 10
	pDisabled.TitleHeight = 32
	planOpts := []string{"Free", "Pro", "Enterprise"}
	plan := ui.NewRadioGroup("b20-plan-rg", planOpts, 0, 0, 0, 120)
	plan.Selected.Set(0)
	plan.Disabled = []bool{false, false, true}
	pDisabled.AddChild(radioCaption("b20-dis-cap", "Enterprise tier is greyed out and cannot be selected."))
	pDisabled.AddChild(plan)

	grid.AddChild(pTheme)
	grid.AddChild(pAlign)
	grid.AddChild(pDisabled)
	page.Body.AddChild(grid)
}
