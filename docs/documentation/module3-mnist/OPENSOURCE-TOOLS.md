<!-- Category: Tools | Module: module3-mnist | Reading time: ~3 min -->
# Module 3 Open-Source Tools: Neural Network Inference Ecosystem

> Tools for neural network training, hardware-aware inference, and
> integration with the MNIST digit recognition module.

---

## Tools Used by This Module

| Tool | Purpose | License |
|------|---------|---------|
| Go toolchain | Training and inference | BSD-style |
| Fyne | GUI rendering | BSD-3-Clause |

No external ML frameworks are integrated in this repo. Training and inference
are implemented in pure Go.

---

## Related Open-Source Tools

### AIHWKIT (IBM)

PyTorch toolkit for analog AI hardware simulation. Supports device-level
noise, drift, and hardware-aware training. If you want to train larger
networks with realistic hardware effects, AIHWKIT provides a mature
PyTorch-native path.

### NeuroSim (Georgia Tech)

Architecture-level benchmark for neuromorphic computing. Provides energy
and area estimates for crossbar-based neural networks. Useful for
estimating what this module's network would cost in real hardware.

### CrossSim (Sandia)

Crossbar array simulator that can run neural network inference with
device non-idealities. Useful for cross-validating accuracy results
from this module.

### TensorFlow Lite / ONNX Runtime

If you want to extend this module to support models trained in external
frameworks, TF Lite and ONNX Runtime provide quantized inference
engines that could serve as comparison baselines.

---

## Integration Notes

- Weight loading: `module3-mnist/pkg/core/network.go`
- Quantization utilities: `module3-mnist/pkg/core/quantize.go`
- To use externally-trained weights: export from PyTorch/TF as JSON arrays,
  load via the weight import path

---

## Tool Comparison

| Tool | Language | Strength | Use Case |
|------|----------|----------|----------|
| This Module (Go) | Go | Interactive GUI, dual-path visualization | Education |
| AIHWKIT | Python | Hardware-aware training, device models | Research |
| NeuroSim | C++ | Energy/area estimation | Benchmarking |
| CrossSim | Python | Non-ideality modeling | Validation |

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
