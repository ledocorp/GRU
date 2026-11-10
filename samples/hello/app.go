// Package hello is the public “Hello, Gru” sample UI (not part of the demo catalog).
//
// Shell: CP-SHELL-PAGE spirit — flex column root + Card content (no examples import).
package hello

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Build mounts a minimal interactive scene under doc.
func Build(doc *ui.Document) {
	w, h := float32(doc.Width), float32(doc.Height)
	if w < 1 {
		w = 360
	}
	if h < 1 {
		h = 640
	}

	// Phone density: larger padding / gaps / hit target (M1 readability).
	pad, gap, cardGap := float32(24), float32(16), float32(12)
	btnH, btnW := float32(48), float32(200)
	if ui.AndroidGLES() {
		pad, gap, cardGap = 32, 24, 16
		btnH, btnW = 56, 240
	}

	// Transparent flex root — DefaultStyle is opaque white and would hide the
	// clear-color backdrop + make phone screens look like an empty sheet.
	root := ui.NewContainer("hello-root", 0, 0, w, h)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.Gap = gap
	root.SetStyle("transparent")
	root.SetStyleOverrides(ui.Style{Padding: pad})
	doc.Root = root

	title := ui.NewLabel("hello-title", "Hello, Gru", 0, 0, 0, 0)
	title.SetStyle("form-field-caption")
	root.AddChild(title)

	sub := ui.NewLabel("hello-sub",
		"Standalone sample outside the demo catalog. Tap the button below.",
		0, 0, 0, 0)
	sub.SetStyle("form-label")
	sub.Wrap = true
	root.AddChild(sub)

	card := ui.NewCard("hello-card", "Getting started", 0, 0, 0, 0)
	card.Gap = cardGap
	// Cards default AutoHeight; keep content shrink-wrapped (not full-window fill).
	card.AutoHeight = true
	root.AddChild(card)

	statusSig := ui.NewSignal("Click the button.")
	status := ui.NewLabel("hello-status", statusSig.Get(), 0, 0, 0, 0)
	status.SetStyle("form-value")
	status.Wrap = true
	statusSig.Subscribe(func() {
		status.Text.Set(statusSig.Get())
		status.MarkDirty()
		card.MarkDirty()
	})
	card.AddChild(status)

	clicks := 0
	btn := ui.NewButton("hello-btn", "Click me", 0, 0, 0, btnH)
	btn.SetStyle("button")
	btn.PreferredWidth = btnW
	btn.OnClick = func() {
		clicks++
		statusSig.Set(fmt.Sprintf("Clicks: %d", clicks))
	}
	card.AddChild(btn)

	doc.ForceFullLayout()
	ui.SettleDocumentMount(doc)

	// Android NativeActivity: one TraceLog of geometry so logcat proves layout.
	if ui.AndroidGLES() {
		tb, sb, cb, bb := title.Bounds(), status.Bounds(), card.Bounds(), btn.Bounds()
		rl.TraceLog(rl.LogWarning,
			"hello layout title=(%.0f,%.0f %.0fx%.0f) status=(%.0f,%.0f %.0fx%.0f) card=(%.0f,%.0f %.0fx%.0f) btn=(%.0f,%.0f %.0fx%.0f) font=%s screen=%dx%d render=%dx%d dpi=%.2f rem=%.0f",
			tb.X, tb.Y, tb.Width, tb.Height,
			sb.X, sb.Y, sb.Width, sb.Height,
			cb.X, cb.Y, cb.Width, cb.Height,
			bb.X, bb.Y, bb.Width, bb.Height,
			ui.UIFontBackend(),
			rl.GetScreenWidth(), rl.GetScreenHeight(),
			rl.GetRenderWidth(), rl.GetRenderHeight(),
			ui.DisplayScale, ui.RootFontSize)
	}
}
