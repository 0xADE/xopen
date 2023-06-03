package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
)

var data = make(chan string, 1)

func ui() fyne.Window {
	a := app.New()
	w := a.NewWindow("Hello World")

	label := widget.NewLabel("Hello World!")
	w.SetContent(label)
	go updateLabel(label)
	return w
}

func updateLabel(label *widget.Label) {
	for {
		select {
		case text := <-data:
			if text != "" {
				label.SetText(text)
			}
		}
	}
}
