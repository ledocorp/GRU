//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch13PaginationScene{} }) }

type batch13PaginationScene struct {
	BaseScene
}

func (s *batch13PaginationScene) Title() string { return "Batch 13 · Pagination" }

func (s *batch13PaginationScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5PagPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func pagCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func mountPag(id string, total int, cur *ui.Signal[int]) *ui.Pagination {
	pg := ui.NewPagination(id, total, cur, 0, 0, 0, 0)
	pg.SetFlexGrow(0)
	return pg
}

func (s *batch13PaginationScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b13",
		"Widget Batch 13 · Pagination",
		"Prev / numbered pages / next — centered in each panel when space allows.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b13-grid", 12)

	pMain := ui.NewPanel("b13-main", "Table footer", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5PagPanel(pMain, 12, 12, 6, 4, 4)
	pMain.Gap = 10
	pMain.TitleHeight = 32

	pageIdx := ui.NewSignal(0)
	pages := mountPag("b13-pages", 12, pageIdx)
	status := ui.NewLabel("b13-status", "", 0, 0, 0, 0)
	status.SetStyle("form-value")
	status.Wrap = true
	pageIdx.Subscribe(func() {
		status.Text.Set(fmt.Sprintf("Showing page %d of 12", pageIdx.Get()+1))
	})
	pageIdx.Set(pageIdx.Get())

	pMain.AddChild(pagCaption("b13-cap", "All twelve page numbers inline — typical table footer when page count is modest."))
	pMain.AddChild(pages)
	pMain.AddChild(status)

	pFew := ui.NewPanel("b13-few", "Few pages", 0, 0, 0, 0)
	pFew.AutoHeight = true
	setSpans5PagPanel(pFew, 12, 12, 6, 4, 4)
	pFew.Gap = 10
	pFew.TitleHeight = 32
	fewIdx := ui.NewSignal(1)
	few := mountPag("b13-few-pg", 3, fewIdx)
	pFew.AddChild(few)

	pWide := ui.NewPanel("b13-wide", "Wide strip", 0, 0, 0, 0)
	pWide.AutoHeight = true
	setSpans5PagPanel(pWide, 12, 12, 12, 4, 4)
	pWide.Gap = 10
	pWide.TitleHeight = 32
	wideIdx := ui.NewSignal(4)
	wide := mountPag("b13-wide-pg", 9, wideIdx)
	pWide.AddChild(pagCaption("b13-wide-cap", "Nine pages — all page buttons visible inline."))
	pWide.AddChild(wide)

	pMany := ui.NewPanel("b13-many", "Large catalog", 0, 0, 0, 0)
	pMany.AutoHeight = true
	setSpans5PagPanel(pMany, 12, 12, 12, 12, 12)
	pMany.Gap = 10
	pMany.TitleHeight = 32
	manyIdx := ui.NewSignal(49)
	many := mountPag("b13-many-pg", 240, manyIdx)
	manyStatus := ui.NewLabel("b13-many-status", "", 0, 0, 0, 0)
	manyStatus.SetStyle("form-value")
	manyStatus.Wrap = true
	manyIdx.Subscribe(func() {
		page := manyIdx.Get() + 1
		manyStatus.Text.Set(fmt.Sprintf("Page %d of 240 · 1 … %d … 240", page, page))
	})
	manyIdx.Set(manyIdx.Get())
	pMany.AddChild(pagCaption("b13-many-cap", "240 pages collapse to 1 … window … 240. Prev/next stay pinned; scroll the middle numbers when narrow."))
	pMany.AddChild(many)
	pMany.AddChild(manyStatus)

	grid.AddChild(pMain)
	grid.AddChild(pFew)
	grid.AddChild(pWide)
	grid.AddChild(pMany)
	page.Body.AddChild(grid)
}
