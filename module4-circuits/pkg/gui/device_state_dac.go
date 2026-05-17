//go:build legacy_fyne

// Package gui — device_state_dac.go
// DAC/ADC voltage configuration, WL/BL mode selection, passive-mode control,
// DAC presets, output-state getters, operation-mode management, and array resize.
package gui

// GetReadRange returns the voltage range for read/compute operations
func (ds *DeviceState) GetReadRange() VoltageRange {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.readRange
}

// GetWriteRange returns the voltage range for write operations
func (ds *DeviceState) GetWriteRange() VoltageRange {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.writeRange
}

// GetDACRangeMode returns the current DAC range mode
func (ds *DeviceState) GetDACRangeMode() DACRangeMode {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.dacRangeMode
}

// SetDACRangeMode sets the DAC range mode (read vs write)
func (ds *DeviceState) SetDACRangeMode(mode DACRangeMode) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.dacRangeMode = mode
}

// GetCurrentVoltageRange returns the voltage range for the current mode
func (ds *DeviceState) GetCurrentVoltageRange() VoltageRange {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.dacRangeMode == DACRangeWrite {
		return ds.writeRange
	}
	return ds.readRange
}

// SetPassiveMode sets whether the device is in passive mode (0T1R)
// In passive mode, all WLs are ALWAYS on and cannot be changed
func (ds *DeviceState) SetPassiveMode(passive bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.isPassive = passive
	if passive {
		// Force all WLs on
		ds.wlMode = WLAll
		for i := range ds.activeRows {
			ds.activeRows[i] = true
		}
	}
}

// IsPassiveMode returns true if in passive mode (all WLs always on)
func (ds *DeviceState) IsPassiveMode() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.isPassive
}

// SetWLSingle activates only the specified row
// In passive mode, this is ignored - all WLs stay on
func (ds *DeviceState) SetWLSingle(row int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.setWLSingleLocked(row)
}

func (ds *DeviceState) setWLSingleLocked(row int) {
	if ds.isPassive {
		return // Passive mode: all WLs always on, ignore
	}
	ds.wlMode = WLSingle
	ds.selectedRow = row
	for i := range ds.activeRows {
		ds.activeRows[i] = (i == row)
	}
}

// SetWLAll activates all rows for MVM
func (ds *DeviceState) SetWLAll() {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.setWLAllLocked()
}

func (ds *DeviceState) setWLAllLocked() {
	ds.wlMode = WLAll
	for i := range ds.activeRows {
		ds.activeRows[i] = true
	}
}

// SetWLCustom sets a custom WL pattern
// In passive mode, this is ignored - all WLs stay on
func (ds *DeviceState) SetWLCustom(pattern []bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ds.isPassive {
		return // Passive mode: all WLs always on, ignore
	}
	ds.wlMode = WLCustom
	copy(ds.activeRows, pattern)
}

// SetDACVoltage sets voltage for a single column
func (ds *DeviceState) SetDACVoltage(col int, voltage float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.setDACVoltageLocked(col, voltage)
}

func (ds *DeviceState) setDACVoltageLocked(col int, voltage float64) {
	if col >= 0 && col < ds.cols {
		ds.dacVoltages[col] = voltage
		ds.dacMode = DACManual
	}
}

// SetDACPreset applies a preset pattern using material-derived voltage ranges
func (ds *DeviceState) SetDACPreset(preset DACMode, params ...float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.dacMode = preset

	switch preset {
	case DACReadPreset:
		// Use read range from material calibration
		// Only selected column gets read voltage, others are 0
		ds.dacRangeMode = DACRangeRead
		voltage := ds.readRange.Max * 0.5 // Default to 50% of safe read range
		if len(params) > 0 {
			voltage = params[0]
		}
		// Clamp to read range
		if voltage > ds.readRange.Max {
			voltage = ds.readRange.Max
		}
		for i := range ds.dacVoltages {
			if i == ds.selectedCol {
				ds.dacVoltages[i] = voltage
			} else {
				ds.dacVoltages[i] = 0
			}
		}

	case DACWritePreset:
		// Use write range from material calibration
		ds.dacRangeMode = DACRangeWrite
		// Default to a positive write pulse for selected column
		writeVoltage := ds.writeRange.Max * 0.5
		if len(params) > 0 {
			writeVoltage = params[0]
		}
		// Clamp to write range
		if writeVoltage < ds.writeRange.Min {
			writeVoltage = ds.writeRange.Min
		}
		if writeVoltage > ds.writeRange.Max {
			writeVoltage = ds.writeRange.Max
		}
		for i := range ds.dacVoltages {
			if i == ds.selectedCol {
				ds.dacVoltages[i] = writeVoltage
			} else {
				ds.dacVoltages[i] = 0
			}
		}

	case DACInputVector:
		// Convert input vector to per-column voltages for MVM.
		//
		// Physics meaning:
		//   - Each column j is driven with an analog input Vj.
		//   - Row currents follow I_i = sum_j (G_ij * Vj).
		//
		// Mapping (units):
		//   - UI supplies "byte-like" codes in the range 0..255.
		//   - We map 0 -> 0V and 255 -> readRange.Max (compute-safe full-scale).
		//
		// Bounds / clamping:
		//   - Any param below 0 is clamped to 0.
		//   - Any param above 255 is clamped to 255.
		ds.dacRangeMode = DACRangeRead
		for i := range ds.dacVoltages {
			if i >= len(params) {
				continue
			}
			code := params[i]
			if code < 0 {
				code = 0
			}
			if code > 255 {
				code = 255
			}
			normalized := code / 255.0
			ds.dacVoltages[i] = normalized * ds.readRange.Max
		}

	case DACRandom:
		// Random voltages in read range (compute-safe)
		// Note: actual random generation done by caller
		ds.dacRangeMode = DACRangeRead
	}
}

// SetDACVoltageForState sets the write voltage for a target state (0 to numLevels-1)
// Maps the state to the appropriate voltage in the write range
// numLevels specifies the quantization levels used by the app (typically 30 for FeCIM)
func (ds *DeviceState) SetDACVoltageForState(col int, targetState int, numLevels int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if col < 0 || col >= ds.cols {
		return
	}

	// Use provided numLevels, fallback to writeRange if not specified
	if numLevels <= 0 {
		numLevels = ds.writeRange.NumLevels
	}

	// Clamp target state
	if targetState < 0 {
		targetState = 0
	}
	if targetState >= numLevels {
		targetState = numLevels - 1
	}

	// Linear interpolation within write range
	// Maps level 0 -> writeRange.Min, level (numLevels-1) -> writeRange.Max
	normalized := float64(targetState) / float64(numLevels-1)
	voltage := ds.writeRange.Min + normalized*(ds.writeRange.Max-ds.writeRange.Min)

	ds.dacVoltages[col] = voltage
	ds.dacRangeMode = DACRangeWrite
	ds.dacMode = DACManual
}

// CalculateVoltageForState calculates the write voltage for a target state without setting it
// Used for UI preview - actual voltage is only applied when user presses "Program Cell"
func (ds *DeviceState) CalculateVoltageForState(targetState int, numLevels int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if numLevels <= 0 {
		numLevels = ds.writeRange.NumLevels
	}

	// Clamp target state
	if targetState < 0 {
		targetState = 0
	}
	if targetState >= numLevels {
		targetState = numLevels - 1
	}

	// Linear interpolation within write range
	normalized := float64(targetState) / float64(numLevels-1)
	return ds.writeRange.Min + normalized*(ds.writeRange.Max-ds.writeRange.Min)
}

// SetAllDACVoltages sets all DAC columns to the same voltage
func (ds *DeviceState) SetAllDACVoltages(voltage float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.setAllDACVoltagesLocked(voltage)
}

func (ds *DeviceState) setAllDACVoltagesLocked(voltage float64) {
	ds.dacMode = DACManual
	for i := range ds.dacVoltages {
		ds.dacVoltages[i] = voltage
	}
}

// SetSelectedCell sets the currently selected cell
func (ds *DeviceState) SetSelectedCell(row, col int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.selectedRow = row
	ds.selectedCol = col
	if ds.wlMode == WLSingle {
		ds.setWLSingleLocked(row)
	}
}

// GetRowCurrent returns the computed current for a row
func (ds *DeviceState) GetRowCurrent(row int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row >= 0 && row < ds.rows {
		return ds.rowCurrents[row]
	}
	return 0
}

// GetRowVoltage returns the TIA output voltage for a row
func (ds *DeviceState) GetRowVoltage(row int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row >= 0 && row < ds.rows {
		return ds.rowVoltages[row]
	}
	return 0
}

// GetRowLevel returns the ADC output level for a row
func (ds *DeviceState) GetRowLevel(row int) int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row >= 0 && row < ds.rows {
		return ds.rowLevels[row]
	}
	return 0
}

// IsSaturated returns whether a row's output is saturated
func (ds *DeviceState) IsSaturated(row int) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row >= 0 && row < ds.rows {
		return ds.saturated[row]
	}
	return false
}

// IsRowActive returns whether a row's WL is active
func (ds *DeviceState) IsRowActive(row int) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row >= 0 && row < ds.rows {
		return ds.activeRows[row]
	}
	return false
}

// GetDACVoltage returns the DAC voltage for a column
func (ds *DeviceState) GetDACVoltage(col int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if col >= 0 && col < ds.cols {
		return ds.dacVoltages[col]
	}
	return 0
}

// GetWLMode returns the current WL selection mode
func (ds *DeviceState) GetWLMode() WLMode {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.wlMode
}

// GetDACMode returns the current DAC preset mode
func (ds *DeviceState) GetDACMode() DACMode {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.dacMode
}

// GetSelectedRow returns the selected row index
func (ds *DeviceState) GetSelectedRow() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.selectedRow
}

// GetSelectedCol returns the selected column index
func (ds *DeviceState) GetSelectedCol() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.selectedCol
}

// GetOperationMode returns the current operation mode (READ/WRITE/COMPUTE)
func (ds *DeviceState) GetOperationMode() OpMode {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.opMode
}

// SetOperationMode sets the operation mode
// This is called by the UI; actual WL/DAC configuration is done in tab_unified.go
func (ds *DeviceState) SetOperationMode(mode OpMode) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.opMode = mode
}

// ClassifyOperation returns a string describing the current operation mode
func (ds *DeviceState) ClassifyOperation() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	switch ds.opMode {
	case OpModeRead:
		return "READ"
	case OpModeWrite:
		return "WRITE"
	case OpModeCompute:
		return "COMPUTE (MVM)"
	default:
		return "IDLE"
	}
}

// Resize updates the device state dimensions
func (ds *DeviceState) Resize(rows, cols int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if rows != ds.rows {
		ds.rows = rows
		ds.activeRows = make([]bool, rows)
		ds.wlVoltages = make([]float64, rows) // V/2 scheme support
		ds.rowCurrents = make([]float64, rows)
		ds.rowVoltages = make([]float64, rows)
		ds.rowLevels = make([]int, rows)
		ds.saturated = make([]bool, rows)
		ds.coupledCellVoltages = nil
		ds.coupledCellCurrents = nil
		// Reset to single row 0
		if rows > 0 {
			ds.activeRows[0] = true
		}
	}

	if cols != ds.cols {
		ds.cols = cols
		ds.dacVoltages = make([]float64, cols)
		ds.coupledCellVoltages = nil
		ds.coupledCellCurrents = nil
		// Reset to read preset (use material-derived safe read voltage)
		readVoltage := ds.readRange.Max * 0.5 // 50% of max safe read voltage
		for i := range ds.dacVoltages {
			ds.dacVoltages[i] = readVoltage
		}
	}

	// Ensure selected cell is within bounds
	if ds.selectedRow >= ds.rows {
		ds.selectedRow = 0
	}
	if ds.selectedCol >= ds.cols {
		ds.selectedCol = 0
	}
}
