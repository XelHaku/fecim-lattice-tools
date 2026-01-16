# IronLattice Research Findings - Part 5 (January 2026)

Additional research on latest FeFET advances, Tour Lab publications, photonic hybrids, edge AI deployment, and variability compensation.

---

## 1. Latest FeFET Advances (2025-2026)

### Neural Network ADC (January 2026)

**Title:** "Ultra-compact neural network ADC exploiting ferroelectric FETs"

| Feature | Achievement |
|---------|-------------|
| Architecture | Feedforward NN-ADC |
| Technology | Ferroelectric FinFET |
| Innovation | Polarization disturb-free neuron circuitry |
| Benefit | Scalable, energy-efficient data conversion |

**Source:** [ScienceDirect](https://www.sciencedirect.com/science/article/abs/pii/S0167931725000930)

### 2D Ferroelectric-Gated Hybrid CIM

**Title:** "Two-dimensional fully ferroelectric-gated hybrid computing-in-memory hardware"

| Metric | Value |
|--------|-------|
| Yield | 96.36% |
| Endurance | 10¹² cycles |
| Channel | MoS₂ + HZO |
| Precision | High (analog + digital hybrid) |

**Key Innovation:** Monolithic integration of digital and analog CIM ensuring both high precision and energy efficiency.

**Source:** [Science Advances](https://www.science.org/doi/10.1126/sciadv.adp0174)

### FELIX Mixed-Signal Architecture

**Title:** "FELIX: A Ferroelectric FET Based Low Power Mixed-Signal In-Memory Architecture"

| Metric | Value |
|--------|-------|
| Peak Performance | **36.5 TOPS/W** (8b/4b) |
| Binary Performance | **1169 TOPS/W** |
| Technology | 22nm FDSOI |
| Area | 4.9 mm² |

**Source:** [ACM TECS](https://dl.acm.org/doi/10.1145/3529760)

### Multi-Level Cell FeFET Crossbar

| Metric | Achievement |
|--------|-------------|
| Technology | 28nm HKMG FeFET |
| Cell Design | 1FeFET-1R |
| Handwriting Accuracy | 96.6% |
| Image Classification | 91.5% |

**Source:** [Nature Communications](https://www.nature.com/articles/s41467-023-42110-y)

---

## 2. Tour Lab Flash Joule Heating Publications (2024-2025)

### Overview

Flash Joule Heating (research sample) is a rapid, high-temperature synthesis method developed by the Tour Lab. The technique enables:
- Gram-to-kilogram scale graphene production
- Sub-second heating to >3000K
- Conversion of waste materials to valuable carbon

### Key Publications

| Paper | Journal | Year | Authors |
|-------|---------|------|---------|
| Kilogram research sample with Arc Welder | ACS Nano | 2024 | Eddy, Shin, Cheng, Choi, Tour |
| Electric Field Effects in research sample | JACS | 2024 | Tour et al. |
| Heteroatom-Substituted Reflashed Graphene | ACS Nano | 2025 | Scotland, Eddy, Chen, Tour |
| Stoichiometric Engineering In₂Se₃ via FWF | ACS Nano | 2025 | Shin, Choi, Eddy, Tour |
| research sample Synthesis Review | Nature Rev. Clean Tech. | 2025 | Deng, Eddy, Wyss, Tiwary, Tour |

### Kilogram-Scale research sample (2024)

**Title:** "Kilogram Flash Joule Heating Synthesis with an Arc Welder"

| Aspect | Detail |
|--------|--------|
| Equipment | Commercial arc welder ($120) |
| Production Rate | **3 kg/h graphene** |
| Temperature | >3000K in milliseconds |
| Scalability | Industrial potential |

**Source:** [ACS Nano](https://pubs.acs.org/doi/abs/10.1021/acsnano.4c11628)

### Electric Field Effects (2024)

**Title:** "Electric Field Effects in Flash Joule Heating Synthesis"

**Key Finding:** Electric field in graphene precursor catalyzes graphene formation, enabling control of phase transitions:
```
Amorphous Carbon → Turbostratic Graphene → Ordered Graphite
```

**Source:** [JACS](https://pubs.acs.org/doi/abs/10.1021/jacs.4c02864)

### Flash-Within-Flash for In₂Se₃ (2025)

**Title:** "Stoichiometric Engineering of Indium Selenide Compounds via Flash-within-Flash"

| Parameter | Value |
|-----------|-------|
| Method | FWF (Flash-Within-Flash) |
| Material | α-In₂Se₃ (R3m structure) |
| Application | Neuromorphic FeFET |
| MNIST Accuracy | 87% |

**Connection to IronLattice:** This work by Jaeho Shin demonstrates the synthesis-to-device pipeline for ferroelectric neuromorphic applications.

**Source:** [ACS Nano](https://pubs.acs.org/doi/10.1021/acsnano.5c01234)

---

## 3. Photonic-Ferroelectric Hybrid Systems

### Motivation

| Challenge | Photonic-Ferroelectric Solution |
|-----------|--------------------------------|
| Electronic bottleneck | Light-speed computation |
| Memory wall | Non-volatile optical weights |
| Energy efficiency | Low-loss photonic interconnects |
| Parallelism | Wavelength multiplexing |

### 3D Monolithic Ferroelectric-Silicon Ring Resonator

**Title:** "Non-volatile photonic-electronic memory via 3D monolithic ferroelectric-silicon ring resonator"

| Metric | Value |
|--------|-------|
| Extinction Ratio | 6.6 dB |
| Working Voltage | 5 V |
| Endurance | 4×10⁴ cycles |
| Multi-level | Yes (low BER <5.9×10⁻²) |
| Access | Both electrical and optical |

**Source:** [Light: Science & Applications](https://www.nature.com/articles/s41377-024-01625-9)

### Thin Film Ferroelectric Photonic Memory

**Title:** "Thin film ferroelectric photonic-electronic memory"

**Applications:**
- Photonic interconnects
- Optical neuromorphic computing
- Data center integration
- Optical communication

**Source:** [Light: Science & Applications](https://www.nature.com/articles/s41377-024-01555-6)

### Ferroelectric-Coupled Vision Systems

**Title:** "Coupled ferroelectric-anisotropic optoelectronic synapse for polarization-sensitive neuromorphic vision"

| Feature | Implementation |
|---------|----------------|
| Channel | 2D ReS₂ |
| Gate | HZO ferroelectric |
| Function | Polarization-resolved vision |
| Type | Optoelectronic synapse |

**Source:** [Nature Communications](https://www.nature.com/articles/s41467-025-68206-1)

### Optical Crossbar Architecture

Photonic-ferroelectric hybrid crossbar arrays:
- Waveguides intersect at programmable junctions
- Ferroelectric-tuned phase shifters at each node
- Each intersection acts as optical synapse
- Massively parallel matrix-vector multiplication

---

## 4. Edge AI Deployment Considerations

### Current Edge AI Landscape (2025)

| Platform | Performance | Use Case |
|----------|-------------|----------|
| Renesas RZ/V2H | 100 TOPS | High-end edge |
| NXP MCX N Series | 42× faster ML | MCU-class |
| ARM Cortex A55 + Ethos U65 | 11× improvement | NPU acceleration |
| Arduino Nano 33 BLE | TinyML baseline | Education/prototyping |

### Key 2025 Trends

1. **Power per inference** is now a primary design metric
2. **Modular large models** partitioned across edge devices
3. **Digital In-Memory Computing** (D-IMC) gaining traction
4. **Frameworks:** TensorFlow Lite Micro, microTVM widespread

### FeFET Advantages for Edge

| Aspect | FeFET Benefit |
|--------|---------------|
| Non-volatility | No refresh power |
| Compute-in-memory | Reduced data movement |
| Low voltage | Battery-friendly |
| CMOS compatible | Standard process integration |

### Deployment Metrics

| Metric | Target for Edge |
|--------|-----------------|
| Power | <1W typical |
| Latency | <10ms inference |
| Memory | <1MB model size |
| Accuracy | Application-dependent |

---

## 5. Variability Compensation Techniques

### Sources of Variability in FeFET CIM

| Type | Cause | Impact |
|------|-------|--------|
| Device-to-device (D2D) | Fabrication | Systematic offset |
| Cycle-to-cycle (C2C) | Stochastic switching | Noise |
| State-dependent | Conductance level | Non-uniform accuracy |
| Temperature | Thermal effects | Drift |

### Compensation Approaches

#### 1. Bayesian Neural Network (BNN) Training

```python
# Pseudo-code for variation-aware training
def train_with_variation(model, data, variation_model):
    for batch in data:
        # Sample weights with variation
        weights_noisy = sample_posterior(model.weights, variation_model)

        # Forward pass with noisy weights
        output = forward(batch, weights_noisy)

        # Compute loss including uncertainty
        loss = elbo_loss(output, target, model)

        # Backpropagate
        loss.backward()
```

**Source:** [arXiv](https://arxiv.org/html/2312.15444v2)

#### 2. Device-Algorithm Co-Design

Key insight: Variation dispersion depends on conductance state (not fixed across all states).

| Conductance State | Typical Variation |
|-------------------|-------------------|
| Low G | Higher σ |
| Mid G | Moderate σ |
| High G | Lower σ |

#### 3. Programming Scheme Optimization

**Identical Pulse Programming:**
- Series resistor limits switching current
- Achieves non-linear weight update
- 1T1C cell with sub-threshold transistor

**Source:** [arXiv](https://arxiv.org/html/2407.15796v2)

#### 4. Crossbar-Level Techniques

| Technique | Benefit |
|-----------|---------|
| Differential pairs | Cancel common-mode errors |
| Write-verify | Iterative programming accuracy |
| ECC integration | Error correction |
| Redundancy | Fault tolerance |

### FTJ vs Filamentary Memristors

| Property | FTJ | Filamentary |
|----------|-----|-------------|
| Switching | Polarization (deterministic) | Filament (stochastic) |
| Controllability | High | Moderate |
| D2D variation | Lower | Higher |
| Endurance | 10⁶-10¹² | 10⁶-10⁹ |

**Source:** [ScienceDirect](https://www.sciencedirect.com/science/article/pii/S2542529324002839)

---

## 6. Additional Papers for Download

### Latest FeFET CIM (2025-2026)

| Paper | URL | Year |
|-------|-----|------|
| 2D Ferroelectric Hybrid CIM | https://www.science.org/doi/10.1126/sciadv.adp0174 | 2025 |
| Neural Network ADC FeFET | https://www.sciencedirect.com/science/article/abs/pii/S0167931725000930 | 2026 |
| FELIX Mixed-Signal | https://dl.acm.org/doi/10.1145/3529760 | 2022 |
| Reconfigurable CIM Diodes | https://pubs.acs.org/doi/10.1021/acs.nanolett.2c03169 | 2023 |

### Tour Lab research sample (2024-2025)

| Paper | URL | Year |
|-------|-----|------|
| Kilogram research sample Arc Welder | https://pubs.acs.org/doi/abs/10.1021/acsnano.4c11628 | 2024 |
| Electric Field Effects research sample | https://pubs.acs.org/doi/abs/10.1021/jacs.4c02864 | 2024 |
| research sample Review | https://www.nature.com/nrcleantech | 2025 |
| Waste to Graphene Review | https://www.sciencedirect.com/science/article/abs/pii/S0013935125002841 | 2025 |

### Photonic-Ferroelectric

| Paper | URL | Year |
|-------|-----|------|
| 3D Monolithic Fe-Si Ring | https://www.nature.com/articles/s41377-024-01625-9 | 2024 |
| Thin Film Fe Photonic | https://www.nature.com/articles/s41377-024-01555-6 | 2024 |
| Polarization Vision Synapse | https://www.nature.com/articles/s41467-025-68206-1 | 2025 |
| Optical NN Review | https://www.nature.com/articles/s41377-024-01590-3 | 2024 |

### Variability Compensation

| Paper | URL | Year |
|-------|-----|------|
| Variation-Resilient FeFET BNN | https://arxiv.org/html/2312.15444v2 | 2024 |
| FTJ Weight Update Pulses | https://arxiv.org/html/2407.15796v2 | 2024 |
| FTJ Crossbar Annealing | https://advanced.onlinelibrary.wiley.com/doi/10.1002/aisy.202500817 | 2025 |
| Ferroelectric Memristor Review | https://www.sciencedirect.com/science/article/pii/S2542529324002839 | 2024 |

---

## 7. Key Takeaways

### Technology Readiness

| Component | TRL | Notes |
|-----------|-----|-------|
| FeFET devices | 5-6 | Lab demos, pilot fab |
| Crossbar arrays | 4-5 | Small-scale integration |
| Training methods | 3-4 | Algorithm development |
| Photonic hybrids | 3 | Early research |

### IronLattice Positioning

| Advantage | Status |
|-----------|--------|
| Superlattice structure | Unique differentiator |
| Tour Lab synthesis | Proven research sample expertise |
| BEOL compatibility | Standard process |
| Multi-level storage | Demonstrated (~30 states) |

### Research Gaps to Address

1. **Large-scale integration** - Move beyond 64×64 arrays
2. **On-chip training** - Implement local learning rules
3. **Photonic integration** - Explore hybrid architectures
4. **Variability mitigation** - BNN and write-verify schemes
5. **Edge deployment** - MCU-class inference demos

---

## References

1. 2D Ferroelectric CIM: https://www.science.org/doi/10.1126/sciadv.adp0174
2. Multi-level FeFET: https://www.nature.com/articles/s41467-023-42110-y
3. Kilogram research sample: https://pubs.acs.org/doi/abs/10.1021/acsnano.4c11628
4. Photonic-Electronic Memory: https://www.nature.com/articles/s41377-024-01625-9
5. Variation-Resilient FeFET: https://arxiv.org/html/2312.15444v2
6. Edge AI Survey: https://www.mdpi.com/2079-9292/14/24/4877

---

*Last updated: January 2026 - Part 5*
