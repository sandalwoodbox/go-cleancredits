package mask

import (
	"fyne.io/fyne/v2/widget"
)

func NewForm(maxFrame, videoWidth, videoHeight int) *widget.Form {
	return widget.NewForm(
		widget.NewFormItem("Frame", widget.NewSlider(0, float64(maxFrame))),
		widget.NewFormItem("Mode", widget.NewSelect([]string{Include, Exclude}, nil)),
		widget.NewFormItem("Grow", widget.NewSlider(0, float64(videoHeight))),

		widget.NewFormItem("Hue / Saturation / Value", widget.NewLabel("")),
		widget.NewFormItem("Hue Min", widget.NewSlider(0, HueMax)),
		widget.NewFormItem("Hue Max", widget.NewSlider(0, HueMax)),
		widget.NewFormItem("Sat Min", widget.NewSlider(0, SatMax)),
		widget.NewFormItem("Sat Max", widget.NewSlider(0, SatMax)),
		widget.NewFormItem("Val Min", widget.NewSlider(0, ValMax)),
		widget.NewFormItem("Val Max", widget.NewSlider(0, ValMax)),

		widget.NewFormItem("Crop", widget.NewLabel("")),
		widget.NewFormItem("Left", widget.NewSlider(0, float64(videoWidth))),
		widget.NewFormItem("Top", widget.NewSlider(0, float64(videoHeight))),
		widget.NewFormItem("Right", widget.NewSlider(0, float64(videoWidth))),
		widget.NewFormItem("Bottom", widget.NewSlider(0, float64(videoHeight))),
	)
}
