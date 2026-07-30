package ui

import "testing"

func TestMergeStyleJSONPaddingZero(t *testing.T) {
	zero := float32(0)
	merged := mergeStyleJSON(nil, styleJSON{Padding: &zero})
	if merged == nil || merged.Padding == nil || *merged.Padding != 0 {
		t.Fatalf("mergeStyleJSON padding = %v, want 0", merged)
	}
	base := CurrentTheme["card"].Padding
	st, err := merged.toStyle(Style{Padding: base})
	if err != nil {
		t.Fatal(err)
	}
	if st.Padding != 0 {
		t.Fatalf("toStyle padding = %.0f, want 0", st.Padding)
	}
}
