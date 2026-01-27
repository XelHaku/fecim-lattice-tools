// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device state for the simulation view.
package gui

import (
	"fecim-lattice-tools/module4-circuits/pkg/peripherals"
)

// OperationMode represents the current operation mode (legacy, kept for compatibility)
type OperationMode int

const (
	ModeWrite OperationMode = iota
	ModeRead
	ModeCompute
)

// WLMode represents word line selection mode
type WLMode int

const (
	WLSingle WLMode = iota // One row selected (for program/read single cell)
	WLAll                  // All rows active (for MVM compute)
	WLCustom               // User-defined pattern
)

// DACMode represents how DAC voltages were set
type DACMode int

const (
	DACManual DACMode = iota // User entered each voltage
	DACReadPreset            // All columns at readVoltage (0.5V)
	DACWritePreset           // Selected column at write voltage
	DACInputVector           // From digital input vector (0-255 -> 0-1V)
	DACRandom                // Random voltages
)

// Voltage thresholds for operation classification
const (
	VoltageThresholdRead  = 0.5  // Max voltage for safe read
	VoltageThresholdWrite = 1.2  // Min voltage for write operation
	VoltageMaxWrite       = 1.5  // Max write voltage
	VoltageMaxCompute     = 1.0  // Max voltage for compute (MVM)
)

// DeviceState holds the unified simulation state
type DeviceState struct {
	// Dimensions
	rows int
	cols int

	// WL configuration
	wlMode     WLMode
	activeRows []bool // true = WL HIGH for that row

	// DAC inputs (per column)
	dacVoltages []float64
	dacMode     DACMode

	// Computed outputs (per row)
	rowCurrents []float64 // TIA input currents (uA)
	rowVoltages []float64 // TIA output voltages (V)
	rowLevels   []int     // ADC output levels

	// Saturation flags
	saturated []bool

	// Selected cell (for single-cell operations)
	selectedRow int
	selectedCol int

	// Peripherals reference
	tia *peripherals.TIA
	adc *peripherals.ADC
}

// NewDeviceState creates a new device state with specified dimensions
func NewDeviceState(rows, cols int, tia *peripherals.TIA, adc *peripherals.ADC) *DeviceState {
	ds := &DeviceState{
		rows:        rows,
		cols:        cols,
		wlMode:      WLSingle,
		activeRows:  make([]bool, rows),
		dacVoltages: make([]float64, cols),
		dacMode:     DACReadPreset,
		rowCurrents: make([]float64, rows),
		rowVoltages: make([]float64, rows),
		rowLevels:   make([]int, rows),
		saturated:   make([]bool, rows),
		selectedRow: 0,
		selectedCol: 0,
		tia:         tia,
		adc:         adc,
	}

	// Initialize with read preset (0.5V all columns)
	ds.SetDACPreset(DACReadPreset, VoltageThresholdRead)

	// Default: single row 0 active
	ds.activeRows[0] = true

	return ds
}

// SetWLSingle activates only the specified row
func (ds *DeviceState) SetWLSingle(row int) {
	ds.wlMode = WLSingle
	ds.selectedRow = row
	for i := range ds.activeRows {
		ds.activeRows[i] = (i == row)
	}
}

// SetWLAll activates all rows for MVM
func (ds *DeviceState) SetWLAll() {
	ds.wlMode = WLAll
	for i := range ds.activeRows {
		ds.activeRows[i] = true
	}
}

// SetWLCustom sets a custom WL pattern
func (ds *DeviceState) SetWLCustom(pattern []bool) {
	ds.wlMode = WLCustom
	copy(ds.activeRows, pattern)
}

// SetDACVoltage sets voltage for a single column
func (ds *DeviceState) SetDACVoltage(col int, voltage float64) {
	if col >= 0 && col < ds.cols {
		ds.dacVoltages[col] = voltage
		ds.dacMode = DACManual
	}
}

// SetDACPreset applies a preset pattern
func (ds *DeviceState) SetDACPreset(preset DACMode, params ...float64) {
	ds.dacMode = preset

	switch preset {
	case DACReadPreset:
		voltage := VoltageThresholdRead
		if len(params) > 0 {
			voltage = params[0]
		}
		for i := range ds.dacVoltages {
			ds.dacVoltages[i] = voltage
		}

	case DACWritePreset:
		// Set selected column to write voltage, others to 0
		writeVoltage := VoltageMaxWrite
		if len(params) > 0 {
			writeVoltage = params[0]
		}
		for i := range ds.dacVoltages {
			if i == ds.selectedCol {
				ds.dacVoltages[i] = writeVoltage
			} else {
				ds.dacVoltages[i] = 0
			}
		}

	case DACInputVector:
		// Convert input vector (0-255) to voltage (0-1V)
		// Params should be the input values as float64
		for i := range ds.dacVoltages {
			if i < len(params) {
				ds.dacVoltages[i] = params[i] / 255.0
			}
		}

	case DACRandom:
		// Random voltages between 0 and 1V (compute-safe)
		// Note: actual random generation done by caller
		// This just marks the mode
	}
}

// SetAllDACVoltages sets all DAC columns to the same voltage
func (ds *DeviceState) SetAllDACVoltages(voltage float64) {
	ds.dacMode = DACManual
	for i := range ds.dacVoltages {
		ds.dacVoltages[i] = voltage
	}
}

// SetSelectedCell sets the currently selected cell
func (ds *DeviceState) SetSelectedCell(row, col int) {
	ds.selectedRow = row
	ds.selectedCol = col
	if ds.wlMode == WLSingle {
		ds.SetWLSingle(row)
	}
}

// Compute runs the device simulation given the weight matrix
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
	for r := 0; r < ds.rows; r++ {
		if !ds.activeRows[r] {
			ds.rowCurrents[r] = 0
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		// Sum currents from all active columns
		totalCurrent := 0.0
		for c := 0; c < ds.cols; c++ {
			voltage := ds.dacVoltages[c]
			if voltage < 0.01 {
				continue
			}

			// Get cell conductance from weight
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			// Conductance: 1-100 uS based on level
			conductance := 1.0 + float64(level)/float64(quantLevels-1)*99.0
			current := conductance * voltage // I = G * V (in uA since G is in uS)
			totalCurrent += current
		}

		ds.rowCurrents[r] = totalCurrent

		// TIA conversion: current (A) to voltage (V)
		if ds.tia != nil {
			ds.rowVoltages[r] = ds.tia.Convert(totalCurrent * 1e-6) // uA to A
		}

		// ADC conversion: voltage to level
		if ds.adc != nil {
			ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r])
		}

		// Check saturation (TIA saturates around 100 uA)
		ds.saturated[r] = totalCurrent > 100.0 || ds.rowLevels[r] >= 31
	}
}

// GetRowCurrent returns the computed current for a row
func (ds *DeviceState) GetRowCurrent(row int) float64 {
	if row >= 0 && row < ds.rows {
		return ds.rowCurrents[row]
	}
	return 0
}

// GetRowVoltage returns the TIA output voltage for a row
func (ds *DeviceState) GetRowVoltage(row int) float64 {
	if row >= 0 && row < ds.rows {
		return ds.rowVoltages[row]
	}
	return 0
}

// GetRowLevel returns the ADC output level for a row
func (ds *DeviceState) GetRowLevel(row int) int {
	if row >= 0 && row < ds.rows {
		return ds.rowLevels[row]
	}
	return 0
}

// IsSaturated returns whether a row's output is saturated
func (ds *DeviceState) IsSaturated(row int) bool {
	if row >= 0 && row < ds.rows {
		return ds.saturated[row]
	}
	return false
}

// IsRowActive returns whether a row's WL is active
func (ds *DeviceState) IsRowActive(row int) bool {
	if row >= 0 && row < ds.rows {
		return ds.activeRows[row]
	}
	return false
}

// GetDACVoltage returns the DAC voltage for a column
func (ds *DeviceState) GetDACVoltage(col int) float64 {
	if col >= 0 && col < ds.cols {
		return ds.dacVoltages[col]
	}
	return 0
}

// GetWLMode returns the current WL selection mode
func (ds *DeviceState) GetWLMode() WLMode {
	return ds.wlMode
}

// GetDACMode returns the current DAC preset mode
func (ds *DeviceState) GetDACMode() DACMode {
	return ds.dacMode
}

// GetSelectedRow returns the selected row index
func (ds *DeviceState) GetSelectedRow() int {
	return ds.selectedRow
}

// GetSelectedCol returns the selected column index
func (ds *DeviceState) GetSelectedCol() int {
	return ds.selectedCol
}

// ClassifyOperation determines what operation the current configuration represents
func (ds *DeviceState) ClassifyOperation() string {
	// Check if any column has write voltage
	hasWriteVoltage := false
	hasReadVoltage := false
	for _, v := range ds.dacVoltages {
		if v >= VoltageThresholdWrite {
			hasWriteVoltage = true
		}
		if v > 0.01 && v <= VoltageThresholdRead {
			hasReadVoltage = true
		}
	}

	activeRowCount := 0
	for _, active := range ds.activeRows {
		if active {
			activeRowCount++
		}
	}

	// Classify based on WL mode and voltage levels
	switch {
	case activeRowCount == 1 && hasWriteVoltage:
		return "PROGRAM"
	case activeRowCount == 1 && hasReadVoltage:
		return "READ"
	case activeRowCount > 1 && !hasWriteVoltage:
		return "COMPUTE (MVM)"
	case activeRowCount > 1 && hasWriteVoltage:
		return "BULK PROGRAM (CAUTION)"
	default:
		return "IDLE"
	}
}

// Resize updates the device state dimensions
func (ds *DeviceState) Resize(rows, cols int) {
	if rows != ds.rows {
		ds.rows = rows
		ds.activeRows = make([]bool, rows)
		ds.rowCurrents = make([]float64, rows)
		ds.rowVoltages = make([]float64, rows)
		ds.rowLevels = make([]int, rows)
		ds.saturated = make([]bool, rows)
		// Reset to single row 0
		if rows > 0 {
			ds.activeRows[0] = true
		}
	}

	if cols != ds.cols {
		ds.cols = cols
		ds.dacVoltages = make([]float64, cols)
		// Reset to read preset
		for i := range ds.dacVoltages {
			ds.dacVoltages[i] = VoltageThresholdRead
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
