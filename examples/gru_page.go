package examples

import "github.com/ledocorp/gru/ui"

// mountGRUPageShell mounts the flex page shell with a sticky header and compiled spec body.
func mountGRUPageShell(doc *ui.Document, shellID, hdrTitle, hdrSubtitle string, body ui.Node) *ui.Viewport {
	page := MountAppPage(doc, shellID, hdrTitle, hdrSubtitle)
	page.Body.Gap = 12
	if root, ok := body.(*ui.Container); ok {
		root.SetFlexGrow(0)
	}
	page.Body.AddChild(body)
	page.Body.MarkDirty()
	page.Frame.MarkDirty()
	doc.Root.MarkDirty()
	return page.Body
}

