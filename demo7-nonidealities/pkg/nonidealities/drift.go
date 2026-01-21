package nonidealities

import (
	"math"
	"math/rand"
)

// DriftSimulator models conductance drift over time in memory devices.
// While FeCIM (ferroelectric) has much better retention than
// other technologies, we model small drift effects for completeness.
type DriftSimulator struct {
	Rows         int         // Number of rows
	Cols         int         // Number of columns
	Conductances [][]float64 // Current conductance matrix (S)
	InitialConds [][]float64 // Initial conductance values
	Levels       int         // Number of discrete levels
	GMin         float64     // Minimum conductance (S)
	GMax         float64     // Maximum conductance (S)

	// Drift parameters
	DriftCoeff  float64 // Drift coefficient (typically 0.01-0.1 for RRAM, much lower for FeFET)
	ReadDisturb float64 // Read disturb probability per read
	Temperature float64 // Operating temperature (K)
	Time        float64 // Elapsed time (seconds)

	// Statistics
	DriftHistory []DriftSnapshot
}

// DriftSnapshot captures state at a point in time.
type DriftSnapshot struct {
	Time            float64   // Time point
	AvgDrift        float64   // Average drift from initial
	MaxDrift        float64   // Maximum drift
	NumLevelChanges int       // Number of cells that changed level
	WorstCellRow    int       // Row of worst-drifted cell
	WorstCellCol    int       // Column of worst-drifted cell
	Conductances    []float64 // Sample of conductances (first row for brevity)
}

// NewDriftSimulator creates a drift simulator.
func NewDriftSimulator(rows, cols int, levels int) *DriftSimulator {
	conductances := make([][]float64, rows)
	initialConds := make([][]float64, rows)

	gMin := 1e-6   // 1µS minimum
	gMax := 100e-6 // 100µS maximum

	for i := range conductances {
		conductances[i] = make([]float64, cols)
		initialConds[i] = make([]float64, cols)
		for j := range conductances[i] {
			// Initialize to random level
			level := rand.Intn(levels)
			g := gMin + (gMax-gMin)*float64(level)/float64(levels-1)
			conductances[i][j] = g
			initialConds[i][j] = g
		}
	}

	return &DriftSimulator{
		Rows:         rows,
		Cols:         cols,
		Conductances: conductances,
		InitialConds: initialConds,
		Levels:       levels,
		GMin:         gMin,
		GMax:         gMax,
		DriftCoeff:   0.001, // Very low for FeFET (0.001 vs 0.05 for RRAM)
		ReadDisturb:  1e-6,  // Very low read disturb for FeFET
		Temperature:  300,   // Room temperature
		Time:         0,
		DriftHistory: make([]DriftSnapshot, 0),
	}
}

// SetConductanceLevel sets a cell to a specific discrete level.
func (d *DriftSimulator) SetConductanceLevel(row, col, level int) {
	if row >= 0 && row < d.Rows && col >= 0 && col < d.Cols {
		if level < 0 {
			level = 0
		}
		if level >= d.Levels {
			level = d.Levels - 1
		}
		g := d.GMin + (d.GMax-d.GMin)*float64(level)/float64(d.Levels-1)
		d.Conductances[row][col] = g
		d.InitialConds[row][col] = g
	}
}

// SetWeightMatrix sets conductances from a weight matrix.
func (d *DriftSimulator) SetWeightMatrix(weights [][]int) {
	for i := 0; i < d.Rows && i < len(weights); i++ {
		for j := 0; j < d.Cols && j < len(weights[i]); j++ {
			level := weights[i][j]
			if level < 0 {
				level = 0
			}
			if level >= d.Levels {
				level = d.Levels - 1
			}
			g := d.GMin + (d.GMax-d.GMin)*float64(level)/float64(d.Levels-1)
			d.Conductances[i][j] = g
			d.InitialConds[i][j] = g
		}
	}
}

// SimulateTimeStep advances simulation by dt seconds.
func (d *DriftSimulator) SimulateTimeStep(dt float64) {
	d.Time += dt

	// Thermal activation factor
	kB := 1.38e-23 // Boltzmann constant
	Ea := 0.5      // Activation energy (eV) - typical for FeFET
	eV := 1.6e-19  // eV to J
	thermalFactor := math.Exp(-Ea * eV / (kB * d.Temperature))

	for i := 0; i < d.Rows; i++ {
		for j := 0; j < d.Cols; j++ {
			// Drift model: G(t) = G0 * (t/t0)^v
			// For FeFET, v (drift coefficient) is very small
			g0 := d.InitialConds[i][j]

			// Time-dependent drift with thermal activation
			// Using log model for stability
			if d.Time > 0 {
				logT := math.Log(d.Time + 1) // +1 to avoid log(0)
				drift := g0 * d.DriftCoeff * logT * thermalFactor

				// Add random component (device-to-device variation)
				drift += (rand.Float64() - 0.5) * 0.001 * g0 * thermalFactor

				d.Conductances[i][j] = g0 + drift

				// Clamp to valid range
				if d.Conductances[i][j] < d.GMin {
					d.Conductances[i][j] = d.GMin
				}
				if d.Conductances[i][j] > d.GMax {
					d.Conductances[i][j] = d.GMax
				}
			}
		}
	}
}

// SimulateRead simulates read disturb on a cell.
func (d *DriftSimulator) SimulateRead(row, col int, numReads int) {
	if row < 0 || row >= d.Rows || col < 0 || col >= d.Cols {
		return
	}

	for n := 0; n < numReads; n++ {
		if rand.Float64() < d.ReadDisturb {
			// Small conductance change due to read disturb
			change := (rand.Float64() - 0.5) * 0.0001 * d.Conductances[row][col]
			d.Conductances[row][col] += change

			if d.Conductances[row][col] < d.GMin {
				d.Conductances[row][col] = d.GMin
			}
			if d.Conductances[row][col] > d.GMax {
				d.Conductances[row][col] = d.GMax
			}
		}
	}
}

// RecordSnapshot records current state.
func (d *DriftSimulator) RecordSnapshot() {
	avgDrift := 0.0
	maxDrift := 0.0
	numLevelChanges := 0
	worstRow, worstCol := 0, 0

	levelWidth := (d.GMax - d.GMin) / float64(d.Levels-1)

	for i := 0; i < d.Rows; i++ {
		for j := 0; j < d.Cols; j++ {
			drift := math.Abs(d.Conductances[i][j] - d.InitialConds[i][j])
			avgDrift += drift
			if drift > maxDrift {
				maxDrift = drift
				worstRow = i
				worstCol = j
			}

			// Check if level changed
			initialLevel := int((d.InitialConds[i][j] - d.GMin) / levelWidth)
			currentLevel := int((d.Conductances[i][j] - d.GMin) / levelWidth)
			if initialLevel != currentLevel {
				numLevelChanges++
			}
		}
	}
	avgDrift /= float64(d.Rows * d.Cols)

	// Sample first row conductances
	sample := make([]float64, d.Cols)
	copy(sample, d.Conductances[0])

	snapshot := DriftSnapshot{
		Time:            d.Time,
		AvgDrift:        avgDrift,
		MaxDrift:        maxDrift,
		NumLevelChanges: numLevelChanges,
		WorstCellRow:    worstRow,
		WorstCellCol:    worstCol,
		Conductances:    sample,
	}
	d.DriftHistory = append(d.DriftHistory, snapshot)
}

// GetCurrentLevel returns the current quantized level for a cell.
func (d *DriftSimulator) GetCurrentLevel(row, col int) int {
	if row < 0 || row >= d.Rows || col < 0 || col >= d.Cols {
		return 0
	}

	g := d.Conductances[row][col]
	level := int((g-d.GMin)/(d.GMax-d.GMin)*float64(d.Levels-1) + 0.5)
	if level < 0 {
		level = 0
	}
	if level >= d.Levels {
		level = d.Levels - 1
	}
	return level
}

// GetInitialLevel returns the initial quantized level for a cell.
func (d *DriftSimulator) GetInitialLevel(row, col int) int {
	if row < 0 || row >= d.Rows || col < 0 || col >= d.Cols {
		return 0
	}

	g := d.InitialConds[row][col]
	level := int((g-d.GMin)/(d.GMax-d.GMin)*float64(d.Levels-1) + 0.5)
	if level < 0 {
		level = 0
	}
	if level >= d.Levels {
		level = d.Levels - 1
	}
	return level
}

// DriftStats contains statistics about drift.
type DriftStats struct {
	ElapsedTime          float64 // Time elapsed (s)
	AvgDrift             float64 // Average conductance drift (S)
	MaxDrift             float64 // Maximum conductance drift (S)
	AvgDriftPercent      float64 // Average drift as percentage
	MaxDriftPercent      float64 // Maximum drift as percentage
	NumLevelErrors       int     // Number of cells with level errors
	LevelErrorRate       float64 // Percentage of cells with errors
	RetentionPrediction  float64 // Predicted 10-year retention (%)
	TechnologyComparison TechDriftComparison
}

// TechDriftComparison compares drift across technologies.
type TechDriftComparison struct {
	FeFETDrift     float64 // FeCIM (FeFET) drift coefficient
	RRAMDrift      float64 // RRAM drift coefficient
	PCMDrift       float64 // Phase-change memory drift coefficient
	FlashDrift     float64 // Flash memory drift coefficient
	FeFETAdvantage float64 // FeFET advantage factor
}

// GetStats returns drift statistics.
func (d *DriftSimulator) GetStats() DriftStats {
	avgDrift := 0.0
	maxDrift := 0.0
	numLevelErrors := 0
	levelWidth := (d.GMax - d.GMin) / float64(d.Levels-1)

	for i := 0; i < d.Rows; i++ {
		for j := 0; j < d.Cols; j++ {
			drift := math.Abs(d.Conductances[i][j] - d.InitialConds[i][j])
			avgDrift += drift
			if drift > maxDrift {
				maxDrift = drift
			}

			initialLevel := int((d.InitialConds[i][j] - d.GMin) / levelWidth)
			currentLevel := int((d.Conductances[i][j] - d.GMin) / levelWidth)
			if initialLevel != currentLevel {
				numLevelErrors++
			}
		}
	}
	avgDrift /= float64(d.Rows * d.Cols)

	avgGMid := (d.GMax + d.GMin) / 2
	avgDriftPercent := avgDrift / avgGMid * 100
	maxDriftPercent := maxDrift / avgGMid * 100

	levelErrorRate := float64(numLevelErrors) / float64(d.Rows*d.Cols) * 100

	// Estimate 10-year retention
	// For FeFET, retention is excellent (>10 years at room temperature)
	tenYearSeconds := 10 * 365.25 * 24 * 3600
	logTenYear := math.Log(tenYearSeconds)
	kB := 1.38e-23
	Ea := 0.5
	eV := 1.6e-19
	thermalFactor := math.Exp(-Ea * eV / (kB * d.Temperature))
	predictedDrift := d.DriftCoeff * logTenYear * thermalFactor
	retention := (1 - predictedDrift) * 100
	if retention > 100 {
		retention = 99.99
	}
	if retention < 0 {
		retention = 0
	}

	// Technology comparison
	comparison := TechDriftComparison{
		FeFETDrift:     0.001,        // Very low for FeFET
		RRAMDrift:      0.05,         // Higher for RRAM
		PCMDrift:       0.1,          // Higher for PCM
		FlashDrift:     0.02,         // Medium for Flash
		FeFETAdvantage: 0.05 / 0.001, // 50x better than RRAM
	}

	return DriftStats{
		ElapsedTime:          d.Time,
		AvgDrift:             avgDrift,
		MaxDrift:             maxDrift,
		AvgDriftPercent:      avgDriftPercent,
		MaxDriftPercent:      maxDriftPercent,
		NumLevelErrors:       numLevelErrors,
		LevelErrorRate:       levelErrorRate,
		RetentionPrediction:  retention,
		TechnologyComparison: comparison,
	}
}

// RefreshCell refreshes a cell to its nearest valid level.
func (d *DriftSimulator) RefreshCell(row, col int) {
	if row < 0 || row >= d.Rows || col < 0 || col >= d.Cols {
		return
	}

	// Get current level and set to exact value
	level := d.GetCurrentLevel(row, col)
	g := d.GMin + (d.GMax-d.GMin)*float64(level)/float64(d.Levels-1)
	d.Conductances[row][col] = g
	d.InitialConds[row][col] = g
}

// RefreshAll refreshes all cells to their nearest valid levels.
func (d *DriftSimulator) RefreshAll() {
	for i := 0; i < d.Rows; i++ {
		for j := 0; j < d.Cols; j++ {
			d.RefreshCell(i, j)
		}
	}
}

// Reset resets all cells to their initial values.
func (d *DriftSimulator) Reset() {
	for i := 0; i < d.Rows; i++ {
		for j := 0; j < d.Cols; j++ {
			d.Conductances[i][j] = d.InitialConds[i][j]
		}
	}
	d.Time = 0
	d.DriftHistory = make([]DriftSnapshot, 0)
}

// CompareTechnologies runs drift simulation for different technologies.
func CompareTechnologies(rows, cols int, simulationTime float64) map[string]DriftStats {
	results := make(map[string]DriftStats)

	// Technology parameters
	technologies := map[string]struct {
		driftCoeff  float64
		readDisturb float64
	}{
		"FeCIM (FeFET)": {0.001, 1e-6},
		"RRAM":          {0.05, 1e-4},
		"PCM":           {0.1, 1e-3},
		"Flash":         {0.02, 1e-5},
	}

	for name, params := range technologies {
		sim := NewDriftSimulator(rows, cols, 30)
		sim.DriftCoeff = params.driftCoeff
		sim.ReadDisturb = params.readDisturb

		// Simulate
		numSteps := 100
		dt := simulationTime / float64(numSteps)
		for step := 0; step < numSteps; step++ {
			sim.SimulateTimeStep(dt)
		}

		results[name] = sim.GetStats()
	}

	return results
}
