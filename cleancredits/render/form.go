package render

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"gocv.io/x/gocv"

	"github.com/sandalwoodbox/go-cleancredits/cleancredits/pipeline"
	"github.com/sandalwoodbox/go-cleancredits/cleancredits/settings"
	ccWidget "github.com/sandalwoodbox/go-cleancredits/cleancredits/widget"
)

type Form struct {
	Container     *fyne.Container
	Window        fyne.Window
	ProgressBar   *widget.ProgressBar
	ProgressLabel *widget.Label
	Pipeline      *pipeline.Pipeline

	StartFrame    binding.Int
	EndFrame      binding.Int
	InpaintRadius binding.Int
}

func NewForm(frameCount int, p *pipeline.Pipeline, w fyne.Window) Form {
	f := Form{
		Window:        w,
		Pipeline:      p,
		StartFrame:    binding.NewInt(),
		EndFrame:      binding.NewInt(),
		InpaintRadius: binding.NewInt(),
	}
	err := f.InpaintRadius.Set(3)
	if err != nil {
		fmt.Println("Error setting default inpaint radius")
	}
	f.ProgressBar = widget.NewProgressBar()
	f.ProgressBar.Hide()
	f.ProgressLabel = widget.NewLabel("")
	f.ProgressLabel.Hide()
	f.Container = container.New(
		layout.NewVBoxLayout(),
		container.New(
			layout.NewGridLayout(3),
			widget.NewLabel("Start frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.StartFrame), ccWidget.NewIntEntryWithData(f.StartFrame),
			widget.NewLabel("End frame"), ccWidget.NewIntSliderWithData(0, frameCount-1, f.EndFrame), ccWidget.NewIntEntryWithData(f.EndFrame),
			widget.NewLabel("Inpaint radius"), ccWidget.NewIntSliderWithData(0, 10, f.InpaintRadius), ccWidget.NewIntEntryWithData(f.InpaintRadius),
			widget.NewButton("Render", f.ShowRenderSave), widget.NewLabel(""), widget.NewLabel(""),
		),
		container.New(
			layout.NewVBoxLayout(),
			f.ProgressBar,
			f.ProgressLabel,
		),
	)
	f.OnChange(func() {
		f.ProgressBar.Hide()
		f.ProgressLabel.Hide()
	})
	return f
}

func (f Form) OnChange(fn func()) {
	l := binding.NewDataListener(fn)
	f.StartFrame.AddListener(l)
	f.EndFrame.AddListener(l)
	f.InpaintRadius.AddListener(l)
}

func (f Form) Settings() (settings.Render, error) {
	startFrame, err := f.StartFrame.Get()
	if err != nil {
		return settings.Render{}, fmt.Errorf("getting startFrame: %v", err)
	}
	endFrame, err := f.EndFrame.Get()
	if err != nil {
		return settings.Render{}, fmt.Errorf("getting endFrame: %v", err)
	}
	inpaintRadius, err := f.InpaintRadius.Get()
	if err != nil {
		return settings.Render{}, fmt.Errorf("getting inpaintRadius: %v", err)
	}
	return settings.Render{
		StartFrame:    startFrame,
		EndFrame:      endFrame,
		InpaintRadius: inpaintRadius,
	}, nil
}

func (f *Form) ShowRenderSave() {
	dialog.ShowFileSave(func(writer fyne.URIWriteCloser, err error) {
		if err != nil {
			dialog.ShowError(fmt.Errorf("error choosing save file: %v", err), f.Window)
		}
		if writer == nil {
			fmt.Println("No file selected")
			return
		}
		f.Render(writer.URI().Path())
	}, f.Window)
}

func (f *Form) Render(path string) {
	f.ProgressLabel.SetText("")
	f.ProgressLabel.Show()
	rs, err := f.Settings()
	if err != nil {
		f.ProgressLabel.SetText(fmt.Sprintf("Error gettings settings: %v", err))
		return
	}
	frameCount := rs.EndFrame - rs.StartFrame + 1
	// Steps: load frame, clean frame, render frame
	stepCount := frameCount * 3
	f.ProgressBar.Min = 0
	f.ProgressBar.Max = float64(stepCount)
	f.ProgressBar.SetValue(0)
	f.ProgressBar.Show()

	codec := f.Pipeline.VideoCapture.CodecString()
	fps := f.Pipeline.VideoCapture.Get(gocv.VideoCaptureFPS)

	mask, err := pipeline.ImageToMatGray(*f.Pipeline.MaskWithOverrides)
	if err != nil {
		mask.Close()
		f.ProgressLabel.SetText(fmt.Sprintf("converting MaskWithOverrides to mat: %v", err))
		return
	}
	defer mask.Close()

	out, err := gocv.VideoWriterFile(path, codec, fps, f.Pipeline.VideoWidth, f.Pipeline.VideoHeight, true)
	for i := rs.StartFrame; i <= rs.EndFrame; i++ {
		m := gocv.NewMat()
		err := pipeline.LoadFrame(f.Pipeline.VideoCapture, i, &m)
		if err != nil {
			f.ProgressLabel.SetText(err.Error())
			return
		}
		f.ProgressBar.SetValue(f.ProgressBar.Value + 1)
		f.ProgressLabel.SetText("%d/%d loaded")

		masked := gocv.NewMat()
		gocv.Inpaint(m, mask, &masked, float32(rs.InpaintRadius), gocv.Telea)
		f.ProgressBar.SetValue(f.ProgressBar.Value + 1)
		f.ProgressLabel.SetText("%d/%d inpainted")

		out.Write(masked)
		m.Close()
		masked.Close()
		f.ProgressBar.SetValue(f.ProgressBar.Value + 1)
		f.ProgressLabel.SetText("%d/%d rendered out")
	}
	err = out.Close()
	if err != nil {
		f.ProgressLabel.SetText("Error finalizing output")
	}
	f.ProgressLabel.SetText(fmt.Sprintf("Finished rendering %d-%d", rs.StartFrame, rs.EndFrame))
}
