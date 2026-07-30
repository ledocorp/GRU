//go:build !notepad

// Package examples (continued)
package examples

import (
	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &timelineScene{} }) }

// timelineScene demonstrates the Timeline widget for activity and event history.
type timelineScene struct {
	BaseScene
}

func (s *timelineScene) Title() string { return "Timeline (Go)" }

func (s *timelineScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "timeline",
		"Timeline",
		"Vertical activity feed with track, dots, and timestamps — early widget demo, panel-hosted.")
	page.Body.Gap = 12

	panel := ui.NewPanel("timeline-panel", "Activity", 0, 0, 0, 0)
	panel.AutoHeight = true
	panel.Gap = 10
	panel.TitleHeight = 34
	panel.AddChild(FlexCopy("timeline-hint", "form-value",
		"Resize the window to confirm AutoHeight keeps the feed flush with the panel body."))

	tl := ui.NewTimeline("timeline-feed", []ui.TimelineEvent{
		{
			Title:    "Project created",
			Subtitle: "Gru workspace initialized",
			Time:     "09:00",
		},
		{
			Title:    "Widgets batch merged",
			Subtitle: "Batch 1–3 controls available in package ui",
			Time:     "10:15",
		},
		{
			Title:    "Shell demos added",
			Subtitle: "App Shell + Desktop Shell scenes registered",
			Time:     "11:40",
		},
		{
			Title:    "Command palette shipped",
			Subtitle: "Ctrl+Shift+P in Desktop Shell (Go)",
			Time:     "14:05",
		},
		{
			Title:    "Build succeeded",
			Subtitle: "All tests passed · ready for review",
			Time:     "16:30",
		},
	}, 0, 0, 0, 0)
	panel.AddChild(tl)
	page.Body.AddChild(panel)

	FinishShellMount(doc)
}

func (s *timelineScene) OnUpdate(_ *ui.Document, _ float32) {}
