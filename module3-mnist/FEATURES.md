# Module 3: MNIST - Features

Neural Network Digit Recognition Demo

---

## Features

- **Dual-Mode Inference** - Side-by-side FP32 vs CIM (quantized + noise) comparison
- **Drawing Canvas** - 28×28 pixel digit input with 3 brush sizes
- **Quantization Control** - Levels selector (available QAT weights); per-layer PTQ supported in core
- **Noise Injection** - Adjustable Gaussian read noise (0-20% in UI)
- **DAC/ADC Simulation** - 3-16 bit resolution (core; fixed in Dual‑Mode UI)
- **Quick Demo** - 5-step automated walkthrough (includes 2-level quantization cliff)
- **Failure Mode Demos** - Binary weights + high noise via Quick Demo/Noisy preset
- **Layer Activation View** - Visualize hidden layer activations
- **Confusion Matrix** - Classification error heatmap
- **Weight Comparison Widget** - FP32 vs quantized weight visualization

## GUI Variants

- **MNISTApp** (single‑mode): activations + confusion matrix + per‑class metrics
- **DualModeApp** (FP vs CIM): comparison card, quantization/energy widgets, quick demo

## Physics Models

| Model | Description |
|-------|-------------|
| **Weight Quantization** | Symmetric linear mapping to N discrete levels |
| **Read Noise** | Gaussian multiplicative (σ/μ) |
| **DAC Quantization** | Input voltage resolution (2^bits levels) |
| **ADC Quantization** | Output current resolution (2^bits levels) |
| **Energy Model** | 10 fJ/bit per MAC (≈50 fJ @ 30-level baseline) + ADC/DAC overhead |

## Key Parameters

| Parameter | Default | Range |
|-----------|---------|-------|
| FeCIM Levels | 30 | 2-30 |
| Noise σ/μ | 0.01 (1%) | 0.0-0.20 (UI), 0.0-0.5 (core clamp) |
| ADC Bits | 8 | 3-16 (core) |
| DAC Bits | 8 | 3-16 (core) |
| Hidden Size | 128 | CLI/weights (GUI fixed) |
| Input | 784 | Fixed (28×28) |
| Output | 10 | Fixed (digits 0-9) |

## Network Architectures

- **Standard**: 784 → 128 (ReLU) → 10 (Softmax)
- **Calibration**: 784 → 10 (Softmax) — Single layer

## Accuracy Benchmarks

| Configuration | Accuracy |
|---------------|----------|
| FP32 (ideal) | ~98% |
| 30-level demo baseline (claim), low noise | 92-96% |
| Peer-reviewed FeCIM | 96.6-98.24% |
| 2 levels (binary) | ~50% |
