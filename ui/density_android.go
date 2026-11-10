//go:build android

package ui

/*
#include <stdint.h>
#include <android/configuration.h>
#include <android_native_app_glue.h>

extern struct android_app *GetAndroidApp(void);

// gruReadDensityDpi returns AConfiguration densityDpi, or 0 if unavailable.
static int32_t gruReadDensityDpi(void) {
	struct android_app *app = GetAndroidApp();
	if (app == NULL || app->config == NULL) {
		return 0;
	}
	return AConfiguration_getDensity(app->config);
}
*/
import "C"

import rl "github.com/gen2brain/raylib-go/raylib"

// readAndroidDensity maps densityDpi → scale (160 → 1.0). Falls back to width
// buckets when the configuration is not ready yet.
func readAndroidDensity() float32 {
	dpi := int32(C.gruReadDensityDpi())
	if dpi >= 120 {
		return float32(dpi) / 160
	}
	// Fallback: same buckets as atlas DisplayScale, expressed as density.
	w := rl.GetScreenWidth()
	switch {
	case w >= 1000:
		return 2.625 // ~420dpi phone / AVD
	case w >= 720:
		return 2.0
	case w >= 480:
		return 1.5
	default:
		return 1.25
	}
}
