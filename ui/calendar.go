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
func dayCellDisc(cell rl.Rectangle) (rl.Vector2, float32) {
	side := cell.Width
	if cell.Height < side {
		side = cell.Height
	}
	r := side * 0.38
	return rl.NewVector2(cell.X+cell.Width*0.5, cell.Y+cell.Height*0.5), r
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
		rl.DrawCircleV(center, radius, mul(p.Fill))
	}
	if p.Ring.A > 0 && p.RingWidth > 0 {
		inner := radius - p.RingWidth
		if inner < 0.5 {
			inner = 0.5
		}
		rl.DrawRing(center, inner, radius, 0, 360, calDayDiscSegs, mul(p.Ring))
	}
}
