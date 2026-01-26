---
Module: module4-circuits
Name: Peripheral Circuits Visualizer
Entry: cmd/circuits-gui/main.go
Package: multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/gui
Theme: FeCIMTheme
Architecture: Unified 3-view design with embedded interface
Last Updated: 2026-01-25
---

## Bugs Summary

### Fixed Bugs
- [x] BUG-M4-002: Array cell click coordinate calculation (FIXED: uses asymmetric margins)
- [x] BUG-M4-003: computeInputRowContainer now initialized in app.go:143
- [x] BUG-M4-005: Race condition in drawSharedArray (FIXED: currentMode read once under lock at line 253)
- [x] BUG-M4-001: Operations panel visibility sync on mode change
- [x] BUG-M4-004: Missing canvas refresh in Start() for shared array

### Open Bugs
(none)

### Physics Issues (from HZO_PARAMETERS.md research) - ALL FIXED
- [x] PHYS-001: WRITE voltage range corrected (1.2-1.5V for 10nm HZO)
- [x] PHYS-002: READ voltage slider max fixed (now ≤0.5V)
- [x] PHYS-003: COMPUTE voltage note updated ("0.3-0.5V COMPUTE-safe")

### UX Issues (NEW)
- [ ] UX-001: COMPUTE button redundant (auto-compute already triggers on input change)
- [ ] UX-002: Export buttons not implemented (EXPORT SVG, EXPORT SPECS)

---

## Recent Changes (2026-01-25)

### Input Row Position Fixed
- Input vector now appears **ABOVE crossbar array** in COMPUTE mode (was in right panel)
- Uses `computeInputRowContainer` which is shown/hidden based on mode

### Animate Button Enhanced
- Visual feedback during ANIMATE sequence:
  - Step 1: DAC boxes flash **bright yellow**
  - Step 2: Array cells get **cyan border overlay**
  - Step 3: ADC boxes flash **bright green**
- New fields: `animationStep` (0-3), `animationActive` (bool) in app.go:189-191

## Screens

### Main Window (app.go:242-250)
**Purpose**: Top-level window with 3-view architecture
**File**: app.go:242-250
**State**: window, fyneApp, mainTabs (legacy), currentMode (operations)
**Layout**:
```
Border
├─ Top: Header with inline view selector
│  ├─ Label: "FeCIM Peripheral Circuits Visualizer"
│  ├─ Spacer
│  ├─ Label: "View:"
│  ├─ Select: ["OPERATIONS", "COMPARISON", "REFERENCE"]
│  ├─ Spacer
│  └─ Label: "3 Views | DAC -> FeFET -> TIA -> ADC"
├─ Bottom: Footer
│  └─ Label: "FeCIM Ferroelectric Compute-in-Memory | Based on Dr. Tour's Research"
└─ Center: Stack container
   ├─ OPERATIONS view (visible on start)
   ├─ COMPARISON view (hidden)
   └─ REFERENCE view (hidden)
```

---

### OPERATIONS View (tab_operations.go:134-179)
**Purpose**: Unified Write/Read/Compute operations with shared array
**File**: tab_operations.go:134-179
**State**: currentMode (ModeWrite/ModeRead/ModeCompute)
**Layout**:
```
Border
├─ Top: Mode selector panel
│  ├─ HBox
│  │  ├─ Label: "Mode:"
│  │  ├─ RadioGroup: ["WRITE", "READ", "COMPUTE"] horizontal
│  │  ├─ Spacer
│  │  └─ operationsStatusLabel (dynamic status)
│  ├─ operationsModeHelp (mode-specific help text)
│  └─ Separator
├─ Bottom: Action buttons (mode-specific, stacked)
│  ├─ opsWriteButtons (visible in WRITE mode)
│  │  ├─ opsProgramBtn: "PROGRAM CELL" (HighImportance)
│  │  └─ opsProgramRandomBtn: "RANDOM ARRAY"
│  ├─ opsReadButtons (visible in READ mode)
│  │  ├─ opsReadBtn: "READ CELL" (HighImportance)
│  │  └─ opsVerifyBtn: "VERIFY ARRAY"
│  └─ opsComputeButtons (visible in COMPUTE mode)
│     ├─ opsComputeBtn: "COMPUTE" (HighImportance)
│     ├─ opsAnimateBtn: "ANIMATE"
│     └─ opsResetBtn: "RESET"
└─ Center: HSplit (55% array, 45% config panel)
   ├─ Left: Shared array section
   │  ├─ titleLabel: "CROSSBAR ARRAY" (bold, centered)
   │  ├─ computeInputRowContainer (VBox, visible in COMPUTE mode only)
   │  │  ├─ Separator
   │  │  ├─ HBox header
   │  │  │  ├─ Label: "INPUT VECTOR (0-255)" (bold)
   │  │  │  ├─ Spacer
   │  │  │  └─ Button: "RANDOM BITS"
   │  │  ├─ HBox input row (8 columns: x0-x7)
   │  │  │  └─ For each column:
   │  │  │     ├─ Label: "x{i}"
   │  │  │     └─ Entry: opsComputeInputs[i] (45px width)
   │  │  ├─ Label: "0-1V READ-safe (below Ec)" (italic)
   │  │  └─ Separator
   │  ├─ TappableArrayCanvas (480x420px) - INTERACTIVE
   │  │  └─ Raster: drawSharedArray with click detection
   │  │     ├─ Integrated DAC boxes (top, purple, per column)
   │  │     ├─ Integrated ADC boxes (right, green, per row)
   │  │     ├─ Cell grid (8x8 default, color-coded by level)
   │  │     ├─ Selected cell highlight (yellow border)
   │  │     └─ Mode-specific overlays:
   │  │        ├─ WRITE: Arrow pointing to selected cell
   │  │        ├─ READ: Cyan probe circle on selected cell
   │  │        └─ COMPUTE: Input arrows (top), output arrows (right)
   │  ├─ legendLabel: "Level: Low (blue) -> High (red) | Yellow = Selected | Click to select"
   │  ├─ sharedCellInfoLabel: Dynamic cell info
   │  └─ sharedArrayInfoLabel: "Array: 8x8 | 30 levels"
   └─ Right: Mode-specific config panel (stacked, visibility toggled)
      ├─ writeConfigPanel (visible in WRITE mode)
      ├─ readConfigPanel (visible in READ mode)
      └─ computeConfigPanel (visible in COMPUTE mode)
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| TappableArrayCanvas | Custom Widget | Clickable array with integrated DAC/ADC | tab_operations.go:36-128 | sharedArrayCanvas, sharedArrayCellSize, sharedArrayOffsetX/Y | Tapped() -> onArrayCellTapped() | BUG-M4-002 |
| operationsStatusLabel | Label | Dynamic status display | app.go:133 | Text updated by all actions | All action handlers | - |
| operationsModeHelp | Label | Mode-specific help text | app.go:134 | Text set by updateModeHelp() | currentMode state | - |
| sharedCellInfoLabel | Label | Selected cell info | app.go:138 | Text set by updateSharedCellInfo() | selectedRow, selectedCol, arrayWeights | - |
| sharedArrayInfoLabel | Label | Array dimensions | app.go:139 | Text set on init | arrayRows, arrayCols, quantLevels | - |
| computeInputRowContainer | VBox | Input entry row (COMPUTE mode only) | app.go:141-142 | Hidden by default, shown in COMPUTE | currentMode | BUG-M4-003 |

**Data Flow**:

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| Mode selection | RadioGroup OnChanged | currentMode, panel visibility, button visibility, array overlay | tab_operations.go:184-560 |
| Array cell click | TappableArrayCanvas.Tapped() | selectedRow, selectedCol, cell info label, data path labels | tab_operations.go:66-120, 656-666 |
| Level slider change (WRITE) | opsWriteLevelSlider.OnChanged | targetLevel, data path labels, pulse canvas | tab_operations.go:678-686 |
| Voltage slider change (READ) | opsReadVoltageSlider.OnChanged | readVoltage, zone canvas | tab_operations.go:924-930 |
| Input change (COMPUTE) | Entry.OnChanged | inputVector, voltage labels, MVM computation, output labels | tab_operations.go:1148-1162 |
| PROGRAM button | opsProgramBtn.OnClicked | arrayWeights[row][col], array canvas, cell info | tab_operations.go:1484-1498 |
| READ button | opsReadBtn.OnClicked | opsReadResultsLabel (TIA/ADC simulation) | tab_operations.go:1516-1569 |
| COMPUTE button | opsComputeBtn.OnClicked | outputVector, output labels, math label | tab_operations.go:1615-1618 |

---

### WRITE Mode Panel (tab_operations.go:673-768)
**Purpose**: Configure and execute cell programming
**File**: tab_operations.go:673-768
**State**: targetLevel, vMin, vMax, pulseWidth, selectedRow, selectedCol
**Layout**:
```
VBox (writeConfigPanel)
├─ TARGET LEVEL section
│  ├─ Label: "TARGET LEVEL" (bold)
│  ├─ opsWriteLevelLabel: "Target Level: 15 (0-29)"
│  ├─ opsWriteLevelSlider: 0-29 range
│  ├─ Label: "Each level = stable polarization state (~4.9 bits/cell)"
│  └─ Separator
├─ VOLTAGE RANGE section
│  ├─ Label: "VOLTAGE RANGE" (bold)
│  ├─ HBox
│  │  ├─ Label: "Vmin:"
│  │  ├─ Entry: vMinEntry (default 2.0V)
│  │  ├─ Label: "V"
│  │  ├─ Label: "Vmax:"
│  │  ├─ Entry: vMaxEntry (default 5.0V)
│  │  └─ Label: "V"
│  ├─ Label: "Write voltage must exceed Ec (~1.5 MV/cm)"
│  └─ Separator
├─ DATA PATH section
│  ├─ Label: "DATA PATH" (bold)
│  ├─ HBox
│  │  ├─ LabeledBox: "DIGITAL" (ColorPrimary)
│  │  │  └─ opsWriteDigitalLabel: "Level:15\n01111"
│  │  ├─ Label: "->"
│  │  ├─ LabeledBox: "DAC" (ColorAccent)
│  │  │  └─ opsWriteDACLabel: "3.55V"
│  │  ├─ Label: "->"
│  │  └─ LabeledBox: "FeFET" (ColorInfo)
│  │     └─ opsWriteFeFETLabel: "[3,5]\n50.0uS"
│  ├─ Label: "Digital level -> DAC voltage -> FeFET polarization"
│  └─ Separator
└─ PROGRAMMING PULSE section
   ├─ Label: "PROGRAMMING PULSE" (bold)
   └─ opsWritePulseCanvas (350x120px)
      └─ Raster: drawOpsWritePulse (voltage vs time waveform)
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| opsWriteLevelSlider | Slider | Target level selection (0-29) | tab_operations.go:676-686 | targetLevel | OnChanged updates labels and pulse | - |
| opsWriteDigitalLabel | Label | Shows target level in decimal + binary | tab_operations.go:712 | Updated by updateOpsWriteDataPath() | targetLevel | - |
| opsWriteDACLabel | Label | Shows DAC voltage output | tab_operations.go:713 | Updated by updateOpsWriteDataPath() | targetLevel, vMin, vMax | - |
| opsWriteFeFETLabel | Label | Shows cell position and conductance | tab_operations.go:714 | Updated by updateOpsWriteDataPath() | selectedRow, selectedCol, targetLevel | - |
| opsWritePulseCanvas | Raster | Programming pulse waveform | tab_operations.go:727 | Redrawn by drawOpsWritePulse() | targetLevel, vMin, vMax | - |

---

### READ Mode Panel (tab_operations.go:919-1025)
**Purpose**: Sense cell conductance with low voltage
**File**: tab_operations.go:919-1025
**State**: readVoltage, tiaGain, adcBits
**Layout**:
```
VBox (readConfigPanel)
├─ READ PARAMETERS section
│  ├─ Label: "READ PARAMETERS" (bold)
│  ├─ opsReadVoltageLabel: "Read Voltage: 0.50 V"
│  ├─ opsReadVoltageSlider: 0.1-1.5V range
│  ├─ Label: "SAFE: < 1.0V | DANGER: > 2.0V"
│  ├─ HBox
│  │  ├─ Label: "TIA Gain:"
│  │  ├─ Select: ["1", "10", "100"] kOhm
│  │  └─ Label: "kOhm"
│  ├─ HBox
│  │  ├─ Label: "ADC Bits:"
│  │  └─ Select: ["4", "5", "6", "7", "8", "10", "12"]
│  └─ Separator
├─ VOLTAGE ZONES section
│  ├─ Label: "VOLTAGE ZONES" (bold)
│  └─ opsReadZoneCanvas (250x150px)
│     └─ Raster: drawOpsReadZone (colored voltage zones)
├─ DATA PATH section
│  ├─ Label: "DATA PATH" (bold)
│  ├─ HBox
│  │  ├─ LabeledBox: "FeFET" (ColorInfo)
│  │  ├─ Label: "->"
│  │  ├─ LabeledBox: "TIA" (ColorWarning)
│  │  ├─ Label: "->"
│  │  ├─ LabeledBox: "ADC" (ColorSuccess)
│  │  ├─ Label: "->"
│  │  └─ LabeledBox: "DIGITAL" (ColorPrimary)
│  ├─ Label: "FeFET current -> TIA voltage -> ADC -> Level"
│  └─ Separator
└─ READ RESULTS section
   ├─ Label: "READ RESULTS" (bold)
   └─ opsReadResultsLabel: Monospace multi-line result display
      ├─ "Cell [--,--] Read Results"
      ├─ "Programmed Level: --"
      ├─ "Read Current:     -- uA"
      ├─ "TIA Voltage:      -- mV"
      ├─ "ADC Raw:          --"
      ├─ "Decoded Level:    --"
      └─ "Match:            --"
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| opsReadVoltageSlider | Slider | Read voltage control (0.1-1.5V) | tab_operations.go:922-930 | readVoltage | OnChanged refreshes zone canvas | - |
| opsReadZoneCanvas | Raster | Voltage zone visualization | tab_operations.go:955 | Redrawn by drawOpsReadZone() | readVoltage | - |
| opsReadResultsLabel | Label | Read operation results (7 lines) | tab_operations.go:972 | Updated by onOpsRead() | selectedRow, selectedCol, arrayWeights | - |

---

### COMPUTE Mode Panel (tab_operations.go:1134-1257)
**Purpose**: Matrix-vector multiplication (MVM) in memory
**File**: tab_operations.go:1134-1257
**State**: inputVector, outputVector, arrayWeights
**Layout**:
```
VBox (computeConfigPanel)
├─ INPUT VECTOR section (shown ABOVE array via computeInputRowContainer)
│  ├─ HBox header
│  │  ├─ Label: "INPUT VECTOR (0-255)" (bold)
│  │  ├─ Spacer
│  │  └─ Button: "RANDOM BITS"
│  ├─ HBox input row (8 entries: x0-x7)
│  │  └─ For each column:
│  │     ├─ Label: "x{i}"
│  │     └─ Entry: 45px width, 0-255 range
│  └─ Label: "0-1V READ-safe (below Ec)" (italic)
├─ OUTPUT VECTOR section
│  ├─ Separator
│  ├─ Label: "OUTPUT VECTOR" (bold)
│  ├─ Label: "I_row -> TIA (10k) -> ADC (5-bit):"
│  ├─ Grid (2 columns, 8 rows)
│  │  └─ opsComputeOutputLabels[0-7]: "y{i}: 50.1uA -> 0.50V -> L16"
│  └─ Label: "IDEAL CROSSBAR: No IR drop or sneak paths (see Module 2)" (italic)
├─ MATH section
│  ├─ Separator
│  ├─ Label: "MATH (Row 0 Breakdown)" (bold)
│  └─ opsComputeMathLabel: Monospace KCL equation display
│     ├─ "I0 = 50*0.50 + 75*0.75 + ... (KCL sum)"
│     ├─ "   = 160.5 uA"
│     └─ "ALL ROWS IN PARALLEL!"
└─ PERFORMANCE section
   ├─ Separator
   ├─ Label: "PERFORMANCE" (bold)
   ├─ Label: "DAC: 5ns | Array settle: 5ns | ADC: 10ns"
   ├─ Label: "TOTAL: ~20ns for full MVM!"
   └─ Label: "GPU equivalent: ~1000 cycles"
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| opsComputeInputs | Entry array | Input vector (x0-x7) | tab_operations.go:1136 | inputVector | OnChanged triggers computeAndUpdateAll() | - |
| opsComputeVoltageLabels | Label array | Voltage display per input | tab_operations.go:1137 | Updated by Entry.OnChanged | inputVector | - |
| opsComputeOutputLabels | Label array | Output display (y0-y7) | tab_operations.go:1176 | Updated by computeAndUpdateAll() | outputVector | - |
| opsComputeMathLabel | Label | KCL equation breakdown (row 0) | tab_operations.go:1183 | Updated by updateOpsComputeMath() | arrayWeights, inputVector | - |
| randomBitsBtn | Button | Randomize input vector | tab_operations.go:1192 | Triggers updateOpsComputeInputs() | inputVector | - |

**Data Flow**:

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| Input entry change | Entry.OnChanged | inputVector[i], voltage label, MVM computation | tab_operations.go:1148-1162 |
| RANDOM BITS button | Button.OnClicked | All inputVector values, all entries, MVM | tab_operations.go:1193-1200 |
| computeAndUpdateAll() | Multiple triggers | outputVector (via MVM), output labels, math label | tab_operations.go:1282-1332 |
| MVM computation | computeAndUpdateAll() | For each row: sum(G_ij * V_j) -> outputVector[i] | tab_operations.go:1284-1297 |

---

### COMPARISON View (tab_comparison.go:20-71)
**Purpose**: Compare FeFET vs GPU vs CPU architectures
**File**: tab_comparison.go:20-71
**State**: compArraySize (8, 16, 32, 64)
**Layout**:
```
Border
├─ Top: Header
│  ├─ RichText: "COMPARISON: Compare FeFET crossbar..." (markdown, word wrap)
│  └─ Separator
├─ Bottom: Button bar
│  ├─ Separator
│  ├─ HBox
│  │  ├─ runBtn: "RUN COMPARISON" (HighImportance)
│  │  ├─ animateBtn: "ANIMATE"
│  │  ├─ scaleBtn: "SCALE UP"
│  │  ├─ Spacer
│  │  └─ compStatusLabel: "8×8 Matrix-Vector Multiply Comparison"
└─ Center: VBox
   ├─ Grid (2 columns)
   │  ├─ VBox: Architecture comparison
   │  │  ├─ Label: "ARCHITECTURE COMPARISON"
   │  │  └─ compArchCanvas (400x200px)
   │  │     └─ Raster: drawCompArch (CPU/GPU/FeFET blocks)
   │  └─ VBox: Timing comparison
   │     ├─ Label: "TIMING COMPARISON"
   │     └─ compTimingCanvas (400x150px)
   │        └─ Raster: drawCompTiming (horizontal bars: CPU 500ns, GPU 50ns, FeFET 20ns)
   ├─ Separator
   └─ Grid (2 columns)
      ├─ VBox: Energy comparison
      │  ├─ Label: "ENERGY COMPARISON"
      │  └─ compEnergyCanvas (400x200px)
      │     └─ Raster: drawCompEnergy (bars: CPU 64000pJ, GPU 6400pJ, FeFET 3.2pJ)
      └─ VBox: Live comparison table
         ├─ Label: "LIVE COMPARISON"
         └─ Grid (4 columns)
            ├─ Headers: ["", "Time", "Energy", "TOPS/W"]
            ├─ CPU row: ["CPU", "500 ns", "64,000 pJ", "0.5"]
            ├─ GPU row: ["GPU", "50 ns", "6,400 pJ", "5.0"]
            └─ FeFET row: ["FeFET", "20 ns", "3.2 pJ", "2,000"]
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| compArchCanvas | Raster | Architecture diagram | tab_comparison.go:75 | Redrawn by drawCompArch() | None | - |
| compTimingCanvas | Raster | Timing bar chart | tab_comparison.go:157 | Redrawn by drawCompTiming() | None | - |
| compEnergyCanvas | Raster | Energy bar chart | tab_comparison.go:243 | Redrawn by drawCompEnergy() | None | - |
| compTableLabels | Label array (16) | Comparison table cells | tab_comparison.go:315 | Updated by onScaleUpComparison() | compArraySize | - |
| compStatusLabel | Label | Status display | tab_comparison.go:102 | Updated by all button handlers | None | - |

**Data Flow**:

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| RUN COMPARISON button | Button.OnClicked | Refreshes all canvases, updates status | tab_comparison.go:354-371 |
| ANIMATE button | Button.OnClicked | Step-by-step status updates (goroutine) | tab_comparison.go:373-397 |
| SCALE UP button | Button.OnClicked | compArraySize (cycles 8->16->32->64->8), table values | tab_comparison.go:399-441 |

---

### REFERENCE View (tab_reference.go:25-53)
**Purpose**: Timing diagrams + specifications reference
**File**: tab_reference.go:25-53
**State**: timingOpSelect, specArraySizeSelect
**Layout**:
```
Border
├─ Top: Section selector
│  ├─ HBox
│  │  ├─ Label: "Reference:"
│  │  ├─ Select: ["TIMING DIAGRAMS", "SPECIFICATIONS"]
│  │  └─ Spacer
│  └─ Separator
└─ Center: Stack
   ├─ refTimingSection (visible by default)
   └─ refSpecsSection (hidden by default)
```

---

### REFERENCE - Timing Section (tab_reference.go:74-118)
**Purpose**: Signal waveform timing diagrams
**File**: tab_reference.go:74-118
**State**: timingOpSelect ("WRITE", "READ", "COMPUTE")
**Layout**:
```
Border
├─ Top: Header
│  ├─ RichText: "TIMING DIAGRAMS: View signal waveforms..." (markdown, word wrap)
│  ├─ Separator
│  └─ HBox
│     ├─ Label: "OPERATION:"
│     └─ timingOpSelect: Select dropdown
├─ Bottom: Button bar
│  ├─ Separator
│  └─ HBox
│     ├─ animateBtn: "ANIMATE"
│     ├─ exportBtn: "EXPORT SVG"
│     ├─ Spacer
│     └─ timingStatusLabel: "Select operation to view timing"
└─ Center: VScroll
   └─ VBox
      ├─ Label: "WRITE TIMING"
      ├─ timingWriteCanvas (600x200px)
      │  └─ Raster: drawTimingWrite (6 signals: CLK, ROW_SEL, COL_SEL, DAC_EN, V_PROG, DONE)
      ├─ Separator
      ├─ Label: "READ TIMING"
      ├─ timingReadCanvas (600x180px)
      │  └─ Raster: drawTimingRead (5 signals: CLK, V_READ, I_SENSE, ADC_EN, DATA_OUT)
      ├─ Separator
      ├─ Label: "COMPUTE TIMING"
      └─ timingComputeCanvas (600x200px)
         └─ Raster: drawTimingCompute (6 signals: CLK, INPUT_VALID, DAC_ALL, ARRAY_SETTLE, ADC_ALL, OUTPUT_VALID)
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| timingOpSelect | Select | Operation selector (WRITE/READ/COMPUTE) | tab_reference.go:80 | None | OnChanged refreshes canvases | - |
| timingWriteCanvas | Raster | WRITE timing diagram (70ns total) | tab_reference.go:122 | Redrawn by drawTimingWrite() | None | - |
| timingReadCanvas | Raster | READ timing diagram (20ns total) | tab_reference.go:252 | Redrawn by drawTimingRead() | None | - |
| timingComputeCanvas | Raster | COMPUTE timing diagram (20ns total) | tab_reference.go:388 | Redrawn by drawTimingCompute() | None | - |
| timingStatusLabel | Label | Timing status display | tab_reference.go:94 | Updated by onAnimateTiming() | None | - |

**Data Flow**:

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| ANIMATE button | Button.OnClicked | Step-by-step status updates per selected operation | tab_reference.go:571-620 |
| EXPORT SVG button | Button.OnClicked | Status message (not implemented) | tab_reference.go:623-628 |

---

### REFERENCE - Specs Section (tab_reference.go:634-726)
**Purpose**: Detailed component specifications
**File**: tab_reference.go:634-726
**State**: specArraySizeSelect, specQuantLevelSelect, specDACBitsSelect, specADCBitsSelect, specTIAGainSelect
**Layout**:
```
Border
├─ Top: Header
│  ├─ RichText: "SPECIFICATIONS: Detailed electrical..." (markdown, word wrap)
│  └─ Separator
├─ Bottom: Button bar
│  ├─ Separator
│  └─ HBox
│     ├─ exportBtn: "EXPORT SPECS"
│     ├─ compareBtn: "COMPARE TO GPU"
│     ├─ Spacer
│     └─ specStatusLabel: "System specifications"
└─ Center: VScroll
   └─ HBox (2 columns)
      ├─ VBox (left column)
      │  ├─ Label: "ARRAY CONFIGURATION" (bold)
      │  ├─ Spacer
      │  ├─ HBox
      │  │  ├─ Label: "Array Size:"
      │  │  ├─ specArraySizeSelect: ["8", "16", "32", "64", "128"]
      │  │  ├─ Label: "×"
      │  │  ├─ specArraySizeSelect (same)
      │  │  └─ Label: "= 1024 cells"
      │  ├─ HBox
      │  │  ├─ Label: "Quantization:"
      │  │  ├─ specQuantLevelSelect: ["2", "4", "8", "16", "30", "32", "64", "128", "256"]
      │  │  └─ Label: "levels (~4.9 bits/cell)"
      │  ├─ Label: "Total Storage: 1024 × 4.9 = 5017 bits"
      │  ├─ Separator
      │  ├─ Label: "DAC SPECIFICATIONS" (bold)
      │  ├─ HBox
      │  │  ├─ Label: "Resolution:"
      │  │  ├─ specDACBitsSelect: ["4", "5", "6", "7", "8", "10", "12"]
      │  │  └─ Label: "bits"
      │  ├─ Label: Multi-line specs (count, range, timing, power, INL, DNL)
      │  ├─ Separator
      │  ├─ Label: "ADC SPECIFICATIONS" (bold)
      │  ├─ HBox
      │  │  ├─ Label: "Resolution:"
      │  │  ├─ specADCBitsSelect: ["4", "5", "6", "7", "8", "10", "12"]
      │  │  └─ Label: "bits"
      │  └─ Label: Multi-line specs (count, range, timing, power, ENOB, SNR)
      └─ VBox (right column)
         ├─ Label: "TIA SPECIFICATIONS" (bold)
         ├─ HBox
         │  ├─ Label: "Gain:"
         │  ├─ specTIAGainSelect: ["1", "10", "100"]
         │  └─ Label: "kOhm"
         ├─ Label: Multi-line specs (count, gain, bandwidth, range, noise)
         ├─ Separator
         ├─ Label: "FeFET CELL SPECIFICATIONS" (bold)
         ├─ Grid (2 columns)
         │  ├─ Material: "HfZrO2 (HZO)"
         │  ├─ Thickness: "10 nm"
         │  ├─ Levels: "30 discrete states (~4.9 bits/cell)"
         │  ├─ Conductance: "1 µS to 100 µS"
         │  ├─ Read Voltage: "0.5 V"
         │  ├─ Write Voltage: "2.0 V to 5.0 V"
         │  ├─ Write Time: "50 ns"
         │  ├─ Endurance: "10^12 cycles"
         │  ├─ Retention: "10 years"
         │  └─ Cell Size: "~0.01 µm²"
         ├─ Separator
         ├─ Label: "SYSTEM SUMMARY" (bold)
         └─ specSummaryLabel: Monospace table
            └─ Component | Count | Power | Area | Latency table
```

**Components**:

| Component | Type | Purpose | File:Line | State | Bindings | Bug |
|-----------|------|---------|-----------|-------|----------|-----|
| specArraySizeSelect | Select | Array size selector | tab_reference.go:731 | None | OnChanged updates specSummaryLabel | - |
| specQuantLevelSelect | Select | Quantization level selector | tab_reference.go:738 | None | None | - |
| specDACBitsSelect | Select | DAC resolution selector | tab_reference.go:757 | None | None | - |
| specADCBitsSelect | Select | ADC resolution selector | tab_reference.go:785 | None | None | - |
| specTIAGainSelect | Select | TIA gain selector | tab_reference.go:812 | None | None | - |
| specSummaryLabel | Label | System summary table (monospace) | tab_reference.go:902 | Updated by updateSpecSummary() | specArraySizeSelect | - |
| specStatusLabel | Label | Status display | tab_reference.go:661 | Updated by button handlers | None | - |

**Data Flow**:

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| Array size change | specArraySizeSelect.OnChanged | updateSpecSummary() recalculates throughput/efficiency | tab_reference.go:732-734, 906-937 |
| EXPORT SPECS button | Button.OnClicked | Status message (not implemented) | tab_reference.go:939-943 |
| COMPARE TO GPU button | Button.OnClicked | Status message with comparison summary | tab_reference.go:946-951 |

---

BugDetails:
  - id: BUG-M4-001
    component: writeConfigPanel, readConfigPanel, computeConfigPanel
    severity: Medium
    description: Panel visibility may not update correctly when mode changes via RadioGroup
    expected: Only the panel for the selected mode should be visible
    actual: Multiple panels may be visible simultaneously or none visible
    file: tab_operations.go:562-595
    suggested_fix:
      // Hide ALL panels first
      ca.writeConfigPanel.Hide()
      ca.readConfigPanel.Hide()
      ca.computeConfigPanel.Hide()

      // Then show only the selected panel
      switch mode {
      case ModeWrite:
          ca.writeConfigPanel.Show()
      case ModeRead:
          ca.readConfigPanel.Show()
      case ModeCompute:
          ca.computeConfigPanel.Show()
      }

  - id: BUG-M4-002
    component: TappableArrayCanvas.Tapped()
    severity: High
    description: Click coordinate calculation must account for asymmetric margins (top: 70px for DAC, right: 70px for ADC)
    expected: Clicking on a cell selects that cell
    actual: Click may select wrong cell due to incorrect offset calculation
    file: tab_operations.go:66-120
    suggested_fix:
      // Use EXACT same asymmetric margins as drawSharedArray
      topMargin := 70   // Space for integrated DAC boxes
      rightMargin := 70 // Space for integrated ADC boxes
      bottomMargin := 30
      leftMargin := 30

  - id: BUG-M4-003
    component: computeInputRowContainer (VBox)
    severity: High
    description: computeInputRowContainer is used in createSharedArraySection() but not declared in CircuitsApp struct
    expected: Container should be declared and initialized
    actual: Field missing from struct, causing compilation error or nil reference
    file: tab_operations.go:228
    suggested_fix: |
      // Add after line 186 in app.go
      computeInputRowContainer *fyne.Container

  - id: BUG-M4-004
    component: EmbeddedCircuitsApp.Start()
    severity: Low
    description: Start() refreshes legacy canvases but not sharedArrayCanvas
    expected: All visible canvases should refresh when tab is selected
    actual: Shared array may show stale rendering
    file: embedded.go:73-93
    suggested_fix: |
      func (e *EmbeddedCircuitsApp) Start() {
          // Refresh operations view canvases
          e.refreshSharedArray()

          // Refresh legacy canvases (for backward compatibility)
          e.refreshWriteArray()
          // ... etc
      }

  - id: BUG-M4-005
    component: drawSharedArray()
    severity: Medium
    description: currentMode is read multiple times within drawSharedArray without holding lock
    expected: All state reads should be atomic within a single RLock/RUnlock pair
    actual: Mode could change mid-draw, causing inconsistent rendering
    file: tab_operations.go:242-526
    suggested_fix: |
      ca.mu.RLock()
      rows := ca.arrayRows
      cols := ca.arrayCols
      weights := ca.arrayWeights
      selectedRow := ca.selectedRow
      selectedCol := ca.selectedCol
      levels := ca.quantLevels
      mode := ca.currentMode // Read once and reuse
      ca.mu.RUnlock()

---

## State Machine

### OperationMode State Transitions
```
Initial State: ModeWrite

ModeWrite
  └─> "READ" selected -> ModeRead
  └─> "COMPUTE" selected -> ModeCompute

ModeRead
  └─> "WRITE" selected -> ModeWrite
  └─> "COMPUTE" selected -> ModeCompute

ModeCompute
  └─> "WRITE" selected -> ModeWrite
  └─> "READ" selected -> ModeRead
```

**Actions on State Change**:
1. Update panel visibility (writeConfigPanel, readConfigPanel, computeConfigPanel)
2. Update button visibility (opsWriteButtons, opsReadButtons, opsComputeButtons)
3. Update mode help text (operationsModeHelp)
4. Refresh shared array canvas (with mode-specific overlay)
5. Update cell info label
6. If entering COMPUTE: call computeAndUpdateAll()

---

## Key Patterns

### 1. Tappable Canvas Pattern
Custom widget wrapping canvas.Raster with Tapped() interface:
```go
type TappableArrayCanvas struct {
    widget.BaseWidget
    raster *canvas.Raster
    onTap  func(row, col int)
    ca     *CircuitsApp
}

func (t *TappableArrayCanvas) Tapped(e *fyne.PointEvent) {
    // Convert screen coordinates to grid coordinates
    col := (int(e.Position.X) - offsetX) / cellSize
    row := (int(e.Position.Y) - offsetY) / cellSize
    t.onTap(row, col)
}

func (t *TappableArrayCanvas) Cursor() desktop.Cursor {
    return desktop.PointerCursor // Show pointer on hover
}
```

### 2. Integrated DAC/ADC Visualization
Array canvas draws peripheral components directly:
```go
// DAC boxes at TOP of each column (integrated visualization)
dacY := offsetY - dacBoxHeight - 10
for c := 0; c < min(8, cols); c++ {
    dacX := offsetX + c*cellSize + 2
    drawRect(img, dacX, dacY, dacBoxWidth, dacBoxHeight, dacColor)
    // Show voltage and column label
}

// ADC boxes at RIGHT of each row
adcX := offsetX + gridW + 8
for r := 0; r < min(8, rows); r++ {
    adcY := offsetY + r*cellSize + 2
    drawRect(img, adcX, adcY, adcBoxWidth, adcBoxHeight, adcColor)
    // Show ADC level and row label
}
```

### 3. Auto-Compute on Input Change
COMPUTE mode updates output immediately when inputs change:
```go
ca.opsComputeInputs[i].OnChanged = func(s string) {
    var v int
    fmt.Sscanf(s, "%d", &v)
    ca.mu.Lock()
    ca.inputVector[idx] = v
    ca.mu.Unlock()
    // Auto-compute on input change (no button press needed)
    ca.computeAndUpdateAll()
}
```

### 4. Mode-Specific Array Overlays
drawSharedArray() renders different overlays per mode:
```go
switch mode {
case ModeWrite:
    // Show target level indicator (arrow pointing to cell)
case ModeRead:
    // Show read probe indicator (cyan circle)
case ModeCompute:
    // Show input arrows (top), output arrows (right)
}
```

### 5. Bitmap Font Text Rendering
Custom 5x7 pixel font for canvas text:
```go
drawSimpleText(img, "Level 15", x, y, color.White)
// Uses fontPatterns map with '0', '1' patterns per character
```

### 6. Thick Timing Diagram Lines
Timing diagrams use thick horizontal lines for better visibility:
```go
drawThickHorizontalLine(img, x1, x2, y, thickness, signalColor)
// Draws horizontal line with vertical thickness (not single pixel)
```

---

## Thread Safety

### Mutex Protection
All shared state accessed via ca.mu (RWMutex):
- arrayWeights (read/write)
- inputVector, outputVector (read/write)
- selectedRow, selectedCol (read/write)
- currentMode (read/write)
- Configuration values (vMin, vMax, readVoltage, etc.) (read/write)

### Canvas Refresh Pattern
All canvas refresh calls wrapped in fyne.Do():
```go
fyne.Do(func() {
    ca.sharedArrayCanvas.Refresh()
})
```

### Goroutine Usage
Goroutines used for:
1. VERIFY button: Background array verification (tab_operations.go:1584-1611)
2. ANIMATE button: Step-by-step status updates (tab_operations.go:1661-1675)
3. Comparison ANIMATE: Step-by-step animation (tab_comparison.go:386-396)
4. Timing ANIMATE: Step-by-step status updates (tab_reference.go:610-620)

---

## Performance Optimizations

### 1. Pre-copy Input/Output Vectors in drawSharedArray
Avoids repeated RLock calls in tight loops:
```go
// BEFORE loop: Copy data once
inputVectorCopy := make([]int, dacColCount)
ca.mu.RLock()
copy(inputVectorCopy, ca.inputVector[:dacColCount])
ca.mu.RUnlock()

// IN loop: Use pre-copied data (no lock needed)
for c := 0; c < dacColCount; c++ {
    inputVal := inputVectorCopy[c]
    // ... render DAC box
}
```

### 2. Min Display Count for Large Arrays
Only render first 8 rows/columns in visualizations:
```go
for c := 0; c < min(8, cols); c++ {
    // Render first 8 columns only
}
```

---

## Physics Constants

| Parameter | Value | Source | File:Line |
|-----------|-------|--------|-----------|
| FeCIMLevels | 30 | Dr. Tour COSM 2025 | app.go:26 |
| DefaultSize | 8 | Default array size | app.go:28 |
| DefaultDACBits | 8 | 8-bit DAC resolution | app.go:29 |
| DefaultADCBits | 8 | 8-bit ADC resolution | app.go:30 |
| vMin (write) | 2.0V | Write voltage range | app.go:197 |
| vMax (write) | 5.0V | Write voltage range | app.go:198 |
| readVoltage | 0.5V | Safe read voltage | app.go:200 |
| tiaGain | 10.0 kOhm | TIA transimpedance | app.go:201 |
| Conductance range | 1-100 µS | Per-cell conductance | helpers.go (calculated) |

---

## External Dependencies

### Fyne GUI Framework
- fyne.io/fyne/v2
- fyne.io/fyne/v2/app
- fyne.io/fyne/v2/canvas
- fyne.io/fyne/v2/container
- fyne.io/fyne/v2/layout
- fyne.io/fyne/v2/widget
- fyne.io/fyne/v2/driver/desktop (for Cursor() interface)

### Internal Packages
- multilayer-ferroelectric-cim-visualizer/module4-circuits/pkg/peripherals (DAC, ADC, TIA, ChargePump)
- multilayer-ferroelectric-cim-visualizer/shared/theme (FeCIMTheme, ColorPrimary, etc.)
- multilayer-ferroelectric-cim-visualizer/shared/widgets (DebugInteraction)

### Peripheral Components (module4-circuits/pkg/peripherals)
- DAC: DefaultDAC() - 8-bit, 0-1V output
- ADC: DefaultADC() - 8-bit, 0-1V input, Convert(voltage) returns level
- TIA: DefaultTIA() - 10kOhm gain, Convert(current) returns voltage
- ChargePump: DefaultChargePump() - Voltage boost for writes

---

## Notes

1. **Unified Operations View**: This is the PRIMARY interface, replacing the legacy 7-tab design. WRITE/READ/COMPUTE are modes within a single view, not separate tabs.

2. **Integrated DAC/ADC Visualization**: DAC boxes are rendered at the TOP of each column, ADC boxes at the RIGHT of each row, directly in the array canvas. This provides a compact, integrated data flow visualization.

3. **Auto-Compute**: COMPUTE mode automatically recalculates MVM whenever inputs change. No need to press COMPUTE button explicitly (though button still works for manual refresh).

4. **Click Detection Complexity**: Array click detection must account for asymmetric margins due to integrated DAC/ADC boxes. See BUG-M4-002 for details.

5. **Missing Field Warning**: computeInputRowContainer is used but not declared in app.go struct. See BUG-M4-003 for fix.

6. **Bitmap Font**: Custom 5x7 pixel font for canvas text rendering, supporting alphanumeric + basic symbols. Defined in font.go with fontPatterns map.

7. **Timing Diagrams**: Show precise nanosecond-scale timing for WRITE (70ns), READ (20ns), and COMPUTE (20ns) operations. Use thick lines (drawThickHorizontalLine) for better visibility.

8. **Comparison View**: Compares FeFET vs GPU vs CPU. Conservative energy claims per CLAUDE.md accuracy policy (10-100x savings, not 10M×).

9. **Specifications**: Comprehensive component specs including DAC/ADC resolution, TIA gain, FeFET cell parameters (30 levels, 10nm HZO, 10^12 cycle endurance).

10. **Embedded Interface**: Implements BuildContent(), Start(), Stop() for integration with main visualizer (cmd/fecim-visualizer).
