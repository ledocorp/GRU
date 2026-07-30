package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestBottomNavigationBarIntrinsicHeight(t *testing.T) {
	tab := NewSignal(0)
	bn := NewBottomNavigationBar("nav", []BottomNavItem{{Icon: "A", Label: "One"}}, tab, 0, 0, 200, 0)
	bn.SetBounds(rl.NewRectangle(0, 0, 200, 80))
	bn.Layout()
	if h := bn.Bounds().Height; h != bottomNavDefaultH {
		t.Fatalf("expected height %v, got %v", bottomNavDefaultH, h)
	}
}

func TestFABLayoutAnchorsBottomRight(t *testing.T) {
	body := NewContainer("body", 0, 0, 400, 300)
	fab := NewFAB("fab", "+", "", nil, 0, 0, 0, 0)
	fab.Anchor = body
	body.SetBounds(rl.NewRectangle(0, 0, 400, 300))
	fab.Layout()
	b := fab.Bounds()
	wantR := float32(400) - fabMargin
	wantB := float32(300) - fabMargin - fabBottomLift
	if b.X+b.Width > wantR+1 || b.X+b.Width < wantR-fabMiniDiameter-2 {
		t.Fatalf("FAB x=%v w=%v expected right edge near %v", b.X, b.Width, wantR)
	}
	if b.Y+b.Height > wantB+1 || b.Y+b.Height < wantB-fabMiniDiameter-2 {
		t.Fatalf("FAB y=%v h=%v expected bottom near %v", b.Y, b.Height, wantB)
	}
}

func TestAvatarLayoutSquare(t *testing.T) {
	a := NewAvatar("av", "", "AB", 0, 0, 48, 36)
	a.Layout()
	b := a.Bounds()
	if b.Width != b.Height {
		t.Fatalf("expected square bounds, got %vx%v", b.Width, b.Height)
	}
}
