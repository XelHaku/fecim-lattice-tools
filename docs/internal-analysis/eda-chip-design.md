# Research Synthesis: EDA and Chip Design for FeCIM

> **Note:** Internal analysis note. Values are reported/illustrative and not validated by this codebase.

> **Internal Analysis Document** - FeCIM Lattice Tools Project

## 1. Executive Summary

This document synthesizes the Electronic Design Automation (EDA) workflows, PDK integrations, and fabrication pathways for the FeCIM project. The project uses the open-source OpenLane RTL-to-GDSII flow with SkyWater 130nm (SKY130) and GlobalFoundries 180nm (GF180MCU) PDKs. Key innovations include custom FeCIM cell injection, analog block integration, and three design modes (Storage, Memory, Compute).

---

## 2. RTL-to-GDSII Flow

The complete flow transforms hardware description to manufacturable layout.

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  RTL/Verilog│───▶│  Synthesis  │───▶│  Floorplan  │───▶│  Placement  │
│  (Module 6) │    │   (Yosys)   │    │ (OpenROAD)  │    │ (OpenROAD)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                                                                │
┌─────────────┐    ┌─────────────┐    ┌─────────────┐           ▼
│    GDSII    │◀───│   Signoff   │◀───│   Routing   │◀───│     CTS     │
│   (Final)   │    │   (Magic)   │    │ (OpenROAD)  │    │ (OpenROAD)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
```

### 2.1 Synthesis (Yosys)
- **Purpose**: RTL → Gate-level netlist
- **Key Variables**: `SYNTH_ELABORATE_ONLY` (for structural netlists)
- **FeCIM Usage**: Structural elaboration without logic mapping for analog blocks

### 2.2 Floorplan (OpenROAD)
- **Purpose**: Define die area, core utilization, IO placement
- **Key Variables**: `FP_DEF_TEMPLATE` (floorplan template), `DIE_AREA`
- **FeCIM Usage**: Pre-define FeCIM array regions

### 2.3 Placement (OpenROAD)
- **Purpose**: Position standard cells in the core area
- **Key Variables**: `PLACEMENT_CURRENT_DEF` (inject pre-placed DEF)
- **FeCIM Usage**: Fixed crossbar array positions via DEF injection

### 2.4 Clock Tree Synthesis (CTS)
- **Purpose**: Build clock distribution network
- **Key Variables**: `CLOCK_PERIOD`, `CTS_TARGET_SKEW`
- **FeCIM Usage**: Standard CTS for peripheral digital logic

### 2.5 Routing (OpenROAD)
- **Purpose**: Global + detailed routing of signal nets
- **Key Variables**: `ROUTING_CORES`, `DETAILED_ROUTER`
- **FeCIM Usage**: Route around pre-placed analog blocks

### 2.6 Signoff (Magic/KLayout/netgen)
- **Purpose**: DRC, LVS verification, GDSII generation
- **Tools**: Magic (DRC), netgen (LVS), KLayout (visualization)
- **FeCIM Usage**: Validate custom cells against PDK rules

---

## 3. OpenLane Integration

OpenLane provides the complete RTL-to-GDSII framework with customization hooks.

### 3.1 Key Configuration Variables

| Variable | Purpose | FeCIM Usage |
|----------|---------|-------------|
| `EXTRA_LEFS` | Custom cell LEF files | FeCIM bitcell definition |
| `EXTRA_GDS_FILES` | GDS for GDSII merge | Physical layout overlay |
| `PLACEMENT_CURRENT_DEF` | Inject pre-placed DEF | Fixed FeCIM cell positions |
| `FP_DEF_TEMPLATE` | Floorplan template | Pin alignment |
| `SYNTH_ELABORATE_ONLY` | Skip logic mapping | Structural netlists |
| `PL_SKIP_INITIAL_PLACEMENT` | Use pre-placed cells | Fixed crossbar array |
| `CELL_PAD` | Cell padding | Array spacing |
| `VERILOG_FILES` | Input RTL | Module 6 generated .v files |

### 3.2 FeCIM Cell Injection Strategy

```json
{
  "DESIGN_NAME": "fecim_crossbar_4x4",
  "VERILOG_FILES": ["fecim_crossbar.v"],
  "EXTRA_LEFS": ["fecim_bitcell.lef"],
  "EXTRA_GDS_FILES": ["fecim_bitcell.gds"],
  "PLACEMENT_CURRENT_DEF": "fecim_preplaced.def",
  "PL_SKIP_INITIAL_PLACEMENT": 1,
  "SYNTH_ELABORATE_ONLY": 1
}
```

### 3.3 Validation Workflow

1. **Yosys Check**: `yosys -p "read_verilog fecim.v; proc; check"`
2. **OpenROAD Check**: Verify DEF compatibility
3. **DRC Check**: `magic -dnull -noconsole -T sky130A.tech drc.tcl`
4. **LVS Check**: `netgen -batch lvs fecim.spice fecim_extracted.spice`

---

## 4. PDK Support

| PDK | Node | Cell Library | Site | Cell Height | Status |
|-----|------|--------------|------|-------------|--------|
| **SKY130** | 130nm | sky130_fd_sc_hd | unithd | 2.72 µm | Default |
| **GF180MCU** | 180nm | gf180mcu_fd_sc_mcu7t5v0 | unit7t5v0 | 4.07 µm | Supported |
| **IHP_SG13G2** | 130nm | (BiCMOS, RRAM) | - | - | Experimental |

### 4.1 SKY130 Details

- **Metal Stack**: 5 metal layers (li1, met1-5)
- **High Density Lib**: sky130_fd_sc_hd (High Density)
- **High Speed Lib**: sky130_fd_sc_hs (High Speed)
- **Track Pitch**: 0.46 µm horizontal, 0.34 µm vertical
- **Via Size**: 0.17 µm × 0.17 µm

### 4.2 GF180MCU Details

- **Metal Stack**: 5 metal layers
- **Voltage**: 5V I/O support (useful for charge pump)
- **Specialty**: High-voltage transistors available

---

## 5. Export Formats

Module 6 generates 8 output formats for EDA tool compatibility.

| Format | Extension | Purpose | Tool Compatibility |
|--------|-----------|---------|-------------------|
| **Verilog** | .v | Structural netlist | Yosys, OpenLane, VCS |
| **DEF** | .def | Cell placement (FIXED) | OpenLane, OpenROAD |
| **LEF** | .lef | Cell library abstraction | OpenLane P&R |
| **Liberty** | .lib | Timing library | OpenSTA, OpenLane |
| **SPICE** | .sp | Analog simulation | ngspice, HSPICE |
| **GDS** | .gds | Physical layout | KLayout, Magic |
| **JSON** | .json | OpenLane configuration | OpenLane |
| **CSV** | .csv | Weight/mapping data | Analysis tools |

---

## 6. Fabrication Pathways

### 6.1 Tiny Tapeout
- **Access**: Low-cost multi-project wafer
- **Size**: 150 µm × 170 µm tiles
- **Cost**: ~$100-500 per tile
- **Turnaround**: 3-6 months
- **URL**: https://tinytapeout.com

### 6.2 Efabless OpenMPW
- **Access**: Free shuttles (Google-sponsored)
- **PDK**: SKY130
- **Size**: Full-size designs (up to ~1mm²)
- **Turnaround**: 4-6 months
- **URL**: https://efabless.com/open_mpw

### 6.3 IHP SG13S Shuttle
- **Access**: European shuttle program
- **Features**: RRAM integration possible
- **Cost**: Research pricing available
- **URL**: https://www.ihp-microelectronics.com

---

## 7. BEOL Integration Research

Recent advances in Back-End-of-Line (BEOL) ferroelectric integration.

| Organization | Topic | Year | Key Finding |
|--------------|-------|------|-------------|
| **CEA-Leti** | 22nm BEOL FeRAM | Dec 2024 | 0.0028 µm² functional capacitors, 3D integration |
| **Samsung** | 3D Vertical FeFET NAND | 2024 | 128-layer demonstrated, 5-bit/cell |
| **GlobalFoundries** | 22FDX FeFET | 2024 | Production FeFET option available |
| **Fraunhofer IPMS** | AEC-Q100 Qualification | 2024 | Automotive -40°C to 150°C testing |
| **IHP** | Open PDK with RRAM | 2025 | SG13S memristive module |

### 7.1 CEA-Leti 22nm BEOL Platform

- **Capacitor Area**: 0.0028 µm²
- **Process**: 3D BEOL-compatible
- **Thermal Budget**: <500°C (CMOS compatible)
- **Application**: Embedded FeRAM for MCUs

---

## 8. FeCIM-Specific Considerations

### 8.1 Three Design Modes

| Mode | Application | Array Focus | Key Metric |
|------|-------------|-------------|------------|
| **Storage** | NAND replacement | High density | 4.9 bits/cell |
| **Memory** | DRAM replacement | Zero refresh | Low standby power |
| **Compute** | AI accelerator | MVM accuracy | 885 TOPS/W |

### 8.2 Custom Cell Integration

FeCIM requires custom analog cells that standard PDKs don't include:

1. **FeFET Bitcell**: Custom LEF/GDS for ferroelectric transistor
2. **TIA Block**: Analog transimpedance amplifier macro
3. **Charge Pump**: High-voltage generation circuit
4. **DAC/ADC**: Mixed-signal peripherals

### 8.3 Analog-Digital Interface

```
Digital Peripherals (OpenLane Standard Flow)
    │
    ├── Clock/Reset Distribution
    ├── Control Logic
    └── Digital I/O
           │
           ▼
[Analog Boundary - Custom Cells]
           │
    ├── DAC Array (Input encoding)
    ├── FeCIM Crossbar (Pre-placed)
    ├── TIA Array (Current sensing)
    └── ADC Array (Output quantization)
```

---

## 9. References

### EDA Tools
- **OpenLane**: https://github.com/The-OpenROAD-Project/OpenLane
- **OpenROAD**: https://github.com/The-OpenROAD-Project/OpenROAD
- **Yosys**: https://github.com/YosysHQ/yosys
- **Magic**: https://github.com/RTimothyEdwards/magic
- **KLayout**: https://www.klayout.de

### PDK Documentation
- **SKY130**: https://skywater-pdk.readthedocs.io
- **GF180MCU**: https://gf180mcu-pdk.readthedocs.io

### BEOL Research
- CEA-Leti 22nm BEOL FeRAM Platform, IEDM December 2024
- Samsung FeFET Nature 2025, DOI: 10.1038/s41586-025-09793-3
- 3D Stacking READMEs: `/docs/research-papers/by-topic/21-3d-stacking/README.md`

### Internal Documentation
- `/docs/eda/guides/integration.md` - OpenLane integration guide
- `/docs/eda/WORKFLOW.md` - Complete RTL-to-GDSII workflow
- `/docs/pdk-reference/SKY130_QUICK_REFERENCE.md` - PDK technical reference
