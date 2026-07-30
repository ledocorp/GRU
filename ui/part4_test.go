package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestGaugeValueSignal(t *testing.T) {
	v := NewSignal(float32(0.72))
	g := NewGauge("g", v, "CPU", 0, 0, 80, 80)
	if g.Value.Get() != 0.72 {
		t.Fatalf("value = %v", g.Value.Get())
	}
}

func TestChartEmptySeries(t *testing.T) {
	c := NewChart("c", ChartLine, nil, 0, 0, 200, 100)
	if len(c.Series) != 0 {
		t.Fatal("expected nil series")
	}
}

func TestChartSetSeries(t *testing.T) {
	c := NewChart("c", ChartBar, []float32{1, 2}, 0, 0, 200, 100)
	c.SetSeries([]float32{10, 20, 30})
	if len(c.Series) != 3 {
		t.Fatalf("series len = %d", len(c.Series))
	}
}

func TestFileDropZoneFilter(t *testing.T) {
	z := NewFileDropZone("z", []string{".png"}, nil, 0, 0, 100, 50)
	got := z.filter([]string{"a.png", "b.txt"})
	if len(got) != 1 || got[0] != "a.png" {
		t.Fatalf("filter = %v", got)
	}
}

func TestPropertyTableLayout(t *testing.T) {
	theme := NewSignal("Dark")
	pt := NewPropertyTable("pt", []PropertyRow{
		{Key: "Theme", Value: theme},
		{Key: "Version", Value: NewSignal("1.0")},
	}, 0, 0, 300, 120)
	pt.SetBounds(rl.NewRectangle(0, 0, 300, 120))
	pt.Layout()
	if pt.form == nil {
		t.Fatal("form is nil")
	}
	if pt.form.FieldCount() != 2 {
		t.Fatalf("form fields = %d", pt.form.FieldCount())
	}
}

func TestNotificationHistoryAppend(t *testing.T) {
	ClearNotificationHistory()
	appendNotificationHistory("test", ToastInfo)
	if len(notificationHistory) != 1 {
		t.Fatalf("history len = %d", len(notificationHistory))
	}
	ClearNotificationHistory()
}
