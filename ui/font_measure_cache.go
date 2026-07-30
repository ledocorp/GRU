package ui

import "sync"

const sdfMeasureCacheMax = 4096

type sdfMeasureCacheKey struct {
	text     string
	fontSize uint16
	flags    uint8 // bit0 bold, bit1 italic, bit2 mono, bit3 preview
}

var (
	sdfMeasureCacheMu  sync.RWMutex
	sdfMeasureCache    map[sdfMeasureCacheKey]float32
	sdfMeasureCacheLen int
)

func sdfMeasureCacheFlags(bold, italic, mono, preview bool) uint8 {
	return shapedMeasureCacheFlags(bold, italic, mono, preview)
}

func sdfMeasureCacheLookup(key sdfMeasureCacheKey) (float32, bool) {
	sdfMeasureCacheMu.RLock()
	defer sdfMeasureCacheMu.RUnlock()
	w, ok := sdfMeasureCache[key]
	return w, ok
}

func sdfMeasureCacheStore(key sdfMeasureCacheKey, w float32) {
	sdfMeasureCacheMu.Lock()
	defer sdfMeasureCacheMu.Unlock()
	if sdfMeasureCache == nil {
		sdfMeasureCache = make(map[sdfMeasureCacheKey]float32, 512)
	}
	if _, exists := sdfMeasureCache[key]; !exists {
		sdfMeasureCacheLen++
	}
	sdfMeasureCache[key] = w
	if sdfMeasureCacheLen > sdfMeasureCacheMax {
		sdfMeasureCache = make(map[sdfMeasureCacheKey]float32, 512)
		sdfMeasureCacheLen = 0
	}
}

func sdfMeasureCacheClear() {
	sdfMeasureCacheMu.Lock()
	defer sdfMeasureCacheMu.Unlock()
	sdfMeasureCache = nil
	sdfMeasureCacheLen = 0
}

func cachedMeasureTextSDF(text string, fontSize float32, bold, italic, mono, preview bool) float32 {
	key := sdfMeasureCacheKey{
		text:     text,
		fontSize: uint16(fontSize*2 + 0.5),
		flags:    sdfMeasureCacheFlags(bold, italic, mono, preview),
	}
	if w, ok := sdfMeasureCacheLookup(key); ok {
		return w
	}
	w := measureTextSDF(text, fontSize, bold, italic, mono, preview)
	sdfMeasureCacheStore(key, w)
	return w
}
