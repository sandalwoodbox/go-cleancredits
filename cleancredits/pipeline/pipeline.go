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
	FrameCache    *FrameCache
	VideoWidth    int
	VideoHeight   int
	DisplayWidth  int
	DisplayHeight int

	// Cached images
	Mask              *image.Image
	MaskWithInput     *image.Image
	MaskWithOverrides *image.Image
	Display           gocv.Mat
	Zoomed            gocv.Mat

	// Last rendered settings
	DisplayFrameNumber int
	MaskSettings       settings.Mask
	DrawSettings       settings.Draw
	DisplaySettings    settings.Display
	RenderSettings     settings.Render

	// Partial render status
	MaskChanged bool
}

func NewPipeline(vc *gocv.VideoCapture, displayWidth, displayHeight int) (*Pipeline, error) {
	w := int(vc.Get(gocv.VideoCaptureFrameWidth))
	h := int(vc.Get(gocv.VideoCaptureFrameHeight))
	cache, err := NewFrameCache(vc, false)
	if err != nil {
		return nil, fmt.Errorf("creating frame cache: %v", err)
	}
	return &Pipeline{
		VideoCapture:       vc,
		FrameCache:         cache,
		VideoWidth:         w,
		VideoHeight:        h,
		DisplayWidth:       displayWidth,
		DisplayHeight:      displayHeight,
		DisplayFrameNumber: -1,
		Display:            gocv.NewMat(),
		Zoomed:             gocv.NewMat(),
		MaskSettings:       settings.Mask{Frame: -1},
		DrawSettings:       settings.Draw{Frame: -1},
		RenderSettings:     settings.Render{Frame: -1},
		DisplaySettings:    settings.Display{Zoom: "-1"},
	}, nil
}

func (p *Pipeline) UpdateMask(ms settings.Mask, drawSettings settings.Draw) error {
	maskFrameChanged := ms.Frame != p.MaskSettings.Frame
	maskFrameMat, err := p.FrameCache.LoadFrame(ms.Frame)
	if err != nil {
		return fmt.Errorf("loading frame %d/%s: %v\n",
			ms.Frame,
			strconv.FormatFloat(p.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
			err)
	}
	if maskFrameChanged {
		p.MaskSettings.Frame = ms.Frame
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
	frameChanged := frame != p.DisplayFrameNumber
	displayFrameMat, err := p.FrameCache.LoadFrame(frame)
	if err != nil {
		return nil, fmt.Errorf("loading frame %d/%s: %v\n",
			p.DisplayFrameNumber,
			strconv.FormatFloat(p.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
			err)
	}
	if frameChanged {
		p.DisplayFrameNumber = frame
	}

	modeChanged := frameChanged || p.DisplaySettings.Mode != ds.Mode || p.MaskChanged || (ds.Mode == display.ViewPreview && rs.InpaintRadius != p.RenderSettings.InpaintRadius)
	if modeChanged {
		p.Display.Close()
		switch ds.Mode {
		case display.ViewOriginal:
			p.Display = displayFrameMat.Clone()
		case display.ViewMask:
			mask, err := ImageToMatGray(*p.MaskWithOverrides)
			defer mask.Close()
			if err != nil {
				return nil, fmt.Errorf("converting p.MaskWithOverrides to mat: %v", err)
			}
			p.Display = gocv.NewMat()
			gocv.BitwiseAndWithMask(displayFrameMat, displayFrameMat, &p.Display, mask)
		case display.ViewDraw:
			// TODO: Display draw layer
			p.Display = displayFrameMat.Clone()
		default: // display.ViewPreview
			mask, err := ImageToMatGray(*p.MaskWithOverrides)
			defer mask.Close()
			if err != nil {
				return nil, fmt.Errorf("converting p.MaskWithOverrides to mat: %v", err)
			}
			p.Display = gocv.NewMat()
			gocv.Inpaint(displayFrameMat, mask, &p.Display, float32(rs.InpaintRadius), gocv.Telea)
			p.RenderSettings = rs
		}
		p.MaskChanged = false
	}

	zoomChanged := modeChanged || p.zoomChanged(ds)
	if zoomChanged {
		zf := display.ZoomLevelMap[ds.Zoom]
		if zf == 0 {
			zf = math.Min(float64(p.DisplayWidth)/float64(p.VideoWidth), float64(p.DisplayHeight)/float64(p.VideoHeight))
		}
		r := ZoomCropRectangle(zf, ds.AnchorX, ds.AnchorY, p.VideoWidth, p.VideoHeight, p.DisplayWidth, p.DisplayHeight)
		p.Zoomed = gocv.NewMatWithSize(p.DisplayHeight, p.DisplayWidth, gocv.MatTypeCV8UC3)
		gocv.Resize(p.Display.Region(r), &p.Zoomed, image.Point{}, zf, zf, gocv.InterpolationNearestNeighbor)
	}
	zoomed, err := p.Zoomed.ToImage()
	if err != nil {
		return nil, fmt.Errorf("converting zoomed to image: %v", err)
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
