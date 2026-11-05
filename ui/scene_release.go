package ui

import (
	"runtime"
)

// ReleaseSceneResources frees GPU / heap retained by the outgoing Studio scene.
// Call after scene.Destroy() and before building the next scene.
//
// Does not unload process-wide S1 chrome strikes or Remix icon atlases — those
// stay warm across Directory tabs. Working Set may not fully return to the OS
// (Go heap + DWM), but Private bytes and GPU textures should drop.
func ReleaseSceneResources(doc *Document) {
	if doc != nil {
		doc.UnloadCache()
	}
	sdfMeasureCacheClear()
	shapedMeasureCacheClear()
	unloadShapedGlyphCache()
	initShapedGlyphCache()
	runtime.GC()
}
