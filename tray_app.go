package main

import (
	"fmt"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"golang.org/x/sys/windows"
)

var (
	AppVersion          = "dev"
	currentOverlay      windows.GUID
	currentOverlayIndex int
	overlays            = []windows.GUID{Efficiency, Balanced, Performance}
	eff                 *systray.MenuItem
	bal                 *systray.MenuItem
	perf                *systray.MenuItem
	hotkeyCh            = make(chan struct{}, 5) // buffered to prevent blocking
)

func onReady() {
	if !RegisterHotKey(0, 1) {
		fmt.Println("Failed to register hotkey Alt+P")
	}

	systray.SetTitle("Power Mode Switcher")
	info := systray.AddMenuItem("made with ❤️ by akinos", "")
	systray.AddMenuItem(AppVersion, "")
	systray.AddSeparator()
	eff = systray.AddMenuItem("Best Efficiency", "")
	bal = systray.AddMenuItem("Balanced", "")
	perf = systray.AddMenuItem("Best Performance", "")
	systray.AddSeparator()
	startupItem := systray.AddMenuItemCheckbox("Run at Startup", "Launch app when Windows starts", isAutoStartEnabled())
	exit := systray.AddMenuItem("Exit", "Exit app")

	pluggedIn := isPluggedIn()

	if current, err := GetCurrentPowerOverlay(pluggedIn); err == nil {
		currentOverlay = *current
	} else {
		currentOverlay = Balanced
	}

	for i, g := range overlays {
		if g == currentOverlay {
			currentOverlayIndex = i
			break
		}
	}

	setTrayIcon(currentOverlay)
	updateTooltip(currentOverlay)
	updateCheckedMenu(currentOverlay)

	go func() {
		// Prevent click interaction — consume and ignore
		for {
			<-info.ClickedCh
		}
	}()

	// 🔁 Handle menu clicks
	go func() {
		for {
			select {
			case <-eff.ClickedCh:
				setOverlay(Efficiency, isPluggedIn())
			case <-bal.ClickedCh:
				setOverlay(Balanced, isPluggedIn())
			case <-perf.ClickedCh:
				setOverlay(Performance, isPluggedIn())
			case <-exit.ClickedCh:
				systray.Quit()
			case <-startupItem.ClickedCh:
				if startupItem.Checked() {
					removeAutoStart()
					startupItem.Uncheck()
				} else {
					err := addAutoStart()
					if err != nil {
						fmt.Println("Failed to enable startup:", err)
					} else {
						startupItem.Check()
					}
				}

			}
		}
	}()

	// 🔁 Hotkey listener goroutine
	go func() {
		var msg struct {
			hwnd    uintptr
			message uint32
			wParam  uintptr
			lParam  uintptr
			time    uint32
			pt      struct{ x, y int32 }
		}

		for {
			ret, _, _ := syscall.Syscall6(
				user32.NewProc("GetMessageW").Addr(),
				4,
				uintptr(unsafe.Pointer(&msg)),
				0, 0, 0,
				0, 0,
			)
			if ret == 0 {
				break
			}
			if msg.message == 0x0312 && msg.wParam == 1 { // WM_HOTKEY and ID=1
				select {
				case hotkeyCh <- struct{}{}:
				default:
					// If channel is full, drop (prevent overflow)
				}
			}
		}
	}()

	// 🔁 Hotkey processor (debounced)
	go func() {
		for range hotkeyCh {
			rotatePowerMode()
			time.Sleep(150 * time.Millisecond)
		}
	}()

	// 🔁 External power overlay polling
	go func() {
		for {
			time.Sleep(5 * time.Second)
			pluggedIn := isPluggedIn()
			if overlay, err := GetCurrentPowerOverlay(pluggedIn); err == nil {
				if *overlay != currentOverlay {
					currentOverlay = *overlay
					for i, g := range overlays {
						if g == currentOverlay {
							currentOverlayIndex = i
							break
						}
					}
					setTrayIcon(currentOverlay)
					updateTooltip(currentOverlay)
					updateCheckedMenu(currentOverlay)
				}
			}
		}
	}()
}

func onExit() {
	UnregisterHotKey(0, 1)

}
func updateCheckedMenu(guid windows.GUID) {
	switch {
	case guid == Efficiency:
		eff.Check()
		bal.Uncheck()
		perf.Uncheck()
	case guid == Balanced:
		eff.Uncheck()
		bal.Check()
		perf.Uncheck()
	case guid == Performance:
		eff.Uncheck()
		bal.Uncheck()
		perf.Check()
	}
}

func setOverlay(guid windows.GUID, pluggedIn bool) {
	if err := SetPowerOverlay(guid, pluggedIn); err != nil {
		fmt.Println("Error setting power mode:", err)
		return
	}

	currentOverlay = guid
	fmt.Println("Power mode set to:", GetPowerModeName(guid))
	setTrayIcon(guid)
	updateTooltip(guid)

	// Update checkmarks
	switch {
	case guid == Efficiency:
		eff.Check()
		bal.Uncheck()
		perf.Uncheck()
	case guid == Balanced:
		eff.Uncheck()
		bal.Check()
		perf.Uncheck()
	case guid == Performance:
		eff.Uncheck()
		bal.Uncheck()
		perf.Check()
	}
}

func isPluggedIn() bool {
	if status, err := GetSystemPowerStatus(); err == nil {
		return status.ACLineStatus == 1
	}
	return false
}

func setTrayIcon(guid windows.GUID) {
	switch {
	case guid == Efficiency:
		systray.SetIcon(iconEfficiency)
	case guid == Balanced:
		systray.SetIcon(iconBalanced)
	case guid == Performance:
		systray.SetIcon(iconPerformance)
	default:
		systray.SetIcon(iconBalanced) // fallback
	}
}

func updateTooltip(guid windows.GUID) {
	mode := GetPowerModeName(guid)
	systray.SetTooltip("Current Power Mode: " + mode)
}

func rotatePowerMode() {
	pluggedIn := isPluggedIn()
	currentOverlayIndex = (currentOverlayIndex + 1) % len(overlays)
	newMode := overlays[currentOverlayIndex]
	setOverlay(newMode, pluggedIn)
}
