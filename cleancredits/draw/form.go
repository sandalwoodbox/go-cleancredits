package draw

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/settings"
	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

const (
	Include = "Always inpaint"
	Exclude = "Never inpaint"
	Reset   = "Reset"
)

type Form struct {
	Container *fyne.Container

	Frame binding.Int
	Mode  binding.String
	Size  binding.Int
}

func NewForm(frameCount int) Form {
	f := Form{
		Frame: binding.NewInt(),
		Mode:  binding.NewString(),
		Size:  binding.NewInt(),
	}
	err := f.Mode.Set(Include)
	if err != nil {
		fmt.Println("Error setting draw mode: ", err)
	}
	f.Container = container.New(
		layout.NewVBoxLayout(),
		container.New(
			layout.NewGridLayout(3),
			widget.NewLabel("Frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.Frame), ccWidget.NewIntEntryWithData(f.Frame),
			widget.NewLabel("Mode"), widget.NewSelectWithData([]string{Include, Exclude}, f.Mode), widget.NewLabel(""),
			widget.NewLabel("Size"), ccWidget.NewIntSliderWithData(0, 100, f.Size), ccWidget.NewIntEntryWithData(f.Size),
		),
	)
	return f
}

func (f Form) OnChange(fn func()) {
	l := binding.NewDataListener(fn)
	f.Frame.AddListener(l)
	f.Mode.AddListener(l)
}

func (f Form) Settings() (settings.Draw, error) {
	frame, err := f.Frame.Get()
	if err != nil {
		return settings.Draw{}, fmt.Errorf("getting frame: %v", err)
	}
	return settings.Draw{
		Frame: frame,
	}, nil
}
