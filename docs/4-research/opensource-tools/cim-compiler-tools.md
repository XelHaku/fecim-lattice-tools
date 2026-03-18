# CIM Compiler & Mapping Frameworks

**Open-Source Compiler and Mapping Tools for Compute-in-Memory Accelerators**

*Last Updated: March 2026*

---

## Overview

Traditional DNN compilers (TVM, XLA, TensorRT) target architectures with separate memory and compute units. CIM accelerators require specialized compilers that handle weight-to-crossbar mapping, crossbar tiling, precision allocation, ADC/DAC scheduling, and non-ideality-aware optimization. This document catalogs open-source frameworks addressing these needs.

### Why CIM Compilers Matter for FeCIM

```
Traditional DNN compiler:           CIM-specific compiler:
  Model → Operator graph              Model → Weight mapping
  → Instruction scheduling            → Crossbar tile partitioning
  → Register allocation               → ADC/DAC precision allocation
  → Machine code                      → Analog-aware scheduling
                                       → Non-ideality compensation
```

---

## 1. CIM-MLC (Multi-Level Compilation for CIM)

### Overview
Multi-level compilation framework that transforms DNN computing into row-wise read/write operations on CIM crossbars and applies multi-level scheduling strategies to optimize DNN inference.

- **Paper:** "CIM-MLC: A Multi-level Compilation Stack for Computing-In-Memory Accelerators" (2024)
- **arXiv:** [2401.12428](https://arxiv.org/abs/2401.12428)
- **Status:** Research prototype

### Key Features
- Transforms DNN compute graph into CIM-specific operations
- Multi-level scheduling: operator-level, tile-level, instruction-level
- Weight mapping optimization for crossbar arrays
- Supports both SRAM-based and NVM-based CIM
- Accounts for ADC/DAC quantization effects

### FeCIM Relevance: HIGH
Directly applicable to mapping Module 3 MNIST networks onto FeCIM crossbar tiles. The multi-level scheduling could optimize write-verify cycles for ISPP programming.

### Limitations
- Research prototype — not production-ready
- Limited to inference (no training support)
- No ferroelectric-specific device models

---

## 2. COMPASS Apache TVM (Arm China)

### Overview
Enhanced Apache TVM fork for wide-range Neural Network model support, optimization, and heterogeneous execution, including CIM accelerator backends.

- **GitHub:** https://github.com/Arm-China/Compass_Apache_TVM
- **License:** Apache-2.0
- **Status:** Active development

### Key Features
- Graph partition via TVM BYOC (Bring Your Own Codegen) framework
- Heterogeneous execution across CPU, GPU, CIM accelerators
- Optimized weight quantization and mapping
- Operator fusion for CIM-compatible graphs
- Hardware-specific calibration hooks

### FeCIM Relevance: MEDIUM
Provides the compiler infrastructure to map DNNs to FeCIM hardware. The BYOC framework could integrate FeCIM-specific backends for crossbar mapping.

### Limitations
- Primarily designed for Arm's Zhouyi NPU, requires adaptation for FeCIM
- No ferroelectric device models built-in
- Documentation mostly focused on Arm hardware

---

## 3. EasyACIM (DAC 2024)

### Overview
End-to-end automated analog CIM with synthesizable architecture and agile design space exploration.

- **Conference:** DAC 2024
- **Status:** Research publication

### Key Features
- Automated analog CIM architecture generation
- Design space exploration for area/power/accuracy trade-offs
- Synthesizable RTL output compatible with EDA tools
- Support for various ADC topologies
- Crossbar array sizing optimization

### FeCIM Relevance: HIGH
Could generate Module 4 peripheral circuit configurations automatically and explore DAC/ADC trade-offs for FeCIM arrays.

### Limitations
- Conference paper — code availability unclear
- Focused on SRAM-based CIM (needs ferroelectric adaptation)

---

## 4. SEGA-dcim (DATE 2025)

### Overview
Design space exploration-guided automatic digital CIM compiler with multiple precision support.

- **Conference:** DATE 2025
- **Status:** Research publication

### Key Features
- Multiple precision support (2-bit to 8-bit)
- Automatic design space exploration
- Digital CIM architecture generation
- Energy-accuracy Pareto optimization

### FeCIM Relevance: MEDIUM
The multi-precision support (2-8 bit) overlaps with FeCIM's configurable quantization levels. Design space exploration methodology transferable.

### Limitations
- Digital CIM focus (not analog/ferroelectric)
- May not handle analog non-idealities

---

## 5. C4CAM (ASPLOS 2024)

### Overview
Compiler for CAM-based (Content-Addressable Memory) in-memory accelerators.

- **Paper:** "C4CAM: A Compiler for CAM-based In-memory Accelerators" (ASPLOS 2024)
- **DOI:** [10.1145/3620666.3651386](https://doi.org/10.1145/3620666.3651386)
- **Status:** Research publication

### Key Features
- First compiler targeting CAM-based CIM architectures
- Pattern matching and associative computing optimization
- Novel dataflow transformations for CAM operations

### FeCIM Relevance: LOW
CAM architecture differs from crossbar MVM. However, FeCIM-based ternary CAM (TCAM) is an emerging application where this could be relevant.

---

## 6. XbarSim (2024)

### Overview
Decomposition-based memristive crossbar simulator that addresses limitations in CrossSim and badcrossbar.

- **arXiv:** [2410.19993](https://arxiv.org/abs/2410.19993)
- **Status:** Research publication (2024)

### Key Features
- Decomposition-based solving (faster than full nodal analysis for large arrays)
- Handles non-idealities: sneak paths, IR drop, device variation
- Claims improved accuracy over badcrossbar for large arrays
- Python-based implementation

### FeCIM Relevance: HIGH
Could supplement badcrossbar for Module 2 validation, especially for larger crossbar arrays where decomposition-based solving is more efficient.

### How it compares
| Feature | badcrossbar | CrossSim | XbarSim |
|---------|------------|----------|---------|
| Method | Full nodal analysis | Behavioral + circuit sim | Decomposition-based |
| Speed (large arrays) | Slow (O(n³)) | Fast (GPU) | Fast (decomposition) |
| Accuracy | Exact | High (behavioral) | High (decomposed) |
| GPU support | No | Yes | No |
| Python | Yes | Yes | Yes |

---

## 7. PEtra (ETH Zurich)

### Overview
Processing Element framework for analog/digital CIM architecture exploration.

- **arXiv:** [2410.16016](https://arxiv.org/abs/2410.16016)
- **Status:** Research publication

### Key Features
- Modular processing element abstraction for CIM
- Supports both analog and digital CIM variants
- Energy/area/latency modeling
- Integration with standard NN frameworks

### FeCIM Relevance: MEDIUM
Architecture-level exploration could inform Module 6 EDA design decisions for FeCIM processing elements.

---

## Comparison Matrix

| Tool | Type | License | CIM Type | GPU | Python | Maturity | FeCIM Relevance |
|------|------|---------|----------|-----|--------|----------|----------------|
| CIM-MLC | Compiler | Research | NVM/SRAM | No | Yes | Prototype | HIGH |
| COMPASS TVM | Compiler | Apache-2 | General | Yes | Yes | Active | MEDIUM |
| EasyACIM | Generator | Research | Analog | No | Yes | Paper | HIGH |
| SEGA-dcim | Compiler | Research | Digital | No | Yes | Paper | MEDIUM |
| C4CAM | Compiler | Research | CAM | No | Yes | Paper | LOW |
| XbarSim | Simulator | Research | Memristive | No | Yes | Paper | HIGH |
| PEtra | Framework | Research | Both | No | Yes | Paper | MEDIUM |

---

## Integration with FeCIM Modules

### Module 2 (Crossbar) — XbarSim validation
```
module2 behavioral model
    ↓
Compare vs. badcrossbar (baseline)
    ↓
Compare vs. XbarSim (large array validation)
    ↓
Report accuracy delta
```

### Module 3 (MNIST) — CIM-MLC mapping
```
Trained 5-bit model (Brevitas)
    ↓
CIM-MLC tiling and scheduling
    ↓
Map to crossbar tiles
    ↓
Validate via CrossSim
```

### Module 4 (Circuits) — EasyACIM exploration
```
ADC/DAC requirements
    ↓
EasyACIM design space exploration
    ↓
Optimal peripheral architecture
    ↓
ngspice validation
```

---

## 8. CINM / Cinnamon (ASPLOS 2024)

### Overview
First end-to-end MLIR-based compilation flow for CIM and CNM (Compute-Near-Memory) accelerators, using hierarchical intermediate representations.

- **GitHub:** https://github.com/tud-ccc/Cinnamon
- **Paper:** ASPLOS 2024
- **arXiv:** [2301.07486](https://arxiv.org/abs/2301.07486)
- **License:** Open source
- **Status:** Active (TU Dresden)

### Key Features
- Device-agnostic AND device-aware compilation
- Progressive lowering through domain-specific and device-specific MLIR dialects
- Supports PCM, RRAM-based CIM accelerators and UPMEM CNM
- Built on MLIR for extensible multi-level IR
- FeCIM-specific dialects could be added to the framework

### FeCIM Relevance: HIGH
The MLIR-based architecture is the most extensible path for building a future FeCIM compiler dialect. More general and architecturally sound than CIM-MLC.

---

## 9. CIMFlow (DAC 2025)

### Overview
Integrated CIM architecture framework combining ISA design, MLIR-based compiler, and SystemC-based simulator in one end-to-end workflow.

- **Website:** https://www.cimflow.org/
- **Paper:** DAC 2025
- **arXiv:** [2505.01107](https://arxiv.org/abs/2505.01107)
- **License:** Open source
- **Status:** Active (May 2025)

### Key Features
- ISA design + MLIR-based compiler + SystemC-based simulator
- End-to-end from DNN workloads to digital CIM evaluation
- Advanced partitioning and parallelism strategies
- Up to 2.8x speedup and 61.7% energy reduction
- Addresses CIM capacity constraints

### FeCIM Relevance: HIGH
Best reference architecture for building an integrated FeCIM toolchain. CIM-MLC only provides the compiler; CIMFlow provides compiler + simulator + ISA in one framework.

---

## 10. OpenACM (January 2026)

### Overview
Open-source accuracy-aware compiler for SRAM-based approximate digital CIM macros.

- **GitHub:** https://github.com/ShenShan123/OpenACM
- **arXiv:** [2601.11292](https://arxiv.org/abs/2601.11292)
- **License:** Open source

### Key Features
- Four components: PE Compiler, Multiplier Compiler, SRAM Macro Compiler, Flow-Script Generator
- Supports exact and approximate multipliers of arbitrary bit widths
- Generates backend scripts for OpenROAD (synthesis, PnR, sign-off)
- Up to 64% energy savings with negligible accuracy loss

### FeCIM Relevance: MEDIUM
Demonstrates the full CIM macro design flow from specification to GDS. The methodology is transferable to FeCIM macro generation.

---

## Updated Comparison Matrix

| Tool | Type | License | CIM Type | GPU | Python | Maturity | FeCIM Relevance |
|------|------|---------|----------|-----|--------|----------|----------------|
| CIM-MLC | Compiler | Research | NVM/SRAM | No | Yes | Beta | HIGH |
| COMPASS TVM | Compiler | Apache-2 | General | Yes | Yes | Active | MEDIUM |
| EasyACIM | Generator | Research | Analog | No | Yes | Paper | HIGH |
| SEGA-dcim | Compiler | Research | Digital | No | Yes | Paper | MEDIUM |
| C4CAM | Compiler | Research | CAM | No | Yes | Paper | LOW |
| XbarSim | Simulator | Research | Memristive | No | Yes | Paper | HIGH |
| PEtra | Framework | Research | Both | No | Yes | Paper | MEDIUM |
| **CINM** | Compiler | Open | PCM/RRAM/CNM | No | Yes | Active | **HIGH** |
| **CIMFlow** | Framework | Open | Digital CIM | No | Yes | Active | **HIGH** |
| **OpenACM** | Macro compiler | Open | SRAM CIM | No | Yes | Active | MEDIUM |

---

## Recommended Reading Order

1. Start with **CIMFlow** for the most complete CIM toolchain reference
2. Read **CINM/Cinnamon** if planning to build a FeCIM-specific MLIR dialect
3. Study **CIM-MLC** for crossbar-specific tiling and scheduling
4. Use **XbarSim** for Module 2 validation of large arrays
5. Explore **EasyACIM** or **OpenACM** for Module 4/6 peripheral macro generation

---

**Last Updated:** March 2026
**Category:** CIM Compiler & Mapping Frameworks
**Tools Documented:** 10
