package overlay

import "image/color"

func DimmedCellColor(base color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(float64(base.R) * 0.55),
		G: uint8(float64(base.G) * 0.55),
		B: uint8(float64(base.B) * 0.55),
		A: 255,
	}
}
