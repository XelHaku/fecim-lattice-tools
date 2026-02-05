---
Module: module4-circuits
Name: Peripheral Circuits Visualizer
Entry: cmd/circuits-gui/main.go
Package: fecim-lattice-tools/module4-circuits/pkg/gui
Theme: FeCIMTheme
Architecture: Unified 3-view design with embedded interface
Last Updated: 2026-02-03
---

Conventions:
  - File paths are relative to module4-circuits unless noted
  - Widget types refer to Fyne (`widget.*`, `container.*`, `canvas.*`) or shared widgets
  - Bindings list event handlers or UI update calls impacting the component

Related (authoritative) control semantics:
- **Unified GUI control contract (semantics + units + clamping):** `docs/development/GUI/GUI.module4.unified-controls.md`

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
- [x] PHYS-001: WRITE voltage range corrected (derived from material Vc)
- [x] PHYS-002: READ voltage slider max fixed (derived from FieldMinRatio * Vc)
- [x] PHYS-003: COMPUTE voltage note updated (uses read range for compute-safe)
- [x] PHYS-004: Voltage ranges now loaded from physics.yaml calibration section

### UX Issues - ALL FIXED
- [x] UX-001: COMPUTE button redundant (auto-compute implemented on input change)
- [x] UX-002: Export buttons - FIXED (2026-01-26): Now show "Coming soon" dialogs with helpful workarounds
- [x] UX-003: Mode selection refactored (2026-01-27): Mode buttons replace RadioGroup
- [x] UX-004: No target level selector for WRITE - FIXED (2026-01-27): Write level slider (0-29) added
- [x] UX-005: Input vector entries not visible - FIXED (2026-01-27): Persistent entries in COMPUTE mode panel
- [x] UX-C1: Cell selection fill obscured state - FIXED (2026-01-27): Gold border only, no fill
- [x] UX-H2: No target cell indicator - FIXED (2026-01-27): "Target: Row X, Col Y" label in write panel
- [x] UX-H3: No undo functionality - FIXED (2026-01-27): Undo button with single-level history
- [x] UX-H4: WL labels unclear - FIXED (2026-01-27): "Row 0" labels, disabled in passive mode
- [x] UX-M4-001: ADC saturation threshold hardcoded - FIXED (2026-01-28): Uses dynamic `adc.Bits` to calculate max level
- [x] UX-M4-002: Fallback numLevels undocumented - FIXED (2026-01-28): Added comment documenting FeCIMLevels consistency

---

## Recent Changes (2026-02-03)

### Physics Alignment
- Updated timing/energy labels and reference diagrams to match `docs/peripheral-circuits/PHYSICS.md` (Read ~76ns, Write ~203ns).
- Updated comparison/specs panels to use the same read-latency baseline and revised FeFET timing labels.

---

## Recent Changes (2026-01-28)

### Voltage Rules Implementation (Major Feature)

**4-Phase Write Sequence:**
- Implements proper ferroelectric write timing: RESET → HOLD → WRITE → HOLD
- Animated timing diagram shows phase progression
- Each phase has configurable duration (ns)
- State machine in `WriteSequenceState` struct

**ISPP (Incremental Step Pulse Programming):**
- Iterative write algorithm with overshoot detection
- Automatic reset-to-saturation when overshoot detected
- Direction tracking (ascending/descending) per cell
- Max 10 iterations with status display
- States: `ISPPSuccess`, `ISPPOvershoot`, `ISPPMaxIter`, `ISPPCancelled`

**V/2 Half-Select Visualization:**
- Gold overlay for target cell during WRITE
- Amber overlay for half-selected cells (V/2 voltage)
- Architecture-specific: critical for passive (0T1R) mode
- Shows sneak path risk in crossbar arrays

**Per-Level Voltage Calibration:**
- 30-level voltage arrays for fine-grained control
- Linear interpolation between calibration points
- Hysteresis direction tracking per cell

### UX Improvements
- **Dynamic ADC saturation** - Saturation check now uses `adc.Bits` to calculate max level (2^bits - 1) instead of hardcoded 31
- **Documented fallback constant** - Added comment explaining `numLevels := 30` fallback matches `FeCIMLevels` constant

---

## Recent Changes (2026-01-27)

### UX Improvements (Phase 2)
- **Cell selection feedback** - Gold border only, no fill (C1 fix) - preserves visibility of state color
- **Target cell label** - "Target: Row X, Col Y" label in write mode panel (H2 fix)
- **Undo functionality** - Single-level undo for array changes with dedicated button (H3 fix)
- **WL checkbox labels** - Changed "WL0" to "Row 0" for clarity (H4 fix)
- **Passive mode enforcement** - In 0T1R architecture:
  - All WL checkboxes always checked and disabled (cannot turn off)
  - DeviceState ignores WL change requests when `isPassive=true`
  - Defense-in-depth: UI + data layer both enforce constraint
- **Dynamic quantLevels** - Write level slider uses `ca.quantLevels` instead of hardcoded 29
- **Mid-level array initialization** - Array cells start at mid-level (15 for 30 states, not 0)
- **Blue-gray-red color mapping** - New color scheme:
  - Blue gradient for levels below mid (0 to mid-1)
  - Gray for mid-level (neutral state)
  - Red gradient for levels above mid (mid+1 to max)

### Research Documentation (New)
- **docs/research/circuits.CIM-fundamentals.md** - Physics basis for READ/WRITE/COMPUTE operations
- **docs/research/MODULE4-PHYSICS-IMPROVEMENTS.md** - Gap analysis with severity ratings
- **docs/plans/module4-plan-improvements.md** - 12-task implementation plan across 4 phases

### Mode-First UX Redesign (Phase 1)
- **Mode bar at top** - READ/WRITE/COMPUTE buttons now prominent at top of OPERATIONS view
- **Write level slider** - 0-29 slider with real-time voltage display (addresses UX-004)
- **Input vector panel** - 8 persistent entry fields for COMPUTE mode (addresses UX-005)
- **Contextual panels** - Show/hide based on mode (WRITE shows slider, COMPUTE shows entries)
- **Mode buttons removed from WL selector** - now only at top mode bar

### Major Refactor: Unified Device Simulation View
- **Replaced** `tab_operations.go` with `tab_unified.go` - single unified device simulation
- **New file** `device_state.go` - DeviceState struct manages all simulation state
- **Mode buttons** replace RadioGroup: READ, WRITE, COMPUTE buttons with visual highlighting
- **Material selector** - dropdown to select ferroelectric material (FeCIM HZO, etc.)
- **Architecture toggle** - PASSIVE/1T1R/2T1R buttons
- **Dynamic voltage ranges** - derived from physics.yaml and material properties (no hardcoded values)

### Voltage Range System
- All voltage thresholds now derived from material properties:
  - **Coercive voltage (Vc)** = Ec × thickness (from material)
  - **Read range**: 0 to FieldMinRatio × Vc (from physics.yaml calibration.field_min_ratio), capped at 1.0V and floored at 0.1V for ADC/DAC practicality
  - **Write range**: Vc to FieldMaxRatio × Vc (from physics.yaml calibration.field_max_ratio)
- DAC preset buttons show actual voltage ranges based on selected material
- No hardcoded voltage constants in device_state.go

### Operation Mode System
- **OpMode enum** replaces OperationMode:
  - `OpModeRead`: Single row active, safe voltage (0 to read max)
  - `OpModeWrite`: Single row active, write voltage on selected column
  - `OpModeCompute`: All rows active, input vector (0 to read max for MVM)
- Mode buttons auto-configure WL and DAC settings when clicked

---

## UI Analysis (2026-01-30)

### 1. Current Layout Issues (from visual analysis)

**CRITICAL Issues:**
- **Massive left sidebar empty space**: 30-40% of horizontal screen space wasted on the WL selector container
  - WL selector (150px width) is barely filled while array gets squeezed
  - Sidebar becomes increasingly wasteful as array grows beyond 8×8
  - Example: 32×32 array needs ~450px but only gets ~250px due to sidebar allocation

- **Mode controls buried at bottom**: Operation mode buttons (READ/WRITE/COMPUTE) are in the top mode bar but easily missed
  - Controls should be more prominent for discoverability
  - Top positioning is good but could be more visually integrated with operation name

- **Write slider hidden at bottom edge**: The write level slider (mfuxWriteLevelSlider) is in the write mode panel
  - Primary control for WRITE mode is not always visible without scrolling
  - Should be in top toolbar for guaranteed visibility

- **Cell details illegible at 32x32+ arrays**: sharedCellInfoLabel shows "Cell [r,c]: State N | G=XXµS | BL=X.XXV | Material"
  - Text overlaps with array or is too small to read
  - Cell coordinates become hard to interpret with large grids

**HIGH Priority Issues:**
- **WL labels take too much vertical space**: "Row 0", "Row 1", ... labels in the checkbox list
  - Each checkbox with label adds 30-35px height
  - For 8 rows, that's 240+ px of vertical space used just for labels
  - Could be replaced with compact numbering or icon indicators

- **Missing legend**: Color meaning unclear without reference
  - Users cannot determine what blue/gray/red gradient represents without documentation
  - Should show level mapping (blue=low conductance, gray=mid, red=high)

- **Architecture mode change weak feedback**: Architecture buttons (PASSIVE/1T1R/2T1R) don't clearly show active state
  - Should use same button importance highlighting as mode buttons
  - Current state requires close inspection

- **MVM output not displayed**: Compute operation result (row currents, ADC levels) not shown
  - sharedArrayInfoLabel shows array dimensions only
  - Should display compute results or provide separate output panel

- **Control panel grouping flat (no visual hierarchy)**: Too many controls at same visual level
  - Mode bar, DAC section, action buttons all blend together
  - Difficult to understand what controls work together
  - Separator line between signal chain and mode bar helps, but more structure needed

### 2. Recommended New Layout (Option B: Top-Heavy Toolbar)

The following layout moves the mode controls and key inputs to a unified top toolbar, freeing vertical space and improving discoverability.

**Layout Structure**:
```
Border
├─ Top: VBox (toolbarSection) ~100px
│  ├─ configRow: HBox [Material, ArraySize, ADCBits, |, READ, WRITE, COMPUTE, |, PASSIVE, 1T1R, 2T1R]
│  │  ├─ Label: "Config:" (bold)
│  │  ├─ materialSelector: Dropdown [FeCIM HZO, HZO (Si-doped), ...]
│  │  ├─ Spacer (small)
│  │  ├─ Label: "Array:"
│  │  ├─ arraySizeSelector: Dropdown [8×8, 16×16, 32×32, 64×64]
│  │  ├─ Spacer (small)
│  │  ├─ Label: "ADC Bits:"
│  │  ├─ adcBitsSpinner: Spinner control [4-12 bits]
│  │  ├─ Separator (vertical line)
│  │  ├─ modeReadBtn: "READ"
│  │  ├─ modeWriteBtn: "WRITE"
│  │  ├─ modeComputeBtn: "COMPUTE"
│  │  ├─ Separator (vertical line)
│  │  ├─ archPassiveBtn: "PASSIVE"
│  │  ├─ arch1T1RBtn: "1T1R"
│  │  └─ arch2T1RBtn: "2T1R"
│  ├─ modePanelStack: Stack [writeModePanel, computeModePanel, emptyPanel] ~35px
│  │  ├─ writeModePanel: HBox
│  │  │  ├─ Label: "Write Level:" (bold)
│  │  │  ├─ mfuxWriteLevelSlider (0-29, wide)
│  │  │  ├─ mfuxWriteLevelLabel: "Level: 15"
│  │  │  ├─ mfuxWriteVoltageLabel: "Voltage: 1.20V"
│  │  │  └─ Spacer
│  │  └─ computeModePanel: HBox
│  │     ├─ Label: "Input Vector (0-255):" (bold)
│  │     ├─ HBox (8 input entries, compact)
│  │     │  ├─ Entry (x0)
│  │     │  ├─ Entry (x1)
│  │     │  └─ ... x7
│  │     └─ HBox buttons [Random, Clear]
│  └─ actionRow: HBox ~30px
│     ├─ dacRangeLabel: "DAC: Read (0-0.5V)"
│     ├─ Spacer
│     ├─ writeBtn: "Write"
│     ├─ mvmBtn: "MVM"
│     ├─ Separator (vertical line)
│     ├─ undoBtn: "Undo"
│     ├─ randomBtn: "Random"
│     ├─ resetBtn: "Reset"
│     ├─ Separator (vertical line)
│     ├─ toolsBtn: "Tools▼" (dropdown: Export, Compare, Settings)
│     └─ Spacer
├─ Center: VBox (arraySection) EXPANDS
│  ├─ UnifiedTappableCanvas (flexible size, maximum vertical space)
│  └─ infoRow: HBox ~20px [sharedCellInfoLabel, Spacer, sharedArrayInfoLabel]
│     ├─ sharedCellInfoLabel: "Cell [r,c]: State N | G=XXµS"
│     ├─ Spacer
│     ├─ sharedArrayInfoLabel: "Array: 8×8 | 30 levels | MVM: [Y0=X, Y1=X, ...]"
│     └─ Spacer
└─ Bottom: HBox (statusBar) ~20px
   └─ operationsModeHelp: "READ mode: Single row, safe voltage (0-0.5V)"
```

**Key Changes from Current Layout**:

1. **Material, ArraySize, ADCBits moved to top configRow**
   - Currently in separate locations or missing from UI
   - Consolidation improves visual grouping

2. **Mode buttons (READ/WRITE/COMPUTE) promoted to top toolbar**
   - Currently in "Mode:" section below signal chain header
   - Top position improves discoverability and reduces scrolling

3. **Architecture buttons (PASSIVE/1T1R/2T1R) moved to configRow**
   - Currently separate from mode controls
   - Grouping shows they work together

4. **Write level slider integrated into toolbar**
   - Currently in writeModePanel which can be hidden
   - Top placement ensures visibility and quick access

5. **Input vector entries in actionRow (compact layout)**
   - Currently in computeModePanel which can be hidden
   - More compact representation (8 narrow fields instead of tall column)

6. **Array canvas gets Center position (EXPANDS)**
   - Currently constrained by left sidebar (HSplit 10%/90%)
   - Full horizontal space allows larger cell visualization

7. **WL selector removed entirely**
   - Functionality preserved: all WL modes configured by operation mode buttons
   - READ mode = single WL selected by clicking cell
   - WRITE mode = single WL selected by clicking cell
   - COMPUTE mode = all WLs active (automatic)
   - No need for explicit WL checkboxes

8. **DAC "Set All (V)" entry removed from toolbar**
   - Manual voltage override replaced with mode-based presets
   - Reduces control clutter

9. **MVM output displayed in infoRow**
   - Shows computed currents/levels after COMPUTE operation
   - Example: "MVM: [Y0=85, Y1=120, ...]" or similar

10. **Status bar simplified to single operationsModeHelp label**
    - Reduces visual clutter at bottom
    - Essential info only

### 3. Implementation Changes Required

Files to modify:

- **tab_unified.go** - Restructure `createUnifiedView()` layout
  - Current: Top (signal chain + mode bar + DAC) → Center (HSplit sidebar + array) → Bottom (actions)
  - New: Top (toolbar) → Center (array with info) → Bottom (status)
  - Remove HSplit layout for WL selector
  - Consolidate mode/architecture/config controls into single toolbar VBox

- **tab_unified_canvas.go** - Remove fixed MinSize constraint
  - Current: `MinSize(850, 550)` forces large minimum canvas
  - New: Remove or reduce to `MinSize(600, 400)` to allow flexible scaling
  - Canvas should expand to fill available Center space

- **device_state.go** - No changes needed
  - Existing OpMode/WL logic works with new UI
  - WL mode still set by mode button clicks

### 4. Benefits of New Layout

- **Better space utilization**: Array canvas gets maximum vertical/horizontal space
- **Improved discoverability**: Mode/architecture controls prominent at top
- **Reduced scrolling**: All primary controls visible without scrolling
- **Better visual hierarchy**: Toolbar organizes controls into logical sections (Config, Mode, Action)
- **Larger cell visualization**: 32×32+ arrays more readable
- **Single entry point**: No hidden/nested WL controls to discover
- **Professional appearance**: Top toolbar matches modern application design patterns (browsers, IDEs, etc.)

---

## File Structure

| File | Purpose |
|------|---------|
| `app.go` | Main CircuitsApp struct, window setup, view switching |
| `device_state.go` | DeviceState struct, voltage ranges, simulation logic |
| `tab_unified.go` | Unified device simulation view (replaces tab_operations.go) |
| `tab_comparison.go` | FeFET vs GPU vs CPU comparison view |
| `tab_reference.go` | Timing diagrams and specifications reference |
| `tab_reference_timing.go` | Timing diagram drawing functions |
| `tab_reference_specs.go` | Specifications section |
| `drawing.go` | Primitive drawing functions |
| `helpers.go` | DAC/TIA/ADC box drawing helpers |
| `font.go` | Bitmap font patterns for canvas text |
| `embedded.go` | Embedded interface for main app integration |

---

## Screens

### Main Window (app.go:274-367)
**Purpose**: Top-level window with 3-view architecture
**File**: app.go:274-367
**State**: window, fyneApp, deviceState
**Layout**:
```
Border
├─ Top: Header with inline view selector
│  ├─ Label: "View:"
│  ├─ Select: ["OPERATIONS", "COMPARISON", "REFERENCE"]
│  ├─ Spacer
│  └─ Label: "3 Views | DAC -> FeFET -> TIA -> ADC"
├─ Bottom: Footer
│  └─ Label: "FeCIM Ferroelectric Compute-in-Memory | Based on Published Research"
└─ Center: Stack container
   ├─ OPERATIONS view (visible on start) - from tab_unified.go
   ├─ COMPARISON view (hidden)
   └─ REFERENCE view (hidden)
```

---

### OPERATIONS View (tab_unified.go:29-76)
**Purpose**: Unified device simulation with Mode-First UX
**File**: tab_unified.go:29-76
**State**: deviceState (OpMode, VoltageRange, material, activeRows, dacVoltages)
**Layout**:
```
Border
├─ Top: VBox
│  ├─ Signal chain header
│  │  ├─ HBox
│  │  │  ├─ Label: "SIGNAL CHAIN: DAC -> Array -> TIA -> ADC" (bold)
│  │  │  ├─ Spacer
│  │  │  ├─ Material selector: [FeCIM HZO, HZO (Si-doped), ...]
│  │  │  ├─ Spacer
│  │  │  ├─ Architecture toggle: [PASSIVE] [1T1R] [2T1R]
│  │  │  ├─ Spacer
│  │  │  └─ operationsStatusLabel
│  │  ├─ operationsModeHelp (mode + architecture help text)
│  │  └─ Separator
│  ├─ Mode bar (Mode-First UX)
│  │  ├─ HBox
│  │  │  ├─ Label: "Mode:" (bold)
│  │  │  ├─ modeReadBtn: "READ" (highlighted when active)
│  │  │  ├─ modeWriteBtn: "WRITE"
│  │  │  ├─ modeComputeBtn: "COMPUTE"
│  │  │  └─ Spacer
│  ├─ Mode panels (Stack - only one visible)
│  │  ├─ writeModePanel (shown in WRITE mode)
│  │  │  ├─ Label: "Target Write Level:" (bold)
│  │  │  └─ Border
│  │  │     ├─ Left: mfuxWriteLevelLabel "Level: 15"
│  │  │     ├─ Right: mfuxWriteVoltageLabel "Voltage: 1.20V"
│  │  │     └─ Center: mfuxWriteLevelSlider (0-29)
│  │  └─ computeModePanel (shown in COMPUTE mode)
│  │     ├─ Label: "Input Vector (0-255):" (bold)
│  │     ├─ HBox (8 entry columns)
│  │     │  ├─ VBox: [x0 label, entry]
│  │     │  ├─ VBox: [x1 label, entry]
│  │     │  └─ ... x7
│  │     └─ HBox: [Random] [Clear]
│  └─ DAC section
│     ├─ HBox
│     │  ├─ dacRangeLabel: "DAC: Read (0-0.5V)" (auto-updated by mode)
│     │  ├─ Spacer
│     │  ├─ Label: "Set All (V):"
│     │  └─ Entry: allEntry (manual voltage override)
├─ Bottom: Action buttons
│  ├─ HBox
│  │  ├─ Button: "Write Cell" (HighImportance)
│  │  ├─ Button: "Sense Row"
│  │  ├─ Button: "Compute MVM"
│  │  ├─ Spacer
│  │  ├─ Button: "Animate"
│  │  ├─ Button: "Random Array"
│  │  └─ Button: "Reset Array"
└─ Center: HSplit (10% WL selector, 90% array)
   ├─ Left: Word Line selector
   │  ├─ Label: "WORD LINES" (bold)
   │  └─ WL checkboxes: WL0, WL1, ... WL7
   └─ Right: Array visualization
      ├─ UnifiedTappableCanvas (400x350px)
      │  └─ Raster: drawUnifiedArray
      │     ├─ DAC boxes (top, per column, voltage-colored)
      │     ├─ WL lines (horizontal, orange=active, dim=inactive)
      │     ├─ BL lines (vertical, red=write, blue=read, dim=off)
      │     ├─ Cell grid (color-coded by conductance level)
      │     ├─ TIA+ADC boxes (right, per row)
      │     ├─ 1T1R/2T1R transistors (if active architecture)
      │     ├─ Operation label (top-left): "READ", "WRITE", "COMPUTE (MVM)"
      │     └─ Architecture badge (top-right): "PASSIVE", "1T1R", "2T1R"
      ├─ legendLabel
      ├─ sharedCellInfoLabel: "Cell [r,c]: State N | G=XXµS | BL=X.XXV | Material"
      └─ sharedArrayInfoLabel: "Array: 8x8 | 30 levels"
```

---

### DeviceState (device_state.go)
**Purpose**: Central state management for device simulation
**File**: device_state.go
**Key Fields**:
```go
type DeviceState struct {
    // Dimensions
    rows, cols int

    // Operation mode (READ/WRITE/COMPUTE)
    opMode OpMode

    // WL configuration
    wlMode     WLMode      // WLSingle, WLAll, WLCustom
    activeRows []bool      // true = WL HIGH
    isPassive  bool        // When true, ALL WLs always on (0T1R architecture)

    // DAC inputs
    dacVoltages  []float64   // Voltage per column
    dacMode      DACMode     // DACReadPreset, DACWritePreset, etc.
    dacRangeMode DACRangeMode // DACRangeRead, DACRangeWrite

    // Voltage ranges (derived from material + physics.yaml)
    readRange   VoltageRange  // 0 to FieldMinRatio*Vc
    writeRange  VoltageRange  // Vc to FieldMaxRatio*Vc
    calibParams CalibrationParams // From physics.yaml

    // Computed outputs
    rowCurrents []float64   // TIA input (µA)
    rowVoltages []float64   // TIA output (V)
    rowLevels   []int       // ADC output
    saturated   []bool

    // Selection
    selectedRow, selectedCol int

    // Physics
    material *ferroelectric.HZOMaterial
    tia      *peripherals.TIA
    adc      *peripherals.ADC
}
```

**Key Methods**:
| Method | Purpose |
|--------|---------|
| `NewDeviceState(rows, cols, tia, adc)` | Create with dimensions and peripherals |
| `SetMaterial(mat)` | Change material, recalculates voltage ranges |
| `SetOperationMode(mode)` | Set READ/WRITE/COMPUTE mode |
| `SetWLSingle(row)` | Activate only specified row (ignored if isPassive=true) |
| `SetWLAll()` | Activate all rows for MVM |
| `SetPassiveMode(passive)` | Enable/disable passive mode enforcement |
| `SetDACPreset(preset, params...)` | Apply voltage preset |
| `SetDACVoltageForState(col, level)` | Set write voltage for target state |
| `Compute(weights, levels)` | Run MVM simulation |
| `GetReadRange() / GetWriteRange()` | Get voltage ranges for current material |
| `ClassifyOperation()` | Get operation name string |
| `AdvanceWritePhase()` | Move to next phase in 4-phase write sequence |
| `GetWritePhaseInfo()` | Get current write sequence state for UI |
| `StartISPP(row, col, target, current)` | Begin ISPP loop for cell programming |
| `ISPPIterate(newLevel)` | Perform one ISPP iteration, returns result |
| `GetISPPStatus()` | Get current ISPP state for UI display |
| `CancelISPP()` | Cancel active ISPP operation |
| `EnableHalfSelectVisualization(row, col, V)` | Enable V/2 overlay for write |
| `DisableHalfSelectVisualization()` | Disable V/2 overlay |
| `IsHalfSelected(row, col)` | Check if cell is in half-select state |

---

### 4-Phase Write Sequence (device_state.go:772-908)

**Purpose**: Implements proper ferroelectric write timing for reliable cell programming.

**Phases**:
```
PhaseIdle → PhaseReset → PhaseHold1 → PhaseWrite → PhaseHold2 → PhaseIdle
    │           │            │            │            │
    │        ~10ns        ~5ns        ~50ns        ~5ns
    └─────────────────────────────────────────────────────┘
```

**WritePhase Enum**:
```go
type WritePhase int
const (
    PhaseIdle   WritePhase = iota  // No write in progress
    PhaseReset                      // Apply reset pulse (opposite polarity)
    PhaseHold1                      // Stabilization after reset
    PhaseWrite                      // Apply write pulse (target polarity)
    PhaseHold2                      // Stabilization after write
)
```

**WriteSequenceState Struct**:
```go
type WriteSequenceState struct {
    Active       bool        // True if sequence in progress
    Phase        WritePhase  // Current phase
    TargetRow    int         // Row being written
    TargetCol    int         // Column being written
    TargetLevel  int         // Target conductance level
    CurrentLevel int         // Current cell level
}
```

**UI Integration**: Animated timing diagram shows phase progression with color-coded pulses.

---

### ISPP State Machine (device_state.go:910-1068)

**Purpose**: Incremental Step Pulse Programming for precise cell programming with overshoot detection.

**Algorithm**:
1. Start with initial voltage estimate for target level
2. Apply write pulse, read back actual level
3. If level matches target → Success
4. If level overshoots → Reset to saturation, restart
5. If level undershoots → Increment voltage, repeat
6. After max iterations → Report max-iter failure

**ISPPResult Enum**:
```go
type ISPPResult int
const (
    ISPPContinue  ISPPResult = iota  // Need more iterations
    ISPPSuccess                       // Target level reached
    ISPPOvershoot                     // Overshot target, needs reset
    ISPPMaxIter                       // Max iterations reached
    ISPPCancelled                     // User cancelled
)
```

**ISPPState Struct**:
```go
type ISPPState struct {
    Active       bool       // True if ISPP loop running
    Iteration    int        // Current iteration (0-9)
    MaxIter      int        // Maximum iterations (default: 10)
    Direction    int        // +1 ascending, -1 descending
    TargetRow    int
    TargetCol    int
    TargetLevel  int
    LastResult   ISPPResult
}
```

**Constants**:
- `ISPPMaxIterations = 10` - Maximum pulse iterations
- `ISPPVoltageStep = 0.05` - Voltage increment per iteration (V)

---

### V/2 Half-Select Visualization (device_state.go:1071-1165)

**Purpose**: Visualize half-select disturb risk in passive (0T1R) crossbar arrays.

**Physics**: In passive arrays without transistor isolation:
- Selected WL = +V/2, selected BL = −V/2 (symmetric half‑select)
- Selected cell sees full write voltage ( +V/2 − (−V/2) = V )
- Cells in same row or column see ±V/2 relative to ground
- Unselected cells remain at 0V
- These V/2 cells may experience disturb over time

**HalfSelectVisualization Struct**:
```go
type HalfSelectVisualization struct {
    Enabled        bool      // True when visualization active
    TargetRow      int       // Selected cell row
    TargetCol      int       // Selected cell column
    FullVoltage    float64   // Voltage on target cell
    HalfVoltage    float64   // V/2 voltage on half-selected cells
    HalfSelectRows []int     // Rows with V/2 (same column)
    HalfSelectCols []int     // Columns with V/2 (same row)
}
```

**UI Colors**:
- **Gold** (`#FFD700`): Target cell receiving full write voltage
- **Amber** (`#FFBF00`): Half-selected cells receiving V/2

**Constant**: `HalfSelectVoltageRatio = 0.5`

---

### Voltage Range Configuration

Voltage ranges are derived from physics.yaml and material properties:

```yaml
# config/physics.yaml
calibration:
  field_min_ratio: 0.7   # Read max = 0.7 * Vc
  field_max_ratio: 2.5   # Write max = 2.5 * Vc
```

**Calculation**:
```go
func (ds *DeviceState) updateVoltageRanges() {
    Vc := ds.material.CoerciveVoltage()  // Vc = Ec × thickness

    // Read range: 0 to FieldMinRatio * Vc (safe, non-destructive)
    safeReadMax := ds.calibParams.FieldMinRatio * Vc

    // Write range: Vc to FieldMaxRatio * Vc (exceeds coercive)
    writeMin := Vc
    writeMax := ds.calibParams.FieldMaxRatio * Vc
}
```

---

### Operation Mode Rules (from docs/peripheral-circuits/circuits.operations.md)

| Mode | WL Selection | DAC Voltage | Effect |
|------|--------------|-------------|--------|
| READ | Single row | 0 to read max | Sense conductance, no change |
| WRITE | Single row | Vc to write max | Program cell state |
| COMPUTE | All rows | 0 to read max (input vector) | MVM multiply, no change |

**Mode Button Behavior** (`setOperationMode()`):
```go
switch mode {
case OpModeRead:
    ca.deviceState.SetWLSingle(selectedRow)
    ca.deviceState.SetDACPreset(DACReadPreset)
case OpModeWrite:
    ca.deviceState.SetWLSingle(selectedRow)
    ca.deviceState.SetDACPreset(DACWritePreset)
case OpModeCompute:
    ca.deviceState.SetWLAll()
    // Keep read range voltages for compute
}
```

---

### Components

| Component | Type | Purpose | File:Line | State |
|-----------|------|---------|-----------|-------|
| UnifiedTappableCanvas | Custom Widget | Clickable array with DAC/TIA/ADC | tab_unified.go:460-538 | sharedArrayCanvas |
| modeReadBtn | Button | Set READ mode (top mode bar) | tab_unified.go:1432 | opMode |
| modeWriteBtn | Button | Set WRITE mode (top mode bar) | tab_unified.go:1435 | opMode |
| modeComputeBtn | Button | Set COMPUTE mode (top mode bar) | tab_unified.go:1438 | opMode |
| mfuxWriteLevelSlider | Slider | Target write level (0 to quantLevels-1) | tab_unified.go:1459 | via SetDACVoltageForState() |
| mfuxWriteLevelLabel | Label | Shows "Level: N" | tab_unified.go:1466 | Updated by onWriteLevelChanged() |
| mfuxWriteVoltageLabel | Label | Shows "Voltage: X.XXV" | tab_unified.go:1468 | Updated by onWriteLevelChanged() |
| mfuxWriteTargetLabel | Label | Shows "Target: Row X, Col Y" | tab_unified.go | Updated by cell selection |
| undoHistoryBtn | Button | Undo last array change | tab_unified.go | Enabled when hasUndoHistory=true |
| mfuxInputVectorEntry | []Entry | 8 input vector entries (0-255) | tab_unified.go:1503 | inputVector via onInputVectorEntryChanged() |
| writeModePanel | Container | Write mode controls (slider) | tab_unified.go:45 | Show/Hide via updateModePanels() |
| computeModePanel | Container | Compute mode controls (entries) | tab_unified.go:49 | Show/Hide via updateModePanels() |
| dacRangeLabel | Label | Shows current DAC range mode (auto-updated by operation mode) | tab_unified.go:147 | dacRangeMode |
| materialSelector | Select | Choose ferroelectric material | tab_unified.go:97-122 | material |
| archPassiveBtn | Button | Select passive (0T1R) architecture | tab_unified.go:1356 | architecture |
| arch1T1RBtn | Button | Select 1T1R architecture | tab_unified.go:1357 | architecture |
| arch2T1RBtn | Button | Select 2T1R architecture | tab_unified.go:1358 | architecture |
| operationsModeHelp | Label | Mode + architecture help text | tab_unified.go:79 | Updated by updateOperationClassification() |
| writePhaseCanvas | Raster | 4-phase timing diagram animation | tab_unified.go | WriteSequenceState |
| isppStatusLabel | Label | Shows "ISPP: Iter N/10, Direction: ↑" | tab_unified.go | ISPPState |
| isppStartBtn | Button | Start ISPP programming loop | tab_unified.go | Triggers StartISPP() |
| isppCancelBtn | Button | Cancel active ISPP operation | tab_unified.go | Triggers CancelISPP() |
| halfSelectOverlay | Canvas Layer | V/2 visualization (gold/amber) | tab_unified.go | HalfSelectVisualization |

---

### Data Flow

| Trigger | Source | Updates | File |
|---------|--------|---------|------|
| Mode button click | modeReadBtn/modeWriteBtn/modeComputeBtn | opMode, WL config, DAC config, button highlighting, mode panels | tab_unified.go:314-352 |
| Write level slider | mfuxWriteLevelSlider | DAC voltage via SetDACVoltageForState(), level/voltage labels | tab_unified.go:1489-1507 |
| Input vector entry | mfuxInputVectorEntry[i] | inputVector[i], DAC voltages via SetDACPreset(DACInputVector) | tab_unified.go:1527-1551 |
| Material selection | materialSelector | material, voltage ranges | tab_unified.go:104-115 |
| Architecture change | archPassiveBtn/arch1T1RBtn/arch2T1RBtn | architecture, WL state, transistor display | tab_unified.go:1384-1424 |
| Cell click | UnifiedTappableCanvas.Tapped() | selectedRow, selectedCol, WL (if single mode) | tab_unified.go:1039-1051 |
| Write Cell button | programBtn | arrayWeights[row][col], starts 4-phase sequence | tab_unified.go:1058-1089 |
| Compute MVM button | computeBtn | WL all, recompute | tab_unified.go:1103-1110 |
| ISPP Start | isppStartBtn | StartISPP(), begins iteration loop | tab_unified.go |
| ISPP Iterate | Timer/callback | ISPPIterate(), updates status | tab_unified.go |
| Write phase advance | Timer (500ms) | AdvanceWritePhase(), updates timing diagram | tab_unified.go |
| Half-select enable | Write mode entry | EnableHalfSelectVisualization() | tab_unified.go |

---

### COMPARISON View (tab_comparison.go:20-71)
**Purpose**: Compare FeFET vs GPU vs CPU architectures
**File**: tab_comparison.go:20-71
**State**: compArraySize (8, 16, 32, 64)
*(Layout unchanged from previous version)*

---

### REFERENCE View (tab_reference.go:25-53)
**Purpose**: Timing diagrams + specifications reference
**File**: tab_reference.go:25-53
**State**: timingOpSelect, specArraySizeSelect
*(Layout unchanged from previous version)*

---

## State Machine

### OpMode State Transitions
```
Initial State: OpModeRead

OpModeRead
  └─> "WRITE" button clicked -> OpModeWrite
  └─> "COMPUTE" button clicked -> OpModeCompute

OpModeWrite
  └─> "READ" button clicked -> OpModeRead
  └─> "COMPUTE" button clicked -> OpModeCompute

OpModeCompute
  └─> "READ" button clicked -> OpModeRead
  └─> "WRITE" button clicked -> OpModeWrite
```

**Actions on Mode Change**:
1. Update `deviceState.opMode`
2. Configure WL (single for READ/WRITE, all for COMPUTE)
3. Configure DAC preset (read vs write range)
4. Update mode button highlighting
5. **Show/hide mode panels** (writeModePanel for WRITE, computeModePanel for COMPUTE)
6. Update WL checkboxes
7. Update DAC range label
8. Refresh array canvas
9. Update operation classification help text

---

## Key Patterns

### 1. Unified Tappable Canvas Pattern
```go
type UnifiedTappableCanvas struct {
    widget.BaseWidget
    raster *canvas.Raster
    onTap  func(row, col int)
    ca     *CircuitsApp
}

func (t *UnifiedTappableCanvas) Tapped(e *fyne.PointEvent) {
    // Convert screen coordinates to grid coordinates
    col := (int(e.Position.X) - offsetX) / cellSize
    row := (int(e.Position.Y) - offsetY) / cellSize
    t.onTap(row, col)
}
```

### 2. Material-Derived Voltage Ranges
```go
// No hardcoded values - all derived from material + config
Vc := material.CoerciveVoltage()  // Ec × thickness
readMax := calibParams.FieldMinRatio * Vc
writeMax := calibParams.FieldMaxRatio * Vc
```

### 3. Mode Button Highlighting
```go
func (ca *CircuitsApp) updateModeButtons() {
    // Reset all to low importance
    ca.modeReadBtn.Importance = widget.LowImportance
    ca.modeWriteBtn.Importance = widget.LowImportance
    ca.modeComputeBtn.Importance = widget.LowImportance

    // Highlight active mode
    switch ca.deviceState.GetOperationMode() {
    case OpModeRead:
        ca.modeReadBtn.Importance = widget.HighImportance
    case OpModeWrite:
        ca.modeWriteBtn.Importance = widget.HighImportance
    case OpModeCompute:
        ca.modeComputeBtn.Importance = widget.HighImportance
    }
}
```

### 4. Automatic DAC Range Updates (Mode-Based)
```go
// DAC ranges are automatically set by operation mode (no separate buttons)
// This follows VOLTAGE_RULES.md: mode dictates voltage ranges
func (ca *CircuitsApp) setOperationMode(mode OperationMode) {
    switch mode {
    case OpModeRead:
        ca.deviceState.SetDACPreset(DACReadPreset)    // 0-0.5V
    case OpModeWrite:
        ca.deviceState.SetDACPreset(DACWritePreset)   // 1.2-1.5V
    case OpModeCompute:
        if ca.deviceState.GetDACMode() == DACWritePreset {
            ca.deviceState.SetDACPreset(DACReadPreset) // Keep safe range
        }
    }
    ca.updateDACRangeModeLabel() // Update "DAC: Read/Write (X-YV)"
}
```

### 5. Architecture-Aware WL Handling (Defense-in-Depth)
```go
// UI Layer: Disable checkboxes in passive mode
func (ca *CircuitsApp) updateWLCheckboxes() {
    isPassive := ca.architecture == sharedwidgets.Architecture0T1R
    for i, check := range ca.unifiedWLChecks {
        if isPassive {
            check.SetChecked(true)
            check.Disable()  // User cannot uncheck
        } else {
            check.Enable()
            check.SetChecked(isActive)
        }
    }
}

// Data Layer: Ignore WL changes when passive
func (ds *DeviceState) SetWLSingle(row int) {
    if ds.isPassive {
        return  // Passive mode: all WLs always on, ignore
    }
    // ... normal logic
}
```

### 5b. Undo History Pattern
```go
func (ca *CircuitsApp) saveUndoHistory() {
    ca.mu.Lock()
    ca.undoHistory = make([][]int, len(ca.arrayWeights))
    for i := range ca.arrayWeights {
        ca.undoHistory[i] = make([]int, len(ca.arrayWeights[i]))
        copy(ca.undoHistory[i], ca.arrayWeights[i])
    }
    ca.hasUndoHistory = true
    ca.mu.Unlock()
    fyne.Do(func() {
        if ca.undoHistoryBtn != nil { ca.undoHistoryBtn.Enable() }
    })
}
```

### 6. Mode-First UX Panel Show/Hide
```go
func (ca *CircuitsApp) updateModePanels(mode OpMode) {
    fyne.Do(func() {
        // Hide all panels first
        ca.writeModePanel.Hide()
        ca.computeModePanel.Hide()

        // Show relevant panel based on mode
        switch mode {
        case OpModeWrite:
            ca.writeModePanel.Show()
        case OpModeCompute:
            ca.computeModePanel.Show()
        // OpModeRead: clean view, no extra panel
        }
    })
}
```

### 7. Write Level Slider to DAC Voltage
```go
func (ca *CircuitsApp) onWriteLevelChanged(level int) {
    // Use material-derived voltage mapping
    ca.deviceState.SetDACVoltageForState(selectedCol, level)
    voltage := ca.deviceState.GetDACVoltage(selectedCol)
    ca.mfuxWriteLevelLabel.SetText(fmt.Sprintf("Level: %d", level))
    ca.mfuxWriteVoltageLabel.SetText(fmt.Sprintf("Voltage: %.2fV", voltage))
}
```

### 8. Input Vector to DAC Voltages
```go
func (ca *CircuitsApp) onInputVectorEntryChanged(col int, valueStr string) {
    value, _ := strconv.Atoi(valueStr)  // 0-255
    ca.inputVector[col] = value

    // Convert all inputs to DAC voltages (0-255 -> read range)
    params := make([]float64, len(ca.inputVector))
    for i, v := range ca.inputVector {
        params[i] = float64(v)
    }
    ca.deviceState.SetDACPreset(DACInputVector, params...)
}
```

### 9. Level-to-Color Mapping (Blue-Gray-Red)
```go
func levelToColor(level, maxLevel int) color.RGBA {
    mid := (maxLevel - 1) / 2
    if level < mid {
        // Below mid: Blue gradient (low conductance)
        t := float64(level) / float64(mid)
        r = uint8(40 + t*40)
        g = uint8(60 + t*60)
        b = uint8(180 + t*40)
    } else if level > mid {
        // Above mid: Red gradient (high conductance)
        t := float64(level-mid) / float64(maxLevel-1-mid)
        r = uint8(180 + t*75)
        g = uint8(100 - t*60)
        b = uint8(80 - t*40)
    } else {
        // At mid: Gray (neutral state)
        r, g, b = 140, 140, 150
    }
    return color.RGBA{r, g, b, 255}
}
```

### 10. Cell Selection Visual (Border Only)
```go
// Draw gold border for selected cell - NO FILL (preserves state color)
if r == selectedRow && c == selectedCol {
    borderColor := color.RGBA{255, 215, 0, 255}  // Gold
    for b := 0; b < 2; b++ {  // 2-pixel border
        // Draw top, bottom, left, right border lines
        // Cell interior retains levelToColor() result
    }
}
```

---

## Error Handling

### Voltage Range Validation

All voltage inputs are validated within material-derived bounds:

```go
// Column bounds checking (device_state.go:295-298)
func (ds *DeviceState) SetDACVoltage(col int, voltage float64) {
    if col >= 0 && col < ds.cols {
        ds.dacVoltages[col] = voltage
        ds.dacMode = DACManual
    }
    // Silently ignores out-of-bounds column
}

// Target state clamping for write level (device_state.go:364-375)
func (ds *DeviceState) SetDACVoltageForState(col int, targetState int) {
    if col < 0 || col >= ds.cols {
        return  // Out-of-bounds column ignored
    }

    // Clamp target state to 0..NumLevels-1
    if targetState < 0 {
        targetState = 0
    }
    if targetState >= ds.writeRange.NumLevels {
        targetState = ds.writeRange.NumLevels - 1
    }
}

// Voltage clamping in preset application (device_state.go:314-335)
case DACReadPreset:
    voltage := ds.readRange.Max * 0.5
    if len(params) > 0 {
        voltage = params[0]
    }
    if voltage > ds.readRange.Max {
        voltage = ds.readRange.Max  // Clamp to safe read range
    }

case DACWritePreset:
    writeVoltage := (ds.writeRange.Min + ds.writeRange.Max) / 2
    if writeVoltage < ds.writeRange.Min {
        writeVoltage = ds.writeRange.Min
    }
    if writeVoltage > ds.writeRange.Max {
        writeVoltage = ds.writeRange.Max
    }
```

### Fallback Behavior for physics.yaml

Configuration loading gracefully handles missing/invalid physics.yaml:

```go
// device_state.go:76-89
func loadCalibrationParams() CalibrationParams {
    cfg, err := physics.Load()
    if err != nil || cfg == nil {
        // Fallback: field_min_ratio=0.7, field_max_ratio=2.5
        // These are standard values from typical physics.yaml
        return CalibrationParams{
            FieldMinRatio: 0.7,
            FieldMaxRatio: 2.5,
        }
    }
    return CalibrationParams{
        FieldMinRatio: cfg.Calibration.FieldMinRatio,
        FieldMaxRatio: cfg.Calibration.FieldMaxRatio,
    }
}
```

**Fallback Action**: If physics.yaml is unavailable or corrupted:
- Read max voltage: 0.5 × Vc (safe non-destructive reading)
- Write max voltage: 2.5 × Vc (standard overwrite window)
- System remains fully functional without explicit error dialogs

### Array Bounds Checking

All array access validates indices before operation:

```go
// Row/column validation in cell operations (device_state.go:395-401)
func (ds *DeviceState) SetSelectedCell(row, col int) {
    // No explicit validation - relies on slice bounds
    // GUI ensures valid indices via click coordinate conversion
    ds.selectedRow = row
    ds.selectedCol = col
}

// Column validation in Compute (device_state.go:424-426)
for c := 0; c < ds.cols; c++ {
    level := 0
    if r < len(weights) && c < len(weights[r]) {
        level = weights[r][c]  // Safe access to weights matrix
    }
    // Proceeds with level=0 if weights unavailable
}
```

### Invalid Mode Transition Handling

Mode transitions are always safe - no invalid state transitions exist:

```go
// State transitions (tab_unified.go:314-352)
// All mode buttons can transition to any other mode
switch mode {
case OpModeRead:
    ca.deviceState.SetWLSingle(selectedRow)      // Valid
    ca.deviceState.SetDACPreset(DACReadPreset)   // Valid
case OpModeWrite:
    ca.deviceState.SetWLSingle(selectedRow)      // Valid
    ca.deviceState.SetDACPreset(DACWritePreset)  // Valid
case OpModeCompute:
    ca.deviceState.SetWLAll()                    // Valid
    // Keeps read range for compute-safe operation
}

// State machine always remains consistent
```

### Material Selection Error Handling

Material selection gracefully handles fallback:

```go
// Device_state.go:403-436 Compute function
var conductanceS float64
if ds.material != nil {
    conductanceS = ds.material.DiscreteLevel(level, quantLevels)
} else {
    // Fallback: linear mapping 1-100 µS
    conductanceS = (1.0 + float64(level)/float64(quantLevels-1)*99.0) * 1e-6
}
```

If material selector fails to load or material is nil:
- Fallback uses simple linear conductance model (1-100 µS)
- Simulation continues without explicit error
- UI material selector initialized with AllMaterials() from ferroelectric package

---

## Testing

### Unit Test Strategy for DeviceState Methods

Test structure follows `docs/development/TESTING.md` patterns:

```bash
# Run peripheral circuit tests (includes device state logic)
go test -v ./shared/peripherals

# Run GUI tests (headless widget tests)
go test -v ./module4-circuits/pkg/gui
```

**DeviceState Method Testing Approach**:
| Method | Test Category | Validation |
|--------|---------------|-----------|
| SetDACVoltage() | Bounds checking | Verify col in [0, cols), voltage stored |
| SetDACVoltageForState() | State-to-voltage mapping | Level clamping, linear interpolation |
| SetWLSingle(row) | WL mode switching | activeRows[row]=true, others false |
| SetWLAll() | WL mode switching | All activeRows=true |
| SetOperationMode() | Mode transitions | OpMode updated, WL/DAC reconfigured |
| Compute() | Physics simulation | Row currents, TIA conversion, ADC levels |
| GetReadRange() | Configuration | Returns material-derived range |

### GUI Interaction Testing Approach

**Interactive elements tested via:**

```go
// Mode button state (tab_unified.go - manual testing required)
// - Click modeReadBtn → opMode=OpModeRead, modeReadBtn.Importance=High
// - Click modeWriteBtn → opMode=OpModeWrite, modeWriteBtn.Importance=High
// - Button text updates reflect current mode

// Write level slider (tab_unified.go:1489-1507)
// - Slider range: 0-29
// - Labels update: "Level: N", "Voltage: X.XXV"
// - DAC voltage set via SetDACVoltageForState()

// Input vector entries (tab_unified.go:1527-1551)
// - 8 entry fields accept 0-255 values
// - DAC voltages updated via SetDACPreset(DACInputVector, ...)
// - Invalid input (non-numeric) ignored by strconv.Atoi

// Canvas refresh (tab_unified.go:460-538)
// - Tapped(e *fyne.PointEvent) converts screen coords to grid
// - Calculate col = (X - offsetX) / cellSize
// - Calculate row = (Y - offsetY) / cellSize
```

### Example Test Cases

**Voltage Range Calculations** (integration testing):

```go
// Pseudo-code for test verification
func TestVoltageRangeCalculations() {
    // Setup
    material := ferroelectric.GetHZOMaterial()  // e.g., FeCIM HZO
    ds := NewDeviceState(8, 8, tia, adc)
    ds.SetMaterial(material)

    // Read range validation
    Vc := material.CoerciveVoltage()
    expectedReadMax := 0.7 * Vc  // Default FieldMinRatio
    assert(ds.GetReadRange().Max == expectedReadMax)

    // Write range validation
    expectedWriteMin := Vc
    expectedWriteMax := 2.5 * Vc  // Default FieldMaxRatio
    assert(ds.GetWriteRange().Min == expectedWriteMin)
    assert(ds.GetWriteRange().Max == expectedWriteMax)
}

// Write level slider verification
func TestWriteLevelToVoltageMapping() {
    for level := 0; level < 30; level++ {
        ds.SetDACVoltageForState(0, level)
        voltage := ds.GetDACVoltage(0)

        // Verify linear interpolation
        expected := writeRange.Min +
                   (float64(level)/(30-1)) *
                   (writeRange.Max - writeRange.Min)
        assert(voltage ≈ expected)  // With 0.01V tolerance
    }
}

// Input vector voltage mapping
func TestInputVectorToDAC() {
    inputs := []float64{0, 64, 128, 192, 255}
    ds.SetDACPreset(DACInputVector, inputs...)

    for col, input := range inputs {
        voltage := ds.GetDACVoltage(col)
        expected := readRange.Min +
                   (input/255.0) *
                   (readRange.Max - readRange.Min)
        assert(voltage ≈ expected)
    }
}

// Bounds checking for set operations
func TestSetDACVoltageOutOfBounds() {
    ds.SetDACVoltage(-1, 0.5)   // Column < 0
    ds.SetDACVoltage(8, 0.5)    // Column >= cols
    ds.SetDACVoltage(4, 0.5)    // Valid column

    // Only column 4 should be updated
    assert(ds.GetDACVoltage(-1) == 0.0)
    assert(ds.GetDACVoltage(8) == 0.0)
    assert(ds.GetDACVoltage(4) == 0.5)
}
```

---

## Reference Tab Architecture

### tab_reference_specs.go Structure

**Purpose**: Display comprehensive electrical and physical specifications for all peripheral components

**Organization**:
- `createReferenceSpecsSection()` - Main section builder (line 20)
- `createSpecArraySection()` - Array configuration (rows, cols, cell density, topology)
- `createSpecDACSection()` - DAC parameters (resolution bits, conversion time, INL/DNL, settling)
- `createSpecADCSection()` - ADC parameters (resolution, conversion time, noise floor)
- `createSpecTIASection()` - TIA specifications (gain, bandwidth, output range, input-referred noise)
- `createSpecFeFETSection()` - FeFET cell parameters (Pr, Ec, coercive voltage, conductance levels)
- `createSpecSummarySection()` - System-level summary (power, area, energy per operation)

**Data Sources**:
- Array size from `ca.deviceState` (rows, cols)
- Material properties from `ca.deviceState.material` (HZOMaterial)
- Peripheral specs from `ca.deviceState.tia` and `ca.deviceState.adc`
- Physics config from `config/physics.Load()`

**Export Functions**:
- `onExportSpecs()` - Export specs to CSV/JSON (stub - "Coming soon")
- `onCompareToGPU()` - Show comparison dialog with GPU equivalent specs

### tab_reference_timing.go Structure

**Purpose**: Visualize timing relationships and signal waveforms for write, read, and compute operations

**Organization**:
- `createReferenceTimingSection()` - Main section builder with operation selector (line 23)
- `createTimingWriteSection()` - WRITE operation timing diagram
- `createTimingReadSection()` - READ operation timing diagram
- `createTimingComputeSection()` - COMPUTE (MVM) operation timing diagram

**Drawing Functions**:
- `drawTimingWrite(w, h int)` - Render WRITE operation signals (line 76)
  - Signals: CLK, ROW_SEL, COL_SEL, WL_PULSE, BL_VOLTAGE, SENSE_EN, ADC_CONVERT
  - Time resolution: nanoseconds
  - Color-coded signal levels (high=cyan, low=dark)

- `drawTimingRead(w, h int)` - Render READ operation signals
  - Signals: CLK, ROW_SEL, BL_VOLTAGE (low for sense), SENSE_EN (acquire), ADC (convert/output)

- `drawTimingCompute(w, h int)` - Render COMPUTE operation signals
  - Signals: CLK (controls input vector shift), WL_ALL (all high), BL_VOLTAGE (per column), SENSE_EN (parallel), ADC_OUTPUT (per row)

**Key Timing Parameters** (from drawing code):
- Clock period: ~10ns
- Pulse widths: Row/column select pulses 5-10 timing units
- Sense window: 8-15 timing units
- ADC conversion: 20+ timing units

**Features**:
- Operator selector dropdown (WRITE/READ/COMPUTE) at top
- Animation button for step-through demonstration
- Export button for SVG output (stub - "Coming soon")
- Status label shows selected operation timing

**Canvas Properties**:
- Minimum size: 600×200 pixels (timing waveforms)
- Light background color (FeCIMTheme.ColorDarkBG)
- Signal colors: Cyan (active), Orange (timing markers), Gray (inactive)

---

## Thread Safety

### Mutex Protection
All shared state accessed via ca.mu (RWMutex):
- arrayWeights (read/write)
- inputVector, outputVector (read/write)
- architecture (read/write)
- animationStep, animationActive (read/write)

### DeviceState Thread Safety
DeviceState methods should be called under appropriate locking in CircuitsApp.

### Canvas Refresh Pattern
All canvas refresh calls wrapped in fyne.Do():
```go
fyne.Do(func() {
    ca.sharedArrayCanvas.Refresh()
})
```

---

## Physics Constants (Now Dynamic)

| Parameter | Source | Calculation |
|-----------|--------|-------------|
| Coercive Voltage (Vc) | material.CoerciveVoltage() | Ec × thickness |
| Read Max Voltage | physics.yaml + material | FieldMinRatio × Vc (capped at 1.0V, floor 0.1V) |
| Write Min Voltage | material | Vc |
| Write Max Voltage | physics.yaml + material | min(FieldMaxRatio × Vc, MaxPracticalVoltage) |
| Max Practical Voltage | device_state.go:92 | 3.0V (hardware limit) |
| FeCIM Levels | app.go:25 | 30 |
| Default Array Size | app.go:27 | 8×8 |

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
- fecim-lattice-tools/shared/peripherals (DAC, ADC, TIA, ChargePump)
- fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric (HZOMaterial, AllMaterials)
- fecim-lattice-tools/config/physics (Load physics.yaml config)
- fecim-lattice-tools/shared/theme (FeCIMTheme)
- fecim-lattice-tools/shared/widgets (DebugInteraction, Architecture constants)

---

## Performance Characteristics

### Canvas Refresh Frequency

**Expected refresh timing**:
- **Normal interaction**: 60 FPS (16.7 ms per frame) when user idle
- **Active slider/entry**: 30-60 FPS (driven by Fyne event loop)
- **Animation mode**: 16-30 FPS (animationStep cycles through 0→1→2→3→0)
  - Each step takes ~500ms for visual clarity
  - Total animation cycle: ~2 seconds

**Memory Considerations per Array Size**:
```
8×8 array:    ~1.2 KB (arrayWeights) + Raster canvas memory
16×16 array:  ~4.8 KB (arrayWeights) + Raster canvas memory
32×32 array:  ~19.2 KB (arrayWeights) + Raster canvas memory
64×64 array:  ~76.8 KB (arrayWeights) + Raster canvas memory
```

**Raster Canvas Memory** (tab_unified.go:460-538):
- Default size: 400×350 pixels
- RGBA image buffer: 400 × 350 × 4 bytes = ~560 KB per refresh
- Raster redraws on every canvas.Refresh() call
- Painting is synchronized with fyne.Do() for thread safety

### Animation Goroutine Lifecycle

**Animation flow** (tab_unified.go:1097-1137):

```
User clicks "Animate" button
    ↓
animationActive = true
animationStep = 1 (DAC highlight phase)
    ↓
Wait 500ms
    ↓
animationStep = 2 (Array highlight phase)
    ↓
Wait 500ms
    ↓
animationStep = 3 (ADC highlight phase)
    ↓
Wait 500ms
    ↓
animationStep = 0 (reset)
animationActive = false
```

**Goroutine Pattern**:
- Animation runs in separate goroutine spawned by button handler
- Uses `time.Sleep(500ms)` for pacing between steps
- All state updates protected by ca.mu (mutex)
- Canvas refresh called on each step via `fyne.Do()`
- Animation completes cleanly without blocking UI thread

### Voltage Range Calculation Performance

**Material coercive voltage lookup**: O(1)
```go
Vc := material.CoerciveVoltage()  // Cached in material struct
```

**Voltage range updates**: O(1)
- Called when material selected: device_state.go:SetMaterial()
- Called when physics.yaml loaded on startup

**DAC voltage updates**: O(cols)
```go
SetDACPreset() → loops cols times to update dacVoltages array
SetDACVoltageForState() → single column update O(1)
```

**Compute simulation**: O(rows × cols)
- Called on every MVM or array refresh
- Inner loop: sum currents from all columns for each row (device_state.go:404-459)
- TIA + ADC conversion: O(rows)

### Input Vector Processing

**Input entry validation**: O(8) for 8 columns
```go
onInputVectorEntryChanged() → strconv.Atoi() → SetDACPreset(DACInputVector)
```

**DAC voltage calculation from input**: O(cols)
```go
// Maps each input 0-255 to readRange.Min → readRange.Max
normalized := params[i] / 255.0
ds.dacVoltages[i] = ds.readRange.Min + normalized*(ds.readRange.Max-ds.readRange.Min)
```

### Known Performance Bottlenecks

1. **Large array rendering**: 64×64 array redraws can take 2-3 frames to complete
   - Mitigation: Limit to 32×32 for smooth interaction
   - Consider hardware acceleration in future

2. **Material property lookup** in Compute loop: Currently O(1) with fallback
   - If material nil: falls back to linear model (negligible cost)
   - No caching required - calls are infrequent

3. **Canvas refresh serialization**: fyne.Do() ensures thread safety but adds latency
   - Each refresh waits for Fyne event loop
   - Typical latency: 1-5ms

---

## Troubleshooting Guide

### Canvas Not Refreshing

**Symptom**: Array visualization appears frozen after voltage/mode changes

**Root Cause**: Canvas refresh not called or called outside fyne.Do()

**Solution**:
```go
// ✓ Correct: Wrapped in fyne.Do()
fyne.Do(func() {
    ca.sharedArrayCanvas.Refresh()
})

// ✗ Wrong: Direct refresh (may cause data races)
ca.sharedArrayCanvas.Refresh()
```

**Debugging**:
1. Check if mode button click triggers `updateModeButtons()` call
2. Verify `ca.sharedArrayCanvas` is initialized in BuildContent()
3. Enable logging in updateUnifiedArrayDisplay() (tab_unified.go)
4. Check Fyne event queue: ensure app.Run() is active

**Reference**: Thread Safety section, Canvas Refresh Pattern subsection

---

### WL Checkboxes Not Updating in Passive Mode

**Symptom**: WL checkbox states don't match expected pattern when switching to PASSIVE architecture

**Root Cause**: Passive mode (0T1R) always has all WLs active regardless of mode - no transistor gating

**Expected Behavior**:
- PASSIVE mode: All WL checkboxes always checked AND DISABLED (user cannot uncheck)
- 1T1R/2T1R: WL checkboxes enabled and reflect mode:
  - READ: Single row checked
  - WRITE: Single row checked
  - COMPUTE: All rows checked

**Solution (Defense-in-Depth)**:

**UI Layer** (tab_unified.go):
```go
func (ca *CircuitsApp) updateWLCheckboxes() {
    isPassive := ca.architecture == sharedwidgets.Architecture0T1R
    for i, check := range ca.unifiedWLChecks {
        fyne.Do(func() {
            if isPassive {
                check.SetChecked(true)
                check.Disable()  // Cannot uncheck in passive mode
            } else {
                check.Enable()
                check.SetChecked(isActive)
            }
        })
    }
}
```

**Data Layer** (device_state.go):
```go
func (ds *DeviceState) SetPassiveMode(passive bool) {
    ds.isPassive = passive
    if passive {
        ds.wlMode = WLAll
        for i := range ds.activeRows {
            ds.activeRows[i] = true
        }
    }
}

func (ds *DeviceState) SetWLSingle(row int) {
    if ds.isPassive {
        return  // Ignore - passive mode enforces all WLs on
    }
    // ... normal logic
}
```

**Debugging**:
1. Check `ca.architecture` value (should be `sharedwidgets.Architecture0T1R` for passive)
2. Verify `ca.deviceState.isPassive` is set correctly
3. Verify `ca.unifiedWLChecks` are initialized before `SetChecked()` is called
4. Check for nil pointer if checkbox OnChanged fires before array is populated

---

### Voltage Ranges Showing Unexpected Values

**Symptom**: DAC preset buttons show incorrect voltage ranges (e.g., "Read (0.0-0.0V)")

**Root Cause**:
1. Material not set (nil pointer)
2. physics.yaml missing or malformed
3. Calibration parameters loading failed

**Expected Ranges** (for HZO materials):
```
Material: FeCIM HZO (typical)
Vc ≈ 1.0V (from material.CoerciveVoltage())
Read range: 0 → 0.5×Vc = 0.0V → 0.5V
Write range: 1.0×Vc → 2.5×Vc = 1.0V → 2.5V
(Read range is capped at 1.0V and floored at 0.1V for ADC/DAC practicality)
```

**Solution**:
```go
// device_state.go:242-256 - Update voltage ranges on material change
func (ds *DeviceState) SetMaterial(mat *ferroelectric.HZOMaterial) {
    ds.material = mat
    if ds.material == nil {
        return  // Falls back to previous material
    }

    Vc := ds.material.CoerciveVoltage()
    if Vc <= 0 {
        return  // Invalid coercive voltage
    }

    // Update read/write ranges (with practical caps)
    safeReadMax := ds.calibParams.FieldMinRatio * Vc
    if safeReadMax > 1.0 {
        safeReadMax = 1.0
    }
    if safeReadMax < 0.1 {
        safeReadMax = 0.1
    }
    ds.readRange.Max = safeReadMax

    ds.writeRange.Min = Vc
    ds.writeRange.Max = ds.calibParams.FieldMaxRatio * Vc
    ds.writeRange.Max = min(ds.writeRange.Max, MaxPracticalVoltage)  // Hardware limit
}
```

**Debugging**:
1. Check material selector dropdown - ensure material is selected
2. Verify physics.yaml exists at `config/physics.yaml`
3. In device_state.go, add logging to loadCalibrationParams():
   ```go
   cfg, err := physics.Load()
   if err != nil {
       log.Printf("physics.yaml load failed: %v", err)  // Fallback to defaults
   }
   ```
4. Check calibration parameters: `log.Printf("FieldMinRatio=%f, FieldMaxRatio=%f", calibParams.FieldMinRatio, calibParams.FieldMaxRatio)`

**References**:
- Voltage Range Configuration section (line 254)
- Error Handling: Fallback Behavior section
- Physics Constants section

---

### Material Selector Not Populating

**Symptom**: Material selector dropdown appears empty or shows no options

**Root Cause**: `ferroelectric.AllMaterials()` returns empty or nil slice

**Solution**:
```go
// tab_unified.go:97-122 - Material selector initialization
materials := ferroelectric.AllMaterials()
materialNames := make([]string, len(materials))
for i, m := range materials {
    materialNames[i] = m.String()  // e.g., "FeCIM HZO", "HZO (Si-doped)", etc.
}

ca.materialSelector = widget.NewSelect(materialNames, func(s string) {
    // Find material by name and set it
    for _, mat := range ferroelectric.AllMaterials() {
        if mat.String() == s {
            ca.deviceState.SetMaterial(mat)
            ca.updateDACPresetLabels()
            ca.sharedArrayCanvas.Refresh()
            return
        }
    }
})
```

**Debugging**:
1. Check module1-hysteresis/pkg/ferroelectric/ferroelectric.go for AllMaterials() definition
2. Verify materials are registered in material initialization code
3. Test AllMaterials() directly:
   ```bash
   go test -run TestMaterialSelection ./module4-circuits/pkg/gui/...
   ```
4. Check if ferroelectric package import is correct: `"fecim-lattice-tools/module1-hysteresis/pkg/ferroelectric"`

**References**: Peripheral Circuits Visualizer (module4) README for material list

---

### High Memory Usage During Animation

**Symptom**: Memory usage spikes to 100+ MB during "Animate" button operation

**Root Cause**: Repeated canvas redraw creates temporary image buffers

**Solution - Optimize Animation**:
1. Reduce canvas size from 400×350 to 320×280 pixels
2. Limit animation to specific array sizes (8×8 only, disable for 64×64)
3. Reuse Raster instead of recreating it per frame

```go
// Recommended optimization
const AnimationCanvasWidth = 320
const AnimationCanvasHeight = 280
const MaxAnimationArraySize = 8  // Limit animation to 8×8 arrays only

if rows > MaxAnimationArraySize || cols > MaxAnimationArraySize {
    showDialog("Animation disabled for large arrays (>8×8)")
    return
}
```

**Memory Per Animation Cycle**:
- Single frame: ~400 KB (400×350×4 bytes RGBA)
- 4-step animation: ~1.6 MB peak
- Should complete in <5 seconds without lingering allocations

**References**: Performance Characteristics section

---

## Notes

1. **Unified Device Simulation**: The OPERATIONS view is now a true device simulator. Configure WL and DAC inputs, see outputs in real-time. No artificial "modes" - the hardware is the same, only inputs differ.

2. **Material Selection**: The material selector loads all materials from `ferroelectric.AllMaterials()`. Changing material updates voltage ranges and recalculates outputs.

3. **Architecture Toggle**: Switches between PASSIVE (0T1R), 1T1R, and 2T1R. Passive mode always has all WLs active and DISABLED (user cannot turn off). 1T1R/2T1R draw transistor symbols.

4. **Dynamic Voltage Ranges**: All voltage thresholds are derived from physics.yaml calibration parameters and material properties. No hardcoded values. DAC preset button labels update automatically.

5. **Mode Buttons vs RadioGroup**: The new mode buttons (READ/WRITE/COMPUTE) replace the old RadioGroup. They provide better visual feedback with button importance highlighting.

6. **Calibration from physics.yaml**: The `CalibrationParams` struct loads `field_min_ratio` and `field_max_ratio` from `config/physics.yaml`. These define operating regions relative to coercive voltage.

7. **Operation Classification Help**: The `operationsModeHelp` label shows mode-specific guidance including voltage ranges and architecture-specific notes (e.g., sneak paths in passive mode).

8. **Embedded Interface**: Implements BuildContent(), Start(), Stop() for integration with main visualizer (cmd/fecim-lattice-tools).

9. **Cell Selection Feedback**: Selected cells show gold 2-pixel border only (no fill). This preserves visibility of the conductance state color within the cell.

10. **Color Mapping**: Uses blue-gray-red gradient based on level relative to mid-point:
    - Blue: Low conductance states (0 to mid-1)
    - Gray: Neutral/mid state
    - Red: High conductance states (mid+1 to max)

11. **Undo Functionality**: Single-level undo for array changes. History saved before each write operation. Button enabled only when history exists.

12. **Passive Mode Enforcement**: Defense-in-depth approach:
    - UI: Checkboxes disabled in passive mode
    - Data: DeviceState.SetWLSingle() ignored when isPassive=true

13. **Dynamic quantLevels**: Write slider range and array initialization use `ca.quantLevels` (default 30) instead of hardcoded values.

14. **Mid-Level Initialization**: Array cells start at mid-level (quantLevels/2 = 15 for 30 states), representing neutral polarization state.

15. **4-Phase Write Sequence**: Proper ferroelectric write timing (RESET→HOLD→WRITE→HOLD) ensures reliable programming. Animated timing diagram shows each phase.

16. **ISPP Programming**: Incremental Step Pulse Programming with overshoot detection. If cell overshoots target, automatically resets to saturation and restarts. Max 10 iterations.

17. **V/2 Half-Select Visualization**: In passive (0T1R) mode, shows cells receiving half voltage during write. Gold = target cell, Amber = half-selected cells at risk of disturb.

18. **Hysteresis Direction Tracking**: Each cell tracks whether it was last programmed ascending or descending. Affects voltage calculation for accurate programming.

19. **Per-Level Voltage Calibration**: 30-level voltage arrays allow fine-grained control. Uses linear interpolation between calibration points.

20. **Charge Pump Modeling**: The charge pump is treated as a regulated write‑rail supply (used for timing/power analysis and output limits); the GUI applies the resulting write voltages directly rather than simulating pump dynamics per cell.

---

## Related Documentation

- **Physics Basis**: `docs/research/circuits.CIM-fundamentals.md` - READ/WRITE/COMPUTE physics
- **Gap Analysis**: `docs/research/MODULE4-PHYSICS-IMPROVEMENTS.md` - Critical/High severity gaps
- **Implementation Plan**: `docs/plans/module4-plan-improvements.md` - 12-task phased implementation
