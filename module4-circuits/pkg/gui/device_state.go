// Package gui provides Fyne-based GUI components for peripheral circuit visualization.
// This file contains the unified device state for the simulation view.
package gui

import (
	"fmt"
	"math"
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

// updateVoltageRanges calculates voltage ranges from material properties and calibration config
// Read range: 0 to FieldMinRatio * Vc (below coercive voltage, non-destructive sensing)
// Write range: Vc to FieldMaxRatio * Vc (exceeds coercive voltage for polarization switching)
//
// From physics.yaml calibration section:
//
//	field_min_ratio: 0.5  -> Read max = 0.5 * Vc
//	field_max_ratio: 2.5  -> Write max = 2.5 * Vc
func (ds *DeviceState) updateVoltageRanges() {
	// Ensure material is set - use default FeCIM if not
	if ds.material == nil {
		ds.material = sharedphysics.FeCIMMaterial()
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

	// Write range: bipolar and derived from material coercive voltage.
	// Ferroelectric WRITE needs both polarities (ERASE/PROGRAM), so map DAC
	// to [-Vmax, +Vmax], where Vmax scales from material Vc.
	writeMaxAbs := ds.calibParams.FieldMaxRatio * Vc
	if writeMaxAbs > MaxPracticalVoltage {
		writeMaxAbs = MaxPracticalVoltage
	}

	ds.writeRange = VoltageRange{
		Min:       -writeMaxAbs,
		Max:       +writeMaxAbs,
		StepSize:  (2 * writeMaxAbs) / float64(numLevels-1),
		NumLevels: numLevels,
	}
}

// SetMaterial changes the ferroelectric material used for conductance calculation
func (ds *DeviceState) SetMaterial(mat *sharedphysics.HZOMaterial) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.material = mat
	ds.updateVoltageRanges() // Recalculate voltage ranges for new material

	// Initialize ISPP calculator with material's coercive voltage
	if mat != nil {
		ec := mat.CoerciveVoltage()
		numLevels := mat.GetNumLevels()
		ds.isppCalc = sharedphysics.NewISPPCalculator(ec, numLevels)
		ds.cellGeometry.Film = sharedphysics.GeometryFromMaterial(mat)
	}
}

// GetMaterial returns the current material
func (ds *DeviceState) GetMaterial() *sharedphysics.HZOMaterial {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.material
}

// SetDACNonlinearity enables/disables DAC nonlinearity in compute
func (ds *DeviceState) SetDACNonlinearity(enable bool) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.enableDACNonlinearity = enable
}

// IsDACNonlinearityEnabled returns whether DAC nonlinearity is applied
func (ds *DeviceState) IsDACNonlinearityEnabled() bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.enableDACNonlinearity
}

// SetPeripheralPVT configures temperature + process corner for DAC/ADC nonlinearity models.
func (ds *DeviceState) SetPeripheralPVT(temperatureK float64, corner peripherals.ProcessCorner) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if temperatureK <= 0 {
		temperatureK = 300.0
	}
	ds.peripheralTemperature = temperatureK
	switch corner {
	case peripherals.CornerFast, peripherals.CornerSlow, peripherals.CornerTypical:
		ds.processCorner = corner
	default:
		ds.processCorner = peripherals.CornerTypical
	}
}

// GetPeripheralPVT returns the current peripheral temperature and process corner.
func (ds *DeviceState) GetPeripheralPVT() (float64, peripherals.ProcessCorner) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.peripheralTemperature, ds.processCorner
}

// SetCouplingMode enables/disables array coupling simulation.
func (ds *DeviceState) SetCouplingMode(mode arraysim.CouplingMode) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.couplingMode = mode
	switch mode {
	case arraysim.CouplingIdeal:
		ds.arrayEngine = nil
		ds.coupledCellVoltages = nil
		ds.coupledCellCurrents = nil
	case arraysim.CouplingTierA:
		ds.arrayEngine = arraysim.NewTierASolver()
	case arraysim.CouplingTierB:
		ds.arrayEngine = arraysim.NewTierBSolver()
	default:
		ds.arrayEngine = arraysim.NewTierASolver()
	}
}

// GetCouplingMode returns the active coupling model.
func (ds *DeviceState) GetCouplingMode() arraysim.CouplingMode {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.couplingMode
}

// SetCellGeometry updates the cell geometry used for coupling calculations.
func (ds *DeviceState) SetCellGeometry(geom arraysim.CellGeometry) {
	ds.mu.Lock()
	ds.cellGeometry = geom.WithDefaults()
	ds.mu.Unlock()
}

// GetCellGeometry returns the current cell geometry.
func (ds *DeviceState) GetCellGeometry() arraysim.CellGeometry {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.cellGeometry
}

// SetWireParams updates wire resistance parameters for coupled solvers.
func (ds *DeviceState) SetWireParams(wire arraysim.WireParams) {
	ds.mu.Lock()
	ds.wireParams = wire
	ds.mu.Unlock()
}

// SetSelectorSeriesParams configures Tier-A selector series-R and leakage model.
func (ds *DeviceState) SetSelectorSeriesParams(enabled bool, ronOhm, leakageConductance float64) {
	ds.mu.Lock()
	ds.selectorEnabled = enabled
	ds.selectorRon = ronOhm
	ds.selectorLeakageConductance = leakageConductance
	ds.mu.Unlock()
}

// SetDACBits changes the DAC resolution (4-8 bits) and recreates the DAC peripheral.
func (ds *DeviceState) SetDACBits(bits int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if bits < 4 || bits > 8 {
		return
	}

	base := peripherals.DefaultDAC()
	if ds.dac != nil {
		base.VrefLow = ds.dac.VrefLow
		base.VrefHigh = ds.dac.VrefHigh
		base.INL = ds.dac.INL
		base.DNL = ds.dac.DNL
		base.SettleTime = ds.dac.SettleTime
	}
	base.Bits = bits
	ds.dac = base
}

// GetDACBits returns the current DAC resolution in bits.
func (ds *DeviceState) GetDACBits() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.dac != nil {
		return ds.dac.Bits
	}
	return peripherals.DefaultDAC().Bits
}

// SetADCBits changes the ADC resolution (5, 6, 7, or 8 bits)
func (ds *DeviceState) SetADCBits(bits int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()
	if ds.adc != nil && bits >= 5 && bits <= 8 {
		ds.adc.Bits = bits
	}
}

// GetADCBits returns the current ADC resolution in bits
func (ds *DeviceState) GetADCBits() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.adc != nil {
		return ds.adc.Bits
	}
	return 5 // Default
}

// GetADCLevels returns the number of ADC output levels (2^bits)
func (ds *DeviceState) GetADCLevels() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.adc != nil {
		return 1 << ds.adc.Bits
	}
	return 32 // Default 5-bit
}

// GetMaterialName returns the name of the current material
func (ds *DeviceState) GetMaterialName() string {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.material != nil {
		return ds.material.Name
	}
	return "Unknown"
}

// SetISPPEngine changes the ISPP engine used by the write path.
func (ds *DeviceState) SetISPPEngine(engine ISPPEngine) {
	ds.mu.Lock()
	ds.isppEngine = engine
	ds.mu.Unlock()
}

// GetISPPEngine returns the active ISPP engine.
func (ds *DeviceState) GetISPPEngine() ISPPEngine {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.isppEngine
}

// conductanceBounds returns the material conductance bounds with a safe fallback.
func (ds *DeviceState) conductanceBounds() (float64, float64) {
	mat := ds.material
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	gmin := mat.Gmin
	gmax := mat.Gmax
	if gmin == 0 && gmax == 0 {
		gmin = 1e-6   // Match material.DiscreteLevel fallback
		gmax = 100e-6 // Match material.DiscreteLevel fallback
	}
	scale := ds.cellGeometry.Film.ConductanceScale(sharedphysics.GeometryFromMaterial(mat))
	return gmin * scale, gmax * scale
}

// levelToConductance maps a discrete level to conductance using the active material.
func (ds *DeviceState) levelToConductance(level, levels int) float64 {
	mat := ds.material
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	levels = ds.resolveConductanceLevels(levels)
	baseG := mat.DiscreteLevel(level, levels)
	referenceGeom := sharedphysics.GeometryFromMaterial(mat)
	return baseG * ds.cellGeometry.Film.ConductanceScale(referenceGeom)
}

// resolveConductanceLevels picks quantization levels used for level→conductance mapping.
// READ defaults to material-native levels so material selection propagates into READ current.
func (ds *DeviceState) resolveConductanceLevels(quantLevels int) int {
	if ds.opMode == OpModeRead && ds.material != nil {
		if materialLevels := ds.material.GetNumLevels(); materialLevels > 0 {
			return materialLevels
		}
	}
	if quantLevels > 0 {
		return quantLevels
	}
	if ds.material != nil {
		if materialLevels := ds.material.GetNumLevels(); materialLevels > 0 {
			return materialLevels
		}
	}
	return 30
}

// conductanceToLevel maps conductance to a discrete level using material bounds.
func (ds *DeviceState) conductanceToLevel(gPhys float64, levels int) int {
	if levels <= 0 {
		levels = ds.writeRange.NumLevels
	}
	if levels <= 0 {
		levels = 30
	}
	gmin, gmax := ds.conductanceBounds()
	if gmax <= gmin {
		return 0
	}
	// Use the model-aware inverse so conductanceToLevel is consistent with
	// levelToConductance (which uses mat.DiscreteLevel → PolarizationToConductanceWithParams
	// with the exponential model). Linear inverse introduces 5–10 level errors at mid-range.
	model := sharedphysics.ConductanceExponential // default matches ParseConductanceModel default
	if ds.material != nil {
		model = sharedphysics.ParseConductanceModel(ds.material.ConductanceModel)
	}
	gNorm := sharedphysics.PhysicalToNormalizedModel(gPhys, gmin, gmax, model)
	return sharedphysics.GetLevel(gNorm, levels)
}

// programLevelFromCoupledVoltage advances one cell's programmed level using the
// *actual* cell voltage seen after DAC + array coupling (IR drop / half-select).
// This is used by the level-engine ISPP write path to avoid ideal-voltage updates.
func (ds *DeviceState) programLevelFromCoupledVoltage(currentLevel int, effectiveV float64, pulseWidth float64, quantLevels int) int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	levels := ds.resolveConductanceLevels(quantLevels)
	if levels < 2 {
		levels = 2
	}
	if currentLevel < 0 {
		currentLevel = 0
	}
	if currentLevel >= levels {
		currentLevel = levels - 1
	}
	if math.Abs(effectiveV) < 1e-12 {
		return currentLevel
	}

	mat := ds.material
	if mat == nil {
		mat = sharedphysics.FeCIMMaterial()
	}
	if mat.Ps == 0 {
		return currentLevel
	}

	if pulseWidth <= 0 {
		pulseWidth = mat.Tau
	}
	if pulseWidth <= 0 {
		pulseWidth = float64(PhaseWriteDurationNs) * 1e-9
	}

	gmin, gmax := ds.conductanceBounds()
	currentG := ds.levelToConductance(currentLevel, levels)
	conductanceModel := sharedphysics.ParseConductanceModel(mat.ConductanceModel)
	currentP := sharedphysics.ConductanceToPolarizationModel(currentG, gmin, gmax, mat.Ps, conductanceModel)

	solver := sharedphysics.NewLKSolver()
	solver.ConfigureFromMaterial(mat)
	solver.Temperature = 300
	solver.EnableNoise = false
	solver.UseNLS = false
	if !solver.UseMaterialAlpha {
		solver.UpdateParams()
	}
	solver.SetState(currentP)

	geom := ds.cellGeometry.WithDefaults()
	eField := geom.Film.ElectricField(effectiveV)
	solver.Step(eField, pulseWidth)

	newP := solver.GetState()
	newG := sharedphysics.PolarizationToConductanceWithParams(newP, mat.Ps, gmin, gmax, sharedphysics.ParseConductanceModel(mat.ConductanceModel), mat.KvT, mat.VGSReadV, mat.VT0V)
	newLevel := ds.conductanceToLevel(newG, levels)
	if newLevel < 0 {
		newLevel = 0
	}
	if newLevel >= levels {
		newLevel = levels - 1
	}
	return newLevel
}

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

// DACWriteVoltage converts a target analog write voltage into the DAC's actual output.
// Returns the applied voltage and the DAC code used. This models DAC quantization (and
// optional nonlinearity) before the voltage is applied to the array.
func (ds *DeviceState) DACWriteVoltage(targetVoltage float64) (float64, int) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.dacWriteVoltageLocked(targetVoltage)
}

// AppliedWriteVoltageForLevel returns the DAC-applied write voltage for a target level.
// This is useful for UI previews where the applied voltage (post-DAC) is desired.
func (ds *DeviceState) AppliedWriteVoltageForLevel(level int, ascending bool) float64 {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}
	target := ds.getVoltageForLevelInternal(level, ascending)
	applied, _ := ds.dacWriteVoltageLocked(target)
	return applied
}

// GetEffectiveCellVoltage returns the effective cell voltage considering WL/BL biasing.
// For passive (0T1R) mode, Vcell = WL - BL. For 1T1R/2T1R, Vcell = BL.
func (ds *DeviceState) GetEffectiveCellVoltage(row, col int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.effectiveCellVoltageLocked(row, col)
}

// GetCoupledCellCurrent returns the last coupled per-cell current (A) from arraysim.
// Returns 0 when coupling is disabled, no solve has run yet, or indices are out of range.
func (ds *DeviceState) GetCoupledCellCurrent(row, col int) float64 {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if row < 0 || row >= ds.rows || col < 0 || col >= ds.cols {
		return 0
	}
	if ds.coupledCellCurrents == nil {
		return 0
	}
	if row >= len(ds.coupledCellCurrents) || col >= len(ds.coupledCellCurrents[row]) {
		return 0
	}
	return ds.coupledCellCurrents[row][col]
}

// GetCoupledCellSnapshot returns deep copies of the latest per-cell coupled
// voltages (V) and currents (A). Nil slices are returned when no coupled solve
// result is available.
func (ds *DeviceState) GetCoupledCellSnapshot() ([][]float64, [][]float64) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	copy2D := func(src [][]float64) [][]float64 {
		if src == nil {
			return nil
		}
		dst := make([][]float64, len(src))
		for r := range src {
			dst[r] = make([]float64, len(src[r]))
			copy(dst[r], src[r])
		}
		return dst
	}

	return copy2D(ds.coupledCellVoltages), copy2D(ds.coupledCellCurrents)
}

func (ds *DeviceState) dacWriteVoltageLocked(targetVoltage float64) (float64, int) {
	if ds.dac == nil {
		return targetVoltage, -1
	}
	minV := ds.dac.VrefLow
	maxV := ds.dac.VrefHigh
	if targetVoltage < minV {
		targetVoltage = minV
	}
	if targetVoltage > maxV {
		targetVoltage = maxV
	}

	span := maxV - minV
	if span == 0 {
		return targetVoltage, 0
	}

	normalized := (targetVoltage - minV) / span
	code := int(math.Round(normalized * float64(ds.dac.Levels()-1)))
	if code < 0 {
		code = 0
	}
	maxCode := ds.dac.Levels() - 1
	if code > maxCode {
		code = maxCode
	}

	var applied float64
	if ds.enableDACNonlinearity {
		applied = ds.dac.ConvertWithCondition(code, ds.peripheralTemperature, ds.processCorner)
	} else {
		applied = ds.dac.Convert(code)
	}

	return applied, code
}

func (ds *DeviceState) effectiveCellVoltageLocked(row, col int) float64 {
	if row < 0 || row >= ds.rows || col < 0 || col >= ds.cols {
		return 0
	}
	if ds.couplingMode != arraysim.CouplingIdeal && ds.coupledCellVoltages != nil {
		if row < len(ds.coupledCellVoltages) && col < len(ds.coupledCellVoltages[row]) {
			v := ds.coupledCellVoltages[row][col]
			if math.Abs(v) < 1e-12 {
				return 0
			}
			return v
		}
	}
	bl := ds.dacVoltages[col]
	if ds.isPassive {
		wl := ds.wlVoltages[row]
		v := wl - bl
		if math.Abs(v) < 1e-12 {
			return 0
		}
		return v
	}
	if math.Abs(bl) < 1e-12 {
		return 0
	}
	return bl
}

// applyDACNonlinearityLocked applies DAC nonlinearity for read/compute path voltages.
// Caller must hold ds.mu.
func (ds *DeviceState) applyDACNonlinearityLocked(voltage float64) float64 {
	if !ds.enableDACNonlinearity || ds.dac == nil || ds.dacRangeMode != DACRangeRead {
		return voltage
	}
	voltageMag := math.Abs(voltage)
	normalized := 0.0
	if ds.readRange.Max > 0 {
		normalized = voltageMag / ds.readRange.Max
	}
	if normalized > 1.0 {
		normalized = 1.0
	}
	if normalized < 0 {
		normalized = 0
	}
	level := int(normalized * float64(ds.dac.Levels()-1))
	applied := ds.dac.ConvertWithCondition(level, ds.peripheralTemperature, ds.processCorner)
	return math.Copysign(applied, voltage)
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
		//   - Row currents follow I_i = Σ_j (G_ij × Vj).
		//
		// Mapping (units):
		//   - UI supplies "byte-like" codes in the range 0..255.
		//   - We map 0 → 0V and 255 → readRange.Max (compute-safe full-scale).
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

// ============================================================================
// DAC-ONLY COLUMN DRIVE (Passive 0T1R Mode Only)
// ============================================================================
//
// For passive (0T1R) WRITE operations the DAC drives the selected BL while all
// WLs are grounded through the TIA virtual ground. There is no V/2 splitting.
//
// Target cell (SET operation):
//   - All WLs: 0V (grounded / TIA virtual ground)
//   - Selected BL (DAC): -V_write (full write voltage)
//   - Effective ΔV = WL − BL = 0 − (−V_write) = +V_write (full switching)
//
// Column disturb (same column, different row):
//   - WL = 0, BL = -V_write → ΔV = +V_write (FULL write — entire column switches)
//
// Same-row cells (different column):
//   - WL = 0, BL = 0 → ΔV = 0 (safe — unselected BLs grounded)
//
// Unselected cells (different row AND column):
//   - WL = 0, BL = 0 → ΔV = 0 (no disturb)
//
// For 1T1R/2T1R modes, transistor gate on selected row completes the circuit;
// only the selected cell [row,col] can switch (transistor isolation).
// ============================================================================

// ApplyHalfSelectWrite applies voltage biasing for passive (0T1R) write operation.
// Implements DAC-Only Column Drive: since rows are grounded (TIA virtual ground),
// the full write voltage is applied to the selected column.
//
// Target cell (SET operation):
//   - All WLs: 0V (Grounded / TIA Virtual Ground)
//   - Selected BL (DAC): -V_write (SET: positive ΔV across cell)
//   - Effective ΔV = WL - BL = 0 - (-V_write) = +V_write (full switching)
//
// Consequence: this performs a COLUMN WRITE — all cells in the selected column
// see the full V_write. Unselected columns see 0V (no disturb).
//
// For 1T1R/2T1R modes, transistor isolation eliminates need for column drive.
func (ds *DeviceState) ApplyHalfSelectWrite(targetRow, targetCol int, writeVoltage float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.isPassive {
		// Non-passive modes use transistor isolation
		// Apply full voltage to selected BL only, WL controls transistor gate
		ds.setDACVoltageLocked(targetCol, writeVoltage)
		return
	}

	// DAC-Only Drive Scheme: rows grounded, selected column driven to -V_write
	for i := range ds.wlVoltages {
		ds.wlVoltages[i] = 0
	}

	dacV := -writeVoltage
	for i := range ds.dacVoltages {
		if i == targetCol {
			ds.dacVoltages[i] = dacV
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
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.resetWriteVoltagesLocked()
}

func (ds *DeviceState) resetWriteVoltagesLocked() {
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
	ds.mu.RLock()
	defer ds.mu.RUnlock()
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
	ds.mu.Lock()
	defer ds.mu.Unlock()
	ds.selectedRow = row
	ds.selectedCol = col
	if ds.wlMode == WLSingle {
		ds.setWLSingleLocked(row)
	}
}

// Compute runs the device simulation given the weight matrix
func (ds *DeviceState) Compute(weights [][]int, quantLevels int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.couplingMode != arraysim.CouplingIdeal && ds.arrayEngine != nil {
		if ds.computeWithArraysimLocked(weights, quantLevels) {
			ds.LogCompute(weights, quantLevels)
			return
		}
	}

	// Fallback: ideal computation path (existing behavior).
	ds.coupledCellVoltages = nil
	ds.coupledCellCurrents = nil
	ds.computeIdealLocked(weights, quantLevels)
	ds.LogCompute(weights, quantLevels)
}

func (ds *DeviceState) computeIdealLocked(weights [][]int, quantLevels int) {
	// During programming, the sense chain (TIA/ADC) is disconnected.
	// Don't compute row currents using write-level WL/BL voltages.
	if ds.dacRangeMode == DACRangeWrite {
		for r := 0; r < ds.rows; r++ {
			ds.rowCurrents[r] = 0
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
		}
		return
	}

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
			voltage := ds.effectiveCellVoltageLocked(r, c)

			// Apply DAC nonlinearity if enabled (read/compute path only).
			if ds.enableDACNonlinearity && ds.dac != nil && ds.dacRangeMode == DACRangeRead {
				voltage = ds.applyDACNonlinearityLocked(voltage)
			} else if ds.isPassive {
				// For passive mode, use WL/BL effective voltage even without DAC nonlinearity.
				voltage = ds.wlVoltages[r] - ds.dacVoltages[c]
			}

			if math.Abs(voltage) < 0.01 {
				continue
			}

			// Get cell conductance from weight using material physics model
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}

			// Use shared level→conductance mapping so ideal and coupled paths
			// respect the same material + geometry scaling dependencies.
			conductanceS := ds.levelToConductance(level, quantLevels)

			// Convert to µS for current calculation.
			conductanceUS := conductanceS * 1e6
			current := conductanceUS * voltage // I = G * V (signed µA since G is in µS)
			totalCurrent += current
		}

		ds.rowCurrents[r] = totalCurrent

		if ds.tia == nil {
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		currentA := totalCurrent * 1e-6 // µA to A

		// Count active columns (columns with |effective voltage| > 0.01V)
		activeColCount := 0
		for c := 0; c < ds.cols; c++ {
			v := ds.effectiveCellVoltageLocked(r, c)
			if math.Abs(v) > 0.01 {
				activeColCount++
			}
		}

		effectiveGain := ds.tia.Gain
		if activeColCount > 1 {
			// MVM mode: scale gain based on number of active columns to reduce saturation.
			effectiveGain = ds.tia.Gain / float64(activeColCount)
		}

		vref := 0.0
		if activeColCount <= 1 {
			// Single-cell read path keeps the TIA output offset.
			vref = ds.tia.OutputOffset
		}

		ds.rowVoltages[r], ds.rowLevels[r], ds.saturated[r] = ds.convertSenseLocked(currentA, effectiveGain, vref)
	}
}

func (ds *DeviceState) convertSenseLocked(rowCurrentA, gainOhm, vrefV float64) (float64, int, bool) {
	if ds.tia == nil {
		return 0, 0, false
	}

	if ds.adc == nil {
		vout := vrefV + rowCurrentA*gainOhm
		if vout < 0 {
			vout = 0
		}
		if vout > ds.tia.MaxOutputVoltage {
			vout = ds.tia.MaxOutputVoltage
		}
		return vout, 0, false
	}

	sense := arraysim.SenseChain{
		TIA: arraysim.TIAConfig{
			Rf:   gainOhm,
			Vref: vrefV,
			Vmin: 0,
			Vmax: ds.tia.MaxOutputVoltage,
		},
		ADC: arraysim.ADCConfig{
			Bits: ds.adc.Bits,
			Vmin: ds.adc.VrefLow,
			Vmax: ds.adc.VrefHigh,
		},
	}
	res := sense.ConvertCurrent(rowCurrentA)
	adcMaxLevel := (1 << ds.adc.Bits) - 1
	saturated := res.TIASaturated || res.ADCSaturated || res.Code >= adcMaxLevel
	return res.Vout, res.Code, saturated
}

func (ds *DeviceState) computeWithArraysimLocked(weights [][]int, quantLevels int) bool {
	if ds.arrayEngine == nil {
		return false
	}

	rows := ds.rows
	cols := ds.cols
	if rows == 0 || cols == 0 {
		return true
	}

	conductance := make([][]float64, rows)
	for r := 0; r < rows; r++ {
		conductance[r] = make([]float64, cols)
		for c := 0; c < cols; c++ {
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}
			conductance[r][c] = ds.levelToConductance(level, quantLevels)
		}
	}

	blApplied := make([]float64, cols)
	for c := 0; c < cols; c++ {
		blApplied[c] = ds.applyDACNonlinearityLocked(ds.dacVoltages[c])
	}

	wlApplied := make([]float64, rows)
	if ds.isPassive {
		copy(wlApplied, ds.wlVoltages)
	}

	blSolve := blApplied
	if !ds.isPassive {
		blSolve = make([]float64, cols)
		for c := 0; c < cols; c++ {
			blSolve[c] = -blApplied[c]
		}
	}

	params := arraysim.SolveParams{
		WLVoltages:      wlApplied,
		BLVoltages:      blSolve,
		Conductance:     conductance,
		ActiveRows:      ds.activeRows,
		SelectorEnabled: ds.selectorEnabled,
		SelectorRon:     ds.selectorRon,
		Selector: arraysim.SelectorDeviceParams{
			Enabled:        ds.selectorLeakageConductance > 0,
			OnConductance:  math.Inf(1),
			OffConductance: ds.selectorLeakageConductance,
		},
		Geometry: ds.cellGeometry,
		Wire:     ds.wireParams,
	}
	result, err := ds.arrayEngine.Solve(params)
	if err != nil {
		return false
	}

	ds.coupledCellVoltages = result.CellVoltages
	ds.coupledCellCurrents = result.CellCurrents

	for r := 0; r < rows; r++ {
		if !ds.activeRows[r] {
			ds.rowCurrents[r] = 0
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		rowCurrentA := 0.0
		if r < len(result.RowCurrents) {
			rowCurrentA = result.RowCurrents[r]
		} else if r < len(result.CellCurrents) {
			for c := 0; c < len(result.CellCurrents[r]); c++ {
				rowCurrentA += result.CellCurrents[r][c]
			}
		}
		// arraysim uses solver-current sign conventions; sense chain expects read-magnitude polarity.
		rowCurrentA = math.Abs(rowCurrentA)

		ds.rowCurrents[r] = rowCurrentA * 1e6

		if ds.tia == nil {
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		activeColCount := 0
		expectedMaxCurrentA := 0.0
		for c := 0; c < cols; c++ {
			vcell := blApplied[c]
			if ds.isPassive {
				vcell = wlApplied[r] - blApplied[c]
			}
			if math.Abs(vcell) > 0.01 {
				activeColCount++
			}
			expectedMaxCurrentA += conductance[r][c] * math.Abs(vcell)
		}

		effectiveGain := ds.tia.Gain
		if activeColCount > 1 {
			effectiveGain = effectiveGain / float64(activeColCount)
		}
		if expectedMaxCurrentA > 0 {
			targetV := 0.9 * ds.tia.MaxOutputVoltage
			if expectedMaxCurrentA*effectiveGain > targetV {
				effectiveGain = targetV / expectedMaxCurrentA
			}
		}

		vref := 0.0
		if activeColCount <= 1 {
			vref = ds.tia.OutputOffset
		}

		ds.rowVoltages[r], ds.rowLevels[r], ds.saturated[r] = ds.convertSenseLocked(rowCurrentA, effectiveGain, vref)
	}

	return true
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

// ============================================================================
// 1. PER-LEVEL VOLTAGE CALIBRATION
// ============================================================================

// PerLevelVoltageCalibration holds calibrated voltages for each level.
// Ascending and descending voltages are now distinct (physics-derived ISPP tanh mapping).
type PerLevelVoltageCalibration struct {
	AscendingVoltages  []float64 // Voltages for writing up (level 0→max)
	DescendingVoltages []float64 // Voltages for writing down (level max→0)
}

// initVoltageCalibrationInternal initializes the per-level voltage arrays (internal, no locking).
// Uses a Preisach-derived tanh mapping: ascending V_L = Ec*(1+atanh(P_L/Ps))*d,
// descending V_L = -Ec*(1+atanh(-P_L/Ps))*d — consistent with ispp.go PredictState.
// Falls back to linear interpolation when material Ec/thickness are unavailable.
// Caller must hold appropriate lock.
func (ds *DeviceState) initVoltageCalibrationInternal() {
	numLevels := ds.writeRange.NumLevels
	cal := &PerLevelVoltageCalibration{
		AscendingVoltages:  make([]float64, numLevels),
		DescendingVoltages: make([]float64, numLevels),
	}

	// Physics-based ISPP voltage calibration using Preisach-derived tanh mapping.
	//
	// For a ferroelectric with coercive field Ec and film thickness d, the remanent
	// polarization after an ISPP ascending pulse to peak field E_peak (starting from -Ps)
	// is approximately: P_rem = Ps * tanh((E_peak − Ec) / Delta)
	// Using Delta=Ec (same approximation as AdaptiveISPP.PredictState in shared/physics/ispp.go):
	//   E_peak = Ec * (1 + atanh(P_L/Ps))    → always ≥ 0 for P_L in (-Ps, +Ps)
	//   V_asc = E_peak * thickness (clamped to [0, WriteRange.Max])
	//
	// For descending (starting from +Ps):
	//   E_peak_desc = Ec * (1 + atanh(-P_L/Ps))
	//   V_desc = -E_peak_desc * thickness (clamped to [WriteRange.Min, 0])
	//
	// Level → normalized P: P_L/Ps = 2*L/(N-1) - 1 (linear in gNorm space)
	// Mid-level (L ≈ N/2): V_asc ≈ Ec*d = Vc (coercive voltage) — physically exact.
	var ec, thickness float64
	mat := ds.material
	if mat != nil && mat.Ec > 0 && mat.Thickness > 0 {
		ec = mat.Ec
		thickness = mat.Thickness
	}
	minV := ds.writeRange.Min
	maxV := ds.writeRange.Max
	const atanhClamp = 0.9999 // Avoid atanh(±1) → ±∞ divergence
	for i := 0; i < numLevels; i++ {
		ratio := 2.0*float64(i)/float64(numLevels-1) - 1.0 // P_L/Ps ∈ (-1, +1)
		if ratio > atanhClamp {
			ratio = atanhClamp
		}
		if ratio < -atanhClamp {
			ratio = -atanhClamp
		}

		if ec > 0 && thickness > 0 {
			// Ascending: positive pulse to program from erase state (−Ps → P_L)
			vAsc := ec * (1.0 + math.Atanh(ratio)) * thickness
			if vAsc > maxV {
				vAsc = maxV
			}
			if vAsc < 0 {
				vAsc = 0 // Ascending pulses are always non-negative
			}
			cal.AscendingVoltages[i] = vAsc

			// Descending: negative pulse to erase from program state (+Ps → P_L)
			vDesc := -ec * (1.0 + math.Atanh(-ratio)) * thickness
			if vDesc < minV {
				vDesc = minV
			}
			if vDesc > 0 {
				vDesc = 0 // Descending pulses are always non-positive
			}
			cal.DescendingVoltages[i] = vDesc
		} else {
			// Fallback: linear interpolation when material parameters are unavailable
			voltage := minV + float64(i)*(maxV-minV)/float64(numLevels-1)
			cal.AscendingVoltages[i] = voltage
			cal.DescendingVoltages[i] = voltage
		}
	}

	ds.voltageCalibration = cal
}

// InitVoltageCalibration initializes the per-level voltage arrays using the Preisach-based
// tanh ISPP mapping (ascending/descending are now distinct, physics-derived voltages).
func (ds *DeviceState) InitVoltageCalibration() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.initVoltageCalibrationInternal()
}

// getVoltageForLevelInternal returns the calibrated write voltage (internal, no locking)
// Caller must hold appropriate lock
func (ds *DeviceState) getVoltageForLevelInternal(level int, ascending bool) float64 {
	// Clamp level to valid range
	maxLevel := len(ds.voltageCalibration.AscendingVoltages) - 1
	if level < 0 {
		level = 0
	}
	if level > maxLevel {
		level = maxLevel
	}

	if ascending {
		return ds.voltageCalibration.AscendingVoltages[level]
	}
	return ds.voltageCalibration.DescendingVoltages[level]
}

// GetVoltageForLevel returns the calibrated write voltage for a target level
// direction: true = ascending (increasing level), false = descending (decreasing level)
func (ds *DeviceState) GetVoltageForLevel(level int, ascending bool) float64 {
	ds.mu.Lock()
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}
	result := ds.getVoltageForLevelInternal(level, ascending)
	ds.mu.Unlock()
	return result
}

// GetLevelForVoltage estimates the nearest discrete level for a given write voltage.
// Uses the calibrated per-level voltage table (ascending/descending) to avoid ad-hoc mapping.
func (ds *DeviceState) GetLevelForVoltage(voltage float64, ascending bool) int {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	levels := ds.voltageCalibration.AscendingVoltages
	if !ascending {
		levels = ds.voltageCalibration.DescendingVoltages
	}
	if len(levels) == 0 {
		return 0
	}

	closest := 0
	minDiff := math.Abs(levels[0] - voltage)
	for i, v := range levels {
		diff := math.Abs(v - voltage)
		if diff < minDiff {
			minDiff = diff
			closest = i
		}
	}

	return closest
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
	LastLevel map[string]int                 // key: "row,col" -> last written level
	Direction map[string]HysteresisDirection // key: "row,col" -> last direction
}

// cellKey generates a map key for a cell coordinate
func cellKey(row, col int) string {
	return fmt.Sprintf("%d,%d", row, col)
}

// RecordWrite updates the hysteresis state after a successful write
func (ds *DeviceState) RecordWrite(row, col, newLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	key := cellKey(row, col)

	oldLevel, exists := ds.hysteresisState.LastLevel[key]
	if exists {
		if newLevel > oldLevel {
			ds.hysteresisState.Direction[key] = DirectionAscending
		} else if newLevel < oldLevel {
			ds.hysteresisState.Direction[key] = DirectionDescending
		}
		// If equal, keep previous direction
	} else {
		ds.hysteresisState.Direction[key] = DirectionUnknown
	}

	ds.hysteresisState.LastLevel[key] = newLevel
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
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	key := cellKey(row, col)
	if dir, exists := ds.hysteresisState.Direction[key]; exists {
		return dir
	}
	return DirectionUnknown
}

// ============================================================================
// 3. 5-PHASE PROGRAM-VERIFY SEQUENCE STATE MACHINE
// ============================================================================

// WritePhase represents the current phase in a program-verify sequence
type WritePhase int

const (
	PhaseIdle   WritePhase = iota // No write in progress
	PhaseReset                    // Applying -V_sat (100ns)
	PhaseHold1                    // Zero field hold (50ns)
	PhaseWrite                    // Applying calibrated voltage (200ns)
	PhaseHold2                    // Zero field hold (50ns)
	PhaseVerify                   // Read/verify at low voltage (80ns)
)

// Phase timing constants (in nanoseconds for display, not real-time)
const (
	PhaseResetDurationNs  = 100
	PhaseHold1DurationNs  = 50
	PhaseWriteDurationNs  = 200
	PhaseHold2DurationNs  = 50
	PhaseVerifyDurationNs = 80
)

// WriteSequenceState holds the state of an active program-verify sequence
type WriteSequenceState struct {
	Active       bool
	Phase        WritePhase
	TargetRow    int
	TargetCol    int
	TargetLevel  int
	CurrentLevel int
	WriteVoltage float64 // Calibrated voltage for target level
	PhaseVoltage float64 // Actual applied voltage for current phase
	Progress     float64 // 0.0 to 1.0 progress through sequence
}

// StartWriteSequence begins a program-verify sequence
func (ds *DeviceState) StartWriteSequence(row, col, targetLevel, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	direction := ds.GetWriteDirection(row, col, currentLevel, targetLevel)
	// Initialize voltage calibration if needed
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	ascending := direction == DirectionAscending

	// Ensure voltage calibration is initialized
	ds.initVoltageCalibrationInternal()

	// Skip RESET when staying on the same branch, or after the first ISPP pulse.
	// Only force RESET on direction change (first pulse) or explicit overshoot flag.
	lastDir := DirectionUnknown
	if dir, exists := ds.hysteresisState.Direction[cellKey(row, col)]; exists {
		lastDir = dir
	}
	sameBranch := lastDir == DirectionUnknown || lastDir == direction
	startPhase := PhaseReset
	if !ds.forceResetNextSeq {
		if ds.isppState.Iteration > 0 || sameBranch {
			startPhase = PhaseHold1
		}
	}
	ds.forceResetNextSeq = false

	ds.writeSequenceState.Active = true
	ds.writeSequenceState.Phase = startPhase
	ds.writeSequenceState.TargetRow = row
	ds.writeSequenceState.TargetCol = col
	ds.writeSequenceState.TargetLevel = targetLevel
	ds.writeSequenceState.CurrentLevel = currentLevel
	calibrated := ds.getVoltageForLevelInternal(targetLevel, ascending)
	writeVoltage := calibrated
	if ds.isppState.Active && ds.isppState.Voltage > 0 {
		// Use the current ISPP pulse voltage when available.
		writeVoltage = ds.isppState.Voltage
	}
	applied, _ := ds.dacWriteVoltageLocked(writeVoltage)
	ds.writeSequenceState.WriteVoltage = applied
	if startPhase == PhaseHold1 {
		ds.writeSequenceState.Progress = 0.2
	} else {
		ds.writeSequenceState.Progress = 0.0
	}
	ds.updateWriteSequencePhaseVoltageLocked()
}

// AdvanceWritePhase moves to the next phase in the sequence
// Returns true if sequence is complete
func (ds *DeviceState) AdvanceWritePhase() bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.writeSequenceState.Active {
		return true
	}

	switch ds.writeSequenceState.Phase {
	case PhaseReset:
		ds.writeSequenceState.Phase = PhaseHold1
		ds.writeSequenceState.Progress = 0.2
	case PhaseHold1:
		ds.writeSequenceState.Phase = PhaseWrite
		ds.writeSequenceState.Progress = 0.4
	case PhaseWrite:
		ds.writeSequenceState.Phase = PhaseHold2
		ds.writeSequenceState.Progress = 0.6
	case PhaseHold2:
		ds.writeSequenceState.Phase = PhaseVerify
		ds.writeSequenceState.Progress = 0.8
	case PhaseVerify:
		ds.writeSequenceState.Phase = PhaseIdle
		ds.writeSequenceState.Active = false
		ds.writeSequenceState.Progress = 1.0
		ds.writeSequenceState.PhaseVoltage = 0.0
		return true
	}
	ds.updateWriteSequencePhaseVoltageLocked()
	return false
}

func (ds *DeviceState) updateWriteSequencePhaseVoltageLocked() {
	switch ds.writeSequenceState.Phase {
	case PhaseWrite:
		ds.writeSequenceState.PhaseVoltage = ds.writeSequenceState.WriteVoltage
	case PhaseVerify:
		// Use a safe read voltage for verify (below coercive voltage).
		verifyVoltage := ds.readRange.Max * 0.5
		if verifyVoltage < 0 {
			verifyVoltage = 0
		}
		ds.writeSequenceState.PhaseVoltage = verifyVoltage
	default:
		ds.writeSequenceState.PhaseVoltage = 0.0
	}
}

// GetWritePhaseInfo returns the current write sequence state for UI display
func (ds *DeviceState) GetWritePhaseInfo() WriteSequenceState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.writeSequenceState
}

// CancelWriteSequence aborts the current write sequence
func (ds *DeviceState) CancelWriteSequence() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.writeSequenceState.Active = false
	ds.writeSequenceState.Phase = PhaseIdle
	ds.writeSequenceState.Progress = 0.0
	ds.writeSequenceState.PhaseVoltage = 0.0
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
	case PhaseVerify:
		return "VERIFY"
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
	case PhaseVerify:
		return PhaseVerifyDurationNs
	default:
		return 0
	}
}

// ============================================================================
// 4. ISPP STATE MACHINE WITH OVERSHOOT HANDLING
// ============================================================================

// ISPPEngine selects which ISPP implementation to use.
type ISPPEngine int

const (
	ISPPEngineLevel ISPPEngine = iota // Fast, level-based ISPP (legacy)
	ISPPEngineLK                      // Physics-based ISPP using L-K solver
)

func (e ISPPEngine) String() string {
	switch e {
	case ISPPEngineLevel:
		return "Preisach (Level-based)"
	case ISPPEngineLK:
		return "Landau-Khalatnikov (Physics ODE)"
	default:
		return "Unknown"
	}
}

// ISPPResult represents the result of an ISPP iteration
type ISPPResult int

const (
	ISPPResultContinue      ISPPResult = iota // Continue iterating
	ISPPResultVerified                        // Target level reached
	ISPPResultOvershoot                       // Overshoot detected, reset needed
	ISPPResultMaxIterations                   // Max iterations reached
	ISPPResultNotActive                       // ISPP was not active
)

// ISPP constants
const (
	ISPPMaxIterations = 40 // More pulses for finer convergence (matched to shared/physics)
)

// ISPPHistoryPoint stores one write-verify iteration snapshot.
type ISPPHistoryPoint struct {
	Iteration int
	Level     int
	Voltage   float64
}

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
	History      []ISPPHistoryPoint
}

// StartISPP begins an ISPP loop for a cell
func (ds *DeviceState) StartISPP(row, col, targetLevel, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	// Early exit if already at target level
	if currentLevel == targetLevel {
		ds.isppState.Active = false
		ds.isppState.Iteration = 0
		ds.isppState.TargetLevel = targetLevel
		ds.isppState.CurrentLevel = currentLevel
		ds.isppState.Verified = true
		ds.isppState.Complete = true
		ds.isppState.Success = true
		ds.isppState.History = []ISPPHistoryPoint{{Iteration: 0, Level: currentLevel, Voltage: 0}}
		return
	}

	// Use shared ISPP calculator to determine direction
	sharedDirection := sharedphysics.GetDirection(currentLevel, targetLevel)

	// Map to local HysteresisDirection type
	var localDirection HysteresisDirection
	switch sharedDirection {
	case sharedphysics.DirectionAscending:
		localDirection = DirectionAscending
	case sharedphysics.DirectionDescending:
		localDirection = DirectionDescending
	default:
		localDirection = DirectionUnknown
	}

	// Ensure voltage calibration is initialized
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	// Calculate starting voltage using shared calculator
	ascending := localDirection == DirectionAscending
	calibratedVoltage := ds.getVoltageForLevelInternal(targetLevel, ascending)
	isppCalc := ds.ensureISPPCalculatorLocked()
	startVoltage := isppCalc.CalculateStartVoltage(calibratedVoltage)

	ds.isppState.Active = true
	ds.isppState.Iteration = 0
	ds.isppState.MaxIter = ISPPMaxIterations
	ds.isppState.TargetRow = row
	ds.isppState.TargetCol = col
	ds.isppState.TargetLevel = targetLevel
	ds.isppState.CurrentLevel = currentLevel
	ds.isppState.Voltage = startVoltage
	ds.isppState.Direction = localDirection
	ds.isppState.Verified = false
	ds.isppState.Complete = false
	ds.isppState.Success = false
	ds.isppState.History = []ISPPHistoryPoint{{Iteration: 0, Level: currentLevel, Voltage: startVoltage}}
}

// ISPPIterate performs one write-verify iteration
// Returns the result indicating whether to continue, success, overshoot, or max iterations
func (ds *DeviceState) ISPPIterate(newCurrentLevel int) ISPPResult {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.isppState.Active {
		return ISPPResultNotActive
	}

	ds.isppState.CurrentLevel = newCurrentLevel
	ds.isppState.Iteration++
	ds.isppState.History = append(ds.isppState.History, ISPPHistoryPoint{
		Iteration: ds.isppState.Iteration,
		Level:     ds.isppState.CurrentLevel,
		Voltage:   ds.isppState.Voltage,
	})

	// Map local direction to shared direction type
	var sharedDirection sharedphysics.HysteresisDirection
	switch ds.isppState.Direction {
	case DirectionAscending:
		sharedDirection = sharedphysics.DirectionAscending
	case DirectionDescending:
		sharedDirection = sharedphysics.DirectionDescending
	default:
		sharedDirection = sharedphysics.DirectionUnknown
	}

	// Use shared ISPP calculator to check result
	isppCalc := ds.ensureISPPCalculatorLocked()
	result := isppCalc.CheckResult(
		ds.isppState.CurrentLevel,
		ds.isppState.TargetLevel,
		sharedDirection,
		ds.isppState.Iteration,
	)

	// Map shared result to local result type and update state
	switch result {
	case sharedphysics.ISPPSuccess:
		ds.isppState.Verified = true
		ds.isppState.Complete = true
		ds.isppState.Success = true
		ds.isppState.Active = false
		return ISPPResultVerified

	case sharedphysics.ISPPOvershoot:
		return ISPPResultOvershoot

	case sharedphysics.ISPPMaxPulses:
		ds.isppState.Complete = true
		ds.isppState.Success = false
		ds.isppState.Active = false
		return ISPPResultMaxIterations

	case sharedphysics.ISPPContinue:
		// Calculate next voltage using shared calculator
		ds.isppState.Voltage = isppCalc.CalculateNextVoltage(ds.isppState.Voltage, sharedDirection)
		return ISPPResultContinue

	default:
		return ISPPResultContinue
	}
}

// HandleOvershoot performs RESET-to-saturation when write overshoots target
// Returns true if reset was performed
func (ds *DeviceState) HandleOvershoot(row, col int) bool {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.isppState.Active {
		return false
	}

	// Reset to saturation based on direction
	if ds.isppState.Direction == DirectionAscending {
		// Ascending overshoot: reset to level 0 (negative saturation)
		ds.isppState.CurrentLevel = 0
		ds.isppState.Direction = DirectionAscending // Keep ascending for retry
	} else {
		// Descending overshoot: reset to max level (positive saturation)
		ds.isppState.CurrentLevel = ds.writeRange.NumLevels - 1
		ds.isppState.Direction = DirectionDescending // Keep descending for retry
	}

	// Ensure voltage calibration is initialized
	if ds.voltageCalibration == nil {
		ds.initVoltageCalibrationInternal()
	}

	// Recalculate voltage for target from new position
	ascending := ds.isppState.Direction == DirectionAscending
	calibratedVoltage := ds.getVoltageForLevelInternal(ds.isppState.TargetLevel, ascending)

	// Use shared calculator for starting voltage after reset
	if ds.isppCalc != nil {
		ds.isppState.Voltage = ds.isppCalc.CalculateStartVoltage(calibratedVoltage)
	} else {
		ds.isppState.Voltage = calibratedVoltage
	}

	ds.forceResetNextSeq = true

	return true
}

// GetISPPStatus returns the current ISPP state for UI display
func (ds *DeviceState) GetISPPStatus() ISPPState {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	status := ds.isppState
	if len(ds.isppState.History) > 0 {
		status.History = append([]ISPPHistoryPoint(nil), ds.isppState.History...)
	}
	return status
}

// CancelISPP aborts the current ISPP loop
func (ds *DeviceState) CancelISPP() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.isppState.Active = false
	ds.isppState.Complete = true
	ds.isppState.Success = false
}

func (ds *DeviceState) ensureISPPCalculatorLocked() *sharedphysics.ISPPCalculator {
	if ds.isppCalc != nil {
		return ds.isppCalc
	}
	ec := 1.0
	numLevels := ds.writeRange.NumLevels
	if ds.material != nil {
		ec = ds.material.CoerciveVoltage()
		numLevels = ds.material.GetNumLevels()
	}
	if ec <= 0 {
		ec = 1.0
	}
	if numLevels < 2 {
		numLevels = 30
	}
	ds.isppCalc = sharedphysics.NewISPPCalculator(ec, numLevels)
	return ds.isppCalc
}

func (ds *DeviceState) beginISPPTracking(row, col, targetLevel, currentLevel int, direction HysteresisDirection, maxIter int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if maxIter <= 0 {
		maxIter = ISPPMaxIterations
	}

	ds.isppState.Active = true
	ds.isppState.Iteration = 0
	ds.isppState.MaxIter = maxIter
	ds.isppState.TargetRow = row
	ds.isppState.TargetCol = col
	ds.isppState.TargetLevel = targetLevel
	ds.isppState.CurrentLevel = currentLevel
	ds.isppState.Voltage = 0
	ds.isppState.Direction = direction
	ds.isppState.Verified = false
	ds.isppState.Complete = false
	ds.isppState.Success = false
	ds.isppState.History = []ISPPHistoryPoint{{Iteration: 0, Level: currentLevel, Voltage: 0}}
}

func (ds *DeviceState) updateISPPTracking(iteration int, voltage float64, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.isppState.Iteration = iteration
	ds.isppState.Voltage = voltage
	ds.isppState.CurrentLevel = currentLevel
	ds.isppState.History = append(ds.isppState.History, ISPPHistoryPoint{
		Iteration: iteration,
		Level:     currentLevel,
		Voltage:   voltage,
	})
}

func (ds *DeviceState) endISPPTracking(success bool, currentLevel int) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.isppState.Active = false
	ds.isppState.Complete = true
	ds.isppState.Success = success
	ds.isppState.Verified = success
	ds.isppState.CurrentLevel = currentLevel
}

// ============================================================================
// 5. COLUMN-WRITE VISUALIZATION STATE
// ============================================================================

// HalfSelectVoltageRatio is kept for backward compatibility; in DAC-only column drive
// the column disturb voltage equals the full write voltage (ratio = 1.0 effectively).
// The name is historical; do not rely on this being 0.5 (V/2 scheme is not used).
const HalfSelectVoltageRatio = 0.5

// HalfSelectVisualization holds the state for column-write overlay visualization.
// Despite the name, this implements DAC-only column drive, not a V/2 half-select scheme.
type HalfSelectVisualization struct {
	Enabled        bool
	FullVoltage    float64
	HalfVoltage    float64 // Set equal to FullVoltage in DAC-only mode (no V/2 splitting)
	SelectedRow    int
	SelectedCol    int
	HalfSelectRows []int // Rows disturbed at full voltage (same column — all switch)
	HalfSelectCols []int // Always empty in DAC-only mode (same-row cells see 0V)
}

// EnableHalfSelectVisualization enables the column-write overlay for a passive write operation.
// In DAC-Only Column Drive, all rows in the selected column see full write voltage (disturbed).
// Cells in the same row see 0V (safe — row is grounded, unselected BL is 0V).
func (ds *DeviceState) EnableHalfSelectVisualization(row, col int, fullVoltage float64) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.halfSelectState.Enabled = true
	ds.halfSelectState.FullVoltage = fullVoltage
	// In DAC-Only drive the whole column is disturbed at full voltage; no V/2 partial stress.
	ds.halfSelectState.HalfVoltage = fullVoltage

	ds.halfSelectState.SelectedRow = row
	ds.halfSelectState.SelectedCol = col

	// All other rows in the same column are disturbed (WL=0, BL=-V → ΔV=+V, full disturb)
	ds.halfSelectState.HalfSelectRows = make([]int, 0)
	for r := 0; r < ds.rows; r++ {
		if r != row {
			ds.halfSelectState.HalfSelectRows = append(ds.halfSelectState.HalfSelectRows, r)
		}
	}

	// Same-row cells see 0V (WL=0, BL=0) — no disturb, so HalfSelectCols is empty.
	ds.halfSelectState.HalfSelectCols = make([]int, 0)
}

// DisableHalfSelectVisualization disables the V/2 overlay
func (ds *DeviceState) DisableHalfSelectVisualization() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	ds.halfSelectState.Enabled = false
	ds.halfSelectState.HalfSelectRows = nil
	ds.halfSelectState.HalfSelectCols = nil
}

// GetHalfSelectState returns the current V/2 visualization state
func (ds *DeviceState) GetHalfSelectState() HalfSelectVisualization {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	return ds.halfSelectState
}

// IsHalfSelected returns true if the given cell is in half-select state
func (ds *DeviceState) IsHalfSelected(row, col int) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	if !ds.halfSelectState.Enabled {
		return false
	}

	// Check if in half-select row (same column as selected)
	if col == ds.halfSelectState.SelectedCol {
		for _, r := range ds.halfSelectState.HalfSelectRows {
			if r == row {
				return true
			}
		}
	}

	// Check if in half-select column (same row as selected)
	if row == ds.halfSelectState.SelectedRow {
		for _, c := range ds.halfSelectState.HalfSelectCols {
			if c == col {
				return true
			}
		}
	}

	return false
}
