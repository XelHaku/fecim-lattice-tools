# 3D Stacking Architectures

This directory focuses on three-dimensional integration and vertical stacking of ferroelectric memory arrays. 3D architectures are critical for achieving the density and bandwidth required for competitive compute-in-memory systems, enabling massive parallelism in the vertical dimension.

## Papers in this Directory

### ferroelectric_transistors_nand_flash_2025.pdf
2025 work on ferroelectric transistor arrays adopting 3D NAND Flash architecture paradigms. Demonstrates CMOS-compatible vertical gate-all-around (GAA) FeFET structures.

### full_spectrum_3d_ferroelectric_memory_2025.pdf
Comprehensive 2025 study covering full spectrum of 3D ferroelectric memory architectures from simple stacking to monolithic 3D integration with through-silicon vias (TSV) and wafer bonding.

### highly_scaled_3d_fefet_array_2023.pdf
2023 demonstration of highly scaled 3D FeFET arrays achieving sub-20nm feature sizes with multi-layer vertical integration, showing path to terabit-scale density.

### vertical_nand_ferroelectric_paradigm_2024.pdf
2024 analysis of vertical NAND-like ferroelectric memory paradigms, comparing string architectures, access schemes, and peripheral circuit integration strategies.

## Key Findings

1. **NAND Architecture Leverage**: Ferroelectric memories can directly adopt proven 3D NAND manufacturing techniques (vertical channels, gate-all-around), accelerating commercialization.

2. **Density Scaling**: 3D FeFET arrays demonstrate 10-100× density improvement over 2D implementations, with demonstrated layer counts of 8-32 and roadmaps to 128+ layers.

3. **CMOS BEOL Integration**: Critical finding that ferroelectric films can be deposited in back-end-of-line (BEOL) processes enables true monolithic 3D integration with logic layers.

4. **Thermal Budget Compatibility**: Low-temperature ALD processes for HfO₂-ZrO₂ (<400°C) are compatible with BEOL thermal budgets, unlike high-temperature alternatives.

5. **Vertical Bandwidth Advantage**: 3D integration provides massive internal bandwidth (multi-Tb/s) within the memory stack, solving the memory wall problem for CIM applications.

## Relevance to FeCIM

3D stacking is transformational for the FeCIM Lattice Tools project:

- **Density Projections**: Our technology comparisons (module5) must account for 10-100× density gains from 3D integration
- **Thermal Modeling**: 3D stacking introduces significant thermal challenges requiring updated thermal models in our crossbar simulations
- **Bandwidth Calculations**: Internal 3D bandwidth enables larger crossbar arrays than 2D constraints would allow
- **Manufacturing Realism**: CEA-Leti 2024 demonstration of 22nm BEOL FeFETs validates our assumption of CMOS compatibility
- **Future Architecture Module**: These papers provide foundation for a 3D architecture exploration module

**Design Impact**: Our crossbar module (module2) currently assumes 2D arrays. 3D stacking research suggests we should model vertically-integrated crossbars with layer-specific IR drop and thermal effects.

## Related Topics

- **[08-industry-reports](../08-industry-reports/)** - Wafer-scale integration manufacturing context
- **[02-crossbar-arrays](../04-cim-architectures/)** - 2D crossbar foundations extended by 3D stacking
- **[18-ald-process-control](../18-ald-process-control/)** - Low-temperature ALD critical for BEOL compatibility
- **[19-variability-yield](../19-variability-yield/)** - Yield challenges amplified in 3D structures
- **[04-peripheral-circuits](../07-memory-architectures/)** - Peripheral circuit integration in 3D architectures
