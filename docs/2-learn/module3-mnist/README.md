# Module 3: MNIST Neural Network Inference

**Navigation:** [← Back to Learn](../README.md) | [ELI5](./eli5.md) | [Physics](./physics.md) | [Features](./features.md)

---

## Overview

Module 3 demonstrates neural network inference on the MNIST handwritten digit dataset using both floating-point (FP) and compute-in-memory (CIM) paths. It provides side-by-side comparison to visualize the impact of quantization and hardware non-idealities on accuracy.

**Key Concept:** Compare ideal software inference with hardware-realistic CIM inference that includes weight/activation quantization and multiplicative noise.

---

## Quick Links

### For Beginners
- **[ELI5 Explanation](./eli5.md)** - What is MNIST and why it matters

### For Developers
- **[Physics Reference](./physics.md)** - Network equations, quantization, noise models
- **[Features](./features.md)** - Dual-mode inference, visualization, workflows

### For Researchers
- **[Open-Source Tools](./tools.md)** - PyTorch, TensorFlow integration

---

## Module Contents

```
module3-mnist/
├── pkg/core/               # Inference engine
│   ├── network_inference.go  # FP vs CIM dual-mode network
│   ├── quantize.go           # Weight/activation quantization
│   └── energy.go             # Energy modeling
├── pkg/mnist/              # Dataset loader
├── pkg/gui/                # Fyne visualization
│   ├── dualmode.go           # Side-by-side comparison UI
│   └── drawing.go            # Interactive digit canvas
├── pkg/training/           # Offline training utilities
├── cmd/mnist-gui/          # GUI entry point
└── cmd/train-network/      # Training tools
```

---

## Quick Start

### GUI Mode
```bash
fecim-lattice-tools mnist
```

### Headless Inference
```bash
fecim-lattice-tools --mode mnist --quantization-levels 30 --noise 0.05
```

### Draw Your Own Digit
1. Launch GUI
2. Use mouse to draw on canvas
3. Click "Recognize" to see FP vs CIM predictions

---

## What You'll Learn

1. **Neural Network Basics**
   - Fully-connected layers
   - ReLU activation
   - Softmax output
   - Forward propagation

2. **Quantization Effects**
   - Weight quantization (30 levels default)
   - Activation quantization
   - Accuracy vs precision trade-offs

3. **Hardware Non-Idealities**
   - Multiplicative noise (σ/µ model)
   - Quantization error accumulation
   - Layer-by-layer degradation

4. **CIM Performance**
   - Energy efficiency gains
   - Accuracy preservation techniques
   - Real-time inference speed

---

## Key Features

- **Dual-mode inference:** FP and CIM side-by-side
- **Interactive drawing canvas:** Test your own handwriting
- **Confidence visualization:** Compare prediction probabilities
- **Activation heatmaps:** See intermediate layer outputs
- **Confusion matrix:** Analyze classification errors
- **Energy widget:** Model-based energy comparison
- **Quantization control:** Adjust levels (8, 16, 30, 64)
- **Noise injection:** Simulate device variation (0-20%)

---

## Network Architecture

### Small MLP (for real-time visualization)

```
Input:  784 (28×28 pixels)
   ↓
Hidden: 128 neurons (ReLU)
   ↓
Output: 10 classes (Softmax)

Total parameters: ~100K
```

### Dual Inference Paths

```
┌─────────────────────────────────────────────┐
│           Input Image (28×28)               │
└──────────────┬──────────────┬───────────────┘
               │              │
     ┌─────────▼─────┐   ┌────▼──────────┐
     │  FP Path      │   │  CIM Path     │
     │ (float32)     │   │ (quantized)   │
     ├───────────────┤   ├───────────────┤
     │ W1_fp × x     │   │ Q(W1) × Q(x)  │
     │ + b1_fp       │   │ + Q(b1)       │
     │ ReLU          │   │ + noise       │
     │               │   │ ReLU          │
     ├───────────────┤   ├───────────────┤
     │ W2_fp × h     │   │ Q(W2) × Q(h)  │
     │ + b2_fp       │   │ + Q(b2)       │
     │ Softmax       │   │ + noise       │
     │               │   │ Softmax       │
     └───────┬───────┘   └────┬──────────┘
             │                │
             ▼                ▼
       FP Prediction    CIM Prediction
       (confidence)     (confidence)
```

---

## Quantization

### Weight Quantization (30 levels default)

```
Symmetric quantization to [-Wmax, +Wmax]:

Q(w) = -Wmax + round((w + Wmax) / (2*Wmax) * (L-1)) * (2*Wmax/(L-1))

where:
  L = 30 (quantization levels)
  Wmax = max(|weights|) in layer
```

### Noise Model

```
Multiplicative Gaussian noise:

w_noisy = w + N(0, 1) × |w| × (σ/µ)

where:
  σ/µ = noise coefficient (0.0 to 0.2)
  N(0,1) = standard normal random variable
```

---

## Accuracy Comparison

### Typical Results (30-level quantization)

| Mode | Noise | Accuracy | Notes |
|------|-------|----------|-------|
| **FP (baseline)** | 0% | 97.8% | Full precision |
| **CIM (ideal)** | 0% | 97.2% | Quantization only |
| **CIM (5% noise)** | 5% | 96.1% | Typical device variation |
| **CIM (10% noise)** | 10% | 94.3% | High variation |

*Note: These are model-based estimates with a small network. See `HONESTY_AUDIT.md` for validation status.*

---

## Energy Efficiency (Modeled)

```
Energy per inference (MACs × energy/MAC):

MACs = 784×128 + 128×10 = 101,632

FP (GPU):     101,632 × 10 pJ   ≈ 1,016 pJ
CIM (FeFET):  101,632 × 0.01 pJ ≈ 1 pJ

Efficiency gain: ~1000× (model-based)
```

---

## Documentation Index

| Document | Purpose | Audience |
|----------|---------|----------|
| [eli5.md](./eli5.md) | Beginner intro to MNIST | Beginners |
| [physics.md](./physics.md) | Network equations, quantization | Developers |
| [features.md](./features.md) | Workflows, extension points | Developers |
| [tools.md](./tools.md) | PyTorch/TF integration | Researchers |

---

## Evidence Status

- **Demonstrated:** Network architecture, quantization, dual-mode inference are implemented and verifiable
- **Modeled:** Energy estimates, accuracy predictions are simulation models
- **Aspirational:** Production deployment, hardware validation are future work

---

## Related Modules

- **[Module 1: Hysteresis](../module1-hysteresis/README.md)** - Conductance levels for weights
- **[Module 2: Crossbar](../module2-crossbar/README.md)** - MVM hardware for layer computation
- **[Module 5: Comparison](../module5-comparison/README.md)** - Architecture-level benchmarks

---

## Testing

```bash
# Run inference tests
go test ./module3-mnist/pkg/core

# Test quantization
go test -run TestQuantization ./module3-mnist/pkg/core

# Full integration test
go test ./module3-mnist/...
```

---

## Source Code

- **GitHub:** [module3-mnist/](../../../module3-mnist/)
- **Network:** `pkg/core/network_inference.go`
- **Quantization:** `pkg/core/quantize.go`
- **GUI:** `pkg/gui/dualmode.go`

---

**Last Updated:** 2026-02-16
**Maintainer:** FeCIM Lattice Tools Project
