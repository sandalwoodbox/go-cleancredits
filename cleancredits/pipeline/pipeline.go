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
	DisplayFrame      gocv.Mat
	MaskFrame         gocv.Mat
	Mask              gocv.Mat
	MaskWithInput     gocv.Mat
	MaskWithOverrides gocv.Mat
	Inpainted         gocv.Mat
	Display           gocv.Mat
	Zoomed            gocv.Mat

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
		VideoCapture:      vc,
		VideoWidth:        w,
		VideoHeight:       h,
		DisplayFrame:      gocv.NewMat(),
		MaskFrame:         gocv.NewMat(),
		Mask:              gocv.NewMat(),
		MaskWithInput:     gocv.NewMat(),
		MaskWithOverrides: gocv.NewMat(),
		Inpainted:         gocv.NewMat(),
		Display:           gocv.NewMat(),
		Zoomed:            gocv.NewMat(),
	}
}

func (p Pipeline) UpdateMask(ms mask.Settings, drawSettings draw.Settings) error {
	if p.MaskFrame.Closed() {
		return fmt.Errorf("MaskFrame is closed")
	}
	maskFrameChanged := ms.Frame != p.MaskSettings.Frame || p.MaskFrame.Empty()
	fmt.Printf("maskFrameChanged: %t, %t\n", ms.Frame != p.MaskSettings.Frame, p.MaskFrame.Empty())
	if maskFrameChanged {
		p.MaskFrame.Close()
		mat := gocv.NewMat()
		err := LoadFrame(p.VideoCapture, ms.Frame, &mat)
		if err != nil {
			return fmt.Errorf("loading frame %d/%s: %v\n",
				ms.Frame,
				strconv.FormatFloat(p.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
				err)
		}
		fmt.Printf("Loaded mask frame: cols %d, rows %d, empty %t\n", mat.Cols(), mat.Rows(), mat.Empty())
		p.MaskFrame = mat
		p.MaskSettings.Frame = ms.Frame
	}

	if p.Mask.Closed() {
		return fmt.Errorf("Mask is closed")
	}
	maskSettingsChanged := maskFrameChanged || p.maskSettingsChanged(ms) || p.Mask.Empty()
	fmt.Printf("maskSettingsChanged: %t, %t, %t\n", maskFrameChanged, p.maskSettingsChanged(ms), p.Mask.Empty())
	fmt.Printf("Old: %v\nNew: %v\n", p.MaskSettings, ms)
	if maskSettingsChanged {
		p.Mask.Close()
		p.Mask = gocv.NewMat()
		RenderMask(p.MaskFrame, &p.Mask, ms)
		p.MaskSettings = ms
		p.MaskChanged = true
	}

	if p.MaskWithInput.Closed() {
		return fmt.Errorf("MaskWithInput is closed")
	}
	p.MaskWithInput.Close()
	// TODO: Take layers into account
	p.MaskWithInput = gocv.NewMat()
	p.Mask.CopyTo(&p.MaskWithInput)

	if p.MaskWithOverrides.Closed() {
		return fmt.Errorf("MaskWithOverrides is closed")
	}
	p.MaskWithOverrides.Close()
	// TODO: Take overrides (drawn) into account
	p.MaskWithOverrides = gocv.NewMat()
	p.MaskWithInput.CopyTo(&p.MaskWithOverrides)
	return nil
}

func (p Pipeline) ApplyMask(frame int, ds display.Settings, dst *gocv.Mat) error {
	if p.DisplayFrame.Closed() {
		return fmt.Errorf("DisplayFrame is closed")
	}
	frameChanged := frame != p.DisplayFrameNumber || p.DisplayFrame.Empty()
	fmt.Printf("frameChanged: %t, %t\n", frame != p.DisplayFrameNumber, p.DisplayFrame.Empty())
	if frameChanged {
		LoadFrame(p.VideoCapture, frame, &p.DisplayFrame)
		p.DisplayFrameNumber = frame
	}

	if p.Display.Closed() {
		return fmt.Errorf("Display is closed")
	}
	modeChanged := frameChanged || p.DisplaySettings.Mode != ds.Mode || p.MaskChanged || p.Display.Empty()
	fmt.Printf("modeChanged: %t, %t, %t, %t\n", frameChanged, p.DisplaySettings.Mode != ds.Mode, p.MaskChanged, p.Display.Empty())
	if modeChanged {
		switch ds.Mode {
		case display.ViewOriginal:
			gocv.CvtColor(p.DisplayFrame, &p.Display, gocv.ColorBGRToRGBA)
		case display.ViewMask:
			gocv.BitwiseAnd(p.DisplayFrame, p.MaskWithOverrides, &p.Display)
			gocv.CvtColor(p.Display, &p.Display, gocv.ColorBGRToRGBA)
		case display.ViewDraw:
			// TODO: Display draw layer
			gocv.CvtColor(p.DisplayFrame, &p.Display, gocv.ColorBGRToRGBA)
		default: // display.ViewPreview
			gocv.CvtColor(p.DisplayFrame, &p.Display, gocv.ColorBGRToRGB)
			// TODO: allow setting inpaint radius
			gocv.Inpaint(p.Display, p.MaskWithOverrides, &p.Display, 3, gocv.Telea)
		}
		p.MaskChanged = false
	}

	if p.Zoomed.Closed() {
		return fmt.Errorf("Zoomed is closed")
	}
	zoomChanged := modeChanged || p.zoomChanged(ds) || p.Zoomed.Empty()
	fmt.Printf("zoomChanged: %t, %t, %t\n", modeChanged, p.zoomChanged(ds), p.Zoomed.Empty())
	if zoomChanged {
		zf := display.ZoomLevelMap[ds.Zoom]
		r := ZoomCropRectangle(zf, ds.AnchorX, ds.AnchorY, p.VideoWidth, p.VideoHeight, 720, 480)
		fmt.Printf("Zoomed rectangle: %v\n", r)
		fmt.Printf("Display cols: %d rows: %d\n", p.Display.Cols(), p.Display.Rows())
		if r.Min.X < 0 || r.Size().X < 0 || r.Min.X+r.Size().X > p.Display.Cols() || r.Min.Y < 0 || r.Size().Y < 0 || r.Min.Y+r.Size().Y > p.Display.Rows() {
			return fmt.Errorf("Zoomed rectangle out of bounds: %v\n", r)
		}
		gocv.Resize(p.Display.Region(r), &p.Zoomed, image.Point{}, zf, zf, gocv.InterpolationNearestNeighbor)
	}
	p.DisplaySettings = ds
	return nil
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
