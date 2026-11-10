package ui

// Borderless app mount helpers — Studio-equivalent chrome without importing examples.
// Canonical contract: docs/GRU_CORE.md + docs/ARCHIVE/RESPONSIVE_RESIZE_LAG_2026-08-21.md

// SettleDocumentMount runs width-first layout passes after mount (FinishShellMount).
func SettleDocumentMount(doc *Document) {
	if doc == nil || doc.Root == nil {
		return
	}
	for pass := 0; pass < 5; pass++ {
		MarkResizeLayoutDirtySubtree(doc.Root)
		InvalidateAutoHeightTextMeasures(doc.Root)
		doc.Root.MarkDirty()
		doc.Root.Layout()
	}
}

// MountBorderlessDocument applies chrome inset + SyncBorderlessLayout + settle.
// Call once after Build / scene remount — NOT on every resize drag tick.
func MountBorderlessDocument(doc *Document, fullW, fullH int32) {
	WireBorderlessTitleBarOS()
	if doc == nil {
		return
	}
	if doc.ChromeTop() <= 0 {
		doc.SetChromeTop(TitleBarHeight)
	}
	doc.Resize(fullW, fullH)
	doc.SyncBorderlessLayout(fullW, fullH)
	SettleDocumentMount(doc)
	doc.InvalidatePaint()
}

// ResizeBorderlessDocument updates shell layout for a client size change.
// ForceFullLayout only when rem actually changes (S1_SHIP_FLOOR) — not every drag pixel.
func ResizeBorderlessDocument(doc *Document, fullW, fullH int32) {
	if doc == nil || fullW < 1 || fullH < 1 {
		return
	}
	prevRem := RootFontSize
	doc.Resize(fullW, fullH)
	if RootFontSize != prevRem {
		doc.ForceFullLayout()
	}
	doc.InvalidatePaint()
}

// SyncBorderlessClientSize mirrors Studio syncWindowFromDisplay for thin hosts:
// defer SSAA RT recreate mid-drag, type scale, overlays, document, title bar.
// Call whenever GetScreenWidth/Height changes — never stretch an old RT into a new client.
func SyncBorderlessClientSize(doc *Document, titleBar *TitleBar, fullW, fullH int32) {
	if fullW < 1 || fullH < 1 {
		return
	}
	// Live drag: defer GPU RT recreate (layout still runs). Do NOT call
	// ApplyDisplayAwareSupersampling every pixel — RescaleSupersampling forced
	// ResizeWindowTextures and undid defer (ARCHIVE 2026-08-21).
	DeferWindowTextures(fullW, fullH)
	prevDPI := DisplayScale
	RefreshDisplayScale()
	if DisplayScale != prevDPI {
		if ApplyDisplayAwareSupersampling(fullW, fullH) && doc != nil {
			doc.UnloadCache()
		}
	}
	SetTypeScaleResizeHold(false)
	prevRem := RootFontSize
	RefreshTypeScaleFromWindow(fullW, fullH)
	Toasts.SetWindowSize(fullW, fullH)
	Tooltips.SetWindowSize(fullW, fullH)
	if doc != nil {
		doc.Resize(fullW, fullH)
		if RootFontSize != prevRem {
			doc.ForceFullLayout()
		}
		doc.InvalidatePaint()
	}
	if titleBar != nil {
		titleBar.SetSize(fullW, fullH)
	}
}

// SettleBorderlessClientTextures applies any deferred RT recreate once chrome is idle.
// Call once per frame after Update when not edge-dragging (Studio settle flush).
func SettleBorderlessClientTextures(titleBar *TitleBar) {
	if titleBar != nil && titleBar.IsResizing() {
		return
	}
	if FillClientChromeResizing() {
		return
	}
	FlushPendingWindowTextures()
}

// SyncBorderlessClientIfChanged applies SyncBorderlessClientSize when the live
// GLFW client differs from fullW/fullH. Call after titleBar.Update and
// DrainFillClientChromeResize — those SetWindowSize in-frame.
//
// Mid-drag: defer RT + shell layout (2026-08-21). On settle call
// SettleBorderlessClientTextures. Never stretch the previous RT.
func SyncBorderlessClientIfChanged(doc *Document, titleBar *TitleBar, fullW, fullH int32) (w, h int32, changed bool) {
	w, h = ClientLayoutSize()
	if w < 1 {
		w = fullW
	}
	if h < 1 {
		h = fullH
	}
	if w == fullW && h == fullH {
		return w, h, false
	}
	SyncBorderlessClientSize(doc, titleBar, w, h)
	return w, h, true
}

// SyncBorderlessChromeFrame updates native + drawn rounded chrome from the title bar.
// Call once per frame while borderless (same as Studio main loop).
func SyncBorderlessChromeFrame(titleBar *TitleBar) {
	if titleBar == nil {
		return
	}
	BindFillClientChromeTitleBar(titleBar)
	rounded := titleBar.BorderlessRoundedChrome()
	ApplyNativeBorderlessRoundedCorners(rounded)
	SetBorderlessRoundedChrome(rounded)
}

// SyncBorderlessInputFrame wires chrome drag / wheel suppress flags after titleBar.Update.
func SyncBorderlessInputFrame(titleBar *TitleBar, root Node, dt float32) {
	_ = root
	_ = dt
	if titleBar == nil {
		return
	}
	SetChromeTitleBarDragging(titleBar.IsDragging() || titleBar.IsTitleClickPending())
	SetChromeWindowMoving(titleBar.IsDragging() || titleBar.IsTitleClickPending() || titleBar.IsResizing() || FillClientChromeResizing())
	SetWheelSuppressBandY(TitleBarHeight)
}
