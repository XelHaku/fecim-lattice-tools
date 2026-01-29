// Package gui provides Fyne-based GUI components for MNIST visualization.
// energy_widget.go implements P1.3: Energy Efficiency Live Visualization
package gui

import (
	"fmt"
	"image"
	"image/color"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
)

// EnergyConfig defines energy parameters for calculations.
type EnergyConfig struct {
	// Per-operation energy costs (in femtojoules)
	EnergyPerMAC_FeCIM float64 // ~50 fJ per MAC for FeCIM
	EnergyPerMAC_GPU   float64 // ~2000 fJ per MAC for GPU (data movement dominates)
	EnergyPerDAC       float64 // ~10 pJ per DAC conversion
	EnergyPerADC       float64 // ~20 pJ per ADC conversion (6-bit)
}

// DefaultEnergyConfig returns energy parameters based on FeCIM research.
func DefaultEnergyConfig() EnergyConfig {
	return EnergyConfig{
		EnergyPerMAC_FeCIM: 50e-15,   // 50 femtojoules
		EnergyPerMAC_GPU:   2000e-15, // 2 picojoules (due to data movement)
		EnergyPerDAC:       10e-12,   // 10 picojoules
		EnergyPerADC:       20e-12,   // 20 picojoules
	}
}

// EnergyWidget visualizes energy efficiency comparison between FeCIM and GPU.
// Shows per-inference energy and running total for the session.
type EnergyWidget struct {
	widget.BaseWidget

	mu sync.RWMutex

	// Configuration
	config EnergyConfig

	// Network architecture
	inputSize  int
	hiddenSize int
	outputSize int

	// Running totals
	totalInferences  int
	totalEnergyFeCIM float64 // in Joules
	totalEnergyGPU   float64 // in Joules

	// Last inference
	lastEnergyFeCIM float64 // in nanojoules
	lastEnergyGPU   float64 // in nanojoules
	efficiencyRatio float64

	// Visual components
	titleLabel *widget.Label
	statsLabel *widget.Label
	raster     *canvas.Raster
}

// NewEnergyWidget creates a new energy visualization widget.
func NewEnergyWidget(inputSize, hiddenSize, outputSize int) *EnergyWidget {
	ew := &EnergyWidget{
		config:     DefaultEnergyConfig(),
		inputSize:  inputSize,
		hiddenSize: hiddenSize,
		outputSize: outputSize,
	}
	ew.titleLabel = widget.NewLabelWithStyle("Energy Efficiency", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	ew.statsLabel = widget.NewLabel("Ready - Draw a digit to begin")
	ew.ExtendBaseWidget(ew)
	return ew
}

// SetNetworkSize updates the network architecture for energy calculations.
func (ew *EnergyWidget) SetNetworkSize(inputSize, hiddenSize, outputSize int) {
	ew.mu.Lock()
	ew.inputSize = inputSize
	ew.hiddenSize = hiddenSize
	ew.outputSize = outputSize
	ew.mu.Unlock()
}

// RecordInference records a new inference and updates energy totals.
func (ew *EnergyWidget) RecordInference() {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	ew.totalInferences++

	// Calculate MACs
	// Layer 1: inputSize × hiddenSize
	// Layer 2: hiddenSize × outputSize
	layer1MACs := ew.inputSize * ew.hiddenSize
	layer2MACs := ew.hiddenSize * ew.outputSize
	totalMACs := layer1MACs + layer2MACs

	// Calculate ADC/DAC counts
	numDACs := ew.inputSize                  // One DAC per input
	numADCs := ew.hiddenSize + ew.outputSize // ADCs at each layer output

	// FeCIM energy calculation
	// E = (MACs × energy_per_MAC) + (DACs × DAC_energy) + (ADCs × ADC_energy)
	fecimEnergy := float64(totalMACs)*ew.config.EnergyPerMAC_FeCIM +
		float64(numDACs)*ew.config.EnergyPerDAC +
		float64(numADCs)*ew.config.EnergyPerADC

	// GPU energy calculation (dominated by data movement)
	gpuEnergy := float64(totalMACs) * ew.config.EnergyPerMAC_GPU

	// Convert to nanojoules for display
	ew.lastEnergyFeCIM = fecimEnergy * 1e9 // Convert J to nJ
	ew.lastEnergyGPU = gpuEnergy * 1e9

	// Calculate efficiency ratio
	if fecimEnergy > 0 {
		ew.efficiencyRatio = gpuEnergy / fecimEnergy
	}

	// Update running totals (in Joules)
	ew.totalEnergyFeCIM += fecimEnergy
	ew.totalEnergyGPU += gpuEnergy

	// Update stats label
	ew.updateStatsLabel()
	fyne.Do(func() {
		ew.Refresh()
	})
}

func (ew *EnergyWidget) updateStatsLabel() {
	// Convert totals to appropriate units using shared physics package
	fecimStr := physics.FormatEnergy(ew.totalEnergyFeCIM)
	gpuStr := physics.FormatEnergy(ew.totalEnergyGPU)

	ew.statsLabel.SetText(fmt.Sprintf("Session: %d inferences | FeCIM: %s | GPU: %s | %.0fx savings",
		ew.totalInferences, fecimStr, gpuStr, ew.efficiencyRatio))
}

// Reset clears all counters.
func (ew *EnergyWidget) Reset() {
	ew.mu.Lock()
	ew.totalInferences = 0
	ew.totalEnergyFeCIM = 0
	ew.totalEnergyGPU = 0
	ew.lastEnergyFeCIM = 0
	ew.lastEnergyGPU = 0
	ew.efficiencyRatio = 0
	ew.mu.Unlock()

	ew.statsLabel.SetText("Ready - Draw a digit to begin")
	fyne.Do(func() {
		ew.Refresh()
	})
}

// GetEfficiencyRatio returns the current efficiency ratio (GPU/FeCIM).
func (ew *EnergyWidget) GetEfficiencyRatio() float64 {
	ew.mu.RLock()
	defer ew.mu.RUnlock()
	return ew.efficiencyRatio
}

// MinSize returns the minimum size for the widget.
func (ew *EnergyWidget) MinSize() fyne.Size {
	return fyne.NewSize(400, 200)
}

// CreateRenderer implements fyne.Widget.
func (ew *EnergyWidget) CreateRenderer() fyne.WidgetRenderer {
	ew.raster = canvas.NewRaster(ew.generateImage)

	content := container.NewBorder(
		container.NewVBox(
			ew.titleLabel,
			widget.NewSeparator(),
		),
		ew.statsLabel,
		nil, nil,
		container.NewMax(ew.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// generateImage creates the energy comparison visualization.
func (ew *EnergyWidget) generateImage(w, h int) image.Image {
	if w < 10 {
		w = 400
	}
	if h < 10 {
		h = 140
	}

	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 30, 45, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	ew.mu.RLock()
	fecimEnergy := ew.lastEnergyFeCIM
	gpuEnergy := ew.lastEnergyGPU
	ratio := ew.efficiencyRatio
	totalInf := ew.totalInferences
	ew.mu.RUnlock()

	padding := 15
	barHeight := 25
	labelWidth := 70

	// Section 1: Per-Inference Comparison (UI-022 fix: added clear labels with units)
	y := padding

	// Title
	drawSimpleText(img, "Per Inference Energy (Horowitz 2014 model):", padding, y, color.RGBA{180, 180, 200, 255})
	y += 15

	// GPU bar (full width represents GPU energy)
	maxBarWidth := w - 2*padding - labelWidth - 120
	gpuBarWidth := maxBarWidth
	if gpuBarWidth < 10 {
		gpuBarWidth = 10
	}

	// GPU label and bar with units
	drawSimpleText(img, "FP32:", padding, y+8, color.RGBA{255, 100, 100, 255})
	drawBar(img, padding+labelWidth, y, gpuBarWidth, barHeight, color.RGBA{200, 80, 80, 255})
	// UI-022 fix: Clear unit labeling
	gpuLabel := fmt.Sprintf("%.2f nJ", gpuEnergy)
	if gpuEnergy >= 1000 {
		gpuLabel = fmt.Sprintf("%.2f µJ", gpuEnergy/1000)
	}
	drawSimpleText(img, gpuLabel, padding+labelWidth+gpuBarWidth+5, y+8, color.RGBA{200, 150, 150, 255})
	y += barHeight + 5

	// FeCIM bar (proportionally smaller)
	fecimBarWidth := 1
	if ratio > 0 && gpuEnergy > 0 {
		fecimBarWidth = int(float64(gpuBarWidth) / ratio)
		if fecimBarWidth < 1 {
			fecimBarWidth = 1
		}
	}

	// FeCIM label and bar with units
	drawSimpleText(img, "FeCIM:", padding, y+8, color.RGBA{100, 255, 180, 255})
	drawBar(img, padding+labelWidth, y, fecimBarWidth, barHeight, color.RGBA{80, 200, 150, 255})
	// UI-022 fix: Clear unit labeling (nJ or pJ)
	fecimLabel := fmt.Sprintf("%.2f nJ", fecimEnergy)
	if fecimEnergy < 1 {
		fecimLabel = fmt.Sprintf("%.0f pJ", fecimEnergy*1000)
	}
	drawSimpleText(img, fecimLabel, padding+labelWidth+fecimBarWidth+5, y+8, color.RGBA{150, 200, 180, 255})
	y += barHeight + 10

	// Efficiency highlight box
	if ratio > 0 {
		boxX := w/2 - 80
		boxY := y
		boxW := 160
		boxH := 35

		// Box background
		for bx := boxX; bx < boxX+boxW; bx++ {
			for by := boxY; by < boxY+boxH; by++ {
				img.Set(bx, by, color.RGBA{40, 60, 80, 255})
			}
		}

		// Border
		borderColor := color.RGBA{0, 200, 255, 255}
		for bx := boxX; bx < boxX+boxW; bx++ {
			img.Set(bx, boxY, borderColor)
			img.Set(bx, boxY+boxH-1, borderColor)
		}
		for by := boxY; by < boxY+boxH; by++ {
			img.Set(boxX, by, borderColor)
			img.Set(boxX+boxW-1, by, borderColor)
		}

		// Efficiency text
		effText := fmt.Sprintf("%.0fx MORE EFFICIENT", ratio)
		drawSimpleText(img, effText, boxX+20, boxY+12, color.RGBA{0, 230, 255, 255})
		y += boxH + 5
	}

	// Section 2: Why CIM wins (bottom section)
	if y < h-30 {
		y = h - 35
		drawSimpleText(img, "Physics does the math! No data movement needed.", padding, y, color.RGBA{120, 120, 140, 255})
		y += 12

		if totalInf > 0 {
			drawSimpleText(img, fmt.Sprintf("Session: %d inferences tracked", totalInf), padding, y, color.RGBA{100, 100, 120, 255})
		}
	}

	return img
}

// drawBar draws a filled rectangle for the bar chart.
func drawBar(img *image.RGBA, x, y, w, h int, c color.RGBA) {
	// Main bar
	for bx := x; bx < x+w && bx < img.Bounds().Dx(); bx++ {
		for by := y; by < y+h && by < img.Bounds().Dy(); by++ {
			img.Set(bx, by, c)
		}
	}

	// Lighter top edge for 3D effect
	lighter := color.RGBA{
		R: uint8(min(255, int(c.R)+40)),
		G: uint8(min(255, int(c.G)+40)),
		B: uint8(min(255, int(c.B)+40)),
		A: c.A,
	}
	for bx := x; bx < x+w && bx < img.Bounds().Dx(); bx++ {
		img.Set(bx, y, lighter)
	}
}

// drawSimpleText draws text using a simple bitmap font.
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.RGBA) {
	charWidth := 7
	for i, ch := range text {
		cx := x + i*charWidth
		drawSimpleChar(img, ch, cx, y, c)
	}
}

// Basic 5x7 font patterns
var fontPatterns = map[rune][]string{
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
	'x': {"00000", "00000", "10001", "01010", "00100", "01010", "10001"},
	'n': {"00000", "00000", "10110", "11001", "10001", "10001", "10001"},
	'J': {"00111", "00010", "00010", "00010", "00010", "10010", "01100"},
	'G': {"01110", "10001", "10000", "10111", "10001", "10001", "01110"},
	'P': {"11110", "10001", "10001", "11110", "10000", "10000", "10000"},
	'U': {"10001", "10001", "10001", "10001", "10001", "10001", "01110"},
	'F': {"11111", "10000", "10000", "11110", "10000", "10000", "10000"},
	'e': {"00000", "00000", "01110", "10001", "11111", "10000", "01110"},
	'C': {"01110", "10001", "10000", "10000", "10000", "10001", "01110"},
	'I': {"01110", "00100", "00100", "00100", "00100", "00100", "01110"},
	'M': {"10001", "11011", "10101", "10101", "10001", "10001", "10001"},
	'E': {"11111", "10000", "10000", "11110", "10000", "10000", "11111"},
	'R': {"11110", "10001", "10001", "11110", "10100", "10010", "10001"},
	'O': {"01110", "10001", "10001", "10001", "10001", "10001", "01110"},
	'N': {"10001", "11001", "10101", "10011", "10001", "10001", "10001"},
	'S': {"01110", "10001", "10000", "01110", "00001", "10001", "01110"},
	's': {"00000", "00000", "01110", "10000", "01110", "00001", "11110"},
	'i': {"00100", "00000", "01100", "00100", "00100", "00100", "01110"},
	'o': {"00000", "00000", "01110", "10001", "10001", "10001", "01110"},
	'c': {"00000", "00000", "01110", "10000", "10000", "10001", "01110"},
	'f': {"00110", "01000", "01000", "11100", "01000", "01000", "01000"},
	'r': {"00000", "00000", "10110", "11001", "10000", "10000", "10000"},
	'a': {"00000", "00000", "01110", "00001", "01111", "10001", "01111"},
	't': {"00100", "00100", "01110", "00100", "00100", "00100", "00011"},
	'h': {"10000", "10000", "10110", "11001", "10001", "10001", "10001"},
	'm': {"00000", "00000", "11010", "10101", "10101", "10001", "10001"},
	'd': {"00001", "00001", "01101", "10011", "10001", "10001", "01111"},
	'v': {"00000", "00000", "10001", "10001", "10001", "01010", "00100"},
	'y': {"00000", "00000", "10001", "10001", "01111", "00001", "01110"},
	'k': {"10000", "10000", "10010", "10100", "11000", "10100", "10010"},
	'g': {"00000", "00000", "01111", "10001", "01111", "00001", "01110"},
	'W': {"10001", "10001", "10001", "10101", "10101", "10101", "01010"},
	'p': {"00000", "00000", "11110", "10001", "11110", "10000", "10000"},
	'!': {"00100", "00100", "00100", "00100", "00100", "00000", "00100"},
	'D': {"11100", "10010", "10001", "10001", "10001", "10010", "11100"},
	'w': {"00000", "00000", "10001", "10001", "10101", "10101", "01010"},
	'l': {"01100", "00100", "00100", "00100", "00100", "00100", "01110"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'u': {"00000", "00000", "10001", "10001", "10001", "10011", "01101"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'b': {"10000", "10000", "11110", "10001", "10001", "10001", "11110"},
	'q': {"00000", "00000", "01101", "10011", "10001", "01111", "00001"},
	'j': {"00010", "00000", "00110", "00010", "00010", "10010", "01100"},
	'z': {"00000", "00000", "11111", "00010", "00100", "01000", "11111"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	'(': {"00010", "00100", "01000", "01000", "01000", "00100", "00010"},
	')': {"01000", "00100", "00010", "00010", "00010", "00100", "01000"},
}

// drawSimpleChar draws a single character.
func drawSimpleChar(img *image.RGBA, ch rune, x, y int, c color.RGBA) {
	drawScaledChar(img, ch, x, y, 1, c)
}

// drawScaledText draws text with a scaling factor.
func drawScaledText(img *image.RGBA, text string, x, y int, scale int, c color.RGBA) {
	charWidth := 7 * scale
	for i, ch := range text {
		cx := x + i*charWidth
		drawScaledChar(img, ch, cx, y, scale, c)
	}
}

// drawScaledChar draws a single character with scaling.
func drawScaledChar(img *image.RGBA, ch rune, x, y int, scale int, c color.RGBA) {
	pattern, ok := fontPatterns[ch]
	if !ok {
		return
	}

	for dy, row := range pattern {
		for dx, pixel := range row {
			if pixel == '1' {
				// Draw scaled pixel
				for sy := 0; sy < scale; sy++ {
					for sx := 0; sx < scale; sx++ {
						px := x + dx*scale + sx
						py := y + dy*scale + sy
						if px >= 0 && px < img.Bounds().Dx() && py >= 0 && py < img.Bounds().Dy() {
							img.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
