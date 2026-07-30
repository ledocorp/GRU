//go:build !android && !windows

package tray

import (
	"fyne.io/systray"
)

var active Config

// Start launches the tray icon goroutine. No-op when Icon is empty.
func Start(c Config) {
	if len(c.Icon) == 0 {
		return
	}
	active = c
	go systray.Run(onReady, onExit)
}

// Stop removes the tray icon.
func Stop() {
	systray.Quit()
}

func onReady() {
	systray.SetIcon(active.Icon)
	if active.Tooltip != "" {
		systray.SetTooltip(active.Tooltip)
	}
	show := systray.AddMenuItem("Show", "Show window")
	quit := systray.AddMenuItem("Quit", "Quit")
	go func() {
		for {
			select {
			case <-show.ClickedCh:
				if active.OnShow != nil {
					active.OnShow()
				}
			case <-quit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	if active.OnQuit != nil {
		active.OnQuit()
	}
}
