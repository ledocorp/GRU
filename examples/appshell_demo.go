//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"time"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &appShellScene{} }) }

// appShellScene demonstrates Strategy 2 Part 1 shell widgets: AppBar,
// BottomNavigationBar, FAB, Drawer, BottomSheet, Avatar, and ListTile content.
type appShellScene struct {
	BaseScene
	tab          *ui.Signal[int]
	pageTitle    *ui.Signal[string]
	bottom       *ui.BottomNavigationBar
	drawerRoot   *ui.Container
	sheetActions *ui.Container
	menuBtn      *ui.IconButton
	moreBtn      *ui.IconButton
}

func (s *appShellScene) Title() string { return "App Shell (Go)" }

func (s *appShellScene) Build(doc *ui.Document) {
	s.tab = ui.NewSignal(0)
	s.pageTitle = ui.NewSignal("Home")

	mount := MountAppShellRoot(doc, "appshell")
	shell := mount.Shell

	PreloadScenePhosphor(doc)
	s.menuBtn = AppBarLeadingMenu("appshell-menu")
	s.moreBtn = AppBarMenuButton("appshell-more")
	bar := ui.NewAppBar("appshell-bar", "github.com/ledocorp/gru", 0, 0, 0, 0)
	bar.SetLeading(s.menuBtn)
	bar.AddTrailing(s.moreBtn)
	s.moreBtn.OnClick = func() { s.showAppBarOverflowMenu() }

	vp := NewAppShellScrollViewport("appshell-vp")
	vp.Gap = 8

	MountSceneHeader(vp, "appshell-hdr", "App Shell (Go)",
		"Menu opens Drawer · overflow opens menu · FAB (+) opens BottomSheet · BottomNav switches tabs.")

	bottom := ui.NewBottomNavigationBar("appshell-nav", []ui.BottomNavItem{
		{Phosphor: ui.PhosphorHouse, Label: "Home"},
		{Phosphor: ui.PhosphorMagnifyingGlass, Label: "Search", Badge: "3"},
		{Phosphor: ui.PhosphorUser, Label: "Profile"},
	}, s.tab, 0, 0, 0, 0)
	bottom.OnChange = func(i int) {
		titles := []string{"Home", "Search", "Profile"}
		if i >= 0 && i < len(titles) {
			s.pageTitle.Set(titles[i])
		}
	}

	s.buildPageContent(vp)
	s.drawerRoot = s.buildDrawerContent()
	s.sheetActions = s.buildSheetActions()
	s.menuBtn.OnClick = func() { ui.OpenDrawer(s.drawerRoot) }

	fab := ui.NewFAB("appshell-fab", "+", "", nil, 0, 0, 0, 0)
	fab.SetPhosphorIcon(ui.PhosphorPlus, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorPlus, ui.PhosphorRegular)
	fab.Anchor = vp
	fab.OnClick = func() {
		ui.ShowBottomSheet(s.sheetActions, 0.38)
	}

	s.bottom = bottom

	shell.AddChild(bar)
	shell.AddChild(vp)
	shell.AddChild(bottom)
	doc.Root.AddChild(fab)

	s.pageTitle.Subscribe(func() {
		bar.TitleSig.Set(s.pageTitle.Get())
	})
	bar.TitleSig.Set(s.pageTitle.Get())
}

func (s *appShellScene) showAppBarOverflowMenu() {
	if s.moreBtn == nil {
		return
	}
	b := s.moreBtn.Bounds()
	ui.ShowContextMenu([]ui.ContextMenuItem{
		{Label: "Share link", Action: ui.CloseContextMenu},
		{Label: "Add to favorites", Action: ui.CloseContextMenu},
		{Divider: true},
		{Label: "Report issue", Action: ui.CloseContextMenu},
		{Label: "Help & support", Action: func() {
			ui.CloseContextMenu()
			ui.ShowToast("Help opened", ui.ToastInfo, 1500*time.Millisecond)
		}},
	}, b.X+b.Width, b.Y+b.Height)
}

func (s *appShellScene) selectTab(index int) {
	s.tab.Set(index)
	if s.bottom != nil && s.bottom.OnChange != nil {
		s.bottom.OnChange(index)
	}
}

func (s *appShellScene) drawerAction(label string, tab int) {
	if tab >= 0 {
		s.selectTab(tab)
	}
	ui.ShowToast("Menu · "+label, ui.ToastInfo, 1500*time.Millisecond)
	ui.CloseDrawer()
}

func (s *appShellScene) buildDrawerContent() *ui.Container {
	root := ui.NewContainer("appshell-drawer", 0, 0, 0, 0)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.Gap = 4
	root.SetStyle("transparent")

	hdr := ui.NewLabel("appshell-drawer-title", "Menu", 0, 0, 0, 0)
	hdr.SetStyle("form-label")
	root.AddChild(hdr)
	sub := ui.NewLabel("appshell-drawer-sub", "Switch tabs or open settings", 0, 0, 0, 0)
	sub.SetStyle("form-value")
	root.AddChild(sub)

	items := []struct {
		title, sub string
		tab        int
	}{
		{"Home", "Main feed", 0},
		{"Search", "Find content", 1},
		{"Profile", "Account settings", 2},
	}
	for _, it := range items {
		tile := ui.NewListTile("appshell-drawer-"+it.title, it.title, it.sub, 0, 0, 0, 0)
		tab := it.tab
		name := it.title
		tile.OnClick = func() { s.drawerAction(name, tab) }
		root.AddChild(tile)
	}

	sep := ui.NewSeparator("appshell-drawer-sep", "", 0, 0, 0, 0)
	root.AddChild(sep)

	settings := ui.NewListTile("appshell-drawer-settings", "Settings", "App preferences", 0, 0, 0, 0)
	settings.SetTrailing(ui.NewIcon("appshell-drawer-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 18, 18))
	settings.OnClick = func() { s.drawerAction("Settings", -1) }
	root.AddChild(settings)
	return root
}

func (s *appShellScene) buildSheetActions() *ui.Container {
	root := ui.NewContainer("appshell-sheet", 0, 0, 0, 0)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.Gap = 4
	root.SetStyle("transparent")

	header := ui.NewContainer("appshell-sheet-header", 0, 0, 0, 44)
	header.LayoutType = ui.LayoutFlex
	header.FlexDirection = ui.FlexRow
	header.Gap = 0
	header.SetStyle("transparent")
	back := AppBarBackButton("appshell-sheet-back")
	back.OnClick = func() { ui.CloseBottomSheet() }
	title := ui.NewLabel("appshell-sheet-title", "Quick actions", 0, 0, 0, 44)
	title.SetStyle("form-label")
	title.SetFlexGrow(1)
	spacer := ui.NewContainer("appshell-sheet-spacer", 0, 0, 44, 44)
	spacer.SetStyle("transparent")
	header.AddChild(back)
	header.AddChild(title)
	header.AddChild(spacer)
	root.AddChild(header)

	for _, label := range []string{"Share link", "Add to favorites", "Report issue", "Help & support"} {
		lbl := label
		tile := ui.NewListTile("appshell-sheet-"+lbl, lbl, "", 0, 0, 0, 0)
		tile.OnClick = func() { ui.CloseBottomSheet() }
		root.AddChild(tile)
	}
	return root
}

func (s *appShellScene) buildPageContent(vp *ui.Viewport) {
	profile := ui.NewListTile("appshell-profile", "User", "user@example.com", 0, 0, 0, 0)
	av := ui.NewAvatar("appshell-avatar", "", "U", 0, 0, 40, 40)
	av.ShowStatus = true
	av.StatusOnline = true
	profile.SetLeading(av)
	profile.SetTrailing(ui.NewIcon("appshell-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 18, 18))
	profile.OnClick = func() {
		ui.ShowToast("Profile row", ui.ToastInfo, 2*time.Second)
	}
	vp.AddChild(profile)

	wifi := ui.NewListTile("appshell-wifi", "Wi-Fi", "Connected", 0, 0, 0, 0)
	wifi.SetRowMode(ui.ListTileSwitchOnly)
	wifi.SetTrailing(ui.NewToggle("appshell-wifi-tog", true, 0, 0, 52, 28))
	vp.AddChild(wifi)

	for i := 1; i <= 6; i++ {
		tile := ui.NewListTile(
			fmt.Sprintf("appshell-item-%d", i),
			fmt.Sprintf("Item %d", i),
			"Tap to log",
			0, 0, 0, 0,
		)
		n := i
		tile.OnClick = func() {
			ui.ShowToast(fmt.Sprintf("Opened item %d", n), ui.ToastSuccess, 1500*time.Millisecond)
		}
		vp.AddChild(tile)
	}
}

func (s *appShellScene) OnUpdate(doc *ui.Document, _ float32) {
	if s.bottom != nil {
		UpdateShellFooterAutoHide(s.bottom, doc)
	}
	if ui.OverlayBlocksSceneInput() {
		return
	}

	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return
	}
	mouse := rl.GetMousePosition()
	if hit := ui.FindInteractiveAt(doc.Root, mouse); hit != nil {
		if _, ok := hit.(*ui.TextInput); ok {
			doc.SetFocus(hit)
			return
		}
	}
	doc.SetFocus(nil)
}
