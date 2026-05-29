# ADR 0001: Level Calibration Workflow — UI-Neutral Service Architecture

The Module 1 Level Calibration Workflow needs a user interface. The existing Fyne-bound calibration path (`pkg/gui/calibration_runtime.go`) is tagged `legacy_fyne` and cannot be built in the zero-CGO default shell. We decided to extract the user-visible level-mapping logic into a UI-neutral domain service that both the Default UI Surface (gogpu) and the Legacy Fyne Surface can consume, rather than duplicating or porting the Fyne-specific implementation.

Update: the authoritative Level Calibration Engine now lives in `shared/physics/level_calibration.go` so downstream modules can reuse it without importing Module 1 internals. The older `module1-hysteresis/pkg/algo/CalibrationManager` remains an adaptive write-controller calibration helper, not the Level Calibration Workflow engine.

## Context

- The Default UI Surface must provide Level Calibration Workflow parity with the Legacy Fyne Surface before the latter is removed.
- Calibration logic (Preisach Delta tuning, ISPP level mapping, temperature interpolation) is purely computational — it does not depend on Fyne rendering.
- The current Level Calibration Workflow computation lives in `shared/physics/level_calibration.go`; the Fyne-specific `calibration_runtime.go` only adds status display and user interaction.
- `module1-hysteresis/pkg/algo/CalibrationManager` is write-controller adaptive tuning state and should not be treated as the user-facing Level Calibration Engine.
- Gogpu UI and Fyne UI use different rendering APIs; sharing the computational core avoids porting physics code.

## Decision

- **Level Calibration Engine** lives in `shared/physics/` as a UI-neutral service with a clean `Calibrate(Inputs) -> Summary` interface.
- **Level Calibration State** (not-calibrated/stale/fresh) is derived from the engine output, not from UI timestamps.
- **Level Calibration Export** is an explicit user action that writes artifacts (JSON/CSV) from the engine summary; it does not depend on the UI framework.
- The Legacy Fyne Surface may keep `calibration_runtime.go` as a thin adapter that calls the engine and formats results.

## Considered Options

1. **Port the Fyne calibration to gogpu directly** — rejected because it would duplicate the computational code and create a maintenance burden when the Fyne Surface is removed.
2. **Keep calibration in Fyne-only path** — rejected because it would leave the Default UI Surface without calibration parity.
3. **Extract UI-neutral service** — chosen. The computation is straightforward math; the UI layer is thin and replaceable.

## Consequences

- `shared/physics/` now carries the Level Calibration Engine contract (`shared/physics/level_calibration.go`), which downstream modules (Module 4 ISPP, Module 6 EDA export) can also consume.
- The Legacy Fyne Surface's `calibration_runtime.go` becomes a thin wrapper; its test coverage is not critical for the Default UI Surface migration.
- New user-facing Level Calibration Workflow computation should be added to `shared/physics/`, not to GUI code. Adaptive write-controller tuning may remain in Module 1 internals when it is not a cross-module workflow.
