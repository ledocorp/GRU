// Package ui (continued) — DocumentSpec surface capabilities (Phase C6).
package ui

import (
	"fmt"
	"strings"
)

// DocBlockCapabilities declares egui-style shell features for section, card, and surface blocks.
// Omitted fields leave the surface at its type defaults (panel title bar, card inset header, etc.).
//
// Example:
//
//	{
//	  "type": "card",
//	  "title": "Filters",
//	  "capabilities": {
//	    "collapsible": true,
//	    "collapsed": false,
//	    "vScroll": true
//	  },
//	  "children": [ … ]
//	}
type DocBlockCapabilities struct {
	TitleBar      *bool    `json:"titleBar"`
	Collapsible   *bool    `json:"collapsible"`
	Closable      *bool    `json:"closable"`
	Movable       *bool    `json:"movable"`
	Resizable     *bool    `json:"resizable"`
	Constrain     *bool    `json:"constrain"`
	HScroll       *bool    `json:"hScroll"`
	VScroll       *bool    `json:"vScroll"`
	Anchored      *bool    `json:"anchored"`
	Anchor        string   `json:"anchor"`    // top-left, center, bottom-right, …
	DragMode      string   `json:"dragMode"`  // off, touch, titleBar, anywhere
	CloseOnEscape *bool    `json:"closeOnEscape"`
	MinWidth      float32  `json:"minWidth"`
	FloatX        *float32 `json:"floatX"`
	FloatY        *float32 `json:"floatY"`
	Collapsed     *bool    `json:"collapsed"` // initial state when collapsible
	OnDismiss     string   `json:"onDismiss"` // BuildContext action name
}

type surfaceDocKind int

const (
	surfaceKindAuto surfaceDocKind = iota
	surfaceKindPanel
	surfaceKindCard
)

func blockEffectiveCapabilities(block DocBlock) *DocBlockCapabilities {
	if block.Capabilities != nil {
		return block.Capabilities
	}
	if block.Props == nil {
		return nil
	}
	raw, ok := block.Props["capabilities"]
	if !ok || raw == nil {
		return nil
	}
	switch v := raw.(type) {
	case map[string]any:
		return docCapabilitiesFromMap(v)
	default:
		return nil
	}
}

func docCapabilitiesFromMap(m map[string]any) *DocBlockCapabilities {
	if len(m) == 0 {
		return nil
	}
	c := &DocBlockCapabilities{}
	if v, ok := m["titleBar"].(bool); ok {
		c.TitleBar = &v
	}
	if v, ok := m["collapsible"].(bool); ok {
		c.Collapsible = &v
	}
	if v, ok := m["closable"].(bool); ok {
		c.Closable = &v
	}
	if v, ok := m["movable"].(bool); ok {
		c.Movable = &v
	}
	if v, ok := m["resizable"].(bool); ok {
		c.Resizable = &v
	}
	if v, ok := m["constrain"].(bool); ok {
		c.Constrain = &v
	}
	if v, ok := m["hScroll"].(bool); ok {
		c.HScroll = &v
	}
	if v, ok := m["vScroll"].(bool); ok {
		c.VScroll = &v
	}
	if v, ok := m["anchored"].(bool); ok {
		c.Anchored = &v
	}
	if v, ok := m["closeOnEscape"].(bool); ok {
		c.CloseOnEscape = &v
	}
	if v, ok := m["anchor"].(string); ok {
		c.Anchor = v
	}
	if v, ok := m["dragMode"].(string); ok {
		c.DragMode = v
	}
	if v, ok := m["onDismiss"].(string); ok {
		c.OnDismiss = v
	}
	if v, ok := m["collapsed"].(bool); ok {
		c.Collapsed = &v
	}
	if v, ok := m["minWidth"].(float64); ok {
		c.MinWidth = float32(v)
	}
	if v, ok := m["floatX"].(float64); ok {
		f := float32(v)
		c.FloatX = &f
	}
	if v, ok := m["floatY"].(float64); ok {
		f := float32(v)
		c.FloatY = &f
	}
	return c
}

func (c *DocBlockCapabilities) isEmpty() bool {
	if c == nil {
		return true
	}
	return c.TitleBar == nil && c.Collapsible == nil && c.Closable == nil &&
		c.Movable == nil && c.Resizable == nil && c.Constrain == nil &&
		c.HScroll == nil && c.VScroll == nil && c.Anchored == nil &&
		c.CloseOnEscape == nil && c.Collapsed == nil &&
		c.Anchor == "" && c.DragMode == "" && c.OnDismiss == "" &&
		c.MinWidth == 0 && c.FloatX == nil && c.FloatY == nil
}

func validateDocCapabilities(caps *DocBlockCapabilities, block DocBlock, path string) error {
	if caps == nil || caps.isEmpty() {
		return nil
	}
	if caps.DragMode != "" {
		if _, ok := parseDocDragMode(caps.DragMode); !ok {
			return docBlockError(block, path, "unknown capabilities.dragMode %q", caps.DragMode)
		}
	}
	if caps.Anchor != "" {
		if _, _, ok := parseDocAnchor(caps.Anchor); !ok {
			return docBlockError(block, path, "unknown capabilities.anchor %q", caps.Anchor)
		}
	}
	if caps.MinWidth < 0 {
		return docBlockError(block, path, "capabilities.minWidth must be >= 0")
	}
	return nil
}

func resolveSurfaceDocKind(block DocBlock) (surfaceDocKind, error) {
	switch strings.ToLower(strings.TrimSpace(block.Variant)) {
	case "", "card":
		if block.Preset != "" {
			if p, ok := LookupPreset(block.Preset); ok {
				switch p.Component {
				case "panel":
					return surfaceKindPanel, nil
				case "card":
					return surfaceKindCard, nil
				}
			}
		}
		return surfaceKindCard, nil
	case "panel", "section":
		return surfaceKindPanel, nil
	default:
		return surfaceKindAuto, fmt.Errorf("unknown surface variant %q (use panel, section, or card)", block.Variant)
	}
}

func buildDocSurfaceAt(block DocBlock, ctx *BuildContext, path string, forceKind surfaceDocKind) (Node, error) {
	id := block.ID
	if id == "" {
		id = "doc-" + block.Type
	}
	kind := forceKind
	if kind == surfaceKindAuto {
		var err error
		kind, err = resolveSurfaceDocKind(block)
		if err != nil {
			return nil, docBlockError(block, path, "%s", err.Error())
		}
	}
	title := block.Title
	if title == "" && block.Type == "section" {
		title = block.Text
	}
	switch kind {
	case surfaceKindPanel:
		p := NewPanel(id, title, 0, 0, block.Width, block.Height)
		p.Gap = docBlockGap(block, 10)
		if err := addDocChildren(p, block.Children, ctx, path); err != nil {
			return nil, err
		}
		if block.Preset != "" {
			_ = p.SetStylePreset(block.Preset, PresetPropsFromMap(block.Props))
		} else {
			applyDocStyle(&p.Element, block)
		}
		applyDocLayout(&p.Element, block)
		if err := applyDocBlockCapabilities(block, &p.SurfaceShell, ctx, path); err != nil {
			return nil, err
		}
		return p, nil
	default:
		card := NewCard(id, title, 0, 0, block.Width, block.Height)
		card.Gap = docBlockGap(block, 10)
		if err := addDocChildren(card, block.Children, ctx, path); err != nil {
			return nil, err
		}
		if block.Preset != "" {
			_ = card.SetStylePreset(block.Preset, PresetPropsFromMap(block.Props))
		} else {
			applyDocStyle(&card.Element, block)
		}
		if v := block.Variant; v == "callout" || v == "code" {
			applyCardRecipeVariant(card, v)
		}
		applyDocLayout(&card.Element, block)
		applyCardChromeTextToBody(card, block)
		if err := applyDocBlockCapabilities(block, &card.SurfaceShell, ctx, path); err != nil {
			return nil, err
		}
		return card, nil
	}
}

func applyDocBlockCapabilities(block DocBlock, sh *SurfaceShell, ctx *BuildContext, path string) error {
	caps := blockEffectiveCapabilities(block)
	if caps == nil || caps.isEmpty() {
		return nil
	}
	if err := validateDocCapabilities(caps, block, path); err != nil {
		return err
	}
	f := surfaceFeaturesConfig(sh)
	if caps.TitleBar != nil {
		f.TitleBar = *caps.TitleBar
	}
	if caps.Collapsible != nil {
		f.Collapsible = *caps.Collapsible
	}
	if caps.Closable != nil {
		f.Closable = *caps.Closable
	}
	if caps.Movable != nil {
		f.Movable = *caps.Movable
	}
	if caps.Resizable != nil {
		f.Resizable = *caps.Resizable
	}
	if caps.Constrain != nil {
		f.Constrain = *caps.Constrain
	}
	if caps.HScroll != nil {
		f.HScroll = *caps.HScroll
	}
	if caps.VScroll != nil {
		f.VScroll = *caps.VScroll
	}
	if caps.Anchored != nil {
		f.Anchored = *caps.Anchored
	}
	if caps.CloseOnEscape != nil {
		f.CloseOnEscape = *caps.CloseOnEscape
	}
	if caps.MinWidth > 0 {
		f.MinWidth = caps.MinWidth
	}
	if caps.Anchor != "" {
		ax, ay, ok := parseDocAnchor(caps.Anchor)
		if !ok {
			return docBlockError(block, path, "unknown capabilities.anchor %q", caps.Anchor)
		}
		f.AnchorX = ax
		f.AnchorY = ay
		f.Anchored = true
	}
	if caps.DragMode != "" {
		mode, ok := parseDocDragMode(caps.DragMode)
		if !ok {
			return docBlockError(block, path, "unknown capabilities.dragMode %q", caps.DragMode)
		}
		f.DragMode = mode
	}
	if caps.FloatX != nil || caps.FloatY != nil {
		f.FloatPosition = true
		if caps.FloatX != nil {
			f.FloatX = *caps.FloatX
		}
		if caps.FloatY != nil {
			f.FloatY = *caps.FloatY
		}
		f.Anchored = false
	}
	if caps.OnDismiss != "" {
		if ctx == nil || ctx.Actions == nil || ctx.Actions[caps.OnDismiss] == nil {
			return docBlockError(block, path, "unknown action %q for capabilities.onDismiss", caps.OnDismiss)
		}
		f.OnDismiss = ctx.Actions[caps.OnDismiss]
		if caps.Closable == nil {
			f.Closable = true
		}
	}
	surfaceSyncFeatures(sh)
	if caps.Collapsible != nil && *caps.Collapsible {
		expanded := true
		if caps.Collapsed != nil {
			expanded = !*caps.Collapsed
		}
		if cb := surfaceCollapseBehavior(sh); cb != nil {
			cb.Expanded.Set(expanded)
		}
	}
	return nil
}

func parseDocDragMode(s string) (PanelDragMode, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "off":
		return PanelDragOff, true
	case "touch", "ontouch", "on-touch":
		return PanelDragOnTouch, true
	case "titlebar", "title-bar", "title":
		return PanelDragTitleBar, true
	case "anywhere", "any":
		return PanelDragAnywhere, true
	default:
		return PanelDragOff, false
	}
}

func parseDocAnchor(s string) (PanelAnchorH, PanelAnchorV, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "top-left", "topleft", "nw":
		return PanelAnchorLeft, PanelAnchorTop, true
	case "top", "top-center", "north":
		return PanelAnchorHCenter, PanelAnchorTop, true
	case "top-right", "topright", "ne":
		return PanelAnchorRight, PanelAnchorTop, true
	case "left", "center-left", "west":
		return PanelAnchorLeft, PanelAnchorVCenter, true
	case "center", "middle":
		return PanelAnchorHCenter, PanelAnchorVCenter, true
	case "right", "center-right", "east":
		return PanelAnchorRight, PanelAnchorVCenter, true
	case "bottom-left", "bottomleft", "sw":
		return PanelAnchorLeft, PanelAnchorBottom, true
	case "bottom", "bottom-center", "south":
		return PanelAnchorHCenter, PanelAnchorBottom, true
	case "bottom-right", "bottomright", "se":
		return PanelAnchorRight, PanelAnchorBottom, true
	default:
		return PanelAnchorLeft, PanelAnchorTop, false
	}
}
