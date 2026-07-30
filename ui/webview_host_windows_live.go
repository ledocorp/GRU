//go:build windows && webview2

// Windows live WebView2 host. HWND raise/passthrough/bounds rules are documented in
// docs/WEBVIEW2_HOST.md §13 — read before changing present(), SetBounds, or passthrough.

package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
	"github.com/wailsapp/go-webview2/pkg/edge"
	"github.com/wailsapp/go-webview2/webviewloader"
	"golang.org/x/sys/windows"
)

var (
	winUser32            = windows.NewLazySystemDLL("user32.dll")
	winGdi32             = windows.NewLazySystemDLL("gdi32.dll")
	winPeekMessageW      = winUser32.NewProc("PeekMessageW")
	winTranslateMessage  = winUser32.NewProc("TranslateMessage")
	winDispatchMessageW  = winUser32.NewProc("DispatchMessageW")
	winGetClientRect     = winUser32.NewProc("GetClientRect")
	winGetWindowRect     = winUser32.NewProc("GetWindowRect")
	winScreenToClient    = winUser32.NewProc("ScreenToClient")
	winSetWindowPos      = winUser32.NewProc("SetWindowPos")
	winSetWindowRgn      = winUser32.NewProc("SetWindowRgn")
	winEnumChildWindows  = winUser32.NewProc("EnumChildWindows")
	winGetClassNameW      = winUser32.NewProc("GetClassNameW")
	winGetWindowLongPtrW  = winUser32.NewProc("GetWindowLongPtrW")
	winSetWindowLongPtrW  = winUser32.NewProc("SetWindowLongPtrW")
	winSetFocus           = winUser32.NewProc("SetFocus")
	winGetFocus           = winUser32.NewProc("GetFocus")
	winGetParent          = winUser32.NewProc("GetParent")
	winCreateRoundRectRgn = winGdi32.NewProc("CreateRoundRectRgn")
	enumRaiseChromeCB     uintptr
	enumDemoteChromeCB    uintptr
	enumClipChromeCB      uintptr
	enumClearChromeCB     uintptr
	enumPassThroughOnCB   uintptr
	enumPassThroughOffCB  uintptr
)

const (
	winMsgPMRemove          = 0x0001
	webViewPumpPerFrame     = 6
	webViewPumpIdlePerFrame = 4
	// webViewWindowCornerRadius matches Win11 DWM default rounding at 96 DPI.
	webViewWindowCornerRadius = 8
	hwndTop                = 0
	hwndBottom             = 1
	swpNoMove           = 0x0002
	swpNoSize           = 0x0001
	swpNoActivate       = 0x0010
	swpShowWindow = 0x0040
	wsExTransparent = 0x00000020
)

func gwlExStyleIdx() uintptr {
	var v int32 = -20
	return uintptr(v)
}

type winMsg struct {
	HWND    uintptr
	Message uint32
	WParam  uintptr
	LParam  uintptr
	Time    uint32
	Pt      struct {
		X int32
		Y int32
	}
}

type winRECT struct {
	Left, Top, Right, Bottom int32
}

const gruBridgeInitScriptFallback = `(function(){
  const gru = window.gru = window.gru || {};
  gru.postMessage = function(s) {
    if (window.chrome && window.chrome.webview) window.chrome.webview.postMessage(s);
  };
})();`

var (
	gruBridgeInitOnce sync.Once
	gruBridgeInitBody string
)

func gruBridgeInitScript() string {
	gruBridgeInitOnce.Do(func() {
		if b, err := ReadAssetFile("assets/web/gru.js"); err == nil && len(b) > 0 {
			gruBridgeInitBody = string(b)
			return
		}
		gruBridgeInitBody = gruBridgeInitScriptFallback
	})
	return gruBridgeInitBody
}

type winPOINT struct {
	X int32
	Y int32
}

type chromeClipEnumData struct {
	parent      uintptr
	want        edge.Rect
	radius      int32
	passthrough bool
}

func init() {
	enumRaiseChromeCB = syscall.NewCallback(enumRaiseChromeProc)
	enumDemoteChromeCB = syscall.NewCallback(enumDemoteChromeProc)
	enumClipChromeCB = syscall.NewCallback(enumApplyChromeClipProc)
	enumClearChromeCB = syscall.NewCallback(enumClearChromeClipProc)
	enumPassThroughOnCB = syscall.NewCallback(enumPassThroughOnProc)
	enumPassThroughOffCB = syscall.NewCallback(enumPassThroughOffProc)
}

func webViewHostPlatformSupported() bool {
	_, err := webviewloader.GetAvailableCoreWebView2BrowserVersionString("")
	return err == nil
}

func afterSyncWebViewHosts() {
	webViewForceBoundsSync = false
}

func presentWebViewHostsPlatform() {
	if activeWebViewHostCount() == 0 {
		return
	}
	webViewHostMu.Lock()
	hosts := append([]WebViewHost(nil), webViewHostAlive...)
	dirty := webViewHostsDirty
	webViewHostsDirty = false
	webViewHostMu.Unlock()
	syncWebViewKeyboardFocusFromOS()
	var pumpParent uintptr
	for _, host := range hosts {
		if h, ok := host.(*webViewLiveHost); ok {
			if pumpParent == 0 {
				pumpParent = h.parentHWND
			}
			h.present()
		}
	}
	pump := webViewPumpIdlePerFrame
	if dirty || activeWebViewHostCount() > 0 {
		pump = webViewPumpPerFrame
	}
	if !WebViewDeferChromeRaise() {
		pumpWin32Messages(pumpParent, webViewPresentPumpMax(pump))
	}
}

func pumpWin32Messages(parentHWND uintptr, max int) {
	if max <= 0 || parentHWND == 0 {
		return
	}
	var msg winMsg
	for i := 0; i < max; i++ {
		ret, _, _ := winPeekMessageW.Call(
			uintptr(unsafe.Pointer(&msg)),
			parentHWND,
			0,
			0,
			winMsgPMRemove,
		)
		if ret == 0 {
			break
		}
		winTranslateMessage.Call(uintptr(unsafe.Pointer(&msg)))
		winDispatchMessageW.Call(uintptr(unsafe.Pointer(&msg)))
	}
}

type webViewLiveHost struct {
	mu              sync.Mutex
	chromium        *edge.Chromium
	parentHWND      uintptr
	panel           *WebViewPanel
	status          WebViewHostStatus
	errMsg          string
	onMessage       func(string)
	visible         bool
	lastBounds      rl.Rectangle
	lastHWNDBounds  edge.Rect
	hasBounds       bool
	clipFillClient     bool
	settingsApplied    bool
	inputPassthrough   bool
	keyboardFocused    bool
	destroyed          bool
}

func newWebViewHostPlatform(parentHWND uintptr) WebViewHost {
	h := &webViewLiveHost{parentHWND: parentHWND, visible: true}
	if parentHWND == 0 {
		h.status = WebViewHostUnsupported
		return h
	}
	if !webViewHostPlatformSupported() {
		h.status = WebViewHostError
		h.errMsg = "WebView2 runtime not installed"
		return h
	}
	if err := h.init(); err != nil {
		h.status = WebViewHostError
		h.errMsg = err.Error()
		return h
	}
	h.status = WebViewHostReady
	registerWebViewHost(h)
	return h
}

func (h *webViewLiveHost) init() error {
	ch := edge.NewChromium()
	ch.DataPath = filepath.Join(os.Getenv("AppData"), "github.com/ledocorp/gru", "WebView2")
	ch.SetErrorCallback(func(err error) {
		h.mu.Lock()
		h.status = WebViewHostError
		if err != nil {
			h.errMsg = err.Error()
		}
		h.mu.Unlock()
	})
	ch.MessageCallback = func(message string, _ *edge.ICoreWebView2, _ *edge.ICoreWebView2WebMessageReceivedEventArgs) {
		h.mu.Lock()
		cb := h.onMessage
		h.mu.Unlock()
		deliverWebViewMessage(message, cb)
	}
	ch.ProcessFailedCallback = func(_ *edge.ICoreWebView2, _ *edge.ICoreWebView2ProcessFailedEventArgs) {
		h.mu.Lock()
		h.status = WebViewHostError
		h.errMsg = "WebView2 renderer process exited"
		h.mu.Unlock()
	}
	h.chromium = ch

	if !ch.Embed(h.parentHWND) {
		return syscall.EINVAL
	}
	ch.Init(gruBridgeInitScript())
	return nil
}

func (h *webViewLiveHost) bindPanel(wv *WebViewPanel) {
	h.mu.Lock()
	h.panel = wv
	h.settingsApplied = false
	h.mu.Unlock()
}

func (h *webViewLiveHost) tryApplyWebSettings() {
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady {
		return
	}
	h.mu.Lock()
	if h.settingsApplied {
		h.mu.Unlock()
		return
	}
	panel := h.panel
	h.mu.Unlock()
	if h.chromium.GetController() == nil {
		return
	}
	wv, err := h.chromium.GetController().GetCoreWebView2()
	if err != nil || wv == nil {
		return
	}
	settings, err := wv.GetSettings()
	if err != nil || settings == nil {
		return
	}
	policy := WebViewHostPolicyCurrent()
	if panel != nil {
		policy = panel.effectivePolicy()
	}
	_ = settings.PutAreDevToolsEnabled(policy.DevToolsEnabled)
	_ = settings.PutAreDefaultContextMenusEnabled(policy.DefaultContextMenusEnabled)
	_ = settings.PutAreBrowserAcceleratorKeysEnabled(policy.BrowserAcceleratorKeysEnabled)
	_ = settings.PutIsStatusBarEnabled(policy.StatusBarEnabled)
	_ = settings.PutIsZoomControlEnabled(policy.ZoomControlEnabled)
	_ = settings.PutAreHostObjectsAllowed(policy.HostObjectsAllowed)
	_ = h.chromium.PutIsGeneralAutofillEnabled(policy.AutofillEnabled)
	_ = h.chromium.PutIsPasswordAutosaveEnabled(policy.AutofillEnabled)
	_ = h.chromium.PutIsSwipeNavigationEnabled(policy.SwipeNavigationEnabled)

	h.chromium.AcceleratorKeyCallback = func(vk uint) bool {
		return webViewAcceleratorHandled(vk, policy)
	}
	_ = EnsureWebViewMediaVHost(h.chromium)

	h.mu.Lock()
	h.settingsApplied = true
	h.mu.Unlock()
}

func (h *webViewLiveHost) bindOnMessage(cb func(string)) {
	h.mu.Lock()
	h.onMessage = cb
	h.mu.Unlock()
}

func (h *webViewLiveHost) setClipFillClient(clip bool) {
	h.mu.Lock()
	h.clipFillClient = clip
	h.mu.Unlock()
}

// fillClientChrome reports a visible FillClient host (window-chrome resize experiment).
func (h *webViewLiveHost) fillClientChrome() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.clipFillClient && h.visible && h.hasBounds && !h.destroyed && h.status == WebViewHostReady
}

func (h *webViewLiveHost) Navigate(url string) error {
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady {
		return nil
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return nil
	}
	if !allowedWebNavigation(h.panelPolicy(), h.panel, url) {
		h.status = WebViewHostError
		h.errMsg = "blocked URL (HTTPS allowlist)"
		return syscall.EACCES
	}
	if err := ensureMediaVHostForNavigate(h.chromium, url); err != nil {
		return err
	}
	h.chromium.Navigate(url)
	return nil
}

func (h *webViewLiveHost) SetHTML(html string) error {
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady {
		return nil
	}
	html = strings.TrimSpace(html)
	if html == "" {
		return nil
	}
	h.chromium.NavigateToString(html)
	return nil
}

func (h *webViewLiveHost) SetBounds(r rl.Rectangle) error {
	if h.chromium == nil || h.status != WebViewHostReady || h.destroyed {
		return nil
	}
	if r.Width < 1 || r.Height < 1 {
		if h.hasBounds && (webViewForceBoundsSync || ChromeWindowMoving()) {
			return nil
		}
		_ = h.chromium.Hide()
		h.hasBounds = false
		return nil
	}
	bounds := hwndBoundsFromPanel(r, h.parentHWND)
	force := webViewForceBoundsSync
	if !force && h.hasBounds && webViewBoundsEqual(h.lastBounds, r) && edgeRectEqual(h.lastHWNDBounds, bounds) {
		h.applyChromeClip(bounds)
		return nil
	}
	h.lastBounds = r
	h.hasBounds = true
	markWebViewHostsDirty()
	return h.applyBounds(r)
}

func (h *webViewLiveHost) applyBounds(r rl.Rectangle) error {
	bounds := hwndBoundsFromPanel(r, h.parentHWND)
	h.lastHWNDBounds = bounds
	h.chromium.ResizeWithBounds(&bounds)
	_ = h.chromium.NotifyParentWindowPositionChanged()
	h.applyChromeClip(bounds)
	if h.visible {
		_ = h.chromium.Show()
	}
	h.setInputPassthrough(false)
	if webViewForceBoundsSync {
		_ = h.chromium.NotifyParentWindowPositionChanged()
	}
	return nil
}

func (h *webViewLiveHost) applyChromeClip(bounds edge.Rect) {
	if h.parentHWND == 0 {
		return
	}
	h.mu.Lock()
	fillClient := h.clipFillClient
	h.mu.Unlock()
	// FillClient: never SetWindowRgn round-rect — top L/R radii open crescents
	// under the flat title-bar edge. DWM already rounds the outer window.
	if fillClient {
		clearChromeClip(h.parentHWND, bounds)
		return
	}
	clearChromeClip(h.parentHWND, bounds)
}

func (h *webViewLiveHost) present() {
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady || !h.visible || !h.hasBounds {
		return
	}
	// §13.5 only: demote orphans, Show (never hide for title drag), raise when not deferred.
	// Do not invent extra demote passes — they buried FillClient (webviewhello lock).
	demoteStrayChromeHWNDs(h.parentHWND, h.lastHWNDBounds)
	_ = h.chromium.Show()
	if WebViewDeferChromeRaise() {
		return
	}
	raiseChromeChildWindows(h.parentHWND, h.lastHWNDBounds)
}

func (h *webViewLiveHost) holdsWebKeyboard() bool {
	h.mu.Lock()
	kf := h.keyboardFocused
	panel := h.panel
	h.mu.Unlock()
	return kf || (webViewKeyboardPanel != nil && panel == webViewKeyboardPanel)
}

func (h *webViewLiveHost) hostKeepsWebKeyboard() bool {
	return h.holdsWebKeyboard()
}

func (h *webViewLiveHost) setInputPassthrough(passthrough bool) {
	if h.destroyed || h.status != WebViewHostReady {
		return
	}
	if passthrough && h.holdsWebKeyboard() {
		passthrough = false
	}
	h.mu.Lock()
	prev := h.inputPassthrough
	if h.inputPassthrough == passthrough {
		h.mu.Unlock()
		return
	}
	h.inputPassthrough = passthrough
	want := h.lastHWNDBounds
	parent := h.parentHWND
	panelID := ""
	if h.panel != nil {
		panelID = h.panel.ID()
	}
	h.mu.Unlock()
	logWebViewDebug("passthrough", fmt.Sprintf("panel=%s %t→%t kbd=%t", panelID, prev, passthrough, h.holdsWebKeyboard()))
	h.applyChromeInputPassthrough(parent, want, passthrough)
}

func (h *webViewLiveHost) applyChromeInputPassthrough(parent uintptr, want edge.Rect, passthrough bool) {
	if parent == 0 {
		return
	}
	data := chromeClipEnumData{parent: parent, want: want, passthrough: passthrough}
	cb := enumPassThroughOffCB
	if passthrough {
		cb = enumPassThroughOnCB
	}
	winEnumChildWindows.Call(parent, cb, uintptr(unsafe.Pointer(&data)))
}

func enumPassThroughOnProc(hwnd, lParam uintptr) uintptr {
	return enumPassThroughProc(hwnd, lParam)
}

func enumPassThroughOffProc(hwnd, lParam uintptr) uintptr {
	return enumPassThroughProc(hwnd, lParam)
}

func enumPassThroughProc(hwnd, lParam uintptr) uintptr {
	if hwnd == 0 || lParam == 0 {
		return 1
	}
	data := (*chromeClipEnumData)(unsafe.Pointer(lParam))
	class := windowClassName(hwnd)
	if !strings.Contains(class, "Chrome") && !strings.Contains(class, "WebView") {
		return 1
	}
	child, ok := chromeChildBoundsInParent(hwnd, data.parent)
	if !ok || !chromeBoundsNear(child, data.want, chromeHWNDMatchTolerance) {
		return 1
	}
	setChromeHWNDPassthrough(hwnd, data.passthrough)
	return 1
}

// chromeContentTopPhysical is the client Y (physical px) below borderless title chrome.
func chromeContentTopPhysical(parentHWND uintptr) int32 {
	doc := ActiveDocument()
	if doc == nil || doc.ChromeTop() <= 0 {
		return 0
	}
	return int32(doc.ChromeTop() * hwndClientScale(parentHWND))
}

func setChromeHWNDPassthrough(hwnd uintptr, enable bool) {
	style, _, _ := winGetWindowLongPtrW.Call(hwnd, gwlExStyleIdx())
	if enable {
		if style&wsExTransparent != 0 {
			return
		}
		winSetWindowLongPtrW.Call(hwnd, gwlExStyleIdx(), style|wsExTransparent)
		return
	}
	if style&wsExTransparent == 0 {
		return
	}
	winSetWindowLongPtrW.Call(hwnd, gwlExStyleIdx(), style&^wsExTransparent)
}

func (h *webViewLiveHost) SetVisible(visible bool) {
	if h.visible != visible && WebViewDebugEnabled {
		h.mu.Lock()
		panelID := ""
		if h.panel != nil {
			panelID = h.panel.ID()
		}
		h.mu.Unlock()
		logWebViewDebug("visible", fmt.Sprintf("panel=%s %t→%t", panelID, h.visible, visible))
	}
	h.visible = visible
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady {
		return
	}
	if visible {
		if h.hasBounds {
			_ = h.applyBounds(h.lastBounds)
		}
	} else {
		_ = h.chromium.Hide()
	}
}

func (h *webViewLiveHost) SetFocus() {
	if !h.destroyed && h.chromium != nil && h.status == WebViewHostReady {
		h.chromium.Focus()
	}
}

func (h *webViewLiveHost) Blur() {
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady {
		return
	}
	h.mu.Lock()
	h.keyboardFocused = false
	panel := h.panel
	h.mu.Unlock()
	if panel == webViewKeyboardPanel {
		notifyWebViewLostKeyboard()
	}
	h.chromium.ReleaseFocus()
	if h.parentHWND != 0 {
		winSetFocus.Call(h.parentHWND)
	}
}

func (h *webViewLiveHost) PostMessage(json string) error {
	if h.destroyed || h.chromium == nil || h.status != WebViewHostReady {
		return nil
	}
	ctrl := h.chromium.GetController()
	if ctrl == nil {
		return nil
	}
	wv, err := ctrl.GetCoreWebView2()
	if err != nil || wv == nil {
		return err
	}
	return wv.PostWebMessageAsString(json)
}

func (h *webViewLiveHost) evalScript(script string) {
	if !h.destroyed && h.chromium != nil && h.status == WebViewHostReady {
		h.chromium.Eval(script)
	}
}

func (h *webViewLiveHost) Destroy() {
	if h.destroyed {
		return
	}
	h.destroyed = true
	unregisterWebViewHost(h)
	if h.chromium != nil {
		h.chromium.ShuttingDown()
		_ = h.chromium.Hide()
		h.status = WebViewHostStub
		zero := edge.Rect{}
		h.chromium.ResizeWithBounds(&zero)
		h.chromium = nil
	}
	h.status = WebViewHostStub
	h.hasBounds = false
}

func (h *webViewLiveHost) Status() WebViewHostStatus { return h.status }
func (h *webViewLiveHost) LastError() string         { return h.errMsg }

func hwndClientScale(parentHWND uintptr) float32 {
	if parentHWND == 0 {
		return DisplayScale
	}
	var rc winRECT
	ok, _, _ := winGetClientRect.Call(parentHWND, uintptr(unsafe.Pointer(&rc)))
	if ok == 0 {
		return DisplayScale
	}
	physicalW := float32(rc.Right - rc.Left)
	logicalW := float32(0)
	if doc := ActiveDocument(); doc != nil && doc.Width > 0 {
		logicalW = float32(doc.Width)
	}
	if logicalW <= 0 {
		logicalW = physicalW / DisplayScale
	}
	if physicalW <= 0 || logicalW <= 0 {
		return DisplayScale
	}
	s := physicalW / logicalW
	if s < 1 {
		return 1
	}
	if s > 3 {
		return 3
	}
	return s
}

func hwndBoundsFromPanel(r rl.Rectangle, parentHWND uintptr) edge.Rect {
	s := hwndClientScale(parentHWND)
	x := int32(r.X * s)
	y := int32(r.Y * s)
	w := int32(r.Width * s)
	h := int32(r.Height * s)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return edge.Rect{
		Left:   x,
		Top:    y,
		Right:  x + w,
		Bottom: y + h,
	}
}

func raiseChromeChildWindows(parentHWND uintptr, want edge.Rect) {
	if parentHWND == 0 || enumRaiseChromeCB == 0 {
		return
	}
	data := chromeClipEnumData{parent: parentHWND, want: want}
	winEnumChildWindows.Call(parentHWND, enumRaiseChromeCB, uintptr(unsafe.Pointer(&data)))
}

// demoteStrayChromeHWNDs pushes Chrome/WebView child HWNDs that are not
// chromeBoundsNear(want, 8px) to HWND_BOTTOM. WebView2 creates full-client-sized
// orphan surfaces; raising them caused title-bar drag/maximize to break (§13.5).
func demoteStrayChromeHWNDs(parentHWND uintptr, want edge.Rect) {
	if parentHWND == 0 || enumDemoteChromeCB == 0 {
		return
	}
	data := chromeClipEnumData{parent: parentHWND, want: want}
	winEnumChildWindows.Call(parentHWND, enumDemoteChromeCB, uintptr(unsafe.Pointer(&data)))
}

const chromeHWNDMatchTolerance = int32(8)

func enumDemoteChromeProc(hwnd, lParam uintptr) uintptr {
	if hwnd == 0 || lParam == 0 {
		return 1
	}
	data := (*chromeClipEnumData)(unsafe.Pointer(lParam))
	class := windowClassName(hwnd)
	if !strings.Contains(class, "Chrome") && !strings.Contains(class, "WebView") {
		return 1
	}
	child, ok := chromeChildBoundsInParent(hwnd, data.parent)
	if !ok {
		return 1
	}
	if chromeBoundsNear(child, data.want, chromeHWNDMatchTolerance) {
		return 1
	}
	winSetWindowPos.Call(
		hwnd,
		hwndBottom,
		0,
		0,
		0,
		0,
		swpNoMove|swpNoSize|swpNoActivate,
	)
	return 1
}

func enumRaiseChromeProc(hwnd, lParam uintptr) uintptr {
	if hwnd == 0 || lParam == 0 {
		return 1
	}
	data := (*chromeClipEnumData)(unsafe.Pointer(lParam))
	class := windowClassName(hwnd)
	if !strings.Contains(class, "Chrome") && !strings.Contains(class, "WebView") {
		return 1
	}
	child, ok := chromeChildBoundsInParent(hwnd, data.parent)
	if !ok || !chromeBoundsNear(child, data.want, chromeHWNDMatchTolerance) {
		return 1
	}
	if top := chromeContentTopPhysical(data.parent); top > 0 {
		// §13.3: never raise an HWND that invades the title bar band (no underlap slack).
		if child.Top < top {
			return 1
		}
	}
	winSetWindowPos.Call(
		hwnd,
		hwndTop,
		0,
		0,
		0,
		0,
		swpNoMove|swpNoSize|swpNoActivate|swpShowWindow,
	)
	return 1
}

func windowClassName(hwnd uintptr) string {
	if hwnd == 0 {
		return ""
	}
	buf := make([]uint16, 256)
	n, _, _ := winGetClassNameW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return ""
	}
	return windows.UTF16ToString(buf[:n])
}

func webViewBoundsEqual(a, b rl.Rectangle) bool {
	return a.X == b.X && a.Y == b.Y && a.Width == b.Width && a.Height == b.Height
}

func edgeRectEqual(a, b edge.Rect) bool {
	return a.Left == b.Left && a.Top == b.Top && a.Right == b.Right && a.Bottom == b.Bottom
}

func (h *webViewLiveHost) panelPolicy() WebViewHostPolicy {
	if h.panel != nil {
		return h.panel.effectivePolicy()
	}
	return WebViewHostPolicyCurrent()
}

func enumApplyChromeClipProc(hwnd, lParam uintptr) uintptr {
	if hwnd == 0 || lParam == 0 {
		return 1
	}
	data := (*chromeClipEnumData)(unsafe.Pointer(lParam))
	class := windowClassName(hwnd)
	if !strings.Contains(class, "Chrome") && !strings.Contains(class, "WebView") {
		return 1
	}
	child, ok := chromeChildBoundsInParent(hwnd, data.parent)
	if !ok || !chromeBoundsNear(child, data.want, 3) {
		return 1
	}
	applyRoundedHWNDRegion(hwnd, data.radius)
	return 1
}

func clearChromeClip(parent uintptr, want edge.Rect) {
	if parent == 0 {
		return
	}
	data := chromeClipEnumData{parent: parent, want: want, radius: 0}
	winEnumChildWindows.Call(parent, enumClearChromeCB, uintptr(unsafe.Pointer(&data)))
}

func enumClearChromeClipProc(hwnd, lParam uintptr) uintptr {
	if hwnd == 0 || lParam == 0 {
		return 1
	}
	data := (*chromeClipEnumData)(unsafe.Pointer(lParam))
	class := windowClassName(hwnd)
	if !strings.Contains(class, "Chrome") && !strings.Contains(class, "WebView") {
		return 1
	}
	child, ok := chromeChildBoundsInParent(hwnd, data.parent)
	if !ok || !chromeBoundsNear(child, data.want, 3) {
		return 1
	}
	winSetWindowRgn.Call(hwnd, 0, 1)
	return 1
}

func chromeChildBoundsInParent(hwnd, parent uintptr) (edge.Rect, bool) {
	var rc winRECT
	if ok, _, _ := winGetWindowRect.Call(hwnd, uintptr(unsafe.Pointer(&rc))); ok == 0 {
		return edge.Rect{}, false
	}
	pt := winPOINT{X: rc.Left, Y: rc.Top}
	if ok, _, _ := winScreenToClient.Call(parent, uintptr(unsafe.Pointer(&pt))); ok == 0 {
		return edge.Rect{}, false
	}
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	return edge.Rect{
		Left:   pt.X,
		Top:    pt.Y,
		Right:  pt.X + w,
		Bottom: pt.Y + h,
	}, true
}

func chromeBoundsNear(a, b edge.Rect, tol int32) bool {
	return abs32(a.Left-b.Left) <= tol &&
		abs32(a.Top-b.Top) <= tol &&
		abs32(a.Right-b.Right) <= tol &&
		abs32(a.Bottom-b.Bottom) <= tol
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func applyRoundedHWNDRegion(hwnd uintptr, radius int32) {
	if hwnd == 0 {
		return
	}
	var rc winRECT
	if ok, _, _ := winGetClientRect.Call(hwnd, uintptr(unsafe.Pointer(&rc))); ok == 0 {
		return
	}
	w := rc.Right - rc.Left
	h := rc.Bottom - rc.Top
	if w < 2 || h < 2 {
		return
	}
	rgn, _, _ := winCreateRoundRectRgn.Call(
		0,
		0,
		uintptr(w+1),
		uintptr(h+1),
		uintptr(radius*2),
		uintptr(radius*2),
	)
	if rgn == 0 {
		return
	}
	winSetWindowRgn.Call(hwnd, rgn, 1)
}

func syncWebViewKeyboardFocusFromOS() {
	fg, _, _ := winGetFocus.Call()
	if fg == 0 {
		return
	}

	webViewHostMu.Lock()
	hosts := append([]WebViewHost(nil), webViewHostAlive...)
	webViewHostMu.Unlock()

	var owner *webViewLiveHost
	for _, host := range hosts {
		h, ok := host.(*webViewLiveHost)
		if !ok || h.destroyed || h.status != WebViewHostReady {
			continue
		}
		if h.chromeOwnsFocusHWND(fg) {
			owner = h
			break
		}
	}
	if owner == nil {
		// Do not clear keyboard state here — GetFocus flickers to the parent HWND
		// between frames; Blur / SetFocus(native) / RouteScenePointerFocus clear instead.
		return
	}

	for _, host := range hosts {
		h, ok := host.(*webViewLiveHost)
		if !ok {
			continue
		}
		if h == owner {
			h.mu.Lock()
			was := h.keyboardFocused
			h.keyboardFocused = true
			panel := h.panel
			h.mu.Unlock()
			if !was && panel != nil {
				notifyWebViewGotKeyboard(panel)
			}
			h.setInputPassthrough(false)
		} else {
			h.mu.Lock()
			h.keyboardFocused = false
			h.mu.Unlock()
		}
	}
}

func (h *webViewLiveHost) chromeOwnsFocusHWND(fg uintptr) bool {
	if fg == 0 || h.parentHWND == 0 {
		return false
	}
	if !hwndDescendantOf(fg, h.parentHWND) {
		return false
	}
	for cur := fg; cur != 0 && cur != h.parentHWND; {
		class := windowClassName(cur)
		if strings.Contains(class, "Chrome") || strings.Contains(class, "WebView") {
			return true
		}
		cur, _, _ = winGetParent.Call(cur)
	}
	return false
}

func hwndDescendantOf(hwnd, ancestor uintptr) bool {
	for cur := hwnd; cur != 0; {
		if cur == ancestor {
			return true
		}
		cur, _, _ = winGetParent.Call(cur)
	}
	return false
}
