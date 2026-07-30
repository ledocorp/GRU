package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAssetPathFindsRemixIcon(t *testing.T) {
	// From package ui cwd, repo assets live one level up — ResolveAssetPath also
	// searches parent roots after the assetfs walk fix.
	candidates := []string{
		"assets/fonts/remixicon.ttf",
		filepath.Join("..", "assets", "fonts", "remixicon.ttf"),
	}
	var found string
	for _, c := range candidates {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			found = c
			break
		}
	}
	if found == "" {
		t.Skip("remixicon.ttf not found relative to test cwd")
	}
	got := ResolveAssetPath("assets/fonts/remixicon.ttf")
	if _, err := os.Stat(got); err != nil {
		// Fall back: explicit relative path must still work as-is
		got2 := ResolveAssetPath(filepath.ToSlash(found))
		if _, err2 := os.Stat(got2); err2 != nil {
			t.Fatalf("ResolveAssetPath(%q)=%q err=%v; fallback %q err=%v", "assets/fonts/remixicon.ttf", got, err, got2, err2)
		}
	}
}
