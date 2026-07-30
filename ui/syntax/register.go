package syntax

import "github.com/ledocorp/gru/ui"

func init() {
	curated := initCuratedStyles()
	ui.RegisterHighlightSyntax(highlightSyntax)
	ui.RegisterChromaStyle(setChromaStyle, chromaStyleName, curated)
}
