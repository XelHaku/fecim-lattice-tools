// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"math"
	"math/rand"
)

// AttentionConfig holds configuration for attention layers.
type AttentionConfig struct {
	EmbedDim    int     // Embedding dimension (d_model)
	NumHeads    int     // Number of attention heads
	HeadDim     int     // Dimension per head (EmbedDim / NumHeads)
	DropoutRate float64 // Dropout rate for attention weights
	UseBias     bool    // Use bias in projections
	Causal      bool    // Use causal (autoregressive) masking
}

// DefaultAttentionConfig returns default attention configuration.
func DefaultAttentionConfig(embedDim, numHeads int) *AttentionConfig {
	return &AttentionConfig{
		EmbedDim:    embedDim,
		NumHeads:    numHeads,
		HeadDim:     embedDim / numHeads,
		DropoutRate: 0.1,
		UseBias:     true,
		Causal:      false,
	}
}

// ScaledDotProductAttention implements scaled dot-product attention.
// Attention(Q, K, V) = softmax(QK^T / sqrt(d_k)) V
type ScaledDotProductAttention struct {
	config   *AttentionConfig
	scale    float64
	Training bool
}

// NewScaledDotProductAttention creates a new attention layer.
func NewScaledDotProductAttention(config *AttentionConfig) *ScaledDotProductAttention {
	if config == nil {
		config = DefaultAttentionConfig(64, 1)
	}
	return &ScaledDotProductAttention{
		config:   config,
		scale:    1.0 / math.Sqrt(float64(config.HeadDim)),
		Training: false,
	}
}

// Forward computes attention.
// Q: [seqLen, headDim], K: [seqLen, headDim], V: [seqLen, headDim]
// Returns: [seqLen, headDim]
func (a *ScaledDotProductAttention) Forward(Q, K, V [][]float64) [][]float64 {
	seqLen := len(Q)
	headDim := len(Q[0])

	// Compute attention scores: QK^T
	scores := make([][]float64, seqLen)
	for i := range scores {
		scores[i] = make([]float64, seqLen)
		for j := range scores[i] {
			// Dot product Q[i] · K[j]
			for k := 0; k < headDim; k++ {
				scores[i][j] += Q[i][k] * K[j][k]
			}
			scores[i][j] *= a.scale
		}
	}

	// Apply causal mask if needed
	if a.config.Causal {
		for i := range scores {
			for j := i + 1; j < seqLen; j++ {
				scores[i][j] = math.Inf(-1)
			}
		}
	}

	// Softmax
	attnWeights := softmax2D(scores)

	// Apply dropout during training
	if a.Training && a.config.DropoutRate > 0 {
		attnWeights = applyDropout2D(attnWeights, a.config.DropoutRate)
	}

	// Compute weighted sum: attn_weights @ V
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, headDim)
		for j := 0; j < seqLen; j++ {
			for k := 0; k < headDim; k++ {
				output[i][k] += attnWeights[i][j] * V[j][k]
			}
		}
	}

	return output
}

// MultiHeadAttention implements multi-head attention mechanism.
type MultiHeadAttention struct {
	config *AttentionConfig

	// Linear projections: [embedDim][embedDim]
	WQ [][]float64 // Query projection
	WK [][]float64 // Key projection
	WV [][]float64 // Value projection
	WO [][]float64 // Output projection

	// Biases
	BQ []float64
	BK []float64
	BV []float64
	BO []float64

	attention *ScaledDotProductAttention
	Training  bool
}

// NewMultiHeadAttention creates a new multi-head attention layer.
func NewMultiHeadAttention(config *AttentionConfig) *MultiHeadAttention {
	if config == nil {
		config = DefaultAttentionConfig(64, 8)
	}

	mha := &MultiHeadAttention{
		config:    config,
		WQ:        initWeights(config.EmbedDim, config.EmbedDim),
		WK:        initWeights(config.EmbedDim, config.EmbedDim),
		WV:        initWeights(config.EmbedDim, config.EmbedDim),
		WO:        initWeights(config.EmbedDim, config.EmbedDim),
		attention: NewScaledDotProductAttention(config),
		Training:  false,
	}

	if config.UseBias {
		mha.BQ = make([]float64, config.EmbedDim)
		mha.BK = make([]float64, config.EmbedDim)
		mha.BV = make([]float64, config.EmbedDim)
		mha.BO = make([]float64, config.EmbedDim)
	}

	return mha
}

// Forward computes multi-head attention.
// Input: [seqLen, embedDim]
// Returns: [seqLen, embedDim]
func (mha *MultiHeadAttention) Forward(input [][]float64) [][]float64 {
	seqLen := len(input)
	embedDim := mha.config.EmbedDim
	numHeads := mha.config.NumHeads
	headDim := mha.config.HeadDim

	mha.attention.Training = mha.Training

	// Linear projections
	Q := matmul(input, mha.WQ)
	K := matmul(input, mha.WK)
	V := matmul(input, mha.WV)

	// Add biases
	if mha.config.UseBias {
		addBias(Q, mha.BQ)
		addBias(K, mha.BK)
		addBias(V, mha.BV)
	}

	// Split into heads and compute attention
	headOutputs := make([][][]float64, numHeads)
	for h := 0; h < numHeads; h++ {
		// Extract head slice
		Qh := extractHead(Q, h, headDim)
		Kh := extractHead(K, h, headDim)
		Vh := extractHead(V, h, headDim)

		// Compute attention for this head
		headOutputs[h] = mha.attention.Forward(Qh, Kh, Vh)
	}

	// Concatenate heads
	concat := concatHeads(headOutputs, seqLen, embedDim)

	// Output projection
	output := matmul(concat, mha.WO)
	if mha.config.UseBias {
		addBias(output, mha.BO)
	}

	return output
}

// GetCrossbarDimensions returns crossbar sizes needed for Q, K, V, O projections.
func (mha *MultiHeadAttention) GetCrossbarDimensions() (qkv, output [2]int) {
	// Q, K, V projections: embedDim x embedDim each
	qkv = [2]int{mha.config.EmbedDim, mha.config.EmbedDim}
	// Output projection: embedDim x embedDim
	output = [2]int{mha.config.EmbedDim, mha.config.EmbedDim}
	return
}

// GetProjectionMatrices returns all projection matrices for crossbar programming.
func (mha *MultiHeadAttention) GetProjectionMatrices() (WQ, WK, WV, WO [][]float64) {
	return mha.WQ, mha.WK, mha.WV, mha.WO
}

// SlidingWindowAttention implements hardware-friendly sliding window attention.
// Limits attention to M most recent tokens instead of full sequence.
// More suitable for analog CIM due to fixed memory requirements.
type SlidingWindowAttention struct {
	config     *AttentionConfig
	WindowSize int     // Maximum window size (M)
	scale      float64 // 1/sqrt(d_k)

	// Cached keys and values for sliding window
	KeyCache   [][]float64 // [WindowSize][HeadDim]
	ValueCache [][]float64 // [WindowSize][HeadDim]
	CachePos   int         // Current position in circular buffer

	Training bool
}

// NewSlidingWindowAttention creates sliding window attention.
func NewSlidingWindowAttention(config *AttentionConfig, windowSize int) *SlidingWindowAttention {
	if config == nil {
		config = DefaultAttentionConfig(64, 1)
	}

	swa := &SlidingWindowAttention{
		config:     config,
		WindowSize: windowSize,
		scale:      1.0 / math.Sqrt(float64(config.HeadDim)),
		KeyCache:   make([][]float64, windowSize),
		ValueCache: make([][]float64, windowSize),
		CachePos:   0,
		Training:   false,
	}

	// Initialize cache
	for i := 0; i < windowSize; i++ {
		swa.KeyCache[i] = make([]float64, config.HeadDim)
		swa.ValueCache[i] = make([]float64, config.HeadDim)
	}

	return swa
}

// UpdateCache adds new key-value pair to sliding window.
func (swa *SlidingWindowAttention) UpdateCache(key, value []float64) {
	copy(swa.KeyCache[swa.CachePos], key)
	copy(swa.ValueCache[swa.CachePos], value)
	swa.CachePos = (swa.CachePos + 1) % swa.WindowSize
}

// Forward computes attention for single query against cached keys/values.
// Query: [HeadDim]
// Returns: [HeadDim]
func (swa *SlidingWindowAttention) Forward(query []float64) []float64 {
	headDim := swa.config.HeadDim

	// Compute attention scores against all cached keys
	scores := make([]float64, swa.WindowSize)
	for i := 0; i < swa.WindowSize; i++ {
		for k := 0; k < headDim; k++ {
			scores[i] += query[k] * swa.KeyCache[i][k]
		}
		scores[i] *= swa.scale
	}

	// Softmax
	attnWeights := softmax1D(scores)

	// Apply dropout
	if swa.Training && swa.config.DropoutRate > 0 {
		applyDropout1D(attnWeights, swa.config.DropoutRate)
	}

	// Weighted sum of values
	output := make([]float64, headDim)
	for i := 0; i < swa.WindowSize; i++ {
		for k := 0; k < headDim; k++ {
			output[k] += attnWeights[i] * swa.ValueCache[i][k]
		}
	}

	return output
}

// ResetCache clears the sliding window cache.
func (swa *SlidingWindowAttention) ResetCache() {
	for i := range swa.KeyCache {
		for j := range swa.KeyCache[i] {
			swa.KeyCache[i][j] = 0
		}
		for j := range swa.ValueCache[i] {
			swa.ValueCache[i][j] = 0
		}
	}
	swa.CachePos = 0
}

// CrossbarAttention implements attention optimized for crossbar array mapping.
// Uses HardSigmoid activation instead of softmax for hardware efficiency.
type CrossbarAttention struct {
	config *AttentionConfig
	scale  float64

	// Use HardSigmoid instead of softmax
	UseHardSigmoid bool

	// Quantization parameters
	InputBits  int // PWM input quantization
	WeightBits int // K/V storage quantization
	OutputBits int // ADC output quantization

	Training bool
}

// NewCrossbarAttention creates attention optimized for crossbar arrays.
func NewCrossbarAttention(config *AttentionConfig) *CrossbarAttention {
	if config == nil {
		config = DefaultAttentionConfig(64, 1)
	}

	return &CrossbarAttention{
		config:         config,
		scale:          1.0 / math.Sqrt(float64(config.HeadDim)),
		UseHardSigmoid: true,
		InputBits:      4,  // 16 levels for input PWM
		WeightBits:     3,  // 8 levels for K/V storage
		OutputBits:     5,  // 32 levels for output ADC
		Training:       false,
	}
}

// Forward computes crossbar-friendly attention.
func (ca *CrossbarAttention) Forward(Q, K, V [][]float64) [][]float64 {
	seqLen := len(Q)
	headDim := len(Q[0])

	// Quantize inputs if not training
	if !ca.Training {
		Q = quantize2D(Q, ca.InputBits)
		K = quantize2D(K, ca.WeightBits)
		V = quantize2D(V, ca.WeightBits)
	}

	// Compute attention scores: QK^T (simulated crossbar MVM)
	scores := make([][]float64, seqLen)
	for i := range scores {
		scores[i] = make([]float64, seqLen)
		for j := range scores[i] {
			// Crossbar dot product
			for k := 0; k < headDim; k++ {
				scores[i][j] += Q[i][k] * K[j][k]
			}
			scores[i][j] *= ca.scale
		}
	}

	// Apply causal mask
	if ca.config.Causal {
		for i := range scores {
			for j := i + 1; j < seqLen; j++ {
				scores[i][j] = -1e9 // Large negative for HardSigmoid
			}
		}
	}

	// Use HardSigmoid instead of Softmax for hardware
	var attnWeights [][]float64
	if ca.UseHardSigmoid {
		attnWeights = hardSigmoidNormalize(scores)
	} else {
		attnWeights = softmax2D(scores)
	}

	// Quantize attention output
	if !ca.Training {
		attnWeights = quantize2D(attnWeights, ca.OutputBits)
	}

	// Compute weighted sum: attn_weights @ V (second crossbar MVM)
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, headDim)
		for j := 0; j < seqLen; j++ {
			for k := 0; k < headDim; k++ {
				output[i][k] += attnWeights[i][j] * V[j][k]
			}
		}
	}

	// Final quantization
	if !ca.Training {
		output = quantize2D(output, ca.OutputBits)
	}

	return output
}

// GetCrossbarConfig returns crossbar array configuration for this attention layer.
func (ca *CrossbarAttention) GetCrossbarConfig() map[string]interface{} {
	return map[string]interface{}{
		"qk_crossbar_rows": ca.config.HeadDim,
		"qk_crossbar_cols": ca.config.HeadDim, // For K storage
		"av_crossbar_rows": ca.config.HeadDim, // For attention weights
		"av_crossbar_cols": ca.config.HeadDim, // For V storage
		"input_bits":       ca.InputBits,
		"weight_bits":      ca.WeightBits,
		"output_bits":      ca.OutputBits,
		"use_hardsigmoid":  ca.UseHardSigmoid,
	}
}

// HardSigmoid implements hardware-friendly activation.
// HardSigmoid(x) = clip(0.2x + 0.5, 0, 1)
func HardSigmoid(x float64) float64 {
	result := 0.2*x + 0.5
	if result < 0 {
		return 0
	}
	if result > 1 {
		return 1
	}
	return result
}

// HardSigmoidVec applies HardSigmoid element-wise.
func HardSigmoidVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = HardSigmoid(v)
	}
	return result
}

// hardSigmoidNormalize applies HardSigmoid and normalizes per row.
func hardSigmoidNormalize(scores [][]float64) [][]float64 {
	result := make([][]float64, len(scores))
	for i := range result {
		result[i] = make([]float64, len(scores[i]))

		// Apply HardSigmoid
		sum := 0.0
		for j := range result[i] {
			result[i][j] = HardSigmoid(scores[i][j])
			sum += result[i][j]
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

// ReLUAttention implements ReLU-based attention (alternative to softmax).
// More hardware-friendly but may require different training approach.
func ReLUAttention(scores [][]float64) [][]float64 {
	result := make([][]float64, len(scores))
	for i := range result {
		result[i] = make([]float64, len(scores[i]))

		sum := 0.0
		for j := range result[i] {
			if scores[i][j] > 0 {
				result[i][j] = scores[i][j]
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

// Helper functions

func initWeights(rows, cols int) [][]float64 {
	w := make([][]float64, rows)
	stddev := math.Sqrt(2.0 / float64(rows+cols))
	for i := range w {
		w[i] = make([]float64, cols)
		for j := range w[i] {
			w[i][j] = rand.NormFloat64() * stddev
		}
	}
	return w
}

func matmul(A, B [][]float64) [][]float64 {
	m := len(A)
	if m == 0 {
		return nil
	}
	k := len(A[0])
	n := len(B[0])

	C := make([][]float64, m)
	for i := range C {
		C[i] = make([]float64, n)
		for j := 0; j < n; j++ {
			for l := 0; l < k; l++ {
				C[i][j] += A[i][l] * B[l][j]
			}
		}
	}
	return C
}

func addBias(A [][]float64, bias []float64) {
	for i := range A {
		for j := range A[i] {
			if j < len(bias) {
				A[i][j] += bias[j]
			}
		}
	}
}

func extractHead(X [][]float64, headIdx, headDim int) [][]float64 {
	result := make([][]float64, len(X))
	start := headIdx * headDim
	for i := range result {
		result[i] = make([]float64, headDim)
		for j := 0; j < headDim; j++ {
			if start+j < len(X[i]) {
				result[i][j] = X[i][start+j]
			}
		}
	}
	return result
}

func concatHeads(heads [][][]float64, seqLen, embedDim int) [][]float64 {
	result := make([][]float64, seqLen)
	for i := range result {
		result[i] = make([]float64, embedDim)
	}

	headDim := embedDim / len(heads)
	for h, head := range heads {
		for i := range head {
			for j := range head[i] {
				result[i][h*headDim+j] = head[i][j]
			}
		}
	}
	return result
}

func softmax1D(x []float64) []float64 {
	// Find max for numerical stability
	maxVal := x[0]
	for _, v := range x {
		if v > maxVal {
			maxVal = v
		}
	}

	// Compute exp and sum
	result := make([]float64, len(x))
	sum := 0.0
	for i, v := range x {
		result[i] = math.Exp(v - maxVal)
		sum += result[i]
	}

	// Normalize
	for i := range result {
		result[i] /= sum
	}
	return result
}

func softmax2D(x [][]float64) [][]float64 {
	result := make([][]float64, len(x))
	for i := range result {
		result[i] = softmax1D(x[i])
	}
	return result
}

func applyDropout1D(x []float64, rate float64) {
	scale := 1.0 / (1.0 - rate)
	for i := range x {
		if rand.Float64() < rate {
			x[i] = 0
		} else {
			x[i] *= scale
		}
	}
}

func applyDropout2D(x [][]float64, rate float64) [][]float64 {
	result := make([][]float64, len(x))
	scale := 1.0 / (1.0 - rate)
	for i := range result {
		result[i] = make([]float64, len(x[i]))
		for j := range result[i] {
			if rand.Float64() < rate {
				result[i][j] = 0
			} else {
				result[i][j] = x[i][j] * scale
			}
		}
	}
	return result
}

func quantize2D(x [][]float64, bits int) [][]float64 {
	levels := float64(int(1) << bits)
	result := make([][]float64, len(x))

	// Find min/max
	minVal, maxVal := x[0][0], x[0][0]
	for i := range x {
		for j := range x[i] {
			if x[i][j] < minVal {
				minVal = x[i][j]
			}
			if x[i][j] > maxVal {
				maxVal = x[i][j]
			}
		}
	}

	scale := (maxVal - minVal) / levels
	if scale == 0 {
		scale = 1
	}

	for i := range result {
		result[i] = make([]float64, len(x[i]))
		for j := range result[i] {
			// Quantize and dequantize
			q := math.Round((x[i][j] - minVal) / scale)
			result[i][j] = q*scale + minVal
		}
	}
	return result
}
