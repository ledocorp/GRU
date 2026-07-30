package markdown

import "github.com/ledocorp/gru/ui"

func init() {
	ui.RegisterMarkdownToDocumentSpec(markdownToDocumentSpecGoldmark)
}
