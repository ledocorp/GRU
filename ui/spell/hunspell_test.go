package spell

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ledocorp/gru/ui"
)

func TestTryHunspellCheckerFallbackPath(t *testing.T) {
	dir := t.TempDir()
	aff := filepath.Join(dir, "en_US.aff")
	dic := filepath.Join(dir, "en_US.dic")
	if err := os.WriteFile(aff, []byte(testHunspellAFF), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dic, []byte(testHunspellDIC), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GRU_DICT_DIR", dir)
	t.Setenv("GORY_DICT_DIR", "")

	c, err := ui.TryHunspellChecker("xyzzy")
	if err != nil {
		t.Fatalf("TryHunspellChecker: %v", err)
	}
	if !c.Check("hello") {
		t.Error("expected hello to pass")
	}
	if c.Check("helo") {
		t.Error("expected helo to fail")
	}
	if !c.Check("xyzzy") {
		t.Error("expected extra word xyzzy to pass")
	}
}

func TestTryHunspellCheckerRealDict(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Skip(err)
	}
	// Walk up to module root from ui/spell test package.
	root := wd
	for i := 0; i < 6; i++ {
		aff := filepath.Join(root, "assets", "dict", "en_US", "en_US.aff")
		dic := filepath.Join(root, "assets", "dict", "en_US", "en_US.dic")
		if fileExists(aff) && fileExists(dic) {
			t.Setenv("GRU_DICT_DIR", filepath.Join(root, "assets", "dict", "en_US"))
			t.Setenv("GORY_DICT_DIR", "")
			c, err := ui.TryHunspellChecker()
			if err != nil {
				t.Fatalf("TryHunspellChecker: %v", err)
			}
			if !c.Check("hello") {
				t.Error("expected hello to pass")
			}
			if c.Check("xyzzy") {
				t.Error("expected xyzzy to fail")
			}
			if c.Check("teh") {
				t.Error("expected teh to fail")
			}
			return
		}
		parent := filepath.Dir(root)
		if parent == root {
			break
		}
		root = parent
	}
	t.Skip("assets/dict/en_US not found (run scripts/build/fetch-en-us-dict.ps1)")
}

func TestTryHunspellCheckerMissingDict(t *testing.T) {
	t.Setenv("GRU_DICT_DIR", filepath.Join(t.TempDir(), "missing"))
	t.Setenv("GORY_DICT_DIR", "")
	if _, err := ui.TryHunspellChecker(); err != ui.ErrHunspellDictNotFound {
		t.Fatalf("got err %v, want ErrHunspellDictNotFound", err)
	}
}

func TestTryHunspellCheckerLegacyDictDirEnv(t *testing.T) {
	dir := t.TempDir()
	aff := filepath.Join(dir, "en_US.aff")
	dic := filepath.Join(dir, "en_US.dic")
	if err := os.WriteFile(aff, []byte(testHunspellAFF), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dic, []byte(testHunspellDIC), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GRU_DICT_DIR", "")
	t.Setenv("GORY_DICT_DIR", dir)

	c, err := ui.TryHunspellChecker()
	if err != nil {
		t.Fatalf("TryHunspellChecker: %v", err)
	}
	if !c.Check("hello") {
		t.Error("expected hello to pass via GORY_DICT_DIR alias")
	}
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}

// Minimal Hunspell fixture (hello + world).
const testHunspellAFF = `SET UTF-8
TRY esianrtolcdugmphbyfvkwzESIANRTOLCDUGMPHBYFVKWZ
`

const testHunspellDIC = `2
hello
world
`
