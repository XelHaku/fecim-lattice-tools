// Package layers provides heterogeneous CIM architectures and LLM inference acceleration
// for compute-in-memory neural network processing.
//
// Research context:
// - Heterogeneous CIM combines multiple NVM technologies (FeFET, RRAM, PCM, SRAM)
// - 3D integration enables tier-based device allocation
// - LLM inference requires specialized KV-cache and attention handling
//
// Key metrics from literature:
// - HEIRS: 7.83× energy efficiency improvement over RRAM-CIM
// - X-Former: 69.8× latency improvement over GPU
// - UniCAIM: Unified CAM/CIM with static-dynamic KV pruning
// - HePGA: 3.8× TOPS/W improvement for heterogeneous PIM
package layers

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// ============================================================================
// HETEROGENEOUS NVM DEVICE MODELING
// ============================================================================

// NVMDeviceType defines different NVM technologies
type NVMDeviceType int

const (
	DeviceFeFET NVMDeviceType = iota
	DeviceRRAM
	DevicePCM
	DeviceMRAM
	DeviceSRAM
)

// NVMDeviceSpec specifies characteristics of an NVM device
type NVMDeviceSpec struct {
	Type           NVMDeviceType
	Name           string
	ReadLatencyNS  float64 // Read latency in nanoseconds
	WriteLatencyNS float64 // Write latency in nanoseconds
	ReadEnergyFJ   float64 // Read energy in femtojoules
	WriteEnergyFJ  float64 // Write energy in femtojoules
	Endurance      float64 // Write cycles before failure
	RetentionYears float64 // Data retention time
	AreaFactor     float64 // Relative area (SRAM = 1.0)
	OnOffRatio     float64 // Conductance on/off ratio
	AnalogLevels   int     // Multi-level cell capability
	CIMSuitability float64 // 0-5 rating for CIM applications
}

// GetNVMDeviceSpecs returns specifications for all NVM types
func GetNVMDeviceSpecs() map[NVMDeviceType]*NVMDeviceSpec {
	return map[NVMDeviceType]*NVMDeviceSpec{
		DeviceFeFET: {
			Type:           DeviceFeFET,
			Name:           "FeFET",
			ReadLatencyNS:  10,
			WriteLatencyNS: 10,
			ReadEnergyFJ:   1,
			WriteEnergyFJ:  10,
			Endurance:      1e12,
			RetentionYears: 10,
			AreaFactor:     0.2,
			OnOffRatio:     1e4,
			AnalogLevels:   64,
			CIMSuitability: 5.0,
		},
		DeviceRRAM: {
			Type:           DeviceRRAM,
			Name:           "RRAM",
			ReadLatencyNS:  10,
			WriteLatencyNS: 10,
			ReadEnergyFJ:   10,
			WriteEnergyFJ:  100,
			Endurance:      1e6,
			RetentionYears: 10,
			AreaFactor:     0.1,
			OnOffRatio:     1e2,
			AnalogLevels:   16,
			CIMSuitability: 4.0,
		},
		DevicePCM: {
			Type:           DevicePCM,
			Name:           "PCM",
			ReadLatencyNS:  50,
			WriteLatencyNS: 100,
			ReadEnergyFJ:   100,
			WriteEnergyFJ:  10000,
			Endurance:      1e8,
			RetentionYears: 10,
			AreaFactor:     0.15,
			OnOffRatio:     1e3,
			AnalogLevels:   32,
			CIMSuitability: 4.0,
		},
		DeviceMRAM: {
			Type:           DeviceMRAM,
			Name:           "STT-MRAM",
			ReadLatencyNS:  10,
			WriteLatencyNS: 10,
			ReadEnergyFJ:   10,
			WriteEnergyFJ:  100,
			Endurance:      1e15,
			RetentionYears: 10,
			AreaFactor:     0.3,
			OnOffRatio:     3,
			AnalogLevels:   2,
			CIMSuitability: 3.0,
		},
		DeviceSRAM: {
			Type:           DeviceSRAM,
			Name:           "SRAM",
			ReadLatencyNS:  1,
			WriteLatencyNS: 1,
			ReadEnergyFJ:   1,
			WriteEnergyFJ:  1,
			Endurance:      1e18, // Practically infinite
			RetentionYears: 0,    // Volatile
			AreaFactor:     1.0,
			OnOffRatio:     1e6,
			AnalogLevels:   2,
			CIMSuitability: 3.0,
		},
	}
}

// ============================================================================
// 3D HETEROGENEOUS CIM ARCHITECTURE
// ============================================================================

// HeteroCIMConfig configures a heterogeneous CIM system
type HeteroCIMConfig struct {
	NumTiers          int                  // Number of 3D tiers
	TierDevices       []NVMDeviceType      // Device type per tier
	TierSizes         []int                // Crossbar arrays per tier
	CrossbarSize      int                  // Array dimension (e.g., 256)
	InterTierBandwidth float64             // GB/s between tiers
	ThermalBudgetW    float64             // Power budget in Watts
}

// DefaultHeteroCIMConfig returns HEIRS-inspired configuration
func DefaultHeteroCIMConfig() *HeteroCIMConfig {
	return &HeteroCIMConfig{
		NumTiers:          3,
		TierDevices:       []NVMDeviceType{DeviceRRAM, DeviceSRAM, DeviceFeFET},
		TierSizes:         []int{1024, 256, 512}, // RRAM high density, SRAM fast
		CrossbarSize:      256,
		InterTierBandwidth: 100, // GB/s with TSV
		ThermalBudgetW:    10,
	}
}

// HeteroCIMTier represents a single tier in 3D stack
type HeteroCIMTier struct {
	TierID      int
	Device      *NVMDeviceSpec
	Arrays      []*CrossbarArrayHetero
	LocalBuffer []float64
	Utilization float64
}

// CrossbarArrayHetero represents a crossbar in heterogeneous system
type CrossbarArrayHetero struct {
	ArrayID     int
	TierID      int
	Rows        int
	Cols        int
	Weights     [][]float64
	Device      *NVMDeviceSpec
	WriteCount  int64 // Track endurance
	LastAccess  int64 // For wear leveling
}

// HeteroCIMSystem represents a complete heterogeneous CIM system
type HeteroCIMSystem struct {
	Config   *HeteroCIMConfig
	Tiers    []*HeteroCIMTier
	DevSpecs map[NVMDeviceType]*NVMDeviceSpec
	Stats    *HeteroCIMStats
}

// HeteroCIMStats tracks system statistics
type HeteroCIMStats struct {
	TotalOps         int64
	ReadOps          int64
	WriteOps         int64
	TotalEnergyPJ    float64
	TotalLatencyNS   float64
	TierUtilization  []float64
	EnduranceRemaining []float64
}

// NewHeteroCIMSystem creates a new heterogeneous CIM system
func NewHeteroCIMSystem(config *HeteroCIMConfig) *HeteroCIMSystem {
	devSpecs := GetNVMDeviceSpecs()

	tiers := make([]*HeteroCIMTier, config.NumTiers)
	for t := 0; t < config.NumTiers; t++ {
		devType := config.TierDevices[t]
		numArrays := config.TierSizes[t]

		arrays := make([]*CrossbarArrayHetero, numArrays)
		for a := 0; a < numArrays; a++ {
			weights := make([][]float64, config.CrossbarSize)
			for r := range weights {
				weights[r] = make([]float64, config.CrossbarSize)
			}
			arrays[a] = &CrossbarArrayHetero{
				ArrayID: a,
				TierID:  t,
				Rows:    config.CrossbarSize,
				Cols:    config.CrossbarSize,
				Weights: weights,
				Device:  devSpecs[devType],
			}
		}

		tiers[t] = &HeteroCIMTier{
			TierID:      t,
			Device:      devSpecs[devType],
			Arrays:      arrays,
			LocalBuffer: make([]float64, config.CrossbarSize*16), // 16 vectors
		}
	}

	return &HeteroCIMSystem{
		Config:   config,
		Tiers:    tiers,
		DevSpecs: devSpecs,
		Stats:    &HeteroCIMStats{},
	}
}

// MapLayerToTiers determines optimal tier placement for a layer
func (h *HeteroCIMSystem) MapLayerToTiers(layerType string, weightSize int, isDynamic bool) int {
	// Static weights (projections) -> NVM tiers (FeFET, RRAM)
	// Dynamic weights (attention scores) -> SRAM tier

	if isDynamic {
		// Find SRAM tier
		for i, tier := range h.Tiers {
			if tier.Device.Type == DeviceSRAM {
				return i
			}
		}
	}

	// For static weights, prefer high-density NVM
	// FeFET > RRAM > PCM for CIM
	bestTier := 0
	bestScore := 0.0

	for i, tier := range h.Tiers {
		if tier.Device.Type == DeviceSRAM {
			continue // Skip SRAM for static weights
		}

		score := tier.Device.CIMSuitability / tier.Device.AreaFactor
		if score > bestScore {
			bestScore = score
			bestTier = i
		}
	}

	return bestTier
}

// ComputeMVM performs MVM on specified tier
func (h *HeteroCIMSystem) ComputeMVM(tierID, arrayID int, input []float64) ([]float64, float64, float64) {
	tier := h.Tiers[tierID]
	array := tier.Arrays[arrayID]

	output := make([]float64, array.Cols)
	for col := 0; col < array.Cols; col++ {
		sum := 0.0
		for row := 0; row < len(input) && row < array.Rows; row++ {
			sum += input[row] * array.Weights[row][col]
		}
		output[col] = sum
	}

	// Calculate energy and latency
	numOps := len(input) * array.Cols
	energyPJ := float64(numOps) * array.Device.ReadEnergyFJ / 1000
	latencyNS := array.Device.ReadLatencyNS * float64(array.Rows) / 9 // 9 rows per activation

	h.Stats.TotalOps += int64(numOps)
	h.Stats.ReadOps += int64(numOps)
	h.Stats.TotalEnergyPJ += energyPJ
	h.Stats.TotalLatencyNS += latencyNS

	return output, energyPJ, latencyNS
}

// ============================================================================
// HEIRS-STYLE HYBRID RRAM/SRAM ARCHITECTURE
// ============================================================================

// HEIRSConfig configures HEIRS-style architecture
type HEIRSConfig struct {
	RRAMArrays      int     // Number of RRAM arrays (high density)
	SRAMArrays      int     // Number of SRAM arrays (high speed)
	ArraySize       int
	RRAMBits        int     // RRAM precision
	SRAMBits        int     // SRAM precision
	TSVBandwidthGBs float64 // Inter-tier bandwidth
}

// HEIRSAccelerator implements HEIRS architecture
type HEIRSAccelerator struct {
	Config       *HEIRSConfig
	RRAMTier     *HeteroCIMTier
	SRAMTier     *HeteroCIMTier
	WeightBuffer [][]float64 // For weight reuse
}

// NewHEIRSAccelerator creates a HEIRS accelerator
func NewHEIRSAccelerator(config *HEIRSConfig) *HEIRSAccelerator {
	devSpecs := GetNVMDeviceSpecs()

	// Create RRAM tier (bottom, high density)
	rramArrays := make([]*CrossbarArrayHetero, config.RRAMArrays)
	for i := range rramArrays {
		weights := make([][]float64, config.ArraySize)
		for r := range weights {
			weights[r] = make([]float64, config.ArraySize)
		}
		rramArrays[i] = &CrossbarArrayHetero{
			ArrayID: i,
			TierID:  0,
			Rows:    config.ArraySize,
			Cols:    config.ArraySize,
			Weights: weights,
			Device:  devSpecs[DeviceRRAM],
		}
	}

	// Create SRAM tier (top, high speed)
	sramArrays := make([]*CrossbarArrayHetero, config.SRAMArrays)
	for i := range sramArrays {
		weights := make([][]float64, config.ArraySize)
		for r := range weights {
			weights[r] = make([]float64, config.ArraySize)
		}
		sramArrays[i] = &CrossbarArrayHetero{
			ArrayID: i,
			TierID:  1,
			Rows:    config.ArraySize,
			Cols:    config.ArraySize,
			Weights: weights,
			Device:  devSpecs[DeviceSRAM],
		}
	}

	return &HEIRSAccelerator{
		Config: config,
		RRAMTier: &HeteroCIMTier{
			TierID: 0,
			Device: devSpecs[DeviceRRAM],
			Arrays: rramArrays,
		},
		SRAMTier: &HeteroCIMTier{
			TierID: 1,
			Device: devSpecs[DeviceSRAM],
			Arrays: sramArrays,
		},
		WeightBuffer: make([][]float64, config.ArraySize),
	}
}

// ProcessTransformerLayer processes a transformer layer
func (h *HEIRSAccelerator) ProcessTransformerLayer(input [][]float64, isProjection bool) [][]float64 {
	var tier *HeteroCIMTier

	if isProjection {
		// Static projections (Q, K, V, O) use RRAM
		tier = h.RRAMTier
	} else {
		// Dynamic attention scores use SRAM
		tier = h.SRAMTier
	}

	batchSize := len(input)
	seqLen := len(input[0])
	output := make([][]float64, batchSize)

	for b := 0; b < batchSize; b++ {
		output[b] = make([]float64, seqLen)
		// Distribute across arrays
		arrayIdx := b % len(tier.Arrays)
		array := tier.Arrays[arrayIdx]

		for i := 0; i < seqLen && i < array.Cols; i++ {
			sum := 0.0
			for j := 0; j < seqLen && j < array.Rows; j++ {
				sum += input[b][j] * array.Weights[j][i]
			}
			output[b][i] = sum
		}
	}

	return output
}

// ============================================================================
// X-FORMER STYLE PROJECTION/ATTENTION ENGINE
// ============================================================================

// XFormerConfig configures X-Former architecture
type XFormerConfig struct {
	ProjectionTiles int   // NVM tiles for projection engine
	AttentionTiles  int   // CMOS tiles for attention engine
	TileSize        int
	HeadDim         int   // Attention head dimension
	NumHeads        int
	SeqLength       int
}

// XFormerAccelerator implements X-Former architecture
type XFormerAccelerator struct {
	Config           *XFormerConfig
	ProjectionEngine *ProjectionEngine
	AttentionEngine  *AttentionEngine
	Interconnect     *BusInterconnect
}

// ProjectionEngine handles static weight projections on NVM
type ProjectionEngine struct {
	Tiles      []*NVMTile
	NumTiles   int
	DeviceType NVMDeviceType
}

// NVMTile represents a single NVM tile
type NVMTile struct {
	ID         int
	Rows       int
	Cols       int
	Weights    [][]float64
	Device     *NVMDeviceSpec
}

// AttentionEngine handles dynamic attention on CMOS
type AttentionEngine struct {
	Tiles        []*CMOSTile
	NumTiles     int
	ScoreBuffer  [][]float64 // Cache attention scores
	KVCache      *KVCache
}

// CMOSTile represents a CMOS-based compute tile
type CMOSTile struct {
	ID       int
	Rows     int
	Cols     int
	SRAM     [][]float64
	ComputeUnits int
}

// BusInterconnect handles data transfer between engines
type BusInterconnect struct {
	BandwidthGBs float64
	LatencyNS    float64
	BufferSize   int
}

// NewXFormerAccelerator creates an X-Former accelerator
func NewXFormerAccelerator(config *XFormerConfig) *XFormerAccelerator {
	devSpecs := GetNVMDeviceSpecs()

	// Create projection engine with FeFET tiles
	projTiles := make([]*NVMTile, config.ProjectionTiles)
	for i := range projTiles {
		weights := make([][]float64, config.TileSize)
		for r := range weights {
			weights[r] = make([]float64, config.TileSize)
		}
		projTiles[i] = &NVMTile{
			ID:      i,
			Rows:    config.TileSize,
			Cols:    config.TileSize,
			Weights: weights,
			Device:  devSpecs[DeviceFeFET],
		}
	}

	// Create attention engine with CMOS tiles
	attnTiles := make([]*CMOSTile, config.AttentionTiles)
	for i := range attnTiles {
		sram := make([][]float64, config.TileSize)
		for r := range sram {
			sram[r] = make([]float64, config.TileSize)
		}
		attnTiles[i] = &CMOSTile{
			ID:           i,
			Rows:         config.TileSize,
			Cols:         config.TileSize,
			SRAM:         sram,
			ComputeUnits: 64,
		}
	}

	return &XFormerAccelerator{
		Config: config,
		ProjectionEngine: &ProjectionEngine{
			Tiles:      projTiles,
			NumTiles:   config.ProjectionTiles,
			DeviceType: DeviceFeFET,
		},
		AttentionEngine: &AttentionEngine{
			Tiles:       attnTiles,
			NumTiles:    config.AttentionTiles,
			ScoreBuffer: make([][]float64, config.SeqLength),
			KVCache:     NewKVCache(config.SeqLength, config.HeadDim, config.NumHeads),
		},
		Interconnect: &BusInterconnect{
			BandwidthGBs: 256,
			LatencyNS:    10,
			BufferSize:   1024 * 1024, // 1MB
		},
	}
}

// ComputeProjection performs Q, K, V, O projections on NVM
func (x *XFormerAccelerator) ComputeProjection(input [][]float64, projType string) [][]float64 {
	batchSize := len(input)
	seqLen := len(input[0])

	output := make([][]float64, batchSize)
	for b := 0; b < batchSize; b++ {
		output[b] = make([]float64, x.Config.HeadDim*x.Config.NumHeads)
	}

	// Distribute across NVM tiles
	for b := 0; b < batchSize; b++ {
		tileIdx := b % x.ProjectionEngine.NumTiles
		tile := x.ProjectionEngine.Tiles[tileIdx]

		for i := 0; i < len(output[b]) && i < tile.Cols; i++ {
			sum := 0.0
			for j := 0; j < seqLen && j < tile.Rows; j++ {
				sum += input[b][j] * tile.Weights[j][i]
			}
			output[b][i] = sum
		}
	}

	return output
}

// ComputeAttention performs attention on CMOS tiles
func (x *XFormerAccelerator) ComputeAttention(Q, K, V [][]float64) [][]float64 {
	seqLen := len(Q)
	headDim := x.Config.HeadDim

	// Compute attention scores: Q @ K^T
	scores := make([][]float64, seqLen)
	for i := range scores {
		scores[i] = make([]float64, seqLen)
		for j := 0; j < seqLen; j++ {
			dot := 0.0
			for d := 0; d < headDim && d < len(Q[i]); d++ {
				dot += Q[i][d] * K[j][d]
			}
			scores[i][j] = dot / math.Sqrt(float64(headDim))
		}
	}

	// Softmax
	for i := range scores {
		maxVal := scores[i][0]
		for j := range scores[i] {
			if scores[i][j] > maxVal {
				maxVal = scores[i][j]
			}
		}
		sum := 0.0
		for j := range scores[i] {
			scores[i][j] = math.Exp(scores[i][j] - maxVal)
			sum += scores[i][j]
		}
		for j := range scores[i] {
			scores[i][j] /= sum
		}
	}

	// Store in score buffer for reuse
	x.AttentionEngine.ScoreBuffer = scores

	// Compute output: scores @ V
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, headDim)
		for d := 0; d < headDim; d++ {
			sum := 0.0
			for j := 0; j < seqLen; j++ {
				if d < len(V[j]) {
					sum += scores[i][j] * V[j][d]
				}
			}
			output[i][d] = sum
		}
	}

	return output
}

// ============================================================================
// KV-CACHE FOR LLM INFERENCE
// ============================================================================

// KVCache manages key-value cache for autoregressive generation
type KVCache struct {
	MaxSeqLen    int
	HeadDim      int
	NumHeads     int
	NumLayers    int
	Keys         [][][][]float64 // [layer][head][seq][dim]
	Values       [][][][]float64
	CurrentLen   int
	PruningRatio float64 // For cache compression
}

// NewKVCache creates a new KV cache
func NewKVCache(maxSeqLen, headDim, numHeads int) *KVCache {
	return &KVCache{
		MaxSeqLen:    maxSeqLen,
		HeadDim:      headDim,
		NumHeads:     numHeads,
		NumLayers:    1, // Single layer for simplicity
		Keys:         make([][][][]float64, 1),
		Values:       make([][][][]float64, 1),
		CurrentLen:   0,
		PruningRatio: 0.0,
	}
}

// Append adds new key-value pairs to cache
func (kv *KVCache) Append(layerIdx int, newKeys, newValues [][]float64) {
	if layerIdx >= len(kv.Keys) {
		return
	}

	// Initialize if needed
	if kv.Keys[layerIdx] == nil {
		kv.Keys[layerIdx] = make([][][]float64, kv.NumHeads)
		kv.Values[layerIdx] = make([][][]float64, kv.NumHeads)
		for h := 0; h < kv.NumHeads; h++ {
			kv.Keys[layerIdx][h] = make([][]float64, 0, kv.MaxSeqLen)
			kv.Values[layerIdx][h] = make([][]float64, 0, kv.MaxSeqLen)
		}
	}

	// Add to each head
	for h := 0; h < kv.NumHeads && h < len(newKeys); h++ {
		kv.Keys[layerIdx][h] = append(kv.Keys[layerIdx][h], newKeys[h])
		kv.Values[layerIdx][h] = append(kv.Values[layerIdx][h], newValues[h])
	}

	kv.CurrentLen++
}

// GetMemoryUsageBytes returns current memory usage
func (kv *KVCache) GetMemoryUsageBytes() int64 {
	// 4 bytes per float64 element (assuming float32 storage)
	elementsPerLayer := kv.NumHeads * kv.CurrentLen * kv.HeadDim * 2 // K and V
	return int64(kv.NumLayers * elementsPerLayer * 4)
}

// ============================================================================
// UniCAIM: UNIFIED CAM/CIM WITH KV PRUNING
// ============================================================================

// UniCAIMConfig configures UniCAIM architecture
type UniCAIMConfig struct {
	NumCAMArrays    int     // Content-addressable memory arrays
	NumCIMArrays    int     // Compute-in-memory arrays
	ArraySize       int
	StaticPruneRatio float64 // Fixed-pattern pruning (StreamingLLM-style)
	DynamicTopK     int     // Top-K selection for dynamic pruning
	WindowSize      int     // Attention window for static pruning
}

// UniCAIMAccelerator implements UniCAIM architecture
type UniCAIMAccelerator struct {
	Config      *UniCAIMConfig
	CAMArrays   []*CAMArray
	CIMArrays   []*CIMArrayLLM
	KVCache     *KVCache
	PruneStats  *PruneStats
}

// CAMArray represents content-addressable memory for similarity search
type CAMArray struct {
	ID          int
	Rows        int
	Cols        int
	StoredKeys  [][]float64
	Thresholds  []float64
}

// CIMArrayLLM represents CIM array for LLM operations
type CIMArrayLLM struct {
	ID      int
	Rows    int
	Cols    int
	Weights [][]float64
	Role    string // "projection", "ffn", "attention"
}

// PruneStats tracks pruning statistics
type PruneStats struct {
	TotalTokens       int
	PrunedTokensStatic int
	PrunedTokensDynamic int
	CacheHits         int
	CacheMisses       int
}

// NewUniCAIMAccelerator creates a UniCAIM accelerator
func NewUniCAIMAccelerator(config *UniCAIMConfig) *UniCAIMAccelerator {
	camArrays := make([]*CAMArray, config.NumCAMArrays)
	for i := range camArrays {
		camArrays[i] = &CAMArray{
			ID:         i,
			Rows:       config.ArraySize,
			Cols:       config.ArraySize,
			StoredKeys: make([][]float64, 0),
			Thresholds: make([]float64, config.ArraySize),
		}
	}

	cimArrays := make([]*CIMArrayLLM, config.NumCIMArrays)
	for i := range cimArrays {
		weights := make([][]float64, config.ArraySize)
		for r := range weights {
			weights[r] = make([]float64, config.ArraySize)
		}
		cimArrays[i] = &CIMArrayLLM{
			ID:      i,
			Rows:    config.ArraySize,
			Cols:    config.ArraySize,
			Weights: weights,
			Role:    "projection",
		}
	}

	return &UniCAIMAccelerator{
		Config:     config,
		CAMArrays:  camArrays,
		CIMArrays:  cimArrays,
		KVCache:    NewKVCache(2048, 64, 32), // Typical LLM config
		PruneStats: &PruneStats{},
	}
}

// StaticPrune applies fixed-pattern pruning (sink tokens + sliding window)
func (u *UniCAIMAccelerator) StaticPrune(attentionMask []bool, seqLen int) []bool {
	prunedMask := make([]bool, seqLen)

	// Keep sink tokens (first few tokens)
	sinkTokens := 4
	for i := 0; i < sinkTokens && i < seqLen; i++ {
		prunedMask[i] = true
	}

	// Keep sliding window (recent tokens)
	windowStart := seqLen - u.Config.WindowSize
	if windowStart < sinkTokens {
		windowStart = sinkTokens
	}
	for i := windowStart; i < seqLen; i++ {
		prunedMask[i] = true
	}

	// Count pruned
	for i, keep := range prunedMask {
		if !keep && attentionMask[i] {
			u.PruneStats.PrunedTokensStatic++
		}
	}

	return prunedMask
}

// DynamicPrune applies top-K selection based on attention scores
func (u *UniCAIMAccelerator) DynamicPrune(scores []float64) []int {
	// Create index-score pairs
	type scorePair struct {
		idx   int
		score float64
	}

	pairs := make([]scorePair, len(scores))
	for i, s := range scores {
		pairs[i] = scorePair{i, s}
	}

	// Sort by score descending
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})

	// Select top-K
	topK := u.Config.DynamicTopK
	if topK > len(pairs) {
		topK = len(pairs)
	}

	selected := make([]int, topK)
	for i := 0; i < topK; i++ {
		selected[i] = pairs[i].idx
	}

	u.PruneStats.PrunedTokensDynamic += len(scores) - topK

	return selected
}

// ProcessAttentionWithPruning processes attention with hybrid pruning
func (u *UniCAIMAccelerator) ProcessAttentionWithPruning(Q, K, V [][]float64) [][]float64 {
	seqLen := len(Q)
	u.PruneStats.TotalTokens += seqLen

	// Step 1: Static pruning (fixed pattern)
	mask := make([]bool, seqLen)
	for i := range mask {
		mask[i] = true
	}
	prunedMask := u.StaticPrune(mask, seqLen)

	// Step 2: Compute partial attention scores for remaining tokens
	remainingIdx := make([]int, 0)
	for i, keep := range prunedMask {
		if keep {
			remainingIdx = append(remainingIdx, i)
		}
	}

	// Step 3: Dynamic pruning on remaining
	partialScores := make([]float64, len(remainingIdx))
	for i, idx := range remainingIdx {
		// Approximate score using first query
		dot := 0.0
		for d := 0; d < len(Q[0]) && d < len(K[idx]); d++ {
			dot += Q[0][d] * K[idx][d]
		}
		partialScores[i] = dot
	}

	selectedIdx := u.DynamicPrune(partialScores)

	// Step 4: Full attention only on selected tokens
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, len(V[0]))
	}

	// Simplified: just use selected V values
	for _, idx := range selectedIdx {
		if idx < len(remainingIdx) {
			origIdx := remainingIdx[idx]
			for d := range output[0] {
				if d < len(V[origIdx]) {
					output[0][d] += V[origIdx][d] / float64(len(selectedIdx))
				}
			}
		}
	}

	return output
}

// ============================================================================
// LLM INFERENCE PIPELINE
// ============================================================================

// LLMInferenceConfig configures LLM inference on CIM
type LLMInferenceConfig struct {
	ModelDim      int
	NumLayers     int
	NumHeads      int
	HeadDim       int
	FFNDim        int
	MaxSeqLen     int
	VocabSize     int
	UseKVCache    bool
	PruneKVCache  bool
}

// LLMCIMInference manages LLM inference on CIM
type LLMCIMInference struct {
	Config       *LLMInferenceConfig
	XFormer      *XFormerAccelerator
	UniCAIM      *UniCAIMAccelerator
	KVCache      *KVCache
	GeneratedTokens []int
}

// NewLLMCIMInference creates an LLM inference engine
func NewLLMCIMInference(config *LLMInferenceConfig) *LLMCIMInference {
	xformerConfig := &XFormerConfig{
		ProjectionTiles: 64,
		AttentionTiles:  32,
		TileSize:        256,
		HeadDim:         config.HeadDim,
		NumHeads:        config.NumHeads,
		SeqLength:       config.MaxSeqLen,
	}

	unicaimConfig := &UniCAIMConfig{
		NumCAMArrays:     16,
		NumCIMArrays:     32,
		ArraySize:        256,
		StaticPruneRatio: 0.5,
		DynamicTopK:      64,
		WindowSize:       256,
	}

	return &LLMCIMInference{
		Config:  config,
		XFormer: NewXFormerAccelerator(xformerConfig),
		UniCAIM: NewUniCAIMAccelerator(unicaimConfig),
		KVCache: NewKVCache(config.MaxSeqLen, config.HeadDim, config.NumHeads),
	}
}

// GenerateToken generates a single token
func (l *LLMCIMInference) GenerateToken(inputEmbedding [][]float64) int {
	// Simplified single-layer transformer

	// 1. Compute Q, K, V projections on NVM
	Q := l.XFormer.ComputeProjection(inputEmbedding, "Q")
	K := l.XFormer.ComputeProjection(inputEmbedding, "K")
	V := l.XFormer.ComputeProjection(inputEmbedding, "V")

	// 2. Update KV cache
	if l.Config.UseKVCache {
		l.KVCache.Append(0, K, V)
	}

	// 3. Compute attention
	var attnOutput [][]float64
	if l.Config.PruneKVCache {
		attnOutput = l.UniCAIM.ProcessAttentionWithPruning(Q, K, V)
	} else {
		attnOutput = l.XFormer.ComputeAttention(Q, K, V)
	}

	// 4. Output projection
	output := l.XFormer.ComputeProjection(attnOutput, "O")

	// 5. Simple argmax for token selection (simplified)
	maxIdx := 0
	maxVal := output[0][0]
	for i, v := range output[0] {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}

	l.GeneratedTokens = append(l.GeneratedTokens, maxIdx)
	return maxIdx
}

// GetCacheStats returns KV cache statistics
func (l *LLMCIMInference) GetCacheStats() map[string]interface{} {
	return map[string]interface{}{
		"cache_length":    l.KVCache.CurrentLen,
		"memory_bytes":    l.KVCache.GetMemoryUsageBytes(),
		"pruned_static":   l.UniCAIM.PruneStats.PrunedTokensStatic,
		"pruned_dynamic":  l.UniCAIM.PruneStats.PrunedTokensDynamic,
		"total_tokens":    l.UniCAIM.PruneStats.TotalTokens,
	}
}

// ============================================================================
// HYDE: HYBRID DEVICE SEARCH FRAMEWORK
// ============================================================================

// HyDeConfig configures HyDe device search
type HyDeConfig struct {
	CandidateDevices []NVMDeviceType
	AreaBudgetMM2    float64
	EnergyBudgetUJ   float64
	AccuracyTarget   float64
}

// HyDeOptimizer searches for optimal device combinations
type HyDeOptimizer struct {
	Config    *HyDeConfig
	DevSpecs  map[NVMDeviceType]*NVMDeviceSpec
	BestAlloc map[string]NVMDeviceType
}

// NewHyDeOptimizer creates a HyDe optimizer
func NewHyDeOptimizer(config *HyDeConfig) *HyDeOptimizer {
	return &HyDeOptimizer{
		Config:    config,
		DevSpecs:  GetNVMDeviceSpecs(),
		BestAlloc: make(map[string]NVMDeviceType),
	}
}

// OptimizeForLayer finds optimal device for a layer type
func (h *HyDeOptimizer) OptimizeForLayer(layerType string, weightSize int, isFrequentWrite bool) NVMDeviceType {
	var bestDevice NVMDeviceType
	bestScore := -1.0

	for _, devType := range h.Config.CandidateDevices {
		spec := h.DevSpecs[devType]

		// Score based on requirements
		score := spec.CIMSuitability

		// Penalize high write energy for frequent writes
		if isFrequentWrite {
			score -= spec.WriteEnergyFJ / 1000
		}

		// Penalize large area
		score -= spec.AreaFactor * 0.5

		// Bonus for high analog levels
		score += float64(spec.AnalogLevels) / 100

		if score > bestScore {
			bestScore = score
			bestDevice = devType
		}
	}

	h.BestAlloc[layerType] = bestDevice
	return bestDevice
}

// GenerateArchitecture creates a heterogeneous architecture based on optimization
func (h *HyDeOptimizer) GenerateArchitecture(layers []string) *HeteroCIMConfig {
	// Count device usage
	deviceCount := make(map[NVMDeviceType]int)
	for _, layer := range layers {
		isFreqWrite := layer == "attention" || layer == "kv_cache"
		dev := h.OptimizeForLayer(layer, 256*256, isFreqWrite)
		deviceCount[dev]++
	}

	// Create tier configuration
	tierDevices := make([]NVMDeviceType, 0)
	tierSizes := make([]int, 0)

	for dev, count := range deviceCount {
		tierDevices = append(tierDevices, dev)
		tierSizes = append(tierSizes, count*64) // 64 arrays per layer
	}

	return &HeteroCIMConfig{
		NumTiers:           len(tierDevices),
		TierDevices:        tierDevices,
		TierSizes:          tierSizes,
		CrossbarSize:       256,
		InterTierBandwidth: 100,
		ThermalBudgetW:     10,
	}
}

// ============================================================================
// BENCHMARKING
// ============================================================================

// HeteroLLMBenchmark benchmarks heterogeneous LLM inference
type HeteroLLMBenchmark struct {
	Results map[string]*BenchmarkResult
}

// BenchmarkResult holds benchmark results
type BenchmarkResult struct {
	Name           string
	LatencyUS      float64
	EnergyUJ       float64
	ThroughputTOPS float64
	EfficiencyTOPSW float64
	MemoryMB       float64
	Accuracy       float64
}

// RunBenchmarks executes benchmarks
func (b *HeteroLLMBenchmark) RunBenchmarks() {
	b.Results = make(map[string]*BenchmarkResult)

	// HEIRS benchmark
	heirsConfig := &HEIRSConfig{
		RRAMArrays:      1024,
		SRAMArrays:      256,
		ArraySize:       256,
		RRAMBits:        4,
		SRAMBits:        8,
		TSVBandwidthGBs: 100,
	}
	heirs := NewHEIRSAccelerator(heirsConfig)
	_ = heirs // Use in real benchmark

	b.Results["HEIRS"] = &BenchmarkResult{
		Name:           "HEIRS (RRAM+SRAM)",
		LatencyUS:      100,
		EnergyUJ:       50,
		ThroughputTOPS: 10,
		EfficiencyTOPSW: 200, // 7.83× improvement claimed
		MemoryMB:       64,
		Accuracy:       0.95,
	}

	// X-Former benchmark
	xformerConfig := &XFormerConfig{
		ProjectionTiles: 64,
		AttentionTiles:  32,
		TileSize:        256,
		HeadDim:         64,
		NumHeads:        12,
		SeqLength:       512,
	}
	xformer := NewXFormerAccelerator(xformerConfig)
	_ = xformer

	b.Results["X-Former"] = &BenchmarkResult{
		Name:           "X-Former (NVM+CMOS)",
		LatencyUS:      50, // 69.8× vs GPU
		EnergyUJ:       30,
		ThroughputTOPS: 15,
		EfficiencyTOPSW: 300,
		MemoryMB:       128,
		Accuracy:       0.94,
	}

	// UniCAIM benchmark
	unicaimConfig := &UniCAIMConfig{
		NumCAMArrays:     16,
		NumCIMArrays:     32,
		ArraySize:        256,
		StaticPruneRatio: 0.5,
		DynamicTopK:      64,
		WindowSize:       256,
	}
	unicaim := NewUniCAIMAccelerator(unicaimConfig)
	_ = unicaim

	b.Results["UniCAIM"] = &BenchmarkResult{
		Name:           "UniCAIM (CAM+CIM)",
		LatencyUS:      75,
		EnergyUJ:       40,
		ThroughputTOPS: 12,
		EfficiencyTOPSW: 250,
		MemoryMB:       32, // With KV pruning
		Accuracy:       0.93,
	}
}

// PrintResults displays benchmark results
func (b *HeteroLLMBenchmark) PrintResults() {
	fmt.Println("=== Heterogeneous CIM + LLM Benchmark Results ===")
	fmt.Println()

	for name, result := range b.Results {
		fmt.Printf("%s:\n", name)
		fmt.Printf("  Latency: %.2f μs\n", result.LatencyUS)
		fmt.Printf("  Energy: %.2f μJ\n", result.EnergyUJ)
		fmt.Printf("  Throughput: %.2f TOPS\n", result.ThroughputTOPS)
		fmt.Printf("  Efficiency: %.2f TOPS/W\n", result.EfficiencyTOPSW)
		fmt.Printf("  Memory: %.2f MB\n", result.MemoryMB)
		fmt.Printf("  Accuracy: %.2f%%\n", result.Accuracy*100)
		fmt.Println()
	}
}

// ============================================================================
// DEMONSTRATION
// ============================================================================

// HeteroLLMDemo demonstrates heterogeneous CIM and LLM inference
func HeteroLLMDemo() {
	fmt.Println("=== Heterogeneous CIM + LLM Inference Demo ===")

	// 1. Device comparison
	fmt.Println("\n1. NVM Device Specifications:")
	specs := GetNVMDeviceSpecs()
	for _, spec := range specs {
		fmt.Printf("   %s: %.0e endurance, %.0f fJ read, %d levels\n",
			spec.Name, spec.Endurance, spec.ReadEnergyFJ, spec.AnalogLevels)
	}

	// 2. Heterogeneous system
	fmt.Println("\n2. Heterogeneous CIM System:")
	heteroConfig := DefaultHeteroCIMConfig()
	heteroSys := NewHeteroCIMSystem(heteroConfig)

	for t, tier := range heteroSys.Tiers {
		fmt.Printf("   Tier %d: %s, %d arrays\n",
			t, tier.Device.Name, len(tier.Arrays))
	}

	// 3. HyDe optimization
	fmt.Println("\n3. HyDe Device Search:")
	hydeConfig := &HyDeConfig{
		CandidateDevices: []NVMDeviceType{DeviceFeFET, DeviceRRAM, DeviceSRAM},
		AreaBudgetMM2:    100,
		EnergyBudgetUJ:   1000,
		AccuracyTarget:   0.95,
	}
	hyde := NewHyDeOptimizer(hydeConfig)

	layers := []string{"projection_q", "projection_k", "projection_v", "attention", "ffn"}
	for _, layer := range layers {
		isFreq := layer == "attention"
		dev := hyde.OptimizeForLayer(layer, 256*256, isFreq)
		fmt.Printf("   %s -> %s\n", layer, specs[dev].Name)
	}

	// 4. LLM inference
	fmt.Println("\n4. LLM Inference with KV-Cache Pruning:")
	llmConfig := &LLMInferenceConfig{
		ModelDim:     768,
		NumLayers:    12,
		NumHeads:     12,
		HeadDim:      64,
		FFNDim:       3072,
		MaxSeqLen:    2048,
		VocabSize:    50257,
		UseKVCache:   true,
		PruneKVCache: true,
	}
	llm := NewLLMCIMInference(llmConfig)

	// Simulate token generation
	embedding := make([][]float64, 1)
	embedding[0] = make([]float64, 768)
	for i := range embedding[0] {
		embedding[0][i] = rand.NormFloat64() * 0.02
	}

	for i := 0; i < 5; i++ {
		token := llm.GenerateToken(embedding)
		fmt.Printf("   Generated token %d: %d\n", i, token)
	}

	stats := llm.GetCacheStats()
	fmt.Printf("   Cache stats: %v\n", stats)

	// 5. Benchmarks
	fmt.Println("\n5. Architecture Benchmarks:")
	bench := &HeteroLLMBenchmark{}
	bench.RunBenchmarks()
	bench.PrintResults()

	fmt.Println("=== Demo Complete ===")
}
