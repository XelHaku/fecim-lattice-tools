//go:build legacy_fyne

// Package gui — device_state_config.go
// Configuration setters and getters for DeviceState: material, coupling,
// peripheral (PVT, DAC/ADC bits), cell geometry, wire params, and ISPP engine.
// DAC voltage presets and WL/BL mode selection live in device_state_dac.go.
package gui

import (
	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	"fecim-lattice-tools/shared/peripherals"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

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
	return 6 // Default: 6-bit (64 codes) to resolve 30 conductance levels
}

// GetADCLevels returns the number of ADC output levels (2^bits)
func (ds *DeviceState) GetADCLevels() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()
	if ds.adc != nil {
		return 1 << ds.adc.Bits
	}
	return 64 // Default 6-bit
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
