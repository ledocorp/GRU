package appinstance

import (
	"os"
	"path/filepath"
	"strings"
)

func pendingOpenFile() string {
	if SetPendingOpenPath != nil {
		if p := strings.TrimSpace(SetPendingOpenPath()); p != "" {
			return p
		}
	}
	return ""
}

func writePendingOpen(path string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return nil
	}
	file := pendingOpenFile()
	if file == "" {
		return os.ErrInvalid
	}
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		return err
	}
	tmp := file + ".tmp"
	if err := os.WriteFile(tmp, []byte(path), 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, file)
}

func consumePendingOpen() string {
	file := pendingOpenFile()
	if file == "" {
		return ""
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return ""
	}
	_ = os.Remove(file)
	return strings.TrimSpace(string(data))
}
