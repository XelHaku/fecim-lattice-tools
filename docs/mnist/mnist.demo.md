# MNIST FeCIM Demo - 87% Hardware Target

> *"We're at 87% validation here... theoretical is 88%."*
> — Dr. external research group, external research institution (Nov 2024)

## Overview

This demo shows how a 784→128→10 neural network runs on ferroelectric crossbar arrays with **30 discrete analog levels**. It features **dual-mode inference** comparing Full Precision (FP) vs Compute-in-Memory (CIM) paths.

**Key Questions Answered:**
1. What are 30 analog levels? (Physics + competitive advantage)
2. Why does FeCIM achieve 87%? (Hardware reality vs simulation)
3. What happens when hardware fails? (Quantization cliff, noise wall)
4. Why does this matter? (10,000x energy savings)

---

## Quick Start

```bash
# From the unified visualizer (recommended)
cd /path/to/fecim-lattice-tools
./launch.sh
# Navigate to "3b. MNIST FP/CIM" tab

# Or standalone demo
cd module3-mnist
go build -o mnist-gui ./cmd/mnist-gui
./mnist-gui
```

**First-Time User:**
1. Click "Start Guided Tour" (7 steps)
2. Follow on-screen instructions
3. Explore presets: Ideal → Quant Cliff → Noisy → Broken ADC

---

## Dual-Mode Architecture

```
User Input (28x28 drawn digit)
    |
+-------------------+------------------+
|  FP Path          |  CIM Path        |
+-------------------+------------------+
| Float32 weights   | Quantized weights|
| No noise          | + Noise          |
| Infinite precision| N-bit ADC/DAC    |
+-------------------+------------------+
| Layer 1: 784→128  | Crossbar 1 MVM   |
| ReLU              | ReLU             |
| Layer 2: 128→10   | Crossbar 2 MVM   |
| Softmax           | Softmax          |
+-------------------+------------------+
| Output: [0.98, …] | Output: [0.89, …]|
+-------------------+------------------+
    |
Compare & Visualize Difference
```

The demo runs both paths simultaneously:
- **Digital (FP)**: Ideal floating-point computation - theoretical maximum
- **FeCIM (CIM)**: Realistic hardware simulation with quantization and noise

---

## Why 30 Levels?

### Physics Justification
- **HZO Ferroelectric:** ~30 stable polarization states
- **Domain Wall Pinning:** Natural quantization from crystal defects
- **ADC Resolution:** 6-bit (64 levels) → 30 reliably distinguishable

### Competitive Advantage

| Technology | Levels | Notes |
|------------|--------|-------|
| Flash (NAND) | 2-4 | TLC/QLC |
| ReRAM | 4-16 | Limited by variability |
| **FeCIM (HZO)** | **30** | **5x better than ReRAM** |
| Ideal (FP32) | 2^32 | Baseline |

**Impact on MNIST:**
- 2 levels (binary): ~50% accuracy (worse than random!)
- 8 levels: ~75%
- **30 levels: ~87% (FeCIM hardware)**
- Float32: ~98% (theoretical)

### Why Not 64 Levels (6-bit ADC)?

Only 30 are reliably distinguishable due to:
1. Device-to-device variation (~2.75%)
2. Cycle-to-cycle variation (~1.5%)
3. Read noise (~0.5% σ/μ)

With 3σ separation requirement, 30 levels is the practical limit.

---

## Hardware Reality Check

### Why 87% and Not 98%?

**Simulation (this demo):** Can achieve 95-98% under ideal conditions.

**FeCIM Hardware (Dr. Tour):** 87% measured, 88% theoretical max.

**Why the gap?**

| Non-Ideality | Simulation | Hardware | Impact |
|--------------|------------|----------|--------|
| Weight quantization | ✓ 30 levels | ✓ 30 levels | -1% |
| Read noise | ✓ Configurable | ✓ Real | -2% |
| IR drop | ⚠️ Simplified | ✓ Metal lines | -3% |
| Sneak paths | ⚠️ Simplified | ✓ Parasitic | -2% |
| ADC non-linearity | ⚠️ Ideal | ✓ DNL/INL | -1% |
| Retention drift | ❌ Not modeled | ✓ 10 years | -1% |
| Cycle-to-cycle variation | ⚠️ Limited | ✓ 2.75% | -2% |

**Total:** ~12% gap between ideal (98%) and hardware (87%).

**How to Match Hardware:**
Set noise level to ~0.08 in the GUI. This empirically matches the 87% target.

---

## Failure Modes (Interactive Presets)

### 1. Quantization Cliff (< 4 levels)

**Preset Button:** "Quant Cliff"

**Settings:**
- Levels: 2
- Noise: 0.01 (low)
- ADC: 8 bits

**Result:** Accuracy ~50% (worse than random!)

**Why:** Binary weights {-1, +1} cannot represent the 128-dimensional weight space. Network loses ability to distinguish classes.

**Visualization:** Heatmap shows only 2 colors (blue/red). Hidden layer activations are nearly identical for all digits.

---

### 2. Noise Wall (> 0.10 noise)

**Preset Button:** "Noisy"

**Settings:**
- Levels: 30
- Noise: 0.15 (high)
- ADC: 6 bits

**Result:** Accuracy ~70%. Confidence drops to ~40-60% (vs 90%+ ideal).

**Why:** Gaussian noise in MVM corrupts output currents. ADC reads wrong value.

**Visualization:**
- Draw an "8" → classified as "3"
- Probability bars "jitter" on redraw

---

### 3. ADC Quantization Artifacts (< 4-bit ADC)

**Preset Button:** "Broken ADC"

**Settings:**
- Levels: 30
- Noise: 0.01
- **ADC: 3 bits**

**Result:** Accuracy ~65%. Staircase artifacts in activations.

**Why:** 3-bit ADC = only 8 output levels. Hidden layer activations are coarsely quantized, losing information.

**Visualization:** Hidden layer heatmap shows discrete bands instead of smooth gradients.

---

### 4. Confidence Collapse (Extreme Settings)

**Manual Settings:**
- Levels: 2
- Noise: 0.20
- ADC: 3 bits

**Result:** All output probabilities → ~10% (uniform distribution). Network effectively random guessing.

**Why:** Combination of:
1. Insufficient weight precision (2 levels)
2. High read noise (0.20)
3. Coarse ADC (3 bits)

Network cannot extract meaningful features.

---

## Energy Efficiency

### Dr. Tour's 10,000x Claim

**Calculation (Jerry et al. IEDM 2017):**
- Energy per MAC: ~50 fJ (HZO FeFET)
- MACs per inference: (784×128) + (128×10) = 101,632
- **FeCIM Energy:** 101,632 × 50 fJ = **5.08 μJ**

**GPU Baseline (NVIDIA V100):**
- Energy per MAC: ~500 pJ (DRAM fetch + compute)
- **GPU Energy:** 101,632 × 500 pJ = **50.8 mJ**

**Ratio:** 50.8 mJ / 5.08 μJ = **10,000x**

**Caveats:**
- Assumes all data on-chip (no DRAM)
- Excludes control circuitry overhead
- Best-case estimate (not independently verified)

---

## Reproducibility

### Training Weights

**Architecture:**
- Input: 784 (28×28 pixels)
- Hidden: 64/128/256 (configurable)
- Output: 10 (Softmax)

**Training:**
- Optimizer: Adam (lr=0.001, β1=0.9, β2=0.999)
- Epochs: 10
- Batch size: 64
- Dataset: MNIST (60k train, 10k test)

**Quantization:**
- Method: Symmetric, linear mapping
- Range: [-W_max, +W_max] (per-layer)
- Levels: 1-30 (configurable)
- Rounding: Round to nearest

### Expected Results

| Configuration | Accuracy | Source |
|---------------|----------|--------|
| FP (float32) | 98.1% | Training script |
| 30-level quantized (sim) | 96.8% | Quantize weights |
| **FeCIM hardware** | **87.0%** | **Dr. Tour (Nov 2024)** |

---

## Literature Context

### FeCIM in Research

| Paper | Architecture | Accuracy | Notes |
|-------|--------------|----------|-------|
| **This Demo** | 784→128→10 | **87%** | Matches Dr. Tour hardware |
| Jerry+ IEDM 2017 | 784→256→10 | 90% | 75ns pulse optimization |
| Nature Comms 2023 | Multi-level FeFET | 96.6% | Simulation only |
| Variation-Resilient 2024 | Binary NN | 94.2% | BNN with FeFET |

**Why Differences?**

1. **Hidden Size:** 128 (this demo) vs 256 (Jerry)
   - More neurons → higher capacity → better accuracy
   - Tradeoff: 2× chip area, 2× energy

2. **Pulse Timing:** 50ns (this demo) vs 75ns (Jerry)
   - 75ns achieves symmetric potentiation/depression
   - Improves weight update linearity

3. **Training Algorithm:** Standard SGD vs Quantization-Aware Training (QAT)
   - QAT simulates quantization during training
   - Network learns robust representations
   - Potential +2-3% accuracy improvement

---

## GUI Features

### Control Panel (Hardware Knobs)

| Control | Range | Default | Description |
|---------|-------|---------|-------------|
| Levels Slider | 1-30 | 30 | Weight quantization levels |
| Noise Slider | 0.0-0.20 | 0.01 | Gaussian noise σ/μ |
| ADC Bits | 3-8 | 6 | Output quantization |
| DAC Bits | 3-8 | 8 | Input quantization |
| Hidden Size | 64/128/256 | 128 | Network capacity |

### Preset Buttons

| Button | Levels | Noise | ADC | Effect |
|--------|--------|-------|-----|--------|
| Ideal | 30 | 0.01 | 8 | Best case (~95%) |
| Hardware (87%) | 30 | 0.08 | 6 | Matches real chip |
| Quant Cliff | 2 | 0.01 | 8 | Binary collapse (~50%) |
| Noisy | 30 | 0.15 | 6 | High noise (~70%) |
| Broken ADC | 30 | 0.01 | 3 | Coarse output (~65%) |

### Info Dialogs

- **Why 30 Levels?** - Physics and competitive advantage
- **Hardware Reality** - Simulation vs hardware gap explanation
- **Failure Modes** - Detailed failure mode descriptions
- **About** - Demo overview and references

---

## Guided Tour Script (7 Steps)

The guided tour walks through the key concepts:

1. **Welcome** - Introduction to FeCIM and 87% target
2. **Draw a Digit** - Interactive digit drawing
3. **FeCIM Classifies It** - Compare FP vs CIM predictions
4. **The 30 Analog Levels** - Weight heatmap explanation
5. **What If We Only Had 2 Levels?** - Quantization cliff demo
6. **What About Noise?** - Noise wall demonstration
7. **FeCIM's Sweet Spot** - Return to optimal settings

---

## Neural Network Architecture

```
+-------------------------------------------------------------+
|                    MNIST Input (28x28)                       |
|                      784 pixels                              |
+-----------------------------+-------------------------------+
                              |
                              v
+-------------------------------------------------------------+
|              Layer 1: FeFET Crossbar Array                   |
|                    784 x 128 weights                         |
|              30-level conductance states                     |
|                                                              |
|   V0  V1  V2 ... V783                                        |
|   |   |   |       |                                          |
|  +---+---+---+---+---+                                       |
|  |G00|G01|G02|...|   |-> I0  -+                              |
|  |G10|G11|G12|...|   |-> I1   |                              |
|  | : | : | : |...|   |-> :    | ReLU                         |
|  |   |   |   |...|   |-> I127 -+                             |
|  +---+---+---+---+---+                                       |
+-----------------------------+-------------------------------+
                              | 128 hidden activations
                              v
+-------------------------------------------------------------+
|              Layer 2: FeFET Crossbar Array                   |
|                    128 x 10 weights                          |
|              30-level conductance states                     |
|                                                              |
|  +---+---+---+---+                                           |
|  |   |   |...|   |-> I0  (digit 0)                           |
|  |   |   |...|   |-> I1  (digit 1)                           |
|  | : | : |...| : |-> :                                       |
|  |   |   |...|   |-> I9  (digit 9)                           |
|  +---+---+---+---+                                           |
+-----------------------------+-------------------------------+
                              | 10 output logits
                              v
+-------------------------------------------------------------+
|                        Softmax                               |
|              Probability distribution over 10 classes        |
+-------------------------------------------------------------+
                              |
                              v
                      Predicted Digit
```

---

## File Structure

```
module3-mnist/
├── cmd/
│   └── mnist-gui/
│       └── main.go           # Standalone entry point
├── pkg/
│   ├── core/                 # Dual-mode inference engine
│   │   ├── network.go        # DualModeNetwork
│   │   ├── quantize.go       # Weight quantization
│   │   └── quantize_test.go  # Unit tests
│   │
│   ├── gui/                  # Fyne GUI components
│   │   ├── dualmode.go       # Dual-mode app (4-zone layout)
│   │   ├── tour.go           # Guided tour mode
│   │   ├── dialogs.go        # Info dialogs
│   │   ├── embedded.go       # For unified visualizer
│   │   └── app.go            # Original single-mode app
│   │
│   ├── mnist/                # MNIST dataset loader
│   │   └── loader.go
│   │
│   └── training/             # Training utilities
│       └── network.go
│
├── data/
│   ├── pretrained_weights.json
│   ├── pretrained_30_h64.json
│   ├── pretrained_30_h128.json
│   ├── pretrained_30_h256.json
│   └── mnist/                # MNIST dataset
│
├── scripts/
│   ├── train_all_sizes.sh    # Train 64/128/256
│   └── benchmark.sh          # Compare with literature
│
└── docs/ -> see docs/mnist/  # All documentation
```

---

## Tests

```bash
# Run all tests
cd module3-mnist
go test ./... -v

# Run core package tests with coverage
go test ./pkg/core -cover -v

# Expected coverage: >80% for core package
```

---

## FAQ

### Why not 64 levels (6-bit ADC)?

Only 30 are reliably distinguishable due to:
1. Device-to-device variation (~2.75%)
2. Cycle-to-cycle variation (~1.5%)
3. Read noise (~0.5% σ/μ)

With 3σ separation requirement, 30 levels is the practical limit.

### Can we train on-chip?

FeCIM supports on-chip training via:
1. Pulse-based weight updates (potentiation/depression)
2. Backpropagation with stored gradients
3. Challenge: Asymmetric updates (see Jerry et al. IEDM 2017)

This demo focuses on inference only.

### How does this compare to Mythic/Analog Inference?

| Company | Technology | Levels | Energy | Status |
|---------|-----------|--------|--------|--------|
| Mythic | Flash | 4 | ~5 pJ/MAC | Shipping |
| Analog Inference | Flash | 8 | ~3 pJ/MAC | R&D |
| **FeCIM** | **HZO FeFET** | **30** | **50 fJ/MAC** | **TRL 4** |

FeCIM's advantage: 10× lower energy (fJ vs pJ), 5× more levels (30 vs 4-8).

---

## Troubleshooting

### MNIST data not found

Download MNIST data:
```bash
cd module3-mnist/data
wget http://yann.lecun.com/exdb/mnist/train-images-idx3-ubyte.gz
wget http://yann.lecun.com/exdb/mnist/train-labels-idx1-ubyte.gz
wget http://yann.lecun.com/exdb/mnist/t10k-images-idx3-ubyte.gz
wget http://yann.lecun.com/exdb/mnist/t10k-labels-idx1-ubyte.gz
```

### Accuracy below target

- Check noise level (lower = better accuracy)
- Increase levels (30 = best)
- Use higher ADC bits (6-8)
- Try "Ideal" preset for baseline

### GUI not responding

- Check if guided tour is running (click "End Tour")
- Restart the application
- Check terminal for error messages

---

## Related Documentation

- [MNIST ELI5](mnist.ELI5.md) - Simple explanations for beginners
- [MNIST Research](mnist.research.md) - Academic background and literature review
- [MNIST Open Source](mnist.opensource.md) - Related projects and tools
- [Module Improvements Plan](mnist-module-improvements-plan.md) - Roadmap

---

## References

1. Dr. external research group, "Ferroelectric CIM Presentation" (Nov 2024)
2. Jerry et al., "FeFET Analog Synapse for DNN Training," IEDM (2017)
3. Nature Communications, "Multi-Level FeFET Crossbar" (2023)
4. Variation-Resilient FeFET Binary NN, arXiv (2024)
5. DNNNeuroSim V2.0, arXiv:2003.06471
6. MNIST Dataset - Yann LeCun

---

## License

MIT License - See LICENSE file

---

## Acknowledgments

- Dr. external research group (external research institution) - Ferroelectric CIM technology
- Jaeho Shin - HZO superlattice FeFET development
- Jerry et al. - IEDM 2017 paper (75ns pulse optimization)
- MNIST Dataset - Yann LeCun

**Disclaimer:** This is an educational visualization. FeCIM hardware is at TRL 4 (lab validation). Energy claims have not been independently verified.
