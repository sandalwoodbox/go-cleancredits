package cleancredits

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/cleaner"
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
		c := cleaner.New(vc)
		w.SetContent(c.Container)
	}, w)
}
