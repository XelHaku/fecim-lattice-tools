# Work Plan: OpenLane Integration for FeCIM EDA Validation (REVISED)

## Context

### Original Request
Integrate real OpenLane tools into the FeCIM EDA validation pipeline, including OpenROAD placement/DRC validation, with support for both Docker and native installation modes.

### Critic Feedback Addressed

| Issue | Critic Feedback | Resolution |
|-------|-----------------|------------|
| 1 | Don't clone OpenLane repo - just use Docker image | Removed repo cloning; use `docker pull` for Docker image |
| 2 | Fix Docker command pattern | Use `--entrypoint openroad` not bare command |
| 3 | Fix OpenROAD TCL for clockless design | FeCIM crossbar has no clock; use `check_placement` not STA |
| 4 | Magic can't read DEF | Use OpenROAD's internal DRC (`check_placement -verbose`) |
| 5 | LVS is netgen not Magic | Deferred to Phase 2 (requires GDS extraction) |
| 6 | Add PDK setup | Document volare setup for SKY130 PDK |

### Scope Revision

**Phase 1 (This Plan):**
- Docker image pull (NOT repo clone)
- OpenROAD placement check and cell usage (NOT STA - design is clockless)
- OpenROAD internal DRC via `check_placement`
- Native mode detection for tools in PATH
- UI integration
- PDK setup guidance via volare

**Phase 2 (Future - Out of Scope):**
- Magic DRC (requires GDS, not DEF)
- Netgen LVS (requires GDS extraction from Magic)
- Full OpenLane flow integration
- GDS export from OpenROAD

### Research Findings

**Existing Codebase Patterns:**
- `pkg/validation/yosys.go`: Uses `exec.Command` for tool invocation with `exec.LookPath` for availability checks
- `pkg/validation/def_validator.go`: Internal Go parser for DEF syntax validation
- `pkg/validation/cross_check.go`: Cross-file consistency checking (LEF/LIB/V)
- `pkg/gui/tabs/builder_validation_tab.go`: Fyne GUI with goroutine-based validation, `fyne.Do()` for UI updates
- `pkg/config/types.go`: Clean config structs with SKY130 defaults
- `pkg/export/openlane_config.go`: Already generates OpenLane v2.0 config.json

**Key Technical Corrections:**

1. **Docker Command (Correct Pattern):**
```bash
docker run --rm --entrypoint openroad \
  -v $(pwd):/design -w /design \
  -v $PDK_ROOT:/pdk:ro \
  -e TECH_LEF=/pdk/sky130A/libs.tech/openlane/sky130_fd_sc_hd/sky130_fd_sc_hd.tlef \
  -e CELL_LEF=/pdk/sky130A/libs.ref/sky130_fd_sc_hd/lef/sky130_fd_sc_hd.lef \
  -e DEF_FILE=/design/placement.def \
  ghcr.io/the-openroad-project/openlane:latest \
  -no_splash -exit /design/check_placement.tcl
```

2. **OpenROAD TCL for Clockless Design (FeCIM Crossbar):**
```tcl
# FeCIM crossbar has no clock - use placement validation, not STA
read_lef $env(TECH_LEF)
read_lef $env(CELL_LEF)
read_def $env(DEF_FILE)

# Placement validation (this IS the DRC for placement)
check_placement -verbose

# Cell usage report
report_cell_usage

# Design summary
report_design

exit
```

3. **PDK Setup via Volare:**
```bash
pip install volare
volare enable --pdk sky130 sky130A
export PDK_ROOT=~/.volare
```

**SKY130 PDK Paths (via volare):**
- Tech LEF: `$PDK_ROOT/sky130A/libs.tech/openlane/sky130_fd_sc_hd/sky130_fd_sc_hd.tlef`
- Cell LEF: `$PDK_ROOT/sky130A/libs.ref/sky130_fd_sc_hd/lef/sky130_fd_sc_hd.lef`
- Liberty: `$PDK_ROOT/sky130A/libs.ref/sky130_fd_sc_hd/lib/sky130_fd_sc_hd__tt_025C_1v80.lib`

---

## Work Objectives

### Core Objective
Enable real EDA validation using OpenROAD tools with Docker image pull (not repo clone) and native mode detection, focused on placement validation for clockless FeCIM crossbar designs.

### Deliverables
1. **OpenLane Manager** (`pkg/openlane/manager.go`) - Docker image management, native tool detection
2. **OpenLane Runner** (`pkg/openlane/runner.go`) - Tool execution with correct Docker `--entrypoint` pattern
3. **Placement Validation** (`pkg/validation/openlane.go`) - OpenROAD `check_placement` and cell usage (NOT STA)
4. **Configuration** (`pkg/openlane/config.go`) - PDK paths, volare setup guidance, timeouts
5. **UI Updates** (`pkg/gui/tabs/builder_validation_tab.go`) - OpenLane status, PDK setup help, new validations

### Definition of Done
- [ ] Docker image pulled via `docker pull` (NOT repo clone)
- [ ] Docker mode uses `--entrypoint openroad` pattern
- [ ] Native mode detects `openroad` in PATH
- [ ] OpenROAD `check_placement` returns structured results
- [ ] OpenROAD cell usage report returns structured results
- [ ] PDK setup guidance shown when PDK_ROOT missing
- [ ] UI shows Docker image status (pulled/not pulled)
- [ ] UI shows PDK status with volare setup instructions
- [ ] All new validation types integrated into "Validate All"
- [ ] Tests pass: `go test ./module6-eda/...`

---

## Guardrails

### Must Have
- Follow existing validation pattern (`error` return, tool availability check)
- Use `fyne.Do()` for all UI updates from goroutines
- Use `--entrypoint openroad` for Docker commands (NOT bare command)
- Use `check_placement` for validation (NOT STA - clockless design)
- Support both Docker and native modes transparently
- Structured result types (not just pass/fail)
- Timeout handling for long-running operations
- PDK setup guidance via volare

### Must NOT Have
- Cloning OpenLane repo (use Docker image pull only)
- STA analysis (design is clockless)
- Magic DRC (Phase 2 - requires GDS)
- LVS checks (Phase 2 - requires GDS extraction)
- Blocking UI during long operations
- Hard-coded paths (use config + environment)
- Docker-only mode (must have native fallback)
- Breaking existing Yosys/DEF/Cross-check validations

---

## Task Flow and Dependencies

```
[T1: Manager] ─────┬─────> [T3: Validation]
                   │              │
[T2: Runner] ──────┘              │
                                  v
[T4: Config] ─────────────> [T5: UI Updates]
                                  │
                                  v
                           [T6: Integration Tests]
```

**Parallel Groups:**
- Group A: T1 + T2 (can be done together, define interfaces)
- Group B: T3 + T4 (after Group A)
- Group C: T5 (after Group B)
- Group D: T6 (after Group C)

---

## Detailed TODOs

### T1: OpenLane Manager (`pkg/openlane/manager.go`)

**Purpose:** Manage Docker image and detect native tools (NO REPO CLONING)

**Acceptance Criteria:**
- [ ] `IsDockerImagePulled() bool` - Check if `ghcr.io/the-openroad-project/openlane` is pulled
- [ ] `PullDockerImage(progress func(string)) error` - Pull image with progress
- [ ] `GetDockerImageVersion() (string, error)` - Get image tag/version
- [ ] `DetectMode() Mode` - Return Docker, Native, or None
- [ ] `IsDockerAvailable() bool` - Check docker command availability
- [ ] `IsNativeOpenROADAvailable() bool` - Check if `openroad` is in PATH
- [ ] `IsPDKInstalled() bool` - Check if PDK_ROOT is set and valid
- [ ] `GetPDKSetupInstructions() string` - Return volare setup instructions

**Implementation Notes:**
```go
package openlane

type Mode int

const (
    ModeNone Mode = iota
    ModeDocker
    ModeNative
)

type Manager struct {
    dockerImage string // ghcr.io/the-openroad-project/openlane:latest
    mode        Mode
    pdkRoot     string
}

func NewManager() *Manager
func (m *Manager) IsDockerImagePulled() bool
func (m *Manager) PullDockerImage(progress func(string)) error  // docker pull
func (m *Manager) DetectMode() Mode
func (m *Manager) IsPDKInstalled() bool
func (m *Manager) GetPDKSetupInstructions() string

// Check Docker image existence:
// docker images -q ghcr.io/the-openroad-project/openlane:latest
// If empty string returned, image not pulled

// Check native tool:
// exec.LookPath("openroad")
```

**File:** `<local-path>`

---

### T2: OpenLane Runner (`pkg/openlane/runner.go`)

**Purpose:** Execute OpenROAD with correct Docker `--entrypoint` pattern

**Acceptance Criteria:**
- [ ] `RunOpenROAD(tclScript string, workDir string) (*Result, error)` - Execute OpenROAD
- [ ] Docker mode: Use `--entrypoint openroad` pattern (NOT bare command)
- [ ] Native mode: Direct `openroad` invocation
- [ ] `RunWithTimeout(cmd, timeout) ([]byte, error)` - Timeout wrapper
- [ ] Environment variables passed correctly (TECH_LEF, CELL_LEF, DEF_FILE)

**Implementation Notes:**
```go
type Runner struct {
    manager *Manager
    config  *Config
}

func NewRunner(manager *Manager, config *Config) *Runner

func (r *Runner) RunOpenROAD(tclScript string, workDir string) (*Result, error)

type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
    Duration time.Duration
}
```

**Docker Command Pattern (CORRECT):**
```bash
docker run --rm --entrypoint openroad \
  -v /path/to/design:/design -w /design \
  -v $PDK_ROOT:/pdk:ro \
  -e TECH_LEF=/pdk/sky130A/libs.tech/openlane/sky130_fd_sc_hd/sky130_fd_sc_hd.tlef \
  -e CELL_LEF=/pdk/sky130A/libs.ref/sky130_fd_sc_hd/lef/sky130_fd_sc_hd.lef \
  -e DEF_FILE=/design/placement.def \
  ghcr.io/the-openroad-project/openlane:latest \
  -no_splash -exit /design/script.tcl
```

**Native Command:**
```bash
TECH_LEF=... CELL_LEF=... DEF_FILE=... openroad -no_splash -exit script.tcl
```

**File:** `<local-path>`

---

### T3: Placement Validation (`pkg/validation/openlane.go`)

**Purpose:** Validate placement using OpenROAD (NOT STA - clockless design)

**Acceptance Criteria:**
- [ ] `RunPlacementCheck(defPath string) (*PlacementResult, error)` - check_placement validation
- [ ] `RunCellUsage(defPath string) (*CellUsageResult, error)` - Cell usage report
- [ ] Results are structured with violation lists
- [ ] TCL script uses `check_placement -verbose` (NOT `report_checks`)

**Implementation Notes:**
```go
type PlacementResult struct {
    Passed         bool
    ViolationCount int
    Violations     []PlacementViolation
    RawOutput      string
}

type PlacementViolation struct {
    Cell     string
    Issue    string  // "overlap", "out_of_bounds", "unplaced"
    Location string  // "x,y"
    Message  string
}

type CellUsageResult struct {
    TotalCells   int
    CellTypes    map[string]int  // cell_name -> count
    UtilizationPct float64
    RawOutput    string
}
```

**OpenROAD TCL Script for Clockless FeCIM Design:**
```tcl
# check_placement.tcl - For clockless FeCIM crossbar
# NOTE: No STA since design has no clock

read_lef $env(TECH_LEF)
read_lef $env(CELL_LEF)
read_def $env(DEF_FILE)

# Placement validation (this is the DRC for placement)
puts "=== PLACEMENT CHECK ==="
check_placement -verbose

# Cell usage report
puts "=== CELL USAGE ==="
report_cell_usage

# Design summary
puts "=== DESIGN SUMMARY ==="
report_design

exit
```

**File:** `<local-path>`

---

### T4: Configuration (`pkg/openlane/config.go`)

**Purpose:** Centralized configuration with PDK setup guidance

**Acceptance Criteria:**
- [ ] `Config` struct with PDK path, preferred mode, timeouts
- [ ] `DefaultConfig()` with SKY130 defaults
- [ ] `LoadConfig(path string)` from JSON file
- [ ] `SaveConfig(path string)` to JSON file
- [ ] Environment variable overrides (PDK_ROOT)
- [ ] `GetVolareSetupInstructions()` - Return volare setup commands

**Implementation Notes:**
```go
type Config struct {
    PDKRoot        string        `json:"pdk_root"`        // ~/.volare or custom
    PDKVariant     string        `json:"pdk_variant"`     // sky130A
    SCLibrary      string        `json:"sc_library"`      // sky130_fd_sc_hd
    PreferredMode  Mode          `json:"preferred_mode"`  // docker, native
    TimeoutPlacement time.Duration `json:"timeout_placement"` // 5 min
    DockerImage    string        `json:"docker_image"`    // ghcr.io/the-openroad-project/openlane:latest
}

func DefaultConfig() *Config {
    pdkRoot := os.Getenv("PDK_ROOT")
    if pdkRoot == "" {
        pdkRoot = filepath.Join(os.Getenv("HOME"), ".volare")
    }
    return &Config{
        PDKRoot:         pdkRoot,
        PDKVariant:      "sky130A",
        SCLibrary:       "sky130_fd_sc_hd",
        PreferredMode:   ModeDocker,
        TimeoutPlacement: 5 * time.Minute,
        DockerImage:     "ghcr.io/the-openroad-project/openlane:latest",
    }
}

func GetVolareSetupInstructions() string {
    return `To set up SKY130 PDK using volare:

1. Install volare:
   pip install volare

2. Enable SKY130 PDK:
   volare enable --pdk sky130 sky130A

3. Set PDK_ROOT environment variable:
   export PDK_ROOT=~/.volare

4. (Optional) Add to shell profile:
   echo 'export PDK_ROOT=~/.volare' >> ~/.bashrc
`
}
```

**Config File Location:** `~/.fecim/openlane-config.json`

**File:** `<local-path>`

---

### T5: UI Updates (`pkg/gui/tabs/builder_validation_tab.go`)

**Purpose:** Integrate OpenLane validation with PDK setup guidance

**Acceptance Criteria:**
- [ ] Docker image status indicator (Pulled/Not Pulled/Pulling)
- [ ] Pull Image button (disabled when pulled, shows progress during pull)
- [ ] Mode indicator (Docker/Native/None)
- [ ] PDK status indicator with setup instructions link
- [ ] New validation checkbox: Placement Check
- [ ] Detailed results display with expandable violation lists
- [ ] Progress bar during validation
- [ ] All UI updates via `fyne.Do()`

**UI Layout Changes:**
```
+-- OpenLane Status Section ----------------+
| Docker Image: [Pulled]  Mode: [Docker]    |
| [Pull Image]                              |
+-------------------------------------------+
| PDK Status: [Not Found]                   |
| [Show Setup Instructions]                 |
+-------------------------------------------+

+-- Validation Options ---------------------+
| [x] Yosys    [x] DEF    [x] Cross-check   |
| [x] Placement Check (OpenROAD)            |
| [Validate Selected]  [Validate All]       |
+-------------------------------------------+

+-- Results --------------------------------+
| Yosys: PASS  | DEF: PASS | Cross: PASS   |
| Placement: PASS (0 violations)            |
| Cell Usage: 128 cells, 45% utilization    |
| [Show Details...]                         |
+-------------------------------------------+
```

**Implementation Notes:**
- Add `dockerImageStatus *widget.Label`
- Add `pdkStatus *widget.Label`
- Add `pullImageBtn *widget.Button` with progress callback
- Add `showPdkSetupBtn *widget.Button` - shows volare instructions in dialog
- Add `placementResult *widget.Label`
- Modify `validateAllBtn` handler to include placement check
- Add async image pull with progress updates

**File:** `<local-path>`

---

### T6: Integration Tests

**Purpose:** Verify end-to-end functionality

**Acceptance Criteria:**
- [ ] Test Manager Docker image detection (mock `docker images`)
- [ ] Test Manager native tool detection (mock `exec.LookPath`)
- [ ] Test Runner Docker command construction (verify `--entrypoint` pattern)
- [ ] Test validation result parsing from OpenROAD output
- [ ] Test config load/save with volare paths
- [ ] Integration test with real OpenROAD (skipped if not available)

**Files:**
- `<local-path>`
- `<local-path>`
- `<local-path>`

---

## Commit Strategy

| Commit | Scope | Message |
|--------|-------|---------|
| 1 | T1 | `feat(eda): Add OpenLane manager for Docker image and native tool detection` |
| 2 | T2 | `feat(eda): Add OpenLane runner with correct --entrypoint Docker pattern` |
| 3 | T4 | `feat(eda): Add OpenLane config with volare PDK setup guidance` |
| 4 | T3 | `feat(eda): Add OpenROAD placement validation (clockless design)` |
| 5 | T5 | `feat(eda): Integrate OpenLane placement check into builder UI` |
| 6 | T6 | `test(eda): Add OpenLane integration tests` |

---

## Success Criteria

1. **Docker Image:** Running app with no image shows "Not Pulled" with working Pull button
2. **PDK Setup:** When PDK_ROOT missing, shows volare setup instructions
3. **Mode Detection:** Correctly identifies Docker vs Native based on available tools
4. **Validation Flow:** "Validate All" runs placement check with structured results
5. **Error Handling:** Graceful degradation when tools/PDK unavailable
6. **Performance:** UI remains responsive during validation
7. **Tests:** `go test ./module6-eda/...` passes

---

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| Docker not available | Native mode fallback with `openroad` in PATH |
| Docker pull fails | Retry logic, clear error message |
| PDK not installed | Clear volare setup instructions shown |
| Long validation times | Timeout + cancellation support |
| OpenROAD output parsing errors | Fallback to raw output display |

---

## Out of Scope (Phase 2)

| Feature | Reason | Future Approach |
|---------|--------|-----------------|
| Magic DRC | Magic reads GDS, not DEF | Export GDS from OpenROAD first |
| Netgen LVS | Requires GDS extraction | Magic extracts netlist, Netgen compares |
| Full OpenLane flow | Overkill for validation | Consider if user requests synthesis+P&R |
| STA analysis | FeCIM crossbar is clockless | Add when supporting clocked designs |

---

## PDK Setup Reference

**Volare Setup (Recommended):**
```bash
# Install volare
pip install volare

# Enable SKY130 PDK
volare enable --pdk sky130 sky130A

# Set environment variable
export PDK_ROOT=~/.volare

# Verify installation
ls $PDK_ROOT/sky130A/libs.ref/sky130_fd_sc_hd/
```

**Manual Setup (Alternative):**
```bash
# Clone skywater-pdk
git clone https://github.com/google/skywater-pdk.git
cd skywater-pdk
git submodule update --init libraries/sky130_fd_sc_hd/latest
export PDK_ROOT=$(pwd)
```

---

## Estimated Effort

| Task | Estimate |
|------|----------|
| T1: Manager | 1 hour |
| T2: Runner | 1 hour |
| T3: Validation | 1.5 hours |
| T4: Config | 30 min |
| T5: UI Updates | 1.5 hours |
| T6: Tests | 1 hour |
| **Total** | **6.5 hours** |
