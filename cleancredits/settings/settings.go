package settings

type Display struct {
	Mode    string
	Zoom    string
	AnchorX int
	AnchorY int
}

type Draw struct {
	Frame int
}

type Mask struct {
	Frame      int
	Mode       string
	HueMin     int
	HueMax     int
	SatMin     int
	SatMax     int
	ValMin     int
	ValMax     int
	Grow       int
	CropLeft   int
	CropTop    int
	CropRight  int
	CropBottom int
}

type Render struct {
	Frame         int
	StartFrame    int
	EndFrame      int
	InpaintRadius int
}
