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
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	configphysics "fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
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
	maxSenseRfKOhm      = 100.0
	minSenseADCVref     = 0.0
	maxSenseADCVref     = 1.5
	minSenseADCVrefSpan = 1e-3
)

func formatCurrentA(currentA float64) string {
	absCurrent := math.Abs(currentA)
	unit := "A"
	scale := 1.0
	switch {
	case absCurrent >= 1:
		unit = "A"
		scale = 1.0
	case absCurrent >= 1e-3:
		unit = "mA"
		scale = 1e-3
	case absCurrent >= 1e-6:
		unit = "uA"
		scale = 1e-6
	case absCurrent >= 1e-9:
		unit = "nA"
		scale = 1e-9
	default:
		unit = "pA"
		scale = 1e-12
	}
	return fmt.Sprintf("%.3g %s", currentA/scale, unit)
}

// ============================================================================
// UNIFIED DEVICE SIMULATION VIEW
// ============================================================================

// createUnifiedView creates the unified device simulation view
// Layout: Controls at TOP (toolbar), Array canvas in CENTER (expands), minimal status at BOTTOM
func (ca *CircuitsApp) createUnifiedView() fyne.CanvasObject {
	// Initialize device state
	ca.deviceState = NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)
	ca.operationsStatusLabel = widget.NewLabel("Ready")

	// In passive (0T1R) mode, all WLs are always active - no transistor gating
	if ca.architecture == sharedwidgets.Architecture0T1R {
		ca.deviceState.SetPassiveMode(true)
	}

	// ============================================================
	// TOP: Compact Toolbar (~100px total)
	// ============================================================

	// Row 1: Config + Mode + Architecture (unified row)
	configModeRow := ca.createUnifiedConfigModeRow()

	// Row 2: Action buttons
	actionRow := ca.createUnifiedActionRow()

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
	ca.sharedArrayCanvas = tappableArray.raster

	// Initialize empty WL checks array (some code may reference it)
	ca.unifiedWLChecks = make([]*widget.Check, 0)

	// Cell info display (updated on cell click)
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")

	// Array info (updated on resize)
	totalCells := ca.arrayRows * ca.arrayCols
	bitCapacity := float64(totalCells) * 4.9
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("%dx%d array | %d levels | ~%.0f bits",
		ca.arrayRows, ca.arrayCols, ca.quantLevels, bitCapacity))

	// Compact info row below canvas
	infoRow := container.NewHBox(
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

	// Initialize architecture info (compact single-line)
	ca.passiveVoltagePanel = ca.createCompactPassivePanel()
	ca.activeVoltagePanel = ca.createCompactActivePanel()
	ca.passiveVoltagePanel.Hide() // Hidden initially (1T1R default)

	statusBar := container.NewHBox(
		ca.operationsModeHelp,
		layout.NewSpacer(),
		widget.NewLabel("DAC -> Array -> TIA -> ADC"),
	)

	// Initialize button states for default READ mode
	ca.updateActionButtons()
	ca.updateModePanels(ca.deviceState.GetOperationMode())
	ca.updateSensePanel()

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

	// ADC bits selector
	adcBitsSelector := ca.createADCBitsSelector()

	// Coupling model toggle
	couplingToggle := ca.createCouplingToggle()

	// Mode buttons
	ca.modeReadBtn = NewTooltipButton(
		"READ",
		"READ mode: safe read voltage on selected column; no programming.",
		ca.window,
		func() { ca.setOperationMode(OpModeRead) },
	)
	ca.modeWriteBtn = NewTooltipButton(
		"WRITE",
		"WRITE mode: arms write range and WL gating; no voltage until Program Cell.",
		ca.window,
		func() { ca.setOperationMode(OpModeWrite) },
	)
	ca.modeComputeBtn = NewTooltipButton(
		"COMPUTE",
		"COMPUTE mode: applies input vector across all rows for MVM.",
		ca.window,
		func() { ca.setOperationMode(OpModeCompute) },
	)

	// Set initial highlight (READ mode by default)
	ca.modeReadBtn.Importance = widget.HighImportance

	// Architecture toggle
	archToggle := ca.createArchitectureToggle()

	// Single row: Material | Array | ADC | Sep | Mode buttons | Spacer | Architecture
	return container.NewHBox(
		materialSelector,
		arraySizeSelector,
		adcBitsSelector,
		couplingToggle,
		widget.NewSeparator(),
		widget.NewLabel("Mode:"),
		ca.modeReadBtn,
		ca.modeWriteBtn,
		ca.modeComputeBtn,
		layout.NewSpacer(),
		archToggle,
	)
}

// createUnifiedActionRow creates the action buttons row
func (ca *CircuitsApp) createUnifiedActionRow() fyne.CanvasObject {
	// Primary action buttons
	ca.actionWriteCellBtn = NewTooltipButton(
		"Program Cell",
		"Apply DAC write pulse to selected cell (ISPP). Passive arrays use V/2 half-select.",
		ca.window,
		func() { ca.onUnifiedProgram() },
	)
	ca.actionWriteCellBtn.Importance = widget.HighImportance

	ca.actionComputeBtn = widget.NewButton("MVM", func() {
		ca.onUnifiedCompute()
	})

	// Utility buttons
	ca.undoHistoryBtn = widget.NewButton("Undo", func() {
		ca.onUndo()
	})
	ca.undoHistoryBtn.Disable()

	ca.actionRandomArrayBtn = widget.NewButton("Random Array", func() {
		ca.onUnifiedRandomArray()
	})

	ca.actionResetArrayBtn = widget.NewButton("Reset Array", func() {
		ca.onUnifiedReset()
	})

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
		fyne.Do(func() {
			if ca.sharedArrayCanvas != nil {
				ca.sharedArrayCanvas.Refresh()
			}
		})
	}

	ca.actionFitBtn = widget.NewButton("Fit", func() {
		logAction("button_zoom_fit")
		ca.zoomSlider.SetValue(1.0)
	})

	// Export button
	exportBtn := widget.NewButton("Export", func() {
		logAction("button_export")
		ca.exportSimulationData()
	})

	// Row: Program Cell | MVM | Sep | Undo | Random Array | Reset Array | Export | Sep | Zoom controls | Spacer | Tools status
	return container.NewHBox(
		ca.actionWriteCellBtn,
		ca.actionComputeBtn,
		widget.NewSeparator(),
		ca.undoHistoryBtn,
		ca.actionRandomArrayBtn,
		ca.actionResetArrayBtn,
		exportBtn,
		widget.NewSeparator(),
		widget.NewLabel("Zoom:"),
		ca.zoomSlider,
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

	idealBtn := widget.NewButton("Ideal", nil)
	approxBtn := widget.NewButton("Approx (Tier A)", nil)

	apply := func(mode arraysim.CouplingMode) {
		idealBtn.Importance = widget.LowImportance
		approxBtn.Importance = widget.LowImportance
		switch mode {
		case arraysim.CouplingTierA:
			approxBtn.Importance = widget.HighImportance
		default:
			idealBtn.Importance = widget.HighImportance
		}
		idealBtn.Refresh()
		approxBtn.Refresh()
	}

	setMode := func(mode arraysim.CouplingMode) {
		if current == mode {
			return
		}
		current = mode
		if ca.deviceState != nil {
			ca.deviceState.SetCouplingMode(mode)
		}
		apply(mode)
		ca.recomputeAndRefresh()
	}

	idealBtn.OnTapped = func() { setMode(arraysim.CouplingIdeal) }
	approxBtn.OnTapped = func() { setMode(arraysim.CouplingTierA) }

	ca.couplingIdealBtn = idealBtn
	ca.couplingApproxBtn = approxBtn
	ca.couplingToggle = container.NewGridWithColumns(2, idealBtn, approxBtn)

	apply(current)

	return container.NewHBox(widget.NewLabel("Coupling:"), ca.couplingToggle)
}

// ValidArraySizes defines the supported array dimensions
var ValidArraySizes = []int{1, 2, 4, 8, 16, 32, 64}

// createArraySizeSelector creates a dropdown to select array size (1x1 to 128x128)
func (ca *CircuitsApp) createArraySizeSelector() fyne.CanvasObject {
	options := make([]string, len(ValidArraySizes))
	for i, size := range ValidArraySizes {
		options[i] = fmt.Sprintf("%dx%d", size, size)
	}

	selector := widget.NewSelect(options, func(selected string) {
		// Parse size from "NxN" format
		var rows, cols int
		n, _ := fmt.Sscanf(selected, "%dx%d", &rows, &cols)
		if n == 2 && rows > 0 && rows <= MaxArraySize && cols > 0 && cols <= MaxArraySize {
			logInput("array_size=%s", selected)
			ca.resizeArray(rows, cols)
		}
	})

	// Set default selection
	selector.SetSelected(fmt.Sprintf("%dx%d", ca.arrayRows, ca.arrayCols))

	return container.NewHBox(widget.NewLabel("Array:"), selector)
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
	bitCapacity := float64(totalCells) * 4.9 // ~4.9 bits per 30-level cell
	fyne.Do(func() {
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

	return container.NewHBox(widget.NewLabel("Material:"), ca.materialBtn)
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

// createADCBitsSelector creates a dropdown to select ADC resolution (5-8 bits)
func (ca *CircuitsApp) createADCBitsSelector() fyne.CanvasObject {
	options := []string{"5-bit (32)", "6-bit (64)", "7-bit (128)", "8-bit (256)"}

	selector := widget.NewSelect(options, func(selected string) {
		var bits int
		switch selected {
		case "5-bit (32)":
			bits = 5
		case "6-bit (64)":
			bits = 6
		case "7-bit (128)":
			bits = 7
		case "8-bit (256)":
			bits = 8
		default:
			bits = 5
		}
		logInput("adc_bits=%d", bits)
		ca.deviceState.SetADCBits(bits)
		ca.recomputeAndRefresh()
		levels := 1 << bits
		ca.operationsStatusLabel.SetText(fmt.Sprintf("ADC: %d-bit (%d levels, 0-%d)", bits, levels, levels-1))
	})

	// Set default selection
	selector.SetSelected("5-bit (32)")

	return container.NewHBox(widget.NewLabel("ADC:"), selector)
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

	fyne.Do(func() {
		// Only show in WRITE mode where voltage range matters
		if mode == OpModeWrite {
			var text string
			if rangeMode == DACRangeWrite {
				text = fmt.Sprintf("DAC: Write (%.1f-%.1fV)", currentRange.Min, currentRange.Max)
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
// so WL configuration is skipped and WRITE uses a V/2 half-select scheme.
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
		readVoltage := ca.deviceState.GetReadRange().Max * 0.4 // ~0.2V safe read
		if readVoltage < 0.1 {
			readVoltage = 0.2
		}
		ca.deviceState.SetAllDACVoltages(0)
		ca.deviceState.SetDACVoltage(ca.deviceState.GetSelectedCol(), readVoltage)

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
	ca.updateWLCheckboxes()
	ca.updateWLHelpLabel()
	ca.updateDACRangeModeLabel()
	ca.recomputeAndRefresh()
}

// updateWLHelpLabel is a no-op - WL UI has been removed
func (ca *CircuitsApp) updateWLHelpLabel() {
	// No-op: WL UI removed
}

// updateModeButtons updates the mode button highlighting
func (ca *CircuitsApp) updateModeButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()

	fyne.Do(func() {
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

// updateActionButtons enables/disables action buttons based on current mode
func (ca *CircuitsApp) updateActionButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()

	fyne.Do(func() {
		// Program Cell: only in WRITE mode
		if ca.actionWriteCellBtn != nil {
			if mode == OpModeWrite {
				ca.actionWriteCellBtn.Enable()
				ca.actionWriteCellBtn.Importance = widget.HighImportance
			} else {
				ca.actionWriteCellBtn.Disable()
				ca.actionWriteCellBtn.Importance = widget.MediumImportance
			}
			ca.actionWriteCellBtn.Refresh()
		}

		// Compute MVM: only in COMPUTE mode
		if ca.actionComputeBtn != nil {
			if mode == OpModeCompute {
				ca.actionComputeBtn.Enable()
			} else {
				ca.actionComputeBtn.Disable()
			}
		}
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

	// Clamp voltage to reasonable range
	if voltage < 0 {
		voltage = 0
	}
	if voltage > 2.0 {
		voltage = 2.0
	}

	logInput("dac_col=%d voltage=%.3f", col, voltage)
	ca.deviceState.SetDACVoltage(col, voltage)
	ca.recomputeAndRefresh()
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
		fyne.Do(func() {
			ca.senseRfEntry.SetText(fmt.Sprintf("%.1f", rfKOhm))
		})
	}
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
		fyne.Do(func() {
			if ca.senseAdcVminEntry != nil {
				ca.senseAdcVminEntry.SetText(fmt.Sprintf("%.2f", vmin))
			}
			if ca.senseAdcVmaxEntry != nil {
				ca.senseAdcVmaxEntry.SetText(fmt.Sprintf("%.2f", vmax))
			}
		})
	}
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
	ca.updateDACEntries()
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
	ca.updateDACEntries()
	ca.recomputeAndRefresh()
}

// onWLChanged is a no-op - WL checkboxes have been removed
// Row selection is now done via cell clicks in onUnifiedCellTapped
func (ca *CircuitsApp) onWLChanged(row int, checked bool) {
	// No-op: WL checkboxes removed from UI
}

// setWLModeSingle sets WL mode to single (only selected row)
func (ca *CircuitsApp) setWLModeSingle() {
	selectedRow := ca.deviceState.GetSelectedRow()
	ca.deviceState.SetWLSingle(selectedRow)
	ca.updateWLCheckboxes()
	ca.recomputeAndRefresh()
}

// setWLModeAll sets WL mode to all rows active
func (ca *CircuitsApp) setWLModeAll() {
	ca.deviceState.SetWLAll()
	ca.updateWLCheckboxes()
	ca.recomputeAndRefresh()
}

// onUnifiedCellTapped handles cell selection
// In READ/WRITE mode: selects row transistor (WL) and column transistor (BL)
func (ca *CircuitsApp) onUnifiedCellTapped(row, col int) {
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

// recomputeAndRefreshNow forces an immediate recompute + UI refresh.
func (ca *CircuitsApp) recomputeAndRefreshNow() {
	ca.mu.RLock()
	weights := ca.arrayWeights
	levels := ca.quantLevels
	ca.mu.RUnlock()

	// Run device simulation
	ca.deviceState.Compute(weights, levels)

	// Update output display
	ca.updateOutputDisplay()

	// Update cell info
	ca.updateCellInfo()
	ca.updateSensePanel()

	// Update operation classification
	ca.updateOperationClassification()

	// Refresh array canvas
	ca.refreshUnifiedArray()
}

func (ca *CircuitsApp) scheduleRecomputeAndRefresh(force bool) {
	if force {
		ca.recomputeAndRefreshNow()
		return
	}

	now := time.Now()
	ca.uiUpdateMu.Lock()
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
	ca.uiUpdateMu.Unlock()

	time.AfterFunc(delay, func() {
		ca.uiUpdateMu.Lock()
		ca.lastUIUpdate = time.Now()
		ca.pendingUIUpd = false
		ca.uiUpdateMu.Unlock()
		ca.recomputeAndRefreshNow()
	})
}

// refreshUnifiedArray refreshes the array canvas
func (ca *CircuitsApp) refreshUnifiedArray() {
	if ca.sharedArrayCanvas != nil {
		fyne.Do(func() {
			ca.sharedArrayCanvas.Refresh()
		})
	}
}

// updateDACEntries is a no-op - DAC values are shown on the diagram
func (ca *CircuitsApp) updateDACEntries() {
	// DAC values are displayed in the array diagram's DAC boxes
}

// updateWLCheckboxes is a no-op - WL checkboxes have been removed
// WL state is now managed automatically based on mode and architecture
func (ca *CircuitsApp) updateWLCheckboxes() {
	// No-op: WL checkboxes removed from UI
	// Row selection is done via cell clicks
}

// updateWLCheckboxesForArchitecture is a no-op - WL UI has been removed
func (ca *CircuitsApp) updateWLCheckboxesForArchitecture() {
	// No-op: WL UI removed
}

// updateOutputDisplay is a no-op - outputs are shown on the diagram
func (ca *CircuitsApp) updateOutputDisplay() {
	// Outputs are displayed in the array diagram's TIA/ADC boxes
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
	expectedCurrent := conductanceUS * math.Abs(effectiveVoltage) // uA

	// Get actual row output (includes all cells in row if active)
	rowCurrent := ca.deviceState.GetRowCurrent(selectedRow)
	rowVoltage := ca.deviceState.GetRowVoltage(selectedRow)
	adcLevel := ca.deviceState.GetRowLevel(selectedRow)
	isActive := ca.deviceState.IsRowActive(selectedRow)
	isPassive := ca.deviceState.IsPassiveMode()

	fyne.Do(func() {
		// Build detailed info string with signal chain data
		var infoStr string
		if isActive && math.Abs(effectiveVoltage) > 0.01 {
			// Show full signal chain: G -> I -> TIA -> ADC
			if isPassive {
				infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fuS | WL=%.2fV BL=%.2fV -> Vcell=%.2fV -> I=%.1fuA -> TIA=%.2fV -> ADC=%d | %s",
					selectedRow, selectedCol, level, levels-1, conductanceUS, wlVoltage, blVoltage, effectiveVoltage, expectedCurrent, rowVoltage, adcLevel, matName)
			} else {
				infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fuS | BL=%.2fV -> I=%.1fuA -> TIA=%.2fV -> ADC=%d | %s",
					selectedRow, selectedCol, level, levels-1, conductanceUS, blVoltage, expectedCurrent, rowVoltage, adcLevel, matName)
			}
		} else {
			// Cell not being sensed
			infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fuS | (Row %s, BL=%.2fV) | %s",
				selectedRow, selectedCol, level, levels-1, conductanceUS,
				map[bool]string{true: "ON", false: "OFF"}[isActive], blVoltage, matName)
		}
		ca.sharedCellInfoLabel.SetText(infoStr)
	})

	// Also update array info label with total row current
	if ca.sharedArrayInfoLabel != nil {
		fyne.Do(func() {
			ca.sharedArrayInfoLabel.SetText(fmt.Sprintf("Array: %dx%d | %d levels | Row %d sum: I=%.1fuA",
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

// updateSensePanel updates the compact sense-chain readout.
func (ca *CircuitsApp) updateSensePanel() {
	if ca.sensePanel == nil || ca.deviceState == nil {
		return
	}

	row := ca.deviceState.GetSelectedRow()
	currentA := ca.deviceState.GetRowCurrent(row) * 1e-6
	voltage := ca.deviceState.GetRowVoltage(row)
	code := ca.deviceState.GetRowLevel(row)
	levels := ca.deviceState.GetADCLevels()
	saturated := ca.deviceState.IsSaturated(row)
	isActive := ca.deviceState.IsRowActive(row)
	mode := ca.deviceState.GetOperationMode()

	rangeText := "n/a"
	lsbText := "n/a"
	if sense, ok := ca.senseChainConfig(); ok {
		imin, imax := sense.CurrentRange()
		lsb := sense.CurrentLSB()
		if lsb > 0 {
			rangeText = fmt.Sprintf("%s .. %s", formatCurrentA(imin), formatCurrentA(imax))
			lsbText = formatCurrentA(lsb)
		}
	}

	codeText := fmt.Sprintf("%d", code)
	if levels > 1 {
		codeText = fmt.Sprintf("%d/%d", code, levels-1)
	}

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

	satText := "OK"
	if saturated {
		satText = "SAT"
	}

	fyne.Do(func() {
		if ca.senseTitleLabel != nil {
			ca.senseTitleLabel.SetText(titleText)
		}
		if ca.senseRowLabel != nil {
			ca.senseRowLabel.SetText(rowText)
		}
		if ca.senseCurrentLabel != nil {
			ca.senseCurrentLabel.SetText(formatCurrentA(currentA))
		}
		if ca.senseVoltageLabel != nil {
			ca.senseVoltageLabel.SetText(fmt.Sprintf("%.3f V", voltage))
		}
		if ca.senseCodeLabel != nil {
			ca.senseCodeLabel.SetText(codeText)
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
			helpText = fmt.Sprintf("WRITE: %.1f-%.1fV. Passive: V/2 scheme reduces half-select disturb. Use Program Cell to apply pulse.", writeRange.Min, writeRange.Max)
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

	fyne.Do(func() {
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

	fyne.Do(func() {
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

		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	toggle := sharedwidgets.NewArchitectureToggle(sharedwidgets.ArchitectureToggleOptions{
		Initial:      ca.architecture,
		Style:        sharedwidgets.ArchitectureToggleStylePlain,
		LabelPassive: "PASSIVE",
		Label1T1R:    "1T1R",
		Label2T1R:    "2T1R",
		OnChanged:    handleArchChange,
	})

	ca.archPassiveBtn = toggle.PassiveButton
	ca.arch1T1RBtn = toggle.OneT1RButton
	ca.arch2T1RBtn = toggle.TwoT1RButton
	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	archLabel := widget.NewLabel("Array:")
	return container.NewHBox(archLabel, ca.archToggle)
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

	fyne.Do(func() {
		if ca.mfuxWriteLevelLabel != nil {
			ca.mfuxWriteLevelLabel.SetText(fmt.Sprintf("L:%d", level))
		}
		if ca.mfuxWriteVoltageLabel != nil {
			if math.Abs(appliedVoltage-targetVoltage) > 0.01 {
				ca.mfuxWriteVoltageLabel.SetText(fmt.Sprintf("%.2fV (DAC %.2fV)", targetVoltage, appliedVoltage))
			} else {
				ca.mfuxWriteVoltageLabel.SetText(fmt.Sprintf("%.2fV", appliedVoltage))
			}
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

	ca.senseCurrentLabel = widget.NewLabel("0 A")
	ca.senseVoltageLabel = widget.NewLabel("0.000 V")
	ca.senseCodeLabel = widget.NewLabel("0")
	ca.senseSaturationLabel = widget.NewLabel("OK")

	metricsRow := container.NewHBox(
		widget.NewLabel("Irow (A):"), ca.senseCurrentLabel,
		widget.NewLabel("Vout (V):"), ca.senseVoltageLabel,
		widget.NewLabel("ADC:"), ca.senseCodeLabel,
		widget.NewLabel("SAT:"), ca.senseSaturationLabel,
	)

	ca.senseRangeLabel = widget.NewLabel("n/a")
	ca.senseLSBLabel = widget.NewLabel("n/a")

	rangeTooltip := NewTooltipButton("?", "Measurable current range after TIA rails and ADC references.\nI = (V - Vref) / Rf, using Vmin/Vmax = max/min(TIA, ADC).", ca.window, nil)
	rangeTooltip.Importance = widget.LowImportance
	lsbTooltip := NewTooltipButton("?", "LSB current = (Vmax_eff - Vmin_eff) / (2^bits - 1) / Rf.", ca.window, nil)
	lsbTooltip.Importance = widget.LowImportance

	rangeRow := container.NewHBox(
		widget.NewLabel("I_range (A):"), ca.senseRangeLabel, rangeTooltip,
		widget.NewLabel("LSB (A):"), ca.senseLSBLabel, lsbTooltip,
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

	controlsRow := container.NewHBox(
		widget.NewLabel("Rf (kΩ):"), ca.senseRfEntry,
		widget.NewLabel("ADC Vmin (V):"), ca.senseAdcVminEntry,
		widget.NewLabel("Vmax (V):"), ca.senseAdcVmaxEntry,
	)

	return container.NewVBox(
		headerRow,
		metricsRow,
		rangeRow,
		controlsRow,
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
		fyne.Do(func() {
			ca.computeInputTitle.SetText(fmt.Sprintf("Input Vector (%d inputs, 0–255 → 0–%.2fV):", ca.arrayCols, readMax))
		})
	}

	// Rebuild entries
	fyne.Do(func() {
		ca.buildComputeInputEntries()
		ca.computeInputContainer.Refresh()
	})
}

// onInputVectorEntryChanged handles input vector code changes.
//
// Physics meaning:
//   - Each entry is a digital code d (0..255) mapped to an analog column voltage
//     V = (d/255) * readRange.Max for compute-safe MVM.
//
// Bounds / clamping:
//   - Codes are clamped to [0,255].
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
	if value > 255 {
		value = 255
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
		ca.inputVector[i] = rand.Intn(256)
		valuesCopy[i] = ca.inputVector[i]
	}
	ca.mu.Unlock()
	logAction("input_vector_randomized len=%d", len(valuesCopy))

	// Update entry widgets (no lock - use copy)
	fyne.Do(func() {
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
	fyne.Do(func() {
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
	fyne.Do(func() {
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
