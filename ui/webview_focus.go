// Package ui (continued) — WebView2 focus handoff with native widgets.

//

// See docs/WEBVIEW2_HOST.md §5.

//

// Contract: passthrough (UpdateWebViewPointerPolicy) is separate from focus

// (RouteScenePointerFocus). Only web panel ↔ native text fields participate in

// focus routing; buttons, modals, directory, and title bar use their own handlers.

package ui



import (

	rl "github.com/gen2brain/raylib-go/raylib"

)



// ReleaseWebKeyboardFocus moves OS keyboard focus out of every live WebView2 host.

func ReleaseWebKeyboardFocus() {

	BlurAllWebViewHosts()

}



// BlurAllWebViewHosts moves keyboard focus out of every live WebView2 surface and

// enables click-through so the next pointer event reaches native widgets.

func BlurAllWebViewHosts() {

	notifyWebViewLostKeyboard()

	webViewHostMu.Lock()

	hosts := append([]WebViewHost(nil), webViewHostAlive...)

	webViewHostMu.Unlock()

	for _, h := range hosts {

		h.Blur()

		if pass, ok := h.(webViewInputPassHost); ok {
			pass.setInputPassthrough(false)
		}

	}

}



// UpdateWebViewPointerPolicy refreshes WS_EX_TRANSPARENT on every live host from

// the current cursor position. Call after SyncWebViewHosts and before focus

// routing so clicks outside the panel reach native widgets (§5).

func UpdateWebViewPointerPolicy(doc *Document) {

	if doc == nil || doc.Root == nil {

		return

	}

	updateWebViewPointerPolicyWalk(doc.Root, doc)

}



func updateWebViewPointerPolicyWalk(n Node, doc *Document) {

	if n == nil {

		return

	}

	if wv, ok := n.(*WebViewPanel); ok {

		wv.syncPointerPolicy(doc)

	}

	for _, ch := range n.Children() {

		updateWebViewPointerPolicyWalk(ch, doc)

	}

}



// AcquireWebFocus clears native doc focus and gives the OS web host keyboard input.

func (wv *WebViewPanel) AcquireWebFocus(doc *Document) {

	if wv == nil || wv.IsHidden() {

		return

	}

	if doc != nil {

		doc.SetFocus(nil)

	}

	if wv.host != nil {
		wv.host.SetFocus()
	}

	setWebViewKeyboardPanel(wv)

}



// RouteScenePointerFocus hands keyboard focus between WebViewPanel and native

// TextInput/TextEditor only. Does not run for title bar, footer, modals, or

// other widgets — those keep their existing click handlers (§5).

func RouteScenePointerFocus(doc *Document) {

	if doc == nil || doc.Root == nil || ScenePointerBlocked() {

		return

	}

	pos, ok := sceneFocusClickPoint()

	if !ok {

		return

	}



	if native := nativeTextFocusAt(doc.Root, pos); native != nil {

		doc.SetFocus(native)

		if PointerClickPending() {

			PointerClickMarkUsed()

		}

		clearFocusHandoffClick()

		return

	}



	if wv := WebViewPanelUnderPoint(doc.Root, pos); wv != nil {

		wv.AcquireWebFocus(doc)

		if PointerClickPending() {

			PointerClickMarkUsed()

		}

		clearFocusHandoffClick()

		return

	}



	// Click outside web panel and native text — release web keyboard focus only.

	if PointerClickPending() {

		ReleaseWebKeyboardFocus()

	}

	clearFocusHandoffClick()

}



// RouteScenePointerFocusAfterPresent runs after Win32 message pump / PresentWebViewHosts

// so WebView2 GotFocus-style OS focus changes are visible the same frame.

func RouteScenePointerFocusAfterPresent(doc *Document) {

	if doc == nil || doc.Root == nil || !focusHandoffPending {

		return

	}

	RouteScenePointerFocus(doc)

}



// sceneFocusClickPoint returns the press position for focus routing after layout.

// Uses the reserved focus handoff click first (immune to pointer latch consumers).

func sceneFocusClickPoint() (rl.Vector2, bool) {

	if pos, ok := PeekFocusHandoffClick(); ok {

		return pos, true

	}

	if PointerClickPending() {

		return PointerClickPosition(), true

	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {

		return rl.GetMousePosition(), true

	}

	return rl.Vector2{}, false

}



func nativeTextFocusAt(root Node, p rl.Vector2) Node {

	if root == nil {

		return nil

	}

	for _, child := range hitTestChildOrder(root.Children()) {

		if child.IsHidden() {

			continue

		}

		if hit := nativeTextFocusAt(child, p); hit != nil {

			return hit

		}

	}

	switch n := root.(type) {

	case *TextInput:

		if !n.IsHidden() && n.IsInteractive() && rl.CheckCollisionPointRec(p, n.Bounds()) {

			return n

		}

	case *TextEditor:

		if !n.IsHidden() && n.IsInteractive() && rl.CheckCollisionPointRec(p, n.Bounds()) {

			return n

		}

	}

	return nil

}



// WebViewPanelUnderPoint returns the deepest WebViewPanel containing p, if any.

func WebViewPanelUnderPoint(root Node, p rl.Vector2) *WebViewPanel {

	if root == nil {

		return nil

	}

	for _, child := range hitTestChildOrder(root.Children()) {

		if child.IsHidden() {

			continue

		}

		if wv := WebViewPanelUnderPoint(child, p); wv != nil {

			return wv

		}

		if wv, ok := child.(*WebViewPanel); ok && !wv.IsHidden() {
			doc := ActiveDocument()
			hit := wv.Bounds()
			if doc != nil {
				if vis := wv.visibleHostRect(doc); vis.Width >= 1 && vis.Height >= 1 {
					hit = vis
				}
			}
			if rl.CheckCollisionPointRec(p, hit) {
				return wv
			}
		}

	}

	return nil

}

