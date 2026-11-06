// hello is the public Gru sample — GRU core spine then a tiny scene (docs/GRU_CORE.md).
//
//	go run ./cmd/hello
//
// Android: built as c-shared into packaging/android/hello (see docs/ANDROID_PHASE.md M1).
package main

import (
	"fmt"
	"os"
	"runtime"

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

func isAndroid() bool { return runtime.GOOS == "android" }

func main() {
	if !isAndroid() {
		runApp()
	}
}

func runApp() {
	rl.SetTraceLogLevel(rl.LogWarning)

	flags := uint32(rl.FlagVsyncHint)
	if !isAndroid() {
		flags |= uint32(rl.FlagMsaa4xHint | rl.FlagWindowResizable |
			rl.FlagWindowUndecorated | rl.FlagWindowHidden)
	}
	rl.SetConfigFlags(flags)

	startW, startH := initW, initH
	if isAndroid() {
		// 0×0 → raylib SetupFramebuffer uses full ANativeWindow (avoids 1920
		// letterbox on 2400-tall phones).
		startW, startH = 0, 0
	}
	rl.InitWindow(startW, startH, "Hello Gru")
	defer rl.CloseWindow()
	rl.SetExitKey(0)

	windowW, windowH := startW, startH
	if isAndroid() {
		ui.EnsureAndroidInsetFlags()
		windowW = int32(rl.GetScreenWidth())
		windowH = int32(rl.GetScreenHeight())
		if windowW > windowH {
			windowW, windowH = windowH, windowW
		}
		if windowW < 1 {
			windowW = 360
		}
		if windowH < 1 {
			windowH = 640
		}
	} else {
		rl.SetWindowMinSize(int(ui.MinClientWidth), 320)
	}

	ui.ApplyUIOptimizations()
	ui.RefreshDisplayScale()
	// Type scale before atlases so Android bitmap bake can track rem (M1 quality).
	ui.RefreshTypeScaleFromWindow(windowW, windowH)
	ui.InitDisplayAwareAtlases()
	defer ui.UnloadSDFFonts()
	defer ui.Icons.UnloadAll()

	useSSAA := !isAndroid()
	if useSSAA {
		ui.InitSupersampling(windowW, windowH)
		defer ui.UnloadSupersampling()
	}

	if isAndroid() {
		fmt.Fprintf(os.Stderr, "hello android: fontReady backend=%s path=%q size=%dx%d dpi=%.2f rem=%.0f atlas=%d\n",
			ui.UIFontBackend(), ui.LoadedUIFontPath, windowW, windowH, ui.DisplayScale, ui.RootFontSize,
			ui.EffectiveBitmapFontAtlasSize())
	}
	ui.Toasts.SetWindowSize(windowW, windowH)
	ui.Tooltips.SetWindowSize(windowW, windowH)
	defer ui.Toasts.Unload()

	doc := ui.NewDocument(windowW, windowH)
	useChrome := !isAndroid()
	if useChrome {
		doc.SetChromeTop(ui.TitleBarHeight)
	} else if isAndroid() {
		ui.ApplyAndroidSafeInsets(doc, 0, windowW, windowH)
	} else {
		doc.SetChromeTop(0)
	}
	hello.Build(doc)
	if useChrome {
		ui.MountBorderlessDocument(doc, windowW, windowH)
		doc.EnableUIRenderTexture(true)
	} else {
		// Critical on NativeActivity: Resize + settle so AutoHeight labels/buttons
		// get real bounds (Build alone is not enough after ANativeWindow sizing).
		ui.MountFullscreenDocument(doc, windowW, windowH)
	}

	shouldClose := false
	var titleBar *ui.TitleBar
	if useChrome {
		titleBar = ui.NewTitleBar("Hello Gru", ui.TitleBarStyleDark,
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
		ui.SetOverlayChromeInsets(ui.TitleBarHeight, 0)
		ui.ApplyNativeBorderlessRoundedCorners(true)
	} else {
		ui.SetOverlayChromeInsets(0, 0)
	}

	bg := ui.WindowChromeFill()
	borderless := useChrome
	windowShown := isAndroid() // NativeActivity is already visible
	rl.SetTargetFPS(60)
	settleFrames := 0

	fmt.Fprintln(os.Stderr, "hello: GRU_CORE spine")

	for !rl.WindowShouldClose() && !shouldClose {
		if isAndroid() && rl.IsKeyPressed(rl.KeyBack) {
			shouldClose = true
		}

		ui.PreparePointerInput()
		ui.SetActiveDocument(doc)
		dt := rl.GetFrameTime()
		doc.DrainQueue()
		bg = ui.WindowChromeFill()

		nw, nh := int32(rl.GetScreenWidth()), int32(rl.GetScreenHeight())
		if isAndroid() && nw > nh {
			nw, nh = nh, nw
		}
		if nw < 1 {
			nw = windowW
		}
		if nh < 1 {
			nh = windowH
		}
		if nw != windowW || nh != windowH {
			windowW, windowH = nw, nh
			if useChrome && titleBar != nil {
				ui.SyncBorderlessClientSize(doc, titleBar, windowW, windowH)
			} else {
				ui.MountFullscreenDocument(doc, windowW, windowH)
				ui.Toasts.SetWindowSize(windowW, windowH)
				ui.Tooltips.SetWindowSize(windowW, windowH)
				ui.RefreshTypeScaleFromWindow(windowW, windowH)
				settleFrames = 0
			}
		}

		// A few settle passes after Android surface ready / rotate.
		if isAndroid() && settleFrames < 5 {
			doc.ForceFullLayout()
			ui.SettleDocumentMount(doc)
			if settleFrames == 4 {
				if n := ui.FindNodeByID(doc.Root, "hello-btn"); n != nil {
					bb := n.Bounds()
					rl.TraceLog(rl.LogWarning, "hello settle btn=(%.0f,%.0f %.0fx%.0f)",
						bb.X, bb.Y, bb.Width, bb.Height)
				}
				if n := ui.FindNodeByID(doc.Root, "hello-card"); n != nil {
					cb := n.Bounds()
					rl.TraceLog(rl.LogWarning, "hello settle card=(%.0f,%.0f %.0fx%.0f)",
						cb.X, cb.Y, cb.Width, cb.Height)
				}
				if n := ui.FindNodeByID(doc.Root, "hello-title"); n != nil {
					tb := n.Bounds()
					rl.TraceLog(rl.LogWarning, "hello settle title=(%.0f,%.0f %.0fx%.0f)",
						tb.X, tb.Y, tb.Width, tb.Height)
				}
			}
			settleFrames++
		}

		if useChrome && titleBar != nil {
			titleBar.Update(windowW, windowH)
			if w, h, ok := ui.SyncBorderlessClientIfChanged(doc, titleBar, windowW, windowH); ok {
				windowW, windowH = w, h
			}
			ui.SyncBorderlessChromeFrame(titleBar)
			titleBar.ApplyResizeCursor(windowW, windowH)
			ui.SyncBorderlessInputFrame(titleBar, doc.Root, dt)
		} else {
			ui.PrepareWheelScroll(doc.Root)
		}

		doc.Root.Update(dt)
		if doc.Root.IsDirty() {
			doc.Root.Layout()
		}
		ui.Toasts.Update(dt)
		ui.Tooltips.Update(dt)

		if useSSAA {
			if (doc.NeedsRedraw() || ui.SuperTargetNeedsFill()) && ui.SuperTargetDrawable() {
				ui.BeginSuperFrame(bg, borderless, windowW, windowH)
				doc.Root.Draw()
				if titleBar != nil {
					titleBar.Draw()
				}
				ui.EndSuperFrame()
				ui.RecordCacheMiss()
			} else {
				ui.RecordCacheHit()
			}
		}

		rl.BeginDrawing()
		if useSSAA {
			ui.BeginDrawingBorderless(bg, borderless, windowW, windowH)
			if ui.SuperTargetDrawable() {
				ui.BlitToScreenBorderless(windowW, windowH, borderless)
			}
		} else {
			rl.ClearBackground(bg)
			doc.Root.Draw()
		}
		ui.Tooltips.Draw()
		ui.Toasts.Draw()
		rl.EndDrawing()
		ui.SettleBorderlessClientTextures(titleBar)

		if !windowShown {
			rl.ClearWindowState(uint32(rl.FlagWindowHidden))
			windowShown = true
		}
	}

	if isAndroid() {
		os.Exit(0)
	}
}
