//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device simulation view that replaces separate WRITE/READ/COMPUTE modes.
package gui

import (
	"fmt"
	"image/color"
	"math"
	"math/rand"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	configphysics "fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedexport "fecim-lattice-tools/shared/export"
	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// createFixedHeightContainer wraps content in a container with minimum height
func createFixedHeightContainer(content fyne.CanvasObject, minHeight float32) fyne.CanvasObject {
	// Create invisible spacer rectangle with minimum size
	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(0, minHeight))

	// Stack content over spacer - content expands to spacer's min height
	return container.NewStack(spacer, content)
}

const (
	minSenseRfKOhm      = 1.0
	maxSenseRfKOhm      = 1000.0
	minSenseADCVref     = 0.0
	maxSenseADCVref     = 1.5
	minSenseADCVrefSpan = 1e-3
)

type senseMeasurementPreset struct {
	Name string
	Rf   float64
	Vmin float64
	Vmax float64
}

var senseMeasurementPresets = []senseMeasurementPreset{
	{Name: "Balanced", Rf: 10.0, Vmin: 0.00, Vmax: 1.00},
	{Name: "High Sensitivity", Rf: 50.0, Vmin: 0.00, Vmax: 1.00},
	{Name: "Wide Current Range", Rf: 5.0, Vmin: 0.00, Vmax: 1.20},
	{Name: "Low-Current Focus", Rf: 100.0, Vmin: 0.10, Vmax: 0.90},
	{Name: "Ultra-Low Current", Rf: 500.0, Vmin: 0.20, Vmax: 0.80},
}

const (
	customSensePresetName = "Custom"
	actionLabelProgram    = "Program Cell"
	actionLabelCompute    = "Run MVM"
	actionLabelUndo       = "Undo"
	actionLabelRandom     = "Random Array"
	actionLabelReset      = "Reset Array"
	actionLabelExport     = "Export"
	labelOverlay          = "Overlay:"
	labelZoom             = "Zoom:"

	// ReadVoltageFraction is the fraction of max read range used as the safe read voltage.
	ReadVoltageFraction = 0.4

	// MinReadVoltageThreshold is the threshold below which the read voltage is replaced by FallbackReadVoltage.
	MinReadVoltageThreshold = 0.1

	// FallbackReadVoltage is the default read voltage when the computed value is too low.
	FallbackReadVoltage = 0.2

	// MaxDACVoltageV is the maximum voltage that can be applied via DAC input.
	MaxDACVoltageV = 2.0

	// MaxDACCode is the maximum digital input code for DAC (8-bit).
	MaxDACCode = 255

	// DACCodeRange is the total number of DAC codes (0 to MaxDACCode).
	DACCodeRange = 256
)

func formatCurrentA(currentA float64) string {
	return formatSignedScaled(currentA, []scaledUnit{
		{unit: "A", scale: 1.0},
		{unit: "mA", scale: 1e-3},
		{unit: "uA", scale: 1e-6},
		{unit: "nA", scale: 1e-9},
		{unit: "pA", scale: 1e-12},
	})
}

// techNodeSelectorParams returns (selectorRon, leakageS, nodeNm) for a technology node name.
func techNodeSelectorParams(name string) (selectorRon, leakageS, nodeNm float64) {
	switch name {
	case "14nm":
		return 0.8e3, 5e-11, 14
	case "28nm":
		return 1.5e3, 2e-11, 28
	case "65nm":
		return 3.5e3, 5e-12, 65
	default:
		return 7.0e3, 1e-12, 130
	}
}

// safeReadVoltage computes a safe read voltage from the device's read range.
// Returns ReadVoltageFraction * readRange.Max, with a floor of FallbackReadVoltage.
func (ca *CircuitsApp) safeReadVoltage() float64 {
	v := ca.deviceState.GetReadRange().Max * ReadVoltageFraction
	if v < MinReadVoltageThreshold {
		v = FallbackReadVoltage
	}
	return v
}

// parseBitsFromSelection extracts the bit width from a selector string like "5-bit (32)".
// Returns the parsed bits and true, or 0 and false if parsing fails.
func parseBitsFromSelection(selected string) (int, bool) {
	var bits int
	_, err := fmt.Sscanf(selected, "%d-bit", &bits)
	if err != nil {
		return 0, false
	}
	return bits, true
}

type scaledUnit struct {
	unit  string
	scale float64
}

func formatSignedScaled(value float64, units []scaledUnit) string {
	absValue := math.Abs(value)
	if absValue < 1e-12 {
		return fmt.Sprintf("0 %s", units[0].unit)
	}

	chosen := units[len(units)-1]
	for _, candidate := range units {
		if absValue >= candidate.scale {
			chosen = candidate
			break
		}
	}
	scaled := value / chosen.scale
	absScaled := math.Abs(scaled)
	format := "%+.3f"
	switch {
	case absScaled >= 100:
		format = "%+.0f"
	case absScaled >= 10:
		format = "%+.1f"
	case absScaled >= 1:
		format = "%+.2f"
	}
	return fmt.Sprintf(format+" %s", scaled, chosen.unit)
}

// ============================================================================
// UNIFIED DEVICE SIMULATION VIEW
// ============================================================================

// createUnifiedView creates the unified device simulation view
// Layout: Controls at TOP (toolbar), Array canvas in CENTER (expands), minimal status at BOTTOM
func (ca *CircuitsApp) createUnifiedView() fyne.CanvasObject {
	ca.beginUIRefreshSuspension()
	resumeRefresh := true
	defer func() {
		if resumeRefresh {
			ca.endUIRefreshSuspension()
		}
	}()

	// Initialize device state
	ca.deviceState = NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)
	ca.operationsStatusLabel = widget.NewLabel("Ready")
	ca.operationsStatusLabel.Truncation = fyne.TextTruncateEllipsis

	// In passive (0T1R) mode, all WLs are always active - no transistor gating
	if ca.architecture == sharedwidgets.Architecture0T1R {
		ca.deviceState.SetPassiveMode(true)
	}

	// ============================================================
	// TOP: Compact Toolbar (~100px total)
	// ============================================================

	// Row 1: Config + Mode + Architecture (unified row)
	configModeRow := ca.createUnifiedConfigModeRow()
	// Make the top rows resilient on narrow windows: allow horizontal scrolling
	// instead of overlap/truncation.
	configModeRow = container.NewHScroll(configModeRow)

	// Row 2: Action buttons
	actionRow := ca.createUnifiedActionRow()
	actionRow = container.NewHScroll(actionRow)

	// Mode-specific panels (only visible when mode is active)
	writePanelContent := ca.createCompactWritePanel()
	ca.writeModePanel = container.NewVBox(writePanelContent)
	ca.writeModePanel.Hide() // Hidden by default (READ mode)

	computePanelContent := ca.createComputeModePanel()
	ca.computeModePanel = container.NewVBox(computePanelContent)
	ca.computeModePanel.Hide() // Hidden by default (READ mode)

	// Sense panel (visible in READ/COMPUTE)
	sensePanelContent := ca.createSensePanel()
	ca.sensePanel = container.NewVBox(sensePanelContent)
	ca.sensePanel.Hide()

	// Stack panels (only one visible at a time)
	modePanelStack := container.NewStack(ca.writeModePanel, ca.computeModePanel)

	// Toolbar section: Config/Mode row + Action row + Mode panel + Sense panel
	toolbarSection := container.NewVBox(
		configModeRow,
		actionRow,
		modePanelStack,
		ca.sensePanel,
	)

	// ============================================================
	// CENTER: Array Canvas (expands to fill available space)
	// ============================================================

	// Create tappable array canvas
	tappableArray := NewUnifiedTappableCanvas(ca, ca.drawUnifiedArray, ca.onUnifiedCellTapped)
	tappableArray.SetMinSize(fyne.NewSize(400, 300)) // Reduced min size - will expand
	sharedwidgets.SetAccessibleLabel(tappableArray, "Crossbar array visualization")
	ca.sharedArrayCanvas = tappableArray.raster

	// Initialize empty WL checks array (some code may reference it)
	ca.unifiedWLChecks = make([]*widget.Check, 0)

	// Cell info display (updated on cell click)
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")
	ca.sharedCellDisplayToggle = widget.NewButton("Show V / Show I: V", func() {
		ca.showCurrentInCellInfo = !ca.showCurrentInCellInfo
		mode := "V"
		if ca.showCurrentInCellInfo {
			mode = "I"
		}
		ca.sharedCellDisplayToggle.SetText(fmt.Sprintf("Show V / Show I: %s", mode))
		ca.updateCellInfo()
	})
	ca.sharedCellDisplayToggle.Importance = widget.LowImportance

	// Array info (updated on resize)
	totalCells := ca.arrayRows * ca.arrayCols
	bitCapacity := float64(totalCells) * sharedphysics.BitsPerCell
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("%dx%d array | %d levels | ~%.0f bits",
		ca.arrayRows, ca.arrayCols, ca.quantLevels, bitCapacity))

	// Truncate labels that may carry long dynamic text at narrow widths
	ca.sharedCellInfoLabel.Truncation = fyne.TextTruncateEllipsis
	ca.sharedArrayInfoLabel.Truncation = fyne.TextTruncateEllipsis

	// Compact info row below canvas
	infoRow := container.NewHBox(
		ca.sharedCellDisplayToggle,
		ca.sharedCellInfoLabel,
		layout.NewSpacer(),
		ca.sharedArrayInfoLabel,
	)

	// Array section: canvas + info row
	arraySection := container.NewBorder(nil, infoRow, nil, nil, tappableArray)

	// ============================================================
	// BOTTOM: Minimal Status Bar (~20px)
	// ============================================================

	// Status/info label
	ca.operationsModeHelp = widget.NewLabel("Click cells to select")
	ca.operationsModeHelp.TextStyle = fyne.TextStyle{Italic: true}
	ca.operationsModeHelp.Truncation = fyne.TextTruncateEllipsis

	// Initialize architecture info (compact single-line)
	ca.passiveVoltagePanel = ca.createCompactPassivePanel()
	ca.activeVoltagePanel = ca.createCompactActivePanel()
	ca.passiveVoltagePanel.Hide() // Hidden initially (1T1R default)

	pipelineLabel := widget.NewLabel("DAC -> Array -> TIA -> ADC")
	pipelineLabel.Truncation = fyne.TextTruncateEllipsis
	statusBar := container.NewBorder(nil, nil, nil, pipelineLabel, ca.operationsModeHelp)

	// Initialize button states for default READ mode
	ca.updateActionButtons()
	ca.updateModePanels(ca.deviceState.GetOperationMode())

	ca.endUIRefreshSuspension()
	resumeRefresh = false
	ca.recomputeAndRefreshNow()

	return container.NewBorder(
		toolbarSection, // top: compact controls
		statusBar,      // bottom: minimal status
		nil, nil,
		arraySection, // center: array canvas (EXPANDS)
	)
}

// createUnifiedConfigModeRow creates a single row with config + mode buttons + architecture
func (ca *CircuitsApp) createUnifiedConfigModeRow() fyne.CanvasObject {
	// Material selector
	materialSelector := ca.createMaterialSelector()

	// Array size selector
	arraySizeSelector := ca.createArraySizeSelector()

	// Technology node selector
	techNodeSelector := ca.createTechnologyNodeSelector()

	// DAC bits selector
	dacBitsSelector := ca.createDACBitsSelector()

	// ADC bits selector
	adcBitsSelector := ca.createADCBitsSelector()

	// Coupling model toggle
	couplingToggle := ca.createCouplingToggle()

	// Mode buttons
	ca.modeReadBtn = widget.NewButton("READ", func() { ca.setOperationMode(OpModeRead) })
	ca.modeWriteBtn = widget.NewButton("WRITE", func() { ca.setOperationMode(OpModeWrite) })
	ca.modeComputeBtn = widget.NewButton("COMPUTE", func() { ca.setOperationMode(OpModeCompute) })
	modeInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Operation Modes",
			"READ: Safe read voltage on selected column; no programming.\n\n"+
				"WRITE: Arms write range and WL gating; no voltage until Program Cell.\n\n"+
				"COMPUTE: Applies input vector across all rows for MVM.", ca.window)
	})
	modeInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(modeInfo, "Operation modes info")

	// Set initial highlight (READ mode by default)
	ca.modeReadBtn.Importance = widget.HighImportance

	// Architecture toggle
	archToggle := ca.createArchitectureToggle()

	// Single row: Material | Array | ADC | Sep | Mode buttons | Spacer | Architecture
	return container.NewHBox(
		materialSelector,
		techNodeSelector,
		arraySizeSelector,
		dacBitsSelector,
		adcBitsSelector,
		couplingToggle,
		widget.NewSeparator(),
		widget.NewLabel("Mode:"),
		ca.modeReadBtn,
		ca.modeWriteBtn,
		ca.modeComputeBtn,
		modeInfo,
		layout.NewSpacer(),
		archToggle,
	)
}

// createUnifiedActionRow creates the action buttons row
func (ca *CircuitsApp) createUnifiedActionRow() fyne.CanvasObject {
	// Primary action buttons
	ca.actionWriteCellBtn = widget.NewButton(actionLabelProgram, func() { ca.onUnifiedProgram() })
	sharedwidgets.SetAccessibleLabel(ca.actionWriteCellBtn, "Program selected cell")
	ca.actionWriteCellBtn.Importance = widget.HighImportance
	programInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Program Cell",
			"Apply DAC write pulse to selected cell (ISPP).\nPassive arrays: full column write (all WL=0V, selected BL=−V_write).", ca.window)
	})
	programInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(programInfo, "Program cell info")

	ca.actionComputeBtn = widget.NewButton(actionLabelCompute, func() {
		ca.onUnifiedCompute()
	})
	sharedwidgets.SetAccessibleLabel(ca.actionComputeBtn, "Run matrix-vector multiply")

	// Utility buttons
	ca.undoHistoryBtn = widget.NewButton(actionLabelUndo, func() {
		ca.onUndo()
	})
	sharedwidgets.SetAccessibleLabel(ca.undoHistoryBtn, "Undo last array operation")
	ca.undoHistoryBtn.Disable()

	ca.actionRandomArrayBtn = widget.NewButton(actionLabelRandom, func() {
		ca.onUnifiedRandomArray()
	})
	sharedwidgets.SetAccessibleLabel(ca.actionRandomArrayBtn, "Fill array with random conductance levels")

	ca.actionResetArrayBtn = widget.NewButton(actionLabelReset, func() {
		ca.onUnifiedReset()
	})
	sharedwidgets.SetAccessibleLabel(ca.actionResetArrayBtn, "Reset all cells to default level")

	// Tools button with status indicators
	toolWidgets := sharedwidgets.NewToolValidationWidgets(sharedwidgets.ToolValidationOptions{
		Window:              ca.window,
		ButtonLabel:         "Tools",
		DialogTitle:         "Validation Tools",
		StatusLabelMode:     sharedwidgets.ToolStatusSymbolOnly,
		MessageStyle:        sharedwidgets.ToolMessageASCII,
		IncludeInstall:      true,
		IncludeInstallNotes: true,
	})

	// Zoom controls
	ca.zoomLabel = widget.NewLabel("100%")
	ca.zoomSlider = widget.NewSlider(0.5, 3.0)
	ca.zoomSlider.Step = 0.1
	ca.zoomSlider.Value = 1.0
	ca.zoomSlider.OnChanged = func(v float64) {
		ca.mu.Lock()
		ca.zoomLevel = v
		ca.mu.Unlock()
		logInput("zoom=%.2f", v)
		ca.zoomLabel.SetText(fmt.Sprintf("%.0f%%", v*100))
		sharedwidgets.SafeDo(func() {
			if ca.sharedArrayCanvas != nil {
				ca.sharedArrayCanvas.Refresh()
			}
		})
	}

	zoomOutBtn := widget.NewButton("−", func() {
		ca.zoomSlider.SetValue(ca.zoomSlider.Value - ca.zoomSlider.Step)
	})
	sharedwidgets.SetAccessibleLabel(zoomOutBtn, "Zoom out")
	zoomOutBtn.Importance = widget.LowImportance
	zoomInBtn := widget.NewButton("+", func() {
		ca.zoomSlider.SetValue(ca.zoomSlider.Value + ca.zoomSlider.Step)
	})
	sharedwidgets.SetAccessibleLabel(zoomInBtn, "Zoom in")
	zoomInBtn.Importance = widget.LowImportance
	zoomSliderWrap := container.NewGridWrap(fyne.NewSize(320, 36), ca.zoomSlider)

	ca.actionFitBtn = widget.NewButton("Fit", func() {
		logAction("button_zoom_fit")
		ca.zoomSlider.SetValue(1.0)
	})
	sharedwidgets.SetAccessibleLabel(ca.actionFitBtn, "Reset zoom to 100 percent")

	// Export button
	exportBtn := sharedexport.CreateExportButton(actionLabelExport, func() {
		logAction("button_export")
		ca.exportSimulationData()
	}, ca.window)
	sharedwidgets.SetAccessibleLabel(exportBtn, "Export current simulation state")

	ca.readOverlaySelect = widget.NewSelect([]string{"Off", "Vcell", "Icell"}, func(mode string) {
		if mode == "" {
			mode = "Off"
		}
		ca.mu.Lock()
		ca.readOverlayMode = mode
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
	})
	ca.readOverlaySelect.SetSelected(ca.readOverlayMode)
	sharedwidgets.SetAccessibleLabel(ca.readOverlaySelect, "Read overlay mode")
	sharedwidgets.SetAccessibleLabel(ca.zoomSlider, "Array zoom level")

	// Row: Program Cell | MVM | Sep | Undo | Random Array | Reset Array | Export | Overlay | Sep | Zoom controls | Spacer | Tools status
	return container.NewHBox(
		ca.actionWriteCellBtn,
		programInfo,
		ca.actionComputeBtn,
		widget.NewSeparator(),
		ca.undoHistoryBtn,
		ca.actionRandomArrayBtn,
		ca.actionResetArrayBtn,
		exportBtn,
		widget.NewLabel(labelOverlay),
		ca.readOverlaySelect,
		widget.NewSeparator(),
		widget.NewLabel(labelZoom),
		zoomOutBtn,
		zoomSliderWrap,
		zoomInBtn,
		ca.zoomLabel,
		ca.actionFitBtn,
		layout.NewSpacer(),
		toolWidgets.CrossSimStatus,
		toolWidgets.BadCrossbarStatus,
		toolWidgets.Button,
	)
}

// createCouplingToggle creates the Ideal vs Tier A coupling toggle.
func (ca *CircuitsApp) createCouplingToggle() fyne.CanvasObject {
	current := arraysim.CouplingIdeal
	if ca.deviceState != nil {
		current = ca.deviceState.GetCouplingMode()
	}

	labelToMode := map[string]arraysim.CouplingMode{
		"Ideal":  arraysim.CouplingIdeal,
		"Tier-A": arraysim.CouplingTierA,
		"Tier-B": arraysim.CouplingTierB,
	}
	modeToLabel := map[arraysim.CouplingMode]string{
		arraysim.CouplingIdeal: "Ideal",
		arraysim.CouplingTierA: "Tier-A",
		arraysim.CouplingTierB: "Tier-B",
	}

	selector := widget.NewSelect([]string{"Ideal", "Tier-A", "Tier-B"}, func(selected string) {
		mode, ok := labelToMode[selected]
		if !ok {
			mode = arraysim.CouplingIdeal
		}
		if ca.deviceState != nil {
			ca.deviceState.SetCouplingMode(mode)
		}
		ca.recomputeAndRefresh()
	})
	selector.SetSelected(modeToLabel[current])
	ca.couplingTierSelect = selector

	couplingInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Fidelity Tier",
			"Ideal: no array parasitics (fastest).\n\nTier-A: approximate IR-drop + sneak coupling.\n\nTier-B: DC nodal reference solver (highest fidelity, slower, small arrays).", ca.window)
	})
	couplingInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(couplingInfo, "Fidelity tier info")
	return container.NewHBox(widget.NewLabel("Fidelity:"), selector, couplingInfo)
}

// ValidArraySizes defines the supported array dimensions
var ValidArraySizes = []int{1, 2, 4, 8, 16, 32, 64, 128}

// createArraySizeSelector creates a dropdown to select array size (1x1 to 128x128)
func (ca *CircuitsApp) createArraySizeSelector() fyne.CanvasObject {
	options := make([]string, len(ValidArraySizes))
	for i, size := range ValidArraySizes {
		options[i] = fmt.Sprintf("%dx%d", size, size)
	}

	selector := widget.NewSelect(options, func(selected string) {
		// Parse size from "NxN" format
		var rows, cols int
		n, err := fmt.Sscanf(selected, "%dx%d", &rows, &cols)
		if err != nil || n != 2 {
			dialog.ShowError(fmt.Errorf("invalid array size format %q; expected NxN", selected), ca.window)
			return
		}
		if rows <= 0 || rows > MaxArraySize || cols <= 0 || cols > MaxArraySize {
			dialog.ShowError(fmt.Errorf("array size %dx%d is outside supported range 1..%d", rows, cols, MaxArraySize), ca.window)
			return
		}
		logInput("array_size=%s", selected)
		ca.resizeArray(rows, cols)
	})

	// Set default selection
	selector.SetSelected(fmt.Sprintf("%dx%d", ca.arrayRows, ca.arrayCols))
	sharedwidgets.SetAccessibleLabel(selector, "Array size selector")

	sizeInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Array Size",
			"Crossbar array dimensions (rows x columns).\nLarger arrays increase compute parallelism but worsen sneak-path and IR-drop effects.", ca.window)
	})
	sizeInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(sizeInfo, "Array size info")
	return container.NewHBox(widget.NewLabel("Size:"), selector, sizeInfo)
}

// resizeArray changes the array dimensions and reinitializes all related state
func (ca *CircuitsApp) resizeArray(rows, cols int) {
	ca.mu.Lock()

	// Skip if no change
	if rows == ca.arrayRows && cols == ca.arrayCols {
		ca.mu.Unlock()
		return
	}

	ca.arrayRows = rows
	ca.arrayCols = cols

	// Reinitialize weight matrix
	ca.arrayWeights = make([][]int, rows)
	ca.halfSelectResidue = make([][]float64, rows)
	for i := range ca.arrayWeights {
		ca.arrayWeights[i] = make([]int, cols)
		ca.halfSelectResidue[i] = make([]float64, cols)
		// Initialize to mid-level (like a fresh device)
		for j := range ca.arrayWeights[i] {
			ca.arrayWeights[i][j] = ca.quantLevels / 2
		}
	}

	// Reinitialize input/output vectors
	ca.inputVector = make([]int, cols)
	ca.outputVector = make([]float64, rows)

	ca.mu.Unlock()

	// Reset disturb engine on resize (will be lazy-reinited with new dimensions)
	ca.writeDisturbEngine = nil

	// Reinitialize device state with new dimensions
	ca.deviceState = NewDeviceState(rows, cols, ca.tia, ca.adc)

	// Preserve architecture mode
	if ca.architecture == sharedwidgets.Architecture0T1R {
		ca.deviceState.SetPassiveMode(true)
	}

	// Update compute mode panel inputs
	ca.rebuildComputeInputs()

	// Update status (must be on UI thread)
	totalCells := rows * cols
	bitCapacity := float64(totalCells) * sharedphysics.BitsPerCell // ~4.9 bits per 30-level cell
	sharedwidgets.SafeDo(func() {
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("Array resized: %dx%d (%d cells, %.1f bits)",
				rows, cols, totalCells, bitCapacity))
		}
	})

	// Refresh display
	ca.recomputeAndRefresh()
}

// Note: createSignalChainHeader removed - functionality moved to createConfigurationSection
// and the signal chain label is now inline in createUnifiedView

// createMaterialSelector creates the ferroelectric material selection button (like module 1)
func (ca *CircuitsApp) createMaterialSelector() fyne.CanvasObject {
	// Get initial material name
	initialName := ca.deviceState.GetMaterialName()

	// Material button - shows current material, opens picker on click
	ca.materialBtn = widget.NewButton(initialName, func() {
		// Get current material ID for pre-selection in picker
		currentID := ca.getCurrentMaterialID()
		sharedwidgets.ShowMaterialPicker(ca.window, currentID, func(materialID string, mat *configphysics.Material) {
			if mat == nil {
				return
			}
			// Convert config material to HZO material and set it
			materials := sharedphysics.AllMaterials()
			for _, m := range materials {
				if m.Name == mat.Name {
					ca.deviceState.SetMaterial(m)
					ca.materialBtn.SetText(m.Name)
					logInput("material=%s", m.Name)
					ca.updateDACRangeModeLabel()
					ca.recomputeAndRefresh()
					ca.operationsStatusLabel.SetText(fmt.Sprintf("Material: %s (Vc=%.2fV)", m.Name, m.CoerciveVoltage()))
					break
				}
			}
		})
	})

	materialInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Material",
			"Ferroelectric material preset.\nDetermines coercive voltage, remnant polarization, and write voltage range.", ca.window)
	})
	materialInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(materialInfo, "Material info")
	return container.NewHBox(widget.NewLabel("Material:"), ca.materialBtn, materialInfo)
}

func (ca *CircuitsApp) createTechnologyNodeSelector() fyne.CanvasObject {
	options := []string{"130nm", "65nm", "28nm", "14nm"}
	selector := widget.NewSelect(options, func(selected string) {
		ca.applyTechnologyNode(selected)
	})
	selector.SetSelected("130nm")
	ca.techNodeSelect = selector
	sharedwidgets.SetAccessibleLabel(selector, "Technology node selector")

	info := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Technology Node",
			"Process node updates wire resistance, selector series resistance, and leakage assumptions used by the coupled read model.", ca.window)
	})
	info.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(info, "Technology node info")

	return container.NewHBox(widget.NewLabel("Node:"), selector, info)
}

func (ca *CircuitsApp) applyTechnologyNode(node string) {
	tech := sharedphysics.TechnologyNodeFromName(node)
	geom := arraysim.CellGeometry{
		PitchX:           tech.CellPitchX,
		PitchY:           tech.CellRowHeight,
		WireWidth:        tech.MetalWidth,
		WireThickness:    tech.MetalThickness,
		MetalResistivity: tech.MetalResistivity,
	}

	selectorRon, leakageS, nodeNm := techNodeSelectorParams(tech.Name)
	ca.selectedTechNodeNm = nodeNm

	if ca.deviceState != nil {
		ca.deviceState.SetCellGeometry(geom)
		ca.deviceState.SetWireParams(arraysim.WireParams{})
		ca.deviceState.SetSelectorSeriesParams(true, selectorRon, leakageS)
	}
	ca.updateFootprintReference()
	ca.recomputeAndRefresh()
}

func (ca *CircuitsApp) updateFootprintReference() {
	if ca.refCellFootprintLbl == nil || ca.refCellDensityLbl == nil {
		return
	}
	arch := ca.architecture
	if arch == "" {
		arch = sharedwidgets.Architecture0T1R
	}
	fp := sharedphysics.CalculateFootprint(arch, ca.selectedTechNodeNm)
	ca.refCellFootprintLbl.SetText(fmt.Sprintf("%.4f µm²", fp.TotalArea*1e12))
	ca.refCellDensityLbl.SetText(fmt.Sprintf("%.2f M cells/mm²", fp.DensityCellsPerMM2()/1e6))
}

// getCurrentMaterialID returns the ID of the currently selected material
func (ca *CircuitsApp) getCurrentMaterialID() string {
	matName := ca.deviceState.GetMaterialName()
	// Load config and find ID by matching display name
	cfg, err := configphysics.Load()
	if err != nil {
		return "fecim_hzo" // Default fallback
	}
	for id, mat := range cfg.Materials {
		if mat.Name == matName {
			return id
		}
	}
	return "fecim_hzo" // Default fallback
}

// createDACBitsSelector creates a dropdown to select DAC resolution (4-8 bits)
func (ca *CircuitsApp) createDACBitsSelector() fyne.CanvasObject {
	options := []string{"4-bit (16)", "5-bit (32)", "6-bit (64)", "7-bit (128)", "8-bit (256)"}

	selector := widget.NewSelect(options, func(selected string) {
		bits, ok := parseBitsFromSelection(selected)
		if !ok {
			dialog.ShowError(fmt.Errorf("unsupported DAC selection %q", selected), ca.window)
			return
		}

		logInput("dac_bits=%d", bits)
		ca.dacBits = bits
		ca.deviceState.SetDACBits(bits)
		ca.recomputeAndRefresh()
		levels := 1 << bits
		ca.operationsStatusLabel.SetText(fmt.Sprintf("DAC: %d-bit (%d levels, 0-%d)", bits, levels, levels-1))
	})

	selector.SetSelected("5-bit (32)")
	sharedwidgets.SetAccessibleLabel(selector, "DAC resolution selector")

	dacInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("DAC Resolution",
			"Digital-to-analog converter bit width.\nHigher resolution improves write-voltage granularity but increases area and energy.", ca.window)
	})
	dacInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(dacInfo, "DAC resolution info")
	return container.NewHBox(widget.NewLabel("DAC:"), selector, dacInfo)
}

// createADCBitsSelector creates a dropdown to select ADC resolution (5-8 bits)
func (ca *CircuitsApp) createADCBitsSelector() fyne.CanvasObject {
	options := []string{"5-bit (32)", "6-bit (64)", "7-bit (128)", "8-bit (256)"}

	selector := widget.NewSelect(options, func(selected string) {
		bits, ok := parseBitsFromSelection(selected)
		if !ok {
			dialog.ShowError(fmt.Errorf("unsupported ADC selection %q", selected), ca.window)
			return
		}
		logInput("adc_bits=%d", bits)
		ca.deviceState.SetADCBits(bits)
		ca.recomputeAndRefresh()
		levels := 1 << bits
		ca.operationsStatusLabel.SetText(fmt.Sprintf("ADC: %d-bit (%d levels, 0-%d)", bits, levels, levels-1))
	})

	// Set default selection
	selector.SetSelected("5-bit (32)")
	sharedwidgets.SetAccessibleLabel(selector, "ADC resolution selector")

	adcInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("ADC Resolution",
			"Analog-to-digital converter bit width.\nHigher resolution distinguishes more conductance levels but increases area and power.", ca.window)
	})
	adcInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(adcInfo, "ADC resolution info")
	return container.NewHBox(widget.NewLabel("ADC:"), selector, adcInfo)
}

// createDACInputSection creates the DAC status and manual control
func (ca *CircuitsApp) createDACInputSection() fyne.CanvasObject {
	// Initialize DAC entries array (used by updateDACEntries but not displayed)
	// Use all columns - no limit
	ca.unifiedDACEntries = make([]*widget.Entry, ca.arrayCols)
	ca.unifiedDACLabels = make([]*widget.Label, ca.arrayCols)

	// Range mode indicator - shows current DAC voltage range based on operation mode
	// Note: DAC range is set automatically by mode (READ/WRITE/COMPUTE)
	// Random input is available in COMPUTE mode panel
	ca.dacRangeLabel = widget.NewLabel("DAC: Read Range")
	ca.dacRangeLabel.TextStyle = fyne.TextStyle{Italic: true}

	// "Set All" entry for bulk voltage (manual override)
	allEntry := widget.NewEntry()
	allEntry.SetPlaceHolder("0.50")
	allEntry.OnSubmitted = func(s string) {
		logInput("dac_all_submit=%s", s)
		ca.setAllUnifiedDACVoltages(s)
	}

	return container.NewHBox(
		ca.dacRangeLabel,
		layout.NewSpacer(),
		widget.NewLabel("Set All (V):"), allEntry,
	)
}

// updateDACRangeModeLabel updates the DAC range mode indicator based on operation mode
// Only shown in WRITE mode; hidden in READ and COMPUTE modes
func (ca *CircuitsApp) updateDACRangeModeLabel() {
	if ca.dacRangeLabel == nil || ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()
	rangeMode := ca.deviceState.GetDACRangeMode()
	currentRange := ca.deviceState.GetCurrentVoltageRange()

	sharedwidgets.SafeDo(func() {
		// Only show in WRITE mode where voltage range matters
		if mode == OpModeWrite {
			var text string
			if rangeMode == DACRangeWrite {
				text = fmt.Sprintf("DAC: Write (-%.1fV RESET .. +%.1fV SET)", currentRange.Max, currentRange.Max)
			} else {
				text = fmt.Sprintf("DAC: Read (0-%.1fV)", currentRange.Max)
			}
			ca.dacRangeLabel.SetText(text)
			ca.dacRangeLabel.Show()
		} else {
			// Hide in READ and COMPUTE modes
			ca.dacRangeLabel.Hide()
		}
	})
}

// setOperationMode sets the operation mode and configures WL/DAC accordingly.
//
// Physics meaning:
//   - READ: bias a *single selected column* with a small, non-destructive sense voltage
//     in the material-derived read range (≈ 0 .. readRange.Max).
//   - WRITE: arm the material-derived write range (≈ writeRange.Min .. writeRange.Max)
//     but keep the array at 0V until the user explicitly presses "Program Cell".
//   - COMPUTE: apply a per-column input vector (digital 0–255) mapped to voltages
//     0 .. readRange.Max and enable all rows for MVM (I = G×V).
//
// Bounds / clamping notes:
//   - Read/compute voltages should stay below Vc to avoid disturb.
//   - Write pulses are clamped to the write range and practical DAC limits.
//
// NOTE (architecture): In passive mode (0T1R), WLs are effectively ALWAYS on (no gating),
// so WL configuration is skipped and WRITE uses DAC-only column drive (all WL=0V, selected BL=−V_write).
func (ca *CircuitsApp) setOperationMode(mode OpMode) {
	if ca.deviceState == nil {
		return
	}
	prev := ca.deviceState.GetOperationMode()
	ca.deviceState.SetOperationMode(mode)
	if prev != mode {
		logAction("mode_switch %s -> %s", opModeLabel(prev), opModeLabel(mode))
	}

	// In passive mode, all WLs are always on - skip WL configuration
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R

	switch mode {
	case OpModeRead:
		// Single row active (only in 1T1R/2T1R)
		if !isPassive {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.deviceState.SetDACRangeMode(DACRangeRead)
		// Per VOLTAGE_RULES.md: Only selected column gets read voltage
		// Ground all columns first, then apply to selected column only
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(ca.deviceState.GetSelectedCol(), ca.safeReadVoltage())

	case OpModeWrite:
		// Single row active (only in 1T1R/2T1R)
		// DAC voltages stay at 0 (write requires explicit action to avoid accidents)
		if !isPassive {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.deviceState.SetDACRangeMode(DACRangeWrite)
		ca.deviceState.SetAllDACVoltages(0) // Safe: no voltage until explicit write

	case OpModeCompute:
		// All rows active for MVM
		if !isPassive {
			ca.deviceState.SetWLAll()
		}
		ca.deviceState.SetDACRangeMode(DACRangeRead)
		// Apply input vector as DAC voltages (MVM: I = G x V)
		// Input vector values (0-255) map to voltage range
		params := make([]float64, len(ca.inputVector))
		for i, v := range ca.inputVector {
			params[i] = float64(v)
		}
		ca.deviceState.SetDACPreset(DACInputVector, params...)
	}

	ca.updateModeButtons()
	ca.updateActionButtons()  // Enable/disable action buttons based on mode
	ca.updateModePanels(mode) // Show/hide mode-specific panels
	ca.updateDACRangeModeLabel()
	ca.recomputeAndRefresh()
}

// updateModeButtons updates the mode button highlighting
func (ca *CircuitsApp) updateModeButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()

	sharedwidgets.SafeDo(func() {
		// Reset all to low importance
		if ca.modeReadBtn != nil {
			ca.modeReadBtn.Importance = widget.LowImportance
			ca.modeReadBtn.Refresh()
		}
		if ca.modeWriteBtn != nil {
			ca.modeWriteBtn.Importance = widget.LowImportance
			ca.modeWriteBtn.Refresh()
		}
		if ca.modeComputeBtn != nil {
			ca.modeComputeBtn.Importance = widget.LowImportance
			ca.modeComputeBtn.Refresh()
		}

		// Highlight active mode
		switch mode {
		case OpModeRead:
			if ca.modeReadBtn != nil {
				ca.modeReadBtn.Importance = widget.HighImportance
				ca.modeReadBtn.Refresh()
			}
		case OpModeWrite:
			if ca.modeWriteBtn != nil {
				ca.modeWriteBtn.Importance = widget.HighImportance
				ca.modeWriteBtn.Refresh()
			}
		case OpModeCompute:
			if ca.modeComputeBtn != nil {
				ca.modeComputeBtn.Importance = widget.HighImportance
				ca.modeComputeBtn.Refresh()
			}
		}
	})
}

func (ca *CircuitsApp) isProgrammingActive() bool {
	ca.mu.RLock()
	active := ca.programmingActive
	ca.mu.RUnlock()
	return active
}

func (ca *CircuitsApp) setProgrammingActive(active bool) {
	ca.mu.Lock()
	ca.programmingActive = active
	ca.mu.Unlock()

	sharedwidgets.SafeDo(func() {
		if ca.mfuxWriteLevelSlider != nil {
			if active {
				ca.mfuxWriteLevelSlider.Disable()
			} else {
				ca.mfuxWriteLevelSlider.Enable()
			}
		}
		if ca.archPassiveBtn != nil {
			if active {
				ca.archPassiveBtn.Disable()
			} else {
				ca.archPassiveBtn.Enable()
			}
		}
		if ca.arch1T1RBtn != nil {
			if active {
				ca.arch1T1RBtn.Disable()
			} else {
				ca.arch1T1RBtn.Enable()
			}
		}
		if ca.arch2T1RBtn != nil {
			if active {
				ca.arch2T1RBtn.Disable()
			} else {
				ca.arch2T1RBtn.Enable()
			}
		}
		if ca.isppEngineSelect != nil {
			if active {
				ca.isppEngineSelect.Disable()
			} else {
				ca.isppEngineSelect.Enable()
			}
		}
		if ca.actionWriteCellBtn != nil {
			if active {
				ca.actionWriteCellBtn.SetText("Programming...")
			} else {
				ca.actionWriteCellBtn.SetText(actionLabelProgram)
			}
		}
		if active && ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText("PROGRAMMING — controls locked")
		}
	})

	ca.updateActionButtons()
}

// updateActionButtons enables/disables and shows/hides action buttons based on current mode
func (ca *CircuitsApp) updateActionButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()
	programmingActive := ca.isProgrammingActive()

	sharedwidgets.SafeDo(func() {
		// Program Cell: visible only in WRITE mode
		if ca.actionWriteCellBtn != nil {
			if mode == OpModeWrite {
				ca.actionWriteCellBtn.Show()
				if programmingActive {
					ca.actionWriteCellBtn.Disable()
				} else {
					ca.actionWriteCellBtn.Enable()
				}
				ca.actionWriteCellBtn.Importance = widget.HighImportance
			} else {
				ca.actionWriteCellBtn.Hide()
				ca.actionWriteCellBtn.Disable()
				ca.actionWriteCellBtn.Importance = widget.MediumImportance
			}
			ca.actionWriteCellBtn.Refresh()
		}

		// MVM: visible only in COMPUTE mode
		if ca.actionComputeBtn != nil {
			if mode == OpModeCompute {
				ca.actionComputeBtn.Show()
				ca.actionComputeBtn.Enable()
			} else {
				ca.actionComputeBtn.Hide()
				ca.actionComputeBtn.Disable()
			}
		}

		// Random Array: visible in WRITE and COMPUTE modes (not READ - read is non-destructive)
		if ca.actionRandomArrayBtn != nil {
			if mode == OpModeRead {
				ca.actionRandomArrayBtn.Hide()
			} else {
				ca.actionRandomArrayBtn.Show()
			}
		}

		// Reset Array: visible in WRITE and COMPUTE modes
		if ca.actionResetArrayBtn != nil {
			if mode == OpModeRead {
				ca.actionResetArrayBtn.Hide()
			} else {
				ca.actionResetArrayBtn.Show()
			}
		}

		// Undo: always visible (can undo in any mode)
		// Export: always visible (can export in any mode)
		// These don't need mode filtering
	})
}

// ============================================================================
// EVENT HANDLERS
// ============================================================================

// onDACVoltageChanged handles DAC voltage input changes
func (ca *CircuitsApp) onDACVoltageChanged(col int, voltageStr string) {
	voltage, err := strconv.ParseFloat(voltageStr, 64)
	if err != nil {
		return
	}

	// Clamp voltage to valid DAC range
	if voltage < 0 {
		voltage = 0
	}
	if voltage > MaxDACVoltageV {
		voltage = MaxDACVoltageV
	}

	logInput("dac_col=%d voltage=%.3f", col, voltage)
	ca.deviceState.SetDACVoltage(col, voltage)
	ca.recomputeAndRefresh()
}

func (ca *CircuitsApp) setSensePresetSelection(name string) {
	if ca.sensePresetSelect == nil {
		return
	}
	ca.sensePresetUpdating = true
	ca.sensePresetSelect.SetSelected(name)
	ca.sensePresetUpdating = false
}

func (ca *CircuitsApp) setSensePresetCustom() {
	ca.setSensePresetSelection(customSensePresetName)
}

func (ca *CircuitsApp) applySensePreset(name string) {
	for _, preset := range senseMeasurementPresets {
		if preset.Name != name {
			continue
		}
		if ca.tia == nil || ca.adc == nil {
			return
		}
		ca.tia.Gain = preset.Rf * 1e3
		ca.tiaGain = preset.Rf
		ca.adc.VrefLow = preset.Vmin
		ca.adc.VrefHigh = preset.Vmax

		sharedwidgets.SafeDo(func() {
			if ca.senseRfEntry != nil {
				ca.senseRfEntry.SetText(fmt.Sprintf("%.1f", preset.Rf))
			}
			if ca.senseAdcVminEntry != nil {
				ca.senseAdcVminEntry.SetText(fmt.Sprintf("%.2f", preset.Vmin))
			}
			if ca.senseAdcVmaxEntry != nil {
				ca.senseAdcVmaxEntry.SetText(fmt.Sprintf("%.2f", preset.Vmax))
			}
		})
		ca.recomputeAndRefresh()
		return
	}
}

// applySenseRf updates the TIA feedback resistance (Rf) from UI input.
func (ca *CircuitsApp) applySenseRf(valueStr string) {
	if ca.tia == nil {
		return
	}
	rfKOhm, err := strconv.ParseFloat(valueStr, 64)
	if err != nil {
		rfKOhm = ca.tia.Gain / 1e3
	}

	// Physics: Rf sets transimpedance gain (Vout = I * Rf + Vref).
	// Larger Rf increases sensitivity but reduces measurable current range.
	if rfKOhm < minSenseRfKOhm {
		rfKOhm = minSenseRfKOhm
	}
	if rfKOhm > maxSenseRfKOhm {
		rfKOhm = maxSenseRfKOhm
	}

	ca.tia.Gain = rfKOhm * 1e3
	ca.tiaGain = rfKOhm

	if ca.senseRfEntry != nil {
		sharedwidgets.SafeDo(func() {
			ca.senseRfEntry.SetText(fmt.Sprintf("%.1f", rfKOhm))
		})
	}
	ca.setSensePresetCustom()
	ca.recomputeAndRefresh()
}

// applySenseADCRange updates the ADC reference range (Vmin/Vmax) from UI input.
func (ca *CircuitsApp) applySenseADCRange() {
	if ca.adc == nil {
		return
	}

	vmin := ca.adc.VrefLow
	vmax := ca.adc.VrefHigh
	if ca.senseAdcVminEntry != nil {
		if parsed, err := strconv.ParseFloat(ca.senseAdcVminEntry.Text, 64); err == nil {
			vmin = parsed
		}
	}
	if ca.senseAdcVmaxEntry != nil {
		if parsed, err := strconv.ParseFloat(ca.senseAdcVmaxEntry.Text, 64); err == nil {
			vmax = parsed
		}
	}

	// Physics: ADC Vref window defines the measurable voltage span before quantization.
	// Clamp to a realistic window and enforce a non-zero span.
	if vmin < minSenseADCVref {
		vmin = minSenseADCVref
	}
	if vmax > maxSenseADCVref {
		vmax = maxSenseADCVref
	}
	if vmin > maxSenseADCVref-minSenseADCVrefSpan {
		vmin = maxSenseADCVref - minSenseADCVrefSpan
	}
	if vmax <= vmin+minSenseADCVrefSpan {
		vmax = vmin + minSenseADCVrefSpan
	}

	ca.adc.VrefLow = vmin
	ca.adc.VrefHigh = vmax

	if ca.senseAdcVminEntry != nil || ca.senseAdcVmaxEntry != nil {
		sharedwidgets.SafeDo(func() {
			if ca.senseAdcVminEntry != nil {
				ca.senseAdcVminEntry.SetText(fmt.Sprintf("%.2f", vmin))
			}
			if ca.senseAdcVmaxEntry != nil {
				ca.senseAdcVmaxEntry.SetText(fmt.Sprintf("%.2f", vmax))
			}
		})
	}
	ca.setSensePresetCustom()
	ca.recomputeAndRefresh()
}

// setUnifiedDACPreset applies a DAC preset (called by mode changes)
func (ca *CircuitsApp) setUnifiedDACPreset(preset DACMode) {
	switch preset {
	case DACReadPreset:
		ca.deviceState.SetDACPreset(DACReadPreset)
	case DACWritePreset:
		ca.deviceState.SetDACPreset(DACWritePreset)
	}
	ca.updateDACRangeModeLabel()
	ca.recomputeAndRefresh()
}

// setAllUnifiedDACVoltages sets all DAC voltages to the same value
func (ca *CircuitsApp) setAllUnifiedDACVoltages(voltageStr string) {
	voltage, err := strconv.ParseFloat(voltageStr, 64)
	if err != nil {
		return
	}
	logInput("dac_all=%.3f", voltage)
	ca.deviceState.SetAllDACVoltages(voltage)
	ca.recomputeAndRefresh()
}

// setWLModeSingle sets WL mode to single (only selected row)
func (ca *CircuitsApp) setWLModeSingle() {
	selectedRow := ca.deviceState.GetSelectedRow()
	ca.deviceState.SetWLSingle(selectedRow)
	ca.recomputeAndRefresh()
}

// setWLModeAll sets WL mode to all rows active
func (ca *CircuitsApp) setWLModeAll() {
	ca.deviceState.SetWLAll()
	ca.recomputeAndRefresh()
}

// onUnifiedCellTapped handles cell selection
// In READ/WRITE mode: selects row transistor (WL) and column transistor (BL)
func (ca *CircuitsApp) onUnifiedCellTapped(row, col int) {
	if ca.isProgrammingActive() {
		return
	}
	logInput("cell_selected row=%d col=%d", row, col)
	ca.deviceState.SetSelectedCell(row, col)

	mode := ca.deviceState.GetOperationMode()
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R

	// In READ/WRITE mode (non-passive): select ONLY this row (single transistor)
	if !isPassive && (mode == OpModeRead || mode == OpModeWrite) {
		ca.deviceState.SetWLSingle(row)
	}

	// Update target cell label in write mode panel
	ca.updateWriteTargetLabel()

	ca.recomputeAndRefresh()
	ca.updateCellInfo()

	// Auto-sense in READ mode when clicking a cell
	if mode == OpModeRead {
		ca.onUnifiedRead()
	}
}

// ============================================================================
// UI UPDATE HELPERS
// ============================================================================

const uiRefreshMinInterval = 50 * time.Millisecond

// recomputeAndRefresh runs computation and updates all UI elements (throttled).
func (ca *CircuitsApp) recomputeAndRefresh() {
	ca.scheduleRecomputeAndRefresh(false)
}

func (ca *CircuitsApp) beginUIRefreshSuspension() {
	ca.uiUpdateMu.Lock()
	defer ca.uiUpdateMu.Unlock()

	ca.suspendUIRefresh = true
	ca.lastUIUpdate = time.Time{}
	ca.pendingUIUpd = false
	if ca.uiUpdateTimer != nil {
		ca.uiUpdateTimer.Stop()
		ca.uiUpdateTimer = nil
	}
}

func (ca *CircuitsApp) endUIRefreshSuspension() {
	ca.uiUpdateMu.Lock()
	defer ca.uiUpdateMu.Unlock()

	ca.suspendUIRefresh = false
	ca.lastUIUpdate = time.Time{}
}

func (ca *CircuitsApp) cancelPendingUIRefresh() {
	ca.uiUpdateMu.Lock()
	defer ca.uiUpdateMu.Unlock()

	ca.pendingUIUpd = false
	ca.lastUIUpdate = time.Time{}
	if ca.uiUpdateTimer != nil {
		ca.uiUpdateTimer.Stop()
		ca.uiUpdateTimer = nil
	}
}

// recomputeAndRefreshNow forces an immediate recompute + UI refresh.
func (ca *CircuitsApp) recomputeAndRefreshNow() {
	// Hold the read lock for the duration of Compute() so concurrent writers
	// (e.g. ISPP animation goroutines) cannot mutate the backing slices while
	// the device simulator iterates.
	ca.mu.RLock()
	weights := ca.arrayWeights
	levels := ca.quantLevels
	ca.deviceState.Compute(weights, levels)
	ca.mu.RUnlock()

	// Update cell info
	ca.updateCellInfo()
	ca.updateSensePanel()

	// Update operation classification
	ca.updateOperationClassification()

	// Refresh array canvas
	ca.refreshUnifiedArray()
}

func (ca *CircuitsApp) scheduleRecomputeAndRefresh(force bool) {
	ca.uiUpdateMu.Lock()
	if ca.suspendUIRefresh {
		ca.uiUpdateMu.Unlock()
		return
	}
	if force {
		ca.lastUIUpdate = time.Now()
		ca.pendingUIUpd = false
		if ca.uiUpdateTimer != nil {
			ca.uiUpdateTimer.Stop()
			ca.uiUpdateTimer = nil
		}
		ca.uiUpdateMu.Unlock()
		ca.recomputeAndRefreshNow()
		return
	}

	now := time.Now()
	if ca.lastUIUpdate.IsZero() || now.Sub(ca.lastUIUpdate) >= uiRefreshMinInterval {
		ca.lastUIUpdate = now
		ca.uiUpdateMu.Unlock()
		ca.recomputeAndRefreshNow()
		return
	}
	if ca.pendingUIUpd {
		ca.uiUpdateMu.Unlock()
		return
	}
	delay := uiRefreshMinInterval - now.Sub(ca.lastUIUpdate)
	ca.pendingUIUpd = true
	ca.uiUpdateTimer = time.AfterFunc(delay, func() {
		ca.uiUpdateMu.Lock()
		if ca.suspendUIRefresh {
			ca.pendingUIUpd = false
			ca.uiUpdateTimer = nil
			ca.uiUpdateMu.Unlock()
			return
		}
		ca.lastUIUpdate = time.Now()
		ca.pendingUIUpd = false
		ca.uiUpdateTimer = nil
		ca.uiUpdateMu.Unlock()
		ca.recomputeAndRefreshNow()
	})
	ca.uiUpdateMu.Unlock()
}

// refreshUnifiedArray refreshes the array canvas
func (ca *CircuitsApp) refreshUnifiedArray() {
	if ca.sharedArrayCanvas != nil {
		sharedwidgets.SafeDo(func() {
			ca.sharedArrayCanvas.Refresh()
		})
	}
}

// updateCellInfo updates the cell info display with detailed circuit data
func (ca *CircuitsApp) updateCellInfo() {
	if ca.sharedCellInfoLabel == nil {
		return
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()

	ca.mu.RLock()
	var level int
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		level = ca.arrayWeights[selectedRow][selectedCol]
	}
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Use material's physics-based conductance calculation
	var conductanceUS float64
	material := ca.deviceState.GetMaterial()
	if material != nil {
		conductanceUS = material.DiscreteLevel(level, levels) * 1e6 // S to uS
	} else {
		conductanceUS = 1.0 + float64(level)/float64(levels-1)*99.0
	}

	blVoltage := ca.deviceState.GetDACVoltage(selectedCol)
	wlVoltage := ca.deviceState.GetWLVoltage(selectedRow)
	effectiveVoltage := ca.deviceState.GetEffectiveCellVoltage(selectedRow, selectedCol)
	matName := ca.deviceState.GetMaterialName()

	// Calculate expected current I = G x V
	expectedCurrent := conductanceUS * effectiveVoltage // uA (signed)

	// Get actual row output (includes all cells in row if active)
	rowCurrent := ca.deviceState.GetRowCurrent(selectedRow)
	rowVoltage := ca.deviceState.GetRowVoltage(selectedRow)
	adcLevel := ca.deviceState.GetRowLevel(selectedRow)
	isActive := ca.deviceState.IsRowActive(selectedRow)
	isPassive := ca.deviceState.IsPassiveMode()

	sharedwidgets.SafeDo(func() {
		rowState := "OFF"
		if isActive {
			rowState = "ON"
		}
		voltageField := fmt.Sprintf("V_cell (V): %+.2f V", effectiveVoltage)
		if isPassive {
			voltageField = fmt.Sprintf("WL (V): %+.2f V | BL (V): %+.2f V | V_cell (V): %+.2f V", wlVoltage, blVoltage, effectiveVoltage)
		}

		primaryField := voltageField
		if ca.showCurrentInCellInfo {
			primaryField = fmt.Sprintf("I_cell (µA): %s", formatMetricICellUA(expectedCurrent))
		}

		infoStr := fmt.Sprintf(
			"Cell [%d,%d] | Level: %s/%d | G (µS): %s | %s | V_TIA (V): %s | ADC Code: %s | Row: %s | Material: %s",
			selectedRow,
			selectedCol,
			formatMetricLevel(level),
			levels-1,
			formatMetricConductanceUS(conductanceUS),
			primaryField,
			formatMetricVTIAMV(rowVoltage),
			formatMetricADCCode(adcLevel),
			rowState,
			matName,
		)
		ca.sharedCellInfoLabel.SetText(infoStr)
	})

	// Also update array info label with total row current
	if ca.sharedArrayInfoLabel != nil {
		sharedwidgets.SafeDo(func() {
			ca.sharedArrayInfoLabel.SetText(fmt.Sprintf("Array: %dx%d | Quantization Levels: %d | Row %d I_row (µA): %+.2f µA",
				ca.arrayRows, ca.arrayCols, ca.quantLevels, selectedRow, rowCurrent))
		})
	}
}

func (ca *CircuitsApp) senseChainConfig() (arraysim.SenseChain, bool) {
	if ca.tia == nil || ca.adc == nil {
		return arraysim.SenseChain{}, false
	}
	return arraysim.SenseChain{
		TIA: arraysim.TIAConfig{
			Rf:   ca.tia.Gain,
			Vref: ca.tia.OutputOffset,
			Vmin: 0,
			Vmax: ca.tia.MaxOutputVoltage,
		},
		ADC: arraysim.ADCConfig{
			Bits: ca.adc.Bits,
			Vmin: ca.adc.VrefLow,
			Vmax: ca.adc.VrefHigh,
		},
	}, true
}

func (ca *CircuitsApp) composedSenseSNRdB(currentA float64, sense arraysim.SenseChain) float64 {
	if ca.tia == nil {
		return math.NaN()
	}
	signalRMS := math.Abs(currentA)
	bandwidth := ca.tia.Bandwidth
	if bandwidth <= 0 {
		bandwidth = 1
	}

	// Input-referred noise composition (A_rms): TIA input noise + shot noise + ADC quantization.
	tiaNoiseRMS := math.Abs(ca.tia.InputNoiseRMS) * math.Sqrt(bandwidth)
	shotNoiseRMS := peripherals.ShotNoiseCurrentRMS(math.Abs(currentA), bandwidth)
	quantNoiseRMS := 0.0
	if lsb := sense.CurrentLSB(); lsb > 0 {
		quantNoiseRMS = lsb / math.Sqrt(12)
	}
	totalNoiseRMS := math.Sqrt(peripherals.TotalNoiseVariance(tiaNoiseRMS, shotNoiseRMS, quantNoiseRMS))
	return peripherals.SNRDB(signalRMS, totalNoiseRMS)
}

// updateSensePanel updates the compact sense-chain readout.
func (ca *CircuitsApp) updateSensePanel() {
	if ca.sensePanel == nil || ca.deviceState == nil {
		return
	}

	row := ca.deviceState.GetSelectedRow()
	currentA := ca.deviceState.GetRowCurrent(row) * 1e-6
	voltage := ca.deviceState.GetRowVoltage(row)
	code := ca.deviceState.GetRowLevel(row)
	saturated := ca.deviceState.IsSaturated(row)
	isActive := ca.deviceState.IsRowActive(row)
	mode := ca.deviceState.GetOperationMode()

	rangeText := "n/a"
	lsbText := "n/a"
	snrText := "n/a"
	enobText := "n/a"
	effSNRText := "n/a"
	if sense, ok := ca.senseChainConfig(); ok {
		imin, imax := sense.CurrentRange()
		lsb := sense.CurrentLSB()
		if lsb > 0 {
			rangeText = fmt.Sprintf("%s .. %s", formatCurrentA(imin), formatCurrentA(imax))
			lsbText = formatCurrentA(lsb)
		}
		snrDB := ca.composedSenseSNRdB(currentA, sense)
		if math.IsInf(snrDB, 1) {
			snrText = "+Inf dB"
		} else if math.IsNaN(snrDB) {
			snrText = "n/a"
		} else {
			snrText = fmt.Sprintf("%.1f dB", snrDB)
		}
		if ca.adc != nil {
			enob := ca.adc.ENOB()
			enobText = fmt.Sprintf("%.2f bits", enob)
			effSNR := ca.adc.EffectiveSNR()
			effSNRText = fmt.Sprintf("%.1f dB", effSNR)
		}
	}

	codeText := formatMetricADCCode(code)

	rowText := fmt.Sprintf("Row %d", row)
	if !isActive {
		rowText = fmt.Sprintf("Row %d (inactive)", row)
	}

	titleText := "Sense (TIA+ADC)"
	if mode == OpModeRead {
		titleText = "Sense (READ)"
	} else if mode == OpModeCompute {
		titleText = "Sense (COMPUTE)"
	}

	satText := "Linear"
	if saturated {
		satText = "SATURATED (clipped)"
	}

	sharedwidgets.SafeDo(func() {
		if ca.senseTitleLabel != nil {
			ca.senseTitleLabel.SetText(titleText)
		}
		if ca.senseRowLabel != nil {
			ca.senseRowLabel.SetText(rowText)
		}
		if ca.senseCurrentLabel != nil {
			ca.senseCurrentLabel.SetText(formatMetricICellUA(currentA * 1e6))
		}
		if ca.senseVoltageLabel != nil {
			ca.senseVoltageLabel.SetText(fmt.Sprintf("V_TIA out: %s (I × R_fb)", formatMetricVTIAMV(voltage)))
		}
		if ca.senseCodeLabel != nil {
			ca.senseCodeLabel.SetText(fmt.Sprintf("Code: %s", codeText))
		}
		if ca.senseSNRLabel != nil {
			ca.senseSNRLabel.SetText(snrText)
		}
		if ca.senseSaturationLabel != nil {
			ca.senseSaturationLabel.SetText(satText)
		}
		if ca.senseRangeLabel != nil {
			ca.senseRangeLabel.SetText(rangeText)
		}
		if ca.senseLSBLabel != nil {
			ca.senseLSBLabel.SetText(lsbText)
		}
		if ca.senseENOBLabel != nil {
			ca.senseENOBLabel.SetText(enobText)
		}
		if ca.senseEffSNRLabel != nil {
			ca.senseEffSNRLabel.SetText(effSNRText)
		}
	})
}

// updateOperationClassification updates the operation classification display
func (ca *CircuitsApp) updateOperationClassification() {
	if ca.operationsModeHelp == nil || ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()
	arch := ca.architecture
	readRange := ca.deviceState.GetReadRange()
	writeRange := ca.deviceState.GetWriteRange()

	var helpText string
	switch mode {
	case OpModeRead:
		if arch == sharedwidgets.Architecture2T1R {
			helpText = fmt.Sprintf("READ: Single row, 0-%.1fV. 2T1R provides perfect isolation.", readRange.Max)
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = fmt.Sprintf("READ: Single row, 0-%.1fV. 1T1R transistor isolates selected row.", readRange.Max)
		} else {
			helpText = fmt.Sprintf("READ: 0-%.1fV. Passive array - sneak currents add 5-20%% error.", readRange.Max)
		}
	case OpModeWrite:
		if arch == sharedwidgets.Architecture2T1R {
			helpText = fmt.Sprintf("WRITE: Single row, %.1f-%.1fV. 2T1R selects single cell. Use Program Cell to apply pulse.", writeRange.Min, writeRange.Max)
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = fmt.Sprintf("WRITE: Single row, %.1f-%.1fV. 1T1R gates selected row. Use Program Cell to apply pulse.", writeRange.Min, writeRange.Max)
		} else {
			helpText = fmt.Sprintf("WRITE: %.1f-%.1fV. Passive: DAC drives selected BL; entire column switches. Use Program Cell to apply pulse.", writeRange.Min, writeRange.Max)
		}
	case OpModeCompute:
		if arch == sharedwidgets.Architecture0T1R {
			helpText = fmt.Sprintf("COMPUTE: All rows, 0-%.1fV. Passive natural MVM mode (~76ns).", readRange.Max)
		} else {
			helpText = fmt.Sprintf("COMPUTE: All transistors ON, 0-%.1fV. Full MVM in ~76ns.", readRange.Max)
		}
	default:
		helpText = "Select a mode: READ, WRITE, or COMPUTE."
	}

	sharedwidgets.SafeDo(func() {
		ca.operationsModeHelp.SetText(helpText)
	})
}

// H2 FIX: updateWriteTargetLabel updates the target cell display in write mode panel
func (ca *CircuitsApp) updateWriteTargetLabel() {
	if ca.mfuxWriteTargetLabel == nil || ca.deviceState == nil {
		return
	}

	row := ca.deviceState.GetSelectedRow()
	col := ca.deviceState.GetSelectedCol()

	sharedwidgets.SafeDo(func() {
		ca.mfuxWriteTargetLabel.SetText(fmt.Sprintf("[%d,%d]", row, col))
	})
}

// ============================================================================
// ARCHITECTURE TOGGLE
// ============================================================================

// createArchitectureToggle creates the PASSIVE/1T1R/2T1R toggle buttons
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
	handleArchChange := func(arch string) {
		ca.mu.Lock()
		ca.architecture = arch
		ca.mu.Unlock()
		logInput("architecture=%s", arch)

		if arch == sharedwidgets.Architecture0T1R {
			// Passive mode: all WLs always active, cannot be changed
			ca.deviceState.SetPassiveMode(true)
		} else {
			// 1T1R/2T1R: disable passive mode, set WLs based on current operation mode
			ca.deviceState.SetPassiveMode(false)
			// Preserve WL state based on operation mode
			if ca.deviceState.GetOperationMode() == OpModeCompute {
				ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
			} else {
				ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
			}
		}

		ca.updateFootprintReference()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	toggle := sharedwidgets.NewArchitectureToggle(sharedwidgets.ArchitectureToggleOptions{
		Initial:      ca.architecture,
		Style:        sharedwidgets.ArchitectureToggleStylePlain,
		LabelPassive: "0T1R",
		Label1T1R:    "1T1R",
		Label2T1R:    "2T1R",
		OnChanged:    handleArchChange,
	})

	ca.archPassiveBtn = toggle.PassiveButton
	ca.arch1T1RBtn = toggle.OneT1RButton
	ca.arch2T1RBtn = toggle.TwoT1RButton
	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	archLabel := widget.NewLabel("Arch:")
	archInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Architecture",
			"0T1R (Passive): No transistor; simpler but suffers sneak paths.\n\n"+
				"1T1R: One transistor per cell; row selection via word-line gating.\n\n"+
				"2T1R: Two transistors; enables bidirectional programming.", ca.window)
	})
	archInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(archInfo, "Crossbar architecture info")
	return container.NewHBox(archLabel, ca.archToggle, archInfo)
}

// ============================================================================
// MODE-FIRST UX PANELS (Phase 1)
// ============================================================================

// createWriteModePanel creates the write mode panel with a *state/level* slider.
//
// Physics meaning:
//   - The slider selects the desired discrete FeFET state index L ∈ [0, quantLevels-1]
//     (i.e., a conductance/polarization level), not a voltage.
//   - The UI shows a voltage *preview* computed from the material-derived write range.
//   - The actual programming pulse(s) are applied only when the user presses "Program Cell"
//     (ISPP-style write/verify loop).
//
// Bounds / clamping:
//   - Slider is integer-stepped and clamped by its min/max.
//   - Voltage preview clamps to the write range.
//
// This addresses UX-004: No target level selector for WRITE mode
func (ca *CircuitsApp) createWriteModePanel() fyne.CanvasObject {
	// Slider: 0 to (quantLevels-1) - uses configured level count
	maxLevel := ca.quantLevels - 1
	midLevel := ca.quantLevels / 2
	ca.mfuxWriteLevelSlider = widget.NewSlider(0, float64(maxLevel))
	ca.mfuxWriteLevelSlider.Step = 1
	ca.mfuxWriteLevelSlider.Value = float64(midLevel)
	ca.mfuxWriteLevelSlider.OnChanged = func(v float64) {
		ca.onWriteLevelChanged(int(v))
	}

	ca.mfuxWriteLevelLabel = widget.NewLabel(fmt.Sprintf("Level: %d", midLevel))
	ca.mfuxWriteLevelLabel.TextStyle = fyne.TextStyle{Monospace: true}

	ca.mfuxWriteVoltageLabel = widget.NewLabel("Voltage: 1.00V")
	ca.mfuxWriteVoltageLabel.TextStyle = fyne.TextStyle{Monospace: true}

	// H2 FIX: Add target cell display
	ca.mfuxWriteTargetLabel = widget.NewLabel("Target: Row 0, Col 0")
	ca.mfuxWriteTargetLabel.TextStyle = fyne.TextStyle{Bold: true}

	// Layout: Title row with target cell, then slider with value labels
	titleLabel := widget.NewLabelWithStyle("Target Write Level:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	headerRow := container.NewHBox(
		titleLabel,
		layout.NewSpacer(),
		ca.mfuxWriteTargetLabel,
	)

	sliderRow := container.NewBorder(nil, nil,
		ca.mfuxWriteLevelLabel,
		ca.mfuxWriteVoltageLabel,
		ca.mfuxWriteLevelSlider,
	)

	return container.NewVBox(
		headerRow,
		sliderRow,
	)
}

// onWriteLevelChanged handles target state changes from the WRITE slider.
//
// Physics meaning:
//   - Updates the displayed "nominal write pulse amplitude" associated with the requested
//     final state index, based on the configured write voltage window.
//
// Safety / bounds:
//   - This is a *preview only* and does NOT apply voltage to the DAC.
//   - Voltage is only applied when the user presses the "Program Cell" button.
func (ca *CircuitsApp) onWriteLevelChanged(level int) {
	if ca.deviceState == nil {
		return
	}

	// Calculate voltage for display only (don't apply to DAC)
	targetVoltage := ca.deviceState.CalculateVoltageForState(level, ca.quantLevels)
	appliedVoltage := targetVoltage
	logInput("write_level=%d voltage=%.3f", level, targetVoltage)

	// Determine direction based on current cell level vs target
	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()
	ca.mu.RLock()
	currentLevel := 0
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		currentLevel = ca.arrayWeights[selectedRow][selectedCol]
	}
	ca.mu.RUnlock()

	dirLabel := "SET"
	if level < currentLevel {
		dirLabel = "RESET"
	} else if level == currentLevel {
		dirLabel = "—"
	}

	// Show bipolar write range from material
	writeRange := ca.deviceState.GetWriteRange()

	sharedwidgets.SafeDo(func() {
		if ca.mfuxWriteLevelLabel != nil {
			ca.mfuxWriteLevelLabel.SetText(fmt.Sprintf("L:%d", level))
		}
		if ca.mfuxWriteVoltageLabel != nil {
			voltText := fmt.Sprintf("%s %.2fV [±%.1fV]", dirLabel, appliedVoltage, writeRange.Max)
			ca.mfuxWriteVoltageLabel.SetText(voltText)
		}
	})
}

// createComputeModePanel creates the compute mode panel with input vector entries
// Supports variable array sizes: shows individual entries for small arrays, grid for large
func (ca *CircuitsApp) createComputeModePanel() fyne.CanvasObject {
	cols := ca.arrayCols

	// Title with array size + units info.
	// Physics meaning: each entry is a digital code (0–255) mapped to a column voltage
	// Vj = (code/255) * readRange.Max (compute-safe range).
	readMax := 1.0
	if ca.deviceState != nil {
		readMax = ca.deviceState.GetReadRange().Max
	}
	titleLabel := widget.NewLabelWithStyle(
		fmt.Sprintf("Input Vector (%d inputs, 0–255 → 0–%.2fV):", cols, readMax),
		fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ca.computeInputTitle = titleLabel

	// Random button to populate with random values
	ca.computeRandomBtn = widget.NewButton("Random Inputs", func() {
		ca.randomizeInputVectorEntries()
	})

	// Clear button
	ca.computeClearBtn = widget.NewButton("Clear Inputs", func() {
		ca.clearInputVectorEntries()
	})

	// Title row with buttons
	titleRow := container.NewHBox(titleLabel, layout.NewSpacer(), ca.computeRandomBtn, ca.computeClearBtn)

	// Horizontal container for input entries
	ca.computeInputContainer = container.NewHBox()
	ca.buildComputeInputEntries()

	// Horizontal scroll for large arrays
	scrollContent := container.NewHScroll(ca.computeInputContainer)
	scrollContent.SetMinSize(fyne.NewSize(400, 40)) // Width for scrolling, shorter height

	return container.NewVBox(
		titleRow,
		scrollContent,
	)
}

// createSensePanel creates the compact sense-chain panel for READ/COMPUTE modes.
func (ca *CircuitsApp) createSensePanel() fyne.CanvasObject {
	titleLabel := widget.NewLabelWithStyle("Sense (TIA+ADC)", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})
	ca.senseTitleLabel = titleLabel
	ca.senseRowLabel = widget.NewLabel("Row 0")

	headerRow := container.NewHBox(titleLabel, layout.NewSpacer(), ca.senseRowLabel)

	ca.senseCurrentLabel = widget.NewLabel("+0.00 µA")
	ca.senseVoltageLabel = widget.NewLabel("V_TIA out: 0.0 mV (I × R_fb)")
	ca.senseCodeLabel = widget.NewLabel("Code: 0")
	ca.senseCodeLabel.TextStyle = fyne.TextStyle{Monospace: true}
	ca.senseSNRLabel = widget.NewLabel("SNR n/a")
	ca.senseSaturationLabel = widget.NewLabel("Linear")
	ca.senseSaturationLabel.TextStyle = fyne.TextStyle{Bold: true}

	metricsRow := container.NewGridWithColumns(2,
		container.NewHBox(widget.NewLabel("I_cell (µA):"), ca.senseCurrentLabel),
		container.NewHBox(widget.NewLabel("V_TIA (V):"), ca.senseVoltageLabel),
		container.NewHBox(widget.NewLabel("ADC Code:"), ca.senseCodeLabel),
		container.NewHBox(widget.NewLabel("SNR:"), ca.senseSNRLabel),
		container.NewHBox(widget.NewLabel("Status:"), ca.senseSaturationLabel),
	)

	ca.senseRangeLabel = widget.NewLabel("n/a")
	ca.senseLSBLabel = widget.NewLabel("n/a")

	rangeInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("Measurable Range",
			"Measurable current range after TIA rails and ADC references.\nI = (V - Vref) / Rf, using Vmin/Vmax = max/min(TIA, ADC).", ca.window)
	})
	rangeInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(rangeInfo, "Measurable current range info")
	lsbInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("LSB Current",
			"LSB current = (Vmax_eff - Vmin_eff) / (2^bits - 1) / Rf.", ca.window)
	})
	lsbInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(lsbInfo, "Current LSB info")

	rangeRow := container.NewGridWithColumns(2,
		container.NewHBox(widget.NewLabel("Measurable I-range:"), ca.senseRangeLabel, rangeInfo),
		container.NewHBox(widget.NewLabel("Current LSB:"), ca.senseLSBLabel, lsbInfo),
	)

	ca.senseENOBLabel = widget.NewLabel("n/a")
	ca.senseEffSNRLabel = widget.NewLabel("n/a")

	enobInfo := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
		dialog.ShowInformation("ENOB",
			"Effective Number of Bits accounts for INL/DNL non-idealities.\nENOB = Bits - log\u2082(\u221a(1 + INL\u00b2 + DNL\u00b2)).\nIdeal ENOB equals ADC bits.", ca.window)
	})
	enobInfo.Importance = widget.LowImportance
	sharedwidgets.SetAccessibleLabel(enobInfo, "ENOB info")

	fidelityRow := container.NewGridWithColumns(2,
		container.NewHBox(widget.NewLabel("ENOB:"), ca.senseENOBLabel, enobInfo),
		container.NewHBox(widget.NewLabel("Effective SNR:"), ca.senseEffSNRLabel),
	)

	ca.senseRfEntry = widget.NewEntry()
	ca.senseRfEntry.SetPlaceHolder("10.0")
	if ca.tia != nil {
		ca.senseRfEntry.SetText(fmt.Sprintf("%.1f", ca.tia.Gain/1e3))
	} else {
		ca.senseRfEntry.SetText(fmt.Sprintf("%.1f", minSenseRfKOhm))
	}
	ca.senseRfEntry.OnSubmitted = func(s string) {
		ca.applySenseRf(s)
	}

	ca.senseAdcVminEntry = widget.NewEntry()
	ca.senseAdcVmaxEntry = widget.NewEntry()
	ca.senseAdcVminEntry.SetPlaceHolder("0.0")
	ca.senseAdcVmaxEntry.SetPlaceHolder("1.0")
	if ca.adc != nil {
		ca.senseAdcVminEntry.SetText(fmt.Sprintf("%.2f", ca.adc.VrefLow))
		ca.senseAdcVmaxEntry.SetText(fmt.Sprintf("%.2f", ca.adc.VrefHigh))
	} else {
		ca.senseAdcVminEntry.SetText(fmt.Sprintf("%.2f", minSenseADCVref))
		ca.senseAdcVmaxEntry.SetText(fmt.Sprintf("%.2f", maxSenseADCVref))
	}
	ca.senseAdcVminEntry.OnSubmitted = func(string) {
		ca.applySenseADCRange()
	}
	ca.senseAdcVmaxEntry.OnSubmitted = func(string) {
		ca.applySenseADCRange()
	}

	presetOptions := make([]string, 0, len(senseMeasurementPresets)+1)
	for _, p := range senseMeasurementPresets {
		presetOptions = append(presetOptions, p.Name)
	}
	presetOptions = append(presetOptions, customSensePresetName)
	ca.sensePresetSelect = widget.NewSelect(presetOptions, func(selected string) {
		if ca.sensePresetUpdating || selected == "" || selected == customSensePresetName {
			return
		}
		ca.applySensePreset(selected)
	})
	ca.setSensePresetSelection(senseMeasurementPresets[0].Name)

	controlWithHelp := func(label, title, msg string, input fyne.CanvasObject) fyne.CanvasObject {
		info := widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
			dialog.ShowInformation(title, msg, ca.window)
		})
		info.Importance = widget.LowImportance
		sharedwidgets.SetAccessibleLabel(info, title+" info")
		labelRow := container.NewHBox(widget.NewLabel(label), info)
		row := container.NewBorder(nil, nil, labelRow, nil, input)
		return container.NewGridWrap(fyne.NewSize(260, 36), row)
	}

	controlsContent := container.NewHBox(
		controlWithHelp("Preset:", "Sense Preset",
			"Preset configures the read chain end-to-end (DAC Vread → array current → TIA gain → ADC range). Choose by expected current magnitude.", ca.sensePresetSelect),
		controlWithHelp("Rf (kΩ):", "TIA Feedback Resistor (Rf)",
			"TIA stage converts row current to voltage: Vout = Irow × Rf. Larger Rf increases sensitivity but can saturate sooner.", ca.senseRfEntry),
		controlWithHelp("ADC Vmin (V):", "ADC Lower Reference (Vmin)",
			"ADC maps TIA output to codes between Vmin and Vmax. Set Vmin near the minimum expected TIA output in READ mode.", ca.senseAdcVminEntry),
		controlWithHelp("ADC Vmax (V):", "ADC Upper Reference (Vmax)",
			"ADC full-scale top for the read chain. Keep Vmax above expected TIA output peaks to avoid clipping and preserve linear code mapping.", ca.senseAdcVmaxEntry),
	)
	controlsScroll := container.NewHScroll(controlsContent)
	controlsScroll.SetMinSize(fyne.NewSize(980, 44))

	return container.NewVBox(
		headerRow,
		metricsRow,
		rangeRow,
		fidelityRow,
		controlsScroll,
	)
}

// buildComputeInputEntries builds the input entries based on current array size
func (ca *CircuitsApp) buildComputeInputEntries() {
	cols := ca.arrayCols

	// Clear existing entries
	if ca.computeInputContainer != nil {
		ca.computeInputContainer.RemoveAll()
	}

	// Allocate entry arrays
	ca.mfuxInputVectorEntry = make([]*widget.Entry, cols)
	ca.mfuxInputVectorLabels = make([]*widget.Label, cols)

	// Single horizontal row with inline label:entry pairs
	for i := 0; i < cols; i++ {
		idx := i

		// Inline label
		label := widget.NewLabel(fmt.Sprintf("x%d:", i))
		label.TextStyle = fyne.TextStyle{Monospace: true}
		ca.mfuxInputVectorLabels[i] = label

		entry := widget.NewEntry()
		entry.SetPlaceHolder("0")
		entry.SetText("0")
		entry.OnChanged = func(s string) {
			ca.onInputVectorEntryChanged(idx, s)
		}
		ca.mfuxInputVectorEntry[i] = entry

		// Add pair: label then entry
		ca.computeInputContainer.Add(label)
		ca.computeInputContainer.Add(entry)
	}
}

// rebuildComputeInputs rebuilds the compute input panel after array resize
func (ca *CircuitsApp) rebuildComputeInputs() {
	if ca.computeInputContainer == nil {
		return
	}

	// Update title (keep units explicit).
	if ca.computeInputTitle != nil {
		readMax := 1.0
		if ca.deviceState != nil {
			readMax = ca.deviceState.GetReadRange().Max
		}
		sharedwidgets.SafeDo(func() {
			ca.computeInputTitle.SetText(fmt.Sprintf("Input Vector (%d inputs, 0–255 → 0–%.2fV):", ca.arrayCols, readMax))
		})
	}

	// Rebuild entries
	sharedwidgets.SafeDo(func() {
		ca.buildComputeInputEntries()
		ca.computeInputContainer.Refresh()
	})
}

// onInputVectorEntryChanged handles input vector code changes.
//
// Physics meaning:
//   - Each entry is a digital code d (0..MaxDACCode) mapped to an analog column voltage
//     V = (d/MaxDACCode) * readRange.Max for compute-safe MVM.
//
// Bounds / clamping:
//   - Codes are clamped to [0, MaxDACCode].
//   - DAC updates are only applied in COMPUTE mode to avoid unintentionally biasing
//     the array in READ/WRITE modes.
func (ca *CircuitsApp) onInputVectorEntryChanged(col int, valueStr string) {
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return
	}

	// Clamp to valid range
	if value < 0 {
		value = 0
	}
	if value > MaxDACCode {
		value = MaxDACCode
	}
	logInput("input_vector x%d=%d", col, value)

	// Always store the value (for when user switches to COMPUTE mode)
	ca.mu.Lock()
	if col < len(ca.inputVector) {
		ca.inputVector[col] = value
	}
	ca.mu.Unlock()

	// Only apply to DAC if in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		return
	}

	// Convert all inputs to DAC voltages
	params := make([]float64, len(ca.inputVector))
	ca.mu.RLock()
	for i, v := range ca.inputVector {
		params[i] = float64(v)
	}
	ca.mu.RUnlock()

	ca.deviceState.SetDACPreset(DACInputVector, params...)
	ca.recomputeAndRefresh()
}

// randomizeInputVectorEntries fills entries with random 0-255 values
// Only applies to DAC if in COMPUTE mode
func (ca *CircuitsApp) randomizeInputVectorEntries() {
	// Generate random values and copy for UI update
	ca.mu.Lock()
	valuesCopy := make([]int, len(ca.inputVector))
	for i := range ca.inputVector {
		ca.inputVector[i] = rand.Intn(DACCodeRange)
		valuesCopy[i] = ca.inputVector[i]
	}
	ca.mu.Unlock()
	logAction("input_vector_randomized len=%d", len(valuesCopy))

	// Update entry widgets (no lock - use copy)
	sharedwidgets.SafeDo(func() {
		for i, entry := range ca.mfuxInputVectorEntry {
			if entry != nil && i < len(valuesCopy) {
				entry.SetText(strconv.Itoa(valuesCopy[i]))
			}
		}
	})

	// Only apply to DAC if in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		return
	}

	// Apply to DAC
	params := make([]float64, len(valuesCopy))
	for i, v := range valuesCopy {
		params[i] = float64(v)
	}

	ca.deviceState.SetDACPreset(DACInputVector, params...)
	ca.recomputeAndRefresh()
}

// clearInputVectorEntries sets all entries to 0
// Only applies to DAC if in COMPUTE mode
func (ca *CircuitsApp) clearInputVectorEntries() {
	ca.mu.Lock()
	for i := range ca.inputVector {
		ca.inputVector[i] = 0
	}
	// IMPORTANT: Unlock BEFORE fyne.Do to prevent deadlock.
	// SetText triggers OnChanged which acquires ca.mu.
	ca.mu.Unlock()
	logAction("input_vector_cleared len=%d", len(ca.inputVector))

	// Update entry widgets (no lock held - safe)
	sharedwidgets.SafeDo(func() {
		for _, entry := range ca.mfuxInputVectorEntry {
			if entry != nil {
				entry.SetText("0")
			}
		}
	})

	// Only apply to DAC if in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		return
	}

	// Apply to DAC
	params := make([]float64, len(ca.inputVector))
	ca.deviceState.SetDACPreset(DACInputVector, params...)
	ca.recomputeAndRefresh()
}

// updateModePanels shows/hides mode-specific panels based on current mode
func (ca *CircuitsApp) updateModePanels(mode OpMode) {
	sharedwidgets.SafeDo(func() {
		// Hide all panels first
		if ca.writeModePanel != nil {
			ca.writeModePanel.Hide()
		}
		if ca.computeModePanel != nil {
			ca.computeModePanel.Hide()
		}
		if ca.sensePanel != nil {
			ca.sensePanel.Hide()
		}

		// Show relevant panel
		switch mode {
		case OpModeWrite:
			if ca.writeModePanel != nil {
				ca.writeModePanel.Show()
				// Update slider to reflect current selection
				if ca.mfuxWriteLevelSlider != nil {
					// Trigger an update to sync voltage display
					ca.onWriteLevelChanged(int(ca.mfuxWriteLevelSlider.Value))
				}
				// H2 FIX: Update target cell label when entering write mode
				ca.updateWriteTargetLabel()
			}
		case OpModeCompute:
			if ca.computeModePanel != nil {
				ca.computeModePanel.Show()
			}
			if ca.sensePanel != nil {
				ca.sensePanel.Show()
			}
			// OpModeRead: no special panel needed (clean view)
		case OpModeRead:
			if ca.sensePanel != nil {
				ca.sensePanel.Show()
			}
		}
	})
}
