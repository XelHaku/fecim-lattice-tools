# Work Plan: EDA Module Tab Consolidation

## Context

### Original Request
Consolidate EDA module (module6-eda) from 7 tabs to 2 tabs:
1. **Builder & Validation** - ONE unified view combining all build/export/validate functionality
2. **Learn** - Keep existing educational content

### Current State
The module6-eda has 7 active tabs plus 5 unused legacy files:

**Active Tabs (currently in use):**
| Tab | File | Lines | Purpose |
|-----|------|-------|---------|
| 1. Cell Builder | `cell_builder_tab.go` | 181 | LEF/LIB/V generation for bitcell |
| 2. Array Builder | `array_builder_tab.go` | 147 | Array dimensions, mode, architecture |
| 3. Verilog Export | `verilog_export_tab.go` | 105 | Generate Verilog netlist |
| 4. DEF Export | `def_export_tab.go` | 183 | Generate placement DEF |
| 5. Validation | `validation_tab.go` | 170 | Yosys/DEF/cross-file validation |
| 6. Learn | `learn_tab.go` | 289 | Educational content (KEEP AS-IS) |
| 7. Export All | `export_all_tab.go` | 211 | One-click package generation |

**Legacy Files (never used by current tabs):**
| File | Lines | Used By | Action |
|------|-------|---------|--------|
| `compiler_tab.go` | 147 | Nothing (requires AppState) | DELETE |
| `export_tab.go` | 198 | Nothing (requires AppState) | DELETE |
| `hdl_tab.go` | 348 | Nothing (requires AppState) | DELETE |
| `layout_tab.go` | 277 | Nothing (requires AppState) | DELETE |
| `state.go` | 12 | Only legacy tabs | DELETE |

### Identified Issues

1. **DEF generation duplicated** - `generateDEF()` in `def_export_tab.go:91-139` and `generateDEFContent()` in `export_all_tab.go:164-210` are identical 49-line functions
2. **No shared state** - Each tab creates its own widgets, no coordination
3. **Fragmented workflow** - User must navigate 6 tabs to complete basic flow
4. **Validation scattered** - Validate buttons in tabs 3, 4, and 5

---

## Work Objectives

### Core Objective
Create a single unified "Builder & Validation" tab that replaces tabs 1-5 and 7, providing a streamlined workflow from configuration to validated export.

### Deliverables
1. New file `builder_validation_tab.go` (~400-500 lines)
2. Updated `app.go` with 2-tab structure
3. Updated `embedded.go` with 2-tab structure
4. Deleted legacy and consolidated files (11 files total)

### Definition of Done
- [ ] Module builds without errors: `go build ./module6-eda/...`
- [ ] Only 2 tabs visible in UI: "Builder & Validation" and "Learn"
- [ ] All functionality preserved (cell config, array config, Verilog, DEF, validation, export-all)
- [ ] Single source of truth for DEF generation (no duplication)
- [ ] Tests pass: `go test ./module6-eda/...`

---

## Must Have / Must NOT Have

### MUST HAVE
- Single entry point for all build/export/validate operations
- All input fields from tabs 1-2 (cell name, dimensions, timing, array rows/cols, mode, architecture)
- Split preview showing both Verilog and DEF simultaneously
- "Generate All" button that creates all files in one action
- "Validate All" button with inline status display
- Progress feedback during generation
- All existing functionality from tabs 1-5, 7

### MUST NOT HAVE
- Duplicate DEF generation code
- Any reference to AppState (legacy pattern)
- Any reference to legacy tabs (compiler, export, hdl, layout)
- Unused imports or dead code
- Breaking changes to Learn tab

---

## Task Flow

```
Task 1 (Create Unified Tab)
    |
    v
Task 2 (Update Entrypoints) --depends on--> Task 1
    |
    v
Task 3 (Delete Files) --depends on--> Task 2
    |
    v
Task 4 (Verify Build) --depends on--> Task 3
```

---

## Detailed TODOs

### Task 1: Create Unified Builder & Validation Tab
**File:** `<local-path>`

**Acceptance Criteria:**
- [ ] File created with proper package declaration
- [ ] `MakeBuilderValidationTab(cfg *config.ArrayConfig) fyne.CanvasObject` function exists
- [ ] Imports only what's needed (no legacy imports)
- [ ] Contains single `generateDEF()` function (extracted from `def_export_tab.go:91-139`)

**Structure:**
```go
// Section 1: Cell Configuration (from cell_builder_tab.go)
// - Cell name, width, height, cell type
// - Rise time, fall time, input cap, leakage
// - "Generate Cell Files" button

// Section 2: Array Configuration (from array_builder_tab.go)
// - Rows, columns entries
// - Mode selector (storage/memory/compute)
// - Architecture selector (passive/1t1r)
// - Statistics display (total cells, area, WL/BL lengths)

// Section 3: Preview Panel (from verilog_export_tab.go + def_export_tab.go)
// - Tabbed preview: Verilog | DEF | Layout Viz
// - Stats display (instances, lines, size)

// Section 4: Actions Bar
// - "Generate All" button (from export_all_tab.go logic)
// - "Validate All" button (from validation_tab.go)
// - "Save Package" button (export to directory)
// - Status label with progress

// Section 5: Validation Results (inline, from validation_tab.go)
// - Yosys result
// - DEF result
// - Cross-check result
// - Log output area
```

**Key Implementation Details:**
1. Extract `generateDEF(cfg config.ArrayConfig) string` as the ONLY DEF generator
2. Use `export.GenerateArrayVerilog()` for Verilog (already exists in export package)
3. Use `export.GenerateLEF()`, `export.GenerateLiberty()`, `export.GenerateCellVerilog()` for cell files
4. Wire validation to `validation.ValidateVerilogWithCell()`, `validation.ValidateDEF()`, `validation.CrossCheckFiles()`
5. Use goroutine for "Generate All" with progress updates via `fyne.Do()`

**Layout:**
```
+------------------------------------------+
| Cell Config       | Array Config         |
| (form fields)     | (form fields + stats)|
+------------------------------------------+
| Preview: [Verilog] [DEF] [Layout]        |
| +--------------------------------------+ |
| | (scrollable preview content)         | |
| +--------------------------------------+ |
+------------------------------------------+
| [Generate All] [Validate All] [Save Pkg] |
| Status: Ready                            |
+------------------------------------------+
| Validation Results (collapsible)         |
| Yosys: - | DEF: - | Cross: -            |
| (log output)                            |
+------------------------------------------+
```

---

### Task 2: Update Entry Points
**Files:**
- `<local-path>`
- `<local-path>`

**Acceptance Criteria:**
- [ ] `app.go` creates only 2 views: "1. Builder & Validation", "2. Learn"
- [ ] `embedded.go` creates only 2 tabs
- [ ] No references to removed tab functions
- [ ] Default view is "1. Builder & Validation"

**Changes to `app.go`:**
```go
// Remove these lines:
cellBuilderContent := tabs.MakeCellBuilderTab()
arrayBuilderContent := tabs.MakeArrayBuilderTab(arrayConfig)
verilogContent := tabs.MakeVerilogExportTab(arrayConfig)
defContent := tabs.MakeDEFExportTab(arrayConfig)
validationContent := tabs.MakeValidationTab(arrayConfig)
exportAllContent := tabs.MakeExportAllTab(arrayConfig)

// Replace with:
builderContent := tabs.MakeBuilderValidationTab(arrayConfig)
learnContent := tabs.MakeLearnTab(&tabs.AppState{}, w)  // Note: AppState still needed by Learn tab visuals

viewNames := []string{
    "1. Builder & Validation",
    "2. Learn",
}

allViews := []fyne.CanvasObject{
    builderContent,
    learnContent,
}
```

**Changes to `embedded.go`:**
```go
// Replace AppTabs content with:
container.NewAppTabs(
    container.NewTabItem("1. Builder & Validation", tabs.MakeBuilderValidationTab(arrayConfig)),
    container.NewTabItem("2. Learn", tabs.MakeLearnTab(nil, window)),
)
```

---

### Task 3: Delete Obsolete Files
**Files to delete:**

**Legacy (never used):**
- `<local-path>`
- `<local-path>`
- `<local-path>`
- `<local-path>`
- `<local-path>`

**Consolidated (replaced by builder_validation_tab.go):**
- `<local-path>`
- `<local-path>`
- `<local-path>`
- `<local-path>`
- `<local-path>`
- `<local-path>`

**Acceptance Criteria:**
- [ ] All 11 files deleted
- [ ] No orphaned imports in remaining files
- [ ] `git status` shows deletions

---

### Task 4: Verify Build and Test
**Commands:**
```bash
cd <local-path>
go build ./module6-eda/...
go test ./module6-eda/...
go build ./cmd/fecim-visualizer
```

**Acceptance Criteria:**
- [ ] `go build ./module6-eda/...` exits 0
- [ ] `go test ./module6-eda/...` exits 0
- [ ] `go build ./cmd/fecim-visualizer` exits 0
- [ ] Running `./fecim-visualizer` shows 2 tabs in Module 6

---

## Commit Strategy

### Commit 1: Add unified builder tab
```
feat(eda): Add unified Builder & Validation tab

Combines functionality from 6 separate tabs into one streamlined view:
- Cell configuration (LEF/LIB/V generation)
- Array configuration (rows, cols, mode, architecture)
- Verilog and DEF preview with split view
- Inline validation with Yosys, DEF, and cross-check
- One-click "Generate All" and "Validate All" buttons

Consolidates duplicate DEF generation into single function.
```

### Commit 2: Update entry points and clean up
```
refactor(eda): Consolidate to 2-tab UI structure

- Update app.go and embedded.go to use new unified tab
- Remove 6 obsolete tab files (cell_builder, array_builder,
  verilog_export, def_export, validation, export_all)
- Remove 5 legacy tab files (compiler, export, hdl, layout, state)
- Keep Learn tab unchanged

Reduces module from 7 tabs to 2 tabs.
```

---

## Success Criteria

| Criterion | How to Verify |
|-----------|---------------|
| Build passes | `go build ./module6-eda/... && go build ./cmd/fecim-visualizer` |
| Tests pass | `go test ./module6-eda/...` |
| 2 tabs only | Run app, count tabs in Module 6 |
| No duplication | `grep -r "generateDEF" module6-eda/` returns 1 result |
| All functionality preserved | Manual test: configure cell, configure array, generate, validate, export |
| Learn tab works | Click Learn tab, navigate topics |

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| AppState required by Learn tab visuals | Keep `state.go` type definition inline in `learn_tab.go` if needed, or pass nil |
| Export package functions change | Only use existing public functions, no modifications |
| Fyne layout issues in unified view | Use proven patterns: HSplit, VSplit, Border layouts |
| Large file (400+ lines) | Well-structured sections with clear comments |

---

## Files Summary

### Files to CREATE (1)
- `module6-eda/pkg/gui/tabs/builder_validation_tab.go`

### Files to MODIFY (2)
- `module6-eda/pkg/gui/app.go`
- `module6-eda/pkg/gui/embedded.go`

### Files to DELETE (11)
- `module6-eda/pkg/gui/tabs/cell_builder_tab.go`
- `module6-eda/pkg/gui/tabs/array_builder_tab.go`
- `module6-eda/pkg/gui/tabs/verilog_export_tab.go`
- `module6-eda/pkg/gui/tabs/def_export_tab.go`
- `module6-eda/pkg/gui/tabs/validation_tab.go`
- `module6-eda/pkg/gui/tabs/export_all_tab.go`
- `module6-eda/pkg/gui/tabs/compiler_tab.go`
- `module6-eda/pkg/gui/tabs/export_tab.go`
- `module6-eda/pkg/gui/tabs/hdl_tab.go`
- `module6-eda/pkg/gui/tabs/layout_tab.go`
- `module6-eda/pkg/gui/tabs/state.go`

### Files to KEEP (unchanged)
- `module6-eda/pkg/gui/tabs/learn_tab.go`
- `module6-eda/pkg/gui/tabs/learn_visuals.go`
