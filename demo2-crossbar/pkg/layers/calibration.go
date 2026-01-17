// Package layers provides hardware calibration and fault tolerance utilities for CIM.
// Implements stuck-at fault handling, variation compensation, and reservoir computing.
package layers

import (
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// STUCK-AT FAULT HANDLING
// ============================================================================

// FaultType represents types of stuck-at faults.
type FaultType int

const (
	FaultNone FaultType = iota
	FaultStuckHigh       // Stuck at high resistance (low conductance)
	FaultStuckLow        // Stuck at low resistance (high conductance)
	FaultOpen            // Open circuit
	FaultShort           // Short circuit
)

// FaultMap stores fault locations in a crossbar array.
type FaultMap struct {
	Rows      int
	Cols      int
	Faults    map[int]map[int]FaultType // [row][col] -> fault type
	FaultRate float64                   // Overall fault rate
}

// NewFaultMap creates a new fault map.
func NewFaultMap(rows, cols int) *FaultMap {
	return &FaultMap{
		Rows:   rows,
		Cols:   cols,
		Faults: make(map[int]map[int]FaultType),
	}
}

// InjectRandomFaults adds random stuck-at faults.
func (f *FaultMap) InjectRandomFaults(rate float64, highLowRatio float64) {
	f.FaultRate = rate
	numFaults := int(float64(f.Rows*f.Cols) * rate)

	for i := 0; i < numFaults; i++ {
		row := rand.Intn(f.Rows)
		col := rand.Intn(f.Cols)

		if f.Faults[row] == nil {
			f.Faults[row] = make(map[int]FaultType)
		}

		// Determine fault type based on ratio
		if rand.Float64() < highLowRatio {
			f.Faults[row][col] = FaultStuckHigh
		} else {
			f.Faults[row][col] = FaultStuckLow
		}
	}
}

// GetFault returns the fault type at a position.
func (f *FaultMap) GetFault(row, col int) FaultType {
	if rowFaults, ok := f.Faults[row]; ok {
		if fault, ok := rowFaults[col]; ok {
			return fault
		}
	}
	return FaultNone
}

// CountFaults returns total number of faults.
func (f *FaultMap) CountFaults() int {
	count := 0
	for _, rowFaults := range f.Faults {
		count += len(rowFaults)
	}
	return count
}

// ============================================================================
// FAULT TOLERANCE STRATEGIES
// ============================================================================

// FaultToleranceConfig configures fault tolerance settings.
type FaultToleranceConfig struct {
	Strategy         string  // "remap", "redundancy", "training", "hybrid"
	RedundancyFactor float64 // Extra capacity for redundancy
	RetrainingEpochs int     // Epochs for fault-aware retraining
	DropConnectRate  float64 // Rate for drop-connect training
}

// DefaultFaultToleranceConfig returns standard settings.
func DefaultFaultToleranceConfig() *FaultToleranceConfig {
	return &FaultToleranceConfig{
		Strategy:         "hybrid",
		RedundancyFactor: 1.2,
		RetrainingEpochs: 10,
		DropConnectRate:  0.1,
	}
}

// FaultTolerantMapper maps weights to faulty crossbar.
type FaultTolerantMapper struct {
	Config   *FaultToleranceConfig
	FaultMap *FaultMap
}

// NewFaultTolerantMapper creates a new mapper.
func NewFaultTolerantMapper(config *FaultToleranceConfig, faultMap *FaultMap) *FaultTolerantMapper {
	return &FaultTolerantMapper{
		Config:   config,
		FaultMap: faultMap,
	}
}

// RemapWeights remaps weights avoiding faulty cells.
func (m *FaultTolerantMapper) RemapWeights(weights [][]float64) ([][]float64, *RemapInfo) {
	rows := len(weights)
	if rows == 0 {
		return weights, nil
	}
	cols := len(weights[0])

	info := &RemapInfo{
		OriginalRows: rows,
		OriginalCols: cols,
		Remappings:   make(map[int]map[int][2]int),
	}

	// Create output with extra rows/cols for remapping
	extraRows := int(float64(rows) * (m.Config.RedundancyFactor - 1))
	extraCols := int(float64(cols) * (m.Config.RedundancyFactor - 1))

	remapped := make([][]float64, rows+extraRows)
	for i := range remapped {
		remapped[i] = make([]float64, cols+extraCols)
	}

	// Track available spare locations
	spareRow := rows
	spareCol := cols

	// Map weights, remapping faulty cells
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			fault := m.FaultMap.GetFault(i, j)

			if fault == FaultNone {
				// No fault, direct mapping
				remapped[i][j] = weights[i][j]
			} else {
				// Find spare location
				newRow, newCol := m.findSpareLocation(spareRow, spareCol, extraRows, extraCols)
				remapped[newRow][newCol] = weights[i][j]

				// Record remapping
				if info.Remappings[i] == nil {
					info.Remappings[i] = make(map[int][2]int)
				}
				info.Remappings[i][j] = [2]int{newRow, newCol}

				// Handle stuck value
				if fault == FaultStuckHigh {
					remapped[i][j] = 0 // Min conductance
				} else if fault == FaultStuckLow {
					remapped[i][j] = 1 // Max conductance
				}
			}
		}
	}

	info.RemappedRows = rows + extraRows
	info.RemappedCols = cols + extraCols

	return remapped, info
}

// findSpareLocation finds an available spare cell.
func (m *FaultTolerantMapper) findSpareLocation(startRow, startCol, maxExtraRows, maxExtraCols int) (int, int) {
	// Simple linear search for spare location
	for r := startRow; r < startRow+maxExtraRows; r++ {
		for c := startCol; c < startCol+maxExtraCols; c++ {
			if m.FaultMap.GetFault(r, c) == FaultNone {
				return r, c
			}
		}
	}
	// Fallback
	return startRow, startCol
}

// RemapInfo contains remapping information.
type RemapInfo struct {
	OriginalRows int
	OriginalCols int
	RemappedRows int
	RemappedCols int
	Remappings   map[int]map[int][2]int // [orig_row][orig_col] -> [new_row, new_col]
}

// ============================================================================
// VARIATION COMPENSATION
// ============================================================================

// VariationConfig configures variation compensation.
type VariationConfig struct {
	D2DVariation    float64 // Device-to-device variation (σ)
	C2CVariation    float64 // Cycle-to-cycle variation (σ)
	TemperatureCoef float64 // Temperature coefficient (%/°C)
	DriftRate       float64 // Conductance drift rate (%/hour)
}

// DefaultVariationConfig returns typical variation parameters.
func DefaultVariationConfig() *VariationConfig {
	return &VariationConfig{
		D2DVariation:    0.05, // 5% device-to-device
		C2CVariation:    0.02, // 2% cycle-to-cycle
		TemperatureCoef: 0.01, // 1% per °C
		DriftRate:       0.001, // 0.1% per hour
	}
}

// VariationCompensator compensates for device variations.
type VariationCompensator struct {
	Config             *VariationConfig
	CalibrationWeights [][]float64 // Reference weights
	MeasuredWeights    [][]float64 // Actual measured weights
	CorrectionFactors  [][]float64 // Compensation factors
}

// NewVariationCompensator creates a new compensator.
func NewVariationCompensator(config *VariationConfig) *VariationCompensator {
	return &VariationCompensator{
		Config: config,
	}
}

// Calibrate measures and computes correction factors.
func (v *VariationCompensator) Calibrate(targetWeights, measuredWeights [][]float64) {
	rows := len(targetWeights)
	cols := len(targetWeights[0])

	v.CalibrationWeights = targetWeights
	v.MeasuredWeights = measuredWeights
	v.CorrectionFactors = make([][]float64, rows)

	for i := 0; i < rows; i++ {
		v.CorrectionFactors[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			if measuredWeights[i][j] != 0 {
				v.CorrectionFactors[i][j] = targetWeights[i][j] / measuredWeights[i][j]
			} else {
				v.CorrectionFactors[i][j] = 1.0
			}
		}
	}
}

// CompensateInput adjusts input to account for weight variation.
func (v *VariationCompensator) CompensateInput(input []float64) []float64 {
	if v.CorrectionFactors == nil {
		return input
	}

	compensated := make([]float64, len(input))
	copy(compensated, input)

	// Apply row-wise correction (simplified)
	for i := range compensated {
		if i < len(v.CorrectionFactors) {
			avgCorrection := 0.0
			for j := range v.CorrectionFactors[i] {
				avgCorrection += v.CorrectionFactors[i][j]
			}
			avgCorrection /= float64(len(v.CorrectionFactors[i]))
			compensated[i] *= avgCorrection
		}
	}

	return compensated
}

// SimulateVariation adds realistic variation to weights.
func (v *VariationCompensator) SimulateVariation(weights [][]float64) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])

	varied := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		varied[i] = make([]float64, cols)
		for j := 0; j < cols; j++ {
			// Device-to-device variation (fixed per device)
			d2d := 1 + rand.NormFloat64()*v.Config.D2DVariation

			// Cycle-to-cycle variation
			c2c := 1 + rand.NormFloat64()*v.Config.C2CVariation

			varied[i][j] = weights[i][j] * d2d * c2c
		}
	}

	return varied
}

// ============================================================================
// RESERVOIR COMPUTING
// ============================================================================

// ReservoirConfig configures FeFET reservoir computing.
type ReservoirConfig struct {
	NumNodes        int     // Virtual nodes in reservoir
	DelayTau        float64 // Delay line time constant (s)
	InputScaling    float64 // Input scaling factor
	FeedbackStrength float64 // Feedback connection strength
	SpectralRadius  float64 // Target spectral radius
	LeakRate        float64 // Leaky integrator rate
	WashoutSteps    int     // Initial steps to discard
}

// DefaultReservoirConfig returns standard reservoir settings.
func DefaultReservoirConfig() *ReservoirConfig {
	return &ReservoirConfig{
		NumNodes:         100,
		DelayTau:         0.001, // 1 ms
		InputScaling:     0.5,
		FeedbackStrength: 0.1,
		SpectralRadius:   0.9,
		LeakRate:         0.3,
		WashoutSteps:     100,
	}
}

// FeFETReservoir implements reservoir computing with FeFET dynamics.
type FeFETReservoir struct {
	Config          *ReservoirConfig
	InputWeights    [][]float64 // Win: input to reservoir
	ReservoirWeights [][]float64 // W: reservoir internal
	OutputWeights   [][]float64 // Wout: reservoir to output
	State           []float64   // Current reservoir state
	History         [][]float64 // State history for training
}

// NewFeFETReservoir creates a new reservoir.
func NewFeFETReservoir(config *ReservoirConfig, inputSize, outputSize int) *FeFETReservoir {
	r := &FeFETReservoir{
		Config: config,
		State:  make([]float64, config.NumNodes),
	}

	// Initialize input weights (sparse, random)
	r.InputWeights = make([][]float64, config.NumNodes)
	for i := 0; i < config.NumNodes; i++ {
		r.InputWeights[i] = make([]float64, inputSize)
		// Sparse connectivity (~10%)
		for j := 0; j < inputSize; j++ {
			if rand.Float64() < 0.1 {
				r.InputWeights[i][j] = (rand.Float64()*2 - 1) * config.InputScaling
			}
		}
	}

	// Initialize reservoir weights (sparse, scaled for spectral radius)
	r.ReservoirWeights = make([][]float64, config.NumNodes)
	for i := 0; i < config.NumNodes; i++ {
		r.ReservoirWeights[i] = make([]float64, config.NumNodes)
		for j := 0; j < config.NumNodes; j++ {
			if rand.Float64() < 0.1 { // 10% connectivity
				r.ReservoirWeights[i][j] = rand.Float64()*2 - 1
			}
		}
	}
	r.scaleSpectralRadius()

	// Output weights initialized later via training
	r.OutputWeights = make([][]float64, outputSize)
	for i := 0; i < outputSize; i++ {
		r.OutputWeights[i] = make([]float64, config.NumNodes)
	}

	return r
}

// scaleSpectralRadius scales reservoir weights to target spectral radius.
func (r *FeFETReservoir) scaleSpectralRadius() {
	// Compute approximate spectral radius using power iteration
	n := len(r.ReservoirWeights)
	v := make([]float64, n)
	for i := range v {
		v[i] = rand.Float64()
	}

	// Power iteration
	for iter := 0; iter < 100; iter++ {
		// Matrix-vector multiply
		newV := make([]float64, n)
		for i := 0; i < n; i++ {
			sum := 0.0
			for j := 0; j < n; j++ {
				sum += r.ReservoirWeights[i][j] * v[j]
			}
			newV[i] = sum
		}

		// Normalize
		norm := 0.0
		for _, val := range newV {
			norm += val * val
		}
		norm = math.Sqrt(norm)
		if norm > 0 {
			for i := range newV {
				newV[i] /= norm
			}
		}
		v = newV
	}

	// Estimate spectral radius
	Wv := make([]float64, n)
	for i := 0; i < n; i++ {
		sum := 0.0
		for j := 0; j < n; j++ {
			sum += r.ReservoirWeights[i][j] * v[j]
		}
		Wv[i] = sum
	}

	rho := 0.0
	for i := range Wv {
		rho += Wv[i] * v[i]
	}
	rho = math.Abs(rho)

	// Scale to target
	if rho > 0 {
		scale := r.Config.SpectralRadius / rho
		for i := range r.ReservoirWeights {
			for j := range r.ReservoirWeights[i] {
				r.ReservoirWeights[i][j] *= scale
			}
		}
	}
}

// Step advances reservoir by one time step.
func (r *FeFETReservoir) Step(input []float64) []float64 {
	n := r.Config.NumNodes

	// Compute new state: x(t+1) = (1-α)x(t) + α*tanh(Win*u + W*x)
	newState := make([]float64, n)

	for i := 0; i < n; i++ {
		// Input contribution
		inputSum := 0.0
		for j := range input {
			if j < len(r.InputWeights[i]) {
				inputSum += r.InputWeights[i][j] * input[j]
			}
		}

		// Recurrent contribution
		recurrentSum := 0.0
		for j := 0; j < n; j++ {
			recurrentSum += r.ReservoirWeights[i][j] * r.State[j]
		}

		// Leaky integration with tanh nonlinearity
		activation := math.Tanh(inputSum + recurrentSum)
		newState[i] = (1-r.Config.LeakRate)*r.State[i] + r.Config.LeakRate*activation
	}

	r.State = newState
	return newState
}

// Collect gathers reservoir states for training.
func (r *FeFETReservoir) Collect(inputs [][]float64) [][]float64 {
	r.History = make([][]float64, 0, len(inputs))

	// Reset state
	r.State = make([]float64, r.Config.NumNodes)

	for i, input := range inputs {
		state := r.Step(input)

		// Skip washout period
		if i >= r.Config.WashoutSteps {
			stateCopy := make([]float64, len(state))
			copy(stateCopy, state)
			r.History = append(r.History, stateCopy)
		}
	}

	return r.History
}

// TrainReadout trains the output layer using ridge regression.
func (r *FeFETReservoir) TrainReadout(targets [][]float64, lambda float64) {
	if len(r.History) == 0 || len(targets) == 0 {
		return
	}

	// Align lengths
	n := len(r.History)
	if n > len(targets) {
		n = len(targets)
	}

	states := r.History[:n]
	targetSlice := targets[:n]

	// Ridge regression: Wout = Y * X^T * (X * X^T + λI)^-1
	// Simplified: use pseudo-inverse with regularization

	numNodes := r.Config.NumNodes
	numOutputs := len(targetSlice[0])

	// Compute X^T * X + λI
	XtX := make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		XtX[i] = make([]float64, numNodes)
		for j := 0; j < numNodes; j++ {
			sum := 0.0
			for t := 0; t < n; t++ {
				sum += states[t][i] * states[t][j]
			}
			XtX[i][j] = sum
			if i == j {
				XtX[i][j] += lambda // Regularization
			}
		}
	}

	// Compute X^T * Y
	XtY := make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		XtY[i] = make([]float64, numOutputs)
		for k := 0; k < numOutputs; k++ {
			sum := 0.0
			for t := 0; t < n; t++ {
				sum += states[t][i] * targetSlice[t][k]
			}
			XtY[i][k] = sum
		}
	}

	// Solve (X^T*X + λI) * Wout^T = X^T * Y
	// Use simple iterative method for now
	WoutT := make([][]float64, numNodes)
	for i := 0; i < numNodes; i++ {
		WoutT[i] = make([]float64, numOutputs)
		for k := 0; k < numOutputs; k++ {
			WoutT[i][k] = XtY[i][k] / (XtX[i][i] + 1e-6)
		}
	}

	// Transpose to get Wout
	r.OutputWeights = make([][]float64, numOutputs)
	for k := 0; k < numOutputs; k++ {
		r.OutputWeights[k] = make([]float64, numNodes)
		for i := 0; i < numNodes; i++ {
			r.OutputWeights[k][i] = WoutT[i][k]
		}
	}
}

// Predict generates output from reservoir state.
func (r *FeFETReservoir) Predict() []float64 {
	numOutputs := len(r.OutputWeights)
	output := make([]float64, numOutputs)

	for k := 0; k < numOutputs; k++ {
		sum := 0.0
		for i := range r.State {
			if i < len(r.OutputWeights[k]) {
				sum += r.OutputWeights[k][i] * r.State[i]
			}
		}
		output[k] = sum
	}

	return output
}

// ============================================================================
// DELAY LINE RESERVOIR
// ============================================================================

// DelayLineConfig configures delay line reservoir.
type DelayLineConfig struct {
	NumTaps       int     // Number of delay taps
	TapSpacing    float64 // Time between taps (s)
	NonlinearMask []float64 // Mask values for nonlinearity
}

// DelayLineReservoir implements single-node delay line reservoir.
type DelayLineReservoir struct {
	Config      *DelayLineConfig
	DelayBuffer []float64 // Circular buffer
	BufferIdx   int
	MaskWeights []float64
}

// NewDelayLineReservoir creates a new delay line reservoir.
func NewDelayLineReservoir(config *DelayLineConfig) *DelayLineReservoir {
	r := &DelayLineReservoir{
		Config:      config,
		DelayBuffer: make([]float64, config.NumTaps),
		MaskWeights: make([]float64, config.NumTaps),
	}

	// Initialize mask (random or from config)
	if len(config.NonlinearMask) == config.NumTaps {
		copy(r.MaskWeights, config.NonlinearMask)
	} else {
		for i := 0; i < config.NumTaps; i++ {
			r.MaskWeights[i] = rand.Float64()*2 - 1
		}
	}

	return r
}

// Process runs input through delay line.
func (d *DelayLineReservoir) Process(input float64) []float64 {
	// Inject input with mask
	maskedInput := input * d.MaskWeights[d.BufferIdx]

	// Apply FeFET-like nonlinearity (polarization dynamics)
	d.DelayBuffer[d.BufferIdx] = math.Tanh(maskedInput + 0.5*d.DelayBuffer[(d.BufferIdx+d.Config.NumTaps-1)%d.Config.NumTaps])

	// Advance buffer
	d.BufferIdx = (d.BufferIdx + 1) % d.Config.NumTaps

	// Return all taps as reservoir state
	state := make([]float64, d.Config.NumTaps)
	for i := 0; i < d.Config.NumTaps; i++ {
		idx := (d.BufferIdx + i) % d.Config.NumTaps
		state[i] = d.DelayBuffer[idx]
	}

	return state
}

// ============================================================================
// IR-DROP COMPENSATION
// ============================================================================

// IRDropConfig configures IR-drop compensation.
type IRDropConfig struct {
	LineResistance float64 // Ohms per cell
	MaxCurrent     float64 // Maximum current (A)
	VDD            float64 // Supply voltage
}

// IRDropCompensator compensates for voltage drop in crossbar.
type IRDropCompensator struct {
	Config     *IRDropConfig
	DropMatrix [][]float64 // Estimated voltage drop per cell
}

// NewIRDropCompensator creates a new compensator.
func NewIRDropCompensator(config *IRDropConfig) *IRDropCompensator {
	return &IRDropCompensator{
		Config: config,
	}
}

// EstimateIRDrop computes voltage drop matrix for given weights.
func (c *IRDropCompensator) EstimateIRDrop(weights [][]float64, inputVoltages []float64) [][]float64 {
	rows := len(weights)
	cols := len(weights[0])

	c.DropMatrix = make([][]float64, rows)
	for i := 0; i < rows; i++ {
		c.DropMatrix[i] = make([]float64, cols)
	}

	// Simplified model: drop accumulates along wordlines and bitlines
	for i := 0; i < rows; i++ {
		cumulativeDrop := 0.0
		for j := 0; j < cols; j++ {
			// Current through cell: I = G * V
			conductance := weights[i][j]
			if conductance < 0 {
				conductance = -conductance
			}
			cellCurrent := conductance * inputVoltages[i]

			// Voltage drop: V_drop = I * R
			drop := cellCurrent * c.Config.LineResistance * float64(j)
			cumulativeDrop += drop

			c.DropMatrix[i][j] = cumulativeDrop
		}
	}

	return c.DropMatrix
}

// CompensateOutput adjusts output to account for IR-drop.
func (c *IRDropCompensator) CompensateOutput(output []float64, weights [][]float64) []float64 {
	if c.DropMatrix == nil {
		return output
	}

	compensated := make([]float64, len(output))
	for j := range output {
		correction := 1.0
		// Average correction factor for this column
		for i := range weights {
			if i < len(c.DropMatrix) && j < len(c.DropMatrix[i]) {
				dropFraction := c.DropMatrix[i][j] / c.Config.VDD
				correction += dropFraction
			}
		}
		correction /= float64(len(weights))
		compensated[j] = output[j] * (1 + correction)
	}

	return compensated
}

// ============================================================================
// CALIBRATION PIPELINE
// ============================================================================

// CalibrationPipeline orchestrates complete calibration.
type CalibrationPipeline struct {
	FaultMap       *FaultMap
	FaultMapper    *FaultTolerantMapper
	VarCompensator *VariationCompensator
	IRCompensator  *IRDropCompensator
}

// NewCalibrationPipeline creates a calibration pipeline.
func NewCalibrationPipeline(rows, cols int) *CalibrationPipeline {
	faultMap := NewFaultMap(rows, cols)
	return &CalibrationPipeline{
		FaultMap:       faultMap,
		FaultMapper:    NewFaultTolerantMapper(DefaultFaultToleranceConfig(), faultMap),
		VarCompensator: NewVariationCompensator(DefaultVariationConfig()),
		IRCompensator:  NewIRDropCompensator(&IRDropConfig{
			LineResistance: 1.0,
			MaxCurrent:     1e-6,
			VDD:            0.9,
		}),
	}
}

// Calibrate runs full calibration procedure.
func (p *CalibrationPipeline) Calibrate(targetWeights [][]float64, faultRate float64) (*CalibratedArray, error) {
	result := &CalibratedArray{
		OriginalWeights: targetWeights,
	}

	// Step 1: Inject faults (for simulation)
	p.FaultMap.InjectRandomFaults(faultRate, 0.5)
	result.FaultCount = p.FaultMap.CountFaults()

	// Step 2: Remap around faults
	remapped, remapInfo := p.FaultMapper.RemapWeights(targetWeights)
	result.RemappedWeights = remapped
	result.RemapInfo = remapInfo

	// Step 3: Simulate variation
	varied := p.VarCompensator.SimulateVariation(remapped)
	result.VariedWeights = varied

	// Step 4: Calibrate for variation
	p.VarCompensator.Calibrate(remapped, varied)
	result.CorrectionFactors = p.VarCompensator.CorrectionFactors

	// Step 5: Estimate IR-drop
	inputVoltages := make([]float64, len(targetWeights))
	for i := range inputVoltages {
		inputVoltages[i] = 0.5 // Example input
	}
	p.IRCompensator.EstimateIRDrop(varied, inputVoltages)
	result.IRDropMatrix = p.IRCompensator.DropMatrix

	return result, nil
}

// CalibratedArray contains calibration results.
type CalibratedArray struct {
	OriginalWeights   [][]float64
	RemappedWeights   [][]float64
	VariedWeights     [][]float64
	CorrectionFactors [][]float64
	IRDropMatrix      [][]float64
	FaultCount        int
	RemapInfo         *RemapInfo
}

// ============================================================================
// TIMING ANALYSIS
// ============================================================================

// TimingAnalyzer analyzes reservoir timing characteristics.
type TimingAnalyzer struct {
	RelaxationTimes []float64
	MemoryCapacity  float64
}

// AnalyzeMemoryCapacity computes short-term memory capacity.
func (t *TimingAnalyzer) AnalyzeMemoryCapacity(reservoir *FeFETReservoir, testLength int) float64 {
	// Generate random input sequence
	inputs := make([][]float64, testLength)
	for i := 0; i < testLength; i++ {
		inputs[i] = []float64{rand.Float64()*2 - 1}
	}

	// Collect states
	states := reservoir.Collect(inputs)

	// Compute memory capacity: sum of R² for delayed reconstructions
	mc := 0.0
	maxDelay := 50
	if maxDelay > len(states)-1 {
		maxDelay = len(states) - 1
	}

	for delay := 1; delay <= maxDelay; delay++ {
		// Compute correlation between state and delayed input
		r2 := t.computeDelayedR2(states, inputs[reservoir.Config.WashoutSteps:], delay)
		mc += r2
	}

	t.MemoryCapacity = mc
	return mc
}

// computeDelayedR2 computes R² for delay reconstruction.
func (t *TimingAnalyzer) computeDelayedR2(states [][]float64, inputs [][]float64, delay int) float64 {
	n := len(states) - delay
	if n <= 0 {
		return 0
	}

	// Simple linear regression using first state component
	var sumX, sumY, sumXY, sumX2, sumY2 float64
	for i := 0; i < n; i++ {
		x := states[i+delay][0]
		y := inputs[i][0]
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
		sumY2 += y * y
	}

	nf := float64(n)
	num := nf*sumXY - sumX*sumY
	den := math.Sqrt((nf*sumX2 - sumX*sumX) * (nf*sumY2 - sumY*sumY))

	if den == 0 {
		return 0
	}

	r := num / den
	return r * r
}

// MeasureRelaxationTime estimates relaxation time constants.
func (t *TimingAnalyzer) MeasureRelaxationTime(reservoir *FeFETReservoir) []float64 {
	// Impulse response
	impulse := []float64{1.0}
	zero := []float64{0.0}

	// Reset and apply impulse
	reservoir.State = make([]float64, reservoir.Config.NumNodes)
	reservoir.Step(impulse)

	// Collect decay
	decaySteps := 100
	decay := make([][]float64, decaySteps)
	for i := 0; i < decaySteps; i++ {
		state := reservoir.Step(zero)
		decay[i] = make([]float64, len(state))
		copy(decay[i], state)
	}

	// Estimate tau for each node (exponential fit)
	t.RelaxationTimes = make([]float64, reservoir.Config.NumNodes)
	for node := 0; node < reservoir.Config.NumNodes; node++ {
		// Find when state drops to 1/e
		initialVal := math.Abs(decay[0][node])
		if initialVal < 1e-10 {
			continue
		}
		threshold := initialVal / math.E

		for step := 1; step < decaySteps; step++ {
			if math.Abs(decay[step][node]) < threshold {
				t.RelaxationTimes[node] = float64(step) * reservoir.Config.DelayTau
				break
			}
		}
	}

	// Sort and return median
	sorted := make([]float64, len(t.RelaxationTimes))
	copy(sorted, t.RelaxationTimes)
	sort.Float64s(sorted)

	return t.RelaxationTimes
}
