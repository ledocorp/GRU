package ui

import "time"

// PaceFrame sleeps until the wall-clock budget for targetFPS is met.
//
// Raylib SetTargetFPS is unreliable below the display refresh when FlagVsyncHint
// is enabled — the loop can still spin at ~60 Hz while the idle policy reports
// DeepIdleFPS. Call once at the end of each main-loop iteration after EndDrawing.
func PaceFrame(frameStart time.Time, targetFPS int) {
	if d := frameBudgetRemaining(frameStart, targetFPS, time.Now()); d > 0 {
		time.Sleep(d)
	}
}

// frameBudgetRemaining returns how long the caller should still wait to honor
// targetFPS. Exported for tests only via the helper below.
func frameBudgetRemaining(frameStart time.Time, targetFPS int, now time.Time) time.Duration {
	if targetFPS <= 0 || targetFPS >= ActiveFPS {
		return 0
	}
	budget := time.Second / time.Duration(targetFPS)
	elapsed := now.Sub(frameStart)
	if elapsed >= budget {
		return 0
	}
	return budget - elapsed
}
