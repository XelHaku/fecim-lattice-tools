# Citable Facts Database

**Last updated:** 2026-05-04
**Total verified facts:** 28

This file is the cross-paper index for facts used by FeCIM Lattice Tools. Facts added only after reading the source and recording the same fact in `citations/papers/{key}.md`.

## Rules

- Preserve exact values and units from the source.
- Include source location such as figure, table, section, or page when available.
- Include experimental or simulation conditions.
- Record evidence level: peer-reviewed, conference, preprint, textbook, or documentation.
- Use `[VERIFY]` only for temporary entries that must not be cited in public claims.

## Index

1. [Ferroelectric Material Properties](#ferroelectric-material-properties)
2. [Crossbar Physics](#crossbar-physics)
3. [Peripheral Circuits](#peripheral-circuits)
4. [Inference Benchmarks](#inference-benchmarks)
5. [Energy, Area, And Timing](#energy-area-and-timing)
6. [EDA And Tooling](#eda-and-tooling)

## Ferroelectric Material Properties

### HZO (Si-doped, 10nm, Park 2015)

- **Pr** = `24` `µC/cm²` [claim: hzo-remanent-polarization-range]
  - Source: Park et al., Adv. Mater. 27, 1811 (2015), Fig 2a
  - Conditions: 10 nm, TiN electrodes, Si-doped, RT, 1 kHz
  - Evidence: peer-reviewed
  - FeCIM DefaultHZO uses 24.5 µC/cm² (conservative midpoint)

- **Ec** = `1.0` `MV/cm`
  - Source: Park et al., Adv. Mater. 27, 1811 (2015), Fig 2a
  - Conditions: same as Pr
  - Evidence: peer-reviewed
  - Literature range: 0.8-1.5 MV/cm

- **Pr/Ec range** = `Pr: 15-34 µC/cm², Ec: 0.8-1.5 MV/cm` [claim: hzo-remanent-polarization-range]
  - Source: Park et al., Adv. Mater. 27, 1811 (2015) — doping/temperature study
  - Evidence: peer-reviewed

### HZO LGD Coefficients (Materlik 2015)

- **β** = `-6.720e8` `J·m⁵/C⁴` [claim: materlik-lgd-coefficients]
  - Source: Materlik et al., J. Appl. Phys. 117, 134109 (2015), Table I
  - Conditions: DFT, 0K, orthorhombic Pca21 HfO2
  - Evidence: peer-reviewed
  - Negative β → first-order phase transition

- **γ** = `1.950e10` `J·m⁹/C⁶` [claim: materlik-lgd-coefficients]
  - Source: same as β
  - Evidence: peer-reviewed

- **Tc (Curie Temperature)** = `598` `K` (computational) / `723` `K` (experimental)
  - Source: Materlik 2015 (598K), FeCIM uses 723K from Park 2015
  - Evidence: peer-reviewed

### NLS Switching (Guo 2018 / Alessandri 2018)

- **Activation Field (Ea)** = `1.9` `MV/cm`
  - Source: Alessandri et al., IEEE EDL 39, 1780 (2018) — NLS model fit
  - Conditions: HZO capacitor, RT
  - Evidence: peer-reviewed

- **Intrinsic Switching Time (τ_inf)** = `100` `ps` (1.0e-10 s)
  - Source: Guo et al., APL 112, 262903 (2018)
  - Conditions: infinite field extrapolation
  - Evidence: peer-reviewed

- **NLSSigma** = `1.5` (log-normal, dimensionless)
  - Source: Guo et al., APL 112, 262903 (2018)
  - Evidence: peer-reviewed

### FeFET Synapse (Jerry 2017)

- **Multi-level states demonstrated**: `32` levels
  - Source: Jerry et al., IEEE IEDM 2017, DOI 10.1109/IEDM.2017.8268338
  - Conditions: FeFET synapse, pulse programming
  - Evidence: conference

### HZO Endurance (Mueller 2013)

- **Endurance cycles** = `10^10` demonstrated
  - Source: Mueller et al., IEEE IEDM 2013, DOI 10.1109/IEDM.2013.6724605
  - Evidence: peer-reviewed

### HZO Superlattice (Cheema 2020)

- **Pr (superlattice)** = `22` `µC/cm²` at 5 nm
  - Source: Cheema et al., Nature 580, 478 (2020), Fig 2c
  - Conditions: HZO/ZrO2 nanolaminate, 5 nm total
  - Evidence: peer-reviewed

### AlScN (2022)

- **Pr (Pt electrode, 200nm)** = `~80` `µC/cm²`
  - Source: micromachines 13, 1629 (2022), Fig 6a, DOI 10.3390/mi13101629
  - Conditions: AlScN, Pt electrode, 200 nm
  - Evidence: peer-reviewed

- **Ec (Pt electrode, 200nm)** = `~3.5` `MV/cm`
  - Source: same as above
  - Evidence: peer-reviewed

- **Pr (Mo electrode, 200nm)** = `~60` `µC/cm²`
  - Source: micromachines 13, 1629 (2022), Fig 6b
  - Evidence: peer-reviewed

### PZT Thin Film (Bi 2024)

- **Pr (trace A)** = `~35` `µC/cm²`
  - Source: Bi et al., Nanomaterials 14, 50432 (2024), Fig 2, DOI 10.3390/nano14050432
  - Evidence: peer-reviewed

### BTO (Jaiswal 2021)

- **Pr (trilayer)** = `~15` `µC/cm²`
  - Source: Jaiswal et al., Crystals 11, 1192 (2021), DOI 10.3390/cryst11101192
  - Evidence: peer-reviewed

### HZO Wake-Up (Kim 2020)

- **Wake-up cycles** = `~10^3` for TiN/HZO
  - Source: Kim et al., Materials 13, 2968 (2020), Fig 3a, DOI 10.3390/ma13132968
  - Evidence: peer-reviewed

### HZO Wake-Up (Pesic 2016)

- **Wake-up mechanism**: oxygen vacancy redistribution
  - Source: Pesic et al., Adv. Funct. Mater. 2016, DOI 10.1002/adfm.201603182
  - Evidence: peer-reviewed

## Crossbar Physics

### CrossSim Reference (Sandia 2024)

- **SOR solver** for IR drop + sneak path modeling
  - Source: CrossSim SAND2024-05171C, DOI 10.2172/2563881
  - Evidence: technical report
  - FeCIM ported SOR solver to `shared/crossbar/solver.go`

### BadCrossbar (Joksas 2020)

- **Exact nodal analysis** via KCL sparse solver (SuperLU)
  - Source: Joksas et al., SoftwareX 2020, DOI 10.1016/j.softx.2020.100617
  - Evidence: peer-reviewed

### FeFET Multi-Level (Soliman 2023)

- **Multi-level FeFET crossbar**: 4-state (2-bit) FeFET cells demonstrated
  - Source: Soliman et al., Nature Comms 14, 6348 (2023), DOI 10.1038/s41467-023-42110-y
  - Evidence: peer-reviewed

### 200-State HZO FTJ (2024)

- **140 conductance states** on HZO FTJ
  - Source: Oh et al., IEEE EDL 38(6), 732 (2017) — 32 states
  - Source: Song et al., Adv. Science 2024 — 140 states
  - [VERIFY] 90-state claim from Science Advances 2024, DOI 10.1126/sciadv.adp0174

### HZO FTJ Accuracy (2025)

- **98.24% MNIST accuracy** — NOT a FeCIM device claim (HZO FTJ reservoir computing)
  - Source: J. Alloys Compounds 2025, DOI 10.1016/j.jallcom.2025.181869
  - Evidence: peer-reviewed
  - Note: Different architecture (FTJ), not FeCIM crossbar

## Peripheral Circuits

[VERIFY] ADC energy benchmarks from ISSCC 2022-2024 survey — used in `shared/peripherals/adc.go`
[VERIFY] SAR ADC INL/DNL defaults (0.5/0.25 LSB) — literature survey values, not measured

## Inference Benchmarks

### FeCIM MNIST Baseline

- **80% MNIST accuracy** — educational target, not validated device measurement
  - FeCIM internal baseline at 30 quantization levels
  - Evidence: educational default (not externally validated)

### External Benchmarks (not FeCIM)

- **98.24%** HZO FTJ reservoir computing (J. Alloys Comp. 2025)
- **97%** current-limited FeFET (IEEE TED 2022, DOI 10.1109/TED.2022.3216973)
- **96.64%** 4-state FeFET σ=38mV (Nature Comms 14, 6348, 2023)

## Energy, Area, And Timing

[VERIFY] All energy/area/timing claims in Module 5 (Comparison) are educational estimates
[VERIFY] TIA bandwidth and gain parameters from `shared/peripherals/tia.go` are heuristic baselines
[VERIFY] Charge pump efficiency (0.70-0.75) from `shared/peripherals/chargepump.go` is a calibration constant

## EDA And Tooling

### SKY130 Process

- **Minimum metal width** = `0.17` `µm`
- **Minimum metal spacing** = `0.17` `µm`
- **Cell pitch** = `0.46` `µm` × `2.72` `µm` (standard cell row)
- Source: SkyWater SKY130 PDK documentation
- Evidence: documentation

### OpenLane Flow

- **Supported PDKs**: SKY130, GF180, IHP130
  - Source: OpenLane documentation
  - FeCIM only implements SKY130 code generation
