// Package gui provides Fyne-based GUI components for architecture comparison.
package gui

import (
	"fmt"
	"image"
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

// EnergyBarChart displays energy per MAC comparison.
type EnergyBarChart struct {
	widget.BaseWidget

	cpuSpec   EnergySpec
	gpuSpec   EnergySpec
	fecimSpec EnergySpec

	raster *canvas.Raster
}

// NewEnergyBarChart creates a new energy bar chart.
func NewEnergyBarChart() *EnergyBarChart {
	e := &EnergyBarChart{}
	e.ExtendBaseWidget(e)
	return e
}

// SetValues updates the energy specifications.
func (e *EnergyBarChart) SetValues(cpu, gpu, fecim EnergySpec) {
	e.cpuSpec = cpu
	e.gpuSpec = gpu
	e.fecimSpec = fecim
	e.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (e *EnergyBarChart) CreateRenderer() fyne.WidgetRenderer {
	e.raster = canvas.NewRaster(e.generateImage)

	// Create source labels
	cpuSourceLabel := widget.NewLabel(fmt.Sprintf("[1] %s", e.cpuSpec.Source))
	cpuSourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	gpuSourceLabel := widget.NewLabel(fmt.Sprintf("[2] %s", e.gpuSpec.Source))
	gpuSourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	fecimSourceLabel := widget.NewLabel(fmt.Sprintf("[3] %s", e.fecimSpec.Source))
	fecimSourceLabel.TextStyle = fyne.TextStyle{Italic: true}

	sources := container.NewVBox(
		cpuSourceLabel,
		gpuSourceLabel,
		fecimSourceLabel,
	)

	content := container.NewBorder(
		nil,
		sources,
		nil, nil,
		container.NewMax(e.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (e *EnergyBarChart) MinSize() fyne.Size {
	return fyne.NewSize(400, 200)
}

// generateImage creates the bar chart.
func (e *EnergyBarChart) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 100 || h < 100 {
		return img
	}

	// Chart dimensions
	padding := 60
	chartWidth := w - 2*padding
	chartHeight := h - 2*padding - 40 // Leave room for labels

	// Max value for scaling (CPU is largest)
	maxVal := e.cpuSpec.EnergyFJ
	if maxVal == 0 {
		maxVal = 1000
	}

	// Bar dimensions
	barHeight := chartHeight / 4
	barSpacing := chartHeight / 4

	// Colors
	cpuColor := color.RGBA{200, 100, 100, 255}   // Red
	gpuColor := color.RGBA{200, 180, 100, 255}   // Yellow
	fecimColor := color.RGBA{100, 200, 150, 255} // Green

	// Draw CPU bar
	cpuWidth := int(float64(chartWidth) * e.cpuSpec.EnergyFJ / maxVal)
	drawBar(img, padding, padding, cpuWidth, barHeight, cpuColor)

	// Draw GPU bar
	gpuWidth := int(float64(chartWidth) * e.gpuSpec.EnergyFJ / maxVal)
	drawBar(img, padding, padding+barSpacing, gpuWidth, barHeight, gpuColor)

	// Draw FeCIM bar
	fecimWidth := int(float64(chartWidth) * e.fecimSpec.EnergyFJ / maxVal)
	if fecimWidth < 5 {
		fecimWidth = 5 // Minimum visible width
	}
	drawBar(img, padding, padding+2*barSpacing, fecimWidth, barHeight, fecimColor)

	// Draw axis
	axisColor := color.RGBA{100, 120, 150, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}
	for y := padding; y < h-padding; y++ {
		img.Set(padding, y, axisColor)
	}

	return img
}

// drawBar draws a filled rectangle.
func drawBar(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			img.Set(x+dx, y+dy, c)
		}
	}
}

// ArchitectureDiagram shows Von Neumann vs CIM architecture.
type ArchitectureDiagram struct {
	widget.BaseWidget
	raster *canvas.Raster
}

// NewArchitectureDiagram creates a new architecture diagram.
func NewArchitectureDiagram() *ArchitectureDiagram {
	a := &ArchitectureDiagram{}
	a.ExtendBaseWidget(a)
	return a
}

// CreateRenderer implements fyne.Widget.
func (a *ArchitectureDiagram) CreateRenderer() fyne.WidgetRenderer {
	a.raster = canvas.NewRaster(a.generateImage)

	titleLabel := widget.NewLabel("Why the Difference?")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}
	titleLabel.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(
		titleLabel,
		nil, nil, nil,
		container.NewMax(a.raster),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (a *ArchitectureDiagram) MinSize() fyne.Size {
	return fyne.NewSize(400, 150)
}

// generateImage creates the architecture comparison diagram.
func (a *ArchitectureDiagram) generateImage(w, h int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, w, h))

	// Background
	bgColor := color.RGBA{25, 35, 55, 255}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, bgColor)
		}
	}

	if w < 200 || h < 80 {
		return img
	}

	// Left side: Von Neumann (CPU/GPU)
	// Draw CPU box
	cpuColor := color.RGBA{200, 100, 100, 255}
	drawBox(img, 30, 30, 60, 40, cpuColor)

	// Draw DRAM box
	dramColor := color.RGBA{100, 100, 200, 255}
	drawBox(img, 130, 30, 60, 40, dramColor)

	// Draw arrow between them (bidirectional)
	arrowColor := color.RGBA{255, 200, 100, 255}
	for x := 95; x < 125; x++ {
		img.Set(x, 50, arrowColor)
		img.Set(x, 51, arrowColor)
	}

	// Right side: CIM (FeCIM)
	// Draw combined box
	cimColor := color.RGBA{100, 200, 150, 255}
	midX := w/2 + 50
	drawBox(img, midX, 30, 100, 40, cimColor)

	// Draw "NO DATA MOVEMENT" text indicator (simple line)
	noMoveColor := color.RGBA{100, 255, 150, 255}
	for y := 75; y < 85; y++ {
		for x := midX; x < midX+100; x++ {
			img.Set(x, y, noMoveColor)
		}
	}

	return img
}

// drawBox draws a rectangle outline.
func drawBox(img *image.RGBA, x, y, width, height int, c color.RGBA) {
	// Fill
	fillColor := color.RGBA{c.R / 2, c.G / 2, c.B / 2, 255}
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			img.Set(x+dx, y+dy, fillColor)
		}
	}
	// Border
	for dx := 0; dx < width; dx++ {
		img.Set(x+dx, y, c)
		img.Set(x+dx, y+height-1, c)
	}
	for dy := 0; dy < height; dy++ {
		img.Set(x, y+dy, c)
		img.Set(x+width-1, y+dy, c)
	}
}

// DataCenterCalculator shows power and cost calculations.
type DataCenterCalculator struct {
	widget.BaseWidget

	workloadLabel   *widget.Label
	macsLabel       *widget.Label
	inferencesLabel *widget.Label

	cpuEnergyLabel   *widget.Label
	gpuEnergyLabel   *widget.Label
	fecimEnergyLabel *widget.Label

	cpuPowerLabel   *widget.Label
	gpuPowerLabel   *widget.Label
	fecimPowerLabel *widget.Label

	cpuCostLabel   *widget.Label
	gpuCostLabel   *widget.Label
	fecimCostLabel *widget.Label

	savingsLabel *widget.Label
}

// NewDataCenterCalculator creates a new calculator widget.
func NewDataCenterCalculator() *DataCenterCalculator {
	d := &DataCenterCalculator{}
	d.workloadLabel = widget.NewLabel("Workload: -")
	d.macsLabel = widget.NewLabel("MACs/inference: -")
	d.inferencesLabel = widget.NewLabel("Inferences/sec: -")

	d.cpuEnergyLabel = widget.NewLabel("CPU: - µJ/inf")
	d.gpuEnergyLabel = widget.NewLabel("GPU: - µJ/inf")
	d.fecimEnergyLabel = widget.NewLabel("FeCIM: - µJ/inf")

	d.cpuPowerLabel = widget.NewLabel("CPU: - W")
	d.gpuPowerLabel = widget.NewLabel("GPU: - W")
	d.fecimPowerLabel = widget.NewLabel("FeCIM: - W")

	d.cpuCostLabel = widget.NewLabel("CPU: $-/month")
	d.gpuCostLabel = widget.NewLabel("GPU: $-/month")
	d.fecimCostLabel = widget.NewLabel("FeCIM: $-/month")

	d.savingsLabel = widget.NewLabel("Savings: -")
	d.savingsLabel.TextStyle = fyne.TextStyle{Bold: true}

	d.ExtendBaseWidget(d)
	return d
}

// SetResults updates the calculator with new results.
func (d *DataCenterCalculator) SetResults(
	workload string,
	macs int,
	inferences float64,
	cpuEnergy, gpuEnergy, fecimEnergy float64,
	cpuPower, gpuPower, fecimPower float64,
	cpuCost, gpuCost, fecimCost float64,
) {
	d.workloadLabel.SetText(fmt.Sprintf("Workload: %s", workload))
	d.macsLabel.SetText(fmt.Sprintf("MACs/inference: %s", formatNumber(float64(macs))))
	d.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", inferences))

	d.cpuEnergyLabel.SetText(fmt.Sprintf("CPU: %.2f µJ/inf [1]", cpuEnergy))
	d.gpuEnergyLabel.SetText(fmt.Sprintf("GPU: %.2f µJ/inf [2]", gpuEnergy))
	d.fecimEnergyLabel.SetText(fmt.Sprintf("FeCIM: %.2f µJ/inf [3]*", fecimEnergy))

	d.cpuPowerLabel.SetText(fmt.Sprintf("CPU: %.1f W", cpuPower))
	d.gpuPowerLabel.SetText(fmt.Sprintf("GPU: %.1f W", gpuPower))
	d.fecimPowerLabel.SetText(fmt.Sprintf("FeCIM: %.2f W*", fecimPower))

	d.cpuCostLabel.SetText(fmt.Sprintf("CPU: $%.0f/month", cpuCost))
	d.gpuCostLabel.SetText(fmt.Sprintf("GPU: $%.0f/month", gpuCost))
	d.fecimCostLabel.SetText(fmt.Sprintf("FeCIM: $%.0f/month*", fecimCost))

	savingsVsGPU := (gpuCost - fecimCost) / gpuCost * 100
	d.savingsLabel.SetText(fmt.Sprintf("Savings vs GPU: %.0f%% (if claims hold)*", savingsVsGPU))

	d.Refresh()
}

// CreateRenderer implements fyne.Widget.
func (d *DataCenterCalculator) CreateRenderer() fyne.WidgetRenderer {
	// Configuration section
	configBox := container.NewVBox(
		d.workloadLabel,
		d.macsLabel,
		d.inferencesLabel,
	)

	// Energy section
	energyLabel := widget.NewLabel("Energy per Inference:")
	energyLabel.TextStyle = fyne.TextStyle{Bold: true}
	energyBox := container.NewVBox(
		energyLabel,
		d.cpuEnergyLabel,
		d.gpuEnergyLabel,
		d.fecimEnergyLabel,
	)

	// Power section
	powerLabel := widget.NewLabel("Total Power:")
	powerLabel.TextStyle = fyne.TextStyle{Bold: true}
	powerBox := container.NewVBox(
		powerLabel,
		d.cpuPowerLabel,
		d.gpuPowerLabel,
		d.fecimPowerLabel,
	)

	// Cost section
	costLabel := widget.NewLabel("Monthly Cost (@$0.10/kWh):")
	costLabel.TextStyle = fyne.TextStyle{Bold: true}
	costBox := container.NewVBox(
		costLabel,
		d.cpuCostLabel,
		d.gpuCostLabel,
		d.fecimCostLabel,
		widget.NewSeparator(),
		d.savingsLabel,
	)

	// Disclaimer
	disclaimer := widget.NewLabel("* FeCIM estimates based on Dr. Tour's claims. NOT verified.")
	disclaimer.TextStyle = fyne.TextStyle{Italic: true}

	content := container.NewVBox(
		configBox,
		widget.NewSeparator(),
		container.NewGridWithColumns(3, energyBox, powerBox, costBox),
		widget.NewSeparator(),
		disclaimer,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (d *DataCenterCalculator) MinSize() fyne.Size {
	return fyne.NewSize(500, 250)
}

// VerifiedClaimsTable shows what's verified vs claimed.
type VerifiedClaimsTable struct {
	widget.BaseWidget
}

// NewVerifiedClaimsTable creates a new table.
func NewVerifiedClaimsTable() *VerifiedClaimsTable {
	v := &VerifiedClaimsTable{}
	v.ExtendBaseWidget(v)
	return v
}

// CreateRenderer implements fyne.Widget.
func (v *VerifiedClaimsTable) CreateRenderer() fyne.WidgetRenderer {
	titleLabel := widget.NewLabel("Verified vs Claimed")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	verifiedLabel := widget.NewLabel("VERIFIED (from Dr. Tour):")
	verifiedLabel.TextStyle = fyne.TextStyle{Bold: true}

	verified := container.NewVBox(
		widget.NewLabel("  30 discrete analog levels"),
		widget.NewLabel("  87% MNIST accuracy"),
		widget.NewLabel("  CMOS compatible fab"),
		widget.NewLabel("  Non-volatile (no refresh)"),
	)

	claimedLabel := widget.NewLabel("CLAIMED (not verified):")
	claimedLabel.TextStyle = fyne.TextStyle{Bold: true}

	claimed := container.NewVBox(
		widget.NewLabel("  10M× lower than NAND"),
		widget.NewLabel("  1000× lower than DRAM"),
		widget.NewLabel("  80-90% DC energy savings"),
	)

	statusLabel := widget.NewLabel("Status: TRL 4 (Lab only)")
	statusLabel.TextStyle = fyne.TextStyle{Bold: true, Italic: true}

	content := container.NewVBox(
		titleLabel,
		widget.NewSeparator(),
		verifiedLabel,
		verified,
		widget.NewSeparator(),
		claimedLabel,
		claimed,
		widget.NewSeparator(),
		statusLabel,
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (v *VerifiedClaimsTable) MinSize() fyne.Size {
	return fyne.NewSize(200, 280)
}

// formatNumber formats large numbers with K, M, B, T suffixes.
func formatNumber(n float64) string {
	switch {
	case n >= 1e12:
		return fmt.Sprintf("%.1fT", n/1e12)
	case n >= 1e9:
		return fmt.Sprintf("%.1fB", n/1e9)
	case n >= 1e6:
		return fmt.Sprintf("%.1fM", n/1e6)
	case n >= 1e3:
		return fmt.Sprintf("%.1fK", n/1e3)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}
