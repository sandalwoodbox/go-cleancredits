package pipeline

import (
	"fmt"
	"image"
	"strconv"

	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/display"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/draw"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
)

type Pipeline struct {
	VideoCapture *gocv.VideoCapture

	// Cached images
	MaskFrame         gocv.Mat
	Mask              gocv.Mat
	MaskWithInput     image.Image
	MaskWithOverrides image.Image
	Inpainted         image.Image
	Preview           image.Image
	Zoomed            image.Image

	// Last rendered settings
	MaskSettings    mask.Settings
	DrawSettings    draw.Settings
	DisplaySettings display.Settings
}

func NewPipeline(vc *gocv.VideoCapture) Pipeline {
	return Pipeline{
		VideoCapture: vc,
	}
}

func (p Pipeline) UpdateMask(maskSettings mask.Settings, drawSettings draw.Settings, displaySettings display.Settings) {
	if maskSettings.Frame != p.MaskSettings.Frame {
		mat, err := LoadFrame(p.VideoCapture, maskSettings.Frame)
		if err != nil {
			fmt.Printf("Error loading frame %d/%s: %v\n",
				maskSettings.Frame,
				strconv.FormatFloat(p.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
				err)
			return
		}
		p.MaskFrame = mat
		p.MaskSettings.Frame = maskSettings.Frame
	}

	if p.maskSettingsChanged(maskSettings) {
		p.Mask = RenderMask(p.MaskFrame, maskSettings)
	}
	// img, err := mat.ToImage()
	// if err != nil {
	// 	fmt.Println("Error converting mat to image: ", err)
	// 	return
	// }
}

func (p Pipeline) ApplyMask(frame int) image.Image {
	return EmptyImage()
}

func (p Pipeline) maskSettingsChanged(ms mask.Settings) bool {
	switch {
	case ms.HueMin != p.MaskSettings.HueMin,
		ms.HueMax != p.MaskSettings.HueMax,
		ms.SatMin != p.MaskSettings.SatMin,
		ms.SatMax != p.MaskSettings.SatMax,
		ms.ValMin != p.MaskSettings.ValMin,
		ms.ValMax != p.MaskSettings.ValMax,
		ms.Grow != p.MaskSettings.Grow,
		ms.CropLeft != p.MaskSettings.CropLeft,
		ms.CropTop != p.MaskSettings.CropTop,
		ms.CropRight != p.MaskSettings.CropRight,
		ms.CropBottom != p.MaskSettings.CropBottom:
		return true
	}
	return false
}
