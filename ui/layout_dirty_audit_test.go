package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// layoutDirtyClearCase exercises widgets whose Layout() must clear layoutDirty
// (see docs/IDLE_INVARIANTS.md). Add a row when overriding Layout on Element embeds.
type layoutDirtyClearCase struct {
	name string
	mk   func() Node
}

func TestLayoutOverridesClearLayoutDirty(t *testing.T) {
	cases := []layoutDirtyClearCase{
		{"TabView", func() Node {
			tv := NewTabView("tv", 0, 0, 400, 200)
			tv.AddTab("A", NewLabel("a", "A", 0, 0, 0, 20))
			return tv
		}},
		{"Breadcrumbs", func() Node { return NewBreadcrumbs("bc", []string{"Home", "Page"}, 0, 0, 0, 0) }},
		{"TreeView", func() Node {
			root := NewTreeNode("root", "Root")
			root.AddChild(NewTreeNode("child", "Child"))
			return NewTreeView("tree", root, 0, 0, 200, 120)
		}},
		{"Rating", func() Node { return NewRating("rate", NewSignal(float32(3)), 5, 0, 0, 120, 0) }},
		{"Spinner", func() Node { return NewSpinner("spin", 0, 0, 48) }},
		{"Avatar", func() Node { return NewAvatar("av", "", "AB", 0, 0, 0, 0) }},
		{"NavigationRail", func() Node {
			return NewNavigationRail("rail", []BottomNavItem{{Icon: "⌂", Label: "Home"}}, NewSignal(0), 0, 0, 0, 400)
		}},
		{"ColorWell", func() Node {
			return NewColorWell("cw", rl.Blue, nil, 0, 0, 0, 0)
		}},
		{"Pagination", func() Node { return NewPagination("pg", 10, NewSignal(1), 0, 0, 0, 0) }},
		{"Timeline", func() Node {
			return NewTimeline("tl", []TimelineEvent{{Title: "One", Subtitle: "Detail"}}, 0, 0, 0, 0)
		}},
		{"FAB", func() Node {
			anchor := NewContainer("anchor", 0, 0, 200, 200)
			fab := NewFAB("fab", "+", "", func() {}, 0, 0, 56, 56)
			fab.Anchor = anchor
			anchor.AddChild(fab)
			return fab
		}},
		{"AppBar", func() Node {
			bar := NewAppBar("bar", "Title", 0, 0, 320, 0)
			bar.AddTrailing(NewIconButton("t", "x", "", 0, 0, 36, 36))
			return bar
		}},
		{"Carousel", func() Node {
			idx := NewSignal(0)
			return NewCarousel("c", []Node{NewLabel("s", "Slide", 0, 0, 0, 20)}, idx, 0, 0, 200, 120)
		}},
		{"Viewport", func() Node {
			vp := NewViewport("vp", 0, 0, 400, 300)
			vp.SetStyle("page-scroll")
			vp.AddChild(NewHeader("hdr", "Title", "Subtitle", 0, 0, 0, 0))
			return vp
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			n := tc.mk()
			if n == nil {
				t.Fatal("mk returned nil")
			}
			n.MarkDirty()
			if el, ok := n.(interface{ Layout() }); ok {
				el.Layout()
			} else {
				t.Fatal("node has no Layout")
			}
			if n.IsDirty() {
				t.Fatalf("%s still layout-dirty after Layout()", tc.name)
			}
		})
	}
}
