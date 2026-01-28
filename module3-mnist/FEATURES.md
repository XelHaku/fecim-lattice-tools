# Module 3: MNIST - Features

## Features

- **Dual-Mode Inference** — Side-by-side FP32 vs CIM (quantized + noise) comparison
- **Drawing Canvas** — 28×28 pixel digit input with 3 brush sizes
- **Quantization Control** — 2-30 levels, per-layer PTQ support
- **Noise Injection** — Adjustable Gaussian read noise (0-50%)
- **DAC/ADC Simulation** — 3-16 bit resolution
- **Educational Tour** — 5-step guided demo + 30-second quick demo
- **Failure Mode Demos** — Binary weights, high noise demonstrations

## Physics Models

| Model | Description |
|-------|-------------|
| **Weight Quantization** | Symmetric linear mapping to N discrete levels |
| **Read Noise** | Gaussian multiplicative (Johnson noise model) |
| **DAC Quantization** | Input voltage resolution (2^bits levels) |
| **ADC Quantization** | Output current resolution (2^bits levels) |
| **Energy Model** | 10 fJ/bit per MAC (50 fJ @ 30 levels) |

## Key Parameters

| Parameter | Default | Range |
|-----------|---------|-------|
| FeCIM Levels | 30 | 2-30 |
| Noise σ/μ | 0.01 (1%) | 0.0-0.5 |
| ADC Bits | 8 | 3-16 |
| DAC Bits | 8 | 3-16 |
| Hidden Size | 128 | 32-512 |
| Input | 784 | Fixed (28×28) |
| Output | 10 | Fixed (digits 0-9) |

## Network Architectures

- **Standard**: 784 → 128 (ReLU) → 10 (Softmax)
- **Calibration**: 784 → 10 (Softmax) — Single layer

## Accuracy Benchmarks

| Configuration | Accuracy |
|---------------|----------|
| FP32 (ideal) | ~98% |
| 30 levels, low noise | 92-96% |
| Peer-reviewed FeCIM | 96.6-98.24% |
| 2 levels (binary) | ~50% |
