package pipeline

import (
	"fmt"
	"image"
	"math"

	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/settings"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/utils"
)

func EmptyImage() image.Image {
	return image.NewRGBA(image.Rect(0, 0, 1920, 1080))
}

func LoadFrame(vc *gocv.VideoCapture, n int, dst *gocv.Mat) error {
	vc.Set(
		gocv.VideoCapturePosFrames,
		float64(n),
	)
	ok := vc.Read(dst)
	if !ok {
		return fmt.Errorf("invalid frame number: %d", n)
	}
	return nil
}

func RenderMask(mat gocv.Mat, dst *gocv.Mat, s settings.Mask) {
	// wait := 1
	// w := gocv.NewWindow("RenderMask")
	// w.IMShow(mat)
	// w.SetWindowTitle("RenderMask - mat")
	// w.WaitKey(wait)
	if s.CropTop > s.CropBottom {
		s.CropTop, s.CropBottom = s.CropBottom, s.CropTop
	}
	if s.CropLeft > s.CropRight {
		s.CropLeft, s.CropRight = s.CropRight, s.CropLeft
	}
	s.CropBottom = utils.ClampInt(s.CropBottom, 0, mat.Rows())
	s.CropTop = utils.ClampInt(s.CropTop, 0, mat.Rows())
	s.CropLeft = utils.ClampInt(s.CropLeft, 0, mat.Cols())
	s.CropRight = utils.ClampInt(s.CropRight, 0, mat.Cols())

	// fmt.Printf("mat dims: %d x %d, %d\n", mat.Cols(), mat.Rows(), mat.Channels())
	frameHSV := gocv.NewMat()
	defer frameHSV.Close()
	gocv.CvtColor(mat, &frameHSV, gocv.ColorBGRToHSV)
	// fmt.Printf("frameHSV dims: %d x %d, %d\n", frameHSV.Cols(), frameHSV.Rows(), frameHSV.Channels())
	// w.IMShow(frameHSV)
	// w.SetWindowTitle("RenderMask - frameHSV")
	// w.WaitKey(wait)

	hsvMask := gocv.NewMat()
	defer hsvMask.Close()
	gocv.InRangeWithScalar(
		frameHSV,
		gocv.NewScalar(float64(s.HueMin), float64(s.SatMin), float64(s.ValMin), 0),
		gocv.NewScalar(float64(s.HueMax), float64(s.SatMax), float64(s.ValMax), 0),
		&hsvMask,
	)
	// fmt.Printf("hsvMask dims: %d x %d, %d\n", hsvMask.Cols(), hsvMask.Rows(), hsvMask.Channels())
	// w.IMShow(hsvMask)
	// w.SetWindowTitle("RenderMask - hsvMask")
	// w.WaitKey(wait)

	var grown gocv.Mat
	if s.Grow > 0 {
		grown = gocv.NewMat()
		kernel := gocv.Ones(s.Grow, s.Grow, gocv.MatTypeCV8U)
		defer kernel.Close()
		gocv.Dilate(hsvMask, &grown, kernel)
	} else {
		grown = hsvMask.Clone()
	}
	defer grown.Close()
	// fmt.Printf("grown dims: %d x %d, %d\n", grown.Cols(), grown.Rows(), grown.Channels())
	// w.IMShow(grown)
	// w.SetWindowTitle("RenderMask - grown")
	// w.WaitKey(wait)

	bboxMask := gocv.Zeros(hsvMask.Rows(), hsvMask.Cols(), gocv.MatTypeCV8U)
	defer bboxMask.Close()
	for x := s.CropLeft; x < s.CropRight; x++ {
		for y := s.CropTop; y < s.CropBottom; y++ {
			bboxMask.SetUCharAt(y, x, 255)
		}
	}
	// fmt.Printf("Crop: l%d, t%d, r%d, b%d\n", s.CropLeft, s.CropTop, s.CropRight, s.CropBottom)
	// fmt.Printf("bboxMask dims: %d x %d, %d\n", bboxMask.Cols(), bboxMask.Rows(), bboxMask.Channels())
	// w.IMShow(bboxMask)
	// w.SetWindowTitle("RenderMask - bboxMask")
	// w.WaitKey(wait)
	gocv.BitwiseAnd(grown, bboxMask, dst)
	gocv.BitwiseAndWithMask(grown, grown, dst, bboxMask)

	// fmt.Printf("dst dims: %d x %d, %d\n", dst.Cols(), dst.Rows(), dst.Channels())
	// w.IMShow(*dst)
	// w.SetWindowTitle("RenderMask - dst")
	// w.WaitKey(wait)
}

func CombineMasks(mode string, top gocv.Mat, bottom, dst *gocv.Mat) {
	if mode == mask.Include {
		if bottom == nil {
			top.CopyTo(dst)
			return
		}
		gocv.BitwiseOr(top, *bottom, dst)
		return
	}
	inv := gocv.NewMat()
	defer inv.Close()
	gocv.BitwiseNot(top, &inv)
	if bottom == nil {
		inv.CopyTo(dst)
		return
	}
	gocv.BitwiseAnd(inv, *bottom, dst)
}

func ZoomCropRectangle(zoomFactor float64, anchorX, anchorY, videoWidth, videoHeight, maxWidth, maxHeight int) image.Rectangle {
	// zoom width and height are the dimensions of the box in the original
	// image that will be zoomed in (or out) and shown to the user. This should
	// be the largest possible width/height that will fit in the display box
	// after the zoom is applied.
	zoomWidth := int(math.Min(float64(videoWidth)*zoomFactor, float64(maxWidth)) / zoomFactor)
	zoomHeight := int(math.Min(float64(videoHeight)*zoomFactor, float64(maxHeight)) / zoomFactor)

	// Crop x and y are set at half the zoom width away from the zoom center,
	// in order to center it. However, the zoom center may be near an edge, so
	// we need to clip it between 0 and the farthest right point possible that
	// won't spill over the video width. If the entire frame should be visible,
	// the crop x and y will always be 0.
	cropX := utils.ClampInt(anchorX-(zoomWidth/2), 0, int(math.Max(float64(videoWidth-zoomWidth), 0)))
	cropY := utils.ClampInt(
		anchorY-(zoomHeight/2),
		0,
		max(videoHeight-zoomHeight, 0),
	)
	return image.Rect(cropX, cropY, cropX+zoomWidth, cropY+zoomHeight)
}

func ImageToMatGray(i image.Image) (gocv.Mat, error) {
	mRGB, err := gocv.ImageToMatRGB(i)
	if err != nil {
		mRGB.Close()
		return gocv.NewMat(), fmt.Errorf("converting p.Mask to mat: %v", err)
	}
	defer mRGB.Close()
	mGray := gocv.NewMat()
	err = gocv.CvtColor(mRGB, &mGray, gocv.ColorBGRToGray)
	if err != nil {
		mGray.Close()
		return gocv.NewMat(), fmt.Errorf("converting maskMat to gray: %v", err)
	}
	return mGray, nil
}
