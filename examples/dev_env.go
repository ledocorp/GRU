// Package examples — development helpers (Air, local iteration).
package examples

import (
	"os"

	"github.com/ledocorp/gru/examples/startuppath"
)

func devEnvAlias(gru, gory string) string {
	if v := os.Getenv(gru); v != "" {
		return v
	}
	return os.Getenv(gory)
}

// DevSceneTitle returns GRU_SCENE when set (GORY_SCENE alias; exact match to Scene.Title()).
func DevSceneTitle() string { return devEnvAlias("GRU_SCENE", "GORY_SCENE") }

// DevOpenFilePath returns GRU_OPEN_FILE when set (GORY_OPEN_FILE alias; path passed to Notepad on Build).
func DevOpenFilePath() string { return devEnvAlias("GRU_OPEN_FILE", "GORY_OPEN_FILE") }

// StartupOpenFilePath returns a file to open on launch: first non-flag CLI arg,
// then GRU_OPEN_FILE / GORY_OPEN_FILE (Air/dev). Windows Explorer "Open with" passes the selected
// file as argv[1].
func StartupOpenFilePath() string {
	return startuppath.Resolve(os.Args, DevOpenFilePath())
}

// defaultStartupSceneTitle is implemented in dev_env_startup_*.go (demo vs -tags notepad).

// ResolveStartupSceneIndex picks the scene index for main.go startup.
// When GRU_SCENE (or GORY_SCENE alias) is set, the first factory whose Title() matches wins.
// On Android, defaults to Notepad when unset.
func ResolveStartupSceneIndex(factories []func() Scene, defaultIndex int) int {
	title := DevSceneTitle()
	if title == "" {
		title = defaultStartupSceneTitle()
	}
	if title == "" {
		return defaultIndex
	}
	for i, f := range factories {
		if f().Title() == title {
			return i
		}
	}
	return defaultIndex
}
