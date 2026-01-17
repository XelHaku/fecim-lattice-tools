// Package layers provides neural network layer implementations for CIM simulation.
// recurrent.go implements recurrent layers (LSTM, GRU) for sequence modeling.
package layers

import (
	"math"
)

// =============================================================================
// LSTM (Long Short-Term Memory)
// =============================================================================

// LSTMConfig holds configuration for LSTM layer
type LSTMConfig struct {
	InputSize    int     // Input feature dimension
	HiddenSize   int     // Hidden state dimension
	NumLayers    int     // Number of stacked LSTM layers
	Bidirectional bool   // Use bidirectional LSTM
	DropoutRate  float64 // Dropout between layers (not on last)
	UseBias      bool    // Include bias terms
}

// DefaultLSTMConfig returns default LSTM configuration
func DefaultLSTMConfig(inputSize, hiddenSize int) *LSTMConfig {
	return &LSTMConfig{
		InputSize:     inputSize,
		HiddenSize:    hiddenSize,
		NumLayers:     1,
		Bidirectional: false,
		DropoutRate:   0.0,
		UseBias:       true,
	}
}

// LSTMCell implements a single LSTM cell
// Gates: forget (f), input (i), cell (g), output (o)
// Equations:
//   f_t = sigmoid(W_f * [h_{t-1}, x_t] + b_f)
//   i_t = sigmoid(W_i * [h_{t-1}, x_t] + b_i)
//   g_t = tanh(W_g * [h_{t-1}, x_t] + b_g)
//   o_t = sigmoid(W_o * [h_{t-1}, x_t] + b_o)
//   c_t = f_t * c_{t-1} + i_t * g_t
//   h_t = o_t * tanh(c_t)
type LSTMCell struct {
	config     *LSTMConfig

	// Weight matrices: [hiddenSize, inputSize + hiddenSize]
	Wf [][]float64 // Forget gate weights
	Wi [][]float64 // Input gate weights
	Wg [][]float64 // Cell gate weights
	Wo [][]float64 // Output gate weights

	// Bias vectors: [hiddenSize]
	Bf []float64
	Bi []float64
	Bg []float64
	Bo []float64

	// Hardware-friendly activation option
	UseHardActivations bool
}

// NewLSTMCell creates a new LSTM cell
func NewLSTMCell(config *LSTMConfig) *LSTMCell {
	inputPlusHidden := config.InputSize + config.HiddenSize

	cell := &LSTMCell{
		config: config,
		Wf:     make([][]float64, config.HiddenSize),
		Wi:     make([][]float64, config.HiddenSize),
		Wg:     make([][]float64, config.HiddenSize),
		Wo:     make([][]float64, config.HiddenSize),
	}

	// Initialize weight matrices
	for i := 0; i < config.HiddenSize; i++ {
		cell.Wf[i] = make([]float64, inputPlusHidden)
		cell.Wi[i] = make([]float64, inputPlusHidden)
		cell.Wg[i] = make([]float64, inputPlusHidden)
		cell.Wo[i] = make([]float64, inputPlusHidden)
	}

	// Initialize biases if enabled
	if config.UseBias {
		cell.Bf = make([]float64, config.HiddenSize)
		cell.Bi = make([]float64, config.HiddenSize)
		cell.Bg = make([]float64, config.HiddenSize)
		cell.Bo = make([]float64, config.HiddenSize)

		// Initialize forget gate bias to 1.0 (common practice)
		for i := range cell.Bf {
			cell.Bf[i] = 1.0
		}
	}

	return cell
}

// Forward performs one step of LSTM
// x: input at time t [inputSize]
// hPrev: previous hidden state [hiddenSize]
// cPrev: previous cell state [hiddenSize]
// Returns: (h_t, c_t)
func (l *LSTMCell) Forward(x, hPrev, cPrev []float64) ([]float64, []float64) {
	hiddenSize := l.config.HiddenSize

	// Concatenate input and previous hidden state
	concat := append(hPrev, x...)

	// Compute gates
	ft := l.computeGate(l.Wf, l.Bf, concat, true)  // forget gate (sigmoid)
	it := l.computeGate(l.Wi, l.Bi, concat, true)  // input gate (sigmoid)
	gt := l.computeGate(l.Wg, l.Bg, concat, false) // cell gate (tanh)
	ot := l.computeGate(l.Wo, l.Bo, concat, true)  // output gate (sigmoid)

	// Update cell state: c_t = f_t * c_{t-1} + i_t * g_t
	ct := make([]float64, hiddenSize)
	for i := 0; i < hiddenSize; i++ {
		ct[i] = ft[i]*cPrev[i] + it[i]*gt[i]
	}

	// Compute hidden state: h_t = o_t * tanh(c_t)
	ht := make([]float64, hiddenSize)
	for i := 0; i < hiddenSize; i++ {
		if l.UseHardActivations {
			ht[i] = ot[i] * hardTanh(ct[i])
		} else {
			ht[i] = ot[i] * math.Tanh(ct[i])
		}
	}

	return ht, ct
}

// computeGate computes a single gate activation
func (l *LSTMCell) computeGate(W [][]float64, b, input []float64, useSigmoid bool) []float64 {
	hiddenSize := l.config.HiddenSize
	output := make([]float64, hiddenSize)

	// Matrix-vector multiplication
	for i := 0; i < hiddenSize; i++ {
		sum := 0.0
		for j := 0; j < len(input); j++ {
			sum += W[i][j] * input[j]
		}
		if l.config.UseBias && b != nil {
			sum += b[i]
		}

		// Apply activation
		if useSigmoid {
			if l.UseHardActivations {
				output[i] = hardSigmoid(sum)
			} else {
				output[i] = sigmoid(sum)
			}
		} else {
			if l.UseHardActivations {
				output[i] = hardTanh(sum)
			} else {
				output[i] = math.Tanh(sum)
			}
		}
	}

	return output
}

// ForwardSequence processes an entire sequence
// inputs: [seqLen][inputSize]
// Returns: outputs [seqLen][hiddenSize], final (h, c)
func (l *LSTMCell) ForwardSequence(inputs [][]float64) ([][]float64, []float64, []float64) {
	seqLen := len(inputs)
	hiddenSize := l.config.HiddenSize

	// Initialize states
	h := make([]float64, hiddenSize)
	c := make([]float64, hiddenSize)

	outputs := make([][]float64, seqLen)

	for t := 0; t < seqLen; t++ {
		h, c = l.Forward(inputs[t], h, c)
		outputs[t] = make([]float64, hiddenSize)
		copy(outputs[t], h)
	}

	return outputs, h, c
}

// =============================================================================
// GRU (Gated Recurrent Unit)
// =============================================================================

// GRUConfig holds configuration for GRU layer
type GRUConfig struct {
	InputSize     int     // Input feature dimension
	HiddenSize    int     // Hidden state dimension
	NumLayers     int     // Number of stacked GRU layers
	Bidirectional bool    // Use bidirectional GRU
	DropoutRate   float64 // Dropout between layers
	UseBias       bool    // Include bias terms
	ResetAfter    bool    // Reset gate applied after matrix mult (cuDNN style)
}

// DefaultGRUConfig returns default GRU configuration
func DefaultGRUConfig(inputSize, hiddenSize int) *GRUConfig {
	return &GRUConfig{
		InputSize:     inputSize,
		HiddenSize:    hiddenSize,
		NumLayers:     1,
		Bidirectional: false,
		DropoutRate:   0.0,
		UseBias:       true,
		ResetAfter:    true,
	}
}

// GRUCell implements a single GRU cell
// Gates: reset (r), update (z)
// Equations:
//   r_t = sigmoid(W_r * [h_{t-1}, x_t] + b_r)
//   z_t = sigmoid(W_z * [h_{t-1}, x_t] + b_z)
//   n_t = tanh(W_n * x_t + b_n + r_t * (U_n * h_{t-1} + b_nh))
//   h_t = (1 - z_t) * n_t + z_t * h_{t-1}
type GRUCell struct {
	config *GRUConfig

	// Weight matrices for input: [hiddenSize, inputSize]
	Wr [][]float64 // Reset gate input weights
	Wz [][]float64 // Update gate input weights
	Wn [][]float64 // New gate input weights

	// Weight matrices for hidden: [hiddenSize, hiddenSize]
	Ur [][]float64 // Reset gate hidden weights
	Uz [][]float64 // Update gate hidden weights
	Un [][]float64 // New gate hidden weights

	// Bias vectors: [hiddenSize]
	Br  []float64 // Reset gate bias (input)
	Bz  []float64 // Update gate bias (input)
	Bn  []float64 // New gate bias (input)
	Brh []float64 // Reset gate bias (hidden)
	Bzh []float64 // Update gate bias (hidden)
	Bnh []float64 // New gate bias (hidden)

	// Hardware-friendly activation option
	UseHardActivations bool
}

// NewGRUCell creates a new GRU cell
func NewGRUCell(config *GRUConfig) *GRUCell {
	cell := &GRUCell{
		config: config,
		Wr:     make([][]float64, config.HiddenSize),
		Wz:     make([][]float64, config.HiddenSize),
		Wn:     make([][]float64, config.HiddenSize),
		Ur:     make([][]float64, config.HiddenSize),
		Uz:     make([][]float64, config.HiddenSize),
		Un:     make([][]float64, config.HiddenSize),
	}

	// Initialize weight matrices
	for i := 0; i < config.HiddenSize; i++ {
		cell.Wr[i] = make([]float64, config.InputSize)
		cell.Wz[i] = make([]float64, config.InputSize)
		cell.Wn[i] = make([]float64, config.InputSize)
		cell.Ur[i] = make([]float64, config.HiddenSize)
		cell.Uz[i] = make([]float64, config.HiddenSize)
		cell.Un[i] = make([]float64, config.HiddenSize)
	}

	// Initialize biases if enabled
	if config.UseBias {
		cell.Br = make([]float64, config.HiddenSize)
		cell.Bz = make([]float64, config.HiddenSize)
		cell.Bn = make([]float64, config.HiddenSize)
		cell.Brh = make([]float64, config.HiddenSize)
		cell.Bzh = make([]float64, config.HiddenSize)
		cell.Bnh = make([]float64, config.HiddenSize)
	}

	return cell
}

// Forward performs one step of GRU
// x: input at time t [inputSize]
// hPrev: previous hidden state [hiddenSize]
// Returns: h_t
func (g *GRUCell) Forward(x, hPrev []float64) []float64 {
	hiddenSize := g.config.HiddenSize

	// Compute reset and update gates
	rt := make([]float64, hiddenSize)
	zt := make([]float64, hiddenSize)

	for i := 0; i < hiddenSize; i++ {
		// r_t = sigmoid(W_r * x + U_r * h + b_r)
		rSum := g.Br[i] + g.Brh[i]
		for j := 0; j < g.config.InputSize; j++ {
			rSum += g.Wr[i][j] * x[j]
		}
		for j := 0; j < hiddenSize; j++ {
			rSum += g.Ur[i][j] * hPrev[j]
		}

		// z_t = sigmoid(W_z * x + U_z * h + b_z)
		zSum := g.Bz[i] + g.Bzh[i]
		for j := 0; j < g.config.InputSize; j++ {
			zSum += g.Wz[i][j] * x[j]
		}
		for j := 0; j < hiddenSize; j++ {
			zSum += g.Uz[i][j] * hPrev[j]
		}

		if g.UseHardActivations {
			rt[i] = hardSigmoid(rSum)
			zt[i] = hardSigmoid(zSum)
		} else {
			rt[i] = sigmoid(rSum)
			zt[i] = sigmoid(zSum)
		}
	}

	// Compute new gate with reset applied
	nt := make([]float64, hiddenSize)
	for i := 0; i < hiddenSize; i++ {
		nSum := g.Bn[i]
		for j := 0; j < g.config.InputSize; j++ {
			nSum += g.Wn[i][j] * x[j]
		}

		// Apply reset gate: r_t * (U_n * h + b_nh)
		hiddenSum := g.Bnh[i]
		for j := 0; j < hiddenSize; j++ {
			hiddenSum += g.Un[i][j] * hPrev[j]
		}
		nSum += rt[i] * hiddenSum

		if g.UseHardActivations {
			nt[i] = hardTanh(nSum)
		} else {
			nt[i] = math.Tanh(nSum)
		}
	}

	// Compute output: h_t = (1 - z_t) * n_t + z_t * h_{t-1}
	ht := make([]float64, hiddenSize)
	for i := 0; i < hiddenSize; i++ {
		ht[i] = (1-zt[i])*nt[i] + zt[i]*hPrev[i]
	}

	return ht
}

// ForwardSequence processes an entire sequence
// inputs: [seqLen][inputSize]
// Returns: outputs [seqLen][hiddenSize], final h
func (g *GRUCell) ForwardSequence(inputs [][]float64) ([][]float64, []float64) {
	seqLen := len(inputs)
	hiddenSize := g.config.HiddenSize

	// Initialize state
	h := make([]float64, hiddenSize)

	outputs := make([][]float64, seqLen)

	for t := 0; t < seqLen; t++ {
		h = g.Forward(inputs[t], h)
		outputs[t] = make([]float64, hiddenSize)
		copy(outputs[t], h)
	}

	return outputs, h
}

// =============================================================================
// Bidirectional RNN
// =============================================================================

// BiRNN wraps a recurrent cell for bidirectional processing
type BiRNN struct {
	ForwardCell  interface{} // *LSTMCell or *GRUCell
	BackwardCell interface{}
	CellType     string // "lstm" or "gru"
}

// NewBiLSTM creates a bidirectional LSTM
func NewBiLSTM(config *LSTMConfig) *BiRNN {
	return &BiRNN{
		ForwardCell:  NewLSTMCell(config),
		BackwardCell: NewLSTMCell(config),
		CellType:     "lstm",
	}
}

// NewBiGRU creates a bidirectional GRU
func NewBiGRU(config *GRUConfig) *BiRNN {
	return &BiRNN{
		ForwardCell:  NewGRUCell(config),
		BackwardCell: NewGRUCell(config),
		CellType:     "gru",
	}
}

// ForwardSequence processes sequence bidirectionally
// Returns concatenated [forward || backward] outputs
func (b *BiRNN) ForwardSequence(inputs [][]float64) [][]float64 {
	seqLen := len(inputs)

	var fwdOutputs, bwdOutputs [][]float64

	switch b.CellType {
	case "lstm":
		fwdCell := b.ForwardCell.(*LSTMCell)
		bwdCell := b.BackwardCell.(*LSTMCell)

		// Forward pass
		fwdOutputs, _, _ = fwdCell.ForwardSequence(inputs)

		// Backward pass (reverse inputs)
		reversedInputs := reverseSequence(inputs)
		bwdOutputs, _, _ = bwdCell.ForwardSequence(reversedInputs)
		bwdOutputs = reverseSequence(bwdOutputs) // Un-reverse outputs

	case "gru":
		fwdCell := b.ForwardCell.(*GRUCell)
		bwdCell := b.BackwardCell.(*GRUCell)

		fwdOutputs, _ = fwdCell.ForwardSequence(inputs)

		reversedInputs := reverseSequence(inputs)
		bwdOutputs, _ = bwdCell.ForwardSequence(reversedInputs)
		bwdOutputs = reverseSequence(bwdOutputs)
	}

	// Concatenate forward and backward outputs
	outputs := make([][]float64, seqLen)
	for t := 0; t < seqLen; t++ {
		outputs[t] = append(fwdOutputs[t], bwdOutputs[t]...)
	}

	return outputs
}

// =============================================================================
// Crossbar-Optimized Recurrent Layers
// =============================================================================

// CrossbarLSTMConfig extends LSTM config for crossbar mapping
type CrossbarLSTMConfig struct {
	*LSTMConfig
	TileSize     int  // Crossbar tile dimension
	WeightBits   int  // Weight quantization bits
	InputBits    int  // Input (PWM) bits
	OutputBits   int  // ADC output bits
	FuseGates    bool // Fuse all 4 gates into single crossbar
}

// DefaultCrossbarLSTMConfig returns crossbar-optimized LSTM config
func DefaultCrossbarLSTMConfig(inputSize, hiddenSize int) *CrossbarLSTMConfig {
	return &CrossbarLSTMConfig{
		LSTMConfig: DefaultLSTMConfig(inputSize, hiddenSize),
		TileSize:   64,
		WeightBits: 6,
		InputBits:  8,
		OutputBits: 6,
		FuseGates:  true,
	}
}

// CrossbarLSTM implements LSTM optimized for crossbar execution
type CrossbarLSTM struct {
	config *CrossbarLSTMConfig
	cell   *LSTMCell

	// Crossbar tiles for fused gates: [4*hiddenSize, inputSize + hiddenSize]
	FusedWeightTiles [][][]float64
	NumTiles         int
}

// NewCrossbarLSTM creates a crossbar-optimized LSTM
func NewCrossbarLSTM(config *CrossbarLSTMConfig) *CrossbarLSTM {
	cell := NewLSTMCell(config.LSTMConfig)
	cell.UseHardActivations = true // Use hardware-friendly activations

	cl := &CrossbarLSTM{
		config: config,
		cell:   cell,
	}

	// Create fused weight tiles
	if config.FuseGates {
		cl.createFusedTiles()
	}

	return cl
}

// createFusedTiles creates crossbar tiles for fused gate computation
func (cl *CrossbarLSTM) createFusedTiles() {
	hiddenSize := cl.config.HiddenSize
	inputPlusHidden := cl.config.InputSize + cl.config.HiddenSize
	tileSize := cl.config.TileSize

	// Fused weight matrix: [4*hiddenSize, inputSize + hiddenSize]
	// Rows: [Wf; Wi; Wg; Wo]
	fusedRows := 4 * hiddenSize
	fusedCols := inputPlusHidden

	// Calculate number of tiles needed
	tilesRow := (fusedRows + tileSize - 1) / tileSize
	tilesCol := (fusedCols + tileSize - 1) / tileSize
	cl.NumTiles = tilesRow * tilesCol

	// Create tile structure
	cl.FusedWeightTiles = make([][][]float64, tilesRow*tilesCol)

	tileIdx := 0
	for tr := 0; tr < tilesRow; tr++ {
		for tc := 0; tc < tilesCol; tc++ {
			tile := make([][]float64, tileSize)
			for i := 0; i < tileSize; i++ {
				tile[i] = make([]float64, tileSize)
				globalRow := tr*tileSize + i
				for j := 0; j < tileSize; j++ {
					globalCol := tc*tileSize + j
					if globalRow < fusedRows && globalCol < fusedCols {
						// Map to appropriate gate
						gateIdx := globalRow / hiddenSize
						gateRow := globalRow % hiddenSize
						var weights [][]float64
						switch gateIdx {
						case 0:
							weights = cl.cell.Wf
						case 1:
							weights = cl.cell.Wi
						case 2:
							weights = cl.cell.Wg
						case 3:
							weights = cl.cell.Wo
						}
						if gateRow < len(weights) && globalCol < len(weights[gateRow]) {
							tile[i][j] = weights[gateRow][globalCol]
						}
					}
				}
			}
			cl.FusedWeightTiles[tileIdx] = tile
			tileIdx++
		}
	}
}

// GetCrossbarMapping returns tile mapping information
func (cl *CrossbarLSTM) GetCrossbarMapping() map[string]interface{} {
	hiddenSize := cl.config.HiddenSize
	inputPlusHidden := cl.config.InputSize + cl.config.HiddenSize

	return map[string]interface{}{
		"fused_matrix_rows":    4 * hiddenSize,
		"fused_matrix_cols":    inputPlusHidden,
		"tile_size":            cl.config.TileSize,
		"num_tiles":            cl.NumTiles,
		"weight_bits":          cl.config.WeightBits,
		"total_weights":        4 * hiddenSize * inputPlusHidden,
		"crossbar_utilization": cl.computeUtilization(),
	}
}

// computeUtilization calculates crossbar utilization
func (cl *CrossbarLSTM) computeUtilization() float64 {
	hiddenSize := cl.config.HiddenSize
	inputPlusHidden := cl.config.InputSize + cl.config.HiddenSize
	tileSize := cl.config.TileSize

	totalWeights := float64(4 * hiddenSize * inputPlusHidden)
	totalSlots := float64(cl.NumTiles * tileSize * tileSize)

	return totalWeights / totalSlots
}

// =============================================================================
// Minimal RNN Cell (for reservoir computing)
// =============================================================================

// MinimalRNNCell implements a simple RNN: h_t = tanh(W_h * h_{t-1} + W_x * x_t + b)
type MinimalRNNCell struct {
	InputSize  int
	HiddenSize int
	Wh         [][]float64 // Hidden-to-hidden weights
	Wx         [][]float64 // Input-to-hidden weights
	B          []float64   // Bias
	UseHard    bool        // Use hardware-friendly tanh
}

// NewMinimalRNN creates a simple RNN cell
func NewMinimalRNN(inputSize, hiddenSize int) *MinimalRNNCell {
	cell := &MinimalRNNCell{
		InputSize:  inputSize,
		HiddenSize: hiddenSize,
		Wh:         make([][]float64, hiddenSize),
		Wx:         make([][]float64, hiddenSize),
		B:          make([]float64, hiddenSize),
	}

	for i := 0; i < hiddenSize; i++ {
		cell.Wh[i] = make([]float64, hiddenSize)
		cell.Wx[i] = make([]float64, inputSize)
	}

	return cell
}

// Forward performs one RNN step
func (r *MinimalRNNCell) Forward(x, hPrev []float64) []float64 {
	h := make([]float64, r.HiddenSize)

	for i := 0; i < r.HiddenSize; i++ {
		sum := r.B[i]
		for j := 0; j < r.InputSize; j++ {
			sum += r.Wx[i][j] * x[j]
		}
		for j := 0; j < r.HiddenSize; j++ {
			sum += r.Wh[i][j] * hPrev[j]
		}

		if r.UseHard {
			h[i] = hardTanh(sum)
		} else {
			h[i] = math.Tanh(sum)
		}
	}

	return h
}

// ForwardSequence processes entire sequence
func (r *MinimalRNNCell) ForwardSequence(inputs [][]float64) ([][]float64, []float64) {
	seqLen := len(inputs)
	h := make([]float64, r.HiddenSize)
	outputs := make([][]float64, seqLen)

	for t := 0; t < seqLen; t++ {
		h = r.Forward(inputs[t], h)
		outputs[t] = make([]float64, r.HiddenSize)
		copy(outputs[t], h)
	}

	return outputs, h
}

// =============================================================================
// Sequence-to-Sequence Components
// =============================================================================

// Seq2SeqEncoder wraps recurrent cell as encoder
type Seq2SeqEncoder struct {
	Cell     interface{} // *LSTMCell, *GRUCell, or *MinimalRNNCell
	CellType string
}

// NewSeq2SeqEncoder creates an encoder
func NewSeq2SeqEncoder(cellType string, inputSize, hiddenSize int) *Seq2SeqEncoder {
	encoder := &Seq2SeqEncoder{CellType: cellType}

	switch cellType {
	case "lstm":
		encoder.Cell = NewLSTMCell(DefaultLSTMConfig(inputSize, hiddenSize))
	case "gru":
		encoder.Cell = NewGRUCell(DefaultGRUConfig(inputSize, hiddenSize))
	default:
		encoder.Cell = NewMinimalRNN(inputSize, hiddenSize)
	}

	return encoder
}

// Encode processes input sequence and returns context
func (e *Seq2SeqEncoder) Encode(inputs [][]float64) (context []float64, cellState []float64) {
	switch e.CellType {
	case "lstm":
		cell := e.Cell.(*LSTMCell)
		_, h, c := cell.ForwardSequence(inputs)
		return h, c
	case "gru":
		cell := e.Cell.(*GRUCell)
		_, h := cell.ForwardSequence(inputs)
		return h, nil
	default:
		cell := e.Cell.(*MinimalRNNCell)
		_, h := cell.ForwardSequence(inputs)
		return h, nil
	}
}

// =============================================================================
// Helper Functions
// =============================================================================

// sigmoid activation
func sigmoid(x float64) float64 {
	return 1.0 / (1.0 + math.Exp(-x))
}

// hardSigmoid: hardware-friendly sigmoid approximation
func hardSigmoid(x float64) float64 {
	result := 0.2*x + 0.5
	if result < 0 {
		return 0
	}
	if result > 1 {
		return 1
	}
	return result
}

// hardTanh: hardware-friendly tanh approximation
func hardTanh(x float64) float64 {
	if x < -1 {
		return -1
	}
	if x > 1 {
		return 1
	}
	return x
}

// reverseSequence reverses a sequence
func reverseSequence(seq [][]float64) [][]float64 {
	n := len(seq)
	reversed := make([][]float64, n)
	for i := 0; i < n; i++ {
		reversed[i] = seq[n-1-i]
	}
	return reversed
}

// GetRecurrentParamCount returns total parameter count
func GetRecurrentParamCount(cellType string, inputSize, hiddenSize int) int {
	switch cellType {
	case "lstm":
		// 4 gates × (inputSize + hiddenSize) × hiddenSize + 4 × hiddenSize (bias)
		return 4 * (inputSize + hiddenSize) * hiddenSize + 4*hiddenSize
	case "gru":
		// 3 gates × (inputSize × hiddenSize + hiddenSize × hiddenSize) + 6 × hiddenSize
		return 3*inputSize*hiddenSize + 3*hiddenSize*hiddenSize + 6*hiddenSize
	default:
		// Simple RNN: (inputSize + hiddenSize) × hiddenSize + hiddenSize
		return (inputSize+hiddenSize)*hiddenSize + hiddenSize
	}
}

// GetRecurrentFLOPs returns FLOPs per timestep
func GetRecurrentFLOPs(cellType string, inputSize, hiddenSize int) int {
	switch cellType {
	case "lstm":
		// 4 gates MVMs + element-wise ops
		mvmFlops := 4 * 2 * (inputSize + hiddenSize) * hiddenSize
		elementWise := 5 * hiddenSize // c_t and h_t computation
		return mvmFlops + elementWise
	case "gru":
		// 3 gates + interpolation
		mvmFlops := 3*2*inputSize*hiddenSize + 3*2*hiddenSize*hiddenSize
		elementWise := 4 * hiddenSize
		return mvmFlops + elementWise
	default:
		return 2*(inputSize+hiddenSize)*hiddenSize + hiddenSize
	}
}
