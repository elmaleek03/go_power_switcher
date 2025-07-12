package main

import (
	"syscall"
	"unsafe"
)

// Step 1: Define the structure
type SystemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

// Step 2: Declare and call the Windows API function
var (
	modkernel32              = syscall.NewLazyDLL("kernel32.dll")
	procGetSystemPowerStatus = modkernel32.NewProc("GetSystemPowerStatus")
)

func GetSystemPowerStatus() (*SystemPowerStatus, error) {
	var sps SystemPowerStatus
	ret, _, err := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&sps)))
	if ret == 0 {
		return nil, err
	}
	return &sps, nil
}
