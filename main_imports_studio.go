//go:build !prism && !grudemo

package main

import (
	_ "github.com/ledocorp/gru/apps/prism"
	_ "github.com/ledocorp/gru/foundry/scenes"
)
