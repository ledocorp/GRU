//go:build !notepad

// Package examples — Go-only nested Card demo (callout/code inside a card).
// Validates docs/GO_FIRST_UI.md: nesting is a Card layout concern, not JSON-only.
package examples

import (
	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &cardNestScene{} }) }

type cardNestScene struct {
	BaseScene
}

func (s *cardNestScene) Title() string { return "Card Nest (Go)" }

func (s *cardNestScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "card-nest",
		"Card Nest",
		"Nested Card shells — callout and code variants inside one page card.")
	page.Body.Gap = 12

	card := ui.NewCard("page-card", "Nested cards", 0, 0, 0, 0)
	card.Gap = 14
	page.Body.AddChild(card)

	card.AddChild(ui.NewRichText("intro", []ui.TextSpan{
		{Text: "Outer card with ", Variant: "muted"},
		{Text: "callout", Variant: "code"},
		{Text: " and ", Variant: "muted"},
		{Text: "code", Variant: "code"},
		{Text: " children — same widgets as .gru callout/code blocks.", Variant: "muted"},
	}, 0, 0, 0, 0))

	callout := ui.NewCard("callout", "Tip", 0, 0, 0, 0)
	callout.SetStyleVariant("card", "callout")
	callout.AddChild(ui.NewRichText("callout-text", []ui.TextSpan{
		{Text: "Scrollable app pages use one section in the viewport, then cards for each topic."},
	}, 0, 0, 0, 0))
	card.AddChild(callout)

	code := ui.NewCard("code", "Example", 0, 0, 0, 0)
	code.SetStyleVariant("card", "code")
	code.AddChild(ui.NewRichText("code-text", []ui.TextSpan{
		{Text: "{\n  \"type\": \"callout\",\n  \"title\": \"Tip\",\n  \"text\": \"…\"\n}", Variant: "code"},
	}, 0, 0, 0, 0))
	card.AddChild(code)

	FinishShellMount(doc)
}
