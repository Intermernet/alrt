package main

import (
	"os"
)

var abletonBase = os.Getenv("HOME") + "/Library/Preferences/Ableton/"

func getDefaultTemplate(ver string) string {
	return abletonBase + "Live " + ver + "/Template.als"
}
