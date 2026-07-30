package examples

import "github.com/ledocorp/gru/ui"

// batchCaption is panel helper copy — transparent body text (not "default",
// which paints a white bordered box behind the string).
func batchCaption(id, text string) *ui.RichText {
	return FlexCopy(id, "form-value", text)
}
