//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"time"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &shellDesktopScene{} }) }

// shellDesktopScene demonstrates NavigationRail, MenuBar, AppBar, and StatusBar
// using MountDesktopPageShell (ARCHITECTURE §5.6 / §11.4 resize contract).
//
//	Root → shell (flex column) → MenuBar (full width) → workspace row → rail | main → StatusBar (full width)
type shellDesktopScene struct {
	BaseScene
	tab         *ui.Signal[int]
	pageTitle   *ui.Signal[string]
	statusMode  *ui.Label
	statusReady *ui.Label
	vp          *ui.Viewport
}

func (s *shellDesktopScene) Title() string { return "Desktop Shell (Go)" }

func (s *shellDesktopScene) Build(doc *ui.Document) {
	s.tab = ui.NewSignal(0)
	s.pageTitle = ui.NewSignal("Home")

	shell, workspace := MountDesktopPageShell(doc, "desktop")

	PreloadScenePhosphor(doc)

	rail := ui.NewNavigationRail("desktop-rail", []ui.BottomNavItem{
		{Phosphor: ui.PhosphorHouse, Label: "Home"},
		{Phosphor: ui.PhosphorEnvelope, Label: "Inbox"},
		{Phosphor: ui.PhosphorGear, Label: "Settings"},
	}, s.tab, 0, 0, 0, 0)

	main := ui.NewContainer("desktop-main", 0, 0, 0, 0)
	main.LayoutType = ui.LayoutFlex
	main.FlexDirection = ui.FlexColumn
	main.SetStyle("transparent")
	main.SetFlexGrow(1)

	menuBar := ui.NewMenuBar("desktop-menubar", []ui.MenuBarMenu{
		{Label: "File", Items: []ui.ContextMenuItem{
			{Label: "New", Action: func() {
				ui.ShowToast("File · New", ui.ToastInfo, 1500*time.Millisecond)
			}},
			{Label: "Open Demo Directory", Action: func() {
				if NavigateToScene != nil {
					NavigateToScene(DirectorySceneTitle)
				}
			}},
			{Divider: true},
			{Label: "Exit", Action: func() {
				ui.ShowToast("Exit (demo only)", ui.ToastInfo, 1500*time.Millisecond)
			}},
		}},
		{Label: "Edit", Items: []ui.ContextMenuItem{
			{Label: "Undo", Disabled: true},
			{Label: "Redo", Disabled: true},
			{Divider: true},
			{Label: "Preferences", Action: func() {
				ui.ShowToast("Edit · Preferences", ui.ToastInfo, 1500*time.Millisecond)
			}},
		}},
		{Label: "View", Items: []ui.ContextMenuItem{
			{Label: "Command palette", Action: func() {
				ui.ShowCommandPaletteFromChord(s.commandItems())
				ui.CloseContextMenu()
			}},
			{Label: "Toggle perf overlay (F11)", Action: func() {
				ui.ShowPerfOverlay = !ui.ShowPerfOverlay
				ui.CloseContextMenu()
			}},
			{Label: "Help", Action: func() {
				ui.ShowToast("View · Help", ui.ToastInfo, 1500*time.Millisecond)
				ui.CloseContextMenu()
			}},
		}},
	}, 0, 0, 0, 0)

	bar := ui.NewAppBar("desktop-bar", "Home", 0, 0, 0, 0)
	bar.AddTrailing(AppBarMenuButton("desktop-menu"))

	vp := ui.NewViewport("desktop-vp", 0, 0, 0, 0)
	vp.SetStyle("transparent")
	vp.FlexDirection = ui.FlexColumn
	vp.Gap = 8
	vp.SetFlexGrow(1)

	s.statusReady = ui.NewLabel("desktop-status-ready", "Ready", 0, 0, 0, 0)
	s.statusReady.SetStyle("form-value")
	s.statusMode = ui.NewLabel("desktop-status-mode", "View: Home", 0, 0, 0, 0)
	s.statusMode.SetStyle("form-value")
	statusCenter := ui.NewLabel("desktop-status-center", "Desktop Shell (Go)", 0, 0, 0, 0)
	statusCenter.SetStyle("form-value")
	statusRight := ui.NewLabel("desktop-status-pos", "Ln 1, Col 1", 0, 0, 0, 0)
	statusRight.SetStyle("form-value")

	statusBar := ui.NewStatusBar("desktop-status", 0, 0, 0, 0)
	statusBar.AddLeft(s.statusReady)
	statusBar.AddLeft(s.statusMode)
	statusBar.AddCenter(statusCenter)
	statusBar.AddRight(statusRight)

	main.AddChild(bar)
	main.AddChild(vp)

	workspace.AddChild(rail)
	workspace.AddChild(main)

	shell.RemoveChild("desktop-workspace")
	shell.AddChild(menuBar)
	shell.AddChild(workspace)
	shell.AddChild(statusBar)
	shell.MarkDirty()

	s.vp = vp

	s.buildPage(vp, s.tab.Get())

	rail.OnChange = func(i int) { s.loadTab(i) }

	s.pageTitle.Subscribe(func() {
		bar.TitleSig.Set(s.pageTitle.Get())
	})
	bar.TitleSig.Set(s.pageTitle.Get())

	back := AppBarBackButton("desktop-back")
	back.OnClick = func() {
		if NavigateToScene != nil {
			NavigateToScene(DirectorySceneTitle)
		}
	}
	bar.SetLeading(back)

	FinishShellMount(doc)
}

func (s *shellDesktopScene) buildPage(vp *ui.Viewport, index int) {
	titles := []string{"Home", "Inbox", "Settings"}
	hints := []string{
		"MenuBar spans full width · Ctrl+Shift+P opens Command Palette · StatusBar at bottom.",
		"Jump to Inbox + Table from the Demo Directory to see DataTable composition.",
		"Open Settings (Go) from the directory for ListTile + Rating patterns.",
	}
	title := titles[0]
	hint := hints[0]
	if index >= 0 && index < len(titles) {
		title = titles[index]
		hint = hints[index]
	}

	hdr := ui.NewHeader("desktop-hdr", title, hint, 0, 0, 0, 0)
	hdr.SetStyle("header")
	vp.AddChild(hdr)

	if index == 0 {
		link := ui.NewListTile("desktop-link-timeline", "Open Timeline (Go)", "Activity feed demo", 0, 0, 0, 0)
		link.SetTrailing(ui.NewIcon("desktop-link-timeline-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 24, 24))
		link.OnClick = func() {
			if NavigateToScene != nil {
				NavigateToScene("Timeline (Go)")
			}
		}
		vp.AddChild(link)
	}

	if index == 1 {
		link := ui.NewListTile("desktop-link-inbox", "Open Inbox + Table (Go)", "Full table demo", 0, 0, 0, 0)
		link.SetTrailing(ui.NewIcon("desktop-link-inbox-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 24, 24))
		link.OnClick = func() {
			if NavigateToScene != nil {
				NavigateToScene("Inbox + Table (Go)")
			}
		}
		vp.AddChild(link)
	}

	for i := 0; i < 4; i++ {
		card := ui.NewCard(fmt.Sprintf("desktop-card-%d-%d", index, i), fmt.Sprintf("Panel %d", i+1), 0, 0, 0, 0)
		card.AutoHeight = true
		lbl := ui.NewLabel(fmt.Sprintf("desktop-card-lbl-%d-%d", index, i),
			"Sample content for "+title+".", 0, 0, 0, 0)
		lbl.SetStyle("form-value")
		lbl.Align = ui.LabelAlignLeft
		lbl.Wrap = true
		card.AddChild(lbl)
		vp.AddChild(card)
	}
}

func (s *shellDesktopScene) loadTab(i int) {
	titles := []string{"Home", "Inbox", "Settings"}
	if i >= 0 && i < len(titles) {
		if s.tab != nil {
			s.tab.Set(i)
		}
		s.pageTitle.Set(titles[i])
		if s.statusMode != nil {
			s.statusMode.Text.Set("View: " + titles[i])
		}
	}
	if s.vp == nil {
		return
	}
	for len(s.vp.Children()) > 0 {
		s.vp.RemoveChild(s.vp.Children()[0].ID())
	}
	s.buildPage(s.vp, i)
	s.vp.MarkDirty()
}

func (s *shellDesktopScene) commandItems() []ui.CommandPaletteItem {
	return []ui.CommandPaletteItem{
		{
			Label: "View: Home", Subtitle: "Desktop Shell", Keywords: "rail tab",
			Action: func() { s.loadTab(0) },
		},
		{
			Label: "View: Inbox", Subtitle: "Desktop Shell", Keywords: "rail tab mail",
			Action: func() { s.loadTab(1) },
		},
		{
			Label: "View: Settings", Subtitle: "Desktop Shell", Keywords: "rail tab gear",
			Action: func() { s.loadTab(2) },
		},
		{
			Label: "Toggle perf overlay", Subtitle: "View", Keywords: "f11 debug fps",
			Shortcut: "F11",
			Action: func() { ui.ShowPerfOverlay = !ui.ShowPerfOverlay },
		},
		{
			Label: "Go to Timeline (Go)", Subtitle: "Navigation", Keywords: "activity history log",
			Action: func() {
				if NavigateToScene != nil {
					NavigateToScene("Timeline (Go)")
				}
			},
		},
		{
			Label: "Go to App Shell (Go)", Subtitle: "Navigation", Keywords: "mobile drawer fab",
			Action: func() {
				if NavigateToScene != nil {
					NavigateToScene("App Shell (Go)")
				}
			},
		},
		{
			Label: "Go to Demo Directory", Subtitle: "Navigation", Keywords: "launcher home",
			Action: func() {
				if NavigateToScene != nil {
					NavigateToScene(DirectorySceneTitle)
				}
			},
		},
	}
}

func (s *shellDesktopScene) OnUpdate(doc *ui.Document, dt float32) {
	_ = doc
	_ = dt
	if ui.KeyChordCtrlShiftP() {
		if ui.IsCommandPaletteOpen() {
			ui.CloseCommandPalette()
		} else {
			ui.ShowCommandPaletteFromChord(s.commandItems())
		}
	}
	if rl.IsKeyPressed(rl.KeyF1) {
		ui.ShowToast("F1 — Ctrl+Shift+P command palette · Demo Directory in footer", ui.ToastInfo, 2*time.Second)
	}
}
