// Package gui provides Fyne-based GUI components for MNIST visualization.
// quantization_widget.go implements P1.1: Real-Time Quantization Visualization
package gui

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/mathutil"
)

// QuantizationSample represents a single weight before and after quantization.
type QuantizationSample struct {
	OriginalValue  float64 // FP32 value
	QuantizedValue float64 // Value after 30-level quantization
	Level          int     // Discrete level (0-29)
	Error          float64 // Absolute quantization error
}

// QuantizationWidget visualizes the N-level quantization process in real-time.
// This is the key educational widget showing why FeCIM's discrete levels matter.
type QuantizationWidget struct {
	widget.BaseWidget

	mu           sync.RWMutex
	samples      []QuantizationSample // Current weight samples being displayed
	numLevels    int                  // Current quantization levels (2-30)
	totalError   float64              // Average precision loss
	animProgress float32              // Animation progress 0.0-1.0

	// Visual components
	titleLabel *widget.Label
	infoLabel  *widget.Label
	raster     *canvas.Raster
}

// NewQuantizationWidget creates a new quantization visualization widget.
func NewQuantizationWidget() *QuantizationWidget {
	qw := &QuantizationWidget{
		samples:      make([]QuantizationSample, 0, 5),
		numLevels:    30,
		animProgress: 1.0,
	}
	qw.titleLabel = widget.NewLabelWithStyle("Weight Quantization", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	qw.infoLabel = widget.NewLabel("Draw a digit to see quantization")
	qw.ExtendBaseWidget(qw)
	return qw
}

// SetNumLevels updates the quantization levels (affects visualization).
func (qw *QuantizationWidget) SetNumLevels(levels int) {
	qw.mu.Lock()
	qw.numLevels = levels
	qw.mu.Unlock()
	fyne.Do(func() {
		qw.Refresh()
	})
}

// UpdateWithWeights samples random weights and shows their quantization.
func (qw *QuantizationWidget) UpdateWithWeights(weights [][]float64, count int) {
	if len(weights) == 0 || len(weights[0]) == 0 {
		return
	}

	qw.mu.Lock()
	defer qw.mu.Unlock()

	levels := qw.numLevels
	if levels < 2 {
		levels = 2
	}

	// Determine max magnitude (same as core.QuantizeWeights)
	wMax := 0.0
	for i := range weights {
		for j := range weights[i] {
			if abs := math.Abs(weights[i][j]); abs > wMax {
				wMax = abs
			}
		}
	}

	// Sample random weights
	rows := len(weights)
	cols := len(weights[0])
	qw.samples = qw.samples[:0]

	// Use deterministic sampling for reproducibility
	step := (rows * cols) / (count + 1)
	if step < 1 {
		step = 1
	}

	totalError := 0.0
	for i := 0; i < count && i*step < rows*cols; i++ {
		idx := (i + 1) * step
		row := idx / cols
		col := idx % cols

		if row >= rows {
			break
		}

		original := weights[row][col]

		// Quantize using the same symmetric mapping as core.QuantizeWeights
		level := 0
		quantizedOriginal := 0.0
		if wMax > 0 {
			normalized := (original + wMax) / (2.0 * wMax)
			normalized = mathutil.Clamp01(normalized)
			bin := int(math.Round(normalized * float64(levels-1)))
			if bin < 0 {
				bin = 0
			}
			if bin >= levels {
				bin = levels - 1
			}
			level = bin
			levelStep := 2.0 * wMax / float64(levels-1)
			quantizedOriginal = -wMax + float64(bin)*levelStep
		}

		absError := math.Abs(original - quantizedOriginal)
		relError := 0.0
		if math.Abs(original) > 0.001 {
			relError = absError / math.Abs(original)
		}

		qw.samples = append(qw.samples, QuantizationSample{
			OriginalValue:  original,
			QuantizedValue: quantizedOriginal,
			Level:          level,
			Error:          relError,
		})

		totalError += relError
	}

	if len(qw.samples) > 0 {
		qw.totalError = totalError / float64(len(qw.samples))
	}

	// Update info label
	impactText := "OK"
	if qw.totalError > 0.05 {
		impactText = "Noticeable"
	}
	if qw.totalError > 0.10 {
		impactText = "High"
	}

	qw.infoLabel.SetText(fmt.Sprintf("Precision loss: %.2f%% | Impact: %s | Levels: %d",
		qw.totalError*100, impactText, qw.numLevels))

	fyne.Do(func() {
		qw.Refresh()
	})
}

// Clear resets the widget to idle state.
func (qw *QuantizationWidget) Clear() {
	qw.mu.Lock()
	qw.samples = qw.samples[:0]
	qw.totalError = 0
	qw.mu.Unlock()

	qw.infoLabel.SetText("Draw a digit to see quantization")
	fyne.Do(func() {
		qw.Refresh()
	})
}

// MinSize returns the minimum size for the widget.
func (qw *QuantizationWidget) MinSize() fyne.Size {
	return fyne.NewSize(260, 140)
}

// CreateRenderer implements fyne.Widget.
func (qw *QuantizationWidget) CreateRenderer() fyne.WidgetRenderer {
	qw.raster = canvas.NewRaster(qw.generateImage)

	content := container.NewBorder(
		container.NewVBox(
			qw.titleLabel,
			widget.NewSeparator(),
		),
		qw.infoLabel,
		nil, nil,
		container.NewMax(qw.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the quantization visualization.
func (qw *QuantizationWidget) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 350
	}
	if h < 10 {
		h = 120
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 30, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	qw.mu.RLock()
	samples := qw.samples
	numLevels := qw.numLevels
	qw.mu.RUnlock()

	if len(samples) == 0 {
		// Draw idle state message
		qw.drawText(img, "Ready - Draw to see quantization", w/2-100, h/2-8, color.RGBA{100, 100, 120, 255})
		return img
	}

	// Layout: each sample gets a row
	rowHeight := (h - 20) / len(samples)
	if rowHeight > 35 {
		rowHeight = 35
	}

	padding := 10
	leftCol := padding
	midCol := w / 2
	rightCol := w - padding - 100

	// Header
	headerY := 5
	qw.drawText(img, "FP (Float32)", leftCol, headerY, color.RGBA{150, 150, 170, 255})
	qw.drawText(img, "->", midCol-10, headerY, color.RGBA{0, 200, 255, 255})
	qw.drawText(img, fmt.Sprintf("%d Levels", numLevels), midCol+10, headerY, color.RGBA{150, 150, 170, 255})

	// Draw each sample
	for i, sample := range samples {
		y := 20 + i*rowHeight

		// Original value (left side)
		origText := fmt.Sprintf("%.6f", sample.OriginalValue)
		qw.drawText(img, origText, leftCol, y, color.RGBA{200, 200, 220, 255})

		// Arrow
		arrowX := midCol - 15
		arrowColor := color.RGBA{0, 180, 220, 255}
		for ax := arrowX; ax < arrowX+30; ax++ {
			img.Set(ax, y+6, arrowColor)
		}
		// Arrowhead
		img.Set(arrowX+28, y+5, arrowColor)
		img.Set(arrowX+28, y+7, arrowColor)
		img.Set(arrowX+27, y+4, arrowColor)
		img.Set(arrowX+27, y+8, arrowColor)

		// Quantized value (right side)
		quantText := fmt.Sprintf("%.4f (L%d)", sample.QuantizedValue, sample.Level)
		qw.drawText(img, quantText, midCol+20, y, color.RGBA{0, 220, 180, 255})

		// Level indicator (visual bar showing position in 0-29 range)
		barX := rightCol
		barWidth := 80
		barHeight := 8
		barY := y + 2

		// Background bar
		for bx := barX; bx < barX+barWidth; bx++ {
			for by := barY; by < barY+barHeight; by++ {
				img.Set(bx, by, color.RGBA{50, 50, 70, 255})
			}
		}

		// Level marker (filled portion + marker)
		fillWidth := int(float64(barWidth) * float64(sample.Level) / float64(numLevels-1))
		for bx := barX; bx < barX+fillWidth; bx++ {
			for by := barY; by < barY+barHeight; by++ {
				img.Set(bx, by, color.RGBA{0, 180, 150, 200})
			}
		}

		// Marker at current level
		markerX := barX + fillWidth
		for by := barY - 1; by < barY+barHeight+1; by++ {
			img.Set(markerX, by, color.RGBA{255, 255, 255, 255})
		}

		// Draw level ticks
		for l := 0; l < numLevels; l += 5 {
			tickX := barX + int(float64(barWidth)*float64(l)/float64(numLevels-1))
			img.Set(tickX, barY+barHeight, color.RGBA{100, 100, 120, 255})
		}
	}

	return img
}

// drawText draws simple text on the image (basic bitmap font simulation).
func (qw *QuantizationWidget) drawText(img *image.RGBA, text string, x, y int, c color.RGBA) {
	// Simple character drawing - each char is approximately 7 pixels wide
	charWidth := 7
	for i, ch := range text {
		cx := x + i*charWidth
		qw.drawChar(img, ch, cx, y, c)
	}
}

// drawChar draws a single character (simplified bitmap approach).
func (qw *QuantizationWidget) drawChar(img *image.RGBA, ch rune, x, y int, c color.RGBA) {
	// Basic 5x7 font patterns for digits and common chars
	patterns := map[rune][]string{
		'0': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
		'1': {"00100", "01100", "00100", "00100", "00100", "00100", "01110"},
		'2': {"01110", "10001", "00001", "00110", "01000", "10000", "11111"},
		'3': {"01110", "10001", "00001", "00110", "00001", "10001", "01110"},
		'4': {"00010", "00110", "01010", "10010", "11111", "00010", "00010"},
		'5': {"11111", "10000", "11110", "00001", "00001", "10001", "01110"},
		'6': {"01110", "10000", "10000", "11110", "10001", "10001", "01110"},
		'7': {"11111", "00001", "00010", "00100", "01000", "01000", "01000"},
		'8': {"01110", "10001", "10001", "01110", "10001", "10001", "01110"},
		'9': {"01110", "10001", "10001", "01111", "00001", "00001", "01110"},
		'.': {"00000", "00000", "00000", "00000", "00000", "01100", "01100"},
		'-': {"00000", "00000", "00000", "11111", "00000", "00000", "00000"},
		':': {"00000", "01100", "01100", "00000", "01100", "01100", "00000"},
		'%': {"11001", "11010", "00100", "01000", "01011", "10011", "00000"},
		' ': {"00000", "00000", "00000", "00000", "00000", "00000", "00000"},
		'(': {"00010", "00100", "01000", "01000", "01000", "00100", "00010"},
		')': {"01000", "00100", "00010", "00010", "00010", "00100", "01000"},
		'>': {"10000", "01000", "00100", "00010", "00100", "01000", "10000"},
		'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
		'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
		'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
		'W': {"10001", "10001", "10001", "10101", "10101", "10101", "01010"},
		'a': {"00000", "00000", "01110", "00001", "01111", "10001", "01111"},
		'i': {"00100", "00000", "01100", "00100", "00100", "00100", "01110"},
		't': {"00100", "00100", "01110", "00100", "00100", "00100", "00011"},
		'n': {"00000", "00000", "10110", "11001", "10001", "10001", "10001"},
		'g': {"00000", "00000", "01111", "10001", "01111", "00001", "01110"},
		'f': {"00110", "01000", "01000", "11100", "01000", "01000", "01000"},
		'o': {"00000", "00000", "01110", "10001", "10001", "10001", "01110"},
		'r': {"00000", "00000", "10110", "11001", "10000", "10000", "10000"},
		'e': {"00000", "00000", "01110", "10001", "11111", "10000", "01110"},
		'c': {"00000", "00000", "01110", "10000", "10000", "10001", "01110"},
		's': {"00000", "00000", "01110", "10000", "01110", "00001", "11110"},
		'l': {"01100", "00100", "00100", "00100", "00100", "00100", "01110"},
		'v': {"00000", "00000", "10001", "10001", "10001", "01010", "00100"},
	}

	pattern, ok := patterns[ch]
	if !ok {
		// Default: draw a small rectangle for unknown chars
		for dy := 0; dy < 7; dy++ {
			for dx := 0; dx < 5; dx++ {
				if dx == 0 || dx == 4 || dy == 0 || dy == 6 {
					if x+dx >= 0 && x+dx < img.Bounds().Dx() && y+dy >= 0 && y+dy < img.Bounds().Dy() {
						img.Set(x+dx, y+dy, c)
					}
				}
			}
		}
		return
	}

	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == '1' {
				px := x + dx
				py := y + dy
				if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
					img.Set(px, py, c)
				}
			}
		}
	}
}
