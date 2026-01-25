# Module 4 Circuits Refactoring Plan

**Created**: 2026-01-25
**Status**: READY FOR EXECUTION (Revision 3)
**Complexity**: MEDIUM-HIGH
**Estimated Files**: 8 files (consolidated from 11)

---

## 0. MAJOR CHANGE: View Consolidation (NEW)

**User Requirement**: "The less views the better" - consolidate 6 views → 3 views

| Current (6 views) | New (3 views) |
|-------------------|---------------|
| WRITE | → |
| READ | → **OPERATIONS** (combined) |
| COMPUTE | → |
| COMPARISON | → **COMPARISON** (unchanged) |
| TIMING | → |
| SPECS | → **REFERENCE** (combined) |

### New Architecture

**View 1: OPERATIONS** - Combined WRITE/READ/COMPUTE with mode selector
- Mode toggle: [WRITE] [READ] [COMPUTE]
- Shared array (always visible, left panel)
- Mode-specific config and data path (right panel)

**View 2: COMPARISON** - FeFET vs CPU/GPU (keep as-is)

**View 3: REFERENCE** - Combined TIMING + SPECS
- Sub-sections for timing diagrams and specifications

---

## 1. Requirements Summary

### Primary Objective
Consolidate 6 views into 3 views AND refactor monolithic app.go into organized files.

### Current State
- **Location**: `<local-path>`
- **app.go**: 3129 lines containing all 6 tabs and helper functions
- **embedded.go**: 99 lines (wrapper, unchanged)

### Critical Bug
**Timing diagrams render as sparse dots instead of continuous waveforms.**

Root cause analysis (lines 2349-2377 in `drawTimingWrite`):
```go
for pct := 0; pct <= 100; pct++ {
    x := marginLeft + pct*plotW/100
    // ...
    img.Set(x, lineY, cyanColor)  // Only sets ONE pixel
}
```

The loop iterates 101 times (0-100) but `plotW` may be 500+ pixels, creating gaps between points. The horizontal lines need interpolation or thicker rendering.

**Note on existing code patterns:**
- `drawTimingRead` (line 2425) and `drawTimingCompute` (line 2553) use `prevLineY` to track vertical transitions
- The fix must integrate horizontal interpolation WITH the existing `prevLineY` vertical transition logic
- `drawTimingWrite` (line 2305) uses a different pattern (`prevHigh` boolean) for transitions

---

## 2. Target File Structure (REVISED)

```
module4-circuits/pkg/gui/
├── app.go              # CircuitsApp struct, init, Run(), createMainLayout() (~250 lines)
├── theme.go            # DELETE - use shared/theme instead
├── tab_operations.go   # NEW: Combined WRITE/READ/COMPUTE with mode selector (~1000 lines)
├── tab_comparison.go   # Comparison tab (unchanged, apply theme) (~370 lines)
├── tab_reference.go    # NEW: Combined TIMING + SPECS (~600 lines)
├── font.go             # fontPatterns, drawSimpleText (~130 lines)
├── helpers.go          # drawRect, min, max, section headers, ComponentBox (~150 lines)
├── drawing.go          # drawThickHorizontalLine, drawArrow (~100 lines)
└── embedded.go         # EmbeddedCircuitsApp (unchanged)
```

**Key Changes:**
- DELETE `theme.go` - use `shared/theme` instead
- DELETE `tab_write.go`, `tab_read.go`, `tab_compute.go` - consolidated into `tab_operations.go`
- DELETE `tab_timing.go`, `tab_specs.go` - consolidated into `tab_reference.go`
- NEW `tab_operations.go` - unified WRITE/READ/COMPUTE with mode selector
- NEW `tab_reference.go` - unified TIMING + SPECS

---

## 3. Acceptance Criteria

### Functional Requirements
- [ ] All 6 tabs (Write, Read, Compute, Comparison, Timing, Specs) render correctly
- [ ] Timing diagrams display continuous waveforms (not sparse dots)
- [ ] All buttons and interactions work as before
- [ ] No regressions in existing functionality

### Code Quality Requirements
- [ ] Each file has a clear single responsibility
- [ ] All files in same package `gui` with consistent imports
- [ ] No circular dependencies
- [ ] All tests pass: `go test ./module4-circuits/...`
- [ ] Build succeeds: `go build ./...`

### Timing Diagram Fix Requirements
- [ ] `drawTimingWrite` renders continuous horizontal lines
- [ ] `drawTimingRead` renders continuous horizontal lines
- [ ] `drawTimingCompute` renders continuous horizontal lines
- [ ] Vertical transitions visible at signal changes
- [ ] Lines are 2-3 pixels thick for visibility

---

## 4. Implementation Steps

### Phase 1: Create Helper Files (Risk: LOW)

#### Task 1.1: Create helpers.go
**File**: `<local-path>`
**Extract from app.go lines**: 794-802 (drawRect), 2987-2999 (min/max)
**Content**:
- `drawRect()` function (794-802)
- `min()` function (2987-2992)
- `max()` function (2994-2999)

#### Task 1.2: Create drawing.go (NEW FILE)
**File**: `<local-path>`
**Content**: New shared helper for timing diagram rendering
```go
package gui

import (
    "image"
    "image/color"
)

// drawThickHorizontalLine draws a horizontal line from x1 to x2 at y with specified thickness.
// This is used by timing diagram functions to render continuous waveforms instead of sparse dots.
func drawThickHorizontalLine(img *image.RGBA, x1, x2, y int, thickness int, c color.Color) {
    h := img.Bounds().Dy()
    halfT := thickness / 2
    for px := x1; px <= x2; px++ {
        for dy := -halfT; dy <= halfT; dy++ {
            py := y + dy
            if py >= 0 && py < h {
                img.Set(px, py, c)
            }
        }
    }
}
```

#### Task 1.3: Create font.go
**File**: `<local-path>`
**Extract from app.go lines**: 3001-3129
**Content**:
- `drawSimpleText()` (3006-3008)
- `drawSimpleChar()` (3011-3013)
- `drawScaledText()` (3016-3022)
- `drawScaledChar()` (3025-3047)
- `fontPatterns` map (3049-end of file)

#### Task 1.4: Create theme.go
**File**: `<local-path>`
**Extract from app.go lines**: 36-89
**Content**:
- All `colorXxx` variables (36-55)
- `feCIMTheme` struct and methods (57-89)

### Phase 2: Create Tab Files (Risk: MEDIUM)

#### Task 2.1: Create tab_write.go
**File**: `<local-path>`
**Extract from app.go lines**: 335-1017
**Functions to include**:
- `createWriteTab()` (339-417)
- `createWriteConfigSection()` (419-521)
- `createWriteCellSection()` (523-582)
- `createWriteDataPathSection()` (584-611)
- `createLabeledBox()` (613-628)
- `createLabeledBoxWithLabel()` (630-644)
- `updateWriteDataPath()` (646-675)
- `createWritePulseSection()` (677-681)
- `drawWritePulse()` (683-791)
- `refreshWritePulse()` (804-810)
- `createWriteArraySection()` (812-816)
- `drawWriteArray()` (818-905)
- `refreshWriteArray()` (907-913)
- `createWriteMappingSection()` (915-918)
- `getMappingText()` (920-989)
- `onProgramCell()` (991-1004)
- `onProgramRandomArray()` (1006-1017)

#### Task 2.2: Create tab_read.go
**File**: `<local-path>`
**Extract from app.go lines**: 1019-1467
**Functions to include**:
- `createReadTab()` (1023-1098)
- `createReadConfigSection()` (1100-1163)
- `createReadCellSection()` (1165-1203)
- `createReadDataPathSection()` (1205-1245)
- `createReadZoneSection()` (1247-1251)
- `drawReadZone()` (1253-1365)
- `refreshReadZone()` (1367-1373)
- `createReadResultsSection()` (1375-1388)
- `onReadCell()` (1390-1454)
- `onReadAllCells()` (1456-1460)
- `onVerifyArray()` (1462-1467)

#### Task 2.3: Create tab_compute.go
**File**: `<local-path>`
**Extract from app.go lines**: 1469-1889
**Functions to include**:
- `createComputeTab()` (1473-1548)
- `createComputeConfigSection()` (1550-1580)
- `createComputeInputSection()` (1582-1639)
- `updateComputeInputs()` (1641-1653)
- `createComputeVizSection()` (1655-1659)
- `drawComputeViz()` (1661-1752) - **NOTE: function is drawComputeViz, NOT drawComputeArray**
- `createComputeMathSection()` (1754-1764)
- `createComputeOutputSection()` (1766-1781)
- `onCompute()` (1783-1816)
- `updateComputeMath()` (1818-1849)
- `onAnimateCompute()` (1851-1869)
- `onResetCompute()` (1871-1889)

#### Task 2.4: Create tab_comparison.go
**File**: `<local-path>`
**Extract from app.go lines**: 1891-2246
**Functions to include**:
- `createComparisonTab()` (1895-1947)
- `createCompArchSection()` (1949-2029)
- `drawCompArch()` (1955-2029) - nested
- `createCompTimingSection()` (2031-2115)
- `drawCompTiming()` (2037-2115) - nested
- `createCompEnergySection()` (2117-2186)
- `drawCompEnergy()` (2123-2186) - nested
- `createCompTableSection()` (2188-2227)
- `onRunComparison()` (2229-2246)

#### Task 2.5: Create tab_timing.go (INCLUDES BUG FIX)
**File**: `<local-path>`
**Extract from app.go lines**: 2248-2719
**Functions to include**:
- `createTimingTab()` (2252-2297)
- `createTimingWriteSection()` (2299-2303)
- `drawTimingWrite()` (2305-2417) - **FIX REQUIRED**
- `createTimingReadSection()` (2419-2423)
- `drawTimingRead()` (2425-2506) - **FIX REQUIRED**
- `createTimingComputeSection()` (2547-2551)
- `drawTimingCompute()` (2553-2705) - **FIX REQUIRED**
- `refreshTimingDiagrams()` (2707-2719) - **EXISTS at line 2707**

**BUG FIX for timing functions**:

**IMPORTANT**: Three different patterns exist in the timing functions:
1. `drawTimingWrite` uses `prevHigh` boolean for transitions
2. `drawTimingRead` uses `prevLineY` for vertical transition tracking
3. `drawTimingCompute` uses `prevLineY` for vertical transition tracking

The fix must add horizontal line interpolation WHILE preserving the existing vertical transition logic.

**NEW SHARED HELPER** (create in `drawing.go`):
```go
// drawThickHorizontalLine draws a horizontal line from x1 to x2 at y with thickness
func drawThickHorizontalLine(img *image.RGBA, x1, x2, y int, thickness int, c color.Color) {
    h := img.Bounds().Dy()
    halfT := thickness / 2
    for px := x1; px <= x2; px++ {
        for dy := -halfT; dy <= halfT; dy++ {
            py := y + dy
            if py >= 0 && py < h {
                img.Set(px, py, c)
            }
        }
    }
}
```

**FIX for drawTimingWrite** (uses prevHigh pattern):
```go
// BEFORE (sparse dots, line 2376):
img.Set(x, lineY, cyanColor)

// AFTER (continuous lines):
// Add prevX tracking before loop:
prevX := marginLeft
// Inside loop, after setting lineY:
drawThickHorizontalLine(img, prevX, x, lineY, 3, cyanColor)
prevX = x
```

**FIX for drawTimingRead and drawTimingCompute** (use prevLineY pattern):
```go
// BEFORE (sparse dots, lines 2504 and 2638):
img.Set(x, lineY, cyanColor)

// AFTER (continuous lines):
// Add prevX tracking at start of signal loop:
prevX := marginLeft
// Inside loop, BEFORE the vertical transition code:
if prevX < x {
    // Draw horizontal line at PREVIOUS y level up to current x
    yToDraw := lineY
    if prevLineY != -1 {
        yToDraw = prevLineY  // Use previous y for horizontal segment
    }
    drawThickHorizontalLine(img, prevX, x, yToDraw, 3, cyanColor)
}
// After vertical transition code, update prevX:
prevX = x
// Keep existing: prevLineY = lineY
```

#### Task 2.6: Create tab_specs.go
**File**: `<local-path>`
**Extract from app.go lines**: 2721-2984
**Functions to include**:
- `createSpecsTab()` (2725-2818)
- `createSpecArraySection()` (2820-2841)
- `createSpecDACSection()` (2843-2869)
- `createSpecADCSection()` (2871-2897)
- `createSpecTIASection()` (2899-2923)
- `createSpecFeFETSection()` (2925-2967)
- `createSpecSummarySection()` (2969-2984)

### Phase 3: Refactor app.go (Risk: MEDIUM)

#### Task 3.1: Slim down app.go
**File**: `<local-path>`
**Keep only**:
- Package declaration and imports (1-25)
- Constants (27-34)
- `CircuitsApp` struct (91-180)
- `NewCircuitsApp()` (182-214)
- `initializeArray()` (216-231)
- `Run()` (233-243)
- `createMainLayout()` (245-333)

**Remove**: All tab-specific code (moved to tab_*.go files)

### Phase 4: Verification (Risk: LOW)

#### Task 4.1: Build verification
```bash
go build ./module4-circuits/...
```

#### Task 4.2: Test verification
```bash
go test ./module4-circuits/...
```

#### Task 4.3: Visual verification
- Launch application: `./fecim-visualizer`
- Navigate to each tab
- Verify timing diagrams show continuous waveforms
- Test all button interactions

---

## 5. Risk Assessment

### High Risk Areas
| Area | Risk | Mitigation |
|------|------|------------|
| Import resolution | Functions call each other across files | All files in same package `gui` |
| Method receivers | All methods on `*CircuitsApp` | Keep receiver consistent |
| Timing bug fix | Logic change could break rendering | Test each signal type independently |

### Medium Risk Areas
| Area | Risk | Mitigation |
|------|------|------------|
| Variable scope | Color constants used everywhere | Move to theme.go, accessible package-wide |
| Canvas references | UI elements stored in struct | Struct unchanged, methods split |

### Low Risk Areas
| Area | Risk | Mitigation |
|------|------|------------|
| Font patterns | Pure data, no dependencies | Simple extraction |
| Helper functions | Utility functions, no state | Simple extraction |

---

## 6. Verification Steps

### Pre-Implementation Checks
- [ ] Confirm go.mod exists and is valid
- [ ] Run existing tests to establish baseline
- [ ] Build current code to confirm it compiles

### Post-Implementation Checks
- [ ] `go build ./...` succeeds
- [ ] `go test ./module4-circuits/...` passes
- [ ] `gofmt -w module4-circuits/pkg/gui/` runs clean
- [ ] Visual test: Timing tab shows continuous waveforms
- [ ] Visual test: All 6 tabs render correctly
- [ ] Visual test: All buttons respond correctly

### Regression Test Checklist
- [ ] WRITE tab: Program cell works
- [ ] WRITE tab: Random array works
- [ ] READ tab: Read cell shows calculations
- [ ] READ tab: Voltage zone diagram updates
- [ ] COMPUTE tab: MVM computation works
- [ ] COMPARISON tab: Architecture diagrams render
- [ ] TIMING tab: **Continuous waveforms (not dots)**
- [ ] SPECS tab: Configuration changes apply

---

## 7. Commit Strategy

### Commit 1: Extract helper files
```
refactor(module4): extract font, drawing, and helper utilities

- Create font.go with fontPatterns and text drawing
- Create helpers.go with drawRect, min, max utilities
- Create drawing.go with drawThickHorizontalLine helper
- Create theme.go with colors and feCIMTheme
```

### Commit 2: Extract tab files
```
refactor(module4): split tabs into separate files

- Create tab_write.go, tab_read.go, tab_compute.go
- Create tab_comparison.go, tab_timing.go, tab_specs.go
- Slim down app.go to core initialization only
```

### Commit 3: Fix timing diagrams
```
fix(module4): render timing diagrams as continuous waveforms

- Fix sparse dot rendering in drawTimingWrite (prevHigh pattern)
- Fix sparse dot rendering in drawTimingRead (prevLineY pattern)
- Fix sparse dot rendering in drawTimingCompute (prevLineY pattern)
- Use drawThickHorizontalLine helper with 3px thickness
- Preserve vertical transition rendering logic
```

---

## 8. Definition of Done

- [ ] app.go reduced from 3129 to ~200 lines
- [ ] All 11 files compile without errors (including new drawing.go)
- [ ] All existing tests pass
- [ ] Timing diagrams display continuous waveforms
- [ ] No visual regressions in any tab
- [ ] Code formatted with gofmt
- [ ] All commits follow conventional commit format

---

## 9. File Dependencies Graph

```
app.go (main struct)
   ├── theme.go (colors, theme)
   ├── helpers.go (utilities: drawRect, min, max)
   ├── drawing.go (timing diagram line drawing: drawThickHorizontalLine)
   ├── font.go (text rendering)
   ├── tab_write.go → helpers.go, font.go, theme.go
   ├── tab_read.go → helpers.go, font.go, theme.go
   ├── tab_compute.go → helpers.go, font.go, theme.go
   ├── tab_comparison.go → helpers.go, font.go, theme.go
   ├── tab_timing.go → helpers.go, font.go, theme.go, drawing.go
   └── tab_specs.go → helpers.go, font.go, theme.go

embedded.go → app.go (wraps CircuitsApp)
```

---

## 10. Success Criteria

| Metric | Target |
|--------|--------|
| app.go line count | < 250 lines |
| Total file count | 11 files |
| Build status | PASS |
| Test status | PASS |
| Timing diagram bug | FIXED |
| Visual regressions | NONE |

---

## 11. Revision History

| Version | Date | Changes |
|---------|------|---------|
| 1 | 2026-01-25 | Initial plan |
| 2 | 2026-01-25 | Critic feedback addressed: fixed function names (drawComputeViz not drawComputeArray), added missing functions (updateComputeInputs, updateComputeMath), enumerated all Specs tab functions (createSpecDACSection, createSpecADCSection, createSpecTIASection, createSpecFeFETSection), corrected refreshTimingDiagrams location (line 2707), added drawThickHorizontalLine helper in new drawing.go, improved timing bug fix to integrate with prevLineY pattern |

---

**PLAN_READY: .omc/plans/module4-refactor.md**
