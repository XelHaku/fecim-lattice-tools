// Package widgets provides reusable UI components.
package widgets

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// ColorLegend displays a gradient bar with min/max labels and optional units.
type ColorLegend struct {
	widget.BaseWidget

	// Display properties
	minValue float64
	maxValue float64
	units    string
	vertical bool

	// Colormap function (normalized 0-1 to color)
	colorFunc func(float64) color.RGBA

	// Visual components
	raster   *canvas.Raster
	minLabel *widget.Label
	maxLabel *widget.Label
}

// NewColorLegend creates a new color legend widget.
// colorFunc takes a normalized value [0,1] and returns a color.
func NewColorLegend(minValue, maxValue float64, units string, vertical bool, colorFunc func(float64) color.RGBA) *ColorLegend {
	cl := &ColorLegend{
		minValue:  minValue,
		maxValue:  maxValue,
		units:     units,
		vertical:  vertical,
		colorFunc: colorFunc,
	}

	// Create labels
	minText := fmt.Sprintf("%.3f", minValue)
	maxText := fmt.Sprintf("%.3f", maxValue)
	if units != "" {
		minText += " " + units
		maxText += " " + units
	}

	cl.minLabel = widget.NewLabel(minText)
	cl.minLabel.TextStyle = fyne.TextStyle{Monospace: true}
	cl.maxLabel = widget.NewLabel(maxText)
	cl.maxLabel.TextStyle = fyne.TextStyle{Monospace: true}

	cl.ExtendBaseWidget(cl)
	return cl
}

// SetRange updates the min/max values and refreshes the legend.
func (cl *ColorLegend) SetRange(minValue, maxValue float64) {
	cl.minValue = minValue
	cl.maxValue = maxValue

	minText := fmt.Sprintf("%.3f", minValue)
	maxText := fmt.Sprintf("%.3f", maxValue)
	if cl.units != "" {
		minText += " " + cl.units
		maxText += " " + cl.units
	}

	fyne.Do(func() {
		cl.minLabel.SetText(minText)
		cl.maxLabel.SetText(maxText)
		cl.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (cl *ColorLegend) CreateRenderer() fyne.WidgetRenderer {
	cl.raster = canvas.NewRaster(cl.generateGradient)

	var content fyne.CanvasObject
	if cl.vertical {
		// Vertical layout: max label at top, gradient bar, min label at bottom
		cl.raster.SetMinSize(fyne.NewSize(40, 150))
		content = container.NewVBox(
			cl.maxLabel,
			container.NewPadded(cl.raster),
			cl.minLabel,
		)
	} else {
		// Horizontal layout: min label, gradient bar, max label
		cl.raster.SetMinSize(fyne.NewSize(150, 30))
		content = container.NewHBox(
			cl.minLabel,
			container.NewPadded(cl.raster),
			cl.maxLabel,
		)
	}

	return widget.NewSimpleRenderer(content)
}

// MinSize returns the minimum size of the legend.
func (cl *ColorLegend) MinSize() fyne.Size {
	if cl.vertical {
		return fyne.NewSize(80, 200)
	}
	return fyne.NewSize(250, 60)
}

// generateGradient creates the gradient bar image.
func (cl *ColorLegend) generateGradient(w, h int) image.Image {
	if w < 2 || h < 2 {
		w = 20
		h = 100
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	if cl.vertical {
		// Vertical gradient: top = max (t=1), bottom = min (t=0)
		for y := 0; y < h; y++ {
			t := 1.0 - float64(y)/float64(h-1) // Inverted: top is high
			c := cl.colorFunc(t)
			for x := 0; x < w; x++ {
				img.Set(x, y, c)
			}
		}
	} else {
		// Horizontal gradient: left = min (t=0), right = max (t=1)
		for x := 0; x < w; x++ {
			t := float64(x) / float64(w-1)
			c := cl.colorFunc(t)
			for y := 0; y < h; y++ {
				img.Set(x, y, c)
			}
		}
	}

	return img
}

// Standard colormap functions for convenience

// ViridisColor returns a Viridis colormap color for normalized value t [0,1].
func ViridisColor(t float64) color.RGBA {
	r := 0.267 + t*(0.993*t-0.068)
	g := 0.005 + t*(0.991-0.149*t)
	b := 0.329 + t*(0.288-0.147*t)
	return color.RGBA{
		R: uint8(clampFloat(r, 0, 1) * 255),
		G: uint8(clampFloat(g, 0, 1) * 255),
		B: uint8(clampFloat(b, 0, 1) * 255),
		A: 255,
	}
}

// BlueWhiteRedColor returns a diverging blue-white-red color for t [0,1].
func BlueWhiteRedColor(t float64) color.RGBA {
	if t < 0.5 {
		// Blue to white
		s := t * 2
		return color.RGBA{
			R: uint8(s * 255),
			G: uint8(s * 255),
			B: 255,
			A: 255,
		}
	}
	// White to red
	s := (t - 0.5) * 2
	return color.RGBA{
		R: 255,
		G: uint8((1 - s) * 255),
		B: uint8((1 - s) * 255),
		A: 255,
	}
}

// GreenToRedColor returns a green-to-red color for t [0,1].
func GreenToRedColor(t float64) color.RGBA {
	r := uint8(t * 255)
	g := uint8((1 - t) * 200)
	b := uint8(50)
	return color.RGBA{r, g, b, 255}
}

// BlueToYellowColor returns a blue-to-yellow color for t [0,1].
func BlueToYellowColor(t float64) color.RGBA {
	r := uint8(t * 255)
	g := uint8(t * 200)
	b := uint8((1 - t) * 150)
	return color.RGBA{r, g, b, 255}
}

// ErrorColor returns a black-yellow-red color for error visualization t [0,1].
func ErrorColor(t float64) color.RGBA {
	if t < 0.5 {
		// Black to yellow
		s := t * 2
		return color.RGBA{
			R: uint8(s * 255),
			G: uint8(s * 200),
			B: 0,
			A: 255,
		}
	}
	// Yellow to red
	s := (t - 0.5) * 2
	return color.RGBA{
		R: 255,
		G: uint8((1 - s) * 200),
		B: 0,
		A: 255,
	}
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
