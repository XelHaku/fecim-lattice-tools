# Module 6 EDA API Reference

Complete API documentation for the FeCIM design tool suite. This module transforms array configurations into multi-format design files for fabrication and simulation.

## Table of Contents

- [Compiler Package](#compiler-package)
- [Export Package](#export-package)
- [Validation Package](#validation-package)
- [OpenLane Package](#openlane-package)
- [Common Types](#common-types)

---

## Compiler Package

The compiler transforms array configurations into physical designs with complete cell-to-conductance mapping.

### GenerateDesign

Creates a complete array design from configuration and optional weights.

```go
func GenerateDesign(config *ArrayConfig) (*ArrayDesign, error)
```

**Behavior:**
- If `config.Mode == ModeCompute` and weights are provided, performs weight quantization
- Otherwise, generates blank initialized array
- Returns error only if weight matrix exceeds array dimensions

**Example:**

```go
// Compute mode with weights (neural network inference)
config := compiler.NewComputeConfig(64, 64)
weights := [][]float64{
    {0.1, -0.2, 0.3},
    {-0.4, 0.5, -0.6},
}
config.ComputeConfig.InitialWeights = weights
design, err := compiler.GenerateDesign(config)
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Design: %d cells, PSNR=%.2f dB\n",
    design.Stats.TotalCells, design.Stats.QuantPSNR)
```

### GenerateBlank

Creates an initialized array without weights.

```go
func GenerateBlank(config *ArrayConfig) *ArrayDesign
```

**Behavior:**
- Always succeeds (no error possible)
- Initializes all cells to Level 0 (minimum conductance)
- Suitable for Storage and Memory modes

**Example:**

```go
// Storage mode (NAND replacement)
config := compiler.NewStorageConfig(256, 256)
design := compiler.GenerateBlank(config)
fmt.Printf("Created storage chip: %.2f mm²\n", design.Stats.AreaMM2)
```

### NewArrayConfig / NewStorageConfig / NewMemoryConfig / NewComputeConfig

Factory functions create mode-specific configurations with sensible defaults.

```go
func NewArrayConfig(mode OperationMode, rows, cols int) *ArrayConfig
func NewStorageConfig(rows, cols int) *ArrayConfig
func NewMemoryConfig(rows, cols int) *ArrayConfig
func NewComputeConfig(rows, cols int) *ArrayConfig
```

**Builder Pattern Methods:**

```go
// Switch to 1T1R architecture (sneak path mitigation)
config.With1T1R()

// Switch to 2T1R architecture (individual cell addressing)
config.With2T1R()

// Set weights for compute mode
config.WithWeights(weights)
```

**Example:**

```go
// 256x256 storage array with 1T1R
config := compiler.NewStorageConfig(256, 256).
    With1T1R()

// 128x128 memory array
config := compiler.NewMemoryConfig(128, 128)

// 64x64 compute with pre-trained weights
config := compiler.NewComputeConfig(64, 64).
    WithWeights(mnistWeights)
```

---

## ArrayConfig Type

Complete configuration for array design.

### Fields

```go
type ArrayConfig struct {
    // Array dimensions
    Name      string        // Design name (default: "fecim_crossbar")
    Mode      OperationMode // Storage, Memory, or Compute
    ArrayRows int           // Number of rows
    ArrayCols int           // Number of columns

    // Technology
    Technology   string // "SKY130", "GF180MCU", "IHP_SG13G2"
    Architecture string // "passive", "1t1r", "2t1r"

    // Physical parameters
    CellPitch float64 // Cell width in microns (default: 0.46)
    RowHeight float64 // Cell height in microns (default: 2.72)
    Levels    int     // Conductance levels (default: 30)

    // Electrical parameters
    GMin     float64 // Min conductance (μS, default: 10.0)
    GMax     float64 // Max conductance (μS, default: 100.0)
    VProgMin float64 // Min programming voltage (V, default: 2.0)
    VProgMax float64 // Max programming voltage (V, default: 5.0)
    TPulse   float64 // Programming pulse width (ns, default: 50.0)

    // Peripheral configuration
    Peripherals PeripheralConfig

    // Mode-specific configurations
    StorageConfig *StorageArrayConfig
    MemoryConfig  *MemoryArrayConfig
    ComputeConfig *ComputeArrayConfig
}
```

### OperationMode Constants

```go
const (
    ModeStorage  // NAND replacement: high-density, data retention
    ModeMemory   // DRAM replacement: high-speed, zero-refresh
    ModeCompute  // AI accelerator: analog compute-in-memory
)
```

### Architecture Constants

```go
const (
    ArchPassive = "passive" // WL[], BL[] only (sneak-path susceptible)
    Arch1T1R    = "1t1r"    // WL[], BL[], SL[] (sneak-path mitigated)
    Arch2T1R    = "2t1r"    // WL[], BL[], SL[], CSL[] (full control)
)
```

### Technology Constants

```go
const (
    TechSKY130 = "SKY130"       // 130nm (open-source)
    TechGF180  = "GF180MCU"     // 180nm (open-source)
    TechIHP    = "IHP_SG13G2"   // 130nm BiCMOS
)
```

---

## ArrayDesign Type

Complete output of design compilation.

```go
type ArrayDesign struct {
    Config *ArrayConfig     // Input configuration
    Cells  []CellAssignment // Physical cell placements
    Stats  DesignStats      // Design metrics
}
```

### CellAssignment Type

```go
type CellAssignment struct {
    Row         int     // Row index
    Col         int     // Column index
    Level       int     // Programmed conductance level (0 to Levels-1)
    Conductance float64 // Conductance in microsiemens
    Resistance  float64 // Resistance in ohms (1e6/Conductance)
    ProgramV    float64 // Programming voltage in volts

    // Compute mode only:
    InitialWeight float64 // Original weight value before quantization
}
```

### DesignStats Type

```go
type DesignStats struct {
    TotalCells     int     // Physical array size
    ActiveCells    int     // Programmed cells (compute mode)
    AreaMM2        float64 // Total die area
    PowerMW        float64 // Estimated power consumption

    // Compute mode with weights:
    QuantMSE  float64 // Quantization mean squared error
    QuantPSNR float64 // Quantization PSNR in dB
    WeightMin float64 // Original weight range
    WeightMax float64 // Original weight range
}
```

---

## Export Package

Multi-format output generation for design files and netlists.

### ExportJSON

Writes complete design to JSON file (for version control and analysis).

```go
func ExportJSON(design *ArrayDesign, path string) error
```

**Output Structure:**
```json
{
  "config": { /* full ArrayConfig */ },
  "cells": [ /* all CellAssignments */ ],
  "stats": { /* DesignStats */ }
}
```

**Example:**

```go
design, _ := compiler.GenerateDesign(config)
err := export.ExportJSON(design, "output/design.json")
```

### ExportCSV

Writes cell data to CSV file (for spreadsheet tools).

```go
func ExportCSV(design *ArrayDesign, path string) error
```

**Output (Compute Mode with Weights):**
```
row,col,weight,level,conductance_uS,resistance_ohm,program_V
0,0,0.100000,15,50.50,19802.00,3.500000
0,1,-0.200000,9,30.90,32361.00,3.000000
```

**Output (Storage/Memory Mode):**
```
row,col,level,conductance_uS,resistance_ohm,program_V
0,0,0,1.00,1000000.00,2.000000
```

### ExportVerilog

Generates structural Verilog netlist for simulation and synthesis.

```go
func ExportVerilog(design *ArrayDesign, path string) error
func ExportVerilogWithConfig(design *ArrayDesign, path string, config VerilogConfig) error
```

**Verilog Features:**
- Module ports: `WL[]` (word lines), `BL[]` (bit lines), optionally `SL[]` (source lines)
- Cell instances: `fecim_bit` or `fecim_1t1r` with LEVEL parameter
- Comments with weight/conductance values (compute mode)
- Compatible with Yosys synthesis and iverilog simulation

**VerilogConfig Type:**

```go
type VerilogConfig struct {
    ModuleName string // Default: "fecim_crossbar"
    CellName   string // Default: "fecim_bit"
}
```

**Example:**

```go
design, _ := compiler.GenerateDesign(config)

// Default configuration
export.ExportVerilog(design, "output/crossbar.v")

// Custom configuration
cfg := export.VerilogConfig{
    ModuleName: "ai_accelerator",
    CellName:   "fecim_1t1r",
}
export.ExportVerilogWithConfig(design, "output/crossbar.v", cfg)
```

**Example Output (passive, 2x2):**
```verilog
module fecim_crossbar (
    input  wire [1:0] WL,  // Word Lines (row select)
    inout  wire [1:0] BL,  // Bit Lines (column data)
    inout  wire       VPWR, // Power supply
    inout  wire       VGND  // Ground
);

    parameter ROWS = 2;
    parameter COLS = 2;
    parameter LEVELS = 30;
    parameter MODE = "Compute";
    parameter ARCHITECTURE = "passive";

    // Cell [0,0]: weight=0.1000, level=15, G=50.50 uS
    fecim_bit #(.LEVEL(15)) R_0_0 (
        .WL  (WL[0]),
        .BL  (BL[0]),
        .VPWR (VPWR),
        .VGND (VGND)
    );

    // ... more cells ...
endmodule
```

### ExportDEF

Generates DEF (Design Exchange Format) placement file for OpenLane.

```go
func ExportDEF(design *ArrayDesign, path string) error
func ExportDEFWithConfig(design *ArrayDesign, path string, config DEFConfig) error
```

**DEFConfig Type:**

```go
type DEFConfig struct {
    DesignName   string  // Default: "fecim_crossbar"
    CellName     string  // Default: "fecim_bit"
    CellWidth    float64 // Default: 0.46 (SKY130)
    CellHeight   float64 // Default: 2.72 (SKY130)
    OriginX      float64 // Array origin X in microns (default: 10.0)
    OriginY      float64 // Array origin Y in microns (default: 10.0)
    MarginX      float64 // Die margin X (default: 20.0)
    MarginY      float64 // Die margin Y (default: 20.0)
    DatabaseUnit int     // Units per micron (default: 1000)
}
```

**DEF Features:**
- FIXED placement (locked for OpenLane injection)
- ROW definitions for placement validation
- PIN definitions for all port types
- NET connectivity (WL[], BL[], SL[], CSL[], VPWR, VGND)
- Compatible with OpenROAD/OpenLane

**Example:**

```go
design, _ := compiler.GenerateDesign(config)

// Default configuration (SKY130)
export.ExportDEF(design, "output/crossbar.def")

// Custom technology node
defCfg := export.DefaultDEFConfig()
defCfg.CellWidth = 0.64    // GF180MCU pitch
defCfg.CellHeight = 4.07   // GF180MCU row height
export.ExportDEFWithConfig(design, "output/crossbar.def", defCfg)
```

**OpenLane Integration:**

```json
{
    "VERILOG_FILES": "dir::crossbar.v",
    "PLACEMENT_CURRENT_DEF": "dir::crossbar.def",
    "SYNTH_ELABORATE_ONLY": 1,
    "PL_SKIP_INITIAL_PLACEMENT": 1
}
```

### ExportSPICE

Generates SPICE netlist for analog simulation.

```go
func ExportSPICE(design *ArrayDesign, path string, vdd float64) error
func GenerateSPICE(design *ArrayDesign, vdd float64) string
```

**SPICE Features:**
- Resistor network representing conductance values
- Word line drivers, bit line loads
- ngspice/HSPICE compatible
- Simplified FeFET model (resistance only)

**Example:**

```go
design, _ := compiler.GenerateDesign(config)

// Export with 1.8V supply (SKY130)
export.ExportSPICE(design, "output/crossbar.sp", 1.8)

// Get SPICE text for programmatic processing
spiceText := export.GenerateSPICE(design, 1.8)
```

### LEF and Liberty Generation

Generates cell abstraction files for place-and-route tools.

```go
func GenerateLEF(cfg config.CellConfig) string
func GenerateLiberty(cfg config.CellConfig) string
```

**Note:** These generate placeholder timing values requiring characterization via SPICE.

---

## Validation Package

Validates design files and extracts statistics.

### ValidateDEF

Performs basic DEF syntax validation.

```go
func ValidateDEF(defPath string) error
```

**Validation Checks:**
- Required keywords present (VERSION, DESIGN, UNITS, DIEAREA, COMPONENTS)
- COMPONENTS count matches actual instances
- END keywords properly matched

**Example:**

```go
err := validation.ValidateDEF("output/crossbar.def")
if err != nil {
    log.Fatalf("DEF validation failed: %v", err)
}
```

### GetDEFStats

Extracts design statistics from DEF file.

```go
func GetDEFStats(defPath string) (map[string]interface{}, error)
```

**Returned Fields:**
- `design_name`: Design name from DESIGN keyword
- `component_count`: Number of COMPONENTS
- `die_area`: DIEAREA specification

**Example:**

```go
stats, err := validation.GetDEFStats("output/crossbar.def")
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Design: %s, Components: %d\n",
    stats["design_name"], stats["component_count"])
```

---

## OpenLane Package

Configuration and execution of OpenLane EDA tool.

### LoadConfig / SaveConfig

Loads or saves OpenLane configuration from JSON.

```go
func LoadConfig(path string) (*Config, error)
func SaveConfig(cfg *Config, path string) error
```

**Config Type:**

```go
type Config struct {
    PDKRoot          string        // PDK installation path
    PDKVariant       string        // Default: "sky130A"
    SCLibrary        string        // Default: "sky130_fd_sc_hd"
    PreferredMode    Mode          // ModeDocker or ModeNative
    TimeoutPlacement time.Duration // Default: 5min
    TimeoutSynthesis time.Duration // Default: 10min
    TimeoutRouting   time.Duration // Default: 15min
    DockerImage      string        // Default: OpenLane latest
}
```

**Example:**

```go
// Load default config (uses ~/.volare or PDK_ROOT env var)
cfg, err := openlane.LoadConfig(openlane.GetConfigPath())
if err != nil {
    cfg = openlane.DefaultConfig() // Fallback to defaults
}

// Modify for GF180MCU
cfg.PDKVariant = "gf180mcuC"
cfg.SCLibrary = "gf180mcu_fd_sc_mcu7t5v0"

// Save to user config
openlane.SaveConfig(cfg, openlane.GetConfigPath())
```

### Runner

Executes OpenLane tools in Docker or native mode.

```go
type Runner struct { /* ... */ }

func NewRunner(manager *Manager, config *Config) *Runner
func (r *Runner) RunOpenROAD(scriptName, workDir string, envVars map[string]string) (*Result, error)
func (r *Runner) RunYosys(yosysCmd, workDir string) (*Result, error)
```

**Result Type:**

```go
type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Duration time.Duration
}
```

**Example:**

```go
manager := openlane.NewManager()
runner := openlane.NewRunner(manager, cfg)

// Run Yosys synthesis
result, err := runner.RunYosys(
    "read_verilog crossbar.v; hierarchy -check; write_json crossbar.json",
    "output/")
if err != nil {
    log.Printf("Yosys failed: %v\n%s", err, result.Stderr)
}

// Run OpenROAD from TCL script
env := map[string]string{
    "CELL_LEF": "output/fecim_bit.lef",
    "DEF_FILE": "output/crossbar.def",
}
result, err := runner.RunOpenROAD("place.tcl", "output/", env)
if result.ExitCode != 0 {
    log.Printf("Placement failed: %s", result.Stderr)
}
```

---

## Common Types

### PeripheralConfig

Configuration for DAC, ADC, and TIA peripherals.

```go
type PeripheralConfig struct {
    DACBits   int     // Resolution in bits (default: 8)
    ADCBits   int     // Resolution in bits (default: 8)
    TIAGain   float64 // Transimpedance gain in ohms (default: 10kΩ)
    VDD       float64 // Supply voltage (default: 1.8V)
    ClockFreq float64 // Operating frequency in MHz (default: 100MHz)
}
```

### StorageArrayConfig

Storage mode specific parameters.

```go
type StorageArrayConfig struct {
    RetentionYears  float64 // Data retention target (default: 10)
    EnduranceCycles int     // Write endurance (default: 1e6)
    ErrorCorrection string  // "none", "SECDED", "BCH" (default: "SECDED")
}
```

### MemoryArrayConfig

Memory mode specific parameters.

```go
type MemoryArrayConfig struct {
    AccessTimeNs  float64 // Read access time target in ns (default: 10)
    WriteTimeNs   float64 // Write time target in ns (default: 50)
    BandwidthGBps float64 // Bandwidth target in GB/s (default: 10)
    PowerBudgetMW float64 // Power budget in mW (default: 100)
}
```

### ComputeArrayConfig

Compute mode specific parameters.

```go
type ComputeArrayConfig struct {
    InitialWeights  [][]float64 // Optional pre-trained weights
    QuantLevels     int         // Quantization levels (default: 30)
    AccumulatorBits int         // MAC accumulator width (default: 16)
    ActivationFunc  string      // "none", "relu", "sigmoid" (default: "none")
}
```

---

## Complete Usage Example

End-to-end workflow from design to OpenLane integration.

```go
package main

import (
    "log"
    "fecim-lattice-tools/module6-eda/pkg/compiler"
    "fecim-lattice-tools/module6-eda/pkg/export"
)

func main() {
    // Step 1: Create compute array configuration
    config := compiler.NewComputeConfig(64, 64).
        With1T1R().
        WithWeights(mnistWeights)  // From trained model

    // Step 2: Generate design with weight mapping
    design, err := compiler.GenerateDesign(config)
    if err != nil {
        log.Fatalf("Compilation failed: %v", err)
    }
    log.Printf("Generated %d cells (%.2f mm²), PSNR=%.1f dB",
        design.Stats.TotalCells, design.Stats.AreaMM2, design.Stats.QuantPSNR)

    // Step 3: Export to multiple formats
    export.ExportJSON(design, "output/design.json")
    export.ExportCSV(design, "output/cells.csv")
    export.ExportVerilog(design, "output/crossbar.v")
    export.ExportDEF(design, "output/crossbar.def")
    export.ExportSPICE(design, "output/crossbar.sp", 1.8)

    // Step 4: Validate outputs
    if err := validation.ValidateDEF("output/crossbar.def"); err != nil {
        log.Fatalf("DEF validation failed: %v", err)
    }
    stats, _ := validation.GetDEFStats("output/crossbar.def")
    log.Printf("DEF: %s, %d components", stats["design_name"], stats["component_count"])

    // Step 5: Integrate with OpenLane
    cfg := openlane.DefaultConfig()
    manager := openlane.NewManager()
    runner := openlane.NewRunner(manager, cfg)

    result, err := runner.RunYosys(
        "read_verilog output/crossbar.v; hierarchy -check",
        "output/")
    if err != nil {
        log.Printf("Yosys error: %v\n%s", err, result.Stderr)
    }
}
```

---

## Design Principles

1. **Multi-Mode Support**: One API handles Storage, Memory, and Compute modes with appropriate defaults
2. **Zero-Copy Export**: Generate formats directly without intermediate representations
3. **Architecture Agnostic**: Passive, 1T1R, and 2T1R architectures handled transparently
4. **Technology Portable**: SKY130, GF180MCU, and IHP support through configuration
5. **OpenLane Ready**: DEF and Verilog files integrate directly with OpenLane flow

---

## Error Handling

All functions follow Go conventions:
- Functions that can fail return `(result, error)`
- Functions that cannot fail return result directly
- Errors are descriptive and include context

```go
// Can fail (weight exceeds array)
design, err := compiler.GenerateDesign(config)
if err != nil {
    log.Fatalf("Design generation failed: %v", err)
}

// Cannot fail (no constraints violated)
design := compiler.GenerateBlank(config)
```

---

## Performance Notes

- **Array Compilation**: O(rows × cols) for weight mapping
- **Export**: O(cells) for CSV/JSON, O(cells) for DEF/Verilog
- **No memory leaks**: All allocations pre-sized or explicitly managed
- **Thread-safe**: Separate design instances for concurrent operations

---

## References

- FeCIM Technology: Dr. external research group, COSM 2025
- LEF/DEF 5.8: Si2/OpenAccess Coalition
- Liberty Timing: Synopsys standard format
- OpenLane: https://openlane.readthedocs.io/
