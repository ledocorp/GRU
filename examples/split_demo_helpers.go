// Package examples — shared helpers for SplitView / ResizablePanel demos (Phase A).
package examples

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

const splitSectionHeight = float32(420)

// splitHint is a wrapped muted paragraph for section intros.
func splitHint(id, text string) *ui.RichText {
	rt := ui.NewRichText(id, []ui.TextSpan{{Text: text, Variant: "muted"}}, 0, 0, 0, 0)
	rt.Wrap = true
	return rt
}

// splitPane builds a fixed-height panel suited for SplitView panes: title chrome,
// scrollable viewport, one RichText block (never a stack of Labels).
func splitPane(id, title string, lines []string, code bool) *ui.Panel {
	p := ui.NewPanel(id, title, 0, 0, 0, 0)
	p.AutoHeight = false
	p.ClipChildren = true
	p.Gap = 0
	p.TitleHeight = 28

	vp := ui.NewViewport(id+"-vp", 0, 0, 0, 0)
	vp.SetStyle("transparent")
	vp.SetFlexGrow(1)
	vp.Gap = 0

	span := ui.TextSpan{Text: strings.Join(lines, "\n")}
	if code {
		span.Variant = "code"
	} else {
		span.Style = "default"
	}
	rt := ui.NewRichText(id+"-body", []ui.TextSpan{span}, 0, 0, 0, 0)
	rt.Wrap = true
	vp.AddChild(rt)
	p.AddChild(vp)
	return p
}

func configureSplitSectionPanel(p *ui.Panel) {
	p.AutoHeight = false
	p.ClipChildren = true
}
