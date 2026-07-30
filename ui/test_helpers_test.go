package ui

import "testing"

// findNodeByID walks the subtree for the first node with the given id.
func findNodeByID(root Node, id string) Node {
	if root == nil {
		return nil
	}
	if root.ID() == id {
		return root
	}
	for _, ch := range root.Children() {
		if found := findNodeByID(ch, id); found != nil {
			return found
		}
	}
	return nil
}

// ensureTestFonts loads shaped faces for unit tests that need non-zero text measure.
// Skips when no system TTF is available (headless CI without fonts).
func ensureTestFonts(t *testing.T) {
	t.Helper()
	if !InitShapedFonts() {
		t.Skip("no TTF faces for text measure")
	}
	SetTextEngineMode(TextEngineShaped)
}
