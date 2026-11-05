package ui

import "testing"

func TestWindowChromeFillDarkVsLight(t *testing.T) {
	prev := currentAppearanceMode
	t.Cleanup(func() { SetAppearance(prev) })
	SetAppearance(AppearanceLight)
	light := WindowChromeFill()
	if light.A != 255 || int(light.R)+int(light.G)+int(light.B) < 600 {
		t.Fatalf("light fill should be pale opaque, got %+v", light)
	}
	SetAppearance(AppearanceDark)
	dark := WindowChromeFill()
	if dark.A != 255 || int(dark.R)+int(dark.G)+int(dark.B) > 120 {
		t.Fatalf("dark fill should be near-black opaque, got %+v", dark)
	}
}
