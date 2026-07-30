// Package ui (continued) — dismiss and escape surface behaviors (Phase C3).
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

// DismissBehavior adds a × close control and hides the shell on dismiss.
type DismissBehavior struct {
	shell     *SurfaceShell
	OnDismiss func()
	active    bool
	hoverClose bool
}

// NewDismissBehavior creates a dismiss plugin. Attach with SurfaceShell.AttachBehavior
// or enable via Panel/Card SetClosable(true).
func NewDismissBehavior() *DismissBehavior {
	return &DismissBehavior{active: true}
}

// Active reports whether the close control is shown and interactive.
func (d *DismissBehavior) Active() bool {
	return d.active
}

// SetActive toggles the close control without detaching the behavior.
func (d *DismissBehavior) SetActive(on bool) {
	if d.active == on {
		return
	}
	d.active = on
	if d.shell != nil {
		d.shell.MarkDrawDirty()
		d.shell.MarkDirty()
	}
}

// Dismiss hides the shell and invokes OnDismiss.
func (d *DismissBehavior) Dismiss() {
	if d.shell == nil {
		return
	}
	d.shell.Hide()
	if d.OnDismiss != nil {
		d.OnDismiss()
	}
}

// AttachShell wires the behavior to a shell.
func (d *DismissBehavior) AttachShell(sh *SurfaceShell) {
	d.shell = sh
}

func (d *DismissBehavior) Update(dt float32) {
	if d.shell == nil || d.shell.IsHidden() || !d.active {
		return
	}
	mouse := rl.GetMousePosition()
	btn := surfaceCloseBtnRect(d.shell)
	prev := d.hoverClose
	d.hoverClose = rl.CheckCollisionPointRec(mouse, btn)
	if d.hoverClose != prev {
		d.shell.MarkDrawDirty()
	}
	if d.hoverClose && rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		d.Dismiss()
	}
}

func (d *DismissBehavior) LayoutAfterBody(sh *SurfaceShell) {}

func (d *DismissBehavior) DrawOverlay(sh *SurfaceShell) {
	if sh == nil || !d.active {
		return
	}
	if sh.panelFeatures != nil {
		return
	}
	d.drawCloseButton(sh)
}

func (d *DismissBehavior) drawCloseButton(sh *SurfaceShell) {
	btn := surfaceCloseBtnRect(sh)
	style := GetThemeStyle("panel-title")
	rl.DrawRectangleRec(btn, style.BackgroundColor)
	col := style.TextColor
	if d.hoverClose {
		col = brightenColor(col, 30)
	}
	drawSurfacePhosphorIcon(btn, PhosphorX, col)
}

func (d *DismissBehavior) HeaderInteractive() bool { return d.active }

// EscapeBehavior dismisses via Escape when the pointer is over the shell.
type EscapeBehavior struct {
	shell   *SurfaceShell
	dismiss *DismissBehavior
	Enabled bool
}

// NewEscapeBehavior creates an escape-key dismiss plugin.
func NewEscapeBehavior(dismiss *DismissBehavior) *EscapeBehavior {
	return &EscapeBehavior{dismiss: dismiss}
}

func (e *EscapeBehavior) AttachShell(sh *SurfaceShell) {
	e.shell = sh
}

func (e *EscapeBehavior) Update(dt float32) {
	if e.shell == nil || e.shell.IsHidden() || !e.Enabled {
		return
	}
	if e.dismiss == nil || !e.dismiss.Active() {
		return
	}
	if !rl.IsKeyPressed(rl.KeyEscape) {
		return
	}
	mouse := rl.GetMousePosition()
	if rl.CheckCollisionPointRec(mouse, e.shell.Bounds()) {
		e.dismiss.Dismiss()
	}
}

func (e *EscapeBehavior) LayoutAfterBody(sh *SurfaceShell) {}
func (e *EscapeBehavior) DrawOverlay(sh *SurfaceShell)     {}
func (e *EscapeBehavior) HeaderInteractive() bool          { return false }
