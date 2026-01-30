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
		RwordLine: 2.5,    // 2.5 Ohm per cell pitch
		RbitLine:  2.5,    // 2.5 Ohm per cell pitch
		Rcontact:  50,     // 50 Ohm contact resistance
		CwordLine: 0.2e-15, // 0.2 fF per cell pitch (typical 45nm)
		CbitLine:  0.2e-15, // 0.2 fF per cell pitch
		CellPitch: 90e-9,  // 90 nm cell pitch (2× minimum feature size for 45nm)
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
//   τ_elmore = R_total × C_total / 2 (for load at end)
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
		"maxDrop":     result.MaxIRDrop,
		"avgDrop":     result.AvgIRDrop,
		"variance":    result.IRDropVariance,
		"worstCell":   result.WorstCaseCell,
	}, result)
	return result
}

// AnalyzeIRDropIterative performs physics-based IR drop analysis with configurable solver.
func (a *Array) AnalyzeIRDropIterative(input []float64, params *WireParams, config *IRDropSolverConfig) *IRDropAnalysis {
	if params == nil {
		params = DefaultWireParams()
	}
	if config == nil {
		config = DefaultIRDropSolverConfig()
	}

	rows := a.config.Rows
	cols := a.config.Cols

	// Initialize voltage matrices
	// Per PHYSICS.md: V_wl(i,j) = V_input, V_bl(i,j) = 0
	wlVoltage := make([][]float64, rows)
	blVoltage := make([][]float64, rows)
	effVoltage := make([][]float64, rows)
	cellCurrent := make([][]float64, rows) // Current through each cell

	for i := 0; i < rows; i++ {
		wlVoltage[i] = make([]float64, cols)
		blVoltage[i] = make([]float64, cols)
		effVoltage[i] = make([]float64, cols)
		cellCurrent[i] = make([]float64, cols)

		// Initialize: word line voltage = 1V (normalized driver voltage)
		// bit line voltage = 0V (ground at sense amp)
		for j := 0; j < cols; j++ {
			wlVoltage[i][j] = 1.0 // All word lines driven at 1V
			blVoltage[i][j] = 0   // Bit lines start at ground
		}
	}

	// Iterative relaxation solver
	// Couples the voltage drops and currents until convergence
	for iter := 0; iter < config.MaxIterations; iter++ {
		maxChange := 0.0

		// Store old voltages for convergence check
		oldWL := make([][]float64, rows)
		oldBL := make([][]float64, rows)
		for i := 0; i < rows; i++ {
			oldWL[i] = make([]float64, cols)
			oldBL[i] = make([]float64, cols)
			copy(oldWL[i], wlVoltage[i])
			copy(oldBL[i], blVoltage[i])
		}

		// Step 1: Calculate cell currents based on current voltages
		// I_ij = V_eff × G_ij × input_scale where V_eff = V_wl - V_bl
		// The input vector scales the current (represents DAC modulation)
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				effVoltage[i][j] = wlVoltage[i][j] - blVoltage[i][j]
				if effVoltage[i][j] < 0 {
					effVoltage[i][j] = 0
				}
				// Get physical conductance (in Siemens)
				gPhys := a.GetPhysicalConductanceForCell(i, j)
				// Scale by input (represents DAC output affecting current)
				inputScale := 1.0
				if j < len(input) {
					inputScale = input[j]
				}
				cellCurrent[i][j] = effVoltage[i][j] * gPhys * inputScale
			}
		}

		// Step 2: Update word line voltages
		// Driver topology: WL drivers on LEFT (column 0)
		// V_wl(i,j) = V_driver - sum of drops from driver to position j
		// Drop per segment = R_wl × (current through that segment)
		// Current through segment k→k+1 = sum of cell currents from columns k+1 to end
		for i := 0; i < rows; i++ {
			driverV := 1.0 // All word lines driven at normalized 1V

			cumulativeDrop := 0.0
			for j := 0; j < cols; j++ {
				// Current flowing through wire segment before position j
				// is sum of all cell currents from this column onwards
				if j > 0 {
					var segmentCurrent float64
					for k := j; k < cols; k++ {
						segmentCurrent += cellCurrent[i][k]
					}
					cumulativeDrop += params.RwordLine * segmentCurrent
				}

				// New word line voltage with damping
				newWL := driverV - cumulativeDrop
				if newWL < 0 {
					newWL = 0
				}
				wlVoltage[i][j] = config.Damping*newWL + (1-config.Damping)*oldWL[i][j]
			}
		}

		// Step 3: Update bit line voltages
		// Topology: BL sense amps/ground at TOP (row 0)
		// V_bl(i,j) = sum of drops from ground to position i
		// Drop per segment = R_bl × (sum of currents from row 0 to i)
		for j := 0; j < cols; j++ {
			cumulativeDrop := 0.0
			for i := 0; i < rows; i++ {
				// Current flowing through wire segment from i-1 to i
				// is sum of all cell currents from row 0 to row i-1
				var segmentCurrent float64
				for k := 0; k < i; k++ {
					segmentCurrent += cellCurrent[k][j]
				}

				// Voltage rise from ground (drop going upward = rise going downward)
				if i > 0 {
					cumulativeDrop += params.RbitLine * segmentCurrent
				}

				// Add contact resistance at the cell
				cumulativeDrop += params.Rcontact * cellCurrent[i][j] * 0.001 // Scaled contact effect

				// New bit line voltage with damping
				newBL := cumulativeDrop
				blVoltage[i][j] = config.Damping*newBL + (1-config.Damping)*oldBL[i][j]
			}
		}

		// Check convergence
		for i := 0; i < rows; i++ {
			for j := 0; j < cols; j++ {
				changeWL := math.Abs(wlVoltage[i][j] - oldWL[i][j])
				changeBL := math.Abs(blVoltage[i][j] - oldBL[i][j])
				if changeWL > maxChange {
					maxChange = changeWL
				}
				if changeBL > maxChange {
					maxChange = changeBL
				}
			}
		}

		// Converged when max voltage change is below tolerance
		if maxChange < config.Tolerance {
			break
		}
	}

	// Final effective voltage calculation
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			effVoltage[i][j] = wlVoltage[i][j] - blVoltage[i][j]
			if effVoltage[i][j] < 0 {
				effVoltage[i][j] = 0
			}
		}
	}

	// Calculate statistics
	var maxDrop, sumDrop float64
	var worstCell [2]int
	count := 0

	// Reference voltage for drop calculation (normalized input)
	refVoltage := 1.0

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			drop := refVoltage - effVoltage[i][j]
			if drop < 0 {
				drop = 0
			}
			if drop > maxDrop {
				maxDrop = drop
				worstCell = [2]int{i, j}
			}
			sumDrop += drop
			count++
		}
	}

	avgDrop := sumDrop / float64(count)

	// Calculate variance
	var variance float64
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			drop := refVoltage - effVoltage[i][j]
			if drop < 0 {
				drop = 0
			}
			variance += (drop - avgDrop) * (drop - avgDrop)
		}
	}
	variance /= float64(count)

	return &IRDropAnalysis{
		WordLineVoltages: wlVoltage,
		BitLineVoltages:  blVoltage,
		EffectiveVoltage: effVoltage,
		MaxIRDrop:        maxDrop,
		AvgIRDrop:        avgDrop,
		IRDropVariance:   variance,
		WorstCaseCell:    worstCell,
	}
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

// AnalyzeSneakPathsWithIsolation analyzes sneak paths with a configurable isolation factor.
// isolationFactor values:
//   - 1.0:    0T1R (passive) - full sneak path impact
//   - 0.001:  1T1R - ~1000x isolation from single transistor
//   - 0.0001: 2T1R - ~10000x isolation from dual transistor AND-gate
//
// Physics model per VOLTAGE_RULES.md Section 3.3 (MVM/Compute):
// During MVM in passive (0T1R) mode, ALL word lines are active simultaneously.
// This enables 3-cell sneak paths: WL_i → cell(i,j) → BL_j → cell(k,j) → WL_k → ...
//
// IMPORTANT: Only OFF-DIAGONAL three-cell paths are true sneak paths during MVM:
//   - Same-row cells contribute to DIFFERENT column outputs (not sneak)
//   - Same-column cells receive DIFFERENT input voltages (part of MVM computation)
//   - Only off-diagonal paths form parasitic loops back to the selected cell's readout
//
// Per IEEE/industry standards (CrossSim, arXiv 2025), sneak ratio for passive
// arrays should be 5-20%, not 100%+. The three-cell series model is:
//
//	G_path = 1/(1/G1 + 1/G2 + 1/G3)
//	I_sneak = V × G_path
func (a *Array) AnalyzeSneakPathsWithIsolation(selectedRow, selectedCol int, isolationFactor float64) *SneakPathAnalysis {
	rows := a.config.Rows
	cols := a.config.Cols

	getLog().Input("AnalyzeSneakPathsWithIsolation", map[string]interface{}{
		"selectedRow":     selectedRow,
		"selectedCol":     selectedCol,
		"isolationFactor": isolationFactor,
		"rows":            rows,
		"cols":            cols,
	})

	sneakMap := make([][]float64, rows)
	for i := range sneakMap {
		sneakMap[i] = make([]float64, cols)
	}

	// Calculate signal current (selected cell)
	signalG := a.cells[selectedRow][selectedCol].Conductance
	if signalG < 1e-10 {
		signalG = 1e-10 // Avoid division by zero
	}
	getLog().Calculation("SignalConductance", map[string]interface{}{
		"signalG": signalG,
	}, nil)

	var totalSneak float64
	var offDiagCount int
	var offDiagSneak float64

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i == selectedRow || j == selectedCol {
				// Skip selected cell AND same-row/same-column cells
				// Same-row: contributes to different column output (not sneak)
				// Same-column: receives different input voltage (part of MVM)
				continue
			}

			// Off-diagonal: Three-cell sneak path
			// Path: cell(selectedRow, j) → cell(i, j) → cell(i, selectedCol)
			g1 := a.cells[selectedRow][j].Conductance // Same row as target, other column
			g2 := a.cells[i][j].Conductance           // The off-diagonal cell
			g3 := a.cells[i][selectedCol].Conductance // Same column as target, other row

			var sneakG float64
			if g1 > 0 && g2 > 0 && g3 > 0 {
				// Series conductance: 1/G_total = 1/G1 + 1/G2 + 1/G3
				sneakG = 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
			}
			offDiagCount++
			offDiagSneak += sneakG

			// Apply architecture isolation factor
			sneakG *= isolationFactor

			sneakMap[i][j] = sneakG
			totalSneak += sneakG
		}
	}

	// Calculate statistics over off-diagonal cells only
	var maxRatio, sumRatio float64
	var maxRatioRow, maxRatioCol int
	count := 0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i == selectedRow || j == selectedCol {
				continue // Only count off-diagonal cells
			}
			ratio := sneakMap[i][j] / signalG
			if ratio > maxRatio {
				maxRatio = ratio
				maxRatioRow = i
				maxRatioCol = j
			}
			sumRatio += ratio
			count++
		}
	}

	avgRatio := 0.0
	if count > 0 {
		avgRatio = sumRatio / float64(count)
	}

	getLog().Calculation("SneakPathAnalysis", map[string]interface{}{
		"offDiagCount":    offDiagCount,
		"offDiagSneak":    offDiagSneak,
		"totalSneak":      totalSneak,
		"maxRatio":        maxRatio,
		"maxRatioCell":    []int{maxRatioRow, maxRatioCol},
		"avgRatio":        avgRatio,
		"isolationFactor": isolationFactor,
	}, nil)

	return &SneakPathAnalysis{
		SneakCurrents: sneakMap,
		MaxSneakRatio: maxRatio,
		AvgSneakRatio: avgRatio,
		TotalSneak:    totalSneak,
		TotalSignal:   signalG,
	}
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

			g := a.cells[i][j].Conductance * a.cells[i][j].NoiseFactor
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
