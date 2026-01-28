# Module 6: EDA Design Suite - Architecture Documentation

## Overview

Module 6 (EDA - Electronic Design Automation) is the fabrication-grade design infrastructure for the FeCIM Lattice Tools suite. It converts neural network weights and design specifications into complete, fabrication-ready chip layouts through a modular pipeline of compilation, format export, and multi-stage validation.

**Core Purpose**: Transform high-level FeCIM array specifications into industry-standard physical design files (GDSII via OpenLane) for semiconductor fabrication.

**Supported Technology Nodes**: SkyWater 130nm (SKY130), GlobalFoundries 180nm (GF180MCU), IHP 130nm (IHP_SG13G2)

**Operation Modes**:
- **Storage**: High-density non-volatile NAND replacement (30 levels/cell, ~4.9 bits/cell)
- **Memory**: High-speed zero-refresh DRAM replacement (10ns access time)
- **Compute**: AI accelerator with optional pre-trained weight initialization

## Architecture Overview

The module is organized into seven core packages working in a coordinated pipeline:

```
Input (ArrayConfig with optional weights)
    ↓
[Compiler] → [Export] → [Validation] → [OpenLane] → [GUI]
    ↓           ↓           ↓            ↓
Physical   Multiple    Cross-file   Tool
Design      Formats     Checks      Integration
```

### Data Flow: Weights to GDSII

```
1. Configuration Phase
   └─ ArrayConfig specifies: mode, size, technology, peripherals

2. Compilation Phase (pkg/compiler/)
   ├─ GenerateDesign() → ArrayDesign
   ├─ mapWeights() [compute mode with weights]
   └─ GenerateBlank() [memory/storage modes]

3. Export Phase (pkg/export/)
   ├─ JSON → Data interchange format
   ├─ CSV  → Spreadsheet format
   ├─ SPICE→ ngspice/HSPICE simulation netlist
   ├─ Verilog → Structural netlist (hierarchical)
   │  ├─ Array-level (fecim_crossbar module)
   │  └─ Cell-level (fecim_bit or fecim_1t1r instances)
   ├─ DEF → Placement constraints for OpenLane
   ├─ LEF → Library cell abstraction
   ├─ Liberty → Timing/power characteristics
   └─ SVG → Visual layout for debugging

4. Validation Phase (pkg/validation/)
   ├─ Yosys (Verilog syntax)
   ├─ DEF Geometry (placement bounds)
   ├─ OpenLane DRC/placement
   ├─ Cross-file consistency
   ├─ Circuit image rendering
   └─ Layout image rendering

5. OpenLane Integration (pkg/openlane/)
   ├─ Mode detection (Docker/native)
   ├─ Runner execution (Yosys/OpenROAD)
   ├─ Manager (PDK discovery)
   └─ Config (flow parameters)

6. Physical Design Generation
   └─ OpenLane flow → GDSII (fabrication-ready)
```

---

## Package Details

### 1. Compiler Package (`pkg/compiler/`)

**Responsibility**: Converts ArrayConfig + weights into ArrayDesign physical layout.

**Key Types**:

#### OperationMode (enum)
```go
const (
    ModeStorage OperationMode = iota  // NAND replacement
    ModeMemory                         // DRAM replacement
    ModeCompute                        // AI accelerator
)
```

#### ArrayConfig
Complete array design specification:
- **Basic**: Name, Mode, ArrayRows, ArrayCols
- **Technology**: Technology node, Architecture (passive/1T1R/2T1R), CellPitch, RowHeight
- **Electrical**: GMin/GMax (μS), VProgMin/VProgMax (V), TPulse (ns), Levels (default: 30)
- **Peripherals**: DACBits, ADCBits, TIAGain (Ω), VDD, ClockFreq (MHz)
- **Mode-Specific**:
  - StorageArrayConfig: RetentionYears, EnduranceCycles, ErrorCorrection
  - MemoryArrayConfig: AccessTimeNs, WriteTimeNs, BandwidthGBps, PowerBudgetMW
  - ComputeArrayConfig: InitialWeights, QuantLevels, AccumulatorBits, ActivationFunc

**Constructor Patterns**:
```go
NewArrayConfig(mode, rows, cols)      // Base factory
NewStorageConfig(rows, cols)           // Storage-optimized
NewMemoryConfig(rows, cols)            // Memory-optimized
NewComputeConfig(rows, cols)           // Compute-optimized
config.With1T1R()                      // Architecture selection
config.With2T1R()
config.WithWeights(weights)            // Attach weights (compute mode)
```

#### CellAssignment
Physical cell after compilation:
- **Position**: Row, Col (array coordinates)
- **Programming**: Level (0 to Levels-1), Conductance (μS), Resistance (Ω)
- **Programming Voltage**: ProgramV (V)
- **Compute Mode**: InitialWeight (mapped from weights)

#### ArrayDesign
Complete design output:
- Config: ArrayConfig used to generate
- Cells: []CellAssignment (one per physical cell)
- Stats: DesignStats (power, area, utilization)

#### DesignStats
- TotalCells, ActiveCells
- AreaMM2, PowerMW, ThroughputGOPS
- **Compute mode only**: QuantMSE, QuantPSNR (dB), WeightMin/Max

**Core Functions**:

| Function | Purpose |
|----------|---------|
| `GenerateDesign(config)` | Main entry point: dispatch to mapWeights or GenerateBlank |
| `GenerateBlank(config)` | Initialize array without weights (Level 0, GMin) |
| `mapWeights(config)` | Quantize weight matrix to 30 levels, calculate stats |

**Weight Quantization Logic** (mapWeights):
1. Find weight range: [wMin, wMax]
2. Normalize weight: `norm = (w + wAbsMax) / (2 * wAbsMax)` (range [0,1])
3. Quantize: `level = round(norm * (Levels-1))`
4. Dequantize for stats: `qValue = -wAbsMax + qNorm * (2 * wAbsMax)`
5. Map to conductance: `G = GMin + qNorm * (GMax - GMin)`
6. Calculate MSE and PSNR for quality assessment

---

### 2. Export Package (`pkg/export/`)

**Responsibility**: Convert ArrayDesign into industry-standard formats for simulation, synthesis, placement, and fabrication.

**Supported Formats**:

| Format | File | Purpose | Consumer |
|--------|------|---------|----------|
| JSON | `.json` | Complete design data, version control | Data analysis, APIs |
| CSV | `.csv` | Tabular cell assignments | Spreadsheets, pandas |
| SPICE | `.sp` | Resistor network simulation netlist | ngspice, HSPICE |
| Verilog | `.v` | Structural cell instances | Yosys, Vivado, simulation |
| DEF | `.def` | Cell placements with coordinates | OpenLane placement |
| LEF | `.lef` | Cell macro abstraction | OpenLane library |
| Liberty | `.lib` | Timing/power models | OpenLane STA/power |
| SVG | `.svg` | 2D visual layout | Debug visualization |
| OpenLane Config | `config.json` | Flow parameter overrides | OpenLane runner |

**Module Structure**:

```
export/
├── json.go                # JSON serialization
├── csv.go                 # CSV tabular format
├── spice.go               # SPICE netlist (resistors + nets)
├── verilog.go             # Array-level module generation
├── cell_verilog.go        # Cell library (fecim_bit, fecim_1t1r)
├── array_verilog.go       # Legacy array instantiation
├── def.go                 # DEF placement file (FIXED keyword)
├── lef.go                 # LEF library cell abstraction
├── liberty.go             # Liberty timing/power file
├── svg.go                 # SVG visualization
├── openlane_config.go     # OpenLane configuration override
└── lattice_generator.go   # Batch design generation
```

**Key Patterns**:

#### Verilog Generation (verilog.go)
- Architecture support: passive (WL/BL) vs 1T1R (WL/BL/SL)
- Module declaration with port parameters
- Cell instantiation: `R_{row}_{col}` naming convention
- Comments include weight values and conductance for compute mode
- Filters active cells (weight matrix bounds for compute mode)

**Example Output Structure**:
```verilog
module fecim_crossbar (
    input wire [rows-1:0] WL,
    inout wire [cols-1:0] BL,
    input wire [cols-1:0] SL,  // 1T1R only
    input wire VDD, VSS
);
    parameter ROWS = N;
    parameter COLS = M;
    parameter LEVELS = 30;
    parameter MODE = "Compute";

    fecim_bit #(.LEVEL(5)) R_0_0 (.WL(WL[0]), .BL(BL[0]), .VDD(VDD), .VSS(VSS));
    // ... more cells
endmodule
```

#### DEF Generation (def.go)
- FIXED placement keyword for OpenLane PLACEMENT_CURRENT_DEF
- Coordinate calculation: X = OriginX + Col * CellWidth, Y = OriginY + Row * CellHeight
- Units: Database units (default 1000 DBU = 1 μm)
- Die boundary includes margins for power/ground rails

#### LEF Export (lef.go)
- MACRO definition (cell abstraction)
- PIN declarations (WL, BL, SL, VDD, VSS)
- LAYER definitions (metal1, poly)
- Geometry: PORT and OBS (obstruction) statements

#### Liberty Export (liberty.go)
- Timing arcs (input→output delays)
- Power consumption (dynamic, leakage)
- Pin capacitances
- **Note**: Placeholder values—requires SPICE characterization for accuracy

**OpenLane Integration Pattern**:
```go
// Generated files inject into OpenLane config
config := map[string]interface{}{
    "VERILOG_FILES": []string{"dir::crossbar.v"},
    "PLACEMENT_CURRENT_DEF": "dir::crossbar.def",
    "SYNTH_ELABORATE_ONLY": 1,
    "PL_SKIP_INITIAL_PLACEMENT": 1,
}
```

---

### 3. Configuration Package (`pkg/config/`)

**Responsibility**: Type definitions for cell and array configurations with sensible defaults.

**Key Types**:

#### CellConfig
Single FeCIM bitcell specification:
- **Geometry**: Name, Width (0.46 μm SKY130), Height (2.72 μm SKY130), CellType
- **Technology**: PDK target (sky130, gf180, ihp)
- **Timing** (PLACEHOLDER—requires SPICE characterization):
  - RiseTime, FallTime (ns)
  - InputCap (pF), LeakagePower (nW)

#### ArrayConfig
Array-level configuration:
- Rows, Cols, Mode, Architecture, Technology
- CellWidth, CellHeight (derived from CellConfig)

**Defaults**:
- Cell: 0.46 × 2.72 μm (SKY130 unithd site)
- Array: 4 × 4 (minimal test array)
- Mode: "storage", Architecture: "passive"

---

### 4. Validation Package (`pkg/validation/`)

**Responsibility**: Multi-stage verification ensuring design correctness before fabrication.

**Validators**:

#### Yosys Validation (yosys.go)
**Function**: Verilog syntax and hierarchy checking via Yosys.

| Function | Purpose |
|----------|---------|
| `ValidateVerilog(path)` | Single file syntax check |
| `ValidateVerilogWithCell(arrayPath, cellPath)` | Array + cell library validation |

**Execution Modes**:
- **Docker**: Volume mount workDir → /design, run yosys in container
- **Native**: Direct execution (requires yosys in PATH)

**Commands**:
```yosys
read_verilog array.v
read_verilog -lib cell.v  // Cell as blackbox library
hierarchy -check          // Verify module instantiation
```

#### DEF Validator (def_validator.go)
**Purpose**: Geometric correctness of placement.
- Coordinate bounds validation
- Cell overlap detection
- Die boundary compliance

#### OpenLane Validator (openlane.go)
**Purpose**: Placement and DRC checks via OpenLane.
- Cell legalization
- Design rule checking (spacing, width violations)
- Routing feasibility

#### Cross-Check Validator (cross_check.go)
**Purpose**: Consistency across LEF, Liberty, Verilog.

| Check | Purpose |
|-------|---------|
| Cell name match | Same cell name in all three formats |
| Pin names match | Port/pin definitions align |
| Pin count match | All formats list same I/O pins |

**Extraction Methods**:
- LEF: Regex for `MACRO` and `PIN` keywords
- Liberty: Regex for `cell(...)` and `pin(...)` blocks
- Verilog: Regex for `module` and `(input|output|inout)` ports

#### Layout Visualization (circuit_image.go, layout_image.go)
**Purpose**: Visual debugging of physical designs.
- Render cell placements as 2D grid
- Show conductance levels as color gradient
- Export as PNG for inspection

---

### 5. OpenLane Package (`pkg/openlane/`)

**Responsibility**: Tool integration for physical design flow (Yosys → OpenROAD → GDS).

**Key Components**:

#### Config (config.go)
OpenLane flow parameters:

| Parameter | Purpose | Default |
|-----------|---------|---------|
| PDKRoot | PDK installation path | `~/.volare` |
| PDKVariant | PDK version | `sky130A` |
| SCLibrary | Standard cell library | `sky130_fd_sc_hd` |
| PreferredMode | Execution mode | Docker |
| Timeouts | Tool-specific limits | 5-15 minutes |
| DockerImage | Container image | `ghcr.io/the-openroad-project/openlane:latest` |

**Methods**:
- `DefaultConfig()`: SKY130 setup with volare defaults
- `LoadConfig(path)`: Read from ~/.fecim/openlane-config.json
- `SaveConfig(cfg, path)`: Persist configuration
- `GetTechLEFPath()`: Return path to tech.lef
- `GetCellLEFPath()`: Return path to cell library LEF
- `GetLibertyPath()`: Return path to timing model

#### Manager (manager.go)
Tool availability detection and Docker image management.

| Function | Purpose |
|----------|---------|
| `DetectMode()` | Best available execution mode (Docker > Native > None) |
| `IsDockerAvailable()` | Check if docker command exists |
| `IsDockerImagePulled()` | Check if OpenLane image is cached locally |
| `PullDockerImage(progress)` | Download image with progress callback |
| `IsNativeOpenROADAvailable()` | Check for openroad in PATH |
| `IsNativeYosysAvailable()` | Check for yosys in PATH |
| `IsPDKInstalled()` | Verify PDK_ROOT contains valid sky130A |

**Mode Enum**:
```go
const (
    ModeNone Mode = iota    // No tools available
    ModeDocker              // Use Docker container
    ModeNative              // Use native binaries
)
```

#### Runner (runner.go)
Tool execution with timeout and environment variable support.

| Function | Purpose |
|----------|---------|
| `RunYosys(cmd, workDir)` | Execute Yosys synthesis commands |
| `RunOpenROAD(scriptName, workDir, envVars)` | Run OpenROAD TCL scripts |

**Result Type**:
```go
type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Duration time.Duration
}
```

**Docker Integration**:
- Volume mount workDir → /design
- Set DISPLAY for X11 GUI (if available)
- Pass PDK_ROOT and other env vars via `-e` flags
- Timeout management per tool phase

---

### 6. GUI Package (`pkg/gui/`)

**Responsibility**: Fyne-based interactive design environment.

**Structure**:

```
gui/
├── app.go                    # Main application controller
├── embedded.go               # Embedded app interface (BuildContent/Start/Stop)
├── tabs/
│   ├── learn_tab.go         # Educational overview
│   ├── builder_validation_tab.go  # Design builder + validation
│   ├── learn_visuals.go     # Visualization tab
│   ├── learn_visuals_cell.go      # Cell-level rendering
│   ├── learn_visuals_array.go     # Array-level rendering
│   └── learn_visuals_transistor.go # Circuit diagram
└── widgets/
    └── layout_canvas.go     # Placement visualization widget
```

**Embedded App Interface**:
```go
type EmbeddedApp interface {
    BuildContent() fyne.CanvasObject  // Return UI
    Start()                            // Initialize
    Stop()                             // Cleanup
}
```

---

## Design Patterns

### 1. Pipeline Architecture
- **Compiler** → **Export** → **Validate** → **OpenLane** → **Fabrication**
- Each stage is independent; formats are standard (JSON, Verilog, DEF)
- Easy to substitute stages or add new exporters

### 2. Configuration Factory Pattern
```go
config := NewComputeConfig(64, 64)
config.With1T1R()
config.WithWeights(weights)
design, _ := compiler.GenerateDesign(config)
```

### 3. Format Pluggability
All exporters follow consistent interface:
```go
// Export formats accept ArrayDesign + optional config
export.ExportJSON(design, "output.json")
export.ExportVerilog(design, "output.v")
export.ExportDEF(design, "output.def")
// ... etc
```

### 4. Mode-Based Behavior
ArrayDesign behavior adapts to Mode:
- **Storage**: Optimizes retention/endurance, ignores weights
- **Memory**: Optimizes speed/power, ignores weights
- **Compute**: Maps weights, calculates quantization stats

### 5. Architecture Abstraction
Cell instantiation changes based on `config.Architecture`:
- **Passive** (ArchPassive): Simple resistor, WL + BL
- **1T1R** (Arch1T1R): Transistor selector, WL + BL + SL
- **2T1R** (Arch2T1R): Dual transistors, WL + BL + SL + CSL (future)

---

## Typical Workflows

### Workflow 1: Design Storage Array (No Weights)
```go
config := compiler.NewStorageConfig(256, 256)
config.Technology = compiler.TechSKY130
config.StorageConfig.ErrorCorrection = "BCH"

design, _ := compiler.GenerateDesign(config)

// Export for analysis
export.ExportJSON(design, "storage_design.json")
export.ExportVerilog(design, "storage.v")
export.ExportDEF(design, "storage.def")

// Validate before OpenLane
validation.ValidateVerilog("storage.v")
```

### Workflow 2: Design Compute Array with Weights
```go
// Load pre-trained weights from MNIST model
weights := loadNNWeights("model.pt")

config := compiler.NewComputeConfig(128, 128)
config.ComputeConfig.InitialWeights = weights

design, _ := compiler.GenerateDesign(config)

// Export with statistics
export.ExportJSON(design, "compute_design.json")
export.ExportVerilog(design, "compute.v")
export.ExportDEF(design, "compute.def")

// Debug quantization
fmt.Printf("Quantization Error: %.4f dB PSNR\n", design.Stats.QuantPSNR)
```

### Workflow 3: Full OpenLane Integration
```go
// Generate files
config := compiler.NewComputeConfig(64, 64)
design, _ := compiler.GenerateDesign(config)

export.ExportVerilog(design, "output/crossbar.v")
export.ExportDEF(design, "output/crossbar.def")
export.ExportLEF(design, "output/crossbar.lef")
export.ExportLiberty(design, "output/crossbar.lib")

// Validate designs
validation.ValidateVerilog("output/crossbar.v")
validation.CrossCheckFiles("output/crossbar.lef", "output/crossbar.lib", "output/crossbar.v")

// Prepare OpenLane config
manager := openlane.NewManager()
config := openlane.DefaultConfig()
runner := openlane.NewRunner(manager, config)

// Run synthesis and place-and-route
result, _ := runner.RunYosys("read_verilog crossbar.v; hierarchy -check", "output")
// OpenLane P&R generates GDSII...
```

---

## Technology-Specific Details

### SkyWater 130nm (SKY130)
- **Cell Pitch**: 0.46 μm (unithd site width × 2)
- **Row Height**: 2.72 μm (standard cell)
- **Metal Layers**: 6 (M1-M6)
- **Standard Cell Library**: sky130_fd_sc_hd (High Density)
- **PDK Setup**: `volare enable --pdk sky130 sky130A`

### GlobalFoundries 180nm (GF180MCU)
- **Cell Pitch**: 0.84 μm
- **Row Height**: 3.6 μm
- **Metal Layers**: 4 (M1-M4)
- **Standard Cell Library**: gf180mcu_fd_sc_mcu7t5v0

### IHP 130nm BiCMOS (IHP_SG13G2)
- **Cell Pitch**: 0.34 μm
- **Row Height**: 2.48 μm
- **Unique Feature**: SiGe bipolar transistors for analog circuits
- **Use Case**: High-speed, low-power analog peripherals (TIA, DAC)

---

## Key Design Decisions

### 1. Why 30 Conductance Levels?
Based on Dr. external research group's FeCIM device characterization (COSM 2025), the HfO₂-ZrO₂ superlattice reliably programs to 30 distinct conductance states, yielding ~4.9 bits/cell. This is the balance point between write time and resolution.

### 2. Quantization Strategy
**Linear quantization in conductance space** (not weight space) because:
- Resistor networks are linear in conductance
- Weight quantization must respect device physics (G_min to G_max)
- Mapping: weight → normalized [0,1] → conductance level → G value

### 3. DEF FIXED Keyword
OpenLane's PLACEMENT_CURRENT_DEF + SYNTH_ELABORATE_ONLY prevents standard-cell synthesis from moving hand-placed FeCIM cells, preserving the designed array topology.

### 4. Passive vs 1T1R Trade-off
- **Passive**: Minimal area, susceptible to sneak paths, used for analysis
- **1T1R**: Area penalty (2× pitch), eliminates sneak paths, production-ready

### 5. Placeholder Timing Values
Liberty files contain estimated timing (0.1 ns rise time). Real values require:
1. SPICE compact model of FeFET (e.g., based on Song et al. 2024)
2. Characterization runs across PVT corners
3. Liberty file generation (e.g., with Liberate tool)

---

## Integration Points

### With Module 2 (Crossbar Arrays)
- Module 2 provides simulation models and MVM calculations
- Module 6 takes those validated designs into fabrication-ready format

### With Module 3 (MNIST)
- MNIST weights are exported to Compute ArrayConfig
- Quantization stats help assess network accuracy loss

### With Module 4 (Peripherals)
- DAC/ADC/TIA specifications set DACBits, ADCBits, TIAGain
- Liberty timing includes peripheral cell models

### With Main Application
- Module 6 GUI tabs integrate into unified Fyne application
- Shared theme, logging, and widget libraries

---

## Testing Strategy

**Test Files**:
- `pkg/compiler/compiler_test.go`: Weight mapping, quantization MSE
- `pkg/compiler/compiler_extended_test.go`: Large arrays, edge cases
- `pkg/export/export_test.go`: Format generation correctness
- `pkg/export/verilog_test.go`: Module hierarchy, cell instantiation
- `pkg/export/def_test.go`: Coordinate calculations
- `pkg/export/svg_test.go`: Visualization rendering
- `pkg/export/lattice_generator_test.go`: Batch generation

**Test Coverage**:
- All operation modes (Storage, Memory, Compute)
- All architectures (Passive, 1T1R, 2T1R)
- Weight quantization accuracy
- Format output validity
- Edge cases (empty arrays, mismatched dimensions)

**Running Tests**:
```bash
go test ./module6-eda/pkg/...
go test -v ./module6-eda/pkg/compiler/  # Verbose compiler tests
```

---

## Performance Characteristics

| Scenario | Compilation Time | Export Time | Validation Time |
|----------|-----------------|-------------|-----------------|
| 64×64 array, no weights | <1 ms | <10 ms | ~1 s (Yosys) |
| 256×256 array, no weights | <5 ms | ~50 ms | ~3 s (Yosys) |
| 64×64 compute with weights | ~10 ms | ~15 ms | ~2 s (validation) |
| 512×512 array | ~50 ms | ~500 ms | ~10 s (full validation) |

**Bottleneck**: Tool execution (Yosys, OpenROAD) not code compilation.

---

## Future Enhancements

1. **SPICE Library Generation**: Automated characterization to replace placeholder timing values
2. **Automatic Tile Generation**: For arrays > 1K×1K, decompose into smaller OpenLane tiles
3. **Power Estimation**: Per-cell leakage based on conductance level
4. **Thermal Analysis**: Integration with OpenROAD thermal solver
5. **Alternative Cell Architectures**: ReRAM, eMRAM backends
6. **GDS Export**: Direct GDSII generation without relying on OpenLane
7. **Parasitic Extraction**: RLC models for post-placement simulation
8. **Machine Learning Accelerator**: Auto-layout optimization for NN inference arrays

---

## References

- **OpenLane Flow**: https://openlane.readthedocs.io/
- **SkyWater PDK**: https://skywater-pdk.readthedocs.io/
- **Dr. external research group, FeCIM Fundamentals**: COSM 2025 transcript (docs/video-transcripts/)
- **Verilog/LEF/DEF Standards**: IEEE 1364, LEF/DEF 5.8 spec
- **Liberty Format**: Open Access Tutorial

---

## Glossary

- **ArrayConfig**: High-level specification (size, mode, technology)
- **ArrayDesign**: Low-level physical design (cells, coordinates, conductances)
- **Crossbar**: NxM grid of FeCIM devices with WL/BL select
- **DEF**: Design Exchange Format (placement constraints)
- **FeCIM**: Ferroelectric Compute-in-Memory (HfO₂-ZrO₂ superlattice device)
- **LEF**: Library Exchange Format (cell abstractions)
- **Level**: Programmed conductance state (0-29 for 30-level device)
- **OpenLane**: Automated RTL-to-GDSII flow (uses Yosys, OpenROAD, KLayout)
- **Quantization**: Mapping continuous weights to discrete levels
- **Sneak Path**: Unwanted current leakage through unselected cells in passive crossbar
- **1T1R**: One Transistor One Resistor (mitigates sneak paths)

---

**Last Updated**: January 2026
**Maintained By**: FeCIM Lattice Tools Team
**License**: Apache 2.0 (See LICENSE file)
