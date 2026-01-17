// Package layers provides neural network layer implementations for CIM inference.
// attention_cim.go implements CIM-optimized attention mechanisms and power gating.
//
// References:
// - Nature Computational Science (2025): Analog attention, 70,000× energy reduction
// - Scientific Reports (2024): Memristor-based transformer accelerator
// - Hybrid analog-digital attention (2024): 14.8 TOPS/W, 75% token pruning
// - Sparse attention with LSH: O(L log L) complexity

package layers

import (
	"fmt"
	"math"
	"sort"
)

// =============================================================================
// CIM-Optimized Attention Mechanism
// =============================================================================

// CIMAttentionConfig configures CIM attention implementation
type CIMAttentionConfig struct {
	// Model dimensions
	HeadDim       int // Dimension per attention head
	NumHeads      int // Number of attention heads
	SeqLength     int // Maximum sequence length
	HiddenDim     int // Hidden dimension (HeadDim × NumHeads)

	// CIM hardware constraints
	CrossbarSize  int     // NxN crossbar array size
	WeightBits    int     // Weight quantization bits
	ADCBits       int     // ADC resolution
	NoiseLevel    float64 // Analog noise percentage

	// Optimization settings
	EnableKVCache     bool    // Enable KV cache in crossbar
	EnableTokenPruning bool   // Enable dynamic token pruning
	PruneRatio        float64 // Fraction of tokens to prune (0.75 = 75%)
	EnableSparseAttn  bool    // Enable sparse attention (LSH)
	LSHBuckets        int     // Number of LSH hash buckets

	// Power management
	EnablePowerGating bool    // Enable power gating for inactive arrays
	SleepThreshold    float64 // Activation threshold for sleep mode
}

// DefaultCIMAttentionConfig returns typical configuration
func DefaultCIMAttentionConfig() *CIMAttentionConfig {
	return &CIMAttentionConfig{
		HeadDim:          64,
		NumHeads:         8,
		SeqLength:        512,
		HiddenDim:        512,
		CrossbarSize:     64,
		WeightBits:       6,
		ADCBits:          6,
		NoiseLevel:       0.02,
		EnableKVCache:    true,
		EnableTokenPruning: true,
		PruneRatio:       0.75,
		EnableSparseAttn: false,
		LSHBuckets:       32,
		EnablePowerGating: true,
		SleepThreshold:   0.01,
	}
}

// CIMAttention implements compute-in-memory attention
type CIMAttention struct {
	Config *CIMAttentionConfig

	// Weight matrices stored in crossbar
	WQ [][]float64 // Query projection
	WK [][]float64 // Key projection
	WV [][]float64 // Value projection
	WO [][]float64 // Output projection

	// KV Cache (stored in gain-cell crossbar)
	KCache [][]float64 // [SeqLength][HiddenDim]
	VCache [][]float64 // [SeqLength][HiddenDim]
	CacheLen int

	// Power gating state
	ArrayPowerState []bool // true = active, false = sleeping
	PowerGater      *PowerGatingController
}

// NewCIMAttention creates a CIM-optimized attention layer
func NewCIMAttention(config *CIMAttentionConfig) *CIMAttention {
	if config == nil {
		config = DefaultCIMAttentionConfig()
	}

	attn := &CIMAttention{
		Config: config,
		KCache: make([][]float64, config.SeqLength),
		VCache: make([][]float64, config.SeqLength),
	}

	for i := 0; i < config.SeqLength; i++ {
		attn.KCache[i] = make([]float64, config.HiddenDim)
		attn.VCache[i] = make([]float64, config.HiddenDim)
	}

	// Initialize power gating
	numArrays := (config.HiddenDim * config.HiddenDim) / (config.CrossbarSize * config.CrossbarSize)
	attn.ArrayPowerState = make([]bool, numArrays)
	for i := range attn.ArrayPowerState {
		attn.ArrayPowerState[i] = true // Start all active
	}

	if config.EnablePowerGating {
		attn.PowerGater = NewPowerGatingController(&PowerGatingConfig{
			NumArrays:      numArrays,
			SleepThreshold: config.SleepThreshold,
			WakeupLatency:  10, // cycles
			SleepPower:     0.01, // fraction of active power
		})
	}

	return attn
}

// InitWeights initializes attention weight matrices
func (a *CIMAttention) InitWeights(wq, wk, wv, wo [][]float64) {
	a.WQ = wq
	a.WK = wk
	a.WV = wv
	a.WO = wo
}

// Forward computes attention with CIM optimizations
func (a *CIMAttention) Forward(input [][]float64) ([][]float64, *AttentionStats) {
	seqLen := len(input)
	stats := &AttentionStats{
		TotalTokens:   seqLen,
		ActiveArrays:  0,
		SleepingArrays: 0,
	}

	// Step 1: Project Q, K, V using crossbar MVM
	Q := a.projectQKV(input, a.WQ)
	K := a.projectQKV(input, a.WK)
	V := a.projectQKV(input, a.WV)

	// Step 2: Update KV cache (in gain-cell crossbar)
	if a.Config.EnableKVCache {
		a.updateKVCache(K, V)
	}

	// Step 3: Compute attention scores
	var scores [][]float64
	var mask []bool

	if a.Config.EnableTokenPruning {
		// Analog CIM core prunes low-score tokens
		scores, mask = a.prunedAttentionScores(Q, K)
		stats.PrunedTokens = countPruned(mask)
	} else if a.Config.EnableSparseAttn {
		// LSH-based sparse attention
		scores = a.sparseAttentionScores(Q, K)
	} else {
		// Full attention
		scores = a.fullAttentionScores(Q, K)
	}

	// Step 4: Apply softmax
	attnWeights := a.softmax(scores)

	// Step 5: Weighted sum of values
	context := a.weightedSum(attnWeights, V, mask)

	// Step 6: Output projection
	output := a.projectQKV(context, a.WO)

	// Update power gating statistics
	if a.PowerGater != nil {
		for _, active := range a.ArrayPowerState {
			if active {
				stats.ActiveArrays++
			} else {
				stats.SleepingArrays++
			}
		}
		stats.PowerSaved = a.PowerGater.EstimatePowerSavings()
	}

	return output, stats
}

// projectQKV performs matrix projection using crossbar MVM
func (a *CIMAttention) projectQKV(input [][]float64, weights [][]float64) [][]float64 {
	if len(weights) == 0 {
		return input // Passthrough if no weights
	}

	seqLen := len(input)
	outDim := len(weights)
	output := make([][]float64, seqLen)

	for i := 0; i < seqLen; i++ {
		output[i] = make([]float64, outDim)
		for j := 0; j < outDim && j < len(weights); j++ {
			sum := 0.0
			for k := 0; k < len(input[i]) && k < len(weights[j]); k++ {
				sum += input[i][k] * weights[j][k]
			}
			// Add analog noise
			if a.Config.NoiseLevel > 0 {
				sum *= (1.0 + (randFloat()-0.5)*a.Config.NoiseLevel*2)
			}
			output[i][j] = sum
		}
	}

	return output
}

// fullAttentionScores computes full attention score matrix
func (a *CIMAttention) fullAttentionScores(Q, K [][]float64) [][]float64 {
	seqLen := len(Q)
	scores := make([][]float64, seqLen)
	scale := 1.0 / math.Sqrt(float64(a.Config.HeadDim))

	for i := 0; i < seqLen; i++ {
		scores[i] = make([]float64, seqLen)
		for j := 0; j < seqLen; j++ {
			dot := 0.0
			for k := 0; k < len(Q[i]) && k < len(K[j]); k++ {
				dot += Q[i][k] * K[j][k]
			}
			scores[i][j] = dot * scale
		}
	}

	return scores
}

// prunedAttentionScores uses analog CIM to prune low-score tokens
func (a *CIMAttention) prunedAttentionScores(Q, K [][]float64) ([][]float64, []bool) {
	seqLen := len(Q)
	scores := make([][]float64, seqLen)
	mask := make([]bool, seqLen) // true = pruned
	scale := 1.0 / math.Sqrt(float64(a.Config.HeadDim))

	// Compute approximate scores in analog domain
	approxScores := make([]float64, seqLen)
	for j := 0; j < seqLen; j++ {
		// Average attention received by each token
		totalScore := 0.0
		for i := 0; i < seqLen; i++ {
			dot := 0.0
			for k := 0; k < len(Q[i]) && k < len(K[j]); k++ {
				dot += Q[i][k] * K[j][k]
			}
			totalScore += dot * scale
		}
		approxScores[j] = totalScore / float64(seqLen)
	}

	// Sort and prune lowest scoring tokens
	type tokenScore struct {
		idx   int
		score float64
	}
	sorted := make([]tokenScore, seqLen)
	for i := range approxScores {
		sorted[i] = tokenScore{i, approxScores[i]}
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].score < sorted[j].score
	})

	// Mark tokens for pruning
	pruneCount := int(float64(seqLen) * a.Config.PruneRatio)
	for i := 0; i < pruneCount; i++ {
		mask[sorted[i].idx] = true
	}

	// Compute full scores only for unpruned tokens
	for i := 0; i < seqLen; i++ {
		scores[i] = make([]float64, seqLen)
		for j := 0; j < seqLen; j++ {
			if mask[j] {
				scores[i][j] = math.Inf(-1) // Will become 0 after softmax
			} else {
				dot := 0.0
				for k := 0; k < len(Q[i]) && k < len(K[j]); k++ {
					dot += Q[i][k] * K[j][k]
				}
				scores[i][j] = dot * scale
			}
		}
	}

	return scores, mask
}

// sparseAttentionScores uses LSH for O(L log L) attention
func (a *CIMAttention) sparseAttentionScores(Q, K [][]float64) [][]float64 {
	seqLen := len(Q)
	scores := make([][]float64, seqLen)
	scale := 1.0 / math.Sqrt(float64(a.Config.HeadDim))

	// Hash queries and keys into buckets
	qBuckets := a.lshHash(Q)
	kBuckets := a.lshHash(K)

	// Only compute attention within same bucket
	for i := 0; i < seqLen; i++ {
		scores[i] = make([]float64, seqLen)
		for j := 0; j < seqLen; j++ {
			if qBuckets[i] == kBuckets[j] {
				dot := 0.0
				for k := 0; k < len(Q[i]) && k < len(K[j]); k++ {
					dot += Q[i][k] * K[j][k]
				}
				scores[i][j] = dot * scale
			} else {
				scores[i][j] = math.Inf(-1)
			}
		}
	}

	return scores
}

// lshHash computes LSH bucket assignments
func (a *CIMAttention) lshHash(vectors [][]float64) []int {
	buckets := make([]int, len(vectors))

	// Simple angular LSH: hash = sign(random_projection)
	for i, vec := range vectors {
		hash := 0
		for b := 0; b < int(math.Log2(float64(a.Config.LSHBuckets))); b++ {
			// Random projection (in practice, stored in crossbar)
			proj := 0.0
			for k := range vec {
				proj += vec[k] * (float64((k*b)%17)/8.0 - 1.0) // Pseudo-random
			}
			if proj > 0 {
				hash |= (1 << b)
			}
		}
		buckets[i] = hash % a.Config.LSHBuckets
	}

	return buckets
}

// softmax applies softmax to attention scores
func (a *CIMAttention) softmax(scores [][]float64) [][]float64 {
	result := make([][]float64, len(scores))

	for i := range scores {
		result[i] = make([]float64, len(scores[i]))

		// Find max for numerical stability
		maxVal := math.Inf(-1)
		for _, s := range scores[i] {
			if s > maxVal && !math.IsInf(s, -1) {
				maxVal = s
			}
		}
		if math.IsInf(maxVal, -1) {
			maxVal = 0
		}

		// Compute exp and sum
		sum := 0.0
		for j, s := range scores[i] {
			if math.IsInf(s, -1) {
				result[i][j] = 0
			} else {
				result[i][j] = math.Exp(s - maxVal)
				sum += result[i][j]
			}
		}

		// Normalize
		if sum > 0 {
			for j := range result[i] {
				result[i][j] /= sum
			}
		}
	}

	return result
}

// weightedSum computes attention-weighted sum of values
func (a *CIMAttention) weightedSum(weights [][]float64, V [][]float64, mask []bool) [][]float64 {
	seqLen := len(weights)
	dim := 0
	if len(V) > 0 {
		dim = len(V[0])
	}

	result := make([][]float64, seqLen)
	for i := 0; i < seqLen; i++ {
		result[i] = make([]float64, dim)
		for j := 0; j < len(V); j++ {
			if mask != nil && mask[j] {
				continue // Skip pruned tokens
			}
			for k := 0; k < dim; k++ {
				result[i][k] += weights[i][j] * V[j][k]
			}
		}
	}

	return result
}

// updateKVCache updates the KV cache in gain-cell crossbar
func (a *CIMAttention) updateKVCache(K, V [][]float64) {
	for i := 0; i < len(K) && a.CacheLen < a.Config.SeqLength; i++ {
		copy(a.KCache[a.CacheLen], K[i])
		copy(a.VCache[a.CacheLen], V[i])
		a.CacheLen++
	}
}

// AttentionStats holds attention computation statistics
type AttentionStats struct {
	TotalTokens    int
	PrunedTokens   int
	ActiveArrays   int
	SleepingArrays int
	PowerSaved     float64 // Percentage
}

func countPruned(mask []bool) int {
	count := 0
	for _, m := range mask {
		if m {
			count++
		}
	}
	return count
}

// Simple pseudo-random for simulation
func randFloat() float64 {
	return 0.5 // Placeholder - would use rand.Float64()
}

// =============================================================================
// Power Gating Controller
// =============================================================================

// PowerGatingConfig configures power gating behavior
type PowerGatingConfig struct {
	NumArrays       int
	SleepThreshold  float64 // Activity below this triggers sleep
	WakeupLatency   int     // Cycles to wake up
	SleepPower      float64 // Power consumption when sleeping (fraction)

	// Advanced settings
	PredictiveWakeup bool    // Use prediction to pre-wake arrays
	HysteresisMargin float64 // Prevent frequent switching
}

// PowerGatingController manages array power states
type PowerGatingController struct {
	Config        *PowerGatingConfig
	ArrayStates   []ArrayPowerState
	ActivityHistory [][]float64 // Per-array activity history
	TotalEnergy   float64
	SavedEnergy   float64
}

// ArrayPowerState tracks individual array state
type ArrayPowerState struct {
	Active        bool
	CyclesSleeping int
	CyclesActive  int
	WakeupPending bool
	WakeupCounter int
}

// NewPowerGatingController creates a power gating controller
func NewPowerGatingController(config *PowerGatingConfig) *PowerGatingController {
	pg := &PowerGatingController{
		Config:          config,
		ArrayStates:     make([]ArrayPowerState, config.NumArrays),
		ActivityHistory: make([][]float64, config.NumArrays),
	}

	for i := range pg.ArrayStates {
		pg.ArrayStates[i].Active = true
		pg.ActivityHistory[i] = make([]float64, 0, 100)
	}

	return pg
}

// UpdateActivity records array activity and updates power states
func (pg *PowerGatingController) UpdateActivity(arrayIdx int, activity float64) {
	if arrayIdx >= len(pg.ArrayStates) {
		return
	}

	state := &pg.ArrayStates[arrayIdx]

	// Record activity
	pg.ActivityHistory[arrayIdx] = append(pg.ActivityHistory[arrayIdx], activity)
	if len(pg.ActivityHistory[arrayIdx]) > 100 {
		pg.ActivityHistory[arrayIdx] = pg.ActivityHistory[arrayIdx][1:]
	}

	// Compute average activity
	avgActivity := 0.0
	for _, a := range pg.ActivityHistory[arrayIdx] {
		avgActivity += a
	}
	avgActivity /= float64(len(pg.ActivityHistory[arrayIdx]))

	// State machine transitions
	if state.Active {
		state.CyclesActive++
		pg.TotalEnergy += 1.0

		// Check if should sleep
		if avgActivity < pg.Config.SleepThreshold {
			state.Active = false
			state.CyclesSleeping = 0
			pg.SavedEnergy += (1.0 - pg.Config.SleepPower)
		}
	} else {
		state.CyclesSleeping++
		pg.TotalEnergy += pg.Config.SleepPower

		// Check if should wake
		if activity > pg.Config.SleepThreshold+pg.Config.HysteresisMargin {
			state.WakeupPending = true
			state.WakeupCounter = pg.Config.WakeupLatency
		}

		// Handle wakeup
		if state.WakeupPending {
			state.WakeupCounter--
			if state.WakeupCounter <= 0 {
				state.Active = true
				state.WakeupPending = false
				state.CyclesActive = 0
			}
		}
	}
}

// PredictWakeup predicts which arrays will be needed
func (pg *PowerGatingController) PredictWakeup(inputPattern []float64) []int {
	if !pg.Config.PredictiveWakeup {
		return nil
	}

	// Simple prediction: wake arrays that were active for similar patterns
	toWake := make([]int, 0)
	for i, state := range pg.ArrayStates {
		if !state.Active {
			// Check if this array is likely needed
			if len(pg.ActivityHistory[i]) > 0 {
				lastActivity := pg.ActivityHistory[i][len(pg.ActivityHistory[i])-1]
				if lastActivity > pg.Config.SleepThreshold*0.5 {
					toWake = append(toWake, i)
				}
			}
		}
	}

	return toWake
}

// WakeArrays forces specific arrays to wake up
func (pg *PowerGatingController) WakeArrays(indices []int) {
	for _, idx := range indices {
		if idx < len(pg.ArrayStates) {
			pg.ArrayStates[idx].WakeupPending = true
			pg.ArrayStates[idx].WakeupCounter = pg.Config.WakeupLatency
		}
	}
}

// EstimatePowerSavings returns power savings percentage
func (pg *PowerGatingController) EstimatePowerSavings() float64 {
	if pg.TotalEnergy == 0 {
		return 0
	}
	return (pg.SavedEnergy / pg.TotalEnergy) * 100
}

// GetActiveArrayCount returns number of active arrays
func (pg *PowerGatingController) GetActiveArrayCount() int {
	count := 0
	for _, state := range pg.ArrayStates {
		if state.Active {
			count++
		}
	}
	return count
}

// =============================================================================
// 2D Ferroelectric Material Models
// =============================================================================

// Material2DType defines 2D ferroelectric material variants
type Material2DType string

const (
	MaterialIn2Se3    Material2DType = "In2Se3"
	MaterialSnS       Material2DType = "SnS"
	MaterialSnSe      Material2DType = "SnSe"
	MaterialBiSnSe    Material2DType = "Bi-SnSe"
	MaterialCuInP2S6  Material2DType = "CuInP2S6"
)

// Material2DConfig configures 2D ferroelectric properties
type Material2DConfig struct {
	Material          Material2DType
	Layers            int     // Number of atomic layers
	Polarization      float64 // Spontaneous polarization (µC/cm²)
	CoerciveField     float64 // Coercive field (kV/cm)
	CurieTemp         float64 // Curie temperature (K)
	BandgapEV         float64 // Band gap (eV)

	// Memristor properties
	OnOffRatio        float64
	MemoryWindow      float64 // Voltage window (V)
	SwitchingSpeed    float64 // Switching time (ns)
	Endurance         float64 // Cycles
	Retention         float64 // Seconds

	// Synapse properties
	NumConductanceStates int
	Linearity            float64 // Weight update linearity (0-1)
}

// Get2DMaterialConfig returns properties for specific material
func Get2DMaterialConfig(material Material2DType) *Material2DConfig {
	switch material {
	case MaterialIn2Se3:
		return &Material2DConfig{
			Material:             MaterialIn2Se3,
			Layers:               5,
			Polarization:         2.0,
			CoerciveField:        50.0,
			CurieTemp:            500.0, // Room temp stable
			BandgapEV:            1.3,
			OnOffRatio:           1e8,
			MemoryWindow:         16.0, // ±8V (2025 paper)
			SwitchingSpeed:       100.0,
			Endurance:            1e8,
			Retention:            1e5,
			NumConductanceStates: 64,
			Linearity:            0.85,
		}
	case MaterialSnS:
		return &Material2DConfig{
			Material:             MaterialSnS,
			Layers:               1,
			Polarization:         2.6, // 2.6 × 10^-10 C/m
			CoerciveField:        30.0,
			CurieTemp:            400.0, // Room temp
			BandgapEV:            1.3,
			OnOffRatio:           1e4,
			MemoryWindow:         4.0,
			SwitchingSpeed:       50.0,
			Endurance:            1e9,
			Retention:            1e6,
			NumConductanceStates: 32,
			Linearity:            0.75,
		}
	case MaterialSnSe:
		return &Material2DConfig{
			Material:             MaterialSnSe,
			Layers:               1,
			Polarization:         4.84, // Highest in group-IV
			CoerciveField:        40.0,
			CurieTemp:            390.0, // ~380-400 K
			BandgapEV:            0.9,
			OnOffRatio:           1e5,
			MemoryWindow:         5.0,
			SwitchingSpeed:       20.0,
			Endurance:            1e10,
			Retention:            1e5,
			NumConductanceStates: 64,
			Linearity:            0.80,
		}
	case MaterialBiSnSe:
		return &Material2DConfig{
			Material:             MaterialBiSnSe,
			Layers:               10,
			Polarization:         3.5,
			CoerciveField:        24.0, // Symmetric ±2.4V
			CurieTemp:            350.0,
			BandgapEV:            1.1,
			OnOffRatio:           1e6,
			MemoryWindow:         4.8,
			SwitchingSpeed:       30.0,
			Endurance:            1e8,
			Retention:            1e6,
			NumConductanceStates: 128,
			Linearity:            0.92, // Enhanced by Bi doping
		}
	case MaterialCuInP2S6:
		return &Material2DConfig{
			Material:             MaterialCuInP2S6,
			Layers:               5,
			Polarization:         4.0,
			CoerciveField:        100.0,
			CurieTemp:            315.0,
			BandgapEV:            2.9,
			OnOffRatio:           1e5,
			MemoryWindow:         10.0,
			SwitchingSpeed:       50.0,
			Endurance:            1e7,
			Retention:            5e6, // >2 months
			NumConductanceStates: 21,
			Linearity:            0.70,
		}
	default:
		return Get2DMaterialConfig(MaterialIn2Se3)
	}
}

// Material2DSynapse simulates a 2D ferroelectric synapse
type Material2DSynapse struct {
	Config           *Material2DConfig
	CurrentState     int     // Conductance state index
	Conductance      float64 // Current conductance
	PolarizationDir  int     // +1 or -1
	CycleCount       int
}

// NewMaterial2DSynapse creates a new 2D FE synapse
func NewMaterial2DSynapse(config *Material2DConfig) *Material2DSynapse {
	return &Material2DSynapse{
		Config:          config,
		CurrentState:    config.NumConductanceStates / 2,
		PolarizationDir: 1,
	}
}

// Potentiate increases synaptic weight
func (s *Material2DSynapse) Potentiate(pulseAmplitude, pulseWidth float64) {
	if pulseAmplitude < s.Config.CoerciveField*0.001 {
		return // Below threshold
	}

	// Nonlinear update with linearity factor
	delta := 1.0
	if s.CurrentState > s.Config.NumConductanceStates/2 {
		// Approaching saturation, reduce delta
		saturationFactor := float64(s.CurrentState) / float64(s.Config.NumConductanceStates)
		delta *= (1.0 - saturationFactor) * s.Config.Linearity
	}

	s.CurrentState += int(delta)
	if s.CurrentState >= s.Config.NumConductanceStates {
		s.CurrentState = s.Config.NumConductanceStates - 1
	}

	s.updateConductance()
	s.CycleCount++
}

// Depress decreases synaptic weight
func (s *Material2DSynapse) Depress(pulseAmplitude, pulseWidth float64) {
	if pulseAmplitude < s.Config.CoerciveField*0.001 {
		return
	}

	delta := 1.0
	if s.CurrentState < s.Config.NumConductanceStates/2 {
		saturationFactor := 1.0 - float64(s.CurrentState)/float64(s.Config.NumConductanceStates)
		delta *= (1.0 - saturationFactor) * s.Config.Linearity
	}

	s.CurrentState -= int(delta)
	if s.CurrentState < 0 {
		s.CurrentState = 0
	}

	s.updateConductance()
	s.CycleCount++
}

// updateConductance calculates conductance from state
func (s *Material2DSynapse) updateConductance() {
	// Log-linear conductance model
	gMin := 1e-9  // 1 nS
	gMax := 1e-6  // 1 µS
	stateNorm := float64(s.CurrentState) / float64(s.Config.NumConductanceStates-1)
	s.Conductance = gMin * math.Pow(s.Config.OnOffRatio, stateNorm)
	if s.Conductance > gMax {
		s.Conductance = gMax
	}
}

// GetWeight returns normalized weight [0, 1]
func (s *Material2DSynapse) GetWeight() float64 {
	return float64(s.CurrentState) / float64(s.Config.NumConductanceStates-1)
}

// EstimateEnergy returns energy per switching event (fJ)
func (s *Material2DSynapse) EstimateEnergy() float64 {
	// E ~ P_s × E_c × Area × thickness
	// Simplified for 2D: much lower than bulk
	area := 1e-12 // 1 µm²
	thickness := float64(s.Config.Layers) * 0.7e-9 // ~0.7 nm per layer
	return s.Config.Polarization * 1e-6 * s.Config.CoerciveField * 1e5 * area * thickness * 1e15
}

// =============================================================================
// CIM Attention Performance Estimation
// =============================================================================

// CIMAttentionPerformance holds performance metrics
type CIMAttentionPerformance struct {
	// Throughput
	TokensPerSecond     float64
	AttentionLatencyUs  float64

	// Power
	TotalPowerW         float64
	ArrayPowerW         float64
	PeripheralPowerW    float64
	PowerWithGating     float64

	// Energy
	EnergyPerTokenJ     float64
	EnergyVsGPU         float64 // Ratio vs GPU baseline

	// Efficiency
	TOPSW               float64
	TokensPerJoule      float64
}

// EstimateAttentionPerformance calculates CIM attention metrics
func EstimateAttentionPerformance(config *CIMAttentionConfig) *CIMAttentionPerformance {
	perf := &CIMAttentionPerformance{}

	// Number of arrays needed
	qkvSize := config.HiddenDim * config.HiddenDim
	arrayCapacity := config.CrossbarSize * config.CrossbarSize
	numArrays := (qkvSize * 4) / arrayCapacity // Q, K, V, O projections

	// MACs per token
	macsPerToken := int64(config.SeqLength) * int64(config.HiddenDim) * 4 // QKV + attention

	// Latency (assuming 1 GHz crossbar operation)
	cyclesPerMVM := config.CrossbarSize // One cycle per row
	mvmsPerToken := macsPerToken / int64(arrayCapacity)
	perf.AttentionLatencyUs = float64(mvmsPerToken*int64(cyclesPerMVM)) / 1000.0

	// Throughput
	perf.TokensPerSecond = 1e6 / perf.AttentionLatencyUs

	// Power model
	powerPerArray := 0.1 // mW per active array
	perf.ArrayPowerW = float64(numArrays) * powerPerArray / 1000.0
	perf.PeripheralPowerW = perf.ArrayPowerW * 0.3 // 30% overhead
	perf.TotalPowerW = perf.ArrayPowerW + perf.PeripheralPowerW

	// Power with gating
	if config.EnablePowerGating {
		activeRatio := 1.0 - config.PruneRatio*0.5 // Rough estimate
		perf.PowerWithGating = perf.TotalPowerW * activeRatio
	} else {
		perf.PowerWithGating = perf.TotalPowerW
	}

	// Energy
	perf.EnergyPerTokenJ = perf.PowerWithGating / perf.TokensPerSecond

	// Comparison to GPU (A100 baseline: ~0.5 µJ per token for attention)
	gpuEnergyPerToken := 0.5e-6
	perf.EnergyVsGPU = gpuEnergyPerToken / perf.EnergyPerTokenJ

	// Efficiency
	tops := float64(macsPerToken) * 2 * perf.TokensPerSecond / 1e12
	perf.TOPSW = tops / perf.PowerWithGating
	perf.TokensPerJoule = 1.0 / perf.EnergyPerTokenJ

	return perf
}

// String returns formatted performance summary
func (p *CIMAttentionPerformance) String() string {
	return fmt.Sprintf(`CIM Attention Performance:
  Throughput:    %.0f tokens/sec
  Latency:       %.2f µs/token
  Power:         %.3f W (%.3f W with gating)
  Energy/Token:  %.2f nJ
  vs GPU:        %.0f× more efficient
  TOPS/W:        %.1f
  Tokens/Joule:  %.2e`,
		p.TokensPerSecond,
		p.AttentionLatencyUs,
		p.TotalPowerW, p.PowerWithGating,
		p.EnergyPerTokenJ*1e9,
		p.EnergyVsGPU,
		p.TOPSW,
		p.TokensPerJoule)
}
