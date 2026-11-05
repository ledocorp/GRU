package ui

import "testing"

func TestChromeNeedsPostBlitOnlyCaptionButtons(t *testing.T) {
	tb := &TitleBar{}
	if tb.ChromeNeedsPostBlit() {
		t.Fatal("idle title bar must not post-blit")
	}
	tb.hoverMin = true
	if !tb.ChromeNeedsPostBlit() {
		t.Fatal("min hover must post-blit plates")
	}
	tb.hoverMin = false
	tb.hoverClose = true
	if !tb.ChromeNeedsPostBlit() {
		t.Fatal("close hover must post-blit plates")
	}
}

func TestCaptionHoverMask(t *testing.T) {
	tb := &TitleBar{hoverClose: true, hoverMax: true}
	if m := tb.CaptionHoverMask(); m != 3 {
		t.Fatalf("mask = %d, want 3 (close+max)", m)
	}
}
