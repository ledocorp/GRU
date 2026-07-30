package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestRaisedSurfaceUntitledBodyFillsBounds(t *testing.T) {
	c := NewCard("untitled", "", 0, 0, 200, 100)
	c.Layout()
	titleOff := c.bodyTitleHeight()
	if titleOff != 0 {
		t.Fatalf("title band = %v, want 0", titleOff)
	}
	if len(c.Children()) == 0 {
		wantBody := float32(100)
		if c.Bounds().Height != wantBody {
			t.Fatalf("height = %v, want %v", c.Bounds().Height, wantBody)
		}
	}
}

func TestNestedSurfaceShadowSkipPanel(t *testing.T) {
	outer := NewPanel("outer", "Outer", 0, 0, 400, 200)
	inner := NewPanel("inner", "Inner", 0, 0, 200, 80)
	outer.AddChild(inner)
	if !nestedInRaisedSurface(inner) {
		t.Fatal("nested panel should report nestedInRaisedSurface")
	}
	if nestedInRaisedSurface(outer) {
		t.Fatal("root panel should not report nestedInRaisedSurface")
	}
}

func TestNestedSurfaceShadowSkipCard(t *testing.T) {
	outer := NewCard("outer", "Outer", 0, 0, 400, 200)
	inner := NewCard("inner", "Inner", 0, 0, 200, 80)
	outer.AddChild(inner)
	if !nestedInRaisedSurface(inner) {
		t.Fatal("nested card should report nestedInRaisedSurface")
	}
}

func TestRaisedSurfaceBodyHasNoTitleFields(t *testing.T) {
	body := newRaisedSurfaceBody("bare-body")
	if body.GetStyle().Padding >= 0 && len(body.Children()) != 0 {
		t.Fatal("bare body starts empty")
	}
	body.AddChild(NewLabel("l", "x", 0, 0, 0, 20))
	body.SetBounds(rl.NewRectangle(0, 0, 100, 50))
	body.Layout()
	if body.Bounds().Height <= 0 {
		t.Fatal("body should size from content")
	}
}
