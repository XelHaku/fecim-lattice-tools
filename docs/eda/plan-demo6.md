# Demo 6: FeCIM Design Suite (Preview)
## Complete Implementation Plan

**Date:** 2026-01-21  
**Author:** @XelHaku  
**Status:** EDA in Diapers → Real EDA Bridge

---

## 🚀 QUICK START: Implementation Checklist
**START HERE. This section tells you exactly what to build, in what order, with what tests.**

### Prerequisites (30 minutes)
```bash
# 1. Install Go 1.21+
go version  # Should be 1.21 or higher

# 2. Install Fyne dependencies (Linux)
sudo apt-get install -y gcc libgl1-mesa-dev xorg-dev

# 3. Create project
mkdir -p module6-eda && cd module6-eda
go mod init module6-eda

# 4. Install dependencies
go get fyne.io/fyne/v2@latest
go get github.com/go-echarts/go-echarts/v2@latest

# 5. (Optional) Install external tools for full integration
sudo apt-get install -y ngspice  # For Tab 4 simulation
# KLayout: https://www.klayout.de/build.html
```

### MVP Definition (Build This First!)

**MINIMUM VIABLE PRODUCT (Week 1-2):**

*   ✅ **Tab 1: Compiler** (CORE - must work)
    *   └── Compile weights → JSON output
*   ✅ **Tab 5: Export** (CORE - must work)
    *   └── Export JSON + CSV + SPICE netlist
*   ⚠️ **Tab 2: Layout** (VISUAL ONLY)
    *   └── Show crossbar grid, no GDSII export yet
*   ⏸️ **Tab 3: Explorer** (DEFER)
*   ⏸️ **Tab 4: Simulate** (DEFER - requires ngspice)
*   ⏸️ **Tab 6: Learn** (DEFER - static content)

**MVP = Tab 1 + Tab 5 + simple Tab 2 visualization**

### Implementation Order (Copy-Paste Ready)

#### PHASE 1: Core Data Structures (Day 1)
*   [ ] Create `pkg/compiler/types.go`
    *   `CompileConfig` struct
    *   `CrossbarMapping` struct
    *   `CellAssignment` struct
    *   `CompilationStats` struct
*   [ ] Test: `go build ./pkg/compiler`

#### PHASE 2: Compiler Logic (Day 2-3)
*   [ ] Create `pkg/compiler/compiler.go`
    *   `NewDefaultConfig()`
    *   `Compile(weights, config) → mapping`
    *   `quantizeSymmetric()`
    *   `levelToConductance()`
*   [ ] Create `pkg/compiler/compiler_test.go`
    *   `TestCompile_SmallMatrix`
    *   `TestCompile_Quantization`
    *   `TestCompile_Statistics`
*   [ ] Test: `go test ./pkg/compiler -v`

#### PHASE 3: Export Functions (Day 4-5)
*   [ ] Create `pkg/export/json.go`
    *   `ExportJSON(mapping, path)`
*   [ ] Create `pkg/export/csv.go`
    *   `ExportCSV(mapping, path)`
*   [ ] Create `pkg/export/spice.go`
    *   `GenerateSPICE(mapping, config) → string`
    *   `ExportSPICE(mapping, path)`
*   [ ] Test: `go test ./pkg/export -v`
*   [ ] Verify: Open output files, check format

#### PHASE 4: Basic GUI Shell (Day 6-7)
*   [ ] Create `cmd/eda-gui/main.go`
    *   Fyne app with 6 tabs (empty)
*   [ ] Create `pkg/gui/app.go`
    *   `CreateMainWindow()`
    *   `CreateTabs()`
*   [ ] Test: `go run ./cmd/eda-gui`
*   [ ] Verify: Window opens with 6 tab buttons

#### PHASE 5: Tab 1 GUI (Day 8-9)
*   [ ] Create `pkg/gui/tabs/compiler_tab.go`
    *   File picker for weights JSON
    *   Config inputs (rows, cols, levels)
    *   Compile button
    *   Results display
*   [ ] Test: Load sample weights, compile, see stats

#### PHASE 6: Tab 5 GUI (Day 10)
*   [ ] Create `pkg/gui/tabs/export_tab.go`
    *   Checkboxes for export formats
    *   Directory picker
    *   Export button
    *   File list after export
*   [ ] Test: Export all formats, verify files

#### PHASE 7: Tab 2 Visualization (Day 11-12)
*   [ ] Create `pkg/gui/tabs/layout_tab.go`
    *   Canvas showing crossbar grid
    *   Color by conductance level
    *   Cell click → show details
*   [ ] Test: Visual inspection

**MVP COMPLETE! 🎉**

---

### Sample Test Data

Create `data/sample_weights_8x8.json`:

```json
{
  "name": "test_8x8",
  "rows": 8,
  "cols": 8,
  "weights": [
    [0.1, -0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8],
    [-0.1, 0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8],
    [0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75, -0.85],
    [-0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75, 0.85],
    [0.05, -0.15, 0.25, -0.35, 0.45, -0.55, 0.65, -0.75],
    [-0.05, 0.15, -0.25, 0.35, -0.45, 0.55, -0.65, 0.75],
    [0.2, -0.3, 0.4, -0.5, 0.6, -0.7, 0.8, -0.9],
    [-0.2, 0.3, -0.4, 0.5, -0.6, 0.7, -0.8, 0.9]
  ]
}
```

### Simplified File Structure (MVP)

```text
module6-eda/
├── cmd/
│   └── eda-gui/
│       └── main.go           # 50 lines - app entry
│
├── pkg/
│   ├── compiler/
│   │   ├── types.go          # 80 lines - structs only
│   │   ├── compiler.go       # 150 lines - compile logic
│   │   └── compiler_test.go  # 100 lines - tests
│   │
│   ├── export/
│   │   ├── json.go           # 30 lines
│   │   ├── csv.go            # 40 lines
│   │   ├── spice.go          # 100 lines
│   │   └── export_test.go    # 60 lines
│   │
│   └── gui/
│       ├── app.go            # 80 lines - main window
│       └── tabs/
│           ├── compiler_tab.go  # 150 lines
│           ├── layout_tab.go    # 100 lines (visual only)
│           └── export_tab.go    # 100 lines
│
├── data/
│   └── sample_weights_8x8.json
│
├── go.mod
└── go.sum
```

---

## 📋 COPY-PASTE STARTER TEMPLATES

### Template 1: types.go (Start Here)
`pkg/compiler/types.go`
```go
package compiler

// CompileConfig holds all parameters for compilation
type CompileConfig struct {
	ArrayRows int     `json:"array_rows"` // Target array rows
	ArrayCols int     `json:"array_cols"` // Target array cols
	Levels    int     `json:"levels"`     // Quantization levels (2-30)
	GMin      float64 `json:"g_min"`      // Min conductance (μS)
	GMax      float64 `json:"g_max"`      // Max conductance (μS)
	VProgMin  float64 `json:"v_prog_min"` // Min programming voltage (V)
	VProgMax  float64 `json:"v_prog_max"` // Max programming voltage (V)
	TPulse    float64 `json:"t_pulse"`    // Pulse width (ns)
}

// DefaultConfig returns standard FeCIM parameters
func DefaultConfig() CompileConfig {
	return CompileConfig{
		ArrayRows: 128,
		ArrayCols: 128,
		Levels:    30,
		GMin:      1.0,
		GMax:      100.0,
		VProgMin:  2.0,
		VProgMax:  5.0,
		TPulse:    50.0,
	}
}

// CellAssignment represents one programmed FeFET cell
type CellAssignment struct {
	Row         int     `json:"row"`
	Col         int     `json:"col"`
	WeightValue float64 `json:"weight_value"` // Original weight
	QuantLevel  int     `json:"quant_level"`  // 0 to Levels-1
	Conductance float64 `json:"conductance"`  // μS
	ProgramV    float64 `json:"program_v"`    // V
}

// CrossbarMapping is the complete compilation output
type CrossbarMapping struct {
	Config CompileConfig    `json:"config"`
	Cells  []CellAssignment `json:"cells"` // Flat array for simplicity
	Stats  Stats            `json:"stats"`
}

// Stats holds compilation statistics
type Stats struct {
	TotalCells   int     `json:"total_cells"`
	UsedCells    int     `json:"used_cells"`
	Utilization  float64 `json:"utilization"`   // 0.0 to 1.0
	WeightMin    float64 `json:"weight_min"`
	WeightMax    float64 `json:"weight_max"`
	QuantMSE     float64 `json:"quant_mse"`     // Mean squared error
	QuantPSNR    float64 `json:"quant_psnr_db"` // dB
	UniqueLevels int     `json:"unique_levels"`
}
```

### Template 2: compiler.go (Core Logic)
`pkg/compiler/compiler.go`
```go
package compiler

import (
	"fmt"
	"math"
)

// Compile transforms a weight matrix into crossbar cell assignments
func Compile(weights [][]float64, config CompileConfig) (*CrossbarMapping, error) {
	// Validate
	if len(weights) == 0 || len(weights[0]) == 0 {
		return nil, fmt.Errorf("empty weight matrix")
	}

	rows := len(weights)
	cols := len(weights[0])

	if rows > config.ArrayRows || cols > config.ArrayCols {
		return nil, fmt.Errorf("weights %dx%d exceed array %dx%d",
			rows, cols, config.ArrayRows, config.ArrayCols)
	}

	// Find weight range for symmetric quantization
	wMin, wMax := weights[0][0], weights[0][0]
	for i := range weights {
		for j := range weights[i] {
			if weights[i][j] < wMin {
				wMin = weights[i][j]
			}
			if weights[i][j] > wMax {
				wMax = weights[i][j]
			}
		}
	}
	wAbsMax := math.Max(math.Abs(wMin), math.Abs(wMax))

	// Compile each weight
	var cells []CellAssignment
	var mseSum float64
	levelsUsed := make(map[int]bool)

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			w := weights[i][j]

			// Quantize: map [-wAbsMax, +wAbsMax] to [0, Levels-1]
			normalized := (w + wAbsMax) / (2 * wAbsMax) // 0 to 1
			level := int(math.Round(normalized * float64(config.Levels-1)))
			level = clamp(level, 0, config.Levels-1)
			levelsUsed[level] = true

			// Dequantize to get quantized value
			qNorm := float64(level) / float64(config.Levels-1)
			qValue := -wAbsMax + qNorm*(2*wAbsMax)
			mseSum += (w - qValue) * (w - qValue)

			// Map to physical parameters
			gNorm := float64(level) / float64(config.Levels-1)
			conductance := config.GMin + gNorm*(config.GMax-config.GMin)
			progV := config.VProgMin + gNorm*(config.VProgMax-config.VProgMin)

			cells = append(cells, CellAssignment{
				Row:         i,
				Col:         j,
				WeightValue: w,
				QuantLevel:  level,
				Conductance: conductance,
				ProgramV:    progV,
			})
		}
	}

	// Calculate statistics
	numCells := rows * cols
	mse := mseSum / float64(numCells)
	psnr := 100.0
	if mse > 0 {
		psnr = 10 * math.Log10((wAbsMax*wAbsMax)/mse)
	}

	return &CrossbarMapping{
		Config: config,
		Cells:  cells,
		Stats: Stats{
			TotalCells:   config.ArrayRows * config.ArrayCols,
			UsedCells:    numCells,
			Utilization:  float64(numCells) / float64(config.ArrayRows*config.ArrayCols),
			WeightMin:    wMin,
			WeightMax:    wMax,
			QuantMSE:     mse,
			QuantPSNR:    psnr,
			UniqueLevels: len(levelsUsed),
		},
	}, nil
}

func clamp(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
```

### Template 3: compiler_test.go (Verify It Works)
`pkg/compiler/compiler_test.go`
```go
package compiler

import (
	"testing"
)

func TestCompile_Basic(t *testing.T) {
	weights := [][]float64{
		{0.1, -0.2, 0.3},
		{-0.4, 0.5, -0.6},
	}

	config := DefaultConfig()
	config.ArrayRows = 8
	config.ArrayCols = 8

	mapping, err := Compile(weights, config)
	if err != nil {
		t.Fatalf("Compile failed: %v", err)
	}

	// Check basic stats
	if mapping.Stats.UsedCells != 6 {
		t.Errorf("Expected 6 used cells, got %d", mapping.Stats.UsedCells)
	}

	if mapping.Stats.Utilization < 0 || mapping.Stats.Utilization > 1 {
		t.Errorf("Invalid utilization: %f", mapping.Stats.Utilization)
	}

	// Check all cells have valid conductance
	for _, cell := range mapping.Cells {
		if cell.Conductance < config.GMin || cell.Conductance > config.GMax {
			t.Errorf("Cell conductance %f outside range [%f, %f]",
				cell.Conductance, config.GMin, config.GMax)
		}
	}

	t.Logf("Stats: %+v", mapping.Stats)
}
```

### Template 4: main.go (App Entry)
`cmd/eda-gui/main.go`
```go
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main() {
	a := app.New()
	w := a.NewWindow("Demo 6: FeCIM Design Suite (Preview)")
	w.Resize(fyne.NewSize(1200, 800))

	// Create tabs
	tabs := container.NewAppTabs(
		container.NewTabItem("1. Compiler", makeCompilerTab()),
		container.NewTabItem("2. Layout", makePlaceholderTab("Layout visualization coming soon")),
		container.NewTabItem("3. Explorer", makePlaceholderTab("Design space explorer coming soon")),
		container.NewTabItem("4. Simulate", makePlaceholderTab("Simulation bridge coming soon")),
		container.NewTabItem("5. Export", makeExportTab()),
		container.NewTabItem("6. Learn", makePlaceholderTab("Learning resources coming soon")),
	)
	tabs.SetTabLocation(container.TabLocationTop)

	// Add preview banner
	banner := widget.NewLabel("⚠️ PREVIEW: Bridge to open-source EDA tools (ngspice, KLayout, CiMLoop)")
	banner.Alignment = fyne.TextAlignCenter

	content := container.NewBorder(banner, nil, nil, nil, tabs)
	w.SetContent(content)
	w.ShowAndRun()
}

func makePlaceholderTab(message string) fyne.CanvasObject {
	return container.NewCenter(widget.NewLabel(message))
}

func makeCompilerTab() fyne.CanvasObject {
	// TODO: Implement in pkg/gui/tabs/compiler_tab.go
	return widget.NewLabel("Compiler Tab - Load weights, configure, compile")
}

func makeExportTab() fyne.CanvasObject {
	// TODO: Implement in pkg/gui/tabs/export_tab.go
	return widget.NewLabel("Export Tab - JSON, CSV, SPICE, GDSII script")
}
```

### Template 5: spice.go (SPICE Export)
`pkg/export/spice.go`
```go
package export

import (
	"fmt"
	"os"
	"strings"

	"module6-eda/pkg/compiler"
)

// GenerateSPICE creates ngspice-compatible netlist
func GenerateSPICE(mapping *compiler.CrossbarMapping, vdd float64) string {
	var sb strings.Builder

	sb.WriteString("* FeCIM Crossbar SPICE Netlist\n")
	sb.WriteString("* Generated by Demo 6: FeCIM Design Suite\n")
	sb.WriteString(fmt.Sprintf("* Array: %d cells\n\n", len(mapping.Cells)))

	// Power supply
	sb.WriteString(fmt.Sprintf(".param VDD = %.2f\n", vdd))
	sb.WriteString("VDD vdd 0 DC {VDD}\n\n")

	// Find array dimensions
	maxRow, maxCol := 0, 0
	for _, cell := range mapping.Cells {
		if cell.Row > maxRow {
			maxRow = cell.Row
		}
		if cell.Col > maxCol {
			maxCol = cell.Col
		}
	}

	// Word line drivers (inputs)
	sb.WriteString("* Word Line Drivers\n")
	for row := 0; row <= maxRow; row++ {
		sb.WriteString(fmt.Sprintf("VWL%d wl%d 0 DC 0\n", row, row))
	}
	sb.WriteString("\n")

	// Bit line loads
	sb.WriteString("* Bit Line Loads (1k for sensing)\n")
	for col := 0; col <= maxCol; col++ {
		sb.WriteString(fmt.Sprintf("RBL%d bl%d 0 1k\n", col, col))
	}
	sb.WriteString("\n")

	// FeFET cells as resistors (simplified model)
	sb.WriteString("* FeFET Cells (resistor model: R = 1/G)\n")
	for _, cell := range mapping.Cells {
		resistance := 1e6 / cell.Conductance // Ω from μS
		sb.WriteString(fmt.Sprintf("R_%d_%d wl%d bl%d %.2f\n",
			cell.Row, cell.Col, cell.Row, cell.Col, resistance))
	}
	sb.WriteString("\n")

	// Analysis
	sb.WriteString("* Analysis\n")
	sb.WriteString(".op\n")
	sb.WriteString(".control\n")
	sb.WriteString("run\n")
	for col := 0; col <= maxCol; col++ {
		sb.WriteString(fmt.Sprintf("print i(RBL%d)\n", col))
	}
	sb.WriteString(".endc\n")
	sb.WriteString(".end\n")

	return sb.String()
}

// ExportSPICE writes netlist to file
func ExportSPICE(mapping *compiler.CrossbarMapping, path string, vdd float64) error {
	content := GenerateSPICE(mapping, vdd)
	return os.WriteFile(path, []byte(content), 0644)
}
```

---

## Executive Summary

**WHAT DEMO 6 IS:**
*   ├── NOT: Production EDA ($100K Cadence replacement)
*   ├── NOT: Full tapeout tool
*   ├── NOT: Foundry-ready output
*   ├── IS: Bridge between your visualizers and real EDA
*   ├── IS: Missing tool for FeCIM design
*   ├── IS: Proof you understand the full stack
*   └── IS: Foundation for real design suite

**THE GAP YOU FILL:**
> "There is no 'OpenROAD for Analog.' You cannot click a button and get a routed FeFET crossbar."

**UNTIL NOW.**

---

## 6-Tab Architecture

```text
┌─────────────────────────────────────────────────────────────┐
│ Demo 6: FeCIM Design Suite (Preview)                       │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│ [Tab 1] [Tab 2] [Tab 3] [Tab 4] [Tab 5] [Tab 6]           │
│ Compiler Layout  Explorer Simulate Export  Learn           │
│                                                             │
│ ⚠️ PREVIEW: Bridge to real open-source EDA tools          │
└─────────────────────────────────────────────────────────────┘
```

### Tab 1: Crossbar Compiler
*   **Purpose:** Translator between AI and hardware.
*   **Input:** Neural network weights (from Demo 3).
*   **Output:** Physical cell assignments + programming parameters.

### Tab 2: Layout Generator
*   **Purpose:** Blueprint of the chip.
*   **Input:** CrossbarMapping from Tab 1.
*   **Output:** Visual layout + GDSII export.

### Tab 3: Design Space Explorer
*   **Purpose:** "What if?" calculator.
*   **Input:** Design parameters (size, levels, ADC, technology).
*   **Output:** Area, power, throughput, efficiency estimates.

### Tab 4: Simulation Bridge
*   **Purpose:** "Run the circuit and see what happens".
*   **Input:** CrossbarMapping + Simulation parameters.
*   **Output:** ngspice simulation results.

### Tab 5: Export Center
*   **Purpose:** "Save your work in every format".
*   **Input:** All generated data.
*   **Output:** Multiple file formats (JSON, CSV, SPICE, KLayout script, CiMLoop YAML).

### Tab 6: Learning Center
*   **Purpose:** "Learn chip design here".
*   **Input:** User's curiosity.
*   **Output:** Guided learning resources (Zero to ASIC, Tiny Tapeout, etc.).

---
