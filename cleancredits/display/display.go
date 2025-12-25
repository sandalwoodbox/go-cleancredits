package display

import (
	"fmt"
	"slices"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

const (
	ViewMask     = "Areas to inpaint"
	ViewDraw     = "Overrides"
	ViewPreview  = "Preview"
	ViewOriginal = "Original"
)

const ZoomFit = "Fit"

var ZoomLevelMap = map[string]float32{
	ZoomFit: 0,
	"10%":   .10,
	"25%":   .25,
	"50%":   .5,
	"100%":  1,
	"150%":  1.5,
	"200%":  2,
	"300%":  3,
	"400%":  4,
	"500%":  5,
}

var ZoomLevels = []string{
	ZoomFit,
	"10%",
	"25%",
	"50%",
	"100%",
	"150%",
	"200%",
	"300%",
	"400%",
	"500%",
}

type Display struct {
	VideoCapture  *gocv.VideoCapture
	RenderChannel chan struct{}
	Container     *fyne.Container
	Image         *canvas.Image
	SelectedTab   binding.String
	MaskForm      mask.Form

	Mode    binding.String
	Zoom    binding.String
	AnchorX binding.Int
	AnchorY binding.Int
}

func NewDisplay(vc *gocv.VideoCapture, selectedTab binding.String, maskForm mask.Form) Display {
	img := &canvas.Image{}
	d := Display{
		VideoCapture:  vc,
		RenderChannel: make(chan struct{}),
		Image:         img,
		SelectedTab:   selectedTab,
		MaskForm:      maskForm,

		Mode:    binding.NewString(),
		Zoom:    binding.NewString(),
		AnchorX: binding.NewInt(),
		AnchorY: binding.NewInt(),
	}
	d.Mode.Set(ViewMask)
	d.Zoom.Set(ZoomFit)
	anchorXEntry := ccWidget.NewIntEntryWithData(d.AnchorX)
	anchorYEntry := ccWidget.NewIntEntryWithData(d.AnchorY)
	d.Container = container.New(
		layout.NewVBoxLayout(),
		container.New(
			layout.NewHBoxLayout(),
			widget.NewLabel("View"),
			widget.NewSelectWithData(
				[]string{
					ViewMask,
					ViewDraw,
					ViewPreview,
					ViewOriginal,
				},
				d.Mode),
			widget.NewLabel("Zoom"),
			widget.NewSelectWithData(
				ZoomLevels, d.Zoom,
			),
			widget.NewButtonWithIcon("", theme.ZoomInIcon(), d.ZoomIn),
			widget.NewButtonWithIcon("", theme.ZoomOutIcon(), d.ZoomOut),
			widget.NewLabel("Anchor X"),
			anchorXEntry,
			widget.NewLabel("Y"),
			anchorYEntry,
		),
		d.Image,
	)
	scheduleRenderListener := binding.NewDataListener(func() {
		d.ScheduleRender()
	})
	selectedTab.AddListener(scheduleRenderListener)
	maskForm.OnChange(func() {
		d.ScheduleRender()
	})
	d.Mode.AddListener(scheduleRenderListener)
	d.Zoom.AddListener(scheduleRenderListener)
	d.AnchorX.AddListener(scheduleRenderListener)
	d.AnchorY.AddListener(scheduleRenderListener)

	go func() {
		for range d.RenderChannel {
			fyne.DoAndWait(d.Render)
		}
	}()
	return d
}

func (d Display) ZoomIn() {
	z, err := d.Zoom.Get()
	if err != nil {
		fmt.Println("Error getting Zoom: ", err)
	}
	if z == ZoomFit {
		fmt.Println("Zoom in/out from Fit not supported")
		return
	}
	i := slices.Index(ZoomLevels, z)
	if i < len(ZoomLevels)-1 {
		err = d.Zoom.Set(ZoomLevels[i+1])
		if err != nil {
			fmt.Println("Error setting Zoom: ", err)
		}
	}
}

func (d Display) ZoomOut() {
	z, err := d.Zoom.Get()
	if err != nil {
		fmt.Println("Error getting Zoom: ", err)
	}
	if z == ZoomFit {
		fmt.Println("Zoom in/out from Fit not supported")
		return
	}
	i := slices.Index(ZoomLevels, z)
	if i > 1 {
		err = d.Zoom.Set(ZoomLevels[i-1])
		if err != nil {
			fmt.Println("Error setting Zoom: ", err)
		}
	}
}

func (d Display) ScheduleRender() {
	select {
	case d.RenderChannel <- struct{}{}:
		// fmt.Println("Render scheduled")
	default:
		// fmt.Println("Render already scheduled")
	}
}

func (d Display) Render() {
	// fmt.Println("Render")
	mat := gocv.NewMat()
	defer mat.Close()

	frame, err := d.MaskForm.Frame.Get()
	if err != nil {
		fmt.Println("Error getting frame number: ", err)
		return
	}
	d.VideoCapture.Set(
		gocv.VideoCapturePosFrames,
		float64(frame),
	)
	d.VideoCapture.Read(&mat)
	img, err := mat.ToImage()
	if err != nil {
		fmt.Printf(
			"Error loading frame %d/%s: %v\n",
			frame,
			strconv.FormatFloat(d.VideoCapture.Get(gocv.VideoCaptureFrameCount), 'f', -1, 64),
			err)
		return
	}
	d.Image.FillMode = canvas.ImageFillContain
	d.Image.SetMinSize(fyne.NewSize(720, 480))
	d.Image.Image = img
	d.Image.Refresh()
}
