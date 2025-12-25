package display

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/data/binding"
	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
)

type Display struct {
	VideoCapture *gocv.VideoCapture
	Image        *canvas.Image
	SelectedTab  binding.String
	MaskForm     mask.Form
}

func NewDisplay(vc *gocv.VideoCapture, selectedTab binding.String, maskForm mask.Form) Display {
	d := Display{
		VideoCapture: vc,
		Image:        &canvas.Image{},
		SelectedTab:  selectedTab,
		MaskForm:     maskForm,
	}
	selectedTab.AddListener(binding.NewDataListener(func() {
		d.Render()
	}))
	maskForm.OnChange(func() {
		d.Render()
	})
	return d
}

func (d Display) Render() {
	mat := gocv.NewMat()
	defer mat.Close()

	frame, err := d.MaskForm.Frame.Get()
	if err != nil {
		fmt.Println("Error getting frame number: ", err)
		return
	}
	d.VideoCapture.Set(
		gocv.VideoCapturePosFrames,
		frame,
	)
	d.VideoCapture.Read(&mat)
	img, err := mat.ToImage()
	if err != nil {
		fmt.Printf(
			"Error loading frame %s/%s: %v\n",
			strconv.FormatFloat(frame, 'f', -1, 64),
			strconv.FormatFloat(d.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
			err)
		return
	}
	d.Image.FillMode = canvas.ImageFillContain
	d.Image.SetMinSize(fyne.NewSize(720, 480))
	d.Image.Image = img
	d.Image.Refresh()
}
