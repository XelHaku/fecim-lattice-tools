// Package layers provides calibration and transformer CIM simulation
// for IronLattice ferroelectric compute-in-memory systems.
//
// This module implements:
// - Write-verify calibration algorithms
// - Multi-tile residual learning
// - Closed-loop conductance programming
// - Gain cell memory for KV cache
// - Transformer attention accelerators
// - Softmax approximation units
//
// Based on research from:
// - Science 2024: Programming memristor arrays with arbitrarily high precision
// - arXiv 2025: Multi-tile residual learning for limited conductance states
// - Nature Computational Science 2025: Analog IMC attention mechanism
// - IEEE 2025: KV-CIM sparse attention accelerator
package layers

import (
	"math"
	"sync"
)

// =============================================================================
// Write-Verify Calibration Algorithms
// =============================================================================

// WriteVerifyConfig configures write-verify calibration parameters
type WriteVerifyConfig struct {
	TargetConductance  float64 // Target conductance in Siemens
	Tolerance          float64 // Acceptable tolerance (e.g., 0.02 = 2%)
	MaxIterations      int     // Maximum write-verify iterations
	PulseVoltageSet    float64 // SET pulse voltage
	PulseVoltageReset  float64 // RESET pulse voltage
	PulseWidth         float64 // Pulse duration in nanoseconds
	VerifyVoltage      float64 // Verify read voltage
	ConductanceMin     float64 // Minimum programmable conductance
	ConductanceMax     float64 // Maximum programmable conductance
	RelaxationAware    bool    // Enable relaxation-aware programming
	RelaxationMargin   float64 // Extra margin for drift compensation
}

// DefaultWriteVerifyConfig returns standard configuration
func DefaultWriteVerifyConfig() *WriteVerifyConfig {
	return &WriteVerifyConfig{
		TargetConductance:  50e-6,   // 50 µS
		Tolerance:          0.02,    // 2%
		MaxIterations:      20,
		PulseVoltageSet:    2.0,     // 2V SET
		PulseVoltageReset:  -1.5,    // -1.5V RESET
		PulseWidth:         100.0,   // 100 ns
		VerifyVoltage:      0.2,     // 200 mV read
		ConductanceMin:     1e-6,    // 1 µS LRS bound
		ConductanceMax:     100e-6,  // 100 µS HRS bound
		RelaxationAware:    true,
		RelaxationMargin:   0.1,     // 10% extra margin
	}
}

// WriteVerifyCalibrator implements closed-loop conductance programming
type WriteVerifyCalibrator struct {
	Config            *WriteVerifyConfig
	CurrentConductance float64
	IterationCount     int
	Converged          bool

	// Statistics
	TotalPulses        int
	SetPulses          int
	ResetPulses        int
	AverageError       float64
	FinalError         float64

	// Relaxation tracking
	InitialConductance float64
	RelaxedConductance float64
	RelaxationDrift    float64  // Percentage drift
}

// NewWriteVerifyCalibrator creates a new calibrator
func NewWriteVerifyCalibrator(config *WriteVerifyConfig) *WriteVerifyCalibrator {
	if config == nil {
		config = DefaultWriteVerifyConfig()
	}
	return &WriteVerifyCalibrator{
		Config: config,
	}
}

// Calibrate performs write-verify calibration to target conductance
func (wvc *WriteVerifyCalibrator) Calibrate(initialConductance float64) (*ConductanceResult, error) {
	wvc.CurrentConductance = initialConductance
	wvc.InitialConductance = initialConductance
	wvc.IterationCount = 0
	wvc.Converged = false

	target := wvc.Config.TargetConductance

	// Apply relaxation margin if enabled
	if wvc.Config.RelaxationAware {
		// Overshoot target to compensate for drift
		if wvc.CurrentConductance < target {
			target *= (1.0 + wvc.Config.RelaxationMargin)
		} else {
			target *= (1.0 - wvc.Config.RelaxationMargin)
		}
	}

	totalError := 0.0

	for i := 0; i < wvc.Config.MaxIterations; i++ {
		wvc.IterationCount++

		// Verify current conductance
		error := math.Abs(wvc.CurrentConductance-wvc.Config.TargetConductance) / wvc.Config.TargetConductance
		totalError += error

		// Check convergence
		if error <= wvc.Config.Tolerance {
			wvc.Converged = true
			wvc.FinalError = error
			break
		}

		// Apply appropriate pulse
		if wvc.CurrentConductance < wvc.Config.TargetConductance {
			// Need to increase conductance - SET operation
			wvc.applySetPulse()
			wvc.SetPulses++
		} else {
			// Need to decrease conductance - RESET operation
			wvc.applyResetPulse()
			wvc.ResetPulses++
		}
		wvc.TotalPulses++
	}

	wvc.AverageError = totalError / float64(wvc.IterationCount)
	wvc.FinalError = math.Abs(wvc.CurrentConductance-wvc.Config.TargetConductance) / wvc.Config.TargetConductance

	// Simulate relaxation effect
	if wvc.Config.RelaxationAware {
		wvc.simulateRelaxation()
	}

	return &ConductanceResult{
		FinalConductance:  wvc.CurrentConductance,
		TargetConductance: wvc.Config.TargetConductance,
		Error:             wvc.FinalError,
		Iterations:        wvc.IterationCount,
		Converged:         wvc.Converged,
	}, nil
}

func (wvc *WriteVerifyCalibrator) applySetPulse() {
	// Simplified conductance increase model
	// Based on gap reduction during SET
	delta := (wvc.Config.ConductanceMax - wvc.CurrentConductance) * 0.15
	wvc.CurrentConductance += delta

	// Clamp to valid range
	if wvc.CurrentConductance > wvc.Config.ConductanceMax {
		wvc.CurrentConductance = wvc.Config.ConductanceMax
	}
}

func (wvc *WriteVerifyCalibrator) applyResetPulse() {
	// Simplified conductance decrease model
	// Based on gap formation during RESET
	delta := (wvc.CurrentConductance - wvc.Config.ConductanceMin) * 0.12
	wvc.CurrentConductance -= delta

	// Clamp to valid range
	if wvc.CurrentConductance < wvc.Config.ConductanceMin {
		wvc.CurrentConductance = wvc.Config.ConductanceMin
	}
}

func (wvc *WriteVerifyCalibrator) simulateRelaxation() {
	// Model conductance relaxation/drift over time
	// Write termination can cause ~50% more drift
	driftFactor := 0.03 // 3% baseline drift

	if !wvc.Config.RelaxationAware {
		driftFactor = 0.05 // 5% without compensation
	}

	wvc.RelaxedConductance = wvc.CurrentConductance * (1.0 - driftFactor)
	wvc.RelaxationDrift = driftFactor * 100
}

// ConductanceResult stores calibration results
type ConductanceResult struct {
	FinalConductance  float64
	TargetConductance float64
	Error             float64
	Iterations        int
	Converged         bool
}

// =============================================================================
// Self-Terminating Write (STW)
// =============================================================================

// STWConfig configures self-terminating write parameters
type STWConfig struct {
	TargetCurrent     float64 // Target current for termination
	VoltageRamp       float64 // Voltage ramp rate V/ns
	MaxPulseWidth     float64 // Maximum pulse duration
	PrecisionBits     int     // Target precision (typically 2-bit for STW)
	TemperatureC      float64 // Operating temperature
}

// SelfTerminatingWrite implements STW for fast programming
type SelfTerminatingWrite struct {
	Config           *STWConfig
	ProgrammedStates []float64
	ActualPulseWidth float64
	EnergyConsumed   float64  // pJ
	SpeedupFactor    float64  // vs traditional W&V
}

// NewSelfTerminatingWrite creates a new STW controller
func NewSelfTerminatingWrite(config *STWConfig) *SelfTerminatingWrite {
	if config == nil {
		config = &STWConfig{
			TargetCurrent:  100e-6,  // 100 µA
			VoltageRamp:    0.1,     // 0.1 V/ns
			MaxPulseWidth:  1000,    // 1 µs max
			PrecisionBits:  2,       // 2-bit STW
			TemperatureC:   25,
		}
	}
	return &SelfTerminatingWrite{
		Config:        config,
		SpeedupFactor: 4.7, // STW typically 4.7x faster than W&V
	}
}

// Program performs self-terminating write to target state
func (stw *SelfTerminatingWrite) Program(targetState int) (float64, error) {
	// Calculate target conductance from state
	numStates := 1 << stw.Config.PrecisionBits
	stepSize := (100e-6 - 1e-6) / float64(numStates-1)
	targetConductance := 1e-6 + float64(targetState)*stepSize

	// Simulate STW - terminates when target current reached
	// Simplified model: pulse width proportional to conductance change needed
	stw.ActualPulseWidth = 50.0 + float64(targetState)*25.0 // ns

	if stw.ActualPulseWidth > stw.Config.MaxPulseWidth {
		stw.ActualPulseWidth = stw.Config.MaxPulseWidth
	}

	// Energy model: E = V^2 * G * t
	voltage := 2.0 // typical SET voltage
	stw.EnergyConsumed = voltage * voltage * targetConductance * stw.ActualPulseWidth * 1e-9 * 1e12 // pJ

	stw.ProgrammedStates = append(stw.ProgrammedStates, targetConductance)

	return targetConductance, nil
}

// =============================================================================
// Multi-Tile Residual Learning
// =============================================================================

// MultiTileConfig configures multi-tile residual learning
type MultiTileConfig struct {
	NumTiles          int     // Number of crossbar tiles
	TilePrecisionBits int     // Bits per tile (typically 4)
	ArraySize         int     // Crossbar array dimensions
	LearningRate      float64 // Base learning rate
	ResidualDecay     float64 // Residual error decay factor
}

// DefaultMultiTileConfig returns standard configuration
func DefaultMultiTileConfig() *MultiTileConfig {
	return &MultiTileConfig{
		NumTiles:          4,      // 4 tiles for 16-bit effective precision
		TilePrecisionBits: 4,      // 4-bit per tile (typical ReRAM limit)
		ArraySize:         64,     // 64x64 crossbars
		LearningRate:      0.01,
		ResidualDecay:     0.5,
	}
}

// MultiTileResidualLearner implements residual learning across tiles
type MultiTileResidualLearner struct {
	Config         *MultiTileConfig
	Tiles          []*CrossbarTile
	ResidualErrors [][]float64
	EffectiveWeights [][]float64

	// Training metrics
	TrainingLoss   float64
	Iterations     int
	Accuracy       float64
}

// CrossbarTile represents a single analog crossbar array
type CrossbarTile struct {
	Weights       [][]float64
	Conductances  [][]float64
	QuantError    [][]float64
	TileIndex     int
	ScaleFactor   float64  // Weight scaling for this tile
}

// NewMultiTileResidualLearner creates a new multi-tile learner
func NewMultiTileResidualLearner(config *MultiTileConfig) *MultiTileResidualLearner {
	if config == nil {
		config = DefaultMultiTileConfig()
	}

	learner := &MultiTileResidualLearner{
		Config: config,
		Tiles:  make([]*CrossbarTile, config.NumTiles),
	}

	// Initialize tiles with decreasing scale factors
	for i := 0; i < config.NumTiles; i++ {
		scaleFactor := math.Pow(2.0, -float64(i*config.TilePrecisionBits))
		learner.Tiles[i] = &CrossbarTile{
			Weights:      make([][]float64, config.ArraySize),
			Conductances: make([][]float64, config.ArraySize),
			QuantError:   make([][]float64, config.ArraySize),
			TileIndex:    i,
			ScaleFactor:  scaleFactor,
		}

		for j := 0; j < config.ArraySize; j++ {
			learner.Tiles[i].Weights[j] = make([]float64, config.ArraySize)
			learner.Tiles[i].Conductances[j] = make([]float64, config.ArraySize)
			learner.Tiles[i].QuantError[j] = make([]float64, config.ArraySize)
		}
	}

	// Initialize residual errors
	learner.ResidualErrors = make([][]float64, config.ArraySize)
	learner.EffectiveWeights = make([][]float64, config.ArraySize)
	for i := 0; i < config.ArraySize; i++ {
		learner.ResidualErrors[i] = make([]float64, config.ArraySize)
		learner.EffectiveWeights[i] = make([]float64, config.ArraySize)
	}

	return learner
}

// UpdateWeights performs multi-tile weight update with residual compensation
func (mtrl *MultiTileResidualLearner) UpdateWeights(gradients [][]float64) error {
	arraySize := mtrl.Config.ArraySize
	numLevels := 1 << mtrl.Config.TilePrecisionBits

	// Compute target weight update
	targetUpdate := make([][]float64, arraySize)
	for i := 0; i < arraySize; i++ {
		targetUpdate[i] = make([]float64, arraySize)
		for j := 0; j < arraySize; j++ {
			targetUpdate[i][j] = -mtrl.Config.LearningRate * gradients[i][j]
		}
	}

	// Sequential tile programming with residual learning
	residual := targetUpdate

	for t := 0; t < mtrl.Config.NumTiles; t++ {
		tile := mtrl.Tiles[t]

		for i := 0; i < arraySize; i++ {
			for j := 0; j < arraySize; j++ {
				// Quantize residual to tile precision
				scaledValue := residual[i][j] / tile.ScaleFactor
				quantizedValue := quantizeToLevels(scaledValue, numLevels)

				// Update tile weights
				tile.Weights[i][j] += quantizedValue * tile.ScaleFactor

				// Compute quantization error for next tile
				tile.QuantError[i][j] = residual[i][j] - quantizedValue*tile.ScaleFactor

				// Propagate residual to next tile
				if t < mtrl.Config.NumTiles-1 {
					residual[i][j] = tile.QuantError[i][j]
				}
			}
		}
	}

	// Compute effective weights by summing all tiles
	for i := 0; i < arraySize; i++ {
		for j := 0; j < arraySize; j++ {
			mtrl.EffectiveWeights[i][j] = 0
			for t := 0; t < mtrl.Config.NumTiles; t++ {
				mtrl.EffectiveWeights[i][j] += mtrl.Tiles[t].Weights[i][j]
			}
			mtrl.ResidualErrors[i][j] = residual[i][j]
		}
	}

	mtrl.Iterations++
	return nil
}

func quantizeToLevels(value float64, numLevels int) float64 {
	// Uniform quantization
	step := 2.0 / float64(numLevels-1)
	clamped := math.Max(-1.0, math.Min(1.0, value))
	quantized := math.Round(clamped/step) * step
	return quantized
}

// ComputeEffectivePrecision calculates effective bits of precision
func (mtrl *MultiTileResidualLearner) ComputeEffectivePrecision() float64 {
	// Effective precision = numTiles * tilePrecision (approximately)
	// Actual is slightly less due to residual errors
	basePrecision := float64(mtrl.Config.NumTiles * mtrl.Config.TilePrecisionBits)

	// Compute average residual error
	var totalError float64
	count := 0
	for i := 0; i < mtrl.Config.ArraySize; i++ {
		for j := 0; j < mtrl.Config.ArraySize; j++ {
			totalError += math.Abs(mtrl.ResidualErrors[i][j])
			count++
		}
	}
	avgError := totalError / float64(count)

	// Precision loss due to residual
	precisionLoss := math.Log2(avgError + 1e-10)
	effectivePrecision := basePrecision + precisionLoss

	if effectivePrecision < 0 {
		effectivePrecision = 0
	}

	return effectivePrecision
}

// =============================================================================
// Gain Cell Memory for KV Cache
// =============================================================================

// GainCellConfig configures gain cell memory parameters
type GainCellConfig struct {
	NumRows           int     // Number of rows
	NumCols           int     // Number of columns
	RetentionTimeMs   float64 // Data retention time
	ReadBandwidthGBps float64 // Read bandwidth
	WriteBandwidthGBps float64 // Write bandwidth
	CapacitanceFf     float64 // Storage capacitance in fF
	LeakagePA         float64 // Leakage current in pA
	RefreshPeriodMs   float64 // Refresh period
	EnergyPerAccessPJ float64 // Energy per read/write
}

// DefaultGainCellConfig returns standard configuration
func DefaultGainCellConfig() *GainCellConfig {
	return &GainCellConfig{
		NumRows:           256,
		NumCols:           64,
		RetentionTimeMs:   100,    // 100 ms retention
		ReadBandwidthGBps: 100,    // 100 GB/s
		WriteBandwidthGBps: 50,    // 50 GB/s
		CapacitanceFf:     10,     // 10 fF storage cap
		LeakagePA:         2,      // 2 pA leakage (OS-FET)
		RefreshPeriodMs:   50,     // 50 ms refresh
		EnergyPerAccessPJ: 0.1,    // 0.1 pJ/access
	}
}

// GainCellMemory implements gain cell crossbar for KV cache
type GainCellMemory struct {
	Config         *GainCellConfig
	Storage        [][]float64  // Analog values (voltage levels)
	LastAccess     [][]int64    // Last access timestamp (ns)
	RefreshCounter int

	// Performance metrics
	TotalReads     int
	TotalWrites    int
	TotalRefreshes int
	EnergyConsumed float64  // pJ
}

// NewGainCellMemory creates a new gain cell memory array
func NewGainCellMemory(config *GainCellConfig) *GainCellMemory {
	if config == nil {
		config = DefaultGainCellConfig()
	}

	mem := &GainCellMemory{
		Config:     config,
		Storage:    make([][]float64, config.NumRows),
		LastAccess: make([][]int64, config.NumRows),
	}

	for i := 0; i < config.NumRows; i++ {
		mem.Storage[i] = make([]float64, config.NumCols)
		mem.LastAccess[i] = make([]int64, config.NumCols)
	}

	return mem
}

// WriteVector writes a vector to specified row
func (gcm *GainCellMemory) WriteVector(row int, values []float64) error {
	for i := 0; i < len(values) && i < gcm.Config.NumCols; i++ {
		gcm.Storage[row][i] = values[i]
		gcm.LastAccess[row][i] = 0 // Reset timestamp
	}
	gcm.TotalWrites++
	gcm.EnergyConsumed += gcm.Config.EnergyPerAccessPJ * float64(len(values))
	return nil
}

// ReadVector reads a vector from specified row
func (gcm *GainCellMemory) ReadVector(row int) []float64 {
	result := make([]float64, gcm.Config.NumCols)
	copy(result, gcm.Storage[row])
	gcm.TotalReads++
	gcm.EnergyConsumed += gcm.Config.EnergyPerAccessPJ * float64(gcm.Config.NumCols)
	return result
}

// DotProduct computes dot product in-memory (analog computation)
func (gcm *GainCellMemory) DotProduct(query []float64, row int) float64 {
	var sum float64
	for i := 0; i < len(query) && i < gcm.Config.NumCols; i++ {
		sum += query[i] * gcm.Storage[row][i]
	}
	gcm.EnergyConsumed += gcm.Config.EnergyPerAccessPJ * 0.5 // More efficient than read
	return sum
}

// ParallelDotProducts computes multiple dot products in parallel
func (gcm *GainCellMemory) ParallelDotProducts(query []float64) []float64 {
	results := make([]float64, gcm.Config.NumRows)

	// Simulate parallel analog computation
	for row := 0; row < gcm.Config.NumRows; row++ {
		for col := 0; col < gcm.Config.NumCols && col < len(query); col++ {
			results[row] += query[col] * gcm.Storage[row][col]
		}
	}

	// Energy: single parallel operation
	gcm.EnergyConsumed += gcm.Config.EnergyPerAccessPJ * float64(gcm.Config.NumCols)

	return results
}

// Refresh performs refresh to maintain data integrity
func (gcm *GainCellMemory) Refresh() {
	gcm.TotalRefreshes++
	// Rewrite all stored values (simplified)
	gcm.EnergyConsumed += gcm.Config.EnergyPerAccessPJ * float64(gcm.Config.NumRows*gcm.Config.NumCols) * 0.1
}

// =============================================================================
// KV Cache CIM Accelerator
// =============================================================================

// KVCacheConfig configures KV cache accelerator
type KVCacheConfig struct {
	MaxSequenceLength int
	HeadDim           int
	NumHeads          int
	NumLayers         int
	PrecisionBits     int
	UseGainCells      bool
	SparseThreshold   float64  // Attention sparsity threshold
}

// DefaultKVCacheConfig returns standard configuration
func DefaultKVCacheConfig() *KVCacheConfig {
	return &KVCacheConfig{
		MaxSequenceLength: 4096,
		HeadDim:           64,
		NumHeads:          32,
		NumLayers:         32,
		PrecisionBits:     16,
		UseGainCells:      true,
		SparseThreshold:   0.01,
	}
}

// KVCacheCIM implements in-memory KV cache storage and computation
type KVCacheCIM struct {
	Config       *KVCacheConfig
	KeyCache     []*GainCellMemory   // Per-layer, per-head
	ValueCache   []*GainCellMemory
	CacheLength  []int               // Current length per layer

	// Performance tracking
	mu               sync.Mutex
	TotalTokens      int
	AttentionOps     int
	EnergyEfficiency float64  // GOPS/W
}

// NewKVCacheCIM creates a new KV cache accelerator
func NewKVCacheCIM(config *KVCacheConfig) *KVCacheCIM {
	if config == nil {
		config = DefaultKVCacheConfig()
	}

	totalCaches := config.NumLayers * config.NumHeads

	kvc := &KVCacheCIM{
		Config:      config,
		KeyCache:    make([]*GainCellMemory, totalCaches),
		ValueCache:  make([]*GainCellMemory, totalCaches),
		CacheLength: make([]int, config.NumLayers),
	}

	// Initialize gain cell memories for each head
	gcConfig := &GainCellConfig{
		NumRows:           config.MaxSequenceLength,
		NumCols:           config.HeadDim,
		RetentionTimeMs:   100,
		EnergyPerAccessPJ: 0.1,
	}

	for i := 0; i < totalCaches; i++ {
		kvc.KeyCache[i] = NewGainCellMemory(gcConfig)
		kvc.ValueCache[i] = NewGainCellMemory(gcConfig)
	}

	return kvc
}

// AppendKV appends new key-value pairs to cache
func (kvc *KVCacheCIM) AppendKV(layer int, keys, values [][]float64) error {
	kvc.mu.Lock()
	defer kvc.mu.Unlock()

	for head := 0; head < kvc.Config.NumHeads; head++ {
		cacheIdx := layer*kvc.Config.NumHeads + head
		pos := kvc.CacheLength[layer]

		for i := 0; i < len(keys); i++ {
			kvc.KeyCache[cacheIdx].WriteVector(pos+i, keys[i])
			kvc.ValueCache[cacheIdx].WriteVector(pos+i, values[i])
		}
	}

	kvc.CacheLength[layer] += len(keys)
	kvc.TotalTokens += len(keys)

	return nil
}

// ComputeAttention computes attention scores and weighted values
func (kvc *KVCacheCIM) ComputeAttention(layer int, queries [][]float64) ([][]float64, error) {
	kvc.mu.Lock()
	defer kvc.mu.Unlock()

	seqLen := kvc.CacheLength[layer]
	numQueries := len(queries)

	// Output: weighted sum of values
	outputs := make([][]float64, numQueries)
	for i := range outputs {
		outputs[i] = make([]float64, kvc.Config.HeadDim)
	}

	for head := 0; head < kvc.Config.NumHeads; head++ {
		cacheIdx := layer*kvc.Config.NumHeads + head

		for q := 0; q < numQueries; q++ {
			// Compute attention scores via parallel dot products (in-memory)
			scores := kvc.KeyCache[cacheIdx].ParallelDotProducts(queries[q])

			// Scale by sqrt(d_k)
			scale := 1.0 / math.Sqrt(float64(kvc.Config.HeadDim))
			for i := 0; i < seqLen; i++ {
				scores[i] *= scale
			}

			// Apply sparse threshold
			for i := 0; i < seqLen; i++ {
				if scores[i] < kvc.Config.SparseThreshold {
					scores[i] = 0
				}
			}

			// Softmax (computed digitally or with approximation)
			softmaxScores := softmax(scores[:seqLen])

			// Weighted sum of values (in-memory)
			for i := 0; i < seqLen; i++ {
				if softmaxScores[i] > 0 {
					values := kvc.ValueCache[cacheIdx].ReadVector(i)
					for j := 0; j < kvc.Config.HeadDim; j++ {
						outputs[q][j] += softmaxScores[i] * values[j]
					}
				}
			}
		}
	}

	kvc.AttentionOps += numQueries
	return outputs, nil
}

func softmax(scores []float64) []float64 {
	if len(scores) == 0 {
		return scores
	}

	// Find max for numerical stability
	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}

	// Compute exp and sum
	result := make([]float64, len(scores))
	var sum float64
	for i, s := range scores {
		result[i] = math.Exp(s - maxScore)
		sum += result[i]
	}

	// Normalize
	if sum > 0 {
		for i := range result {
			result[i] /= sum
		}
	}

	return result
}

// =============================================================================
// Softmax Approximation Unit
// =============================================================================

// SoftmaxApproxConfig configures softmax approximation
type SoftmaxApproxConfig struct {
	Method       string  // "lut", "polynomial", "piecewise"
	LUTSize      int     // Lookup table entries
	PolyDegree   int     // Polynomial degree
	PrecisionBits int    // Output precision
}

// SoftmaxApproxUnit implements hardware-efficient softmax
type SoftmaxApproxUnit struct {
	Config    *SoftmaxApproxConfig
	LUT       []float64  // Lookup table for exp()
	PolyCoeffs []float64 // Polynomial coefficients

	// Metrics
	MaxError  float64
	AvgError  float64
	EnergyPJ  float64
}

// NewSoftmaxApproxUnit creates a new softmax approximation unit
func NewSoftmaxApproxUnit(config *SoftmaxApproxConfig) *SoftmaxApproxUnit {
	if config == nil {
		config = &SoftmaxApproxConfig{
			Method:        "lut",
			LUTSize:       256,
			PolyDegree:    3,
			PrecisionBits: 8,
		}
	}

	sau := &SoftmaxApproxUnit{
		Config: config,
	}

	// Initialize LUT for exp() approximation
	sau.LUT = make([]float64, config.LUTSize)
	for i := 0; i < config.LUTSize; i++ {
		x := -8.0 + 16.0*float64(i)/float64(config.LUTSize-1)
		sau.LUT[i] = math.Exp(x)
	}

	// Polynomial coefficients for exp(x) approximation: 1 + x + x^2/2 + x^3/6
	sau.PolyCoeffs = []float64{1.0, 1.0, 0.5, 0.166667}

	return sau
}

// Compute applies softmax with approximation
func (sau *SoftmaxApproxUnit) Compute(scores []float64) []float64 {
	switch sau.Config.Method {
	case "lut":
		return sau.computeLUT(scores)
	case "polynomial":
		return sau.computePolynomial(scores)
	case "piecewise":
		return sau.computePiecewise(scores)
	default:
		return softmax(scores)
	}
}

func (sau *SoftmaxApproxUnit) computeLUT(scores []float64) []float64 {
	if len(scores) == 0 {
		return scores
	}

	// Find max
	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}

	// LUT-based exp
	result := make([]float64, len(scores))
	var sum float64
	for i, s := range scores {
		x := s - maxScore
		// Map to LUT index
		idx := int((x + 8.0) / 16.0 * float64(sau.Config.LUTSize-1))
		if idx < 0 {
			idx = 0
		}
		if idx >= sau.Config.LUTSize {
			idx = sau.Config.LUTSize - 1
		}
		result[i] = sau.LUT[idx]
		sum += result[i]
	}

	// Normalize
	if sum > 0 {
		for i := range result {
			result[i] /= sum
		}
	}

	sau.EnergyPJ += 0.05 * float64(len(scores)) // LUT is efficient
	return result
}

func (sau *SoftmaxApproxUnit) computePolynomial(scores []float64) []float64 {
	if len(scores) == 0 {
		return scores
	}

	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}

	result := make([]float64, len(scores))
	var sum float64
	for i, s := range scores {
		x := s - maxScore
		// Polynomial approximation: exp(x) ≈ 1 + x + x^2/2 + x^3/6
		expApprox := sau.PolyCoeffs[0]
		xPow := 1.0
		for j := 1; j < len(sau.PolyCoeffs); j++ {
			xPow *= x
			expApprox += sau.PolyCoeffs[j] * xPow
		}
		if expApprox < 0 {
			expApprox = 0
		}
		result[i] = expApprox
		sum += result[i]
	}

	if sum > 0 {
		for i := range result {
			result[i] /= sum
		}
	}

	sau.EnergyPJ += 0.1 * float64(len(scores)) // More compute than LUT
	return result
}

func (sau *SoftmaxApproxUnit) computePiecewise(scores []float64) []float64 {
	if len(scores) == 0 {
		return scores
	}

	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}

	result := make([]float64, len(scores))
	var sum float64
	for i, s := range scores {
		x := s - maxScore
		// Piecewise linear approximation
		var expApprox float64
		if x > 0 {
			expApprox = 1.0 + x
		} else if x > -1 {
			expApprox = 1.0 + x + 0.5*x*x
		} else if x > -3 {
			expApprox = math.Exp(-1) * (1.0 + (x + 1.0))
		} else {
			expApprox = 0.05 // Floor value
		}
		if expApprox < 0 {
			expApprox = 0
		}
		result[i] = expApprox
		sum += result[i]
	}

	if sum > 0 {
		for i := range result {
			result[i] /= sum
		}
	}

	sau.EnergyPJ += 0.03 * float64(len(scores)) // Very efficient
	return result
}

// =============================================================================
// Transformer CIM Accelerator
// =============================================================================

// TransformerCIMConfig configures transformer accelerator
type TransformerCIMConfig struct {
	HiddenDim         int
	FFNDim            int
	NumHeads          int
	NumLayers         int
	MaxSequenceLength int
	UseAnalogAttention bool
	UseAnalogFFN      bool
	KVCacheEnabled    bool
	SparsityRatio     float64  // Expected attention sparsity
}

// DefaultTransformerCIMConfig returns standard configuration
func DefaultTransformerCIMConfig() *TransformerCIMConfig {
	return &TransformerCIMConfig{
		HiddenDim:         768,
		FFNDim:            3072,
		NumHeads:          12,
		NumLayers:         12,
		MaxSequenceLength: 4096,
		UseAnalogAttention: true,
		UseAnalogFFN:      true,
		KVCacheEnabled:    true,
		SparsityRatio:     0.9,  // 90% sparse attention
	}
}

// TransformerCIMAccelerator implements full transformer acceleration
type TransformerCIMAccelerator struct {
	Config         *TransformerCIMConfig
	KVCache        *KVCacheCIM
	FFNCrossbars   []*MultiTileResidualLearner
	QKVProjections []*MultiTileResidualLearner
	SoftmaxUnit    *SoftmaxApproxUnit

	// Performance metrics
	TotalTokens          int
	InferenceLatencyUs   float64
	EnergyPerTokenPJ     float64
	EffectiveTOPSW       float64
}

// NewTransformerCIMAccelerator creates a new transformer accelerator
func NewTransformerCIMAccelerator(config *TransformerCIMConfig) *TransformerCIMAccelerator {
	if config == nil {
		config = DefaultTransformerCIMConfig()
	}

	accel := &TransformerCIMAccelerator{
		Config:         config,
		FFNCrossbars:   make([]*MultiTileResidualLearner, config.NumLayers),
		QKVProjections: make([]*MultiTileResidualLearner, config.NumLayers*3),
	}

	// Initialize KV cache
	if config.KVCacheEnabled {
		kvConfig := &KVCacheConfig{
			MaxSequenceLength: config.MaxSequenceLength,
			HeadDim:           config.HiddenDim / config.NumHeads,
			NumHeads:          config.NumHeads,
			NumLayers:         config.NumLayers,
			UseGainCells:      true,
		}
		accel.KVCache = NewKVCacheCIM(kvConfig)
	}

	// Initialize multi-tile arrays for linear layers
	mtConfig := &MultiTileConfig{
		NumTiles:          4,
		TilePrecisionBits: 4,
		ArraySize:         64,
	}

	for i := 0; i < config.NumLayers; i++ {
		accel.FFNCrossbars[i] = NewMultiTileResidualLearner(mtConfig)
		// Q, K, V projections
		for j := 0; j < 3; j++ {
			accel.QKVProjections[i*3+j] = NewMultiTileResidualLearner(mtConfig)
		}
	}

	// Initialize softmax unit
	accel.SoftmaxUnit = NewSoftmaxApproxUnit(&SoftmaxApproxConfig{
		Method:        "lut",
		LUTSize:       256,
		PrecisionBits: 8,
	})

	return accel
}

// ProcessToken processes a single token through the transformer
func (tca *TransformerCIMAccelerator) ProcessToken(input []float64) ([]float64, error) {
	hidden := input

	for layer := 0; layer < tca.Config.NumLayers; layer++ {
		// Self-attention
		attentionOutput, err := tca.processAttentionLayer(layer, hidden)
		if err != nil {
			return nil, err
		}

		// Residual connection + LayerNorm (simplified)
		for i := range hidden {
			hidden[i] = (hidden[i] + attentionOutput[i]) / 2.0
		}

		// FFN
		ffnOutput := tca.processFFNLayer(layer, hidden)

		// Residual connection
		for i := range hidden {
			hidden[i] = (hidden[i] + ffnOutput[i]) / 2.0
		}
	}

	tca.TotalTokens++
	return hidden, nil
}

func (tca *TransformerCIMAccelerator) processAttentionLayer(layer int, hidden []float64) ([]float64, error) {
	// Simplified: use KV cache attention
	if tca.Config.KVCacheEnabled && tca.KVCache != nil {
		// Generate Q, K, V from hidden
		query := make([][]float64, 1)
		query[0] = hidden[:tca.Config.HiddenDim/tca.Config.NumHeads]

		outputs, err := tca.KVCache.ComputeAttention(layer, query)
		if err != nil {
			return nil, err
		}

		// Pad output to hidden dim
		result := make([]float64, tca.Config.HiddenDim)
		for i := range outputs[0] {
			if i < len(result) {
				result[i] = outputs[0][i]
			}
		}
		return result, nil
	}

	return hidden, nil
}

func (tca *TransformerCIMAccelerator) processFFNLayer(layer int, hidden []float64) []float64 {
	// Two-layer FFN with GELU
	// Up-projection: hidden_dim -> ffn_dim
	intermediate := make([]float64, tca.Config.FFNDim)
	for i := range intermediate {
		// Simplified linear transformation
		intermediate[i] = hidden[i%len(hidden)] * 0.5
	}

	// GELU activation
	for i := range intermediate {
		intermediate[i] = gelu(intermediate[i])
	}

	// Down-projection: ffn_dim -> hidden_dim
	output := make([]float64, tca.Config.HiddenDim)
	for i := range output {
		output[i] = intermediate[i%len(intermediate)] * 0.5
	}

	return output
}

func gelu(x float64) float64 {
	return 0.5 * x * (1.0 + math.Tanh(math.Sqrt(2.0/math.Pi)*(x+0.044715*x*x*x)))
}

// GetPerformanceMetrics returns performance statistics
func (tca *TransformerCIMAccelerator) GetPerformanceMetrics() map[string]float64 {
	// Estimated metrics based on literature
	// X-Former: 85x latency, 7.5x energy vs GPU
	// Nature 2025: 70,000x energy, 100x speed vs GPU
	// MCBP: 22,740 GOPS/W (31x vs A100)

	metrics := make(map[string]float64)
	metrics["total_tokens"] = float64(tca.TotalTokens)

	// Estimate based on analog CIM benefits
	baseLatencyPerToken := 10.0 // µs baseline for digital
	if tca.Config.UseAnalogAttention {
		metrics["latency_per_token_us"] = baseLatencyPerToken / 85.0
	} else {
		metrics["latency_per_token_us"] = baseLatencyPerToken
	}

	// Energy estimates
	baseEnergyPerToken := 1000.0 // pJ baseline
	if tca.Config.UseAnalogAttention && tca.Config.UseAnalogFFN {
		metrics["energy_per_token_pj"] = baseEnergyPerToken / 70.0
	} else {
		metrics["energy_per_token_pj"] = baseEnergyPerToken
	}

	// TOPS/W estimate
	metrics["effective_tops_w"] = 22740.0 * (1.0 - tca.Config.SparsityRatio*0.5)

	// KV cache memory footprint
	if tca.Config.KVCacheEnabled {
		// Memory per token = 2 * num_layers * hidden_dim * precision_bytes
		bytesPerToken := 2 * tca.Config.NumLayers * tca.Config.HiddenDim * 2 // FP16
		metrics["kv_cache_bytes_per_token"] = float64(bytesPerToken)
		metrics["kv_cache_total_mb"] = float64(tca.Config.MaxSequenceLength*bytesPerToken) / (1024 * 1024)
	}

	return metrics
}

// =============================================================================
// X-Former Hybrid Architecture
// =============================================================================

// XFormerConfig configures X-Former hybrid accelerator
type XFormerConfig struct {
	NVMCrossbars     int     // Number of NVM crossbars for static ops
	CMOSProcessors   int     // Number of CMOS PEs for dynamic ops
	CrossbarSize     int     // NVM crossbar dimensions
	PEArraySize      int     // CMOS PE array size
	StaticWeightRatio float64 // Fraction of weights in NVM
}

// XFormerAccelerator implements hybrid NVM-CMOS architecture
type XFormerAccelerator struct {
	Config           *XFormerConfig
	NVMArrays        []*MultiTileResidualLearner
	CMOSPipeline     *CMOSProcessingPipeline

	// Performance
	LatencySpeedup   float64  // vs GPU (target: 85x)
	EnergySpeedup    float64  // vs GPU (target: 7.5x)
}

// CMOSProcessingPipeline handles dynamic operations
type CMOSProcessingPipeline struct {
	NumPEs           int
	PEUtilization    float64
	DynamicOps       []string
	EnergyPerOpPJ    float64
}

// NewXFormerAccelerator creates X-Former accelerator
func NewXFormerAccelerator(config *XFormerConfig) *XFormerAccelerator {
	if config == nil {
		config = &XFormerConfig{
			NVMCrossbars:     16,
			CMOSProcessors:   8,
			CrossbarSize:     128,
			PEArraySize:      16,
			StaticWeightRatio: 0.7,
		}
	}

	accel := &XFormerAccelerator{
		Config:    config,
		NVMArrays: make([]*MultiTileResidualLearner, config.NVMCrossbars),
	}

	// Initialize NVM arrays
	mtConfig := &MultiTileConfig{
		NumTiles:          2,
		TilePrecisionBits: 4,
		ArraySize:         config.CrossbarSize,
	}
	for i := 0; i < config.NVMCrossbars; i++ {
		accel.NVMArrays[i] = NewMultiTileResidualLearner(mtConfig)
	}

	// Initialize CMOS pipeline
	accel.CMOSPipeline = &CMOSProcessingPipeline{
		NumPEs:        config.CMOSProcessors,
		PEUtilization: 0.85,
		DynamicOps:    []string{"softmax", "layernorm", "gelu", "elementwise"},
		EnergyPerOpPJ: 1.0,
	}

	// Target speedups from literature
	accel.LatencySpeedup = 85.0
	accel.EnergySpeedup = 7.5

	return accel
}

// ProcessLayer processes one transformer layer through X-Former
func (xf *XFormerAccelerator) ProcessLayer(input [][]float64) ([][]float64, error) {
	// Static operations (linear projections) -> NVM crossbars
	// Dynamic operations (softmax, layernorm) -> CMOS PEs

	// Simplified: pass through
	return input, nil
}

// =============================================================================
// Array Calibration Controller
// =============================================================================

// ArrayCalibrationConfig configures full-array calibration
type ArrayCalibrationConfig struct {
	ArrayRows         int
	ArrayCols         int
	TargetPrecision   int     // bits
	CalibrationMode   string  // "sequential", "parallel", "adaptive"
	MaxCalibrationTime float64 // seconds
	ErrorThreshold    float64
}

// ArrayCalibrationController manages array-wide calibration
type ArrayCalibrationController struct {
	Config            *ArrayCalibrationConfig
	CellCalibrators   [][]*WriteVerifyCalibrator
	CalibrationStatus [][]bool

	// Metrics
	TotalCells       int
	CalibratedCells  int
	FailedCells      int
	AverageError     float64
	CalibrationTime  float64  // seconds
}

// NewArrayCalibrationController creates array calibration controller
func NewArrayCalibrationController(config *ArrayCalibrationConfig) *ArrayCalibrationController {
	if config == nil {
		config = &ArrayCalibrationConfig{
			ArrayRows:         64,
			ArrayCols:         64,
			TargetPrecision:   4,
			CalibrationMode:   "adaptive",
			MaxCalibrationTime: 10.0,
			ErrorThreshold:    0.05,
		}
	}

	acc := &ArrayCalibrationController{
		Config:            config,
		CellCalibrators:   make([][]*WriteVerifyCalibrator, config.ArrayRows),
		CalibrationStatus: make([][]bool, config.ArrayRows),
		TotalCells:        config.ArrayRows * config.ArrayCols,
	}

	for i := 0; i < config.ArrayRows; i++ {
		acc.CellCalibrators[i] = make([]*WriteVerifyCalibrator, config.ArrayCols)
		acc.CalibrationStatus[i] = make([]bool, config.ArrayCols)

		for j := 0; j < config.ArrayCols; j++ {
			acc.CellCalibrators[i][j] = NewWriteVerifyCalibrator(DefaultWriteVerifyConfig())
		}
	}

	return acc
}

// CalibrateArray performs full array calibration
func (acc *ArrayCalibrationController) CalibrateArray(targetWeights [][]float64) error {
	var wg sync.WaitGroup
	var mu sync.Mutex

	totalError := 0.0
	calibrated := 0
	failed := 0

	for i := 0; i < acc.Config.ArrayRows; i++ {
		for j := 0; j < acc.Config.ArrayCols; j++ {
			wg.Add(1)
			go func(row, col int) {
				defer wg.Done()

				// Convert weight to conductance
				weight := targetWeights[row][col]
				targetG := weightToConductance(weight)

				// Update calibrator target
				acc.CellCalibrators[row][col].Config.TargetConductance = targetG

				// Calibrate cell
				initialG := 10e-6 // Start from some initial state
				result, err := acc.CellCalibrators[row][col].Calibrate(initialG)

				mu.Lock()
				defer mu.Unlock()

				if err != nil || !result.Converged {
					failed++
					acc.CalibrationStatus[row][col] = false
				} else {
					calibrated++
					acc.CalibrationStatus[row][col] = true
					totalError += result.Error
				}
			}(i, j)
		}
	}

	wg.Wait()

	acc.CalibratedCells = calibrated
	acc.FailedCells = failed
	if calibrated > 0 {
		acc.AverageError = totalError / float64(calibrated)
	}

	return nil
}

func weightToConductance(weight float64) float64 {
	// Map weight [-1, 1] to conductance [1µS, 100µS]
	normalized := (weight + 1.0) / 2.0 // [0, 1]
	conductance := 1e-6 + normalized*99e-6
	return conductance
}

// GetYield returns calibration yield percentage
func (acc *ArrayCalibrationController) GetYield() float64 {
	return float64(acc.CalibratedCells) / float64(acc.TotalCells) * 100
}

// =============================================================================
// Performance Estimator
// =============================================================================

// PerformanceEstimator estimates CIM accelerator performance
type PerformanceEstimator struct {
	ModelParams      int64    // Total model parameters
	SequenceLength   int
	BatchSize        int
	Precision        int      // bits
	ArrayEfficiency  float64  // crossbar utilization
	AnalogSpeedup    float64  // vs digital baseline
	MemoryBandwidth  float64  // GB/s
}

// EstimateLLMInference estimates LLM inference performance
func (pe *PerformanceEstimator) EstimateLLMInference() map[string]float64 {
	results := make(map[string]float64)

	// Memory requirements
	kvCacheSize := float64(pe.SequenceLength) * float64(pe.BatchSize) * float64(pe.Precision/8) * 2
	results["kv_cache_size_gb"] = kvCacheSize / 1e9

	// Compute requirements (MACs per token)
	macsPerToken := float64(pe.ModelParams) * 2 // ~2 MACs per param
	results["macs_per_token"] = macsPerToken

	// Throughput estimation
	// MCBP: 22,740 GOPS/W
	// Nature 2025 IMC: 70,000x energy reduction
	topsW := 22740.0 * pe.ArrayEfficiency
	results["effective_tops_w"] = topsW

	// Latency per token
	// X-Former: 85x speedup vs GPU
	baseLatencyMs := macsPerToken / (100e12) * 1000 // 100 TOPS baseline
	analogLatencyMs := baseLatencyMs / pe.AnalogSpeedup
	results["latency_per_token_ms"] = analogLatencyMs

	// Tokens per second
	if analogLatencyMs > 0 {
		results["tokens_per_second"] = 1000.0 / analogLatencyMs
	}

	// Energy per token
	energyPerMac := 1.0 / (topsW * 1e12) * 1e12 // pJ
	results["energy_per_token_mj"] = macsPerToken * energyPerMac / 1e9

	return results
}
