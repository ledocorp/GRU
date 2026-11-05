//go:build grudemo

package examples

// Public grudemo: no Studio footer sample-host launcher (ghostclock / island / …).

func SampleHostsEnabled() bool { return false }

func SampleHostHintX() float32 { return 92 }

func DrawSampleHostButtons(windowW, windowH int32) {
	_ = windowW
	_ = windowH
}

func TryLaunchSampleHostClick(windowW, windowH int32) (string, error) {
	_ = windowW
	_ = windowH
	return "", nil
}
