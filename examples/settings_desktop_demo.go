//go:build !notepad

// Package examples — desktop settings under MenuBar (no AppBar). Reference for Notepad-style in-shell settings.
package examples

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &settingsDesktopScene{} }) }

type settingsDesktopScene struct {
	BaseScene
	inSettings   bool
	workArea     *ui.Container
	settingsPage *ui.Container
	placeholder  *ui.Label
	setWrapTog   *ui.Toggle
	setStatusTog *ui.Toggle
	setSyntaxTog *ui.Toggle
	setDenseTog  *ui.Toggle
	themePick    *ui.Signal[string]
	accentPick   *ui.Signal[string]
	wordWrap     bool
	showStatus   bool
	syntaxOn     bool
	denseLists   bool
}

func (s *settingsDesktopScene) Title() string { return "Settings · Desktop (Go)" }

func (s *settingsDesktopScene) Build(doc *ui.Document) {
	PreloadScenePhosphor(doc)
	ui.Phosphor.EnsureLoaded(ui.PhosphorTextWrap, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorInfoI, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorCodeView, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorSun, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorStar, ui.PhosphorRegular)
	ui.Phosphor.EnsureLoaded(ui.PhosphorMoon, ui.PhosphorRegular)

	s.wordWrap = true
	s.showStatus = true
	s.syntaxOn = false
	s.denseLists = false
	s.themePick = ui.NewSignal("System")
	s.accentPick = ui.NewSignal("Indigo")

	root := MountEdgeToEdgeRoot(doc, "setdesk", false)
	root.Gap = 0

	menuBar := ui.NewMenuBar("setdesk-menubar", []ui.MenuBarMenu{
		{Label: "File", ItemsFunc: s.fileMenuItems},
		{Label: "Help", Items: []ui.ContextMenuItem{
			{Label: "About this demo", Action: func() {
				ui.ShowToast("Desktop settings — CP-SETTINGS-DESKTOP-01", ui.ToastInfo, 2500)
			}},
		}},
	}, 0, 0, 0, 0)
	menuBar.SetFlexGrow(0)

	body := ui.NewContainer("setdesk-body", 0, 0, 0, 0)
	body.LayoutType = ui.LayoutFlex
	body.FlexDirection = ui.FlexColumn
	body.SetStyle("settings-body")
	body.SetFlexGrow(1)

	s.workArea = ui.NewContainer("setdesk-work", 0, 0, 0, 0)
	s.workArea.LayoutType = ui.LayoutFlex
	s.workArea.FlexDirection = ui.FlexColumn
	s.workArea.SetStyle("settings-body")
	s.workArea.SetFlexGrow(1)

	pad := ui.NewContainer("setdesk-work-pad", 0, 0, 0, 0)
	pad.LayoutType = ui.LayoutFlex
	pad.FlexDirection = ui.FlexColumn
	pad.SetStyle("page-scroll")
	pad.SetFlexGrow(1)
	pad.Gap = 8

	s.placeholder = ui.NewLabel("setdesk-placeholder",
		"Work area — open File → Settings for the desktop settings page.\n\n"+
			"MenuBar stays visible; there is no AppBar (edge-to-edge shell).",
		0, 0, 0, 0)
	s.placeholder.SetStyle("form-value")
	s.placeholder.Align = ui.LabelAlignLeft
	s.placeholder.Wrap = true
	pad.AddChild(s.placeholder)
	s.workArea.AddChild(pad)

	s.settingsPage = s.buildDesktopSettingsPage()

	body.AddChild(s.workArea)
	body.AddChild(s.settingsPage)

	status := ui.NewStatusBar("setdesk-status", 0, 0, 0, 0)
	hint := ui.NewLabel("setdesk-status-hint", "File → Settings · Escape to close", 0, 0, 0, 0)
	hint.SetStyle("statusbar-label")
	hint.Align = ui.LabelAlignLeft
	status.SetColumns([]ui.StatusBarColumn{
		{Weight: 100, Align: ui.LabelAlignLeft, Nodes: []ui.Node{hint}},
	})

	root.AddChild(menuBar)
	root.AddChild(body)
	root.AddChild(status)
	FinishShellMount(doc)
}

func (s *settingsDesktopScene) fileMenuItems() []ui.ContextMenuItem {
	items := []ui.ContextMenuItem{
		{Label: "Settings…", Action: func() { s.showSettingsPage() }},
	}
	if s.inSettings {
		items = append(items, ui.ContextMenuItem{Divider: true})
		items = append(items, ui.ContextMenuItem{
			Label: "Back to main",
			Action: func() { s.closeSettingsPage() },
		})
	}
	return items
}

func (s *settingsDesktopScene) buildDesktopSettingsPage() *ui.Container {
	page := ui.NewContainer("setdesk-settings-page", 0, 0, 0, 0)
	page.LayoutType = ui.LayoutFlex
	page.FlexDirection = ui.FlexColumn
	page.SetStyle("settings-flat-band")
	page.SetFlexGrow(1)

	vp := ui.NewViewport("setdesk-settings-vp", 0, 0, 0, 0)
	vp.SetStyle("settings-scroll")
	vp.FlexDirection = ui.FlexColumn
	vp.SetFlexGrow(1)
	vp.Gap = 24

	vp.AddChild(settingsPageTitleRow("setdesk-page-title", "Settings", func() {
		s.closeSettingsPage()
	}))

	intro := ui.NewLabel("setdesk-settings-intro",
		"Flat grey in-shell settings: section labels + ListTile rows (no Panel chrome).",
		0, 0, 0, 0)
	intro.SetStyle("form-value")
	intro.Align = ui.LabelAlignLeft
	intro.Wrap = true
	vp.AddChild(intro)

	s.setWrapTog = ui.NewToggle("setdesk-wrap", s.wordWrap, 0, 0, 52, 28)
	s.setStatusTog = ui.NewToggle("setdesk-status-tog", s.showStatus, 0, 0, 52, 28)
	s.setSyntaxTog = ui.NewToggle("setdesk-syntax", s.syntaxOn, 0, 0, 52, 28)
	s.setDenseTog = ui.NewToggle("setdesk-dense", s.denseLists, 0, 0, 52, 28)

	vp.AddChild(settingsSection("setdesk-sec-editor", "Editor"))
	vp.AddChild(settingsFlatToggleRow("setdesk-wrap-row", "Word wrap", "Wrap long lines in the editor", ui.PhosphorTextWrap, s.setWrapTog))

	vp.AddChild(settingsSection("setdesk-sec-view", "View"))
	vp.AddChild(settingsFlatToggleRow("setdesk-status-row", "Status bar", "Show line, column, and encoding info", ui.PhosphorInfoI, s.setStatusTog))
	vp.AddChild(settingsFlatToggleRow("setdesk-dense-row", "Dense list rows", "Compact ListTile height in side panes", ui.PhosphorList, s.setDenseTog))

	vp.AddChild(settingsSection("setdesk-sec-syntax", "Syntax"))
	vp.AddChild(settingsFlatToggleRow("setdesk-syntax-row", "Syntax highlight", "Colorize source in the editor", ui.PhosphorCodeView, s.setSyntaxTog))

	vp.AddChild(settingsSection("setdesk-sec-appearance", "Appearance"))

	motionTog := ui.NewToggle("setdesk-motion", false, 0, 0, 52, 28)
	vp.AddChild(settingsFlatToggleRow("setdesk-motion-row", "Reduce motion", "Fewer transitions and animations", ui.PhosphorMoon, motionTog))

	themeCB := ui.NewComboBox("setdesk-theme-cb",
		[]string{"System", "Light", "Dark"}, s.themePick, 0, 0, 168, 40)
	vp.AddChild(settingsFlatComboRow("setdesk-theme-row", "Theme", "Match system or force light/dark", ui.PhosphorSun, themeCB))

	vp.AddChild(settingsSection("setdesk-sec-accent", "Accent"))

	accentCB := ui.NewComboBox("setdesk-accent-cb",
		[]string{"Indigo", "Blue", "Teal", "Rose"}, s.accentPick, 0, 0, 168, 40)
	vp.AddChild(settingsFlatComboRow("setdesk-accent-row", "Accent color", "Buttons, links, and focus ring", ui.PhosphorStar, accentCB))

	s.wireDesktopSettingsToggles()
	s.themePick.Subscribe(func() {
		ui.ShowToast("Theme: "+s.themePick.Get(), ui.ToastInfo, 1500)
	})
	s.accentPick.Subscribe(func() {
		ui.ShowToast("Accent: "+s.accentPick.Get(), ui.ToastInfo, 1500)
	})

	page.AddChild(vp)
	page.Hide()
	return page
}

func (s *settingsDesktopScene) wireDesktopSettingsToggles() {
	apply := func() {
		if s.setWrapTog == nil {
			return
		}
		s.wordWrap = s.setWrapTog.Value.Get()
		s.showStatus = s.setStatusTog.Value.Get()
		s.syntaxOn = s.setSyntaxTog.Value.Get()
		s.denseLists = s.setDenseTog.Value.Get()
		ui.ShowToast("Settings updated", ui.ToastInfo, 1200)
	}
	s.setWrapTog.Value.Subscribe(func() { apply() })
	s.setStatusTog.Value.Subscribe(func() { apply() })
	s.setSyntaxTog.Value.Subscribe(func() { apply() })
	s.setDenseTog.Value.Subscribe(func() { apply() })
}

func (s *settingsDesktopScene) syncDesktopSettingsToggles() {
	if s.setWrapTog == nil {
		return
	}
	s.setWrapTog.Value.Set(s.wordWrap)
	s.setStatusTog.Value.Set(s.showStatus)
	s.setSyntaxTog.Value.Set(s.syntaxOn)
	s.setDenseTog.Value.Set(s.denseLists)
}

func (s *settingsDesktopScene) showSettingsPage() {
	if s.settingsPage == nil {
		return
	}
	ui.CloseContextMenu()
	s.syncDesktopSettingsToggles()
	s.inSettings = true
	s.workArea.Hide()
	s.settingsPage.Show()
	if s.settingsPage.Parent() != nil {
		s.settingsPage.Parent().MarkDirty()
	}
}

func (s *settingsDesktopScene) closeSettingsPage() {
	if s.settingsPage == nil {
		return
	}
	s.inSettings = false
	s.settingsPage.Hide()
	s.workArea.Show()
	if s.workArea.Parent() != nil {
		s.workArea.Parent().MarkDirty()
	}
}

func (s *settingsDesktopScene) OnUpdate(_ *ui.Document, _ float32) {
	if !s.inSettings {
		return
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		s.closeSettingsPage()
	}
}
