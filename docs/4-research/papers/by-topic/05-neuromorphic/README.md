# Neuromorphic Computing with Ferroelectric Devices

## Overview

This directory contains research on neuromorphic computing implementations using ferroelectric materials, with emphasis on bio-inspired spiking neural networks (SNNs), synaptic plasticity emulation, and event-driven computing paradigms. While FeCIM Lattice Tools focuses on analog compute-in-memory for traditional DNNs, neuromorphic approaches offer alternative pathways for ultra-low-power, brain-inspired computing using the same ferroelectric device technology.

## Papers in this Directory

### Ferroelectric Synaptic Devices
- **`FeFET_Synapse_Neuromorphic_arXiv.pdf`** - Comprehensive study of FeFET devices as artificial synapses with multi-level potentiation/depression
- **`Ferroelectric_Devices_AI_Applications_arXiv.pdf`** - Overview of ferroelectric devices (FeFET, FeCAP, FTJ) for AI applications
- **`ferroelectric_synaptic_transistors_review_2024.pdf`** - 2024 review of ferroelectric synaptic transistors: materials, devices, and applications

### Novel Materials and Approaches
- **`2D_Ferroelectric_Materials_Review_arXiv.pdf`** - Review of 2D ferroelectric materials (MoS₂, WSe₂, α-In₂Se₃) for ultra-scaled neuromorphic devices
- **`2d_spintronics_neuromorphic_2024.pdf`** - 2D spintronic devices for neuromorphic computing (combining ferroelectrics with spintronics)
- **`neuromorphic_spintronics_review_2024.pdf`** - Comprehensive review of spintronic neuromorphic computing

## Key Findings

### Neuromorphic vs. Analog CIM Paradigms

| Aspect | FeCIM (Analog CIM) | Neuromorphic FeFET |
|--------|-------------------|-------------------|
| **Computing Model** | Synchronous MVM | Event-driven spikes |
| **Precision** | 30 analog states (demo baseline; simulation baseline) | 2-16 states typical |
| **Energy/Op** | ~1 pJ/MAC | ~0.01 pJ/spike |
| **Throughput** | High (parallel MVM) | Variable (spike-dependent) |
| **Training** | Backpropagation (off-chip) | STDP (on-chip possible) |
| **Best For** | Inference on trained models | Adaptive, low-power edge AI |

### FeFET as Artificial Synapse

**Synaptic Behaviors Demonstrated:**
1. **Long-term potentiation/depression (LTP/LTD)**: Gradual weight increase/decrease with repeated pulses
2. **Spike-timing-dependent plasticity (STDP)**: Weight change depends on pre/post-spike timing (Δt)
3. **Short-term plasticity**: Temporary weight changes for working memory
4. **Paired-pulse facilitation**: Enhanced response to closely spaced spikes

**FeFET Advantages for Synapses:**
- **Multi-level states**: 8-30 levels enable analog-like weight storage (30 is demo baseline; simulation baseline)
- **Non-volatility**: Weights persist without power (unlike SRAM-based synapses)
- **Low energy**: ~1 fJ/spike for weight update (1000× less than digital)
- **Scalability**: CMOS-compatible, 3D stackable

### Spike-Timing-Dependent Plasticity (STDP)

**STDP Learning Rule:**
```
Δw ∝ exp(-|Δt|/τ) × sign(Δt)
where Δt = t_post - t_pre
      τ = time constant (typically 20 ms)
```

**FeFET Implementation:**
- Pre-spike pulse on gate, post-spike on drain
- Coincident spikes (Δt ≈ 0): Maximum potentiation
- Non-coincident: Weak or no change
- Demonstrated: 96% classification on simplified MNIST (28×28 → 784 synapses)

### 2D Ferroelectric Materials for Neuromorphic

**Why 2D Materials Matter:**
1. **Ultra-thin**: Sub-5nm thickness enables aggressive scaling
2. **Interface-free**: Atomically smooth, minimal defects
3. **Electrostatic control**: Better gate control at scaled dimensions
4. **Heterostructures**: Van der Waals stacking for novel functionalities

**Promising 2D Ferroelectrics:**
- **α-In₂Se₃**: Room-temperature ferroelectricity, high on/off ratio (10⁶)
- **CuInP₂S₆**: Large polarization, low coercive field
- **Twisted bilayer materials**: Emergent ferroelectricity in twisted MoS₂, WSe₂

**Performance:**
- **Retention**: 10⁴ s demonstrated (vs. 10⁹ s for HfO₂, needs improvement)
- **Endurance**: 10⁶ cycles (vs. 10¹² for bulk ferroelectrics)
- **Energy**: 0.1 fJ/update (10× better than HfO₂ FeFET)

### Spintronic-Ferroelectric Hybrid Neuromorphic

**Multiferroic Approach:**
Combine ferroelectric switching with magnetic domain dynamics for:
1. **Non-volatile spintronics**: Ferroelectric controls magnetic state
2. **Stochastic computing**: Thermal fluctuations enable probabilistic neurons
3. **Domain wall motion**: Analog weight storage in magnetic domain positions

**Demonstrated Capabilities:**
- **Stochastic neuron**: Sigmoid activation via thermal fluctuations
- **Low-power switching**: 1 fJ/bit (magnetoelectric coupling)
- **Radiation-hard**: Magnetic states robust to ionizing radiation

### Energy Efficiency Comparison

**Per-Synaptic-Operation Energy:**
- **SRAM (digital)**: 10 pJ (baseline)
- **HfO₂ FeFET**: 1 fJ (10000× better)
- **2D FeFET**: 0.1 fJ (100000× better)
- **Spintronic-FE**: 1 fJ (10000× better, with stochasticity)

**Full Inference (10⁴ synapses, 100 spikes):**
- **Digital SNN (SRAM)**: 10 µJ
- **FeFET SNN**: 1 nJ (10000× better)
- **2D FeFET SNN**: 0.1 nJ (100000× better)

### Learning Mechanisms

**On-Chip Learning (STDP):**
- **Unsupervised**: Feature extraction, clustering (demonstrated)
- **Supervised**: Classification with teacher signal (requires external circuit)
- **Reinforcement**: Reward-modulated STDP (theoretical)

**Challenges:**
- **Asymmetric updates**: Potentiation easier than depression (ionic motion)
- **Variability**: Device-to-device σ/μ = 10-20% (higher than CIM tolerance)
- **Limited linearity**: Weight updates nonlinear (compensated by STDP rules)

## Relevance to FeCIM

### Technology Synergies

**Shared Device Technology:**
1. **HfO₂-ZrO₂ FeFET**: Same base device for CIM and neuromorphic
2. **30 analog states (demo baseline; simulation baseline)**: Sufficient for both MVM weights and synaptic weights
3. **CMOS integration**: Compatible process flows
4. **3D stacking**: Both benefit from vertical integration

**Complementary Strengths:**
- **FeCIM**: High-throughput inference on pre-trained DNNs
- **Neuromorphic**: Ultra-low-power adaptive learning for edge AI

### Hybrid Architecture Opportunities

**Two-Tier System:**
1. **Neuromorphic front-end**: Sparse event-driven sensing (e.g., DVS camera)
2. **FeCIM back-end**: Dense MVM computation on extracted features
3. **Combined benefit**: 1000× energy savings from sparsity + 100× from analog CIM

**Example: Event-Based Vision:**
```
DVS Camera (events) → FeFET SNN (feature extraction)
                   → FeCIM Crossbar (classification)
                   → Decision
```
**Total energy**: <10 nJ/inference (vs. 10 mJ for frame-based GPU)

### Learning Paradigm Comparison

| Paradigm | FeCIM Current | FeFET Neuromorphic | Hybrid Future |
|----------|---------------|-------------------|---------------|
| **Training** | Backprop (off-chip) | STDP (on-chip) | Transfer learning |
| **Accuracy** | 96-98% (MNIST) | 88-92% (MNIST) | 95%+ (ensemble) |
| **Energy** | 10 µJ/inference | 1 nJ/inference | 5 nJ/inference |
| **Latency** | 500 ns | 10 ms (100 Hz spikes) | 1 ms |
| **Adaptability** | Fixed | High | Medium |

### Potential FeCIM Extensions

**Neuromorphic Features for FeCIM Lattice Tools:**
1. **Spike-based input encoding**: Convert pixel intensities to spike rates
2. **STDP-inspired weight updates**: On-chip fine-tuning after deployment
3. **Event-driven inference**: Skip computations for zero activations
4. **Stochastic rounding**: Use FeFET variability for regularization

## Related Topics

- **[01-ferroelectric-materials](../01-ferroelectric-materials/)** - Base materials for both CIM and neuromorphic
- **[04-cim-architectures](../04-cim-architectures/)** - Crossbar architectures applicable to both paradigms
- **[02-training-algorithms](../02-training-algorithms/)** - Comparison with backpropagation-based training
- **[06-photonic-computing](../06-photonic-computing/)** - Alternative ultra-low-power computing approach
- **[13-low-power-design](../07-memory-architectures/)** - Power optimization techniques

## Implementation Considerations for FeCIM

### Short-Term (Educational/Comparison)
Add neuromorphic module to FeCIM Lattice Tools:
- **Module 8**: Spiking neural network simulator
- **Demo**: STDP learning on simple patterns
- **Comparison**: Energy/latency vs. analog CIM approach

### Medium-Term (Hybrid Architecture)
Explore hybrid designs:
- **Sparse activation encoding**: Use SNN-inspired techniques to skip zero-weight computations
- **Adaptive precision**: Neuromorphic-inspired dynamic bit-width adjustment

### Long-Term (Research)
2D ferroelectric devices:
- **Beyond HfO₂**: Explore α-In₂Se₃, CuInP₂S₆ for next-gen devices
- **Multiferroic**: Combine ferroelectric and magnetic for stochastic computing

## References for FeCIM Development

**High Priority (Understanding Context):**
1. `FeFET_Synapse_Neuromorphic_arXiv.pdf` - Comprehensive FeFET synapse study
2. `ferroelectric_synaptic_transistors_review_2024.pdf` - State-of-the-art review

**Medium Priority (Future Directions):**
3. `2D_Ferroelectric_Materials_Review_arXiv.pdf` - Next-generation materials
4. `2d_spintronics_neuromorphic_2024.pdf` - Hybrid approaches

**Advanced Topics:**
5. `neuromorphic_spintronics_review_2024.pdf` - Alternative computing paradigms
