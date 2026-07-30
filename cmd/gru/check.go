package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

type checkOptions struct {
	skipLint     bool
	skipVuln     bool
	skipFullTest bool
	fullLint     bool
}

// runCheck mirrors scripts/check-local.ps1 — local quality gate without PowerShell.
func runCheck(args []string) error {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	skipLint := fs.Bool("skip-lint", false, "skip staticcheck / golangci-lint")
	skipVuln := fs.Bool("skip-vuln", false, "skip govulncheck")
	skipFullTest := fs.Bool("skip-full-test", false, "skip full go test ./ui/...")
	fullLint := fs.Bool("full-lint", false, "run golangci-lint or full staticcheck (not U1000 only)")
	fs.Usage = func() {
		fmt.Fprintln(os.Stderr, "usage: gru check [-skip-lint] [-skip-vuln] [-skip-full-test] [-full-lint]")
	}
	if err := fs.Parse(args); err != nil {
		return err
	}
	opts := checkOptions{
		skipLint:     *skipLint,
		skipVuln:     *skipVuln,
		skipFullTest: *skipFullTest,
		fullLint:     *fullLint,
	}
	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	lintPaths := []string{
		"./ui/...",
		"./examples/...",
		"./cmd/...",
		"./internal/...",
		".",
	}
	if err := checkStep(root, "go vet", "go", "vet", "./..."); err != nil {
		return err
	}
	if err := checkStep(root, "go build", "go", "build", "./..."); err != nil {
		return err
	}
	hasNotepad := fileExists(filepath.Join(root, "examples", "notepad_demo.go"))
	hasDemo := fileExists(filepath.Join(root, "examples", "public_catalog.go"))
	if hasNotepad {
		if err := checkStep(root, "go build notepad", "go", "build", "-tags", "notepad", "./..."); err != nil {
			return err
		}
	}
	if hasDemo {
		if err := checkStep(root, "go build grudemo", "go", "build", "-tags", "grudemo", "."); err != nil {
			return err
		}
		if err := checkStep(root, "public allowlist",
			"go", "test", "./examples/",
			"-run", "TestPublicAllowlist",
			"-count", "1", "-timeout", "5m"); err != nil {
			return err
		}
	}
	if fileExists(filepath.Join(root, "cmd", "hello", "main.go")) {
		if err := checkStep(root, "go build hello", "go", "build", "./cmd/hello"); err != nil {
			return err
		}
	}
	if !opts.skipFullTest {
		if err := checkStep(root, "go test ui", "go", "test", "./ui/...", "-count=1", "-timeout", "10m"); err != nil {
			return err
		}
	}
	if err := checkStep(root, "idle rails",
		"go", "test", "./ui/...",
		"-run", "Idle|Wake|Wheel|Dirty|SplitView|Carousel|LayoutOverrides|AppBarScroll|IdleQuick",
		"-count=1", "-timeout", "5m"); err != nil {
		return err
	}
	if !opts.skipVuln {
		fmt.Println("\n== govulncheck ==")
		exit, err := runGoTool(root, "golang.org/x/vuln/cmd/govulncheck", "govulncheck", "./...")
		if err != nil {
			return err
		}
		if exit == 3 {
			fmt.Fprintln(os.Stderr, "govulncheck: reachable vulnerabilities — see output above.")
			fmt.Fprintln(os.Stderr, "stdlib-only? Upgrade Go to 1.26.4+. Temporary: gru check -skip-vuln")
			return fmt.Errorf("govulncheck reported vulnerabilities (exit 3)")
		}
		if exit != 0 {
			return fmt.Errorf("govulncheck failed (exit %d)", exit)
		}
	}
	if !opts.skipLint {
		if err := runLint(root, opts.fullLint, lintPaths); err != nil {
			return err
		}
	}
	fmt.Println("\nAll local checks passed.")
	return nil
}

func checkStep(dir, name string, args ...string) error {
	fmt.Printf("\n== %s ==\n", name)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return fmt.Errorf("%s failed (exit %d)", name, ee.ExitCode())
		}
		return fmt.Errorf("%s: %w", name, err)
	}
	return nil
}

func runGoTool(dir, module, binName string, args ...string) (int, error) {
	if path, err := exec.LookPath(binName); err == nil {
		cmd := exec.Command(path, args...)
		cmd.Dir = dir
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				return ee.ExitCode(), nil
			}
			return -1, err
		}
		return 0, nil
	}
	cmd := exec.Command("go", append([]string{"run", module + "@latest"}, args...)...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			return ee.ExitCode(), nil
		}
		return -1, err
	}
	return 0, nil
}

func runLint(dir string, fullLint bool, paths []string) error {
	if fullLint {
		if _, err := exec.LookPath("golangci-lint"); err == nil {
			lintArgs := append([]string{"run"}, paths...)
			return checkStep(dir, "golangci-lint (full)", append([]string{"golangci-lint"}, lintArgs...)...)
		}
		fmt.Println("\n== staticcheck full (golangci-lint not installed) ==")
		for _, path := range paths {
			fmt.Printf("  staticcheck %s\n", path)
			exit, err := runGoTool(dir, "honnef.co/go/tools/cmd/staticcheck", "staticcheck", path)
			if err != nil {
				return err
			}
			if exit != 0 {
				return fmt.Errorf("staticcheck %s failed (exit %d)", path, exit)
			}
		}
		return nil
	}
	fmt.Println("\n== staticcheck U1000 (hygiene baseline) ==")
	for _, path := range []string{"./ui/...", "./examples/..."} {
		fmt.Printf("  staticcheck -checks U1000 %s\n", path)
		exit, err := runGoTool(dir, "honnef.co/go/tools/cmd/staticcheck", "staticcheck", "-checks", "U1000", path)
		if err != nil {
			return err
		}
		if exit != 0 {
			return fmt.Errorf("staticcheck U1000 %s failed (exit %d)", path, exit)
		}
	}
	fmt.Println("Use -full-lint for golangci-lint / full staticcheck (SA rules).")
	return nil
}
