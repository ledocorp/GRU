//go:build !notepad && !prism

package examples

import "runtime"

func defaultStartupSceneTitle() string {
	if runtime.GOOS == "android" {
		return NotepadSceneTitle
	}
	return ""
}
