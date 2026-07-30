// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Popup layout constants (shared by [DatePickerMgr] draw/update).
const (
	dpPad        float32 = 14
	dpHeaderH    float32 = 40
	dpDowH       float32 = 24
	dpCellW      float32 = 44
	dpCellH      float32 = 44
	dpNavBtnW    float32 = 36
	dpPopInnerW  float32 = dpPad*2 + float32(calDaysPerWeek)*dpCellW
	dpPopInnerH  float32 = dpPad + dpHeaderH + dpDowH + float32(calWeeks)*dpCellH + dpPad
	dpPopMaxScale float32 = 1.15 // popup width may grow with field, up to this × base
	dpFadeIn     float32 = 0.12
	dpFadeOut    float32 = 0.10
	dpSkipFrames int     = 2
	dpDayFontSize float32 = 16
)

// DatePicker is a compact field that shows the current [Signal] date and opens
// a full calendar overlay when clicked. The popup is managed by [DatePickerMgr]
// so it is never clipped by parent Viewports, Panels, or Modals.
//
// Selected dates are always stored as local-midnight [time.Time] values; any
// time-of-day component on write is truncated.
//
// Keyboard (while popup is open): arrow keys move the focused day, Page Up /
// Page Down change month, Enter commits and closes, Escape cancels (restores
// the value from when the popup opened) and closes. Clicking a day commits;
// clicking outside the popup cancels.
//
// # Wiring
//
// Call [DatePickerMgr.Update] once per frame after [Document.Root] Update and
// [DatePickerMgr.Draw] after other overlay managers in main.go.
//
// # LLM Prompt Template
//
//	d := ui.NewSignal(time.Now())
//	dp := ui.NewDatePicker("due", d, 0, 0, 0, 38)
//	form.AddChild(dp)
//
// Demo scenes: **Batch 3 DatePicker**, **Batch 11 DateRangePicker** (range variant).
type DatePicker struct {
	Element

	// Value is the reactive selected calendar date. Subscribers fire on every Set.
	Value *Signal[time.Time]

	// DateFormat is the Go time layout string used to format the field text
	// (default "2006-01-02").
	DateFormat string

	// Placeholder is shown when Value is the zero time.
	Placeholder string

	hovered bool

	// visiblePage is the first day of the month page last shown while this
	// picker was the active popup target (updated by DatePickerMgr).
	visiblePage time.Time
}

// NewDatePicker creates a DatePicker bound to value. The id must be unique
// within the document for inspector selection. Pass width 0 to stretch in flex
// layouts; height 34–40 matches text inputs and dropdowns.
func NewDatePicker(id string, value *Signal[time.Time], x, y, w, h float32) *DatePicker {
	if value == nil {
		panic("ui.NewDatePicker: value must not be nil")
	}
	dp := &DatePicker{
		Element:     NewElement(id, x, y, w, h),
		Value:       value,
		DateFormat:  "2006-01-02",
		Placeholder: "Select date…",
	}
	dp.styleName = "datepicker"
	dp.ZIndex = 5
	dp.Value.Subscribe(func() { dp.MarkDrawDirty() })
	return dp
}

// VisibleMonth returns the calendar month page (first local midnight of that
// month) last driven by the popup while this picker was active. When the popup
// has never been opened, it defaults to the month of the current Value (or the
// current month if Value is zero).
func (dp *DatePicker) VisibleMonth() time.Time {
	if !dp.visiblePage.IsZero() {
		return dp.visiblePage
	}
	t := dateTruncLocal(dp.Value.Get())
	if t.IsZero() {
		now := time.Now()
		y, m, _ := now.Date()
		return monthStart(y, m)
	}
	y, m, _ := t.Date()
	return monthStart(y, m)
}

// IsInteractive implements [Node].
func (dp *DatePicker) IsInteractive() bool { return true }

// IsPopupOpen reports whether this picker currently owns the global calendar popup.
func (dp *DatePicker) IsPopupOpen() bool {
	return DatePickerMgr.isTarget(dp)
}

// Update handles hover and opening the calendar via [DatePickerMgr].
func (dp *DatePicker) Update(_ float32) {
	if dp.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	b := dp.Bounds()
	prevH := dp.hovered
	dp.hovered = rl.CheckCollisionPointRec(mouse, b)
	if dp.hovered != prevH {
		dp.MarkDrawDirty()
	}

	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if DatePickerMgr.open && !DatePickerMgr.closing && DatePickerMgr.target != dp {
			return
		}
		if dp.hovered {
			if DatePickerMgr.isTarget(dp) {
				DatePickerMgr.cancelAndClose()
				return
			}
			DatePickerMgr.Open(dp, b.X, b.Y+b.Height+4)
		}
	}
}

// OverlayExemptRects implements overlayExempter.
func (dp *DatePicker) OverlayExemptRects() []rl.Rectangle {
	if !DatePickerMgr.isTarget(dp) {
		return nil
	}
	return []rl.Rectangle{dp.Bounds(), DatePickerMgr.PopupBounds()}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (dp *DatePicker) ClearOverlayPointerState() {
	if !dp.hovered {
		return
	}
	dp.hovered = false
	dp.MarkDrawDirty()
}

// Layout is a no-op; bounds come from the parent layout engine.
func (dp *DatePicker) Layout() { dp.layoutDirty = false }

// Draw renders the closed field (popup is drawn by [DatePickerMgr.Draw]).
func (dp *DatePicker) Draw() {
	dp.drawInternal()
	dp.drawDirty = false
}

func (dp *DatePicker) drawInternal() {
	if dp.IsHidden() {
		return
	}
	style := dp.GetStyle()
	bounds := fieldPaintBounds(dp.Bounds(), style)
	open := DatePickerMgr.isTarget(dp)

	bg := style.BackgroundColor
	if open {
		bg = rl.ColorBrightness(bg, -0.04)
	} else if dp.hovered {
		bg = rl.ColorBrightness(bg, -0.06)
	}
	borderCol := style.BorderColor
	if open || dp.hovered {
		borderCol = rl.NewColor(99, 90, 249, 255)
	}

	r := float32(0)
	if style.CornerRadius > 0 {
		half := bounds.Height / 2
		if bounds.Width/2 < half {
			half = bounds.Width / 2
		}
		if half > 0 {
			r = style.CornerRadius / half
			if r > 1 {
				r = 1
			}
		}
	}
	if r > 0 {
		rl.DrawRectangleRounded(bounds, r, 8, bg)
		if style.BorderWidth > 0 {
			rl.DrawRectangleRoundedLinesEx(bounds, r, 8, style.BorderWidth, borderCol)
		}
	} else {
		rl.DrawRectangleRec(bounds, bg)
		if style.BorderWidth > 0 {
			rl.DrawRectangleLinesEx(bounds, style.BorderWidth, borderCol)
		}
	}

	val := dp.Value.Get()
	label := dp.Placeholder
	drawStyle := style
	if val.IsZero() {
		drawStyle.TextColor = rl.NewColor(140, 142, 155, 255)
	} else {
		label = val.In(time.Local).Format(dp.DateFormat)
	}
	const fieldTextPad = float32(12)
	const fieldArrowGutter = float32(32)
	textMaxW := bounds.Width - fieldTextPad - fieldArrowGutter
	if textMaxW < 8 {
		textMaxW = 8
	}
	drawTextS(truncateTextS(label, textMaxW, drawStyle), int32(bounds.X+fieldTextPad), TextPosY(bounds, style), drawStyle)

	// Calendar glyph / chevron
	arrowR := float32(9)
	arrowX := bounds.X + bounds.Width - 26
	arrowY := bounds.Y + bounds.Height/2
	arrowCol := rl.NewColor(110, 112, 128, 255)
	if open {
		rl.DrawTriangle(
			rl.NewVector2(arrowX, arrowY+arrowR*0.5),
			rl.NewVector2(arrowX+arrowR*2, arrowY+arrowR*0.5),
			rl.NewVector2(arrowX+arrowR, arrowY-arrowR*0.5),
			arrowCol,
		)
	} else {
		rl.DrawTriangle(
			rl.NewVector2(arrowX, arrowY-arrowR*0.5),
			rl.NewVector2(arrowX+arrowR*2, arrowY-arrowR*0.5),
			rl.NewVector2(arrowX+arrowR, arrowY+arrowR*0.5),
			arrowCol,
		)
	}
}

// ── datePickerManager (singleton overlay) ───────────────────────────────────

type datePickerManager struct {
	open    bool
	closing bool
	target  *DatePicker

	original  time.Time
	focusDate time.Time
	hoverDate time.Time // mouse-over day (light grey); distinct from focus/selection

	viewYear  int
	viewMonth time.Month

	popX, popY float32
	popW, popH float32
	cellW      float32 // day column width (may widen with anchor field)

	// Last anchor field rect — repopup when the field moves or grows.
	lastFieldX float32
	lastFieldY float32
	lastFieldW float32
	lastFieldH float32

	fade       float32
	skipFrames int

	// Cached geometry for hit testing (screen space).
	prevBtn, nextBtn rl.Rectangle
	gridOriginX      float32
	gridOriginY      float32
}

// DatePickerMgr is the package-level singleton. Wire [DatePickerMgr.Update] and
// [DatePickerMgr.Draw] from main.go alongside other overlay managers.
var DatePickerMgr = &datePickerManager{}

func (m *datePickerManager) isTarget(dp *DatePicker) bool {
	return m.open && m.target == dp && !m.closing
}

// IsOpen reports whether a calendar popup is visible (including fade-out).
func (m *datePickerManager) IsOpen() bool { return m.open }

// PopupBounds returns the screen-space calendar popup rectangle while open.
func (m *datePickerManager) PopupBounds() rl.Rectangle {
	if !m.open {
		return rl.Rectangle{}
	}
	return rl.NewRectangle(m.popX, m.popY, m.popW, m.popH)
}

// IsAnimating reports fade-in, fade-out, or open settling frames.
func (m *datePickerManager) IsAnimating() bool {
	if !m.open {
		return false
	}
	return m.closing || m.fade < 1 || m.skipFrames > 0
}

func (m *datePickerManager) snapClose() {
	m.open = false
	m.closing = false
	m.target = nil
	m.fade = 0
	m.skipFrames = 0
}

func (m *datePickerManager) setVisiblePage(dp *DatePicker) {
	dp.visiblePage = monthStart(m.viewYear, m.viewMonth)
	dp.MarkDrawDirty()
}

// Open shows the calendar anchored below (or above) the given picker field.
func (m *datePickerManager) Open(dp *DatePicker, anchorX, anchorY float32) {
	if dp == nil || dp.Value == nil {
		return
	}
	if DateRangePickerMgr.IsOpen() {
		DateRangePickerMgr.snapClose()
	}
	// Revert another open picker before stealing focus.
	if m.open && m.target != nil && m.target != dp {
		m.target.Value.Set(m.original)
		m.target.MarkDrawDirty()
		m.snapClose()
	}

	m.target = dp
	m.original = dateTruncLocal(dp.Value.Get())
	if m.original.IsZero() {
		now := time.Now()
		m.focusDate = dateTruncLocal(now)
	} else {
		m.focusDate = m.original
	}
	y, mo, _ := m.focusDate.Date()
	m.viewYear, m.viewMonth = y, mo
	m.setVisiblePage(dp)

	field := dp.Bounds()
	m.layoutPopupAnchored(field, anchorX, anchorY)

	m.open = true
	m.closing = false
	m.fade = 0
	m.skipFrames = dpSkipFrames
	m.layoutHitRects()
}

// layoutPopupAnchored sizes the calendar to the trigger field (slightly wider when
// the field grows) and clamps position to the screen.
func (m *datePickerManager) layoutPopupAnchored(field rl.Rectangle, anchorX, anchorY float32) {
	m.cellW = dpCellW
	m.popW = dpPopInnerW
	if field.Width > m.popW {
		m.popW = field.Width
		maxW := dpPopInnerW * dpPopMaxScale
		if m.popW > maxW {
			m.popW = maxW
		}
	}
	gridW := m.popW - 2*dpPad
	m.cellW = gridW / float32(calDaysPerWeek)
	if m.cellW < dpCellW {
		m.cellW = dpCellW
		m.popW = 2*dpPad + float32(calDaysPerWeek)*m.cellW
	}
	m.popH = dpPad + dpHeaderH + dpDowH + float32(calWeeks)*dpCellH + dpPad

	sw := float32(rl.GetScreenWidth())
	sh := float32(rl.GetScreenHeight())
	px := anchorX
	py := anchorY
	if px+m.popW > sw-8 {
		px = sw - m.popW - 8
	}
	if px < 8 {
		px = 8
	}
	if py+m.popH > sh-40 {
		py = field.Y - 4 - m.popH
	}
	if py < 8 {
		py = 8
	}
	m.popX = px
	m.popY = py
	m.lastFieldX = field.X
	m.lastFieldY = field.Y
	m.lastFieldW = field.Width
	m.lastFieldH = field.Height
}

func (m *datePickerManager) syncPopupToField() {
	if m.target == nil || !m.open || m.closing {
		return
	}
	field := m.target.Bounds()
	if field.X == m.lastFieldX && field.Y == m.lastFieldY &&
		field.Width == m.lastFieldW && field.Height == m.lastFieldH {
		return
	}
	anchorY := field.Y + field.Height + 4
	m.layoutPopupAnchored(field, field.X, anchorY)
	m.layoutHitRects()
}

func (m *datePickerManager) layoutHitRects() {
	innerLeft := m.popX + dpPad
	headerY := m.popY + dpPad
	m.prevBtn = rl.NewRectangle(innerLeft, headerY+4, dpNavBtnW, dpHeaderH-8)
	m.nextBtn = rl.NewRectangle(m.popX+m.popW-dpPad-dpNavBtnW, headerY+4, dpNavBtnW, dpHeaderH-8)
	m.gridOriginX = innerLeft
	m.gridOriginY = headerY + dpHeaderH + dpDowH
}

func (m *datePickerManager) monthHeaderText() string {
	return fmt.Sprintf("%s %d", m.viewMonth.String(), m.viewYear)
}

func (m *datePickerManager) cellRect(col, row int) rl.Rectangle {
	cw := m.cellW
	if cw <= 0 {
		cw = dpCellW
	}
	x := m.gridOriginX + float32(col)*cw
	y := m.gridOriginY + float32(row)*dpCellH
	return rl.NewRectangle(x, y, cw, dpCellH)
}

func (m *datePickerManager) cancelAndClose() {
	if m.target != nil {
		m.target.Value.Set(m.original)
		m.target.MarkDrawDirty()
	}
	m.closing = true
}

func (m *datePickerManager) commitAndClose() {
	if m.target != nil {
		m.target.Value.Set(dateTruncLocal(m.focusDate))
		m.target.MarkDrawDirty()
	}
	m.closing = true
}

// Update drives fade animation, keyboard, and mouse for the open popup.
// Call once per frame after [Document.Root].Update.
func (m *datePickerManager) Update(dt float32) {
	if !m.open {
		return
	}

	if m.closing {
		m.fade -= dt / dpFadeOut
		if m.fade <= 0 {
			m.snapClose()
		}
		return
	}

	if m.fade < 1 {
		m.fade += dt / dpFadeIn
		if m.fade > 1 {
			m.fade = 1
		}
	}

	if m.skipFrames > 0 {
		m.skipFrames--
		return
	}

	m.syncPopupToField()

	mouse := rl.GetMousePosition()
	popRec := rl.NewRectangle(m.popX, m.popY, m.popW, m.popH)
	fieldRec := rl.Rectangle{}
	if m.target != nil {
		fieldRec = m.target.Bounds()
	}

	grid := buildMonthGrid(m.viewYear, m.viewMonth)
	// Hover: light grey disc (hoverDate), separate from keyboard focusDate.
	hoverHit := false
	for i, cell := range grid {
		if !cell.InMonth {
			continue
		}
		row, col := i/calDaysPerWeek, i%calDaysPerWeek
		cr := m.cellRect(col, row)
		if pointInDayCell(mouse, cr) {
			hoverHit = true
			if m.hoverDate != cell.Date {
				m.hoverDate = cell.Date
				if m.target != nil {
					m.target.MarkDrawDirty()
				}
			}
			break
		}
	}
	if !hoverHit && !m.hoverDate.IsZero() {
		m.hoverDate = time.Time{}
		if m.target != nil {
			m.target.MarkDrawDirty()
		}
	}

	// Escape cancels.
	if rl.IsKeyPressed(rl.KeyEscape) {
		m.cancelAndClose()
		return
	}
	// Enter commits.
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		m.commitAndClose()
		return
	}
	// Month navigation.
	if rl.IsKeyPressed(rl.KeyPageUp) {
		m.shiftMonth(-1)
	}
	if rl.IsKeyPressed(rl.KeyPageDown) {
		m.shiftMonth(1)
	}

	// Arrow keys: move by day; view month follows when crossing boundary.
	if rl.IsKeyPressed(rl.KeyRight) {
		m.moveFocusByDays(1)
	}
	if rl.IsKeyPressed(rl.KeyLeft) {
		m.moveFocusByDays(-1)
	}
	if rl.IsKeyPressed(rl.KeyDown) {
		m.moveFocusByDays(7)
	}
	if rl.IsKeyPressed(rl.KeyUp) {
		m.moveFocusByDays(-7)
	}

	// Mouse: month buttons, day commit, click-outside cancel.
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		switch {
		case rl.CheckCollisionPointRec(mouse, m.prevBtn):
			m.shiftMonth(-1)
			return
		case rl.CheckCollisionPointRec(mouse, m.nextBtn):
			m.shiftMonth(1)
			return
		}

		for i, cell := range grid {
			if !cell.InMonth {
				continue
			}
			row, col := i/calDaysPerWeek, i%calDaysPerWeek
			cr := m.cellRect(col, row)
			if pointInDayCell(mouse, cr) {
				m.focusDate = cell.Date
				m.commitAndClose()
				return
			}
		}

		// Click outside popup and field → cancel.
		if !rl.CheckCollisionPointRec(mouse, popRec) && !rl.CheckCollisionPointRec(mouse, fieldRec) {
			m.cancelAndClose()
			return
		}
	}
}

func (m *datePickerManager) shiftMonth(delta int) {
	_, _, pref := m.focusDate.Date()
	first := monthStart(m.viewYear, m.viewMonth).AddDate(0, delta, 0)
	m.viewYear, m.viewMonth, _ = first.Date()
	// Keep same day-of-month when possible inside new month.
	m.focusDate = clampDayInMonth(m.viewYear, m.viewMonth, pref)
	if m.target != nil {
		m.setVisiblePage(m.target)
	}
}

func (m *datePickerManager) moveFocusByDays(delta int) {
	next := m.focusDate.AddDate(0, 0, delta)
	next = dateTruncLocal(next)
	y, mo, _ := next.Date()
	if y != m.viewYear || mo != m.viewMonth {
		m.viewYear, m.viewMonth = y, mo
		if m.target != nil {
			m.setVisiblePage(m.target)
		}
	}
	m.focusDate = next
	if m.target != nil {
		m.target.MarkDrawDirty()
	}
}

// Draw renders the popup overlay at the end of the frame.
func (m *datePickerManager) Draw() {
	if !m.open || m.fade <= 0 {
		return
	}
	mul := func(c rl.Color) rl.Color {
		c.A = uint8(float32(c.A) * m.fade)
		return c
	}

	popRec := rl.NewRectangle(
		float32(int32(m.popX+0.5)),
		float32(int32(m.popY+0.5)),
		float32(int32(m.popW+0.5)),
		float32(int32(m.popH+0.5)),
	)
	shorter := popRec.Width
	if popRec.Height < shorter {
		shorter = popRec.Height
	}
	popRound := float32(0.06)
	if shorter > 0 {
		popRound = 10 / (shorter / 2)
		if popRound > 1 {
			popRound = 1
		}
	}

	rl.DrawRectangleRounded(rl.NewRectangle(popRec.X+2, popRec.Y+4, popRec.Width, popRec.Height), popRound, 12, mul(rl.NewColor(0, 0, 20, 55)))
	rl.DrawRectangleRounded(popRec, popRound, 12, mul(rl.NewColor(255, 255, 255, 255)))
	drawPopupRoundedBorder(popRec, 10, 1.5, mul(rl.NewColor(198, 200, 214, 255)))

	title := m.monthHeaderText()
	tw := MeasureText(title, 17)
	tx := m.popX + (m.popW-tw)/2
	DrawText(title, tx, m.popY+dpPad+8, 17, mul(rl.NewColor(40, 42, 58, 255)))

	// Nav buttons
	drawMiniNav(m.prevBtn, true, mul)
	drawMiniNav(m.nextBtn, false, mul)

	// Weekday labels
	dows := []string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	cw := m.cellW
	if cw <= 0 {
		cw = dpCellW
	}
	for c, d := range dows {
		cx := m.gridOriginX + float32(c)*cw
		cy := m.popY + dpPad + dpHeaderH
		tw2 := MeasureText(d, 12)
		DrawText(d, cx+(cw-tw2)/2, cy+4, 12, mul(rl.NewColor(120, 122, 140, 255)))
	}

	today := dateTruncLocal(time.Now())
	sel := time.Time{}
	if m.target != nil {
		sel = dateTruncLocal(m.target.Value.Get())
	}

	grid := buildMonthGrid(m.viewYear, m.viewMonth)
	for i, cell := range grid {
		row, col := i/calDaysPerWeek, i%calDaysPerWeek
		cr := m.cellRect(col, row)
		if !cell.InMonth {
			continue
		}
		isToday := cell.Date.Equal(today)
		isSel := !sel.IsZero() && cell.Date.Equal(sel)
		isFocus := cell.Date.Equal(m.focusDate)
		isHover := !m.hoverDate.IsZero() && cell.Date.Equal(m.hoverDate)

		paint := CalendarDayPaint{}
		if isSel {
			paint.Fill = rl.NewColor(238, 242, 255, 255)
			paint.Ring = rl.NewColor(79, 70, 229, 255)
			paint.RingWidth = 1.5
		} else if isHover {
			paint.Fill = rl.NewColor(226, 228, 234, 255)
		} else if isFocus {
			paint.Fill = rl.NewColor(237, 235, 255, 255)
			paint.Ring = rl.NewColor(99, 90, 249, 255)
			paint.RingWidth = 2
		} else if isToday {
			paint.Fill = rl.NewColor(224, 231, 255, 255)
		}
		paintCalendarDay(cr, paint, mul)

		ds := fmt.Sprintf("%d", cell.Date.Day())
		dtw := MeasureText(ds, dpDayFontSize)
		tc := mul(rl.NewColor(38, 40, 52, 255))
		if isSel || (isFocus && !isHover) {
			tc = mul(rl.NewColor(55, 48, 163, 255))
		}
		DrawText(ds, cr.X+(cr.Width-dtw)/2, cr.Y+(dpCellH-float32(dpDayFontSize))/2, dpDayFontSize, tc)
	}
}

func drawMiniNav(r rl.Rectangle, left bool, mul func(rl.Color) rl.Color) {
	midY := r.Y + r.Height/2
	rl.DrawRectangleRounded(r, 0.35, 6, mul(rl.NewColor(240, 240, 246, 255)))
	rl.DrawRectangleRoundedLinesEx(r, 0.35, 6, 1, mul(rl.NewColor(200, 202, 215, 255)))
	c := mul(rl.NewColor(70, 72, 90, 255))
	if left {
		rl.DrawTriangle(
			rl.NewVector2(r.X+r.Width*0.62, midY-r.Height*0.22),
			rl.NewVector2(r.X+r.Width*0.62, midY+r.Height*0.22),
			rl.NewVector2(r.X+r.Width*0.32, midY),
			c,
		)
	} else {
		rl.DrawTriangle(
			rl.NewVector2(r.X+r.Width*0.38, midY-r.Height*0.22),
			rl.NewVector2(r.X+r.Width*0.38, midY+r.Height*0.22),
			rl.NewVector2(r.X+r.Width*0.68, midY),
			c,
		)
	}
}
