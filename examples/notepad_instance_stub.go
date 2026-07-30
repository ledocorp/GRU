//go:build !notepad

package examples

func StartupInstance(forwardPath string) (exit bool, stop func()) {
	return false, func() {}
}

func SetInstanceOpenFileHandler(func(string)) {}
