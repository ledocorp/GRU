package ui

// Density is the device scale where 160dpi (mdpi) = 1.0.
// Android: densityDpi/160 from AConfiguration (see RefreshDensity).
// Desktop: 1.0 — chrome stays authored in logical pixels; type uses rem.
var deviceDensity float32 = 1

// Phone chrome tokens (dp). Convert with [Dp]. Desktop call sites ignore these
// and keep fixed px constants.
const (
	// Tool strip == cell height so stacked icons fill the band (no empty chrome pad).
	PhoneToolStripHDp     = float32(48)
	PhoneRibbonCellWDp    = float32(48)
	PhoneRibbonCellHDp    = float32(48)
	PhoneRibbonCaptionDp  = float32(10) // cancel rem for captions (like BottomNav labels)
	PhoneRibbonPadHDp     = float32(4)  // outer left/right strip pad
	PhoneRibbonItemGapDp  = float32(2)  // gap between cells in a section
	PhoneRibbonSectionGapDp = float32(10)
	PhoneBottomNavHDp     = float32(56) // icon + label + tight pad
	PhoneBottomNavIconDp  = float32(24)
	PhoneBottomNavLabelDp = float32(11) // absolute label size via EffectiveFontSize cancel
	PhoneTouchTargetMinDp = float32(48)
)

// Density returns the current device density scale (160dpi → 1.0).
func Density() float32 {
	if deviceDensity < 0.5 {
		return 1
	}
	return deviceDensity
}

// Dp converts density-independent pixels to framebuffer pixels.
//
//	h := ui.Dp(56) // Material-sized app bar on any density
func Dp(v float32) float32 {
	return v * Density()
}

// RefreshDensity updates [Density] from the platform. Call after InitWindow /
// alongside [RefreshDisplayScale].
func RefreshDensity() {
	if AndroidGLES() {
		deviceDensity = readAndroidDensity()
		if deviceDensity < 1 {
			deviceDensity = 1
		}
		if deviceDensity > 4 {
			deviceDensity = 4
		}
		return
	}
	deviceDensity = 1
}
