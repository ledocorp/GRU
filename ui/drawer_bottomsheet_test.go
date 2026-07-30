package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestDrawerPanelWidthClamped(t *testing.T) {
	d := &drawerManager{width: 900, host: DefaultOverlayHost(OverlayAnimSlideLeft)}
	d.host.Progress = 1
	w := d.panelWidth(400)
	if w > 400*0.85+0.1 {
		t.Fatalf("panel width should clamp to 85%% of screen, got %v", w)
	}
	if w < 200 {
		t.Fatalf("panel width below minimum: %v", w)
	}
}

func TestDrawerPanelRectSlidesFromLeft(t *testing.T) {
	d := &drawerManager{width: 280, host: DefaultOverlayHost(OverlayAnimSlideLeft)}
	d.host.Progress = 0
	r0 := d.panelRect(800, 600)
	if r0.X > -1 {
		t.Fatalf("closed drawer should start off-screen left, x=%v", r0.X)
	}
	d.host.Progress = 1
	r1 := d.panelRect(800, 600)
	if r1.X < 0 || r1.X > 1 {
		t.Fatalf("open drawer should sit at x=0, got %v", r1.X)
	}
	if r1.Width != 280 {
		t.Fatalf("width=%v", r1.Width)
	}
}

func TestDrawerContentInsets(t *testing.T) {
	d := &drawerManager{width: 280, host: DefaultOverlayHost(OverlayAnimSlideLeft)}
	d.host.Progress = 1
	d.host.SetContentInsets(48, 36)
	r := d.panelRect(800, 600)
	if r.Y < 47.5 || r.Y > 48.5 {
		t.Fatalf("panel y=%v want 48", r.Y)
	}
	wantH := float32(600 - 48 - 36)
	if r.Height < wantH-0.5 || r.Height > wantH+0.5 {
		t.Fatalf("panel height=%v want %v", r.Height, wantH)
	}
}

func TestBottomSheetHeightFraction(t *testing.T) {
	b := &bottomSheetManager{height: 0.5, host: DefaultOverlayHost(OverlayAnimSlideBottom)}
	h := b.sheetHeight(1000)
	if h < 499 || h > 501 {
		t.Fatalf("expected half screen height, got %v", h)
	}
}

func TestBottomSheetRectSlidesFromBottom(t *testing.T) {
	b := &bottomSheetManager{height: 300, host: DefaultOverlayHost(OverlayAnimSlideBottom)}
	b.host.Progress = 0
	r0 := b.sheetRect(400, 800)
	if r0.Y < 799 {
		t.Fatalf("closed sheet should start below viewport, y=%v", r0.Y)
	}
	b.host.Progress = 1
	r1 := b.sheetRect(400, 800)
	wantY := float32(800 - 300)
	if r1.Y < wantY-0.5 || r1.Y > wantY+0.5 {
		t.Fatalf("open sheet y=%v want %v", r1.Y, wantY)
	}
}

func TestDrawerOpenClose(t *testing.T) {
	old := DrawerMgr
	t.Cleanup(func() { DrawerMgr = old })
	DrawerMgr = &drawerManager{
		host:            DefaultOverlayHost(OverlayAnimSlideLeft),
		CloseOnBackdrop: true,
		CloseOnEscape:   true,
	}
	panel := NewContainer("drawer-test", 0, 0, 100, 100)
	panel.SetBounds(rl.NewRectangle(0, 0, 100, 100))
	OpenDrawer(panel)
	if !IsDrawerOpen() {
		t.Fatal("expected open")
	}
	CloseDrawer()
	if IsDrawerOpen() {
		t.Fatal("expected closed after CloseDrawer")
	}
}
