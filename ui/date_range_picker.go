// Package ui (continued)
package ui

import (
	"fmt"
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// DateRangePicker is a compact field for start/end calendar dates. The range
// calendar popup is managed by [DateRangePickerMgr] (same wiring as [DatePickerMgr]
// in main.go). Click a start day, then an end day; the range commits and closes.
//
// # LLM Prompt Template
//
//	start, end := ui.NewSignal(time.Now()), ui.NewSignal(time.Now().AddDate(0,0,7))
//	drp := ui.NewDateRangePicker("range", start, end, 0, 0, 0, 40)
type DateRangePicker struct {
	Element
	Start        *Signal[time.Time]
	End          *Signal[time.Time]
	DateFormat   string
	Placeholder  string
	hovered      bool
	visiblePage  time.Time
}

// NewDateRangePicker creates a range picker. start and end must not be nil.
func NewDateRangePicker(id string, start, end *Signal[time.Time], x, y, w, h float32) *DateRangePicker {
	if start == nil || end == nil {
		panic("ui.NewDateRangePicker: start and end must not be nil")
	}
	drp := &DateRangePicker{
		Element:     NewElement(id, x, y, w, h),
		Start:       start,
		End:         end,
		DateFormat:  "2006-01-02",
		Placeholder: "Select range…",
	}
	drp.styleName = "daterangepicker"
	drp.ZIndex = 5
	start.Subscribe(func() { drp.MarkDrawDirty() })
	end.Subscribe(func() { drp.MarkDrawDirty() })
	return drp
}

func normalizeDateRange(a, b time.Time) (time.Time, time.Time) {
	a = dateTruncLocal(a)
	b = dateTruncLocal(b)
	if a.IsZero() || b.IsZero() {
		return a, b
	}
	if b.Before(a) {
		a, b = b, a
	}
	return a, b
}

func (drp *DateRangePicker) formatFieldLabel() string {
	s, e := drp.Start.Get(), drp.End.Get()
	s, e = dateTruncLocal(s), dateTruncLocal(e)
	if s.IsZero() && e.IsZero() {
		return drp.Placeholder
	}
	if s.IsZero() {
		return e.In(time.Local).Format(drp.DateFormat)
	}
	if e.IsZero() {
		return s.In(time.Local).Format(drp.DateFormat)
	}
	s, e = normalizeDateRange(s, e)
	return s.In(time.Local).Format(drp.DateFormat) + " – " + e.In(time.Local).Format(drp.DateFormat)
}

func (drp *DateRangePicker) IsPopupOpen() bool {
	return DateRangePickerMgr.isTarget(drp)
}

func (drp *DateRangePicker) Update(_ float32) {
	if drp.IsHidden() {
		return
	}
	mouse := rl.GetMousePosition()
	b := drp.Bounds()
	prev := drp.hovered
	drp.hovered = rl.CheckCollisionPointRec(mouse, b)
	if drp.hovered != prev {
		drp.MarkDrawDirty()
	}
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		if DateRangePickerMgr.open && !DateRangePickerMgr.closing && DateRangePickerMgr.target != drp {
			return
		}
		if drp.hovered {
			if DateRangePickerMgr.isTarget(drp) {
				DateRangePickerMgr.cancelAndClose()
				return
			}
			if DatePickerMgr.IsOpen() {
				DatePickerMgr.snapClose()
			}
			DateRangePickerMgr.Open(drp, b.X, b.Y+b.Height+4)
		}
	}
}

// OverlayExemptRects implements overlayExempter.
func (drp *DateRangePicker) OverlayExemptRects() []rl.Rectangle {
	if !DateRangePickerMgr.isTarget(drp) {
		return nil
	}
	return []rl.Rectangle{drp.Bounds(), DateRangePickerMgr.PopupBounds()}
}

// ClearOverlayPointerState implements overlayPointerClearer.
func (drp *DateRangePicker) ClearOverlayPointerState() {
	if !drp.hovered {
		return
	}
	drp.hovered = false
	drp.MarkDrawDirty()
}

func (drp *DateRangePicker) Layout() { drp.layoutDirty = false }

func (drp *DateRangePicker) Draw() {
	drp.drawInternal()
	drp.drawDirty = false
}

func (drp *DateRangePicker) drawInternal() {
	if drp.IsHidden() {
		return
	}
	style := drp.GetStyle()
	bounds := fieldPaintBounds(drp.Bounds(), style)
	open := DateRangePickerMgr.isTarget(drp)

	bg := style.BackgroundColor
	if open {
		bg = rl.ColorBrightness(bg, -0.04)
	} else if drp.hovered {
		bg = rl.ColorBrightness(bg, -0.06)
	}
	borderCol := style.BorderColor
	if open || drp.hovered {
		borderCol = focusRingIndigo
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

	label := drp.formatFieldLabel()
	drawStyle := style
	if label == drp.Placeholder {
		drawStyle.TextColor = rl.NewColor(140, 142, 155, 255)
	}
	const fieldTextPad = float32(12)
	const fieldArrowGutter = float32(32)
	textMaxW := bounds.Width - fieldTextPad - fieldArrowGutter
	if textMaxW < 8 {
		textMaxW = 8
	}
	drawTextS(truncateTextS(label, textMaxW, drawStyle), int32(bounds.X+fieldTextPad), TextPosY(bounds, style), drawStyle)

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

func (drp *DateRangePicker) IsInteractive() bool { return true }

// ── dateRangePickerManager ───────────────────────────────────────────────────

type dateRangePickerManager struct {
	open    bool
	closing bool
	target  *DateRangePicker

	origStart, origEnd time.Time
	draftStart, draftEnd time.Time
	awaitingEnd        bool

	fade       float32
	skipFrames int

	viewYear  int
	viewMonth time.Month
	focusDate time.Time
	hoverDate time.Time // mouse-over day (light grey); distinct from focus/selection

	popX, popY, popW, popH float32
	cellW                  float32
	prevBtn, nextBtn       rl.Rectangle
	gridOriginX, gridOriginY float32
	lastFieldX, lastFieldY, lastFieldW, lastFieldH float32
}

// DateRangePickerMgr is the package-level calendar overlay for range pickers.
var DateRangePickerMgr = &dateRangePickerManager{}

func (m *dateRangePickerManager) isTarget(drp *DateRangePicker) bool {
	return m.open && m.target == drp && !m.closing
}

func (m *dateRangePickerManager) IsOpen() bool { return m.open }

// PopupBounds returns the screen-space range calendar popup rectangle while open.
func (m *dateRangePickerManager) PopupBounds() rl.Rectangle {
	if !m.open {
		return rl.Rectangle{}
	}
	return rl.NewRectangle(m.popX, m.popY, m.popW, m.popH)
}

func (m *dateRangePickerManager) IsAnimating() bool {
	if !m.open {
		return false
	}
	return m.closing || m.fade < 1 || m.skipFrames > 0
}

func (m *dateRangePickerManager) snapClose() {
	m.open = false
	m.closing = false
	m.target = nil
	m.fade = 0
	m.skipFrames = 0
}

func (m *dateRangePickerManager) setVisiblePage(drp *DateRangePicker) {
	drp.visiblePage = monthStart(m.viewYear, m.viewMonth)
	drp.MarkDrawDirty()
}

func (m *dateRangePickerManager) Open(drp *DateRangePicker, anchorX, anchorY float32) {
	if drp == nil {
		return
	}
	if m.open && m.target != nil && m.target != drp {
		m.cancelAndClose()
		m.snapClose()
	}
	if DatePickerMgr.IsOpen() {
		DatePickerMgr.snapClose()
	}

	m.target = drp
	m.origStart = dateTruncLocal(drp.Start.Get())
	m.origEnd = dateTruncLocal(drp.End.Get())
	m.draftStart = m.origStart
	m.draftEnd = m.origEnd
	m.awaitingEnd = false

	ref := m.draftStart
	if ref.IsZero() {
		ref = m.draftEnd
	}
	if ref.IsZero() {
		ref = dateTruncLocal(time.Now())
	}
	m.focusDate = ref
	y, mo, _ := m.focusDate.Date()
	m.viewYear, m.viewMonth = y, mo
	m.setVisiblePage(drp)

	field := drp.Bounds()
	m.layoutPopupAnchored(field, anchorX, anchorY)
	m.open = true
	m.closing = false
	m.fade = 0
	m.skipFrames = dpSkipFrames
	m.layoutHitRects()
}

func (m *dateRangePickerManager) layoutPopupAnchored(field rl.Rectangle, anchorX, anchorY float32) {
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
	px, py := anchorX, anchorY
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
	m.popX, m.popY = px, py
	m.lastFieldX, m.lastFieldY = field.X, field.Y
	m.lastFieldW, m.lastFieldH = field.Width, field.Height
}

func (m *dateRangePickerManager) syncPopupToField() {
	if m.target == nil || !m.open || m.closing {
		return
	}
	field := m.target.Bounds()
	if field.X == m.lastFieldX && field.Y == m.lastFieldY &&
		field.Width == m.lastFieldW && field.Height == m.lastFieldH {
		return
	}
	m.layoutPopupAnchored(field, field.X, field.Y+field.Height+4)
	m.layoutHitRects()
}

func (m *dateRangePickerManager) layoutHitRects() {
	innerLeft := m.popX + dpPad
	headerY := m.popY + dpPad
	m.prevBtn = rl.NewRectangle(innerLeft, headerY+4, dpNavBtnW, dpHeaderH-8)
	m.nextBtn = rl.NewRectangle(m.popX+m.popW-dpPad-dpNavBtnW, headerY+4, dpNavBtnW, dpHeaderH-8)
	m.gridOriginX = innerLeft
	m.gridOriginY = headerY + dpHeaderH + dpDowH
}

func (m *dateRangePickerManager) cellRect(col, row int) rl.Rectangle {
	cw := m.cellW
	if cw <= 0 {
		cw = dpCellW
	}
	x := m.gridOriginX + float32(col)*cw
	y := m.gridOriginY + float32(row)*dpCellH
	return rl.NewRectangle(x, y, cw, dpCellH)
}

func (m *dateRangePickerManager) cancelAndClose() {
	if m.target != nil {
		m.target.Start.Set(m.origStart)
		m.target.End.Set(m.origEnd)
		m.target.MarkDrawDirty()
	}
	m.closing = true
}

func (m *dateRangePickerManager) commitAndClose() {
	if m.target != nil {
		s, e := normalizeDateRange(m.draftStart, m.draftEnd)
		if !s.IsZero() && e.IsZero() {
			e = s
		}
		m.target.Start.Set(s)
		m.target.End.Set(e)
		m.target.MarkDrawDirty()
	}
	m.closing = true
}

func (m *dateRangePickerManager) pickDay(d time.Time) {
	d = dateTruncLocal(d)
	if !m.awaitingEnd {
		m.draftStart = d
		m.draftEnd = time.Time{}
		m.awaitingEnd = true
		m.focusDate = d
	} else {
		m.draftEnd = d
		m.commitAndClose()
		return
	}
	if m.target != nil {
		m.target.MarkDrawDirty()
	}
}

func (m *dateRangePickerManager) shiftMonth(delta int) {
	_, _, pref := m.focusDate.Date()
	first := monthStart(m.viewYear, m.viewMonth).AddDate(0, delta, 0)
	m.viewYear, m.viewMonth, _ = first.Date()
	m.focusDate = clampDayInMonth(m.viewYear, m.viewMonth, pref)
	if m.target != nil {
		m.setVisiblePage(m.target)
	}
}

func (m *dateRangePickerManager) moveFocusByDays(delta int) {
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

func (m *dateRangePickerManager) Update(dt float32) {
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

	if rl.IsKeyPressed(rl.KeyEscape) {
		m.cancelAndClose()
		return
	}
	if rl.IsKeyPressed(rl.KeyEnter) || rl.IsKeyPressed(rl.KeyKpEnter) {
		if m.awaitingEnd && m.draftEnd.IsZero() {
			m.draftEnd = m.draftStart
		}
		m.commitAndClose()
		return
	}
	if rl.IsKeyPressed(rl.KeyPageUp) {
		m.shiftMonth(-1)
	}
	if rl.IsKeyPressed(rl.KeyPageDown) {
		m.shiftMonth(1)
	}
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
				m.pickDay(cell.Date)
				return
			}
		}
		if !rl.CheckCollisionPointRec(mouse, popRec) && !rl.CheckCollisionPointRec(mouse, fieldRec) {
			m.cancelAndClose()
		}
	}
}

func (m *dateRangePickerManager) Draw() {
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
	shadowRec := rl.NewRectangle(popRec.X+2, popRec.Y+4, popRec.Width, popRec.Height)
	rl.DrawRectangleRounded(shadowRec, 0.06, 12, mul(rl.NewColor(0, 0, 20, 55)))
	rl.DrawRectangleRounded(popRec, 0.06, 12, mul(rl.NewColor(252, 252, 255, 255)))
	drawPopupRoundedBorder(popRec, 10, 1.5, mul(rl.NewColor(198, 200, 214, 255)))

	title := fmt.Sprintf("%s %d", m.viewMonth.String(), m.viewYear)
	tw := MeasureText(title, 17)
	DrawText(title, m.popX+(m.popW-tw)/2, m.popY+dpPad+8, 17, mul(rl.NewColor(40, 42, 58, 255)))

	drawMiniNav(m.prevBtn, true, mul)
	drawMiniNav(m.nextBtn, false, mul)

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
	rangeLo, rangeHi := normalizeDateRange(m.draftStart, m.draftEnd)
	hasBoth := !rangeLo.IsZero() && !rangeHi.IsZero()
	draftStart := dateTruncLocal(m.draftStart)
	hasStartOnly := !draftStart.IsZero() && dateTruncLocal(m.draftEnd).IsZero()

	grid := buildMonthGrid(m.viewYear, m.viewMonth)
	for i, cell := range grid {
		if !cell.InMonth {
			continue
		}
		row, col := i/calDaysPerWeek, i%calDaysPerWeek
		cr := m.cellRect(col, row)
		isToday := cell.Date.Equal(today)
		isFocus := cell.Date.Equal(m.focusDate)
		isHover := !m.hoverDate.IsZero() && cell.Date.Equal(m.hoverDate)

		isEndpoint := false
		if hasBoth {
			isEndpoint = cell.Date.Equal(rangeLo) || cell.Date.Equal(rangeHi)
		} else if hasStartOnly {
			isEndpoint = cell.Date.Equal(draftStart)
		}
		inBetween := hasBoth && !cell.Date.Before(rangeLo) && !cell.Date.After(rangeHi) && !isEndpoint

		paint := CalendarDayPaint{}
		if inBetween || isEndpoint {
			// Endpoints share the range fill so the selection reads as one band;
			// indigo ring still marks first/last day.
			paint.Fill = rl.NewColor(237, 235, 255, 255)
		} else if isToday {
			paint.Fill = rl.NewColor(224, 231, 255, 255)
		}
		if isHover && !isEndpoint {
			// Light grey hover disc — distinct from selected (indigo ring) and today.
			paint.Fill = rl.NewColor(226, 228, 234, 255)
		}
		if isEndpoint {
			paint.Ring = rl.NewColor(79, 70, 229, 255)
			paint.RingWidth = 2
		} else if isFocus && !isHover {
			paint.Ring = rl.NewColor(99, 90, 249, 255)
			paint.RingWidth = 2
		}
		paintCalendarDay(cr, paint, mul)

		ds := fmt.Sprintf("%d", cell.Date.Day())
		dtw := MeasureText(ds, dpDayFontSize)
		tc := mul(rl.NewColor(38, 40, 52, 255))
		if isEndpoint || (isFocus && !isHover) {
			tc = mul(rl.NewColor(55, 48, 163, 255))
		}
		DrawText(ds, cr.X+(cr.Width-dtw)/2, cr.Y+(dpCellH-float32(dpDayFontSize))/2, dpDayFontSize, tc)
	}
}
