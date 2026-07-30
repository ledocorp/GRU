// Package ui (continued) — WebView2 OS keyboard focus tracking (§5).

package ui



// webViewKeyboardPanel is the panel that currently owns OS keyboard input, if any.

var webViewKeyboardPanel *WebViewPanel



// WebViewHostHoldsKeyboard reports whether a live WebView2 host has OS keyboard focus.

func WebViewHostHoldsKeyboard() bool {

	return webViewKeyboardPanel != nil

}



func setWebViewKeyboardPanel(wv *WebViewPanel) {

	webViewKeyboardPanel = wv

}



func clearWebViewKeyboardPanel() {

	webViewKeyboardPanel = nil

}



// notifyWebViewGotKeyboard clears native doc focus when the web host receives OS keyboard input.

func notifyWebViewGotKeyboard(wv *WebViewPanel) {

	if wv == nil {

		return

	}

	setWebViewKeyboardPanel(wv)

	if doc := ActiveDocument(); doc != nil && doc.Focused != nil {

		doc.SetFocus(nil)

	}

}



func notifyWebViewLostKeyboard() {

	clearWebViewKeyboardPanel()

}

