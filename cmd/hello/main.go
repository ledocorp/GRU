// hello is the public Gru sample — GRU core spine then a tiny scene (docs/GRU_CORE.md).
//
//	go run ./cmd/hello
package main

import (
	"fmt"
	"os"

	"github.com/ledocorp/gru/samples/hello"
	"github.com/ledocorp/gru/ui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	// Modest default client — SSAA RTs scale with window size; 960×640 made hello
	// look ~40 MB heavier than calc for no product reason (see docs/LEAN_GRU_SPIKE.md).
	initW int32 = 720
	initH int32 = 520
)

func main() {
	rl.SetTraceLogLevel(rl.LogWarning)
	rl.SetConfigFlags(uint32(rl.FlagMsaa4xHint | rl.FlagVsyncHint |
		rl.FlagWindowResizable | rl.FlagWindowUndecorated | rl.FlagWindowHidden))
	rl.InitWindow(initW, initH, "Hello Gru")
	defer rl.CloseWindow()
	rl.SetExitKey(0)
	rl.SetWindowMinSize(int(ui.MinClientWidth), 320)

	ui.ApplyUIOptimizations()
	ui.RefreshDisplayScale()
	ui.InitDisplayAwareAtlases()
	defer ui.UnloadSDFFonts()
	defer ui.Icons.UnloadAll()

	windowW, windowH := initW, initH
	ui.InitSupersampling(windowW, windowH)
	defer ui.UnloadSupersampling()
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	defer ui.Toasts.Unload()

	doc := ui.NewDocument(windowW, windowH)
	doc.SetChromeTop(ui.TitleBarHeight)
	hello.Build(doc)
	ui.MountBorderlessDocument(doc, windowW, windowH)
	doc.EnableUIRenderTexture(true)

	shouldClose := false
	titleBar := ui.NewTitleBar("Hello Gru", ui.TitleBarStyleDark,
		func() { shouldClose = true },
		func() { rl.MinimizeWindow() },
		func() {
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
	rl.SetTargetFPS(60)
	ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)
	ui.ApplyNativeBorderlessRoundedCorners(true)

	fmt.Fprintln(os.Stderr, "hello: GRU_CORE spine")

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
		if nw != windowW || nh != windowH {
			windowW, windowH = nw, nh
			ui.SyncBorderlessClientSize(doc, titleBar, windowW, windowH)
		}

		titleBar.Update(windowW, windowH)
		ui.SyncBorderlessChromeFrame(titleBar)
		titleBar.ApplyResizeCursor(windowW, windowH)
		ui.SyncBorderlessInputFrame(titleBar, doc.Root, dt)

		doc.Root.Update(dt)
		if doc.Root.IsDirty() {
			doc.Root.Layout()
		}
		ui.Toasts.Update(dt)
		ui.Tooltips.Update(dt)

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

		if !windowShown {
			rl.ClearWindowState(uint32(rl.FlagWindowHidden))
			windowShown = true
		}
	}
}
