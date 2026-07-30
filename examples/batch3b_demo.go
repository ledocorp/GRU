//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch3bScene{} }) }

// batch3bScene demonstrates the Accordion widget (Batch 3b).
//
// Three panels in a responsive page grid (same col-spans as Batch 3 · Badge):
//
//   - "Standalone" — three stand-alone accordions (one starts expanded).
//   - "In a Panel" — accordions with mixed content.
//   - "Reactive" — programmatic Expanded control via Buttons.
type batch3bScene struct {
	BaseScene
}

func (s *batch3bScene) Title() string { return "Batch 3b · Accordion" }

func (s *batch3bScene) OnUpdate(_ *ui.Document, _ float32) {}

// setSpans5AccordionPanel lays out accordion demo panels full-width on XS–MD,
// three columns (4+4+4) on LG/XL — same breakpoints as Badge batch demos.
func setSpans5AccordionPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func b3bCaption(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-label",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func b3bBody(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{
		Text:    text,
		Variant: "form-value",
	}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

func (s *batch3bScene) Build(doc *ui.Document) {
	// ── Page shell → main viewport (see page_shell.go) ────────────────────────
	page := MountAppPage(doc, "b3b",
		"Widget Batch 3b · Accordion",
		"Expand/collapse with height tween. Uses the Batch 3 responsive grid (stacked xs–md, three columns lg+) instead of FlexRow+flex-grow, which overstretched AutoHeight panels and clipped accordions.")
	page.Body.Gap = 12

	// ── Responsive page grid — see batch3_demo.go pattern ──────────────────────
	grid := NewBatchPageGrid("b3b-grid", 12)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 1: Standalone Accordions
	// ══════════════════════════════════════════════════════════════════════════
	pStandalone := ui.NewPanel("p-b3b-solo", "Standalone", 0, 0, 0, 0)
	pStandalone.AutoHeight = true
	setSpans5AccordionPanel(pStandalone, 12, 12, 12, 4, 4)
	pStandalone.Gap = 8
	pStandalone.TitleHeight = 32

	pStandalone.AddChild(b3bCaption("solo-intro", "Click any header to expand or collapse."))

	acc1 := ui.NewAccordion("solo-acc1", "What is Gru?", 0, 0, 0, 0)
	acc1.AddChild(b3bBody("acc1-body",
		"Gru is a retained-mode UI library for raylib-go with widgets, signals, and themes."))
	pStandalone.AddChild(acc1)

	// Accordion 2: starts expanded.
	acc2 := ui.NewAccordion("solo-acc2", "Installation", 0, 0, 0, 0)
	acc2.AddChild(b3bBody("acc2-body",
		"1. go get github.com/gen2brain/raylib-go\n2. go get github.com/ledocorp/gru\n3. Import and use ui.NewAccordion(...)."))
	acc2.Expanded.Set(true)
	pStandalone.AddChild(acc2)

	// Accordion 3: mixed Badge + Label content.
	acc3 := ui.NewAccordion("solo-acc3", "Status Badges", 0, 0, 0, 0)
	badgeRow := ui.NewContainer("acc3-brow", 0, 0, 0, 22)
	badgeRow.FlexDirection = ui.FlexRow
	badgeRow.Gap = 8
	badgeRow.SetStyle("transparent")
	badgeRow.AddChild(ui.NewBadge("acc3-b1", "Stable", ui.BadgeSuccess, 0, 0, 0, 22))
	badgeRow.AddChild(ui.NewBadge("acc3-b2", "v1.0.0", ui.BadgePrimary, 0, 0, 0, 22))
	badgeRow.AddChild(ui.NewBadge("acc3-b3", "MIT", ui.BadgeDefault, 0, 0, 0, 22))
	acc3.AddChild(badgeRow)
	noteLbl := ui.NewLabel("acc3-note", "All release channels are healthy.", 0, 0, 0, 0)
	noteLbl.SetStyle("form-value")
	noteLbl.Align = ui.LabelAlignLeft
	noteLbl.Wrap = true
	acc3.AddChild(noteLbl)
	pStandalone.AddChild(acc3)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 2: Accordions inside a nested panel
	// ══════════════════════════════════════════════════════════════════════════
	pNested := ui.NewPanel("p-b3b-nested", "In a Panel", 0, 0, 0, 0)
	pNested.AutoHeight = true
	setSpans5AccordionPanel(pNested, 12, 12, 12, 4, 4)
	pNested.Gap = 8
	pNested.TitleHeight = 32

	pNested.AddChild(b3bCaption("nest-intro", "Accordions composed with other widgets."))

	// Separator then two accordions.
	pNested.AddChild(ui.NewSeparator("nest-sep1", "", 0, 0, 0, 3))

	accA := ui.NewAccordion("nest-accA", "Typography settings", 0, 0, 0, 0)
	fsRow := ui.NewContainer("accA-fs", 0, 0, 0, 22)
	fsRow.FlexDirection = ui.FlexRow
	fsRow.Gap = 8
	fsRow.SetStyle("transparent")
	fsLbl := ui.NewLabel("accA-fslbl", "Font size:", 0, 0, 80, 22)
	fsLbl.SetStyle("form-label")
	fsBadge := ui.NewBadge("accA-fsbadge", "17 px", ui.BadgeInfo, 0, 0, 0, 22)
	fsRow.AddChild(fsLbl)
	fsRow.AddChild(fsBadge)
	accA.AddChild(fsRow)
	boldRow := ui.NewContainer("accA-bold", 0, 0, 0, 22)
	boldRow.FlexDirection = ui.FlexRow
	boldRow.Gap = 8
	boldRow.SetStyle("transparent")
	boldLbl := ui.NewLabel("accA-boldlbl", "Bold headers:", 0, 0, 100, 22)
	boldLbl.SetStyle("form-label")
	boldBadge := ui.NewBadge("accA-boldbadge", "Enabled", ui.BadgeSuccess, 0, 0, 0, 22)
	boldRow.AddChild(boldLbl)
	boldRow.AddChild(boldBadge)
	accA.AddChild(boldRow)
	pNested.AddChild(accA)

	accB := ui.NewAccordion("nest-accB", "Color theme", 0, 0, 0, 0)
	accB.AddChild(b3bBody("accB-body",
		"Active theme: Gru Light. All styles editable via GetThemeStyle()."))
	themeChips := ui.NewContainer("accB-chips", 0, 0, 0, 22)
	themeChips.FlexDirection = ui.FlexRow
	themeChips.Gap = 6
	themeChips.SetStyle("transparent")
	themeChips.AddChild(ui.NewBadge("accB-c1", "Primary", ui.BadgePrimary, 0, 0, 0, 22))
	themeChips.AddChild(ui.NewBadge("accB-c2", "Success", ui.BadgeSuccess, 0, 0, 0, 22))
	themeChips.AddChild(ui.NewBadge("accB-c3", "Danger", ui.BadgeDanger, 0, 0, 0, 22))
	accB.AddChild(themeChips)
	pNested.AddChild(accB)

	accC := ui.NewAccordion("nest-accC", "Layout engine", 0, 0, 0, 0)
	accC.AddChild(b3bBody("accC-body",
		"FlexRow / FlexColumn containers. Gap, Padding, and SetFlexGrow(n) supported."))
	accC.Expanded.Set(true)
	pNested.AddChild(accC)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 3: Reactive control via Expanded signal
	// ══════════════════════════════════════════════════════════════════════════
	pReactive := ui.NewPanel("p-b3b-reactive", "Reactive", 0, 0, 0, 0)
	pReactive.AutoHeight = true
	setSpans5AccordionPanel(pReactive, 12, 12, 12, 4, 4)
	pReactive.Gap = 8
	pReactive.TitleHeight = 32

	pReactive.AddChild(b3bCaption("react-intro", "Buttons drive Expanded.Set() directly."))

	stateLbl := ui.NewLabel("react-state", "State: collapsed", 0, 0, 0, 0)
	stateLbl.SetStyle("form-value")
	stateLbl.Align = ui.LabelAlignLeft
	pReactive.AddChild(stateLbl)

	accR := ui.NewAccordion("react-acc", "Controlled Section", 0, 0, 0, 0)
	accR.AddChild(b3bBody("react-body",
		"This accordion is driven entirely by the buttons below — no header click needed."))
	rCountBadge := ui.NewBadge("react-count", "toggles: 0", ui.BadgePrimary, 0, 0, 0, 22)
	accR.AddChild(rCountBadge)
	pReactive.AddChild(accR)

	// Toggle counter tracked as a closure variable.
	toggleCount := 0
	accR.OnToggle = func(expanded bool) {
		toggleCount++
		rCountBadge.Text.Set(fmt.Sprintf("toggles: %d", toggleCount))
		if expanded {
			stateLbl.Text.Set("State: expanded")
		} else {
			stateLbl.Text.Set("State: collapsed")
		}
	}

	// Button row (tight under accordion — no separator gap).
	btnExpand := ui.NewButton("react-expand", "Expand", 0, 0, 96, 34)
	btnExpand.SetStyle("primary")
	btnExpand.OnClick = func() {
		if !accR.Expanded.Get() {
			accR.Expanded.Set(true)
		}
	}

	btnCollapse := ui.NewButton("react-collapse", "Collapse", 0, 0, 96, 34)
	btnCollapse.SetStyle("button")
	btnCollapse.OnClick = func() {
		if accR.Expanded.Get() {
			accR.Expanded.Set(false)
		}
	}

	btnToggle := ui.NewButton("react-toggle", "Toggle", 0, 0, 80, 34)
	btnToggle.SetStyle("default")
	btnToggle.OnClick = func() { accR.Toggle() }

	btnRow := ui.NewContainer("react-btnrow", 0, 0, 0, 34)
	btnRow.FlexDirection = ui.FlexRow
	btnRow.Gap = 8
	btnRow.SetStyle("transparent")
	btnRow.AddChild(btnExpand)
	btnRow.AddChild(btnCollapse)
	btnRow.AddChild(btnToggle)
	pReactive.AddChild(btnRow)

	// Additional informational separator + note.
	pReactive.AddChild(ui.NewSeparator("react-sep2", "", 0, 0, 0, 8))
	pReactive.AddChild(b3bBody("react-note",
		"Subscribe to Expanded for reactive updates:\nacc.Expanded.Subscribe(func(){ ... })"))

	// ── Assemble ───────────────────────────────────────────────────────────────
	grid.AddChild(pStandalone)
	grid.AddChild(pNested)
	grid.AddChild(pReactive)

	page.Body.AddChild(grid)

}
