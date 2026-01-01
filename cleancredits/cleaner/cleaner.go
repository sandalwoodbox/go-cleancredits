package cleaner

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/display"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/draw"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/pipeline"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/preview"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/render"
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
	RenderForm  render.Form
	SelectedTab binding.String

	UpdateChannel chan struct{}
	Pipeline      *pipeline.Pipeline
	Preview       preview.Preview
}

func New(vc *gocv.VideoCapture, w fyne.Window) Cleaner {
	videoWidth := int(vc.Get(gocv.VideoCaptureFrameWidth))
	videoHeight := int(vc.Get(gocv.VideoCaptureFrameHeight))
	frameCount := int(vc.Get(gocv.VideoCaptureFrameCount))
	displayWidth := 720
	displayHeight := 480
	p := pipeline.NewPipeline(vc, displayWidth, displayHeight)
	c := Cleaner{
		VideoCapture:  vc,
		MaskForm:      mask.NewForm(frameCount, videoWidth, videoHeight),
		DrawForm:      draw.NewForm(frameCount),
		DisplayForm:   display.NewForm(videoWidth, videoHeight),
		SelectedTab:   binding.NewString(),
		UpdateChannel: make(chan struct{}),
		Pipeline:      &p,
		Preview:       preview.NewPreview(displayWidth, displayHeight),
	}
	c.RenderForm = render.NewForm(frameCount, c.Pipeline, w)
	maskTab := container.NewTabItem(MaskTabName, c.MaskForm.Container)
	drawTab := container.NewTabItem(DrawTabName, c.DrawForm.Container)
	renderTab := container.NewTabItem(RenderTabName, c.RenderForm.Container)
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
	right := container.New(
		layout.NewVBoxLayout(),
		c.DisplayForm.Container,
		c.Preview.Container,
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
	c.RenderForm.OnChange(scheduleUpdate)
	// Change draw tab frame when mask frame changes (but not vice versa)
	c.MaskForm.Frame.AddListener(binding.NewDataListener(func() {
		f, err := c.MaskForm.Frame.Get()
		if err != nil {
			fmt.Println("Error getting MaskForm.Frame: ", err)
			return
		}
		err = c.DrawForm.Frame.Set(f)
		if err != nil {
			fmt.Println("Error getting DrawForm.Frame: ", err)
			return
		}
	}))

	return c
}

func (c *Cleaner) UpdatePipeline() {
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
	renderSettings, err := c.RenderForm.Settings()
	if err != nil {
		fmt.Println("Error getting render settings: ", err)
		return
	}

	tabName, err := c.SelectedTab.Get()
	if err != nil {
		fmt.Println("Error getting selected tab: ", err)
		return
	}

	err = c.Pipeline.UpdateMask(
		maskSettings,
		drawSettings,
	)
	if err != nil {
		fmt.Println("Error updating mask: ", err)
		return
	}
	fNum := maskSettings.Frame
	switch tabName {
	case "Draw":
		fNum = drawSettings.Frame
	case "Render":
		fNum = renderSettings.Frame
	}
	img, err := c.Pipeline.ApplyMask(fNum, displaySettings, renderSettings)
	if err != nil {
		fmt.Println("Error applying mask: ", err)
		return
	}

	c.Preview.SetImage(img)
}
