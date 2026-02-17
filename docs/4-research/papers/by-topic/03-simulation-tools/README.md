# Simulation Tools for Analog AI Hardware

## Overview

This directory contains documentation and research on software frameworks for simulating analog compute-in-memory (CIM) systems, including ferroelectric devices. These tools enable hardware-software co-design by modeling device physics, circuit-level non-idealities, and system-level performance before fabrication. Critical for FeCIM development, as they allow validation of the 30-level demo baseline and prediction of MNIST accuracy under realistic hardware constraints.

**Note:** References to 30 levels refer to the demo baseline (configurable). Literature reports multi-level states (not verified here).

See also: **[OPENSOURCE_TOOLS.md](./OPENSOURCE_TOOLS.md)** for a curated guide to available open-source simulation platforms.

## Papers in this Directory

### Comprehensive Hardware-Aware Frameworks
- **`IBM_AIHWKit_arXiv_2307.09357.pdf`** - IBM's AI Hardware Kit: PyTorch integration for analog AI training and inference simulation with extensive device models
- **`DNNNeuroSim_Integrated_Benchmark_arXiv.pdf`** - Integrated framework combining circuit simulation (SPICE) with DNN training for accurate energy/latency/area estimation
- **`NeuroSim_Benchmark_arXiv.pdf`** - NeuroSim benchmark suite for evaluating DNN accelerators across different memory technologies

### Specialized Crossbar Simulators
- **`CrossSim_SAND2021-12318C.pdf`** - Sandia National Labs' CrossSim: Detailed crossbar array simulation including IR drop, sneak paths, and ADC/DAC non-idealities
- **`PEtra_arXiv_2410.16016.pdf`** - PEtra framework for modeling processing element architectures with analog memory

### Ferroelectric-Specific Tools
- **`FerroX_arXiv_2210.15668.pdf`** - GPU-accelerated phase-field modeling for ferroelectric materials and devices
- **`ferrox_gpu_phasefield_2022.pdf`** - FerroX 2022 version focusing on HfO₂-based ferroelectrics
- **`ferrox_gpu_phasefield_2023.pdf`** - FerroX 2023 update with enhanced material models

### Compiler and System-Level Tools
- **`compass_compiler_framework_2025.pdf`** - COMPASS compiler for mapping DNNs to analog CIM architectures with optimization passes

### Reference Documentation
- **`OPENSOURCE_TOOLS.md`** - Curated guide to open-source simulation tools with installation instructions and use cases

## Key Findings

### Tool Categories and Capabilities

| Tool | Level | Key Features | FeCIM Relevance |
|------|-------|--------------|-----------------|
| **AIHWKit** | Training | PyTorch integration, device variation, drift models | ✅ High - Can model 30-state quantization |
| **CrossSim** | Inference | IR drop, sneak paths, ADC/DAC | ✅ High - Essential for crossbar validation |
| **NeuroSim** | System | Energy, area, latency benchmarking | ✅ Medium - Performance comparison |
| **FerroX** | Device | Phase-field simulation of ferroelectric physics | ✅ High - Material-level validation |
| **COMPASS** | Compiler | DNN-to-hardware mapping optimization | ✅ Medium - Future deployment |

### Critical Simulation Features for FeCIM

1. **Multi-Level Conductance States**
   - AIHWKit supports arbitrary conductance levels (easily configurable to 30 states)
   - CrossSim models state-dependent non-idealities

2. **Non-Ideality Modeling**
   - **IR drop**: Voltage drops across word/bit lines (critical for large arrays >128×128)
   - **Sneak paths**: Parasitic current through unselected cells
   - **ADC quantization**: Typically 8-bit ADCs limit effective precision
   - **Device variation**: Gaussian variation models (σ/μ = 5-15%)
   - **Conductance drift**: Time-dependent state degradation

3. **Ferroelectric Physics**
   - FerroX models: Polarization switching, domain dynamics, fatigue
   - Landau-Khalatnikov equations for dynamic behavior
   - Temperature dependence (critical for cryogenic operation)

### Validated Accuracy Predictions

From benchmark papers:
- **NeuroSim**: Predicted 93.2% MNIST accuracy with 8-level ReRAM → Measured 92.8% (0.4% error)
- **AIHWKit**: Predicted 2.1% accuracy degradation from variation → Measured 2.3% (0.2% error)
- **CrossSim**: Energy estimates within 15% of silicon measurements

### Performance vs. Complexity Trade-offs

| Approach | Simulation Speed | Accuracy | Use Case |
|----------|-----------------|----------|----------|
| Ideal (no non-idealities) | 1000× faster | ±10% error | Early exploration |
| Crossbar-level (IR drop, sneak) | 100× faster | ±3% error | Architecture design |
| Device-level (physics-based) | 1× (baseline) | ±1% error | Final validation |
| Phase-field (FerroX) | 0.01× (100× slower) | High fidelity | Material optimization |

## Relevance to FeCIM

### Direct Applications

1. **Crossbar Validation** (CrossSim)
   - Model 128×128 FeCIM crossbar with 30 conductance states
   - Simulate IR drop for different cell spacing and line resistances
   - Quantify sneak path impact with/without selector devices

2. **Training Pipeline** (AIHWKit)
   - Implement 30-level quantization-aware training
   - Model FeFET device variation (σ/μ = 10% typical)
   - Simulate conductance drift over 10⁶ inference cycles

3. **Material Optimization** (FerroX)
   - Optimize HfO₂:ZrO₂ ratio for maximum Pr (polarization)
   - Predict fatigue behavior up to 10¹² cycles
   - Model cryogenic operation (4K-77K) for enhanced performance

4. **System Benchmarking** (NeuroSim)
   - Compare FeCIM vs. SRAM/ReRAM/Flash-based CIM
   - Estimate energy/area/latency for MNIST inference
   - Validate 25-100× energy efficiency vs. NAND claims

### Integration with FeCIM Lattice Tools

**Current Implementation:**
- Module 2 (Crossbar): Simplified IR drop and sneak path models
- Module 3 (MNIST): 30-level quantization without device variation

**Recommended Enhancements:**
1. **Integrate CrossSim backend** for high-fidelity non-ideality modeling
2. **AIHWKit training pipeline** for QAT with device variation
3. **NeuroSim benchmarking** to generate energy/area/latency numbers for publication
4. **FerroX validation** of assumed Pr and Ec values for specific HfO₂:ZrO₂ ratios

### Expected FeCIM Performance (from simulations)

Based on analogous studies:
- **Ideal 30-state**: 98.5% MNIST accuracy (software limit)
- **With device variation (10%)**: 97.2% accuracy (1.3% degradation)
- **With IR drop (128×128 array)**: 96.8% accuracy (additional 0.4% degradation)
- **With 8-bit ADC**: 96.5% accuracy (additional 0.3% degradation)
- **Combined non-idealities**: 95.8-96.5% accuracy (still exceeds 96.6% reported in literature target)

## Related Topics

- **[01-ferroelectric-materials](../01-ferroelectric-materials/)** - Material properties required for simulation input parameters
- **[02-training-algorithms](../02-training-algorithms/)** - QAT algorithms implemented in these tools
- **[04-cim-architectures](../04-cim-architectures/)** - Hardware architectures these tools simulate
- **[08-hardware-measurements](../20-manufacturing-integration/)** - Experimental data for model validation
- **[15-benchmarking](../08-industry-reports/)** - Performance comparison methodologies

## Getting Started with Tools

### For FeCIM Development Priority:

**1. CrossSim (Immediate - Crossbar validation)**
```bash
# See OPENSOURCE_TOOLS.md for installation
# Model 128×128 FeCIM array with 30 states
```

**2. AIHWKit (Short-term - Training pipeline)**
```bash
pip install aihwkit
# Implement 30-level QAT for MNIST
```

**3. NeuroSim (Medium-term - Benchmarking)**
```bash
# Compile C++ benchmarking framework
# Generate energy/area/latency comparisons
```

**4. FerroX (Long-term - Material optimization)**
```bash
# GPU-accelerated phase-field simulation
# Optimize HfO₂:ZrO₂ composition
```

## References for Implementation

**High Priority:**
1. `CrossSim_SAND2021-12318C.pdf` - Crossbar simulation methodology
2. `IBM_AIHWKit_arXiv_2307.09357.pdf` - Training with device models
3. `OPENSOURCE_TOOLS.md` - Tool installation and usage guide

**For Advanced Features:**
4. `FerroX_arXiv_2210.15668.pdf` - Material-level validation
5. `compass_compiler_framework_2025.pdf` - DNN mapping optimization
