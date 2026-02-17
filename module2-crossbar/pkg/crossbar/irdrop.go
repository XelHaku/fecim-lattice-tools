package crossbar

import (
	"math"
)

// IRDropSimulator models voltage drop along metal lines in a crossbar.
type IRDropSimulator struct {
	Rows         int         // Number of rows in crossbar
	Cols         int         // Number of columns in crossbar
	RowResist    float64     // Resistance per row segment (Ohms)
	ColResist    float64     // Resistance per column segment (Ohms)
	Temperature  float64     // Operating temperature in Kelvin (default 300K)
	VoltageIn    []float64   // Input voltages (row drivers)
	VoltageOut   []float64   // Output voltages (column sense)
	Conductances [][]float64 // Conductance matrix (S)

	// Results
	RowVoltages  [][]float64 // Voltage at each row position
	ColVoltages  [][]float64 // Voltage at each column position
	CellCurrents [][]float64 // Actual current through each cell
	IRDropMap    [][]float64 // IR drop at each cell position
}

// NewIRDropSimulator creates an IR drop simulator for a crossbar.
func NewIRDropSimulator(rows, cols int) *IRDropSimulator {
	getLog().Input("NewIRDropSimulator", map[string]interface{}{
		"rows": rows,
		"cols": cols,
	})

	// Initialize conductance matrix
	conductances := make([][]float64, rows)
	for i := range conductances {
		conductances[i] = make([]float64, cols)
		for j := range conductances[i] {
			conductances[i][j] = 50e-6 // 50µS typical (middle of range)
		}
	}

	sim := &IRDropSimulator{
		Rows:         rows,
		Cols:         cols,
		RowResist:    2.5, // 2.5 Ω per segment (typical metal interconnect)
		ColResist:    2.5, // 2.5 Ω per segment
		Temperature:  300.0,
		VoltageIn:    make([]float64, rows),
		VoltageOut:   make([]float64, cols),
		Conductances: conductances,
		RowVoltages:  make([][]float64, rows),
		ColVoltages:  make([][]float64, rows),
		CellCurrents: make([][]float64, rows),
		IRDropMap:    make([][]float64, rows),
	}

	getLog().Output("NewIRDropSimulator", sim)
	return sim
}

// SetTemperature sets the operating temperature and adjusts wire resistances.
// Uses copper TCR = 0.00393/K for resistance scaling.
func (ir *IRDropSimulator) SetTemperature(tempK float64) {
	if tempK < 0 {
		tempK = 300 // Default to room temperature
	}
	ir.Temperature = tempK

	// Apply temperature coefficient of resistance (TCR) for copper
	// R(T) = R(300K) * (1 + 0.00393 * (T - 300))
	const copperTCR = 0.00393
	const refTemp = 300.0
	const baseRowResist = 2.5 // Base resistance at 300K
	const baseColResist = 2.5

	factor := 1.0 + copperTCR*(tempK-refTemp)
	ir.RowResist = baseRowResist * factor
	ir.ColResist = baseColResist * factor
}

// GetTemperatureEffect returns the resistance multiplier due to temperature.
func (ir *IRDropSimulator) GetTemperatureEffect() float64 {
	const copperTCR = 0.00393
	const refTemp = 300.0
	return 1.0 + copperTCR*(ir.Temperature-refTemp)
}

// SetConductance sets conductance for a specific cell.
func (ir *IRDropSimulator) SetConductance(row, col int, g float64) {
	if row >= 0 && row < ir.Rows && col >= 0 && col < ir.Cols {
		ir.Conductances[row][col] = g
	}
}

// SetInputVoltage sets input voltage for a row.
func (ir *IRDropSimulator) SetInputVoltage(row int, v float64) {
	if row >= 0 && row < ir.Rows {
		ir.VoltageIn[row] = v
	}
}

// SetAllInputs sets all row input voltages.
func (ir *IRDropSimulator) SetAllInputs(voltages []float64) {
	for i := 0; i < ir.Rows && i < len(voltages); i++ {
		ir.VoltageIn[i] = voltages[i]
	}
}

// Simulate runs the IR drop simulation using Gauss-Seidel relaxation on KCL.
func (ir *IRDropSimulator) Simulate(iterations int) {
	getLog().Input("IRDropSimulator.Simulate", map[string]interface{}{
		"iterations": iterations,
		"rows":       ir.Rows,
		"cols":       ir.Cols,
		"rowResist":  ir.RowResist,
		"colResist":  ir.ColResist,
	})

	// Initialize arrays
	for i := range ir.RowVoltages {
		ir.RowVoltages[i] = make([]float64, ir.Cols)
		ir.ColVoltages[i] = make([]float64, ir.Cols)
		ir.CellCurrents[i] = make([]float64, ir.Cols)
		ir.IRDropMap[i] = make([]float64, ir.Cols)
	}

	for i := 0; i < ir.Rows; i++ {
		for j := 0; j < ir.Cols; j++ {
			ir.RowVoltages[i][j] = ir.VoltageIn[i]
			ir.ColVoltages[i][j] = ir.VoltageOut[j]
		}
	}

	if iterations <= 0 {
		iterations = 1
	}

	gRow := 0.0
	if ir.RowResist > 0 {
		gRow = 1.0 / ir.RowResist
	}
	gCol := 0.0
	if ir.ColResist > 0 {
		gCol = 1.0 / ir.ColResist
	}

	for iter := 0; iter < iterations; iter++ {
		// Update row-wire nodes
		for i := 0; i < ir.Rows; i++ {
			for j := 0; j < ir.Cols; j++ {
				gCell := ir.Conductances[i][j]
				den := gCell
				num := gCell * ir.ColVoltages[i][j]

				// Left neighbor or driver
				if j == 0 {
					den += gRow
					num += gRow * ir.VoltageIn[i]
				} else {
					den += gRow
					num += gRow * ir.RowVoltages[i][j-1]
				}

				// Right neighbor (open at far end)
				if j < ir.Cols-1 {
					den += gRow
					num += gRow * ir.RowVoltages[i][j+1]
				}

				if den > 0 {
					ir.RowVoltages[i][j] = num / den
				}
			}
		}

		// Update column-wire nodes
		for i := 0; i < ir.Rows; i++ {
			for j := 0; j < ir.Cols; j++ {
				gCell := ir.Conductances[i][j]
				den := gCell
				num := gCell * ir.RowVoltages[i][j]

				// Down neighbor or sense node
				if i == ir.Rows-1 {
					den += gCol
					num += gCol * ir.VoltageOut[j]
				} else {
					den += gCol
					num += gCol * ir.ColVoltages[i+1][j]
				}

				// Up neighbor (open at top)
				if i > 0 {
					den += gCol
					num += gCol * ir.ColVoltages[i-1][j]
				}

				if den > 0 {
					ir.ColVoltages[i][j] = num / den
				}
			}
		}
	}

	// Currents and IR drop map
	var maxDrop float64
	for i := 0; i < ir.Rows; i++ {
		for j := 0; j < ir.Cols; j++ {
			vActual := ir.RowVoltages[i][j] - ir.ColVoltages[i][j]
			ir.CellCurrents[i][j] = vActual * ir.Conductances[i][j]

			vIdeal := ir.VoltageIn[i] - ir.VoltageOut[j]
			ir.IRDropMap[i][j] = math.Abs(vIdeal - vActual)
			if ir.IRDropMap[i][j] > maxDrop {
				maxDrop = ir.IRDropMap[i][j]
			}
		}
	}

	getLog().Calculation("IRDropSimulator.Simulate", map[string]interface{}{
		"maxDrop": maxDrop,
		"avgDrop": ir.GetAvgIRDrop(),
	}, nil)
}

// GetMaxIRDrop returns the maximum IR drop in the array.
func (ir *IRDropSimulator) GetMaxIRDrop() float64 {
	maxDrop := 0.0
	for i := 0; i < ir.Rows; i++ {
		for j := 0; j < ir.Cols; j++ {
			if ir.IRDropMap[i][j] > maxDrop {
				maxDrop = ir.IRDropMap[i][j]
			}
		}
	}
	return maxDrop
}

// GetAvgIRDrop returns the average IR drop in the array.
func (ir *IRDropSimulator) GetAvgIRDrop() float64 {
	total := 0.0
	count := 0
	for i := 0; i < ir.Rows; i++ {
		for j := 0; j < ir.Cols; j++ {
			total += ir.IRDropMap[i][j]
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / float64(count)
}

// GetOutputCurrents returns the output currents (sum of each column).
func (ir *IRDropSimulator) GetOutputCurrents() []float64 {
	outputs := make([]float64, ir.Cols)
	for j := 0; j < ir.Cols; j++ {
		for i := 0; i < ir.Rows; i++ {
			outputs[j] += ir.CellCurrents[i][j]
		}
	}
	return outputs
}

// GetIdealOutputs returns ideal outputs (no IR drop).
func (ir *IRDropSimulator) GetIdealOutputs() []float64 {
	outputs := make([]float64, ir.Cols)
	for j := 0; j < ir.Cols; j++ {
		for i := 0; i < ir.Rows; i++ {
			outputs[j] += ir.VoltageIn[i] * ir.Conductances[i][j]
		}
	}
	return outputs
}

// GetOutputError returns the relative error between ideal and actual outputs.
func (ir *IRDropSimulator) GetOutputError() []float64 {
	actual := ir.GetOutputCurrents()
	ideal := ir.GetIdealOutputs()
	errors := make([]float64, ir.Cols)

	for j := range errors {
		if ideal[j] != 0 {
			errors[j] = math.Abs(actual[j]-ideal[j]) / math.Abs(ideal[j]) * 100
		}
	}
	return errors
}

// IRDropStats contains statistics about IR drop.
type IRDropStats struct {
	MaxIRDrop      float64 // Maximum IR drop (V)
	AvgIRDrop      float64 // Average IR drop (V)
	MaxOutputError float64 // Maximum output error (%)
	AvgOutputError float64 // Average output error (%)
	WorstCellRow   int     // Row of worst-affected cell
	WorstCellCol   int     // Column of worst-affected cell
	Temperature    float64 // Operating temperature (K)
	ResistFactor   float64 // Resistance multiplier due to temperature
}

// GetStats returns IR drop statistics.
func (ir *IRDropSimulator) GetStats() IRDropStats {
	errors := ir.GetOutputError()

	maxErr := 0.0
	avgErr := 0.0
	for _, e := range errors {
		avgErr += e
		if e > maxErr {
			maxErr = e
		}
	}
	avgErr /= float64(len(errors))

	// Find worst cell
	worstRow, worstCol := 0, 0
	maxDrop := 0.0
	for i := 0; i < ir.Rows; i++ {
		for j := 0; j < ir.Cols; j++ {
			if ir.IRDropMap[i][j] > maxDrop {
				maxDrop = ir.IRDropMap[i][j]
				worstRow = i
				worstCol = j
			}
		}
	}

	return IRDropStats{
		MaxIRDrop:      ir.GetMaxIRDrop(),
		AvgIRDrop:      ir.GetAvgIRDrop(),
		MaxOutputError: maxErr,
		AvgOutputError: avgErr,
		WorstCellRow:   worstRow,
		WorstCellCol:   worstCol,
		Temperature:    ir.Temperature,
		ResistFactor:   ir.GetTemperatureEffect(),
	}
}

// IRDropMitigation represents mitigation strategies.
type IRDropMitigation struct {
	UseWidenedLines   bool    // Use wider metal lines
	UseHierarchical   bool    // Use hierarchical bus structure
	UseTiledArray     bool    // Break into smaller tiles
	LineWidthIncrease float64 // Factor to increase line width
	TileSize          int     // Size of tiles if tiling enabled
}

// ApplyMitigation applies mitigation strategies.
func (ir *IRDropSimulator) ApplyMitigation(mit IRDropMitigation) {
	if mit.UseWidenedLines && mit.LineWidthIncrease > 1 {
		// Wider lines = lower resistance (R inversely proportional to width)
		ir.RowResist /= mit.LineWidthIncrease
		ir.ColResist /= mit.LineWidthIncrease
	}

	if mit.UseHierarchical {
		// Simple hierarchical bus model: reduce effective series resistance.
		ir.RowResist *= 0.5
		ir.ColResist *= 0.5
	}

	if mit.UseTiledArray && mit.TileSize > 0 {
		// Tiling reduces the effective wire length seen by any cell.
		maxDim := ir.Rows
		if ir.Cols > maxDim {
			maxDim = ir.Cols
		}
		factor := float64(mit.TileSize) / float64(maxDim)
		if factor > 0 && factor < 1 {
			ir.RowResist *= factor
			ir.ColResist *= factor
		}
	}

	// Re-simulate with new parameters
	ir.Simulate(100)
}
