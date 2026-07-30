package ui

import "testing"

func TestImageResetTextureBumpsLoadGen(t *testing.T) {
	img := NewImage("t", "path.png", 0, 0, 10, 10)
	before := img.loadGen
	img.ResetTexture()
	if img.loadGen != before+1 {
		t.Fatalf("loadGen: got %d want %d", img.loadGen, before+1)
	}
	if img.loaded {
		t.Fatal("expected unloaded after ResetTexture")
	}
}
