// Package crossbar implements ferroelectric crossbar array simulation.
// This file implements write disturb (half-select stress) modeling.
//
// H10: Per Dr. Tour critique - add write disturb model for half-selected cells.
//
// Physics basis:
//   - During crossbar write operations, the target cell receives full voltage
//   - Half-selected cells (same row OR same column) receive partial voltage
//   - This partial voltage causes small polarization disturbance over time
//   - In passive (0T1R) arrays, half-select stress is a major reliability concern
//   - In active (1T1R/2T1R) arrays, transistors reduce half-select stress
//
// Write disturb mechanisms:
//   1. Field-induced switching: |V| < Vc but repeated stress accumulates
//   2. Charge injection: Incomplete switching creates trapped charge
//   3. Domain wall motion: Sub-threshold fields cause gradual domain drift
//
// References:
//   - IEEE IEDM 2022 (FeFET write disturb characterization)
//   - Nature Electronics 2023 (crossbar endurance with half-select)
//   - CrossSim documentation (half-select stress modeling)
package crossbar

import (
	"math"
	"sync"
)

// WriteDisturbConfig configures write disturb modeling.
type WriteDisturbConfig struct {
	Enable bool // Enable write disturb modeling

	// Half-select stress parameters
	// HalfSelectRatio: Voltage ratio for half-selected cells (typically 0.5)
	HalfSelectRatio float64

	// StressAccumulationRate: How much stress accumulates per write operation
	// Expressed as fraction of switching threshold per half-select exposure
	// Typical: 1e-4 to 1e-3 (0.01% to 0.1% per exposure)
	StressAccumulationRate float64

	// StressThreshold: Accumulated stress that causes 1-level shift
	// When stress exceeds this, the cell's level changes
	StressThreshold float64

	// Architecture dependency
	// 1T1R arrays have much lower disturb due to transistor isolation
	Architecture1T1R bool // If true, apply 1T1R reduction factor

	// 1T1R reduction factor (typically 0.01-0.1)
	// Transistor isolation reduces stress by 10-100x
	Architecture1T1RReduction float64
}

// DefaultWriteDisturbConfig returns default write disturb settings.
func DefaultWriteDisturbConfig() *WriteDisturbConfig {
	return &WriteDisturbConfig{
		Enable:                    false,                  // Disabled by default
		HalfSelectRatio:           0.5,                    // V/2 scheme
		StressAccumulationRate:    1e-4,                   // 0.01% per exposure
		StressThreshold:           1.0,                    // 1 unit = 1 level shift
		Architecture1T1R:          true,                   // Default to 1T1R
		Architecture1T1RReduction: 0.1,                    // 10x reduction with transistor
	}
}

// WriteDisturbEngine tracks and applies write disturb effects.
type WriteDisturbEngine struct {
	config *WriteDisturbConfig
	mu     sync.RWMutex

	// Stress accumulation matrix
	// stress[row][col] = accumulated stress on cell (row, col)
	stress [][]float64
	rows   int
	cols   int

	// Statistics
	TotalWriteOps     int     // Total write operations
	TotalHalfSelects  int     // Total half-select exposures
	DisturbedCells    int     // Cells that shifted due to disturb
	MaxStress         float64 // Maximum accumulated stress
	AvgStress         float64 // Average stress across array
}

// NewWriteDisturbEngine creates a new write disturb engine.
func NewWriteDisturbEngine(rows, cols int, config *WriteDisturbConfig) *WriteDisturbEngine {
	if config == nil {
		config = DefaultWriteDisturbConfig()
	}

	stress := make([][]float64, rows)
	for i := range stress {
		stress[i] = make([]float64, cols)
	}

	return &WriteDisturbEngine{
		config: config,
		stress: stress,
		rows:   rows,
		cols:   cols,
	}
}

// RecordWrite records a write operation and applies half-select stress.
// targetRow, targetCol: the cell being written
// The function applies stress to all half-selected cells.
func (e *WriteDisturbEngine) RecordWrite(targetRow, targetCol int) {
	if !e.config.Enable {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	e.TotalWriteOps++

	// Calculate stress increment
	stressIncrement := e.config.StressAccumulationRate
	if e.config.Architecture1T1R {
		stressIncrement *= e.config.Architecture1T1RReduction
	}

	// Apply stress to half-selected cells (same row OR same column, not both)
	for r := 0; r < e.rows; r++ {
		for c := 0; c < e.cols; c++ {
			// Skip the target cell (full voltage, intentional write)
			if r == targetRow && c == targetCol {
				continue
			}

			// Check if half-selected
			isHalfSelected := (r == targetRow) || (c == targetCol)
			if isHalfSelected {
				e.stress[r][c] += stressIncrement
				e.TotalHalfSelects++
			}
		}
	}
}

// RecordBatchWrite records multiple write operations (e.g., parallel column writes).
func (e *WriteDisturbEngine) RecordBatchWrite(targetCells [][2]int) {
	for _, cell := range targetCells {
		e.RecordWrite(cell[0], cell[1])
	}
}

// ApplyDisturbEffects applies accumulated stress effects to conductance matrix.
// Returns the number of cells that experienced level shifts.
func (e *WriteDisturbEngine) ApplyDisturbEffects(conductances [][]float64, levels int) int {
	if !e.config.Enable {
		return 0
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	disturbCount := 0
	gRange := GMax - GMin
	levelStep := gRange / float64(levels-1)

	for r := 0; r < e.rows && r < len(conductances); r++ {
		for c := 0; c < e.cols && c < len(conductances[r]); c++ {
			// Check if stress exceeds threshold
			if e.stress[r][c] >= e.config.StressThreshold {
				// Apply 1-level shift (random direction, slightly biased toward center)
				currentG := conductances[r][c]
				midG := (GMin + GMax) / 2

				// Bias toward center (regression to mean under stress)
				if currentG > midG {
					conductances[r][c] -= levelStep
				} else {
					conductances[r][c] += levelStep
				}

				// Clamp to valid range
				if conductances[r][c] < GMin {
					conductances[r][c] = GMin
				}
				if conductances[r][c] > GMax {
					conductances[r][c] = GMax
				}

				// Reduce stress (partial reset after disturbance)
				e.stress[r][c] -= e.config.StressThreshold
				disturbCount++
			}
		}
	}

	e.DisturbedCells += disturbCount
	e.updateStats()
	return disturbCount
}

// GetStressMatrix returns a copy of the stress matrix.
func (e *WriteDisturbEngine) GetStressMatrix() [][]float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	result := make([][]float64, e.rows)
	for r := 0; r < e.rows; r++ {
		result[r] = make([]float64, e.cols)
		copy(result[r], e.stress[r])
	}
	return result
}

// GetCellStress returns the stress level for a specific cell.
func (e *WriteDisturbEngine) GetCellStress(row, col int) float64 {
	e.mu.RLock()
	defer e.mu.RUnlock()

	if row >= 0 && row < e.rows && col >= 0 && col < e.cols {
		return e.stress[row][col]
	}
	return 0
}

// GetStressStats returns statistics about stress distribution.
func (e *WriteDisturbEngine) GetStressStats() WriteDisturbStats {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return WriteDisturbStats{
		TotalWriteOps:    e.TotalWriteOps,
		TotalHalfSelects: e.TotalHalfSelects,
		DisturbedCells:   e.DisturbedCells,
		MaxStress:        e.MaxStress,
		AvgStress:        e.AvgStress,
		CellsAtRisk:      e.countCellsAtRisk(),
	}
}

// WriteDisturbStats contains write disturb statistics.
type WriteDisturbStats struct {
	TotalWriteOps    int     // Total write operations performed
	TotalHalfSelects int     // Total half-select exposures
	DisturbedCells   int     // Cells that shifted due to disturb
	MaxStress        float64 // Maximum stress on any cell
	AvgStress        float64 // Average stress across array
	CellsAtRisk      int     // Cells with stress > 50% of threshold
}

// updateStats updates internal statistics (called with lock held).
func (e *WriteDisturbEngine) updateStats() {
	var sum, max float64
	count := 0

	for r := 0; r < e.rows; r++ {
		for c := 0; c < e.cols; c++ {
			s := e.stress[r][c]
			sum += s
			if s > max {
				max = s
			}
			count++
		}
	}

	e.MaxStress = max
	if count > 0 {
		e.AvgStress = sum / float64(count)
	}
}

// countCellsAtRisk counts cells with stress > 50% of threshold.
func (e *WriteDisturbEngine) countCellsAtRisk() int {
	threshold := e.config.StressThreshold * 0.5
	count := 0

	for r := 0; r < e.rows; r++ {
		for c := 0; c < e.cols; c++ {
			if e.stress[r][c] > threshold {
				count++
			}
		}
	}
	return count
}

// Reset clears all accumulated stress.
func (e *WriteDisturbEngine) Reset() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for r := 0; r < e.rows; r++ {
		for c := 0; c < e.cols; c++ {
			e.stress[r][c] = 0
		}
	}

	e.TotalWriteOps = 0
	e.TotalHalfSelects = 0
	e.DisturbedCells = 0
	e.MaxStress = 0
	e.AvgStress = 0
}

// Resize adjusts the engine for a new array size.
func (e *WriteDisturbEngine) Resize(rows, cols int) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.rows = rows
	e.cols = cols
	e.stress = make([][]float64, rows)
	for i := range e.stress {
		e.stress[i] = make([]float64, cols)
	}

	e.TotalWriteOps = 0
	e.TotalHalfSelects = 0
	e.DisturbedCells = 0
	e.MaxStress = 0
	e.AvgStress = 0
}

// EstimateDisturbRate estimates the disturb rate based on write pattern.
// writesPerCell: average writes per cell in the workload
// Returns the expected fraction of cells that will experience disturb.
func EstimateDisturbRate(writesPerCell float64, config *WriteDisturbConfig) float64 {
	if config == nil {
		config = DefaultWriteDisturbConfig()
	}
	if !config.Enable {
		return 0
	}

	// Each write causes (rows + cols - 2) half-selects in the array
	// Average half-selects per cell ≈ 2 * writesPerCell (simplified)
	avgHalfSelects := 2 * writesPerCell

	// Stress accumulation
	stressRate := config.StressAccumulationRate
	if config.Architecture1T1R {
		stressRate *= config.Architecture1T1RReduction
	}

	expectedStress := avgHalfSelects * stressRate

	// Probability of exceeding threshold (assuming normal distribution around expected)
	// Using simplified model: P(disturb) ≈ expectedStress / threshold
	disturbProb := expectedStress / config.StressThreshold
	if disturbProb > 1.0 {
		disturbProb = 1.0
	}

	return disturbProb
}

// CompareArchitectures compares write disturb between passive and active arrays.
func CompareArchitectures(writesPerCell float64) (passiveRate, activeRate float64) {
	passiveConfig := DefaultWriteDisturbConfig()
	passiveConfig.Enable = true
	passiveConfig.Architecture1T1R = false

	activeConfig := DefaultWriteDisturbConfig()
	activeConfig.Enable = true
	activeConfig.Architecture1T1R = true

	passiveRate = EstimateDisturbRate(writesPerCell, passiveConfig)
	activeRate = EstimateDisturbRate(writesPerCell, activeConfig)

	return passiveRate, activeRate
}

// HalfSelectVoltage calculates the voltage seen by half-selected cells.
func HalfSelectVoltage(writeVoltage float64, scheme string) float64 {
	switch scheme {
	case "V/2":
		return writeVoltage / 2
	case "V/3":
		return writeVoltage / 3
	case "floating":
		// Floating BL/WL - voltage depends on parasitics
		return writeVoltage * 0.4 // Approximate
	default:
		return writeVoltage / 2
	}
}

// IsDisturbCritical checks if half-select voltage exceeds safety margin.
func IsDisturbCritical(halfSelectV, coerciveV, safetyMargin float64) bool {
	// Disturb is critical if |V_half| > (1 - margin) * |Vc|
	threshold := math.Abs(coerciveV) * (1 - safetyMargin)
	return math.Abs(halfSelectV) > threshold
}
