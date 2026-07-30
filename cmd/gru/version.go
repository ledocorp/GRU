package main

import (
	"fmt"

	"github.com/ledocorp/gru/internal/version"
)

func printVersion() {
	fmt.Printf("gru %s\n", version.Tool)
	fmt.Printf("app %s (%s)\n", version.App, version.Module)
	fmt.Printf("product %s\n", version.Product)
}
