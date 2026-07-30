// WebView flash debugger — GRU_WEBVIEW_DEBUG=1 (GORY_WEBVIEW_DEBUG alias) or Shift+F11 while running.
//
// Logs passthrough, visibility, and present skips when the embedded panel blanks.
package main

import (
	"github.com/ledocorp/gru/ui"
	"fmt"
)

func toggleWebViewDebug() {
	ui.WebViewDebugEnabled = !ui.WebViewDebugEnabled
	fmt.Printf("Gru webview debug: %v (Shift+F11 toggle — passthrough/visible trace)\n", ui.WebViewDebugEnabled)
}
