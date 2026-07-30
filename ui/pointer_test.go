package ui

import (
	"testing"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestPointerClickSuppressTileBodyLeavesSwitch(t *testing.T) {
	ptrClickPending = true
	ptrClickPos = rl.NewVector2(10, 28)
	ptrClickUsed = false

	tile := rl.NewRectangle(0, 0, 100, 56)
	sw := rl.NewRectangle(40, 40, 20, 20)
	PointerClickSuppressTileBody(tile, sw)
	if PointerClickPending() || PointerClickHandled() {
		t.Fatal("expected body click on tile outside switch to be suppressed")
	}

	ptrClickPending = true
	ptrClickPos = rl.NewVector2(50, 50)
	ptrClickUsed = false
	PointerClickSuppressTileBody(tile, sw)
	if !PointerClickPending() {
		t.Fatal("click on switch area must not be suppressed before toggle consumes")
	}
	if !PointerClickConsume(sw) {
		t.Fatal("toggle should still consume click on switch hit")
	}
}

func TestPointerClickPendingAcrossFrames(t *testing.T) {
	ptrClickPending = true
	ptrClickPos = rl.NewVector2(50, 50)
	ptrClickUsed = false
	ptrLeftWasDown = true

	miss := rl.NewRectangle(0, 0, 10, 10)
	if PointerClickConsume(miss) {
		t.Fatal("miss should not consume")
	}
	if !ptrClickPending || ptrClickUsed {
		t.Fatal("pending gesture should stay until consumed")
	}

	hit := rl.NewRectangle(40, 40, 20, 20)
	if !PointerClickConsume(hit) {
		t.Fatal("second frame should deliver pending click")
	}
}

func TestPointerClickPendingSurvivesButtonRelease(t *testing.T) {
	ptrClickPending = true
	ptrClickPos = rl.NewVector2(50, 50)
	ptrClickUsed = false
	ptrLeftWasDown = true
	ptrClickPendingSince = time.Now()

	ptrLeftWasDown = false
	if !PointerClickPending() {
		t.Fatal("pending click must survive button release until consumed")
	}

	hit := rl.NewRectangle(40, 40, 20, 20)
	if !PointerClickConsume(hit) {
		t.Fatal("expected consume after release")
	}
}

func TestPointerClickConsumeSingleWinner(t *testing.T) {
	ptrClickPending = true
	ptrClickPos = rl.NewVector2(50, 50)
	ptrClickUsed = false

	hit := rl.NewRectangle(40, 40, 20, 20)
	if !PointerClickConsume(hit) {
		t.Fatal("expected first consume")
	}
	if PointerClickConsume(hit) {
		t.Fatal("expected only one consume per gesture")
	}
}
