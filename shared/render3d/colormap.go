package render3d

import "image/color"

// ColormapFunc maps a normalized value t in [0,1] to an RGBA color.
type ColormapFunc func(t float64) color.RGBA

// GetColormap returns the colormap function for the given name.
// Supported: "viridis", "plasma", "coolwarm". Default is viridis.
func GetColormap(name string) ColormapFunc {
	switch name {
	case "viridis":
		return ViridisColor
	case "plasma":
		return PlasmaColor
	case "coolwarm":
		return CoolwarmColor
	default:
		return ViridisColor
	}
}

// ViridisColor approximates the Viridis colormap (perceptually uniform, colorblind-safe).
func ViridisColor(t float64) color.RGBA {
	t = clamp01(t)
	r := 0.267 + t*(0.993*t-0.068)
	g := 0.005 + t*(0.991-0.149*t)
	b := 0.329 + t*(0.288-0.147*t)
	return color.RGBA{
		R: uint8(clamp01(r) * 255),
		G: uint8(clamp01(g) * 255),
		B: uint8(clamp01(b) * 255),
		A: 255,
	}
}

// PlasmaColor approximates the Plasma colormap (perceptually uniform, colorblind-safe).
func PlasmaColor(t float64) color.RGBA {
	t = clamp01(t)
	r := 0.05 + t*0.89
	g := 0.03 + t*0.95*t
	b := 0.53 - t*0.40
	return color.RGBA{
		R: uint8(clamp01(r) * 255),
		G: uint8(clamp01(g) * 255),
		B: uint8(clamp01(b) * 255),
		A: 255,
	}
}

// CoolwarmColor produces a diverging blue-white-red colormap.
func CoolwarmColor(t float64) color.RGBA {
	t = clamp01(t)
	if t < 0.5 {
		s := t * 2
		return color.RGBA{
			R: uint8(s * 255),
			G: uint8(s * 255),
			B: 255,
			A: 255,
		}
	}
	s := (t - 0.5) * 2
	return color.RGBA{
		R: 255,
		G: uint8((1 - s) * 255),
		B: uint8((1 - s) * 255),
		A: 255,
	}
}

// clamp01 clamps v to [0, 1].
func clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
