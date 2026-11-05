// Package ui (continued)
// See node.go for the full package documentation.
package ui

import (
	"time"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// Calendar layout uses a fixed 6×7 grid (42 cells). Weeks start on Sunday to
// match Go's [time.Weekday] numbering (Sunday = 0).
const (
	calWeeks         = 6
	calDaysPerWeek   = 7
	calGridCellCount = calWeeks * calDaysPerWeek // 42
	calDayDiscSegs   = int32(48)
)

// CalendarPopupOpen reports whether a date or date-range calendar overlay is
// open. While true, scene widgets should not react to hover/wheel underneath.
func CalendarPopupOpen() bool {
	return DatePickerMgr.IsOpen() || DateRangePickerMgr.IsOpen()
}

// CalendarCell describes one slot in a month grid, including leading and
// trailing days from adjacent months that pad the first/last row.
type CalendarCell struct {
	Date    time.Time // local date at midnight
	InMonth bool      // true when Date falls in the requested year/month
}

// dateTruncLocal returns t normalized to the local calendar date at midnight.
func dateTruncLocal(t time.Time) time.Time {
	if t.IsZero() {
		return t
	}
	y, m, d := t.In(time.Local).Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

// monthStart returns the first moment of the given calendar month in local time.
func monthStart(year int, month time.Month) time.Time {
	return time.Date(year, month, 1, 0, 0, 0, 0, time.Local)
}

// daysInMonth returns the number of days in year/month.
func daysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.Local).Day()
}

// buildMonthGrid fills a 6×7 grid for the given calendar month. Index 0 is the
// top-left cell (Sunday column). Cells before the first of the month show the
// tail of the previous month; cells after the last day show the start of the
// next month.
func buildMonthGrid(year int, month time.Month) [calGridCellCount]CalendarCell {
	first := monthStart(year, month)
	lead := int(first.Weekday()) // Sunday = 0

	var out [calGridCellCount]CalendarCell
	for i := 0; i < calGridCellCount; i++ {
		off := i - lead
		d := first.AddDate(0, 0, off)
		y2, m2, _ := d.Date()
		in := m2 == month && y2 == year
		out[i] = CalendarCell{Date: dateTruncLocal(d), InMonth: in}
	}
	return out
}

// clampDayInMonth picks preferredDay in year/month, clamped to the last valid
// day (e.g. 31 → 28 when moving from January to February).
func clampDayInMonth(year int, month time.Month, preferredDay int) time.Time {
	dim := daysInMonth(year, month)
	d := preferredDay
	if d < 1 {
		d = 1
	}
	if d > dim {
		d = dim
	}
	return time.Date(year, month, d, 0, 0, 0, 0, time.Local)
}

// dayCellDisc returns the center and outer radius for a day indicator disc that
// fits inside the grid cell with room for a 2px ring stroke.
// Center and radius are whole pixels so the ring SDF and day number share one origin.
func dayCellDisc(cell rl.Rectangle) (rl.Vector2, float32) {
	side := cell.Width
	if cell.Height < side {
		side = cell.Height
	}
	r := float32(int32(side*0.44 + 0.5))
	if r < 10 {
		r = 10
	}
	cx := float32(int32(cell.X + cell.Width*0.5 + 0.5))
	cy := float32(int32(cell.Y + cell.Height*0.5 + 0.5))
	return rl.NewVector2(cx, cy), r
}

// pointInDayCell is true when pt lies inside the day disc for a grid cell.
func pointInDayCell(pt rl.Vector2, cell rl.Rectangle) bool {
	c, r := dayCellDisc(cell)
	dx, dy := pt.X-c.X, pt.Y-c.Y
	return dx*dx+dy*dy <= r*r
}

// CalendarDayPaint describes fill and/or ring for one calendar day disc.
type CalendarDayPaint struct {
	Fill      rl.Color
	Ring      rl.Color
	RingWidth float32
}

// paintCalendarDay draws a circular day indicator inside a grid cell.
func paintCalendarDay(cell rl.Rectangle, p CalendarDayPaint, mul func(rl.Color) rl.Color) {
	if p.Fill.A == 0 && p.Ring.A == 0 {
		return
	}
	center, radius := dayCellDisc(cell)
	if p.Fill.A > 0 {
		DrawCoverageCircle(center.X, center.Y, radius, mul(p.Fill))
	}
	if p.Ring.A > 0 && p.RingWidth > 0 {
		// Slightly outside the fill; width ≥2 reads smoother under coverage AA at 1×.
		rw := p.RingWidth
		if rw < 2 {
			rw = 2
		}
		outer := radius + 1
		inner := outer - rw
		if inner < 0.5 {
			inner = 0.5
		}
		DrawCoverageRing(center.X, center.Y, inner, outer, 0, 360, mul(p.Ring))
	}
}

// calendarDayTextStyle matches form-label captions (S1 sticky strikes via drawTextS).
// Calendar used to call DrawText at token 13 — off-ladder soft scale looked muddy
// next to same-size form labels on the right.
func calendarDayTextStyle() Style {
	s := GetThemeStyle("form-label")
	s.NoCaptionBold = true
	return s
}

func calendarDowTextStyle() Style {
	s := GetThemeStyle("form-label")
	s.NoCaptionBold = true
	s.TextColor = rl.NewColor(120, 122, 140, 255)
	return s
}

func calendarTitleTextStyle() Style {
	s := GetThemeStyle("form-field-caption")
	s.Bold = true
	s.FontSize = 17
	if s.MinFontSize < 16 {
		s.MinFontSize = 16
	}
	s.TextColor = rl.NewColor(40, 42, 58, 255)
	return s
}

// drawCalendarDayNumber centers a day-of-month label on the day disc using the
// same drawTextS / EffectiveFontSize path as form captions (ClearType strikes).
func drawCalendarDayNumber(label string, center rl.Vector2, col rl.Color) {
	st := calendarDayTextStyle()
	st.TextColor = col
	tw := float32(measureTextS(label, st))
	cap := TextCapHeight(st)
	ascent := TextInkHeight(st)
	if ascent < 1 {
		ascent = EffectiveFontSize(st)
	}
	if cap < 1 {
		cap = ascent * 0.75
	}
	x := int32(center.X - tw/2 + 0.5)
	// Zero-height bounds → mid at disc center; cap-vs-ascent matches buttons.
	y := controlTextPosYMath(center.Y, 0, cap, ascent)
	drawTextS(label, x, y, st)
}

// calendarTextColor keeps full-alpha ink when the popup is settled — partial
// alpha destroys ClearType/LCD sharpness during/after fade.
func calendarTextColor(base rl.Color, fade float32) rl.Color {
	if fade >= 0.999 {
		return base
	}
	base.A = uint8(float32(base.A) * fade)
	return base
}
