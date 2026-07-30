package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// BeginFrameCursor resets to the default arrow before the document Update pass.
// Interactive widgets (RichText, Icon, SplitView, …) override while hovered;
// without this reset, I-beam / hand cursors stick after the pointer leaves.
func BeginFrameCursor() {
	rl.SetMouseCursor(rl.MouseCursorDefault)
}
