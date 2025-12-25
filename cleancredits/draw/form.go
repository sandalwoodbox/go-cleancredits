package draw

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

const (
	Include = "Always inpaint"
	Exclude = "Never inpaint"
)

type Form struct {
	Container *fyne.Container

	Frame binding.Int
	Mode  binding.String
}

func NewForm(frameCount int) Form {
	f := Form{
		Frame: binding.NewInt(),
		Mode:  binding.NewString(),
	}
	f.Container = container.New(
		layout.NewGridLayout(3),
		widget.NewLabel("Frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.Frame), ccWidget.NewIntEntryWithData(f.Frame),
		widget.NewLabel("Mode"), widget.NewSelectWithData([]string{Include, Exclude}, f.Mode), widget.NewLabel(""),
	)
	return f
}

func (f Form) OnChange(fn func()) {
	l := binding.NewDataListener(fn)
	f.Frame.AddListener(l)
	f.Mode.AddListener(l)
}
