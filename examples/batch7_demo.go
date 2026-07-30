//go:build !notepad

// Package examples (continued)
package examples

import (
	"fmt"
	"time"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &batch7Scene{} }) }

// batch7Scene demonstrates the Toast notification system (Batch 7).
//
// Three responsive panels:
//
//   - "Toast Levels" - fire each severity and sticky variants.
//   - "Timing"       - durations, stacking, and click callbacks.
//   - "Behavior"     - what the overlay manager is demonstrating.
type batch7Scene struct {
	BaseScene
}

func (s *batch7Scene) Title() string { return "Batch 7 · Toast / Notification" }

func (s *batch7Scene) OnUpdate(_ *ui.Document, _ float32) {}

func setSpans5ToastPanel(p *ui.Panel, xs, sm, md, lg, xl int) {
	p.SetColSpan(ui.BreakpointXS, xs)
	p.SetColSpan(ui.BreakpointSM, sm)
	p.SetColSpan(ui.BreakpointMD, md)
	p.SetColSpan(ui.BreakpointLG, lg)
	p.SetColSpan(ui.BreakpointXL, xl)
}

func toastButton(id, label, style string, onClick func()) *ui.Button {
	b := ui.NewButton(id, label, 0, 0, 0, 34)
	b.SetStyle(style)
	b.OnClick = onClick
	return b
}

func addToastText(p *ui.Panel, id, text string) {
	p.AddChild(FlexCopy(id, "form-value", text))
}

func (s *batch7Scene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "b7",
		"Widget Batch 7 · Toast / Notification",
		"Global overlay cards for transient status, sticky alerts, stacked messages, and click callbacks.")
	page.Body.Gap = 12

	grid := NewBatchPageGrid("b7-grid", 12)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 1: severity levels and sticky notifications.
	// ══════════════════════════════════════════════════════════════════════════
	pLevels := ui.NewPanel("p-b7-levels", "Toast Levels", 0, 0, 0, 0)
	setSpans5ToastPanel(pLevels, 12, 12, 12, 4, 4)
	pLevels.Gap = 10
	pLevels.TitleHeight = 32

	addToastText(pLevels, "b7-levels-note", "Use these to preview each palette and icon.")
	pLevels.AddChild(ui.NewSeparator("b7-sep-levels", "Timed toasts", 0, 0, 0, 22))

	pLevels.AddChild(toastButton("b7-info", "Info Toast", "button", func() {
		ui.ShowToast("A neutral informational message.", ui.ToastInfo, 3*time.Second)
	}))
	pLevels.AddChild(toastButton("b7-success", "Success Toast", "primary", func() {
		ui.ShowToast("File saved successfully.", ui.ToastSuccess, 3*time.Second)
	}))
	pLevels.AddChild(toastButton("b7-warn", "Warning Toast", "button", func() {
		ui.ShowToast("Disk space is running low.", ui.ToastWarning, 3*time.Second)
	}))
	pLevels.AddChild(toastButton("b7-error", "Error Toast", "danger", func() {
		ui.ShowToast("Connection to server lost.", ui.ToastError, 3*time.Second)
	}))

	pLevels.AddChild(ui.NewSeparator("b7-sep-sticky", "Sticky toasts", 0, 0, 0, 22))
	pLevels.AddChild(toastButton("b7-sticky-info", "Sticky Info", "button", func() {
		ui.ShowToast("Sticky info: click the card to dismiss.", ui.ToastInfo, 0)
	}))
	pLevels.AddChild(toastButton("b7-sticky-warn", "Sticky Warning", "button", func() {
		ui.ShowToast("Sticky warning: action required.", ui.ToastWarning, 0)
	}))
	pLevels.AddChild(toastButton("b7-dismiss", "Dismiss All", "danger", func() {
		ui.DismissAll()
	}))

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 2: durations, stacking, and callbacks.
	// ══════════════════════════════════════════════════════════════════════════
	pTiming := ui.NewPanel("p-b7-timing", "Timing", 0, 0, 0, 0)
	setSpans5ToastPanel(pTiming, 12, 12, 12, 4, 4)
	pTiming.Gap = 10
	pTiming.TitleHeight = 32

	pTiming.AddChild(ui.NewSeparator("b7-sep-duration", "Durations", 0, 0, 0, 22))
	durRow := ui.NewContainer("b7-duration-row", 0, 0, 0, 34)
	durRow.FlexDirection = ui.FlexRow
	durRow.Gap = 6
	durRow.SetStyle("transparent")
	for _, seconds := range []int{1, 3, 5} {
		secs := seconds
		b := ui.NewButton(fmt.Sprintf("b7-dur-%d", secs), fmt.Sprintf("%ds", secs), 0, 0, 0, 34)
		b.SetFlexGrow(1)
		b.SetStyle("button")
		b.OnClick = func() {
			ui.ShowToast(fmt.Sprintf("Auto-dismisses in %d second(s).", secs), ui.ToastInfo, time.Duration(secs)*time.Second)
		}
		durRow.AddChild(b)
	}
	pTiming.AddChild(durRow)

	pTiming.AddChild(ui.NewSeparator("b7-sep-stack", "Stacking", 0, 0, 0, 22))
	pTiming.AddChild(toastButton("b7-stack3", "Stack 3 Toasts", "button", func() {
		ui.ShowToast("First toast: oldest.", ui.ToastInfo, 4*time.Second)
		ui.ShowToast("Second toast: stacks above.", ui.ToastSuccess, 3*time.Second)
		ui.ShowToast("Third toast: newest.", ui.ToastWarning, 2*time.Second)
	}))
	pTiming.AddChild(toastButton("b7-stack5", "Stack 5 Toasts", "button", func() {
		levels := []ui.ToastLevel{ui.ToastInfo, ui.ToastSuccess, ui.ToastWarning, ui.ToastError, ui.ToastInfo}
		for i, level := range levels {
			ui.ShowToast(fmt.Sprintf("Stack item %d of 5.", i+1), level, 5*time.Second)
		}
	}))

	pTiming.AddChild(ui.NewSeparator("b7-sep-action", "Action label", 0, 0, 0, 22))
	actionLog := ui.NewSignal("Action: none")
	actionResult, actionDisplay := FlexCopyPair("b7-action-result", "form-label", actionLog.Get())
	actionLog.Subscribe(func() {
		actionDisplay.Set(actionLog.Get())
	})
	pTiming.AddChild(toastButton("b7-action-undo", "Toast with Undo", "primary", func() {
		ui.ShowToastWithAction("Item deleted.", ui.ToastInfo, 5*time.Second, "Undo", func() {
			actionLog.Set("Action: Undo clicked")
		})
	}))
	pTiming.AddChild(actionResult)

	pTiming.AddChild(ui.NewSeparator("b7-sep-callback", "Clickable callbacks", 0, 0, 0, 22))
	callbackLog := ui.NewSignal("Callback: none")
	callbackResult, callbackDisplay := FlexCopyPair("b7-callback-result", "form-label", callbackLog.Get())
	callbackLog.Subscribe(func() {
		callbackDisplay.Set(callbackLog.Get())
	})

	pTiming.AddChild(toastButton("b7-click-info", "Clickable Info", "button", func() {
		ui.ShowToastClickable("Click for info callback.", ui.ToastInfo, 0, func() {
			callbackLog.Set("Callback: info toast clicked")
		})
	}))
	pTiming.AddChild(toastButton("b7-click-success", "Clickable Success", "primary", func() {
		ui.ShowToastClickable("Upload complete — click to view.", ui.ToastSuccess, 5*time.Second, func() {
			callbackLog.Set("Callback: success toast clicked")
		})
	}))
	pTiming.AddChild(toastButton("b7-click-error", "Clickable Error", "danger", func() {
		ui.ShowToastClickable("Build failed — click to inspect.", ui.ToastError, 0, func() {
			callbackLog.Set("Callback: error toast clicked")
		})
	}))
	pTiming.AddChild(callbackResult)

	// ══════════════════════════════════════════════════════════════════════════
	// Panel 3: explanation and API surface.
	// ══════════════════════════════════════════════════════════════════════════
	pBehavior := ui.NewPanel("p-b7-behavior", "Behavior", 0, 0, 0, 0)
	setSpans5ToastPanel(pBehavior, 12, 12, 12, 4, 4)
	pBehavior.Gap = 10
	pBehavior.TitleHeight = 32

	pBehavior.AddChild(ui.NewSeparator("b7-sep-feature", "What this shows", 0, 0, 0, 22))
	features := []string{
		"Global overlay manager",
		"Bottom-right anchored cards",
		"Slide and fade animations",
		"Timed progress bar",
		"Sticky duration = 0",
		"Click card to dismiss",
		"Max stack count: 5",
	}
	for i, feature := range features {
		addToastText(pBehavior, fmt.Sprintf("b7-feature-%d", i), "- "+feature)
	}

	pBehavior.AddChild(ui.NewSeparator("b7-sep-api", "API", 0, 0, 0, 22))
	apiLines := []string{
		"ui.ShowToast(msg, level, dur)",
		"ui.ShowToastWithAction(..., label, fn)",
		"ui.ShowToastClickable(..., fn)",
		"ui.DismissAll()",
		"ui.ActiveToastCount()",
	}
	for i, line := range apiLines {
		pBehavior.AddChild(FlexCopy(fmt.Sprintf("b7-api-%d", i), "form-label", line))
	}

	pBehavior.AddChild(ui.NewSeparator("b7-sep-tips", "Tips", 0, 0, 0, 22))
	tips := []string{
		"Click a toast card to dismiss it.",
		"Clickable toasts run a callback first.",
		"Newest toast appears at the bottom.",
		"Dismiss All clears the current stack.",
	}
	for i, tip := range tips {
		addToastText(pBehavior, fmt.Sprintf("b7-tip-%d", i), "- "+tip)
	}

	grid.AddChild(pLevels)
	grid.AddChild(pTiming)
	grid.AddChild(pBehavior)

	page.Body.AddChild(grid)
}
