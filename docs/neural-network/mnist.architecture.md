# MNIST Module Architecture

**Version:** 1.0
**Date:** 2026-02-02

This document describes the architecture of the MNIST module (`module3-mnist/`), which implements dual-path neural network inference comparing ideal floating-point (FP) computation with realistic Ferroelectric Compute-in-Memory (FeCIM) hardware simulation.

---

## Table of Contents

1. [Overview](#overview)
2. [Package Structure](#package-structure)
3. [Core Components](#core-components)
4. [Data Flow](#data-flow)
5. [Key Abstractions](#key-abstractions)
6. [Threading Model](#threading-model)
7. [Configuration](#configuration)

---

## Overview

The MNIST module demonstrates FeCIM's capabilities for neural network inference by:

1. **Dual-Path Inference**: Running the same input through both ideal FP and realistic CIM paths
2. **Hardware Simulation**: Modeling quantization, noise, ADC/DAC effects, and energy consumption
3. **Interactive Visualization**: Real-time comparison of predictions, probabilities, and energy metrics

### Architecture Goals

- **Educational**: Demonstrate FeCIM benefits clearly
- **Accurate**: Model realistic hardware behavior with reported in literature parameters
- **Interactive**: Allow users to explore parameter effects in real-time

---

## Package Structure

```
module3-mnist/
├── cmd/
│   ├── mnist/              # Standalone CLI tool
│   ├── mnist-gui/          # Standalone GUI app
│   ├── train-network/      # Two-layer training utility
│   ├── train-ptq/          # Post-training quantization
│   └── train-single-layer/ # Single-layer training
├── pkg/
│   ├── core/               # Network implementation, inference, quantization
│   ├── gui/                # Fyne-based GUI components
│   ├── mnist/              # MNIST data loading
│   └── training/           # Network training utilities
└── data/                   # MNIST data files (downloaded separately)
```

---

## Core Components

### 1. DualModeNetwork (`pkg/core/network.go`)

The central abstraction representing a neural network with dual inference paths.

```
┌─────────────────────────────────────────────────────────────┐
│                     DualModeNetwork                         │
├─────────────────────────────────────────────────────────────┤
│ Architecture:                                               │
│   InputSize:  784  (28×28 pixels)                          │
│   HiddenSize: 128  (configurable)                          │
│   OutputSize: 10   (digit classes 0-9)                     │
├─────────────────────────────────────────────────────────────┤
│ FP Weights (Float64):           │ Quant Weights (30-level): │
│   FPWeights1 [128][784]         │   QuantWeights1 [128][784]│
│   FPWeights2 [10][128]          │   QuantWeights2 [10][128] │
│   FPBias1    [128]              │   QuantBias1    [128]     │
│   FPBias2    [10]               │   QuantBias2    [10]      │
├─────────────────────────────────────────────────────────────┤
│ Single-Layer Mode (784→10):                                 │
│   SingleLayerWeights [10][784]                             │
│   QuantSingleLayerWeights [10][784]                        │
└─────────────────────────────────────────────────────────────┘
```

**Key Methods:**
- `Infer(input) → InferenceResult`: Runs both FP and CIM paths
- `InferFPOnly(input)`: Fast FP-only inference
- `InferCIMOnly(input)`: Fast CIM-only inference
- `RequantizeWeights()`: Apply current quantization settings
- `LoadWeights(filename)`: Load pre-trained weights

### 2. NetworkConfig (`pkg/core/network.go`)

Configuration for CIM inference behavior.

| Field | Type | Description |
|-------|------|-------------|
| `NumLevels` | int | Quantization levels (2-30) |
| `NoiseLevel` | float64 | Noise coefficient σ/μ (0.0-0.20) |
| `ADCBits` | int | ADC resolution (3-16 bits) |
| `DACBits` | int | DAC resolution (3-16 bits) |
| `EnableSneak` | bool | Sneak-path simulation flag (reserved for future non-idealities) |
| `IRDrop` | bool | IR-drop simulation flag (reserved for future non-idealities) |
| `SingleLayer` | bool | Use 784→10 architecture |
| `PerLayerQuant` | bool | Enable per-layer quantization |
| `Layer1Levels` | int | Layer 1 quantization levels |
| `Layer2Levels` | int | Layer 2 quantization levels |

### 3. InferenceResult (`pkg/core/network.go`)

Result container for dual-path inference.

```go
type InferenceResult struct {
    // FP Path
    FPLogits        []float64 // Pre-softmax logits
    FPPrediction    int       // Predicted digit (0-9)
    FPConfidence    float64   // Confidence (0-1)
    FPProbabilities []float64 // All class probabilities

    // CIM Path
    CIMLogits        []float64
    CIMPrediction    int
    CIMConfidence    float64
    CIMProbabilities []float64

    // Comparison
    Agree        bool    // Do predictions match?
    Disagreement float64 // KL divergence
    EnergyUsed   float64 // Energy in μJ

    // Intermediate activations (for visualization)
    FPHidden  []float64 // Hidden layer activations (nil in single-layer mode)
    CIMHidden []float64
}
```

### 4. DualModeApp (`pkg/gui/dualmode.go`)

Main GUI application controller coordinating all UI components.

**Responsibilities:**
- Widget lifecycle management
- Event handling (drawing, button clicks, sliders)
- Inference orchestration
- State synchronization

**Key UI Zones:**
1. **Input Zone**: Drawing canvas, random sample buttons
2. **Results Zone**: Predictions, confidence bars, comparison
3. **Controls Zone**: Hardware parameter sliders
4. **Weights Zone**: Weight visualization, quantization stats

---

## Data Flow

### Inference Flow

```
User Input (Drawing/Sample)
        │
        ▼
┌───────────────────┐
│   onDigitChanged  │
│   (event handler) │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│   runInference    │
│   (orchestrator)  │
└─────────┬─────────┘
          │
          ▼
┌───────────────────────────────────────────┐
│           DualModeNetwork.Infer           │
├───────────────────┬───────────────────────┤
│     FP Path       │       CIM Path        │
├───────────────────┼───────────────────────┤
│ 1. forwardFP L1   │ 1. quantizeDAC(input) │
│ 2. ReLU           │ 2. forwardCIM L1      │
│ 3. forwardFP L2   │ 3. quantizeADC        │
│ 4. softmax        │ 4. addNoise           │
│                   │ 5. ReLU               │
│                   │ 6. forwardCIM L2      │
│                   │ 7. quantizeADC        │
│                   │ 8. addNoise           │
│                   │ 9. softmax            │
└───────────────────┴───────────────────────┘
          │
          ▼
┌───────────────────┐
│  InferenceResult  │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ updateResultDisplays │
│   (UI update)     │
└───────────────────┘
```

### Weight Loading Flow

```
Weight File (JSON)
        │
        ▼
┌───────────────────┐
│   LoadWeights()   │
│   (parse JSON)    │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ FPWeights1/2      │ ← Store as-is for FP path
│ FPBias1/2         │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ RequantizeWeights │
│ (apply 30-level)  │
└─────────┬─────────┘
          │
          ▼
┌───────────────────┐
│ QuantWeights1/2   │ ← Used for CIM path
│ QuantBias1/2      │
└───────────────────┘
```

---

## Key Abstractions

### 1. Quantization (`pkg/core/quantize.go`)

Converts continuous weights to discrete FeCIM levels using symmetric range mapping.

```go
// QuantizeWeights maps [-Wmax, +Wmax] → N discrete levels (symmetric)
normalized := (w + wMax) / (2.0 * wMax)
bin := int(math.Round(normalized * float64(levels-1)))
bin = clamp(bin, 0, levels-1)
levelStep := 2.0 * wMax / float64(levels-1)
quantized := -wMax + float64(bin)*levelStep
```

### 2. ADC/DAC Simulation (`pkg/core/network_inference.go`)

```go
// quantizeDAC: Input voltage quantization (N-bit)
// quantizeADC: Output current quantization (N-bit)
```

### 3. Noise Injection (`pkg/core/quantize.go`)

```go
// AddGaussianNoise: σ/μ multiplicative noise
// noiseLevel scales with |value| to model device/read variability
result[i] = v + rng.NormFloat64()*math.Abs(v)*noiseLevel
```

---

## Threading Model

### GUI Thread Safety

All UI updates use Fyne's thread-safe mechanism:

```go
fyne.Do(func() {
    // UI updates here
})
```

### Network Thread Safety

`DualModeNetwork` uses `sync.RWMutex`:

- **Read operations** (`Infer`, `GetWeights`): Use `RLock`
- **Write operations** (`SetNumLevels`, `RequantizeWeights`): Use `Lock`

### RNG Thread Safety

Separate mutex (`rngMu`) for noise generation prevents races during parallel inference:

```go
func (net *DualModeNetwork) safeNoise(values []float64, level float64) []float64 {
    net.rngMu.Lock()
    defer net.rngMu.Unlock()
    // Generate noise
}
```

---

## Configuration

### Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `FECIM_DEBUG_SEED` | Fixed RNG seed for reproducible training + inference noise | `42` |

### Weight Files

Located in `module3-mnist/data/`:

- `pretrained_weights.json` - Default 30-level weights
- `pretrained_weights_{N}.json` - Optional QAT/PTQ weights for specific levels
- `pretrained_weights_ptq.json` - PTQ fallback (if present)
- `single_layer_weights.json` - Optional 784→10 weights

### Data Files

Located in `module3-mnist/data/`:

- `train-images-idx3-ubyte.gz` - Training images (60,000)
- `train-labels-idx1-ubyte.gz` - Training labels
- `t10k-images-idx3-ubyte.gz` - Test images (10,000)
- `t10k-labels-idx1-ubyte.gz` - Test labels

---

## References

- [FeCIM Honesty Audit](../../docs/4-research/honesty-audit.md) - Verified claims and sources
- [MNIST Fixes TODO](mnist.fixes.todo.md) - Issue tracking
- [Development Reference](../development/SCRIPT_REFERENCE.md) - Code patterns

---

*Last updated: 2026-01-27*
