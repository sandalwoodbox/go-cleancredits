package cleaner

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/display"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/draw"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/pipeline"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/preview"
)

const (
	MaskTabName   = "Mask"
	DrawTabName   = "Draw"
	RenderTabName = "Render"
)

type Cleaner struct {
	Container    *fyne.Container
	VideoCapture *gocv.VideoCapture

	MaskForm    mask.Form
	DrawForm    draw.Form
	DisplayForm display.Form
	SelectedTab binding.String

	UpdateChannel chan struct{}
	Pipeline      pipeline.Pipeline
	Preview       preview.Preview
}

func New(vc *gocv.VideoCapture) Cleaner {
	videoWidth := int(vc.Get(gocv.VideoCaptureFrameWidth))
	videoHeight := int(vc.Get(gocv.VideoCaptureFrameWidth))
	frameCount := int(vc.Get(gocv.VideoCaptureFrameCount))
	c := Cleaner{
		VideoCapture:  vc,
		MaskForm:      mask.NewForm(frameCount, videoWidth, videoHeight),
		DrawForm:      draw.NewForm(frameCount),
		DisplayForm:   display.NewForm(),
		SelectedTab:   binding.NewString(),
		UpdateChannel: make(chan struct{}),
		Pipeline:      pipeline.NewPipeline(vc),
		Preview:       preview.NewPreview(),
	}
	maskTab := container.NewTabItem(MaskTabName, c.MaskForm.Container)
	drawTab := container.NewTabItem(DrawTabName, c.DrawForm.Container)
	renderTab := container.NewTabItem(RenderTabName, widget.NewLabel("Render tab"))
	left := container.NewAppTabs(maskTab, drawTab, renderTab)
	left.OnSelected = func(ti *container.TabItem) {
		switch ti {
		case maskTab:
			c.SelectedTab.Set(MaskTabName)
		case drawTab:
			c.SelectedTab.Set(DrawTabName)
		case renderTab:
			c.SelectedTab.Set(RenderTabName)
		}
	}
	imgPreview := preview.NewPreview()
	right := container.New(
		layout.NewVBoxLayout(),
		c.DisplayForm.Container,
		imgPreview.Image,
	)

	c.Container = container.New(layout.NewHBoxLayout(), left, right)

	// Update pipeline & preview when forms change
	scheduleUpdate := func() {
		select {
		case c.UpdateChannel <- struct{}{}:
			// fmt.Println("Update scheduled")
		default:
			// fmt.Println("Update already scheduled")
		}
	}
	go func() {
		for range c.UpdateChannel {
			fyne.DoAndWait(c.UpdatePipeline)
		}
	}()
	scheduleUpdateListener := binding.NewDataListener(scheduleUpdate)
	c.SelectedTab.AddListener(scheduleUpdateListener)
	c.MaskForm.OnChange(scheduleUpdate)
	c.DrawForm.OnChange(scheduleUpdate)
	c.DisplayForm.OnChange(scheduleUpdate)

	return c
}

func (c Cleaner) UpdatePipeline() {
	maskSettings, err := c.MaskForm.Settings()
	if err != nil {
		fmt.Println("Error getting mask settings: ", err)
		return
	}
	drawSettings, err := c.DrawForm.Settings()
	if err != nil {
		fmt.Println("Error getting draw settings: ", err)
		return
	}
	displaySettings, err := c.DisplayForm.Settings()
	if err != nil {
		fmt.Println("Error getting display settings: ", err)
		return
	}

	tabName, err := c.SelectedTab.Get()
	if err != nil {
		fmt.Println("Error getting selected tab: ", err)
		return
	}

	c.Pipeline.UpdateMask(
		maskSettings,
		drawSettings,
		displaySettings,
	)
	fNum := maskSettings.Frame
	switch tabName {
	case "Draw":
		fNum = drawSettings.Frame
	}
	c.Preview.SetImage(c.Pipeline.ApplyMask(fNum))
}
