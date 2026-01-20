# Ferroelectric CIM Research Summary - January 2026

Consolidated summary of research findings across all documentation.

---

## Executive Overview

### Ferroelectric CIM Technology

**Company:** external research institution spinout (Tour Lab)
**Funding:** $50K One Small Step Grant (2025)
**Key Innovation:** HfO₂/ZrO₂ superlattice ferroelectric compute-in-memory

| Claimed Performance | vs. NAND | vs. DRAM |
|---------------------|----------|----------|
| Energy | 10⁶× lower | 10³× lower |
| Speed | 10⁶× faster | Comparable |
| Operating Voltage | 90% reduction | Lower |

### Key Personnel
- **Dr. external research group** - Principal Investigator, external research institution
- **Dr. Jaeho Shin** - Device Engineer, superlattice inventor
- **Tawfik Jarjour** - Commercialization Lead

---

## Research Documentation

| Document | Topics | Papers |
|----------|--------|--------|
| Part 1 | Core technology, benchmarks, simulation tools | 30+ |
| Part 2 | STDP, SNNs, technology comparisons, Landau coefficients | 25+ |
| Part 3 | CAM, transformers, reliability, manufacturing | 20+ |
| Part 4 | NC-FET, reservoir computing, 3D, security, Vulkan | 25+ |
| Part 5 | Latest FeFET, Tour Lab research sample, photonic hybrids, edge AI, variability | 30+ |
| Part 6 | CIM compilers, thermal management, applications, code analysis | 25+ |
| Part 7 | Research groups, patents, testing methods, commercialization | 20+ |
| Part 8 | **Vulkan GPU compute, Go bindings, real-time hysteresis visualization** | 20+ |
| **Total** | **Comprehensive coverage** | **195+ papers** |

---

## Key Technical Parameters

### Landau-Devonshire Coefficients for HZO

```go
params := &LandauParams{
    Alpha0: 1.72e6,      // V·m·K⁻¹·C⁻¹ (universal)
    Tc:     500,         // K
    Beta:   -2.5e9,      // First-order transition
    Gamma:  1.5e11,      // Stability
    Ps:     25e-2,       // C/m² (25 μC/cm²)
    Ec:     1.0e8,       // V/m (1 MV/cm)
}
```

### Crossbar Non-Idealities

```go
nonIdealities := &CrossbarParams{
    WireResistance:  1.0,    // Ω/segment
    OnOffRatio:      1000,   // Sneak current suppression
    D2DVariation:    0.05,   // 5%
    ReadNoise:       0.02,   // 2%
}
```

### Training Configuration

```go
trainingConfig := &TrainingConfig{
    WeightBits:     4,      // INT4 quantization
    NoiseInjection: 0.05,   // 5% Gaussian
    LearningRate:   0.01,
    TargetAccuracy: 0.87,   // Match claims
}
```

---

## Benchmark Comparisons

### MNIST Accuracy (State of the Art)

| Technology | Accuracy | Year |
|------------|----------|------|
| Ferroelectric memristor (RC) | 98.78% | 2025 |
| RRAM crossbar | 98.1% | 2024 |
| Multi-level FeFET | 96.6% | 2023 |
| All-ferroelectric SNN | 94.9% | 2024 |
| **Ferroelectric CIM (claimed)** | **87%** | 2025 |

### Energy Comparison (per MAC)

| Technology | Energy | Voltage |
|------------|--------|---------|
| **FeFET/FTJ** | **10-100 fJ** | 0.1-4 V |
| ReRAM | 10-100 fJ | 0.5-3 V |
| STT-MRAM | 100-1000 fJ | 0.5-1.5 V |
| PCM | 1000-10000 fJ | 1-3 V |

### Endurance

| Device | Cycles | Notes |
|--------|--------|-------|
| Si-based FeFET | <10⁶ | Interface limited |
| Oxide-channel FeFET | 2×10⁷ | No IL degradation |
| Superlattice FeFET | >10⁹ | Best reported |
| TSMC BEOL target | 10¹² | Engineering goal |

---

## Key Demonstrations

### Transformer Acceleration
- **70,000× energy reduction** vs. GPU (Analog IMC Attention)
- **100× speed improvement** for LLM inference

### CAM Performance
- **10× energy reduction** vs. SRAM CAM
- **60× reduction** for memory-augmented neural networks

### Reservoir Computing
- **98.1% accuracy** on speech recognition
- **1000× speed improvement** vs. software

### Security (PUF)
- **1.89 fJ/bit** readout energy
- **49.8% uniqueness** (ideal: 50%)

---

## Simulation Tools

| Tool | Purpose | URL |
|------|---------|-----|
| FerroX | GPU phase-field | github.com/AMReX-Microelectronics/FerroX |
| IBM AIHWKit | Analog CIM training | github.com/IBM/aihwkit |
| Preisach Verilog-A | Circuit hysteresis | github.com/DavidTobar456/pfecapRevision |

---

## Demo Implementation Plan

### Demo 1: Hysteresis Visualizer
- TDGL/Preisach physics on Vulkan GPU
- Interactive P-E curves
- 30 analog state visualization
- Domain coloring

### Demo 2: Crossbar MVM
- 8×8 or 16×16 array
- Real-time I-V visualization
- Non-ideality toggles
- O(1) complexity demonstration

### Demo 3: MNIST Inference
- 784×128×10 network
- Hardware-aware training
- INT4 quantization
- 87% accuracy target

---

## Industry Landscape

### Key Companies

| Company | Focus | Status |
|---------|-------|--------|
| FMC (Dresden) | DRAM+, 3D CACHE+ | €100M C-round (2025) |
| GlobalFoundries | FeFET eNVM | 28nm/22nm FDSOI |
| TSMC | OS-FeFET | BEOL research |
| Samsung | HZO FeFET | Pilot production |
| **Ferroelectric CIM** | AI CIM | $50K seed |

### Market Projections

| Segment | 2024 | 2034 | CAGR |
|---------|------|------|------|
| FeRAM total | $474M | $852M | 6.1% |
| FeFET embedded | $1.37B | $10.95B | **23.5%** |
| Neuromorphic CIM | - | - | **35%+** |

---

## Download Plan Summary

**Total Sections:** 49 priority areas
**Total URLs:** ~290 papers and resources

| Category | Sections |
|----------|----------|
| Core HfO₂/ZrO₂ | 1-3 |
| CIM/Neuromorphic | 4-10 |
| Simulation/Tools | 11-17 |
| STDP/Training | 18-23 |
| Advanced Topics | 24-28 |
| CAM/Transformer | 29-33 |
| NC/RC/3D/Security | 34-38 |
| Latest/research sample/Photonic/Edge | 39-43 |
| Compilers/Thermal/Apps | 44-46 |
| Groups/Patents/Testing | 47-49 |

---

## Next Steps

1. **Validate parameters** - Cross-reference Landau coefficients with experimental data
2. **Implement Vulkan TDGL** - GPU-accelerated phase-field simulation
3. **Build crossbar demo** - MVM with realistic non-idealities
4. **Train MNIST model** - Hardware-aware training for 87% target
5. **Document API** - Prepare for open-source release

---

## References

### Primary Documentation
1. `docs/RESEARCH_FINDINGS_2026.md` - Core findings
2. `docs/RESEARCH_FINDINGS_2026_PART2.md` - STDP, SNNs, comparisons
3. `docs/RESEARCH_FINDINGS_2026_PART3.md` - CAM, transformers, reliability
4. `docs/RESEARCH_FINDINGS_2026_PART4.md` - NC, reservoir, 3D, security
5. `docs/RESEARCH_FINDINGS_2026_PART5.md` - research sample, photonics, edge AI, variability
6. `docs/RESEARCH_FINDINGS_2026_PART6.md` - Compilers, thermal, applications
7. `docs/RESEARCH_FINDINGS_2026_PART7.md` - Research groups, patents, testing
8. `docs/RESEARCH_FINDINGS_2026_PART8.md` - **Vulkan GPU compute, Go bindings**
9. `docs/VULKAN_DEMO_GUIDE.md` - **Demo1 Vulkan implementation guide**
10. `docs/CURRICULUM_DETAILED.md` - Implementation guide
11. `papers/DOWNLOAD_PLAN.md` - Paper collection (310+ URLs)

### Key External Sources
- Rice Innovation: https://news.rice.edu/news/2025/rice-innovation-awards-fourth-cycle-one-small-step-grants
- FerroX: https://github.com/AMReX-Microelectronics/FerroX
- IBM AIHWKit: https://github.com/IBM/aihwkit
- Universal Curie Constant: https://www.sciencedirect.com/science/article/abs/pii/S2211285520302901

---

*Last updated: January 2026*
