//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &batch18ColorWellScene{} }) }

type batch18ColorWellScene struct {
	BaseScene
}

func (s *batch18ColorWellScene) Title() string { return "Batch 18 · ColorWell" }

func (s *batch18ColorWellScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5WellPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func wellCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

var themeSwatches = []rl.Color{
	rl.NewColor(79, 70, 229, 255),
	rl.NewColor(59, 130, 246, 255),
	rl.NewColor(34, 197, 94, 255),
	rl.NewColor(234, 179, 8, 255),
	rl.NewColor(249, 115, 22, 255),
	rl.NewColor(239, 68, 68, 255),
	rl.NewColor(139, 92, 246, 255),
	rl.NewColor(100, 116, 139, 255),
}

func colorHex(c rl.Color) string {
	return fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B)
}

func (s *batch18ColorWellScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b18",
		"Widget Batch 18 · ColorWell",
		"Preset swatch row — no HSV popup. Full picker → Batch 26 · ColorPicker.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b18-grid", 12)

	pAccent := ui.NewPanel("b18-accent", "Theme accent", 0, 0, 0, 0)
	pAccent.AutoHeight = true
	setSpans5WellPanel(pAccent, 12, 12, 6, 5, 5)
	pAccent.Gap = 10
	pAccent.TitleHeight = 32
	accent := ui.NewSignal(themeSwatches[0])
	well := ui.NewColorWell("b18-accent-well", accent.Get(), themeSwatches, 0, 0, 0, 0)
	accentLbl := ui.NewLabel("b18-accent-lbl", "", 0, 0, 0, 0)
	accentLbl.SetStyle("form-value")
	accent.Subscribe(func() {
		accentLbl.Text.Set("Accent: " + colorHex(accent.Get()))
	})
	accent.Set(accent.Get())
	pAccent.AddChild(wellCaption("b18-accent-cap", "Click a swatch — selected ring shows the active color."))
	pAccent.AddChild(well)
	pAccent.AddChild(accentLbl)

	pTag := ui.NewPanel("b18-tag", "Tag colors", 0, 0, 0, 0)
	pTag.AutoHeight = true
	setSpans5WellPanel(pTag, 12, 12, 6, 5, 5)
	pTag.Gap = 10
	pTag.TitleHeight = 32
	tagSwatches := []rl.Color{
		rl.NewColor(219, 234, 254, 255),
		rl.NewColor(220, 252, 231, 255),
		rl.NewColor(254, 249, 195, 255),
		rl.NewColor(255, 228, 230, 255),
	}
	tag := ui.NewSignal(tagSwatches[0])
	tagWell := ui.NewColorWell("b18-tag-well", tag.Get(), tagSwatches, 0, 0, 0, 0)
	tagLbl := ui.NewLabel("b18-tag-lbl", "Pastel tag palette", 0, 0, 0, 0)
	tagLbl.SetStyle("form-value")
	tag.Subscribe(func() {
		tagLbl.Text.Set("Tag fill: " + colorHex(tag.Get()))
	})
	pTag.AddChild(tagWell)
	pTag.AddChild(tagLbl)

	pRow := ui.NewPanel("b18-row", "Form row", 0, 0, 0, 0)
	pRow.AutoHeight = true
	setSpans5WellPanel(pRow, 12, 12, 12, 12, 12)
	pRow.Gap = 10
	pRow.TitleHeight = 32
	row := ui.NewContainer("b18-form-row", 0, 0, 0, 0)
	row.FlexDirection = ui.FlexRow
	row.Gap = 16
	row.SetStyle("transparent")
	lbl := ui.NewLabel("b18-label", "Highlight", 0, 0, 120, 36)
	lbl.SetStyle("form-label")
	highlight := ui.NewSignal(themeSwatches[2])
	hWell := ui.NewColorWell("b18-highlight", highlight.Get(), themeSwatches[:5], 0, 0, 0, 0)
	row.AddChild(lbl)
	row.AddChild(hWell)
	pRow.AddChild(wellCaption("b18-row-cap", "Typical settings form pairing — label + swatch row."))
	pRow.AddChild(row)

	grid.AddChild(pAccent)
	grid.AddChild(pTag)
	grid.AddChild(pRow)
	page.Body.AddChild(grid)
}
