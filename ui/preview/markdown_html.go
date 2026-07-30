package preview

import (
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/ledocorp/gru/ui"
)

// htmlFragmentToSpans turns a small HTML fragment into RichText spans (bold/italic/etc.).
// Goldmark emits HTML blocks as raw tags; this renders them like formatted markdown.
func htmlFragmentToSpans(fragment string, flags inlineFlags) []ui.TextSpan {
	fragment = strings.TrimSpace(fragment)
	if fragment == "" {
		return []ui.TextSpan{{Text: ""}}
	}
	if !strings.Contains(fragment, "<") {
		return []ui.TextSpan{applyFlags(ui.TextSpan{Text: fragment}, flags)}
	}
	doc, err := html.Parse(strings.NewReader("<!DOCTYPE html><html><body>" + fragment + "</body></html>"))
	if err != nil {
		return []ui.TextSpan{{Text: fragment}}
	}
	body := doc
	if n := findHTMLNode(doc, atom.Body); n != nil {
		body = n
	}
	var spans []ui.TextSpan
	walkHTML(body, flags, &spans)
	if len(spans) == 0 {
		return []ui.TextSpan{{Text: fragment}}
	}
	return mergeSpans(spans)
}

func findHTMLNode(n *html.Node, a atom.Atom) *html.Node {
	if n == nil {
		return nil
	}
	if n.Type == html.ElementNode && n.DataAtom == a {
		return n
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if found := findHTMLNode(c, a); found != nil {
			return found
		}
	}
	return nil
}

func walkHTML(n *html.Node, flags inlineFlags, out *[]ui.TextSpan) {
	if n == nil {
		return
	}
	switch n.Type {
	case html.TextNode:
		text := strings.ReplaceAll(n.Data, "\r\n", "\n")
		if text == "" {
			return
		}
		*out = append(*out, applyFlags(ui.TextSpan{Text: text}, flags))
	case html.ElementNode:
		f := flags
		switch n.DataAtom {
		case atom.Strong, atom.B:
			f.bold = true
		case atom.Em, atom.I:
			f.italic = true
		case atom.Code, atom.Samp:
			for c := n.FirstChild; c != nil; c = c.NextSibling {
				if c.Type == html.TextNode && c.Data != "" {
					*out = append(*out, applyFlags(ui.TextSpan{Text: c.Data, Variant: "code"}, f))
				} else {
					walkHTML(c, f, out)
				}
			}
			return
		case atom.Br:
			*out = append(*out, ui.TextSpan{Text: "\n"})
			return
		case atom.H1, atom.H2, atom.H3, atom.H4, atom.H5, atom.H6:
			f.bold = true
			fallthrough
		case atom.P, atom.Div:
			if len(*out) > 0 && (*out)[len(*out)-1].Text != "\n" {
				*out = append(*out, ui.TextSpan{Text: "\n"})
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkHTML(c, f, out)
		}
		if n.DataAtom == atom.P || n.DataAtom == atom.Div {
			*out = append(*out, ui.TextSpan{Text: "\n"})
		}
	default:
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walkHTML(c, flags, out)
		}
	}
}
