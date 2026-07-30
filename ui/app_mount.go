package ui

// Borderless app mount helpers — Studio-equivalent chrome without importing examples.
// Canonical contract: docs/GRU_CORE.md

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

// ResizeBorderlessDocument updates layout for a client size change (Studio resize path).
// Does not re-run mount settle passes — those belong in MountBorderlessDocument only.
func ResizeBorderlessDocument(doc *Document, fullW, fullH int32) {
	if doc == nil || fullW < 1 || fullH < 1 {
		return
	}
	doc.Resize(fullW, fullH)
	doc.ForceFullLayout()
	doc.InvalidatePaint()
}

// SyncBorderlessClientSize mirrors Studio syncWindowFromDisplay for thin hosts:
// SSAA RT resize + type scale + overlays + document + title bar.
// Call whenever GetScreenWidth/Height changes — never blit an old RT to a new client.
func SyncBorderlessClientSize(doc *Document, titleBar *TitleBar, fullW, fullH int32) {
	if fullW < 1 || fullH < 1 {
		return
	}
	// ResizeWindowTextures is also reached via ApplyDisplayAwareSupersampling →
	// RescaleSupersampling; call explicitly so hosts match Studio and stay correct
	// even if SSAA helpers change.
	ResizeWindowTextures(fullW, fullH)
	if ApplyDisplayAwareSupersampling(fullW, fullH) && doc != nil {
		doc.UnloadCache()
	}
	RefreshTypeScaleFromWindow(fullW, fullH)
	Toasts.SetWindowSize(fullW, fullH)
	Tooltips.SetWindowSize(fullW, fullH)
	ResizeBorderlessDocument(doc, fullW, fullH)
	if titleBar != nil {
		titleBar.SetSize(fullW, fullH)
	}
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

// SyncBorderlessInputFrame mirrors Studio's post-titleBar input routing for thin hosts:
// chrome drag flags, title-band wheel suppress, PrepareWheelScroll, and ListTile
// switch-row pointer pass (must run before doc.Root.Update).
// Call after titleBar.Update and before doc.Root.Update.
func SyncBorderlessInputFrame(titleBar *TitleBar, root Node, dt float32) {
	if titleBar != nil {
		// Match Studio: defer WebView raise for drag AND title-click pending (§13.5).
		SetChromeTitleBarDragging(titleBar.IsDragging() || titleBar.IsTitleClickPending())
		SetChromeWindowMoving(titleBar.IsDragging() || titleBar.IsTitleClickPending() || titleBar.IsResizing() || FillClientChromeResizing())
		SetWheelSuppressBandY(TitleBarHeight)
	} else {
		SetChromeTitleBarDragging(false)
		SetChromeWindowMoving(false)
		SetWheelSuppressBandY(0)
	}
	PrepareWheelScroll(root)
	ProcessSwitchListTilePointers(root, dt)
}
