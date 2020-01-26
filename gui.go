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

func runGUI(a fyne.App, minBPM, maxBPM int) int {
	var rt int
	//a := app.New()
	a.Settings().SetTheme(theme.LightTheme())

	//fmt.Println(a.Settings().Scale())

	w := a.NewWindow("Ableton Live Tempo Randomizer")

	minLabel := widget.NewLabel("Minimum BPM")
	maxLabel := widget.NewLabel("Maximum BPM")

	minSlider := widget.NewSlider(20, 999)
	min := widget.NewEntry()

	maxSlider := widget.NewSlider(20, 999)
	max := widget.NewEntry()

	randomizer := widget.NewButton("Gerenate Random BPM", func() {
		rt = randomTempo(int(minSlider.Value), int(maxSlider.Value))
		a.Quit()
	})

	quitButton := widget.NewButton("Quit", func() {
		a.Quit()
		os.Exit(0)
	})

	minSlider.Value = float64(minBPM)
	min.SetText(fmt.Sprintf("%d", minBPM))

	maxSlider.Value = float64(maxBPM)
	max.SetText(fmt.Sprintf("%d", maxBPM))

	minSlider.OnChanged = func(float64) {
		if minSlider.Value >= maxSlider.Value {
			maxSlider.Value = minSlider.Value
			max.SetText(fmt.Sprintf("%.f", minSlider.Value))
		}
		min.SetText(fmt.Sprintf("%.f", minSlider.Value))
	}
	min.OnChanged = func(string) {
		minVal, err := strconv.Atoi(min.Text)
		if err != nil {
			min.SetText("")
			return
		}
		if minVal >= 20 && minVal <= 999 {
			if minVal >= int(maxSlider.Value) {
				maxSlider.Value = minSlider.Value
				max.SetText(fmt.Sprintf("%.f", minSlider.Value))
			}
			minSlider.Value = float64(minVal)
			minSlider.Refresh()
		}
		if minVal > 999 {
			min.SetText(fmt.Sprintf("%d", 999))
		}
	}

	maxSlider.OnChanged = func(float64) {
		if maxSlider.Value <= minSlider.Value {
			minSlider.Value = maxSlider.Value
			min.SetText(fmt.Sprintf("%.f", maxSlider.Value))
		}
		max.SetText(fmt.Sprintf("%.f", maxSlider.Value))
	}
	max.OnChanged = func(string) {
		maxVal, err := strconv.Atoi(max.Text)
		if err != nil {
			min.SetText("")
			return
		}
		if maxVal >= 20 && maxVal <= 999 {
			if maxVal <= int(minSlider.Value) {
				minSlider.Value = maxSlider.Value
				min.SetText(fmt.Sprintf("%.f", maxSlider.Value))
			}
			maxSlider.Value = float64(maxVal)
			maxSlider.Refresh()
		}
		if maxVal > 999 {
			max.SetText(fmt.Sprintf("%d", 999))
		}
	}

	minBox := widget.NewVBox(
		minLabel,
		minSlider,
		min,
	)

	maxBox := widget.NewVBox(
		maxLabel,
		maxSlider,
		max,
	)

	controls := fyne.NewContainerWithLayout(
		layout.NewGridLayout(2),
		minBox,
		maxBox,
	)

	all := fyne.NewContainerWithLayout(
		layout.NewVBoxLayout(),
		controls,
		randomizer,
		quitButton,
	)

	w.Resize(fyne.NewSize(600, 200))

	w.SetContent(all)

	w.ShowAndRun()
	return rt
}
