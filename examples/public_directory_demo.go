//go:build !notepad

package examples

import (
	"fmt"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &publicDirectoryScene{} }) }

// publicDirectoryScene is the grudemo scene picker (not Studio Foundry Directory).
type publicDirectoryScene struct {
	BaseScene
}

func (s *publicDirectoryScene) Title() string { return PublicDirectorySceneTitle }

func (s *publicDirectoryScene) Build(doc *ui.Document) {
	page := MountAppPage(doc, "pubdir",
		"Demo Index",
		"Pick a curated scene — or press Tab to cycle. F12 opens the Inspector.")
	page.Body.Gap = 10

	page.Body.AddChild(ui.NewPlainText("pubdir-hint", "form-label",
		fmt.Sprintf("%d public demos · click a row to open", len(PublicSceneTitles())),
		0, 0, 0, 0))

	list := ui.NewCard("pubdir-list", "Scenes", 0, 0, 0, 0)
	list.Gap = 4
	page.Body.AddChild(list)

	for _, title := range PublicSceneTitles() {
		title := title
		if title == PublicDirectorySceneTitle {
			continue
		}
		tile := ui.NewListTile("pubdir-"+title, title, "Open", 0, 0, 0, 0)
		tile.OnClick = func() {
			if NavigateToScene(title) {
				ui.ShowToast(title, ui.ToastInfo, 1200)
			}
		}
		list.AddChild(tile)
	}

	FinishShellMount(doc)
}
