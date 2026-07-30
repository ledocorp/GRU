package ui

import (
	"testing"
	"time"
)

func TestFrameBudgetRemainingDeepIdle(t *testing.T) {
	start := time.Now()
	now := start.Add(5 * time.Millisecond)
	if d := frameBudgetRemaining(start, DeepIdleFPS, now); d < 95*time.Millisecond || d > 105*time.Millisecond {
		t.Fatalf("DeepIdle budget remaining = %v, want ~100ms", d)
	}
}

func TestFrameBudgetRemainingActiveSkips(t *testing.T) {
	start := time.Now()
	if d := frameBudgetRemaining(start, ActiveFPS, start.Add(2*time.Millisecond)); d != 0 {
		t.Fatalf("ActiveFPS should not pace, got %v", d)
	}
}

func TestFrameBudgetRemainingElapsedSkips(t *testing.T) {
	start := time.Now()
	now := start.Add(120 * time.Millisecond)
	if d := frameBudgetRemaining(start, DeepIdleFPS, now); d != 0 {
		t.Fatalf("elapsed budget should be 0, got %v", d)
	}
}
