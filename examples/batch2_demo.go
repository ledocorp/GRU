//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"strings"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &batch2Scene{} }) }

// batch2Scene demonstrates the SearchBar widget (Batch 2).
//
// Bootstrap-style layout: full-width search field, one help/status line, fixed-height
// bordered list-group (VirtualList). Variant panel shows debounce/icon/width options
// without coupling to the main filter list.
type batch2Scene struct {
	BaseScene
	searchBars []*ui.SearchBar
}

func (s *batch2Scene) Title() string { return "Batch 2 · SearchBar" }

func setSpans5Panel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func sbCaption(id, text string) *ui.RichText { return batchCaption(id, text) }

func (s *batch2Scene) OnUpdate(d *ui.Document, _ float32) {
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return
	}
	mouse := rl.GetMousePosition()
	for _, sb := range s.searchBars {
		if rl.CheckCollisionPointRec(mouse, sb.Bounds()) {
			d.SetFocus(sb)
			return
		}
	}
	d.SetFocus(nil)
}

func sbListRow(name string, selected bool) ui.Node {
	const rowH float32 = 36
	lbl := ui.NewLabel("b2-row", name, 0, 0, 0, rowH)
	lbl.Align = ui.LabelAlignLeft
	lbl.Truncate = true
	if selected {
		lbl.SetStyle("list-selected")
	} else {
		lbl.SetStyle("list-row")
	}
	return lbl
}

func (s *batch2Scene) Build(doc *ui.Document) {
	PreloadScenePhosphor(doc)
	page := MountAppPage(doc, "b2",
		"Widget Batch 2 · SearchBar",
		"Pill search field, debounced Query signal, and filterable list. HSV picker → Batch 26.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b2-grid", 12)

	pMain := ui.NewPanel("p-main", "Live filter", 0, 0, 0, 0)
	pMain.AutoHeight = true
	setSpans5Panel(pMain, 12, 12, 6, 6, 6)
	pMain.Gap = 8
	pMain.TitleHeight = 32

	sb := ui.NewSearchBar("sb-main", "Search fruits…", 0, 0, 0, 38)
	sb.DebounceDelay = 0.3

	statusLabel := ui.NewLabel("sb-status", fmt.Sprintf("%d items", 11), 0, 0, 0, 0)
	statusLabel.SetStyle("form-value")

	const listH float32 = 220
	items := []string{
		"Apple", "Banana", "Cherry", "Date", "Elderberry",
		"Fig", "Grape", "Honeydew", "Jackfruit", "Kiwi",
		"Long fruit name that should truncate inside the list row",
	}
	fruitBinding := ui.NewListBinding(items)
	fruitList := ui.NewVirtualList("sb-fruit-list", fruitBinding,
		func(name string, _ int, selected bool) ui.Node {
			return sbListRow(name, selected)
		}, 36, 0, 0, 0, listH)
	fruitList.SetStyle("list")

	updateStatus := func(live, query string) {
		q := strings.ToLower(strings.TrimSpace(query))
		filtered := make([]string, 0, len(items))
		for _, name := range items {
			if q == "" || strings.Contains(strings.ToLower(name), q) {
				filtered = append(filtered, name)
			}
		}
		fruitBinding.SetItems(filtered)

		switch {
		case query == "" && live == "":
			statusLabel.Text.Set(fmt.Sprintf("%d items · type to filter (Query debounced 0.3 s)", len(items)))
		case query == "":
			statusLabel.Text.Set(fmt.Sprintf("typing %q · %d chars · filter applies after pause", live, len(live)))
		case len(filtered) == len(items):
			statusLabel.Text.Set(fmt.Sprintf("Query %q · all %d items", query, len(items)))
		default:
			statusLabel.Text.Set(fmt.Sprintf("Query %q · %d of %d items", query, len(filtered), len(items)))
		}
		statusLabel.MarkDirty()
	}

	sb.Text.Subscribe(func() { updateStatus(sb.GetText(), sb.Query.Get()) })
	sb.Query.Subscribe(func() { updateStatus(sb.GetText(), sb.Query.Get()) })

	pMain.AddChild(sbCaption("b2-main-cap", "Search icon + inset border; one status line; scrollable list with clipped rows."))
	pMain.AddChild(sb)
	pMain.AddChild(statusLabel)
	pMain.AddChild(fruitList)

	pVariants := ui.NewPanel("p-variants", "Variants", 0, 0, 0, 0)
	pVariants.AutoHeight = true
	setSpans5Panel(pVariants, 12, 12, 6, 6, 6)
	pVariants.Gap = 8
	pVariants.TitleHeight = 32

	variantStatus := func(lbl *ui.Label, q string) {
		if q == "" {
			lbl.Text.Set("Query: (none)")
		} else {
			lbl.Text.Set(fmt.Sprintf("Query: %q", q))
		}
		lbl.MarkDirty()
	}

	sb2 := ui.NewSearchBar("sb-instant", "Instant (no debounce)…", 0, 0, 0, 38)
	sb2.ShowIcon = false
	sb2.DebounceDelay = 0
	sb2Status := ui.NewLabel("v1-status", "Query: (none)", 0, 0, 0, 0)
	sb2Status.SetStyle("form-value")
	sb2.Query.Subscribe(func() { variantStatus(sb2Status, sb2.Query.Get()) })

	sb3 := ui.NewSearchBar("sb-slow", "Slow debounce (1 s)…", 0, 0, 0, 38)
	sb3.ShowIcon = true
	sb3.DebounceDelay = 1.0
	sb3Status := ui.NewLabel("v2-status", "Query: (none)", 0, 0, 0, 0)
	sb3Status.SetStyle("form-value")
	sb3.Query.Subscribe(func() { variantStatus(sb3Status, sb3.Query.Get()) })

	sb4Row := ui.NewContainer("sb4-row", 0, 0, 0, 38)
	sb4Row.FlexDirection = ui.FlexRow
	sb4Row.SetStyle("transparent")
	sb4 := ui.NewSearchBar("sb-narrow", "Fixed 200 px…", 0, 0, 200, 38)
	sb4.ShowIcon = true
	sb4.DebounceDelay = 0.3
	sb4Status := ui.NewLabel("v3-status", "Query: (none)", 0, 0, 0, 0)
	sb4Status.SetStyle("form-value")
	sb4.Query.Subscribe(func() { variantStatus(sb4Status, sb4.Query.Get()) })

	pVariants.AddChild(sbCaption("b2-v-cap", "Each variant reports its own Query — they do not filter the left list."))
	pVariants.AddChild(sbCaption("b2-v1-cap", "Instant · no icon"))
	pVariants.AddChild(sb2)
	pVariants.AddChild(sb2Status)
	pVariants.AddChild(sbCaption("b2-v2-cap", "Icon · DebounceDelay = 1.0 s"))
	pVariants.AddChild(sb3)
	pVariants.AddChild(sb3Status)
	sb4Row.AddChild(sb4)
	pVariants.AddChild(sbCaption("b2-v3-cap", "Fixed 200 px width"))
	pVariants.AddChild(sb4Row)
	pVariants.AddChild(sb4Status)
	pVariants.AddChild(sbCaption("b2-hint", "Click a field to focus · Esc clears · × when non-empty"))

	grid.AddChild(pMain)
	grid.AddChild(pVariants)
	page.Body.AddChild(grid)

	s.searchBars = []*ui.SearchBar{sb, sb2, sb3, sb4}
}
