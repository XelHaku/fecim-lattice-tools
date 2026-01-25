# FECIM TO WAFER: Complete Design-to-Fabrication Roadmap

> **Document Purpose**: Comprehensive guide for taking FeCIM array designs from Module 6 EDA tools to physical silicon fabrication.
>
> **Last Updated**: 2026-01-24
> **Document Version**: 1.0
> **Author**: XelHaku / FeCIM Visualizer Project

---

## Table of Contents

1. [Executive Summary](#1-executive-summary)
2. [Technology Foundation](#2-technology-foundation)
3. [Module 6 EDA Design Suite](#3-module-6-eda-design-suite)
4. [OpenLane Integration Strategy](#4-openlane-integration-strategy)
5. [Process Design Kits (PDKs)](#5-process-design-kits-pdks)
6. [Fabrication Pathways](#6-fabrication-pathways)
7. [Simulation & Validation](#7-simulation--validation)
8. [Custom FeFET Cell Design](#8-custom-fefet-cell-design)
9. [Complete Design-to-Wafer Workflow](#9-complete-design-to-wafer-workflow)
10. [Cost & Timeline Analysis](#10-cost--timeline-analysis)
11. [Risk Assessment & Mitigation](#11-risk-assessment--mitigation)
12. [References](#12-references)

---

## 1. Executive Summary

### Vision

Transform FeCIM (Ferroelectric Compute-in-Memory) array designs from our Module 6 EDA tools into production-ready silicon through open-source EDA flows and accessible fabrication pathways.

### Key Findings

| Aspect | Status | Details |
|--------|--------|---------|
| **EDA Tools** | ✅ Ready | Module 6 generates OpenLane-compatible Verilog/DEF |
| **OpenLane Integration** | ✅ Validated | PLACEMENT_CURRENT_DEF injection works |
| **SKY130 PDK** | ✅ Available | Open-source, well-documented |
| **FeFET Support** | ⚠️ Custom Required | No standard PDK includes FeFET devices |
| **Fabrication** | ✅ Multiple Options | $0-$15k for standard CMOS; custom FeFET requires R&D |

### Technology Readiness

| Component | TRL | Notes |
|-----------|-----|-------|
| HfO₂-ZrO₂ FeFET Devices | 4 | Lab validation (Tour Lab, academic) |
| CIM Array Architecture | 5-6 | Component validation, subsystem demos |
| OpenLane/SKY130 Flow | 9 | Production-ready |
| Module 6 EDA Suite | 4-5 | Functional prototype |

### Recommended Path

**Phase 1 (Immediate)**: Standard CMOS peripheral circuits via ChipFoundry/Tiny Tapeout
**Phase 2 (6-12 months)**: Custom FeFET BEOL post-processing via university cleanroom
**Phase 3 (12-24 months)**: Integrated FeFET+CMOS via foundry R&D partnership

---

## 2. Technology Foundation

### 2.1 Ferroelectric Memory Principles

FeCIM leverages **HfO₂-ZrO₂ (HZO) superlattice** ferroelectric materials to achieve:

| Parameter | Value | Source |
|-----------|-------|--------|
| **Discrete States** | 30 levels (~4.9 bits/cell) | Dr. Tour COSM 2025; Jerry 2017: 32 states |
| **Remanent Polarization (Pr)** | 15-34 μC/cm² | Nature Commun. 2025, ACS 2020 |
| **Coercive Field (Ec)** | 1.0-1.5 MV/cm | Nature Commun. 2025 |
| **Endurance Target** | 10¹² cycles | Achieved in HZO superlattice (PMC 2024) |
| **Data Retention** | >10 years | 4nm HZO capacitors demonstrated |
| **CMOS Compatibility** | Yes | BEOL integration <400°C |

### 2.2 Three Operating Modes

Module 6 supports three FeCIM operation modes:

```
┌─────────────────────────────────────────────────────────────────────┐
│                    FeCIM Operation Modes                             │
├────────────────────┬────────────────────┬────────────────────────────┤
│   Storage Mode     │   Memory Mode      │   Compute Mode             │
│   ─────────────    │   ───────────      │   ────────────             │
│   NAND replacement │   DRAM replacement │   AI accelerator           │
│   No weights       │   No weights       │   Weights optional         │
│   10+ year retain  │   ~10ns access     │   Analog MVM               │
│   30 levels/cell   │   Fast switching   │   Matrix-vector multiply   │
└────────────────────┴────────────────────┴────────────────────────────┘
```

### 2.3 Dr. Tour / IronLattice Research Status

**IronLattice** is Dr. external research group's ferroelectric CIM technology developed at external research institution:

| Claim | Status | Source |
|-------|--------|--------|
| 30 analog states | ✅ Confirmed | COSM 2025 presentation |
| 87% MNIST accuracy | ✅ Hardware-tested | COSM 2025 presentation |
| 10¹² cycle endurance | ⚠️ Target, not achieved | Dr. Tour: "still have to get this up to 10¹²" |
| TRL 4 status | ✅ Confirmed | Dr. Tour explicitly stated at COSM 2025 |
| 10M× vs NAND energy | ❌ Unverified | Dr. Tour claim only; peer-reviewed max: 25-100× |

**Verified Funding**: Rice Innovation "One Small Step Grant" ($50,000) January 2025

---

## 3. Module 6 EDA Design Suite

### 3.1 Overview

Module 6 is a universal chip design tool that generates physical layouts for FeCIM arrays ready for OpenLane/OpenROAD fabrication flow.

### 3.2 Generated Outputs

| Format | Extension | Use Case |
|--------|-----------|----------|
| **Verilog** | `.v` | Structural netlist for OpenLane synthesis |
| **DEF** | `.def` | Cell placement for OpenLane injection |
| **SPICE** | `.sp` | ngspice/HSPICE simulation |
| **JSON** | `.json` | Full design data, version control |
| **CSV** | `.csv` | Spreadsheet analysis |

### 3.3 7-Tab GUI Interface

| Tab | Name | Status | Purpose |
|-----|------|--------|---------|
| 1 | **Configure** | ✅ | Array parameters (mode, size, peripherals) |
| 2 | **Layout** | ✅ | Visual crossbar grid visualization |
| 3 | **HDL** | ✅ | Verilog netlist + DEF generation |
| 4 | **Explorer** | Placeholder | Design space analysis |
| 5 | **Simulate** | Placeholder | ngspice simulation bridge |
| 6 | **Export** | ✅ | Multi-format output |
| 7 | **Learn** | ✅ | OpenLane documentation |

### 3.4 CLI Usage

```bash
# Storage mode - High-density non-volatile storage
go run ./cmd/eda-cli -mode storage -rows 256 -cols 256 -name storage_array

# Memory mode - High-speed DRAM replacement
go run ./cmd/eda-cli -mode memory -rows 128 -cols 128 -name memory_array

# Compute mode - AI accelerator with weights
go run ./cmd/eda-cli -mode compute -input weights.json -rows 64 -cols 64 -output ./output

# Full options
go run ./cmd/eda-cli \
  -mode compute \
  -input data/sample_weights_8x8.json \
  -output ./output \
  -name my_design \
  -rows 8 -cols 8 \
  -levels 30 \
  -tech SKY130 \
  -arch 1T1R \
  -vdd 1.8 \
  -json=true -csv=true -spice=true -verilog=true -def=true
```

---

## 4. OpenLane Integration Strategy

### 4.1 OpenLane Flow Overview

```
┌─────────────────────────────────────────────────────────────────────┐
│                      OpenLane ASIC Flow                              │
├─────────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐                                                    │
│  │ RTL Design  │ ◄── Verilog from Module 6                          │
│  └──────┬──────┘                                                    │
│         ▼                                                           │
│  ┌─────────────┐     SYNTH_ELABORATE_ONLY=1                         │
│  │  Synthesis  │ ◄── Skip logic synthesis for structural netlists   │
│  │   (Yosys)   │                                                    │
│  └──────┬──────┘                                                    │
│         ▼                                                           │
│  ┌─────────────┐     FP_DEF_TEMPLATE                                │
│  │ Floorplan   │ ◄── Use FeCIM DEF for die area/pins                │
│  └──────┬──────┘                                                    │
│         ▼                                                           │
│  ┌─────────────┐     PLACEMENT_CURRENT_DEF ◄───────────────┐        │
│  │  Placement  │ ◄── PL_SKIP_INITIAL_PLACEMENT=1           │        │
│  └──────┬──────┘                              ┌────────────┴──────┐ │
│         ▼                                     │ Module 6 DEF      │ │
│  ┌─────────────┐                              │ Generator         │ │
│  │     CTS     │     RUN_CTS=0 (optional)     │ - FIXED cells     │ │
│  └──────┬──────┘                              │ - 1T1R layout     │ │
│         ▼                                     └───────────────────┘ │
│  ┌─────────────┐                                                    │
│  │   Routing   │                                                    │
│  │ (TritonRoute)│                                                   │
│  └──────┬──────┘                                                    │
│         ▼                                                           │
│  ┌─────────────┐                                                    │
│  │   Signoff   │ ◄── DRC (Magic), LVS (netgen), STA (OpenSTA)       │
│  └──────┬──────┘                                                    │
│         ▼                                                           │
│  ┌─────────────┐                                                    │
│  │    GDSII    │ ◄── Final layout for fabrication                   │
│  └─────────────┘                                                    │
└─────────────────────────────────────────────────────────────────────┘
```

### 4.2 Key Configuration Variables

```json
{
    "PDK": "sky130A",
    "STD_CELL_LIBRARY": "sky130_fd_sc_hd",
    "DESIGN_NAME": "fecim_crossbar",

    "EXTRA_LEFS": "/path/to/fecim_cell.lef",
    "EXTRA_GDS_FILES": "/path/to/fecim_cell.gds",
    "EXTRA_LIBS": "/path/to/fecim_cell.lib",

    "FP_DEF_TEMPLATE": "/path/to/fecim_crossbar.def",
    "PLACEMENT_CURRENT_DEF": "/path/to/fecim_crossbar.def",
    "PL_SKIP_INITIAL_PLACEMENT": 1,

    "SYNTH_ELABORATE_ONLY": 1,
    "VERILOG_FILES_BLACKBOX": "/path/to/fecim_cell.v",

    "RUN_CTS": 0,
    "CLOCK_PERIOD": "10.0"
}
```

### 4.3 OpenLane 2.0 Python-Based Flow

OpenLane 2.0 uses a Python-based configuration with `MACROS` dictionary:

```python
# config.py for OpenLane 2.0
from openlane.flows import Classic
from openlane.steps import Yosys, OpenROAD

class FeCIMFlow(Classic):
    config = {
        "DESIGN_NAME": "fecim_array",
        "PDK": "sky130A",
        "MACROS": {
            "fecim_crossbar": {
                "instances": {
                    "crossbar_inst": {
                        "location": [10, 10],
                        "orientation": "N"
                    }
                }
            }
        }
    }
```

### 4.4 DEF Format Requirements

```def
VERSION 5.8 ;
NAMESCASESENSITIVE ON ;
DIVIDERCHAR "/" ;
BUSBITCHARS "[]" ;
DESIGN fecim_crossbar_64x64 ;
UNITS DISTANCE MICRONS 1000 ;

DIEAREA ( 0 0 ) ( 30000 30000 ) ;

COMPONENTS 4096 ;
  - cell_0_0 fecim_bit + FIXED ( 0 0 ) N ;
  - cell_0_1 fecim_bit + FIXED ( 460 0 ) N ;
  - cell_0_2 fecim_bit + FIXED ( 920 0 ) N ;
  ...
END COMPONENTS

PINS 128 ;
  - WL[0] + NET WL[0] + DIRECTION INPUT + USE SIGNAL
    + PORT
      + LAYER met1 ( 0 0 ) ( 140 140 ) ;
  ...
END PINS

END DESIGN
```

**Key Points**:
- **FIXED**: Cells that placement tools must not move
- **PLACED**: Cells that may be adjusted during optimization
- Cell pitch: 460nm (SKY130 site width × 1)
- Cell height: 2720nm (SKY130 standard cell height)

---

## 5. Process Design Kits (PDKs)

### 5.1 SKY130 PDK (Primary Target)

| Parameter | Value | Notes |
|-----------|-------|-------|
| **Technology Node** | 130nm | SkyWater Technology Foundry |
| **License** | Apache 2.0 | Fully open-source |
| **Cell Height** | 2.72 μm | 8 horizontal routing tracks |
| **Site Width** | 0.46 μm | Unit cell placement grid |
| **Site Name** | `unithd` | For LEF SITE definition |
| **Power Rails** | VPWR, VGND | On met1, top/bottom |
| **VDD** | 1.8V typical | 1.65V-1.95V range |

#### Metal Layer Stack

| Layer | Name | GDS# | Min Width | Min Space |
|-------|------|------|-----------|-----------|
| Local Interconnect | li1 | 67/20 | 0.17 μm | 0.17 μm |
| Metal 1 | met1 | 68/20 | 0.14 μm | 0.14 μm |
| Metal 2 | met2 | 69/20 | 0.14 μm | 0.14 μm |
| Metal 3 | met3 | 70/20 | 0.30 μm | 0.30 μm |
| Metal 4 | met4 | 71/20 | 0.30 μm | 0.30 μm |
| Metal 5 | met5 | 72/20 | 1.60 μm | 1.60 μm |

#### Pin Placement for FeCIM Cells

| Pin Type | Layer | Placement | Notes |
|----------|-------|-----------|-------|
| Word Line (WL) | met1 | Left/right edge | Horizontal routing |
| Bit Line (BL) | met2 | Top/bottom edge | Vertical routing |
| Select Line (SL) | met1 | Left/right edge | 1T1R architecture |
| Power (VPWR) | met1 | Top edge | 0.48 μm width |
| Ground (VGND) | met1 | Bottom edge | 0.48 μm width |

### 5.2 IHP SG13G2 PDK (Alternative)

| Parameter | Value | Notes |
|-----------|-------|-------|
| **Technology** | 130nm BiCMOS | SiGe:C npn-HBT |
| **Performance** | fT/fmax = 350/450 GHz | High-speed capability |
| **Open-Source PDK** | Yes | IHP-Open-PDK on GitHub |
| **MEMRES Module** | Available | RRAM option (custom) |
| **Pricing** | €7,300/mm² | Academic discount via EUROPRACTICE |

### 5.3 GF180MCU PDK

| Parameter | Value | Notes |
|-----------|-------|-------|
| **Technology** | 180nm CMOS | GlobalFoundries |
| **License** | Apache 2.0 | Fully open-source |
| **Google MPW** | Free shuttles | Competitive selection |
| **EUROPRACTICE** | €913-€1,000/mm² | Academic pricing |

### 5.4 PDK Tool Support Matrix

| PDK | Magic | KLayout | OpenLane | Xschem | ngspice |
|-----|-------|---------|----------|--------|---------|
| SKY130 | ✅ | ✅ | ✅ | ✅ | ✅ |
| GF180MCU | ✅ | ✅ | ✅ | ✅ | ✅ |
| IHP SG13G2 | ⚠️ | ✅ | ⚠️ | ✅ | ✅ |

---

## 6. Fabrication Pathways

### 6.1 Option Comparison Matrix

| Program | Cost | Process | Timeline | FeFET Support | Academic |
|---------|------|---------|----------|---------------|----------|
| **ChipFoundry ChipCreate** | $14,950 | SKY130 | 5 months | ❌ No | ❌ No |
| **Tiny Tapeout** | €70-€2,240 | SKY130/IHP/GF180 | ~6 months | ❌ No | ✅ Low-cost |
| **IHP SG13G2 Direct** | €7,300/mm² | 130nm BiCMOS | 4-5 months | ⚠️ Custom | ✅ EUROPRACTICE |
| **GF180MCU OpenMPW** | **FREE** | 180nm CMOS | 6-9 months | ❌ No | ✅ Free |
| **EUROPRACTICE** | €913-€1,000/mm² | Various | 3-6 months | ❌ No | ✅ Mini@sic |
| **MOSIS 2.0** | Contact | Various | Variable | ⚠️ R&D focus | ✅ University |

### 6.2 ChipFoundry / ChipCreate (Efabless Successor)

**Status**: Efabless collapsed early 2025; ChipFoundry/ChipCreate is the successor.

| Parameter | Value |
|-----------|-------|
| **Cost** | $14,950 per project |
| **Deliverables** | 100 QFN packaged parts or bare die |
| **Additional die** | $3,000 for 50 extra units |
| **Timeline** | 5 months from submission |
| **PDK** | SKY130 (130nm) |
| **ReRAM Support** | Available November 2025+ |

**Links**:
- [ChipFoundry Main Site](https://chipfoundry.io/)
- [ChipCreate FAQ](https://chipfoundry.io/faqs)

### 6.3 Tiny Tapeout

| Parameter | Value |
|-----------|-------|
| **Single tile** | €70 minimum |
| **Project range** | €70 to €2,240 (1-32 tiles) |
| **Analog pins** | €40/pin (first 2), €100/pin (additional) |
| **Processes** | SKY130, IHP SG13G2, GF180MCU |

**Important**: IHP25b chips remain **property of IHP** (loan only, not ownership).

**Links**:
- [Tiny Tapeout](https://tinytapeout.com/)
- [Tiny Tapeout FAQ](https://tinytapeout.com/faq/)

### 6.4 IHP Direct Research Shuttles

**2026 Shuttle Dates** (SG13G2):
- June 10, June 29, July 26, December, March

| Parameter | Value |
|-----------|-------|
| **Standard pricing** | €7,300/mm² |
| **Minimum area** | 0.8 mm² |
| **Deliverables** | 40 diced samples with E-test |
| **LBE module** | €5,000 per order |
| **MEMRES module** | Available (details TBD) |

**Links**:
- [IHP MPW Service](https://www.ihp-microelectronics.com/services/research-and-prototyping-service/mpw-prototyping-service)
- [IHP-Open-PDK](https://github.com/IHP-GmbH/IHP-Open-PDK)

### 6.5 Google OpenMPW (GF180MCU)

| Parameter | Value |
|-----------|-------|
| **Cost** | **FREE** (competitive selection) |
| **Capacity** | 40 projects per shuttle |
| **PDK License** | Apache 2.0 |

**Links**:
- [Google GF180MCU Announcement](https://opensource.googleblog.com/2022/10/announcing-globalfoundries-open-mpw-shuttle-program.html)
- [GF180MCU PDK](https://github.com/google/gf180mcu-pdk)

### 6.6 EUROPRACTICE (Academic Access)

**Available Foundries**:
- ams OSRAM: 0.18μ CMOS €1,650/mm² (€1,500 discounted)
- GlobalFoundries: 130nm-12nm, €913-€31,240/mm²
- IHP: €3,825-€9,000/mm²
- TSMC: 0.13μm-7nm
- X-FAB, UMC, STMicroelectronics

**Eligibility**: EU/associated countries with academic membership

**Links**:
- [EUROPRACTICE](https://europractice-ic.com/)
- [2026 Schedules](https://europractice-ic.com/schedules-prices-2026/)

---

## 7. Simulation & Validation

### 7.1 Pre-Silicon Simulation Tools

| Tool | Purpose | Level | Speed |
|------|---------|-------|-------|
| **CrossSim 3.1.1** | CIM array simulation | Architecture | Minutes |
| **NeuroSim V2.1** | Energy/area estimation | System | Hours |
| **AIHWKIT 1.0.0** | IBM analog AI training | Algorithm | Hours |
| **CiMLoop** | Architecture exploration | DSE | Minutes |
| **ngspice** | Circuit simulation | Transistor | Hours-Days |

### 7.2 CrossSim (Sandia National Labs)

```bash
# Installation
pip install crosssim

# Basic usage for FeCIM arrays
from crosssim import CoreParams, NeuralCore

core_params = CoreParams(
    rows=64, cols=64,
    cell_type="FeFET",
    Gmin_uS=1.0, Gmax_uS=100.0,
    num_levels=30
)
```

**Features**:
- Device-to-device variability modeling
- IR drop simulation
- Sneak path analysis for passive arrays
- 30-level quantization support

**Links**: [CrossSim GitHub](https://github.com/sandialabs/cross-sim)

### 7.3 NeuroSim V2.1 (Georgia Tech)

| Feature | Description |
|---------|-------------|
| **Purpose** | Benchmarking neural network accelerators |
| **Includes** | NeuroSim core + DNN inference framework |
| **Output** | Energy, latency, area estimates |
| **Calibration** | Validated against silicon (RRAM) |

**Links**: [NeuroSim](https://github.com/neurosim)

### 7.4 ngspice Simulation for FeCIM Cells

```spice
* FeCIM 1T1R Cell - ngspice simulation
.include /path/to/sky130_fd_pr/models/sky130.lib.spice tt

* FeFET compact model (Verilog-A via OpenVAF)
.include fefet_model.va

* 1T1R cell
M1 BL SL FE_TOP VSS sky130_fd_pr__nfet_01v8 W=0.42u L=0.15u
XFE FE_TOP WL fefet_1t1r Vth_0=0.5 Pr=15u Ec=1.2M

* Transient analysis
.tran 1n 100n
.control
run
plot V(BL) V(WL)
.endc
.end
```

### 7.5 Validation Checklist

- [ ] Functional simulation (Verilog testbench)
- [ ] CrossSim array-level accuracy verification
- [ ] ngspice transistor-level timing
- [ ] NeuroSim energy/area estimation
- [ ] DRC clean (Magic)
- [ ] LVS clean (netgen)
- [ ] STA timing closure (OpenSTA)

---

## 8. Custom FeFET Cell Design

### 8.1 Material Stack: HfO₂-ZrO₂ Superlattice

**BEOL-Compatible Process** (<400°C):

```
┌─────────────────────────────────┐
│          TiN (Top Electrode)    │  50-100nm PVD
├─────────────────────────────────┤
│    HfO₂-ZrO₂ Superlattice       │  4-10nm ALD
│    (Ferroelectric Layer)        │
├─────────────────────────────────┤
│          TiN (Bottom Electrode) │  50-100nm PVD
├─────────────────────────────────┤
│          Si Substrate           │
└─────────────────────────────────┘
```

### 8.2 ALD Deposition Process

**Precursors**:
- TEMAH/TEMAZ (Tetrakis[ethylmethylamino]hafnium/zirconium) - most common
- TDMAH/TDMAZ (lower carbon, better electrical performance)

**Process Recipe**:
```
Cycle Structure:
- Source feeding: 2s
- Source purge: 20s
- Ozone feeding: 3s (O₃ preferred over H₂O)
- Ozone purge: 10s

Growth rate: 0.13 nm/cycle
Substrate temperature: 240-280°C
Total cycles: ~60-80 for 8nm film
```

**Critical Parameters**:
- Temperature: 250°C (BEOL-compatible)
- Oxygen source: O₃ (lower impurity, lower leakage)
- Crystallization anneal: 400°C for 30s-1hr (RTA or furnace)

### 8.3 TiN Electrode Requirements

TiN provides:
- {111} texture transfer to HZO for better ferroelectricity
- CMOS compatibility (standard fab material)
- Thermal stability at 400°C
- Suitable work function for switching

**Enhancement**: Controlled ozone oxidation of bottom TiN improves endurance.

### 8.4 LEF File for Custom FeCIM Cell

```lef
VERSION 5.8 ;
BUSBITCHARS "[]" ;
DIVIDERCHAR "/" ;

MACRO fecim_1t1r_cell
  CLASS CORE ;
  FOREIGN fecim_1t1r_cell 0.0 0.0 ;
  ORIGIN 0.0 0.0 ;
  SIZE 1.84 BY 2.72 ;      # 4 sites × standard height
  SYMMETRY X Y ;
  SITE unithd ;

  PIN WL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
        RECT 0.0 1.20 0.14 1.52 ;
    END
  END WL

  PIN BL
    DIRECTION INOUT ;
    USE SIGNAL ;
    PORT
      LAYER met2 ;
        RECT 0.85 2.58 1.13 2.72 ;
    END
  END BL

  PIN SL
    DIRECTION INPUT ;
    USE SIGNAL ;
    PORT
      LAYER met1 ;
        RECT 1.70 1.20 1.84 1.52 ;
    END
  END SL

  PIN VPWR
    DIRECTION INOUT ;
    USE POWER ;
    PORT
      LAYER met1 ;
        RECT 0.0 2.48 1.84 2.72 ;
    END
  END VPWR

  PIN VGND
    DIRECTION INOUT ;
    USE GROUND ;
    PORT
      LAYER met1 ;
        RECT 0.0 0.0 1.84 0.24 ;
    END
  END VGND

  OBS
    LAYER li1 ;
      RECT 0.1 0.3 1.74 2.42 ;
    LAYER met1 ;
      RECT 0.2 0.4 1.64 2.32 ;
  END
END fecim_1t1r_cell

END LIBRARY
```

### 8.5 Liberty Timing Model (Stub)

```lib
library(fecim_1t1r) {
  technology (cmos);
  delay_model : table_lookup;
  time_unit : "1ns";
  voltage_unit : "1V";
  current_unit : "1mA";
  capacitive_load_unit (1, pf);

  nom_process : 1;
  nom_temperature : 25;
  nom_voltage : 1.8;

  cell(fecim_1t1r_cell) {
    area : 5.0048;  /* 1.84 × 2.72 μm² */

    pin(WL) {
      direction : input;
      capacitance : 0.002;
    }

    pin(BL) {
      direction : inout;
      capacitance : 0.005;
      max_capacitance : 0.5;
    }

    pin(SL) {
      direction : input;
      capacitance : 0.002;
    }

    pin(VPWR) {
      direction : inout;
      capacitance : 0.01;
    }

    pin(VGND) {
      direction : inout;
      capacitance : 0.01;
    }
  }
}
```

### 8.6 Foundry Capabilities for FeFET

| Foundry | FeFET Status | Notes |
|---------|--------------|-------|
| **TSMC** | R&D | Collaboration with Georgia Tech demonstrated |
| **GlobalFoundries** | R&D | 28nm/22nm FD-SOI FeFET demos |
| **Samsung** | R&D | Accelerating commercialization |
| **Intel** | R&D | Historical FeFET array demos |
| **IHP** | Custom | MEMRES module available |

**Key Insight**: No foundry offers FeFET via standard MPW programs. Requires custom R&D collaboration.

---

## 9. Complete Design-to-Wafer Workflow

### 9.1 Phase 0: Requirements (Week 1-2)

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 0: REQUIREMENTS DEFINITION                                │
├─────────────────────────────────────────────────────────────────┤
│ □ Define operation mode (Storage/Memory/Compute)                │
│ □ Specify array dimensions (rows × cols)                        │
│ □ Choose architecture (passive or 1T1R)                         │
│ □ Define conductance range (Gmin, Gmax)                         │
│ □ Set accuracy requirements (for compute mode)                  │
│ □ Identify target fabrication pathway                           │
│ □ Establish budget and timeline constraints                     │
└─────────────────────────────────────────────────────────────────┘
```

### 9.2 Phase 1: Design Entry (Week 2-4)

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 1: DESIGN ENTRY                                           │
├─────────────────────────────────────────────────────────────────┤
│ □ Configure Module 6 parameters                                 │
│   - go run ./cmd/eda-cli -mode compute -rows 64 -cols 64 ...    │
│                                                                 │
│ □ Load weights (compute mode only)                              │
│   - Format: JSON with "weights" array                           │
│                                                                 │
│ □ Generate outputs:                                             │
│   - Verilog netlist (.v)                                        │
│   - DEF placement (.def)                                        │
│   - SPICE netlist (.sp)                                         │
│   - Design data (.json, .csv)                                   │
│                                                                 │
│ □ Review quantization metrics (PSNR, MSE for compute mode)      │
└─────────────────────────────────────────────────────────────────┘
```

### 9.3 Phase 2: Simulation & Verification (Week 4-8)

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 2: SIMULATION & VERIFICATION                              │
├─────────────────────────────────────────────────────────────────┤
│ □ Functional verification                                       │
│   - Verilog testbench simulation (iverilog/Verilator)           │
│   - Verify I/O behavior                                         │
│                                                                 │
│ □ Array-level simulation                                        │
│   - CrossSim: accuracy, variability, IR drop                    │
│   - NeuroSim: energy, latency, area estimates                   │
│                                                                 │
│ □ Circuit-level verification                                    │
│   - ngspice: cell timing, power consumption                     │
│   - Verify against Liberty model                                │
│                                                                 │
│ □ Pre-layout DRC check                                          │
│   - Magic DRC on custom cells                                   │
└─────────────────────────────────────────────────────────────────┘
```

### 9.4 Phase 3: OpenLane Integration (Week 8-12)

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 3: OPENLANE INTEGRATION                                   │
├─────────────────────────────────────────────────────────────────┤
│ □ Prepare custom cell files                                     │
│   - LEF (abstract view)                                         │
│   - Liberty (timing)                                            │
│   - GDS (layout) - if available                                 │
│   - Verilog (behavioral model)                                  │
│                                                                 │
│ □ Configure OpenLane                                            │
│   - Set EXTRA_LEFS, EXTRA_LIBS                                  │
│   - Set PLACEMENT_CURRENT_DEF                                   │
│   - Set PL_SKIP_INITIAL_PLACEMENT=1                             │
│   - Set SYNTH_ELABORATE_ONLY=1 (structural netlist)             │
│                                                                 │
│ □ Run OpenLane flow                                             │
│   - flow.tcl or Python-based flow                               │
│   - Monitor each stage for errors                               │
│                                                                 │
│ □ Post-layout verification                                      │
│   - DRC clean (Magic)                                           │
│   - LVS clean (netgen)                                          │
│   - STA timing closure (OpenSTA)                                │
└─────────────────────────────────────────────────────────────────┘
```

### 9.5 Phase 4: Tape-Out Preparation (Week 12-16)

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 4: TAPE-OUT PREPARATION                                   │
├─────────────────────────────────────────────────────────────────┤
│ □ Final GDSII generation                                        │
│   - OpenLane final output                                       │
│   - Stream out with correct layer mapping                       │
│                                                                 │
│ □ Final DRC/LVS sign-off                                        │
│   - Run foundry DRC deck                                        │
│   - Full-chip LVS                                               │
│                                                                 │
│ □ Prepare submission package                                    │
│   - GDSII file                                                  │
│   - Layer mapping file                                          │
│   - Design summary document                                     │
│                                                                 │
│ □ Submit to shuttle                                             │
│   - Meet registration deadline                                  │
│   - Meet GDS submission deadline                                │
│   - Pay fabrication fees                                        │
└─────────────────────────────────────────────────────────────────┘
```

### 9.6 Phase 5: Fabrication & Test (Week 16-36)

```
┌─────────────────────────────────────────────────────────────────┐
│ PHASE 5: FABRICATION & TEST                                     │
├─────────────────────────────────────────────────────────────────┤
│ □ Wafer fabrication (8-16 weeks)                                │
│   - Monitor foundry status updates                              │
│                                                                 │
│ □ Packaging (2-4 weeks)                                         │
│   - QFN, WCSP, or bare die                                      │
│                                                                 │
│ □ Receive silicon                                               │
│   - Verify chip count and packaging                             │
│   - Review E-test data                                          │
│                                                                 │
│ □ Silicon characterization                                      │
│   - Basic functionality test                                    │
│   - I-V characterization                                        │
│   - Timing measurements                                         │
│   - Power measurements                                          │
│                                                                 │
│ □ Application testing                                           │
│   - MNIST inference (compute mode)                              │
│   - Write/read cycles (storage/memory mode)                     │
│   - Endurance testing                                           │
└─────────────────────────────────────────────────────────────────┘
```

---

## 10. Cost & Timeline Analysis

### 10.1 Standard CMOS Path (Peripheral Circuits)

| Phase | Duration | Cost |
|-------|----------|------|
| Design Entry | 2-4 weeks | $0 (Module 6) |
| Simulation | 2-4 weeks | $0 (open-source tools) |
| OpenLane Integration | 2-4 weeks | $0 |
| Tape-Out Prep | 2-4 weeks | $0 |
| Fabrication | 12-24 weeks | $0-$15,000 |
| **Total** | **5-9 months** | **$0-$15,000** |

### 10.2 Custom FeFET Path (Full Integration)

| Phase | Duration | Cost |
|-------|----------|------|
| Standard CMOS (above) | 5-9 months | $0-$15,000 |
| FeFET Process Development | 6-12 months | $50,000-$200,000 |
| BEOL Post-Processing | 2-4 months | $10,000-$50,000 |
| Integration & Test | 2-4 months | $5,000-$20,000 |
| **Total** | **12-24 months** | **$65,000-$285,000** |

### 10.3 Budget Breakdown by Pathway

| Pathway | Budget | Timeline | Ownership |
|---------|--------|----------|-----------|
| **Google OpenMPW** | FREE | 6-9 months | ✅ Yes |
| **Tiny Tapeout SKY130** | €300-€2,500 | 6 months | ✅ Yes |
| **Tiny Tapeout IHP** | €70-€2,240 | 6 months | ❌ Loan only |
| **ChipFoundry ChipCreate** | $14,950 | 5 months | ✅ Yes |
| **IHP Direct** | €5,840+ | 4-5 months | ✅ Yes |
| **EUROPRACTICE** | €900+/mm² | 3-6 months | ✅ Yes |
| **Custom FeFET R&D** | $65k-$285k | 12-24 months | ✅ Yes |

### 10.4 MPW Cost Comparison by Node

| Node | Typical Cost | Notes |
|------|--------------|-------|
| 180nm | $15k-$30k | Legacy, mature |
| 130nm (SKY130) | $0-$15k | Open-source options |
| 65nm | $45k-$70k | |
| 28nm | $70k-$130k | |
| FinFET (16nm, 7nm) | $180k-$350k+ | |

---

## 11. Risk Assessment & Mitigation

### 11.1 Technical Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| FeFET device variability | High | High | Use 1T1R architecture, calibration |
| OpenLane DRC failures | Medium | Medium | Iterative layout refinement |
| Timing closure failure | Medium | Medium | Conservative clock constraints |
| IR drop in large arrays | High | Medium | Segmented power distribution |
| Sneak paths (passive) | High | High | Use 1T1R, limit array size |

### 11.2 Schedule Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Shuttle delay | Medium | Medium | Buffer time in schedule |
| Foundry capacity | Low | High | Register early, backup options |
| Custom process negotiation | High | High | Start discussions early |

### 11.3 Cost Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Respin required | Medium | High | Thorough pre-silicon verification |
| Custom FeFET cost overrun | High | High | Phase approach, academic partnerships |
| Tool licensing | Low | Low | Use open-source tools |

### 11.4 Go/No-Go Decision Points

1. **After Simulation**: Proceed if CrossSim accuracy >85%, NeuroSim estimates within budget
2. **After DRC/LVS**: Proceed if clean with <10 waivers
3. **After Fabrication Quote**: Proceed if within budget ±20%
4. **After Silicon Test**: Proceed to production if yield >50% and specs met

---

## 12. References

### Project Documentation

- [Module 6 Demo Guide](./eda.demo.md)
- [Module 6 Overview](./README.md)
- [OpenLane Study](./OPENLANE_STUDY.md)
- [OpenLane Integration Guide](./eda.integration.md)
- [SKY130 Quick Reference](./SKY130.md)
- [EDA Research](./eda.research.md)
- [EDA Open Source Ecosystem](./eda.opensource.md)

### External Resources

**Open-Source EDA**
- [OpenLane Documentation](https://openlane.readthedocs.io/)
- [SKY130 PDK](https://github.com/google/skywater-pdk)
- [GF180MCU PDK](https://github.com/google/gf180mcu-pdk)
- [IHP-Open-PDK](https://github.com/IHP-GmbH/IHP-Open-PDK)

**Fabrication Services**
- [ChipFoundry](https://chipfoundry.io/)
- [Tiny Tapeout](https://tinytapeout.com/)
- [EUROPRACTICE](https://europractice-ic.com/)
- [IHP MPW Service](https://www.ihp-microelectronics.com/services/research-and-prototyping-service/mpw-prototyping-service)

**Simulation Tools**
- [CrossSim](https://github.com/sandialabs/cross-sim)
- [NeuroSim](https://github.com/neurosim)
- [AIHWKIT](https://github.com/IBM/aihwkit)
- [ngspice](https://ngspice.sourceforge.io/)

**FeFET Research**
- [Roadmap on Ferroelectric Hafnia/Zirconia (APL)](https://pubs.aip.org/aip/apm/article/11/8/089201/2908480/Roadmap-on-ferroelectric-hafnia-and-zirconia-based)
- [HZO Superlattice Endurance (ResearchGate)](https://www.researchgate.net/publication/357111874_HfO2-ZrO2_Superlattice_Ferroelectric_Capacitor_with_Improved_Endurance_Performance_and_Higher_Fatigue_Recovery_Capability)
- [BEOL FeFET Integration (ACS)](https://pubs.acs.org/doi/10.1021/acsami.0c00877)
- [Ferroelectric Stability in Superlattices (Nature)](https://www.nature.com/articles/s41467-025-61758-2)

**Industry Analysis**
- [What Ever Happened to Next-Gen Ferroelectric Memories?](https://marklapedus.substack.com/p/what-ever-happened-to-next-gen-ferroelectric)
- [Ferroelectric Memory Market 2025-2029](https://www.macnifico.pt/news-en/ferroelectric-memory-devices-2025-breakthroughs-set-to-double-market-growth-by-2029/86662/)

---

## Appendix A: Quick Reference Commands

```bash
# Module 6 CLI - Generate all outputs
go run ./cmd/eda-cli \
  -mode compute \
  -input weights.json \
  -rows 64 -cols 64 \
  -name fecim_array \
  -tech SKY130 \
  -arch 1T1R \
  -output ./output \
  -json=true -csv=true -spice=true -verilog=true -def=true

# OpenLane - Run flow with custom cells
cd $OPENLANE_ROOT
./flow.tcl -design fecim_array -config_file config.json

# Magic - DRC check
magic -dnull -noconsole -T sky130A << EOF
drc off
gds read output/fecim_array.gds
load fecim_array
drc check
drc catchup
drc count total
quit
EOF

# ngspice - Simulate FeCIM cell
ngspice -b fecim_cell.sp -o fecim_cell.log

# CrossSim - Run array simulation
python3 -c "
from crosssim import CrossSimParameters
params = CrossSimParameters()
params.load('fecim_config.json')
params.run_inference('mnist_test.npz')
"
```

---

## Appendix B: Glossary

| Term | Definition |
|------|------------|
| **ALD** | Atomic Layer Deposition - thin film deposition technique |
| **BEOL** | Back-End-of-Line - metal interconnect layers in CMOS |
| **CIM** | Compute-in-Memory - processing data where it's stored |
| **DEF** | Design Exchange Format - placement/routing file format |
| **DRC** | Design Rule Check - verify layout meets fab rules |
| **FeFET** | Ferroelectric Field-Effect Transistor |
| **GDSII** | Graphic Data System II - standard IC layout format |
| **HZO** | HfO₂-ZrO₂ - hafnium zirconium oxide ferroelectric |
| **LEF** | Library Exchange Format - abstract cell definition |
| **LVS** | Layout vs. Schematic - verify layout matches netlist |
| **MPW** | Multi-Project Wafer - shared fabrication run |
| **MVM** | Matrix-Vector Multiply - core CIM operation |
| **PDK** | Process Design Kit - fab-specific design files |
| **Pr** | Remanent Polarization - ferroelectric property |
| **RTL** | Register Transfer Level - digital design abstraction |
| **STA** | Static Timing Analysis - verify timing constraints |
| **TRL** | Technology Readiness Level (1-9 scale) |
| **1T1R** | One-Transistor-One-Resistor - memory cell architecture |

---

**Document End**

*This document will be updated as the FeCIM technology and open-source EDA ecosystem evolves.*
