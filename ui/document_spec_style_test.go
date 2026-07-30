package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDocumentSpecStyleFieldsCompile(t *testing.T) {
	ctx := NewBuildContext()
	hex := "#F59E0B"
	pad := float32(12)
	fs := int32(16)
	bw := float32(2)
	br := float32(8)
	fg := float32(1)
	bold := true
	fd := float32(1)
	mfs := int32(12)
	border := "#B45309"

	style := DocBlockStyle{
		BackgroundColor: &hex,
		TextColor:       strPtr("#1F2937"),
		FontSize:        &fs,
		Padding:         &pad,
		BorderWidth:     &bw,
		BorderColor:     &border,
		FlexGrow:        &fg,
		CornerRadius:    &br,
		Bold:            &bold,
		FontDensity:     &fd,
		MinFontSize:     &mfs,
	}
	if _, err := style.compile(); err != nil {
		t.Fatalf("compile style: %v", err)
	}

	block := DocBlock{
		Type:  "card",
		ID:    "style-smoke",
		Title: "Styled",
		Style: &style,
		Children: []DocBlock{
			{Type: "text", ID: "t", Text: "body"},
		},
	}
	if err := validateDocBlock(block, "style-smoke"); err != nil {
		t.Fatalf("validate: %v", err)
	}
	if _, err := buildDocBlockAt(block, ctx, "style-smoke"); err != nil {
		t.Fatalf("build: %v", err)
	}
}

func TestDocumentSpecPresetsMatchRegistry(t *testing.T) {
	docPresets := []string{
		"surface-card", "callout-tip", "code-block", "neo-glow-card",
		"glass-panel", "glass-panel-dark", "glass-card",
		"primary-button", "ghost-button", "danger-button",
	}
	registered := ListPresetNames()
	regSet := make(map[string]bool, len(registered))
	for _, p := range registered {
		regSet[p] = true
	}
	for _, p := range docPresets {
		if !regSet[p] {
			t.Errorf("documented preset %q not in ListPresetNames()", p)
		}
	}
}

func TestGalleryGRUPresetsRegistered(t *testing.T) {
	path := filepath.Join("..", "pages", "gallery.gru")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skipf("gallery.gru not found: %v", err)
	}
	spec, err := ParseDocumentSpec(data)
	if err != nil {
		t.Fatalf("parse gallery.gru: %v", err)
	}
	for _, preset := range CollectDocumentSpecPresets(spec) {
		if _, ok := LookupPreset(preset); !ok {
			t.Errorf("gallery.gru references unknown preset %q", preset)
		}
	}
}

func TestThemeV2ComponentsUsedByDocumentSpec(t *testing.T) {
	// Components referenced by presets + common block styling in authoring docs.
	want := []string{"button", "card", "panel", "badge"}
	names := ThemeV2ComponentNames()
	set := make(map[string]bool, len(names))
	for _, n := range names {
		set[n] = true
	}
	for _, c := range want {
		if !set[c] {
			t.Errorf("Theme v2 missing component %q (DocumentSpec styling depends on it)", c)
		}
	}
}

func TestBadgeVariantsAvailableInTheme(t *testing.T) {
	want := []string{"default", "primary", "success", "warning", "danger", "info"}
	got := ThemeV2ComponentVariants("badge")
	set := make(map[string]bool, len(got))
	for _, v := range got {
		set[v] = true
	}
	for _, v := range want {
		if !set[v] {
			t.Errorf("badge variant %q missing from Theme v2", v)
		}
	}
}

func strPtr(s string) *string { return &s }
