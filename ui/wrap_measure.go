package ui

import (
	"sync"
	"unicode/utf8"
)

// wrapASCIISingle holds interned single-rune strings for ASCII wrap measure (no alloc).
var wrapASCIISingle [128]string

func init() {
	for i := 32; i < 127; i++ {
		wrapASCIISingle[i] = string(rune(i))
	}
}

type wrapAdvKey struct {
	r     rune
	size  uint16
	flags uint8
}

var wrapAdvCache sync.Map

func wrapStyleFlags(s Style) uint8 {
	var f uint8
	if styleDrawBold(s) {
		f |= 1
	}
	if s.Italic {
		f |= 2
	}
	if s.Mono {
		f |= 4
	}
	if s.PreviewFont {
		f |= 8
	}
	return f
}

// measureWrapRuneAdvance reports pixel width for one rune on the POD wrap hot path.
// Bypasses string-keyed SDF/shaped measure caches (map insert churn); uses a
// rune-keyed side cache warmed once per distinct rune+style.
func measureWrapRuneAdvance(r rune, s Style) float32 {
	fs := EffectiveFontSize(s)
	key := wrapAdvKey{r: r, size: uint16(fs*2 + 0.5), flags: wrapStyleFlags(s)}
	if v, ok := wrapAdvCache.Load(key); ok {
		return v.(float32)
	}
	text := wrapRuneString(r)
	bold, italic, mono, preview := styleDrawBold(s), s.Italic, s.Mono, s.PreviewFont
	var w float32
	if shapedTextReady() && textDrawUsesShaped(text, fs, bold, italic, mono, preview) {
		if raw, ok := shapedMeasureTextFRaw(text, fs, bold, italic, mono, preview); ok {
			w = raw * shapedMeasureRatioFor(fs, preview, text)
		}
	}
	if w <= 0 {
		w = measureTextSDF(text, fs, bold, italic, mono, preview)
	}
	wrapAdvCache.Store(key, w)
	return w
}

func wrapRuneString(r rune) string {
	if r >= 32 && r < 127 {
		return wrapASCIISingle[r]
	}
	var buf [utf8.UTFMax]byte
	n := utf8.EncodeRune(buf[:], r)
	return string(buf[:n])
}
