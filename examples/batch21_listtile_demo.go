//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"strings"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch21ListTileScene{} }) }

type batch21ListTileScene struct {
	BaseScene
}

func (s *batch21ListTileScene) Title() string { return "Batch 21 · ListTile" }

func (s *batch21ListTileScene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5TilePanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func tileChevron(id string, size float32) *ui.Icon {
	return ui.NewIcon(id, ui.PhosphorCaretRight, ui.PhosphorRegular, 0, 0, size, size)
}

func (s *batch21ListTileScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b21",
		"Widget Batch 21 · ListTile",
		"Settings-style rows — navigation taps, switch-only rows, leading badges.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("b21-grid", 12)

	pNav := ui.NewPanel("b21-nav", "Navigation rows", 0, 0, 0, 0)
	pNav.AutoHeight = true
	setSpans5TilePanel(pNav, 12, 12, 6, 6, 6)
	pNav.Gap = 4
	pNav.TitleHeight = 32

	account := ui.NewListTile("b21-account", "Account", "user@example.com", 0, 0, 0, 0)
	account.SetTrailing(tileChevron("b21-chev", 18))
	account.OnClick = func() {
		account.Subtitle = "Opened account settings"
		account.MarkDrawDirty()
	}

	billing := ui.NewListTile("b21-billing", "Billing", "Visa ending 4242", 0, 0, 0, 0)
	badge := ui.NewBadge("b21-badge", "Pro", ui.BadgePrimary, 0, 0, 0, 0)
	billing.SetLeading(badge)
	billing.SetTrailing(tileChevron("b21-bill-chev", 18))

	pNav.AddChild(batchCaption("b21-nav-cap", "Full-row tap target with optional leading/trailing slots."))
	pNav.AddChild(account)
	pNav.AddChild(billing)

	pSwitch := ui.NewPanel("b21-switch", "Switch rows", 0, 0, 0, 0)
	pSwitch.AutoHeight = true
	setSpans5TilePanel(pSwitch, 12, 12, 6, 6, 6)
	pSwitch.Gap = 4
	pSwitch.TitleHeight = 32

	wifiTog := ui.NewToggle("b21-wifi-tog", true, 0, 0, 52, 28)
	wifi := ui.NewListTile("b21-wifi", "Wi-Fi", "Home network", 0, 0, 0, 0)
	wifi.SetRowMode(ui.ListTileSwitchOnly)
	wifi.SetTrailing(wifiTog)
	wifiTog.Value.Subscribe(func() {
		if wifiTog.Value.Get() {
			wifi.Subtitle = "Home network"
		} else {
			wifi.Subtitle = "Off"
		}
		wifi.MarkDrawDirty()
	})

	notifTog := ui.NewToggle("b21-notif-tog", false, 0, 0, 52, 28)
	notif := ui.NewListTile("b21-notif", "Notifications", "Banners and sounds", 0, 0, 0, 0)
	notif.SetRowMode(ui.ListTileSwitchOnly)
	notif.SetTrailing(notifTog)

	pSwitch.AddChild(batchCaption("b21-sw-cap", "Locked rows — only the trailing toggle receives clicks."))
	pSwitch.AddChild(wifi)
	pSwitch.AddChild(notif)

	pDense := ui.NewPanel("b21-dense", "Compact list", 0, 0, 0, 0)
	pDense.AutoHeight = true
	setSpans5TilePanel(pDense, 12, 12, 12, 12, 12)
	pDense.Gap = 4
	pDense.TitleHeight = 32
	pDense.AddChild(batchCaption("b21-dense-cap",
		"Dense = less vertical padding for long grouped lists. Subtitle is the current value; tap the row to open a detail screen (Clear / export actions belong there, not on the row)."))
	storage := []struct{ name, size string }{
		{"Cache", "124 MB used"},
		{"Cookies", "38 MB used"},
		{"Local storage", "12 MB used"},
	}
	for i, item := range storage {
		t := ui.NewListTile(fmt.Sprintf("b21-dense-%d", i), item.name, item.size, 0, 0, 0, 0)
		t.Dense = true
		t.SetTrailing(tileChevron(fmt.Sprintf("b21-d-%d", i), 18))
		name := item.name
		t.OnClick = func() {
			t.Subtitle = "Opening " + strings.ToLower(name) + "…"
			t.MarkDrawDirty()
		}
		pDense.AddChild(t)
	}

	grid.AddChild(pNav)
	grid.AddChild(pSwitch)
	grid.AddChild(pDense)
	page.Body.AddChild(grid)
}
