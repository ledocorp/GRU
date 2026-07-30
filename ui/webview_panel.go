// Package ui (continued) — WebViewPanel reserves layout space for an embedded web surface.
//
// See docs/WEBVIEW2_HOST.md (especially §13 HWND pitfalls) and webview_host.go.
package ui

import (
	"encoding/json"
	"strings"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// WebViewPanel is a ui.Node that owns a bounded web host region inside the retained tree.
//
// When the live host is inactive, Draw renders a placeholder frame so layout and demos
// work on all platforms. Call SyncWebViewHosts from the main loop when B+2.2 links WebView2.
//
// # LLM Prompt Template
//
//	wv := ui.NewWebViewPanel("foundry-module", "https://example.com", 0, 0, 0, 400)
//	wv.OnMessage = func(payload string) { /* bridge */ }
//	center.AddChild(wv)
//
// Demos: **WebView Foundry Demo** (nested module), **WebView Full Client** (Gru chrome + web),
// **WebView Focus Handoff** (TextEditor + web §5), **WebView Video (T2b)** (gru-video-player).
type WebViewPanel struct {
	Element
	URL        string
	HTML       string
	HostStatus *Signal[string]
	OnMessage  func(payload string)
	// FillClient draws an edge-to-edge placeholder (no debug border). Use when web owns
	// the full client below the title bar (Wails-shaped apps).
	FillClient bool

	// Per-panel policy overrides (nil = use ui.WebViewHostPolicyCurrent()).
	DevToolsEnabled            *bool
	DefaultContextMenusEnabled *bool
	BrowserAcceleratorKeysEnabled *bool
	// AllowHTTP permits non-localhost http:// when global policy does not.
	AllowHTTP bool
	// AllowFileBundle permits file:// navigations (bundle path validation TBD).
	AllowFileBundle bool

	host         WebViewHost
	hostParent   uintptr
	hovered      bool
	themeSynced  bool
	contentSynced bool
}

type webViewMessageBinder interface {
	bindOnMessage(func(string))
}

type webViewScriptHost interface {
	evalScript(string)
}

type webViewClipStyleHost interface {
	setClipFillClient(bool)
}

type webViewPanelBinder interface {
	bindPanel(*WebViewPanel)
	tryApplyWebSettings()
}

type webViewInputPassHost interface {
	setInputPassthrough(bool)
}

// webViewKeyboardHost reports OS keyboard focus on the live HWND (independent of webViewKeyboardPanel).
type webViewKeyboardHost interface {
	hostKeepsWebKeyboard() bool
}

// NewWebViewPanel creates a web panel. Pass w=0, h=0 for flex sizing in parents.
func NewWebViewPanel(id, url string, x, y, w, h float32) *WebViewPanel {
	wv := &WebViewPanel{
		Element:    NewElement(id, x, y, w, h),
		URL:        strings.TrimSpace(url),
		HostStatus: NewSignal(string(WebViewHostStub)),
	}
	wv.styleName = "webview-panel"
	if w > 0 {
		wv.PreferredWidth = w
	}
	return wv
}

// PostMessage sends a JSON envelope to the JS context (no-op when host is stub).
func (wv *WebViewPanel) PostMessage(json string) {
	if wv.host != nil {
		_ = wv.host.PostMessage(json)
	}
}

// BridgeCapabilities returns read-only flags exposed to JS via gru.capabilities.
func (wv *WebViewPanel) BridgeCapabilities() map[string]any {
	if wv == nil {
		return map[string]any{"runtime": 1}
	}
	p := wv.effectivePolicy()
	caps := map[string]any{
		"runtime":   1,
		"fileMedia": p.AllowFileBundle,
		"http":      p.AllowInsecureHTTP,
		"devtools":  p.DevToolsEnabled,
		"mediaHost": "gru.media",
	}
	// FillClient only: page may drive S/E/W/SE/SW window resize via chrome.resize.
	if wv.FillClient {
		caps["windowChromeResize"] = true
	}
	return caps
}
// EmitBridge posts a gru.bridge v1 message to the page (toast, theme, commands).
// In full-client layouts, prefer name "toast" — native Gru toasts draw under the HWND host.
func (wv *WebViewPanel) EmitBridge(name string, payload any) error {
	env := map[string]any{"type": "gru.bridge", "v": 1, "name": name}
	if payload != nil {
		env["payload"] = payload
	}
	b, err := json.Marshal(env)
	if err != nil {
		return err
	}
	if wv.host != nil {
		return wv.host.PostMessage(string(b))
	}
	return nil
}

// ApplyTheme sets data-gru-theme on the hosted document when a live host exists (B+2.2).
func (wv *WebViewPanel) ApplyTheme(dark bool) {
	theme := "light"
	if dark {
		theme = "dark"
	}
	if sh, ok := wv.host.(webViewScriptHost); ok && wv.host.Status() == WebViewHostReady {
		sh.evalScript(`document.documentElement.dataset.gruTheme="` + theme + `"`)
	}
	wv.PostMessage(`{"type":"gru.bridge","v":1,"name":"theme","payload":{"theme":"` + theme + `"}}`)
}

func (wv *WebViewPanel) ensureHost(parentHWND uintptr) {
	if wv.host != nil && wv.hostParent == parentHWND {
		return
	}
	if wv.host != nil {
		wv.host.Destroy()
		wv.host = nil
	}
	wv.hostParent = parentHWND
	if parentHWND == 0 {
		wv.HostStatus.Set(string(WebViewHostUnsupported))
		return
	}
	wv.host = NewWebViewHost(parentHWND)
	if binder, ok := wv.host.(webViewMessageBinder); ok {
		binder.bindOnMessage(wv.OnMessage)
	}
	st := wv.host.Status()
	wv.HostStatus.Set(string(st))
	if st == WebViewHostReady {
		wv.themeSynced = false
		wv.contentSynced = false
	}
}

// syncHost drives the live WebView2 COM controller for this panel each SyncWebViewHosts pass.
//
// Visibility: hidden when the panel is hidden, hostBounds empty (except during
// WebViewForceBoundsSyncActive), or WebViewHostOccluded (modal/drawer). Title-bar
// drag and edge resize do not hide the host.
//
// Navigation runs once (contentSynced); resize must not re-navigate.
// syncPointerPolicy keeps the HWND opaque — see docs/WEBVIEW2_HOST.md §13.4.
func (wv *WebViewPanel) syncHost(doc *Document) {
	if wv.IsHidden() {
		if wv.host != nil {
			wv.host.SetVisible(false)
		}
		return
	}
	b := wv.hostBounds(doc)
	if b.Width < 1 || b.Height < 1 {
		if wv.host != nil && WebViewForceBoundsSyncActive() {
			wv.syncPointerPolicy(doc)
			return
		}
		if wv.host != nil {
			wv.host.SetVisible(false)
		}
		return
	}
	overlayHidden := WebViewHostOccluded()
	hwnd := uintptr(0)
	if doc != nil {
		hwnd = doc.PlatformWindowHandle()
	}
	wv.ensureHost(hwnd)
	if wv.host == nil {
		return
	}
	if binder, ok := wv.host.(webViewMessageBinder); ok {
		binder.bindOnMessage(wv.OnMessage)
	}
	if clip, ok := wv.host.(webViewClipStyleHost); ok {
		clip.setClipFillClient(wv.FillClient)
	}
	if binder, ok := wv.host.(webViewPanelBinder); ok {
		binder.bindPanel(wv)
		binder.tryApplyWebSettings()
	}
	if wv.host.Status() == WebViewHostReady && !wv.contentSynced {
		var navErr error
		if wv.URL != "" {
			navErr = wv.host.Navigate(wv.URL)
		} else if wv.HTML != "" {
			navErr = wv.host.SetHTML(wv.HTML)
		} else {
			wv.contentSynced = true
		}
		if navErr == nil && (wv.URL != "" || wv.HTML != "") {
			wv.contentSynced = true
		}
	}
	_ = wv.host.SetBounds(b)
	if wv.host.Status() == WebViewHostReady && !wv.themeSynced {
		wv.ApplyTheme(ThemeIsDark())
		_ = wv.EmitBridge("capabilities", wv.BridgeCapabilities())
		wv.themeSynced = true
	}
	if overlayHidden {
		wv.host.SetVisible(false)
		wv.syncPointerPolicy(doc)
		return
	}
	wv.host.SetVisible(true)
	wv.syncPointerPolicy(doc)
}

// syncPointerPolicy keeps the live host opaque. WS_EX_TRANSPARENT shows the OpenGL
// UI through the WebView surface (looks blank). Side-by-side layouts route focus
// via RouteScenePointerFocus; the HWND bounds do not cover native widgets beside it.
func (wv *WebViewPanel) syncPointerPolicy(doc *Document) {
	if wv.host == nil || wv.IsHidden() {
		return
	}
	pass, ok := wv.host.(webViewInputPassHost)
	if !ok {
		return
	}
	pass.setInputPassthrough(false)
}

// IsInteractive implements Node.
func (wv *WebViewPanel) IsInteractive() bool { return true }

// Update tracks hover for chrome; click-to-focus is handled by RouteScenePointerFocus.
func (wv *WebViewPanel) Update(_ float32) {
	if wv.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	wv.hovered = rl.CheckCollisionPointRec(mouse, wv.Bounds())
}

func (wv *WebViewPanel) Layout() {
	wv.layoutDirty = false
}

// Draw renders placeholder chrome when the live host is not ready.
func (wv *WebViewPanel) Draw() {
	if wv.IsHidden() {
		return
	}
	if wv.host != nil && wv.host.Status() == WebViewHostReady {
		wv.drawDirty = false
		return
	}
	b := wv.Bounds()
	if wv.FillClient {
		wv.drawFillClientPlaceholder(b)
		wv.drawDirty = false
		return
	}
	style := wv.GetStyle()
	bg := style.BackgroundColor
	if bg.A == 0 {
		bg = rl.NewColor(248, 250, 252, 255)
	}
	rl.DrawRectangleRec(b, bg)
	if style.BorderWidth > 0 {
		rl.DrawRectangleLinesEx(b, style.BorderWidth, style.BorderColor)
	}

	status := wv.HostStatus.Get()
	title := "WebView panel"
	if wv.URL != "" {
		title = wv.URL
	}
	subStyle := style
	subStyle.FontSize = 14
	subStyle.TextColor = rl.NewColor(100, 116, 139, 255)
	drawTextS(title, int32(b.X+12), int32(b.Y+12), style)
	drawTextS("Host: "+status, int32(b.X+12), int32(b.Y+36), subStyle)
	if wv.host != nil && wv.host.LastError() != "" {
		drawTextS(wv.host.LastError(), int32(b.X+12), int32(b.Y+56), subStyle)
	} else {
		drawTextS("Offline preview · rebuild with -tags webview2 for live Edge", int32(b.X+12), int32(b.Y+56), subStyle)
	}
	wv.drawDirty = false
}

func (wv *WebViewPanel) drawFillClientPlaceholder(b rl.Rectangle) {
	bg := rl.NewColor(248, 250, 252, 255)
	rl.DrawRectangleRec(b, bg)

	titleStyle := GetThemeStyle("header")
	titleStyle.FontSize = 26
	titleStyle.TextColor = rl.NewColor(15, 23, 42, 255)
	bodyStyle := GetThemeStyle("form-value")
	bodyStyle.FontSize = 15
	bodyStyle.TextColor = rl.NewColor(71, 85, 105, 255)

	cx := b.X + b.Width/2
	cy := b.Y + b.Height/2 - 40
	title := "Gru Web Shell"
	titleW := float32(measureTextS(title, titleStyle))
	drawTextS(title, int32(cx-titleW/2), int32(cy), titleStyle)

	sub := "Full client below title bar — Wails-shaped layout"
	subW := float32(measureTextS(sub, bodyStyle))
	drawTextS(sub, int32(cx-subW/2), int32(cy+36), bodyStyle)

	status := wv.HostStatus.Get()
	if wv.host != nil && wv.host.LastError() != "" {
		status = wv.host.LastError()
	}
	hintStyle := bodyStyle
	hintStyle.FontSize = 12
	hintStyle.TextColor = rl.NewColor(148, 163, 184, 255)
	hint := "Host: " + status + " · offline preview (add -tags webview2 for live)"
	hintW := float32(measureTextS(hint, hintStyle))
	drawTextS(hint, int32(b.X+b.Width-hintW-12), int32(b.Y+b.Height-24), hintStyle)
}

// webViewFillClientTopInset: FillClient tops at ChromeTop + inset.
// Must stay ≥ 0 — §13.3 forbids HWND into the title bar band (negative
// underlap stole title-bar hits / broke drag). Seam is title-bar hairline, not HWND.
const webViewFillClientTopInset = float32(0)

// visibleHostRect is the OS child-surface rectangle in logical client coordinates.
//
// It starts from layout Bounds(), clamps below Document.ChromeTop() (FillClient
// may add a non-negative inset), then intersects ancestor Panel/Viewport clip
// bounds — the same contract as TextEditor scissor. PresentWebViewHosts maps
// this rect to physical HWND bounds. See docs/WEBVIEW2_HOST.md §13.3.
//
// FillClient stays edge-flush on L/R/bottom. Grip insets and click-through are
// rejected (§13.10). Window + HWND stay tied via live ResizeWithBounds (§13.7).
func (wv *WebViewPanel) visibleHostRect(doc *Document) rl.Rectangle {
	b := wv.Bounds()
	chromeTop := float32(0)
	if doc != nil {
		chromeTop = doc.ChromeTop()
	}
	if chromeTop > 0 {
		desiredTop := chromeTop
		if wv.FillClient {
			desiredTop = chromeTop + webViewFillClientTopInset
			if desiredTop < chromeTop {
				desiredTop = chromeTop
			}
		}
		if b.Y < desiredTop {
			b.Height -= desiredTop - b.Y
			b.Y = desiredTop
		} else if wv.FillClient && b.Y > desiredTop {
			b.Height += b.Y - desiredTop
			b.Y = desiredTop
		}
	}
	if clip, ok := ancestorClipBounds(wv); ok {
		b = intersectRects(b, clip)
	}
	if b.Width < 1 || b.Height < 1 {
		return rl.Rectangle{}
	}
	return b
}

func (wv *WebViewPanel) hostBounds(doc *Document) rl.Rectangle {
	return wv.visibleHostRect(doc)
}

// UnmountHost destroys the OS host; call before removing from tree or on scene swap.
func (wv *WebViewPanel) UnmountHost() {
	if wv.host != nil {
		wv.host.Destroy()
		wv.host = nil
	}
	wv.hostParent = 0
	wv.themeSynced = false
	wv.contentSynced = false
	wv.HostStatus.Set(string(WebViewHostStub))
}
