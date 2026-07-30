//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch4Scene{} }) }

// batch4Scene demonstrates the Stepper / ProgressIndicator widget (Batch 4).
//
// Three responsive panels:
//
//   - "Horizontal" — 4-step wizard with Next / Back buttons.
//   - "Vertical"   — 5-step pipeline; steps are click-to-jump.
//   - "Reactive"   — stepper driven entirely by signal writes.
type batch4Scene struct {
	BaseScene
}

func (s *batch4Scene) Title() string { return "Batch 4 · Stepper" }

func (s *batch4Scene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5StepperPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func b4Caption(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-label",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func b4Status(id, text string) (*ui.RichText, *ui.Signal[string]) {
	return FlexCopyPair(id, "form-value", text)
}

func (s *batch4Scene) Build(doc *ui.Document) {
	// ── Page shell → main viewport (see page_shell.go) ────────────────────────
	page := MountAppPage(doc, "b4",
		"Widget Batch 4 · Stepper",
		"Responsive step progress indicators using intrinsic-height panels in the Batch page grid.")
	page.Body.Gap = 12

	grid := NewBatchPageGrid("b4-grid", 12)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 1: Horizontal Stepper  (Next / Back buttons)
	// ══════════════════════════════════════════════════════════════════════════
	pHoriz := ui.NewPanel("p-b4-horiz", "Horizontal", 0, 0, 0, 0)
	pHoriz.AutoHeight = true
	setSpans5StepperPanel(pHoriz, 12, 12, 12, 4, 4)
	pHoriz.Gap = 10
	pHoriz.TitleHeight = 32

	horizSteps := []ui.StepItem{
		{Title: "Account", Subtitle: "Create login"},
		{Title: "Profile", Subtitle: "Your details"},
		{Title: "Review", Subtitle: "Check info"},
		{Title: "Confirm", Subtitle: "Submit"},
	}
	horizStepper := ui.NewStepper("h-stepper", horizSteps, 0, 0, 0, 96)

	horizStatus, horizStatusText := b4Status("h-status", "Step 1 of 4 - Account")
	horizDesc := b4Caption("h-desc", "Create your login credentials to get started.")

	stepDescs := []string{
		"Create your login credentials to get started.",
		"Fill in your personal profile information.",
		"Review all the information you have entered.",
		"Submit your details to complete registration.",
	}

	// Reactive description + status update.
	ui.NewEffect(func() {
		cur := horizStepper.CurrentStep.Get()
		if cur < 0 || cur >= len(horizSteps) {
			return
		}
		horizStatusText.Set(fmt.Sprintf("Step %d of %d - %s", cur+1, len(horizSteps), horizSteps[cur].Title))
		horizDesc.SetSpans([]ui.TextSpan{{
			Text:    stepDescs[cur],
			Variant: "form-label",
		}})
		horizDesc.MarkDirty()
	})

	btnRow := ui.NewContainer("h-btn-row", 0, 0, 0, 34)
	btnRow.FlexDirection = ui.FlexRow
	btnRow.Gap = 8
	btnRow.SetStyle("transparent")

	prevBtn := ui.NewButton("h-prev", "Back", 0, 0, 88, 34)
	prevBtn.SetStyle("button")
	prevBtn.OnClick = func() {
		cur := horizStepper.CurrentStep.Get()
		if cur > 0 {
			horizStepper.CurrentStep.Set(cur - 1)
		}
	}

	nextBtn := ui.NewButton("h-next", "Next", 0, 0, 88, 34)
	nextBtn.SetStyle("primary")
	nextBtn.OnClick = func() {
		cur := horizStepper.CurrentStep.Get()
		if cur < len(horizSteps)-1 {
			horizStepper.CurrentStep.Set(cur + 1)
		}
	}

	resetBtn := ui.NewButton("h-reset", "Reset", 0, 0, 74, 34)
	resetBtn.SetStyle("button")
	resetBtn.OnClick = func() { horizStepper.CurrentStep.Set(0) }

	btnRow.AddChild(prevBtn)
	btnRow.AddChild(nextBtn)
	btnRow.AddChild(resetBtn)

	// Separator + clickable variant below.
	sep1 := ui.NewSeparator("h-sep1", "Clickable variant", 0, 0, 0, 18)

	clickSteps := []ui.StepItem{
		{Title: "Init"},
		{Title: "Build"},
		{Title: "Test"},
		{Title: "Deploy"},
	}
	clickStepper := ui.NewStepper("h-click", clickSteps, 0, 0, 0, 72)
	clickStepper.Clickable = true
	clickStepper.CircleRadius = 14

	clickNote, _ := b4Status("h-click-note", "Click a step circle to jump directly.")
	clickStatus, clickStatusText := b4Status("h-click-status", "Step 1 - Init")
	clickStepper.OnStepClick = func(idx int) {
		clickStatusText.Set(fmt.Sprintf("Step %d - %s", idx+1, clickSteps[idx].Title))
	}

	pHoriz.AddChild(horizStepper)
	pHoriz.AddChild(horizStatus)
	pHoriz.AddChild(horizDesc)
	pHoriz.AddChild(btnRow)
	pHoriz.AddChild(sep1)
	pHoriz.AddChild(clickStepper)
	pHoriz.AddChild(clickNote)
	pHoriz.AddChild(clickStatus)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 2: Vertical Stepper  (click-to-jump, 5 steps)
	// ══════════════════════════════════════════════════════════════════════════
	pVert := ui.NewPanel("p-b4-vert", "Vertical (Clickable)", 0, 0, 0, 0)
	pVert.AutoHeight = true
	setSpans5StepperPanel(pVert, 12, 12, 12, 4, 4)
	pVert.Gap = 10
	pVert.TitleHeight = 32

	vertSteps := []ui.StepItem{
		{Title: "Initialize", Subtitle: "Set up workspace"},
		{Title: "Configure", Subtitle: "Adjust settings"},
		{Title: "Build", Subtitle: "Compile assets"},
		{Title: "Test", Subtitle: "Run test suite"},
		{Title: "Deploy", Subtitle: "Push to production"},
	}
	vertStepper := ui.NewStepper("v-stepper", vertSteps, 0, 0, 0, 320)
	vertStepper.Direction = ui.StepperVertical
	vertStepper.Clickable = true

	vertNote, _ := b4Status("v-note", "Click a step to navigate to it.")
	vertStatus, vertStatusText := b4Status("v-status", "Step 1 of 5 - Initialize")

	ui.NewEffect(func() {
		cur := vertStepper.CurrentStep.Get()
		if cur < 0 || cur >= len(vertSteps) {
			return
		}
		vertStatusText.Set(fmt.Sprintf("Step %d of %d - %s", cur+1, len(vertSteps), vertSteps[cur].Title))
	})

	vBtnRow := ui.NewContainer("v-btn-row", 0, 0, 0, 34)
	vBtnRow.FlexDirection = ui.FlexRow
	vBtnRow.Gap = 8
	vBtnRow.SetStyle("transparent")

	vPrevBtn := ui.NewButton("v-prev", "Back", 0, 0, 88, 34)
	vPrevBtn.SetStyle("button")
	vPrevBtn.OnClick = func() {
		cur := vertStepper.CurrentStep.Get()
		if cur > 0 {
			vertStepper.CurrentStep.Set(cur - 1)
		}
	}

	vNextBtn := ui.NewButton("v-next", "Next", 0, 0, 88, 34)
	vNextBtn.SetStyle("primary")
	vNextBtn.OnClick = func() {
		cur := vertStepper.CurrentStep.Get()
		if cur < len(vertSteps)-1 {
			vertStepper.CurrentStep.Set(cur + 1)
		}
	}

	vBtnRow.AddChild(vPrevBtn)
	vBtnRow.AddChild(vNextBtn)

	pVert.AddChild(vertStepper)
	pVert.AddChild(vertNote)
	pVert.AddChild(vertStatus)
	pVert.AddChild(vBtnRow)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 3: Reactive — signal-driven step control
	// ══════════════════════════════════════════════════════════════════════════
	pReactive := ui.NewPanel("p-b4-reactive", "Reactive", 0, 0, 0, 0)
	pReactive.AutoHeight = true
	setSpans5StepperPanel(pReactive, 12, 12, 12, 4, 4)
	pReactive.Gap = 10
	pReactive.TitleHeight = 32

	reactSteps := []ui.StepItem{
		{Title: "Plan", Subtitle: "Define scope"},
		{Title: "Execute", Subtitle: "Do the work"},
		{Title: "Review", Subtitle: "Inspect output"},
		{Title: "Close", Subtitle: "Archive and ship"},
	}
	reactStepper := ui.NewStepper("r-stepper", reactSteps, 0, 0, 0, 96)
	reactStepper.Clickable = true

	rStatusLbl, rStatusText := b4Status("r-status", "")
	rSubLbl := b4Caption("r-sub", "")

	reactDescs := []string{
		"Define project scope, milestones, and deliverables.",
		"Carry out the planned work and track progress.",
		"Inspect outputs for quality and completeness.",
		"Archive results, write docs, and ship the release.",
	}

	ui.NewEffect(func() {
		cur := reactStepper.CurrentStep.Get()
		if cur < 0 || cur >= len(reactSteps) {
			return
		}
		rStatusText.Set(fmt.Sprintf("%s  [%d/%d]", reactSteps[cur].Title, cur+1, len(reactSteps)))
		rSubLbl.SetSpans([]ui.TextSpan{{
			Text:    reactDescs[cur],
			Variant: "form-value",
		}})
		rSubLbl.MarkDirty()
	})

	// Separator + quick-jump buttons.
	rSep := ui.NewSeparator("r-sep", "Jump to step", 0, 0, 0, 18)

	rBtnRow := ui.NewContainer("r-btn-row", 0, 0, 0, 34)
	rBtnRow.FlexDirection = ui.FlexRow
	rBtnRow.Gap = 6
	rBtnRow.SetStyle("transparent")

	stepLabels := []string{"Plan", "Execute", "Review", "Close"}
	stepStyles := []string{"button", "primary", "primary", "button"}
	for i, lbl := range stepLabels {
		idx := i // capture
		style := stepStyles[i]
		btn := ui.NewButton(fmt.Sprintf("r-step-%d", i), lbl, 0, 0, 74, 34)
		btn.SetStyle(style)
		btn.OnClick = func() { reactStepper.CurrentStep.Set(idx) }
		rBtnRow.AddChild(btn)
	}

	rCtrlRow := ui.NewContainer("r-ctrl-row", 0, 0, 0, 34)
	rCtrlRow.FlexDirection = ui.FlexRow
	rCtrlRow.Gap = 8
	rCtrlRow.SetStyle("transparent")

	completeBtn := ui.NewButton("r-complete", "Complete", 0, 0, 100, 34)
	completeBtn.SetStyle("primary")
	completeBtn.OnClick = func() { reactStepper.CurrentStep.Set(len(reactSteps) - 1) }

	rResetBtn := ui.NewButton("r-reset", "Reset", 0, 0, 74, 34)
	rResetBtn.SetStyle("button")
	rResetBtn.OnClick = func() { reactStepper.CurrentStep.Set(0) }

	rCtrlRow.AddChild(completeBtn)
	rCtrlRow.AddChild(rResetBtn)

	// Variant: large-radius stepper (visual comparison).
	rSep2 := ui.NewSeparator("r-sep2", "Compact phases", 0, 0, 0, 18)

	bigSteps := []ui.StepItem{
		{Title: "Alpha"},
		{Title: "Beta"},
		{Title: "GA"},
	}
	bigStepper := ui.NewStepper("r-big", bigSteps, 0, 0, 0, 76)
	bigStepper.CircleRadius = 16
	bigStepper.Clickable = true

	rBigStatus, rBigStatusText := b4Status("r-big-status", "Phase: Alpha")
	bigStepper.OnStepClick = func(idx int) {
		rBigStatusText.Set(fmt.Sprintf("Phase: %s", bigSteps[idx].Title))
	}

	pReactive.AddChild(reactStepper)
	pReactive.AddChild(rStatusLbl)
	pReactive.AddChild(rSubLbl)
	pReactive.AddChild(rSep)
	pReactive.AddChild(rBtnRow)
	pReactive.AddChild(rCtrlRow)
	pReactive.AddChild(rSep2)
	pReactive.AddChild(bigStepper)
	pReactive.AddChild(rBigStatus)

	// ── Assemble ──────────────────────────────────────────────────────────────
	grid.AddChild(pHoriz)
	grid.AddChild(pVert)
	grid.AddChild(pReactive)

	page.Body.AddChild(grid)
}
