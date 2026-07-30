package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestAutoHeightContainerShrinksAfterProbeHeight(t *testing.T) {
	c := NewContainer("auto", 0, 0, 320, 0)
	c.SetStyle("transparent")
	c.Gap = 8

	a := NewLabel("a", "A", 0, 0, 0, 24)
	b := NewLabel("b", "B", 0, 0, 0, 24)
	c.AddChild(a)
	c.AddChild(b)

	c.SetBounds(rl.NewRectangle(0, 0, 320, 4096))
	c.Layout()

	if c.Bounds().Height >= 4096 {
		t.Fatalf("AutoHeight container kept probe height: got %.0f", c.Bounds().Height)
	}
	want := float32(24 + 8 + 24)
	if c.Bounds().Height != want {
		t.Fatalf("AutoHeight container height = %.0f, want %.0f", c.Bounds().Height, want)
	}
}

func TestAutoHeightCardMeasuresChildren(t *testing.T) {
	card := NewCard("card", "Generated Card", 0, 0, 320, 0)
	rt := NewRichText("text", []TextSpan{{Text: "Card text should measure into the card body."}}, 0, 0, 0, 0)
	card.AddChild(rt)

	card.SetBounds(rl.NewRectangle(0, 0, 320, 4096))
	card.Layout()

	if card.Bounds().Height >= 4096 {
		t.Fatalf("AutoHeight card kept probe height: got %.0f", card.Bounds().Height)
	}
	if card.Bounds().Height <= card.TitleHeight {
		t.Fatalf("AutoHeight card did not include child content: got %.0f", card.Bounds().Height)
	}
	if rt.Bounds().Height <= 0 {
		t.Fatalf("RichText child was not measured: got %.0f", rt.Bounds().Height)
	}
}

func TestAutoHeightCardMeasuresFromZeroHeight(t *testing.T) {
	card := NewCard("card", "Generated Card", 0, 0, 320, 0)
	rt := NewRichText("text", []TextSpan{{Text: "Generated cards start at zero height in DocumentSpec."}}, 0, 0, 0, 0)
	card.AddChild(rt)

	card.Layout()

	if card.Bounds().Height <= card.TitleHeight {
		t.Fatalf("AutoHeight card did not expand from zero height: got %.0f", card.Bounds().Height)
	}
	if card.Bounds().Height >= 4096 {
		t.Fatalf("AutoHeight card kept probe height: got %.0f", card.Bounds().Height)
	}
}
