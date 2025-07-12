package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	powrprof                              = windows.NewLazySystemDLL("powrprof.dll")
	modPowrProf                           = windows.NewLazySystemDLL("powrprof.dll")
	procPowerGetUserConfiguredACPowerMode = modPowrProf.NewProc("PowerGetUserConfiguredACPowerMode")
	procPowerGetUserConfiguredDCPowerMode = modPowrProf.NewProc("PowerGetUserConfiguredDCPowerMode")
	procPowerSetACMode                    = powrprof.NewProc("PowerSetUserConfiguredACPowerMode")
	procPowerSetDCMode                    = powrprof.NewProc("PowerSetUserConfiguredDCPowerMode")
	Efficiency                            = windows.GUID{Data1: 0x961cc777, Data2: 0x2547, Data3: 0x4f9d, Data4: [8]byte{0x81, 0x74, 0x7d, 0x86, 0x18, 0x1b, 0x8a, 0x7a}}
	Balanced                              = windows.GUID{} // Empty
	Performance                           = windows.GUID{Data1: 0xded574b5, Data2: 0x45a0, Data3: 0x4f42, Data4: [8]byte{0x87, 0x37, 0x46, 0x34, 0x5c, 0x09, 0xc2, 0x38}}
)

func SetPowerOverlay(mode windows.GUID, pluggedIn bool) error {
	var ret uintptr
	var err error
	if pluggedIn {
		ret, _, err = procPowerSetACMode.Call(uintptr(unsafe.Pointer(&mode)))
	} else {
		ret, _, err = procPowerSetDCMode.Call(uintptr(unsafe.Pointer(&mode)))
	}
	if ret != 0 {
		return fmt.Errorf("PowerSetUserConfiguredPowerMode failed: %v", err)
	}
	return nil
}

func GetCurrentPowerOverlay(pluggedIn bool) (*windows.GUID, error) {
	var guid windows.GUID

	proc := procPowerGetUserConfiguredACPowerMode
	if !pluggedIn {
		proc = procPowerGetUserConfiguredDCPowerMode
	}

	ret, _, _ := proc.Call(uintptr(unsafe.Pointer(&guid)))
	if ret != 0 {
		return nil, fmt.Errorf("PowerGetUserConfiguredPowerMode failed with code 0x%x", ret)
	}

	return &guid, nil
}

func GetPowerModeName(guid windows.GUID) string {
	switch {
	case guid == Efficiency:
		return "Best Efficiency"
	case guid == Balanced:
		return "Balanced"
	case guid == Performance:
		return "Best Performance"
	default:
		return "Unknown Mode"
	}
}
