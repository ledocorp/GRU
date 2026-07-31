// Demo debug helpers for the grudemo launcher (F5–F11 / env toggles).
//
// Consolidated at repo root so public export stays a short allowlist — not a
// scatter of debug_*.go files. Keys:
//
//	F5  caret / idle blockers          F6  benchmark (main)
//	F7  resize FPS + idle guard        F8  resize propagation
//	F10 draw-dirty trace               F11 perf overlay (main)
//	Shift+F11 WebView flash
//
// Env: GRU_DEBUG, GRU_IDLE_GUARD, GRU_WEBVIEW_DEBUG (GORY_* aliases).
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/ui"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func debugVerbose() bool {
	v := envAlias("GRU_DEBUG", "GORY_DEBUG")
	if appReleaseMode() {
		return v == "1"
	}
	return v != "0"
}

// ── Caret (F5) ───────────────────────────────────────────────────────────────

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

// ── Draw-dirty (F10) ─────────────────────────────────────────────────────────

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

// ── Idle guard (env / F7) ────────────────────────────────────────────────────

const (
	idleGuardStreakFrames = 120
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

// ── Resize FPS (F7) ──────────────────────────────────────────────────────────

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

// ── Resize propagation (F8) ──────────────────────────────────────────────────

func logBounds(tag string, n ui.Node) {
	if n == nil {
		log.Printf("[resize-debug] %s: <nil>", tag)
		return
	}
	r := n.Bounds()
	log.Printf("[resize-debug] %s id=%s X=%.1f Y=%.1f W=%.1f H=%.1f", tag, n.ID(), r.X, r.Y, r.Width, r.Height)
}

// TestResizePropagation builds an isolated Document and logs shell→viewport→grid
// bounds across a few Resize calls (F8). Does not modify the running demo doc.
func TestResizePropagation() {
	const w0, h0 int32 = 1280, 720
	doc := ui.NewDocument(w0, h0)
	shell, vp := examples.MountFlexPageShell(doc, "rzdbg")
	vp.SetStyle("transparent")

	grid := ui.NewContainer("rzdbg-grid", 0, 0, 0, 0)
	grid.LayoutType = ui.LayoutGrid
	grid.GridColumns = 12
	grid.Gap = 12
	grid.SetStyle("page-grid")

	sample := ui.NewPanel("rzdbg-panel", "Sample", 0, 0, 0, 120)
	sample.SetColSpan(ui.BreakpointXS, 12)
	sample.SetColSpan(ui.BreakpointMD, 6)
	sample.SetColSpan(ui.BreakpointLG, 4)

	grid.AddChild(sample)
	vp.AddChild(grid)

	doc.Root.MarkDirty()
	doc.Root.Layout()

	log.Println("[resize-debug] === BEFORE (initial layout, 1280×720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)

	doc.Resize(640, h0)
	log.Println("[resize-debug] === AFTER doc.Resize(640, 720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)

	doc.Resize(1600, h0)
	log.Println("[resize-debug] === AFTER doc.Resize(1600, 720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)

	doc.Resize(w0, h0)
	log.Println("[resize-debug] === AFTER doc.Resize(restore 1280×720) ===")
	logBounds("shell", shell)
	logBounds("main viewport", vp)
	logBounds("grid", grid)
	logBounds("sample panel", sample)
	log.Println("[resize-debug] === done (isolated doc discarded) ===")
}

// ── WebView flash (Shift+F11) ────────────────────────────────────────────────

func toggleWebViewDebug() {
	ui.WebViewDebugEnabled = !ui.WebViewDebugEnabled
	fmt.Printf("Gru webview debug: %v (Shift+F11 toggle — passthrough/visible trace)\n", ui.WebViewDebugEnabled)
}
