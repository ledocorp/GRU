package spell

import "github.com/ledocorp/gru/ui"

func init() {
	ui.RegisterHunspellFactory(newSpellChecker)
}
