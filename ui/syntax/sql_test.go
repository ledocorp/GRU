package syntax

import (
	"testing"

	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestHighlightSQLColors(t *testing.T) {
	editorBg := ui.GetThemeStyle("text-editor").BackgroundColor
	for _, theme := range []string{"github", "monokai", "pygments", "vs"} {
		ui.SetChromaStyle(theme)
		src := "SELECT id, name FROM users WHERE id = 1; -- end"
		spans := ui.HighlightSyntax(src, "sql")
		colors := map[rl.Color]int{}
		badContrast := 0
		for _, sp := range spans {
			if sp.Color.A == 0 {
				continue
			}
			if !syntaxColorContrastsEditor(sp.Color, editorBg) {
				badContrast += len(sp.Text)
			}
			colors[sp.Color] += len(sp.Text)
		}
		if badContrast > 0 {
			t.Fatalf("%s: %d bytes with failing contrast on editor bg", theme, badContrast)
		}
		if len(colors) < 2 {
			t.Fatalf("%s: expected multiple syntax colors, got %d", theme, len(colors))
		}
	}
}
