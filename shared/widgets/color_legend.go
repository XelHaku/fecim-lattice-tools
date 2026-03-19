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

	"fecim-lattice-tools/shared/mathutil"
)

// ColorLegend displays a gradient bar with min/max labels and optional units.
type ColorLegend struct {
	widget.BaseWidget

	// Display properties
	minValue float64
	maxValue float64
	units    string
	vertical bool

	// Colormap support
	colormapName string
	colorFunc    func(float64) color.RGBA

	// Visual components
	raster   *canvas.Raster
	minLabel *widget.Label
	maxLabel *widget.Label

	// Vertical layout text labels (need separate reference for updates)
	verticalMinText *canvas.Text
	verticalMaxText *canvas.Text
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

	// Create labels with numeric values
	cl.minLabel = widget.NewLabel(cl.formatLabel(minValue))
	cl.minLabel.TextStyle = fyne.TextStyle{Monospace: true}
	cl.minLabel.Alignment = fyne.TextAlignCenter
	cl.maxLabel = widget.NewLabel(cl.formatLabel(maxValue))
	cl.maxLabel.TextStyle = fyne.TextStyle{Monospace: true}
	cl.maxLabel.Alignment = fyne.TextAlignCenter

	cl.ExtendBaseWidget(cl)
	return cl
}

// NewColorLegendWithColormap creates a ColorLegend using a named colormap.
func NewColorLegendWithColormap(minValue, maxValue float64, units string, vertical bool, colormapName string) *ColorLegend {
	colorFunc := GetColormapFunc(colormapName)
	cl := NewColorLegend(minValue, maxValue, units, vertical, colorFunc)
	cl.colormapName = colormapName
	return cl
}

// formatLabel formats a numeric value for display.
func (cl *ColorLegend) formatLabel(value float64) string {
	if cl.units != "" {
		// Check if value is integer-like
		if value == float64(int(value)) {
			return fmt.Sprintf("%d %s", int(value), cl.units)
		}
		return fmt.Sprintf("%.2f %s", value, cl.units)
	}
	if value == float64(int(value)) {
		return fmt.Sprintf("%d", int(value))
	}
	return fmt.Sprintf("%.2f", value)
}

// SetRange updates the min/max values and refreshes the legend.
func (cl *ColorLegend) SetRange(minValue, maxValue float64) {
	cl.minValue = minValue
	cl.maxValue = maxValue

	fyne.Do(func() {
		// Update horizontal layout labels
		cl.minLabel.SetText(cl.formatLabel(minValue))
		cl.maxLabel.SetText(cl.formatLabel(maxValue))

		// Update vertical layout text labels if they exist
		if cl.verticalMinText != nil {
			cl.verticalMinText.Text = cl.formatLabel(minValue)
			cl.verticalMinText.Refresh()
		}
		if cl.verticalMaxText != nil {
			cl.verticalMaxText.Text = cl.formatLabel(maxValue)
			cl.verticalMaxText.Refresh()
		}

		cl.Refresh()
	})
}

// SetColormap changes the colormap by name.
func (cl *ColorLegend) SetColormap(name string) {
	cl.colormapName = name
	cl.colorFunc = GetColormapFunc(name)
	if IsStartupStabilizing() {
		return
	}
	fyne.Do(func() {
		cl.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (cl *ColorLegend) CreateRenderer() fyne.WidgetRenderer {
	cl.raster = canvas.NewRaster(cl.generateGradient)

	var content fyne.CanvasObject
	if cl.vertical {
		// Vertical layout with numeric labels overlaid on gradient
		cl.raster.SetMinSize(fyne.NewSize(60, 180))

		// Position max label at top, min label at bottom
		// Store references so SetRange can update them
		cl.verticalMaxText = canvas.NewText(cl.formatLabel(cl.maxValue), color.White)
		cl.verticalMaxText.TextSize = 14
		cl.verticalMaxText.Alignment = fyne.TextAlignLeading

		cl.verticalMinText = canvas.NewText(cl.formatLabel(cl.minValue), color.White)
		cl.verticalMinText.TextSize = 14
		cl.verticalMinText.Alignment = fyne.TextAlignLeading

		// Local references for layout
		maxLabelText := cl.verticalMaxText
		minLabelText := cl.verticalMinText

		// Create intermediate labels if widget is tall enough
		// Labels at 0, 10, 20, max (assuming range is roughly 0-30 for FeCIM)
		intermediateLabels := []fyne.CanvasObject{}

		// Only add intermediate labels if height > 100px
		if cl.raster.MinSize().Height > 100 {
			// Add labels at reasonable intervals
			labelValues := []float64{}

			// Always include 0, 10, 20 if they're in range
			for labelVal := 0.0; labelVal <= cl.maxValue; labelVal += 10 {
				if labelVal >= cl.minValue && labelVal <= cl.maxValue {
					labelValues = append(labelValues, labelVal)
				}
			}

			// Ensure max is included
			if len(labelValues) == 0 || labelValues[len(labelValues)-1] != cl.maxValue {
				labelValues = append(labelValues, cl.maxValue)
			}

			// Create text objects for intermediate labels
			for _, labelVal := range labelValues {
				if labelVal == cl.minValue || labelVal == cl.maxValue {
					continue // Skip min/max as they're handled separately
				}

				labelText := canvas.NewText(cl.formatLabel(labelVal), color.White)
				labelText.TextSize = 14
				labelText.Alignment = fyne.TextAlignLeading
				intermediateLabels = append(intermediateLabels, labelText)
			}
		}

		// Use BorderLayout to position labels
		if len(intermediateLabels) > 0 {
			// With intermediate labels, use custom container
			labelContainer := container.NewVBox(append([]fyne.CanvasObject{
				container.NewHBox(maxLabelText),
			}, append(intermediateLabels, container.NewHBox(minLabelText))...)...)

			content = container.NewBorder(
				nil, nil,
				labelContainer, nil,
				cl.raster,
			)
		} else {
			// Original layout without intermediate labels
			content = container.NewBorder(
				container.NewHBox(maxLabelText),
				container.NewHBox(minLabelText),
				nil, nil,
				cl.raster,
			)
		}
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
		R: uint8(mathutil.Clamp(r, 0, 1) * 255),
		G: uint8(mathutil.Clamp(g, 0, 1) * 255),
		B: uint8(mathutil.Clamp(b, 0, 1) * 255),
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
// Uses minimum brightness values for visibility on dark theme backgrounds.
func BlueToYellowColor(t float64) color.RGBA {
	r := uint8(50 + t*205)  // 50-255: visible at all levels
	g := uint8(80 + t*140)  // 80-220: visible at all levels
	b := uint8(200 - t*150) // 200-50: blue fades as yellow increases
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


// GetColormapFunc returns the colormap function for a given name.
func GetColormapFunc(name string) func(float64) color.RGBA {
	switch name {
	case "viridis":
		return ViridisColor
	case "plasma":
		return PlasmaColor
	case "coolwarm":
		return BlueWhiteRedColor
	case "fecim":
		return FeCIMColor
	case "diverging":
		return BlueWhiteRedColor
	default:
		return ViridisColor
	}
}

// PlasmaColor returns a Plasma colormap color for normalized value t [0,1].
func PlasmaColor(t float64) color.RGBA {
	r := 0.05 + t*0.89
	g := 0.03 + t*0.95*t
	b := 0.53 - t*0.40

	return color.RGBA{
		R: uint8(mathutil.Clamp(r, 0, 1) * 255),
		G: uint8(mathutil.Clamp(g, 0, 1) * 255),
		B: uint8(mathutil.Clamp(b, 0, 1) * 255),
		A: 255,
	}
}

// FeCIMColor returns a 30-level inspired colormap for normalized value t [0,1].
func FeCIMColor(t float64) color.RGBA {
	if t < 0.2 {
		s := t * 5
		return color.RGBA{
			R: uint8(60 + s*20),
			G: uint8(s * 100),
			B: uint8(120 + s*80),
			A: 255,
		}
	} else if t < 0.4 {
		s := (t - 0.2) * 5
		return color.RGBA{
			R: uint8(80 - s*50),
			G: uint8(100 + s*155),
			B: uint8(200 - s*50),
			A: 255,
		}
	} else if t < 0.6 {
		s := (t - 0.4) * 5
		return color.RGBA{
			R: uint8(30 + s*180),
			G: uint8(255 - s*55),
			B: uint8(150 - s*50),
			A: 255,
		}
	} else if t < 0.8 {
		s := (t - 0.6) * 5
		return color.RGBA{
			R: uint8(210 + s*45),
			G: uint8(200 - s*100),
			B: uint8(100 - s*60),
			A: 255,
		}
	}
	s := (t - 0.8) * 5
	return color.RGBA{
		R: uint8(255),
		G: uint8(100 - s*50),
		B: uint8(40 + s*20),
		A: 255,
	}
}
