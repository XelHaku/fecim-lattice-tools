//go:build legacy_fyne

// Package gui — device_state_compute.go
// Compute (MVM) path for DeviceState: ideal and coupled-array simulation,
// TIA/ADC sense-chain conversion, and coupled-voltage level programming.
package gui

import (
	"math"

	"fecim-lattice-tools/module4-circuits/pkg/arraysim"
	sharedphysics "fecim-lattice-tools/shared/physics"
)

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

			if math.Abs(voltage) < 1e-9 {
				continue
			}

			// Get cell conductance from weight using material physics model
			level := 0
			if r < len(weights) && c < len(weights[r]) {
				level = weights[r][c]
			}

			// Use shared level->conductance mapping so ideal and coupled paths
			// respect the same material + geometry scaling dependencies.
			conductanceS := ds.levelToConductance(level, quantLevels)

			// Convert to uS for current calculation.
			conductanceUS := conductanceS * 1e6
			current := conductanceUS * voltage // I = G * V (signed uA since G is in uS)
			totalCurrent += current
		}

		ds.rowCurrents[r] = totalCurrent

		if ds.tia == nil {
			ds.rowVoltages[r] = 0
			ds.rowLevels[r] = 0
			ds.saturated[r] = false
			continue
		}

		currentA := totalCurrent * 1e-6 // uA to A

		// Count active columns (columns with |effective voltage| > 0.01V)
		activeColCount := 0
		for c := 0; c < ds.cols; c++ {
			v := ds.effectiveCellVoltageLocked(r, c)
			if math.Abs(v) > 1e-9 {
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
			if math.Abs(vcell) > 1e-9 {
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
	currentP := sharedphysics.ConductanceToPolarizationWithParams(currentG, mat.Ps, gmin, gmax, sharedphysics.ParseConductanceModel(mat.ConductanceModel), mat.KvT, mat.VGSReadV, mat.VT0V)

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
