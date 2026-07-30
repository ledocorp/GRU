package spell

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeAffForGoSpell(t *testing.T) {
	dir := t.TempDir()
	aff := filepath.Join(dir, "test.aff")
	if err := os.WriteFile(aff, []byte("SET UTF8\nTRY abc\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	path, cleanup, err := normalizeAffForGoSpell(aff)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(raw) != "SET UTF-8\nTRY abc\n" {
		t.Fatalf("got %q", raw)
	}
}
