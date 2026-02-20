<!-- Category: Features | Module: module4-circuits | Reading time: ~5 min -->
# Module 4 Features: Peripheral Circuit Simulator

> Complete feature inventory for the CIM circuit interface module.

---

## Core Operations

### WRITE -- Program Individual Cells

| Feature | Status |
|---------|--------|
| Select target cell (row, col) | IMPLEMENTED |
| Select target level (0-29) | IMPLEMENTED |
| ISPP write engine (shared with Module 1) | IMPLEMENTED |
| DAC voltage generation | IMPLEMENTED |
| Charge pump voltage boosting | IMPLEMENTED |
| Half-select disturb tracking | IMPLEMENTED |
| Per-cell signed voltage overlay | IMPLEMENTED |
| Write data path visualization (Digital->DAC->FeFET) | IMPLEMENTED |

### READ -- Sense Individual Cells

| Feature | Status |
|---------|--------|
| Safe read voltage application | IMPLEMENTED |
| TIA current-to-voltage conversion | IMPLEMENTED |
| ADC digitization | IMPLEMENTED |
| Read data path visualization (FeFET->TIA->ADC->Digital) | IMPLEMENTED |
| Per-cell signed current display | IMPLEMENTED |
| Composed SNR metric | IMPLEMENTED |

### COMPUTE -- Analog Matrix-Vector Multiplication

| Feature | Status |
|---------|--------|
| Input vector configuration | IMPLEMENTED |
| Analog MVM (V_in x G_matrix) | IMPLEMENTED |
| Column-by-column TIA/ADC readout | IMPLEMENTED |
| Output vector display | IMPLEMENTED |
| Softmax classification | IMPLEMENTED |

---

## Crossbar Array

| Feature | Status |
|---------|--------|
| Configurable NxN size (4-128) | IMPLEMENTED |
| Visual array canvas with zoom | IMPLEMENTED |
| Color-coded conductance heatmap | IMPLEMENTED |
| Cell selection by click | IMPLEMENTED |
| Conductance level labels per cell | IMPLEMENTED |
| Signed voltage/current overlay toggle | IMPLEMENTED |

---

## Architecture Selection

| Feature | Status |
|---------|--------|
| 0T1R passive crossbar | IMPLEMENTED |
| 1T1R single-transistor selector | IMPLEMENTED |
| 2T1R dual-transistor selector | IMPLEMENTED |
| Sneak path physics per architecture | IMPLEMENTED |
| Half-select voltage indicators | IMPLEMENTED |
| MOSFET selector device model (W/L, Vth) | PLANNED |
| Selector I-V curve in read/write path | PLANNED |

---

## Signal Chain Components

### DAC (Digital-to-Analog Converter)

| Feature | Status |
|---------|--------|
| Configurable bit resolution (4-8 bit) | IMPLEMENTED |
| DNL within +/-1 LSB | IMPLEMENTED |
| INL modeled | PARTIAL |
| Voltage output range [VrefN, VrefP] | IMPLEMENTED |

### TIA (Transimpedance Amplifier)

| Feature | Status |
|---------|--------|
| Configurable gain (k-Ohm) | IMPLEMENTED |
| I x Gain + offset conversion | IMPLEMENTED |
| Output clamping | IMPLEMENTED |
| SNR reporting | IMPLEMENTED |

### ADC (Analog-to-Digital Converter)

| Feature | Status |
|---------|--------|
| Configurable bit resolution (5-8 bit) | IMPLEMENTED |
| Monotonicity guaranteed | IMPLEMENTED |
| SAR noise model (thermal, metastability, reference drift) | IMPLEMENTED |
| ENOB/SNR reporting | IMPLEMENTED |

### Charge Pump

| Feature | Status |
|---------|--------|
| Dickson-style voltage boost | IMPLEMENTED |
| Configurable boost factor | IMPLEMENTED |
| UI control for charge pump parameters | PARTIAL |

### Noise Models

| Feature | Status |
|---------|--------|
| Thermal noise (4kTR*BW) | IMPLEMENTED |
| 1/f flicker noise | IMPLEMENTED |
| Shot noise (2qI*BW) | IMPLEMENTED |
| Quantization noise (LSB^2/12) | IMPLEMENTED |
| Total noise composition (variance sum) | IMPLEMENTED |
| SNR calculation | IMPLEMENTED |
| Noise injection into simulation chain | PARTIAL |

---

## Array Coupling and IR Drop

| Feature | Status |
|---------|--------|
| Tier 0: Ideal (no coupling) | IMPLEMENTED |
| Tier A: Approximate IR-drop model | IMPLEMENTED |
| Tier B: Full DC nodal MNA/KCL solver | IMPLEMENTED |
| Wire resistance from geometry | IMPLEMENTED |
| Cell geometry: pitch, wire width/thickness | IMPLEMENTED |
| Boundary conditions (drive R, termination R) | IMPLEMENTED |
| Selector device series resistance | PLANNED |

---

## Views

### Operations View (primary)
- Unified WRITE/READ/COMPUTE interface
- Array visualization with zoom and cell selection
- Data path animation
- Architecture toggle (0T1R/1T1R/2T1R)
- Per-cell signed V/I overlay

### Comparison View
- CPU vs GPU vs FeFET performance table
- TOPS/W benchmarks (static, modeled)

### Reference View
- Timing diagrams for WRITE/READ/COMPUTE cycles
- Peripheral specification tables

---

## Integration Points

| Integration | Status |
|-------------|--------|
| Module 1 -> Module 4: Shared Preisach/conductance | IMPLEMENTED |
| Module 2 -> Module 4: Crossbar array solver | IMPLEMENTED |
| Module 4 -> Module 6: Timing/power back-annotation | PLANNED |

---

## Status Legend

- **IMPLEMENTED**: Code exists, tests pass, feature is functional
- **PARTIAL**: Core logic exists but incomplete (missing UI, test, or placeholder values)
- **PLANNED**: Specified in TODO, not yet coded

---
*FeCIM Lattice Tools -- Simulation baseline, not validated device data.*
