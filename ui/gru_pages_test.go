package ui

import (
	"os"
	"path/filepath"
	"testing"
)

// TestPagesGRUCompile ensures shipped pages/*.gru files parse and compile.
// Run from module root: go test ./ui -run TestPagesGRUCompile
func TestPagesGRUCompile(t *testing.T) {
	ctx := NewBuildContext()
	ctx.Actions["noop"] = func() {}
	ctx.Actions["saveSettings"] = func() {}
	ctx.Actions["resetSettings"] = func() {}
	ctx.Actions["refreshMetrics"] = func() {}
	ctx.Actions["openDocs"] = func() {}
	ctx.Actions["signIn"] = func() {}
	ctx.Actions["openSettings"] = func() {}

	pages := []string{
		"example.gru",
		"authoring.gru",
		"settings.gru",
		"dashboard.gru",
		"docs.gru",
		"gallery.gru",
		"sign-in.gru",
		"profile.gru",
	}
	for _, name := range pages {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join("pages", name)
			if _, err := os.Stat(path); err != nil {
				t.Skipf("pages/%s not found (run tests from module root)", name)
			}
			node, err := LoadGRU(path, ctx)
			if err != nil {
				t.Fatalf("LoadGRU(%s): %v", path, err)
			}
			if node == nil {
				t.Fatalf("LoadGRU(%s): nil root", path)
			}
		})
	}

	templates := []string{
		"settings-form-card.gru",
		"actions-card.gru",
		"metric-card.gru",
		"callout-and-code.gru",
		"status-footer.gru",
		"controls-row.gru",
		"sign-in-form.gru",
		"badges-row.gru",
		"progress-field.gru",
	}
	for _, name := range templates {
		t.Run("templates/"+name, func(t *testing.T) {
			path := filepath.Join("pages", "templates", name)
			if _, err := os.Stat(path); err != nil {
				t.Skipf("pages/templates/%s not found (run tests from module root)", name)
			}
			node, err := LoadGRU(path, ctx)
			if err != nil {
				t.Fatalf("LoadGRU(%s): %v", path, err)
			}
			if node == nil {
				t.Fatalf("LoadGRU(%s): nil root", path)
			}
		})
	}
}
