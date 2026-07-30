package appicon

import (
	"bytes"
	"image/color"
	"image/png"
	"os"
	"runtime"
	"testing"
)

func TestEmbeddedWindowSizes(t *testing.T) {
	for name, data := range map[string][]byte{
		"window-16.png":      window16PNG,
		"window-32.png":      window32PNG,
		"window-48.png":      window48PNG,
		"window-dark-16.png": windowDark16PNG,
		"window-dark-32.png": windowDark32PNG,
		"window-dark-48.png": windowDark48PNG,
	} {
		if len(data) == 0 {
			t.Fatalf("%s embed empty — run: go run ./scripts/build/gen_app_icon.go", name)
		}
		img, err := png.Decode(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("decode %s: %v", name, err)
		}
		want := 16
		switch name {
		case "window-32.png", "window-dark-32.png":
			want = 32
		case "window-48.png", "window-dark-48.png":
			want = 48
		}
		if img.Bounds().Dx() != want || img.Bounds().Dy() != want {
			t.Fatalf("%s size = %v, want %dx%d", name, img.Bounds(), want, want)
		}
	}
}

func TestEmbeddedWindowIconValidPNG(t *testing.T) {
	if len(window32PNG) == 0 {
		t.Fatal("window-32.png embed is empty — run: go run ./scripts/build/gen_app_icon.go")
	}
	img, err := png.Decode(bytes.NewReader(window32PNG))
	if err != nil {
		t.Fatalf("decode window icon: %v", err)
	}
	b := img.Bounds()
	if b.Dx() != 32 || b.Dy() != 32 {
		t.Fatalf("window icon size = %dx%d, want 32x32", b.Dx(), b.Dy())
	}
}

func TestTrayPNGBytesRespectsDarkPreference(t *testing.T) {
	SetPreferDarkIcon(false)
	if !bytes.Equal(TrayPNGBytes(), notify16PNG) {
		t.Fatal("TrayPNGBytes light should match notify16PNG embed")
	}
	SetPreferDarkIcon(true)
	if !bytes.Equal(TrayPNGBytes(), notifyDark16PNG) {
		t.Fatal("TrayPNGBytes dark should match notifyDark16PNG embed")
	}
	SetPreferDarkIcon(false)
}

func TestEmbeddedNotify16PNG(t *testing.T) {
	if len(notify16PNG) == 0 {
		t.Fatal("notify-16.png embed empty — run: go run ./scripts/build/gen_app_icon.go")
	}
	img, err := png.Decode(bytes.NewReader(notify16PNG))
	if err != nil {
		t.Fatalf("decode notify-16: %v", err)
	}
	if img.Bounds().Dx() != 16 || img.Bounds().Dy() != 16 {
		t.Fatalf("notify-16 size = %v, want 16x16", img.Bounds())
	}
}

func TestEmbeddedNotifyDark16PNG(t *testing.T) {
	if len(notifyDark16PNG) == 0 {
		t.Fatal("notify-dark-16.png embed empty — run: go run ./scripts/build/gen_app_icon.go")
	}
	img, err := png.Decode(bytes.NewReader(notifyDark16PNG))
	if err != nil {
		t.Fatalf("decode notify-dark-16: %v", err)
	}
	if img.Bounds().Dx() != 16 || img.Bounds().Dy() != 16 {
		t.Fatalf("notify-dark-16 size = %v, want 16x16", img.Bounds())
	}
}

func TestEmbeddedNotifyIconValidPNG(t *testing.T) {
	if len(notify32PNG) == 0 {
		t.Fatal("notify-32.png embed is empty — run: go run ./scripts/build/gen_app_icon.go")
	}
	img, err := png.Decode(bytes.NewReader(notify32PNG))
	if err != nil {
		t.Fatalf("decode notify icon: %v", err)
	}
	if img.Bounds().Dx() != 32 || img.Bounds().Dy() != 32 {
		t.Fatalf("notify icon size = %v, want 32x32", img.Bounds())
	}
}

func TestNotifyIconFourQuadrantColors(t *testing.T) {
	assertQuadrantColors(t, notify32PNG, false)
	assertQuadrantColors(t, notifyDark32PNG, true)
}

func assertQuadrantColors(t *testing.T, pngData []byte, dark bool) {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	at := func(x, y int) color.RGBA {
		return color.RGBAModel.Convert(img.At(x, y)).(color.RGBA)
	}
	purpleTL := colorDist{minR: 0, maxR: 80, minG: 0, maxG: 50, minB: 0, maxB: 130}
	if dark {
		purpleTL = colorDist{minR: 120, maxR: 255, minG: 80, maxG: 220, minB: 200, maxB: 255}
	}
	quads := []struct {
		name string
		x, y int
		want colorDist
	}{
		{"purple TL", 8, 8, purpleTL},
		{"teal TR", 24, 8, colorDist{minR: 0, maxR: 80, minG: 160, maxG: 255, minB: 140, maxB: 255}},
		{"orange BL", 8, 24, colorDist{minR: 200, maxR: 255, minG: 100, maxG: 180, minB: 0, maxB: 80}},
		{"lilac BR", 24, 24, colorDist{minR: 160, maxR: 255, minG: 140, maxG: 220, minB: 200, maxB: 255}},
	}
	for _, q := range quads {
		c := at(q.x, q.y)
		if c.A < 128 {
			t.Fatalf("%s: transparent pixel at (%d,%d)", q.name, q.x, q.y)
		}
		if !q.want.match(c) {
			t.Fatalf("%s: got R=%d G=%d B=%d", q.name, c.R, c.G, c.B)
		}
	}
}

type colorDist struct {
	minR, maxR, minG, maxG, minB, maxB uint8
}

func (d colorDist) match(c color.RGBA) bool {
	in := func(v, lo, hi uint8) bool { return v >= lo && v <= hi }
	return in(c.R, d.minR, d.maxR) && in(c.G, d.minG, d.maxG) && in(c.B, d.minB, d.maxB)
}

func TestEmbeddedAppICOValid(t *testing.T) {
	if len(appICO) == 0 {
		t.Fatal("app.ico embed empty — run: go run ./scripts/build/gen_app_icon.go")
	}
	if len(appICO) < 6 || appICO[0] != 0 || appICO[1] != 0 || appICO[2] != 1 || appICO[3] != 0 {
		t.Fatal("app.ico missing ICO header")
	}
}

func TestEmbeddedNotifyICOValid(t *testing.T) {
	if len(notifyICO) == 0 {
		t.Fatal("notify.ico embed empty — run: go run ./scripts/build/gen_app_icon.go")
	}
	if len(notifyICO) < 6 || notifyICO[0] != 0 || notifyICO[1] != 0 || notifyICO[2] != 1 || notifyICO[3] != 0 {
		t.Fatal("notify.ico missing ICO header")
	}
}

func TestTrayIconBytesWindowsUsesICO(t *testing.T) {
	if runtime.GOOS == "windows" {
		b := TrayIconBytes()
		if len(b) < 6 || b[2] != 1 {
			t.Fatal("TrayIconBytes on Windows must be .ico data")
		}
		return
	}
	if len(TrayIconBytes()) == 0 {
		t.Fatal("TrayIconBytes empty")
	}
}

func TestPNGPathFallbackWhenPackagingPresent(t *testing.T) {
	// When run from repo root after gen, packaging PNG should resolve.
	if p := PNGPath(); p != "" {
		if _, err := os.Stat(p); err != nil {
			t.Fatalf("PNGPath %q not found: %v", p, err)
		}
	}
}
