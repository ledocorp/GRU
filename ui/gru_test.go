package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGRU(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.gru")
	const json = `{"id":"t","children":[{"type":"section","id":"s","title":"S","children":[{"type":"text","id":"x","text":"hi"}]}]}`
	if err := os.WriteFile(path, []byte(json), 0o644); err != nil {
		t.Fatal(err)
	}
	node, err := LoadGRU(path, NewBuildContext())
	if err != nil {
		t.Fatalf("LoadGRU: %v", err)
	}
	if node == nil {
		t.Fatal("nil node")
	}
}

func TestReadGRUFilePagesFallback(t *testing.T) {
	dir := t.TempDir()
	pages := filepath.Join(dir, "pages")
	if err := os.Mkdir(pages, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pages, "x.gru"), []byte(`{"id":"x"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	if _, err := ReadGRUFile("x.gru"); err != nil {
		t.Fatalf("ReadGRUFile fallback: %v", err)
	}
}
