// Package hello is the public “Hello, Gru” sample UI (not part of the demo catalog).
//
// Shell: CP-SHELL-PAGE spirit — flex column root + Card content (no examples import).
package hello

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

// Build mounts a minimal interactive scene under doc.
func Build(doc *ui.Document) {
	w, h := float32(doc.Width), float32(doc.Height)
	root := ui.NewContainer("hello-root", 0, 0, w, h)
	root.LayoutType = ui.LayoutFlex
	root.FlexDirection = ui.FlexColumn
	root.Gap = 16
	root.SetStyleOverrides(ui.Style{Padding: 24})
	doc.Root = root

	root.AddChild(ui.NewPlainText("hello-title", "form-value", "Hello, Gru", 0, 0, 0, 0))
	root.AddChild(ui.NewPlainText("hello-sub", "form-label",
		"Standalone sample outside the demo catalog. Edit samples/hello and re-run.",
		0, 0, 0, 0))

	card := ui.NewCard("hello-card", "Getting started", 0, 0, 0, 0)
	card.Gap = 12
	root.AddChild(card)

	statusSig := ui.NewSignal("Click the button.")
	status := ui.NewPlainText("hello-status", "form-label", statusSig.Get(), 0, 0, 0, 0)
	ui.BindPlainText(status, statusSig)
	card.AddChild(status)

	clicks := 0
	btn := ui.NewButton("hello-btn", "Click me", 0, 0, 0, 0)
	btn.OnClick = func() {
		clicks++
		statusSig.Set(fmt.Sprintf("Clicks: %d", clicks))
	}
	card.AddChild(btn)

	doc.Root.MarkDirty()
	for i := 0; i < 5; i++ {
		doc.Root.Layout()
	}
}
