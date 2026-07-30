// Package studio implements Gru Studio chrome — the internal demo launcher side panel.
//
// LLM Prompt Template:
// Add Studio panel rows in devtools/studio/panel.go; handle returned Action in main.go.
package studio

import (
	"fmt"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	PanelWidth = 300
	navBarH    = 36
	menuBtnW   = 76
	rowH       = 40
	pad        = 14
	headerH    = 68 // title + scene subtitle + divider (subtitle was clipped at 56)
)

// Action is a one-shot command from a panel row click.
type Action int

const (
	ActionNone Action = iota
	ActionOpenDirectory
	ActionNextScene
	ActionToggleBenchmark
	ActionTogglePerfOverlay
	ActionToggleResizeFPS
	ActionToggleCaretDebug
	ActionToggleDrawDirty
	ActionToggleWebViewDebug
	ActionToggleChrome
	ActionToggleInspector
)

// ToolState mirrors dev-tool toggles for panel labels.
type ToolState struct {
	SceneIndex     int
	SceneCount     int
	SceneTitle     string
	BenchmarkOn    bool
	PerfOverlayOn  bool
	ResizeFPSOn    bool
	CaretDebugOn   bool
	DrawDirtyOn    bool
	WebViewDebugOn bool
	BorderlessOn   bool
	InspectorOn    bool
}

// Panel is the slide-out Gru Studio menu.
type Panel struct {
	Open      bool
	hoverRow  int // index into rows(); -1 = none
	menuHover bool
}

// MenuButtonRect is the footer control that opens the panel (left side).
func MenuButtonRect(windowW, windowH int32) rl.Rectangle {
	return rl.NewRectangle(8, float32(windowH-navBarH)+6, menuBtnW, float32(navBarH)-12)
}

// PanelRect is the slide-out area when Open.
func (p *Panel) PanelRect(windowW, windowH int32) rl.Rectangle {
	h := float32(windowH)
	if windowH > navBarH {
		h = float32(windowH - navBarH)
	}
	return rl.NewRectangle(0, 0, PanelWidth, h)
}

// BlocksSceneInput reports whether the panel should eat scene pointer events.
func (p *Panel) BlocksSceneInput(windowW, windowH int32) bool {
	if !p.Open {
		return false
	}
	mouse := rl.GetMousePosition()
	r := p.PanelRect(windowW, windowH)
	return rl.CheckCollisionPointRec(mouse, r)
}

// Toggle opens or closes the panel.
func (p *Panel) Toggle() { p.Open = !p.Open }

// Close hides the panel.
func (p *Panel) Close() { p.Open = false }

func (p *Panel) refreshHover(windowW, windowH int32) {
	mouse := rl.GetMousePosition()
	p.menuHover = rl.CheckCollisionPointRec(mouse, MenuButtonRect(windowW, windowH))
	p.hoverRow = -1
	if !p.Open {
		return
	}
	panel := p.PanelRect(windowW, windowH)
	if !rl.CheckCollisionPointRec(mouse, panel) {
		return
	}
	relY := mouse.Y - panel.Y
	if relY < headerH {
		return
	}
	idx := int((relY - headerH) / rowH)
	rows := p.rows(ToolState{})
	if idx >= 0 && idx < len(rows) && rows[idx].action != ActionNone {
		p.hoverRow = idx
	}
}

func (p *Panel) applyCursor() {
	if p.menuHover || p.hoverRow >= 0 {
		rl.SetMouseCursor(rl.MouseCursorPointingHand)
	}
}

// Update handles menu button, outside dismiss, and row clicks. Returns an action for main.
func (p *Panel) Update(windowW, windowH int32, state ToolState) Action {
	p.refreshHover(windowW, windowH)

	if ui.PointerClickConsume(MenuButtonRect(windowW, windowH)) {
		p.Toggle()
		return ActionNone
	}
	if !p.Open {
		return ActionNone
	}

	panel := p.PanelRect(windowW, windowH)
	mouse := rl.GetMousePosition()
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !rl.CheckCollisionPointRec(mouse, panel) &&
		!rl.CheckCollisionPointRec(mouse, MenuButtonRect(windowW, windowH)) {
		p.Close()
		return ActionNone
	}

	rows := p.rows(state)
	for i, row := range rows {
		if row.action == ActionNone {
			continue
		}
		if ui.PointerClickConsume(rowRect(panel, i)) {
			p.Close()
			return row.action
		}
	}
	return ActionNone
}

type panelRow struct {
	label  string
	action Action
}

func (p *Panel) rows(state ToolState) []panelRow {
	on := func(v bool) string {
		if v {
			return "ON"
		}
		return "off"
	}
	scene := fmt.Sprintf("Next scene (Tab) %d/%d", state.SceneIndex+1, state.SceneCount)
	return []panelRow{
		{label: "Open Directory", action: ActionOpenDirectory},
		{label: scene, action: ActionNextScene},
		{label: "Tools", action: ActionNone},
		{label: fmt.Sprintf("F6 Benchmark [%s]", on(state.BenchmarkOn)), action: ActionToggleBenchmark},
		{label: fmt.Sprintf("F11 Perf overlay [%s]", on(state.PerfOverlayOn)), action: ActionTogglePerfOverlay},
		{label: fmt.Sprintf("F7 Resize FPS trace [%s]", on(state.ResizeFPSOn)), action: ActionToggleResizeFPS},
		{label: fmt.Sprintf("F5 Caret debug [%s]", on(state.CaretDebugOn)), action: ActionToggleCaretDebug},
		{label: fmt.Sprintf("F10 Draw-dirty trace [%s]", on(state.DrawDirtyOn)), action: ActionToggleDrawDirty},
		{label: fmt.Sprintf("Shift+F11 WebView debug [%s]", on(state.WebViewDebugOn)), action: ActionToggleWebViewDebug},
		{label: fmt.Sprintf("F9 Borderless chrome [%s]", on(state.BorderlessOn)), action: ActionToggleChrome},
		{label: fmt.Sprintf("F12 Inspector [%s]", on(state.InspectorOn)), action: ActionToggleInspector},
	}
}

func rowRect(panel rl.Rectangle, idx int) rl.Rectangle {
	y := panel.Y + headerH + float32(idx)*rowH
	return rl.NewRectangle(panel.X+6, y+2, panel.Width-12, rowH-4)
}

// PaintFooter draws the footer Menu button and open slide-out panel.
// Call from the SSAA overlay pass together with nav bar backgrounds.
func (p *Panel) PaintFooter(windowW, windowH int32, state ToolState) {
	p.refreshHover(windowW, windowH)
	p.applyCursor()
	p.drawMenuButton(windowW, windowH)
	if p.Open {
		p.drawPanel(windowW, windowH, state)
	}
}

// PaintOverlays draws the footer Menu button (with hover) and open panel.
func (p *Panel) PaintOverlays(windowW, windowH int32, state ToolState) {
	p.PaintFooter(windowW, windowH, state)
}

func (p *Panel) drawMenuButton(windowW, windowH int32) {
	btn := MenuButtonRect(windowW, windowH)
	bg := rl.NewColor(55, 60, 78, 255)
	if p.menuHover || p.Open {
		bg = rl.NewColor(79, 70, 229, 220)
	}
	rl.DrawRectangleRounded(btn, 0.25, 6, bg)
	if p.menuHover || p.Open {
		rl.DrawRectangleRoundedLines(btn, 0.25, 6, rl.NewColor(140, 130, 255, 255))
	}
	label := "Menu"
	if p.Open {
		label = "Menu -"
	}
	menu := ui.ChromeFooterButtonStyle()
	menu.TextColor = rl.White
	tw := ui.MeasureChromeText(label, menu)
	x := btn.X + (btn.Width-tw)/2
	y := ui.ChromeTextCenterY(btn, menu)
	ui.DrawChromeText(label, x, y, menu)
}

func (p *Panel) drawPanel(windowW, windowH int32, state ToolState) {
	panel := p.PanelRect(windowW, windowH)
	rl.DrawRectangleRec(panel, rl.NewColor(22, 24, 34, 248))
	rl.DrawRectangleLinesEx(panel, 1, rl.NewColor(70, 75, 100, 255))

	title := ui.ChromeTitleStyle()
	ui.DrawChromeText("Gru Studio", pad, panel.Y+10, title)

	sub := state.SceneTitle
	if len(sub) > 36 {
		sub = sub[:33] + "..."
	}
	muted := ui.ChromeMutedStyle()
	ui.DrawChromeText(sub, pad, panel.Y+34, muted)

	dividerY := int32(panel.Y + headerH - 8)
	rl.DrawRectangle(int32(panel.X+pad), dividerY, int32(panel.Width-pad*2), 1,
		rl.NewColor(60, 65, 88, 255))

	section := ui.ChromeMutedStyle()
	section.Bold = true

	body := ui.ChromeBodyStyle()
	rows := p.rows(state)
	for i, row := range rows {
		if row.action == ActionNone {
			r := rowRect(panel, i)
			y := ui.ChromeTextCenterY(r, section)
			ui.DrawChromeText(row.label, pad, y, section)
			continue
		}
		r := rowRect(panel, i)
		if p.hoverRow == i {
			rl.DrawRectangleRounded(r, 0.2, 4, rl.NewColor(79, 70, 229, 255))
			rl.DrawRectangleRoundedLines(r, 0.2, 4, rl.NewColor(140, 130, 255, 255))
		}
		y := ui.ChromeTextCenterY(r, body)
		ui.DrawChromeText(row.label, r.X+10, y, body)
	}
}
