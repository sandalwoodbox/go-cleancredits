package mask

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
)

type Form struct {
	Container *fyne.Container

	Frame binding.Float
	Mode  binding.String
	Grow  binding.Float

	HueMin binding.Float
	HueMax binding.Float
	SatMin binding.Float
	SatMax binding.Float
	ValMin binding.Float
	ValMax binding.Float

	CropLeft   binding.Float
	CropTop    binding.Float
	CropRight  binding.Float
	CropBottom binding.Float
}

func NewForm(frameCount, videoWidth, videoHeight int) Form {
	form := Form{
		Frame: binding.NewFloat(),
		Mode:  binding.NewString(),
		Grow:  binding.NewFloat(),

		HueMin: binding.NewFloat(),
		HueMax: binding.NewFloat(),
		SatMin: binding.NewFloat(),
		SatMax: binding.NewFloat(),
		ValMin: binding.NewFloat(),
		ValMax: binding.NewFloat(),

		CropLeft:   binding.NewFloat(),
		CropTop:    binding.NewFloat(),
		CropRight:  binding.NewFloat(),
		CropBottom: binding.NewFloat(),
	}
	form.Container = container.New(
		layout.NewGridLayout(3),
		widget.NewLabel("Frame"), widget.NewSliderWithData(0, float64(frameCount)-1, form.Frame), widget.NewEntryWithData(binding.FloatToStringWithFormat(form.Frame, "%0.0f")),
		widget.NewLabel("Mode"), widget.NewSelectWithData([]string{Include, Exclude}, form.Mode), widget.NewLabel(""),
		widget.NewLabel("Grow"), widget.NewSliderWithData(0, float64(videoHeight), form.Grow), widget.NewLabel(""),

		widget.NewLabel("Hue / Saturation / Value"), widget.NewLabel(""), widget.NewLabel(""),
		widget.NewLabel("Hue Min"), widget.NewSliderWithData(0, HueMax, form.HueMin), widget.NewLabel(""),
		widget.NewLabel("Hue Max"), widget.NewSliderWithData(0, HueMax, form.HueMax), widget.NewLabel(""),
		widget.NewLabel("Sat Min"), widget.NewSliderWithData(0, SatMax, form.SatMin), widget.NewLabel(""),
		widget.NewLabel("Sat Max"), widget.NewSliderWithData(0, SatMax, form.SatMax), widget.NewLabel(""),
		widget.NewLabel("Val Min"), widget.NewSliderWithData(0, ValMax, form.ValMin), widget.NewLabel(""),
		widget.NewLabel("Val Max"), widget.NewSliderWithData(0, ValMax, form.ValMax), widget.NewLabel(""),

		widget.NewLabel("Crop"), widget.NewLabel(""), widget.NewLabel(""),
		widget.NewLabel("Left"), widget.NewSliderWithData(0, float64(videoWidth), form.CropLeft), widget.NewLabel(""),
		widget.NewLabel("Top"), widget.NewSliderWithData(0, float64(videoHeight), form.CropTop), widget.NewLabel(""),
		widget.NewLabel("Right"), widget.NewSliderWithData(0, float64(videoWidth), form.CropRight), widget.NewLabel(""),
		widget.NewLabel("Bottom"), widget.NewSliderWithData(0, float64(videoHeight), form.CropBottom), widget.NewLabel(""),
	)
	return form
}

func (f Form) OnChange(fn func()) {
	f.Frame.AddListener(binding.NewDataListener(fn))
	f.Mode.AddListener(binding.NewDataListener(fn))
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
