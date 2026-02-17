# Module 4: Peripheral Circuits — Feature Specification

**Last updated:** 2026-02-12  
**Status:** Each feature marked as IMPLEMENTED, PARTIAL, PLANNED, or INVESTIGATION

---

## 1. Overview

Module 4 is the **peripheral circuit simulator** for FeCIM crossbar arrays. It models the complete analog signal chain from digital input to digital output, with configurable array sizes, architecture selection, and noise-aware readout.

**Three views:** Operations (WRITE/READ/COMPUTE), Comparison (CPU/GPU/FeFET benchmarks), Reference (timing diagrams + specs).

---

## 2. Core Operations

### 2.1 WRITE — Program Individual Cells
| Feature | Status | Implementation |
|---------|--------|----------------|
| Select target cell (row, col) | IMPLEMENTED | `tab_unified.go` row/col selectors |
| Select target level (0–29) | IMPLEMENTED | Level slider + label |
| ISPP write engine (shared with Module 1) | IMPLEMENTED | `shared/physics/ispp.go` SSOT |
| DAC voltage generation | IMPLEMENTED | `shared/peripherals/dac.go` |
| Charge pump voltage boosting | IMPLEMENTED | `shared/peripherals/chargepump.go` |
| Half-select disturb tracking | IMPLEMENTED | `tab_unified_voltage.go` residue accumulation |
| Per-cell signed voltage overlay | IMPLEMENTED | `tab_unified_voltage.go` + `device_state.go` |
| Write data path visualization (Digital→DAC→FeFET) | IMPLEMENTED | `tab_unified_actions.go` |

### 2.2 READ — Sense Individual Cells
| Feature | Status | Implementation |
|---------|--------|----------------|
| Safe read voltage application | IMPLEMENTED | `device_state.go` readVoltage |
| TIA current→voltage conversion | IMPLEMENTED | `shared/peripherals/tia.go` (I×Gain + offset) |
| ADC digitization | IMPLEMENTED | `shared/peripherals/adc.go` |
| Read data path visualization (FeFET→TIA→ADC→Digital) | IMPLEMENTED | `tab_unified_actions.go` |
| Per-cell signed current display | IMPLEMENTED | V/I toggle (`8008783`) |
| Composed SNR metric | IMPLEMENTED | Sense panel (`a08323c`) |

### 2.3 COMPUTE — Analog Matrix-Vector Multiplication
| Feature | Status | Implementation |
|---------|--------|----------------|
| Input vector configuration | IMPLEMENTED | `tab_unified.go` input vector UI |
| Analog MVM (V_in × G_matrix) | IMPLEMENTED | `device_state.go` compute path |
| Column-by-column TIA→ADC readout | IMPLEMENTED | Shared peripheral chain |
| Output vector display | IMPLEMENTED | Compute log panel |
| Softmax classification | IMPLEMENTED | `shared/compute/` pipeline |

---

## 3. Crossbar Array

| Feature | Status | Implementation |
|---------|--------|----------------|
| Configurable NxN size (4–128) | IMPLEMENTED | Size selector dropdown, `MaxArraySize=128` (`cba7c17`) |
| Visual array canvas with zoom | IMPLEMENTED | `tab_unified_canvas.go`, `tab_unified_drawing.go` |
| Color-coded conductance levels | IMPLEMENTED | Heatmap in drawing code |
| Cell selection (click) | IMPLEMENTED | Canvas click handler |
| Conductance level display per cell | IMPLEMENTED | Cell label overlay |
| **Signed voltage/current overlay** | **IMPLEMENTED** | V/I toggle button (`8008783`) |

---

## 4. Architecture Selection

| Feature | Status | Implementation |
|---------|--------|----------------|
| 0T1R (passive crossbar) | IMPLEMENTED | `device_state.go`, `arraysim/` |
| 1T1R (single transistor selector) | IMPLEMENTED | Architecture toggle + solver path |
| 2T1R (dual transistor selector) | IMPLEMENTED | Full electrical isolation via masks |
| Sneak path physics (0T1R present, 1T1R suppressed, 2T1R eliminated) | IMPLEMENTED | `module2-crossbar/pkg/crossbar/enhanced.go` |
| Half-select voltage indicators | IMPLEMENTED | `tab_unified_voltage.go` |
| **MOSFET selector device model (W/L, Vth, Ion/Ioff)** | **PLANNED** | M4-CMOS-01: Replace boolean mask with physical transistor |
| **Selector I-V curve in read/write path** | **PLANNED** | M4-CMOS-04: Series conductance model |

---

## 5. Signal Chain — Peripheral Components

### 5.1 DAC (Digital-to-Analog Converter)
| Feature | Status | Implementation |
|---------|--------|----------------|
| Configurable bit resolution (4–8 bit) | IMPLEMENTED | DAC bits selector in UI (`cba7c17`) |
| DNL within ±1 LSB | IMPLEMENTED | `shared/peripherals/dac.go` + `linearity_test.go` |
| INL modeled | PARTIAL | Modeled but not explicitly asserted ±1 LSB in tests |
| Voltage output range [VrefN, VrefP] | IMPLEMENTED | DAC parameters |

### 5.2 TIA (Transimpedance Amplifier)
| Feature | Status | Implementation |
|---------|--------|----------------|
| Configurable gain (kΩ) | IMPLEMENTED | UI presets + slider |
| I×Gain + offset conversion | IMPLEMENTED | `shared/peripherals/tia.go` |
| Output clamping | IMPLEMENTED | Min/max clamp |
| SNR reporting | IMPLEMENTED | TIA SNR method |

### 5.3 ADC (Analog-to-Digital Converter)
| Feature | Status | Implementation |
|---------|--------|----------------|
| Configurable bit resolution (5–8 bit) | IMPLEMENTED | ADC bits selector |
| Monotonicity guaranteed | IMPLEMENTED | Tested in `linearity_test.go` |
| SAR noise model (thermal, metastability, reference drift) | IMPLEMENTED | `adc.go` |
| ENOB/SNR reporting | IMPLEMENTED | ADC analysis methods |

### 5.4 Charge Pump
| Feature | Status | Implementation |
|---------|--------|----------------|
| Dickson-style voltage boost | IMPLEMENTED | `shared/peripherals/chargepump.go` |
| Boost factor configurable | IMPLEMENTED | Parameters + tests |
| **UI control for charge pump parameters** | **PARTIAL** | Model exists, no dedicated UI knob |

### 5.5 Noise Models
| Feature | Status | Implementation |
|---------|--------|----------------|
| Thermal noise (4kTR·BW) | IMPLEMENTED | `shared/peripherals/noise.go` |
| 1/f (flicker) noise | IMPLEMENTED | `noise.go` (`4525cd5`) |
| Shot noise (2qI·BW) | IMPLEMENTED | `noise.go` |
| Quantization noise (LSB²/12) | IMPLEMENTED | `noise.go` |
| Total noise composition (variance sum) | IMPLEMENTED | `TotalNoiseVariance()` |
| SNR calculation | IMPLEMENTED | `SNRDB()` |
| **Noise injection into simulation chain** | **PARTIAL** | Composition is production-ready; chain-level injection is fragmented |

---

## 6. Array Coupling & IR Drop

| Feature | Status | Implementation |
|---------|--------|----------------|
| Tier 0: Ideal (no coupling) | IMPLEMENTED | `CouplingIdeal` in `arraysim/` |
| Tier A: Approximate IR-drop model | IMPLEMENTED | Iterative solver with damping + contact R |
| Tier B: Full DC nodal MNA/KCL | IMPLEMENTED | Matrix-free PCG solver |
| Wire resistance from geometry (ρ·L/A) | IMPLEMENTED | `arraysim/types.go` WireParams |
| Cell geometry: pitch, wire width/thickness, metal resistivity | IMPLEMENTED | `CellGeometry` with SKY130 defaults |
| Boundary conditions (drive R, termination R) | IMPLEMENTED | `BoundaryParams` |
| **Selector device series resistance** | **PLANNED** | M4-CMOS-04: Currently boolean mask, needs conductance model |

---

## 7. Cell Physics

| Feature | Status | Implementation |
|---------|--------|----------------|
| Ferroelectric film: thickness=10nm, area=100nm² | IMPLEMENTED | `shared/physics/cell_geometry.go` |
| E = V/t field conversion | IMPLEMENTED | `CellGeometry.ElectricField()` |
| Q = P·A charge conversion | IMPLEMENTED | `CellGeometry.ChargeFromPolarization()` |
| G = σ·A/t conductance | IMPLEMENTED | `CellGeometry.ConductanceFromConductivity()` |
| P→G transfer function | IMPLEMENTED | `shared/physics/transfer.go` |
| 30 discrete conductance levels | IMPLEMENTED | `HZOMaterial.DiscreteLevel()` |
| **Cell layout footprint (FeFET + selector + routing)** | **PLANNED** | M4-CMOS-02: Area in F² for density calculation |
| **Technology node scaling** | **PLANNED** | M4-CMOS-03: Impact on wire R, transistor params |
| **Gate capacitance loading on wordline** | **PLANNED** | M4-CMOS-05: WL RC = R_wire × (C_wire + N×C_gate) |

---

## 8. Views

### 8.1 Operations View (primary)
- Unified WRITE/READ/COMPUTE interface
- Array visualization with zoom and cell selection
- Data path animation (Digital→DAC→Crossbar→TIA→ADC→Digital)
- Architecture toggle (0T1R/1T1R/2T1R)
- Per-cell signed V/I overlay with toggle

### 8.2 Comparison View
- CPU vs GPU vs FeFET performance table
- TOPS/W static benchmarks (0.5 / 5.0 / 22)
- **Note:** SRAM/ReRAM/MRAM comparison lives in Module 5, not here

### 8.3 Reference View
- Timing diagrams: WRITE/READ/COMPUTE cycle waveforms (ns-level)
- Peripheral component specification tables
- Reference voltage diagrams

---

## 9. Integration Points

| Integration | Status | Notes |
|-------------|--------|-------|
| Module 1 → Module 4: Shared Preisach/conductance | IMPLEMENTED | Via `shared/physics/` (verified `d5ad236`) |
| Module 4 → Module 6: Timing/power back-annotation | PLANNED | M6-TECH-02 |
| Module 2 → Module 4: Crossbar array solver | IMPLEMENTED | `module2-crossbar/` sneak path + scaling |
| Module 4 ↔ Module 6: Shared technology config | PLANNED | M6-TECH-01 |

---

## 10. Investigation & Planned Enhancements

See `TODO.md` sections:
- **M4-CMOS-01..06**: MOSFET selector, cell footprint, technology scaling, gate capacitance
- **M6-TECH-01**: Shared TechnologyNode type
- **M6-POWER-01**: Dynamic power model feeding back into Module 4 displays

### Key investigations needed:

1. **Selector transistor impact on read margin**: How much does finite Ron degrade sense margin at 64×64? At 128×128?
2. **Wordline RC delay vs array size**: At what N does WL delay exceed pulse width? Depends on C_gate loading.
3. **Half-select disturb accumulation rate**: How many write cycles before a half-selected cell drifts one level? Currently tracked but not bounded.
4. **Thermal noise floor vs quantization noise**: At what ADC resolution does thermal noise dominate? Determines useful ADC bits.
5. **Charge pump efficiency**: Dickson model assumes ideal caps — what's realistic efficiency at 3V boost?

---

## Evidence Status Legend

- **IMPLEMENTED**: Code exists, tests pass, feature is functional
- **PARTIAL**: Core logic exists but incomplete (missing UI, missing test, or placeholder values)
- **PLANNED**: In TODO.md with clear specification, not yet coded
- **INVESTIGATION**: Needs research/analysis before implementation can begin
