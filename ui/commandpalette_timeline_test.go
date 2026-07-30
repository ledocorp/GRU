package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestMatchCommandQuery(t *testing.T) {
	if !matchCommandQuery("", "Open File", "File", "new") {
		t.Fatal("empty query should match all")
	}
	if !matchCommandQuery("open file", "Open File", "", "") {
		t.Fatal("expected multi-word match")
	}
	if matchCommandQuery("zzzz", "Open File", "", "") {
		t.Fatal("expected miss")
	}
	if !matchCommandQuery("pref", "Settings", "", "preferences config") {
		t.Fatal("expected keyword prefix match")
	}
}

func TestCommandPaletteRefilter(t *testing.T) {
	m := &commandPaletteManager{
		items: []CommandPaletteItem{
			{Label: "Home"},
			{Label: "Settings", Keywords: "preferences"},
		},
	}
	m.query = "pref"
	m.refilter()
	if len(m.filtered) != 1 || m.filtered[0] != 1 {
		t.Fatalf("filtered = %v, want [1]", m.filtered)
	}
}

func TestTimelineAutoHeight(t *testing.T) {
	tl := NewTimeline("tl", []TimelineEvent{
		{Title: "A", Subtitle: "one"},
		{Title: "B", Subtitle: "two"},
	}, 0, 0, 400, 0)
	tl.SetBounds(rl.NewRectangle(0, 0, 400, 0))
	tl.Layout()
	if tl.Bounds().Height < timelineEventMinH*2 {
		t.Fatalf("height = %v, expected at least two rows", tl.Bounds().Height)
	}
}
