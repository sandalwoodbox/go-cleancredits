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
}

func NewPreview() Preview {
	i := image.NewRGBA(image.Rect(0, 0, 1, 1))
	i.Set(0, 0, color.RGBA{0, 0, 0, 0})
	img := canvas.NewImageFromImage(i)
	img.FillMode = canvas.ImageFillContain
	img.SetMinSize(fyne.NewSize(720, 480))
	p := Preview{
		Image:     img,
		Container: container.NewStack(img),
	}

	return p
}

func (p Preview) SetImage(img image.Image) {
	p.Image.Image = img
	p.Image.Refresh()
}
