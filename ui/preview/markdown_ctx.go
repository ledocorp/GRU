package preview

import (
	"strings"
	"unicode"

	"github.com/ledocorp/gru/ui"
)

// mdBuildCtx carries link wiring and heading anchors for one markdown build.
type mdBuildCtx struct {
	source             []byte
	idPrefix           string
	anchors            map[string]ui.Node
	footnoteIndexToRef map[int]string
	onLink             func(string)
	doc                *ui.Document // async math render in preview
}

func newBuildCtx(idPrefix string, anchors map[string]ui.Node, onLink func(string)) *mdBuildCtx {
	if anchors == nil {
		anchors = make(map[string]ui.Node)
	}
	return &mdBuildCtx{
		idPrefix:           idPrefix,
		anchors:            anchors,
		footnoteIndexToRef: make(map[int]string),
		onLink:             onLink,
	}
}

func (c *mdBuildCtx) footnoteRef(idx int) string {
	if c == nil {
		return itoa(idx + 1)
	}
	if ref, ok := c.footnoteIndexToRef[idx]; ok && ref != "" {
		return ref
	}
	return itoa(idx)
}

func (c *mdBuildCtx) wireRichText(rt *ui.RichText) {
	if rt == nil {
		return
	}
	if c.onLink != nil {
		rt.OnLinkClick = c.onLink
	}
}

// registerAnchor records heading and other non-footnote-return targets.
func (c *mdBuildCtx) registerAnchor(slug string, node ui.Node) {
	if c == nil || c.anchors == nil || slug == "" || node == nil {
		return
	}
	c.anchors[slug] = node
}

// MarkdownLinkHandler opens http(s) links and scrolls to #anchors (headings, footnotes, return).
func MarkdownLinkHandler(scroll *ui.Viewport, lane ui.Node, anchors map[string]ui.Node) func(string) {
	return func(link string) {
		link = strings.TrimSpace(link)
		if link == "" {
			return
		}
		if strings.HasPrefix(link, "#") {
			slug := markdownAnchorSlug(strings.TrimPrefix(link, "#"))
			if target := resolvePreviewJumpTarget(lane, anchors, slug); target != nil {
				target.Show()
				if scroll != nil && lane != nil {
					lane.Layout()
					scroll.Layout()
					scroll.ScrollToShowNode(target, lane)
					scroll.MarkDirty()
				}
			}
			return
		}
		if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
			_ = ui.OpenBrowser(link)
		}
	}
}

// markdownAnchorSlug normalizes heading text or fragment ids (e.g. "1-headings").
func markdownAnchorSlug(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	var b strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			prevDash = false
			continue
		}
		if !prevDash && b.Len() > 0 {
			b.WriteByte('-')
			prevDash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

func headingAnchorID(level int, text string) string {
	slug := markdownAnchorSlug(text)
	if slug == "" {
		return ""
	}
	_ = level
	return slug
}

