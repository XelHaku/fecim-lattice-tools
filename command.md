/ralph-loop "BUILD DEMO 6: FeCIM Design Suite

GOAL: Create the EDA bridge tool from plan-demo6.md

PLAN FILE: <local-path>
Look for section: COPY-PASTE STARTER TEMPLATES

---
PHASE 1: Setup (Do First)
---

1. Create project:
   mkdir -p demo6-eda/cmd/eda-gui
   mkdir -p demo6-eda/pkg/compiler
   mkdir -p demo6-eda/pkg/export
   mkdir -p demo6-eda/pkg/gui/tabs
   mkdir -p demo6-eda/data
   cd demo6-eda
   go mod init demo6-eda
   go get fyne.io/fyne/v2@latest

2. Create pkg/compiler/types.go
   Copy Template 1 from plan-demo6.md
   Structs: CompileConfig, CellAssignment, CrossbarMapping, Stats

3. Test: go build ./pkg/compiler

---
PHASE 2: Compiler (Core Logic)
---

1. Create pkg/compiler/compiler.go
   Copy Template 2 from plan-demo6.md
   Main function: Compile(weights [][]float64, config CompileConfig) (*CrossbarMapping, error)

2. Create pkg/compiler/compiler_test.go
   Copy Template 3 from plan-demo6.md

3. Test: go test ./pkg/compiler -v
   ALL TESTS MUST PASS

---
PHASE 3: Export Functions
---

1. Create pkg/export/json.go
   Function: ExportJSON(mapping, path) - uses json.MarshalIndent

2. Create pkg/export/csv.go
   Function: ExportCSV(mapping, path) - row,col,weight,level,conductance

3. Create pkg/export/spice.go
   Copy Template 5 from plan-demo6.md
   Function: GenerateSPICE(mapping, vdd) returns SPICE netlist string

4. Test: go test ./pkg/export -v

---
PHASE 4: GUI Shell
---

1. Create cmd/eda-gui/main.go
   Copy Template 4 from plan-demo6.md
   Creates Fyne window with 6 tabs

2. Test: go run ./cmd/eda-gui
   Window should open with 6 tab buttons

---
PHASE 5: Tab 1 - Compiler Tab
---

1. Create pkg/gui/tabs/compiler_tab.go
   - Load Weights button with file picker
   - Config inputs: rows, cols, levels
   - Compile button
   - Results text showing stats

2. Wire into main.go

3. Test: Load sample JSON, compile, see stats

---
PHASE 6: Tab 5 - Export Tab
---

1. Create pkg/gui/tabs/export_tab.go
   - Checkboxes for JSON, CSV, SPICE
   - Export button
   - Shows list of created files

2. Test full workflow: Tab1 compile then Tab5 export

---
PHASE 7: Tab 2 - Layout Tab
---

1. Create pkg/gui/tabs/layout_tab.go
   - Canvas showing colored grid
   - Blue = low conductance, Red = high
   - Click cell shows details

---
MVP DONE after Phase 7
---

---
SAMPLE DATA - Create this file:
---

File: data/sample_weights_8x8.json

{
  \"rows\": 8,
  \"cols\": 8,
  \"weights\": [
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

---
KEY FORMULAS
---

Quantize: level = round((weight + maxWeight) / (2 * maxWeight) * (Levels-1))
Conductance: G = GMin + (level / (Levels-1)) * (GMax - GMin)
Resistance: R = 1000000 / G (ohms from microsiemens)

---
COMMANDS
---

Build: go build ./...
Test: go test ./... -v
Run: go run ./cmd/eda-gui

---
SUCCESS = MVP working with Tab 1, Tab 2, Tab 5
---

Output: Weights compile to cells, export to JSON/CSV/SPICE, visual grid shows" --max-iterations 1000
