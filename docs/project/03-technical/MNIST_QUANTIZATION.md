## Plan: MNIST Weights Quantized to N Levels

---

## Overview

```
PyTorch (train) → Export (float32) → Go (load) → Quantize (N levels) → Crossbar
```

---

## Step 1: Train or Download Pretrained Weights

### Option A: Download Pretrained (Fastest)

```bash
# Simple MNIST MLP weights exist everywhere
wget https://github.com/pytorch/examples/raw/main/mnist/main.py
python main.py --epochs 5 --save-model
# Outputs: mnist_cnn.pt
```

### Option B: Train Minimal Network (More Control)

```python
# train_mnist.py
import torch
import torch.nn as nn
from torchvision import datasets, transforms

class SimpleMLP(nn.Module):
    def __init__(self):
        super().__init__()
        self.fc1 = nn.Linear(784, 128)  # Layer 1: 784×128
        self.fc2 = nn.Linear(128, 10)   # Layer 2: 128×10
    
    def forward(self, x):
        x = x.view(-1, 784)
        x = torch.relu(self.fc1(x))
        x = self.fc2(x)
        return x

# Train
model = SimpleMLP()
# ... training loop ...

# Save weights
torch.save({
    'fc1_weight': model.fc1.weight.data,  # [128, 784]
    'fc1_bias': model.fc1.bias.data,      # [128]
    'fc2_weight': model.fc2.weight.data,  # [10, 128]
    'fc2_bias': model.fc2.bias.data,      # [10]
}, 'mnist_mlp.pt')
```

---

## Step 2: Export to Portable Format

```python
# export_weights.py
import torch
import numpy as np
import json

# Load trained model
checkpoint = torch.load('mnist_mlp.pt')

# Convert to numpy
weights = {
    'fc1_weight': checkpoint['fc1_weight'].numpy().tolist(),
    'fc1_bias': checkpoint['fc1_bias'].numpy().tolist(),
    'fc2_weight': checkpoint['fc2_weight'].numpy().tolist(),
    'fc2_bias': checkpoint['fc2_bias'].numpy().tolist(),
}

# Save as JSON (portable, Go can read)
with open('mnist_weights.json', 'w') as f:
    json.dump(weights, f)

# Or save as binary (smaller, faster)
np.savez('mnist_weights.npz',
    fc1_weight=checkpoint['fc1_weight'].numpy(),
    fc1_bias=checkpoint['fc1_bias'].numpy(),
    fc2_weight=checkpoint['fc2_weight'].numpy(),
    fc2_bias=checkpoint['fc2_bias'].numpy(),
)
```

---

## Step 3: Go - Load Weights

```go
// pkg/weights/loader.go
package weights

import (
    "encoding/json"
    "os"
)

type NetworkWeights struct {
    FC1Weight [][]float64 `json:"fc1_weight"` // [128][784]
    FC1Bias   []float64   `json:"fc1_bias"`   // [128]
    FC2Weight [][]float64 `json:"fc2_weight"` // [10][128]
    FC2Bias   []float64   `json:"fc2_bias"`   // [10]
}

func LoadWeights(path string) (*NetworkWeights, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    var weights NetworkWeights
    decoder := json.NewDecoder(file)
    err = decoder.Decode(&weights)
    if err != nil {
        return nil, err
    }
    
    return &weights, nil
}
```

---

## Step 4: Quantize to N Levels

```go
// pkg/weights/quantize.go
package weights

import "math"

// QuantizedWeights holds weights as integer levels
type QuantizedWeights struct {
    FC1Weight [][]int // [128][784] values 0 to N-1
    FC1Bias   []int   // [128]
    FC2Weight [][]int // [10][128]
    FC2Bias   []int   // [10]
    
    NumLevels int
    MinVal    float64 // original min (for reconstruction)
    MaxVal    float64 // original max (for reconstruction)
}

// Quantize converts float weights to N discrete levels
func Quantize(weights *NetworkWeights, numLevels int) *QuantizedWeights {
    // Step 1: Find global min/max across all weights
    minVal := math.MaxFloat64
    maxVal := -math.MaxFloat64
    
    for _, row := range weights.FC1Weight {
        for _, v := range row {
            minVal = math.Min(minVal, v)
            maxVal = math.Max(maxVal, v)
        }
    }
    for _, row := range weights.FC2Weight {
        for _, v := range row {
            minVal = math.Min(minVal, v)
            maxVal = math.Max(maxVal, v)
        }
    }
    // Include biases
    for _, v := range weights.FC1Bias {
        minVal = math.Min(minVal, v)
        maxVal = math.Max(maxVal, v)
    }
    for _, v := range weights.FC2Bias {
        minVal = math.Min(minVal, v)
        maxVal = math.Max(maxVal, v)
    }
    
    // Step 2: Quantize function
    quantize := func(val float64) int {
        // Normalize to [0, 1]
        normalized := (val - minVal) / (maxVal - minVal)
        // Scale to [0, numLevels-1]
        level := int(normalized * float64(numLevels-1) + 0.5)
        // Clamp
        if level < 0 {
            level = 0
        }
        if level >= numLevels {
            level = numLevels - 1
        }
        return level
    }
    
    // Step 3: Apply to all weights
    q := &QuantizedWeights{
        NumLevels: numLevels,
        MinVal:    minVal,
        MaxVal:    maxVal,
    }
    
    // FC1 Weight
    q.FC1Weight = make([][]int, len(weights.FC1Weight))
    for i, row := range weights.FC1Weight {
        q.FC1Weight[i] = make([]int, len(row))
        for j, v := range row {
            q.FC1Weight[i][j] = quantize(v)
        }
    }
    
    // FC1 Bias
    q.FC1Bias = make([]int, len(weights.FC1Bias))
    for i, v := range weights.FC1Bias {
        q.FC1Bias[i] = quantize(v)
    }
    
    // FC2 Weight
    q.FC2Weight = make([][]int, len(weights.FC2Weight))
    for i, row := range weights.FC2Weight {
        q.FC2Weight[i] = make([]int, len(row))
        for j, v := range row {
            q.FC2Weight[i][j] = quantize(v)
        }
    }
    
    // FC2 Bias
    q.FC2Bias = make([]int, len(weights.FC2Bias))
    for i, v := range weights.FC2Bias {
        q.FC2Bias[i] = quantize(v)
    }
    
    return q
}

// Dequantize converts level back to float (for computation)
func (q *QuantizedWeights) Dequantize(level int) float64 {
    normalized := float64(level) / float64(q.NumLevels-1)
    return q.MinVal + normalized*(q.MaxVal-q.MinVal)
}

// ToFloat converts quantized weight matrix back to float
func (q *QuantizedWeights) FC1WeightFloat() [][]float64 {
    result := make([][]float64, len(q.FC1Weight))
    for i, row := range q.FC1Weight {
        result[i] = make([]float64, len(row))
        for j, level := range row {
            result[i][j] = q.Dequantize(level)
        }
    }
    return result
}
```

---

## Step 5: Map Levels to Conductance

```go
// pkg/crossbar/conductance.go
package crossbar

// LevelToConductance maps discrete level to physical conductance
func LevelToConductance(level int, numLevels int, Gmin, Gmax float64) float64 {
    // Linear mapping
    ratio := float64(level) / float64(numLevels-1)
    return Gmin + ratio*(Gmax-Gmin)
}

// For Ferroelectric CIM HZO:
// Gmin ≈ 1 μS (low conductance state)
// Gmax ≈ 100 μS (high conductance state)

type CrossbarConfig struct {
    NumLevels int
    Gmin      float64 // minimum conductance (Siemens)
    Gmax      float64 // maximum conductance (Siemens)
}

func DefaultFerroelectric CIMConfig() CrossbarConfig {
    return CrossbarConfig{
        NumLevels: 30,
        Gmin:      1e-6,   // 1 μS
        Gmax:      100e-6, // 100 μS
    }
}
```

---

## Step 6: Full Pipeline Usage

```go
// main.go
package main

import (
    "fmt"
    "multilayer-ferroelectric-cim-visualizer/pkg/weights"
    "multilayer-ferroelectric-cim-visualizer/pkg/crossbar"
)

func main() {
    // Load float weights
    w, err := weights.LoadWeights("mnist_weights.json")
    if err != nil {
        panic(err)
    }
    
    // Quantize to 30 levels (Ferroelectric CIM)
    q30 := weights.Quantize(w, 30)
    fmt.Printf("Quantized to %d levels\n", q30.NumLevels)
    
    // Or quantize to different levels for comparison
    q8 := weights.Quantize(w, 8)    // 3-bit
    q16 := weights.Quantize(w, 16)  // 4-bit
    q64 := weights.Quantize(w, 64)  // 6-bit
    
    // Convert to conductance for crossbar
    config := crossbar.DefaultFerroelectric CIMConfig()
    config.NumLevels = 30
    
    // Example: get conductance for a weight
    level := q30.FC1Weight[0][0]
    G := crossbar.LevelToConductance(level, config.NumLevels, config.Gmin, config.Gmax)
    fmt.Printf("Level %d → Conductance %.2f μS\n", level, G*1e6)
}
```

---

## Step 7: Test Accuracy vs Number of Levels

```go
// pkg/weights/accuracy.go
package weights

// TestQuantizationAccuracy compares accuracy at different bit depths
func TestQuantizationAccuracy(original *NetworkWeights, testData []TestSample) {
    levels := []int{2, 4, 8, 16, 30, 64, 128, 256}
    
    for _, n := range levels {
        q := Quantize(original, n)
        accuracy := RunInference(q, testData)
        fmt.Printf("Levels: %3d | Accuracy: %.2f%%\n", n, accuracy*100)
    }
}

// Expected output:
// Levels:   2 | Accuracy: 45.00%  (binary, terrible)
// Levels:   4 | Accuracy: 78.00%
// Levels:   8 | Accuracy: 84.00%
// Levels:  16 | Accuracy: 86.00%
// Levels:  30 | Accuracy: 87.00%  ← Ferroelectric CIM target
// Levels:  64 | Accuracy: 87.50%
// Levels: 128 | Accuracy: 87.80%
// Levels: 256 | Accuracy: 88.00%  ← theoretical max
```

---

## File Structure

```
multilayer-ferroelectric-cim-visualizer/
├── scripts/
│   ├── train_mnist.py       # Train simple MLP
│   └── export_weights.py    # Export to JSON
├── data/
│   └── mnist_weights.json   # Portable weights
├── pkg/
│   ├── weights/
│   │   ├── loader.go        # Load JSON
│   │   ├── quantize.go      # Quantize to N levels
│   │   └── accuracy.go      # Test accuracy
│   └── crossbar/
│       └── conductance.go   # Level → conductance
```

---

## Summary

| Step | Tool | Output |
|------|------|--------|
| Train | Python/PyTorch | mnist_mlp.pt |
| Export | Python | mnist_weights.json |
| Load | Go | NetworkWeights struct |
| Quantize | Go | QuantizedWeights (N levels) |
| Map | Go | Conductance values |
| Test | Go | Accuracy vs levels |

---

## Key Function

```go
q := weights.Quantize(floatWeights, 30)  // Ferroelectric CIM
q := weights.Quantize(floatWeights, 8)   // Compare to 3-bit
q := weights.Quantize(floatWeights, 256) // Compare to 8-bit
```

**One function. Any number of levels.**

---

Ready to integrate into the ralph command?
