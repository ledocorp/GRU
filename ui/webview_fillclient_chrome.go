package ui

import (
	"encoding/json"
	"strings"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// FillClient window-chrome resize (experiment): the page owns S/E/W/SE/SW grips
// under the title bar and posts gru.bridge chrome.resize; Go applies the same
// SetWindowSize/Position path as TitleBar. Scoped to FillClient hosts only —
// Focus Handoff and nested embeds are unchanged. See WEBVIEW2_HOST.md §13.12.

var (
	fillClientChromeMu       sync.Mutex
	fillClientChromeTitleBar *TitleBar
	fillClientResize         fillClientChromeResizeState
)

type fillClientChromeResizeState struct {
	active               bool
	edge                 resizeEdge
	startMX, startMY     float32
	startWX, startWY     int32
	startWW, startWH     int32
	pendingMX, pendingMY float32
	hasPendingMove       bool
	wantStart            bool
	startEdge            resizeEdge
	startScreenMX        float32
	startScreenMY        float32
	wantEnd              bool
}

// BindFillClientChromeTitleBar records the borderless TitleBar so FillClient
// web resize keeps tracked window origin in sync (W/SW moves). Call each frame
// from SyncBorderlessChromeFrame / Studio main.
func BindFillClientChromeTitleBar(tb *TitleBar) {
	fillClientChromeMu.Lock()
	fillClientChromeTitleBar = tb
	fillClientChromeMu.Unlock()
}

// FillClientChromeResizing reports an active web-driven window resize.
func FillClientChromeResizing() bool {
	fillClientChromeMu.Lock()
	defer fillClientChromeMu.Unlock()
	return fillClientResize.active || fillClientResize.wantStart
}

// WebViewFillClientHostsActive reports a visible FillClient WebView host.
// Full Client keeps ActiveFPS while focused — nested embeds use WebViewIdleFPS.
func WebViewFillClientHostsActive() bool {
	return webViewFillClientHostsActive()
}

func webViewFillClientHostsActive() bool {
	webViewHostMu.Lock()
	defer webViewHostMu.Unlock()
	for _, h := range webViewHostAlive {
		live, ok := h.(interface{ fillClientChrome() bool })
		if ok && live.fillClientChrome() {
			return true
		}
	}
	return false
}

type fillClientChromeEnvelope struct {
	Type    string          `json:"type"`
	V       int             `json:"v"`
	Name    string          `json:"name"`
	Payload json.RawMessage `json:"payload"`
}

type fillClientChromeResizePayload struct {
	Phase   string  `json:"phase"`
	Edge    string  `json:"edge"`
	ScreenX float32 `json:"screenX"`
	ScreenY float32 `json:"screenY"`
}

// noteFillClientChromeResize records chrome.resize on the WebView COM thread
// without QueueMain (avoids flooding the 128-slot buffer and dropping end).
// Main applies via DrainFillClientChromeResize. Returns true if consumed.
func noteFillClientChromeResize(payload string) bool {
	payload = strings.TrimSpace(payload)
	if payload == "" || !strings.Contains(payload, "chrome.resize") {
		return false
	}
	var env fillClientChromeEnvelope
	if err := json.Unmarshal([]byte(payload), &env); err != nil {
		return false
	}
	if env.Type != "gru.bridge" || env.Name != "chrome.resize" {
		return false
	}
	if !webViewFillClientHostsActive() {
		return true
	}
	var p fillClientChromeResizePayload
	if len(env.Payload) > 0 {
		_ = json.Unmarshal(env.Payload, &p)
	}
	phase := strings.ToLower(strings.TrimSpace(p.Phase))
	fillClientChromeMu.Lock()
	switch phase {
	case "start":
		edge := parseFillClientResizeEdge(p.Edge)
		if edge != edgeNone {
			fillClientResize.wantStart = true
			fillClientResize.startEdge = edge
			fillClientResize.startScreenMX = p.ScreenX
			fillClientResize.startScreenMY = p.ScreenY
			fillClientResize.wantEnd = false
		}
	case "move":
		fillClientResize.pendingMX = p.ScreenX
		fillClientResize.pendingMY = p.ScreenY
		fillClientResize.hasPendingMove = true
	case "end", "cancel":
		fillClientResize.pendingMX = p.ScreenX
		fillClientResize.pendingMY = p.ScreenY
		fillClientResize.hasPendingMove = true
		fillClientResize.wantEnd = true
	}
	fillClientChromeMu.Unlock()
	Wake(WakeResize, "fillclient-chrome")
	return true
}

// tryHandleFillClientChromeResize is kept for tests; prefer noteFillClientChromeResize
// from deliverWebViewMessage (COM thread) + DrainFillClientChromeResize (main).
func tryHandleFillClientChromeResize(payload string) bool {
	return noteFillClientChromeResize(payload)
}

func parseFillClientResizeEdge(s string) resizeEdge {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "s":
		return edgeS
	case "e":
		return edgeE
	case "w":
		return edgeW
	case "se":
		return edgeSE
	case "sw":
		return edgeSW
	default:
		return edgeNone
	}
}

func beginFillClientChromeResize(edge resizeEdge, screenMX, screenMY float32) {
	if edge == edgeNone || rl.IsWindowMaximized() {
		return
	}
	fillClientChromeMu.Lock()
	tb := fillClientChromeTitleBar
	fillClientResize.active = true
	fillClientResize.edge = edge
	fillClientResize.startMX = screenMX
	fillClientResize.startMY = screenMY
	fillClientResize.hasPendingMove = false
	fillClientResize.wantStart = false
	if tb != nil {
		tb.seedPos()
		fillClientResize.startWX = tb.trackedX
		fillClientResize.startWY = tb.trackedY
		fillClientResize.startWW = tb.windowW
		fillClientResize.startWH = tb.windowH
		if fillClientResize.startWW < 1 {
			fillClientResize.startWW = int32(rl.GetScreenWidth())
		}
		if fillClientResize.startWH < 1 {
			fillClientResize.startWH = int32(rl.GetScreenHeight())
		}
	} else {
		pos := rl.GetWindowPosition()
		fillClientResize.startWX = int32(pos.X)
		fillClientResize.startWY = int32(pos.Y)
		fillClientResize.startWW = int32(rl.GetScreenWidth())
		fillClientResize.startWH = int32(rl.GetScreenHeight())
	}
	fillClientChromeMu.Unlock()
	SetChromeWindowMoving(true)
	MarkWebViewHostsResize()
}

// DrainFillClientChromeResize applies pending start/move/end on the main goroutine.
// Call once per frame before SyncWebViewHosts.
func DrainFillClientChromeResize() {
	fillClientChromeMu.Lock()
	wantStart := fillClientResize.wantStart
	startEdge := fillClientResize.startEdge
	startMX := fillClientResize.startScreenMX
	startMY := fillClientResize.startScreenMY
	if wantStart {
		fillClientResize.wantStart = false
	}
	hasMove := fillClientResize.hasPendingMove
	mx := fillClientResize.pendingMX
	my := fillClientResize.pendingMY
	if hasMove {
		fillClientResize.hasPendingMove = false
	}
	wantEnd := fillClientResize.wantEnd
	if wantEnd {
		fillClientResize.wantEnd = false
	}
	fillClientChromeMu.Unlock()

	if wantStart {
		beginFillClientChromeResize(startEdge, startMX, startMY)
	}
	if hasMove {
		moveFillClientChromeResize(mx, my)
	}
	if wantEnd {
		endFillClientChromeResize()
	}
}

func moveFillClientChromeResize(screenMX, screenMY float32) {
	fillClientChromeMu.Lock()
	st := fillClientResize
	tb := fillClientChromeTitleBar
	if !st.active || st.edge == edgeNone {
		fillClientChromeMu.Unlock()
		return
	}
	dx := screenMX - st.startMX
	dy := screenMY - st.startMY
	edge := st.edge
	startWX, startWY := st.startWX, st.startWY
	startWW, startWH := st.startWW, st.startWH
	fillClientChromeMu.Unlock()

	x := int(startWX)
	y := int(startWY)
	w := int(startWW)
	h := int(startWH)
	idX := int(dx)
	idY := int(dy)

	shrinkW := func(d int) int {
		if w-d < minWinW {
			d = w - minWinW
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

	switch edge {
	case edgeE:
		w += growW(idX)
	case edgeW:
		d := shrinkW(idX)
		x += d
		w -= d
	case edgeS:
		h += growH(idY)
	case edgeSE:
		w += growW(idX)
		h += growH(idY)
	case edgeSW:
		d := shrinkW(idX)
		x += d
		w -= d
		h += growH(idY)
	default:
		return
	}

	if x != int(startWX) || y != int(startWY) {
		if tb != nil {
			tb.moveWindow(x, y)
		} else {
			rl.SetWindowPosition(x, y)
		}
	}
	rl.SetWindowSize(w, h)
	if tb != nil {
		tb.SetSize(int32(w), int32(h))
	}
	SetChromeWindowMoving(true)
	MarkWebViewHostsResize()
}

func endFillClientChromeResize() {
	fillClientChromeMu.Lock()
	fillClientResize.active = false
	fillClientResize.edge = edgeNone
	fillClientResize.hasPendingMove = false
	fillClientResize.wantStart = false
	tb := fillClientChromeTitleBar
	resizing := false
	if tb != nil {
		resizing = tb.IsResizing() || tb.IsDragging()
	}
	fillClientChromeMu.Unlock()
	if !resizing {
		SetChromeWindowMoving(false)
	}
}
