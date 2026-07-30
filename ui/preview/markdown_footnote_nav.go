package preview

import (
	"strings"

	"github.com/ledocorp/gru/ui"
)

// walkPreviewNodes visits nodes in document order; stop when fn returns true.
func walkPreviewNodes(root ui.Node, fn func(ui.Node) bool) {
	var visit func(ui.Node) bool
	visit = func(n ui.Node) bool {
		if n == nil {
			return false
		}
		if fn(n) {
			return true
		}
		for _, ch := range n.Children() {
			if visit(ch) {
				return true
			}
		}
		return false
	}
	visit(root)
}

// findFirstFootnoteRef returns the first inline [^ref] RichText in the preview tree.
// Navigation is owned by the widget tree, not Goldmark anchor registration.
func findFirstFootnoteRef(root ui.Node, ref string) ui.Node {
	ref = strings.TrimSpace(ref)
	if ref == "" || root == nil {
		return nil
	}
	wantLink := "#" + footnoteDefAnchor(ref)
	var found ui.Node
	walkPreviewNodes(root, func(n ui.Node) bool {
		rt, ok := n.(*ui.RichText)
		if !ok {
			return false
		}
		for _, sp := range rt.Spans {
			if sp.Variant == "footnote-ref" && sp.Link == wantLink {
				found = rt
				return true
			}
		}
		return false
	})
	return found
}

// findFootnoteDef returns the footnote definition RichText for ref (bottom [ref] line).
func findFootnoteDef(root ui.Node, ref string) ui.Node {
	ref = strings.TrimSpace(ref)
	if ref == "" || root == nil {
		return nil
	}
	label := footnoteBracketLabel(ref)
	var found ui.Node
	walkPreviewNodes(root, func(n ui.Node) bool {
		rt, ok := n.(*ui.RichText)
		if !ok || !footnoteDefRichText(rt, label) {
			return false
		}
		found = rt
		return true
	})
	return found
}

func firstRichText(n ui.Node) *ui.RichText {
	if rt, ok := n.(*ui.RichText); ok {
		return rt
	}
	for _, ch := range n.Children() {
		if rt := firstRichText(ch); rt != nil {
			return rt
		}
	}
	return nil
}

func footnoteDefRichText(rt *ui.RichText, label string) bool {
	for _, sp := range rt.Spans {
		if sp.Variant != "footnote-ref" || sp.Link != "" {
			continue
		}
		if strings.HasPrefix(sp.Text, label) {
			return true
		}
	}
	return false
}

// resolvePreviewJumpTarget finds a scroll target for #slug (headings, footnotes, return).
func resolvePreviewJumpTarget(lane ui.Node, anchors map[string]ui.Node, slug string) ui.Node {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return nil
	}
	if strings.HasPrefix(slug, "fn-ref-") {
		return findFirstFootnoteRef(lane, strings.TrimPrefix(slug, "fn-ref-"))
	}
	if strings.HasPrefix(slug, "fn-") {
		if n := findFootnoteDef(lane, strings.TrimPrefix(slug, "fn-")); n != nil {
			return n
		}
	}
	if anchors != nil {
		return anchors[slug]
	}
	return nil
}
