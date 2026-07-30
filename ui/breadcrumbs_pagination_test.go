package ui

import "testing"

func TestBreadcrumbsMeasureWidth(t *testing.T) {
	bc := NewBreadcrumbs("bc", []string{"Inbox", "Message"}, 0, 0, 0, 0)
	if bc.Bounds().Width < 40 {
		t.Fatalf("width too small: %.0f", bc.Bounds().Width)
	}
}

func TestPaginationPageCount(t *testing.T) {
	cur := NewSignal(2)
	p := NewPagination("p", 5, cur, 0, 0, 0, 0)
	if p.controlCount() != 7 {
		t.Fatalf("controls = %d, want 7 (prev+5+next)", p.controlCount())
	}
}

func TestPaginationNextPinnedWhenOverflow(t *testing.T) {
	cur := NewSignal(0)
	p := NewPagination("p", 12, cur, 0, 0, 120, 40)
	p.Layout()
	next := p.nextBounds()
	b := p.Bounds()
	if next.X+next.Width > b.X+b.Width+0.5 {
		t.Fatalf("next extends past widget: next=%v bounds=%v", next, b)
	}
	if next.X < b.X {
		t.Fatalf("next before widget: next=%v bounds=%v", next, b)
	}
}

func TestBreadcrumbsSetItems(t *testing.T) {
	bc := NewBreadcrumbs("bc", []string{"A"}, 0, 0, 0, 0)
	w0 := bc.Bounds().Width
	bc.SetItems([]string{"A", "B", "C"})
	if bc.Bounds().Width <= w0 {
		t.Fatalf("width should grow after SetItems")
	}
}
