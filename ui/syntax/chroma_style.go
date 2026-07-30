package syntax

import (
	"strings"
	"sync"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/styles"
)

// curatedChromaStyles are editor/preview syntax themes exposed in Gru Notepad settings.
// Names must exist in Chroma's styles registry.
var curatedChromaStyles = []string{
	"github",
	"monokai",
	"dracula",
	"nord",
	"one-dark",
	"solarized-dark",
	"solarized-light",
	"vs",
	"pygments",
	"catppuccin-mocha",
	"gruvbox-dark",
	"atom-one-dark",
}

var (
	chromaStyleMu sync.RWMutex
	chromaStyle   = "github"
)

func initCuratedStyles() []string {
	validated := make([]string, 0, len(curatedChromaStyles))
	for _, name := range curatedChromaStyles {
		if styles.Get(name) != nil {
			validated = append(validated, name)
		}
	}
	if len(validated) > 0 {
		curatedChromaStyles = validated
	}
	chromaStyle = "github"
	if styles.Get(chromaStyle) == nil {
		chromaStyle = "monokai"
	}
	return curatedChromaStyles
}

func chromaStyleName() string {
	chromaStyleMu.RLock()
	defer chromaStyleMu.RUnlock()
	return chromaStyle
}

func setChromaStyle(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "github"
	}
	if styles.Get(name) == nil {
		if styles.Get("github") != nil {
			name = "github"
		} else {
			name = "monokai"
		}
	}
	chromaStyleMu.Lock()
	chromaStyle = name
	chromaStyleMu.Unlock()
}

func resolveChromaStyle() *chroma.Style {
	name := chromaStyleName()
	if s := styles.Get(name); s != nil {
		return s
	}
	if s := styles.Get("github"); s != nil {
		return s
	}
	return styles.Get("monokai")
}
