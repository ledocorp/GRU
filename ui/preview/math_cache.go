package preview

import (
	"crypto/sha256"
	"encoding/hex"
	"sync"

	rl "github.com/gen2brain/raylib-go/raylib"
)

const mathTextureCacheMax = 64

type mathTextureCache struct {
	mu      sync.Mutex
	entries map[string]mathTextureEntry
	order   []string
	max     int
}

type mathTextureEntry struct {
	tex rl.Texture2D
	w   int32
	h   int32
}

var globalMathTextureCache = &mathTextureCache{
	entries: make(map[string]mathTextureEntry),
	max:     mathTextureCacheMax,
}

func mathCacheKey(source string, display bool) string {
	h := sha256.Sum256([]byte(source))
	mode := "inline"
	if display {
		mode = "display"
	}
	return mode + "|" + hex.EncodeToString(h[:8])
}

func (c *mathTextureCache) get(key string) (mathTextureEntry, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || e.tex.ID == 0 {
		return mathTextureEntry{}, false
	}
	c.touch(key)
	return e, true
}

func (c *mathTextureCache) put(key string, e mathTextureEntry) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if old, ok := c.entries[key]; ok && old.tex.ID != 0 && old.tex.ID != e.tex.ID {
		rl.UnloadTexture(old.tex)
	}
	c.entries[key] = e
	c.touch(key)
	for len(c.order) > c.max {
		evict := c.order[0]
		c.order = c.order[1:]
		if old, ok := c.entries[evict]; ok {
			if old.tex.ID != 0 {
				rl.UnloadTexture(old.tex)
			}
			delete(c.entries, evict)
		}
	}
}

func (c *mathTextureCache) touch(key string) {
	for i, k := range c.order {
		if k == key {
			c.order = append(c.order[:i], c.order[i+1:]...)
			break
		}
	}
	c.order = append(c.order, key)
}
