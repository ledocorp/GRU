// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"sort"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Accordion is a titled expandable/collapsible container.
//
// # Layout
//
// The Accordion always renders its header row (TitleHeight px tall) plus an
// animated content area that transitions between 0 px (collapsed) and the
// natural content height (expanded).  Children are laid out vertically inside
// the content area with the Accordion's Gap and Padding applied.
//
// The accordion height therefore varies over time while the animation is
// running.  Parent containers (Panel, Viewport) should not give the Accordion
// a fixed height unless they intentionally want to clip it — set h=0 and let
// the accordion control its own height.
//
// # Expand / Collapse
//
// Clicking the title row toggles the Expanded signal.  The same toggle can be
// driven programmatically:
//
//	acc.Expanded.Set(true)   // expand
//	acc.Expanded.Set(false)  // collapse
//
// An OnToggle callback is fired synchronously after each toggle.
//
// # Animation
//
// The content area height is interpolated from its current value to the target
// height using EaseInOutQuad over AnimDuration seconds (default 0.11s).  The
// animation drives MarkDirty so the layout engine reflows the accordion and its
// ancestors every frame while the animation is running.
//
// # Styles
//
//	"accordion"        — header background, text colour, font size, padding, corner radius
//	"accordion-open"   — header background when expanded (accent shade)
//	"accordion-body"   — content area background, padding
//
// # LLM Prompt Template
//
//	acc := ui.NewAccordion("settings-acc", "Advanced Settings", 0, 0, 440, 0)
//	acc.AddChild(ui.NewLabel("lbl1", "Debug mode", 0, 0, 0, 26))
//	acc.AddChild(ui.NewToggle("dbg", nil, 0, 0, 44, 24))
//	panel.AddChild(acc)
//
// Demo scenes: **Batch 3b Accordion Demo**.
type Accordion struct {
	// Element is the base; accordion is a leaf + children hybrid.
	// We embed Element directly (not Container) because we control child layout
	// ourselves — the content height changes every animation frame.
	Element

	// Title is the text shown in the header row.
	Title string

	// Expanded is the reactive expanded/collapsed state.
	// Subscribe to it to react to external programmatic changes.
	Expanded *Signal[bool]

	// OnToggle is called (with the new state) when the accordion is toggled,
	// either by a click or by Expanded.Set.
	OnToggle func(expanded bool)

	// AnimDuration is the expand/collapse animation duration in seconds.
	// Default is 0.11.
	AnimDuration float32

	// TitleHeight is the pixel height of the clickable header row. Default 34.
	TitleHeight float32

	// Gap between children inside the content area (default 6).
	Gap float32

	// children is the list of content nodes (managed manually, not via Container).
	children []Node

	// contentH is the target full height of the content (computed from children).
	contentH float32

	// animH is the current animated content height (0 → contentH when expanding).
	animH float32

	// tween is the active height animation; nil when idle.
	tween *Tween

	// hoverHeader tracks whether the mouse is over the header for hover feedback.
	hoverHeader bool

	// pendingExpand defers the first open until width is known (Expanded.Set during Build).
	pendingExpand bool

	lastLayoutW float32
}

// NewAccordion creates an Accordion with the given title.
//
//	id    — unique node ID
//	title — text displayed in the clickable header
//	x, y  — position (overridden by parent layout containers)
//	w     — width; 0 = fill parent via cross-axis stretch
//	h     — pass 0 so the accordion manages its own height
func NewAccordion(id, title string, x, y, w, h float32) *Accordion {
	a := &Accordion{
		Element:      NewElement(id, x, y, w, h),
		Title:        title,
		Expanded:     NewSignal(false),
		AnimDuration: 0.11,
		TitleHeight:  34,
		Gap:          6,
	}
	// Override height to TitleHeight regardless of what the caller passed.
	// This ensures parent layout containers see the correct height from frame 1
	// even when the accordion is created collapsed (h=0 is the common idiom).
	a.bounds.Height = a.TitleHeight
	a.styleName = "accordion"

	// When Expanded changes (whether from a click or programmatic Set) start
	// the appropriate animation and notify the caller.
	a.Expanded.Subscribe(func() {
		exp := a.Expanded.Get()
		if exp {
			a.startExpand()
		} else {
			a.startCollapse()
		}
		if a.OnToggle != nil {
			a.OnToggle(exp)
		}
	})

	return a
}

// ─── Public API ───────────────────────────────────────────────────────────────

// AddChild appends a child widget to the accordion's content area.
func (a *Accordion) AddChild(child Node) {
	a.children = append(a.children, child)
	child.SetParent(a)
	a.recomputeContentH()
	a.MarkDirty()
}

// Children returns the content-area children (satisfies Node).
func (a *Accordion) Children() []Node { return a.children }

// Toggle flips the expanded state, triggering the animation.
func (a *Accordion) Toggle() {
	a.Expanded.Set(!a.Expanded.Get())
}

// ─── Node implementation ──────────────────────────────────────────────────────

// IsInteractive returns true — the header row is a click target.
func (a *Accordion) IsInteractive() bool { return true }

// AnimationActive reports whether the expand/collapse tween is currently moving.
func (a *Accordion) AnimationActive() bool {
	return !a.IsHidden() && a.tween != nil && a.tween.IsActive
}

// AnimationSource returns a compact source label for perf diagnostics.
func (a *Accordion) AnimationSource() string { return a.ID() }

// Update handles mouse hover/click and advances the animation each frame.
func (a *Accordion) Update(dt float32) {
	if a.IsHidden() {
		return
	}

	// ── Advance tween ─────────────────────────────────────────────────────────
	if a.tween != nil && a.tween.IsActive {
		a.tween.Update(dt)
		// tween.OnUpdate sets animH and calls MarkDirty — no extra call needed.
	} else if a.tween != nil && !a.tween.IsActive {
		a.tween = nil
	}

	// ── Header hover and click ────────────────────────────────────────────────
	mouse := rl.GetMousePosition()
	headerRect := rl.NewRectangle(a.bounds.X, a.bounds.Y, a.bounds.Width, a.TitleHeight)
	prevHover := a.hoverHeader
	a.hoverHeader = rl.CheckCollisionPointRec(mouse, headerRect)
	if a.hoverHeader != prevHover {
		a.MarkDrawDirty()
	}

	if a.hoverHeader && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		a.Toggle()
	}

	// ── Update children ───────────────────────────────────────────────────────
	if a.animH > 0 {
		for _, child := range a.children {
			if !child.IsHidden() {
				UpdateNodeOverlayAware(child, dt)
			}
		}
	}
}

// Layout positions children inside the content area and refreshes contentH.
// bounds.Height is kept up-to-date by tween callbacks (and by NewAccordion for
// the initial collapsed height); this pass must use post-Layout child heights
// so targets match wrapped / AutoHeight descendants.
func (a *Accordion) Layout() {
	if a.IsHidden() {
		return
	}
	if a.lastLayoutW > 0 && a.bounds.Width != a.lastLayoutW {
		for _, ch := range a.children {
			if !ch.IsHidden() {
				ch.MarkDirty()
			}
		}
		a.MarkDirty()
	}

	measured := a.layoutContentArea()
	a.contentH = measured

	if a.pendingExpand && a.Expanded.Get() && a.bounds.Width > 0 {
		a.pendingExpand = false
		a.animH = measured
		a.tween = nil
	}

	// Keep outer bounds aligned with expanded/collapsed state so grid/flex siblings
	// are not placed on top of an open body (common when Expanded.Set runs at Build
	// before the parent assigns width).
	prevH := a.bounds.Height
	if a.tween == nil {
		switch {
		case a.Expanded.Get():
			if a.animH != measured {
				a.animH = measured
			}
		default:
			a.animH = 0
		}
	}
	wantH := a.TitleHeight + a.animH
	if a.bounds.Height != wantH {
		a.bounds.Height = wantH
	}
	if a.bounds.Height != prevH {
		if p := a.ParentNode(); p != nil {
			p.MarkDirty()
		}
	}

	if !a.IsDirty() && a.tween == nil {
		a.lastLayoutW = a.bounds.Width
		syncLayoutExtent(a)
		return
	}

	a.layoutDirty = false
	a.lastLayoutW = a.bounds.Width
	syncLayoutExtent(a)
}

// layoutContentArea arranges visible content children in the body band and
// returns the body height (top padding + rows + inner gaps + bottom padding),
// i.e. the value animH approaches when fully expanded.
func (a *Accordion) layoutContentArea() float32 {
	style := GetThemeStyle("accordion-body")
	pad := style.Padding
	if pad == 0 {
		pad = 12
	}
	contentW := a.bounds.Width - 2*pad
	if contentW < 0 {
		contentW = 0
	}
	y := a.bounds.Y + a.TitleHeight + pad

	var visible []Node
	for _, ch := range a.children {
		if !ch.IsHidden() {
			visible = append(visible, ch)
		}
	}
	prepareAccordionBodyContent(contentW, visible)
	collapsed := !a.Expanded.Get() && a.tween == nil && a.animH <= 0
	if collapsed {
		inner := float32(0)
		measureY := a.bounds.Y + a.TitleHeight + pad
		for i, child := range visible {
			cb := child.Bounds()
			cb.X = a.bounds.X + pad
			cb.Y = measureY
			if cb.Width == 0 || cb.Width > contentW {
				cb.Width = contentW
			}
			child.SetBounds(cb)
			fitSubtreeLabels(contentW, []Node{child})
			child.Layout()
			inner += child.Bounds().Height
			if i < len(visible)-1 {
				inner += a.Gap
			}
		}
		return inner + 2*pad
	}
	for i, child := range visible {
		cb := child.Bounds()
		cb.X = a.bounds.X + pad
		cb.Y = y
		if cb.Width == 0 || cb.Width > contentW {
			cb.Width = contentW
		}
		child.SetBounds(cb)
		fitSubtreeLabels(contentW, []Node{child})
		child.Layout()
		h := child.Bounds().Height
		y += h
		if i < len(visible)-1 {
			y += a.Gap
		}
	}
	y += pad
	bodyH := y - (a.bounds.Y + a.TitleHeight)
	if bodyH < 0 {
		return 0
	}
	return bodyH
}

// Draw renders the header and the animated content area.
func (a *Accordion) Draw() {
	if a.IsHidden() {
		return
	}
	a.drawInternal()
	a.drawDirty = false
}

func (a *Accordion) drawInternal() {
	if a.animH > 0 {
		a.drawOpenPanelChrome()
		return
	}
	a.drawCollapsedChrome()
}

func (a *Accordion) drawOpenPanelChrome() {
	bounds := a.bounds
	panelSt := GetThemeStyle("panel")
	headerSt := GetThemeStyle("accordion-open")
	bodySt := GetThemeStyle("accordion-body")
	r := panelSt.CornerRadius
	if r <= 0 {
		r = PresetSurfaceCornerRadius
	}

	totalH := a.TitleHeight + a.animH
	shell := rl.NewRectangle(bounds.X, bounds.Y, bounds.Width, totalH)
	bodyBg := bodySt.BackgroundColor

	// Header strip: rounded top, square bottom (matches panel title-band geometry).
	headerFill := rl.NewRectangle(bounds.X, bounds.Y, bounds.Width, a.TitleHeight)
	drawFlatBottomChromeFill(headerFill, r, headerSt.BackgroundColor)
	if a.hoverHeader {
		hoverRect := rl.NewRectangle(bounds.X, bounds.Y, bounds.Width, a.TitleHeight)
		rl.DrawRectangleRec(hoverRect, rl.NewColor(0, 0, 0, 10))
	}

	// Body band: flat top at the divider, rounded bottom cap on the shell.
	if a.animH > 0 {
		bodyRect := rl.NewRectangle(bounds.X, bounds.Y+a.TitleHeight, bounds.Width, a.animH)
		rl.DrawRectangleRec(bodyRect, bodyBg)
		if a.animH >= r {
			cap := rl.NewRectangle(bounds.X, bounds.Y+totalH-r, bounds.Width, r)
			capRound := chromeRoundness(cap, r)
			if capRound > 0 {
				rl.DrawRectangleRounded(cap, capRound, 8, bodyBg)
			}
		}
	}

	pad := headerSt.Padding
	if pad == 0 {
		pad = 14
	}
	chevSize := float32(8)
	textY := int32(bounds.Y) + (int32(a.TitleHeight)-headerSt.FontSize)/2
	maxTitleW := bounds.Width - 2*pad - chevSize - 8
	if maxTitleW < 8 {
		maxTitleW = 8
	}
	titleText := truncateTextS(a.Title, maxTitleW, headerSt)
	drawTextS(titleText, int32(bounds.X)+int32(pad), textY, headerSt)

	chevX := bounds.X + bounds.Width - pad - chevSize
	chevY := bounds.Y + a.TitleHeight/2
	progress := float32(1)
	if a.contentH > 0 {
		progress = a.animH / a.contentH
		if progress > 1 {
			progress = 1
		}
	}
	if a.Expanded.Get() && a.tween == nil {
		progress = 1
	}
	drawChevron(chevX, chevY, chevSize, progress, headerSt.TextColor)

	sepY := int32(bounds.Y + a.TitleHeight)
	rl.DrawLine(int32(bounds.X), sepY, int32(bounds.X+bounds.Width), sepY, headerSt.BorderColor)

	clip := rl.NewRectangle(bounds.X, bounds.Y+a.TitleHeight, bounds.Width, a.animH)
	clip = intersectRectsWithViewportAncestors(clip, a)
	if clip.Width > 0 && clip.Height > 0 {
		beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		sorted := make([]Node, len(a.children))
		copy(sorted, a.children)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].GetZIndex() < sorted[j].GetZIndex()
		})
		for _, child := range sorted {
			child.Draw()
			beginScissorMode(int32(clip.X), int32(clip.Y), int32(clip.Width), int32(clip.Height))
		}
		rl.EndScissorMode()
	}

	if panelSt.BorderWidth > 0 {
		borderBounds := chromeBorderBounds(shell, panelSt.BorderWidth)
		bodyRoundness := chromeRoundness(borderBounds, r)
		if bodyRoundness > 0 {
			rl.DrawRectangleRoundedLinesEx(borderBounds, bodyRoundness, 8, panelSt.BorderWidth, panelSt.BorderColor)
		} else {
			rl.DrawRectangleLinesEx(borderBounds, panelSt.BorderWidth, panelSt.BorderColor)
		}
	}
}

func (a *Accordion) drawCollapsedChrome() {
	bounds := a.bounds
	headerStyle := GetThemeStyle("accordion")
	bodyStyle := GetThemeStyle("accordion-body")

	r := headerStyle.CornerRadius
	if bodyStyle.CornerRadius > 0 && bodyStyle.CornerRadius < r {
		r = bodyStyle.CornerRadius
	}
	roundness := accordionRoundness(bounds.Width, a.TitleHeight, r)
	totalRect := rl.NewRectangle(bounds.X, bounds.Y, bounds.Width, a.TitleHeight)

	if a.hoverHeader {
		drawAccordionHeaderBand(bounds, a.TitleHeight, roundness, false, headerStyle.BackgroundColor)
		hover := rl.NewColor(0, 0, 0, 14)
		drawAccordionHeaderBand(bounds, a.TitleHeight, roundness, false, hover)
	} else {
		drawAccordionHeaderBand(bounds, a.TitleHeight, roundness, false, headerStyle.BackgroundColor)
	}

	pad := headerStyle.Padding
	if pad == 0 {
		pad = 14
	}
	chevSize := float32(8)
	textY := int32(bounds.Y) + (int32(a.TitleHeight)-headerStyle.FontSize)/2
	maxTitleW := bounds.Width - 2*pad - chevSize - 8
	if maxTitleW < 8 {
		maxTitleW = 8
	}
	titleText := truncateTextS(a.Title, maxTitleW, headerStyle)
	drawTextS(titleText, int32(bounds.X)+int32(pad), textY, headerStyle)

	chevX := bounds.X + bounds.Width - pad - chevSize
	chevY := bounds.Y + a.TitleHeight/2
	drawChevron(chevX, chevY, chevSize, 0, headerStyle.TextColor)

	bw := headerStyle.BorderWidth
	if bw > 0 {
		bc := headerStyle.BorderColor
		if roundness > 0 {
			rl.DrawRectangleRoundedLinesEx(totalRect, roundness, 6, bw, bc)
		} else {
			rl.DrawRectangleLinesEx(totalRect, bw, bc)
		}
	}
}

// ─── Internal helpers ─────────────────────────────────────────────────────────

func accordionRoundness(width, height, radius float32) float32 {
	if radius <= 0 || height <= 0 {
		return 0
	}
	shorter := width
	if height < shorter {
		shorter = height
	}
	rn := radius / (shorter / 2)
	if rn > 1 {
		return 1
	}
	return rn
}

func drawAccordionHeaderBand(bounds rl.Rectangle, titleHeight, roundness float32, open bool, color rl.Color) {
	titleRect := rl.NewRectangle(bounds.X, bounds.Y, bounds.Width, titleHeight)
	if roundness <= 0 {
		rl.DrawRectangleRec(titleRect, color)
		return
	}
	if open {
		// Match panel title bar: rounded cap + square bottom half at the divider.
		rl.DrawRectangleRounded(titleRect, roundness, 6, color)
		if titleRect.Height > titleHeight/2 {
			square := rl.NewRectangle(bounds.X, bounds.Y+titleHeight/2, bounds.Width, titleHeight/2)
			rl.DrawRectangleRec(square, color)
		}
		return
	}
	rl.DrawRectangleRounded(titleRect, roundness, 6, color)
}

// recomputeContentH recalculates the natural full height of all children.
// When the accordion already has a width from the parent, this runs the same
// arrange pass as Layout so wrapped content and post-Layout heights match.
// Before width is known (e.g. first AddChild during tree build), fall back to
// summing current child bounds.
func (a *Accordion) recomputeContentH() {
	style := GetThemeStyle("accordion-body")
	pad := style.Padding
	if pad == 0 {
		pad = 12
	}
	if a.bounds.Width > 2*pad+4 {
		a.contentH = a.layoutContentArea()
		return
	}
	total := pad * 2
	visible := 0
	for _, child := range a.children {
		if !child.IsHidden() {
			total += child.Bounds().Height
			visible++
		}
	}
	if visible > 1 {
		total += float32(visible-1) * a.Gap
	}
	a.contentH = total
}

// startExpand begins the height animation toward contentH.
func (a *Accordion) startExpand() {
	a.recomputeContentH()
	if a.bounds.Width <= 0 && a.lastLayoutW <= 0 {
		a.pendingExpand = true
		return
	}
	a.pendingExpand = false
	start := a.animH
	end := a.contentH
	if end <= 0 {
		a.animH = 0
		a.bounds.Height = a.TitleHeight
		a.MarkDirty()
		return
	}
	if start >= end {
		// Already fully expanded — snap to final state, no animation.
		a.animH = end
		a.bounds.Height = a.TitleHeight + end
		a.tween = nil
		a.MarkDirty()
		return
	}
	dur := a.AnimDuration * (1 - start/end) // proportional duration for interrupted anim
	if dur < 0.04 {
		dur = 0.04
	}
	a.tween = NewTween(start, end, dur, EaseInOutQuad,
		func(v float32) {
			a.animH = v
			// Update bounds.Height immediately so parent layout containers
			// (Panel, Container) see the correct height before Layout() runs.
			a.bounds.Height = a.TitleHeight + v
			a.MarkDirty()
		},
		func() {
			a.animH = a.contentH
			a.bounds.Height = a.TitleHeight + a.contentH
			a.tween = nil
			a.MarkDirty()
		},
	)
}

// startCollapse begins the height animation toward 0.
func (a *Accordion) startCollapse() {
	a.pendingExpand = false
	start := a.animH
	if start <= 0 {
		// Already collapsed — snap to final state.
		a.animH = 0
		a.bounds.Height = a.TitleHeight
		a.tween = nil
		a.MarkDirty()
		return
	}
	if a.contentH <= 0 {
		a.animH = 0
		a.bounds.Height = a.TitleHeight
		a.MarkDirty()
		return
	}
	dur := a.AnimDuration * (start / a.contentH) // proportional duration for interrupted anim
	if dur < 0.04 {
		dur = 0.04
	}
	a.tween = NewTween(start, 0, dur, EaseInOutQuad,
		func(v float32) {
			a.animH = v
			// Update bounds.Height immediately so parent layout containers
			// (Panel, Container) see the correct height before Layout() runs.
			a.bounds.Height = a.TitleHeight + v
			a.MarkDirty()
		},
		func() {
			a.animH = 0
			a.bounds.Height = a.TitleHeight
			a.tween = nil
			a.MarkDirty()
		},
	)
}

// drawChevron draws a right-pointing chevron (>) rotated by progress*90°.
// progress=0 → pointing right (▶, collapsed)
// progress=1 → pointing down (▼, expanded)
func drawChevron(cx, cy, size, progress float32, col rl.Color) {
	// Angle sweeps from -90° (right) to 0° (down) as progress goes 0→1.
	// Actually let's think of it as: at progress=0 we draw ▶ (pointing right),
	// at progress=1 we draw ▼ (pointing down). Rotate by progress * 90°.
	angle := progress * 90 // degrees
	rad := angle * (3.14159265 / 180)
	cos := float32(1)
	sin := float32(0)
	// Simple cos/sin approximation using the math package is not imported here,
	// so we use a lookup: raylib has no rotate-draw for primitives, use manual math.
	// We import nothing extra — use rl.Vector2Rotate if available, else manual.
	// Manual: cosine and sine via Taylor or just use rl helpers.
	// rl has no direct trig, but we can use the Go runtime math.
	// We handle this by computing cos/sin inline via approximations or import.
	// Since animation.go already imports "math", we delegate via a small helper.
	cos, sin = cossin(rad)

	// ▶ shape: three points of an arrowhead centred at (cx, cy).
	// tip is to the right at progress=0.
	halfSize := size * 0.5

	// Base two points (back of chevron), tip is front.
	// Unrotated: tip=(+halfSize, 0), back top=(−halfSize, −halfSize), back bot=(−halfSize, +halfSize)
	tipX := cx + cos*halfSize - sin*0
	tipY := cy + sin*halfSize + cos*0

	topX := cx + cos*(-halfSize) - sin*(-halfSize)
	topY := cy + sin*(-halfSize) + cos*(-halfSize)

	botX := cx + cos*(-halfSize) - sin*halfSize
	botY := cy + sin*(-halfSize) + cos*halfSize

	rl.DrawTriangle(
		rl.NewVector2(tipX, tipY),
		rl.NewVector2(botX, botY),
		rl.NewVector2(topX, topY),
		col,
	)
}

// cossin returns cos(rad) and sin(rad) using a simple polynomial approximation
// that avoids importing "math" into this file while maintaining reasonable
// accuracy for the small angle range [0, π/2].
func cossin(rad float32) (c, s float32) {
	// Use double-precision via cast to float64 via the rl math package.
	// Since we already link against math implicitly through raylib-go, it's
	// available.  Alternatively, use a 4-term Taylor series — adequate for
	// 90° range with error < 0.0003.
	//
	// Taylor: sin(x) ≈ x - x³/6 + x⁵/120 - x⁷/5040
	//         cos(x) ≈ 1 - x²/2 + x⁴/24 - x⁶/720
	x := float64(rad)
	x2 := x * x
	x3 := x2 * x
	x4 := x2 * x2
	x5 := x4 * x
	x6 := x4 * x2
	x7 := x6 * x
	sinV := x - x3/6 + x5/120 - x7/5040
	cosV := 1 - x2/2 + x4/24 - x6/720
	return float32(cosV), float32(sinV)
}
