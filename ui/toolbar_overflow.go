// Package ui (continued) — toolbar overflow menu activation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// ToolbarOverflowKind selects how a too-narrow toolbar exposes hidden items.
type ToolbarOverflowKind int

const (
	// ToolbarOverflowScroll keeps all items live and pans the band horizontally (default).
	ToolbarOverflowScroll ToolbarOverflowKind = iota
	// ToolbarOverflowMenu hides trailing items behind a "…" text menu (simple commands only).
	ToolbarOverflowMenu
)

func (tb *Toolbar) overflowKind() ToolbarOverflowKind {
	if !tb.Overflow {
		return ToolbarOverflowScroll
	}
	if tb.OverflowKind == ToolbarOverflowMenu {
		return ToolbarOverflowMenu
	}
	return ToolbarOverflowScroll
}

func (tb *Toolbar) usesOverflowMenu() bool {
	return tb.Overflow && tb.overflowKind() == ToolbarOverflowMenu
}

func (tb *Toolbar) usesHorizontalScroll() bool {
	return tb.Overflow && tb.overflowKind() == ToolbarOverflowScroll
}

// activateOverflowItem runs the same outcome as clicking the control (menu mode popup).
func (tb *Toolbar) activateOverflowItem(item *ToolbarItem, rowRect rl.Rectangle) {
	if item == nil || item.widget == nil {
		return
	}
	switch w := item.widget.(type) {
	case *Button:
		if w.ToggleBinding != nil {
			w.ToggleBinding.Set(!w.ToggleBinding.Get())
			w.MarkDrawDirty()
		} else if w.OnClick != nil {
			w.OnClick()
		}
	case *IconButton:
		if w.Checked != nil {
			w.Checked.Set(!w.Checked.Get())
			w.MarkDrawDirty()
		}
		if w.OnClick != nil {
			w.OnClick()
		}
	case *Toggle:
		w.toggle()
	case *Dropdown:
		w.OpenFromToolbarOverflow(rowRect)
		item.hidden = false
		tb.overflowOpen = false
	case *ToolbarWordToggle:
		if w.Value != nil {
			w.Value.Set(!w.Value.Get())
			w.MarkDrawDirty()
		}
	case *ToolbarSpinBox:
		w.nudge(1)
	default:
	}
}
