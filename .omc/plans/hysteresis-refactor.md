# Hysteresis Module Refactoring Plan

## Context

### Original Request
Refactor the hysteresis module (`module1-hysteresis/`) for easier maintenance.

### Research Findings

**Current File Sizes (Lines of Code):**
| File | Lines | Primary Concerns |
|------|-------|------------------|
| `simulation.go` | 1803 | Calibration, phase management, waveform logic, UI updates, energy tracking |
| `gui.go` | 618 | App struct (176 fields!), UI creation, debug logging, lifecycle |
| `controls.go` | 413 | Control panel creation, event handlers |
| `info.go` | 276 | Info panels, educational content, dialogs |
| `keyboard.go` | 218 | Keyboard shortcuts |
| `export.go` | 221 | Data export functionality |
| `embedded.go` | 89 | Embedded app interface |
| `theme.go` | 123 | Theme definitions |
| `preisach_advanced.go` | 503 | Physics model (well-structured) |
| `material.go` | 496 | Material definitions (well-structured) |

**Widget Files (Reasonably Sized):**
- `widgets/peplot.go`: 420 lines
- `widgets/level.go`: 461 lines
- `widgets/phase.go`: 278 lines
- `widgets/cell.go`: 255 lines
- `widgets/mode.go`: 165 lines

### Key Issues Identified

1. **Massive `simulation.go` (1803 lines)** - Contains 7+ distinct concerns:
   - Calibration data structures and persistence (~350 lines)
   - Temperature-aware calibration (~200 lines)
   - Main simulation loop (~100 lines)
   - Manual waveform animation (~250 lines)
   - Write/Read Demo waveform logic (~500 lines)
   - Time-resolved waveform logic (~70 lines)
   - UI update logic (~300 lines)

2. **Bloated `App` struct (176+ fields in gui.go)** - Mixes:
   - Physics state (material, preisach, polarization)
   - Simulation state (running, paused, simTime)
   - Waveform state (wrdPhase, manualAnimating, timeResIndex)
   - Calibration state (calibrationUp/Down, tempCalibrations)
   - UI components (20+ widget references)
   - Demo metrics (wrdTotalWrites, wrdSuccessWrites)

3. **Cross-cutting concerns not separated:**
   - Logging mixed with business logic
   - UI updates interleaved with physics
   - State persistence scattered across methods

---

## Work Objectives

### Core Objective
Extract cohesive subsystems from the monolithic `simulation.go` and `App` struct to create a maintainable, testable architecture.

### Deliverables
1. New `calibration/` subpackage for level calibration logic
2. New `waveforms/` subpackage for waveform generators
3. Extracted `state.go` for simulation state management
4. Refactored `App` struct with embedded sub-components
5. Clear interface boundaries between physics, state, and UI

### Definition of Done
- [ ] No file exceeds 500 lines (excluding tests)
- [ ] `App` struct has fewer than 50 direct fields
- [ ] All calibration logic is in `calibration/` package
- [ ] All waveform logic is in `waveforms/` package
- [ ] Existing tests pass without modification
- [ ] New unit tests for extracted components
- [ ] No functionality regression

---

## Must Have / Must NOT Have

### Must Have (Guardrails)
- Preserve all existing physics behavior
- Maintain thread-safety (mutex patterns)
- Keep backward compatibility for embedded interface
- Preserve calibration file format (v2 JSON)
- Keep `fyne.Do()` pattern for UI updates

### Must NOT Have
- No changes to `ferroelectric/` package internals
- No changes to widget implementations
- No new external dependencies
- No breaking changes to public APIs
- No removal of features

---

## Task Flow and Dependencies

```
[Task 1] Extract CalibrationManager
    |
    v
[Task 2] Extract WaveformGenerator Interface
    |
    +---> [Task 3] ManualWaveform
    |
    +---> [Task 4] WriteReadWaveform
    |
    +---> [Task 5] TimeResolvedWaveform
    |
    v
[Task 6] Extract SimulationState
    |
    v
[Task 7] Refactor App Struct
    |
    v
[Task 8] Update simulation.go (Main Loop Only)
    |
    v
[Task 9] Add Unit Tests
    |
    v
[Task 10] Integration Testing
```

---

## Detailed TODOs

### Task 1: Extract CalibrationManager

**File:** `module1-hysteresis/pkg/gui/calibration/manager.go`

**Extract from simulation.go:**
- `TempCalibration` struct (lines 16-26)
- `CalibrationData` struct (lines 28-45)
- `saveCalibration()` method (lines 62-116)
- `loadCalibration()` method (lines 118-220)
- `initializeTempCalibrationBounds()` (lines 222-242)
- `loadTempCalibration()` (lines 244-267)
- `validateCalibration()` (lines 269-283)
- `countDuplicates()` helper (lines 285-306)
- `loadCalibrationForTemperature()` (lines 308-343)
- `findNearestCalibrations()` (lines 345-374)
- `interpolateCalibrations()` (lines 376-412)
- `hasCalibrationNear()` (lines 414-422)
- `onTemperatureChanged()` (lines 424-444)
- `calibrateLevelsAtTemperature()` (lines 1575-1608)
- `calibrateLevels()` (lines 1610-1803)

**New Interface:**
```go
type CalibrationManager interface {
    Load() bool
    Save() error
    GetCalibrationUp(level int) float64
    GetCalibrationDown(level int) float64
    OnTemperatureChanged(newTemp float64)
    CalibrateAtTemperature(tempK float64)
    UpdateCalibration(level int, ascending bool, error int, currentE, Ec float64)
}
```

**Acceptance Criteria:**
- [ ] CalibrationManager is fully self-contained
- [ ] No direct access to App fields
- [ ] Thread-safe with internal mutex
- [ ] Existing calibration files load correctly

---

### Task 2: Extract WaveformGenerator Interface

**File:** `module1-hysteresis/pkg/gui/waveforms/generator.go`

**Define Interface:**
```go
type WaveformGenerator interface {
    // Update returns the E-field for this tick
    Update(dt float64, state *SimulationState) float64

    // GetPhase returns current phase for UI display
    GetPhase() int

    // GetPhaseMode returns "wrd", "manual", or ""
    GetPhaseMode() string

    // Reset reinitializes the waveform
    Reset()

    // IsAnimating returns true if in active animation
    IsAnimating() bool
}

type SimulationState struct {
    ElectricField  float64
    Polarization   float64
    DiscreteLevel  int
    NormalizedP    float64
    NumLevels      int
    Ec             float64
    Frequency      float64
    SimTime        float64
}
```

**Acceptance Criteria:**
- [ ] Interface is minimal and complete
- [ ] SimulationState captures all needed state
- [ ] No UI dependencies in interface

---

### Task 3: Extract ManualWaveform

**File:** `module1-hysteresis/pkg/gui/waveforms/manual.go`

**Extract from simulation.go lines 480-727:**
- Manual mode phase logic (RESET/HOLD_RESET/WRITE/HOLD_WRITE)
- Phase timing calculations
- Ramp rate logic
- Calibration adjustment logic

**Acceptance Criteria:**
- [ ] All 4 phases work correctly
- [ ] Click-to-level animation preserved
- [ ] Calibration feedback works

---

### Task 4: Extract WriteReadWaveform

**File:** `module1-hysteresis/pkg/gui/waveforms/writeread.go`

**Extract from simulation.go lines 747-1137:**
- 6-phase Write/Read Demo logic
- Dr. Tour metrics tracking
- Debug logging
- Energy calculation

**New Struct:**
```go
type WriteReadWaveform struct {
    // Phase state
    phase       int
    phaseTimer  float64
    targetLevel int
    startLevel  int
    readLevel   int

    // Metrics
    totalWrites   int
    successWrites int
    totalEnergyfJ float64
    cycleEnergy   float64

    // Phase transition data for logging
    resetStartP, resetEndP float64
    writeStartP, writeEndP float64
    readStartP             float64
    // ... etc
}
```

**Acceptance Criteria:**
- [ ] All 6 phases work correctly
- [ ] Metrics accumulate correctly
- [ ] Energy tracking accurate
- [ ] Debug log still works

---

### Task 5: Extract TimeResolvedWaveform

**File:** `module1-hysteresis/pkg/gui/waveforms/timeresolved.go`

**Extract from simulation.go lines 1138-1209:**
- KAI dynamics animation
- Precomputed data management
- Animation index tracking

**Acceptance Criteria:**
- [ ] Animation plays correctly
- [ ] Loop behavior preserved
- [ ] Data arrays managed properly

---

### Task 6: Extract SimulationState

**File:** `module1-hysteresis/pkg/gui/state.go`

**Extract from gui.go App struct:**
- Physics state fields (lines 47-60)
- History fields (lines 62-65)
- UI runtime state (lines 67-73)
- Move `eHistory`, `pHistory` management here

**New Struct:**
```go
type SimulationState struct {
    mu sync.RWMutex

    // Physics
    ElectricField float64
    Polarization  float64
    NormalizedP   float64
    DiscreteLevel int

    // History
    EHistory   []float64
    PHistory   []float64
    MaxHistory int

    // Runtime
    Running   bool
    Paused    bool
    SimTime   float64

    // Configuration
    NumLevels int
    Frequency float64
    Waveform  WaveformType
}
```

**Acceptance Criteria:**
- [ ] All state access is thread-safe
- [ ] History management is encapsulated
- [ ] State transitions are clear

---

### Task 7: Refactor App Struct

**File:** `module1-hysteresis/pkg/gui/gui.go`

**Reduce App to composition:**
```go
type App struct {
    fyneApp    fyne.App
    mainWindow fyne.Window

    // Sub-components (composition)
    state       *SimulationState
    calibration *calibration.Manager
    waveform    waveforms.WaveformGenerator

    // Physics (unchanged)
    material  *ferroelectric.HZOMaterial
    preisach  *ferroelectric.MayergoyzPreisach
    materials []*ferroelectric.HZOMaterial
    matIndex  int

    // UI widgets (keep as-is)
    plot           *widgets.PEPlot
    levelIndicator *widgets.LevelIndicator
    // ... (widget references stay)
}
```

**Acceptance Criteria:**
- [ ] App struct under 50 direct fields
- [ ] Sub-components are clear
- [ ] No functionality loss

---

### Task 8: Update simulation.go (Main Loop Only)

**File:** `module1-hysteresis/pkg/gui/simulation.go`

**Reduce to:**
- `simulationLoop()` - now just orchestrates
- `updateUI()` - moved from 300 lines to delegation
- Remove all extracted code

**Target size:** ~200-300 lines

**Acceptance Criteria:**
- [ ] Main loop is readable
- [ ] Delegates to waveforms package
- [ ] UI updates are clean

---

### Task 9: Add Unit Tests

**Files:**
- `calibration/manager_test.go`
- `waveforms/manual_test.go`
- `waveforms/writeread_test.go`
- `waveforms/timeresolved_test.go`
- `state_test.go`

**Test Coverage:**
- Calibration save/load round-trip
- Temperature interpolation
- Phase transitions
- Energy calculations
- State thread-safety

**Acceptance Criteria:**
- [ ] 80%+ coverage on new packages
- [ ] No race conditions (go test -race)
- [ ] Tests run in <5 seconds

---

### Task 10: Integration Testing

**Verify:**
- [ ] Existing widget tests pass
- [ ] Manual mode click-to-level works
- [ ] Write/Read Demo cycles correctly
- [ ] Temperature changes recalibrate
- [ ] Calibration persists across restarts
- [ ] Embedded app works in unified visualizer

---

## Commit Strategy

| Commit | Description |
|--------|-------------|
| 1 | `refactor(hysteresis): extract calibration package` |
| 2 | `refactor(hysteresis): add waveform generator interface` |
| 3 | `refactor(hysteresis): extract manual waveform` |
| 4 | `refactor(hysteresis): extract writeread waveform` |
| 5 | `refactor(hysteresis): extract timeresolved waveform` |
| 6 | `refactor(hysteresis): extract simulation state` |
| 7 | `refactor(hysteresis): slim down App struct` |
| 8 | `refactor(hysteresis): clean up simulation loop` |
| 9 | `test(hysteresis): add unit tests for extracted components` |
| 10 | `test(hysteresis): integration testing and verification` |

---

## Success Criteria

### Quantitative
- [ ] `simulation.go` reduced from 1803 to <300 lines
- [ ] `gui.go` App struct reduced from 176+ to <50 fields
- [ ] No file exceeds 500 lines
- [ ] Test coverage >80% on new code
- [ ] Zero test regressions

### Qualitative
- [ ] Clear separation of concerns
- [ ] Each file has single responsibility
- [ ] Easy to locate code for any feature
- [ ] New waveform types easy to add

---

## Risk Identification and Mitigations

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| Breaking thread-safety | Medium | High | Keep mutex patterns, add race tests |
| Calibration regression | Low | High | Extensive round-trip tests |
| UI responsiveness | Low | Medium | Keep fyne.Do() pattern |
| Phase timing drift | Medium | Medium | Unit tests for phase transitions |
| Energy calculation errors | Low | Medium | Golden file tests |

---

## Testing Strategy

### Phase 1: Smoke Tests
Before any refactoring:
```bash
go test ./module1-hysteresis/...
```
Document baseline passing tests.

### Phase 2: Incremental Testing
After each extraction:
- Run full test suite
- Manual verification of affected mode
- Check for race conditions

### Phase 3: Final Verification
- Complete E2E testing
- Performance benchmarking
- UI responsiveness testing

---

## Appendix: Current Code Structure Analysis

### simulation.go Concern Map (1803 lines)

| Lines | Concern | Complexity |
|-------|---------|------------|
| 1-60 | Calibration structs | Low |
| 62-220 | Calibration persistence | Medium |
| 222-444 | Temperature handling | Medium |
| 446-478 | Simulation loop setup | Low |
| 480-727 | Manual waveform (4 phases) | High |
| 729-746 | Sine/Triangle waveforms | Low |
| 747-1137 | Write/Read Demo (6 phases) | Very High |
| 1138-1209 | Time-resolved waveform | Medium |
| 1210-1263 | Physics update | Medium |
| 1264-1573 | UI update | High |
| 1575-1803 | Calibration algorithms | High |

### App Struct Field Categories

| Category | Count | Examples |
|----------|-------|----------|
| Fyne refs | 2 | fyneApp, mainWindow |
| Physics | 4 | material, preisach, materials, matIndex |
| State | 12 | electricField, polarization, running, paused... |
| WRD Demo | 20 | wrdPhase, wrdTargetLevel, wrdTotalWrites... |
| Manual Mode | 5 | manualAnimating, manualTargetLevel... |
| TimeRes | 4 | timeResAnimating, timeResDataTimes... |
| Calibration | 12 | calibrationUp, calibrationDown, tempCalibrations... |
| UI Widgets | 25+ | plot, levelIndicator, eFieldSlider... |
| Logging | 4 | logEntries, maxLogLines, lastLogPhase... |

**Total: ~88+ fields** (some nested in sub-structs)

---

## Notes

- The `ferroelectric/` package is well-structured and should not be modified
- Widget files are reasonably sized and don't need refactoring
- The embedded interface (`BuildContent`, `Start`, `Stop`) must be preserved
- Consider adding a `WaveformType` enum to the waveforms package

