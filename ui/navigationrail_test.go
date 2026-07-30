package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestNavigationRailSelection(t *testing.T) {
	tab := NewSignal(1)
	rail := NewNavigationRail("rail", []BottomNavItem{
		{Icon: "A", Label: "One"},
		{Icon: "B", Label: "Two"},
	}, tab, 0, 0, 0, 200)

	if tab.Get() != 1 {
		t.Fatalf("selected = %d, want 1", tab.Get())
	}
	if rail.bounds.Width != navigationRailWidth {
		t.Fatalf("width = %.0f, want %.0f", rail.bounds.Width, navigationRailWidth)
	}
	rail.SetBounds(rl.NewRectangle(0, 0, 12, 200))
	if rail.bounds.Width != navigationRailWidth {
		t.Fatalf("after SetBounds(12) width = %.0f, want %.0f", rail.bounds.Width, navigationRailWidth)
	}

	changed := -1
	rail.OnChange = func(i int) { changed = i }
	rail.hoverIdx = 0
	rail.Selected.Set(0)
	rail.OnChange(0)
	if changed != 0 {
		t.Fatalf("OnChange index = %d, want 0", changed)
	}
}
