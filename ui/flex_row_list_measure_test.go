package ui

import (
	"strings"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Regression: preview list rows are flex rows with a fixed inset + flex-grow RichText.
// Pre-measure must use remaining width so wrapped copy height matches pass 2.
func TestFlexRowFlexGrowRichTextMeasuresAtRemainingWidth(t *testing.T) {
	row := NewContainer("row", 0, 0, 300, 0)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.AutoHeight = true
	row.Gap = 10

	inset := NewContainer("in", 0, 0, 60, 0)
	inset.SetStyle("transparent")

	words := strings.Fields(strings.Repeat("list ", 30))
	spans := make([]TextSpan, len(words))
	for i, w := range words {
		spans[i] = TextSpan{Text: w + " "}
	}
	rt := NewRichText("rt", spans, 0, 0, 0, 0)
	rt.Wrap = true
	rt.AutoHeight = true
	rt.SetFlexGrow(1)

	row.AddChild(inset)
	row.AddChild(rt)

	row.SetBounds(rl.NewRectangle(0, 0, 300, 800))
	row.Layout()

	ref := NewRichText("ref", spans, 0, 0, 230, 0)
	ref.Wrap = true
	ref.AutoHeight = true
	ref.SetBounds(rl.NewRectangle(0, 0, 230, 0))
	ref.Layout()

	if rt.Bounds().Height < ref.Bounds().Height-0.5 {
		t.Fatalf("row text height %.1f < ref at 230px width %.1f", rt.Bounds().Height, ref.Bounds().Height)
	}
}
