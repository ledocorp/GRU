package preview

import (
	"strings"

	"github.com/yuin/goldmark/ast"

	"github.com/ledocorp/gru/ui"
)

func isHTTPURL(s string) bool {
	s = strings.TrimSpace(s)
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func paragraphImage(p *ast.Paragraph) *ast.Image {
	if p == nil || p.ChildCount() != 1 {
		return nil
	}
	img, ok := p.FirstChild().(*ast.Image)
	if !ok {
		return nil
	}
	return img
}

func buildImageBlock(ctx *mdBuildCtx, id string, img *ast.Image, source []byte) ui.Node {
	url := strings.TrimSpace(string(img.Destination))
	alt := strings.TrimSpace(plainInline(inlineSpans(img, source, inlineFlags{}, ctx)))
	if alt == "" {
		alt = "Image"
	}
	title := string(img.Title)

	// Flat container (not Card) — avoids raised-surface shadow/fill bleeding over
	// sibling blocks while the preview lane scrolls.
	wrap := ui.NewContainer(id, 0, 0, 0, 0)
	wrap.SetStyle("preview-image-wrap")
	wrap.LayoutType = ui.LayoutFlex
	wrap.FlexDirection = ui.FlexColumn
	wrap.Gap = 8
	wrap.AutoHeight = true

	const frameH = float32(260)
	frame := ui.NewContainer(id+"-frame", 0, 0, 0, frameH)
	frame.SetStyle("preview-image-frame")
	frame.LayoutType = ui.LayoutFlex
	frame.FlexDirection = ui.FlexColumn
	frame.AutoHeight = false

	localPath := resolvePreviewImagePath(previewImagePath(url))
	wimg := ui.NewImage(id+"-img", localPath, 0, 0, 0, frameH)
	wimg.FitMode = ui.FitContain
	wimg.SetFlexGrow(1)
	frame.AddChild(wimg)
	wrap.AddChild(frame)

	caption := []ui.TextSpan{{Text: alt, Variant: "muted"}}
	if title != "" {
		caption = append(caption, ui.TextSpan{Text: " — " + title, Variant: "muted"})
	}
	cap := ui.NewRichText(id+"-cap", caption, 0, 0, 0, 0)
	cap.SetStyle("richtext-preview")
	cap.AutoHeight = true
	ctx.wireRichText(cap)
	wrap.AddChild(cap)

	return wrap
}
