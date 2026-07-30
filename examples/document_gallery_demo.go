//go:build !notepad

package examples

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() { Register(func() Scene { return &documentGalleryScene{} }) }

// documentGalleryScene shows Theme v2 / DocumentSpec variants side by side (pages/gallery.gru).
type documentGalleryScene struct {
	BaseScene
	gru GRUPageReloader
}

func (s *documentGalleryScene) Title() string { return "Gallery (.gru)" }

func (s *documentGalleryScene) Destroy() { s.gru.Close() }

func (s *documentGalleryScene) OnUpdate(doc *ui.Document, _ float32) {
	s.gru.Poll(doc)
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		return
	}
	focusClickedTextInput(doc)
}

func (s *documentGalleryScene) Build(doc *ui.Document) {
	ctx := ui.NewBuildContext()
	ctx.LinkHandler = func(link string) {
		NavigateFromDemoLink(link)
	}
	ctx.Actions["noop"] = func() {}

	s.gru = GRUPageReloader{
		LogicalPath:      "pages/gallery.gru",
		Ctx:              ctx,
		PreserveControls: true,
	}

	compiled, err := s.gru.Compile()
	if err != nil {
		page := MountAppPage(doc, "doc-gallery-err", "Gallery (.gru)", "Load error")
		msg := ui.NewLabel("gallery-err", "Could not load pages/gallery.gru:\n"+err.Error()+"\n\nRun from the repo root so pages/ resolves.", 0, 0, 0, 0)
		msg.SetStyle("form-value")
		msg.Wrap = true
		msg.Align = ui.LabelAlignLeft
		card := ui.NewCard("gallery-err-card", "Missing gallery", 0, 0, 0, 0)
		card.SetFlexGrow(1)
		card.AddChild(msg)
		page.Body.AddChild(card)
		FinishShellMount(doc)
		return
	}

	s.gru.MountShell(doc, "doc-gallery",
		"Gallery (.gru)",
		"DocumentSpec control gallery — edit pages/gallery.gru to hot-reload.",
		compiled,
	)
}
