package main

import (
	"fmt"
	"os"
	"strconv"

	"fyne.io/fyne"
	"fyne.io/fyne/layout"
	"fyne.io/fyne/theme"
	"fyne.io/fyne/widget"
)

func runGUI(a fyne.App) int {
	var rt int
	a.Settings().SetTheme(theme.LightTheme())

	w := a.NewWindow("Ableton Live Tempo Randomizer")

	verLabel := widget.NewLabel("Installed Versions")
	versions := widget.NewSelect(vers, func(string) {})
	versions.SetSelected(*ver)
	versions.OnChanged = func(string) {
		*ver = versions.Selected
	}
	minLabel := widget.NewLabel("Minimum BPM")
	maxLabel := widget.NewLabel("Maximum BPM")

	minSlider := widget.NewSlider(20, 999)
	minEntry := widget.NewEntry()

	maxSlider := widget.NewSlider(20, 999)
	maxEntry := widget.NewEntry()

	randomizer := widget.NewButton("Gerenate Random BPM", func() {
		rt = randomTempo(int(minSlider.Value), int(maxSlider.Value))
		a.Quit()
	})

	quitButton := widget.NewButton("Quit", func() {
		a.Quit()
		os.Exit(0)
	})

	minSlider.Value = float64(*min)
	minEntry.SetText(fmt.Sprintf("%d", *min))

	maxSlider.Value = float64(*max)
	maxEntry.SetText(fmt.Sprintf("%d", *max))

	minSlider.OnChanged = func(float64) {
		if minSlider.Value >= maxSlider.Value {
			maxSlider.Value = minSlider.Value
			maxEntry.SetText(fmt.Sprintf("%.f", minSlider.Value))
		}
		minEntry.SetText(fmt.Sprintf("%.f", minSlider.Value))
	}
	minEntry.OnChanged = func(string) {
		minVal, err := strconv.Atoi(minEntry.Text)
		if err != nil {
			minEntry.SetText("")
			return
		}
		if minVal >= 20 && minVal <= 999 {
			if minVal >= int(maxSlider.Value) {
				maxSlider.Value = minSlider.Value
				maxEntry.SetText(fmt.Sprintf("%.f", minSlider.Value))
			}
			minSlider.Value = float64(minVal)
			minSlider.Refresh()
		}
		if minVal > 999 {
			minEntry.SetText(fmt.Sprintf("%d", 999))
		}
	}

	maxSlider.OnChanged = func(float64) {
		if maxSlider.Value <= minSlider.Value {
			minSlider.Value = maxSlider.Value
			minEntry.SetText(fmt.Sprintf("%.f", maxSlider.Value))
		}
		maxEntry.SetText(fmt.Sprintf("%.f", maxSlider.Value))
	}
	maxEntry.OnChanged = func(string) {
		maxVal, err := strconv.Atoi(maxEntry.Text)
		if err != nil {
			minEntry.SetText("")
			return
		}
		if maxVal >= 20 && maxVal <= 999 {
			if maxVal <= int(minSlider.Value) {
				minSlider.Value = maxSlider.Value
				minEntry.SetText(fmt.Sprintf("%.f", maxSlider.Value))
			}
			maxSlider.Value = float64(maxVal)
			maxSlider.Refresh()
		}
		if maxVal > 999 {
			maxEntry.SetText(fmt.Sprintf("%d", 999))
		}
	}

	minBox := widget.NewVBox(
		minLabel,
		minSlider,
		minEntry,
	)

	maxBox := widget.NewVBox(
		maxLabel,
		maxSlider,
		maxEntry,
	)

	controls := fyne.NewContainerWithLayout(
		layout.NewGridLayout(2),
		minBox,
		maxBox,
	)

	all := fyne.NewContainerWithLayout(
		layout.NewVBoxLayout(),
		verLabel,
		versions,
		controls,
		randomizer,
		quitButton,
	)

	w.Resize(fyne.NewSize(600, 200))

	w.SetContent(all)

	w.ShowAndRun()
	return rt
}
