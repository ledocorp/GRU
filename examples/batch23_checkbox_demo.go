//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch23CheckboxScene{} }) }

// batch23CheckboxScene demonstrates Checkbox with labelled rows.
// Recipe: CP-SHELL-PAGE + NewBatchPageGrid.
type batch23CheckboxScene struct {
	BaseScene
}

func (s *batch23CheckboxScene) Title() string { return "Batch 23 · Checkbox" }

func (s *batch23CheckboxScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5CheckPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func checkRow(id, label string, box *ui.Checkbox) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 10
	row.AutoHeight = true
	row.SetStyle("transparent")
	row.AddChild(box)
	lbl := ui.NewPlainText(id+"-lbl", "form-value", label, 0, 0, 0, 0)
	lbl.SetFlexGrow(1)
	row.AddChild(lbl)
	return row
}

func (s *batch23CheckboxScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b23",
		"Checkbox",
		"Boolean checkbox with checkmark — pair with a label in a row.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b23-grid", 12)

	pForm := ui.NewPanel("b23-form", "Preferences", 0, 0, 0, 0)
	pForm.AutoHeight = true
	setSpans5CheckPanel(pForm, 12, 12, 6, 5, 5)
	pForm.Gap = 8
	pForm.TitleHeight = 32

	email := ui.NewCheckbox("b23-email", true, 0, 0, 24, 24)
	marketing := ui.NewCheckbox("b23-mkt", false, 0, 0, 24, 24)
	summary, summaryDisplay := FlexCopyPair("b23-sum", "form-value", "")
	refresh := func() {
		summaryDisplay.Set(fmt.Sprintf("Email alerts: %v · Marketing: %v", email.Value.Get(), marketing.Value.Get()))
	}
	email.Value.Subscribe(refresh)
	marketing.Value.Subscribe(refresh)
	refresh()

	pForm.AddChild(batchCaption("b23-cap", "Click the box — standard size is 24×24 px."))
	pForm.AddChild(checkRow("b23-email-row", "Send email notifications", email))
	pForm.AddChild(checkRow("b23-mkt-row", "Receive product updates", marketing))
	pForm.AddChild(summary)

	pTerms := ui.NewPanel("b23-terms", "Agreement", 0, 0, 0, 0)
	pTerms.AutoHeight = true
	setSpans5CheckPanel(pTerms, 12, 12, 6, 5, 5)
	pTerms.Gap = 8
	pTerms.TitleHeight = 32
	terms := ui.NewCheckbox("b23-terms-box", false, 0, 0, 24, 24)
	termsStatus, termsStatusDisplay := FlexCopyPair("b23-terms-status", "form-value", "")
	submit := ui.NewButton("b23-submit", "Continue", 0, 0, 120, 36)
	submit.SetStyle("button")
	canSubmit := false
	terms.Value.Subscribe(func() {
		canSubmit = terms.Value.Get()
	})
	submit.OnClick = func() {
		if canSubmit {
			termsStatusDisplay.Set("Continued — terms accepted.")
		} else {
			termsStatusDisplay.Set("Accept the terms to continue.")
		}
	}
	pTerms.AddChild(checkRow("b23-terms-row", "I agree to the terms of service", terms))
	pTerms.AddChild(batchCaption("b23-terms-hint", "Continue only works after the box is checked."))
	pTerms.AddChild(submit)
	pTerms.AddChild(termsStatus)

	pDisabled := ui.NewPanel("b23-off", "Disabled", 0, 0, 0, 0)
	pDisabled.AutoHeight = true
	setSpans5CheckPanel(pDisabled, 12, 12, 12, 12, 12)
	pDisabled.Gap = 8
	pDisabled.TitleHeight = 32
	off := ui.NewCheckbox("b23-off", true, 0, 0, 24, 24)
	off.Disabled = true
	pDisabled.AddChild(checkRow("b23-off-row", "Locked on (disabled)", off))

	grid.AddChild(pForm)
	grid.AddChild(pTerms)
	grid.AddChild(pDisabled)
	page.Body.AddChild(grid)
}
