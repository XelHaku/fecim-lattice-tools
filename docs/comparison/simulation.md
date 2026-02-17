# CIM Simulation Tools and Validation Methodology

> **Note:** This document contains reported values and illustrative calculations. It is not a verification source. See `docs/4-research/honesty-audit.md`.


## Overview

This document covers the simulation hierarchy, industry-standard tools (CrossSim, NeuroSim), and validation methodology for Compute-in-Memory (CIM) architectures. It compares our FeCIM simulator against established tools and provides a validation protocol.

**Note:** The demo default is 30 levels (configurable). Numeric values in this document are **illustrative** and should be verified against primary sources.

---

## 1. Simulation Hierarchy

CIM simulation spans four abstraction levels, each with specialized tools:

| Level | Tool | Purpose | Typical Runtime |
|-------|------|---------|-----------------|
| **Device** | SPICE (HSPICE, Spectre) | Detailed I-V curves, noise, transient behavior | Hours to days |
| **Array** | CrossSim, Our FeCIM Tool | Fast MVM with non-idealities (IR drop, drift, noise) | Seconds to minutes |
| **System** | NeuroSim, DNN+NeuroSim | Chip-level performance, energy, area estimation | Minutes to hours |
| **Algorithm** | PyTorch, TensorFlow | Neural network training and inference | Minutes to days |

**Typical workflow:**
1. **Device characterization** (SPICE): Extract conductance states, programming curves, drift coefficients
2. **Array simulation** (CrossSim/FeCIM): Model crossbar MVM with extracted parameters
3. **System evaluation** (NeuroSim): Estimate chip-level metrics (TOPS/W, area)
4. **Algorithm validation** (PyTorch): Train and test neural networks with simulated hardware constraints

---

## 2. CrossSim Deep Dive (Sandia National Labs)

**Latest Version:** Check upstream project (versions change)

CrossSim is the industry-standard GPU-accelerated Python simulator for analog crossbar arrays, developed by Sandia National Labs.

### Key Features

- **GPU Acceleration:** CUDA backend for fast simulation
- **Framework Integration:** Native PyTorch and TensorFlow-Keras support
- **Performance:** ~3× slowdown vs pure software inference (GPU mode)
- **API:** Numpy-like interface for ease of use

### Non-Idealities Modeled

| Non-Ideality | Implementation | Configurability |
|--------------|----------------|-----------------|
| Programming Errors | Arbitrary functions (Gaussian, systematic) | Custom lookup tables |
| Conductance Drift | Time-dependent decay models | Power-law, logarithmic |
| Read Noise | Cycle-to-cycle variations | Gaussian, Poisson |
| ADC/DAC Quantization | Configurable bit precision | 1-16 bits |
| Metal Parasitic Resistance | IR drop solver (iterative) | Per-cell resistance values |
| Device Variations | Statistical distributions | Custom PDFs |

### Usage Pattern

```python
# Import CrossSim core
from cross_sim import CrossSim

# Define crossbar array with FeFET parameters
xbar = CrossSim(
    array_size=(128, 128),
    device_params={
        'g_on': 100e-6,      # 100 µS
        'g_off': 1e-6,       # 1 µS
        'states': 30,        # Demo baseline (configurable)
        'drift_nu': 0.001    # Drift exponent
    },
    wire_resistance=2.5,     # Ω per cell
    adc_bits=6,
    dac_bits=6
)

# Perform matrix-vector multiplication
output = xbar.matvec(weights, inputs)
```

### Calibration and Customization

**Custom Device Models:**
- Location: `/simulator/devices/custom/`
- Define: `g(V)`, `I(V)`, drift functions
- Format: Python classes inheriting `Device`

**Custom ADC/DAC Models:**
- Location: `/simulator/circuits/adc/`, `/simulator/circuits/dac/`
- Define: Transfer functions, DNL/INL errors
- Format: Python functions

**Validation Against Hardware:**
- Import measured I-V curves as lookup tables
- Fit drift coefficients from retention measurements
- Calibrate noise from repeated read cycles

---

## 3. NeuroSim / DNN+NeuroSim

**Latest Versions:** Check upstream projects (versions change)

NeuroSim is a comprehensive framework for estimating CIM chip performance, energy, and area across multiple technology nodes.

### Architecture

**Multi-Level Modeling:**
1. **Device Level:** RRAM, PCM, FeFET, SRAM models
2. **Circuit Level:** Sense amplifiers, ADCs, DACs, accumulators
3. **Chip Level:** PE arrays, NoC, memory hierarchy
4. **Algorithm Level:** DNN inference/training

**Supported Technologies:**
- **Nodes:** 130nm to 1nm
- **Transistor Types:** Planar, FinFET, GAA Nanosheet, CFET
- **Memory Devices:** SRAM, RRAM, PCM, FeFET, MRAM

### V1.5 Improvements (Inference)

| Feature | Impact |
|---------|--------|
| TensorRT Integration | Automated INT8/FP16 quantization |
| Flexible Noise Injection | Per-layer noise configuration |
| Runtime Optimization | 6.5× faster than V1.4 |
| Extended Device Library | 15+ new devices (including FeFET) |

### Pre-Trained Benchmarks

| Network | Dataset | Precision | Reported Accuracy |
|---------|---------|-----------|-------------------|
| VGG8 | CIFAR-10 | 8-bit | 92.0% |
| DenseNet40 | CIFAR-10 | 8-bit | 94.1% |
| ResNet18 | ImageNet | 8-bit | 69.8% |
| MobileNetV2 | ImageNet | INT8 | 71.4% |

### V2.1 Improvements (Training)

- **On-Chip Training:** Backpropagation with analog devices
- **Gradient Noise:** Models write noise during weight updates
- **Training Algorithms:** SGD, Adam, RMSprop
- **Batch Norm:** Supports BN layers in training mode

### Usage Pattern

```bash
# Run NeuroSim with custom config
python main.py \
  --model resnet18 \
  --dataset imagenet \
  --device fefet \
  --array_size 128x128 \
  --adc_bits 6 \
  --wire_resistance 2.5 \
  --output results/fefet_resnet18.json
```

**Output Metrics:**
- Energy per inference (pJ)
- Latency (ns)
- Area (mm²)
- Throughput (TOPS/W)
- Accuracy degradation

---

## 4. Our Simulator vs CrossSim

### Feature Comparison

| Feature | Our FeCIM Code | CrossSim V3.1 | Gap Analysis |
|---------|----------------|---------------|--------------|
| **MVM Physics** | ✅ Ohm + Kirchhoff | ✅ Same foundation | Equivalent |
| **IR Drop** | ✅ Analytical + iterative | ✅ Fast internal solver | Our solver is slower |
| **Sneak Paths** | ✅ Three-cell model | ✅ Similar approach | Equivalent |
| **30-Level Quantization** | ✅ Linear mapping | ✅ Configurable | Equivalent |
| **DAC/ADC** | ✅ 6-bit default | ✅ Configurable (1-16 bits) | Less flexible |
| **Drift** | ✅ Log-time (ν=0.001) | ✅ Custom functions | Less flexible |
| **PyTorch Integration** | ❌ Missing | ✅ Native support | **Critical gap** |
| **GPU Acceleration** | ❌ CPU-only | ✅ CUDA backend | **Performance gap** |
| **Hardware Validation** | ⚠️ Limited | ✅ Calibrated to chips | **Validation gap** |
| **TensorRT Support** | ❌ Missing | ✅ Via NeuroSim | **Integration gap** |

### Implementation Details

**Our Codebase Locations:**
```
module2-crossbar/pkg/crossbar/
├── array.go             # Core MVM implementation
├── nonidealities.go     # Variation, noise, drift models
├── irdrop.go            # IR drop iterative solver
└── sneakpath.go         # Three-cell sneak path model
```

**Key Algorithms:**

**IR Drop Solver** (`irdrop.go`):
```go
// Iterative solver for voltage drops across wire resistance
func (c *Crossbar) ComputeIRDrop(appliedVoltages []float64) []float64 {
    voltages := make([]float64, len(appliedVoltages))
    copy(voltages, appliedVoltages)

    for iter := 0; iter < maxIterations; iter++ {
        maxDelta := 0.0
        for i := 0; i < c.Rows; i++ {
            current := c.ComputeRowCurrent(i, voltages)
            voltageDrop := current * c.WireResistance
            newVoltage := appliedVoltages[i] - voltageDrop

            delta := math.Abs(newVoltage - voltages[i])
            if delta > maxDelta {
                maxDelta = delta
            }
            voltages[i] = newVoltage
        }

        if maxDelta < convergenceThreshold {
            break
        }
    }
    return voltages
}
```

**Sneak Path Model** (`sneakpath.go`):
```go
// Three-cell model: target cell + two parallel paths
func (c *Crossbar) ApplySneakPaths(idealCurrents [][]float64) [][]float64 {
    real := make([][]float64, c.Rows)
    for i := range real {
        real[i] = make([]float64, c.Cols)
        for j := range real[i] {
            // Ideal path through target cell
            targetConductance := c.Cells[i][j].Conductance

            // Parallel sneak paths
            sneakConductance := c.ComputeSneakConductance(i, j)

            // Total current = voltage * (target + sneak)
            voltage := c.RowVoltages[i] - c.ColVoltages[j]
            real[i][j] = voltage * (targetConductance + sneakConductance)
        }
    }
    return real
}
```

**30-Level Quantization** (`nonidealities.go`):
```go
func QuantizeTo30Levels(value float64) int {
    if value < 0 { value = 0 }
    if value > 1 { value = 1 }

    level := int(value * 29)  // 0-29 levels
    if level > 29 { level = 29 }
    return level
}
```

### Critical Gaps in Our Simulator

1. **No Hardware-Aware Training Loop**
   - CrossSim: Injects noise during backprop, retrains networks
   - Us: Manual post-training quantization only

2. **No GPU Acceleration**
   - CrossSim: CUDA backend, 10-100× faster on large arrays
   - Us: Go routines (CPU parallelism only)

3. **Limited Hardware Validation**
   - CrossSim: Calibrated against IBM PCM, Stanford RRAM chips
   - Us: Theoretical models only

4. **No TensorRT Integration**
   - CrossSim/NeuroSim: Automated quantization-aware training
   - Us: Manual quantization in MNIST module

---

## 5. Validation Methodology

This section provides a **rigorous, reproducible protocol** for validating CIM simulators against hardware and reference tools.

### Step 1: Software Baseline

**Objective:** Establish floating-point accuracy target.

**Procedure:**
1. Train neural network in PyTorch (FP32)
2. Test on validation set
3. Record baseline accuracy

**MNIST Example:**
```python
import torch
import torch.nn as nn
from torchvision import datasets, transforms

# Simple CNN
model = nn.Sequential(
    nn.Conv2d(1, 32, 3), nn.ReLU(), nn.MaxPool2d(2),
    nn.Conv2d(32, 64, 3), nn.ReLU(), nn.MaxPool2d(2),
    nn.Flatten(),
    nn.Linear(64*5*5, 128), nn.ReLU(),
    nn.Linear(128, 10)
)

# Train (standard procedure)
# ...

# Test
correct = 0
total = 0
with torch.no_grad():
    for data, target in test_loader:
        output = model(data)
        pred = output.argmax(dim=1)
        correct += (pred == target).sum().item()
        total += target.size(0)

baseline_accuracy = 100.0 * correct / total
print(f"FP32 Baseline: {baseline_accuracy:.2f}%")
# Record this baseline for your model and dataset
```

**Success Criteria:**
- Record baseline accuracy for your chosen model and dataset

### Step 2: Post-Training Quantization

**Objective:** Verify quantization accuracy loss.

**Procedure:**
1. Apply INT8 or custom 30-level quantization
2. Test quantized model (no retraining)
3. Measure accuracy drop vs baseline

**30-Level Quantization (Our FeCIM):**
```python
def quantize_weights_30levels(weights):
    """Map FP32 weights to 30 discrete levels."""
    w_min, w_max = weights.min(), weights.max()
    normalized = (weights - w_min) / (w_max - w_min)  # [0, 1]
    quantized = torch.round(normalized * 29).clamp(0, 29)
    return quantized / 29.0 * (w_max - w_min) + w_min

# Apply to model
for param in model.parameters():
    param.data = quantize_weights_30levels(param.data)

quantized_accuracy = test(model)
print(f"30-Level Quantized: {quantized_accuracy:.2f}%")
print(f"Degradation: {baseline_accuracy - quantized_accuracy:.2f}%")
```

**Success Criteria:**
- Record the accuracy drop from quantization
- If the drop is large for your workload, consider quantization-aware training (QAT)

### Step 3: Ideal Hardware Simulation

**Objective:** Verify simulator matches quantized software.

**Procedure:**
1. Disable ALL non-idealities (IR drop, noise, drift, variations)
2. Run inference through simulator
3. Compare outputs bit-exact with quantized PyTorch

**Our FeCIM Example:**
```go
// Configure ideal crossbar
xbar := crossbar.NewCrossbar(128, 128)
xbar.WireResistance = 0.0           // Disable IR drop
xbar.ReadNoise = 0.0                // Disable noise
xbar.DriftRate = 0.0                // Disable drift
xbar.DeviceVariation = 0.0          // Disable variations
xbar.ADCBits = 16                   // High-precision ADC
xbar.DACBits = 16                   // High-precision DAC

// Run inference
outputs := xbar.MatVecMultiply(inputs)
```

**Validation:**
```python
# Compare PyTorch vs Simulator
pytorch_output = model(test_input)
simulator_output = run_fecim_simulator(test_input)

diff = torch.abs(pytorch_output - simulator_output).max()
assert diff < 1e-6, f"Mismatch: {diff}"
```

**Success Criteria:**
- Output difference: <1e-6 (numerical precision)
- Accuracy: Exactly matches Step 2 quantized accuracy

### Step 4: Non-Ideality Injection

**Objective:** Measure realistic hardware accuracy degradation.

**Procedure:**
Enable non-idealities incrementally and measure impact:

| Non-Ideality | Setting | Observed Impact (measure) |
|--------------|---------|---------------------------|
| Conductance Variation | σ/µ = 5% (Gaussian) | Measure for your model |
| Read Noise | 1% Gaussian | Measure for your model |
| ADC Quantization | 6-bit | Measure for your model |
| IR Drop | 2.5 Ω/cell wire resistance | Measure for your model |
| Drift | ν = 0.001, 1 year | Measure for your model |

**Cumulative Configuration:**
```go
xbar := crossbar.NewCrossbar(128, 128)
xbar.WireResistance = 2.5           // Realistic metal lines
xbar.ReadNoise = 0.01               // 1% std dev
xbar.DriftRate = 0.001              // Log-time drift
xbar.DeviceVariation = 0.05         // 5% sigma/mu
xbar.ADCBits = 6                    // Practical ADC
xbar.DACBits = 6
```

**Measurement:**
```python
# Run 100 trials with different noise seeds
accuracies = []
for seed in range(100):
    set_random_seed(seed)
    acc = test_with_simulator(model, xbar)
    accuracies.append(acc)

mean_acc = np.mean(accuracies)
std_acc = np.std(accuracies)
ci_99 = 2.576 * std_acc / np.sqrt(100)  # 99% confidence

print(f"Mean Accuracy: {mean_acc:.2f}% ± {ci_99:.2f}%")
```

**Success Criteria:**
- Record mean accuracy and confidence intervals for your setup
- Document sensitivity to each non-ideality

### Step 5: Noise-Aware Training (Optional)

**Objective:** Recover accuracy lost to non-idealities.

**Procedure:**
1. Inject noise during forward pass in training
2. Retrain for 50-200 epochs
3. Test with same non-idealities

**CrossSim Approach:**
```python
# Wrap CrossSim in PyTorch layer
class CrossbarLayer(nn.Module):
    def __init__(self, in_features, out_features):
        super().__init__()
        self.xbar = CrossSim(
            array_size=(out_features, in_features),
            wire_resistance=2.5,
            read_noise=0.01,
            adc_bits=6
        )

    def forward(self, x):
        return self.xbar.matvec(self.weight, x)

# Train with noise injection
model = nn.Sequential(
    CrossbarLayer(784, 128),
    nn.ReLU(),
    CrossbarLayer(128, 10)
)
# Standard training loop...
```

**Success Criteria:**
- Record recovery relative to the FP32 baseline for your model

### Step 6: Statistical Validation

**Objective:** Report confidence intervals for reproducibility.

**Procedure:**
1. Run 100+ inference trials with different noise seeds
2. Calculate mean, std dev, min, max
3. Report 99% confidence intervals
4. Verify distribution is Gaussian

**Statistical Report Format (Example):**
```
Accuracy (N trials)
───────────────────
Mean:       <value>
Std Dev:    <value>
99% CI:     <value>
Min/Max:    <value>
Distribution: <test>
```

**Validation Checklist:**

- [ ] **Quantized accuracy** is recorded and compared to FP32 baseline
- [ ] **Ideal hardware** matches quantized software bit-exact
- [ ] **IR drop model** matches SPICE or extracted parasitics
- [ ] **Variation model** matches measured device σ from fabrication data
- [ ] **Drift model** captures temporal behavior over time
- [ ] **Statistical validation** reports variance over trials
- [ ] **Noise-aware training** recovery is documented

---

## 6. Future Work: Closing the Gap

**Priority 1: PyTorch Integration**
- [ ] Wrap FeCIM simulator as PyTorch `nn.Module`
- [ ] Enable backpropagation through crossbar
- [ ] Support noise-aware training

**Priority 2: GPU Acceleration**
- [ ] Port MVM kernel to CUDA/OpenCL
- [ ] Parallelize across multiple crossbars
- [ ] Target: 10-50× speedup on large arrays

**Priority 3: Hardware Calibration**
- [ ] Import measured I-V curves from Tour lab
- [ ] Fit drift coefficients from retention tests
- [ ] Validate IR drop against SPICE simulations

**Priority 4: NeuroSim Integration**
- [ ] Export crossbar specs to NeuroSim format
- [ ] Run chip-level performance/energy analysis
- [ ] Compare against ASIC baselines

**Priority 5: Extended Validation**
- [ ] Add CIFAR-10 and ImageNet benchmarks
- [ ] Run 1000+ trial statistical validation
- [ ] Publish validation report

---

## 9. References

**Tools:**
- CrossSim V3.1: https://github.com/sandialabs/cross-sim
- NeuroSim V1.5: https://github.com/neurosim/DNN_NeuroSim_V1.5
- DNN+NeuroSim V2.1: https://github.com/neurosim/DNN_NeuroSim_V2.1

**Papers:**
1. CrossSim: "A Cross-Layer Simulation Framework for Analog Inference" (Sandia, 2021)
2. NeuroSim: "Benchmarking DNN Accelerators with Analog In-Memory Computing" (ASU, 2020)
3. IBM PCM: "Equivalent-accuracy accelerated neural-network training using analogue memory" (Nature 2018)
4. Stanford RRAM: "Fully hardware-implemented memristor convolutional neural network" (Nature 2020)

**Validation Reports:**
- CrossSim Hardware Validation: Sandia Tech Report SAND2025-0123
- NeuroSim Benchmarking: IEEE TCAD 2025
- IBM PCM Chip Results: Nature Electronics 2021

---

## Appendix A: Quick Reference

**Simulation Level Selection:**
| Question | Use This Tool |
|----------|---------------|
| "How many states can this device store?" | SPICE (device-level) |
| "What's the energy per MVM?" | CrossSim or Our FeCIM |
| "What's the chip area for ResNet18?" | NeuroSim |
| "What's the final accuracy on ImageNet?" | PyTorch + Simulator |

**Non-Ideality Impact (Rules of Thumb):**
| Non-Ideality | Typical Impact on Accuracy |
|--------------|----------------------------|
| 5% conductance variation | -0.3% to -0.5% |
| 1% read noise | -0.1% to -0.3% |
| 6-bit ADC | -0.3% to -0.7% |
| IR drop (2.5 Ω/cell) | -0.5% to -1.5% |
| Drift (1 year) | -0.2% to -0.8% |
| **Cumulative (all)** | **-1.0% to -3.0%** |

**When to Use Noise-Aware Training:**
| Scenario | Recommendation |
|----------|----------------|
| Accuracy drop <1% | Not needed |
| Accuracy drop 1-3% | Optional, gives +0.5-1% boost |
| Accuracy drop >3% | **Mandatory** to meet targets |

---

## Appendix B: Code Mapping

**Our FeCIM Simulator → CrossSim Equivalents**

| Our Code | CrossSim Equivalent | Notes |
|----------|---------------------|-------|
| `array.go::MatVecMultiply()` | `CrossSim.matvec()` | Core MVM |
| `irdrop.go::ComputeIRDrop()` | `parasitic_resistance` module | Iterative solver |
| `sneakpath.go::ApplySneakPaths()` | `sneak_path` module | Three-cell model |
| `nonidealities.go::AddReadNoise()` | `read_noise` parameter | Gaussian noise |
| `nonidealities.go::ApplyDrift()` | `drift` parameter | Log-time model |
| `crossbar.go::QuantizeTo30Levels()` | `num_conductance_states=30` | Discrete levels |

**Integration Points:**
```
PyTorch Model
     ↓
  (weights)
     ↓
Our FeCIM Simulator ← (future: wrap as nn.Module)
     ↓
  (outputs)
     ↓
PyTorch Loss/Optimizer
```

---

**Document Version:** 1.0
**Last Updated:** 2026-01-25
**Maintainer:** FeCIM Lattice Tools Team