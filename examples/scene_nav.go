package examples

import "github.com/ledocorp/gru/ui"

// NavigateToScene is set by main.go to switch demo scenes by Title().
// Returns true when the scene exists (even if already active).
var NavigateToScene func(title string) bool

// demoLinkToScene maps demo:// URIs to registered scene titles.
var demoLinkToScene = map[string]string{
	"demo://settings":  "Settings (DocumentSpec)",
	"demo://dashboard": "Dashboard (.gru)",
	"demo://docs":      "Docs (.gru)",
	"demo://authoring": "Authoring (.gru)",
	"demo://gallery":   "Gallery (.gru)",
	"demo://signin":    "Sign in (.gru)",
}

// NavigateFromDemoLink switches scenes for known demo:// page links.
func NavigateFromDemoLink(link string) bool {
	if NavigateToScene == nil {
		return false
	}
	title, ok := demoLinkToScene[link]
	if !ok {
		return false
	}
	return NavigateToScene(title)
}

// HandleDemoLink navigates when possible; otherwise writes a status message.
func HandleDemoLink(link string, status *ui.RichText) {
	if NavigateFromDemoLink(link) {
		return
	}
	if status == nil {
		return
	}
	status.SetSpans([]ui.TextSpan{{Text: docsLinkMessage(link), Variant: "muted"}})
}

func docsLinkMessage(link string) string {
	switch link {
	case "demo://authoring":
		return "Tab to Authoring (.gru) scene"
	case "demo://fixture-link":
		return "Tab to Document + Theme Foundation → Stability Fixture panel"
	default:
		return "Link: " + link
	}
}
