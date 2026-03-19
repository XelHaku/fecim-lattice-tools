// Package gui — device_state_read.go
// Read-path helpers for DeviceState: conductance mapping, effective cell voltage
// resolution, DAC write-voltage quantization, and DAC nonlinearity application.
package gui

import (
	"math"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

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

// resolveConductanceLevels picks quantization levels used for level->conductance mapping.
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
		levels = sharedphysics.DefaultLevels
	}
	gmin, gmax := ds.conductanceBounds()
	if gmax <= gmin {
		return 0
	}
	gNorm := (gPhys - gmin) / (gmax - gmin)
	return sharedphysics.GetLevel(gNorm, levels)
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
