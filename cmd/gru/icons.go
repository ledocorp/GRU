package main

import (
	"fmt"
	"os"
	"os/exec"
)

// runIcons handles: gru icons regen
func runIcons(args []string) error {
	if len(args) == 0 || args[0] != "regen" {
		return fmt.Errorf("usage: gru icons regen")
	}
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	cmd := exec.Command("go", "run", "./scripts/build/gen_app_icon.go")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	fmt.Println("icons: embedded PNG/ICO, resource.syso, and packaging paths updated")
	fmt.Println("next: go run .   (or gru package windows)")
	return nil
}
