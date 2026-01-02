package pipeline

import (
	"fmt"
	"image"
	"path"
	"reflect"
	"strings"
	"testing"

	"gocv.io/x/gocv"
	"gocv.io/x/gocv/contrib"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/settings"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/utils"
)

func TestClampInt(t *testing.T) {
	cases := []struct {
		name              string
		n, min, max, want int
	}{
		{
			name: "in range",
			n:    5,
			min:  1,
			max:  10,
			want: 5,
		},
		{
			name: "below range",
			n:    -1,
			min:  1,
			max:  10,
			want: 1,
		},
		{
			name: "above range",
			n:    11,
			min:  1,
			max:  10,
			want: 10,
		},
		{
			name: "at bottom",
			n:    1,
			min:  1,
			max:  10,
			want: 1,
		},
		{
			name: "at top",
			n:    10,
			min:  1,
			max:  10,
			want: 10,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := utils.ClampInt(tc.n, tc.min, tc.max)
			if got != tc.want {
				t.Fatalf("Clamp(%d, %d, %d) return incorrect value. got %d, want %d", tc.n, tc.min, tc.max, got, tc.want)
			}
		})
	}
}

func sliceToHSVMat(sl [][][]uint8) gocv.Mat {
	h := gocv.NewMatWithSize(len(sl), len(sl[0]), gocv.MatTypeCV8U)
	defer h.Close()
	s := gocv.NewMatWithSize(len(sl), len(sl[0]), gocv.MatTypeCV8U)
	defer s.Close()
	v := gocv.NewMatWithSize(len(sl), len(sl[0]), gocv.MatTypeCV8U)
	defer v.Close()
	for ri, row := range sl {
		for ci, col := range row {
			h.SetUCharAt(ri, ci, col[0])
			s.SetUCharAt(ri, ci, col[1])
			v.SetUCharAt(ri, ci, col[2])
		}
	}
	m := gocv.NewMatWithSize(len(sl), len(sl[0]), gocv.MatTypeCV8UC3)
	gocv.Merge([]gocv.Mat{h, s, v}, &m)
	return m
}

func sliceToGrayscaleMat(sl [][]uint8) gocv.Mat {
	m := gocv.NewMatWithSize(len(sl), len(sl[0]), gocv.MatTypeCV8U)
	for ri, row := range sl {
		for ci, col := range row {
			m.SetUCharAt(ri, ci, col)
		}
	}
	return m
}

func compareMats(t *testing.T, got, want gocv.Mat) {
	hashes := []contrib.ImgHashBase{
		contrib.PHash{},
		contrib.AverageHash{},
		contrib.BlockMeanHash{},
		contrib.BlockMeanHash{Mode: contrib.BlockMeanHashMode1},
		contrib.ColorMomentHash{},
		contrib.NewMarrHildrethHash(),
		contrib.NewRadialVarianceHash(),
	}
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
	// if !reflect.DeepEqual(got.GetUCharAt(0, 0), want.GetUCharAt(0, 0)) {
	// 	t.Fatalf("First pixel not equal. got %v, want %v", got.GetUCharAt(0, 0), want.GetUCharAt(0, 0))
	// }
	// if !reflect.DeepEqual(got.GetUCharAt(720-1, 1080-1), want.GetUCharAt(720-1, 1080-1)) {
	// 	t.Fatalf("Last pixel not equal. got %v, want %v", got.GetUCharAt(720-1, 1080-1), want.GetUCharAt(720-1, 1080-1))
	// }

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

		gw := gocv.NewWindow(fmt.Sprintf("%s got", t.Name()))
		defer gw.Close()
		gw.ResizeWindow(1920, 1080)
		gw.IMShow(got)
		gw.WaitKey(1)
		ww := gocv.NewWindow(fmt.Sprintf("%s want", t.Name()))
		defer ww.Close()
		ww.ResizeWindow(1920, 1080)
		ww.IMShow(want)
		ww.WaitKey(5000)
	}
}

func TestRenderMask_horses(t *testing.T) {
	cases := []struct {
		name                string
		hueMin, hueMax      int
		satMin, satMax      int
		valMin, valMax      int
		grow                int
		cropLeft, cropRight int
		cropTop, cropBottom int
		inputMask           string
		want                string
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
			name:       "frame mask",
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
			inputMask:  "horses-720p-mask.png",
			want:       "input_mask.png",
		},
	}

	vc, err := gocv.VideoCaptureFile("testdata/horses-720p.mp4")
	if err != nil {
		t.Fatalf("Error loading video file: %v", err)
	}
	frame := gocv.NewMat()
	defer frame.Close()
	vc.Set(gocv.VideoCapturePosFrames, 0)
	ok := vc.Read(&frame)
	if !ok {
		t.Fatalf("Error loading frame: %v", err)
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ms := settings.Mask{
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
			RenderMask(frame, &got, ms)

			if tc.inputMask != "" {
				m := gocv.IMRead(
					path.Join("testdata", tc.inputMask),
					gocv.IMReadGrayScale,
				)
				defer m.Close()
				CombineMasks(mask.Include, got, &m, &got)
			}

			want := gocv.IMRead(
				path.Join("testdata/render_mask_output", tc.want),
				gocv.IMReadGrayScale,
			)
			defer want.Close()
			compareMats(t, got, want)
		})
	}
}

func TestRenderMask_manual(t *testing.T) {
	cases := []struct {
		name                string
		hsv                 [][][]uint8
		hueMin, hueMax      int
		satMin, satMax      int
		valMin, valMax      int
		grow                int
		cropLeft, cropRight int
		cropTop, cropBottom int
		inputMask           [][]uint8
		want                [][]uint8
	}{
		{
			name: "keep all",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{50, 50, 50},
				},
				{
					{100, 100, 100},
					{150, 150, 150},
				},
			},
			hueMax:     179,
			satMax:     255,
			valMax:     255,
			cropRight:  2,
			cropBottom: 2,
			want: [][]uint8{
				{255, 255},
				{255, 255},
			},
		},
		{
			name: "hue partial",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{50, 50, 50},
				},
				{
					{100, 100, 100},
					{150, 150, 150},
				},
			},
			hueMin:     25,
			hueMax:     179,
			satMax:     255,
			valMax:     255,
			cropRight:  2,
			cropBottom: 2,
			want: [][]uint8{
				{0, 255},
				{255, 255},
			},
		},
		{
			name: "sat partial",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{50, 50, 50},
				},
				{
					{100, 100, 100},
					{150, 150, 150},
				},
			},
			hueMax:     179,
			satMin:     25,
			satMax:     255,
			valMax:     255,
			cropRight:  2,
			cropBottom: 2,
			want: [][]uint8{
				{0, 255},
				{255, 255},
			},
		},
		{
			name: "val partial",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{50, 50, 50},
				},
				{
					{100, 100, 100},
					{150, 150, 150},
				},
			},
			hueMax:     179,
			satMax:     255,
			valMin:     25,
			valMax:     255,
			cropRight:  2,
			cropBottom: 2,
			want: [][]uint8{
				{0, 255},
				{255, 255},
			},
		},
		{
			name: "hsv partial",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{50, 50, 50},
				},
				{
					{100, 100, 100},
					{150, 150, 150},
				},
			},
			hueMin:     25,
			hueMax:     179,
			satMin:     75,
			satMax:     255,
			valMin:     110,
			valMax:     255,
			cropRight:  2,
			cropBottom: 2,
			want: [][]uint8{
				{0, 0},
				{0, 255},
			},
		},
		{
			name: "hsv grow",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{100, 100, 100},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
			},
			hueMin:     25,
			hueMax:     179,
			satMin:     75,
			satMax:     255,
			valMin:     25,
			valMax:     255,
			grow:       2,
			cropRight:  3,
			cropBottom: 3,
			want: [][]uint8{
				{0, 0, 0},
				{0, 255, 255},
				{0, 255, 255},
			},
		},
		{
			name: "hsv grow 3",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{100, 100, 100},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
			},
			hueMin:     25,
			hueMax:     179,
			satMin:     75,
			satMax:     255,
			valMin:     25,
			valMax:     255,
			grow:       3,
			cropRight:  4,
			cropBottom: 4,
			want: [][]uint8{
				{0, 0, 0, 0, 0},
				{0, 255, 255, 255, 0},
				{0, 255, 255, 255, 0},
				{0, 255, 255, 255, 0},
				{0, 0, 0, 0, 0},
			},
		},
		{
			name: "crop 1",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{50, 50, 50},
				},
				{
					{100, 100, 100},
					{150, 150, 150},
				},
			},
			hueMax:     179,
			satMax:     255,
			valMax:     255,
			cropRight:  1,
			cropBottom: 1,
			want: [][]uint8{
				{255, 0},
				{0, 0},
			},
		},
		{
			name: "crop 3",
			hsv: [][][]uint8{
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
				{
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
					{1, 1, 1},
				},
			},
			hueMax:     179,
			satMax:     255,
			valMax:     255,
			cropLeft:   1,
			cropRight:  4,
			cropTop:    1,
			cropBottom: 4,
			want: [][]uint8{
				{0, 0, 0, 0, 0},
				{0, 255, 255, 255, 0},
				{0, 255, 255, 255, 0},
				{0, 255, 255, 255, 0},
				{0, 0, 0, 0, 0},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ms := settings.Mask{
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
			hsv := sliceToHSVMat(tc.hsv)
			bgr := gocv.NewMat()
			defer bgr.Close()
			gocv.CvtColor(hsv, &bgr, gocv.ColorHSVToBGR)
			got := gocv.NewMat()
			defer got.Close()
			want := sliceToGrayscaleMat(tc.want)
			defer want.Close()
			RenderMask(bgr, &got, ms)
			compareMats(t, got, want)
		})
	}
}

func TestCombineMasks(t *testing.T) {
	cases := []struct {
		name   string
		mode   string
		top    [][]uint8
		bottom [][]uint8
		want   [][]uint8
	}{
		{
			name: "include all",
			mode: mask.Include,
			top: [][]uint8{
				{255, 255},
				{255, 255},
			},
			bottom: [][]uint8{
				{0, 255},
				{255, 255},
			},
			want: [][]uint8{
				{255, 255},
				{255, 255},
			},
		},
		{
			name: "include none",
			mode: mask.Include,
			top: [][]uint8{
				{0, 0},
				{0, 0},
			},
			bottom: [][]uint8{
				{0, 255},
				{255, 255},
			},
			want: [][]uint8{
				{0, 255},
				{255, 255},
			},
		},
		{
			name: "include partial",
			mode: mask.Include,
			top: [][]uint8{
				{255, 0},
				{0, 255},
			},
			bottom: [][]uint8{
				{0, 0},
				{255, 255},
			},
			want: [][]uint8{
				{255, 0},
				{255, 255},
			},
		},
		{
			name: "exclude all",
			mode: mask.Exclude,
			top: [][]uint8{
				{255, 255},
				{255, 255},
			},
			bottom: [][]uint8{
				{255, 255},
				{255, 255},
			},
			want: [][]uint8{
				{0, 0},
				{0, 0},
			},
		},
		{
			name: "exclude none",
			mode: mask.Exclude,
			top: [][]uint8{
				{0, 0},
				{0, 0},
			},
			bottom: [][]uint8{
				{255, 0},
				{255, 255},
			},
			want: [][]uint8{
				{255, 0},
				{255, 255},
			},
		},
		{
			name: "exclude partial",
			mode: mask.Exclude,
			top: [][]uint8{
				{255, 0},
				{0, 255},
			},
			bottom: [][]uint8{
				{255, 0},
				{255, 255},
			},
			want: [][]uint8{
				{0, 0},
				{255, 0},
			},
		},
		{
			name: "exclude no bottom",
			mode: mask.Exclude,
			top: [][]uint8{
				{255, 0},
				{0, 255},
			},
			want: [][]uint8{
				{0, 255},
				{255, 0},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			top := sliceToGrayscaleMat(tc.top)
			defer top.Close()
			var bottom *gocv.Mat
			if tc.bottom != nil {
				m := sliceToGrayscaleMat(tc.bottom)
				bottom = &m
			}

			got := gocv.NewMat()
			defer got.Close()
			want := sliceToGrayscaleMat(tc.want)
			defer want.Close()
			CombineMasks(tc.mode, top, bottom, &got)
			compareMats(t, got, want)
		})
	}
}

func TestZoomCropRectangle(t *testing.T) {
	cases := []struct {
		name            string
		zoomFactor      float64
		anchorCoords    []int
		videoDimensions []int
		maxDimensions   []int
		want            image.Rectangle
	}{
		// Zoom in - video width == display width
		{
			name:            "video width == display width, 0x0, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 0x0, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 500, 500),
		},
		{
			name:            "video width == display width, 0x0, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 333, 333),
		},
		{
			name:            "video width == display width, 1000x1000, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 1000x1000, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(500, 500, 1000, 1000),
		},
		{
			name:            "video width == display width, 1000x1000, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(667, 667, 1000, 1000),
		},
		{
			name:            "video width == display width, 0x749, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 0x749, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 499, 500, 999),
		},
		{
			name:            "video width == display width, 0x749, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 583, 333, 916),
		},
		{
			name:            "video width == display width, 749x0, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 749x0, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(499, 0, 999, 500),
		},
		{
			name:            "video width == display width, 749x0, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(583, 0, 916, 333),
		},
		// Zoom in - video width > display width
		{
			name:            "video width > display width, 0x0, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video width > display width, 0x0, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 50, 50),
		},
		{
			name:            "video width > display width, 0x0, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 33, 33),
		},
		{
			name:            "video width > display width, 1000x1000, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(900, 900, 1000, 1000),
		},
		{
			name:            "video width > display width, 1000x1000, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(950, 950, 1000, 1000),
		},
		{
			name:            "video width > display width, 1000x1000, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(967, 967, 1000, 1000),
		},
		{
			name:            "video width > display width, 0x749, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 699, 100, 799),
		},
		{
			name:            "video width > display width, 0x749, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 724, 50, 774),
		},
		{
			name:            "video width > display width, 0x749, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 733, 33, 766),
		},
		{
			name:            "video width > display width, 749x0, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(699, 0, 799, 100),
		},
		{
			name:            "video width > display width, 749x0, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(724, 0, 774, 50),
		},
		{
			name:            "video width > display width, 749x0, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(733, 0, 766, 33),
		},
		// Zoom in - video size < max size
		{
			name:            "video size < max size, 0x0, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x0, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x0, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x0, 10x",
			zoomFactor:      10.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x0, 20x",
			zoomFactor:      20.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 50, 50),
		},
		{
			name:            "video size < max size, 0x0, 30x",
			zoomFactor:      30.0,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 33, 33),
		},
		{
			name:            "video size < max size, 100x100, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 100x100, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 100x100, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 100x100, 10x",
			zoomFactor:      10.0,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 100x100, 20x",
			zoomFactor:      20.0,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(50, 50, 100, 100),
		},
		{
			name:            "video size < max size, 100x100, 30x",
			zoomFactor:      30.0,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(67, 67, 100, 100),
		},
		{
			name:            "video size < max size, 0x74, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{0, 74},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x74, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{0, 74},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x74, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{0, 74},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x74, 10x",
			zoomFactor:      10.0,
			anchorCoords:    []int{0, 74},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 0x74, 20x",
			zoomFactor:      20.0,
			anchorCoords:    []int{0, 74},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 49, 50, 99),
		},
		{
			name:            "video size < max size, 0x74, 30x",
			zoomFactor:      30.0,
			anchorCoords:    []int{0, 74},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 58, 33, 91),
		},
		{
			name:            "video size < max size, 74x0, 1x",
			zoomFactor:      1.0,
			anchorCoords:    []int{74, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 74x0, 2x",
			zoomFactor:      2.0,
			anchorCoords:    []int{74, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 74x0, 3x",
			zoomFactor:      3.0,
			anchorCoords:    []int{74, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 74x0, 10x",
			zoomFactor:      10.0,
			anchorCoords:    []int{74, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video size < max size, 74x0, 20x",
			zoomFactor:      20.0,
			anchorCoords:    []int{74, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(49, 0, 99, 50),
		},
		{
			name:            "video size < max size, 74x0, 30x",
			zoomFactor:      30.0,
			anchorCoords:    []int{74, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(58, 0, 91, 33),
		},
		// Zoom out - video width == display width
		{
			name:            "video width == display width, 0x0, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 0x0, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 0x0, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 1000x1000, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 1000x1000, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video width == display width, 1000x1000, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		// Zoom out - video width < display width
		{
			name:            "video width < display width, 0x0, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video width < display width, 0x0, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video width < display width, 0x0, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video width < display width, 100x100, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video width < display width, 100x100, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		{
			name:            "video width < display width, 100x100, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{100, 100},
			videoDimensions: []int{100, 100},
			maxDimensions:   []int{1000, 1000},
			want:            image.Rect(0, 0, 100, 100),
		},
		// Zoom out - video size > max size
		{
			name:            "video size > max size, 0x0, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 200, 200),
		},
		{
			name:            "video size > max size, 0x0, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 500, 500),
		},
		{
			name:            "video size > max size, 0x0, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{0, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video size > max size, 1000x1000, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(800, 800, 1000, 1000),
		},
		{
			name:            "video size > max size, 1000x1000, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(500, 500, 1000, 1000),
		},
		{
			name:            "video size > max size, 1000x1000, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{1000, 1000},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video size > max size, 0x749, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 649, 200, 849),
		},
		{
			name:            "video size > max size, 0x749, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 499, 500, 999),
		},
		{
			name:            "video size > max size, 0x749, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{0, 749},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 1000, 1000),
		},
		{
			name:            "video size > max size, 749x0, .5x",
			zoomFactor:      0.5,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(649, 0, 849, 200),
		},
		{
			name:            "video size > max size, 749x0, .2x",
			zoomFactor:      0.2,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(499, 0, 999, 500),
		},
		{
			name:            "video size > max size, 749x0, .1x",
			zoomFactor:      0.1,
			anchorCoords:    []int{749, 0},
			videoDimensions: []int{1000, 1000},
			maxDimensions:   []int{100, 100},
			want:            image.Rect(0, 0, 1000, 1000),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			anchorX, anchorY := tc.anchorCoords[0], tc.anchorCoords[1]
			videoWidth, videoHeight := tc.videoDimensions[0], tc.videoDimensions[1]
			maxWidth, maxHeight := tc.maxDimensions[0], tc.maxDimensions[1]
			got := ZoomCropRectangle(tc.zoomFactor, anchorX, anchorY, videoWidth, videoHeight, maxWidth, maxHeight)
			if got != tc.want {
				t.Fatalf("ZoomCropRectangle(%s) return unexpected result. got %v, want %v", tc.name, got, tc.want)
			}
		})
	}
}
