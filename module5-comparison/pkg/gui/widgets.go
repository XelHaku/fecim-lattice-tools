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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"fecim-lattice-tools/shared/physics"
	sharedutils "fecim-lattice-tools/shared/utils"
)

// Re-export shared physics formatters for package-level use
var (
	// FormatEnergy formats energy in Joules with appropriate SI prefix (fJ to J)
	FormatEnergy = physics.FormatEnergy
	// FormatEnergyUJ formats energy given in microjoules
	FormatEnergyUJ = physics.FormatEnergyUJ
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
	sharedutils.DrawRect(img, x, y, width, height, c)
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

// DataCenterCalculator shows power and cost calculations (model inputs).
// HERO: Dynamic "$XX MILLION ANNUAL SAVINGS" (model output)
type DataCenterCalculator struct {
	widget.BaseWidget

	// Hero display
	heroSavingsText  *canvas.Text
	heroSavingsLabel *canvas.Text

	// Comparison display
	gpuCostText    *canvas.Text
	fecimCostText  *canvas.Text
	savingsPercent *canvas.Text

	// Configuration display
	workloadLabel   *widget.Label
	inferencesLabel *widget.Label

	// Internal values for calculations
	currentWorkload   string
	currentInferences float64
	currentGpuCost    float64
	currentFecimCost  float64
	annualSavings     float64
}

const dataCenterServerScale = 10000.0

func annualSavingsForScale(gpuCost, fecimCost float64) float64 {
	monthlyGpuCost := gpuCost * dataCenterServerScale
	monthlyFecimCost := fecimCost * dataCenterServerScale
	return (monthlyGpuCost - monthlyFecimCost) * 12
}

// NewDataCenterCalculator creates a new calculator widget.
func NewDataCenterCalculator() *DataCenterCalculator {
	d := &DataCenterCalculator{}
	d.workloadLabel = widget.NewLabel("Workload: MNIST")
	d.inferencesLabel = widget.NewLabel("Scale: 10,000 servers")

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
	d.currentWorkload = workload
	d.currentInferences = inferences
	d.currentGpuCost = gpuCost
	d.currentFecimCost = fecimCost

	// Calculate annual savings (monthly * 12 * 10,000 server scale factor)
	monthlyGpuCost := gpuCost * dataCenterServerScale
	monthlyFecimCost := fecimCost * dataCenterServerScale
	d.annualSavings = annualSavingsForScale(gpuCost, fecimCost)

	fyne.Do(func() {
		if d.workloadLabel != nil {
			d.workloadLabel.SetText(fmt.Sprintf("Workload: %s", workload))
		}
		if d.inferencesLabel != nil {
			d.inferencesLabel.SetText(fmt.Sprintf("Inferences/sec: %.0f", inferences))
		}

		// Update hero savings display
		if d.heroSavingsText != nil {
			d.heroSavingsText.Text = formatHeroMoney(d.annualSavings)
			canvas.Refresh(d.heroSavingsText)
		}

		// Update comparison display
		if d.gpuCostText != nil {
			d.gpuCostText.Text = fmt.Sprintf("GPU Baseline: %s/month", formatCost(monthlyGpuCost))
			canvas.Refresh(d.gpuCostText)
		}
		if d.fecimCostText != nil {
			d.fecimCostText.Text = fmt.Sprintf("FeCIM (model input; TRL 4): %s/month", formatCost(monthlyFecimCost))
			canvas.Refresh(d.fecimCostText)
		}
		if d.savingsPercent != nil {
			savingsPct := (gpuCost - fecimCost) / gpuCost * 100
			d.savingsPercent.Text = fmt.Sprintf("MODEL SAVINGS: %.0f%%", savingsPct)
			canvas.Refresh(d.savingsPercent)
		}

		d.Refresh()
	})
}

// formatHeroMoney formats large money values for hero display.
func formatHeroMoney(amount float64) string {
	if amount >= 1e9 {
		return fmt.Sprintf("$%.1fB", amount/1e9)
	} else if amount >= 1e6 {
		return fmt.Sprintf("$%.0fM", amount/1e6)
	} else if amount >= 1e3 {
		return fmt.Sprintf("$%.0fK", amount/1e3)
	}
	return fmt.Sprintf("$%.0f", amount)
}

// CreateRenderer implements fyne.Widget.
func (d *DataCenterCalculator) CreateRenderer() fyne.WidgetRenderer {
	// === HERO SECTION: COMPACT ANNUAL SAVINGS ===
	d.heroSavingsText = canvas.NewText("$0M", color.RGBA{46, 204, 113, 255}) // Green for money
	d.heroSavingsText.TextSize = 48
	d.heroSavingsText.TextStyle = fyne.TextStyle{Bold: true}
	d.heroSavingsText.Alignment = fyne.TextAlignCenter

	d.heroSavingsLabel = canvas.NewText("ANNUAL SAVINGS (MODEL)", color.RGBA{0, 212, 255, 255})
	d.heroSavingsLabel.TextSize = 14
	d.heroSavingsLabel.TextStyle = fyne.TextStyle{Bold: true}
	d.heroSavingsLabel.Alignment = fyne.TextAlignCenter

	// === CONFIGURATION ROW ===
	configRow := container.NewHBox(
		d.workloadLabel,
		widget.NewLabel("|"),
		d.inferencesLabel,
		widget.NewLabel("|"),
		widget.NewLabel("Scale: 10,000 servers"),
	)

	// === COMPARISON SECTION ===
	d.gpuCostText = canvas.NewText("GPU Baseline: $0/month", color.RGBA{231, 76, 60, 255}) // Red
	d.gpuCostText.TextSize = 12
	d.gpuCostText.Alignment = fyne.TextAlignCenter

	d.fecimCostText = canvas.NewText("FeCIM (model input; TRL 4): $0/month", color.RGBA{46, 204, 113, 255}) // Green
	d.fecimCostText.TextSize = 12
	d.fecimCostText.Alignment = fyne.TextAlignCenter

	d.savingsPercent = canvas.NewText("MODEL SAVINGS: 99%", color.RGBA{0, 212, 255, 255}) // Cyan
	d.savingsPercent.TextSize = 16
	d.savingsPercent.TextStyle = fyne.TextStyle{Bold: true}
	d.savingsPercent.Alignment = fyne.TextAlignCenter

	comparisonSection := container.NewHBox(
		layout.NewSpacer(),
		container.NewCenter(d.gpuCostText),
		container.NewCenter(d.fecimCostText),
		container.NewCenter(d.savingsPercent),
		layout.NewSpacer(),
	)

	// === CITATION ===
	citation := canvas.NewText("Model input references (not validated): NVIDIA H100 datasheets, Intel/AMD datasheets | Cost: $0.10/kWh", color.RGBA{160, 180, 200, 255})
	citation.TextSize = 9
	citation.TextStyle = fyne.TextStyle{Italic: true}
	citation.Alignment = fyne.TextAlignCenter

	content := container.NewVBox(
		container.NewHBox(
			layout.NewSpacer(),
			container.NewVBox(
				container.NewCenter(d.heroSavingsText),
				container.NewCenter(d.heroSavingsLabel),
			),
			layout.NewSpacer(),
		),
		container.NewCenter(configRow),
		comparisonSection,
		container.NewCenter(citation),
	)

	return widget.NewSimpleRenderer(content)
}

// MinSize returns minimum size.
func (d *DataCenterCalculator) MinSize() fyne.Size {
	return fyne.NewSize(600, 180)
}

// VerifiedClaimsTable shows model inputs vs scenario inputs.
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
	titleLabel := widget.NewLabel("Model Inputs vs Scenario Inputs")
	titleLabel.TextStyle = fyne.TextStyle{Bold: true}

	verifiedLabel := widget.NewLabel("MODEL INPUT REFERENCES:")
	verifiedLabel.TextStyle = fyne.TextStyle{Bold: true}

	verified := container.NewVBox(
		widget.NewLabel("  Analog levels (literature ranges; not validated here)"),
		widget.NewLabel("  MNIST accuracy (literature ranges; not validated here)"),
		widget.NewLabel("  CMOS compatibility (assumed)"),
	)

	claimedLabel := widget.NewLabel("SCENARIO INPUTS (NOT VALIDATED):")
	claimedLabel.TextStyle = fyne.TextStyle{Bold: true}

	claimed := container.NewVBox(
		widget.NewLabel("  30 levels (conference claim)"),
		widget.NewLabel("  25-100× lower than NAND (scenario input)"),
		widget.NewLabel("  1000× lower than DRAM (scenario input)"),
		widget.NewLabel("  80-90% DC energy savings (scenario input)"),
	)

	statusLabel := widget.NewLabel("Status: TRL 4 (Lab only) — model inputs only")
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

// Note: formatEnergy moved to shared/physics package
// Use physics.FormatEnergy(joules) or physics.FormatEnergyUJ(microjoules)

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
	beforeTitle := widget.NewLabelWithStyle("Before (GPU, model)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
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
	afterTitle := widget.NewLabelWithStyle("After (FeCIM, model)", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	var afterRackWidgets []fyne.CanvasObject
	for i := 0; i < 2; i++ {
		d.afterRacks[i] = canvas.NewRectangle(color.RGBA{80, 180, 120, 255})
		d.afterRacks[i].SetMinSize(fyne.NewSize(8, 30))
		afterRackWidgets = append(afterRackWidgets, d.afterRacks[i])
	}
	afterRacksRow := container.NewHBox(afterRackWidgets...)
	d.afterLabel = widget.NewLabel("100 W")
	afterCol := container.NewVBox(afterTitle, afterRacksRow, d.afterLabel)

	// Savings - larger text size for prominence
	d.savingsLabel = canvas.NewText("Model: 90% Less", color.RGBA{0, 212, 255, 255})
	d.savingsLabel.TextSize = 16
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
		d.savingsLabel.Text = fmt.Sprintf("Model: %.0f%% Less", savings)
		canvas.Refresh(d.savingsLabel)
	}
}

// drawSimpleText draws text using a simple bitmap font.
func drawSimpleText(img *image.RGBA, text string, x, y int, c color.RGBA) {
	sharedutils.DrawSimpleText(img, text, x, y, c)
}

// drawSimpleChar draws a single character.
func drawSimpleChar(img *image.RGBA, ch rune, x, y int, c color.RGBA) {
	sharedutils.DrawSimpleChar(img, ch, x, y, c)
}

// drawScaledText draws text with a scaling factor.
func drawScaledText(img *image.RGBA, text string, x, y int, scale int, c color.RGBA) {
	sharedutils.DrawScaledText(img, text, x, y, scale, c)
}

// drawScaledChar draws a single character with scaling.
func drawScaledChar(img *image.RGBA, ch rune, x, y int, scale int, c color.RGBA) {
	sharedutils.DrawScaledChar(img, ch, x, y, scale, c)
}
