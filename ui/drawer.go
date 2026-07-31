// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── Drawer ──────────────────────────────────────────────────────────────────
//
// DrawerMgr provides a slide-in side panel over a dimmed scrim (Material drawer).
// Animation, scrim, backdrop, and Escape use OverlayHost (C5).
//
//	ui.OpenDrawer(navContent)
//	ui.CloseDrawer()
//
// Optional open state signal:
//
//	open := ui.NewSignal(false)
//	ui.DrawerMgr.BindOpen(open)
//
// Drive from main.go:
//
//	ui.DrawerMgr.Update(dt)
//	ui.DrawerMgr.Draw() // inside drawScreenOverlays
//
// # LLM Prompt Template
//
//	ui.OpenDrawer(navContent)
//	// main loop: ui.DrawerMgr.Update(dt); ui.DrawerMgr.Draw()
//
// Demo scenes: **AppShell Demo**, **Shell Desktop Demo**.
//
// Styled via "drawer" theme key (falls back to modal surface).

const drawerDefaultW = float32(280)

type drawerManager struct {
	host *OverlayHost

	content Node
	width   float32
	CloseOnBackdrop bool
	CloseOnEscape   bool
	openSignal      *Signal[bool]
}

// DrawerMgr is the package-level drawer overlay manager.
var DrawerMgr = &drawerManager{
	host:            DefaultOverlayHost(OverlayAnimSlideLeft),
	CloseOnBackdrop: true,
	CloseOnEscape:   true,
}

func (d *drawerManager) syncHostOptions() {
	d.host.CloseOnBackdrop = d.CloseOnBackdrop
	d.host.CloseOnEscape = d.CloseOnEscape
}

// BindOpen syncs DrawerMgr open/close with a bool signal (optional).
func (d *drawerManager) BindOpen(sig *Signal[bool]) {
	d.openSignal = sig
	if sig != nil {
		sig.Subscribe(func() {
			if sig.Get() {
				if d.content != nil {
					d.beginOpen()
				}
			} else {
				d.beginClose()
			}
		})
	}
}

func (d *drawerManager) setOpenSignal(v bool) {
	if d.openSignal != nil && d.openSignal.Get() != v {
		d.openSignal.Set(v)
	}
}

// OpenDrawer slides in content from the left. Replaces any existing drawer content.
func OpenDrawer(content Node) {
	if content == nil {
		return
	}
	DrawerMgr.content = content
	DrawerMgr.beginOpen()
}

func (d *drawerManager) beginOpen() {
	if d.content == nil {
		return
	}
	d.syncHostOptions()
	d.host.BeginOpen()
	d.setOpenSignal(true)
}

// CloseDrawer begins the slide-out animation.
func CloseDrawer() {
	DrawerMgr.beginClose()
}

func (d *drawerManager) beginClose() {
	if !d.host.Open {
		return
	}
	d.host.BeginClose()
	d.setOpenSignal(false)
}

func (d *drawerManager) reset() {
	d.host.Reset()
	d.setOpenSignal(false)
}

// IsDrawerOpen is true while the drawer is logically open (before animation ends).
func IsDrawerOpen() bool { return DrawerMgr.host.IsOpen() }

// IsDrawerVisible is true while the drawer is rendered (including close animation).
func IsDrawerVisible() bool { return DrawerMgr.host.Open }

func (d *drawerManager) IsAnimating() bool { return d.host.IsAnimating() }

func (d *drawerManager) panelWidth(sw float32) float32 {
	w := d.width
	if w <= 0 {
		w = drawerDefaultW
	}
	maxW := sw * 0.85
	if w > maxW {
		w = maxW
	}
	if w < 200 {
		w = 200
	}
	return w
}

// SetContentInsets limits the drawer panel and scrim to the content band below
// window chrome (custom title bar) and above footer chrome (launcher nav).
// Pass 0 for either inset when that chrome is absent (e.g. full-bleed mobile).
func (d *drawerManager) SetContentInsets(top, bottom float32) {
	d.host.SetContentInsets(top, bottom)
}

func (d *drawerManager) panelRect(sw, sh float32) rl.Rectangle {
	band := d.host.ContentBand(sw, sh)
	w := d.panelWidth(sw)
	return d.host.SlidePanelRect(band, w)
}

// ContentBandHitRect is the scrim/content band used to block scene pointer hit-tests.
func (d *drawerManager) ContentBandHitRect(sw, sh float32) rl.Rectangle {
	return d.host.ContentBand(sw, sh)
}

// Update advances animation and handles input.
func (d *drawerManager) Update(dt float32) {
	if !d.host.Open {
		return
	}
	d.syncHostOptions()
	d.host.AdvanceAnimation(dt)
	if !d.host.Open {
		return
	}

	if !d.host.InputReady() {
		d.host.TickSkipFrames()
		return
	}
	if d.host.HandleEscape(d.beginClose) {
		return
	}

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	band := d.host.ContentBand(sw, sh)
	panel := d.panelRect(sw, sh)
	if d.host.HandleBackdrop(panel, band, d.beginClose) {
		return
	}

	mouse := rl.GetMousePosition()
	if rl.CheckCollisionPointRec(mouse, panel) && d.content != nil {
		inner := d.contentInnerRect(panel)
		layoutOverlaySubtree(d.content, inner)
		d.content.Update(dt)
	}
}

func (d *drawerManager) contentInnerRect(panel rl.Rectangle) rl.Rectangle {
	pad := GetThemeStyle("drawer").Padding
	return rl.NewRectangle(panel.X+pad, panel.Y+pad, panel.Width-2*pad, panel.Height-2*pad)
}

// Draw renders scrim + sliding panel. Call from drawScreenOverlays.
func (d *drawerManager) Draw() {
	if !d.host.Open || d.host.Progress <= 0 {
		return
	}
	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	a := d.host.Opacity()
	band := d.host.ContentBand(sw, sh)
	d.host.DrawScrimRect(band)

	panel := d.panelRect(sw, sh)
	style := GetThemeStyle("drawer")
	bg := style.BackgroundColor
	bg.A = uint8(float32(bg.A) * a)
	rl.DrawRectangleRec(panel, bg)
	bc := style.BorderColor
	bc.A = uint8(float32(bc.A) * a)
	rl.DrawRectangle(int32(panel.X+panel.Width-1), int32(panel.Y), 1, int32(panel.Height), bc)

	if d.content != nil && !d.host.Closing {
		drawOverlaySubtree(d.content, d.contentInnerRect(panel))
	}
}

// DebugInfo returns inspector status text.
func (d *drawerManager) DebugInfo() string {
	if !d.host.Open {
		return "drawer: closed"
	}
	state := "open"
	if d.host.Closing {
		state = "closing"
	}
	return fmt.Sprintf("drawer: %s  progress:%.2f", state, d.host.Progress)
}
