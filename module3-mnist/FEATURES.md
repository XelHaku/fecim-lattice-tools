# Module 3: MNIST - Features

Neural-network digit recognition demo with FP vs CIM comparison.

---

## Features

- **Dual-Mode Inference** - Side-by-side FP32 vs CIM (quantized + noise)
- **Drawing Canvas** - 28x28 input with brush sizes and smoothing
- **Quantization Controls** - Levels selector (shows only available QAT weights)
- **Noise Injection** - Gaussian multiplicative read noise (UI: 0-20%)
- **DAC/ADC Modeling** - Adjustable bit-depth (core 3-16; UI defaults to 8-bit)
- **Quick Demo** - Guided 5-step walkthrough (ideal -> failure modes -> recover)
- **Activation & Metrics** - Hidden-layer activations, confusion matrix, per-class stats
- **Energy Widget** - Model-based energy estimate (not measured hardware)
- **Weight Comparison** - FP vs quantized weight visualization

---

## GUI Variants

- **MNISTApp** (single-mode): activations + confusion matrix + metrics
- **DualModeApp** (FP vs CIM): side-by-side comparison + presets + quick demo

---

## Physics/Modeling (Core)

| Model | Description |
|---|---|
| **Weight Quantization** | Symmetric linear mapping to N discrete levels |
| **Read Noise** | Gaussian multiplicative (sigma/u) |
| **DAC Quantization** | Input voltage resolution (2^bits levels) |
| **ADC Quantization** | Output current resolution (2^bits levels) |
| **Energy Model** | Simple, configurable estimate for relative comparisons (not hardware) |

---

## Key Defaults

| Parameter | Default | Notes |
|---|---:|---|
| Levels | 30 | Demo baseline (configurable) |
| Noise sigma/u | 0.01 | UI default (1%) |
| ADC / DAC | 8 / 8 | UI defaults |
| Hidden Size | 128 | 784 -> 128 -> 10 network |

Defaults are simulation inputs, not measured hardware values.

---

## Accuracy Notes

- Reported accuracy is computed from the current weights and settings.
- External benchmarks are treated as literature notes and are **not** simulator claims. See `docs/comparison/HONESTY_AUDIT.md`.
