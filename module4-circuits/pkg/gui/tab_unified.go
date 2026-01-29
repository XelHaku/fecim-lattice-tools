// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device simulation view that replaces separate WRITE/READ/COMPUTE modes.
package gui

import (
	"fmt"
	"math/rand"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	configphysics "fecim-lattice-tools/config/physics"
	sharedphysics "fecim-lattice-tools/shared/physics"
	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

// ============================================================================
// UNIFIED DEVICE SIMULATION VIEW
// ============================================================================

// createUnifiedView creates the unified device simulation view
// Replaces the old mode-based createOperationsView()
func (ca *CircuitsApp) createUnifiedView() fyne.CanvasObject {
	// Initialize device state
	ca.deviceState = NewDeviceState(ca.arrayRows, ca.arrayCols, ca.tia, ca.adc)
	ca.operationsStatusLabel = widget.NewLabel("Ready")

	// In passive (0T1R) mode, all WLs are always active - no transistor gating
	if ca.architecture == sharedwidgets.Architecture0T1R {
		ca.deviceState.SetPassiveMode(true)
	}

	// 1. Signal chain header
	signalChainHeader := ca.createSignalChainHeader()

	// 2. Mode bar at top (Mode-First UX)
	modeBar := ca.createModeBar()

	// 3. Mode-specific panels (initially hidden, shown based on mode)
	writePanelContent := ca.createEnhancedWriteModePanel()
	ca.writeModePanel = container.NewVBox(writePanelContent)
	ca.writeModePanel.Hide() // Hidden by default (READ mode)

	computePanelContent := ca.createComputeModePanel()
	ca.computeModePanel = container.NewVBox(computePanelContent)
	ca.computeModePanel.Hide() // Hidden by default (READ mode)

	// Stack the mode panels (only one visible at a time)
	modePanelStack := container.NewStack(ca.writeModePanel, ca.computeModePanel)

	// Initialize architecture-specific voltage panels
	ca.passiveVoltagePanel = ca.createPassiveVoltagePanel()
	ca.activeVoltagePanel = ca.createActiveVoltagePanel()
	ca.passiveVoltagePanel.Hide() // Hidden initially (1T1R default)
	archVoltageStack := container.NewStack(ca.passiveVoltagePanel, ca.activeVoltagePanel)

	// 4. DAC input section
	dacSection := ca.createDACInputSection()

	// Update DAC range mode label with current voltage range
	ca.updateDACRangeModeLabel()

	// 5. Main visualization area (center)
	mainSection := ca.createMainSimSection()

	// 6. Action buttons (bottom)
	actionSection := ca.createUnifiedActionSection()

	// Initialize button states for default READ mode
	ca.updateActionButtons()

	// Top section: signal chain header, mode bar, mode panels, architecture voltage panels, DAC presets
	topSection := container.NewVBox(
		signalChainHeader,
		modeBar,
		modePanelStack,
		archVoltageStack,
		dacSection,
	)

	return container.NewBorder(
		topSection,    // top
		actionSection, // bottom
		nil, nil,
		mainSection, // center
	)
}

// createSignalChainHeader creates the signal chain indicator
func (ca *CircuitsApp) createSignalChainHeader() fyne.CanvasObject {
	// Architecture toggle
	archToggle := ca.createArchitectureToggle()

	// Material selector
	materialSelector := ca.createMaterialSelector()

	// ADC bits selector
	adcBitsSelector := ca.createADCBitsSelector()

	chainLabel := widget.NewLabelWithStyle(
		"SIGNAL CHAIN: DAC -> Array -> TIA -> ADC",
		fyne.TextAlignCenter,
		fyne.TextStyle{Bold: true},
	)

	// Operation classification label (updates based on configuration)
	ca.operationsModeHelp = widget.NewLabel("Configuration: Click cells or adjust voltages")
	ca.operationsModeHelp.TextStyle = fyne.TextStyle{Italic: true}

	// Circuit specs summary - shows current configuration
	adcBits := 5
	if ca.adc != nil {
		adcBits = ca.adc.Bits
	}
	adcLevels := 1 << adcBits
	circuitSpecsLabel := widget.NewLabel(fmt.Sprintf("ADC: %d-bit (%d levels, 0-%d)", adcBits, adcLevels, adcLevels-1))
	circuitSpecsLabel.TextStyle = fyne.TextStyle{Monospace: true}

	return container.NewVBox(
		container.NewHBox(
			chainLabel,
			layout.NewSpacer(),
			materialSelector,
			adcBitsSelector,
			layout.NewSpacer(),
			archToggle,
			layout.NewSpacer(),
			circuitSpecsLabel,
		),
		ca.operationsModeHelp,
		widget.NewSeparator(),
	)
}

// createMaterialSelector creates the ferroelectric material selection dropdown with browse button
func (ca *CircuitsApp) createMaterialSelector() fyne.CanvasObject {
	materials := sharedphysics.AllMaterials()
	materialNames := make([]string, len(materials))
	for i, m := range materials {
		materialNames[i] = m.Name
	}

	selector := widget.NewSelect(materialNames, func(selected string) {
		// Find the material and set it
		for _, m := range materials {
			if m.Name == selected {
				ca.deviceState.SetMaterial(m)
				ca.updateDACRangeModeLabel() // Update mode indicator
				ca.recomputeAndRefresh()
				ca.operationsStatusLabel.SetText(fmt.Sprintf("Material: %s (Vc=%.2fV)", selected, m.CoerciveVoltage()))
				break
			}
		}
	})

	// Set default selection to FeCIM material
	selector.SetSelected("FeCIM HZO")

	// Material picker button for detailed view
	pickerBtn := sharedwidgets.CreateMaterialPickerButton(
		ca.window,
		"fecim_hzo", // Default material ID
		func(materialID string, mat *configphysics.Material) {
			if mat == nil {
				return
			}
			// Update dropdown to match selection
			selector.SetSelected(mat.Name)
			// The dropdown's OnChanged will handle the rest
		},
	)

	return container.NewHBox(widget.NewLabel("Material:"), selector, pickerBtn)
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
	maxCols := min(8, ca.arrayCols)
	ca.unifiedDACEntries = make([]*widget.Entry, maxCols)
	ca.unifiedDACLabels = make([]*widget.Label, maxCols)

	// Range mode indicator - shows current DAC voltage range based on operation mode
	// Note: DAC range is set automatically by mode (READ/WRITE/COMPUTE)
	// Random input is available in COMPUTE mode panel
	ca.dacRangeLabel = widget.NewLabel("DAC: Read Range")
	ca.dacRangeLabel.TextStyle = fyne.TextStyle{Italic: true}

	// "Set All" entry for bulk voltage (manual override)
	allEntry := widget.NewEntry()
	allEntry.SetPlaceHolder("0.50")
	allEntry.OnSubmitted = func(s string) {
		ca.setAllUnifiedDACVoltages(s)
	}

	return container.NewHBox(
		ca.dacRangeLabel,
		layout.NewSpacer(),
		widget.NewLabel("Set All (V):"), allEntry,
	)
}

// updateDACRangeModeLabel updates the DAC range mode indicator based on operation mode
func (ca *CircuitsApp) updateDACRangeModeLabel() {
	if ca.dacRangeLabel == nil || ca.deviceState == nil {
		return
	}

	rangeMode := ca.deviceState.GetDACRangeMode()
	currentRange := ca.deviceState.GetCurrentVoltageRange()

	var text string
	if rangeMode == DACRangeWrite {
		text = fmt.Sprintf("DAC: Write (%.1f-%.1fV)", currentRange.Min, currentRange.Max)
	} else {
		text = fmt.Sprintf("DAC: Read (0-%.1fV)", currentRange.Max)
	}

	fyne.Do(func() {
		ca.dacRangeLabel.SetText(text)
	})
}

// createMainSimSection creates the main simulation visualization area
func (ca *CircuitsApp) createMainSimSection() fyne.CanvasObject {
	// WL checkboxes removed - row selection is done by clicking cells
	// WL state is determined automatically by mode and architecture:
	// - Passive (0T1R): All WLs always on
	// - 1T1R/2T1R READ/WRITE: Selected row only (via cell click)
	// - COMPUTE: All WLs on for MVM

	// Initialize empty WL checks array (some code may reference it)
	ca.unifiedWLChecks = make([]*widget.Check, 0)

	// Array canvas with DAC inputs at top, TIA/ADC outputs at right
	return ca.createUnifiedArraySection()
}

// setOperationMode sets the operation mode and configures WL/DAC accordingly
// READ: Single row, safe voltage (0-0.5V)
// WRITE: Single row, write voltage (1.2-1.5V on selected column)
// COMPUTE: All rows active, input vector (0-1V)
// NOTE: In passive mode (0T1R), all WLs are ALWAYS on - WL configuration is skipped
func (ca *CircuitsApp) setOperationMode(mode OpMode) {
	if ca.deviceState == nil {
		return
	}

	ca.deviceState.SetOperationMode(mode)

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
	ca.updateActionButtons() // Enable/disable action buttons based on mode
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

// createUnifiedArraySection creates the array visualization section
func (ca *CircuitsApp) createUnifiedArraySection() fyne.CanvasObject {
	// Create tappable array canvas - larger size for better visualization
	tappableArray := NewUnifiedTappableCanvas(ca, ca.drawUnifiedArray, ca.onUnifiedCellTapped)
	tappableArray.SetMinSize(fyne.NewSize(850, 600)) // Large canvas for detailed visualization
	ca.sharedArrayCanvas = tappableArray.raster

	// Cell info display
	ca.sharedCellInfoLabel = widget.NewLabel("Click a cell to select")

	// Array size info with capacity calculation
	totalCells := ca.arrayRows * ca.arrayCols
	bitCapacity := float64(totalCells) * 4.9 // ~4.9 bits per 30-level cell
	ca.sharedArrayInfoLabel = widget.NewLabel(fmt.Sprintf("Array: %dx%d (%d cells) | %d levels (~%.0f bits)",
		ca.arrayRows, ca.arrayCols, totalCells, ca.quantLevels, bitCapacity))

	// Legend with energy info
	legendLabel := widget.NewLabel("States: Low G (blue) -> High G (red) | Energy: READ ~45fJ, WRITE ~55fJ, MVM ~50fJ/cell")
	legendLabel.TextStyle = fyne.TextStyle{Italic: true}

	return container.NewVBox(
		tappableArray,
		legendLabel,
		ca.sharedCellInfoLabel,
		ca.sharedArrayInfoLabel,
	)
}


// createUnifiedActionSection creates the action buttons
func (ca *CircuitsApp) createUnifiedActionSection() fyne.CanvasObject {
	// Program button - only enabled in WRITE mode
	ca.actionWriteCellBtn = widget.NewButton("Write Cell", func() {
		ca.onUnifiedProgram()
	})
	ca.actionWriteCellBtn.Importance = widget.HighImportance

	// Sense button - only enabled in READ mode
	ca.actionReadBtn = widget.NewButton("Sense Row", func() {
		ca.onUnifiedRead()
	})

	// Compute button - only enabled in COMPUTE mode
	ca.actionComputeBtn = widget.NewButton("Compute MVM", func() {
		ca.onUnifiedCompute()
	})

	// Animate button
	animateBtn := widget.NewButton("Animate", func() {
		ca.onUnifiedAnimate()
	})

	// H3 FIX: Undo button
	ca.undoHistoryBtn = widget.NewButton("Undo", func() {
		ca.onUndo()
	})
	ca.undoHistoryBtn.Disable() // Initially disabled (no history)

	// Reset array button
	resetBtn := widget.NewButton("Reset Array", func() {
		ca.onUnifiedReset()
	})

	// Random array button
	randomBtn := widget.NewButton("Random Array", func() {
		ca.onUnifiedRandomArray()
	})

	return container.NewHBox(
		ca.actionWriteCellBtn, ca.actionReadBtn, ca.actionComputeBtn,
		layout.NewSpacer(),
		ca.undoHistoryBtn, animateBtn, randomBtn, resetBtn,
	)
}

// updateActionButtons enables/disables action buttons based on current mode
func (ca *CircuitsApp) updateActionButtons() {
	if ca.deviceState == nil {
		return
	}

	mode := ca.deviceState.GetOperationMode()

	fyne.Do(func() {
		// Write Cell: only in WRITE mode
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

		// Read/Sense: only in READ mode
		if ca.actionReadBtn != nil {
			if mode == OpModeRead {
				ca.actionReadBtn.Enable()
			} else {
				ca.actionReadBtn.Disable()
			}
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

	ca.deviceState.SetDACVoltage(col, voltage)
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
	ca.deviceState.SetSelectedCell(row, col)

	mode := ca.deviceState.GetOperationMode()
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R

	// In READ/WRITE mode (non-passive): select ONLY this row (single transistor)
	if !isPassive && (mode == OpModeRead || mode == OpModeWrite) {
		ca.deviceState.SetWLSingle(row)
	}

	// Update target cell label in write mode panel
	ca.updateWriteTargetLabel()

	// Cell click only selects the cell - does NOT apply voltages
	// Voltages are only applied when user presses action buttons

	ca.recomputeAndRefresh()
	ca.updateCellInfo()
}

// ============================================================================
// UI UPDATE HELPERS
// ============================================================================

// recomputeAndRefresh runs computation and updates all UI elements
func (ca *CircuitsApp) recomputeAndRefresh() {
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

	// Update operation classification
	ca.updateOperationClassification()

	// Refresh array canvas
	ca.refreshUnifiedArray()
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

	voltage := ca.deviceState.GetDACVoltage(selectedCol)
	matName := ca.deviceState.GetMaterialName()

	// Calculate expected current I = G x V
	expectedCurrent := conductanceUS * voltage // uA

	// Get actual row output (includes all cells in row if active)
	rowCurrent := ca.deviceState.GetRowCurrent(selectedRow)
	rowVoltage := ca.deviceState.GetRowVoltage(selectedRow)
	adcLevel := ca.deviceState.GetRowLevel(selectedRow)
	isActive := ca.deviceState.IsRowActive(selectedRow)

	fyne.Do(func() {
		// Build detailed info string with signal chain data
		var infoStr string
		if isActive && voltage > 0.01 {
			// Show full signal chain: G -> I -> TIA -> ADC
			infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fuS | BL=%.2fV -> I=%.1fuA -> TIA=%.2fV -> ADC=%d | %s",
				selectedRow, selectedCol, level, levels-1, conductanceUS, voltage, expectedCurrent, rowVoltage, adcLevel, matName)
		} else {
			// Cell not being sensed
			infoStr = fmt.Sprintf("Cell [%d,%d]: State %d/%d | G=%.1fuS | (Row %s, BL=%.2fV) | %s",
				selectedRow, selectedCol, level, levels-1, conductanceUS,
				map[bool]string{true: "ON", false: "OFF"}[isActive], voltage, matName)
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
			helpText = fmt.Sprintf("WRITE: Single row, %.1f-%.1fV. 2T1R selects single cell.", writeRange.Min, writeRange.Max)
		} else if arch == sharedwidgets.Architecture1T1R {
			helpText = fmt.Sprintf("WRITE: Single row, %.1f-%.1fV. 1T1R gates selected row.", writeRange.Min, writeRange.Max)
		} else {
			helpText = fmt.Sprintf("WRITE: %.1f-%.1fV. Passive: V/2 scheme reduces half-select disturb.", writeRange.Min, writeRange.Max)
		}
	case OpModeCompute:
		if arch == sharedwidgets.Architecture0T1R {
			helpText = fmt.Sprintf("COMPUTE: All rows, 0-%.1fV. Passive natural MVM mode (~75ns).", readRange.Max)
		} else {
			helpText = fmt.Sprintf("COMPUTE: All transistors ON, 0-%.1fV. Full MVM in ~75ns.", readRange.Max)
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
		ca.mfuxWriteTargetLabel.SetText(fmt.Sprintf("Target: Row %d, Col %d", row, col))
	})
}

// ============================================================================
// ARCHITECTURE TOGGLE
// ============================================================================

// createArchitectureToggle creates the PASSIVE/1T1R/2T1R toggle buttons
func (ca *CircuitsApp) createArchitectureToggle() fyne.CanvasObject {
	ca.archPassiveBtn = widget.NewButton("PASSIVE", nil)
	ca.arch1T1RBtn = widget.NewButton("1T1R", nil)
	ca.arch2T1RBtn = widget.NewButton("2T1R", nil)

	// Helper to update button styles based on selection
	updateArchButtons := func() {
		ca.archPassiveBtn.Importance = widget.LowImportance
		ca.arch1T1RBtn.Importance = widget.LowImportance
		ca.arch2T1RBtn.Importance = widget.LowImportance
		switch ca.architecture {
		case sharedwidgets.Architecture0T1R:
			ca.archPassiveBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture1T1R:
			ca.arch1T1RBtn.Importance = widget.HighImportance
		case sharedwidgets.Architecture2T1R:
			ca.arch2T1RBtn.Importance = widget.HighImportance
		default:
			ca.archPassiveBtn.Importance = widget.HighImportance
		}
		ca.archPassiveBtn.Refresh()
		ca.arch1T1RBtn.Refresh()
		ca.arch2T1RBtn.Refresh()
	}

	// Set initial state
	updateArchButtons()

	// Wire up callbacks
	ca.archPassiveBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture0T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture0T1R
		ca.mu.Unlock()
		updateArchButtons()
		// Passive mode: all WLs always active, cannot be changed
		ca.deviceState.SetPassiveMode(true)
		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	ca.arch1T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture1T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture1T1R
		ca.mu.Unlock()
		updateArchButtons()
		// 1T1R: disable passive mode, set WLs based on current operation mode
		ca.deviceState.SetPassiveMode(false)
		// Preserve WL state based on operation mode
		if ca.deviceState.GetOperationMode() == OpModeCompute {
			ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
		} else {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	ca.arch2T1RBtn.OnTapped = func() {
		if ca.architecture == sharedwidgets.Architecture2T1R {
			return
		}
		ca.mu.Lock()
		ca.architecture = sharedwidgets.Architecture2T1R
		ca.mu.Unlock()
		updateArchButtons()
		// 2T1R: disable passive mode, set WLs based on current operation mode
		ca.deviceState.SetPassiveMode(false)
		// Preserve WL state based on operation mode
		if ca.deviceState.GetOperationMode() == OpModeCompute {
			ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
		} else {
			ca.deviceState.SetWLSingle(ca.deviceState.GetSelectedRow())
		}
		ca.updateWLCheckboxesForArchitecture()
		ca.recomputeAndRefresh()
		ca.updateArchitectureSpecificUI()
	}

	ca.archToggle = container.NewGridWithColumns(3, ca.archPassiveBtn, ca.arch1T1RBtn, ca.arch2T1RBtn)

	archLabel := widget.NewLabel("Array:")
	return container.NewHBox(archLabel, ca.archToggle)
}

// ============================================================================
// MODE-FIRST UX PANELS (Phase 1)
// ============================================================================

// createModeBar creates the top-level mode selection bar
// This replaces the mode buttons previously buried in createWLSelector()
func (ca *CircuitsApp) createModeBar() fyne.CanvasObject {
	ca.modeReadBtn = widget.NewButton("READ", func() {
		ca.setOperationMode(OpModeRead)
	})
	ca.modeWriteBtn = widget.NewButton("WRITE", func() {
		ca.setOperationMode(OpModeWrite)
	})
	ca.modeComputeBtn = widget.NewButton("COMPUTE", func() {
		ca.setOperationMode(OpModeCompute)
	})

	// Set initial highlight (READ mode by default)
	ca.modeReadBtn.Importance = widget.HighImportance

	return container.NewHBox(
		widget.NewLabelWithStyle("Mode:", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ca.modeReadBtn,
		ca.modeWriteBtn,
		ca.modeComputeBtn,
		layout.NewSpacer(),
	)
}

// createWriteModePanel creates the write mode panel with level slider
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

// onWriteLevelChanged handles write level slider changes
// Only updates UI labels - does NOT apply voltage to DAC
// Voltage is only applied when user presses "Write Cell" button
func (ca *CircuitsApp) onWriteLevelChanged(level int) {
	if ca.deviceState == nil {
		return
	}

	// Calculate voltage for display only (don't apply to DAC)
	voltage := ca.deviceState.CalculateVoltageForState(level, ca.quantLevels)

	fyne.Do(func() {
		if ca.mfuxWriteLevelLabel != nil {
			ca.mfuxWriteLevelLabel.SetText(fmt.Sprintf("Level: %d", level))
		}
		if ca.mfuxWriteVoltageLabel != nil {
			ca.mfuxWriteVoltageLabel.SetText(fmt.Sprintf("Voltage: %.2fV", voltage))
		}
	})
}

// createComputeModePanel creates the compute mode panel with input vector entries
// This addresses UX-005: Input vector entries not visible
func (ca *CircuitsApp) createComputeModePanel() fyne.CanvasObject {
	maxCols := min(8, ca.arrayCols)
	ca.mfuxInputVectorEntry = make([]*widget.Entry, maxCols)
	ca.mfuxInputVectorLabels = make([]*widget.Label, maxCols)

	entriesBox := container.NewHBox()
	for i := 0; i < maxCols; i++ {
		idx := i
		entry := widget.NewEntry()
		entry.SetPlaceHolder("0")
		entry.SetText("0")
		entry.OnChanged = func(s string) {
			ca.onInputVectorEntryChanged(idx, s)
		}
		ca.mfuxInputVectorEntry[i] = entry

		label := widget.NewLabel(fmt.Sprintf("x%d", i))
		label.TextStyle = fyne.TextStyle{Monospace: true}
		ca.mfuxInputVectorLabels[i] = label

		// Each column: label above entry
		col := container.NewVBox(label, entry)
		entriesBox.Add(col)
	}

	// Random button to populate with random values
	randomBtn := widget.NewButton("Random", func() {
		ca.randomizeInputVectorEntries()
	})

	// Clear button
	clearBtn := widget.NewButton("Clear", func() {
		ca.clearInputVectorEntries()
	})

	titleLabel := widget.NewLabelWithStyle("Input Vector (0-255):", fyne.TextAlignLeading, fyne.TextStyle{Bold: true})

	return container.NewVBox(
		titleLabel,
		entriesBox,
		container.NewHBox(randomBtn, clearBtn),
	)
}

// onInputVectorEntryChanged handles input vector entry changes
// Only applies DAC changes in COMPUTE mode to prevent state corruption
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
		// OpModeRead: no special panel needed (clean view)
		}
	})
}
