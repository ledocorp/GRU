// Package examples — shared file-dialog helpers (all platforms).
package examples

import (
	"errors"
	"path/filepath"
)

// ErrFileDialogCancelled is returned when the user dismisses a native dialog.
var ErrFileDialogCancelled = errors.New("file dialog cancelled")

// ErrFileDialogUnavailable is returned on platforms without native pickers yet (Android v1).
var ErrFileDialogUnavailable = errors.New("file dialog unavailable on this platform")

// DefaultSaveName returns a filename hint from the current path or untitled.txt.
func DefaultSaveName(path string) string {
	if path != "" {
		return filepath.Base(path)
	}
	return "Untitled.txt"
}
