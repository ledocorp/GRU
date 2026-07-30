// Caret debugger — press F5 in the demo to toggle stderr tracing for:
//   - space / char input (focus, cursor, buffer)
//   - caret draw color and lifecycle phase
//   - idle FPS blockers (needsRedraw, dirty nodes, wake)
//
// F6 benchmark · F7 resize FPS · F10 draw-dirty · F11 perf overlay
package main

import (
	"github.com/ledocorp/gru/ui"
	"fmt"
	"time"
)

var (
	caretDebugLastLog time.Time
	caretDebugPeriod  = 400 * time.Millisecond
)

func toggleCaretDebug() {
	ui.CaretDebugEnabled = !ui.CaretDebugEnabled
	caretDebugLastLog = time.Time{}
	fmt.Printf("Gru caret debug: %v (F5 toggle — space/color/caret phase/idle blockers)\n", ui.CaretDebugEnabled)
}

func observeCaretDebugFrame(root ui.Node, in resizeFPSFrameInput) {
	if !ui.CaretDebugEnabled || root == nil {
		return
	}
	now := in.Now
	if now.Sub(caretDebugLastLog) < caretDebugPeriod {
		return
	}
	caretDebugLastLog = now

	reports := ui.CollectDirtyReports(root, 8)
	ui.LogFocusedEditorCaret(root)
	ui.CaretDebugLine(
		"idle scene=%q target=%d state=%s clean=%t cacheHit=%t redraw=%t needsRedraw=%t rootDirty=%t wake=%s dirty=%s",
		in.Scene, in.TargetFPS, in.IdleState, in.CleanForIdle, in.CacheHit, in.FullRedraw,
		in.NeedsRedraw, in.RootDirty, in.EndWake.Reasons, ui.FormatDirtyReports(reports),
	)
}
