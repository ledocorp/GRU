package ui

import "sync"

const shapedMeasureCacheMax = 2048

type shapedMeasureCacheKey struct {
	text     string
	fontSize uint16
	flags    uint8 // bit0 bold, bit1 italic, bit2 mono, bit3 preview
}

var (
	shapedMeasureCacheMu  sync.RWMutex
	shapedMeasureCache    map[shapedMeasureCacheKey]float32
	shapedMeasureCacheLen int
)

func shapedMeasureCacheFlags(bold, italic, mono, preview bool) uint8 {
	var f uint8
	if bold {
		f |= 1
	}
	if italic {
		f |= 2
	}
	if mono {
		f |= 4
	}
	if preview {
		f |= 8
	}
	return f
}

func shapedMeasureCacheLookup(key shapedMeasureCacheKey) (float32, bool) {
	shapedMeasureCacheMu.RLock()
	defer shapedMeasureCacheMu.RUnlock()
	w, ok := shapedMeasureCache[key]
	return w, ok
}

func shapedMeasureCacheStore(key shapedMeasureCacheKey, w float32) {
	shapedMeasureCacheMu.Lock()
	defer shapedMeasureCacheMu.Unlock()
	if shapedMeasureCache == nil {
		shapedMeasureCache = make(map[shapedMeasureCacheKey]float32, 256)
	}
	if _, exists := shapedMeasureCache[key]; !exists {
		shapedMeasureCacheLen++
	}
	shapedMeasureCache[key] = w
	if shapedMeasureCacheLen > shapedMeasureCacheMax {
		shapedMeasureCache = make(map[shapedMeasureCacheKey]float32, 256)
		shapedMeasureCacheLen = 0
	}
}

func shapedMeasureCacheClear() {
	shapedMeasureCacheMu.Lock()
	defer shapedMeasureCacheMu.Unlock()
	shapedMeasureCache = nil
	shapedMeasureCacheLen = 0
}
