package ui

import (
	"image"
	"os"
	"testing"
)

func TestInterVFBoldHasHeavierInkThanRegular(t *testing.T) {
	path := "../assets/fonts/Inter/Inter-VariableFont_opsz,wght.ttf"
	data, err := os.ReadFile(path)
	if err != nil {
		t.Skip(err)
	}
	reg, gR, inkR, okR := packFTGreyAtlasOutlineVars(data, 16, 4, nil)
	if !okR || reg == nil {
		t.Fatal("regular pack failed")
	}
	bold, gB, inkB, okB := packFTGreyAtlasOutlineVars(data, 16, 4, interVariableBoldVariations(16))
	if !okB || bold == nil {
		t.Fatal("bold pack failed")
	}
	if inkB <= inkR {
		t.Fatalf("expected bold ink > regular: bold=%d regular=%d", inkB, inkR)
	}
	cover := func(img *image.RGBA, g ftGreyGlyph) int {
		if !g.hasInk {
			return 0
		}
		x0, y0 := int(g.src.X), int(g.src.Y)
		w, h := int(g.src.Width), int(g.src.Height)
		n := 0
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				_, _, _, a := img.At(x0+x, y0+y).RGBA()
				n += int(a >> 8)
			}
		}
		return n
	}
	heavier := 0
	for _, r := range []rune{'a', 'b', 'c', 'p', 's', 'w', 'e', 'o'} {
		cr := cover(reg, gR[r])
		cb := cover(bold, gB[r])
		t.Logf("%q cover regular=%d bold=%d Δ=%d", string(r), cr, cb, cb-cr)
		if cb > cr {
			heavier++
		}
	}
	if heavier < 5 {
		t.Fatalf("bold VF wght should darken most glyphs; heavier=%d/8 inkR=%d inkB=%d", heavier, inkR, inkB)
	}
}

func TestOutlineVarsForAuxPathInterVF(t *testing.T) {
	vars := outlineVarsForAuxPath("assets/fonts/Inter/Inter-VariableFont_opsz,wght.ttf", true, 16)
	if len(vars) < 1 {
		t.Fatal("expected wght vars for Inter VF bold")
	}
	if outlineVarsForAuxPath("assets/fonts/Inter/Inter-VariableFont_opsz,wght.ttf", false, 16) != nil {
		t.Fatal("regular should not set VF axes")
	}
	if outlineVarsForAuxPath("assets/fonts/poppins/Poppins-Bold.ttf", true, 16) != nil {
		t.Fatal("static bold should not set VF axes")
	}
}
