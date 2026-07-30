package ui

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrHunspellDictNotFound is returned when no en_US .aff/.dic pair is found.
var ErrHunspellDictNotFound = errors.New("hunspell en_US dictionary not found")

const (
	hunspellAffName = "en_US.aff"
	hunspellDicName = "en_US.dic"
)

// HunspellDictDirNames are subdirectories searched under each base path.
var HunspellDictDirNames = []string{
	"",
	"en_US",
	"dict",
	filepath.Join("dict", "en_US"),
	filepath.Join("assets", "dict", "en_US"),
}

// ResolveHunspellDict finds the first directory containing en_US.aff and en_US.dic.
// Search order: GRU_DICT_DIR (GORY_DICT_DIR alias), executable dir, user config gru-notepad/dict, assets/dict.
func ResolveHunspellDict() (affPath, dicPath string, ok bool) {
	for _, base := range hunspellSearchBases() {
		for _, sub := range HunspellDictDirNames {
			dir := base
			if sub != "" {
				dir = filepath.Join(base, sub)
			}
			aff := filepath.Join(dir, hunspellAffName)
			dic := filepath.Join(dir, hunspellDicName)
			if hunspellFileExists(aff) && hunspellFileExists(dic) {
				return aff, dic, true
			}
		}
	}
	return "", "", false
}

func hunspellSearchBases() []string {
	var bases []string
	if env := EnvOr("GRU_DICT_DIR", "GORY_DICT_DIR"); env != "" {
		bases = append(bases, env)
	}
	if exe, err := os.Executable(); err == nil {
		bases = append(bases, filepath.Dir(exe))
	}
	if cfg, err := os.UserConfigDir(); err == nil {
		bases = append(bases, filepath.Join(cfg, "gru-notepad", "dict"))
	}
	if wd, err := os.Getwd(); err == nil {
		bases = append(bases, filepath.Join(wd, "assets", "dict"))
	}
	return bases
}

func hunspellFileExists(path string) bool {
	st, err := os.Stat(path)
	return err == nil && !st.IsDir()
}
