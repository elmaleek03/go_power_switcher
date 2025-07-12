package main

import (
	"os"

	"golang.org/x/sys/windows/registry"
)

const appName = "PowerModeSwitcher"

func addAutoStart() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetStringValue(appName, exePath)
}

func removeAutoStart() error {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.DeleteValue(appName)
}

func isAutoStartEnabled() bool {
	key, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Run`, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer key.Close()
	val, _, err := key.GetStringValue(appName)
	if err != nil {
		return false
	}
	exePath, _ := os.Executable()
	return val == exePath
}
