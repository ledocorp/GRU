// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// TitleBarHeight is the pixel height of the custom title bar.
const TitleBarHeight = float32(48)

// TitleBarShadowGap is the band below the title bar drop shadow where scene
// chrome should not paint an opaque AppBar (see examples/shell_desktop_demo.go).
const TitleBarShadowGap = float32(6)

// titleChromeRadius is the corner radius for borderless window chrome (all four corners).
const titleChromeRadius = float32(10)

// titleFontSize is the base title label size in the title bar. Actual render
// size also includes GlobalFontScale.
const titleFontSize = float32(20)

// TitleBarAppIconSize is the system window icon drawn at the top-left when ShowAppIcon is set.
const TitleBarAppIconSize = float32(24)

// Resize-grip hit zones.
const (
	sideGrip   = float32(6)  // E / W / S edge strip width
	topGrip    = float32(4)  // N edge — thin so it doesn't eat the title-drag zone
	cornerGrip = float32(16) // corner square — large enough to grab comfortably
)

// Minimum window dimensions enforced during borderless edge resize (matches
// main.go client clamp — ui.MinClientWidth).
var (
	minWinW = int(MinClientWidth)
	minWinH = 320
)

// titleBarNativeDblClickMaximize is wired from main on Windows (ShowWindow / NC
// messages). When nil, title-bar double-click uses OnMaximize.
var titleBarNativeDblClickMaximize func()

// titleBarDoubleClickInterval is the second-click window for maximize (OS default on Windows).
var titleBarDoubleClickInterval = 0.35

// SetTitleBarNativeDoubleClickMaximize registers OS-native maximize/restore for
// title-bar double-click. Minimize/maximize/close buttons still use OnMinimize etc.
func SetTitleBarNativeDoubleClickMaximize(fn func()) {
	titleBarNativeDblClickMaximize = fn
}

// SetTitleBarDoubleClickInterval sets the double-click timing window in seconds.
func SetTitleBarDoubleClickInterval(seconds float64) {
	if seconds > 0 {
		titleBarDoubleClickInterval = seconds
	}
}

// Windows-style title-bar button widths (right-aligned).
const (
	closeBtnW = float32(48)
	otherBtnW = float32(46)
)

// TitleBarStyle selects the visual theme.
type TitleBarStyle int

const (
	TitleBarStyleDark   TitleBarStyle = iota // dark bg, light text
	TitleBarStyleLight                       // light bg, dark text
	TitleBarStyleAccent                      // indigo bg
)

// resizeEdge identifies which edge or corner is being dragged.
type resizeEdge int

const (
	edgeNone resizeEdge = iota
	edgeN
	edgeS
	edgeE
	edgeW
	edgeNE
	edgeNW
	edgeSE
	edgeSW
)

// TitleBar renders and handles the custom window chrome.
//
// Key design rules:
//   - Hover is reset to false at the top of every Update() and recomputed from
//     scratch, so stale values from mid-resize can never persist into Draw().
//   - Button zones always take priority over resize grip zones.
//   - applyResize intentionally does NOT write back to tb.windowW/H.  Those
//     values always come from Update()'s parameter which matches the live SSAA
//     texture size.  Writing a "future" size here would push the buttons past
//     the texture boundary and clip them off-screen for one frame.
//
// Window-position tracking: GetWindowPosition() on Windows returns a value
// that includes the invisible DWM shadow/frame and is therefore offset from
// the coordinate space that SetWindowPosition uses. To avoid a jump on every
// drag/resize, we maintain trackedX/trackedY — our own ground-truth record of
// where the window is. It is seeded from GetWindowPosition() once (when we
// first go borderless) and thereafter updated only by our own SetWindowPosition
// calls via moveWindow().
// TitleBarLeadingIcon draws the optional top-left app icon (e.g. GRU lettermark).
type TitleBarLeadingIcon func(rect rl.Rectangle)

type TitleBar struct {
	Title        string
	ShowTitle    bool // when false, centered title text is omitted
	ShowAppIcon  bool // small icon at top-left (uses DrawLeadingIcon)
	DrawLeadingIcon TitleBarLeadingIcon
	Style        TitleBarStyle
	HandleResize bool // enable edge/corner resize grips (borderless mode)
	OnClose      func()
	OnMinimize   func()
	OnMaximize   func()

	// drag-to-move
	dragging    bool
	dragStartMX float32 // screen-absolute mouse X when drag began
	dragStartMY float32
	dragStartWX int32 // window X when drag began
	dragStartWY int32

	// resize
	resizing      bool
	resizeEdge    resizeEdge
	resizeStartMX float32
	resizeStartMY float32
	resizeStartWX int32
	resizeStartWY int32
	resizeStartWW int32
	resizeStartWH int32

	// trackedX/Y: our own record of window position — never read from
	// GetWindowPosition() after the initial seed because that API returns
	// a value that includes the invisible DWM frame offset on Windows.
	trackedX  int32
	trackedY  int32
	posSeeded bool // true once trackedX/Y have been seeded

	// double-click timing
	lastClickTime      float64
	titleClickPending  bool // Windows: wait for 2nd click or move before drag
	titleClickStartMX  float32
	titleClickStartMY  float32

	// hover — reset every Update(), never stale
	hoverClose bool
	hoverMin   bool
	hoverMax   bool
	closeArmed bool // close fires on mouse-up while still over close (avoids slip-off accidents)

	// chromeCursorActive tracks whether ApplyResizeCursor set an edge/corner cursor
	// so we can reset to default when the pointer leaves the OS resize band.
	chromeCursorActive bool

	// current logical window size — set by Update(), used by Draw()
	windowW int32
	windowH int32

	// saved bounds from the last non-maximized frame (restore-on-drag target).
	savedWX, savedWY, savedWW, savedWH int32
	hasSavedBounds bool
}

// IsResizing reports an active borderless edge/corner resize drag.
func (tb *TitleBar) IsResizing() bool { return tb.resizing }

// IsDragging reports an active title-bar move drag.
func (tb *TitleBar) IsDragging() bool { return tb.dragging }

// IsTitleClickPending reports the window between first and second title-bar click
// (double-click maximize or drag-out). WebView present defers during this interval.
func (tb *TitleBar) IsTitleClickPending() bool { return tb.titleClickPending }

// BorderlessRoundedChrome reports whether borderless fill/title chrome should
// use rounded corners. True while dragging (including restore-out-of-maximize).
func (tb *TitleBar) BorderlessRoundedChrome() bool {
	if tb.dragging || tb.titleClickPending {
		return true
	}
	return !rl.IsWindowMaximized()
}

// NewTitleBar constructs a TitleBar. Pass nil for any callback to disable
// the corresponding button.
func NewTitleBar(title string, style TitleBarStyle, onClose, onMin, onMax func()) *TitleBar {
	return &TitleBar{
		Title:     title,
		ShowTitle: true,
		Style:      style,
		OnClose:    onClose,
		OnMinimize: onMin,
		OnMaximize: onMax,
	}
}

// moveWindow moves the window to (x, y) in screen coordinates and updates
// our tracked position. Always use this instead of calling SetWindowPosition
// directly, to keep trackedX/Y in sync.
func (tb *TitleBar) moveWindow(x, y int) {
	rl.SetWindowPosition(x, y)
	tb.trackedX = int32(x)
	tb.trackedY = int32(y)
}

func (tb *TitleBar) rememberNormalBounds() {
	if rl.IsWindowMaximized() {
		return
	}
	tb.savedWX = tb.trackedX
	tb.savedWY = tb.trackedY
	tb.savedWW = tb.windowW
	tb.savedWH = tb.windowH
	tb.hasSavedBounds = true
}

// restoreIfMaximizedForDrag exits maximized and sizes/positions the window for
// a title-bar drag (Windows-style snap-restore under the cursor).
func (tb *TitleBar) restoreIfMaximizedForDrag(screenMX, screenMY float32, mouse rl.Vector2) {
	if !rl.IsWindowMaximized() {
		return
	}
	w, h := int(tb.savedWW), int(tb.savedWH)
	if !tb.hasSavedBounds || w < int(minWinW) || h < int(minWinH) {
		w = int(minWinW * 2)
		if w < 960 {
			w = 960
		}
		h = 640
	}
	fracX := mouse.X / float32(tb.windowW)
	if fracX < 0 {
		fracX = 0
	}
	if fracX > 1 {
		fracX = 1
	}
	newX := int(screenMX - float32(w)*fracX)
	newY := int(screenMY - mouse.Y)
	if newY < 0 {
		newY = 0
	}
	rl.RestoreWindow()
	rl.SetWindowSize(w, h)
	tb.moveWindow(newX, newY)
}

func (tb *TitleBar) beginDragFromTitle(screenMX, screenMY float32, mouse rl.Vector2) {
	tb.restoreIfMaximizedForDrag(screenMX, screenMY, mouse)
	tb.armDragFromTitle(screenMX, screenMY)
}

// armDragFromTitle starts a move drag without restore-from-maximize.
// First click on a normal window arms grab immediately; maximized windows wait
// for a 4px move (or a second click) so double-click restore still works.
func (tb *TitleBar) armDragFromTitle(screenMX, screenMY float32) {
	ResetWheelScrollGesture()
	tb.dragging = true
	tb.dragStartMX = screenMX
	tb.dragStartMY = screenMY
	tb.dragStartWX = tb.trackedX
	tb.dragStartWY = tb.trackedY
}

// seedPos seeds trackedX/Y from GetWindowPosition() exactly once per
// borderless session. After that we trust our own records.
func (tb *TitleBar) seedPos() {
	if !tb.posSeeded {
		p := rl.GetWindowPosition()
		tb.trackedX = int32(p.X)
		tb.trackedY = int32(p.Y)
		tb.posSeeded = true
	}
}

// ResetPos forces the tracked position to be re-seeded from GetWindowPosition()
// on the next Update(). Call this whenever the window mode changes (e.g. after
// toggling FlagWindowUndecorated) so the seed reflects the new window frame.
func (tb *TitleBar) ResetPos() {
	tb.posSeeded = false
}

// ─── Update ───────────────────────────────────────────────────────────────────

// Update processes mouse input for this frame.
// windowW/windowH must match the live SSAA texture size (i.e. the globals from
// main.go, NOT a value written by applyResize).
func (tb *TitleBar) Update(windowW, windowH int32) {
	tb.windowW = windowW
	tb.windowH = windowH

	// Seed our tracked window position if not done yet this borderless session.
	tb.seedPos()

	// Reset hover unconditionally — must not bleed across frames.
	tb.hoverClose = false
	tb.hoverMin = false
	tb.hoverMax = false

	mouse := rl.GetMousePosition()
	// Screen-absolute cursor using our tracked window position — avoids the
	// DWM-frame offset that GetWindowPosition() returns on Windows.
	screenMX := float32(tb.trackedX) + mouse.X
	screenMY := float32(tb.trackedY) + mouse.Y

	leftPressed := rl.IsMouseButtonPressed(rl.MouseButtonLeft)
	leftReleased := rl.IsMouseButtonReleased(rl.MouseButtonLeft)

	// ── Button hover (computed every frame, takes priority over resize) ───────
	closeX := float32(windowW) - closeBtnW
	maxX := closeX - otherBtnW
	minX := maxX - otherBtnW

	tb.hoverClose = rl.CheckCollisionPointRec(mouse,
		rl.NewRectangle(closeX, 0, closeBtnW, TitleBarHeight))
	tb.hoverMax = tb.OnMaximize != nil && rl.CheckCollisionPointRec(mouse,
		rl.NewRectangle(maxX, 0, otherBtnW, TitleBarHeight))
	tb.hoverMin = tb.OnMinimize != nil && rl.CheckCollisionPointRec(mouse,
		rl.NewRectangle(minX, 0, otherBtnW, TitleBarHeight))

	// ── Button clicks — highest priority, always checked ─────────────────────
	if leftPressed {
		if tb.hoverClose {
			tb.closeArmed = true
		} else {
			tb.closeArmed = false
		}
		if tb.hoverMax {
			tb.OnMaximize()
			tb.lastClickTime = 0
			return
		}
		if tb.hoverMin {
			tb.OnMinimize()
			return
		}
	}
	if leftReleased {
		if tb.closeArmed && tb.hoverClose && tb.OnClose != nil {
			tb.OnClose()
		}
		tb.closeArmed = false
	}

	// ── Active resize drag ────────────────────────────────────────────────────
	if tb.HandleResize && tb.resizing {
		dx := screenMX - tb.resizeStartMX
		dy := screenMY - tb.resizeStartMY
		tb.applyResize(dx, dy)
		if leftReleased {
			tb.resizing = false
			tb.resizeEdge = edgeNone
		}
		return // skip drag / double-click while resizing
	}

	// ── Resize initiation (click to start; cursor in ApplyResizeCursor) ───────
	if tb.HandleResize {
		edge := tb.edgeAt(mouse, windowW, windowH)
		if leftPressed && edge != edgeNone {
			ResetWheelScrollGesture()
			tb.resizing = true
			tb.resizeEdge = edge
			tb.resizeStartMX = screenMX
			tb.resizeStartMY = screenMY
			tb.resizeStartWX = tb.trackedX
			tb.resizeStartWY = tb.trackedY
			tb.resizeStartWW = windowW
			tb.resizeStartWH = windowH
			return
		}
	}

	// ── Title-area drag / double-click to toggle maximize ────────────────────
	barRect := rl.NewRectangle(0, 0, float32(windowW), TitleBarHeight)
	if leftPressed &&
		rl.CheckCollisionPointRec(mouse, barRect) &&
		!tb.hoverClose && !tb.hoverMax && !tb.hoverMin {

		now := rl.GetTime()
		interval := titleBarDoubleClickInterval
		if titleBarNativeDblClickMaximize == nil {
			interval = 0.35
		}
		if now-tb.lastClickTime < interval {
			if titleBarNativeDblClickMaximize != nil {
				titleBarNativeDblClickMaximize()
			} else if tb.OnMaximize != nil {
				tb.OnMaximize()
			}
			tb.lastClickTime = 0
			tb.titleClickPending = false
			tb.dragging = false
		} else if titleBarNativeDblClickMaximize != nil {
			// Both behaviors (§13.5 + double-click):
			//  - always mark pending so WebViewDeferChromeRaise stays on
			//  - if already restored: armDrag immediately (grab works like before)
			//  - if maximized: wait for 4px move before restore+drag (dblclick works)
			tb.lastClickTime = now
			tb.titleClickPending = true
			tb.titleClickStartMX = screenMX
			tb.titleClickStartMY = screenMY
			tb.dragStartWX = tb.trackedX
			tb.dragStartWY = tb.trackedY
			if !rl.IsWindowMaximized() {
				tb.armDragFromTitle(screenMX, screenMY)
			}
		} else {
			tb.lastClickTime = now
			tb.beginDragFromTitle(screenMX, screenMY, mouse)
		}
	}

	leftDown := rl.IsMouseButtonDown(rl.MouseButtonLeft)
	if leftReleased {
		tb.dragging = false
		tb.titleClickPending = false
	}

	// Maximized → drag-out: only after 4px so the first click of a double-click
	// does not restore before the second click arrives.
	if tb.titleClickPending && leftDown && rl.IsWindowMaximized() {
		dx := screenMX - tb.titleClickStartMX
		dy := screenMY - tb.titleClickStartMY
		if dx*dx+dy*dy > 16 {
			tb.titleClickPending = false
			tb.beginDragFromTitle(screenMX, screenMY, mouse)
		}
	}

	if tb.dragging {
		newX := int(tb.dragStartWX) + int(screenMX-tb.dragStartMX)
		newY := int(tb.dragStartWY) + int(screenMY-tb.dragStartMY)
		if newY < 0 {
			newY = 0
		}
		tb.moveWindow(newX, newY)
	}

	if !rl.IsWindowMaximized() {
		tb.rememberNormalBounds()
	}
}

// ─── Resize helpers ───────────────────────────────────────────────────────────

// edgeAt returns the resize zone under the mouse.
//
// Corners are checked FIRST, before the button-zone exclusion. This is
// intentional: the NE corner physically overlaps the close button, but
// button clicks are checked before resize starts in Update(), so clicking
// the close button still works — only a drag activates the corner grip.
// The cursor feedback will show a resize arrow in the corner area, which
// matches Windows 11 native behaviour.
func (tb *TitleBar) edgeAt(mouse rl.Vector2, w, h int32) resizeEdge {
	x, y := mouse.X, mouse.Y
	fw, fh := float32(w), float32(h)
	cg := cornerGrip
	sg := sideGrip

	// ── Corners take absolute priority (largest hit target, best UX). ─────────
	switch {
	case x < cg && y < cg:
		return edgeNW
	case x > fw-cg && y < cg:
		return edgeNE
	case x < cg && y > fh-cg:
		return edgeSW
	case x > fw-cg && y > fh-cg:
		return edgeSE
	}

	// ── Button zone: block edge grips in the title bar button strip. ──────────
	// Corners above already handled, so this only blocks edge grips (N / E)
	// from overlapping the min/max/close buttons.
	btnZoneW := closeBtnW + otherBtnW + otherBtnW
	if x >= fw-btnZoneW && y < TitleBarHeight {
		return edgeNone
	}

	// ── Edge grips. ───────────────────────────────────────────────────────────
	switch {
	case y < topGrip:
		return edgeN
	case y > fh-sg:
		return edgeS
	case x < sg:
		return edgeW
	case x > fw-sg:
		return edgeE
	}
	return edgeNone
}

func (tb *TitleBar) setEdgeCursor(e resizeEdge) {
	switch e {
	case edgeN, edgeS:
		rl.SetMouseCursor(rl.MouseCursorResizeNS)
	case edgeE, edgeW:
		rl.SetMouseCursor(rl.MouseCursorResizeEW)
	case edgeNE, edgeSW:
		rl.SetMouseCursor(rl.MouseCursorResizeNESW)
	case edgeNW, edgeSE:
		rl.SetMouseCursor(rl.MouseCursorResizeNWSE)
	default:
		rl.SetMouseCursor(rl.MouseCursorDefault)
	}
}

// ApplyResizeCursor sets borderless window edge/corner resize cursors when
// HandleResize is enabled. Call after the scene tree Update each frame so
// widget cursor resets do not overwrite chrome grips. When the pointer leaves
// the resize band, restores the default arrow (scene nodes may set I-beam etc.
// on the following frame).
func (tb *TitleBar) ApplyResizeCursor(windowW, windowH int32) {
	if !tb.HandleResize {
		return
	}
	var edge resizeEdge
	if tb.resizing {
		edge = tb.resizeEdge
	} else {
		edge = tb.edgeAt(rl.GetMousePosition(), windowW, windowH)
	}
	if edge != edgeNone {
		tb.setEdgeCursor(edge)
		tb.chromeCursorActive = true
		return
	}
	if tb.chromeCursorActive {
		rl.SetMouseCursor(rl.MouseCursorDefault)
		tb.chromeCursorActive = false
	}
}

// applyResize computes and applies updated window bounds from an absolute drag
// delta.  Does NOT write tb.windowW/H — see struct comment for why.
func (tb *TitleBar) applyResize(dx, dy float32) {
	x := int(tb.resizeStartWX)
	y := int(tb.resizeStartWY)
	w := int(tb.resizeStartWW)
	h := int(tb.resizeStartWH)

	idX := int(dx)
	idY := int(dy)

	shrinkW := func(d int) int {
		if w-d < minWinW {
			d = w - minWinW
		}
		return d
	}
	shrinkH := func(d int) int {
		if h-d < minWinH {
			d = h - minWinH
		}
		return d
	}
	growW := func(d int) int {
		if w+d < minWinW {
			d = minWinW - w
		}
		return d
	}
	growH := func(d int) int {
		if h+d < minWinH {
			d = minWinH - h
		}
		return d
	}

	switch tb.resizeEdge {
	case edgeE:
		w += growW(idX)
	case edgeW:
		d := shrinkW(idX)
		x += d
		w -= d
	case edgeS:
		h += growH(idY)
	case edgeN:
		d := shrinkH(idY)
		y += d
		h -= d
	case edgeSE:
		w += growW(idX)
		h += growH(idY)
	case edgeSW:
		d := shrinkW(idX)
		x += d
		w -= d
		h += growH(idY)
	case edgeNE:
		w += growW(idX)
		d := shrinkH(idY)
		y += d
		h -= d
	case edgeNW:
		dw := shrinkW(idX)
		x += dw
		w -= dw
		dh := shrinkH(idY)
		y += dh
		h -= dh
	}

	// Only move the window if its origin actually changed — unnecessary
	// SetWindowPosition calls trigger extra WM_MOVE messages on Windows and
	// cause a subtle jitter on edges/corners that don't change position (E/S/SE).
	if x != int(tb.resizeStartWX) || y != int(tb.resizeStartWY) {
		tb.moveWindow(x, y)
	}
	rl.SetWindowSize(w, h)
}

// titleBarRoundness maps a corner radius to DrawRectangleRounded roundness.
func titleBarRoundness(width, height, cornerRadius float32) float32 {
	if cornerRadius <= 0 || height <= 0 {
		return 0
	}
	half := height / 2
	if width/2 < half {
		half = width / 2
	}
	if half <= 0 {
		return 0
	}
	r := cornerRadius / half
	if r > 1 {
		return 1
	}
	return r
}

// DrawBorderlessWindowFill paints the full client area with rounded corners
// (10px radius, same as title bar). Call after ClearBackground(rl.Blank).
// When maximized, corners are square so the window fills the screen edge-to-edge.
//
// On Windows with DWM corner preference, use a solid rect: the OS clips the
// silhouette. DrawRectangleRounded triangulation leaves faint seams (cross
// hairlines) on large empty fills — very visible on thin hosts like calc.
func DrawBorderlessWindowFill(w, h int32, fill rl.Color) {
	if w < 1 || h < 1 {
		return
	}
	// Pad 1px past client edges so Camera2D/SSAA coverage has no mid-seam gaps
	// from conservative rasterization of exact client bounds.
	r := rl.NewRectangle(-1, -1, float32(w)+2, float32(h)+2)
	if !BorderlessChromeRounded() || nativeBorderlessUsesOSCornerClip() {
		rl.DrawRectangleRec(r, fill)
		return
	}
	rn := titleBarRoundness(r.Width, r.Height, titleChromeRadius)
	if rn > 0 {
		rl.DrawRectangleRounded(r, rn, 32, fill)
	} else {
		rl.DrawRectangle(-1, -1, w+2, h+2, fill)
	}
}

// drawTitleBarBackground fills the title bar; borderless mode uses rounded top corners.
func drawTitleBarBackground(w int32, bg rl.Color, borderlessChrome bool) {
	bar := rl.NewRectangle(0, 0, float32(w), TitleBarHeight)
	if borderlessChrome && BorderlessChromeRounded() {
		rn := titleBarRoundness(bar.Width, bar.Height, titleChromeRadius)
		rl.DrawRectangleRounded(bar, rn, 8, bg)
		// Square off the bottom half so only the top edge is rounded.
		square := rl.NewRectangle(0, TitleBarHeight/2, bar.Width, TitleBarHeight/2)
		rl.DrawRectangleRec(square, bg)
	} else {
		rl.DrawRectangle(0, 0, w, int32(TitleBarHeight), bg)
	}
}

// Title-bar window controls — Remix Icon font (assets/fonts/remixicon.ttf).
// slotScale sets footprint; strokeScale (<1) draws a lighter-weight glyph centered in the slot.
const titleBarIconStrokeThin = float32(0.78)

func titleBarChromeIconSize(rect rl.Rectangle) float32 {
	sz := rect.Height * 0.58
	if rect.Width > 0 && rect.Width*0.72 < sz {
		sz = rect.Width * 0.72
	}
	if sz > 32 {
		sz = 32
	}
	if sz < 20 {
		sz = 20
	}
	return sz
}

func drawTitleBarChromeIcon(rect rl.Rectangle, cp rune, col rl.Color, slotScale, strokeScale float32) {
	if slotScale <= 0 {
		slotScale = 1
	}
	if strokeScale <= 0 {
		strokeScale = titleBarIconStrokeThin
	}
	base := titleBarChromeIconSize(rect)
	sz := base * slotScale
	inner := snapPhosphorRect(rl.NewRectangle(
		rect.X+(rect.Width-sz)/2,
		rect.Y+(rect.Height-sz)/2,
		sz, sz,
	))
	drawRemixIcon(inner, cp, col, strokeScale)
}

// drawTitleBarButtonHover draws a subtle rounded hover plate inside a button zone.
func drawTitleBarButtonHover(rect rl.Rectangle, col rl.Color) {
	inset := float32(6)
	inner := rl.NewRectangle(rect.X+inset, rect.Y+inset, rect.Width-2*inset, rect.Height-2*inset)
	if inner.Width < 4 || inner.Height < 4 {
		rl.DrawRectangleRec(rect, col)
		return
	}
	rn := titleBarRoundness(inner.Width, inner.Height, 6)
	if rn > 0 {
		rl.DrawRectangleRounded(inner, rn, 6, col)
	} else {
		rl.DrawRectangleRec(inner, col)
	}
}

// ─── Draw ─────────────────────────────────────────────────────────────────────

// Draw renders the title bar using the dims stored by the most recent Update().
func (tb *TitleBar) Draw() {
	w := float32(tb.windowW)
	borderlessChrome := tb.HandleResize

	// Theme colours — prefer CurrentTheme["titlebar"] when SetAppearance has run.
	// (Restored from NB_061026 — Style-first broke Studio Full Client seam.)
	var bg, textCol, hoverBg rl.Color
	if tbStyle, ok := CurrentTheme["titlebar"]; ok && tbStyle.BackgroundColor.A > 0 {
		bg = tbStyle.BackgroundColor
		textCol = tbStyle.TextColor
		if hover, ok := CurrentTheme["titlebar-hover"]; ok && hover.BackgroundColor.A > 0 {
			hoverBg = hover.BackgroundColor
		} else {
			hoverBg = rl.NewColor(255, 255, 255, 18)
		}
	} else {
		switch tb.Style {
		case TitleBarStyleLight:
			bg = rl.NewColor(248, 249, 252, 255)
			textCol = rl.NewColor(28, 30, 48, 255)
			hoverBg = rl.NewColor(0, 0, 0, 14)
		case TitleBarStyleAccent:
			bg = rl.NewColor(79, 70, 229, 255)
			textCol = rl.NewColor(255, 255, 255, 255)
			hoverBg = rl.NewColor(255, 255, 255, 28)
		default: // Dark
			bg = rl.NewColor(20, 22, 36, 255)
			textCol = rl.NewColor(218, 220, 236, 255)
			hoverBg = rl.NewColor(255, 255, 255, 18)
		}
	}

	drawTitleBarBackground(tb.windowW, bg, borderlessChrome)

	// Hairline under title bar — shared with ribbon/menubar below (no drop shadow).
	// DrawLineEx at Height-0.5 matches NB_061026 / Studio (do not move onto content band).
	lineCol := rl.NewColor(222, 226, 230, 255)
	if tbStyle, ok := CurrentTheme["titlebar"]; ok && tbStyle.BorderColor.A > 0 {
		lineCol = tbStyle.BorderColor
	} else if ribbon, ok := CurrentTheme["toolbar-ribbon"]; ok && ribbon.BorderColor.A > 0 {
		lineCol = ribbon.BorderColor
	}
	ruleY := TitleBarHeight - 0.5
	rl.DrawLineEx(
		rl.NewVector2(0, ruleY),
		rl.NewVector2(w, ruleY),
		1, lineCol,
	)

	btnTotalW := closeBtnW + otherBtnW + otherBtnW

	if tb.ShowAppIcon && tb.DrawLeadingIcon != nil {
		pad := float32(12)
		icon := TitleBarAppIconSize
		iconRect := rl.NewRectangle(pad, (TitleBarHeight-icon)/2, icon, icon)
		tb.DrawLeadingIcon(iconRect)
	}

	// Title — centered in the FULL window width, but pulled left if it would
	// overlap the right-side buttons.
	if tb.ShowTitle && tb.Title != "" {
		titleAreaW := w - btnTotalW // safe zone: left edge to first button
		titleStyle := GetThemeStyle("default")
		titleStyle.FontSize = int32(titleFontSize)
		titleStyle.TextColor = textCol
		titleStyle.Bold = false
		titleFS := EffectiveFontSize(titleStyle)
		titleW := float32(measureTextS(tb.Title, titleStyle))
		titleY := (TitleBarHeight - titleFS) / 2
		if titleW < titleAreaW {
			titleX := (w - titleW) / 2 // ideal: dead-centre in window
			if titleX+titleW > w-btnTotalW {
				// Would overlap buttons — shift left to flush with button area
				titleX = (titleAreaW - titleW) / 2
			}
			if titleX < 12 {
				titleX = 12
			}
			drawTextS(tb.Title, int32(titleX), int32(titleY+1), titleStyle)
		}
	}

	// ── Control buttons ───────────────────────────────────────────────────────
	closeX := w - closeBtnW
	maxX := closeX - otherBtnW
	minX := maxX - otherBtnW

	closeRect := rl.NewRectangle(closeX, 0, closeBtnW, TitleBarHeight)
	maxRect := rl.NewRectangle(maxX, 0, otherBtnW, TitleBarHeight)
	minRect := rl.NewRectangle(minX, 0, otherBtnW, TitleBarHeight)

	// Hover backgrounds (rounded inset plates)
	if tb.hoverMin {
		drawTitleBarButtonHover(minRect, hoverBg)
	}
	if tb.hoverMax {
		drawTitleBarButtonHover(maxRect, hoverBg)
	}
	if tb.hoverClose {
		drawTitleBarButtonHover(closeRect, rl.NewColor(196, 43, 28, 255))
	}

	symCol := textCol
	closeSymCol := textCol
	if tb.hoverClose {
		closeSymCol = rl.White
	}

	// ─ Minimize (em dash)
	if tb.OnMinimize != nil {
		drawTitleBarChromeIcon(minRect, RemixSubtract, symCol, 1.0, titleBarIconStrokeThin)
	}

	// □ Maximize / restore-to-window
	if tb.OnMaximize != nil {
		if rl.IsWindowMaximized() {
			drawTitleBarChromeIcon(maxRect, RemixCheckboxMultipleBlank, symCol, 0.75, titleBarIconStrokeThin)
		} else {
			drawTitleBarChromeIcon(maxRect, RemixSquare, symCol, 0.75, titleBarIconStrokeThin)
		}
	}

	// × Close
	drawTitleBarChromeIcon(closeRect, RemixClose, closeSymCol, 1.0, titleBarIconStrokeThin)
}

// DrawResizeGrip draws a small phosphor resize hint in the bottom-right corner.
func (tb *TitleBar) DrawResizeGrip() {
	if !tb.HandleResize {
		return
	}
	w := float32(tb.windowW)
	h := float32(tb.windowH)
	const icon = float32(12)
	inner := rl.NewRectangle(
		w-cornerGrip+(cornerGrip-icon)*0.5,
		h-cornerGrip+(cornerGrip-icon)*0.5,
		icon, icon,
	)
	col := rl.NewColor(140, 145, 170, 170)
	Phosphor.EnsureLoaded(PhosphorResize, PhosphorRegular)
	Phosphor.Draw(inner, PhosphorResize, PhosphorRegular, col)
}

// SetSize updates the logical window dimensions used by Draw() without
// processing any input. Call this after a same-frame resize (applyResize has
// already called SetWindowSize) so that Draw() positions buttons at the correct
// right-edge for this frame rather than waiting until the next frame.
func (tb *TitleBar) SetSize(w, h int32) {
	tb.windowW = w
	tb.windowH = h
}

// ─── Style helpers ────────────────────────────────────────────────────────────

// TitleBarStyleName returns the human-readable name for a style constant.
func TitleBarStyleName(s TitleBarStyle) string {
	switch s {
	case TitleBarStyleLight:
		return "Light"
	case TitleBarStyleAccent:
		return "Accent (Indigo)"
	default:
		return "Dark"
	}
}
