package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestFlexRowKeepsNavigationRailWidthOnNarrowParent(t *testing.T) {
	row := NewContainer("row", 0, 0, 400, 300)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.SetBounds(rl.NewRectangle(0, 0, 400, 300))

	tab := NewSignal(0)
	rail := NewNavigationRail("rail", []BottomNavItem{
		{Icon: "A", Label: "Home"},
	}, tab, 0, 0, 0, 0)
	main := NewContainer("main", 0, 0, 0, 0)
	main.SetFlexGrow(1)

	row.AddChild(rail)
	row.AddChild(main)
	row.MarkDirty()
	row.Layout()

	if got := rail.Bounds().Width; got != navigationRailWidth {
		t.Fatalf("wide parent: rail width %.0f, want %.0f", got, navigationRailWidth)
	}

	row.SetBounds(rl.NewRectangle(0, 0, 120, 300))
	row.InvalidateLayoutPassCache()
	row.MarkDirty()
	row.Layout()

	if got := rail.Bounds().Width; got != navigationRailWidth {
		t.Fatalf("narrow parent: rail width %.0f, want %.0f", got, navigationRailWidth)
	}
}

func TestDesktopPageShellWorkspaceRowPreservesRailWidth(t *testing.T) {
	doc := NewDocument(800, 600)
	shell := NewContainer("shell", 0, 0, 800, 600)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn
	workspace := NewContainer("ws", 0, 0, 0, 0)
	workspace.LayoutType = LayoutFlex
	workspace.FlexDirection = FlexRow
	workspace.SetFlexGrow(1)
	tab := NewSignal(0)
	rail := NewNavigationRail("rail", []BottomNavItem{{Label: "Home"}}, tab, 0, 0, 0, 0)
	main := NewContainer("main", 0, 0, 0, 0)
	main.SetFlexGrow(1)
	workspace.AddChild(rail)
	workspace.AddChild(main)
	shell.AddChild(workspace)
	doc.Root.AddChild(shell)
	doc.Root.MarkDirty()
	doc.Root.Layout()

	if got := rail.Bounds().Width; got != navigationRailWidth {
		t.Fatalf("desktop shell row: rail width %.0f, want %.0f", got, navigationRailWidth)
	}
}

func TestDocumentResizePreservesDesktopRail(t *testing.T) {
	doc := NewDocument(1280, 720)
	doc.SetChromeTop(40)
	w, h := float32(1280), float32(680)
	shell := NewContainer("shell", 0, 0, w, h)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn
	workspace := NewContainer("ws", 0, 0, 0, 0)
	workspace.LayoutType = LayoutFlex
	workspace.FlexDirection = FlexRow
	workspace.SetFlexGrow(1)
	tab := NewSignal(0)
	rail := NewNavigationRail("rail", []BottomNavItem{{Label: "Home"}, {Label: "Settings"}}, tab, 0, 0, 0, 0)
	main := NewContainer("main", 0, 0, 0, 0)
	main.SetFlexGrow(1)
	workspace.AddChild(rail)
	workspace.AddChild(main)
	shell.AddChild(workspace)
	doc.Root.AddChild(shell)
	doc.Height = int32(h)
	doc.Width = 1280

	for _, nw := range []int32{1280, 960, 720, 640, 900, 1280} {
		doc.Resize(nw, 720)
		if got := rail.Bounds().Width; got != navigationRailWidth {
			t.Fatalf("after Resize(%d): rail width %.0f, want %.0f", nw, got, navigationRailWidth)
		}
		if got := rail.Bounds().Height; got < 200 {
			t.Fatalf("after Resize(%d): rail height %.0f too small", nw, got)
		}
	}
}
