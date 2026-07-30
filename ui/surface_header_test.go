package ui

import "testing"

func TestSurfaceHeaderHeightUntitled(t *testing.T) {
	h := SurfaceHeader{Title: "", TitleHeight: 36, Mode: HeaderModeTitleBar}
	if h.Height() != 0 {
		t.Fatalf("untitled height = %v, want 0", h.Height())
	}
}

func TestSurfaceHeaderHeightTitled(t *testing.T) {
	h := SurfaceHeader{Title: "Settings", TitleHeight: 36, Mode: HeaderModeInset}
	if h.Height() != 36 {
		t.Fatalf("height = %v, want 36", h.Height())
	}
}

func TestSurfaceHeaderGlassDefer(t *testing.T) {
	h := SurfaceHeader{Title: "Glass", TitleHeight: 36, Mode: HeaderModeGlass}
	if !h.DefersUntilPostSheen() {
		t.Fatal("glass header should defer until post sheen")
	}
	h.Mode = HeaderModeTitleBar
	if h.DefersUntilPostSheen() {
		t.Fatal("title bar should not defer")
	}
}

func TestSurfaceHeaderModeNone(t *testing.T) {
	h := SurfaceHeader{Title: "ignored", TitleHeight: 36, Mode: HeaderModeNone}
	if h.Height() != 0 {
		t.Fatalf("HeaderModeNone height = %v, want 0", h.Height())
	}
}
