// Package ui (continued)
// See node.go for the full package documentation.
package ui

import rl "github.com/gen2brain/raylib-go/raylib"

const (
	timelinePadX       float32 = 8
	timelineTrackW     float32 = 28
	timelineDotR       float32 = 6
	timelineLineW      float32 = 2
	timelineEventMinH  float32 = 56
	timelineTitleGap   float32 = 4
	timelineTimeW      float32 = 72
)

// TimelineEvent is one entry in a vertical timeline.
type TimelineEvent struct {
	Title    string
	Subtitle string
	Time     string // display time label (e.g. "10:42 AM")
}

// Timeline renders a vertical activity/history feed with a track, dots, and text.
//
// Pass h=0 for AutoHeight. Width should be set by the flex parent (w=0 in columns).
//
// # LLM Prompt Template
//
//	tl := ui.NewTimeline("activity", []ui.TimelineEvent{
//	    {Title: "Build succeeded", Subtitle: "All tests passed", Time: "09:15"},
//	}, 0, 0, 0, 0)
//	vp.AddChild(tl)
//
// Demo scenes: **Timeline Demo**.
type Timeline struct {
	Element
	Events []TimelineEvent
}

// NewTimeline creates a timeline. h=0 enables AutoHeight.
func NewTimeline(id string, events []TimelineEvent, x, y, w, h float32) *Timeline {
	tl := &Timeline{
		Element: NewElement(id, x, y, w, h),
		Events:  events,
	}
	tl.styleName = "timeline"
	if h == 0 {
		tl.AutoHeight = true
	}
	return tl
}

// IsInteractive implements Node.
func (t *Timeline) IsInteractive() bool { return false }

// Update implements Node.Update (no-op).
func (t *Timeline) Update(_ float32) {}

func (t *Timeline) eventHeight(ev TimelineEvent, contentW float32) float32 {
	style := t.GetStyle()
	subStyle := GetThemeStyle("timeline-subtitle")
	titleFS := EffectiveFontSize(style)
	subFS := EffectiveFontSize(subStyle)
	h := titleFS + timelineTitleGap + 8
	if ev.Subtitle != "" {
		h += subFS + 4
	}
	if h < timelineEventMinH {
		h = timelineEventMinH
	}
	return h
}

func (t *Timeline) contentWidth(totalW float32) float32 {
	w := totalW - timelineTrackW - timelinePadX*2
	if w < 48 {
		return 48
	}
	return w
}

// Layout sets AutoHeight from event count and wrapped subtitles.
func (t *Timeline) Layout() {
	defer func() { t.layoutDirty = false }()
	if t.IsHidden() {
		return
	}
	b := t.Bounds()
	contentW := t.contentWidth(b.Width)
	var total float32
	for _, ev := range t.Events {
		total += t.eventHeight(ev, contentW)
	}
	if total < timelineEventMinH && len(t.Events) > 0 {
		total = timelineEventMinH
	}
	want := total + timelinePadX
	if t.IsAutoHeight() && (b.Height < want-0.5 || b.Height > want+0.5) {
		b.Height = want
		t.setBoundsNoMark(b)
	}
}

// Draw implements Node.Draw.
func (t *Timeline) Draw() { t.drawInternal() }

func (t *Timeline) drawInternal() {
	if t.IsHidden() || len(t.Events) == 0 {
		return
	}
	b := t.Bounds()
	style := t.GetStyle()
	subStyle := GetThemeStyle("timeline-subtitle")
	timeStyle := GetThemeStyle("timeline-time")
	accent := rl.NewColor(79, 70, 229, 255)
	lineColor := rl.NewColor(210, 214, 228, 255)
	dotBorder := style.BackgroundColor
	if dotBorder.A == 0 {
		dotBorder = rl.White
	}

	trackX := b.X + timelinePadX + timelineTrackW/2
	contentX := b.X + timelinePadX + timelineTrackW
	contentW := t.contentWidth(b.Width)
	y := b.Y + timelinePadX / 2

	for i, ev := range t.Events {
		eh := t.eventHeight(ev, contentW)
		cy := y + eh/2
		if i < len(t.Events)-1 {
			rl.DrawRectangle(int32(trackX-timelineLineW/2), int32(cy+timelineDotR),
				int32(timelineLineW), int32(eh-timelineDotR*2+8), lineColor)
		}
		rl.DrawCircleV(rl.NewVector2(trackX, cy), timelineDotR+1.5, dotBorder)
		rl.DrawCircleV(rl.NewVector2(trackX, cy), timelineDotR, accent)

		titleStyle := style
		titleStyle.Bold = true
		drawTextS(ev.Title, int32(contentX), int32(y+6), titleStyle)

		if ev.Time != "" {
			tw := measureTextS(ev.Time, timeStyle)
			drawTextS(ev.Time, int32(b.X+b.Width-timelinePadX-float32(tw)), int32(y+6), timeStyle)
		}

		textY := y + 6 + float32(EffectiveFontSize(titleStyle)) + timelineTitleGap
		if ev.Subtitle != "" {
			drawTextS(ev.Subtitle, int32(contentX), int32(textY), subStyle)
		}
		y += eh
	}
}
