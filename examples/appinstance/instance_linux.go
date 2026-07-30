//go:build linux

package appinstance

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/ledocorp/gru/x11util"

	"golang.org/x/sys/unix"
)

func startup(forwardPath string, onOpen func(path string)) (exit bool, stop func()) {
	lockPath := instanceLockPath()
	if lockPath == "" {
		return false, func() {}
	}
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return false, func() {}
	}

	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return false, func() {}
	}

	if err := unix.Flock(int(lockFile.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = lockFile.Close()
		_ = writePendingOpen(forwardPath)
		_ = x11util.RaiseNotepadWindow()
		return true, func() {}
	}

	var stopPoll func()
	var once sync.Once
	release := func() {
		once.Do(func() {
			if stopPoll != nil {
				stopPoll()
			}
			_ = unix.Flock(int(lockFile.Fd()), unix.LOCK_UN)
			_ = lockFile.Close()
		})
	}

	stopPoll = startPendingPoll(onOpen)
	return false, release
}

func instanceLockPath() string {
	if file := pendingOpenFile(); file != "" {
		return filepath.Join(filepath.Dir(file), "instance.lock")
	}
	if rd := os.Getenv("XDG_RUNTIME_DIR"); rd != "" {
		return filepath.Join(rd, "gru-notepad.lock")
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "gru-notepad", "instance.lock")
}
