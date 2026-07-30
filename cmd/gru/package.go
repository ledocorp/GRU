package main

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ledocorp/gru/internal/version"
)

func runPackage(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("missing app (hello, calc, leandemo, demo, notepad) or legacy platform (windows, linux)")
	}
	appID, platform, rest, err := parseAppPlatform(args)
	if err != nil {
		return err
	}
	if len(rest) > 0 {
		return fmt.Errorf("unexpected args after platform: %v", rest)
	}
	prod, ok := products[appID]
	if !ok {
		return fmt.Errorf("unknown app %q", appID)
	}
	target, err := platformTarget(platform)
	if err != nil {
		return err
	}
	switch target.goos {
	case "windows":
		return packageProductWindows(prod)
	case "linux":
		return packageProductLinux(prod)
	default:
		return fmt.Errorf("unsupported GOOS %s", target.goos)
	}
}

func packageProductWindows(prod appProduct) error {
	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	target := buildTarget{goos: "windows", goarch: "amd64", ext: ".exe"}
	out := filepath.Join(root, "dist", prod.binaryWin+target.ext)
	if err := buildProduct(prod, target, out); err != nil {
		return err
	}
	dist := filepath.Join(root, "dist")
	stageDir := filepath.Join(dist, "stage-"+prod.id+"-win")
	_ = os.RemoveAll(stageDir)
	if err := os.MkdirAll(stageDir, 0o755); err != nil {
		return err
	}
	if err := copyFile(out, filepath.Join(stageDir, prod.binaryWin+target.ext)); err != nil {
		return err
	}
	if err := stagePublicRuntime(root, stageDir, prod); err != nil {
		return err
	}
	zipPath := filepath.Join(dist, fmt.Sprintf("%s-win-amd64.zip", prod.binaryWin))
	entries := []string{prod.binaryWin + target.ext, "assets"}
	if fileExists(filepath.Join(stageDir, "README.txt")) {
		entries = append(entries, "README.txt")
	}
	if prod.stageDict {
		entries = append(entries, "dict")
	}
	if prod.needIcons {
		entries = append(entries, "icons")
	}
	if err := writeZip(zipPath, stageDir, entries); err != nil {
		return err
	}
	fmt.Printf("package: %s\n", zipPath)
	return nil
}

func packageProductLinux(prod appProduct) error {
	if runtime.GOOS == "windows" {
		return fmt.Errorf("linux package must be built inside WSL Ubuntu (CGO cross-compile from Windows is unsupported)\nSee docs/NOTEPAD_SHIP_GATE.md")
	}
	root, err := findRepoRoot()
	if err != nil {
		return err
	}
	target := buildTarget{goos: "linux", goarch: "amd64", ext: ""}
	out := filepath.Join(root, "dist", prod.binaryNix)
	if err := buildProduct(prod, target, out); err != nil {
		return err
	}
	stage := filepath.Join(root, "dist", prod.binaryNix+"-linux-amd64")
	if err := os.RemoveAll(stage); err != nil {
		return err
	}
	if err := os.MkdirAll(stage, 0o755); err != nil {
		return err
	}
	binaryPath := filepath.Join(stage, prod.binaryNix)
	if err := copyFile(out, binaryPath); err != nil {
		return err
	}
	if err := os.Chmod(binaryPath, 0o755); err != nil {
		fmt.Printf("note: chmod skipped (%v) — common on /mnt/d; tarball will still be executable\n", err)
	}
	if err := stagePublicRuntime(root, stage, prod); err != nil {
		return err
	}
	if prod.id == "notepad" {
		desktop := filepath.Join(root, "packaging", "gru-notepad.desktop")
		if fileExists(desktop) {
			_ = copyFile(desktop, filepath.Join(stage, "gru-notepad.desktop"))
		}
	}
	tarPath := filepath.Join(root, "dist", prod.binaryNix+"-linux-amd64.tar.gz")
	if err := writeTarGz(tarPath, stage, filepath.Base(stage)); err != nil {
		return err
	}
	fmt.Printf("package: %s\n", tarPath)
	return nil
}

func stagePublicRuntime(root, dest string, prod appProduct) error {
	if prod.id == "notepad" {
		if err := stageReleaseExtras(root, dest); err != nil {
			return err
		}
		if err := stageNotepadAssets(root, dest); err != nil {
			return err
		}
		return stageAppIcons(root, dest)
	}
	// hello / demo: fonts (+ web for demo)
	srcFonts := filepath.Join(root, "assets", "fonts")
	if _, err := os.Stat(srcFonts); err == nil {
		dst := filepath.Join(dest, "assets", "fonts")
		if err := copyRuntimeFontsDir(srcFonts, dst); err != nil {
			return err
		}
		fmt.Println("staged assets/fonts")
	} else {
		fmt.Println("note: assets/fonts missing — run from module root after sync-public-export")
	}
	if prod.stageWeb {
		srcWeb := filepath.Join(root, "assets", "web")
		if _, err := os.Stat(srcWeb); err == nil {
			if err := copyDir(srcWeb, filepath.Join(dest, "assets", "web")); err != nil {
				return err
			}
			fmt.Println("staged assets/web")
		}
		gallery := filepath.Join(root, "pages", "gallery.gru")
		if fileExists(gallery) {
			if err := os.MkdirAll(filepath.Join(dest, "pages"), 0o755); err != nil {
				return err
			}
			if err := copyFile(gallery, filepath.Join(dest, "pages", "gallery.gru")); err != nil {
				return err
			}
			fmt.Println("staged pages/gallery.gru")
		}
	}
	readme := filepath.Join(dest, "README.txt")
	_ = os.WriteFile(readme, []byte(fmt.Sprintf(
		"%s\n\nRun this binary with cwd = this folder so assets/ resolves.\nModule: %s\n",
		prod.product, version.Module,
	)), 0o644)
	return nil
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

func stageReleaseExtras(root, dest string) error {
	readme := filepath.Join(root, "packaging", "README-release.txt")
	if fileExists(readme) {
		if err := copyFile(readme, filepath.Join(dest, "README.txt")); err != nil {
			return err
		}
	}
	dictSrc := filepath.Join(root, "assets", "dict", "en_US")
	dictDst := filepath.Join(dest, "dict", "en_US")
	if fileExists(filepath.Join(dictSrc, "en_US.aff")) && fileExists(filepath.Join(dictSrc, "en_US.dic")) {
		if err := os.MkdirAll(dictDst, 0o755); err != nil {
			return err
		}
		for _, name := range []string{"en_US.aff", "en_US.dic"} {
			if err := copyFile(filepath.Join(dictSrc, name), filepath.Join(dictDst, name)); err != nil {
				return err
			}
		}
		fmt.Println("staged dict/en_US")
	} else {
		fmt.Println("note: no dictionary in assets/dict/en_US — run scripts/build/fetch-en-us-dict.ps1")
	}
	return nil
}

// stageNotepadAssets copies runtime fonts (Remix + editor mono) beside the binary.
// Phosphor PNG trees are omitted — Notepad uses Remix icon font only.
func stageNotepadAssets(root, dest string) error {
	src := filepath.Join(root, "assets", "fonts")
	dst := filepath.Join(dest, "assets", "fonts")
	if _, err := os.Stat(src); err != nil {
		fmt.Println("note: assets/fonts missing — icons may fail at runtime")
		return nil
	}
	if err := copyRuntimeFontsDir(src, dst); err != nil {
		return err
	}
	fmt.Println("staged assets/fonts")
	return nil
}

func releaseLDFlags() string {
	return productLDFlags(products["notepad"])
}

// ldflagsX formats a -X linker definition. Values with spaces must be quoted or
// Windows link fails (splits "Gru Notepad" into separate arguments).
func ldflagsX(importPath, value string) string {
	if strings.ContainsAny(value, " \t\r\n") {
		return fmt.Sprintf(`-X "%s=%s"`, importPath, value)
	}
	return fmt.Sprintf("-X %s=%s", importPath, value)
}

// stageAppIcons copies hicolor PNG for Linux .desktop Icon=gru-notepad.
func stageAppIcons(root, dest string) error {
	src := filepath.Join(root, "packaging", "icons", "hicolor", "256x256", "apps", "gru-notepad.png")
	if !fileExists(src) {
		fmt.Println("note: packaging/icons/.../gru-notepad.png missing — run: go run ./scripts/build/gen_app_icon.go")
		return nil
	}
	dst := filepath.Join(dest, "icons", "hicolor", "256x256", "apps", "gru-notepad.png")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := copyFile(src, dst); err != nil {
		return err
	}
	fmt.Println("staged icons/hicolor/256x256/apps/gru-notepad.png")
	return nil
}

func ensureAppIconAssets(root string) error {
	ico := filepath.Join(root, "packaging", "icons", "gru-notepad.ico")
	if fileExists(ico) {
		return nil
	}
	fmt.Println("generating app icon assets…")
	cmd := exec.Command("go", "run", "./scripts/build/gen_app_icon.go")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// copyRuntimeFontsDir copies only font/CSS assets needed at runtime (skips npm metadata).
func copyRuntimeFontsDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !runtimeFontFile(path) {
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return copyFile(path, target)
	})
}

func runtimeFontFile(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".ttf", ".otf", ".woff", ".woff2", ".css":
		return true
	default:
		return false
	}
}

func writeZip(zipPath, baseDir string, entries []string) error {
	_ = os.Remove(zipPath)
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	w := zip.NewWriter(f)
	defer w.Close()
	for _, entry := range entries {
		path := filepath.Join(baseDir, filepath.FromSlash(entry))
		if err := addToZip(w, baseDir, path); err != nil {
			return err
		}
	}
	return w.Close()
}

func addToZip(w *zip.Writer, baseDir, path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if info.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}
		for _, e := range entries {
			if err := addToZip(w, baseDir, filepath.Join(path, e.Name())); err != nil {
				return err
			}
		}
		return nil
	}
	rel, err := filepath.Rel(baseDir, path)
	if err != nil {
		return err
	}
	rel = filepath.ToSlash(rel)
	hdr, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	hdr.Name = rel
	hdr.Method = zip.Deflate
	out, err := w.CreateHeader(hdr)
	if err != nil {
		return err
	}
	in, err := os.Open(path)
	if err != nil {
		return err
	}
	defer in.Close()
	_, err = io.Copy(out, in)
	return err
}

func writeTarGz(tarPath, srcDir, prefix string) error {
	_ = os.Remove(tarPath)
	f, err := os.Create(tarPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	defer gz.Close()
	tw := tar.NewWriter(gz)
	defer tw.Close()
	return filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := filepath.ToSlash(filepath.Join(prefix, rel))
		hdr, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		hdr.Name = name
		if !info.IsDir() && strings.HasSuffix(name, "/gru-notepad") {
			hdr.Mode = 0o755
		}
		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(tw, in)
		return err
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

func fileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
