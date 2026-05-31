# Scientific Claims Matrix

This matrix maps each reported scientific claim to executable verification artifacts.

| Claim | Test/Script | Evidence Path | Tolerance/Criterion |
|-------|-------------|---------------|---------------------|
| Pr=19.17 µC/cm² (HZO, Materlik 2015) | `TestPhysicsRegressionCurves`, `scripts/module1_automation.sh --fast` | validation/testdata/ | ±5% |
| Ec=1.16 MV/cm | `TestPhysicsRegressionCurves`, `scripts/module1_automation.sh --fast` | validation/testdata/ | ±5% |
| MNIST 80% accuracy | TestFullStackMNIST | validation/ | ≥80% |
| Energy 44.94 fJ/cell | ~~TestTransientPulse~~ (test removed/renamed — needs investigation) | shared/physics/ | 10-100 fJ range |
| **(M4)** Power conservation < 1% | `scripts/module4_automation.sh --fast` (TestThermodynamics*) | module4-circuits/pkg/arraysim/ | < 1% |
| **(M4)** KCL residual < 1e-12 | `scripts/module4_automation.sh --fast` (TestKirchhoff*) | module4-circuits/pkg/arraysim/ | < 1e-12 |
| **(M4)** MVM accuracy < 5% (Tier-A) | `scripts/module4_automation.sh --full` (TestComputeMVM*) | module4-circuits/pkg/gui/ | < 5% |
| **(M4)** BER < 5% | `scripts/module4_automation.sh --full` (TestComputeBER*) | module4-circuits/pkg/gui/ | < 5% |
| **(M4)** Read margin > 3σ | `scripts/module4_automation.sh --full` (TestReadMarginBER*) | module4-circuits/pkg/gui/ | > 3σ |
| **(M4)** DAC INL < 1 LSB | `scripts/module4_automation.sh --full` (TestPeripheralsINLDNL*) | shared/peripherals/ | < 1 LSB |
| **(M4)** Retention ΔG/G < 1% | `scripts/module4_automation.sh --full` (TestRetention*) | module4-circuits/pkg/gui/ | < 1% |
| **(M4)** Write disturb bounded | `scripts/module4_automation.sh --full` (TestWriteDisturb*) | module4-circuits/pkg/gui/ | bounded |
| Discontinuities all physical (0 spurious) | `TestHeadlessISPPContinuityValidation_PreisachVsLK` | cmd/fecim-lattice-tools/ | 0 spurious |
| ISPP convergence (all target levels reached) | `TestISPPConverges_*`, `scripts/module1_automation.sh --fast` | module1-hysteresis/pkg/controller/ | all levels hit within pulse budget |
| Literature anchoring (RMSE vs experimental datasets) | `TestExperimentalDataValidation` | validation/ | RMSE within thresholds |
| Array ISPP with disturb tracking | TestArrayISPP | shared/physics/ | MaxDisturb < 0.3 |

## Literature Benchmarks (2024-2026, Peer-Reviewed)

These are external benchmarks from published papers — NOT claims made by this simulator.
Use them to validate simulator accuracy relative to the state of the art.

### FeFET Crossbar (Soliman et al., Nature Commun. 14, 6348, 2023)
- 28nm HKMG, 32x32 array, 4 conductance states (2-bit)
- Vth sigma: 38-40 mV
- MNIST accuracy: 96.64% (at sigma=40mV), 97.25% (at sigma=30mV)
- Read latency: ~1 ns, write energy: <1 fJ, write duration: ~1 us
- Energy efficiency: 885.4 TOPS/W
- DOI: 10.1038/s41467-023-42110-y

### 2D Ferroelectric-Gated Hybrid CIM (Science Advances, 2024)
- HZO 20nm, MoS2 channel, 90 conductance states
- 2Pr = 49.5 uC/cm2, 2Vc = ~4.2 V
- ON/OFF ratio: >10^7, endurance: >10^12 cycles
- C2C variation: 0.3%, D2D variation: 0.5%
- Energy efficiency: 26.3 TOPS/W, 3 fJ/bit
- Dynamic tracking accuracy: 99.8%
- DOI: 10.1126/sciadv.adp0174

### HZO Preisach Model Calibration (Sci. Reports, 2021)
- Automated Preisach parameter extraction from experimental P-E loops
- Validated on fluorite-structure ferroelectrics (HfO2, HZO)
- Pr and Ec tend to decrease with increasing coercive field
- DOI: 10.1038/s41598-021-91492-w

### HZO FeFET Reliability (Microsystem Technologies, 2025)
- Endurance: >10^9 cycles with Al2O3 interlayers
- Coercive field reduction: 43.5% via superlattice strategy
- Fatigue suppression via leakage current reduction
- DOI: 10.1007/s00542-025-05919-9

### Analog CIM Precision (npj Unconventional Computing, 2025)
- Systematic analysis of accuracy loss sources in analog CIM
- ADC resolution vs inference accuracy tradeoffs quantified
- DOI: 10.1038/s44335-025-00044-2

## Simulator vs Literature Gaps

| Parameter | Our Simulator | Best Literature | Gap | Action |
|-----------|---------------|-----------------|-----|--------|
| Conductance states | 30 (default) | 90 (Science Adv. 2024) | Within range | FTJ-140 preset covers this |
| C2C variation | 3% base (heuristic) | 0.3% (Science Adv. 2024) | 10x too pessimistic | Update C2CSigmaBase for MoS2 preset |
| D2D variation | 2% (heuristic) | 0.5% (Science Adv. 2024) | 4x too pessimistic | Update ProcessVariation defaults |
| ON/OFF ratio | 100:1 (default) | >10^7 (Science Adv. 2024) | FTJ has much higher | Add FTJ material variant |
| MNIST accuracy (CIM) | ~60% (8-bit) | 96.64% (28nm FeFET) | Large gap | Expected — our model includes all non-idealities |
| Energy per MAC | ~1 pJ (estimated) | <1 fJ (28nm FeFET) | 1000x | Update energy model with measured data |
| Endurance | 10^10 (default) | >10^12 (Science Adv.) | Within range | Literature preset covers this |
| 2Pr | 24.5 uC/cm2 (HZO default) | 49.5 uC/cm2 (MoS2/HZO) | 2D FE is higher | Add 2D-FE material preset |

## Known Limits

- VK-1: Simulator uses heuristic C2C/D2D variation, not device-calibrated
- VK-2: Energy model uses estimates, not measured per-MAC values
- VK-3: No fabricated device data — all validation is simulation-internal
- L07: Preisach calibration uses Park 2015 data only (no 2024-2026 datasets)
- L08: No FTJ (tunnel junction) device model — only FeFET conductance
- L09: No 2D ferroelectric (MoS2/In2Se3) device-specific calibration
- L10: MNIST accuracy gap between sim (~60% CIM) and literature (96%+) unexplained

