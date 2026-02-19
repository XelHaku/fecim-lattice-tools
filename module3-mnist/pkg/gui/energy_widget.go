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

	"fecim-lattice-tools/module3-mnist/pkg/core"
	"fecim-lattice-tools/shared/canvas"
	"fecim-lattice-tools/shared/physics"
)

// EnergyConfig defines energy parameters for calculations.
type EnergyConfig struct {
	// GPU energy model (kept separate from FeCIM core model).
	//
	// FeCIM energy is computed via module3-mnist/pkg/core (bit-scaled fJ/MAC + ADC/DAC overhead)
	// so that GUI and core stay aligned.
	EnergyPerMAC_GPU float64 // ~500 pJ per MAC including DRAM access (Horowitz 2014)
}

// DefaultEnergyConfig returns energy parameters for the GUI.
func DefaultEnergyConfig() EnergyConfig {
	return EnergyConfig{
		EnergyPerMAC_GPU: 500e-12, // 500 picojoules per MAC including DRAM access (Horowitz 2014)
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
	ew.statsLabel.Wrapping = fyne.TextWrapWord
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
//
// FeCIM energy is computed using the shared core model (bit-scaled energy per MAC + ADC/DAC overhead).
// GPU energy uses a simple per-MAC constant (EnergyPerMAC_GPU).
func (ew *EnergyWidget) RecordInference(cfg *core.NetworkConfig) {
	ew.mu.Lock()
	defer ew.mu.Unlock()

	ew.totalInferences++

	est := core.EstimateInferenceEnergyJ(cfg, ew.inputSize, ew.hiddenSize, ew.outputSize)

	// GPU energy calculation (simple model dominated by data movement)
	gpuEnergy := float64(est.TotalMACs) * ew.config.EnergyPerMAC_GPU

	// Convert to nanojoules for display
	ew.lastEnergyFeCIM = est.TotalJ * 1e9
	ew.lastEnergyGPU = gpuEnergy * 1e9

	// Calculate efficiency ratio
	if est.TotalJ > 0 {
		ew.efficiencyRatio = gpuEnergy / est.TotalJ
	}

	// Update running totals (in Joules)
	ew.totalEnergyFeCIM += est.TotalJ
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
	return fyne.NewSize(280, 160)
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
	utils.DrawSimpleText(img, "Per Inference Energy (Horowitz 2014 model):", padding, y, color.RGBA{180, 180, 200, 255})
	y += 15

	// GPU bar (full width represents GPU energy)
	maxBarWidth := w - 2*padding - labelWidth - 120
	gpuBarWidth := maxBarWidth
	if gpuBarWidth < 10 {
		gpuBarWidth = 10
	}

	// GPU label and bar with units
	utils.DrawSimpleText(img, "FP32:", padding, y+8, color.RGBA{255, 100, 100, 255})
	drawBar(img, padding+labelWidth, y, gpuBarWidth, barHeight, color.RGBA{200, 80, 80, 255})
	// UI-022 fix: Clear unit labeling
	gpuLabel := fmt.Sprintf("%.2f nJ", gpuEnergy)
	if gpuEnergy >= 1000 {
		gpuLabel = fmt.Sprintf("%.2f µJ", gpuEnergy/1000)
	}
	utils.DrawSimpleText(img, gpuLabel, padding+labelWidth+gpuBarWidth+5, y+8, color.RGBA{200, 150, 150, 255})
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
	utils.DrawSimpleText(img, "FeCIM:", padding, y+8, color.RGBA{100, 255, 180, 255})
	drawBar(img, padding+labelWidth, y, fecimBarWidth, barHeight, color.RGBA{80, 200, 150, 255})
	// UI-022 fix: Clear unit labeling (nJ or pJ)
	fecimLabel := fmt.Sprintf("%.2f nJ", fecimEnergy)
	if fecimEnergy < 1 {
		fecimLabel = fmt.Sprintf("%.0f pJ", fecimEnergy*1000)
	}
	utils.DrawSimpleText(img, fecimLabel, padding+labelWidth+fecimBarWidth+5, y+8, color.RGBA{150, 200, 180, 255})
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
		utils.DrawSimpleText(img, effText, boxX+20, boxY+12, color.RGBA{0, 230, 255, 255})
		y += boxH + 5
	}

	// Section 2: Why CIM wins (bottom section)
	if y < h-30 {
		y = h - 35
		utils.DrawSimpleText(img, "Physics does the math! No data movement needed.", padding, y, color.RGBA{120, 120, 140, 255})
		y += 12

		if totalInf > 0 {
			utils.DrawSimpleText(img, fmt.Sprintf("Session: %d inferences tracked", totalInf), padding, y, color.RGBA{100, 100, 120, 255})
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
