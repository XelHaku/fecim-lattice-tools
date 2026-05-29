# ADR 0003: ISPP Convergence Policy Module

Module 1 and shared physics both need ISPP write-verify convergence behavior. The previous architecture kept collapsed-bounds recovery and related convergence heuristics inside concrete writer implementations. We introduced a UI-neutral ISPP Convergence Policy Module so these decisions can become shared, receipt-producing policy instead of duplicated controller implementation detail.

## Context

- Module 1's waveform-based controller (`module1-hysteresis/pkg/controller/writer.go`) owns level-oriented write state for P-E demonstrations.
- Shared physics' L-K controller (`shared/physics/ispp_write.go`) owns voltage/conductance write behavior for physics and downstream modules.
- ADR 0002 established that guard-band correction, overshoot handling, and bounds recovery are deliberate convergence behavior, not incidental implementation details.
- Architecture review evidence showed two writer implementations and two test clusters carrying similar bounds/overshoot knowledge, weakening locality.

## Decision

- The **ISPP Convergence Policy Module** lives in `shared/physics/isppconv`.
- Its interface is UI-neutral and unit-neutral: callers pass pulse magnitude bounds in their own native units (electric field for Module 1, voltage for L-K adapters).
- Policy functions return **ISPP Convergence Receipts** that state whether a rule changed bounds and whether recovery reset to full range.
- Concrete writers remain adapters that own their solver/waveform state, but delegate shared convergence policy decisions to `isppconv`.
- The first extracted policy is collapsed-bounds recovery via `RecoverCollapsedBounds`.

## Considered Options

1. **Keep policy in Module 1 writer only** — rejected because shared physics has an independent WriteController used by Module 4 and validation paths.
2. **Move all writer state into shared physics immediately** — rejected because it is too broad and risks behavior drift in the Default UI Surface and headless regressions.
3. **Extract one small policy module with receipts** — chosen. It deepens the seam incrementally while preserving existing adapters and tests.

## Consequences

- Bounds recovery now has one public test surface in `shared/physics/isppconv` plus adapter-level regression tests in Module 1 and shared physics.
- Future ISPP guard, overshoot, or convergence-acceptance rules should be considered for `isppconv` when they are shared by more than one adapter.
- Writer implementations should keep solver-specific state local and avoid moving UI, waveform, or logging concerns into `isppconv`.
- New policy extraction must remain behavior-preserving unless introduced with a failing test that describes the intended behavior change.
