// Draw-dirty debugger — press F10 to toggle per-frame dirty-node tracing.
//
// When enabled, logs the first dirty nodes whenever a full SSAA redraw runs
// without an interactive wake (mouse/keyboard/resize). Use with F6 benchmark
// lines to correlate redraw=60 with the widgets still marking the tree dirty.
package main

import (
	"github.com/ledocorp/gru/ui"
	"fmt"
	"time"
)

var (
	drawDirtyDebug     bool
	drawDirtyLastLog   time.Time
	drawDirtyLogPeriod = 500 * time.Millisecond
)

func toggleDrawDirtyDebug() {
	drawDirtyDebug = !drawDirtyDebug
	drawDirtyLastLog = time.Time{}
	fmt.Printf("Gru draw-dirty debug: %v (F10 toggle — logs dirty nodes on unexplained redraws)\n", drawDirtyDebug)
}

func drawDirtyDebugOn() bool { return drawDirtyDebug }

func observeDrawDirtyFrame(root ui.Node, fullRedraw bool, endWake ui.WakeSummary, needsRedraw, rootDirty bool) {
	if !drawDirtyDebug || root == nil {
		return
	}
	if !fullRedraw && !needsRedraw {
		return
	}
	now := time.Now()
	if now.Sub(drawDirtyLastLog) < drawDirtyLogPeriod {
		return
	}
	drawDirtyLastLog = now
	reports := ui.CollectDirtyReports(root, 12)
	fmt.Printf("Gru draw-dirty trace: redraw=%t needsRedraw=%t rootDirty=%t wake=%s nodes=%s\n",
		fullRedraw, needsRedraw, rootDirty, endWake.Reasons, ui.FormatDirtyReports(reports))
}
