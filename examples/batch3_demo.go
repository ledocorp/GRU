//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch3Scene{} }) }

// batch3Scene demonstrates the Badge widget (Batch 3).
type batch3Scene struct {
	BaseScene
}

func (s *batch3Scene) Title() string { return "Batch 3 · Badge" }

func (s *batch3Scene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5BadgePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func b3Caption(id, text string) *ui.RichText {
	return batchCaption(id, text)
}

func b3Hint(id, text string) *ui.RichText {
	return batchCaption(id, text)
}

func addBadgeFlowRow(parent ui.Node, id string, gap float32, badges []*ui.Badge) {
	row := ui.NewContainer(id, 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.SetFlexWrap(true)
	row.ClipChildren = true
	row.Gap = gap
	row.AutoHeight = true
	row.SetStyle("transparent")
	for _, b := range badges {
		row.AddChild(b)
	}
	parent.AddChild(row)
}

var badgeVariantSamples = []struct {
	text    string
	variant ui.BadgeVariant
}{
	{"Default", ui.BadgeDefault},
	{"Primary", ui.BadgePrimary},
	{"Success", ui.BadgeSuccess},
	{"Warning", ui.BadgeWarning},
	{"Danger", ui.BadgeDanger},
	{"Info", ui.BadgeInfo},
}

func newVariantBadges(idPrefix string, shape ui.BadgeShape, h float32) []*ui.Badge {
	out := make([]*ui.Badge, 0, len(badgeVariantSamples))
	for i, v := range badgeVariantSamples {
		b := ui.NewBadge(fmt.Sprintf("%s-%d", idPrefix, i), v.text, v.variant, 0, 0, 0, h)
		b.Shape = shape
		out = append(out, b)
	}
	return out
}

func (s *batch3Scene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b3",
		"Badge",
		"Pill-shaped status chips with six color variants and an optional close button.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b3-grid", 16)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel: Variants
	// ══════════════════════════════════════════════════════════════════════════
	pVariants := ui.NewPanel("p-b3-variants", "Variants", 0, 0, 0, 0)
	pVariants.AutoHeight = true
	setSpans5BadgePanel(pVariants, 12, 12, 12, 4, 4)
	pVariants.Gap = 12

	pVariants.AddChild(b3Caption("var-intro",
		"Six BadgeVariant presets. Width auto-sizes to label text."))

	addBadgeFlowRow(pVariants, "var-pills", 8, newVariantBadges("var-pill", ui.BadgeShapePill, 26))

	pVariants.AddChild(ui.NewSeparator("var-sep1", "", 0, 0, 0, 8))

	pVariants.AddChild(b3Caption("long-intro",
		"Longer labels — same pill styling, wider chips:"))
	addBadgeFlowRow(pVariants, "long-pills", 8, []*ui.Badge{
		ui.NewBadge("b-long1", "In Progress", ui.BadgeInfo, 0, 0, 0, 26),
		ui.NewBadge("b-long2", "Needs Review", ui.BadgeWarning, 0, 0, 0, 26),
		ui.NewBadge("b-long3", "Deployed", ui.BadgeSuccess, 0, 0, 0, 26),
	})

	pVariants.AddChild(ui.NewSeparator("var-sep2", "", 0, 0, 0, 8))

	pVariants.AddChild(b3Caption("rect-intro",
		"Rectangle shape — identical colors and padding, sharp corners only:"))
	addBadgeFlowRow(pVariants, "rect-variants", 8, newVariantBadges("rect", ui.BadgeShapeRect, 26))
	addBadgeFlowRow(pVariants, "rect-long", 8, func() []*ui.Badge {
		long := []struct {
			text    string
			variant ui.BadgeVariant
		}{
			{"In Progress", ui.BadgeInfo},
			{"Needs Review", ui.BadgeWarning},
			{"Deployed", ui.BadgeSuccess},
		}
		out := make([]*ui.Badge, 0, len(long))
		for i, l := range long {
			b := ui.NewBadge(fmt.Sprintf("rect-long-%d", i), l.text, l.variant, 0, 0, 0, 26)
			b.Shape = ui.BadgeShapeRect
			out = append(out, b)
		}
		return out
	}())

	// ══════════════════════════════════════════════════════════════════════════
	// Panel: Close Chips
	// ══════════════════════════════════════════════════════════════════════════
	pClose := ui.NewPanel("p-b3-close", "Close Chips", 0, 0, 0, 0)
	pClose.AutoHeight = true
	setSpans5BadgePanel(pClose, 12, 12, 6, 4, 4)
	pClose.Gap = 10

	pClose.AddChild(b3Caption("cl-hdr",
		"Click the circle × to dismiss a chip. Dismissed count is tracked below."))

	dismissedCount := 0
	dismissLabel := ui.NewLabel("cl-dismissed", "Dismissed: 0", 0, 0, 0, 0)
	dismissLabel.SetStyle("form-value")
	dismissLabel.Align = ui.LabelAlignLeft

	type closeChip struct {
		text    string
		variant ui.BadgeVariant
	}
	chips := []closeChip{
		{"Go", ui.BadgePrimary},
		{"raylib", ui.BadgeInfo},
		{"retained-mode", ui.BadgeDefault},
		{"Signals", ui.BadgePrimary},
		{"gg vectors", ui.BadgeSuccess},
		{"SDF fonts", ui.BadgeInfo},
		{"Inspector", ui.BadgeWarning},
		{"VirtualList", ui.BadgeDefault},
		{"Overlay", ui.BadgeDanger},
		{"FlexLayout", ui.BadgeSuccess},
		{"Batch 3", ui.BadgePrimary},
		{"Badge", ui.BadgeWarning},
	}

	closeBadges := make([]*ui.Badge, 0, len(chips))
	for i, c := range chips {
		chip := ui.NewBadge(fmt.Sprintf("chip-%d", i), c.text, c.variant, 0, 0, 0, 28)
		chip.SetCloseButton(true)
		chipCopy := chip
		chip.OnClose = func() {
			chipCopy.Hide()
			dismissedCount++
			dismissLabel.Text.Set(fmt.Sprintf("Dismissed: %d", dismissedCount))
			dismissLabel.MarkDirty()
		}
		closeBadges = append(closeBadges, chip)
	}
	addBadgeFlowRow(pClose, "cl-wrap", 10, closeBadges)
	pClose.AddChild(dismissLabel)

	pClose.AddChild(ui.NewSeparator("cl-sep", "", 0, 0, 0, 8))
	pClose.AddChild(b3Hint("cl-hint",
		"Dismissed badges call badge.Hide(). OnClose runs synchronously on click."))

	// ══════════════════════════════════════════════════════════════════════════
	// Panel: Live Demo
	// ══════════════════════════════════════════════════════════════════════════
	pLive := ui.NewPanel("p-b3-live", "Live Demo", 0, 0, 0, 0)
	pLive.AutoHeight = true
	setSpans5BadgePanel(pLive, 12, 12, 6, 4, 4)
	pLive.Gap = 10

	pLive.AddChild(b3Caption("lv-hdr", "Badges react to a shared counter signal."))

	counter := ui.NewSignal(0)
	statusBadge := ui.NewBadge("lv-status", "Count: 0", ui.BadgeDefault, 0, 0, 0, 28)
	versionBadge := ui.NewBadge("lv-version", "v2.4.1", ui.BadgePrimary, 0, 0, 0, 26)

	btnInc := ui.NewButton("lv-inc", "+ Inc", 0, 0, 0, 34)
	btnInc.SetStyle("primary")
	btnDec := ui.NewButton("lv-dec", "− Dec", 0, 0, 0, 34)
	btnDec.SetStyle("button")
	btnReset := ui.NewButton("lv-reset", "Reset", 0, 0, 0, 34)
	btnReset.SetStyle("default")

	counter.Subscribe(func() {
		n := counter.Get()
		statusBadge.Text.Set(fmt.Sprintf("Count: %d", n))
		switch {
		case n < 0:
			statusBadge.SetVariant(ui.BadgeDanger)
		case n == 0:
			statusBadge.SetVariant(ui.BadgeDefault)
		case n < 5:
			statusBadge.SetVariant(ui.BadgeInfo)
		case n < 10:
			statusBadge.SetVariant(ui.BadgeSuccess)
		default:
			statusBadge.SetVariant(ui.BadgeWarning)
		}
	})

	btnInc.OnClick = func() { counter.Set(counter.Get() + 1) }
	btnDec.OnClick = func() { counter.Set(counter.Get() - 1) }
	btnReset.OnClick = func() { counter.Set(0) }

	btnRow := ui.NewContainer("lv-btn-row", 0, 0, 0, 38)
	btnRow.FlexDirection = ui.FlexRow
	btnRow.Gap = 8
	btnRow.SetStyle("transparent")
	btnRow.AddChild(btnInc)
	btnRow.AddChild(btnDec)
	btnRow.AddChild(btnReset)
	btnInc.SetFlexGrow(1)
	btnDec.SetFlexGrow(1)
	btnReset.SetFlexGrow(1)

	badgeRow := ui.NewContainer("lv-badge-row", 0, 0, 0, 0)
	badgeRow.FlexDirection = ui.FlexRow
	badgeRow.Gap = 10
	badgeRow.SetStyle("transparent")
	badgeRow.AddChild(statusBadge)
	badgeRow.AddChild(versionBadge)

	pLive.AddChild(btnRow)
	pLive.AddChild(badgeRow)

	pLive.AddChild(b3Caption("lv-thresh-hdr", "Variant thresholds:"))

	type threshold struct {
		desc    string
		variant ui.BadgeVariant
	}
	thresholds := []threshold{
		{"n < 0   → Danger", ui.BadgeDanger},
		{"n = 0   → Default", ui.BadgeDefault},
		{"n < 5   → Info", ui.BadgeInfo},
		{"n < 10  → Success", ui.BadgeSuccess},
		{"n ≥ 10  → Warning", ui.BadgeWarning},
	}
	for i, t := range thresholds {
		tr := ui.NewContainer(fmt.Sprintf("lv-th-%d", i), 0, 0, 0, 0)
		tr.FlexDirection = ui.FlexRow
		tr.Gap = 10
		tr.SetStyle("transparent")

		lbl := ui.NewLabel(fmt.Sprintf("lv-th-lbl-%d", i), t.desc, 0, 0, 0, 0)
		lbl.SetStyle("form-label")
		lbl.Align = ui.LabelAlignLeft
		lbl.Wrap = true
		lbl.Truncate = true
		lbl.SetFlexGrow(1)

		dot := ui.NewBadge(fmt.Sprintf("lv-th-dot-%d", i),
			badgeVariantSamples[t.variant].text, t.variant, 0, 0, 0, 26)
		dot.Shape = ui.BadgeShapeRect

		tr.AddChild(lbl)
		tr.AddChild(dot)
		pLive.AddChild(tr)
	}

	grid.AddChild(pVariants)
	grid.AddChild(pClose)
	grid.AddChild(pLive)
	page.Body.AddChild(grid)
}
