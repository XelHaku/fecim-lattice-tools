# MNIST FeCIM Demo - FP vs CIM Comparison

## Overview

This demo shows how a 784→128→10 neural network runs on ferroelectric crossbar arrays with a **30-level baseline** (conference claim; pending peer review). It features **dual-mode inference** comparing Full Precision (FP) vs Compute-in-Memory (CIM) paths, with peer-reviewed accuracy context in the UI (96.6–98.24%).

**Key Questions Answered:**
1. What is the 30-level baseline? (Physics + competitive advantage)
2. How do FP vs CIM results diverge? (Quantization + noise effects)
3. What happens when hardware degrades? (Quantization cliff, noise wall)
4. Why does this matter? (Verified energy-efficiency advantage range)

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
1. Click **Quick Demo** (5 steps) or draw a digit
2. Follow on-screen status and comparison card
3. Explore presets: Ideal → Hardware → Noisy

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
- **HZO Ferroelectric:** Demo baseline uses 30 states (conference claim); peer‑reviewed devices report 32–140 states
- **Domain Wall Pinning:** Natural quantization from crystal defects
- **ADC Resolution:** 6-bit (64 levels) → 30 reliably distinguishable

### Competitive Advantage

| Technology | Levels | Notes |
|------------|--------|-------|
| Flash (NAND) | 2-4 | TLC/QLC |
| ReRAM | 4-16 | Limited by variability |
| **FeCIM (HZO)** | **30 (conference-claim baseline)** | **5x better than ReRAM** |
| Ideal (FP32) | 2^32 | Baseline |

**Impact on MNIST (illustrative):**
- 2 levels (binary): ~50% accuracy (worse than random!)
- 8 levels: ~75%
- **30 levels (demo baseline): ~92–96% in simulation** (depends on noise/ADC/DAC)
- Float32: ~98% (theoretical)

### Why Not 64 Levels (6-bit ADC)?

Only 30 are reliably distinguishable due to:
1. Device-to-device variation (~2.75%)
2. Cycle-to-cycle variation (~1.5%)
3. Read noise (~0.5% σ/μ)

With 3σ separation requirement, 30 levels is a **conservative demo baseline** (practical limits vary by process).

---

## Hardware Reality Check

### Why CIM Diverges from FP

**Simulation (this demo):** With low noise and the 30-level baseline, CIM can approach FP accuracy.

**Hardware (literature):** Peer‑reviewed FeFET/FTJ work reports 96.6–98.24% (simulation); a conference‑only 87% claim exists but is unverified and not used as a target here.

**Why the gap?**

| Non-Ideality | Modeled Here | Notes |
|--------------|-------------|-------|
| Weight quantization | ✓ | 30-level symmetric quantization (demo baseline) |
| Read noise | ✓ | σ/μ multiplicative noise (configurable) |
| IR drop | ⚠️ Simplified | Not part of CIM inference path yet |
| Sneak paths | ⚠️ Simplified | Not part of CIM inference path yet |
| ADC non-linearity | ⚠️ Ideal | ADC modeled as uniform quantizer |
| Retention drift | ❌ Not modeled | Long-term drift out of scope |
| Cycle-to-cycle variation | ✓ | Covered by noise term |

**Takeaway:** Increasing noise, reducing levels, or lowering ADC resolution visibly degrades confidence and agreement. Exact mapping to a specific hardware demo depends on device variability and is not calibrated in this UI.

---

## Failure Modes (Quick Demo + Noise Preset)

### 1. Quantization Cliff (2 levels)

**Trigger:** Quick Demo step 4 (forces 2 levels via PTQ if no QAT weights exist).

**Result:** Accuracy collapses toward ~50% (worse than random).

**Why:** Binary weights cannot represent the 128‑dimensional weight space.

---

### 2. Noise Wall (> 0.10 noise)

**Preset Button:** "Noisy"

**Settings:** Levels=30, Noise=0.15 (high)

**Result:** Accuracy drops and confidence degrades.

**Why:** σ/μ noise corrupts analog currents before ADC readout.

---

### 3. ADC Quantization Artifacts (Advanced)

ADC resolution is fixed in the Dual‑Mode UI. To explore ADC artifacts, adjust `ADCBits` in code/CLI (e.g., 3‑bit).

---

## Energy Efficiency

### Energy Model (MAC-Level Estimate)

**Calculation (Jerry et al. IEDM 2017):**
- Energy per MAC: ~10 fJ/bit × log2(levels) (≈50 fJ @ 30-level demo baseline)
- MACs per inference: (784×128) + (128×10) = 101,632
- **FeCIM Energy:** 101,632 × 50 fJ ≈ **5.08 μJ** (plus small ADC/DAC overhead)

**GPU Baseline (NVIDIA V100):**
- Energy per MAC: ~500 pJ (DRAM fetch + compute)
- **GPU Energy:** 101,632 × 500 pJ = **50.8 mJ**

**Ratio:** Theoretical MAC-level ratio is large, but the **verified project claim** is **25–100×** efficiency vs NAND (Samsung Nature 2025). The UI uses that verified range.

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
| 30-level quantized (sim) | 96.8% | Quantize weights (demo baseline) |
| Conference-only claim (unverified) | ~87% | Dr. Tour (Nov 2024, not peer-reviewed) |

---

## Literature Context

### FeCIM in Research

| Paper | Architecture | Accuracy | Notes |
|-------|--------------|----------|-------|
| **This Demo** | 784→128→10 | **92–96% (sim, noise‑dependent)** | UI highlights peer‑reviewed baselines |
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
| Levels Select | QAT levels available in `module3-mnist/data/` | 30 | Only levels with trained weights are shown |
| Noise Slider | 0.0-0.20 | 0.01 | Gaussian noise σ/μ |

**Note:** ADC/DAC resolution and hidden size are fixed in the Dual‑Mode UI. They can be adjusted via code/CLI if needed.

### Preset Buttons

| Button | Levels | Noise | ADC | Effect |
|--------|--------|-------|-----|--------|
| Ideal | 30 | 0.01 | 8 | Best case (simulation) |
| Hardware | 30 | 0.03 | 8 | Production‑like noise (illustrative) |
| Noisy | 30 | 0.15 | 8 | High noise (accuracy drop) |

### Info Dialogs

Click **ℹ Info** to open a tabbed dialog with:
**Why 30 Levels?**, **Hardware Reality**, **Failure Modes**, and **About**.

---

## Quick Demo Script (5 Steps)

The **Quick Demo** button runs an automated walkthrough:

1. **Welcome** - Intro to FeCIM + 30‑level advantage
2. **Ideal** - Load a digit at the 30-level baseline (low noise)
3. **Success** - FP and CIM agree at the 30-level baseline
4. **Break It** - Switch to 2 levels (binary collapse)
5. **Restore** - Return to the 30-level baseline and conclude

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
│   ├── mnist/                # CLI (train/evaluate/interactive)
│   └── mnist-gui/            # Standalone GUI (MNISTApp)
├── pkg/
│   ├── core/                 # Dual-mode inference engine
│   │   ├── network.go        # DualModeNetwork + weight loading
│   │   ├── network_inference.go
│   │   ├── network_quantization.go
│   │   └── quantize.go       # Weight quantization + noise
│   │
│   ├── gui/                  # Fyne GUI components
│   │   ├── dualmode.go       # Dual-mode app (4-zone layout)
│   │   ├── dualmode_*.go     # Controls, inference, demo, weights
│   │   ├── dialogs.go        # Info dialogs
│   │   ├── comparison_card.go
│   │   ├── energy_widget.go
│   │   ├── weight_comparison.go
│   │   ├── embedded.go       # For unified visualizer
│   │   └── app.go            # Single-mode MNISTApp
│   │
│   ├── mnist/                # MNIST dataset loader
│   │   └── loader.go
│   │
│   └── training/             # Training utilities
│       └── network.go
│
├── data/
│   ├── pretrained_weights.json
│   ├── pretrained_weights_{N}.json    # Optional level-specific weights
│   ├── pretrained_weights_ptq.json
│   ├── single_layer_weights.json
│   └── *-idx*-ubyte.gz                 # MNIST dataset
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

With 3σ separation requirement, 30 levels is a **conservative demo baseline** (practical limits vary by process).

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
| **FeCIM** | **HZO FeFET** | **30 (conference-claim baseline)** | **50 fJ/MAC** | **TRL 4** |

FeCIM's advantage: 10× lower energy (fJ vs pJ), 5× more levels (30-level demo baseline vs 4-8).

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
- Use higher ADC bits (via code/CLI)
- Try "Ideal" preset for baseline

### GUI not responding

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
