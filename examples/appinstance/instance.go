// Package appinstance forwards startup files to an already-running Gru Notepad.
package appinstance

// SetPendingOpenPath returns the pending_open.txt path for duplicate-launch handoff.
// Set from examples before Startup (Windows + Linux).
var SetPendingOpenPath func() string

// Startup runs once at process launch. When another instance is already running,
// the file path (if any) is forwarded and exit is true. Otherwise listen for
// later Open-with launches until stop is called.
func Startup(forwardPath string, onOpen func(path string)) (exit bool, stop func()) {
	return startup(forwardPath, onOpen)
}
