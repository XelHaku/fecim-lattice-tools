# Module 2: Crossbar Array Simulation

## Overview

Module 2 simulates ferroelectric crossbar arrays for Compute-in-Memory (CIM) operations. It models analog matrix-vector multiplication (MVM) and includes non-idealities such as IR drop, sneak paths, drift, temperature scaling, and variation.

**Simulation defaults:**
- Discrete conductance levels are **configurable** with a default of 30.
- Architectures supported: passive (0T1R) and transistor-isolated (1T1R).

All values in this module are **model parameters** for education and exploration, not device measurements.

---

## Key Features

- **MVM simulation** with DAC/ADC quantization
- **Differential arrays** for signed weights
- **Non-idealities**: IR drop, sneak paths, drift, variation, half-select disturb
- **Temperature scaling** for model parameters
- **Visualization**: heatmaps, parametric sweeps, and per-tab analysis
- **Weight I/O**: JSON/binary serialization for reuse
- **Optional training/inference helpers** for MNIST-style demos

---

## Quick Start

### Build and Run

```bash
# Build the main unified application
go build -o fecim-lattice-tools ./cmd/fecim-lattice-tools
./fecim-lattice-tools

# Or use the launch script
./launch.sh
```

### Command-Line Tools

```bash
# Crossbar GUI (standard mode)
go run ./module2-crossbar/cmd/crossbar-gui

# Crossbar GUI (enhanced mode)
go run ./module2-crossbar/cmd/crossbar-gui -enhanced

# Neural network inference (terminal)
go run ./module2-crossbar/cmd/crossbar-gui inference -size=64 -layers=3
```

### Simple Example

```go
package main

import (
    "fmt"
    "fecim-lattice-tools/module2-crossbar/pkg/crossbar"
)

func main() {
    cfg := &crossbar.Config{
        Rows:       64,
        Cols:       784,
        NoiseLevel: 0.02,
        ADCBits:    8,
        DACBits:    8,
    }
    array, _ := crossbar.NewArray(cfg)

    // Program random weights (quantized to the configured discrete levels)
    weights := make([][]float64, 64)
    for i := range weights {
        weights[i] = make([]float64, 784)
        for j := range weights[i] {
            weights[i][j] = rand.Float64() // [0, 1]
        }
    }
    array.ProgramWeightMatrix(weights)

    input := make([]float64, 784)
    for i := range input {
        input[i] = rand.Float64()
    }
    output, _ := array.MVM(input)
    fmt.Printf("Output shape: [%d]\n", len(output))

    opts := crossbar.DefaultMVMOptions()
    opts.Architecture = "0T1R"
    result, _ := array.MVMWithNonIdealities(input, opts)

    fmt.Printf("Ideal output:  %v\n", result.IdealOutput[:4])
    fmt.Printf("Actual output: %v\n", result.ActualOutput[:4])
    fmt.Printf("RMSE:          %.4f\n", result.RMSE)
}
```

---

## Core Concepts

### Discrete Conductance Levels

The simulator stores weights as discrete conductance levels. The default is 30 levels, but this is configurable.

```go
// Quantize any value to the nearest level
quantized := crossbar.QuantizeToLevels(0.542)  // Returns ~14/29

// Get discrete level index (0-29 for default 30 levels)
level := crossbar.GetLevel(0.5)
```

### Matrix-Vector Multiplication (MVM)

```
I[i] = Sum(G[i][j] * V[j])
```

- Inputs are voltages
- Stored weights are conductances
- Outputs are currents (optionally quantized by ADC)

---

## API Quick Reference

### Creating Arrays

```go
cfg := &crossbar.Config{
    Rows:       128,
    Cols:       784,
    NoiseLevel: 0.02,
    ADCBits:    8,
    DACBits:    8,
}

cfg.ConductanceModel = crossbar.ConductanceExponential
cfg.Endurance = crossbar.DefaultEnduranceConfig()
cfg.ProcessVariation = crossbar.DefaultProcessVariationConfig()
cfg.HalfSelect = crossbar.DefaultHalfSelectConfig()

array, err := crossbar.NewArray(cfg)
if err != nil {
    log.Fatal(err)
}
```

### Programming Weights

```go
err := array.ProgramWeight(row, col, 0.75)  // Quantized to configured levels

weights := make([][]float64, 128)
for i := range weights {
    weights[i] = make([]float64, 784)
}
err := array.ProgramWeightMatrix(weights)
```

### Ideal MVM

```go
output, err := array.MVM(input)
```

### MVM with Non-Idealities

```go
opts := &crossbar.MVMOptions{
    EnableIRDrop:     true,
    EnableSneakPaths: true,
    EnableVariation:  true,
    EnableDrift:      false,
    Temperature:      300.0,
    Architecture:     "0T1R", // "0T1R" or "1T1R"
}

result, err := array.MVMWithNonIdealities(input, opts)
```

### Differential Arrays (Signed Weights)

```go
pos, _ := crossbar.NewArray(cfg)
neg, _ := crossbar.NewArray(cfg)

diffArr := crossbar.NewDifferentialArray(pos, neg)
diffArr.ProgramWeights(signedWeights)

output, _ := diffArr.MVM(input)
```

---

## Testing

```bash
cd <local-path>
go test ./module2-crossbar/...
```

---

## Related Documentation

- `docs/crossbar/` for educational notes
- `docs/development/TESTING.md` for test conventions
- `CLAUDE.md` for project-wide guidance

Note: Some research notes under `docs/` include external claims and literature summaries. Treat them as references only; verified claims are tracked in `docs/comparison/HONESTY_AUDIT.md`.
