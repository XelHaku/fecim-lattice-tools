//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the specifications section of the REFERENCE tab.
package gui

import (
	"fmt"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// REFERENCE SPECS SECTION
// ============================================================================

func (ca *CircuitsApp) createReferenceSpecsSection() fyne.CanvasObject {
	// Header with description
	headerLabel := widget.NewRichTextFromMarkdown("**SPECIFICATIONS**: Detailed electrical and physical parameters for all peripheral components (DAC, ADC, TIA) and FeFET cells. Includes array configuration, conversion times, power consumption, and device characteristics.")
	headerLabel.Wrapping = fyne.TextWrapWord

	// Array configuration
	arraySection := ca.createSpecArraySection()

	// DAC specs
	dacSection := ca.createSpecDACSection()

	// ADC specs
	adcSection := ca.createSpecADCSection()

	// TIA specs
	tiaSection := ca.createSpecTIASection()

	// FeFET cell specs
	fefetSection := ca.createSpecFeFETSection()

	// System summary
	summarySection := ca.createSpecSummarySection()

	// Buttons
	exportBtn := widget.NewButton("Export Specs", ca.onExportSpecs)
	compareBtn := widget.NewButton("Compare to GPU", ca.onCompareToGPU)

	ca.specStatusLabel = widget.NewLabel("System specifications")

	buttonBox := container.NewHBox(
		exportBtn,
		compareBtn,
		layout.NewSpacer(),
		ca.specStatusLabel,
	)

	// Layout in a grid with improved visual hierarchy
	arrayHeader := widget.NewLabelWithStyle("ARRAY CONFIGURATION", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	dacHeader := widget.NewLabelWithStyle("DAC SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	adcHeader := widget.NewLabelWithStyle("ADC SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	tiaHeader := widget.NewLabelWithStyle("TIA SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	fefetHeader := widget.NewLabelWithStyle("FeFET CELL SPECIFICATIONS", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	summaryHeader := widget.NewLabelWithStyle("SYSTEM SUMMARY", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	leftCol := container.NewVBox(
		arrayHeader,
		arraySection,
		widget.NewSeparator(),
		dacHeader,
		dacSection,
		widget.NewSeparator(),
		adcHeader,
		adcSection,
	)

	rightCol := container.NewVBox(
		tiaHeader,
		tiaSection,
		widget.NewSeparator(),
		fefetHeader,
		fefetSection,
		widget.NewSeparator(),
		summaryHeader,
		summarySection,
	)

	mainContent := container.NewHBox(
		container.NewPadded(leftCol),
		widget.NewSeparator(),
		container.NewPadded(rightCol),
	)

	return container.NewBorder(
		container.NewVBox(headerLabel, widget.NewSeparator()),
		container.NewVBox(widget.NewSeparator(), buttonBox),
		nil,
		nil,
		container.NewVScroll(container.NewHScroll(mainContent)),
	)
}

func (ca *CircuitsApp) createSpecArraySection() fyne.CanvasObject {
	sizeOptions := []string{"8", "16", "32", "64", "128"}
	ca.specArraySizeSelect = widget.NewSelect(sizeOptions, func(s string) {
		logInput("spec_array_size=%s", s)
		// Update the summary when size changes
		ca.updateSpecSummary()
	})
	ca.specArraySizeSelect.SetSelected("32")

	levelOptions := []string{"2", "4", "8", "16", "30", "32", "64", "128", "256"}
	ca.specQuantLevelSelect = widget.NewSelect(levelOptions, func(s string) {
		logInput("spec_quant_levels=%s", s)
	})
	ca.specQuantLevelSelect.SetSelected("30")

	// Calculate storage
	cells := 32 * 32
	bitsPerCell := math.Log2(30)
	totalBits := float64(cells) * bitsPerCell

	specGrid := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Parameter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Notes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("Array Size"),
		container.NewHBox(ca.specArraySizeSelect, widget.NewLabel("×"), ca.specArraySizeSelect),
		widget.NewLabel(fmt.Sprintf("%d cells", cells)),

		widget.NewLabel("Quantization Levels"),
		ca.specQuantLevelSelect,
		widget.NewLabel(fmt.Sprintf("~%.1f bits/cell", bitsPerCell)),

		widget.NewLabel("Total Storage"),
		widget.NewLabel(fmt.Sprintf("%.0f bits", totalBits)),
		widget.NewLabel(fmt.Sprintf("%d cells × %.1f bits", cells, bitsPerCell)),
	)

	return container.NewPadded(specGrid)
}

func (ca *CircuitsApp) createSpecDACSection() fyne.CanvasObject {
	dacBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specDACBitsSelect = widget.NewSelect(dacBitsOptions, func(s string) {
		logInput("spec_dac_bits=%s", s)
	})
	ca.specDACBitsSelect.SetSelected("8")

	specGrid := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Parameter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Notes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("Count"),
		widget.NewLabel("32"),
		widget.NewLabel("one per column"),

		widget.NewLabel("Resolution"),
		container.NewHBox(ca.specDACBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel("256 levels"),

		widget.NewLabel("Output Range (read)"),
		widget.NewLabel("0V to 1.0V"),
		widget.NewLabel("non-destructive read"),

		widget.NewLabel("Output Range (write)"),
		widget.NewLabel("2V to 5V"),
		widget.NewLabel("exceeds coercive field"),

		widget.NewLabel("Conversion Time"),
		widget.NewLabel("5 ns"),
		widget.NewLabel("digital to analog latency"),

		widget.NewLabel("Power per DAC"),
		widget.NewLabel("0.1 mW"),
		widget.NewLabel("static + dynamic"),

		widget.NewLabel("Total DAC Power"),
		widget.NewLabel("3.2 mW"),
		widget.NewLabel("for 32 DACs"),

		widget.NewLabel("INL"),
		widget.NewLabel("< 0.5 LSB"),
		widget.NewLabel("integral nonlinearity"),

		widget.NewLabel("DNL"),
		widget.NewLabel("< 0.5 LSB"),
		widget.NewLabel("differential nonlinearity"),

		widget.NewLabel("Rise/Fall Time"),
		widget.NewLabel("2-5 ns"),
		widget.NewLabel("signal edge transitions"),
	)

	helpText := widget.NewLabel("DAC converts digital level (0-29) to precise analog voltage for programming FeFET cells")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewPadded(specGrid),
		widget.NewSeparator(),
		container.NewPadded(helpText),
	)
}

func (ca *CircuitsApp) createSpecADCSection() fyne.CanvasObject {
	adcBitsOptions := []string{"4", "5", "6", "7", "8", "10", "12"}
	ca.specADCBitsSelect = widget.NewSelect(adcBitsOptions, func(s string) {
		logInput("spec_adc_bits=%s", s)
	})
	ca.specADCBitsSelect.SetSelected("8")

	specGrid := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Parameter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Notes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("Count"),
		widget.NewLabel("32"),
		widget.NewLabel("one per row"),

		widget.NewLabel("Resolution"),
		container.NewHBox(ca.specADCBitsSelect, widget.NewLabel("bits")),
		widget.NewLabel("256 levels"),

		widget.NewLabel("Input Range"),
		widget.NewLabel("0V to 1.0V"),
		widget.NewLabel("after TIA conversion"),

		widget.NewLabel("Conversion Time"),
		widget.NewLabel("10 ns"),
		widget.NewLabel("analog to digital latency"),

		widget.NewLabel("Power per ADC"),
		widget.NewLabel("0.5 mW"),
		widget.NewLabel("conversion energy"),

		widget.NewLabel("Total ADC Power"),
		widget.NewLabel("16 mW"),
		widget.NewLabel("for 32 ADCs"),

		widget.NewLabel("ENOB"),
		widget.NewLabel("7.5 bits"),
		widget.NewLabel("effective resolution with noise"),

		widget.NewLabel("SNR"),
		widget.NewLabel("46 dB"),
		widget.NewLabel("signal-to-noise ratio"),

		widget.NewLabel("Sample Rate"),
		widget.NewLabel("100 MSPS"),
		widget.NewLabel("samples per second"),
	)

	helpText := widget.NewLabel("ADC digitizes analog current from TIA, converting continuous values to discrete digital levels")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewPadded(specGrid),
		widget.NewSeparator(),
		container.NewPadded(helpText),
	)
}

func (ca *CircuitsApp) createSpecTIASection() fyne.CanvasObject {
	tiaGainOptions := []string{"1", "10", "100"}
	ca.specTIAGainSelect = widget.NewSelect(tiaGainOptions, func(s string) {
		logInput("spec_tia_gain=%s", s)
	})
	ca.specTIAGainSelect.SetSelected("10")

	specGrid := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Parameter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Notes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("Count"),
		widget.NewLabel("32"),
		widget.NewLabel("one per row"),

		widget.NewLabel("Gain (R_f)"),
		container.NewHBox(ca.specTIAGainSelect, widget.NewLabel("kOhm")),
		widget.NewLabel("transimpedance gain"),

		widget.NewLabel("Bandwidth"),
		widget.NewLabel("100 MHz"),
		widget.NewLabel("frequency response"),

		widget.NewLabel("Input Current Range"),
		widget.NewLabel("0 to 100 µA"),
		widget.NewLabel("cell current range"),

		widget.NewLabel("Output Voltage"),
		widget.NewLabel("0 to 1.0 V"),
		widget.NewLabel("V_out = I_in × R_f"),

		widget.NewLabel("Noise"),
		widget.NewLabel("< 1 µA RMS"),
		widget.NewLabel("input-referred noise"),

		widget.NewLabel("Response Time"),
		widget.NewLabel("~2 ns"),
		widget.NewLabel("settling time"),
	)

	helpText := widget.NewLabel("TIA (Transimpedance Amplifier) converts tiny FeFET currents to measurable voltages for ADC")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewPadded(specGrid),
		widget.NewSeparator(),
		container.NewPadded(helpText),
	)
}

func (ca *CircuitsApp) createSpecFeFETSection() fyne.CanvasObject {
	if ca.refCellFootprintLbl == nil {
		ca.refCellFootprintLbl = widget.NewLabel("-")
	}
	if ca.refCellDensityLbl == nil {
		ca.refCellDensityLbl = widget.NewLabel("-")
	}

	specGrid := container.NewGridWithColumns(3,
		widget.NewLabelWithStyle("Parameter", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Value", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Notes", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("Material"),
		widget.NewLabel("HfZrO2 (HZO)"),
		widget.NewLabel("ferroelectric superlattice"),

		widget.NewLabel("Thickness"),
		widget.NewLabel("10 nm"),
		widget.NewLabel("ferroelectric layer"),

		widget.NewLabel("Levels"),
		widget.NewLabel("30 states (claim)"),
		widget.NewLabel("~4.9 bits/cell; pending peer review"),

		widget.NewLabel("Conductance Range"),
		widget.NewLabel("1 µS to 100 µS"),
		widget.NewLabel("programmable range"),

		widget.NewLabel("Read Voltage"),
		widget.NewLabel("0.5 V"),
		widget.NewLabel("empirical example; safe read is assumed V_read << Vc≈Ec×t"),

		widget.NewLabel("Write Voltage"),
		widget.NewLabel("2.0 V to 5.0 V"),
		widget.NewLabel("exceeds coercive field Ec"),

		widget.NewLabel("Write Time"),
		widget.NewLabel("50 ns"),
		widget.NewLabel("polarization switching"),

		widget.NewLabel("Endurance"),
		widget.NewLabel("10^9 cycles (10^12 target)"),
		widget.NewLabel("write/erase lifetime"),

		widget.NewLabel("Retention"),
		widget.NewLabel("10 years"),
		widget.NewLabel("assumed headline value unless calibrated to a specific dataset"),

		widget.NewLabel("Cell Footprint"),
		ca.refCellFootprintLbl,
		widget.NewLabel("from shared/physics CalculateFootprint"),

		widget.NewLabel("Array Density"),
		ca.refCellDensityLbl,
		widget.NewLabel("cells per mm²"),
	)
	ca.updateFootprintReference()

	helpText := widget.NewLabel("Note: Rise/fall times typically 2-10 ns; capacitance 0.1-10 pF; leakage < 1 nW per cell")
	helpText.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		container.NewPadded(specGrid),
		widget.NewSeparator(),
		container.NewPadded(helpText),
	)
}

func (ca *CircuitsApp) createSpecSummarySection() fyne.CanvasObject {
	// Calculate initial summary based on default size (32x32)
	size := 32
	cells := size * size
	throughput := float64(cells) / 20.0 // MACs per ns = GOPS

	// Component summary table
	summaryGrid := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Component", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Count", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Power", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Area", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Latency", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("FeFET Array"),
		widget.NewLabel(fmt.Sprintf("%d", cells)),
		widget.NewLabel("0.1 mW"),
		widget.NewLabel("0.01 mm²"),
		widget.NewLabel("5 ns"),

		widget.NewLabel("DACs"),
		widget.NewLabel(fmt.Sprintf("%d", size)),
		widget.NewLabel("3.2 mW"),
		widget.NewLabel("0.02 mm²"),
		widget.NewLabel("10 ns"),

		widget.NewLabel("TIAs"),
		widget.NewLabel(fmt.Sprintf("%d", size)),
		widget.NewLabel("1.6 mW"),
		widget.NewLabel("0.01 mm²"),
		widget.NewLabel("11 ns"),

		widget.NewLabel("ADCs"),
		widget.NewLabel(fmt.Sprintf("%d", size)),
		widget.NewLabel("16 mW"),
		widget.NewLabel("0.04 mm²"),
		widget.NewLabel("50 ns"),

		widget.NewLabel("Control"),
		widget.NewLabel("1"),
		widget.NewLabel("0.5 mW"),
		widget.NewLabel("0.01 mm²"),
		widget.NewLabel("2 ns"),

		widget.NewLabelWithStyle("TOTAL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		widget.NewLabelWithStyle("21.4 mW", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("0.09 mm²", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("76 ns", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Performance metrics
	perfGrid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Throughput:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%d MACs (Ops) / 76ns = %.1f GOPS", cells, throughput)),

		widget.NewLabelWithStyle("Efficiency:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%.1f GOPS / 21.4 mW = %d GOPS/W", throughput, int(throughput*1000/21.4))),
	)

	ca.specSummaryLabel = container.NewVBox(
		container.NewPadded(summaryGrid),
		widget.NewSeparator(),
		container.NewPadded(perfGrid),
	)

	return ca.specSummaryLabel
}

func (ca *CircuitsApp) updateSpecSummary() {
	if ca.specSummaryLabel == nil || ca.specArraySizeSelect == nil {
		return
	}

	// Get current array size
	var size int
	fmt.Sscanf(ca.specArraySizeSelect.Selected, "%d", &size)
	if size == 0 {
		size = 32 // default
	}

	cells := size * size
	throughput := float64(cells) / 76.0 // MACs per ns = GOPS

	// Component summary table
	summaryGrid := container.NewGridWithColumns(5,
		widget.NewLabelWithStyle("Component", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Count", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Power", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Area", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("Latency", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),

		widget.NewLabel("FeFET Array"),
		widget.NewLabel(fmt.Sprintf("%d", cells)),
		widget.NewLabel("0.1 mW"),
		widget.NewLabel("0.01 mm²"),
		widget.NewLabel("5 ns"),

		widget.NewLabel("DACs"),
		widget.NewLabel(fmt.Sprintf("%d", size)),
		widget.NewLabel("3.2 mW"),
		widget.NewLabel("0.02 mm²"),
		widget.NewLabel("10 ns"),

		widget.NewLabel("TIAs"),
		widget.NewLabel(fmt.Sprintf("%d", size)),
		widget.NewLabel("1.6 mW"),
		widget.NewLabel("0.01 mm²"),
		widget.NewLabel("11 ns"),

		widget.NewLabel("ADCs"),
		widget.NewLabel(fmt.Sprintf("%d", size)),
		widget.NewLabel("16 mW"),
		widget.NewLabel("0.04 mm²"),
		widget.NewLabel("50 ns"),

		widget.NewLabel("Control"),
		widget.NewLabel("1"),
		widget.NewLabel("0.5 mW"),
		widget.NewLabel("0.01 mm²"),
		widget.NewLabel("2 ns"),

		widget.NewLabelWithStyle("TOTAL", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(""),
		widget.NewLabelWithStyle("21.4 mW", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("0.09 mm²", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabelWithStyle("76 ns", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
	)

	// Performance metrics
	perfGrid := container.NewGridWithColumns(2,
		widget.NewLabelWithStyle("Throughput:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%d MACs (Ops) / 76ns = %.1f GOPS", cells, throughput)),

		widget.NewLabelWithStyle("Efficiency:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel(fmt.Sprintf("%.1f GOPS / 21.4 mW = %d GOPS/W", throughput, int(throughput*1000/21.4))),
	)

	newContent := container.NewVBox(
		container.NewPadded(summaryGrid),
		widget.NewSeparator(),
		container.NewPadded(perfGrid),
	)

	// Replace the content
	ca.specSummaryLabel.Objects = newContent.Objects
	ca.specSummaryLabel.Refresh()
}

func (ca *CircuitsApp) onExportSpecs() {
	logAction("specs_export")
	// Export specifications to JSON and CSV
	ca.exportSimulationData()
}

func (ca *CircuitsApp) onCompareToGPU() {
	logAction("specs_compare_gpu")
	// Show comparison summary in status label
	sharedwidgets.SafeDo(func() {
		ca.specStatusLabel.SetText("FeFET vs GPU: ~6.6x faster than CPU (76ns vs 500ns). FeFET per-MAC energy ~46 fJ in this model.")
	})
}
