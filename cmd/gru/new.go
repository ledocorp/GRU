package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode"
)

// runNew scaffolds a native or WebView hello starter under cmd/ and samples/.
//
//	gru new myapp
//	gru new myapp --webview
func runNew(args []string) error {
	fs := flag.NewFlagSet("new", flag.ContinueOnError)
	webview := fs.Bool("webview", false, "scaffold WebView shell (chrome + HTML body)")
	fs.SetOutput(os.Stderr)
	if err := fs.Parse(args); err != nil {
		return err
	}
	rest := fs.Args()
	if len(rest) != 1 {
		return fmt.Errorf("usage: gru new <name> [--webview]")
	}
	name, err := sanitizeAppName(rest[0])
	if err != nil {
		return err
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}

	if *webview {
		return scaffoldFrom(repoRoot, name, "webviewhello", "webviewhello", true)
	}
	return scaffoldFrom(repoRoot, name, "hello", "hello", false)
}

func sanitizeAppName(raw string) (string, error) {
	s := strings.TrimSpace(strings.ToLower(raw))
	s = strings.ReplaceAll(s, "-", "")
	s = strings.ReplaceAll(s, "_", "")
	if s == "" {
		return "", fmt.Errorf("empty app name")
	}
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			return "", fmt.Errorf("app name %q must be letters/digits only", raw)
		}
	}
	if !unicode.IsLetter(rune(s[0])) {
		return "", fmt.Errorf("app name must start with a letter")
	}
	reserved := map[string]bool{
		"hello": true, "calc": true, "leandemo": true, "demo": true,
		"notepad": true, "webviewhello": true, "ui": true, "examples": true,
	}
	if reserved[s] {
		return "", fmt.Errorf("%q is a reserved product id", s)
	}
	return s, nil
}

func scaffoldFrom(repoRoot, name, cmdSrc, sampleSrc string, webview bool) error {
	cmdDst := filepath.Join(repoRoot, "cmd", name)
	sampleDst := filepath.Join(repoRoot, "samples", name)
	if dirExists(cmdDst) || dirExists(sampleDst) {
		return fmt.Errorf("%s already exists — choose another name", name)
	}

	cmdSrcDir := filepath.Join(repoRoot, "cmd", cmdSrc)
	sampleSrcDir := filepath.Join(repoRoot, "samples", sampleSrc)
	if err := copyDir(cmdSrcDir, cmdDst); err != nil {
		return err
	}
	if err := copyDir(sampleSrcDir, sampleDst); err != nil {
		_ = os.RemoveAll(cmdDst)
		return err
	}

	title := titleCaseName(name)
	repls := []struct{ old, new string }{
		{"github.com/ledocorp/gru/samples/" + sampleSrc, "github.com/ledocorp/gru/samples/" + name},
		{"package " + sampleSrc, "package " + name},
		{sampleSrc + ".", name + "."},
	}
	if webview {
		repls = append(repls,
			struct{ old, new string }{"Hello WebView", title},
			struct{ old, new string }{"webviewhello:", name + ":"},
		)
	} else {
		repls = append(repls,
			struct{ old, new string }{"Hello Gru", title},
			struct{ old, new string }{"Hello, Gru", title},
			struct{ old, new string }{"hello:", name + ":"},
		)
	}

	for _, dir := range []string{cmdDst, sampleDst} {
		if err := rewriteGoFiles(dir, repls); err != nil {
			return err
		}
	}

	kind := "native"
	runHint := fmt.Sprintf("go run ./cmd/%s", name)
	if webview {
		kind = "webview"
		runHint = fmt.Sprintf("go run -tags webview2 ./cmd/%s", name)
	}
	fmt.Printf("ok: created %s starter %q\n", kind, name)
	fmt.Printf("  cmd/%s   samples/%s\n", name, name)
	fmt.Printf("  %s\n", runHint)
	fmt.Println("  Import ui widgets / ui/syntax / ui/preview as needed — docs/CAPABILITY_GUIDE.md")
	return nil
}

func titleCaseName(name string) string {
	if name == "" {
		return name
	}
	r := []rune(name)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func dirExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && st.IsDir()
}

var goFileRe = regexp.MustCompile(`\.go$`)

func rewriteGoFiles(dir string, repls []struct{ old, new string }) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() || !goFileRe.MatchString(e.Name()) {
			continue
		}
		path := filepath.Join(dir, e.Name())
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		s := string(b)
		for _, r := range repls {
			s = strings.ReplaceAll(s, r.old, r.new)
		}
		if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
			return err
		}
	}
	return nil
}
