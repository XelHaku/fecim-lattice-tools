# WRD Phase-Boundary Logging Spec (Throttled)

Date: 2026-02-11  
Owner: module1-hysteresis GUI

## Scope
Define deterministic, low-noise logging at WRD phase boundaries for GUI parity checks.

## Phase model
UI/ops phase boundaries are reduced to:

- `PROGRAM` (phase id `0`)
- `VERIFY` (phase id `1`)
- `RESULT` (phase id `2`)

These map from controller/field state and are consumed by both:
- `PhaseIndicator` widget
- `LevelIndicator` target highlight + mode

## Required boundary log lines
Emit only when **display phase changes**:

- `▓▓ PROGRAM L<target> | ±E>Ec`
- `▒▒ VERIFY L<target> | E=0`
- `●● RESULT L<target> <status> [<rate>% rate]`

Status formatting:
- exact match: `✓ MATCH`
- off by 1: `△ ±1 (got <read>)`
- miss >1: `✗ miss (got <read>)`

## Throttle policy
To avoid GUI log spam from rapid phase oscillation:

- Min interval between boundary logs for same target: **400 ms**
- Always allow first boundary log for a target
- Always allow boundary log when target changes
- Still update internal `lastLogPhase` on every transition, even when emission is suppressed

## Implementation anchors
- Constant: `wrdPhaseBoundaryLogMinInterval`
- Gate: `shouldEmitWRDPhaseBoundaryLog(wrdTarget int)`
- Integration point: `refreshGUI()` WRD phase-transition logging path

## Parity contract
A single `widgetSnapshot` is the source of truth for phase + target widget state:

- `phaseWidgetSnapshot` (`mode`, `phase`)
- `targetWidgetSnapshot` (`level`, `highlight`, `mode`)

This prevents target-marker/phase-desync between controller state and rendered widgets.
