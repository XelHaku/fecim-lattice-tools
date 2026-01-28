package core

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
	SingleLayer bool    // Calibration Mode: use single-layer (784→10) like hardware demo

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

	// Single-Layer Weights (Calibration Mode: 784→10 direct)
	// This matches the hardware demo architecture
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

	// GPU acceleration flag
	useGPU bool
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

	// Initialize single-layer weights (Calibration Mode: 784→10)
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
	// Single-layer weights (Calibration Mode: 784→10 direct)
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

// ScanAvailableQATLevels scans the data directory for available weight files.
// Returns sorted list of levels that have trained weights.
// Pattern: pretrained_weights.json (30) and pretrained_weights_{N}.json (N levels)
func ScanAvailableQATLevels(dataDir string) []int {
	levels := make(map[int]bool)

	// Check for default 30-level weights
	defaultPath := filepath.Join(dataDir, "pretrained_weights.json")
	if _, err := os.Stat(defaultPath); err == nil {
		levels[30] = true
	}

	// Scan for level-specific weight files: pretrained_weights_{N}.json
	pattern := filepath.Join(dataDir, "pretrained_weights_*.json")
	matches, err := filepath.Glob(pattern)
	if err == nil {
		for _, match := range matches {
			base := filepath.Base(match)
			// Skip PTQ weights file
			if base == "pretrained_weights_ptq.json" {
				continue
			}
			// Extract level number from pretrained_weights_{N}.json
			var level int
			if _, err := fmt.Sscanf(base, "pretrained_weights_%d.json", &level); err == nil {
				levels[level] = true
			}
		}
	}

	// Convert map to sorted slice
	result := make([]int, 0, len(levels))
	for l := range levels {
		result = append(result, l)
	}
	sort.Ints(result)

	// If no weights found, return just 30 as fallback
	if len(result) == 0 {
		return []int{30}
	}

	return result
}

// GetWeightsFilename returns the appropriate weights filename for a quantization level.
// Returns the exact match if available, otherwise returns the default 30-level weights.
func GetWeightsFilename(dataDir string, levels int) string {
	// Check for exact match first (level-specific file)
	if levels != 30 {
		levelPath := filepath.Join(dataDir, fmt.Sprintf("pretrained_weights_%d.json", levels))
		if _, err := os.Stat(levelPath); err == nil {
			return levelPath
		}
	}
	// Default to 30-level weights (backward compatible)
	return filepath.Join(dataDir, "pretrained_weights.json")
}

// GetBestMatchingWeightsLevel returns the closest available QAT level for a given target.
func GetBestMatchingWeightsLevel(dataDir string, targetLevels int) int {
	available := ScanAvailableQATLevels(dataDir)
	if len(available) == 0 {
		return 30
	}

	bestMatch := available[0]
	bestDiff := 1000
	for _, l := range available {
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

	// Load single-layer weights if provided (Calibration Mode: 784→10)
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
		// Accuracy depends on quantization noise; peer-reviewed FeCIM: 96-98%, software baseline: 98-99%
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

// LoadWeightsForLevel loads weights optimized for a specific quantization level.
// It looks for level-specific weight files (e.g., pretrained_weights_20.json)
// and falls back to the default 30-level weights if not found.
func (net *DualModeNetwork) LoadWeightsForLevel(dataDir string, levels int) error {
	// Find the best matching weights file
	bestLevel := GetBestMatchingWeightsLevel(dataDir, levels)
	filename := GetWeightsFilename(dataDir, bestLevel)

	// Check if the file exists
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		// Fall back to default weights
		filename = filepath.Join(dataDir, "pretrained_weights.json")
	}

	return net.LoadWeights(filename)
}
