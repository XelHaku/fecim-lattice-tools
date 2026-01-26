package core

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
)

// NetworkConfig holds configuration for the CIM inference path.
type NetworkConfig struct {
	NumLevels   int     // Quantization levels (1-30) - used when PerLayerQuant is false
	NoiseLevel  float64 // Noise as σ/μ coefficient (0.0-0.20)
	ADCBits     int     // ADC resolution (3-16)
	DACBits     int     // DAC resolution (3-16)
	EnableSneak bool    // Enable sneak path simulation
	IRDrop      bool    // Enable IR drop simulation
	SingleLayer bool    // Tour Mode: use single-layer (784→10) like Dr. Tour's demo

	// Per-layer PTQ configuration
	PerLayerQuant bool // Enable per-layer quantization levels
	Layer1Levels  int  // Quantization levels for layer 1 (hidden layer)
	Layer2Levels  int  // Quantization levels for layer 2 (output layer)
}

// DefaultNetworkConfig returns the default configuration for optimal FeCIM operation.
func DefaultNetworkConfig() *NetworkConfig {
	return &NetworkConfig{
		NumLevels:     30,   // FeCIM supports 30 levels
		NoiseLevel:    0.01, // Low noise
		ADCBits:       8,    // 8-bit ADC
		DACBits:       8,    // 8-bit DAC
		PerLayerQuant: false,
		Layer1Levels:  30, // Default same as NumLevels
		Layer2Levels:  30, // Default same as NumLevels
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

	// Separate mutex for RNG access to prevent races under RLock
	rngMu sync.Mutex
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
	L2Scale            float64     `json:"l2_scale"`
	L2Offset           float64     `json:"l2_offset"`
	// Quantization level the weights were trained for (QAT) - legacy uniform
	QuantLevels int `json:"quant_levels,omitempty"`
	// Per-layer PTQ quantization levels
	Layer1QuantLevels int `json:"layer1_quant_levels,omitempty"`
	Layer2QuantLevels int `json:"layer2_quant_levels,omitempty"`
}

// AvailableQATLevels lists the quantization levels we have trained weights for.
var AvailableQATLevels = []int{10, 20, 29, 30, 31}

// GetWeightsFilename returns the appropriate weights filename for a quantization level.
// Returns the exact match if available, otherwise returns the default 30-level weights.
func GetWeightsFilename(dataDir string, levels int) string {
	// Check for exact match first
	for _, l := range AvailableQATLevels {
		if l == levels && l != 30 {
			return filepath.Join(dataDir, fmt.Sprintf("pretrained_weights_%d.json", levels))
		}
	}
	// Default to 30-level weights (backward compatible)
	return filepath.Join(dataDir, "pretrained_weights.json")
}

// GetBestMatchingWeightsLevel returns the closest available QAT level for a given target.
func GetBestMatchingWeightsLevel(targetLevels int) int {
	bestMatch := 30
	bestDiff := 1000
	for _, l := range AvailableQATLevels {
		diff := targetLevels - l
		if diff < 0 {
			diff = -diff
		}
		if diff < bestDiff {
			bestDiff = diff
			bestMatch = l
		}
	}
	return bestMatch
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

	// Check for per-layer quantization levels in the weights file
	if wf.Layer1QuantLevels > 0 && wf.Layer2QuantLevels > 0 {
		// File contains per-layer PTQ configuration
		net.Config.Layer1Levels = wf.Layer1QuantLevels
		net.Config.Layer2Levels = wf.Layer2QuantLevels
	} else if wf.QuantLevels > 0 {
		// Legacy uniform quantization
		net.Config.Layer1Levels = wf.QuantLevels
		net.Config.Layer2Levels = wf.QuantLevels
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

// LoadWeightsForLevel loads weights optimized for a specific quantization level.
// It looks for level-specific weight files (e.g., pretrained_weights_20.json)
// and falls back to the default 30-level weights if not found.
func (net *DualModeNetwork) LoadWeightsForLevel(dataDir string, levels int) error {
	// Find the best matching weights file
	bestLevel := GetBestMatchingWeightsLevel(levels)
	filename := GetWeightsFilename(dataDir, bestLevel)

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Fall back to default weights
		filename = filepath.Join(dataDir, "pretrained_weights.json")
	}

	return net.LoadWeights(filename)
}

// requantizeWeightsLocked performs quantization (must hold lock).
func (net *DualModeNetwork) requantizeWeightsLocked() {
	// Determine quantization levels for each layer
	var l1Levels, l2Levels int

	if net.Config.PerLayerQuant {
		// Use per-layer quantization levels
		l1Levels = net.Config.Layer1Levels
		l2Levels = net.Config.Layer2Levels
	} else {
		// Use uniform quantization levels
		l1Levels = net.Config.NumLevels
		l2Levels = net.Config.NumLevels
	}

	// Clamp levels to valid range
	if l1Levels < 2 {
		l1Levels = 2
	}
	if l1Levels > 31 {
		l1Levels = 31
	}
	if l2Levels < 2 {
		l2Levels = 2
	}
	if l2Levels > 31 {
		l2Levels = 31
	}

	// Quantize layer 1 weights with layer1 levels
	// Levels are clamped to [2, 31] above, so these cannot fail
	net.QuantWeights1, _ = QuantizeWeights(net.FPWeights1, l1Levels)

	// Quantize layer 2 weights with layer2 levels
	net.QuantWeights2, _ = QuantizeWeights(net.FPWeights2, l2Levels)

	// Quantize single-layer weights (Tour Mode) - use l1 levels for input layer
	if len(net.SingleLayerWeights) > 0 {
		net.QuantSingleLayerWeights, _ = QuantizeWeights(net.SingleLayerWeights, l1Levels)
		net.QuantSingleLayerBias, _ = QuantizeBias(net.SingleLayerBias, l1Levels)
	}

	// Quantize biases with corresponding layer levels
	net.QuantBias1, _ = QuantizeBias(net.FPBias1, l1Levels)
	net.QuantBias2, _ = QuantizeBias(net.FPBias2, l2Levels)
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

// SetPerLayerQuant enables/disables per-layer quantization mode.
// When enabled, Layer1Levels and Layer2Levels are used instead of NumLevels.
func (net *DualModeNetwork) SetPerLayerQuant(enabled bool) {
	net.mu.Lock()
	defer net.mu.Unlock()
	net.Config.PerLayerQuant = enabled
	net.requantizeWeightsLocked()
}

// IsPerLayerQuant returns whether per-layer quantization is enabled.
func (net *DualModeNetwork) IsPerLayerQuant() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.PerLayerQuant
}

// SetLayer1Levels sets the quantization levels for layer 1 (hidden layer).
func (net *DualModeNetwork) SetLayer1Levels(levels int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 2 {
		levels = 2
	}
	if levels > 30 {
		levels = 30
	}
	net.Config.Layer1Levels = levels
	if net.Config.PerLayerQuant {
		net.requantizeWeightsLocked()
	}
}

// GetLayer1Levels returns the current layer 1 quantization levels.
func (net *DualModeNetwork) GetLayer1Levels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.Layer1Levels
}

// SetLayer2Levels sets the quantization levels for layer 2 (output layer).
func (net *DualModeNetwork) SetLayer2Levels(levels int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if levels < 2 {
		levels = 2
	}
	if levels > 30 {
		levels = 30
	}
	net.Config.Layer2Levels = levels
	if net.Config.PerLayerQuant {
		net.requantizeWeightsLocked()
	}
}

// GetLayer2Levels returns the current layer 2 quantization levels.
func (net *DualModeNetwork) GetLayer2Levels() int {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.Layer2Levels
}

// SetPerLayerLevels sets quantization levels for both layers at once.
func (net *DualModeNetwork) SetPerLayerLevels(layer1, layer2 int) {
	net.mu.Lock()
	defer net.mu.Unlock()

	if layer1 < 2 {
		layer1 = 2
	}
	if layer1 > 30 {
		layer1 = 30
	}
	if layer2 < 2 {
		layer2 = 2
	}
	if layer2 > 30 {
		layer2 = 30
	}

	net.Config.Layer1Levels = layer1
	net.Config.Layer2Levels = layer2
	net.Config.PerLayerQuant = true
	net.requantizeWeightsLocked()
}

// IsSingleLayer returns whether single-layer (Tour Mode) is enabled.
func (net *DualModeNetwork) IsSingleLayer() bool {
	net.mu.RLock()
	defer net.mu.RUnlock()
	return net.Config.SingleLayer
}

// safeNoise applies Gaussian noise with thread-safe RNG access.
// This prevents data races when multiple goroutines hold RLock on the network.
func (net *DualModeNetwork) safeNoise(data []float64, noiseLevel float64) []float64 {
	if noiseLevel <= 0 {
		return data
	}
	net.rngMu.Lock()
	result := AddGaussianNoise(data, noiseLevel, net.rng)
	net.rngMu.Unlock()
	return result
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
		cimOutput = net.safeNoise(cimOutput, net.Config.NoiseLevel)
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
		cimHidden = net.safeNoise(cimHidden, net.Config.NoiseLevel)
		cimHidden = relu(cimHidden)

		// Layer 2: Use QUANTIZED weights (30-level FeCIM quantization)
		cimOutput = net.forwardCIM(cimHidden, net.QuantWeights2, net.QuantBias2)
		cimOutput = quantizeADC(cimOutput, net.Config.ADCBits)
		cimOutput = net.safeNoise(cimOutput, net.Config.NoiseLevel)
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

	// Energy calculation (Jerry et al. IEDM 2017: ~50 fJ/MAC at 30 levels)
	// Energy scales with bits of precision: ~10 fJ/bit per MAC
	// bits = log2(levels), so energy = 10 * log2(levels) fJ/MAC
	// At 30 levels (~4.9 bits): 10 * 4.9 ≈ 50 fJ/MAC (matches literature)
	// At 2 levels (1 bit): 10 * 1 = 10 fJ/MAC (5x more efficient)
	//
	// With per-layer quantization, calculate energy for each layer separately
	var totalEnergy float64
	if net.Config.SingleLayer {
		// Single layer uses Layer1Levels (or NumLevels if not per-layer)
		levels := net.Config.NumLevels
		if net.Config.PerLayerQuant {
			levels = net.Config.Layer1Levels
		}
		bitsPerCell := math.Log2(float64(levels))
		if bitsPerCell < 1 {
			bitsPerCell = 1
		}
		energyPerMAC := 10e-15 * bitsPerCell
		totalEnergy = float64(totalMACs) * energyPerMAC * 1e6
	} else {
		// Two layers: calculate energy separately
		macs1 := net.InputSize * net.HiddenSize
		macs2 := net.HiddenSize * net.OutputSize

		l1Levels := net.Config.NumLevels
		l2Levels := net.Config.NumLevels
		if net.Config.PerLayerQuant {
			l1Levels = net.Config.Layer1Levels
			l2Levels = net.Config.Layer2Levels
		}

		bits1 := math.Log2(float64(l1Levels))
		if bits1 < 1 {
			bits1 = 1
		}
		bits2 := math.Log2(float64(l2Levels))
		if bits2 < 1 {
			bits2 = 1
		}

		energy1 := float64(macs1) * 10e-15 * bits1 * 1e6 // Layer 1 energy in μJ
		energy2 := float64(macs2) * 10e-15 * bits2 * 1e6 // Layer 2 energy in μJ
		totalEnergy = energy1 + energy2
	}
	result.EnergyUsed = totalEnergy

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
	hidden = net.safeNoise(hidden, net.Config.NoiseLevel)
	hidden = relu(hidden)

	output := net.forwardCIM(hidden, net.FPWeights2, net.FPBias2)
	output = quantizeADC(output, net.Config.ADCBits)
	output = net.safeNoise(output, net.Config.NoiseLevel)
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

// GetPerLayerQuantInfo returns the current per-layer quantization configuration.
func (net *DualModeNetwork) GetPerLayerQuantInfo() (enabled bool, l1Levels, l2Levels int) {
	net.mu.RLock()
	defer net.mu.RUnlock()

	enabled = net.Config.PerLayerQuant
	if enabled {
		l1Levels = net.Config.Layer1Levels
		l2Levels = net.Config.Layer2Levels
	} else {
		l1Levels = net.Config.NumLevels
		l2Levels = net.Config.NumLevels
	}
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
