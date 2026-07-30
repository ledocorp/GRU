package main

import (
	"github.com/ledocorp/gru/ui"
	"fmt"
	"os"
)

// applyTextEngineMode configures shaped vs SDF per T1.6-R.
//
// Notepad release (-tags notepad) defaults to shaped text. Dev builds opt in with
// GRU_SHAPED_TEXT=1 (GORY_SHAPED_TEXT alias). Set GRU_SHAPED_TEXT=0 to force SDF (regression compare).
func applyTextEngineMode() {
	switch shapedTextEnv() {
	case "0":
		return
	case "1":
		ui.SetTextEngineMode(ui.TextEngineShaped)
	default:
		if appReleaseMode() {
			ui.SetTextEngineMode(ui.TextEngineShaped)
		}
	}
}

func shapedTextEnv() string {
	if v := os.Getenv("GRU_SHAPED_TEXT"); v != "" {
		return v
	}
	return os.Getenv("GORY_SHAPED_TEXT")
}

// notepadReleaseRenderLine is the one stderr ship-gate render summary (-tags notepad).
func notepadReleaseRenderLine() string {
	return fmt.Sprintf("Gru Notepad render: text=%s icons=%s ssaa=%.1fx dpi=%.2f sdfAtlas=%d",
		ui.TextEngineBackendName(), ui.Phosphor.IconFontSummary(), ui.EffectiveSupersamplingScale(),
		ui.DisplayScale, ui.EffectiveSDFAtlasSize())
}
