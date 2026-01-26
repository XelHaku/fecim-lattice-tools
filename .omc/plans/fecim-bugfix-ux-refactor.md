# FeCIM Lattice Tools: Bug Fix, UI/UX, and Refactoring Plan

**Created**: 2026-01-25
**Plan Type**: Comprehensive (Bug Fixes + UI/UX + Refactoring)
**Source**: HYPER_ANALYSIS_REPORT.md analysis + codebase review
**Iteration**: 2 (Revised after Critic feedback)

---

## Context

### Original Request
Fix bugs, UI/UX issues, and refactor large files in the FeCIM Lattice Tools project as identified in the HYPER_ANALYSIS_REPORT.md.

### Summary of Findings
- **1 Potential Bug**: Module 1 info.go physics text (INVESTIGATE - may be false positive)
- **4 High Priority UI/UX Issues**: Missing labels, poor readability
- **3 Medium Priority UI/UX Issues**: Text sizing, tooltips
- **8 Files > 500 Lines**: Need structured refactoring

---

## Phase 0: Gap Analysis

### Additional Issues Discovered

| Issue | Location | Severity | Description |
|-------|----------|----------|-------------|
| Missing units | Multiple modules | MEDIUM | Charts lack unit labels (V, uA, mV) |
| Thread safety audit | All GUI files | LOW | Verify all goroutine UI updates use `fyne.Do()` |
| Widget duplication | module2-crossbar widgets.go | LOW | ColorLegend duplicated in shared/widgets |

### Architecture Patterns to Preserve
1. **EmbeddedApp Interface**: `BuildContent()`, `Start()`, `Stop()` - all modules follow this
2. **Thread Safety**: `fyne.Do()` for UI updates from goroutines
3. **30-Level Quantization**: `crossbar.QuantizeTo30Levels()` for FeCIM physics
4. **Module Independence**: No cross-module imports except `shared/`

---

## Phase 1: Investigation - Verify Bug Existence

### Task 1.1: Module 1 Physics Text Verification [INVESTIGATE - POSSIBLE FALSE POSITIVE]

**File**: `<local-path>`

**Current Behavior (verified in code)**:
```go
// Line 106: isWrite logic is CORRECT
isWrite := math.Abs(a.electricField) > a.material.Ec

// Lines 127-134: Manual mode shows correct text
if isWrite {
    return fmt.Sprintf("██ WRITING LEVEL %d ██\n\n"+
        "Electric field E > Ec.\n"+
        "Domains are switching.\n"...)
}
return fmt.Sprintf("░░ HOLDING LEVEL %d ░░\n\n"+
    "E-field is low or zero.\n"...)

// Lines 143-155: Sine/Triangle modes dynamically update
phaseText := "░░ READING ░░"
if isWrite {
    phaseText = "██ WRITING ██"  // Correctly switches based on E vs Ec
}

// Lines 188-195: WriteReadDemo case 2 (READ) correctly shows "Sense pulse: |E| < Ec"
```

**Analysis Result: CODE IS CORRECT**

The HYPER_ANALYSIS_REPORT.md bug claim appears to be a **FALSE POSITIVE**:
1. The `isWrite` logic at line 106 is mathematically correct: `|E| > Ec` means write
2. The Sine/Triangle mode dynamically shows "READING" vs "WRITING" based on instantaneous E-field
3. The WriteReadDemo READ phase correctly shows "|E| < Ec"

**What the report MAY have confused**:
- During Sine/Triangle waveforms, E oscillates continuously
- When E crosses Ec, the indicator switches from "READING" to "WRITING"
- This is CORRECT physics behavior, not a bug

**Proposed Action**:
1. **VERIFY** by running the app and observing behavior in all waveform modes
2. **NO CODE CHANGE NEEDED** if physics text matches actual E vs Ec relationship
3. **OPTIONAL ENHANCEMENT**: Add clarifying text for educational purposes:
   - "Note: During continuous waveforms, mode switches as E crosses Ec threshold"

**Acceptance Criteria**:
- [ ] Manual test: Run each waveform mode, verify text reflects actual physics
- [ ] Confirm "E > Ec" only displays when |E| actually exceeds Ec
- [ ] Document finding (bug confirmed or false positive)

**Risk**: NONE - Investigation only, no code changes unless bug confirmed

---

## Phase 2: High Priority UI/UX Improvements

### Task 2.1: Add Axis Labels to Charts/Waveforms

**Affected Files**:
1. `<local-path>` (lines 830-932)
2. `<local-path>` (lines 127-248, 257-386, 394-555)
3. `<local-path>` (AccuracyWaterfall)

**Current Issue**:
- Waveforms show time axis labels but lack Y-axis unit labels
- Heatmaps lack colorbar value labels

**Changes Required**:

**tab_operations.go drawOpsWritePulse (lines 830-932)**:
- Add Y-axis label: "Voltage (V)"
- Add X-axis label: "Time (ns)"

**tab_reference.go drawTimingWrite/Read/Compute**:
- Add Y-axis signal names (already present)
- Add X-axis time labels (already present)
- Add title for each diagram

**widgets.go AccuracyWaterfall**:
- Add Y-axis label: "Accuracy (%)"
- Add X-axis label: "Degradation Stage"

**Acceptance Criteria**:
- [ ] All waveforms have X-axis and Y-axis labels with units
- [ ] Labels visible at all window sizes
- [ ] Font size appropriate (10-12pt)

**Risk**: LOW - Visual additions only

---

### Task 2.2: Improve Module 4 Specifications Readability

**File**: `<local-path>`
**Lines**: 634-951 (createReferenceSpecsSection and related functions)

**Current Issues**:
1. Text too small and dense
2. Specifications displayed as raw strings
3. Poor visual hierarchy

**Changes Required**:

1. **Convert spec strings to structured tables**:
   - Use `widget.NewTable()` or `container.NewGridWithColumns(2)`
   - Each row: Parameter name | Value | Unit

2. **Add section dividers and spacing**:
   - Use `widget.NewSeparator()` between sections
   - Add `container.NewPadded()` for breathing room

3. **Increase font sizes**:
   - Headers: Bold, 14pt
   - Specs: Monospace, 11pt
   - Help text: Italic, 10pt

**Example Refactor for createSpecDACSection (lines 755-780)**:
```go
// FROM:
specs := `Count:             32 (one per column)
Resolution:        8 bits (256 levels)
...`

// TO:
specTable := widget.NewTable(
    func() (int, int) { return 9, 3 }, // 9 rows, 3 cols
    func() fyne.CanvasObject { return widget.NewLabel("") },
    func(cell widget.TableCellID, obj fyne.CanvasObject) {
        // Parameter, Value, Unit
    },
)
```

**Acceptance Criteria**:
- [ ] All specs in structured table format
- [ ] Clear visual hierarchy (headers, sections)
- [ ] Readable at default window size
- [ ] Specs update when dropdowns change

**Risk**: MEDIUM - Significant UI restructuring

---

### Task 2.3: Improve Module 3 Weight Matrix Visualization (SCOPED DOWN)

**File**: `<local-path>`
**Lines**: 1252-1317 (drawWeightHeatmap), 1522+ (showZoomedHeatmap)

**Current Issue**:
- 784x128 weight matrix too dense to inspect
- Zoom functionality exists but may not be discoverable

**Scope Adjustment (per Critic feedback)**:
Mouse wheel zoom and pan/drag are NOT natively supported by Fyne's canvas.Raster.
Implementing them would require significant custom widget development (out of scope).

**Changes Required (ACHIEVABLE)**:

1. **Verify existing zoom works**:
   - Test `showZoomedHeatmap()` function at line 1522
   - Ensure zoom button is visible and functional
   - Document how to access zoomed view

2. **Add hover tooltip for cell values** (achievable with Fyne):
   - Create overlay label that follows mouse
   - Use existing `Mouseable` interface or `desktop.Hoverable`
   - Show: "Weight[row,col] = value"

3. **Add grid lines for small matrices**:
   - If matrix is < 64x64, overlay thin grid lines
   - Optional toggle (default: off for large matrices)

**DEFERRED to future iteration**:
- Mouse wheel zoom (requires custom canvas widget)
- Pan/drag navigation (requires custom canvas widget)

**Acceptance Criteria**:
- [ ] Zoom button opens larger view (verify existing functionality)
- [ ] Hover shows cell value and indices
- [ ] Grid visible on small matrices only
- [ ] No Fyne limitations violated

**Risk**: LOW - Using existing Fyne capabilities only

---

### Task 2.4: Add Colorbar Min/Max Labels + Consolidate ColorLegend

**Files**:
1. `<local-path>` (lines 18-189: ColorLegend)
2. `<local-path>`

**Current Issue**:
- Colorbar shows gradient but min/max values not clearly labeled
- Module2 has its own ColorLegend duplicating shared version
- Different API signatures between the two implementations

**API Differences (must bridge)**:
```go
// module2-crossbar version (lines 35-48):
NewColorLegend(minLabel, maxLabel, unit string, levels int) *ColorLegend

// shared/widgets version (lines 36-48):
NewColorLegend(minValue, maxValue float64, units string, vertical bool, colorFunc func(float64) color.RGBA) *ColorLegend
```

**Changes Required**:

1. **Consolidate ColorLegend widgets**:
   - Migrate module2 to use `shared/widgets.ColorLegend`
   - Create adapter function or update shared widget API
   - Remove duplicate in `module2-crossbar/pkg/gui/widgets.go` (lines 18-189)

2. **Update callers** (interface migration required):
   - `<local-path>` (lines 46-48)
   - `<local-path>` (lines 42-48)
   - Already uses shared: `tabs/irdrop_tab.go`, `tabs/sneak_tab.go`

3. **Add numeric labels**:
   - Min value at bottom
   - Max value at top
   - Mid-point label (optional)

4. **Update existing usage**:
   - Module 2 crossbar heatmaps
   - Module 3 weight heatmaps (already uses shared)
   - Module 5 comparison charts

**Acceptance Criteria**:
- [ ] Single ColorLegend implementation in shared/widgets
- [ ] Module2 callers updated to use shared widget
- [ ] No duplicate ColorLegend type in module2-crossbar
- [ ] Min/max values visible on all colorbars
- [ ] Labels update when data range changes

**Risk**: MEDIUM - Interface migration required, but limited to 2-3 files

---

## Phase 3: Medium Priority UI/UX

### Task 3.1: Increase Chart Label Font Sizes

**Specific Files and Lines (from grep search)**:

| File | Line | Current | Change To |
|------|------|---------|-----------|
| `module2-crossbar/pkg/gui/widgets.go` | 86 | `TextSize = 9` | `TextSize = 10` |
| `module2-crossbar/pkg/gui/widgets.go` | 96 | `TextSize = 9` | `TextSize = 10` |
| `module2-crossbar/pkg/gui/widgets.go` | 552 | `TextSize = 9` | `TextSize = 10` |
| `module2-crossbar/pkg/gui/widgets.go` | 559 | `TextSize = 9` | `TextSize = 10` |
| `module2-crossbar/pkg/gui/widgets.go` | 782 | `TextSize = 9` | `TextSize = 10` |
| `module2-crossbar/pkg/gui/widgets.go` | 798 | `TextSize = 8` | `TextSize = 10` |
| `module2-crossbar/pkg/gui/tabs/drift_tab.go` | 195 | `TextSize = 9` | `TextSize = 10` |

**Note**: Files with `TextSize = 96` (hero.go, market.go) are intentionally large for visual impact - DO NOT CHANGE.

**Changes Required**:
- Minimum font size: 10pt for labels
- Chart titles: 12-14pt bold
- Axis labels: 10-11pt

**Acceptance Criteria**:
- [ ] All 7 locations updated
- [ ] All text readable without zoom
- [ ] Consistent sizing across modules
- [ ] Hero/market large text unchanged

**Risk**: LOW - Font size changes only

---

### Task 3.2: Add Educational Tooltips

**Files**:
- `<local-path>`
- `<local-path>`
- `<local-path>`
- `<local-path>`

**Pattern**:
```go
// Add tooltip button next to labels
container.NewHBox(
    widget.NewLabel("Ec:"),
    widget.NewButtonWithIcon("", theme.InfoIcon(), func() {
        dialog.ShowInformation("Coercive Field (Ec)",
            "The electric field required to switch polarization...", window)
    }),
)
```

**Priority Tooltips**:
1. Ec (Coercive Field)
2. Pr (Remanent Polarization)
3. 30 Levels explanation
4. ADC/DAC bit resolution
5. TIA gain

**Acceptance Criteria**:
- [ ] Info icon next to key parameters
- [ ] Tooltip explains physics concept
- [ ] Consistent tooltip style across modules

**Risk**: LOW - Additive changes

---

### Task 3.3: Label Colorbar Tick Marks

**File**: `<local-path>`

**Current State**:
- Tick marks drawn at every 5th level (line 170-183)
- Numeric labels only at 0, 10, 20 (line 91-99)

**Changes Required**:
- Add labels at every 10th level: 0, 10, 20, 29
- Ensure labels don't overlap on small widgets

**Acceptance Criteria**:
- [ ] Tick labels visible at standard sizes
- [ ] Labels match actual data range

**Risk**: LOW

---

## Phase 4: File Refactoring

### Refactoring Principles

1. **Extract by Functionality**: Group related code
2. **Preserve Interfaces**: No API changes
3. **Incremental**: One file at a time
4. **Test After Each**: Run `go test ./...` after each refactor

---

### Task 4.1: Refactor module4-circuits tab_operations.go (1770 lines)

**File**: `<local-path>`

**Current Structure**:
- Lines 1-130: TappableArrayCanvas widget
- Lines 131-180: createOperationsView (unified view setup)
- Lines 181-240: createModeSelector, createSharedArraySection
- Lines 241-544: drawSharedArray (complex rendering)
- Lines 545-700: Write mode panel and functions
- Lines 700-1055: Read mode panel and functions
- Lines 1055-1430: Compute mode panel and functions
- Lines 1430-1770: Action handlers

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `tab_operations_write.go` | ~300 | createWriteModePanel, updateOpsWriteDataPath, drawOpsWritePulse |
| `tab_operations_read.go` | ~350 | createReadModePanel, drawOpsReadZone, onOpsRead, onOpsVerify |
| `tab_operations_compute.go` | ~400 | createComputeModePanel, computeAndUpdateAll, updateOpsComputeMath |
| `tab_operations.go` | ~720 | Core structure, TappableArrayCanvas, drawSharedArray, mode switching |

**Acceptance Criteria**:
- [ ] Each file < 500 lines
- [ ] No exported API changes
- [ ] `go build` passes
- [ ] `go test ./module4-circuits/...` passes

**Risk**: MEDIUM - Large refactor, ensure imports correct

---

### Task 4.2: Refactor module3-mnist dualmode.go (1545 lines)

**File**: `<local-path>`

**Current Structure**:
- Lines 1-165: Types and NewDualModeApp
- Lines 166-335: BuildContent, Start, Stop, createMainLayout
- Lines 335-480: createHeader, createDrawingZone, createResultsZone
- Lines 480-630: createControlsZone
- Lines 630-725: createWeightZone
- Lines 725-965: Inference methods (runInference, runInferenceAnimated)
- Lines 965-1165: Quick demo methods
- Lines 1165-1545: Weight visualization and utility methods

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `dualmode_controls.go` | ~300 | createControlsZone, applyPreset, applyPresetWithMode |
| `dualmode_weights.go` | ~300 | createWeightZone, drawWeightHeatmap, updateWeightHeatmap |
| `dualmode_inference.go` | ~350 | runInference, runInferenceAnimated, updateResultDisplays |
| `dualmode_demo.go` | ~200 | StartQuickDemo, StopQuickDemo, waitOrStop |
| `dualmode.go` | ~400 | Core app, layout, zones |

**Acceptance Criteria**:
- [ ] Each file < 400 lines
- [ ] No exported API changes
- [ ] All tests pass

**Risk**: MEDIUM

---

### Task 4.3: Refactor module1-hysteresis vulkan.go (1472 lines) [DEFERRED - HIGH RISK]

**File**: `<local-path>`

**Risk Assessment**: HIGH - Graphics code is tightly coupled, Vulkan has strict object lifecycle

**Mitigation Strategy (per Critic feedback)**:
1. **DEFER** this task until after Phase 2-3 UI/UX changes are stable
2. When ready, use conservative approach:
   - Keep ALL EXPORTED functions in `vulkan.go`
   - Only extract PRIVATE HELPER functions to new files
   - Preserve initialization/cleanup order

**Proposed Split (CONSERVATIVE)**:

| New File | Lines | Content |
|----------|-------|---------|
| `vulkan_shaders_internal.go` | ~300 | Private shader compilation helpers |
| `vulkan_pipeline_internal.go` | ~300 | Private pipeline setup helpers |
| `vulkan.go` | ~870 | ALL exports + core init + cleanup |

**Prerequisites before starting**:
- [ ] All Phase 2-3 tasks complete and stable
- [ ] Full test suite passing
- [ ] Manual visual verification of Module 1

**Acceptance Criteria**:
- [ ] All exported functions remain in vulkan.go
- [ ] Private helpers extracted to _internal.go files
- [ ] No change to public API
- [ ] Visual rendering unchanged

**Risk**: HIGH - Only attempt after other refactors proven stable

---

### Task 4.4: Refactor module2-crossbar widgets.go (1087 lines)

**File**: `<local-path>`

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `widgets_legend.go` | ~200 | ColorLegend (or remove - use shared) |
| `widgets_metrics.go` | ~150 | MetricsPanel, ComparisonBadge |
| `widgets_waterfall.go` | ~350 | AccuracyWaterfall, waterfallRenderer |
| `widgets_comparison.go` | ~300 | BeforeAfterToggle |
| `widgets.go` | ~100 | Common utilities, colormap functions |

**Acceptance Criteria**:
- [ ] Remove ColorLegend duplication (use shared/widgets)
- [ ] Each file focused on single widget type

**Risk**: LOW - Widget files are loosely coupled

---

### Task 4.5: Refactor module2-crossbar app_enhanced.go (987 lines)

**File**: `<local-path>`

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `app_tabs.go` | ~350 | Tab creation methods |
| `app_controls.go` | ~300 | Control panel, sliders, buttons |
| `app_analysis.go` | ~200 | Analysis functions, metrics |
| `app_enhanced.go` | ~200 | Core App struct, BuildContent, Start, Stop |

**Risk**: LOW

---

### Task 4.6: Refactor module6-eda learn_visuals.go (960 lines)

**File**: `<local-path>`

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `learn_visuals_transistor.go` | ~200 | Transistor visualization |
| `learn_visuals_cell.go` | ~200 | Cell structure visualization |
| `learn_visuals_array.go` | ~200 | Array layout visualization |
| `learn_visuals.go` | ~360 | Core tab, navigation |

**Risk**: LOW

---

### Task 4.7: Refactor module3-mnist network.go (957 lines)

**File**: `<local-path>`

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `network_inference.go` | ~300 | Infer, InferFP, InferCIM |
| `network_quantization.go` | ~250 | Quantize, GetQuantWeights |
| `network_config.go` | ~150 | Setters, getters, config loading |
| `network.go` | ~300 | Core struct, NewDualModeNetwork, LoadWeights |

**Risk**: LOW - Clean functional separation

---

### Task 4.8: Refactor module4-circuits tab_reference.go (951 lines)

**File**: `<local-path>`

**Proposed Split**:

| New File | Lines | Content |
|----------|-------|---------|
| `tab_reference_timing.go` | ~400 | Timing diagram section, draw functions |
| `tab_reference_specs.go` | ~450 | Specifications section, spec creation |
| `tab_reference.go` | ~100 | createReferenceTab, section switching |

**Risk**: LOW

---

## Verification Steps

### After Each Phase

1. **Build Check**:
   ```bash
   go build ./...
   ```

2. **Test Suite**:
   ```bash
   go test ./...
   ```

3. **Visual Verification**:
   ```bash
   ./launch.sh
   # Navigate to affected module
   # Verify UI changes
   ```

4. **Thread Safety Audit** (per Critic feedback):
   ```bash
   # Find all goroutine UI updates
   rg "go func|goroutine" --type go -A5 | grep -E "widget\.|canvas\.|Refresh"

   # Verify all use fyne.Do()
   rg "fyne\.Do\(" --type go
   ```

   **Rule**: Any UI update from a goroutine MUST be wrapped in `fyne.Do(func() { ... })`

### Phase-Specific Verification

| Phase | Verification |
|-------|--------------|
| Phase 1 | Run Module 1, verify text matches mode in all waveforms (expect: NO BUG) |
| Phase 2.1 | Screenshot charts, verify axis labels visible |
| Phase 2.2 | Open Module 4 Specs, verify readability |
| Phase 2.3 | Draw digit in Module 3, verify zoom button works, test hover tooltip |
| Phase 2.4 | Verify colorbar labels on all heatmaps |
| Phase 3 | Visual inspection of label sizes, tooltips |
| Phase 4 | Full test suite + visual inspection + thread safety audit |

### Thread Safety Checklist

Before claiming Phase 4 complete:
- [ ] Run grep for goroutine patterns
- [ ] Verify all UI updates use `fyne.Do()`
- [ ] No direct widget modifications from background goroutines
- [ ] Test with race detector: `go test -race ./...`

---

## Commit Strategy

### Suggested Commits

1. `fix(m1): clarify E-field physics text in Write/Read Demo mode`
2. `feat(m4): add axis labels to operations waveforms`
3. `feat(m4): improve specifications view readability`
4. `feat(m3): enhance weight matrix visualization with zoom`
5. `refactor(shared): consolidate ColorLegend widget`
6. `style: increase chart label font sizes across modules`
7. `feat: add educational tooltips for physics parameters`
8. `refactor(m4): split tab_operations.go by mode`
9. `refactor(m3): split dualmode.go by functionality`
10. `refactor(m2): split widgets.go by widget type`
11. `refactor(m4): split tab_reference.go into timing/specs`

---

## Success Criteria

- [ ] All tests pass (`go test ./...`)
- [ ] Race detector clean (`go test -race ./...`)
- [ ] No regressions in UI functionality
- [ ] All files < 500 lines after refactoring (except vulkan.go - deferred)
- [ ] Axis labels visible on all charts
- [ ] Colorbar values labeled (single implementation in shared/widgets)
- [ ] Specifications readable in Module 4
- [ ] Weight matrix zoom button works, hover tooltip functional
- [ ] Physics text verified correct (or fixed if bug confirmed)
- [ ] Thread safety audit complete

---

## Risk Assessment

| Risk | Mitigation |
|------|------------|
| Phase 1 bug is false positive | Investigation first, code change only if confirmed |
| Breaking UI during refactor | Incremental refactor, test after each step |
| Missing imports after split | IDE assistance, go build check |
| Visual regressions | Screenshot comparison before/after |
| Thread safety issues | Explicit audit step, race detector test |
| ColorLegend API mismatch | Create adapter or update shared widget API |
| Vulkan refactor breaks rendering | DEFER until other refactors stable, conservative approach |
| Mouse wheel zoom not possible | Already scoped down - use existing Fyne capabilities only |

---

## Dependencies

### Required Tools
- Go 1.21+
- Fyne v2.4+
- IDE with Go support (for refactoring)

### No External Dependencies Required
- All changes use existing Fyne widgets
- No new package imports needed

---

PLAN_READY: .omc/plans/fecim-bugfix-ux-refactor.md
