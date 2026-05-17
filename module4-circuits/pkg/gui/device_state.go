//go:build legacy_fyne

// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device state type definitions, constants, and constructor.
// Methods are split across domain-focused files:
//   - device_state_config.go  — configuration setters/getters, voltage ranges, resize
//   - device_state_read.go    — conductance mapping, effective cell voltage, DAC quantization
//   - device_state_compute.go — MVM compute path (ideal and coupled-array simulation)
//   - device_state_write.go   — write path, ISPP, hysteresis, voltage calibration, half-select
package gui

import (
	"sync"

	"fecim-lattice-tools/config/physics"
	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
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
	OpModeRead    OpMode = iota // READ: Single row active, safe voltage (0-1.0V)
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
	DACManual      DACMode = iota // User entered each voltage
	DACReadPreset                 // Selected column at readVoltage, others 0 (single cell read)
	DACWritePreset                // Selected column at write voltage, others 0 (single cell write)
	DACInputVector                // From digital input vector (0-255 -> 0-1V)
	DACRandom                     // Random voltages
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
		// Fallback: field_min_ratio=0.7, field_max_ratio=2.5 (safe non-destructive reads)
		return CalibrationParams{
			FieldMinRatio: 0.7,
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
	// Mutex for thread-safe access to state
	mu sync.RWMutex

	// Dimensions
	rows int
	cols int

	// Passive mode flag - when true, ALL WLs are always on (0T1R architecture)
	isPassive bool

	// Operation mode (READ/WRITE/COMPUTE)
	opMode OpMode // Current operation mode

	// WL configuration (derived from opMode)
	wlMode     WLMode
	activeRows []bool    // true = WL HIGH for that row
	wlVoltages []float64 // WL voltages for V/2 scheme (passive mode write)

	// DAC inputs (per column)
	dacVoltages  []float64
	dacMode      DACMode
	dacRangeMode DACRangeMode // Current DAC range (read vs write)

	// Voltage ranges (derived from material + calibration config)
	readRange   VoltageRange      // 0 to FieldMinRatio*Vc for read/compute
	writeRange  VoltageRange      // Vc to FieldMaxRatio*Vc for write operations
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
	material *sharedphysics.HZOMaterial

	// Peripherals reference
	tia *peripherals.TIA
	adc *peripherals.ADC
	dac *peripherals.DAC

	// Coupling simulation (Tier A arraysim)
	couplingMode               arraysim.CouplingMode
	arrayEngine                arraysim.Engine
	cellGeometry               arraysim.CellGeometry
	wireParams                 arraysim.WireParams
	selectorEnabled            bool
	selectorRon                float64
	selectorLeakageConductance float64
	coupledCellVoltages        [][]float64 // Last coupled Vcell (V)
	coupledCellCurrents        [][]float64 // Last coupled Icell (A)

	enableDACNonlinearity bool                      // Apply DAC INL/DNL in compute path
	peripheralTemperature float64                   // Temperature for peripheral nonlinearity model (K)
	processCorner         peripherals.ProcessCorner // Process corner for peripheral nonlinearity

	// Embedded state machines (previously global)
	hysteresisState    HysteresisState
	writeSequenceState WriteSequenceState
	isppState          ISPPState
	halfSelectState    HalfSelectVisualization
	voltageCalibration *PerLevelVoltageCalibration
	forceResetNextSeq  bool

	// Shared ISPP calculator for voltage math
	isppCalc *sharedphysics.ISPPCalculator

	// ISPP engine selection
	isppEngine ISPPEngine
}

// NewDeviceState creates a new device state with specified dimensions
// Loads calibration parameters from physics.yaml for voltage range calculation
func NewDeviceState(rows, cols int, tia *peripherals.TIA, adc *peripherals.ADC) *DeviceState {
	defaultMaterial := sharedphysics.FeCIMMaterial()
	defaultGeometry := arraysim.DefaultCellGeometry()
	defaultGeometry.Film = sharedphysics.GeometryFromMaterial(defaultMaterial)

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
		material:     defaultMaterial,         // Default to FeCIM material
		calibParams:  loadCalibrationParams(), // Load from physics.yaml
		tia:          tia,
		adc:          adc,
		dac:          peripherals.DefaultDAC(),
		// Default to Tier A so READ path uses coupled array-level simulation.
		couplingMode:               arraysim.CouplingTierA,
		arrayEngine:                arraysim.NewTierASolver(),
		cellGeometry:               defaultGeometry,
		wireParams:                 arraysim.WireParams{},
		selectorEnabled:            false,
		selectorRon:                0,
		selectorLeakageConductance: 0,
		peripheralTemperature:      300.0,
		processCorner:              peripherals.CornerTypical,
		// Initialize embedded state machines
		hysteresisState: HysteresisState{
			LastLevel: make(map[string]int),
			Direction: make(map[string]HysteresisDirection),
		},
		isppState:  ISPPState{MaxIter: ISPPMaxIterations},
		isppEngine: ISPPEngineLevel,
	}

	// Initialize ISPP calculator with default material
	if defaultMaterial != nil {
		ec := defaultMaterial.CoerciveVoltage()
		numLevels := defaultMaterial.GetNumLevels()
		ds.isppCalc = sharedphysics.NewISPPCalculator(ec, numLevels)
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
