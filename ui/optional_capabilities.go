package ui

import "strings"

// Optional capability hooks — default package ui stays free of chroma, goldmark,
// gospell, and media/imaging. Import ui/syntax, ui/markdown, and/or ui/spell
// (blank import is enough) to register real implementations.

var (
	highlightSyntaxImpl        func(source, language string) []TextSpan
	setChromaStyleImpl         func(name string)
	chromaStyleNameImpl        func() string
	markdownToDocumentSpecImpl func(id, title, source string) DocumentSpec
	hunspellFactory            func(affPath, dicPath string, extraWords ...string) (SpellChecker, error)
)

// CuratedChromaStyles are editor/preview syntax themes exposed in settings UIs.
// Default is a single placeholder; ui/syntax replaces this with the curated list.
var CuratedChromaStyles = []string{"github"}

// RegisterHighlightSyntax installs the Chroma-backed highlighter (called from ui/syntax).
func RegisterHighlightSyntax(fn func(source, language string) []TextSpan) {
	highlightSyntaxImpl = fn
}

// RegisterChromaStyle installs style get/set and the curated theme name list (ui/syntax).
func RegisterChromaStyle(set func(name string), get func() string, curated []string) {
	setChromaStyleImpl = set
	chromaStyleNameImpl = get
	if len(curated) > 0 {
		CuratedChromaStyles = curated
	}
}

// RegisterMarkdownToDocumentSpec installs the goldmark DocumentSpec bridge (ui/markdown).
func RegisterMarkdownToDocumentSpec(fn func(id, title, source string) DocumentSpec) {
	markdownToDocumentSpecImpl = fn
}

// RegisterHunspellFactory installs the gospell Hunspell backend (ui/spell).
func RegisterHunspellFactory(fn func(affPath, dicPath string, extraWords ...string) (SpellChecker, error)) {
	hunspellFactory = fn
}

// NormalizeSyntaxLanguage maps common fence labels and file extensions to lexer names.
// Lives in core ui (no chroma) so callers can normalize without linking syntax.
func NormalizeSyntaxLanguage(lang string) string {
	switch strings.ToLower(strings.TrimSpace(lang)) {
	case "go", "golang":
		return "go"
	case "js", "javascript":
		return "javascript"
	case "ts", "typescript":
		return "typescript"
	case "py", "python":
		return "python"
	case "c":
		return "c"
	case "cpp", "c++", "cc", "cxx", "hpp", "hxx":
		return "cpp"
	case "java":
		return "java"
	case "lua":
		return "lua"
	case "rs", "rust":
		return "rust"
	case "sh", "bash", "shell", "zsh":
		return "bash"
	case "yml":
		return "yaml"
	case "md", "markdown":
		return "markdown"
	case "sql":
		return "sql"
	default:
		return strings.ToLower(strings.TrimSpace(lang))
	}
}

// HighlightSyntax tokenizes source into RichText spans when ui/syntax is linked;
// otherwise returns a single plain span.
func HighlightSyntax(source, language string) []TextSpan {
	if highlightSyntaxImpl != nil {
		return highlightSyntaxImpl(source, language)
	}
	source = strings.ReplaceAll(source, "\r\n", "\n")
	if source == "" {
		return nil
	}
	return []TextSpan{{Text: source}}
}

// ChromaStyleName returns the active Chroma style registry name (default "github").
func ChromaStyleName() string {
	if chromaStyleNameImpl != nil {
		return chromaStyleNameImpl()
	}
	return "github"
}

// SetChromaStyle selects a Chroma styles registry name for HighlightSyntax.
// No-op until ui/syntax is imported.
func SetChromaStyle(name string) {
	if setChromaStyleImpl != nil {
		setChromaStyleImpl(name)
	}
}

func markdownToDocumentSpecGoldmark(id, title, source string) DocumentSpec {
	if markdownToDocumentSpecImpl != nil {
		return markdownToDocumentSpecImpl(id, title, source)
	}
	return DocumentSpec{ID: id, Title: title}
}

// NewHunspellChecker loads dictionary files via the registered Hunspell factory.
// Returns ErrHunspellDictNotFound when ui/spell is not imported or load fails.
func NewHunspellChecker(affPath, dicPath string, extraWords ...string) (SpellChecker, error) {
	if hunspellFactory == nil {
		return nil, ErrHunspellDictNotFound
	}
	return hunspellFactory(affPath, dicPath, extraWords...)
}

// TryHunspellChecker resolves dictionary paths via [ResolveHunspellDict] and returns a
// checker, or nil and an error when files are missing or ui/spell is not linked.
func TryHunspellChecker(extraWords ...string) (SpellChecker, error) {
	if hunspellFactory == nil {
		return nil, ErrHunspellDictNotFound
	}
	aff, dic, ok := ResolveHunspellDict()
	if !ok {
		return nil, ErrHunspellDictNotFound
	}
	return hunspellFactory(aff, dic, extraWords...)
}
