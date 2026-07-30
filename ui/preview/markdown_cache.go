package preview

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"sync"

	"github.com/yuin/goldmark/ast"

	"github.com/ledocorp/gru/ui"
)

type parseCacheEntry struct {
	blocks []ast.Node
	src    []byte
}

var markdownParseCache sync.Map // key: sha256 prefix → *parseCacheEntry

func markdownCacheKey(source string) string {
	sum := sha256.Sum256([]byte(source))
	return hex.EncodeToString(sum[:12])
}

func normalizeMarkdownSource(source string) []byte {
	return []byte(strings.ReplaceAll(source, "\r\n", "\n"))
}

// ParseMarkdownBlocksCached returns parsed blocks, reusing a prior parse for the same source.
func ParseMarkdownBlocksCached(source string) ([]ast.Node, []byte) {
	norm := normalizeMarkdownSource(source)
	key := markdownCacheKey(source)
	if v, ok := markdownParseCache.Load(key); ok {
		e := v.(*parseCacheEntry)
		if bytes.Equal(e.src, norm) {
			return e.blocks, e.src
		}
	}
	blocks, src := ParseMarkdownBlocks(source)
	entry := &parseCacheEntry{
		blocks: blocks,
		src:    append([]byte(nil), src...),
	}
	markdownParseCache.Store(key, entry)
	return entry.blocks, entry.src
}

// WarmMarkdownCache pre-parses source on a worker so the first preview open is faster.
func WarmMarkdownCache(source string) {
	if source == "" {
		return
	}
	key := markdownCacheKey(source)
	if _, ok := markdownParseCache.Load(key); ok {
		return
	}
	ui.SubmitAsyncBg(func() {
		ParseMarkdownBlocksCached(source)
	})
}
