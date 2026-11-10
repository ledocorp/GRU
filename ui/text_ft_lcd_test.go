package ui

import (
	"testing"

	"github.com/golang/freetype/truetype"
	rl "github.com/gen2brain/raylib-go/raylib"
	"golang.org/x/image/font"
)

func TestPreferFTLCDChromeOptIn(t *testing.T) {
	t.Cleanup(func() {
		t.Setenv("GRU_TEXT_LCD", "")
		t.Setenv("GORY_TEXT_LCD", "")
		t.Setenv("GRU_TEXT_STRIKES", "")
		t.Setenv("GORY_TEXT_STRIKES", "")
		t.Setenv("GRU_SSAA", "")
	})
	t.Setenv("GRU_TEXT_LCD", "")
	t.Setenv("GRU_TEXT_STRIKES", "")
	BaseSupersamplingScale = 1
	if dwriteLCDAvailable() {
		if !PreferFTLCDChrome() {
			t.Fatal("Windows S1 (DWrite) must auto-enable LCD")
		}
	} else if PreferFTLCDChrome() {
		t.Fatal("without DWrite, S1 must default greyscale sticky (no lite LCD)")
	}
	t.Setenv("GRU_TEXT_STRIKES", "0")
	if PreferFTLCDChrome() {
		t.Fatal("GRU_TEXT_STRIKES=0 (Path A) must not auto-enable LCD")
	}
	t.Setenv("GRU_TEXT_STRIKES", "")
	t.Setenv("GRU_TEXT_LCD", "1")
	if !PreferFTLCDChrome() {
		t.Fatal("env=1 enables")
	}
	t.Setenv("GRU_TEXT_LCD", "0")
	t.Setenv("GRU_TEXT_STRIKES", "1")
	t.Setenv("GRU_SSAA", "1")
	if PreferFTLCDChrome() {
		t.Fatal("GRU_TEXT_LCD=0 must opt out even for S1")
	}
}

func TestPackFTLCDAtlasHasRGB(t *testing.T) {
	path := firstExisting(fontCandidates...)
	if path == "" {
		t.Skip("no UI font")
	}
	data, err := ReadAssetFile(path)
	if err != nil {
		t.Fatal(err)
	}
	ttf, err := truetype.Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	face := truetype.NewFace(ttf, &truetype.Options{
		Size:              32,
		DPI:               72,
		Hinting:           font.HintingFull,
		GlyphCacheEntries: 8,
	})
	defer face.Close()
	atlas, glyphs, ink, ok := packFTLCDAtlas(face, 32)
	if !ok {
		t.Fatalf("pack failed ink=%d", ink)
	}
	g, ok := glyphs['H']
	if !ok || !g.hasInk {
		t.Fatalf("glyph H missing: %+v", g)
	}
	found := false
	for y := int(g.src.Y); y < int(g.src.Y+g.src.Height); y++ {
		for x := int(g.src.X); x < int(g.src.X+g.src.Width); x++ {
			r, gch, b, _ := atlas.At(x, y).RGBA()
			if r>>8 > 40 || gch>>8 > 40 || b>>8 > 40 {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatal("expected LCD RGB coverage in H")
	}
}

func TestLCDBgHeuristic(t *testing.T) {
	prevL, prevD := ftLCDLightBackdrop, ftLCDDarkBackdrop
	t.Cleanup(func() {
		ftLCDLightBackdrop = prevL
		ftLCDDarkBackdrop = prevD
	})
	SetFTLCDLightBackdrop(rl.NewColor(228, 229, 235, 255))
	SetFTLCDDarkBackdrop(rl.NewColor(30, 30, 30, 255))
	bgDarkText := lcdBgForText(rl.NewColor(30, 30, 40, 255))
	if bgDarkText.X < 0.85 || bgDarkText.X > 0.95 {
		t.Fatalf("dark text should mix against page grey (~0.89), got %+v", bgDarkText)
	}
	bgMuted := lcdBgForText(rl.NewColor(72, 78, 100, 255))
	if bgMuted.X < 0.85 {
		t.Fatalf("muted grey text should use light backdrop, got %+v", bgMuted)
	}
	bgLightText := lcdBgForText(rl.NewColor(240, 240, 250, 255))
	if bgLightText.X > 0.2 {
		t.Fatalf("light text should mix against dark backdrop (~0.12), got %+v", bgLightText)
	}
}

func TestLCDSurfaceBackdropOverridesPageGrey(t *testing.T) {
	prevStack := lcdSurfaceBackdropStack
	prevL := ftLCDLightBackdrop
	lcdSurfaceBackdropStack = nil
	t.Cleanup(func() {
		lcdSurfaceBackdropStack = prevStack
		ftLCDLightBackdrop = prevL
	})
	SetFTLCDLightBackdrop(rl.NewColor(228, 229, 235, 255))
	if !PushLCDSurfaceBackdrop(rl.NewColor(255, 255, 255, 255)) {
		t.Fatal("opaque white tile fill should push")
	}
	defer PopLCDSurfaceBackdrop(true)
	bg := lcdBgForText(rl.NewColor(28, 32, 48, 255))
	if bg.X < 0.98 || bg.Y < 0.98 {
		t.Fatalf("list-tile dark text should mix against white, got %+v", bg)
	}
	if lcdBgForText(rl.NewColor(240, 240, 250, 255)).X < 0.98 {
		t.Fatal("light ink on a pushed fill must mix against that fill, not generic dark")
	}
}

func TestLCDSurfaceBackdropLightInkOnButtonFill(t *testing.T) {
	prevStack := lcdSurfaceBackdropStack
	lcdSurfaceBackdropStack = nil
	t.Cleanup(func() { lcdSurfaceBackdropStack = prevStack })
	purple := rl.NewColor(79, 70, 229, 255)
	if !PushLCDSurfaceBackdrop(purple) {
		t.Fatal("opaque button fill should push")
	}
	defer PopLCDSurfaceBackdrop(true)
	bg := lcdBgForText(rl.NewColor(255, 255, 255, 255))
	if bg.X < 0.29 || bg.X > 0.33 || bg.Z < 0.88 {
		t.Fatalf("white button label must mix against primary fill, got %+v", bg)
	}
}

func TestLCDWidgetFillMixOutput(t *testing.T) {
	prevStack := lcdSurfaceBackdropStack
	lcdSurfaceBackdropStack = nil
	t.Cleanup(func() { lcdSurfaceBackdropStack = prevStack })
	if lcdWidgetFillMixOutput() {
		t.Fatal("empty stack must not force widget-fill opaque mix")
	}
	if !PushLCDSurfaceBackdrop(rl.NewColor(79, 70, 229, 255)) {
		t.Fatal("button fill should push")
	}
	defer PopLCDSurfaceBackdrop(true)
	if !lcdWidgetFillMixOutput() {
		t.Fatal("pushed opaque fill must use widget-fill opaque mix output")
	}
}

func TestLCDClearTypeSaturatedDest(t *testing.T) {
	grey := lcdClearTypeForDest(1, rl.NewVector3(0.12, 0.12, 0.14))
	if grey < 0.99 {
		t.Fatalf("neutral dest should keep ClearType, got %v", grey)
	}
	purple := lcdClearTypeForDest(1, colorToLCDBackdrop(rl.NewColor(79, 70, 229, 255)))
	if purple > 0.4 {
		t.Fatalf("primary fill should collapse RGB ClearType, got %v", purple)
	}
}

func TestLCDLiftHostileMixDestMatchesHover(t *testing.T) {
	primary := colorToLCDBackdrop(rl.NewColor(79, 70, 229, 255))
	lifted := lcdLiftHostileMixDest(primary)
	want := rl.NewVector3(
		primary.X+(1-primary.X)*0.25,
		primary.Y+(1-primary.Y)*0.25,
		primary.Z+(1-primary.Z)*0.25,
	)
	if abs32(lifted.X-want.X) > 0.001 || abs32(lifted.Y-want.Y) > 0.001 || abs32(lifted.Z-want.Z) > 0.001 {
		t.Fatalf("hostile rest mix should match hover ColorBrightness(+0.25): got %+v want %+v", lifted, want)
	}
	bright := lcdLiftHostileMixDest(rl.NewVector3(0.75, 0.70, 0.90))
	if bright.X != 0.75 || bright.Y != 0.70 {
		t.Fatalf("already-bright dest must skip lift, got %+v", bright)
	}
}

func TestLCDClearTypeChromeLightIsGreyscale(t *testing.T) {
	green := colorToLCDBackdrop(rl.NewColor(22, 163, 74, 255))
	page := lcdClearTypeForDestOpt(1, green, false)
	if page < 0.1 {
		t.Fatalf("page soft-CT on green should stay >0, got %v", page)
	}
	chrome := lcdClearTypeForDestOpt(1, green, true)
	if chrome != 0 {
		t.Fatalf("C25 chrome light-on-sat must be greyscale, got %v", chrome)
	}
}

func abs32(v float32) float32 {
	if v < 0 {
		return -v
	}
	return v
}
