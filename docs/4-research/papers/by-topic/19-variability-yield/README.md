# Variability and Yield

This directory addresses manufacturing variability, device-to-device variations, and yield challenges in ferroelectric memory and compute-in-memory systems. Understanding and mitigating variability is critical for transitioning from laboratory demonstrations to commercial production.

## Papers in this Directory

### fefet_nvmemory_challenges_2023.pdf
2023 comprehensive analysis of FeFET non-volatile memory challenges including retention degradation, read disturb, programming variability, and temperature sensitivity. Proposes mitigation strategies for each failure mechanism.

### fefet_storage_technology_2024.pdf
2024 review of FeFET storage technology focusing on manufacturing yield improvement through process optimization, device design, and circuit-level compensation techniques.

### temperature_variability_fefet_modeling_2024.pdf
2024 modeling study of temperature-dependent variability in FeFETs from cryogenic (4K) to automotive (150°C) operating ranges. Includes statistical models for Monte Carlo simulation.

## Key Findings

1. **Dominant Variation Sources**:
   - Grain size variation in polycrystalline HZO films (±20% Pr variation)
   - Interfacial layer thickness fluctuations (±15% Ec variation)
   - Doping concentration non-uniformity (±10% coercive field shift)

2. **Temperature Dependence**: Coercive field Ec decreases ~30% from 4K to 400K, while remanent polarization Pr decreases ~40%. Variability increases with temperature.

3. **Read Disturb Mitigation**: Non-destructive reads require read voltage <50% of coercive voltage, establishing design constraints for sense amplifiers.

4. **Retention vs Programming Trade-off**: Aggressive programming (high voltage) improves initial state but accelerates long-term retention loss. Optimal programming reduces lifetime by finding balance.

5. **Yield Improvement Pathways**:
   - Error correction codes (ECC) tolerate 5-10% defect rates
   - Adaptive programming algorithms compensate for device variation
   - Redundancy schemes repair defective crossbar rows/columns

## Relevance to FeCIM

Variability modeling is essential for realistic FeCIM system simulation:

- **Crossbar Non-Idealities**: Our module2 must incorporate statistical device variation, not just ideal devices
- **Monte Carlo Simulation**: Temperature and process variation models enable Monte Carlo analysis of crossbar inference accuracy
- **Yield Prediction**: Understanding defect densities informs realistic system-level yield estimates
- **Design Margins**: Variation data establishes required margins for programming voltages and sense circuitry
- **Calibration Requirements**: Device variation quantifies the need for per-device or per-crossbar calibration

**Current Gap**: Our simulations currently use ideal devices. These papers provide data to add realistic variation models to modules 1 and 2.

**Action Items**:
1. Add Gaussian variation to Pr and Ec in Preisach model (σ/μ ≈ 0.15-0.20)
2. Implement temperature-dependent parameter scaling in hysteresis module
3. Add Monte Carlo mode to crossbar simulator with statistical device populations
4. Model ECC overhead in MNIST accuracy calculations

## Related Topics

- **[01-material-physics](../01-ferroelectric-materials/)** - Physical origins of variation (grain boundaries, defects)
- **[18-ald-process-control](../18-ald-process-control/)** - Process optimization to reduce variation at source
- **[02-crossbar-arrays](../04-cim-architectures/)** - Impact of device variation on crossbar MVM accuracy
- **[10-cim-compilers-mapping](../10-cim-compilers-mapping/)** - Compiler techniques for variation-aware mapping
- **[06-neural-networks-inference](../10-cim-compilers-mapping/)** - Accuracy degradation due to weight variation
- **[15-3d-stacking-architectures](../15-3d-stacking-architectures/)** - Yield challenges amplified in 3D structures
