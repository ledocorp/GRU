//go:build !notepad && !prism

package examples

import "runtime"

func defaultStartupSceneTitle() string {
	if runtime.GOOS == "android" {
		// App Shell validates insets / rem / BottomNav. Notepad is a separate
		// product tree (gru-notepad); wire after App Shell polish. GRU_SCENE overrides.
		return "App Shell (Go)"
	}
	return ""
}
