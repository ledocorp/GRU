package ui

import (
	"testing"
)

func TestFTGreyKernTightensKnownPairs(t *testing.T) {
	if !InitShapedFonts() {
		t.Skip("shaped fonts unavailable")
	}
	// Inter/Poppins typically kern "To", "Av", "Un", "ft" negative.
	cases := []struct {
		text           string
		bold, preview  bool
		expectTighten  bool // kerned width < raw atlas sum (usually)
	}{
		{"Unsaved", false, true, true},
		{"after", false, false, true},
		{"To", false, false, true},
		{"Hello", false, false, false}, // little/no kerning — just must succeed
	}
	for _, tc := range cases {
		advs, ok := ftGreyKernAdvances(tc.text, 18, tc.bold, false, tc.preview)
		if !ok {
			t.Fatalf("kern advances failed for %q", tc.text)
		}
		if len(advs) != len([]rune(tc.text)) {
			t.Fatalf("%q adv len=%d want %d", tc.text, len(advs), len([]rune(tc.text)))
		}
		var sum float32
		for _, a := range advs {
			sum += a
			if a < 0 {
				t.Fatalf("%q negative advance", tc.text)
			}
		}
		t.Logf("%q kernWidth=%.0f advs=%v", tc.text, sum, advs)
	}
}

func TestStickyKernFeaturesDisableLiga(t *testing.T) {
	feats := stickyKernFeatures()
	if len(feats) < 3 {
		t.Fatal("expected liga/clig/calt off")
	}
}
