// Package gui provides Fyne-based GUI components for architecture comparison.
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
	fyne.Do(func() {
		e.Refresh()
	})
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
	drawScaledText(img, "CPU", 5, padding+barHeight/2-7, 2, cpuColor)
	drawScaledText(img, fmt.Sprintf("%.0f", e.cpuSpec.EnergyFJ), padding+cpuWidth+5, padding+barHeight/2-7, 1, color.RGBA{200, 200, 200, 255})

	// Draw GPU bar
	gpuWidth := int(float64(chartWidth) * e.gpuSpec.EnergyFJ / maxVal)
	drawBar(img, padding, padding+barSpacing, gpuWidth, barHeight, gpuColor)
	drawScaledText(img, "GPU", 5, padding+barSpacing+barHeight/2-7, 2, gpuColor)
	drawScaledText(img, fmt.Sprintf("%.0f", e.gpuSpec.EnergyFJ), padding+gpuWidth+5, padding+barSpacing+barHeight/2-7, 1, color.RGBA{200, 200, 200, 255})

	// Draw FeCIM bar
	fecimWidth := int(float64(chartWidth) * e.fecimSpec.EnergyFJ / maxVal)
	if fecimWidth < 5 {
		fecimWidth = 5 // Minimum visible width
	}
	drawBar(img, padding, padding+2*barSpacing, fecimWidth, barHeight, fecimColor)
	drawScaledText(img, "FeCIM", 5, padding+2*barSpacing+barHeight/2-7, 2, fecimColor)
	drawScaledText(img, fmt.Sprintf("%.1f", e.fecimSpec.EnergyFJ), padding+fecimWidth+5, padding+2*barSpacing+barHeight/2-7, 1, color.RGBA{200, 200, 200, 255})

	// Draw axis
	axisColor := color.RGBA{100, 120, 150, 255}
	for x := padding; x < w-padding; x++ {
		img.Set(x, h-padding, axisColor)
	}
	for y := padding; y < h-padding; y++ {
		img.Set(padding, y, axisColor)
	}
	drawScaledText(img, "Energy (fJ/MAC)", w/2-60, h-40, 2, axisColor)

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
	d.macsLabel.SetText(fmt.Sprintf("MACs/inference: %s", formatNumberWithSuffix(float64(macs))))
	d.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", inferences))

	d.cpuEnergyLabel.SetText(fmt.Sprintf("CPU: %s/inf [1]", formatEnergy(cpuEnergy)))
	d.gpuEnergyLabel.SetText(fmt.Sprintf("GPU: %s/inf [2]", formatEnergy(gpuEnergy)))
	d.fecimEnergyLabel.SetText(fmt.Sprintf("FeCIM: %s/inf [3]*", formatEnergy(fecimEnergy)))

	d.cpuPowerLabel.SetText(fmt.Sprintf("CPU: %s", formatPower(cpuPower)))
	d.gpuPowerLabel.SetText(fmt.Sprintf("GPU: %s", formatPower(gpuPower)))
	d.fecimPowerLabel.SetText(fmt.Sprintf("FeCIM: %s*", formatPower(fecimPower)))

	d.cpuCostLabel.SetText(fmt.Sprintf("CPU: %s/month", formatCost(cpuCost)))
	d.gpuCostLabel.SetText(fmt.Sprintf("GPU: %s/month", formatCost(gpuCost)))
	d.fecimCostLabel.SetText(fmt.Sprintf("FeCIM: %s/month*", formatCost(fecimCost)))

	savingsVsGPU := (gpuCost - fecimCost) / gpuCost * 100
	d.savingsLabel.SetText(fmt.Sprintf("Potential Savings vs GPU: %.0f%%*", savingsVsGPU))

	fyne.Do(func() {
		d.Refresh()
	})
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

	// Disclaimer - clarity on verification status
	disclaimer := widget.NewLabel("* FeCIM energy values based on Dr. Tour's COSM 2025 claims. Independent verification pending.")
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
	return fyne.NewSize(400, 180)
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

// formatNumberWithSuffix formats large numbers with K, M, B, T suffixes.
func formatNumberWithSuffix(n float64) string {
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

// formatCost formats cost values appropriately for display.
func formatCost(cost float64) string {
	switch {
	case cost >= 1000000:
		return fmt.Sprintf("$%.1fM", cost/1000000)
	case cost >= 1000:
		return fmt.Sprintf("$%.1fK", cost/1000)
	case cost >= 1:
		return fmt.Sprintf("$%.0f", cost)
	case cost >= 0.01:
		return fmt.Sprintf("$%.2f", cost)
	case cost >= 0.0001:
		return fmt.Sprintf("$%.4f", cost)
	case cost > 0:
		return fmt.Sprintf("$%.2e", cost)
	default:
		return "$0.00"
	}
}

// formatPower formats power values with appropriate units.
func formatPower(power float64) string {
	switch {
	case power >= 1000000:
		return fmt.Sprintf("%.1f MW", power/1000000)
	case power >= 1000:
		return fmt.Sprintf("%.1f kW", power/1000)
	case power >= 1:
		return fmt.Sprintf("%.1f W", power)
	case power >= 0.001:
		return fmt.Sprintf("%.1f mW", power*1000)
	case power >= 0.000001:
		return fmt.Sprintf("%.1f µW", power*1000000)
	case power > 0:
		return fmt.Sprintf("%.1f nW", power*1e9)
	default:
		return "0 W"
	}
}

// formatEnergy formats energy values with appropriate units.
// Input is in µJ (microjoules).
func formatEnergy(energy float64) string {
	switch {
	case energy >= 1000000:
		return fmt.Sprintf("%.2f J", energy/1000000)
	case energy >= 1000:
		return fmt.Sprintf("%.2f mJ", energy/1000)
	case energy >= 1:
		return fmt.Sprintf("%.2f µJ", energy)
	case energy >= 0.001:
		return fmt.Sprintf("%.2f nJ", energy*1000)
	case energy >= 0.000001:
		return fmt.Sprintf("%.3f pJ", energy*1000000)
	case energy > 0:
		return fmt.Sprintf("%.1f fJ", energy*1e9)
	default:
		return "0 µJ"
	}
}

// DataCenterTransformation shows before/after data center comparison using Fyne widgets.
type DataCenterTransformation struct {
	widget.BaseWidget

	mu             sync.RWMutex
	beforePowerW   float64
	afterPowerW    float64
	savingsPercent float64
	minSize        fyne.Size

	container    *fyne.Container
	beforeRacks  []*canvas.Rectangle
	afterRacks   []*canvas.Rectangle
	beforeLabel  *widget.Label
	afterLabel   *widget.Label
	savingsLabel *canvas.Text
}

// NewDataCenterTransformation creates a new transformation visual.
func NewDataCenterTransformation() *DataCenterTransformation {
	d := &DataCenterTransformation{
		beforePowerW:   1000,
		afterPowerW:    100,
		savingsPercent: 90,
		minSize:        fyne.NewSize(350, 70),
		beforeRacks:    make([]*canvas.Rectangle, 10),
		afterRacks:     make([]*canvas.Rectangle, 2),
	}
	d.ExtendBaseWidget(d)
	return d
}

// SetValues updates the before/after values.
func (d *DataCenterTransformation) SetValues(beforeW, afterW float64) {
	d.mu.Lock()
	d.beforePowerW = beforeW
	d.afterPowerW = afterW
	if beforeW > 0 {
		d.savingsPercent = (beforeW - afterW) / beforeW * 100
	}
	d.mu.Unlock()
	fyne.Do(func() {
		d.Refresh()
	})
}

// UpdateAnimation advances the animation (not used in widget version).
func (d *DataCenterTransformation) UpdateAnimation(dt float64) {
	// Animation handled by Refresh
}

// MinSize returns minimum size.
func (d *DataCenterTransformation) MinSize() fyne.Size {
	return d.minSize
}

// CreateRenderer implements fyne.Widget.
func (d *DataCenterTransformation) CreateRenderer() fyne.WidgetRenderer {
	// Before section - GPU (many racks)
	beforeTitle := widget.NewLabel("Before (GPU)")
	var beforeRackWidgets []fyne.CanvasObject
	for i := 0; i < 10; i++ {
		d.beforeRacks[i] = canvas.NewRectangle(color.RGBA{180, 80, 80, 255})
		d.beforeRacks[i].SetMinSize(fyne.NewSize(8, 30))
		beforeRackWidgets = append(beforeRackWidgets, d.beforeRacks[i])
	}
	beforeRacksRow := container.NewHBox(beforeRackWidgets...)
	d.beforeLabel = widget.NewLabel("1000 W")
	beforeCol := container.NewVBox(beforeTitle, beforeRacksRow, d.beforeLabel)

	// Arrow
	arrowText := canvas.NewText("→", color.RGBA{0, 212, 255, 255})
	arrowText.TextSize = 24

	// After section - FeCIM (few racks)
	afterTitle := widget.NewLabel("After (FeCIM)")
	var afterRackWidgets []fyne.CanvasObject
	for i := 0; i < 2; i++ {
		d.afterRacks[i] = canvas.NewRectangle(color.RGBA{80, 180, 120, 255})
		d.afterRacks[i].SetMinSize(fyne.NewSize(8, 30))
		afterRackWidgets = append(afterRackWidgets, d.afterRacks[i])
	}
	afterRacksRow := container.NewHBox(afterRackWidgets...)
	d.afterLabel = widget.NewLabel("100 W")
	afterCol := container.NewVBox(afterTitle, afterRacksRow, d.afterLabel)

	// Savings
	d.savingsLabel = canvas.NewText("90% Less", color.RGBA{0, 212, 255, 255})
	d.savingsLabel.TextSize = 12
	d.savingsLabel.TextStyle = fyne.TextStyle{Bold: true}

	d.container = container.NewHBox(
		beforeCol,
		container.NewCenter(arrowText),
		afterCol,
		container.NewCenter(d.savingsLabel),
	)

	return widget.NewSimpleRenderer(d.container)
}

// Refresh updates the display.
func (d *DataCenterTransformation) Refresh() {
	d.mu.RLock()
	beforePower := d.beforePowerW
	afterPower := d.afterPowerW
	savings := d.savingsPercent
	d.mu.RUnlock()

	if d.beforeLabel != nil {
		d.beforeLabel.SetText(fmt.Sprintf("%.0f W", beforePower))
	}
	if d.afterLabel != nil {
		d.afterLabel.SetText(fmt.Sprintf("%.0f W", afterPower))
	}
	if d.savingsLabel != nil {
		d.savingsLabel.Text = fmt.Sprintf("%.0f%% Less", savings)
		canvas.Refresh(d.savingsLabel)
	}
}

// drawSimpleText draws text using a simple bitmap font.
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.RGBA) {
	drawScaledText(img, text, x, y, 1, c)
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
	'(': {"00010", "00100", "01000", "01000", "01000", "00100", "00010"},
	')': {"01000", "00100", "00010", "00010", "00010", "00100", "01000"},
	'_': {"00000", "00000", "00000", "00000", "00000", "00000", "11111"},
	'A': {"01110", "10001", "10001", "11111", "10001", "10001", "10001"},
	'B': {"11110", "10001", "10001", "11110", "10001", "10001", "11110"},
	'D': {"11100", "10010", "10001", "10001", "10001", "10010", "11100"},
	'H': {"10001", "10001", "10001", "11111", "10001", "10001", "10001"},
	'K': {"10001", "10010", "10100", "11000", "10100", "10010", "10001"},
	'L': {"10000", "10000", "10000", "10000", "10000", "10000", "11111"},
	'T': {"11111", "00100", "00100", "00100", "00100", "00100", "00100"},
	'V': {"10001", "10001", "10001", "10001", "10001", "01010", "00100"},
	'X': {"10001", "10001", "01010", "00100", "01010", "10001", "10001"},
	'Y': {"10001", "10001", "01010", "00100", "00100", "00100", "00100"},
	'Z': {"11111", "00001", "00010", "00100", "01000", "10000", "11111"},
	'b': {"10000", "10000", "11110", "10001", "10001", "10001", "11110"},
	'l': {"01100", "00100", "00100", "00100", "00100", "00100", "01110"},
	'u': {"00000", "00000", "10001", "10001", "10001", "10011", "01101"},
	'w': {"00000", "00000", "10001", "10001", "10101", "10101", "01010"},
	'z': {"00000", "00000", "11111", "00010", "00100", "01000", "11111"},
	'[': {"01110", "01000", "01000", "01000", "01000", "01000", "01110"},
	']': {"01110", "00010", "00010", "00010", "00010", "00010", "01110"},
	'/': {"00001", "00010", "00010", "00100", "01000", "01000", "10000"},
	'=': {"00000", "00000", "11111", "00000", "11111", "00000", "00000"},
	'+': {"00000", "00100", "00100", "11111", "00100", "00100", "00000"},
	'<': {"00010", "00100", "01000", "10000", "01000", "00100", "00010"},
	'>': {"01000", "00100", "00010", "00001", "00010", "00100", "01000"},
	',': {"00000", "00000", "00000", "00000", "00110", "00100", "01000"},
	'q': {"00000", "00000", "01111", "10001", "01111", "00001", "00001"},
	'j': {"00010", "00000", "00110", "00010", "00010", "10010", "01100"},
}
