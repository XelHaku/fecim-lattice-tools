package crossbar

// Package crossbar — FeCAP (ferroelectric capacitor) crossbar support.
//
// LIT-P2-01/02/04: Capacitance matrix data structure + charge-domain MVM +
// transient pulse energy model.
//
// In FeCAP arrays, each cell is a ferroelectric capacitor rather than a FeFET:
//   - State encoded as capacitance C (not conductance G)
//   - MVM is charge-domain: Q[col] = Σ_row C[row][col] × V[row]
//   - Sneak-path and IR-drop problems are eliminated (capacitor blocks DC)
//   - Sensing uses a charge amplifier (not TIA)
//
// Reference: Adv. Intell. Syst. 2022, 128×128 FeCAP demo, 3.8 pJ/MVM.

import (
	"fmt"
	"math"
	"sync/atomic"
)

// CellType identifies the operating mode of the crossbar array.
type CellType int

const (
	// CellTypeFeFET is the default transistor-gated mode; computation is
	// current-domain: I = G × V.
	CellTypeFeFET CellType = iota
	// CellTypeFeCAP is the capacitive mode; computation is charge-domain:
	// Q = C × V. Sneak paths and IR drop are absent (capacitor blocks DC).
	CellTypeFeCAP
)

// CapacitanceModel describes how cell capacitance scales with normalized state.
type CapacitanceModel int

const (
	// CapModelLinear scales linearly: C = Cmin + cNorm*(Cmax-Cmin).
	CapModelLinear CapacitanceModel = iota
	// CapModelExponential scales log-linearly: C = Cmin * exp(ln(Cmax/Cmin)*cNorm).
	CapModelExponential
)

// FeCAP defaults based on HZO-stack literature (Adv. Intell. Syst. 2022).
const (
	// DefaultCMin is the minimum cell capacitance (LRS-equivalent, cNorm=0).
	// ~0.5 fF for a 0.1 µm² HZO cell at 8 nm thickness, εr≈25.
	DefaultCMin = 0.5e-15 // 0.5 fF

	// DefaultCMax is the maximum cell capacitance (HRS-equivalent, cNorm=1).
	// ~2 fF, ×4 modulation ratio observed in literature.
	DefaultCMax = 2.0e-15 // 2.0 fF

	// DefaultFeCAPPulseDuration is the word-line pulse width for charge transfer.
	// 10 ns enables reliable switching and read in 28 nm CMOS back-end.
	DefaultFeCAPPulseDuration = 10e-9 // 10 ns

	// DefaultCfb is the charge amplifier feedback capacitance used by
	// MVMChargeQuantized. Sized so that Cmax×Vmax / Cfb ≈ 1 V (ADC full scale).
	// With Cmax=2 fF and Vmax=1 V: Cfb = Rows*Cmax*1V/1V = Rows*2 fF.
	// Using 64*2 fF = 128 fF as a 64-row default.
	DefaultCfb = 128e-15 // 128 fF
)

// DefaultFeCAPConfig returns a Config pre-filled for FeCAP operation.
// Rows, Cols, ADCBits, and DACBits should be set by the caller.
func DefaultFeCAPConfig(rows, cols int) *Config {
	return &Config{
		Rows:             rows,
		Cols:             cols,
		ADCBits:          4, // 4-bit per literature consensus
		DACBits:          4, // 4-bit per literature consensus
		CellType:         CellTypeFeCAP,
		CMin:             DefaultCMin,
		CMax:             DefaultCMax,
		PulseDuration:    DefaultFeCAPPulseDuration,
		CapacitanceModel: CapModelLinear,
		Endurance:        DefaultEnduranceConfig(),
	}
}

// GetPhysicalCapacitance maps a normalized capacitance state (0–1) to a
// physical capacitance in Farads using the array's CapacitanceModel.
func (a *Array) GetPhysicalCapacitance(cNorm float64) float64 {
	if cNorm < 0 {
		cNorm = 0
	}
	if cNorm > 1 {
		cNorm = 1
	}
	cMin := a.config.CMin
	cMax := a.config.CMax
	if cMin <= 0 {
		cMin = DefaultCMin
	}
	if cMax <= 0 {
		cMax = DefaultCMax
	}
	switch a.config.CapacitanceModel {
	case CapModelExponential:
		if cMin <= 0 {
			return cMax
		}
		return cMin * math.Exp(math.Log(cMax/cMin)*cNorm)
	default: // CapModelLinear
		return cMin + cNorm*(cMax-cMin)
	}
}

// GetPhysicalCapacitanceForCell returns the effective physical capacitance of
// cell (row, col) accounting for per-cell noise variation.
func (a *Array) GetPhysicalCapacitanceForCell(row, col int) float64 {
	cell := a.cells[row][col]
	cPhys := a.GetPhysicalCapacitance(cell.Capacitance)
	// Apply the same per-cell noise factor as conductance cells.
	if cell.NoiseFactor != 0 {
		cPhys *= (1 + cell.NoiseFactor*a.config.NoiseLevel)
	}
	return cPhys
}

// ProgramCapacitance writes a normalized capacitance weight (0–1) to cell
// (row, col). The caller is responsible for translating from a logical weight
// to a normalized capacitance using the array's CMin/CMax range.
func (a *Array) ProgramCapacitance(row, col int, weight float64) error {
	if row < 0 || row >= a.config.Rows {
		return fmt.Errorf("row %d out of range [0, %d)", row, a.config.Rows)
	}
	if col < 0 || col >= a.config.Cols {
		return fmt.Errorf("col %d out of range [0, %d)", col, a.config.Cols)
	}
	if weight < 0 {
		weight = 0
	}
	if weight > 1 {
		weight = 1
	}
	a.cells[row][col].Capacitance = weight
	a.cells[row][col].SwitchingCount++
	atomic.AddInt64(&a.totalWrites, 1)
	getLog().Input("ProgramCapacitance", map[string]interface{}{
		"row": row, "col": col, "weight": weight,
	})
	return nil
}

// ProgramCapacitanceMatrix writes an entire capacitance weight matrix (rows×cols,
// values in [0,1]) to the array. Returns the first error encountered.
func (a *Array) ProgramCapacitanceMatrix(weights [][]float64) error {
	for r, row := range weights {
		for c, w := range row {
			if err := a.ProgramCapacitance(r, c, w); err != nil {
				return err
			}
		}
	}
	return nil
}

// GetCapacitanceMatrix returns the normalized capacitance state (0–1) of every
// cell as a rows×cols slice.
func (a *Array) GetCapacitanceMatrix() [][]float64 {
	mat := make([][]float64, a.config.Rows)
	for r := range mat {
		mat[r] = make([]float64, a.config.Cols)
		for c := range mat[r] {
			mat[r][c] = a.cells[r][c].Capacitance
		}
	}
	return mat
}

// GetEffectiveCapacitanceMatrix returns the physical capacitance (Farads) of
// every cell, including noise variation.
func (a *Array) GetEffectiveCapacitanceMatrix() [][]float64 {
	mat := make([][]float64, a.config.Rows)
	for r := range mat {
		mat[r] = make([]float64, a.config.Cols)
		for c := range mat[r] {
			mat[r][c] = a.GetPhysicalCapacitanceForCell(r, c)
		}
	}
	return mat
}

// MVMCharge performs a charge-domain matrix-vector multiply for FeCAP arrays.
//
// Each row voltage is DAC-quantized then multiplied by the physical capacitance
// of each cell. The resulting per-column charge accumulation is:
//
//	Q[j] = Σ_i  C_phys[i][j] × V_dac[i]
//
// The raw charge vector is returned (Coulombs). Use a ChargeAmplifier to
// convert to voltage for ADC quantization, or call MVMChargeQuantized for
// the full read chain.
//
// Returns an error if the array is configured in FeFET (conductance) mode or
// if the input length does not match the number of rows.
func (a *Array) MVMCharge(input []float64) ([]float64, error) {
	if a.config.CellType != CellTypeFeCAP {
		return nil, fmt.Errorf("MVMCharge requires CellTypeFeCAP; got CellTypeFeFET — use MVM instead")
	}
	if len(input) != a.config.Rows {
		return nil, fmt.Errorf("input length %d != rows %d", len(input), a.config.Rows)
	}

	charge := make([]float64, a.config.Cols)
	for r, v := range input {
		vDAC := a.quantizeDAC(v)
		for c := range charge {
			charge[c] += a.GetPhysicalCapacitanceForCell(r, c) * vDAC
		}
	}

	atomic.AddInt64(&a.totalReads, 1)
	getLog().Calculation("MVMCharge", map[string]interface{}{
		"rows": a.config.Rows, "cols": a.config.Cols,
	}, charge)
	return charge, nil
}

// MVMChargeQuantized runs the full FeCAP read chain:
//  1. DAC quantize each row input voltage.
//  2. Charge accumulation: Q[j] = Σ_i C[i][j] × V_dac[i].
//  3. Charge-to-voltage via charge amplifier: V_out = Q / Cfb.
//  4. ADC quantize each column output voltage.
//
// cfb is the charge amplifier feedback capacitance (F). Use
// DefaultChargeAmplifier().Cfb as a sensible default.
func (a *Array) MVMChargeQuantized(input []float64, cfb float64) ([]float64, error) {
	charge, err := a.MVMCharge(input)
	if err != nil {
		return nil, err
	}
	if cfb <= 0 {
		cfb = DefaultCfb
	}
	result := make([]float64, len(charge))
	for j, q := range charge {
		vOut := q / cfb
		result[j] = a.quantizeADC(vOut)
	}
	return result, nil
}

// MVMChargeEnergy returns the total capacitor charging energy (Joules) for
// a single FeCAP MVM read with the given row input voltages.
//
// Per-cell energy: E_cell = ½ × C_phys × V²
// This is the energy drawn from the word-line driver during the charge phase.
func (a *Array) MVMChargeEnergy(input []float64) float64 {
	if len(input) != a.config.Rows {
		return 0
	}
	total := 0.0
	for r, v := range input {
		vDAC := a.quantizeDAC(v)
		for c := 0; c < a.config.Cols; c++ {
			cPhys := a.GetPhysicalCapacitanceForCell(r, c)
			total += 0.5 * cPhys * vDAC * vDAC
		}
	}
	return total
}
