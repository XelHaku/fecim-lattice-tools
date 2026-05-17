//go:build legacy_fyne

package overlay

import "image/color"

func DimmedCellColor(base color.RGBA) color.RGBA {
	return color.RGBA{
		R: uint8(float64(base.R) * 0.75),
		G: uint8(float64(base.G) * 0.75),
		B: uint8(float64(base.B) * 0.75),
		A: 255,
	}
}
