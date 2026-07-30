package ui

import (
	"fmt"
	"testing"
)

// buildDesktopShellTree mirrors examples/shell_desktop_demo.go layout for regression tests.
func buildDesktopShellTree(t *testing.T, docW, docH float32) (*Container, *Viewport, *Card) {
	t.Helper()
	tab := NewSignal(0)
	root := NewContainer("root", 0, 0, docW, docH)
	root.LayoutType = LayoutAbsolute

	shell := NewContainer("desktop-shell", 0, 0, docW, docH)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn
	shell.SetStyle("transparent")

	menuBar := NewMenuBar("desktop-menubar", []MenuBarMenu{{Label: "File"}}, 0, 0, 0, 0)

	workspace := NewContainer("desktop-workspace", 0, 0, 0, 0)
	workspace.LayoutType = LayoutFlex
	workspace.FlexDirection = FlexRow
	workspace.SetStyle("transparent")
	workspace.SetFlexGrow(1)

	rail := NewNavigationRail("desktop-rail", []BottomNavItem{
		{Label: "Home"}, {Label: "Inbox"}, {Label: "Settings"},
	}, tab, 0, 0, 0, 0)

	main := NewContainer("desktop-main", 0, 0, 0, 0)
	main.LayoutType = LayoutFlex
	main.FlexDirection = FlexColumn
	main.SetStyle("transparent")
	main.SetFlexGrow(1)

	bar := NewAppBar("desktop-bar", "Home", 0, 0, 0, 0)

	body := NewContainer("desktop-body", 0, 0, 0, 0)
	body.SetStyle("appshell-content")
	body.SetFlexGrow(1)

	vp := NewViewport("desktop-vp", 0, 0, 0, 0)
	vp.SetStyle("transparent")
	vp.FlexDirection = FlexColumn
	vp.Gap = 8
	vp.SetFlexGrow(1)

	hdr := NewHeader("desktop-hdr", "Home", "MenuBar spans full width.", 0, 0, 0, 0)
	vp.AddChild(hdr)

	card := NewCard("desktop-card-0-0", "Panel 1", 0, 0, 0, 0)
	card.AutoHeight = true
	card.AddChild(NewLabel("desktop-card-lbl-0-0", "Sample content for Home.", 0, 0, 0, 0))
	vp.AddChild(card)

	body.AddChild(vp)
	main.AddChild(bar)
	main.AddChild(body)
	workspace.AddChild(rail)
	workspace.AddChild(main)

	statusBar := NewStatusBar("desktop-status", 0, 0, 0, 0)

	shell.AddChild(menuBar)
	shell.AddChild(workspace)
	shell.AddChild(statusBar)
	root.AddChild(shell)

	root.Layout()
	return shell, vp, card
}

func TestDesktopShellWorkspaceFillsShell(t *testing.T) {
	const w, h = float32(1280), float32(720)
	shell, _, _ := buildDesktopShellTree(t, w, h)
	ws := findNodeByID(shell, "desktop-workspace").(*Container)
	if ws.Bounds().Height < 400 {
		t.Fatalf("workspace height %.0f, want most of shell below menubar", ws.Bounds().Height)
	}
}

func TestDesktopShellMainFillsWorkspace(t *testing.T) {
	const w, h = float32(1280), float32(720)
	shell, _, _ := buildDesktopShellTree(t, w, h)
	ws := findNodeByID(shell, "desktop-workspace").(*Container)
	main := findNodeByID(shell, "desktop-main").(*Container)
	if main.Bounds().Height < ws.Bounds().Height-4 {
		t.Fatalf("main height %.0f, workspace %.0f — main should fill row cross-axis", main.Bounds().Height, ws.Bounds().Height)
	}
	if main.Bounds().Width < w-navigationRailWidth-100 {
		t.Fatalf("main width %.0f too narrow", main.Bounds().Width)
	}
}

func TestDesktopShellViewportFillsBody(t *testing.T) {
	const w, h = float32(1280), float32(720)
	_, vp, _ := buildDesktopShellTree(t, w, h)
	body := findNodeByID(vp.ParentNode(), "desktop-body")
	if body == nil {
		// walk from vp
		p := vp.ParentNode()
		for p != nil {
			if p.ID() == "desktop-body" {
				body = p
				break
			}
			p = p.ParentNode()
		}
	}
	if vp.Bounds().Height < 200 {
		t.Fatalf("viewport height %.0f, want flex-grow fill in main column", vp.Bounds().Height)
	}
}

func TestDesktopShellCardNotTiny(t *testing.T) {
	const w, h = float32(1280), float32(720)
	_, _, card := buildDesktopShellTree(t, w, h)
	if card.Bounds().Height < 60 {
		t.Fatalf("desktop card height %.0f, want intrinsic >= 60", card.Bounds().Height)
	}
}

func TestFlexRowStretchesWhenRowHasFlexGrow(t *testing.T) {
	row := NewContainer("row", 0, 0, 400, 300)
	row.LayoutType = LayoutFlex
	row.FlexDirection = FlexRow
	row.SetFlexGrow(1)

	fixed := NewContainer("fixed", 0, 0, 80, 0)
	fixed.AutoHeight = false

	main := NewContainer("main", 0, 0, 0, 0)
	main.SetFlexGrow(1)

	row.AddChild(fixed)
	row.AddChild(main)
	row.Layout()

	if main.Bounds().Height < 280 {
		t.Fatalf("main height %.0f, want cross-axis stretch in flex-grow row", main.Bounds().Height)
	}
}

func TestViewportScrollWithFlexGrowGrid(t *testing.T) {
	const w, h = float32(800), float32(600)
	shell := NewContainer("shell", 0, 0, w, h)
	shell.LayoutType = LayoutFlex
	shell.FlexDirection = FlexColumn

	vp := NewViewport("vp", 0, 0, 0, 0)
	vp.SetFlexGrow(1)
	vp.Gap = 12

	hdr := NewHeader("hdr", "Title", "Subtitle", 0, 0, 0, 64)

	grid := NewContainer("grid", 0, 0, 0, 0)
	grid.LayoutType = LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	grid.SetFlexGrow(1)
	if !grid.AutoHeight {
		t.Fatal("grid with h=0 should keep AutoHeight for intrinsic snap")
	}

	for i := 0; i < 4; i++ {
		p := NewPanel(fmt.Sprintf("p%d", i), fmt.Sprintf("Panel %d", i), 0, 0, 0, 180)
		p.SetColSpan(BreakpointXS, 12)
		grid.AddChild(p)
	}

	vp.AddChild(hdr)
	vp.AddChild(grid)
	shell.AddChild(vp)

	root := NewContainer("root", 0, 0, w, h)
	root.LayoutType = LayoutAbsolute
	root.AddChild(shell)
	root.Layout()

	if overflow := vp.overflowScrollY(); overflow < 100 {
		t.Fatalf("overflowScrollY %.0f, want scroll when flex-grow grid content exceeds viewport", overflow)
	}
}

func TestDesktopShellChromeOrder(t *testing.T) {
	const w, h = float32(1280), float32(720)
	shell, _, _ := buildDesktopShellTree(t, w, h)
	kids := shell.Children()
	if len(kids) != 3 {
		t.Fatalf("shell children = %d, want menubar/workspace/statusbar", len(kids))
	}
	if kids[0].ID() != "desktop-menubar" || kids[2].ID() != "desktop-status" {
		t.Fatalf("shell order: %s, %s, %s", kids[0].ID(), kids[1].ID(), kids[2].ID())
	}
	mb := kids[0].Bounds()
	sb := kids[2].Bounds()
	if mb.Width < w-20 {
		t.Fatalf("menubar width %.0f", mb.Width)
	}
	if sb.Y <= mb.Y {
		t.Fatalf("statusbar should be below menubar")
	}
}
