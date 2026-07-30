// Package tray hosts a desktop notification-area icon (separate goroutine; safe with raylib/GLFW).
//
// LLM Prompt Template: "Add tray menu item X — extend tray_desktop.go onReady; keep Start/Stop in main after InitWindow."
package tray

// Config holds tray icon bytes and callbacks invoked from the systray thread.
type Config struct {
	Icon    []byte
	Tooltip string
	OnShow  func()
	OnQuit  func()
}
