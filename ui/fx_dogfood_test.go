package ui

import (
	"os"
	"testing"

	"github.com/ledocorp/gru/ui/fxglyph"
	"github.com/ledocorp/gru/ui/fxmeasure"
)

func TestFxDogfoodHUD(t *testing.T) {
	h := FxDogfoodHUD()
	if h == "" {
		t.Fatal("empty HUD")
	}
	if EditorWrapPathName() != "pod-go" && EditorWrapPathName() != "pod-fx" {
		t.Fatalf("unexpected wrap path %q", EditorWrapPathName())
	}
}

func TestFxDogfoodHUDGlyphOptOut(t *testing.T) {
	if !fxglyph.Enabled() {
		t.Skip("fx not linked")
	}
	prev := os.Getenv("GRU_FX_GLYPH")
	t.Setenv("GRU_FX_GLYPH", "0")
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("GRU_FX_GLYPH")
		} else {
			_ = os.Setenv("GRU_FX_GLYPH", prev)
		}
	})
	if fxGlyphBackendActive() {
		t.Fatal("glyph fx should be off when GRU_FX_GLYPH=0")
	}
	if fxGlyphDogfoodState() != "gg-fallback" {
		t.Fatalf("state=%q want gg-fallback", fxGlyphDogfoodState())
	}
}

func TestFxDogfoodHUDWrapFXDefault(t *testing.T) {
	if !fxmeasure.Enabled() {
		t.Skip("fx not linked")
	}
	prev := os.Getenv("GRU_FX_WRAP")
	_ = os.Unsetenv("GRU_FX_WRAP")
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("GRU_FX_WRAP")
		} else {
			_ = os.Setenv("GRU_FX_WRAP", prev)
		}
	})
	if EditorWrapPathName() != "pod-fx" {
		t.Fatalf("path=%q want pod-fx by default", EditorWrapPathName())
	}
	if FxDogfoodHUD() == "" {
		t.Fatal("empty")
	}
}

func TestFxDogfoodHUDWrapGoOptOut(t *testing.T) {
	if !fxmeasure.Enabled() {
		t.Skip("fx not linked")
	}
	prev := os.Getenv("GRU_FX_WRAP")
	t.Setenv("GRU_FX_WRAP", "0")
	t.Cleanup(func() {
		if prev == "" {
			_ = os.Unsetenv("GRU_FX_WRAP")
		} else {
			_ = os.Setenv("GRU_FX_WRAP", prev)
		}
	})
	if EditorWrapPathName() != "pod-go" {
		t.Fatalf("path=%q want pod-go when GRU_FX_WRAP=0", EditorWrapPathName())
	}
}
