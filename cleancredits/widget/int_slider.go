package widget

import (
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type IntSlider struct {
	widget.Slider
}

func NewIntSlider(min, max int) *IntSlider {
	s := &IntSlider{
		Slider: widget.Slider{
			Value:       0,
			Min:         float64(min),
			Max:         float64(max),
			Step:        1,
			Orientation: widget.Horizontal,
		},
	}
	s.ExtendBaseWidget(s)
	return s
}

func NewIntSliderWithData(min, max int, data binding.Int) *IntSlider {
	s := NewIntSlider(min, max)
	s.Bind(binding.IntToFloat(data))

	return s
}
