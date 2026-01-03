package preview

import (
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
)

type Preview struct {
	Image     *canvas.Image
	Container *fyne.Container
	Width     int
	Height    int
}

func NewPreview(w, h int) Preview {
	i := image.NewRGBA(image.Rect(0, 0, 1, 1))
	i.Set(0, 0, color.RGBA{0, 0, 0, 0})
	img := canvas.NewImageFromImage(i)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(float32(w), float32(h)))
	p := Preview{
		Image:     img,
		Container: container.NewCenter(img),
		Width:     w,
		Height:    h,
	}

	return p
}

func (p Preview) SetImage(img image.Image) {
	p.Image.Image = img
	r := img.Bounds()
	if r.Dx() < p.Width && r.Dy() < p.Height {
		p.Image.FillMode = canvas.ImageFillOriginal
	} else {
		p.Image.FillMode = canvas.ImageFillContain
	}
	p.Image.SetMinSize(fyne.NewSize(float32(p.Width), float32(p.Height)))
	p.Image.Refresh()
}
