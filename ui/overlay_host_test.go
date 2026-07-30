package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestOverlayHostFadeScaleOpenClose(t *testing.T) {
	h := DefaultOverlayHost(OverlayAnimFadeScale)
	h.BeginOpen()
	if !h.Open || !h.IsOpen() {
		t.Fatal("expected open after BeginOpen")
	}
	if h.Alpha != 0 || h.Scale != h.ScaleFrom {
		t.Fatalf("fade scale should start at alpha=0 scale=%v, got alpha=%v scale=%v", h.ScaleFrom, h.Alpha, h.Scale)
	}

	h.AdvanceAnimation(0.2)
	if h.Alpha <= 0 {
		t.Fatal("expected alpha to advance")
	}

	h.BeginClose()
	if h.IsOpen() {
		t.Fatal("IsOpen should be false immediately after BeginClose")
	}
	if !h.Closing {
		t.Fatal("expected closing")
	}

	for i := 0; i < 20 && h.Open; i++ {
		h.AdvanceAnimation(0.05)
	}
	if h.Open {
		t.Fatal("expected fully closed after fade-out")
	}
}

func TestOverlayHostSlideProgress(t *testing.T) {
	h := DefaultOverlayHost(OverlayAnimSlideLeft)
	h.BeginOpen()
	h.AdvanceAnimation(h.SlideTime)
	if h.Progress < 0.999 {
		t.Fatalf("expected progress ~1, got %v", h.Progress)
	}

	h.BeginClose()
	for i := 0; i < 20 && h.Open; i++ {
		h.AdvanceAnimation(0.05)
	}
	if h.Open {
		t.Fatal("expected slide closed")
	}
}

func TestScaledCenterBox(t *testing.T) {
	box := rl.NewRectangle(100, 100, 200, 100)
	s := ScaledCenterBox(box, 0.5)
	if s.Width != 100 || s.Height != 50 {
		t.Fatalf("scaled size=%v×%v", s.Width, s.Height)
	}
	if s.X != 150 || s.Y != 125 {
		t.Fatalf("scaled origin=(%v,%v) want (150,125)", s.X, s.Y)
	}
}

func TestOverlayHostContentBand(t *testing.T) {
	h := DefaultOverlayHost(OverlayAnimSlideLeft)
	h.SetContentInsets(48, 36)
	band := h.ContentBand(800, 600)
	if band.Y != 48 {
		t.Fatalf("band y=%v", band.Y)
	}
	if band.Height != 516 {
		t.Fatalf("band h=%v want 516", band.Height)
	}
}

func TestOverlayHostSlidePanelRect(t *testing.T) {
	h := DefaultOverlayHost(OverlayAnimSlideLeft)
	band := rl.NewRectangle(0, 0, 800, 600)
	h.Progress = 0
	r0 := h.SlidePanelRect(band, 280)
	if r0.X > -1 {
		t.Fatalf("closed drawer off-screen, x=%v", r0.X)
	}
	h.Progress = 1
	r1 := h.SlidePanelRect(band, 280)
	if r1.X < 0 || r1.X > 1 {
		t.Fatalf("open drawer at x=0, got %v", r1.X)
	}
}

func TestOverlayHostSlideSheetRect(t *testing.T) {
	h := DefaultOverlayHost(OverlayAnimSlideBottom)
	h.Progress = 0
	r0 := h.SlideSheetRect(400, 800, 300)
	if r0.Y < 799 {
		t.Fatalf("closed sheet below viewport, y=%v", r0.Y)
	}
	h.Progress = 1
	r1 := h.SlideSheetRect(400, 800, 300)
	if r1.Y < 499 || r1.Y > 501 {
		t.Fatalf("open sheet y=%v want 500", r1.Y)
	}
}
