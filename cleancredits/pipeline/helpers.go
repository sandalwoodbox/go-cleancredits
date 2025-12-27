package pipeline

import (
	"fmt"
	"image"
	"math"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"gocv.io/x/gocv"
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

func RenderMask(mat gocv.Mat, dst *gocv.Mat, s mask.Settings) {
	if s.CropBottom > s.CropTop {
		s.CropBottom, s.CropTop = s.CropTop, s.CropBottom
	}
	if s.CropLeft > s.CropRight {
		s.CropLeft, s.CropRight = s.CropRight, s.CropLeft
	}
	s.CropBottom = int(math.Min(math.Max(float64(s.CropBottom), 0), float64(mat.Rows())))
	s.CropTop = int(math.Min(math.Max(float64(s.CropTop), 0), float64(mat.Rows())))
	s.CropLeft = int(math.Min(math.Max(float64(s.CropLeft), 0), float64(mat.Cols())))
	s.CropRight = int(math.Min(math.Max(float64(s.CropRight), 0), float64(mat.Cols())))

	fmt.Printf("mat dims: %d x %d, %d\n", mat.Cols(), mat.Rows(), mat.Channels())
	frameHSV := gocv.NewMat()
	defer frameHSV.Close()
	gocv.CvtColor(mat, &frameHSV, gocv.ColorBGRToHSV)
	fmt.Printf("frameHSV dims: %d x %d, %d\n", frameHSV.Cols(), frameHSV.Rows(), frameHSV.Channels())

	hsvMask := gocv.NewMat()
	defer hsvMask.Close()
	gocv.InRangeWithScalar(
		frameHSV,
		gocv.NewScalar(float64(s.HueMin), float64(s.SatMin), float64(s.ValMin), 0),
		gocv.NewScalar(float64(s.HueMax), float64(s.SatMax), float64(s.ValMax), 0),
		&hsvMask,
	)
	fmt.Printf("hsvMask dims: %d x %d, %d\n", hsvMask.Cols(), hsvMask.Rows(), hsvMask.Channels())

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
	fmt.Printf("grown dims: %d x %d, %d\n", grown.Cols(), grown.Rows(), grown.Channels())

	bboxMask := gocv.Zeros(hsvMask.Rows(), hsvMask.Cols(), gocv.MatTypeCV8U)
	defer bboxMask.Close()

	cropArea := gocv.Ones(s.CropBottom-s.CropTop, s.CropRight-s.CropLeft, gocv.MatTypeCV8U)
	defer cropArea.Close()

	rect := image.Rect(s.CropLeft, s.CropTop, s.CropRight, s.CropBottom)
	r := bboxMask.Region(rect)
	gocv.Add(r, cropArea, &r)

	fmt.Printf("bboxMask dims: %d x %d, %d\n", bboxMask.Cols(), bboxMask.Rows(), bboxMask.Channels())
	gocv.BitwiseAnd(grown, bboxMask, dst)

	fmt.Printf("dst dims: %d x %d, %d\n", dst.Cols(), dst.Rows(), dst.Channels())
}
