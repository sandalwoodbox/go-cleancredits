package pipeline

import (
	"fmt"
	"image"
	"math"
	"strconv"

	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/display"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/settings"
)

type Pipeline struct {
	VideoCapture  *gocv.VideoCapture
	VideoWidth    int
	VideoHeight   int
	DisplayWidth  int
	DisplayHeight int

	// Cached images
	DisplayFrame      *image.Image
	MaskFrame         *image.Image
	Mask              *image.Image
	MaskWithInput     *image.Image
	MaskWithOverrides *image.Image
	Inpainted         *image.Image
	Display           *image.Image
	Zoomed            *image.Image

	// Last rendered settings
	DisplayFrameNumber int
	MaskSettings       settings.Mask
	DrawSettings       settings.Draw
	DisplaySettings    settings.Display
	RenderSettings     settings.Render

	// Partial render status
	MaskChanged bool
}

func NewPipeline(vc *gocv.VideoCapture, displayWidth, displayHeight int) Pipeline {
	w := int(vc.Get(gocv.VideoCaptureFrameWidth))
	h := int(vc.Get(gocv.VideoCaptureFrameHeight))
	return Pipeline{
		VideoCapture:  vc,
		VideoWidth:    w,
		VideoHeight:   h,
		DisplayWidth:  displayWidth,
		DisplayHeight: displayHeight,
	}
}

func (p *Pipeline) UpdateMask(ms settings.Mask, drawSettings settings.Draw) error {
	maskFrameChanged := ms.Frame != p.MaskSettings.Frame || p.MaskFrame == nil

	var maskFrameMat gocv.Mat
	defer maskFrameMat.Close()
	var err error
	if maskFrameChanged {
		maskFrameMat = gocv.NewMat()
		err := LoadFrame(p.VideoCapture, ms.Frame, &maskFrameMat)
		if err != nil {
			return fmt.Errorf("loading frame %d/%s: %v\n",
				ms.Frame,
				strconv.FormatFloat(p.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
				err)
		}
		i, err := maskFrameMat.ToImage()
		if err != nil {
			return fmt.Errorf("converting maskFrame to image: %v", err)
		}
		p.MaskFrame = &i
		p.MaskSettings.Frame = ms.Frame
	} else {
		maskFrameMat, err = gocv.ImageToMatRGB(*p.MaskFrame)
		if err != nil {
			return fmt.Errorf("converting p.MaskFrame to mat: %v", err)
		}
	}

	maskSettingsChanged := maskFrameChanged || p.maskSettingsChanged(ms) || p.Mask == nil
	var maskMat gocv.Mat
	defer maskMat.Close()
	if maskSettingsChanged {
		maskMat = gocv.NewMat()
		RenderMask(maskFrameMat, &maskMat, ms)
		i, err := maskMat.ToImage()
		if err != nil {
			return fmt.Errorf("converting mask to image: %v", err)
		}
		p.Mask = &i
		p.MaskSettings = ms
		p.MaskChanged = true
	} else {
		maskMat, err = ImageToMatGray(*p.Mask)
		if err != nil {
			return fmt.Errorf("converting p.Mask to mat: %v", err)
		}
	}

	// TODO: Take layers into account
	p.MaskWithInput = p.Mask

	// TODO: Take overrides (drawn) into account
	p.MaskWithOverrides = p.MaskWithInput
	return nil
}

func (p *Pipeline) ApplyMask(frame int, ds settings.Display, rs settings.Render) (image.Image, error) {
	frameChanged := frame != p.DisplayFrameNumber || p.DisplayFrame == nil
	var displayFrameMat gocv.Mat
	defer displayFrameMat.Close()
	var err error
	if frameChanged {
		displayFrameMat = gocv.NewMat()
		err = LoadFrame(p.VideoCapture, frame, &displayFrameMat)
		if err != nil {
			return nil, fmt.Errorf("loading frame: %v", err)
		}
		i, err := displayFrameMat.ToImage()
		if err != nil {
			return nil, fmt.Errorf("converting displayFrame to image: %v", err)
		}
		p.DisplayFrame = &i
		p.DisplayFrameNumber = frame
	} else {
		displayFrameMat, err = gocv.ImageToMatRGB(*p.DisplayFrame)
		if err != nil {
			return nil, fmt.Errorf("converting p.DisplayFrame to mat: %v", err)
		}
	}

	modeChanged := frameChanged || p.DisplaySettings.Mode != ds.Mode || p.MaskChanged || p.Display == nil || (ds.Mode == display.ViewPreview && rs.InpaintRadius != p.RenderSettings.InpaintRadius)
	var displayMat gocv.Mat
	defer displayMat.Close()
	if modeChanged {
		switch ds.Mode {
		case display.ViewOriginal:
			displayMat = displayFrameMat.Clone()
		case display.ViewMask:
			displayMat = gocv.NewMat()
			m, err := ImageToMatGray(*p.MaskWithOverrides)
			defer m.Close()
			if err != nil {
				return nil, fmt.Errorf("converting p.MaskWithOverrides to mat: %v", err)
			}
			gocv.BitwiseAndWithMask(displayFrameMat, displayFrameMat, &displayMat, m)
		case display.ViewDraw:
			// TODO: Display draw layer
			displayMat = displayFrameMat.Clone()
		default: // display.ViewPreview
			displayMat = gocv.NewMat()
			m, err := ImageToMatGray(*p.MaskWithOverrides)
			defer m.Close()
			if err != nil {
				return nil, fmt.Errorf("converting p.MaskWithOverrides to mat: %v", err)
			}
			gocv.Inpaint(displayFrameMat, m, &displayMat, float32(p.RenderSettings.InpaintRadius), gocv.Telea)
		}
		i, err := displayMat.ToImage()
		if err != nil {
			return nil, fmt.Errorf("converting display to image: %v", err)
		}
		p.Display = &i
		p.MaskChanged = false
	} else {
		displayMat, err = gocv.ImageToMatRGB(*p.Display)
		if err != nil {
			return nil, fmt.Errorf("converting p.Display to mat: %v", err)
		}
	}

	zoomChanged := modeChanged || p.zoomChanged(ds) || p.Zoomed == nil
	var zoomed image.Image
	if zoomChanged {
		zf := display.ZoomLevelMap[ds.Zoom]
		if zf == 0 {
			zf = math.Min(float64(p.DisplayWidth)/float64(p.VideoWidth), float64(p.DisplayHeight)/float64(p.VideoHeight))
		}
		r := ZoomCropRectangle(zf, ds.AnchorX, ds.AnchorY, p.VideoWidth, p.VideoHeight, p.DisplayWidth, p.DisplayHeight)
		m := gocv.NewMatWithSize(p.DisplayHeight, p.DisplayWidth, gocv.MatTypeCV8UC3)
		gocv.Resize(displayMat.Region(r), &m, image.Point{}, zf, zf, gocv.InterpolationNearestNeighbor)
		zoomed, err = m.ToImage()
		if err != nil {
			return nil, fmt.Errorf("converting display to image: %v", err)
		}
		p.Zoomed = &zoomed
	} else {
		zoomed = *p.Zoomed
	}
	p.DisplaySettings = ds
	return zoomed, nil
}

func (p Pipeline) maskSettingsChanged(ms settings.Mask) bool {
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

func (p Pipeline) zoomChanged(ds settings.Display) bool {
	switch {
	case ds.Zoom != p.DisplaySettings.Zoom,
		ds.AnchorX != p.DisplaySettings.AnchorX,
		ds.AnchorY != p.DisplaySettings.AnchorY:
		return true
	}
	return false
}
