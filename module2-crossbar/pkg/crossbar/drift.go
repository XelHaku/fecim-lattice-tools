package crossbar

import (
	"math"
	"math/rand"
)

// DriftModel specifies the drift coefficient source.
type DriftModel int

const (
	// DriftModelAssumed uses the default assumed coefficient (0.001)
	// WARNING: No peer-reviewed source for this exact value.
	DriftModelAssumed DriftModel = iota
	// DriftModelLiterature uses coefficients derived from literature
	// Based on HZO FeFET retention studies showing >10 year retention at RT.
	DriftModelLiterature
	// DriftModelMeasured allows user-specified coefficients from calibration
	DriftModelMeasured
)

// FeFETDriftCoefficients contains documented drift parameters.
// Source notes:
//   - FeFET retention: >10 years at 85°C demonstrated (Fraunhofer IPMS 2024)
//   - IEEE IRPS 2022: 10^9 cycle endurance with minimal drift
//   - Nano Letters 2024: V:HfO₂ shows 10^12 endurance
//
// The exact drift coefficient is estimated from retention requirements:
// For 10-year retention with <1 level drift, coefficient must be <0.001.
// This is a DERIVED estimate, not a directly measured value.
var FeFETDriftCoefficients = struct {
	Assumed    float64 // Conservative estimate (no direct measurement)
	Literature float64 // Derived from retention requirements
	RRAM       float64 // RRAM comparison (much higher)
	PCM        float64 // PCM comparison (highest)
	Flash      float64 // Flash comparison
}{
	Assumed:    0.001,  // ⚠️ ASSUMED VALUE - no direct peer-reviewed measurement
	Literature: 0.0005, // Derived: assumes <0.5 level drift over 10 years
	RRAM:       0.05,   // Literature: RRAM drift typically 5-10%
	PCM:        0.1,    // Literature: PCM has highest drift
	Flash:      0.02,   // Literature: Flash moderate drift
}

// DriftSimulator models conductance drift over time in memory devices.
// While FeCIM (ferroelectric) has much better retention than
// other technologies, we model small drift effects for completeness.
//
// IMPORTANT: The FeFET drift coefficient is estimated, not directly measured.
// Literature demonstrates excellent retention (>10 years) but does not
// provide explicit drift coefficients in the same format as RRAM/PCM papers.
// See FeFETDriftCoefficients for sources and assumptions.
type DriftSimulator struct {
	Rows         int         // Number of rows
	Cols         int         // Number of columns
	Conductances [][]float64 // Current conductance matrix (S)
	InitialConds [][]float64 // Initial conductance values
	Levels       int         // Number of discrete levels
	GMin         float64     // Minimum conductance (S)
	GMax         float64     // Maximum conductance (S)

	// Drift parameters
	DriftCoeff  float64    // Drift coefficient (see FeFETDriftCoefficients for sources)
	DriftModel  DriftModel // Source of drift coefficient
	ReadDisturb float64    // Read disturb probability per read
	Temperature float64    // Operating temperature (K)
	Time        float64    // Elapsed time (seconds)

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
	getLog().Input("NewDriftSimulator", map[string]interface{}{
		"rows":   rows,
		"cols":   cols,
		"levels": levels,
	})

	conductances := make([][]float64, rows)
	initialConds := make([][]float64, rows)

	gMin := GMin   // Use package constant
	gMax := GMax   // Use package constant

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

	sim := &DriftSimulator{
		Rows:         rows,
		Cols:         cols,
		Conductances: conductances,
		InitialConds: initialConds,
		Levels:       levels,
		GMin:         gMin,
		GMax:         gMax,
		// ⚠️ ASSUMED VALUE: FeFET drift coefficient 0.001 is estimated from retention requirements.
		// No direct peer-reviewed measurement exists for this exact value.
		// Derived from: >10 year retention at 85°C requires coefficient <0.001.
		// Compare to literature values: RRAM ~0.05, PCM ~0.1, Flash ~0.02.
		// See FeFETDriftCoefficients for detailed source notes.
		DriftCoeff:   FeFETDriftCoefficients.Assumed,
		DriftModel:   DriftModelAssumed,
		ReadDisturb:  1e-6, // Very low read disturb for FeFET
		Temperature:  300,  // Room temperature
		Time:         0,
		DriftHistory: make([]DriftSnapshot, 0),
	}

	getLog().Output("NewDriftSimulator", sim)
	return sim
}

// NewDriftSimulatorWithModel creates a drift simulator with specified coefficient source.
func NewDriftSimulatorWithModel(rows, cols int, levels int, model DriftModel) *DriftSimulator {
	sim := NewDriftSimulator(rows, cols, levels)
	sim.SetDriftModel(model)
	return sim
}

// SetDriftModel changes the drift coefficient source and updates the coefficient.
func (d *DriftSimulator) SetDriftModel(model DriftModel) {
	d.DriftModel = model
	switch model {
	case DriftModelLiterature:
		d.DriftCoeff = FeFETDriftCoefficients.Literature
	case DriftModelMeasured:
		// Keep existing coefficient (user should set it separately)
	case DriftModelAssumed:
		fallthrough
	default:
		d.DriftCoeff = FeFETDriftCoefficients.Assumed
	}
}

// SetMeasuredDriftCoeff sets a user-specified drift coefficient.
// Use this when you have calibration data from real device measurements.
func (d *DriftSimulator) SetMeasuredDriftCoeff(coeff float64) {
	d.DriftModel = DriftModelMeasured
	d.DriftCoeff = coeff
}

// GetDriftModelInfo returns information about the current drift model.
type DriftModelInfo struct {
	Model       DriftModel
	ModelName   string
	Coefficient float64
	IsAssumed   bool   // True if this value is estimated, not measured
	SourceNote  string // Citation or derivation note
}

// GetDriftModelInfo returns metadata about the current drift model configuration.
func (d *DriftSimulator) GetDriftModelInfo() *DriftModelInfo {
	info := &DriftModelInfo{
		Model:       d.DriftModel,
		Coefficient: d.DriftCoeff,
	}

	switch d.DriftModel {
	case DriftModelLiterature:
		info.ModelName = "Literature-Derived"
		info.IsAssumed = true // Still derived, not directly measured
		info.SourceNote = "Derived from HZO FeFET >10 year retention requirement. " +
			"Sources: Fraunhofer IPMS 2024, IEEE IRPS 2022."
	case DriftModelMeasured:
		info.ModelName = "User-Measured"
		info.IsAssumed = false
		info.SourceNote = "User-provided coefficient from device calibration."
	case DriftModelAssumed:
		fallthrough
	default:
		info.ModelName = "Assumed (Default)"
		info.IsAssumed = true
		info.SourceNote = "⚠️ ASSUMED VALUE: No direct peer-reviewed measurement. " +
			"Estimated to be ~50x better than RRAM based on retention studies."
	}

	return info
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

// SimulateTimeStep advances simulation by dt seconds.
func (d *DriftSimulator) SimulateTimeStep(dt float64) {
	getLog().Calculation("SimulateTimeStep", map[string]interface{}{
		"dt":          dt,
		"currentTime": d.Time,
		"driftCoeff":  d.DriftCoeff,
		"temperature": d.Temperature,
	}, nil)

	d.Time += dt

	// Thermal activation factor
	kB := 1.38e-23 // Boltzmann constant
	Ea := 0.5      // Activation energy (eV) - typical for FeFET
	eV := 1.6e-19  // eV to J
	thermalFactor := math.Exp(-Ea * eV / (kB * d.Temperature))

	var maxDrift float64
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

				if math.Abs(drift) > maxDrift {
					maxDrift = math.Abs(drift)
				}
			}
		}
	}

	getLog().Calculation("SimulateTimeStep", map[string]interface{}{
		"newTime":  d.Time,
		"maxDrift": maxDrift,
	}, nil)
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
	// ⚠️ IMPORTANT: FeFET drift coefficient 0.001 is ASSUMED (derived from retention requirements)
	// No direct peer-reviewed measurement exists. This comparison is illustrative only.
	// Sources:
	//   - FeFET retention: >10 years at 85°C (Fraunhofer IPMS 2024) → implies low drift
	//   - RRAM drift: Literature reports 5-10% drift (various sources)
	//   - PCM drift: Highest among emerging memories (resistance drift well documented)
	//   - Flash: Moderate drift, well characterized in industry
	comparison := TechDriftComparison{
		FeFETDrift:     d.DriftCoeff, // Current configured value (may be assumed or measured)
		RRAMDrift:      FeFETDriftCoefficients.RRAM,
		PCMDrift:       FeFETDriftCoefficients.PCM,
		FlashDrift:     FeFETDriftCoefficients.Flash,
		FeFETAdvantage: FeFETDriftCoefficients.RRAM / d.DriftCoeff, // Advantage vs RRAM
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
