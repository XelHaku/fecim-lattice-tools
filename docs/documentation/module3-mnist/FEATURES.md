<!-- Category: Features | Module: module3-mnist | Reading time: ~4 min -->
# Module 3 Features: MNIST Digit Recognition

> Feature reference for the neural network module -- dual-path inference,
> visualization, and hardware-effect controls.

---

## Feature Reference Table

| Feature | GUI Location | What It Shows | Model |
|---------|-------------|---------------|-------|
| Dual-Path Inference | Main panel | Side-by-side full-precision vs CIM results | MLP with quantization |
| Confidence Bars | Output panel | Per-digit probability comparison | Softmax output |
| Drawing Canvas | Input panel | Draw digits live for immediate inference | Direct pixel input |
| Noise Slider | Controls panel | Adjust sigma/mu from 0.00 to 0.20 | Multiplicative Gaussian |
| Quantization Level | Controls panel | Number of discrete weight levels | Uniform symmetric |
| Activation Visualization | Hidden layer panel | See which neurons fire for each input | ReLU activations |
| Accuracy Counter | Status bar | Running accuracy over test samples | Correct / total |
| Analysis Tab | Analysis panel | Accuracy sweeps across levels, bits, noise | SweepQuantizationLevels() |

---

## Key Workflows

### 1. Test Set Inference

Load pretrained weights, run inference on MNIST test images (10,000), compare
full-precision and CIM accuracy side by side.

### 2. Live Drawing

Draw a digit on the canvas. The network classifies it in real time through
both paths. See which digits the CIM path confuses.

### 3. Parameter Exploration

Adjust noise level and quantization levels. Watch accuracy change. The
analysis tab shows sweep curves (accuracy vs levels, accuracy vs noise)
as ASCII bar charts.

---

## Code Layout

### Inference Path (Runtime)

| Component | Location |
|-----------|----------|
| Dual-mode network | `module3-mnist/pkg/core/` |
| Quantization utilities | `module3-mnist/pkg/core/quantize.go` |
| Energy model | `module3-mnist/pkg/core/` |
| GUI widgets | `module3-mnist/pkg/gui/` |
| MNIST data loader | `module3-mnist/pkg/mnist/` |
| Accuracy sweep | `shared/neural/accuracy_sweep.go` |

### Training Path (Offline)

| Component | Location |
|-----------|----------|
| Training network | `module3-mnist/pkg/training/` |
| Train CLI | `module3-mnist/cmd/train-network/main.go` |
| PTQ CLI | `module3-mnist/cmd/train-ptq/main.go` |

---

## Running Module 3

```bash
# GUI (default)
./fecim-lattice-tools mnist

# CLI inference
./fecim-lattice-tools mnist --headless
```

---

## Extension Points

- Add new weight files with different quantization-aware training
- Extend visualization for more layers or activation metrics
- Integrate other datasets (Fashion-MNIST, CIFAR-10) for comparison
- Add new quantization schemes (per-layer, non-uniform)

---

## Known Limitations

1. Network is intentionally small (784-128-10) for interactivity, not accuracy
2. Training is offline; GUI focuses on inference only
3. Hardware non-idealities are simplified and modeled, not device-calibrated
4. Noise slider range (0.00-0.20) is clamped in core to match UI/docs
5. No ADC/DAC quantization at this level (handled in Module 2 crossbar)

---

## Integration with Other Modules

| Source Module | What It Provides |
|--------------|-----------------|
| Module 2 (Crossbar) | Crossbar arrays execute MVM in network layers |
| Module 1 (Hysteresis) | Polarization levels map to quantized weights |
| Module 4 (Circuits) | DAC/ADC resolution affects effective quantization |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
