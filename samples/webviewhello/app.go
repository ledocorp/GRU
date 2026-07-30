// Package webviewhello is the WebView shell sample — native chrome + HTML body.
//
// Layout matches examples/webview_full_demo.go (§13.2): keep Document.Root,
// add a transparent flex shell child, WebViewPanel with FillClient=true.
// Build with -tags webview2 on Windows for a live WebView2 host.
package webviewhello

import (
	"github.com/ledocorp/gru/ui"
)

// Same starter HTML spirit as examples/webview_full_demo.go (no bridge required).
const starterHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width, initial-scale=1"/>
<title>Hello WebView</title>
<style>
  :root { font-family: system-ui, Segoe UI, sans-serif; color: #0f172a; background: #f8fafc; }
  body { margin: 0; padding: 32px 28px; }
  h1 { margin: 0 0 8px; font-size: 28px; font-weight: 650; }
  p { margin: 0; color: #64748b; line-height: 1.55; max-width: 36rem; }
  code { font-size: 0.92em; background: #e2e8f0; padding: 0.1em 0.35em; border-radius: 4px; }
</style>
</head>
<body>
  <h1>Hello, WebView</h1>
  <p>Native Gru chrome (window + title bar). Everything below is HTML.
  Run with <code>-tags webview2</code> on Windows for a live host.</p>
</body>
</html>`

// Build mounts a full-client WebView under borderless chrome (MountAppShellRoot shape).
func Build(doc *ui.Document) {
	// Mirror examples.MountAppShellRoot — do NOT replace doc.Root.
	w, h := float32(doc.Width), float32(doc.Height)
	shell := ui.NewContainer("wv-hello-shell", 0, 0, w, h)
	shell.LayoutType = ui.LayoutFlex
	shell.FlexDirection = ui.FlexColumn
	shell.Gap = 0
	shell.SetStyle("transparent")
	shell.SetFlexGrow(1)
	doc.Root.AddChild(shell)

	wv := ui.NewWebViewPanel("wv-hello", "", 0, 0, 0, 0)
	wv.SetFlexGrow(1)
	wv.FillClient = true
	wv.HTML = starterHTML
	shell.AddChild(wv)

	doc.Root.MarkDirty()
}
