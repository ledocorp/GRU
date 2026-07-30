package ui

import (
	"testing"
	"time"
)

func TestResizeHoldTrackerBurstExtendsTail(t *testing.T) {
	var hold ResizeHoldTracker
	t0 := time.Now()
	hold.NoteDimensionChange(t0)
	if !hold.Active(t0.Add(700 * time.Millisecond)) {
		t.Fatal("expected hold active 700ms after first step")
	}
	hold.NoteDimensionChange(t0.Add(900 * time.Millisecond))
	if !hold.Active(t0.Add(2500 * time.Millisecond)) {
		t.Fatal("burst tail should extend past single-step hold")
	}
}

func TestResizeHoldTrackerIdleCooldown(t *testing.T) {
	var hold ResizeHoldTracker
	t0 := time.Now()
	hold.NoteDimensionChange(t0)
	afterHold := t0.Add(ResizeHoldDuration + 100*time.Millisecond)
	if !hold.KeepActiveFPS(afterHold) {
		t.Fatal("cooldown should keep FPS active briefly after hold expires")
	}
	afterCooldown := t0.Add(ResizeHoldDuration + ResizeIdleCooldown + 100*time.Millisecond)
	if hold.KeepActiveFPS(afterCooldown) {
		t.Fatal("cooldown should end after ResizeIdleCooldown")
	}
}
