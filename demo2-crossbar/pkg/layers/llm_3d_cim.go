// llm_3d_cim.go - LLM Inference Optimization and 3D CIM Integration
// Iteration 151: KV cache management, attention acceleration, 3D FeNAND, chiplet integration
//
// Key research:
// - Analog attention: 70,000× energy reduction, 100× speedup vs GPU
// - MoE on 3D AIMC: Expert → physical layer mapping
// - 3D FeNAND: 4000× density, 1000× TOPS/mm² improvement
// - 3D-CIMlet: 12× energy efficiency for edge LLM

package layers

import (
	"fmt"
	"math"
	"math/rand"
)

// =============================================================================
// PART 1: LLM INFERENCE OPTIMIZATION ON CIM
// =============================================================================

// KVCacheConfig configures KV cache management
type KVCacheConfig struct {
	MaxSequenceLength int
	NumHeads          int
	HeadDim           int
	NumLayers         int
	CacheType         KVCacheType
	QuantBits         int     // KV quantization
	EvictionPolicy    string  // "lru", "attention_score", "sliding_window"
	OffloadToHost     bool    // CPU memory offload
	CompressionRatio  float64 // KV compression
}

// KVCacheType represents storage type for KV cache
type KVCacheType string

const (
	KVCacheGPU      KVCacheType = "gpu_hbm"
	KVCacheCPU      KVCacheType = "cpu_dram"
	KVCacheCXL      KVCacheType = "cxl_memory"
	KVCacheAnalog   KVCacheType = "analog_gain_cell"
	KVCacheHybrid   KVCacheType = "hybrid_analog_digital"
)

// KVCache manages key-value cache for LLM inference
type KVCache struct {
	Config        *KVCacheConfig
	Keys          [][][][]float64 // [layer][batch][seq][head_dim]
	Values        [][][][]float64
	CurrentLength int
	HitRate       float64
	MemoryUsageMB float64
	EnergyPerAccess float64 // fJ
}

// NewKVCache creates a KV cache
func NewKVCache(config *KVCacheConfig) *KVCache {
	cache := &KVCache{
		Config: config,
		Keys:   make([][][][]float64, config.NumLayers),
		Values: make([][][][]float64, config.NumLayers),
	}

	// Initialize cache storage
	for l := 0; l < config.NumLayers; l++ {
		cache.Keys[l] = make([][][]float64, 1) // batch=1 for simplicity
		cache.Values[l] = make([][][]float64, 1)
		cache.Keys[l][0] = make([][]float64, config.MaxSequenceLength)
		cache.Values[l][0] = make([][]float64, config.MaxSequenceLength)
		for s := 0; s < config.MaxSequenceLength; s++ {
			cache.Keys[l][0][s] = make([]float64, config.NumHeads*config.HeadDim)
			cache.Values[l][0][s] = make([]float64, config.NumHeads*config.HeadDim)
		}
	}

	// Calculate memory usage
	totalElements := config.NumLayers * config.MaxSequenceLength *
		config.NumHeads * config.HeadDim * 2 // K and V
	bytesPerElement := float64(config.QuantBits) / 8.0
	cache.MemoryUsageMB = float64(totalElements) * bytesPerElement / (1024 * 1024)

	// Set energy based on cache type
	switch config.CacheType {
	case KVCacheGPU:
		cache.EnergyPerAccess = 20.0 // fJ
	case KVCacheCPU:
		cache.EnergyPerAccess = 50.0
	case KVCacheCXL:
		cache.EnergyPerAccess = 30.0
	case KVCacheAnalog:
		cache.EnergyPerAccess = 0.001 // Much lower for analog
	case KVCacheHybrid:
		cache.EnergyPerAccess = 5.0
	}

	return cache
}

// AppendKV appends new key-value pair to cache
func (kv *KVCache) AppendKV(layer int, key, value []float64) {
	if kv.CurrentLength >= kv.Config.MaxSequenceLength {
		// Eviction needed
		kv.evict()
	}

	copy(kv.Keys[layer][0][kv.CurrentLength], key)
	copy(kv.Values[layer][0][kv.CurrentLength], value)

	if layer == kv.Config.NumLayers-1 {
		kv.CurrentLength++
	}
}

// evict removes oldest entries based on policy
func (kv *KVCache) evict() {
	switch kv.Config.EvictionPolicy {
	case "sliding_window":
		// Remove first half
		halfLen := kv.Config.MaxSequenceLength / 2
		for l := 0; l < kv.Config.NumLayers; l++ {
			copy(kv.Keys[l][0], kv.Keys[l][0][halfLen:])
			copy(kv.Values[l][0], kv.Values[l][0][halfLen:])
		}
		kv.CurrentLength = halfLen
	default:
		// LRU - just remove oldest
		kv.CurrentLength = kv.Config.MaxSequenceLength - 1
	}
}

// GetKV retrieves keys and values for attention
func (kv *KVCache) GetKV(layer int) ([][]float64, [][]float64) {
	return kv.Keys[layer][0][:kv.CurrentLength], kv.Values[layer][0][:kv.CurrentLength]
}

// =============================================================================
// ANALOG ATTENTION MECHANISM
// =============================================================================

// AnalogAttentionConfig configures analog attention accelerator
type AnalogAttentionConfig struct {
	NumHeads        int
	HeadDim         int
	MaxSeqLength    int
	GainCellEnabled bool     // Use gain cell for KV storage
	TokenPruning    bool     // Prune low-attention tokens
	PruningRatio    float64  // Fraction to prune (e.g., 0.75)
	KernelApprox    bool     // Use kernel approximation for softmax
	QuantBits       int
}

// AnalogAttentionAccelerator implements attention on analog CIM
// Based on Nature Comp Sci 2025: 70,000× energy, 100× speedup
type AnalogAttentionAccelerator struct {
	Config            *AnalogAttentionConfig
	GainCellArray     *GainCellMemory
	QKCrossbar        *AnalogCrossbar
	KVCrossbar        *AnalogCrossbar
	TokenScores       []float64
	PrunedIndices     []int
	EnergyReduction   float64 // vs GPU
	SpeedupFactor     float64
}

// GainCellMemory represents volatile analog memory for KV cache
type GainCellMemory struct {
	Rows            int
	Cols            int
	Storage         [][]float64
	RetentionTimeMs float64
	WriteEnergyFJ   float64
	ReadEnergyFJ    float64
	RefreshPeriodMs float64
}

// AnalogCrossbar represents analog compute array
type AnalogCrossbar struct {
	Rows       int
	Cols       int
	Weights    [][]float64
	NoiseLevel float64
}

// NewAnalogAttentionAccelerator creates attention accelerator
func NewAnalogAttentionAccelerator(config *AnalogAttentionConfig) *AnalogAttentionAccelerator {
	acc := &AnalogAttentionAccelerator{
		Config:          config,
		EnergyReduction: 70000.0, // From literature
		SpeedupFactor:   100.0,   // From literature
	}

	if config.GainCellEnabled {
		acc.GainCellArray = &GainCellMemory{
			Rows:            config.MaxSeqLength,
			Cols:            config.NumHeads * config.HeadDim,
			Storage:         make([][]float64, config.MaxSeqLength),
			RetentionTimeMs: 100.0,
			WriteEnergyFJ:   0.1,
			ReadEnergyFJ:    0.01,
			RefreshPeriodMs: 50.0,
		}
		for i := 0; i < config.MaxSeqLength; i++ {
			acc.GainCellArray.Storage[i] = make([]float64, config.NumHeads*config.HeadDim)
		}
	}

	// Initialize crossbars for QK^T computation
	acc.QKCrossbar = &AnalogCrossbar{
		Rows:       config.MaxSeqLength,
		Cols:       config.HeadDim,
		NoiseLevel: 0.02,
	}

	return acc
}

// ComputeAttention performs analog attention computation
func (a *AnalogAttentionAccelerator) ComputeAttention(Q, K, V [][]float64) [][]float64 {
	seqLen := len(Q)
	headDim := a.Config.HeadDim

	// Step 1: Compute QK^T scores (analog MVM)
	scores := make([][]float64, seqLen)
	for i := 0; i < seqLen; i++ {
		scores[i] = make([]float64, seqLen)
		for j := 0; j < seqLen; j++ {
			dot := 0.0
			for k := 0; k < headDim && k < len(Q[i]) && k < len(K[j]); k++ {
				// Add analog noise
				noise := 1.0 + rand.NormFloat64()*a.QKCrossbar.NoiseLevel
				dot += Q[i][k] * K[j][k] * noise
			}
			scores[i][j] = dot / math.Sqrt(float64(headDim))
		}
	}

	// Step 2: Token pruning (if enabled)
	if a.Config.TokenPruning {
		scores, V = a.pruneTokens(scores, V)
	}

	// Step 3: Softmax (kernel approximation if enabled)
	attention := a.applySoftmax(scores)

	// Step 4: Attention @ V (analog MVM)
	output := make([][]float64, seqLen)
	for i := 0; i < seqLen; i++ {
		output[i] = make([]float64, headDim)
		for k := 0; k < headDim; k++ {
			sum := 0.0
			for j := 0; j < len(attention[i]) && j < len(V); j++ {
				if k < len(V[j]) {
					sum += attention[i][j] * V[j][k]
				}
			}
			output[i][k] = sum
		}
	}

	return output
}

// pruneTokens removes low-attention tokens
func (a *AnalogAttentionAccelerator) pruneTokens(scores [][]float64, V [][]float64) ([][]float64, [][]float64) {
	seqLen := len(scores)
	keepCount := int(float64(seqLen) * (1.0 - a.Config.PruningRatio))
	if keepCount < 1 {
		keepCount = 1
	}

	// Calculate token importance (sum of attention scores)
	importance := make([]float64, seqLen)
	for j := 0; j < seqLen; j++ {
		for i := 0; i < seqLen; i++ {
			importance[j] += scores[i][j]
		}
	}

	// Find top-k tokens to keep
	type indexScore struct {
		idx   int
		score float64
	}
	ranked := make([]indexScore, seqLen)
	for i := 0; i < seqLen; i++ {
		ranked[i] = indexScore{i, importance[i]}
	}

	// Sort by score (descending)
	for i := 0; i < len(ranked)-1; i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].score > ranked[i].score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	// Keep top tokens
	a.PrunedIndices = make([]int, keepCount)
	for i := 0; i < keepCount; i++ {
		a.PrunedIndices[i] = ranked[i].idx
	}

	// Create pruned scores and V
	prunedScores := make([][]float64, seqLen)
	prunedV := make([][]float64, keepCount)

	for i := 0; i < seqLen; i++ {
		prunedScores[i] = make([]float64, keepCount)
		for j := 0; j < keepCount; j++ {
			prunedScores[i][j] = scores[i][a.PrunedIndices[j]]
		}
	}

	for j := 0; j < keepCount; j++ {
		prunedV[j] = V[a.PrunedIndices[j]]
	}

	return prunedScores, prunedV
}

// applySoftmax applies softmax with optional kernel approximation
func (a *AnalogAttentionAccelerator) applySoftmax(scores [][]float64) [][]float64 {
	seqLen := len(scores)
	if seqLen == 0 {
		return scores
	}
	colLen := len(scores[0])

	attention := make([][]float64, seqLen)
	for i := 0; i < seqLen; i++ {
		attention[i] = make([]float64, colLen)

		// Find max for numerical stability
		maxScore := scores[i][0]
		for j := 1; j < colLen; j++ {
			if scores[i][j] > maxScore {
				maxScore = scores[i][j]
			}
		}

		// Compute exp and sum
		sum := 0.0
		for j := 0; j < colLen; j++ {
			attention[i][j] = math.Exp(scores[i][j] - maxScore)
			sum += attention[i][j]
		}

		// Normalize
		for j := 0; j < colLen; j++ {
			attention[i][j] /= sum
		}
	}

	return attention
}

// =============================================================================
// MIXTURE OF EXPERTS ON 3D AIMC
// =============================================================================

// MoEConfig configures Mixture of Experts
type MoEConfig struct {
	NumExperts      int
	ExpertDim       int
	HiddenDim       int
	TopK            int     // Number of experts per token
	LoadBalancing   bool
	CapacityFactor  float64 // Expert capacity
	Use3DAIMC       bool    // Map to 3D layers
}

// MoE3DAIMCAccelerator maps MoE to 3D analog in-memory computing
// Based on Nature Comp Sci: Expert → physical layer mapping
type MoE3DAIMCAccelerator struct {
	Config           *MoEConfig
	ExpertLayers     []*Expert3DLayer
	Router           *MoERouter
	LoadBalanceLoss  float64
	ThroughputTPS    float64 // Tokens per second
	EnergyEfficiency float64 // TOPS/W
}

// Expert3DLayer represents one expert mapped to physical 3D layer
type Expert3DLayer struct {
	ExpertID     int
	PhysicalLayer int
	WeightsUp    [][]float64 // d_model -> d_ff
	WeightsDown  [][]float64 // d_ff -> d_model
	Activated    bool
	Utilization  float64
}

// MoERouter routes tokens to experts
type MoERouter struct {
	RouterWeights [][]float64
	TopK          int
	Temperature   float64
}

// NewMoE3DAIMCAccelerator creates MoE accelerator
func NewMoE3DAIMCAccelerator(config *MoEConfig) *MoE3DAIMCAccelerator {
	acc := &MoE3DAIMCAccelerator{
		Config:       config,
		ExpertLayers: make([]*Expert3DLayer, config.NumExperts),
		Router: &MoERouter{
			RouterWeights: make([][]float64, config.ExpertDim),
			TopK:          config.TopK,
			Temperature:   1.0,
		},
	}

	// Initialize experts (each maps to physical 3D layer)
	for i := 0; i < config.NumExperts; i++ {
		acc.ExpertLayers[i] = &Expert3DLayer{
			ExpertID:      i,
			PhysicalLayer: i, // Direct mapping: expert i -> layer i
			WeightsUp:     make([][]float64, config.ExpertDim),
			WeightsDown:   make([][]float64, config.HiddenDim),
		}

		// Initialize weights
		for j := 0; j < config.ExpertDim; j++ {
			acc.ExpertLayers[i].WeightsUp[j] = make([]float64, config.HiddenDim)
			for k := 0; k < config.HiddenDim; k++ {
				acc.ExpertLayers[i].WeightsUp[j][k] = rand.NormFloat64() * 0.02
			}
		}
		for j := 0; j < config.HiddenDim; j++ {
			acc.ExpertLayers[i].WeightsDown[j] = make([]float64, config.ExpertDim)
			for k := 0; k < config.ExpertDim; k++ {
				acc.ExpertLayers[i].WeightsDown[j][k] = rand.NormFloat64() * 0.02
			}
		}
	}

	// Initialize router
	for i := 0; i < config.ExpertDim; i++ {
		acc.Router.RouterWeights[i] = make([]float64, config.NumExperts)
		for j := 0; j < config.NumExperts; j++ {
			acc.Router.RouterWeights[i][j] = rand.NormFloat64() * 0.01
		}
	}

	return acc
}

// Forward performs MoE forward pass with 3D mapping
func (m *MoE3DAIMCAccelerator) Forward(input [][]float64) [][]float64 {
	batchSize := len(input)
	output := make([][]float64, batchSize)

	for b := 0; b < batchSize; b++ {
		// Step 1: Route token to experts
		expertScores := m.route(input[b])
		topExperts, topScores := m.topK(expertScores)

		// Step 2: Compute expert outputs (on 3D layers)
		output[b] = make([]float64, m.Config.ExpertDim)
		totalWeight := 0.0

		for i := 0; i < len(topExperts); i++ {
			expertIdx := topExperts[i]
			weight := topScores[i]
			totalWeight += weight

			expert := m.ExpertLayers[expertIdx]
			expert.Activated = true
			expert.Utilization += 1.0 / float64(batchSize)

			// Expert computation (on physical layer)
			expertOut := m.computeExpert(expert, input[b])

			// Weighted sum
			for j := 0; j < len(output[b]) && j < len(expertOut); j++ {
				output[b][j] += weight * expertOut[j]
			}
		}

		// Normalize
		if totalWeight > 0 {
			for j := range output[b] {
				output[b][j] /= totalWeight
			}
		}
	}

	// Compute load balance loss
	if m.Config.LoadBalancing {
		m.computeLoadBalanceLoss()
	}

	return output
}

// route computes routing scores
func (m *MoE3DAIMCAccelerator) route(token []float64) []float64 {
	scores := make([]float64, m.Config.NumExperts)

	for j := 0; j < m.Config.NumExperts; j++ {
		for i := 0; i < len(token) && i < len(m.Router.RouterWeights); i++ {
			scores[j] += token[i] * m.Router.RouterWeights[i][j]
		}
	}

	// Softmax
	maxScore := scores[0]
	for j := 1; j < len(scores); j++ {
		if scores[j] > maxScore {
			maxScore = scores[j]
		}
	}

	sum := 0.0
	for j := range scores {
		scores[j] = math.Exp((scores[j] - maxScore) / m.Router.Temperature)
		sum += scores[j]
	}
	for j := range scores {
		scores[j] /= sum
	}

	return scores
}

// topK selects top-k experts
func (m *MoE3DAIMCAccelerator) topK(scores []float64) ([]int, []float64) {
	type idxScore struct {
		idx   int
		score float64
	}

	ranked := make([]idxScore, len(scores))
	for i, s := range scores {
		ranked[i] = idxScore{i, s}
	}

	// Sort descending
	for i := 0; i < len(ranked)-1; i++ {
		for j := i + 1; j < len(ranked); j++ {
			if ranked[j].score > ranked[i].score {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}

	k := m.Config.TopK
	if k > len(ranked) {
		k = len(ranked)
	}

	indices := make([]int, k)
	scores_out := make([]float64, k)
	for i := 0; i < k; i++ {
		indices[i] = ranked[i].idx
		scores_out[i] = ranked[i].score
	}

	return indices, scores_out
}

// computeExpert runs expert on its 3D layer
func (m *MoE3DAIMCAccelerator) computeExpert(expert *Expert3DLayer, input []float64) []float64 {
	// Up projection
	hidden := make([]float64, m.Config.HiddenDim)
	for j := 0; j < m.Config.HiddenDim; j++ {
		for i := 0; i < len(input) && i < len(expert.WeightsUp); i++ {
			hidden[j] += input[i] * expert.WeightsUp[i][j]
		}
		// GELU activation
		hidden[j] = hidden[j] * 0.5 * (1.0 + math.Tanh(math.Sqrt(2.0/math.Pi)*(hidden[j]+0.044715*math.Pow(hidden[j], 3))))
	}

	// Down projection
	output := make([]float64, m.Config.ExpertDim)
	for j := 0; j < m.Config.ExpertDim; j++ {
		for i := 0; i < m.Config.HiddenDim && i < len(hidden); i++ {
			output[j] += hidden[i] * expert.WeightsDown[i][j]
		}
	}

	return output
}

// computeLoadBalanceLoss calculates load balance auxiliary loss
func (m *MoE3DAIMCAccelerator) computeLoadBalanceLoss() {
	// f_i = fraction of tokens routed to expert i
	// P_i = average probability assigned to expert i
	// Loss = alpha * sum(f_i * P_i)

	N := float64(m.Config.NumExperts)
	loss := 0.0

	for _, expert := range m.ExpertLayers {
		// Ideally uniform: 1/N
		fi := expert.Utilization
		loss += fi * fi // Penalize imbalance
	}

	m.LoadBalanceLoss = N * loss
}

// =============================================================================
// PART 2: 3D INTEGRATION FOR CIM
// =============================================================================

// Integration3DConfig configures 3D CIM integration
type Integration3DConfig struct {
	NumLayers         int
	LayerType         Layer3DType
	TSVPitch          float64 // Micrometers
	ThermalTSVRatio   float64 // Fraction of TSVs for thermal
	BondingType       string  // "microbump", "hybrid", "direct"
	TotalThickness    float64 // mm
	PowerBudgetW      float64
}

// Layer3DType represents 3D memory layer type
type Layer3DType string

const (
	Layer3DFeNAND    Layer3DType = "fenand"      // Ferroelectric NAND
	Layer3DIGZO      Layer3DType = "igzo_dram"   // IGZO 2T0C DRAM
	Layer3DRRAM      Layer3DType = "rram"        // ReRAM
	Layer3DSRAM      Layer3DType = "sram_cim"    // SRAM CIM
	Layer3DHybrid    Layer3DType = "hybrid"      // Mixed technologies
)

// Stack3DCIM represents a 3D stacked CIM system
type Stack3DCIM struct {
	Config           *Integration3DConfig
	Layers           []*CIMLayer3D
	TSVNetwork       *TSVNetwork
	ThermalModel     *ThermalModel3D
	TotalAreaMm2     float64
	TotalPowerW      float64
	DensityImprovement float64 // vs 2D
	EfficiencyTOPSW  float64
}

// CIMLayer3D represents one layer in 3D stack
type CIMLayer3D struct {
	LayerIndex      int
	LayerType       Layer3DType
	ArraySize       int
	WeightBits      int
	NumConductance  int   // MLC states
	AreaMm2         float64
	PowerW          float64
	TemperatureC    float64
}

// TSVNetwork represents through-silicon via network
type TSVNetwork struct {
	TotalTSVs       int
	SignalTSVs      int
	ThermalTSVs     int
	PowerTSVs       int
	TSVPitch        float64
	TSVDiameter     float64
	Bandwidth       float64 // GB/s
}

// ThermalModel3D models thermal behavior
type ThermalModel3D struct {
	LayerTemperatures []float64
	HotspotTemp       float64
	AmbientTemp       float64
	ThermalResistance [][]float64 // Layer-to-layer
	CoolingType       string       // "passive", "active", "liquid"
}

// NewStack3DCIM creates 3D CIM stack
func NewStack3DCIM(config *Integration3DConfig) *Stack3DCIM {
	stack := &Stack3DCIM{
		Config: config,
		Layers: make([]*CIMLayer3D, config.NumLayers),
	}

	// Initialize layers
	for i := 0; i < config.NumLayers; i++ {
		stack.Layers[i] = &CIMLayer3D{
			LayerIndex:     i,
			LayerType:      config.LayerType,
			ArraySize:      256,
			WeightBits:     8,
			NumConductance: 256, // 8-bit MLC
		}

		switch config.LayerType {
		case Layer3DFeNAND:
			stack.Layers[i].NumConductance = 256 // ≥256 levels per SK hynix
			stack.Layers[i].AreaMm2 = 0.01       // Very dense
		case Layer3DIGZO:
			stack.Layers[i].NumConductance = 16
			stack.Layers[i].AreaMm2 = 0.05
		case Layer3DRRAM:
			stack.Layers[i].NumConductance = 64
			stack.Layers[i].AreaMm2 = 0.02
		case Layer3DSRAM:
			stack.Layers[i].NumConductance = 2 // Binary
			stack.Layers[i].AreaMm2 = 0.1
		}
	}

	// Initialize TSV network
	stack.TSVNetwork = &TSVNetwork{
		TSVPitch:    config.TSVPitch,
		TSVDiameter: config.TSVPitch * 0.5,
	}

	// Calculate TSV counts
	areaPerLayer := stack.Layers[0].AreaMm2
	tsvDensity := 1.0 / (config.TSVPitch * config.TSVPitch * 1e-6) // per mm²
	stack.TSVNetwork.TotalTSVs = int(areaPerLayer * tsvDensity)
	stack.TSVNetwork.ThermalTSVs = int(float64(stack.TSVNetwork.TotalTSVs) * config.ThermalTSVRatio)
	stack.TSVNetwork.SignalTSVs = stack.TSVNetwork.TotalTSVs - stack.TSVNetwork.ThermalTSVs
	stack.TSVNetwork.PowerTSVs = stack.TSVNetwork.SignalTSVs / 10

	// Initialize thermal model
	stack.ThermalModel = &ThermalModel3D{
		LayerTemperatures: make([]float64, config.NumLayers),
		AmbientTemp:       25.0,
		CoolingType:       "passive",
	}

	// Calculate density improvement
	// 3D FeNAND: 4000× vs 2D (from SK hynix IEDM 2024)
	if config.LayerType == Layer3DFeNAND {
		stack.DensityImprovement = 4000.0
		stack.EfficiencyTOPSW = 1000.0 // 1000× TOPS/mm² improvement
	} else {
		stack.DensityImprovement = float64(config.NumLayers)
		stack.EfficiencyTOPSW = float64(config.NumLayers) * 10.0
	}

	return stack
}

// SimulateThermal runs thermal simulation
func (s *Stack3DCIM) SimulateThermal(workloadW float64) {
	// Simple thermal model: heat flows up
	// Bottom layer is hottest

	baseTemp := s.ThermalModel.AmbientTemp
	tempRisePerLayer := workloadW / float64(s.Config.NumLayers) * 5.0 // 5°C/W per layer

	for i := 0; i < s.Config.NumLayers; i++ {
		// Heat accumulates from bottom
		heatAccum := float64(s.Config.NumLayers-i) * tempRisePerLayer
		s.Layers[i].TemperatureC = baseTemp + heatAccum
		s.ThermalModel.LayerTemperatures[i] = s.Layers[i].TemperatureC
	}

	// Find hotspot
	s.ThermalModel.HotspotTemp = s.ThermalModel.LayerTemperatures[0]
	for _, t := range s.ThermalModel.LayerTemperatures {
		if t > s.ThermalModel.HotspotTemp {
			s.ThermalModel.HotspotTemp = t
		}
	}
}

// =============================================================================
// 3D FeNAND FOR CIM
// =============================================================================

// FeNAND3DConfig configures 3D ferroelectric NAND
type FeNAND3DConfig struct {
	NumLayers         int
	StringLength      int     // Cells per string
	ConductanceLevels int     // MLC states (≥256 for SK hynix)
	GateStackType     string  // "hzo", "superlattice"
	MemoryWindowV     float64
	ProgramVoltage    float64
	EraseVoltage      float64
}

// FeNAND3DArray represents 3D ferroelectric NAND array
// Based on SK hynix IEDM 2024: 4000× density, 87.8% accuracy
type FeNAND3DArray struct {
	Config            *FeNAND3DConfig
	ConductanceStates [][][]float64 // [layer][row][col]
	AccuracyMNIST     float64
	DensityImprovement float64
	TOPSPerMm2        float64
}

// NewFeNAND3DArray creates 3D FeNAND array
func NewFeNAND3DArray(config *FeNAND3DConfig) *FeNAND3DArray {
	arr := &FeNAND3DArray{
		Config:             config,
		ConductanceStates:  make([][][]float64, config.NumLayers),
		AccuracyMNIST:      0.878,  // 87.8% from literature
		DensityImprovement: 4000.0, // vs 2D
		TOPSPerMm2:         1000.0, // 1000× vs 2D
	}

	// Initialize states
	for l := 0; l < config.NumLayers; l++ {
		arr.ConductanceStates[l] = make([][]float64, config.StringLength)
		for r := 0; r < config.StringLength; r++ {
			arr.ConductanceStates[l][r] = make([]float64, config.StringLength)
		}
	}

	return arr
}

// ProgramWeight programs a weight into the array
func (f *FeNAND3DArray) ProgramWeight(layer, row, col int, weight float64) {
	if layer < 0 || layer >= f.Config.NumLayers {
		return
	}
	if row < 0 || row >= f.Config.StringLength {
		return
	}
	if col < 0 || col >= f.Config.StringLength {
		return
	}

	// Quantize to available conductance levels
	levels := float64(f.Config.ConductanceLevels)
	normalized := (weight + 1.0) / 2.0 // Map [-1,1] to [0,1]
	quantized := math.Round(normalized*levels) / levels
	f.ConductanceStates[layer][row][col] = quantized*2.0 - 1.0 // Back to [-1,1]
}

// ComputeMVM performs matrix-vector multiplication
func (f *FeNAND3DArray) ComputeMVM(layer int, input []float64) []float64 {
	if layer < 0 || layer >= f.Config.NumLayers {
		return nil
	}

	stringLen := f.Config.StringLength
	output := make([]float64, stringLen)

	for i := 0; i < stringLen; i++ {
		sum := 0.0
		for j := 0; j < stringLen && j < len(input); j++ {
			sum += f.ConductanceStates[layer][i][j] * input[j]
		}
		output[i] = sum
	}

	return output
}

// =============================================================================
// CHIPLET INTEGRATION
// =============================================================================

// ChipletConfig configures CIM chiplet
type ChipletConfig struct {
	ChipletType      ChipletType
	TechnologyNode   int     // nm
	ArraySize        int
	MemoryType       string  // "rram", "fefet", "sram", "edram"
	InterfaceType    string  // "ucle", "aib", "lipincon"
	Bandwidth        float64 // GB/s
}

// ChipletType represents chiplet function
type ChipletType string

const (
	ChipletCIMRRAM   ChipletType = "cim_rram"
	ChipletCIMFeFET  ChipletType = "cim_fefet"
	ChipletCIMSRAM   ChipletType = "cim_sram"
	ChipletCIMEDRAM  ChipletType = "cim_edram"
	ChipletDigital   ChipletType = "digital_compute"
	ChipletRouter    ChipletType = "router_nop"
	ChipletIO        ChipletType = "io_interface"
)

// Chiplet3DCIMlet represents a CIM chiplet
type Chiplet3DCIMlet struct {
	Config          *ChipletConfig
	ID              int
	Position        [2]int  // X, Y on package
	StackPosition   int     // Z for 3D
	EnergyEfficiency float64 // TOPS/W
	AreaEfficiency  float64 // TOPS/mm²
}

// HeterogeneousPackage represents 2.5D/3D chiplet package
// Based on 3D-CIMlet: 12× energy efficiency
type HeterogeneousPackage struct {
	PackageType      string // "2.5d", "3d", "hybrid"
	Chiplets         []*Chiplet3DCIMlet
	Interposer       *Interposer
	NetworkOnPackage *NoP
	TotalTOPS        float64
	TotalPowerW      float64
	EnergyEffImprove float64 // vs 2D baseline
	EDPReduction     float64 // Energy-delay product
}

// Interposer represents silicon interposer or RDL
type Interposer struct {
	Technology      string  // "passive_si", "active_si", "rdl"
	InterconnectPitch float64 // um
	Bandwidth       float64 // TB/s
}

// NoP represents Network on Package
type NoP struct {
	Topology       string // "mesh", "ring", "tree"
	RouterLatency  int    // cycles
	LinkBandwidth  float64 // GB/s per link
}

// NewHeterogeneousPackage creates chiplet package
func NewHeterogeneousPackage(packageType string, chiplets []*ChipletConfig) *HeterogeneousPackage {
	pkg := &HeterogeneousPackage{
		PackageType:      packageType,
		Chiplets:         make([]*Chiplet3DCIMlet, len(chiplets)),
		EnergyEffImprove: 1.0,
		EDPReduction:     0.0,
	}

	// Create chiplets
	for i, cfg := range chiplets {
		pkg.Chiplets[i] = &Chiplet3DCIMlet{
			Config:   cfg,
			ID:       i,
			Position: [2]int{i % 4, i / 4},
		}

		// Set efficiency based on type
		switch cfg.ChipletType {
		case ChipletCIMRRAM:
			pkg.Chiplets[i].EnergyEfficiency = 100.0
			pkg.Chiplets[i].AreaEfficiency = 50.0
		case ChipletCIMFeFET:
			pkg.Chiplets[i].EnergyEfficiency = 50.0
			pkg.Chiplets[i].AreaEfficiency = 100.0
		case ChipletCIMSRAM:
			pkg.Chiplets[i].EnergyEfficiency = 20.0
			pkg.Chiplets[i].AreaEfficiency = 10.0
		case ChipletCIMEDRAM:
			pkg.Chiplets[i].EnergyEfficiency = 80.0
			pkg.Chiplets[i].AreaEfficiency = 40.0
		}
	}

	// Initialize interposer
	switch packageType {
	case "2.5d":
		pkg.Interposer = &Interposer{
			Technology:        "passive_si",
			InterconnectPitch: 10.0, // um
			Bandwidth:         2.0,  // TB/s
		}
		pkg.EnergyEffImprove = 9.3 // From 3D-CIMlet paper
	case "3d":
		pkg.Interposer = &Interposer{
			Technology:        "hybrid_bond",
			InterconnectPitch: 1.0, // um
			Bandwidth:         10.0,
		}
		pkg.EnergyEffImprove = 12.0 // From 3D-CIMlet paper
		pkg.EDPReduction = 0.925    // 92.5% EDP reduction
	}

	// Initialize NoP
	pkg.NetworkOnPackage = &NoP{
		Topology:      "mesh",
		RouterLatency: 2,
		LinkBandwidth: 100.0, // GB/s
	}

	return pkg
}

// MapLLMToChiplets maps LLM layers to chiplets
func (h *HeterogeneousPackage) MapLLMToChiplets(modelLayers int) map[int]int {
	mapping := make(map[int]int)

	// Round-robin mapping
	for l := 0; l < modelLayers; l++ {
		chipletIdx := l % len(h.Chiplets)
		mapping[l] = chipletIdx
	}

	return mapping
}

// =============================================================================
// BENCHMARK AND EVALUATION
// =============================================================================

// LLM3DBenchmark benchmarks LLM inference on 3D CIM
type LLM3DBenchmark struct {
	ModelName           string
	Parameters          int64
	SequenceLength      int
	TokensPerSecond     float64
	PrefillLatencyMs    float64
	DecodeLatencyMs     float64
	EnergyPerToken      float64 // μJ
	KVCacheMemoryMB     float64
	EnergyReduction     float64 // vs GPU
	SpeedupVsGPU        float64
}

// RunLLM3DBenchmark evaluates LLM on 3D CIM
func RunLLM3DBenchmark(modelName string, params int64, seqLen int) *LLM3DBenchmark {
	bench := &LLM3DBenchmark{
		ModelName:      modelName,
		Parameters:     params,
		SequenceLength: seqLen,
	}

	// Estimate based on model size
	paramsB := float64(params) / 1e9

	// Prefill: O(seq² * d)
	bench.PrefillLatencyMs = float64(seqLen*seqLen) * paramsB * 0.001

	// Decode: O(seq * d) per token
	bench.DecodeLatencyMs = float64(seqLen) * paramsB * 0.0001

	// Tokens per second
	bench.TokensPerSecond = 1000.0 / bench.DecodeLatencyMs

	// KV cache: 2 * layers * heads * head_dim * seq * bytes
	numLayers := int(math.Log2(paramsB) * 10)
	bench.KVCacheMemoryMB = float64(2*numLayers*32*128*seqLen*2) / (1024 * 1024)

	// Energy (analog attention gives 70,000× reduction for attention)
	// Attention is ~70% of energy in digital
	digitalEnergyPerToken := paramsB * 10.0 // μJ baseline
	analogAttentionEnergy := digitalEnergyPerToken * 0.7 / 70000.0
	otherEnergy := digitalEnergyPerToken * 0.3
	bench.EnergyPerToken = analogAttentionEnergy + otherEnergy

	// Overall metrics
	bench.EnergyReduction = digitalEnergyPerToken / bench.EnergyPerToken
	bench.SpeedupVsGPU = 100.0 // From literature

	return bench
}

// PrintLLM3DBenchmark formats benchmark results
func PrintLLM3DBenchmark(bench *LLM3DBenchmark) string {
	return fmt.Sprintf(`LLM 3D CIM Benchmark
====================
Model: %s
Parameters: %.1fB
Sequence Length: %d

Performance:
  Prefill Latency: %.2f ms
  Decode Latency: %.3f ms/token
  Tokens/Second: %.0f

Memory:
  KV Cache: %.1f MB

Efficiency vs GPU:
  Energy Reduction: %.0f×
  Speedup: %.0f×
  Energy/Token: %.4f μJ
`, bench.ModelName,
		float64(bench.Parameters)/1e9,
		bench.SequenceLength,
		bench.PrefillLatencyMs,
		bench.DecodeLatencyMs,
		bench.TokensPerSecond,
		bench.KVCacheMemoryMB,
		bench.EnergyReduction,
		bench.SpeedupVsGPU,
		bench.EnergyPerToken)
}
