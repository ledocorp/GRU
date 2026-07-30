// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

var (
	toggleTrackOff  = rl.NewColor(235, 237, 245, 255)
	toggleTrackOn   = rl.NewColor(79, 70, 229, 255)
	toggleThumbFill = rl.NewColor(255, 255, 255, 255)
)

const (
	toggleThumbPad         = float32(2)
	toggleInnerRatio       = float32(0.88) // inner disc radius vs thumb outer (thin color ring)
	toggleRoundSegments    = 16
	toggleTrackWidthRatio  = float32(0.90) // visual track width vs allocated bounds width
	toggleTrackHeightRatio = float32(0.9075) // 0.825 × 1.10 — visual track height vs allocated bounds height
)

// Toggle is a pill track (segmented-style) with a sliding circular thumb inside.
// Recommended: width ~1.9× height (e.g. 47×25).
//
// # LLM Prompt Template
//
//	tog := ui.NewToggle("dark", false, 0, 0, 52, 28)
//	tog.Value.Subscribe(func() { applyTheme(tog.Value.Get()) })
//	row.AddChild(tog)
//
// Demo scenes: **Batch 22 Toggle**, **Form Demo**, **Settings Demo** (ListTile trailing).
type Toggle struct {
	Element
	Value            *Signal[bool]
	Disabled         bool
	hostedInListTile bool // when true, ListTile owns pointer input (Update is a no-op)
	hovered          bool
	pointerHeld      bool // suppress repeat toggle while button held after a hit
	thumbOffset      float32
	tween            *Tween
}

// NewToggle creates a Toggle.
func NewToggle(id string, initialValue bool, x, y, w, h float32) *Toggle {
	initial := float32(0)
	if initialValue {
		initial = 1.0
	}
	t := &Toggle{
		Element:     NewElement(id, x, y, w, h),
		Value:       NewSignal(initialValue),
		thumbOffset: initial,
	}
	t.styleName = "toggle"
	t.Element.SetStyleVariant("toggle", "default")
	t.Value.Subscribe(func() {
		if t.tween == nil {
			if t.Value.Get() {
				t.thumbOffset = 1
			} else {
				t.thumbOffset = 0
			}
		}
		t.MarkDrawDirty()
	})
	return t
}

// PillBounds is the stadium track rectangle used for drawing the switch chrome.
func (tg *Toggle) PillBounds() rl.Rectangle { return tg.pillRect(tg.Bounds()) }

// hitBounds is the pointer target. ListTile-hosted switches use the slot only
// (no expanded touch rect that can bleed into the next row).
func (tg *Toggle) hitBounds() rl.Rectangle {
	b := tg.Bounds()
	if b.Width < 1 || b.Height < 1 {
		return tg.PillBounds()
	}
	if tg.hostedInListTile {
		return b
	}
	const minTouch = float32(44)
	if b.Width < minTouch {
		pad := (minTouch - b.Width) / 2
		b.X -= pad
		b.Width = minTouch
	}
	if b.Height < minTouch {
		pad := (minTouch - b.Height) / 2
		b.Y -= pad
		b.Height = minTouch
	}
	return b
}

func (tg *Toggle) pillRect(bounds rl.Rectangle) rl.Rectangle {
	w := bounds.Width * toggleTrackWidthRatio
	h := bounds.Height * toggleTrackHeightRatio
	minH := w * 0.45 * toggleTrackHeightRatio
	if h < minH {
		h = minH
	}
	if h > w {
		h = w
	}
	padX := (bounds.Width - w) / 2
	padY := (bounds.Height - h) / 2
	return rl.NewRectangle(bounds.X+padX, bounds.Y+padY, w, h)
}

func (tg *Toggle) thumbGeometry(pill rl.Rectangle) (cx, cy, outerR, innerR float32) {
	w, h := pill.Width, pill.Height
	thumbD := h - 2*toggleThumbPad
	if thumbD < h*0.72 {
		thumbD = h * 0.72
	}
	outerR = thumbD / 2
	innerR = outerR * toggleInnerRatio
	travel := w - thumbD - 2*toggleThumbPad
	if travel < 0 {
		travel = 0
	}
	cy = pill.Y + h/2
	cx = pill.X + toggleThumbPad + outerR + tg.thumbOffset*travel
	return cx, cy, outerR, innerR
}

// Update implements Node.Update. When hostedInListTile, ListTile calls
// updateHostedPointer instead so row rules never compete with the switch.
func (tg *Toggle) Update(dt float32) {
	if tg.hostedInListTile {
		return
	}
	tg.updatePointer(dt)
}

// updateHostedPointer runs pointer handling for a toggle in a ListTile switch row.
func (tg *Toggle) updateHostedPointer(dt float32) {
	tg.updatePointer(dt)
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (tg *Toggle) ClearOverlayPointerState() {
	if !tg.hovered {
		return
	}
	tg.hovered = false
	tg.pointerHeld = false
	tg.MarkDrawDirty()
}

func (tg *Toggle) updatePointer(dt float32) {
	if tg.IsHidden() {
		return
	}
	if tg.Disabled {
		tg.hovered = false
		return
	}

	if tg.tween != nil {
		tg.tween.Update(dt)
		if !tg.tween.IsActive {
			tg.tween = nil
			tg.MarkDrawDirty()
		}
	}

	hit := tg.hitBounds()
	mouse := rl.GetMousePosition()
	inside := rl.CheckCollisionPointRec(mouse, hit)
	down := rl.IsMouseButtonDown(rl.MouseLeftButton)

	prevHovered := tg.hovered
	tg.hovered = inside
	if tg.hovered != prevHovered {
		tg.MarkDrawDirty()
	}

	if !down {
		tg.pointerHeld = false
		return
	}

	// Down-edge inside hit: works at DeepIdleFPS (IsMouseButtonPressed is one frame).
	if inside && !tg.pointerHeld {
		tg.pointerHeld = true
		tg.toggle()
		PointerClickMarkUsed()
	}
}

func (tg *Toggle) toggle() {
	newVal := !tg.Value.Get()
	startOffset := tg.thumbOffset
	endOffset := float32(0)
	if newVal {
		endOffset = 1
	}
	// Short tween so rapid flips stay responsive.
	tg.tween = NewTween(startOffset, endOffset, 0.12, EaseOutQuad,
		func(v float32) {
			tg.thumbOffset = v
			tg.MarkDrawDirty()
		},
		nil)
	tg.Value.Set(newVal)
	tg.MarkDrawDirty()
	Wake(WakeAnimation|WakeInput, tg.ID())
}

// AnimationActive implements AnimationReporter — keeps redraw/wake during thumb tween.
func (tg *Toggle) AnimationActive() bool {
	return tg.tween != nil && tg.tween.IsActive
}

// AnimationSource implements AnimationReporter.
func (tg *Toggle) AnimationSource() string { return tg.ID() }

// DrawAnimationOverlay implements AnimationOverlayDrawer.
// Thumb motion is baked into the SSAA cache each tween frame (MarkDrawDirty).
func (tg *Toggle) DrawAnimationOverlay() {}

// InteractionOverlayActive implements InteractionOverlayPainter.
func (tg *Toggle) InteractionOverlayActive() bool { return false }

// DrawInteractionOverlay implements InteractionOverlayPainter.
func (tg *Toggle) DrawInteractionOverlay() {}

// Layout is a no-op for this leaf widget.
func (tg *Toggle) Layout() { tg.layoutDirty = false }

// Draw implements Node.Draw.
func (tg *Toggle) Draw() { tg.drawInternal() }

func (tg *Toggle) drawInternal() {
	if tg.IsHidden() {
		return
	}
	pill := tg.PillBounds()
	w, h := pill.Width, pill.Height
	if w < 10 || h < 10 {
		return
	}

	roundness := float32(1)
	t := tg.thumbOffset
	trackColor := lerpColor(toggleTrackOff, toggleTrackOn, t)
	if tg.Disabled {
		trackColor = rl.NewColor(220, 222, 230, 255)
	} else if tg.hovered && rl.IsMouseButtonDown(rl.MouseLeftButton) {
		trackColor = rl.ColorBrightness(trackColor, -0.08)
	} else if tg.hovered {
		trackColor = rl.ColorBrightness(trackColor, -0.04)
	}

	rl.DrawRectangleRounded(pill, roundness, toggleRoundSegments, trackColor)

	cx, cy, outerR, innerR := tg.thumbGeometry(pill)
	thumbFill := toggleThumbFill
	if tg.Disabled {
		thumbFill = rl.NewColor(245, 246, 250, 255)
	}
	rl.DrawCircle(int32(cx), int32(cy), outerR, thumbFill)
	if innerR > 1 {
		inner := rl.ColorBrightness(thumbFill, -0.06)
		rl.DrawCircle(int32(cx), int32(cy), innerR, inner)
	}
}

// IsInteractive implements Node.IsInteractive.
func (tg *Toggle) IsInteractive() bool { return !tg.Disabled }
