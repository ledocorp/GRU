//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch22ToggleScene{} }) }

type batch22ToggleScene struct {
	BaseScene
}

func (s *batch22ToggleScene) Title() string { return "Batch 22 · Toggle" }

func (s *batch22ToggleScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5TogglePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func toggleCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func toggleRow(id, label string, tog *ui.Toggle) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	row.FlexDirection = ui.FlexRow
	row.Gap = 12
	row.SetStyle("transparent")
	lbl := ui.NewLabel(id+"-lbl", label, 0, 0, 0, 0)
	lbl.SetStyle("form-label")
	lbl.SetFlexGrow(1)
	tog.SetFlexGrow(0)
	row.AddChild(lbl)
	row.AddChild(tog)
	return row
}

func (s *batch22ToggleScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b22",
		"Widget Batch 22 · Toggle",
		"Pill track switch — animated thumb, recommended 52×28 px.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b22-grid", 12)

	pMain := ui.NewPanel("b22-main", "Settings", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5TogglePanel(pMain, 12, 12, 6, 5, 5)
	pMain.Gap = 10
	pMain.TitleHeight = 32

	wifi := ui.NewToggle("b22-wifi", true, 0, 0, 52, 28)
	dark := ui.NewToggle("b22-dark", false, 0, 0, 52, 28)
	status := ui.NewLabel("b22-status", "", 0, 0, 0, 0)
	status.SetStyle("form-value")
	status.Wrap = true
	refresh := func() {
		status.Text.Set(fmt.Sprintf("Wi-Fi %v · Dark mode %v", wifi.Value.Get(), dark.Value.Get()))
	}
	wifi.Value.Subscribe(refresh)
	dark.Value.Subscribe(refresh)
	refresh()

	pMain.AddChild(toggleCaption("b22-cap", "Click the track to flip — thumb animates between off and on."))
	pMain.AddChild(toggleRow("b22-wifi-row", "Wi-Fi", wifi))
	pMain.AddChild(toggleRow("b22-dark-row", "Dark mode", dark))
	pMain.AddChild(status)

	pSize := ui.NewPanel("b22-size", "Sizes", 0, 0, 0, 0)
	pSize.AutoHeight = true
	setSpans5TogglePanel(pSize, 12, 12, 6, 5, 5)
	pSize.Gap = 10
	pSize.TitleHeight = 32
	compact := ui.NewToggle("b22-compact", true, 0, 0, 44, 24)
	standard := ui.NewToggle("b22-std", false, 0, 0, 52, 28)
	pSize.AddChild(toggleRow("b22-compact-row", "Compact (44×24)", compact))
	pSize.AddChild(toggleRow("b22-std-row", "Standard (52×28)", standard))

	pOff := ui.NewPanel("b22-off", "Disabled", 0, 0, 0, 0)
	pOff.AutoHeight = true
	setSpans5TogglePanel(pOff, 12, 12, 12, 12, 12)
	pOff.Gap = 10
	pOff.TitleHeight = 32
	off := ui.NewToggle("b22-off-tog", true, 0, 0, 52, 28)
	off.Disabled = true
	pOff.AddChild(off)

	grid.AddChild(pMain)
	grid.AddChild(pSize)
	grid.AddChild(pOff)
	page.Body.AddChild(grid)
}
