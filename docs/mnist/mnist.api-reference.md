# MNIST Module API Reference

**Version:** 1.0
**Date:** 2026-01-27
**Package:** `fecim-lattice-tools/module3-mnist/pkg/core`

---

## Table of Contents

1. [Types](#types)
2. [Network Creation](#network-creation)
3. [Inference](#inference)
4. [Configuration](#configuration)
5. [Weight Management](#weight-management)
6. [Quantization](#quantization)
7. [Utilities](#utilities)

---

## Types

### NetworkConfig

Configuration for the CIM inference path.

```go
type NetworkConfig struct {
    NumLevels     int     // Quantization levels (2-30)
    NoiseLevel    float64 // Noise coefficient σ/μ (0.0-0.20)
    ADCBits       int     // ADC resolution (3-16)
    DACBits       int     // DAC resolution (3-16)
    EnableSneak   bool    // Enable sneak path simulation
    IRDrop        bool    // Enable IR drop simulation
    SingleLayer   bool    // Use single-layer (784→10) architecture
    PerLayerQuant bool    // Enable per-layer quantization
    Layer1Levels  int     // Quantization levels for layer 1
    Layer2Levels  int     // Quantization levels for layer 2
}
```

### DualModeNetwork

Neural network with dual inference paths (FP and CIM).

```go
type DualModeNetwork struct {
    InputSize  int  // Input dimensions (784 for MNIST)
    HiddenSize int  // Hidden layer size
    OutputSize int  // Output classes (10 for MNIST)
    Config     *NetworkConfig
    // ... internal fields
}
```

### InferenceResult

Results from dual-path inference.

```go
type InferenceResult struct {
    // FP Path Results
    FPLogits        []float64  // Raw logits
    FPProbabilities []float64  // Softmax probabilities
    FPPrediction    int        // Predicted class (0-9)
    FPConfidence    float64    // Confidence (0-1)
    FPHidden        []float64  // Hidden activations

    // CIM Path Results
    CIMLogits        []float64
    CIMProbabilities []float64
    CIMPrediction    int
    CIMConfidence    float64
    CIMHidden        []float64

    // Comparison Metrics
    Agree        bool    // Do FP and CIM agree?
    Disagreement float64 // KL divergence
    EnergyUsed   float64 // Energy consumption (μJ)
}
```

### QuantizationStats

Statistics about weight quantization.

```go
type QuantizationStats struct {
    MinError    float64 // Minimum quantization error
    MaxError    float64 // Maximum quantization error
    MeanError   float64 // Mean quantization error
    StdError    float64 // Standard deviation of error
    UniqueLevels int    // Number of unique levels used
}
```

---

## Network Creation

### NewDualModeNetwork

Creates a new dual-mode network with Xavier initialization.

```go
func NewDualModeNetwork(inputSize, hiddenSize, outputSize int) *DualModeNetwork
```

**Parameters:**
- `inputSize`: Input dimensions (784 for MNIST)
- `hiddenSize`: Hidden layer neurons (typically 64-256)
- `outputSize`: Output classes (10 for MNIST)

**Returns:** Initialized network with random weights

**Example:**
```go
net := core.NewDualModeNetwork(784, 128, 10)
```

### DefaultNetworkConfig

Returns default configuration for optimal FeCIM operation.

```go
func DefaultNetworkConfig() *NetworkConfig
```

**Returns:** Config with NumLevels=30, NoiseLevel=0.01, ADCBits=8, DACBits=8

---

## Inference

### Infer

Runs dual-path inference (FP and CIM) and returns comparison results.

```go
func (net *DualModeNetwork) Infer(input []float64) *InferenceResult
```

**Parameters:**
- `input`: Normalized pixel values [0,1], length must match InputSize

**Returns:** `*InferenceResult` with FP and CIM predictions, or `nil` if input length is wrong

**Example:**
```go
pixels := make([]float64, 784)
// ... populate pixels
result := net.Infer(pixels)
fmt.Printf("FP: %d (%.1f%%), CIM: %d (%.1f%%)\n",
    result.FPPrediction, result.FPConfidence*100,
    result.CIMPrediction, result.CIMConfidence*100)
```

### InferFPOnly

Runs only the floating-point path (faster, for evaluation).

```go
func (net *DualModeNetwork) InferFPOnly(input []float64) (prediction int, confidence float64, probs []float64)
```

**Returns:** Predicted class, confidence, and full probability distribution

### InferCIMOnly

Runs only the CIM path with quantization effects.

```go
func (net *DualModeNetwork) InferCIMOnly(input []float64) (prediction int, confidence float64, probs []float64)
```

**Returns:** Predicted class, confidence, and full probability distribution

---

## Configuration

### SetNumLevels / GetNumLevels

Set/get global quantization levels.

```go
func (net *DualModeNetwork) SetNumLevels(levels int)
func (net *DualModeNetwork) GetNumLevels() int
```

**Parameters:**
- `levels`: Number of quantization levels (2-30, clamped)

**Note:** Automatically calls `RequantizeWeights()` after change.

### SetNoiseLevel

Set noise injection level for CIM simulation.

```go
func (net *DualModeNetwork) SetNoiseLevel(noise float64)
```

**Parameters:**
- `noise`: Noise coefficient σ/μ (0.0-0.20, clamped)

### SetADCBits / SetDACBits

Set ADC/DAC resolution.

```go
func (net *DualModeNetwork) SetADCBits(bits int)
func (net *DualModeNetwork) SetDACBits(bits int)
```

**Parameters:**
- `bits`: Resolution in bits (3-16, clamped)

### SetSingleLayer / IsSingleLayer

Enable/check single-layer mode (784→10).

```go
func (net *DualModeNetwork) SetSingleLayer(enabled bool)
func (net *DualModeNetwork) IsSingleLayer() bool
```

### Per-Layer Quantization

```go
func (net *DualModeNetwork) SetPerLayerQuant(enabled bool)
func (net *DualModeNetwork) IsPerLayerQuant() bool
func (net *DualModeNetwork) SetLayer1Levels(levels int)
func (net *DualModeNetwork) GetLayer1Levels() int
func (net *DualModeNetwork) SetLayer2Levels(levels int)
func (net *DualModeNetwork) GetLayer2Levels() int
func (net *DualModeNetwork) SetPerLayerLevels(layer1, layer2 int)
func (net *DualModeNetwork) GetPerLayerQuantInfo() (enabled bool, l1Levels, l2Levels int)
```

---

## Weight Management

### LoadWeights

Load weights from JSON file.

```go
func (net *DualModeNetwork) LoadWeights(filename string) error
```

**Parameters:**
- `filename`: Path to weights JSON file

**Returns:** Error if file cannot be read or parsed

**Example:**
```go
err := net.LoadWeights("weights/mnist_weights_30lvl.json")
if err != nil {
    log.Fatal(err)
}
```

### LoadWeightsForLevel

Load QAT weights for specific quantization level.

```go
func (net *DualModeNetwork) LoadWeightsForLevel(dataDir string, levels int) error
```

**Parameters:**
- `dataDir`: Directory containing weight files
- `levels`: Target quantization level (finds best match from AvailableQATLevels)

### GetWeightsFilename

Get standard filename for weights at given level.

```go
func GetWeightsFilename(dataDir string, levels int) string
```

**Returns:** Path like `dataDir/weights/mnist_weights_30lvl.json`

### GetBestMatchingWeightsLevel

Find closest available QAT level.

```go
func GetBestMatchingWeightsLevel(targetLevels int) int
```

**Parameters:**
- `targetLevels`: Desired quantization level

**Returns:** Closest level from `AvailableQATLevels`

### GetFPWeights / GetQuantWeights

Retrieve weight matrices.

```go
func (net *DualModeNetwork) GetFPWeights() (w1, w2 [][]float64, b1, b2 []float64)
func (net *DualModeNetwork) GetQuantWeights() (w1, w2 [][]float64, b1, b2 []float64)
```

**Returns:** Weight matrices and bias vectors for both layers

---

## Quantization

### RequantizeWeights

Apply current quantization settings to weights.

```go
func (net *DualModeNetwork) RequantizeWeights()
```

**Note:** Called automatically by `SetNumLevels()` and other config changes.

### QuantizeWeights

Quantize a weight matrix to N levels.

```go
func QuantizeWeights(fpWeights [][]float64, levels int) ([][]float64, error)
```

**Parameters:**
- `fpWeights`: Full-precision weight matrix
- `levels`: Number of quantization levels (≥2)

**Returns:** Quantized weight matrix, error if levels < 2

### QuantizeBias

Quantize a bias vector to N levels.

```go
func QuantizeBias(fpBias []float64, levels int) ([]float64, error)
```

### ComputeQuantizationStats

Calculate statistics about quantization error.

```go
func ComputeQuantizationStats(original, quantized [][]float64) QuantizationStats
```

**Returns:** Statistics including min/max/mean error and unique levels used

### GetQuantizationStats

Get stats for network's current quantization.

```go
func (net *DualModeNetwork) GetQuantizationStats() (layer1Stats, layer2Stats QuantizationStats)
```

---

## Utilities

### RandomSource

Thread-safe random number generator.

```go
type RandomSource struct {
    // ... internal state
}

func NewRandomSource(seed uint64) *RandomSource
func (r *RandomSource) NormFloat64() float64  // Standard normal
func (r *RandomSource) Float64() float64      // Uniform [0,1)
func (r *RandomSource) Intn(n int) int        // Uniform [0,n)
```

### AddGaussianNoise

Add Gaussian noise to values.

```go
func AddGaussianNoise(values []float64, noiseLevel float64, rng *RandomSource) []float64
```

**Parameters:**
- `values`: Input values
- `noiseLevel`: Noise coefficient (σ = noiseLevel × mean)
- `rng`: Random source

---

## Constants

### AvailableQATLevels

Pre-trained QAT weight levels.

```go
var AvailableQATLevels = []int{10, 20, 29, 30, 31}
```

### FeCIMLevels

Maximum FeCIM quantization levels.

```go
const FeCIMLevels = 30
```

---

## Thread Safety

All public methods on `DualModeNetwork` are thread-safe:

- **Read operations** (`Infer`, `GetWeights`, `GetNumLevels`): Use `RLock`
- **Write operations** (`SetNumLevels`, `RequantizeWeights`, `LoadWeights`): Use exclusive `Lock`

Multiple goroutines can safely call `Infer()` concurrently.

---

## Error Handling

| Function | Error Condition |
|----------|----------------|
| `LoadWeights` | File not found, invalid JSON |
| `QuantizeWeights` | levels < 2 |
| `Infer` | Returns `nil` if input length wrong |

---

*Last updated: 2026-01-27*
