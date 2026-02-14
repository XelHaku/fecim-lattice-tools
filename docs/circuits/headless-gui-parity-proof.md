# RG-PAR-01 — Headless/GUI physics parity proof (module4-circuits)

Goal: prove that **headless** and **GUI** execution paths in `module4-circuits/` share the *same physics implementation* (no GUI-only / headless-only physics divergence).

This document is evidence-based: it lists the concrete physics functions/types used by READ/WRITE/COMPUTE and the packages that provide them.

## 1) Audit results: no headless-only physics code

Search for headless-specific implementations in non-test code:

- `grep -rn 'func.*headless\|func.*Headless' module4-circuits/ --include='*.go' | grep -v test`
- Result: **no matches** (no headless-only physics functions).

Conclusion: there is **no separate headless physics fork** inside module4.

## 2) Where physics lives (authoritative sources)

Physics used by module4 comes from only two places:

1. **`fecim-lattice-tools/shared/physics`**
   - Material presets (`HZOMaterial`), LK solver, ISPP calculator, conductance ↔ polarization mappings, geometry scalings, etc.
2. **`fecim-lattice-tools/module4-circuits/pkg/arraysim`**
   - Array/network solvers and invariants (Tier A/B solvers, transient solve). This is shared by both GUI and headless CLI.

No other package provides an alternative physics pipeline for headless vs GUI.

## 3) GUI path: physics functions and their import paths

Primary GUI state machine:

- File: `module4-circuits/pkg/gui/device_state.go`
- Imports physics as: `sharedphysics "fecim-lattice-tools/shared/physics"`

### READ path physics usage (GUI)

The GUI uses shared physics for:

- Materials/geometry:
  - `sharedphysics.HZOMaterial`
  - `sharedphysics.GeometryFromMaterial(mat)`
  - `sharedphysics.DefaultCellGeometry()` (defaults)
  - `CellGeometry.Film.ConductanceScale(...)` (geometry scaling)

- Level ↔ conductance mapping:
  - `sharedphysics.GetLevel(gNorm, levels)`

- Conductance ↔ polarization mapping (for LK and coupled voltage programming):
  - `sharedphysics.ConductanceToPolarization(currentG, gmin, gmax, mat.Ps)`
  - `sharedphysics.PolarizationToConductance(p, |Ps|, gmin, gmax)`
  - `sharedphysics.PolarizationToConductanceWithParams(...)`
  - `sharedphysics.ParseConductanceModel(mat.ConductanceModel)`

### WRITE / ISPP physics usage (GUI)

ISPP control is delegated to shared physics:

- `sharedphysics.NewISPPCalculator(ec, numLevels)`
- `sharedphysics.GetDirection(currentLevel, targetLevel)`
- `sharedphysics.HysteresisDirection` (`DirectionAscending`, `DirectionDescending`, `DirectionUnknown`)
- `sharedphysics.ISPPResult` / status enums (`ISPPSuccess`, `ISPPContinue`, etc.)

LK solver is also shared:

- `sharedphysics.NewLKSolver()`
- `solver.ConfigureFromMaterial(mat)`
- `solver.Step(E, dt)`
- `solver.SetState(P)` / `solver.GetState()`

### COMPUTE path physics usage (GUI)

Compute operations in module4 depend on the same `DeviceState` material + geometry + mapping functions above, and on `pkg/arraysim` for circuit-level solving.

## 4) Headless path: same physics packages

Module4 headless execution exists primarily via:

- CLI entry: `module4-circuits/cmd/circuits/main.go`
- Shared solver layer: `module4-circuits/pkg/arraysim`

`pkg/arraysim` imports and uses `fecim-lattice-tools/shared/physics` directly (same as GUI):

Examples:

- File: `module4-circuits/pkg/arraysim/transient.go`
  - `sharedphysics.NewLKSolver()`
  - `solver.ConfigureFromMaterial(mat)`
  - `sharedphysics.PolarizationToConductance(...)`
  - defaulting: `sharedphysics.DefaultHZO()`

- File: `module4-circuits/pkg/arraysim/array_config.go`, `types.go`
  - `sharedphysics.CellGeometry`
  - `sharedphysics.DefaultCellGeometry()`
  - `sharedphysics.GeometryFromMaterial(mat)`

Therefore the headless circuit simulation path consumes the **same** material definitions, mappings, and LK/ISPP primitives as the GUI.

## 5) No duplicate conductance/ISPP/LK/Preisach implementations in GUI-only code

Targeted scan for duplicated physics implementations in non-test code:

- `grep -rn 'func.*Conductance\|func.*ISPP\|func.*Solve\|func.*LandauK\|func.*Preisach' module4-circuits/ --include='*.go' | grep -v test`

Findings:

- GUI defines **control/state plumbing** (e.g. `DeviceState.StartISPP`, `DeviceState.ISPPIterate`) but delegates physics to `shared/physics`.
- `pkg/arraysim` defines **circuit/network solvers** (`TierASolver`, `TierBSolver`, transient solve, PCG, reference dense solve). These are not duplicated between GUI and headless; both use `pkg/arraysim`.
- No GUI-only or headless-only copies of:
  - ferroelectric material presets
  - conductance mapping formulas
  - LK solver
  - ISPP next-pulse decision logic

## 6) Tests that explicitly enforce parity

The repo already contains parity-oriented tests in the GUI package:

- `module4-circuits/pkg/gui/headless_gui_physics_parity_test.go`

These tests exercise GUI-visible paths while asserting equivalence against the shared/headless calculation paths.

## Conclusion

- **Headless and GUI share identical physics sources** for module4.
- Physics primitives (materials, LK, ISPP, conductance mapping, geometry) come from **`shared/physics`**.
- Circuit/array solving comes from **`module4-circuits/pkg/arraysim`**, used by both headless CLI and GUI.
- The codebase contains **no headless-only / GUI-only physics forks**.
