// Asset loading: desktop resolves paths beside the executable; Android uses rl.OpenAsset.
// See docs/ANDROID_CODE.md §5 and docs/NOTEPAD_RELEASE.md (release layout).
package ui

import (
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	assetRootsOnce sync.Once
	assetRoots     []string
)

func initAssetRoots() {
	var roots []string
	seen := map[string]struct{}{}
	add := func(dir string) {
		if dir == "" {
			return
		}
		abs, err := filepath.Abs(dir)
		if err != nil {
			abs = dir
		}
		if _, ok := seen[abs]; ok {
			return
		}
		seen[abs] = struct{}{}
		roots = append(roots, abs)
	}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		add(dir)
		// dist/GruDemo.exe → repo root; also walk a few parents for nested out dirs
		for i := 0; i < 4; i++ {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
			add(dir)
			if st, err := os.Stat(filepath.Join(dir, "assets", "fonts", "remixicon.ttf")); err == nil && !st.IsDir() {
				break
			}
		}
	}
	if wd, err := os.Getwd(); err == nil {
		add(wd)
		dir := wd
		for i := 0; i < 4; i++ {
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
			add(dir)
			if st, err := os.Stat(filepath.Join(dir, "assets", "fonts", "remixicon.ttf")); err == nil && !st.IsDir() {
				break
			}
		}
	}
	assetRoots = roots
}

// ResolveAssetPath returns the first existing file for a repo-relative asset path
// (e.g. assets/fonts/remixicon.css). Search order: path as-is, then each root + path.
func ResolveAssetPath(path string) string {
	if runtime.GOOS == "android" {
		return path
	}
	if st, err := os.Stat(path); err == nil && !st.IsDir() {
		if abs, err := filepath.Abs(path); err == nil {
			return abs
		}
		return path
	}
	assetRootsOnce.Do(initAssetRoots)
	clean := filepath.FromSlash(path)
	if wd, err := os.Getwd(); err == nil {
		candidate := filepath.Join(wd, clean)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
	}
	for _, root := range assetRoots {
		candidate := filepath.Join(root, clean)
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate
		}
	}
	return path
}

// ReadAssetFile loads a file from disk (desktop) or APK assets (Android).
// Paths are usually repo-relative, e.g. assets/fonts/foo.ttf.
func ReadAssetFile(path string) ([]byte, error) {
	if runtime.GOOS != "android" {
		return os.ReadFile(ResolveAssetPath(path))
	}
	candidates := assetPathCandidates(path)
	var lastErr error
	for _, p := range candidates {
		a, err := rl.OpenAsset(p)
		if err != nil {
			lastErr = err
			continue
		}
		data, err := io.ReadAll(a)
		_ = a.Close()
		if err != nil {
			lastErr = err
			continue
		}
		return data, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, os.ErrNotExist
}

func assetPathCandidates(path string) []string {
	trimmed := strings.TrimPrefix(path, "assets/")
	if trimmed == path {
		return []string{path}
	}
	return []string{trimmed, path}
}

// AndroidGLES returns true on Android where desktop GLSL 330 / MSAA are unsafe.
func AndroidGLES() bool { return runtime.GOOS == "android" }
