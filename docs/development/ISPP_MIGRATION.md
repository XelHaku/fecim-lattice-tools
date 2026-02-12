# ISPP Migration Plan (G04c)

**Status:** Proposed (design approved for implementation sequencing)  
**Owners:** Module 1 (hysteresis), Module 4 (circuits), shared physics maintainers  
**Related TODOs:** G04b (one-source-of-truth ISPP engine), G04c (this migration plan)

---

## 1) Problem Statement

ISPP behavior is currently split across multiple call sites with overlapping responsibilities:

- `shared/physics/ispp_write.go` (`WriteController`) — physics-aware write/verify loop driven by `LKSolver` and conductance target.
- `shared/physics/ispp_legacy.go` (`ISPPCalculator`) — level-space helper used by Module 4 legacy/faster loop.
- `module1-hysteresis/pkg/controller/writer.go` — module-local state machine (`Apply/Wait/Verify/Reset`) with its own bracketing/overshoot logic.

This causes drift risk (different convergence behavior, retries, overshoot handling, tolerance semantics), and makes cross-module regression parity harder.

---

## 2) Migration Goals

1. **Single authoritative ISPP engine API** in `shared` used by both Module 1 and Module 4.
2. Preserve two execution modes under one interface:
   - **Physics mode** (L-K / solver-backed)
   - **Fast level mode** (legacy-compatible, educational UI speed)
3. Maintain backward compatibility during migration through adapters.
4. Publish a clear deprecation schedule for old call sites.

Non-goals:
- Replacing all GUI animation/state code in one step.
- Reworking non-ISPP hysteresis pipelines.

---

## 3) Proposed Shared API Surface

Create package:

- `shared/ispp` (new)

### 3.1 Core Types

```go
type EngineMode string

const (
    EngineModePhysics EngineMode = "physics" // solver-backed (LK/Preisach backend wrappers)
    EngineModeFast    EngineMode = "fast"    // level-space stepping
)

type Config struct {
    Mode          EngineMode
    NumLevels     int
    TargetLevel   int
    MaxIterations int
    ToleranceLvls int

    // Pulse/field knobs (used by physics engine; ignored by fast mode as needed)
    PulseWidthS   float64
    MinVoltageV   float64
    MaxVoltageV   float64
    StepPercentEc float64

    // Retry/overshoot policy
    AllowResetRecovery bool
    MaxResets          int
}

type Snapshot struct {
    Attempt       int
    Phase         string // start|predict|apply|verify|overshoot|reset|success|failed
    CurrentLevel  int
    TargetLevel   int
    VoltageV      float64
    ErrorLevels   int
    Overshoots    int
    Complete      bool
    Success       bool
    FailureReason string
}

type Result struct {
    Attempts      int
    FinalLevel    int
    Overshoots    int
    Success       bool
    FailureReason string
}
```

### 3.2 Runtime Interfaces

```go
// Engine is the single entry point consumed by modules.
type Engine interface {
    Run(ctx Context, cfg Config) Result
}

// Context provides module-owned read/apply hooks (adapter boundary).
type Context interface {
    CurrentLevel() int
    TargetLevel() int
    ApplyPulse(voltageV float64, pulseWidthS float64) error
    VerifyLevel() (int, error)
    Reset(direction int) error // direction +1/-1 semantics documented in package
    OnEvent(Snapshot)
}
```

### 3.3 Built-in Engines

- `PhysicsEngine` (wraps `shared/physics.WriteController` behavior)
- `FastEngine` (wraps `shared/physics.ISPPCalculator` behavior)

### 3.4 Compatibility Adapters (temporary)

- `shared/ispp/compat/module1`
- `shared/ispp/compat/module4`

These adapters map module-specific state transitions/events into `Context` and preserve existing UI signals.

---

## 4) Adapter Interfaces for Callers

## 4.1 Module 1 Adapter

Current caller:
- `module1-hysteresis/pkg/controller/writer.go`

Adapter target:

```go
// Module1Adapter bridges existing simulation state machine to shared/ispp.Context
type Module1Adapter struct {
    // references to existing model/controller state
}

func (a *Module1Adapter) CurrentLevel() int
func (a *Module1Adapter) TargetLevel() int
func (a *Module1Adapter) ApplyPulse(voltageV, pulseWidthS float64) error
func (a *Module1Adapter) VerifyLevel() (int, error)
func (a *Module1Adapter) Reset(direction int) error
func (a *Module1Adapter) OnEvent(s sharedispp.Snapshot)
```

Mapping notes:
- Keep module1 phase labels (`APPLY/WAIT/VERIFY/RESET`) as display-level labels, but internal control comes from shared engine snapshots.
- Existing guard-band and logging lines become `OnEvent` projections (no independent convergence math).

## 4.2 Module 4 Adapter

Current callers:
- `module4-circuits/pkg/gui/device_state.go` (`StartISPP`, `ISPPIterate`, `HandleOvershoot`)
- `module4-circuits/pkg/gui/tab_unified_voltage.go` (L-K shared controller path)

Adapter target:

```go
// Module4Adapter bridges array/cell operations to shared/ispp.Context
type Module4Adapter struct {
    // device state + selected row/col
}

func (a *Module4Adapter) CurrentLevel() int
func (a *Module4Adapter) TargetLevel() int
func (a *Module4Adapter) ApplyPulse(voltageV, pulseWidthS float64) error
func (a *Module4Adapter) VerifyLevel() (int, error)
func (a *Module4Adapter) Reset(direction int) error
func (a *Module4Adapter) OnEvent(s sharedispp.Snapshot)
```

Mapping notes:
- Fast mode routes through existing level-space behavior (legacy parity).
- Physics mode routes through existing L-K pulse application, including half-select/disturb visualization hooks.
- Existing UI status messages should be fed from `Snapshot` so both modes share output semantics.

---

## 5) Migration Phases

## Phase 0 — Design freeze (this doc)
- Confirm API names and event schema.
- Lock behavior parity matrix (Preisach/LK, fast/physics, target sweeps).

## Phase 1 — Shared package introduction
- Add `shared/ispp` with interfaces/types and adapter shims.
- Implement wrappers that call current `shared/physics` internals (no behavior change target).

## Phase 2 — Module 4 adoption first
- Replace direct `ISPPCalculator` and direct `WriteController` invocations in Module 4 with `shared/ispp.Engine`.
- Keep existing GUI API stable (`StartISPP`, `ISPPIterate`, etc.) via adapter wrapper methods.

## Phase 3 — Module 1 adoption
- Replace convergence/bracketing core in `module1-hysteresis/pkg/controller/writer.go` with `shared/ispp.Engine` calls through adapter.
- Retain phase/UI cadence but move decision logic to shared engine.

## Phase 4 — Validation + cleanup
- Add cross-module regression assertions for matching end-state and attempt/overshoot accounting.
- Remove dead duplicated logic and old direct entry points.

---

## 6) Deprecation Timeline (Old Call Sites)

### Release N (start migration)
- Add deprecation notices in comments:
  - `shared/physics/ispp_legacy.go` exported constructor/helpers marked `Deprecated: use shared/ispp`.
  - `module1-hysteresis/pkg/controller/writer.go` internal convergence methods marked migration-owned.
- No removals.

### Release N+1 (dual-path)
- Default modules to `shared/ispp` adapters.
- Keep legacy paths behind explicit fallback flags/envs for rollback.

### Release N+2 (removal)
- Remove direct module usage of:
  - `shared/physics.ISPPCalculator` from module code paths.
  - module-local convergence math in Module 1 writer.
- Keep minimal compatibility wrappers only if external packages require them.

### Release N+3 (hard cleanup)
- Remove compatibility wrappers not used internally.
- Keep docs/tests only for shared API.

---

## 7) Acceptance Criteria

1. Module 1 and Module 4 call `shared/ispp.Engine` (no direct per-module convergence core).
2. Headless regression suites still pass for Preisach + LK targets.
3. Event/log output retains sufficient diagnostics (attempts, overshoots, failure reason).
4. TODO G04b and G04c can be marked complete with commit evidence.

---

## 8) Risks and Mitigations

- **Risk:** UI behavior drift during adapter swap.  
  **Mitigation:** keep phase labels/status sourced from snapshots and assert with focused GUI tests.

- **Risk:** physics parity regressions under LK edge cases.  
  **Mitigation:** run existing headless regression (`scripts/run_headless_ispp_regressions.sh`) before and after each phase.

- **Risk:** hidden dependency on module-local retry semantics.  
  **Mitigation:** encode retry policy explicitly in `Config` and cover in adapter tests.

---

## 9) Immediate Next Implementation Task (for G04b)

Implement `shared/ispp` scaffolding + Module 4 adapter first (smaller blast radius), then Module 1 migration.
