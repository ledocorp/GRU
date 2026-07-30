package startuppath

import "testing"

func TestFirstFileArg(t *testing.T) {
	got := FirstFileArg([]string{"GruNotepad.exe", `C:\notes\a.txt`})
	if got != `C:\notes\a.txt` {
		t.Fatalf("got %q", got)
	}
}

func TestFirstFileArgSkipsFlags(t *testing.T) {
	got := FirstFileArg([]string{"gru-notepad", "-v", "/tmp/x.md"})
	if got != "/tmp/x.md" {
		t.Fatalf("got %q", got)
	}
}

func TestResolvePrefersCLI(t *testing.T) {
	got := Resolve([]string{"exe", "cli.txt"}, "env.txt")
	if got != "cli.txt" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveEnvFallback(t *testing.T) {
	got := Resolve([]string{"exe"}, "env.txt")
	if got != "env.txt" {
		t.Fatalf("got %q", got)
	}
}
