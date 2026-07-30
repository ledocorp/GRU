//go:build !notepad

package examples

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

func init() { Register(func() Scene { return &webViewFocusScene{} }) }

// webViewFocusScene dogfoods §5 focus handoff: native TextEditor vs embedded WebView2.
type webViewFocusScene struct {
	BaseScene
	editor *ui.TextEditor
}

func (s *webViewFocusScene) Title() string { return "WebView Focus Handoff" }

func (s *webViewFocusScene) Destroy() { ui.DestroyAllWebViewHosts() }

func (s *webViewFocusScene) OnUpdate(doc *ui.Document, _ float32) {
	_ = doc
}

const webViewFocusHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8"/>
<style>
  :root { font-family: system-ui, Segoe UI, sans-serif; color: #0f172a; background: #fff; }
  [data-gru-theme="dark"] { color: #e2e8f0; background: #1e293b; }
  body { margin: 0; padding: 20px; }
  h2 { margin: 0 0 8px; font-size: 18px; }
  p { margin: 0 0 12px; color: #64748b; font-size: 14px; line-height: 1.45; }
  textarea { width: 100%; min-height: 120px; box-sizing: border-box; padding: 10px;
    border: 1px solid #cbd5e1; border-radius: 8px; font: inherit; resize: vertical; }
  #web-focus { margin-top: 10px; font-size: 13px; color: #475569; }
</style>
<script src="https://gru.media/web/gru.js"></script>
</head>
<body>
  <h2>Web side</h2>
  <p>Click here, type, then click the native editor on the left — keys should follow focus.</p>
  <textarea id="web-input" placeholder="Web textarea — type here when focused"></textarea>
  <div id="web-focus">Web focus: unknown</div>
  <script>
    var ta = document.getElementById('web-input');
    var status = document.getElementById('web-focus');
    function post(name, payload) {
      if (!window.gru || !window.gru.postMessage) return;
      window.gru.postMessage(JSON.stringify({type:'gru.bridge',v:1,name:name,payload:payload}));
    }
    ta.addEventListener('focus', function(){ status.textContent = 'Web focus: active'; post('web-focus',{active:true}); });
    ta.addEventListener('blur', function(){ status.textContent = 'Web focus: inactive'; post('web-focus',{active:false}); });
  </script>
</body>
</html>`

func (s *webViewFocusScene) Build(doc *ui.Document) {
	mount := MountAppShellRoot(doc, "webview-focus")
	shell := mount.Shell

	hdr := ui.NewHeader("webview-focus-hdr", "Focus handoff",
		"Click the native editor or web panel — keyboard follows focus.", 0, 0, 0, 0)
	hdr.SetStyle("header")
	shell.AddChild(hdr)

	row := ui.NewContainer("webview-focus-row", 0, 0, 0, 0)
	row.LayoutType = ui.LayoutFlex
	row.FlexDirection = ui.FlexRow
	row.Gap = 12
	row.SetFlexGrow(1)

	nativeCol := ui.NewPanel("webview-focus-native", "Text editor", 0, 0, 0, 0)
	nativeCol.SetFlexGrow(1)
	nativeCol.SetStyle("card")

	s.editor = ui.NewTextEditor("webview-focus-editor",
		"Native editor — click here and type.\n\nSwitch to the web panel on the right; this field should blur and stop receiving keys.", 0, 0, 0, 0)
	s.editor.SetFlexGrow(1)
	s.editor.WordWrap = true
	nativeCol.AddChild(s.editor)

	webStatus := ui.NewSignal("Click web panel to focus")
	statusLbl := ui.NewLabel("webview-focus-status", webStatus.Get(), 0, 0, 0, 0)
	statusLbl.Text = webStatus
	statusLbl.SetStyle("form-value")
	statusLbl.Wrap = false
	nativeCol.AddChild(statusLbl)

	modalBtn := ui.NewButton("webview-focus-modal", "Open modal (hides web host)", 0, 0, 0, 36)
	modalBtn.OnClick = func() {
		body := ui.NewLabel("webview-focus-modal-body", "Modal open — web host should be hidden until dismissed.", 0, 0, 0, 0)
		body.Wrap = true
		ui.ShowModal("Focus test", body, []ui.ModalButton{
			{Label: "Close", Style: "primary", Action: ui.CloseModal},
		})
	}
	nativeCol.AddChild(modalBtn)

	wv := ui.NewWebViewPanel("webview-focus-web", "", 0, 0, 0, 0)
	wv.SetFlexGrow(1)
	wv.HTML = webViewFocusHTML
	wv.OnMessage = func(payload string) {
		switch {
		case strings.Contains(payload, `"name":"web-focus"`) && strings.Contains(payload, `"active":true`):
			webStatus.Set("Web focus: active (from page)")
		case strings.Contains(payload, `"name":"web-focus"`) && strings.Contains(payload, `"active":false`):
			webStatus.Set("Web focus: inactive")
		}
	}

	row.AddChild(nativeCol)
	row.AddChild(wv)
	shell.AddChild(row)
	FinishShellMount(doc)

	doc.SetFocus(s.editor)
}
