package mask

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
	HueMax = 179
	SatMax = 255
	ValMax = 255
)

const (
	Include = "Always inpaint"
	Exclude = "Never inpaint"
)

type Form struct {
	Container *fyne.Container

	Frame binding.Int
	Mode  binding.String // TODO: implement this more fully - it's the mode of the current mask
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
		Mode:  binding.NewString(),

		HueMin: binding.NewInt(),
		HueMax: binding.NewInt(),
		SatMin: binding.NewInt(),
		SatMax: binding.NewInt(),
		ValMin: binding.NewInt(),
		ValMax: binding.NewInt(),
		Grow:   binding.NewInt(),

		CropLeft:   binding.NewInt(),
		CropTop:    binding.NewInt(),
		CropRight:  binding.NewInt(),
		CropBottom: binding.NewInt(),
	}
	err := f.Mode.Set(Include)
	if err != nil {
		fmt.Println("Error setting Mode: ", err)
	}
	err = f.HueMax.Set(HueMax)
	if err != nil {
		fmt.Println("Error setting HueMax: ", err)
	}
	err = f.SatMax.Set(SatMax)
	if err != nil {
		fmt.Println("Error setting SatMax: ", err)
	}
	err = f.ValMax.Set(ValMax)
	if err != nil {
		fmt.Println("Error setting ValMax: ", err)
	}
	err = f.CropRight.Set(videoWidth)
	if err != nil {
		fmt.Println("Error setting CropRight: ", err)
	}
	err = f.CropBottom.Set(videoHeight)
	if err != nil {
		fmt.Println("Error setting CropBottom: ", err)
	}
	f.Container = container.New(
		layout.NewVBoxLayout(),
		container.New(
			layout.NewGridLayout(3),
			widget.NewLabel("Frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.Frame), ccWidget.NewIntEntryWithData(0, frameCount-1, f.Frame),

			widget.NewLabel("Hue / Saturation / Value"), widget.NewLabel(""), widget.NewLabel(""),
			widget.NewLabel("Hue Min"), ccWidget.NewIntSliderWithData(0, HueMax, f.HueMin), ccWidget.NewIntEntryWithData(0, HueMax, f.HueMin),
			widget.NewLabel("Hue Max"), ccWidget.NewIntSliderWithData(0, HueMax, f.HueMax), ccWidget.NewIntEntryWithData(0, HueMax, f.HueMax),
			widget.NewLabel("Sat Min"), ccWidget.NewIntSliderWithData(0, SatMax, f.SatMin), ccWidget.NewIntEntryWithData(0, SatMax, f.SatMin),
			widget.NewLabel("Sat Max"), ccWidget.NewIntSliderWithData(0, SatMax, f.SatMax), ccWidget.NewIntEntryWithData(0, SatMax, f.SatMax),
			widget.NewLabel("Val Min"), ccWidget.NewIntSliderWithData(0, ValMax, f.ValMin), ccWidget.NewIntEntryWithData(0, ValMax, f.ValMin),
			widget.NewLabel("Val Max"), ccWidget.NewIntSliderWithData(0, ValMax, f.ValMax), ccWidget.NewIntEntryWithData(0, ValMax, f.ValMax),
			widget.NewLabel("Grow"), ccWidget.NewIntSliderWithData(0, 20, f.Grow), ccWidget.NewIntEntryWithData(0, 20, f.Grow),

			widget.NewLabel("Crop"), widget.NewLabel(""), widget.NewLabel(""),
			widget.NewLabel("Left"), ccWidget.NewIntSliderWithData(0, videoWidth, f.CropLeft), ccWidget.NewIntEntryWithData(0, videoWidth, f.CropLeft),
			widget.NewLabel("Top"), ccWidget.NewIntSliderWithData(0, videoHeight, f.CropTop), ccWidget.NewIntEntryWithData(0, videoHeight, f.CropTop),
			widget.NewLabel("Right"), ccWidget.NewIntSliderWithData(0, videoWidth, f.CropRight), ccWidget.NewIntEntryWithData(0, videoWidth, f.CropRight),
			widget.NewLabel("Bottom"), ccWidget.NewIntSliderWithData(0, videoHeight, f.CropBottom), ccWidget.NewIntEntryWithData(0, videoHeight, f.CropBottom),
		),
	)
	return f
}

func (f Form) OnChange(fn func()) {
	l := binding.NewDataListener(fn)
	f.Frame.AddListener(l)
	f.Grow.AddListener(l)

	f.HueMin.AddListener(l)
	f.HueMax.AddListener(l)
	f.SatMin.AddListener(l)
	f.SatMax.AddListener(l)
	f.ValMin.AddListener(l)
	f.ValMax.AddListener(l)

	f.CropLeft.AddListener(l)
	f.CropTop.AddListener(l)
	f.CropRight.AddListener(l)
	f.CropBottom.AddListener(l)
}

func (f Form) Settings() (settings.Mask, error) {
	frame, err := f.Frame.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting frame: %v", err)
	}
	mode, err := f.Mode.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting mask mode: %v", err)
	}

	hueMin, err := f.HueMin.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting hue min: %v", err)
	}
	hueMax, err := f.HueMax.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting hue max: %v", err)
	}
	satMin, err := f.SatMin.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting sat min: %v", err)
	}
	satMax, err := f.SatMax.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting sat max: %v", err)
	}
	valMin, err := f.ValMin.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting val min: %v", err)
	}
	valMax, err := f.ValMax.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting val max: %v", err)
	}

	grow, err := f.Grow.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting grow: %v", err)
	}
	cropLeft, err := f.CropLeft.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting cropLeft: %v", err)
	}
	cropTop, err := f.CropTop.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting cropTop: %v", err)
	}
	cropRight, err := f.CropRight.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting cropRight: %v", err)
	}
	cropBottom, err := f.CropBottom.Get()
	if err != nil {
		return settings.Mask{}, fmt.Errorf("getting cropBottom: %v", err)
	}
	return settings.Mask{
		Frame:      frame,
		Mode:       mode,
		HueMin:     hueMin,
		HueMax:     hueMax,
		SatMin:     satMin,
		SatMax:     satMax,
		ValMin:     valMin,
		ValMax:     valMax,
		Grow:       grow,
		CropLeft:   cropLeft,
		CropTop:    cropTop,
		CropRight:  cropRight,
		CropBottom: cropBottom,
	}, nil
}
