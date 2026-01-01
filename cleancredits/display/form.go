package display

import (
	"fmt"
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/settings"
	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

const (
	ViewMask     = "Areas to inpaint"
	ViewDraw     = "Overrides"
	ViewPreview  = "Preview"
	ViewOriginal = "Original"
)

const ZoomFit = "Fit"

var ZoomLevelMap = map[string]float64{
	ZoomFit: 0,
	"10%":   .10,
	"25%":   .25,
	"50%":   .5,
	"100%":  1,
	"150%":  1.5,
	"200%":  2,
	"300%":  3,
	"400%":  4,
	"500%":  5,
}

var ZoomLevels = []string{
	ZoomFit,
	"10%",
	"25%",
	"50%",
	"100%",
	"150%",
	"200%",
	"300%",
	"400%",
	"500%",
}

type Form struct {
	Container *fyne.Container

	Mode    binding.String
	Zoom    binding.String
	AnchorX binding.Int
	AnchorY binding.Int
}

func NewForm(width, height int) Form {
	f := Form{
		Mode:    binding.NewString(),
		Zoom:    binding.NewString(),
		AnchorX: binding.NewInt(),
		AnchorY: binding.NewInt(),
	}
	err := f.Mode.Set(ViewMask)
	if err != nil {
		fmt.Println("Error setting mode: ", err)
	}
	err = f.Zoom.Set(ZoomFit)
	if err != nil {
		fmt.Println("Error setting zoom: ", err)
	}
	err = f.AnchorX.Set(width / 2)
	if err != nil {
		fmt.Println("Error setting anchorX: ", err)
	}
	err = f.AnchorY.Set(height / 2)
	if err != nil {
		fmt.Println("Error setting anchorY: ", err)
	}
	anchorXEntry := ccWidget.NewIntEntryWithData(0, width, f.AnchorX)
	anchorYEntry := ccWidget.NewIntEntryWithData(0, height, f.AnchorY)
	f.Container =
		container.New(
			layout.NewHBoxLayout(),
			widget.NewLabel("View"),
			widget.NewSelectWithData(
				[]string{
					ViewMask,
					ViewDraw,
					ViewPreview,
					ViewOriginal,
				},
				f.Mode),
			widget.NewLabel("Zoom"),
			widget.NewSelectWithData(
				ZoomLevels, f.Zoom,
			),
			widget.NewButtonWithIcon("", theme.ZoomInIcon(), f.ZoomIn),
			widget.NewButtonWithIcon("", theme.ZoomOutIcon(), f.ZoomOut),
			widget.NewLabel("Anchor X"),
			anchorXEntry,
			widget.NewLabel("Y"),
			anchorYEntry,
		)
	return f
}

func (f Form) OnChange(fn func()) {
	l := binding.NewDataListener(fn)
	f.Mode.AddListener(l)
	f.Zoom.AddListener(l)
	f.AnchorX.AddListener(l)
	f.AnchorY.AddListener(l)
}

func (f Form) Settings() (settings.Display, error) {
	mode, err := f.Mode.Get()
	if err != nil {
		return settings.Display{}, fmt.Errorf("getting mode: %v", err)
	}

	zoom, err := f.Zoom.Get()
	if err != nil {
		return settings.Display{}, fmt.Errorf("getting zoom: %v", err)
	}

	anchorX, err := f.AnchorX.Get()
	if err != nil {
		return settings.Display{}, fmt.Errorf("getting anchorX: %v", err)
	}

	anchorY, err := f.AnchorY.Get()
	if err != nil {
		return settings.Display{}, fmt.Errorf("getting anchorY: %v", err)
	}
	return settings.Display{
		Mode:    mode,
		Zoom:    zoom,
		AnchorX: anchorX,
		AnchorY: anchorY,
	}, nil
}

func (f Form) ZoomIn() {
	z, err := f.Zoom.Get()
	if err != nil {
		fmt.Println("Error getting Zoom: ", err)
	}
	if z == ZoomFit {
		fmt.Println("Zoom in/out from Fit not supported")
		return
	}
	i := slices.Index(ZoomLevels, z)
	if i < len(ZoomLevels)-1 {
		err = f.Zoom.Set(ZoomLevels[i+1])
		if err != nil {
			fmt.Println("Error setting Zoom: ", err)
		}
	}
}

func (f Form) ZoomOut() {
	z, err := f.Zoom.Get()
	if err != nil {
		fmt.Println("Error getting Zoom: ", err)
	}
	if z == ZoomFit {
		fmt.Println("Zoom in/out from Fit not supported")
		return
	}
	i := slices.Index(ZoomLevels, z)
	if i > 1 {
		err = f.Zoom.Set(ZoomLevels[i-1])
		if err != nil {
			fmt.Println("Error setting Zoom: ", err)
		}
	}
}
