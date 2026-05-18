package crossbar

import (
	"math"
)

// ThermalConfig configures the compact RC thermal model for crossbar arrays.
//
// The model uses a lumped-element RC network where each cell has a thermal
// resistance Theta_JA (junction-to-ambient) and thermal capacitance C_th.
// Steady-state temperature rise: dT = P * R_th.
// Transient time constant: tau = R_th * C_th.
//
// References:
//   - Cai et al., "Thermal Design of Resistive Crossbar Arrays", IEEE TED (2020)
//   - Li et al., "Thermal Effects in RRAM Crossbar Arrays", DATE (2019)
type ThermalConfig struct {
	AmbientTempK       float64 // Ambient temperature (K), default 300
	ThermalResistance  float64 // Theta_JA: junction-to-ambient (K/W), default 20
	ThermalCapacitance float64 // C_th: thermal capacitance (J/K), default 1e-6
	SubstrateTempK     float64 // Substrate temperature (K), default 300
}

// DefaultThermalConfig returns reasonable defaults for an HZO-based
// ferroelectric crossbar array on a silicon substrate.
//
// Parameters are educational defaults and should be calibrated for
// specific technology nodes and packaging.
func DefaultThermalConfig() ThermalConfig {
	return ThermalConfig{
		AmbientTempK:       300,  // Room temperature
		ThermalResistance:  20,   // 20 K/W typical for thin-film devices
		ThermalCapacitance: 1e-6, // 1 uJ/K typical for nanoscale devices
		SubstrateTempK:     300,  // Room temperature substrate
	}
}

// ThermalState tracks the thermal state of a crossbar array.
type ThermalState struct {
	CellTemperatures [][]float64 // Temperature of each cell (K)
	PeakTempK        float64     // Maximum temperature across array
	AvgTempK         float64     // Average temperature
	PowerDensityWm2  float64     // Current power density (W/m^2)
}

// ThermalModel provides compact thermal simulation for crossbar arrays.
//
// The model implements a first-order RC thermal network where each cell
// is treated as an independent thermal node coupled to the ambient through
// a thermal resistance. This is a simplification that ignores lateral heat
// spreading between cells; for sub-100nm pitches, lateral coupling becomes
// significant and a full thermal FEM would be needed.
type ThermalModel struct {
	config ThermalConfig
	rows   int
	cols   int
	state  ThermalState
}

// NewThermalModel creates a thermal model for the given array dimensions.
func NewThermalModel(rows, cols int, config ThermalConfig) *ThermalModel {
	if rows <= 0 {
		rows = 1
	}
	if cols <= 0 {
		cols = 1
	}

	temps := make([][]float64, rows)
	for i := range temps {
		temps[i] = make([]float64, cols)
		for j := range temps[i] {
			temps[i][j] = config.AmbientTempK
		}
	}

	return &ThermalModel{
		config: config,
		rows:   rows,
		cols:   cols,
		state: ThermalState{
			CellTemperatures: temps,
			PeakTempK:        config.AmbientTempK,
			AvgTempK:         config.AmbientTempK,
			PowerDensityWm2:  0,
		},
	}
}

// ComputeSteadyState computes the steady-state temperature of each cell
// given a power dissipation map (W per cell).
//
// For steady-state RC network: T_cell = T_ambient + P_cell * R_th
//
// powerMap[i][j] is the power dissipated in cell (i,j) in Watts.
// If powerMap dimensions don't match the array, only the overlapping
// region is computed; remaining cells stay at ambient.
func (tm *ThermalModel) ComputeSteadyState(powerMap [][]float64) ThermalState {
	rth := tm.config.ThermalResistance
	tAmb := tm.config.AmbientTempK

	temps := make([][]float64, tm.rows)
	for i := range temps {
		temps[i] = make([]float64, tm.cols)
	}

	peakT := tAmb
	sumT := 0.0
	totalPower := 0.0
	count := 0

	for i := 0; i < tm.rows; i++ {
		for j := 0; j < tm.cols; j++ {
			power := 0.0
			if i < len(powerMap) && j < len(powerMap[i]) {
				power = powerMap[i][j]
				if power < 0 {
					power = 0
				}
			}

			cellT := tAmb + power*rth
			temps[i][j] = cellT
			totalPower += power

			if cellT > peakT {
				peakT = cellT
			}
			sumT += cellT
			count++
		}
	}

	avgT := tAmb
	if count > 0 {
		avgT = sumT / float64(count)
	}

	state := ThermalState{
		CellTemperatures: temps,
		PeakTempK:        peakT,
		AvgTempK:         avgT,
		PowerDensityWm2:  totalPower / float64(tm.rows*tm.cols),
	}
	tm.state = state
	return state
}

// ComputeTransient performs a time-stepping thermal transient simulation.
//
// Starting from the current state (or ambient if first call), it integrates
// the first-order ODE:
//
//	C_th * dT/dt = P - (T - T_ambient) / R_th
//
// using forward Euler with time step dt for the given number of steps.
// Returns a slice of ThermalState snapshots (one per step).
//
// powerMap[i][j] is the constant power dissipated in cell (i,j) in Watts
// during the entire transient.
func (tm *ThermalModel) ComputeTransient(powerMap [][]float64, dt float64, steps int) []ThermalState {
	if steps <= 0 {
		return nil
	}
	if dt <= 0 {
		dt = 1e-9 // Default 1 ns
	}

	rth := tm.config.ThermalResistance
	cth := tm.config.ThermalCapacitance
	tAmb := tm.config.AmbientTempK

	// Initialize from current state.
	temps := make([][]float64, tm.rows)
	for i := range temps {
		temps[i] = make([]float64, tm.cols)
		copy(temps[i], tm.state.CellTemperatures[i])
	}

	snapshots := make([]ThermalState, steps)

	for step := 0; step < steps; step++ {
		peakT := tAmb
		sumT := 0.0
		totalPower := 0.0

		for i := 0; i < tm.rows; i++ {
			for j := 0; j < tm.cols; j++ {
				power := 0.0
				if i < len(powerMap) && j < len(powerMap[i]) {
					power = powerMap[i][j]
					if power < 0 {
						power = 0
					}
				}
				totalPower += power

				// Forward Euler: C_th * dT = (P - (T-Tamb)/Rth) * dt
				heatLoss := (temps[i][j] - tAmb) / rth
				dT := (power - heatLoss) * dt / cth
				temps[i][j] += dT

				// Clamp to ambient (can't go below ambient in this model).
				if temps[i][j] < tAmb {
					temps[i][j] = tAmb
				}

				if temps[i][j] > peakT {
					peakT = temps[i][j]
				}
				sumT += temps[i][j]
			}
		}

		// Snapshot the state.
		snapTemps := make([][]float64, tm.rows)
		for i := range snapTemps {
			snapTemps[i] = make([]float64, tm.cols)
			copy(snapTemps[i], temps[i])
		}

		snapshots[step] = ThermalState{
			CellTemperatures: snapTemps,
			PeakTempK:        peakT,
			AvgTempK:         sumT / float64(tm.rows*tm.cols),
			PowerDensityWm2:  totalPower / float64(tm.rows*tm.cols),
		}
	}

	// Update model state to final snapshot.
	tm.state = snapshots[steps-1]
	return snapshots
}

// PowerFromMVM estimates the power dissipation per cell during a matrix-vector
// multiply operation on the given crossbar array.
//
// For each cell, power is computed as:
//
//	P(i,j) = V(j)^2 * G(i,j)
//
// where V(j) is the input voltage on column j (normalized [0,1] mapped to
// [0, Vread]) and G(i,j) is the cell conductance in Siemens.
//
// The Vread scaling uses a default of 0.2V (typical FeFET read voltage).
// Conductances are taken from the normalized [0,1] values stored in the array,
// mapped to physical range [GMin, GMax].
func (tm *ThermalModel) PowerFromMVM(arr *Array, input []float64) [][]float64 {
	if arr == nil {
		return nil
	}

	const vRead = 0.2 // Default read voltage (V)
	rows := arr.Rows()
	cols := arr.Cols()

	powerMap := make([][]float64, rows)
	for i := range powerMap {
		powerMap[i] = make([]float64, cols)
	}

	gMatrix := arr.GetConductanceMatrix()

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Input voltage for this column.
			v := 0.0
			if j < len(input) {
				v = input[j] * vRead // Normalize to physical voltage
			}

			// Physical conductance from normalized value.
			gNorm := gMatrix[i][j]
			gPhys := GMin + gNorm*(GMax-GMin)

			// P = V^2 * G
			powerMap[i][j] = v * v * gPhys
		}
	}

	return powerMap
}

// TimeConstant returns the thermal time constant tau = R_th * C_th (seconds).
func (tm *ThermalModel) TimeConstant() float64 {
	return tm.config.ThermalResistance * tm.config.ThermalCapacitance
}

// SteadyStateTemp returns the expected steady-state temperature rise for
// a given power dissipation (K).
func SteadyStateTemp(power float64, config ThermalConfig) float64 {
	return config.AmbientTempK + power*config.ThermalResistance
}

// MaxAllowedPower returns the maximum power per cell that keeps the junction
// temperature below maxTempK (in Kelvin).
func MaxAllowedPower(maxTempK float64, config ThermalConfig) float64 {
	dT := maxTempK - config.AmbientTempK
	if dT <= 0 {
		return 0
	}
	return dT / config.ThermalResistance
}

// TransientTemp returns the analytical temperature at time t for a step
// power input P, assuming first-order RC response:
//
//	T(t) = T_amb + P * R_th * (1 - exp(-t/tau))
//
// where tau = R_th * C_th.
func TransientTemp(power, t float64, config ThermalConfig) float64 {
	tau := config.ThermalResistance * config.ThermalCapacitance
	if tau <= 0 {
		return config.AmbientTempK + power*config.ThermalResistance
	}
	return config.AmbientTempK + power*config.ThermalResistance*(1-math.Exp(-t/tau))
}
