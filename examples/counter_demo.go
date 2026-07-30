//go:build !notepad

// Package examples contains self-contained Gru demo scenes.
package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &counterScene{} }) }

// counterScene demonstrates reactive Signals, effects, progress, and tweens.
// Recipe: CP-SHELL-PAGE + NewBatchPageGrid; PlainText for body copy (not Label).
type counterScene struct {
	BaseScene
	tweens []*ui.Tween
}

func (s *counterScene) Title() string { return "Counter Demo" }

func setSpans5CounterPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func counterPanel(id, title string) *ui.Panel {
	p := ui.NewPanel(id, title, 0, 0, 0, 0)
	p.AutoHeight = true
	p.Gap = 10
	p.TitleHeight = 32
	return p
}

func counterButton(id, text string, w float32, primary bool, onClick func()) *ui.Button {
	btn := ui.NewButton(id, text, 0, 0, w, 36)
	if primary {
		btn.SetStyle("primary")
	} else {
		btn.SetStyle("button")
	}
	btn.OnClick = onClick
	return btn
}

func counterRow(id string) *ui.Container {
	row := ui.NewContainer(id, 0, 0, 0, 40)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 8
	row.AutoHeight = false
	row.SetStyle("transparent")
	return row
}

func (s *counterScene) bounce(btn *ui.Button) {
	t := ui.NewTween(1.0, 1.08, 0.10, ui.EaseOutQuad,
		func(v float32) { btn.Scale = v },
		func() { btn.Scale = 1.0 },
	)
	s.tweens = append(s.tweens, t)
}

func (s *counterScene) Build(doc *ui.Document) {
	s.tweens = s.tweens[:0]
	counter := ui.NewSignal(0)
	autoReset := ui.NewSignal(false)

	page := MountAppPage(doc, "counter",
		"Counter Demo",
		"Signals, effects, progress, derived values, and button tween feedback.")
	page.Body.Gap = 16

	grid := NewBatchPageGrid("counter-grid", 12)
	grid.AddChild(s.buildCounterPanel(counter, autoReset))
	grid.AddChild(s.buildProgressPanel(counter))
	grid.AddChild(s.buildStatsPanel(counter))
	grid.AddChild(s.buildControlsPanel(counter, autoReset))
	page.Body.AddChild(grid)
}

func (s *counterScene) buildCounterPanel(counter *ui.Signal[int], autoReset *ui.Signal[bool]) *ui.Panel {
	p := counterPanel("counter-main-panel", "Counter")
	setSpans5CounterPanel(p, 12, 12, 6, 6, 6)
	p.AddChild(batchCaption("counter-main-hint",
		"Click buttons to update a shared signal. Readouts refresh through effects."))

	countLbl, countDisplay := FlexCopyPair("counter-count", "form-value", "Count: 0")
	ui.NewEffect(func() {
		countDisplay.Set(fmt.Sprintf("Count: %d", counter.Get()))
	})
	p.AddChild(countLbl)

	row := counterRow("counter-actions")
	incBtn := counterButton("counter-inc", "Increment", 128, true, nil)
	incBtn.OnClick = func() {
		next := counter.Get() + 1
		if autoReset.Get() && next > 100 {
			next = 0
		}
		counter.Set(next)
		s.bounce(incBtn)
	}
	row.AddChild(incBtn)
	row.AddChild(counterButton("counter-reset", "Reset", 96, false, func() { counter.Set(0) }))
	p.AddChild(row)
	return p
}

func (s *counterScene) buildProgressPanel(counter *ui.Signal[int]) *ui.Panel {
	p := counterPanel("counter-progress-panel", "Goal Progress")
	setSpans5CounterPanel(p, 12, 12, 6, 6, 6)
	p.AddChild(batchCaption("counter-progress-hint",
		"Progress clamps visually at 100 while the count can continue."))

	progressLbl, progressDisplay := FlexCopyPair("counter-progress", "form-value", "0 / 100")
	progressBar := ui.NewProgressBar("counter-progress-bar", 0, 0, 0, 0, 24)
	ui.NewEffect(func() {
		v := counter.Get()
		capped := v
		if capped > 100 {
			capped = 100
		}
		if capped < 0 {
			capped = 0
		}
		progressDisplay.Set(fmt.Sprintf("%d / 100", v))
		progressBar.Value.Set(float32(capped) / 100)
	})
	p.AddChild(progressLbl)
	p.AddChild(progressBar)
	return p
}

func (s *counterScene) buildStatsPanel(counter *ui.Signal[int]) *ui.Panel {
	p := counterPanel("counter-stats-panel", "Derived Values")
	setSpans5CounterPanel(p, 12, 12, 6, 6, 6)
	p.AddChild(batchCaption("counter-stats-hint",
		"Computed labels stay in sync with the same source signal."))

	doubleLbl, doubleDisplay := FlexCopyPair("counter-double", "form-value", "Double: 0")
	squareLbl, squareDisplay := FlexCopyPair("counter-square", "form-value", "Square: 0")
	parityLbl, parityDisplay := FlexCopyPair("counter-parity", "form-value", "Parity: even")
	ui.NewEffect(func() {
		n := counter.Get()
		parity := "even"
		if n%2 != 0 {
			parity = "odd"
		}
		doubleDisplay.Set(fmt.Sprintf("Double: %d", n*2))
		squareDisplay.Set(fmt.Sprintf("Square: %d", n*n))
		parityDisplay.Set("Parity: " + parity)
	})
	p.AddChild(doubleLbl)
	p.AddChild(squareLbl)
	p.AddChild(parityLbl)
	return p
}

func (s *counterScene) buildControlsPanel(counter *ui.Signal[int], autoReset *ui.Signal[bool]) *ui.Panel {
	p := counterPanel("counter-controls-panel", "Controls")
	setSpans5CounterPanel(p, 12, 12, 6, 6, 6)
	p.AddChild(batchCaption("counter-controls-hint",
		"Step buttons and a compact auto-reset toggle."))

	row := counterRow("counter-step-row")
	row.AddChild(counterButton("counter-minus", "Minus 1", 100, false, func() { counter.Set(counter.Get() - 1) }))
	row.AddChild(counterButton("counter-plus-five", "Plus 5", 100, false, func() { counter.Set(counter.Get() + 5) }))
	row.AddChild(counterButton("counter-plus-ten", "Plus 10", 108, false, func() { counter.Set(counter.Get() + 10) }))
	p.AddChild(row)

	toggleRow := ui.NewContainer("counter-toggle-row", 0, 0, 0, 28)
	toggleRow.LayoutType = ui.LayoutFlex
	toggleRow.FlexDirection = ui.FlexRow
	toggleRow.Gap = 10
	toggleRow.AutoHeight = false
	toggleRow.SetStyle("transparent")
	toggle := ui.NewToggle("counter-auto-reset", false, 0, 0, 52, 28)
	toggle.Value = autoReset
	toggle.Value.Subscribe(func() { toggle.MarkDirty() })
	toggleRow.AddChild(toggle)

	toggleLbl, toggleDisplay := FlexCopyPair("counter-auto-reset-label", "form-value", "Auto reset after 100: Off")
	toggleLbl.SetFlexGrow(1)
	autoReset.Subscribe(func() {
		if autoReset.Get() {
			toggleDisplay.Set("Auto reset after 100: On")
		} else {
			toggleDisplay.Set("Auto reset after 100: Off")
		}
	})
	toggleRow.AddChild(toggleLbl)
	p.AddChild(toggleRow)
	return p
}

func (s *counterScene) OnUpdate(d *ui.Document, dt float32) {
	active := s.tweens[:0]
	for _, tw := range s.tweens {
		tw.Update(dt)
		if tw.IsActive {
			active = append(active, tw)
		}
	}
	s.tweens = active
	if len(s.tweens) > 0 {
		ui.Wake(ui.WakeAnimation, "counter-tween")
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		d.SetFocus(nil)
	}
}
