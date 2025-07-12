package main

import (
	"syscall"
)

var (
	user32             = syscall.NewLazyDLL("user32.dll")
	procRegisterHotKey = user32.NewProc("RegisterHotKey")
	procUnregisterHotKey = user32.NewProc("UnregisterHotKey")
)

const (
	MOD_ALT = 0x0001
	VK_P    = 0x50
)

func RegisterHotKey(hwnd uintptr, id int) bool {
	ret, _, _ := procRegisterHotKey.Call(hwnd, uintptr(id), MOD_ALT, VK_P)
	return ret != 0
}

func UnregisterHotKey(hwnd uintptr, id int) {
	procUnregisterHotKey.Call(hwnd, uintptr(id))
}
