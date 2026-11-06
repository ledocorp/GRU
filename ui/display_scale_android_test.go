package ui

import "testing"

func TestAndroidDisplayScaleFromWidth(t *testing.T) {
	cases := []struct {
		w    int
		want float32
	}{
		{1080, 2.5},
		{1440, 2.5},
		{800, 2.0},
		{540, 1.5},
		{320, 1.25},
	}
	for _, c := range cases {
		if got := androidDisplayScaleFromWidth(c.w); got != c.want {
			t.Fatalf("width %d: got %.2f want %.2f", c.w, got, c.want)
		}
	}
}
