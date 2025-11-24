package display

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"gocv.io/x/gocv"
)

type Display struct {
	VideoCapture *gocv.VideoCapture
	Frame        *int
}

func NewDisplay(vc *gocv.VideoCapture) Display {
	d := Display{
		VideoCapture: vc,
		Frame:        new(int),
	}
	d.SetFrame(0)
	return d
}

func (d Display) SetFrame(frame int) {
	*d.Frame = frame
}

func (d Display) Render() *canvas.Image {
	mat := gocv.NewMat()
	defer mat.Close()

	d.VideoCapture.Set(gocv.VideoCapturePosFrames, float64(*d.Frame))
	d.VideoCapture.Read(&mat)
	frame, err := mat.ToImage()
	if err != nil {
		fmt.Println("Error loading frame: ", err)
	}
	fImg := canvas.NewImageFromImage(frame)
	fImg.FillMode = canvas.ImageFillContain
	fImg.SetMinSize(fyne.NewSize(720, 480))
	return fImg
}
