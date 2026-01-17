// Package layers provides neural network layer implementations for crossbar arrays.
package layers

import (
	"math"
	"math/rand"
)

// TransformerConfig holds configuration for transformer blocks.
type TransformerConfig struct {
	EmbedDim     int     // Embedding dimension (d_model)
	NumHeads     int     // Number of attention heads
	FFNDim       int     // Feed-forward network hidden dimension (typically 4*EmbedDim)
	DropoutRate  float64 // Dropout rate
	LayerNormEps float64 // LayerNorm epsilon
	PreNorm      bool    // Use pre-normalization (modern transformers) vs post-norm
	UseBias      bool    // Use bias in linear projections
	Causal       bool    // Causal attention (for decoders)
}

// DefaultTransformerConfig returns default transformer configuration.
func DefaultTransformerConfig(embedDim int) *TransformerConfig {
	return &TransformerConfig{
		EmbedDim:     embedDim,
		NumHeads:     8,
		FFNDim:       4 * embedDim,
		DropoutRate:  0.1,
		LayerNormEps: 1e-5,
		PreNorm:      true, // Modern default
		UseBias:      true,
		Causal:       false,
	}
}

// FeedForwardNetwork implements the FFN component of transformer.
// FFN(x) = Linear(GELU(Linear(x)))
type FeedForwardNetwork struct {
	EmbedDim    int
	FFNDim      int
	DropoutRate float64

	// Weights
	W1 [][]float64 // [EmbedDim][FFNDim]
	B1 []float64   // [FFNDim]
	W2 [][]float64 // [FFNDim][EmbedDim]
	B2 []float64   // [EmbedDim]

	Training bool
}

// NewFeedForwardNetwork creates a new FFN layer.
func NewFeedForwardNetwork(embedDim, ffnDim int, dropoutRate float64) *FeedForwardNetwork {
	ffn := &FeedForwardNetwork{
		EmbedDim:    embedDim,
		FFNDim:      ffnDim,
		DropoutRate: dropoutRate,
		W1:          initWeightsXavier(embedDim, ffnDim),
		B1:          make([]float64, ffnDim),
		W2:          initWeightsXavier(ffnDim, embedDim),
		B2:          make([]float64, embedDim),
		Training:    false,
	}
	return ffn
}

// Forward applies FFN to input.
// Input: [seqLen, embedDim] -> Output: [seqLen, embedDim]
func (ffn *FeedForwardNetwork) Forward(input [][]float64) [][]float64 {
	seqLen := len(input)

	// First linear + GELU
	hidden := make([][]float64, seqLen)
	for i := range hidden {
		hidden[i] = make([]float64, ffn.FFNDim)
		for j := 0; j < ffn.FFNDim; j++ {
			sum := ffn.B1[j]
			for k := 0; k < ffn.EmbedDim; k++ {
				sum += input[i][k] * ffn.W1[k][j]
			}
			hidden[i][j] = GELU(sum)
		}
	}

	// Apply dropout after activation
	if ffn.Training && ffn.DropoutRate > 0 {
		hidden = applyDropout2D(hidden, ffn.DropoutRate)
	}

	// Second linear
	output := make([][]float64, seqLen)
	for i := range output {
		output[i] = make([]float64, ffn.EmbedDim)
		for j := 0; j < ffn.EmbedDim; j++ {
			sum := ffn.B2[j]
			for k := 0; k < ffn.FFNDim; k++ {
				sum += hidden[i][k] * ffn.W2[k][j]
			}
			output[i][j] = sum
		}
	}

	return output
}

// GetCrossbarDimensions returns crossbar sizes for FFN.
func (ffn *FeedForwardNetwork) GetCrossbarDimensions() (fc1, fc2 [2]int) {
	fc1 = [2]int{ffn.EmbedDim, ffn.FFNDim}
	fc2 = [2]int{ffn.FFNDim, ffn.EmbedDim}
	return
}

// TransformerBlock implements a single transformer encoder/decoder block.
// Contains: LayerNorm -> MultiHeadAttention -> Residual -> LayerNorm -> FFN -> Residual
type TransformerBlock struct {
	config *TransformerConfig

	// Sub-layers
	Attention *MultiHeadAttention
	FFN       *FeedForwardNetwork

	// Layer normalization
	LN1 *LayerNorm // Before/after attention
	LN2 *LayerNorm // Before/after FFN

	// Dropout
	DropoutAttn *Dropout
	DropoutFFN  *Dropout

	Training bool
}

// NewTransformerBlock creates a new transformer block.
func NewTransformerBlock(config *TransformerConfig) *TransformerBlock {
	if config == nil {
		config = DefaultTransformerConfig(512)
	}

	attnConfig := &AttentionConfig{
		EmbedDim:    config.EmbedDim,
		NumHeads:    config.NumHeads,
		HeadDim:     config.EmbedDim / config.NumHeads,
		DropoutRate: config.DropoutRate,
		UseBias:     config.UseBias,
		Causal:      config.Causal,
	}

	block := &TransformerBlock{
		config:      config,
		Attention:   NewMultiHeadAttention(attnConfig),
		FFN:         NewFeedForwardNetwork(config.EmbedDim, config.FFNDim, config.DropoutRate),
		LN1:         NewLayerNorm(config.EmbedDim),
		LN2:         NewLayerNorm(config.EmbedDim),
		DropoutAttn: NewDropout(config.DropoutRate),
		DropoutFFN:  NewDropout(config.DropoutRate),
		Training:    false,
	}

	block.LN1.Epsilon = config.LayerNormEps
	block.LN2.Epsilon = config.LayerNormEps

	return block
}

// Forward applies transformer block to input.
// Input: [seqLen, embedDim] -> Output: [seqLen, embedDim]
func (tb *TransformerBlock) Forward(input [][]float64) [][]float64 {
	tb.Attention.Training = tb.Training
	tb.FFN.Training = tb.Training
	tb.DropoutAttn.Training = tb.Training
	tb.DropoutFFN.Training = tb.Training

	var output [][]float64

	if tb.config.PreNorm {
		// Pre-normalization (GPT-2, LLaMA style)
		// x = x + Attention(LayerNorm(x))
		// x = x + FFN(LayerNorm(x))

		// Attention sub-block
		normed := tb.applyLayerNorm(input, tb.LN1)
		attnOut := tb.Attention.Forward(normed)
		attnOut = tb.DropoutAttn.Forward2D(attnOut)
		x := addResidual(input, attnOut)

		// FFN sub-block
		normed = tb.applyLayerNorm(x, tb.LN2)
		ffnOut := tb.FFN.Forward(normed)
		ffnOut = tb.DropoutFFN.Forward2D(ffnOut)
		output = addResidual(x, ffnOut)
	} else {
		// Post-normalization (Original transformer, BERT style)
		// x = LayerNorm(x + Attention(x))
		// x = LayerNorm(x + FFN(x))

		// Attention sub-block
		attnOut := tb.Attention.Forward(input)
		attnOut = tb.DropoutAttn.Forward2D(attnOut)
		x := addResidual(input, attnOut)
		x = tb.applyLayerNorm(x, tb.LN1)

		// FFN sub-block
		ffnOut := tb.FFN.Forward(x)
		ffnOut = tb.DropoutFFN.Forward2D(ffnOut)
		output = addResidual(x, ffnOut)
		output = tb.applyLayerNorm(output, tb.LN2)
	}

	return output
}

// applyLayerNorm applies layer normalization to batch.
func (tb *TransformerBlock) applyLayerNorm(input [][]float64, ln *LayerNorm) [][]float64 {
	output := make([][]float64, len(input))
	for i := range output {
		output[i] = ln.Forward(input[i])
	}
	return output
}

// GetCrossbarDimensions returns total crossbar requirements.
func (tb *TransformerBlock) GetCrossbarDimensions() map[string][2]int {
	qkv, out := tb.Attention.GetCrossbarDimensions()
	fc1, fc2 := tb.FFN.GetCrossbarDimensions()

	return map[string][2]int{
		"attention_q":   qkv,
		"attention_k":   qkv,
		"attention_v":   qkv,
		"attention_out": out,
		"ffn_fc1":       fc1,
		"ffn_fc2":       fc2,
	}
}

// SetTraining sets training mode for all sub-layers.
func (tb *TransformerBlock) SetTraining(training bool) {
	tb.Training = training
	tb.Attention.Training = training
	tb.FFN.Training = training
	tb.DropoutAttn.Training = training
	tb.DropoutFFN.Training = training
}

// TransformerEncoder implements a stack of transformer encoder blocks.
type TransformerEncoder struct {
	config *TransformerConfig
	Blocks []*TransformerBlock
	FinalLN *LayerNorm // Final layer norm (if pre-norm)

	Training bool
}

// NewTransformerEncoder creates a transformer encoder with N layers.
func NewTransformerEncoder(config *TransformerConfig, numLayers int) *TransformerEncoder {
	if config == nil {
		config = DefaultTransformerConfig(512)
	}

	encoder := &TransformerEncoder{
		config: config,
		Blocks: make([]*TransformerBlock, numLayers),
	}

	for i := 0; i < numLayers; i++ {
		encoder.Blocks[i] = NewTransformerBlock(config)
	}

	// Final layer norm for pre-norm architecture
	if config.PreNorm {
		encoder.FinalLN = NewLayerNorm(config.EmbedDim)
		encoder.FinalLN.Epsilon = config.LayerNormEps
	}

	return encoder
}

// Forward applies all transformer blocks sequentially.
func (enc *TransformerEncoder) Forward(input [][]float64) [][]float64 {
	x := input

	for _, block := range enc.Blocks {
		block.Training = enc.Training
		x = block.Forward(x)
	}

	// Apply final layer norm for pre-norm
	if enc.config.PreNorm && enc.FinalLN != nil {
		output := make([][]float64, len(x))
		for i := range output {
			output[i] = enc.FinalLN.Forward(x[i])
		}
		x = output
	}

	return x
}

// GetTotalCrossbars returns total number of crossbar arrays needed.
func (enc *TransformerEncoder) GetTotalCrossbars() int {
	// Each block needs: 4 for attention (Q, K, V, O) + 2 for FFN
	return len(enc.Blocks) * 6
}

// CrossbarTransformerBlock implements transformer optimized for crossbar deployment.
// Uses HardSigmoid attention and fused operations.
type CrossbarTransformerBlock struct {
	config *TransformerConfig

	// Crossbar-optimized attention
	Attention *CrossbarAttention

	// FFN with quantization
	FFN *FeedForwardNetwork

	// RMSNorm instead of LayerNorm (hardware-friendly)
	RMSNorm1 *RMSNorm
	RMSNorm2 *RMSNorm

	// Quantization parameters
	InputBits  int
	WeightBits int
	OutputBits int

	Training bool
}

// NewCrossbarTransformerBlock creates crossbar-optimized transformer block.
func NewCrossbarTransformerBlock(config *TransformerConfig) *CrossbarTransformerBlock {
	if config == nil {
		config = DefaultTransformerConfig(512)
	}

	attnConfig := &AttentionConfig{
		EmbedDim:    config.EmbedDim,
		NumHeads:    config.NumHeads,
		HeadDim:     config.EmbedDim / config.NumHeads,
		DropoutRate: config.DropoutRate,
		UseBias:     config.UseBias,
		Causal:      config.Causal,
	}

	return &CrossbarTransformerBlock{
		config:     config,
		Attention:  NewCrossbarAttention(attnConfig),
		FFN:        NewFeedForwardNetwork(config.EmbedDim, config.FFNDim, config.DropoutRate),
		RMSNorm1:   NewRMSNorm(config.EmbedDim),
		RMSNorm2:   NewRMSNorm(config.EmbedDim),
		InputBits:  4,
		WeightBits: 4,
		OutputBits: 8,
		Training:   false,
	}
}

// Forward applies crossbar-optimized transformer block.
func (ctb *CrossbarTransformerBlock) Forward(input [][]float64) [][]float64 {
	ctb.Attention.Training = ctb.Training
	ctb.FFN.Training = ctb.Training

	// Pre-norm with RMSNorm (more hardware-friendly)
	normed := make([][]float64, len(input))
	for i := range normed {
		normed[i] = ctb.RMSNorm1.Forward(input[i])
	}

	// Crossbar attention (uses HardSigmoid)
	// Note: CrossbarAttention expects Q, K, V separately
	// For self-attention, we use input as Q, K, V
	attnOut := ctb.Attention.Forward(normed, normed, normed)

	// Residual connection
	x := addResidual(input, attnOut)

	// FFN sub-block
	normed = make([][]float64, len(x))
	for i := range normed {
		normed[i] = ctb.RMSNorm2.Forward(x[i])
	}

	ffnOut := ctb.FFN.Forward(normed)
	output := addResidual(x, ffnOut)

	// Quantize output if not training
	if !ctb.Training {
		output = quantize2D(output, ctb.OutputBits)
	}

	return output
}

// PositionalEncoding implements sinusoidal positional encoding.
type PositionalEncoding struct {
	MaxLen   int
	EmbedDim int
	Encoding [][]float64 // Precomputed encodings
}

// NewPositionalEncoding creates positional encoding.
func NewPositionalEncoding(maxLen, embedDim int) *PositionalEncoding {
	pe := &PositionalEncoding{
		MaxLen:   maxLen,
		EmbedDim: embedDim,
		Encoding: make([][]float64, maxLen),
	}

	// Precompute sinusoidal encodings
	for pos := 0; pos < maxLen; pos++ {
		pe.Encoding[pos] = make([]float64, embedDim)
		for i := 0; i < embedDim; i++ {
			angle := float64(pos) / math.Pow(10000, float64(2*(i/2))/float64(embedDim))
			if i%2 == 0 {
				pe.Encoding[pos][i] = math.Sin(angle)
			} else {
				pe.Encoding[pos][i] = math.Cos(angle)
			}
		}
	}

	return pe
}

// Forward adds positional encoding to input.
func (pe *PositionalEncoding) Forward(input [][]float64) [][]float64 {
	seqLen := len(input)
	output := make([][]float64, seqLen)

	for i := range output {
		output[i] = make([]float64, len(input[i]))
		for j := range output[i] {
			output[i][j] = input[i][j]
			if i < pe.MaxLen && j < pe.EmbedDim {
				output[i][j] += pe.Encoding[i][j]
			}
		}
	}

	return output
}

// TokenEmbedding implements learnable token embeddings.
type TokenEmbedding struct {
	VocabSize int
	EmbedDim  int
	Embedding [][]float64 // [VocabSize][EmbedDim]
}

// NewTokenEmbedding creates token embedding layer.
func NewTokenEmbedding(vocabSize, embedDim int) *TokenEmbedding {
	te := &TokenEmbedding{
		VocabSize: vocabSize,
		EmbedDim:  embedDim,
		Embedding: make([][]float64, vocabSize),
	}

	// Initialize with scaled random values
	scale := 1.0 / math.Sqrt(float64(embedDim))
	for i := range te.Embedding {
		te.Embedding[i] = make([]float64, embedDim)
		for j := range te.Embedding[i] {
			te.Embedding[i][j] = rand.NormFloat64() * scale
		}
	}

	return te
}

// Forward looks up embeddings for token indices.
func (te *TokenEmbedding) Forward(tokens []int) [][]float64 {
	output := make([][]float64, len(tokens))
	for i, tok := range tokens {
		output[i] = make([]float64, te.EmbedDim)
		if tok >= 0 && tok < te.VocabSize {
			copy(output[i], te.Embedding[tok])
		}
	}
	return output
}

// ReservoirLayer implements a reservoir computing layer using FeFET dynamics.
// Suitable for temporal sequence processing.
type ReservoirLayer struct {
	NumNodes   int
	InputDim   int
	OutputDim  int
	SpectralRadius float64

	// Reservoir weights (random, fixed)
	Win  [][]float64 // Input weights [NumNodes][InputDim]
	Wres [][]float64 // Recurrent weights [NumNodes][NumNodes]

	// Readout weights (trainable)
	Wout [][]float64 // Output weights [OutputDim][NumNodes]

	// Reservoir state
	State []float64

	// FeFET-like dynamics
	LeakRate float64 // State decay rate
	STMDecay float64 // Short-term memory decay

	Training bool
}

// NewReservoirLayer creates a reservoir computing layer.
func NewReservoirLayer(inputDim, numNodes, outputDim int) *ReservoirLayer {
	rl := &ReservoirLayer{
		NumNodes:       numNodes,
		InputDim:       inputDim,
		OutputDim:      outputDim,
		SpectralRadius: 0.9,
		Win:            make([][]float64, numNodes),
		Wres:           make([][]float64, numNodes),
		Wout:           make([][]float64, outputDim),
		State:          make([]float64, numNodes),
		LeakRate:       0.3,
		STMDecay:       0.1,
		Training:       false,
	}

	// Initialize input weights (sparse, random)
	inputSparsity := 0.2
	for i := range rl.Win {
		rl.Win[i] = make([]float64, inputDim)
		for j := range rl.Win[i] {
			if rand.Float64() < inputSparsity {
				rl.Win[i][j] = rand.NormFloat64() * 0.5
			}
		}
	}

	// Initialize reservoir weights (sparse, random)
	resSparsity := 0.1
	for i := range rl.Wres {
		rl.Wres[i] = make([]float64, numNodes)
		for j := range rl.Wres[i] {
			if rand.Float64() < resSparsity {
				rl.Wres[i][j] = rand.NormFloat64()
			}
		}
	}

	// Scale reservoir weights to achieve desired spectral radius
	rl.scaleSpectralRadius()

	// Initialize output weights
	for i := range rl.Wout {
		rl.Wout[i] = make([]float64, numNodes)
	}

	return rl
}

// scaleSpectralRadius scales reservoir weights to achieve target spectral radius.
func (rl *ReservoirLayer) scaleSpectralRadius() {
	// Estimate spectral radius using power iteration
	v := make([]float64, rl.NumNodes)
	for i := range v {
		v[i] = rand.Float64()
	}

	for iter := 0; iter < 100; iter++ {
		// v = Wres @ v
		newV := make([]float64, rl.NumNodes)
		for i := range newV {
			for j := range v {
				newV[i] += rl.Wres[i][j] * v[j]
			}
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

	// Compute Rayleigh quotient
	Wv := make([]float64, rl.NumNodes)
	for i := range Wv {
		for j := range v {
			Wv[i] += rl.Wres[i][j] * v[j]
		}
	}

	eigenval := 0.0
	for i := range v {
		eigenval += v[i] * Wv[i]
	}
	eigenval = math.Abs(eigenval)

	// Scale to target spectral radius
	if eigenval > 0 {
		scale := rl.SpectralRadius / eigenval
		for i := range rl.Wres {
			for j := range rl.Wres[i] {
				rl.Wres[i][j] *= scale
			}
		}
	}
}

// Forward processes input through reservoir.
// Input: [InputDim] -> Output: [OutputDim]
func (rl *ReservoirLayer) Forward(input []float64) []float64 {
	// Compute input contribution
	inputContrib := make([]float64, rl.NumNodes)
	for i := range inputContrib {
		for j := 0; j < len(input) && j < rl.InputDim; j++ {
			inputContrib[i] += rl.Win[i][j] * input[j]
		}
	}

	// Compute recurrent contribution
	recurrentContrib := make([]float64, rl.NumNodes)
	for i := range recurrentContrib {
		for j := range rl.State {
			recurrentContrib[i] += rl.Wres[i][j] * rl.State[j]
		}
	}

	// Update state with leaky integration (FeFET-like dynamics)
	for i := range rl.State {
		preActivation := inputContrib[i] + recurrentContrib[i]
		// Leaky integration with tanh nonlinearity
		rl.State[i] = (1-rl.LeakRate)*rl.State[i] + rl.LeakRate*math.Tanh(preActivation)
	}

	// Compute output
	output := make([]float64, rl.OutputDim)
	for i := range output {
		for j := range rl.State {
			output[i] += rl.Wout[i][j] * rl.State[j]
		}
	}

	return output
}

// ForwardSequence processes a sequence of inputs.
func (rl *ReservoirLayer) ForwardSequence(inputs [][]float64) [][]float64 {
	outputs := make([][]float64, len(inputs))
	for i, input := range inputs {
		outputs[i] = rl.Forward(input)
	}
	return outputs
}

// CollectStates returns reservoir states for entire sequence (for training readout).
func (rl *ReservoirLayer) CollectStates(inputs [][]float64) [][]float64 {
	states := make([][]float64, len(inputs))
	for i, input := range inputs {
		rl.Forward(input) // Updates internal state
		states[i] = make([]float64, rl.NumNodes)
		copy(states[i], rl.State)
	}
	return states
}

// ResetState resets reservoir state to zero.
func (rl *ReservoirLayer) ResetState() {
	for i := range rl.State {
		rl.State[i] = 0
	}
}

// TrainReadout trains output weights using ridge regression.
func (rl *ReservoirLayer) TrainReadout(states [][]float64, targets [][]float64, regularization float64) {
	// Simple ridge regression: Wout = Y @ X^T @ (X @ X^T + λI)^(-1)
	// For simplicity, use pseudo-inverse approach

	numSamples := len(states)
	if numSamples == 0 {
		return
	}

	// Compute X^T @ X + λI
	XTX := make([][]float64, rl.NumNodes)
	for i := range XTX {
		XTX[i] = make([]float64, rl.NumNodes)
		for j := range XTX[i] {
			for k := 0; k < numSamples; k++ {
				XTX[i][j] += states[k][i] * states[k][j]
			}
			if i == j {
				XTX[i][j] += regularization
			}
		}
	}

	// Compute X^T @ Y
	XTY := make([][]float64, rl.NumNodes)
	for i := range XTY {
		XTY[i] = make([]float64, rl.OutputDim)
		for j := range XTY[i] {
			for k := 0; k < numSamples; k++ {
				if j < len(targets[k]) {
					XTY[i][j] += states[k][i] * targets[k][j]
				}
			}
		}
	}

	// Solve (X^T @ X + λI) @ Wout^T = X^T @ Y
	// Using simple iterative method for this demo
	for i := 0; i < rl.OutputDim; i++ {
		// Extract column of XTY
		b := make([]float64, rl.NumNodes)
		for j := range b {
			b[j] = XTY[j][i]
		}

		// Solve using conjugate gradient (simplified)
		x := make([]float64, rl.NumNodes)
		for iter := 0; iter < 100; iter++ {
			for j := range x {
				sum := b[j]
				for k := range x {
					if k != j {
						sum -= XTX[j][k] * x[k]
					}
				}
				x[j] = sum / XTX[j][j]
			}
		}

		// Store in Wout
		for j := range x {
			rl.Wout[i][j] = x[j]
		}
	}
}

// Helper functions

func initWeightsXavier(fanIn, fanOut int) [][]float64 {
	w := make([][]float64, fanIn)
	stddev := math.Sqrt(2.0 / float64(fanIn+fanOut))
	for i := range w {
		w[i] = make([]float64, fanOut)
		for j := range w[i] {
			w[i][j] = rand.NormFloat64() * stddev
		}
	}
	return w
}

func addResidual(x, residual [][]float64) [][]float64 {
	output := make([][]float64, len(x))
	for i := range output {
		output[i] = make([]float64, len(x[i]))
		for j := range output[i] {
			output[i][j] = x[i][j]
			if i < len(residual) && j < len(residual[i]) {
				output[i][j] += residual[i][j]
			}
		}
	}
	return output
}

// GELU implements Gaussian Error Linear Unit activation.
// GELU(x) ≈ 0.5 * x * (1 + tanh(sqrt(2/π) * (x + 0.044715 * x³)))
func GELU(x float64) float64 {
	return 0.5 * x * (1 + math.Tanh(math.Sqrt(2/math.Pi)*(x+0.044715*x*x*x)))
}

// GELUVec applies GELU element-wise.
func GELUVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = GELU(v)
	}
	return result
}

// SiLU implements Sigmoid Linear Unit (Swish) activation.
// SiLU(x) = x * sigmoid(x)
func SiLU(x float64) float64 {
	return x / (1 + math.Exp(-x))
}

// SiLUVec applies SiLU element-wise.
func SiLUVec(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = SiLU(v)
	}
	return result
}
