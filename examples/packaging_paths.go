//go:build !notepad

package examples

import "github.com/ledocorp/gru/internal/appicon"

// AppIconPNG returns the baked OS app icon PNG path, or "" if missing.
// Regenerate with: go run ./scripts/build/gen_app_icon.go
func AppIconPNG() string { return appicon.PNGPath() }
