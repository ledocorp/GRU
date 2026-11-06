package ui

import (
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

// gruLog prints a Gru: diagnostic line. On Android, fmt stdout is invisible in
// logcat — route through raylib TraceLog (Warning) so S1 soft-bold/outline/
// sticky messages show under the default LogWarning filter.
func gruLog(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	if AndroidGLES() {
		rl.TraceLog(rl.LogWarning, "%s", msg)
		return
	}
	fmt.Println(msg)
}
