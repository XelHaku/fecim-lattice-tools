// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device state for the simulation view.
package gui

import (
	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"
	"fecim-lattice-tools/module4-circuits/pkg/peripherals"
)

// OperationMode represents the current operation mode (legacy, kept for compatibility)
type OperationMode int

const (
	ModeWrite OperationMode = iota
	ModeRead
	ModeCompute
)

// OpMode represents the unified operation mode for the device simulation
// This controls both WL selection and DAC voltage range automatically
type OpMode int

const (
	OpModeRead    OpMode = iota // READ: Single row active, safe voltage (0-0.5V)
	OpModeWrite                 // WRITE: Single row active, write voltage (Vc to 1.3*Vc)
	OpModeCompute               // COMPUTE: All rows active, input vector (0-1V)
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
	DACReadPreset            // Selected column at readVoltage, others 0 (single cell read)
	DACWritePreset           // Selected column at write voltage, others 0 (single cell write)
	DACInputVector           // From digital input vector (0-255 -> 0-1V)
	DACRandom                // Random voltages
)

// DACRangeMode represents the DAC output range mode
type DACRangeMode int

const (
	DACRangeRead  DACRangeMode = iota // 0 to 1V (read/compute safe zone)
	DACRangeWrite                     // MinWriteV to MaxWriteV (write zone)
)

// VoltageRange holds the min/max voltages for a given operation
// All values are derived from material properties (Ec, thickness) via physics.yaml config
type VoltageRange struct {
	Min       float64 // Minimum voltage (derived from material)
	Max       float64 // Maximum voltage (derived from material)
	StepSize  float64 // Voltage step between states = (Max-Min)/(NumLevels-1)
	NumLevels int     // Number of discrete levels (from material.NumLevels)
}

// CalibrationParams holds voltage calculation parameters from physics.yaml
// These define the operating voltage regions relative to coercive voltage (Vc)
type CalibrationParams struct {
	FieldMinRatio float64 // Read max = FieldMinRatio * Vc (from calibration.field_min_ratio)
	FieldMaxRatio float64 // Write max = FieldMaxRatio * Vc (from calibration.field_max_ratio)
}

// loadCalibrationParams loads calibration ratios from physics.yaml config
// Falls back to sensible defaults if config is unavailable
func loadCalibrationParams() CalibrationParams {
	cfg, err := physics.Load()
	if err != nil || cfg == nil {
		// Fallback: field_min_ratio=0.5, field_max_ratio=2.5 from typical physics.yaml
		return CalibrationParams{
			FieldMinRatio: 0.5,
			FieldMaxRatio: 2.5,
		}
	}
	return CalibrationParams{
		FieldMinRatio: cfg.Calibration.FieldMinRatio,
		FieldMaxRatio: cfg.Calibration.FieldMaxRatio,
	}
}

// MaxPracticalVoltage: Hardware DAC/driver limit (prevents unrealistic voltages)
const MaxPracticalVoltage = 3.0

// DeviceState holds the unified simulation state
type DeviceState struct {
	// Dimensions
	rows int
	cols int

	// Passive mode flag - when true, ALL WLs are always on (0T1R architecture)
	isPassive bool

	// Operation mode (READ/WRITE/COMPUTE)
	opMode OpMode // Current operation mode

	// WL configuration (derived from opMode)
	wlMode     WLMode
	activeRows []bool // true = WL HIGH for that row

	// DAC inputs (per column)
	dacVoltages  []float64
	dacMode      DACMode
	dacRangeMode DACRangeMode // Current DAC range (read vs write)

	// Voltage ranges (derived from material + calibration config)
	readRange   VoltageRange     // 0 to FieldMinRatio*Vc for read/compute
	writeRange  VoltageRange     // Vc to FieldMaxRatio*Vc for write operations
	calibParams CalibrationParams // Loaded from physics.yaml

	// Computed outputs (per row)
	rowCurrents []float64 // TIA input currents (uA)
	rowVoltages []float64 // TIA output voltages (V)
	rowLevels   []int     // ADC output levels

	// Saturation flags
	saturated []bool

	// Selected cell (for single-cell operations)
	selectedRow int
	selectedCol int

	// Material physics model (from hysteresis calibration)
	material *ferroelectric.HZOMaterial

	// Peripherals reference
	tia *peripherals.TIA
	adc *peripherals.ADC
}

// NewDeviceState creates a new device state with specified dimensions
// Loads calibration parameters from physics.yaml for voltage range calculation
func NewDeviceState(rows, cols int, tia *peripherals.TIA, adc *peripherals.ADC) *DeviceState {
	ds := &DeviceState{
		rows:         rows,
		cols:         cols,
		opMode:       OpModeRead, // Default to READ mode
		wlMode:       WLSingle,
		activeRows:   make([]bool, rows),
		dacVoltages:  make([]float64, cols),
		dacMode:      DACReadPreset,
		dacRangeMode: DACRangeRead,
		rowCurrents:  make([]float64, rows),
		rowVoltages:  make([]float64, rows),
		rowLevels:    make([]int, rows),
		saturated:    make([]bool, rows),
		selectedRow:  0,
		selectedCol:  0,
		material:     ferroelectric.FeCIMMaterial(), // Default to FeCIM material
		calibParams:  loadCalibrationParams(),       // Load from physics.yaml
		tia:          tia,
		adc:          adc,
	}

	// Calculate voltage ranges from material + calibration config
	ds.updateVoltageRanges()

	// Initialize with read preset (uses read range)
	ds.SetDACRangeMode(DACRangeRead)
	ds.SetDACPreset(DACReadPreset)

	// Default: single row 0 active
	ds.activeRows[0] = true

	return ds
}

// updateVoltageRanges calculates voltage ranges from material properties and calibration config
// Read range: 0 to FieldMinRatio * Vc (below coercive voltage, non-destructive sensing)
// Write range: Vc to FieldMaxRatio * Vc (exceeds coercive voltage for polarization switching)
//
// From physics.yaml calibration section:
//   field_min_ratio: 0.5  -> Read max = 0.5 * Vc
//   field_max_ratio: 2.5  -> Write max = 2.5 * Vc
func (ds *DeviceState) updateVoltageRanges() {
	// Get material's coercive voltage (Vc = Ec * thickness)
	Vc := 1.0 // Fallback if no material
	numLevels := 30
	if ds.material != nil {
		Vc = ds.material.CoerciveVoltage()
		numLevels = ds.material.GetNumLevels()
	}

	// Read range: 0 to FieldMinRatio * Vc
	// This is the safe sensing zone below coercive voltage
	safeReadMax := ds.calibParams.FieldMinRatio * Vc
	if safeReadMax > 1.0 {
		safeReadMax = 1.0 // Cap at 1V for practical DAC range
	}
	if safeReadMax < 0.1 {
		safeReadMax = 0.1 // Minimum useful read voltage
	}

	ds.readRange = VoltageRange{
		Min:       0,
		Max:       safeReadMax,
		StepSize:  safeReadMax / float64(numLevels-1),
		NumLevels: numLevels,
	}

	// Write range: Vc to FieldMaxRatio * Vc
	// Must exceed Vc to switch polarization
	writeMin := Vc
	writeMax := ds.calibParams.FieldMaxRatio * Vc
	if writeMax > MaxPracticalVoltage {
		writeMax = MaxPracticalVoltage
	}

	ds.writeRange = VoltageRange{
		Min:       writeMin,
		Max:       writeMax,
		StepSize:  (writeMax - writeMin) / float64(numLevels-1),
		NumLevels: numLevels,
	}
}

// SetMaterial changes the ferroelectric material used for conductance calculation
func (ds *DeviceState) SetMaterial(mat *ferroelectric.HZOMaterial) {
	ds.material = mat
	ds.updateVoltageRanges() // Recalculate voltage ranges for new material
}

// GetMaterial returns the current material
func (ds *DeviceState) GetMaterial() *ferroelectric.HZOMaterial {
	return ds.material
}

// GetMaterialName returns the name of the current material
func (ds *DeviceState) GetMaterialName() string {
	if ds.material != nil {
		return ds.material.Name
	}
	return "Unknown"
}

// GetReadRange returns the voltage range for read/compute operations
func (ds *DeviceState) GetReadRange() VoltageRange {
	return ds.readRange
}

// GetWriteRange returns the voltage range for write operations
func (ds *DeviceState) GetWriteRange() VoltageRange {
	return ds.writeRange
}

// GetDACRangeMode returns the current DAC range mode
func (ds *DeviceState) GetDACRangeMode() DACRangeMode {
	return ds.dacRangeMode
}

// SetDACRangeMode sets the DAC range mode (read vs write)
func (ds *DeviceState) SetDACRangeMode(mode DACRangeMode) {
	ds.dacRangeMode = mode
}

// GetCurrentVoltageRange returns the voltage range for the current mode
func (ds *DeviceState) GetCurrentVoltageRange() VoltageRange {
	if ds.dacRangeMode == DACRangeWrite {
		return ds.writeRange
	}
	return ds.readRange
}

// SetPassiveMode sets whether the device is in passive mode (0T1R)
// In passive mode, all WLs are ALWAYS on and cannot be changed
func (ds *DeviceState) SetPassiveMode(passive bool) {
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
	return ds.isPassive
}

// SetWLSingle activates only the specified row
// In passive mode, this is ignored - all WLs stay on
func (ds *DeviceState) SetWLSingle(row int) {
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
	ds.wlMode = WLAll
	for i := range ds.activeRows {
		ds.activeRows[i] = true
	}
}

// SetWLCustom sets a custom WL pattern
// In passive mode, this is ignored - all WLs stay on
func (ds *DeviceState) SetWLCustom(pattern []bool) {
	if ds.isPassive {
		return // Passive mode: all WLs always on, ignore
	}
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

// SetDACPreset applies a preset pattern using material-derived voltage ranges
func (ds *DeviceState) SetDACPreset(preset DACMode, params ...float64) {
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
		// Default to middle of write range for selected column
		writeVoltage := (ds.writeRange.Min + ds.writeRange.Max) / 2
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
		// Convert input vector (0-255) to voltage using read range
		// Maps 0-255 to readRange.Min-readRange.Max
		ds.dacRangeMode = DACRangeRead
		for i := range ds.dacVoltages {
			if i < len(params) {
				normalized := params[i] / 255.0
				ds.dacVoltages[i] = ds.readRange.Min + normalized*(ds.readRange.Max-ds.readRange.Min)
			}
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

			// Get cell conductance from weight using material physics model
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}

			// Use material's DiscreteLevel for physics-accurate conductance
			// DiscreteLevel returns conductance in Siemens (S)
			var conductanceS float64
			if ds.material != nil {
				conductanceS = ds.material.DiscreteLevel(level, quantLevels)
			} else {
				// Fallback: linear mapping 1-100 µS
				conductanceS = (1.0 + float64(level)/float64(quantLevels-1)*99.0) * 1e-6
			}

			// Convert to µS for current calculation
			conductanceUS := conductanceS * 1e6
			current := conductanceUS * voltage // I = G * V (in µA since G is in µS)
			totalCurrent += current
		}

		ds.rowCurrents[r] = totalCurrent

		// TIA conversion: current (A) to voltage (V)
		if ds.tia != nil {
			ds.rowVoltages[r] = ds.tia.Convert(totalCurrent * 1e-6) // µA to A
		}

		// ADC conversion: voltage to level
		if ds.adc != nil {
			ds.rowLevels[r] = ds.adc.Convert(ds.rowVoltages[r])
		}

		// Check saturation (TIA saturates around 100 µA)
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

// GetOperationMode returns the current operation mode (READ/WRITE/COMPUTE)
func (ds *DeviceState) GetOperationMode() OpMode {
	return ds.opMode
}

// SetOperationMode sets the operation mode
// This is called by the UI; actual WL/DAC configuration is done in tab_unified.go
func (ds *DeviceState) SetOperationMode(mode OpMode) {
	ds.opMode = mode
}

// ClassifyOperation returns a string describing the current operation mode
func (ds *DeviceState) ClassifyOperation() string {
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
