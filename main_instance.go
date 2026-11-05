//go:build !notepad

package main

import "github.com/ledocorp/gru/examples"

func startupProductInstance() (exit bool, stop func()) {
	return examples.StartupInstance(examples.StartupOpenFilePath())
}
