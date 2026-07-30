//go:build !notepad

package examples

import (
	"testing"

	"github.com/ledocorp/gru/ui"
)

// TestPublicAllowlistScenesRegistered ensures every frozen public title has a
// registered factory (smoke gate for grudemo).
func TestPublicAllowlistScenesRegistered(t *testing.T) {
	have := make(map[string]bool, RegisteredSceneCount())
	for i := 0; i < RegisteredSceneCount(); i++ {
		have[RegistryTitle(i)] = true
	}
	missing := make([]string, 0)
	for _, title := range PublicSceneTitles() {
		if !have[title] {
			missing = append(missing, title)
		}
	}
	if len(missing) > 0 {
		t.Fatalf("public allowlist titles not registered (%d): %v", len(missing), missing)
	}
	filtered := FilterPublicFactories(Registered())
	if len(filtered) != len(PublicSceneTitles()) {
		t.Fatalf("FilterPublicFactories returned %d, want %d", len(filtered), len(PublicSceneTitles()))
	}
}

// TestPublicAllowlistScenesBuild mounts each public scene once (headless doc).
// Catches panics and missing assets that would flash unfinished UI on Tab.
func TestPublicAllowlistScenesBuild(t *testing.T) {
	filtered := FilterPublicFactories(Registered())
	if len(filtered) == 0 {
		t.Fatal("no public factories")
	}
	for _, factory := range filtered {
		factory := factory
		sc := factory()
		title := sc.Title()
		t.Run(title, func(t *testing.T) {
			doc := ui.NewDocument(1280, 800)
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("Build panic: %v", r)
				}
				sc.Destroy()
			}()
			sc.Build(doc)
			if doc.Root == nil || len(doc.Root.Children()) == 0 {
				t.Fatal("Build left empty document root")
			}
		})
	}
}
