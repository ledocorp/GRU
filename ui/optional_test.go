// Blank-import optional ui backends so package ui tests exercise real highlighting,
// goldmark DocumentSpec, and Hunspell when those packages are present.
//
// Must be package ui_test (not ui) to avoid an import cycle: ui → markdown → ui.
package ui_test

import (
	_ "github.com/ledocorp/gru/ui/markdown"
	_ "github.com/ledocorp/gru/ui/spell"
	_ "github.com/ledocorp/gru/ui/syntax"
)
