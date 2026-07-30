// Resize FPS debugger — press F7 in the demo to toggle detailed stderr logging.
//
// Logs every target-FPS transition with a full snapshot of resize hold, idle
// policy inputs, and frame timing. While enabled, also emits throttled lines
// during active resize (~4/sec) so you can correlate actual FPS with hold state.
//
// F6 = benchmark summary (1 line/sec). F7 = this trace. F8 = layout propagation.
package main

import (
	"github.com/ledocorp/gru/ui"
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	resizeFPSDebug     bool
	resizeFPSLastBurst time.Time
)

func toggleResizeFPSDebug() {
	resizeFPSDebug = !resizeFPSDebug
	resizeFPSLastBurst = time.Time{}
	setIdleGuardFromResizeFPSDebug(resizeFPSDebug)
	fmt.Printf("Gru resize FPS debug: %v (F7 toggle — logs target changes + idle guard)\n", resizeFPSDebug)
}

func resizeFPSDebugOn() bool { return resizeFPSDebug }

// resizeFPSFrameInput is the per-frame snapshot passed from the main loop.
type resizeFPSFrameInput struct {
	Now              time.Time
	Scene            string
	WindowW          int32
	WindowH          int32
	FrameResized     bool
	IsResizing       bool
	ResizeHoldActive bool
	Hold             ui.ResizeHoldSnapshot
	EndWake          ui.WakeSummary
	CleanForIdle     bool
	CacheHit         bool
	RootDirty        bool
	NeedsRedraw      bool
	QueueDrained     int
	FullRedraw       bool
	UpdateMs         float32
	LayoutMs         float32
	DrawMs           float32
	TotalMs          float32
	TargetFPS        int
	PrevTargetFPS    int
	IdleState        string
	PolicyChanged    bool
}

func observeResizeFPSFrame(in resizeFPSFrameInput) {
	if !debugVerbose() {
		return
	}
	if !resizeFPSDebug && !in.PolicyChanged {
		return
	}

	resizeBurst := in.FrameResized || in.IsResizing || in.ResizeHoldActive
	throttledBurst := resizeFPSDebug && resizeBurst &&
		(in.Now.Sub(resizeFPSLastBurst) >= 250*time.Millisecond || in.PolicyChanged)
	if in.PolicyChanged {
		logResizeFPSTransition(in)
	} else if throttledBurst {
		resizeFPSLastBurst = in.Now
		logResizeFPSBurst(in)
	}
}

func logResizeFPSTransition(in resizeFPSFrameInput) {
	dropReason := explainFPSDrop(in)
	fmt.Printf(
		"Gru resize-fps TRANSITION %d→%d state=%s scene=%q win=%dx%d\n"+
			"  hold: active=%t keep=%t cooldown=%t msLeft=%d msSinceAct=%d msSinceDim=%d burst=%d\n"+
			"  gesture: resized=%t resizing=%t resizeHoldActive=%t\n"+
			"  idle: clean=%t cacheHit=%t rootDirty=%t needsRedraw=%t queue=%d wake=%s\n"+
			"  frame: redraw=%t update=%.1fms layout=%.1fms draw=%.1fms total=%.1fms\n"+
			"  %s\n",
		in.PrevTargetFPS, in.TargetFPS, in.IdleState, in.Scene, in.WindowW, in.WindowH,
		in.Hold.HoldActive, in.Hold.KeepActive, in.Hold.InPostCooldown,
		in.Hold.MsUntilHoldEnd, in.Hold.MsSinceLastActivity, in.Hold.MsSinceLastDim, in.Hold.BurstSteps,
		in.FrameResized, in.IsResizing, in.ResizeHoldActive,
		in.CleanForIdle, in.CacheHit, in.RootDirty, in.NeedsRedraw, in.QueueDrained, in.EndWake.Reasons,
		in.FullRedraw, in.UpdateMs, in.LayoutMs, in.DrawMs, in.TotalMs,
		dropReason,
	)
}

func logResizeFPSBurst(in resizeFPSFrameInput) {
	fmt.Printf(
		"Gru resize-fps burst target=%d state=%s actual~=%.0ffps win=%dx%d resized=%t resizing=%t hold=%t keep=%t msLeft=%d redraw=%t update=%.0fms layout=%.0fms draw=%.0fms wake=%s\n",
		in.TargetFPS, in.IdleState, 1000.0/max(in.TotalMs, 0.1), in.WindowW, in.WindowH,
		in.FrameResized, in.IsResizing, in.Hold.HoldActive, in.Hold.KeepActive, in.Hold.MsUntilHoldEnd,
		in.FullRedraw, in.UpdateMs, in.LayoutMs, in.DrawMs, in.EndWake.Reasons,
	)
}

func explainFPSDrop(in resizeFPSFrameInput) string {
	if in.TargetFPS >= ui.ActiveFPS {
		return "reason: bumped to active (wake/blockers/dirty/redraw or resize-hold)"
	}
	if in.ResizeHoldActive {
		return "reason: UNEXPECTED drop while resizeHoldActive=true — file a bug"
	}
	parts := make([]string, 0, 6)
	if in.Hold.KeepActive {
		parts = append(parts, "keepActive still true but resizeHoldActive false — check main loop wiring")
	}
	if !in.Hold.HoldActive && in.Hold.MsSinceLastActivity >= int(ui.ResizeIdleCooldown/time.Millisecond) {
		parts = append(parts, fmt.Sprintf("post-cooldown expired (%dms since activity)", in.Hold.MsSinceLastActivity))
	} else if !in.Hold.HoldActive {
		parts = append(parts, fmt.Sprintf("hold window expired (%dms since activity, cooldown=%dms)",
			in.Hold.MsSinceLastActivity, int(ui.ResizeIdleCooldown/time.Millisecond)))
	}
	if in.CleanForIdle {
		parts = append(parts, "cleanForIdle=true")
	} else {
		if !in.CacheHit {
			parts = append(parts, "not clean: cache miss")
		}
		if in.RootDirty {
			parts = append(parts, "not clean: root dirty")
		}
		if in.NeedsRedraw {
			parts = append(parts, "not clean: needs redraw")
		}
		if in.QueueDrained > 0 {
			parts = append(parts, "not clean: queue drained")
		}
	}
	if in.EndWake.Reasons == 0 {
		parts = append(parts, "endWake=none")
	}
	if !in.IsResizing && !in.FrameResized {
		parts = append(parts, "no active gesture")
	}
	if rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsMouseButtonDown(rl.MouseRightButton) {
		parts = append(parts, "mouse still down (input wake should have fired)")
	}
	if len(parts) == 0 {
		return "reason: grace window elapsed with clean cache"
	}
	return "reason: " + joinReasons(parts)
}

func joinReasons(parts []string) string {
	out := parts[0]
	for i := 1; i < len(parts); i++ {
		out += "; " + parts[i]
	}
	return out
}
