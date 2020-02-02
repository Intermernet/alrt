package main

import (
	"os"
	"syscall"
)

var abletonBase = os.Getenv("USERPROFILE") + "\\AppData\\Roaming\\Ableton\\"

func getDefaultTemplate(ver string) string {
	return abletonBase + "Live " + ver + "\\Preferences\\Template.als"
}

func pressEnterKey() {
	const keyEnter = 28
	var dll = syscall.NewLazyDLL("user32.dll")
	var procKeyBd = dll.NewProc("keybd_event")
	procKeyBd.Call(uintptr(keyEnter), uintptr(keyEnter+0x80), uintptr(0|0x0008), 0)
	procKeyBd.Call(uintptr(keyEnter), uintptr(keyEnter+0x80), uintptr(0x0002|0x0008), 0)
}
