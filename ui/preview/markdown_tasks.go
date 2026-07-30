package preview

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	extast "github.com/yuin/goldmark/extension/ast"

	"github.com/ledocorp/gru/ui"
)

// parseTaskPrefix detects GFM task markers at the start of list item text: [ ], [x], [X].
// Preview owns task UI; Goldmark TaskCheckBox is not used for display.
func parseTaskPrefix(s string) (task bool, checked bool, body string) {
	s = strings.TrimSpace(s)
	if len(s) < 3 || s[0] != '[' {
		return false, false, s
	}
	close := strings.IndexByte(s, ']')
	if close < 2 {
		return false, false, s
	}
	inside := strings.TrimSpace(s[1:close])
	if inside != "" && !strings.EqualFold(inside, "x") {
		return false, false, s
	}
	checked = strings.EqualFold(inside, "x")
	body = strings.TrimSpace(s[close+1:])
	return true, checked, body
}

// taskItemFromListItem detects tasks from source [x]/[ ] when present, else from
// Goldmark's TaskCheckBox node (detection only — display is always our Remix icons).
func taskItemFromListItem(li *ast.ListItem, source []byte) (task bool, checked bool) {
	if task, checked, _ := parseTaskPrefix(listItemPlainText(li, source)); task {
		return true, checked
	}
	_ = ast.Walk(li, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if cb, ok := n.(*extast.TaskCheckBox); ok {
			task, checked = true, cb.IsChecked
			return ast.WalkStop, nil
		}
		return ast.WalkContinue, nil
	})
	return task, checked
}

func listItemPlainText(li *ast.ListItem, source []byte) string {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		if p, ok := c.(*ast.Paragraph); ok {
			return blockPlainText(p, source)
		}
	}
	return ""
}

func blockPlainText(n ast.Node, source []byte) string {
	var b strings.Builder
	_ = ast.Walk(n, func(nn ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := nn.(*ast.Text); ok {
			b.WriteString(unescapeMarkdownPunct(string(t.Value(source))))
			if t.HardLineBreak() {
				b.WriteByte('\n')
			} else if t.SoftLineBreak() {
				b.WriteByte('\n') // keep $$ display-math line structure
			}
		}
		if s, ok := nn.(*ast.String); ok {
			b.WriteString(unescapeMarkdownPunct(string(s.Value)))
		}
		return ast.WalkContinue, nil
	})
	return b.String()
}

func listItemBodySpans(li *ast.ListItem, source []byte, ctx *mdBuildCtx) []ui.TextSpan {
	for c := li.FirstChild(); c != nil; c = c.NextSibling() {
		switch c.(type) {
		case *ast.List, *extast.TaskCheckBox:
			continue
		default:
			return inlineSpans(c, source, inlineFlags{}, ctx)
		}
	}
	return nil
}

func taskListMarker(id string, checked bool) *ui.Icon {
	name := ui.PhosphorCheckboxBlank
	if checked {
		name = ui.PhosphorCheckbox
	}
	ic := ui.NewIcon(id, name, ui.PhosphorRegular, 0, 0, 22, 22)
	ic.SetStyle("preview-task-icon")
	ui.Phosphor.EnsureLoaded(name, ui.PhosphorRegular)
	return ic
}
