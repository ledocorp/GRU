// Package ui — .gru file loading (DocumentSpec JSON on disk).
// See docs/GO_FIRST_UI.md: .gru is declarative authoring; LoadGRU compiles to the
// same nodes as BuildDocumentSpec / hand-written Go.
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadGRUFile reads a .gru file (UTF-8 JSON DocumentSpec). Relative paths are
// tried as given and under pages/ (e.g. "settings.gru" → "pages/settings.gru").
// Run the app from the Gru module root so pages/*.gru resolve.
func ReadGRUFile(path string) ([]byte, error) {
	var lastErr error
	for _, p := range gruPathCandidates(path) {
		data, err := os.ReadFile(p)
		if err == nil {
			return data, nil
		}
		lastErr = err
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("no candidates")
	}
	return nil, fmt.Errorf("ui/gru: read %q: %w", path, lastErr)
}

// GRUResolvedPath returns the on-disk path ReadGRUFile would use (first candidate that exists).
func GRUResolvedPath(path string) (string, error) {
	for _, p := range gruPathCandidates(path) {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("ui/gru: resolve %q: file not found", path)
}

func gruPathCandidates(path string) []string {
	if filepath.IsAbs(path) {
		return []string{path}
	}
	norm := filepath.ToSlash(path)
	var out []string
	add := func(p string) {
		if p == "" {
			return
		}
		out = append(out, p)
	}
	add(path)
	if !strings.HasPrefix(norm, "pages/") {
		add(filepath.Join("pages", filepath.Base(path)))
	}
	// Walk cwd + executable parents so dist/*.exe still finds pages/*.gru
	var bases []string
	if wd, err := os.Getwd(); err == nil {
		bases = append(bases, wd)
	}
	if exe, err := os.Executable(); err == nil {
		bases = append(bases, filepath.Dir(exe))
	}
	seen := map[string]struct{}{}
	for _, base := range bases {
		dir := base
		for i := 0; i < 5; i++ {
			abs, err := filepath.Abs(dir)
			if err != nil {
				abs = dir
			}
			if _, ok := seen[abs]; ok {
				break
			}
			seen[abs] = struct{}{}
			add(filepath.Join(abs, path))
			if !strings.HasPrefix(norm, "pages/") {
				add(filepath.Join(abs, "pages", filepath.Base(path)))
			} else {
				add(filepath.Join(abs, filepath.FromSlash(norm)))
			}
			parent := filepath.Dir(dir)
			if parent == dir {
				break
			}
			dir = parent
		}
	}
	return out
}

// LoadGRU reads a .gru file and compiles it with LoadDocumentSpec.
//
// Example:
//
//	ctx := ui.NewBuildContext()
//	ctx.Actions["save"] = func() { snap := ctx.ControlSnapshot() }
//	root, err := ui.LoadGRU("pages/settings.gru", ctx)
func LoadGRU(path string, ctx *BuildContext) (Node, error) {
	data, err := ReadGRUFile(path)
	if err != nil {
		return nil, err
	}
	return LoadDocumentSpec(data, ctx)
}
