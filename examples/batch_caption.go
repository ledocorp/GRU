package examples

import "github.com/ledocorp/gru/ui"

// batchCaption is panel helper / hint copy — form-label (muted caption), not
// form-value (near-black 18pt readout). Using form-value made Counter Demo
// body look heavy next to the lean header-subtitle intro.
func batchCaption(id, text string) *ui.RichText {
	return FlexCopy(id, "form-label", text)
}
