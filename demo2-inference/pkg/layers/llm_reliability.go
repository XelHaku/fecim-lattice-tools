// llm_reliability.go - LLM Acceleration and Reliability Modeling for CIM
//
// This module implements:
// - CIM-based LLM inference acceleration
// - KV cache management for transformers
// - Attention mechanism optimization
// - Ferroelectric device reliability modeling
// - Endurance and retention simulation
// - Aging and degradation prediction
//
// Based on research findings:
// - KV cache uses 70-80% of inference energy
// - CIM achieves 23-1894× efficiency gain
// - Si-FeFET endurance: <10⁶ cycles
// - HZO superlattice: 10¹⁰ cycles
// - Charge trapping main fatigue mechanism
//
// References:
// - "Memory Is All You Need" (arXiv 2406.08413)
// - FeFET reliability reviews (Journal of Semiconductors 2024)
// - IBM AIMC for attention (Nature 2025)

package layers

import (
	"math"
	"math/rand"
)

// ================== LLM CIM Acceleration ==================

// LLMCIMConfig configures CIM for LLM inference
type LLMCIMConfig struct {
	// Model parameters
	ModelName        string
	NumParameters    int64   // Total parameters
	HiddenDim        int     // Hidden dimension
	NumLayers        int     // Transformer layers
	NumHeads         int     // Attention heads
	HeadDim          int     // Dimension per head
	VocabSize        int     // Vocabulary size
	MaxSeqLength     int     // Maximum sequence length

	// CIM configuration
	ArraySize        int     // Crossbar array size
	NumArrays        int     // Number of arrays
	WeightPrecision  int     // Weight bits
	ActivationPrecision int  // Activation bits
	KVCachePrecision int     // KV cache bits

	// Memory hierarchy
	OnChipMemoryMB   float64 // On-chip memory (MB)
	OffChipBandwidth float64 // Off-chip bandwidth (GB/s)

	// Efficiency targets
	TargetTOPSW      float64 // Target TOPS/W
}

// DefaultLLMCIMConfig returns configuration for a 7B model
func DefaultLLMCIMConfig() *LLMCIMConfig {
	return &LLMCIMConfig{
		ModelName:           "LLaMA-7B",
		NumParameters:       7000000000, // 7B
		HiddenDim:           4096,
		NumLayers:           32,
		NumHeads:            32,
		HeadDim:             128,
		VocabSize:           32000,
		MaxSeqLength:        4096,

		ArraySize:           256,
		NumArrays:           1024,
		WeightPrecision:     4,  // INT4
		ActivationPrecision: 8,  // INT8
		KVCachePrecision:    8,

		OnChipMemoryMB:      64,
		OffChipBandwidth:    100, // 100 GB/s

		TargetTOPSW:         10,
	}
}

// GPT2Config returns configuration for GPT-2
func GPT2Config() *LLMCIMConfig {
	return &LLMCIMConfig{
		ModelName:           "GPT-2",
		NumParameters:       117000000, // 117M
		HiddenDim:           768,
		NumLayers:           12,
		NumHeads:            12,
		HeadDim:             64,
		VocabSize:           50257,
		MaxSeqLength:        1024,

		ArraySize:           128,
		NumArrays:           256,
		WeightPrecision:     8,
		ActivationPrecision: 8,
		KVCachePrecision:    8,

		OnChipMemoryMB:      16,
		OffChipBandwidth:    50,

		TargetTOPSW:         15,
	}
}

// KVCache represents the key-value cache for transformers
type KVCache struct {
	Config       *LLMCIMConfig
	KeyCache     [][][]float64 // [layer][seq][head_dim]
	ValueCache   [][][]float64
	CacheLength  int
	MaxLength    int

	// CIM-specific
	KeyArrays    []*CIMArray
	ValueArrays  []*CIMArray
	WriteCount   int64
	ReadCount    int64
	WriteEnergy  float64 // pJ
	ReadEnergy   float64 // pJ
}

// CIMArray represents a CIM array for KV storage
type CIMArray struct {
	Rows         int
	Cols         int
	Conductances [][]float64
	WriteCount   int64
	LastWriteTime float64
}

// NewKVCache creates a KV cache with CIM arrays
func NewKVCache(config *LLMCIMConfig) *KVCache {
	numLayers := config.NumLayers
	maxLen := config.MaxSeqLength
	headDim := config.HeadDim * config.NumHeads

	keyCache := make([][][]float64, numLayers)
	valueCache := make([][][]float64, numLayers)
	keyArrays := make([]*CIMArray, numLayers)
	valueArrays := make([]*CIMArray, numLayers)

	for l := 0; l < numLayers; l++ {
		keyCache[l] = make([][]float64, maxLen)
		valueCache[l] = make([][]float64, maxLen)
		for s := 0; s < maxLen; s++ {
			keyCache[l][s] = make([]float64, headDim)
			valueCache[l][s] = make([]float64, headDim)
		}

		// Create CIM arrays for K and V
		keyArrays[l] = &CIMArray{
			Rows:         maxLen,
			Cols:         headDim,
			Conductances: make([][]float64, maxLen),
		}
		valueArrays[l] = &CIMArray{
			Rows:         maxLen,
			Cols:         headDim,
			Conductances: make([][]float64, maxLen),
		}
		for s := 0; s < maxLen; s++ {
			keyArrays[l].Conductances[s] = make([]float64, headDim)
			valueArrays[l].Conductances[s] = make([]float64, headDim)
		}
	}

	return &KVCache{
		Config:      config,
		KeyCache:    keyCache,
		ValueCache:  valueCache,
		CacheLength: 0,
		MaxLength:   maxLen,
		KeyArrays:   keyArrays,
		ValueArrays: valueArrays,
	}
}

// Append appends new K,V vectors to the cache
func (kv *KVCache) Append(layer int, key, value []float64) {
	if kv.CacheLength >= kv.MaxLength {
		// Evict oldest entry
		kv.CacheLength = kv.MaxLength - 1
	}

	pos := kv.CacheLength
	copy(kv.KeyCache[layer][pos], key)
	copy(kv.ValueCache[layer][pos], value)

	// Update CIM arrays (write operation)
	for i, v := range key {
		kv.KeyArrays[layer].Conductances[pos][i] = v
	}
	for i, v := range value {
		kv.ValueArrays[layer].Conductances[pos][i] = v
	}

	kv.KeyArrays[layer].WriteCount++
	kv.ValueArrays[layer].WriteCount++
	kv.WriteCount += 2

	// Estimate write energy (based on array size and precision)
	bitsWritten := float64(len(key)+len(value)) * float64(kv.Config.KVCachePrecision)
	kv.WriteEnergy += bitsWritten * 10 // ~10 pJ/bit for volatile memory

	if layer == kv.Config.NumLayers-1 {
		kv.CacheLength++
	}
}

// ComputeAttention computes attention using CIM
func (kv *KVCache) ComputeAttention(layer int, query []float64) ([]float64, *AttentionMetrics) {
	cfg := kv.Config
	metrics := &AttentionMetrics{}

	// Q × K^T (dot products)
	scores := make([]float64, kv.CacheLength)
	for i := 0; i < kv.CacheLength; i++ {
		dot := 0.0
		for j := 0; j < len(query); j++ {
			dot += query[j] * kv.KeyCache[layer][i][j]
		}
		scores[i] = dot / math.Sqrt(float64(cfg.HeadDim))
	}

	// Softmax
	maxScore := scores[0]
	for _, s := range scores {
		if s > maxScore {
			maxScore = s
		}
	}
	expSum := 0.0
	for i := range scores {
		scores[i] = math.Exp(scores[i] - maxScore)
		expSum += scores[i]
	}
	for i := range scores {
		scores[i] /= expSum
	}

	// Attention × V
	output := make([]float64, len(query))
	for i := 0; i < kv.CacheLength; i++ {
		for j := 0; j < len(output); j++ {
			output[j] += scores[i] * kv.ValueCache[layer][i][j]
		}
	}

	// Update metrics
	kv.ReadCount += int64(kv.CacheLength * 2) // Read K and V
	metrics.QKOps = int64(kv.CacheLength * len(query))
	metrics.SoftmaxOps = int64(kv.CacheLength * 3) // exp, sum, div
	metrics.AVOps = int64(kv.CacheLength * len(output))
	metrics.TotalOps = metrics.QKOps + metrics.SoftmaxOps + metrics.AVOps

	// Read energy (much lower than write)
	bitsRead := float64(kv.CacheLength*len(query)*2) * float64(cfg.KVCachePrecision)
	readEnergy := bitsRead * 0.1 // ~0.1 pJ/bit for read
	kv.ReadEnergy += readEnergy
	metrics.EnergyPJ = readEnergy

	return output, metrics
}

// AttentionMetrics contains attention computation metrics
type AttentionMetrics struct {
	QKOps       int64
	SoftmaxOps  int64
	AVOps       int64
	TotalOps    int64
	EnergyPJ    float64
	LatencyNs   float64
}

// GetKVCacheStats returns KV cache statistics
func (kv *KVCache) GetKVCacheStats() map[string]interface{} {
	cfg := kv.Config

	// Memory usage
	kvMemoryBytes := int64(kv.CacheLength) * int64(cfg.NumLayers) * int64(cfg.HeadDim*cfg.NumHeads) * 2 * int64(cfg.KVCachePrecision) / 8

	return map[string]interface{}{
		"CacheLength":     kv.CacheLength,
		"MaxLength":       kv.MaxLength,
		"MemoryUsageMB":   float64(kvMemoryBytes) / 1e6,
		"WriteCount":      kv.WriteCount,
		"ReadCount":       kv.ReadCount,
		"WriteEnergyPJ":   kv.WriteEnergy,
		"ReadEnergyPJ":    kv.ReadEnergy,
		"WriteReadRatio":  kv.WriteEnergy / math.Max(1, kv.ReadEnergy),
	}
}

// ================== LLM Inference Accelerator ==================

// LLMAccelerator represents a CIM-based LLM accelerator
type LLMAccelerator struct {
	Config        *LLMCIMConfig
	KVCaches      []*KVCache
	WeightArrays  [][]*CIMArray // [layer][sublayer]

	// Performance tracking
	TotalTokens   int64
	TotalOps      int64
	TotalEnergy   float64 // pJ
	TotalLatency  float64 // ns
}

// NewLLMAccelerator creates an LLM accelerator
func NewLLMAccelerator(config *LLMCIMConfig) *LLMAccelerator {
	// Create KV cache (shared across batch)
	kvCache := NewKVCache(config)

	// Weight arrays for each layer
	weightArrays := make([][]*CIMArray, config.NumLayers)
	for l := 0; l < config.NumLayers; l++ {
		// Each layer has: QKV projection, output projection, FFN
		weightArrays[l] = make([]*CIMArray, 4)

		// QKV: hidden_dim × 3*hidden_dim
		weightArrays[l][0] = &CIMArray{
			Rows: config.HiddenDim,
			Cols: 3 * config.HiddenDim,
		}
		// Output: hidden_dim × hidden_dim
		weightArrays[l][1] = &CIMArray{
			Rows: config.HiddenDim,
			Cols: config.HiddenDim,
		}
		// FFN up: hidden_dim × 4*hidden_dim
		weightArrays[l][2] = &CIMArray{
			Rows: config.HiddenDim,
			Cols: 4 * config.HiddenDim,
		}
		// FFN down: 4*hidden_dim × hidden_dim
		weightArrays[l][3] = &CIMArray{
			Rows: 4 * config.HiddenDim,
			Cols: config.HiddenDim,
		}
	}

	return &LLMAccelerator{
		Config:       config,
		KVCaches:     []*KVCache{kvCache},
		WeightArrays: weightArrays,
	}
}

// GenerateToken generates a single token
func (acc *LLMAccelerator) GenerateToken(inputToken int) (int, *TokenMetrics) {
	cfg := acc.Config
	metrics := &TokenMetrics{}

	// Embedding lookup (simple for now)
	hidden := make([]float64, cfg.HiddenDim)
	for i := range hidden {
		hidden[i] = rand.NormFloat64() * 0.02
	}

	// Process through layers
	for l := 0; l < cfg.NumLayers; l++ {
		// Self-attention
		// QKV projection
		qkvOps := int64(cfg.HiddenDim * 3 * cfg.HiddenDim)
		metrics.QKVProjOps += qkvOps

		// Extract Q, K, V (simplified)
		query := hidden[:cfg.HeadDim*cfg.NumHeads]
		key := hidden[:cfg.HeadDim*cfg.NumHeads]
		value := hidden[:cfg.HeadDim*cfg.NumHeads]

		// Update KV cache
		acc.KVCaches[0].Append(l, key, value)

		// Compute attention
		attnOutput, attnMetrics := acc.KVCaches[0].ComputeAttention(l, query)
		metrics.AttentionOps += attnMetrics.TotalOps
		metrics.AttentionEnergy += attnMetrics.EnergyPJ

		// Output projection
		outputOps := int64(cfg.HiddenDim * cfg.HiddenDim)
		metrics.OutputProjOps += outputOps

		// Residual connection
		for i := range attnOutput {
			if i < len(hidden) {
				hidden[i] += attnOutput[i]
			}
		}

		// FFN
		ffnOps := int64(cfg.HiddenDim*4*cfg.HiddenDim) * 2 // Up + down
		metrics.FFNOps += ffnOps
	}

	// Final output
	metrics.TotalOps = metrics.QKVProjOps + metrics.AttentionOps +
		metrics.OutputProjOps + metrics.FFNOps

	// Energy estimation
	// Static weights: ~1 pJ/op for CIM
	// Dynamic (attention): ~10 pJ/op due to writes
	staticOps := metrics.QKVProjOps + metrics.OutputProjOps + metrics.FFNOps
	dynamicOps := metrics.AttentionOps
	metrics.TotalEnergy = float64(staticOps)*1 + float64(dynamicOps)*10 + metrics.AttentionEnergy

	// Latency (assume 100 MHz, pipelined)
	metrics.LatencyNs = float64(cfg.NumLayers) * 1000 // ~1µs per layer

	// Update accelerator stats
	acc.TotalTokens++
	acc.TotalOps += metrics.TotalOps
	acc.TotalEnergy += metrics.TotalEnergy
	acc.TotalLatency += metrics.LatencyNs

	// Output token (simplified)
	outputToken := rand.Intn(cfg.VocabSize)
	return outputToken, metrics
}

// TokenMetrics contains per-token metrics
type TokenMetrics struct {
	QKVProjOps     int64
	AttentionOps   int64
	AttentionEnergy float64
	OutputProjOps  int64
	FFNOps         int64
	TotalOps       int64
	TotalEnergy    float64 // pJ
	LatencyNs      float64
}

// GetEfficiency returns efficiency metrics
func (acc *LLMAccelerator) GetEfficiency() map[string]float64 {
	timeSeconds := acc.TotalLatency / 1e9
	energyJoules := acc.TotalEnergy / 1e12

	topsw := 0.0
	if energyJoules > 0 {
		topsw = float64(acc.TotalOps) / energyJoules / 1e12
	}

	tokensPerSecond := 0.0
	if timeSeconds > 0 {
		tokensPerSecond = float64(acc.TotalTokens) / timeSeconds
	}

	return map[string]float64{
		"TotalTokens":      float64(acc.TotalTokens),
		"TotalOps":         float64(acc.TotalOps),
		"TotalEnergyPJ":    acc.TotalEnergy,
		"TotalLatencyNs":   acc.TotalLatency,
		"TOPSW":            topsw,
		"TokensPerSecond":  tokensPerSecond,
		"EnergyPerToken":   acc.TotalEnergy / math.Max(1, float64(acc.TotalTokens)),
	}
}

// ================== Reliability Modeling ==================

// ReliabilityConfig configures reliability simulation
type ReliabilityConfig struct {
	// Device type
	DeviceType       string // "FeFET", "ReRAM", "PCM"
	MaterialType     string // "HZO", "AlScN", "superlattice"

	// Endurance parameters
	MaxEndurance     float64 // Maximum cycles before failure
	EnduranceSlope   float64 // Weibull slope

	// Retention parameters
	RetentionActivationEnergy float64 // eV
	RetentionTime25C float64 // Retention time at 25°C (s)

	// Fatigue parameters
	ChargeTrapDensity float64 // Trap density (cm⁻²)
	TrapGenRate       float64 // Trap generation rate per cycle
	DepolarizationRate float64 // Polarization loss rate

	// Variation
	CycleToVariation  float64 // Variance increase per cycle
	InitialVariation  float64 // Initial device variation
}

// DefaultReliabilityConfig returns typical FeFET reliability parameters
func DefaultReliabilityConfig() *ReliabilityConfig {
	return &ReliabilityConfig{
		DeviceType:       "FeFET",
		MaterialType:     "HZO",
		MaxEndurance:     1e10,   // 10^10 for HZO superlattice
		EnduranceSlope:   2.5,    // Weibull slope

		RetentionActivationEnergy: 1.2, // 1.2 eV
		RetentionTime25C: 10 * 365 * 24 * 3600, // 10 years

		ChargeTrapDensity: 1e12, // 10^12 cm⁻²
		TrapGenRate:       1e-10, // Per cycle
		DepolarizationRate: 1e-12, // Per cycle

		CycleToVariation: 1e-8, // Variance increase per cycle
		InitialVariation: 0.05, // 5% initial variation
	}
}

// SiFeFETConfig returns Si-based FeFET (lower endurance)
func SiFeFETConfig() *ReliabilityConfig {
	cfg := DefaultReliabilityConfig()
	cfg.MaterialType = "HZO_Si"
	cfg.MaxEndurance = 1e6 // Only 10^6 for Si-based
	cfg.TrapGenRate = 1e-8 // Higher trap generation
	cfg.ChargeTrapDensity = 1e13
	return cfg
}

// DeviceState represents the reliability state of a device
type DeviceState struct {
	Config          *ReliabilityConfig
	CycleCount      int64
	TrappedCharge   float64 // Accumulated trapped charge
	PolarizationLoss float64 // Polarization degradation (0-1)
	MemoryWindow    float64 // Current memory window (V)
	InitialMW       float64 // Initial memory window
	Variation       float64 // Current variation level
	Failed          bool
	FailureMode     string
}

// NewDeviceState creates initial device state
func NewDeviceState(config *ReliabilityConfig, initialMW float64) *DeviceState {
	return &DeviceState{
		Config:          config,
		CycleCount:      0,
		TrappedCharge:   0,
		PolarizationLoss: 0,
		MemoryWindow:    initialMW,
		InitialMW:       initialMW,
		Variation:       config.InitialVariation,
		Failed:          false,
		FailureMode:     "",
	}
}

// ApplyCycle simulates one write cycle
func (ds *DeviceState) ApplyCycle() {
	if ds.Failed {
		return
	}

	cfg := ds.Config
	ds.CycleCount++

	// Charge trapping
	ds.TrappedCharge += cfg.TrapGenRate * cfg.ChargeTrapDensity

	// Polarization degradation (fatigue)
	ds.PolarizationLoss += cfg.DepolarizationRate

	// Memory window narrowing
	chargeEffect := ds.TrappedCharge / (cfg.ChargeTrapDensity * 100) // Normalize
	polarizationEffect := ds.PolarizationLoss
	ds.MemoryWindow = ds.InitialMW * (1 - chargeEffect - polarizationEffect)

	// Variation increase
	ds.Variation = cfg.InitialVariation + float64(ds.CycleCount)*cfg.CycleToVariation

	// Check for failure (Weibull distribution)
	failureProbability := 1 - math.Exp(-math.Pow(float64(ds.CycleCount)/cfg.MaxEndurance, cfg.EnduranceSlope))
	if rand.Float64() < failureProbability*0.001 { // Scale down for simulation
		ds.Failed = true
		ds.FailureMode = "endurance"
	}

	// Memory window closure failure
	if ds.MemoryWindow < ds.InitialMW*0.1 {
		ds.Failed = true
		ds.FailureMode = "memory_window_closure"
	}
}

// PredictRetention predicts retention time at a temperature
func (ds *DeviceState) PredictRetention(tempCelsius float64) float64 {
	cfg := ds.Config

	// Arrhenius model
	// t = t_ref × exp(Ea/k × (1/T - 1/T_ref))
	kB := 8.617e-5 // Boltzmann constant (eV/K)
	tRef := 298.0  // 25°C in K
	t := tempCelsius + 273.15

	retentionFactor := math.Exp(cfg.RetentionActivationEnergy / kB * (1/t - 1/tRef))
	baseRetention := cfg.RetentionTime25C

	// Degradation from cycling
	cycleDegradation := 1 - ds.PolarizationLoss - ds.TrappedCharge/(cfg.ChargeTrapDensity*100)
	cycleDegradation = math.Max(0.1, cycleDegradation)

	return baseRetention * retentionFactor * cycleDegradation
}

// GetReliabilityMetrics returns current reliability metrics
func (ds *DeviceState) GetReliabilityMetrics() map[string]interface{} {
	cfg := ds.Config

	remainingLife := 1.0 - float64(ds.CycleCount)/cfg.MaxEndurance
	mwRetention := ds.MemoryWindow / ds.InitialMW

	return map[string]interface{}{
		"CycleCount":       ds.CycleCount,
		"MaxEndurance":     cfg.MaxEndurance,
		"RemainingLife":    remainingLife,
		"MemoryWindow":     ds.MemoryWindow,
		"MWRetention":      mwRetention,
		"TrappedCharge":    ds.TrappedCharge,
		"PolarizationLoss": ds.PolarizationLoss,
		"Variation":        ds.Variation,
		"Failed":           ds.Failed,
		"FailureMode":      ds.FailureMode,
		"Retention25C":     ds.PredictRetention(25),
		"Retention85C":     ds.PredictRetention(85),
	}
}

// ================== Array-Level Reliability ==================

// ArrayReliability tracks reliability of an entire array
type ArrayReliability struct {
	Config        *ReliabilityConfig
	Devices       [][]*DeviceState
	Rows          int
	Cols          int
	FailedCells   int
	TotalCycles   int64

	// Statistics
	AverageMW     float64
	MaxVariation  float64
	MTTF          float64 // Mean time to failure (cycles)
}

// NewArrayReliability creates array-level reliability tracking
func NewArrayReliability(config *ReliabilityConfig, rows, cols int, initialMW float64) *ArrayReliability {
	devices := make([][]*DeviceState, rows)
	for r := 0; r < rows; r++ {
		devices[r] = make([]*DeviceState, cols)
		for c := 0; c < cols; c++ {
			devices[r][c] = NewDeviceState(config, initialMW)
		}
	}

	return &ArrayReliability{
		Config:  config,
		Devices: devices,
		Rows:    rows,
		Cols:    cols,
	}
}

// ApplyWritePattern applies a write pattern to the array
func (ar *ArrayReliability) ApplyWritePattern(pattern [][]bool) {
	for r := 0; r < ar.Rows && r < len(pattern); r++ {
		for c := 0; c < ar.Cols && c < len(pattern[r]); c++ {
			if pattern[r][c] {
				ar.Devices[r][c].ApplyCycle()
				ar.TotalCycles++
			}
		}
	}
	ar.updateStatistics()
}

// ApplyUniformCycles applies cycles uniformly
func (ar *ArrayReliability) ApplyUniformCycles(numCycles int) {
	for n := 0; n < numCycles; n++ {
		for r := 0; r < ar.Rows; r++ {
			for c := 0; c < ar.Cols; c++ {
				ar.Devices[r][c].ApplyCycle()
				ar.TotalCycles++
			}
		}
	}
	ar.updateStatistics()
}

// updateStatistics updates array-level statistics
func (ar *ArrayReliability) updateStatistics() {
	totalMW := 0.0
	maxVar := 0.0
	ar.FailedCells = 0

	for r := 0; r < ar.Rows; r++ {
		for c := 0; c < ar.Cols; c++ {
			dev := ar.Devices[r][c]
			totalMW += dev.MemoryWindow
			if dev.Variation > maxVar {
				maxVar = dev.Variation
			}
			if dev.Failed {
				ar.FailedCells++
			}
		}
	}

	totalCells := ar.Rows * ar.Cols
	ar.AverageMW = totalMW / float64(totalCells)
	ar.MaxVariation = maxVar

	// Estimate MTTF (when 50% cells fail)
	if ar.FailedCells > 0 {
		ar.MTTF = float64(ar.TotalCycles) / float64(ar.FailedCells) * float64(totalCells/2)
	} else {
		ar.MTTF = ar.Config.MaxEndurance
	}
}

// GetArrayHealth returns array health metrics
func (ar *ArrayReliability) GetArrayHealth() map[string]interface{} {
	totalCells := ar.Rows * ar.Cols
	healthyCells := totalCells - ar.FailedCells
	healthRatio := float64(healthyCells) / float64(totalCells)

	return map[string]interface{}{
		"TotalCells":    totalCells,
		"HealthyCells":  healthyCells,
		"FailedCells":   ar.FailedCells,
		"HealthRatio":   healthRatio,
		"TotalCycles":   ar.TotalCycles,
		"AverageMW":     ar.AverageMW,
		"MaxVariation":  ar.MaxVariation,
		"MTTF":          ar.MTTF,
		"CanOperate":    healthRatio > 0.99, // 99% healthy threshold
	}
}

// PredictLifetime predicts remaining array lifetime
func (ar *ArrayReliability) PredictLifetime(cyclesPerSecond float64) float64 {
	// Find minimum remaining life across all devices
	minRemainingCycles := ar.Config.MaxEndurance

	for r := 0; r < ar.Rows; r++ {
		for c := 0; c < ar.Cols; c++ {
			dev := ar.Devices[r][c]
			if !dev.Failed {
				remaining := ar.Config.MaxEndurance - float64(dev.CycleCount)
				if remaining < minRemainingCycles {
					minRemainingCycles = remaining
				}
			}
		}
	}

	// Convert to time
	if cyclesPerSecond > 0 {
		return minRemainingCycles / cyclesPerSecond // seconds
	}
	return minRemainingCycles
}
