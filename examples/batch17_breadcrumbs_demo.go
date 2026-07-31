//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"strings"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch17BreadcrumbsScene{} }) }

type batch17BreadcrumbsScene struct {
	BaseScene
	path []string
}

func (s *batch17BreadcrumbsScene) Title() string { return "Batch 17 · Breadcrumbs" }

func (s *batch17BreadcrumbsScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5CrumbPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func crumbCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func (s *batch17BreadcrumbsScene) Build(doc *ui.Document) {
	s.path = []string{"Home", "Projects", "Gru UI"}

	page := MountAppPage(doc, "b17",
		"Widget Batch 17 · Breadcrumbs",
		"Clickable path trail — earlier segments navigate back; the last segment is the current page.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b17-grid", 12)

	pNav := ui.NewPanel("b17-nav", "Navigation", 0, 0, 0, 0)
	pNav.AutoHeight = true
	setSpans5CrumbPanel(pNav, 12, 12, 8, 6, 6)
	pNav.Gap = 10
	pNav.TitleHeight = 32

	crumbs := ui.NewBreadcrumbs("b17-crumbs", s.path, 0, 0, 0, 0)
	status := ui.NewLabel("b17-status", "", 0, 0, 0, 0)
	status.SetStyle("form-value")
	status.Wrap = true

	updateStatus := func() {
		status.Text.Set("Current: " + s.path[len(s.path)-1])
	}
	updateStatus()

	crumbs.OnClick = func(i int) {
		if i < 0 || i >= len(s.path) {
			return
		}
		s.path = s.path[:i+1]
		crumbs.SetItems(s.path)
		updateStatus()
	}

	drillBtn := ui.NewButton("b17-drill", "Drill deeper →", 0, 0, 0, 36)
	drillBtn.OnClick = func() {
		next := fmt.Sprintf("Level %d", len(s.path))
		s.path = append(s.path, next)
		crumbs.SetItems(s.path)
		updateStatus()
	}

	pNav.AddChild(crumbCaption("b17-cap", "Click a segment to pop the trail back to that level."))
	pNav.AddChild(crumbs)
	pNav.AddChild(drillBtn)
	pNav.AddChild(status)

	pLong := ui.NewPanel("b17-long", "Long path", 0, 0, 0, 0)
	pLong.AutoHeight = true
	setSpans5CrumbPanel(pLong, 12, 12, 4, 6, 6)
	pLong.Gap = 10
	pLong.TitleHeight = 32

	longItems := []string{
		"Workspace", "North America", "Engineering", "Platform", "UI", "Widgets", "Forms",
	}
	longCrumbs := ui.NewBreadcrumbs("b17-long-crumbs", longItems, 0, 0, 0, 0)
	longLbl := ui.NewLabel("b17-long-lbl", strings.Join(longItems, " / "), 0, 0, 0, 0)
	longLbl.SetStyle("form-value")
	longLbl.Wrap = true
	longCrumbs.OnClick = func(i int) {
		longLbl.Text.Set("Jumped to: " + longItems[i])
	}

	pLong.AddChild(longCrumbs)
	pLong.AddChild(longLbl)

	grid.AddChild(pNav)
	grid.AddChild(pLong)
	page.Body.AddChild(grid)
}
