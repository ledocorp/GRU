// Package ui (continued) — DocumentSpec toolbar / ribbon blocks (Sprint G).
package ui

import "fmt"

// buildDocToolbar compiles a flat command-bar toolbar from DocBlock children.
// Groups use type "toolbarGroup" (or "group"); items are nested children.
// Toolbar-level props: overflow (bool), overflowKind ("scroll"|"menu"), itemGap, ribbon (bool, not yet).
func buildDocToolbar(block DocBlock, ctx *BuildContext, path string) (Node, error) {
	id := block.ID
	if id == "" {
		id = "toolbar"
	}
	h := block.Height
	tb := NewToolbar(id, 0, 0, docBlockWidth(block), h)
	tb.Overflow = docBlockBool(block, "overflow", false)
	switch docBlockStringProp(block, "overflowKind") {
	case "menu":
		tb.OverflowKind = ToolbarOverflowMenu
	case "scroll":
		tb.OverflowKind = ToolbarOverflowScroll
	}
	if gap := docBlockFloat(block, "itemGap", 0); gap > 0 {
		tb.ItemGap = gap
	} else if gap := docBlockGap(block, 0); gap > 0 {
		tb.ItemGap = gap
	}
	if docBlockBool(block, "ribbon", false) {
		return nil, docBlockError(block, path, "ribbon toolbar JSON is not implemented yet; use Go AddRibbonTab or flat toolbar")
	}
	hasGroup := false
	for i := range block.Children {
		if isDocToolbarGroup(block.Children[i].Type) {
			hasGroup = true
			break
		}
	}
	if !hasGroup && len(block.Children) > 0 {
		tb.AddGroup("main", "")
		for i := range block.Children {
			if err := addDocToolbarItem(tb, "main", block.Children[i], ctx, fmt.Sprintf("%s.children[%d]", path, i)); err != nil {
				return nil, err
			}
		}
	} else {
		for i := range block.Children {
			child := block.Children[i]
			childPath := fmt.Sprintf("%s.children[%d]", path, i)
			if isDocToolbarGroup(child.Type) {
				gid := child.ID
				if gid == "" {
					gid = fmt.Sprintf("group-%d", i)
				}
				glabel := child.Title
				if glabel == "" {
					glabel = child.Label
				}
				tb.AddGroup(gid, glabel)
				for j := range child.Children {
					if err := addDocToolbarItem(tb, gid, child.Children[j], ctx, fmt.Sprintf("%s.children[%d]", childPath, j)); err != nil {
						return nil, err
					}
				}
				continue
			}
			return nil, docBlockError(child, childPath, "toolbar child must be toolbarGroup or item inside a group")
		}
	}
	applyDocStyle(&tb.Element, block)
	applyDocLayout(&tb.Element, block)
	registerDocControlValue(ctx, id, func() any { return tb })
	return tb, nil
}

func isDocToolbarGroup(typ string) bool {
	switch typ {
	case "toolbarGroup", "group":
		return true
	default:
		return false
	}
}

func addDocToolbarItem(tb *Toolbar, groupID string, item DocBlock, ctx *BuildContext, path string) error {
	typ := item.Type
	if typ == "" {
		return docBlockError(item, path, "toolbar item type is required")
	}
	switch typ {
	case "separator", "toolbarSeparator", "divider":
		tb.AddSeparator(groupID, item.ID)
	case "button":
		label := item.Text
		if label == "" {
			label = item.Label
		}
		var fn func()
		if action := docBlockOnClick(item); action != "" {
			if ctx == nil || ctx.Actions == nil || ctx.Actions[action] == nil {
				return docBlockError(item, path, "unknown onClick action %q", action)
			}
			fn = ctx.Actions[action]
		}
		itemID := item.ID
		if itemID == "" {
			itemID = "btn"
		}
		tb.AddButton(groupID, itemID, label, fn)
	case "toggleLabel", "toggle-label":
		label := item.Label
		if label == "" {
			label = item.Text
		}
		sig := docBindBool(ctx, docBlockBindKey(item), item.Checked)
		itemID := item.ID
		if itemID == "" {
			itemID = "toggle-label"
		}
		tb.AddToggleLabel(groupID, itemID, label, sig)
		registerDocControlValue(ctx, itemID, func() any { return sig.Get() })
	case "menu", "menuButton":
		face := item.Label
		if face == "" {
			face = item.Title
		}
		if face == "" {
			face = "Menu"
		}
		opts := docBlockOptions(item)
		if len(opts) == 0 {
			return docBlockError(item, path, "menu item requires options")
		}
		itemID := item.ID
		if itemID == "" {
			itemID = "menu"
		}
		var onSelect func(int)
		if action := docBlockStringProp(item, "onSelect"); action != "" {
			if ctx == nil || ctx.Actions == nil || ctx.Actions[action] == nil {
				return docBlockError(item, path, "unknown onSelect action %q", action)
			}
			onSelect = func(int) { ctx.Actions[action]() }
		}
		tb.AddMenuButton(groupID, itemID, face, opts, onSelect)
	case "wordToggle", "word-toggle":
		label := item.Label
		if label == "" {
			label = item.Text
		}
		sig := docBindBool(ctx, docBlockBindKey(item), item.Checked)
		itemID := item.ID
		if itemID == "" {
			itemID = "word-toggle"
		}
		tb.AddWordToggle(groupID, itemID, label, sig)
		registerDocControlValue(ctx, itemID, func() any { return sig.Get() })
	case "spinBox", "spin-box":
		minF, maxF := docBlockRange(item, 0, 100)
		minV := float64(minF)
		maxV := float64(maxF)
		step := float64(docBlockFloat(item, "step", 1))
		if step <= 0 {
			step = 1
		}
		sig := docBindFloat64(ctx, docBlockBindKey(item), docBlockFloat64Value(item, minV))
		itemID := item.ID
		if itemID == "" {
			itemID = "spin"
		}
		tb.AddSpinBox(groupID, itemID, sig, minV, maxV, step)
		registerDocControlValue(ctx, itemID, func() any { return sig.Get() })
	case "icon", "iconButton", "icon-button":
		iconName := docBlockStringProp(item, "icon")
		if iconName == "" {
			iconName = docBlockStringProp(item, "phosphor")
		}
		if iconName == "" {
			return docBlockError(item, path, "icon item requires props.icon (Remix / Phosphor name)")
		}
		label := item.Text
		tooltip := docBlockStringProp(item, "tooltip")
		var fn func()
		if action := docBlockOnClick(item); action != "" {
			if ctx == nil || ctx.Actions == nil || ctx.Actions[action] == nil {
				return docBlockError(item, path, "unknown onClick action %q", action)
			}
			fn = ctx.Actions[action]
		}
		itemID := item.ID
		if itemID == "" {
			itemID = "icon"
		}
		tb.AddPhosphorIconButton(groupID, itemID, iconName, label, tooltip, fn)
	default:
		return docBlockError(item, path, "unsupported toolbar item type %q", typ)
	}
	return nil
}

func docBlockBindKey(block DocBlock) string {
	if k := docBlockStringProp(block, "bind"); k != "" {
		return k
	}
	if block.Props != nil {
		if k, _ := block.Props["bindingKey"].(string); k != "" {
			return k
		}
	}
	return block.ID
}

func docBindBool(ctx *BuildContext, key string, initial bool) *Signal[bool] {
	if key == "" {
		return NewSignal(initial)
	}
	if ctx != nil && ctx.BoolSignals != nil {
		if s, ok := ctx.BoolSignals[key]; ok {
			return s
		}
	}
	s := NewSignal(initial)
	if ctx != nil {
		if ctx.BoolSignals == nil {
			ctx.BoolSignals = make(map[string]*Signal[bool])
		}
		ctx.BoolSignals[key] = s
	}
	return s
}

func docBindFloat64(ctx *BuildContext, key string, initial float64) *Signal[float64] {
	if key == "" {
		return NewSignal(initial)
	}
	if ctx != nil && ctx.Float64Signals != nil {
		if s, ok := ctx.Float64Signals[key]; ok {
			return s
		}
	}
	s := NewSignal(initial)
	if ctx != nil {
		if ctx.Float64Signals == nil {
			ctx.Float64Signals = make(map[string]*Signal[float64])
		}
		ctx.Float64Signals[key] = s
	}
	return s
}

func docBlockFloat64Value(block DocBlock, fallback float64) float64 {
	if block.Value == nil {
		return fallback
	}
	if !docBlockValueIsNumber(block.Value) {
		return fallback
	}
	return float64(docBlockNumericValue(block, float32(fallback)))
}
