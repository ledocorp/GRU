package ui

import (
	"os"
	"path/filepath"
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestPhosphorFontTTFPath(t *testing.T) {
	dir := t.TempDir()
	weightDir := filepath.Join(dir, "Fonts", "regular")
	if err := os.MkdirAll(weightDir, 0o755); err != nil {
		t.Fatal(err)
	}
	ttf := filepath.Join(weightDir, "Phosphor.ttf")
	if err := os.WriteFile(ttf, []byte("fake"), 0o644); err != nil {
		t.Fatal(err)
	}
	r := NewPhosphorRegistry(dir)
	if got := r.FontTTFPath(PhosphorRegular); got != ttf {
		t.Fatalf("FontTTFPath = %q, want %q", got, ttf)
	}
	if got := r.FontTTFPath(PhosphorBold); got != "" {
		t.Fatalf("bold without file = %q", got)
	}
}

func TestPhosphorFontDrawSize(t *testing.T) {
	prev := BaseSupersamplingScale
	t.Cleanup(func() { BaseSupersamplingScale = prev })
	BaseSupersamplingScale = 2 // Path A-like — no sticky quantize
	dst := rl.NewRectangle(10, 20, 48, 40)
	if got := phosphorFontDrawSize(dst); got != 40 {
		t.Fatalf("size = %v, want 40 (min of w/h)", got)
	}
	dst = rl.NewRectangle(0, 0, 56, 56)
	if got := phosphorFontDrawSize(dst); got != 56 {
		t.Fatalf("size = %v, want 56", got)
	}
}

func TestPhosphorIconFontSummaryNoTTF(t *testing.T) {
	r := NewPhosphorRegistry(t.TempDir())
	if got := r.IconFontSummary(); got != "png/svg (remix off)" {
		t.Fatalf("summary = %q", got)
	}
}

func TestPhosphorFontDrawSizeNo40Cap(t *testing.T) {
	prev := BaseSupersamplingScale
	t.Cleanup(func() { BaseSupersamplingScale = prev })
	BaseSupersamplingScale = 2
	dst := rl.NewRectangle(0, 0, 56, 56)
	got := phosphorFontDrawSize(dst)
	if got <= 40 {
		t.Fatalf("expected sizes above legacy 40px cap to pass through, got %v", got)
	}
}

func TestQuantizeIconDrawSizeSticky(t *testing.T) {
	// Bias favors larger sticky: 19 → 20 (not 18).
	if got := quantizeIconDrawSize(19); got != 20 {
		t.Fatalf("19 → %v, want sticky 20 (larger bias)", got)
	}
	if got := quantizeIconDrawSize(24); got != 24 {
		t.Fatalf("exact 24 → %v", got)
	}
	prev := BaseSupersamplingScale
	t.Cleanup(func() { BaseSupersamplingScale = prev })
	BaseSupersamplingScale = 1
	dst := rl.NewRectangle(0, 0, 19, 19)
	got := phosphorFontDrawSize(dst)
	if got != 20 {
		t.Fatalf("SSAA1× draw size for 19px slot = %v, want sticky 20", got)
	}
}

func TestIconStrikeWarmSizesOnLadder(t *testing.T) {
	for _, sz := range iconStrikeWarmSizes {
		found := false
		for _, s := range stickyIconDrawSizes {
			if int32(s) == sz {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("warm strike size %d not on stickyIconDrawSizes", sz)
		}
	}
}
