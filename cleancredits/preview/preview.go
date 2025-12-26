package preview

import (
	"image"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
)

type Preview struct {
	Image *canvas.Image
}

func NewPreview() Preview {
	img := &canvas.Image{}
	p := Preview{
		Image: img,
	}
	p.Image.FillMode = canvas.ImageFillContain
	p.Image.SetMinSize(fyne.NewSize(720, 480))

	return p
}

func (p Preview) SetImage(img image.Image) {
	p.Image.Image = img
	p.Image.Refresh()
}
