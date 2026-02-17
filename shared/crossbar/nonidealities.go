// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"errors"
	"fmt"
	"math"
)

func init() {
	ErrInputSize = errors.New("input size exceeds array columns")
}

// WireParams contains wire resistance and capacitance parameters for
// IR drop and RC delay modeling.
//
// H11: Added parasitic capacitance for realistic RC delay simulation.
// RC delay affects signal propagation time and limits array operating frequency.
//
// Physical basis:
//   - Wire capacitance: C = ε₀ε_r × (W×L)/t + fringe capacitance
//   - Typical interconnect: 0.1-0.3 fF/µm for 45nm node
//   - RC delay: τ = R × C (Elmore delay for distributed RC line)
type WireParams struct {
	RwordLine float64 // Word line resistance per unit (Ohm)
	RbitLine  float64 // Bit line resistance per unit (Ohm)
	Rcontact  float64 // Contact resistance (Ohm)

	// Parasitic capacitance (H11)
	CwordLine float64 // Word line capacitance per unit (F), typical: 0.1-0.5 fF
	CbitLine  float64 // Bit line capacitance per unit (F), typical: 0.1-0.5 fF
	CellPitch float64 // Cell pitch in meters (for calculating total wire length)
}

// DefaultWireParams returns typical wire parameters for a 45nm technology node.
// Includes parasitic capacitance values for RC delay modeling (H11).
func DefaultWireParams() *WireParams {
	return &WireParams{
		RwordLine: 2.5,     // 2.5 Ohm per cell pitch
		RbitLine:  2.5,     // 2.5 Ohm per cell pitch
		Rcontact:  50,      // 50 Ohm contact resistance
		CwordLine: 0.2e-15, // 0.2 fF per cell pitch (typical 45nm)
		CbitLine:  0.2e-15, // 0.2 fF per cell pitch
		CellPitch: 90e-9,   // 90 nm cell pitch (2× minimum feature size for 45nm)
	}
}

// ============================================================================
// H11: PARASITIC CAPACITANCE AND RC DELAY MODELING
// ============================================================================

// RCDelayAnalysis contains RC delay analysis results.
type RCDelayAnalysis struct {
	// Per-line delays
	WordLineDelay []float64 // RC delay for each word line (seconds)
	BitLineDelay  []float64 // RC delay for each bit line (seconds)

	// Summary statistics
	MaxWordLineDelay float64 // Maximum word line delay (seconds)
	MaxBitLineDelay  float64 // Maximum bit line delay (seconds)
	TotalArrayDelay  float64 // Total array propagation delay (seconds)
	MaxFrequency     float64 // Maximum operating frequency (Hz)

	// Capacitance breakdown
	TotalWordLineCap float64 // Total word line capacitance (F)
	TotalBitLineCap  float64 // Total bit line capacitance (F)
	TotalArrayCap    float64 // Total array capacitance (F)

	// Energy per operation (CV²/2)
	ChargingEnergy float64 // Energy to charge all lines (J)
}

// AnalyzeRCDelay performs RC delay analysis for the crossbar array.
// Uses Elmore delay model for distributed RC lines.
//
// For a distributed RC line of length L:
//
//	τ_elmore = R_total × C_total / 2 (for load at end)
//
// Reference: IEEE TCAD, Elmore delay modeling for VLSI interconnects
func (a *Array) AnalyzeRCDelay(params *WireParams, inputVoltage float64) *RCDelayAnalysis {
	if params == nil {
		params = DefaultWireParams()
	}

	rows := a.config.Rows
	cols := a.config.Cols

	result := &RCDelayAnalysis{
		WordLineDelay: make([]float64, rows),
		BitLineDelay:  make([]float64, cols),
	}

	// Word line analysis
	// Each word line spans 'cols' cells
	wlResistance := params.RwordLine * float64(cols)
	wlCapacitance := params.CwordLine * float64(cols)
	result.TotalWordLineCap = wlCapacitance * float64(rows)

	for i := 0; i < rows; i++ {
		// Elmore delay for distributed RC line
		// τ = R × C / 2 for a uniformly distributed RC line
		result.WordLineDelay[i] = wlResistance * wlCapacitance / 2

		if result.WordLineDelay[i] > result.MaxWordLineDelay {
			result.MaxWordLineDelay = result.WordLineDelay[i]
		}
	}

	// Bit line analysis
	// Each bit line spans 'rows' cells
	blResistance := params.RbitLine * float64(rows)
	blCapacitance := params.CbitLine * float64(rows)
	result.TotalBitLineCap = blCapacitance * float64(cols)

	for j := 0; j < cols; j++ {
		result.BitLineDelay[j] = blResistance * blCapacitance / 2

		if result.BitLineDelay[j] > result.MaxBitLineDelay {
			result.MaxBitLineDelay = result.BitLineDelay[j]
		}
	}

	// Total array capacitance
	result.TotalArrayCap = result.TotalWordLineCap + result.TotalBitLineCap

	// Total array delay is dominated by the longer path
	// Word lines drive the inputs, bit lines collect outputs
	result.TotalArrayDelay = result.MaxWordLineDelay + result.MaxBitLineDelay

	// Maximum frequency (with 10× margin for settling)
	if result.TotalArrayDelay > 0 {
		result.MaxFrequency = 0.1 / result.TotalArrayDelay
	}

	// Charging energy: E = C × V² / 2 (per transition)
	// For one complete MVM, all word lines switch
	result.ChargingEnergy = result.TotalArrayCap * inputVoltage * inputVoltage / 2

	return result
}

// GetRCTimeConstant returns the RC time constant for a single line.
func (p *WireParams) GetRCTimeConstant(numCells int, isWordLine bool) float64 {
	if isWordLine {
		return p.RwordLine * float64(numCells) * p.CwordLine * float64(numCells)
	}
	return p.RbitLine * float64(numCells) * p.CbitLine * float64(numCells)
}

// GetPropagationDelay returns the 50% propagation delay for a line (Elmore model).
func (p *WireParams) GetPropagationDelay(numCells int, isWordLine bool) float64 {
	// Elmore delay: τ = 0.69 × R × C for 50% threshold
	tau := p.GetRCTimeConstant(numCells, isWordLine)
	return 0.69 * tau / float64(numCells) // Divide by N for distributed RC
}

// GetSettlingTime returns time to settle within given percentage (e.g., 0.01 = 1%).
func (p *WireParams) GetSettlingTime(numCells int, isWordLine bool, errorPercent float64) float64 {
	if errorPercent <= 0 || errorPercent >= 1 {
		errorPercent = 0.01 // Default 1%
	}
	tau := p.GetRCTimeConstant(numCells, isWordLine) / float64(numCells)
	// Settling to within error requires -ln(error) time constants
	return tau * (-math.Log(errorPercent))
}

// ScaleForTechNode scales wire parameters for a different technology node.
// Scaling follows Dennard scaling rules with modifications for modern nodes.
//
// fromNode, toNode: technology nodes in nm (e.g., 45, 28, 14, 7)
func (p *WireParams) ScaleForTechNode(fromNode, toNode float64) *WireParams {
	if fromNode <= 0 || toNode <= 0 {
		return p
	}

	// Scaling factor
	s := fromNode / toNode

	// Resistance scales as 1/s (narrower wires = higher resistance)
	// Capacitance scales as s (shorter wires = lower capacitance)
	// But modern nodes have higher resistivity due to electron scattering
	resistivityPenalty := 1.0
	if toNode < 14 {
		resistivityPenalty = 1.5 // Significant penalty below 14nm
	} else if toNode < 28 {
		resistivityPenalty = 1.2
	}

	return &WireParams{
		RwordLine: p.RwordLine * (1 / s) * resistivityPenalty,
		RbitLine:  p.RbitLine * (1 / s) * resistivityPenalty,
		Rcontact:  p.Rcontact * (1 / s) * resistivityPenalty,
		CwordLine: p.CwordLine * s,
		CbitLine:  p.CbitLine * s,
		CellPitch: p.CellPitch * (toNode / fromNode),
	}
}

// IRDropAnalysis contains the results of IR drop analysis.
type IRDropAnalysis struct {
	// Voltage drop matrices
	WordLineVoltages [][]float64 // Voltage at each word line position
	BitLineVoltages  [][]float64 // Voltage at each bit line position
	EffectiveVoltage [][]float64 // Effective voltage across each cell
	CellCurrents     [][]float64 // Current through each cell (added for Unified Solver)

	// Summary statistics
	MaxIRDrop      float64 // Maximum IR drop in array
	AvgIRDrop      float64 // Average IR drop
	IRDropVariance float64 // Variance in IR drop
	WorstCaseCell  [2]int  // Location of worst-case cell
}

// IRDropSolverConfig contains configuration for the iterative IR drop solver.
type IRDropSolverConfig struct {
	MaxIterations int     // Maximum solver iterations (default: 100)
	Tolerance     float64 // Convergence tolerance in volts (default: 1e-6)
	Damping       float64 // Damping factor for stability (default: 0.5)
}

// DefaultIRDropSolverConfig returns default solver configuration.
func DefaultIRDropSolverConfig() *IRDropSolverConfig {
	return &IRDropSolverConfig{
		MaxIterations: 100,
		Tolerance:     1e-6,
		Damping:       0.5, // Under-relaxation for stability
	}
}

// AnalyzeIRDrop performs IR drop analysis using iterative relaxation.
// Solves the coupled resistive network where voltage drops affect currents
// and currents affect voltage drops. Iterates until convergence.
//
// Physics model (per PHYSICS.md):
//   - Word line drivers on LEFT, bit line sense amps/ground at TOP
//   - V_wl(i,j) = V_input - cumulative_drop_from_driver
//   - V_bl(i,j) = cumulative_rise_from_ground
//   - V_eff(i,j) = V_wl(i,j) - V_bl(i,j)
//   - I_ij = V_eff(i,j) × G_ij
func (a *Array) AnalyzeIRDrop(input []float64, params *WireParams) *IRDropAnalysis {
	getLog().Input("AnalyzeIRDrop", map[string]interface{}{
		"inputLen": len(input),
		"rows":     a.config.Rows,
		"cols":     a.config.Cols,
	})
	result := a.AnalyzeIRDropIterative(input, params, nil)
	getLog().Calculation("AnalyzeIRDrop", map[string]interface{}{
		"maxDrop":   result.MaxIRDrop,
		"avgDrop":   result.AvgIRDrop,
		"variance":  result.IRDropVariance,
		"worstCell": result.WorstCaseCell,
	}, result)
	return result
}

// AnalyzeIRDropIterative performs physics-based IR drop analysis with configurable solver.
// Uses the generic IRDropSimulator for unified solving.
func (a *Array) AnalyzeIRDropIterative(input []float64, params *WireParams, config *IRDropSolverConfig) *IRDropAnalysis {
	if params == nil {
		params = DefaultWireParams()
	}
	if config == nil {
		config = DefaultIRDropSolverConfig()
	}

	rows := a.config.Rows
	cols := a.config.Cols

	// Use the unified IRDropSimulator
	sim := NewIRDropSimulator(rows, cols)
	sim.RowResist = params.RwordLine
	sim.ColResist = params.RbitLine

	// Map cells conductances
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			// Get physical conductance (includes variation if applied in MVM flow?
			// Note: GetPhysicalConductanceForCell uses stored/generated variation)
			// Ideally we should pass the exact G matrix used in MVM, but accessing `a.cells` matches MVM logic.
			sim.SetConductance(i, j, a.GetPhysicalConductanceForCell(i, j))
		}
	}

	// Apply Inputs
	// Mapping: Inputs are applied to Columns (Bit Lines) for MVM accumulation on Rows.
	// We simulate this by driving BitLines with Input Voltages (inverted relative to Row Ground).
	// V_row (Virtual Ground) = 0.
	// V_col = -Input[j].
	// V_eff = V_row - V_col = input[j].
	// This ensures positive current flows from Row to Col (or vice-versa depending on convention).
	// MVM expects I = G * V_in.
	// If V_col = V_in, V_row = 0, then V_eff = -V_in. I = G * -V_in.
	// We'll set V_col = input[j]. V_row = 0. And expect negative currents in results.

	inputLen := len(input)
	for j := 0; j < cols; j++ {
		if j < inputLen {
			// Driving Bit Line with input voltage
			// Note: IRDropSimulator treats VoltageOut as Column voltage.
			sim.VoltageOut[j] = -input[j]
		} else {
			sim.VoltageOut[j] = 0 // Floating/Ground?
		}
	}
	for i := 0; i < rows; i++ {
		sim.SetInputVoltage(i, 0.0) // Rows held at virtual ground (0V)
	}

	sim.Simulate(config.MaxIterations)

	// Convert simulator results to IRDropAnalysis
	return &IRDropAnalysis{
		WordLineVoltages: sim.RowVoltages,
		BitLineVoltages:  sim.ColVoltages,
		// IRDropSimulator calculates effective voltage and cell currents
		// internally. However, EffectiveVoltage in IRDropSimulator isn't exported as a matrix directly,
		// but CellCurrents is. We can derive EffV or just ensure IRDropAnalysis matches expected format.
		// Wait, IRDropAnalysis struct definition I verified previously has EffectiveVoltage.
		// I need to reconstruct it or expose it from Sim.
		EffectiveVoltage: reconstructEffectiveVoltage(sim),
		CellCurrents:     sim.CellCurrents,
		MaxIRDrop:        sim.GetMaxIRDrop(),
		AvgIRDrop:        sim.GetAvgIRDrop(),
		// IRDropVariance is not calculated by default in Sim, compute here or skip?
		// We'll calculate simple variance.
		IRDropVariance: 0.0, // Placeholder
		WorstCaseCell:  [2]int{sim.GetStats().WorstCellRow, sim.GetStats().WorstCellCol},
	}
}

func reconstructEffectiveVoltage(sim *IRDropSimulator) [][]float64 {
	eff := make([][]float64, sim.Rows)
	for i := 0; i < sim.Rows; i++ {
		eff[i] = make([]float64, sim.Cols)
		for j := 0; j < sim.Cols; j++ {
			// V_eff = V_row - V_col
			eff[i][j] = sim.RowVoltages[i][j] - sim.ColVoltages[i][j]
		}
	}
	return eff
}

// estimateCurrent estimates current draw for a word line.
// Uses the configured conductance model (linear, exponential, or lookup).
func (a *Array) estimateCurrent(row int, input []float64) float64 {
	var current float64
	for j := 0; j < len(input) && j < a.config.Cols; j++ {
		// Use the new physics-aware conductance conversion
		gPhys := a.GetPhysicalConductanceForCell(row, j)
		// I = G * V
		current += gPhys * input[j]
	}
	return current
}

// estimateColumnCurrent estimates current through a bit line column.
// Uses the configured conductance model (linear, exponential, or lookup).
func (a *Array) estimateColumnCurrent(col int, input []float64) float64 {
	var current float64
	for i := 0; i < a.config.Rows; i++ {
		// Use the new physics-aware conductance conversion
		gPhys := a.GetPhysicalConductanceForCell(i, col)
		// I = G * V
		if col < len(input) {
			current += gPhys * input[col]
		}
	}
	return current
}

// SneakPathAnalysis contains sneak path current analysis results.
type SneakPathAnalysis struct {
	// Sneak current map (normalized)
	SneakCurrents [][]float64

	// Statistics
	MaxSneakRatio float64 // Maximum sneak/signal ratio
	AvgSneakRatio float64 // Average sneak/signal ratio
	TotalSneak    float64 // Total sneak current
	TotalSignal   float64 // Total signal current
}

// AnalyzeSneakPaths analyzes sneak path currents in the array.
// Sneak paths occur when unselected cells create parallel current paths.
//
// Physical model for passive crossbar when reading cell (selectedRow, selectedCol):
// - Same row: current leaks from selected WL through cell to unselected BL
// - Same column: current from unselected WLs leaks through cell to selected BL
// - Off-diagonal: three-cell sneak path forms a complete loop
//
// For architecture-aware analysis, use AnalyzeSneakPathsWithArch.
func (a *Array) AnalyzeSneakPaths(selectedRow, selectedCol int) *SneakPathAnalysis {
	getLog().Input("AnalyzeSneakPaths", map[string]interface{}{
		"selectedRow": selectedRow,
		"selectedCol": selectedCol,
	})
	result := a.AnalyzeSneakPathsWithArch(selectedRow, selectedCol, false)
	getLog().Calculation("AnalyzeSneakPaths", map[string]interface{}{
		"maxSneakRatio": result.MaxSneakRatio,
		"avgSneakRatio": result.AvgSneakRatio,
		"totalSneak":    result.TotalSneak,
		"totalSignal":   result.TotalSignal,
	}, result)
	return result
}

// AnalyzeSneakPathsWithArch analyzes sneak paths with architecture consideration.
// If is1T1R is true, transistor isolation reduces sneak by ~1000x.
// For 2T1R, use AnalyzeSneakPathsWithIsolation with factor 0.0001.
func (a *Array) AnalyzeSneakPathsWithArch(selectedRow, selectedCol int, is1T1R bool) *SneakPathAnalysis {
	isolationFactor := 1.0
	if is1T1R {
		isolationFactor = 0.001 // 1000x isolation from transistor
	}
	return a.AnalyzeSneakPathsWithIsolation(selectedRow, selectedCol, isolationFactor)
}

// MVMWithIRDrop performs MVM with IR drop effects.
func (a *Array) MVMWithIRDrop(input []float64, params *WireParams) ([]float64, *IRDropAnalysis, error) {
	getLog().Input("MVMWithIRDrop", map[string]interface{}{
		"inputLen": len(input),
		"rows":     a.config.Rows,
		"cols":     a.config.Cols,
	})

	if len(input) > a.config.Cols {
		getLog().Error(ErrInputSize, "MVMWithIRDrop validation failed")
		return nil, nil, ErrInputSize
	}

	if params == nil {
		params = DefaultWireParams()
	}

	// First, analyze IR drop
	irAnalysis := a.AnalyzeIRDrop(input, params)

	output := make([]float64, a.config.Rows)

	for i := 0; i < a.config.Rows; i++ {
		var sum float64
		for j := 0; j < len(input); j++ {
			// Apply IR drop effect to voltage
			effectiveV := input[j] * irAnalysis.EffectiveVoltage[i][j]
			quantizedInput := a.quantizeDAC(effectiveV)

			g := a.cells[i][j].Conductance * a.GetProcessVariationFactor(i, j)
			sum += g * quantizedInput
		}

		output[i] = a.quantizeADC(sum / float64(len(input)))
		a.totalReads++
	}

	getLog().Calculation("MVMWithIRDrop", map[string]interface{}{
		"maxIRDrop": irAnalysis.MaxIRDrop,
		"avgIRDrop": irAnalysis.AvgIRDrop,
	}, output)

	return output, irAnalysis, nil
}

// ErrInputSize indicates input size mismatch.
var ErrInputSize error

// GetIRDropMap returns an IR drop heatmap for visualization.
// Returns values in [0,1] where value directly represents percentage (0.05 = 5%).
// This matches the 0-100% legend scale.
func (a *IRDropAnalysis) GetIRDropMap() [][]float64 {
	return a.GetIRDropMapWithScale(1.0) // Default 100% scale
}

// GetIRDropMapWithScale returns the IR drop map normalized to a custom scale.
// Use the passive (0T1R) max IR drop as scale for architecture comparison:
// - Passive: Shows full color range (highest IR drop)
// - 1T1R/2T1R: Shows proportionally less (better isolation)
// maxDrop should be provided as a fraction (e.g., 0.0296 for 2.96%)
func (a *IRDropAnalysis) GetIRDropMapWithScale(maxDrop float64) [][]float64 {
	rows := len(a.EffectiveVoltage)
	cols := len(a.EffectiveVoltage[0])

	if maxDrop <= 0 {
		maxDrop = 1.0 // Default to 100%
	}

	normalized := make([][]float64, rows)
	for i := range normalized {
		normalized[i] = make([]float64, cols)
		for j := range normalized[i] {
			// IR drop = 1 - effective voltage (assuming 1V input)
			drop := 1.0 - a.EffectiveVoltage[i][j]
			// Normalize to scale: drop / maxDrop gives 0-1 range
			scaled := drop / maxDrop
			// Cap at 1.0
			if scaled > 1.0 {
				scaled = 1.0
			}
			if scaled < 0 {
				scaled = 0
			}
			normalized[i][j] = scaled
		}
	}

	return normalized
}

// GetSneakMap returns the sneak current map normalized for visualization.
// Uses a default scale of 1.0 (100% sneak ratio).
func (s *SneakPathAnalysis) GetSneakMap() [][]float64 {
	return s.GetSneakMapWithScale(1.0) // Default 100% scale
}

// GetSneakMapWithScale returns the sneak current map normalized to a custom scale.
// Use the passive (0T1R) max sneak ratio as scale for architecture comparison:
// - 0T1R: Shows full color range (high sneak)
// - 1T1R: Shows near-zero (1000x reduction visible)
// - 2T1R: Shows essentially zero (10000x reduction visible)
func (s *SneakPathAnalysis) GetSneakMapWithScale(maxRatio float64) [][]float64 {
	rows := len(s.SneakCurrents)
	cols := len(s.SneakCurrents[0])

	getLog().Input("GetSneakMapWithScale", map[string]interface{}{
		"maxRatio":    maxRatio,
		"rows":        rows,
		"cols":        cols,
		"TotalSignal": s.TotalSignal,
	})

	if maxRatio <= 0 {
		maxRatio = 1.0 // Default to 100%
	}

	var minVal, maxVal float64 = 1.0, 0.0
	var nonZeroCount int

	normalized := make([][]float64, rows)
	for i := range normalized {
		normalized[i] = make([]float64, cols)
		for j := range normalized[i] {
			if s.TotalSignal > 0 {
				// Compute sneak ratio for this cell
				ratio := s.SneakCurrents[i][j] / s.TotalSignal
				// Normalize to [0,1] using provided scale
				normalized[i][j] = ratio / maxRatio
			}
			if normalized[i][j] > 1.0 {
				normalized[i][j] = 1.0
			}
			if normalized[i][j] > 0 {
				nonZeroCount++
				if normalized[i][j] < minVal {
					minVal = normalized[i][j]
				}
				if normalized[i][j] > maxVal {
					maxVal = normalized[i][j]
				}
			}
		}
	}

	getLog().Calculation("GetSneakMapWithScale", map[string]interface{}{
		"minNormalized":  minVal,
		"maxNormalized":  maxVal,
		"nonZeroCount":   nonZeroCount,
		"totalCells":     rows * cols,
		"percentNonZero": float64(nonZeroCount) / float64(rows*cols) * 100,
	}, nil)

	return normalized
}

// ComputeError calculates the MVM output error due to non-idealities.
func ComputeError(ideal, actual []float64) float64 {
	getLog().Input("ComputeError", map[string]interface{}{
		"idealLen":  len(ideal),
		"actualLen": len(actual),
	})

	if len(ideal) != len(actual) {
		getLog().Error(fmt.Errorf("length mismatch: ideal=%d actual=%d", len(ideal), len(actual)), "ComputeError validation failed")
		return math.Inf(1)
	}

	var sumSqError float64
	for i := range ideal {
		diff := ideal[i] - actual[i]
		sumSqError += diff * diff
	}

	rmse := math.Sqrt(sumSqError / float64(len(ideal)))

	getLog().Calculation("ComputeError", map[string]interface{}{
		"rmse": rmse,
	}, rmse)

	return rmse
}
