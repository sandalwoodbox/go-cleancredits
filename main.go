package main

import (
	"fyne.io/fyne/v2/app"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits"
)

func main() {
	a := app.NewWithID("com.github.sandalwoodbox.cleancredits")
	w := cleancredits.NewMainWindow(a)

	w.ShowAndRun()
}
