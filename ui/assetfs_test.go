package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveAssetPathBesideCWD(t *testing.T) {
	dir := t.TempDir()
	assetDir := filepath.Join(dir, "assets", "fonts")
	if err := os.MkdirAll(assetDir, 0o755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(assetDir, "test.txt")
	if err := os.WriteFile(file, []byte("ok"), 0o644); err != nil {
		t.Fatal(err)
	}
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(orig)

	got := ResolveAssetPath("assets/fonts/test.txt")
	if got != file {
		t.Fatalf("ResolveAssetPath = %q, want %q", got, file)
	}
}
