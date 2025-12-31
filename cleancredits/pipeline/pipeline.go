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
	VideoWidth   int
	VideoHeight  int

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
	MaskSettings       mask.Settings
	DrawSettings       draw.Settings
	DisplaySettings    display.Settings

	// Partial render status
	MaskChanged bool
}

func NewPipeline(vc *gocv.VideoCapture) Pipeline {
	w := int(vc.Get(gocv.VideoCaptureFrameWidth))
	h := int(vc.Get(gocv.VideoCaptureFrameHeight))
	return Pipeline{
		VideoCapture: vc,
		VideoWidth:   w,
		VideoHeight:  h,
	}
}

func (p *Pipeline) UpdateMask(ms mask.Settings, drawSettings draw.Settings) error {
	maskFrameChanged := ms.Frame != p.MaskSettings.Frame || p.MaskFrame == nil

	fmt.Printf("maskFrameChanged: %t, %t\n", ms.Frame != p.MaskSettings.Frame, p.MaskFrame == nil)
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
		fmt.Printf("Loaded mask frame: cols %d, rows %d, empty %t\n", maskFrameMat.Cols(), maskFrameMat.Rows(), maskFrameMat.Empty())
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
	fmt.Printf("maskSettingsChanged: %t, %t, %t\n", maskFrameChanged, p.maskSettingsChanged(ms), p.Mask == nil)
	fmt.Printf("Old: %v\nNew: %v\n", p.MaskSettings, ms)
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

func (p *Pipeline) ApplyMask(frame int, ds display.Settings) (image.Image, error) {
	frameChanged := frame != p.DisplayFrameNumber || p.DisplayFrame == nil
	fmt.Printf("frameChanged: %t, %t\n", frame != p.DisplayFrameNumber, p.DisplayFrame == nil)
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

	modeChanged := frameChanged || p.DisplaySettings.Mode != ds.Mode || p.MaskChanged || p.Display == nil
	fmt.Printf("modeChanged: %t, %t, %t, %t\n", frameChanged, p.DisplaySettings.Mode != ds.Mode, p.MaskChanged, p.Display == nil)
	var displayMat gocv.Mat
	defer displayMat.Close()
	if modeChanged {
		displayMat = gocv.NewMat()
		switch ds.Mode {
		case display.ViewOriginal:
			gocv.CvtColor(displayFrameMat, &displayMat, gocv.ColorBGRToRGBA)
		case display.ViewMask:
			m, err := ImageToMatGray(*p.MaskWithOverrides)
			defer m.Close()
			if err != nil {
				return nil, fmt.Errorf("converting p.MaskWithOverrides to mat: %v", err)
			}
			fmt.Printf("displayFrameMat continuous: %t\n", displayFrameMat.IsContinuous())
			fmt.Printf("m continuous: %t\n", m.IsContinuous())
			gocv.BitwiseAndWithMask(displayFrameMat, displayFrameMat, &displayMat, m)
			fmt.Printf("displayMat size: %dx%d\n", displayMat.Cols(), displayMat.Rows())
			fmt.Printf("displayMat continuous: %t\n", displayMat.IsContinuous())
			gocv.CvtColor(displayMat, &displayMat, gocv.ColorBGRToRGBA)
			fmt.Printf("displayMat continuous: %t\n", displayMat.IsContinuous())
		case display.ViewDraw:
			// TODO: Display draw layer
			gocv.CvtColor(displayFrameMat, &displayMat, gocv.ColorBGRToRGBA)
		default: // display.ViewPreview
			m, err := ImageToMatGray(*p.MaskWithOverrides)
			defer m.Close()
			if err != nil {
				return nil, fmt.Errorf("converting p.MaskWithOverrides to mat: %v", err)
			}
			gocv.CvtColor(displayFrameMat, &displayMat, gocv.ColorBGRToRGB)
			// TODO: allow setting inpaint radius
			gocv.Inpaint(displayMat, m, &displayMat, 3, gocv.Telea)
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
	fmt.Printf("zoomChanged: %t, %t, %t\n", modeChanged, p.zoomChanged(ds), p.Zoomed == nil)
	var zoomed image.Image
	if zoomChanged {
		zf := display.ZoomLevelMap[ds.Zoom]
		displayWidth := 720
		displayHeight := 480
		r := ZoomCropRectangle(zf, ds.AnchorX, ds.AnchorY, p.VideoWidth, p.VideoHeight, displayWidth, displayHeight)
		fmt.Printf("Zoomed rectangle: %v\n", r)
		if r.Min.X < 0 || r.Size().X < 0 || r.Min.X+r.Size().X > displayWidth || r.Min.Y < 0 || r.Size().Y < 0 || r.Min.Y+r.Size().Y > displayHeight {
			return nil, fmt.Errorf("Zoomed rectangle out of bounds: %v\n", r)
		}
		m := gocv.NewMatWithSize(displayHeight, displayWidth, gocv.MatTypeCV8UC3)
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

func (p Pipeline) zoomChanged(ds display.Settings) bool {
	switch {
	case ds.Zoom != p.DisplaySettings.Zoom,
		ds.AnchorX != p.DisplaySettings.AnchorX,
		ds.AnchorY != p.DisplaySettings.AnchorY:
		return true
	}
	return false
}
