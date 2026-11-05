//go:build !notepad

package examples

// StartupInstance is a no-op outside the Notepad product SKU.
// Product single-instance lives in apps/notepad (notepad build tag).
func StartupInstance(forwardPath string) (exit bool, stop func()) {
	return false, func() {}
}

// SetInstanceOpenFileHandler is a no-op outside the Notepad product SKU.
func SetInstanceOpenFileHandler(func(string)) {}
