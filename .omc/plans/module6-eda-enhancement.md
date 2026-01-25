# Module 6 EDA Enhancement Plan

> **Plan Type**: Feature Enhancement
> **Created**: 2026-01-24
> **Status**: READY FOR REVIEW
> **Estimated Complexity**: MEDIUM-HIGH

---

## 1. Context

### 1.1 Original Request

Enhance Module 6 EDA with:
1. OpenLane config.json Generator (enhanced version)
2. CrossSim/NeuroSim Config Export
3. Stub Liberty Timing Files (.lib)
4. Verilog Testbenches for MVM verification
5. Future: Peripheral Integration (DAC/ADC/TIA)
6. Documentation updates in `docs/eda/`

### 1.2 Current State Analysis

**Existing Export Capabilities** (`module6-eda/pkg/export/`):

| File | Format | Status |
|------|--------|--------|
| `verilog.go` | `.v` netlist | Complete |
| `array_verilog.go` | `.v` structural | Complete |
| `cell_verilog.go` | `.v` bitcell | Complete |
| `def.go` | `.def` placement | Complete |
| `spice.go` | `.sp` SPICE | Complete |
| `json.go` | `.json` data | Complete |
| `csv.go` | `.csv` data | Complete |
| `lef.go` | `.lef` abstract | Complete |
| `liberty.go` | `.lib` timing | Basic (stub only) |
| `openlane_config.go` | `config.json` | Basic (needs enhancement) |

**Key Types** (`module6-eda/pkg/compiler/types.go`):
- `ArrayConfig`: Main design configuration with mode, technology, architecture
- `ArrayDesign`: Complete design output with cells and stats
- `CellAssignment`: Individual cell programming
- `PeripheralConfig`: DAC/ADC/TIA parameters (exists but not used in exports)

**Configuration Types** (`module6-eda/pkg/config/types.go`):
- `CellConfig`: Single bitcell parameters for LEF/Liberty generation
- `ArrayConfig`: Array-level parameters for crossbar generation

### 1.3 Research Findings

**OpenLane 2.0 MACROS Format** (from [OpenLane Documentation](https://openlane2.readthedocs.io/en/latest/usage/using_macros.html)):
- Uses Python dataclass `Macro` with `lef`, `gds`, `lib` views
- JSON format with wildcards for timing corners: `"*_tt_025C_1v80": ["dir::lib0.lib"]`
- Instance naming follows hierarchical dot notation when flattened

**CrossSim 3.1 Parameters** (from [CrossSim GitHub](https://github.com/sandialabs/cross-sim)):
- JSON load/save support in `CrossSimParameters`
- Key parameters: rows, cols, Gmin/Gmax (uS), num_levels, cell_type
- Device models: conductance drift, programming errors, cycle-to-cycle noise
- ADC precision, bit slicing, weight resolution

**NeuroSim Configuration** (from [NeuroSim GitHub](https://github.com/neurosim/NeuroSim)):
- C++ header files (`Param.cpp`, `Cell.h`)
- `memcelltype`: 1=SRAM, 2=RRAM, 3=FeFET
- `accesstype`: 1=CMOS, 2=BJT, 3=diode, 4=none (crossbar)
- ADC modes: SAR vs MLSA, current vs voltage sensing
- Architecture: buffer sizes, tile dimensions

---

## 2. Work Objectives

### 2.1 Core Objective

Extend Module 6 export capabilities to support external simulator integration (CrossSim, NeuroSim) and enhanced OpenLane configuration, enabling complete design-to-simulation-to-fabrication workflow.

### 2.2 Deliverables

| Deliverable | Priority | Files |
|-------------|----------|-------|
| Enhanced OpenLane config.json | P1 | `openlane_config.go` |
| CrossSim config export | P1 | `crosssim.go` (new) |
| NeuroSim config export | P1 | `neurosim.go` (new) |
| Verilog testbench generator | P2 | `testbench.go` (new) |
| Enhanced Liberty timing | P2 | `liberty.go` (update) |
| Peripheral export preparation | P3 | `peripheral.go` (new, stub) |
| Documentation | P1 | `docs/eda/simulator-export.md` |

### 2.3 Definition of Done

- [ ] All new export functions have unit tests
- [ ] Generated files pass format validation (where possible)
- [ ] Documentation includes usage examples with disclaimers
- [ ] Existing tests continue to pass (`go test ./module6-eda/...`)
- [ ] Code follows existing patterns in `pkg/export/`

---

## 3. Guardrails

### 3.1 Must Have

- Maintain backward compatibility with existing `ArrayDesign` and `ArrayConfig` types
- Include clear PLACEHOLDER/DISCLAIMER comments for uncharacterized values
- Follow existing code patterns (generator functions returning strings)
- Proper error handling for edge cases
- Unit tests for all new exporters

### 3.2 Must NOT Have

- No breaking changes to existing export APIs
- No modification of `pkg/compiler/` types (use adapters if needed)
- No hard dependencies on external tools (CrossSim, NeuroSim are Python/C++)
- No fabrication-ready claims (educational/research tool disclaimer)

---

## 4. Task Flow

```
Phase 1: OpenLane Enhancement
    |
    v
Phase 2: Simulator Exports (CrossSim, NeuroSim)
    |
    v
Phase 3: Testbench & Liberty Enhancement
    |
    v
Phase 4: Peripheral Stub & Documentation
    |
    v
Phase 5: Testing & Verification
```

---

## 5. Detailed TODOs

### Phase 1: OpenLane Config Enhancement

**Task 1.1**: Enhance `openlane_config.go` for OpenLane 2.0 MACROS

**File**: `<local-path>`

**Changes**:
```go
// Add OpenLane 2.0 MACROS format support
// Current: simple JSON with flat structure
// New: MACROS dictionary with timing corner wildcards

type OpenLaneConfig struct {
    DesignName string
    PDK        string  // "sky130A", "gf180mcu", "ihp-sg13g2"
    Macros     map[string]MacroConfig
    // ... existing fields
}

type MacroConfig struct {
    LEF       string
    GDS       string
    Instances map[string]InstanceConfig
    Lib       map[string][]string  // Timing corner wildcards
}
```

**Acceptance Criteria**:
- [x] Generates valid OpenLane 2.0 JSON format
- [x] Supports MACROS dictionary with instance placement
- [x] Includes timing corner wildcards for Liberty files
- [x] Test validates JSON structure

---

### Phase 2: Simulator Config Exports

**Task 2.1**: Create CrossSim Configuration Exporter

**File**: `<local-path>` (NEW)

**Implementation**:
```go
// CrossSim 3.1 compatible JSON configuration
// Reference: https://github.com/sandialabs/cross-sim

type CrossSimConfig struct {
    // Array dimensions
    Rows int `json:"rows"`
    Cols int `json:"cols"`

    // Device parameters
    CellType    string  `json:"cell_type"`  // "FeFET"
    GminUS      float64 `json:"Gmin_uS"`
    GmaxUS      float64 `json:"Gmax_uS"`
    NumLevels   int     `json:"num_levels"` // 30 for FeCIM

    // Non-ideality models
    Variability     float64 `json:"variability_sigma"`
    DriftEnabled    bool    `json:"drift_enabled"`
    ReadNoiseStdDev float64 `json:"read_noise_std"`

    // ADC/DAC precision
    InputBits  int `json:"input_bits"`
    OutputBits int `json:"output_bits"`
    ADCRangeMin float64 `json:"adc_range_min"`
    ADCRangeMax float64 `json:"adc_range_max"`
}

func GenerateCrossSimConfig(design *compiler.ArrayDesign) string
func ExportCrossSimConfig(design *compiler.ArrayDesign, path string) error
```

**Acceptance Criteria**:
- [x] Exports JSON compatible with CrossSim 3.1 parameters
- [x] Maps ArrayConfig fields to CrossSim equivalents
- [x] Includes FeCIM-specific defaults (30 levels, 1-100 uS range)
- [x] Test validates JSON structure and field presence

---

**Task 2.2**: Create NeuroSim Configuration Exporter

**File**: `<local-path>` (NEW)

**Implementation**:
```go
// NeuroSim 2.1+ compatible configuration
// Reference: https://github.com/neurosim/DNN_NeuroSim_V2.1

type NeuroSimConfig struct {
    // Memory cell type (maps to memcelltype in Param.cpp)
    MemCellType int // 1=SRAM, 2=RRAM, 3=FeFET
    AccessType  int // 1=CMOS, 2=BJT, 3=diode, 4=none(crossbar)

    // Array dimensions
    NumRowSubArray int
    NumColSubArray int

    // Device parameters
    ResistanceOn  float64 // Ohms (low resistance state)
    ResistanceOff float64 // Ohms (high resistance state)
    NumLevels     int

    // Read circuit
    ReadVoltage float64 // V
    ReadPulseWidth float64 // ns

    // ADC configuration
    NumADC int
    ADCBits int
    SARADC bool // false=MLSA, true=SAR
    CurrentMode bool // false=VSA, true=CSA
}

func GenerateNeuroSimConfig(design *compiler.ArrayDesign) string
func ExportNeuroSimConfig(design *compiler.ArrayDesign, path string) error
```

**Acceptance Criteria**:
- [x] Exports configuration compatible with NeuroSim Param.cpp structure
- [x] Uses memcelltype=3 for FeFET by default
- [x] Maps conductance to resistance correctly (R = 1e6/G_uS)
- [x] Includes clear comments mapping to NeuroSim variables

---

### Phase 3: Testbench & Liberty Enhancement

**Task 3.1**: Create Verilog Testbench Generator

**File**: `<local-path>` (NEW)

**Implementation**:
```go
// Verilog testbench for MVM verification
// Generates stimulus for crossbar array testing

type TestbenchConfig struct {
    ModuleName    string
    InputVectors  [][]float64 // Test input patterns
    ExpectedOutputs [][]float64 // Expected MVM results
    ClockPeriodNs float64
    SimTimeNs     float64
}

func GenerateTestbench(design *compiler.ArrayDesign, cfg TestbenchConfig) string
func GenerateMVMTestbench(design *compiler.ArrayDesign) string // Default MVM test
func ExportTestbench(design *compiler.ArrayDesign, path string) error
```

**Testbench Structure**:
```verilog
`timescale 1ns/1ps
module tb_fecim_crossbar;
    // Clock and reset
    reg clk;
    reg rst_n;

    // Array interface
    reg [ROWS-1:0] WL;
    wire [COLS-1:0] BL;

    // DUT instantiation
    fecim_crossbar_NxM dut (...);

    // Test stimulus
    initial begin
        // Apply input vectors
        // Check output currents
        // Report pass/fail
    end
endmodule
```

**Acceptance Criteria**:
- [x] Generates syntactically valid Verilog testbench
- [x] Includes clock generation for synchronous testing
- [x] Provides basic MVM verification (apply WL, read BL)
- [x] Includes $display/$finish for simulation control
- [x] Test validates syntax with Icarus Verilog (if available)

---

**Task 3.2**: Enhance Liberty Timing File Generator

**File**: `<local-path>`

**Enhancements**:
```go
// Add multi-corner support for timing analysis
// Add NLDM (Non-Linear Delay Model) tables (placeholder values)

type LibertyConfig struct {
    CellConfig config.CellConfig
    Corner     string  // "tt", "ff", "ss"
    Temperature float64 // Celsius
    Voltage     float64 // V
}

func GenerateLibertyMultiCorner(cfg config.CellConfig) map[string]string
func GenerateLibertyWithNLDM(cfg config.CellConfig) string
```

**New Liberty Features**:
- Operating conditions for multiple corners (tt/ff/ss)
- Temperature derating (25C, -40C, 125C)
- Voltage corners (1.65V, 1.8V, 1.95V for SKY130)
- NLDM table templates (placeholder 2x2 tables)

**Acceptance Criteria**:
- [x] Generates Liberty with operating_conditions for multiple corners
- [x] Includes NLDM table structure (even with placeholder values)
- [x] Clear PLACEHOLDER disclaimers remain
- [x] Existing tests continue to pass

---

### Phase 4: Peripheral Stub & Documentation

**Task 4.1**: Create Peripheral Export Stub

**File**: `<local-path>` (NEW)

**Implementation**:
```go
// Stub for future peripheral circuit integration
// DAC/ADC/TIA export for complete CIM accelerator

type PeripheralExportConfig struct {
    IncludeDAC bool
    IncludeADC bool
    IncludeTIA bool
    DACBits    int
    ADCBits    int
    TIAGain    float64
}

// Placeholder functions for future implementation
func GenerateDACVerilog(cfg compiler.PeripheralConfig) string
func GenerateADCVerilog(cfg compiler.PeripheralConfig) string
func GenerateTIASpice(cfg compiler.PeripheralConfig) string
```

**Acceptance Criteria**:
- [x] Stub functions return TODO comments
- [x] Uses existing `PeripheralConfig` from compiler types
- [x] Documentation explains future roadmap

---

**Task 4.2**: Create Simulator Export Documentation

**File**: `<local-path>` (NEW)

**Content Structure**:
```markdown
# FeCIM Simulator Export Guide

## Overview
How to use Module 6 exports with external CIM simulators

## CrossSim Integration
- Installation and setup
- Loading exported JSON
- Running inference simulation
- Example Python code

## NeuroSim Integration
- Building NeuroSim
- Configuring Param.cpp from export
- Running area/power estimation
- Interpreting results

## Testbench Usage
- Running with Icarus Verilog
- Waveform viewing (GTKWave)
- MVM verification methodology

## Limitations and Disclaimers
- PLACEHOLDER values
- Educational/research use only
- Path to characterization
```

**Acceptance Criteria**:
- [x] Clear step-by-step instructions
- [x] Code examples for each simulator
- [x] Explicit disclaimers about placeholder values
- [x] References to official simulator documentation

---

**Task 4.3**: Update Module 6 README

**File**: `<local-path>`

**Updates**:
- Add simulator export to capabilities table
- Add testbench generation to outputs
- Reference new `simulator-export.md`
- Update "What Gets Generated" section

---

### Phase 5: Testing & Verification

**Task 5.1**: Add Unit Tests for New Exporters

**File**: `<local-path>` (NEW)

**Tests**:
```go
func TestGenerateCrossSimConfig(t *testing.T)
func TestGenerateNeuroSimConfig(t *testing.T)
func TestGenerateTestbench(t *testing.T)
func TestGenerateLibertyMultiCorner(t *testing.T)
```

**Acceptance Criteria**:
- [x] All new functions have corresponding tests
- [x] Tests validate JSON/Verilog structure
- [x] Edge cases covered (empty array, single cell, large array)

---

**Task 5.2**: Integration Test

**Command**: `go test ./module6-eda/...`

**Acceptance Criteria**:
- [x] All existing tests pass
- [x] All new tests pass
- [x] No race conditions (if applicable)

---

## 6. Commit Strategy

| Commit | Scope | Message |
|--------|-------|---------|
| 1 | Phase 1 | `feat(eda): enhance OpenLane config.json for v2.0 MACROS format` |
| 2 | Phase 2 | `feat(eda): add CrossSim and NeuroSim config exporters` |
| 3 | Phase 3 | `feat(eda): add Verilog testbench generator for MVM verification` |
| 4 | Phase 3 | `feat(eda): enhance Liberty with multi-corner and NLDM support` |
| 5 | Phase 4 | `feat(eda): add peripheral export stubs for future DAC/ADC/TIA` |
| 6 | Phase 4 | `docs(eda): add simulator export guide and update README` |
| 7 | Phase 5 | `test(eda): add unit tests for new export functions` |

---

## 7. Success Criteria

### Functional

- [ ] `GenerateCrossSimConfig()` produces valid JSON loadable by CrossSim
- [ ] `GenerateNeuroSimConfig()` produces configuration matching NeuroSim Param.cpp structure
- [ ] `GenerateTestbench()` produces Verilog that compiles with `iverilog`
- [ ] `GenerateOpenLaneConfig()` produces OpenLane 2.0 compatible JSON
- [ ] Enhanced Liberty includes multi-corner operating conditions

### Quality

- [ ] 100% test coverage for new functions
- [ ] All existing tests pass
- [ ] Documentation reviewed for accuracy
- [ ] Code follows existing patterns in `pkg/export/`

### Documentation

- [ ] `simulator-export.md` provides complete usage guide
- [ ] README updated with new capabilities
- [ ] All placeholder values clearly marked

---

## 8. Dependencies and Risks

### Dependencies

| Dependency | Type | Mitigation |
|------------|------|------------|
| Existing `ArrayDesign` type | Code | Use adapters, no modifications |
| CrossSim JSON format | External | Follow published documentation |
| NeuroSim C++ structure | External | Generate comments, not compilable C++ |

### Risks

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| CrossSim format changes | Low | Medium | Version-specific generator |
| NeuroSim incompatibility | Low | Low | Export as reference, not direct import |
| Testbench too complex | Medium | Low | Start with simple MVM verification |

---

## 9. References

### External Documentation

- [OpenLane 2.0 MACROS](https://openlane2.readthedocs.io/en/latest/usage/using_macros.html)
- [CrossSim GitHub](https://github.com/sandialabs/cross-sim)
- [NeuroSim GitHub](https://github.com/neurosim/NeuroSim)
- [Liberty Format Reference](https://people.eecs.berkeley.edu/~alanmi/publications/other/liberty07_03.pdf)

### Internal Documentation

- `<local-path>`
- `<local-path>`
- `<local-path>`

---

**PLAN_READY: .omc/plans/module6-eda-enhancement.md**
