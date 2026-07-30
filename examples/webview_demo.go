//go:build !notepad

package examples

import (
	"time"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &webViewDemoScene{} }) }

// webViewDemoScene — header + navigation rail + embedded WebViewPanel (hybrid UI).
type webViewDemoScene struct {
	BaseScene
}

func (s *webViewDemoScene) Title() string { return "WebView Module Demo" }

func (s *webViewDemoScene) OnUpdate(_ *ui.Document, _ float32) {}

func (s *webViewDemoScene) Destroy() { ui.DestroyAllWebViewHosts() }

const webViewModuleHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<style>
  :root { font-family: system-ui, Segoe UI, sans-serif; color: #0f172a; background: #fff; margin: 0; }
  [data-gru-theme="dark"] { color: #e2e8f0; background: #0f172a; }
  body { margin: 0; padding: 24px; }
  h1 { margin: 0 0 8px; font-size: 24px; }
  p { color: #64748b; line-height: 1.5; }
</style>
</head>
<body>
  <h1>Web module</h1>
  <p>Navigation rail on the left; web content fills the rest. Bridge v0 via <code>window.gru.postMessage</code>.</p>
</body>
</html>`

func (s *webViewDemoScene) Build(doc *ui.Document) {
	mount := MountAppShellRoot(doc, "webview")
	shell := mount.Shell

	hdr := ui.NewHeader("webview-hdr", "Web module", "Rail + embedded web panel", 0, 0, 0, 0)
	hdr.SetStyle("header")
	shell.AddChild(hdr)

	row := ui.NewContainer("webview-row", 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 0
	row.SetFlexGrow(1)

	PreloadScenePhosphor(doc)
	rail := ui.NewNavigationRail("webview-rail", []ui.BottomNavItem{
		{Phosphor: ui.PhosphorHouse, Label: "Home"},
		{Phosphor: ui.PhosphorEnvelope, Label: "Chat"},
		{Phosphor: ui.PhosphorGear, Label: "Settings"},
	}, ui.NewSignal(0), 0, 0, 0, 0)

	wv := ui.NewWebViewPanel("webview-module", "", 0, 0, 0, 0)
	wv.SetFlexGrow(1)
	wv.HTML = webViewModuleHTML
	wv.OnMessage = func(payload string) {
		_ = wv.EmitBridge("toast", map[string]string{"text": "Go: " + payload, "level": "info"})
		ui.ShowToast("Bridge: "+payload, ui.ToastInfo, 2*time.Second)
	}

	row.AddChild(rail)
	row.AddChild(wv)
	shell.AddChild(row)
	FinishShellMount(doc)
}
