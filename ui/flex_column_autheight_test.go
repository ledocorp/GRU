package ui

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Regression: preview markdown stacks RichText blocks in a flex column; on width
// shrink pass 2 must re-layout AutoHeight children before advancing Y.
func TestFlexColumnAutoHeightRichTextStackOnWidthShrink(t *testing.T) {
	col := NewContainer("col", 0, 0, 200, 0)
	col.LayoutType = LayoutFlex
	col.FlexDirection = FlexColumn
	col.AutoHeight = true
	col.Gap = 8

	rt1 := NewRichText("rt1", []TextSpan{{Text: strings.Repeat("wrap ", 40)}}, 0, 0, 0, 0)
	rt1.Wrap = true
	rt1.AutoHeight = true
	rt2 := NewRichText("rt2", []TextSpan{{Text: "Second preview block"}}, 0, 0, 0, 0)
	rt2.Wrap = true
	rt2.AutoHeight = true
	col.AddChild(rt1)
	col.AddChild(rt2)

	col.SetBounds(rl.NewRectangle(0, 0, 200, 800))
	col.Layout()

	col.InvalidateLayoutPassCache()
	col.MarkDirty()
	col.SetBounds(rl.NewRectangle(0, 0, 120, 800))
	col.Layout()

	b1 := rt1.Bounds()
	b2 := rt2.Bounds()
	if b2.Y < b1.Y+b1.Height-0.5 {
		t.Fatalf("blocks overlap after shrink: rt1 bottom=%.1f rt2 top=%.1f (heights %.1f / %.1f)",
			b1.Y+b1.Height, b2.Y, b1.Height, b2.Height)
	}
}

// Regression: preview lane stacks heading RichText, table Card, then body copy.
// Card Layout after pass-2 Y stack must not leave the paragraph overlapping.
func TestFlexColumnAutoHeightHeadingTableCardStackOnWidthShrink(t *testing.T) {
	col := NewContainer("preview-lane", 0, 0, 320, 0)
	col.LayoutType = LayoutFlex
	col.FlexDirection = FlexColumn
	col.AutoHeight = true
	col.Gap = 12

	h1 := NewRichText("h1", []TextSpan{{Text: "# Preview heading " + strings.Repeat("wrap ", 18), Variant: "h1"}}, 0, 0, 0, 0)
	h1.Wrap = true
	h1.AutoHeight = true

	tableCard := NewCard("table", "", 0, 0, 0, 0)
	tableCard.Title = ""
	tableCard.TitleHeight = 0
	tableCard.AutoHeight = true
	tableCard.Gap = 0
	row := NewContainer("row", 0, 0, 0, 0)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.Gap = 8
	row.AutoHeight = true
	row.PreferredWidth = 280
	cellA := NewRichText("c1", []TextSpan{{Text: strings.Repeat("cell ", 30)}}, 0, 0, 120, 0)
	cellA.Wrap = true
	cellA.AutoHeight = true
	cellA.PreferredWidth = 120
	cellB := NewRichText("c2", []TextSpan{{Text: strings.Repeat("wide ", 25)}}, 0, 0, 120, 0)
	cellB.Wrap = true
	cellB.AutoHeight = true
	cellB.PreferredWidth = 120
	row.AddChild(cellA)
	row.AddChild(cellB)
	tableCard.AddChild(row)

	body := NewRichText("body", []TextSpan{{Text: "Paragraph below the table."}}, 0, 0, 0, 0)
	body.Wrap = true
	body.AutoHeight = true

	col.AddChild(h1)
	col.AddChild(tableCard)
	col.AddChild(body)

	col.SetBounds(rl.NewRectangle(0, 0, 320, 900))
	col.Layout()

	col.InvalidateLayoutPassCache()
	col.MarkDirty()
	col.SetBounds(rl.NewRectangle(0, 0, 180, 900))
	col.Layout()

	cardB := tableCard.Bounds()
	h1B := h1.Bounds()
	bodyB := body.Bounds()
	if cardB.Y < h1B.Y+h1B.Height-0.5 {
		t.Fatalf("table overlaps heading: h1 bottom=%.1f card top=%.1f", h1B.Y+h1B.Height, cardB.Y)
	}
	if bodyB.Y < cardB.Y+cardB.Height-0.5 {
		t.Fatalf("body overlaps table: card bottom=%.1f body top=%.1f (h1 h=%.1f card h=%.1f)",
			cardB.Y+cardB.Height, bodyB.Y, h1B.Height, cardB.Height)
	}
}
