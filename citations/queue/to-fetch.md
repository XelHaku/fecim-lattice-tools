# Papers To Fetch

Add candidate papers here before creating full records.

## Priority 1 — Core FeFET Device Physics

- [ ] Mulaosmanovic et al., "Switching Kinetics in Nanoscale Hafnium Oxide Based Ferroelectric Field-Effect Transistors" — ACS AMI 2017, DOI 10.1021/acsami.6b13866
  - Why: FeFET device I-V characteristics, subthreshold swing, needed for compact model
- [ ] Fengler et al., "Comparison of Hafnia and PZT based ferroelectrics for future memory applications" — IEEE TED 2017
  - Why: Endurance and retention comparison HZO vs PZT
- [ ] Lederer et al., "On the Variability of the Memory Window in Ferroelectric HfO2-Based MFM and MFIS Structures" — IEEE IMW 2021
  - Why: C2C and D2D variation data for Module 4 Monte Carlo
- [ ] Dünkel et al., "A FeFET based super-low-power ultra-fast embedded NVM technology" — IEEE IEDM 2017
  - Why: FeFET memory cell characteristics for 1T1R array model
- [ ] Yurchuk et al., "Charge-Trapping Phenomena in HfO2-Based FeFET" — IEEE IRPS 2014
  - Why: Si:HfO2 endurance data, retention vs temperature
- [ ] Toprasertpong et al., "Bias Temperature Instability in Ferroelectric FET" — IEEE IEDM 2019
  - Why: FeFET reliability/BTI for aging engine

## Priority 2 — Additional Material Systems

- [ ] Lomenzo et al., "Ferroelectric HfO2 for Memory Applications" — various 2015-2016
  - Why: Undoped HfO2 antiferroelectric behavior
- [ ] Starschich et al., "Gd:HfO2 ferroelectric properties" — 2017
  - Why: Doping-dependent Ec/Pr for additional material presets
- [ ] Mueller et al., "Si:HfO2 Ferroelectric Memories" — Nano Letters 2012
  - Why: Early Si:HfO2 material parameters
- [ ] Ni et al., "HZO polarization and benchmarking" — APL 2018
  - Why: Pr/Ec benchmarking survey for Module 5 comparison

## Priority 3 — Circuit and System-Level

- [ ] Murmann ADC Survey — ISSCC 2022-2024
  - Why: Calibrate ADC energy/area parameters against measured data
- [ ] CIM-MLC compiler — arXiv 2401.12428 (2024)
  - Why: Crossbar mapping compiler for Module 3 MNIST pipeline
- [ ] arXiv 2312.15444 — State-dependent C2C variation in FeFET
  - Why: Replace constant C2C sigma with state-dependent model
- [ ] arXiv 2024 — Column-shared ADC sharing ratio optimization
  - Why: ADC count vs energy tradeoff for Module 4

## Priority 4 — Crossbar and EDA

- [ ] Slesazeck et al., "2T-1C FeCAP crossbar arrays" — 2020-2023
  - Why: FeCAP (capacitive) crossbar model eliminates IR drop
- [ ] SkyWater SKY130 PDK documentation
  - Why: DRC rule deck, layer stack for Module 6 EDA
