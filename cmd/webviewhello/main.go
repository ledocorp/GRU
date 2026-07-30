// webviewhello — Gru WebView shell starter (window + title bar + HTML body).
//
//	go run ./cmd/webviewhello
//	go run -tags webview2 ./cmd/webviewhello   # live WebView2 on Windows
//
// Host loop mirrors Studio main.go §13.6 (docs/WEBVIEW2_HOST.md). Do not freestyle
// TitleBar or HWND policy here — see examples/webview_full_demo.go for the scene.
package main

import (
	"fmt"
	"os"

	"github.com/ledocorp/gru/samples/webviewhello"
	"github.com/ledocorp/gru/ui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	initW int32 = 960
	initH int32 = 640
)

func main() {
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetConfigFlags(uint32(rl.FlagMsaa4xHint | rl.FlagVsyncHint |
		rl.FlagWindowResizable | rl.FlagWindowUndecorated | rl.FlagWindowHidden))
	rl.InitWindow(initW, initH, "Hello WebView")
	defer rl.CloseWindow()
	rl.SetExitKey(0)
	rl.SetWindowMinSize(int(ui.MinClientWidth), 320)

	ui.ApplyUIOptimizations()
	ui.RefreshDisplayScale()
	ui.InitDisplayAwareAtlases()
	defer ui.UnloadSDFFonts()
	defer ui.Icons.UnloadAll()
	defer ui.DestroyAllWebViewHosts()

	windowW, windowH := initW, initH
	ui.InitSupersampling(windowW, windowH)
	defer ui.UnloadSupersampling()
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	defer ui.Toasts.Unload()

	doc := ui.NewDocument(windowW, windowH)
	doc.SetChromeTop(ui.TitleBarHeight)
	webviewhello.Build(doc)
	ui.MountBorderlessDocument(doc, windowW, windowH)
	doc.EnableUIRenderTexture(true)
	if rl.IsWindowReady() {
		doc.SetPlatformWindowHandle(uintptr(rl.GetWindowHandle()))
	}
	ui.SyncWebViewHosts(doc)

	shouldClose := false
	titleBar := ui.NewTitleBar("Hello WebView", ui.TitleBarStyleDark,
		func() { shouldClose = true },
		func() { rl.MinimizeWindow() },
		func() {
			ui.NotifyWebViewLayoutJump()
			if rl.IsWindowMaximized() {
				rl.RestoreWindow()
			} else {
				rl.MaximizeWindow()
			}
		})
	titleBar.HandleResize = true
	titleBar.SetSize(windowW, windowH)

	bg := rl.NewColor(228, 229, 235, 255)
	borderless := true
	windowShown := false
	wasMaximized := false
	rl.SetTargetFPS(60)
	ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)
	ui.ApplyNativeBorderlessRoundedCorners(true)

	fmt.Fprintln(os.Stderr, "webviewhello: Studio §13.6 host loop + full-client FillClient")

	for !rl.WindowShouldClose() && !shouldClose {
		ui.PreparePointerInput()
		ui.SetActiveDocument(doc)
		dt := rl.GetFrameTime()
		doc.DrainQueue()

		nw, nh := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
		if nw < 1 {
			nw = windowW
		}
		if nh < 1 {
			nh = windowH
		}
		maximized := rl.IsWindowMaximized()
		if maximized != wasMaximized {
			wasMaximized = maximized
			ui.NotifyWebViewLayoutJump()
			ui.MarkWebViewHostsResize()
		}
		if nw != windowW || nh != windowH {
			windowW, windowH = nw, nh
			ui.SyncBorderlessClientSize(doc, titleBar, windowW, windowH)
			ui.MarkWebViewHostsResize()
		}

		// §13.6 — same order as Studio main.go (do not reorder).
		titleBar.Update(windowW, windowH)
		ui.SetChromeTitleBarDragging(titleBar.IsDragging() || titleBar.IsTitleClickPending())
		ui.SetChromeWindowMoving(titleBar.IsDragging() || titleBar.IsTitleClickPending() || titleBar.IsResizing() || ui.FillClientChromeResizing())
		ui.SetWheelSuppressBandY(ui.TitleBarHeight)
		ui.SyncBorderlessChromeFrame(titleBar)
		titleBar.ApplyResizeCursor(windowW, windowH)
		ui.PrepareWheelScroll(doc.Root)
		ui.ProcessSwitchListTilePointers(doc.Root, dt)

		mouse := rl.GetMousePosition()
		titleBarBlocks := borderless && mouse.Y < ui.TitleBarHeight
		resizeInputBlock := borderless && (titleBar.IsResizing() || titleBar.IsDragging() || ui.FillClientChromeResizing())
		ui.SetScenePointerBlocked(titleBarBlocks || resizeInputBlock)
		ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)

		if titleBar.IsDragging() {
			ui.ResetWheelScrollGesture()
		}

		doc.Root.Update(dt)
		ui.SetScenePointerBlocked(false)

		if doc.Root.IsDirty() || ui.SubtreeLayoutDirty(doc.Root) {
			doc.Root.Layout()
		}
		ui.Toasts.Update(dt)
		ui.Tooltips.Update(dt)

		if rl.IsWindowReady() {
			doc.SetPlatformWindowHandle(uintptr(rl.GetWindowHandle()))
		}
		ui.DrainFillClientChromeResize()
		ui.SyncWebViewHosts(doc)
		ui.UpdateWebViewPresentBudget(ui.ActiveFPS)

		focusInputBlock := resizeInputBlock
		if titleBarBlocks {
			if pos, ok := ui.PeekFocusHandoffClick(); ok && pos.Y >= ui.TitleBarHeight {
				// content handoff — still route
			} else {
				focusInputBlock = true
			}
		}
		if !ui.OverlayBlocksSceneInput() && !focusInputBlock {
			ui.RouteScenePointerFocus(doc)
		}

		if doc.NeedsRedraw() && ui.SuperTargetDrawable() {
			ui.BeginSuperFrame(bg, borderless, windowW, windowH)
			doc.Root.Draw()
			titleBar.Draw()
			ui.EndSuperFrame()
			ui.RecordCacheMiss()
		} else {
			ui.RecordCacheHit()
		}

		rl.BeginDrawing()
		ui.BeginDrawingBorderless(bg, borderless, windowW, windowH)
		if ui.SuperTargetDrawable() {
			ui.BlitToScreenBorderless(windowW, windowH, borderless)
		}
		ui.Tooltips.Draw()
		ui.Toasts.Draw()
		rl.EndDrawing()
		ui.PresentWebViewHosts()
		if !ui.OverlayBlocksSceneInput() && !resizeInputBlock {
			ui.RouteScenePointerFocusAfterPresent(doc)
		}

		if !windowShown {
			rl.ClearWindowState(uint32(rl.FlagWindowHidden))
			windowShown = true
		}
	}
}
