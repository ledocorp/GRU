// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	rl "github.com/gen2brain/raylib-go/raylib"
)

// Document is the top-level owner of the UI tree.
//
// # LLM Prompt Template — main loop + idle cache (SSAA)
//
//	doc := ui.NewDocument(1280, 720)
//	doc.Root = ui.NewContainer("root", 0, 0, 1280, 720)
//	// build scene under doc.Root …
//	for !rl.WindowShouldClose() {
//	    dt := rl.GetFrameTime()
//	    doc.DrainQueue()
//	    doc.Root.Update(dt)
//	    if doc.Root.IsDirty() { doc.Root.Layout() }
//	    if doc.NeedsRedraw() {
//	        ui.BeginSuperFrame(bg)
//	        doc.Root.Draw()
//	        ui.EndSuperFrame()
//	    }
//	    rl.BeginDrawing()
//	    ui.BlitToScreen(int32(w), int32(h))
//	    rl.EndDrawing()
//	}
//
// # LLM Prompt Template
//
//	doc := ui.NewDocument(1280, 720)
//	doc.Root = ui.NewContainer("root", 0, 0, 1280, 720)
//	doc.Root.LayoutType = ui.LayoutFlex
//	doc.Root.FlexDirection = ui.FlexColumn
//	// each frame:
//	doc.DrainQueue()
//	doc.Root.Update(dt)
//	if doc.Root.IsDirty() { doc.Root.Layout() }
//	doc.Root.Draw()
//
// It holds the root Container (which fills the entire window), the window
// dimensions, and the currently focused Node. Every frame the application
// should call:
//
//	doc.DrainQueue()          // flush any goroutine-queued callbacks first
//	doc.Root.Update(dt)
//	if doc.Root.IsDirty() { doc.Root.Layout() }
//	doc.Root.Draw()
//
// # Document-level UI Cache
//
// When enabled via EnableUIRenderTexture(true), the Document keeps a GPU
// texture that holds the last rendered frame of the widget tree. On frames
// where the root is not dirty, the draw pass is skipped entirely and the
// cached texture is blitted instead. This is a significant win for static or
// near-static UIs.
//
// In SSAA mode, the supersampling target (superTarget) already acts as an
// implicit GPU cache between frames — BeginSuperFrame only clears and redraws
// it when NeedsRedraw() returns true. The Document's own uiRT is only
// allocated in 1× fallback mode (SupersamplingActive() == false), where no
// such persistent texture already exists.
//
// Typical usage in main.go (SSAA path):
//
//	if doc.NeedsRedraw() {
//	    ui.BeginSuperFrame(bgColor)
//	    doc.Root.Draw()
//	    ui.EndSuperFrame()
//	    ui.RecordCacheMiss()
//	} else {
//	    ui.RecordCacheHit()
//	}
//	rl.BeginDrawing()
//	ui.BlitToScreen(windowW, windowH)
//	rl.EndDrawing()
//
// # Concurrency / Main-Thread Callbacks
//
// Background goroutines must not call raylib or mutate widget state directly.
// Use QueueMain to post a callback that will be executed on the main goroutine
// at the start of the next frame, safely before Update/Layout/Draw.
//
// Example (async image or data load):
//
//	go func() {
//	    data, err := fetchFromNetwork(url)
//	    if err != nil { return }
//	    doc.QueueMain(func() {
//	        label.Text.Set(data.Title)    // safe: runs on main thread
//	        imageWidget.SetTexture(data.Texture)
//	        doc.Root.MarkDirty()
//	    })
//	}()
//
// Focus management: call doc.SetFocus(node) to direct keyboard events to a
// specific widget. SetFocus emits EventBlur on the previously focused node
// and EventFocus on the new one. Passing nil clears focus.
//
// SetActiveDocument is called each frame from the main loop so overlay widgets
// (modal TextInput, etc.) can sync keyboard handling with doc.Focused.
var activeDocument *Document

// SetActiveDocument records the document for the current frame.
func SetActiveDocument(d *Document) { activeDocument = d }

// ActiveDocument returns the document set for the current frame, or nil.
func ActiveDocument() *Document { return activeDocument }

// SetPlatformWindowHandle stores the native window handle for embedded hosts (WebView2).
func (d *Document) SetPlatformWindowHandle(hwnd uintptr) {
	if d == nil {
		return
	}
	d.platformWindowHandle = hwnd
}

// PlatformWindowHandle returns the native window handle set by the platform loop.
func (d *Document) PlatformWindowHandle() uintptr {
	if d == nil {
		return 0
	}
	return d.platformWindowHandle
}

type Document struct {
	Root    *Container // Root container
	Width   int32      // Content width (matches client width)
	Height  int32      // Content height (client height minus chromeTop)
	Focused Node       // Currently focused node

	// chromeTop is the logical Y inset for the widget tree (borderless title bar).
	chromeTop float32
	// chromeBottom reserves the launcher nav strip (main.go) so scene content and
	// hit targets do not sit under the footer Directory control.
	chromeBottom float32

	uiCacheEnabled bool               // true when UI-texture caching is on
	uiRT           rl.RenderTexture2D // 1× cache used in non-SSAA mode

	// taskQueue holds callbacks queued from background goroutines via QueueMain.
	// Buffered at 128 slots — enough for any realistic burst of async results.
	// DrainQueue empties it at the start of each frame on the main goroutine.
	taskQueue chan func()

	// platformWindowHandle is the native window handle (HWND on Windows) for WebView2 child surfaces.
	platformWindowHandle uintptr
}

// NewDocument creates a new Document with a root container matching the window size.
// The root uses LayoutAbsolute so direct children keep explicit bounds from Build()
// and from Document.Resize (flex on the root would overwrite X/Y and leave fixed-width
// children narrower than the window after resize).
func NewDocument(width, height int32) *Document {
	root := NewContainer("root", 0, 0, float32(width), float32(height))
	root.LayoutType = LayoutAbsolute
	return &Document{
		Root:      root,
		Width:     width,
		Height:    height,
		taskQueue: make(chan func(), 128),
	}
}

// ChromeTop returns the logical Y inset for scene content (borderless title bar).
func (d *Document) ChromeTop() float32 {
	if d == nil {
		return 0
	}
	return d.chromeTop
}

// SetChromeTop sets the top inset for the root container (e.g. ui.TitleBarHeight
// in borderless mode). Call before Resize when toggling undecorated chrome.
func (d *Document) SetChromeTop(top float32) {
	if d == nil {
		return
	}
	if d.chromeTop == top {
		return
	}
	d.chromeTop = top
	if d.Root != nil {
		d.Root.MarkDirty()
	}
}

// SetChromeBottom reserves space at the bottom of the client area (launcher nav bar).
func (d *Document) SetChromeBottom(bottom float32) {
	if d == nil {
		return
	}
	if d.chromeBottom == bottom {
		return
	}
	d.chromeBottom = bottom
	if d.Root != nil {
		d.Root.MarkDirty()
	}
}

// contentHeight returns client height minus chromeTop and chromeBottom, at least 1.
func (d *Document) contentHeight(fullH int32) int32 {
	h := fullH - int32(d.chromeTop) - int32(d.chromeBottom)
	if h < 1 {
		return 1
	}
	return h
}

// syncRootBounds sets Root to the content band below chromeTop.
func (d *Document) syncRootBounds(w, fullH int32) {
	if d == nil || d.Root == nil {
		return
	}
	h := d.contentHeight(fullH)
	d.Root.SetBounds(rl.NewRectangle(0, d.chromeTop, float32(w), float32(h)))
}

// fitRootChildrenToContent snaps each direct child of Root to the content
// band below chromeTop. Bounds are screen-absolute (widgets draw at bounds.X/Y).
func (d *Document) fitRootChildrenToContent() {
	if d == nil || d.Root == nil {
		return
	}
	r := rl.NewRectangle(0, d.chromeTop, float32(d.Width), float32(d.Height))
	for _, ch := range d.Root.children {
		ch.SetBounds(r)
	}
}

// SyncBorderlessLayout applies chromeTop, root bounds, shell flex, and a full
// layout pass. Call after toggling F9 or when the client size changes in
// borderless mode.
func (d *Document) SyncBorderlessLayout(fullW, fullH int32) {
	if d == nil || d.Root == nil {
		return
	}
	contentH := d.contentHeight(fullH)
	d.Width = fullW
	d.Height = contentH
	d.syncRootBounds(fullW, fullH)
	d.fitRootChildrenToContent()
	d.applyShellFlexAndSyncRootLayout()
}

// ─── Responsive Breakpoints ───────────────────────────────────────────────────

// Breakpoint describes the current responsive layout tier based on pixel width.
type Breakpoint int

const (
	BreakpointXS Breakpoint = iota // < 480 px  (phone portrait)
	BreakpointSM                   // 480–767 px (phone landscape / small tablet)
	BreakpointMD                   // 768–1023 px (tablet)
	BreakpointLG                   // 1024–1279 px (desktop)
	BreakpointXL                   // ≥ 1280 px  (wide desktop)
)

// String returns a human-readable label for the breakpoint.
func (b Breakpoint) String() string {
	switch b {
	case BreakpointXS:
		return "xs  (<480px)"
	case BreakpointSM:
		return "sm  (480–767px)"
	case BreakpointMD:
		return "md  (768–1023px)"
	case BreakpointLG:
		return "lg  (1024–1279px)"
	default:
		return "xl  (≥1280px)"
	}
}

// CurrentBreakpoint maps a pixel width to the corresponding Breakpoint tier.
// Pass any width — window width, container width, or a simulated value.
func CurrentBreakpoint(width float32) Breakpoint {
	switch {
	case width < 480:
		return BreakpointXS
	case width < 768:
		return BreakpointSM
	case width < 1024:
		return BreakpointMD
	case width < 1280:
		return BreakpointLG
	default:
		return BreakpointXL
	}
}

// ActiveBreakpoint is the package-level reactive breakpoint Signal.
// The demo launcher sets it when the window width changes (and once after the
// first scene build). Demos that read this signal can reflow by tier; others
// can ignore it and rely on Document bounds alone.
var ActiveBreakpoint = NewSignal(BreakpointXL)

// ApplyUIOptimizations disables 3D/game-specific OpenGL state that is
// unnecessary for a 2D UI application. Call once after rl.InitWindow, before
// the main loop.
//
// What is called and why:
//   - DisableDepthTest: no depth buffer needed for flat 2D; eliminates a
//     per-fragment depth comparison for every draw call.
//   - DisableBackfaceCulling: raylib's 2D quads are always front-facing;
//     the culling check is unnecessary work.
//   - SetGesturesEnabled(0): gesture recognition is a touch/mobile feature;
//     disabling it removes the per-frame OS gesture-tracking work on desktop.
//
// Audio: Gru intentionally does NOT call rl.InitAudioDevice(). Doing so
// would open a system audio context and start an audio mixer thread even
// when no sound is ever played. If a Gru application needs audio in the
// future, InitAudioDevice should be called lazily on first use.
func ApplyUIOptimizations() {
	rl.DisableDepthTest()
	rl.DisableBackfaceCulling()
	rl.SetGesturesEnabled(0)
}

// Add appends a node as a direct child of the absolute root. Demos that need
// a resize-safe page shell typically mount a full-window flex column + main
// Viewport first (see the examples package, page_shell.go), then add content
// to that Viewport instead of calling Add with a bare Viewport.
func (d *Document) Add(node Node) {
	d.Root.AddChild(node)
}

// SetFocus sets the currently focused node.
func (d *Document) SetFocus(node Node) {
	if d.Focused == node {
		return
	}
	if d.Focused != nil {
		d.Focused.Emit(EventBlur, nil)
	}
	d.Focused = node
	if node != nil {
		ReleaseWebKeyboardFocus()
		node.Emit(EventFocus, nil)
	}
}

// EnableUIRenderTexture turns the Document-level GPU cache on or off.
//
// When on and supersampling is inactive, a 1× RenderTexture2D is allocated and
// used as the cache target. In SSAA mode, the existing superTarget already
// provides frame-level caching; no extra texture is needed there.
//
// Default is false (cache disabled) so that existing code is unaffected.
func (d *Document) EnableUIRenderTexture(on bool) {
	if on == d.uiCacheEnabled {
		return
	}
	d.uiCacheEnabled = on
	d.RefreshRenderCache()
}

// RefreshRenderCache reconciles the document cache with the current render
// mode. SSAA uses the global superTarget as the cache; native 1x mode needs the
// document-owned uiRT. Call after changing supersampling scale.
func (d *Document) RefreshRenderCache() {
	if d == nil {
		return
	}
	if d.uiCacheEnabled && !SupersamplingActive() {
		if d.uiRT.ID != 0 {
			return
		}
		w, h := clampDim(d.Width, d.Height)
		d.uiRT = rl.LoadRenderTexture(w, h)
		rl.SetTextureFilter(d.uiRT.Texture, rl.FilterBilinear)
	} else if d.uiRT.ID != 0 {
		rl.UnloadRenderTexture(d.uiRT)
		d.uiRT = rl.RenderTexture2D{}
	}
}

// IsUICacheEnabled returns true when the UI-texture cache has been requested.
func (d *Document) IsUICacheEnabled() bool { return d.uiCacheEnabled }

// NeedsRedraw returns true when the GPU texture cache is stale and a full
// widget draw pass must run this frame.
//
// The check is dirty-flag–based rather than uiCacheEnabled-based:
//
//   - layoutDirty (IsDirty): a widget changed size or position — full redraw.
//   - drawDirty (DbgDrawDirty): a purely visual change (hover, blink, color) —
//     redraw still needed even though no layout ran.
//
// In SSAA mode the superTarget acts as the implicit frame cache: skipping
// BeginSuperFrame on clean frames means zero GPU work for static UI — the
// blit from the previous frame's superTarget is already on screen.
//
// In 1× mode, the cache is the uiRT RenderTexture2D (allocated only when
// uiCacheEnabled == true). Without a cache texture, draw always runs.
//
// Design invariant: Layout() only clears layoutDirty. Draw() only clears
// drawDirty. So checking both after the Layout pass catches all change types.
func (d *Document) NeedsRedraw() bool {
	// No cache available in 1× mode without EnableUIRenderTexture → always draw.
	if !d.uiCacheEnabled && !SupersamplingActive() {
		return true
	}
	if d.uiCacheEnabled && !SupersamplingActive() && d.uiRT.ID == 0 {
		return true
	}
	// Redraw whenever any widget has a pending visual or layout change.
	return SubtreeNeedsRedraw(d.Root)
}

// UIRenderTexture returns the 1× fallback cache texture.
// Returns a zero-value RenderTexture2D if SSAA is active (superTarget is used
// instead) or if EnableUIRenderTexture has not been called with true.
func (d *Document) UIRenderTexture() rl.RenderTexture2D { return d.uiRT }

// UnloadCache releases the Document's uiRT GPU texture.
// Call this when a scene is unloaded or on application shutdown to avoid GPU
// resource leaks. Safe to call even if caching is disabled.
func (d *Document) UnloadCache() {
	if d.uiRT.ID != 0 {
		rl.UnloadRenderTexture(d.uiRT)
		d.uiRT = rl.RenderTexture2D{}
	}
}

// InvalidatePaint marks the tree dirty for a full repaint into superTarget or uiRT.
//
// Call after scene swaps — blitting a valid cache texture from the previous scene
// looks frozen until something (e.g. resize) forces NeedsRedraw.
//
// LLM Prompt Template:
//   "After loadScene or NavigateToScene, call doc.InvalidatePaint() — never
//    ClearDrawDirtySubtree on a fresh tree. See docs/IDLE_INVARIANTS.md §3."
func (d *Document) InvalidatePaint() {
	if d == nil || d.Root == nil {
		return
	}
	d.Root.MarkDrawDirty()
	if d.uiCacheEnabled && !SupersamplingActive() {
		d.UnloadCache()
	}
}

// ForceFullLayout marks the entire widget tree layout- and draw-dirty, prepares
// every Viewport for a full layoutFlex pass on the next Layout, walks
// Node.Children() via MarkResizeLayoutDirtySubtree, and drops the UI-texture
// cache. Resize delegates here; call directly if the window size is unchanged but
// everything must reflow.
func (d *Document) ForceFullLayout() {
	if d == nil || d.Root == nil {
		return
	}
	MarkResizeLayoutDirtySubtree(d.Root)
	InvalidateAutoHeightTextMeasures(d.Root)
	// SSAA uses superTarget (resized via ResizeWindowTextures); unloading the
	// optional 1× document cache is redundant and added cost on every resize frame.
	if !SupersamplingActive() {
		d.UnloadCache()
	}
}

// RelayoutTreeForResize prepares viewports and marks the full tree layout-dirty
// without unloading the document UI cache. Use when logical width/height are
// unchanged (e.g. repeated IsWindowResized at the minimum) so GPU textures stay
// valid while flex geometry is recomputed.
func (d *Document) RelayoutTreeForResize() {
	if d == nil || d.Root == nil {
		return
	}
	MarkResizeLayoutDirtySubtree(d.Root)
	InvalidateAutoHeightTextMeasures(d.Root)
}

// relayoutShellsForResize invalidates each root page shell's layout cache and marks
// only the shell containers dirty. Flex/grid children are marked dirty by
// Container.Layout when the shell's bounds change (boundsChanged +
// flexChildDependsOnParentFlex). Walking the full subtree here forced layout and
// draw on every leaf during window resize (e.g. Responsive demo ~8 redraws/s).
func (d *Document) relayoutShellsForResize() {
	if d == nil || d.Root == nil {
		return
	}
	for _, ch := range d.Root.children {
		if c, ok := ch.(*Container); ok {
			c.InvalidateLayoutPassCache()
			c.MarkDirty()
		}
	}
	d.Root.MarkDirty()
}

// applyShellFlexAndSyncRootLayout runs layoutFlex on every direct flex root child
// (page-shell, edge-to-edge app shells, etc.), syncs MountFlexPageShell viewports,
// then Root.Layout.
func (d *Document) applyShellFlexAndSyncRootLayout() {
	if d == nil || d.Root == nil {
		return
	}
	for _, ch := range d.Root.children {
		shell, ok := ch.(*Container)
		if !ok || shell.LayoutType != LayoutFlex {
			continue
		}
		shell.InvalidateLayoutPassCache()
		shell.MarkDirty()
	}
	d.Root.MarkDirty()
	d.Root.Layout()
	assertPinnedFlexRowWidths(d.Root)
	// Page-shell: sync the main viewport to the shell inner rect after flex layout.
	for _, ch := range d.Root.children {
		shell, ok := ch.(*Container)
		if !ok || shell.LayoutType != LayoutFlex {
			continue
		}
		mainVP := findShellScrollViewport(shell)
		if mainVP != nil && shellViewportDirectChild(shell, mainVP) {
			if SyncShellScrollViewportWidth(shell, mainVP) {
				mainVP.Layout()
			}
		}
	}
	d.Root.MarkDrawDirty()
}

// Resize updates the document's logical dimensions to match a window resize.
// It calls Root.SetBounds, propagates the size delta to all direct children,
// Root.MarkDirty(), then applyShellFlexAndSyncRootLayout so the page shell's
// layoutFlex runs and pushes new bounds into flex-grow / w=0 children (see Container.Layout).
// Call this whenever rl.GetScreenWidth() / rl.GetScreenHeight() changes.
// Resize updates layout for client size fullW×fullH (full window). Content
// height is fullH minus SetChromeTop (borderless title bar).
func (d *Document) Resize(fullW, fullH int32) {
	if fullW < 1 || fullH < 1 {
		return
	}
	if fullW < MinClientWidth {
		fullW = MinClientWidth
	}
	contentH := d.contentHeight(fullH)
	// Same logical size can still follow a platform resize event (e.g. after
	// sync paths). Do not UnloadCache here — hammering the UI RT every frame at
	// the minimum width breaks redraw; RelayoutTreeForResize + shell flex is enough.
	if d.Width == fullW && d.Height == contentH {
		d.syncRootBounds(fullW, fullH)
		d.fitRootChildrenToContent()
		d.relayoutShellsForResize()
		d.applyShellFlexAndSyncRootLayout()
		return
	}
	d.Width = fullW
	d.Height = contentH
	d.syncRootBounds(fullW, fullH)
	d.fitRootChildrenToContent()
	d.relayoutShellsForResize()
	d.applyShellFlexAndSyncRootLayout()
}

// findShellScrollViewport returns the primary scroll viewport for a root shell:
// MountFlexPageShell (vp direct child) or MountDesktopPageShell (vp inside main column).
// assertPinnedFlexRowWidths re-applies pinned widths after shell layout (e.g.
// NavigationRail at 80px) so resize passes cannot leave flex-grow siblings overlapping.
func assertPinnedFlexRowWidths(root Node) {
	var walk func(Node)
	walk = func(n Node) {
		if n == nil || n.IsHidden() {
			return
		}
		if row, ok := n.(*Container); ok && row.LayoutType == LayoutFlex && row.FlexDirection == FlexRow {
			for _, ch := range row.children {
				if pw, ok := PinnedMainAxisWidth(ch); ok {
					b := ch.Bounds()
					if b.Width != pw {
						b.Width = pw
						ch.SetBounds(b)
					}
				}
			}
		}
		for _, ch := range n.Children() {
			walk(ch)
		}
	}
	walk(root)
}

// ShellScrollViewportBand returns the horizontal band (x, width) for the main
// page-scroll viewport in a page-shell flex column: left inset by shell padding,
// flush to the window right edge for the scrollbar gutter.
func ShellScrollViewportBand(shell *Container) (x, width float32) {
	if shell == nil {
		return 0, 0
	}
	b := shell.Bounds()
	padL, _, padR, _ := flexContentInsets(shell)
	w := b.Width - padL - padR
	if w < 1 {
		w = 1
	}
	return b.X + padL, w
}

// SyncShellScrollViewportWidth assigns the flush-right horizontal band to a
// page-scroll viewport inside a page-shell (app-shell scenes only).
func SyncShellScrollViewportWidth(shell *Container, vp *Viewport) bool {
	if shell == nil || vp == nil || shell.styleName != "page-shell" || vp.styleName != "page-scroll" {
		return false
	}
	wantX, wantW := ShellScrollViewportBand(shell)
	b := vp.Bounds()
	if b.X == wantX && b.Width == wantW {
		return false
	}
	b.X = wantX
	b.Width = wantW
	vp.setBoundsNoMark(b)
	vp.MarkDirty()
	return true
}

// ShellMainViewportBounds returns the full inner rectangle of a page-shell when
// it has a single scroll child (bootstrap / tests). Multi-child shells (sticky
// header + body) rely on flex layout for Y/Height; use ShellScrollViewportBand
// for horizontal sync on resize.
func ShellMainViewportBounds(shell *Container) rl.Rectangle {
	b := shell.Bounds()
	_, padT, _, padB := flexContentInsets(shell)
	x, w := ShellScrollViewportBand(shell)
	inner := rl.NewRectangle(x, b.Y+padT, w, b.Height-padT-padB)
	if inner.Height < 1 {
		inner.Height = 1
	}
	return inner
}

func shellViewportDirectChild(shell *Container, vp *Viewport) bool {
	for _, ch := range shell.children {
		if ch == vp {
			return true
		}
	}
	return false
}

func findShellScrollViewport(shell *Container) *Viewport {
	for _, ch := range shell.children {
		if v, ok := ch.(*Viewport); ok {
			return v
		}
	}
	for _, ch := range shell.children {
		if col, ok := ch.(*Container); ok {
			if v := findViewportInContainer(col); v != nil {
				return v
			}
		}
	}
	return nil
}

func findViewportInContainer(c *Container) *Viewport {
	for _, ch := range c.children {
		if v, ok := ch.(*Viewport); ok {
			return v
		}
		if sub, ok := ch.(*Container); ok {
			if v := findViewportInContainer(sub); v != nil {
				return v
			}
		}
	}
	return nil
}

// ─── Concurrency: main-thread callback queue ──────────────────────────────────

// QueueMain safely schedules fn to run on the main goroutine at the start of
// the next frame, inside DrainQueue(). It is safe to call from any goroutine.
//
// Use this whenever a background goroutine (image load, network fetch, file
// read) needs to update widget state or call raylib APIs, both of which must
// happen on the main thread.
//
// If the internal 128-slot buffer is full (extremely rare), the callback is
// dropped and a warning is emitted to the raylib log. In practice this cannot
// happen unless hundreds of goroutines fire simultaneously without any frame
// having run in between.
func (d *Document) QueueMain(fn func()) {
	select {
	case d.taskQueue <- fn:
		Wake(WakeDataUpdate, "queue-main")
	default:
		rl.TraceLog(rl.LogWarning, "Gru: QueueMain buffer full — callback dropped")
	}
}

// DrainQueue runs all callbacks that have been queued via QueueMain since the
// last call to DrainQueue. Call this once at the very start of each frame,
// before doc.Root.Update(dt), so that goroutine results are visible to the
// Update → Layout → Draw pipeline in the same frame they arrive.
//
// DrainQueue is a no-op when no callbacks are pending — the select default
// branch returns immediately with zero allocations.
func (d *Document) DrainQueue() {
	d.DrainQueueCount()
}

// DrainQueueCount is like DrainQueue but returns how many callbacks ran.
func (d *Document) DrainQueueCount() int {
	count := 0
	for {
		select {
		case fn := <-d.taskQueue:
			fn()
			count++
		default:
			return count
		}
	}
}

// TaskQueueLen returns the number of callbacks currently waiting in the
// main-thread task queue. Safe to call from any goroutine (len on a buffered
// channel is atomic on all Go platforms). Useful for monitoring async work
// in the Inspector or perf overlay.
func (d *Document) TaskQueueLen() int { return len(d.taskQueue) }
