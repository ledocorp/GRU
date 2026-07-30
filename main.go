// Package main is the Gru demo launcher.
//
// Press Tab (or the on-screen button) to cycle through the available demo scenes.
// Each scene is built by a function in the examples/ package.
//
//go:generate go run ./scripts/build/gen_app_icon.go
//go:generate go run github.com/josephspurrier/goversioninfo/cmd/goversioninfo@v1.4.1 -64=true -icon=packaging/icons/gru-notepad.ico -o=resource.syso packaging/versioninfo.json
package main

import (
	"github.com/ledocorp/gru/devtools/studio"
	"github.com/ledocorp/gru/examples"
	"github.com/ledocorp/gru/examples/appinstance"
	"github.com/ledocorp/gru/internal/appicon"
	"github.com/ledocorp/gru/internal/tray"
	"github.com/ledocorp/gru/ui"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func init() {
	if envAliasSet("GRU_CARET_DEBUG", "GORY_CARET_DEBUG") {
		ui.CaretDebugEnabled = true
	}
	if envAliasSet("GRU_WEBVIEW_DEBUG", "GORY_WEBVIEW_DEBUG") {
		ui.WebViewDebugEnabled = true
	}
	ui.InitWebViewHostPolicy(appReleaseMode())
}

const (
	initWindowW = 1280 // Initial OS window size (can be resized by user)
	initWindowH = 720  // 16:9 — keep in sync with supersample/doc logical size at startup
)

// windowW / windowH track the current logical dimensions and are updated
// every frame from rl.GetScreenWidth() / rl.GetScreenHeight() so that resize
// (both OS-native and custom edge-drag) is reflected in layout and rendering.
var windowW, windowH int32 = initWindowW, initWindowH

// readClientSizeClamped returns the current raylib client size, enforcing
// ui.MinClientWidth so the layout canvas never runs narrower than policy.
func readClientSizeClamped() (w, h int32) {
	if isAndroidApp() {
		w = int32(rl.GetScreenWidth())
		h = int32(rl.GetScreenHeight())
		w, h = appNormalizeAndroidClientSize(w, h)
		if w < 1 {
			w = 360
		}
		if h < 1 {
			h = 640
		}
		return w, h
	}
	const maxAttempts = 6
	for attempt := 0; attempt < maxAttempts; attempt++ {
		w = int32(rl.GetScreenWidth())
		h = int32(rl.GetScreenHeight())
		if w < 1 || h < 1 {
			return w, h
		}
		if w >= ui.MinClientWidth {
			return w, h
		}
		rl.SetWindowSize(int(ui.MinClientWidth), int(h))
	}
	w = int32(rl.GetScreenWidth())
	h = int32(rl.GetScreenHeight())
	if w < ui.MinClientWidth {
		w = ui.MinClientWidth
	}
	return w, h
}

const (
	navBarH      = 36
	navDirBtnW   = float32(108)
	navDirBtnPad = float32(8)
)

// demoNavBottom returns chrome reserved for the demo launcher footer, or 0 when hidden.
func demoNavBottom(scene examples.Scene) int32 {
	if scene != nil && scene.HideDemoNav() {
		return 0
	}
	return navBarH
}

// perfOverlayVisible is true in dev when the user toggled F11 (nav-bar stats strip).
func perfOverlayVisible() bool {
	return !appReleaseMode() && ui.ShowPerfOverlay
}

// engineDebugHUDVisible gates the green raylib DrawFPS corner counter.
// Off by default in Studio and grudemo — use F12 Inspector for FPS.
// Opt in with GRU_RAYLIB_FPS=1 (GORY_RAYLIB_FPS alias) when debugging the raw Raylib counter.
func engineDebugHUDVisible() bool {
	return envAliasEq("GRU_RAYLIB_FPS", "GORY_RAYLIB_FPS", "1")
}

func navBarRect(windowW, windowH int32) rl.Rectangle {
	return rl.NewRectangle(0, float32(windowH-navBarH), float32(windowW), float32(navBarH))
}

func navDirectoryButtonRect(windowW, windowH int32) rl.Rectangle {
	x := float32(windowW) - navDirBtnPad - navDirBtnW
	return rl.NewRectangle(x, float32(windowH-navBarH)+6, navDirBtnW, float32(navBarH)-12)
}

func slimNavHint(current, total int, title string) string {
	if examples.PublicDemoMode() {
		return fmt.Sprintf("Tab · %d/%d · %s  ·  F12  ·  Scenes", current+1, total, title)
	}
	return fmt.Sprintf("%d/%d · %s", current+1, total, title)
}

func studioToolState(current, total int, title string, benchmarkMode, borderless bool, inspector *ui.Inspector) studio.ToolState {
	return studio.ToolState{
		SceneIndex:     current,
		SceneCount:     total,
		SceneTitle:     title,
		BenchmarkOn:    benchmarkMode,
		PerfOverlayOn:  ui.ShowPerfOverlay,
		ResizeFPSOn:    resizeFPSDebugOn(),
		CaretDebugOn:   ui.CaretDebugEnabled,
		DrawDirtyOn:    drawDirtyDebugOn(),
		WebViewDebugOn: ui.WebViewDebugEnabled,
		BorderlessOn:   borderless,
		InspectorOn:    inspector != nil && inspector.IsVisible(),
	}
}

func drawNavBar(windowW, windowH int32, navBg rl.Color) {
	navRect := navBarRect(windowW, windowH)
	rl.DrawRectangleRec(navRect, navBg)
	btn := navDirectoryButtonRect(windowW, windowH)
	rl.DrawRectangleRounded(btn, 0.25, 6, rl.NewColor(79, 70, 229, 255))
}

func drawNavBarLabels(windowW, windowH int32, hint string, textColor rl.Color) {
	hintStyle := ui.ChromeFooterHintStyle()
	hintStyle.TextColor = textColor
	hintX := float32(92)
	rightReserve := navDirBtnW + navDirBtnPad
	hintRect := rl.NewRectangle(hintX, float32(windowH-navBarH), float32(windowW)-hintX-rightReserve, float32(navBarH))
	hintY := ui.ChromeTextCenterY(hintRect, hintStyle)
	ui.DrawChromeText(hint, hintX, hintY, hintStyle)

	btn := navDirectoryButtonRect(windowW, windowH)
	dirStyle := ui.ChromeFooterButtonStyle()
	dirStyle.TextColor = rl.White
	label := "Directory"
	if examples.PublicDemoMode() {
		label = "Scenes"
	}
	tw := ui.MeasureChromeText(label, dirStyle)
	dirX := btn.X + (btn.Width-tw)/2
	dirY := ui.ChromeTextCenterY(btn, dirStyle)
	ui.DrawChromeText(label, dirX, dirY, dirStyle)
}

// drawFooterOverlay paints the full launcher footer (backgrounds + labels) in the
// SSAA overlay pass so text is not composited over a bilinear-downscaled bar.
func drawFooterOverlay(windowW, windowH int32, navBg rl.Color, navHint string, navTextColor rl.Color, panel *studio.Panel, state studio.ToolState) {
	drawNavBar(windowW, windowH, navBg)
	drawNavBarLabels(windowW, windowH, navHint, navTextColor)
	if panel != nil && !examples.PublicDemoMode() {
		panel.PaintFooter(windowW, windowH, state)
	}
}

func drawDevChromeOverlays(windowW, windowH int32, bottomChrome int32, panel *studio.Panel, state studio.ToolState, inspector *ui.Inspector, navBg rl.Color, navHint string, navTextColor rl.Color) {
	if appReleaseMode() {
		return
	}
	if bottomChrome > 0 {
		drawFooterOverlay(windowW, windowH, navBg, navHint, navTextColor, panel, state)
	}
	if inspector != nil {
		inspector.Draw()
	}
}

// drawPostChromeOverlays composites interaction chrome, launcher footer, and inspector
// in one SSAA overlay pass when available (avoids pixelated 1× footer text).
func drawPostChromeOverlays(windowW, windowH int32, root ui.Node, bottomChrome int32, panel *studio.Panel, state studio.ToolState, inspector *ui.Inspector, navBg rl.Color, navHint string, navTextColor rl.Color) {
	if ui.SupersamplingActive() && ui.OverlayTargetDrawable() {
		ui.BeginOverlaySuperFrame()
		if root != nil {
			ui.DrawInteractionOverlays(root)
		}
		drawDevChromeOverlays(windowW, windowH, bottomChrome, panel, state, inspector, navBg, navHint, navTextColor)
		ui.EndOverlaySuperFrame()
		ui.BlitOverlayToScreen(windowW, windowH)
		return
	}
	if root != nil {
		ui.DrawInteractionOverlays(root)
	}
	drawDevChromeOverlays(windowW, windowH, bottomChrome, panel, state, inspector, navBg, navHint, navTextColor)
}

func drawLauncherChrome(windowW, windowH int32, navBg rl.Color, panel *studio.Panel, state studio.ToolState) {
	_ = windowW
	_ = windowH
	_ = navBg
	_ = panel
	_ = state
}

// syncWindowFromDisplay updates windowW/H, SSAA targets, document, breakpoints,
// and title-bar cached size when the OS client area changed (resize, maximize,
// restore, borderless edge drag). Call after any code that may change window
// dimensions in-frame (e.g. titleBar double-click maximize).
func syncWindowFromDisplay(doc *ui.Document, titleBar *ui.TitleBar, borderless bool) {
	newW, newH := readClientSizeClamped()
	if newW < 1 || newH < 1 {
		return
	}
	if newW == windowW && newH == windowH {
		return
	}
	windowW, windowH = newW, newH
	ui.ResizeWindowTextures(windowW, windowH)
	if ui.ApplyDisplayAwareSupersampling(windowW, windowH) && doc != nil {
		doc.UnloadCache()
	}
	// Update RootFontSize from window width (not height) before doc.Resize,
	// so layout/measure uses the correct EffectiveFontSize immediately.
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	if doc != nil {
		doc.Resize(windowW, windowH)
	}
	if bp := ui.CurrentBreakpoint(float32(windowW)); ui.ActiveBreakpoint.Get() != bp {
		ui.ActiveBreakpoint.Set(bp)
	}
	titleBar.SetSize(windowW, windowH)
	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.NotificationCenterMgr.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	if borderless {
		applyNativeBorderlessRoundedCorners(titleBar.BorderlessRoundedChrome())
		ui.SetBorderlessRoundedChrome(titleBar.BorderlessRoundedChrome())
	}
}

func toggleBorderlessChrome(
	borderless *bool,
	doc *ui.Document,
	titleBar *ui.TitleBar,
	activeScene examples.Scene,
	windowW, windowH int32,
) {
	if !appUsesCustomChrome() {
		return
	}
	if !*borderless {
		*borderless = true
		rl.SetWindowState(uint32(rl.FlagWindowUndecorated))
		examples.ConfigureTitleBar(titleBar, activeScene)
		titleBar.HandleResize = true
		titleBar.ResetPos()
		doc.SetChromeTop(ui.TitleBarHeight)
		doc.SetChromeBottom(float32(demoNavBottom(activeScene)))
		syncWindowFromDisplay(doc, titleBar, *borderless)
		doc.SyncBorderlessLayout(windowW, windowH)
		doc.UnloadCache()
	} else {
		*borderless = false
		rl.ClearWindowState(uint32(rl.FlagWindowUndecorated))
		titleBar.HandleResize = false
		doc.SetChromeTop(0)
		doc.SetChromeBottom(float32(demoNavBottom(activeScene)))
		syncWindowFromDisplay(doc, titleBar, *borderless)
		doc.Resize(windowW, windowH)
		doc.ForceFullLayout()
		doc.UnloadCache()
	}
	applyNativeBorderlessRoundedCorners(*borderless)
	ui.ArmViewportScrollRecovery(2)
}

func main() {
	if !isAndroidApp() {
		runApp()
	}
}

func runApp() {
	exit, stopInstance := examples.StartupInstance(examples.StartupOpenFilePath())
	if exit {
		return
	}
	defer stopInstance()

	// ── Logging: suppress info/debug noise in normal runs ────────────────────
	// Must be called before SetConfigFlags / InitWindow so that startup
	// messages from raylib are also filtered. Change to LogAll while debugging.
	rl.SetTraceLogLevel(rl.LogWarning)

	// ── Window quality flags ────────────────────────────────────────────────
	// Must be set before InitWindow.
	//   FlagMsaa4xHint     — ask the GPU for 4× MSAA on the default framebuffer
	//   FlagVsyncHint      — enable VSync (smooth tearing-free rendering)
	//   FlagWindowResizable  — allow the window to be resized (OS chrome when decorated,
	//                           custom edge grips when borderless)
	//   FlagWindowUndecorated — use Gru chrome by default for one consistent layout path
	// FlagWindowHighdpi is omitted so framebuffer pixels match layout coordinates
	// (Document + SSAA logical size track GetScreenWidth/Height).
	rl.SetConfigFlags(appInitWindowFlags())

	initW, initH := appInitWindowSize()
	rl.InitWindow(initW, initH, appInitWindowTitle())
	appinstance.RaiseRunningInstance = platformWindowFocus
	if isAndroidApp() {
		windowW = int32(rl.GetScreenWidth())
		windowH = int32(rl.GetScreenHeight())
		windowW, windowH = appNormalizeAndroidClientSize(windowW, windowH)
		if windowW < 1 {
			windowW = 360
		}
		if windowH < 1 {
			windowH = 640
		}
	}
	setupPlatformWindowHooks()
	if appUsesCustomChrome() {
		applyNativeBorderlessRoundedCorners(true)
		platformWindowFocus()
	}
	// After HWND exists and DWM chrome is applied (DWM can reset WM_SETICON if set too early).
	appicon.ApplyWindowIcon()
	// GLFW-enforced minimum client size (desktop only).
	if appUsesCustomChrome() {
		rl.SetWindowMinSize(int(ui.MinClientWidth), 320)
	}
	defer rl.CloseWindow()

	// Disable the default ESC-closes-window behaviour so that widgets that use
	// Escape internally (SearchBar clear, TextInput cancel, etc.) do not quit
	// the app. The window can still be closed via the title-bar × button.
	rl.SetExitKey(0)

	// ── Global render state (UI app — set once, persist for app lifetime) ————
	// See ui.ApplyUIOptimizations for the rationale. Audio device is NOT
	// initialised: Gru is a pure UI engine and starting the audio subsystem
	// (rl.InitAudioDevice) would spin up a mixer thread for no benefit.
	ui.ApplyUIOptimizations()

	rl.SetTargetFPS(60)

	// ── High-resolution font + icon atlases (T1.9 DPR-scaled source) ─────────
	// Base sizes: 176 pt SDF/bitmap text, 512 pt Remix icons — scaled by DisplayScale
	// (125% DPI → ~220 / ~640). SSAA stays separate; do not multiply SSAA × DPI.
	ui.RefreshDisplayScale()
	ui.IconsEagerWarm = true // Studio catalog needs WarmGlyphNames at first paint
	ui.InitDisplayAwareAtlases()
	applyTextEngineMode()
	defer ui.UnloadSDFFonts()

	// ── 2× Supersampled RenderTexture (SSAA) ────────────────────────────────
	if isAndroidApp() {
		// OpenGL ES on Android: skip 2x SSAA (1x direct draw).
		ui.RenderScale = 1
	} else {
		ui.InitSupersampling(windowW, windowH)
		ui.ApplyDisplayAwareSupersampling(windowW, windowH)
		defer ui.UnloadSupersampling()
	}
	// HWND + GPU ready — re-apply icons after heavy init (matches prior “delayed but working” behavior).
	appicon.ReapplyWindowIcon()
	if debugVerbose() {
		fmt.Printf("Gru render: text=%s  icons=%s  ssaa=%.1fx  dpi=%.2f  sdfAtlas=%d  remixAtlas=%d\n",
			ui.TextEngineBackendName(), ui.Phosphor.IconFontSummary(), ui.EffectiveSupersamplingScale(),
			ui.DisplayScale, ui.EffectiveSDFAtlasSize(), ui.EffectiveRemixAtlasSize())
	} else if appReleaseMode() {
		// One stderr line for ship-gate render verification (§A.4); no idle/resize spam.
		fmt.Fprintln(os.Stderr, notepadReleaseRenderLine())
	}
	// Seed RootFontSize once at startup. Resize events later keep it in sync.
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.NotificationCenterMgr.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	rl.TraceLog(rl.LogInfo, "Gru display DPI %.2f, SSAA %.2fx (FontSize = screen px)",
		ui.DisplayScale, ui.EffectiveSupersamplingScale())
	defer ui.Toasts.Unload()
	defer ui.Icons.UnloadAll()

	// ── Detect and log render capability tier ────────────────────────────────
	ui.CurrentRenderMode = ui.DetectRenderCapability()
	rl.TraceLog(rl.LogInfo, "Gru render mode: %s", ui.RenderCapabilityString())

	// ── Demo registry ─────────────────────────────────────────────────────────
	// Scenes self-register via init() in each demo file; we just retrieve them.
	factories := examples.Registered()
	if examples.PublicDemoMode() {
		factories = examples.FilterPublicFactories(factories)
		if len(factories) == 0 {
			fmt.Fprintln(os.Stderr, "grudemo: no public scenes registered — check examples/public_catalog.go")
			return
		}
		fmt.Fprintf(os.Stderr, "grudemo: %d public scenes (start=%s)\n", len(factories), examples.PublicDemoStartTitle)
	}
	current := 0
	if examples.PublicDemoMode() {
		current = examples.IndexOfPublicStart(factories)
	} else {
		for i := range factories {
			if examples.RegistryTitle(i) == examples.DirectorySceneTitle {
				current = i
				break
			}
		}
	}
	current = examples.ResolveStartupSceneIndex(factories, current)

	// ── Custom title bar ───────────────────────────────────────────────────────
	// Gru starts with its own borderless chrome so layout, input, resize, and
	// titlebar behavior all run through the same polished engine path. Press F9
	// to temporarily return to native Windows chrome while debugging.
	var closeRequested bool
	var trayShowPending atomic.Bool
	var trayQuitPending atomic.Bool
	titleBar := ui.NewTitleBar(
		"Gru",
		ui.TitleBarStyleDark,
		func() { platformWindowClose(&closeRequested) },
		platformWindowMinimize,
		platformWindowToggleMaximize,
	)
	titleBar.HandleResize = true
	titleBar.DrawLeadingIcon = appicon.DrawTitleBarIcon
	titleBar.ResetPos()
	titleBar.SetSize(windowW, windowH)

	// ── Scene lifecycle helpers ───────────────────────────────────────────────
	var (
		doc         *ui.Document
		activeScene examples.Scene
		borderless  = appUsesCustomChrome() // F9 toggles on desktop only
	)
	defer func() {
		ui.DestroyAllWebViewHosts()
		if activeScene != nil {
			activeScene.Destroy()
		}
	}()
	inspector := ui.NewInspector()
	var studioPanel studio.Panel
	idlePolicy := ui.NewRenderIdlePolicy(time.Now())

	loadScene := func(index int) {
		studioPanel.Close()
		ui.DestroyAllWebViewHosts()
		if activeScene != nil {
			activeScene.Destroy()
		}
		ui.ResetTransientOverlays()
		ui.ClearTooltipEntries()
		activeScene = factories[index]()
		doc = ui.NewDocument(windowW, windowH)
		doc.SetChromeBottom(float32(demoNavBottom(activeScene)))
		if borderless {
			doc.SetChromeTop(ui.TitleBarHeight)
		}
		activeScene.Build(doc)
		examples.PreloadScenePhosphor(doc)
		doc.Resize(windowW, windowH)
		if borderless {
			doc.SyncBorderlessLayout(windowW, windowH)
		}
		examples.FinishShellMount(doc)
		if rl.IsWindowReady() {
			doc.SetPlatformWindowHandle(uintptr(rl.GetWindowHandle()))
		}
		ui.SyncWebViewHosts(doc)
		inspector.ResetForSceneChange()
		if examples.PublicDemoMode() {
			rl.SetWindowTitle(fmt.Sprintf("Gru Demo — %s", activeScene.Title()))
		} else {
			rl.SetWindowTitle(fmt.Sprintf("Gru — %s", activeScene.Title()))
		}
		examples.ConfigureTitleBar(titleBar, activeScene)
		// Enable the document-level UI cache: in SSAA mode this causes
		// NeedsRedraw to skip the draw pass on clean frames (the superTarget
		// persists and is blitted instead). In 1× fallback mode it allocates
		// a RenderTexture2D so the same skip-and-blit logic applies.
		doc.EnableUIRenderTexture(true)
		if doc.Root != nil {
			if doc.Root.IsDirty() {
				doc.Root.Layout()
			}
			doc.InvalidatePaint()
		}
		idlePolicy.NoteSceneLoad(time.Now())
	}

	// Build title→index before the first loadScene so grace wall-clock is not
	// burned instantiating every factory after NoteSceneLoad.
	sceneTitleIndex := make(map[string]int, len(factories))
	for i, f := range factories {
		sc := f()
		sceneTitleIndex[sc.Title()] = i
		sc.Destroy()
	}

	loadScene(current)
	if appReleaseMode() {
		ui.ShowPerfOverlay = false
	}

	// ── Launcher UI (always-on-top nav bar) ───────────────────────────────────
	navBg := rl.NewColor(30, 32, 42, 255)
	navTextColor := rl.NewColor(200, 202, 215, 255)
	bgColor := rl.NewColor(228, 229, 235, 255)

	var frameCount int // used for periodic cache-hit-rate logging

	prevMouse := rl.GetMousePosition()
	lastPolicyLog := time.Now()
	lastTargetFPS := ui.ActiveFPS
	rl.SetTargetFPS(int32(ui.ActiveFPS))
	benchmarkMode := false
	benchmarkLastLog := time.Now()
	benchmarkFrames := 0
	benchmarkRedraws := 0
	benchmarkBlits := 0
	processStats := newProcessSampler()
	var resizeHold ui.ResizeHoldTracker
	var resizeHoldWasActive bool
	var pendingNavWake ui.WakeSummary
	var lastWindowMaximized bool
	windowShown := false // desktop starts FlagWindowHidden; reveal after first present
	taskbarIconAfterFrame := false
	var lastSceneSwitch time.Time
	const sceneSwitchDebounce = 180 * time.Millisecond

	switchScene := func(next int) {
		now := time.Now()
		if !lastSceneSwitch.IsZero() && now.Sub(lastSceneSwitch) < sceneSwitchDebounce {
			return
		}
		lastSceneSwitch = now
		if doc != nil {
			doc.UnloadCache()
		}
		current = next
		pendingNavWake.Add(ui.WakeScene|ui.WakeInput, "navigate")
		idlePolicy.NoteInteractiveWake(now)
		loadScene(current)
		rl.SetTargetFPS(int32(ui.ActiveFPS))
		lastTargetFPS = ui.ActiveFPS
	}

	examples.NavigateToScene = func(title string) bool {
		idx, ok := sceneTitleIndex[title]
		if !ok {
			return false
		}
		if idx == current {
			return true
		}
		switchScene(idx)
		return true
	}
	ui.ActiveBreakpoint.Set(ui.CurrentBreakpoint(float32(windowW)))

	// PresentHook redraws the cached frame + overlays without a full update pass.
	// File dialogs call this so dropdown menus vanish before blocking zenity.
	ui.PresentHook = func() {
		if doc == nil || !rl.IsWindowReady() {
			return
		}
		bottomChrome := demoNavBottom(activeScene)
		var navRect rl.Rectangle
		if bottomChrome > 0 {
			navRect = navBarRect(windowW, windowH)
		}
		navHint := slimNavHint(current, len(factories), activeScene.Title())
		presentStudio := studioToolState(current, len(factories), activeScene.Title(), benchmarkMode, borderless, inspector)

		if ui.SupersamplingActive() {
			rl.BeginDrawing()
			ui.BeginDrawingBorderless(bgColor, borderless, windowW, windowH)
			if ui.SuperTargetDrawable() {
				ui.BlitToScreenBorderless(windowW, windowH, borderless)
			}
			ui.DrawAnimationOverlays(doc.Root, navRect)
			drawPostChromeOverlays(windowW, windowH, doc.Root, bottomChrome, &studioPanel, presentStudio, inspector, navBg, navHint, navTextColor)
			drawScreenOverlays(windowW, windowH, doc.Root)
			rl.EndDrawing()
			return
		}

		rl.BeginDrawing()
		ui.BeginDrawingBorderless(bgColor, borderless, windowW, windowH)
		uiRT := doc.UIRenderTexture()
		if uiRT.ID != 0 {
			src := rl.NewRectangle(0, 0, float32(uiRT.Texture.Width), -float32(uiRT.Texture.Height))
			dst := rl.NewRectangle(0, 0, float32(windowW), float32(windowH))
			rl.DrawTexturePro(uiRT.Texture, src, dst, rl.NewVector2(0, 0), 0, rl.White)
		}
		if bottomChrome > 0 {
			drawLauncherChrome(windowW, windowH, navBg, &studioPanel, presentStudio)
		}
		ui.DrawAnimationOverlays(doc.Root, navRect)
		drawPostChromeOverlays(windowW, windowH, doc.Root, bottomChrome, &studioPanel, presentStudio, inspector, navBg, navHint, navTextColor)
		drawScreenOverlays(windowW, windowH, doc.Root)
		rl.EndDrawing()
	}

	if !isAndroidApp() {
		tray.Start(tray.Config{
			Icon:    appicon.TrayIconBytes(),
			Tooltip: appInitWindowTitle(),
			OnShow:  func() { trayShowPending.Store(true) },
			OnQuit:  func() { trayQuitPending.Store(true) },
		})
		defer tray.Stop()
	}

	// Re-arm after tray/setup so scene-load grace covers the first presents,
	// not wall-clock spent building the title index / chrome.
	idlePolicy.NoteSceneLoad(time.Now())
	if examples.PublicDemoMode() || debugVerbose() {
		fmt.Fprintln(os.Stderr, "Gru: presenting first frame (window stays hidden until ready)…")
	}

	// ── Main loop ─────────────────────────────────────────────────────────────
	for !rl.WindowShouldClose() {
		if androidBackPressed() && !ui.OverlayBlocksSceneInput() {
			break
		}
		frameStart := time.Now()
		windowFocused := rl.IsWindowFocused()
		frameWake := ui.DrainWakeSignals()
		ui.PreparePointerInput()
		frameWake = frameWake.Merge(sampleInputWake(&prevMouse, windowFocused))
		frameWake = frameWake.Merge(ui.SampleChromeHoverWake(windowFocused, windowW, windowH))
		if doc != nil && doc.Root != nil {
			frameWake = frameWake.Merge(ui.SampleEditorKeyboardWake(doc.Root))
		}
		if ui.TypingGestureActive() {
			frameWake.Add(ui.WakeKeyboard, "typing-gesture")
		}
		if ui.ScrollGestureActive() {
			frameWake.Add(ui.WakeScroll, "scroll-gesture")
		}
		if ui.PointerClickPending() || ui.PointerInputActive() {
			frameWake.Add(ui.WakeInput, "pointer-latch")
		}
		if windowFocused && (frameWake.Reasons&ui.WakeInput != 0 ||
			ui.PointerClickPending() || ui.PointerInputActive()) {
			idlePolicy.NoteInteractiveWake(frameStart)
			if lastTargetFPS != ui.ActiveFPS {
				rl.SetTargetFPS(int32(ui.ActiveFPS))
				lastTargetFPS = ui.ActiveFPS
			}
		}
		// Deep idle runs at ~10 FPS — bump immediately on wheel so the next ticks
		// are sampled fast enough for slow scroll (policy settles at ScrollFPS or ActiveFPS).
		if frameWake.Reasons&ui.WakeScroll != 0 && lastTargetFPS < ui.ScrollFPS {
			idlePolicy.NoteInteractiveWake(frameStart)
			rl.SetTargetFPS(int32(ui.ScrollFPS))
			lastTargetFPS = ui.ScrollFPS
		}
		// Same for editor typing: deep idle misses key repeat and GetCharPressed ticks.
		if (frameWake.Reasons&ui.WakeKeyboard != 0 || ui.TypingGestureActive()) && lastTargetFPS < ui.ActiveFPS {
			idlePolicy.NoteInteractiveWake(frameStart)
			rl.SetTargetFPS(int32(ui.ActiveFPS))
			lastTargetFPS = ui.ActiveFPS
		}
		ui.SetActiveDocument(doc)
		dt := rl.GetFrameTime()

		// ── Drain main-thread callbacks from background goroutines ────────────
		// Must run before Update so any goroutine results (async image loads,
		// network responses, etc.) are visible in the same frame they arrive.
		queueDrained := doc.DrainQueueCount()
		if queueDrained > 0 {
			frameWake.Add(ui.WakeDataUpdate, "queue-main")
		}

		prevW, prevH := windowW, windowH
		syncWindowFromDisplay(doc, titleBar, borderless)
		// Borderless edge drag updates the OS client area in-frame; run title bar
		// before resize hold so IsResizing() matches the current gesture.
		if borderless {
			syncWindowFromDisplay(doc, titleBar, borderless)
			applyNativeBorderlessRoundedCorners(titleBar.BorderlessRoundedChrome())
			ui.SetBorderlessRoundedChrome(titleBar.BorderlessRoundedChrome())
			ui.BindFillClientChromeTitleBar(titleBar)
		}
		dimChanged := windowW != prevW || windowH != prevH
		// Use actual client-size change only — IsWindowResized() alone can stay
		// true at the minimum width and was forcing full relayout every frame.
		frameResized := dimChanged
		if frameResized {
			ui.ArmViewportScrollRecovery(2)
			resizeHold.NoteDimensionChange(frameStart)
			ui.MarkWebViewHostsResize()
			doc.ForceFullLayout()
			doc.InvalidatePaint()
		}
		if rl.IsWindowReady() {
			zoomed := rl.IsWindowMaximized()
			if zoomed != lastWindowMaximized {
				ui.NotifyWebViewLayoutJump()
				doc.InvalidatePaint()
				lastWindowMaximized = zoomed
			}
		}
		if ui.WebViewLayoutJumpActive(frameStart) {
			resizeHold.NoteGesture(frameStart)
			frameWake.Add(ui.WakeResize, "webview-layout-jump")
		}
		if titleBar.IsResizing() {
			resizeHold.NoteGesture(frameStart)
		}
		resizeActive := resizeHold.Active(frameStart) || titleBar.IsResizing() || frameResized
		if resizeActive {
			frameWake.Add(ui.WakeResize, "window")
			if lastTargetFPS != ui.ActiveFPS {
				rl.SetTargetFPS(int32(ui.ActiveFPS))
				lastTargetFPS = ui.ActiveFPS
			}
		}

		navHint := slimNavHint(current, len(factories), activeScene.Title())

		// ── Input: footer Directory button (after overlays — see below) ─────────

		// ── Input: Tab cycles scenes (dev launcher only) ─────────────────────
		if !appReleaseMode() && rl.IsKeyPressed(rl.KeyTab) {
			frameWake.Add(ui.WakeKeyboard|ui.WakeScene, "tab")
			switchScene((current + 1) % len(factories))
		}

		// ── Input: debug keys (dev builds only) ─────────────────────────────
		if !appReleaseMode() {
		// ── Input: F5 — caret / space / idle trace (see debug_caret.go) ────────
		if rl.IsKeyPressed(rl.KeyF5) {
			frameWake.Add(ui.WakeKeyboard, "f5")
			toggleCaretDebug()
		}

		// ── Input: F7 — resize FPS / idle policy trace (see debug_resize_fps.go) ──
		if rl.IsKeyPressed(rl.KeyF7) {
			frameWake.Add(ui.WakeKeyboard, "f7")
			toggleResizeFPSDebug()
		}

		// ── Input: F10 — draw-dirty node trace (see debug_draw_dirty.go) ───────
		if rl.IsKeyPressed(rl.KeyF10) {
			frameWake.Add(ui.WakeKeyboard, "f10")
			toggleDrawDirtyDebug()
		}

		// ── Input: WebView passthrough trace (GRU_WEBVIEW_DEBUG=1 only) ───────
		if rl.IsKeyDown(rl.KeyLeftShift) && rl.IsKeyPressed(rl.KeyF11) {
			frameWake.Add(ui.WakeKeyboard, "f11")
			toggleWebViewDebug()
		}

		// ── Input: F8 — isolated resize propagation log (stderr, see debug_resize.go) ──
		if rl.IsKeyPressed(rl.KeyF8) {
			frameWake.Add(ui.WakeKeyboard, "f8")
			TestResizePropagation()
		}

		// ── Input: F6 toggles one-line/sec benchmark logging ─────────────────
		if rl.IsKeyPressed(rl.KeyF6) {
			frameWake.Add(ui.WakeKeyboard, "f6")
			benchmarkMode = !benchmarkMode
			benchmarkLastLog = frameStart
			benchmarkFrames, benchmarkRedraws, benchmarkBlits = 0, 0, 0
			fmt.Printf("Gru benchmark logging: %v  (text=%s  icons=%s  dpi=%.2f  sdfAtlas=%d  remixAtlas=%d)\n",
				benchmarkMode, ui.UIFontBackend(), ui.Phosphor.IconFontSummary(),
				ui.DisplayScale, ui.EffectiveSDFAtlasSize(), ui.EffectiveRemixAtlasSize())
		}

		// ── Input: F11 toggles perf overlay in nav bar ────────────────────────
		if rl.IsKeyPressed(rl.KeyF11) {
			frameWake.Add(ui.WakeKeyboard|ui.WakeOverlay, "f11")
			ui.ShowPerfOverlay = !ui.ShowPerfOverlay
		}
		} // end !appReleaseMode debug keys

		// ── Input: F9 toggles borderless window + custom title bar (desktop) ──
		if appUsesCustomChrome() && rl.IsKeyPressed(rl.KeyF9) {
			frameWake.Add(ui.WakeKeyboard|ui.WakeResize, "f9")
			toggleBorderlessChrome(&borderless, doc, titleBar, activeScene, windowW, windowH)
		}

		// ── Per-scene update ──────────────────────────────────────────────────
		t0 := time.Now()
		if trayShowPending.Swap(false) {
			platformWindowShow()
			appicon.ReapplyWindowIcon()
		}
		if trayQuitPending.Swap(false) {
			closeRequested = true
		}
		if closeRequested {
			if activeScene != nil {
				activeScene.Destroy()
				activeScene = nil
			}
			break
		}
		ui.PrepareOverlayHitTest(doc.Root)
		if borderless {
			titleBar.Update(windowW, windowH)
			// Include titleClickPending — IsDragging alone left a hole before the
			// 4px drag-out where WebView could raise and eat the gesture (§13.5).
			ui.SetChromeTitleBarDragging(titleBar.IsDragging() || titleBar.IsTitleClickPending())
			ui.SetChromeWindowMoving(titleBar.IsDragging() || titleBar.IsTitleClickPending() || titleBar.IsResizing() || ui.FillClientChromeResizing())
			ui.SetWheelSuppressBandY(ui.TitleBarHeight)
		} else {
			ui.SetChromeWindowMoving(false)
			ui.SetWheelSuppressBandY(0)
		}
		ui.PrepareWheelScroll(doc.Root)
		// Layout before pointer handling. Re-layout while the button is down so the
		// first click at idle FPS sees correct switch bounds (not just when dirty).
		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}
		// Hovering a list-tile switch or ribbon cell wakes ActiveFPS before the click.
		listTileWake := ui.CollectListTileSwitchWake(doc.Root)
		ribbonWake := ui.CollectRibbonIconWake(doc.Root)
		frameWake = frameWake.Merge(listTileWake).Merge(ribbonWake)
		if (listTileWake.Any() || ribbonWake.Any()) && lastTargetFPS != ui.ActiveFPS {
			rl.SetTargetFPS(int32(ui.ActiveFPS))
			lastTargetFPS = ui.ActiveFPS
		}
		// Switch rows before the full tree so the latched click reaches the toggle first.
		drawerTop := float32(0)
		if borderless {
			drawerTop = ui.TitleBarHeight
		}
		bottomChrome := demoNavBottom(activeScene)
		ui.SetOverlayChromeInsets(drawerTop, float32(bottomChrome))

		studioState := studioToolState(current, len(factories), activeScene.Title(), benchmarkMode, borderless, inspector)
		if bottomChrome > 0 && !appReleaseMode() && !examples.PublicDemoMode() {
			switch studioPanel.Update(windowW, windowH, studioState) {
			case studio.ActionOpenDirectory:
				dest := examples.DirectorySceneTitle
				if examples.PublicDemoMode() {
					dest = examples.PublicDirectorySceneTitle
				}
				if examples.NavigateToScene(dest) {
					frameWake.Add(ui.WakeScene, "studio-directory")
				}
			case studio.ActionNextScene:
				frameWake.Add(ui.WakeKeyboard|ui.WakeScene, "studio-tab")
				switchScene((current + 1) % len(factories))
			case studio.ActionToggleBenchmark:
				frameWake.Add(ui.WakeKeyboard, "studio")
				benchmarkMode = !benchmarkMode
				benchmarkLastLog = frameStart
				benchmarkFrames, benchmarkRedraws, benchmarkBlits = 0, 0, 0
			case studio.ActionTogglePerfOverlay:
				frameWake.Add(ui.WakeKeyboard|ui.WakeOverlay, "studio")
				ui.ShowPerfOverlay = !ui.ShowPerfOverlay
			case studio.ActionToggleResizeFPS:
				frameWake.Add(ui.WakeKeyboard, "studio")
				toggleResizeFPSDebug()
			case studio.ActionToggleCaretDebug:
				frameWake.Add(ui.WakeKeyboard, "studio")
				toggleCaretDebug()
			case studio.ActionToggleDrawDirty:
				frameWake.Add(ui.WakeKeyboard, "studio")
				toggleDrawDirtyDebug()
			case studio.ActionToggleWebViewDebug:
				frameWake.Add(ui.WakeKeyboard, "studio")
				toggleWebViewDebug()
			case studio.ActionToggleChrome:
				frameWake.Add(ui.WakeKeyboard|ui.WakeResize, "studio")
				toggleBorderlessChrome(&borderless, doc, titleBar, activeScene, windowW, windowH)
			case studio.ActionToggleInspector:
				frameWake.Add(ui.WakeKeyboard, "studio")
				inspector.Toggle()
			}
		}
		studioBlocks := bottomChrome > 0 && !appReleaseMode() && !examples.PublicDemoMode() && studioPanel.BlocksSceneInput(windowW, windowH)

		// Overlays before scene pointer handling so taps hit drawer/sheet controls
		// instead of list rows underneath.
		ui.ModalMgr.Update(dt)
		ui.CommandPaletteMgr.Update(dt)
		ui.DrawerMgr.Update(dt)
		ui.BottomSheetMgr.Update(dt)
		ui.ColorPickerMgr.Update(dt)
		ui.DatePickerMgr.Update(dt)
		ui.DateRangePickerMgr.Update(dt)
		if bottomChrome > 0 && !examples.PublicDemoMode() && !ui.OverlayBlocksSceneInput() &&
			ui.PointerClickConsume(navDirectoryButtonRect(windowW, windowH)) {
			if examples.NavigateToScene(examples.DirectorySceneTitle) {
				frameWake.Add(ui.WakeScene, "directory")
			}
		}
		if bottomChrome > 0 && examples.PublicDemoMode() && !ui.OverlayBlocksSceneInput() &&
			ui.PointerClickConsume(navDirectoryButtonRect(windowW, windowH)) {
			if examples.NavigateToScene(examples.PublicDirectorySceneTitle) {
				frameWake.Add(ui.WakeScene, "demo-index")
			}
		}
		resizeInputBlock := borderless && (titleBar.IsResizing() || titleBar.IsDragging() || ui.FillClientChromeResizing() || resizeHold.KeepActiveFPS(frameStart))
		mouse := rl.GetMousePosition()
		titleBarBlocks := borderless && mouse.Y < ui.TitleBarHeight
		ui.SetScenePointerBlocked(ui.OverlayBlocksSceneInput() || resizeInputBlock || titleBarBlocks || studioBlocks)

		if borderless && titleBar.IsDragging() {
			ui.ResetWheelScrollGesture()
		}
		if !ui.ScenePointerBlocked() {
			ui.ProcessSwitchListTilePointers(doc.Root, dt)
		}
		ui.BeginFrameCursor()
		doc.Root.Update(dt)
		// Widgets may Wake() mid-Update (e.g. Toggle flip). Drain now and bump
		// FPS immediately — waiting until the next frame at DeepIdle (~10 FPS)
		// makes rapid clicks feel laggy.
		if midWake := ui.DrainWakeSignals(); midWake.Any() {
			frameWake = frameWake.Merge(midWake)
			if windowFocused && (midWake.Reasons&ui.WakeInput != 0 ||
				midWake.Reasons&ui.WakeAnimation != 0) {
				idlePolicy.NoteInteractiveWake(time.Now())
				if lastTargetFPS != ui.ActiveFPS {
					rl.SetTargetFPS(int32(ui.ActiveFPS))
					lastTargetFPS = ui.ActiveFPS
				}
			}
		}
		// Dismiss context menu before focus routing so the same outside-click can
		// hand keyboard focus back to native TextEditor / TextInput (§5).
		ui.ContextMenuMgr.Update(dt)
		activeScene.OnUpdate(doc, dt)
		if rl.IsWindowReady() {
			doc.SetPlatformWindowHandle(uintptr(rl.GetWindowHandle()))
		}
		ui.SetScenePointerBlocked(false)
		if borderless {
			titleBar.ApplyResizeCursor(windowW, windowH)
		}

		animationWake := ui.CollectAnimationWake(doc.Root)
		if !windowFocused {
			animationWake = ui.WakeSummary{}
		}
		frameWake = frameWake.Merge(animationWake)
		if windowFocused {
			frameWake = frameWake.Merge(ui.CollectInteractionOverlayWake(doc.Root))
		}
		updateMs := float32(time.Since(t0).Microseconds()) / 1000.0

		// ── Inspector update (F12 widget debugger — dev/demo only) ────────────
		if !appReleaseMode() {
			inspector.Update(doc.Root, doc, windowW, windowH)
		}

		// ── Overlay managers (tooltip / toast fade) ──
		ui.Tooltips.Update(dt)
		ui.Toasts.Update(dt)
		ui.NotificationCenterMgr.Update(dt)
		if anim := ui.OverlayAnimationWake(); anim != 0 {
			frameWake.Add(anim, "overlay-anim")
		}

		// ── Layout (dirty subtree only) ───────────────────────────────────────
		t1 := time.Now()
		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}
		if ui.ViewportScrollRecoveryActive() {
			ui.InvalidateViewportScrollFastPath(doc.Root)
			doc.Root.MarkDirty()
			doc.Root.Layout()
			ui.TickViewportScrollRecovery()
		}
		layoutMs := float32(time.Since(t1).Microseconds()) / 1000.0

		// WebView HWND bounds follow layout (after Layout, before focus routing).
		ui.DrainFillClientChromeResize()
		ui.SyncWebViewHosts(doc)
		focusInputBlock := resizeInputBlock
		if titleBarBlocks {
			if pos, ok := ui.PeekFocusHandoffClick(); ok && pos.Y >= ui.TitleBarHeight {
				// Content click — route focus even if the cursor is now over the title bar.
			} else {
				focusInputBlock = true
			}
		}
		if !ui.OverlayBlocksSceneInput() && !focusInputBlock {
			ui.RouteScenePointerFocus(doc)
		}

		frameCount++

		// Log cache hit rate every 60 frames (~1 s at 60 FPS).
		if frameCount%60 == 0 {
			rl.TraceLog(rl.LogInfo, "Gru cache hit rate: %.1f%% (%.1f× SSAA)",
				ui.CacheHitRate()*100, ui.RenderScale)
		}

		var navRect rl.Rectangle
		if bottomChrome > 0 {
			navRect = navBarRect(windowW, windowH)
		}

		// ── Draw ──────────────────────────────────────────────────────────────
		t2 := time.Now()
		fullRedraw := false
		resizePaintHold := frameResized || ui.WebViewLayoutJumpActive(frameStart) ||
			resizeHold.Active(frameStart) || titleBar.IsResizing()
		if resizePaintHold {
			doc.InvalidatePaint()
			if !doc.NeedsRedraw() {
				doc.Root.MarkDrawDirty()
			}
		}

		if !windowFocused && doc != nil && doc.Root != nil && !doc.Root.IsDirty() &&
			!(resizeHold.Active(frameStart) || titleBar.IsResizing()) {
			ui.ClearDrawDirtySubtree(doc.Root)
		}

		if ui.SupersamplingActive() {
			// ── 2× supersampled pass (skip when cache is valid) ───────────────
			if doc.NeedsRedraw() && ui.SuperTargetDrawable() {
				fullRedraw = true
				ui.BeginSuperFrame(bgColor, borderless, windowW, windowH)

				doc.Root.Draw()

				if bottomChrome > 0 {
					drawLauncherChrome(windowW, windowH, navBg, &studioPanel, studioState)
				}

				// Perf overlay (F11) — inside super pass for AA
				if perfOverlayVisible() {
					drawPerfOverlay(windowW, windowH)
				}

				// Title bar drawn on top of everything inside the super pass.
				if borderless {
					titleBar.Draw()
				}

				ui.EndSuperFrame()
				ui.RecordCacheMiss()
				// Bubbled drawDirty on transparent wrappers (editor-area, split) may
				// outlive child Draw(); match cache-hit handling so idle can reach 10 FPS.
				ui.ClearDrawDirtySubtree(doc.Root)
			} else if doc.NeedsRedraw() {
				// GL/window not ready for off-screen pass this frame — retry next frame.
				doc.Root.MarkDrawDirty()
				ui.RecordCacheMiss()
			} else {
				ui.RecordCacheHit()
				if doc.Root != nil {
					ui.ClearDrawDirtySubtree(doc.Root)
				}
			}

			// ── Blit 2× → 1× with GPU bilinear downscale ─────────────────────
			rl.BeginDrawing()
			ui.BeginDrawingBorderless(bgColor, borderless, windowW, windowH)
			if ui.SuperTargetDrawable() {
				ui.BlitToScreenBorderless(windowW, windowH, borderless)
			} else if doc.NeedsRedraw() {
				// SSAA RT not ready yet (early frames / GL init) — draw once to avoid a black flash.
				doc.Root.Draw()
				if bottomChrome > 0 {
					drawLauncherChrome(windowW, windowH, navBg, &studioPanel, studioState)
				}
				if borderless {
					titleBar.Draw()
				}
				ui.ClearDrawDirtySubtree(doc.Root)
			}
			ui.DrawAnimationOverlays(doc.Root, navRect)
			drawPostChromeOverlays(windowW, windowH, doc.Root, bottomChrome, &studioPanel, studioState, inspector, navBg, navHint, navTextColor)
			if engineDebugHUDVisible() && rl.IsWindowReady() {
				rl.DrawFPS(windowW-80, 6)
			}
			// Overlay managers (tooltips, modals, pickers, toasts) through SSAA when active.
			drawScreenOverlays(windowW, windowH, doc.Root)
			rl.EndDrawing()
			ui.PresentWebViewHosts()
			if !ui.OverlayBlocksSceneInput() && !resizeInputBlock {
				ui.RouteScenePointerFocusAfterPresent(doc)
			}
			if !taskbarIconAfterFrame {
				appicon.ReapplyWindowIcon()
				taskbarIconAfterFrame = true
			}
			if !windowShown {
				if rl.IsWindowHidden() {
					rl.ClearWindowState(uint32(rl.FlagWindowHidden))
				}
				platformWindowFocus()
				windowShown = true
				appicon.ReapplyWindowIcon()
				idlePolicy.NoteSceneLoad(time.Now())
			}

		} else {
			// ── Fallback: direct 1× rendering (no supersampling) ─────────────
			doc.RefreshRenderCache()
			rl.BeginDrawing()
			ui.BeginDrawingBorderless(bgColor, borderless, windowW, windowH)

			if doc.NeedsRedraw() {
				uiRT := doc.UIRenderTexture()
				if uiRT.ID != 0 {
					fullRedraw = true
					// Render into the 1× cache texture
					rl.BeginTextureMode(uiRT)
					if borderless {
						rl.ClearBackground(rl.Blank)
						ui.DrawBorderlessWindowFill(windowW, windowH, bgColor)
					} else {
						rl.ClearBackground(bgColor)
					}
					doc.Root.Draw()
					if bottomChrome > 0 {
						drawLauncherChrome(windowW, windowH, navBg, &studioPanel, studioState)
					}
					if perfOverlayVisible() {
						drawPerfOverlay(windowW, windowH)
					}
					if borderless {
						titleBar.Draw()
					}
					rl.EndTextureMode()
					src := rl.NewRectangle(0, 0, float32(uiRT.Texture.Width), -float32(uiRT.Texture.Height))
					dst := rl.NewRectangle(0, 0, float32(windowW), float32(windowH))
					rl.DrawTexturePro(uiRT.Texture, src, dst, rl.NewVector2(0, 0), 0, rl.White)
					ui.ClearDrawDirtySubtree(doc.Root)
				} else {
					fullRedraw = true
					// No cache: draw directly to screen (BeginDrawingBorderless already ran)
					doc.Root.Draw()
					if bottomChrome > 0 {
						drawLauncherChrome(windowW, windowH, navBg, &studioPanel, studioState)
					}
					if perfOverlayVisible() {
						drawPerfOverlay(windowW, windowH)
					}
					if borderless {
						titleBar.Draw()
					}
					ui.ClearDrawDirtySubtree(doc.Root)
				}
				ui.RecordCacheMiss()
			} else {
				// Cache valid: blit without redraw
				uiRT := doc.UIRenderTexture()
				if uiRT.ID != 0 {
					src := rl.NewRectangle(0, 0, float32(uiRT.Texture.Width), -float32(uiRT.Texture.Height))
					dst := rl.NewRectangle(0, 0, float32(windowW), float32(windowH))
					rl.DrawTexturePro(uiRT.Texture, src, dst, rl.NewVector2(0, 0), 0, rl.White)
				}
				if bottomChrome > 0 {
					drawLauncherChrome(windowW, windowH, navBg, &studioPanel, studioState)
				}
				ui.RecordCacheHit()
			}

			ui.DrawAnimationOverlays(doc.Root, navRect)
			drawPostChromeOverlays(windowW, windowH, doc.Root, bottomChrome, &studioPanel, studioState, inspector, navBg, navHint, navTextColor)
			if engineDebugHUDVisible() && rl.IsWindowReady() {
				rl.DrawFPS(windowW-80, 6)
			}
			// Overlay managers always drawn on top of everything
			drawScreenOverlays(windowW, windowH, doc.Root)
			rl.EndDrawing()
			ui.PresentWebViewHosts()
			if !ui.OverlayBlocksSceneInput() && !resizeInputBlock {
				ui.RouteScenePointerFocusAfterPresent(doc)
			}
			if !taskbarIconAfterFrame {
				appicon.ReapplyWindowIcon()
				taskbarIconAfterFrame = true
			}
			if !windowShown {
				if rl.IsWindowHidden() {
					rl.ClearWindowState(uint32(rl.FlagWindowHidden))
				}
				platformWindowFocus()
				windowShown = true
				appicon.ReapplyWindowIcon()
				idlePolicy.NoteSceneLoad(time.Now())
			}
		}

		drawMs := float32(time.Since(t2).Microseconds()) / 1000.0
		cacheHitThisFrame := !fullRedraw
		if fullRedraw {
			benchmarkRedraws++
		} else {
			benchmarkBlits++
		}
		benchmarkFrames++
		frameEnd := time.Now()
		// Re-sample resize + pointer at frame end so idle policy does not drop FPS
		// on clean blit frames using stale start-of-frame wake (common during drag).
		endWake := frameWake
		if pendingNavWake.Any() {
			endWake = endWake.Merge(pendingNavWake)
			pendingNavWake = ui.WakeSummary{}
		}
		if !windowFocused {
			endWake = ui.WakeSummaryForBackground(endWake)
		}
		if borderless && titleBar.IsResizing() {
			resizeHold.NoteGesture(frameEnd)
			endWake.Add(ui.WakeResize, "window-end")
		}
		if rl.IsMouseButtonDown(rl.MouseLeftButton) || rl.IsMouseButtonDown(rl.MouseRightButton) {
			endWake.Add(ui.WakeInput, "mouse-button-end")
		}
		if ui.PointerClickPending() {
			endWake.Add(ui.WakeInput, "pointer-pending-end")
			idlePolicy.NoteInteractiveWake(frameEnd)
		}
		if ui.TypingGestureActive() {
			endWake.Add(ui.WakeKeyboard, "typing-gesture-end")
		}
		if fullRedraw && resizeHold.RecentActivity(frameEnd) {
			resizeHold.NoteHeavyFrame(frameEnd)
		}
		resizeBurstHold := resizeHold.Active(frameEnd) || titleBar.IsResizing() || ui.FillClientChromeResizing()
		resizeHoldActive := resizeHold.KeepActiveFPS(frameEnd) || titleBar.IsResizing() || ui.FillClientChromeResizing()
		if doc != nil && doc.Root != nil && doc.Root.IsDirty() && resizeHold.RecentActivity(frameEnd) {
			resizeHold.NoteLayoutSettling(frameEnd)
			resizeBurstHold = true
			resizeHoldActive = true
		}
		if resizeHoldWasActive && !resizeBurstHold && doc != nil && doc.Root != nil && !doc.Root.IsDirty() {
			ui.ClearDrawDirtySubtree(doc.Root)
		}
		resizeHoldWasActive = resizeBurstHold
		cleanForIdle := cacheHitThisFrame && !doc.NeedsRedraw() && !resizeBurstHold && queueDrained == 0
		var idleBlockers ui.WakeReason
		if ui.AnyDropdownOpen(doc.Root) {
			idleBlockers |= ui.WakeOverlay
		}
		if resizeBurstHold {
			idleBlockers |= ui.WakeResize
		}
		idleBlockers |= ui.OverlayIdleBlockers()
		idleWake := ui.WakeSummaryForIdlePolicy(endWake)
		prevTargetFPS := lastTargetFPS
		targetFPS := idlePolicy.Update(frameEnd, idleWake, cleanForIdle, idleBlockers, resizeBurstHold)
		if !windowFocused && !resizeBurstHold && idleBlockers == 0 &&
			!ui.WebViewHostsActive() &&
			doc != nil && doc.Root != nil && !doc.Root.IsDirty() {
			targetFPS = ui.DeepIdleFPS
		}
		if ui.WebViewHostsActive() {
			webInteractive := resizeHoldActive || idleWake.Reasons != 0 || endWake.Reasons != 0 ||
				ui.PointerClickPending() || !cleanForIdle || ui.WebViewHostHoldsKeyboard() ||
				ui.FillClientChromeResizing()
			// Full Client: mouse lives in the HWND — raylib sees no input wake.
			// Keep ActiveFPS while focused so Present/pump and chrome.resize stay live.
			if ui.WebViewFillClientHostsActive() && windowFocused {
				webInteractive = true
			}
			if targetFPS < ui.WebViewIdleFPS {
				targetFPS = ui.WebViewIdleFPS
			}
			if ui.WebViewFillClientHostsActive() && windowFocused && targetFPS < ui.ActiveFPS {
				targetFPS = ui.ActiveFPS
			}
			if !webInteractive && targetFPS > ui.WebViewIdleFPS {
				targetFPS = ui.WebViewIdleFPS
			}
		}
		ui.UpdateWebViewPresentBudget(targetFPS)
		// One-shot SSAA redraw after async data only — do not drop FPS after user
		// clicks, scene navigation, or other interactive redraws.
		if windowFocused && targetFPS == ui.ActiveFPS && !resizeBurstHold &&
			idleWake.Reasons == 0 && fullRedraw && doc != nil && doc.Root != nil &&
			!doc.NeedsRedraw() && !doc.Root.IsDirty() && queueDrained == 0 &&
			idlePolicy.LastReason() == ui.WakeDataUpdate {
			targetFPS = ui.DeepIdleFPS
		}
		if windowFocused && ui.PointerClickPending() && targetFPS < ui.ActiveFPS {
			targetFPS = ui.ActiveFPS
		}
		policyChanged := targetFPS != lastTargetFPS
		if policyChanged {
			rl.SetTargetFPS(int32(targetFPS))
			lastTargetFPS = targetFPS
			if debugVerbose() && !resizeFPSDebug {
				fmt.Printf("Gru idle policy: %s target=%d frameWake=%s lastWake=%s cacheHit=%t redraw=%t resizeHold=%t\n",
					idlePolicy.State(), targetFPS, endWake.Reasons, idlePolicy.LastReason(), cacheHitThisFrame, fullRedraw, resizeHoldActive)
			}
		} else if debugVerbose() && ui.ShowPerfOverlay && time.Since(lastPolicyLog) >= 5*time.Second {
			lastPolicyLog = time.Now()
			fmt.Printf("Gru frame: state=%s target=%d wake=%s cacheHit=%t redraw=%t queue=%d\n",
				idlePolicy.State(), targetFPS, frameWake.Reasons, cacheHitThisFrame, fullRedraw, queueDrained)
		}
		if !appReleaseMode() && benchmarkMode && time.Since(benchmarkLastLog) >= time.Second {
			elapsed := time.Since(benchmarkLastLog).Seconds()
			actualFPS := float64(benchmarkFrames) / elapsed
			pm := processStats.Sample()
			fmt.Printf("Gru bench scene=%q fps=%.1f target=%d scale=%.1fx dpi=%.2f sdfAtlas=%d state=%s wake=%s anim=%d[%s] redraw=%d blit=%d hit=%.0f%% queue=%d cpu=%.1f%% rss=%.1fMB heap=%.1fMB\n",
				activeScene.Title(), actualFPS, targetFPS, ui.RenderScale, ui.DisplayScale, ui.EffectiveSDFAtlasSize(),
				idlePolicy.State(), frameWake.Reasons,
				len(animationWake.Sources), shortSources(animationWake.Sources, 3), benchmarkRedraws, benchmarkBlits,
				ui.CacheHitRate()*100, queueDrained, pm.CPUPercent, pm.WorkingMB, pm.HeapMB)
			benchmarkLastLog = frameStart
			benchmarkFrames, benchmarkRedraws, benchmarkBlits = 0, 0, 0
		}

		observeIdleGuard(idleGuardInput{
			CleanForIdle: cleanForIdle,
			TargetFPS:    targetFPS,
			EndWake:      endWake,
			Root:         doc.Root,
		})

		ui.PaceFrame(frameStart, targetFPS)
		frameWallMs := float32(time.Since(frameStart).Seconds()) * 1000.0

		// ── Update PerfStats for Inspector and overlay ────────────────────────
		ui.PerfStats.UpdateMs = updateMs
		ui.PerfStats.LayoutMs = layoutMs
		ui.PerfStats.DrawMs = drawMs
		ui.PerfStats.TotalMs = frameWallMs
		ui.PerfStats.WakeReasons = frameWake.Reasons
		ui.PerfStats.IdleState = idlePolicy.State()
		ui.PerfStats.TargetFPS = targetFPS
		ui.PerfStats.FullRedraw = fullRedraw
		ui.PerfStats.QueueDrained = queueDrained
		ui.PerfStats.AnimationActive = animationWake.Any()
		ui.PerfStats.AnimationCount = len(animationWake.Sources)
		ui.PerfStats.AnimationSources = shortSources(animationWake.Sources, 3)

		holdSnap := resizeHold.Snapshot(frameEnd)
		observeDrawDirtyFrame(doc.Root, fullRedraw, endWake, doc.NeedsRedraw(), doc.Root.IsDirty())
		observeCaretDebugFrame(doc.Root, resizeFPSFrameInput{
			Now:              frameEnd,
			Scene:            activeScene.Title(),
			WindowW:          windowW,
			WindowH:          windowH,
			FrameResized:     frameResized,
			IsResizing:       borderless && titleBar.IsResizing(),
			ResizeHoldActive: resizeHoldActive,
			Hold:             holdSnap,
			EndWake:          endWake,
			CleanForIdle:     cleanForIdle,
			CacheHit:         cacheHitThisFrame,
			RootDirty:        doc.Root.IsDirty(),
			NeedsRedraw:      doc.NeedsRedraw(),
			QueueDrained:     queueDrained,
			FullRedraw:       fullRedraw,
			UpdateMs:         updateMs,
			LayoutMs:         layoutMs,
			DrawMs:           drawMs,
			TotalMs:          frameWallMs,
			TargetFPS:        targetFPS,
			PrevTargetFPS:    prevTargetFPS,
			IdleState:        idlePolicy.State(),
			PolicyChanged:    policyChanged,
		})
		observeResizeFPSFrame(resizeFPSFrameInput{
			Now:              frameEnd,
			Scene:            activeScene.Title(),
			WindowW:          windowW,
			WindowH:          windowH,
			FrameResized:     frameResized,
			IsResizing:       borderless && titleBar.IsResizing(),
			ResizeHoldActive: resizeHoldActive,
			Hold:             holdSnap,
			EndWake:          endWake,
			CleanForIdle:     cleanForIdle,
			CacheHit:         cacheHitThisFrame,
			RootDirty:        doc.Root.IsDirty(),
			NeedsRedraw:      doc.NeedsRedraw(),
			QueueDrained:     queueDrained,
			FullRedraw:       fullRedraw,
			UpdateMs:         updateMs,
			LayoutMs:         layoutMs,
			DrawMs:           drawMs,
			TotalMs:          frameWallMs,
			TargetFPS:        targetFPS,
			PrevTargetFPS:    prevTargetFPS,
			IdleState:        idlePolicy.State(),
			PolicyChanged:    policyChanged,
		})
	}
	if isAndroidApp() {
		os.Exit(0)
	}
}

// drawScreenOverlays renders tooltips, modals, context menus, pickers, and toasts.
// When SSAA is active they share one supersampled overlay pass so SDF text stays
// as sharp as in-tree widgets (calendar popups and tooltips were soft at 1×).
func drawScreenOverlays(windowW, windowH int32, root ui.Node) {
	if ui.SupersamplingActive() && ui.OverlayTargetDrawable() {
		ui.BeginOverlaySuperFrame()
		ui.Tooltips.Draw()
		ui.ModalMgr.Draw()
		ui.CommandPaletteMgr.Draw()
		ui.DrawerMgr.Draw()
		ui.BottomSheetMgr.Draw()
		ui.ContextMenuMgr.Draw()
		ui.ColorPickerMgr.Draw()
		ui.DatePickerMgr.Draw()
		ui.DateRangePickerMgr.Draw()
		ui.NotificationCenterMgr.Draw()
		ui.Toasts.Draw()
		ui.DrawOpenMenuPopups(root)
		ui.EndOverlaySuperFrame()
		ui.BlitOverlayToScreen(windowW, windowH)
		return
	}
	ui.Tooltips.Draw()
	ui.ModalMgr.Draw()
	ui.CommandPaletteMgr.Draw()
	ui.DrawerMgr.Draw()
	ui.BottomSheetMgr.Draw()
	ui.ContextMenuMgr.Draw()
	ui.ColorPickerMgr.Draw()
	ui.DatePickerMgr.Draw()
	ui.DateRangePickerMgr.Draw()
	ui.NotificationCenterMgr.Draw()
	ui.Toasts.Draw()
	ui.DrawOpenMenuPopups(root)
}

// drawInteractionOverlay renders hover/focus/press chrome through the same
// supersampled overlay path as the main UI cache so text/icons do not degrade
// when interaction overlays are active.
func drawInteractionOverlay(windowW, windowH int32, root ui.Node) {
	if root == nil {
		return
	}
	if ui.SupersamplingActive() && ui.OverlayTargetDrawable() {
		ui.BeginOverlaySuperFrame()
		ui.DrawInteractionOverlays(root)
		ui.EndOverlaySuperFrame()
		ui.BlitOverlayToScreen(windowW, windowH)
		return
	}
	ui.DrawInteractionOverlays(root)
}

// drawPerfOverlay renders the F11 performance strip on the right side of the nav bar.
// It is drawn inside whatever render pass is active (super or 1×).
func drawPerfOverlay(windowW, windowH int32) {
	s := ui.PerfStats
	cacheStr := "miss"
	if s.CacheHit {
		cacheStr = "HIT"
	}
	hitPct := int(ui.CacheHitRate() * 100)
	line := fmt.Sprintf("U:%.1f L:%.1f D:%.1f ms  cache:%s %d%%  %.1f× SSAA  %s/%dfps  icons:%s  wake:%s  anim:%d %s",
		s.UpdateMs, s.LayoutMs, s.DrawMs, cacheStr, hitPct, ui.RenderScale, s.IdleState, s.TargetFPS,
		ui.Phosphor.IconFontSummary(), s.WakeReasons, s.AnimationCount, s.AnimationSources)
	tw := ui.MeasureChromeText(line, ui.ChromeDimStyle())
	x := float32(windowW) - tw - 12
	bar := rl.NewRectangle(x-4, float32(windowH-navBarH), tw+8, float32(navBarH))
	perfStyle := ui.ChromeDimStyle()
	perfStyle.TextColor = rl.NewColor(160, 200, 230, 180)
	y := ui.ChromeTextCenterY(bar, perfStyle)
	ui.DrawChromeText(line, x, y, perfStyle)
}

func shortSources(sources []string, max int) string {
	if len(sources) == 0 {
		return ""
	}
	seen := make(map[string]struct{}, len(sources))
	out := make([]string, 0, max)
	for _, src := range sources {
		if src == "" {
			continue
		}
		if _, ok := seen[src]; ok {
			continue
		}
		seen[src] = struct{}{}
		out = append(out, src)
		if len(out) == max {
			break
		}
	}
	if len(out) == 0 {
		return ""
	}
	suffix := ""
	if len(seen) < len(sources) || len(sources) > len(out) {
		suffix = "+"
	}
	return strings.Join(out, ",") + suffix
}

func sampleInputWake(prevMouse *rl.Vector2, windowFocused bool) ui.WakeSummary {
	var out ui.WakeSummary
	mouse := rl.GetMousePosition()
	moved := mouse.X != prevMouse.X || mouse.Y != prevMouse.Y
	if moved && ui.WakeOnMouseMove {
		out.Add(ui.WakeInput, "mouse")
	}
	*prevMouse = mouse

	if rl.GetMouseWheelMove() != 0 {
		out.Add(ui.WakeScroll, "wheel")
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) ||
		rl.IsMouseButtonPressed(rl.MouseRightButton) ||
		rl.IsMouseButtonDown(rl.MouseLeftButton) ||
		rl.IsMouseButtonDown(rl.MouseRightButton) {
		out.Add(ui.WakeInput, "mouse-button")
	}
	if rl.IsKeyPressed(rl.KeyTab) ||
		rl.IsKeyPressed(rl.KeyF6) ||
		rl.IsKeyPressed(rl.KeyF8) ||
		rl.IsKeyPressed(rl.KeyF9) ||
		rl.IsKeyPressed(rl.KeyF11) ||
		rl.IsKeyPressed(rl.KeyF12) ||
		rl.IsKeyPressed(rl.KeyEnter) ||
		rl.IsKeyPressed(rl.KeySpace) ||
		rl.IsKeyPressed(rl.KeyBackspace) ||
		rl.IsKeyPressed(rl.KeyEscape) {
		out.Add(ui.WakeKeyboard, "key")
	}
	return out
}

