//go:build !notepad

// Package examples (continued)
package examples

import (
	"time"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &settingsScene{} }) }

// settingsScene is a Go-first settings flow: AppBar, sections, ListTiles,
// filter badges (chips), and Rating — complements App Shell (Go).
type settingsScene struct {
	BaseScene
	filterGo   *ui.Signal[bool]
	filterBeta *ui.Signal[bool]
	rating     *ui.Signal[float32]
}

func (s *settingsScene) Title() string { return "Settings (Go)" }

func (s *settingsScene) Build(doc *ui.Document) {
	s.filterGo = ui.NewSignal(true)
	s.filterBeta = ui.NewSignal(false)
	s.rating = ui.NewSignal(float32(4))

	mount := MountAppShellRoot(doc, "settings")
	shell := mount.Shell
	PreloadScenePhosphor(doc)

	backBtn := AppBarBackButton("settings-back")
	backBtn.OnClick = func() {
		ui.ShowToast("Back", ui.ToastInfo, 1500*time.Millisecond)
	}
	menuBtn := AppBarMenuButton("settings-menu")
	bar := ui.NewAppBar("settings-bar", "Settings", 0, 0, 0, 0)
	bar.SetLeading(backBtn)
	bar.AddTrailing(menuBtn)

	vp := NewAppShellScrollViewport("settings-vp")
	sub := ui.NewLabel("settings-sub", "ListTiles, filter badges, and Rating.", 0, 0, 0, 0)
	sub.SetStyle("form-value")
	sub.Align = ui.LabelAlignLeft
	sub.Wrap = true
	vp.AddChild(sub)

	s.buildSettingsContent(vp)

	shell.AddChild(bar)
	shell.AddChild(vp)
}

func (s *settingsScene) buildSettingsContent(vp *ui.Viewport) {
	vp.AddChild(settingsSection("settings-sec-account", "Account"))
	profile := ui.NewListTile("settings-profile", "User", "user@example.com", 0, 0, 0, 0)
	av := ui.NewAvatar("settings-avatar", "", "U", 0, 0, 40, 40)
	profile.SetLeading(av)
	profile.SetTrailing(ui.NewIcon("settings-prof-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 24, 24))
	profile.OnClick = func() {
		ui.ShowToast("Open profile", ui.ToastInfo, 2*time.Second)
	}
	vp.AddChild(profile)

	vp.AddChild(settingsSection("settings-sec-prefs", "Preferences"))
	wifi := ui.NewListTile("settings-wifi", "Wi-Fi", "Connected", 0, 0, 0, 0)
	wifi.SetTrailing(ui.NewToggle("settings-wifi-tog", true, 0, 0, 52, 28))
	vp.AddChild(wifi)

	notif := ui.NewListTile("settings-notif", "Notifications", "Push and email", 0, 0, 0, 0)
	notif.SetTrailing(ui.NewToggle("settings-notif-tog", true, 0, 0, 52, 28))
	vp.AddChild(notif)

	theme := ui.NewListTile("settings-theme", "Theme", "System", 0, 0, 0, 0)
	theme.SetTrailing(ui.NewIcon("settings-theme-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 24, 24))
	theme.OnClick = func() { ui.ShowToast("Theme picker", ui.ToastInfo, 2*time.Second) }
	vp.AddChild(theme)

	vp.AddChild(settingsSection("settings-sec-filters", "Filters"))
	chipRow := ui.NewContainer("settings-chips", 0, 0, 0, 0)
	chipRow.LayoutType = ui.LayoutFlex
	chipRow.FlexDirection = ui.FlexRow
	chipRow.Gap = 10
	chipRow.SetStyle("settings-chip-row")
	badgeGo := ui.NewBadge("settings-badge-go", "Go", ui.BadgePrimary, 0, 0, 0, 32)
	badgeGo.Selected = s.filterGo
	badgeBeta := ui.NewBadge("settings-badge-beta", "Beta", ui.BadgeWarning, 0, 0, 0, 32)
	badgeBeta.Selected = s.filterBeta
	badgeBeta.CloseButton = true
	badgeBeta.OnClose = func() {
		s.filterBeta.Set(false)
		badgeBeta.Hide()
		ui.ShowToast("Beta filter removed", ui.ToastInfo, 2*time.Second)
	}
	chipRow.AddChild(badgeGo)
	chipRow.AddChild(badgeBeta)
	vp.AddChild(chipRow)

	vp.AddChild(settingsSection("settings-sec-feedback", "Feedback"))
	feedbackRow := ui.NewContainer("settings-rating-row", 0, 0, 0, 0)
	feedbackRow.LayoutType = ui.LayoutFlex
	feedbackRow.FlexDirection = ui.FlexRow
	feedbackRow.Gap = 10
	feedbackRow.SetStyle("settings-chip-row")
	lbl := ui.NewLabel("settings-rating-lbl", "Rate Gru", 0, 0, 0, 0)
	lbl.SetStyle("settings-row-label")
	lbl.Align = ui.LabelAlignLeft
	lbl.PreferredWidth = 96
	feedbackRow.AddChild(lbl)
	feedbackRow.AddChild(ui.NewRating("settings-rating", s.rating, 5, 0, 0, 0, 0))
	vp.AddChild(feedbackRow)

	vp.AddChild(settingsSection("settings-sec-about", "About"))
	about := ui.NewListTile("settings-about", "Version", "0.9 dev", 0, 0, 0, 0)
	about.SetTrailing(ui.NewIcon("settings-about-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 24, 24))
	about.OnClick = func() { ui.ShowToast("About Gru", ui.ToastInfo, 2*time.Second) }
	vp.AddChild(about)

	legal := ui.NewListTile("settings-legal", "Privacy", "Policy and terms", 0, 0, 0, 0)
	legal.SetTrailing(ui.NewIcon("settings-legal-chev", ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, 24, 24))
	legal.OnClick = func() { ui.ShowToast("Legal", ui.ToastInfo, 2*time.Second) }
	vp.AddChild(legal)
}

func (s *settingsScene) OnUpdate(doc *ui.Document, _ float32) {
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
