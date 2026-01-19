package nonidealities

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
	StartRow    int     // Starting row
	StartCol    int     // Starting column
	PathCells   [][2]int // Cells along the path
	PathCurrent float64 // Current through this path
	PathLength  int     // Number of cells in path
}

// NewSneakPathAnalyzer creates a sneak path analyzer.
func NewSneakPathAnalyzer(rows, cols int) *SneakPathAnalyzer {
	conductances := make([][]float64, rows)
	for i := range conductances {
		conductances[i] = make([]float64, cols)
		for j := range conductances[i] {
			conductances[i][j] = 50e-6 // 50µS typical
		}
	}

	return &SneakPathAnalyzer{
		Rows:          rows,
		Cols:          cols,
		Conductances:  conductances,
		SneakCurrents: make([][]float64, rows),
	}
}

// SetConductance sets conductance for a specific cell.
func (sp *SneakPathAnalyzer) SetConductance(row, col int, g float64) {
	if row >= 0 && row < sp.Rows && col >= 0 && col < sp.Cols {
		sp.Conductances[row][col] = g
	}
}

// SetConductancePattern sets conductances based on a weight pattern.
func (sp *SneakPathAnalyzer) SetConductancePattern(weights [][]int, levels int) {
	gMin := 1e-6   // 1µS minimum conductance
	gMax := 100e-6 // 100µS maximum conductance

	for i := 0; i < sp.Rows && i < len(weights); i++ {
		for j := 0; j < sp.Cols && j < len(weights[i]); j++ {
			// Linear mapping from level to conductance
			level := weights[i][j]
			sp.Conductances[i][j] = gMin + (gMax-gMin)*float64(level)/float64(levels-1)
		}
	}
}

// AnalyzeTarget analyzes sneak paths for a target cell.
func (sp *SneakPathAnalyzer) AnalyzeTarget(targetRow, targetCol int, voltage float64) {
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

// GetSneakCurrentMap returns 2D map of sneak currents.
func (sp *SneakPathAnalyzer) GetSneakCurrentMap() [][]float64 {
	return sp.SneakCurrents
}

// GetTopSneakPaths returns the n largest sneak paths.
func (sp *SneakPathAnalyzer) GetTopSneakPaths(n int) []SneakPath {
	if n > len(sp.SneakPaths) {
		n = len(sp.SneakPaths)
	}

	// Simple selection of top n (not optimal but works for small n)
	result := make([]SneakPath, n)
	used := make([]bool, len(sp.SneakPaths))

	for i := 0; i < n; i++ {
		maxIdx := -1
		maxCurrent := -1.0
		for j, path := range sp.SneakPaths {
			if !used[j] && path.PathCurrent > maxCurrent {
				maxCurrent = path.PathCurrent
				maxIdx = j
			}
		}
		if maxIdx >= 0 {
			result[i] = sp.SneakPaths[maxIdx]
			used[maxIdx] = true
		}
	}

	return result
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

// ArraySneakAnalysis analyzes sneak paths for entire array operation.
type ArraySneakAnalysis struct {
	Analyzer        *SneakPathAnalyzer
	CellSneakRatios [][]float64 // Sneak ratio at each position
	AvgSneakRatio   float64
	MaxSneakRatio   float64
	WorstCellRow    int
	WorstCellCol    int
}

// AnalyzeArray performs full array sneak path analysis.
func (sp *SneakPathAnalyzer) AnalyzeArray(voltage float64) ArraySneakAnalysis {
	analysis := ArraySneakAnalysis{
		Analyzer:        sp,
		CellSneakRatios: make([][]float64, sp.Rows),
	}

	for i := range analysis.CellSneakRatios {
		analysis.CellSneakRatios[i] = make([]float64, sp.Cols)
	}

	totalRatio := 0.0
	maxRatio := 0.0

	for i := 0; i < sp.Rows; i++ {
		for j := 0; j < sp.Cols; j++ {
			sp.AnalyzeTarget(i, j, voltage)
			ratio := sp.TotalSneakRatio
			analysis.CellSneakRatios[i][j] = ratio
			totalRatio += ratio

			if ratio > maxRatio {
				maxRatio = ratio
				analysis.WorstCellRow = i
				analysis.WorstCellCol = j
			}
		}
	}

	analysis.AvgSneakRatio = totalRatio / float64(sp.Rows*sp.Cols)
	analysis.MaxSneakRatio = maxRatio

	return analysis
}
