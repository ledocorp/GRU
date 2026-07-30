package ui

import "testing"

func TestOverlayContentBandRespectsInsets(t *testing.T) {
	SetOverlayChromeInsets(48, 36)
	band := OverlayContentBand(800, 600)
	if band.Y != 48 {
		t.Fatalf("band.Y = %v want 48", band.Y)
	}
	if band.Height != 600-48-36 {
		t.Fatalf("band.Height = %v want %v", band.Height, 600-48-36)
	}
	SetOverlayChromeInsets(0, 0)
}
