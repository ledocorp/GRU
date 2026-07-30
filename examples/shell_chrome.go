package examples

import (
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// shellFooterLift is the distance from the window bottom edge that reveals
// auto-hidden shell chrome (bottom nav, pagination bar). Clears taskbar overlap.
const shellFooterLift = float32(96)

// UpdateShellFooterAutoHide shows footer chrome when the pointer is near the
// bottom of the window; hides it otherwise so scroll content gets full height.
func UpdateShellFooterAutoHide(footer ui.Node, doc *ui.Document) {
	if footer == nil || doc == nil {
		return
	}
	mouse := rl.GetMousePosition()
	threshold := float32(doc.Height) - shellFooterLift
	nearBottom := mouse.Y >= threshold
	if nearBottom && footer.IsHidden() {
		footer.Show()
	} else if !nearBottom && !footer.IsHidden() {
		footer.Hide()
	}
}
