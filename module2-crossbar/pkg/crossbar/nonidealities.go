// Package crossbar implements ferroelectric crossbar array simulation.
package crossbar

import (
	"errors"
	"math"
)

func init() {
	ErrInputSize = errors.New("input size exceeds array columns")
}

// WireParams contains wire resistance parameters for IR drop modeling.
type WireParams struct {
	RwordLine float64 // Word line resistance per unit (Ohm)
	RbitLine  float64 // Bit line resistance per unit (Ohm)
	Rcontact  float64 // Contact resistance (Ohm)
}

// DefaultWireParams returns typical wire parameters for a 45nm technology node.
func DefaultWireParams() *WireParams {
	return &WireParams{
		RwordLine: 2.5, // 2.5 Ohm per cell pitch
		RbitLine:  2.5, // 2.5 Ohm per cell pitch
		Rcontact:  50,  // 50 Ohm contact resistance
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
	return a.AnalyzeIRDropIterative(input, params, nil)
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
	return a.AnalyzeSneakPathsWithArch(selectedRow, selectedCol, false)
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
func (a *Array) AnalyzeSneakPathsWithIsolation(selectedRow, selectedCol int, isolationFactor float64) *SneakPathAnalysis {
	rows := a.config.Rows
	cols := a.config.Cols

	sneakMap := make([][]float64, rows)
	for i := range sneakMap {
		sneakMap[i] = make([]float64, cols)
	}

	// Calculate signal current (selected cell)
	signalG := a.cells[selectedRow][selectedCol].Conductance
	if signalG < 1e-10 {
		signalG = 1e-10 // Avoid division by zero
	}

	var totalSneak float64

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i == selectedRow && j == selectedCol {
				continue // Skip selected cell
			}

			var sneakG float64

			if i == selectedRow {
				// Same row as selected cell (but different column)
				// Sneak path: selected WL → this cell → BL j → return path
				// The cell's conductance directly contributes to sneak current
				// Higher conductance = more current leaks to this BL
				cellG := a.cells[i][j].Conductance

				// Sum return paths through all cells in column j (except this row)
				var returnPathG float64
				for k := 0; k < rows; k++ {
					if k != selectedRow {
						returnPathG += a.cells[k][j].Conductance
					}
				}

				// Series combination: this cell in series with parallel return paths
				if cellG > 0 && returnPathG > 0 {
					sneakG = (cellG * returnPathG) / (cellG + returnPathG)
				}

			} else if j == selectedCol {
				// Same column as selected cell (but different row)
				// Sneak path: WL i → cells in row i → return to this cell → selected BL
				// Current from other word lines can leak through this cell
				cellG := a.cells[i][j].Conductance

				// Sum paths from other columns in this row
				var feedPathG float64
				for k := 0; k < cols; k++ {
					if k != selectedCol {
						feedPathG += a.cells[i][k].Conductance
					}
				}

				// Series combination: feed paths in series with this cell
				if cellG > 0 && feedPathG > 0 {
					sneakG = (cellG * feedPathG) / (cellG + feedPathG)
				}

			} else {
				// Off-diagonal cell: three-cell sneak path
				// Path: selected WL → cell(sr,j) → BL j → cell(i,j) → WL i → cell(i,sc) → selected BL
				g1 := a.cells[selectedRow][j].Conductance // Entry point on selected row
				g2 := a.cells[i][j].Conductance           // This cell (corner of path)
				g3 := a.cells[i][selectedCol].Conductance // Exit point on selected column

				if g1 > 0 && g2 > 0 && g3 > 0 {
					// Three conductances in series: G_total = 1/(1/g1 + 1/g2 + 1/g3)
					sneakG = 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
				}
			}

			// Apply architecture isolation factor
			sneakG *= isolationFactor

			sneakMap[i][j] = sneakG
			totalSneak += sneakG
		}
	}

	// Calculate statistics
	var maxRatio, sumRatio float64
	count := 0
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if i == selectedRow && j == selectedCol {
				continue
			}
			ratio := sneakMap[i][j] / signalG
			if ratio > maxRatio {
				maxRatio = ratio
			}
			sumRatio += ratio
			count++
		}
	}

	avgRatio := 0.0
	if count > 0 {
		avgRatio = sumRatio / float64(count)
	}

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
	if len(input) > a.config.Cols {
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

	if maxRatio <= 0 {
		maxRatio = 1.0 // Default to 100%
	}

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
		}
	}

	return normalized
}

// ComputeError calculates the MVM output error due to non-idealities.
func ComputeError(ideal, actual []float64) float64 {
	if len(ideal) != len(actual) {
		return math.Inf(1)
	}

	var sumSqError float64
	for i := range ideal {
		diff := ideal[i] - actual[i]
		sumSqError += diff * diff
	}

	return math.Sqrt(sumSqError / float64(len(ideal)))
}
