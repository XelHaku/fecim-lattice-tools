// transformer_adc_cim.go - Transformer/LLM CIM Acceleration and ADC Optimization
// IronLattice Visualization Project - Iteration 123
//
// This module implements simulation models for:
// 1. Transformer/LLM CIM acceleration architectures
// 2. Attention mechanism mapping to crossbar arrays
// 3. KV-cache optimization for in-memory computing
// 4. ADC optimization techniques (adaptive, sparsity-aware, ADC-less)
// 5. In-memory ADC (IMADC) with dual functionality
// 6. Resonant time-domain CIM architecture
//
// Research basis:
// - arXiv 2406.08413: Memory Is All You Need (CIM for LLM survey)
// - HARDSEA: 28.5× acceleration, 1,894.3× energy efficiency vs RTX 3090
// - HALO: 18× speedup for LLaMA-2 7B with HBM CiD + analog CiM
// - Nature Communications 2025: Memristor-based adaptive ADC
// - Wiley 2025: Multifunctional In-Memory ADC (IMADC)
// - GLSVLSI 2024: Resonant Time-Domain CiM (28.05 TOPS/W)

package layers

import (
	"encoding/json"
	"math"
	"math/rand"
	"sync"
)

// ============================================================================
// TRANSFORMER CIM ACCELERATION
// ============================================================================

// TransformerCIMConfig defines configuration for transformer acceleration
type TransformerCIMConfig struct {
	// Model dimensions
	HiddenSize     int `json:"hidden_size"`      // d_model (e.g., 768, 1024, 4096)
	NumHeads       int `json:"num_heads"`        // Number of attention heads
	HeadDim        int `json:"head_dim"`         // Dimension per head
	FFNHiddenSize  int `json:"ffn_hidden_size"`  // FFN intermediate size (4x hidden typically)
	NumLayers      int `json:"num_layers"`       // Number of transformer layers
	VocabSize      int `json:"vocab_size"`       // Vocabulary size
	MaxSeqLen      int `json:"max_seq_len"`      // Maximum sequence length

	// CIM array configuration
	ArrayRows      int     `json:"array_rows"`       // Crossbar rows (64-256 typical)
	ArrayCols      int     `json:"array_cols"`       // Crossbar columns
	WeightBits     int     `json:"weight_bits"`      // Weight precision (4-8 bits)
	ActivationBits int     `json:"activation_bits"`  // Activation precision

	// ADC configuration
	ADCBits        int     `json:"adc_bits"`         // ADC resolution (4-8 bits)
	ADCsPerArray   int     `json:"adcs_per_array"`   // Number of ADCs per crossbar
	UseAdaptiveADC bool    `json:"use_adaptive_adc"` // Enable adaptive quantization

	// Sparsity settings
	AttentionSparsity float64 `json:"attention_sparsity"` // Attention sparsity ratio
	WeightSparsity    float64 `json:"weight_sparsity"`    // Weight matrix sparsity
}

// KVCache represents key-value cache for autoregressive generation
type KVCache struct {
	Keys       [][][]float64 `json:"keys"`        // [layer][seq_len][head_dim]
	Values     [][][]float64 `json:"values"`      // [layer][seq_len][head_dim]
	CacheLen   int           `json:"cache_len"`   // Current cached length
	MaxLen     int           `json:"max_len"`     // Maximum cache length
	NumLayers  int           `json:"num_layers"`
	NumHeads   int           `json:"num_heads"`
	HeadDim    int           `json:"head_dim"`
	mu         sync.RWMutex
}

// NewKVCache creates a new KV cache
func NewKVCache(numLayers, numHeads, headDim, maxLen int) *KVCache {
	keys := make([][][]float64, numLayers)
	values := make([][][]float64, numLayers)

	for l := 0; l < numLayers; l++ {
		keys[l] = make([][]float64, 0, maxLen)
		values[l] = make([][]float64, 0, maxLen)
	}

	return &KVCache{
		Keys:      keys,
		Values:    values,
		CacheLen:  0,
		MaxLen:    maxLen,
		NumLayers: numLayers,
		NumHeads:  numHeads,
		HeadDim:   headDim,
	}
}

// Append adds new key-value pairs to cache
func (kv *KVCache) Append(layer int, key, value []float64) {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	if kv.CacheLen < kv.MaxLen {
		kv.Keys[layer] = append(kv.Keys[layer], key)
		kv.Values[layer] = append(kv.Values[layer], value)
		if layer == 0 {
			kv.CacheLen++
		}
	}
}

// GetMemoryFootprint returns KV cache memory in bytes
func (kv *KVCache) GetMemoryFootprint(bytesPerElement int) int64 {
	// 2 (K+V) * layers * seq_len * num_heads * head_dim * bytes
	return int64(2 * kv.NumLayers * kv.CacheLen * kv.NumHeads * kv.HeadDim * bytesPerElement)
}

// ============================================================================
// HARDSEA-STYLE HYBRID ATTENTION ACCELERATOR
// ============================================================================

// HARDSEAConfig defines HARDSEA accelerator parameters
type HARDSEAConfig struct {
	// Performance targets (from paper)
	// 28.5× acceleration vs RTX 3090
	// 1,894.3× energy efficiency
	// 921.6 GOPs throughput
	// 943.7 GOPs/W efficiency

	NumPEs           int     `json:"num_pes"`           // Processing elements
	SRAMCIMSize      int     `json:"sram_cim_size"`     // SRAM-CIM capacity (KB)
	SparsityThreshold float64 `json:"sparsity_threshold"` // Attention sparsity threshold

	// Hybrid analog-digital
	AnalogPrecision  int     `json:"analog_precision"`  // Analog compute bits
	DigitalPrecision int     `json:"digital_precision"` // Digital compute bits
}

// HARDSEA implements hybrid analog-digital sparse attention accelerator
type HARDSEA struct {
	Config       *HARDSEAConfig
	SRAMArrays   []*SRAMCIMArray
	SparseEngine *SparseAttentionEngine
	Stats        *HARDSEAStats
	mu           sync.RWMutex
}

// HARDSEAStats tracks performance statistics
type HARDSEAStats struct {
	TotalOps         int64   `json:"total_ops"`
	SparseOps        int64   `json:"sparse_ops"`
	DenseOps         int64   `json:"dense_ops"`
	EnergyConsumed   float64 `json:"energy_consumed"`   // pJ
	CyclesExecuted   int64   `json:"cycles_executed"`
	SparsityAchieved float64 `json:"sparsity_achieved"`
}

// SRAMCIMArray represents SRAM-based CIM array for attention
type SRAMCIMArray struct {
	Rows       int
	Cols       int
	Data       [][]float64
	BitLines   []float64
	WordLines  []bool
	mu         sync.RWMutex
}

// NewSRAMCIMArray creates a new SRAM-CIM array
func NewSRAMCIMArray(rows, cols int) *SRAMCIMArray {
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
	}

	return &SRAMCIMArray{
		Rows:      rows,
		Cols:      cols,
		Data:      data,
		BitLines:  make([]float64, cols),
		WordLines: make([]bool, rows),
	}
}

// ComputeMAC performs in-SRAM MAC operation
func (s *SRAMCIMArray) ComputeMAC(input []float64) []float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	output := make([]float64, s.Cols)

	// Activate word lines based on input
	for i := range s.WordLines {
		if i < len(input) {
			s.WordLines[i] = input[i] > 0.5
		}
	}

	// Accumulate on bit lines
	for j := 0; j < s.Cols; j++ {
		sum := 0.0
		for i := 0; i < s.Rows && i < len(input); i++ {
			sum += input[i] * s.Data[i][j]
		}
		output[j] = sum
	}

	return output
}

// SparseAttentionEngine handles sparse attention computation
type SparseAttentionEngine struct {
	SparsityThreshold float64
	BlockSize         int
	mu                sync.RWMutex
}

// ComputeSparseAttention computes attention with sparsity
func (e *SparseAttentionEngine) ComputeSparseAttention(Q, K, V [][]float64) ([][]float64, float64) {
	e.mu.Lock()
	defer e.mu.Unlock()

	seqLen := len(Q)
	headDim := len(Q[0])

	// Compute attention scores
	scores := make([][]float64, seqLen)
	for i := range scores {
		scores[i] = make([]float64, seqLen)
		for j := range scores[i] {
			// Q[i] · K[j]^T / sqrt(d)
			dot := 0.0
			for k := 0; k < headDim; k++ {
				dot += Q[i][k] * K[j][k]
			}
			scores[i][j] = dot / math.Sqrt(float64(headDim))
		}
	}

	// Apply sparsity mask (top-k or threshold)
	totalElements := seqLen * seqLen
	sparseElements := 0

	for i := range scores {
		// Find threshold for this row
		rowMax := scores[i][0]
		for j := range scores[i] {
			if scores[i][j] > rowMax {
				rowMax = scores[i][j]
			}
		}
		threshold := rowMax - e.SparsityThreshold*math.Abs(rowMax)

		// Apply threshold
		for j := range scores[i] {
			if scores[i][j] < threshold {
				scores[i][j] = 0
				sparseElements++
			}
		}
	}

	sparsity := float64(sparseElements) / float64(totalElements)

	// Softmax per row
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

	// Compute output: Attention × V
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, headDim)
		for j := 0; j < seqLen; j++ {
			if scores[i][j] > 0 { // Skip zero attention weights
				for k := 0; k < headDim; k++ {
					output[i][k] += scores[i][j] * V[j][k]
				}
			}
		}
	}

	return output, sparsity
}

// NewHARDSEA creates a new HARDSEA accelerator
func NewHARDSEA(config *HARDSEAConfig) *HARDSEA {
	arrays := make([]*SRAMCIMArray, config.NumPEs)
	for i := range arrays {
		arrays[i] = NewSRAMCIMArray(256, 256) // Standard size
	}

	return &HARDSEA{
		Config:     config,
		SRAMArrays: arrays,
		SparseEngine: &SparseAttentionEngine{
			SparsityThreshold: config.SparsityThreshold,
			BlockSize:         64,
		},
		Stats: &HARDSEAStats{},
	}
}

// ============================================================================
// HALO ARCHITECTURE (HBM CiD + Analog CiM)
// ============================================================================

// HALOConfig defines HALO heterogeneous accelerator parameters
type HALOConfig struct {
	// HBM-based Compute-in-DRAM
	HBMCapacity      int     `json:"hbm_capacity"`      // GB
	HBMBandwidth     float64 `json:"hbm_bandwidth"`     // GB/s
	CiDThroughput    float64 `json:"cid_throughput"`    // TOPS

	// On-chip analog CiM
	CiMArrays        int     `json:"cim_arrays"`        // Number of CiM arrays
	CiMArraySize     int     `json:"cim_array_size"`    // Rows per array
	CiMPrecision     int     `json:"cim_precision"`     // Bits

	// 2.5D integration
	InterleaverDepth int     `json:"interleaver_depth"` // Pipeline depth
}

// HALO implements heterogeneous memory-centric accelerator
type HALO struct {
	Config        *HALOConfig
	HBMController *HBMCiDController
	CiMEngine     *AnalogCiMEngine
	Scheduler     *HALOScheduler
	Stats         *HALOStats
	mu            sync.RWMutex
}

// HALOStats tracks HALO performance
type HALOStats struct {
	PrefillLatency  float64 `json:"prefill_latency"`  // ms
	DecodeLatency   float64 `json:"decode_latency"`   // ms per token
	TotalTokens     int64   `json:"total_tokens"`
	EnergyPerToken  float64 `json:"energy_per_token"` // mJ
}

// HBMCiDController manages HBM-based compute-in-DRAM
type HBMCiDController struct {
	Capacity    int     // GB
	Bandwidth   float64 // GB/s
	CiDUnits    []*CiDUnit
	mu          sync.RWMutex
}

// CiDUnit represents a compute-in-DRAM unit
type CiDUnit struct {
	BankID      int
	RowBuffer   []float64
	Accumulator float64
}

// AnalogCiMEngine manages on-chip analog CiM
type AnalogCiMEngine struct {
	Arrays       []*AnalogCiMArray
	ADCs         []*AdaptiveADC
	mu           sync.RWMutex
}

// AnalogCiMArray represents analog CiM crossbar
type AnalogCiMArray struct {
	Rows        int
	Cols        int
	Conductances [][]float64
	mu          sync.RWMutex
}

// HALOScheduler manages workload distribution
type HALOScheduler struct {
	PrefillQueue [][]float64
	DecodeQueue  [][]float64
	mu           sync.RWMutex
}

// NewHALO creates a new HALO accelerator
func NewHALO(config *HALOConfig) *HALO {
	// Initialize HBM CiD
	cidUnits := make([]*CiDUnit, 8) // 8 HBM stacks typical
	for i := range cidUnits {
		cidUnits[i] = &CiDUnit{
			BankID:    i,
			RowBuffer: make([]float64, 1024),
		}
	}

	hbmController := &HBMCiDController{
		Capacity:  config.HBMCapacity,
		Bandwidth: config.HBMBandwidth,
		CiDUnits:  cidUnits,
	}

	// Initialize analog CiM
	arrays := make([]*AnalogCiMArray, config.CiMArrays)
	adcs := make([]*AdaptiveADC, config.CiMArrays)
	for i := range arrays {
		conductances := make([][]float64, config.CiMArraySize)
		for j := range conductances {
			conductances[j] = make([]float64, config.CiMArraySize)
		}
		arrays[i] = &AnalogCiMArray{
			Rows:         config.CiMArraySize,
			Cols:         config.CiMArraySize,
			Conductances: conductances,
		}
		adcs[i] = NewAdaptiveADC(&AdaptiveADCConfig{
			BaseBits:     6,
			MaxBits:      8,
			AdaptiveMode: true,
		})
	}

	cimEngine := &AnalogCiMEngine{
		Arrays: arrays,
		ADCs:   adcs,
	}

	return &HALO{
		Config:        config,
		HBMController: hbmController,
		CiMEngine:     cimEngine,
		Scheduler:     &HALOScheduler{},
		Stats:         &HALOStats{},
	}
}

// ProcessPrefill handles prefill phase (compute-bound)
func (h *HALO) ProcessPrefill(tokens []int, embeddings [][]float64) [][]float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	seqLen := len(tokens)
	hiddenSize := len(embeddings[0])

	// Prefill uses analog CiM for parallel MVM
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, hiddenSize)
		// Parallel computation across CiM arrays
		for j := 0; j < hiddenSize; j++ {
			sum := 0.0
			for k := 0; k < len(embeddings[i]); k++ {
				// Use CiM for dot product
				arrayIdx := j % len(h.CiMEngine.Arrays)
				if k < h.CiMEngine.Arrays[arrayIdx].Rows {
					sum += embeddings[i][k] * h.CiMEngine.Arrays[arrayIdx].Conductances[k][j%h.CiMEngine.Arrays[arrayIdx].Cols]
				}
			}
			output[i][j] = sum
		}
	}

	h.Stats.PrefillLatency = float64(seqLen) * 0.1 // Simplified: 0.1ms per token

	return output
}

// ProcessDecode handles decode phase (memory-bound)
func (h *HALO) ProcessDecode(kvCache *KVCache, query []float64) []float64 {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Decode uses HBM CiD for KV cache access
	// Single token generation is memory-bound

	output := make([]float64, len(query))

	// Simulate HBM access for KV cache
	for i := range h.HBMController.CiDUnits {
		unit := h.HBMController.CiDUnits[i]
		// CiD performs MAC in DRAM
		for j := range unit.RowBuffer {
			if j < len(query) {
				unit.Accumulator += unit.RowBuffer[j] * query[j]
			}
		}
	}

	h.Stats.DecodeLatency = 1.0 // Simplified: 1ms per token
	h.Stats.TotalTokens++

	return output
}

// ============================================================================
// ADC OPTIMIZATION
// ============================================================================

// AdaptiveADCConfig defines adaptive ADC parameters
type AdaptiveADCConfig struct {
	BaseBits       int     `json:"base_bits"`       // Base resolution (e.g., 6)
	MaxBits        int     `json:"max_bits"`        // Maximum resolution (e.g., 8)
	AdaptiveMode   bool    `json:"adaptive_mode"`   // Enable adaptive quantization
	SparsityAware  bool    `json:"sparsity_aware"`  // Exploit input sparsity
	EnergyPerBit   float64 `json:"energy_per_bit"`  // fJ per bit of resolution
}

// AdaptiveADC implements memristor-based adaptive ADC
// From Nature Communications 2025: 15.1× energy efficiency, 12.9× area reduction
type AdaptiveADC struct {
	Config            *AdaptiveADCConfig
	Thresholds        []float64         // Adaptive quantization thresholds
	CAMCells          []*CAMCell        // Content-addressable memory for threshold
	Stats             *ADCStats
	mu                sync.RWMutex
}

// CAMCell represents analog CAM cell for threshold comparison
type CAMCell struct {
	Threshold     float64
	OverlapMargin float64 // Programmable overlap boundary
}

// ADCStats tracks ADC performance
type ADCStats struct {
	Conversions      int64   `json:"conversions"`
	AverageBits      float64 `json:"average_bits"`
	TotalEnergy      float64 `json:"total_energy"`      // fJ
	SparsitySkipped  int64   `json:"sparsity_skipped"`  // Conversions skipped due to sparsity
}

// NewAdaptiveADC creates a new adaptive ADC
func NewAdaptiveADC(config *AdaptiveADCConfig) *AdaptiveADC {
	numLevels := 1 << config.MaxBits
	thresholds := make([]float64, numLevels)
	camCells := make([]*CAMCell, numLevels)

	// Initialize uniform thresholds (will be adapted)
	for i := range thresholds {
		thresholds[i] = float64(i) / float64(numLevels-1)
		camCells[i] = &CAMCell{
			Threshold:     thresholds[i],
			OverlapMargin: 0.01, // 1% overlap
		}
	}

	return &AdaptiveADC{
		Config:     config,
		Thresholds: thresholds,
		CAMCells:   camCells,
		Stats:      &ADCStats{},
	}
}

// AdaptThresholds optimizes thresholds for output distribution
func (a *AdaptiveADC) AdaptThresholds(distribution []float64) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if len(distribution) == 0 {
		return
	}

	// Compute histogram of distribution
	numBins := len(a.Thresholds)
	histogram := make([]int, numBins)

	minVal, maxVal := distribution[0], distribution[0]
	for _, v := range distribution {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}

	rangeVal := maxVal - minVal
	if rangeVal == 0 {
		rangeVal = 1.0
	}

	for _, v := range distribution {
		bin := int((v - minVal) / rangeVal * float64(numBins-1))
		if bin >= numBins {
			bin = numBins - 1
		}
		histogram[bin]++
	}

	// Compute cumulative distribution for optimal quantization
	cumsum := make([]int, numBins)
	cumsum[0] = histogram[0]
	for i := 1; i < numBins; i++ {
		cumsum[i] = cumsum[i-1] + histogram[i]
	}

	total := cumsum[numBins-1]
	if total == 0 {
		total = 1
	}

	// Set thresholds at equal-probability intervals (Lloyd-Max style)
	targetPerBin := total / numBins
	currentBin := 0
	for i := range a.Thresholds {
		targetCumsum := (i + 1) * targetPerBin
		for currentBin < numBins-1 && cumsum[currentBin] < targetCumsum {
			currentBin++
		}
		a.Thresholds[i] = minVal + float64(currentBin)/float64(numBins-1)*rangeVal
		a.CAMCells[i].Threshold = a.Thresholds[i]
	}
}

// Convert performs adaptive ADC conversion
func (a *AdaptiveADC) Convert(analogValue float64, inputSparsity float64) (int, int) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.Stats.Conversions++

	// Sparsity-aware bit reduction
	effectiveBits := a.Config.MaxBits
	if a.Config.SparsityAware && inputSparsity > 0 {
		// If input is sparse, output range is limited
		// Can use fewer MSBs
		maxPossibleOutput := 1.0 - inputSparsity
		bitsNeeded := int(math.Ceil(math.Log2(maxPossibleOutput*float64(1<<a.Config.MaxBits) + 1)))
		if bitsNeeded < a.Config.BaseBits {
			bitsNeeded = a.Config.BaseBits
		}
		effectiveBits = bitsNeeded
	}

	// Binary search using CAM cells
	low, high := 0, (1<<effectiveBits)-1
	result := 0

	for low <= high {
		mid := (low + high) / 2
		thresholdIdx := mid * (len(a.Thresholds) - 1) / ((1 << effectiveBits) - 1)
		if thresholdIdx >= len(a.Thresholds) {
			thresholdIdx = len(a.Thresholds) - 1
		}

		if analogValue >= a.CAMCells[thresholdIdx].Threshold {
			result = mid
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	// Update statistics
	a.Stats.AverageBits = (a.Stats.AverageBits*float64(a.Stats.Conversions-1) + float64(effectiveBits)) / float64(a.Stats.Conversions)
	a.Stats.TotalEnergy += float64(effectiveBits) * a.Config.EnergyPerBit

	return result, effectiveBits
}

// ============================================================================
// IN-MEMORY ADC (IMADC)
// ============================================================================

// IMADCConfig defines in-memory ADC parameters
// From Wiley 2025: 45 μm², 29.6 fJ/op
type IMADCConfig struct {
	NumLevels       int     `json:"num_levels"`        // Quantization levels
	AreaUM2         float64 `json:"area_um2"`          // Area in μm² (45 typical)
	EnergyFJ        float64 `json:"energy_fj"`         // Energy per conversion (29.6 fJ)
	NonlinearActiv  bool    `json:"nonlinear_activ"`   // Dual function: ADC + activation
	ActivationType  string  `json:"activation_type"`   // "relu", "sigmoid", "tanh"
}

// IMADC implements in-memory ADC with dual functionality
type IMADC struct {
	Config         *IMADCConfig
	FlashTFTs      []*ChargeTrapTFT
	ThermometerOut []bool
	Stats          *IMADCStats
	mu             sync.RWMutex
}

// ChargeTrapTFT represents charge-trap flash TFT for IMADC
type ChargeTrapTFT struct {
	Threshold   float64 // Programmable threshold voltage
	ChargeState float64 // Trapped charge level
}

// IMADCStats tracks IMADC performance
type IMADCStats struct {
	Conversions   int64   `json:"conversions"`
	TotalEnergy   float64 `json:"total_energy"`   // fJ
	Activations   int64   `json:"activations"`    // Activation function applications
}

// NewIMADC creates a new in-memory ADC
func NewIMADC(config *IMADCConfig) *IMADC {
	if config.AreaUM2 == 0 {
		config.AreaUM2 = 45.0 // From paper
	}
	if config.EnergyFJ == 0 {
		config.EnergyFJ = 29.6 // From paper
	}

	tfts := make([]*ChargeTrapTFT, config.NumLevels)
	for i := range tfts {
		tfts[i] = &ChargeTrapTFT{
			Threshold:   float64(i+1) / float64(config.NumLevels),
			ChargeState: 0,
		}
	}

	return &IMADC{
		Config:         config,
		FlashTFTs:      tfts,
		ThermometerOut: make([]bool, config.NumLevels),
		Stats:          &IMADCStats{},
	}
}

// ConvertWithActivation performs ADC + optional nonlinear activation
func (im *IMADC) ConvertWithActivation(analogValue float64) (int, float64) {
	im.mu.Lock()
	defer im.mu.Unlock()

	im.Stats.Conversions++
	im.Stats.TotalEnergy += im.Config.EnergyFJ

	// Thermometer code conversion
	digitalValue := 0
	for i, tft := range im.FlashTFTs {
		if analogValue >= tft.Threshold {
			im.ThermometerOut[i] = true
			digitalValue = i + 1
		} else {
			im.ThermometerOut[i] = false
		}
	}

	// Apply activation function if enabled
	activatedValue := float64(digitalValue) / float64(im.Config.NumLevels)
	if im.Config.NonlinearActiv {
		im.Stats.Activations++
		switch im.Config.ActivationType {
		case "relu":
			if activatedValue < 0 {
				activatedValue = 0
			}
		case "sigmoid":
			activatedValue = 1.0 / (1.0 + math.Exp(-activatedValue*6+3))
		case "tanh":
			activatedValue = math.Tanh(activatedValue*2 - 1)
		}
	}

	return digitalValue, activatedValue
}

// ============================================================================
// RESONANT TIME-DOMAIN CIM (rTD-CiM)
// ============================================================================

// RTDCiMConfig defines resonant time-domain CIM parameters
// From GLSVLSI 2024: 2.36 TOPS, 28.05 TOPS/W
type RTDCiMConfig struct {
	ArraySize       int     `json:"array_size"`        // 8KB SRAM typical
	Technology      string  `json:"technology"`        // "28nm" TSMC
	Throughput      float64 `json:"throughput"`        // TOPS (2.36)
	Efficiency      float64 `json:"efficiency"`        // TOPS/W (28.05)
	ResonantFreq    float64 `json:"resonant_freq"`     // Resonant frequency (GHz)
}

// RTDCiM implements ADC-less resonant time-domain CIM
type RTDCiM struct {
	Config        *RTDCiMConfig
	SRAMArray     [][]int       // Binary SRAM cells
	TimeDomainAcc []float64     // Time-domain accumulators
	Oscillators   []*LCOscillator
	Stats         *RTDCiMStats
	mu            sync.RWMutex
}

// LCOscillator represents LC oscillator for time-domain computing
type LCOscillator struct {
	Frequency float64 // Current frequency
	Phase     float64 // Current phase
	Amplitude float64 // Signal amplitude
}

// RTDCiMStats tracks rTD-CiM performance
type RTDCiMStats struct {
	MACOperations  int64   `json:"mac_operations"`
	TotalCycles    int64   `json:"total_cycles"`
	EnergyConsumed float64 `json:"energy_consumed"` // pJ
}

// NewRTDCiM creates a new resonant time-domain CIM
func NewRTDCiM(config *RTDCiMConfig) *RTDCiM {
	arrayRows := 256
	arrayCols := 256

	sram := make([][]int, arrayRows)
	for i := range sram {
		sram[i] = make([]int, arrayCols)
	}

	oscillators := make([]*LCOscillator, arrayCols)
	for i := range oscillators {
		oscillators[i] = &LCOscillator{
			Frequency: config.ResonantFreq,
			Phase:     0,
			Amplitude: 1.0,
		}
	}

	return &RTDCiM{
		Config:        config,
		SRAMArray:     sram,
		TimeDomainAcc: make([]float64, arrayCols),
		Oscillators:   oscillators,
		Stats:         &RTDCiMStats{},
	}
}

// ComputeMAC performs ADC-less MAC using time-domain encoding
func (r *RTDCiM) ComputeMAC(input []int) []int {
	r.mu.Lock()
	defer r.mu.Unlock()

	output := make([]int, len(r.TimeDomainAcc))

	// Time-domain accumulation
	// MAC result encoded as oscillator phase shift
	for col := range r.TimeDomainAcc {
		phaseAccum := 0.0
		for row := 0; row < len(input) && row < len(r.SRAMArray); row++ {
			if input[row] == 1 && r.SRAMArray[row][col] == 1 {
				// Both input and weight are 1: add to phase
				phaseAccum += r.Oscillators[col].Frequency
			}
		}

		// Convert phase to digital via zero-crossing detection
		// No ADC needed - purely digital counting
		r.Oscillators[col].Phase += phaseAccum
		output[col] = int(phaseAccum) // Simplified: phase directly gives count

		r.Stats.MACOperations += int64(len(input))
	}

	r.Stats.TotalCycles++

	return output
}

// ============================================================================
// PIM-LLM ARCHITECTURE (1-bit LLMs)
// ============================================================================

// PIMLLMConfig defines PIM-LLM hybrid architecture parameters
// From research: 80× tokens/s improvement, 70% tokens/joule increase
type PIMLLMConfig struct {
	// Analog PIM for projection layers
	PIMArrays       int     `json:"pim_arrays"`
	PIMPrecision    int     `json:"pim_precision"`     // 1-bit for BitNet style

	// Digital systolic array for attention
	SystolicSize    int     `json:"systolic_size"`
	SystolicPrec    int     `json:"systolic_prec"`     // Higher precision for attention

	// Model parameters
	HiddenSize      int     `json:"hidden_size"`
	NumHeads        int     `json:"num_heads"`
}

// PIMLLM implements hybrid PIM architecture for 1-bit LLMs
type PIMLLM struct {
	Config        *PIMLLMConfig
	PIMEngine     *BinaryPIMEngine
	SystolicArray *DigitalSystolicArray
	Stats         *PIMLLMStats
	mu            sync.RWMutex
}

// PIMLLMStats tracks PIM-LLM performance
type PIMLLMStats struct {
	TokensGenerated int64   `json:"tokens_generated"`
	TokensPerSecond float64 `json:"tokens_per_second"`
	TokensPerJoule  float64 `json:"tokens_per_joule"`
	TotalEnergy     float64 `json:"total_energy"` // mJ
}

// BinaryPIMEngine handles 1-bit weight projection layers
type BinaryPIMEngine struct {
	Arrays    []*BinaryPIMArray
	mu        sync.RWMutex
}

// BinaryPIMArray represents 1-bit PIM crossbar
type BinaryPIMArray struct {
	Rows    int
	Cols    int
	Weights [][]int8 // -1, +1 only
}

// DigitalSystolicArray handles high-precision attention
type DigitalSystolicArray struct {
	Size    int
	PEs     [][]float64
	mu      sync.RWMutex
}

// NewPIMLLM creates a new PIM-LLM accelerator
func NewPIMLLM(config *PIMLLMConfig) *PIMLLM {
	// Initialize binary PIM arrays
	pimArrays := make([]*BinaryPIMArray, config.PIMArrays)
	for i := range pimArrays {
		weights := make([][]int8, 256)
		for j := range weights {
			weights[j] = make([]int8, 256)
			for k := range weights[j] {
				// Random binary weights for initialization
				if rand.Float32() > 0.5 {
					weights[j][k] = 1
				} else {
					weights[j][k] = -1
				}
			}
		}
		pimArrays[i] = &BinaryPIMArray{
			Rows:    256,
			Cols:    256,
			Weights: weights,
		}
	}

	pimEngine := &BinaryPIMEngine{Arrays: pimArrays}

	// Initialize systolic array
	pes := make([][]float64, config.SystolicSize)
	for i := range pes {
		pes[i] = make([]float64, config.SystolicSize)
	}

	systolic := &DigitalSystolicArray{
		Size: config.SystolicSize,
		PEs:  pes,
	}

	return &PIMLLM{
		Config:        config,
		PIMEngine:     pimEngine,
		SystolicArray: systolic,
		Stats:         &PIMLLMStats{},
	}
}

// ProcessProjection handles projection layer with 1-bit PIM
func (p *PIMLLM) ProcessProjection(input []float64) []float64 {
	p.PIMEngine.mu.Lock()
	defer p.PIMEngine.mu.Unlock()

	// Sign-based activation for binary compatibility
	binaryInput := make([]int8, len(input))
	for i, v := range input {
		if v >= 0 {
			binaryInput[i] = 1
		} else {
			binaryInput[i] = -1
		}
	}

	// Compute across PIM arrays
	outputSize := 0
	for _, arr := range p.PIMEngine.Arrays {
		outputSize += arr.Cols
	}
	output := make([]float64, outputSize)

	outIdx := 0
	for _, arr := range p.PIMEngine.Arrays {
		for col := 0; col < arr.Cols; col++ {
			sum := int(0)
			for row := 0; row < arr.Rows && row < len(binaryInput); row++ {
				// XNOR + popcount for binary MAC
				sum += int(binaryInput[row] * arr.Weights[row][col])
			}
			output[outIdx] = float64(sum)
			outIdx++
		}
	}

	return output
}

// ProcessAttention handles attention with digital systolic array
func (p *PIMLLM) ProcessAttention(Q, K, V [][]float64) [][]float64 {
	p.SystolicArray.mu.Lock()
	defer p.SystolicArray.mu.Unlock()

	seqLen := len(Q)
	headDim := len(Q[0])

	// Full-precision attention on systolic array
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, headDim)

		// Simplified attention computation
		for j := 0; j < seqLen; j++ {
			// Q·K^T
			score := 0.0
			for k := 0; k < headDim; k++ {
				score += Q[i][k] * K[j][k]
			}
			score /= math.Sqrt(float64(headDim))
			score = math.Exp(score) // Part of softmax

			// Attention × V
			for k := 0; k < headDim; k++ {
				output[i][k] += score * V[j][k]
			}
		}
	}

	return output
}

// GenerateToken simulates single token generation
func (p *PIMLLM) GenerateToken() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Stats.TokensGenerated++
	// Simplified: assume 80× improvement means ~80 tokens/s
	p.Stats.TokensPerSecond = 80.0
	p.Stats.TokensPerJoule = 1.7 // 70% improvement over baseline 1.0
}

// ============================================================================
// TRANSFORMER CIM ACCELERATOR (INTEGRATED)
// ============================================================================

// TransformerCIMAccelerator combines all transformer CIM components
type TransformerCIMAccelerator struct {
	Config       *TransformerCIMConfig
	HARDSEA      *HARDSEA
	HALO         *HALO
	PIMLLM       *PIMLLM
	KVCache      *KVCache
	AdaptiveADC  *AdaptiveADC
	IMADC        *IMADC
	RTDCiM       *RTDCiM
	mu           sync.RWMutex
}

// NewTransformerCIMAccelerator creates an integrated transformer accelerator
func NewTransformerCIMAccelerator(config *TransformerCIMConfig) *TransformerCIMAccelerator {
	return &TransformerCIMAccelerator{
		Config: config,
		HARDSEA: NewHARDSEA(&HARDSEAConfig{
			NumPEs:            16,
			SRAMCIMSize:       256,
			SparsityThreshold: 0.9,
			AnalogPrecision:   6,
			DigitalPrecision:  8,
		}),
		HALO: NewHALO(&HALOConfig{
			HBMCapacity:      16,
			HBMBandwidth:     1000.0,
			CiDThroughput:    10.0,
			CiMArrays:        32,
			CiMArraySize:     256,
			CiMPrecision:     6,
			InterleaverDepth: 4,
		}),
		PIMLLM: NewPIMLLM(&PIMLLMConfig{
			PIMArrays:    16,
			PIMPrecision: 1,
			SystolicSize: 64,
			SystolicPrec: 8,
			HiddenSize:   config.HiddenSize,
			NumHeads:     config.NumHeads,
		}),
		KVCache: NewKVCache(config.NumLayers, config.NumHeads, config.HeadDim, config.MaxSeqLen),
		AdaptiveADC: NewAdaptiveADC(&AdaptiveADCConfig{
			BaseBits:      config.ADCBits,
			MaxBits:       config.ADCBits + 2,
			AdaptiveMode:  config.UseAdaptiveADC,
			SparsityAware: true,
			EnergyPerBit:  5.0,
		}),
		IMADC: NewIMADC(&IMADCConfig{
			NumLevels:      64,
			AreaUM2:        45.0,
			EnergyFJ:       29.6,
			NonlinearActiv: true,
			ActivationType: "relu",
		}),
		RTDCiM: NewRTDCiM(&RTDCiMConfig{
			ArraySize:    8192,
			Technology:   "28nm",
			Throughput:   2.36,
			Efficiency:   28.05,
			ResonantFreq: 1.0,
		}),
	}
}

// ============================================================================
// SERIALIZATION
// ============================================================================

// TransformerCIMState holds serializable state
type TransformerCIMState struct {
	Config         *TransformerCIMConfig `json:"config"`
	KVCacheLen     int                   `json:"kv_cache_len"`
	KVCacheMemory  int64                 `json:"kv_cache_memory_bytes"`
	ADCStats       *ADCStats             `json:"adc_stats"`
	IMADCStats     *IMADCStats           `json:"imadc_stats"`
	RTDCiMStats    *RTDCiMStats          `json:"rtd_cim_stats"`
	PIMLLMStats    *PIMLLMStats          `json:"pim_llm_stats"`
}

// ExportState exports accelerator state
func (t *TransformerCIMAccelerator) ExportState() ([]byte, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	state := &TransformerCIMState{
		Config:        t.Config,
		KVCacheLen:    t.KVCache.CacheLen,
		KVCacheMemory: t.KVCache.GetMemoryFootprint(2), // FP16
		ADCStats:      t.AdaptiveADC.Stats,
		IMADCStats:    t.IMADC.Stats,
		RTDCiMStats:   t.RTDCiM.Stats,
		PIMLLMStats:   t.PIMLLM.Stats,
	}

	return json.MarshalIndent(state, "", "  ")
}
