package mask

const (
	Include = "Always inpaint"
	Exclude = "Never inpaint"
)

const (
	HueMax = 179
	SatMax = 255
	ValMax = 255
)

type Layer struct {
	Frame                 int
	Mode                  string
	Grow                  int
	HueMin, HueMax        int
	SatMin, SatMax        int
	ValMin, ValMax        int
	CropLeft, CropTop     int
	CropRight, CropBottom int
}

func NewLayer(frame int, videoWidth, videoHeight int) Layer {
	return Layer{
		Frame:      frame,
		Mode:       Include,
		HueMax:     HueMax,
		SatMax:     SatMax,
		ValMax:     ValMax,
		CropRight:  videoWidth,
		CropBottom: videoHeight,
	}
}
