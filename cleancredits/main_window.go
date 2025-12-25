package cleancredits

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/display"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/mask"
	"gocv.io/x/gocv"
)

const (
	MaskTabName   = "Mask"
	DrawTabName   = "Draw"
	RenderTabName = "Render"
)

func NewMainWindow(a fyne.App) fyne.Window {
	w := a.NewWindow("cleancredits")
	w.SetMaster()

	button := widget.NewButton("Open video file", func() { openVideo(w) })
	content := container.New(layout.NewCenterLayout(), button)
	w.Resize(fyne.NewSize(720, 480))
	w.SetContent(content)
	return w
}

func openVideo(w fyne.Window) {
	dialog.ShowFileOpen(func(reader fyne.URIReadCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("error loading file: %v", err), w)
			return
		}
		if reader == nil {
			fmt.Println("No file selected")
			return
		}

		videoPath := reader.URI().Path()
		reader.Close()
		vc, err := gocv.VideoCaptureFile(videoPath)
		if err != nil {
			fmt.Println("Error loading video file: ", err)
			w.Close()
			return
		}
		w.SetContent(buildCleaner(vc))
	}, w)
}

func buildCleaner(vc *gocv.VideoCapture) fyne.CanvasObject {
	videoWidth := int(vc.Get(gocv.VideoCaptureFrameWidth))
	videoHeight := int(vc.Get(gocv.VideoCaptureFrameWidth))
	frameCount := int(vc.Get(gocv.VideoCaptureFrameCount))
	maskForm := mask.NewForm(frameCount, videoWidth, videoHeight)
	selectedTab := binding.NewString()

	maskTab := container.NewTabItem(MaskTabName, maskForm.Container)
	drawTab := container.NewTabItem(DrawTabName, widget.NewLabel("Draw tab"))
	renderTab := container.NewTabItem(RenderTabName, widget.NewLabel("Render tab"))
	left := container.NewAppTabs(maskTab, drawTab, renderTab)
	left.OnSelected = func(ti *container.TabItem) {
		switch ti {
		case maskTab:
			selectedTab.Set(MaskTabName)
		case drawTab:
			selectedTab.Set(DrawTabName)
		case renderTab:
			selectedTab.Set(RenderTabName)
		}
	}

	display := display.NewDisplay(vc, selectedTab, maskForm)
	content := container.New(layout.NewHBoxLayout(), left, display.Container)

	return content
}
