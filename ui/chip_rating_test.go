package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestBadgeFilterChipAutoWidth(t *testing.T) {
	sel := NewSignal(true)
	b := NewBadge("b", "Go", BadgePrimary, 0, 0, 0, 32)
	b.Selected = sel
	if b.Bounds().Width < 24 {
		t.Fatalf("auto width too small: %.0f", b.Bounds().Width)
	}
}

func TestBadgeCloseButtonResizesAfterEnable(t *testing.T) {
	b := NewBadge("chip", "Go", BadgePrimary, 0, 0, 0, 28)
	w0 := b.Bounds().Width
	b.SetCloseButton(true)
	w1 := b.Bounds().Width
	if w1 <= w0+10 {
		t.Fatalf("close width = %.0f, want > %.0f", w1, w0+10)
	}
}

func TestBadgeFlexRowKeepsMinWidth(t *testing.T) {
	row := NewContainer("row", 0, 0, 120, 32)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	b := NewBadge("b", "retained-mode", BadgeDefault, 0, 0, 0, 26)
	b.SetCloseButton(true)
	row.AddChild(b)
	row.SetBounds(rl.NewRectangle(0, 0, 120, 32))
	row.Layout()
	if b.Bounds().Width < b.GetMinWidth()-0.5 {
		t.Fatalf("badge squeezed to %.0f, min %.0f", b.Bounds().Width, b.GetMinWidth())
	}
}

func TestRatingIntrinsicWidth(t *testing.T) {
	sig := NewSignal(float32(2))
	r := NewRating("r", sig, 5, 0, 0, 0, 0)
	want := float32(5)*ratingStarSize + 4*ratingStarGap
	if r.Bounds().Width != want {
		t.Fatalf("width = %.0f, want %.0f", r.Bounds().Width, want)
	}
}

func TestRatingFlexRowKeepsWidth(t *testing.T) {
	row := NewContainer("row", 0, 0, 400, 40)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.Gap = 10
	lbl := NewLabel("lbl", "Rate Gru", 0, 0, 0, 22)
	lbl.PreferredWidth = 96
	r := NewRating("r", NewSignal(float32(3)), 5, 0, 0, 0, 0)
	row.AddChild(lbl)
	row.AddChild(r)
	row.SetBounds(rl.NewRectangle(0, 0, 400, 40))
	row.Layout()
	want := float32(5)*ratingStarSize + 4*ratingStarGap
	if r.Bounds().Width < want-0.5 {
		t.Fatalf("rating width %.0f, want >= %.0f", r.Bounds().Width, want)
	}
}

func TestRatingStarBounds(t *testing.T) {
	r := NewRating("r", NewSignal(float32(0)), 5, 0, 0, 160, 28)
	r.SetBounds(rl.NewRectangle(0, 0, 160, 28))
	b0 := r.starBounds(0)
	if b0.Width != ratingStarSize {
		t.Fatalf("star width = %.0f", b0.Width)
	}
}
