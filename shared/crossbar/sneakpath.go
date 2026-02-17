package crossbar

import (
	"math"
)

// SneakPathAnalyzer analyzes parasitic current paths in crossbar arrays.
// In a crossbar, reading/writing to one cell can cause current to flow
// through unintended paths via other cells.
type SneakPathAnalyzer struct {
	Rows         int         // Number of rows
	Cols         int         // Number of columns
	Conductances [][]float64 // Conductance matrix (S)
	TargetRow    int         // Row of target cell
	TargetCol    int         // Column of target cell

	// Results
	SneakCurrents   [][]float64 // Sneak current through each cell
	TotalSneakRatio float64     // Ratio of sneak current to target current
	SneakPaths      []SneakPath // Individual sneak paths identified
}

// SneakPath represents a single sneak current path.
type SneakPath struct {
	StartRow    int      // Starting row
	StartCol    int      // Starting column
	PathCells   [][2]int // Cells along the path
	PathCurrent float64  // Current through this path
	PathLength  int      // Number of cells in path
}

// NewSneakPathAnalyzer creates a sneak path analyzer.
func NewSneakPathAnalyzer(rows, cols int) *SneakPathAnalyzer {
	getLog().Input("NewSneakPathAnalyzer", map[string]interface{}{
		"rows": rows,
		"cols": cols,
	})

	conductances := make([][]float64, rows)
	for i := range conductances {
		conductances[i] = make([]float64, cols)
		for j := range conductances[i] {
			conductances[i][j] = 50e-6 // 50µS typical
		}
	}

	analyzer := &SneakPathAnalyzer{
		Rows:          rows,
		Cols:          cols,
		Conductances:  conductances,
		SneakCurrents: make([][]float64, rows),
	}

	getLog().Output("NewSneakPathAnalyzer", analyzer)
	return analyzer
}

// SetConductance sets conductance for a specific cell.
func (sp *SneakPathAnalyzer) SetConductance(row, col int, g float64) {
	if row >= 0 && row < sp.Rows && col >= 0 && col < sp.Cols {
		sp.Conductances[row][col] = g
	}
}

// AnalyzeTarget analyzes sneak paths for a target cell.
//
// Physics model per VOLTAGE_RULES.md Section 3.3 (MVM/Compute):
// During MVM in passive (0T1R) mode, ALL word lines are active simultaneously.
// This enables 3-cell sneak paths through the array.
func (sp *SneakPathAnalyzer) AnalyzeTarget(targetRow, targetCol int, voltage float64) {
	getLog().Input("SneakPathAnalyzer.AnalyzeTarget", map[string]interface{}{
		"targetRow": targetRow,
		"targetCol": targetCol,
		"voltage":   voltage,
	})

	sp.TargetRow = targetRow
	sp.TargetCol = targetCol

	// Initialize sneak current array
	for i := range sp.SneakCurrents {
		sp.SneakCurrents[i] = make([]float64, sp.Cols)
	}

	// Calculate target current (direct path)
	targetCurrent := voltage * sp.Conductances[targetRow][targetCol]

	// Calculate sneak currents through parallel paths
	// A sneak path goes: target row -> other column -> other row -> target column
	totalSneakCurrent := 0.0
	sp.SneakPaths = make([]SneakPath, 0)

	for otherRow := 0; otherRow < sp.Rows; otherRow++ {
		if otherRow == targetRow {
			continue
		}

		for otherCol := 0; otherCol < sp.Cols; otherCol++ {
			if otherCol == targetCol {
				continue
			}

			// Three-cell sneak path:
			// Cell 1: (targetRow, otherCol) - on target row
			// Cell 2: (otherRow, otherCol) - connects rows
			// Cell 3: (otherRow, targetCol) - on target column

			g1 := sp.Conductances[targetRow][otherCol]
			g2 := sp.Conductances[otherRow][otherCol]
			g3 := sp.Conductances[otherRow][targetCol]

			// Series conductance through sneak path
			if g1 > 0 && g2 > 0 && g3 > 0 {
				// For three resistors in series: 1/G_total = 1/G1 + 1/G2 + 1/G3
				gSeries := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
				sneakCurrent := voltage * gSeries

				sp.SneakCurrents[otherRow][otherCol] += sneakCurrent
				totalSneakCurrent += sneakCurrent

				// Record this sneak path
				path := SneakPath{
					StartRow: targetRow,
					StartCol: targetCol,
					PathCells: [][2]int{
						{targetRow, otherCol},
						{otherRow, otherCol},
						{otherRow, targetCol},
					},
					PathCurrent: sneakCurrent,
					PathLength:  3,
				}
				sp.SneakPaths = append(sp.SneakPaths, path)
			}
		}
	}

	// Calculate sneak ratio
	if targetCurrent > 0 {
		sp.TotalSneakRatio = totalSneakCurrent / targetCurrent
	} else {
		sp.TotalSneakRatio = math.Inf(1)
	}

	getLog().Calculation("SneakPathAnalyzer.AnalyzeTarget", map[string]interface{}{
		"targetCurrent":     targetCurrent,
		"totalSneakCurrent": totalSneakCurrent,
		"sneakRatio":        sp.TotalSneakRatio,
		"numPaths":          len(sp.SneakPaths),
	}, nil)
}

// SneakPathStats contains statistics about sneak paths.
type SneakPathStats struct {
	TargetCurrent      float64 // Current through target cell (A)
	TotalSneakCurrent  float64 // Total sneak current (A)
	SneakRatio         float64 // Sneak current / target current
	NumSneakPaths      int     // Number of sneak paths
	WorstSneakPath     float64 // Largest single sneak current
	SignalToNoiseRatio float64 // SNR in dB
}

// GetStats returns sneak path statistics.
func (sp *SneakPathAnalyzer) GetStats(voltage float64) SneakPathStats {
	targetCurrent := voltage * sp.Conductances[sp.TargetRow][sp.TargetCol]

	totalSneakCurrent := 0.0
	worstSneak := 0.0
	for _, path := range sp.SneakPaths {
		totalSneakCurrent += path.PathCurrent
		if path.PathCurrent > worstSneak {
			worstSneak = path.PathCurrent
		}
	}

	snr := 0.0
	if totalSneakCurrent > 0 {
		snr = 20 * math.Log10(targetCurrent/totalSneakCurrent)
	} else {
		snr = 100 // Very high SNR (no sneak)
	}

	return SneakPathStats{
		TargetCurrent:      targetCurrent,
		TotalSneakCurrent:  totalSneakCurrent,
		SneakRatio:         sp.TotalSneakRatio,
		NumSneakPaths:      len(sp.SneakPaths),
		WorstSneakPath:     worstSneak,
		SignalToNoiseRatio: snr,
	}
}

// SneakMitigation represents mitigation strategies for sneak paths.
type SneakMitigation struct {
	UseSelector       bool    // Use selector device (1T1R, 1S1R)
	SelectorOnOff     float64 // Selector on/off ratio
	UseHalfSelect     bool    // Use half-select voltage scheme
	HalfSelectVoltage float64 // Half-select voltage (V)
}

// AnalyzeWithMitigation analyzes with mitigation applied.
func (sp *SneakPathAnalyzer) AnalyzeWithMitigation(targetRow, targetCol int, voltage float64, mit SneakMitigation) SneakPathStats {
	sp.TargetRow = targetRow
	sp.TargetCol = targetCol

	// Initialize sneak current array
	for i := range sp.SneakCurrents {
		sp.SneakCurrents[i] = make([]float64, sp.Cols)
	}

	targetCurrent := voltage * sp.Conductances[targetRow][targetCol]

	totalSneakCurrent := 0.0
	sp.SneakPaths = make([]SneakPath, 0)

	for otherRow := 0; otherRow < sp.Rows; otherRow++ {
		if otherRow == targetRow {
			continue
		}

		for otherCol := 0; otherCol < sp.Cols; otherCol++ {
			if otherCol == targetCol {
				continue
			}

			g1 := sp.Conductances[targetRow][otherCol]
			g2 := sp.Conductances[otherRow][otherCol]
			g3 := sp.Conductances[otherRow][targetCol]

			// Apply mitigation effects
			effectiveVoltage := voltage
			conductanceMultiplier := 1.0

			if mit.UseSelector {
				// Selector reduces off-path conductance
				conductanceMultiplier = 1.0 / mit.SelectorOnOff
				g1 *= conductanceMultiplier
				g2 *= conductanceMultiplier
				g3 *= conductanceMultiplier
			}

			if mit.UseHalfSelect {
				// Half-select reduces effective voltage on sneak paths
				effectiveVoltage = voltage - mit.HalfSelectVoltage
				if effectiveVoltage < 0 {
					effectiveVoltage = 0
				}
			}

			if g1 > 0 && g2 > 0 && g3 > 0 {
				gSeries := 1.0 / (1.0/g1 + 1.0/g2 + 1.0/g3)
				sneakCurrent := effectiveVoltage * gSeries

				sp.SneakCurrents[otherRow][otherCol] += sneakCurrent
				totalSneakCurrent += sneakCurrent

				path := SneakPath{
					StartRow: targetRow,
					StartCol: targetCol,
					PathCells: [][2]int{
						{targetRow, otherCol},
						{otherRow, otherCol},
						{otherRow, targetCol},
					},
					PathCurrent: sneakCurrent,
					PathLength:  3,
				}
				sp.SneakPaths = append(sp.SneakPaths, path)
			}
		}
	}

	if targetCurrent > 0 {
		sp.TotalSneakRatio = totalSneakCurrent / targetCurrent
	} else {
		sp.TotalSneakRatio = math.Inf(1)
	}

	snr := 0.0
	if totalSneakCurrent > 0 {
		snr = 20 * math.Log10(targetCurrent/totalSneakCurrent)
	} else {
		snr = 100
	}

	worstSneak := 0.0
	for _, path := range sp.SneakPaths {
		if path.PathCurrent > worstSneak {
			worstSneak = path.PathCurrent
		}
	}

	return SneakPathStats{
		TargetCurrent:      targetCurrent,
		TotalSneakCurrent:  totalSneakCurrent,
		SneakRatio:         sp.TotalSneakRatio,
		NumSneakPaths:      len(sp.SneakPaths),
		WorstSneakPath:     worstSneak,
		SignalToNoiseRatio: snr,
	}
}
