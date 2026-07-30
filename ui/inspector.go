// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"reflect"
	"runtime"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// ─── Inspector ───────────────────────────────────────────────────────────────
//
// Inspector is a toggleable overlay panel for debugging the widget tree.
//
// Usage in main.go:
//
//	insp := ui.NewInspector()
//	// inside update loop:
//	insp.Update(doc.Root, windowW, windowH)
//	// inside draw loop (after widget draw, before EndDrawing):
//	insp.Draw()
//
// Controls:
//   - F12           — toggle inspector visibility
//   - Left-click    — select the topmost widget under the cursor
//   - Escape        — deselect current widget
//   - Tree +/-      — expand/collapse branch (Shift+click row toggles too)

const (
	inspW        = 340 // panel width
	inspLineH    = 26  // pixels per tree row
	inspPad      = 10  // inner horizontal padding
	inspHeaderH  = 100 // top bar height
	inspDetailH  = 300 // detail pane height at the bottom
	inspFoldW    = 18  // click target for expand/collapse
)

var (
	inspBg        = rl.NewColor(18, 20, 30, 220)
	inspHeader    = rl.NewColor(30, 34, 50, 255)
	inspBorder    = rl.NewColor(70, 75, 100, 255)
	inspTextDim   = rl.NewColor(130, 135, 160, 255)
	inspTextBrt   = rl.NewColor(220, 222, 235, 255)
	inspSel       = rl.NewColor(80, 105, 200, 80)
	inspSelBorder = rl.NewColor(100, 130, 255, 200)
	inspHover     = rl.NewColor(255, 200, 60, 120)
	inspKeyColor  = rl.NewColor(140, 200, 255, 255)
	inspValColor  = rl.NewColor(200, 255, 170, 255)
)

func inspDrawText(text string, x, y float32, color rl.Color) {
	s := ChromeDimStyle()
	s.TextColor = color
	DrawChromeText(text, x, y, s)
}

func inspDrawTextBold(text string, x, y float32, color rl.Color) {
	s := ChromeTitleStyle()
	s.TextColor = color
	DrawChromeText(text, x, y, s)
}

func inspDrawTreeText(text string, x, y float32, color rl.Color) {
	s := ChromeInspectorTreeStyle()
	s.TextColor = color
	DrawChromeText(text, x, y, s)
}

// treeRow is a single line in the widget-tree view.
type treeRow struct {
	node        Node
	depth       int
	label       string
	hasChildren bool
	collapsed   bool
}

// Inspector is the live widget-tree debugger.
type Inspector struct {
	visible   bool
	dockRight bool // true = dock on right edge (default)
	selected  Node // currently selected widget
	hovered   Node // widget under cursor this frame (for highlight)
	rows      []treeRow
	scrollY   float32 // tree scroll offset (pixels)
	panelX    int32   // left edge of inspector panel (= windowW - inspW)
	panelH    int32   // total height of inspector panel (= windowH)
	doc       *Document
	collapsed map[string]bool // node ID → branch collapsed
}

// NewInspector creates a new Inspector in the hidden state.
func NewInspector() *Inspector {
	return &Inspector{collapsed: make(map[string]bool), dockRight: true}
}

// Toggle shows or hides the inspector panel.
func (ins *Inspector) Toggle() { ins.visible = !ins.visible }

// ToggleDock moves the panel between the left and right edge.
func (ins *Inspector) ToggleDock() { ins.dockRight = !ins.dockRight }

// DockRight reports whether the panel is on the right edge.
func (ins *Inspector) DockRight() bool { return ins.dockRight }

func (ins *Inspector) panelLeft(windowW int32) int32 {
	if ins.dockRight {
		return windowW - int32(inspW)
	}
	return 0
}

func (ins *Inspector) dockToggleRect() rl.Rectangle {
	x := float32(ins.panelX) + float32(inspW) - float32(inspPad) - 80
	return rl.NewRectangle(x, 72, 80, 22)
}

// ResetForSceneChange clears selection and tree UI state when the active scene reloads.
func (ins *Inspector) ResetForSceneChange() {
	ins.selected = nil
	ins.hovered = nil
	ins.scrollY = 0
	ins.collapsed = make(map[string]bool)
}

// Update should be called every frame before Draw.
func (ins *Inspector) Update(root Node, doc *Document, windowW, windowH int32) {
	ins.doc = doc
	ins.panelX = ins.panelLeft(windowW)
	ins.panelH = windowH

	if rl.IsKeyPressed(rl.KeyF12) {
		ins.visible = !ins.visible
	}
	if !ins.visible {
		return
	}

	if rl.IsKeyPressed(rl.KeyEscape) {
		ins.selected = nil
	}

	wheel := rl.GetMouseWheelMove()
	if wheel != 0 {
		ins.scrollY -= wheel * float32(inspLineH*3)
		if ins.scrollY < 0 {
			ins.scrollY = 0
		}
	}

	ins.rows = ins.rows[:0]
	ins.buildRows(root, 0)

	mouse := rl.GetMousePosition()
	outsidePanel := ins.pointerOutsidePanel(mouse)

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && rl.CheckCollisionPointRec(mouse, ins.dockToggleRect()) {
		ins.ToggleDock()
	}

	ins.hovered = nil
	if outsidePanel {
		ins.hovered = ins.hitTest(root, mouse)
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && outsidePanel {
		ins.selected = ins.hovered
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) && !outsidePanel {
		treeTop := int32(inspHeaderH)
		treeBot := ins.panelH - int32(inspDetailH)
		if mouse.Y >= float32(treeTop) && mouse.Y < float32(treeBot) {
			clickRow := int((mouse.Y - float32(treeTop) + ins.scrollY) / float32(inspLineH))
			if clickRow >= 0 && clickRow < len(ins.rows) {
				row := ins.rows[clickRow]
				foldX := float32(ins.panelX+inspPad) + float32(row.depth)*14
				if row.hasChildren && mouse.X >= foldX && mouse.X < foldX+inspFoldW {
					id := row.node.ID()
					ins.collapsed[id] = !ins.collapsed[id]
				} else if rl.IsKeyDown(rl.KeyLeftShift) || rl.IsKeyDown(rl.KeyRightShift) {
					if row.hasChildren {
						id := row.node.ID()
						ins.collapsed[id] = !ins.collapsed[id]
					}
				} else {
					ins.selected = row.node
				}
			}
		}
	}
}

func (ins *Inspector) pointerOutsidePanel(mouse rl.Vector2) bool {
	if ins.dockRight {
		return mouse.X < float32(ins.panelX)
	}
	return mouse.X >= float32(ins.panelX)+inspW
}

// Draw renders the inspector overlay.
func (ins *Inspector) Draw() {
	if !ins.visible {
		return
	}

	if ins.selected != nil && !ins.nodeStillLive(ins.selected) {
		ins.selected = nil
	}

	if ins.hovered != nil && ins.hovered != ins.selected && ins.nodeStillLive(ins.hovered) {
		b := ins.hovered.Bounds()
		rl.DrawRectangleRec(b, inspHover)
		rl.DrawRectangleLinesEx(b, 1.5, rl.NewColor(255, 200, 60, 220))
	}
	if ins.selected != nil && ins.nodeStillLive(ins.selected) {
		b := ins.selected.Bounds()
		rl.DrawRectangleRec(b, inspSel)
		rl.DrawRectangleLinesEx(b, 2, inspSelBorder)
	}

	panelRect := rl.NewRectangle(float32(ins.panelX), 0, inspW, float32(ins.panelH))
	rl.DrawRectangleRec(panelRect, inspBg)
	rl.DrawRectangleLinesEx(panelRect, 1, inspBorder)

	rl.DrawRectangle(ins.panelX, 0, int32(inspW), inspHeaderH, inspHeader)

	titleStyle := ChromeTitleStyle()
	titleStyle.TextColor = inspTextBrt
	titleY := ChromeTextCenterY(rl.NewRectangle(float32(ins.panelX), 0, inspW, 28), titleStyle)
	inspDrawTextBold("Inspector  [F12]", float32(ins.panelX+inspPad), titleY, inspTextBrt)
	{
		dockLabel := "Dock Left"
		if !ins.dockRight {
			dockLabel = "Dock Right"
		}
		dr := ins.dockToggleRect()
		if rl.CheckCollisionPointRec(rl.GetMousePosition(), dr) {
			rl.DrawRectangleRec(dr, rl.NewColor(79, 70, 229, 100))
			rl.SetMouseCursor(rl.MouseCursorPointingHand)
		}
		dockStyle := ChromeBodyStyle()
		dockStyle.TextColor = inspTextBrt
		y := ChromeTextCenterY(dr, dockStyle)
		DrawChromeText(dockLabel, dr.X+6, y, dockStyle)
	}
	{
		fps := rl.GetFPS()
		fpsStr := fmt.Sprintf("%d fps", fps)
		var fpsColor rl.Color
		switch {
		case fps >= 55:
			fpsColor = rl.NewColor(100, 220, 120, 255)
		case fps >= 35:
			fpsColor = rl.NewColor(240, 200, 60, 255)
		default:
			fpsColor = rl.NewColor(220, 80, 80, 255)
		}
		fpsStyle := ChromeTitleStyle()
		fpsStyle.FontSize = 18
		fpsStyle.MinFontSize = 16
		fpsStyle.TextColor = fpsColor
		fw := MeasureChromeText(fpsStr, fpsStyle)
		fpsY := ChromeTextCenterY(rl.NewRectangle(float32(ins.panelX), 0, inspW, 28), fpsStyle)
		DrawChromeText(fpsStr, float32(ins.panelX)+float32(inspW)-fw-float32(inspPad), fpsY, fpsStyle)
	}

	statY := float32(34)
	if ins.selected != nil && ins.nodeStillLive(ins.selected) {
		tag := fmt.Sprintf("sel: %s", ins.selected.ID())
		inspDrawText(tag, float32(ins.panelX+inspPad), statY, inspTextDim)
		statY += 16
	}

	{
		s := PerfStats
		cacheLabel := "miss"
		if s.CacheHit {
			cacheLabel = "HIT"
		}
		timingLine := fmt.Sprintf("U:%.1f L:%.1f D:%.1f ms  %s",
			s.UpdateMs, s.LayoutMs, s.DrawMs, cacheLabel)
		inspDrawText(timingLine, float32(ins.panelX+inspPad), statY, inspTextDim)
		statY += 16
	}

	{
		goroutines := runtime.NumGoroutine()
		active := ActiveJobCount()
		qlen := 0
		if ins.doc != nil {
			qlen = ins.doc.TaskQueueLen()
		}
		concStr := fmt.Sprintf("goroutines: %d  jobs: %d  queue: %d", goroutines, active, qlen)
		inspDrawText(concStr, float32(ins.panelX+inspPad), statY, inspTextDim)
		statY += 16
	}

	inspDrawText("Shift+click row to fold branch", float32(ins.panelX+inspPad), statY, inspTextDim)

	treeTop := int32(inspHeaderH)
	treeBot := ins.panelH - int32(inspDetailH)
	beginScissorMode(ins.panelX, treeTop, int32(inspW), treeBot-treeTop)

	for i, row := range ins.rows {
		y := float32(treeTop) + float32(i)*float32(inspLineH) - ins.scrollY
		if y+float32(inspLineH) < float32(treeTop) || y > float32(treeBot) {
			continue
		}

		isSelected := row.node == ins.selected
		if isSelected {
			rl.DrawRectangle(ins.panelX, int32(y), int32(inspW), int32(inspLineH),
				rl.NewColor(80, 105, 200, 60))
		}

		col := inspTextBrt
		if row.node.IsHidden() {
			col = inspTextDim
		}

		rowRect := rl.NewRectangle(float32(ins.panelX), y, inspW, float32(inspLineH))
		treeStyle := ChromeInspectorTreeStyle()
		treeY := ChromeTextCenterY(rowRect, treeStyle)

		x := float32(ins.panelX+inspPad) + float32(row.depth)*14
		if row.hasChildren {
			glyph := "-"
			if row.collapsed {
				glyph = "+"
			}
			inspDrawTreeText(glyph, x, treeY, inspTextDim)
			x += inspFoldW
		} else {
			x += inspFoldW
		}
		inspDrawTreeText(row.label, x, treeY, col)
	}

	rl.EndScissorMode()

	rl.DrawRectangle(ins.panelX, treeBot, int32(inspW), 1, inspBorder)

	if ins.selected != nil && ins.nodeStillLive(ins.selected) {
		beginScissorMode(ins.panelX, treeBot, int32(inspW), int32(inspDetailH))
		ins.drawDetail(ins.selected, treeBot)
		rl.EndScissorMode()
	} else {
		y0 := float32(treeBot + 10)
		inspDrawText("Click a widget or tree row to inspect",
			float32(ins.panelX+inspPad), y0, inspTextDim)
		if ModalMgr.host.Open {
			inspDrawText(ModalMgr.DebugInfo(),
				float32(ins.panelX+inspPad), y0+float32(inspLineH),
				rl.NewColor(255, 200, 100, 220))
		}
		if IsContextMenuOpen() {
			inspDrawText("ctx: "+ContextMenuMgr.DebugInfo(),
				float32(ins.panelX+inspPad), y0+float32(inspLineH)*2,
				rl.NewColor(130, 220, 255, 220))
		}
		if n := ActiveToastCount(); n > 0 {
			inspDrawText(fmt.Sprintf("toasts: %d active", n),
				float32(ins.panelX+inspPad), y0+float32(inspLineH)*3,
				rl.NewColor(255, 180, 80, 220))
		}
	}
}

// IsVisible reports whether the inspector panel is shown.
func (ins *Inspector) IsVisible() bool { return ins.visible }

func (ins *Inspector) nodeStillLive(n Node) bool {
	if n == nil || ins.doc == nil || ins.doc.Root == nil {
		return false
	}
	var found bool
	var walk func(Node)
	walk = func(cur Node) {
		if cur == n {
			found = true
			return
		}
		for _, ch := range cur.Children() {
			walk(ch)
		}
	}
	walk(ins.doc.Root)
	return found
}

func (ins *Inspector) buildRows(n Node, depth int) {
	typeName := reflect.TypeOf(n).Elem().Name()
	if b, ok := n.(*Badge); ok {
		typeName = fmt.Sprintf("Badge (%s)", b.Shape)
	}
	hidden := ""
	if n.IsHidden() {
		hidden = " [H]"
	}
	label := fmt.Sprintf("%s  %s%s", typeName, n.ID(), hidden)
	kids := n.Children()
	collapsed := ins.collapsed[n.ID()]
	ins.rows = append(ins.rows, treeRow{
		node:        n,
		depth:       depth,
		label:       label,
		hasChildren: len(kids) > 0,
		collapsed:   collapsed,
	})
	if collapsed {
		return
	}
	for _, child := range kids {
		ins.buildRows(child, depth+1)
	}
}

func (ins *Inspector) hitTest(n Node, p rl.Vector2) Node {
	if n.IsHidden() {
		return nil
	}
	if !rl.CheckCollisionPointRec(p, n.Bounds()) {
		return nil
	}
	var childHit Node
	for i := len(n.Children()) - 1; i >= 0; i-- {
		if hit := ins.hitTest(n.Children()[i], p); hit != nil {
			childHit = hit
			break
		}
	}
	if lt, ok := n.(*ListTile); ok {
		return lt.InspectorPickTarget(childHit)
	}
	if childHit != nil {
		return childHit
	}
	return n
}

func (ins *Inspector) drawDetail(n Node, top int32) {
	x := float32(ins.panelX + inspPad)
	y := float32(top) + 8
	lineH := float32(inspLineH)

	drawKV := func(key, val string) {
		keyStyle := ChromeTitleStyle()
		keyStyle.TextColor = inspKeyColor
		valStyle := ChromeInspectorTreeStyle()
		valStyle.TextColor = inspValColor
		DrawChromeText(key+":", x, y, keyStyle)
		kw := MeasureChromeText(key+":", keyStyle)
		DrawChromeText(val, x+kw+4, y, valStyle)
		y += lineH
	}

	b := n.Bounds()
	typeName := reflect.TypeOf(n).Elem().Name()
	if bdg, ok := n.(*Badge); ok {
		typeName = fmt.Sprintf("Badge (%s)", bdg.Shape)
	}

	drawKV("type", typeName)
	drawKV("id", n.ID())
	drawKV("style", n.StyleName())
	drawKV("bounds", fmt.Sprintf("(%.0f,%.0f) %.0fx%.0f", b.X, b.Y, b.Width, b.Height))

	type dbgElement interface {
		DbgLayoutDirty() bool
		DbgDrawDirty() bool
		DbgCachePolicy() string
		DbgCacheDirty() bool
	}
	if de, ok := n.(dbgElement); ok {
		drawKV("layoutDirty", fmt.Sprintf("%v", de.DbgLayoutDirty()))
		drawKV("drawDirty", fmt.Sprintf("%v", de.DbgDrawDirty()))
		drawKV("cache", de.DbgCachePolicy())
		if de.DbgCacheDirty() {
			drawKV("cacheDirty", "true")
		}
	} else {
		drawKV("dirty", fmt.Sprintf("%v", n.IsDirty()))
	}
	drawKV("usesScissor", fmt.Sprintf("%v", n.UsesScissor()))
	drawKV("hidden", fmt.Sprintf("%v", n.IsHidden()))
	drawKV("interactive", fmt.Sprintf("%v", n.IsInteractive()))
	drawKV("zIndex", fmt.Sprintf("%d", n.GetZIndex()))
	drawKV("flexGrow", fmt.Sprintf("%.2f", n.GetFlexGrow()))
	drawKV("children", fmt.Sprintf("%d", len(n.Children())))

	switch w := n.(type) {
	case *Tooltip:
		drawKV("text", fmt.Sprintf("%q", w.Text.Get()))
		drawKV("delay", fmt.Sprintf("%.2fs", w.Delay))
		drawKV("visible", fmt.Sprintf("%v", w.visible))
		drawKV("alpha", fmt.Sprintf("%.2f", w.alpha))
		if w.Target != nil {
			tName := reflect.TypeOf(w.Target).Elem().Name()
			drawKV("target", tName+":"+w.Target.ID())
		}
	case *TabView:
		drawKV("tabs", fmt.Sprintf("%d", len(w.tabs)))
		drawKV("active", fmt.Sprintf("%d", w.Active.Get()))
		if idx := w.Active.Get(); idx >= 0 && idx < len(w.tabs) {
			drawKV("active tab", fmt.Sprintf("%q", w.tabs[idx].title))
		}
		drawKV("vertical", fmt.Sprintf("%v", w.Vertical))
		for i, t := range w.tabs {
			state := ""
			if t.disabled {
				state = " [disabled]"
			}
			drawKV(fmt.Sprintf("  tab[%d]", i), fmt.Sprintf("%q%s", t.title, state))
		}
	case *Spinner:
		drawKV("active", fmt.Sprintf("%v", w.Active.Get()))
		drawKV("determinate", fmt.Sprintf("%v", w.Determinate))
		if w.Determinate {
			drawKV("progress", fmt.Sprintf("%.0f%%", w.Progress.Get()*100))
		} else {
			drawKV("angle", fmt.Sprintf("%.1f°", w.angle))
			drawKV("speed", fmt.Sprintf("%.0f°/s", w.Speed))
		}
		if w.Label != "" {
			pos := "below"
			if w.LabelPos == LabelRight {
				pos = "right"
			}
			drawKV("label", fmt.Sprintf("%q (%s)", w.Label, pos))
		}
	case *RadioGroup:
		dir := "vertical"
		if !w.Vertical {
			dir = "horizontal"
		}
		drawKV("layout", dir)
		drawKV("options", fmt.Sprintf("%d", len(w.Options)))
		idx := w.Selected.Get()
		if idx >= 0 && idx < len(w.Options) {
			drawKV("selected", fmt.Sprintf("[%d] %q", idx, w.Options[idx]))
		} else {
			drawKV("selected", "none")
		}
		for i, opt := range w.Options {
			suffix := ""
			if i == idx {
				suffix = " ●"
			}
			if i < len(w.Disabled) && w.Disabled[i] {
				suffix += " [disabled]"
			}
			drawKV(fmt.Sprintf("  [%d]", i), fmt.Sprintf("%q%s", opt, suffix))
		}
	case *TreeView:
		drawKV("nodes", fmt.Sprintf("%d total", w.TotalNodeCount()))
		drawKV("expanded", fmt.Sprintf("%d", w.ExpandedNodeCount()))
		drawKV("visible rows", fmt.Sprintf("%d", len(w.rows)))
		drawKV("scrollY", fmt.Sprintf("%.0fpx", w.ScrollY))
		sel := w.Selected.Get()
		if sel != nil {
			drawKV("selected", fmt.Sprintf("%q", sel.Label))
			if path := w.SelectedPath(); path != "" {
				drawKV("path", path)
			}
		} else {
			drawKV("selected", "none")
		}
	case *Form:
		dir := "two-column"
		if w.Vertical {
			dir = "vertical"
		}
		drawKV("layout", dir)
		drawKV("fields", fmt.Sprintf("%d", w.FieldCount()))
		drawKV("labelW", fmt.Sprintf("%.0fpx", w.LabelW))
		drawKV("rowH", fmt.Sprintf("%.0fpx", w.RowH))
		for i, ff := range w.fields {
			annotation := ""
			if ff.errorMsg != "" {
				annotation = " [error: " + ff.errorMsg + "]"
			}
			widgetType := "nil"
			if ff.widget != nil {
				widgetType = reflect.TypeOf(ff.widget).Elem().Name()
			}
			drawKV(fmt.Sprintf("  [%d]", i), fmt.Sprintf("%q → %s%s", ff.label, widgetType, annotation))
		}
	case *Button:
		drawKV("text", fmt.Sprintf("%q", w.Text.Get()))
	case *Label:
		drawKV("text", fmt.Sprintf("%q", w.Text.Get()))
	case *TextInput:
		drawKV("text", fmt.Sprintf("%q", w.Text.Get()))
	case *SearchBar:
		drawKV("text", fmt.Sprintf("%q", w.GetText()))
		drawKV("query", fmt.Sprintf("%q", w.Query.Get()))
		if rem := w.DebounceRemaining(); rem > 0 {
			drawKV("debouncing", fmt.Sprintf("%.2fs left", rem))
		}
		drawKV("debounceDelay", fmt.Sprintf("%.2fs", w.DebounceDelay))
		drawKV("showIcon", fmt.Sprintf("%v", w.ShowIcon))
		if w.Placeholder != "" {
			drawKV("placeholder", fmt.Sprintf("%q", w.Placeholder))
		}
	case *Slider:
		drawKV("value", fmt.Sprintf("%.2f", w.Value.Get()))
		drawKV("min/max", fmt.Sprintf("%.2f / %.2f", w.MinValue, w.MaxValue))
	case *Checkbox:
		drawKV("checked", fmt.Sprintf("%v", w.Value.Get()))
	case *Toggle:
		drawKV("on", fmt.Sprintf("%v", w.Value.Get()))
	case *Dropdown:
		idx := w.SelectedIndex.Get()
		sel := ""
		if idx >= 0 && idx < len(w.Options) {
			sel = w.Options[idx]
		}
		drawKV("selected", fmt.Sprintf("[%d] %q", idx, sel))
	case *DatePicker:
		t := w.Value.Get()
		if t.IsZero() {
			drawKV("value", "zero (placeholder)")
		} else {
			drawKV("value", t.In(time.Local).Format("2006-01-02"))
		}
		vm := w.VisibleMonth()
		drawKV("visible month", vm.Format("January 2006"))
		drawKV("popup", fmt.Sprintf("%v", DatePickerMgr.isTarget(w)))
	case *FilePicker:
		drawKV("mode", string(w.Mode()))
		drawKV("current path", fmt.Sprintf("%q", w.CurrentPath.Get()))
		drawKV("selection", fmt.Sprintf("%q", w.Selection.Get()))
		drawKV("loading", fmt.Sprintf("%v", w.IsLoading()))
	case *Panel:
		drawKV("title", fmt.Sprintf("%q", w.Title))
	case *Header:
		drawKV("title", fmt.Sprintf("%q", w.Title))
		if w.Subtitle != "" {
			drawKV("subtitle", fmt.Sprintf("%q", w.Subtitle))
		}
	case *ColorPicker:
		c := w.Value.Get()
		h, s, v, _ := colorToHSVA(c)
		drawKV("rgba", fmt.Sprintf("(%d,%d,%d,%d)", c.R, c.G, c.B, c.A))
		drawKV("hsv", fmt.Sprintf("(%.0f°, %.0f%%, %.0f%%)", h, s*100, v*100))
		drawKV("hex", fmt.Sprintf("#%02X%02X%02X", c.R, c.G, c.B))
		drawKV("open", fmt.Sprintf("%v", ColorPickerMgr.open))
		drawKV("showAlpha", fmt.Sprintf("%v", w.ShowAlpha))
	case *Badge:
		drawKV("text", fmt.Sprintf("%q", w.Text.Get()))
		drawKV("variant", fmt.Sprintf("%s (%d)", badgeStyleName[w.Variant], int(w.Variant)))
		drawKV("shape", w.Shape.String())
		drawKV("closeButton", fmt.Sprintf("%v", w.CloseButton))
		drawKV("autoSize", fmt.Sprintf("%v", w.autoSize))
	case *Accordion:
		drawKV("title", fmt.Sprintf("%q", w.Title))
		drawKV("expanded", fmt.Sprintf("%v", w.Expanded.Get()))
		drawKV("animating", fmt.Sprintf("%v", w.tween != nil && w.tween.IsActive))
		drawKV("animH", fmt.Sprintf("%.1f / %.1f", w.animH, w.contentH))
		drawKV("children", fmt.Sprintf("%d", len(w.children)))
	case *Stepper:
		drawKV("steps", fmt.Sprintf("%d", len(w.Steps)))
		drawKV("current", fmt.Sprintf("%d", w.CurrentStep.Get()))
		drawKV("animProgress", fmt.Sprintf("%.3f", w.animProgress))
		drawKV("animating", fmt.Sprintf("%v", w.tween != nil && w.tween.IsActive))
		direction := "horizontal"
		if w.Direction == StepperVertical {
			direction = "vertical"
		}
		drawKV("direction", direction)
		drawKV("clickable", fmt.Sprintf("%v", w.Clickable))
		drawKV("hoverStep", fmt.Sprintf("%d", w.hoverStep))
	case *KeyboardShortcut:
		drawKV("action", fmt.Sprintf("%q", w.Action.Get()))
		drawKV("keys", fmt.Sprintf("%q", w.Keys.Get()))
		drawKV("tokens", fmt.Sprintf("%v", w.keyTokens))
		drawKV("symbol", fmt.Sprintf("%q", w.Symbol))
		drawKV("spread", fmt.Sprintf("%v", w.Spread))
		drawKV("clickable", fmt.Sprintf("%v", w.OnClick != nil))
		drawKV("autoSize", fmt.Sprintf("%v", w.autoSize))
	case dataTableWidget:
		cols, rows, sc, sa, sel := w.dtInfo()
		drawKV("columns", fmt.Sprintf("%d", cols))
		drawKV("rows", fmt.Sprintf("%d", rows))
		sortStr := "none"
		if sc >= 0 {
			dir := "↑ asc"
			if !sa {
				dir = "↓ desc"
			}
			sortStr = fmt.Sprintf("col %d %s", sc, dir)
		}
		drawKV("sort", sortStr)
		drawKV("selected", fmt.Sprintf("%d", sel))
	case *SplitView:
		dir, ratioStr, sizeStr := w.SplitViewInfo()
		drawKV("direction", dir)
		drawKV("ratio", ratioStr)
		drawKV("size", sizeStr)
		drawKV("splitterW", fmt.Sprintf("%.0fpx", w.SplitterW))
		drawKV("minFirst", fmt.Sprintf("%.0fpx", w.MinFirst))
		drawKV("minSecond", fmt.Sprintf("%.0fpx", w.MinSecond))
		drawKV("dragging", fmt.Sprintf("%v", w.dragging))
		drawKV("hover", fmt.Sprintf("%v", w.hoverSplit))
		if w.first != nil {
			drawKV(w.firstPaneLabel(), reflect.TypeOf(w.first).Elem().Name()+":"+w.first.ID())
		}
		if w.second != nil {
			drawKV(w.secondPaneLabel(), reflect.TypeOf(w.second).Elem().Name()+":"+w.second.ID())
		}
	case *Toolbar:
		groups, items, active := w.ToolbarInfo()
		drawKV("groups", groups)
		drawKV("totalItems", items)
		drawKV("ribbon", fmt.Sprintf("%v", w.Ribbon))
		drawKV("overflow", fmt.Sprintf("%v", w.Overflow))
		if w.Ribbon {
			drawKV("activeGroup", active)
		}
	case *FloatLayer:
		drawKV("role", "viewport overlay host (full viewport bounds)")
	}
}
