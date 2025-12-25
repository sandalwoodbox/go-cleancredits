package mask

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

const (
	HueMax = 179
	SatMax = 255
	ValMax = 255
)

type Form struct {
	Container *fyne.Container

	Frame binding.Int
	Grow  binding.Int

	HueMin binding.Int
	HueMax binding.Int
	SatMin binding.Int
	SatMax binding.Int
	ValMin binding.Int
	ValMax binding.Int

	CropLeft   binding.Int
	CropTop    binding.Int
	CropRight  binding.Int
	CropBottom binding.Int
}

func NewForm(frameCount, videoWidth, videoHeight int) Form {
	f := Form{
		Frame: binding.NewInt(),
		Grow:  binding.NewInt(),

		HueMin: binding.NewInt(),
		HueMax: binding.NewInt(),
		SatMin: binding.NewInt(),
		SatMax: binding.NewInt(),
		ValMin: binding.NewInt(),
		ValMax: binding.NewInt(),

		CropLeft:   binding.NewInt(),
		CropTop:    binding.NewInt(),
		CropRight:  binding.NewInt(),
		CropBottom: binding.NewInt(),
	}
	f.Container = container.New(
		layout.NewGridLayout(3),
		widget.NewLabel("Frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.Frame), ccWidget.NewIntEntryWithData(f.Frame),
		widget.NewLabel("Grow"), ccWidget.NewIntSliderWithData(0, videoHeight, f.Grow), widget.NewLabel(""),

		widget.NewLabel("Hue / Saturation / Value"), widget.NewLabel(""), widget.NewLabel(""),
		widget.NewLabel("Hue Min"), ccWidget.NewIntSliderWithData(0, HueMax, f.HueMin), widget.NewLabel(""),
		widget.NewLabel("Hue Max"), ccWidget.NewIntSliderWithData(0, HueMax, f.HueMax), widget.NewLabel(""),
		widget.NewLabel("Sat Min"), ccWidget.NewIntSliderWithData(0, SatMax, f.SatMin), widget.NewLabel(""),
		widget.NewLabel("Sat Max"), ccWidget.NewIntSliderWithData(0, SatMax, f.SatMax), widget.NewLabel(""),
		widget.NewLabel("Val Min"), ccWidget.NewIntSliderWithData(0, ValMax, f.ValMin), widget.NewLabel(""),
		widget.NewLabel("Val Max"), ccWidget.NewIntSliderWithData(0, ValMax, f.ValMax), widget.NewLabel(""),

		widget.NewLabel("Crop"), widget.NewLabel(""), widget.NewLabel(""),
		widget.NewLabel("Left"), ccWidget.NewIntSliderWithData(0, videoWidth, f.CropLeft), widget.NewLabel(""),
		widget.NewLabel("Top"), ccWidget.NewIntSliderWithData(0, videoHeight, f.CropTop), widget.NewLabel(""),
		widget.NewLabel("Right"), ccWidget.NewIntSliderWithData(0, videoWidth, f.CropRight), widget.NewLabel(""),
		widget.NewLabel("Bottom"), ccWidget.NewIntSliderWithData(0, videoHeight, f.CropBottom), widget.NewLabel(""),
	)
	return f
}

func (f Form) OnChange(fn func()) {
	f.Frame.AddListener(binding.NewDataListener(fn))
	f.Grow.AddListener(binding.NewDataListener(fn))

	f.HueMin.AddListener(binding.NewDataListener(fn))
	f.HueMax.AddListener(binding.NewDataListener(fn))
	f.SatMin.AddListener(binding.NewDataListener(fn))
	f.SatMax.AddListener(binding.NewDataListener(fn))
	f.ValMin.AddListener(binding.NewDataListener(fn))
	f.ValMax.AddListener(binding.NewDataListener(fn))

	f.CropLeft.AddListener(binding.NewDataListener(fn))
	f.CropTop.AddListener(binding.NewDataListener(fn))
	f.CropRight.AddListener(binding.NewDataListener(fn))
	f.CropBottom.AddListener(binding.NewDataListener(fn))
}
