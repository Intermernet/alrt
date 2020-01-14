package main

import (
	"os"
)

var abletonBase = os.Getenv("USERPROFILE") + "\\AppData\\Roaming\\Ableton\\"

func getDefaultTemplate(ver string) string {
	return abletonBase + "Live " + ver + "\\Preferences\\Template.als"
}
