package ui

import (
	"testing"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func TestPreferGeometryCoverageAAAtSSAA1(t *testing.T) {
	prev := BaseSupersamplingScale
	t.Cleanup(func() { BaseSupersamplingScale = prev })
	BaseSupersamplingScale = 1
	if !PreferGeometryCoverageAA() {
		t.Fatal("SSAA 1× must prefer coverage AA")
	}
	BaseSupersamplingScale = 1.5
	if PreferGeometryCoverageAA() {
		t.Fatal("SSAA 1.5× must use tessellated path (Path A softens)")
	}
}

func TestCornerRoundnessFromRadius(t *testing.T) {
	rr := cornerRoundnessFromRadius(rl.NewRectangle(0, 0, 100, 40), 20)
	if rr < 0.9 || rr > 1.01 {
		t.Fatalf("pill roundness=%v want ~1", rr)
	}
}
