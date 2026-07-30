package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestCardDrawStyleWithDocumentRecipeOverrides(t *testing.T) {
	card := NewCard("doc-callout", "Callout", 0, 0, 200, 80)
	applyCardRecipeVariant(card, "callout")
	applyElementPadding(&card.Element, 14)

	st := card.GetStyle()
	want, ok := MergedComponentVariantStyle("card", "callout")
	if !ok {
		t.Fatal("callout variant missing from theme")
	}
	if st.BackgroundColor != want.BackgroundColor {
		t.Fatalf("drawStyle background = %+v, want callout %+v", st.BackgroundColor, want.BackgroundColor)
	}
	if st.Padding != 14 {
		t.Fatalf("drawStyle padding = %.0f, want 14", st.Padding)
	}
}

func TestCardDrawStyleWithBlockStyleOverride(t *testing.T) {
	bg := "#FCD34D"
	border := "#B45309"
	card := NewCard("amber", "style override", 0, 0, 200, 60)
	applyElementPadding(&card.Element, 14)
	applyDocBlockStyleOverrides(&card.Element, DocBlock{
		Style: &DocBlockStyle{BackgroundColor: &bg, BorderColor: &border},
	})

	st := card.GetStyle()
	wantBg, err := parseHexColor(bg)
	if err != nil {
		t.Fatal(err)
	}
	if st.BackgroundColor != wantBg {
		t.Fatalf("drawStyle background = %+v, want amber %+v", st.BackgroundColor, wantBg)
	}
}

func TestNestedAutoHeightCardInAutoHeightCard(t *testing.T) {
	outer := NewCard("outer", "Surfaces", 0, 0, 400, 0)
	outer.AddChild(NewRichText("intro", []TextSpan{{Text: "Parent card intro."}}, 0, 0, 0, 0))

	inner := NewCard("inner", "Tip", 0, 0, 0, 0)
	inner.SetStyleVariant("card", "callout")
	inner.AddChild(NewRichText("callout", []TextSpan{
		{Text: "Nested callout inside a parent card — same layout path as callout blocks in .gru."},
	}, 0, 0, 0, 0))
	outer.AddChild(inner)

	outer.SetBounds(rl.NewRectangle(0, 0, 400, 4096))
	outer.Layout()

	if outer.Bounds().Height >= 4096 {
		t.Fatalf("outer kept probe height: %.0f", outer.Bounds().Height)
	}
	if inner.Bounds().Height <= inner.TitleHeight {
		t.Fatalf("inner card clipped: height %.0f", inner.Bounds().Height)
	}
	innerBottom := inner.Bounds().Y + inner.Bounds().Height
	outerBottom := outer.Bounds().Y + outer.Bounds().Height
	if innerBottom > outerBottom+1 {
		t.Fatalf("inner bottom %.0f past outer bottom %.0f", innerBottom, outerBottom)
	}
}

func TestNestedCardFillsParentBodyWidth(t *testing.T) {
	const outerW = float32(400)
	outer := NewCard("outer", "", 0, 0, outerW, 0)
	inner := NewCard("inner", "Code", 0, 0, 0, 0)
	inner.SetStyleVariant("card", "code")
	inner.AddChild(NewRichText("code", []TextSpan{{Text: "fn main() {}", Variant: "code"}}, 0, 0, 0, 0))
	outer.AddChild(inner)

	outer.Layout()

	padding := float32(16)
	wantW := outerW - 2*padding
	if w := inner.Bounds().Width; w < wantW-1 || w > wantW+1 {
		t.Fatalf("inner width = %.0f, want ~%.0f (full body band)", w, wantW)
	}
}

func TestNestedCardAutoHeightIncludesShadowBleed(t *testing.T) {
	outer := NewCard("outer", "", 0, 0, 400, 0)
	inner := NewCard("inner", "Tip", 0, 0, 0, 0)
	inner.SetStyleVariant("card", "callout")
	inner.AddChild(NewRichText("callout", []TextSpan{{Text: "Nested callout shadow room."}}, 0, 0, 0, 0))
	outer.AddChild(inner)

	outer.Layout()

	padding := float32(16)
	innerBottom := inner.Bounds().Y + inner.Bounds().Height
	outerBottom := outer.Bounds().Y + outer.Bounds().Height
	wantMin := innerBottom + RaisedSurfaceShadowBleed + padding - 1
	if outerBottom < wantMin {
		t.Fatalf("outer bottom %.0f < inner+shadow+pad %.0f", outerBottom, wantMin)
	}
}

func TestCardPaddingOverrideAffectsChildX(t *testing.T) {
	flush := NewCard("flush", "padding: 0", 0, 0, 200, 0)
	applyElementPadding(&flush.Element, 0)
	pipeFlush := NewRichText("pipe-flush", []TextSpan{{Text: "|"}}, 0, 0, 12, 24)
	flush.AddChild(pipeFlush)

	inset := NewCard("inset", "padding: xl", 0, 0, 200, 0)
	applyElementPadding(&inset.Element, 24)
	pipeInset := NewRichText("pipe-inset", []TextSpan{{Text: "|"}}, 0, 0, 12, 24)
	inset.AddChild(pipeInset)

	flush.SetBounds(rl.NewRectangle(0, 0, 200, 100))
	inset.SetBounds(rl.NewRectangle(0, 0, 200, 100))
	flush.Layout()
	inset.Layout()

	delta := pipeInset.Bounds().X - pipeFlush.Bounds().X
	if delta < 20 || delta > 28 {
		t.Fatalf("inset pipe X delta = %.0f, want ~24px padding difference", delta)
	}
}

func TestCardZeroPaddingHonoredInLayout(t *testing.T) {
	card := NewCard("flush", "padding: 0", 0, 0, 200, 0)
	applyElementPadding(&card.Element, 0)

	row := NewContainer("row", 0, 0, 0, 0)
	applyElementPadding(&row.Element, 0)
	row.FlexDirection = FlexRow
	pipe := NewRichText("pipe", []TextSpan{{Text: "|"}}, 0, 0, 28, 24)
	row.AddChild(pipe)
	card.AddChild(row)

	card.Layout()

	bodyLeft := card.bounds.X
	if pipe.Bounds().X > bodyLeft+0.5 {
		t.Fatalf("pipe X %.0f, want flush at body left %.0f", pipe.Bounds().X, bodyLeft)
	}
}

func TestFixedHeightPresetCardRichTextStaysInBody(t *testing.T) {
	card := NewCard("outer", "neo-glow-card", 0, 0, 320, 120)
	glow := float32(0.55)
	if err := card.SetStylePreset("neo-glow-card", PresetProps{GlowIntensity: &glow}); err != nil {
		t.Fatal(err)
	}
	rt := NewRichText("body", []TextSpan{{
		Text: "Dark indigo card with outer halo + inner glow rings (glowIntensity prop).",
	}}, 0, 0, 0, 0)
	card.AddChild(rt)

	card.Layout()

	padding := card.GetStyle().Padding
	titleOff := card.TitleHeight
	maxBottom := card.Bounds().Y + card.Bounds().Height - padding
	rtBottom := rt.Bounds().Y + rt.Bounds().Height
	if rtBottom > maxBottom+1 {
		t.Fatalf("richtext bottom %.0f exceeds card body bottom %.0f", rtBottom, maxBottom)
	}
	if rt.Bounds().Y < card.Bounds().Y+titleOff+padding-0.5 {
		t.Fatalf("richtext Y %.0f above body top", rt.Bounds().Y)
	}
	if rt.GetStyle().Padding != 0 {
		t.Fatalf("preset richtext padding = %v, want 0 (chrome pads body)", rt.GetStyle().Padding)
	}
}

func TestNestedCardInFixedHeightCardClampsToBody(t *testing.T) {
	outer := NewCard("outer", "Fixed", 0, 0, 320, 120)
	inner := NewCard("inner", "Tip", 0, 0, 0, 0)
	inner.SetStyleVariant("card", "callout")
	long := NewRichText("long", []TextSpan{
		{Text: "Line one.\nLine two.\nLine three.\nLine four.\nLine five.\nLine six."},
	}, 0, 0, 0, 0)
	inner.AddChild(long)
	outer.AddChild(inner)

	outer.Layout()

	padding := float32(16)
	titleOff := outer.TitleHeight
	maxBottom := outer.Bounds().Y + outer.Bounds().Height - padding
	innerBottom := inner.Bounds().Y + inner.Bounds().Height
	if innerBottom > maxBottom+1 {
		t.Fatalf("inner bottom %.0f exceeds outer body bottom %.0f", innerBottom, maxBottom)
	}
	if inner.Bounds().Y < outer.Bounds().Y+titleOff+padding-0.5 {
		t.Fatalf("inner Y %.0f above body top", inner.Bounds().Y)
	}
}
