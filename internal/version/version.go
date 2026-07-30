// Package version holds release strings injected by cmd/gru at link time.
package version

// Defaults match cmd/gru/version.go; overridden via -ldflags on release builds.
var (
	Tool    = "0.1.0"
	App     = "0.1.0-dev"
	Product = "Gru Notepad"
	Module  = "github.com/ledocorp/gru"
)
