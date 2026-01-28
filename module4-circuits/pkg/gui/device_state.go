// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device state for the simulation view.
package gui

import (
	"fmt"
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
	activeRows []bool   // true = WL HIGH for that row
	wlVoltages []float64 // WL voltages for V/2 scheme (passive mode write)

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
		wlVoltages:   make([]float64, rows), // WL voltages for V/2 scheme
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
	// Ensure material is set - use default FeCIM if not
	if ds.material == nil {
		ds.material = ferroelectric.FeCIMMaterial()
	}

	// Get material's coercive voltage (Vc = Ec * thickness)
	// All values derived from material properties - no hardcoded fallbacks
	Vc := ds.material.CoerciveVoltage()
	numLevels := ds.material.GetNumLevels()

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

// CalculateVoltageForState calculates the write voltage for a target state without setting it
// Used for UI preview - actual voltage is only applied when user presses "Write Cell"
func (ds *DeviceState) CalculateVoltageForState(targetState int, numLevels int) float64 {
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
	ds.dacMode = DACManual
	for i := range ds.dacVoltages {
		ds.dacVoltages[i] = voltage
	}
}

// ============================================================================
// V/2 HALF-SELECT SCHEME (Passive 0T1R Mode Only)
// ============================================================================
//
// Per VOLTAGE_RULES.md Section 3.2 and 6.1:
// For passive (0T1R) WRITE operations, use V/2 half-select biasing to minimize
// write disturb on half-selected cells (same row or column as target).
//
// Target cell (SET operation):
//   - Selected WL: +V_write/2 (positive half)
//   - Selected BL: -V_write/2 (negative half)
//   - Effective ΔV = WL - BL = +V/2 - (-V/2) = V_write (full switching)
//
// Half-selected cells (same row or same column):
//   - Same row: WL = +V/2, BL = 0 → ΔV = +V/2 (below Vc, minimal disturb)
//   - Same col: WL = 0, BL = -V/2 → ΔV = +V/2 (below Vc, minimal disturb)
//
// Unselected cells (diagonal):
//   - WL = 0, BL = 0 → ΔV = 0 (no disturb)
//
// For 1T1R/2T1R modes, transistor isolation eliminates need for V/2.
// ============================================================================

// ApplyHalfSelectWrite applies V/2 biasing for passive (0T1R) write operation
// writeVoltage is the full write voltage (derived from material's Vc)
// For SET: WL = +V/2, BL = -V/2, giving target cell ΔV = +V_write
// For ERASE: WL = -V/2, BL = +V/2, giving target cell ΔV = -V_write
func (ds *DeviceState) ApplyHalfSelectWrite(targetRow, targetCol int, writeVoltage float64) {
	if !ds.isPassive {
		// Non-passive modes use transistor isolation, not V/2
		// Apply full voltage to selected BL only, WL controls transistor gate
		ds.SetDACVoltage(targetCol, writeVoltage)
		return
	}

	// V/2 half-select for passive mode
	halfV := writeVoltage / 2.0

	// Set WL voltages: selected row gets +V/2 (for SET), others get 0
	for i := range ds.wlVoltages {
		if i == targetRow {
			ds.wlVoltages[i] = halfV // +V/2 for selected WL
		} else {
			ds.wlVoltages[i] = 0 // Unselected WLs grounded
		}
	}

	// Set BL (DAC) voltages: selected column gets -V/2 (for SET), others get 0
	for i := range ds.dacVoltages {
		if i == targetCol {
			ds.dacVoltages[i] = -halfV // -V/2 for selected BL
		} else {
			ds.dacVoltages[i] = 0 // Unselected BLs grounded
		}
	}

	ds.dacMode = DACManual
	ds.dacRangeMode = DACRangeWrite
}

// ResetWriteVoltages returns all WL and BL voltages to 0V after write operation
// Should be called after write completes to put array in safe idle state
func (ds *DeviceState) ResetWriteVoltages() {
	// Reset all WL voltages to 0
	for i := range ds.wlVoltages {
		ds.wlVoltages[i] = 0
	}
	// Reset all DAC (BL) voltages to 0
	for i := range ds.dacVoltages {
		ds.dacVoltages[i] = 0
	}
	ds.dacMode = DACManual
}

// GetWLVoltage returns the WL voltage for a specific row
func (ds *DeviceState) GetWLVoltage(row int) float64 {
	if row >= 0 && row < len(ds.wlVoltages) {
		return ds.wlVoltages[row]
	}
	return 0
}

// GetHalfSelectVoltage returns V/2 value derived from material's write voltage
// This is the voltage seen by half-selected cells (below Vc, minimal disturb)
func (ds *DeviceState) GetHalfSelectVoltage() float64 {
	// Use middle of write range as reference
	fullWriteV := (ds.writeRange.Min + ds.writeRange.Max) / 2
	return fullWriteV / 2.0
}

// IsUsingHalfSelect returns true if V/2 scheme is active (passive mode write)
func (ds *DeviceState) IsUsingHalfSelect() bool {
	return ds.isPassive && ds.opMode == OpModeWrite
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
		// ADC saturation: 5-bit ADC has 32 levels (0-31), level 31 indicates max/saturated
		adcMaxLevel := 31 // Default 5-bit ADC max level
		if ds.adc != nil {
			adcMaxLevel = (1 << ds.adc.Bits) - 1 // 2^bits - 1
		}
		ds.saturated[r] = totalCurrent > 100.0 || ds.rowLevels[r] >= adcMaxLevel
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
		ds.wlVoltages = make([]float64, rows) // V/2 scheme support
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

// ============================================================================
// 1. PER-LEVEL VOLTAGE CALIBRATION
// ============================================================================

// PerLevelVoltageCalibration holds calibrated voltages for each of 30 levels
// Linear interpolation used as simplified demo (not physics-accurate Preisach model)
type PerLevelVoltageCalibration struct {
	AscendingVoltages  [30]float64 // Voltages for writing up (level 0→29)
	DescendingVoltages [30]float64 // Voltages for writing down (level 29→0)
}

// voltageCalibration holds the per-level calibration data
// Initialized once at startup via InitVoltageCalibration()
var voltageCalibration *PerLevelVoltageCalibration

// InitVoltageCalibration initializes the per-level voltage arrays using linear interpolation
// This is a simplified demo - real devices would use non-linear Preisach-derived values
func (ds *DeviceState) InitVoltageCalibration() {
	cal := &PerLevelVoltageCalibration{}

	// Linear interpolation: Level 0 → WriteRange.Min, Level 29 → WriteRange.Max
	minV := ds.writeRange.Min
	maxV := ds.writeRange.Max
	step := (maxV - minV) / 29.0

	for i := 0; i < 30; i++ {
		voltage := minV + float64(i)*step
		cal.AscendingVoltages[i] = voltage
		cal.DescendingVoltages[i] = voltage // Same for simplified model
	}

	voltageCalibration = cal
}

// GetVoltageForLevel returns the calibrated write voltage for a target level
// direction: true = ascending (increasing level), false = descending (decreasing level)
func (ds *DeviceState) GetVoltageForLevel(level int, ascending bool) float64 {
	if voltageCalibration == nil {
		ds.InitVoltageCalibration()
	}

	// Clamp level to valid range
	if level < 0 {
		level = 0
	}
	if level > 29 {
		level = 29
	}

	if ascending {
		return voltageCalibration.AscendingVoltages[level]
	}
	return voltageCalibration.DescendingVoltages[level]
}

// ============================================================================
// 2. HYSTERESIS DIRECTION TRACKING
// ============================================================================

// HysteresisDirection indicates the write direction on the hysteresis curve
type HysteresisDirection int

const (
	DirectionUnknown    HysteresisDirection = iota
	DirectionAscending                      // Writing to higher level
	DirectionDescending                     // Writing to lower level
)

// HysteresisState tracks the last written level and direction per cell
type HysteresisState struct {
	LastLevel map[string]int                // key: "row,col" -> last written level
	Direction map[string]HysteresisDirection // key: "row,col" -> last direction
}

// hysteresisState is the global hysteresis tracking state
var hysteresisState *HysteresisState

// getHysteresisState returns the global hysteresis state, initializing if needed
func getHysteresisState() *HysteresisState {
	if hysteresisState == nil {
		hysteresisState = &HysteresisState{
			LastLevel: make(map[string]int),
			Direction: make(map[string]HysteresisDirection),
		}
	}
	return hysteresisState
}

// cellKey generates a map key for a cell coordinate
func cellKey(row, col int) string {
	return fmt.Sprintf("%d,%d", row, col)
}

// RecordWrite updates the hysteresis state after a successful write
func (ds *DeviceState) RecordWrite(row, col, newLevel int) {
	hs := getHysteresisState()
	key := cellKey(row, col)

	oldLevel, exists := hs.LastLevel[key]
	if exists {
		if newLevel > oldLevel {
			hs.Direction[key] = DirectionAscending
		} else if newLevel < oldLevel {
			hs.Direction[key] = DirectionDescending
		}
		// If equal, keep previous direction
	} else {
		hs.Direction[key] = DirectionUnknown
	}

	hs.LastLevel[key] = newLevel
}

// GetWriteDirection determines the write direction for a target level
func (ds *DeviceState) GetWriteDirection(row, col, currentLevel, targetLevel int) HysteresisDirection {
	if targetLevel > currentLevel {
		return DirectionAscending
	} else if targetLevel < currentLevel {
		return DirectionDescending
	}
	return DirectionUnknown
}

// GetLastHysteresisDirection returns the last write direction for a cell
func (ds *DeviceState) GetLastHysteresisDirection(row, col int) HysteresisDirection {
	hs := getHysteresisState()
	key := cellKey(row, col)
	if dir, exists := hs.Direction[key]; exists {
		return dir
	}
	return DirectionUnknown
}

// ============================================================================
// 3. 4-PHASE WRITE SEQUENCE STATE MACHINE
// ============================================================================

// WritePhase represents the current phase in a 4-phase write sequence
type WritePhase int

const (
	PhaseIdle   WritePhase = iota // No write in progress
	PhaseReset                    // Applying -V_sat (100ns)
	PhaseHold1                    // Zero field hold (50ns)
	PhaseWrite                    // Applying calibrated voltage (200ns)
	PhaseHold2                    // Zero field hold (50ns)
)

// Phase timing constants (in nanoseconds for display, not real-time)
const (
	PhaseResetDurationNs = 100
	PhaseHold1DurationNs = 50
	PhaseWriteDurationNs = 200
	PhaseHold2DurationNs = 50
)

// WriteSequenceState holds the state of an active 4-phase write
type WriteSequenceState struct {
	Active       bool
	Phase        WritePhase
	TargetRow    int
	TargetCol    int
	TargetLevel  int
	CurrentLevel int
	WriteVoltage float64 // Calibrated voltage for target level
	Progress     float64 // 0.0 to 1.0 progress through sequence
}

// writeSequenceState is the global write sequence state
var writeSequenceState *WriteSequenceState

// getWriteSequenceState returns the global write sequence state
func getWriteSequenceState() *WriteSequenceState {
	if writeSequenceState == nil {
		writeSequenceState = &WriteSequenceState{}
	}
	return writeSequenceState
}

// StartWriteSequence begins a 4-phase write sequence
func (ds *DeviceState) StartWriteSequence(row, col, targetLevel int) {
	ws := getWriteSequenceState()

	direction := ds.GetWriteDirection(row, col, 0, targetLevel) // Assume starting from current
	ascending := direction == DirectionAscending

	ws.Active = true
	ws.Phase = PhaseReset
	ws.TargetRow = row
	ws.TargetCol = col
	ws.TargetLevel = targetLevel
	ws.WriteVoltage = ds.GetVoltageForLevel(targetLevel, ascending)
	ws.Progress = 0.0
}

// AdvanceWritePhase moves to the next phase in the sequence
// Returns true if sequence is complete
func (ds *DeviceState) AdvanceWritePhase() bool {
	ws := getWriteSequenceState()
	if !ws.Active {
		return true
	}

	switch ws.Phase {
	case PhaseReset:
		ws.Phase = PhaseHold1
		ws.Progress = 0.25
	case PhaseHold1:
		ws.Phase = PhaseWrite
		ws.Progress = 0.5
	case PhaseWrite:
		ws.Phase = PhaseHold2
		ws.Progress = 0.75
	case PhaseHold2:
		ws.Phase = PhaseIdle
		ws.Active = false
		ws.Progress = 1.0
		return true
	}
	return false
}

// GetWritePhaseInfo returns the current write sequence state for UI display
func (ds *DeviceState) GetWritePhaseInfo() WriteSequenceState {
	ws := getWriteSequenceState()
	return *ws
}

// CancelWriteSequence aborts the current write sequence
func (ds *DeviceState) CancelWriteSequence() {
	ws := getWriteSequenceState()
	ws.Active = false
	ws.Phase = PhaseIdle
	ws.Progress = 0.0
}

// GetPhaseName returns a human-readable name for a write phase
func GetPhaseName(phase WritePhase) string {
	switch phase {
	case PhaseIdle:
		return "IDLE"
	case PhaseReset:
		return "RESET"
	case PhaseHold1:
		return "HOLD"
	case PhaseWrite:
		return "WRITE"
	case PhaseHold2:
		return "HOLD"
	default:
		return "UNKNOWN"
	}
}

// GetPhaseDuration returns the duration in nanoseconds for a phase
func GetPhaseDuration(phase WritePhase) int {
	switch phase {
	case PhaseReset:
		return PhaseResetDurationNs
	case PhaseHold1:
		return PhaseHold1DurationNs
	case PhaseWrite:
		return PhaseWriteDurationNs
	case PhaseHold2:
		return PhaseHold2DurationNs
	default:
		return 0
	}
}

// ============================================================================
// 4. ISPP STATE MACHINE WITH OVERSHOOT HANDLING
// ============================================================================

// ISPPResult represents the result of an ISPP iteration
type ISPPResult int

const (
	ISPPResultContinue      ISPPResult = iota // Continue iterating
	ISPPResultVerified                        // Target level reached
	ISPPResultOvershoot                       // Overshoot detected, reset needed
	ISPPResultMaxIterations                   // Max iterations reached
)

// ISPP constants
const (
	ISPPMaxIterations   = 5
	ISPPToleranceLevels = 0 // Exact match required
)

// ISPPState holds the state of an active ISPP (Incremental Step Pulse Programming) loop
type ISPPState struct {
	Active       bool
	Iteration    int
	MaxIter      int
	TargetRow    int
	TargetCol    int
	TargetLevel  int
	CurrentLevel int
	Voltage      float64
	Direction    HysteresisDirection
	Verified     bool
	Complete     bool
	Success      bool
}

// isppState is the global ISPP state
var isppState *ISPPState

// getISPPState returns the global ISPP state
func getISPPState() *ISPPState {
	if isppState == nil {
		isppState = &ISPPState{MaxIter: ISPPMaxIterations}
	}
	return isppState
}

// StartISPP begins an ISPP loop for a cell
func (ds *DeviceState) StartISPP(row, col, targetLevel, currentLevel int) {
	is := getISPPState()

	direction := ds.GetWriteDirection(row, col, currentLevel, targetLevel)
	ascending := direction == DirectionAscending

	is.Active = true
	is.Iteration = 0
	is.MaxIter = ISPPMaxIterations
	is.TargetRow = row
	is.TargetCol = col
	is.TargetLevel = targetLevel
	is.CurrentLevel = currentLevel
	is.Voltage = ds.GetVoltageForLevel(targetLevel, ascending)
	is.Direction = direction
	is.Verified = false
	is.Complete = false
	is.Success = false
}

// ISPPIterate performs one write-verify iteration
// Returns the result indicating whether to continue, success, overshoot, or max iterations
func (ds *DeviceState) ISPPIterate(newCurrentLevel int) ISPPResult {
	is := getISPPState()
	if !is.Active {
		return ISPPResultVerified
	}

	is.CurrentLevel = newCurrentLevel
	is.Iteration++

	// Check if target reached (within tolerance)
	diff := is.TargetLevel - is.CurrentLevel
	if diff < 0 {
		diff = -diff
	}
	if diff <= ISPPToleranceLevels {
		is.Verified = true
		is.Complete = true
		is.Success = true
		is.Active = false
		return ISPPResultVerified
	}

	// Check for overshoot
	if is.Direction == DirectionAscending && is.CurrentLevel > is.TargetLevel {
		return ISPPResultOvershoot
	}
	if is.Direction == DirectionDescending && is.CurrentLevel < is.TargetLevel {
		return ISPPResultOvershoot
	}

	// Check max iterations
	if is.Iteration >= is.MaxIter {
		is.Complete = true
		is.Success = false
		is.Active = false
		return ISPPResultMaxIterations
	}

	// Adjust voltage for next iteration
	voltageStep := (ds.writeRange.Max - ds.writeRange.Min) / 60.0 // Fine step
	if is.Direction == DirectionAscending {
		is.Voltage += voltageStep
		if is.Voltage > ds.writeRange.Max {
			is.Voltage = ds.writeRange.Max
		}
	} else {
		is.Voltage -= voltageStep
		if is.Voltage < ds.writeRange.Min {
			is.Voltage = ds.writeRange.Min
		}
	}

	return ISPPResultContinue
}

// HandleOvershoot performs RESET-to-saturation when write overshoots target
// Returns true if reset was performed
func (ds *DeviceState) HandleOvershoot(row, col int) bool {
	is := getISPPState()
	if !is.Active {
		return false
	}

	// Reset to saturation based on direction
	if is.Direction == DirectionAscending {
		// Ascending overshoot: reset to level 0 (negative saturation)
		is.CurrentLevel = 0
		is.Direction = DirectionAscending // Keep ascending for retry
	} else {
		// Descending overshoot: reset to level 29 (positive saturation)
		is.CurrentLevel = 29
		is.Direction = DirectionDescending // Keep descending for retry
	}

	// Recalculate voltage for target from new position
	ascending := is.Direction == DirectionAscending
	is.Voltage = ds.GetVoltageForLevel(is.TargetLevel, ascending)

	return true
}

// GetISPPStatus returns the current ISPP state for UI display
func (ds *DeviceState) GetISPPStatus() ISPPState {
	is := getISPPState()
	return *is
}

// CancelISPP aborts the current ISPP loop
func (ds *DeviceState) CancelISPP() {
	is := getISPPState()
	is.Active = false
	is.Complete = true
	is.Success = false
}

// ============================================================================
// 5. V/2 HALF-SELECT VISUALIZATION STATE
// ============================================================================

// HalfSelectVoltageRatio is the V/2 ratio for half-selected cells
const HalfSelectVoltageRatio = 0.5

// HalfSelectVisualization holds the state for V/2 overlay visualization
type HalfSelectVisualization struct {
	Enabled        bool
	FullVoltage    float64
	HalfVoltage    float64
	SelectedRow    int
	SelectedCol    int
	HalfSelectRows []int // Rows with V/2 (same column, different rows)
	HalfSelectCols []int // Cols with V/2 (same row, different columns)
}

// halfSelectState is the global half-select visualization state
var halfSelectState *HalfSelectVisualization

// getHalfSelectStateInternal returns the global half-select state
func getHalfSelectStateInternal() *HalfSelectVisualization {
	if halfSelectState == nil {
		halfSelectState = &HalfSelectVisualization{}
	}
	return halfSelectState
}

// EnableHalfSelectVisualization enables V/2 overlay for a write operation
// Only meaningful for 0T1R (passive) architecture
func (ds *DeviceState) EnableHalfSelectVisualization(row, col int, fullVoltage float64) {
	hs := getHalfSelectStateInternal()

	hs.Enabled = true
	hs.FullVoltage = fullVoltage
	hs.HalfVoltage = fullVoltage * HalfSelectVoltageRatio
	hs.SelectedRow = row
	hs.SelectedCol = col

	// All other rows in the same column get V/2
	hs.HalfSelectRows = make([]int, 0)
	for r := 0; r < ds.rows; r++ {
		if r != row {
			hs.HalfSelectRows = append(hs.HalfSelectRows, r)
		}
	}

	// All other columns in the same row get V/2
	hs.HalfSelectCols = make([]int, 0)
	for c := 0; c < ds.cols; c++ {
		if c != col {
			hs.HalfSelectCols = append(hs.HalfSelectCols, c)
		}
	}
}

// DisableHalfSelectVisualization disables the V/2 overlay
func (ds *DeviceState) DisableHalfSelectVisualization() {
	hs := getHalfSelectStateInternal()
	hs.Enabled = false
	hs.HalfSelectRows = nil
	hs.HalfSelectCols = nil
}

// GetHalfSelectState returns the current V/2 visualization state
func (ds *DeviceState) GetHalfSelectState() HalfSelectVisualization {
	hs := getHalfSelectStateInternal()
	return *hs
}

// IsHalfSelected returns true if the given cell is in half-select state
func (ds *DeviceState) IsHalfSelected(row, col int) bool {
	hs := getHalfSelectStateInternal()
	if !hs.Enabled {
		return false
	}

	// Check if in half-select row (same column as selected)
	if col == hs.SelectedCol {
		for _, r := range hs.HalfSelectRows {
			if r == row {
				return true
			}
		}
	}

	// Check if in half-select column (same row as selected)
	if row == hs.SelectedRow {
		for _, c := range hs.HalfSelectCols {
			if c == col {
				return true
			}
		}
	}

	return false
}
