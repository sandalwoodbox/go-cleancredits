package pipeline

import (
	"fmt"
	"path"
	"reflect"
	"strings"
	"testing"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
)

func TestRenderMask(t *testing.T) {
	cases := []struct {
		name                                     string
		hueMin, hueMax                           int
		satMin, satMax                           int
		valMin, valMax                           int
		grow                                     int
		cropLeft, cropRight, cropTop, cropBottom int
		inputMask                                string
		want                                     string
	}{
		{
			name:       "keep",
			hueMin:     0,
			hueMax:     179,
			satMin:     0,
			satMax:     255,
			valMin:     0,
			valMax:     255,
			grow:       0,
			cropLeft:   0,
			cropRight:  1080,
			cropTop:    0,
			cropBottom: 720,
			want:       "keep.png",
		},
		{
			name:       "mask",
			hueMin:     10,
			hueMax:     20,
			satMin:     30,
			satMax:     50,
			valMin:     40,
			valMax:     60,
			grow:       1,
			cropLeft:   100,
			cropRight:  500,
			cropTop:    150,
			cropBottom: 700,
			want:       "mask.png",
		},
		{
			name:       "out of bounds",
			hueMin:     -5,
			hueMax:     200,
			satMin:     -10,
			satMax:     290,
			valMin:     -50,
			valMax:     365,
			grow:       500,
			cropLeft:   2000,
			cropRight:  2000,
			cropTop:    9001,
			cropBottom: 9001,
			want:       "out_of_bounds.png",
		},
		{
			name:       "inverted out of bounds",
			hueMin:     200,
			hueMax:     -5,
			satMin:     290,
			satMax:     -10,
			valMin:     365,
			valMax:     -50,
			grow:       -1,
			cropLeft:   -20,
			cropRight:  -21,
			cropTop:    -200,
			cropBottom: -2000,
			want:       "inverted_out_of_bounds.png",
		},
		{
			name:       "higher grow",
			hueMin:     30,
			hueMax:     60,
			satMin:     20,
			satMax:     200,
			valMin:     0,
			valMax:     240,
			grow:       3,
			cropLeft:   20,
			cropRight:  900,
			cropTop:    35,
			cropBottom: 500,
			want:       "higher_grow.png",
		},
		{
			name:       "no hue change",
			hueMin:     0,
			hueMax:     179,
			satMin:     0,
			satMax:     45,
			valMin:     221,
			valMax:     255,
			grow:       2,
			cropLeft:   302,
			cropRight:  762,
			cropTop:    593,
			cropBottom: 678,
			want:       "no_hue_change.png",
		},
		{
			name:       "unusual numbers",
			hueMin:     21,
			hueMax:     87,
			satMin:     11,
			satMax:     143,
			valMin:     49,
			valMax:     109,
			grow:       0,
			cropLeft:   588,
			cropRight:  985,
			cropTop:    455,
			cropBottom: 709,
			want:       "unusual_numbers.png",
		},
		{
			name:       "input mask",
			hueMin:     -5,
			hueMax:     200,
			satMin:     -10,
			satMax:     290,
			valMin:     -50,
			valMax:     365,
			grow:       500,
			cropLeft:   2000,
			cropRight:  2000,
			cropTop:    9001,
			cropBottom: 9001,
			inputMask:  "horses-720p-mask",
			want:       "input_mask.png",
		},
	}

	vc, err := gocv.VideoCaptureFile("testdata/horses-720p.mp4")
	if err != nil {
		t.Fatalf("Error loading video file: %v", err)
	}
	input := gocv.NewMat()
	defer input.Close()
	LoadFrame(vc, 0, &input)

	hashes := []contrib.ImgHashBase{
		contrib.PHash{},
		contrib.AverageHash{},
		contrib.BlockMeanHash{},
		contrib.BlockMeanHash{Mode: contrib.BlockMeanHashMode1},
		contrib.ColorMomentHash{},
		contrib.NewMarrHildrethHash(),
		contrib.NewRadialVarianceHash(),
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel()
			ms := mask.Settings{
				HueMin:     tc.hueMin,
				HueMax:     tc.hueMax,
				SatMin:     tc.satMin,
				SatMax:     tc.satMax,
				ValMin:     tc.valMin,
				ValMax:     tc.valMax,
				Grow:       tc.grow,
				CropLeft:   tc.cropLeft,
				CropRight:  tc.cropRight,
				CropTop:    tc.cropTop,
				CropBottom: tc.cropBottom,
			}
			got := gocv.NewMat()
			defer got.Close()
			RenderMask(input, &got, ms)
			want := gocv.IMRead(
				path.Join("testdata/render_mask_output", tc.want),
				gocv.IMReadGrayScale,
			)
			if got.Cols() != want.Cols() || got.Rows() != want.Rows() {
				t.Fatalf("dimensions not equal. got %d x %d, want %d x %d", got.Cols(), got.Rows(), want.Cols(), want.Rows())
			}
			if got.Channels() != want.Channels() {
				t.Fatalf("number of channels not equal. got %d, want %d", got.Channels(), want.Channels())
			}
			if got.Type() != want.Type() {
				t.Fatalf("type not equal. got %s, want %s", got.Type().String(), want.Type().String())
			}
			if got.ElemSize() != want.ElemSize() {
				t.Fatalf("ElemSize not equal. got %d, want %d", got.ElemSize(), want.ElemSize())
			}
			if !reflect.DeepEqual(got.GetUCharAt(0, 0), want.GetUCharAt(0, 0)) {
				t.Fatalf("First pixel not equal. got %v, want %v", got.GetUCharAt(0, 0), want.GetUCharAt(0, 0))
			}

			gotImg, err := got.ToImage()
			if err != nil {
				t.Errorf("Error converting got Mat to img: %v", err)
			}
			wantImg, err := want.ToImage()
			if err != nil {
				t.Errorf("Error converting want Mat to img: %v", err)
			}
			if !reflect.DeepEqual(gotImg, wantImg) {
				t.Errorf("rendered mask not equal to expected mask. Similarity:\n")
				for _, hash := range hashes {
					name := strings.TrimPrefix(fmt.Sprintf("%T", hash), "contrib.")
					gotHash := gocv.NewMat()
					wantHash := gocv.NewMat()
					hash.Compute(got, &gotHash)
					hash.Compute(want, &wantHash)
					if gotHash.Empty() {
						t.Errorf("error computing %s got hash", name)
					}
					if wantHash.Empty() {
						t.Errorf("error computing %s want hash", name)
					}
					similarity := hash.Compare(gotHash, wantHash)
					gotHash.Close()
					wantHash.Close()
					t.Errorf("%s: similarity %g\n", name, similarity)
				}

				gw := gocv.NewWindow(fmt.Sprintf("%s got", tc.name))
				defer gw.Close()
				gw.ResizeWindow(1920, 1080)
				gw.IMShow(got)
				gw.WaitKey(1)
				ww := gocv.NewWindow(fmt.Sprintf("%s want", tc.name))
				defer ww.Close()
				ww.ResizeWindow(1920, 1080)
				ww.IMShow(want)
				ww.WaitKey(30000)
			}
		})
	}
}
