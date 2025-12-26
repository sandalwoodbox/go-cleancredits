package pipeline

import (
	"fmt"
	"image"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"gocv.io/x/gocv"
)

func EmptyImage() image.Image {
	return image.NewRGBA(image.Rect(0, 0, 1920, 1080))
}

func LoadFrame(vc *gocv.VideoCapture, n int) (gocv.Mat, error) {
	mat := gocv.NewMat()
	defer mat.Close()
	vc.Set(
		gocv.VideoCapturePosFrames,
		float64(n),
	)
	ok := vc.Read(&mat)
	if !ok {
		return mat, fmt.Errorf("invalid frame number: %d", n)
	}
	return mat, nil
}

func RenderMask(mat gocv.Mat, s mask.Settings) gocv.Mat {
	frameHSV := gocv.NewMat()
	defer frameHSV.Close()
	gocv.CvtColor(mat, &frameHSV, gocv.ColorBGRToHSV)

	hsvMask := gocv.NewMat()
	defer hsvMask.Close()
	gocv.InRangeWithScalar(
		frameHSV,
		gocv.NewScalar(float64(s.HueMin), float64(s.SatMin), float64(s.ValMin), 0),
		gocv.NewScalar(float64(s.HueMax), float64(s.SatMax), float64(s.ValMax), 0),
		&hsvMask,
	)

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

	bboxMask := gocv.Zeros(hsvMask.Rows(), hsvMask.Cols(), gocv.MatTypeCV8U)
	defer bboxMask.Close()
	merged := gocv.NewMat()
	defer merged.Close()
	gocv.Merge([]gocv.Mat{bboxMask, bboxMask, bboxMask}, &bboxMask)
	mask := gocv.NewMat()
	gocv.BitwiseAnd(grown, bboxMask, &mask)
	return mask
}
