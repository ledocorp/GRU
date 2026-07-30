package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ledocorp/gru/internal/version"
)

// appProduct is a buildable / packagable Gru app surface.
type appProduct struct {
	id        string // hello | demo | notepad
	pkg       string // go package path relative to module root
	tags      string // build tags (comma-separated); linux may append ,x11
	binaryWin string
	binaryNix string
	product   string // ldflags Product
	needIcons bool   // regenerate packaging icons before build
	stageWeb  bool   // include assets/web in package
	stageDict bool   // include spell dict (Notepad)
}

var products = map[string]appProduct{
	"hello": {
		id:        "hello",
		pkg:       "./cmd/hello",
		tags:      "",
		binaryWin: "HelloGru",
		binaryNix: "hello-gru",
		product:   "Hello Gru",
		needIcons: false,
		stageWeb:  false,
		stageDict: false,
	},
	"calc": {
		id:        "calc",
		pkg:       "./cmd/calc",
		tags:      "",
		binaryWin: "GruCalc",
		binaryNix: "gru-calc",
		product:   "Gru Calculator",
		needIcons: false,
		stageWeb:  false,
		stageDict: false,
	},
	"leandemo": {
		id:        "leandemo",
		pkg:       "./cmd/leandemo",
		tags:      "",
		binaryWin: "GruLeanDemo",
		binaryNix: "gru-lean-demo",
		product:   "Gru Lean Public Demo",
		needIcons: false,
		stageWeb:  false,
		stageDict: false,
	},
	"webviewhello": {
		id:        "webviewhello",
		pkg:       "./cmd/webviewhello",
		tags:      "",
		binaryWin: "HelloWebView",
		binaryNix: "hello-webview",
		product:   "Hello WebView",
		needIcons: false,
		stageWeb:  false,
		stageDict: false,
	},
	"demo": {
		id:        "demo",
		pkg:       ".",
		tags:      "grudemo",
		binaryWin: "GruDemo",
		binaryNix: "gru-demo",
		product:   "Gru Demo",
		needIcons: false,
		stageWeb:  true,
		stageDict: false,
	},
	"notepad": {
		id:        "notepad",
		pkg:       ".",
		tags:      "notepad",
		binaryWin: "GruNotepad",
		binaryNix: "gru-notepad",
		product:   "Gru Notepad",
		needIcons: true,
		stageWeb:  false,
		stageDict: true,
	},
}

type buildTarget struct {
	goos   string
	goarch string
	ext    string
}

func runBuild(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing app (hello, calc, leandemo, demo, notepad) or legacy platform (windows, linux)")
	}

	appID, platform, rest, err := parseAppPlatform(args)
	if err != nil {
		return err
	}
	prod, ok := products[appID]
	if !ok {
		return fmt.Errorf("unknown app %q (use hello, demo, or notepad)", appID)
	}

	target, err := platformTarget(platform)
	if err != nil {
		return err
	}

	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	out := fs.String("o", "", "output file path")
	arch := fs.String("arch", target.goarch, "GOARCH (linux: amd64 or arm64)")
	webview2 := fs.Bool("webview2", false, "hello webview / demo: add webview2 build tag (Windows)")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *arch != "" {
		target.goarch = *arch
	}
	if *webview2 {
		switch prod.id {
		case "demo", "webviewhello":
			if prod.tags == "" {
				prod.tags = "webview2"
			} else {
				prod.tags = prod.tags + ",webview2"
			}
		default:
			return fmt.Errorf("-webview2 is only valid for gru build demo|webviewhello")
		}
	}

	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	outPath := *out
	if outPath == "" {
		dir := filepath.Join(repoRoot, "dist")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		bin := prod.binaryNix
		if target.goos == "windows" {
			bin = prod.binaryWin
		}
		outPath = filepath.Join(dir, bin+target.ext)
	} else {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return err
		}
	}

	if err := buildProduct(prod, target, outPath); err != nil {
		return err
	}
	fmt.Printf("ok: %s\n", outPath)
	if target.goos == "windows" && runtime.GOOS == "windows" {
		fmt.Println("note: unsigned Windows builds may trigger SmartScreen — see docs/PACKAGING_AND_CLI.md")
	}
	return nil
}

// parseAppPlatform accepts:
//
//	hello [windows|linux] …
//	demo [windows|linux] …
//	notepad [windows|linux] …
//	windows|linux …          (legacy → notepad)
func parseAppPlatform(args []string) (appID, platform string, rest []string, err error) {
	first := strings.ToLower(args[0])
	rest = args[1:]
	switch first {
	case "hello", "calc", "leandemo", "webviewhello", "demo", "notepad":
		appID = first
		if len(rest) > 0 {
			p := strings.ToLower(rest[0])
			if p == "windows" || p == "win" || p == "linux" {
				platform = p
				rest = rest[1:]
			}
		}
		if platform == "" {
			platform = hostPlatform()
		}
		return appID, platform, rest, nil
	case "windows", "win", "linux":
		fmt.Fprintln(os.Stderr, "note: gru build windows|linux defaults to notepad; prefer: gru build notepad windows")
		return "notepad", first, rest, nil
	default:
		return "", "", nil, fmt.Errorf("unknown %q — use: gru build hello|calc|leandemo|webviewhello|demo|notepad [windows|linux]", first)
	}
}

func hostPlatform() string {
	if runtime.GOOS == "windows" {
		return "windows"
	}
	return "linux"
}

func platformTarget(platform string) (buildTarget, error) {
	switch strings.ToLower(platform) {
	case "windows", "win":
		return buildTarget{goos: "windows", goarch: "amd64", ext: ".exe"}, nil
	case "linux":
		return buildTarget{goos: "linux", goarch: "amd64", ext: ""}, nil
	default:
		return buildTarget{}, fmt.Errorf("unsupported platform %q (use windows or linux)", platform)
	}
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	var fallback string
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			hasMain := fileExists(filepath.Join(dir, "main.go"))
			hasHello := fileExists(filepath.Join(dir, "cmd", "hello", "main.go"))
			if hasMain || hasHello {
				// Prefer the private monorepo (has staging drafts) over a nested
				// staging/export tree that also contains go.mod + main.go.
				if fileExists(filepath.Join(dir, "staging", "PUBLIC_README.md")) ||
					fileExists(filepath.Join(dir, "cmd", "gru", "export.go")) {
					return dir, nil
				}
				if fallback == "" {
					fallback = dir
				}
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	if fallback != "" {
		return fallback, nil
	}
	return "", fmt.Errorf("could not find Gru module root (go.mod + main.go or cmd/hello)")
}

func buildProduct(prod appProduct, target buildTarget, outPath string) error {
	if target.goos == "linux" && runtime.GOOS == "windows" {
		return fmt.Errorf("linux build must run inside WSL Ubuntu (CGO cross-compile from Windows is unsupported)")
	}
	repoRoot, err := findRepoRoot()
	if err != nil {
		return err
	}
	if prod.needIcons {
		if err := ensureAppIconAssets(repoRoot); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}

	tags := prod.tags
	if target.goos == "linux" {
		if tags == "" {
			tags = "x11"
		} else if !strings.Contains(tags, "x11") {
			tags = tags + ",x11"
		}
	}

	fmt.Printf("building %s %s/%s -> %s\n", prod.id, target.goos, target.goarch, outPath)
	args := []string{"build", "-trimpath", "-ldflags", productLDFlags(prod), "-o", outPath}
	if tags != "" {
		args = append(args, "-tags", tags)
	}
	args = append(args, prod.pkg)

	cmd := exec.Command("go", args...)
	cmd.Dir = repoRoot
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED=1",
		"GOOS="+target.goos,
		"GOARCH="+target.goarch,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func productLDFlags(prod appProduct) string {
	parts := []string{"-s", "-w"}
	for _, x := range []struct{ path, value string }{
		{"github.com/ledocorp/gru/internal/version.Tool", version.Tool},
		{"github.com/ledocorp/gru/internal/version.App", version.App},
		{"github.com/ledocorp/gru/internal/version.Product", prod.product},
		{"github.com/ledocorp/gru/internal/version.Module", version.Module},
	} {
		parts = append(parts, ldflagsX(x.path, x.value))
	}
	return strings.Join(parts, " ")
}

// buildNotepad keeps older call sites (package.go) working.
func buildNotepad(target buildTarget, outPath string) error {
	return buildProduct(products["notepad"], target, outPath)
}
