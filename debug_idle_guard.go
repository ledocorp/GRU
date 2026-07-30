// Idle guard — warns when hands-off frames cannot reach cleanForIdle (NeedsRedraw pinned).
//
// Enable with GRU_IDLE_GUARD=1 (GORY_IDLE_GUARD alias) or toggle with F7 (resize FPS debug).
// See docs/IDLE_INVARIANTS.md.
package main

import (
	"github.com/ledocorp/gru/ui"
	"fmt"
	"os"
	"time"
)

const (
	idleGuardStreakFrames = 120 // ~2 s at 60 FPS before first warning
	idleGuardWarnCooldown = 5 * time.Second
)

var idleGuard struct {
	envEnabled   bool
	manualEnable bool
	streak       int
	lastWarn     time.Time
}

func init() {
	idleGuard.envEnabled = envAliasEq("GRU_IDLE_GUARD", "GORY_IDLE_GUARD", "1")
}

func idleGuardActive() bool {
	return idleGuard.envEnabled || idleGuard.manualEnable || resizeFPSDebug
}

func setIdleGuardFromResizeFPSDebug(on bool) {
	idleGuard.manualEnable = on
	if !on {
		idleGuard.streak = 0
	}
}

type idleGuardInput struct {
	CleanForIdle bool
	TargetFPS    int
	EndWake      ui.WakeSummary
	Root         ui.Node
}

func observeIdleGuard(in idleGuardInput) {
	if !idleGuardActive() || in.Root == nil {
		return
	}
	handsOff := in.EndWake.Reasons&ui.WakeInteractive == 0 && !ui.PointerInputActive()
	if handsOff && !in.CleanForIdle {
		idleGuard.streak++
		if idleGuard.streak < idleGuardStreakFrames {
			return
		}
		now := time.Now()
		if now.Sub(idleGuard.lastWarn) < idleGuardWarnCooldown {
			return
		}
		idleGuard.lastWarn = now
		fmt.Fprintf(os.Stderr,
			"Gru IDLE GUARD: hands-off but not idle-ready for %d+ frames (target=%d fps). Blockers: %s\n"+
				"  Hint: F10 draw-dirty trace · docs/IDLE_INVARIANTS.md · go test ./ui -run Idle\n",
			idleGuard.streak, in.TargetFPS, ui.NotIdleReason(in.Root))
		return
	}
	idleGuard.streak = 0
}
