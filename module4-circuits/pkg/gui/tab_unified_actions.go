// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains action handlers for the unified view.
package gui

import (
	"fmt"
	"math/rand"
	"time"

	sharedwidgets "fecim-lattice-tools/shared/widgets"
)

func formatReadStatusLine(row, col, level int, currentUA, tiaVoltageV float64, adcCode int) string {
	return fmt.Sprintf("READ [%d,%d]: State=%d | I=%+.2f µA -> TIA=%+.2f V -> ADC=%d | ~76ns, ~46fJ",
		row, col, level, currentUA, tiaVoltageV, adcCode)
}

// ============================================================================
// ACTION HANDLERS
// ============================================================================

// onUnifiedProgram programs the selected cell using Write-ReadVerify loop
// This simulates ISPP (Incremental Step Pulse Programming) behavior
func (ca *CircuitsApp) onUnifiedProgram() {
	if ca.isProgrammingActive() {
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText("PROGRAMMING — controls locked")
		}
		return
	}

	// Check if ISPP already in progress - prevent concurrent operations
	isppStatus := ca.deviceState.GetISPPStatus()
	if isppStatus.Active {
		ca.setProgrammingActive(true)
		return
	}

	// Mode validation: only allowed in WRITE mode
	if ca.deviceState.GetOperationMode() != OpModeWrite {
		ca.operationsStatusLabel.SetText("Error: Switch to WRITE mode first")
		return
	}

	// Get target level directly from slider (the user's intent)
	targetLevel := int(ca.mfuxWriteLevelSlider.Value)
	if targetLevel < 0 {
		targetLevel = 0
	}
	if targetLevel >= ca.quantLevels {
		targetLevel = ca.quantLevels - 1
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()
	logAction("write_start row=%d col=%d target=%d", selectedRow, selectedCol, targetLevel)

	// H3 FIX: Save current state to undo history before modifying
	ca.saveUndoHistory()

	ca.setProgrammingActive(true)

	// Run Write-ReadVerify loop in background goroutine
	go ca.runISPPWithAnimation(selectedRow, selectedCol, targetLevel)
}

// writeReadVerifyLoop performs animated Write-ReadVerify iterations
// Simulates ISPP: apply pulse, read back, adjust if needed, repeat until target reached
func (ca *CircuitsApp) writeReadVerifyLoop(row, col, targetLevel int, startVoltage float64) {
	const maxIterations = 5
	const iterationDelay = 300 * time.Millisecond

	writeRange := ca.deviceState.GetWriteRange()
	voltage := startVoltage
	currentLevel := 0

	// Get current level
	ca.mu.Lock()
	if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
		currentLevel = ca.arrayWeights[row][col]
	}
	ca.mu.Unlock()

	// Passive (0T1R) mode: DAC drives the selected BL; all WLs are grounded.
	isPassive := ca.deviceState.IsPassiveMode()
	if isPassive {
		defer func() {
			ca.deviceState.DisableHalfSelectVisualization()
			ca.updateHalfSelectVisualization()
		}()
	}

	for iteration := 1; iteration <= maxIterations; iteration++ {
		// === WRITE PHASE ===
		appliedVoltage, _ := ca.applyWriteVoltages(row, col, voltage)
		if isPassive {
			sharedwidgets.SafeDo(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("WRITE [%d,%d]: Column drive BL=−%.2fV WL=0V (iter %d/%d)",
					row, col, appliedVoltage, iteration, maxIterations))
			})
			ca.deviceState.EnableHalfSelectVisualization(row, col, appliedVoltage)
			ca.updateHalfSelectVisualization()
		} else {
			sharedwidgets.SafeDo(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("WRITE [%d,%d]: V=%.2fV (iter %d/%d)", row, col, appliedVoltage, iteration, maxIterations))
			})
		}
		ca.applyHalfSelectDisturb(row, col)
		ca.recomputeAndRefresh()
		time.Sleep(iterationDelay / 2)

		// Simulate write: move current level toward target
		// In real hardware, the level change depends on voltage amplitude
		ca.mu.Lock()
		if row < len(ca.arrayWeights) && col < len(ca.arrayWeights[row]) {
			if currentLevel < targetLevel {
				// Increase by 1-2 levels per pulse (simulated partial switching)
				step := 1
				if targetLevel-currentLevel > 3 {
					step = 2
				}
				currentLevel += step
				if currentLevel > targetLevel {
					currentLevel = targetLevel
				}
			} else if currentLevel > targetLevel {
				// Decrease by 1-2 levels per pulse
				step := 1
				if currentLevel-targetLevel > 3 {
					step = 2
				}
				currentLevel -= step
				if currentLevel < targetLevel {
					currentLevel = targetLevel
				}
			}
			ca.arrayWeights[row][col] = currentLevel
		}
		ca.mu.Unlock()

		// === READ/VERIFY PHASE ===
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("VERIFY [%d,%d]: Read level %d (target %d)", row, col, currentLevel, targetLevel))
		})

		// Reset write voltages before applying read voltage
		// Reset write voltages and return array to safe idle state
		ca.deviceState.ResetWriteVoltages()

		// Set DAC to read voltage for verification
		readVoltage := ca.deviceState.GetReadRange().Max * 0.5
		ca.deviceState.SetDACVoltage(col, readVoltage)
		ca.recomputeAndRefresh()
		time.Sleep(iterationDelay / 2)

		// Check if target reached
		// Write cycle: ~203ns per iteration, Energy: ~2.2pJ (pump + verify)
		if currentLevel == targetLevel {
			totalTimeNs := iteration * 203
			totalEnergyFJ := iteration * 2200
			sharedwidgets.SafeDo(func() {
				ca.operationsStatusLabel.SetText(fmt.Sprintf("WRITE [%d,%d] = State %d | %d iter | ~%dns, ~%dfJ",
					row, col, targetLevel, iteration, totalTimeNs, totalEnergyFJ))
			})
			// Return all voltages to 0 (safe idle state)
			ca.deviceState.ResetWriteVoltages()
			ca.recomputeAndRefresh()
			return
		}

		// Adjust voltage for next iteration (ISPP: increment voltage if undershoot)
		if currentLevel < targetLevel {
			// Need higher voltage to switch more domains
			voltageStep := (writeRange.Max - writeRange.Min) / 80.0 // ~1.25% of range per step
			voltage += voltageStep
			if voltage > writeRange.Max {
				voltage = writeRange.Max
			}
		} else {
			// Need lower voltage (less aggressive write)
			voltageStep := (writeRange.Max - writeRange.Min) / 80.0 // ~1.25% of range per step
			voltage -= voltageStep
			if voltage < writeRange.Min {
				voltage = writeRange.Min
			}
		}
	}

	// Max iterations reached
	sharedwidgets.SafeDo(func() {
		ca.operationsStatusLabel.SetText(fmt.Sprintf("PARTIAL [%d,%d] = State %d (target was %d, max iterations)",
			row, col, currentLevel, targetLevel))
	})
	// Return all voltages to 0 (safe idle state)
	ca.deviceState.ResetWriteVoltages()
	ca.recomputeAndRefresh()
}

// onUnifiedRead senses the selected cell per VOLTAGE_RULES.md:
// - Selected WL: Active (1T1R) or read voltage (0T1R)
// - Selected BL: Read voltage (0.1-0.5V)
// - Unselected BLs: 0V (grounded to minimize sneak paths)
func (ca *CircuitsApp) onUnifiedRead() {
	// Mode validation: only allowed in READ mode
	if ca.deviceState.GetOperationMode() != OpModeRead {
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText("Error: Switch to READ mode first")
		})
		return
	}

	selectedRow := ca.deviceState.GetSelectedRow()
	selectedCol := ca.deviceState.GetSelectedCol()
	logAction("read row=%d col=%d", selectedRow, selectedCol)

	// Per VOLTAGE_RULES.md: Only selected row active
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R
	if !isPassive {
		ca.deviceState.SetWLSingle(selectedRow)
	}

	// Per VOLTAGE_RULES.md Section 3.1 and 4.1:
	// - Selected BL: Read voltage (0.2V typical)
	// - Unselected BLs: 0V (ground)
	// Ground all columns first, then apply read voltage only to selected column
	ca.deviceState.SetAllDACVoltages(0)
	ca.deviceState.SetDACVoltage(selectedCol, ca.safeReadVoltage())
	ca.deviceState.SetDACRangeMode(DACRangeRead)

	// Recompute with proper biasing
	ca.recomputeAndRefresh()

	// Get results for the selected cell
	current := ca.deviceState.GetRowCurrent(selectedRow)
	tiaVoltage := ca.deviceState.GetRowVoltage(selectedRow)
	adcLevel := ca.deviceState.GetRowLevel(selectedRow)

	// Get the cell's conductance for display
	ca.mu.RLock()
	level := 0
	if selectedRow < len(ca.arrayWeights) && selectedCol < len(ca.arrayWeights[selectedRow]) {
		level = ca.arrayWeights[selectedRow][selectedCol]
	}
	ca.mu.RUnlock()

	// Update status with single-cell sense result including energy/timing
	// Read cycle: ~76ns, Energy: DAC(14.4fJ) + TIA(6.3fJ) + ADC(25fJ) = ~46fJ
	sharedwidgets.SafeDo(func() {
		ca.operationsStatusLabel.SetText(formatReadStatusLine(selectedRow, selectedCol, level, current, tiaVoltage, adcLevel))
	})
}

// onUnifiedCompute runs MVM computation with current input vector
func (ca *CircuitsApp) onUnifiedCompute() {
	// Mode validation: only allowed in COMPUTE mode
	if ca.deviceState.GetOperationMode() != OpModeCompute {
		ca.operationsStatusLabel.SetText("Error: Switch to COMPUTE mode first")
		return
	}
	logAction("compute_mvm rows=%d cols=%d", ca.arrayRows, ca.arrayCols)

	// Ensure all rows are active for MVM
	ca.deviceState.SetWLAll()

	// Apply input vector to DAC (convert 0-255 to read range voltages)
	ca.mu.RLock()
	params := make([]float64, len(ca.inputVector))
	for i, v := range ca.inputVector {
		params[i] = float64(v)
	}
	ca.mu.RUnlock()
	ca.deviceState.SetDACPreset(DACInputVector, params...)

	ca.recomputeAndRefresh()

	// Save compute log for debugging
	// MVM: ~76ns (parallel row read), Energy: N x ~46fJ where N = active cells
	activeCells := ca.arrayRows * ca.arrayCols
	energyFJ := activeCells * 46 // ~46fJ per cell (read path)
	if ComputeLogEnabled() {
		if err := SaveComputeLog(); err != nil {
			ca.operationsStatusLabel.SetText(fmt.Sprintf("MVM done (log error: %v)", err))
			return
		}
		ca.operationsStatusLabel.SetText(fmt.Sprintf("MVM complete: %dx%d array | ~76ns, ~%dfJ total | saved log",
			ca.arrayRows, ca.arrayCols, energyFJ))
		return
	}

	ca.operationsStatusLabel.SetText(fmt.Sprintf("MVM complete: %dx%d array | ~76ns, ~%dfJ total",
		ca.arrayRows, ca.arrayCols, energyFJ))
}

// onUnifiedAnimate animates the signal flow
func (ca *CircuitsApp) onUnifiedAnimate() {
	ca.mu.Lock()
	ca.animationActive = true
	ca.mu.Unlock()

	ca.operationsStatusLabel.SetText("Animating...")

	go func() {
		// Step 1: DAC
		if ca.shouldStop() {
			return
		}
		ca.mu.Lock()
		ca.animationStep = 1
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText("Step 1: DAC conversion (10ns)")
		})
		if ca.sleep(600) {
			return // Interrupted
		}

		// Step 2: Array
		if ca.shouldStop() {
			return
		}
		ca.mu.Lock()
		ca.animationStep = 2
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText("Step 2: Array settle (5ns)")
		})
		if ca.sleep(600) {
			return // Interrupted
		}

		// Step 3: ADC
		if ca.shouldStop() {
			return
		}
		ca.mu.Lock()
		ca.animationStep = 3
		ca.mu.Unlock()
		ca.refreshUnifiedArray()
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText("Step 3: TIA+ADC conversion (~61ns)")
		})
		if ca.sleep(600) {
			return // Interrupted
		}

		// Complete
		ca.mu.Lock()
		ca.animationStep = 0
		ca.animationActive = false
		ca.mu.Unlock()
		ca.recomputeAndRefresh()
		sharedwidgets.SafeDo(func() {
			ca.operationsStatusLabel.SetText("Complete in ~76ns")
		})
	}()
}

// onUnifiedReset resets the array to mid-level (neutral state)
func (ca *CircuitsApp) onUnifiedReset() {
	// Save current state to undo history before resetting
	ca.saveUndoHistory()
	logAction("reset_array rows=%d cols=%d", ca.arrayRows, ca.arrayCols)

	// Reset all array weights to mid-level (e.g., 15 for 30 levels)
	midLevel := ca.quantLevels / 2
	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = midLevel
		}
	}
	for r := range ca.halfSelectResidue {
		for c := range ca.halfSelectResidue[r] {
			ca.halfSelectResidue[r][c] = 0
		}
	}
	ca.mu.Unlock()

	// Reset DAC to read preset (uses material-derived voltage range)
	ca.deviceState.SetDACPreset(DACReadPreset)
	ca.updateDACRangeModeLabel()

	// Reset WL based on operation mode (only in 1T1R/2T1R - passive keeps all on)
	isPassive := ca.architecture == sharedwidgets.Architecture0T1R
	if !isPassive {
		if ca.deviceState.GetOperationMode() == OpModeCompute {
			ca.deviceState.SetWLAll() // COMPUTE needs all rows for MVM
		} else {
			ca.deviceState.SetWLSingle(0)
		}
	}

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Reset to mid-level complete")
}

// onUnifiedRandomArray randomizes the array weights
func (ca *CircuitsApp) onUnifiedRandomArray() {
	// H3 FIX: Save current state to undo history before modifying
	ca.saveUndoHistory()
	logAction("randomize_array rows=%d cols=%d", ca.arrayRows, ca.arrayCols)

	ca.mu.Lock()
	for r := range ca.arrayWeights {
		for c := range ca.arrayWeights[r] {
			ca.arrayWeights[r][c] = rand.Intn(ca.quantLevels)
		}
	}
	for r := range ca.halfSelectResidue {
		for c := range ca.halfSelectResidue[r] {
			ca.halfSelectResidue[r][c] = 0
		}
	}
	ca.mu.Unlock()

	// Apply default read voltage after randomization so DAC/TIA/ADC reflect the new array state
	if ca.deviceState != nil {
		mode := ca.deviceState.GetOperationMode()
		if mode == OpModeRead {
			// Apply read voltage to selected column
			ca.deviceState.SetAllDACVoltages(0)
			ca.deviceState.SetDACVoltage(ca.deviceState.GetSelectedCol(), ca.safeReadVoltage())
		} else if mode == OpModeCompute {
			// Re-apply input vector as DAC voltages
			params := make([]float64, len(ca.inputVector))
			for i, v := range ca.inputVector {
				params[i] = float64(v)
			}
			ca.deviceState.SetDACPreset(DACInputVector, params...)
		}
	}

	ca.recomputeAndRefresh()
	ca.operationsStatusLabel.SetText("Array randomized")
}

// H3 FIX: saveUndoHistory saves the current array state for undo
func (ca *CircuitsApp) saveUndoHistory() {
	ca.mu.Lock()
	// Create a deep copy of current array state
	ca.undoHistory = make([][]int, len(ca.arrayWeights))
	for i := range ca.arrayWeights {
		ca.undoHistory[i] = make([]int, len(ca.arrayWeights[i]))
		copy(ca.undoHistory[i], ca.arrayWeights[i])
	}
	ca.hasUndoHistory = true
	ca.mu.Unlock() // Release lock before UI update to avoid deadlock

	// Enable undo button
	sharedwidgets.SafeDo(func() {
		if ca.undoHistoryBtn != nil {
			ca.undoHistoryBtn.Enable()
		}
	})
}

// H3 FIX: onUndo restores the previous array state
func (ca *CircuitsApp) onUndo() {
	ca.mu.Lock()
	if !ca.hasUndoHistory || ca.undoHistory == nil {
		ca.mu.Unlock()
		return
	}
	logAction("undo_array rows=%d cols=%d", ca.arrayRows, ca.arrayCols)

	// Restore array from history with defensive length check
	for i := range ca.arrayWeights {
		if i < len(ca.undoHistory) && len(ca.arrayWeights[i]) == len(ca.undoHistory[i]) {
			copy(ca.arrayWeights[i], ca.undoHistory[i])
		}
	}

	// Clear history (single-level undo only)
	ca.undoHistory = nil
	ca.hasUndoHistory = false
	ca.mu.Unlock() // Release lock before UI updates to avoid deadlock

	// Disable undo button
	sharedwidgets.SafeDo(func() {
		if ca.undoHistoryBtn != nil {
			ca.undoHistoryBtn.Disable()
		}
		if ca.operationsStatusLabel != nil {
			ca.operationsStatusLabel.SetText("Undo complete")
		}
	})

	ca.recomputeAndRefresh()
}
