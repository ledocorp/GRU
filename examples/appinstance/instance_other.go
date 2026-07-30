//go:build !windows && !linux

package appinstance

func startup(forwardPath string, onOpen func(path string)) (exit bool, stop func()) {
	return false, func() {}
}
