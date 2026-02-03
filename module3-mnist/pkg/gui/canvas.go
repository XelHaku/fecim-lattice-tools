// Package gui provides Fyne-based GUI components for MNIST visualization.
package gui

import (
	"image"
	"image/color"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
)

// BrushSize represents predefined brush sizes.
type BrushSize int

const (
	BrushThin   BrushSize = iota // 1.0 radius
	BrushMedium                  // 1.5 radius
	BrushThick                   // 2.5 radius
)

// DigitCanvas provides an interactive 28x28 drawing area for MNIST digits.
type DigitCanvas struct {
	widget.BaseWidget

	// 28x28 pixel values (0.0 to 1.0)
	pixels [28][28]float64

	// Drawing state
	isDrawing   bool
	brushRadius float64
	brushSize   BrushSize

	// Rendering
	raster *canvas.Raster

	// Pixel count tracking
	activePixels int // Count of pixels with value > 0.1

	// Input tracking
	lastInputSource InputSource

	// Callback when digit changes
	OnDigitChanged func(pixels []float64)

	// Callback for pixel count updates
	OnPixelCountChanged func(count int, total int)
}

// NewDigitCanvas creates a new digit drawing canvas.
func NewDigitCanvas() *DigitCanvas {
	dc := &DigitCanvas{
		brushRadius:     1.5,
		brushSize:       BrushMedium,
		lastInputSource: InputProgrammatic,
	}
	dc.ExtendBaseWidget(dc)
	return dc
}

// SetBrushSize changes the brush size.
func (dc *DigitCanvas) SetBrushSize(size BrushSize) {
	dc.brushSize = size
	switch size {
	case BrushThin:
		dc.brushRadius = 1.0
	case BrushMedium:
		dc.brushRadius = 1.5
	case BrushThick:
		dc.brushRadius = 2.5
	}
}

// GetBrushSize returns the current brush size.
func (dc *DigitCanvas) GetBrushSize() BrushSize {
	return dc.brushSize
}

// GetActivePixelCount returns the number of active (drawn) pixels.
func (dc *DigitCanvas) GetActivePixelCount() int {
	return dc.activePixels
}

// updatePixelCount recalculates the active pixel count.
func (dc *DigitCanvas) updatePixelCount() {
	count := 0
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			if dc.pixels[i][j] > 0.1 {
				count++
			}
		}
	}
	dc.activePixels = count
	if dc.OnPixelCountChanged != nil {
		dc.OnPixelCountChanged(count, 784)
	}
}

// Clear resets the canvas to blank.
// Note: fyne.Do() is used defensively - it's a no-op when already on the main thread
// (e.g., from button callbacks), but ensures safety if called from other contexts.
func (dc *DigitCanvas) Clear() {
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			dc.pixels[i][j] = 0
		}
	}
	dc.lastInputSource = InputProgrammatic
	dc.activePixels = 0
	fyne.Do(func() {
		dc.Refresh()
	})
	dc.updatePixelCount()
	dc.notifyChange()
}

// GetPixels returns the 784-element flattened pixel array.
func (dc *DigitCanvas) GetPixels() []float64 {
	result := make([]float64, 784)
	for i := 0; i < 28; i++ {
		for j := 0; j < 28; j++ {
			result[i*28+j] = dc.pixels[i][j]
		}
	}
	return result
}

// SetPixels sets the canvas from a 784-element array.
// Note: fyne.Do() is used defensively - it's a no-op when already on the main thread
// (e.g., from button callbacks), but ensures safety if called from goroutines.
func (dc *DigitCanvas) SetPixels(pixels []float64) {
	for i := 0; i < 28 && i*28 < len(pixels); i++ {
		for j := 0; j < 28 && i*28+j < len(pixels); j++ {
			dc.pixels[i][j] = pixels[i*28+j]
		}
	}
	dc.lastInputSource = InputProgrammatic
	dc.updatePixelCount()
	fyne.Do(func() {
		dc.Refresh()
	})
}

// CreateRenderer implements fyne.Widget.
func (dc *DigitCanvas) CreateRenderer() fyne.WidgetRenderer {
	dc.raster = canvas.NewRaster(dc.generateImage)
	dc.raster.ScaleMode = canvas.ImageScalePixels

	return widget.NewSimpleRenderer(dc.raster)
}

// MinSize returns the minimum size for the digit canvas.
// Enlarged to 350x350 for better drawing experience (increased vertical space).
func (dc *DigitCanvas) MinSize() fyne.Size {
	return fyne.NewSize(350, 350) // Larger canvas for easier digit drawing
}

// generateImage creates the canvas image.
func (dc *DigitCanvas) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{20, 20, 30, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	// Calculate cell size
	cellW := float64(w) / 28.0
	cellH := float64(h) / 28.0

	// Check if canvas is empty
	isEmpty := true
	for py := 0; py < 28 && isEmpty; py++ {
		for px := 0; px < 28; px++ {
			if dc.pixels[py][px] > 0.01 {
				isEmpty = false
				break
			}
		}
	}

	// Draw "Draw here" hint when empty
	if isEmpty && w > 50 && h > 50 {
		hintColor := color.RGBA{60, 80, 100, 255}
		// Draw a simple hand-draw icon (circle with finger pointer suggestion)
		cx, cy := w/2, h/2
		radius := w / 6
		if h/6 < radius {
			radius = h / 6
		}
		// Draw dashed circle hint
		for angle := 0.0; angle < 360; angle += 15 {
			rad := angle * math.Pi / 180
			x := cx + int(float64(radius)*math.Cos(rad))
			y := cy + int(float64(radius)*math.Sin(rad))
			if x >= 0 && x < w && y >= 0 && y < h {
				img.Set(x, y, hintColor)
				if x+1 < w {
					img.Set(x+1, y, hintColor)
				}
				if y+1 < h {
					img.Set(x, y+1, hintColor)
				}
			}
		}
		// Draw small dot at center to suggest "draw here"
		dotSize := 4
		for dy := -dotSize; dy <= dotSize; dy++ {
			for dx := -dotSize; dx <= dotSize; dx++ {
				if dx*dx+dy*dy <= dotSize*dotSize {
					px, py := cx+dx, cy+dy
					if px >= 0 && px < w && py >= 0 && py < h {
						img.Set(px, py, hintColor)
					}
				}
			}
		}
	}

	// Draw pixels
	for py := 0; py < 28; py++ {
		for px := 0; px < 28; px++ {
			value := dc.pixels[py][px]
			if value > 0 {
				// Use FeCIM cyan-to-white gradient
				intensity := uint8(value * 255)
				c := color.RGBA{
					R: uint8(float64(intensity) * 0.7),
					G: intensity,
					B: intensity,
					A: 255,
				}

				// Fill cell
				x0 := int(float64(px) * cellW)
				x1 := int(float64(px+1) * cellW)
				y0 := int(float64(py) * cellH)
				y1 := int(float64(py+1) * cellH)

				for y := y0; y < y1; y++ {
					for x := x0; x < x1; x++ {
						if x >= 0 && x < w && y >= 0 && y < h {
							img.Set(x, y, c)
						}
					}
				}
			}
		}
	}

	// Draw grid lines (subtle)
	gridColor := color.RGBA{40, 40, 50, 255}
	for i := 0; i <= 28; i++ {
		x := int(float64(i) * cellW)
		y := int(float64(i) * cellH)
		for j := 0; j < h; j++ {
			if x >= 0 && x < w {
				img.Set(x, j, gridColor)
			}
		}
		for j := 0; j < w; j++ {
			if y >= 0 && y < h {
				img.Set(j, y, gridColor)
			}
		}
	}

	return img
}

// Tapped handles tap events for drawing.
func (dc *DigitCanvas) Tapped(e *fyne.PointEvent) {
	dc.draw(e.Position)
}

// TappedSecondary clears the canvas on right-click.
func (dc *DigitCanvas) TappedSecondary(e *fyne.PointEvent) {
	dc.Clear()
}

// Dragged handles drag events for continuous drawing.
func (dc *DigitCanvas) Dragged(e *fyne.DragEvent) {
	dc.draw(e.Position)
}

// DragEnd handles end of drag.
func (dc *DigitCanvas) DragEnd() {
	dc.isDrawing = false
}

// MouseDown implements desktop.Mouseable.
func (dc *DigitCanvas) MouseDown(e *desktop.MouseEvent) {
	dc.isDrawing = true
	dc.draw(e.Position)
}

// MouseUp implements desktop.Mouseable.
func (dc *DigitCanvas) MouseUp(e *desktop.MouseEvent) {
	dc.isDrawing = false
}

// draw paints pixels at the given position.
// Note: This is always called from Fyne event handlers (Tapped, Dragged, MouseDown),
// which are dispatched on the main thread, so no fyne.Do() wrapper is needed.
func (dc *DigitCanvas) draw(pos fyne.Position) {
	dc.lastInputSource = InputUser
	size := dc.Size()
	cellW := size.Width / 28.0
	cellH := size.Height / 28.0

	// Convert to pixel coordinates
	px := int(pos.X / cellW)
	py := int(pos.Y / cellH)

	// Apply brush with soft falloff
	radius := dc.brushRadius
	for dy := -2; dy <= 2; dy++ {
		for dx := -2; dx <= 2; dx++ {
			nx := px + dx
			ny := py + dy

			if nx >= 0 && nx < 28 && ny >= 0 && ny < 28 {
				dist := math.Sqrt(float64(dx*dx + dy*dy))
				if dist <= radius {
					// Soft brush with falloff
					intensity := 1.0 - (dist / (radius + 0.5))
					if intensity > dc.pixels[ny][nx] {
						dc.pixels[ny][nx] = clamp(intensity, 0, 1)
					}
				}
			}
		}
	}

	dc.Refresh()
	dc.updatePixelCount()
	dc.notifyChange()
}

// notifyChange calls the change callback if set.
func (dc *DigitCanvas) notifyChange() {
	if dc.OnDigitChanged != nil {
		dc.OnDigitChanged(dc.GetPixels())
	}
}

// LastInputSource returns the most recent input source (user vs programmatic).
func (dc *DigitCanvas) LastInputSource() InputSource {
	return dc.lastInputSource
}

// clamp restricts a value to a range.
func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Ensure interfaces are implemented
var _ fyne.Tappable = (*DigitCanvas)(nil)
var _ fyne.SecondaryTappable = (*DigitCanvas)(nil)
var _ fyne.Draggable = (*DigitCanvas)(nil)
var _ desktop.Mouseable = (*DigitCanvas)(nil)
