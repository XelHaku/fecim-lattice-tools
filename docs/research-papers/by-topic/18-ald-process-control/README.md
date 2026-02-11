# ALD Process Control

This directory focuses on atomic layer deposition (ALD) process optimization for HfO₂-ZrO₂ ferroelectric thin films. Precise control of composition, crystalline phase, and film thickness at the atomic scale is critical for achieving reliable ferroelectric properties and high analog state resolution.

## Papers in this Directory

### fefet_heterogeneous_codoped_hfo2_2025.pdf
2025 study on heterogeneous co-doping strategies in HfO₂-based FeFETs. Investigates combined Zr-Si doping for enhanced ferroelectric stability and multi-level storage capability.

### hzo_coald_wakefree_2024.pdf
2024 work on co-ALD (simultaneous precursor exposure) for HZO films, eliminating wake-up effects and achieving as-deposited ferroelectric behavior. Critical for manufacturing yield.

### hzo_ferroelectric_memristor_2023.pdf
2023 demonstration of HZO-based ferroelectric memristors with controlled ALD process parameters for analog resistance switching. Achieves 30+ distinguishable resistance states.

### hzo_multilayer_ferroelectricity_ald_2022.pdf
2022 investigation of multilayer HZO stacks deposited by ALD, showing how layer interfaces and thickness ratios control polarization and coercive field distributions.

### rapid_cooling_ultrathin_hzo_ald_2025.pdf
2025 breakthrough using rapid thermal cooling after ALD to stabilize orthorhombic phase in ultrathin (<5nm) HZO films. Enables aggressive device scaling while maintaining ferroelectricity.

## Key Findings

1. **Composition Control**: Zr:Hf ratios of 50:50 to 60:40 (atomic %) produce optimal ferroelectric properties, with ±2% tolerance requiring precise ALD cycle control.

2. **Phase Stabilization**: Rapid thermal annealing (RTA) with cooling rates >100°C/s suppresses monoclinic phase formation, critical for ultrathin films <5nm.

3. **Wake-Up Elimination**: Co-ALD techniques achieve ferroelectric behavior without field cycling, improving manufacturing throughput and device-to-device uniformity.

4. **Doping Strategies**: Si co-doping (1-3 atomic %) enhances orthorhombic phase stability and increases remanent polarization Pr by 20-40%.

5. **Multilayer Engineering**: Alternating HfO₂/ZrO₂ superlattices with 1-2nm layer thickness show superior fatigue resistance (10¹²+ cycles) compared to mixed films.

## Relevance to FeCIM

ALD process control is foundational to achieving the 30-level analog resolution central to this project:

- **Material Parameters**: Our physics models (module1) use Pr and Ec values directly dependent on ALD process optimization
- **Device Uniformity**: Wake-up-free processes reduce device-to-device variation, improving crossbar yield simulations
- **Scaling Roadmap**: Rapid cooling techniques validate our assumption of sub-5nm scaling potential
- **30-Level Achievement**: The HZO memristor paper's demonstration of 30+ states provides experimental validation for our quantization strategy
- **Endurance Modeling**: Superlattice fatigue data (10¹² cycles) justifies our endurance assumptions in technology comparisons

**Calibration Data**: These papers provide experimental data for calibrating our Preisach model parameters and coercive field distributions.

## Related Topics

- **[01-material-physics](../01-ferroelectric-materials/)** - Fundamental physics of HZO ferroelectricity enabled by ALD control
- **[19-variability-yield](../19-variability-yield/)** - Process-induced variation that ALD optimization must minimize
- **[15-3d-stacking-architectures](../15-3d-stacking-architectures/)** - Low-temperature ALD enabling BEOL 3D integration
- **[05-endurance-reliability](../19-variability-yield/)** - ALD-optimized films achieving 10¹²+ cycle endurance
- **[02-crossbar-arrays](../04-cim-architectures/)** - Uniform device characteristics critical for crossbar functionality
