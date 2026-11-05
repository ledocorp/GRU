// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	segmentedDefaultH = float32(36)
	segmentedPad      = float32(3)
)

// SegmentedControl is an iOS-style mutually exclusive option strip: one rounded
// track with a sliding pill behind the selected label. Pass an explicit height
// (e.g. 36) or h=0 for intrinsic height; avoid FlexGrow on the main axis in a
// column layout or the control will stretch vertically.
//
// # LLM Prompt Template
//
//	filter := ui.NewSignal(0)
//	seg := ui.NewSegmentedControl("view", []string{"All", "Active", "Done"}, filter, 0, 0, 320, 36)
//	toolbar.AddChild(seg)
//
// Demo scenes: **Batch 14 SegmentedControl**.
type SegmentedControl struct {
	Element
	Options  []string
	Selected *Signal[int]
	hoverIdx int
}

// NewSegmentedControl creates a control with options and a reactive selected index.
func NewSegmentedControl(id string, options []string, selected *Signal[int], x, y, w, h float32) *SegmentedControl {
	if selected == nil {
		selected = NewSignal(0)
	}
	if len(options) == 0 {
		options = []string{"-"}
	}
	if selected.Get() < 0 || selected.Get() >= len(options) {
		selected.Set(0)
	}
	sc := &SegmentedControl{
		Element:  NewElement(id, x, y, w, h),
		Options:  options,
		Selected: selected,
		hoverIdx: -1,
	}
	sc.styleName = "segmented"
	if h == 0 {
		sc.AutoHeight = true
	}
	sc.Selected.Subscribe(func() { sc.MarkDirty() })
	return sc
}

// IsInteractive implements Node.
func (sc *SegmentedControl) IsInteractive() bool { return true }

func (sc *SegmentedControl) innerBounds() rl.Rectangle {
	b := sc.Bounds()
	return rl.NewRectangle(b.X+segmentedPad, b.Y+segmentedPad, b.Width-2*segmentedPad, b.Height-2*segmentedPad)
}

func (sc *SegmentedControl) segmentBounds(i int) rl.Rectangle {
	inner := sc.innerBounds()
	n := len(sc.Options)
	if n <= 0 {
		return inner
	}
	w := inner.Width / float32(n)
	return rl.NewRectangle(inner.X+float32(i)*w, inner.Y, w, inner.Height)
}

// Update handles segment selection.
func (sc *SegmentedControl) Update(_ float32) {
	if sc.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	sc.hoverIdx = -1
	for i := range sc.Options {
		if rl.CheckCollisionPointRec(mouse, sc.segmentBounds(i)) {
			sc.hoverIdx = i
			break
		}
	}
	if !rl.IsMouseButtonPressed(rl.MouseLeftButton) || sc.hoverIdx < 0 {
		return
	}
	if sc.Selected.Get() != sc.hoverIdx {
		sc.Selected.Set(sc.hoverIdx)
	}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (sc *SegmentedControl) ClearOverlayPointerState() {
	sc.hoverIdx = -1
}

// Layout clamps to intrinsic height when AutoHeight is set.
func (sc *SegmentedControl) Layout() {
	defer func() { sc.layoutDirty = false }()
	if !sc.IsAutoHeight() {
		return
	}
	b := sc.Bounds()
	if b.Height < segmentedDefaultH-0.5 || b.Height > segmentedDefaultH+0.5 {
		b.Height = segmentedDefaultH
		sc.setBoundsNoMark(b)
	}
}

// Draw implements Node.Draw.
func (sc *SegmentedControl) Draw() { sc.drawInternal() }

func (sc *SegmentedControl) drawInternal() {
	if sc.IsHidden() {
		return
	}
	b := sc.Bounds()
	inner := sc.innerBounds()
	if inner.Width < 1 || inner.Height < 1 {
		return
	}

	trackStyle, _ := sc.ResolveStyle(StyleStateNone)
	if trackStyle.BackgroundColor.A == 0 {
		trackStyle.BackgroundColor = rl.NewColor(235, 237, 245, 255)
	}
	roundness := float32(0.35)
	if inner.Height > 0 {
		roundness = (inner.Height * 0.35) / (inner.Height / 2)
		if roundness > 1 {
			roundness = 1
		}
	}
	// Outer track — tess fill (hairline AA fringe left dark arcs outside the pill).
	if roundness > 0 {
		rl.DrawRectangleRounded(inner, roundness, 64, trackStyle.BackgroundColor)
	} else {
		rl.DrawRectangleRec(inner, trackStyle.BackgroundColor)
	}

	sel := sc.Selected.Get()
	if sel < 0 {
		sel = 0
	}
	if sel >= len(sc.Options) {
		sel = len(sc.Options) - 1
	}
	pillFill := rl.NewColor(255, 255, 255, 255)
	if len(sc.Options) > 0 {
		pill := sc.segmentBounds(sel)
		pillPad := float32(2)
		pill = rl.NewRectangle(pill.X+pillPad, pill.Y+pillPad, pill.Width-2*pillPad, pill.Height-2*pillPad)
		if pill.Width < 1 {
			pill.Width = 1
		}
		if pill.Height < 1 {
			pill.Height = 1
		}
		// Match track end-cap curvature so the selected pill sits cleanly inside.
		pillRound := roundness
		if pill.Height > 0 && pill.Height < inner.Height {
			pillRound = (pill.Height * 0.5) / (pill.Height / 2)
			if pillRound > 1 {
				pillRound = 1
			}
			if roundness < pillRound {
				pillRound = roundness
			}
		}
		if pillRound > 0 {
			rl.DrawRectangleRounded(pill, pillRound, 64, pillFill)
		} else {
			rl.DrawRectangleRec(pill, pillFill)
		}
	}

	for i, opt := range sc.Options {
		seg := sc.segmentBounds(i)
		labelStyle := trackStyle
		dest := trackStyle.BackgroundColor
		if i == sel {
			labelStyle.TextColor = rl.NewColor(30, 27, 75, 255)
			dest = pillFill
		} else {
			labelStyle.TextColor = rl.NewColor(90, 94, 118, 255)
		}
		if labelStyle.FontSize <= 0 {
			labelStyle.FontSize = 14
		}
		tw := measureTextS(opt, labelStyle)
		tx := int32(seg.X + (seg.Width-float32(tw))/2)
		ty := controlTextPosY(seg, labelStyle)
		popLCD := PushLCDSurfaceBackdrop(dest)
		drawTextS(opt, tx, ty, labelStyle)
		PopLCDSurfaceBackdrop(popLCD)
	}
	_ = b
}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (sc *SegmentedControl) InteractionOverlayActive() bool {
	if sc == nil || sc.IsHidden() || sc.hoverIdx < 0 {
		return false
	}
	sel := sc.Selected.Get()
	return sc.hoverIdx != sel
}

// DrawInteractionOverlay paints a hover pill matching the selected segment
// (PRESENT_1X_COMPOSITE — do not scissor the full track; that left an uneven wrap).
func (sc *SegmentedControl) DrawInteractionOverlay() {
	if !sc.InteractionOverlayActive() {
		return
	}
	i := sc.hoverIdx
	if i < 0 || i >= len(sc.Options) {
		return
	}
	trackStyle, _ := sc.ResolveStyle(StyleStateNone)
	if trackStyle.BackgroundColor.A == 0 {
		trackStyle.BackgroundColor = rl.NewColor(235, 237, 245, 255)
	}
	seg := sc.segmentBounds(i)
	pillPad := float32(2)
	pill := rl.NewRectangle(seg.X+pillPad, seg.Y+pillPad, seg.Width-2*pillPad, seg.Height-2*pillPad)
	if pill.Width < 1 {
		pill.Width = 1
	}
	if pill.Height < 1 {
		pill.Height = 1
	}
	pillRound := float32(1)
	if pill.Height > 0 {
		pillRound = (pill.Height * 0.5) / (pill.Height / 2)
		if pillRound > 1 {
			pillRound = 1
		}
	}
	hoverFill := rl.NewColor(255, 255, 255, 220)
	rl.DrawRectangleRounded(pill, pillRound, 64, hoverFill)

	labelStyle := trackStyle
	labelStyle.TextColor = rl.NewColor(55, 58, 78, 255)
	if labelStyle.FontSize <= 0 {
		labelStyle.FontSize = 14
	}
	popLCD := PushLCDSurfaceBackdrop(hoverFill)
	tw := measureTextS(sc.Options[i], labelStyle)
	tx := int32(seg.X + (seg.Width-float32(tw))/2)
	ty := controlTextPosY(seg, labelStyle)
	drawTextS(sc.Options[i], tx, ty, labelStyle)
	PopLCDSurfaceBackdrop(popLCD)
}
