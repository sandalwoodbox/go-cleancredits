package render

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

type Form struct {
	Container *fyne.Container

	StartFrame    binding.Int
	EndFrame      binding.Int
	InpaintRadius binding.Int
}

type RenderSettings struct {
	InpaintRadius int
}

func NewForm(frameCount int) Form {
	f := Form{
		StartFrame:    binding.NewInt(),
		EndFrame:      binding.NewInt(),
		InpaintRadius: binding.NewInt(),
	}
	f.Container = container.New(
		layout.NewGridLayout(3),
		widget.NewLabel("Start frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.StartFrame), ccWidget.NewIntEntryWithData(f.StartFrame),
		widget.NewLabel("End frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.EndFrame), ccWidget.NewIntEntryWithData(f.EndFrame),
		widget.NewLabel("Inpaint radius"), ccWidget.NewIntSliderWithData(0, 10, f.InpaintRadius), ccWidget.NewIntEntryWithData(f.InpaintRadius),
	)
	return f
}

func (f Form) OnChange(fn func()) {
	l := binding.NewDataListener(fn)
	f.StartFrame.AddListener(l)
	f.EndFrame.AddListener(l)
	f.InpaintRadius.AddListener(l)
}
