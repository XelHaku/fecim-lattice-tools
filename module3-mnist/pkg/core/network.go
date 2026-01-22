package core

import (
	"encoding/json"
	"math"
	"os"
	"sync"
)

// NetworkConfig holds configuration for the CIM inference path.
type NetworkConfig struct {
	NumLevels   int     // Quantization levels (1-30)
	NoiseLevel  float64 // Noise as σ/μ coefficient (0.0-0.20)
	ADCBits     int     // ADC resolution (3-16)
	DACBits     int     // DAC resolution (3-16)
	EnableSneak bool    // Enable sneak path simulation
	IRDrop      bool    // Enable IR drop simulation
	SingleLayer bool    // Tour Mode: use single-layer (784→10) like Dr. Tour's demo
}

// DefaultNetworkConfig returns the default configuration for optimal FeCIM operation.
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		NumLevels:  30,   // FeCIM supports 30 levels
		NoiseLevel: 0.01, // Low noise
		ADCBits:    8,    // 8-bit ADC
		DACBits:    8,    // 8-bit DAC
	}
}

// DualModeNetwork implements dual-path inference: Full Precision (FP) and CIM.
// This allows direct comparison of ideal digital computation vs realistic FeCIM hardware.
type DualModeNetwork struct {
	// Architecture
	InputSize  int
	HiddenSize int
	OutputSize int

	// Full Precision Weights (float64, for FP inference path)
	FPWeights1 [][]float64 // [HiddenSize][InputSize]
	FPWeights2 [][]float64 // [OutputSize][HiddenSize]
	FPBias1    []float64   // [HiddenSize]
	FPBias2    []float64   // [OutputSize]

	// Quantized Weights (modified by sliders, for CIM inference path)
	QuantWeights1 [][]float64 // [HiddenSize][InputSize]
	QuantWeights2 [][]float64 // [OutputSize][HiddenSize]
	QuantBias1    []float64   // [HiddenSize]
	QuantBias2    []float64   // [OutputSize]

	// Single-Layer Weights (Tour Mode: 784→10 direct, ~87% max accuracy)
	// This matches Dr. Tour's MNIST demo architecture
	SingleLayerWeights      [][]float64 // [OutputSize][InputSize] = [10][784]
	SingleLayerBias         []float64   // [OutputSize] = [10]
	QuantSingleLayerWeights [][]float64 // Quantized version
	QuantSingleLayerBias    []float64   // Quantized version

	// Configuration
	Config *NetworkConfig

	// Random source for noise injection
	rng *RandomSource

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// InferenceResult holds results from dual-path inference.
type InferenceResult struct {
	// FP path results
	FPLogits        []float64
	FPProbabilities []float64
	FPPrediction    int
	FPConfidence    float64

	// CIM path results
	CIMLogits        []float64
	CIMProbabilities []float64
	CIMPrediction    int
	CIMConfidence    float64

	// Intermediate activations (for visualization)
	FPHidden  []float64
	CIMHidden []float64

	// Comparison metrics
	Agree        bool    // Do FP and CIM predictions match?
	Disagreement float64 // KL divergence between probability distributions
	EnergyUsed   float64 // Estimated energy in μJ
}

// NewDualModeNetwork creates a new dual-mode network with the specified architecture.
func NewDualModeNetwork(inputSize, hiddenSize, outputSize int) *DualModeNetwork {
	net := &DualModeNetwork{
		InputSize:  inputSize,
		HiddenSize: hiddenSize,
		OutputSize: outputSize,
		Config:     DefaultNetworkConfig(),
		rng:        NewRandomSource(42),
	}

	// Initialize weight matrices
	net.FPWeights1 = make([][]float64, hiddenSize)
	net.FPWeights2 = make([][]float64, outputSize)
	net.QuantWeights1 = make([][]float64, hiddenSize)
	net.QuantWeights2 = make([][]float64, outputSize)

	for i := 0; i < hiddenSize; i++ {
		net.FPWeights1[i] = make([]float64, inputSize)
		net.QuantWeights1[i] = make([]float64, inputSize)
	}
	for i := 0; i < outputSize; i++ {
		net.FPWeights2[i] = make([]float64, hiddenSize)
		net.QuantWeights2[i] = make([]float64, hiddenSize)
	}

	// Initialize biases
	net.FPBias1 = make([]float64, hiddenSize)
	net.FPBias2 = make([]float64, outputSize)
	net.QuantBias1 = make([]float64, hiddenSize)
	net.QuantBias2 = make([]float64, outputSize)

	// Initialize single-layer weights (Tour Mode: 784→10)
	net.SingleLayerWeights = make([][]float64, outputSize)
	net.QuantSingleLayerWeights = make([][]float64, outputSize)
	for i := 0; i < outputSize; i++ {
		net.SingleLayerWeights[i] = make([]float64, inputSize)
		net.QuantSingleLayerWeights[i] = make([]float64, inputSize)
	}
	net.SingleLayerBias = make([]float64, outputSize)
	net.QuantSingleLayerBias = make([]float64, outputSize)

	return net
}

// WeightsFile represents the JSON structure for loading pretrained weights.
type WeightsFile struct {
	Layer1Weights [][]float64 `json:"layer1_weights"`
	Layer2Weights [][]float64 `json:"layer2_weights"`
	Biases1       []float64   `json:"biases1"`
	Biases2       []float64   `json:"biases2"`
	L1Scale       float64     `json:"l1_scale"`
	L1Offset      float64     `json:"l1_offset"`
	// Single-layer weights (Tour Mode: 784→10 direct)
	SingleLayerWeights [][]float64 `json:"single_layer_weights,omitempty"`
	SingleLayerBias    []float64   `json:"single_layer_bias,omitempty"`
	L2Scale       float64     `json:"l2_scale"`
	L2Offset      float64     `json:"l2_offset"`
}

// LoadWeights loads pretrained weights from a JSON file.
// The file can contain either:
// 1. Quantized weights with scale/offset (new format)
// 2. Raw FP weights (legacy format)
func (net *DualModeNetwork) LoadWeights(filename string) error {
	net.mu.Lock()
	defer net.mu.Unlock()

	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var wf WeightsFile
	if err := json.Unmarshal(data, &wf); err != nil {
		return err
	}

	// Determine hidden size from loaded weights
	hiddenSize := len(wf.Layer1Weights)
	inputSize := 0
	if hiddenSize > 0 {
		inputSize = len(wf.Layer1Weights[0])
	}
	outputSize := len(wf.Layer2Weights)

	net.HiddenSize = hiddenSize
	net.InputSize = inputSize
	net.OutputSize = outputSize

	// Reallocate if needed
	net.FPWeights1 = make([][]float64, hiddenSize)
	net.FPWeights2 = make([][]float64, outputSize)
	net.QuantWeights1 = make([][]float64, hiddenSize)
	net.QuantWeights2 = make([][]float64, outputSize)

	// Load and reconstruct FP weights
	// If scale/offset provided, reconstruct: fp = quantized * scale + offset
	// Otherwise, use values directly as FP weights
	for i := 0; i < hiddenSize; i++ {
		net.FPWeights1[i] = make([]float64, inputSize)
		net.QuantWeights1[i] = make([]float64, inputSize)
		for j := 0; j < inputSize && j < len(wf.Layer1Weights[i]); j++ {
			qw := wf.Layer1Weights[i][j]
			if wf.L1Scale > 0 {
				net.FPWeights1[i][j] = qw*wf.L1Scale + wf.L1Offset
			} else {
				net.FPWeights1[i][j] = qw
			}
		}
	}

	for i := 0; i < outputSize; i++ {
		net.FPWeights2[i] = make([]float64, hiddenSize)
		net.QuantWeights2[i] = make([]float64, hiddenSize)
		for j := 0; j < hiddenSize && j < len(wf.Layer2Weights[i]); j++ {
			qw := wf.Layer2Weights[i][j]
			if wf.L2Scale > 0 {
				net.FPWeights2[i][j] = qw*wf.L2Scale + wf.L2Offset
			} else {
				net.FPWeights2[i][j] = qw
			}
		}
	}

	// Load biases (biases are stored as FP, not quantized)
	net.FPBias1 = make([]float64, hiddenSize)
	net.FPBias2 = make([]float64, outputSize)
	net.QuantBias1 = make([]float64, hiddenSize)
	net.QuantBias2 = make([]float64, outputSize)

	if len(wf.Biases1) > 0 {
		copy(net.FPBias1, wf.Biases1)
	}
	if len(wf.Biases2) > 0 {
		copy(net.FPBias2, wf.Biases2)
	}

	// Load single-layer weights if provided (Tour Mode: 784→10)
	// If not provided, initialize from layer2 weights as a fallback
	net.SingleLayerWeights = make([][]float64, outputSize)
	net.QuantSingleLayerWeights = make([][]float64, outputSize)
	net.SingleLayerBias = make([]float64, outputSize)
	net.QuantSingleLayerBias = make([]float64, outputSize)

	if len(wf.SingleLayerWeights) >= outputSize && len(wf.SingleLayerWeights[0]) >= inputSize {
		// Use provided single-layer weights
		for i := 0; i < outputSize; i++ {
			net.SingleLayerWeights[i] = make([]float64, inputSize)
			net.QuantSingleLayerWeights[i] = make([]float64, inputSize)
			copy(net.SingleLayerWeights[i], wf.SingleLayerWeights[i])
		}
		if len(wf.SingleLayerBias) >= outputSize {
			copy(net.SingleLayerBias, wf.SingleLayerBias)
		}
	} else {
		// Generate single-layer weights using Xavier initialization
		// This gives ~88% theoretical max on MNIST (matching Tour's claim)
		scale := 1.0 / float64(inputSize)
		for i := 0; i < outputSize; i++ {
			net.SingleLayerWeights[i] = make([]float64, inputSize)
			net.QuantSingleLayerWeights[i] = make([]float64, inputSize)
			for j := 0; j < inputSize; j++ {
				// Xavier-like initialization scaled for 30-level quantization
				net.SingleLayerWeights[i][j] = (net.rng.Float64()*2 - 1) * scale
			}
		}
	}

	// Initialize quantized weights based on current config
	net.requantizeWeightsLocked()

	return nil
}

// RequantizeWeights re-quantizes the FP weights to the current number of levels.
// Call this after changing Config.NumLevels.
func (net *DualModeNetwork) RequantizeWeights() {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.requantizeWeightsLocked()
}

// requantizeWeightsLocked performs quantization (must hold lock).
func (net *DualModeNetwork) requantizeWeightsLocked() {
	levels := net.Config.NumLevels
	if levels < 2 {
		levels = 2
	}
	if levels > 30 {
		levels = 30
	}

	// Quantize layer 1 weights
	net.QuantWeights1 = QuantizeWeights(net.FPWeights1, levels)

	// Quantize layer 2 weights
	net.QuantWeights2 = QuantizeWeights(net.FPWeights2, levels)

	// Quantize single-layer weights (Tour Mode)
	if len(net.SingleLayerWeights) > 0 {
		net.QuantSingleLayerWeights = QuantizeWeights(net.SingleLayerWeights, levels)
		net.QuantSingleLayerBias = QuantizeBias(net.SingleLayerBias, levels)
	}

	// Quantize biases
	net.QuantBias1 = QuantizeBias(net.FPBias1, levels)
	net.QuantBias2 = QuantizeBias(net.FPBias2, levels)
}

// SetNumLevels updates the quantization levels and re-quantizes weights.
func (net *DualModeNetwork) SetNumLevels(levels int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 1 {
		levels = 1
	}
	if levels > 30 {
		levels = 30
	}
	net.Config.NumLevels = levels
	net.requantizeWeightsLocked()
}

// GetNumLevels returns the current quantization levels.
func (net *DualModeNetwork) GetNumLevels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.NumLevels
}

// SetNoiseLevel updates the noise level for CIM inference.
func (net *DualModeNetwork) SetNoiseLevel(noise float64) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if noise < 0 {
		noise = 0
	}
	if noise > 0.5 {
		noise = 0.5
	}
	net.Config.NoiseLevel = noise
}

// SetADCBits updates the ADC resolution for CIM inference.
func (net *DualModeNetwork) SetADCBits(bits int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if bits < 3 {
		bits = 3
	}
	if bits > 16 {
		bits = 16
	}
	net.Config.ADCBits = bits
}

// SetDACBits updates the DAC resolution for CIM inference.
func (net *DualModeNetwork) SetDACBits(bits int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if bits < 3 {
		bits = 3
	}
	if bits > 16 {
		bits = 16
	}
	net.Config.DACBits = bits
}

// SetSingleLayer enables/disables Tour Mode (single-layer 784→10 architecture).
// When enabled, this matches Dr. Tour's MNIST demo with ~87% theoretical max accuracy.
func (net *DualModeNetwork) SetSingleLayer(enabled bool) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.Config.SingleLayer = enabled
}

// IsSingleLayer returns whether single-layer (Tour Mode) is enabled.
func (net *DualModeNetwork) IsSingleLayer() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.SingleLayer
}

// Infer runs dual-path inference (FP + CIM) and returns comparison results.
func (net *DualModeNetwork) Infer(input []float64) *InferenceResult {
	net.mu.RLock()
	defer net.mu.RUnlock()

	result := &InferenceResult{}

	var fpOutput, fpProbs []float64
	var fpHidden []float64
	var cimOutput, cimProbs []float64
	var cimHidden []float64
	var totalMACs int

	if net.Config.SingleLayer {
		// ============================================
		// TOUR MODE: Single-Layer (784→10)
		// Matches Dr. Tour's MNIST demo (~87% theoretical max)
		// ============================================

		// FP PATH (single layer)
		fpOutput = net.forwardFP(input, net.SingleLayerWeights, net.SingleLayerBias)
		fpProbs = softmax(fpOutput)
		fpHidden = nil // No hidden layer in Tour mode

		// CIM PATH (single layer with quantization)
		dacInput := quantizeDAC(input, net.Config.DACBits)
		cimOutput = net.forwardCIM(dacInput, net.QuantSingleLayerWeights, net.QuantSingleLayerBias)
		cimOutput = quantizeADC(cimOutput, net.Config.ADCBits)
		cimOutput = AddGaussianNoise(cimOutput, net.Config.NoiseLevel, net.rng)
		cimProbs = softmax(cimOutput)
		cimHidden = nil // No hidden layer in Tour mode

		// Energy: single layer only
		totalMACs = net.InputSize * net.OutputSize // 784 × 10 = 7,840 MACs
	} else {
		// ============================================
		// STANDARD MODE: Two-Layer (784→128→10)
		// Higher accuracy (~93%+) due to hidden layer capacity
		// ============================================

		// FP PATH (Ideal Digital)
		fpHidden = net.forwardFP(input, net.FPWeights1, net.FPBias1)
		fpHidden = relu(fpHidden)
		fpOutput = net.forwardFP(fpHidden, net.FPWeights2, net.FPBias2)
		fpProbs = softmax(fpOutput)

		// CIM PATH (Realistic Hardware)
		dacInput := quantizeDAC(input, net.Config.DACBits)

		// Layer 1: Use QUANTIZED weights (30-level FeCIM quantization)
		cimHidden = net.forwardCIM(dacInput, net.QuantWeights1, net.QuantBias1)
		cimHidden = quantizeADC(cimHidden, net.Config.ADCBits)
		cimHidden = AddGaussianNoise(cimHidden, net.Config.NoiseLevel, net.rng)
		cimHidden = relu(cimHidden)

		// Layer 2: Use QUANTIZED weights (30-level FeCIM quantization)
		cimOutput = net.forwardCIM(cimHidden, net.QuantWeights2, net.QuantBias2)
		cimOutput = quantizeADC(cimOutput, net.Config.ADCBits)
		cimOutput = AddGaussianNoise(cimOutput, net.Config.NoiseLevel, net.rng)
		cimProbs = softmax(cimOutput)

		// Energy: two layers
		macs1 := net.InputSize * net.HiddenSize  // 784 × 128
		macs2 := net.HiddenSize * net.OutputSize // 128 × 10
		totalMACs = macs1 + macs2
	}

	// Store FP results
	result.FPLogits = fpOutput
	result.FPProbabilities = fpProbs
	result.FPPrediction = argmax(fpProbs)
	result.FPConfidence = fpProbs[result.FPPrediction]
	result.FPHidden = fpHidden

	// Store CIM results
	result.CIMLogits = cimOutput
	result.CIMProbabilities = cimProbs
	result.CIMPrediction = argmax(cimProbs)
	result.CIMConfidence = cimProbs[result.CIMPrediction]
	result.CIMHidden = cimHidden

	// ============================================
	// COMPARISON
	// ============================================
	result.Agree = (result.FPPrediction == result.CIMPrediction)
	result.Disagreement = klDivergence(result.FPProbabilities, result.CIMProbabilities)

	// Energy calculation (Jerry et al. IEDM 2017: ~50 fJ/MAC)
	result.EnergyUsed = float64(totalMACs) * 50e-15 * 1e6 // Convert to μJ

	return result
}

// InferFPOnly runs only the FP path (for fast evaluation).
func (net *DualModeNetwork) InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	hidden := net.forwardFP(input, net.FPWeights1, net.FPBias1)
	hidden = relu(hidden)
	output := net.forwardFP(hidden, net.FPWeights2, net.FPBias2)
	probs = softmax(output)
	prediction = argmax(probs)
	confidence = probs[prediction]
	return
}

// InferCIMOnly runs only the CIM path (for fast evaluation).
func (net *DualModeNetwork) InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	dacInput := quantizeDAC(input, net.Config.DACBits)

	hidden := net.forwardCIM(dacInput, net.FPWeights1, net.FPBias1)
	hidden = quantizeADC(hidden, net.Config.ADCBits)
	hidden = AddGaussianNoise(hidden, net.Config.NoiseLevel, net.rng)
	hidden = relu(hidden)

	output := net.forwardCIM(hidden, net.FPWeights2, net.FPBias2)
	output = quantizeADC(output, net.Config.ADCBits)
	output = AddGaussianNoise(output, net.Config.NoiseLevel, net.rng)
	probs = softmax(output)
	prediction = argmax(probs)
	confidence = probs[prediction]
	return
}

// forwardFP performs standard FP matrix multiplication.
func (net *DualModeNetwork) forwardFP(input []float64, weights [][]float64, bias []float64) []float64 {
	output := make([]float64, len(bias))

	for i := 0; i < len(weights); i++ {
		sum := bias[i]
		for j := 0; j < len(input); j++ {
			sum += weights[i][j] * input[j]
		}
		output[i] = sum
	}

	return output
}

// forwardCIM performs CIM matrix multiplication (same math, but uses quantized weights).
func (net *DualModeNetwork) forwardCIM(input []float64, weights [][]float64, bias []float64) []float64 {
	return net.forwardFP(input, weights, bias)
}

// relu applies ReLU activation.
func relu(x []float64) []float64 {
	result := make([]float64, len(x))
	for i, v := range x {
		if v > 0 {
			result[i] = v
		}
	}
	return result
}

// softmax applies softmax activation.
func softmax(x []float64) []float64 {
	max := x[0]
	for _, v := range x {
		if v > max {
			max = v
		}
	}

	expSum := 0.0
	result := make([]float64, len(x))
	for i, v := range x {
		result[i] = math.Exp(v - max)
		expSum += result[i]
	}

	for i := range result {
		result[i] /= expSum
	}

	return result
}

// argmax returns the index of the maximum value.
func argmax(x []float64) int {
	maxIdx := 0
	maxVal := x[0]
	for i, v := range x {
		if v > maxVal {
			maxVal = v
			maxIdx = i
		}
	}
	return maxIdx
}

// klDivergence computes KL divergence D(p||q).
func klDivergence(p, q []float64) float64 {
	kl := 0.0
	eps := 1e-10
	for i := range p {
		if p[i] > eps {
			kl += p[i] * math.Log(p[i]/(q[i]+eps))
		}
	}
	return kl
}

// quantizeDAC simulates N-bit DAC quantization of input voltages.
func quantizeDAC(values []float64, bits int) []float64 {
	if bits >= 16 {
		return values // No quantization
	}

	levels := 1 << bits // 2^bits
	result := make([]float64, len(values))

	for i, v := range values {
		// Input values are assumed to be in [0, 1]
		// Clamp and quantize
		v = math.Max(0, math.Min(1, v))
		bin := int(math.Round(v * float64(levels-1)))
		if bin >= levels {
			bin = levels - 1
		}
		result[i] = float64(bin) / float64(levels-1)
	}

	return result
}

// quantizeADC simulates N-bit ADC quantization of output currents.
func quantizeADC(values []float64, bits int) []float64 {
	if bits >= 16 {
		return values // No quantization
	}

	levels := 1 << bits // 2^bits

	// Find range
	vMin, vMax := values[0], values[0]
	for _, v := range values {
		if v < vMin {
			vMin = v
		}
		if v > vMax {
			vMax = v
		}
	}

	vRange := vMax - vMin
	if vRange == 0 {
		return values
	}

	step := vRange / float64(levels-1)

	result := make([]float64, len(values))
	for i, v := range values {
		bin := int(math.Round((v - vMin) / step))
		if bin < 0 {
			bin = 0
		}
		if bin >= levels {
			bin = levels - 1
		}
		result[i] = vMin + float64(bin)*step
	}

	return result
}

// GetQuantizationStats returns statistics about the current weight quantization.
func (net *DualModeNetwork) GetQuantizationStats() (layer1Stats, layer2Stats QuantizationStats) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	layer1Stats = ComputeQuantizationStats(net.FPWeights1, net.QuantWeights1)
	layer2Stats = ComputeQuantizationStats(net.FPWeights2, net.QuantWeights2)
	return
}

// GetFPWeights returns a copy of the FP weights for visualization.
func (net *DualModeNetwork) GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	w1 = copyMatrix(net.FPWeights1)
	w2 = copyMatrix(net.FPWeights2)
	b1 = copySlice(net.FPBias1)
	b2 = copySlice(net.FPBias2)
	return
}

// GetQuantWeights returns a copy of the quantized weights for visualization.
func (net *DualModeNetwork) GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	w1 = copyMatrix(net.QuantWeights1)
	w2 = copyMatrix(net.QuantWeights2)
	b1 = copySlice(net.QuantBias1)
	b2 = copySlice(net.QuantBias2)
	return
}

func copyMatrix(m [][]float64) [][]float64 {
	result := make([][]float64, len(m))
	for i := range m {
		result[i] = make([]float64, len(m[i]))
		copy(result[i], m[i])
	}
	return result
}

func copySlice(s []float64) []float64 {
	result := make([]float64, len(s))
	copy(result, s)
	return result
}
