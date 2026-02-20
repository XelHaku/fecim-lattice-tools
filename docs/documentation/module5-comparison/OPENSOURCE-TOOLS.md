<!-- Category: Open-Source Tools | Module: module5-comparison | Reading time: ~3 min -->
# Module 5 Open-Source Tools: Benchmarking and Comparison

> Tools for benchmarking and comparing computing architectures.

---

## Tools Used in This Module

| Tool | Role |
|------|------|
| Go toolchain | Model calculations and comparison engine |
| Fyne | GUI charts and ROI calculator |

---

## External Benchmarking Tools

These open-source tools are relevant for hardware benchmarking
beyond what the FeCIM comparison module models.

### Neural Network Benchmarks

| Tool | Description |
|------|-------------|
| **MLPerf** | Industry-standard benchmark suite for ML training and inference. Provides comparable results across hardware platforms. |
| **TVM / Apache TVM** | Compiler framework for ML models. Can target different hardware backends and report performance. |
| **ONNX Runtime** | Cross-platform inference engine. Useful for comparing the same model across CPU/GPU/accelerator. |

### Power and Energy Measurement

| Tool | Description |
|------|-------------|
| **RAPL (Intel)** | Running Average Power Limit -- read CPU/GPU energy counters on Intel platforms. |
| **nvidia-smi** | GPU power monitoring for NVIDIA cards. Reports real-time and average power draw. |
| **PowerTOP** | Linux power analysis tool. Shows per-component power breakdown. |

### CIM-Specific Simulation

| Tool | Description |
|------|-------------|
| **NeuroSim+** | Circuit-level CIM simulator from Georgia Tech. Models energy, latency, and area for various CIM architectures. |
| **MNSIM** | Memristor-based neural network simulator. Focuses on crossbar array non-idealities. |
| **CIMulator** | Open-source CIM architecture simulator with configurable peripherals. |

---

## Integration Notes

- The FeCIM comparison module uses analytic models, not NeuroSim+ or
  MNSIM. Integrating those tools would require translating the
  crossbar configuration to their input formats.
- Architecture definitions live in
  `module5-comparison/pkg/comparison/architecture.go`.
- The honesty audit (`docs/4-research/honesty-audit.md`) provides
  context for interpreting comparison results.

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
